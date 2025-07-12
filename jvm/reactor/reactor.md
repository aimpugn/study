# Reactor

- [Reactor](#reactor)
    - [Reactor](#reactor-1)
    - [Reactive Streams, Reactor Core, Netty, 그리고 Reactor Netty](#reactive-streams-reactor-core-netty-그리고-reactor-netty)
        - [Reactive Streams](#reactive-streams)
        - [Reactor Core(Project Reactor)](#reactor-coreproject-reactor)
        - [Netty](#netty)
        - [Reactor Netty](#reactor-netty)
    - [Inbound 및 Outbound 데이터 처리 흐름](#inbound-및-outbound-데이터-처리-흐름)
    - [예제](#예제)
        - [`WebClient` 설정](#webclient-설정)
        - [기본적인 사용법](#기본적인-사용법)

## Reactor

리액터는 "반응하는 자(Reactor)"에서 유래합니다.
외부 이벤트에 반응하여 적절한 처리를 수행하는 구조를 의미합니다.

- 이벤트 기반 처리: 하나 또는 여러 입력을 동시에 처리하여 서비스에 전달
- IO 멀티플렉싱: 하나의 블로킹 객체에서 여러 연결을 동시에 관리하며, 새로운 데이터 처리 시에만 블로킹 상태에서 반환
- 스레드 풀 기반 리소스 재사용: 각 연결마다 전용 스레드를 생성하지 않고 스레드 풀의 스레드를 재사용하여 여러 연결의 비즈니스 로직을 처리

가령 Reactor Pattern은 동시에 들어오는 여러 종류의 이벤트를 처리하기 위한 동시성 디자인 패턴입니다.
Reactor Pattern의 구성요소는 다음과 같습니다:

- Reactor: I/O 이벤트를 감지하고 분배하는 이벤트 루프
- Acceptor: 클라이언트 연결 접수를 담당
- Handler: 읽기, 쓰기 등 구체적인 I/O 이벤트 처리를 담당

Reactor 모델은 세 가지 형태로 발전했습니다:

- Single Reactor Single Thread: 모든 작업을 하나의 스레드에서 처리
- Single Reactor Multi Thread: 이벤트 감지는 단일 스레드, 처리는 멀티 스레드
- Master-Slave Reactor Multi Thread: 연결 수립과 I/O 처리를 분리

## Reactive Streams, Reactor Core, Netty, 그리고 Reactor Netty

대략적인 의존성 방향은 다음과 같습니다.

```sh
상위
[Spring WebFlux WebClient]
          │ Spring에서 활용
[TcpClient, HttpClient]
          │ 제공하는 클라이언트(Reactive + Network)
[Reactor Netty]
          │ + Netty(네트워크 I/O) 결합
          │   TcpClient/TcpServer, HttpClient/HttpServer 등 네트워크 어댑터
[Reactor Core (Project Reactor)]
          │ + Flux, Mono, Scheduler 등 구현
[Reactive Streams 표준]
          │
[Java NIO/OS Epoll, KQueue 등 ]
하위
```

- `TcpClient`:
    Reactor Netty에서 제공하는 TCP 전용 클라이언트입니다.
    Reactive Streams의 백프레셔를 지원합니다.

    ```java
    TcpClient.create().host().port().connectNow()
    ```

- `HttpClient`:
    Reactor Netty에서 제공하는 HTTP 전용 클라이언트입니다.

    ```java
    HttpClient.create().get().uri().response()
    ```

- `WebClient`:
    Spring WebFlux에서 제공하는 HTTP 클라이언트입니다.
    Reactor Netty `HttpClient` 기본 구현체를 사용합니다.

### Reactive Streams

- JVM에서 비동기 스트림 처리를 위한 표준 인터페이스입니다.
- `Publisher`, `Subscriber`, `Subscription`, `Processor` 등의 인터페이스를 정의합니다.
- 백프레셔(backpressure) 지원을 위한 표준 스펙을 제공합니다.

### Reactor Core(Project Reactor)

JVM에서 Non-blocking 리액티브 프로그래밍을 위한 라이브러리로, Reactive Streams 인터페이스의 구현체입니다.
`Flux`(0~N개 원소), `Mono`(0~1개 원소) 등 리액티브 프로그래밍 모델을 제공합니다.

- `Flux`: 0~N 요소를 스트림합니다. `map‧flatMap‧concat` 등 연산자를 제공합니다.
- `Mono`: 0 또는 1 요소를 처리합니다.
- `Scheduler`: 실행 쓰레드 추상화합니다. `boundedElastic`, `parallel`, `single` 등을 제공합니다.
- `Context`: 호출 체인 간 전달용 키/값 스토어입니다. 로깅 MDC 같은 용도로 사용됩니다.

### Netty

Netty는 JVM 위에서 동작하는 비동기 이벤트 기반 네트워크 애플리케이션 프레임워크로, Master-Slave Reactor 멀티스레드 모델을 채택했습니다.
Java NIO + Selector + Reactor Pattern 기반으로 구현되어 고성능 네트워크 I/O를 처리합니다.
OS NIO 셀렉터를 감싸서 `EventLoop` 스레드 모델을 제공하며, 각 `EventLoop`는 단일 스레드에서 이벤트를 감지하고 처리하는 실제 Reactor 역할을 수행합니다.

Netty의 구조:

- Boss Group(Main Reactor): 클라이언트 연결 요청을 처리하고 `NioSocketChannel`을 생성
- Worker Group(Sub Reactor): 실제 IO 작업을 처리하는 이벤트 루프들로 구성
- Executor Group: 비즈니스 로직 처리를 위한 스레드 풀

Netty의 핵심 구성 요소:

- `EventLoopGroup`:
    - 여러 `EventLoop`(쓰레드)를 관리하는 그룹

- `EventLoop`:
    - Selector, 큐 감시 + 핸들러 실행 등 실제 IO 작업을 처리하는 이벤트 루프
    - 기본 `EventLoop` 풀 크기 = 코어 수 × 2 (Netty Default).
    - 요청 수천 건도 몇 개 스레드가 multiplexing 처리.
    - `WebClient`에서 블로킹 I/O(DB JDBC 등)를 호출하면 event-loop 스레드를 점유해 성능 저하가 발생할 수 있는데, `Schedulers.boundedElastic()` 등 별도 스레드로 우회.

- `Channel`:
    - 네트워크 연결을 나타내는 추상화

- `ChannelOperations`:
    - Netty Channel의 생명주기에 맞춰 동작하는 Reactive Bridge 객체
    - 소켓 데이터를 `NettyInbound`, `NettyOutbound`를 통해 Reactor에 publish/subscribe
    - `Channel`마다 생성되고 소멸되는 1:1 대응 관계

- `ChannelPipeline`: 여러 `ChannelHandler`가 연결된 처리 파이프라인

    ```sh
    [Inbound Handler1]  -> [Inbound Handler2]  -> [Inbound Handler3]
                                                          |
    [Outbound Handler1] <- [Outbound Handler2] <- [Outbound Handler3]
    ```

- `ChannelHandler`: IO 이벤트나 IO 작업을 인터셉트하고 처리
    - 등록된 이벤트를 선택적으로 인터셉트하거나 투명하게 전달
    - 이벤트를 다음 Handler로 전달하거나 종료
    - 로그 출력, 예외 처리, 성능 통계, 메시지 인코딩 등

### Reactor Netty

Netty와 Reactive Streams의 구현체인 Reactor Core를 기반으로 구축된 리액티브 네트워크 프레임워크입니다.
네트워크 I/O부터 비즈니스 처리까지 완전 Non-blocking 처리할 수 있도록 기능을 제공합니다.
가령 `TcpClient.create().connect()` 결과를 `Mono` 반환하는 등, Netty 채널을 Reactor Core 스트림(Publisher)로 래핑합니다.
HTTP Keep-Alive, 커넥션 풀, 백프레셔 모두 Reactive 방식으로 처리합니다.

계층 구조:
- 네트워크 계층 (Netty)
    - `EventLoopGroup` (Boss/Worker)
    - `Channel` 및 `ChannelPipeline`
    - 기본 IO 처리, Library/User Defined Codec, 로그, 메트릭, 코덱 등 편의 핸들러

- 브릿지 계층 (Reactor Netty)

    `ChannelOperationsHandler`는 브릿지 컴포넌트로 Netty와 Reactive Streams 사이의 연결 레이어 역할을 수행합니다.

    1. `Connection` 생성 및 바인딩
        - 연결이 수립되면 `ChannelOperations` 객체를 생성
        - 이를 `Channel`에 바인딩하여 생명주기를 동기화

    2. 데이터 위임 처리
        - `Channel`에서 수신된 데이터를 읽어서 `ChannelOperations`에 위임
        - `onInboundNext(msg)` 메서드를 통해 데이터 전달

    3. Reactive Streams 조립
        - 비즈니스 처리를 위한 Reactive Streams의 조립 및 구독 수행
        - `Connection` 내에서 데이터 수신 시 처리 파이프라인 구성

    4. 생명주기 관리
        - `Connection`이 닫힐 때 Reactive Streams 구독 취소
        - 리소스 정리 및 메모리 누수 방지

- 리액티브 계층 (Reactor)
    - `FluxReceive`
        - `ChannelOperations`에서 전달받은 수신 데이터를 Queue에 저장
        - 다운스트림으로 데이터를 발행하는 `Flux` 역할 수행
        - Back-Pressure 구현을 통해 흐름 제어
    - `MonoSendMany`
        - 사용자 정의 핸들러에서 처리된 데이터를 수신
        - 소켓으로 데이터를 전송하는 Publisher 겸 Listener 역할
    - 사용자 정의 Reactive Streams Handler

Reactor Netty의 데이터 플로우는 다음과 같이 이뤄집니다:

1. 데이터 수신: Netty `Channel`에서 데이터 수신
2. 브릿지 처리: `ChannelOperationsHandler`가 데이터를 `ChannelOperations`에 전달
3. 리액티브 변환: `ChannelOperations`가 데이터를 `NettyInbound`/`NettyOutbound`로 변환
4. 스트림 처리: `FluxReceive`를 통해 다운스트림으로 데이터 발행
5. 응답 처리: `MonoSendMany`를 통해 처리된 데이터를 소켓에 전송

## Inbound 및 Outbound 데이터 처리 흐름

- Inbound

    ```sh
    클라이언트 요청
        ↓
    Boss Group (연결 수립)
        ↓
    Worker Group (Event Loop)
        ↓
    ChannelPipeline
        ↓
    Netty Handler 부분 (코덱, 로그, 메트릭 등)
        ↓
    ChannelOperationsHandler (브릿지)
        ↓
    ChannelOperations (Reactive Bridge)
        ↓
    FluxReceive (Reactive Publisher)
        ↓
    사용자 정의 Reactive Handler
    ```

- Outbound

    ```sh
    사용자 정의 Reactive Handler
        ↓
    MonoSendMany (Reactive Subscriber)
        ↓
    ChannelOperations
        ↓
    ChannelOperationsHandler
        ↓
    Netty Handler 부분
        ↓
    ChannelPipeline
        ↓
    소켓 전송
    ```

## 예제

### `WebClient` 설정

```java
// WebClient는 기본적으로 Reactor Netty HttpClient 사용
HttpClient httpClient = HttpClient.create()
    .option(ChannelOption.CONNECT_TIMEOUT_MILLIS, 10000);

WebClient webClient = WebClient.builder()
    .clientConnector(new ReactorClientHttpConnector(httpClient))
    .build();
```

### 기본적인 사용법

```java
// 1. 직접 Reactor Netty 사용
TcpClient.create()
    .host("localhost")
    .port(8080)
    .connect()
    .subscribe();

// 2. Spring WebFlux에서 WebClient 사용 (내부적으로 Reactor Netty 활용)
WebClient.create()
    .get()
    .uri("http://example.com")
    .retrieve()
    .bodyToMono(String.class);
```
