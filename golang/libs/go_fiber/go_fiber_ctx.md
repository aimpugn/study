# Go Fiber CTX

- [Go Fiber CTX](#go-fiber-ctx)
    - [Fiber에서 `ctx.SetUserContext` 작동 방식](#fiber에서-ctxsetusercontext-작동-방식)
    - [Fasthttp와 Fiber의 메모리 재사용](#fasthttp와-fiber의-메모리-재사용)
    - [Fiber 컨텍스트와 UserContext의 처리](#fiber-컨텍스트와-usercontext의-처리)
    - [다른 프레임워크의 유사 메커니즘](#다른-프레임워크의-유사-메커니즘)
    - [`fiber.Ctx` vs `context.Context` (그리고 `UserContext`)](#fiberctx-vs-contextcontext-그리고-usercontext)
        - [`fiber.Ctx`](#fiberctx)
            - [`UserContext`와 `SetUserContext`](#usercontext와-setusercontext)
            - [`UserContext` 라이프사이클](#usercontext-라이프사이클)
            - [`ctx.SetUserContext` 사용 방법](#ctxsetusercontext-사용-방법)
        - [`context.Context`](#contextcontext)
    - [`ctx.Next()`](#ctxnext)
    - [`Locals`와 `SetUserContext`](#locals와-setusercontext)
        - [`Locals` 메서드](#locals-메서드)
            - [작동 원리](#작동-원리)
            - [값이 사라지는 시점](#값이-사라지는-시점)
        - [`SetUserContext` 메서드](#setusercontext-메서드)
        - [어떤 방법을 사용해야 할까?](#어떤-방법을-사용해야-할까)
    - [Examples](#examples)
        - [인증 데이터를 컨텍스트에 저장하고 전달하기](#인증-데이터를-컨텍스트에-저장하고-전달하기)

## Fiber에서 `ctx.SetUserContext` 작동 방식

Go Fiber의 `ctx`는 fasthttp 요청 컨텍스트를 추상화하여 HTTP 요청 및 응답을 다루기 쉽게 설계되었습니다. `UserContext` 기능은 이 추상화의 일부로, 요청의 생명 주기를 통해 데이터를 전달하는 방법을 제공합니다.

Go Fiber에서 `ctx.SetUserContext(context.WithValue(ctx.UserContext(), key, value))`를 호출할 때:

1. `ctx.UserContext()`는 이 요청과 관련된 현재 사용자 특정 컨텍스트를 검색합니다.
2. `context.WithValue(ctx.UserContext(), key, value)`는 사용자 컨텍스트에서 파생된 새 컨텍스트를 생성하며, 이 컨텍스트에는 키-값 쌍이 저장됩니다.
3. `ctx.SetUserContext(...)`는 요청의 사용자 컨텍스트를 이 새 컨텍스트로 업데이트합니다.

이 과정을 통해 인증 세부 정보, 사용자 특정 설정 또는 다른 요청 범위 데이터와 같이 요청의 생명 주기와 관련된 값을 저장하고 검색할 수 있습니다.

## Fasthttp와 Fiber의 메모리 재사용

Fasthttp는 공격적인 메모리 재사용을 통해 높은 성능을 달성하는 것으로 알려져 있습니다. 여러 연결에 대해 재사용되는 요청 및 응답 구조의 풀을 유지하여 가비지 컬렉션 압박을 줄이고 성능을 향상시킵니다.

Fiber를 사용할 때, fasthttp 위에 구축된:

- Fasthttp 요청 및 응답 객체가 재사용됩니다. 그러나 Fiber는 이 세부 정보를 사용자로부터 추상화하여 더 사용자 친화적인 API를 제공합니다.
- Fiber의 `UserContext`는 이 재사용을 염두에 두고 설계되었습니다. 새 사용자 컨텍스트를 설정하면, 이 컨텍스트가 현재 요청에 특정하고 동일한 기본 fasthttp 구조를 재사용하는 미래 요청과 실수로 공유되지 않도록 합니다.

## Fiber 컨텍스트와 UserContext의 처리

- **초기화 및 정리**: 요청이 처리될 때마다 Fiber는 `UserContext`가 적절하게 초기화되고 요청이 처리된 후 정리되도록 합니다. 이는 요청 간 데이터 유출을 방지합니다.
- **동시성**: Fiber와 fasthttp는 동시 환경에서 높은 성능을 발휘하도록 설계되었습니다. 메모리 및 컨텍스트의 재사용은 동시 요청 간 격리를 보장하는 방식으로 관리됩니다.

## 다른 프레임워크의 유사 메커니즘

다른 언어의 다른 웹 프레임워크는 요청 범위 저장소에 대한 유사한 메커니즘을 가질 수 있습니다. 예를 들어, Express.js(Node.js)의 요청 컨텍스트나 ASP.NET Core의 미들웨어 컨텍스트 등이 있습니다. 핵심 아이디어는 전체 처리 파이프라인을 통해 요청별 데이터를 안전하고 효율적으로 전달할 수 있는 방법을 제공하는 것입니다.

요약하자면, 특정 구현 세부 사항과 메모리 처리 최적화는 fasthttp와 Fiber에 특유의 것이지만, 요청 범위 컨텍스트 저장소의 개념은 웹 개발에서 흔히 볼 수 있습니다. Fiber의 `UserContext` 사용은 기본적인 fasthttp 라이브러리의 성능 특성에 최적화된 Go 중심적 방법을 제공합니다.

## `fiber.Ctx` vs `context.Context` (그리고 `UserContext`)

- **context.Context**

    Go의 표준 라이브러리에 포함된 인터페이스로, 요청의 생명주기 관리, 요청 취소 신호 전달, 요청 범위의 값 전달 등을 위해 사용된다.
    동시성을 다루는 프로그램에서 중요한 역할을 한다.

- **fiber.Ctx**

    Fiber 프레임워크에서 HTTP 요청과 응답을 다루기 위해 제공하는 구조체다.
    요청의 헤더, 바디, 쿠키 등을 접근하고, 응답을 구성하는 데 필요한 메서드들을 포함하고 있다.

- **UserContext 함수**

    `fiber.Ctx` 내에서 사용자가 설정한 `context.Context` 인스턴스를 반환한다.
    이를 통해 Fiber의 요청 처리 과정에서 표준 `context.Context`의 기능을 사용할 수 있다.

### `fiber.Ctx`

`fiber.Ctx`는 Fiber 프레임워크에서 HTTP 요청과 응답의 컨텍스트를 관리하는 주요 구조체다.
이 구조체는 요청의 메타데이터, 바디, 헤더 등을 포함하고 있으며, 응답을 구성하는 데 필요한 메서드들을 제공한다.
또한, 라우팅 파라미터, 쿼리 파라미터, 쿠키 등에 접근할 수 있는 메서드들도 포함하고 있다.
`fiber.Ctx`는 Fiber의 핵심적인 부분으로, 웹 애플리케이션에서 요청과 응답을 처리하는 데 필수적이다.

#### `UserContext`와 `SetUserContext`

`UserContext`는 Fiber에서 제공하는 `fiber.Ctx`의 메서드 중 하나로, 사용자가 설정한 `context.Context` 인스턴스를 반환한다.
만약 사용자가 `SetUserContext` 메서드를 통해 `context.Context`를 설정하지 않았다면, `UserContext`는 비어있지 않은 기본 `context.Context`를 반환한다. 이는 Fiber의 요청 처리 파이프라인에서 Go의 표준 `context.Context`를 사용하고자 할 때 유용하다.

이를 통해 Fiber 핸들러 내부에서 표준 `context.Context`를 사용할 수 있게 되며, 다른 라이브러리나 프레임워크와의 호환성을 높여준다.

`SetUserContext` 메서드는 사용자가 지정한 `context.Context` 인스턴스를 `fiber.Ctx`에 설정한다.
이를 통해 사용자는 표준 `context.Context`의 기능을 Fiber의 요청 처리 과정에 통합할 수 있다. 예를 들어:
- 요청 처리 중에 데이터베이스 작업을 취소하기
- 요청 처리에 필요한 메타데이터를 전달하기 등

#### `UserContext` 라이프사이클

Go Fiber에서 `ctx.SetUserContext()`를 사용하여 유저 컨텍스트에 값을 저장하면, 해당 유저 컨텍스트의 라이프사이클은 다음과 같다:

1. 요청(Request) 시작 시점에 유저 컨텍스트가 생성된다.
2. `ctx.SetUserContext()`를 호출하여 유저 컨텍스트에 값을 저장한다.
3. 해당 요청에 대한 처리가 진행되는 동안 유저 컨텍스트는 유지되며, `ctx.UserContext()`를 통해 저장된 값에 접근할 수 있다.
4. 요청 처리가 완료되면, 즉 응답(Response)이 클라이언트에게 전송된 후 유저 컨텍스트는 메모리에서 해제된다.

유저 컨텍스트는 개별 요청(Request)에 한정되어 존재하며, 요청이 완료되면 함께 소멸되고, 다음 요청이 들어오면 새로운 유저 컨텍스트가 생성된다.
각 요청마다 독립적인 유저 컨텍스트를 가지므로, 유저 컨텍스트에 저장된 값은 해당 요청 내에서만 유효하며, 다른 요청 간에 공유되지 않는다.

만약 여러 요청에 걸쳐 유저 정보를 유지해야 한다면, 세션(Session)이나 JWT(JSON Web Token)과 같은 방식을 사용하여 유저 정보를 저장하고 관리해야 한다.

`ctx.SetUserContext`를 호출하면, Fiber는 내부적으로 현재 요청의 컨텍스트(`ctx.UserContext()`)를 새로운 컨텍스트로 대체한다.
이 새로운 컨텍스트는 `context.WithValue`를 통해 생성된 컨텍스트로, 여기에는 사용자 정의 값이 포함될 수 있다.
이렇게 설정된 컨텍스트는 현재 처리 중인 HTTP 요청에 대한 처리가 완료될 때까지 유효하다.

#### `ctx.SetUserContext` 사용 방법

`context` 패키지를 사용하여 요청 처리 과정에서 필요한 데이터를 효율적으로 전달하고 관리하기 위해 사용한다.

```go
// 특정 HTTP 요청을 처리하는 동안 사용자 정보(`&users.User{}`)를 컨텍스트에 저장하여, 
// 요청 처리 과정에서 이 정보를 쉽게 접근하고 사용할 수 있도록 한다. 
// 이는 인증된 사용자의 정보를 요청 처리 과정의 다양한 단계에서 사용해야 할 때 유용하다.
ctx.SetUserContext(
    context.WithValue( // 자식 컨텍스트
        ctx.UserContext(), // 부모 컨텍스트
        userContextKey, 
        &users.User{},
    ),
)
```

1. **요청 스코프 데이터 전달**

    HTTP 요청을 처리하는 과정에서는 사용자 인증 정보, 요청에 대한 메타데이터 등 요청과 관련된 다양한 데이터를 다루게 된다.
    이러한 데이터를 요청 처리 파이프라인의 여러 단계와 함수들 사이에서 전달하고 공유해야 할 필요가 있다.
    `context.Context`를 사용하면 이러한 데이터를 효율적으로 전달하고 관리할 수 있다.

2. **취소 신호 및 타임아웃 전파**

    `context.Context`는 요청 처리 과정에서 취소 신호나 타임아웃 같은 제어 신호를 전파하는 데에도 사용된다.
    예를 들어, 사용자가 요청을 취소하거나 설정된 타임아웃이 도래했을 때, 요청을 처리하는 모든 함수에 전파하여 적절한 조치를 취할 수 있다.

3. **요청 생명주기 관리**

    `context.Context`는 요청의 생명주기와 밀접하게 연결되어 있다.
    요청이 시작될 때 컨텍스트가 생성되고, 요청이 완료되면 컨텍스트에 저장된 모든 데이터와 리소스가 정리된다.
    이를 통해 요청과 관련된 리소스의 생명주기를 효과적으로 관리할 수 있다.

### `context.Context`

`context.Context`는 Go 표준 라이브러리에서 제공하는 인터페이스로, 요청의 실행 시간, 취소 신호, 요청 범위의 값 등을 전달하는 데 사용된다.
이 인터페이스는 Go의 동시성 관리와 관련된 작업에서 중요한 역할을 한다. 예를 들어:
- 데이터베이스 쿼리나 HTTP 요청과 같이 시간이 걸리는 작업을 취소하기
- 요청 사이에 메타데이터를 전달하기 등

## `ctx.Next()`

`gofiber/fiber` 패키지에서 `Ctx.Next()` 메서드는 현재 요청-응답 사이클에서 다음 미들웨어 함수를 실행한다.
이 메서드는 다음 미들웨어 또는 핸들러 함수에서 에러가 발생했을 때 에러 반환한다.

## `Locals`와 `SetUserContext`

```go
// y/go/pkg/mod/github.com/gofiber/fiber/v2@v2.50.0/ctx.go

// Locals makes it possible to pass interface{} values under keys scoped to the request
// and therefore available to all following routes that match the request.
func (c *Ctx) Locals(key interface{}, value ...interface{}) interface{} {
    if len(value) == 0 {
        return c.fasthttp.UserValue(key)
    }
    c.fasthttp.SetUserValue(key, value[0])
    return value[0]
}

// SetUserContext sets a context implementation by user.
func (c *Ctx) SetUserContext(ctx context.Context) {
    c.fasthttp.SetUserValue(userContextKey, ctx)
}
```

Go Fiber 라이브러리의 `Ctx` 객체에 정의된 `Locals` 메서드와 `SetUserContext` 메서드는 모두 컨텍스트 관련 데이터를 저장하고 관리하는 데 사용되지만, 용도와 사용 방법에 차이가 있다.

### `Locals` 메서드

`Locals` 메서드는 Fiber 요청(`Ctx` 객체)의 생명주기 동안 접근 가능한 `interface{}` 타입의 값을 저장하고 검색하는 데 사용된다.
이 메서드는 키-값 쌍으로 데이터를 저장하며, 요청의 처리 과정에서 추가적인 정보를 전달하거나 상태를 공유하는 데 유용하다.

**사용법:**

- 값을 저장할 때: `Locals(key, value)` 형태로 사용하며, 특정 키에 값을 할당합니다.
- 값을 검색할 때: `Locals(key)` 형태로 사용하며, 주어진 키에 해당하는 값을 반환합니다.

**특징:**

- 다양한 타입의 값을 저장할 수 있으며, **요청 범위 내에서만 유효**하다.
- 주로 인증 정보, 요청 관련 메타데이터 등을 저장하는 데 사용된다.

#### 작동 원리

`ctx.Locals` 메서드 (또는 Fasthttp의 `SetUserValue` 함수)는 요청 처리 과정 중에 임의의 값을 저장하고 검색하는 기능을 제공한다.
이러한 값은 요청의 생명주기 동안만 유효하며, 요청이 완료된 후에는 자동으로 제거된다.

- **인메모리 저장**

    `ctx.Locals` 또는 `SetUserValue`를 통해 저장된 값은 메모리 내의 특정 구조(예: 맵 또는 컨텍스트 객체의 필드)에 저장된다.
    이 값은 키-값 쌍으로 관리되며, 요청 처리 도중에 필요한 정보를 임시로 저장하는 데 사용된다

- **요청 생명주기**

    저장된 값은 HTTP 요청을 처리하는 동안에만 접근 가능하며, 요청이 완료되고 응답이 반환된 후에는 이러한 값이 자동으로 제거된다.
    이는 각 요청마다 별도의 컨텍스트(`RequestCtx` 또는 `Ctx`) 인스턴스가 생성되기 때문이다.

#### 값이 사라지는 시점

- **자동 제거**

    Fasthttp (또는 Fiber)는 요청 처리를 완료한 후, 해당 요청의 컨텍스트에 저장된 모든 사용자 정의 값들을 자동으로 제거한다.
    이 과정은 내부적으로 처리되며, 개발자가 명시적으로 값을 제거할 필요는 없다.

- **`io.Closer` 인터페이스**

    Fasthttp는 `SetUserValue`를 통해 저장된 값 중에서 `io.Closer` 인터페이스를 구현한 값에 대해 `Close()` 메서드를 호출한다.
    이는 열린 파일, 데이터베이스 연결 등의 리소스를 안전하게 해제하는 데 사용될 수 있다.

### `SetUserContext` 메서드

`SetUserContext` 메서드는 사용자가 정의한 `context.Context` 인터페이스 구현체를 Fiber 요청(`Ctx` 객체)에 저장하는 데 사용된다.
이 메서드를 통해 저장된 컨텍스트는 주로 Go의 표준 라이브러리나 다른 미들웨어, 서비스 간의 컨텍스트 전달에 사용된다.

**사용법:**

- `SetUserContext(ctx)` 형태로 사용하며, `context.Context` 인터페이스를 구현한 객체를 요청에 저장한다.

**특징:**

- Go 표준 `context.Context`의 기능을 활용할 수 있으며, 타임아웃, 취소 신호 등을 관리할 수 있다.
- 주로 외부 API 호출, 데이터베이스 작업 등과 같이, 표준 컨텍스트 기능이 필요한 작업에서 사용된다.

### 어떤 방법을 사용해야 할까?

두 메서드는 사용 목적과 상황에 따라 달라집니다.

- **`Locals`의 사용**

    요청 범위 내에서의 간단한 데이터 공유나 상태 관리가 필요한 경우 `Locals`를 사용하는 것이 좋다.
    이는 특정 요청에 한정된 정보를 저장하고 공유하는 용도로 적합하다.
  
- **`SetUserContext`의 사용**

    요청 처리 과정에서 표준 `context.Context`의 기능(예: 취소 신호, 타임아웃)을 활용해야 하는 경우 `SetUserContext`를 사용하는 것이 좋다. 이는 더 복잡한 상황에서의 컨텍스트 관리에 적합하다.

## Examples

### 인증 데이터를 컨텍스트에 저장하고 전달하기

Go Fiber는 Go 표준 라이브러리의 `context.Context`와는 다른 방식과 목적으로 컨텍스트를 관리한다.
- Fiber의 컨텍스트: 요청 및 응답과 관련된 메서드와 속성을 포함하는 `*fiber.Ctx` 객체
- Go 표준 라이브러리의 `context.Context`: 주로 요청의 생명주기, 취소 신호, 요청 스코프의 데이터 전달 등을 관리하는 데 사용

Go Fiber에서 인증 결과와 같은 데이터를 요청 컨텍스트에 저장하고, 각 라우트 핸들러로 전달하는 것은 일반적인 패턴이다.
그러나 Go 표준 라이브러리의 `context.Context`를 직접 전달하는 것은 Fiber의 설계와 사용 방법에 따라 다를 수 있다.

```go
package main

import (
    "github.com/gofiber/fiber/v2"
)
// userContextKey 타입은 사용자 정의 컨텍스트에서 사용되는 키의 타입을 나타냅니다.
// 단순 문자열 사용시 다른 코드에서 키가 겹칠 수 있으므로 타입을 정의하여 키의 고유성을 보장합니다.
// 외부로 노출할 필요 없으므로 export 하지 않습니다.
type userContextKey string

// userKey 변수는 userContextKey 타입을 사용하여 사용자 정의 키를 정의합니다.
// 이 키를 사용하여 인증된 사용자 정보를 컨텍스트에 저장하거나 검색할 수 있습니다.
// 외부로 노출할 필요 없으므로 export 하지 않습니다.
const userKey userContextKey = "some.user.key"

// 인증 미들웨어: 모든 요청에 대한 인증을 수행하고 인증된 사용자 정보를 컨텍스트에 저장합니다.
func AuthMiddleware(c *fiber.Ctx) error {
    // 인증 로직 수행 (여기서는 예시로 간단한 구현만을 보여줍니다)
    user := User{ID: "123", Name: "John Doe"}
    
    // 인증된 사용자 정보를 컨텍스트에 저장합니다.
    c.Locals(userKey, &user)
    
    // 요청 동안 유효한 사용자 정의 컨텍스트에 API 요청한 사용자 정보를 저장하여 이 요청의 생애주기 동안 사용합니다.
    // c.SetUserContext(
    //    // Fiber의 컨텍스트 관리 방식과 표준 라이브러리의 `context.Context` 사용 방식 간에는 차이가 있으므로, 
    //    // 각각의 컨텍스트 타입이 제공하는 기능과 메서드 사용에 주의해야 합니다.
    //   context.WithValue(c.UserContext(), userKey, &user)
    // )
    // 
    
    // 다음 핸들러로 진행합니다.
    if err := c.Next(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, errMsgInternalServerError)
    }
    return nil
}

// 사용자 정보를 출력하는 라우트 핸들러
func GetUser(c *fiber.Ctx) error {
    // 컨텍스트에서 사용자 정보를 검색합니다.
    user, ok := c.Locals(userKey).(*User)
    if !ok {
        return c.SendStatus(fiber.StatusUnauthorized)
    }
    
    // 사용자 정보를 응답으로 반환합니다.
    return c.JSON(user)
}

func main() {
    app := fiber.New()
    
    // 모든 요청에 대해 AuthMiddleware를 적용합니다.
    app.Use(AuthMiddleware)
    
    // '/user' 경로에 대한 라우트를 등록합니다. GetUser 핸들러는 인증된 사용자 정보에 접근할 수 있습니다.
    app.Get("/user", GetUser)
    
    app.Listen(":3000")
}

// User는 사용자 정보를 나타내는 예시 구조체입니다.
type User struct {
    ID   string
    Name string
}
```

이 예시에서 `AuthMiddleware`는 인증을 수행하고, 인증된 사용자 정보를 Fiber 컨텍스트의 `Locals` 메서드를 사용하여 저장한다.
각 라우트 핸들러(예: `GetUser`)는 `Locals` 메서드를 사용하여 이 정보에 접근할 수 있다.

Go Fiber의 `*fiber.Ctx` 객체를 사용하여 데이터를 전달하는 방식은 성능을 최우선으로 고려한 Fiber의 설계와 잘 어울립니다:
- 모든 요청과 응답 데이터를 `*fiber.Ctx` 객체 하나로 관리합니다.
    - 따라서 `*fiber.Ctx` 객체를 사용하면 미들웨어 간의 데이터 전달이 매우 빠르게 이루어집니다.
    - 요청과 응답을 처리하는 데 필요한 모든 기능을 통합하여 제공하므로, 별도의 컨텍스트 객체를 사용할 필요가 없습니다.
- 미들웨어 체인을 순차적으로 호출하며, 각 미들웨어에서 `*fiber.Ctx` 객체를 통해 데이터를 공유하므로 성능이 향상됩니다:
    - 함수 호출 스택의 깊이를 최소화
    - 데이터를 복사하지 않고 참조만 전달
- `*fiber.Ctx` 객체는 요청 처리 중 반복적으로 재사용되어, 불필요한 메모리 할당과 해제를 줄입니다. 이는 성능에 큰 영향을 미치는 가비지 컬렉션의 부담을 줄입니다.

그러나 Go 표준 라이브러리의 `context.Context`와 같은 방식으로 데이터를 전달하려면, 이를 `*fiber.Ctx` 객체에 통합하거나 랩핑하는 추가적인 작업이 필요할 수 있습니다.
