# Go fiber middleware

- [Go fiber middleware](#go-fiber-middleware)
    - [미들웨어에서 인증 후 비즈니스 로직에서 인증 정보 재사용](#미들웨어에서-인증-후-비즈니스-로직에서-인증-정보-재사용)
        - [Locals를 사용한 방법](#locals를-사용한-방법)
            - [1. 사용자 정보를 위한 전용 getter 함수 생성](#1-사용자-정보를-위한-전용-getter-함수-생성)
            - [2. 커스텀 컨텍스트 구조체 사용](#2-커스텀-컨텍스트-구조체-사용)

## 미들웨어에서 인증 후 비즈니스 로직에서 인증 정보 재사용

Go Fiber에서 미들웨어에서 사용자 인증 후 해당 사용자 정보를 다음 요청 처리 로직에서 재사용할 수 있게 하는 몇 가지 방법이 있습니다.
가장 일반적이고 권장되는 방법은 Fiber의 컨텍스트(context)를 사용하는 것입니다. 이 방법을 상세히 설명하고 예제 코드를 제공하겠습니다.

### Locals를 사용한 방법

Fiber의 `c.Locals()` 메서드를 사용하여 요청 범위의 로컬 저장소에 데이터를 저장하고 검색할 수 있습니다.

```go
import (
    "github.com/gofiber/fiber/v2"
    "errors"
)

// User 구조체 정의
type User struct {
    ID   string
    Name string
    // 기타 필요한 필드
}

// AuthMiddleware 함수
func AuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 토큰 추출 (예: Authorization 헤더에서)
        token := c.Get("Authorization")

        // 토큰 검증 및 사용자 정보 조회 (예시 함수)
        user, err := validateTokenAndGetUser(token)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid or expired token",
            })
        }

        // 사용자 정보를 컨텍스트에 저장
        c.Locals("user", user)

        return c.Next()
    }
}

// 실제 요청을 처리하는 핸들러
func ProtectedHandler(c *fiber.Ctx) error {
    // 컨텍스트에서 사용자 정보 추출
    user, ok := c.Locals("user").(*User)
    if !ok {
        return errors.New("user not found in context")
    }

    return c.JSON(fiber.Map{
        "message": "Hello, " + user.Name,
        "userID":  user.ID,
    })
}

// 라우터 설정
func SetupRoutes(app *fiber.App) {
    app.Use(AuthMiddleware())
    app.Get("/protected", ProtectedHandler)
}

// validateTokenAndGetUser 함수 (예시)
func validateTokenAndGetUser(token string) (*User, error) {
    // 실제 구현에서는 토큰을 검증하고 데이터베이스에서 사용자 정보를 조회합니다.
    // 이 예제에서는 간단히 더미 데이터를 반환합니다.
    if token == "valid_token" {
        return &User{ID: "123", Name: "John Doe"}, nil
    }
    return nil, errors.New("invalid token")
}
```

이 접근 방식의 장점:

1. 간단하고 직관적입니다.
2. Fiber의 내장 기능을 사용하므로 추가 라이브러리가 필요 없습니다.
3. 요청 범위 내에서 데이터를 안전하게 공유할 수 있습니다.

주의사항:

1. `c.Locals()`에 저장된 데이터는 인터페이스 타입이므로, 사용할 때 타입 단언(type assertion)이 필요합니다.
2. 키 이름("user" 등)을 문자열로 사용하므로 오타의 위험이 있습니다. 상수를 사용하여 이 문제를 완화할 수 있습니다.

더 견고한 방식으로 구현하려면 다음과 같은 방법을 고려할 수 있습니다:

#### 1. 사용자 정보를 위한 전용 getter 함수 생성

```go
func GetUser(c *fiber.Ctx) (*User, error) {
    user, ok := c.Locals("user").(*User)
    if !ok {
        return nil, errors.New("user not found in context")
    }
    return user, nil
}

// 사용 예
func ProtectedHandler(c *fiber.Ctx) error {
    user, err := GetUser(c)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    // user 사용
}
```

#### 2. 커스텀 컨텍스트 구조체 사용

이 방식은 타입 안정성을 제공하고 IDE의 자동 완성 기능을 더 잘 활용할 수 있게 해줍니다.
그러나 구현이 조금 더 복잡해지는 단점이 있습니다.

```go
type CustomContext struct {
    *fiber.Ctx
    User *User
}

func NewCustomContext(c *fiber.Ctx) *CustomContext {
    return &CustomContext{Ctx: c}
}

func (c *CustomContext) SetUser(user *User) {
    c.User = user
    c.Locals("user", user)
}

func AuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        cc := NewCustomContext(c)
        // 인증 로직...
        user, err := validateTokenAndGetUser(token)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid or expired token",
            })
        }
        cc.SetUser(user)
        return cc.Next()
    }
}

func ProtectedHandler(c *fiber.Ctx) error {
    cc := c.(*CustomContext)
    if cc.User == nil {
        return errors.New("user not found in context")
    }
    // cc.User 사용
}

// 라우터 설정 시 미들웨어 추가
app.Use(func(c *fiber.Ctx) error {
    return NewCustomContext(c).Next()
})
```
