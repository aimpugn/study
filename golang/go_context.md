# Go Context

- [Go Context](#go-context)
    - [context?](#context)
    - [구현 배경및 메서드](#구현-배경및-메서드)
    - [사용상의 주의사항](#사용상의-주의사항)
        - [함수의 첫 번째 인자로 전달](#함수의-첫-번째-인자로-전달)
        - [Context 값을 전역 변수로 사용하지 않기](#context-값을-전역-변수로-사용하지-않기)
        - [Context를 과도하게 사용하지 않기](#context를-과도하게-사용하지-않기)
        - [취소 가능한 Context 사용 시 자원 정리](#취소-가능한-context-사용-시-자원-정리)
        - [Context의 불변성](#context의-불변성)
        - [Context의 전파](#context의-전파)
    - [`context` 패키지의 메서드 상세](#context-패키지의-메서드-상세)
        - [`context.Context`](#contextcontext)
        - [`context.WithValue`](#contextwithvalue)
            - [사용자 정의 타입을 사용한 컨텍스트 키](#사용자-정의-타입을-사용한-컨텍스트-키)
    - [기타](#기타)
        - [SA1029: should not use built-in type string as key for value; define your own type to avoid collisions (staticcheck)](#sa1029-should-not-use-built-in-type-string-as-key-for-value-define-your-own-type-to-avoid-collisions-staticcheck)
            - [원인과 이유](#원인과-이유)
    - [로깅과 컨텍스트](#로깅과-컨텍스트)
            - [예시 코드](#예시-코드)
            - [최신 자료 및 참고 문서](#최신-자료-및-참고-문서)

## context?

`context` 패키지는 Go 언어에서 동시성을 관리하는 기본적인 방법을 제공한다.
*동시에 수행되는 작업들 간의 실행 시간, 취소 신호, 그리고 기타 요청 스코프의 값들을 관리하고 전달*하기 위해 설계되었다.

- **요청의 취소와 타임아웃 관리**:

    네트워크 요청, 데이터베이스 쿼리, 또는 다른 오래 걸릴 수 있는 작업을 취소할 수 있는 메커니즘을 제공한다.
    이를 통해 자원의 불필요한 사용을 줄이고, 시스템의 응답성을 높일 수 있다.

- **요청 스코프 내의 값 전달**

    요청이 처리되는 동안 필요한 메타데이터나 구성 정보와 같은 값들을 고루틴 사이에서 전달할 수 있다.

- **계층적인 요청 관리**:

    부모 context로부터 파생된 여러 자식 context를 생성할 수 있으며,
    부모 context의 취소 신호는 모든 자식 context에게 전파된다.

## 구현 배경및 메서드

Go의 고루틴은 매우 가볍고, 수천 개의 고루틴이 동시에 실행될 수 있다.
이러한 고루틴들 사이에서 작업의 취소, 타임아웃, 데이터 전달과 같은 메커니즘이 필요했다.
초기에는 각자 다양한 방식으로 이 문제를 해결하려 했지만, `context` 패키지의 도입으로 공식적이고 일관된 방법이 마련되었다.

- **타임아웃 관리**

    `context.WithTimeout`을 사용하여 특정 시간 후에 작업을 자동으로 취소할 수 있다.

- **취소 신호 전달**

    `context.WithCancel`을 사용하여 취소 신호를 수신할 수 있고, 이 신호는 해당 context에서 파생된 모든 context에게 전파된다.

- **데드라인 설정**

    `context.WithDeadline`을 통해 특정 시점이 되면 작업을 취소할 수 있다.

- **값 전달**

    `context.WithValue`를 사용하여 요청 스코프의 값들을 고루틴에 전달할 수 있다.

## 사용상의 주의사항

- **Context 값을 전역 변수로 사용하지 않기**

    context는 요청이나 특정 작업의 스코프에 대한 정보를 담고 있기 때문에, 전역 변수로 사용하는 것은 적절하지 않다.

- **Context를 과도하게 사용하지 않기**

    `context.Value`를 사용하여 너무 많은 데이터를 전달하면, 코드의 가독성과 유지보수성이 떨어질 수 있다.
    필요한 최소한의 정보만 context를 통해 전달하는 것이 좋다.

- **취소 가능한 Context 사용 시 자원 정리**

    context가 취소될 때, 열려 있는 파일, 데이터베이스 연결, 네트워크 요청 등과 같은 자원을 적절히 정리해야 한다.

### 함수의 첫 번째 인자로 전달

`context`를 함수의 첫 번째 매개변수로 전달하는 방식은 context의 명시적인 전달을 강제하며, 이는 코드의 가독성과 관리 측면에서 여러 장점을 갖는다. 이러한 방식은 함수 호출 체인을 통해 context가 전파되도록 하여, 모든 함수가 동일한 context(또는 그 파생된 context)를 공유할 수 있도록 한다. 이는 특히 타임아웃, 취소 신호, 요청 스코프의 데이터 전달과 같은 상황에서 유용하다.

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel() // 메인 함수 종료 시 모든 파생된 context 취소

    // `doWork` 고루틴을 시작할 때 context를 전달
    go doWork(ctx)

    // 어떤 조건에 따라 작업 취소
    cancel()
}

// `doWork` 고루틴은 `main` 함수에서 생성된 context의 취소 신호를 감지할 수 있으며, 작업을 중단할 수 있다.
//  이는 context의 생명주기 관리를 명확하게 하며, 작업의 취소를 용이하게 한다.
func doWork(ctx context.Context) {
    select {
    case <-time.After(5 * time.Second):
        fmt.Println("work completed")
    case <-ctx.Done():
        fmt.Println("work canceled")
    }
}
```

### Context 값을 전역 변수로 사용하지 않기

`context`는 특정 요청이나 작업의 생명주기와 밀접한 관련이 있으므로, 전역 변수로 사용하는 것은 부적절하다.
전역 변수로 사용할 경우, 요청이나 작업의 범위를 넘어서 `context`가 살아남아 버그나 불명확한 동작을 초래할 수 있다.

```go
// 나쁜 예
var globalCtx context.Context // 전역 변수로 context 사용

// 좋은 예
func main() {
    ctx := context.Background()
    doSomething(ctx)
}

func doSomething(ctx context.Context) {
    // ctx를 함수의 인자로 전달
}
```

### Context를 과도하게 사용하지 않기

`context.Value`를 사용하여 너무 많은 정보를 전달하는 것은 권장되지 않는다.
이는 코드의 가독성을 떨어뜨리고, 디버깅을 어렵게 만들 수 있다. 대신, 필요한 정보만 최소한으로 전달하는 것이 좋다.

```go
// 나쁜 예
ctx = context.WithValue(ctx, "userID", "1234")
ctx = context.WithValue(ctx, "authToken", "abcd")

// 좋은 예
type RequestContext struct {
    UserID    string
    AuthToken string
}

// 해당 구조체 인스턴스를 생성하여 함수에 전달하는 것이 더 명확함
```

### 취소 가능한 Context 사용 시 자원 정리

취소 가능한 context(`context.WithCancel`, `context.WithTimeout`, `context.WithDeadline`)를 사용할 때는 context가 취소될 경우, 열린 자원을 적절히 정리해야 한다.

```go
func fetchResource(ctx context.Context, resourceURL string) error {
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, resourceURL, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close() // 자원 정리
    // 응답 처리
    return nil
}
```

### Context의 불변성

`context.Context`는 변경 불가능한 객체로 설계되어 있다.
따라서 새로운 데이터를 추가하거나, 취소 신호를 생성하기 위해서는 기존 context를 변경하는 것이 아니라, 새로운 context 객체를 생성해야 한다.
이는 함수 호출 체인을 통한 데이터의 안정적인 전달과 취소 가능한 작업의 관리를 가능하게 한다.

```go
func processRequest(userID string) {
    // 먼저 루트 context를 생성
    ctx := context.Background()
    // 사용자 ID를 추가한 새로운 `context` 객체를 파생시킨다.
    ctx = context.WithValue(ctx, "userID", userID)

    // 2분의 타임아웃을 가진 새로운 `context` 객체를 파생시킨다.
    // 이를 통해 context의 불변성을 유지하면서 필요한 정보와 제어 메커니즘을 전달한다.
    ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
    defer cancel()

    // 데이터베이스 작업을 수행하는 하위 함수에 context 전달
    fetchUserData(ctx)
}

func fetchUserData(ctx context.Context) {
    // context에서 userID 값을 추출
    userID := ctx.Value("userID").(string)
    fmt.Println("Fetching data for user", userID)

    // 작업 수행...
}
```

### Context의 전파

`context`는 API 경계를 넘어서 고루틴, HTTP 요청, 외부 서비스 호출 등으로 전파되어야 한다.
이는 모든 관련 작업이 동일한 context(또는 파생된 context)를 공유함으로써, 요청의 생명주기와 관련된 정보와 제어를 일관되게 관리할 수 있다.
context를 명시적으로 전달하는 방식은 Go 어플리케이션에서 동시성을 관리하고, 요청의 생명주기를 효과적으로 제어하는 데 중요한 역할을 한다.

```go
func fetchFromExternalService(ctx context.Context, url string) error {
    // 인자로 전달된 context를 전달하여 외부 서비스 호출 시 동일한 context가 사용되도록 한다.
    // 따라서 타임아웃이나 취소 신호가 해당 HTTP 요청에도 영향을 미치도록 한다.
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    // 응답 처리...
    return nil
}
```

## `context` 패키지의 메서드 상세

### `context.Context`

고(Go) 언어에서 `context.Context`는 고루틴 간에 데이터를 전달하기 위한 목적으로 사용된다.
`context.Context`는 요청 처리의 경계, 취소 신호, 그리고 고루틴 사이에서 값들을 안전하게 전달할 수 있는 기능을 제공한다.

### `context.WithValue`

#### 사용자 정의 타입을 사용한 컨텍스트 키

Go 언어에서, 타입은 값의 집합과 해당 값에 대한 연산의 집합을 정의한다.
서로 다른 타입은 서로 다른 값의 집합을 가지며, 서로 할당할 수 없으므로, 타입 안전성을 보장하는 데 도움이 된다.

따라서 사용자 정의 타입을 사용하면, 해당 키는 고유한 타입을 가지게 되고, 컨텍스트 키 자체의 타입 안전성이 보장된다.
사용자 정의 타입을 사용하여 컨텍스트 키를 정의하는 것은 컴파일 시점에 타입 안전성을 향상시키고 네임스페이스 충돌을 방지한다.
단, 같은 값을 가진 다른 키들 사이의 구별은 할 수 없으므로, 런타임에서의 고유성을 보장하는 것은 아니다.
그리고 사용자 정의 타입을 사용하면, 코드의 의도를 더 명확하게 표현할 수 있다.ㄴ

```go
package main

import (
    "context"
    "fmt"
)

const MyKey = "myKey"

// `MyKeyType` 타입의 키는 해당 타입을 사용하는 코드 내에서만 유효하므로, 
// 다른 패키지나 라이브러리와의 충돌 가능성이 줄어든다.
type MyKeyType string

const CustomMyKey MyKeyType = "myKey"

func main() {
    ctx := context.Background()

    // MyKey를 사용하여 값을 저장
    ctx = context.WithValue(ctx, MyKey, "value1")

    // CustomMyKey를 사용하여 값을 저장
    ctx = context.WithValue(ctx, CustomMyKey, "value2")

    // MyKey를 사용하여 값을 검색한다. 이러면 `CustomMyKey`와 연관된 값을 실수로 가져오는 것을 방지할 수 있다.
    value1 := ctx.Value(MyKey)
    fmt.Println("MyKey 값:", value1)

    // CustomMyKey를 사용하여 값을 검색한다.
    value2 := ctx.Value(CustomMyKey) // MyKey 값: value1
    fmt.Println("CustomMyKey 값:", value2) // CustomMyKey 값: value2
}
```

## 기타

### SA1029: should not use built-in type string as key for value; define your own type to avoid collisions (staticcheck)

이 경고 메시지는 Go의 `context` 패키지를 사용할 때 발생하는 일반적인 문제 중 하나를 지적한다.
`context.WithValue` 함수를 사용하여 컨텍스트에 값을 저장할 때, 키로 기본 타입(`string`, `int` 등)을 사용하는 것은 권장되지 않는다.

#### 원인과 이유

`context.WithValue` 함수는 컨텍스트에 키-값 쌍을 저장하는 데 사용된다.
이 함수는 다양한 미들웨어나 서비스 간에 데이터를 전달하는 데 유용하다.
그러나, 키로 기본 타입을 사용하면 다음과 같은 문제가 발생할 수 있다.

Go의 `context.Context`는 요청 처리의 경계, 취소 신호, 그리고 고루틴 사이에서 값들을 안전하게 전달할 수 있는 기능을 제공한다.
이런 컨텍스트를 통해 값을 저장하거나 검색할 때 `context.WithValue` 함수를 사용하게 되는데, 이 함수는 키-값 쌍을 저장하기 위해 인터페이스 타입을 사용한다.

```go
func WithValue(parent Context, key, val interface{}) Context
```

여기서 `key`와 `val`은 모두 `interface{}` 타입이므로, 어떤 타입의 값이든 저장할 수 있다.
하지만, 이런 유연성은 타입 안전성을 저해하고, 예상치 못한 충돌을 발생시킬 수 있다.

이때 기본 타입(`string`, `int` 등)을 키로 사용하면, 다른 패키지 또는 코드의 다른 부분에서 같은 키를 사용할 때 의도치 않은 충돌이 발생할 수 있다. 예를 들어, 두 개발자가 독립적으로 같은 문자열 `"userID"`를 키로 사용하여 다른 목적의 값들을 저장한다면, 이는 예상치 못한 결과를 초래할 수 있다.

## 로깅과 컨텍스트

context를 인자로 넘겨서 로깅하는 방법은 특히 마이크로서비스 아키텍처와 같은 복잡한 시스템에서 유용한 베스트 프랙티스입니다.
이는 여러 가지 이유로 추천되며, 최신 문서와 실례를 통해 그 이유를 설명할 수 있습니다.

- 장점

    1. **추적 가능성**:

        context를 사용하면 각 요청에 대한 고유한 ID를 로깅에 포함시켜 전체 시스템에서 특정 요청의 흐름을 추적할 수 있습니다.
        이는 분산 시스템에서 오류를 디버깅하고 성능 병목을 찾는 데 매우 유용합니다【96†source】.

    2. **일관된 로깅**:

        context를 통해 로깅하면, 모든 로그 메시지에 동일한 컨텍스트 정보를 포함시킬 수 있습니다.
        이는 로그 메시지 간의 일관성을 높이고, 로그를 더 쉽게 분석할 수 있게 합니다

    3. **유연성**:

        context를 사용하면 각 로거 인스턴스에 추가 정보를 쉽게 첨부할 수 있습니다. 이는 로깅 설정을 더 유연하게 만들어, 상황에 맞는 로깅 정보를 동적으로 추가할 수 있습니다【97†source】.

#### 예시 코드

다음은 Go에서 context를 사용하여 로깅하는 예시입니다:

```go
package main

import (
    "context"
    "log"
)

type Logger struct {
    log *log.Logger
}

func (l *Logger) WithContext(ctx context.Context) *log.Logger {
    return log.New(log.Writer(), log.Prefix(), log.Flags())
}

func main() {
    logger := &Logger{log: log.Default()}

    ctx := context.WithValue(context.Background(), "requestID", "12345")
    logWithCtx := logger.WithContext(ctx)

    logWithCtx.Println("This is a log message with context")
}
```

위 코드에서, `Logger` 구조체는 `WithContext` 메서드를 통해 context를 받아 로깅에 추가적인 정보를 포함시킵니다.

#### 최신 자료 및 참고 문서

- **New Relic의 전문가 가이드**: 로그 관리와 컨텍스트를 통한 로깅의 중요성을 강조하며, 이는 복잡한 시스템에서 로그 메시지의 일관성을 유지하고 디버깅을 용이하게 하는 데 도움이 된다고 설명합니다【95†source】.
- **Sematext의 로깅 베스트 프랙티스**: 로그 메시지에 추가적인 메타데이터를 포함하여 더 나은 가시성과 추적 가능성을 제공하는 것이 중요하다고 설명합니다【96†source】.
- **Kubernetes 컨텍스트 로깅**: Kubernetes 1.24 버전에서 도입된 컨텍스트 기반 로깅은 분산 시스템에서 로그를 효율적으로 관리하고 디버깅하는 데 유용한 방법을 제시합니다【97†source】.

이와 같은 자료들은 컨텍스트 기반 로깅이 왜 중요한지, 그리고 이를 통해 어떻게 더 나은 로그 관리를 할 수 있는지에 대해 신뢰할 수 있는 정보를 제공합니다. 이를 통해 동료나 팀원들에게 컨텍스트 기반 로깅의 중요성을 설명하고, 이를 실천할 수 있는 방법을 제시할 수 있습니다.
