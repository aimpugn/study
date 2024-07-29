# Go Fiber HTTP

- [Go Fiber HTTP](#go-fiber-http)
    - [`fiber.Client`](#fiberclient)
    - [`fiber.Client`와 `http.Client`](#fiberclient와-httpclient)
    - [`defer body.Close()` 불필요](#defer-bodyclose-불필요)
        - [`net/http.Client`](#nethttpclient)
        - [`fiber.Client`](#fiberclient-1)
    - [우려되는 점들](#우려되는-점들)
        - [데이터 오염 우려](#데이터-오염-우려)
        - [동시성 문제](#동시성-문제)
        - [성능과 동기화 메커니즘](#성능과-동기화-메커니즘)

## `fiber.Client`

Fiber의 HTTP 클라이언트 구현인 `fiber.Client`는 FastHTTP의 `HostClient`를 기반으로 구축되었고, 다양한 편리한 헬퍼  메서드들을 제공한다.
예를 들어, HTTP 요청 방법에 대한 전용 메서드들이 있다.

`fiber.Client`는 사용하기 쉽고 직관적이라는 점이 강조된다.
예를 들어, `fiber.Get` 또는 `fiber.Post`와 같은 메서드를 사용하여 HTTP 요청을 시작할 수 있으며, 이를 통해 프록시 시뮬레이션 같은 작업을 수행할 수 있다.

이러한 특징들을 고려할 때, Fiber의 `fiber.Client`는 특히 Fiber 프레임워크 내에서의 사용에 최적화되어 있다.
FastHTTP를 기반으로 하기 때문에 빠른 성능을 제공하는 것이 중요한 설계 요소다.

반면, `net/http`의 `http.Client`는 Go 표준 라이브러리의 일부이며, 더 일반적이고 범용적인 사용을 목적으로 한다.
따라서 `fiber.Client`가 고루틴에서의 동시 사용에 안전하다고 표현하는 것은 Fiber 프레임워크의 사용 사례와 성능 요구사항을 충족하기 위한 것이다.

## `fiber.Client`와 `http.Client`

- Go의 `net/http` 패키지의 `http.Client`와 Go Fiber의 `fiber.Client`는 두 가지 다른 목적과 설계 철학을 가지고 있다고 한다.
- Go Fiber의 문서에서 언급된 "It is safe calling Client methods from concurrently running goroutines"라는 문장은 `fiber.Client`가 고루틴 사이에서 안전하게 사용될 수 있다는 것을 의미하며, 이것은 Go Fiber의 설계 목적 중 하나를 반영한 것이다.
- 그러나 이것이 `net/http`의 `http.Client`가 고루틴에서 안전하지 않다는 것을 의미하지는 않는다. 실제로 `net/http`의 `http.Client`는 고루틴에서 안전하게 사용될 수 있도록 설계되었다. Go의 공식 문서에 따르면 `http.Client`는 고루틴 사이에서 안전하게 공유될 수 있으며, 여러 고루틴에서 동시에 요청을 보내는 데 사용될 수 있다.
- `fiber.Client`가 강조하는 것은 Fiber 프레임워크 내부의 컨텍스트와 통합을 더 잘 제공한다는 점일 수 있다. Fiber는 웹 애플리케이션 개발에 특화된 프레임워크이며, 해당 프레임워크 내에서의 클라이언트 사용에 최적화되어 있을 수 있다. 반면, `net/http`의 `http.Client`는 Go 언어 자체에서 제공하는 더 일반적이고 범용적인 HTTP 클라이언트이다.

## `defer body.Close()` 불필요

### `net/http.Client`

- `net/http` 패키지를 사용할 때, `http.Response` 객체의 `Body`는 **네트워크 스트림에 직접 연결**되어 있다.
- 사용자는 수동으로 `Body` 스트림을 닫아서 네트워크 리소스를 해제해야 하므로, `defer resp.Body.Close()` 호출이 필요하다
- `Body`를 닫지 않으면 네트워크 연결이 열린 상태로 남아 메모리 누수와 같은 문제를 일으킬 수 있다.

### `fiber.Client`

- `fiber.Client`의 경우, 네트워크 응답을 처리하는 방식이 다르다. `fiber.Client`는 내부적으로 `fasthttp` 패키지를 사용한다
- `fasthttp`는 `net/http`와 달리, **요청과 응답에 사용되는 메모리 버퍼를 재사용**한다. 따라서 `fiber.Client`로 요청을 보내고 응답을 받으면, `fasthttp`가 응답 본문을 버퍼에 저장하고 이를 관리한다.
- 이로 인해 사용자가 별도로 응답의 `Body`를 닫을 필요가 없다. 실제로 `fasthttp`의 응답 객체에는 `Body`를 닫는 메서드가 존재하지 않는다.

## 우려되는 점들

### 데이터 오염 우려

- `fasthttp`는 요청과 응답을 처리할 때 사용하는 메모리를 재사용함으로써 효율성을 높입니다. 이것은 메모리 할당과 가비지 컬렉션에 대한 오버헤드를 줄입니다.
- 재사용된 버퍼는 새 요청이나 응답에 사용되기 전에 항상 초기화됩니다. 이는 데이터 오염을 방지하는 중요한 단계입니다. 즉, 이전 요청이나 응답의 데이터가 새 요청/응답에 영향을 미치지 않도록 합니다.

### 동시성 문제

- `fasthttp`는 고루틴에서 안전하게 사용할 수 있도록 설계되었습니다. 즉, 동시성 문제를 고려하여 구현되었습니다.
- `fasthttp`의 내부 구현에서는 고루틴 간의 데이터 충돌을 방지하기 위해 필요한 경우 적절한 동기화 메커니즘을 사용합니다.

### 성능과 동기화 메커니즘

- 성능과 동시성은 항상 균형을 맞춰야 하는 부분입니다. 잠금 메커니즘(예: 뮤텍스, 세마포어)은 필요한 경우에만 사용되며, 이는 성능 저하를 최소화하기 위한 조치입니다.
- `fasthttp`는 성능을 크게 향상시키기 위해 동기화 비용을 최소화하는 방법으로 설계되었습니다. 따라서, 필요한 경우에만 동기화를 수행하며, 이는 성능에 미치는 영향을 가능한 한 낮춥니다.

결론적으로, `fasthttp`와 같은 라이브러리는 성능과 안정성을 모두 고려하여 설계되었으며, 데이터 오염이나 동시성 문제, 성능 저하에 대한 우려는 내부적으로 잘 처리되고 있습니다. 물론, 모든 HTTP 클라이언트 라이브러리와 마찬가지로, 특정 사용 사례에 적합한지는 개별 프로젝트의 요구 사항에 따라 달라질 수 있습니다.
