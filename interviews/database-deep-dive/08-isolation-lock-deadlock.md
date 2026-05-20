# 격리 수준, 락, 데드락은 어떤 동시성 실패를 막고 어떤 실패를 남기는가?

격리 수준은 transaction들이 서로의 변경을 어떤 방식으로 관측할 수 있는지 정하는 계약이고, lock은 그 계약을 실현하기 위해 DBMS가 row, index range, table, predicate, advisory key 같은 자원에 충돌 규칙을 붙이는 실행 수단이다. 둘은 같은 말이 아니다. Isolation은 사용자에게 약속하는 관측 결과이고, lock은 그 약속을 만들기 위해 쓰는 여러 도구 중 하나다. MVCC, predicate conflict tracking, next-key lock, serialization failure, retry도 모두 같은 목표 주변에 있다.

면접에서 이 주제를 잘 답하려면 dirty read, non-repeatable read, phantom read 표를 외운 뒤 멈추면 안 된다. `PostgreSQL READ COMMITTED`, `PostgreSQL REPEATABLE READ`, `PostgreSQL SERIALIZABLE`, `MySQL/InnoDB REPEATABLE READ`가 같은 이름과 다른 구현을 가질 수 있음을 말해야 한다. 또한 row lock, gap lock, next-key lock, predicate lock, advisory lock, table lock을 보호하는 대상별로 나누고, latch와 lock의 차이, deadlock wait graph와 retry boundary까지 이어야 한다.

## 2-5분 개요

짧게 답하면 이렇게 말할 수 있다. SQL isolation level은 동시 transaction이 서로 어떤 중간 결과나 새 row를 볼 수 있는지 정하는 계약이다. DBMS는 그 계약을 MVCC snapshot, row lock, range lock, predicate 성격의 conflict tracking, serialization failure 같은 방식으로 구현한다. Lock은 그중 하나의 실행 수단이고, deadlock은 여러 transaction이 서로 가진 자원을 기다리며 cycle을 만들 때 발생한다.

가장 흔한 오해는 `격리 수준을 높이면 안전하다`와 `lock을 걸면 안전하다`를 같은 말로 보는 것이다. 격리 수준을 높이면 anomaly를 줄일 수 있지만 abort/retry, lock wait, throughput 저하가 생긴다. Lock을 더 넓게 잡으면 특정 충돌은 막을 수 있지만 deadlock과 contention이 늘 수 있다. 반대로 lock을 너무 좁게 잡거나 index가 없어 range lock이 넓어지면, 의도와 다른 row까지 막거나 phantom을 놓칠 수 있다.

PostgreSQL은 MVCC 기반 plain read가 writer를 막지 않는 성격이 강하다. Serializable에서는 Serializable Snapshot Isolation 계열로 위험한 read/write dependency를 감지해 serialization failure를 낼 수 있다. MySQL/InnoDB는 index record lock, gap lock, next-key lock을 통해 range conflict와 phantom을 다룬다. InnoDB의 `row lock`은 실제로 index record와 range에 걸리는 경우가 많으므로, 어떤 index로 접근했는지가 lock 범위를 바꾼다.

Deadlock은 DB가 이상한 상태가 되었다는 뜻이 아니라, 동시 transaction scheduling에서 생길 수 있는 정상적인 실패 모드다. T1이 A를 잡고 B를 기다리며, T2가 B를 잡고 A를 기다리면 cycle이 생긴다. DBMS는 deadlock을 감지하면 보통 한 transaction을 victim으로 골라 abort한다. 애플리케이션은 이 오류를 재시도 가능한 경계로 다뤄야 하고, 같은 자원을 항상 같은 순서로 잡는 설계로 빈도를 줄여야 한다.

면접 답변의 핵심은 네 가지다. 첫째, isolation은 관측 계약이고 lock은 구현 수단이다. 둘째, DBMS별로 같은 isolation 이름의 실제 동작이 다르다. 셋째, lock은 보호 대상과 mode를 기준으로 말해야 한다. 넷째, deadlock은 wait graph cycle이며 retry와 lock ordering이 설계의 일부다.

## 먼저 잡아야 할 작은 모델

작은 모델은 좌석 예약이다. 공연에 남은 좌석이 하나뿐인데 두 transaction이 동시에 예약하려고 한다.

```text
seat table
  show_id=1, seat_no='A1', status='FREE'

T1
  SELECT count(*) FROM seat WHERE show_id=1 AND status='FREE'; -- 1
  INSERT INTO reservation(show_id, user_id) VALUES (1, 'userA');
  UPDATE seat SET status='SOLD' WHERE show_id=1 AND seat_no='A1';

T2
  SELECT count(*) FROM seat WHERE show_id=1 AND status='FREE'; -- 1
  INSERT INTO reservation(show_id, user_id) VALUES (1, 'userB');
  UPDATE seat SET status='SOLD' WHERE show_id=1 AND seat_no='A1';
```

만약 실제로 같은 seat row를 update한다면 row lock이나 write conflict가 둘 중 하나를 막을 수 있다. 하지만 invariant가 `남은 좌석 count > 0이면 예약 row를 추가한다`처럼 aggregate/predicate에 걸려 있으면 단일 row lock만으로 충분하지 않을 수 있다. 특히 예약 row만 insert하고 seat 상태를 늦게 바꾸거나, 조건이 range query로 표현되면 phantom과 write skew가 등장한다.

다음 trace는 row lock과 predicate 보호의 차이를 보여 준다.

```text
row lock으로 충분한 경우
  UPDATE seat
  SET status='SOLD'
  WHERE show_id=1 AND seat_no='A1' AND status='FREE';

  affected rows = 1이면 성공
  affected rows = 0이면 이미 팔림

predicate 보호가 필요한 경우
  SELECT count(*) WHERE show_id=1 AND status='FREE';
  count가 0보다 크면 reservation row INSERT

  이때 같은 predicate를 믿은 두 transaction이 서로 다른 row를 insert할 수 있다.
```

이 모델은 격리 수준 표보다 강하다. Dirty read, non-repeatable read, phantom read, write skew가 모두 `무엇을 읽고 그 읽은 결과를 근거로 무엇을 쓰는가`라는 질문으로 돌아오기 때문이다. Lock도 마찬가지다. 어떤 row를 잡았는지, 어떤 index range를 잡았는지, table 전체를 막았는지, advisory key만 잡았는지에 따라 보호하는 invariant가 달라진다.

## 깊은 메커니즘

### Isolation level은 금지할 현상을 정하지만 제품별 의미가 다르다

SQL 표준은 isolation을 dirty read, non-repeatable read, phantom read 같은 현상으로 설명한다. Dirty read는 commit되지 않은 다른 transaction의 변경을 읽는 것이다. Non-repeatable read는 같은 row를 두 번 읽었는데 중간 commit 때문에 값이 달라지는 것이다. Phantom read는 같은 조건으로 다시 검색했을 때 새 row가 나타나는 것이다.

하지만 실제 DBMS는 표준 표를 그대로 구현하는 것이 아니라 자기 저장 구조와 동시성 제어 방식에 맞게 isolation을 제공한다. PostgreSQL의 `READ COMMITTED`는 statement마다 snapshot을 새로 만들기 때문에 같은 transaction 안의 반복 SELECT가 다른 committed 값을 볼 수 있다. PostgreSQL의 `REPEATABLE READ`는 phantom까지 상당히 막는 snapshot isolation 성격을 가진다. PostgreSQL의 `SERIALIZABLE`은 SSI로 dependency cycle을 감지해 serialization failure를 낼 수 있다.

MySQL/InnoDB는 기본 `REPEATABLE READ`에서 consistent read는 transaction read view를 유지하지만, locking read와 range scan에서는 next-key lock이 중요해진다. `READ COMMITTED`에서는 gap lock 사용이 줄어드는 등 lock behavior가 달라질 수 있다. 따라서 `REPEATABLE READ면 phantom이 발생한다/안 한다`처럼 제품 없이 단정하는 답은 위험하다.

### Row lock은 논리 row가 아니라 접근 경로와 연결된다

`row lock`이라는 말은 쉽게 들리지만, InnoDB에서는 실제로 index record lock을 중심으로 이해해야 한다. `WHERE id = 10 FOR UPDATE`가 primary key index를 사용하면 해당 index record에 lock을 걸 수 있다. 그런데 조건에 맞는 index가 없어서 table scan을 하거나, range scan을 넓게 하면 lock 범위가 커질 수 있다. Lock은 SQL 문장만이 아니라 실행 계획과 index 구조에 영향을 받는다.

```text
좋은 접근 경로
  WHERE user_id = 100 AND status = 'READY'
  index(user_id, status)를 사용
  필요한 record/range만 lock할 가능성이 높다.

나쁜 접근 경로
  WHERE lower(email) = 'a@example.com'
  usable index 없음
  많은 row를 읽으며 더 넓은 lock/wait를 만들 수 있다.
```

PostgreSQL에서도 row-level lock은 `FOR UPDATE`, `FOR NO KEY UPDATE`, `FOR SHARE`, `FOR KEY SHARE`처럼 mode가 나뉜다. Table-level lock도 mode별 충돌 관계가 다르다. 면접에서 `SELECT는 lock을 안 겁니다`라고만 말하면 부족하다. Plain SELECT, locking SELECT, DDL, foreign key check, update/delete가 만드는 lock을 나누어야 한다.

### Gap lock과 next-key lock은 존재하지 않는 row를 둘러싼 충돌을 다룬다

Phantom은 기존 row 하나를 잠그는 것만으로 막히지 않는다. `WHERE price BETWEEN 100 AND 200`에 해당하는 row를 읽은 뒤 그 범위에 새 row가 insert되면, 두 번째 읽기에서 결과가 달라질 수 있다. InnoDB next-key lock은 index record와 그 앞 gap을 함께 잠그는 방식으로 range에 새 row가 끼어드는 일을 막을 수 있다.

```text
index values: 10, 20, 30
query: WHERE value BETWEEN 11 AND 29 FOR UPDATE

record 20과 그 주변 gap을 잠그면
다른 transaction이 value=15 또는 value=25를 insert하려 할 때 기다릴 수 있다.
```

이 설명에서 index가 중요하다. Range를 어떤 index로 읽었는지에 따라 gap과 next-key lock 범위가 달라진다. Unique index로 정확히 하나의 record를 찾는 경우와 non-unique range scan은 lock 범위가 다르다. Index가 없으면 DBMS가 더 많은 record를 검사해야 하고, 그만큼 충돌 범위가 넓어질 수 있다.

PostgreSQL은 InnoDB식 gap/next-key lock 용어로 설명하지 않는다. Serializable에서는 predicate 성격의 read/write dependency를 추적하고, 위험한 구조가 발견되면 serialization failure를 낸다. 그래서 PostgreSQL과 InnoDB를 비교할 때는 `phantom을 막는다`는 목표와 `gap을 잠근다`, `dependency를 감지하고 abort한다`는 구현을 분리해야 한다.

### Predicate lock과 SSI는 읽은 조건 자체를 보호하려는 접근이다

Predicate는 `status='FREE'인 좌석`, `balance < 0인 계좌`, `doctor on_call=true인 row`처럼 row 하나가 아니라 조건이다. Serializable isolation이 진짜 serial execution처럼 보이려면 transaction이 읽은 predicate와 나중에 들어오는 write 사이의 충돌도 다뤄야 한다.

PostgreSQL Serializable Snapshot Isolation은 단순히 모든 predicate를 물리적으로 잠그는 방식이 아니라, snapshot isolation 위에서 read/write dependency를 추적해 serialization anomaly 가능성이 있는 구조를 감지한다. 감지되면 한 transaction을 abort시켜 실제 serial order를 만들 수 없는 결과를 막는다. 이때 애플리케이션은 serialization failure를 retry해야 한다.

이 지점이 면접에서 좋다. `Serializable이면 lock을 많이 걸어서 느립니다`라고만 말하면 PostgreSQL의 SSI를 놓친다. `Serializable이면 retry가 필요 없습니다`라고 말하면 더 틀린다. Strong isolation을 제공하려면 DB가 어떤 transaction을 실패시킬 수 있고, 애플리케이션은 그 실패를 정상 경로로 설계해야 한다.

### Advisory lock은 DB가 의미를 모르는 이름 붙은 락이다

Advisory lock은 row나 table에 자동으로 묶인 lock이 아니라 애플리케이션이 정한 key에 대해 DB lock manager를 빌리는 방식이다. PostgreSQL의 advisory lock이 대표적이다. 예를 들어 `tenant_id=10`에 대한 batch job이 동시에 하나만 돌게 하고 싶다면 tenant id를 advisory lock key로 사용할 수 있다.

이 방식은 유용하지만 DB가 그 key의 업무 의미를 이해하지는 않는다. 어떤 code path는 advisory lock을 잡고 어떤 path는 잡지 않으면 보호가 깨진다. Transaction-level advisory lock과 session-level advisory lock도 생명주기가 다르다. Connection pool을 쓰는 환경에서 session-level lock을 잘못 다루면 lock이 예상보다 오래 남거나 엉뚱한 요청에 영향을 줄 수 있다.

면접 답변에서는 advisory lock을 `분산락 대체재`처럼 과장하면 안 된다. 같은 DB에 연결된 참여자들이 같은 규칙을 지킬 때는 강력한 도구가 될 수 있지만, DB 밖 시스템이나 다른 storage를 자동으로 보호하지 않는다. Lock key 설계, release boundary, timeout, deadlock 가능성을 함께 말해야 한다.

### Latch와 lock은 보호하는 층위가 다르다

Latch는 DBMS 내부 자료구조를 아주 짧게 보호하는 동기화 장치로 이해하면 된다. Buffer page, B-tree page, lock table 같은 내부 구조가 동시에 깨지지 않게 하는 데 쓰인다. Transaction lock은 사용자 transaction 사이의 논리 충돌을 조정한다. 둘을 같은 말로 쓰면 운영 지표를 잘못 해석한다.

예를 들어 lock wait가 길다는 것은 어떤 transaction이 row/table/range lock을 들고 있어서 다른 transaction이 기다린다는 뜻일 수 있다. Latch contention은 CPU와 메모리 내부 구조 경쟁으로 나타날 수 있다. 해결도 다르다. Lock wait는 긴 transaction, 잘못된 update 순서, index 부재, batch 크기, isolation level을 본다. Latch contention은 hot page, buffer manager, index split, internal contention을 본다.

면접에서 latch를 깊게 몰라도 된다. 다만 `lock은 사용자 transaction의 논리 자원 충돌, latch는 엔진 내부 자료구조 보호`라고 구분하면 된다. 이 구분만 있어도 wait event를 해석할 때 한 단계 더 정확해진다.

### Deadlock은 wait graph cycle이고 retry는 설계의 일부다

Deadlock은 간단한 trace로 설명할 수 있다.

```text
T1: UPDATE account SET ... WHERE id='A'; -- A lock 보유
T2: UPDATE account SET ... WHERE id='B'; -- B lock 보유
T1: UPDATE account SET ... WHERE id='B'; -- B를 기다림
T2: UPDATE account SET ... WHERE id='A'; -- A를 기다림

wait graph:
  T1 -> T2
  T2 -> T1
cycle 발생
```

DBMS는 이런 cycle을 감지하면 보통 한 transaction을 victim으로 골라 rollback한다. Victim 선택 기준은 제품별로 다르다. 애플리케이션은 deadlock 오류를 사용자에게 곧바로 실패로 보여 주기보다, transaction 전체가 idempotent하거나 안전하게 다시 실행될 수 있으면 retry해야 한다. 단, retry는 무한 반복이 아니라 backoff, 최대 횟수, observability, side effect 분리와 함께 설계해야 한다.

Deadlock 예방의 기본은 lock order를 통일하는 것이다. 여러 row를 갱신해야 한다면 모든 code path가 key를 정렬한 순서로 update한다. Batch job도 random order로 update하지 않는다. Foreign key가 참조하는 parent/child table 순서, inventory와 order table 순서, 계좌 A/B 순서를 통일한다. 그래도 DB 내부 lock과 index range 때문에 모든 deadlock을 제거할 수는 없으므로 retry는 남긴다.


### Anomaly별로 방어 수단이 다르다

Dirty read는 commit되지 않은 값을 읽는 문제다. 대부분의 현대 OLTP 설정에서는 기본 격리 수준에서 dirty read를 허용하지 않는다. Non-repeatable read는 같은 row를 다시 읽었을 때 값이 달라지는 문제다. Statement snapshot을 새로 만드는 `READ COMMITTED`에서는 자연스럽게 보일 수 있고, transaction snapshot을 유지하면 줄어든다. Phantom은 같은 predicate query를 다시 실행했을 때 새 row가 나타나는 문제다. Write skew는 각 transaction이 서로 다른 row를 쓰지만 함께 보면 invariant가 깨지는 문제다.

이 현상들은 같은 lock 하나로 모두 해결되지 않는다. 단일 row lost update는 row lock이나 optimistic version column으로 막기 쉽다. Range phantom은 index range lock, predicate conflict tracking, serializable retry가 필요할 수 있다. Write skew는 guard row를 도입해 같은 row를 lock하게 만들거나, constraint로 invariant를 DB가 알 수 있게 바꾸거나, serializable isolation과 retry를 쓸 수 있다. 따라서 면접에서는 현상 이름 뒤에 `어떤 읽기 결과를 믿고 어떤 쓰기를 했는가`를 붙여야 한다.

```text
lost update
  T1 reads row X=10
  T2 reads row X=10
  T1 writes X=11
  T2 writes X=11
  expected X=12 but got X=11

write skew
  T1 reads predicate P, writes row A
  T2 reads predicate P, writes row B
  each write is different row
  combined result breaks invariant
```

### Lock wait를 줄이는 일과 correctness를 지키는 일은 함께 봐야 한다

운영에서 lock wait가 늘면 누구나 빨리 줄이고 싶어 한다. 하지만 lock을 줄이는 방향이 invariant를 깨면 안 된다. 예를 들어 `SELECT ... FOR UPDATE`를 제거하면 대기 시간은 줄 수 있지만, 같은 재고를 두 transaction이 동시에 차감할 수 있다. 반대로 모든 작업에 table lock을 걸면 correctness는 단순해질 수 있지만 throughput이 무너진다.

좋은 설계는 보호해야 할 invariant를 먼저 고른 뒤 가장 좁은 자원으로 보호한다. 재고 한 SKU의 수량이면 그 SKU row를 조건부 update한다. 한 사용자 계정의 상태 전이면 user id 기준으로 같은 row 또는 advisory key를 잡는다. 특정 기간의 중복 예약이면 `(resource_id, time_range)`를 constraint나 range lock으로 표현할 수 있는지 본다. 보호 대상이 명확해질수록 lock 범위도 줄어든다.

### Index 설계는 lock 설계이기도 하다

InnoDB에서 range 조건을 locking read로 실행할 때 usable index가 없으면 많은 record와 gap을 검사하며 넓은 lock을 만들 수 있다. PostgreSQL에서도 index가 없으면 더 많은 row를 방문하고 더 오래 transaction이 유지될 수 있다. Lock 문제를 해결하려고 timeout만 늘리면 근본 원인을 놓친다. Query가 어떤 index를 사용해 어떤 후보 row를 읽는지 확인해야 한다.

```text
문제 query
  UPDATE coupon
  SET used=true
  WHERE campaign_id=? AND user_id=? AND used=false;

필요한 관찰
  campaign_id, user_id, used 순서의 index가 있는가?
  affected row가 0일 때 재시도 또는 실패 처리가 있는가?
  같은 campaign의 많은 row를 불필요하게 scan하며 lock하지 않는가?
```

Index가 correctness를 직접 보장할 때도 있다. Unique index는 중복 insert를 막는다. Conditional unique index나 exclusion constraint가 있는 DBMS에서는 `활성 예약은 같은 시간 구간에 하나만` 같은 invariant를 DB에 내릴 수 있다. 이 경우 애플리케이션 lock보다 DB constraint가 더 단순하고 강할 수 있다.

### Retry는 side effect와 함께 설계해야 한다

Deadlock이나 serialization failure를 retry하면 된다고 말하는 것은 절반이다. Retry할 transaction 안에 외부 API 호출이나 메시지 발행이 들어 있으면 재시도 때 side effect가 중복될 수 있다. 따라서 retry 가능한 transaction function은 DB 내부 변경과 idempotency check를 중심으로 만들고, 외부 side effect는 outbox나 별도 멱등성 키로 분리하는 편이 안전하다.

```text
retry-safe boundary
  BEGIN
    check request_id unique key
    update rows
    insert outbox event
  COMMIT
  publisher sends event with event_id

retry-risk boundary
  BEGIN
    call external payment API
    update rows
  deadlock -> rollback -> retry calls payment API again
```

이 구분을 하면 isolation-lock 주제가 transaction boundary 문서와 연결된다. DB가 한 transaction을 abort시키는 것은 correctness를 지키기 위한 정상 동작일 수 있다. 애플리케이션이 그 abort를 재실행할 수 없는 형태로 side effect를 섞어 놓으면, DB의 correctness mechanism이 사용자 장애로 바뀐다.

### Deadlock 분석은 victim query 하나로 끝나지 않는다

Deadlock log에는 보통 victim이 된 transaction의 query가 눈에 잘 띈다. 하지만 root cause는 상대 transaction의 lock order, index range, foreign key check, trigger, batch ordering에 있을 수 있다. 따라서 deadlock 분석은 양쪽 transaction이 잡은 lock과 기다린 lock을 모두 본다.

분석 순서는 단순하다. 먼저 wait graph의 node와 edge를 적는다. 각 node가 어떤 transaction인지, 어떤 statement에서 어떤 lock을 보유했는지, 어떤 lock을 기다렸는지 쓴다. 그 다음 왜 그 자원을 그 순서로 잡았는지 code path를 찾는다. 마지막으로 순서를 통일할 수 있는지, index로 범위를 줄일 수 있는지, transaction을 짧게 나눌 수 있는지, retry가 안전한지 결정한다.

```text
T1 owns order:10, waits inventory:skuA
T2 owns inventory:skuA, waits order:10

possible fix
  every path locks order first, then inventory
  or every path locks inventory first, then order
  choose one and enforce it in service layer
```

### Table lock과 metadata lock은 장애 모양이 다르다

Row lock만 생각하면 DDL과 migration 장애를 놓친다. PostgreSQL의 `ACCESS EXCLUSIVE` lock은 많은 작업을 막을 수 있고, MySQL의 metadata lock은 오래 열린 transaction 때문에 DDL이 기다리거나 뒤따르는 query까지 막히는 상황을 만들 수 있다. 운영에서 `ALTER TABLE` 하나가 서비스 전체를 멈춘 것처럼 보이는 이유가 여기에 있다.

Migration을 설계할 때는 lock level, lock 획득 시점, timeout, online DDL 지원 범위, backfill batch 크기, rollback plan을 확인해야 한다. 면접에서 isolation과 lock을 application transaction만으로 설명하면 schema change의 실제 위험을 놓친다. DBMS별 online DDL 능력이 좋아졌더라도, 모든 변경이 무중단인 것은 아니다.

## DBMS별 경계

### PostgreSQL

PostgreSQL은 table-level lock mode와 row-level lock mode를 공식 문서에서 자세히 나눈다. `ACCESS SHARE`, `ROW EXCLUSIVE`, `SHARE`, `ACCESS EXCLUSIVE` 같은 table lock은 DDL과 DML 충돌을 이해하는 데 중요하다. Row-level lock도 `FOR UPDATE`, `FOR NO KEY UPDATE`, `FOR SHARE`, `FOR KEY SHARE`로 나뉜다. Plain SELECT는 보통 ACCESS SHARE table lock을 잡지만 row write를 막는 lock은 아니다.

Serializable은 PostgreSQL에서 특히 중요하다. MVCC snapshot 위에서 Serializable Snapshot Isolation이 동작하고, dangerous structure가 생기면 serialization failure를 낼 수 있다. 따라서 PostgreSQL Serializable 답변에는 `retry`가 반드시 들어가야 한다. `lock을 다 걸어서 순서대로 실행합니다`라고만 답하면 PostgreSQL의 실제 전략을 놓친다.

Deadlock 관측은 `pg_locks`, `pg_stat_activity`, server log의 deadlock detail, wait event를 본다. Blocker와 waiter를 찾고, 어떤 query가 어떤 lock을 들고 있는지, transaction age가 얼마나 되는지, 같은 자원을 어떤 순서로 잡는지 확인한다.

### MySQL/InnoDB

InnoDB는 row-level locking을 index record 중심으로 봐야 한다. Shared lock, exclusive lock, intention lock, record lock, gap lock, next-key lock, insert intention lock, auto-inc lock 같은 개념이 있다. 특히 next-key lock은 phantom 방어와 range conflict 설명의 중심이다.

InnoDB deadlock 분석에서는 `SHOW ENGINE INNODB STATUS`의 latest detected deadlock이 자주 쓰인다. 어떤 transaction이 어떤 lock을 기다렸고 어떤 lock을 보유했는지 나온다. Performance Schema의 data locks 관련 view도 함께 볼 수 있다. MySQL 8.4 기준 문서를 보면 isolation level별 locking read와 consistent read 차이를 공식적으로 확인할 수 있다.

InnoDB에서는 index 설계가 lock 설계다. 같은 SQL이라도 적절한 index가 있으면 좁은 range만 잠글 수 있고, index가 없거나 조건이 넓으면 불필요한 row와 gap까지 영향을 줄 수 있다. 그래서 lock 문제를 볼 때 query text만 보지 말고 execution plan과 index를 함께 봐야 한다.

### 애플리케이션 경계

애플리케이션은 DB isolation을 보완할 수 있지만 대체하지는 않는다. Optimistic locking의 version column은 lost update를 잡는 데 좋다. Unique constraint는 중복 insert를 막는 데 좋다. Queue serialization은 특정 aggregate의 write를 순서화하는 데 좋다. Distributed lock은 DB 밖 자원까지 포함한 작업 조율에 도움을 줄 수 있다. 그러나 각각은 실패 경계가 다르다.

예를 들어 optimistic lock은 update 시 `WHERE id=? AND version=?` 조건으로 affected row count를 확인한다. 충돌이 나면 사용자가 다시 읽고 재시도해야 한다. 이것은 DB row lock을 오래 들고 있지 않는 장점이 있지만, conflict가 많은 hot row에서는 retry가 많아질 수 있다. Distributed lock은 lock service와 네트워크 failure, lease 만료, clock skew, fencing token을 봐야 한다. DB lock보다 무조건 낫다고 말하면 안 된다.

## 직접 재생해 보기

1. Deadlock을 일부러 만든다.

    Session A:

    ```sql
    BEGIN;
    UPDATE account SET balance = balance - 10 WHERE id = 'A';
    -- Session B가 B를 update할 때까지 대기
    UPDATE account SET balance = balance + 10 WHERE id = 'B';
    ```

    Session B:

    ```sql
    BEGIN;
    UPDATE account SET balance = balance - 10 WHERE id = 'B';
    UPDATE account SET balance = balance + 10 WHERE id = 'A';
    ```

    PASS 신호는 DB가 deadlock을 감지하고 한 transaction을 abort하는 것이다. 그 다음 두 session 모두 A, B 순서로 update하도록 바꾸면 deadlock 가능성이 줄어드는지 확인한다.

2. InnoDB range lock을 확인한다.

    적절한 index가 있는 table에서 `SELECT ... WHERE score BETWEEN 10 AND 20 FOR UPDATE`를 실행하고 다른 session에서 score=15 row를 insert해 본다. 기다림이 발생하면 어떤 index range가 보호되는지 설명한다. Index를 제거하거나 다른 조건으로 바꾸었을 때 lock 범위가 달라지는지도 `EXPLAIN`과 lock view로 본다.

3. PostgreSQL Serializable retry를 만든다.

    Write skew 예시를 사용한다. 두 transaction이 같은 predicate를 읽고 서로 다른 row를 update해 invariant를 깨는 패턴을 만든다. `SERIALIZABLE`에서 한 transaction이 serialization failure로 실패하면 PASS다. 애플리케이션 관점에서는 같은 transaction function을 재실행해야 한다.

4. Advisory lock 생명주기를 비교한다.

    PostgreSQL에서 transaction-level advisory lock과 session-level advisory lock을 각각 잡고 commit/rollback/connection close 때 어떻게 해제되는지 확인한다. Connection pool에서 session-level lock을 쓰면 왜 위험한지 설명할 수 있어야 한다.

5. Latch와 lock 지표를 분리해 말한다.

    실제 환경이 없다면 wait event 예시를 표로 만든다. `row lock wait`, `metadata lock`, `buffer content lock`, `IO wait`가 모두 같은 해결책을 갖지 않는다는 점을 적는다. PASS는 대기 이름을 보고 바로 쿼리를 죽이는 것이 아니라 보호 대상과 owner를 먼저 찾는 것이다.

## 면접 꼬리 질문

- REPEATABLE READ면 phantom read가 항상 발생하나요?

    제품 없이 단정하면 안 된다. 표준 설명과 실제 DBMS 구현이 다르다. PostgreSQL Repeatable Read는 snapshot isolation 성격으로 phantom을 다르게 다루고, InnoDB Repeatable Read는 next-key lock과 consistent read가 함께 등장한다. 공식 문서와 실험으로 제품별 경계를 확인해야 한다.

- Row lock은 정말 row 자체에 걸리나요?

    DBMS마다 설명이 다르지만 InnoDB에서는 index record lock으로 이해해야 한다. 어떤 index를 타는지와 range 조건이 lock 범위를 바꾼다. PostgreSQL에서도 row lock mode와 table lock mode가 따로 있다.

- Deadlock이 발생하면 무조건 DB 문제인가요?

    아니다. 여러 transaction이 자원을 다른 순서로 잡으면 애플리케이션 설계 때문에 생길 수 있다. DB는 deadlock을 감지하고 하나를 abort할 뿐이다. 일관된 lock order, 짧은 transaction, 적절한 index, retry 설계가 필요하다.

- Serializable을 쓰면 retry가 필요 없나요?

    반대다. Serializable이 강한 격리를 제공하려면 어떤 transaction을 실패시킬 수 있다. PostgreSQL Serializable에서는 serialization failure retry가 정상 경로다.

- Latch와 lock의 차이는 무엇인가요?

    Latch는 DBMS 내부 자료구조를 짧게 보호하는 동기화 장치이고, transaction lock은 사용자 transaction 사이의 논리 자원 충돌을 조정한다. 관측 지표와 해결책이 다르다.

- Advisory lock을 쓰면 분산락 문제가 해결되나요?

    같은 DB와 같은 규칙을 공유하는 참여자 사이에서는 유용하다. 하지만 DB 밖 시스템, session 생명주기, connection pool, timeout, fencing 문제는 별도로 봐야 한다. Advisory lock은 DB가 의미를 모르는 key에 대한 약속이다.

## 함정 질문

- Isolation level 표만 외우고 DBMS 차이를 무시하는 답

    표는 출발점일 뿐이다. 같은 이름도 PostgreSQL과 InnoDB에서 snapshot, lock, phantom, retry behavior가 다를 수 있다. 제품별 공식 문서와 재생 실험으로 닫아야 한다.

- Lock을 많이 걸면 안전하고 비용은 없다고 말하는 답

    넓은 lock은 anomaly를 줄일 수 있지만 contention과 deadlock, throughput 저하를 만든다. 필요한 invariant를 가장 좁고 명확한 경계로 보호하는 것이 중요하다.

- Deadlock을 timeout과 같은 말로 보는 답

    Timeout은 오래 기다려서 포기하는 것이고, deadlock은 wait graph cycle이다. Deadlock은 한쪽을 abort해야 진행된다. 로그와 wait graph를 보고 구분해야 한다.

- Predicate 문제를 row lock 하나로 해결하려는 답

    조건에 새 row가 들어오는 phantom이나 write skew는 기존 row 하나를 잡는 것만으로 막히지 않을 수 있다. Range/predicate 보호, serializable retry, guard row, constraint 설계를 봐야 한다.

- Application distributed lock을 DB isolation의 상위 호환으로 말하는 답

    Distributed lock은 다른 실패 모드를 가진다. Lease 만료, lock service 장애, fencing token, clock 문제를 설계하지 않으면 DB lock보다 더 위험할 수 있다.

## 답변을 더 단단하게 만드는 판단 흐름

격리 수준과 lock 질문은 항상 `어떤 불변식이 동시에 깨질 수 있는가`에서 시작한다. 단일 row만 바뀌는 문제라면 row lock이나 조건부 update로 충분할 수 있다. 그러나 좌석 예약, 재고 총량, 당직 의사 최소 1명, 쿠폰 1회 사용처럼 여러 row나 predicate가 만드는 불변식은 단순 row lock만으로 닫히지 않을 수 있다. 면접에서는 먼저 불변식의 모양을 말한 뒤, 그 불변식을 어떤 자원 단위로 보호할지 설명해야 한다.

```text
불변식 모양 -> 보호할 자원 단위 -> 필요한 관측 계약 -> lock/conflict/retry 전략
```

예를 들어 `남은 좌석이 있으면 예약한다`는 조건은 특정 row 하나가 아니라 `show_id=1 AND status='free'`라는 predicate에 기대고 있다. 두 transaction이 같은 predicate를 읽고 서로 다른 reservation row를 insert하면 단일 row conflict가 없는데도 oversell이 생길 수 있다. 이때 가능한 방어는 여러 가지다. 좌석 row 자체를 먼저 잡고 상태를 바꾸거나, show별 guard row를 lock하거나, unique/exclusion constraint로 최종 중복을 막거나, Serializable에서 실패한 transaction을 retry할 수 있다. 중요한 것은 수단 이름이 아니라 어떤 불변식을 어느 계층에서 닫는가다.

InnoDB를 말할 때는 index와 lock 범위를 함께 말해야 한다. InnoDB row lock은 실제로 index record를 통해 잡히는 경우가 많다. 조건에 맞는 index가 있으면 좁은 range를 잠글 수 있지만, index가 없거나 조건이 넓으면 더 많은 record와 gap이 잠길 수 있다. Gap lock과 next-key lock은 phantom insert를 막는 데 쓰일 수 있지만, 예상보다 많은 insert가 기다리는 원인이 되기도 한다. 그래서 lock wait 문제를 볼 때는 SQL 문장만 보지 말고 어떤 index range를 훑었는지 EXPLAIN과 lock 관측을 함께 봐야 한다.

PostgreSQL을 말할 때는 `읽기가 writer를 막지 않는다`와 `Serializable에서 abort가 날 수 있다`를 같이 말해야 한다. PostgreSQL의 MVCC는 plain read를 강하게 살려 주지만, Serializable Snapshot Isolation 계열에서는 위험한 read/write dependency가 감지되면 serialization failure가 발생할 수 있다. 이것은 DB 오류가 아니라 정합성을 지키기 위한 정상적인 거절일 수 있다. 따라서 애플리케이션은 그 transaction을 idempotent하게 재시도할 수 있어야 한다.

Deadlock은 이 논의를 현실로 끌고 내려온다. Deadlock은 오래 기다린다는 뜻이 아니라 wait-for graph에 cycle이 생겼다는 뜻이다. T1이 A를 잡고 B를 기다리고, T2가 B를 잡고 A를 기다리면 둘 중 하나는 abort되어야 한다. DBMS는 보통 deadlock detector로 cycle을 찾고 victim을 고른다. 애플리케이션은 victim이 된 transaction을 재시도할 수 있어야 하고, 가능하면 모든 코드 경로가 같은 순서로 lock을 잡도록 만들어야 한다.

```text
나쁜 순서
  request 1: lock account A -> lock account B
  request 2: lock account B -> lock account A

나은 순서
  모든 transfer: smaller account id 먼저 lock -> larger account id lock
```

이 예시는 단순하지만 강력하다. Deadlock을 완전히 없애지는 못해도, 같은 자원 집합을 같은 순서로 잡으면 cycle 가능성이 크게 줄어든다. Batch update도 마찬가지다. `WHERE id IN (...)` 목록을 random하게 처리하면 transaction마다 lock 순서가 달라질 수 있다. key를 정렬하고 작은 batch로 나누면 lock 보유 시간과 cycle 가능성을 줄일 수 있다.

격리 수준을 높이는 선택은 항상 비용과 함께 말해야 한다. READ COMMITTED는 동시성이 좋지만 같은 transaction 안에서 새 commit을 다시 볼 수 있다. REPEATABLE READ는 반복 조회 일관성을 주지만 제품별 phantom 처리와 lock 비용이 다르다. SERIALIZABLE은 강한 관측 계약을 주지만 serialization failure와 retry 비용이 생길 수 있다. 따라서 답변은 `SERIALIZABLE을 쓰면 됩니다`가 아니라 `이 업무는 어떤 anomaly를 허용하지 못하고, 그 비용으로 어떤 retry/latency를 받아들일 수 있는가`가 되어야 한다.

실무에서는 lock을 DB 안에서만 보지 않는다. Optimistic locking은 row의 version column을 이용해 update 시점에 `내가 읽은 version이 아직 현재인가`를 확인한다. Distributed lock은 DB 밖 자원을 직렬화할 수 있지만 lease 만료, clock, network partition, fencing token 문제가 있다. Queue는 같은 key의 작업을 한 worker로 모아 순서를 만들 수 있지만 throughput과 latency를 바꾼다. 이 수단들은 DB lock의 대체품이라기보다, 보호해야 할 불변식과 실패 비용에 따라 고르는 별도 도구다.

면접에서 답변을 압축하면 다음처럼 말할 수 있다. 격리 수준은 어떤 이상 현상을 허용할지 정하는 계약이고, lock은 특정 자원을 보호하는 실행 수단이며, deadlock은 그 실행 수단들이 서로 순환 대기할 때 생긴다. 제품별로는 InnoDB의 index/gap/next-key lock과 PostgreSQL의 MVCC/SSI/retry 경계를 나누어야 한다. 애플리케이션은 lock order, 짧은 transaction, idempotent retry, 조건부 update, constraint를 함께 써서 업무 불변식을 닫아야 한다.

이 흐름을 검증하려면 실제로 네 가지를 관찰한다. 첫째, blocking session과 blocked session을 찾는다. 둘째, 어떤 SQL과 어떤 index range가 자원을 잡았는지 확인한다. 셋째, deadlock log에서 wait-for cycle을 읽는다. 넷째, retry 뒤에도 업무 side effect가 중복되지 않는지 본다. 관측 없이 `락이 걸렸다`고만 말하면 원인은 닫히지 않는다.

```text
blocked query -> blocker transaction -> locked resource/index range -> invariant under protection -> retry or ordering repair
```

이 마지막 trace가 중요하다. Lock wait를 보면 보통 빨리 kill하고 싶어진다. 하지만 kill은 현상 완화일 뿐이다. 어떤 불변식을 보호하느라 기다렸는지, 그 불변식이 정말 그 lock 범위를 필요로 하는지, code path가 같은 순서로 자원을 잡는지, retry가 안전한지까지 봐야 같은 문제가 반복되지 않는다.

### 방어 수단을 고르는 기준

동시성 문제를 봤을 때 바로 `락을 걸자`로 가면 설계가 거칠어진다. 먼저 충돌이 발생하는 자원이 무엇인지 구분한다. 하나의 row라면 조건부 update나 optimistic version column이 가장 단순할 수 있다. 연속된 범위라면 index range와 next-key/gap lock이 중요해질 수 있다. 여러 row의 aggregate라면 guard row, materialized counter, serializable retry, constraint 재설계가 필요할 수 있다. 외부 시스템까지 걸리면 DB lock만으로는 부족하다.

```text
single row invariant      -> conditional update, row lock, version column
range/predicate invariant -> range lock, exclusion/unique design, serializable retry
aggregate invariant       -> guard row, summary row lock, queue serialization
external side effect      -> idempotency, outbox, compensation, fencing token
```

이 표는 정답표가 아니라 질문 순서다. 예를 들어 쿠폰 1회 사용은 단일 `(coupon_id, user_id)` unique key로 막을 수 있으면 lock을 넓게 잡을 필요가 없다. 반면 `한 시간에 최대 100명 예약` 같은 aggregate 제한은 단순 unique key로 닫히지 않을 수 있다. 이때는 counter row를 갱신하며 row lock을 잡거나, reservation insert 후 집계 검사를 serializable retry로 감싸거나, 예약 요청을 show_id별 queue로 직렬화할 수 있다. 각 선택은 throughput, latency, 실패 시 retry 경험이 다르다.

면접에서 강한 답은 `저라면 lock을 겁니다`가 아니라 `불변식이 단일 row인지, range인지, aggregate인지 먼저 나누고, 그에 맞는 가장 좁은 방어를 고릅니다`라고 말하는 것이다. 이렇게 말하면 DBMS 지식과 애플리케이션 설계 판단이 연결된다. 또한 lock은 강한 도구이지만 관측 가능성이 있어야 한다. 어떤 lock이 누구를 막았는지 볼 수 없으면 장애 때 설명이 추측으로 흐른다. 그래서 설계 단계에서 lock timeout, deadlock retry, blocker query 관측, application trace id를 함께 준비해야 한다.

## 더 깊게 볼 자료

공식 자료:

- [PostgreSQL current: Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [PostgreSQL current: Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- [MySQL 8.4 Reference Manual: InnoDB Transaction Isolation Levels](https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html)
- [MySQL 8.4 Reference Manual: InnoDB Locking](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html)
- [MySQL 8.4 Reference Manual: Deadlocks in InnoDB](https://dev.mysql.com/doc/refman/8.4/en/innodb-deadlocks.html)

저장소 안에서 이어 볼 자료:

- `database/lock.md`
- `database/postgresql/lock.md`
- `database/deep-dive/transactions/13-isolation-anomalies.md`
- `database/deep-dive/transactions/14-locks-latches-deadlocks.md`
- `interviews/database-deep-dive/07-mvcc-snapshot-visibility.md`

이 자료를 볼 때는 `무엇을 읽었고, 무엇을 썼고, 어떤 자원이 기다림을 만들었는가`를 계속 추적한다. 격리 수준은 관측 계약이고, lock은 그 계약을 만들기 위한 수단이며, deadlock은 기다림 관계의 cycle이라는 세 문장으로 다시 압축할 수 있으면 면접 답변의 중심이 잡힌 것이다.
