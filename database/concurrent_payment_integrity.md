# 동시에 들어온 요청은 어떻게 결제를 두 번 만들고 잔액을 틀리게 하는가

- [개요](#개요)
- [첫 번째 예시: 잔액 20,000원에서 5,000원 두 번](#첫-번째-예시-잔액-20000원에서-5000원-두-번)
- [틈에 들어오는 조건](#틈에-들어오는-조건)
- [틈을 막는 것처럼 보이지만 막지 못하는 두 가지](#틈을-막는-것처럼-보이지만-막지-못하는-두-가지)
- [틈을 다루는 네 가지 방법](#틈을-다루는-네-가지-방법)
- [외부 결제 호출이 끼면 생기는 문제](#외부-결제-호출이-끼면-생기는-문제)
- [무엇을 고를 것인가](#무엇을-고를-것인가)
- [이 예제가 증명하는 것과 증명하지 않는 것](#이-예제가-증명하는-것과-증명하지-않는-것)
- [정리](#정리)
- [직접 확인하기](#직접-확인하기)
- [근거](#근거)

## 개요

이 문서는 서버가 요청 사이에 공유해야 할 데이터를 자기 메모리에 두지 않고 모두 DB에 두며, 그런 서버가 여러 대 떠서 DB 하나를 공유하는 구조를 전제합니다. 이 구조에서 요청 두 개가 겹치면 두 종류의 사고가 납니다.

- **같은 결제가 두 번 승인됩니다.** 사용자가 결제 버튼을 두 번 누르거나 네트워크가 재전송하면 같은 요청이 두 번 도착합니다.
- **공유 자원의 값이 틀어집니다.** 서로 다른 요청이 같은 계좌 잔액이나 같은 고객 한도를 동시에 줄이는 경우입니다.

두 사고의 원인은 같습니다. **경쟁 조건(race condition)은 한 요청이 값을 "확인"하고 그 결과로 "쓰기" 전까지의 구간에 다른 요청이 들어와서 생깁니다.** 이 문서는 확인과 쓰기 사이의 구간을 **틈**이라고 부릅니다. 틈이 있는 코드는 요청이 겹치는 순서에 따라 깨지고, 틈이 없는 코드는 아무리 겹쳐도 깨지지 않습니다.

이 문서를 읽은 뒤에는 다음을 자기 말로 설명할 수 있어야 합니다.

- 갱신 손실(lost update)과 쓰기 스큐(write skew)가 어떤 틈에서 생기고, 둘이 어떻게 다른지
- 서버 안의 뮤텍스와 "트랜잭션을 열었다"는 사실이 왜 틈을 막지 못하는지
- 조건부 UPDATE, 비관적 락, 낙관적 락, 규칙을 한 행으로 모으기가 각각 틈을 어떻게 다루고 언제 쓰는지
- 외부 결제 대행사 호출이 끼면 왜 호출 전에 DB에 행을 먼저 넣어 "이 결제는 내가 처리한다"를 기록(선점)하고, 타임아웃을 실패가 아닌 결과 모름으로 기록하는지

모든 주장은 [Go 예제](examples/concurrency_transaction/go/)의 테스트로 재현했습니다. 예제는 SQLite 위에서 돌기 때문에, PostgreSQL과 MySQL의 동작은 실행하지 않고 공식 문서를 근거로 적었습니다. 그런 문장에는 근거가 공식 문서라고 따로 밝힙니다.

이 문서에서 **PG**는 PostgreSQL이 아니라 **결제 대행사(Payment Gateway)**를 뜻합니다. PostgreSQL은 항상 PostgreSQL이라고 씁니다.

## 첫 번째 예시: 잔액 20,000원에서 5,000원 두 번

가장 흔한 출금 코드는 잔액을 읽고, 애플리케이션에서 빼고, 결과를 저장합니다.

```go
// WithdrawReadModifyWrite (examples/concurrency_transaction/go/balance.go)
db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&balance)
if balance < amount { return ErrInsufficientBalance }
// <- 틈
db.ExecContext(ctx, "UPDATE accounts SET balance = ? WHERE id = ?", balance-amount, accountID)
```

한 줄씩 보면 이렇습니다.

1. `SELECT`는 DB의 잔액 컬럼 `accounts.balance`를 변수 `balance`에 **복사**합니다. 변수 `balance`는 읽은 순간의 값에 고정되고, 이후 다른 요청이 `accounts.balance`를 바꿔도 따라 바뀌지 않습니다. 이 문서에서 **복사본**은 이렇게 DB에서 읽어 변수에 담아 둔 값을 가리킵니다.
2. `if`는 잔액 복사본(`balance`)이 출금액(`amount`)보다 적은지 판단합니다.
3. `UPDATE`는 복사본으로 계산한 값(`balance-amount`)을 `accounts.balance`에 그대로 덮어씁니다.

1번과 3번 사이가 틈입니다. 요청 두 개가 이 틈에서 겹치면 다음 순서로 진행됩니다.

| 순서 | 요청 A | 요청 B | DB의 잔액 |
| --- | --- | --- | --- |
| 1 | `SELECT` -> 복사본 20,000 | | 20,000 |
| 2 | | `SELECT` -> 복사본 20,000 | 20,000 |
| 3 | 20,000 >= 5,000, 계산 결과 15,000 | 20,000 >= 5,000, 계산 결과 15,000 | 20,000 |
| 4 | `UPDATE balance = 15000` | | 15,000 |
| 5 | | `UPDATE balance = 15000` | 15,000 |

두 출금 모두 성공했는데 잔액은 15,000원입니다. 차례로 실행했다면 10,000원이어야 합니다. 요청 B는 2번에서 만든 잔액 복사본(20,000)으로 계산했기 때문에, 4번에서 A가 `accounts.balance`에 쓴 15,000을 모른 채 같은 값을 덮어썼습니다. **이렇게 나중에 쓴 쪽이 먼저 쓴 변경을 지우는 현상을 갱신 손실(lost update)이라고 합니다.**

테스트 `TestWithdrawReadModifyWrite_LosesUpdate`가 이 표를 그대로 재현합니다.

```text
balance_test.go:26: success=2 balance=15000 (차례로 실행했다면 10000)
```

예시 금액은 일부러 이렇게 골랐습니다. 잔액 10,000원에서 10,000원씩 두 번 빼면 틀린 결과도 0원이 되고, 이 값은 한 번만 출금된 정상 결과와 구별되지 않습니다. 틀린 결과가 정상 결과와 다른 숫자로 나오도록 금액을 골라야 테스트가 버그를 드러냅니다.

테스트는 두 요청을 틈에서 반드시 만나게 합니다. 함수에 `Pause` 인자를 두고, 테스트가 그 자리에 약속 지점(`Rendezvous`)을 넘깁니다. 약속 지점은 두 요청이 모두 도착할 때까지 먼저 온 요청을 붙잡아 둡니다. 스케줄러가 우연히 두 요청을 겹치게 실행하기를 기다리지 않고, 겹치는 순서를 직접 만드는 장치입니다.

## 틈에 들어오는 조건

요청 B가 A의 틈에 들어오려면, **B의 확인 시각이 A의 확인 시각과 A의 쓰기 시각 사이에 있어야 합니다.** 그래서 도착 간격만 보고는 깨질지 판단할 수 없고, 틈의 길이와 함께 봐야 합니다. 틈이 1밀리초인 코드는 요청 간격이 100밀리초면 거의 깨지지 않습니다. 틈이 200밀리초인 코드는 같은 100밀리초 간격에도 깨집니다.

결제에서 틈이 가장 길어지는 원인은 외부 호출입니다. 예를 하나 보겠습니다. 클라이언트는 결제 요청마다 **멱등성 키(idempotency key)**라는 식별자를 만들고, 재시도할 때도 같은 키를 보냅니다. 서버는 이 키로 같은 결제인지 알아봅니다. 이때 같은 키의 기록이 있는지 먼저 조회하고(check), 없으면 결제하고 기록하는(act) 방식을 **Check-then-Act**라고 부릅니다.

```go
// PayCheckThenAct (examples/concurrency_transaction/go/duplicate_request.go)
SELECT COUNT(*) FROM payments_unguarded WHERE idempotency_key = ?   // 확인
// <- 틈: 아래 PG 호출이 끝나고 INSERT할 때까지 이어집니다
gateway.Approve(ctx, req.Key, req.Amount)                          // PG 승인, 수십에서 수백 밀리초
INSERT INTO payments_unguarded ...                                 // 쓰기
```

같은 키로 요청 두 개가 들어오는 시간표를 예로 들면 다음과 같습니다. 시각은 설명을 위한 예시입니다.

| 시각 | 요청 A | 요청 B | `payments_unguarded`의 행 |
| --- | --- | --- | --- |
| 0 ms | `COUNT` -> 0, PG 승인 요청 전송 | | 없음 |
| 100 ms | (PG 응답 대기 중) | `COUNT` -> 0, PG 승인 요청 전송 | 없음 |
| 150 ms | 승인 응답, `INSERT` | | 1 |
| 250 ms | | 승인 응답, `INSERT` | 2 |

요청 B는 A보다 100밀리초 늦게 왔지만, B의 확인(100 ms)이 A의 틈(0~150 ms) 안에 들어왔기 때문에 똑같이 "없음"을 봤습니다. 외부 호출이 틈을 수백 밀리초로 늘려 놓으면, 사용자의 더블 클릭이나 네트워크 재전송처럼 흔한 간격이 모두 틈 안에 들어옵니다.

테스트 `TestPayCheckThenAct_ChargesTwice`에서 PG가 두 번 승인하고 행도 두 개 생겼습니다.

```text
duplicate_request_test.go:47: pg calls=2 rows=2
```

`payments_unguarded`의 `idempotency_key`에 UNIQUE 제약을 걸어도 이 코드는 구할 수 없습니다. 요청 B의 `INSERT`가 제약에 걸려 실패하는 시점(250 ms)에는 B의 PG 승인이 이미 끝났습니다.

## 틈을 막는 것처럼 보이지만 막지 못하는 두 가지

### 서버 안의 뮤텍스

Go의 `sync.Mutex`나 Java의 `synchronized`는 **같은 프로세스 메모리를 공유하는 스레드끼리** 순서를 세웁니다. 서버를 여러 대 띄우면 서버마다 뮤텍스가 따로 있고, 서버끼리 공유하는 것은 DB뿐입니다.

`TestServerWithdraw_LocalLockProtectsOnlyOneServer`는 같은 조건을 두 번 실행합니다.

```text
balance_test.go:54: missed=true balance=10000     <- 같은 Server
balance_test.go:71: missed=false balance=15000    <- Server 두 개
```

- **같은 Server:** 두 번째 요청이 뮤텍스에 막혀 약속 지점에 오지 못했습니다(`missed=true`). 잔액은 10,000원으로 맞습니다.
- **Server 두 개:** 두 요청이 틈에서 만났고(`missed=false`) 갱신 손실이 그대로 났습니다.

이 테스트는 한 프로세스 안에 `Server` 값 두 개를 만들어 서버 두 대를 흉내 냅니다. 두 값은 DB 말고는 아무것도 공유하지 않으므로 공유 관계는 실제 서버 두 대와 같습니다. 다만 실제 프로세스 두 개를 띄워 실행한 것은 아닙니다.

### 트랜잭션을 열기만 하는 것

"트랜잭션 안에서 읽고 썼으니 안전하다"는 판단도 흔합니다. **트랜잭션은 여러 변경을 한 묶음으로 커밋하거나 되돌리는 단위이지, 일반 `SELECT`로 읽은 행을 다른 요청이 못 바꾸게 잠그는 장치가 아닙니다.** 앞의 출금 코드를 `BEGIN`과 `COMMIT` 사이에 넣어도 틈은 그대로 남고, 틈이 어떤 증상으로 드러나는지는 DB마다 다릅니다.

먼저 알아 둘 용어가 셋 있습니다.

- **스냅숏(snapshot):** 읽기를 시작한 시점의 DB 상태입니다. 그 뒤 다른 트랜잭션이 커밋한 변경은 보지 않고 이 상태를 기준으로 읽습니다. 앞의 "복사본"이 변수 하나라면, 스냅숏은 DB 전체를 한 시점에 고정해 보는 방식입니다.
- **격리 수준:** 한 트랜잭션이 다른 트랜잭션의 커밋된 변경을 언제부터 보게 할지 정하는 설정입니다. 스냅숏을 언제 새로 잡는지가 여기서 갈립니다. READ COMMITTED는 문장을 실행할 때마다 새 스냅숏을 잡아 그 시점에 커밋된 값을 보고, REPEATABLE READ와 SERIALIZABLE은 트랜잭션에서 처음 읽은 시점의 스냅숏을 끝까지 씁니다. PostgreSQL의 기본값은 READ COMMITTED, MySQL InnoDB의 기본값은 REPEATABLE READ입니다. SQLite의 읽기 트랜잭션(아래에서 설명)도 트랜잭션 전체에 스냅숏 하나를 씁니다.
- **행 버전:** PostgreSQL은 `UPDATE`할 때 기존 행을 제자리에서 고치지 않고, 고친 내용을 담은 새 행을 하나 더 만든 뒤 기존 행을 옛것으로 표시합니다. 그래서 같은 계좌에 잔액 20,000인 행과 15,000인 행이 잠시 함께 존재할 수 있고, 이 각각을 행 버전이라고 부릅니다. 트랜잭션은 자기 스냅숏에서 볼 수 있는 버전을 골라 읽습니다.

DB별 증상은 다음과 같습니다.

- **SQLite (이 예제로 확인함):** SQLite의 트랜잭션은 두 종류입니다.

    - **읽기 트랜잭션:** 스냅숏을 고정해 읽기만 합니다. 여러 연결이 동시에 가질 수 있습니다.
    - **쓰기 트랜잭션:** DB 전체에 하나뿐인 쓰기 락을 쥡니다. 한 번에 한 연결만 가질 수 있습니다.

    `BEGIN`의 기본값인 DEFERRED는 시작할 때 종류를 정하지 않고 첫 문장을 보고 정합니다. 첫 문장이 `SELECT`면 읽기 트랜잭션이 되고, 같은 트랜잭션에서 나중에 `UPDATE`를 실행하려면 그 순간 쓰기 락을 얻어 **쓰기 트랜잭션으로 전환**해야 합니다. 공식 문서는 이 전환을 upgrade라고 부르며, 다른 연결이 이미 DB를 수정했거나 수정하는 중이면 전환하지 못하고 `SQLITE_BUSY`를 돌려준다고 설명합니다.

    | 순서 | 연결 A | 연결 B | 쓰기 락 |
    | --- | --- | --- | --- |
    | 1 | `BEGIN`, `SELECT` -> 읽기 트랜잭션, 잔액 20,000 | `BEGIN`, `SELECT` -> 읽기 트랜잭션, 잔액 20,000 | 없음 |
    | 2 | `UPDATE` -> 쓰기 락 획득, 쓰기 트랜잭션으로 전환 | | A |
    | 3 | | `UPDATE` -> 전환하려 했으나 A가 쓰기 락을 쥠 -> `SQLITE_BUSY` | A |

    A가 2번에서 커밋까지 마쳤다면 B는 쓰기 락을 얻을 수 있지만, B의 스냅숏(20,000)은 이미 DB(15,000)와 달라 전환이 거부됩니다. 이때는 확장 코드 `SQLITE_BUSY_SNAPSHOT`(517)을 받습니다. 두 연결이 1번에서 `BEGIN` 대신 처음부터 쓰기 락을 잡는 `BEGIN IMMEDIATE`를 썼다면 이 전환이 필요 없습니다. A가 먼저 쓰기 락을 잡으면 B는 `BEGIN IMMEDIATE`에서 A의 커밋을 기다린 뒤, `SELECT`로 A가 커밋한 15,000을 읽습니다.

    예제는 busy timeout(쓰기 락을 못 잡은 연결이 실패하기 전에 기다리는 최대 시간)을 5초로 설정했지만, 이 오류는 기다리지 않고 바로 납니다. busy timeout은 SQLite에 내장된 busy handler(락을 못 잡았을 때 잠시 쉬었다 다시 시도하게 하는 함수)로 구현됩니다. SQLite 공식 문서는 busy handler를 부르면 교착이 생길 수 있을 때 handler를 부르지 않고 곧바로 `SQLITE_BUSY`를 돌려준다고 설명합니다. 문서가 드는 예는 WAL이 아닌 롤백 저널 모드입니다. 연결 A는 읽기 락을 쥔 채 쓰기 락을 기다리고, 연결 B는 쓰기 락을 쥔 채 커밋하려고 A의 읽기 락이 풀리기를 기다립니다. 서로가 서로를 기다리므로 SQLite는 A를 기다리게 하지 않고 곧바로 `SQLITE_BUSY`를 돌려줍니다.

    이 예제가 쓰는 WAL 모드에서는 읽기 트랜잭션이 쓰기 트랜잭션을 막지 않으므로 이 교착 설명이 그대로 맞지 않습니다. 대신 이런 이유가 추론됩니다. B의 읽기 트랜잭션은 이미 1번 시점의 스냅숏에 묶여 있어서, A의 쓰기가 끝날 때까지 기다려 A가 커밋하면 B의 스냅숏이 낡아 어차피 전환할 수 없습니다. 이 추론은 공식 문서의 `SQLITE_BUSY_SNAPSHOT` 설명과 실행 결과(기다리지 않고 바로 오류)에서 따라 나온 것이며, A가 커밋하지 않고 롤백하는 경우까지 설명하지는 못합니다.

    ```text
    balance_test.go:97: worker 0: err=<nil>
    balance_test.go:97: worker 1: err=write balance: database is locked
    ```

    데이터는 지켜졌지만 요청 하나가 오류로 끝났습니다. 이 방식을 쓰려면 오류가 난 트랜잭션을 처음부터 다시 실행해야 합니다.

- **PostgreSQL READ COMMITTED, 기본 격리 수준 (공식 문서 근거, 실행하지 않음):** 공식 문서에 따르면 뒤에 온 UPDATE는 같은 행을 고치는 앞 트랜잭션이 끝나기를 기다립니다. 앞 트랜잭션이 커밋하면, 그 트랜잭션이 만든 최신 행 버전을 보고 **WHERE 조건이 아직 맞는지만** 다시 확인한 뒤 진행합니다. 순서로 보면 이렇습니다.

    | 순서 | 요청 A | 요청 B | `accounts`의 행 버전 |
    | --- | --- | --- | --- |
    | 1 | `SELECT` -> 잔액 복사본 20,000 | `SELECT` -> 잔액 복사본 20,000 | [20,000] |
    | 2 | 앱에서 계산한 `UPDATE ... SET balance = 15000 WHERE id = 'acc'`, 행 락 획득 | | [20,000], [15,000] (A의 새 버전, 커밋 전) |
    | 3 | | 앱에서 계산한 `UPDATE ... SET balance = 15000 WHERE id = 'acc'` -> A의 락 때문에 대기 | 같음 |
    | 4 | `COMMIT` | | 최신 버전 [15,000] |
    | 5 | | 최신 버전을 보고 `id = 'acc'`가 아직 맞는지 다시 확인 -> 맞음 | |
    | 6 | | `SET balance = 15000` 실행 | 최신 버전 [15,000] |

    5번에서 PostgreSQL이 다시 보는 것은 SQL 문장 안의 조건뿐입니다. B가 1번에서 한 `SELECT`를 다시 실행하지 않고, 앱이 넘긴 숫자 `15000`을 다시 계산하지도 않습니다. PostgreSQL에게 `15000`은 그냥 상수라서 그 값이 낡은 잔액으로 계산됐는지 알 수 없습니다. 그래서 두 출금이 모두 성공했는데 잔액은 15,000원이 되고, 오류는 나지 않습니다.

- **MySQL InnoDB REPEATABLE READ, 기본 격리 수준 (공식 문서 근거, 실행하지 않음):** 공식 문서에 따르면 이 수준에서 `SELECT`는 트랜잭션에서 처음 읽은 시점의 스냅숏을 보지만, UPDATE와 DELETE에는 스냅숏이 반드시 적용되지는 않아 다른 트랜잭션이 방금 커밋한 행을 대상으로 동작할 수 있습니다. UPDATE는 대상 행에 배타 락(다른 트랜잭션이 그 행을 고치거나 잠그지 못하게 하는 락)을 겁니다. 순서로 보면 이렇습니다.

    | 순서 | 요청 A | 요청 B | `accounts.balance` |
    | --- | --- | --- | --- |
    | 1 | `SELECT` -> 잔액 복사본 20,000 | `SELECT` -> 잔액 복사본 20,000 | 20,000 |
    | 2 | 앱에서 계산한 `UPDATE ... SET balance = 15000 WHERE id = 'acc'`, 배타 락 획득 | | 15,000 (커밋 전) |
    | 3 | | 앱에서 계산한 `UPDATE ... SET balance = 15000 WHERE id = 'acc'` -> A의 락 때문에 대기 | 같음 |
    | 4 | `COMMIT` | | 15,000 |
    | 5 | | 최신 커밋 행을 대상으로 `SET balance = 15000` 실행 | 15,000 |

    5번에서 B의 UPDATE는 최신 행을 대상으로 실행되지만, 쓰는 값은 앱이 1번의 복사본 20,000으로 계산해 넘긴 상수 15000입니다. 그래서 PostgreSQL과 마찬가지로 오류 없이 잔액이 15,000으로 남습니다.

같은 코드인데 SQLite에서는 오류가 나고, PostgreSQL과 MySQL의 기본 격리 수준에서는 조용히 틀린 값이 남습니다. 한 DB에서 테스트가 통과했다고 다른 DB에서도 안전하다고 판단하면 안 되는 이유입니다.

## 틈을 다루는 네 가지 방법

| 방법 | 틈을 어떻게 다루는가 | 대표 구현 | 경합에서 진 요청이 받는 것 | Go 함수 |
| --- | --- | --- | --- | --- |
| 조건부 UPDATE | 확인과 쓰기를 SQL 한 문장에 넣어 틈을 없앰 | `UPDATE ... WHERE balance >= ?`, UNIQUE `INSERT` | 0행 또는 제약 위반 | `WithdrawConditional`, `claim` |
| 비관적 락 | 읽기 전에 락을 잡아 커밋까지 다른 요청을 기다리게 함 | `SELECT ... FOR UPDATE`, SQLite `BEGIN IMMEDIATE` | 락 대기 | `WithdrawLocked` |
| 낙관적 락 | 틈은 두고, 쓸 때 누가 바꿨는지 검출해 재시도 | version 비교 `UPDATE`, SERIALIZABLE | 0행이나 직렬화 실패 오류, 그 뒤 재시도 | `WithdrawOptimistic` |
| 규칙을 한 행으로 모으기 | 여러 행에 걸친 규칙을 한 행의 값으로 바꿔 위 방법을 쓸 수 있게 함 | 사용액 행 조건부 `UPDATE` | 0행 | `ChargeReserved` |

`SELECT ... FOR UPDATE`는 비관적 락 절에서, SERIALIZABLE과 직렬화 실패 오류는 낙관적 락 절 끝에서 설명합니다.

### 조건부 UPDATE: 확인과 쓰기를 한 문장에 넣기

규칙이 행 하나 안에서 표현되면 확인을 `WHERE`로 옮깁니다.

```sql
UPDATE accounts
   SET balance = balance - ?          -- (1) 앱의 복사본이 아니라 DB의 현재 값에서 뺍니다
 WHERE id = ? AND balance >= ?        -- (2) 잔액 확인이 같은 문장 안에 있습니다
```

- SQL 안의 `balance`는 Go 변수가 아니라 컬럼 `accounts.balance`입니다.
- **(1)** `SET balance = balance - ?`는 애플리케이션이 과거에 복사한 값이 아니라 DB의 현재 잔액에서 출금액을 뺍니다. 이 문장에는 복사본이 없습니다.
- **(2)** `WHERE balance >= ?`는 현재 잔액이 출금액 이상인지를 같은 문장 안에서 확인하므로, 확인과 쓰기 사이에 애플리케이션 코드가 끼지 않습니다. 그래서 Go 함수 `WithdrawConditional`은 `Pause` 인자를 받지 않습니다. 멈출 틈이 없기 때문입니다.
- 결과는 바뀐 행 수로 판정합니다. 1이면 출금 성공, 0이면 잔액이 모자라거나 계좌가 없는 경우입니다.

두 요청이 겹칠 때 뒤의 요청이 앞의 결과를 보고 판단하도록 만드는 주체는 DB이며, 방식은 DB마다 다릅니다.

- **SQLite:** DB 전체에서 쓰기를 한 번에 하나만 실행하므로, 뒤의 UPDATE는 앞 UPDATE가 자동 커밋된 뒤에 실행됩니다. 그래서 `WHERE balance >= ?`와 `SET balance = balance - ?`가 앞 요청이 줄인 잔액으로 계산됩니다.
- **PostgreSQL READ COMMITTED:** 뒤의 UPDATE는 앞 트랜잭션의 커밋을 기다린 뒤, 앞 트랜잭션이 만든 최신 행 버전으로 `WHERE balance >= ?`를 다시 확인합니다. 앞 요청이 잔액을 0원으로 만들었다면 `balance >= 10000`이 거짓이 되어 0행이 바뀝니다. 조건이 참일 때 실행되는 `SET balance = balance - ?`도 상수가 아니라 컬럼을 쓰는 식이므로 최신 잔액에서 계산됩니다. 공식 문서도 뒤의 UPDATE가 "갱신된 행 버전을 사용해" 작업을 진행한다고 설명합니다. 앞의 Read-Modify-Write와 달리 SQL 밖에서 미리 계산한 값이 없으므로, 다시 확인하는 것만으로 충분합니다.
- **MySQL InnoDB:** 앞 UPDATE가 대상 행에 배타 락을 걸므로 뒤의 UPDATE는 그 락이 풀릴 때까지 기다립니다. 락이 풀리면 스냅숏이 아닌 최신 커밋 상태의 잔액으로 `WHERE balance >= ?`와 `SET balance = balance - ?`를 계산합니다.

잔액 10,000원에서 10,000원 출금 다섯 건을 동시에 보내도 한 건만 성공합니다.

```text
balance_test.go:122: success=1 insufficient=4 other=[] balance=0
```

**PostgreSQL에서는 격리 수준에 따라 결과가 달라집니다.** REPEATABLE READ나 SERIALIZABLE의 트랜잭션은 처음 읽은 시점의 스냅숏만 봐야 합니다. 그래서 뒤의 UPDATE가 고치려는 행이 그 스냅숏 이후 앞 트랜잭션에 의해 바뀌었다면, READ COMMITTED처럼 최신 버전으로 WHERE를 다시 확인해 0행을 돌려주는 대신 `could not serialize access due to concurrent update` 오류로 실패합니다. 이 오류의 SQLSTATE(SQL 표준 오류 코드)는 `40001`이고, 공식 문서는 트랜잭션 전체를 처음부터 재시도하라고 안내합니다.

스키마의 `CHECK (balance >= 0)`는 마지막 방어선입니다. `WHERE`의 잔액 조건을 일부러 지운 코드로 테스트했더니, 네 요청이 `CHECK constraint failed: balance >= 0` 오류로 막혔습니다. 다만 이 제약이 막는 것은 음수 잔액뿐입니다. 앞의 갱신 손실처럼 복사본으로 계산한 양수 값을 덮어쓰는 오류는 막지 못합니다.

### 비관적 락: 읽은 값을 커밋까지 잠그기

규칙이 UPDATE 한 문장에 들어가지 않을 때가 있습니다. 여러 테이블을 읽고 판단한 뒤 써야 하는 경우입니다. 이때는 읽기 전에 락을 잡고 커밋할 때까지 놓지 않습니다.

```sql
-- PostgreSQL, MySQL
BEGIN;
SELECT balance FROM accounts WHERE id = ? FOR UPDATE;   -- 행 락을 잡고 읽습니다
-- 애플리케이션에서 여러 규칙을 확인합니다
UPDATE accounts SET balance = ? WHERE id = ?;
COMMIT;                                                 -- 여기서 락이 풀립니다
```

락 안에서 읽은 값은 커밋 전까지 DB의 실제 값과 같습니다. 그래서 복사본으로 계산해 덮어써도 안전합니다.

SQLite에는 행 락과 `FOR UPDATE`가 없습니다. 대신 `BEGIN IMMEDIATE`가 첫 SQL보다 먼저 **DB 전체**의 쓰기 락을 잡습니다. SQLite 공식 문서에 따르면 `BEGIN IMMEDIATE`가 성공한 뒤로는 커밋할 때까지 그 연결의 작업이 `SQLITE_BUSY`를 받지 않습니다. 이 연결은 아직 아무것도 읽지 않아 스냅숏이 없으므로, 앞 연결이 커밋할 때까지 기다렸다가 쓰기 락을 얻어도 낡은 스냅숏 문제가 없습니다. 그래서 SQLite가 busy timeout까지 기다려 줍니다. 먼저 `SELECT`한 뒤 `UPDATE`하는 일반 `BEGIN` 트랜잭션이 곧바로 `SQLITE_BUSY`를 받는 것과 이 점이 다릅니다.

여기에는 흔한 실수가 하나 있습니다. Go의 database/sql로 `db.Begin()`을 부르면 SQLite에는 일반 `BEGIN`, 즉 DEFERRED가 갑니다. 락을 잡는다고 믿고 이렇게 쓰면 테스트가 통과할 수도 있습니다. 잔액 30,000원에서 10,000원 출금 다섯 건을 보낸 결과를 비교하면 이유가 보입니다.

| 구현 | 성공 | 잔액 부족 | 그 밖의 오류 |
| --- | --- | --- | --- |
| `BEGIN` (DEFERRED) | 3 | 0 | 2 (`database is locked`) |
| `BEGIN IMMEDIATE` | 3 | 2 | 0 |

두 구현 모두 성공은 3건이고 최종 잔액도 0원입니다. 그런데 첫 번째 구현에서 두 요청은 락을 기다린 끝에 잔액 부족으로 거절된 것이 아니라 **락 경합 오류로 실패했습니다.** 테스트가 성공 수와 최종 잔액만 본다면 이 차이를 놓칩니다.

그래서 `TestWithdrawLocked_SerializesRequests`는 두 가지를 함께 확인합니다.

- **그 밖의 오류가 0건인가:** 결과를 성공, 기대한 업무 오류, 그 밖의 오류로 나눠 셉니다.
- **락이 실제로 상대를 막았는가:** 락을 쥔 요청이 멈춘 동안 상대가 약속 지점에 오지 못해야 합니다(`missed=true`).

```text
balance_test.go:152: missed=true success=2 balance=10000
balance_test.go:171: success=3 insufficient=2 other=[] balance=0
```

`BEGIN IMMEDIATE`를 일반 `BEGIN`으로 바꿔 실행하면 이 테스트가 `missed=false success=1 balance=15000`으로 실패했습니다.

비관적 락에는 지켜야 할 규칙이 하나 있습니다. **락을 쥔 채 외부 호출을 하지 않습니다.** PG 응답이 느려지면 그동안 행 락과 DB 커넥션 하나가 함께 묶입니다. 같은 계좌의 후속 요청이 락을 기다리며 커넥션을 하나씩 더 잡고, 커넥션 풀이 바닥나면 이 계좌와 상관없는 API까지 커넥션을 얻지 못해 실패합니다.

### 낙관적 락: 쓰는 순간 충돌을 검출하기

낙관적 락은 틈을 그대로 두고, 쓰는 순간에 "내가 읽은 뒤 누가 바꿨는가"를 확인합니다.

```sql
SELECT balance, version FROM accounts WHERE id = ?;                        -- 잔액과 version을 함께 복사합니다
UPDATE accounts SET balance = ?, version = version + 1
 WHERE id = ? AND version = ?;                                             -- 복사한 version 그대로일 때만 씁니다
```

`version`은 이 행이 몇 번 바뀌었는지 세는 값입니다. 잔액을 바꾸는 **모든** SQL이 `version`을 올려야 이 판단이 맞습니다. 한 곳이라도 version을 올리지 않고 잔액을 바꾸면 낙관적 락은 그 변경을 보지 못합니다.

같은 20,000원 계좌에서 5,000원 출금 두 건이 겹치면 다음 순서로 진행됩니다.

| 순서 | 요청 A | 요청 B | DB (잔액, version) |
| --- | --- | --- | --- |
| 1 | 복사 -> (20,000, 0) | 복사 -> (20,000, 0) | (20,000, 0) |
| 2 | `WHERE version = 0` -> 1행 | | (15,000, 1) |
| 3 | | `WHERE version = 0` -> **0행** | (15,000, 1) |
| 4 | | 다시 복사 -> (15,000, 1) | (15,000, 1) |
| 5 | | `WHERE version = 1` -> 1행 | (10,000, 2) |

```text
balance_test.go:195: worker 0: attempts=1 err=<nil>
balance_test.go:195: worker 1: attempts=2 err=<nil>
```

재시도 횟수에는 상한이 필요합니다. 여기서 충돌은 version 비교 UPDATE가 0행을 바꾼 경우입니다. 0행이 나오려면 다른 요청의 UPDATE가 먼저 성공해 version을 올렸어야 하므로, 충돌이 날 때마다 적어도 한 요청의 출금은 끝납니다. 하지만 같은 요청이 매번 그 다른 요청에게 질 수는 있습니다. 이 예제는 `maxAttempts`를 넘으면 `ErrTooManyConflicts`를 돌려줍니다.

충돌이 드문 곳에서는 락을 기다리지 않아 빠르고, 같은 행에 요청이 몰리면 재시도가 늘어 오히려 느립니다. 잔액처럼 한 행에 요청이 집중되는 자원에는 조건부 UPDATE가 더 단순합니다.

틈은 그대로 두고 충돌을 검출해 재시도하게 하는 일을 DB가 대신 해 주는 방식이 PostgreSQL SERIALIZABLE입니다. 공식 문서에 따르면, 트랜잭션들을 어떤 순서로 하나씩 실행해도 나올 수 없는 결과(직렬화 이상)가 생길 것을 검출하면 그중 한 트랜잭션을 SQLSTATE `40001`로 실패시키고, 애플리케이션은 이 오류를 받아 재시도할 수 있게 준비해야 합니다.

### 여러 행에 걸친 규칙: 쓰기 스큐와 한 행으로 모으기

지금까지의 규칙은 모두 행 하나 안에 있었습니다. 규칙이 여러 행에 걸쳐 있으면 다른 모양의 실패가 생깁니다.

고객별 하루 결제 한도가 10,000원이고, 결제 내역은 `charges` 테이블에 한 건마다 한 행으로 쌓인다고 하겠습니다. 지켜야 할 규칙은 "이 고객의 `charges` 합계는 10,000원 이하"입니다.

```go
// ChargeReadThenInsert (examples/concurrency_transaction/go/daily_limit.go)
SELECT SUM(amount) FROM charges WHERE customer_id = ?     // 합계 복사
if used+amount > limit { return ErrLimitExceeded }
// <- 틈
INSERT INTO charges (customer_id, amount) VALUES (?, ?)   // 새 행 추가
```

| 순서 | 요청 A (6,000원) | 요청 B (6,000원) | `charges` 합계 |
| --- | --- | --- | --- |
| 1 | 합계 -> 0, 0 + 6,000 <= 10,000 | | 0 |
| 2 | | 합계 -> 0, 0 + 6,000 <= 10,000 | 0 |
| 3 | 새 행 `INSERT` | | 6,000 |
| 4 | | 새 행 `INSERT` | **12,000** |

```text
daily_limit_test.go:43: success=2 total=12000 (한도 10000)
```

갱신 손실과 비교하면 차이가 보입니다. 두 요청은 **같은 행을 덮어쓰지 않았습니다.** 각자 새 행을 하나씩 추가했고, 각 요청은 자기가 읽은 합계 기준으로는 옳게 판단했습니다. 두 결정이 합쳐졌을 때 여러 행에 걸친 규칙이 깨졌습니다. **이렇게 각자 옳은 판단이 합쳐져 여러 행에 걸친 규칙을 깨는 현상을 쓰기 스큐(write skew)라고 합니다.**

이 경우 기존 `charges` 행에만 행 락을 걸어서는 막지 못합니다. PostgreSQL이나 MySQL에서 합계를 읽기 전에 이 고객의 기존 `charges` 행을 `SELECT ... FOR UPDATE`로 잠근다고 해 보겠습니다. 첫 결제라면 잠글 행이 0개라서 두 요청 모두 기다리지 않고, 이어서 각자 서로 다른 새 행을 INSERT하므로 같은 행을 두고 기다리는 순간이 한 번도 오지 않습니다. PostgreSQL은 `SUM` 같은 집계와 `FOR UPDATE`를 함께 쓰는 것도 허용하지 않습니다. 막는 방법은 여러 가지이고, 각각 대가가 다릅니다.

- **규칙을 한 행으로 모으기:** 고객마다 한도 행을 두고 사용액을 그 행에 누적합니다. 아래에서 자세히 봅니다.
- **부모 행 잠그기:** 같은 고객의 결제가 모두 거치는 행(고객 행이나 한도 행)을 `FOR UPDATE`로 잠근 뒤 합계를 읽습니다. 같은 고객의 결제가 줄을 섭니다.
- **범위 잠그기:** MySQL InnoDB REPEATABLE READ에서는 인덱스가 있는 조건으로 `FOR UPDATE`를 하면 next-key 락과 gap 락이 그 범위의 새 행 삽입을 막을 수 있습니다. gap 락은 인덱스 레코드 사이의 빈 구간을 잠가 그 구간에 새 행이 들어오지 못하게 하고, next-key 락은 레코드 락과 그 앞 gap 락을 합친 것입니다. 어떤 범위가 잠기는지는 인덱스와 실행 계획에 달려 있습니다. [격리 수준, 락, 데드락 문서](../interviews/database-deep-dive/08-isolation-lock-deadlock.md)의 gap 락 절에서 이어집니다.
- **DB나 테이블 전체 잠그기:** SQLite의 `BEGIN IMMEDIATE`로 감싸면 막힙니다. 대신 모든 고객의 결제가 한 줄로 섭니다.
- **SERIALIZABLE:** PostgreSQL SERIALIZABLE은 쓰기 스큐를 검출해 `40001`로 실패시키므로 재시도가 필요합니다.

규칙을 한 행으로 모으는 코드는 이렇습니다.

```sql
BEGIN;
UPDATE daily_limits SET used_amount = used_amount + ?
 WHERE customer_id = ? AND used_amount + ? <= limit_amount;   -- 0행이면 한도 초과
INSERT INTO charges (customer_id, amount) VALUES (?, ?);      -- 같은 트랜잭션에서 내역 기록
COMMIT;
```

이제 같은 고객의 모든 결제가 `daily_limits`의 같은 행을 UPDATE하므로, 문제가 앞에서 본 단일 행 조건부 UPDATE로 줄어듭니다. 한도 10,000원에 6,000원 두 건이 겹치면 PostgreSQL과 MySQL에서 이렇게 진행됩니다.

| 순서 | 요청 A | 요청 B | `used_amount` |
| --- | --- | --- | --- |
| 1 | `UPDATE`, `WHERE 0 + 6000 <= 10000` 참, 행 락 획득 | | 6,000 (커밋 전) |
| 2 | | 같은 행 `UPDATE` -> A의 락 때문에 대기 | 같음 |
| 3 | `INSERT INTO charges`, `COMMIT` | | 6,000 |
| 4 | | 최신 `used_amount`로 `WHERE 6000 + 6000 <= 10000` 다시 확인 -> 거짓, 0행 | 6,000 |

앞 UPDATE가 잡은 행 락이 커밋까지 유지되므로, 뒤 UPDATE는 앞 트랜잭션이 끝난 뒤 최신 `used_amount`로 조건을 다시 계산합니다. `INSERT`를 같은 트랜잭션에 넣어야 사용액과 결제 내역이 함께 커밋되거나 함께 취소됩니다.

```text
daily_limit_test.go:65: success=2 exceeded=3 other=[] used=8000 total=8000
```

내역 INSERT를 일부러 실패시키는 테스트 `TestChargeReserved_RollsBackReservationWhenInsertFails`에서 사용액도 0원으로 되돌아갔습니다. 한계도 하나 있습니다. SQLite는 DB 전체에 쓰기 락이 하나뿐이라 `ChargeReserved`의 트랜잭션이 한 번에 하나씩만 실행됩니다. 그래서 이 테스트가 통과한 이유가 "규칙을 `daily_limits` 한 행으로 모았기 때문"인지, "SQLite가 원래 모든 쓰기를 줄 세웠기 때문"인지 가릴 수 없습니다. 행마다 따로 락을 거는 PostgreSQL과 MySQL에서도 규칙이 지켜진다는 근거는 위 순서표(공식 문서의 행 락 동작)이며, 실행해 확인한 것은 아닙니다.

## 외부 결제 호출이 끼면 생기는 문제

잔액과 한도의 문제는 DB 안에서 끝납니다. 결제 승인은 PG가 하므로 DB 트랜잭션으로 되돌릴 수 없습니다. 그래서 두 가지 질문이 더 생깁니다.

- 같은 요청이 여러 번 와도 PG를 한 번만 부르려면 어떻게 해야 하는가
- PG 응답을 받지 못했을 때 무엇을 기록해야 하는가

### PG를 부르기 전에 선점한다

앞에서 본 멱등성 키를 다시 씁니다. 서버는 PG를 부르기 전에 이 키로 `PENDING` 행을 INSERT합니다.

```sql
INSERT INTO payments (idempotency_key, account_id, amount, status, updated_at_ms)
VALUES (?, ?, ?, 'PENDING', <DB 현재 시각>);    -- idempotency_key에 UNIQUE 제약이 있습니다
```

이렇게 외부 호출 전에 DB에 행을 넣어, 이 멱등성 키로 PG를 부를 권리를 한 요청이 먼저 가져가는 것을 이 문서는 **선점**이라고 부릅니다. UNIQUE 제약 때문에 같은 키의 INSERT는 한 건만 성공합니다. **성공한 요청만 PG를 부르고**, 제약에 막힌 요청은 PG를 부르지 않은 채 기존 행을 읽어 돌려줍니다. Check-then-Act와 순서가 반대입니다. 기록이 먼저이고 돈의 이동이 나중이므로, 같은 키로 PG를 두 번 부르는 틈은 없습니다. 다만 선점한 요청이 결과를 기록하기 전에 프로세스가 죽거나 PG 응답을 잃으면 행이 `PENDING`이나 `UNKNOWN`으로 남습니다. 이런 결제는 뒤에서 설명할 대사가 확정합니다.

| 순서 | 요청 A | 요청 B | `payments` 행 | PG 호출 수 |
| --- | --- | --- | --- | --- |
| 1 | `INSERT PENDING` -> 성공 | | PENDING | 0 |
| 2 | | `INSERT PENDING` -> UNIQUE 위반 | PENDING | 0 |
| 3 | PG 승인 요청 | 기존 행 읽기 -> PENDING 반환 | PENDING | 1 |
| 4 | 승인 응답, `SUCCEEDED` 기록 | | SUCCEEDED | 1 |
| 5 | | 같은 키로 재요청 -> SUCCEEDED 반환 | SUCCEEDED | 1 |

같은 키로 다섯 요청을 동시에 보낸 테스트에서도 PG 호출은 한 번이었습니다. 테스트는 선점한 요청의 PG 호출을 나머지 네 요청이 모두 `PENDING`을 받아 돌아올 때까지 붙잡아 둡니다. 네 요청이 선점한 요청의 결과 기록보다 늦게 왔다면 `SUCCEEDED`를 받았을 것이므로, 넷 모두 `PENDING`을 받았다는 결과(SUCCEEDED 하나, PENDING 넷)가 곧 선점한 요청이 PG 응답을 기다리는 동안 들어왔다는 증거입니다. 어느 요청이 선점할지는 실행마다 달라집니다.

```text
duplicate_request_test.go:77: worker 0: status=PENDING err=<nil>
duplicate_request_test.go:77: worker 1: status=PENDING err=<nil>
duplicate_request_test.go:77: worker 2: status=PENDING err=<nil>
duplicate_request_test.go:77: worker 3: status=PENDING err=<nil>
duplicate_request_test.go:77: worker 4: status=SUCCEEDED err=<nil>
```

UNIQUE 충돌을 판정하는 방식도 DB마다 다릅니다. "인덱스 페이지의 래치를 먼저 잡은 쪽이 이긴다"는 설명은 정확하지 않습니다. 래치(latch)는 B-Tree 페이지 같은 메모리 구조가 동시 수정으로 깨지지 않게 아주 짧게 잡는 보호 장치일 뿐이고, 유일성은 다음 방식으로 판정합니다.

- **SQLite:** 쓰기를 한 번에 하나만 실행하므로, 뒤의 INSERT는 앞의 커밋이 끝난 뒤에 중복을 발견합니다.
- **PostgreSQL:** 공식 문서에 따르면, 충돌한 행을 넣은 트랜잭션이 아직 커밋되지 않았으면 뒤의 INSERT는 **그 트랜잭션이 끝날 때까지 기다립니다.** 앞 트랜잭션이 커밋하면 유일성 위반이고, 롤백하면 충돌이 아닙니다.
- **MySQL InnoDB:** 공식 문서에 따르면 중복 키 오류가 나면 중복된 인덱스 레코드에 **공유 락**(여러 트랜잭션이 함께 쥘 수 있지만 배타 락과는 함께 쥘 수 없는 락)을 겁니다. 매뉴얼의 세 세션 예시는 이렇게 데드락에 빠집니다.

    1. S1이 키를 INSERT해 그 인덱스 레코드에 배타 락을 쥡니다.
    2. S2와 S3이 같은 키를 INSERT하다 중복을 만나고, 각자 공유 락을 요청하며 기다립니다.
    3. S1이 롤백하면 S2와 S3이 둘 다 공유 락을 얻습니다.
    4. 둘 다 실제로 INSERT하려고 배타 락을 요청하지만, 상대의 공유 락 때문에 서로를 기다립니다.
    5. InnoDB가 한쪽을 데드락 오류(1213)로 끝냅니다.

    그래서 MySQL에서는 선점 INSERT가 1213을 받으면 재시도해야 합니다.

실무에서 이 차이가 중요해지는 경우가 있습니다. 선점 INSERT를 긴 트랜잭션 안에 넣으면, 같은 키로 온 뒤의 요청은 즉시 거절되지 않고 앞 트랜잭션이 끝날 때까지 **블로킹**됩니다. 또 앞 트랜잭션이 롤백하면 뒤의 요청이 새로 선점에 성공합니다. 예제의 `claim`은 INSERT를 자동 커밋 한 문장으로 끝내서 대기 시간을 짧게 유지합니다.

### 키가 같다고 같은 요청은 아니다

같은 키로 금액이나 계좌가 다른 요청이 올 수 있습니다. 이때 기존 결과를 그대로 돌려주면, 99,999원을 요청한 클라이언트가 30,000원짜리 결제의 성공 응답을 받고 자기 요청이 성공했다고 믿게 됩니다. 예제는 선점에 실패하면 저장된 계좌와 금액을 요청과 비교하고, 다르면 `ErrKeyReused`를 돌려줍니다. 실제 시스템은 보통 요청 본문을 정규화한 뒤 해시(request fingerprint)로 저장해 비교합니다.

### 타임아웃은 실패가 아니다

PG 응답은 세 가지로 나뉩니다.

| PG 응답 | 돈이 움직였는가 | 기록할 상태 |
| --- | --- | --- |
| 승인 | 움직임 | `SUCCEEDED` |
| 거절 | 움직이지 않음 | `FAILED` |
| 타임아웃, 연결 끊김, 요청 기한 초과 | **모름** | `UNKNOWN` |

타임아웃은 PG가 요청을 받지 못했다는 뜻일 수도 있고, 승인은 했는데 응답만 돌아오지 못했다는 뜻일 수도 있습니다. 이를 `FAILED`로 기록하면 돈은 나갔는데 결제는 실패로 남고, 사용자가 다시 결제하면 이중 결제가 됩니다. **모르는 결과는 모른다고 기록하고, 나중에 PG에 조회해 확정합니다.** 내부 기록과 외부 기록을 대조해 맞추는 이 작업을 **대사(reconciliation)**라고 합니다.

결과를 기록하는 단계에도 주의할 점이 있습니다. 호출자의 요청 기한이 PG 응답을 기다리는 도중에 끝나면, 같은 기한으로 DB에 기록하려 해도 곧바로 실패합니다. 그러면 행이 `UNKNOWN`이 아니라 `PENDING`으로 남습니다. 예제의 `PayOnce`는 기록할 때 `context.WithoutCancel(ctx)`로 새 문맥을 만듭니다. 이 문맥은 요청 문맥의 값은 이어받지만 요청 기한이 끝나도 함께 끝나지 않으므로, 여기에 기록용 기한을 따로 줍니다. 테스트 `TestPayOnce_CallerDeadlineDuringApprovalIsUnknown`이 이 경우를 확인합니다.

결제 한 건의 상태는 다음과 같이 움직입니다.

```text
                 PG 승인 -----------------------------> SUCCEEDED
                /                                         ^
PENDING -------+-- PG 거절 ---------------------------> FAILED
    |           \                                         ^   ^
    |            결과 모름 ------> UNKNOWN --(대사: PG 조회)-+---+
    |                                                     |
    +-- 결과 기록 전에 프로세스가 죽음, stuckAfter 경과 ----(대사)
```

테스트 `TestPayOnce_TimeoutAfterApprovalIsUnknownUntilReconciled`는 PG가 승인하고 응답만 잃어버린 경우를 재현합니다. 첫 요청과 재요청 모두 `UNKNOWN`을 받고 PG는 다시 호출되지 않습니다. 대사를 거친 뒤에야 `SUCCEEDED`로 확정됩니다.

```text
duplicate_request_test.go:166: reconciled: status=SUCCEEDED pg_transaction_id=pg-order-1-1 calls=1
```

이 흐름은 [애플리케이션 경계, 멱등성, 금액 처리 문서](../interviews/database-deep-dive/12-application-boundaries-idempotency-money-outbox.md)의 멱등성 상태 기계와 같은 상태 이름을 씁니다.

### 대사의 시간 조건과 그 한계

`PENDING`으로 오래 남은 결제도 대사 대상입니다. PG를 부른 프로세스가 결과를 기록하기 전에 죽었을 수 있기 때문입니다. 예제의 `PayOnce`는 PG 호출에 `ApproveTimeout`(5초), 결과 기록에 `recordTimeout`(5초) 기한을 겁니다. 기한이 없으면 "얼마나 기다려야 멈춘 결제인가"를 정할 기준이 없습니다. 정상 처리 중인 결제도 선점 뒤 PG 응답 대기 최대 5초와 결과 기록 최대 5초 동안 `PENDING`일 수 있고, 경과 시간은 선점한 시각부터 잽니다. 그래서 "오래"의 기준인 `stuckAfter`는 두 기한을 더한 10초에 여유를 더한 값보다 길어야 합니다. 경과 시간은 서버 시계가 아니라 DB 시계로 잽니다. 서버마다 시계가 조금씩 다르면 같은 결제를 두고 서버마다 판단이 달라지기 때문입니다.

기준이 너무 짧으면 다음 순서의 사고가 납니다. 테스트 `TestReconcile_TooEarlyConflictsWithInFlightPayment`가 이 순서를 재현합니다.

1. 요청이 `PENDING`을 선점하고 PG 응답을 기다립니다.
2. 대사 작업이 이 결제를 오래된 PENDING으로 판단해 PG에 조회합니다. 아직 승인 기록이 없으므로 `FAILED`로 확정합니다.
3. PG 승인이 도착합니다.
4. 원래 요청이 `SUCCEEDED`를 기록하려 하지만, 상태가 이미 `PENDING`이 아니라서 거부됩니다.

```text
duplicate_request_test.go:259: PG는 pg-order-1-1로 승인했지만 결제는 FAILED로 남았고, 이 충돌은 ErrStateChanged로 드러났습니다
```

4번에서 상태 확정은 `WHERE idempotency_key = ? AND status = 'PENDING'`처럼 **이전 상태를 조건으로 거는 UPDATE**로 합니다. 이 조건이 없으면 뒤늦은 기록이 대사 결과를 조용히 덮어쓰고, 어느 쪽이 맞았는지 흔적도 남지 않습니다. 조건이 있으면 적어도 충돌이 오류로 드러납니다. 잔액의 조건부 UPDATE와 같은 원리를 상태 전이에 적용한 것입니다.

**`stuckAfter`를 길게 잡는 것만으로 이 사고를 완전히 없앨 수는 없습니다.** 다음 순서를 보면 이유가 드러납니다.

1. 요청이 선점한 직후, 가비지 컬렉터가 도는 동안 프로그램 전체가 멈추는 GC 정지 같은 이유로 프로세스가 `stuckAfter`보다 오래 멈춥니다.
2. 그사이 대사 작업이 이 결제를 오래된 PENDING으로 판단해 PG에 조회하고, 아직 승인이 없으니 `FAILED`로 확정합니다.
3. 프로세스가 깨어나 PG를 부르고, PG가 승인합니다.

1번의 정지가 얼마나 길지는 미리 알 수 없으므로, 시간 기준으로는 이 경우를 막을 수 없습니다. 완전한 방어는 PG 쪽에 있습니다. PG가 같은 멱등성 키의 승인을 한 번만 처리하거나, 확정 뒤 도착한 승인을 취소할 수 있어야 합니다. 그래도 `ErrStateChanged`가 났는데 PG가 승인했다면 돈은 나갔고 결제는 실패로 남은 것이므로, 경보를 내고 환불해야 합니다. 이 예제는 그 처리를 구현하지 않습니다.

대사는 조회해서 승인 기록이 없으면 `FAILED`로 확정합니다. 이 판단도 "조회에 나오지 않은 요청은 이후에도 승인되지 않는다"는 PG의 보장이 있어야 맞습니다. 실제 PG에서는 이 보장이 어떤 조건에서 성립하는지 연동 규격에서 확인해야 합니다.

## 무엇을 고를 것인가

| 상황 | 먼저 고려할 방법 | 이유 | 주의할 점 |
| --- | --- | --- | --- |
| 같은 요청의 중복 실행 방지 | UNIQUE 선점 후 외부 호출 | 제약이 한 건만 통과시킴 | 요청 내용 비교, `UNKNOWN` 상태, 대사, PG 쪽 멱등성이 함께 필요 |
| 잔액, 재고처럼 한 행의 값 차감 | 조건부 UPDATE | 틈이 없고 락 유지 시간이 한 문장뿐 | PostgreSQL REPEATABLE READ 이상에서는 0행 대신 `40001` |
| 여러 테이블을 읽고 판단한 뒤 쓰기 | 비관적 락 | 판단에 쓴 값이 커밋까지 고정됨 | 락 안에서 외부 호출 금지, 데드락 대비 |
| 충돌이 드문 편집 | 낙관적 락 | 대기 없이 진행 | 모든 쓰기가 version을 올려야 함, 재시도 상한 |
| 여러 행에 걸친 규칙 | 한 행으로 모으기, 부모 행 락, SERIALIZABLE | 기존 행에만 거는 락은 새 행 삽입을 막지 못함 | SERIALIZABLE은 `40001` 재시도 필수 |

Redis 같은 외부 저장소의 분산 락은 이 표에 넣지 않았습니다. 락을 쥔 프로세스가 GC 정지나 네트워크 지연으로 멈춘 사이 락이 만료되면, 두 프로세스가 동시에 자기가 락을 쥐었다고 믿을 수 있습니다. 펜싱 토큰(fencing token)은 락을 내줄 때마다 1씩 커지는 번호입니다. 쓰기를 받는 저장소가 지금까지 본 가장 큰 번호보다 작은 번호의 쓰기를 거부하면, 멈췄다 깨어난 옛 소유자의 쓰기가 막힙니다. 저장소가 이렇게 거부하지 않는 한 분산 락만으로는 정확성을 보장하지 못한다는 비판이 널리 알려져 있습니다. "정해진 시간이 지나면 상대는 끝났다"는 가정이 프로세스 정지로 깨진다는 점에서, 앞의 `stuckAfter` 한계와 같은 종류입니다. 부하를 DB 앞에서 걸러 내는 용도로는 쓸 수 있지만, 돈의 정확성은 DB 제약과 조건부 쓰기로 지키는 편이 안전합니다.

## 이 예제가 증명하는 것과 증명하지 않는 것

**실행해서 확인한 것**

- SQLite에서 이 문서에 인용한 모든 테스트 출력. 설정은 WAL 모드(변경을 먼저 별도 로그 파일에 쓰는 방식으로, 읽기는 동시에 여럿, 쓰기는 한 번에 하나를 허용)와 busy timeout 5초입니다. `go test -race -count=20`으로 반복해도 모두 통과했습니다. 반복 실행은 결과가 들쭉날쭉하지 않은지 보는 확인이며, 타이밍에 전혀 기대지 않는다는 증명은 아닙니다.
- 지키는 구현을 일부러 망가뜨리면 테스트가 실패한다는 사실([직접 확인하기](#직접-확인하기) 참고)

**공식 문서로 뒷받침했지만 실행하지 않은 것**

- PostgreSQL READ COMMITTED의 WHERE 재평가, REPEATABLE READ와 SERIALIZABLE의 `40001`, UNIQUE 충돌 시 대기
- MySQL InnoDB에서 UPDATE가 최신 커밋 행에 영향을 주는 동작, 중복 키 공유 락과 그로 인한 데드락
- SQLite가 교착 가능성이 있으면 busy handler를 부르지 않는다는 동작. 실행 결과는 이와 맞았지만, 그 이유는 공식 문서에서 가져왔습니다.

**증명하지 않는 것**

- 서버 여러 대는 한 프로세스 안의 `Server` 값 두 개로 흉내 냈습니다. 실제 프로세스 두 개로는 실행하지 않았습니다.
- 락이 상대를 막았다는 판정(`missed=true`)은 약속 지점의 대기 상한 300밀리초에 기댑니다. 기계가 극도로 느려 상대가 300밀리초 넘게 늦게 도착하면, 락이 없어도 같은 판정이 나올 수 있습니다.
- 한도를 한 행으로 모으는 방식의 효과는 SQLite의 직렬 쓰기와 구분해 확인하지 못했습니다.
- `FakeGateway`는 같은 키의 중복 승인을 걸러 주지 않고, 조회에 없으면 이후에도 승인되지 않는다고 가정합니다. 실제 PG의 동작은 연동 규격마다 다릅니다.
- 처리량과 지연 시간은 측정하지 않았습니다.

## 정리

- 경쟁 조건은 **확인과 쓰기 사이의 틈**에 다른 요청의 확인이 들어올 때 생깁니다. 깨질지는 도착 간격과 틈의 길이를 함께 봐야 알 수 있고, 외부 호출이 틈을 가장 길게 만듭니다.
- 서버 안의 뮤텍스는 서버 한 대 안에서만 순서를 세웁니다. 트랜잭션을 열기만 하면 읽은 행이 잠기지 않으며, 증상은 DB마다 다릅니다. SQLite는 오류를 내고, PostgreSQL과 MySQL은 기본 격리 수준에서 조용히 갱신 손실을 남깁니다.
- 틈을 다루는 방법은 네 가지입니다. **조건부 UPDATE**로 틈을 없애고, **비관적 락**으로 읽은 값을 커밋까지 고정하고, **낙관적 락**으로 충돌을 검출해 재시도하고, 여러 행에 걸친 규칙은 **한 행으로 모읍니다.**
- 갱신 손실은 같은 행을 덮어써서 생기고, 쓰기 스큐는 각 요청이 같은 합계를 읽고 옳게 판단해 서로 다른 새 행을 쓴 결과가, 합쳐졌을 때 여러 행에 걸친 규칙을 깨서 생깁니다. 기존 행에만 거는 락은 새 행 삽입을 막지 못합니다.
- 외부 호출이 끼면 **선점을 먼저** 합니다. 같은 키라도 내용이 다르면 거부하고, 결과를 모르는 응답은 `UNKNOWN`으로 남겨 **대사로 확정**합니다. 상태 전이는 이전 상태를 조건으로 거는 UPDATE로 해야 충돌이 드러나고, 시간 기준만으로 막을 수 없는 경우는 PG 쪽 멱등성이 막아야 합니다.
- 테스트는 결과 숫자만으로 판정하면 부족합니다. **예상하지 못한 오류가 0건인지**와 **락이 실제로 상대를 막았는지**를 함께 확인해야 합니다.

### 복습 질문

1. 요청 B가 A보다 100밀리초 늦게 왔는데도 이중 승인이 나는 조건은 무엇인가요? 10밀리초 늦게 왔는데도 이중 승인이 나지 않는 조건은 무엇인가요?
2. 잔액 10,000원에서 10,000원 출금 두 건으로 갱신 손실을 테스트하면 무엇이 문제인가요?
3. 같은 Read-Modify-Write 코드를 트랜잭션 안에 넣었을 때 SQLite와 PostgreSQL READ COMMITTED의 증상은 각각 어떻게 다르고, 그 이유는 무엇인가요?
4. 한도 규칙에서 기존 `charges` 행에 행 락을 걸어도 쓰기 스큐가 막히지 않는 이유는 무엇이고, 대신 무엇을 잠그면 막을 수 있나요?
5. PG 타임아웃을 `FAILED`로 기록하면 어떤 순서로 사고가 나나요? `stuckAfter`를 충분히 길게 잡아도 남는 사고는 무엇이고, 무엇이 그것을 막나요?

## 직접 확인하기

실행 방법과 파일 구성은 [Go 예제 README](examples/concurrency_transaction/go/README.md)에 있습니다.

```bash
cd database/examples/concurrency_transaction/go
go test -v ./...
```

올바른 코드를 일부러 망가뜨리고 테스트가 잡는지 보면 오래 기억에 남습니다. 아래는 이 문서를 쓰면서 실제로 망가뜨려 본 코드와, 그때 실패한 테스트입니다.

| 코드 변경 | 실패한 테스트 | 실패 신호 |
| --- | --- | --- |
| `inImmediateTx`의 `BEGIN IMMEDIATE`를 `BEGIN`으로 | `TestWithdrawLocked_SerializesRequests` | `missed=false success=1 balance=15000` |
| `PayOnce`에서 결과 모름을 `StatusFailed`로 | `TestPayOnce_TimeoutAfterApprovalIsUnknownUntilReconciled` | `want UNKNOWN, got status=FAILED` |
| `PayOnce`의 계좌, 금액 비교 제거 | `TestPayOnce_SameKeyDifferentRequestIsRejected` | `want ErrKeyReused ..., got err=<nil>` |
| `PayOnce`에서 선점 실패를 무시하고 진행 (다섯 요청이 모두 PG를 부르고 `PENDING`을 결과로 바꾸려 하지만, 첫 요청의 UPDATE만 1행이고 나머지는 status가 이미 `PENDING`이 아니라 0행) | `TestPayOnce_SameKeyConcurrentChargesOnce` | `payment state changed by another actor` |
| `PayOnce`의 결과 기록에 요청 ctx를 그대로 사용 | `TestPayOnce_CallerDeadlineDuringApprovalIsUnknown` | `record UNKNOWN: context deadline exceeded` |
| `finish`의 `status = ?` 조건 제거 | `TestReconcile_TooEarlyConflictsWithInFlightPayment` | `want ErrStateChanged, got <nil>` |
| `WithdrawConditional`의 `balance >= ?` 제거 | `TestWithdrawConditional_RejectsOverdraft` | `CHECK constraint failed: balance >= 0` |
| `WithdrawOptimistic`의 `version = ?` 제거 | `TestWithdrawOptimistic_RetriesAfterConflict` | 두 요청 모두 `attempts=1` |
| `ChargeReserved`의 한도 UPDATE를 트랜잭션 밖에서 실행 | `TestChargeReserved_RollsBackReservationWhenInsertFails` | `사용액이 롤백되지 않았습니다: used=4000` |
| `ChargeReserved`의 내역 INSERT를 트랜잭션 밖에서 실행 | `TestChargeReserved_KeepsLimit` | `success=0 ... other=[reserve limit: database is locked ...]` |

바꾼 코드는 반드시 되돌립니다. 테스트가 실패하지 않는 변경을 찾았다면 두 가지 중 하나입니다. 그 변경이 실제로 무해하거나, 테스트가 그 성질을 확인하지 않는 것입니다.

어떤 테스트가 실패했더라도, 같은 변경에서 **엉뚱한 이유로 통과한 다른 테스트**가 없는지도 봐야 합니다. 표의 마지막 행 변경에서 한도 UPDATE는 트랜잭션 `tx` 안에서 실행돼 SQLite 쓰기 락을 쥐고, 내역 INSERT는 `tx`가 아니라 `db`(풀의 다른 연결)로 실행됩니다. 그 INSERT는 같은 함수의 `tx`가 쥔 쓰기 락을 busy timeout 5초 동안 기다리다 `database is locked`로 실패하고, 함수가 오류를 돌려주면서 `tx`는 롤백됩니다. 그래서 `TestChargeReserved_RollsBackReservationWhenInsertFails`는 5초 뒤 사용액 0원으로 통과했지만, 사용액을 되돌린 이유는 주입한 트리거가 아니라 이 락 대기 실패였습니다. 이 변경을 실제로 잡은 것은 `TestChargeReserved_KeepsLimit`입니다. 동시에 들어온 다른 요청들이 그 5초 동안 한도 UPDATE에서 기다리다 `reserve limit: database is locked`로 실패했습니다(오류 문구에서 추론한 순서이며 따로 추적하지는 않았습니다). 또 컴파일이 실패해서 테스트가 멈춘 경우는 검출로 세면 안 됩니다. 이렇게 어느 쪽인지 따져 보는 것 자체가 좋은 복습입니다.

## 근거

- SQLite, [Transaction](https://www.sqlite.org/lang_transaction.html): 기본 BEGIN은 DEFERRED이고, 읽기 트랜잭션이 쓰기 트랜잭션으로 전환(upgrade)하지 못하면 `SQLITE_BUSY`를 반환한다는 설명
- SQLite, [Result and Error Codes](https://www.sqlite.org/rescode.html): `SQLITE_BUSY`, `SQLITE_BUSY_SNAPSHOT`(517), `SQLITE_CONSTRAINT_UNIQUE`(2067), BEGIN IMMEDIATE 이후 BUSY가 나지 않는다는 보장
- SQLite, [Register A Callback To Handle SQLITE_BUSY Errors](https://www.sqlite.org/c3ref/busy_handler.html): 교착 가능성이 있으면 busy handler를 부르지 않고 `SQLITE_BUSY`를 돌려준다는 설명
- PostgreSQL, [Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html): READ COMMITTED의 WHERE 재평가, REPEATABLE READ와 SERIALIZABLE의 `40001`과 재시도 안내
- PostgreSQL, [Index Uniqueness Checks](https://www.postgresql.org/docs/current/index-unique-checks.html): 커밋되지 않은 충돌 행이 있으면 삽입하려는 쪽이 기다린다는 설명
- MySQL 8.4, [Locks Set by Different SQL Statements in InnoDB](https://dev.mysql.com/doc/refman/8.4/en/innodb-locks-set.html): INSERT의 배타 락, 중복 키 시 공유 락과 세 세션 데드락 예시
- MySQL 8.4, [Consistent Nonlocking Reads](https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html): 스냅숏은 SELECT에 적용되고 DML에는 반드시 적용되지는 않는다는 설명
- Martin Kleppmann, [How to do distributed locking](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html): 락 만료와 펜싱 토큰에 대한 비판
