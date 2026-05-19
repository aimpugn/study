# Idempotency And Duplicate Request

## duplicate request and idempotency key

이 절은 `duplicate request and idempotency key`를 단순 용어가 아니라 애플리케이션과 데이터베이스가 서로 책임을 넘기는 경계로 설명한다. 읽은 뒤에는 중복 요청을 UI debounce가 아니라 DB winner election, request fingerprint, 상태별 replay로 다뤄야 하는 이유를 설명할 수 있어야 한다. 이때 특히 버려야 할 오해는 UI debounce만으로 중복 결제를 막을 수 있다고 믿는 함정이다. 이 문서는 기존 메모의 문제의식을 보존하되, 공식 자료와 실제 값 trace와 운영 검증 경로로 설명을 다시 세운다.
먼저 작은 중복 결제 요청 하나에서 시작한다. 사용자가 결제 버튼을 눌렀고 서버는 PG 승인을 마쳤지만 응답이 모바일 네트워크에서 사라질 수 있다. 같은 사용자가 다시 누르거나 클라이언트가 자동 재시도하면 서버에는 같은 업무 의미의 POST가 두 번 들어온다. 이때 DU44의 첫 질문은 버튼을 막았는가가 아니라 외부 부작용 이전에 DB가 winner를 선출했고 같은 key의 요청 의미를 hash로 확인했는가다.
### 근거 경계

이 절에서 하중을 지탱하는 근거는 다음처럼 분리한다.

| 구분 | 사용한 근거 | 이 근거로 닫는 판단 |
| --- | --- | --- |
| 공식 자료 | RFC 9110 HTTP Semantics: https://www.rfc-editor.org/rfc/rfc9110 | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 공식 자료 | IETF HTTPAPI Idempotency-Key draft archive: https://datatracker.ietf.org/doc/draft-ietf-httpapi-idempotency-key-header/ | 표준 API나 vendor가 보장하는 상태, 용어, 실패 규칙 |
| 로컬 seed | `knowledge/cards/K-IDEMPOTENT-EXECUTION-WINNER-ELECTION-REPLAY.md` | 이 저장소에서 이미 잡아 둔 사고 흐름과 실무 사례 |

공식 자료는 사실 판정에 쓰고, 로컬 seed는 이 저장소의 학습 맥락과 실무 감각을 보강하는 데 쓴다. 공식 자료가 직접 말하지 않는 운영 설계는 추론으로 표시하고, 재현 가능한 SQL, 로그, 테스트, 대사 절차로 확인한다.
### 첫 번째 벽돌: 같은 POST가 두 번 들어오는 시간표

HTTP RFC 9110은 PUT, DELETE, 안전한 메서드처럼 의도한 효과가 여러 번 실행해도 한 번과 같은 메서드를 멱등하다고 설명한다. 그러나 결제 생성 같은 POST는 기본적으로 그런 보장을 갖지 않는다. IETF HTTPAPI의 Idempotency-Key draft는 POST나 PATCH 같은 비멱등 메서드를 장애에 견디게 만들기 위한 헤더 필드를 다룬다. 여기서 중요한 점은 헤더 이름 자체가 아니라 서버가 그 키로 어떤 실행 권한과 응답 재사용 정책을 만들었는가다.

```text
T0 browser sends POST /payments key=K-20260519 amount=10000
T1 mobile network stalls before response reaches browser
T2 user taps again or client retries POST /payments key=K-20260519 amount=10000
T3 server must decide: one external charge, one stored result, replay for duplicate
wrong path: find none -> call PG twice -> unique row fails after side effect
right path: INSERT PENDING first -> winner calls PG -> loser waits/reads/replays
```

이 시간표에서 위험한 순간은 T1과 T2 사이가 아니다. 진짜 위험은 서버가 외부 PG 호출 전에 winner를 선출하지 않은 상태다. UI debounce는 T2를 줄일 수 있지만 네트워크 재시도, 브라우저 새로고침, 앱 재전송, 로드밸런서 timeout, 메시지 재배달을 막지 못한다. 그래서 서버 저장소가 외부 부작용 이전에 단 하나의 실행권을 만들어야 한다.
#### DU44-1.1 winner election은 외부 호출보다 앞서야 한다

멱등 처리의 핵심은 '중복 row를 막는다'가 아니라 '중복 부작용을 막는다'이다. `SELECT`로 기존 요청을 찾고 없으면 PG를 호출한 뒤 insert하는 방식은 두 요청이 동시에 none을 볼 수 있다. DB 유니크 제약이 마지막 insert에서 하나를 실패시켜도 이미 PG 호출은 두 번 나갔다. 따라서 첫 DB 작업은 `(idempotency_key)` 유니크 제약이 걸린 `INSERT PENDING`이어야 한다. insert에 성공한 요청만 winner이고, 실패한 요청은 loser로서 기존 row를 읽는다.

운영 함정은 유니크 제약을 걸었다는 사실만으로 안전하다고 믿는 것이다. 유니크 제약은 그 제약이 실행되는 시점 이후의 DB row 중복만 막는다. 외부 호출이 제약보다 앞서면 보호 대상이 늦다. 장애 로그에서 같은 idempotency key로 PG 승인 번호가 두 개 보이면 DB row가 하나여도 멱등은 실패한 것이다.

#### DU44-1.2 request fingerprint가 없으면 다른 요청을 같은 요청으로 오인한다

idempotency key는 재시도 식별자이지 요청 의미 전체가 아니다. 같은 키로 amount, currency, merchant, receiver가 달라진 요청이 들어오면 이는 재시도가 아니라 충돌이다. 그래서 서버는 canonical request를 정규화해 hash를 저장하고, duplicate가 들어오면 stored hash와 incoming hash를 비교해야 한다. 같으면 replay 후보이고, 다르면 409류 충돌이나 도메인 예외로 드러내야 한다.

함정은 클라이언트가 키를 잘 만들 것이라고 믿는 것이다. 모바일 앱 버그, 브라우저 탭 복제, 서버 간 재시도, 운영 수동 재처리에서 키 재사용은 충분히 일어난다. fingerprint 없이 성공 응답을 replay하면 다른 금액 요청에 이전 결제 결과를 돌려주는 silent corruption이 된다.

#### DU44-1.3 상태는 PENDING, SUCCEEDED, FAILED, UNKNOWN을 분리한다

winner가 `PENDING` row를 만든 뒤 PG를 호출하면 그 사이 duplicate loser가 들어올 수 있다. 이때 loser는 성공 응답을 아직 replay할 수 없다. 무조건 실패로 돌리면 클라이언트가 다시 재시도하고, 무조건 새 실행을 허용하면 중복 결제가 난다. 그래서 상태는 진행 중, 성공, 실패, 알 수 없음으로 분리되어야 한다. 특히 timeout은 실패가 아니라 unknown일 수 있다. 외부가 받았는지 모르는 요청은 조회나 webhook, 대사로 닫아야 한다.

실무 함정은 timeout을 곧바로 실패로 표시하고 새 결제를 열어 주는 것이다. 사용자는 한 번 눌렀는데 실제 PG에는 첫 요청이 성공했고 내부는 실패로 보고 두 번째 요청을 또 승인할 수 있다. `UNKNOWN` 상태는 불편하지만 금융 경계에서는 정직한 상태다.

#### DU44-1.4 replay는 응답 계약이다

멱등 duplicate가 성공 row를 읽었을 때 같은 응답을 돌려줄지, 이미 처리됨만 말할지, 최신 상태 조회 링크를 줄지는 API 계약이다. 결제 생성 API는 클라이언트가 첫 응답을 잃었을 때도 주문 완료 화면으로 복구해야 하므로 성공 응답 replay가 강한 선택인 경우가 많다. replay하려면 response payload, external id, status, createdAt, request hash를 저장해야 한다.

함정은 replay payload를 나중에 serializer 변경으로 읽지 못하게 만드는 것이다. 저장된 응답은 과거 버전 클라이언트와 재시도 흐름의 계약이므로 schema version이나 최소 canonical response를 따로 설계하는 편이 안전하다.

#### DU44-1.5 멱등키의 보존 기간과 정리 정책도 계약이다

멱등키 row를 영원히 보관하면 저장소가 커지고 개인정보나 응답 payload 보존 부담이 생긴다. 너무 빨리 지우면 늦은 retry가 새 실행으로 보일 수 있다. 따라서 TTL은 결제 수단, 클라이언트 retry 정책, PG 조회 가능 기간, 정산/대사 주기, 법적 보존 요구를 함께 보고 정한다. 삭제하더라도 충돌 감지에 필요한 최소 digest를 더 오래 보존할 수 있다.

운영 함정은 cleanup batch가 성공 row와 pending row를 같은 기준으로 삭제하는 것이다. pending이나 unknown은 복구 대상이고, succeeded는 replay 대상이며, conflict는 감사 대상이다. 상태별 보존과 알람 기준이 달라야 한다.

#### DU44-1.6 관측은 key, hash, winner, external id를 한 줄로 묶는다

멱등 장애를 디버깅하려면 request id만으로 부족하다. 로그에는 idempotency key, request hash, winner/loser 판정, stored status, external transaction id, replay 여부가 함께 있어야 한다. DB에는 unique violation count, pending age, unknown age, replay count, conflict count가 metric으로 올라와야 한다. 이 값들이 있어야 duplicate 폭증이 정상 retry인지 클라이언트 버그인지 구분할 수 있다.

함정은 PG 승인번호만 보고 내부 idempotency 상태를 놓치는 것이다. 외부는 한 번 성공했는데 내부 pending이 오래 남으면 클라이언트는 계속 진행 중을 보고, 운영자는 성공인지 실패인지 수동 조회를 해야 한다. unknown/pending age 알람이 없으면 이 회색 상태가 정산 시점까지 숨어 있다.



### DB winner election 예시

```sql
CREATE TABLE idempotency_request (
  idempotency_key varchar(120) PRIMARY KEY,
  request_hash char(64) NOT NULL,
  status varchar(20) NOT NULL,
  external_id varchar(120),
  response_payload text,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);

-- winner path
INSERT INTO idempotency_request(idempotency_key, request_hash, status, created_at, updated_at)
VALUES ('K-20260519', 'HASH-A', 'PENDING', now(), now());

-- loser path after duplicate key error
SELECT request_hash, status, external_id, response_payload
FROM idempotency_request WHERE idempotency_key = 'K-20260519';
```

이 schema는 완성형이 아니라 first brick이다. 실제 서비스에서는 merchant scope, operation type, expiry, locked_until, error code, response version, unique key 범위를 추가해야 한다. 중요한 것은 `INSERT PENDING`이 외부 호출보다 먼저 있고, duplicate는 이미 존재하는 row를 기준으로 의미 충돌과 replay 가능성을 판단한다는 순서다.

| 입력 | 저장 row | 서버 판단 | 외부 호출 | 응답 |
| --- | --- | --- | --- | --- |
| key K, hash A, row 없음 | 없음 | winner | 호출 1회 | 결과 저장 후 반환 |
| key K, hash A, SUCCESS | hash A | duplicate replay | 호출 없음 | 저장 응답 반환 |
| key K, hash B, SUCCESS | hash A | conflict | 호출 없음 | 키 재사용 오류 |
| key K, hash A, PENDING old | hash A | recovery needed | 조회 또는 보류 | 진행 중/unknown |
### 관측과 검증

DU44의 검증은 문장 이해가 아니라 실제 관측 지점으로 닫는다. PASS는 기대한 상태 전이가 로그, DB row, 트랜잭션 경계, 응답 코드 중 최소 하나에서 같은 방향으로 보이는 것이다. FAIL은 결과는 맞아 보여도 중간 상태가 설명과 다르게 움직이거나, 장애 상황에서 같은 요청을 다시 실행했을 때 다른 부작용이 생기는 것이다.

```text
DU44 verification checklist
1. run concurrent duplicate requests and assert external action counter is 1
2. send same key with different request hash and assert conflict without external call
3. simulate lost response and verify duplicate receives stored successful response
4. simulate external timeout and verify UNKNOWN is reconciled by query/webhook instead of new charge
```

DU44의 검증은 외부 action counter와 DB 상태를 함께 봐야 한다. 동시에 같은 key를 두 번 보내도 PG 호출은 한 번이어야 하고, 같은 key 다른 hash는 외부 호출 없이 conflict가 되어야 하며, 성공 후 lost response는 저장 응답 replay로 복구되어야 한다. 특히 멱등 경계에서는 성공 응답보다 무엇을 다시 실행하지 않았는지가 더 중요하다.
### 기억할 압축 문장

멱등키는 버튼 중복 방지 이름표가 아니다. 서버가 외부 부작용 전에 winner를 선출하고, 같은 키의 의미를 request hash로 검증하며, 성공 응답을 재사용하거나 unknown을 복구하는 전체 실행 계약이다.

#### DU44-S1 등장 배경: retry가 사용자 경험을 살리면서 중복 부작용을 만든다

네트워크는 실패를 명확하게 알려 주지 않는다. 응답이 사라졌을 때 클라이언트가 재시도할 수 있어야 사용자는 결제 완료 화면을 회복할 수 있다. 하지만 같은 retry가 서버에서는 두 번째 결제 생성이 될 수 있다. 멱등키는 이 양쪽 요구, 즉 retry 허용과 부작용 중복 차단을 함께 만족하기 위해 필요해졌다.

이 배경을 놓치면 멱등키를 단순 중복 제출 방지로 축소한다. 실제 목적은 사용자가 다시 누르거나 클라이언트가 자동 재시도해도 서버가 같은 업무 의미를 한 번만 실행하고, 이미 실행한 결과를 다시 설명할 수 있게 만드는 것이다.

#### DU44-S2 키 범위는 operation scope와 묶인다

멱등키는 전역 문자열 하나로 충분하지 않을 수 있다. 같은 키라도 merchant, operation type, account, endpoint가 다르면 다른 의미일 수 있다. 따라서 unique key는 보통 `(merchant_id, operation, idempotency_key)`처럼 업무 범위를 포함해야 한다.

함정은 클라이언트가 UUID를 보낸다는 이유로 scope를 생략하는 것이다. 운영 수동 재처리나 테스트 도구, 파트너 버그에서 키 충돌은 일어난다. scope가 없으면 한 가맹점의 키가 다른 가맹점 요청을 막는 황당한 장애도 가능하다.

#### DU44-S3 canonical hash는 정규화가 먼저다

request hash를 만들 때 JSON 문자열을 그대로 hash하면 필드 순서, 공백, 기본값 누락 때문에 같은 의미가 다른 hash가 될 수 있다. 먼저 canonical form을 정해야 한다. 금액은 minor unit과 currency로, 고객 id는 내부 canonical id로, optional field는 명시적 기본값으로 정규화한 뒤 hash한다.

이 정규화가 없으면 replay되어야 할 재시도가 conflict로 보이거나, 반대로 다른 의미가 같은 문자열 일부만 보고 같다고 판단될 수 있다. hash는 보안 장식이 아니라 의미 동일성 판정이므로 canonicalization이 핵심이다.

#### DU44-S4 pending timeout은 두 단계로 본다

PENDING이 오래 남았다고 바로 실패는 아니다. 먼저 winner process가 죽었는지, 외부 호출이 나갔는지, external reference가 저장됐는지, webhook이 도착했는지 확인해야 한다. external reference가 있으면 조회로 복구하고, reference가 전혀 없고 외부 호출 전 죽었다는 증거가 있으면 안전 재시도 후보가 된다.

함정은 pending age만 보고 row를 삭제하는 것이다. 삭제 후 같은 키가 들어오면 새 winner가 생기고, 첫 winner가 사실 외부 성공했더라면 중복 부작용이 된다. pending cleanup은 상태 증거 없이 삭제가 아니라 recovery workflow로 가야 한다.

#### DU44-S5 replay payload는 최소 안정 표현을 저장한다

전체 HTTP 응답을 그대로 저장하면 구현은 쉽지만 serializer 변경, 개인정보 보존, schema version 문제가 생긴다. 반대로 external id만 저장하면 클라이언트가 첫 응답을 잃었을 때 같은 결과를 얻기 어렵다. 그래서 replay payload는 versioned minimal response로 설계하는 편이 안전하다.

운영 함정은 response table을 캐시처럼 보고 TTL로 쉽게 지우는 것이다. replay payload는 장애 복구 계약이다. 보존 기간과 masking, schema migration을 API 계약처럼 다뤄야 한다.

#### DU44-S6 외부 idempotency와 내부 idempotency는 서로 보완한다

PG나 외부 API가 idempotency key를 지원하면 중복 부작용 위험은 줄어든다. 하지만 내부 주문 상태, 원장, 응답 replay, 대사 연결은 여전히 내부 저장소가 알아야 한다. 외부 키만 믿으면 내부가 어떤 요청을 보냈는지 재구성할 수 없다.

반대로 내부 멱등만 있고 외부에 키를 보내지 않으면 internal retry 중 외부 호출이 중복될 수 있다. 좋은 설계는 내부 key와 외부 key, external id를 한 row에 묶어 양쪽 재조회가 가능하게 한다.

#### DU44-S7 동시성 테스트는 의도적으로 경쟁 창을 만든다

멱등 처리는 일반 단위 테스트로 잘 깨지지 않는다. 두 thread가 동시에 `find none`을 보고 action으로 진입하도록 latch를 걸어야 실패가 재현된다. 로컬 seed 카드가 이런 race lab을 둔 이유도 우연한 동시성에 기대지 않고 실패 창을 확정적으로 열기 위해서다.

테스트 PASS 조건은 response가 둘 다 성공처럼 보이는 것이 아니다. 외부 action counter가 1이고, loser가 conflict/replay/in-progress 중 올바른 분기로 갔으며, DB에는 같은 key의 row가 하나만 있어야 한다.

#### DU44-S8 운영 metric은 retry 정상성과 버그를 구분한다

duplicate count가 늘었다고 항상 나쁜 것은 아니다. 네트워크 장애 동안 retry가 늘면 replay count가 증가할 수 있다. 문제는 conflict count, pending age, unknown age, external duplicate id, hash mismatch가 같이 늘어나는 경우다.

따라서 dashboard는 `duplicate_total` 하나가 아니라 replay/conflict/in_progress/unknown/recovered를 나눠야 한다. 이 분류가 있어야 클라이언트 버그, PG 지연, 내부 worker 장애를 구분한다.

#### DU44-S9 보안과 악용 경계

멱등키는 공격자가 리소스를 점유하는 입력이 될 수도 있다. 긴 TTL과 큰 response payload를 무제한 허용하면 저장소가 커진다. 키 길이, scope, rate limit, payload size, 인증 주체와의 binding이 필요하다.

함정은 멱등키를 편의 기능으로만 보고 인증 주체와 분리하는 것이다. 다른 사용자의 key를 추측해 replay 결과를 얻을 수 있으면 정보 노출이 된다. idempotency row는 반드시 owner scope 안에서 조회되어야 한다.

#### DU44-S10 최종 설계 질문

중복 요청 설계의 마지막 질문은 같은 요청이 두 번 들어왔을 때 무엇을 하지 않을 것인가다. 새 외부 호출을 하지 않고, 다른 의미를 같은 결과로 replay하지 않고, unknown을 실패로 단정하지 않는 것이 핵심이다.

이 부정형 불변식이 닫히면 구현 선택은 여러 가지가 가능하다. DB unique insert, advisory lock, external key, inbox table을 쓸 수 있지만, 외부 부작용 전 winner election과 의미 충돌 탐지는 빠지면 안 된다.

### DU44 추가 실전 사례: 결제 완료 화면 복구

사용자가 결제 버튼을 누른 뒤 브라우저가 응답을 받기 전에 닫혔다고 가정하자. 서버가 이미 PG 승인을 받았고 내부 DB에도 성공을 기록했다면, 다음 앱 실행에서 같은 idempotency key로 재조회하거나 같은 주문 id로 상태를 조회했을 때 결제 완료 화면을 복구해야 한다. 여기서 replay는 개발 편의가 아니라 사용자 경험과 회계 일관성을 동시에 지키는 장치다. replay가 없으면 사용자는 다시 결제를 시도하고, 고객센터는 PG 승인 내역과 내부 주문 상태를 손으로 맞춰야 한다.

반대로 첫 요청이 PG까지 도달하지 않았다는 증거가 있다면 같은 key로 새 winner를 만들 수 있는가를 따져야 한다. 이 결정은 pending row 삭제가 아니라 상태 전이로 남아야 한다. 예를 들어 `PENDING -> EXPIRED_BEFORE_EXTERNAL_CALL` 같은 상태는 새 실행이 안전했던 이유를 남기고, `UNKNOWN_EXTERNAL_SUBMITTED`는 재조회 없이는 새 실행을 막는다. 이 두 상태를 하나의 `FAILED`로 뭉개면 운영자는 어떤 실패가 안전 재시도인지 구분하지 못한다.

```text
lost response recovery
  stored SUCCESS + same hash -> return saved response, no PG call
  stored PENDING + no external_ref + worker crash proof -> expire and allow controlled retry
  stored UNKNOWN + external_ref exists -> retrieve external status first
  stored SUCCESS + different hash -> reject as key conflict
```

이 사례의 검증은 두 단계다. 먼저 통합 테스트에서 응답 직전 timeout을 강제로 만들고 duplicate 요청이 저장 응답을 받는지 확인한다. 다음으로 운영 로그에서 같은 key의 외부 승인번호가 하나만 있는지, replay count가 증가했는지, conflict가 따로 분류되는지 확인한다.

### DU44 운영 케이스 매트릭스

DU44의 본문을 실제 장애 대응에 쓰려면 `duplicate request and idempotency key`를 하나의 정의가 아니라 반복해서 판정할 수 있는 케이스 묶음으로 바꿔야 한다. 아래 케이스들은 서로 다른 상황을 다루며, 각 케이스는 입력, 상태 변화, 실패 신호, 관측 지점을 함께 남긴다.

| 케이스 | 입력 또는 상황 | 안전한 판정 | 관측/검증 지점 |
| --- | --- | --- | --- |
| lost response | 성공 응답 유실 | stored success replay | replay count |
| hash conflict | 같은 key 다른 amount | 409 conflict without PG call | conflict metric |
| old pending | 오래된 pending | external ref 기준 복구 | pending age |
| external key mismatch | 내부 key와 외부 key 다름 | 키 매핑 실패로 차단 | external id audit |
| TTL expiry | 늦은 retry가 TTL 뒤 도착 | 보존 정책 검토 | expired key log |
| multi tenant key | 다른 merchant 같은 key | scope 포함 unique | tenant scoped index |
| payload version | 응답 schema 변경 | versioned replay | response version |
| abuse | 무작위 key 대량 생성 | rate limit and owner binding | key creation metric |

### DU44 면접식 꼬리 질문과 실전 답변

DU44를 면접이나 설계 리뷰에서 설명할 때는 정의보다 반례 대응이 중요하다. 아래 질문들은 본문을 읽은 사람이 실제로 다시 설명할 수 있는지 확인하기 위한 압축 검증이다.

#### Q1. UI debounce와 서버 멱등키의 차이는 무엇인가?

UI debounce는 사용자의 빠른 연속 클릭을 줄이는 편의 장치다. 서버 멱등키는 네트워크 재시도, 앱 재실행, 로드밸런서 timeout, 메시지 재배달까지 포함해 같은 업무 의미의 부작용을 한 번으로 제한하는 저장소 계약이다.

함정은 버튼을 비활성화했으니 중복 결제가 불가능하다고 믿는 것이다.

#### Q2. 왜 INSERT PENDING이 외부 호출보다 먼저인가?

중복 부작용은 외부 호출에서 생긴다. DB unique insert가 외부 호출 뒤에 있으면 row 중복은 막아도 PG 호출 두 번은 이미 나갔다. 따라서 winner election은 외부 호출 이전에 닫혀야 한다.

함정은 DB row가 하나라서 결제도 하나라고 추론하는 것이다.

#### Q3. request hash는 왜 필요한가?

idempotency key는 재시도 묶음이고 요청 의미 자체가 아니다. 같은 key로 amount나 receiver가 바뀌면 재시도가 아니라 충돌이다. hash가 있어야 같은 key의 같은 의미와 다른 의미를 구분한다.

함정은 클라이언트가 key를 절대 재사용하지 않을 것이라고 믿는 것이다.

#### Q4. timeout 후 duplicate가 오면 무엇을 반환해야 하는가?

stored status가 SUCCESS이면 replay하고, PENDING이면 진행 중으로 알리거나 기다리며, UNKNOWN이면 외부 조회를 먼저 해야 한다. 바로 새 외부 호출을 열면 중복 부작용 위험이 커진다.

함정은 timeout을 실패로 확정해 새 결제를 허용하는 것이다.

#### Q5. 멱등키는 얼마나 보관해야 하는가?

클라이언트 retry 기간, 외부 조회 가능 기간, 정산/대사 주기, 개인정보 보존 부담을 함께 보고 정한다. 너무 짧으면 늦은 retry가 새 실행이 되고, 너무 길면 저장/보안 부담이 커진다.

함정은 cleanup을 단순 용량 관리로만 보고 상태별 위험을 나누지 않는 것이다.

### DU44 마지막 close scenario: 같은 키, 다른 금액, 같은 사용자

사용자 U가 주문 O-1에 대해 `K-1`로 10,000원 결제를 보냈고, 네트워크 실패 후 앱 버그 때문에 같은 `K-1`로 12,000원 결제를 다시 보냈다고 하자. 서버가 key만 보고 replay하면 사용자는 12,000원 결제를 요청했는데 10,000원 성공 응답을 받는다. 반대로 서버가 금액 차이를 감지하면 외부 호출 없이 충돌을 반환하고, 클라이언트는 새 idempotency key로 새 주문 의미를 제출해야 한다.

```text
stored row: key=K-1, hash=H(amount=10000,currency=KRW,order=O-1), status=SUCCESS
incoming:   key=K-1, hash=H(amount=12000,currency=KRW,order=O-1)
server:     hash mismatch -> CONFLICT -> no PG call -> no replay
```

이 시나리오는 request hash가 왜 '선택적 방어'가 아닌지 보여 준다. 멱등은 같은 요청을 여러 번 보내도 같은 효과가 나야 한다는 계약이지, 같은 키를 가진 모든 요청을 같은 요청으로 취급한다는 뜻이 아니다. 운영 로그에는 conflict가 정상 replay와 분리되어야 한다. conflict가 늘면 클라이언트 키 생성 버그, 주문 수정 플로우 결함, 파트너 재시도 규약 위반을 조사해야 한다.

### DU44 최종 보강: 같은 응답을 돌려준다는 말의 정확한 뜻

replay는 매번 현재 상태를 새로 계산해 비슷한 응답을 만드는 것이 아니다. 첫 성공 시점에 클라이언트가 받았어야 할 안정 응답을 저장하고, 같은 의미의 재시도에 외부 호출 없이 돌려주는 것이다. 물론 개인정보나 큰 payload를 그대로 저장할 수 없다면 canonical response를 따로 정의해야 한다. 중요한 것은 중복 요청이 새로운 결제 생성 경로로 들어가지 않는다는 점이다. 이 기준이 닫히면 UI, 모바일 네트워크, 서버 retry, 운영 수동 재처리가 모두 같은 안전 규칙을 공유한다.

### DU44 source boundary note

RFC 9110은 HTTP method의 멱등 의미를 말하지만 결제 POST의 저장소 구현을 대신 정해 주지는 않는다. Idempotency-Key draft도 헤더의 방향을 제시할 뿐, winner election, request hash, replay payload, unknown 복구는 서버가 자기 DB와 외부 PG 계약에 맞게 구현해야 한다.

### DU44 최종 보강: idempotency는 중복 실행 방지가 아니라 결과 소유권 설계다

따닥 요청을 막는 가장 약한 방법은 버튼을 잠그는 것이다. 버튼 잠금은 사용자 경험에는 도움이 되지만, 네트워크 재전송, 모바일 앱 재시도, 브라우저 refresh, worker retry, webhook 중복 전달을 막지 못한다. 그래서 idempotency key는 UI 방어가 아니라 서버가 같은 operation의 소유권을 정하는 장치다. 첫 요청이 winner가 되어 처리 권한을 얻고, 뒤따라온 요청은 새로 실행하지 않고 winner의 상태나 저장된 응답을 읽어야 한다.

실무에서 중요한 질문은 같은 키인가보다 같은 의도인가다. 같은 idempotency key로 금액이나 수취인이 다른 요청이 들어오면 replay가 아니라 충돌로 보아야 한다. 이를 위해 request hash나 canonical payload digest를 저장한다. 같은 키와 같은 hash면 진행 중 상태를 반환하거나 최종 응답을 replay하고, 같은 키와 다른 hash면 명확한 오류로 닫는다. 이 경계를 두지 않으면 공격자나 버그가 같은 키를 이용해 다른 결제를 덮어쓸 수 있다.

검증은 세 번의 요청으로 충분하다. 첫 요청은 row insert에 성공해야 한다. 두 번째 요청은 같은 key와 같은 payload로 들어와 같은 response를 받아야 한다. 세 번째 요청은 같은 key와 다른 amount로 들어와 새 결제를 만들지 않고 충돌로 실패해야 한다. 이 세 결과가 모두 DB row와 로그에서 재현되면 idempotency는 단순 중복 방지가 아니라 결과 소유권 설계로 닫힌다.

운영에서는 TTL도 함께 정해야 한다. idempotency row를 영원히 보관하면 저장 비용과 개인정보 보존 문제가 커지고, 너무 빨리 지우면 느린 재시도나 webhook 재전송이 새 실행으로 바뀔 수 있다. TTL은 업무 위험과 외부 시스템 재시도 기간을 기준으로 잡아야 한다. 결제라면 승인, 취소, 정산, 고객 화면 복구까지 고려해야 하므로 단순 cache 만료 시간처럼 다루면 안 된다.

마지막 검증 질문은 “같은 요청을 다시 받았을 때 외부 PG를 다시 호출하지 않는가”이다. DB winner row가 있어도 코드가 replay 전에 외부 호출을 먼저 하면 멱등성은 깨진다. 로그에는 winner election, request hash 비교, replay source, external call 여부가 분리되어 남아야 한다.

### 추가 판정 질문: replay 가능한 응답은 무엇인가

idempotency를 구현할 때 모든 응답을 그대로 저장할 필요는 없지만, 무엇을 replay할지는 미리 정해야 한다. 결제 성공 응답에는 payment id, status, 승인 시각, 표시할 메시지가 필요할 수 있고, 처리 중 응답에는 polling 위치나 request id가 필요할 수 있다. 반대로 민감한 카드 정보나 과도한 내부 오류 stack은 저장하면 안 된다.

따라서 replay payload는 API 설계의 일부다. 첫 요청이 성공했지만 응답이 클라이언트에 닿지 못했을 때, 두 번째 요청은 사용자가 같은 결과를 안전하게 이어갈 수 있을 만큼 충분한 응답을 받아야 한다. 이 기준이 없으면 멱등 row는 있어도 사용자 경험은 깨진다.

실무에서는 replay payload를 저장 응답 전체와 동일시하지 않는 편이 안전하다. 응답 전체를 저장하면 빠르게 구현할 수 있지만, 개인정보, 내부 오류 문구, serializer 변화, API version 변화가 모두 저장소 계약 안으로 들어온다. 반대로 payment id만 저장하면 사용자는 첫 요청에서 받아야 했던 상태와 안내 문구를 복구하지 못한다. 그래서 보통은 `response_version`, `public_status`, `external_ref`, `display_message_code`, `created_at`처럼 재시도 사용자가 이어서 행동하는 데 필요한 안정 필드만 남긴다.

이 설계가 있으면 장애 대응도 분명해진다. 중복 요청이 들어왔을 때 서버는 "지금 다시 계산한 최신 응답"이 아니라 "처음 성공한 실행의 공개 가능한 결과"를 돌려준다. 외부 상태가 이후에 바뀌었다면 별도 조회 API나 상태 갱신 이벤트가 그 변화를 설명해야 한다. replay 응답이 최신 상태 조회와 섞이면 사용자는 같은 키로 요청할 때마다 다른 결과를 볼 수 있고, 멱등성은 중복 실행 방지에서 결과 재현성까지 확장되지 못한다.

### DU44 운영 복구 훈련: 성공 응답이 사라진 결제

가장 현실적인 따닥 이슈는 사용자가 버튼을 두 번 누른 장면보다, 첫 요청은 성공했지만 응답이 사라진 장면이다. 모바일 네트워크가 끊기거나 로드밸런서 timeout이 나면 사용자는 실패로 느낀다. 하지만 서버는 이미 PG 승인을 받았고 DB에 성공 상태를 저장했을 수 있다. 이때 두 번째 요청을 새 결제로 실행하면 사용자는 같은 주문에 두 번 결제된다. 그래서 idempotency의 중심은 "두 번 눌렀나"가 아니라 "첫 실행의 결과를 다시 찾을 수 있나"다.

```text
request K-900, hash H1
  t1 insert idempotency row: PENDING
  t2 call PG with external key K-900
  t3 PG success auth=A-77
  t4 update row: SUCCESS, response_version=1, payment_id=P-10, auth=A-77
  t5 response lost before client receives it

retry K-900, hash H1
  find SUCCESS row
  no PG call
  return stored public response for P-10
```

이 trace에서 안전을 만드는 지점은 t4다. 성공 응답이 클라이언트에 닿았는지와 상관없이, 서버는 첫 실행의 공개 가능한 결과를 저장했다. 따라서 retry는 새 실행이 아니라 조회와 replay가 된다. 반대로 t3 뒤 t4 전에 서버가 죽었다면 상태는 더 까다롭다. PG에는 성공이 있을 수 있지만 내부 row는 PENDING일 수 있다. 이때는 같은 external key로 PG를 조회하거나 webhook/정산 자료를 기다려야 하며, row를 지우고 새 winner를 만들면 안 된다.

멱등키 설계는 request hash와 external reference를 함께 가져야 한다. request hash는 같은 key가 같은 업무 의미인지 확인하고, external reference는 외부 세계에서 이미 일어난 부작용을 다시 찾게 해 준다. 둘 중 하나만 있으면 복구가 약해진다. hash만 있으면 외부 성공 여부를 찾기 어렵고, external reference만 있으면 같은 key로 다른 금액을 보낸 오류를 놓칠 수 있다.

운영 화면도 이 차이를 보여 줘야 한다. `SUCCESS_REPLAYABLE`, `PENDING_NO_EXTERNAL_REF`, `UNKNOWN_EXTERNAL_SUBMITTED`, `CONFLICT_HASH_MISMATCH`는 모두 다른 다음 행동을 가진다. 첫 상태는 replay하면 되고, 두 번째는 안전 재시도 후보가 될 수 있으며, 세 번째는 외부 조회가 먼저이고, 네 번째는 새 key로 새 업무 의미를 제출해야 한다. 이 상태 언어가 없으면 운영자는 pending row를 삭제하거나 수동 재처리 버튼을 눌러 중복 부작용을 만들 수 있다.

### DU44 마지막 운영 연습: 두 요청이 동시에 winner가 되려는 순간

동시 따닥 요청을 제대로 이해하려면 `SELECT 후 없으면 INSERT` 패턴을 의심해야 한다. 두 요청이 거의 동시에 들어오면 둘 다 row가 없다고 보고 외부 호출로 나갈 수 있다. 안전한 구현은 조회가 아니라 원자적 선출을 먼저 해야 한다. DB unique constraint가 있는 insert, 조건부 update, advisory lock 같은 방식 중 무엇을 쓰든, 외부 PG 호출 전에 오직 하나의 요청만 처리권을 가져야 한다.

```text
bad race
  request A: SELECT key K -> none
  request B: SELECT key K -> none
  request A: call PG
  request B: call PG
  request A/B: one DB row may survive, but two external effects already happened

safe race
  request A: INSERT key K PENDING -> success, winner
  request B: INSERT key K PENDING -> unique conflict, loser
  request A: call PG and store result
  request B: read winner row and replay/conflict/wait
```

이 차이는 DB row 수만 보면 놓칠 수 있다. 나중에 unique key 때문에 row가 하나만 남아도 외부 PG 호출은 이미 두 번 나갔을 수 있다. 그래서 idempotency 검증에는 DB row count뿐 아니라 외부 action counter가 반드시 들어가야 한다. 결제라면 PG mock의 승인 호출 수, 펌뱅킹이라면 전문 발송 수, 메시지라면 publish count를 함께 봐야 한다.

운영 알람도 분리한다. replay count 증가는 네트워크 장애 중 정상 회복 신호일 수 있다. hash conflict 증가는 클라이언트 버그나 파트너 재시도 규약 위반 신호다. pending age 증가는 worker crash나 외부 조회 지연 신호다. 이 세 값을 하나의 duplicate count로 합치면 정상 복구와 위험 신호를 구분하지 못한다. 멱등성은 코드 패턴이 아니라 관측 분류까지 포함하는 운영 계약이다.
