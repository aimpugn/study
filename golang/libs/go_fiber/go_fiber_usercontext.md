# UserContext

## [UserContext](https://docs.gofiber.io/api/ctx#usercontext)?

- Go Fiber의 `UserContext` 메서드는 특정 HTTP 요청의 컨텍스트와 관련된 사용자 정의 데이터를 관리하는 데 사용된다
- 이 메서드가 반환하는 `context.Context` 객체는 하나의 HTTP 요청 동안에만 유효하고, 이는 Go의 HTTP 서버 처리 모델과 Fiber의 요청 처리 방식에 기반하여 구현된다

### 요청 내 유효성 보장 방식

1. HTTP 요청과 `fiber.Ctx`의 연결
    - Go Fiber는 각 HTTP 요청을 처리할 때 고유한 `fiber.Ctx` 인스턴스를 생성한다
    - 이 컨텍스트는 요청의 수명 동안 유효하며, 요청이 완료되면 컨텍스트도 함께 사라진다
2. `UserContext`의 저장과 접근:
    - `UserContext` 메서드는 `fiber.Ctx` 내에 사용자 정의 컨텍스트를
        - 저장
        - 검색
    - 이 메서드는 `Fasthttp`의 `UserValue` 메커니즘을 사용하여 `fiber.Ctx`에 사용자 정의 컨텍스트를 저장한다
        - `UserValue`는 특정 요청과 연결된 키-값 쌍을 저장하는 메커니즘
    - 요청이 시작될 때 `UserContext`가 설정되지 않았다면, `context.Background()`를 사용하여 새로운 기본 컨텍스트를 생성하고 이를 `fiber.Ctx`에 연결한다
3. 요청 수명 주기:
   - 요청이 들어오면, Fiber는 새로운 `fiber.Ctx` 인스턴스를 생성하고 이를 처리 로직에 전달한다
   - `UserContext`는 이 `fiber.Ctx` 인스턴스 내에 저장되므로, 저장된 사용자 정의 컨텍스트는 해당 HTTP 요청이 처리되는 동안에만 유효하다
   - 요청 처리가 완료되면, `fiber.Ctx`와 그 내부에 저장된 모든 데이터, 포함하여 `UserContext`,는 더 이상 유효하지 않게 된다
