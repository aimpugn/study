# wire

- [wire](#wire)
    - [Wire?](#wire-1)
    - [Dependency injection?](#dependency-injection)
    - [Go에서의 자체적인 DI 방법](#go에서의-자체적인-di-방법)
    - [Wire를 사용하는 이유](#wire를-사용하는-이유)
        - [사용시 장점](#사용시-장점)
            - [1. **자동화된 의존성 관리**](#1-자동화된-의존성-관리)
            - [2. **컴파일 타임 검증**](#2-컴파일-타임-검증)
        - [실질적인 예](#실질적인-예)
    - [vs Uber's Dig](#vs-ubers-dig)
        - [사용 방법](#사용-방법)
        - [차이점](#차이점)

## [Wire](https://github.com/google/wire)?

> Wire is a code generation tool that automates connecting components using dependency injection.
> Dependencies between components are represented in `Wire` as function parameters, encouraging explicit initialization instead of global variables.
> Because `Wire` operates without runtime state or reflection, code written to be used with `Wire` is useful even for hand-written initialization.

Google Wire는 `wire` 명령을 사용하여 **의존성 주입 코드를 생성**한다.
`wire` CLI를 실행하면 `wire_gen.go` 파일이 추가되고 필요한 모든 의존성을 주입하는 코드가 자동으로 작성된다.

## Dependency injection?

객체가 다른 객체의 의존성(즉, 필요한 객체)을 직접 생성하지 않고, 외부(생성자, 팩토리, 등록기 등)에서 받아 사용하는 방법을 말한다. 이 패턴은 객체 간의 결합도를 낮추어 코드의 유지보수와 테스트가 쉬워지게 만든다.

## Go에서의 자체적인 DI 방법

Go에서는 생성자 함수를 통해 간단하게 DI를 구현할 수 있다

```go
package main

import "fmt"

// `MessageService` provides operations on messages.
type MessageService interface {
    Send(message string) error
}

// EmailService is an implementation of `MessageService`.
type EmailService struct {}

func (e EmailService) Send(message string) error {
    fmt.Println("Sending email:", message)
    return nil
}

// Notification is a consumer of MessageService that depends on an implementation of the service.
type Notification struct {
    MessageService
}

// NewNotification creates a new Notification with the given `MessageService`.
func NewNotification(service MessageService) *Notification {
    return &Notification{MessageService: service}
}

func main() {
    emailService := EmailService{}
    notification := NewNotification(emailService)
    notification.Send("Hello World!")
}
```

## Wire를 사용하는 이유

위와 같은 방법으로 간단한 의존성을 관리할 수 있지만, 애플리케이션이 커지고 의존성이 복잡해질수록, 수동으로 모든 의존성을 관리하는 것은 오류가 발생하기 쉽고, 코드 유지보수가 어려워질 수 있다.  Wire는 컴파일 타임에 의존성을 해결하므로, 실행 시간에 발생할 수 있는 오류를 예방하고, 의존성 관리 코드를 자동으로 생성하여 개발자의 부담을 줄여준다.

### 사용시 장점

#### 1. **자동화된 의존성 관리**

- **문제 상황**

    애플리케이션이 커짐에 따라 수많은 서비스와 컴포넌트가 서로 의존성을 가지게 된다.
    예를 들어, 요청 처리 과정에서 DB 접근, 외부 API 호출, 로깅, 데이터 처리 등 여러 서비스가 필요할 수 있다.
    수동으로 이 모든 의존성을 초기화하고 관리하는 과정에서는 실수로 의존성을 잘못 설정하거나, 필요한 의존성을 빠뜨리는 일이 발생하기 쉽다.

- **Wire 사용 이점**

    `Wire`는 이러한 의존성을 자동으로 생성하고 연결해 준다.
    개발자는 각 컴포넌트의 생성자만 정의해주면, Wire가 이들을 분석하여 필요한 의존성을 자동으로 주입하는 코드를 생성한다
    이는 인간의 실수를 줄여주며, 의존성 구성 과정에서의 오류 가능성을 감소시킨다.

#### 2. **컴파일 타임 검증**

- **문제 상황**

    런타임 의존성 주입을 사용할 경우, 의존성 문제(예: 필요한 의존성 누락, 잘못된 타입 주입 등)는 애플리케이션이 실제로 실행될 때까지 발견되지 않을 수 있다.
    이는 개발 초기 단계에서 문제를 감지하고 해결하는 데 시간이 더 걸릴 수 있음을 의미한다.

- **Wire 사용 이점**

    `Wire`는 컴파일 타임에 의존성 주입 코드를 생성하므로, 의존성 관련 문제들을 컴파일 시점에 발견할 수 있다.
    즉, 코드가 컴파일되지 않는다면 의존성 문제가 있다는 신호이므로, 개발자는 코드를 배포하기 전에 이러한 문제들을 해결할 수 있다.

### 실질적인 예

```go
// 복잡한 의존성을 가진 서비스 예시

type ServiceA struct {}
type ServiceB struct {}
type ServiceC struct {
    A *ServiceA
    B *ServiceB
}
type ServiceD struct {
    C *ServiceC
}

func NewServiceA() *ServiceA { return &ServiceA{} }
func NewServiceB() *ServiceB { return &ServiceB{} }
func NewServiceC(a *ServiceA, b *ServiceB) *ServiceC { return &ServiceC{A: a, B: b} }
func NewServiceD(c *ServiceC) *ServiceD { return &ServiceD{C: c} }
```

```go
//+build wireinject

// Wire를 통한 자동 의존성 주입 설정

import (
    "github.com/google/wire"
)

func InitializeServiceD() *ServiceD {
    wire.Build(NewServiceD, NewServiceC, NewServiceA, NewServiceB)
    return &ServiceD{}
}
```

위의 예에서 `ServiceD`는 `ServiceC`, `ServiceA`, `ServiceB`에 의존하고 있다.
수동으로 이 모든 의존성을 관리하는 것은 오류를 일으킬 수 있다.
하지만 `Wire`를 사용하면 컴파일 시점에 이 의존성들이 모두 충족되는지 검증된다.
따라서 런타임에 발생할 수 있는 의존성 관련 문제를 예방할 수 있으므로, 의존성을 안전하게 자동으로 처리할 수 있다.

아래 `wire` 명령을 사용하여 의존성 주입 코드를 자동 생성

```go
//+build wireinject

package main

import (
    "github.com/google/wire"
)

func InitializeNotification() *Notification {
    wire.Build(NewNotification, NewEmailService)
    return &Notification{}
}

type EmailService struct{}

func (e *EmailService) Send(message string) error {
    fmt.Println("Sending email:", message)
    return nil
}

func NewEmailService() *EmailService {
    return &EmailService{}
}

type Notification struct {
    MessageService
}

func NewNotification(service MessageService) *Notification {
    return &Notification{MessageService: service}
}

func main() {
    notification := InitializeNotification()
    notification.Send("Hello Wire!")
}
```

## vs Uber's Dig

**Uber's Dig**는 Uber에서 개발한 유연한 리플렉션 기반의 DI 라이브러리입니다. Dig는 런타임에 의존성을 해결하며, 강력한 API를 제공하여 다양한 사용 시나리오를 지원합니다.

### 사용 방법

```bash
go get go.uber.org/dig
```

```go
package main

import (
    "go.uber.org/dig"
    "fmt"
)

type Message string
type Greeter struct {
    Message Message
}
type Event struct {
    Greeter Greeter
}

func NewMessage() Message {
    return Message("Hi there!")
}

func NewGreeter(m Message) Greeter {
    return Greeter{Message: m}
}

func NewEvent(g Greeter) Event {
    return Event{Greeter: g}
}

func main() {
    c := dig.New()
    c.Provide(NewMessage)
    c.Provide(NewGreeter)
    c.Provide(NewEvent)

    if err := c.Invoke(func(e Event) {
        fmt.Println(e.Greeter.Message)
    }); err != nil {
        panic(err)
    }
}
```

### 차이점

- **Google Wire**

    컴파일 타임 DI. Wire는 의존성 주입 코드를 자동으로 생성하여 런타임 오버헤드가 없다.
    설정이 복잡할 수 있지만, 성능과 안정성 측면에서 이점이 있다.

- **Uber's Dig**

    런타임 DI. Dig는 런타임에 의존성을 해결하므로 더 유연하지만, 성능 오버헤드가 있을 수 있다.
    반면, API가 직관적이고 사용하기 쉽다.
