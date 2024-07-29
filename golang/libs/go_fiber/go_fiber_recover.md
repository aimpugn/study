# Go Fiber Recover

## [Recover](https://docs.gofiber.io/api/middleware/recover/)?

- 스택 체인 어디서든 패닉을 복구하고 중앙 집중적인 에러 핸들러에 대한 제어를 처리하는 Fiber용 복구 미들웨어
- 각 핸들러나 미들웨어에서 개별적으로 `panic`을 처리할 필요 없이, `Recover` 미들웨어가 모든 `panic` 상황을 잡아내고 처리
- 주로 패닉 상황을 처리하고, 애플리케이션이 갑작스럽게 중단되는 것을 방지하는 데 사용

## default config

```go
var ConfigDefault = Config{
    Next:              nil,
    EnableStackTrace:  false,
    StackTraceHandler: defaultStackTraceHandler,
}
```

## 예시

```go
// Initialize default config
app.Use(recover.New())

// This panic will be caught by the middleware
app.Get("/", func(c *fiber.Ctx) error {
    panic("I'm an error")
})
```

## 어떻게 복구가 가능한가?

- Go에서 "복구(recover)"라는 개념은 주로 `panic` 발생 시 사용된다
- `panic`은 예상치 못한 오류가 발생했을 때 프로그램의 정상적인 실행 흐름을 중단시키는 메커니즘
- 이때 `recover` 함수를 사용하여 `panic` 상태에서 복구할 수 있다

### `recover` 함수의 원리

1. 패닉 감지:
    - `recover`는 `defer` 문과 함께 사용된다
    - `defer`로 지정된 함수는 함수가 리턴하기 직전에 실행된다.
    - 이 함수 내에서 `recover()`를 호출하면, 현재 발생한 패닉을 감지하고 처리할 수 있다
2. 복구 실행
    - `recover()` 함수가 패닉 상태를 감지하면, 그 상태에서 복구를 시도한다.
    - 즉, 패닉으로 인해 중단된 함수의 실행을 멈추고, `panic`이 발생한 함수가 리턴된다
    - 이로써 프로그램이 완전히 중단되지 않고 계속 실행될 수 있다
3. 프로그램 실행 계속
    - `recover`에 의해 복구가 이루어지면, 패닉이 발생했던 함수의 나머지 부분은 실행되지 않고, 프로그램의 실행 흐름은 `panic`이 발생한 함수를 호출한 곳으로 돌아간다

### `recover`를 사용하는 예시

```go
func mayPanic() {
    panic("a problem")
}

func main() {
    defer func() {
        if r := recover(); r != nil { //  `recover`를 호출하여 프로그램이 완전히 중단되지 않고 복구된다
            fmt.Println("Recovered. Error:\n", r)
        }
    }()
    mayPanic() // -> `panic` 실행
    // 이 시점에서 `defer`로 지정된 함수 실행되고 `recover`를 호출되고, 
    // 제어는 `main` 함수로 돌아간다
    fmt.Println("After mayPanic()") // 이 부분은 실행되지 않음
}
```
