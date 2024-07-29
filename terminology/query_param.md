# Query and Param

- [Query and Param](#query-and-param)
    - [Query와 Param](#query와-param)

## Query와 Param

Query는 [RFC_3986](../standards/RFC_3986.md)을 참고해주세요.

'param'은 'param'은 'parameter'의 줄임말로, 다양한 프로그래밍 컨텍스트에서 단일 속성이나 값을 나타내는 데 사용됩니다.
다음과 같은 특징을 가집니다:

1. **단일성**: *하나의 개별적인 값이나 속성*을 나타냅니다.
2. **구체성**: 특정 데이터 항목이나 설정을 지칭합니다.
3. **컨텍스트 의존성**: 사용되는 맥락에 따라 의미가 달라질 수 있습니다.

이는 URL, 함수, 설정, HTTP 요청 등 여러 영역에서 일관되게 적용되는 개념입니다.

```http
https://api.example.com/users?id=123&sort=name&filter=active
```

`id=123&sort=name&filter=active`에서 `id`, `sort`, `filter` 등이 'query param'에 해당합니다.

다음 예제는 다양한 컨텍스트에서 'param'의 사용을 보여줍니다:

```go
package main

import (
    "fmt"
    "net/http"
    "strconv"
)

// 설정 파라미터
type Config struct {
    Host string
    Port int
}

// 함수 파라미터
func createUser(name string, age int) string {
    return fmt.Sprintf("Created user: %s, age: %d", name, age)
}

// HTTP 핸들러
func handler(w http.ResponseWriter, r *http.Request) {
    // URL 쿼리 파라미터
    query := r.URL.Query()
    action := query.Get("action")

    // HTTP 헤더 파라미터
    userAgent := r.Header.Get("User-Agent")

    // 응답 작성
    fmt.Fprintf(w, "Action: %s, User-Agent: %s", action, userAgent)
}

func main() {
    // 설정 파라미터 사용
    config := Config{Host: "localhost", Port: 8080}

    // 함수 파라미터 사용
    result := createUser("Alice", 30)
    fmt.Println(result)

    // HTTP 서버 시작
    http.HandleFunc("/", handler)
    addr := config.Host + ":" + strconv.Itoa(config.Port)
    fmt.Printf("Server starting on %s\n", addr)
    http.ListenAndServe(addr, nil)
}
```

이 예제에서:
1. `Config` 구조체의 `Host`와 `Port`는 설정 파라미터입니다.
2. `createUser` 함수의 `name`과 `age`는 함수 파라미터입니다.
3. HTTP 핸들러에서 `action`은 URL 쿼리 파라미터이고, `userAgent`는 HTTP 헤더 파라미터입니다.
