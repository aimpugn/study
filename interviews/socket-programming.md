# 소켓이란 무엇이고 소켓 프로그래밍이란 무엇인가

- [소켓이란 무엇이고 소켓 프로그래밍이란 무엇인가](#소켓이란-무엇이고-소켓-프로그래밍이란-무엇인가)
  - [면접에서 먼저 말할 답](#면접에서-먼저-말할-답)
  - [이 질문이 실제로 확인하려는 것](#이-질문이-실제로-확인하려는-것)
  - [운영체제 중립 모델과 OS별 API 차이](#운영체제-중립-모델과-os별-api-차이)
  - [첫 번째 벽돌: POSIX 계열에서 정수 하나가 네트워크 통로가 되는 장면](#첫-번째-벽돌-posix-계열에서-정수-하나가-네트워크-통로가-되는-장면)
  - [소켓은 포트도, 패킷도, HTTP 요청도 아니다](#소켓은-포트도-패킷도-http-요청도-아니다)
  - [서버 소켓의 생애: socket, bind, listen, accept](#서버-소켓의-생애-socket-bind-listen-accept)
  - [클라이언트 소켓의 생애: socket, connect, send, recv, close](#클라이언트-소켓의-생애-socket-connect-send-recv-close)
  - [데이터는 소켓을 지나 어떤 모양으로 움직이는가](#데이터는-소켓을-지나-어떤-모양으로-움직이는가)
  - [소켓 프로그래밍이란 무엇인가](#소켓-프로그래밍이란-무엇인가)
  - [Java 예제로 보는 소켓 프로그래밍](#java-예제로-보는-소켓-프로그래밍)
  - [동시성 모델: 연결마다 스레드인가, 이벤트 루프인가](#동시성-모델-연결마다-스레드인가-이벤트-루프인가)
  - [TCP 소켓과 UDP 소켓은 무엇이 다른가](#tcp-소켓과-udp-소켓은-무엇이-다른가)
  - [실무에서 자주 터지는 실패 모드](#실무에서-자주-터지는-실패-모드)
  - [직접 확인하는 작은 실험](#직접-확인하는-작은-실험)
  - [꼬리 질문 대비](#꼬리-질문-대비)
  - [정리](#정리)
  - [근거와 더 읽을 자료](#근거와-더-읽을-자료)

## 면접에서 먼저 말할 답

소켓(socket)은 애플리케이션이 운영체제의 네트워크 기능을 쓰기 위해 여는 통신 끝점입니다. 특정 운영체제에 묶지 않고 말하면, 소켓은 "내 프로그램이 TCP나 UDP 같은 전송 방식으로 상대와 데이터를 주고받기 위해 운영체제에 맡긴 통로"입니다. Unix/Linux/macOS 같은 POSIX 계열에서는 이 통로가 파일 디스크립터(file descriptor)라는 정수 손잡이로 보이는 경우가 많고, Windows에서는 Winsock의 `SOCKET` 핸들처럼 다른 형태로 보입니다. 겉모양은 달라도 운영체제의 네트워크 스택 안에서 주소, 프로토콜 상태, 송수신 버퍼, 오류와 종료 상태가 관리된다는 점은 같습니다.

소켓 프로그래밍은 이 소켓을 직접 만들고, 주소를 붙이고, 연결을 받거나 맺고, 데이터를 읽고 쓰고, 오류와 종료를 처리하는 프로그램을 작성하는 일입니다. 서버라면 보통 `socket -> bind -> listen -> accept -> read/write -> close` 흐름을 다루고, 클라이언트라면 `socket -> connect -> send/recv -> close` 흐름을 다룹니다.

면접에서 더 짧게 말하면 이렇게 답할 수 있습니다.

> 소켓은 애플리케이션과 커널 네트워크 스택 사이의 통신 끝점이고, 소켓 프로그래밍은 그 끝점을 열어 TCP나 UDP로 데이터를 주고받는 코드를 쓰는 일입니다. 서버에서는 `bind`, `listen`, `accept`로 연결을 받고, 클라이언트에서는 `connect`로 연결을 맺은 뒤 `send`/`recv` 또는 스트림 API로 바이트를 주고받습니다.

## 이 질문이 실제로 확인하려는 것

이 질문은 "소켓은 네트워크 통신에 쓰입니다"라는 정의를 외웠는지보다, 애플리케이션 코드와 운영체제 커널 사이의 경계를 이해하는지 확인합니다. 좋은 답변은 세 가지를 분리해야 합니다.

1. 주소와 이름

    IP 주소와 포트 번호는 네트워크에서 어느 호스트의 어느 서비스로 갈지 가리키는 이름입니다. `bind`는 서버 소켓에 로컬 주소를 붙이고, `connect`는 클라이언트 소켓을 상대 주소로 연결합니다.

2. 커널 객체와 손잡이

    소켓 자체는 커널이 관리하는 통신 객체입니다. 애플리케이션은 그 객체를 직접 만지는 것이 아니라 파일 디스크립터 또는 Java `Socket` 같은 런타임 객체를 통해 커널에 요청합니다.

3. 프로토콜 상태와 데이터 흐름

    TCP 소켓이라면 연결 상태, 순서 보장, 재전송, 흐름 제어, 송수신 버퍼가 함께 움직입니다. HTTP 요청이나 JSON 객체는 그 위에서 애플리케이션이 해석하는 의미이고, 커널이 직접 아는 단위는 주로 바이트와 TCP 상태입니다.

이 세 가지가 섞이면 흔한 오해가 생깁니다. 예를 들어 "`accept()`가 새 포트를 만든다", "소켓은 IP와 포트의 조합 그 자체다", "`send()` 한 번은 `recv()` 한 번으로 그대로 도착한다" 같은 설명은 실제 동작을 놓칩니다.

## 운영체제 중립 모델과 OS별 API 차이

소켓을 특정 운영체제에 묶지 않고 설명할 때는 함수 이름보다 먼저, 거의 모든 구현이 공유하는 공통 흐름을 잡는 편이 안전합니다.

```text
애플리케이션 코드
  |
  | "이 주소와 전송 방식으로 통신하고 싶다"
  v
운영체제 네트워크 API
  |
  | 통신 끝점을 가리키는 손잡이 반환
  v
운영체제 네트워크 스택
  |
  | 주소, 연결 상태, 송수신 버퍼, 오류/종료 상태 관리
  v
네트워크
```

운영체제마다 API 이름, 핸들 타입, 비동기 I/O 모델, 옵션 이름은 달라질 수 있습니다. 그래도 네트워크 프로그램이 풀어야 하는 질문은 대체로 같습니다.

1. 어떤 주소 체계와 전송 방식을 쓸 것인가

    IPv4, IPv6, Unix domain socket 같은 주소 체계와 TCP 스트림, UDP 데이터그램 같은 전송 방식을 고릅니다. POSIX의 `socket(domain, type, protocol)`이나 Windows Winsock의 `socket(af, type, protocol)` 모두 이 세 가지 선택을 받습니다.

2. 이 통신 끝점을 어떻게 가리킬 것인가

    POSIX 계열은 소켓을 파일 디스크립터로 반환합니다. Windows Winsock은 `SOCKET`이라는 별도 핸들 타입을 반환하고, 닫을 때도 일반 `close()`가 아니라 `closesocket()`을 씁니다. Apple 플랫폼에서는 BSD socket도 쓸 수 있지만, 앱 개발에서는 `NWConnection` 같은 Network framework 추상화를 더 높은 수준의 연결 객체로 쓰기도 합니다.

3. 서버 입구와 연결 하나를 어떻게 구분할 것인가

    이름은 달라도 서버에는 보통 "새 연결을 기다리는 입구"와 "특정 상대와 데이터를 주고받는 연결"이 나뉩니다. POSIX에서는 listening socket과 `accept()`가 반환한 connected socket으로 보이고, Java에서는 `ServerSocket`과 `Socket`으로 보이며, Go에서는 `net.Listener`와 `net.Conn`으로 보입니다.

4. 기다림을 어떻게 표현할 것인가

    호출한 스레드가 멈춰 기다리는 blocking call을 쓸 수도 있고, `select`/`poll`/`epoll`/`kqueue`, Windows IOCP, Apple Network framework의 callback/queue처럼 운영체제별 이벤트 모델을 쓸 수도 있습니다. 이 차이는 구현 방식의 차이이지, "소켓은 운영체제 네트워크 스택에 등록된 통신 끝점"이라는 핵심 모델을 바꾸지는 않습니다.

면접에서는 먼저 운영체제 중립 모델로 답하고, 그다음 대표 구현을 붙이면 좋습니다.

> 개념적으로 소켓은 운영체제 네트워크 스택에 만든 통신 끝점입니다. Unix 계열에서는 그 끝점을 파일 디스크립터로 다루고, Windows에서는 Winsock `SOCKET` 핸들로 다루며, Java나 Go 같은 언어는 다시 `Socket`, `ServerSocket`, `Listener`, `Conn` 같은 객체로 감쌉니다. API 모양은 달라도 주소를 붙이고, 연결을 맺거나 받고, 바이트나 데이터그램을 주고받고, 오류와 종료를 처리한다는 생명주기는 거의 같습니다.

## 첫 번째 벽돌: POSIX 계열에서 정수 하나가 네트워크 통로가 되는 장면

가장 작은 대표 장면은 POSIX/BSD 계열 C 소켓 API의 `socket()` 호출입니다. 이 예시는 Linux, macOS, BSD 계열의 공통 감각을 잡기 위한 것이고, 모든 OS가 소켓을 반드시 "작은 정수 fd"로 표현한다는 뜻은 아닙니다.

```c
int fd = socket(AF_INET, SOCK_STREAM, 0);
```

이 한 줄은 "IPv4 주소 체계(`AF_INET`)에서 TCP 스트림(`SOCK_STREAM`) 통신에 쓸 커널 소켓 객체를 만들고, 그 객체를 가리키는 파일 디스크립터 `fd`를 달라"는 요청입니다. Linux `socket(2)` 문서는 `socket()`이 통신 끝점을 만들고 그 끝점을 가리키는 파일 디스크립터를 반환한다고 설명합니다.

아래 그림은 이 한 줄 뒤에 생기는 관계를 단순화한 것입니다.

```text
user space process
┌──────────────────────────────────────────────┐
│ fd table                                     │
│                                              │
│   0 -> stdin                                 │
│   1 -> stdout                                │
│   2 -> stderr                                │
│   3 -> 통신 끝점(socket endpoint) ───────┐    │
└─────────────────────────────────────────│────┘
                                          │ syscall boundary
kernel space                              │
┌─────────────────────────────────────────▼────┐
│ socket object                                │
│                                              │
│ address family: AF_INET                      │
│ type: SOCK_STREAM                            │
│ protocol state: TCP                          │
│ send buffer / receive buffer                 │
│ local address: not bound yet                 │
│ remote address: not connected yet            │
└──────────────────────────────────────────────┘
```

여기서 `fd=3`은 소켓 자체가 아니라 손잡이입니다. 애플리케이션은 `fd=3`을 `bind`, `listen`, `accept`, `connect`, `send`, `recv`, `close` 같은 시스템 호출에 넘깁니다. 커널은 그 숫자를 보고 프로세스의 파일 디스크립터 테이블에서 실제 소켓 객체를 찾아 처리합니다.

Java에서는 숫자 `fd`가 직접 보이지 않을 뿐 같은 구조 위에 올라탑니다. `new Socket(...)`이나 `new ServerSocket(...)`은 JDK 객체를 만들지만, 실제 네트워크 연결과 대기 상태는 운영체제의 소켓과 파일 디스크립터가 담당합니다. Oracle의 `ServerSocket` 문서도 서버 소켓이 네트워크 요청을 기다리는 역할을 한다고 설명합니다.

## 소켓은 포트도, 패킷도, HTTP 요청도 아니다

면접에서 소켓 답변이 흔들리는 이유는 서로 다른 층위의 단어를 한 덩어리로 말하기 때문입니다.

| 개념 | 한 문장 설명 | 소켓과의 관계 |
| --- | --- | --- |
| IP 주소 | 네트워크에서 호스트나 인터페이스를 찾는 주소입니다. | 소켓이 통신할 로컬/원격 주소의 일부입니다. |
| 포트 | 한 호스트 안에서 어떤 전송 계층 서비스로 보낼지 구분하는 번호입니다. | 서버 소켓은 보통 특정 로컬 포트에 `bind` 됩니다. |
| 파일 디스크립터 | 프로세스가 열린 커널 객체를 가리키는 정수 손잡이입니다. | 유닉스 계열에서 소켓은 파일 디스크립터로 접근됩니다. |
| 소켓 | 커널이 관리하는 통신 끝점입니다. | 주소, 프로토콜 상태, 버퍼를 가진 실제 I/O 대상입니다. |
| TCP 연결 | 두 끝점 사이에 만들어진 신뢰성 있는 바이트 스트림 상태입니다. | TCP 연결이 성립하면 양쪽에 연결 소켓이 있습니다. |
| HTTP 요청 | TCP 바이트 스트림 위에서 애플리케이션이 해석하는 메시지입니다. | 커널 소켓은 HTTP 의미를 모르고 바이트만 전달합니다. |

특히 포트와 소켓을 구분해야 합니다. 포트는 주소의 일부이고, 소켓은 그 주소를 붙잡고 커널 안에서 I/O 상태를 관리하는 객체입니다. 서버가 8080 포트에서 대기한다고 해서 "포트 하나가 연결 하나"라는 뜻은 아닙니다. 같은 서버 포트 8080에 여러 클라이언트가 동시에 연결될 수 있고, 커널은 보통 아래 4개 값으로 TCP 연결을 구분합니다.

```text
(client IP, client port, server IP, server port)
```

이 네 값의 조합을 흔히 4-tuple이라고 부릅니다. 한국어로는 "네 가지 주소 요소의 조합"이라고 풀어 말하면 충분합니다. 여기에 프로토콜까지 포함해 TCP/UDP 구분을 명시할 때는 5-tuple이라고 부르기도 합니다.

## 서버 소켓의 생애: socket, bind, listen, accept

서버 소켓을 이해할 때는 "연결을 기다리는 소켓"과 "데이터를 주고받는 소켓"을 나누어야 합니다.

```text
server process

1. socket()
   fd=3
   아직 주소 없음, 아직 연결 없음

2. bind(fd=3, 0.0.0.0:8080)
   fd=3
   local address = 0.0.0.0:8080

3. listen(fd=3, backlog=128)
   fd=3
   state = LISTEN
   새 연결을 받을 큐 준비

4. accept(fd=3)
   fd=11
   state = ESTABLISHED
   이 fd로 실제 read/write 수행
```

`bind`는 소켓에 로컬 주소를 붙이는 단계입니다. Linux `bind(2)` 문서는 주소가 없는 소켓에 주소를 할당하는 동작이라고 설명합니다. `listen`은 그 소켓을 수동 소켓(passive socket), 즉 들어오는 연결 요청을 받을 소켓으로 표시하고 대기 큐의 한계를 정합니다. `accept`는 대기 중인 완성 연결 하나를 꺼내 새 연결 소켓을 만들고, 그 새 소켓을 가리키는 파일 디스크립터를 반환합니다.

이때 가장 중요한 그림은 아래입니다.

```text
LISTEN socket
  fd=3
  local=0.0.0.0:8080
  remote=*:*
  state=LISTEN

CONNECTED sockets returned by accept()
  fd=11  local=10.0.0.5:8080  remote=203.0.113.10:53144  state=ESTABLISHED
  fd=12  local=10.0.0.5:8080  remote=203.0.113.11:53145  state=ESTABLISHED
  fd=13  local=10.0.0.5:8080  remote=203.0.113.12:53146  state=ESTABLISHED
```

`accept()`는 원래 리스닝 소켓을 없애지 않습니다. `fd=3`은 계속 8080 포트에서 새 연결을 기다리고, `accept()`가 반환한 `fd=11`, `fd=12`, `fd=13` 같은 연결 소켓이 각 클라이언트와 데이터를 주고받습니다. Linux `accept(2)` 문서도 새 connected socket은 listening 상태가 아니고, 원래 listening socket은 영향을 받지 않는다고 설명합니다.

TCP 핸드셰이크와 `accept()`의 관계도 자주 묻습니다.

```text
client                                 server kernel / server process
  |                                                |
  | 1. SYN --------------------------------------->| LISTEN socket sees SYN
  |                                                | SYN queue
  | 2. <------------------------------- SYN + ACK  |
  |                                                |
  | 3. ACK --------------------------------------->| accept queue
  |                                                |
  |                         accept(listen_fd) ---->| dequeue one completed connection
  |                         returns conn_fd <------| new connected socket fd
  |                                                |
```

단순화해서 말하면, TCP 연결 완성 자체는 커널 TCP 스택이 처리하고, 서버 프로세스는 `accept()`로 완성된 연결을 사용자 공간으로 꺼내 옵니다. 그래서 "`accept()`가 핸드셰이크를 수행한다"보다 "`accept()`는 핸드셰이크가 끝나 대기 큐에 올라온 연결을 꺼내 연결 소켓 fd를 받는다"가 더 정확합니다.

## 클라이언트 소켓의 생애: socket, connect, send, recv, close

클라이언트는 보통 먼저 상대 주소를 알고 연결을 시도합니다.

```text
client process

1. socket()
   fd=4
   아직 로컬 포트가 명시되지 않았을 수 있음

2. connect(fd=4, 198.51.100.20:443)
   커널이 필요하면 임시 로컬 포트(ephemeral port)를 선택
   TCP 3-way handshake 진행

3. send(fd=4, request bytes)
   애플리케이션 바이트를 커널 send buffer에 넘김

4. recv(fd=4, response bytes)
   커널 receive buffer에서 애플리케이션으로 바이트 복사

5. close(fd=4)
   연결 종료 절차 또는 리소스 정리
```

`connect`는 소켓을 지정된 상대 주소에 연결하는 시스템 호출입니다. TCP라면 이 단계에서 3-way handshake가 진행되고, 성공하면 소켓은 연결된 바이트 스트림으로 동작합니다. 클라이언트가 `bind`를 직접 호출하지 않아도 커널은 보통 연결 시도에 필요한 로컬 IP와 임시 포트를 고릅니다.

서버 입장에서 동시에 많은 클라이언트가 붙어도 같은 서버 포트를 계속 쓸 수 있는 이유가 여기서 나옵니다. 클라이언트마다 임시 로컬 포트가 달라지고, 커널은 네 가지 주소 요소의 조합으로 연결을 구분합니다.

```text
client A  203.0.113.10:53144  ->  server 10.0.0.5:8080
client B  203.0.113.10:53145  ->  server 10.0.0.5:8080
client C  203.0.113.11:61200  ->  server 10.0.0.5:8080
```

서버 로컬 포트는 모두 8080입니다. 새 포트로 연결을 옮긴 것이 아니라, 서로 다른 원격 주소와 원격 포트가 붙은 별도 연결 소켓들이 같은 로컬 서비스 포트를 공유하는 것입니다.

## 데이터는 소켓을 지나 어떤 모양으로 움직이는가

TCP 소켓에서 가장 중요한 성질은 "메시지"가 아니라 "바이트 스트림"이라는 점입니다. TCP는 바이트 순서를 보장하지만, 애플리케이션이 한 번 `send()`한 묶음이 상대방의 한 번 `recv()`로 그대로 도착한다고 보장하지 않습니다.

아래 예시는 클라이언트가 `HELLO`와 `WORLD`를 따로 보낸 상황입니다.

```text
application write calls
  send("HELLO")
  send("WORLD")

TCP stream seen by receiver
  H E L L O W O R L D

possible recv results
  recv() -> "HEL"
  recv() -> "LOWOR"
  recv() -> "LD"

or
  recv() -> "HELLOWORLD"
```

따라서 소켓 프로그래밍에서 "어디까지가 한 메시지인가"는 애플리케이션 프로토콜이 정해야 합니다. HTTP/1.1은 헤더 끝을 `CRLF CRLF`로 구분하고, 본문 길이는 `Content-Length`나 chunked encoding 같은 규칙으로 정합니다. 직접 프로토콜을 만든다면 보통 아래 중 하나를 둡니다.

1. 고정 길이 프레임

    항상 16바이트, 128바이트처럼 정해진 크기만 읽습니다. 단순하지만 가변 길이 메시지에는 낭비가 큽니다.

2. 길이 prefix

    먼저 4바이트 길이를 읽고, 그 길이만큼 본문을 다시 읽습니다. 바이너리 프로토콜에서 흔합니다.

3. 구분자 delimiter

    줄바꿈 `\n`처럼 특정 구분자가 나올 때까지 읽습니다. 텍스트 프로토콜에서 이해하기 쉽지만, 본문 안에 구분자가 들어갈 수 있는지 규칙을 정해야 합니다.

길이 prefix 방식은 아래처럼 움직입니다.

```text
sender application message
  "ping"

wire bytes
  00 00 00 04 70 69 6e 67
  └─ length=4 ┘ └ p  i  n  g ┘

receiver loop
  1. read exactly 4 bytes      -> length = 4
  2. read exactly length bytes -> payload = "ping"
  3. hand payload to protocol parser
```

여기서 `read exactly`가 중요합니다. `recv(fd, buf, 4)`를 한 번 호출했다고 항상 4바이트가 온다고 가정하면 안 됩니다. 필요한 길이가 채워질 때까지 반복해서 읽고, 중간에 0이 오면 상대가 정상 종료한 것으로 처리하며, 오류가 오면 어떤 오류인지에 따라 재시도하거나 연결을 닫아야 합니다.

## 소켓 프로그래밍이란 무엇인가

소켓 프로그래밍은 단지 `Socket` 클래스를 쓰는 일이 아니라, 네트워크 I/O의 생명주기와 실패 조건을 코드로 관리하는 일입니다. 최소한 아래 책임이 들어갑니다.

- 어떤 주소 체계와 전송 방식을 쓸지 고릅니다. 예를 들어 IPv4/IPv6, TCP 스트림, UDP 데이터그램, 유닉스 도메인 소켓 중 무엇을 쓸지 정합니다.
- 서버라면 로컬 주소에 이름을 붙이고(`bind`), 연결 요청을 받을 준비를 하며(`listen`), 완성된 연결을 꺼냅니다(`accept`).
- 클라이언트라면 상대 주소로 연결을 시도합니다(`connect`).
- 데이터 경계를 정합니다. TCP는 바이트 스트림이므로 애플리케이션 프로토콜이 메시지 경계를 만들어야 합니다.
- 부분 읽기, 부분 쓰기, timeout, 연결 종료, reset, interrupted call, backpressure를 처리합니다.
- 동시에 여러 연결을 처리할 모델을 정합니다. 연결마다 스레드를 둘지, non-blocking 소켓과 이벤트 루프를 쓸지, 운영체제 이벤트 API를 쓸지 정합니다.
- 운영 한계를 관리합니다. 파일 디스크립터 수, backlog, 소켓 버퍼, 메모리, 스레드 풀, 커넥션 풀, idle timeout이 모두 실제 장애 지점이 될 수 있습니다.

언어별 API는 달라도 밑바닥 질문은 비슷합니다.

```text
Java ServerSocket.accept()
Python socket.accept()
Go net.Listener.Accept()
C accept()

모두 결국 "완성된 연결 하나를 받아서, 그 연결로 읽고 쓸 수 있는 핸들"을 얻는 추상화다.
```

## Java 예제로 보는 소켓 프로그래밍

Java에서는 `ServerSocket`과 `Socket`이 C의 raw 파일 디스크립터를 직접 보이지 않게 감쌉니다. 하지만 논리는 같습니다. `ServerSocket`은 들어오는 연결을 기다리는 서버 쪽 소켓이고, `accept()`가 반환하는 `Socket`은 특정 클라이언트와 연결된 소켓입니다.

아래 코드는 줄 단위 echo 서버입니다. 학습용으로 단순화했기 때문에 운영 서버처럼 timeout, thread pool, graceful shutdown을 모두 갖춘 코드는 아닙니다.

```java
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.ServerSocket;
import java.net.Socket;
import java.nio.charset.StandardCharsets;

public class LineEchoServer {
    public static void main(String[] args) throws Exception {
        try (ServerSocket server = new ServerSocket(18080)) {
            while (true) {
                try (Socket client = server.accept();
                     var in = new BufferedReader(new InputStreamReader(
                         client.getInputStream(), StandardCharsets.UTF_8));
                     var out = new PrintWriter(client.getOutputStream(), true, StandardCharsets.UTF_8)) {

                    String line = in.readLine();
                    if (line != null) {
                        out.println("echo: " + line);
                    }
                }
            }
        }
    }
}
```

이 코드에서 각 줄의 의미를 운영체제 흐름으로 바꾸면 아래와 같습니다.

```text
new ServerSocket(18080)
  -> socket()
  -> bind(0.0.0.0 or configured address, 18080)
  -> listen(backlog)

server.accept()
  -> accept(listening socket fd)
  -> connected socket fd 반환
  -> Java Socket 객체가 그 연결 fd를 감쌈

client.getInputStream().readLine()
  -> connected socket에서 바이트 읽기
  -> Java reader가 줄바꿈까지 문자로 해석

out.println(...)
  -> 문자열을 바이트로 인코딩
  -> connected socket으로 write/send

client.close()
  -> 연결 소켓 fd 정리
```

여기서 `ServerSocket`과 `Socket`을 섞으면 안 됩니다. `ServerSocket`은 새 연결을 기다리는 입구이고, `accept()`가 반환한 `Socket`은 특정 클라이언트와 데이터를 주고받는 통로입니다. 서버가 계속 여러 클라이언트를 받으려면 `ServerSocket`은 열어 두고, 각 클라이언트용 `Socket`만 요청 처리가 끝난 뒤 닫아야 합니다.

## 동시성 모델: 연결마다 스레드인가, 이벤트 루프인가

소켓 프로그래밍은 연결 하나만 다룰 때보다 여러 연결을 동시에 다룰 때 더 어려워집니다. 서버가 연결 1개만 처리한다면 `accept -> read -> write -> close`를 반복해도 됩니다. 하지만 한 클라이언트의 `read`가 오래 막히면 다음 클라이언트의 `accept`나 처리까지 늦어질 수 있습니다.

대표 모델은 두 가지입니다.

1. 연결마다 스레드 또는 작업을 배정하는 모델

    `accept()`가 연결을 하나 반환할 때마다 새 스레드나 스레드 풀 작업에 넘깁니다. 코드는 직관적이고 blocking I/O로 작성하기 쉽습니다. 대신 연결 수가 많아지면 스레드 수, 스택 메모리, 컨텍스트 스위칭 비용, blocking 대기 시간이 문제가 됩니다.

2. non-blocking 소켓과 이벤트 루프 모델

    소켓을 non-blocking으로 설정하고, `select`, `poll`, `epoll`, `kqueue` 같은 이벤트 통지 API로 "지금 읽거나 쓸 준비가 된 fd"만 처리합니다. Nginx, Netty, Node.js 같은 시스템을 이해할 때 중요한 모델입니다. 코드는 더 복잡하지만, 많은 idle 연결을 적은 스레드로 관리하기 좋습니다.

두 모델의 차이는 아래처럼 볼 수 있습니다.

```text
thread-per-connection

client A fd=11 -> thread A -> blocking read
client B fd=12 -> thread B -> blocking read
client C fd=13 -> thread C -> blocking read

event-loop

fd=11 ┐
fd=12 ├── epoll/kqueue/select -> ready fd list -> one loop handles small steps
fd=13 ┘
```

이벤트 루프는 HTTP/2의 멀티플렉싱과도 구분해야 합니다. `epoll`이나 `kqueue`는 여러 파일 디스크립터의 준비 상태를 감시하는 운영체제 API입니다. HTTP/2 멀티플렉싱은 하나의 TCP 연결 안에서 여러 HTTP 스트림을 프레임으로 섞어 보내는 애플리케이션 프로토콜 규칙입니다. 둘 다 "여러 개를 동시에 다룬다"는 느낌이 있지만, 층위가 다릅니다.

## TCP 소켓과 UDP 소켓은 무엇이 다른가

소켓은 TCP에만 있는 개념이 아닙니다. UDP도 소켓을 사용합니다. 다만 소켓의 타입과 프로토콜 계약이 다릅니다.

| 구분 | TCP 소켓 | UDP 소켓 |
| --- | --- | --- |
| 대표 타입 | `SOCK_STREAM` | `SOCK_DGRAM` |
| 연결 개념 | 연결 상태가 있습니다. | 기본적으로 연결 없는 데이터그램입니다. |
| 데이터 모양 | 순서 있는 바이트 스트림입니다. | 보낸 단위인 데이터그램 경계가 유지됩니다. |
| 신뢰성 | 재전송, 순서 보장, 흐름 제어를 제공합니다. | 애플리케이션이 손실, 순서, 재전송을 직접 다뤄야 할 수 있습니다. |
| 대표 사용 | HTTP/1.1, HTTP/2 over TCP, DB 연결, SSH | DNS, 일부 게임/실시간 통신, QUIC의 기반 전송 |

TCP 소켓에서 "연결"은 커널이 양 끝의 상태를 오래 유지한다는 뜻입니다. UDP 소켓은 상대 주소로 데이터그램을 보낼 수 있지만, TCP처럼 양쪽이 3-way handshake로 연결 상태를 만든 뒤 바이트 스트림을 주고받는 구조는 아닙니다. 그래서 UDP 서버에는 `listen()`과 `accept()` 흐름이 없습니다. 보통 `recvfrom()`으로 데이터그램과 보낸 주소를 함께 받습니다.

## 실무에서 자주 터지는 실패 모드

소켓 질문은 장애 질문으로 이어지기 쉽습니다. 대표 실패를 증상 중심으로 정리하면 아래와 같습니다.

| 증상 | 낮은 계층에서 일어나는 일 | 확인 관점 |
| --- | --- | --- |
| `Address already in use` | 같은 주소와 포트에 이미 리스닝 소켓이 있거나, 재사용 조건이 맞지 않습니다. | `lsof`, `ss`, `netstat`로 누가 해당 포트를 잡았는지 봅니다. |
| `Connection refused` | 목적지에 도달했지만 해당 포트에서 받아 줄 리스닝 소켓이 없거나 거부 응답이 왔습니다. | 서버 프로세스, bind 주소, 방화벽, 컨테이너 포트 매핑을 봅니다. |
| connect timeout | SYN에 대한 응답이 오지 않거나 경로 중간에서 사라집니다. | 라우팅, 보안 그룹, 방화벽, 서버 다운, 패킷 드롭을 봅니다. |
| read timeout | 연결은 됐지만 지정 시간 안에 다음 바이트가 오지 않습니다. | 서버 처리 지연, DB 대기, 프록시 timeout, backpressure를 봅니다. |
| `Connection reset` | 상대나 중간 장비가 TCP RST로 연결을 강제 종료했습니다. | 누가 RST를 보냈는지 패킷, LB 로그, 서버 로그를 같은 시간대에 맞춥니다. |
| `Broken pipe` | 이미 닫히거나 깨진 연결에 쓰려고 했습니다. | 상대 종료 시점, write 시점, retry/close 처리를 봅니다. |
| `Too many open files` | 프로세스가 새 파일 디스크립터를 더 열 수 없습니다. | `ulimit`, systemd limit, fd 누수, 연결 수, accept 실패를 봅니다. |
| accept queue overflow | 완성 연결을 애플리케이션이 충분히 빨리 `accept()`하지 못합니다. | backlog, accept thread, CPU saturation, SYN/accept queue 지표를 봅니다. |

여기서 중요한 점은 error message 하나가 곧 root cause는 아니라는 것입니다. 예를 들어 `Connection reset`은 Java 예외 이름이 아니라 TCP 사건이 애플리케이션에 표면화된 것입니다. 패킷을 보면 서버가 보냈을 수도 있고, 로드밸런서가 보냈을 수도 있고, 클라이언트 쪽 종료가 서버 로그에 반대로 찍힌 것일 수도 있습니다.

## 직접 확인하는 작은 실험

아래 실험은 localhost에서 리스닝 소켓과 연결 소켓을 눈으로 확인하기 위한 최소 예시입니다. macOS에서는 `lsof`가 기본 관측 도구로 편하고, Linux에서는 `ss`가 더 흔합니다.

먼저 Python 표준 라이브러리로 간단한 TCP 서버를 띄웁니다.

```sh
python3 - <<'PY'
import socket

server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
server.bind(("127.0.0.1", 18080))
server.listen(8)
print("LISTEN 127.0.0.1:18080")

conn, addr = server.accept()
print("ACCEPT", addr)
data = conn.recv(1024)
conn.sendall(b"echo: " + data)
conn.close()
server.close()
PY
```

다른 터미널에서 클라이언트를 실행합니다.

```sh
python3 - <<'PY'
import socket

with socket.create_connection(("127.0.0.1", 18080)) as s:
    s.sendall(b"hello\n")
    print(s.recv(1024).decode(), end="")
PY
```

기대 결과는 클라이언트 쪽에 `echo: hello`가 출력되고, 서버 쪽에는 `ACCEPT ('127.0.0.1', <client-port>)` 형태의 로그가 찍히는 것입니다. `<client-port>`는 클라이언트가 받은 임시 포트입니다.

서버를 띄워 둔 동안 관측 명령을 실행하면 리스닝 소켓을 볼 수 있습니다.

```sh
# macOS
lsof -nP -iTCP:18080

# Linux
ss -ltnp 'sport = :18080'
```

PASS 신호는 `127.0.0.1:18080` 또는 `*:18080`이 LISTEN 상태로 보이는 것입니다. FAIL 신호는 아무 것도 안 보이거나, 다른 프로세스가 이미 같은 포트를 점유하고 있거나, 클라이언트가 `Connection refused`를 보는 것입니다.

이 실험에서 꼭 확인할 것은 두 가지입니다.

1. 서버는 하나의 리스닝 소켓으로 `127.0.0.1:18080`을 잡고 있습니다.
2. 클라이언트가 연결되면 서버는 새 연결을 `accept()`하지만, 서버 로컬 포트가 새 번호로 바뀌지는 않습니다.

## 꼬리 질문 대비

### 소켓은 파일인가요

소켓은 일반 파일은 아닙니다. 다만 유닉스 계열에서는 파일 디스크립터라는 같은 손잡이로 접근되기 때문에 `read`, `write`, `close` 같은 유사한 인터페이스를 사용할 수 있습니다. 내부 대상은 디스크 파일의 inode가 아니라 커널의 소켓 객체와 프로토콜 상태입니다.

### 서버에서 `accept()`하면 새 포트가 생기나요

보통 새 포트가 생기지 않습니다. 새로 생기는 것은 연결 소켓과 그 소켓을 가리키는 파일 디스크립터입니다. 연결 소켓의 로컬 포트는 여전히 서버가 listen 중인 포트일 수 있고, 커널은 클라이언트 IP/포트까지 포함한 주소 조합으로 각 연결을 구분합니다.

### TCP 소켓에서 한 번 보낸 데이터는 한 번에 읽히나요

아닙니다. TCP는 바이트 스트림입니다. 송신자가 두 번 보낸 데이터가 수신자에게 한 번에 합쳐져 보일 수도 있고, 한 번 보낸 데이터가 여러 번에 나뉘어 보일 수도 있습니다. 그래서 애플리케이션 프로토콜은 길이 prefix, 구분자, 고정 길이 같은 프레이밍 규칙을 가져야 합니다.

### `ServerSocket`과 `Socket`은 무엇이 다른가요

Java 기준으로 `ServerSocket`은 들어오는 연결을 기다리는 서버 쪽 입구이고, `accept()`가 반환한 `Socket`은 특정 클라이언트와 연결된 통로입니다. 데이터를 실제로 읽고 쓰는 대상은 보통 `accept()`로 얻은 `Socket`입니다.

### HTTP 요청 하나가 소켓 하나인가요

항상 그렇지 않습니다. HTTP/1.0처럼 요청마다 연결을 닫는 단순 모델에서는 비슷하게 보일 수 있지만, HTTP keep-alive에서는 하나의 TCP 연결에서 여러 HTTP 요청/응답이 순서대로 오갈 수 있습니다. HTTP/2에서는 하나의 TCP 연결 안에서 여러 HTTP 스트림이 프레임 단위로 섞일 수도 있습니다. 커널이 아는 것은 TCP 연결과 바이트이고, HTTP 요청 경계는 애플리케이션 프로토콜이 해석합니다.

### 소켓 프로그래밍을 알면 왜 백엔드 개발에 도움이 되나요

타임아웃, 커넥션 풀, keep-alive, reverse proxy, `Connection reset`, `Broken pipe`, `Too many open files`, 대량 연결 처리 같은 문제는 모두 소켓 위에서 표면화됩니다. 프레임워크가 대부분을 감싸 주더라도, 장애가 나면 결국 어느 계층의 소켓 상태가 깨졌는지, 어떤 큐와 버퍼가 막혔는지, 어떤 timeout이 먼저 닫았는지 봐야 합니다.

## 정리

소켓은 애플리케이션이 네트워크를 쓰기 위해 커널에 여는 통신 끝점입니다. 서버는 리스닝 소켓으로 연결을 기다리고, `accept()`로 각 클라이언트에 대한 연결 소켓을 받아 실제 데이터를 읽고 씁니다. 클라이언트는 상대 주소로 `connect()`해 연결 소켓을 만들고, 그 위에서 바이트를 주고받습니다.

소켓 프로그래밍은 이 생명주기를 코드로 다루는 일입니다. 핵심은 API 이름을 외우는 것이 아니라, 포트와 소켓, 리스닝 소켓과 연결 소켓, TCP 바이트 스트림과 애플리케이션 메시지, blocking 모델과 이벤트 루프 모델을 구분하는 것입니다. 이 구분이 서면 면접 답변도 강해지고, 운영 장애를 볼 때도 "어느 계층의 어떤 상태를 확인해야 하는가"가 분명해집니다.

## 근거와 더 읽을 자료

- Linux man-pages: [`socket(2)`](https://www.man7.org/linux/man-pages/man2/socket.2.html), [`bind(2)`](https://man7.org/linux/man-pages/man2/bind.2.html), [`listen(2)`](https://man7.org/linux/man-pages/man2/listen.2.html), [`accept(2)`](https://man7.org/linux/man-pages/man2/accept.2.html), [`connect(2)`](https://man7.org/linux/man-pages/man2/connect.2.html), [`socket(7)`](https://man7.org/linux/man-pages/man7/socket.7.html)
- POSIX / The Open Group: [`socket()`](https://pubs.opengroup.org/onlinepubs/9699919799/functions/socket.html)
- Microsoft Learn: [Winsock `socket()`](https://learn.microsoft.com/en-us/windows/win32/api/winsock2/nf-winsock2-socket), [Winsock functions](https://learn.microsoft.com/en-us/windows/win32/api/winsock/)
- Apple Developer Documentation: [Network framework](https://developer.apple.com/documentation/Network)
- IETF RFC Editor: [RFC 9293 - Transmission Control Protocol (TCP)](https://www.rfc-editor.org/rfc/rfc9293)
- IETF: [RFC 9110 - HTTP Semantics](https://www.ietf.org/rfc/rfc9110.html)
- Oracle Java API: [`ServerSocket`](https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/net/ServerSocket.html), [`Socket`](https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/net/Socket.html)
- 함께 보면 좋은 저장소 문서: [linux-network-backend-runtime.md](linux-network-backend-runtime.md), [network-web-protocols.md](network-web-protocols.md), [../jvm/java/java_socket_connection_reset.md](../jvm/java/java_socket_connection_reset.md)
