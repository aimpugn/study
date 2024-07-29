# Go Fiber Test

- [Go Fiber Test](#go-fiber-test)
    - [App \> Test](#app--test)
    - [recipes \> unit-test](#recipes--unit-test)
    - [핸들러 로직만 테스트](#핸들러-로직만-테스트)
        - [1. 핸들러를 Fiber 라우트에 바인딩](#1-핸들러를-fiber-라우트에-바인딩)
        - [2. 핸들러 로직만 직접 테스트](#2-핸들러-로직만-직접-테스트)
    - [기타](#기타)
        - [참고](#참고)

## [App > Test](https://docs.gofiber.io/api/app/#test)

```go
// Create route with GET method for test:
app.Get("/", func(c *fiber.Ctx) error {
  fmt.Println(c.BaseURL())              // => http://google.com
  fmt.Println(c.Get("X-Custom-Header")) // => hi

  return c.SendString("hello, World!")
})

// http.Request
req := httptest.NewRequest("GET", "http://google.com", nil)
req.Header.Set("X-Custom-Header", "hi")

// http.Response
resp, _ := app.Test(req)

// Do something with results:
if resp.StatusCode == fiber.StatusOK {
  body, _ := io.ReadAll(resp.Body)
  fmt.Println(string(body)) // => Hello, World!
}
```

## [recipes > unit-test](https://github.com/gofiber/recipes/blob/master/unit-test/main_test.go)

```go
func TestIndexRoute(t *testing.T) {
    // Define a structure for specifying input and output
    // data of a single test case. This structure is then used
    // to create a so called test map, which contains all test
    // cases, that should be run for testing this function
    tests := []struct {
        description string

        // Test input
        route string

        // Expected output
        expectedError bool
        expectedCode  int
        expectedBody  string
    }{
        {
            description:   "index route",
            route:         "/",
            expectedError: false,
            expectedCode:  200,
            expectedBody:  "OK",
        },
        {
            description:   "non existing route",
            route:         "/i-dont-exist",
            expectedError: false,
            expectedCode:  404,
            expectedBody:  "Cannot GET /i-dont-exist",
        },
    }

    // Setup the app as it is done in the main function
    app := Setup()

    // Iterate through test single test cases
    for _, test := range tests {
        // Create a new http request with the route
        // from the test case
        req, _ := http.NewRequest(
            "GET",
            test.route,
            nil,
        )

        // Perform the request plain with the app.
        // The -1 disables request latency.
        res, err := app.Test(req, -1)

        // verify that no error occured, that is not expected
        assert.Equalf(t, test.expectedError, err != nil, test.description)

        // As expected errors lead to broken responses, the next
        // test case needs to be processed
        if test.expectedError {
            continue
        }

        // Verify if the status code is as expected
        assert.Equalf(t, test.expectedCode, res.StatusCode, test.description)

        // Read the response body
        body, err := io.ReadAll(res.Body)

        // Reading the response body should work everytime, such that
        // the err variable should be nil
        assert.Nilf(t, err, test.description)

        // Verify, that the reponse body equals the expected body
        assert.Equalf(t, test.expectedBody, string(body), test.description)
    }

    // Setup Setup a fiber app with all of its routes
    func Setup() *fiber.App {
        // Initialize a new app
        app := fiber.New()

        // Register the index route with a simple
        // "OK" response. It should return status
        // code 200
        app.Get("/", func(c *fiber.Ctx) error {
            return c.SendString("OK")
        })

        // Return the configured app
        return app
    }
}
```

## 핸들러 로직만 테스트

`fiber.App`을 사용하는 테스트와 직접 정의한 핸들러를 테스트하는 방법은 약간 다릅니다. `fiber.App` 테스트는 전체 HTTP 서버의 동작을 시뮬레이션합니다. 반면, 개별 핸들러 함수를 테스트하는 것은 해당 함수의 로직에 집중하는 접근 방식입니다.

`type MyApp{}`과 `func (m *MyApp) myHandler(ctx *fiber.Ctx) error` 같은 구조체와 메서드를 가지고 있다면, 테스트는 다음 두 가지 방식 중 하나로 수행할 수 있습니다:

### 1. 핸들러를 Fiber 라우트에 바인딩

`fiber.App` 인스턴스에 핸들러를 바인딩하여 테스트를 수행합니다. 이 방식은 핸들러가 `fiber.Ctx`를 사용하는 방식을 포함하여 전체 HTTP 요청/응답 사이클을 테스트합니다.

```go
func TestMyHandler(t *testing.T) {
    app := fiber.New()

    // MyApp 인스턴스 생성
    myApp := &MyApp{}

    // myHandler를 라우트에 바인딩
    app.Get("/myroute", myApp.myHandler)

    // http.Request 생성
    req, _ := http.NewRequest("GET", "/myroute", nil)

    // fiber.App의 Test 메서드를 사용해 테스트 수행
    resp, err := app.Test(req)
    if err != nil {
        t.Errorf("Error during request: %v", err)
    }

    // 테스트 결과 검증
    // 예: resp.StatusCode를 검사하거나, resp.Body를 읽어 검사 등
}
```

### 2. 핸들러 로직만 직접 테스트

이 방법은 `fiber.Ctx`를 모의(Mock)하여 핸들러 로직만 직접 테스트합니다. `fiber.Ctx`를 모의하기 위해서는 별도의 라이브러리나 모킹 도구를 사용해야 할 수 있습니다. 이 방식은 HTTP 라이프사이클 밖에서 핸들러의 순수한 기능적 측면을 테스트하는 데 적합합니다.

```go
func TestMyHandlerLogic(t *testing.T) {
    // fiber.Ctx 모의 생성
    ctx := MockFiberCtx("GET", "/myroute", nil)

    // MyApp 인스턴스 생성
    myApp := &MyApp{}

    // 핸들러 호출
    err := myApp.myHandler(ctx)
    if err != nil {
        t.Errorf("Error in handler: %v", err)
    }

    // 핸들러의 로직 검증
    // 예: ctx.Response()를 검사하거나, 핸들러가 설정한 특정 값들을 검사 등
}
```

`MockFiberCtx` 함수는 `fiber.Ctx`의 동작을 모방하는 함수로, 실제 `fiber.Ctx` 인터페이스를 구현하는 모의 객체를 생성하거나, Fiber의 공식적인 테스트 유틸리티(있을 경우)를 사용합니다. 현재 Fiber에는 `fiber.Ctx`를 직접 모킹하는 공식적인 방법이 없으므로, 필요한 기능을 모킹하는 사용자 정의 함수나 모킹 라이브러리를 사용해야 합니다.

두 번째 방법은 `fiber.Ctx`에 의존하는 로직이 많은 경우 구현하기 어려울 수 있으며, 전체적인 HTTP 처리 흐름을 테스트하지 않습니다. 따라서 첫 번째 방법이 일반적으로 더 권장되는 접근 방식입니다.

## 기타

### 참고

- [How do I run a set of unit test?](https://github.com/gofiber/fiber/issues/756)
