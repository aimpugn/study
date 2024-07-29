# control

## if ~ else if

### 첫 번째 코드 스니펫

조건들을 완전히 독립적으로 평가합니다.

```go
if ... {
    return
}
if ... {
    return
}
if ... {
    return
}
```

- 독립적인 `if`문들을 사용하여 각 조건을 검사합니다.
- 각 `if`문은 다른 `if`문들과 구조적으로 독립적이며, `else` 구문 없이 연속으로 나열됩니다.
- 한 `if`문의 조건이 참인 경우, 해당 블록 내의 `return` 문으로 인해 함수에서 즉시 탈출합니다. 다음 `if`문은 평가되지 않습니다.
- `return` 문의 사용으로 인해, 실제 실행 흐름에서는 `else if` 구문을 사용한 경우와 유사한 조건 검사 흐름을 가집니다.

### 두 번째 코드 스니펫

구조적으로 연결된 조건들을 평가합니다.

```go
if ... {
    return
} else if ... {
    return
} else if ... {
    return
}
```

- 첫 번째 `if` 다음에 `else if` 구문을 사용하여, 첫 번째 조건이 거짓인 경우에만 다음 조건을 검사합니다.
- `else if`는 구조적으로 이전 조건과 연결되어 있으며, 이전 조건이 거짓일 때만 다음 조건이 평가됩니다.
- 여기서도 각 조건 블록 내의 `return` 문은 조건이 참인 경우 함수에서의 조기 탈출을 의미합니다.

## switch ~ case

### when type alias

```go
package main

import "fmt"

type myType string

const (
    test  myType = "test"
    test2 myType = "test2"
    test3 myType = "test3"
)

func main() {
    tmp := "test3"
    switch myType(tmp) { // `tmp` 변수의 `string` 값을 `myType` 타입으로 변환
    case test: // `myType` 타입의 값과 `case` 절에 지정된 `myType` 타입의 상수 값을 비교
        fmt.Println("1")
    case test2:
        fmt.Println("2")
    case test3:
        fmt.Println("3")
    default:
        fmt.Println(myType(tmp))
    }
}
```

Go 언어에서 `switch` 문은 표현식의 값을 평가하고, 해당 값이 `case` 절에 명시된 값들과 일치하는지를 비교합니다.
여기서 `myType`은 `string`의 별칭 타입이므로, `myType(tmp)` 표현식은 `tmp` 변수의 `string` 값을 `myType` 타입으로 변환합니다.
이후 `switch` 문은 `myType` 타입의 값과 `case` 절에 지정된 `myType` 타입의 상수 값을 비교하게 됩니다.

이 경우, `tmp`의 값이 "test3"이므로, `myType(tmp)`는 `test3` 상수와 일치합니다.
따라서 `switch` 문의 해당 `case` 절이 실행되고, 결과적으로 "3"이 출력됩니다.
