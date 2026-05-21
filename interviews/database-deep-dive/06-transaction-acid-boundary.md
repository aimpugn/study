# 트랜잭션 경계와 ACID는 어디서 끝나고 애플리케이션 불변식은 어디서 시작되는가?

트랜잭션은 SQL을 여러 줄 묶는 문법이 아니라, 업무 상태가 한 번에 넘어가야 하는 경계를 DBMS에게 알려 주는 단위입니다. 계좌 이체라면 출금과 입금이 함께 성공하거나 함께 사라져야 하고, 주문 결제라면 결제 승인, 주문 상태, 재고 차감, 정산 이벤트의 관계를 어느 경계 안에서 닫을지 결정해야 합니다. DBMS는 그 경계 안에서 atomicity, consistency, isolation, durability를 제공하지만, 어떤 row들을 같은 경계에 넣어야 하는지와 외부 시스템의 부작용을 어떻게 묶을지는 애플리케이션 설계가 정합니다.

면접에서 이 주제를 잘 답하려면 ACID 네 글자를 순서대로 외우는 데서 멈추면 안 됩니다. `autocommit=true`일 때 각 statement가 이미 작은 트랜잭션이라는 사실, `BEGIN`과 `COMMIT` 사이에 어떤 연결 세션이 묶이는지, 예외가 발생했을 때 rollback rule이 어디서 결정되는지, `SAVEPOINT`가 전체 실패가 아니라 부분 되돌림을 어떻게 표현하는지까지 이어서 말해야 합니다. 마지막에는 DB 트랜잭션이 이메일, PG 승인, 메시지 발행, 파일 저장 같은 외부 side effect를 자동으로 되돌리지 못한다는 경계를 닫아야 합니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [Autocommit은 보이지 않는 BEGIN/COMMIT을 만든다](#autocommit은-보이지-않는-begincommit을-만든다)
    - [COMMIT은 성공 응답과 durable log 사이의 약속이다](#commit은-성공-응답과-durable-log-사이의-약속이다)
    - [DB COMMIT은 OS flush와 storage acknowledge까지 이어진다](#db-commit은-os-flush와-storage-acknowledge까지-이어진다)
    - [ROLLBACK은 DB 내부 변경을 되돌리지만 시간은 되돌리지 않는다](#rollback은-db-내부-변경을-되돌리지만-시간은-되돌리지-않는다)
    - [SAVEPOINT는 부분 실패를 표현하지만 업무 안전성을 대신 판단하지 않는다](#savepoint는-부분-실패를-표현하지만-업무-안전성을-대신-판단하지-않는다)
    - [Isolation은 boundary 안의 변경을 다른 transaction이 언제 보느냐의 문제다](#isolation은-boundary-안의-변경을-다른-transaction이-언제-보느냐의-문제다)
    - [Framework transaction은 connection과 호출 경계를 통해 실현된다](#framework-transaction은-connection과-호출-경계를-통해-실현된다)
    - [Application invariant는 DB constraint와 transaction boundary를 함께 써야 닫힌다](#application-invariant는-db-constraint와-transaction-boundary를-함께-써야-닫힌다)
    - [Transaction boundary는 너무 넓어도 실패한다](#transaction-boundary는-너무-넓어도-실패한다)
    - [Transaction boundary는 관측 가능한 값으로 검증해야 한다](#transaction-boundary는-관측-가능한-값으로-검증해야-한다)
- [DBMS별 경계](#dbms별-경계)
    - [PostgreSQL](#postgresql)
    - [MySQL/InnoDB](#mysqlinnodb)
    - [애플리케이션 프레임워크](#애플리케이션-프레임워크)
    - [외부 시스템](#외부-시스템)
- [직접 재생해 보기](#직접-재생해-보기)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
    - [다른 시스템으로 옮겨 보기: 업무 상태 전이를 어디서 원자적으로 닫을지 정하는 문제](#다른-시스템으로-옮겨-보기-업무-상태-전이를-어디서-원자적으로-닫을지-정하는-문제)
- [답변을 더 단단하게 만드는 판단 흐름](#답변을-더-단단하게-만드는-판단-흐름)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

짧게 답하면 이렇게 말할 수 있습니다. 트랜잭션은 함께 성공하거나 함께 사라져야 하는 업무 상태 전이를 DBMS가 관찰할 수 있는 단위로 묶는 것입니다. Atomicity는 그 단위가 중간 상태로 남지 않게 하고, isolation은 동시에 실행되는 다른 트랜잭션이 어떤 중간 상태를 볼 수 있는지 정하며, durability는 commit 이후 장애가 나도 결과를 복구할 수 있게 합니다. Consistency는 DB가 모든 업무 규칙을 알아서 이해한다는 뜻이 아니라, constraint와 transaction boundary가 지키는 데이터 규칙이 애플리케이션 불변식과 맞아야 한다는 뜻에 가깝습니다.

`autocommit`이 켜져 있으면 `UPDATE account SET balance = balance - 100 WHERE id = 'A'` 한 줄이 끝나는 순간 commit될 수 있습니다. 바로 다음 줄에서 `UPDATE account SET balance = balance + 100 WHERE id = 'B'`를 실행하기 전에 예외가 나면, DB 입장에서는 첫 번째 update가 이미 성공한 독립 트랜잭션입니다. 업무 입장에서는 돈이 사라진 것처럼 보입니다. 이 차이가 transaction boundary 질문의 출발점입니다.

그래서 서비스 메서드 하나가 하나의 업무 전이를 표현한다면 그 안의 SQL들이 같은 DB connection과 같은 transaction context를 공유해야 합니다. Java/Spring의 `@Transactional`도 이 원리를 바꾸지 않습니다. 프록시가 메서드 진입 시 connection을 가져와 transaction을 열고, 정상 종료 시 commit하며, rollback 대상 예외가 밖으로 나가면 rollback합니다. 같은 클래스 내부 호출, 비동기 실행, 다른 thread, checked exception 처리, 이미 시작된 transaction과의 propagation 설정은 이 경계를 바꿀 수 있습니다.

또 하나의 핵심은 DB 트랜잭션 밖의 세계입니다. 결제 승인 API를 호출한 뒤 DB 저장에 실패하면 DB rollback은 결제 승인 자체를 취소하지 못합니다. Kafka publish가 이미 broker에 들어간 뒤 DB transaction이 rollback되면 메시지와 DB 상태가 갈라질 수 있습니다. 이런 경계에서는 outbox, idempotency key, compensating transaction, saga 같은 설계가 등장합니다. 이들은 DB ACID가 부족해서라기보다, DB가 책임질 수 있는 경계와 외부 시스템 경계가 다르기 때문에 필요합니다.

면접 답변은 보통 다음 흐름이 가장 안전합니다. 먼저 작은 예시를 듭니다. 그 다음 autocommit과 explicit transaction이 그 예시를 어떻게 다르게 처리하는지 보입니다. 이어서 ACID를 DBMS 내부 보장과 애플리케이션 invariant로 나누고, 마지막으로 외부 side effect와 framework transaction boundary를 분리합니다. 이 순서로 말하면 단어는 많지 않아도 실제 시스템을 추적하는 답이 됩니다.

## 먼저 잡아야 할 작은 모델

가장 작은 모델은 계좌 이체입니다.

```text
초기 상태
  account A = 1000
  account B = 500
  총액 = 1500

업무 의도
  A에서 100을 빼고 B에 100을 더합니다.
  성공 후 총액은 여전히 1500이어야 합니다.

정상 완료
  A = 900
  B = 600
  총액 = 1500
```

이 모델에서 SQL은 두 줄뿐입니다.

```sql
UPDATE account SET balance = balance - 100 WHERE id = 'A';
UPDATE account SET balance = balance + 100 WHERE id = 'B';
```

두 줄이 같은 transaction에 있으면 첫 번째 update 뒤 예외가 나도 rollback으로 초기 상태로 돌아갈 수 있습니다. 두 줄이 각각 독립 transaction이면 첫 번째 줄만 commit되고 두 번째 줄이 실패할 수 있습니다. 이때 DBMS는 자기 계약을 어긴 것이 아닙니다. 애플리케이션이 업무상 하나여야 하는 상태 전이를 두 개의 transaction으로 나누었기 때문에 invariant가 깨진 것입니다.

이 작은 모델은 ACID를 더 정확히 읽게 해 줍니다.

```text
Atomicity
  A 차감과 B 증가가 함께 성공하거나 함께 사라집니다.

Consistency
  DB constraint와 애플리케이션 규칙을 통과한 상태에서 다른 상태로 넘어갑니다.
  총액 유지 같은 업무 규칙은 SQL 묶음과 코드가 함께 지켜야 할 수 있습니다.

Isolation
  동시에 다른 transaction이 A나 B를 읽거나 쓸 때 어떤 중간 상태를 볼 수 있는지 정합니다.

Durability
  COMMIT 성공 응답 뒤 DB가 crash해도 A=900, B=600 상태를 복구할 수 있어야 합니다.
```

여기서 consistency를 특히 조심해야 합니다. ACID 문맥에서 C는 transaction이 유효한 상태에서 유효한 상태로 DB를 옮겨야 한다는 의미로 읽는 편이 안전합니다. 하지만 DBMS가 업무 규칙 전체를 자동으로 알고 있다는 뜻은 아닙니다. `balance >= 0` 같은 check constraint는 DB가 직접 알 수 있습니다. 반면 `결제 승인됨이면 주문 상태는 PAID이고, 재고 차감 이벤트는 정확히 한 번 발행되어야 한다` 같은 규칙은 table 구조, unique key, transaction boundary, outbox 처리, 재시도 정책이 함께 지켜야 합니다.

같은 모델을 주문으로 바꾸면 외부 side effect가 보입니다.

```text
주문 결제 처리
  1. PG 승인 API 호출
  2. payment row INSERT
  3. order status PAID로 UPDATE
  4. inventory reserved_count UPDATE
  5. payment_completed event 발행
```

2, 3, 4는 같은 DB transaction 안에 넣을 수 있습니다. 1과 5는 DB 내부 변경이 아닙니다. PG 승인은 외부 시스템에 이미 반영될 수 있고, 메시지 발행은 broker의 durable log에 들어갈 수 있습니다. 따라서 이 흐름에서 transaction boundary는 `어디까지 DB가 rollback할 수 있는가`와 `rollback할 수 없는 행동은 어떤 보상 또는 멱등성으로 다룰 것인가`로 나누어야 합니다.

같은 결제 처리라도 각 단계가 기대는 영속성 경계는 다릅니다.

| 단계 | 주로 남는 위치 | DB rollback으로 되돌릴 수 있는가 | 안전하게 만들 때 필요한 보조 계약 |
| --- | --- | --- | --- |
| PG 승인 API 호출 | PG사의 거래 시스템 | 아닙니다. 취소 API나 정산 보정이 필요할 수 있습니다. | idempotency key, 승인 상태 조회, 보상 호출 |
| `payment` row insert | 현재 DB transaction | 같은 transaction 안이면 rollback됩니다. | unique request id, constraint, retry-safe SQL |
| 주문 상태 변경 | 현재 DB transaction | 같은 transaction 안이면 rollback됩니다. | 상태 전이 constraint, affected row count 확인 |
| 재고 차감 | 현재 DB transaction | 같은 transaction 안이면 rollback됩니다. | 조건부 update, 음수 방지, 충돌 retry |
| outbox row insert | 현재 DB transaction | 같은 transaction 안이면 rollback됩니다. | event id, publisher retry, consumer 멱등성 |
| broker 직접 발행 | broker의 durable log | DB rollback으로는 되돌리지 못합니다. | outbox 전환, producer idempotency, 보상/중복 처리 |

이 표를 보면 `결제 처리는 transaction으로 묶습니다`라는 말이 너무 넓다는 점이 보입니다. DB가 되돌릴 수 있는 상태와 외부 시스템에 이미 남은 상태를 나누어야, 실패 후 재시도와 보상 절차를 설계할 수 있습니다.

## 깊은 메커니즘

### Autocommit은 보이지 않는 BEGIN/COMMIT을 만든다

많은 DB client는 기본적으로 autocommit을 켭니다. 이 상태에서 statement 하나를 보내면 DBMS는 그 statement를 하나의 transaction으로 실행하고 성공하면 commit합니다. 사용자는 `BEGIN`을 쓰지 않았지만 transaction이 없었던 것이 아닙니다. 아주 작은 transaction이 statement마다 자동으로 열린 것입니다.

Autocommit이 기본값으로 자주 쓰이는 배경은 SQL을 대화형으로 실행하거나 단일 statement를 업무 단위로 보는 상황에서는 이 방식이 가장 예측하기 쉽기 때문입니다. `INSERT` 한 줄을 실행했는데 별도 `COMMIT`을 잊어 데이터가 보이지 않는다면 대화형 SQL을 쓰는 사람과 운영자 모두 혼란스럽습니다. 문제는 이 편의가 여러 statement로 이루어진 업무 전이까지 자동으로 묶어 주지는 않는다는 점입니다. DBMS는 statement 하나의 성공과 실패는 잘 알고 있지만, "출금과 입금은 같은 업무"라는 의미는 애플리케이션이 transaction boundary로 알려 줘야 합니다.

이 사실은 장애 trace에서 중요합니다.

```text
autocommit=true

client -> UPDATE account SET balance = balance - 100 WHERE id='A'
DB     -> statement transaction begin
DB     -> row update
DB     -> commit
client -> network error before second UPDATE

결과: A 차감은 이미 durable할 수 있습니다.
```

반대로 explicit transaction은 경계를 넓힙니다.

```text
autocommit=false 또는 BEGIN 사용

BEGIN
  UPDATE account SET balance = balance - 100 WHERE id='A';
  -- 여기서 예외 발생
ROLLBACK

결과: A 차감은 사라집니다.
```

면접에서 `autocommit이 위험한가요?`라고 물으면 항상 위험하다고 답하면 얕습니다. 단일 row insert처럼 statement 하나가 업무 단위와 같다면 autocommit은 자연스럽고 단순합니다. 위험한 경우는 업무 단위가 여러 statement인데 autocommit 때문에 각 statement가 서로 다른 transaction으로 분리될 때입니다. 판단 기준은 SQL 개수가 아니라 업무 invariant입니다.

### COMMIT은 성공 응답과 durable log 사이의 약속이다

COMMIT은 단순히 lock을 푸는 명령이 아닙니다. DBMS는 commit record를 durable한 로그에 남기고, crash recovery가 그 record를 보고 transaction 결과를 재구성할 수 있어야 합니다. PostgreSQL과 InnoDB 모두 세부 구조는 다르지만, 변경된 data page를 즉시 모두 disk에 반영하지 않아도 durable log를 통해 commit 결과를 복구할 수 있다는 원리를 씁니다. 이 영역은 [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)에서 더 깊게 다루고, 여기서는 transaction boundary 관점만 잡습니다.

중요한 경계는 `COMMIT 명령을 보냈다`, `COMMIT 성공 응답을 받았다`, `클라이언트가 성공 응답을 받기 전에 연결이 끊겼다`가 서로 다를 수 있다는 점입니다. 네트워크 장애 때문에 client가 결과를 모르면 DB에서는 이미 commit되었을 수도 있고 rollback되었을 수도 있습니다. 그래서 결제나 주문 같은 작업은 client retry가 중복 실행을 만들지 않도록 idempotency key나 unique request id를 둡니다.

```text
client sends COMMIT
network fails before response

possible DB state
  case 1: commit record durable -> transaction committed
  case 2: server did not commit -> transaction aborted/unknown

safe application design
  retry uses request_id unique key
  second attempt checks existing result before applying side effect again
```

이 문제는 DB가 약해서 생기는 것이 아닙니다. distributed system에서 client와 server 사이의 응답 손실은 항상 가능합니다. transaction boundary를 제대로 잡아도, commit outcome을 애플리케이션이 어떻게 확인하고 재시도할지 설계하지 않으면 사용자에게는 중복 결제나 주문 누락처럼 보일 수 있습니다.

그래서 transaction id, request id, payment id 같은 업무 식별자는 단순 로그 장식이 아닙니다. 성공 응답을 못 받은 client가 다시 요청했을 때, 서버가 "이미 commit된 같은 시도인지", "처음부터 실행되지 않은 새 시도인지"를 구분할 기준이 됩니다. 트랜잭션은 DB 내부 상태 전이를 원자적으로 만들지만, 네트워크 너머의 사용자는 commit 결과를 메시지로만 알 수 있습니다. 그 메시지가 유실될 수 있다는 사실 때문에 멱등성 키와 결과 조회 API가 transaction boundary 바깥의 안전장치로 붙습니다.

### DB COMMIT은 OS flush와 storage acknowledge까지 이어진다

Durability를 설명할 때는 `DB가 COMMIT했습니다`에서 멈추지 않는 편이 좋습니다. PostgreSQL 문서는 WAL record가 영구 저장소로 flush된 뒤에야 data file 변경을 나중에 써도 crash recovery로 복구할 수 있다고 설명합니다. Linux의 `fsync(2)`도 file의 수정된 in-core data를 storage device로 보내고, device가 완료를 보고할 때까지 호출이 block된다고 설명합니다. 즉 commit latency에는 DB lock뿐 아니라 WAL buffer, kernel page cache, filesystem, block layer, device driver, disk 또는 SSD의 flush 응답이 함께 들어올 수 있습니다.

```text
COMMIT path, 단순화

client thread
  -> DB backend records commit in WAL buffer
  -> WAL write moves bytes to kernel page cache
  -> fsync/fdatasync asks kernel to persist WAL file
  -> filesystem maps file blocks to block I/O
  -> block layer queues and dispatches request
  -> device driver submits to disk/SSD
  -> storage reports completion
  -> DB returns COMMIT success
```

이 그림은 `COMMIT이 느립니다`라는 증상을 해석할 때 중요합니다. 원인이 row lock이면 blocker transaction을 봐야 하지만, WAL fsync 시간이 길면 storage latency, checkpoint, dirty page pressure, write cache 설정을 봐야 합니다. CPU scheduler 관점도 들어옵니다. DB backend가 runnable인데 CPU를 못 받는 상태와, `fsync` 안에서 I/O completion을 기다리며 sleep하는 상태는 증상이 모두 latency로 보이지만 원인과 처방이 다릅니다. 그래서 좋은 답변은 `transaction이 오래 걸렸습니다`가 아니라 `lock wait인지, CPU run queue 지연인지, WAL fsync/I/O wait인지, 외부 API 대기인지`를 나누어 관측한다고 말해야 합니다.

COMMIT 지연을 볼 때는 다음처럼 "누가 무엇을 기다리는가"로 나누면 좋습니다.

| 관측 문장 | 가능성이 큰 대기 위치 | 먼저 볼 것 |
| --- | --- | --- |
| row lock wait가 큽니다. | 다른 transaction이 같은 row/range를 잡고 있습니다. | blocker transaction, lock order, transaction age |
| WAL sync 시간이 큽니다. | commit record를 영구 저장소에 밀어 넣는 경로가 느립니다. | WAL write/sync 지표, storage latency, checkpoint |
| CPU 사용률은 높고 run queue가 깁니다. | backend가 실행 기회를 기다립니다. | CPU quota, scheduler/run queue, 실행 중 query 수 |
| transaction duration이 길지만 DB wait가 작습니다. | 애플리케이션이 외부 API나 사용자 입력을 기다릴 수 있습니다. | service trace, 외부 호출 시간, transaction 안의 코드 범위 |

이 표의 목적은 DBMS와 OS를 억지로 섞는 것이 아닙니다. 같은 `느린 COMMIT`처럼 보여도 DB lock manager, WAL flush, CPU scheduling, 외부 API 대기는 서로 다른 층의 원인이므로 관측값을 나누어야 한다는 뜻입니다.

### ROLLBACK은 DB 내부 변경을 되돌리지만 시간은 되돌리지 않는다

ROLLBACK은 transaction 안에서 DBMS가 관리하던 변경을 취소합니다. 하지만 transaction 중 실행한 외부 API 호출, 로그 출력, 이메일 발송, 이미 다른 시스템에 도착한 메시지는 자동으로 취소하지 않습니다. 또한 sequence 값처럼 일부 DB object는 rollback되어도 gap이 남을 수 있습니다. 면접에서는 rollback을 `모든 일이 없던 일이 된다`고 말하면 바로 공격받기 쉽습니다.

예를 들어 주문 번호를 sequence로 발급하고 transaction이 rollback되면 다음 주문 번호가 건너뛸 수 있습니다. 이것은 데이터 정합성 오류가 아닙니다. sequence는 보통 transaction rollback과 독립적으로 증가하도록 설계되어 동시성을 얻습니다. 따라서 `id가 연속이어야 한다`는 요구가 있다면 sequence를 근거로 삼으면 안 됩니다. 업무 번호 연속성은 별도 설계가 필요하고, 많은 경우에는 감사 요구를 제외하면 연속성보다 유일성과 추적성이 더 중요합니다.

외부 API는 더 분명합니다.

```text
BEGIN
  UPDATE order SET status='PAID' WHERE id=10;
  call payment gateway capture API -> success
  INSERT payment_log ... -- 여기서 constraint violation
ROLLBACK

DB 상태: order status는 원래대로 돌아갈 수 있습니다.
외부 상태: PG 승인 또는 매입은 남아 있을 수 있습니다.
```

이 흐름에서는 DB rollback 뒤 PG 취소 API를 호출하거나, 아예 PG 호출을 DB commit 뒤 outbox event 소비자가 처리하게 할 수 있습니다. 어느 쪽이 낫다는 답은 도메인에 따라 달라집니다. 중요한 것은 DB transaction의 rollback 경계와 외부 시스템의 보상 경계를 같은 것으로 말하지 않는 것입니다.

이런 보상 설계가 등장하는 이유는 현대 서비스가 하나의 DB 안에서만 끝나지 않기 때문입니다. 결제사는 자기 원장을 가지고, 메시지 broker는 자기 log를 가지고, 이메일 시스템은 발송 큐를 가집니다. DBMS가 `ROLLBACK`을 수행해도 그 외부 원장과 log까지 시간을 되감지는 못합니다. 따라서 분산 트랜잭션을 모든 곳에 억지로 씌우기보다, 어떤 side effect는 취소 API로 되돌리고, 어떤 side effect는 outbox와 멱등 consumer로 중복을 흡수하며, 어떤 side effect는 정산 보정으로 수렴시키는 식의 경계 설계가 필요해집니다.

### SAVEPOINT는 부분 실패를 표현하지만 업무 안전성을 대신 판단하지 않는다

`SAVEPOINT`는 transaction 안에서 되돌릴 위치를 만듭니다. 어떤 단계가 실패했을 때 전체 transaction을 포기하지 않고 savepoint 이후 작업만 취소할 수 있습니다.

Savepoint가 유용해진 배경은 모든 업무 실패가 "처음부터 아무 일도 없었던 것"과 같지 않기 때문입니다. 대량 정산, 여러 주문 항목 처리, 일부 쿠폰 적용처럼 큰 작업 안에 작은 실패 단위가 있으면, 전체를 버리는 것보다 실패한 조각을 기록하고 남은 조각을 계속 처리하는 편이 업무적으로 맞을 수 있습니다. 다만 이 선택은 기술 기능보다 도메인 상태 모델이 먼저입니다. 부분 실패 상태를 설명할 수 없다면 savepoint는 안전장치가 아니라 이상한 중간 상태를 commit하는 통로가 됩니다.

```sql
BEGIN;
UPDATE order_header SET status = 'PROCESSING' WHERE id = 10;
SAVEPOINT item_step;
UPDATE order_item SET reserved = true WHERE order_id = 10;
-- 일부 item 처리 실패
ROLLBACK TO SAVEPOINT item_step;
UPDATE order_header SET status = 'PARTIAL_FAILED' WHERE id = 10;
COMMIT;
```

이 기능은 batch 처리나 부분 성공을 기록해야 하는 업무에서 유용합니다. 하지만 savepoint가 있다고 해서 남은 상태가 자동으로 올바른 것은 아닙니다. 위 예시에서 `PARTIAL_FAILED` 상태를 허용하는 업무 규칙과 후속 보정 흐름이 있어야 합니다. 그렇지 않다면 부분 rollback 뒤 commit은 기술적으로는 성공하지만 도메인적으로는 이상한 상태를 남깁니다.

Nested transaction이라는 말을 들을 때도 조심해야 합니다. 어떤 프레임워크는 중첩 transaction처럼 보이는 API를 제공하지만 실제 DB에서는 savepoint로 구현될 수 있습니다. 어떤 propagation 설정은 기존 transaction에 참여하고, 어떤 설정은 새 transaction을 열며, 어떤 설정은 transaction이 없으면 실패합니다. 면접에서는 제품별 세부 이름보다 `새 독립 commit 경계인가, 같은 commit 경계 안의 savepoint인가`를 먼저 물어야 합니다.

### Isolation은 boundary 안의 변경을 다른 transaction이 언제 보느냐의 문제다

트랜잭션 경계를 설명할 때 isolation을 빼면 동시성 질문에서 약해집니다. A와 B를 같은 transaction에서 update하더라도 다른 transaction이 중간에 A만 바뀐 상태를 볼 수 있다면 계좌 총액 invariant가 관측상 깨질 수 있습니다. 대부분의 OLTP DBMS는 MVCC와 lock을 이용해 이런 중간 관측을 제한합니다.

다만 isolation은 이 문서의 중심인 boundary와 구분해서 말해야 합니다. Boundary는 어떤 작업을 한 transaction으로 묶을지 정합니다. Isolation level은 그 transaction이 동시에 실행되는 다른 transaction과 어떤 관측 관계를 가질지 정합니다. Boundary가 틀리면 isolation을 높여도 업무 단위가 쪼개져 있습니다. Boundary가 맞아도 isolation이 낮으면 write skew나 phantom 같은 anomaly가 남을 수 있습니다.

예를 들어 두 의사가 동시에 당직에서 빠지는 예시는 단일 row update가 아닙니다.

```text
invariant: 한 과에는 항상 최소 한 명의 당직 의사가 있어야 합니다.
T1 reads: doctor A, B 둘 다 on_call=true
T2 reads: doctor A, B 둘 다 on_call=true
T1 updates: A on_call=false
T2 updates: B on_call=false
둘 다 commit하면 invariant가 깨집니다.
```

이 문제는 `A row와 B row update를 각각 transaction으로 잘 묶었는가`만으로 닫히지 않습니다. Predicate 또는 aggregate invariant를 어떻게 보호할지, serializable retry를 쓸지, 별도 guard row를 lock할지, unique/exclusion constraint를 둘지 설계해야 합니다. 이 지점부터는 [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)과 함께 봐야 합니다.

### Framework transaction은 connection과 호출 경계를 통해 실현된다

`@Transactional` 같은 annotation은 DBMS 기능을 호출하는 프레임워크 장치입니다. 일반적으로 proxy가 public method 진입을 가로채고, transaction manager가 connection을 thread-bound context에 묶고, 메서드가 정상 종료되면 commit합니다. Runtime exception은 rollback하고 checked exception은 rollback하지 않는 기본 규칙처럼 프레임워크별 정책도 있습니다.

이때 흔한 함정은 내부 메서드 호출입니다.

```java
class OrderService {
    public void pay() {
        this.savePayment(); // proxy를 통하지 않으면 @Transactional이 적용되지 않을 수 있습니다.
    }

    @Transactional
    public void savePayment() {
        ...
    }
}
```

또 다른 함정은 async boundary입니다. Transaction context가 thread-local에 묶여 있는데 `@Async`나 별도 executor로 넘어가면 같은 connection과 transaction이 이어지지 않을 수 있습니다. 메시지 listener, scheduler, web request, batch step도 각자 transaction boundary를 어떻게 여는지 확인해야 합니다. 면접 답변에서 annotation 이름만 말하면 부족하고, `어떤 호출이 proxy를 지나고 어떤 connection이 commit되는가`까지 내려가야 합니다.

호출 경계는 작은 call graph로 그리면 더 잘 보입니다.

```text
HTTP request thread
  -> proxy intercepts OrderService.pay()
       -> acquire connection C1
       -> BEGIN on C1
       -> PaymentRepository.insert() uses C1
       -> InventoryRepository.decrease() uses C1
       -> method returns normally
       -> COMMIT on C1

same class internal call
  OrderService.pay()
    -> this.savePayment()
       proxy interception may be skipped
       expected transaction rule may not run

async boundary
  OrderService.pay() on thread A, connection C1
    -> @Async publisher on thread B
       thread B does not automatically inherit C1 transaction
```

이 trace의 핵심은 annotation이 source code에 붙어 있다는 사실보다 runtime에서 어떤 호출이 interceptor를 지나고 어떤 connection이 commit되는지입니다. 프레임워크별 세부는 다를 수 있지만, 면접 답변은 항상 `호출 경로`, `connection`, `commit/rollback 신호`를 분리해 말해야 안전합니다.

### Application invariant는 DB constraint와 transaction boundary를 함께 써야 닫힌다

DB constraint는 강력합니다. Unique key는 중복 요청 방지에 좋고, foreign key는 부모 없는 자식 row를 막으며, check constraint는 단일 row의 값 범위를 지킬 수 있습니다. 하지만 모든 invariant가 constraint 하나로 표현되지는 않습니다. 여러 row의 합계, 외부 시스템 상태, 시간 순서, 이벤트 발행 여부 같은 규칙은 transaction과 애플리케이션 로직이 함께 지켜야 합니다.

실무에서 안전한 설계는 가능하면 invariant를 DB가 알 수 있는 형태로 내립니다. 중복 결제를 막고 싶다면 `payment_request_id`에 unique key를 둡니다. 한 주문에 성공 결제가 하나만 있어야 한다면 `(order_id, success_flag)`를 조건부 unique index로 표현할 수 있는지 봅니다. 재고 차감을 막고 싶다면 `UPDATE inventory SET qty = qty - 1 WHERE sku = ? AND qty > 0`처럼 조건부 update와 affected row count 확인을 씁니다. 이런 방식은 애플리케이션 if문만 믿는 것보다 동시성에 강합니다.

하지만 외부 side effect까지 DB constraint로 넣을 수는 없습니다. 이때 outbox pattern은 transaction boundary를 다음과 같이 바꿉니다.

```text
DB transaction
  update order status
  insert outbox event with unique event_id
COMMIT

separate publisher
  read unsent outbox rows
  publish message
  mark sent or rely on idempotent consumer
```

이 설계는 메시지 발행을 DB transaction 안에 직접 넣지 않습니다. 대신 `발행해야 할 사실`을 DB transaction 안에 기록합니다. 그 다음 별도 publisher가 적어도 한 번 발행하고, consumer는 event id로 중복을 견딥니다. 원자성은 DB 상태와 outbox row 사이에서 닫히고, 외부 메시지 전달은 멱등성과 재시도로 닫힙니다. 외부 부작용을 어디서 자르고 다시 붙일지는 [애플리케이션 경계, 멱등성, 금액 처리, outbox](12-application-boundaries-idempotency-money-outbox.md)에서 이어서 봅니다.

### Transaction boundary는 너무 넓어도 실패한다

트랜잭션을 넓게 잡으면 atomicity는 커지지만 비용도 커집니다. 긴 transaction은 lock을 오래 들고, MVCC old version 정리를 늦추며, connection pool을 점유합니다. 사용자가 결제 화면에서 오래 머무는 동안 DB transaction을 열어 두거나, 외부 API 응답을 기다리는 동안 row lock을 잡고 있으면 다른 요청이 불필요하게 기다릴 수 있습니다. 따라서 좋은 boundary는 `업무상 함께 commit되어야 하는 최소 DB 상태`를 묶고, 느린 외부 작업은 멱등성과 보상 흐름으로 분리하는 쪽을 먼저 검토합니다.

예를 들어 재고 차감과 주문 상태 변경은 같은 transaction에 둘 수 있지만, 이메일 발송을 같은 transaction 안에서 기다릴 필요는 보통 없습니다. 결제 승인도 도메인에 따라 순서가 갈립니다. 먼저 PG 승인 후 DB commit을 한다면 승인 성공 후 DB 실패를 보상해야 하고, 먼저 DB에 승인 요청 상태를 commit한 뒤 worker가 PG를 호출한다면 pending 상태와 재시도 UX를 설계해야 합니다. 둘 중 어느 쪽이든 핵심은 DB rollback이 닿는 범위와 닿지 않는 범위를 숨기지 않는 것입니다.

긴 transaction의 비용은 시간축으로 보면 더 직관적입니다.

```text
bad boundary
  BEGIN
    lock order row
    call PG API and wait 2s
    update inventory
  COMMIT

  다른 요청은 2초 동안 order row 또는 관련 index/range를 기다릴 수 있습니다.
  그동안 connection도 점유되고, MVCC old version cleanup도 늦어질 수 있습니다.

smaller DB boundary
  write payment_request PENDING
  COMMIT

  worker calls PG with idempotency key

  BEGIN
    mark payment APPROVED
    update order/inventory
    insert outbox event
  COMMIT
```

두 번째 흐름이 항상 정답은 아닙니다. 사용자 경험, 승인 취소 가능성, 재고 보류 정책, timeout 정책에 따라 달라집니다. 하지만 이 trace는 transaction을 길게 잡는 선택이 단순히 "안전한 묶음"이 아니라 lock 보유 시간, connection 점유, 외부 실패 보상까지 함께 키운다는 점을 보여 줍니다.

### Transaction boundary는 관측 가능한 값으로 검증해야 한다

트랜잭션 문제는 코드만 읽으면 그럴듯해 보이는 경우가 많습니다. 그래서 관측값을 붙여야 합니다. PostgreSQL에서는 `pg_stat_activity`의 `xact_start`, `state`, wait event를 보고, MySQL/InnoDB에서는 transaction list, lock wait, deadlock output, history list length를 봅니다. 애플리케이션에서는 connection pool active count, transaction duration, rollback count, outbox backlog, idempotency conflict count를 봅니다. 좋은 운영 질문은 `transaction이 열렸는가`가 아니라 `얼마나 오래 열려 있었고, 어떤 자원을 잡고 있었으며, 실패했을 때 어떤 외부 효과가 남았는가`입니다.

## DBMS별 경계

### PostgreSQL

PostgreSQL은 명시적인 transaction block을 `BEGIN`, `COMMIT`, `ROLLBACK`으로 다룹니다. 공식 tutorial은 여러 update가 transaction 안에 있을 때 모두 반영되거나 모두 반영되지 않는 예를 보여 줍니다. PostgreSQL에서는 많은 DDL도 transaction 안에서 rollback될 수 있지만, 모든 명령이 같은 방식으로 되돌아가는 것은 아니므로 운영 명령은 공식 문서를 확인해야 합니다.

PostgreSQL의 MVCC는 transaction isolation과 직접 연결됩니다. `READ COMMITTED`에서는 statement마다 새 snapshot을 보는 식으로 동작하고, `REPEATABLE READ`에서는 첫 non-transaction statement를 기준으로 더 안정된 snapshot을 봅니다. `SERIALIZABLE`에서는 read/write dependency 감지와 serialization failure가 중요해집니다. 이 문서에서는 boundary를 다루므로 세부 visibility는 [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md), 동시성 실패는 [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)으로 넘깁니다.

운영에서 자주 보는 신호는 `pg_stat_activity`의 오래 열린 transaction, `idle in transaction`, lock wait, rollback count, long-running transaction 때문에 늦어지는 vacuum입니다. 오래 열린 transaction은 단순히 connection 하나가 놀고 있는 문제가 아니라 old row version 정리를 막고 lock을 오래 들고 있을 수 있습니다.

### MySQL/InnoDB

InnoDB는 transaction model에서 multi-versioning과 locking을 함께 씁니다. MySQL은 기본적으로 autocommit이 켜져 있으며, explicit transaction을 시작하면 여러 statement를 하나의 commit 경계로 묶을 수 있습니다. InnoDB consistent read와 locking read는 다르게 동작하므로, plain select가 보이는 snapshot과 `SELECT ... FOR UPDATE`가 만드는 lock을 구분해야 합니다.

MySQL에서는 DDL의 atomicity와 implicit commit 경계도 조심해야 합니다. 일부 DDL은 실행 전후로 implicit commit을 유발할 수 있고, version에 따라 atomic DDL 지원 범위가 다릅니다. 따라서 application transaction 안에서 schema 변경까지 함께 rollback될 것처럼 말하면 위험합니다. OLTP transaction 설명과 migration/DDL 설명은 분리하는 편이 안전합니다.

관측값으로는 `SHOW ENGINE INNODB STATUS`, performance schema의 transaction/lock 관련 view, deadlock log, history list length, binary log position 등이 있습니다. Boundary 질문에서는 특히 autocommit 상태, transaction isolation level, connection pool이 실제로 어떤 connection을 재사용하는지 확인해야 합니다.

### 애플리케이션 프레임워크

Spring/JPA/MyBatis 같은 프레임워크는 DB transaction을 더 쓰기 쉽게 만들지만 경계를 없애지 않습니다. Transaction manager가 어떤 data source를 대상으로 하는지, JPA flush가 언제 일어나는지, MyBatis executor가 같은 connection을 쓰는지, propagation이 기존 transaction에 참여하는지 새 transaction을 만드는지 확인해야 합니다.

JPA에서는 entity 변경이 즉시 SQL로 나가지 않고 flush 시점에 DB에 반영될 수 있습니다. 따라서 코드에서 값을 바꾼 시점, SQL이 DB로 나간 시점, commit된 시점이 다를 수 있습니다. 면접에서 ORM을 함께 물으면 `객체 상태 변경`, `flush`, `DB transaction commit`을 분리해 말해야 합니다.

### 외부 시스템

PG, Kafka, email, file storage, cache invalidation은 DB transaction과 다른 durable boundary를 가집니다. Redis cache를 먼저 지우고 DB commit이 실패하면 cache miss 뒤 old DB 값을 다시 채울 수 있습니다. DB commit 후 cache invalidation이 실패하면 stale cache가 남을 수 있습니다. Kafka publish와 DB update가 분리되면 둘 중 하나만 성공하는 창이 생깁니다.

이 영역의 답은 `분산 트랜잭션을 쓰면 됩니다`로 쉽게 닫으면 안 됩니다. 2PC는 참여자와 coordinator의 failure mode를 늘리고, 모든 외부 시스템이 XA 같은 protocol을 잘 지원하지도 않습니다. 많은 서비스에서는 outbox, idempotency, compensating action, reconciliation job을 조합합니다. 중요한 것은 어떤 invariant를 강하게 지키고 어떤 결과는 재시도나 보정으로 수렴시킬지 명시하는 것입니다. 이 판단은 DB 내부 ACID보다 넓으므로 [애플리케이션 경계, 멱등성, 금액 처리, outbox](12-application-boundaries-idempotency-money-outbox.md)와 연결됩니다.

## 직접 재생해 보기

아래 실험은 운영 DB가 아니라 로컬 disposable DB에서 합니다. 핵심은 SQL을 실행하는 것보다 관측할 결과를 먼저 정하는 것입니다.

1. Autocommit과 explicit transaction을 비교합니다.

    ```sql
    DROP TABLE IF EXISTS account;
    CREATE TABLE account (id text PRIMARY KEY, balance int NOT NULL);
    INSERT INTO account VALUES ('A', 1000), ('B', 500);

    -- autocommit 환경에서 첫 UPDATE 뒤 client 오류가 났다고 가정합니다.
    UPDATE account SET balance = balance - 100 WHERE id = 'A';
    SELECT * FROM account ORDER BY id;
    ```

    PASS 신호는 A만 900이 된 상태가 관측되는 것입니다. 이것은 DB가 틀렸다는 뜻이 아니라 각 statement가 독립 transaction이었다는 뜻입니다. 이어서 table을 초기화한 뒤 `BEGIN`, 두 update, 중간 실패, `ROLLBACK`을 실행합니다. PASS 신호는 A와 B가 초기 상태로 돌아가는 것입니다.

2. Savepoint의 부분 rollback을 재생합니다.

    ```sql
    BEGIN;
    UPDATE account SET balance = balance - 100 WHERE id = 'A';
    SAVEPOINT after_debit;
    UPDATE account SET balance = balance + 100 WHERE id = 'B';
    ROLLBACK TO SAVEPOINT after_debit;
    COMMIT;
    SELECT * FROM account ORDER BY id;
    ```

    기술적 PASS 신호는 A 차감만 commit되는 것입니다. 업무적 PASS 여부는 별도입니다. 계좌 이체라면 이 결과는 invariant를 깨므로 실패 설계입니다. Batch 처리에서 실패 항목만 표시하는 업무라면 허용될 수 있습니다. 이 실험은 savepoint가 correctness를 대신 판단하지 않는다는 점을 보여 줍니다.

3. Commit outcome unknown 상황을 설계로 재생합니다.

    SQL만으로 네트워크 단절을 완벽히 만들 필요는 없습니다. 대신 요청 table을 만듭니다.

    ```sql
    CREATE TABLE payment_request (
      request_id text PRIMARY KEY,
      order_id bigint NOT NULL,
      amount int NOT NULL,
      status text NOT NULL
    );
    ```

    같은 `request_id`로 두 번 insert를 시도하면 두 번째는 unique key에 막힙니다. 이 결과는 client가 commit 응답을 잃고 재시도해도 같은 요청을 중복 반영하지 않을 수 있음을 보여 줍니다. PASS 신호는 재시도가 새 결제를 만들지 않고 기존 request를 조회하거나 idempotent하게 종료되는 것입니다.

4. Framework transaction boundary를 call graph로 확인합니다.

    Spring이라면 같은 클래스 내부 호출, public method proxy 호출, `@Async` 호출, checked exception을 각각 그립니다. PASS 신호는 어느 경로에서 transaction interceptor가 실행되는지 설명할 수 있는 것입니다. 실제 코드에서는 test transaction manager log, datasource proxy log, SQL commit/rollback log를 켜고 확인합니다. FAIL 신호는 annotation이 붙어 있다는 이유만으로 모든 경로가 같은 transaction이라고 단정하는 것입니다.

5. 외부 side effect를 DB transaction 밖으로 분리합니다.

    Outbox table을 만들고, order update와 outbox insert를 같은 transaction에 넣습니다. Publisher는 commit된 outbox row만 읽는다고 가정합니다.

    ```text
    transaction: order PAID + outbox row INSERT -> COMMIT
    publisher: outbox row read -> publish -> mark sent
    consumer: event_id unique key로 중복 처리
    ```

    PASS 신호는 DB 상태와 발행해야 할 사실이 같은 transaction에 남는 것입니다. 메시지 중복은 consumer idempotency로 처리합니다. FAIL 신호는 DB commit 전 메시지를 먼저 발행하고 rollback 때 아무 보정도 없는 것입니다.

## 면접 꼬리 질문

- ACID 중 Consistency는 DB가 업무 규칙을 모두 보장한다는 뜻인가요?

    아닙니다. DB constraint가 표현할 수 있는 규칙은 DB가 강하게 지킬 수 있지만, 어떤 row와 외부 행동을 같은 업무 경계로 묶을지는 애플리케이션 설계가 정합니다. 좋은 답은 check/unique/foreign key 같은 DB 규칙과 주문 상태 전이, 결제 승인, 이벤트 발행 같은 업무 invariant를 나누어 말합니다.

- Autocommit은 언제 문제가 되나요?

    Statement 하나가 업무 단위와 같으면 문제가 아닙니다. 여러 statement가 함께 성공해야 하는데 autocommit 때문에 각 statement가 독립 commit될 때 문제가 됩니다. 예시는 계좌 이체의 출금/입금, 주문 상태 변경/재고 차감, 중복 요청 처리 같은 흐름입니다.

- `@Transactional`은 실제로 어디서 transaction을 시작하나요?

    보통 프록시나 AOP interceptor가 메서드 진입을 가로채 connection을 가져오고 DB transaction을 시작합니다. 따라서 내부 호출, final/private method, 다른 thread, propagation 설정, rollback 대상 예외에 따라 기대와 다를 수 있습니다. 답변에는 annotation 이름보다 connection과 commit/rollback 경계를 넣어야 합니다.

- Savepoint와 nested transaction은 같은가요?

    항상 같지 않습니다. Savepoint는 같은 transaction 안의 부분 rollback 지점입니다. 어떤 프레임워크가 nested transaction을 savepoint로 구현할 수 있지만, 독립적으로 commit되는 새 transaction과는 다릅니다. 질문을 받으면 `같은 outer commit 경계 안인가, 별도 commit 경계인가`를 먼저 확인해야 합니다.

- DB transaction 안에서 외부 API를 호출하면 안전한가요?

    DB rollback은 외부 API를 되돌리지 못하므로 그 자체로 안전하지 않습니다. 외부 API가 먼저 성공하고 DB가 rollback되면 보상 호출이나 reconciliation이 필요합니다. 반대로 DB commit 뒤 외부 API가 실패하면 retry/outbox가 필요합니다. 안전성은 호출 순서와 멱등성, 보상 경로까지 포함해 판단합니다.

- Commit 성공 응답을 받기 전에 네트워크가 끊기면 어떻게 하나요?

    결과는 unknown일 수 있습니다. DB에는 commit되었을 수도 있고 아닐 수도 있습니다. 애플리케이션은 같은 요청을 식별하는 idempotency key를 두고, 재시도 시 기존 처리 결과를 조회하거나 중복 반영을 막아야 합니다.

## 함정 질문

- "트랜잭션은 SQL 여러 개를 묶는 것입니다"라고만 답하는 것

    이 답은 너무 작습니다. SQL 여러 개가 아니라 업무 상태 전이가 기준입니다. SQL 두 줄이라도 서로 다른 업무면 나눌 수 있고, SQL 한 줄이라도 외부 side effect와 묶이면 DB transaction만으로는 부족할 수 있습니다.

- "ACID라서 정합성은 DB가 알아서 맞춥니다"라고 말하는 것

    DB가 모르는 업무 규칙은 지킬 수 없습니다. Constraint로 표현한 규칙, isolation으로 보호한 동시성 규칙, 애플리케이션이 정한 상태 전이 규칙을 나누어야 합니다. 이 구분이 없으면 consistency를 가장 그럴듯하지만 가장 위험하게 오해합니다.

- "Rollback하면 모든 것이 원래대로 돌아갑니다"라고 말하는 것

    DB transaction 안의 변경은 되돌릴 수 있지만 sequence gap, 외부 API, 메시지, 로그, cache 같은 것은 별도 경계를 가집니다. 면접에서는 rollback 가능한 상태와 보상해야 하는 상태를 분리해서 답해야 합니다.

- "`@Transactional`을 붙였으니 항상 transaction입니다"라고 말하는 것

    Annotation은 호출 경로와 runtime proxy가 맞아야 효과가 있습니다. 내부 호출, async boundary, transaction manager 설정, exception rule을 보지 않으면 장애 원인을 놓칠 수 있습니다.

- "격리 수준을 높이면 transaction 설계 문제도 사라집니다"라고 말하는 것

    Isolation은 동시에 실행되는 transaction 사이의 관측 규칙입니다. 업무 단위를 잘못 나눈 boundary 문제, 외부 side effect 문제, idempotency 문제를 isolation level만으로 해결할 수는 없습니다.

### 다른 시스템으로 옮겨 보기: 업무 상태 전이를 어디서 원자적으로 닫을지 정하는 문제

다른 시스템으로 옮길 때는 먼저 `함께 성공하거나 함께 사라져야 하는 상태`를 찾습니다. 그 상태가 하나의 DB 안에만 있으면 transaction boundary가 중심이고, 외부 API나 broker까지 걸리면 idempotency, outbox, saga 같은 보조 계약이 필요합니다. 좋은 확인 질문은 `이 작업이 중간에 멈추면 어떤 사실이 남는가`입니다.

## 답변을 더 단단하게 만드는 판단 흐름

트랜잭션 질문은 결국 `어떤 상태 전이를 하나의 사실로 만들 것인가`를 묻습니다. 이 질문을 받으면 먼저 table 이름이나 annotation 이름으로 들어가지 말고, 업무적으로 중간 상태가 남으면 안 되는 불변식을 찾습니다. 이체라면 총액 보존, 주문이라면 주문 헤더와 항목의 일관성, 결제라면 승인 증거와 내부 상태의 대응, 재고라면 음수 방지와 중복 차감 방지가 그 불변식입니다. 그 다음에 그 불변식을 DB가 직접 볼 수 있는 row와 constraint로 얼마나 내릴 수 있는지 확인합니다.

예를 들어 재고 차감은 다음 두 문장이 비슷해 보여도 안전성이 다릅니다.

```sql
-- 위험한 형태: 먼저 읽고 애플리케이션에서 판단합니다.
SELECT quantity FROM inventory WHERE sku = 'A';
UPDATE inventory SET quantity = quantity - 1 WHERE sku = 'A';

-- 더 나은 형태: DB가 보는 조건과 변경을 같은 statement에 묶습니다.
UPDATE inventory
   SET quantity = quantity - 1
 WHERE sku = 'A'
   AND quantity > 0;
```

두 번째 형태도 완전한 답은 아닙니다. affected row count가 1인지 확인해야 하고, 주문 row 생성과 재고 차감이 같은 transaction 안에 있는지 봐야 하며, 결제 승인이나 메시지 발행처럼 DB 밖 행동은 별도 계약으로 빼야 합니다. 하지만 첫 번째 형태보다 훨씬 강한 이유는 `재고가 남아 있는지 확인한다`와 `재고를 줄인다`가 DB가 원자적으로 판단할 수 있는 한 동작에 가까워졌기 때문입니다.

면접에서는 transaction boundary를 세 단계로 말하면 좋습니다. 첫째, 업무 불변식을 말합니다. 둘째, 그 불변식을 지키는 row set, constraint, SQL 조건을 말합니다. 셋째, DB transaction 밖으로 나가는 행동을 말합니다. 이 세 단계를 빠뜨리지 않으면 `@Transactional을 붙였습니다` 같은 구현 표지가 결론을 대신하지 못하게 됩니다.

```text
업무 불변식 -> DB가 볼 수 있는 row/constraint -> transaction boundary -> 외부 side effect 보상/멱등성
```

이 흐름에서 자주 깨지는 지점은 예외 처리입니다. 코드가 예외를 catch한 뒤 정상 return하면 framework는 commit할 수 있습니다. 반대로 rollback 대상이 아닌 checked exception을 던지면 기본 정책에서는 commit될 수 있습니다. 어떤 프레임워크든 핵심은 같습니다. `어떤 예외가 밖으로 나갔는가`, `transaction manager가 그 예외를 rollback 신호로 해석하는가`, `이미 flush된 SQL이 commit되는가 rollback되는가`를 봐야 합니다.

또 하나의 지점은 읽기 전용 transaction입니다. `readOnly=true`는 보통 성능 최적화나 flush 억제 힌트로 쓰이며, 모든 DB에서 쓰기를 물리적으로 불가능하게 만드는 보안 장치라고 단정하면 안 됩니다. 읽기 전용 경계가 중요한 이유는 의도를 드러내고 connection routing, flush mode, lock 사용을 줄이는 데 도움을 줄 수 있기 때문입니다. 하지만 실제 쓰기 차단은 DB 권한, transaction mode, 테스트로 확인해야 합니다.

마지막으로 transaction은 길수록 안전해지는 것이 아닙니다. 너무 짧으면 업무 불변식이 쪼개지고, 너무 길면 lock을 오래 잡고 MVCC cleanup을 막으며 외부 API timeout까지 transaction 안으로 끌어들입니다. 좋은 경계는 `함께 commit되어야 하는 DB 상태`를 포함하되, 오래 걸리거나 실패가 애매한 외부 행동은 outbox나 상태 머신으로 분리합니다. 이 균형을 말할 수 있으면 ACID 답변이 암기에서 설계 판단으로 올라갑니다.

## 더 깊게 볼 자료

공식 자료:

- [PostgreSQL current: Transactions](https://www.postgresql.org/docs/current/tutorial-transactions.html)
- [PostgreSQL current: Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [PostgreSQL current: Write-Ahead Logging](https://www.postgresql.org/docs/current/wal-intro.html)
- [PostgreSQL current: WAL Configuration](https://www.postgresql.org/docs/current/wal-configuration.html)
- [MySQL 8.4 Reference Manual: The InnoDB Transaction Model](https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-model.html)
- [MySQL 8.4 Reference Manual: InnoDB Locking and Transaction Model](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking-transaction-model.html)
- [Linux man-pages: fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html)
- [Linux kernel docs: Multi-Queue Block IO Queueing Mechanism](https://www.kernel.org/doc/html/latest/block/blk-mq.html)

저장소 안에서 이어 볼 자료:

- [database/deep-dive/transactions/11-transaction-lifecycle-acid.md](../../database/deep-dive/transactions/11-transaction-lifecycle-acid.md)
- [database/mvcc.md](../../database/mvcc.md)
- [database/lock.md](../../database/lock.md)
- [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md)
- [격리 수준, 락, 데드락](08-isolation-lock-deadlock.md)

자료를 읽을 때는 용어를 그대로 모으기보다 같은 질문을 반복합니다. 이 기능은 어떤 상태 전이를 한 경계로 묶는가, 실패하면 무엇이 rollback되고 무엇이 남는가, commit 결과를 어떤 관측값으로 확인하는가. 이 세 질문이 닫히면 transaction boundary 답변은 단순 암기가 아니라 설계 판단으로 바뀝니다.
