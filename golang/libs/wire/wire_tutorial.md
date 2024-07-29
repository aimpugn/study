# [Wire tutorial](https://github.com/google/wire/blob/main/_tutorial/README.md)

## `Wire` 없이 의존성 주입 구현

greeter가 특정 메시지로 게스트에게 인사하는 이벤트를 시뮬레이션하는 작은 프로그램을 만들어 보자.

```go
// a message for a greeter
type Message string

// a greeter who conveys that message
type Greeter struct {
    Message Message // <- adding a Message field
}

// an event that starts with the greeter greeting guests
type Event struct {
    Greeter Greeter // <- adding a Greeter field
}

func (e Event) Start() {
    msg := e.Greeter.Greet()
    fmt.Println(msg)
}

// simple initializer 
func NewMessage() Message {
    return Message("Hi there!")
}

func NewGreeter(m Message) Greeter {
    return Greeter{Message: m}
}

func NewEvent(g Greeter) Event {
    return Event{Greeter: g}
}
```

이제 애플리케이션의 모든 구성 요소가 준비되었다.
`Wire`를 사용하지 않고 모든 구성 요소를 초기화하는 데 무엇이 필요한지 살펴보자.
`main` 함수는 다음과 같다.

```go
func main() {
    message := NewMessage() // create a message
    greeter := NewGreeter(message) // create a greeter with that message
    event := NewEvent(greeter) // create an event with that greeter
    // ready to start our event

    event.Start()
}
```

## `Wire` 사용해서 의존성 주입 구현

의존성 주입의 한 가지 단점은 초기화 단계가 너무 많다는 것이다.
Wire를 사용하여 컴포넌트를 초기화하는 프로세스를 더 원활하게 만드는 방법을 살펴보자.

```go
func main() {
    e := InitializeEvent()

    e.Start()
}
```

사용하려는 initializer를 전달하여 `wire.Build` 한 번만 호출하면,
각 컴포넌트를 차례로 초기화하여 다음 컴포넌트로 전달하는 번거로움을 겪지 않아도 된다.

```go
// wire.go
func InitializeEvent() Event {
    //                 ^^^^^ 컴파일러를 만족시키기 위해 Event zero-value 추가
    wire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}
}
```

`Wire`에서 initializers는 특정 유형을 제공하는 함수인 "providers"로 알려져 있다.
`Event`에 값을 추가하더라도 `Wire`는 이를 무시한다는 점에 주의하자.
사실 인젝터의 목적은 `Event`를 구성하는 데 사용할 provider에 대한 정보를 제공하는 것이므로, 파일 상단에 빌드 제약 조건을 사용하여 최종 바이너리에서 제외한다.

```go
//+build wireinject
```

`Wire`에서 전문 용어로, `InitializeEvent`는 "injector"다.
injector를 완성했으므로, `wire` CLI를 사용할 준비가 됐다.
