# sequence diagram

- [sequence diagram](#sequence-diagram)
    - [Messages](#messages)
    - [Examples](#examples)
        - [할인 적용 sequenceDiagram](#할인-적용-sequencediagram)

## Messages

- `->`:

    두 participant 간의 관계를 나타내며, 일반적으로 정보의 흐름이나 연관성을 나타냅니다.

    예를 들어, 시스템 간의 데이터 전송이나, 프로세스 간의 상호작용을 표현할 때 사용됩니다.

- `-->`:

    두 participant 간의 관계를 나타내지만, 일반적으로 약한 연관성을 나타냅니다.

    예를 들어, 시스템 간의 약한 연관성이나, 프로세스 간의 간접적인 상호작용을 표현할 때 사용됩니다.

- `->>`:

    두 participant 간의 관계를 나타내며, 일반적으로 명확한 명령이나 요청을 나타냅니다.

    예를 들어, 시스템 간의 명령이나 요청을 전달하는 경우, 또는 프로세스 간의 명령이나 요청을 전달하는 경우에 사용됩니다.

- `-->>`:

    두 participant 간의 관계를 나타내며, 일반적으로 약한 명령이나 요청을 나타냅니다.

    대개 비동기적 또는 덜 중요한 메시지 전달을 나타냅니다.
    이는 응답이나 결과를 나타낼 때 사용되며, 호출자가 응답을 기다리지 않고 다른 작업을 계속할 수 있음을 의미할 수 있습니다.

    예를 들어, 시스템 간의 약한 명령이나 요청을 전달하는 경우, 또는 프로세스 간의 약한 명령이나 요청을 전달하는 경우에 사용됩니다.

- `-x`:

    두 participant 간의 관계를 나타내며, 일반적으로 오류나 예외 상황을 나타냅니다.

    예를 들어, 시스템 간의 오류 발생이나, 프로세스 간의 예외 상황을 표현할 때 사용됩니다.

- `--x`:

    두 participant 간의 관계를 나타내며, 일반적으로 약한 오류나 예외 상황을 나타냅니다.

    예를 들어, 시스템 간의 약한 오류 발생이나, 프로세스 간의 약한 예외 상황을 표현할 때 사용됩니다.

- `-)`:

    두 participant 간의 관계를 나타내며, 일반적으로 비동기적인 상호작용을 나타냅니다.

    예를 들어, 시스템 간의 비동기적인 데이터 전송이나, 프로세스 간의 비동기적인 상호작용을 표현할 때 사용됩니다.

- `--)`:

    두 participant 간의 관계를 나타내며, 일반적으로 약한 비동기적인 상호작용을 나타냅니다.

    예를 들어, 시스템 간의 약한 비동기적인 데이터 전송이나, 프로세스 간의 약한 비동기적인 상호작용을 표현할 때 사용됩니다.

## Examples

### 할인 적용 sequenceDiagram

```mermaid
sequenceDiagram
    participant Customer as Customer
    participant Client as Client
    participant AwesomeService as AwesomeService
    participant PGs as Payment Gateways

    Customer ->> Client: 상품 구매 시도
    Client ->> AwesomeService: PG사에 대한 결제 요청
    AwesomeService ->> AwesomeService: 프로모션 비용 확보 시도 gRPC
    AwesomeService ->> PGs: 프로모션 비용 확보 된 금액을 차감하여<br>인증 요청

    rect rgb(250, 84, 42)
        PGs -x AwesomeService: 인증 실패
        AwesomeService  ->>  AwesomeService: Recover gRPC
        AwesomeService -x Client: 결제 실패 응답
        Client -x Customer: 상품 구매 실패
        end    

        rect rgb(14, 173, 22)
        AwesomeService ->> Client: 승인 전 Confirm 요청
        alt Confirm process 실패
            rect rgb(250, 84, 42)
            Client -x AwesomeService: Confirm 실패
            AwesomeService ->> AwesomeService: Recover gRPC
            AwesomeService -x Client: 가맹점 요청에 의해 결제를 중단
            Client -x Customer: 상품 구매 실패
            end
        else Confirm process 성공
            Client ->> AwesomeService: Confirm 성공
            AwesomeService ->> PGs: 프로모션 비용 확보 된 금액을 차감하여<br>승인 요청
            alt 승인 실패
                rect rgb(250, 84, 42)
                PGs -x AwesomeService: 승인 실패
                AwesomeService ->> AwesomeService: 결제 실패 처리
                AwesomeService ->> AwesomeService: Recover gRPC
                AwesomeService -x Client: 결제 실패 리디렉션
                Client -x Customer: 상품 구매 실패
                end
            else 승인 성공
                PGs ->> AwesomeService: 승인 성공
                AwesomeService ->> AwesomeService: 결제 성공 처리
                AwesomeService ->> AwesomeService: Paid gRPC
                AwesomeService ->> Client: 결제 성공 리디렉션
                Client ->> Customer: 상품 구매 성공
            end
        end 
    end
```

```mermaid
sequenceDiagram
    participant User as User
    participant System as System
    participant Database as Database

    User ->> System: 자금 확보 요청
    System ->> Database: 자금 확보
    Database- ->> System: 자금 확보 완료

    alt 로직 수행 성공
        System ->> System: 로직 수행
        System ->> Database: 커밋
        Database- ->> System: 커밋 완료
        Note right of System: 커밋 된 경우에만 취소 가능
    else 로직 수행 실패
        System ->> System: 로직 수행
        System ->> Database: 롤백
        Database- ->> System: 롤백 완료
    end
```

```mermaid
flowchart TD
    A[프로모션 비용 확보 시도]
    A --> B{결제}
    B -->|성공| C[할인 확정]
    C --> D[확정된 된 경우에만<br>복구]
    B -->|실패| G[할인 되돌리기]
```
