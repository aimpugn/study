# 트랜잭션 고립 수준과 이상 현상

## SQL 표준 isolation과 anomaly

트랜잭션 고립 수준은 `동시에 실행되는 SQL이 서로의 중간 상태를 어디까지 보아도 되는가`를 정하는 약속이다. 이 절을 읽은 뒤에는 `READ UNCOMMITTED`, `READ COMMITTED`, `REPEATABLE READ`, `SERIALIZABLE`이라는 이름을 단순 암기하지 않고, 두 세션이 같은 행과 같은 조건 범위를 만졌을 때 어떤 결과가 보이면 어떤 이상 현상인지 schedule로 설명할 수 있어야 한다. 특히 중요한 출발점은 “고립 수준 이름이 같으면 모든 DBMS가 같은 방식으로 막는다”는 생각을 버리는 것이다. PostgreSQL 공식 문서는 표준의 네 수준과 세 가지 현상 표를 보여 주면서도 PostgreSQL의 실제 구현에서는 `READ UNCOMMITTED`가 `READ COMMITTED`처럼 동작하고, `REPEATABLE READ`가 표준이 허용하는 phantom read까지 허용하지 않는다고 설명한다. MySQL InnoDB 공식 문서는 기본 고립 수준이 `REPEATABLE READ`이며, consistent read와 locking read가 다르게 동작하고, gap lock과 next-key lock이 범위 삽입을 막는 데 쓰인다고 설명한다. 그래서 이 DU의 핵심 문장은 다음과 같다. 고립 수준은 이름이 아니라 `읽기 시점`, `잠금 범위`, `쓰기 충돌 처리`, `재시도 필요성`을 함께 보아야 정확히 이해된다.

이 개념이 생긴 배경은 단순하다. 한 사람이 장부 파일을 열고 수정하던 시대에는 “작업 중인 값”을 다른 사람이 볼 가능성이 작았다. 다중 사용자 DB가 등장하면서 두 사용자가 같은 계좌, 같은 주문, 같은 재고 범위를 동시에 읽고 쓰게 되었고, DB는 성능을 위해 가능한 한 동시에 실행하면서도 애플리케이션이 견딜 수 없는 중간 상태를 막아야 했다. 가장 강하게 막는 방법은 모든 트랜잭션을 하나씩 순서대로 실행하는 것이지만, 그러면 처리량이 낮다. 반대로 아무 제한 없이 동시에 실행하면 빠르지만, 아직 커밋하지 않은 값, 읽을 때마다 바뀌는 값, 조건 검색 범위에 갑자기 끼어드는 행, 서로는 합리적이지만 합치면 불가능한 결과가 생긴다. 고립 수준은 이 둘 사이에서 어느 현상을 금지하고 어느 동시성을 허용할지 정한 단계적 계약이다.

첫 번째 벽돌은 한 행짜리 계좌표다. 이 작은 표만 있어도 dirty read, nonrepeatable read, phantom read의 차이를 손으로 그릴 수 있다.

```sql
CREATE TABLE account_balance (
    account_id integer PRIMARY KEY,
    owner_name text NOT NULL,
    balance integer NOT NULL,
    status text NOT NULL
);

INSERT INTO account_balance(account_id, owner_name, balance, status)
VALUES
    (1, 'kim', 100, 'ACTIVE'),
    (2, 'lee', 100, 'ACTIVE');
```

아래 schedule에서 `A`와 `B`는 서로 다른 DB 세션이다. 핵심은 SQL 자체보다 “어느 시점의 값을 읽었는가”다.

```text
dirty read 후보

time | session A                         | session B
-----+-----------------------------------+-----------------------------------------
t1   | BEGIN;                            |
t2   | UPDATE account_balance            |
     |    SET balance = 0                |
     |  WHERE account_id = 1;            |
t3   |                                   | BEGIN;
t4   |                                   | SELECT balance FROM account_balance
     |                                   |  WHERE account_id = 1;
t5   | ROLLBACK;                         |
t6   |                                   | -- t4에서 0을 보았다면 B는 사라진 값을 본 것

값의 운명
원래 committed value: account_id=1, balance=100
A의 uncommitted value: account_id=1, balance=0
A rollback 뒤 durable value: account_id=1, balance=100
B가 t4에서 0을 보았다면: dirty read
```

더러운 읽기(dirty read)는 다른 트랜잭션이 아직 커밋하지 않은 값을 읽는 현상이다. 이 값은 이후 rollback으로 사라질 수 있으므로, 읽은 쪽은 존재하지 않았던 사실을 기반으로 판단한다. 대부분의 실무 DBMS는 이 현상을 매우 위험하게 본다. PostgreSQL에서는 `READ UNCOMMITTED`를 요청해도 실제로는 `READ COMMITTED`처럼 커밋된 값만 보게 한다. InnoDB도 일반적인 consistent read에서는 커밋되지 않은 다른 트랜잭션의 값을 읽지 않는 방향으로 동작한다. 그러므로 “READ UNCOMMITTED라는 이름이 있으니 PostgreSQL에서도 dirty read가 가능하다”는 설명은 PostgreSQL 기준으로 틀리다. 이름보다 구현 문서를 보아야 한다는 첫 번째 함정이 여기서 나온다.

반복 불가능 읽기(nonrepeatable read)는 같은 트랜잭션 안에서 같은 행을 두 번 읽었는데, 다른 트랜잭션의 커밋 때문에 두 결과가 달라지는 현상이다.

```text
nonrepeatable read 후보

time | session A                                      | session B
-----+------------------------------------------------+-----------------------------------------
t1   | BEGIN;                                         |
t2   | SELECT balance FROM account_balance            |
     |  WHERE account_id = 1; -- 100                  |
t3   |                                                | BEGIN;
t4   |                                                | UPDATE account_balance
     |                                                |    SET balance = 80
     |                                                |  WHERE account_id = 1;
t5   |                                                | COMMIT;
t6   | SELECT balance FROM account_balance            |
     |  WHERE account_id = 1; -- 80이면 현상 발생     |

읽기 규칙
READ COMMITTED: 각 SELECT가 시작될 때 새 committed snapshot을 잡는다.
REPEATABLE READ 계열: 트랜잭션이 보는 snapshot을 더 오래 유지한다.
SERIALIZABLE: snapshot 유지뿐 아니라 serial order와 맞지 않는 실행도 막거나 실패시킨다.
```

`READ COMMITTED`에서는 이 현상이 자연스럽다. 이름 그대로 statement가 시작될 때 커밋된 데이터를 읽으므로, t2와 t6 사이에 B가 커밋하면 A의 두 번째 SELECT는 새 값을 본다. 이 동작은 버그가 아니라 계약이다. 실무에서 이 계약을 모르고 “트랜잭션 안에서 읽었으니 같은 값일 것”이라고 믿으면, 할인율 검증, 잔액 확인, 재고 확인처럼 한 번 읽은 값을 뒤에서 다시 신뢰하는 코드가 깨진다. 반대로 `REPEATABLE READ`나 snapshot 기반 격리에서는 첫 SELECT 시점 또는 트랜잭션 시작 시점의 snapshot을 유지해 같은 행을 다시 읽을 때 같은 값을 보게 한다. 다만 같은 이름이라도 PostgreSQL과 InnoDB의 세부가 다르므로, 여기서 곧바로 “둘 다 완전히 같다”고 결론 내리면 다음 DU31에서 다룰 함정으로 들어간다.

팬텀 읽기(phantom read)는 같은 조건으로 범위를 다시 조회했는데, 다른 트랜잭션이 삽입하거나 삭제한 행 때문에 결과 집합의 구성 자체가 달라지는 현상이다. nonrepeatable read가 “이미 있던 행의 값이 바뀜”이라면 phantom은 “조건에 맞는 행의 집합이 바뀜”이다.

```sql
CREATE TABLE reservation (
    reservation_id integer PRIMARY KEY,
    room_id integer NOT NULL,
    reserved_date date NOT NULL,
    status text NOT NULL
);

INSERT INTO reservation VALUES
    (10, 7, DATE '2026-05-20', 'CONFIRMED');
```

```text
phantom read 후보

time | session A                                               | session B
-----+---------------------------------------------------------+----------------------------------------
t1   | BEGIN;                                                  |
t2   | SELECT count(*) FROM reservation                        |
     |  WHERE room_id = 7                                      |
     |    AND reserved_date = DATE '2026-05-20'                 |
     |    AND status = 'CONFIRMED'; -- 1                       |
t3   |                                                         | BEGIN;
t4   |                                                         | INSERT INTO reservation
     |                                                         | VALUES (11, 7, DATE '2026-05-20',
     |                                                         |         'CONFIRMED');
t5   |                                                         | COMMIT;
t6   | SELECT count(*) FROM reservation                        |
     |  WHERE room_id = 7                                      |
     |    AND reserved_date = DATE '2026-05-20'                 |
     |    AND status = 'CONFIRMED'; -- 2이면 phantom           |

집합 변화
첫 조회 결과: {reservation_id=10}
두 번째 조회 결과: {reservation_id=10, reservation_id=11}
바뀐 것은 row value 하나가 아니라 predicate(room_id/date/status)가 가리키는 집합이다.
```

이 차이는 실무에서 매우 중요하다. “이미 읽은 행을 잠갔다”는 감각만으로는 아직 존재하지 않던 행의 삽입을 막을 수 없다. 범위 조건을 보호하려면 DBMS가 인덱스 범위, gap, predicate 같은 더 넓은 개념을 다루어야 한다. InnoDB의 next-key lock과 gap lock은 이런 범위 삽입 문제와 연결된다. PostgreSQL의 Serializable Snapshot Isolation은 predicate lock이라는 관측 표면을 사용해 직렬화 불가능한 관계를 추적한다. 둘은 같은 이름의 `SERIALIZABLE`을 향해 가지만, 내부 장치는 다르다.

직렬화 이상(serialization anomaly)은 더 미묘하다. dirty read, nonrepeatable read, phantom read가 “이 특정 값이나 집합이 바뀌었다”처럼 비교적 눈에 보이는 현상이라면, serialization anomaly는 각 트랜잭션이 자기 snapshot 안에서는 타당한 결정을 했는데 전체 결과가 어떤 직렬 순서로도 설명되지 않는 상황이다. 대표 예시는 write skew다.

```sql
CREATE TABLE doctor_on_call (
    doctor_id integer PRIMARY KEY,
    doctor_name text NOT NULL,
    on_call boolean NOT NULL
);

INSERT INTO doctor_on_call VALUES
    (1, 'kim', true),
    (2, 'lee', true);
```

병원 규칙은 `항상 최소 한 명의 의사가 당직이어야 한다`라고 하자. 이 규칙은 단일 행 제약조건만으로는 표현되지 않는다. 두 트랜잭션이 같은 snapshot에서 “나 말고 다른 사람이 당직이다”를 확인하고 각각 자기 행을 off로 바꾸면, 각자 읽은 세계에서는 합리적이지만 최종 상태는 규칙을 깬다.

```text
write skew / serialization anomaly 후보

time | session A                                              | session B
-----+--------------------------------------------------------+----------------------------------------
t1   | BEGIN;                                                 | BEGIN;
t2   | SELECT count(*) FROM doctor_on_call                    |
     |  WHERE on_call = true; -- 2                            |
t3   |                                                        | SELECT count(*) FROM doctor_on_call
     |                                                        |  WHERE on_call = true; -- 2
t4   | UPDATE doctor_on_call SET on_call = false              |
     |  WHERE doctor_id = 1;                                  |
t5   |                                                        | UPDATE doctor_on_call SET on_call = false
     |                                                        |  WHERE doctor_id = 2;
t6   | COMMIT;                                                |
t7   |                                                        | COMMIT 또는 serialization failure

최종 상태 후보
doctor_id=1, on_call=false
doctor_id=2, on_call=false

직렬 순서로 설명하려고 하면:
A가 먼저였다면 B는 t3에서 count=1을 보아야 한다.
B가 먼저였다면 A는 t2에서 count=1을 보아야 한다.
둘 다 count=2를 보고 둘 다 커밋했다면, 어떤 직렬 순서와도 맞지 않는다.
```

이 예제는 “repeatable read면 충분하다”는 오해를 고친다. snapshot이 반복 읽기를 안정시켜도, 서로 다른 행을 갱신하면서 같은 집합 불변식을 깨는 문제는 남을 수 있다. PostgreSQL의 `SERIALIZABLE`은 이런 위험한 read/write dependency를 추적해 어느 한 트랜잭션을 serialization failure로 중단시킬 수 있다. 애플리케이션은 이 실패를 장애로만 보지 말고 retry 가능한 동시성 충돌로 처리해야 한다. MySQL InnoDB에서도 `SERIALIZABLE`은 일반 `SELECT`를 더 잠금 읽기처럼 다루는 방향으로 동시성을 줄여 phantom 계열 문제를 막으려 한다. 하지만 이 역시 세부 동작은 SQL 모양, 인덱스, locking read 여부에 따라 달라진다.

표준 격리 수준을 현상 중심으로 요약하면 아래처럼 읽을 수 있다. 이 표는 “각 DBMS가 그대로 이 표처럼 구현한다”는 뜻이 아니라, 이상 현상을 구분하는 언어를 제공한다.

| isolation level | dirty read | nonrepeatable read | phantom read | serialization anomaly |
|---|---:|---:|---:|---:|
| READ UNCOMMITTED | 표준상 가능 | 가능 | 가능 | 가능 |
| READ COMMITTED | 방지 | 가능 | 가능 | 가능 |
| REPEATABLE READ | 방지 | 방지 | 표준상 가능 | 가능 |
| SERIALIZABLE | 방지 | 방지 | 방지 | 방지해야 함 |

여기서 `표준상 가능`이라는 표현을 조심해야 한다. PostgreSQL 문서는 자체 표에서 `READ UNCOMMITTED`가 dirty read를 허용하지 않고, `REPEATABLE READ`도 phantom read를 허용하지 않는다고 표시한다. 이는 표준보다 강하게 막는 구현이 허용되기 때문이다. 반대로 어떤 DBMS가 더 강하게 막는다고 해서 애플리케이션 설계자가 이름만 보고 안전을 일반화해도 된다는 뜻은 아니다. 공식 문서를 읽을 때는 “이 수준에서 무엇을 허용한다”보다 “이 DBMS가 이 수준을 어떻게 구현하고, 어떤 SQL에서 어떤 snapshot과 lock을 쓰는가”를 확인해야 한다.

실무 함정은 보통 세 갈래로 나타난다.

1. `READ COMMITTED`를 “트랜잭션 안에서는 읽기가 고정된다”로 오해한다.

    이 오해는 운영에서 간헐적인 검증 실패로 나타난다. 예를 들어 주문 생성 트랜잭션이 처음에는 재고가 있다고 읽고, 뒤에서 다시 읽었더니 수량이 달라져 할인, 배송, 결제 판단이 엇갈린다. 이때 개발자가 “분명 한 트랜잭션인데 왜 바뀌지?”라고 묻는다면 격리 수준 계약을 잘못 기억하고 있는 것이다. 해결은 무조건 고립 수준을 올리는 것이 아니라, 필요한 불변식을 어떤 statement, lock, constraint, retry 정책으로 보호할지 정하는 것이다.

2. `REPEATABLE READ`를 “모든 동시성 문제가 사라진다”로 오해한다.

    반복 읽기는 같은 snapshot을 제공하지만, snapshot 안에서 내린 두 결정이 함께 커밋될 때 전역 불변식을 깰 수 있다. write skew가 대표적이다. 단일 행을 서로 업데이트하는 lost update 계열은 DBMS가 감지하거나 잠금으로 막을 수 있지만, 서로 다른 행을 업데이트하면서 집합 규칙을 깨는 경우는 별도 제약, 명시적 locking read, serializable, 또는 애플리케이션 retry 설계가 필요하다.

3. `SERIALIZABLE`을 “느리지만 항상 성공하는 안전 모드”로 오해한다.

    직렬화 가능성은 결과가 어떤 직렬 순서와 같아야 한다는 뜻이지, 모든 동시 실행을 성공시킨다는 뜻이 아니다. PostgreSQL의 serializable은 충돌 관계를 발견하면 한 트랜잭션을 실패시켜 애플리케이션이 다시 시도하게 한다. 따라서 운영 코드에는 SQLSTATE `40001` 같은 serialization failure를 분류하고 idempotent하게 재시도하는 경계가 필요하다. retry가 없으면 안전한 고립 수준을 선택했는데도 사용자는 500 오류를 본다.

관측과 검증은 반드시 두 세션으로 해야 한다. 하나의 세션에서 isolation level만 바꿔 보는 것은 동시성 현상을 만들지 못한다. 아래는 로컬 seed로 둘 수 있는 공통 관측 절차다. 실제 PostgreSQL과 MySQL에서 문법 일부는 다르지만, schedule의 의미는 같다.

```text
검증 루틴

1. 두 터미널을 연다.
2. 같은 test schema를 만든다.
3. session A에서 BEGIN 뒤 첫 SELECT를 실행한다.
4. session B에서 UPDATE 또는 INSERT 후 COMMIT한다.
5. session A에서 같은 SELECT를 다시 실행한다.
6. 값이 바뀌었는지, 집합이 바뀌었는지, commit이 실패했는지 기록한다.

PASS 신호
- dirty read 실험: 다른 세션의 rollback될 값을 보지 않는다.
- nonrepeatable read 실험: READ COMMITTED에서는 두 번째 SELECT가 새 committed 값을 볼 수 있다.
- repeatable read 실험: 같은 snapshot을 유지하는 DBMS에서는 두 번째 SELECT가 첫 결과와 같다.
- serializable write skew 실험: 둘 다 성공하면 위험하고, 하나가 serialization failure로 실패하면 DB가 직렬화 불가능성을 막은 것이다.

FAIL 신호
- isolation level을 확인하지 않고 결과만 비교한다.
- SELECT와 SELECT FOR UPDATE, consistent read와 locking read를 섞어 놓고 같은 실험이라고 말한다.
- 인덱스 없는 predicate 실험과 인덱스 있는 predicate 실험을 같은 lock 범위로 해석한다.
```

공식 자료와 로컬 seed의 경계도 분명히 해야 한다. 이 절의 현상 이름, PostgreSQL의 실제 고립 수준 차이, InnoDB의 기본 고립 수준과 gap/next-key lock 연결은 PostgreSQL `Transaction Isolation` 문서와 MySQL 8.4 `InnoDB Transaction Isolation Levels` 문서를 기준으로 삼았다. 로컬 계획 파일 `database/database-deep-study-plan.md`는 이 절이 고립 수준을 추상 표가 아니라 재현 가능한 두 세션 실험과 운영 함정으로 설명해야 한다는 방향을 제공한 seed다. 이 문서 안의 schedule은 설명과 검증을 위한 synthetic lab이며, 실제 운영 DB의 lock 범위와 실패 메시지는 버전, 설정, 인덱스, SQL 형태에 따라 공식 문서와 로컬 실험으로 다시 확인해야 한다.

마지막으로 기억할 한 문장은 이것이다. 고립 수준은 “얼마나 안전한가”라는 한 줄 등급표가 아니라, 어떤 읽기 시점을 잡고, 어떤 범위를 잠그거나 추적하고, 충돌이 생기면 기다릴지 실패시킬지 정하는 동시성 계약이다. 같은 이름의 격리 수준도 DBMS가 다른 내부 장치를 쓰면 관측 결과가 달라진다. 그래서 실무자는 이름을 외우는 데서 멈추지 않고, 두 세션 schedule을 직접 그려 “무엇을 보았고, 무엇을 못 보았고, 어느 순간 실패해야 안전한가”를 확인해야 한다.

고립 수준을 더 깊게 이해하려면 이상 현상을 “읽기 오류 목록”으로만 외우지 말고, 애플리케이션 불변식이 깨지는 방식으로 다시 읽어야 한다. dirty read는 사라질 값을 읽는 문제라서 회계, 결제, 재고처럼 외부 행동을 유발하는 코드에서 특히 위험하다. nonrepeatable read는 같은 트랜잭션 안에서 판단의 기준점이 움직이는 문제라서 “처음 확인한 값으로 뒤 결정을 한다”는 코드에서 위험하다. phantom read는 조건 범위의 구성원이 바뀌는 문제라서 예약, 쿠폰 발급, 재고 배정처럼 “조건에 맞는 것이 몇 개인가”를 보는 코드에서 위험하다. serialization anomaly는 단일 값이나 단일 범위보다 더 넓은 불변식이 깨지는 문제라서 당직자 수, 계좌 총합, 중복 처리 winner, 한도 합산처럼 여러 행이 함께 의미를 만드는 코드에서 위험하다.

```text
애플리케이션 불변식 관점의 anomaly map

+-----------------------+---------------------------+--------------------------------------+--------------------------------------+
| anomaly               | DB에서 보이는 모양        | 애플리케이션에서 깨지는 판단        | 자주 필요한 대응                      |
+-----------------------+---------------------------+--------------------------------------+--------------------------------------+
| dirty read            | rollback될 값을 읽음      | 존재하지 않은 결제/재고/상태로 행동 | 커밋된 읽기, 외부 side effect 지연    |
| nonrepeatable read    | 같은 row 값이 다시 바뀜   | 같은 요청 안의 기준값 불일치        | snapshot, 재검증, 조건부 UPDATE       |
| phantom read          | predicate 결과 집합 변화  | 범위 count/존재성 판단 불일치       | range lock, unique/constraint, retry  |
| serialization anomaly | 어떤 직렬 순서와도 불일치 | 여러 row 불변식 붕괴                | SERIALIZABLE, materialized guard row  |
+-----------------------+---------------------------+--------------------------------------+--------------------------------------+

읽는 순서
1. 먼저 애플리케이션 불변식을 문장으로 쓴다.
2. 그 불변식이 단일 row, predicate range, 여러 row relation 중 어디에 걸리는지 본다.
3. 그다음 isolation level 이름이 아니라 해당 DBMS의 읽기/잠금/실패 정책을 고른다.
```

예를 들어 “하루에 한 방은 한 번만 예약되어야 한다”는 규칙은 단순 count 조회로 보호하기보다 `(room_id, reserved_date)`에 unique constraint를 두는 편이 더 강하다. 고립 수준을 높여도 애플리케이션이 실패를 retry하지 않거나, 범위를 보호하지 못하거나, 읽기와 쓰기 사이에 외부 API를 호출하면 여전히 틈이 생긴다. 반대로 unique constraint는 두 트랜잭션이 어떤 snapshot을 보았든 최종 쓰기 지점에서 같은 key 충돌을 만든다. 즉 고립 수준은 중요하지만, 모든 비즈니스 불변식을 대신 작성해 주는 장치는 아니다. DB가 가장 잘 막는 것은 명시된 제약과 명확한 충돌이다. 애매한 업무 규칙은 행, key, constraint, lock 대상, retry 정책으로 번역해야 DB가 도와줄 수 있다.

다음은 같은 예약 문제를 세 가지 방식으로 고치는 비교 trace다.

```sql
-- 약한 방식: count만 보고 insert한다.
BEGIN;
SELECT count(*) FROM reservation
 WHERE room_id = 7
   AND reserved_date = DATE '2026-05-20'
   AND status = 'CONFIRMED';
-- 0이라고 믿고 애플리케이션에서 예약 가능 판단
INSERT INTO reservation(reservation_id, room_id, reserved_date, status)
VALUES (21, 7, DATE '2026-05-20', 'CONFIRMED');
COMMIT;

-- 더 강한 방식 1: key 자체를 충돌하게 만든다.
CREATE UNIQUE INDEX reservation_once_per_room_day
    ON reservation(room_id, reserved_date)
 WHERE status = 'CONFIRMED';

-- 더 강한 방식 2: guard row를 잠근다.
CREATE TABLE room_day_guard (
    room_id integer NOT NULL,
    reserved_date date NOT NULL,
    PRIMARY KEY(room_id, reserved_date)
);

BEGIN;
SELECT * FROM room_day_guard
 WHERE room_id = 7 AND reserved_date = DATE '2026-05-20'
 FOR UPDATE;
-- 이 guard row를 잡은 세션만 해당 room/day 예약 판단을 진행한다.
COMMIT;
```

```text
세 방식의 실패 지점

count-only
  A와 B가 모두 count=0을 볼 수 있다.
  둘 다 INSERT까지 성공하면 중복 예약이 된다.

unique constraint
  A와 B가 모두 count=0을 보더라도 INSERT/COMMIT 지점에서 한쪽이 key 충돌을 만난다.
  실패를 사용자에게 어떻게 돌려줄지 또는 retry할지는 애플리케이션 책임이다.

guard row
  A가 guard row를 FOR UPDATE로 잡으면 B는 기다린다.
  대기 시간이 길어지면 lock wait와 connection 점유가 운영 문제가 된다.
```

이 비교가 중요한 이유는 “isolation level을 올리면 된다”는 단순 처방을 막기 위해서다. isolation level은 읽기와 동시 실행의 기본 계약을 바꾸지만, 업무 규칙을 어떤 데이터 구조로 표현할지는 별도 설계다. senior 개발자는 먼저 보호할 불변식을 key나 row로 좁힐 수 있는지 본다. 좁힐 수 있으면 constraint나 conditional update가 가장 재현 가능하다. 좁힐 수 없고 여러 행의 관계를 읽어야 하면 serializable과 retry, 또는 명시적 locking read, 또는 업무 단위 guard row를 검토한다. 어느 쪽이든 실패 신호를 정상적인 동시성 결과로 다루는 코드가 필요하다.

관측 쿼리도 이상 현상별로 달라야 한다. 값이 바뀌었는지 보려면 같은 SELECT 결과를 기록하면 되지만, lock 때문에 기다리는지 보려면 lock view를 봐야 하고, serialization failure를 보려면 SQLSTATE와 로그를 봐야 한다. 다음 표는 실험을 문서화할 때 남겨야 할 최소 관측 항목이다.

| 실험 | PostgreSQL 관측 | MySQL/InnoDB 관측 | PASS로 볼 신호 | FAIL로 볼 신호 |
|---|---|---|---|---|
| dirty read | 두 번째 세션 SELECT 결과, isolation 표시 | 두 번째 세션 SELECT 결과, isolation 표시 | rollback될 값을 보지 않음 | isolation을 확인하지 않고 값만 기록 |
| nonrepeatable read | `SHOW transaction_isolation`, 두 SELECT 결과 | `SELECT @@transaction_isolation`, 두 SELECT 결과 | READ COMMITTED에서 값 변화 재현 | 같은 세션/같은 statement만 테스트 |
| phantom read | count 결과, 실행 plan, 필요 시 lock 관측 | count 결과, index/gap lock 관측 | 범위 결과 변화 또는 차단을 설명 | index 유무를 누락 |
| serialization anomaly | SQLSTATE, error message, retry 성공 여부 | deadlock/lock wait/serializable 동작 구분 | 실패를 retry-required로 분류 | 실패를 모두 장애로만 기록 |

그리고 실험 결과를 문서에 남길 때는 “PostgreSQL에서는 이렇게 나왔다”와 “표준상 이 수준은 이것을 허용한다”를 같은 문장으로 섞지 않는다. 표준은 현상 이름과 금지 최소선을 제공한다. PostgreSQL과 MySQL은 그 최소선보다 강한 동작을 선택할 수 있고, 실제로 그렇게 하는 부분이 있다. 따라서 근거 표기는 항상 `표준/현상 언어`, `PostgreSQL 공식 문서`, `MySQL 공식 문서`, `로컬 실험`으로 분리한다. 이 분리가 없으면 나중에 다른 DBMS, 다른 버전, 다른 SQL 모양을 만났을 때 잘못된 일반화가 된다.

마지막으로 lost update를 별도로 언급해야 한다. 많은 설명에서 lost update를 dirty/nonrepeatable/phantom과 함께 섞어 버리지만, 실무에서는 독립적으로 자주 터진다. 두 트랜잭션이 같은 값을 읽고 각각 계산한 새 값을 덮어쓰면 한쪽 변경이 사라진다. 예를 들어 잔액 100을 두 세션이 읽고 각각 `+10`을 계산해 `110`을 저장하면 최종 값은 120이 아니라 110이 될 수 있다. 이를 막으려면 `UPDATE wallet SET amount = amount + 10 WHERE id = 1`처럼 DB 안에서 원자적으로 계산하거나, version 컬럼을 둔 optimistic locking을 쓰거나, `SELECT ... FOR UPDATE`로 읽기와 쓰기 사이의 소유권을 잡아야 한다. 이 문제는 격리 수준 이름만 외우면 놓치기 쉽지만, 실제 서비스 장애에서는 매우 흔하다.

```text
lost update hand trace

초기 amount=100

A: SELECT amount -> 100
B: SELECT amount -> 100
A: app calculates 100 + 10 = 110
B: app calculates 100 + 10 = 110
A: UPDATE wallet SET amount = 110 WHERE id = 1
B: UPDATE wallet SET amount = 110 WHERE id = 1
최종 amount=110

원래 기대
100 + 10 + 10 = 120

수리 1
UPDATE wallet SET amount = amount + 10 WHERE id = 1;

수리 2
UPDATE wallet
   SET amount = 110, version = version + 1
 WHERE id = 1
   AND version = 3;
-- affected row가 0이면 누군가 먼저 바꾼 것이므로 다시 읽고 재시도한다.
```

이 예제는 고립 수준 학습의 마무리로 적합하다. isolation anomaly는 교과서 표로 시작하지만, 실무에서는 결국 데이터 구조와 SQL 형태, 실패 처리 방식으로 돌아온다. 좋은 DB 코드는 “어떤 isolation level인가”를 주석처럼 적는 데서 끝나지 않고, 어떤 값이 동시에 바뀔 수 있는지, 어떤 충돌을 DB가 감지할 수 있는지, 감지된 실패를 어느 계층에서 retry하거나 사용자 메시지로 바꿀지까지 설계한다.

면접이나 설계 리뷰에서 이 주제를 설명할 때는 마지막에 “어떤 이상 현상을 막고 싶은가”를 업무 언어로 되물어야 한다. `SERIALIZABLE로 올리겠습니다`라는 답은 아직 설계가 아니다. 예를 들어 쿠폰 중복 발급을 막으려는 것인지, 보고서 조회 중 값이 바뀌지 않게 하려는 것인지, 재고 차감을 원자적으로 만들려는 것인지, 여러 지점 잔액 합이 음수가 되지 않게 하려는 것인지가 다르다. 쿠폰 중복 발급은 unique key와 idempotency replay가 더 직접적일 수 있고, 보고서 조회는 snapshot 격리가 맞을 수 있으며, 재고 차감은 조건부 UPDATE가 더 단순할 수 있고, 여러 행 합계 불변식은 serializable retry나 guard row가 필요할 수 있다. 고립 수준은 선택지 중 하나이지, 보호하려는 불변식을 말하지 않아도 되는 면허가 아니다.

```text
업무 요구에서 격리 선택으로 내려가는 trace

요구: 같은 사용자가 같은 쿠폰을 두 번 받을 수 없다.
  -> 움직이는 값: (coupon_code, user_id) key
  -> 가장 좋은 충돌 지점: UNIQUE(coupon_code, user_id)
  -> isolation 보조: READ COMMITTED여도 key 충돌은 막힘
  -> 실패 처리: duplicate key면 기존 지급 결과 replay 또는 이미 지급 메시지

요구: 관리자가 보는 월별 보고서가 조회 중 흔들리지 않아야 한다.
  -> 움직이는 값: 여러 statement가 읽는 snapshot
  -> 가장 좋은 보호 지점: transaction snapshot 또는 exported snapshot
  -> isolation 보조: REPEATABLE READ 계열
  -> 실패 처리: 오래 열린 snapshot이 vacuum/purge를 방해하지 않는지 관측

요구: 최소 한 명의 당직자는 남아야 한다.
  -> 움직이는 값: 여러 row가 만드는 count 불변식
  -> 가장 좋은 충돌 지점: guard row, serializable retry, constraint로 재모델링
  -> isolation 보조: plain repeatable read만으로는 부족할 수 있음
  -> 실패 처리: serialization/deadlock retry와 사용자 메시지 분리
```

이렇게 거꾸로 내려가면 고립 수준 표의 각 칸이 살아난다. dirty read는 외부 행동을 유발하기 전 커밋 여부를 보장해야 한다는 신호이고, nonrepeatable read는 한 요청 안에서 같은 기준을 유지해야 한다는 신호이며, phantom은 조건 범위가 업무 자원이라는 신호이고, serialization anomaly는 여러 행의 관계 자체가 업무 자원이라는 신호다. 그래서 관측도 업무 자원에 맞춰야 한다. 단일 row라면 affected row count와 version을 보고, 범위라면 index와 lock wait를 보고, 여러 행 관계라면 실패율과 retry 성공률, 불변식 위반 로그를 본다. 이 정도로 설명할 수 있어야 isolation level을 단순 암기에서 실제 설계 도구로 옮긴 것이다.

또 하나의 검증 습관은 “결과값”과 “허용된 실행 순서”를 분리해 적는 것이다. 두 트랜잭션이 모두 성공했고 최종 값도 우연히 맞아 보일 수 있다. 하지만 그 결과가 어떤 직렬 순서와도 맞지 않으면 여전히 isolation 문제다. 반대로 한쪽 트랜잭션이 실패했더라도, 그 실패가 직렬화 가능성을 지키기 위한 정상 신호라면 설계가 실패한 것이 아니다. 따라서 lab 기록에는 최종 row만 쓰지 말고, 각 세션이 읽은 snapshot, 각 세션이 쓴 row, commit 또는 abort 시점, retry 여부를 함께 남긴다.

이 기록 방식은 나중에 장애 회고에서도 그대로 쓴다. “값이 틀렸다”가 아니라 “A는 어떤 snapshot을 보고 어떤 결정을 했고, B는 그 사이 어떤 값을 커밋했으며, DB는 어느 시점에 실패시켰어야 했는가”로 써야 같은 실수를 다시 막을 수 있다.

### 마지막 실무 연결: 고립 수준은 장애를 없애는 스위치가 아니라 재시도 계약이다

고립 수준을 높이면 모든 문제가 사라진다고 생각하기 쉽지만, 실제로는 다른 종류의 비용과 실패 신호가 나타난다. `SERIALIZABLE`은 이상 현상을 더 강하게 막을 수 있지만, 어떤 실행은 직렬 순서로 설명할 수 없다고 판단되어 실패할 수 있다. 이 실패는 DB 버그가 아니라 정합성을 지키기 위한 정상 신호다.

따라서 고립 수준을 선택할 때는 애플리케이션 retry 가능성을 함께 설계한다. 같은 command를 다시 실행해도 안전한지, 외부 API 호출은 이미 나갔는지, 사용자에게 성공 화면을 보여 주었는지, idempotency key가 있는지 확인한다. 고립 수준은 DB 내부 설정이지만, 그 결과로 나오는 retry와 abort는 애플리케이션 계약까지 이어진다.

### 추가 판정 질문: 이상 현상 이름보다 업무 불변식을 먼저 말한다

실무 설계에서 dirty read, nonrepeatable read, phantom read 이름을 외우는 것보다 중요한 것은 어떤 업무 불변식이 깨지는지 말하는 것이다. 잔액은 음수가 되면 안 된다, 최소 한 명의 당직자는 남아야 한다, 같은 쿠폰은 한 번만 써야 한다, 예약 범위에는 중복 좌석이 없어야 한다. 이 문장이 먼저 있어야 어떤 고립 수준과 어떤 lock/constraint가 필요한지 정할 수 있다.

따라서 isolation review는 표준 표를 읽고 끝내지 않는다. 업무 불변식을 적고, 두 세션 schedule로 깨뜨려 본 뒤, 선택한 DBMS에서 그 schedule이 어떤 결과를 내는지 실험한다. 그 결과가 실패라면 retry 정책까지 포함해야 한다.


## PostgreSQL과 InnoDB 실제 차이

PostgreSQL과 InnoDB는 둘 다 MVCC를 쓰지만, 같은 고립 수준 이름을 같은 체험으로 제공하지 않는다. 이 절의 목표는 `PostgreSQL Repeatable Read = InnoDB Repeatable Read`라는 단축 이해를 고치는 것이다. 둘 다 반복 읽기를 안정시키지만, PostgreSQL은 statement 또는 transaction snapshot과 Serializable Snapshot Isolation 쪽으로 사고가 이어지고, InnoDB는 consistent read, locking read, record/gap/next-key lock의 조합으로 사고가 이어진다. 이 차이를 모르고 애플리케이션을 옮기면 “로컬 MySQL에서는 막혔는데 PostgreSQL에서는 serialization failure가 난다”, “PostgreSQL에서 되던 SELECT가 MySQL에서는 gap lock을 유발한다”, “같은 REPEATABLE READ인데 새 row가 보이거나 안 보이는 조건이 이상하다” 같은 현상을 만난다.

이 차이가 등장한 배경은 두 엔진이 같은 문제를 서로 다른 저장 구조와 운영 철학으로 풀어 왔기 때문이다. PostgreSQL은 heap tuple에 `xmin`, `xmax` 같은 transaction id 흔적을 남기고, snapshot이 어떤 tuple version을 볼 수 있는지 판단한 뒤, 오래된 version은 vacuum으로 정리한다. InnoDB는 clustered index record를 현재 위치에 두고 undo log chain과 read view를 이용해 과거 committed version을 따라간다. 둘 다 “읽는 쪽이 쓰는 쪽을 매번 기다리지 않게 한다”는 목표는 비슷하지만, version을 어디에 두고, 범위 삽입을 어떻게 막고, serializable에 가까운 보장을 언제 실패로 돌려보낼지가 다르게 발전했다. 그래서 고립 수준 비교는 표준 이름의 역사만 외우는 일이 아니라, 각 엔진이 어떤 저장 구조 위에서 어떤 동시성 지뢰를 먼저 막으려 했는지 읽는 일이다.

가장 먼저 고정할 공통 구조는 읽기 종류다. PostgreSQL 문서는 `READ COMMITTED`에서 SELECT가 쿼리 시작 시점의 snapshot을 보고, `REPEATABLE READ`에서는 트랜잭션 시작 시점의 snapshot을 본다고 설명한다. MySQL InnoDB 문서는 consistent read가 snapshot을 사용하고, `SELECT ... FOR UPDATE`나 `FOR SHARE` 같은 locking read는 잠금을 잡아 동시 변경을 제어한다고 설명한다. 이 구분이 없으면 같은 SELECT라고 생각한 두 실험이 사실은 서로 다른 계약을 시험하게 된다.

```text
같은 단어처럼 보이는 읽기의 분해

PostgreSQL
  plain SELECT
    READ COMMITTED    -> statement snapshot
    REPEATABLE READ   -> transaction snapshot
    SERIALIZABLE      -> transaction snapshot + serialization conflict tracking
  SELECT FOR UPDATE
    row-level lock을 잡고, 동시 UPDATE/DELETE/SELECT FOR UPDATE와 충돌한다.

InnoDB
  consistent read
    undo 기반 read view로 과거 committed version을 읽는다.
  locking read
    SELECT ... FOR UPDATE / FOR SHARE가 record, gap, next-key lock과 연결된다.
  SERIALIZABLE
    autocommit off plain SELECT도 더 강한 잠금 읽기처럼 다루어 동시 삽입/수정을 제한할 수 있다.
```

첫 번째 비교 실험은 nonrepeatable read다.

```sql
CREATE TABLE wallet (
    id integer PRIMARY KEY,
    amount integer NOT NULL
);

INSERT INTO wallet VALUES (1, 100);
```

```text
실험 A: READ COMMITTED에서 같은 행을 두 번 읽기

time | PostgreSQL session A                         | PostgreSQL session B
-----+----------------------------------------------+-----------------------------------------
t1   | BEGIN ISOLATION LEVEL READ COMMITTED;        |
t2   | SELECT amount FROM wallet WHERE id = 1;      | -- 100
t3   |                                              | BEGIN;
t4   |                                              | UPDATE wallet SET amount = 120 WHERE id=1;
t5   |                                              | COMMIT;
t6   | SELECT amount FROM wallet WHERE id = 1;      | -- 120 가능

time | InnoDB session A                              | InnoDB session B
-----+----------------------------------------------+-----------------------------------------
t1   | SET TRANSACTION ISOLATION LEVEL READ COMMITTED; BEGIN;
t2   | SELECT amount FROM wallet WHERE id = 1;      | -- 100
t3   |                                              | BEGIN;
t4   |                                              | UPDATE wallet SET amount = 120 WHERE id=1;
t5   |                                              | COMMIT;
t6   | SELECT amount FROM wallet WHERE id = 1;      | -- 120 가능
```

이 실험에서는 두 엔진이 비슷하게 보인다. `READ COMMITTED`에서는 각 statement가 새로 커밋된 세계를 볼 수 있으므로 두 번째 SELECT가 새 값을 보는 것이 정상이다. 이 때문에 “둘 다 같네”라고 결론 내리기 쉽다. 그러나 차이는 더 높은 고립 수준, 범위 조건, 잠금 읽기에서 드러난다.

두 번째 비교 실험은 `REPEATABLE READ`에서 plain SELECT와 locking read를 분리하는 것이다.

```text
실험 B: REPEATABLE READ plain SELECT

time | session A                                     | session B
-----+-----------------------------------------------+------------------------------------------
t1   | BEGIN ISOLATION LEVEL REPEATABLE READ;        |
t2   | SELECT amount FROM wallet WHERE id = 1;       | -- 100
t3   |                                               | BEGIN;
t4   |                                               | UPDATE wallet SET amount = 120 WHERE id=1;
t5   |                                               | COMMIT;
t6   | SELECT amount FROM wallet WHERE id = 1;       | -- PostgreSQL: 100
     |                                               | -- InnoDB consistent read: 보통 100

읽기 경로
PostgreSQL: transaction snapshot이 id=1의 t2 당시 visible version을 유지한다.
InnoDB: read view와 undo chain이 t2 당시 visible version을 재구성한다.
공통 결과: plain consistent read만 보면 반복 읽기가 안정된다.
```

여기까지도 비슷해 보인다. 하지만 `SELECT ... FOR UPDATE`를 넣으면 목적이 바뀐다. plain SELECT는 “내 snapshot에서 읽기”이고, locking read는 “앞으로 내가 수정하거나 의사결정에 쓸 현재 행을 잠그기”다. InnoDB 문서는 locking read가 검색한 index records에 lock을 설정하고, gap 또는 next-key lock으로 range 삽입을 막을 수 있다고 설명한다. PostgreSQL의 row-level lock은 선택된 row와 충돌하는 업데이트를 막지만, 같은 방식의 gap lock이라는 표현으로 range 전체를 설명하지 않는다. PostgreSQL에서 범위 불변식을 직렬화 수준으로 보호하려면 predicate lock/SSI 관점이 필요하다.

```text
실험 C: 예약 범위 삽입과 잠금 읽기

CREATE TABLE seat_booking (
    id integer PRIMARY KEY,
    event_id integer NOT NULL,
    seat_no integer NOT NULL,
    status text NOT NULL
);
CREATE INDEX seat_booking_event_seat_idx
    ON seat_booking(event_id, seat_no);

session A
BEGIN;
SELECT * FROM seat_booking
 WHERE event_id = 10
   AND seat_no BETWEEN 1 AND 10
 FOR UPDATE;

session B
BEGIN;
INSERT INTO seat_booking VALUES (101, 10, 5, 'CONFIRMED');
COMMIT;

관측 포인트
PostgreSQL: 이미 선택된 row에 row lock을 건다. 비어 있는 범위 자체를 InnoDB gap lock처럼 잠근다고 단순화하면 안 된다.
InnoDB: 인덱스 범위 검색과 locking read가 next-key/gap lock을 만들 수 있고, 범위 안 삽입이 기다릴 수 있다.
```

이 차이는 “존재하지 않는 것을 어떻게 보호하는가”라는 질문으로 이어진다. PostgreSQL에서 plain `REPEATABLE READ` snapshot은 내가 본 세계를 안정시켜 주지만, 다른 트랜잭션이 새 행을 넣는 것 자체를 같은 방식으로 막는다고 생각하면 안 된다. 그 새 행은 내 snapshot에 보이지 않을 수 있다. 반대로 InnoDB의 locking read는 인덱스 범위를 잠그면서 삽입을 기다리게 만들 수 있다. 그러나 이것도 항상 마법처럼 작동하지 않는다. 적절한 인덱스가 없거나 조건이 넓거나 isolation level과 lock option이 다르면 lock 범위와 비용이 달라진다.

세 번째 비교 실험은 write skew다. PostgreSQL `SERIALIZABLE`의 특징은 이 예제에서 잘 보인다.

```text
실험 D: write skew

초기 상태
doctor_on_call
+-----------+-------------+---------+
| doctor_id | doctor_name | on_call |
+-----------+-------------+---------+
| 1         | kim         | true    |
| 2         | lee         | true    |
+-----------+-------------+---------+

PostgreSQL REPEATABLE READ
time | A                                                | B
-----+--------------------------------------------------+-----------------------------------------
t1   | BEGIN ISOLATION LEVEL REPEATABLE READ;           | BEGIN ISOLATION LEVEL REPEATABLE READ;
t2   | SELECT count(*) FROM doctor_on_call              |
     |  WHERE on_call = true; -- 2                      |
t3   |                                                  | SELECT count(*) FROM doctor_on_call
     |                                                  |  WHERE on_call = true; -- 2
t4   | UPDATE doctor_on_call SET on_call=false          |
     |  WHERE doctor_id=1;                              |
t5   |                                                  | UPDATE doctor_on_call SET on_call=false
     |                                                  |  WHERE doctor_id=2;
t6   | COMMIT;                                          | COMMIT 가능

PostgreSQL SERIALIZABLE
같은 schedule에서 둘 중 하나가 could not serialize access 같은 실패를 받을 수 있다.
이 실패는 DB가 결과를 직렬 순서와 맞추기 위해 애플리케이션 재시도를 요구하는 신호다.
```

InnoDB에서 같은 문제를 다룰 때는 plain consistent read, locking read, unique/constraint 설계, `SERIALIZABLE` 또는 명시적 잠금 선택을 함께 봐야 한다. InnoDB `REPEATABLE READ`는 gap/next-key lock이 강하다는 이미지 때문에 모든 write skew를 자동으로 막는다고 오해하기 쉽다. 그러나 plain consistent read로 count를 보고 서로 다른 행을 업데이트하는 형태에서는 읽은 predicate를 모두 잠근 것이 아닐 수 있다. `SELECT ... FOR UPDATE`를 쓰거나, 불변식을 하나의 materialized row로 모아 잠그거나, unique constraint로 충돌을 만들거나, 실패 시 retry할 수 있는 serializable 전략을 선택해야 한다. 정확한 선택은 “무엇을 보호하는가”에 따라 달라진다.

PostgreSQL과 InnoDB의 차이를 한 표로 압축하면 다음과 같다. 이 표는 학습용 요약이며, 운영 판단은 반드시 공식 문서와 실제 SQL/인덱스/버전에서 확인해야 한다.

| 관점 | PostgreSQL | MySQL InnoDB |
|---|---|---|
| 기본 isolation | 보통 READ COMMITTED를 많이 사용하지만 DB 기본값은 배포/설정 확인 필요 | 공식 문서상 InnoDB 기본은 REPEATABLE READ |
| READ UNCOMMITTED | PostgreSQL에서는 READ COMMITTED처럼 동작 | 수준은 제공하지만 consistent read와 engine 동작을 별도로 확인해야 함 |
| READ COMMITTED plain SELECT | statement snapshot | 각 consistent read가 자체 read view를 만들 수 있음 |
| REPEATABLE READ plain SELECT | transaction snapshot, phantom도 PostgreSQL 표에서는 허용하지 않음 | 첫 consistent read의 read view를 유지하는 방향 |
| range protection | row lock과 SSI predicate lock 관점으로 분리해서 봄 | record/gap/next-key lock이 핵심 관찰 포인트 |
| SERIALIZABLE | SSI가 직렬화 불가능성을 감지하고 실패시킬 수 있음 | 더 강한 잠금으로 동시성을 줄이는 쪽의 체감이 큼 |
| 운영 trap | serialization failure retry 누락 | gap/next-key lock으로 예상 밖 대기, deadlock 증가 |

이 차이는 장애 대응에서도 다르게 나타난다. PostgreSQL에서 `SERIALIZABLE`을 켠 뒤 간헐적인 `40001`이 늘었다면, 그것은 DB가 망가졌다는 신호가 아니라 애플리케이션이 retry 경계를 가져야 한다는 신호일 수 있다. `pg_locks`, `pg_stat_activity`, 로그의 serialization failure, 트랜잭션 길이를 함께 봐야 한다. InnoDB에서 같은 업무가 느려졌다면 `SHOW ENGINE INNODB STATUS`, `performance_schema.data_locks`, deadlock log, index 선택, `SELECT ... FOR UPDATE` 범위, gap lock 발생 가능성을 봐야 한다. 단순히 “둘 다 REPEATABLE READ인데 왜 다르지?”라고 묻는 순간 이미 진단 축을 잘못 잡은 것이다.

실무에서 안전한 설계 질문은 다음 순서다.

1. 이 코드가 보호해야 하는 것은 단일 행인가, 조건 범위인가, 여러 행이 만드는 불변식인가.

    단일 행이면 row lock 또는 optimistic update 조건으로 충분할 수 있다. 조건 범위면 phantom과 gap/predicate 문제를 봐야 한다. 여러 행 불변식이면 write skew와 serialization failure 또는 constraint 설계를 봐야 한다.

2. 읽기는 의사결정용 snapshot인가, 곧 수정할 대상에 대한 locking read인가.

    report 조회처럼 snapshot이 목적이면 plain SELECT가 맞을 수 있다. 예약 확정, 재고 차감, 중복 요청 winner election처럼 이후 쓰기와 연결되면 `FOR UPDATE`, unique constraint, 상태 전이 조건, retry가 필요할 수 있다.

3. 실패를 기다림으로 처리할 것인가, 즉시 실패와 retry로 처리할 것인가.

    lock wait는 latency를 늘리고 thread/connection을 붙잡는다. serialization failure와 deadlock victim은 빠르게 실패할 수 있지만 재시도 설계가 없으면 사용자 오류가 된다. 어떤 전략을 쓰든 관측과 보상 경계가 있어야 한다.

4. 인덱스가 lock 범위와 실행 계획을 어떻게 바꾸는가.

    InnoDB next-key lock은 index record와 gap 개념을 따라간다. 인덱스가 없거나 조건이 넓으면 생각보다 넓은 범위를 잠글 수 있다. PostgreSQL도 plan과 predicate lock 관찰이 연결될 수 있으므로, 실행 계획과 잠금 관측을 분리해서 보지 않는다.

아래 local lab seed는 이 절의 검증 경로다. 이 파일 안에 실제 DB 컨테이너를 만들지는 않았지만, `database/deep-dive/labs/transactions/du31-pg-vs-innodb-isolation` 아래에 둘 수 있는 최소 실험 형태를 그대로 적었다. 작성자가 실제 lab을 만들 때는 이 schedule을 SQL 파일 두 개, README, expected-output으로 쪼개면 된다.

```text
du31 lab shape

postgres/session-a.sql
  DROP TABLE IF EXISTS doctor_on_call;
  CREATE TABLE doctor_on_call(...);
  INSERT ...
  BEGIN ISOLATION LEVEL SERIALIZABLE;
  SELECT count(*) ...
  UPDATE ...
  COMMIT;

postgres/session-b.sql
  BEGIN ISOLATION LEVEL SERIALIZABLE;
  SELECT count(*) ...
  UPDATE ...
  COMMIT; -- expected: one side may fail with serialization failure

mysql/session-a.sql
  SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
  START TRANSACTION;
  SELECT count(*) ...
  UPDATE ...
  COMMIT;

mysql/session-b.sql
  SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
  START TRANSACTION;
  SELECT count(*) ...
  UPDATE ...
  COMMIT;

expected-observation.md
  - plain consistent read 결과와 locking read 결과를 분리한다.
  - InnoDB에서 SELECT ... FOR UPDATE를 추가한 버전을 별도로 둔다.
  - PostgreSQL SERIALIZABLE 실패는 retry-required PASS 신호로 기록한다.
```

이 lab에서 PASS는 “두 DB가 같은 결과를 낸다”가 아니다. PASS는 두 엔진의 차이가 예상한 축에서 관측되는 것이다. PostgreSQL에서는 snapshot과 serialization failure를 중심으로 설명이 맞아야 하고, InnoDB에서는 read view와 locking read, gap/next-key lock의 관찰이 맞아야 한다. FAIL은 isolation level만 같게 맞춘 뒤 plain SELECT와 locking read를 섞거나, 실패 메시지를 모두 같은 lock timeout으로 처리하거나, 인덱스 유무를 기록하지 않는 것이다.

공식 출처 경계는 다음처럼 정리한다. PostgreSQL 쪽 판단은 PostgreSQL `Transaction Isolation`과 `Explicit Locking` 문서에서 가져온다. MySQL 쪽 판단은 MySQL 8.4 `InnoDB Transaction Isolation Levels`, `InnoDB Locking`, `InnoDB Deadlocks` 문서에서 가져온다. 로컬 seed는 `database/database-deep-study-plan.md`와 이 저장소의 isolation 학습 계획이다. 이 절에서 든 schedule은 synthetic example이므로 실제 버전에서 메시지와 잠금 관찰 view 이름은 달라질 수 있다. 다만 “같은 이름을 같은 동작으로 보지 말고, 읽기 종류와 잠금 범위, 실패 정책을 분리해 보라”는 원리는 공식 문서와 실험으로 재현 가능하다.

마지막으로 오해를 수리하는 짧은 대조를 남긴다. PostgreSQL `REPEATABLE READ`를 쓰면 phantom이 표준 표보다 강하게 막히는 것처럼 보일 수 있지만, 그렇다고 PostgreSQL `REPEATABLE READ`가 `SERIALIZABLE`과 같은 것은 아니다. write skew 같은 직렬화 이상은 `SERIALIZABLE`과 retry 설계로 다루어야 한다. InnoDB `REPEATABLE READ`는 gap/next-key lock 때문에 더 강하게 막는 것처럼 보이는 순간이 있지만, 그 힘은 plain consistent read가 아니라 locking read와 인덱스 범위에서 주로 나온다. 이 둘을 섞으면 운영에서 “왜 막히지?”와 “왜 안 막혔지?”를 동시에 겪는다.

차이를 더 확실히 보려면 `현재 읽기(current read)`와 `일관된 읽기(consistent read)`를 분리해야 한다. InnoDB 자료를 읽을 때 자주 만나는 함정은 “REPEATABLE READ니까 모든 SELECT가 같은 과거 snapshot만 본다”는 생각이다. InnoDB에서 일반 SELECT는 consistent read로 과거 snapshot을 볼 수 있지만, UPDATE, DELETE, INSERT, `SELECT ... FOR UPDATE`, `SELECT ... FOR SHARE`는 현재 버전과 잠금이 중요해진다. 그래서 같은 트랜잭션 안에서도 “보고서처럼 읽는 SELECT”와 “이후 변경을 위해 소유권을 잡는 읽기”는 같은 말이 아니다. PostgreSQL에서도 plain SELECT와 row-level locking clause가 붙은 SELECT는 다르게 봐야 하지만, InnoDB의 gap/next-key lock처럼 인덱스 간격을 강하게 설명하는 습관을 그대로 옮기면 안 된다.

```text
InnoDB에서 헷갈리기 쉬운 같은 트랜잭션 안의 두 세계

BEGIN;
-- consistent read: read view가 허용하는 과거 committed version을 볼 수 있다.
SELECT amount FROM wallet WHERE id = 1;

-- current write: 실제 현재 row를 바꾸려 하므로 lock, 최신 committed version, 충돌이 중요하다.
UPDATE wallet SET amount = amount + 10 WHERE id = 1;

-- locking read: 읽는 동시에 앞으로의 변경 권한을 조율한다.
SELECT amount FROM wallet WHERE id = 1 FOR UPDATE;
COMMIT;

핵심 질문
- 이 SQL은 과거 snapshot을 읽는가?
- 현재 row에 쓰기 의도를 표현하는가?
- index range의 gap까지 잠글 수 있는가?
- 기다림, deadlock, duplicate key, serialization failure 중 어떤 실패가 정상 신호인가?
```

PostgreSQL 쪽에서 대응되는 함정은 `SERIALIZABLE`을 “모든 SELECT가 잠금을 강하게 잡는 모드”로 이해하는 것이다. PostgreSQL의 serializable은 snapshot isolation 위에 직렬화 불가능한 의존 관계를 감지하는 방식으로 설명된다. 그래서 블로킹이 늘어나는 것만 관찰하려고 하면 핵심을 놓칠 수 있다. 어떤 트랜잭션이 어떤 값을 읽고, 다른 트랜잭션이 그 판단을 무효화하는 값을 썼으며, 그 관계가 순환을 만들었는지 보는 것이 더 중요하다. 관측 표면도 lock wait만이 아니라 serialization failure, retry 성공률, 긴 read-only transaction, predicate lock 관련 view를 함께 봐야 한다.

아래는 운영 장애에서 두 엔진을 구분해 읽는 간단한 진단 흐름이다.

```text
증상: 같은 예약 API가 가끔 실패하거나 느려진다.

1. 먼저 실패 종류를 나눈다.
   - duplicate key / constraint violation
   - lock wait timeout
   - deadlock victim
   - serialization failure
   - application timeout

2. PostgreSQL이면 다음을 먼저 본다.
   - SHOW transaction_isolation;
   - SELECT * FROM pg_stat_activity WHERE state <> 'idle';
   - SELECT * FROM pg_locks;
   - 애플리케이션 로그의 SQLSTATE 40001 또는 40P01
   - retry가 있는지, retry가 같은 외부 side effect를 중복 실행하지 않는지

3. InnoDB이면 다음을 먼저 본다.
   - SELECT @@transaction_isolation;
   - SHOW ENGINE INNODB STATUS\G
   - performance_schema.data_locks / data_lock_waits
   - EXPLAIN으로 index range 확인
   - gap/next-key lock을 넓히는 조건식 또는 누락된 index

4. 둘 다에서 마지막에 확인한다.
   - SQL이 plain SELECT인지 locking read인지
   - 보호할 불변식이 unique/constraint로 표현 가능한지
   - 실패를 retry할 수 있게 transaction boundary가 짧은지
```

시니어 실무자의 함정 메모는 여기서 매우 구체적이어야 한다. `SERIALIZABLE`이나 `FOR UPDATE`를 붙이면 안전해진다는 말은 절반만 맞다. 안전해지는 대신 connection이 기다리고, deadlock 감지 비용이 생기고, 실패가 정상 경로로 올라온다. 특히 외부 결제 API, 메시지 발행, 파일 쓰기 같은 side effect를 트랜잭션 중간에 끼워 넣으면 retry가 곧 중복 실행 위험이 된다. DB 동시성 실패를 retry하려면 트랜잭션 안의 작업이 재실행 가능해야 하고, 외부 side effect는 commit 뒤 outbox나 idempotency key로 분리하는 편이 안전하다. 이 연결은 나중의 application-boundary DU와 이어지지만, isolation을 공부하는 지금부터 기억해야 한다.

PostgreSQL과 InnoDB 차이를 문서화하는 로컬 lab은 다음 형식을 권장한다. 단순히 결과값만 적지 말고, `읽기 종류`, `isolation`, `index`, `expected wait/failure`, `observability command`를 한 줄에 같이 둔다.

| case | engine | SQL shape | isolation | index prerequisite | expected observation | diagnostic command |
|---|---|---|---|---|---|---|
| repeated row read | PostgreSQL | plain SELECT | READ COMMITTED | PK | second read sees committed update | `SHOW transaction_isolation` |
| repeated row read | PostgreSQL | plain SELECT | REPEATABLE READ | PK | second read keeps first snapshot | `SHOW transaction_isolation` |
| write skew | PostgreSQL | plain SELECT + disjoint UPDATE | SERIALIZABLE | PK | one transaction may fail and need retry | app SQLSTATE log |
| range insert | InnoDB | SELECT ... FOR UPDATE | REPEATABLE READ | range index | insert into locked gap may wait | `SHOW ENGINE INNODB STATUS` |
| range insert | InnoDB | plain SELECT | REPEATABLE READ | range index | snapshot read alone does not mean writer is blocked | `performance_schema.data_locks` |
| duplicate booking | both | INSERT with unique key | any practical level | unique key | one writer fails at key conflict | constraint error log |

이 표가 제공하는 학습 효과는 “어느 DB가 더 좋다”가 아니다. 같은 업무 불변식을 보호할 때도 엔진마다 사고의 첫 지점이 다르다는 점이다. PostgreSQL에서는 snapshot과 serialization conflict, 명시적 row lock, constraint를 조합한다. InnoDB에서는 read view와 current/locking read, record/gap/next-key lock, unique constraint를 조합한다. 어느 쪽이든 좋은 설계는 실패 가능성을 없애는 척하지 않고, 실패를 작고 관측 가능하고 재시도 가능한 형태로 만든다.

버전과 설정의 경계도 남겨야 한다. 이 문서는 PostgreSQL current documentation과 MySQL 8.4 documentation을 기준으로 설명한다. 운영 환경이 Aurora, Cloud SQL, RDS, Percona, MariaDB, 오래된 MySQL 5.x, PostgreSQL 확장 설정을 쓰면 세부 관측 view와 기본값, lock wait timeout, deadlock logging, replication interaction이 달라질 수 있다. 그래서 문서의 결론은 “우리 운영에서도 무조건 이렇게 나온다”가 아니라 “이 축으로 확인하라”다. 실제 적용 전에는 `SHOW transaction_isolation`, `SELECT version()`, `SELECT @@version`, DB 파라미터, 실행 계획, 인덱스 정의, 실제 실패 로그를 함께 남기는 것이 PASS다.

마지막 replay 질문은 다음이다. 같은 `REPEATABLE READ`에서 세션 A가 `SELECT count(*) FROM reservation WHERE room_id=7`을 실행했고, 세션 B가 같은 room에 새 예약을 insert했다. A가 두 번째 plain SELECT에서 새 행을 보지 못했다면 안전한가? 답은 “아직 모른다”다. A가 보호하려던 것이 보고서 조회라면 충분할 수 있다. A가 예약 가능 여부를 판단하고 insert하려던 것이라면 plain snapshot만으로는 부족할 수 있다. unique constraint, locking read, guard row, serializable retry 중 무엇이 필요한지 다시 설계해야 한다. 이 “아직 모른다”를 말할 수 있어야 PostgreSQL과 InnoDB의 차이를 실제 설계 판단으로 가져갈 수 있다.

마이그레이션이나 DBMS 교체 리뷰에서는 이 차이를 체크리스트로 고정해야 한다. MySQL에서 PostgreSQL로 옮길 때 `REPEATABLE READ`라는 문자열만 매핑하면 InnoDB의 gap/next-key lock에 기대던 코드가 느슨해질 수 있다. PostgreSQL에서 MySQL로 옮길 때는 PostgreSQL `SERIALIZABLE`의 retry failure를 기대하던 코드가 InnoDB의 lock wait나 deadlock 형태로 보일 수 있다. 또한 MySQL의 `SELECT ... FOR UPDATE`가 어떤 index range를 잠그는지에 기대던 queue worker는 PostgreSQL에서 다른 관측과 다른 plan을 보일 수 있다. 따라서 포팅 리뷰는 다음 네 가지를 반드시 묻는다.

```text
DBMS 이동 isolation review

1. plain SELECT와 locking read를 구분했는가
   - report query
   - decision query
   - update target query

2. InnoDB gap/next-key lock에 기대던 업무 규칙이 있는가
   - 없는 row 확인 뒤 insert
   - range reservation
   - work queue claim

3. PostgreSQL serialization failure retry에 기대던 경계가 있는가
   - SQLSTATE 40001 분류
   - bounded retry
   - idempotent transaction body

4. constraint로 바꿀 수 있는 규칙을 isolation으로만 막고 있지 않은가
   - duplicate claim
   - one active subscription
   - one reservation per resource/time
```

이 리뷰의 PASS 신호는 “동일 isolation 이름을 설정했다”가 아니라, 업무 규칙마다 보호 장치가 다시 선택되었다는 것이다. 어떤 규칙은 unique key로 내려가고, 어떤 규칙은 guard row로 바뀌고, 어떤 규칙은 serializable retry를 유지하고, 어떤 규칙은 locking read와 index를 보강할 수 있다. FAIL 신호는 테스트가 단일 세션 CRUD만 확인하거나, 동시 두 세션에서 wait/failure/retry를 관측하지 않는 것이다. isolation 차이는 단위 테스트 한 줄보다 schedule과 로그에서 드러난다.

마지막으로 운영 문서에는 “정상 실패”를 명시해야 한다. PostgreSQL serializable에서 serialization failure가 발생할 수 있다는 사실, InnoDB에서 deadlock victim이나 lock wait timeout이 발생할 수 있다는 사실, unique constraint가 중복 요청을 막으면서 오류를 낼 수 있다는 사실은 모두 정상적인 동시성 제어의 일부다. 정상 실패를 오류 예외로만 남기면 장애처럼 보이고, 반대로 정상이라는 이유로 무한 retry하면 장애가 된다. 좋은 isolation 설계는 성공 경로보다 실패 경로가 더 선명하다. 실패를 분류하고, 재시도 가능성을 정하고, 재시도해도 외부 side effect가 중복되지 않게 만든다.

문서화할 때는 엔진별 차이를 “PostgreSQL은 낙관적, InnoDB는 비관적” 같은 한 줄 비유로 끝내지 않는다. 그런 말은 기억하기 쉽지만 너무 거칠다. PostgreSQL에도 row lock과 table lock이 있고, InnoDB에도 consistent read snapshot이 있다. 더 정확한 문장은 “PostgreSQL의 serializable 설명은 snapshot 위에서 직렬화 실패를 감지하는 축이 강하고, InnoDB의 repeatable-read 운영 체감은 locking read와 index range lock을 함께 읽어야 한다”에 가깝다. 이 정도로 써야 다른 isolation level, 다른 SQL shape, 다른 index를 만났을 때도 사고가 이어진다.

마지막 transfer check는 간단하다. `주문 수량 합계가 창고 수량을 넘으면 안 된다`는 요구를 받았을 때, 어떤 isolation level을 고를지 바로 말하지 말고 먼저 모델을 세 가지로 바꿔 보라. 첫째, 창고 row 하나에 남은 수량을 두고 조건부 UPDATE를 한다. 둘째, 주문 row들을 합산해 판단하고 serializable retry를 둔다. 셋째, 재고 예약 ledger와 unique/idempotency key를 둔다. 세 모델은 같은 업무 요구를 다른 충돌 지점으로 바꾼다. 이 연습을 할 수 있으면 PostgreSQL과 InnoDB의 차이는 단순 제품 지식이 아니라 설계 선택지가 된다.

팀 문서에는 이 차이를 테스트 이름에도 남긴다. `testRepeatableRead`보다 `postgresSerializableRejectsWriteSkew`, `innodbForUpdateWaitsOnIndexedGap`, `uniqueConstraintPreventsDuplicateReservation`처럼 기대하는 현상을 이름에 넣는 편이 낫다. 그러면 나중에 DBMS 버전이나 드라이버가 바뀌었을 때 실패한 테스트가 어떤 계약이 흔들렸는지 바로 말해 준다. 또한 테스트는 성공 결과만 보지 말고 실패 결과도 assertion으로 둔다. serialization failure가 나야 하는 실험에서 둘 다 커밋되면 FAIL이고, unique constraint가 중복 insert를 막아야 하는 실험에서 둘 다 성공하면 FAIL이다. 이렇게 “실패해야 안전한 지점”을 테스트하면 isolation 차이가 회귀 방지 자산이 된다.

결론적으로 PostgreSQL과 InnoDB의 차이는 제품 취향의 문제가 아니라 관측 가능한 동시성 계약의 차이다. 같은 업무 규칙을 두 엔진에 올릴 때는 읽기 종류, 잠금 범위, 실패 신호, 재시도 경계를 다시 매핑해야 한다. 이 매핑을 하지 않은 채 isolation 이름만 맞추면 테스트는 통과해도 운영의 경합 순간에 다른 결과가 나온다.

따라서 최종 산출물은 설정표가 아니라 두 세션 재현표여야 한다. 어떤 세션이 무엇을 읽고, 무엇을 기다리고, 무엇이 실패했는지 보이면 엔진 차이는 설명 가능한 지식이 된다.

그 표가 없으면 차이는 기억에 남지 않고, 다음 장애 때 다시 제품 이름만 외우게 된다.

재현표는 지식을 운영 언어로 정확히 바꾼다.
