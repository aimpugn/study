# WORK_20260917_CONCURRENCY_TRANSACTION_GO_REWORK

> SSOT는 `AGENTS.md`입니다. 템플릿의 빈 항목은 이 작업의 판단에 쓰이지 않아 생략했습니다.

## 0. Meta

- 작업 제목: 동시성 예제 리뷰 피드백 반영, Go 구현 이동, 줄 단위 설명 재작성
- 작업 유형: `audit | execute | explain`
- 작업 깊이: `full`
- 원문 사용자 요청: "여러 언어로 구현해볼 수 있으니, go로 구현된 것을 적당한 디렉토리로 옮기고, 방금 피드백 내용을 반영한 완성본을 만들어 주세요. 그리고 학습이 중요하므로, 라인 바이 라인으로 핵심을 설명해주세요. /natural-korean-writing 스킬 등을 사용하여 글쓰기 전략을 잘 구성하여 전체적으로 코드와 설명이 잘 이해가 되도록 개선/수정해주세요."
- 선행 작업: `docs/works/WORK_20260915_TRANSACTION_MICROSECOND_CONCURRENCY_UNIQUENESS.md` (이 작업이 대체)
- 대상 자산:
  - `database/examples/concurrency_transaction/README.md` (신규, 언어 공통 시나리오 계약)
  - `database/examples/concurrency_transaction/go/` (루트의 Go 파일을 대체)
  - `database/concurrent_payment_integrity.md` (`database/microsecond_concurrency_transaction_uniqueness.md`를 대체)
- 현재 상태: `COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 리뷰에서 확인한 코드 결함과 문서 오류를 모두 고친 완성본을 만들고, 코드 곁에서 줄 단위로 핵심을 설명한다.
- must_keep:
  - 외부 DB 없이 `go test`로 실행 가능
  - 안티패턴과 해법을 짝지어 비교하는 구조
  - 리뷰 피드백 전 항목 반영 (결함 1-1~1-4, 문서 오류 2-1~2-4, 테스트 약점 3-1~3-3, 누락 주제 4절, 저장소 계약 5절)
- 구조 결정:
  - 다른 언어 구현을 나란히 둘 수 있게 `concurrency_transaction/go/`로 옮김. 저장소의 기존 다언어 배치(`computer_architecture/threads/examples/{go,java,kotlin,rust}`)와 같은 형태.
  - 언어가 바뀌어도 유지할 계약은 상위 `README.md`에 시나리오 표로 둠.
  - 줄 단위 설명은 소스 주석에 둠(study-explanation의 code-first 규칙, AGENTS 9.1). README는 실행과 읽는 순서만, 개념 문서는 원리와 엔진 차이와 트레이스를 맡음.
- Non-goals: PostgreSQL, MySQL 컨테이너로 실제 실행. 이번에는 공식 문서 근거로 적고 "실행하지 않음"을 명시함. 다른 언어 구현은 계약만 두고 만들지 않음.

## 2. Root-First Framing

- 근본 문제: 첫 버전은 "1µs 동시 도착"을 원인으로 설명했지만, 실제 원인은 확인과 쓰기 사이의 틈이다. 이 오해가 문서 전체의 설명 축을 틀어지게 했고, 테스트는 결과 숫자만 보아 락이 없는 구현을 통과시켰다.
- 성공 정의:
  - 알려진 버그를 되살리면 테스트가 실패한다(뮤테이션 확인).
  - 문서의 DB 동작 주장은 실행 결과나 공식 문서 인용으로 뒷받침되고, 실행하지 않은 것은 그렇다고 밝힌다.
  - 독자가 갱신 손실, 쓰기 스큐, 네 가지 대응, 선점과 UNKNOWN과 대사를 자기 말로 설명할 수 있다.

## 3. Reader & Internalization Contract

- 주 독자: DB 동시성을 처음 체계적으로 정리하는 개발자, 결제 도메인 면접 준비자
- teach-back 목표: "경쟁 조건은 틈의 문제다"에서 출발해 네 가지 대응과 외부 호출 시 상태 기계까지 이어서 설명하기
- 막아야 하는 오해와 대조쌍:
  - 동시 도착 vs 틈의 길이 -> 100ms 늦게 와도 이중 승인되는 시간표
  - 트랜잭션 vs 락 -> SQLite는 BUSY, PostgreSQL READ COMMITTED는 조용한 갱신 손실
  - 갱신 손실 vs 쓰기 스큐 -> 같은 행 덮어쓰기 vs 서로 다른 새 행 추가
  - 타임아웃 vs 실패 -> UNKNOWN과 대사
  - 래치 경합 vs 트랜잭션 대기 기반 유일성 판정
- 첫 번째 벽돌: 잔액 20,000원에서 5,000원 두 건의 시간표(틀린 결과가 정상과 구별되는 금액)
- active recall / replay: 복습 질문 5개, 버그를 되살려 테스트 실패를 보는 표 9행
- exemplar:
  - 주석 기준: `computer_architecture/network/l4_failover_lab/src/probe.rs`의 `run` (단계마다 목적 문장을 코드 바로 위에 둠). 그대로 따르지 않은 점: Go 문서 주석은 Rust의 `# Arguments` 절 대신 산문으로 반환값과 오류를 적음(Go 관례).
  - 문서 기준: `database/mvcc.md`의 깊이(트레이스, 엔진별 흐름 비교). 그대로 따르지 않은 점: 역사 서술 비중. 이 주제는 역사보다 실패 트레이스가 이해를 만든다.

## 4. Evidence Ledger

- E-01 SQLite 기본 BEGIN은 DEFERRED, 읽기 트랜잭션은 쓰기로 올라가지 못하면 `SQLITE_BUSY`: [sqlite.org/lang_transaction](https://www.sqlite.org/lang_transaction.html)
- E-02 `SQLITE_BUSY_SNAPSHOT`(517), BEGIN IMMEDIATE 성공 뒤 BUSY 없음, `SQLITE_CONSTRAINT_UNIQUE`(2067): [sqlite.org/rescode](https://www.sqlite.org/rescode.html)
- E-03 PostgreSQL READ COMMITTED의 WHERE 재평가, REPEATABLE READ와 SERIALIZABLE의 40001과 재시도: [postgresql.org transaction-iso](https://www.postgresql.org/docs/current/transaction-iso.html)
- E-04 PostgreSQL UNIQUE 충돌 시 커밋 전 트랜잭션을 기다림: [postgresql.org index-unique-checks](https://www.postgresql.org/docs/current/index-unique-checks.html)
- E-05 InnoDB 중복 키 시 공유 락과 세 세션 데드락: [dev.mysql.com innodb-locks-set](https://dev.mysql.com/doc/refman/8.4/en/innodb-locks-set.html)
- E-06 InnoDB 스냅숏은 SELECT에 적용되고 DML에는 반드시 적용되지는 않음: [dev.mysql.com innodb-consistent-read](https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html)
- E-07 실행 결과: 이전 구현의 비관적 락은 `db.Begin()`(DEFERRED)이라 5건 중 2건이 `database is locked`(517/5)로 실패, `_txlock=immediate`면 성공 3, 잔액 부족 2, 오류 0 (scratchpad 재현)
- E-08 실행 결과: 새 구현 `go test -race -count=20` 통과, `revive [rule.exported]` 통과, `go vet` 통과
- E-09 뮤테이션: 1차 7종 모두 검출, 최종 10종 모두 검출(7.5절)
- E-10 SQLite busy handler는 교착 가능성이 있으면 호출되지 않음: [sqlite.org/c3ref/busy_handler](https://www.sqlite.org/c3ref/busy_handler.html)
- 추론으로 남는 것: PostgreSQL, MySQL에서 같은 Go 코드를 실행했을 때의 증상은 E-03, E-05, E-06에서 따라 나온 결론이며 실행하지 않음

## 5. Decisions and Rejected Alternatives

- 비관적 락 구현: DSN `_txlock=immediate` 대신 고정 연결에 `BEGIN IMMEDIATE`를 직접 보냄. DSN 방식은 코드에서 락을 잡는 줄이 보이지 않아 학습 자산으로 약함. 롤백 실패 시 `driver.ErrBadConn`으로 연결을 버려 열린 트랜잭션이 풀로 돌아가지 않게 함.
- 겹침 강제: 원래의 `WaitGroup` 배리어는 한쪽이 도착하지 못하면 테스트가 영원히 멈춤. 대기 상한이 있는 `Rendezvous`로 바꾸고, 상한 초과(`Missed`)를 "락이 상대를 막았다"는 증거로 사용.
- 중복 요청의 패자 처리: 15ms 폴링 후 `ErrDuplicateRequest` 대신 저장된 상태를 그대로 반환. 임의의 폴링 시간이 사라지고 호출자가 상태로 판단함.
- 문서 주석 형식: `revive`는 `이름 + 공백` 시작을 요구함. 영어 첫 문장(한국어 우선 규칙과 충돌), `이름 는`(어색함)을 버리고 `OpenDB 함수는`, `ErrKeyReused 오류는` 형식을 채택.
- 제거: 하드웨어 지연 표와 MESI 절(결론을 뒷받침하지 않고 일부는 틀림), "100%" 표현, Redlock을 결제 보호 수단으로 평가한 행.

## 6. Verification

- `go vet ./...`: PASS
- `gofmt -l .`: 출력 없음
- `go test -race -count=20 ./...`: PASS
- `revive -set_exit_status` (`[rule.exported]`): PASS
- 뮤테이션: 7.3절, 7.5절
- 독립 검토: 작성에 관여하지 않은 에이전트가 코드, DB 주장, 문서 인용과 실제 출력 일치, 한국어 가독성을 검토함. 지적과 조치는 7절.

## 7. Independent Review and User Feedback

### 7.1 1차 독립 검토 (작성에 관여하지 않은 에이전트)

검토자는 인용 출력 일치, 뮤테이션 7종 재현, race 반복을 직접 확인했고 MAJOR 5건, MINOR 13건을 지적했습니다.

| 지적 | 조치 |
| --- | --- |
| stuckAfter만 길게 잡아도 대사가 안전하지 않음 | `ApproveTimeout`으로 PG 대기 상한, DB 시계(`dbNowMillis`)로 경과 측정, 기록은 `WithoutCancel`+`recordTimeout`. 선점과 PG 호출 사이의 정지는 시간으로 못 막고 PG 쪽 멱등성이 막아야 한다는 한계, `ErrStateChanged` 뒤 환불과 경보 미구현을 문서와 주석에 명시 |
| "도착 간격이 아니라 틈의 길이" 과잉 교정 | "B의 확인이 A의 확인과 쓰기 사이에 들어오면"으로 조건을 다시 씀 |
| "존재하지 않는 행은 잠글 수 없다" 일반화 | "기존 행에만 거는 행 락은 새 행 삽입을 못 막는다"로 좁히고, 부모 행 락, gap 락, DB 락, SERIALIZABLE을 대가와 함께 나열 |
| "READ COMMITTED에서만 0행" 과대 범위 | PostgreSQL 한정으로 좁힘 |
| SQLite 오류 설명이 실제 경로(BUSY 5)와 불일치 | BUSY와 BUSY_SNAPSHOT 두 경우를 모두 설명하고, busy handler를 부르지 않는 이유를 공식 문서로 인용 |
| race 반복이 타이밍 독립성을 증명한다는 과장 | "들쭉날쭉하지 않은지 보는 확인"으로 낮추고 300ms 상한 의존을 한계에 명시 |
| 느린 테스트 이름 오기 | 실제 출력 기준으로 수정 |
| 계약 README가 구현보다 넓게 주장 | "Go 확인 방법" 열(재현 테스트 / 변이 확인) 추가 |
| 고객 없음이 한도 초과로 보고됨 | `ErrCustomerNotFound` 추가, 테스트 추가 |
| 기한 초과 시 PENDING으로 남음 | 기록 문맥 분리, `TestPayOnce_CallerDeadlineDuringApprovalIsUnknown` 추가 |
| MySQL 중복 키 데드락 누락 | `claim` 주석과 문서에 1213 재시도 명시 |
| P1 테스트 약함 | 모든 결과가 같은 기록인지 확인, 다른 계좌 경우 추가, 고정 worker 번호 서술 제거 |
| `ChargeReserved` 원자성 미검증 | 트리거로 INSERT 실패를 주입하는 테스트 추가 |
| MySQL 인용 과장, 기본 격리 수준 누락 | "반드시 적용되지는 않는다", "기본 격리 수준에서"로 수정 |
| "틈이 생길 자리가 없다" 과장 | "같은 키로 PG를 두 번 부르는 틈은 없다"로 좁히고 남는 구간을 대사로 연결 |
| 용어 미정의 (PG, Check-then-Act, 스냅숏 등) | 첫 등장에 정의, P1 시간표 추가, 제목 정리 |
| 개정 이력식 서술, 반복 문장 | 일반 실패 사례로 바꾸고 제거 |
| 주석 계약(9.1) 누락과 중복 | 오류 계약 추가, 본문 주석 보강, 중복 제거 |

### 7.2 사용자 피드백 (작업 중 수신)

- "주의 깊게 봐야 하는 곳은 강조": 본문에 `주의:`(버그가 생기는 자리)와 `핵심:`(막는 줄) 두 종류의 ASCII 상자만 두고, 함수마다 한두 곳으로 제한함. AGENTS 9.1의 "라벨 조각 금지"와 충돌하지 않도록 상자 안은 완결된 문장으로 씀. 읽는 법은 `go/README.md`에 설명함.
- "사진이 무슨 말인가요": 정의 없는 비유였음. 애플리케이션 변수는 "복사본", DB가 트랜잭션에 고정해 보여 주는 상태는 "스냅숏"으로 나누고 각각 정의함.
- "무엇이 무엇인지 명확하게": 본문 주석과 문서의 개념어마다 변수나 컬럼을 붙임(`잔액 복사본(balance)`, `출금액(amount)`, `accounts.balance`). 사용자 예시는 `amount`와 `balance`의 대응이 반대였으며, 이는 원래 주석의 모호함이 실제 오해를 만든다는 증거로 기록함.

### 7.3 2차 검증

- 결과 기록에 요청 ctx를 쓰는 복사본은 처음에 컴파일 오류로 실패해 검출로 세지 않았고, 다시 만들어 `record UNKNOWN: context deadline exceeded`로 검출됨을 확인함.
- 내역 INSERT만 트랜잭션 밖으로 뺀 복사본을 처음에는 롤백 테스트 하나로만 돌려 "살아남음"으로 잘못 판단함. 재검토 지적대로 `TestChargeReserved_KeepsLimit`이 잡으며, 롤백 테스트는 트리거가 아니라 락 대기 실패로 통과한 것이었음. 문서에 "엉뚱한 이유로 통과한 테스트" 사례로 기록함.

### 7.4 좁은 재검토 (같은 검토자, 수정 여부와 회귀만)

- 1차 지적 18건: FIXED 12, BOUNDED 5, PARTIAL 1(#12, 겹침 미강제).
- 새 지적과 조치:
  - N1 (MAJOR) 위 뮤테이션 오판 -> 문서 표에 행 추가, 설명 수정
  - N2 선점 핵심 상자의 인과 오류 -> UNIQUE가 한 건만 통과시키고, INSERT가 PG보다 먼저라 PG도 한 번이라고 수정
  - N3 `claim` 주석이 재시도를 하는 것처럼 읽힘 -> "MySQL로 옮길 때 호출자가 재시도, 이 코드는 SQLite라 하지 않음"
  - N4 WAL 모드에서 기다리지 않는 이유를 롤백 저널 이유로 설명 -> 스냅숏이 낡아 기다려도 소용없다는 WAL 이유로 수정, 추론임을 명시
  - N5 스냅숏 정의가 READ COMMITTED의 문장 단위 스냅숏과 충돌 -> 격리 수준별 차이 추가
  - N6 `stuckAfter` 하한에서 `recordTimeout` 누락 -> `ApproveTimeout + recordTimeout + 여유`
  - N7 강조 상자 불균형 -> `WithdrawInPlainTx` 틈에 주의 상자, `WithdrawLocked`에 핵심 상자
  - #12 -> `holdingGateway`로 선점한 요청의 PG 호출을 나머지 네 요청이 PENDING을 받을 때까지 붙잡아, 매 실행 PENDING 4 / SUCCEEDED 1을 단언

### 7.5 최종 검증

- `gofmt`, `go vet`, `revive [rule.exported]`: PASS
- `go test -race -count=20 ./...`: PASS (21개 테스트)
- 문서 "직접 확인하기" 표의 뮤테이션 10종: 모두 표에 적은 테스트에서 실패(컴파일 실패는 검출로 세지 않음)
- 문서 인용 출력과 실제 `go test -v` 출력 전수 대조: 불일치 0
- 상대 링크: 깨진 링크 0

## 8. Remaining Risks

- PostgreSQL, MySQL 실제 실행으로 문서 주장을 확인하지 않음. 다음 단계 후보: docker compose로 두 DB를 띄워 같은 시나리오 계약을 실행.
- `Rendezvous` 대기 상한 300ms는 매우 느린 CI에서 `Missed` 오판 가능성이 있음. 이 환경에서는 race 모드 20회 반복으로 문제 없음을 확인함.
- `ErrStateChanged` 뒤 환불과 경보, PG 쪽 멱등성 연동은 구현하지 않음.
- `ChargeReserved`의 한 행 모으기 효과는 SQLite 직렬 쓰기와 구분해 확인하지 못함.
