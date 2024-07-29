# fiber routes

## route마다 미들웨어 생성하여 설정

```go
package routes

import (
    "github.com/create-go-app/fiber-go-template/app/controllers"
    "github.com/create-go-app/fiber-go-template/pkg/middleware"
    "github.com/gofiber/fiber/v2"
)

// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App) {
    // Create routes group.
    route := a.Group("/api/v1")

    // Routes for POST method:
    route.Post("/book", middleware.JWTProtected(), controllers.CreateBook)           // create a new book
    route.Post("/user/sign/out", middleware.JWTProtected(), controllers.UserSignOut) // de-authorization user
    route.Post("/token/renew", middleware.JWTProtected(), controllers.RenewTokens)   // renew Access & Refresh tokens

    // Routes for PUT method:
    route.Put("/book", middleware.JWTProtected(), controllers.UpdateBook) // update one book by ID

    // Routes for DELETE method:
    route.Delete("/book", middleware.JWTProtected(), controllers.DeleteBook) // delete one book by ID
}
```

이 코드 구조에는 몇 가지 이유가 있을 수 있습니다:

1. 유연성과 세밀한 제어:
   각 라우트마다 middleware.JWTProtected()를 개별적으로 호출함으로써, 필요에 따라 특정 라우트에 대해 미들웨어를 쉽게 추가하거나 제거할 수 있습니다.
   이는 라우트별로 다른 인증 요구사항이 생길 경우 유용합니다.

2. 미들웨어 설정의 독립성:
   각 라우트에 대해 미들웨어를 독립적으로 설정함으로써, 한 라우트의 미들웨어 변경이 다른 라우트에 영향을 미치지 않습니다.

3. 가독성과 명확성:
   각 라우트에 대해 미들웨어를 명시적으로 지정함으로써, 코드를 읽는 사람이 각 라우트의 보안 요구사항을 쉽게 파악할 수 있습니다.

4. 미들웨어의 동적 구성 가능성:
   middleware.JWTProtected()가 호출될 때마다 새로운 미들웨어 인스턴스를 생성한다면, 이는 각 라우트에 대해 미들웨어를 동적으로 구성할 수 있는 기회를 제공합니다.

5. 성능 영향 최소화:
   Go의 함수 호출은 일반적으로 매우 빠르기 때문에, 이렇게 미들웨어를 여러 번 생성하는 것이 성능에 큰 영향을 미치지 않을 수 있습니다.

6. 프레임워크 설계 철학:
   Fiber 프레임워크가 이러한 방식의 미들웨어 적용을 권장하거나, 이 방식이 프레임워크의 내부 동작과 더 잘 맞을 수 있습니다.

하지만 만약 모든 라우트에 대해 동일한 JWT 보호가 필요하다면, 미들웨어를 한 번만 생성하고 재사용하는 것이 더 효율적일 수 있습니다.

예를 들어:

```go
// PrivateRoutes func for describe group of private routes.
func PrivateRoutes(a *fiber.App) {
    // Create routes group.
    route := a.Group("/api/v1")

    // Create a single instance of the JWT middleware
    jwtMiddleware := middleware.JWTProtected()

    // Routes for POST method:
    route.Post("/book", jwtMiddleware, controllers.CreateBook)
    route.Post("/user/sign/out", jwtMiddleware, controllers.UserSignOut)
    route.Post("/token/renew", jwtMiddleware, controllers.RenewTokens)

    // Routes for PUT method:
    route.Put("/book", jwtMiddleware, controllers.UpdateBook)

    // Routes for DELETE method:
    route.Delete("/book", jwtMiddleware, controllers.DeleteBook)
}
```

이 접근 방식은 코드를 더 DRY(Don't Repeat Yourself)하게 만들고, 미들웨어 설정의 중앙화된 지점을 제공합니다. 그러나 각 라우트에 대해 미들웨어를 개별적으로 지정하는 원래의 방식도 유효하며, 특히 미래의 유연성을 위해 선택될 수 있습니다.

최종적인 선택은 프로젝트의 요구사항, 팀의 선호도, 그리고 예상되는 미래의 변경사항에 따라 달라질 수 있습니다.
