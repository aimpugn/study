# 락, 래치, 데드락

## lock vs latch, row/table/gap/next-key/predicate lock

락(lock)은 트랜잭션 사이의 의미 있는 충돌을 조율하는 장치이고, 래치(latch)는 DB 엔진 내부 자료구조를 아주 짧게 보호하는 장치다. 이 절을 읽은 뒤에는 “락이 많다”라는 말을 들었을 때 그것이 애플리케이션 트랜잭션 락인지, 테이블/행/범위 락인지, InnoDB의 index gap인지, PostgreSQL의 predicate lock 관찰인지, 아니면 buffer page나 B-tree 내부 구조를 지키는 래치 경합인지 분리해서 말할 수 있어야 한다. 중요한 함정은 gap lock과 next-key lock을 row lock의 별명처럼 취급하는 것이다. InnoDB 공식 문서는 record lock이 index record에 걸리고, gap lock은 index record 사이의 간격 삽입을 막으며, next-key lock은 record lock과 그 앞 gap lock의 조합이라고 설명한다. PostgreSQL 공식 문서는 명시적 잠금 문서에서 table-level lock, row-level lock, page-level lock, advisory lock을 구분하고, row-level lock은 데이터 조회 자체를 막지 않고 같은 row를 쓰거나 잠그려는 동작을 막는다고 설명한다.

이 구분이 생긴 배경은 DB가 동시에 두 종류의 안전을 지켜야 하기 때문이다. 첫째, 사용자에게 보이는 데이터 의미를 지켜야 한다. 계좌 잔액을 두 세션이 동시에 덮어쓰면 안 되고, 예약 범위에 동시에 같은 좌석이 들어오면 안 된다. 이 영역이 트랜잭션 락의 세계다. 둘째, DB 엔진 내부의 메모리 자료구조와 디스크 page를 망가뜨리지 않아야 한다. B-tree page를 split하는 순간 두 thread가 같은 포인터를 동시에 바꾸면 index 구조 자체가 깨진다. 이 영역이 래치의 세계다. 트랜잭션 락은 commit/rollback까지 이어질 수 있지만, 래치는 보통 내부 critical section을 지나면 곧 풀린다. 따라서 운영에서 “lock wait”를 본다고 해서 곧바로 “엔진 내부 latch가 문제”라고 말하면 안 되고, 반대로 CPU가 latch spin으로 타는 상황을 “비즈니스 row lock”으로만 보면 원인을 놓친다.

첫 번째 벽돌은 같은 계좌 두 행을 업데이트하는 SQL이다.

```sql
CREATE TABLE transfer_account (
    account_id integer PRIMARY KEY,
    owner_name text NOT NULL,
    balance integer NOT NULL
);

INSERT INTO transfer_account VALUES
    (1, 'kim', 100),
    (2, 'lee', 100);
```

```text
row lock의 기본 움직임

time | session A                                      | session B
-----+------------------------------------------------+------------------------------------------
t1   | BEGIN;                                         |
t2   | UPDATE transfer_account                        |
     |    SET balance = balance - 10                  |
     |  WHERE account_id = 1;                         |
t3   |                                                | BEGIN;
t4   |                                                | UPDATE transfer_account
     |                                                |    SET balance = balance + 10
     |                                                |  WHERE account_id = 1;
t5   |                                                | -- B는 A가 잡은 row lock 때문에 기다림
t6   | COMMIT;                                        |
t7   |                                                | -- A commit 뒤 B가 진행하거나 재평가

상태 변화
account_id=1 row는 A의 UPDATE로 잠금 소유권이 생긴다.
B의 UPDATE는 같은 row를 쓰려 하므로 기다린다.
plain SELECT는 DBMS와 isolation/read type에 따라 과거 committed version을 볼 수 있다.
```

이 예제의 lock은 사용자가 볼 수 있는 트랜잭션 의미를 지킨다. A가 아직 커밋하지 않은 갱신을 B가 동시에 덮어쓰지 못하게 한다. PostgreSQL의 `SELECT FOR UPDATE`는 선택한 row를 update하려는 것처럼 잠그고, 다른 트랜잭션의 UPDATE, DELETE, SELECT FOR UPDATE류를 기다리게 한다. InnoDB의 `SELECT ... FOR UPDATE`도 row 또는 index record에 exclusive 계열 lock을 잡아 다음 쓰기 충돌을 조율한다. 그러나 여기서 말하는 row lock은 “row 객체 하나에 이름표를 붙인다”보다 더 물리적인 면이 있다. InnoDB record lock은 index record lock이다. 테이블에 명시적 index가 없어도 InnoDB는 hidden clustered index를 만들어 record lock을 수행한다. 그래서 InnoDB 잠금을 이해할 때는 항상 “어떤 index range를 검색했는가”를 같이 봐야 한다.

테이블 락은 row보다 넓은 단위다. PostgreSQL에서는 `ACCESS SHARE`, `ROW SHARE`, `ROW EXCLUSIVE`, `SHARE`, `ACCESS EXCLUSIVE` 같은 table-level lock mode가 있고, 이름에 row가 들어가도 table-level lock mode라는 역사적 이름이 있을 수 있다. 예를 들어 plain SELECT는 보통 `ACCESS SHARE`를 잡고, UPDATE/DELETE/INSERT는 target table에 `ROW EXCLUSIVE`를 잡는다. `DROP TABLE`, `TRUNCATE`, `VACUUM FULL` 같은 작업은 강한 `ACCESS EXCLUSIVE`를 잡아 plain SELECT까지 막을 수 있다. 이 지점에서 실무 함정이 나온다. “row만 조금 수정하는 배포”라고 생각했지만, migration이 강한 테이블 락을 잡으면 읽기까지 멈출 수 있다. DDL과 DML의 잠금 모드는 운영 배포에서 반드시 확인해야 한다.

InnoDB의 intention lock은 row lock과 table lock이 공존하기 위한 표지판에 가깝다. 어떤 트랜잭션이 table 안의 개별 row에 shared lock을 잡으려면 table-level intention shared lock을, exclusive row lock을 잡으려면 intention exclusive lock을 먼저 표시한다. intention lock 자체는 대부분 row 작업끼리 서로 막으려는 것이 아니라, 누군가 이 테이블 안에서 row를 잠글 예정이라는 정보를 table-level lock과 조율한다. 따라서 `IX`가 보인다고 해서 “테이블 전체가 exclusive로 막혔다”고 읽으면 안 된다. 이는 관측을 잘못 해석하는 흔한 운영 실수다.

```text
InnoDB multiple granularity sketch

table: transfer_account
  IX by session A
    -> index PRIMARY record account_id=1 has X record lock

table-level LOCK TABLES ... WRITE 요청
  -> IX와 충돌할 수 있으므로 기다림

다른 row update
  -> 같은 table에 IX를 둘 수 있고, 다른 index record면 진행 가능

관측 해석
IX는 "row 안쪽에 exclusive lock을 잡으려 한다"는 표지다.
IX만 보고 모든 row가 막혔다고 결론 내리면 FAIL이다.
```

gap lock과 next-key lock은 row lock의 확장이지만, 의미가 다르다. record lock은 존재하는 index record를 잠근다. gap lock은 index record 사이의 빈 공간, 첫 record 앞, 마지막 record 뒤의 간격을 잠근다. next-key lock은 어떤 index record와 그 앞 gap을 함께 잠근다. 이 구조는 phantom row를 막기 위해 등장했다. 조건 범위를 읽고 “이 범위에는 이런 값이 없다”고 판단하는 동안, 다른 트랜잭션이 그 빈틈에 새 index record를 넣으면 결과 집합이 바뀐다. InnoDB는 `REPEATABLE READ`에서 index search/scan에 next-key lock을 사용해 이런 삽입을 막을 수 있다.

```text
InnoDB index values

PRIMARY or secondary index key order:

    10        11        13        20
----|---------|---------|---------|----
 (-inf,10] (10,11] (11,13] (13,20] (20,+inf)

SELECT * FROM child WHERE id BETWEEN 11 AND 13 FOR UPDATE;

잠금 해석의 핵심
- record 11과 13만 잠긴다고 생각하면 부족하다.
- next-key lock은 record와 그 앞 gap을 함께 보아야 한다.
- gap lock의 목적은 주로 새 index record 삽입을 막는 것이다.
- 같은 gap lock이 항상 서로를 배타적으로 막는다고 생각하면 MySQL 공식 설명과 어긋난다.
```

이 구조 때문에 인덱스가 잠금 범위를 바꾼다. `WHERE id = 100`이 unique index 전체를 사용해 단일 row를 찾으면 gap locking이 필요 없을 수 있다. 반대로 nonunique index, range condition, composite unique index의 일부 컬럼만 쓰는 조건은 gap lock이 생길 수 있다. 따라서 “SELECT FOR UPDATE 한 줄”만 보고 lock 범위를 추정하면 부족하다. `EXPLAIN`, index definition, isolation level, actual predicate를 함께 보아야 한다. 실무에서 InnoDB lock wait가 넓게 퍼질 때는 SQL 문장보다 인덱스 설계가 원인인 경우가 많다.

PostgreSQL의 predicate lock은 InnoDB gap lock과 같은 물건으로 외우면 안 된다. PostgreSQL serializable에서 predicate lock은 직렬화 가능성을 보장하기 위한 충돌 추적의 관찰 표면이다. 사용자가 명시적으로 gap을 잠그는 감각과 다르다. PostgreSQL은 MVCC snapshot과 SSI가 결합되어, 어떤 읽기와 쓰기가 위험한 의존 관계를 만들면 serialization failure로 실패시킬 수 있다. 그래서 PostgreSQL에서 범위 불변식을 다룰 때는 `SERIALIZABLE`과 retry, 명시적 row lock, constraint, advisory lock 같은 선택지를 목적에 맞게 비교해야 한다.

래치는 사용자에게 직접 보이는 SQL 계약이 아니라 엔진 내부의 짧은 보호 장치다. 예를 들어 buffer pool 안의 page를 읽거나, B-tree page split 중 sibling pointer를 바꾸거나, lock table 내부 hash bucket을 수정할 때 여러 worker가 같은 메모리를 동시에 만지면 안 된다. 래치는 이 순간을 보호한다. 트랜잭션 rollback이 래치 획득을 되돌리는 것이 아니고, 사용자가 `COMMIT`까지 래치를 보유하는 것도 아니다. 래치 경합은 보통 CPU 사용률, wait event, mutex/spin 관측으로 보이고, 특정 row를 누가 잡고 있는지와는 다른 진단 도구가 필요하다.

| 구분 | lock | latch |
|---|---|---|
| 보호 대상 | 트랜잭션 의미, row/table/range/advisory resource | 메모리 page, buffer, B-tree 내부 구조, lock manager 자료구조 |
| 보유 시간 | statement 또는 transaction 끝까지 갈 수 있음 | 아주 짧은 critical section |
| 사용자가 관측하는 방식 | `pg_locks`, lock wait, deadlock log, `SHOW ENGINE INNODB STATUS` | wait event, mutex/latch 통계, CPU spin, engine internal metrics |
| rollback 의미 | transaction lock은 rollback/commit과 연결 | latch는 transaction 의미와 직접 연결되지 않음 |
| 실무 대응 | 트랜잭션 순서, index, SQL, timeout, retry | hot page, index design, workload skew, engine/version/config |

운영에서 이 둘을 헷갈리면 조치가 빗나간다. row lock 문제인데 CPU 튜닝만 하면 대기 트랜잭션은 계속 쌓인다. latch 경합 문제인데 고립 수준만 낮추면 hot index page나 monotonic key insert 병목은 남는다. 예를 들어 모든 주문이 `created_at` 순서로 같은 끝 page에 몰리는 index를 강하게 갱신하면 B-tree rightmost page 경합이 생길 수 있다. 이때 애플리케이션 로그에는 “DB가 느림”만 보일 수 있지만, 실제 원인은 특정 row lock이 아니라 내부 page/latch 경합일 수 있다. 반대로 `SELECT ... FOR UPDATE`를 긴 외부 API 호출 동안 잡고 있으면 명백한 트랜잭션 lock wait다. 두 경우 모두 “DB lock”이라고만 말하면 사고가 멈춘다.

관측 루틴은 다음처럼 나눈다.

```text
PostgreSQL lock 관측
  SELECT pid, locktype, mode, granted, relation::regclass, page, tuple
    FROM pg_locks
   ORDER BY granted, pid;

  SELECT pid, wait_event_type, wait_event, state, query
    FROM pg_stat_activity
   WHERE wait_event_type IS NOT NULL;

InnoDB lock 관측
  SHOW ENGINE INNODB STATUS\G

  SELECT * FROM performance_schema.data_locks;
  SELECT * FROM performance_schema.data_lock_waits;

latch/engine wait 관측
  PostgreSQL: wait_event_type이 LWLock, BufferPin, IO 등인지 본다.
  MySQL: performance_schema mutex/rwlock wait summary, InnoDB monitor를 본다.

PASS
  lock wait와 latch/engine wait를 같은 표에 뭉개지 않고 분리한다.

FAIL
  row lock holder도 찾지 않고 isolation만 바꾼다.
  wait_event가 내부 latch 계열인데 SELECT FOR UPDATE만 의심한다.
```

이 절의 공식 출처는 PostgreSQL `Explicit Locking` 문서와 MySQL 8.4 `InnoDB Locking` 문서다. 로컬 seed `database/lock.md`와 `database/postgresql/lock.md`는 lock 주제를 deep-dive로 확장해야 한다는 보존 신호로 사용했다. 두 seed는 매우 짧으므로 본문 사실의 근거로 그대로 확정하지 않았고, 공식 문서와 synthetic SQL trace로 다시 구성했다. 이 문서의 lab 후보는 `database/deep-dive/labs/transactions/du32-lock-latch-gap`이며, 실제 lab을 만들 때는 PostgreSQL `pg_locks` 관측과 MySQL `SHOW ENGINE INNODB STATUS`/`performance_schema.data_locks` 관측을 분리해야 한다.

마지막으로 한 문장으로 정리하면 이렇다. 락은 “다른 트랜잭션이 이 의미 있는 자원을 지금 건드려도 되는가”를 조율하고, 래치는 “엔진 내부 구조를 바꾸는 이 짧은 순간에 다른 worker가 끼어들면 안 된다”를 보장한다. gap lock, next-key lock, predicate lock은 모두 “존재하는 row 하나”를 넘어 범위나 읽기 조건을 다루려는 장치지만, DBMS마다 구현과 관측 표면이 다르다. 실무자는 lock이라는 단어를 들을 때 먼저 보호 대상, 보유 시간, 관측 view, 실패/대기 정책을 나누어야 한다.

락 모드를 실제로 읽을 때는 “강하다/약하다”라는 감각보다 compatibility, 즉 함께 존재할 수 있는지를 먼저 본다. PostgreSQL table lock mode 이름은 역사적 이유로 헷갈릴 수 있다. `ROW EXCLUSIVE`라는 이름이 들어가도 table-level lock mode이며, row 하나를 뜻하지 않는다. 이 mode는 UPDATE, DELETE, INSERT 같은 쓰기 명령이 target table에 잡는 table-level lock이다. 반대로 row-level lock mode인 `FOR UPDATE`, `FOR NO KEY UPDATE`, `FOR SHARE`, `FOR KEY SHARE`는 선택된 row에 대한 다른 writer/locker와의 충돌을 제어한다. 이 두 층을 섞으면 `pg_locks`를 읽을 때 `relation` lock과 `tuple` lock을 같은 의미로 해석하는 실수가 생긴다.

```text
PostgreSQL 잠금 관측을 두 층으로 나누기

table-level
  locktype = relation
  relation = transfer_account
  mode     = RowExclusiveLock, AccessShareLock, AccessExclusiveLock ...
  의미     = 테이블 구조/명령 실행의 호환성 조율

row-level
  locktype = tuple 또는 transactionid 대기와 함께 관측될 수 있음
  mode     = row-level locking clause와 연결
  의미     = 같은 row를 update/delete/lock하려는 트랜잭션 조율

해석 순서
1. plain SELECT가 막혔는가? 그렇다면 ACCESS EXCLUSIVE 같은 강한 table lock을 의심한다.
2. UPDATE만 막혔는가? row-level lock 또는 transactionid wait를 의심한다.
3. DDL과 DML이 충돌하는가? table-level lock mode compatibility를 본다.
4. wait_event_type이 Lock인지 LWLock인지 먼저 나눈다.
```

InnoDB에서는 lock report가 index 중심으로 나온다는 점이 중요하다. `RECORD LOCKS ... index PRIMARY ... locks rec but not gap`이라는 문구를 보면 “record만 잠겼고 gap은 아니다”라고 읽을 수 있다. 반대로 `lock_mode X`만 있고 supremum pseudo-record가 보이면 next-key 또는 gap까지 연결된 범위 잠금을 의심해야 한다. 이 정보는 단순 로그 장식이 아니라, 어떤 index range가 동시 삽입을 막았는지 찾는 단서다.

```text
InnoDB lock report 읽기 손풀이

RECORD LOCKS space id 58 page no 3 n bits 72 index `PRIMARY` of table `test`.`child`
trx id 10080 lock_mode X locks rec but not gap
Record lock, heap no 2 PHYSICAL RECORD:
  0: len 4; hex 8000000a; asc     ;;

읽는 방법
- index `PRIMARY`: 잠금은 PRIMARY index record를 기준으로 잡혔다.
- lock_mode X: exclusive 계열 record lock이다.
- locks rec but not gap: record는 잠겼지만 gap 삽입 보호까지는 아니다.
- heap no 2: page 안의 물리 record 번호다. 애플리케이션 PK와 같은 말이 아니다.

다른 예
RECORD LOCKS ... index `PRIMARY` ... lock_mode X
Record lock, heap no 1 PHYSICAL RECORD:
  0: len 8; hex 73757072656d756d; asc supremum;;

해석
- supremum은 page의 마지막보다 큰 가짜 record다.
- next-key/gap 범위의 끝을 설명할 때 등장할 수 있다.
- "row가 하나 더 있다"가 아니라 index interval을 읽어야 한다.
```

gap lock을 실무에서 가장 자주 밟는 곳은 “없는 것을 확인하고 넣는” 코드다. 예를 들어 쿠폰 코드가 없는지 확인한 뒤 insert하거나, 특정 구간에 예약이 없는지 확인한 뒤 insert하거나, 작업 큐에서 아직 처리되지 않은 row를 찾는 코드가 그렇다. 이때 plain SELECT는 과거 snapshot만 보여 줄 수 있고, locking read는 index gap을 잠가 새 삽입을 기다리게 할 수 있다. 그러나 gap lock이 넓어지면 unrelated insert까지 기다릴 수 있다. 그래서 InnoDB에서는 predicate에 맞는 index를 설계하는 일이 잠금 설계와 직접 연결된다.

```sql
CREATE TABLE coupon_claim (
    claim_id bigint PRIMARY KEY,
    coupon_code varchar(40) NOT NULL,
    user_id bigint NOT NULL,
    claimed_at timestamp NULL,
    KEY coupon_code_user_idx(coupon_code, user_id)
);

-- 같은 coupon_code/user_id 조합이 없는지 보고 넣는 약한 방식
SELECT count(*) FROM coupon_claim
 WHERE coupon_code = 'WELCOME'
   AND user_id = 1001;

-- InnoDB에서 범위를 잠그려는 방식
SELECT claim_id FROM coupon_claim
 WHERE coupon_code = 'WELCOME'
   AND user_id = 1001
 FOR UPDATE;

-- 더 강한 방식: 중복 자체를 key 충돌로 만든다.
CREATE UNIQUE INDEX coupon_claim_once
    ON coupon_claim(coupon_code, user_id);
```

```text
인덱스와 gap 범위

coupon_code_user_idx order:
  ('SPRING',  3)
  ('WELCOME', 1000)
  ('WELCOME', 1002)
  ('WINTER',  9)

WHERE coupon_code='WELCOME' AND user_id=1001 FOR UPDATE

검색 위치:
  ('WELCOME',1000) < ('WELCOME',1001) < ('WELCOME',1002)

잠금 목표:
  이미 있는 record가 아니라 ('WELCOME',1000)과 ('WELCOME',1002) 사이의 gap에
  같은 key가 삽입되는 것을 막는 것이 중요하다.

함정:
  coupon_code만 index에 있고 user_id가 뒤따르지 않으면 훨씬 넓은 WELCOME 범위를 잠글 수 있다.
  index가 없으면 scan과 lock 범위가 커져 unrelated user claim까지 느려질 수 있다.
```

PostgreSQL에서는 같은 업무 문제를 다른 식으로 풀 가능성이 높다. partial unique index나 exclusion constraint처럼 불변식을 제약으로 표현할 수 있으면 그것이 가장 명확하다. 명시적 row lock은 이미 있는 row를 보호하는 데 강하지만, 존재하지 않는 조합을 보호하려면 guard row를 만들거나 serializable retry를 선택하거나, 애초에 unique key 충돌로 문제를 바꾸는 편이 낫다. PostgreSQL에도 advisory lock이 있지만, advisory lock은 DB가 의미를 알지 못한다. 애플리케이션이 같은 key 산식과 같은 획득/해제 규칙을 일관되게 써야만 안전하다. 따라서 advisory lock은 편리한 escape hatch일 수 있지만, 제약조건을 대체하는 기본 장치로 남발하면 나중에 다른 코드 경로가 규칙을 우회한다.

```text
같은 "중복 쿠폰 claim 방지"의 PostgreSQL 후보

후보 A: unique constraint
  CREATE UNIQUE INDEX coupon_claim_once
      ON coupon_claim(coupon_code, user_id);
  장점: 모든 코드 경로에서 같은 key 충돌을 강제한다.
  단점: 실패를 사용자 메시지나 idempotent replay로 바꾸는 처리가 필요하다.

후보 B: guard row + SELECT FOR UPDATE
  coupon_guard(coupon_code, user_id)를 먼저 만들고 그 row를 잠근다.
  장점: 복잡한 업무 단위를 명시적 소유권으로 만들 수 있다.
  단점: guard row lifecycle, hot key, deadlock 순서를 관리해야 한다.

후보 C: pg_advisory_xact_lock(hash(coupon_code,user_id))
  장점: 별도 row 없이 transaction-level advisory lock을 쓸 수 있다.
  단점: DB constraint가 아니므로 모든 writer가 같은 함수를 호출해야 한다.

후보 D: SERIALIZABLE + retry
  장점: 읽기/쓰기 의존 관계를 DB가 감지할 수 있다.
  단점: retry-safe transaction boundary가 필요하고 실패율을 관측해야 한다.
```

락과 래치의 lower-layer 차이를 더 분명히 하려면 “누가 해제하는가”를 보면 된다. 트랜잭션 lock은 commit, rollback, savepoint rollback, statement 종료 같은 SQL 실행 경계와 연결된다. PostgreSQL 문서는 savepoint 이후 획득한 lock이 savepoint rollback 때 해제될 수 있다고 설명한다. InnoDB row/gap lock도 트랜잭션 경계와 연결된다. 반면 latch는 사용자 SQL이 rollback한다고 “업무 의미가 되돌아가서” 해제되는 것이 아니다. 내부 page를 고치기 위해 잡은 latch는 그 code path가 끝나면 풀린다. 사용자가 `ROLLBACK`을 해도 이미 지나간 latch critical section 자체는 과거 일이 된다. 이 차이가 보이면 관측과 대응이 갈린다.

| 질문 | lock이라면 | latch라면 |
|---|---|---|
| 누가 보유자인가 | transaction/session, sometimes statement | engine worker/thread |
| 무엇을 기다리나 | row/table/index range/advisory resource | buffer/page/hash bucket/internal structure |
| 오래 열릴 수 있나 | 긴 transaction 때문에 가능 | 보통 짧지만 hot path에서 반복 경합 |
| 앱 코드로 줄일 수 있나 | SQL 순서, transaction 길이, index, retry로 가능 | workload skew, schema/index pattern, engine setting/version까지 봄 |
| 실패 신호 | deadlock, lock timeout, serialization failure | high CPU, LWLock/mutex wait, throughput collapse |

현장에서 이 표는 장애 첫 10분을 아낀다. 예를 들어 `pg_stat_activity.wait_event_type = Lock`이면 트랜잭션 lock holder를 찾는 것이 우선이다. `wait_event_type = LWLock`이면 내부 경량 lock 계열이므로 row holder만 찾아서는 부족하다. MySQL에서도 `data_locks`와 `events_waits_summary_global_by_event_name`은 다른 질문에 답한다. 전자는 어떤 transaction lock이 어떤 record/gap/table을 잡았는지에 가깝고, 후자는 mutex/rwlock 같은 내부 wait를 볼 수 있다. 한 화면에 “lock”이라는 단어가 있어도 같은 계층이 아니다.

시니어 실무자는 lock 설계에서 “더 강하게 잠그면 안전하다”는 말도 경계한다. 강한 lock은 correctness를 쉽게 만들 수 있지만, 처리량과 tail latency를 크게 해칠 수 있다. 특히 API 서버는 DB connection을 제한된 pool로 사용한다. lock wait가 길어지면 대기 중인 트랜잭션만 느린 것이 아니라 connection pool이 고갈되어 관련 없는 요청까지 실패한다. 그래서 lock을 쓸 때는 세 가지 예산을 함께 정한다.

1. 보유 시간 예산

    lock을 잡은 뒤 외부 API, 사용자 입력, 파일 업로드, 네트워크 호출을 기다리지 않는다. 필요한 값은 lock 전에 준비하고, lock 안에서는 짧은 검증과 쓰기만 한다.

2. 범위 예산

    단일 row로 충분한지, range가 필요한지, guard row로 줄일 수 있는지 본다. InnoDB에서는 index가 range를 줄이는 핵심이고, PostgreSQL에서는 constraint나 materialized guard가 범위를 줄일 수 있다.

3. 실패 예산

    lock wait timeout, deadlock, serialization failure가 어느 정도까지 정상 재시도인지 정한다. 실패율이 높아지면 설계 문제가 있다는 신호로 본다.

이 절의 replay 질문은 다음과 같다. `SELECT * FROM orders WHERE status='READY' ORDER BY id LIMIT 1 FOR UPDATE`가 느릴 때, row lock 문제라고 단정해도 되는가? 답은 아니다. 먼저 어떤 DBMS인지, 어떤 index를 타는지, `SKIP LOCKED`를 쓰는지, 같은 READY row에 writer가 몰리는지, gap/next-key lock이 넓어진 것인지, PostgreSQL에서 tuple lock인지 table lock인지, 내부 latch/IO wait인지 확인해야 한다. 같은 SQL처럼 보여도 queue worker, 예약 시스템, 정산 배치에서는 보호해야 하는 자원과 실패 허용 방식이 다르다. lock 설계는 SQL 한 줄이 아니라 업무 불변식, index, 트랜잭션 길이, 관측 지표가 만나는 지점이다.

마지막으로 lock taxonomy를 실제 장애 메시지와 연결해 보자. “대기 중”이라는 증상 하나에도 여러 원인이 있다. 테이블 DDL이 `ACCESS EXCLUSIVE`를 잡아 plain SELECT까지 막는 경우, 한 row를 오래 잡아 UPDATE만 막는 경우, InnoDB range scan이 gap을 잡아 INSERT를 막는 경우, serializable 충돌 추적이 실패를 내는 경우, 내부 latch 경합으로 CPU가 타는 경우가 모두 다르다. 운영 대응은 이 분류를 빠르게 하는 데서 시작한다.

```text
증상에서 lock 종류로 좁히는 현장 trace

증상: SELECT도 멈춘다.
  PostgreSQL 후보: ACCESS EXCLUSIVE table lock, DDL, VACUUM FULL, TRUNCATE
  확인: pg_locks relation lock mode, pg_stat_activity query
  잘못된 조치: row holder만 찾기

증상: UPDATE만 멈춘다.
  후보: row-level lock, transactionid wait, InnoDB record lock
  확인: blocked writer, holder transaction, row/index key
  잘못된 조치: SELECT query cache나 buffer pool만 보기

증상: INSERT가 특정 범위에서 멈춘다.
  InnoDB 후보: gap/next-key lock
  확인: index range, SHOW ENGINE INNODB STATUS, data_locks
  잘못된 조치: "해당 row가 없는데 왜 잠기지?"라고 끝내기

증상: CPU가 높고 lock holder가 명확하지 않다.
  후보: latch/mutex/LWLock/IO wait/hot page
  확인: wait_event_type, performance_schema waits, engine metrics
  잘못된 조치: isolation level만 낮추기
```

이 trace는 문서용 장식이 아니라 장애 대응 순서다. lock을 제대로 읽는 팀은 먼저 “무엇이 막혔는가”를 좁히고, 그다음 “누가 들고 있는가”, “왜 오래 들고 있는가”, “범위가 왜 넓어졌는가”로 내려간다. 반대로 lock이라는 단어만 보고 transaction timeout, DB parameter, pool size를 동시에 바꾸면 재현성이 사라진다. 특히 pool size를 늘리는 조치는 lock wait를 해결하지 못하고 대기열을 키울 수 있다. holder가 길면 더 많은 connection이 같은 holder 뒤에 붙고, DB와 애플리케이션 모두 더 늦어진다.

PostgreSQL과 InnoDB를 같이 쓰는 조직에서는 용어표도 필요하다.

| 사람이 말한 표현 | PostgreSQL에서 먼저 볼 곳 | InnoDB에서 먼저 볼 곳 | 주의할 오해 |
|---|---|---|---|
| 테이블이 잠겼다 | `pg_locks` relation mode | metadata/table lock, intention lock | intention lock을 전체 차단으로 오해 |
| 행이 잠겼다 | row lock/transactionid wait | record lock on index record | InnoDB record lock은 index record 기준 |
| 범위가 잠겼다 | serializable predicate 관찰, explicit strategy | gap/next-key lock | 두 구현을 같은 장치로 오해 |
| 내부 락이 높다 | LWLock, BufferPin, wait event | mutex/rwlock wait, InnoDB monitor | 트랜잭션 lock과 latch를 혼동 |
| SKIP LOCKED 쓰면 된다 | queue semantics, starvation 검토 | gap/next-key와 worker fairness 검토 | 누락/기아/순서 보장 문제 무시 |

`SKIP LOCKED`도 신중해야 한다. 작업 큐에서 이미 다른 worker가 잡은 row를 건너뛰는 것은 처리량을 높일 수 있지만, 항상 공정한 순서를 보장하지 않는다. 오래 걸리는 row가 계속 건너뛰어지고 뒤 작업만 처리되면 starvation이 생길 수 있다. 또한 잠긴 row를 건너뛰었다는 것은 “처리할 일이 없다”가 아니라 “지금 이 worker가 잡을 수 있는 일이 없다”에 가깝다. queue 설계에서는 claim timestamp, attempt count, visibility timeout, stuck job recovery를 함께 둔다. 이 내용은 lock DU에서 너무 멀어 보이지만, 실제로는 `FOR UPDATE SKIP LOCKED`를 쓰는 순간 바로 운영 문제가 된다.

```sql
-- queue claim 예시: 간단하지만 이 자체가 완전한 운영 설계는 아니다.
WITH picked AS (
    SELECT job_id
      FROM job_queue
     WHERE status = 'READY'
     ORDER BY priority DESC, job_id
     FOR UPDATE SKIP LOCKED
     LIMIT 10
)
UPDATE job_queue q
   SET status = 'RUNNING',
       claimed_at = now()
  FROM picked
 WHERE q.job_id = picked.job_id
RETURNING q.job_id;

-- 관측 질문
-- 1. READY인데 오래 남은 job이 있는가?
-- 2. RUNNING에서 stuck 된 job을 되돌리는 정책이 있는가?
-- 3. priority가 낮은 job이 영구히 밀리지 않는가?
-- 4. 같은 index order가 worker 사이 deadlock을 줄이는가?
```

이처럼 lock은 단순히 막는 장치가 아니라 작업 분배, 실패 복구, 공정성까지 영향을 준다. 그러므로 이 DU의 핵심 replay는 “이 SQL이 어떤 lock을 잡는가”에서 끝나지 않는다. “그 lock이 어떤 업무 자원을 보호하고, 기다리는 동안 어떤 connection과 thread를 붙잡고, 실패하면 어떤 retry/복구 경로를 타는가”까지 이어져야 한다.

마지막으로 문서나 코드 리뷰에서 lock을 설명할 때는 네 단어를 함께 적는 습관을 들인다. `resource`는 보호 대상이고, `mode`는 공유/배타/범위 같은 충돌 규칙이며, `duration`은 언제 풀리는지이고, `observer`는 어디서 확인할지다. 예를 들어 “주문 row를 FOR UPDATE로 잠근다”보다 “resource=orders.id, mode=row exclusive, duration=transaction end, observer=pg_locks/blocked UPDATE”라고 적으면 설계가 훨씬 선명하다. InnoDB range라면 “resource=secondary index interval, mode=next-key, duration=transaction end, observer=SHOW ENGINE INNODB STATUS/data_locks”처럼 쓴다. 이 네 칸을 채우지 못하면 아직 lock을 이해한 것이 아니라 lock이라는 단어만 붙인 것이다.

이 네 칸은 리뷰 질문으로도 좋다. `resource`가 비어 있으면 무엇을 보호하는지 모르는 것이고, `mode`가 비어 있으면 무엇과 충돌하는지 모르는 것이며, `duration`이 비어 있으면 대기 비용을 모르는 것이고, `observer`가 비어 있으면 운영에서 검증할 방법이 없는 것이다. lock 설계는 이 네 질문에 답할 때 비로소 코드가 아니라 운영 가능한 계약이 된다.

그래서 lock 관련 PR을 볼 때는 SQL diff만 보지 않는다. 인덱스 diff, 트랜잭션 경계, timeout 설정, retry 정책, metric 이름까지 같이 본다. lock은 한 줄의 SQL 기능이 아니라 동시 실행 중 어떤 요청을 기다리게 만들지 정하는 시스템 계약이기 때문이다.

이 관점이 있어야 lock을 성능 문제와 정합성 문제의 교차점으로 다룰 수 있다.

그 교차점을 보지 못하면 튜닝과 정합성 보강이 서로를 망친다.

항상 같이 본다.

### 마지막 실무 연결: 잠금은 범위와 수명을 함께 읽어야 한다

lock을 이해할 때는 무엇을 잠갔는가와 얼마나 오래 잠갔는가를 같이 봐야 한다. row lock은 commit까지 이어질 수 있고, table lock은 DDL과 함께 읽기까지 막을 수 있으며, gap/next-key lock은 존재하지 않는 값이 들어올 공간을 막을 수 있다. latch는 내부 자료구조 보호를 위해 훨씬 짧게 잡히지만, 높은 경합에서는 CPU와 latency로 드러날 수 있다.

운영에서 lock wait을 보면 먼저 SQL 하나만 보지 말고 transaction 전체를 본다. 어떤 순서로 자원을 잡았는지, 사용자 응답 생성이나 외부 호출이 transaction 안에 들어갔는지, index 조건이 너무 넓어 gap을 크게 잡았는지 확인한다. 잠금 장애는 대개 한 줄의 UPDATE가 아니라 그 UPDATE를 둘러싼 transaction 수명에서 나온다.

### 추가 판정 질문: 잠금은 데이터의 의미를 지키는가, 내부 구조를 지키는가

lock과 latch를 구분하는 이유는 조치가 다르기 때문이다. 트랜잭션 lock이 문제라면 긴 transaction, 자원 획득 순서, index range, DDL lock을 본다. latch나 내부 경합이 문제라면 hot page, buffer manager, B-tree split, CPU spin 같은 다른 층을 본다. 둘을 섞으면 개발자는 쿼리 순서를 고쳐야 할 문제에 서버 증설을 하고, 내부 경합 문제에 업무 retry만 늘리는 식으로 엇나간다.

운영 로그에는 가능한 한 wait 이름, 대상 relation/index, transaction age, blocker query를 함께 남겨야 한다. 이 네 가지가 있으면 같은 “잠금”이라는 단어도 row conflict인지, range 보호인지, DDL 충돌인지, 내부 경합인지 좁혀갈 수 있다.


## deadlock, wait graph, timeout, retry

데드락(deadlock)은 여러 트랜잭션이 서로가 가진 lock을 기다려 더 이상 진행할 수 없는 상태다. 이 절을 읽은 뒤에는 lock timeout과 deadlock을 같은 장애로 처리하면 왜 위험한지 설명할 수 있어야 한다. timeout은 “너무 오래 기다렸다”는 시간 기반 신호이고, deadlock은 “기다림 그래프에 cycle이 생겨 기다려도 풀리지 않는다”는 구조 기반 신호다. PostgreSQL 공식 문서는 deadlock을 자동 감지해 관련 트랜잭션 중 하나를 abort한다고 설명하고, 여러 object를 잠글 때 일관된 순서로 잠그는 것이 일반적인 방어라고 말한다. MySQL InnoDB 공식 문서도 deadlock 발생 시 하나의 victim transaction을 rollback하며, 마지막 deadlock은 `SHOW ENGINE INNODB STATUS`로 볼 수 있고, 자주 발생하면 모든 deadlock을 error log에 출력하는 설정을 검토할 수 있다고 설명한다.

데드락 처리가 별도 주제로 등장한 배경은 DB가 단순히 “기다리면 언젠가 풀릴 잠금”만 다루지 않기 때문이다. 초창기 단일 사용자식 파일 갱신에서는 한 작업이 끝날 때까지 다른 작업을 막으면 충분했지만, 다중 사용자 DB에서는 수천 개 트랜잭션이 서로 다른 순서로 계좌, 주문, 재고, 인덱스 범위를 잡는다. 이때 A가 B를 기다리고 B가 다시 A를 기다리는 cycle은 시간이 지나도 자연스럽게 풀리지 않는다. 그래서 DB 엔진은 wait graph를 관찰하고, cycle을 발견하면 누군가를 희생시켜 전체 시스템을 앞으로 진행시킨다. 이 역사는 실무 retry 정책에도 그대로 이어진다. deadlock victim은 “데이터가 틀렸다”가 아니라 “DB가 교착을 끊기 위해 정상적으로 한쪽을 되돌렸다”는 신호일 수 있으므로, 애플리케이션은 무조건 장애 화면으로 밀어내기보다 idempotency, 재시도 가능성, 사용자에게 이미 보인 부작용을 함께 판단해야 한다.

첫 번째 벽돌은 계좌 이체의 반대 순서 업데이트다.

```sql
CREATE TABLE account_pair (
    account_id integer PRIMARY KEY,
    balance integer NOT NULL
);

INSERT INTO account_pair VALUES (1, 100), (2, 100);
```

```text
deadlock schedule

time | session A                                      | session B
-----+------------------------------------------------+------------------------------------------
t1   | BEGIN;                                         | BEGIN;
t2   | UPDATE account_pair                            |
     |    SET balance = balance - 10                  |
     |  WHERE account_id = 1;                         |
t3   |                                                | UPDATE account_pair
     |                                                |    SET balance = balance - 20
     |                                                |  WHERE account_id = 2;
t4   | UPDATE account_pair                            |
     |    SET balance = balance + 10                  |
     |  WHERE account_id = 2; -- waits B              |
t5   |                                                | UPDATE account_pair
     |                                                |    SET balance = balance + 20
     |                                                |  WHERE account_id = 1; -- waits A

wait graph

  A holds row(1)
  A waits for row(2), held by B

  B holds row(2)
  B waits for row(1), held by A

  A -> B -> A
  cycle detected: one transaction must be aborted
```

이 trace의 핵심은 기다림이 길어서 문제가 아니라, 기다려도 자연스럽게 풀릴 수 없는 cycle이 생겼다는 점이다. A는 B가 row(2)를 놓아야 진행하고, B는 A가 row(1)을 놓아야 진행한다. 둘 다 상대가 끝나야 끝날 수 있으므로 시스템이 개입해야 한다. DB는 한쪽을 victim으로 골라 rollback시키고, 다른 쪽이 진행하게 한다. 어떤 쪽이 victim이 될지는 애플리케이션이 의존하면 안 된다. 따라서 deadlock 처리는 “A가 항상 진다” 같은 가정이 아니라, 어느 쪽이든 실패할 수 있고 전체 트랜잭션을 재시도할 수 있다는 구조로 만들어야 한다.

lock timeout은 다르다.

```text
lock wait timeout schedule

time | session A                                      | session B
-----+------------------------------------------------+------------------------------------------
t1   | BEGIN;                                         |
t2   | UPDATE account_pair SET balance=90             |
     |  WHERE account_id=1;                           |
t3   | -- 외부 API 응답을 기다리며 60초 동안 열림     |
t4   |                                                | BEGIN;
t5   |                                                | UPDATE account_pair SET balance=80
     |                                                |  WHERE account_id=1; -- waits A
t6   |                                                | lock wait timeout 발생 가능
t7   | COMMIT;                                        |

wait graph
B -> A
cycle 없음

의미
A가 끝나면 B는 진행할 수 있었다.
문제는 cycle이 아니라 대기 시간이 운영 한계를 넘은 것이다.
```

timeout은 구조적으로 풀릴 수 있는 기다림이 너무 길어진 상태일 수 있다. 원인은 긴 트랜잭션, 외부 API 호출, 사용자 입력 대기, 느린 쿼리, 배치 작업, 누락된 인덱스, 대량 업데이트일 수 있다. 이때 단순 retry를 걸면 같은 긴 holder 뒤에 다시 줄을 서서 부하를 키울 수 있다. 반대로 deadlock은 같은 순서 충돌이 반복되면 retry해도 다시 터질 수 있지만, 짧은 random backoff와 lock acquisition order 정렬로 줄일 수 있다. timeout과 deadlock을 같은 catch 블록에서 “그냥 한 번 더 실행”으로 처리하면 운영에서 재시도 폭풍이 생긴다.

데드락을 줄이는 가장 강한 기본기는 lock 순서 고정이다. 여러 row나 여러 table을 잠가야 한다면 모든 코드 경로가 같은 기준으로 정렬한 뒤 잠근다.

```sql
-- 나쁜 패턴: 호출자가 준 순서대로 잠근다.
-- transfer(from=2, to=1)와 transfer(from=1, to=2)가 동시에 오면 순서가 엇갈린다.

BEGIN;
SELECT * FROM account_pair WHERE account_id = :from_id FOR UPDATE;
SELECT * FROM account_pair WHERE account_id = :to_id FOR UPDATE;
UPDATE account_pair SET balance = balance - :amount WHERE account_id = :from_id;
UPDATE account_pair SET balance = balance + :amount WHERE account_id = :to_id;
COMMIT;

-- 더 강한 패턴: 잠금 획득 순서를 account_id 오름차순으로 고정한다.

BEGIN;
SELECT * FROM account_pair
 WHERE account_id IN (:from_id, :to_id)
 ORDER BY account_id
 FOR UPDATE;
UPDATE account_pair SET balance = balance - :amount WHERE account_id = :from_id;
UPDATE account_pair SET balance = balance + :amount WHERE account_id = :to_id;
COMMIT;
```

주의할 점은 SQL의 `ORDER BY`가 언제나 lock 획득 순서를 원하는 대로 보장한다고 단순화하지 않는 것이다. DBMS와 plan에 따라 lock을 잡는 물리 순서는 달라질 수 있다. 그래도 애플리케이션이 먼저 id 목록을 정렬하고, 가능한 한 단일 statement나 명확한 index order로 접근하며, plan을 확인하는 것은 데드락 확률을 줄이는 핵심 습관이다. 더 중요한 것은 모든 코드 경로가 같은 규칙을 쓰는 것이다. 한 경로만 정렬하고 다른 경로가 역순으로 업데이트하면 deadlock 가능성은 남는다.

운영에서 deadlock을 보았을 때 확인할 것은 “누가 나빴나”가 아니라 “cycle이 어떤 resource 순서에서 생겼나”다.

```text
deadlock report 읽기 checklist

1. victim transaction
   - 어떤 SQL이 실패했는가
   - 애플리케이션은 retry했는가

2. held resource
   - 이미 잡은 row/table/index range는 무엇인가
   - 해당 resource를 잡은 앞 SQL은 무엇인가

3. waited resource
   - 기다리던 row/table/index range는 무엇인가
   - 상대 transaction이 왜 그것을 잡고 있었는가

4. acquisition order
   - 두 transaction이 같은 resource set을 다른 순서로 잡았는가
   - 인덱스가 없어 range scan이 넓어졌는가

5. fix path
   - 순서 고정
   - transaction 축소
   - 필요한 index 추가
   - lock scope 축소
   - retry/backoff와 idempotency 보강
```

PostgreSQL에서는 `pg_locks`와 `pg_stat_activity`를 조합해 현재 wait를 볼 수 있고, deadlock 발생 시 로그에 관련 문맥이 남는다. InnoDB에서는 `SHOW ENGINE INNODB STATUS`가 마지막 deadlock 정보를 보여 주며, `innodb_print_all_deadlocks`를 켜면 모든 deadlock을 error log에 남겨 반복 패턴을 볼 수 있다. MySQL 공식 문서는 deadlock 가능성이 isolation level 자체에 의해 없어지는 것이 아니라, write operation의 lock 순서와 timing 때문에 생긴다고 설명한다. 따라서 isolation level을 낮추는 것만으로 deadlock을 해결하려 하면 원인과 맞지 않을 수 있다.

실무 함정은 retry에도 있다. deadlock victim은 보통 retry 가능하지만, 트랜잭션 안에 외부 side effect가 있으면 재시도가 위험하다.

```text
위험한 retry

BEGIN;
UPDATE account_pair SET balance = balance - 100 WHERE account_id = 1;
CALL external_payment_api(); -- 카드 승인, 송금 요청, 메시지 발행
UPDATE account_pair SET balance = balance + 100 WHERE account_id = 2;
COMMIT; -- 여기서 deadlock victim이면?

문제
- DB transaction은 rollback될 수 있다.
- 외부 API 호출은 rollback되지 않을 수 있다.
- retry하면 외부 side effect가 중복될 수 있다.

수리 방향
- 외부 side effect를 DB transaction 밖으로 빼거나 outbox로 지연한다.
- idempotency key를 사용한다.
- retry 가능한 단위와 재실행하면 안 되는 단위를 분리한다.
```

따라서 deadlock 대응은 DB만의 문제가 아니다. 애플리케이션 transaction boundary, 외부 호출, idempotency, outbox, 로그 상관관계가 함께 필요하다. DB가 victim을 고르는 순간 애플리케이션은 “이 요청을 어디서부터 다시 실행할 수 있는가”를 알고 있어야 한다. 이 경계가 없으면 deadlock 자체는 정상적으로 복구 가능한 동시성 현상인데, 서비스는 중복 결제나 중복 메시지 같은 더 큰 장애로 번진다.

local lab seed는 다음처럼 구성할 수 있다.

```text
database/deep-dive/labs/transactions/du33-deadlock-wait-graph/
  README.md
  postgres/
    setup.sql
    session-a.sql
    session-b.sql
    observe.sql
  mysql/
    setup.sql
    session-a.sql
    session-b.sql
    observe.sql
  expected/
    wait-graph.md
    retry-policy.md

expected PASS
- 반대 순서 UPDATE가 deadlock 또는 one-side abort를 만든다.
- 같은 순서 UPDATE 버전에서는 deadlock이 재현되지 않는다.
- lock timeout 실험은 cycle 없는 단방향 wait로 분리된다.
- retry 예시는 외부 side effect 없는 순수 DB transaction에만 적용된다.

expected FAIL
- timeout과 deadlock을 같은 expected file에 섞는다.
- victim이 항상 session B라고 가정한다.
- deadlock 재현 뒤 순서 고정 repair를 검증하지 않는다.
```

마지막으로 데드락을 줄이는 설계 규칙을 다시 정리한다. 트랜잭션은 짧게 유지한다. 여러 자원을 잡을 때는 모든 코드 경로에서 같은 순서를 쓴다. 범위 업데이트와 locking read에는 적절한 인덱스를 둔다. 실패는 `deadlock`, `lock wait timeout`, `serialization failure`, `duplicate key`로 분류한다. deadlock과 serialization failure는 재시도 후보지만, 재시도 단위가 안전해야 한다. lock wait timeout은 holder를 줄이고 SQL/인덱스/트랜잭션 길이를 고치는 쪽이 먼저다. 이 차이를 지키면 deadlock은 “무서운 랜덤 장애”가 아니라, wait graph로 설명하고 재현하고 줄일 수 있는 동시성 현상이 된다.

데드락을 wait graph로 그리는 습관은 장애 분석의 품질을 바꾼다. 로그에는 SQL 조각, transaction id, index record, lock mode가 흩어져 나오지만, 사람이 결론을 내려야 하는 것은 그래프다. 노드는 transaction이고, 화살표는 “왼쪽 transaction이 오른쪽 transaction이 가진 resource를 기다린다”는 뜻이다. cycle이 있으면 deadlock이다. cycle이 없으면 long wait, starvation, lock convoy, connection pool starvation 같은 다른 문제일 수 있다.

```text
wait graph notation

T1 --waits-for--> T2
T2 --waits-for--> T3
T3 --waits-for--> T1

cycle:
T1 -> T2 -> T3 -> T1

deadlock detector가 해야 할 일:
1. cycle을 찾는다.
2. victim을 하나 고른다.
3. victim transaction을 abort/rollback한다.
4. 나머지 transaction이 진행할 수 있게 한다.

애플리케이션이 해야 할 일:
1. victim이 될 수 있음을 정상 실패로 분류한다.
2. transaction 전체를 안전하게 다시 실행한다.
3. 반복되면 lock 획득 순서나 범위를 고친다.
```

세 개 이상의 transaction이 만드는 deadlock도 같은 방식으로 읽는다.

```text
three-way deadlock

T1 holds row(A), waits row(B)
T2 holds row(B), waits row(C)
T3 holds row(C), waits row(A)

그래프
T1 -> T2 -> T3 -> T1

겉보기 증상
- 각 세션은 "한 row만 기다리는 것"처럼 보인다.
- 전체로 보면 cycle이다.
- victim은 T1/T2/T3 중 어느 쪽이든 될 수 있다.
```

이 예제는 “우리 코드는 두 row만 만지니 복잡한 deadlock은 없다”는 오해를 고친다. 각 코드 경로는 단순해도, 업무 흐름이 세 개 이상 겹치면 cycle은 쉽게 커진다. 주문 상태 업데이트가 주문 row를 잡고 재고 row를 기다리고, 재고 보정 배치가 재고 row를 잡고 정산 row를 기다리고, 정산 배치가 정산 row를 잡고 주문 row를 기다리면 세 방향 cycle이 된다. 따라서 deadlock 분석은 단일 SQL이 아니라 업무 경로의 resource acquisition order를 본다.

운영 대응을 retry 중심으로만 두면 또 다른 함정이 생긴다. retry는 victim transaction을 다시 실행해도 같은 입력과 같은 외부 효과가 안전할 때만 맞다. 다음 표처럼 실패 종류별 첫 대응을 나누어야 한다.

| 실패 신호 | 구조 | 즉시 retry 적합성 | 먼저 볼 것 | 잘못된 대응 |
|---|---|---:|---|---|
| deadlock victim | wait graph cycle | 조건부 적합 | lock 순서, transaction 크기, index | 무한 즉시 retry |
| lock wait timeout | cycle 없음, 긴 holder | 낮음 또는 조건부 | holder SQL, 외부 대기, long transaction | timeout만 늘리기 |
| serialization failure | 직렬화 불가능성 감지 | 적합한 경우 많음 | retry-safe boundary, conflict rate | 사용자 오류로만 처리 |
| duplicate key | constraint 충돌 | 업무에 따라 다름 | idempotency, unique design | isolation level 올리기 |
| connection timeout | pool 고갈 가능 | 부적합 | lock wait 누적, slow query, pool size | pool만 늘리기 |

특히 lock wait timeout을 deadlock처럼 재시도하면 증상이 커질 수 있다. holder가 30초 동안 외부 API를 기다리는 구조라면, retry는 같은 holder 뒤에 새 요청을 더 세운다. pool이 차고, API 서버 thread가 막히고, 사용자는 전체 서비스 장애를 본다. 이때는 timeout 값을 늘리기보다 holder transaction을 줄이고, 외부 호출을 transaction 밖으로 빼고, 필요한 row를 더 늦게 잠그고 더 빨리 commit하는 방향이 맞다.

반대로 deadlock은 timeout을 늘린다고 해결되지 않는다. cycle은 시간이 지나도 자연스럽게 풀리지 않는다. deadlock detector가 victim을 고르거나, detector가 꺼져 있으면 lock wait timeout 같은 시간 기반 장치가 결국 한쪽을 깨야 한다. MySQL 공식 문서는 deadlock detection이 기본적으로 켜져 있고, 꺼져 있으면 `innodb_lock_wait_timeout`에 의존한다고 설명한다. 그래서 InnoDB 운영에서는 deadlock detector 비용과 timeout 정책을 모두 이해해야 한다. 고동시성 hot row 환경에서는 detector 자체가 부담이 될 수 있지만, detector를 끄면 cycle 해결이 느려질 수 있다. 이 선택은 단순 튜닝이 아니라 workload와 장애 허용 시간의 trade-off다.

데드락을 줄이는 repair는 보통 네 층이다.

1. resource order repair

    모든 코드 경로가 같은 순서로 row/table을 잡게 한다. 계좌 이체라면 `min(account_id)`부터 잠근다. 주문/재고/정산처럼 table이 여러 개면 `order -> inventory -> settlement` 같은 순서를 문서와 helper로 고정한다. 이 규칙은 사람 기억에 맡기지 말고 repository-level helper나 service policy로 드러내는 편이 좋다.

2. scope repair

    불필요하게 넓은 range scan을 줄인다. InnoDB에서는 `UPDATE ... WHERE status='READY'`처럼 낮은 선택도의 조건이 많은 next-key/record lock을 만들 수 있다. index를 추가하거나, 작업 단위를 shard key로 나누거나, `SKIP LOCKED` queue pattern을 검토한다. PostgreSQL에서도 대량 UPDATE가 많은 row lock과 vacuum pressure를 만들 수 있으므로 batch 크기를 조정한다.

3. duration repair

    lock을 잡은 뒤 수행하는 일을 줄인다. 검증용 SELECT, 외부 API, JSON 직렬화, 파일 생성, remote call이 lock 안에 들어 있으면 holder 시간이 길어진다. transaction boundary를 줄이면 deadlock뿐 아니라 timeout과 pool 고갈도 함께 줄어든다.

4. failure contract repair

    실패를 분류하고 재시도 가능한 단위를 만든다. SQLSTATE, MySQL error number, vendor exception class를 application error taxonomy로 올린다. retry에는 max attempt, backoff, jitter, idempotency key, observability tag가 있어야 한다.

아래는 계좌 이체 deadlock을 줄이는 좀 더 실제적인 pseudo-code다. 핵심은 정렬된 lock acquisition과 외부 side effect 분리다.

```text
transfer(fromId, toId, amount, requestId)

outside transaction
  validate request shape
  check idempotency requestId already completed?
  lockOrder = sort([fromId, toId])

inside transaction
  claim idempotency key or load existing result
  SELECT * FROM account_pair
   WHERE account_id IN (lockOrder[0], lockOrder[1])
   ORDER BY account_id
   FOR UPDATE
  validate balance
  UPDATE debit row
  UPDATE credit row
  INSERT ledger entries
  INSERT outbox event
  COMMIT

after commit
  publish outbox asynchronously
  return committed result

retry rule
  deadlock/serialization failure before commit -> retry whole DB transaction
  duplicate idempotency key -> replay stored result
  external publish failure -> outbox worker retry, not transfer transaction retry
```

이 pseudo-code는 deadlock을 절대 없앤다고 주장하지 않는다. 대신 deadlock이 나도 같은 request를 다시 실행할 수 있는 경계를 만든다. ledger insert와 outbox insert가 같은 DB transaction에 들어 있으므로 commit 전 실패는 사라지고, commit 후 publish 실패는 outbox worker가 처리한다. 외부 side effect가 transaction 안에 없기 때문에 victim rollback 뒤 retry가 상대적으로 안전하다. 이 구조는 나중의 outbox/saga DU와 연결되지만, deadlock retry를 설명할 때 이미 필요하다.

관측 SQL도 repair 검증에 포함해야 한다.

```sql
-- PostgreSQL: 누가 누구를 기다리는지 보기 위한 출발점
SELECT
    blocked.pid AS blocked_pid,
    blocked.query AS blocked_query,
    blocker.pid AS blocker_pid,
    blocker.query AS blocker_query
FROM pg_stat_activity blocked
JOIN pg_locks blocked_locks
  ON blocked_locks.pid = blocked.pid
 AND NOT blocked_locks.granted
JOIN pg_locks blocker_locks
  ON blocker_locks.locktype = blocked_locks.locktype
 AND blocker_locks.database IS NOT DISTINCT FROM blocked_locks.database
 AND blocker_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
 AND blocker_locks.page IS NOT DISTINCT FROM blocked_locks.page
 AND blocker_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
 AND blocker_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
 AND blocker_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
 AND blocker_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
 AND blocker_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
 AND blocker_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
 AND blocker_locks.pid <> blocked_locks.pid
 AND blocker_locks.granted
JOIN pg_stat_activity blocker
  ON blocker.pid = blocker_locks.pid;
```

```sql
-- MySQL/InnoDB: performance_schema가 켜져 있을 때 wait edge 보기
SELECT
    waiting_engine_transaction_id,
    blocking_engine_transaction_id,
    object_schema,
    object_name,
    index_name,
    lock_type,
    lock_mode,
    lock_status,
    lock_data
FROM performance_schema.data_lock_waits w
JOIN performance_schema.data_locks l
  ON w.requesting_engine_lock_id = l.engine_lock_id;
```

이 쿼리들은 운영에 그대로 붙여 넣기 전에 권한, 버전, extension, performance_schema 설정을 확인해야 한다. 여기서 중요한 것은 쿼리 자체보다 관측 모델이다. wait edge를 찾아 graph로 그리고, cycle인지 단방향 wait인지 분리한 뒤, SQL과 resource order로 돌아가야 한다. “DB가 deadlock이라고 했다”에서 멈추면 다음 장애 때 같은 현상이 반복된다.

마지막으로 팀에 남길 runbook은 짧고 단단해야 한다.

```text
deadlock/timeout runbook

1. classify
   - deadlock victim?
   - lock wait timeout?
   - serialization failure?
   - duplicate key?

2. capture
   - failed SQL and transaction boundary
   - holder/waiter SQL
   - locked table/index/row/range
   - isolation level
   - retry attempt count and request id

3. draw
   - wait graph node = transaction
   - edge = waits-for
   - cycle exists? yes/no

4. repair
   - consistent acquisition order
   - narrower index/range
   - shorter transaction
   - constraint/idempotency/outbox
   - bounded retry

5. verify
   - reproduce before
   - apply repair
   - reproduce after
   - confirm lower deadlock rate and no duplicate side effect
```

이 runbook의 PASS 신호는 deadlock을 0으로 만드는 것이 아니다. 모든 deadlock을 완전히 없애는 것은 현실적인 목표가 아닐 수 있다. PASS는 deadlock이 발생해도 request id로 추적되고, victim transaction이 안전하게 retry되며, 반복 패턴은 acquisition order나 index/range repair로 줄어드는 것이다. FAIL은 deadlock을 단순히 “DB 일시 오류”로만 기록하고, 사용자에게 500을 돌려주거나, 외부 결제를 중복 호출하거나, timeout 값을 늘리는 것으로 문제를 숨기는 것이다.

데드락 학습의 마지막 단계는 repair 전후를 같은 실험으로 비교하는 것이다. deadlock을 한 번 재현하고 끝내면 “무서운 현상”만 기억에 남는다. 같은 데이터와 같은 두 세션에서 순서 고정 repair를 적용한 뒤 더 이상 cycle이 생기지 않는 것을 봐야 설계 원리가 남는다.

```text
before: 서로 다른 순서

session A order: row(1) -> row(2)
session B order: row(2) -> row(1)

wait graph:
A holds 1, waits 2
B holds 2, waits 1
A -> B -> A

after: 같은 순서

session A order: row(1) -> row(2)
session B order: row(1) -> row(2)

timeline:
A locks row(1)
B waits row(1)
A locks row(2)
A commits
B locks row(1)
B locks row(2)
B commits

결과:
대기는 있을 수 있지만 cycle은 없다.
```

이 비교에서 중요한 말은 “대기는 남을 수 있다”다. deadlock repair는 모든 wait를 없애는 작업이 아니다. 같은 순서로 잠그면 B가 A를 기다릴 수 있지만, B가 이미 A가 기다리는 다른 resource를 들고 있지 않으므로 cycle이 없다. 따라서 lock wait 시간이 여전히 길다면 다음 단계는 deadlock repair가 아니라 duration repair다. transaction 안의 외부 호출, 대량 scan, 느린 index, 불필요한 batch 크기를 줄여야 한다. 이처럼 deadlock과 timeout을 분리하면 다음 조치가 선명해진다.

실패 분류는 코드에도 남아야 한다. Java/Spring 애플리케이션이라면 vendor exception을 모두 `RuntimeException`으로만 처리하지 말고, SQLSTATE와 vendor code를 보고 재시도 가능한 동시성 실패를 분리한다. PostgreSQL의 deadlock은 보통 SQLSTATE `40P01`, serialization failure는 `40001`로 올라올 수 있다. MySQL deadlock은 error 1213, lock wait timeout은 1205로 알려져 있다. 실제 driver와 framework가 어떤 예외로 감싸는지는 로컬 테스트로 확인해야 하지만, 분류 축 자체는 코드에 있어야 한다.

```text
application error taxonomy sketch

ConcurrencyRetryable
  - PostgreSQL serialization failure 40001
  - PostgreSQL deadlock detected 40P01
  - MySQL deadlock 1213

ConcurrencyMaybeRetryableAfterInspection
  - MySQL lock wait timeout 1205
  - application-side transaction timeout

BusinessConflict
  - duplicate key for idempotency winner
  - optimistic lock affected_rows=0

NotRetryableByDefault
  - syntax error
  - permission denied
  - missing table

정책:
deadlock은 bounded retry 후보이지만, lock wait timeout은 holder를 확인하지 않고 무조건 retry하지 않는다.
duplicate key는 실패가 아니라 replay 또는 이미 처리됨으로 바뀔 수 있다.
```

이 taxonomy가 없으면 장애 대응은 로그 문자열 검색에 의존한다. 문자열은 locale, DB 버전, driver, framework에 따라 바뀔 수 있다. 가능한 한 SQLSTATE, vendor error number, typed exception, affected row count 같은 구조화된 신호를 남긴다. 재시도 로그에는 attempt number, request id, transaction name, locked resource hint를 넣는다. 그래야 retry가 실제로 문제를 줄이는지, 아니면 같은 lock wait를 반복해 pool을 태우는지 관측할 수 있다.

마지막 실무 함정은 deadlock을 “DB가 알아서 rollback했으니 끝”이라고 생각하는 것이다. DB가 victim transaction을 rollback해도 애플리케이션의 in-memory state, 외부 API 호출, 캐시 갱신, 메시지 발행은 자동으로 되돌아가지 않을 수 있다. 따라서 transaction 안에서 무엇이 DB와 함께 rollback되는지, 무엇이 DB 밖에 이미 나갔는지 구분해야 한다. 좋은 설계는 DB transaction 안에서 durable intent를 남기고, 외부 side effect는 commit 이후 outbox worker가 처리하게 만든다. 그러면 deadlock victim이 되어도 외부 세계에는 아직 아무 일도 나가지 않았으므로 retry가 안전해진다.

```text
deadlock-safe side effect boundary

inside DB transaction
  - validate
  - lock rows in canonical order
  - update balances/statuses
  - insert ledger
  - insert outbox(event_id, payload, status='READY')

commit succeeds
  - outbox worker publishes event
  - publish failure retries by event_id

deadlock victim before commit
  - all DB changes rollback
  - outbox row does not exist
  - external publish never happened
  - request can retry DB transaction

deadlock victim after external call inside transaction
  - DB rollback
  - external call may already have happened
  - retry can duplicate side effect
  - this is the design smell to remove
```

이 경계까지 설명해야 deadlock retry가 실무적으로 닫힌다. DB 관점에서는 victim rollback이 충분하지만, 서비스 관점에서는 “재실행해도 같은 결과가 한 번만 외부에 나타나는가”가 더 중요하다. 그래서 deadlock 문서는 lock manager에서 끝나지 않고 idempotency, outbox, request tracing으로 이어진다. 이것이 단순 DB 이론이 아니라 운영 가능한 동시성 설계다.

데드락을 학습한 뒤의 마지막 확인 질문은 “이 실패를 없앨 것인가, 받아들이고 안전하게 복구할 것인가”다. 모든 deadlock을 설계로 제거하려고 하면 transaction이 과하게 직렬화되고 처리량이 떨어질 수 있다. 반대로 deadlock을 전부 정상이라고만 보면 반복되는 순서 오류와 넓은 range scan을 방치한다. 좋은 기준은 빈도와 원인이다. 드물게 발생하고 재시도 성공률이 높으며 외부 side effect가 안전하면 bounded retry로 충분할 수 있다. 특정 API, 특정 index range, 특정 배치 시간에 집중된다면 설계 수리 대상이다. 이 판단을 위해서는 deadlock count, retry count, retry success latency, victim SQL fingerprint, locked index/range를 metric과 log에 남겨야 한다.

```text
deadlock metric sketch

db.deadlock.count{service,operation,engine}
db.deadlock.retry.count{service,operation}
db.deadlock.retry.success{service,operation}
db.deadlock.retry.latency_ms{service,operation}
db.lock_wait.timeout.count{service,operation}
db.lock_wait.holder_seconds{operation}

좋은 대시보드 질문
- deadlock이 특정 operation에 몰리는가?
- retry가 대부분 성공하는가, 아니면 같은 요청이 계속 실패하는가?
- timeout이 deadlock보다 훨씬 많다면 holder duration 문제가 아닌가?
- 배포나 batch 시간과 상관관계가 있는가?
```

이 관측까지 닫히면 deadlock은 운에 맡기는 장애가 아니라 관리 가능한 동시성 비용이 된다. DB는 cycle을 감지하고 victim을 고를 수 있지만, 어떤 업무 단위를 다시 실행해도 되는지, 어떤 실패율부터 설계를 고쳐야 하는지, 어떤 외부 효과가 중복되면 안 되는지는 애플리케이션과 운영의 책임이다. 그래서 최종 결론은 단순하다. deadlock은 “재시도하면 됨”도 아니고 “무조건 DB 문제”도 아니다. wait graph로 원인을 설명하고, 실패 경계를 안전하게 만들고, 반복 패턴은 lock 순서와 범위와 transaction 길이로 줄이는 문제다.
