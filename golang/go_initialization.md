# Go initialization

- [Go initialization](#go-initialization)
    - [The `init` function](#the-init-function)
    - ['모든 변수 선언들이 초기화 값을 평가한 후'란?](#모든-변수-선언들이-초기화-값을-평가한-후란)
    - [예제](#예제)
        - [여러 init이 있는 경우](#여러-init이-있는-경우)
        - [`sqltrace.Register`를 `init`에서 수행하는 예제](#sqltraceregister를-init에서-수행하는-예제)

## [The `init` function](https://go.dev/doc/effective_go#init)

Go 언어에서 `init` 함수는 패키지가 가져올(import) 때 자동으로 한 번 호출되는 특수 함수입니다.
이 함수는 주로 *패키지 초기화*를 위한 설정을 담당하며, 패키지의 다른 부분이 실행되기 전에 필요한 설정을 수행합니다.
`init` 함수는 해당 패키지 내에서만 호출되며, 패키지 사용자는 이 함수를 직접 호출할 수 없습니다.
또한, 해당 패키지가 import 되지 않는다면 init 함수는 호출되지 않습니다.

> Finally, each source file can define its own niladic `init` function to set up whatever state is required. (Actually each file can have multiple `init` functions.)
> And finally means finally: `init` is called after all the variable declarations in the package have evaluated their initializers,
> and those are evaluated only after all the imported packages have been initialized.
>
> Besides initializations that cannot be expressed as declarations, a common use of `init` functions is to verify or repair correctness of the program state before real execution begins.
>

마지막으로, 각 소스 파일은 필요한 상태를 설정하기 위해 자체적인 매개변수가 없는 `init` 함수를 정의할 수 있습니다.
(사실 각 파일은 여러 개의 `init` 함수를 가질 수 있습니다.)
그리고 마지막으로, `init` 함수는 *모든 변수 선언들이 초기화 값을 평가한 후*에 호출됩니다.
이러한 초기화는 모든 가져온 패키지가 초기화된 후에만 평가됩니다.

> - `niladic` 함수: 인수가 없는 함수
> - `monadic` 함수: 하나의 인수를 가지는 함수
> - `dyadic` 함수: 두 개의 인수를 가지는 함수

선언으로 표현할 수 없는 초기화(즉, 선언만으로는 표현할 수 없는 복잡한 초기화 작업) 외에도,
`init` 함수의 일반적인 용도는 실제 실행이 시작되기 전에 프로그램 상태의 정확성을 확인하거나 수정하는 것입니다.

```go
func init() {
    if user == "" {
        log.Fatal("$USER not set")
    }
    if home == "" {
        home = "/home/" + user
    }
    if gopath == "" {
        gopath = home + "/go"
    }
    // gopath may be overridden by --gopath flag on command line.
    flag.StringVar(&gopath, "gopath", gopath, "override default GOPATH")
}
```

## '모든 변수 선언들이 초기화 값을 평가한 후'란?

"모든 변수 선언들이 초기화 값을 평가한 후에"라는 문구는 Go 언어에서 *변수 선언 시 초기화 값을 평가한 후*를 의미합니다.

- 초기화 값(initializer):

    Go 언어에서 변수를 선언할 때 초기 값을 함께 지정할 수 있습니다.
    초기화 값은 변수를 선언할 때 할당되는 초기 값을 의미합니다.
    이 초기화 값은 *프로그램이 실행될 때* 평가됩니다.

    ```go
    var x = 10
    //      ^^ 초기화 값. 변수 선언 시점에 평가(evaluate)되어 변수에 할당됩니다.
    ```

- 평가(evaluate)

    변수를 선언할 때 지정된 초기화 값을 실제로 계산하거나 설정하는 과정을 의미합니다.
    모든 변수 선언이 초기화 값을 평가한 후에 init 함수가 호출됩니다.
    초기화 값 평가 과정은 *프로그램이 실행될 때* 변수에 초기 값을 할당하는 중요한 단계입니다.

다음은 변수 선언과 초기화 값을 평가하는 과정을 보여주는 예시 코드입니다:

```go
package main

import (
    "fmt"
)

// `a`와 `b` 변수는 각각 `initializeA`와 `initializeB` 함수를 호출하여 초기화됩니다.
var (
    // `initializeA` 함수가 호출되어 "Initializing a"를 출력하고, `a` 변수에 `1`을 할당합니다.
    a = initializeA()
    // `initializeB` 함수가 호출되어 "Initializing b"를 출력하고, `b` 변수에 `2`를 할당합니다.
    b = initializeB()
)

func initializeA() int {
    fmt.Println("Initializing a")
    return 1
}

func initializeB() int {
    fmt.Println("Initializing b")
    return 2
}

// 모든 변수 선언이 초기화 값을 평가한 후에 `init` 함수가 호출됩니다.
// 첫 번째 `init` 함수가 호출되어 "첫 번째 init 함수 실행"을 출력합니다.
func init() {
    fmt.Println("첫 번째 init 함수 실행")
}

// 두 번째 `init` 함수가 호출되어 "두 번째 init 함수 실행"을 출력합니다.
func init() {
    fmt.Println("두 번째 init 함수 실행")
}

// 모든 `init` 함수가 실행된 후, `main` 함수가 호출됩니다.
// `main` 함수는 "main 함수 실행"을 출력하고, 변수 `a`와 `b`의 값을 출력합니다.
func main() {
    fmt.Println("main 함수 실행")
    fmt.Println("a:", a)
    fmt.Println("b:", b)
}
```

실행 결과는 다음과 같습니다.

```sh
Initializing a
Initializing b
첫 번째 init 함수 실행
두 번째 init 함수 실행
main 함수 실행
a: 1
b: 2
```

## 예제

### 여러 init이 있는 경우

```go
package sub

import (
    "fmt"
)

// 첫 번째 init 함수
func init() {
    fmt.Println("multiple_init 첫 번째 init 함수 실행")
    // 초기화 작업 1
}

// 두 번째 init 함수
func init() {
    fmt.Println("multiple_init 두 번째 init 함수 실행")
    // 초기화 작업 2
}

// 세 번째 init 함수
func init() {
    fmt.Println("multiple_init 세 번째 init 함수 실행")
    // 초기화 작업 3
}
```

```bash
❯ go run .
multiple_init 첫 번째 init 함수 실행
multiple_init 두 번째 init 함수 실행
multiple_init 세 번째 init 함수 실행
```

### `sqltrace.Register`를 `init`에서 수행하는 예제

```go
func init() {
    sqltrace.Register(
        driverName, 
        &mysql.MySQLDriver{}, 
        sqltrace.WithServiceName(TraceServiceName),
    )
}
```

- `init` 함수에서 `sqltrace.Register`를 호출하는 이유

    **자동 초기화 보장**:

    `init` 함수를 사용하면, 프로그램이 시작할 때 자동으로 SQL 추적 기능이 설정됩니다.
    이는 개발자가 별도로 추적을 활성화할 필요가 없게 만들어, 오류를 줄이고 일관성을 유지하는 데 도움이 됩니다.

    **종속성 관리**:

    `sqltrace.Register`는 데이터베이스 드라이버와 연동하여 SQL 명령을 추적합니다.
    이러한 설정을 `init`에서 수행하면, 해당 패키지가 import되는 순간 자동으로 설정되어,
    후속 코드에서 데이터베이스 연결 및 쿼리 수행 시 바로 추적이 가능합니다.

    **명시적인 설정 호출 불필요**:

    프로그램의 다른 부분에서 이 설정을 명시적으로 호출할 필요가 없어, 코드의 복잡성이 감소합니다.

- `init` 단계에서 수행하지 않을 경우의 영향

    **수동 설정 필요**:

    만약 `init` 함수에서 이러한 설정을 수행하지 않는다면, 개발자는 애플리케이션의 다른 적절한 위치에서 직접 이를 설정해야 합니다.
    이는 코드의 어떤 부분에서는 추적이 활성화되지 않을 수 있는 리스크를 안고 있습니다.

    **코드 중복**:

    다양한 위치에서 데이터베이스 연결을 설정할 때마다 추적 기능도 함께 활성화해야 할 수 있습니다.
    이로 인해 코드 중복이 발생하고, 오류 발생 가능성이 증가할 수 있습니다.

    **실행 순서에 따른 오류**:

    `sqltrace.Register`가 호출되기 전에 데이터베이스 연결이 시도될 경우, 초기 연결에서는 SQL 추적이 수행되지 않을 수 있습니다.
    이는 디버깅을 어렵게 만들고, 시스템의 행동을 예측하기 어렵게 할 수 있습니다.
