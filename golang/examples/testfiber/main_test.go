package main

import (
    "github.com/gofiber/fiber/v3"
    "net/http"
    "net/http/httptest"
    "testing"
)

func BenchmarkFiber(b *testing.B) {
    app := fiber.New()

    app.Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, Fiber!")
    })

    req := httptest.NewRequest("GET", "/", nil)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        resp, _ := app.Test(req)
        _ = resp.Body.Close() // 응답에서 메모리 할당 관리를 위해 닫아줌
    }
}

func BenchmarkFiberImmuable(b *testing.B) {
    app := fiber.New(fiber.Config{
        Immutable: true,
    })

    app.Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, Fiber with Immutable!")
    })

    req := httptest.NewRequest("GET", "/", nil)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        resp, _ := app.Test(req)
        _ = resp.Body.Close() // 응답에서 메모리 할당 관리를 위해 닫아줌
    }
}

func BenchmarkNetHTTP(b *testing.B) {
    // 핸들러 정의
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, http!"))
    })

    // 테스트 서버 생성
    server := httptest.NewServer(handler)
    defer server.Close()

    // 실제 클라이언트 요청 생성
    client := server.Client()

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        // 서버 URL로 요청을 보냄
        req, _ := http.NewRequest("GET", server.URL, nil)
        resp, _ := client.Do(req)
        // HTTP 응답의 Body는 지연 로드(deferred loading) 방식으로 처리됩니다.
        // 즉, 응답 본문은 요청 시 바로 메모리로 로드되지 않으며, 필요할 때만 네트워크에서 데이터를 가져옵니다.
        // 따라서, 클라이언트가 네트워크로부터 데이터를 완전히 읽기 전에 Close()를 호출하는 경우,
        // 연결이 적절히 종료되지 않아서 nil 참조 문제가 발생할 수 있습니다.
        content := make([]byte, 256)
        resp.Body.Read(content)
        _ = resp.Body.Close() // 응답 Body 닫기
    }
}
