# 왜 HTTP 클라이언트 timeout은 보통 요청별보다 클라이언트 기본값으로 수렴할까

짧게 답하면, **표준이 그렇게 강제해서라기보다 낮은 계층의 기본 단위가 request가 아니라 socket, connection, stream이기 때문**입니다. 그래서 `connect timeout`, `socket read timeout` 같은 값은 자연스럽게 연결 쪽으로 붙고, `이번 요청은 5초 안에 끝나야 한다` 같은 값은 요청 전체 마감시간(deadline)이나 취소(cancellation)로 표현되는 경우가 많습니다.

이 문서는 먼저 timeout이 하나가 아니라는 점부터 분리하고, 그다음 왜 운영체제와 HTTP의 발전 과정이 지금의 API 모양을 만들었는지 설명합니다. 끝에서는 JDK, Apache HttpClient, Go, requests, aiohttp, libcurl, OkHttp, Fetch 같은 라이브러리들이 각각 어느 지점을 노출하는지도 정리합니다.

## 먼저, `timeout`은 하나가 아니다

많은 혼란은 여기서 시작합니다. 우리가 보통 "timeout"이라고 한 단어로 부르는 것은 실제로는 서로 다른 층의 제한 시간 묶음입니다.

| 이름 | 실제로 제한하는 것 | 자연스럽게 붙는 곳 |
| --- | --- | --- |
| connect timeout | 새 연결을 맺는 데 허용하는 시간 | socket / connection / client |
| socket read timeout | 소켓에서 다음 바이트를 기다리는 시간 | socket / connection |
| socket write timeout | 소켓에 쓰기가 오래 막힐 때 허용하는 시간 | socket / connection |
| response header timeout | 요청을 보낸 뒤 응답 첫 줄/헤더를 기다리는 시간 | request / exchange |
| pool acquire timeout | pool에서 사용 가능한 연결을 기다리는 시간 | request / connection manager |
| request deadline / call timeout | 요청 전체를 끝내야 하는 총 시간 | request |
| idle keepalive timeout | 놀고 있는 연결을 얼마나 유지할지 | connection pool / client |

이 표를 먼저 머리에 넣으면, 왜 어떤 라이브러리는 "요청별 timeout"을 준다고 해도 사실은 `전체 마감시간`만 주고, 어떤 라이브러리는 `connect/read`를 분리하면서도 클라이언트 쪽에 두는지 이해하기 쉬워집니다.

## 한 장으로 보면 이런 구조다

```text
애플리케이션 코드
  |
  |  "이 요청은 5초 안에 끝나야 해"
  v
HTTP 클라이언트 / connection pool / stream 관리
  |
  |  새 연결이 필요하면 connect timeout 적용
  |  이미 있는 연결이면 connect 단계 자체가 없음
  |  필요하면 요청 전체 deadline / cancellation 관리
  v
socket / file descriptor
  |
  |  SO_RCVTIMEO, SO_SNDTIMEO 같은 socket 옵션
  |  connect(), recv(), send() 같은 syscall
  v
kernel / TCP / NIC driver / network
```

핵심은 **커널은 "HTTP request 객체"를 모른다**는 점입니다. 커널이 아는 것은 소켓과 연결, 그리고 그 위에서 흐르는 바이트입니다. HTTP request라는 개념은 대부분 사용자 공간(user space)의 라이브러리가 만들어서 관리합니다.

## 표준이 직접 이렇게 만들었나

직접적으로는 아닙니다.

HTTP RFC는 메시지 형식, 연결 유지, stream과 frame의 의미 같은 **프로토콜 규칙**을 다룹니다. 하지만 "자바나 Go의 클라이언트 API는 connect timeout을 Builder에 둬라" 같은 식으로 **라이브러리 API 표면**까지 정하지는 않습니다.

- HTTP/1.1 RFC는 persistent connection을 규정하지만, keep-alive 시간을 몇 초로 두고 어떤 setter로 노출할지는 정하지 않습니다. [RFC 7230](https://www.rfc-editor.org/rfc/rfc7230.html)
- HTTP/2 RFC는 하나의 연결에 여러 stream을 동시에 실을 수 있다고 규정합니다. 이건 timeout API에 큰 영향을 주지만, 여전히 API 형태 자체를 규정하지는 않습니다. [RFC 9113](https://www.rfc-editor.org/rfc/rfc9113.html)
- QUIC도 하나의 연결에서 여러 stream을 동시 운반한다고 설명합니다. [RFC 9000](https://www.rfc-editor.org/rfc/rfc9000.html)

즉, **표준이 API를 강제했다기보다, 표준이 만들어 놓은 실행 환경에 라이브러리들이 적응하면서 비슷한 설계로 수렴한 것**에 가깝습니다.

## 왜 낮은 계층은 request보다 socket을 먼저 보나

운영체제와 커널 입장에서 네트워크 I/O의 기본 단위는 보통 소켓입니다.

- `socket(7)`은 `SO_RCVTIMEO`, `SO_SNDTIMEO` 같은 옵션을 **socket option**으로 설명합니다. 이름 그대로 socket에 붙는 옵션입니다. [socket(7)](https://man7.org/linux/man-pages/man7/socket.7.html)
- `connect(2)`는 새 연결을 맺는 시스템 호출입니다. 이 단계는 요청마다 항상 있는 것이 아니라, **새 연결이 필요할 때만** 있습니다. [connect(2)](https://man7.org/linux/man-pages/man2/connect.2.html)
- `recv(2)`는 흥미로운 대비를 보여 줍니다. `MSG_DONTWAIT`는 **per-call 옵션**이지만, `O_NONBLOCK`는 open file description 쪽 설정입니다. 즉 낮은 계층도 "호출마다 다르게"와 "소켓 자체 상태"를 분리해서 봅니다. [recv(2)](https://man7.org/linux/man-pages/man2/recv.2.html)

아주 단순화하면 커널 쪽 그림은 이렇습니다.

```text
fd = socket()
setsockopt(fd, SO_RCVTIMEO, 3s)
connect(fd, addr)
send(fd, bytes)
recv(fd, buf)
```

여기에는 `requestId=1234` 같은 개념이 없습니다. 커널이 보는 것은 `fd` 하나입니다.

그래서 라이브러리가 `socket read timeout` 같은 것을 노출할 때, 기본적으로는 소켓이나 연결에 붙이는 것이 더 자연스럽습니다. 요청별 timeout은 대개 그 위에서 라이브러리가 별도 타이머를 두고 취소하거나, 특정 단계에만 override를 거는 식으로 구현합니다.

여기서 NIC driver도 request를 알지 못합니다. driver는 결국 프레임을 보내고 받는 하드웨어 쪽 경계에 가깝고, "이 바이트가 HTTP 요청 A의 3번째 chunk인지" 같은 해석은 더 위 계층의 책임입니다. TCP 재전송, 혼잡 제어, 수신 버퍼 관리, 연결 상태 전이도 주로 커널 TCP 스택이 맡습니다. 즉 낮은 계층으로 내려갈수록 "요청별 timeout"보다 "연결과 소켓의 상태"가 더 본질적인 단위가 됩니다.

## 왜 connect timeout은 특히 request보다 connection에 더 가깝나

이건 request 단위로 생각하면 오히려 더 헷갈립니다.

JDK `HttpClient.Builder.connectTimeout(...)` 문서는 새 연결을 만들 때 이 timeout이 적용된다는 점을 드러냅니다. 반면 `HttpRequest.Builder.timeout(...)`은 요청 전체의 지정된 시간 제한을 말합니다. [JDK HttpClient.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpClient.Builder.html) [JDK HttpRequest.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpRequest.Builder.html)

왜 이렇게 나뉘느냐를 작은 예로 보면 바로 보입니다.

```text
request A -> 기존 keep-alive 연결 재사용 -> connect 단계 없음
request B -> 새 TCP 연결 필요         -> connect 단계 있음
```

둘 다 "요청"이지만, connect timeout이 실제로 의미 있는 것은 `request B`뿐입니다. 그래서 connect timeout을 request 속성처럼 보이게 만들면, 어떤 요청에는 적용되고 어떤 요청에는 적용되지 않는 값이 됩니다. 이건 사용자에게는 직관적이지 않아 보일 수 있지만, 실제로는 **연결 단계가 request의 항상-존재 속성이 아니기 때문**입니다.

## HTTP/1.0 때는 왜 덜 이상해 보였을까

초기에는 `한 요청 = 한 연결`에 가까웠습니다. HTTP/1.0 시절에는 요청 하나 보내고 연결을 닫는 그림이 더 흔했습니다.

```text
request 1 -> connect -> send -> recv -> close
request 2 -> connect -> send -> recv -> close
```

이 구조에서는 request와 connection이 거의 겹칩니다. 그래서 "요청별 connect timeout"이라고 말해도 큰 어색함이 없습니다.

하지만 HTTP/1.1의 persistent connection이 보편화되면서 상황이 달라집니다. [RFC 7230](https://www.rfc-editor.org/rfc/rfc7230.html)

```text
connect 1번
  -> request A
  -> request B
  -> request C
close
```

이제 timeout은 적어도 두 덩어리로 갈라집니다.

- 이 연결을 새로 만들 때 얼마나 기다릴까
- 이미 만들어진 이 연결을 얼마나 오래 재사용할까

이 시점부터 `request timeout`과 `connection timeout`은 더 이상 같은 말을 할 수 없게 됩니다.

## connection pool이 들어오면 timeout 종류가 더 늘어난다

keep-alive만으로도 계층이 나뉘는데, 여기에 connection pool이 들어오면 또 하나의 대기 구간이 생깁니다.

```text
request
  -> pool에서 연결 빌리기 대기
  -> 없으면 새 연결 생성
  -> 요청 전송
  -> 응답 헤더 대기
  -> 응답 바디 읽기
```

이제 하나의 숫자로는 각 단계의 의미를 제대로 표현하기 어렵습니다.

- pool에서 100ms 넘게 기다리면 실패시키고 싶은가
- 새 TCP/TLS 연결은 1초까지 허용할 것인가
- 응답 전체는 5초 안에 끝나야 하는가

Apache HttpComponents는 이 복잡도를 아예 문서에서 계층별로 나눠 설명합니다.

- socket 설정은 per-socket
- HTTP/1.1, HTTP/2 설정은 per-connection
- request 설정은 per-request
- TLS 설정은 per-host
- connection 설정은 per-route

이 문서 자체가 "timeout은 하나가 아니라 어느 층의 속성이냐가 중요하다"는 걸 보여 줍니다. [Apache Configuration Guide](https://hc.apache.org/httpcomponents-client-5.6.x/configuration.html)

## 발전 과정을 짧게 이어 보면

API가 왜 지금처럼 생겼는지를 가장 짧게 요약하면 아래 흐름입니다.

1. HTTP/1.0에 가까운 시절에는 `한 요청 = 한 연결`이라 request와 connection을 거의 같은 것으로 다뤄도 큰 문제가 없었다.
2. HTTP/1.1 keep-alive가 보편화되면서 "연결 생성"과 "요청 실행"이 분리됐다.
3. connection pool이 들어오면서 "pool 대기", "새 연결 생성", "응답 대기"가 서로 다른 timeout 후보가 됐다.
4. TLS, proxy CONNECT, DNS 같은 단계가 커지면서 connect phase 자체가 길고 복합적인 단계가 됐다.
5. 비동기/event loop 기반 라이브러리는 request task에 timer를 얹기 쉬워져 request deadline, cancellation 모델이 강해졌다.
6. HTTP/2, HTTP/3는 한 연결에서 여러 stream을 동시에 처리하므로 transport timeout을 request별로 단순 매핑하기가 더 어려워졌다.

즉 지금의 API는 어느 날 누가 한 번에 설계한 정답이라기보다, **프로토콜과 런타임이 복잡해질수록 timeout도 계층별로 쪼개질 수밖에 없었던 결과**에 가깝습니다.

## HTTP/2부터는 왜 per-request read timeout이 더 어색해지나

HTTP/2는 하나의 TCP 연결 안에 여러 stream을 동시에 실을 수 있습니다. RFC 9113은 한 연결에 여러 concurrent stream이 있을 수 있고, frame이 서로 interleave될 수 있다고 설명합니다. [RFC 9113](https://www.rfc-editor.org/rfc/rfc9113.html)

그림으로 보면 이렇습니다.

```text
TCP connection 1개
  |- stream 1: GET /a
  |- stream 3: GET /b
  |- stream 5: POST /c
```

이때 socket read timeout을 request별로 생각해 보면 문제가 바로 드러납니다.

```text
socket fd 42
  -> stream 1 은 100ms read timeout을 원함
  -> stream 3 은 10s read timeout을 원함

setsockopt(fd, SO_RCVTIMEO, ?)
```

이 소켓에는 값이 하나만 붙습니다. 그런데 소켓 위에는 여러 request가 동시에 올라가 있습니다. 그러니 낮은 계층의 read timeout을 요청별로 깔끔하게 매핑하기가 어렵습니다.

그래서 HTTP/2 이후에는 라이브러리들이 보통 아래 둘 중 하나를 택합니다.

1. 소켓/연결 쪽 timeout은 shared default로 유지한다.
2. 요청별로는 전체 deadline이나 cancellation을 건다.

이건 "라이브러리가 게을러서"가 아니라, **공유 연결 위에서 request별 transport timeout을 직접 모델링하는 것이 본질적으로 더 까다롭기 때문**입니다.

## HTTP/3와 QUIC가 오면 다시 request별이 쉬워지나

부분적으로는 그렇고, 전체적으로는 여전히 아닙니다.

QUIC는 transport 자체가 여러 stream을 동시에 운반하도록 설계되어 있습니다. [RFC 9000](https://www.rfc-editor.org/rfc/rfc9000.html) 그래서 TCP 위 HTTP/2보다 "stream 취소" 같은 개념을 transport 차원에서 더 잘 표현합니다.

하지만 그렇다고 모든 timeout이 곧바로 request별이 되지는 않습니다.

- DNS
- 연결 수립과 경로 검증
- TLS/암호화 세션
- connection migration
- congestion control
- idle timeout
- pool / session 재사용

이런 것들은 여전히 connection 또는 client 쪽 책임이 큽니다. 그래서 HTTP/3 시대에도 보통은

- connection/client 기본 설정
- request 전체 deadline / cancellation

이 조합이 계속 중요합니다.

## 결국 라이브러리는 어떤 방향으로 수렴했나

대체로 세 가지 패턴이 많습니다.

### 1. 연결 성격의 timeout은 client/connection 기본값

이 패턴은 가장 흔합니다.

- JDK `HttpClient`: connect timeout은 client builder, request timeout은 request builder [JDK HttpClient.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpClient.Builder.html) [JDK HttpRequest.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpRequest.Builder.html)
- Apache HttpClient: socket / connection / request를 계층별로 분리 [Apache Configuration Guide](https://hc.apache.org/httpcomponents-client-5.6.x/configuration.html)
- OkHttp: client builder에 `connectTimeout`, `readTimeout`, `callTimeout` 제공 [OkHttpClient.Builder](https://square.github.io/okhttp/3.x/okhttp/okhttp3/OkHttpClient.Builder.html)

이쪽은 "shared client는 공유하고, transport 성격은 기본 정책으로 둔다"는 철학에 가깝습니다.

### 2. 요청별로는 전체 마감시간이나 cancellation을 준다

이 패턴도 매우 흔합니다.

- JDK `HttpRequest.Builder.timeout(...)`은 요청 전체 제한 시간 성격입니다. [JDK HttpRequest.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpRequest.Builder.html)
- Go `http.Client.Timeout`은 "이 Client가 만드는 request에 대한 시간 제한"을 설명합니다. [Go net/http](https://pkg.go.dev/net/http#Client)
- 브라우저 Fetch는 전통적인 `timeout=` 파라미터보다 `AbortController`, `AbortSignal.timeout(...)` 같은 취소 모델로 가는 편입니다. [DOM AbortSignal.timeout](https://dom.spec.whatwg.org/#abortsignal-timeout) [Fetch Standard](https://fetch.spec.whatwg.org/)

이쪽은 "이번 요청이 언제까지 끝나야 하는가"라는 업무 관점을 표현하기 좋습니다.

### 3. 요청별 phase timeout을 비교적 풍부하게 노출한다

예외처럼 보이지만, 사실 이 경우도 내부를 보면 계층 구분이 살아 있습니다.

- `aiohttp`는 session 기본 timeout을 두고 request별 override도 허용합니다. 동시에 `total`, `connect`, `sock_connect`, `sock_read`를 나눕니다. 즉 "요청별 override"를 주면서도 phase를 계속 구분합니다. [aiohttp Client Quickstart](https://docs.aiohttp.org/en/stable/client_quickstart.html#timeouts)
- `requests`는 `timeout` 인자를 request 호출에 직접 받지만, quickstart는 이것이 "전체 응답 다운로드 시간 제한"이 아니라 **기본 소켓에서 바이트가 오지 않는 시간**과 관련 있다고 설명합니다. [Requests Quickstart](https://requests.readthedocs.io/en/latest/user/quickstart/#timeouts)

즉 per-request API를 준다고 해서 "timeout이 본질적으로 request-native"라는 뜻은 아닙니다. 어떤 timeout인지를 계속 따져봐야 합니다.

### 4. 아예 transfer handle을 중심으로 설계한다

`libcurl`은 이 점이 아주 선명합니다.

- `CURLOPT_TIMEOUT`은 **전체 transfer** 시간 제한
- `CURLOPT_CONNECTTIMEOUT`은 **연결 단계** 시간 제한

libcurl 문서는 connect timeout이 DNS와 handshake까지 포함하는 연결 단계에만 적용되고, 전체 timeout은 transfer 전체를 덮는다고 설명합니다. [CURLOPT_TIMEOUT](https://curl.se/libcurl/c/CURLOPT_TIMEOUT.html) [CURLOPT_CONNECTTIMEOUT](https://curl.se/libcurl/c/CURLOPT_CONNECTTIMEOUT.html)

이 모델이 가능한 이유는 libcurl이 `easy handle` 같은 **transfer 객체**를 중심으로 설정을 갖기 때문입니다. 즉 어떤 추상화 객체가 timeout을 소유하는지가 다릅니다.

## 왜 어떤 라이브러리는 요청별 override를 더 쉽게 제공할까

이것도 우연이 아니라 구조 차이입니다.

### transfer 객체가 명확하면 쉽다

libcurl처럼 "이 handle 하나가 이번 transfer"라는 객체가 있으면, 그 객체에 timeout을 붙이는 것이 자연스럽습니다.

```text
easy handle 1 = request A 설정
easy handle 2 = request B 설정
```

이 경우 timeout도 handle별로 들고 다니기 쉽습니다.

### event loop / coroutine 기반이면 user-space timer를 얹기 쉽다

aiohttp 같은 비동기 라이브러리는 event loop가 이미 작업(task) 단위를 관리합니다. 그러면 `socket read timeout` 자체를 바꾸지 않고도, "이 task는 200ms 안에 다음 단계로 못 가면 취소" 같은 정책을 얹기 쉽습니다.

즉 lower-level timeout을 request에 붙인다기보다, **request task에 timer를 붙이는 것**에 가깝습니다.

### shared client를 불변 객체처럼 다루면 파생 client 전략도 가능하다

일부 라이브러리는 기본 client를 공유하되, 약간 다른 timeout을 가진 파생 client를 가볍게 만들 수 있습니다. 이 경우도 request마다 shared mutable state를 바꾸는 것이 아니라, **새 policy object를 하나 더 만드는 방식**으로 문제를 피합니다.

## 왜 "요청마다 connect/read/write 다 다르게"는 예외처럼 보일까

이제 이유를 한 번에 정리하면 이렇습니다.

1. `connect`는 새 연결이 필요할 때만 의미가 있다.
2. `read/write timeout`은 낮은 계층에서 socket/connection 성격이 강하다.
3. keep-alive와 pool이 request와 connection을 분리했다.
4. HTTP/2와 HTTP/3는 하나의 연결 위에 여러 stream을 동시에 올린다.
5. shared client와 connection pool은 실무적으로 매우 중요하다.
6. 그래서 라이브러리는 transport timeout을 shared policy로, 요청별 요구는 deadline/cancel로 표현하는 쪽이 더 단순하고 안전하다.

한 줄로 요약하면:

> 요청별 전체 마감시간은 "이번 일의 SLA"에 가깝고, connect/read timeout은 "도로와 차량의 운행 정책"에 가깝습니다.

## 도로와 배송 비유로 다시 보면

이 비유가 꽤 잘 맞습니다.

- `client 설정`: 차량 종류, 기본 제한 속도, 연료 정책
- `connection pool`: 차고에서 어떤 차를 꺼낼지
- `connect timeout`: 새 차를 도로에 올려 출발시키는 데 기다릴 시간
- `socket read timeout`: 도로 위에서 다음 신호나 흐름을 기다리는 기본 규칙
- `request deadline`: 이번 배송은 5분 안에 도착해야 한다는 업무 마감

배송마다 "이번 배송은 5분 안"이라고 말하는 것은 자연스럽습니다. 하지만 배송마다 "이번 건은 도로의 기본 신호 대기 시간을 바꾸자"라고 하면, 같은 도로와 같은 차를 공유할 때 곧바로 충돌합니다.

그래서 보통은 이렇게 됩니다.

```text
차량/도로 기본 규칙 = client / connection 기본 설정
이번 배송 마감     = request deadline / cancellation
```

## 라이브러리별로 보면

아래 표는 대표적인 차이를 빠르게 보는 용도입니다.

| 라이브러리 | 기본 모형 | 요청별로 주기 쉬운 것 | 주의점 |
| --- | --- | --- | --- |
| JDK HttpClient | client connect timeout + request timeout | 전체 request timeout | connect timeout은 새 연결일 때만 의미가 큼 |
| Apache HttpClient | socket / connection / request 계층 분리 | request config | 계층을 명확히 구분해서 봐야 함 |
| Go `net/http` | `Client.Timeout` + `Transport` 세부 설정 | 전체 deadline, context | transport와 request context를 함께 봐야 함 |
| Requests | per-call `timeout` 인자 | 호출별 timeout 인자 | 전체 다운로드 deadline이 아니라 socket inactivity 성격 |
| aiohttp | session default + request override | total/connect/sock_connect/sock_read | request override가 가능해도 phase 분리는 유지 |
| libcurl | transfer handle 중심 | transfer 전체 / connect 단계 | handle이 정책 소유자라서 per-transfer가 자연스러움 |
| OkHttp | client builder 중심 | `callTimeout` 등 | shared client + 기본 정책 모델 |
| Fetch | timeout 파라미터보다 abort/cancel | `AbortSignal.timeout()` | 브라우저는 cancellation 쪽으로 기울어짐 |

## 실무에서는 어떻게 생각하면 덜 헷갈릴까

### 1. 먼저 내가 제한하려는 것이 무엇인지 말로 써 본다

- "새 연결이 너무 늦으면 끊고 싶다"
- "응답 전체가 2초 넘으면 실패시키고 싶다"
- "pool에서 연결 기다리다가 100ms 넘으면 실패시키고 싶다"
- "바디 스트리밍 중에 30초 넘게 아무 바이트도 안 오면 끊고 싶다"

이 문장을 먼저 만들면 적절한 timeout 종류가 더 잘 보입니다.

### 2. business SLA는 request deadline으로 표현하는 편이 안전하다

업무적으로 "이 API는 3초 안에 끝나야 한다"는 요구는 보통 request 전체 deadline이나 cancellation이 더 잘 맞습니다.

### 3. transport 정책은 shared client 기본값으로 두는 편이 안전하다

connect/read/write/socket 성격의 값은 pool, keep-alive, multiplexing과 충돌할 가능성이 커서 공유 정책으로 두는 편이 단순합니다.

### 4. 정말 요청별 phase override가 필요하면, 라이브러리의 실제 의미를 먼저 확인한다

예를 들어 `requests timeout=3`은 "전체 3초 deadline"이라고 오해하기 쉽지만, quickstart는 그렇게 설명하지 않습니다. [Requests Quickstart](https://requests.readthedocs.io/en/latest/user/quickstart/#timeouts)

즉 같은 이름이어도 의미가 다를 수 있습니다.

## 작은 의사코드로 마무리

### 단순하고 안전한 기본 모형

```text
client {
  connect_timeout = 1s
  socket_read_timeout = 3s
  keepalive_idle_timeout = 30s
}

request {
  deadline = now + 5s
}

runtime:
  if need_new_connection:
    apply connect_timeout

  send request
  enforce request.deadline in user space
  use shared connection/socket policy underneath
```

### 겉보기에는 편하지만 내부적으로 더 까다로운 모형

```text
request A { connect=100ms, read=100ms }
request B { connect=5s,    read=10s   }

shared HTTP/2 connection 1개

question:
  socket timeout은 누구 기준으로 둘까?
  이미 연결된 상태에서 request A의 connect timeout은 무슨 뜻일까?
```

이 질문에 자연스럽게 답하기 어려워지는 순간부터, 왜 많은 라이브러리가 요청별 phase timeout보다 client 기본 정책 + request deadline 쪽으로 기울었는지 이해할 수 있습니다.

## 직접 확인해 볼 수 있는 포인트

1. JDK 문서를 보면 `connectTimeout`은 `HttpClient.Builder`, `timeout`은 `HttpRequest.Builder`에 있습니다. [JDK HttpClient.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpClient.Builder.html) [JDK HttpRequest.Builder](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpRequest.Builder.html)
2. Linux man page를 보면 `SO_RCVTIMEO`, `SO_SNDTIMEO`는 socket option이고, `MSG_DONTWAIT`는 per-call option으로 따로 나옵니다. [socket(7)](https://man7.org/linux/man-pages/man7/socket.7.html) [recv(2)](https://man7.org/linux/man-pages/man2/recv.2.html)
3. Apache 문서를 보면 configuration level이 per-socket, per-connection, per-request로 분리되어 있습니다. [Apache Configuration Guide](https://hc.apache.org/httpcomponents-client-5.6.x/configuration.html)
4. aiohttp 문서를 보면 session default는 물론 request override도 있지만, 여전히 `total`, `connect`, `sock_connect`, `sock_read`를 나눕니다. [aiohttp Client Quickstart](https://docs.aiohttp.org/en/stable/client_quickstart.html#timeouts)

## 출처

- [RFC 7230: HTTP/1.1 Message Syntax and Routing](https://www.rfc-editor.org/rfc/rfc7230.html)
- [RFC 9113: HTTP/2](https://www.rfc-editor.org/rfc/rfc9113.html)
- [RFC 9000: QUIC: A UDP-Based Multiplexed and Secure Transport](https://www.rfc-editor.org/rfc/rfc9000.html)
- [Linux `socket(7)`](https://man7.org/linux/man-pages/man7/socket.7.html)
- [Linux `connect(2)`](https://man7.org/linux/man-pages/man2/connect.2.html)
- [Linux `recv(2)`](https://man7.org/linux/man-pages/man2/recv.2.html)
- [JDK `HttpClient.Builder`](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpClient.Builder.html)
- [JDK `HttpRequest.Builder`](https://docs.oracle.com/en/java/javase/24/docs/api/java.net.http/java/net/http/HttpRequest.Builder.html)
- [Apache HttpComponents Configuration Guide](https://hc.apache.org/httpcomponents-client-5.6.x/configuration.html)
- [Go `net/http` package](https://pkg.go.dev/net/http)
- [Requests Quickstart: Timeouts](https://requests.readthedocs.io/en/latest/user/quickstart/#timeouts)
- [aiohttp Client Quickstart: Timeouts](https://docs.aiohttp.org/en/stable/client_quickstart.html#timeouts)
- [libcurl `CURLOPT_TIMEOUT`](https://curl.se/libcurl/c/CURLOPT_TIMEOUT.html)
- [libcurl `CURLOPT_CONNECTTIMEOUT`](https://curl.se/libcurl/c/CURLOPT_CONNECTTIMEOUT.html)
- [OkHttp `OkHttpClient.Builder`](https://square.github.io/okhttp/3.x/okhttp/okhttp3/OkHttpClient.Builder.html)
- [DOM Standard: `AbortSignal.timeout()`](https://dom.spec.whatwg.org/#abortsignal-timeout)
- [Fetch Standard](https://fetch.spec.whatwg.org/)
