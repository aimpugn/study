# 트랜잭션 전파, 격리 수준, 락, 데드락은 동시성 실패에서 각각 무엇을 담당하는가?

트랜잭션 전파(transaction propagation)는 애플리케이션 호출 흐름에서 이미 열린 transaction에 참여할지, 새 transaction을 열지, savepoint를 만들지 정하는 경계 규칙입니다. 격리 수준(isolation level)은 그렇게 만들어진 transaction들이 서로의 변경을 어떤 방식으로 관측할 수 있는지 정하는 계약이고, lock은 그 계약이나 업무 불변식을 실현하기 위해 DBMS가 row, index range, table, predicate, advisory key 같은 자원에 충돌 규칙을 붙이는 실행 수단입니다. 셋은 같은 말이 아닙니다. 전파는 호출 경계와 commit/rollback 지점을 정하고, 격리는 관측 결과를 정하며, lock은 그 관측 계약과 쓰기 충돌을 실행 단계에서 조정합니다. MVCC, predicate conflict tracking, next-key lock, serialization failure, retry도 모두 이 목표 주변에 있습니다.

면접에서 이 주제를 잘 답하려면 dirty read, non-repeatable read, phantom read 표를 외운 뒤 멈추면 안 됩니다. `PostgreSQL READ COMMITTED`, `PostgreSQL REPEATABLE READ`, `PostgreSQL SERIALIZABLE`, `MySQL/InnoDB REPEATABLE READ`가 같은 이름과 다른 구현을 가질 수 있음을 말해야 합니다. 또한 Spring의 `REQUIRED`, `REQUIRES_NEW`, `NESTED` 같은 전파 설정은 격리 수준이 아니라 transaction 경계를 호출 그래프에 어떻게 이어 붙일지 정하는 정책이라는 점을 분리해야 합니다. 그 위에서 row lock, gap lock, next-key lock, predicate lock, advisory lock, table lock을 보호 대상별로 나누고, latch와 lock의 차이, deadlock wait graph, retry boundary, 경쟁 조건(race condition)까지 이어야 합니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [격리 수준은 금지할 현상을 정하지만 제품별 의미가 다르다](#격리-수준은-금지할-현상을-정하지만-제품별-의미가-다르다)
    - [전파는 격리가 아니라 transaction 경계의 전파 규칙이다](#전파는-격리가-아니라-transaction-경계의-전파-규칙이다)
    - [Row lock은 논리 row가 아니라 접근 경로와 연결된다](#row-lock은-논리-row가-아니라-접근-경로와-연결된다)
    - [Gap lock과 next-key lock은 존재하지 않는 row를 둘러싼 충돌을 다룬다](#gap-lock과-next-key-lock은-존재하지-않는-row를-둘러싼-충돌을-다룬다)
    - [Predicate lock과 SSI는 읽은 조건 자체를 보호하려는 접근이다](#predicate-lock과-ssi는-읽은-조건-자체를-보호하려는-접근이다)
    - [Advisory lock은 DB가 의미를 모르는 이름 붙은 락이다](#advisory-lock은-db가-의미를-모르는-이름-붙은-락이다)
    - [Latch와 lock은 보호하는 층위가 다르다](#latch와-lock은-보호하는-층위가-다르다)
    - [Deadlock은 wait graph cycle이고 retry는 설계의 일부다](#deadlock은-wait-graph-cycle이고-retry는-설계의-일부다)
    - [Anomaly별로 방어 수단이 다르다](#anomaly별로-방어-수단이-다르다)
    - [경쟁 조건은 anomaly, lock wait, deadlock을 모두 품는 더 넓은 이름이다](#경쟁-조건은-anomaly-lock-wait-deadlock을-모두-품는-더-넓은-이름이다)
    - [Lock wait를 줄이는 일과 correctness를 지키는 일은 함께 봐야 한다](#lock-wait를-줄이는-일과-correctness를-지키는-일은-함께-봐야-한다)
    - [DB lock wait는 OS scheduler와 I/O wait 위에서 관측된다](#db-lock-wait는-os-scheduler와-io-wait-위에서-관측된다)
    - [Index 설계는 lock 설계이기도 하다](#index-설계는-lock-설계이기도-하다)
    - [Retry는 side effect와 함께 설계해야 한다](#retry는-side-effect와-함께-설계해야-한다)
    - [Deadlock 분석은 victim query 하나로 끝나지 않는다](#deadlock-분석은-victim-query-하나로-끝나지-않는다)
    - [Table lock과 metadata lock은 장애 모양이 다르다](#table-lock과-metadata-lock은-장애-모양이-다르다)
- [DBMS별 경계](#dbms별-경계)
    - [PostgreSQL](#postgresql)
    - [MySQL/InnoDB](#mysqlinnodb)
    - [애플리케이션 경계](#애플리케이션-경계)
- [직접 재생해 보기](#직접-재생해-보기)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
- [답변을 더 단단하게 만드는 판단 흐름](#답변을-더-단단하게-만드는-판단-흐름)
    - [방어 수단을 고르는 기준](#방어-수단을-고르는-기준)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

짧게 답하면 이렇게 말할 수 있습니다. 트랜잭션 전파(transaction propagation)는 서비스 메서드 호출이 같은 물리 transaction을 공유할지 새 transaction을 열지 정합니다. SQL 격리 수준(isolation level)은 만들어진 transaction들이 서로 어떤 중간 결과나 새 row를 볼 수 있는지 정하는 계약입니다. DBMS는 그 계약을 MVCC snapshot, row lock, range lock, predicate 성격의 conflict tracking, serialization failure 같은 방식으로 구현합니다. Lock은 그중 하나의 실행 수단이고, deadlock은 여러 transaction이 서로 가진 자원을 기다리며 cycle을 만들 때 발생합니다. 경쟁 조건은 결과가 실행 타이밍에 따라 달라지는 더 넓은 이름이고, dirty read나 phantom read는 그중 DB transaction에서 관측되는 대표 현상입니다. MVCC의 version visibility 자체는 [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md)에서 먼저 잡아 두면 이 문서의 lock 논리가 더 잘 보입니다.

가장 흔한 오해는 `격리 수준을 높이면 안전하다`와 `lock을 걸면 안전하다`를 같은 말로 보는 것입니다. 격리 수준을 높이면 anomaly를 줄일 수 있지만 abort/retry, lock wait, throughput 저하가 생깁니다. Lock을 더 넓게 잡으면 특정 충돌은 막을 수 있지만 deadlock과 contention이 늘 수 있습니다. 반대로 lock을 너무 좁게 잡거나 index가 없어 range lock이 넓어지면, 의도와 다른 row까지 막거나 phantom을 놓칠 수 있습니다.

PostgreSQL은 MVCC 기반 plain read가 writer를 막지 않는 성격이 강합니다. Serializable에서는 Serializable Snapshot Isolation 계열로 위험한 read/write dependency를 감지해 serialization failure를 낼 수 있습니다. 이때 PostgreSQL 문서의 predicate lock은 InnoDB gap lock처럼 writer를 직접 막는 물리적 range lock으로만 이해하면 안 되고, serialization anomaly 가능성을 추적하는 장치로 봐야 합니다. MySQL/InnoDB는 index record lock, gap lock, next-key lock을 통해 range conflict와 phantom을 다룹니다. InnoDB의 `row lock`은 실제로 index record와 range에 걸리는 경우가 많으므로, 어떤 index로 접근했는지가 lock 범위를 바꿉니다.

Deadlock은 DB가 이상한 상태가 되었다는 뜻이 아니라, 동시 transaction scheduling에서 생길 수 있는 정상적인 실패 모드입니다. T1이 A를 잡고 B를 기다리며, T2가 B를 잡고 A를 기다리면 cycle이 생깁니다. DBMS는 deadlock을 감지하면 보통 한 transaction을 victim으로 골라 abort합니다. 애플리케이션은 이 오류를 재시도 가능한 경계로 다뤄야 하고, 같은 자원을 항상 같은 순서로 잡는 설계로 빈도를 줄여야 합니다.

면접 답변의 핵심은 다섯 가지입니다. 첫째, 전파(propagation)는 호출 경계 정책이고 격리(isolation)는 관측 계약이며 lock은 구현 수단입니다. 둘째, DBMS별로 같은 isolation 이름의 실제 동작이 다릅니다. 셋째, lock은 보호 대상과 mode를 기준으로 말해야 합니다. 넷째, deadlock은 wait graph cycle이며 retry와 lock ordering이 설계의 일부입니다. 다섯째, 경쟁 조건(race condition)은 DB 용어 하나가 아니라 애플리케이션 상태 전이가 실행 순서에 의존한다는 더 넓은 문제 이름입니다.

## 먼저 잡아야 할 작은 모델

작은 모델은 좌석 예약입니다. 공연에 남은 좌석이 하나뿐인데 두 transaction이 동시에 예약하려고 합니다.

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

만약 실제로 같은 seat row를 update한다면 row lock이나 write conflict가 둘 중 하나를 막을 수 있습니다. 하지만 invariant가 `남은 좌석 count > 0이면 예약 row를 추가한다`처럼 aggregate/predicate에 걸려 있으면 단일 row lock만으로 충분하지 않을 수 있습니다. 특히 예약 row만 insert하고 seat 상태를 늦게 바꾸거나, 조건이 range query로 표현되면 phantom과 write skew가 등장합니다.

이 좌석 예약을 Spring 서비스 호출로 감싸면 전파(propagation)가 별도 축으로 보입니다.

```text
OrderService.reserveSeat()      @Transactional(REQUIRED)
  -> SeatRepository.markSold()  같은 물리 transaction에 참여
  -> AuditService.writeLog()    @Transactional(REQUIRES_NEW)
  -> CouponService.useCoupon()  @Transactional(NESTED)
```

```text
T0: reserveSeat 진입
    REQUIRED가 물리 transaction P1을 엽니다.

T1: markSold 호출
    REQUIRED는 P1에 참여합니다. commit은 아직 없습니다.

T2: writeLog 호출
    REQUIRES_NEW는 P1을 잠시 묶어 두고 물리 transaction P2를 새로 엽니다.
    P2가 먼저 commit되면, 나중에 P1이 rollback되어도 audit row는 남을 수 있습니다.

T3: useCoupon 호출
    NESTED는 같은 P1 안에 savepoint를 만듭니다.
    coupon 처리만 savepoint로 되돌려도 P1 전체를 계속 진행할 수 있습니다.
```

이 trace에서 전파는 `어떤 commit 경계가 생기는가`를 바꿉니다. 그러나 P1과 P2가 동시에 다른 transaction과 부딪힐 때 어떤 값을 볼지는 isolation이 정하고, 어떤 row나 index range를 기다릴지는 lock이 정합니다. 따라서 `REQUIRES_NEW를 쓰면 동시성 문제가 해결된다`가 아니라 `commit 경계가 분리되므로 성공/실패와 lock 보유 시간이 어떻게 달라지는가`를 물어야 합니다. 특히 바깥 transaction이 connection을 붙잡은 채 안쪽 `REQUIRES_NEW`가 새 connection을 요구하면 connection pool 부족과 대기 문제가 생길 수 있습니다.

다음 trace는 row lock과 predicate 보호의 차이를 보여 줍니다.

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

  이때 같은 predicate를 믿은 두 transaction이 서로 다른 row를 insert할 수 있습니다.
```

이 모델은 격리 수준 표보다 강합니다. Dirty read, non-repeatable read, phantom read, write skew가 모두 `무엇을 읽고 그 읽은 결과를 근거로 무엇을 쓰는가`라는 질문으로 돌아오기 때문입니다. Lock도 마찬가지입니다. 어떤 row를 잡았는지, 어떤 index range를 잡았는지, table 전체를 막았는지, advisory key만 잡았는지에 따라 보호하는 invariant가 달라집니다.

같은 좌석 예약 예시도 실패 모양을 나누면 해결책이 달라집니다.

| 실패 모양 | 읽은 것 | 쓴 것 | 주된 방어 후보 |
| --- | --- | --- | --- |
| 같은 좌석을 두 번 팝니다. | `seat_no='A1'` row | 같은 row의 `status` | 조건부 update, row lock, unique reservation key |
| 남은 좌석 수를 초과합니다. | `status='FREE'`인 row 집합 | 서로 다른 reservation row | guard row, range/predicate 보호, serializable retry |
| 같은 쿠폰을 두 번 씁니다. | coupon 사용 이력 | `(coupon_id, user_id)` 사용 row | unique constraint, idempotency key |
| 결제 API가 두 번 호출됩니다. | DB row만으로는 외부 호출 여부를 완전히 모릅니다. | 외부 PG 상태 | outbox, idempotency key, 보상/조회 API |

이 표의 목적은 "락을 걸면 됩니다"를 더 작은 판단으로 쪼개는 것입니다. 읽은 대상과 쓴 대상이 같은 row인지, 여러 row의 조건인지, DB 밖 상태인지에 따라 DB lock, constraint, retry, outbox 중 무엇이 중심이 될지 달라집니다.

## 깊은 메커니즘

### 격리 수준은 금지할 현상을 정하지만 제품별 의미가 다르다

SQL 표준은 isolation을 dirty read, non-repeatable read, phantom read 같은 현상으로 설명합니다. Dirty read는 commit되지 않은 다른 transaction의 변경을 읽는 것입니다. Non-repeatable read는 같은 row를 두 번 읽었는데 중간 commit 때문에 값이 달라지는 것입니다. Phantom read는 같은 조건으로 다시 검색했을 때 새 row가 나타나는 것입니다.

하지만 실제 DBMS는 표준 표를 그대로 구현하는 것이 아니라 자기 저장 구조와 동시성 제어 방식에 맞게 isolation을 제공합니다. PostgreSQL의 `READ COMMITTED`는 statement마다 snapshot을 새로 만들기 때문에 같은 transaction 안의 반복 SELECT가 다른 committed 값을 볼 수 있습니다. PostgreSQL의 `REPEATABLE READ`는 phantom까지 상당히 막는 snapshot isolation 성격을 가집니다. PostgreSQL의 `SERIALIZABLE`은 SSI로 dependency cycle을 감지해 serialization failure를 낼 수 있습니다.

MySQL/InnoDB는 기본 `REPEATABLE READ`에서 consistent read는 transaction read view를 유지하지만, locking read와 range scan에서는 next-key lock이 중요해집니다. `READ COMMITTED`에서는 gap lock 사용이 줄어드는 등 lock behavior가 달라질 수 있습니다. 따라서 `REPEATABLE READ면 phantom이 발생한다/안 한다`처럼 제품 없이 단정하는 답은 위험합니다.

### 전파는 격리가 아니라 transaction 경계의 전파 규칙이다

Spring 같은 프레임워크에서 말하는 트랜잭션 전파(transaction propagation)는 이미 시작된 transaction이 있을 때 안쪽 메서드가 그 transaction에 참여할지, 독립 transaction을 열지, transaction 없이 실행할지, savepoint를 만들지 정합니다. 기본값인 `REQUIRED`는 현재 transaction이 있으면 참여하고 없으면 새로 시작합니다. 이때 여러 메서드마다 논리적인 transaction scope가 생길 수 있지만, 보통 하나의 물리 transaction에 매핑됩니다.

이 차이는 rollback에서 바로 드러납니다. 안쪽 `REQUIRED` scope가 rollback-only 표시를 남기면 바깥 scope는 정상 commit을 시도하더라도 결국 commit되지 않습니다. 그래서 Spring은 호출자가 commit됐다고 오해하지 않도록 `UnexpectedRollbackException`을 낼 수 있습니다. 반대로 `REQUIRES_NEW`는 바깥 transaction과 독립된 물리 transaction을 새로 열기 때문에 안쪽 commit/rollback이 바깥 transaction의 commit 여부와 분리됩니다. `NESTED`는 대개 같은 물리 transaction 안의 savepoint로 이해하는 편이 안전합니다.

```text
REQUIRED
  outer P1 시작
  inner는 P1에 참여
  commit/rollback은 바깥 P1 경계에서 결정

REQUIRES_NEW
  outer P1은 유지된 채 일시 중단
  inner P2 새로 시작
  P2는 먼저 commit/rollback 가능

NESTED
  outer P1 안에 savepoint 생성
  inner 실패분만 savepoint로 되돌릴 수 있음
```

전파(propagation)는 transaction 경계와 connection 사용량, lock 보유 시간, rollback 전파 방식을 바꾸지만, dirty read나 phantom read를 직접 금지하는 설정은 아닙니다. 격리 수준(isolation level), timeout, read-only 같은 속성도 안쪽 메서드가 기존 transaction에 참여하는 경우에는 바깥 transaction의 성격을 따라갈 수 있습니다. 따라서 면접에서는 `propagation=REQUIRES_NEW라서 격리가 강해집니다`가 아니라 `새 물리 transaction이 생기므로 commit 경계, connection pool, lock release 시점, side effect 생존 여부가 달라집니다`라고 말해야 합니다.

### Row lock은 논리 row가 아니라 접근 경로와 연결된다

`row lock`이라는 말은 쉽게 들리지만, InnoDB에서는 실제로 index record lock을 중심으로 이해해야 합니다. `WHERE id = 10 FOR UPDATE`가 primary key index를 사용하면 해당 index record에 lock을 걸 수 있습니다. 그런데 조건에 맞는 index가 없어서 table scan을 하거나, range scan을 넓게 하면 lock 범위가 커질 수 있습니다. Lock은 SQL 문장만이 아니라 실행 계획과 index 구조에 영향을 받습니다.

```text
좋은 접근 경로
  WHERE user_id = 100 AND status = 'READY'
  index(user_id, status)를 사용
  필요한 record/range만 lock할 가능성이 높습니다.

나쁜 접근 경로
  WHERE lower(email) = 'a@example.com'
  usable index 없음
  많은 row를 읽으며 더 넓은 lock/wait를 만들 수 있습니다.
```

PostgreSQL에서도 row-level lock은 `FOR UPDATE`, `FOR NO KEY UPDATE`, `FOR SHARE`, `FOR KEY SHARE`처럼 mode가 나뉩니다. Table-level lock도 mode별 충돌 관계가 다릅니다. 면접에서 `SELECT는 lock을 안 겁니다`라고만 말하면 부족합니다. Plain SELECT, locking SELECT, DDL, foreign key check, update/delete가 만드는 lock을 나누어야 합니다.

### Gap lock과 next-key lock은 존재하지 않는 row를 둘러싼 충돌을 다룬다

Phantom은 기존 row 하나를 잠그는 것만으로 막히지 않습니다. `WHERE price BETWEEN 100 AND 200`에 해당하는 row를 읽은 뒤 그 범위에 새 row가 insert되면, 두 번째 읽기에서 결과가 달라질 수 있습니다. InnoDB next-key lock은 index record와 그 앞 gap을 함께 잠그는 방식으로 range에 새 row가 끼어드는 일을 막을 수 있습니다.

```text
index values: 10, 20, 30
query: WHERE value BETWEEN 11 AND 29 FOR UPDATE

record 20과 그 주변 gap을 잠그면
다른 transaction이 value=15 또는 value=25를 insert하려 할 때 기다릴 수 있습니다.
```

이 설명에서 index가 중요합니다. Range를 어떤 index로 읽었는지에 따라 gap과 next-key lock 범위가 달라집니다. Unique index로 정확히 하나의 record를 찾는 경우와 non-unique range scan은 lock 범위가 다릅니다. Index가 없으면 DBMS가 더 많은 record를 검사해야 하고, 그만큼 충돌 범위가 넓어질 수 있습니다.

같은 조건도 index가 있느냐에 따라 lock 후보가 달라질 수 있습니다.

```text
index(score): 10 | 20 | 30 | 40

query: score BETWEEN 11 AND 29 FOR UPDATE

candidate range:
  gap (10,20) + record 20 + gap (20,30)

insert score=15
  -> query가 보호한 gap에 들어가므로 기다릴 수 있습니다.

insert score=35
  -> 이 range 밖이므로 다른 조건이 없다면 영향이 작을 수 있습니다.
```

이 그림은 InnoDB를 설명할 때 특히 유용합니다. `row lock`이라는 단어만 들으면 이미 있는 row만 잠그는 것처럼 느껴지지만, phantom을 막으려는 range 읽기에서는 "아직 존재하지 않는 값이 들어올 자리"도 보호 대상이 될 수 있습니다.

PostgreSQL은 InnoDB식 gap/next-key lock 용어로 설명하지 않습니다. Serializable에서는 predicate 성격의 read/write dependency를 추적하고, 위험한 구조가 발견되면 serialization failure를 냅니다. 그래서 PostgreSQL과 InnoDB를 비교할 때는 `phantom을 막는다`는 목표와 `gap을 잠근다`, `dependency를 감지하고 abort한다`는 구현을 분리해야 합니다.

### Predicate lock과 SSI는 읽은 조건 자체를 보호하려는 접근이다

Predicate는 `status='FREE'인 좌석`, `balance < 0인 계좌`, `doctor on_call=true인 row`처럼 row 하나가 아니라 조건입니다. Serializable isolation이 진짜 serial execution처럼 보이려면 transaction이 읽은 predicate와 나중에 들어오는 write 사이의 충돌도 다뤄야 합니다.

PostgreSQL Serializable Snapshot Isolation은 단순히 모든 predicate를 물리적으로 잠그는 방식이 아니라, snapshot isolation 위에서 read/write dependency를 추적해 serialization anomaly 가능성이 있는 구조를 감지합니다. 감지되면 한 transaction을 abort시켜 실제 serial order를 만들 수 없는 결과를 막습니다. PostgreSQL 문서에서는 이 추적 정보를 predicate lock으로 설명하지만, 이 lock은 일반 row lock처럼 다른 transaction을 막아서 deadlock을 만드는 lock과는 성격이 다릅니다. 이때 애플리케이션은 serialization failure를 retry해야 합니다.

이 지점이 면접에서 좋습니다. `Serializable이면 lock을 많이 걸어서 느립니다`라고만 말하면 PostgreSQL의 SSI를 놓칩니다. `Serializable이면 retry가 필요 없습니다`라고 말하면 더 틀립니다. Strong isolation을 제공하려면 DB가 어떤 transaction을 실패시킬 수 있고, 애플리케이션은 그 실패를 정상 경로로 설계해야 합니다.

### Advisory lock은 DB가 의미를 모르는 이름 붙은 락이다

Advisory lock은 row나 table에 자동으로 묶인 lock이 아니라 애플리케이션이 정한 key에 대해 DB lock manager를 빌리는 방식입니다. PostgreSQL의 advisory lock이 대표적입니다. 예를 들어 `tenant_id=10`에 대한 batch job이 동시에 하나만 돌게 하고 싶다면 tenant id를 advisory lock key로 사용할 수 있습니다.

이 방식은 유용하지만 DB가 그 key의 업무 의미를 이해하지는 않습니다. 어떤 code path는 advisory lock을 잡고 어떤 path는 잡지 않으면 보호가 깨집니다. Transaction-level advisory lock과 session-level advisory lock도 생명주기가 다릅니다. Connection pool을 쓰는 환경에서 session-level lock을 잘못 다루면 lock이 예상보다 오래 남거나 엉뚱한 요청에 영향을 줄 수 있습니다.

면접 답변에서는 advisory lock을 `분산락 대체재`처럼 과장하면 안 됩니다. 같은 DB에 연결된 참여자들이 같은 규칙을 지킬 때는 강력한 도구가 될 수 있지만, DB 밖 시스템이나 다른 storage를 자동으로 보호하지 않습니다. Lock key 설계, release boundary, timeout, deadlock 가능성을 함께 말해야 합니다.

### Latch와 lock은 보호하는 층위가 다르다

Latch는 DBMS 내부 자료구조를 아주 짧게 보호하는 동기화 장치로 이해하면 됩니다. Buffer page, B-tree page, lock table 같은 내부 구조가 동시에 깨지지 않게 하는 데 쓰입니다. Transaction lock은 사용자 transaction 사이의 논리 충돌을 조정합니다. 둘을 같은 말로 쓰면 운영 지표를 잘못 해석합니다.

예를 들어 lock wait가 길다는 것은 어떤 transaction이 row/table/range lock을 들고 있어서 다른 transaction이 기다린다는 뜻일 수 있습니다. Latch contention은 CPU와 메모리 내부 구조 경쟁으로 나타날 수 있습니다. 해결도 다릅니다. Lock wait는 긴 transaction, 잘못된 update 순서, index 부재, batch 크기, isolation level을 봅니다. Latch contention은 hot page, buffer manager, index split, internal contention을 봅니다.

면접에서 latch를 깊게 몰라도 됩니다. 다만 `lock은 사용자 transaction의 논리 자원 충돌, latch는 엔진 내부 자료구조 보호`라고 구분하면 됩니다. 이 구분만 있어도 wait event를 해석할 때 한 단계 더 정확해집니다.

### Deadlock은 wait graph cycle이고 retry는 설계의 일부다

Deadlock은 간단한 trace로 설명할 수 있습니다.

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

DBMS는 이런 cycle을 감지하면 보통 한 transaction을 victim으로 골라 rollback합니다. Victim 선택 기준은 제품별로 다릅니다. 애플리케이션은 deadlock 오류를 사용자에게 곧바로 실패로 보여 주기보다, transaction 전체가 idempotent하거나 안전하게 다시 실행될 수 있으면 retry해야 합니다. 단, retry는 무한 반복이 아니라 backoff, 최대 횟수, observability, side effect 분리와 함께 설계해야 합니다.

Deadlock 예방의 기본은 lock order를 통일하는 것입니다. 여러 row를 갱신해야 한다면 모든 code path가 key를 정렬한 순서로 update합니다. Batch job도 random order로 update하지 않습니다. Foreign key가 참조하는 parent/child table 순서, inventory와 order table 순서, 계좌 A/B 순서를 통일합니다. 그래도 DB 내부 lock과 index range 때문에 모든 deadlock을 제거할 수는 없으므로 retry는 남깁니다.

### Anomaly별로 방어 수단이 다르다

Dirty read는 commit되지 않은 값을 읽는 문제입니다. 대부분의 현대 OLTP 설정에서는 기본 격리 수준에서 dirty read를 허용하지 않습니다. Non-repeatable read는 같은 row를 다시 읽었을 때 값이 달라지는 문제입니다. Statement snapshot을 새로 만드는 `READ COMMITTED`에서는 자연스럽게 보일 수 있고, transaction snapshot을 유지하면 줄어듭니다. Phantom은 같은 predicate query를 다시 실행했을 때 새 row가 나타나는 문제입니다. Write skew는 각 transaction이 서로 다른 row를 쓰지만 함께 보면 invariant가 깨지는 문제입니다.

이 현상들은 같은 lock 하나로 모두 해결되지 않습니다. 단일 row lost update는 row lock이나 optimistic version column으로 막기 쉽습니다. Range phantom은 index range lock, predicate conflict tracking, serializable retry가 필요할 수 있습니다. Write skew는 guard row를 도입해 같은 row를 lock하게 만들거나, constraint로 invariant를 DB가 알 수 있게 바꾸거나, serializable isolation과 retry를 쓸 수 있습니다. 따라서 면접에서는 현상 이름 뒤에 `어떤 읽기 결과를 믿고 어떤 쓰기를 했는가`를 붙여야 합니다.

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

### 경쟁 조건은 anomaly, lock wait, deadlock을 모두 품는 더 넓은 이름이다

경쟁 조건(race condition)은 둘 이상의 실행 흐름이 같은 상태를 읽거나 바꾸는데, 최종 결과가 실행 순서와 타이밍에 의존하는 상황입니다. DB isolation anomaly는 그 경쟁 조건이 transaction의 읽기/쓰기 관측에서 드러난 구체적인 모양입니다. Deadlock은 결과가 틀리는 anomaly라기보다 서로의 lock을 기다리다 진행이 멈추는 scheduling 실패입니다. Lock wait는 아직 cycle은 아니지만 어떤 transaction이 다른 transaction의 자원 해제를 기다리는 관측 신호입니다.

```text
race condition
  넓은 문제 이름: 실행 순서가 결과를 바꿉니다.

dirty / non-repeatable / phantom read
  읽기 관측이 isolation 계약에 따라 달라집니다.

lost update / write skew
  읽은 값을 근거로 쓴 결과가 업무 불변식을 깨뜨립니다.

lock wait / deadlock
  DBMS가 충돌을 조정하는 과정에서 기다림 또는 cycle이 생깁니다.
```

이 구분이 있어야 해결책도 정확해집니다. 경쟁 조건(race condition)이라는 말만 듣고 무조건 `synchronized`나 distributed lock을 붙이면 DB가 이미 제공하는 constraint, conditional update, row lock, unique index, serializable retry를 놓칠 수 있습니다. 반대로 isolation level만 올리면 외부 API 중복 호출, 잘못 쪼개진 transaction boundary, cache 갱신 순서 같은 애플리케이션 경쟁은 남을 수 있습니다. 먼저 어떤 공유 상태가 어떤 순서에서 깨지는지 trace를 그린 뒤, 그 상태를 DB가 아는 invariant로 내릴지, 애플리케이션에서 직렬화할지, retry와 멱등성으로 흡수할지 고릅니다.

### Lock wait를 줄이는 일과 correctness를 지키는 일은 함께 봐야 한다

운영에서 lock wait가 늘면 누구나 빨리 줄이고 싶어 합니다. 하지만 lock을 줄이는 방향이 invariant를 깨면 안 됩니다. 예를 들어 `SELECT ... FOR UPDATE`를 제거하면 대기 시간은 줄 수 있지만, 같은 재고를 두 transaction이 동시에 차감할 수 있습니다. 반대로 모든 작업에 table lock을 걸면 correctness는 단순해질 수 있지만 throughput이 무너집니다.

좋은 설계는 보호해야 할 invariant를 먼저 고른 뒤 가장 좁은 자원으로 보호합니다. 재고 한 SKU의 수량이면 그 SKU row를 조건부 update합니다. 한 사용자 계정의 상태 전이면 user id 기준으로 같은 row 또는 advisory key를 잡습니다. 특정 기간의 중복 예약이면 `(resource_id, time_range)`를 constraint나 range lock으로 표현할 수 있는지 봅니다. 보호 대상이 명확해질수록 lock 범위도 줄어듭니다.

### DB lock wait는 OS scheduler와 I/O wait 위에서 관측된다

DB에서 `lock wait`가 보인다고 해서 모든 시간이 DB lock manager 안에서만 흐르는 것은 아닙니다. Transaction을 들고 있는 backend process도 결국 운영체제의 process입니다. CPU를 쓰려면 scheduler가 runnable task 중 하나로 선택해야 하고, WAL flush나 data page read/write가 필요하면 kernel I/O 경로에서 sleep할 수 있습니다. Linux scheduler 문서는 runnable task의 실행 기회를 가상 실행 시간 같은 기준으로 나누고, block layer 문서는 I/O 요청이 software queue와 hardware dispatch queue를 거쳐 device driver로 내려간다고 설명합니다.

```text
T2 waits for row lock held by T1

T1 can be slow because:
  case A: T1 is doing CPU work but scheduler run queue is crowded
  case B: T1 is blocked in fsync or data page read/write
  case C: T1 is waiting for another DB lock
  case D: T1 is waiting for external API inside transaction

T2 symptom:
  DB shows lock wait
  root owner path may be CPU scheduling, I/O wait, nested lock, or external wait
```

이 구분은 운영 대응을 바꿉니다. T1이 다른 lock을 기다리는 중이면 wait graph를 따라가야 하고, T1이 `fsync`나 storage I/O를 기다리는 중이면 WAL sync time, checkpoint, dirty page writeback, disk queue를 봐야 합니다. T1이 외부 API를 transaction 안에서 기다리고 있으면 DB 튜닝보다 transaction boundary를 줄이는 쪽이 먼저입니다. 따라서 `blocked query -> blocker transaction -> blocker가 지금 무엇을 기다리는가 -> lock을 줄일지, I/O를 줄일지, transaction 경계를 줄일지` 순서로 보셔야 합니다.

운영에서 이 구분은 다음 표처럼 분해할 수 있습니다.

| 질문 | 확인할 것 | 판단이 바뀌는 이유 |
| --- | --- | --- |
| 누가 막혔나요? | blocked query, wait event, lock type | 사용자 증상과 직접 연결됩니다. |
| 누가 잡고 있나요? | blocker transaction, query, transaction age | kill 또는 수정 대상이 blocked query가 아닐 수 있습니다. |
| blocker는 무엇을 기다리나요? | CPU runnable, DB lock, I/O wait, 외부 호출 | root cause가 DB lock 밖에 있을 수 있습니다. |
| 어떤 불변식을 보호하나요? | row, range, table, advisory key, 외부 resource | lock을 줄여도 되는지 판단합니다. |
| 재시도해도 안전한가요? | idempotency key, outbox, side effect 분리 | deadlock victim을 다시 실행할 수 있는지 결정합니다. |

이 표를 채우면 "락이 걸렸으니 쿼리를 죽인다"보다 더 나은 결정을 할 수 있습니다. 특히 blocker가 `fsync`나 외부 API를 기다리는 중이면 SQL만 튜닝해서는 같은 장애가 반복됩니다.

### Index 설계는 lock 설계이기도 하다

InnoDB에서 range 조건을 locking read로 실행할 때 usable index가 없으면 많은 record와 gap을 검사하며 넓은 lock을 만들 수 있습니다. PostgreSQL에서도 index가 없으면 더 많은 row를 방문하고 더 오래 transaction이 유지될 수 있습니다. Lock 문제를 해결하려고 timeout만 늘리면 근본 원인을 놓칩니다. Query가 어떤 index를 사용해 어떤 후보 row를 읽는지 확인해야 합니다.

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

Index가 correctness를 직접 보장할 때도 있습니다. Unique index는 중복 insert를 막습니다. Conditional unique index나 exclusion constraint가 있는 DBMS에서는 `활성 예약은 같은 시간 구간에 하나만` 같은 invariant를 DB에 내릴 수 있습니다. 이 경우 애플리케이션 lock보다 DB constraint가 더 단순하고 강할 수 있습니다.

### Retry는 side effect와 함께 설계해야 한다

Deadlock이나 serialization failure를 retry하면 된다고 말하는 것은 절반입니다. Retry할 transaction 안에 외부 API 호출이나 메시지 발행이 들어 있으면 재시도 때 side effect가 중복될 수 있습니다. 따라서 retry 가능한 transaction function은 DB 내부 변경과 idempotency check를 중심으로 만들고, 외부 side effect는 outbox나 별도 멱등성 키로 분리하는 편이 안전합니다.

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

이 구분을 하면 isolation-lock 주제가 [트랜잭션 경계와 ACID](06-transaction-acid-boundary.md)로 연결됩니다. DB가 한 transaction을 abort시키는 것은 correctness를 지키기 위한 정상 동작일 수 있습니다. 애플리케이션이 그 abort를 재실행할 수 없는 형태로 side effect를 섞어 놓으면, DB의 correctness mechanism이 사용자 장애로 바뀝니다. 재시도 경계 밖의 메시지와 결제 호출은 [애플리케이션 경계, 멱등성, 금액 처리, outbox](12-application-boundaries-idempotency-money-outbox.md)의 문제로 분리해야 합니다.

Retry를 설계할 때는 transaction 함수가 "처음부터 다시 실행해도 같은 의미인가"를 확인해야 합니다.

| 재시도 대상 | 안전한 조건 | 위험 신호 |
| --- | --- | --- |
| DB row update | 조건부 update와 affected row count로 결과를 판정합니다. | update 전에 읽은 값을 application memory에 오래 들고 있습니다. |
| unique key insert | 같은 request id가 있으면 기존 결과를 조회합니다. | retry마다 새 id를 만들어 중복 row를 만듭니다. |
| outbox insert | event id가 안정적이고 consumer가 멱등적입니다. | retry 때 같은 업무 event가 다른 id로 여러 번 발행됩니다. |
| 외부 PG/API 호출 | provider idempotency key 또는 상태 조회가 있습니다. | DB deadlock retry가 외부 승인 API를 다시 호출합니다. |

Deadlock과 serialization failure는 DBMS가 정합성을 지키려고 transaction 하나를 포기시키는 경우가 많습니다. 애플리케이션이 재시도할 수 있게 만들려면 DB 내부 변경은 같은 request id로 다시 적용 가능해야 하고, 외부 행동은 transaction 밖의 멱등성 계약으로 분리되어야 합니다.

### Deadlock 분석은 victim query 하나로 끝나지 않는다

Deadlock log에는 보통 victim이 된 transaction의 query가 눈에 잘 띕니다. 하지만 root cause는 상대 transaction의 lock order, index range, foreign key check, trigger, batch ordering에 있을 수 있습니다. 따라서 deadlock 분석은 양쪽 transaction이 잡은 lock과 기다린 lock을 모두 봅니다.

분석 순서는 단순합니다. 먼저 wait graph의 node와 edge를 적습니다. 각 node가 어떤 transaction인지, 어떤 statement에서 어떤 lock을 보유했는지, 어떤 lock을 기다렸는지 씁니다. 그 다음 왜 그 자원을 그 순서로 잡았는지 code path를 찾습니다. 마지막으로 순서를 통일할 수 있는지, index로 범위를 줄일 수 있는지, transaction을 짧게 나눌 수 있는지, retry가 안전한지 결정합니다.

```text
T1 owns order:10, waits inventory:skuA
T2 owns inventory:skuA, waits order:10

possible fix
  every path locks order first, then inventory
  or every path locks inventory first, then order
  choose one and enforce it in service layer
```

### Table lock과 metadata lock은 장애 모양이 다르다

Row lock만 생각하면 DDL과 migration 장애를 놓칩니다. PostgreSQL의 `ACCESS EXCLUSIVE` lock은 많은 작업을 막을 수 있고, MySQL의 metadata lock은 오래 열린 transaction 때문에 DDL이 기다리거나 뒤따르는 query까지 막히는 상황을 만들 수 있습니다. 운영에서 `ALTER TABLE` 하나가 서비스 전체를 멈춘 것처럼 보이는 이유가 여기에 있습니다.

Migration을 설계할 때는 lock level, lock 획득 시점, timeout, online DDL 지원 범위, backfill batch 크기, rollback plan을 확인해야 합니다. 면접에서 isolation과 lock을 application transaction만으로 설명하면 schema change의 실제 위험을 놓칩니다. DBMS별 online DDL 능력이 좋아졌더라도, 모든 변경이 무중단인 것은 아닙니다.

## DBMS별 경계

### PostgreSQL

PostgreSQL은 table-level lock mode와 row-level lock mode를 공식 문서에서 자세히 나눕니다. `ACCESS SHARE`, `ROW EXCLUSIVE`, `SHARE`, `ACCESS EXCLUSIVE` 같은 table lock은 DDL과 DML 충돌을 이해하는 데 중요합니다. Row-level lock도 `FOR UPDATE`, `FOR NO KEY UPDATE`, `FOR SHARE`, `FOR KEY SHARE`로 나뉩니다. Plain SELECT는 보통 ACCESS SHARE table lock을 잡지만 row write를 막는 lock은 아닙니다.

Serializable은 PostgreSQL에서 특히 중요합니다. MVCC snapshot 위에서 Serializable Snapshot Isolation이 동작하고, dangerous structure가 생기면 serialization failure를 낼 수 있습니다. 따라서 PostgreSQL Serializable 답변에는 `retry`가 반드시 들어가야 합니다. `lock을 다 걸어서 순서대로 실행합니다`라고만 답하면 PostgreSQL의 실제 전략을 놓칩니다.

Deadlock 관측은 `pg_locks`, `pg_stat_activity`, server log의 deadlock detail, wait event를 봅니다. Blocker와 waiter를 찾고, 어떤 query가 어떤 lock을 들고 있는지, transaction age가 얼마나 되는지, 같은 자원을 어떤 순서로 잡는지 확인합니다.

### MySQL/InnoDB

InnoDB는 row-level locking을 index record 중심으로 봐야 합니다. Shared lock, exclusive lock, intention lock, record lock, gap lock, next-key lock, insert intention lock, auto-inc lock 같은 개념이 있습니다. 특히 next-key lock은 phantom 방어와 range conflict 설명의 중심입니다.

InnoDB deadlock 분석에서는 `SHOW ENGINE INNODB STATUS`의 latest detected deadlock이 자주 쓰입니다. 어떤 transaction이 어떤 lock을 기다렸고 어떤 lock을 보유했는지 나옵니다. Performance Schema의 data locks 관련 view도 함께 볼 수 있습니다. MySQL 8.4 기준 문서를 보면 isolation level별 locking read와 consistent read 차이를 공식적으로 확인할 수 있습니다.

InnoDB에서는 index 설계가 lock 설계입니다. 같은 SQL이라도 적절한 index가 있으면 좁은 range만 잠글 수 있고, index가 없거나 조건이 넓으면 불필요한 row와 gap까지 영향을 줄 수 있습니다. 그래서 lock 문제를 볼 때 query text만 보지 말고 execution plan과 index를 함께 봐야 합니다.

### 애플리케이션 경계

애플리케이션은 DB isolation을 보완할 수 있지만 대체하지는 않습니다. Optimistic locking의 version column은 lost update를 잡는 데 좋습니다. Unique constraint는 중복 insert를 막는 데 좋습니다. Queue serialization은 특정 aggregate의 write를 순서화하는 데 좋습니다. Distributed lock은 DB 밖 자원까지 포함한 작업 조율에 도움을 줄 수 있습니다. 그러나 각각은 실패 경계가 다릅니다.

예를 들어 optimistic lock은 update 시 `WHERE id=? AND version=?` 조건으로 affected row count를 확인합니다. 충돌이 나면 사용자가 다시 읽고 재시도해야 합니다. 이것은 DB row lock을 오래 들고 있지 않는 장점이 있지만, conflict가 많은 hot row에서는 retry가 많아질 수 있습니다. Distributed lock은 lock service와 네트워크 failure, lease 만료, clock skew, fencing token을 봐야 합니다. DB lock보다 무조건 낫다고 말하면 안 됩니다.

Spring의 트랜잭션 전파(transaction propagation)는 이 애플리케이션 경계에 속합니다. `REQUIRED`는 서비스 경계의 여러 repository 호출을 같은 commit 경계에 묶는 데 적합하고, `REQUIRES_NEW`는 audit log나 실패 기록처럼 바깥 transaction과 별도로 남겨야 하는 작업에 쓰일 수 있습니다. 하지만 `REQUIRES_NEW`는 새 connection을 요구하고, 바깥 transaction이 들고 있던 lock은 그대로 남을 수 있으며, 안쪽 commit이 바깥 rollback보다 먼저 영구화될 수 있습니다. 따라서 전파 선택은 `업무적으로 commit 생존 범위가 달라져도 되는가`, `connection pool이 버틸 수 있는가`, `재시도 때 side effect가 중복되지 않는가`를 함께 봐야 합니다.

## 직접 재생해 보기

1. 전파 경계를 눈으로 그립니다.

    작은 Spring 서비스 세 개를 가정합니다. A는 `REQUIRED`, B는 `REQUIRED`, C는 `REQUIRES_NEW` 또는 `NESTED`로 둡니다. A가 시작한 transaction 안에서 B와 C를 호출할 때 commit 지점, rollback 전파, connection 수, lock release 시점을 표로 적습니다. PASS 신호는 `REQUIRED는 같은 물리 transaction을 공유하고, REQUIRES_NEW는 독립 물리 transaction을 열며, NESTED는 savepoint 성격으로 설명할 수 있는 것`입니다. FAIL 신호는 전파를 dirty read나 phantom read를 막는 격리 수준처럼 설명하는 것입니다.

2. Deadlock을 일부러 만듭니다.

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

    PASS 신호는 DB가 deadlock을 감지하고 한 transaction을 abort하는 것입니다. 그 다음 두 session 모두 A, B 순서로 update하도록 바꾸면 deadlock 가능성이 줄어드는지 확인합니다.

3. InnoDB range lock을 확인합니다.

    적절한 index가 있는 table에서 `SELECT ... WHERE score BETWEEN 10 AND 20 FOR UPDATE`를 실행하고 다른 session에서 score=15 row를 insert해 봅니다. 기다림이 발생하면 어떤 index range가 보호되는지 설명합니다. Index를 제거하거나 다른 조건으로 바꾸었을 때 lock 범위가 달라지는지도 `EXPLAIN`과 lock view로 봅니다.

4. PostgreSQL Serializable retry를 만듭니다.

    Write skew 예시를 사용합니다. 두 transaction이 같은 predicate를 읽고 서로 다른 row를 update해 invariant를 깨는 패턴을 만듭니다. `SERIALIZABLE`에서 한 transaction이 serialization failure로 실패하면 PASS입니다. 애플리케이션 관점에서는 같은 transaction function을 재실행해야 합니다.

5. Advisory lock 생명주기를 비교합니다.

    PostgreSQL에서 transaction-level advisory lock과 session-level advisory lock을 각각 잡고 commit/rollback/connection close 때 어떻게 해제되는지 확인합니다. Connection pool에서 session-level lock을 쓰면 왜 위험한지 설명할 수 있어야 합니다.

6. Latch와 lock 지표를 분리해 말합니다.

    실제 환경이 없다면 wait event 예시를 표로 만듭니다. `row lock wait`, `metadata lock`, `buffer content lock`, `IO wait`가 모두 같은 해결책을 갖지 않는다는 점을 적습니다. PASS는 대기 이름을 보고 바로 쿼리를 죽이는 것이 아니라 보호 대상과 owner를 먼저 찾는 것입니다.

## 면접 꼬리 질문

- 전파와 격리는 어떻게 다른가요?

    전파(propagation)는 호출된 메서드가 기존 transaction에 참여할지 새 transaction을 열지 정하는 애플리케이션 경계 정책입니다. 격리(isolation)는 이미 만들어진 transaction들이 서로 어떤 변경을 볼 수 있는지 정하는 DB 관측 계약입니다. `REQUIRES_NEW`는 새 물리 transaction을 만들 수 있지만 그 자체가 dirty read나 phantom read를 막는 격리 수준은 아닙니다.

- REPEATABLE READ면 phantom read가 항상 발생하나요?

    제품 없이 단정하면 안 됩니다. 표준 설명과 실제 DBMS 구현이 다릅니다. PostgreSQL Repeatable Read는 snapshot isolation 성격으로 phantom을 다르게 다루고, InnoDB Repeatable Read는 next-key lock과 consistent read가 함께 등장합니다. 공식 문서와 실험으로 제품별 경계를 확인해야 합니다.

- Row lock은 정말 row 자체에 걸리나요?

    DBMS마다 설명이 다르지만 InnoDB에서는 index record lock으로 이해해야 합니다. 어떤 index를 타는지와 range 조건이 lock 범위를 바꿉니다. PostgreSQL에서도 row lock mode와 table lock mode가 따로 있습니다.

- Deadlock이 발생하면 무조건 DB 문제인가요?

    아닙니다. 여러 transaction이 자원을 다른 순서로 잡으면 애플리케이션 설계 때문에 생길 수 있습니다. DB는 deadlock을 감지하고 하나를 abort할 뿐입니다. 일관된 lock order, 짧은 transaction, 적절한 index, retry 설계가 필요합니다.

- Serializable을 쓰면 retry가 필요 없나요?

    반대입니다. Serializable이 강한 격리를 제공하려면 어떤 transaction을 실패시킬 수 있습니다. PostgreSQL Serializable에서는 serialization failure retry가 정상 경로입니다.

- Latch와 lock의 차이는 무엇인가요?

    Latch는 DBMS 내부 자료구조를 짧게 보호하는 동기화 장치이고, transaction lock은 사용자 transaction 사이의 논리 자원 충돌을 조정합니다. 관측 지표와 해결책이 다릅니다.

- Advisory lock을 쓰면 분산락 문제가 해결되나요?

    같은 DB와 같은 규칙을 공유하는 참여자 사이에서는 유용합니다. 하지만 DB 밖 시스템, session 생명주기, connection pool, timeout, fencing 문제는 별도로 봐야 합니다. Advisory lock은 DB가 의미를 모르는 key에 대한 약속입니다.

- 경쟁 조건과 phantom read는 같은 말인가요?

    아닙니다. 경쟁 조건(race condition)은 실행 순서와 타이밍에 따라 결과가 달라지는 넓은 문제 이름입니다. Phantom read는 같은 predicate query를 다시 실행했을 때 새 row가 보이는 DB isolation 현상입니다. Phantom read는 race condition의 한 형태로 볼 수 있지만, race condition은 cache 갱신 순서, 외부 API 중복 호출, lock ordering 문제처럼 DB isolation 표 밖에서도 생깁니다.

## 함정 질문

- Isolation level 표만 외우고 DBMS 차이를 무시하는 답

    표는 출발점일 뿐입니다. 같은 이름도 PostgreSQL과 InnoDB에서 snapshot, lock, phantom, retry behavior가 다를 수 있습니다. 제품별 공식 문서와 재생 실험으로 닫아야 합니다.

- Lock을 많이 걸면 안전하고 비용은 없다고 말하는 답

    넓은 lock은 anomaly를 줄일 수 있지만 contention과 deadlock, throughput 저하를 만듭니다. 필요한 invariant를 가장 좁고 명확한 경계로 보호하는 것이 중요합니다.

- Deadlock을 timeout과 같은 말로 보는 답

    Timeout은 오래 기다려서 포기하는 것이고, deadlock은 wait graph cycle입니다. Deadlock은 한쪽을 abort해야 진행됩니다. 로그와 wait graph를 보고 구분해야 합니다.

- Predicate 문제를 row lock 하나로 해결하려는 답

    조건에 새 row가 들어오는 phantom이나 write skew는 기존 row 하나를 잡는 것만으로 막히지 않을 수 있습니다. Range/predicate 보호, serializable retry, guard row, constraint 설계를 봐야 합니다.

- 전파 설정을 격리 수준처럼 말하는 답

    `REQUIRED`, `REQUIRES_NEW`, `NESTED`는 호출 경계가 물리 transaction을 어떻게 공유하거나 나누는지 정합니다. Dirty read, non-repeatable read, phantom read를 어떤 수준까지 허용할지는 DB isolation level과 DBMS 구현이 정합니다. 전파로 commit 경계를 분리하면 side effect 생존 범위와 connection pool 압박이 바뀔 수 있으므로 오히려 별도 검토가 필요합니다.

- Application distributed lock을 DB isolation의 상위 호환으로 말하는 답

    Distributed lock은 다른 실패 모드를 가집니다. Lease 만료, lock service 장애, fencing token, clock 문제를 설계하지 않으면 DB lock보다 더 위험할 수 있습니다.

## 답변을 더 단단하게 만드는 판단 흐름

이 주제는 먼저 축을 나누면 훨씬 덜 헷갈립니다. 전파(propagation)는 `이 코드 호출이 어느 transaction 경계에 들어가는가`를 묻고, 격리(isolation)는 `동시에 실행되는 transaction끼리 무엇을 볼 수 있는가`를 묻고, lock은 `어떤 자원을 누가 기다리는가`를 묻습니다. 경쟁 조건(race condition)은 `실행 순서가 결과를 바꾸는가`라는 더 넓은 질문입니다. 이 네 질문을 섞으면 `REQUIRES_NEW를 쓰면 phantom이 막힌다`, `READ COMMITTED니까 deadlock이 없다`, `lock을 걸었으니 모든 경쟁 조건이 사라진다` 같은 답이 나옵니다.

격리 수준과 lock 질문은 항상 `어떤 불변식이 동시에 깨질 수 있는가`에서 시작합니다. 단일 row만 바뀌는 문제라면 row lock이나 조건부 update로 충분할 수 있습니다. 그러나 좌석 예약, 재고 총량, 당직 의사 최소 1명, 쿠폰 1회 사용처럼 여러 row나 predicate가 만드는 불변식은 단순 row lock만으로 닫히지 않을 수 있습니다. 면접에서는 먼저 불변식의 모양을 말한 뒤, 그 불변식을 어떤 자원 단위로 보호할지 설명해야 합니다.

```text
불변식 모양 -> 보호할 자원 단위 -> 필요한 관측 계약 -> lock/conflict/retry 전략
```

예를 들어 `남은 좌석이 있으면 예약한다`는 조건은 특정 row 하나가 아니라 `show_id=1 AND status='free'`라는 predicate에 기대고 있습니다. 두 transaction이 같은 predicate를 읽고 서로 다른 reservation row를 insert하면 단일 row conflict가 없는데도 oversell이 생길 수 있습니다. 이때 가능한 방어는 여러 가지입니다. 좌석 row 자체를 먼저 잡고 상태를 바꾸거나, show별 guard row를 lock하거나, unique/exclusion constraint로 최종 중복을 막거나, Serializable에서 실패한 transaction을 retry할 수 있습니다. 중요한 것은 수단 이름이 아니라 어떤 불변식을 어느 계층에서 닫을지입니다.

InnoDB를 말할 때는 index와 lock 범위를 함께 말해야 합니다. InnoDB row lock은 실제로 index record를 통해 잡히는 경우가 많습니다. 조건에 맞는 index가 있으면 좁은 range를 잠글 수 있지만, index가 없거나 조건이 넓으면 더 많은 record와 gap이 잠길 수 있습니다. Gap lock과 next-key lock은 phantom insert를 막는 데 쓰일 수 있지만, 예상보다 많은 insert가 기다리는 원인이 되기도 합니다. 그래서 lock wait 문제를 볼 때는 SQL 문장만 보지 말고 어떤 index range를 훑었는지 EXPLAIN과 lock 관측을 함께 봐야 합니다.

PostgreSQL을 말할 때는 `읽기가 writer를 막지 않는다`와 `Serializable에서 abort가 날 수 있다`를 같이 말해야 합니다. PostgreSQL의 MVCC는 plain read를 강하게 살려 주지만, Serializable Snapshot Isolation 계열에서는 위험한 read/write dependency가 감지되면 serialization failure가 발생할 수 있습니다. 이것은 DB 오류가 아니라 정합성을 지키기 위한 정상적인 거절일 수 있습니다. 따라서 애플리케이션은 그 transaction을 idempotent하게 재시도할 수 있어야 합니다.

Deadlock은 이 논의를 현실로 끌고 내려옵니다. Deadlock은 오래 기다린다는 뜻이 아니라 wait-for graph에 cycle이 생겼다는 뜻입니다. T1이 A를 잡고 B를 기다리고, T2가 B를 잡고 A를 기다리면 둘 중 하나는 abort되어야 합니다. DBMS는 보통 deadlock detector로 cycle을 찾고 victim을 고릅니다. 애플리케이션은 victim이 된 transaction을 재시도할 수 있어야 하고, 가능하면 모든 코드 경로가 같은 순서로 lock을 잡도록 만들어야 합니다.

```text
나쁜 순서
  request 1: lock account A -> lock account B
  request 2: lock account B -> lock account A

나은 순서
  모든 transfer: smaller account id 먼저 lock -> larger account id lock
```

이 예시는 단순하지만 강력합니다. Deadlock을 완전히 없애지는 못해도, 같은 자원 집합을 같은 순서로 잡으면 cycle 가능성이 크게 줄어듭니다. Batch update도 마찬가지입니다. `WHERE id IN (...)` 목록을 random하게 처리하면 transaction마다 lock 순서가 달라질 수 있습니다. key를 정렬하고 작은 batch로 나누면 lock 보유 시간과 cycle 가능성을 줄일 수 있습니다.

격리 수준을 높이는 선택은 항상 비용과 함께 말해야 합니다. READ COMMITTED는 동시성이 좋지만 같은 transaction 안에서 새 commit을 다시 볼 수 있습니다. REPEATABLE READ는 반복 조회 일관성을 주지만 제품별 phantom 처리와 lock 비용이 다릅니다. SERIALIZABLE은 강한 관측 계약을 주지만 serialization failure와 retry 비용이 생길 수 있습니다. 따라서 답변은 `SERIALIZABLE을 쓰면 됩니다`가 아니라 `이 업무는 어떤 anomaly를 허용하지 못하고, 그 비용으로 어떤 retry/latency를 받아들일 수 있는가`가 되어야 합니다.

실무에서는 lock을 DB 안에서만 보지 않습니다. Optimistic locking은 row의 version column을 이용해 update 시점에 `내가 읽은 version이 아직 현재인가`를 확인합니다. Distributed lock은 DB 밖 자원을 직렬화할 수 있지만 lease 만료, clock, network partition, fencing token 문제가 있습니다. Queue는 같은 key의 작업을 한 worker로 모아 순서를 만들 수 있지만 throughput과 latency를 바꿉니다. 이 수단들은 DB lock의 대체품이라기보다, 보호해야 할 불변식과 실패 비용에 따라 고르는 별도 도구입니다.

면접에서 답변을 압축하면 다음처럼 말할 수 있습니다. 격리 수준은 어떤 이상 현상을 허용할지 정하는 계약이고, lock은 특정 자원을 보호하는 실행 수단이며, deadlock은 그 실행 수단들이 서로 순환 대기할 때 생깁니다. 제품별로는 InnoDB의 index/gap/next-key lock과 PostgreSQL의 MVCC/SSI/retry 경계를 나누어야 합니다. 애플리케이션은 lock order, 짧은 transaction, idempotent retry, 조건부 update, constraint를 함께 써서 업무 불변식을 닫아야 합니다.

이 흐름을 검증하려면 실제로 네 가지를 관찰합니다. 첫째, blocking session과 blocked session을 찾습니다. 둘째, 어떤 SQL과 어떤 index range가 자원을 잡았는지 확인합니다. 셋째, deadlock log에서 wait-for cycle을 읽습니다. 넷째, retry 뒤에도 업무 side effect가 중복되지 않는지 봅니다. 관측 없이 `락이 걸렸다`고만 말하면 원인은 닫히지 않습니다.

```text
blocked query -> blocker transaction -> locked resource/index range -> invariant under protection -> retry or ordering repair
```

이 마지막 trace가 중요합니다. Lock wait를 보면 보통 빨리 kill하고 싶어집니다. 하지만 kill은 현상 완화일 뿐입니다. 어떤 불변식을 보호하느라 기다렸는지, 그 불변식이 정말 그 lock 범위를 필요로 하는지, code path가 같은 순서로 자원을 잡는지, retry가 안전한지까지 봐야 같은 문제가 반복되지 않습니다.

### 방어 수단을 고르는 기준

동시성 문제를 봤을 때 바로 `락을 걸자`로 가면 설계가 거칠어집니다. 먼저 충돌이 발생하는 자원이 무엇인지 구분합니다. 하나의 row라면 조건부 update나 optimistic version column이 가장 단순할 수 있습니다. 연속된 범위라면 index range와 next-key/gap lock이 중요해질 수 있습니다. 여러 row의 aggregate라면 guard row, materialized counter, serializable retry, constraint 재설계가 필요할 수 있습니다. 외부 시스템까지 걸리면 DB lock만으로는 부족합니다.

```text
single row invariant      -> conditional update, row lock, version column
range/predicate invariant -> range lock, exclusion/unique design, serializable retry
aggregate invariant       -> guard row, summary row lock, queue serialization
external side effect      -> idempotency, outbox, compensation, fencing token
```

이 표는 정답표가 아니라 질문 순서입니다. 예를 들어 쿠폰 1회 사용은 단일 `(coupon_id, user_id)` unique key로 막을 수 있으면 lock을 넓게 잡을 필요가 없습니다. 반면 `한 시간에 최대 100명 예약` 같은 aggregate 제한은 단순 unique key로 닫히지 않을 수 있습니다. 이때는 counter row를 갱신하며 row lock을 잡거나, reservation insert 후 집계 검사를 serializable retry로 감싸거나, 예약 요청을 show_id별 queue로 직렬화할 수 있습니다. 각 선택은 throughput, latency, 실패 시 retry 경험이 다릅니다.

면접에서 강한 답은 `저라면 lock을 겁니다`가 아니라 `불변식이 단일 row인지, range인지, aggregate인지 먼저 나누고, 그에 맞는 가장 좁은 방어를 고릅니다`라고 말하는 것입니다. 이렇게 말하면 DBMS 지식과 애플리케이션 설계 판단이 연결됩니다. 또한 lock은 강한 도구이지만 관측 가능성이 있어야 합니다. 어떤 lock이 누구를 막았는지 볼 수 없으면 장애 때 설명이 추측으로 흐릅니다. 그래서 설계 단계에서 lock timeout, deadlock retry, blocker query 관측, application trace id를 함께 준비해야 합니다.

## 더 깊게 볼 자료

공식 자료:

- [PostgreSQL current: Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [PostgreSQL current: Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html)
- [MySQL 8.4 Reference Manual: InnoDB Transaction Isolation Levels](https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html)
- [MySQL 8.4 Reference Manual: InnoDB Locking](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html)
- [MySQL 8.4 Reference Manual: Deadlocks in InnoDB](https://dev.mysql.com/doc/refman/8.4/en/innodb-deadlocks.html)
- [Spring Framework Reference: Transaction Propagation](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/tx-propagation.html)
- [Spring Framework Reference: Using @Transactional](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html)
- [Linux kernel docs: CFS Scheduler](https://docs.kernel.org/scheduler/sched-design-CFS.html)
- [Linux kernel docs: Multi-Queue Block IO Queueing Mechanism](https://www.kernel.org/doc/html/latest/block/blk-mq.html)
- [Linux man-pages: fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html)

저장소 안에서 이어 볼 자료:

- [jvm/spring/spring_transactional.md](../../jvm/spring/spring_transactional.md)
- [트랜잭션 경계와 ACID](06-transaction-acid-boundary.md)
- [database/lock.md](../../database/lock.md)
- [database/postgresql/lock.md](../../database/postgresql/lock.md)
- [database/deep-dive/transactions/13-isolation-anomalies.md](../../database/deep-dive/transactions/13-isolation-anomalies.md)
- [database/deep-dive/transactions/14-locks-latches-deadlocks.md](../../database/deep-dive/transactions/14-locks-latches-deadlocks.md)
- [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md)

이 자료를 볼 때는 `어느 transaction 경계에 들어갔고, 무엇을 읽었고, 무엇을 썼고, 어떤 자원이 기다림을 만들었는가`를 계속 추적합니다. 전파는 호출 경계 정책이고, 격리 수준은 관측 계약이며, lock은 그 계약을 만들기 위한 수단이고, deadlock은 기다림 관계의 cycle이라는 네 문장으로 다시 압축할 수 있으면 면접 답변의 중심이 잡힌 것입니다.
