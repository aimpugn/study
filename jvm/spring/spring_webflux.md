# Spring Webflux

- [Spring Webflux](#spring-webflux)
    - [coRouter filter](#corouter-filter)
    - [기타](#기타)

## coRouter filter

- `filter`는 `handler`들이 요청을 처리하는 곳 앞뒤로 공통의 동작을 정의 가능
- filter는 사용자가 정의하는 `filterFunction` 이라는 함수를 파라미터로 받음

```kotlin
// 1) ServerRequest `request`
// 2) ServerResponse를 리턴하는 `handler` suspend 함수
// handler(request) 앞뒤로 여러 가지 동작을 정의할 수 있는 간단한 함수
filter { request, handler -> handler(request) }
```

## 기타

- [Spring WebFlux 에서 coRouter filter를 이용하여 request, response 로깅하기](https://medium.com/riiid-teamblog-kr/spring-webflux-%EC%97%90%EC%84%9C-corouter-filter%EB%A5%BC-%EC%9D%B4%EC%9A%A9%ED%95%98%EC%97%AC-request-response-%EB%A1%9C%EA%B9%85%ED%95%98%EA%B8%B0-df56f9d9680)
