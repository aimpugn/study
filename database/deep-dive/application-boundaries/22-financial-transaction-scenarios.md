# Financial Transaction Scenarios

## ledger, balance, payment state machine

이 절은 `ledger, balance, payment state machine`를 단순 용어가 아니라 애플리케이션과 데이터베이스가 서로 책임을 넘기는 경계로 설명한다. 읽은 뒤에는 원장, 잔액, 결제 상태 머신이 각각 다른 책임을 가지며 왜 balance 컬럼 하나로 금융 처리를 끝낼 수 없는지 설명할 수 있어야 한다. 이때 특히 버려야 할 오해는 balance 컬럼만 수정하면 금융 처리가 끝난다고 믿는 함정이다. 이 문서는 기존 메모의 문제의식을 보존하되, 공식 자료와 실제 값 trace와 운영 검증 경로로 설명을 다시 세운다.
먼저 결제 성공 이벤트 하나에서 시작한다. 외부 PG가 성공을 알려도 주문 상태, 원장 row, balance snapshot은 서로 다른 질문에 답한다. 주문 상태는 배송해도 되는지 말하고, 원장은 돈이 왜 움직였는지 설명하며, balance는 지금 얼마로 조회할지 빠르게 보여 준다. 그래서 DU46의 첫 질문은 balance를 얼마나 더할 것인가가 아니라 같은 이벤트를 세 모델이 각자 어떤 책임으로 소비하는가다.
### 근거 경계

이 절에서 하중을 지탱하는 근거는 다음처럼 분리한다.

| 구분 | 사용한 근거 | 이 근거로 닫는 판단 |
| --- | --- | --- |
| 공식 자료 | Stripe PaymentIntents API: https://docs.stripe.com/api/payment_intents | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Stripe PaymentIntents guide: https://docs.stripe.com/payments/intents | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Stripe balance transaction types: https://docs.stripe.com/reports/balance-transaction-types | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 로컬 seed | `domains/payment/payment_operation.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `domains/payment/pg.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `domains/firmbanking/_proto.firmbanking_batch_firm.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |

공식 자료는 사실 판정에 쓰고, 로컬 seed는 이 저장소의 학습 맥락과 실무 감각을 보강하는 데 쓴다. 공식 자료가 직접 말하지 않는 운영 설계는 추론으로 표시하고, 재현 가능한 SQL, 로그, 테스트, 대사 절차로 확인한다.
### 첫 번째 벽돌: 결제 한 건이 세 종류의 기록으로 나뉜다

Stripe PaymentIntents 공식 문서는 PaymentIntent가 결제 생애주기를 추적하고, 하나의 주문이나 고객 세션마다 하나를 만들 것을 권장하며, 여러 status를 거쳐 최대 하나의 성공 charge를 만든다고 설명한다. Balance transaction 문서는 계정 balance로 들어오거나 나가는 모든 활동을 balance transaction으로 나타낸다고 설명한다. 이 두 자료를 합치면 결제 시스템에는 최소 세 종류의 기록이 필요하다는 점이 보인다. 사용자의 결제 시도 상태, 돈의 이동을 설명하는 원장 행, 현재 조회를 빠르게 하기 위한 잔액 또는 snapshot이다.

```text
Payment attempt state:
  order O-100 -> payment_intent PI-1 -> requires_action -> processing -> succeeded

Ledger rows:
  L1 debit customer receivable 10000
  L2 credit merchant payable 9700
  L3 credit fee revenue 300

Balance snapshot:
  merchant available balance += 9700 only after settlement/availability rule
```

상태 머신은 '이 주문을 지금 배송해도 되는가'에 답하고, 원장은 '무슨 돈이 어떤 이유로 움직였는가'에 답하며, 잔액은 '현재 얼마로 조회되는가'에 답한다. 세 질문은 겹치지만 같지 않다. 결제 성공 상태가 곧 정산 가능 잔액 증가를 뜻하지 않을 수 있고, 잔액이 맞아 보여도 원장 행이 없으면 왜 그렇게 되었는지 감사할 수 없다.
#### DU46-1.1 상태 머신은 업무 행동을 제한한다

결제 상태는 단순 문자열이 아니라 다음에 허용되는 행동을 제한하는 규칙이다. `requires_payment_method`에서는 결제 수단을 다시 받아야 하고, `requires_action`에서는 고객 인증이 필요하며, `processing`에서는 기다리거나 webhook을 받아야 하고, `succeeded`에서는 fulfillment를 시작할 수 있다. 상태 전이가 없다면 서비스는 timeout 뒤에 취소해도 되는지, capture해도 되는지, 환불해도 되는지 매번 임의 판단을 하게 된다.

함정은 상태를 화면 표시용 enum으로만 보는 것이다. 운영 장애에서는 상태가 곧 guard다. `processing`을 실패처럼 처리하면 나중에 성공 webhook이 와서 이중 처리되고, `requires_capture`를 성공처럼 배송하면 실제 capture 실패 때 회수 비용이 생긴다.

#### DU46-1.2 원장은 변경 불가능한 설명 기록이다

원장 ledger는 잔액을 만들기 위한 중간 테이블이 아니라 돈의 움직임을 설명하는 사실 기록이다. 좋은 원장 row는 amount, currency, debit/credit, account, event id, external id, effective time, created time, source type을 갖고, 이미 기록된 row를 조용히 수정하지 않는다. 잘못된 기록은 반대 방향의 보정 row로 남겨야 감사 추적이 가능하다.

실무 함정은 balance가 맞으면 ledger를 대충 둬도 된다고 믿는 것이다. 고객 문의나 정산 불일치가 생기면 현재 잔액보다 '왜 이렇게 되었는가'가 중요하다. 원장 행이 없으면 장애 당시의 판단을 재구성할 수 없고, 수동 SQL update는 회계 설명력을 파괴한다.

#### DU46-1.3 잔액은 조회 최적화 snapshot이다

잔액 balance는 원장 행을 매번 합산하지 않기 위해 둔 snapshot일 수 있다. 따라서 balance는 원장에서 유도 가능한 값이어야 하고, 원장과 정기적으로 대조되어야 한다. 실시간 서비스에서는 row lock과 동시성 때문에 balance를 업데이트하지만, 그 업데이트는 원장 행과 같은 트랜잭션에 묶거나 재계산 가능한 방식으로 설계해야 한다.

함정은 balance 컬럼을 source of truth로 삼고 원장을 나중에 맞추는 것이다. 장애 후 balance만 남으면 어떤 거래가 빠졌는지 알 수 없다. 반대로 원장이 완전하면 balance snapshot은 재생성할 수 있다. 그래서 audit 관점의 주인은 원장이고, balance는 빠른 조회를 위한 파생 상태로 보는 편이 안전하다.

#### DU46-1.4 pending과 available은 다른 돈이다

결제 승인 직후 고객에게는 성공처럼 보이지만 정산 시스템에서는 아직 pending balance일 수 있다. Stripe balance transaction 문서도 처음 받은 payment가 pending balance로 반영되고 settlement timing에 따라 available이 된다고 설명한다. 이 차이는 지급 가능 금액, 환불 가능 금액, 정산 예정 금액을 나눌 때 중요하다.

실무 함정은 승인 성공 금액을 바로 가맹점 출금 가능 잔액으로 더하는 것이다. 카드 취소, chargeback, 정산 보류, 수수료 확정 전에는 가용 잔액과 총 발생액이 다를 수 있다. pending/available/reserved를 나누지 않으면 출금 가능 금액을 과대 표시한다.

#### DU46-1.5 상태 전이와 원장 기록은 같은 이벤트를 다르게 본다

`payment_succeeded` 이벤트는 상태 머신에는 succeeded 전이를 만들고, 원장에는 매출 또는 수취 예정 행을 만들며, balance에는 pending 증가를 만들 수 있다. 같은 외부 이벤트를 세 모델이 각자 소비하지만, 각 모델의 책임은 다르다. 상태 머신은 중복 이벤트를 무시하거나 전이 가능성을 검사하고, 원장은 같은 event id로 중복 row를 막으며, balance는 원장 반영 결과를 집계한다.

함정은 이벤트 핸들러 하나에서 상태, 원장, 잔액을 순서 없이 막 수정하는 것이다. 중간 실패가 나면 상태는 성공인데 원장이 없거나, 원장은 있는데 balance가 안 맞는 상태가 된다. 처리 순서와 트랜잭션 경계를 명시하고, 재처리 시 idempotent하게 같은 결과가 나오도록 해야 한다.

#### DU46-1.6 대사는 모델 간 관계를 검증한다

대사 reconciliation은 외부 입금액과 내부 balance를 맞추는 작업만이 아니다. payment state, ledger rows, balance snapshot, payout file, bank deposit이 서로 설명 가능한지 확인하는 절차다. 예를 들어 내부 원장 총합은 1,000건 9,700,000원인데 은행 입금은 9,699,700원이라면 수수료 row, 환불 row, 보류 row, dispute row를 추적해야 한다.

운영 함정은 대사를 월말 수동 작업으로만 두는 것이다. 일별 자동 대사, 차이 threshold, missing external id, duplicate ledger event, balance drift metric이 없으면 작은 불일치가 마감일에 큰 장애로 나타난다.



### 상태, 원장, 잔액을 함께 쓰는 SQL trace

```sql
-- one external success event E-9001 arrives
INSERT INTO payment_event(event_id, payment_id, status) VALUES ('E-9001', 'P-77', 'SUCCEEDED');
UPDATE payment SET status = 'SUCCEEDED' WHERE id = 'P-77' AND status IN ('PROCESSING', 'REQUIRES_CAPTURE');
INSERT INTO ledger_entry(event_id, account_id, direction, amount_minor, currency)
VALUES ('E-9001', 'merchant:42', 'CREDIT', 9700, 'KRW'),
       ('E-9001', 'fee:platform', 'CREDIT', 300, 'KRW');
UPDATE balance_snapshot SET pending_minor = pending_minor + 9700 WHERE account_id = 'merchant:42';
```

| 모델 | 질문 | 중복 방지 키 | 복구 방법 |
| --- | --- | --- | --- |
| payment | 주문을 다음 단계로 보내도 되는가 | payment id + allowed transition | 외부 status 재조회 |
| ledger | 돈이 왜 움직였는가 | event id + ledger line id | 보정 row 추가 |
| balance | 지금 얼마로 보여 줄 것인가 | account id snapshot version | ledger에서 재계산 |

이 trace는 실제 운영 코드의 단순화다. 세 update를 하나의 DB 트랜잭션에 넣을 수도 있고, 원장을 먼저 쓰고 snapshot을 비동기로 갱신할 수도 있다. 어떤 방식을 택하든 각 모델의 owner와 재생성 경로가 보여야 한다. 특히 balance snapshot은 틀릴 수 있는 파생값이므로 원장으로 재계산하는 절차가 있어야 한다.
### 관측과 검증

DU46의 검증은 문장 이해가 아니라 실제 관측 지점으로 닫는다. PASS는 기대한 상태 전이가 로그, DB row, 트랜잭션 경계, 응답 코드 중 최소 하나에서 같은 방향으로 보이는 것이다. FAIL은 결과는 맞아 보여도 중간 상태가 설명과 다르게 움직이거나, 장애 상황에서 같은 요청을 다시 실행했을 때 다른 부작용이 생기는 것이다.

```text
DU46 verification checklist
1. replay the same external event twice and assert no duplicate ledger rows
2. compare ledger sum and balance snapshot for each account periodically
3. test invalid payment state transitions such as processing to requires_payment_method after success
4. match payout/bank deposit rows to ledger reporting categories and fees
```

DU46의 검증은 같은 event id를 여러 모델에 재주입해 본다. 첫 주입은 상태 전이, 원장 row, balance snapshot 또는 rebuild queue를 만들고, 두 번째 주입은 중복으로 무시되어야 한다. PASS는 상태, 원장, balance가 서로 설명 가능한 것이고, FAIL은 balance 숫자는 맞지만 원장이나 상태 전이 근거가 비어 있는 것이다.
### 기억할 압축 문장

결제 상태는 다음 행동을 제한하고, 원장은 돈의 이유를 설명하며, 잔액은 조회를 빠르게 하는 파생 snapshot이다. balance 컬럼 하나만 맞추면 금융 처리가 끝난다는 생각은 감사, 복구, 정산이 시작되는 순간 깨진다.

#### DU46-S1 등장 배경: 현재 잔액만으로는 이유를 설명할 수 없다

금융 시스템은 단순 CRUD 잔액 관리에서 시작할 수 있지만, 고객 문의와 회계 마감이 들어오면 현재 숫자만으로 부족해진다. 왜 이 잔액이 되었는지, 어떤 수수료와 환불과 분쟁이 있었는지, 어떤 외부 보고서와 연결되는지 설명해야 한다. 그래서 원장과 상태 머신과 잔액 snapshot이 분리된다.

이 배경을 놓치면 balance update가 가장 중요한 코드처럼 보인다. 실제로는 balance는 빠른 조회를 위한 결과이고, 원장과 상태 전이가 그 숫자를 설명하는 근거다.

#### DU46-S2 상태 머신은 command를 제한한다

Payment 상태가 `PROCESSING`이면 cancel, capture, fulfill, retry의 허용 여부가 모두 달라진다. 상태 머신은 화면 badge가 아니라 command guard다. 허용되지 않는 전이는 버그를 조기에 드러내는 방어선이다.

함정은 enum 값만 있고 transition table이 없는 것이다. 그러면 handler마다 if문이 흩어지고, 어떤 경로에서는 succeeded에서 failed로 내려가는 불가능한 전이가 생긴다.

#### DU46-S3 원장은 append-only에 가까워야 한다

원장 row는 돈의 이동을 설명하는 사건 기록이다. 이미 발행한 회계 사실을 수정하면 과거 보고서와 현재 보고서가 달라진다. 보정이 필요하면 adjustment entry를 추가해 흐름을 설명한다.

물론 구현상 정정 컬럼이나 status가 있을 수 있지만, 금액 사실 자체를 조용히 덮어쓰는 것은 위험하다. 감사와 대사는 변경 이력의 보존을 요구한다.

#### DU46-S4 잔액 snapshot은 재계산 가능해야 한다

고성능 조회를 위해 account_balance table을 두는 것은 자연스럽다. 그러나 그 값은 ledger sum으로 재계산 가능해야 한다. 재계산할 수 없다면 snapshot이 깨졌을 때 복구 경로가 없다.

운영에서는 주기적으로 ledger sum과 snapshot을 비교한다. 차이가 나면 snapshot을 맞추기 전에 어떤 ledger event가 누락됐는지 먼저 찾아야 한다.

#### DU46-S5 수수료와 net amount를 분리한다

결제 10,000원이 성공해도 가맹점에게 정산될 돈은 수수료를 뺀 9,700원일 수 있다. gross, fee, net을 같은 amount로 뭉개면 매출, 비용, 지급 가능액이 섞인다.

Stripe 보고서들이 gross, fee, net, reporting category를 나누는 이유도 회계 분류를 위해서다. 내부 원장도 이 분류가 추적 가능해야 한다.

#### DU46-S6 pending, available, reserved의 시간축

잔액은 발생 시점과 사용 가능 시점이 다를 수 있다. 카드 승인, 매입, 정산, 지급, 은행 입금은 다른 시간축이다. pending은 벌었지만 아직 쓸 수 없는 돈이고, available은 지급 가능하며, reserved는 분쟁이나 리스크로 잡힌 돈일 수 있다.

함정은 성공 이벤트 하나로 available을 올리는 것이다. 나중에 payout이 부족하거나 dispute가 오면 이미 출금 가능하게 보여 준 돈을 되돌려야 한다.

#### DU46-S7 외부 이벤트 중복과 원장 중복

webhook이나 파일 결과는 중복 도착할 수 있다. 상태 머신은 이미 처리한 event를 무시해야 하고, ledger는 같은 external event id로 중복 row가 생기지 않아야 한다.

중복 방지는 handler 시작부의 편의 코드가 아니라 원장 unique key와 idempotent command의 핵심이다. 중복 원장은 balance를 부풀리고 대사를 망가뜨린다.

#### DU46-S8 부분 실패와 부분 성공

배치 지급이나 대량 정산에서는 100건 중 98건 성공, 2건 실패가 정상 결과일 수 있다. 상태 머신은 파일 전체 상태와 개별 거래 상태를 나누고, 원장은 성공 건만 돈의 이동으로 기록해야 한다.

함정은 파일 단위 성공/실패만 두는 것이다. 일부 성공을 전체 실패로 보상하면 이미 지급된 98건을 다시 지급할 위험이 있다.

#### DU46-S9 대사 차이의 분류

대사에서 차이가 나면 누락, 중복, 수수료, 환불, dispute, 보류, 환율, 지급 지연 중 어디에 속하는지 분류해야 한다. 분류가 곧 다음 조치다.

단순히 차액 row 하나로 맞추면 이번 달은 닫힐 수 있지만 다음 달 원인 분석이 불가능해진다. 차이도 설명 가능한 원장 사건이어야 한다.

#### DU46-S10 최종 설계 질문

금융 거래 모델링의 마지막 질문은 어떤 테이블이 source of truth인가가 아니라 어떤 질문에 어떤 모델이 답하는가다. 상태는 행동 가능성을, 원장은 돈의 이유를, 잔액은 현재 조회를 답한다.

이 세 질문이 분리되면 설계가 복잡해 보이지만 장애 때는 단순해진다. 어느 숫자가 틀렸는지, 어느 이벤트가 빠졌는지, 어느 상태 전이가 잘못됐는지 따로 고칠 수 있기 때문이다.

### DU46 추가 실전 사례: 배송 결정과 정산 결정은 같은 성공을 다르게 읽는다

주문 서비스는 결제 상태가 `SUCCEEDED`가 되면 상품을 배송할 수 있다고 판단할 수 있다. 그러나 정산 서비스는 같은 성공을 바로 지급 가능 금액으로 보지 않을 수 있다. 카드 매입, 수수료 확정, 지급 가능일, dispute reserve가 남아 있기 때문이다. 따라서 같은 외부 성공 이벤트라도 주문 상태 머신, 원장, 잔액 snapshot은 서로 다른 질문에 답한다.

```text
external event: payment_intent.succeeded amount=10000 fee_estimate=300
order state machine: PAID -> fulfillment allowed
ledger: recognize gross receivable 10000 and estimated fee 300
balance: pending +9700, available unchanged until settlement date
settlement later: pending -9700, available +9700 or payout queued
```

이 trace에서 배송 가능 여부와 지급 가능 여부는 분리된다. 배송은 고객 결제 흐름이 완료됐는지를 보고, 지급은 실제 자금이 정산 가능한 상태가 됐는지를 본다. 두 결정을 같은 balance 컬럼 하나로 처리하면, 성공 직후 가맹점이 아직 정산되지 않은 돈을 출금할 수 있거나, 반대로 배송 가능한 주문을 정산 대기 때문에 붙잡는 문제가 생긴다.

원장 설계에서는 event sourcing이라는 이름을 붙이지 않더라도 사건 보존 원리가 필요하다. 결제 성공, 수수료 확정, 환불, chargeback, payout, bank deposit은 모두 돈의 설명을 바꾸는 사건이다. 각 사건은 원장 row를 만들거나 기존 pending을 available로 옮기는 전이를 만든다. balance snapshot은 이 원장 row의 현재 합계 view일 뿐이다.

대사 검증은 다음 쿼리 감각으로 닫을 수 있다.

```sql
SELECT account_id, SUM(CASE direction WHEN 'CREDIT' THEN amount_minor ELSE -amount_minor END) AS ledger_sum
FROM ledger_entry
GROUP BY account_id;

SELECT account_id, pending_minor + available_minor + reserved_minor AS snapshot_sum
FROM balance_snapshot;
```

두 결과가 다르면 곧바로 snapshot을 덮어쓰지 않는다. 먼저 빠진 event, 중복 event, 보류 금액, 수수료 확정 차이, payout timing difference를 분류한다. 이 분류가 끝난 뒤에야 재계산이나 adjustment를 수행한다.

### DU46 마지막 검산: balance drift를 발견했을 때의 안전한 순서

balance snapshot과 ledger sum이 다르면 가장 먼저 해야 할 일은 snapshot을 맞추는 update가 아니다. 먼저 차이가 어느 계정, 어느 currency, 어느 기간에서 시작됐는지 찾고, 해당 기간의 payment event, ledger entry, payout/report row를 시간순으로 펼친다. 그다음 누락 event인지, 중복 event인지, pending에서 available로 옮겨야 할 timing 차이인지, 수수료나 dispute가 늦게 반영된 것인지 분류한다.

```text
balance drift triage
  1. account+currency scope fixed
  2. ledger sum by event time calculated
  3. snapshot version and last rebuilt event checked
  4. external report category matched
  5. repair: rebuild snapshot or insert adjustment, never silent overwrite first
```

이 순서를 지키면 잔액 숫자는 잠깐 늦게 고쳐질 수 있지만 원인은 보존된다. 반대로 바로 balance를 덮어쓰면 dashboard는 조용해져도 다음 대사에서 같은 차이가 재발한다. 금융 시스템의 좋은 복구는 숫자를 맞추는 속도보다 왜 맞지 않았는지를 다시 설명할 수 있는 능력에 가깝다.

### DU46 운영 케이스 매트릭스

DU46의 본문을 실제 장애 대응에 쓰려면 `ledger, balance, payment state machine`를 하나의 정의가 아니라 반복해서 판정할 수 있는 케이스 묶음으로 바꿔야 한다. 아래 케이스들은 서로 다른 상황을 다루며, 각 케이스는 입력, 상태 변화, 실패 신호, 관측 지점을 함께 남긴다.

| 케이스 | 입력 또는 상황 | 안전한 판정 | 관측/검증 지점 |
| --- | --- | --- | --- |
| duplicate webhook | 같은 event 두 번 | idempotent ledger insert | event unique key |
| invalid transition | succeeded에서 failed | transition reject | state transition test |
| snapshot drift | ledger와 balance 차이 | 원인 분류 후 재계산 | drift query |
| payout timing | pending과 available 차이 | 지급 가능일 분리 | available_on report |
| fee category | 수수료 누락 | gross/fee/net 분해 | reporting category |
| partial batch | 98 성공 2 실패 | 개별 row 상태 유지 | trailer match |
| manual adjustment | 수동 보정 | 원장 adjustment row | approval log |
| dispute reserve | 분쟁 발생 | reserved balance 분리 | dispute event |

### DU46 면접식 꼬리 질문과 실전 답변

DU46를 면접이나 설계 리뷰에서 설명할 때는 정의보다 반례 대응이 중요하다. 아래 질문들은 본문을 읽은 사람이 실제로 다시 설명할 수 있는지 확인하기 위한 압축 검증이다.

#### Q1. 원장과 잔액 중 무엇이 더 근본인가?

감사와 복구 관점에서는 원장이 더 근본이다. 잔액 snapshot은 빠른 조회를 위한 파생값이며, 원장 합계로 재계산 가능해야 한다.

함정은 dashboard balance가 맞으면 원장 설명이 없어도 된다고 믿는 것이다.

#### Q2. 결제 성공은 곧 available balance 증가인가?

아니다. 승인 성공, 매입, 수수료 확정, settlement, payout은 시간축이 다르다. 성공은 배송 가능성을 만들 수 있지만 지급 가능 잔액은 나중에 available이 될 수 있다.

함정은 성공 이벤트 하나로 모든 모델을 같은 상태로 올리는 것이다.

#### Q3. 상태 머신은 왜 필요한가?

상태 머신은 다음 command를 제한한다. requires_action, processing, succeeded, canceled, disputed 같은 상태에 따라 capture, cancel, fulfill, refund 가능 여부가 달라진다.

함정은 상태를 화면 표시용 문자열로만 두는 것이다.

#### Q4. balance drift를 바로 update로 고치면 왜 위험한가?

숫자는 맞아져도 원인이 사라진다. missing event, duplicate event, fee, dispute, timing difference를 분류한 뒤 rebuild나 adjustment를 해야 다음 대사에서 재발하지 않는다.

함정은 조용한 수동 SQL이 가장 빠른 복구라고 믿는 것이다.

#### Q5. 부분 성공 배치는 어떻게 모델링하는가?

파일 전체 상태와 개별 거래 상태를 나눠야 한다. 100건 중 98건 성공이면 성공 98건의 원장과 실패 2건의 재처리 후보가 함께 남아야 한다.

함정은 파일 단위 실패로 전체 재지급을 열어 중복 지급을 만드는 것이다.

### DU46 마지막 close scenario: 같은 이벤트를 세 모델이 소비하는 순서

외부에서 `payment_succeeded`가 도착하면 상태 머신, 원장, balance snapshot이 모두 반응한다. 그러나 반응 순서와 실패 복구가 정의되어 있지 않으면 중간 장애가 찢어진 상태를 만든다. 예를 들어 상태는 `SUCCEEDED`가 되었는데 원장 insert가 실패하면 주문은 배송됐지만 돈의 설명이 없다. 원장은 있는데 balance snapshot 갱신이 실패하면 조회 금액은 틀릴 수 있지만 원장으로 재계산할 수 있다.

```text
recommended recovery-friendly order
  1. store event inbox with unique event id
  2. validate allowed payment state transition
  3. insert ledger rows with same event id
  4. update or enqueue balance snapshot rebuild
  5. mark inbox processed after all required effects close
```

이 순서는 유일한 정답은 아니지만 복구 가능성을 드러낸다. snapshot 갱신을 비동기로 미룰 수는 있어도 원장 없는 성공 상태는 위험하다. 상태, 원장, snapshot 중 무엇이 실패해도 같은 event id로 재처리할 수 있어야 한다.

### DU46 최종 보강: 원장 row의 idempotency

원장 row도 멱등해야 한다. 같은 외부 이벤트가 두 번 도착했을 때 payment 상태만 중복 방지하고 ledger insert는 두 번 허용하면 balance가 두 배로 늘어난다. 따라서 ledger에는 `source_event_id`, `line_type`, `account_id` 같은 중복 방지 키가 필요하다. 이 키는 단순 기술 id가 아니라 같은 돈의 움직임을 한 번만 설명한다는 회계 불변식이다.

검증은 같은 webhook fixture를 두 번 주입하는 방식으로 단순하게 만들 수 있다. 첫 번째 주입은 상태 전이와 원장 row를 만들고, 두 번째 주입은 이미 처리된 event로 인식되어 원장 row 수와 balance snapshot이 변하지 않아야 한다. 이 테스트가 없으면 운영에서 webhook 재전송이나 결과 파일 재수신이 곧 중복 매출로 바뀐다.

### DU46 상태와 원장 사이의 감사 질문

감사자가 묻는 질문은 보통 `현재 balance가 얼마인가`에서 끝나지 않는다. `이 balance를 만든 사건은 무엇인가`, `그 사건은 외부 어떤 id와 연결되는가`, `같은 사건이 두 번 반영되지 않았다는 증거는 무엇인가`, `이 상태에서 고객에게 어떤 행동을 허용했는가`까지 이어진다. 상태 머신과 원장이 분리되어 있으면 이 질문에 각각 답할 수 있다. 상태는 행동 허용 근거를 말하고, 원장은 금액 변화 근거를 말한다.

### DU46 source boundary note

Stripe 문서는 PaymentIntent 상태와 balance transaction/reporting 개념을 보여 주는 공식 vendor 근거다. 이 절의 원장/잔액 분리 설명은 그 vendor 모델을 그대로 복사하자는 뜻이 아니라, 결제 상태와 돈의 설명과 조회 snapshot을 분리해야 한다는 일반 설계 원리를 로컬 결제/펌뱅킹 seed에 맞춰 재구성한 것이다.

### DU46 최종 보강: balance drift는 숫자 보정 전에 원인을 분류해야 한다

잔액 snapshot이 원장 합계와 다를 때 가장 위험한 조치는 차액만큼 balance를 직접 수정하는 것이다. 그렇게 하면 화면 숫자는 맞아 보일 수 있지만, 왜 차이가 생겼는지 설명하는 근거가 사라진다. 먼저 중복 이벤트, 누락 이벤트, 잘못된 방향의 원장 row, currency 혼합, 수수료 반영 시점, 환불/분쟁 reserve, snapshot rebuild 실패를 분류해야 한다.

안전한 순서는 원장 합계를 기준으로 snapshot을 재계산하고, 외부 보고서와 대조한 뒤, 원장 자체가 틀렸으면 adjustment entry를 추가하는 것이다. 이미 발행된 원장 row를 조용히 수정하면 과거 정산 파일과 현재 DB가 다른 이야기를 하게 된다. 그래서 금융 시스템의 복구는 `UPDATE balance SET amount = ...`가 아니라 어떤 사건을 어떤 근거로 보정했는가를 남기는 일이다.

테스트도 이 순서를 따라야 한다. 같은 외부 event id를 두 번 넣어도 원장 row가 하나만 생기는지, 상태가 성공으로 바뀌어도 원장 생성이 실패하면 재처리 queue가 남는지, snapshot rebuild가 원장 합계와 같은 값을 만드는지 확인한다. 이 세 검증이 닫혀야 balance는 빠른 조회용 파생값이라는 설명이 실제 코드 경계와 맞아진다.

또 하나의 실무 기준은 원장과 상태의 시간 순서를 분리해 기록하는 것이다. 결제 상태는 `AUTHORIZED`, `CAPTURED`, `CANCELED`처럼 고객 행동과 외부 API 진행을 말하고, 원장은 돈의 증감 이유를 말한다. 상태가 먼저 바뀌고 원장이 나중에 따라오는 설계도 있을 수 있고, 원장 row를 먼저 잡고 상태를 확정하는 설계도 있을 수 있다. 어느 쪽이든 실패 중간 상태가 재처리 가능한지 검증해야 한다.

장애 회고에서는 “성공 상태인데 원장이 없다”, “원장은 있는데 balance snapshot이 늦다”, “balance는 맞는데 원장 event id가 중복이다”를 서로 다른 장애로 분류한다. 이 분류가 있어야 운영자는 숫자 하나를 맞추는 대신 원인을 보정할 수 있다.

이 분류는 알람 설계에도 들어가야 한다. balance drift 알람, ledger duplicate 알람, state transition violation 알람은 서로 다른 담당자와 복구 절차를 가질 수 있다. 하나의 “금융 오류” 알람으로 뭉치면 장애 시간 동안 원인을 다시 분해해야 한다.

### 추가 판정 질문: 상태 전이와 돈 이동은 같은 이벤트를 다르게 읽는다

같은 결제 성공 이벤트라도 상태 머신은 다음 행동을 허용하는 근거로 읽고, 원장은 돈이 왜 움직였는지 설명하는 근거로 읽으며, balance snapshot은 빠른 조회를 위한 결과로 읽는다. 이 세 모델이 같은 이벤트를 소비하더라도 책임은 다르다. 하나가 성공하고 하나가 실패할 수 있으므로 재처리 경로도 분리되어야 한다.

검증은 이벤트 하나를 재주입해 본다. 첫 실행에서 상태는 성공으로 바뀌고 원장 row가 생기며 snapshot이 갱신된다. 중간에 snapshot 갱신만 실패시킨 뒤 같은 event id로 재처리하면 상태와 원장은 중복되지 않고 snapshot만 복구되어야 한다. 이 동작이 보이면 세 모델의 책임 분리가 실제 코드로 닫힌 것이다.

DU46의 실무 설계에서 특히 위험한 지점은 `balance`라는 단어가 너무 편하다는 점이다. 하나의 숫자에 pending, available, reserved, fee, dispute, payout timing을 모두 접으면 API 응답은 단순해지지만, 대사와 복구는 어려워진다. 잔액은 사용자가 보는 숫자이고 원장은 그 숫자가 왜 생겼는지 설명하는 기록이다. 두 층을 분리해야 지금 보이는 금액과 회계적으로 증명 가능한 금액을 서로 맞출 수 있다.

그래서 리뷰에서는 balance update SQL보다 먼저 원장 사건의 이름을 본다. `PAYMENT_AUTHORIZED`, `FEE_RECOGNIZED`, `REFUND_CREATED`, `DISPUTE_RESERVED`, `PAYOUT_SENT`, `BANK_DEPOSIT_CONFIRMED`처럼 돈이 움직인 이유가 드러나야 한다. 사건 이름이 없고 amount delta만 있으면 나중에 같은 +9700원이 매출인지 환불 취소인지 보정인지 구분할 수 없다. 금융 모델의 좋은 이름은 문서 미학이 아니라 대사와 감사의 검색 키다.

### DU46 운영 복구 훈련: balance drift 알람을 받은 순간

잔액 불일치 알람이 울리면 운영자는 숫자를 맞추고 싶은 유혹을 받는다. 하지만 balance drift의 첫 조치는 update가 아니라 범위 고정이다. 어느 account, 어느 currency, 어느 기간, 어느 event 이후부터 차이가 났는지 잡아야 한다. 이 범위를 잡지 않고 전체 balance를 덮어쓰면 오늘 화면은 조용해질 수 있지만 다음 대사에서 같은 차이가 다시 나타난다.

```text
balance drift triage
  scope:
    account=A-100, currency=KRW, event_time between T1 and T2
  compare:
    ledger sum by event id
    balance snapshot version and last_applied_event_id
    external payout/report line
  classify:
    missing ledger, duplicate ledger, late fee, dispute reserve, payout timing, snapshot rebuild failure
  repair:
    rebuild snapshot from ledger or append adjustment entry with evidence
```

이 훈련에서 원장은 단순히 balance를 계산하는 재료가 아니다. 원장은 돈이 움직인 이유의 목록이다. 같은 +9700원이라도 결제 성공으로 생긴 pending인지, 수수료 확정 후 net으로 옮긴 값인지, payout 이후 available에서 빠진 값인지 다르다. 그래서 원장 row에는 event id, line type, account, currency, direction, source reference가 있어야 한다. amount만 있으면 숫자는 맞춰도 이유를 잃는다.

상태 머신은 이 drift 분석을 행동 제한과 연결한다. 결제 상태가 `SUCCEEDED`라고 해서 refund, payout, shipment가 모두 가능하다는 뜻은 아니다. dispute가 열렸거나 reserve가 잡혔거나 정산이 아직 pending이면 다음 command는 달라진다. 좋은 금융 모델은 상태가 다음 행동을 제한하고, 원장이 금액 이유를 설명하며, balance snapshot이 빠른 조회를 제공한다. 세 모델을 분리해야 장애 대응에서 어디를 고칠지 말할 수 있다.

검증은 event replay로 닫는다. 같은 외부 event를 두 번 넣었을 때 원장 row가 중복되지 않아야 하고, snapshot 갱신만 실패시킨 뒤 재처리하면 상태와 원장은 그대로 두고 snapshot만 복구되어야 한다. 이 동작이 가능하면 시스템은 중간 실패를 견딘다. 가능하지 않다면 balance는 빠르지만 fragile한 캐시가 된다.

### DU46 추가 운영 연습: 수동 보정 row를 쓸 때의 기준

수동 보정은 숫자를 맞추는 편한 도구처럼 보이지만, 사실상 새로운 돈의 사건을 쓰는 일이다. 따라서 보정 row에는 방향, 금액, currency, 근거 문서, 승인자, 원인 분류가 있어야 한다. 단순히 `adjustment +1000`만 남기면 다음 대사에서 그 1000원이 누락 수수료인지, 고객 보상인지, payout timing 차이인지 알 수 없다.

```text
adjustment entry minimum
  account_id      : M-100
  currency        : KRW
  direction       : CREDIT
  amount_minor    : 1000
  reason_code     : FEE_REPORT_CORRECTION
  evidence_ref    : payout_report_2026_05_19_line_77
  approved_by     : finance-ops-2
  source_event_id : original payment or reconciliation case id
```

이 구조가 있으면 보정도 원장 원칙을 따른다. 과거 row를 조용히 수정하지 않고, 차이가 왜 생겼는지 설명하는 새 사건을 추가한다. 보정 후 balance snapshot은 원장 합계를 다시 반영해 바뀔 수 있지만, balance 자체가 원인을 품지는 않는다. 원인은 adjustment entry와 evidence에 남아야 한다.

운영에서 자주 밟는 지뢰는 "금액이 작으니 직접 balance만 맞추자"는 판단이다. 작은 금액이라도 원인이 중복 이벤트라면 다음에는 큰 금액으로 반복될 수 있다. 반대로 timing difference라면 시간이 지나면 외부 report가 따라올 수 있다. 그래서 보정 전에는 missing, duplicate, fee, timing, dispute, manual compensation 중 어느 분류인지 먼저 정한다. 분류가 닫히지 않으면 보정이 아니라 조사 상태로 남겨야 한다.

이 연습은 ledger와 balance의 책임을 다시 분리한다. ledger는 사건의 이유를 남기고, balance는 빠른 현재 조회를 제공하며, state machine은 다음 행동을 제한한다. 수동 보정도 이 세 책임을 흐리지 않아야 한다.


## timeout, unknown state, settlement, reconciliation

이 절은 `timeout, unknown state, settlement, reconciliation`를 단순 용어가 아니라 애플리케이션과 데이터베이스가 서로 책임을 넘기는 경계로 설명한다. 읽은 뒤에는 외부 성공, 내부 timeout, webhook 지연, 정산 파일 불일치가 생겼을 때 실패 단정 대신 unknown 상태, 재조회, 대사, 보정으로 닫는 흐름을 설명할 수 있어야 한다. 이때 특히 버려야 할 오해는 timeout을 실패로 단정해 이중 취소나 이중 지급을 만드는 함정이다. 이 문서는 기존 메모의 문제의식을 보존하되, 공식 자료와 실제 값 trace와 운영 검증 경로로 설명을 다시 세운다.
먼저 외부 결제 timeout 하나에서 시작한다. 서버가 요청을 보냈지만 응답을 받지 못했다는 사실은 실패 증거가 아니라 관측 실패다. 외부가 성공했을 수도, 처리 중일 수도, 아예 받지 못했을 수도 있다. 그래서 DU47의 첫 질문은 다시 실행할 것인가가 아니라 어떤 external reference와 어떤 증거 사다리로 unknown을 성공, 실패, 보정, 대기 중 하나로 좁힐 것인가다.
### 근거 경계

이 절에서 하중을 지탱하는 근거는 다음처럼 분리한다.

| 구분 | 사용한 근거 | 이 근거로 닫는 판단 |
| --- | --- | --- |
| 공식 자료 | Stripe Payment status updates: https://docs.stripe.com/payments/payment-intents/verifying-status | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Stripe Balance summary report: https://docs.stripe.com/reports/balance | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Stripe bank reconciliation overview: https://docs.stripe.com/reconciliation/overview | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Stripe payout reconciliation report: https://docs.stripe.com/reports/payout-reconciliation | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | Stripe disputes guide: https://docs.stripe.com/disputes | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 로컬 seed | `domains/payment/payment_operation.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `domains/firmbanking/firmbanking_batch.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |
| 로컬 seed | `domains/firmbanking/_proto.firmbanking_batch_firm.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |

공식 자료는 사실 판정에 쓰고, 로컬 seed는 이 저장소의 학습 맥락과 실무 감각을 보강하는 데 쓴다. 공식 자료가 직접 말하지 않는 운영 설계는 추론으로 표시하고, 재현 가능한 SQL, 로그, 테스트, 대사 절차로 확인한다.
### 첫 번째 벽돌: 성공인지 실패인지 모르는 결제

외부 결제나 펌뱅킹 호출에서 가장 위험한 상태는 실패가 아니라 모름이다. HTTP timeout은 서버가 요청을 받지 않았다는 뜻이 아니다. 은행이나 PG가 요청을 처리했지만 응답이 네트워크에서 사라졌을 수 있고, 내부 DB commit 직전에 프로세스가 죽었을 수도 있다. Stripe의 payment status 자료도 PaymentIntent 상태를 조회하고 webhook으로 성공/실패 이벤트를 받아 업무 행동을 결정하라고 안내한다. 즉 응답을 잃었을 때는 새 실행보다 현재 외부 상태 확인이 먼저다.

```text
T0 internal DB writes PENDING payment P-77
T1 server calls external PG confirm PI-77
T2 socket timeout before response body arrives
T3 internal state must become UNKNOWN, not FAILED
T4 recovery worker retrieves PI-77 or receives webhook
T5 if external succeeded -> mark SUCCEEDED and create ledger
T6 if external absent/failed with proof -> mark FAILED or retry safely
```

이 trace가 DU47의 핵심이다. timeout은 관측 실패이지 업무 실패 증거가 아니다. 외부 시스템이 성공했는지 모르는 동안 내부 상태를 실패로 내려 버리면 고객에게 재시도를 열어 주고 중복 승인 위험을 만든다. 반대로 무조건 성공으로 올리면 실제 실패 거래를 배송하거나 지급할 수 있다. 그래서 unknown은 별도 상태여야 한다.
#### DU47-1.1 unknown 상태는 회피가 아니라 안전장치다

unknown은 개발자가 결정을 미룬 흔적이 아니라 증거가 부족한 상황을 시스템이 정직하게 표현한 상태다. 결제, 송금, 정산처럼 외부 부작용이 있는 작업에서는 실패 단정보다 unknown 보존이 안전하다. unknown row는 recovery owner, next check time, external reference, retry count, last error, idempotency key를 가져야 한다.

함정은 unknown을 운영자에게 부끄러운 상태로 보고 cleanup하는 것이다. unknown을 지우면 외부 성공 여부를 잃고, 나중에 webhook이나 정산 파일이 도착했을 때 연결할 내부 row가 없다. unknown은 알람과 복구 queue의 입력이어야지 쓰레기 row가 아니다.

#### DU47-1.2 재조회는 같은 external reference로 해야 한다

timeout 이후 복구는 새 결제 생성이 아니라 이전 요청의 external reference를 조회하는 데서 시작한다. PaymentIntent id, merchant order id, bank trace number, file sequence, 전문 일련번호 같은 값이 있어야 한다. 이 값이 없으면 외부가 성공했는지 확인할 수 없고, 시스템은 중복 실행과 수동 문의 사이에 갇힌다.

실무 함정은 timeout 뒤 같은 payload로 새 요청을 보내는 것이다. 외부가 idempotency key를 지원하더라도 같은 key를 보내지 않거나, 내부 key와 외부 key를 연결하지 않았다면 새 부작용이 생긴다. 복구의 첫 명령은 `create`가 아니라 `retrieve/status inquiry`여야 한다.

#### DU47-1.3 webhook은 빠른 신호이지 유일한 진실이 아니다

webhook은 결제 상태를 빠르게 알려 주지만 네트워크 지연, 중복 전달, 순서 역전이 가능하다. 따라서 webhook handler는 idempotent해야 하고, 이벤트의 created time과 current external status를 함께 봐야 한다. webhook만 믿고 내부 상태를 되돌리면 오래된 실패 이벤트가 최신 성공 상태를 덮을 수 있다.

함정은 webhook 수신 성공을 업무 성공과 동일시하는 것이다. webhook은 이벤트 전달 성공이고, 내부 DB 반영, 원장 생성, balance 갱신, 알림 발송은 별도 단계다. handler 중간 실패를 대비해 event inbox, 처리 상태, 재처리 가능성을 둬야 한다.

#### DU47-1.4 정산은 일 단위 지연된 진실을 가져온다

정산 settlement와 대사 reconciliation은 실시간 API와 다른 시간축을 가진다. Stripe balance summary report는 일정 기간의 balance activity와 payout을 CSV로 내려받아 월말 대사에 쓰는 구조를 설명한다. 펌뱅킹 로컬 seed도 헤더/트레일러 총건수, 총금액, 결과 파일, 실제 입출금 내역을 대조한다고 설명한다. 즉 오늘의 API 성공은 내일의 정산 파일과 은행 입금으로 다시 확인되어야 한다.

운영 함정은 API 성공 로그만 보고 정산을 닫는 것이다. 수수료, 환불, dispute, 보류, payout 실패가 있으면 API 성공 총액과 은행 입금액은 다를 수 있다. 대사는 차이를 없애는 작업이 아니라 차이를 설명 가능한 category와 원장 row로 분해하는 작업이다.

#### DU47-1.5 보정은 삭제가 아니라 반대 방향 기록이다

정산에서 불일치가 발견되면 기존 원장을 수정하거나 삭제하기보다 보정 adjustment row를 남기는 편이 안전하다. 예를 들어 누락 수수료 300원이 확인되면 fee adjustment ledger를 추가하고, 어떤 보고서/파일/외부 id 때문에 보정했는지 연결한다. 그래야 월말 close 후에도 왜 잔액이 바뀌었는지 감사할 수 있다.

함정은 수동 SQL로 balance만 맞추는 것이다. 순간적으로 dashboard 숫자는 맞지만 원장 합계와 보고서가 더 멀어진다. 수동 보정도 event id, 승인자, 사유, source file, before/after sum을 남겨 재현 가능해야 한다.

#### DU47-1.6 장애 대응 runbook은 상태별로 달라야 한다

timeout, unknown, processing, succeeded, failed, disputed, settled는 서로 다른 runbook을 가져야 한다. unknown은 외부 조회, processing은 대기와 webhook, succeeded는 원장/배송 확인, failed는 재시도 가능성 판단, disputed는 reserve/chargeback 원장, settled는 payout/bank deposit 대사로 간다. 모든 상태를 재시도 버튼 하나로 처리하면 이중 지급이나 이중 취소가 난다.

실무 함정은 운영자 화면에 '재처리' 버튼만 두는 것이다. 재처리 버튼은 어떤 상태에서는 조회이고, 어떤 상태에서는 보정이고, 어떤 상태에서는 새 실행이다. 버튼 이름보다 상태별 command가 중요하다. `retryCreate`, `retrieveExternalStatus`, `applyWebhookAgain`, `createAdjustment`, `markReconciled`를 분리해야 한다.



### timeout 복구와 대사 trace

```text
day 0 10:00 internal P-77 UNKNOWN external_ref=pi_77
day 0 10:02 recovery SELECT payment WHERE status=UNKNOWN AND next_check_at <= now()
day 0 10:02 retrieve pi_77 -> status=succeeded
day 0 10:02 insert ledger event=pi_77 amount=10000 fee=300 net=9700
day 1 12:00 balance report shows charge 10000 fee 300 net 9700 available_on=day 3
day 3 payout report includes net 9700 trace_id=bank-ref-123
day 4 bank statement deposit 9700 -> reconciliation closes P-77
```

| 관측 소스 | 도착 시점 | 답하는 질문 | 내부 반영 |
| --- | --- | --- | --- |
| API response | 즉시 또는 timeout | 외부가 지금 뭐라고 답했나 | 상태 후보 |
| webhook | 비동기 | 외부 상태가 바뀌었나 | event inbox 후 처리 |
| balance report | 일 단위 | 계정 balance 활동이 어떻게 분류됐나 | 원장/수수료 대사 |
| payout report | 지급 주기 | 어떤 거래가 어떤 입금 묶음에 들어갔나 | 정산 묶음 close |
| bank statement | 은행 반영 후 | 실제 현금이 들어왔나 | 최종 cash reconciliation |

이 trace에서 같은 결제는 여러 번 확인된다. 첫 API 응답은 빠르지만 잃어버릴 수 있고, webhook은 중복될 수 있으며, report는 늦지만 회계 설명력이 강하고, 은행 입금은 최종 현금 관측이다. 좋은 시스템은 이 관측들을 서로 대체제로 보지 않고 시간축이 다른 증거로 합성한다.
### 관측과 검증

DU47의 검증은 문장 이해가 아니라 실제 관측 지점으로 닫는다. PASS는 기대한 상태 전이가 로그, DB row, 트랜잭션 경계, 응답 코드 중 최소 하나에서 같은 방향으로 보이는 것이다. FAIL은 결과는 맞아 보여도 중간 상태가 설명과 다르게 움직이거나, 장애 상황에서 같은 요청을 다시 실행했을 때 다른 부작용이 생기는 것이다.

```text
DU47 verification checklist
1. simulate external success with internal timeout and verify UNKNOWN then recovery to SUCCEEDED
2. deliver duplicate/out-of-order webhook events and assert idempotent inbox handling
3. compare ledger net amounts to payout report and bank deposit trace id
4. create an adjustment row for a reconciliation difference without mutating original ledger
```

DU47의 검증은 timeout, webhook, report, bank statement를 시간순으로 재생하는 것이다. timeout 직후에는 UNKNOWN이 보존되어야 하고, 같은 external reference 조회나 결과 파일 매칭이 와야 상태가 좁혀져야 한다. PASS는 새 실행 없이 증거로 상태를 닫는 것이고, FAIL은 timeout을 실패로 단정해 이중 지급이나 이중 취소 가능성을 여는 것이다.
### 기억할 압축 문장

외부 부작용이 있는 금융 경계에서 timeout은 실패 증거가 아니라 관측 실패다. unknown을 보존하고 같은 external reference로 조회하며, webhook과 정산 보고서와 은행 입금을 시간축이 다른 증거로 합성해야 이중 취소와 이중 지급을 피할 수 있다.

#### DU47-S1 등장 배경: 분산 시스템은 실패를 말해 주지 않고 침묵한다

외부 PG나 은행과 통신할 때 가장 흔한 장애는 명확한 거절이 아니라 응답을 받지 못한 상태다. 네트워크가 끊겼는지, 외부가 처리 중인지, 성공했지만 응답만 사라졌는지 알 수 없다. 그래서 금융 시스템은 성공/실패 이분법보다 unknown과 대사를 발전시켰다.

이 배경을 이해하면 timeout을 실패로 단정하는 코드가 왜 위험한지 보인다. 침묵은 실패 증거가 아니라 추가 관측이 필요하다는 신호다.

#### DU47-S2 unknown row에는 복구 재료가 있어야 한다

unknown 상태는 external reference, idempotency key, request hash, last attempt time, next check time, owner worker, retry count를 함께 가져야 한다. 그래야 recovery worker가 무엇을 조회해야 하는지 안다.

함정은 status만 UNKNOWN으로 두는 것이다. 조회할 외부 id가 없으면 unknown은 복구 가능한 상태가 아니라 수동 문의 ticket이 된다.

#### DU47-S3 재시도와 재조회는 다르다

재시도는 같은 작업을 다시 시도하는 것이고, 재조회는 이미 보낸 작업의 결과를 확인하는 것이다. timeout 직후에는 대개 재조회가 먼저다. 외부가 받지 않았다는 증거가 있을 때만 재시도가 후보가 된다.

실무에서는 이 둘을 버튼 하나로 합치면 사고가 난다. 운영자에게는 `상태 조회`, `동일 키 재전송`, `새 요청 생성`이 다른 command로 보여야 한다.

#### DU47-S4 webhook 지연과 순서 역전

webhook은 결제 흐름을 빠르게 복구하지만 이벤트 순서를 보장하지 않을 수 있다. 실패 이벤트가 먼저 오고 성공 이벤트가 나중에 오거나, 같은 이벤트가 두 번 올 수 있다. handler는 event id, created time, 현재 외부 상태를 함께 봐야 한다.

함정은 마지막으로 도착한 webhook을 최신 진실로 보는 것이다. 네트워크 도착 순서와 업무 발생 순서는 다르다. 필요하면 retrieve API로 현재 상태를 확인해 event를 보강한다.

#### DU47-S5 정산 파일은 늦지만 강한 증거다

정산 report나 은행 statement는 실시간은 아니지만 회계 close에 가까운 증거다. API 성공 로그가 있어도 정산 파일에 빠졌다면 수수료, 보류, 실패 반전, payout 지연을 확인해야 한다.

함정은 늦게 온 파일을 단순 후처리로 취급하는 것이다. 금융에서는 늦은 증거가 더 강한 경우가 있다. 대사 결과가 내부 ledger를 검증하고 보정하게 해야 한다.

#### DU47-S6 dispute와 reversal은 성공 이후의 반전이다

결제 성공은 모든 위험이 끝났다는 뜻이 아니다. 분쟁, chargeback, 환불, 지급 실패는 성공 이후에도 돈의 방향을 바꾼다. 상태 머신과 원장은 이런 반전 사건을 정상 경로로 받아들여야 한다.

성공 row를 final처럼 잠가 버리면 나중에 dispute를 수동 SQL로 처리하게 된다. 성공 이후 상태와 원장 보정 경로를 미리 모델링해야 한다.

#### DU47-S7 cut-off와 영업일 캘린더

정산과 펌뱅킹은 영업일, 마감 시간, 은행 처리 창에 묶인다. 오늘 발생한 거래가 오늘 available이 되지 않을 수 있고, 주말이나 공휴일에는 payout이 미뤄질 수 있다.

함정은 cron 시간만 보고 SLA를 설계하는 것이다. 외부 cut-off와 reporting availability를 모르면 정상 지연을 장애로 오판하거나, 실제 지연을 정상으로 방치한다.

#### DU47-S8 대사 차이를 처리하는 순서

차이가 나면 먼저 source file과 internal ledger의 기준 기간, timezone, currency, fee 포함 여부를 맞춘다. 그 다음 missing, duplicate, amount mismatch, timing difference, fee/dispute category로 분류한다. 마지막에 adjustment가 필요하면 원장 row로 남긴다.

바로 balance를 고치는 것은 가장 나중이어야 한다. 원인을 모르는 보정은 다음 대사에서 다시 차이를 만든다.

#### DU47-S9 운영 알람은 unknown age를 본다

unknown 수량보다 중요한 것은 나이다. 방금 생긴 unknown은 정상 복구 대상일 수 있지만, SLA를 넘긴 unknown은 고객 영향과 정산 위험이 커진다. 상태별 age histogram이 필요하다.

함정은 전체 실패율만 보는 것이다. 실패율은 낮아도 unknown 몇 건이 큰 금액이면 회계 리스크가 크다. amount-weighted unknown metric도 필요하다.

#### DU47-S10 최종 설계 질문

timeout/settlement/reconciliation 설계의 마지막 질문은 지금 당장 성공인지 실패인지가 아니라 어떤 증거가 오면 상태를 안전하게 좁힐 수 있는가다. API 조회, webhook, report, bank statement가 각각 어떤 불확실성을 줄이는지 정해야 한다.

이 질문이 닫히면 unknown은 두려운 예외가 아니라 복구 가능한 정상 상태가 된다. 시스템은 실패를 추측하지 않고 증거를 기다리며, 증거가 오면 원장과 상태를 재현 가능한 방식으로 업데이트한다.

### DU47 추가 실전 사례: 은행 파일 결과와 내부 unknown의 만남

펌뱅킹 배치에서는 요청 파일을 보냈고 내부는 timeout을 기록했지만, 다음 날 은행 결과 파일에는 일부 건이 성공으로 들어오는 상황이 가능하다. 이때 내부가 전날 timeout을 실패로 단정하고 같은 건을 새 파일로 다시 보냈다면 중복 지급 위험이 생긴다. 반대로 unknown으로 보존하고 파일 일련번호와 거래 식별자를 남겼다면 결과 파일을 받아 내부 상태를 안전하게 좁힐 수 있다.

```text
D0 17:00 send file F-20260519 seq=77 total=50 rows
D0 17:05 socket timeout -> internal batch UNKNOWN, rows UNKNOWN_EXTERNAL_SUBMITTED
D1 09:00 bank result file F-20260519 seq=77 returns 48 success, 2 failed
D1 09:05 match by file seq + row ref -> 48 SUCCEEDED ledger rows, 2 FAILED retry candidates
D1 09:10 trailer amount matches 48 success sum -> reconciliation stage closes
```

이 흐름에서 timeout은 파일 전체 실패가 아니다. 은행이 파일을 받았는지 모르는 상태다. 결과 파일이 오면 거래별로 상태를 좁히고, trailer의 총건수와 총금액을 내부 집계와 맞춘다. 결과 파일이 오지 않으면 은행 조회나 운영 문의로 접수 여부를 확인해야 한다. 새 파일을 만드는 것은 접수되지 않았다는 증거가 닫힌 뒤의 일이다.

운영 runbook은 상태별 command를 분리해야 한다. `UNKNOWN_EXTERNAL_SUBMITTED`에는 `retrieveOrReceiveResult`가 붙고, `FAILED_VALIDATION`에는 입력 수정 후 새 요청이 붙으며, `SUCCEEDED_NOT_RECONCILED`에는 정산 파일 대기가 붙는다. 같은 '재처리' 버튼으로 묶으면 사람이 버튼의 의미를 기억해야 하고, 야간 장애 때 실수가 난다.

대사에서 PASS는 외부 결과 파일의 성공 건수와 금액, 내부 원장 성공 건수와 금액, 은행 입출금 내역이 같은 기준 기간과 같은 currency에서 설명 가능한 것이다. FAIL은 합계만 맞고 개별 row 매칭이 없거나, 개별 row는 맞지만 trailer 합계가 다르거나, 은행 입금은 맞는데 내부 원장에 보정 이유가 없는 경우다.

### DU47 운영 케이스 매트릭스

DU47의 본문을 실제 장애 대응에 쓰려면 `timeout, unknown state, settlement, reconciliation`를 하나의 정의가 아니라 반복해서 판정할 수 있는 케이스 묶음으로 바꿔야 한다. 아래 케이스들은 서로 다른 상황을 다루며, 각 케이스는 입력, 상태 변화, 실패 신호, 관측 지점을 함께 남긴다.

| 케이스 | 입력 또는 상황 | 안전한 판정 | 관측/검증 지점 |
| --- | --- | --- | --- |
| socket timeout | 응답 없음 | UNKNOWN 보존 | unknown age |
| late webhook | 늦은 성공 이벤트 | 현재 상태 조회 후 반영 | event inbox |
| bank result file | 다음 날 결과 파일 | seq로 내부 row 매칭 | file trailer |
| payout mismatch | 입금액 차이 | fee/dispute/timing 분류 | payout report |
| manual retry | 운영자 재처리 | 조회와 새 실행 분리 | command audit |
| cut off | 마감 이후 접수 | 영업일 캘린더 반영 | calendar rule |
| statement deposit | 은행 입금 확인 | cash reconciliation close | bank trace id |
| stale unknown | SLA 초과 unknown | 알람과 owner 지정 | age histogram |

### DU47 면접식 꼬리 질문과 실전 답변

DU47를 면접이나 설계 리뷰에서 설명할 때는 정의보다 반례 대응이 중요하다. 아래 질문들은 본문을 읽은 사람이 실제로 다시 설명할 수 있는지 확인하기 위한 압축 검증이다.

#### Q1. timeout은 실패인가?

아니다. timeout은 응답 관측 실패다. 외부가 요청을 받았는지, 처리했는지, 성공했는지 모를 수 있으므로 UNKNOWN으로 보존하고 같은 external reference로 조회해야 한다.

함정은 실패로 단정하고 새 실행을 열어 이중 지급이나 이중 취소를 만드는 것이다.

#### Q2. webhook이 오면 바로 믿어도 되는가?

webhook은 중요한 신호지만 중복, 지연, 순서 역전이 가능하다. event id로 idempotent하게 처리하고 필요하면 현재 external status를 retrieve해 보강한다.

함정은 마지막으로 도착한 이벤트가 최신 업무 진실이라고 믿는 것이다.

#### Q3. 정산 파일은 왜 필요한가?

실시간 API는 빠르지만 회계 close의 전체 그림을 주지 않는다. 정산 파일과 payout report, bank statement는 수수료, 보류, 지급 묶음, 실제 입금까지 확인하게 해 준다.

함정은 API 성공 로그만으로 정산을 닫는 것이다.

#### Q4. unknown row는 언제 닫히는가?

외부 조회, webhook, 결과 파일, 은행 입금 같은 증거가 와서 성공/실패/보정 중 하나로 좁혀질 때 닫힌다. 시간이 지났다는 이유만으로 실패가 되지 않는다.

함정은 cleanup batch가 unknown을 오래된 실패처럼 삭제하는 것이다.

#### Q5. 대사 차이는 어떻게 처리하는가?

기간, currency, gross/fee/net 기준을 맞춘 뒤 missing, duplicate, timing, fee, dispute, payout mismatch로 분류한다. 원인이 닫히면 adjustment나 snapshot rebuild를 한다.

함정은 차액만 맞추는 보정 row로 원인을 덮는 것이다.

### DU47 마지막 close scenario: unknown을 닫는 증거의 강도

unknown 상태를 닫을 때 모든 증거가 같은 강도를 갖지는 않는다. 클라이언트가 받은 화면 캡처는 보조 신호이고, PG retrieve API의 현재 status는 강한 운영 신호이며, 정산 report와 은행 입금은 회계 close에 가까운 신호다. 좋은 runbook은 이 증거들을 순서 없이 섞지 않고, 어떤 증거가 어떤 불확실성을 줄이는지 적는다.

```text
evidence ladder for UNKNOWN payment
  client screenshot -> suggests user saw success, not accounting proof
  PG retrieve status -> confirms external payment state
  webhook event -> async signal, must be deduped and ordered
  balance/payout report -> confirms settlement grouping and fees
  bank statement -> confirms cash movement
```

이 사다리는 운영자의 판단을 줄여 준다. unknown이 오래됐을 때 새 결제를 만들지, 외부 조회를 할지, 은행 결과 파일을 기다릴지, 수동 보정을 할지는 증거 강도에 따라 달라진다. 증거 강도 없이 '고객이 성공이라고 했다' 또는 'timeout이니까 실패다'로 닫으면 이중 지급과 미배송이 모두 가능해진다.

### DU47 최종 보강: 운영자에게 보여 줄 상태 언어

unknown 복구가 어려운 이유 중 하나는 운영 화면이 상태의 의미를 제대로 보여 주지 않기 때문이다. `실패`, `처리중`, `재처리` 같은 넓은 단어만 있으면 운영자는 어떤 버튼이 조회이고 어떤 버튼이 새 실행인지 기억해야 한다. 화면과 runbook에는 `외부 제출 여부 모름`, `외부 reference 있음, 조회 필요`, `결과 파일 대기`, `정산 불일치 조사`, `보정 승인 대기`처럼 다음 행동이 드러나는 상태 언어가 필요하다.

이 언어는 개발자 내부 enum 이름과 꼭 같을 필요는 없지만, enum에서 화면 문구와 runbook command로 추적 가능해야 한다. 예를 들어 `UNKNOWN_EXTERNAL_SUBMITTED`는 운영 화면에서 `외부 처리 여부 확인 필요`로 보이고, 가능한 command는 `외부 상태 조회`와 `은행 결과 파일 매칭`이어야 한다. `새 지급 생성`은 이 상태에서 비활성화되어야 한다. 이렇게 UI 제약까지 연결해야 timeout을 실패로 단정하는 사람 실수를 줄일 수 있다.

대사도 같은 원리다. `차이 있음`이라는 상태만 보여 주면 운영자는 엑셀로 맞추려 한다. `수수료 차이 후보`, `입금일 차이 후보`, `dispute reserve 후보`, `내부 원장 누락 후보`처럼 분류가 보이면 다음 확인 자료가 정해진다. 좋은 상태 언어는 복구 절차의 절반이다.

### DU47 자동 복구와 수동 승인의 경계

unknown을 모두 사람이 처리하게 만들면 운영 비용이 커지고, 모두 자동 처리하면 증거가 약한 상태에서 돈을 움직일 수 있다. 자동 복구는 같은 external reference로 조회해 성공/실패가 명확히 닫히는 경우에 적합하다. 수동 승인은 외부 자료가 충돌하거나, 금액이 크거나, 은행/PG 문의가 필요한 경우에 필요하다. 이 경계는 금액, 경과 시간, 외부 상태, 정산 마감 임박 여부로 정할 수 있다.

운영 화면에는 자동 복구 대상과 수동 승인 대상을 분리해 보여야 한다. 그래야 담당자는 모든 unknown을 같은 긴급도로 보지 않고, 먼저 고객 영향과 회계 영향이 큰 건을 처리할 수 있다. 검증은 recovery job이 자동 처리한 건과 수동 queue로 보낸 건의 이유를 로그와 DB row에서 재구성할 수 있는지로 닫는다.

DU47의 마지막 기준도 같다. timeout을 만났을 때 새 실행 버튼을 누르기 전에, 어떤 external reference를 어떤 증거로 조회할지 먼저 말할 수 있어야 한다.

### DU47 source boundary note

Stripe status, report, reconciliation 문서는 timeout 이후 확인할 수 있는 외부 증거의 종류를 보여 준다. 국내 펌뱅킹 배치에서는 PG report 대신 은행 결과 파일, 전문 일련번호, 파일 trailer, 실제 계좌 입출금 내역이 같은 역할을 한다. 이름은 달라도 원리는 같다. timeout을 실패로 단정하지 말고 더 강한 증거로 상태를 좁힌다.

### DU47 최종 보강: unknown을 빨리 없애려는 마음이 가장 큰 장애를 만든다

unknown 상태는 보기 불편하다. 운영 화면에 남아 있으면 담당자는 빨리 성공이나 실패로 바꾸고 싶어진다. 하지만 외부 부작용이 있는 금융 경계에서는 이 불편함이 안전장치다. timeout 직후에는 실제 외부 처리 여부를 모르므로, 내부 상태를 실패로 바꾸고 새 요청을 보내면 이중 지급이나 이중 취소가 생길 수 있다. 반대로 성공으로 단정하면 실제 실패한 건을 배송하거나 포인트를 지급할 수 있다.

좋은 복구 설계는 unknown을 없애는 것이 아니라 좁힌다. 먼저 같은 external reference로 외부 조회를 하고, webhook 또는 결과 파일이 오면 같은 idempotency key와 request hash에 묶는다. 그래도 모순이 있으면 정산 report, 은행 입금 내역, 고객 노출 화면을 증거 강도별로 정리한다. 각 단계는 새 부작용을 만들지 않고 증거를 수집하는 방향이어야 한다.

운영 runbook에는 버튼도 구분되어야 한다. `외부 상태 조회`는 안전한 관측이고, `재전송`은 새 부작용을 만들 수 있으며, `수동 성공 처리`는 강한 승인과 증거 첨부가 필요하다. 이 버튼들이 같은 색과 같은 위치에 있으면 사람은 긴급 상황에서 잘못 누른다. 그래서 unknown 처리의 마지막 품질은 DB schema가 아니라 운영자가 실수하기 어려운 화면과 권한 설계까지 이어진다.

정산 파일은 unknown을 닫는 마지막 증거가 될 때가 많다. API 조회는 성공을 말하지만 수수료와 입금 묶음은 report나 은행 파일에서 드러난다. 따라서 reconciliation job은 단순히 성공/실패를 갱신하는 batch가 아니라 외부 증거와 내부 원장을 같은 기준으로 맞추는 절차다. 금액, currency, fee, settlement date, external reference가 모두 맞아야 한다.

unknown 상태가 오래 남을 때도 무조건 나쁜 것은 아니다. 더 강한 증거를 기다리는 중이라면 상태가 남아 있는 편이 새 부작용보다 안전하다. 다만 owner, next check time, escalation rule이 비어 있으면 unknown은 안전장치가 아니라 방치가 된다. 좋은 시스템은 unknown을 보존하되, 누가 언제 어떤 증거로 닫을지 함께 보존한다.

### 추가 판정 질문: unknown을 닫는 증거는 어디서 왔는가

unknown 상태를 성공이나 실패로 바꿀 때는 증거 출처를 남겨야 한다. 외부 API 조회로 닫았는지, webhook으로 닫았는지, 정산 파일로 닫았는지, 은행 입금으로 닫았는지에 따라 신뢰도와 후속 조치가 다르다. 증거 출처가 없는 상태 변경은 나중에 분쟁이나 회계 문의가 들어왔을 때 다시 설명할 수 없다.

그래서 unknown table에는 status뿐 아니라 evidence_type, evidence_ref, observed_at, operator 또는 worker id가 필요하다. 자동 복구가 했든 사람이 했든, 어떤 증거로 불확실성을 줄였는지 남겨야 한다. 이 기록이 있으면 timeout 장애는 단순 실패 로그가 아니라 복구 가능한 상태 기계가 된다.

마지막으로 unknown 처리는 시간과 책임을 함께 저장해야 한다. `UNKNOWN`이라는 값만 있으면 오래된 row가 방치인지 안전한 대기인지 알 수 없다. `next_check_at`, `owner_team`, `last_evidence_type`, `last_external_status`, `retry_allowed` 같은 필드가 있으면 운영자는 같은 unknown이라도 자동 조회 대기, 수동 확인 필요, 정산 파일 대기, 고객 안내 필요를 나눌 수 있다.

이 구분은 장애 이후 회고에서도 중요하다. unknown이 30분 남아 있었지만 외부 조회가 계속 실패했고 재실행은 금지되어 있었다면 시스템은 안전하게 버틴 것이다. 반대로 unknown이 5분만 남았어도 owner와 다음 확인 시각이 비어 있었다면 방치다. 좋은 금융 시스템은 unknown의 존재 자체를 실패로 보지 않고, unknown을 어떤 증거로 좁혀 가고 있는지를 품질 기준으로 본다.

### DU47 운영 복구 훈련: timeout 뒤 고객이 성공 화면을 봤다고 말할 때

고객이 성공 화면을 봤다는 말은 중요한 단서지만, 그것만으로 회계 상태를 닫을 수는 없다. 화면은 클라이언트가 어느 순간 어떤 응답을 보았는지 알려 줄 뿐이고, 실제 외부 승인, 내부 원장, 정산 묶음, 은행 입금은 서로 다른 증거다. 운영자는 고객의 말을 무시해서도 안 되지만, 그 말을 곧바로 성공 확정으로 바꿔서도 안 된다. 먼저 내부 request id, external reference, webhook event id, PG retrieve 결과, 정산 report를 같은 사건으로 묶어야 한다.

```text
customer says success after timeout
  1. find internal request by idempotency key or order id
  2. check whether external_ref was stored before timeout
  3. retrieve external status using the same reference
  4. match webhook/event inbox by event id and payment id
  5. if settlement period passed, compare payout/report/bank statement
  6. close UNKNOWN with evidence_type and evidence_ref
```

이 순서의 핵심은 새 부작용을 만들지 않는 것이다. 조회, 매칭, 대사는 관측이다. 재전송, 수동 성공 처리, 수동 취소는 상태를 바꾸는 행동이다. unknown 복구 화면에서 이 둘이 같은 버튼처럼 보이면 장애 중 사람은 빠른 버튼을 누른다. 그래서 권한과 UI는 상태 기계의 일부다. `외부 상태 조회`는 넓게 허용할 수 있지만, `수동 성공 확정`은 증거 첨부와 승인 기록이 필요하다.

정산 파일이 늦게 오는 업무에서는 unknown이 하루 이상 남을 수도 있다. 이 자체는 실패가 아니다. 다만 next check time, owner, evidence ladder가 있어야 한다. 예를 들어 은행 결과 파일을 기다리는 unknown은 다음 영업일 오전에 자동 매칭 job이 다시 확인하고, 그래도 닫히지 않으면 담당자 queue로 올라가야 한다. 이런 흐름이 있으면 unknown은 방치가 아니라 관리되는 불확실성이다.

마지막 판단은 항상 같은 질문으로 돌아온다. 지금 하려는 행동은 증거를 더 모으는가, 아니면 돈의 상태를 바꾸는가. 전자라면 안전하게 자동화할 수 있는 경우가 많고, 후자라면 멱등키, 원장, 승인, 대사 근거가 함께 있어야 한다. timeout을 잘 다루는 시스템은 실패를 빨리 단정하는 시스템이 아니라, 모름을 안전하게 보존하고 더 강한 증거로 좁혀 가는 시스템이다.

### DU47 마지막 운영 연습: reconciliation job은 무엇을 소비하는가

대사 작업은 단순히 외부 파일을 읽어 내부 상태를 맞추는 batch가 아니다. 대사 job은 외부 증거를 소비하고, 내부 원장과 상태 머신이 그 증거를 이미 반영했는지 확인하며, 모순이 있으면 보정 후보를 만든다. 이때 외부 증거는 실시간 API, webhook, 정산 report, 은행 입금 내역처럼 강도가 다르다. job은 가장 최신으로 보이는 자료가 아니라 회계적으로 닫을 수 있는 자료를 기준으로 삼아야 한다.

```text
reconciliation input bundle
  internal payment: payment_id, status, external_ref, idempotency_key
  internal ledger : event_id, gross, fee, net, currency, direction
  external report : external_ref, reporting_category, fee, payout_id
  bank statement  : deposit amount, date, trace id
  decision        : matched, missing_internal, missing_external, fee_diff, timing_diff, manual_review
```

이 bundle이 있으면 차이를 곧바로 실패로 보지 않는다. 외부 report에는 payout 묶음 때문에 하루 늦게 보이는 금액이 있을 수 있고, 내부 원장은 이미 성공을 기록했지만 은행 입금은 다음 영업일에 올 수 있다. 반대로 은행 입금은 있는데 내부 원장이 없으면 더 심각하다. 돈은 움직였는데 시스템이 이유를 기록하지 못한 상태이기 때문이다. 따라서 reconciliation의 핵심은 숫자 일치보다 차이 분류다.

unknown 상태도 이 job과 연결된다. 실시간 조회로는 모호했던 payment가 정산 파일에서 성공으로 확인될 수 있고, webhook이 늦게 도착해 이미 닫힌 상태를 다시 건드릴 수 있다. 그래서 event inbox와 reconciliation job은 같은 idempotency 원칙을 공유해야 한다. 같은 외부 증거를 두 번 읽어도 원장 row가 중복되면 안 되고, 이전 판단과 충돌하면 수동 검토 queue로 보내야 한다.

좋은 운영 문서는 이 절차를 사람의 기억에 맡기지 않는다. 어떤 상태에서 어떤 외부 자료를 기다리는지, 어떤 차이는 자동 보정 가능한지, 어떤 차이는 승인자가 필요한지 runbook에 적는다. 금융 트랜잭션 시나리오를 제대로 이해했다는 것은 DB 트랜잭션 하나를 설명하는 것을 넘어, 외부 세계의 증거가 늦게 도착해도 내부 원장과 상태를 일관되게 닫는 방법을 설명할 수 있다는 뜻이다.

### DU46/DU47 연결 연습: 같은 사건을 두 번 보아도 한 번만 돈이 움직여야 한다

금융 시스템은 같은 외부 사건을 여러 경로로 받을 수 있다. 결제 성공은 실시간 API 응답으로 오고, webhook으로 다시 오고, 정산 report에도 나오며, 은행 입금 내역으로 마지막에 확인될 수 있다. 이 네 신호를 모두 새 사건처럼 처리하면 원장과 balance가 계속 부풀어 오른다. 반대로 하나만 믿고 나머지를 버리면 회계 close에서 필요한 수수료, 지급 묶음, 실제 입금 증거를 잃는다.

그래서 event identity와 evidence identity를 나눠야 한다. event identity는 같은 돈의 움직임을 한 번만 원장에 반영하기 위한 키다. evidence identity는 그 사건을 어떤 외부 자료로 확인했는지 남기는 키다. 예를 들어 `payment_intent.succeeded` webhook과 payout report line은 같은 결제 성공을 가리킬 수 있지만, 후자는 수수료와 지급 묶음을 더 잘 설명한다. 원장 row는 중복되면 안 되고, 증거 row는 보강될 수 있어야 한다.

```text
same business event, multiple evidence
  event key      : payment=P-77, movement=CAPTURED, amount=10000 KRW
  ledger effect  : one CREDIT gross row, one FEE row
  evidence #1    : API response auth=A-77
  evidence #2    : webhook event=evt_1
  evidence #3    : payout report line=po_9
  evidence #4    : bank statement trace=bank-20260519-7
```

이 모델을 쓰면 reconciliation이 원장을 다시 쓰는 batch가 아니라 증거를 추가하고 차이를 분류하는 절차가 된다. 원장에 없는 외부 증거는 missing internal 후보가 되고, 내부 원장에 있지만 외부 report에 없는 값은 timing difference나 누락 후보가 된다. 같은 evidence가 두 번 들어오면 inbox/idempotency로 흡수하고, 같은 event에 서로 다른 amount가 붙으면 자동 보정이 아니라 수동 검토로 보내야 한다.

이 연결 연습은 ledger, balance, state machine, unknown 복구를 하나로 묶는다. 상태 머신은 "이 사건으로 다음 행동을 허용해도 되는가"를 답하고, 원장은 "돈이 왜 움직였는가"를 답하며, evidence table은 "그 판단을 어떤 외부 자료로 확인했는가"를 답한다. 이 세 질문이 분리되어 있으면 timeout, webhook 중복, 정산 차이, 은행 입금 지연이 한꺼번에 와도 각 문제를 자기 자리에서 처리할 수 있다.
