# handle error

## 특정 rpc에 대해 특정 에러 경우 클라이언트에 에러 응답해야 할 경우

에러 클래스에, 특정 메서드에서 특정 에러 발생한 경우, 클라이언트에 에러를 반환해야 하는지 여부를 판단하는 메서드를 추가하려고 합니다.
- grpc는 여럿있지만, 각 rpc마다 응답해주는 에러는 공통일 수 있습니다.
- 그래서 A rpc 호출해서 아래 타입의 에러가 발생하면 에러를 응답할 수 있지만, B rpc 호출해서 아래 타입의 에러가 발생하면 에러 응답하지 않을 수 있습니다.

```php
public function ????()
{
    return in_array($this->type, [
        self::ERR_TOTAL_AMOUNT_LESS_THAN_DISCOUNT_AMOUNT,
        self::ERR_PROMOTION_INVALID,
        self::ERR_PROMOTION_NOT_FOUND,
    ]);
}
```

1. 우선 이런 식으로 에러 체크하는 게 괜찮은 설계 방식일까요?
2. 이렇게 한다면, `????` 에 들어갈 적절한 메서드 이름은 어떤 것들이 있을까요? 참고로 에러를 응답할 에러 타입은 추가되거나 줄어들 수 있습니다.

### 1. 에러 체크 방식의 적절성

이런 방식의 에러 체크는 몇 가지 장단점이 있습니다:

장점:
- 간단하고 직관적인 구현
- 에러 유형을 중앙에서 관리할 수 있음
- 특정 RPC 호출에 대한 에러 처리를 쉽게 커스터마이즈할 수 있음

단점:
- 에러 유형이 늘어날수록 관리가 복잡해질 수 있음
- 모든 RPC 호출에 대해 동일한 메서드를 사용하므로, 컨텍스트에 따른 세밀한 제어가 어려울 수 있음
- Open-Closed Principle을 위반할 수 있음 (새로운 에러 유형이 추가될 때마다 메서드를 수정해야 함)

### 대안적 접근 방식

#### 1. 전략 패턴 (Strategy Pattern) 사용

```php
interface ErrorResponseStrategy {
    public function shouldRespond(Error $error): bool;
}

class DefaultErrorResponseStrategy implements ErrorResponseStrategy {
    public function shouldRespond(Error $error): bool {
        return in_array($error->getType(), [
            Error::ERR_TOTAL_AMOUNT_LESS_THAN_DISCOUNT_AMOUNT,
            Error::ERR_PROMOTION_INVALID,
            Error::ERR_PROMOTION_NOT_FOUND,
        ]);
    }
}

class Error {
    private ErrorResponseStrategy $responseStrategy;

    public function setResponseStrategy(ErrorResponseStrategy $strategy) {
        $this->responseStrategy = $strategy;
    }

    public function shouldRespond(): bool {
        return $this->responseStrategy->shouldRespond($this);
    }
}
```

이 방식을 사용하면 다양한 RPC 호출에 대해 서로 다른 전략을 적용할 수 있습니다.

#### 2. 데코레이터 패턴 (Decorator Pattern) 사용

```php
interface ErrorResponse {
    public function shouldRespond(): bool;
}

class BaseErrorResponse implements ErrorResponse {
    protected Error $error;

    public function __construct(Error $error) {
        $this->error = $error;
    }

    public function shouldRespond(): bool {
        return false;
    }
}

class PromotionErrorResponse extends BaseErrorResponse {
    public function shouldRespond(): bool {
        return in_array($this->error->getType(), [
            Error::ERR_PROMOTION_INVALID,
            Error::ERR_PROMOTION_NOT_FOUND,
        ]);
    }
}

class AmountErrorResponse extends BaseErrorResponse {
    public function shouldRespond(): bool {
        return $this->error->getType() === Error::ERR_TOTAL_AMOUNT_LESS_THAN_DISCOUNT_AMOUNT;
    }
}
```

이 방식을 사용하면 에러 유형별로 응답 로직을 분리할 수 있습니다.

#### 3. 열거형 (Enum) 사용 (PHP 8.1 이상)

```php
enum ErrorResponseType {
    case RESPOND;
    case IGNORE;
}

class Error {
    public function getResponseType(): ErrorResponseType {
        return match($this->type) {
            self::ERR_TOTAL_AMOUNT_LESS_THAN_DISCOUNT_AMOUNT,
            self::ERR_PROMOTION_INVALID,
            self::ERR_PROMOTION_NOT_FOUND => ErrorResponseType::RESPOND,
            default => ErrorResponseType::IGNORE
        };
    }

    public function shouldRespond(): bool {
        return $this->getResponseType() === ErrorResponseType::RESPOND;
    }
}
```

이 방식은 타입 안정성을 제공하며, 에러 응답 유형을 명확하게 정의할 수 있습니다.
