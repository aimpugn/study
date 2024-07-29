# sentinel

- [sentinel](#sentinel)
    - [Sentinel Error 란?](#sentinel-error-란)
    - [사용 맥락과 역사](#사용-맥락과-역사)
    - [Go에서의 사용](#go에서의-사용)
    - [Go 외의 사용](#go-외의-사용)
    - [장단점](#장단점)

## Sentinel Error 란?

**Sentinel Error**는 소프트웨어 개발에서 *특정 오류 상황을 나타내기 위해 미리 정의된 고정된 오류 객체 또는 값*입니다.
이러한 오류들은 프로그램의 다른 부분에서 쉽게 식별하고 처리할 수 있도록 고유하게 정의되어 있습니다.

Sentinel 오류는 특정 조건이 발생했을 때 반환되는 상수 오류로, 호출자가 이 오류를 받았을 때 해당 오류가 정확히 무엇을 의미하는지 즉시 알 수 있도록 돕습니다.

## 사용 맥락과 역사

- **용어의 기원**

    Sentinel이라는 용어는 원래 군대나 경비에서 '경계병' 또는 '파수꾼'을 의미합니다.
    소프트웨어에서는 특정 조건이나 상황을 '감시'하고 그 상황이 발생했음을 알리는 값 또는 객체로 사용됩니다.

- **소프트웨어에서의 사용**

    Sentinel 값은 종종 리스트나 트리와 같은 데이터 구조에서 경계를 나타내거나, 특정 오류 상황을 식별하는 데 사용됩니다.
    오류 처리에서 Sentinel 오류를 사용하면 프로그램의 다른 부분에서 그 오류를 쉽게 확인하고 적절히 반응할 수 있습니다.

## Go에서의 사용

Go 언어에서 Sentinel 오류는 매우 흔하게 사용되며, 표준 라이브러리 자체도 여러 Sentinel 오류를 제공합니다.

예를 들어, 파일 시스템 작업을 할 때 `io.EOF`는 파일의 끝에 도달했음을 나타내는 Sentinel 오류입니다.

```go
package main

import (
    "errors"
    "fmt"
    "io"
)

// Custom Sentinel Error
var ErrMyCustomError = errors.New("my custom error occurred")

func mightFail(flag bool) error {
    if flag {
        return ErrMyCustomError
    }
    return nil
}

func main() {
    err := mightFail(true)
    if errors.Is(err, ErrMyCustomError) {
        fmt.Println("Handled My Custom Error:", err)
    }
    if errors.Is(err, io.EOF) {
        fmt.Println("Handled EOF Error")
    }
}
```

## Go 외의 사용

Sentinel 오류는 Go언어에만 한정되지 않고, 많은 다른 프로그래밍 언어와 시스템에서 사용됩니다.

예를 들어, Java에서는 `NullPointerException`이나 `ArrayIndexOutOfBoundsException` 등이 Sentinel 오류의 역할을 할 수 있습니다.

이러한 예외들은 특정 상황에서 발생하며, 이를 통해 오류의 원인을 식별하고 적절히 처리할 수 있습니다.

## 장단점

**장점**:
- 오류의 원인을 명확하게 식별할 수 있어 오류 처리가 간결해집니다.
- 코드의 읽기 쉽고 유지보수하기 쉬워집니다.

**단점**:
- 너무 많은 Sentinel 오류는 코드를 복잡하게 만들 수 있습니다.
- 다양한 오류를 고유한 상수로 관리해야 하므로 오류 유형이 많을 때 관리가 어려워질 수 있습니다.
