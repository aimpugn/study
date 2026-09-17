# Go 구현: 동시 요청에서 결제와 잔액 지키기

[시나리오 계약](../README.md)을 Go와 SQLite로 구현한 예제입니다. 왜 그렇게 동작하는지는 [개념 문서](../../../concurrent_payment_integrity.md)가 설명하고, 줄 단위 설명은 각 소스 파일의 주석에 있습니다. 이 README는 실행 방법과 읽는 순서만 안내합니다.

## 실행

Go 1.22 이상과 C 컴파일러가 필요합니다. SQLite 드라이버 `github.com/mattn/go-sqlite3`가 cgo를 쓰기 때문입니다.

```bash
cd database/examples/concurrency_transaction/go
go test -v ./...
```

모든 테스트가 `PASS`로 끝나야 합니다. 안티패턴 테스트도 PASS가 정상입니다. 안티패턴 테스트는 "버그가 재현됐는가"를 검사하므로, 버그가 재현되지 않으면 FAIL이 납니다.

결과가 실행할 때마다 달라지지 않는지 보려면 반복 실행합니다. `-race`는 Go 코드의 데이터 경쟁을 함께 검사합니다.

```bash
go test -race -count=20 ./...
```

반복 실행이 통과해도 테스트가 타이밍에 전혀 기대지 않는다는 뜻은 아닙니다. 테스트는 `Missed`가 true인 것을 락이 상대를 막은 증거로 보는데, 기계가 매우 느려 상대가 락과 상관없이 300밀리초 넘게 늦게 도착해도 `Missed`는 true가 되기 때문입니다.

## 주석 읽는 법

소스 주석은 코드 바로 위에서 그 줄이 무엇을 위해 무엇을 하는지 설명합니다. 주석에 나오는 개념어 옆에는 가리키는 변수나 컬럼을 붙입니다. 예를 들어 `잔액 복사본(balance)`, `출금액(amount)`, `accounts.balance`처럼 씁니다. 반드시 멈춰서 봐야 할 자리는 두 가지 상자로 표시합니다.

```go
// ------------------------------------------------------------------------
// 주의: 버그가 생기는 자리입니다. 대부분 확인과 쓰기 사이의 틈입니다.
// ------------------------------------------------------------------------

// ------------------------------------------------------------------------
// 핵심: 그 버그를 막는 줄입니다. 조건부 WHERE, 락, 선점 INSERT 같은 것입니다.
// ------------------------------------------------------------------------
```

상자가 없는 주석도 읽어야 하지만, 처음 읽을 때는 상자만 따라가도 각 함수가 어디서 깨지고 어디서 지켜지는지 보입니다.

## 읽는 순서

1. [`interleave.go`](interleave.go): 테스트가 두 요청을 틈(값을 확인한 뒤 그 결과로 쓰기 전까지의 구간)에서 반드시 만나게 하는 장치입니다. `Pause`, `Rendezvous`, `RunTogether`를 먼저 이해해야 나머지 테스트가 읽힙니다.
2. [`db.go`](db.go): 테이블 구조와 SQLite 오류 판별입니다. `accounts.version`, `payments.idempotency_key`의 UNIQUE, `CHECK` 제약, DB 시각 식 `dbNowMillis`가 뒤에서 쓰입니다.
3. [`balance.go`](balance.go): 잔액 한 행을 두고 겹치는 요청입니다. 깨지는 구현 세 개 다음에 지키는 구현 세 개가 나옵니다.
4. [`daily_limit.go`](daily_limit.go): 규칙이 여러 행에 걸친 경우입니다. 쓰기 스큐(각 요청은 옳게 판단했지만 둘이 합쳐져 여러 행에 걸친 규칙이 깨지는 현상)와, 규칙을 한 행으로 모으는 해법을 다룹니다.
5. [`gateway.go`](gateway.go): 승인, 거절, 두 종류의 타임아웃을 흉내 내는 가짜 PG(결제 대행사)입니다.
6. [`duplicate_request.go`](duplicate_request.go): 같은 결제 요청이 여러 번 오는 경우입니다. 선점(PG 호출 전에 PENDING 행을 넣어 이 키로 PG를 부를 권리를 먼저 가져가기), 요청 내용 비교, 결과 모름(`UNKNOWN`), 대사(PG에 조회해 결과 확정)를 다룹니다.

각 `*_test.go`의 테스트 이름과 문서 주석이 그 테스트가 무엇을 증명하는지 설명합니다.

## 테스트가 확인하는 것

| 테스트 | 확인하는 주장 | `go test -v`에서 볼 값 |
| --- | --- | --- |
| `TestWithdrawReadModifyWrite_LosesUpdate` | 읽고 계산해 덮어쓰면 갱신 손실 | `success=2 balance=15000` |
| `TestServerWithdraw_LocalLockProtectsOnlyOneServer` | 뮤텍스는 서버 한 대 안에서만 유효 | 같은 서버 `missed=true balance=10000`, 두 대 `missed=false balance=15000` |
| `TestWithdrawInPlainTx_TransactionAloneDoesNotLock` | 트랜잭션만 열면 틈이 남음 | 한쪽 `database is locked` |
| `TestWithdrawConditional_RejectsOverdraft` | 조건부 UPDATE에는 틈이 없음 | `success=1 insufficient=4 other=[]` |
| `TestWithdrawLocked_SerializesRequests` | `BEGIN IMMEDIATE`가 상대를 실제로 막음 | `missed=true`, `success=3 insufficient=2 other=[]` |
| `TestWithdrawOptimistic_RetriesAfterConflict` | version 충돌 뒤 재시도 | 한쪽 `attempts=2` |
| `TestChargeReadThenInsert_WriteSkewExceedsLimit` | 여러 행 규칙이 쓰기 스큐로 깨짐 | `success=2 total=12000` |
| `TestChargeReserved_KeepsLimit` | 한 행으로 모으면 규칙 유지 | `used=8000 total=8000` |
| `TestChargeReserved_RollsBackReservationWhenInsertFails` | 내역 INSERT 실패 시 사용액도 롤백 | 로그 없음, PASS |
| `TestCharge_UnknownCustomer` | 없는 고객은 한도 초과가 아닌 고객 없음 | 로그 없음, PASS |
| `TestPayCheckThenAct_ChargesTwice` | 확인 후 호출하면 이중 승인 | `pg calls=2 rows=2` |
| `TestPayOnce_SameKeyConcurrentChargesOnce` | 선점 후 호출하면 한 번만 승인, 모두 같은 기록 | 한 요청 `SUCCEEDED`, 나머지 `PENDING` (어느 요청인지는 실행마다 다름) |
| `TestPayOnce_SameKeyDifferentRequestIsRejected` | 같은 키에 금액이나 계좌가 다르면 거부 | 로그 없음, PASS |
| `TestPayOnce_CallerDeadlineDuringApprovalIsUnknown` | 요청 기한이 끝나도 `UNKNOWN` 기록 | 로그 없음, PASS |
| `TestPayOnce_TimeoutAfterApprovalIsUnknownUntilReconciled` | 타임아웃은 `UNKNOWN`, 대사로 확정 | `status=SUCCEEDED ... calls=1` |
| `TestReconcile_TimeoutBeforeApprovalBecomesFailed` | 승인 기록이 없으면 `FAILED` | 로그 없음, PASS |
| `TestPayOnce_DeclineIsFailedImmediately` | 확정 거절은 바로 `FAILED` | 로그 없음, PASS |
| `TestReconcile_ClosesPaymentLeftPendingByCrash` | 오래된 `PENDING`만 대사 | 로그 없음, PASS |
| `TestReconcile_TooEarlyConflictsWithInFlightPayment` | 너무 이른 대사는 충돌로 드러남 | `ErrStateChanged로 드러났습니다` |

`TestServerWithdraw_LocalLockProtectsOnlyOneServer`의 같은 서버 경우와 `TestWithdrawLocked_SerializesRequests`의 첫 하위 테스트는 각각 0.3초쯤 걸립니다. 한쪽이 락에 막혀 약속 지점의 대기 상한 300밀리초를 채우기 때문입니다.

## 문서 게이트

저장소 규칙(AGENTS.md 9.2절)에 따라, 공개 식별자에 문서 주석이 없으면 실패하는 검사를 둡니다.

```bash
go install github.com/mgechev/revive@latest
printf '[rule.exported]\n' > /tmp/revive.toml
revive -set_exit_status -config /tmp/revive.toml ./...
```

`revive`는 문서 주석이 `식별자 이름 + 공백`으로 시작하길 요구합니다. 그래서 주석을 `OpenDB는`이 아니라 `OpenDB 함수는`, `ErrKeyReused 오류는`처럼 씁니다.

## 이 구현의 한계

- SQLite에는 행 락이 없습니다. `BEGIN IMMEDIATE`는 DB 전체를 잠그므로, 서로 다른 계좌의 요청도 줄을 섭니다. PostgreSQL과 MySQL의 `SELECT ... FOR UPDATE`는 행 단위로 잠급니다.
- SQLite는 DB 전체에 쓰기 락이 하나뿐이라 `ChargeReserved`의 트랜잭션이 한 번에 하나씩만 실행됩니다. 그래서 그 테스트가 통과한 이유가 규칙을 한 행으로 모았기 때문인지, SQLite가 원래 모든 쓰기를 줄 세웠기 때문인지 가릴 수 없습니다.
- PostgreSQL과 MySQL의 동작은 소스 주석과 개념 문서에 공식 문서를 근거로 적었으며, 이 코드로 실행하지는 않았습니다.
- 서버 여러 대는 한 프로세스 안의 `Server` 값 여러 개로 흉내 냅니다.
