# Go rod

- [Go rod](#go-rod)
    - [테스트용 패키지 구성](#테스트용-패키지-구성)
    - [예제](#예제)
        - [IFrame 내에서 postMessage를 보내고, 서버와 통신하여 form이 올바르게 구성되었는지 테스트](#iframe-내에서-postmessage를-보내고-서버와-통신하여-form이-올바르게-구성되었는지-테스트)

## 테스트용 패키지 구성

테스트 코드는 프로젝트 구조에 맞게 별도의 패키지로 구성하는 것이 좋습니다.
이때 패키지는 `server` 디렉토리 내에 `tests` 디렉토리를 만들어서 관리할 수 있습니다.
테스트용 패키지의 이름은 `tests`로 하거나 `integration_tests`로 하여 명확하게 구분할 수 있습니다.

```plaintext
server/
    fixtures/
    models/
    repositories/
    templates/
    tests/
        integration_test.go
```

## 예제

### IFrame 내에서 postMessage를 보내고, 서버와 통신하여 form이 올바르게 구성되었는지 테스트

아래는 Go Fiber와 Go Rod를 사용하여 IFrame 내에서 postMessage를 보내고, 서버와 통신하여 form이 올바르게 구성되었는지 테스트하는 코드의 예제입니다.

```go
// integration_test.go
package tests

import (
    "context"
    "testing"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
)

// 테스트를 위한 서버를 설정하고 실행합니다. 
// 이 코드는 `setupServer` 함수에서 설정한 Fiber 애플리케이션을 백그라운드에서 실행합니다.
func setupServer() *fiber.App {
    app := fiber.New()
    // 서버 설정 및 라우팅 설정
    return app
}

func TestPostMessageAndFormRendering(t *testing.T) {
    app := setupServer()
    go func() {
        if err := app.Listen(":3000"); err != nil {
            t.Fatalf("Server failed to start: %v", err)
        }
    }()
    time.Sleep(1 * time.Second) // 서버 시작을 위한 대기 시간

    // Rod 브라우저 설정
    browser := rod.New().MustConnect()
    defer browser.MustClose()

    // Rod 브라우저를 설정하고 테스트할 페이지를 엽니다.
    page := browser.MustPage("http://localhost:3000/some_template")
    page.MustWaitLoad()

    // IFrame 내부로 이동
    iframe := page.MustElement("iframe").MustFrame()

    // postMessage 전송
    iframe.MustEval(`window.postMessage({ type: "testType", data: "testData" }, "*")`)

    // postMessage 수신 및 처리 대기
    result := iframe.MustWaitEvent(`message`, func(js *rod.Event) bool {
        msg := js.Message
        // 메시지 타입 확인
        if msg["type"] == "testType" {
            return true
        }
        return false
    }, 10*time.Second)

    assert.NotNil(t, result, "Expected to receive a postMessage event")

    // 서버와 통신하여 폼 데이터 확인
    formElement := iframe.MustElement("form")
    assert.NotNil(t, formElement, "Expected form element to be present")

    inputValue := formElement.MustElement("input[name='expectedField']").MustProperty("value").String()
    assert.Equal(t, "expectedValue", inputValue, "Form input value mismatch")
}
```
