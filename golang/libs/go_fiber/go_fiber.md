# Go Fiber

## Go Fiber?

> Fiber is a Go web framework built on top of `Fasthttp`, the fastest HTTP engine for Go. It's designed to ease things up for fast development with `zero memory allocation` and performance in mind.

- Fiber는 Fasthttp를 기반으로 하는 Go 언어의 웹 프레임워크.
- Fiber의 핵심 설계 철학:
    - 빠른 개발
    - 제로 메모리 할당
    - 성능 중시
- 이러한 설계는 특히 고성능을 위해 최적화되어 있으며, Fiber 컨텍스트(*fiber.Ctx)에서 반환되는 값들은 기본적으로 변경 불가능한(immutable) 상태가 아니며, 요청 간에 재사용된다.

## [`Zero Allocation`](https://docs.gofiber.io/#zero-allocation)

> 자세한 사항은 [#426](https://github.com/gofiber/fiber/issues/426) 이슈와  [#185](https://github.com/gofiber/fiber/issues/185) 이슈 참고

- 고성능에 최적화 되어서, `fiber.Ctx`에서 반환되는 값들은 기본적으로 변경되지 않고(immutable), 이러한 객체들을 여러 요청에서 재사용한다.  
    - 메모리 사용량을 줄인다
    - 가비지 컬렉터의 부하를 감소시키는 데 도움이 된다
- 기본 설정에서는 `fiber.Ctx`의 값들이 **완전히 불변 상태로 반환되는 것은 아니다**. 반환된 값들은 변경할 수 있으나, 이러한 변경이 다른 요청에 영향을 미칠 수 있습니다.
- 경험상, **반드시** 컨텍스트의 값들을 핸들러 내에서만 사용해야 하고, **반드시** 참조를 유지하지 않아야 한다
- 핸들러에서 되돌아 오자마자, [컨텍스트에서 얻은 모든 값은 향후 다른 요청에서 재사용](https://stackoverflow.com/a/74513794)된다.

    ```go
    func handler(c *fiber.Ctx) error {
        // 변수는 이 핸들러 안에서만 유효하다
        result := c.Params("foo") 

        // ...
    }
    ```

- 이러한 컨텍스트에서 얻은 값들을 핸들러 외부에서 유지해야 하는 경우 내장된 복사 기능을 사용하여 기본 버퍼의 복사본을 만든다

    ```go
    func handler(c *fiber.Ctx) error {
        // 변수는 이 핸들러 안에서만 유효하다
        result := c.Params("foo")

        // 본사본을 만든다
        buffer := make([]byte, len(result))
        copy(buffer, result)
        resultCopy := string(buffer) 
        // 이제 변수는 영원히(forever) 유효하다

        // ...
    }
    ```

- `Immutable` 설정을 사용하여 컨테스트에서 리턴된 모든 값들을 불변(immutable)로 만들어서 어디서나 값을 유지할 수 있게 해준다. 물론 이는 성능 저하라는 대가를 치러야 한다

    ```go
    app := fiber.New(fiber.Config{
        Immutable: true,
    })
    ```

### 완전/불완전 불변

- "완전한 불변"과 "불완전한 불변" 사이의 차이는 **객체가 얼마나 엄격하게 변경으로부터 보호되는지에 관한 것**

1. 완전한 불변성 (Complete Immutability):
   - 완전한 불변성을 가진 객체는 생성 후 그 상태가 전혀 변경될 수 없다
   - 즉, 객체의 어떤 속성도 변경할 수 없으며, 이러한 객체는 여러 부분에서 안전하게 공유될 수 있다
   - 예를 들어, 문자열(`String`)은 한 번 생성되면 그 내용을 변경할 수 없으므로, Go에서 불변 객체의 전형적인 예가 된다.

2. 불완전한 불변성 (Incomplete Immutability):
   - 불완전한 불변성을 가진 객체는 특정 조건 하에서만 불변성을 유지한다
   - 즉, 객체의 일부 속성이 변경될 수 있거나, 특정 메서드를 통해 상태 변경이 가능하다
   - 이러한 객체는 완전히 불변하지 않기 때문에, 공유 시 주의가 필요하며, 동시성 문제가 발생할 수 있다

Go Fiber에서의 상황을 보면:

- 기본적인 `fiber.Ctx` 사용:
    - Fiber는 성능을 위해 객체를 재사용하고, 가능한 한 불변성을 유지하려고 한다
    - 그러나, 이 경우의 불변성은 완전하지 않다. `fiber.Ctx`에서 반환된 값들은 기술적으로 변경 가능할 수 있으며, 이러한 변경이 동시성 문제나 데이터 무결성 문제를 일으킬 수 있다.
- `Immutable` 설정 활성화
    - 이 설정을 사용하면 `fiber.Ctx`에서 반환되는 모든 값들이 완전히 불변 상태가 된다.
    - 즉, 반환된 값들은 변경할 수 없게 되며, 어떠한 상황에서도 데이터의 무결성이 보장된다

## `Use`(미들웨어)와 순서

## `router.go`와 데이터독 에러
