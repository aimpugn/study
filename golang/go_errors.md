# Go Errors

- [Go Errors](#go-errors)
    - [efficientgo error wrap](#efficientgo-error-wrap)
        - [Comparison with errors in switch statements fails on wrapped errors](#comparison-with-errors-in-switch-statements-fails-on-wrapped-errors)
    - [`wrapcheck` lint](#wrapcheck-lint)
        - [error returned from external package is unwrapped](#error-returned-from-external-package-is-unwrapped)
        - [왜 이걸 체크하는지](#왜-이걸-체크하는지)
        - [대처하지 않으면 어떤 이슈가 발생하나요?](#대처하지-않으면-어떤-이슈가-발생하나요)
        - [이런 경우에는 어떻게 wrap을 해야 하는 건가요?](#이런-경우에는-어떻게-wrap을-해야-하는-건가요)

## efficientgo error wrap

### Comparison with errors in switch statements fails on wrapped errors

```go
if err != nil {
    switch errors.Cause(err) {
                  ^^^^^^ Comparison with errors in switch statements fails on wrapped errors
    case domain.ErrExceedingSearchPeriod:
        return fiber.NewError(fiber.StatusBadRequest, "최대 90일간 검색이 가능합니다")
    case domain.ErrInvalidSearchPeriod:
        return fiber.NewError(fiber.StatusBadRequest, "유효하지 않은 검색기간입니다.")
    case domain.ErrUnknownPaymentStatus:
        return fiber.NewError(fiber.StatusBadRequest, "지원되지 않는 상태값입니다.")
    case domain.ErrUnknownSortProperty:
        return fiber.NewError(fiber.StatusBadRequest, "지원되지 않는 정렬규칙입니다.")
    default:
        return fiber.NewError(fiber.StatusInternalServerError, "알 수 없는 에러")
    }
}
```

Go 언어에서 "Comparison with errors in switch statements fails on wrapped errors"라는 경고가 발생하는 것은, Go의 `error` 타입이 인터페이스로 정의되어 있기 때문에 발생할 수 있습니다.

특히, 이 문제는 Go 1.13 버전에서 도입된 에러 래핑(wrapping) 기능과 관련이 있습니다.

> **에러 래핑(Error Wrapping)**
>
> Go 1.13 이상에서는 `%w` 포맷 지시어를 사용하여 `fmt.Errorf`를 통해 에러를 래핑할 수 있습니다.
> 이를 통해 하위 에러를 포함하는 새로운 에러를 생성할 수 있으며, 이러한 에러들은 `errors.Unwrap`, `errors.Is`, `errors.As` 함수를 사용하여 검사할 수 있습니다.

에러를 래핑한 후에, 이러한 래핑된 에러들을 스위치 문에서 직접 비교하려고 할 때 문제가 발생할 수 있습니다.
래핑된 에러는 원본 에러를 내부적으로 포함하고 있기 때문에, 단순한 비교 연산으로는 원본 에러를 정확히 식별할 수 없습니다.

예를 들어, 다음과 같은 코드가 있을 때 문제가 발생할 수 있습니다:

```go
err := fmt.Errorf("an error occurred: %w", io.ErrUnexpectedEOF)

switch err {
case io.ErrUnexpectedEOF:
    fmt.Println("The error is ErrUnexpectedEOF")
default:
    fmt.Println("Unknown error")
}
```

위 코드에서 `switch` 문은 예상대로 작동하지 않을 것입니다. `err`는 `io.ErrUnexpectedEOF`를 내부적으로 래핑하고 있지만, 직접적인 비교에서는 `false`로 평가되기 때문입니다.

래핑된 에러들을 처리할 때는 `errors.Is` 함수를 사용하여 에러를 확인해야 합니다. 이 함수는 래핑 체인을 따라가면서 주어진 에러 타입이 존재하는지 확인합니다. 위의 예를 `errors.Is`를 사용하여 수정하면 다음과 같습니다:

```go
if errors.Is(err, io.ErrUnexpectedEOF) {
    fmt.Println("The error is ErrUnexpectedEOF")
} else {
    fmt.Println("Unknown error")
}
```

이렇게 `errors.Is`를 사용하면 래핑된 에러의 내부를 정확히 검사할 수 있으며, 원하는 에러 타입과 일치하는지 확인할 수 있습니다.

## `wrapcheck` lint

### error returned from external package is unwrapped

`wrapcheck`는 Go 코드에서 에러를 반환할 때 적절히 감싸지 않은 경우를 찾아내는 도구다.
이 도구는 에러 처리를 더 명확하고 디버깅을 용이하게 하기 위해 사용된다.

### 왜 이걸 체크하는지

`wrapcheck`는 에러를 반환할 때 추가적인 정보를 제공하기 위해 에러를 감싸는 것을 권장한다.
- 에러의 출처를 더 명확하게 함
- 에러 처리 과정에서 문제를 더 쉽게 추적할 수 있게 함

특히, 여러 패키지와 함수를 거치는 복잡한 에러 처리 과정에서는 에러의 원래 위치를 추적하기 어려울 수 있다.
따라서, 에러를 감싸는 것은 에러 처리를 더 효율적이고 안전하게 만든다.

### 대처하지 않으면 어떤 이슈가 발생하나요?

에러를 적절히 감싸지 않으면, 에러의 출처를 파악하기 어려워질 수 있습니다. 이는 디버깅을 더 어렵게 만들고, 문제의 원인을 찾는 데 시간을 더 소비하게 됩니다. 또한, 에러 처리 로직이 복잡해질 수 있으며, 코드의 가독성과 유지보수성을 저하시킬 수 있습니다.

### 이런 경우에는 어떻게 wrap을 해야 하는 건가요?

Go에서 에러를 감싸는 방법은 여러 가지가 있습니다. 가장 일반적인 방법은 `fmt.Errorf`를 사용하여 에러 메시지에 추가적인 정보를 추가하는 것입니다. 이를 통해 에러의 출처를 명확하게 할 수 있습니다.

예를 들어, 위의 코드에서 `ctx.Next()` 함수가 에러를 반환하는 경우, 다음과 같이 에러를 감싸는 것이 좋습니다:

```go
if err := ctx.Next(); err != nil {
    return fmt.Errorf("failed to call Next: %w", err)
}
```

여기서 `%w`는 에러를 감싸는 특별한 형식 지정자로, 에러를 감싸는 것을 명시적으로 나타냅니다. 이렇게 하면, 에러의 출처를 더 명확하게 파악할 수 있고, 디버깅 과정을 더 효율적으로 만들 수 있습니다.

이 이슈는 에러 처리와 관련된 것으로, Go 언어에서 에러를 적절히 처리하고 전파하는 것의 중요성을 강조하고 있습니다. 이 경우, `ctx.Next()` 함수에서 반환된 에러가 처리되지 않고 그대로 반환되고 있습니다.

에러를 적절히 처리하지 않으면 다음과 같은 문제가 발생할 수 있습니다:

에러를 적절히 래핑하려면 `fmt.Errorf`나 `errors.Wrap`과 같은 함수를 사용하여 에러에 추가 정보를 제공할 수 있습니다. 예를 들어:

```go
import (
    "fmt"
    "github.com/pkg/errors"
)

func someFunction() error {
    // ...
    if err := ctx.Next(); err != nil {
        return errors.Wrap(err, "failed to process request")
    }
    // ...
}
```

위 예제에서, `errors.Wrap` 함수를 사용하여 `ctx.Next()`에서 반환된 에러를 래핑하고, 추가 정보("failed to process request")를 제공했습니다. 이렇게 하면 에러의 맥락을 유지하면서 에러를 반환할 수 있습니다.

또한, Go 1.13 버전부터는 `fmt.Errorf`에 `%w` 동사(verb)를 사용하여 에러를 래핑할 수 있습니다:

```go
if err := ctx.Next(); err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}
```

이렇게 에러를 래핑하면, 에러 정보를 손실하지 않으면서 에러 추적과 처리를 용이하게 할 수 있습니다. 또한, 일관된 에러 처리 방식을 적용하여 코드의 가독성과 유지보수성을 향상시킬 수 있습니다.
