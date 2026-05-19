# Outbox, Saga, 2PC

## 2PC/XA/JTA vs saga/outbox

2PC/XA/JTA, saga, outbox는 모두 '여러 부작용을 어떻게 일관되게 다룰 것인가'라는 질문에서 나온다. 이 패턴들이 등장한 배경은 단순하다. 하나의 DB 안에서는 commit과 rollback이 비교적 선명하지만, DB 밖의 broker, HTTP API, 다른 서비스, 결제 승인망이 함께 움직이는 순간 한 트랜잭션 경계로 모든 결과를 묶기 어려워진다. 그러나 해결하려는 실패 모델은 서로 다르다. 2PC/XA는 여러 XA-aware resource manager가 하나의 global transaction에 참여할 수 있을 때 쓰는 강한 원자성 모델이다. saga는 긴 업무 흐름을 여러 local transaction과 보상 동작으로 나눈다. outbox는 로컬 DB commit과 외부 메시지 publish 사이의 틈을 durable row로 메운다. 이 절의 핵심은 DB commit과 메시지 publish 사이의 틈을 무시하지 않는 것이다.

이 절에서 공식 근거로 삼은 자료는 다음과 같다. 링크는 단순 참고가 아니라 본문 판단의 경계다. PostgreSQL과 MySQL은 같은 단어를 쓰더라도 로그 이름, 보존 책임, failover 절차, 관측 뷰가 다르므로, 공통 원리와 제품별 실행 방법을 분리해 읽어야 한다.

- Spring Framework Transaction Management: https://docs.spring.io/spring-framework/reference/data-access/transaction.html
- Spring Transaction-bound Events: https://docs.spring.io/spring-framework/reference/data-access/transaction/event.html
- Jakarta Transactions: https://jakarta.ee/specifications/transactions/
- Jakarta Transactions 2.0 Specification: https://jakarta.ee/specifications/transactions/2.0/jakarta-transactions-spec-2.0
- 로컬 seed: knowledge/cards/K-ASYNC-EVENTS-OUTBOX-INBOX-CLAIM-LEASE.md

가장 작은 실패 trace부터 보자. 주문 service가 DB에 주문을 저장하고 Kafka, RabbitMQ, webhook, HTTP API 같은 외부 시스템으로 `OrderCreated`를 보내야 한다고 하자. DB insert와 message publish를 코드에서 연속으로 호출하면 두 줄 사이에 프로세스 crash, network timeout, broker ack 손실, transaction rollback이 들어올 수 있다. 이 틈은 테스트에서 잘 보이지 않지만 운영에서는 배포, OOM, broker 장애, timeout retry와 함께 반복된다.

```text
unsafe dual side effect

BEGIN local DB transaction
  INSERT INTO orders(id=100, status='CREATED')
COMMIT
  process crashes here before publish
PUBLISH OrderCreated(100)

observed state:
  orders row exists
  no message exists
  downstream inventory/search/notification never learns the order
```

반대로 publish가 먼저 성공하고 DB commit이 실패하면 외부 시스템은 존재하지 않는 주문을 보게 된다. 이 두 실패는 서로 반대 방향이지만 원인은 같다. DB와 broker 또는 HTTP endpoint가 같은 원자적 commit 경계에 있지 않다. outbox는 이 문제를 '외부 publish를 DB transaction 안에 넣는다'로 풀지 않는다. 대신 publish해야 할 의도를 outbox row로 같은 DB transaction에 저장하고, 별도 worker가 그 row를 claim해 publish한다. commit이 성공하면 보낼 의도도 남고, commit이 rollback되면 outbox row도 사라진다.

```sql
BEGIN;

INSERT INTO orders(order_id, user_id, status, amount)
VALUES (100, 42, 'CREATED', 39000);

INSERT INTO outbox_events(event_id, aggregate_type, aggregate_id, event_type,
                          payload_json, status, attempt_count, next_attempt_at, created_at)
VALUES ('evt-100', 'ORDER', '100', 'OrderCreated',
        '{"orderId":100,"userId":42,"amount":39000}',
        'PENDING', 0, now(), now());

COMMIT;

-- worker side
UPDATE outbox_events
   SET status = 'DELIVERING', claimed_by = 'worker-1', lease_until = now() + interval '30 seconds'
 WHERE event_id = (
   SELECT event_id FROM outbox_events
    WHERE status IN ('PENDING', 'RETRYING')
      AND next_attempt_at <= now()
    ORDER BY created_at
    LIMIT 1
    FOR UPDATE SKIP LOCKED
 )
RETURNING *;
```

로컬 seed 카드가 강조한 claim lease가 여기서 중요하다. outbox row가 있다고 해서 안전한 것이 아니다. worker 두 개가 동시에 `PENDING` row를 읽고 둘 다 publish하면 중복 메시지가 나간다. 따라서 저장소 primitive는 load가 아니라 claim이어야 한다. claim은 조건부 update, row lock, lease_until, attempt_count를 통해 '이 worker가 일정 시간 처리권을 가진다'는 상태 전이를 원자적으로 만든다. publish 성공 후 `SENT`, 실패 후 backoff를 계산해 `RETRYING`으로 바꾸면 crash 후에도 다시 이어 갈 수 있다.

| 패턴 | 강점 | 깨지는 지점 | 적합한 상황 |
| --- | --- | --- | --- |
| 2PC/XA/JTA | 여러 XA resource를 global transaction으로 묶음 | coordinator recovery, in-doubt, XA 미지원 외부 API | DB와 JMS처럼 XA 참여자가 명확한 enterprise 환경 |
| outbox | 로컬 transaction과 publish intent를 함께 저장 | 중복 전달은 소비자 idempotency 필요 | DB commit 후 메시지 유실 방지 |
| saga | 긴 업무 흐름을 단계와 보상으로 명시 | 보상은 완전한 undo가 아님 | 여러 서비스가 각자 local transaction을 가질 때 |
| 단순 afterCommit publish | 구현이 작고 빠름 | publish 실패가 durable retry로 남지 않음 | 유실 허용 이벤트나 별도 보정이 있는 경우 |

Jakarta Transactions 공식 specification은 transaction manager, resource manager, application server, application program 같은 참여자를 나누고, XAResource를 통해 resource enlistment와 two-phase commit을 수행할 수 있게 한다. 특히 Jakarta Messaging provider가 XAResource를 지원하면 two-phase commit transaction protocol에 참여할 수 있다고 설명한다. 이 모델은 강하지만 전제가 있다. 참여자가 XA-aware여야 하고, transaction manager가 recovery log와 in-doubt 상태를 운영해야 하며, 외부 HTTP API나 대부분의 SaaS webhook은 이 protocol에 참여하지 않는다.

Spring Framework는 transaction abstraction을 제공해 JTA, JDBC, Hibernate, JPA 같은 transaction API 위에서 일관된 programming model을 제공한다. 또한 `@TransactionalEventListener`는 listener를 `BEFORE_COMMIT`, `AFTER_COMMIT`, `AFTER_ROLLBACK`, `AFTER_COMPLETION` 같은 phase에 묶을 수 있다. 그러나 phase를 묶는 것은 durable delivery와 다르다. `AFTER_COMMIT` listener가 실행되다가 process가 죽거나 broker publish가 실패하면, 이미 끝난 DB transaction을 되돌릴 수 없고 실패 event가 자동으로 재시도 queue에 남지도 않는다. 그래서 중요한 외부 전달은 listener callback만으로 닫지 않고 outbox row로 남긴다.

```text
transaction-bound event vs outbox

@TransactionalEventListener(AFTER_COMMIT)
  order committed -> callback runs -> publish now
  failure after commit -> no automatic durable retry unless code stores one elsewhere

outbox
  order row + outbox row committed together
  worker publishes later
  failure after commit -> row remains PENDING/RETRYING and can be retried

conclusion
  transaction phase controls timing.
  outbox controls durable intent and retry state.
```

saga는 outbox와 경쟁하는 말이 아니라 함께 쓰이는 경우가 많다. 예를 들어 주문 생성, 재고 예약, 결제 승인, 배송 요청이 각각 다른 서비스라면 하나의 local DB transaction으로 묶을 수 없다. 주문 서비스는 `OrderCreated` outbox를 남기고, 재고 서비스는 inbox로 중복 수신을 막은 뒤 재고 예약 local transaction과 `StockReserved` outbox를 남긴다. 결제 실패가 오면 saga state는 보상 단계로 이동한다. 여기서 outbox/inbox는 각 local step의 event 전달 신뢰성을 만들고, saga는 여러 step의 업무 상태 전이를 설명한다.

| 시간 | 서비스 | local transaction | outbox/inbox 상태 | 외부 관측 |
| --- | --- | --- | --- | --- |
| T1 | Order | orders=CREATED, outbox OrderCreated=PENDING | OrderCreated worker 대기 | 사용자는 주문 접수 |
| T2 | Inventory | inbox OrderCreated 처리, reservation=HELD, outbox StockReserved=PENDING | dedup_key 기록 | 재고 임시 차감 |
| T3 | Payment | payment authorization 실패, outbox PaymentFailed=PENDING | 실패 event 저장 | 결제 실패 |
| T4 | Order saga | PaymentFailed inbox 처리, order=CANCEL_REQUESTED, outbox ReleaseStock=PENDING | 보상 시작 | 주문 취소 진행 |
| T5 | Inventory | ReleaseStock 처리, reservation=RELEASED | idempotent release | 재고 복원 |

2PC를 무조건 피해야 한다는 뜻은 아니다. 같은 application server 안에서 XA-capable DB와 JMS provider를 묶고, transaction manager 운영 경험이 있으며, latency와 lock 보유 시간을 감당할 수 있다면 2PC가 맞을 수 있다. 그러나 microservice, HTTP callback, SaaS API, cloud managed broker, polyglot service가 섞인 환경에서는 모든 참여자를 XA resource로 만들기 어렵다. 이때 2PC를 흉내 내려고 distributed lock이나 임시 status를 억지로 만들면, recovery protocol 없는 가짜 2PC가 된다.

outbox도 한계가 있다. 보통 at-least-once delivery를 제공하므로 같은 event가 두 번 publish될 수 있다. worker가 publish 성공 후 `SENT`로 표시하기 전에 죽으면, 재시작 후 같은 row를 다시 publish할 수 있다. 따라서 consumer는 idempotency key나 inbox table로 중복을 흡수해야 한다. 이 중복 가능성을 문서에 쓰지 않으면 outbox를 exactly-once 장치로 오해하게 된다.

```sql
-- consumer inbox idempotency sketch
BEGIN;

INSERT INTO inbox_events(consumer_name, event_id, received_at)
VALUES ('inventory-service', 'evt-100', now())
ON CONFLICT (consumer_name, event_id) DO NOTHING;

-- affected rows = 1이면 처음 보는 이벤트이므로 업무 반영
UPDATE stock
   SET reserved = reserved + 1
 WHERE sku = 'SKU-1';

COMMIT;

-- affected rows = 0이면 이미 처리한 이벤트이므로 업무 update를 건너뛴다.
```

관측 지표는 설계 의도를 따라가야 한다. outbox에서는 pending row 수, 가장 오래된 pending age, retrying row 수, attempt_count 분포, lease 만료 횟수, publish latency, last_error 상위 원인이 필요하다. saga에서는 saga instance 상태별 개수, 오래 머문 상태, 보상 성공/실패, step timeout이 필요하다. XA/2PC에서는 in-doubt transaction, prepare 후 대기 시간, recovery scan 결과가 필요하다. 한 패턴의 지표를 다른 패턴에 붙이면 장애를 놓친다.

검증은 failure injection으로 닫는다. 첫째, DB commit 직후 process를 죽였을 때 outbox row가 남아 worker 재시작 후 publish되는지 본다. 둘째, publish 성공 직후 `SENT` update 전에 worker를 죽였을 때 consumer inbox가 중복을 흡수하는지 본다. 셋째, 두 worker를 동시에 돌려 같은 row를 claim하려 할 때 하나만 성공하는지 본다. 넷째, 외부 API가 500을 반환할 때 backoff와 max attempt가 의도대로 움직이는지 본다. PASS는 row 상태와 외부 관측이 trace로 설명되는 것이고, FAIL은 로그에는 에러가 있지만 어떤 event가 유실/중복/재시도 중인지 SQL로 설명하지 못하는 것이다.

### 운영 리허설 카드: commit과 publish 사이의 틈을 다루는 법

1. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.

2. 2PC/XA는 여러 resource manager가 한 global transaction에 참여하도록 prepare와 commit을 나누는 강한 모델이다. Jakarta Transactions는 transaction manager, resource manager, application server, XAResource 같은 역할을 정의한다. 하지만 외부 HTTP API나 대부분의 SaaS webhook 대상은 XA resource가 아니므로, 모든 부작용을 2PC 안으로 넣을 수 있다고 가정하면 설계가 깨진다.

3. Spring의 @TransactionalEventListener는 transaction phase에 event listener를 묶을 수 있지만, AFTER_COMMIT listener에서 외부 publish가 실패했을 때 자동으로 DB transaction을 되돌려 주지는 않는다. listener phase는 callback 시점을 조절하는 도구이고, durable retry queue를 대신하지 않는다. outbox row가 필요한 이유가 여기에 있다.

4. saga는 긴 업무 절차를 여러 local transaction과 보상 작업으로 나누는 사고방식이다. 결제 승인, 재고 예약, 배송 요청처럼 각 단계가 별도 시스템이면 하나의 ACID transaction으로 묶기 어렵다. saga는 실패를 없애지 않고, 실패했을 때 어느 단계까지 갔고 어떤 보상 또는 재시도가 필요한지 상태를 드러낸다.

5. outbox worker는 loadPending 후 send가 아니라 claim 후 send여야 한다. 여러 worker가 같은 pending row를 읽고 동시에 publish하면 중복 메시지가 나간다. 조건부 UPDATE, SKIP LOCKED, lease_until 같은 primitive로 처리권을 먼저 선출하고, publish 결과에 따라 SENT 또는 RETRYING으로 바꾸는 흐름이 필요하다.

6. exactly-once publish는 대부분 애플리케이션 경계에서 환상에 가깝다. outbox는 보통 at-least-once 전달을 만들고, 소비자는 inbox 또는 idempotency key로 중복을 흡수한다. 중복을 없애겠다는 목표보다 중복이 와도 업무 결과가 한 번만 반영되게 만드는 목표가 더 검증 가능하다.

7. outbox table을 업무 table에 섞으면 편해 보이지만 worker의 status, attempt_count, next_attempt_at update가 업무 조회와 같은 index를 흔든다. outbox는 업무 commit과 같은 transaction에 참여하되, 운영 패턴은 별도 queue에 가깝다. 그래서 row envelope를 고정한 별도 table이 관측과 성능 면에서 보통 더 안전하다.

8. 2PC는 coordinator recovery log가 핵심이다. prepare 이후 coordinator가 죽으면 participant는 in-doubt 상태에 남을 수 있고, recovery가 commit 또는 rollback 결정을 다시 알려줘야 한다. 이 비용을 이해하지 못하면 'XA를 켜면 원자성이 해결된다'고 생각하지만, 실제로는 transaction manager 운영과 장애 복구 책임이 새로 생긴다.

9. saga 보상은 undo가 아니다. 이미 외부에 보낸 이메일을 지울 수 없고, 이미 승인된 결제를 취소하더라도 수수료나 고객 경험은 남는다. 보상은 업무적으로 반대 효과를 내는 새 action일 뿐이다. 그래서 saga 설계는 기술 흐름보다 고객에게 보이는 중간 상태와 재시도 정책을 먼저 정해야 한다.

10. outbox 관측은 pending 개수만 보면 부족하다. oldest_pending_age, retry_backoff 분포, last_error 상위 원인, claim lease 만료 수, publish latency, consumer idempotency hit rate가 있어야 한다. 메시지가 안 나간 장애와 중복이 많이 나간 장애는 같은 outbox table에서 다른 지표로 보인다.

11. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.

12. 2PC/XA는 여러 resource manager가 한 global transaction에 참여하도록 prepare와 commit을 나누는 강한 모델이다. Jakarta Transactions는 transaction manager, resource manager, application server, XAResource 같은 역할을 정의한다. 하지만 외부 HTTP API나 대부분의 SaaS webhook 대상은 XA resource가 아니므로, 모든 부작용을 2PC 안으로 넣을 수 있다고 가정하면 설계가 깨진다.

13. Spring의 @TransactionalEventListener는 transaction phase에 event listener를 묶을 수 있지만, AFTER_COMMIT listener에서 외부 publish가 실패했을 때 자동으로 DB transaction을 되돌려 주지는 않는다. listener phase는 callback 시점을 조절하는 도구이고, durable retry queue를 대신하지 않는다. outbox row가 필요한 이유가 여기에 있다.

14. saga는 긴 업무 절차를 여러 local transaction과 보상 작업으로 나누는 사고방식이다. 결제 승인, 재고 예약, 배송 요청처럼 각 단계가 별도 시스템이면 하나의 ACID transaction으로 묶기 어렵다. saga는 실패를 없애지 않고, 실패했을 때 어느 단계까지 갔고 어떤 보상 또는 재시도가 필요한지 상태를 드러낸다.

15. outbox worker는 loadPending 후 send가 아니라 claim 후 send여야 한다. 여러 worker가 같은 pending row를 읽고 동시에 publish하면 중복 메시지가 나간다. 조건부 UPDATE, SKIP LOCKED, lease_until 같은 primitive로 처리권을 먼저 선출하고, publish 결과에 따라 SENT 또는 RETRYING으로 바꾸는 흐름이 필요하다.

16. exactly-once publish는 대부분 애플리케이션 경계에서 환상에 가깝다. outbox는 보통 at-least-once 전달을 만들고, 소비자는 inbox 또는 idempotency key로 중복을 흡수한다. 중복을 없애겠다는 목표보다 중복이 와도 업무 결과가 한 번만 반영되게 만드는 목표가 더 검증 가능하다.

17. outbox table을 업무 table에 섞으면 편해 보이지만 worker의 status, attempt_count, next_attempt_at update가 업무 조회와 같은 index를 흔든다. outbox는 업무 commit과 같은 transaction에 참여하되, 운영 패턴은 별도 queue에 가깝다. 그래서 row envelope를 고정한 별도 table이 관측과 성능 면에서 보통 더 안전하다.

18. 2PC는 coordinator recovery log가 핵심이다. prepare 이후 coordinator가 죽으면 participant는 in-doubt 상태에 남을 수 있고, recovery가 commit 또는 rollback 결정을 다시 알려줘야 한다. 이 비용을 이해하지 못하면 'XA를 켜면 원자성이 해결된다'고 생각하지만, 실제로는 transaction manager 운영과 장애 복구 책임이 새로 생긴다.

19. saga 보상은 undo가 아니다. 이미 외부에 보낸 이메일을 지울 수 없고, 이미 승인된 결제를 취소하더라도 수수료나 고객 경험은 남는다. 보상은 업무적으로 반대 효과를 내는 새 action일 뿐이다. 그래서 saga 설계는 기술 흐름보다 고객에게 보이는 중간 상태와 재시도 정책을 먼저 정해야 한다.

20. outbox 관측은 pending 개수만 보면 부족하다. oldest_pending_age, retry_backoff 분포, last_error 상위 원인, claim lease 만료 수, publish latency, consumer idempotency hit rate가 있어야 한다. 메시지가 안 나간 장애와 중복이 많이 나간 장애는 같은 outbox table에서 다른 지표로 보인다.

21. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.

22. 2PC/XA는 여러 resource manager가 한 global transaction에 참여하도록 prepare와 commit을 나누는 강한 모델이다. Jakarta Transactions는 transaction manager, resource manager, application server, XAResource 같은 역할을 정의한다. 하지만 외부 HTTP API나 대부분의 SaaS webhook 대상은 XA resource가 아니므로, 모든 부작용을 2PC 안으로 넣을 수 있다고 가정하면 설계가 깨진다.

23. Spring의 @TransactionalEventListener는 transaction phase에 event listener를 묶을 수 있지만, AFTER_COMMIT listener에서 외부 publish가 실패했을 때 자동으로 DB transaction을 되돌려 주지는 않는다. listener phase는 callback 시점을 조절하는 도구이고, durable retry queue를 대신하지 않는다. outbox row가 필요한 이유가 여기에 있다.

24. saga는 긴 업무 절차를 여러 local transaction과 보상 작업으로 나누는 사고방식이다. 결제 승인, 재고 예약, 배송 요청처럼 각 단계가 별도 시스템이면 하나의 ACID transaction으로 묶기 어렵다. saga는 실패를 없애지 않고, 실패했을 때 어느 단계까지 갔고 어떤 보상 또는 재시도가 필요한지 상태를 드러낸다.

25. outbox worker는 loadPending 후 send가 아니라 claim 후 send여야 한다. 여러 worker가 같은 pending row를 읽고 동시에 publish하면 중복 메시지가 나간다. 조건부 UPDATE, SKIP LOCKED, lease_until 같은 primitive로 처리권을 먼저 선출하고, publish 결과에 따라 SENT 또는 RETRYING으로 바꾸는 흐름이 필요하다.

26. exactly-once publish는 대부분 애플리케이션 경계에서 환상에 가깝다. outbox는 보통 at-least-once 전달을 만들고, 소비자는 inbox 또는 idempotency key로 중복을 흡수한다. 중복을 없애겠다는 목표보다 중복이 와도 업무 결과가 한 번만 반영되게 만드는 목표가 더 검증 가능하다.

27. outbox table을 업무 table에 섞으면 편해 보이지만 worker의 status, attempt_count, next_attempt_at update가 업무 조회와 같은 index를 흔든다. outbox는 업무 commit과 같은 transaction에 참여하되, 운영 패턴은 별도 queue에 가깝다. 그래서 row envelope를 고정한 별도 table이 관측과 성능 면에서 보통 더 안전하다.

28. 2PC는 coordinator recovery log가 핵심이다. prepare 이후 coordinator가 죽으면 participant는 in-doubt 상태에 남을 수 있고, recovery가 commit 또는 rollback 결정을 다시 알려줘야 한다. 이 비용을 이해하지 못하면 'XA를 켜면 원자성이 해결된다'고 생각하지만, 실제로는 transaction manager 운영과 장애 복구 책임이 새로 생긴다.

29. saga 보상은 undo가 아니다. 이미 외부에 보낸 이메일을 지울 수 없고, 이미 승인된 결제를 취소하더라도 수수료나 고객 경험은 남는다. 보상은 업무적으로 반대 효과를 내는 새 action일 뿐이다. 그래서 saga 설계는 기술 흐름보다 고객에게 보이는 중간 상태와 재시도 정책을 먼저 정해야 한다.

30. outbox 관측은 pending 개수만 보면 부족하다. oldest_pending_age, retry_backoff 분포, last_error 상위 원인, claim lease 만료 수, publish latency, consumer idempotency hit rate가 있어야 한다. 메시지가 안 나간 장애와 중복이 많이 나간 장애는 같은 outbox table에서 다른 지표로 보인다.

31. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.

32. 2PC/XA는 여러 resource manager가 한 global transaction에 참여하도록 prepare와 commit을 나누는 강한 모델이다. Jakarta Transactions는 transaction manager, resource manager, application server, XAResource 같은 역할을 정의한다. 하지만 외부 HTTP API나 대부분의 SaaS webhook 대상은 XA resource가 아니므로, 모든 부작용을 2PC 안으로 넣을 수 있다고 가정하면 설계가 깨진다.

33. Spring의 @TransactionalEventListener는 transaction phase에 event listener를 묶을 수 있지만, AFTER_COMMIT listener에서 외부 publish가 실패했을 때 자동으로 DB transaction을 되돌려 주지는 않는다. listener phase는 callback 시점을 조절하는 도구이고, durable retry queue를 대신하지 않는다. outbox row가 필요한 이유가 여기에 있다.

34. saga는 긴 업무 절차를 여러 local transaction과 보상 작업으로 나누는 사고방식이다. 결제 승인, 재고 예약, 배송 요청처럼 각 단계가 별도 시스템이면 하나의 ACID transaction으로 묶기 어렵다. saga는 실패를 없애지 않고, 실패했을 때 어느 단계까지 갔고 어떤 보상 또는 재시도가 필요한지 상태를 드러낸다.

35. outbox worker는 loadPending 후 send가 아니라 claim 후 send여야 한다. 여러 worker가 같은 pending row를 읽고 동시에 publish하면 중복 메시지가 나간다. 조건부 UPDATE, SKIP LOCKED, lease_until 같은 primitive로 처리권을 먼저 선출하고, publish 결과에 따라 SENT 또는 RETRYING으로 바꾸는 흐름이 필요하다.

36. exactly-once publish는 대부분 애플리케이션 경계에서 환상에 가깝다. outbox는 보통 at-least-once 전달을 만들고, 소비자는 inbox 또는 idempotency key로 중복을 흡수한다. 중복을 없애겠다는 목표보다 중복이 와도 업무 결과가 한 번만 반영되게 만드는 목표가 더 검증 가능하다.

37. outbox table을 업무 table에 섞으면 편해 보이지만 worker의 status, attempt_count, next_attempt_at update가 업무 조회와 같은 index를 흔든다. outbox는 업무 commit과 같은 transaction에 참여하되, 운영 패턴은 별도 queue에 가깝다. 그래서 row envelope를 고정한 별도 table이 관측과 성능 면에서 보통 더 안전하다.

38. 2PC는 coordinator recovery log가 핵심이다. prepare 이후 coordinator가 죽으면 participant는 in-doubt 상태에 남을 수 있고, recovery가 commit 또는 rollback 결정을 다시 알려줘야 한다. 이 비용을 이해하지 못하면 'XA를 켜면 원자성이 해결된다'고 생각하지만, 실제로는 transaction manager 운영과 장애 복구 책임이 새로 생긴다.

39. saga 보상은 undo가 아니다. 이미 외부에 보낸 이메일을 지울 수 없고, 이미 승인된 결제를 취소하더라도 수수료나 고객 경험은 남는다. 보상은 업무적으로 반대 효과를 내는 새 action일 뿐이다. 그래서 saga 설계는 기술 흐름보다 고객에게 보이는 중간 상태와 재시도 정책을 먼저 정해야 한다.

40. outbox 관측은 pending 개수만 보면 부족하다. oldest_pending_age, retry_backoff 분포, last_error 상위 원인, claim lease 만료 수, publish latency, consumer idempotency hit rate가 있어야 한다. 메시지가 안 나간 장애와 중복이 많이 나간 장애는 같은 outbox table에서 다른 지표로 보인다.

41. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.

42. 2PC/XA는 여러 resource manager가 한 global transaction에 참여하도록 prepare와 commit을 나누는 강한 모델이다. Jakarta Transactions는 transaction manager, resource manager, application server, XAResource 같은 역할을 정의한다. 하지만 외부 HTTP API나 대부분의 SaaS webhook 대상은 XA resource가 아니므로, 모든 부작용을 2PC 안으로 넣을 수 있다고 가정하면 설계가 깨진다.

43. Spring의 @TransactionalEventListener는 transaction phase에 event listener를 묶을 수 있지만, AFTER_COMMIT listener에서 외부 publish가 실패했을 때 자동으로 DB transaction을 되돌려 주지는 않는다. listener phase는 callback 시점을 조절하는 도구이고, durable retry queue를 대신하지 않는다. outbox row가 필요한 이유가 여기에 있다.

44. saga는 긴 업무 절차를 여러 local transaction과 보상 작업으로 나누는 사고방식이다. 결제 승인, 재고 예약, 배송 요청처럼 각 단계가 별도 시스템이면 하나의 ACID transaction으로 묶기 어렵다. saga는 실패를 없애지 않고, 실패했을 때 어느 단계까지 갔고 어떤 보상 또는 재시도가 필요한지 상태를 드러낸다.

45. outbox worker는 loadPending 후 send가 아니라 claim 후 send여야 한다. 여러 worker가 같은 pending row를 읽고 동시에 publish하면 중복 메시지가 나간다. 조건부 UPDATE, SKIP LOCKED, lease_until 같은 primitive로 처리권을 먼저 선출하고, publish 결과에 따라 SENT 또는 RETRYING으로 바꾸는 흐름이 필요하다.

46. exactly-once publish는 대부분 애플리케이션 경계에서 환상에 가깝다. outbox는 보통 at-least-once 전달을 만들고, 소비자는 inbox 또는 idempotency key로 중복을 흡수한다. 중복을 없애겠다는 목표보다 중복이 와도 업무 결과가 한 번만 반영되게 만드는 목표가 더 검증 가능하다.

47. outbox table을 업무 table에 섞으면 편해 보이지만 worker의 status, attempt_count, next_attempt_at update가 업무 조회와 같은 index를 흔든다. outbox는 업무 commit과 같은 transaction에 참여하되, 운영 패턴은 별도 queue에 가깝다. 그래서 row envelope를 고정한 별도 table이 관측과 성능 면에서 보통 더 안전하다.

48. 2PC는 coordinator recovery log가 핵심이다. prepare 이후 coordinator가 죽으면 participant는 in-doubt 상태에 남을 수 있고, recovery가 commit 또는 rollback 결정을 다시 알려줘야 한다. 이 비용을 이해하지 못하면 'XA를 켜면 원자성이 해결된다'고 생각하지만, 실제로는 transaction manager 운영과 장애 복구 책임이 새로 생긴다.

49. saga 보상은 undo가 아니다. 이미 외부에 보낸 이메일을 지울 수 없고, 이미 승인된 결제를 취소하더라도 수수료나 고객 경험은 남는다. 보상은 업무적으로 반대 효과를 내는 새 action일 뿐이다. 그래서 saga 설계는 기술 흐름보다 고객에게 보이는 중간 상태와 재시도 정책을 먼저 정해야 한다.

50. outbox 관측은 pending 개수만 보면 부족하다. oldest_pending_age, retry_backoff 분포, last_error 상위 원인, claim lease 만료 수, publish latency, consumer idempotency hit rate가 있어야 한다. 메시지가 안 나간 장애와 중복이 많이 나간 장애는 같은 outbox table에서 다른 지표로 보인다.

51. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.

52. 2PC/XA는 여러 resource manager가 한 global transaction에 참여하도록 prepare와 commit을 나누는 강한 모델이다. Jakarta Transactions는 transaction manager, resource manager, application server, XAResource 같은 역할을 정의한다. 하지만 외부 HTTP API나 대부분의 SaaS webhook 대상은 XA resource가 아니므로, 모든 부작용을 2PC 안으로 넣을 수 있다고 가정하면 설계가 깨진다.

53. Spring의 @TransactionalEventListener는 transaction phase에 event listener를 묶을 수 있지만, AFTER_COMMIT listener에서 외부 publish가 실패했을 때 자동으로 DB transaction을 되돌려 주지는 않는다. listener phase는 callback 시점을 조절하는 도구이고, durable retry queue를 대신하지 않는다. outbox row가 필요한 이유가 여기에 있다.

54. saga는 긴 업무 절차를 여러 local transaction과 보상 작업으로 나누는 사고방식이다. 결제 승인, 재고 예약, 배송 요청처럼 각 단계가 별도 시스템이면 하나의 ACID transaction으로 묶기 어렵다. saga는 실패를 없애지 않고, 실패했을 때 어느 단계까지 갔고 어떤 보상 또는 재시도가 필요한지 상태를 드러낸다.

55. outbox worker는 loadPending 후 send가 아니라 claim 후 send여야 한다. 여러 worker가 같은 pending row를 읽고 동시에 publish하면 중복 메시지가 나간다. 조건부 UPDATE, SKIP LOCKED, lease_until 같은 primitive로 처리권을 먼저 선출하고, publish 결과에 따라 SENT 또는 RETRYING으로 바꾸는 흐름이 필요하다.

56. exactly-once publish는 대부분 애플리케이션 경계에서 환상에 가깝다. outbox는 보통 at-least-once 전달을 만들고, 소비자는 inbox 또는 idempotency key로 중복을 흡수한다. 중복을 없애겠다는 목표보다 중복이 와도 업무 결과가 한 번만 반영되게 만드는 목표가 더 검증 가능하다.

57. outbox table을 업무 table에 섞으면 편해 보이지만 worker의 status, attempt_count, next_attempt_at update가 업무 조회와 같은 index를 흔든다. outbox는 업무 commit과 같은 transaction에 참여하되, 운영 패턴은 별도 queue에 가깝다. 그래서 row envelope를 고정한 별도 table이 관측과 성능 면에서 보통 더 안전하다.

58. 2PC는 coordinator recovery log가 핵심이다. prepare 이후 coordinator가 죽으면 participant는 in-doubt 상태에 남을 수 있고, recovery가 commit 또는 rollback 결정을 다시 알려줘야 한다. 이 비용을 이해하지 못하면 'XA를 켜면 원자성이 해결된다'고 생각하지만, 실제로는 transaction manager 운영과 장애 복구 책임이 새로 생긴다.

59. saga 보상은 undo가 아니다. 이미 외부에 보낸 이메일을 지울 수 없고, 이미 승인된 결제를 취소하더라도 수수료나 고객 경험은 남는다. 보상은 업무적으로 반대 효과를 내는 새 action일 뿐이다. 그래서 saga 설계는 기술 흐름보다 고객에게 보이는 중간 상태와 재시도 정책을 먼저 정해야 한다.

60. outbox 관측은 pending 개수만 보면 부족하다. oldest_pending_age, retry_backoff 분포, last_error 상위 원인, claim lease 만료 수, publish latency, consumer idempotency hit rate가 있어야 한다. 메시지가 안 나간 장애와 중복이 많이 나간 장애는 같은 outbox table에서 다른 지표로 보인다.

61. DB commit과 메시지 publish 사이에는 항상 틈이 있다. commit 후 프로세스가 죽으면 DB에는 주문이 있지만 메시지는 나가지 않고, publish 후 commit이 rollback되면 외부에는 없는 주문 이벤트가 나간다. outbox는 이 틈을 없애는 마법이 아니라 '보낼 의도'를 같은 로컬 transaction 안에 저장해 재시작 후 다시 보낼 수 있게 만드는 구조다.
