# refactoring examples

## 동치인 경우 다른 타입으로 리팩토링

### 두 가지 경우의 수만 가진 타입을 bool 매개변수로 치환

아래와 같이 `SomeProperty` 타입을 `bool` 타입으로 변경함으로써 선택지를 줄이고 코드를 단순화할 수 있습니다.

```go
type SomeProperty string

// 단 두 가지 경우의 수만 있음 
const (
  SomePropertyAscending = SomeProperty("ASC")
  SomePropertyDescending = SomeProperty("DESC")
)
func Test(someProperty SomeProperty)
```

```go
// bool 타입으로 대체
func Test(isAscendingbool)
```

타입 이론에서는 *데이터의 종류*와 *그 데이터가 취할 수 있는 값*을 정의합니다.
`SomeProperty` 타입의 경우, 이 타입은 특정한 문자열 값(`"ASC"` 및 `"DESC"`)만을 가질 수 있도록 제한되었습니다.
이를 열거형 타입이라고 할 수 있습니다.

```go
type SomeProperty string

const (
  SomePropertyAscending = SomeProperty("ASC")
  SomePropertyDescending = SomeProperty("DESC")
)
```

이 타입을 `bool`로 리팩토링하면, 두 가지 선택지(`true`와 `false`)만 가지는 불리언 타입으로 변환됩니다.
불리언 타입은 단 두 가지 값만 가지므로, 이전 열거형 타입의 선택지를 불리언 값으로 대체할 수 있습니다.

```go
func Test(isAscending bool)
```

이 리팩토링은 타입 이론에서 타입의 단순화와 안전성의 관점에서 볼 수 있습니다.

- **단순화**: 열거형 타입을 불리언 타입으로 변환하여 선택지를 두 가지로 단순화했습니다. 이는 코드를 더 간결하게 만들고, 이해하기 쉽게 합니다.
- **타입 안전성**: 불리언 타입은 언어 자체에서 제공하는 타입으로, 오타나 잘못된 값을 넣는 실수를 방지할 수 있습니다. 열거형 타입에서 문자열을 사용할 때 발생할 수 있는 오타를 방지할 수 있습니다.

함수형 프로그래밍에서는 불변성, 순수 함수, 고차 함수 등의 개념을 중요시합니다.
여기서 불리언 타입을 사용함으로써 얻을 수 있는 함수형 패러다임의 장점을 몇 가지 살펴보겠습니다.

함수형 프로그래밍에서는 데이터가 불변적이어야 하며, 변형되지 않아야 합니다.
불리언 타입은 매우 간단하며, 데이터의 변형 가능성이 적습니다.
`SomeProperty` 타입이 문자열로 되어 있는 경우, 데이터의 변형 가능성이 존재합니다.

또한 불리언 타입을 사용하면 코드가 더 단순해지고 명확해집니다.
함수형 프로그래밍에서는 코드의 단순성과 명확성을 중요시합니다.
불리언 값은 그 자체로 명확한 의미를 가지며, 추가적인 변환이 필요 없습니다.

그리고 함수형 언어에서는 패턴 매칭을 사용하여 데이터의 다양한 형태를 처리할 수 있습니다.
Go에서는 패턴 매칭이 직접적으로 지원되지 않지만, `if-else` 구문을 통해 비슷한 효과를 얻을 수 있습니다.
불리언 값은 이러한 구문을 더 단순하게 만듭니다.

```go
func Test(isAscending bool) {
    if isAscending {
        fmt.Println("Ascending order")
    } else {
        fmt.Println("Descending order")
    }
}
```

이와 같이, 불리언 값을 사용하면 코드가 더 단순하고 명확해지며, 이는 함수형 패러다임의 원칙과 일치합니다.

동형사상은 두 구조가 *서로 다른 형태를 가지지만, 동일한 정보를 표현할 수 있는 경우*를 말합니다.
프로그래밍에서 이는 두 가지 타입이 같은 정보의 표현이라는 것을 의미합니다.

두 타입의 동형사상 예:
1. `Option<T>` 타입 (예: `Some(T)`와 `None`)은 빈 리스트 또는 단일 요소를 가진 리스트와 동형입니다.
2. `Either<A, B>` 타입 (예: `Left(A)`와 `Right(B)`)은 두 가지 선택지 중 하나를 가지는 타입과 동형입니다.

원래 코드는 `SomeProperty` 타입으로 두 가지 값만 가질 수 있게 정의되었습니다.

```go
type SomeProperty string

const (
  SomePropertyAscending = SomeProperty("ASC")
  SomePropertyDescending = SomeProperty("DESC")
)
```

이 타입은 "ASC"와 "DESC" 문자열 값으로 제한된 열거형 타입입니다.
이를 다음과 같은 불리언 타입으로 변경할 수 있습니다.

```go
func Test(isAscending bool)
```

여기서 `isAscending`이 `true`이면 "ASC", `false`이면 "DESC"를 의미하도록 합니다. 이는 동형사상에 해당합니다.

동형사상의 증명:
- **SomeProperty 타입**: 두 가지 값 "ASC"와 "DESC"를 가집니다.
- **bool 타입**: 두 가지 값 `true`와 `false`를 가집니다.

이 두 타입은 각각 두 가지 값을 가지므로 동형입니다.
즉, 다음과 같은 동형사상 함수를 정의할 수 있습니다.

```go
func somePropertyToBool(sp SomeProperty) bool {
    return sp == SomePropertyAscending
}

func boolToSomeProperty(b bool) SomeProperty {
    if b {
        return SomePropertyAscending
    }
    return SomePropertyDescending
}
```

이 동형사상 함수는 `SomeProperty` 타입과 `bool` 타입이 정보 손실 없이 상호 변환 가능함을 보여줍니다.

함수형 프로그래밍에서는 함수와 타입의 변환이 중요합니다.
여기서 `SomeProperty`와 `bool` 간의 동형사상은 다음과 같은 이점을 제공합니다.

- 불변성 및 패턴 매칭

    불리언 타입을 사용하면, 패턴 매칭이 단순해지고 불변성이 보장됩니다.
    이는 함수형 프로그래밍에서 중요한 원칙입니다.

    ```go
    func Test(isAscending bool) {
        if isAscending {
            fmt.Println("Ascending order")
        } else {
            fmt.Println("Descending order")
        }
    }
    ```

- 코드 단순화

    타입이 단순해짐에 따라 함수 정의도 단순해집니다.
    이는 함수형 프로그래밍에서 코드의 간결성과 명확성을 중요시하는 이유와 일치합니다.
