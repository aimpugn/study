# Application Boundary, Idempotency, Money, Outbox Deep Dive

데이터베이스 면접에서 "transaction을 어디에 건다"는 질문은 DB 문법 질문이 아니다. 실제 장애는 DBMS 하나의 내부 원리보다 애플리케이션 경계에서 자주 생긴다. connection을 빌렸는데 session state가 남아 있고, Spring `@Transactional`이 걸렸다고 믿었지만 proxy를 지나지 않았고, 외부 결제 API timeout 후 성공 여부를 모르는 상태에서 재시도했으며, 금액 반올림이 한 줄씩 달라 ledger와 balance가 어긋나는 식이다.

이 문서는 application boundary를 DB 바깥의 부수적인 주제로 보지 않는다. connection/session/pool은 transaction의 물리적 통로이고, Spring/JPA/MyBatis는 그 통로를 추상화하는 계층이며, idempotency와 outbox는 "한 번만 일어난 것처럼 보이게 만드는" 운영 계약이다. money 계산은 단순 `BigDecimal` 사용법이 아니라 ledger, balance, payment state, reconciliation이 함께 맞아야 하는 회계 경계다.

## 2-5분 개요

애플리케이션에서 transaction을 설명할 때 가장 먼저 잡아야 할 모델은 "비즈니스 메서드 하나가 논리 transaction처럼 보여도, 실제 DB에서는 connection 하나의 session 상태와 commit/rollback 경계로 실행된다"이다. Spring은 proxy와 transaction manager를 통해 connection을 thread-bound resource처럼 묶고, JPA는 persistence context와 flush timing을 통해 SQL 실행 시점을 늦출 수 있으며, MyBatis는 mapper 호출 시점에 SQL이 직접 나가되 Spring transaction에 참여할 수 있다. 그래서 `@Transactional`은 마법 문구가 아니라 "어떤 호출이 proxy를 통과했고, 어떤 transaction manager가 어떤 connection을 언제 bind했고, 언제 flush/commit/rollback했는가"라는 실행 경계다. DB 내부 transaction의 의미는 [트랜잭션과 ACID 경계](06-transaction-acid-boundary.md)를 먼저 고정해 두면 더 안전하게 이어 읽을 수 있다.

Connection pool은 성능 최적화 장치이면서 session contamination의 위험 지점이다. DB connection은 단순 TCP socket이 아니라 transaction isolation, autocommit, timezone, search_path, role, temporary object, prepared statement 같은 session state를 가질 수 있다. 애플리케이션이 raw SQL로 session setting을 바꾸고 되돌리지 않으면 다음 borrower가 그 상태를 이어받을 수 있다. Pool이 connection을 재사용하기 때문에 빠르지만, 그만큼 "반납 시 깨끗한 상태"라는 계약이 중요하다.

중복 요청과 idempotency는 "버튼을 두 번 눌렀을 때 한 번만 처리한다"보다 넓다. Client retry, gateway timeout, mobile network retry, queue redelivery, worker crash, DB deadlock retry가 모두 같은 business command를 다시 실행시킬 수 있다. 안전한 모델은 먼저 idempotency key와 request fingerprint로 "같은 의도인지"를 판정하고, DB에 PENDING 같은 처리 기록을 원자적으로 선점한 뒤, 결과를 SUCCEEDED/FAILED/UNKNOWN 같은 상태로 닫는 것이다. 외부 side effect가 끼면 단순 unique constraint만으로는 부족하다.

돈 계산은 floating point를 피하고 decimal을 쓰면 끝나는 문제가 아니다. 어느 단위로 저장할지, rounding mode를 어느 boundary에서 적용할지, 세금/수수료/할인의 순서를 어떻게 고정할지, 나눗셈 후 남는 1원 단위를 어떻게 배분할지, ledger와 balance 중 어느 쪽을 진실의 원장으로 볼지 정해야 한다. 금액은 "계산 결과가 맞다"보다 "같은 입력을 다시 계산하면 같은 원장 entry와 같은 잔액이 나온다"가 중요하다.

분산 transaction과 outbox는 완벽한 정답 하나가 아니라 trade-off다. 2PC/XA는 여러 resource가 prepare/commit protocol에 참여할 수 있을 때 atomic commit을 제공하지만, 참여 resource와 운영 복잡도, blocking risk가 있다. Saga는 여러 local transaction과 보상 transaction으로 긴 업무 흐름을 쪼개지만, 중간 상태와 보상 실패를 설계해야 한다. Outbox는 DB transaction 안에 "나중에 외부로 내보낼 메시지"를 함께 저장하고 worker가 재시도하며 발행하는 패턴이다. Outbox는 exactly-once side effect를 공짜로 주지 않는다. 대신 durable intent와 at-least-once delivery, consumer idempotency를 결합해 실제 운영 실패를 다룰 수 있게 한다.

## 먼저 잡아야 할 작은 모델

작은 주문 결제 시스템을 상상하자. 사용자가 결제 버튼을 누르면 애플리케이션은 `payment` row를 만들고, 외부 PG사에 승인 요청을 보내고, 승인 성공 후 ledger entry를 남기고, 주문 상태를 `PAID`로 바꾸며, 알림 메시지를 발행한다.

처음 보기에는 하나의 메서드로 충분해 보인다.

```java
@Transactional
public PaymentResult pay(PayCommand command) {
    Payment payment = paymentRepository.create(command);
    PgResult pgResult = pgClient.approve(command);
    ledgerRepository.append(payment, pgResult);
    orderRepository.markPaid(command.orderId());
    eventPublisher.publish(new PaymentApprovedEvent(payment.id()));
    return PaymentResult.success(payment.id());
}
```

이 코드는 설명용으로 일부러 단순하게 쓴 것이다. 실제 위험은 네 가지다.

첫째, 외부 PG 호출은 DB transaction처럼 rollback되지 않는다. `pgClient.approve`가 성공했는데 뒤의 DB commit이 실패하면 외부 결제는 성공했고 내부 상태는 실패할 수 있다. 반대로 DB에 PENDING을 만들고 PG 호출 중 timeout이 나면, 외부 결제가 성공했는지 실패했는지 모르는 UNKNOWN 상태가 된다.

둘째, `eventPublisher.publish`가 메모리 이벤트나 message broker 직접 발행이라면 DB commit과 원자적으로 묶이지 않을 수 있다. DB commit 전에 발행하면 consumer가 아직 commit되지 않은 상태를 볼 수 있고, DB commit 후 발행하면 commit 직후 process crash로 메시지가 사라질 수 있다.

셋째, 사용자가 같은 요청을 다시 보낼 수 있다. Browser double click, mobile retry, gateway 504, server timeout, worker 재시작이 모두 중복 실행을 만든다. 이때 idempotency key 없이 매번 `payment` row를 만들면 중복 결제나 중복 ledger가 생길 수 있다.

넷째, 금액 계산이 여러 boundary에서 조금씩 다르면 원장이 맞지 않는다. 주문 금액 10,000원, 쿠폰 1,000원, 포인트 333원, 수수료 2.9%, 세금 반올림이 서비스마다 다르게 적용되면 "결제 승인 금액", "ledger debit/credit 합계", "user balance", "정산 금액"이 서로 어긋난다.

작은 모델은 다음처럼 바뀌어야 한다.

```text
request
  |
  | idempotency key + fingerprint 확인
  v
idempotency record(PENDING) 선점
  |
  | local DB transaction 안에서 payment intent / ledger intent / outbox 저장
  v
external call 또는 worker execution
  |
  | 성공, 실패, timeout unknown을 상태로 기록
  v
reconciliation이 최종 사실을 맞춘다
```

여기서 중요한 단어는 "intent"와 "state"다. 결제처럼 외부 side effect가 있는 업무에서는 DB transaction 하나로 세계 전체를 rollback할 수 없다. 대신 내부 DB에는 어떤 일을 하려 했는지, 지금 어디까지 확정됐는지, 어떤 외부 식별자로 다시 조회할 수 있는지, 재시도해도 같은 결과로 수렴하는지를 남긴다.

Connection boundary도 같은 방식으로 본다. 실제 lock 대기와 isolation 증상은 이 boundary 위에서 드러나므로, 증상이 DB 안에서 어떻게 보이는지는 [격리 수준, lock, deadlock](08-isolation-lock-deadlock.md)과 함께 봐야 한다.

```text
HTTP request thread
  -> transaction proxy 진입
  -> transaction manager가 pool에서 connection borrow
  -> connection session에 transaction 시작
  -> JPA/MyBatis SQL 실행
  -> commit 또는 rollback
  -> session state reset 확인
  -> pool에 return
```

이 흐름에서 `@Transactional`이 빠졌거나, self-invocation 때문에 proxy를 지나지 않았거나, 다른 thread로 넘어갔거나, read-only transaction이라고 믿었지만 write SQL이 나갔거나, connection이 반납되지 않았거나, session setting이 남으면 모두 DB 장애처럼 보이는 애플리케이션 경계 문제가 된다.

## 깊은 메커니즘

### Connection은 socket이 아니라 session이다

DB connection을 단순한 네트워크 연결로 보면 pool 문제를 놓친다. Connection에는 DBMS session이 붙어 있고, session에는 여러 상태가 있다. Transaction isolation level, autocommit, current schema, timezone, role, lock timeout, statement timeout, temporary table, prepared statement, user variables 같은 상태가 대표적이다. Connection pool은 이 session을 매번 새로 만들지 않고 재사용한다.

Pool의 장점은 connection 생성 비용을 줄이고 DB 동시 접속 수를 제어한다는 점이다. 하지만 이 장점은 "빌려 간 쪽이 session을 깨끗하게 돌려준다"는 계약 위에 선다. 예를 들어 한 요청이 `SET search_path TO tenant_a`를 실행하고 되돌리지 않은 채 connection을 반환하면, 다음 요청이 tenant_b 데이터를 조회해야 하는데 tenant_a schema를 볼 수 있다. MySQL에서도 `SET time_zone`, user variables, temporary tables, transaction isolation이 예상과 다르게 남으면 장애가 된다.

좋은 pool 설정은 max pool size 숫자 하나가 아니다. Minimum idle, maximum lifetime, idle timeout, connection timeout, validation query, leak detection, max open prepared statement, DB server connection limit을 함께 본다. 애플리케이션 thread 수보다 pool이 너무 작으면 thread가 connection을 기다리며 latency가 커진다. Pool이 너무 크면 DB가 과도한 active query를 처리하다 context switching과 lock contention으로 느려진다. Pool은 "DB를 더 빠르게 만드는 장치"가 아니라 "DB 동시성의 상한을 애플리케이션에서 명확히 거는 장치"이기도 하다.

Transaction 경계에서 connection을 빌리는 시점도 중요하다. Spring은 transaction 시작 시 connection을 즉시 얻을 수도 있고, 실제 SQL이 나갈 때까지 지연될 수도 있다. JPA persistence context가 열려 있어도 DB connection이 항상 물려 있는 것은 아니다. 반대로 long transaction 안에서 외부 API를 호출하면 connection이 오래 점유될 수 있다. HTTP call, file upload, remote service call을 DB transaction 안에 넣으면 pool starvation과 lock 대기, timeout이 함께 발생할 수 있다. 이런 증상은 애플리케이션 코드 문제처럼 시작해도 운영에서는 [운영 관측, 보안, troubleshooting](13-operations-security-troubleshooting.md)의 pool wait, blocker, timeout 지표로 확인된다.

### Spring transaction은 proxy, transaction manager, resource binding의 합성이다

Spring declarative transaction은 보통 proxy 기반으로 동작한다. 외부 caller가 bean proxy를 통해 `@Transactional` method를 호출하면 interceptor가 transaction을 시작하고, method가 정상 종료되면 commit, runtime exception 등 rollback 조건이 맞으면 rollback한다. 같은 class 내부에서 `this.innerTransactionalMethod()`를 호출하면 proxy를 지나지 않아 transaction annotation이 기대대로 적용되지 않을 수 있다.

Transaction manager는 실제 resource 경계를 안다. JDBC/JPA 단일 DB면 `DataSourceTransactionManager` 또는 `JpaTransactionManager`가 connection/entity manager를 thread에 묶는다. 여러 resource를 XA로 묶으려면 JTA transaction manager 같은 다른 경계가 필요하다. `@Transactional`이라는 annotation은 같아 보여도 어떤 transaction manager가 선택됐는지에 따라 실제 commit 대상과 propagation 동작이 달라진다.

Propagation은 "이미 transaction이 있으면 어떻게 할 것인가"를 정한다. `REQUIRED`는 기존 transaction에 참여하고 없으면 새로 만든다. `REQUIRES_NEW`는 기존 transaction을 suspend하고 새 transaction을 만든다. 이 차이는 outbox, audit log, retry log 같은 곳에서 자주 중요하다. 다만 `REQUIRES_NEW`를 남발하면 connection을 추가로 빌려야 하므로 pool size와 deadlock 위험을 함께 봐야 한다.

Rollback rule도 면접 단골 함정이다. Spring의 기본 declarative transaction은 runtime exception과 error에 rollback을 걸고 checked exception은 기본적으로 commit될 수 있다. 물론 설정으로 바꿀 수 있다. 따라서 "exception이 났으니 무조건 rollback"이라고 답하면 위험하다. 또한 transaction method 내부에서 exception을 catch하고 삼켜 버리면 interceptor는 정상 종료로 보고 commit할 수 있다. 비즈니스 실패를 상태로 저장할지 rollback할지는 의도적으로 결정해야 한다.

### JPA와 MyBatis는 SQL 실행 시점과 상태 관리가 다르다

JPA는 persistence context를 둔다. Entity를 조회하면 managed state가 되고, 변경 감지는 flush 시점에 SQL로 변환된다. Transaction commit 전에 flush가 일어날 수 있고, query 실행 전에 flush가 발생할 수도 있다. 그래서 `save`를 호출한 줄과 실제 INSERT가 DB로 나가는 시점이 다를 수 있다. Lazy loading은 transaction/session 경계 밖에서 접근하면 실패하거나, open-session-in-view 같은 설정 때문에 web view 단계까지 persistence context가 살아 남아 예상 못한 query가 나갈 수 있다. 이때 DBMS 엔진이 실제로 어떤 lock과 version을 남기는지는 [MySQL/InnoDB와 PostgreSQL 엔진 비교](11-mysql-postgresql-engine-deep-dive.md)의 저장 구조 차이와 함께 봐야 한다.

MyBatis는 mapper method와 SQL이 더 직접적으로 연결된다. XML 또는 annotation SQL이 호출 시점에 실행되고, result mapping이 객체로 돌아온다. 하지만 MyBatis도 Spring transaction 안에서는 같은 connection에 참여할 수 있다. 같은 service method에서 JPA와 MyBatis를 섞으면 flush timing과 직접 SQL 실행 시점이 어긋날 수 있다. 예를 들어 JPA entity 변경이 아직 flush되지 않았는데 MyBatis select가 같은 row를 읽으면 예상과 다른 값을 볼 수 있다. 이때는 명시적 flush나 경계 분리가 필요하다.

JPA의 optimistic lock과 DB row lock도 구분해야 한다. `@Version`은 update 시 version 조건을 붙여 lost update를 감지하는 애플리케이션-DB 협력 방식이다. `SELECT FOR UPDATE`는 DB row lock을 걸어 다른 transaction의 write를 대기시키는 방식이다. Optimistic lock은 충돌이 드문 경우 throughput이 좋지만 retry와 conflict handling이 필요하다. Pessimistic lock은 충돌을 즉시 대기로 바꾸지만 deadlock과 lock wait timeout을 설계해야 한다.

### Idempotency는 key 저장보다 상태 기계가 중요하다

Idempotency key는 같은 요청을 같은 결과로 수렴시키기 위한 식별자다. 하지만 key만 unique로 두는 것은 충분하지 않다. 같은 key로 다른 amount나 다른 receiver가 들어오면 어떻게 할지, 첫 요청이 timeout 후 PENDING으로 남았을 때 재요청은 기다릴지 조회할지 실패시킬지, 성공 결과를 얼마나 보관할지 정해야 한다.

안전한 idempotency table은 보통 다음 정보를 가진다.

```text
idempotency_key
request_fingerprint
status: PENDING | SUCCEEDED | FAILED | UNKNOWN
business_resource_id
response_snapshot 또는 replay 가능한 결과 참조
external_request_id / external_transaction_id
created_at, updated_at, expires_at
```

처리 흐름은 먼저 `INSERT`로 PENDING record를 만들거나, 이미 존재하면 fingerprint를 비교한다. 같은 key와 같은 fingerprint의 SUCCEEDED면 저장된 결과를 반환한다. 같은 key인데 fingerprint가 다르면 idempotency conflict다. PENDING이면 아직 처리 중인지, lease가 만료됐는지, 외부 조회로 상태를 확인할 수 있는지 판단한다. UNKNOWN이면 reconciliation을 통해 외부 최종 상태를 확인해야 한다.

중요한 점은 PENDING을 외부 호출 뒤에 만들면 늦다는 것이다. 외부 PG 승인 요청이 성공한 뒤 DB insert가 실패하면 다음 retry는 처음 요청을 모른다. 그래서 보통은 local DB transaction에서 idempotency record와 payment intent를 먼저 만들고 commit한 뒤 외부 호출을 하거나, outbox worker가 외부 호출을 담당하게 한다. 어떤 순서든 "외부 side effect를 다시 실행할 수 있는가"와 "이미 실행됐는지 확인할 외부 id가 있는가"를 닫아야 한다.

### Timeout은 실패가 아니라 unknown일 수 있다

외부 API call에서 timeout이 발생하면 application은 응답을 받지 못했다. 하지만 원격 시스템에서는 요청을 처리했을 수 있다. 그래서 timeout을 곧바로 FAILED로 확정하고 재승인을 보내면 중복 결제가 생긴다. 결제, 송금, 포인트 차감, 쿠폰 사용처럼 외부 side effect가 있는 업무에서는 timeout을 UNKNOWN 상태로 두고 외부 transaction id, idempotency key, merchant order id 등으로 조회해야 한다.

UNKNOWN 상태는 불편하지만 정직하다. 업무적으로 "성공인지 실패인지 모른다"를 데이터 모델이 표현해야 reconciliation이 가능하다. Reconciliation은 외부 시스템의 statement, webhook, inquiry API, 내부 ledger를 비교해 최종 상태를 맞추는 과정이다. 좋은 설계는 reconciliation을 예외 운영 수작업으로만 두지 않는다. 상태 전이를 명시하고, 어느 상태에서 어떤 조회를 하며, 외부 결과가 success/failure/absent일 때 내부 상태를 어떻게 닫는지 자동화한다.

### 돈은 decimal type보다 ledger model이 더 중요하다

금액 계산에서 binary floating point를 쓰면 0.1 같은 십진 소수가 정확히 표현되지 않아 누적 오차가 생긴다. Java에서는 `BigDecimal`, DB에서는 `DECIMAL`/`NUMERIC` 같은 십진 고정 정밀 타입을 쓰는 편이 기본이다. 하지만 그것만으로 충분하지 않다. `BigDecimal`도 scale과 rounding mode를 명확히 지정하지 않으면 `divide`에서 예외가 나거나, `equals`와 `compareTo`의 차이로 버그가 생길 수 있다.

돈은 가능하면 minor unit으로도 생각한다. 원화는 1원, 달러는 cent처럼 가장 작은 결제 단위를 정수로 저장하면 일부 계산이 단순해진다. 다만 세금/수수료율 계산이나 다통화 decimal scale이 필요하면 decimal이 필요하다. 어떤 모델을 택하든 핵심은 한 업무 흐름 안에서 rounding boundary가 하나로 고정되는 것이다. Item 단위로 반올림한 합계와 order 전체에서 반올림한 합계는 다를 수 있다. 둘 중 무엇이 계약인지 정하지 않으면 고객 화면, PG 승인 금액, ledger가 어긋난다.

나눗셈과 allocation은 특히 위험하다. 10,000원을 세 사람에게 균등 배분하면 3,333원씩 나누고 1원이 남는다. 이 1원을 누구에게 줄지 규칙이 필요하다. 가장 큰 잔여분 방식, line order 방식, merchant 부담, platform 부담 등 도메인마다 다르다. 중요한 것은 배분 규칙이 deterministic해야 하고, ledger entry 합계가 원 금액과 정확히 같아야 한다는 점이다.

Ledger와 balance도 분리한다. Ledger는 불변 event log에 가깝다. Debit/credit entry가 쌓이고, 각 entry는 왜 생겼는지와 외부 reference를 가진다. Balance는 ledger를 집계한 현재 상태이거나, 빠른 조회를 위해 유지하는 projection이다. Balance를 직접 수정할 수 있게 열어 두면 감사가 어려워진다. 안전한 모델은 ledger를 원천 사실로 두고, balance update를 같은 transaction 안에서 하거나 재계산 가능하게 만드는 것이다.

### Payment state machine은 회계와 외부 세계를 잇는다

결제 상태는 `PENDING`, `AUTHORIZED`, `CAPTURED`, `FAILED`, `CANCELED`, `REFUNDED` 같은 단어를 나열한다고 끝나지 않는다. 각 상태에서 어떤 transition이 허용되는지, transition이 ledger에 어떤 entry를 남기는지, 외부 PG 상태와 내부 상태가 불일치할 때 reconciliation이 어떻게 고치는지가 중요하다.

예를 들어 card authorization과 capture를 분리하는 도메인에서는 승인 성공이 곧 매출 확정이 아니다. `AUTHORIZED` 상태에서는 사용 가능 한도를 잡았지만 아직 정산 대상이 아닐 수 있다. Capture가 성공해야 매출 ledger가 생긴다. Refund는 capture 이후에만 가능하고, partial refund라면 남은 captured amount를 추적해야 한다. 이 상태 모델을 단순 boolean `paid`로 줄이면 부분 환불, 승인 취소, timeout unknown, chargeback을 표현하기 어렵다.

State transition은 DB constraint와 application validation으로 함께 보호한다. 예를 들어 `payment_id + transition_seq` unique key, `external_transaction_id` unique key, amount non-negative constraint, ledger balanced constraint, allowed transition check가 있다. 모든 규칙을 DB check constraint로만 넣을 수는 없지만, 핵심 불변식은 가능한 한 DB에도 둬야 한다. 애플리케이션 버그가 나도 원장이 깨지지 않게 하기 위해서다.

### 2PC/XA, Saga, Outbox는 서로 다른 실패를 줄인다

2PC는 coordinator가 participants에게 prepare를 요청하고, 모두 준비되면 commit을 지시하는 atomic commit protocol이다. XA transaction은 DB와 message broker 같은 XA resource가 이 protocol에 참여할 때 쓸 수 있다. 장점은 여러 resource에 대한 all-or-nothing 보장을 얻는다는 점이다. 단점은 모든 참여자가 XA를 제대로 지원해야 하고, coordinator failure나 prepared transaction 관리, lock 유지, 운영 복구가 어려울 수 있다는 점이다.

Saga는 긴 business transaction을 여러 local transaction으로 나누고, 각 단계가 실패하면 이전 단계를 보상하는 방식이다. 예를 들어 주문 생성, 재고 예약, 결제 승인, 배송 요청을 각각 local transaction으로 처리하고, 배송 요청이 실패하면 결제 취소와 재고 예약 취소를 수행한다. Saga는 분산 lock을 오래 잡지 않지만, 중간 상태가 외부에 보이고 보상 action이 항상 완전한 inverse가 아닐 수 있다. 이미 보낸 이메일이나 이미 사용된 쿠폰처럼 보상이 어려운 side effect는 별도 정책이 필요하다.

Outbox는 local DB transaction 안에서 business row와 outbox row를 함께 저장한다. Worker는 outbox table을 polling하거나 CDC를 통해 읽고 message broker나 외부 API로 발행한다. 발행 성공 후 outbox row를 SENT로 표시한다. Process가 중간에 죽어도 outbox row가 DB에 남아 있으므로 재시도할 수 있다. 하지만 worker가 발행 성공 후 SENT 표시 전에 죽으면 같은 message가 다시 발행될 수 있다. 그래서 consumer idempotency나 inbox table이 필요하다.

`afterCommit` callback은 outbox와 다르다. afterCommit은 DB commit 뒤 어떤 action을 실행하게 해 주지만, commit 직후 process가 죽으면 action intent가 durable하게 남지 않을 수 있다. Outbox는 "나중에 반드시 시도해야 하는 일"을 DB에 저장한다. 면접에서 outbox를 설명할 때 이 차이를 말하면 실제 장애를 이해하고 있다는 신호가 된다.

### Worker claim과 retry는 작은 DB 설계다

Outbox worker가 여러 대라면 같은 row를 동시에 발행하지 않도록 claim 전략이 필요하다. PostgreSQL에서는 `FOR UPDATE SKIP LOCKED`, MySQL 8에서도 유사한 locking read를 활용할 수 있다. Row에는 `status`, `attempt_count`, `next_attempt_at`, `locked_by`, `locked_until`, `last_error` 같은 필드가 들어갈 수 있다. Worker는 due row를 claim하고, 발행하고, 성공/실패 상태를 update한다.

Retry는 무한 반복이 아니다. 일시적 network failure는 exponential backoff로 재시도할 수 있지만, permanent validation error는 dead-letter나 manual review로 보내야 한다. Poison message 하나가 worker batch를 계속 막지 않게 해야 한다. 또한 메시지 payload에는 consumer가 idempotently 처리할 key가 있어야 한다. "producer outbox가 있으니 중복이 없다"가 아니라 "producer는 durable retry를 보장하고, consumer는 duplicate delivery를 견딘다"가 정확한 말이다.

## DBMS별 경계

이 문서의 중심은 애플리케이션 경계지만, DBMS별 차이는 실제 설계에 영향을 준다.

| 경계 | MySQL/InnoDB에서 볼 점 | PostgreSQL에서 볼 점 | 애플리케이션 판단 |
|---|---|---|---|
| transaction isolation | REPEATABLE READ 기본 운영이 많고 gap/next-key lock이 중요하다 | READ COMMITTED 기본, statement snapshot과 row lock/predicate lock 경계가 중요하다 | 같은 `@Transactional`이어도 DB isolation default를 확인한다 |
| duplicate 선점 | `INSERT ... ON DUPLICATE KEY UPDATE`, unique key, lock wait/deadlock 처리 | `INSERT ... ON CONFLICT`, partial unique index, `SKIP LOCKED` 활용 | idempotency key는 DB unique constraint로 강제하고 충돌 결과를 상태 기계로 해석한다 |
| worker claim | locking read와 `SKIP LOCKED` 지원 버전을 확인한다 | `FOR UPDATE SKIP LOCKED`가 outbox queue 구현에 자주 쓰인다 | claim query는 DB별 syntax와 fairness/starvation을 검증한다 |
| decimal | `DECIMAL(M,D)` precision/scale과 strict mode를 확인한다 | `NUMERIC` precision/scale, expression result scale을 확인한다 | DB 타입이 같아 보여도 application rounding boundary를 별도로 고정한다 |
| session state | user variables, time_zone, sql_mode, isolation, temporary table을 조심한다 | search_path, role, timezone, statement_timeout, advisory lock을 조심한다 | pool 반납 시 reset과 connection initialization을 명확히 한다 |
| prepared transaction | XA transaction과 storage engine/resource 지원을 확인한다 | two-phase commit은 prepared transaction 관리와 cleanup이 필요하다 | 2PC는 운영 복구 절차까지 있어야 선택할 수 있다 |
| lock 관측 | performance schema, InnoDB status, processlist를 본다 | `pg_stat_activity`, `pg_locks`, wait_event를 본다 | timeout이 나면 application log와 DB wait graph를 함께 맞춘다 |

Spring/JPA/MyBatis 경계도 DBMS와 독립적이지 않다. PostgreSQL에서 `search_path`를 tenant routing에 쓰면 connection pool contamination이 치명적이다. MySQL에서 transaction isolation을 요청 단위로 바꾼 뒤 되돌리지 않으면 다음 요청의 gap lock 행동이 달라질 수 있다. JPA가 생성한 SQL은 DBMS별 dialect에 따라 lock clause와 pagination 문법이 바뀐다. MyBatis raw SQL은 dialect 차이를 더 직접적으로 드러낸다.

Money 설계에서도 DBMS 경계가 있다. PostgreSQL `NUMERIC`은 arbitrary precision에 가까운 decimal 연산을 제공하지만, 무제한 precision을 방치하면 storage와 연산 비용이 커질 수 있다. MySQL `DECIMAL(M,D)`는 선언한 precision/scale 경계가 명확하다. 어느 쪽이든 application에서는 통화별 minor unit, rounding mode, allocation rule을 DB 타입보다 먼저 계약으로 잡아야 한다.

Outbox 구현에서도 DBMS별 경계가 있다. PostgreSQL은 `FOR UPDATE SKIP LOCKED`와 partial index를 이용해 due row claim을 비교적 자연스럽게 만들 수 있다. MySQL도 버전별 locking read 지원을 확인해야 하고, InnoDB gap lock과 isolation level이 worker concurrency에 어떤 영향을 주는지 실험해야 한다. CDC 기반 outbox를 쓰면 MySQL binlog나 PostgreSQL logical decoding 같은 로그 기반 경계가 들어오며, schema migration과 message ordering을 별도로 검증해야 한다.

## 직접 재생해 보기

아래 실험은 로컬 disposable database와 작은 Spring/JDBC 또는 SQL console에서 수행한다. 핵심은 프레임워크 마법을 외우는 것이 아니라, 어느 경계에서 실제 상태가 바뀌는지 눈으로 보는 것이다.

### Idempotency key 선점 실험

PostgreSQL 예시다. MySQL에서는 `ON CONFLICT` 대신 `ON DUPLICATE KEY` 계열로 바꾼다.

```sql
CREATE TABLE payment_idempotency (
    idempotency_key TEXT PRIMARY KEY,
    request_fingerprint TEXT NOT NULL,
    status TEXT NOT NULL,
    payment_id BIGINT,
    response_body JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO payment_idempotency(idempotency_key, request_fingerprint, status)
VALUES ('pay:order-100:try-1', 'amount=10000;currency=KRW', 'PENDING')
ON CONFLICT (idempotency_key) DO NOTHING;

SELECT * FROM payment_idempotency
WHERE idempotency_key = 'pay:order-100:try-1';
```

PASS 신호는 같은 key의 두 번째 요청이 새 작업을 만들지 않고 기존 row를 보게 되는 것이다. FAIL 신호는 같은 key로 다른 amount가 들어왔는데도 기존 결과를 그대로 replay하는 것이다. 반드시 fingerprint 비교를 추가해야 한다.

### Timeout을 FAILED로 확정하지 않는 상태 전이

```sql
CREATE TABLE payment (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    amount NUMERIC(19, 2) NOT NULL,
    status TEXT NOT NULL,
    external_request_id TEXT NOT NULL UNIQUE,
    external_transaction_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);
```

처리 흐름을 손으로 따라간다.

```text
1. payment row를 PENDING으로 만든다.
2. 외부 승인 요청에 external_request_id를 포함한다.
3. timeout이 나면 FAILED가 아니라 UNKNOWN으로 바꾼다.
4. reconciliation worker가 external_request_id로 외부 상태를 조회한다.
5. 외부 성공이면 CAPTURED 또는 AUTHORIZED로, 외부 실패면 FAILED로 닫는다.
```

PASS 신호는 timeout 후 재요청이 새 external request를 만들지 않고 기존 external_request_id를 조회하는 것이다. FAIL 신호는 timeout을 실패로 단정해 두 번째 승인 요청을 보내는 것이다.

### 금액 배분 불변식 확인

Java 또는 SQL 중 편한 쪽으로 테스트한다. 원칙은 모든 line allocation의 합이 원 금액과 정확히 같아야 한다.

```text
input amount = 10000
receivers = A, B, C

base share = 3333
remainder = 1

rule: 정렬된 receiver 중 앞에서부터 remainder만큼 1원씩 더한다

A = 3334
B = 3333
C = 3333
sum = 10000
```

PASS 신호는 같은 입력을 여러 번 실행해도 같은 배분 결과가 나오고, 합계가 항상 원 금액과 같으며, ledger debit/credit 합계가 0으로 맞는 것이다. FAIL 신호는 반올림을 line마다 다르게 적용해 합계가 9,999원 또는 10,001원이 되는 것이다.

### Spring self-invocation 확인

다음 구조를 작은 프로젝트에서 만든다.

```java
@Service
public class PaymentService {
    public void outer() {
        inner();
    }

    @Transactional
    public void inner() {
        // insert 후 runtime exception 발생
    }
}
```

`outer()`를 외부에서 호출하면 `inner()` 호출이 같은 객체 내부에서 일어나 proxy를 지나지 않을 수 있다. PASS 신호는 transaction log 또는 DB 결과로 rollback이 기대대로 걸리지 않는 상황을 재현하고, 별도 bean으로 분리하거나 self proxy를 거치면 transaction이 적용되는 것을 확인하는 것이다. FAIL 신호는 annotation만 보고 transaction이 무조건 적용됐다고 판단하는 것이다.

### Outbox worker claim 실험

PostgreSQL 예시다.

```sql
CREATE TABLE outbox_event (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'READY',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP NOT NULL DEFAULT now(),
    locked_by TEXT,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

WITH candidates AS (
    SELECT id
    FROM outbox_event
    WHERE status = 'READY'
      AND next_attempt_at <= now()
    ORDER BY id
    FOR UPDATE SKIP LOCKED
    LIMIT 10
)
UPDATE outbox_event o
SET status = 'PROCESSING',
    locked_by = 'worker-1',
    locked_until = now() + interval '30 seconds',
    updated_at = now()
FROM candidates c
WHERE o.id = c.id
RETURNING o.*;
```

두 worker session에서 동시에 실행해도 같은 row가 중복 claim되지 않는지 본다. PASS 신호는 각 worker가 다른 row를 잡는 것이다. FAIL 신호는 발행 성공 후 상태 update 전에 worker가 죽었을 때 row가 영원히 PROCESSING으로 남는 것이다. 이 실패를 막으려면 `locked_until` 만료 후 재claim하거나 recovery job을 둔다.

## 면접 꼬리 질문

### `@Transactional`이 붙어 있으면 모든 DB 작업이 하나로 묶이나요?

아니다. 호출이 Spring proxy를 통과해야 하고, 올바른 transaction manager가 선택되어야 하며, 같은 thread/resource binding 안에서 DB 작업이 실행되어야 한다. Self-invocation, private method, final method, 다른 thread 실행, transaction manager mismatch, exception swallowing은 기대한 transaction 경계를 깨뜨릴 수 있다. JPA flush timing과 MyBatis 직접 SQL 실행 시점도 함께 봐야 한다.

### Connection pool이 커지면 성능이 항상 좋아지나요?

아니다. Pool은 DB 동시성을 제한하는 장치다. 너무 작으면 애플리케이션 thread가 connection을 기다리고, 너무 크면 DB가 과도한 active query와 lock contention을 감당해야 한다. 적정 값은 DB CPU, I/O, query latency, transaction duration, 애플리케이션 concurrency, timeout 정책으로 정한다. Pool exhaustion이 보이면 max pool size를 올리기 전에 long transaction, connection leak, 외부 call inside transaction을 먼저 확인해야 한다.

### Idempotency key만 unique로 걸면 중복 결제가 막히나요?

부분적으로만 막힌다. 같은 key로 같은 요청이 왔는지 fingerprint를 확인해야 하고, PENDING/UNKNOWN 상태를 처리해야 하며, 외부 side effect가 이미 실행됐는지 조회할 수 있어야 한다. Unique key는 선점 장치이고, idempotency는 상태 기계와 replay 정책이다.

### Timeout이 나면 transaction rollback으로 충분하지 않나요?

DB 변경만 rollback할 수 있다. 외부 PG 승인 요청, message 발행, 이메일 전송 같은 side effect는 DB rollback으로 되돌아가지 않는다. Timeout은 실패가 아니라 unknown일 수 있으므로, 외부 id로 조회하고 reconciliation으로 최종 상태를 닫아야 한다.

### Outbox를 쓰면 메시지가 정확히 한 번만 발행되나요?

아니다. Outbox는 local DB transaction과 durable message intent를 묶어 "발행해야 할 일이 사라지지 않게" 해 준다. Worker crash나 broker ack 경계 때문에 같은 메시지가 두 번 발행될 수 있다. Consumer idempotency, inbox table, message key, deduplication이 함께 필요하다.

### 돈 계산에서 `BigDecimal`만 쓰면 충분한가요?

아니다. `BigDecimal`은 binary floating point 문제를 피하는 도구일 뿐이다. Scale, rounding mode, allocation rule, 통화별 minor unit, DB decimal precision, ledger/balance 불변식이 함께 필요하다. 특히 나눗셈과 수수료 계산은 어느 boundary에서 반올림하는지 계약으로 고정해야 한다.

### Saga와 2PC 중 무엇이 더 좋은가요?

목적이 다르다. 2PC/XA는 참여 resource가 prepare/commit protocol을 지원하고 운영 복구를 감당할 수 있을 때 atomic commit을 제공한다. Saga는 긴 업무 흐름을 local transaction과 보상 action으로 나누어 분산 lock을 피하지만, 중간 상태와 보상 실패를 설계해야 한다. 더 좋은 것 하나가 아니라, resource 지원 범위와 실패 비용에 맞는 선택이 필요하다.

## 함정 질문

### "외부 API 호출도 transaction 안에 넣으면 안전해지죠?"

오히려 위험해질 수 있다. DB transaction 안에서 외부 API를 호출하면 connection과 lock을 오래 잡고, 외부 API 성공 후 DB rollback 같은 불일치가 생긴다. 외부 side effect는 DB transaction으로 rollback되지 않는다. 안전한 설계는 local intent를 먼저 durable하게 남기고, 외부 호출 결과를 상태와 reconciliation으로 닫는 것이다.

### "Read-only transaction이면 DB에 아무 영향이 없나요?"

항상 그렇지 않다. Read-only hint는 DBMS와 transaction manager, ORM에 따라 강제 수준이 다르다. 또한 long read-only transaction도 snapshot을 오래 유지해 PostgreSQL vacuum이나 InnoDB undo purge에 영향을 줄 수 있다. "쓰기 SQL이 없으니 운영 비용도 없다"는 말은 위험하다.

### "JPA repository save를 호출했으니 DB에 바로 들어갔겠죠?"

JPA에서는 persistence context와 flush timing이 있다. `save`가 managed entity를 만들거나 변경해도 실제 SQL은 flush/commit 시점에 나갈 수 있다. Query 실행 전에 flush가 일어날 수도 있다. MyBatis처럼 mapper 호출과 SQL 실행이 직접 연결된 모델과 섞을 때는 특히 조심해야 한다.

### "중복 요청은 client에서 버튼 disabled 처리하면 끝 아닌가요?"

아니다. Client 제어는 UX 개선일 뿐이다. Network retry, gateway timeout, server retry, queue redelivery, worker crash는 client button과 무관하게 중복을 만든다. Idempotency는 server-side durable constraint와 상태 기계로 구현해야 한다.

### "Ledger가 있으면 balance table은 필요 없나요?"

조회 성능 때문에 balance projection이 필요할 수 있다. 다만 balance를 원천 사실처럼 직접 수정하면 감사가 어렵다. 안전한 모델은 ledger를 authoritative record로 두고, balance는 같은 transaction에서 갱신하거나 재계산 가능한 projection으로 둔다. Ledger 합계와 balance가 어긋나면 reconciliation 대상이다.

### "2PC를 쓰면 분산 시스템 문제가 사라지나요?"

아니다. 2PC는 atomic commit 문제를 줄여 주지만, coordinator failure, prepared transaction cleanup, participant blocking, timeout, 운영 복구라는 다른 비용을 만든다. 모든 resource가 XA를 잘 지원해야 하고, prepared 상태가 오래 남으면 lock과 resource를 잡을 수 있다. 2PC 선택은 프로토콜뿐 아니라 운영 절차 선택이다.

### 전이 질문: DB transaction 밖으로 새는 실패

다른 업무에서도 외부 side effect가 있으면 같은 구조가 반복됩니다. 이메일, 배송 요청, 포인트 적립, 쿠폰 발급은 DB rollback으로 자동 취소되지 않습니다. 전이 질문은 `중복 요청이 들어오면 무엇을 새로 하지 말아야 하는가`, `응답을 잃으면 무엇을 조회해야 하는가`, `최종 증거는 어디에 남는가`입니다.

## 더 깊게 볼 자료

- [Spring Framework Reference - Declarative Transaction Management](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html): `@Transactional`, proxy, rollback rule, transaction manager 경계를 확인한다.
- [Spring Framework Reference - Transaction Propagation](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/tx-propagation.html): `REQUIRED`, `REQUIRES_NEW`, nested transaction의 실제 의미를 확인한다.
- [MyBatis-Spring Reference - Transactions](https://mybatis.org/spring/transactions.html): MyBatis mapper가 Spring transaction에 참여하는 방식을 확인한다.
- [PostgreSQL Current Documentation - Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html): PostgreSQL isolation level과 snapshot 경계를 확인한다.
- [PostgreSQL Current Documentation - Two-Phase Commit](https://www.postgresql.org/docs/current/two-phase.html): prepared transaction과 운영 주의점을 확인한다.
- [MySQL 8.4 Reference Manual - XA Transactions](https://dev.mysql.com/doc/refman/8.4/en/xa.html): MySQL XA transaction의 지원 범위와 경계를 확인한다.
- [PostgreSQL Current Documentation - Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html): `FOR UPDATE`, row lock, table lock, advisory lock을 확인한다.
- [MySQL 8.4 Reference Manual - InnoDB Locking](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html): InnoDB lock이 idempotency 선점과 worker claim에 미치는 영향을 확인한다.
