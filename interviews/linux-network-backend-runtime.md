# 리눅스 환경과 네트워크를 백엔드 서버 관점에서 이해하기

이 문서는 백엔드 서버가 리눅스 환경에서 네트워크 요청을 받을 때 실제로 어떤 계층을 지나고, 각 계층이 어떤 상태를 소유하며, 장애가 났을 때 어디서 무엇을 확인해야 하는지 설명합니다.
단순히 Linux 명령어나 TCP 상태 이름을 외우는 문서가 아니라, `DNS -> TCP -> TLS -> 프록시 -> Tomcat/Netty -> Spring -> DB -> 응답 송신` 경로를 하나의 실행 흐름으로 따라갑니다.
커널, 소켓, 파일 디스크립터, TCP 상태, 프록시, JVM/Spring/Tomcat, WebFlux, Reactor, Netty, streaming, DB 접근, 운영 장애 분석은 서로 다른 과목이 아닙니다.
실제 서버에서는 한 요청 안에서 이들이 모두 연결됩니다.

이 문서는 `core-interview-guide.md`의 요청 경로 설명을 더 깊게 펼친 심화 브리지 문서입니다.
`core-interview-guide.md`가 면접에서 먼저 꺼낼 큰 줄기를 제공한다면, 이 문서는 그 줄기 뒤의 커널 큐, 버퍼, 런타임 스레드, 이벤트 루프, backpressure, DB/외부 시스템 관측 지점을 길게 추적합니다.
기존 `os-kernel-computer-architecture.md`, `network-web-protocols.md`, `concurrency-async-io.md`, `spring-backend-frameworks.md`, `database-storage-search-nosql.md`를 대체하지 않습니다.
대신 각 문서에 흩어진 개념을 백엔드 서버의 요청 처리라는 하나의 장면으로 다시 연결합니다.

먼저 붙잡을 전체 지도는 아래와 같습니다.

```text
Client
  -> DNS resolver
  -> routing / NAT / firewall
  -> TCP 3-way handshake
  -> TLS handshake
  -> L4/L7 load balancer
  -> Nginx reverse proxy
  -> upstream TCP connection
  -> Tomcat worker or Netty EventLoop
  -> Spring MVC DispatcherServlet or WebFlux DispatcherHandler
  -> service / transaction / connection pool
  -> DB protocol / query execution
  -> response object
  -> proxy response path
  -> kernel send buffer / TCP window
  -> Client receive buffer
```

이 지도에서 각 화살표는 단순한 함수 호출이 아닙니다.
어떤 화살표는 커널이 소유한 TCP 상태 전이이고, 어떤 화살표는 프록시가 HTTP 메시지를 다시 만드는 일이며, 어떤 화살표는 Java 런타임이 byte를 object나 signal로 바꾸는 일입니다.
이 차이를 구분해야 장애 분석도 단단해집니다.

## 목차

- [전체 요청 경로 지도](#전체-요청-경로-지도)
    - [클라이언트 요청은 DNS에서 DB 응답까지 어떤 순서로 왕복하는가](#클라이언트-요청은-dns에서-db-응답까지-어떤-순서로-왕복하는가)
    - [한 요청 경로에서 커널, 프록시, 런타임, 애플리케이션, DB의 책임은 어떻게 나뉘는가](#한-요청-경로에서-커널-프록시-런타임-애플리케이션-db의-책임은-어떻게-나뉘는가)
- [리눅스 실행 기반](#리눅스-실행-기반)
    - [프로세스, 스레드, 파일 디스크립터](#프로세스-스레드-파일-디스크립터)
    - [사용자 공간과 커널 공간, 시스템 콜](#사용자-공간과-커널-공간-시스템-콜)
    - [서버 프로세스는 포트를 어떻게 열고 연결을 받을 준비를 하는가](#서버-프로세스는-포트를-어떻게-열고-연결을-받을-준비를-하는가)
    - [소켓 버퍼와 파일 디스크립터](#소켓-버퍼와-파일-디스크립터)
    - [메모리, page cache, OOM, cgroup](#메모리-page-cache-oom-cgroup)
    - [스케줄링, load average, iowait](#스케줄링-load-average-iowait)
    - [systemd, signal, 로그, ulimit](#systemd-signal-로그-ulimit)
- [네트워크 실행 경로](#네트워크-실행-경로)
    - [DNS, IP, routing, NAT](#dns-ip-routing-nat)
    - [bind, listen, accept는 각각 어느 계층의 일을 하는가](#bind-listen-accept는-각각-어느-계층의-일을-하는가)
    - [TCP 3-way handshake, backlog, accept queue](#tcp-3-way-handshake-backlog-accept-queue)
    - [TCP 상태: ESTABLISHED, CLOSE_WAIT, TIME_WAIT](#tcp-상태-established-close-wait-time-wait)
    - [TLS와 HTTP keep-alive](#tls와-http-keep-alive)
    - [Nginx, L4/L7 load balancer, reverse proxy](#nginx-l4-l7-load-balancer-reverse-proxy)
    - [패킷이 NIC에서 애플리케이션 버퍼까지 가는 길](#패킷이-nic에서-애플리케이션-버퍼까지-가는-길)
    - [요청 수신 경로와 응답 송신 경로는 커널 안에서 어떻게 다르게 보이는가](#요청-수신-경로와-응답-송신-경로는-커널-안에서-어떻게-다르게-보이는가)
    - [방화벽, security group, routing 문제는 connection refused와 timeout을 어떻게 갈라놓는가](#방화벽-security-group-routing-문제는-connection-refused와-timeout을-어떻게-갈라놓는가)
- [I/O multiplexing과 커널 이벤트 모델](#i-o-multiplexing과-커널-이벤트-모델)
    - [select와 poll은 왜 한계가 생겼는가](#select와-poll은-왜-한계가-생겼는가)
    - [epoll은 readiness를 어떤 방식으로 관리하는가](#epoll은-readiness를-어떤-방식으로-관리하는가)
    - [kqueue는 epoll과 어떤 관점에서 비교해야 하는가](#kqueue는-epoll과-어떤-관점에서-비교해야-하는가)
    - [io_uring은 readiness가 아니라 completion을 어떻게 다루는가](#io-uring은-readiness가-아니라-completion을-어떻게-다루는가)
    - [readiness 기반 모델과 completion 기반 모델은 서버 설계를 어떻게 바꾸는가](#readiness-기반-모델과-completion-기반-모델은-서버-설계를-어떻게-바꾸는가)
- [JVM 네트워크 런타임과 비동기 처리](#jvm-네트워크-런타임과-비동기-처리)
    - [Java NIO는 blocking socket 모델과 무엇이 다른가](#java-nio는-blocking-socket-모델과-무엇이-다른가)
    - [Tomcat thread-per-request 모델과 Netty event loop 모델은 어디서 갈라지는가](#tomcat-thread-per-request-모델과-netty-event-loop-모델은-어디서-갈라지는가)
    - [Netty는 selector, event loop, channel, pipeline을 어떻게 조립하는가](#netty는-selector-event-loop-channel-pipeline을-어떻게-조립하는가)
    - [Reactor는 Publisher, Subscriber, backpressure를 어떤 계약으로 묶는가](#reactor는-publisher-subscriber-backpressure를-어떤-계약으로-묶는가)
    - [Spring WebFlux는 Servlet MVC와 요청 처리 모델이 어떻게 다른가](#spring-webflux는-servlet-mvc와-요청-처리-모델이-어떻게-다른가)
    - [WebFlux, Reactor, Netty는 Linux I/O 모델 위에서 어떻게 이어지는가](#webflux-reactor-netty는-linux-i-o-모델-위에서-어떻게-이어지는가)
- [Streaming과 backpressure](#streaming과-backpressure)
    - [HTTP streaming은 응답을 한 번에 만들지 않고 어떻게 흘려보내는가](#http-streaming은-응답을-한-번에-만들지-않고-어떻게-흘려보내는가)
    - [100GB 파일 스트리밍은 page cache, TCP window, backpressure를 어떻게 지나간다](#100gb-파일-스트리밍은-page-cache-tcp-window-backpressure를-어떻게-지나간다)
    - [proxy buffering은 streaming 응답의 의미를 어떻게 바꿀 수 있는가](#proxy-buffering은-streaming-응답의-의미를-어떻게-바꿀-수-있는가)
    - [느린 클라이언트는 서버 스레드, 메모리, 소켓 버퍼에 어떤 압력을 주는가](#느린-클라이언트는-서버-스레드-메모리-소켓-버퍼에-어떤-압력을-주는가)
- [장애 상황으로 묶어 말하기](#장애-상황으로-묶어-말하기)
    - [장애 분석은 증상에서 커널, 런타임, 애플리케이션, 외부 시스템 중 어디로 좁혀 가는가](#장애-분석은-증상에서-커널-런타임-애플리케이션-외부-시스템-중-어디로-좁혀-가는가)
    - [connection refused와 timeout은 어디가 다른가](#connection-refused와-timeout은-어디가-다른가)
    - [CLOSE_WAIT가 쌓이면 무엇을 의심하는가](#close-wait가-쌓이면-무엇을-의심하는가)
    - [TIME_WAIT가 많으면 항상 문제인가](#time-wait가-많으면-항상-문제인가)
    - [too many open files는 왜 네트워크 장애처럼 보이는가](#too-many-open-files는-왜-네트워크-장애처럼-보이는가)
    - [p99 latency가 튈 때 CPU, GC, DB, network를 어떻게 가른다](#p99-latency가-튈-때-cpu-gc-db-network를-어떻게-가른다)
    - [컨테이너에서 OOMKilled가 났을 때 Linux 관점에서 무엇을 본다](#컨테이너에서-oomkilled가-났을-때-linux-관점에서-무엇을-본다)
    - [WebFlux나 Netty로 바꿨는데도 느릴 때 무엇을 의심하는가](#webflux나-netty로-바꿨는데도-느릴-때-무엇을-의심하는가)
    - [커널 파라미터와 애플리케이션 설정은 언제 같이 봐야 하는가](#커널-파라미터와-애플리케이션-설정은-언제-같이-봐야-하는가)
- [면접 질문 세트](#면접-질문-세트)
- [직접 확인하는 명령어](#직접-확인하는-명령어)
- [더 깊게 볼 기존 문서 링크](#더-깊게-볼-기존-문서-링크)
- [근거와 참고 자료](#근거와-참고-자료)

## 전체 요청 경로 지도

이 장은 뒤쪽 세부 항목을 읽기 전에 하나의 요청이 어떤 길을 오가는지 먼저 고정합니다. 여기서 잡은 지도는 뒤의 Linux, network, JVM, streaming, 장애 분석 절에서 계속 다시 참조합니다.

### 클라이언트 요청은 DNS에서 DB 응답까지 어떤 순서로 왕복하는가

요청은 하나의 선이 아니라 여러 번의 주소 결정, 연결 생성, 암호 채널 수립, HTTP 파싱, 프록시 재요청, 런타임 호출, DB 왕복, 응답 쓰기가 이어진 경로입니다. 예를 들어 사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행을 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

여기서 먼저 붙잡을 대상은 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행입니다.
이 대상들은 같은 요청 안에 머물지만 계층이 바뀔 때마다 관측되는 이름과 소유자가 달라집니다.
브라우저에서 시작한 것은 처음에는 도메인 이름이지만, DNS 뒤에는 IP 후보가 되고, TCP 뒤에는 커널이 관리하는 연결 상태가 되며, TLS 뒤에는 암호화된 바이트 흐름이 됩니다.
Nginx가 TLS를 끝내면 같은 사용자의 요청은 다시 upstream HTTP 요청으로 만들어지고, Tomcat이나 Netty는 그 바이트를 런타임의 요청 객체나 이벤트로 바꿉니다.
애플리케이션은 그 요청을 컨트롤러와 서비스 호출로 읽고, DB 접근이 필요하면 또 다른 TCP 연결과 DB 프로토콜 위에서 SQL을 보냅니다.
응답은 이 흐름을 거꾸로 타지만 완전히 대칭은 아닙니다.
DB result set은 Java 객체로, Java 객체는 JSON 바이트로, JSON 바이트는 HTTP body로, HTTP body는 프록시와 커널 send buffer를 거쳐 client receive buffer로 이동합니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 요청 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 요청 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

요청은 하나의 선이 아니라 여러 번의 주소 결정, 연결 생성, 암호 채널 수립, HTTP 파싱, 프록시 재요청, 런타임 호출, DB 왕복, 응답 쓰기가 이어진 경로입니다.
이 문장을 실제 서버로 옮기면 먼저 "사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
요청 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`의 소유권을 확인하는 근거로 사용합니다.
요청 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 요청 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
요청 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

요청 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
dig +trace api.example.com
curl -v --http1.1 https://api.example.com/orders/1
ss -antp | grep ':443\|:8080'
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 요청 경로를 확인할 때는 이름 하나만 보지 말고, 주소 결정 결과, TCP 상태, 프록시 timing, JVM 대기, DB wait가 같은 요청에서 어떻게 이어지는지 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`요청은 하나의 선이 아니라 여러 번의 주소 결정, 연결 생성, 암호 채널 수립, HTTP 파싱, 프록시 재요청, 런타임 호출, DB 왕복, 응답 쓰기가 이어진 경로입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
요청 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 요청 경로의 커널 상태 vs 애플리케이션 상태 | `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 요청 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 요청 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 요청 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 요청 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 요청 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

여기서 비교는 선택지의 우열을 한 번에 정하는 도구가 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "요청은 하나의 선이 아니라 여러 번의 주소 결정, 연결 생성, 암호 채널 수립, HTTP 파싱, 프록시 재요청, 런타임 호출, DB 왕복, 응답 쓰기가 이어진 경로입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`를 먼저 말하고, 그다음 `요청은 하나의 선이 아니라 여러 번의 주소 결정, 연결 생성, 암호 채널 수립, HTTP 파싱, 프록시 재요청, 런타임 호출, DB 왕복, 응답 쓰기가 이어진 경로입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

요청 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

요청 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `dig +trace api.example.com` 같은 확인은 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `사용자가 브라우저에서 주문 조회 API를 호출하고, Spring 서비스가 DB에서 행을 읽어 JSON 응답을 돌려주는 장면`에서는 도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `DNS가 끝나면 HTTP가 바로 간다고 생각하거나, Nginx와 Tomcat 사이의 upstream 연결을 같은 TCP 연결로 합쳐 버리는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

#### 운영 질문으로 다시 압축하기

요청 경로를 실제 면접 답변으로 꺼낼 때는 먼저 요청 경로에서 무엇을 구분하려는지 말해야 합니다. 핵심은 클라이언트가 본 지연과 서버 내부 지연을 같은 이벤트(event)로 합치지 않는 것입니다. 예를 들어 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 요청 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 DNS, TCP, TLS, 프록시, 런타임, DB가 남긴 시간을 한 요청 ID나 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `도메인 이름, TCP 세그먼트, TLS 레코드, HTTP 메시지, SQL 결과 행` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 요청 경로의 세부 메커니즘으로 내려가면 됩니다.

### 한 요청 경로에서 커널, 프록시, 런타임, 애플리케이션, DB의 책임은 어떻게 나뉘는가

책임 경계를 나누면 관측 지표가 정리됩니다. 커널은 연결과 버퍼, 프록시는 HTTP 경계와 upstream, 런타임은 스레드와 이벤트 루프, 애플리케이션은 비즈니스 흐름, DB는 쿼리와 저장소 상태를 맡습니다. 예를 들어 같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

여기서 먼저 붙잡을 대상은 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set입니다.
이 대상들은 같은 요청 안에 머물지만 계층이 바뀔 때마다 관측되는 이름과 소유자가 달라집니다.
브라우저에서 시작한 것은 처음에는 도메인 이름이지만, DNS 뒤에는 IP 후보가 되고, TCP 뒤에는 커널이 관리하는 연결 상태가 되며, TLS 뒤에는 암호화된 바이트 흐름이 됩니다.
Nginx가 TLS를 끝내면 같은 사용자의 요청은 다시 upstream HTTP 요청으로 만들어지고, Tomcat이나 Netty는 그 바이트를 런타임의 요청 객체나 이벤트로 바꿉니다.
애플리케이션은 그 요청을 컨트롤러와 서비스 호출로 읽고, DB 접근이 필요하면 또 다른 TCP 연결과 DB 프로토콜 위에서 SQL을 보냅니다.
응답은 이 흐름을 거꾸로 타지만 완전히 대칭은 아닙니다.
DB result set은 Java 객체로, Java 객체는 JSON 바이트로, JSON 바이트는 HTTP body로, HTTP body는 프록시와 커널 send buffer를 거쳐 client receive buffer로 이동합니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 요청 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 요청 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

책임 경계를 나누면 관측 지표가 정리됩니다. 커널은 연결과 버퍼, 프록시는 HTTP 경계와 upstream, 런타임은 스레드와 이벤트 루프, 애플리케이션은 비즈니스 흐름, DB는 쿼리와 저장소 상태를 맡습니다.
이 문장을 실제 서버로 옮기면 먼저 "같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
요청 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`의 소유권을 확인하는 근거로 사용합니다.
요청 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 요청 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
요청 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

요청 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -ltnp
nginx -T | grep -E 'proxy_(pass|buffering|read_timeout)'
jcmd <pid> Thread.print
SELECT state, wait_event_type, wait_event FROM pg_stat_activity;
```

`요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 요청 경로를 확인할 때는 이름 하나만 보지 말고, 주소 결정 결과, TCP 상태, 프록시 timing, JVM 대기, DB wait가 같은 요청에서 어떻게 이어지는지 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`책임 경계를 나누면 관측 지표가 정리됩니다. 커널은 연결과 버퍼, 프록시는 HTTP 경계와 upstream, 런타임은 스레드와 이벤트 루프, 애플리케이션은 비즈니스 흐름, DB는 쿼리와 저장소 상태를 맡습니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 단일 로그에서만 찾는 것입니다.
요청 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 요청 경로의 커널 상태 vs 애플리케이션 상태 | `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 요청 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 요청 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 요청 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 요청 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 요청 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "책임 경계를 나누면 관측 지표가 정리됩니다. 커널은 연결과 버퍼, 프록시는 HTTP 경계와 upstream, 런타임은 스레드와 이벤트 루프, 애플리케이션은 비즈니스 흐름, DB는 쿼리와 저장소 상태를 맡습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set`를 먼저 말하고, 그다음 `책임 경계를 나누면 관측 지표가 정리됩니다. 커널은 연결과 버퍼, 프록시는 HTTP 경계와 upstream, 런타임은 스레드와 이벤트 루프, 애플리케이션은 비즈니스 흐름, DB는 쿼리와 저장소 상태를 맡습니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

요청 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

요청 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -ltnp` 같은 확인은 `요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서는 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `모든 지연을 애플리케이션 코드 탓으로 보거나, 모든 연결 실패를 방화벽 탓으로 보는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`같은 지연 시간을 두고 커널 큐가 찼는지, 프록시가 buffering 중인지, Tomcat worker가 모자란지, DB가 잠긴 것인지 좁혀 가는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 요청 바이트와 그 바이트를 해석한 HTTP request object, Java 객체, SQL result set이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

## 리눅스 실행 기반

이 장은 Spring이나 Netty가 올라타기 전의 실행 바닥을 설명합니다. 프로세스, 스레드, fd, syscall, memory, scheduler, service manager를 먼저 잡아야 서버 장애가 Java 로그 하나로만 보이지 않습니다.

### 프로세스, 스레드, 파일 디스크립터

백엔드 서버는 프로세스 하나가 아니라 fd table, native thread, heap, direct buffer, socket 객체가 붙어 있는 실행 단위입니다. 예를 들어 Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 먼저 붙잡을 대상은 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

백엔드 서버는 프로세스 하나가 아니라 fd table, native thread, heap, direct buffer, socket 객체가 붙어 있는 실행 단위입니다.
이 문장을 실제 서버로 옮기면 먼저 "Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html), [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ps -o pid,ppid,stat,comm -p <pid>
ls -l /proc/<pid>/fd
lsof -p <pid> | head
jcmd <pid> Thread.print
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 Linux 실행 기반을 확인할 때는 `/proc/<pid>/fd`, `/proc/<pid>/limits`, `ps -L`, `top -H`, `journalctl`처럼 프로세스가 커널 자원을 어떻게 쓰는지 보여 주는 표면을 먼저 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`백엔드 서버는 프로세스 하나가 아니라 fd table, native thread, heap, direct buffer, socket 객체가 붙어 있는 실행 단위입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
Linux 실행 기반을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 Linux 실행 기반을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "백엔드 서버는 프로세스 하나가 아니라 fd table, native thread, heap, direct buffer, socket 객체가 붙어 있는 실행 단위입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 먼저 말하고, 그다음 `백엔드 서버는 프로세스 하나가 아니라 fd table, native thread, heap, direct buffer, socket 객체가 붙어 있는 실행 단위입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ps -o pid,ppid,stat,comm -p <pid>` 같은 확인은 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서는 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `파일 디스크립터를 실제 파일 경로로만 이해해서 socket, pipe, epoll fd, eventfd까지 같은 한도에 걸린다는 점을 놓치는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Spring Boot 프로세스 하나가 8080 포트를 열고 worker thread와 DB connection을 함께 유지하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 `/proc`, `systemctl`, `journalctl`, `ps`, `top`, `ulimit`처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `프로세스의 fd table entry, 커널 open file description, socket 객체, Java Thread` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

### 사용자 공간과 커널 공간, 시스템 콜

Java 메서드 호출과 네트워크 전송 사이에는 시스템 콜 경계가 있습니다. Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣을 때도 실제 이동은 사용자 공간 버퍼, 시스템 콜 인자, 커널 소켓 버퍼를 차례로 지납니다. 이 경계를 알아야 `read()`가 느린 것인지, 애플리케이션 계산이 느린 것인지, send buffer가 차서 `write()`가 밀리는 것인지 구분할 수 있습니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 먼저 붙잡을 대상은 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Java 메서드 호출과 네트워크 전송 사이에는 시스템 콜 경계가 있습니다. 이 경계에서 복사, blocking, errno, signal interruption, buffer pressure가 드러납니다.
이 문장을 실제 서버로 옮기면 먼저 "Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html), [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -f -e trace=network,read,write -p <pid>
perf trace -p <pid>
cat /proc/<pid>/syscall
```

시스템 콜 경계를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 Linux 실행 기반을 확인할 때는 `/proc/<pid>/fd`, `/proc/<pid>/limits`, `ps -L`, `top -H`, `journalctl`처럼 프로세스가 커널 자원을 어떻게 쓰는지 보여 주는 표면을 먼저 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Java 메서드 호출과 네트워크 전송 사이에는 시스템 콜 경계가 있습니다. 이 경계에서 복사, blocking, errno, signal interruption, buffer pressure가 드러납니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 시스템 콜 경계를 단일 로그에서만 찾는 것입니다.
Linux 실행 기반을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

시스템 콜 경계를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 Linux 실행 기반을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

시스템 콜 경계를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Java 메서드 호출과 네트워크 전송 사이에는 시스템 콜 경계가 있습니다. 이 경계에서 복사, blocking, errno, signal interruption, buffer pressure가 드러납니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`를 먼저 말하고, 그다음 `Java 메서드 호출과 네트워크 전송 사이에는 시스템 콜 경계가 있습니다. 이 경계에서 복사, blocking, errno, signal interruption, buffer pressure가 드러납니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서는 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
Tomcat worker가 `read()`로 요청 body 일부를 가져오고 `write()`로 response chunk를 밀어 넣는 장면에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `strace -f -e trace=network,read,write -p <pid>` 같은 확인은 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `JVM 메서드가 곧바로 NIC에 쓴다고 생각하거나, kernel이 Java 객체를 안다고 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

시스템 콜 경계의 관측 근거는 `/proc`, `systemctl`, `journalctl`, `ps`, `top`, `ulimit`처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `사용자 공간 버퍼의 byte, syscall 인자, 커널 socket buffer의 byte` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

### 서버 프로세스는 포트를 어떻게 열고 연결을 받을 준비를 하는가

서버는 요청이 온 뒤 포트를 여는 것이 아니라 시작할 때 로컬 주소와 포트에 socket을 묶고, listening state로 바꾼 뒤 accept 가능한 연결을 기다립니다. 예를 들어 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름을 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 먼저 붙잡을 대상은 socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

서버는 요청이 온 뒤 포트를 여는 것이 아니라 시작할 때 로컬 주소와 포트에 socket을 묶고, listening state로 바꾼 뒤 accept 가능한 연결을 기다립니다.
이 문장을 실제 서버로 옮기면 먼저 "Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html), [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -f -e socket,bind,listen,accept4 java -jar app.jar
ss -ltnp 'sport = :8080'
sysctl net.core.somaxconn
```

포트 준비 과정을 볼 때 확인해야 하는 신호는 "이 프로세스가 실제로 어느 주소와 포트에 리스닝 소켓을 만들었는가"입니다.
`strace`에서는 JVM이 시작되는 동안 `socket`, `bind`, `listen`, `accept4` 계열 호출이 어떤 순서로 나오는지 볼 수 있고, `ss -ltnp`에서는 해당 포트가 LISTEN 상태인지와 어느 프로세스가 소유하는지 확인할 수 있습니다.
`/proc/<pid>/fd`는 그 리스닝 소켓이 프로세스 fd table에 붙어 있음을 보여 주고, `/proc/<pid>/limits`는 새 연결을 받을 여유가 fd 한도에 막히지 않는지 확인하게 해 줍니다.
Spring Boot 로그의 `Tomcat started on port(s)` 같은 문구는 애플리케이션 관점의 시작 완료 신호이므로, 커널이 실제로 포트를 열었는지 확인하는 명령과 함께 봐야 합니다.
Nginx나 로드밸런서가 앞에 있으면 클라이언트가 보는 endpoint와 Tomcat이 실제로 바인딩한 endpoint가 다를 수 있으므로, client-facing 포트와 upstream 포트를 나누어 확인해야 합니다.

어긋나는 신호는 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`서버는 요청이 온 뒤 포트를 여는 것이 아니라 시작할 때 로컬 주소와 포트에 socket을 묶고, listening state로 바꾼 뒤 accept 가능한 연결을 기다립니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 포트 준비 실패를 애플리케이션 로그 하나에서만 찾는 것입니다.
예를 들어 Spring 로그에는 정상 기동처럼 보이는데 `ss -ltnp`에는 리스닝 소켓이 없거나, 반대로 포트는 열렸지만 로드밸런서 health check가 계속 실패할 수 있습니다.
이럴 때는 application log, `journalctl`, `ss`, `/proc/<pid>/fd`, Nginx upstream log, 로드밸런서 health check 결과를 같은 시간대에 맞춰야 합니다.
한 표면에 아무것도 없다는 사실은 그 표면에 증거가 없다는 뜻일 뿐, 연결 생성 경로 전체가 정상이라는 뜻은 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 포트 준비 과정에 적용하면 "서버가 안 뜬다"는 말을 여러 질문으로 나눌 수 있습니다.
먼저 같은 포트를 이미 다른 프로세스가 쓰고 있어서 `bind()`가 실패했는지, systemd socket activation이나 컨테이너 포트 매핑 때문에 예상한 포트와 실제 리스닝 포트가 다른지, 아니면 accept 가능한 연결은 생기지만 worker thread나 event loop가 뒤에서 막히는지 구분합니다.
각각의 확인 명령도 다릅니다.
포트 점유와 리스닝 상태는 `ss -ltnp`와 `lsof -iTCP:<port>`를 보고, systemd 경계는 unit 설정과 `journalctl`을 보며, 런타임 처리 지연은 thread dump나 event-loop 지연 지표를 봅니다.

이 비교는 Spring MVC와 WebFlux 중 어느 쪽이 더 좋은지를 한 번에 정하는 도구가 아닙니다.
포트를 열고 연결을 받을 준비를 하는 단계에서는 두 모델 모두 결국 커널 소켓과 fd를 사용합니다.
차이는 그 뒤의 요청 처리 단계에서 worker thread가 요청을 붙잡고 기다리느냐, event loop와 callback 흐름으로 readiness를 처리하느냐에서 커집니다.
따라서 포트가 열리지 않는 문제에는 MVC/WebFlux 선택보다 bind 주소, 권한, 포트 충돌, container port mapping, listen backlog 같은 관측값이 먼저입니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "서버는 요청이 온 뒤 포트를 여는 것이 아니라 시작할 때 로컬 주소와 포트에 socket을 묶고, listening state로 바꾼 뒤 accept 가능한 연결을 기다립니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화를 먼저 말하고, 그다음 `서버는 요청이 온 뒤 포트를 여는 것이 아니라 시작할 때 로컬 주소와 포트에 socket을 묶고, listening state로 바꾼 뒤 accept 가능한 연결을 기다립니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 포트 준비 흐름이 어느 계층의 상태로 관측되는지 먼저 물어보셔야 합니다.
이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서는 socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `strace -f -e socket,bind,listen,accept4 java -jar app.jar` 같은 확인은 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `클라이언트마다 서버 local port가 새로 생긴다고 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
Spring Boot 내장 Tomcat이 `server.port=8080`을 읽고 리스닝 소켓을 준비하는 장면에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 socket fd가 `socket -> bind -> listen -> accept` 상태로 바뀌는 흐름 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

포트 준비 흐름의 관측 근거는 `ss`, `lsof`, `/proc/<pid>/fd`, `systemctl`, `journalctl`, container runtime 로그처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 커널이 알고 있는 리스닝 소켓 상태이고, Spring Boot 로그가 보여 주는 것은 런타임이 connector 시작을 완료했다고 판단한 시점이며, Nginx upstream log가 보여 주는 것은 프록시가 그 포트로 실제 요청을 보냈을 때의 결과입니다. 이 신호들이 같은 시간대에 맞물리면 가설이 강해지고, 맞물리지 않으면 포트 매핑, 방화벽, health check, 런타임 초기화 순서 중 어느 경계가 어긋났는지 다시 보셔야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `socket -> bind -> listen -> accept`로 이어지는 socket fd 상태 변화 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

### 소켓 버퍼와 파일 디스크립터

fd는 커널 객체를 가리키는 손잡이고, 소켓 버퍼는 그 객체 안에서 네트워크 byte를 임시로 보관하는 공간입니다. 예를 들어 10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 먼저 붙잡을 대상은 NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

fd는 커널 객체를 가리키는 손잡이고, 소켓 버퍼는 그 객체 안에서 네트워크 byte를 임시로 보관하는 공간입니다.
이 문장을 실제 서버로 옮기면 먼저 "10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html), [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -tinp 'sport = :8080'
cat /proc/sys/net/ipv4/tcp_rmem
cat /proc/sys/net/ipv4/tcp_wmem
lsof -p <pid> | grep TCP
```

`NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 Linux 실행 기반을 확인할 때는 `/proc/<pid>/fd`, `/proc/<pid>/limits`, `ps -L`, `top -H`, `journalctl`처럼 프로세스가 커널 자원을 어떻게 쓰는지 보여 주는 표면을 먼저 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`fd는 커널 객체를 가리키는 손잡이고, 소켓 버퍼는 그 객체 안에서 네트워크 byte를 임시로 보관하는 공간입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 단일 로그에서만 찾는 것입니다.
Linux 실행 기반을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 Linux 실행 기반을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "fd는 커널 객체를 가리키는 손잡이고, 소켓 버퍼는 그 객체 안에서 네트워크 byte를 임시로 보관하는 공간입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 먼저 말하고, 그다음 `fd는 커널 객체를 가리키는 손잡이고, 소켓 버퍼는 그 객체 안에서 네트워크 byte를 임시로 보관하는 공간입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -tinp 'sport = :8080'` 같은 확인은 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `10MB upload body가 여러 TCP segment로 들어오고 애플리케이션이 조금씩 읽는 장면`에서는 NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Content-Length 전체가 한 번에 애플리케이션 메모리에 올라온 뒤 처리된다고 보는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer`의 관측 근거는 `/proc`, `systemctl`, `journalctl`, `ps`, `top`, `ulimit`처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `NIC가 받은 byte, 커널 receive buffer, 사용자 공간 ByteBuffer 또는 ServletInputStream buffer` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

### 메모리, page cache, OOM, cgroup

Linux 메모리 압박은 JVM heap만 보아서는 닫히지 않습니다. page cache, direct buffer, native thread stack, cgroup limit, OOM killer 판정이 함께 움직입니다. 예를 들어 컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 따라갈 대상은 anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `fd, 큐, 버퍼, 스레드, 연결 상태`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `fd, 큐, 버퍼, 스레드, 연결 상태`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. 컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Linux 메모리 압박은 JVM heap만 보아서는 닫히지 않습니다. page cache, direct buffer, native thread stack, cgroup limit, OOM killer 판정이 함께 움직입니다.
이 문장을 실제 서버로 옮기면 먼저 "컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `fd, 큐, 버퍼, 스레드, 연결 상태`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html), [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `fd, 큐, 버퍼, 스레드, 연결 상태`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `fd, 큐, 버퍼, 스레드, 연결 상태`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
cat /sys/fs/cgroup/memory.current
cat /sys/fs/cgroup/memory.events
cat /proc/<pid>/status | grep -E 'VmRSS|VmHWM|Threads'
dmesg -T | grep -i 'killed process'
```

`anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 Linux 실행 기반을 확인할 때는 `/proc/<pid>/fd`, `/proc/<pid>/limits`, `ps -L`, `top -H`, `journalctl`처럼 프로세스가 커널 자원을 어떻게 쓰는지 보여 주는 표면을 먼저 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Linux 메모리 압박은 JVM heap만 보아서는 닫히지 않습니다. page cache, direct buffer, native thread stack, cgroup limit, OOM killer 판정이 함께 움직입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 단일 로그에서만 찾는 것입니다.
Linux 실행 기반을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | 컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 Linux 실행 기반을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Linux 메모리 압박은 JVM heap만 보아서는 닫히지 않습니다. page cache, direct buffer, native thread stack, cgroup limit, OOM killer 판정이 함께 움직입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`를 먼저 말하고, 그다음 `Linux 메모리 압박은 JVM heap만 보아서는 닫히지 않습니다. page cache, direct buffer, native thread stack, cgroup limit, OOM killer 판정이 함께 움직입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `cat /sys/fs/cgroup/memory.current` 같은 확인은 `anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하는 경우 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서는 anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 컨테이너 OOM을 Java `OutOfMemoryError`와 같은 이벤트로만 설명하면 부족합니다. cgroup 한도, RSS, direct memory, page cache, OOM killer 기록을 같이 보지 못하기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`컨테이너 안의 Spring 서비스가 파일 스트리밍과 JSON 직렬화를 동시에 하다가 OOMKilled 되는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `fd, 큐, 버퍼, 스레드, 연결 상태`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `fd, 큐, 버퍼, 스레드, 연결 상태`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`anonymous heap, direct memory, page cache, cgroup memory.current, OOM victim`의 관측 근거는 `/proc`, `systemctl`, `journalctl`, `ps`, `top`, `ulimit`처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `fd, 큐, 버퍼, 스레드, 연결 상태` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

### 스케줄링, load average, iowait

load average는 CPU 사용률이 아니라 runnable 또는 일부 uninterruptible 상태 task의 평균 압박을 보여 주는 신호입니다. 예를 들어 CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 runnable thread, run queue, blocked I/O task, CPU time slice를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 먼저 붙잡을 대상은 runnable thread, run queue, blocked I/O task, CPU time slice이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: runnable thread, run queue, blocked I/O task, CPU time slice
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `runnable thread, run queue, blocked I/O task, CPU time slice`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "load average가 8이면 CPU 800% 사용이라고 단정하는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `runnable thread, run queue, blocked I/O task, CPU time slice` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `runnable thread, run queue, blocked I/O task, CPU time slice`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `runnable thread, run queue, blocked I/O task, CPU time slice`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `load average가 8이면 CPU 800% 사용이라고 단정하는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

load average는 CPU 사용률이 아니라 runnable 또는 일부 uninterruptible 상태 task의 평균 압박을 보여 주는 신호입니다.
이 문장을 실제 서버로 옮기면 먼저 "CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 runnable thread, run queue, blocked I/O task, CPU time slice이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `runnable thread, run queue, blocked I/O task, CPU time slice`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html), [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `runnable thread, run queue, blocked I/O task, CPU time slice`의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `runnable thread, run queue, blocked I/O task, CPU time slice`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`runnable thread, run queue, blocked I/O task, CPU time slice`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `runnable thread, run queue, blocked I/O task, CPU time slice`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
uptime
vmstat 1
pidstat -t -p <pid> 1
top -H -p <pid>
```

`runnable thread, run queue, blocked I/O task, CPU time slice`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 Linux 실행 기반을 확인할 때는 `/proc/<pid>/fd`, `/proc/<pid>/limits`, `ps -L`, `top -H`, `journalctl`처럼 프로세스가 커널 자원을 어떻게 쓰는지 보여 주는 표면을 먼저 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `runnable thread, run queue, blocked I/O task, CPU time slice`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `load average가 8이면 CPU 800% 사용이라고 단정하는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
load average가 8이면 CPU 800% 사용이라고 단정하는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`load average는 CPU 사용률이 아니라 runnable 또는 일부 uninterruptible 상태 task의 평균 압박을 보여 주는 신호입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `runnable thread, run queue, blocked I/O task, CPU time slice`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `runnable thread, run queue, blocked I/O task, CPU time slice`를 단일 로그에서만 찾는 것입니다.
Linux 실행 기반을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `runnable thread, run queue, blocked I/O task, CPU time slice` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `load average가 8이면 CPU 800% 사용이라고 단정하는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`runnable thread, run queue, blocked I/O task, CPU time slice`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 Linux 실행 기반을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`runnable thread, run queue, blocked I/O task, CPU time slice`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "load average는 CPU 사용률이 아니라 runnable 또는 일부 uninterruptible 상태 task의 평균 압박을 보여 주는 신호입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: runnable thread, run queue, blocked I/O task, CPU time slice
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `runnable thread, run queue, blocked I/O task, CPU time slice`를 먼저 말하고, 그다음 `load average는 CPU 사용률이 아니라 runnable 또는 일부 uninterruptible 상태 task의 평균 압박을 보여 주는 신호입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `load average가 8이면 CPU 800% 사용이라고 단정하는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `uptime` 같은 확인은 `runnable thread, run queue, blocked I/O task, CPU time slice` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `load average가 8이면 CPU 800% 사용이라고 단정하는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 runnable thread, run queue, blocked I/O task, CPU time slice 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서는 runnable thread, run queue, blocked I/O task, CPU time slice 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `load average가 8이면 CPU 800% 사용이라고 단정하는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 runnable thread, run queue, blocked I/O task, CPU time slice이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`runnable thread, run queue, blocked I/O task, CPU time slice` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`CPU 사용률은 낮은데 load average가 높고 API p99가 튀는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `runnable thread, run queue, blocked I/O task, CPU time slice`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `runnable thread, run queue, blocked I/O task, CPU time slice`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`runnable thread, run queue, blocked I/O task, CPU time slice`의 관측 근거는 `/proc`, `systemctl`, `journalctl`, `ps`, `top`, `ulimit`처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `runnable thread, run queue, blocked I/O task, CPU time slice` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

### systemd, signal, 로그, ulimit

운영에서 서버 프로세스는 단순히 `java -jar`로 떠 있는 것이 아니라 service manager, signal, file limit, stdout/stderr 수집, restart policy 안에서 삽니다. 예를 들어 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 service unit, process signal, journal log, resource limit를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

리눅스 실행 기반을 볼 때는 Java나 Spring 이름을 잠시 내려놓고 프로세스가 커널에게 빌려 쓰는 자원을 먼저 봅니다.
여기서 먼저 붙잡을 대상은 service unit, process signal, journal log, resource limit이며, 이 대상은 애플리케이션 코드 안에만 있는 값이 아니라 프로세스의 descriptor table, 커널 socket 객체, scheduler queue, memory accounting, service manager 상태와 연결됩니다.
운영 장애에서 이 계층을 놓치면 로그에는 `Connection reset`, `Broken pipe`, `OutOfMemoryError`, `too many open files`처럼 서로 다른 문구가 찍히는데, 실제 공통 원인은 fd 한도, cgroup memory limit, thread scheduling 지연, signal 처리 실패일 수 있습니다.
따라서 리눅스 실행 기반은 "서버가 돌아가는 배경"이 아니라 요청 처리의 첫 번째 런타임입니다.

#### 값이 움직이는 순서

여기서는 client request body의 이동보다 서버 프로세스의 생명주기와 운영 한도가 요청 처리에 어떤 영향을 주는지 먼저 봅니다.
기준 장면은 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면이고, 따라갈 대상은 service unit, process signal, journal log, resource limit입니다.

```text
systemctl restart 또는 배포 도구
    |
    | 1. service manager가 프로세스에 SIGTERM을 보냅니다
    v
JVM process
    |
    | 2. shutdown hook, Spring graceful shutdown, Tomcat connector pause가 실행됩니다
    v
in-flight request / socket / DB connection
    |
    | 3. worker thread와 connection pool이 새 요청 수락을 멈추고 남은 요청을 정리합니다
    v
systemd timeout / signal policy / journal
    |
    | 4. 제한 시간 안에 끝나지 않으면 SIGKILL, restart policy, journal log가 다음 증거가 된다
```

`ulimit`도 같은 운영 경계에 있습니다.
`RLIMIT_NOFILE`이 낮으면 새 socket fd, DB connection fd, log file fd가 모두 같은 한도 안에서 경쟁합니다.
그래서 이 항목의 trace는 요청 byte의 흐름이 아니라 `service manager -> signal -> JVM shutdown/drain -> resource limit -> journal evidence` 흐름으로 읽어야 합니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `service unit, process signal, journal log, resource limit` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 Linux 실행 기반의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `service unit, process signal, journal log, resource limit`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 Linux 실행 기반을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `service unit, process signal, journal log, resource limit`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

운영에서 서버 프로세스는 단순히 `java -jar`로 떠 있는 것이 아니라 service manager, signal, file limit, stdout/stderr 수집, restart policy 안에서 살아갑니다.
이 문장을 실제 서버로 옮기면 먼저 "배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 service unit, process signal, journal log, resource limit이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
Linux 실행 기반을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `service unit, process signal, journal log, resource limit`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [systemd.service(5)](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html), [Linux signal(7)](https://www.man7.org/linux/man-pages/man7/signal.7.html), [Linux getrlimit(2)](https://www.man7.org/linux/man-pages/man2/getrlimit.2.html), [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `service unit, process signal, journal log, resource limit`의 소유권을 확인하는 근거로 사용합니다.
Linux 실행 기반의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `service unit, process signal, journal log, resource limit`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 Linux 실행 기반을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`service unit, process signal, journal log, resource limit`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `service unit, process signal, journal log, resource limit`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
Linux 실행 기반을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

Linux 실행 기반의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
systemctl status myapp
journalctl -u myapp -n 200
systemctl show myapp -p LimitNOFILE -p MemoryMax
cat /proc/<pid>/limits
```

`service unit, process signal, journal log, resource limit`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 Linux 실행 기반을 확인할 때는 `/proc/<pid>/fd`, `/proc/<pid>/limits`, `ps -L`, `top -H`, `journalctl`처럼 프로세스가 커널 자원을 어떻게 쓰는지 보여 주는 표면을 먼저 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `service unit, process signal, journal log, resource limit`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`운영에서 서버 프로세스는 단순히`java -jar`로 떠 있는 것이 아니라 service manager, signal, file limit, stdout/stderr 수집, restart policy 안에서 삽니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `service unit, process signal, journal log, resource limit`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `service unit, process signal, journal log, resource limit`를 단일 로그에서만 찾는 것입니다.
Linux 실행 기반을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| Linux 실행 기반의 커널 상태 vs 애플리케이션 상태 | `service unit, process signal, journal log, resource limit` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| Linux 실행 기반의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| Linux 실행 기반의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| Linux 실행 기반의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| Linux 실행 기반의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`service unit, process signal, journal log, resource limit`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 Linux 실행 기반을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`service unit, process signal, journal log, resource limit`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "운영에서 서버 프로세스는 단순히 `java -jar`로 떠 있는 것이 아니라 service manager, signal, file limit, stdout/stderr 수집, restart policy 안에서 살아갑니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: service unit, process signal, journal log, resource limit
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `service unit, process signal, journal log, resource limit`를 먼저 말하고, 그다음 `운영에서 서버 프로세스는 단순히`java -jar`로 떠 있는 것이 아니라 service manager, signal, file limit, stdout/stderr 수집, restart policy 안에서 삽니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 service unit, process signal, journal log, resource limit 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서는 service unit, process signal, journal log, resource limit 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 service unit, process signal, journal log, resource limit이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

Linux 실행 기반의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`service unit, process signal, journal log, resource limit` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
배포 중 `systemctl restart`가 Java 프로세스에 SIGTERM을 보내고 graceful shutdown이 시작되는 장면에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

Linux 실행 기반을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `systemctl status myapp` 같은 확인은 `service unit, process signal, journal log, resource limit` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `프로세스 종료를 항상 kill -9로 보거나, too many open files를 애플리케이션 예외만으로 보는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

#### 운영 질문으로 다시 압축하기

Linux 실행 기반을 실제 면접 답변으로 꺼낼 때는 먼저 프로세스 실행 기반에서 어떤 상태를 확인하려는지 말해야 합니다. 핵심은 Java 코드 밖에 있는 fd, 스레드, 메모리, signal, service manager 상태를 함께 보는 것입니다. 예를 들어 `service unit, process signal, journal log, resource limit`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 Linux 실행 기반을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `service unit, process signal, journal log, resource limit`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`service unit, process signal, journal log, resource limit`의 관측 근거는 `/proc`, `systemctl`, `journalctl`, `ps`, `top`, `ulimit`처럼 프로세스와 커널 경계를 보여 주는 표면에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `service unit, process signal, journal log, resource limit` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 Linux 실행 기반의 세부 메커니즘으로 내려가면 됩니다.

## 네트워크 실행 경로

이 장은 packet과 연결 상태가 커널 안에서 어떤 큐와 상태 이름으로 보이는지 설명합니다. DNS와 routing부터 TCP 상태, TLS, proxy, firewall 증상까지 한 경로로 묶어 봅니다.

### DNS, IP, routing, NAT

DNS는 이름을 주소 후보로 바꾸고, routing은 그 주소로 나갈 인터페이스와 next hop을 고르며, NAT는 패킷의 주소/포트 일부를 경계에서 바꿉니다. 예를 들어 Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

DNS는 이름을 주소 후보로 바꾸고, routing은 그 주소로 나갈 인터페이스와 next hop을 고르며, NAT는 패킷의 주소/포트 일부를 경계에서 바꿉니다.
이 문장을 실제 서버로 옮기면 먼저 "Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
dig api.partner.example
ip route get 203.0.113.10
conntrack -L 2>/dev/null | head
traceroute 203.0.113.10
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`DNS는 이름을 주소 후보로 바꾸고, routing은 그 주소로 나갈 인터페이스와 next hop을 고르며, NAT는 패킷의 주소/포트 일부를 경계에서 바꿉니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "DNS는 이름을 주소 후보로 바꾸고, routing은 그 주소로 나갈 인터페이스와 next hop을 고르며, NAT는 패킷의 주소/포트 일부를 경계에서 바꿉니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 먼저 말하고, 그다음 `DNS는 이름을 주소 후보로 바꾸고, routing은 그 주소로 나갈 인터페이스와 next hop을 고르며, NAT는 패킷의 주소/포트 일부를 경계에서 바꿉니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서는 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `dig api.partner.example` 같은 확인은 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `DNS, routing, NAT 문제를 모두 timeout이라는 한 단어로 뭉개는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Pod에서 외부 결제 API로 나가는 요청이 NAT 게이트웨이를 지나고 응답이 돌아오는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `도메인 이름, destination IP, route table entry, NAT 변환된 5-tuple` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### bind, listen, accept는 각각 어느 계층의 일을 하는가

`bind`는 이름 붙이기, `listen`은 새 연결 입구 만들기, `accept`는 완성 연결을 사용자 공간으로 꺼내는 단계입니다. 예를 들어 Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 socket fd, local endpoint binding, listening queue, connected socket fd를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 socket fd, local endpoint binding, listening queue, connected socket fd이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: socket fd, local endpoint binding, listening queue, connected socket fd
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `socket fd, local endpoint binding, listening queue, connected socket fd`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "`accept`가 TCP handshake를 직접 수행한다고 말하거나 `listen`이 요청 body를 읽는다고 말하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `socket fd, local endpoint binding, listening queue, connected socket fd` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `socket fd, local endpoint binding, listening queue, connected socket fd`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `socket fd, local endpoint binding, listening queue, connected socket fd`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. ``accept`가 TCP handshake를 직접 수행한다고 말하거나`listen`이 요청 body를 읽는다고 말하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

`bind`는 이름 붙이기, `listen`은 새 연결 입구 만들기, `accept`는 완성 연결을 사용자 공간으로 꺼내는 단계입니다.
이 문장을 실제 서버로 옮기면 먼저 "Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 socket fd, local endpoint binding, listening queue, connected socket fd이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `socket fd, local endpoint binding, listening queue, connected socket fd`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `socket fd, local endpoint binding, listening queue, connected socket fd`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `socket fd, local endpoint binding, listening queue, connected socket fd`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`socket fd, local endpoint binding, listening queue, connected socket fd`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `socket fd, local endpoint binding, listening queue, connected socket fd`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -f -e socket,bind,listen,accept4 -p <pid>
ss -ltnp
cat /proc/sys/net/core/somaxconn
```

`socket fd, local endpoint binding, listening queue, connected socket fd`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `socket fd, local endpoint binding, listening queue, connected socket fd`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 ``accept`가 TCP handshake를 직접 수행한다고 말하거나`listen`이 요청 body를 읽는다고 말하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
`accept`가 TCP handshake를 직접 수행한다고 말하거나 `listen`이 요청 body를 읽는다고 말하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
``bind`는 이름 붙이기,`listen`은 새 연결 입구 만들기,`accept`는 완성 연결을 사용자 공간으로 꺼내는 단계입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `socket fd, local endpoint binding, listening queue, connected socket fd`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `socket fd, local endpoint binding, listening queue, connected socket fd`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `socket fd, local endpoint binding, listening queue, connected socket fd` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | ``accept`가 TCP handshake를 직접 수행한다고 말하거나`listen`이 요청 body를 읽는다고 말하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`socket fd, local endpoint binding, listening queue, connected socket fd`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`socket fd, local endpoint binding, listening queue, connected socket fd`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "`bind`는 이름 붙이기, `listen`은 새 연결 입구 만들기, `accept`는 완성 연결을 사용자 공간으로 꺼내는 단계입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: socket fd, local endpoint binding, listening queue, connected socket fd
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `socket fd, local endpoint binding, listening queue, connected socket fd`를 먼저 말하고, 그다음 ``bind`는 이름 붙이기, `listen`은 새 연결 입구 만들기, `accept`는 완성 연결을 사용자 공간으로 꺼내는 단계입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면``accept`가 TCP handshake를 직접 수행한다고 말하거나`listen`이 요청 body를 읽는다고 말하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 socket fd, local endpoint binding, listening queue, connected socket fd이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`socket fd, local endpoint binding, listening queue, connected socket fd` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `strace -f -e socket,bind,listen,accept4 -p <pid>` 같은 확인은 `socket fd, local endpoint binding, listening queue, connected socket fd` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 ``accept`가 TCP handshake를 직접 수행한다고 말하거나`listen`이 요청 body를 읽는다고 말하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 socket fd, local endpoint binding, listening queue, connected socket fd 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Nginx worker가 443 포트에서 listen하고 Tomcat connector가 8080 포트에서 accept하는 장면`에서는 socket fd, local endpoint binding, listening queue, connected socket fd 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `accept()`가 TCP handshake를 직접 수행한다거나 `listen()`이 요청 body를 읽는다는 식으로 말하면 부족합니다. `bind()`는 주소를 붙이고, `listen()`은 passive open 상태와 큐를 준비하며, `accept()`는 이미 성립한 연결을 user-space 서버가 사용할 fd로 꺼내는 단계이기 때문입니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `socket fd, local endpoint binding, listening queue, connected socket fd`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `socket fd, local endpoint binding, listening queue, connected socket fd`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`socket fd, local endpoint binding, listening queue, connected socket fd`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `socket fd, local endpoint binding, listening queue, connected socket fd` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### TCP 3-way handshake, backlog, accept queue

3-way handshake는 커널 TCP 스택의 연결 상태를 맞추는 절차이고, backlog와 accept queue는 그 연결을 애플리케이션이 가져가기 전까지 보관하는 완충 지점입니다. 예를 들어 부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 SYN, SYN-ACK, ACK, SYN backlog, accept queue entry를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 SYN, SYN-ACK, ACK, SYN backlog, accept queue entry이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: SYN, SYN-ACK, ACK, SYN backlog, accept queue entry
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "backlog를 HTTP request queue로만 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `backlog를 HTTP request queue로만 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

3-way handshake는 커널 TCP 스택의 연결 상태를 맞추는 절차이고, backlog와 accept queue는 그 연결을 애플리케이션이 가져가기 전까지 보관하는 완충 지점입니다.
이 문장을 실제 서버로 옮기면 먼저 "부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 SYN, SYN-ACK, ACK, SYN backlog, accept queue entry이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -ant state syn-recv
ss -ltn sport = :8080
sysctl net.ipv4.tcp_max_syn_backlog net.core.somaxconn
netstat -s | grep -i listen
```

`SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `backlog를 HTTP request queue로만 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
backlog를 HTTP request queue로만 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`3-way handshake는 커널 TCP 스택의 연결 상태를 맞추는 절차이고, backlog와 accept queue는 그 연결을 애플리케이션이 가져가기 전까지 보관하는 완충 지점입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `backlog를 HTTP request queue로만 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "3-way handshake는 커널 TCP 스택의 연결 상태를 맞추는 절차이고, backlog와 accept queue는 그 연결을 애플리케이션이 가져가기 전까지 보관하는 완충 지점입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: SYN, SYN-ACK, ACK, SYN backlog, accept queue entry
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 먼저 말하고, 그다음 `3-way handshake는 커널 TCP 스택의 연결 상태를 맞추는 절차이고, backlog와 accept queue는 그 연결을 애플리케이션이 가져가기 전까지 보관하는 완충 지점입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `backlog를 HTTP request queue로만 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`SYN, SYN-ACK, ACK, SYN backlog, accept queue entry` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -ant state syn-recv` 같은 확인은 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `backlog를 HTTP request queue로만 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 SYN, SYN-ACK, ACK, SYN backlog, accept queue entry 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서는 SYN, SYN-ACK, ACK, SYN backlog, accept queue entry 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `backlog를 HTTP request queue로만 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`부하 순간 새 연결이 몰리고 일부 클라이언트가 connect timeout 또는 refused를 보는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 SYN, SYN-ACK, ACK, SYN backlog, accept queue entry이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`SYN, SYN-ACK, ACK, SYN backlog, accept queue entry`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `SYN, SYN-ACK, ACK, SYN backlog, accept queue entry` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### TCP 상태: ESTABLISHED, CLOSE_WAIT, TIME_WAIT

TCP 상태는 커널이 각 연결의 생애를 추적하는 이름이고, 상태마다 누가 다음 close를 해야 하는지가 다릅니다. 예를 들어 운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 TCP state machine entry, FIN, ACK, socket close ownership를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 TCP state machine entry, FIN, ACK, socket close ownership이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: TCP state machine entry, FIN, ACK, socket close ownership
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `TCP state machine entry, FIN, ACK, socket close ownership`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `TCP state machine entry, FIN, ACK, socket close ownership` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `TCP state machine entry, FIN, ACK, socket close ownership`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `TCP state machine entry, FIN, ACK, socket close ownership`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

TCP 상태는 커널이 각 연결의 생애를 추적하는 이름이고, 상태마다 누가 다음 close를 해야 하는지가 다릅니다.
이 문장을 실제 서버로 옮기면 먼저 "운영 서버에서 `ss -ant` 결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 TCP state machine entry, FIN, ACK, socket close ownership이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `TCP state machine entry, FIN, ACK, socket close ownership`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `TCP state machine entry, FIN, ACK, socket close ownership`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `TCP state machine entry, FIN, ACK, socket close ownership`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`TCP state machine entry, FIN, ACK, socket close ownership`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `TCP state machine entry, FIN, ACK, socket close ownership`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -ant state established,close-wait,time-wait | head
ss -ant state close-wait '( sport = :8080 )'
lsof -p <pid> | grep TCP
```

`TCP state machine entry, FIN, ACK, socket close ownership`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `TCP state machine entry, FIN, ACK, socket close ownership`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`TCP 상태는 커널이 각 연결의 생애를 추적하는 이름이고, 상태마다 누가 다음 close를 해야 하는지가 다릅니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `TCP state machine entry, FIN, ACK, socket close ownership`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `TCP state machine entry, FIN, ACK, socket close ownership`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `TCP state machine entry, FIN, ACK, socket close ownership` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`TCP state machine entry, FIN, ACK, socket close ownership`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`TCP state machine entry, FIN, ACK, socket close ownership`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "TCP 상태는 커널이 각 연결의 생애를 추적하는 이름이고, 상태마다 누가 다음 close를 해야 하는지가 다릅니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: TCP state machine entry, FIN, ACK, socket close ownership
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `TCP state machine entry, FIN, ACK, socket close ownership`를 먼저 말하고, 그다음 `TCP 상태는 커널이 각 연결의 생애를 추적하는 이름이고, 상태마다 누가 다음 close를 해야 하는지가 다릅니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -ant state established,close-wait,time-wait | head` 같은 확인은 `TCP state machine entry, FIN, ACK, socket close ownership` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 TCP state machine entry, FIN, ACK, socket close ownership 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서는 TCP state machine entry, FIN, ACK, socket close ownership 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `TIME_WAIT가 보이면 무조건 장애라고 보거나 CLOSE_WAIT를 네트워크 장비 탓으로 돌리는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 TCP state machine entry, FIN, ACK, socket close ownership이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`TCP state machine entry, FIN, ACK, socket close ownership` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`운영 서버에서`ss -ant`결과 CLOSE_WAIT 또는 TIME_WAIT가 많이 보이는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `TCP state machine entry, FIN, ACK, socket close ownership`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `TCP state machine entry, FIN, ACK, socket close ownership`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`TCP state machine entry, FIN, ACK, socket close ownership`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `TCP state machine entry, FIN, ACK, socket close ownership` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### TLS와 HTTP keep-alive

keep-alive는 HTTP 요청을 하나 더 보내기 위해 이미 만든 TCP/TLS 연결을 살려 두는 운영 선택이고, handshake 비용과 TIME_WAIT 부담을 줄입니다. 예를 들어 클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

keep-alive는 HTTP 요청을 하나 더 보내기 위해 이미 만든 TCP/TLS 연결을 살려 두는 운영 선택이고, handshake 비용과 TIME_WAIT 부담을 줄입니다.
이 문장을 실제 서버로 옮기면 먼저 "클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
curl -v --http1.1 https://api.example.com/a https://api.example.com/b
openssl s_client -connect api.example.com:443 -servername api.example.com
ss -antp | grep ESTAB
```

`TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`keep-alive는 HTTP 요청을 하나 더 보내기 위해 이미 만든 TCP/TLS 연결을 살려 두는 운영 선택이고, handshake 비용과 TIME_WAIT 부담을 줄입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "keep-alive는 HTTP 요청을 하나 더 보내기 위해 이미 만든 TCP/TLS 연결을 살려 두는 운영 선택이고, handshake 비용과 TIME_WAIT 부담을 줄입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 먼저 말하고, 그다음 `keep-alive는 HTTP 요청을 하나 더 보내기 위해 이미 만든 TCP/TLS 연결을 살려 두는 운영 선택이고, handshake 비용과 TIME_WAIT 부담을 줄입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서는 TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`클라이언트가 같은 서버에 여러 API를 호출하면서 TCP/TLS 연결을 재사용하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `curl -v --http1.1 https://api.example.com/a https://api.example.com/b` 같은 확인은 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `TLS handshake와 TCP 4-way close를 같은 종류의 handshake로 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `TCP connection, TLS session, HTTP request/response pair, idle keep-alive socket` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### Nginx, L4/L7 load balancer, reverse proxy

L4는 보통 연결과 주소/포트 수준에서 분산하고, L7 reverse proxy는 HTTP 메시지를 이해해 header, path, buffering, upstream 정책을 바꿉니다. 예를 들어 ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

L4는 보통 연결과 주소/포트 수준에서 분산하고, L7 reverse proxy는 HTTP 메시지를 이해해 header, path, buffering, upstream 정책을 바꿉니다.
이 문장을 실제 서버로 옮기면 먼저 "ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
nginx -T | grep -E 'proxy_pass|proxy_buffering|upstream|keepalive'
curl -v -H 'Host: example.com' http://127.0.0.1/
ss -antp | grep nginx
```

`client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`L4는 보통 연결과 주소/포트 수준에서 분산하고, L7 reverse proxy는 HTTP 메시지를 이해해 header, path, buffering, upstream 정책을 바꿉니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "L4는 보통 연결과 주소/포트 수준에서 분산하고, L7 reverse proxy는 HTTP 메시지를 이해해 header, path, buffering, upstream 정책을 바꿉니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 먼저 말하고, 그다음 `L4는 보통 연결과 주소/포트 수준에서 분산하고, L7 reverse proxy는 HTTP 메시지를 이해해 header, path, buffering, upstream 정책을 바꿉니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서는 client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `nginx -T | grep -E 'proxy_pass|proxy_buffering|upstream|keepalive'` 같은 확인은 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `로드밸런서와 reverse proxy를 모두 단순 port forward로 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`ALB가 TLS를 종료하고 Nginx가 upstream Tomcat으로 요청을 보내는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `client TCP 연결, proxy HTTP request, upstream TCP 연결, response buffering` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### 패킷이 NIC에서 애플리케이션 버퍼까지 가는 길

NIC는 HTTP를 모르고, 커널은 TCP byte stream을 복원하며, 애플리케이션은 read 계열 호출로 가능한 byte를 가져옵니다. 예를 들어 NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

NIC는 HTTP를 모르고, 커널은 TCP byte stream을 복원하며, 애플리케이션은 read 계열 호출로 가능한 byte를 가져옵니다.
이 문장을 실제 서버로 옮기면 먼저 "NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ethtool -S eth0 | head
sar -n DEV,TCP,ETCP 1
tcpdump -i eth0 -nn tcp port 8080
strace -e read,recvfrom -p <pid>
```

`Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`NIC는 HTTP를 모르고, 커널은 TCP byte stream을 복원하며, 애플리케이션은 read 계열 호출로 가능한 byte를 가져옵니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "NIC는 HTTP를 모르고, 커널은 TCP byte stream을 복원하며, 애플리케이션은 read 계열 호출로 가능한 byte를 가져옵니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 먼저 말하고, 그다음 `NIC는 HTTP를 모르고, 커널은 TCP byte stream을 복원하며, 애플리케이션은 read 계열 호출로 가능한 byte를 가져옵니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ethtool -S eth0 | head` 같은 확인은 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `NIC가 받은 frame이 커널 네트워크 스택을 지나 Java 프로세스의 buffer로 복사되는 장면`에서는 Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `packet 하나가 Java 객체 하나로 바로 바뀐다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `Ethernet frame, sk_buff, TCP segment, socket receive buffer, ByteBuffer` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### 요청 수신 경로와 응답 송신 경로는 커널 안에서 어떻게 다르게 보이는가

수신은 NIC에서 socket receive buffer로 쌓이고 애플리케이션이 당겨 가는 흐름이며, 송신은 애플리케이션이 send buffer에 밀어 넣고 커널이 전송 가능할 때 내보내는 흐름입니다. 예를 들어 클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 receive queue, user read, user write, send queue, congestion/window update를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 receive queue, user read, user write, send queue, congestion/window update이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: receive queue, user read, user write, send queue, congestion/window update
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `receive queue, user read, user write, send queue, congestion/window update`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "read와 write가 같은 큐 하나를 공유한다고 보는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `receive queue, user read, user write, send queue, congestion/window update` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `receive queue, user read, user write, send queue, congestion/window update`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `receive queue, user read, user write, send queue, congestion/window update`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `read와 write가 같은 큐 하나를 공유한다고 보는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

수신은 NIC에서 socket receive buffer로 쌓이고 애플리케이션이 당겨 가는 흐름이며, 송신은 애플리케이션이 send buffer에 밀어 넣고 커널이 전송 가능할 때 내보내는 흐름입니다.
이 문장을 실제 서버로 옮기면 먼저 "클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 receive queue, user read, user write, send queue, congestion/window update이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `receive queue, user read, user write, send queue, congestion/window update`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `receive queue, user read, user write, send queue, congestion/window update`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `receive queue, user read, user write, send queue, congestion/window update`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`receive queue, user read, user write, send queue, congestion/window update`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `receive queue, user read, user write, send queue, congestion/window update`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -tinp '( sport = :8080 )'
cat /proc/net/tcp | head
tcpdump -i lo -nn tcp port 8080
```

`receive queue, user read, user write, send queue, congestion/window update`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `receive queue, user read, user write, send queue, congestion/window update`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `read와 write가 같은 큐 하나를 공유한다고 보는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
read와 write가 같은 큐 하나를 공유한다고 보는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`수신은 NIC에서 socket receive buffer로 쌓이고 애플리케이션이 당겨 가는 흐름이며, 송신은 애플리케이션이 send buffer에 밀어 넣고 커널이 전송 가능할 때 내보내는 흐름입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `receive queue, user read, user write, send queue, congestion/window update`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `receive queue, user read, user write, send queue, congestion/window update`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `receive queue, user read, user write, send queue, congestion/window update` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `read와 write가 같은 큐 하나를 공유한다고 보는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`receive queue, user read, user write, send queue, congestion/window update`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`receive queue, user read, user write, send queue, congestion/window update`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "수신은 NIC에서 socket receive buffer로 쌓이고 애플리케이션이 당겨 가는 흐름이며, 송신은 애플리케이션이 send buffer에 밀어 넣고 커널이 전송 가능할 때 내보내는 흐름입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: receive queue, user read, user write, send queue, congestion/window update
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `receive queue, user read, user write, send queue, congestion/window update`를 먼저 말하고, 그다음 `수신은 NIC에서 socket receive buffer로 쌓이고 애플리케이션이 당겨 가는 흐름이며, 송신은 애플리케이션이 send buffer에 밀어 넣고 커널이 전송 가능할 때 내보내는 흐름입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `read와 write가 같은 큐 하나를 공유한다고 보는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`receive queue, user read, user write, send queue, congestion/window update` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -tinp '( sport = :8080 )'` 같은 확인은 `receive queue, user read, user write, send queue, congestion/window update` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `read와 write가 같은 큐 하나를 공유한다고 보는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 receive queue, user read, user write, send queue, congestion/window update 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서는 receive queue, user read, user write, send queue, congestion/window update 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `read와 write가 같은 큐 하나를 공유한다고 보는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`클라이언트 upload는 느린데 서버 response도 동시에 나가야 하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 receive queue, user read, user write, send queue, congestion/window update이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `receive queue, user read, user write, send queue, congestion/window update`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `receive queue, user read, user write, send queue, congestion/window update`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`receive queue, user read, user write, send queue, congestion/window update`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `receive queue, user read, user write, send queue, congestion/window update` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

### 방화벽, security group, routing 문제는 connection refused와 timeout을 어떻게 갈라놓는가

refused는 상대 커널이나 중간 장치가 거절 신호를 돌려준 쪽에 가깝고, timeout은 응답이 오지 않아 재시도 끝에 포기하는 쪽에 가깝습니다. 예를 들어 배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 SYN, RST, ICMP, dropped packet, route lookup result를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

네트워크 실행 경로는 프로토콜 이름을 외우는 문서가 아니라, 커널이 어떤 상태를 만들고 어떤 큐에 무엇을 놓는지 따라가는 문서여야 합니다.
여기서 먼저 붙잡을 대상은 SYN, RST, ICMP, dropped packet, route lookup result이며, 이 대상은 TCP/IP stack, routing table, NAT table, socket buffer, proxy upstream 연결 사이를 지납니다.
각 계층은 자기보다 위의 HTTP 의미를 모른 채 자기 계약만 지킨습니다.
서버 운영자는 이 분리를 알아야 `connection refused`, `timeout`, `CLOSE_WAIT`, `TIME_WAIT`, `SSL handshake failure`, `502`, `504` 같은 증상을 보고 어느 계층부터 확인할지 정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: SYN, RST, ICMP, dropped packet, route lookup result
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `SYN, RST, ICMP, dropped packet, route lookup result`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `SYN, RST, ICMP, dropped packet, route lookup result` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 네트워크 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `SYN, RST, ICMP, dropped packet, route lookup result`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 네트워크 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `SYN, RST, ICMP, dropped packet, route lookup result`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

refused는 상대 커널이나 중간 장치가 거절 신호를 돌려준 쪽에 가깝고, timeout은 응답이 오지 않아 재시도 끝에 포기하는 쪽에 가깝습니다.
이 문장을 실제 서버로 옮기면 먼저 "배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 SYN, RST, ICMP, dropped packet, route lookup result이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
네트워크 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `SYN, RST, ICMP, dropped packet, route lookup result`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `SYN, RST, ICMP, dropped packet, route lookup result`의 소유권을 확인하는 근거로 사용합니다.
네트워크 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `SYN, RST, ICMP, dropped packet, route lookup result`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 네트워크 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`SYN, RST, ICMP, dropped packet, route lookup result`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `SYN, RST, ICMP, dropped packet, route lookup result`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
네트워크 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

네트워크 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
nc -vz host 8080
curl -v --connect-timeout 3 http://host:8080
ip route get <ip>
sudo tcpdump -nn host <ip> and tcp
```

`SYN, RST, ICMP, dropped packet, route lookup result`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 네트워크 경로를 확인할 때는 `ss`, `ip route`, `tcpdump`, Nginx upstream timing, 방화벽 로그를 같은 시간대에 놓고 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `SYN, RST, ICMP, dropped packet, route lookup result`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`refused는 상대 커널이나 중간 장치가 거절 신호를 돌려준 쪽에 가깝고, timeout은 응답이 오지 않아 재시도 끝에 포기하는 쪽에 가깝습니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `SYN, RST, ICMP, dropped packet, route lookup result`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `SYN, RST, ICMP, dropped packet, route lookup result`를 단일 로그에서만 찾는 것입니다.
네트워크 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 네트워크 경로의 커널 상태 vs 애플리케이션 상태 | `SYN, RST, ICMP, dropped packet, route lookup result` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 네트워크 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 네트워크 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 네트워크 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 네트워크 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`SYN, RST, ICMP, dropped packet, route lookup result`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 네트워크 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`SYN, RST, ICMP, dropped packet, route lookup result`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "refused는 상대 커널이나 중간 장치가 거절 신호를 돌려준 쪽에 가깝고, timeout은 응답이 오지 않아 재시도 끝에 포기하는 쪽에 가깝습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: SYN, RST, ICMP, dropped packet, route lookup result
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `SYN, RST, ICMP, dropped packet, route lookup result`를 먼저 말하고, 그다음 `refused는 상대 커널이나 중간 장치가 거절 신호를 돌려준 쪽에 가깝고, timeout은 응답이 오지 않아 재시도 끝에 포기하는 쪽에 가깝습니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

네트워크 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `nc -vz host 8080` 같은 확인은 `SYN, RST, ICMP, dropped packet, route lookup result` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 SYN, RST, ICMP, dropped packet, route lookup result 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서는 SYN, RST, ICMP, dropped packet, route lookup result 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `connection refused와 timeout을 같은 네트워크 불안정으로 묶는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 SYN, RST, ICMP, dropped packet, route lookup result이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

네트워크 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`SYN, RST, ICMP, dropped packet, route lookup result` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`배포 후 새 서버 포트가 열리지 않아 어떤 클라이언트는 refused, 어떤 클라이언트는 timeout을 보는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

#### 운영 질문으로 다시 압축하기

네트워크 경로를 실제 면접 답변으로 꺼낼 때는 먼저 네트워크 경로에서 packet, HTTP 메시지, upstream 요청을 어떻게 나눌지 말해야 합니다. 핵심은 packet과 HTTP 메시지와 upstream 요청을 한 단어로 합치지 않는 것입니다. 예를 들어 `SYN, RST, ICMP, dropped packet, route lookup result`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 네트워크 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `SYN, RST, ICMP, dropped packet, route lookup result`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`SYN, RST, ICMP, dropped packet, route lookup result`의 관측 근거는 `ss`, `ip route`, `tcpdump`, proxy log, firewall log를 같은 시간대에 맞추는 것에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `SYN, RST, ICMP, dropped packet, route lookup result` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 네트워크 경로의 세부 메커니즘으로 내려가면 됩니다.

## I/O multiplexing과 커널 이벤트 모델

이 장은 많은 연결을 적은 스레드로 다루기 위해 서버가 커널 이벤트 모델을 어떻게 사용하는지 설명합니다. readiness와 completion의 차이는 Netty, Reactor, WebFlux를 이해하는 바닥입니다.

### select와 poll은 왜 한계가 생겼는가

select와 poll은 여러 fd를 기다릴 수 있게 했지만, 현대적인 대규모 연결 서버에서는 fd 집합 복사와 전체 scan, API 한계가 비용이 됩니다. 예를 들어 연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 fd_set, pollfd array, watched descriptor list를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

이벤트 모델을 이해할 때 가장 먼저 버려야 할 오해는 이벤트 API가 데이터를 대신 읽어 준다는 생각입니다.
여기서 먼저 붙잡을 대상은 fd_set, pollfd array, watched descriptor list이며, 이 대상은 준비 상태 또는 완료 상태를 알려 주는 신호이지 애플리케이션 프로토콜 경계를 대신 판단하는 값이 아닙니다.
readiness 기반 모델에서는 커널이 "지금 읽거나 쓸 수 있다"고 알려 주고, 애플리케이션이 실제 read/write를 수행합니다.
completion 기반 모델에서는 애플리케이션이 "이 read/write를 해 달라"고 작업을 제출하고, 나중에 완료 결과를 받습니다.
두 모델의 차이는 API 이름보다 소유권 차이에 있습니다.
누가 작업을 시작하고, 누가 반복하며, 누가 완료와 오류를 책임지는지가 달라집니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: fd_set, pollfd array, watched descriptor list
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `fd_set, pollfd array, watched descriptor list`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `fd_set, pollfd array, watched descriptor list` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 커널 이벤트 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `fd_set, pollfd array, watched descriptor list`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 커널 이벤트 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `fd_set, pollfd array, watched descriptor list`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

select와 poll은 여러 fd를 기다릴 수 있게 했지만, 현대적인 대규모 연결 서버에서는 fd 집합 복사와 전체 scan, API 한계가 비용이 됩니다.
이 문장을 실제 서버로 옮기면 먼저 "연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 fd_set, pollfd array, watched descriptor list이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
커널 이벤트 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `fd_set, pollfd array, watched descriptor list`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux select(2)](https://man7.org/linux/man-pages/man2/select.2.html), [Linux poll(2)](https://man7.org/linux/man-pages/man2/poll.2.html), [Linux epoll(7)](https://www.man7.org/linux/man-pages/man7/epoll.7.html), [Linux io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html), [FreeBSD kqueue(2)](https://man.freebsd.org/kqueue%282%29)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `fd_set, pollfd array, watched descriptor list`의 소유권을 확인하는 근거로 사용합니다.
커널 이벤트 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `fd_set, pollfd array, watched descriptor list`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 커널 이벤트 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`fd_set, pollfd array, watched descriptor list`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `fd_set, pollfd array, watched descriptor list`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
커널 이벤트 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

커널 이벤트 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
man 2 select
man 2 poll
strace -e select,poll -p <pid>
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 커널 이벤트 모델을 확인할 때는 `strace`로 `select`, `poll`, `epoll_wait`, `io_uring_enter` 호출을 보고, fd 수가 늘 때 대기 비용이 어디서 발생하는지 함께 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `fd_set, pollfd array, watched descriptor list`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`select와 poll은 여러 fd를 기다릴 수 있게 했지만, 현대적인 대규모 연결 서버에서는 fd 집합 복사와 전체 scan, API 한계가 비용이 됩니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `fd_set, pollfd array, watched descriptor list`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
커널 이벤트 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 커널 이벤트 모델의 커널 상태 vs 애플리케이션 상태 | `fd_set, pollfd array, watched descriptor list` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 커널 이벤트 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 커널 이벤트 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 커널 이벤트 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 커널 이벤트 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 커널 이벤트 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`fd_set, pollfd array, watched descriptor list`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "select와 poll은 여러 fd를 기다릴 수 있게 했지만, 현대적인 대규모 연결 서버에서는 fd 집합 복사와 전체 scan, API 한계가 비용이 됩니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: fd_set, pollfd array, watched descriptor list
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `fd_set, pollfd array, watched descriptor list`를 먼저 말하고, 그다음 `select와 poll은 여러 fd를 기다릴 수 있게 했지만, 현대적인 대규모 연결 서버에서는 fd 집합 복사와 전체 scan, API 한계가 비용이 됩니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 fd_set, pollfd array, watched descriptor list 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서는 fd_set, pollfd array, watched descriptor list 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 fd_set, pollfd array, watched descriptor list이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

커널 이벤트 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`fd_set, pollfd array, watched descriptor list` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`연결 수가 수만 개로 늘어났는데 매 loop마다 관심 fd 전체를 다시 훑는 서버 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

커널 이벤트 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `man 2 select` 같은 확인은 `fd_set, pollfd array, watched descriptor list` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `select/poll이 느린 이유를 단순히 오래된 API라서라고만 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

#### 운영 질문으로 다시 압축하기

커널 이벤트 모델을 실제 면접 답변으로 꺼낼 때는 먼저 커널 이벤트 모델에서 readiness 신호와 실제 데이터 처리를 어떻게 나눌지 말해야 합니다. 핵심은 readiness 신호와 실제 데이터 처리 책임을 분리하는 것입니다. 예를 들어 `fd_set, pollfd array, watched descriptor list`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 커널 이벤트 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `fd_set, pollfd array, watched descriptor list`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 `strace`, fd 수, event-loop delay, readiness/completion 호출의 빈도와 대기 시간에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `fd_set, pollfd array, watched descriptor list` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 커널 이벤트 모델의 세부 메커니즘으로 내려가면 됩니다.

### epoll은 readiness를 어떤 방식으로 관리하는가

epoll은 관심 fd를 커널 안 epoll instance에 등록해 두고, 준비된 fd 목록을 기다리는 Linux 전용 readiness 모델입니다. 예를 들어 Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 epoll instance, interest list, ready list, epoll_wait result를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

이벤트 모델을 이해할 때 가장 먼저 버려야 할 오해는 이벤트 API가 데이터를 대신 읽어 준다는 생각입니다.
여기서 먼저 붙잡을 대상은 epoll instance, interest list, ready list, epoll_wait result이며, 이 대상은 준비 상태 또는 완료 상태를 알려 주는 신호이지 애플리케이션 프로토콜 경계를 대신 판단하는 값이 아닙니다.
readiness 기반 모델에서는 커널이 "지금 읽거나 쓸 수 있다"고 알려 주고, 애플리케이션이 실제 read/write를 수행합니다.
completion 기반 모델에서는 애플리케이션이 "이 read/write를 해 달라"고 작업을 제출하고, 나중에 완료 결과를 받습니다.
두 모델의 차이는 API 이름보다 소유권 차이에 있습니다.
누가 작업을 시작하고, 누가 반복하며, 누가 완료와 오류를 책임지는지가 달라집니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: epoll instance, interest list, ready list, epoll_wait result
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `epoll instance, interest list, ready list, epoll_wait result`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "epoll_wait가 byte를 읽어 준다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `epoll instance, interest list, ready list, epoll_wait result` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 커널 이벤트 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `epoll instance, interest list, ready list, epoll_wait result`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 커널 이벤트 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `epoll instance, interest list, ready list, epoll_wait result`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `epoll_wait가 byte를 읽어 준다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

epoll은 관심 fd를 커널 안 epoll instance에 등록해 두고, 준비된 fd 목록을 기다리는 Linux 전용 readiness 모델입니다.
이 문장을 실제 서버로 옮기면 먼저 "Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 epoll instance, interest list, ready list, epoll_wait result이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
커널 이벤트 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `epoll instance, interest list, ready list, epoll_wait result`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux select(2)](https://man7.org/linux/man-pages/man2/select.2.html), [Linux poll(2)](https://man7.org/linux/man-pages/man2/poll.2.html), [Linux epoll(7)](https://www.man7.org/linux/man-pages/man7/epoll.7.html), [Linux io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html), [FreeBSD kqueue(2)](https://man.freebsd.org/kqueue%282%29)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `epoll instance, interest list, ready list, epoll_wait result`의 소유권을 확인하는 근거로 사용합니다.
커널 이벤트 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `epoll instance, interest list, ready list, epoll_wait result`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 커널 이벤트 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`epoll instance, interest list, ready list, epoll_wait result`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `epoll instance, interest list, ready list, epoll_wait result`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
커널 이벤트 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

커널 이벤트 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -f -e epoll_create1,epoll_ctl,epoll_wait -p <pid>
ls -l /proc/<pid>/fd
cat /proc/<pid>/fdinfo/<epoll-fd>
```

`epoll instance, interest list, ready list, epoll_wait result`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 커널 이벤트 모델을 확인할 때는 `strace`로 `select`, `poll`, `epoll_wait`, `io_uring_enter` 호출을 보고, fd 수가 늘 때 대기 비용이 어디서 발생하는지 함께 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `epoll instance, interest list, ready list, epoll_wait result`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `epoll_wait가 byte를 읽어 준다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
epoll_wait가 byte를 읽어 준다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`epoll은 관심 fd를 커널 안 epoll instance에 등록해 두고, 준비된 fd 목록을 기다리는 Linux 전용 readiness 모델입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `epoll instance, interest list, ready list, epoll_wait result`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `epoll instance, interest list, ready list, epoll_wait result`를 단일 로그에서만 찾는 것입니다.
커널 이벤트 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 커널 이벤트 모델의 커널 상태 vs 애플리케이션 상태 | `epoll instance, interest list, ready list, epoll_wait result` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 커널 이벤트 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 커널 이벤트 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 커널 이벤트 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 커널 이벤트 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `epoll_wait가 byte를 읽어 준다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`epoll instance, interest list, ready list, epoll_wait result`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 커널 이벤트 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`epoll instance, interest list, ready list, epoll_wait result`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "epoll은 관심 fd를 커널 안 epoll instance에 등록해 두고, 준비된 fd 목록을 기다리는 Linux 전용 readiness 모델입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: epoll instance, interest list, ready list, epoll_wait result
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `epoll instance, interest list, ready list, epoll_wait result`를 먼저 말하고, 그다음 `epoll은 관심 fd를 커널 안 epoll instance에 등록해 두고, 준비된 fd 목록을 기다리는 Linux 전용 readiness 모델입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `epoll_wait가 byte를 읽어 준다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서는 epoll instance, interest list, ready list, epoll_wait result 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `epoll_wait가 byte를 읽어 준다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 epoll instance, interest list, ready list, epoll_wait result이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

커널 이벤트 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`epoll instance, interest list, ready list, epoll_wait result` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

커널 이벤트 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `strace -f -e epoll_create1,epoll_ctl,epoll_wait -p <pid>` 같은 확인은 `epoll instance, interest list, ready list, epoll_wait result` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `epoll_wait가 byte를 읽어 준다고 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Nginx worker나 Netty event loop가 많은 socket fd 중 읽을 수 있는 fd만 받아 처리하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 epoll instance, interest list, ready list, epoll_wait result 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

#### 운영 질문으로 다시 압축하기

커널 이벤트 모델을 실제 면접 답변으로 꺼낼 때는 먼저 커널 이벤트 모델에서 readiness 신호와 실제 데이터 처리를 어떻게 나눌지 말해야 합니다. 핵심은 readiness 신호와 실제 데이터 처리 책임을 분리하는 것입니다. 예를 들어 `epoll instance, interest list, ready list, epoll_wait result`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 커널 이벤트 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `epoll instance, interest list, ready list, epoll_wait result`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`epoll instance, interest list, ready list, epoll_wait result`의 관측 근거는 `strace`, fd 수, event-loop delay, readiness/completion 호출의 빈도와 대기 시간에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `epoll instance, interest list, ready list, epoll_wait result` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 커널 이벤트 모델의 세부 메커니즘으로 내려가면 됩니다.

### kqueue는 epoll과 어떤 관점에서 비교해야 하는가

kqueue는 BSD 계열의 kernel event queue이고, socket뿐 아니라 vnode, signal, process 같은 다양한 filter를 같은 이벤트 큐에 올릴 수 있습니다. 예를 들어 macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 kqueue fd, kevent filter, EVFILT_READ/WRITE event를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

이벤트 모델을 이해할 때 가장 먼저 버려야 할 오해는 이벤트 API가 데이터를 대신 읽어 준다는 생각입니다.
여기서 먼저 붙잡을 대상은 kqueue fd, kevent filter, EVFILT_READ/WRITE event이며, 이 대상은 준비 상태 또는 완료 상태를 알려 주는 신호이지 애플리케이션 프로토콜 경계를 대신 판단하는 값이 아닙니다.
readiness 기반 모델에서는 커널이 "지금 읽거나 쓸 수 있다"고 알려 주고, 애플리케이션이 실제 read/write를 수행합니다.
completion 기반 모델에서는 애플리케이션이 "이 read/write를 해 달라"고 작업을 제출하고, 나중에 완료 결과를 받습니다.
두 모델의 차이는 API 이름보다 소유권 차이에 있습니다.
누가 작업을 시작하고, 누가 반복하며, 누가 완료와 오류를 책임지는지가 달라집니다.

#### 값이 움직이는 순서

여기서는 Linux 기본 경로가 아니라 BSD/macOS 비교 경로를 따로 떼어 봅니다.
Linux 서버의 주 경로는 epoll이고, kqueue는 FreeBSD와 macOS 계열에서 같은 문제를 다른 커널 이벤트 큐로 푸는 방식입니다.

```text
macOS 또는 BSD process
    |
    | 1. kqueue fd를 만들고 kevent filter를 등록합니다
    v
BSD/macOS kernel event queue
    |
    | 2. socket, vnode, signal, process 같은 대상에서 이벤트가 생기면 queue에 올립니다
    v
user-space runtime
    |
    | 3. kevent 호출 결과를 Java NIO provider나 Netty native transport가 읽습니다
    v
application protocol layer
    |
    | 4. runtime은 read/write 가능 상태를 channel event나 handler 호출로 바꿉니다
```

이 trace를 Linux trace와 섞으면 안 됩니다.
Linux에서 kqueue fd나 kevent filter를 찾는 것은 잘못된 관측이고, macOS에서 `/proc/<pid>/fdinfo`의 epoll 항목을 찾는 것도 잘못된 관측입니다.
이 항목의 목적은 kqueue를 Linux 기능처럼 소개하는 것이 아니라, Java NIO와 Netty가 운영체제별 커널 이벤트 모델 위에 올라갈 때 어떤 부분이 공통 추상화이고 어떤 부분이 플랫폼 고유 구현인지 구분하는 것입니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `kqueue fd, kevent filter, EVFILT_READ/WRITE event` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 커널 이벤트 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 커널 이벤트 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Linux에서 kqueue 설정을 찾거나, macOS에서 epoll fdinfo를 찾는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

kqueue는 BSD 계열의 kernel event queue이고, socket뿐 아니라 vnode, signal, process 같은 다양한 filter를 같은 이벤트 큐에 올릴 수 있습니다.
이 문장을 실제 서버로 옮기면 먼저 "macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 kqueue fd, kevent filter, EVFILT_READ/WRITE event이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
커널 이벤트 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `kqueue fd, kevent filter, EVFILT_READ/WRITE event`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux select(2)](https://man7.org/linux/man-pages/man2/select.2.html), [Linux poll(2)](https://man7.org/linux/man-pages/man2/poll.2.html), [Linux epoll(7)](https://www.man7.org/linux/man-pages/man7/epoll.7.html), [Linux io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html), [FreeBSD kqueue(2)](https://man.freebsd.org/kqueue%282%29)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`의 소유권을 확인하는 근거로 사용합니다.
커널 이벤트 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 커널 이벤트 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`kqueue fd, kevent filter, EVFILT_READ/WRITE event`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
커널 이벤트 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

커널 이벤트 모델은 Linux 서버에서 kqueue를 켜라는 뜻이 아닙니다.
Linux에서는 epoll 계열 관측을 보고, BSD/macOS 비교가 필요할 때만 kqueue와 kevent를 확인합니다.
따라서 이 항목의 재현 명령은 Linux production 서버에 그대로 붙이는 명령이 아니라 macOS/BSD 개발 환경에서 플랫폼 차이를 확인하는 명령으로 읽어야 합니다.

```sh
uname -s
sudo dtruss -f -t kqueue -t kevent <cmd>
man 2 kqueue
```

확인할 수 있는 신호는 `kqueue` 또는 `kevent` 호출이 BSD/macOS 커널 이벤트 큐를 통해 준비 상태를 돌려주는 모습입니다.
Linux 환경에서는 같은 목적을 `epoll_wait`와 `/proc/<pid>/fdinfo/<epoll-fd>` 같은 Linux 전용 표면으로 확인합니다.
어긋나는 신호는 두 플랫폼의 관측 표면을 섞는 경우입니다.
Linux에서 kqueue 설정을 찾거나 macOS에서 epoll fdinfo를 찾는다면, Java NIO나 Netty의 cross-platform abstraction과 운영체제별 native transport를 혼동한 것입니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Linux에서 kqueue 설정을 찾거나, macOS에서 epoll fdinfo를 찾는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`kqueue는 BSD 계열의 kernel event queue이고, socket뿐 아니라 vnode, signal, process 같은 다양한 filter를 같은 이벤트 큐에 올릴 수 있습니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 단일 로그에서만 찾는 것입니다.
커널 이벤트 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 커널 이벤트 모델의 커널 상태 vs 애플리케이션 상태 | `kqueue fd, kevent filter, EVFILT_READ/WRITE event` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 커널 이벤트 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 커널 이벤트 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 커널 이벤트 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 커널 이벤트 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Linux에서 kqueue 설정을 찾거나, macOS에서 epoll fdinfo를 찾는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 커널 이벤트 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "kqueue는 BSD 계열의 kernel event queue이고, socket뿐 아니라 vnode, signal, process 같은 다양한 filter를 같은 이벤트 큐에 올릴 수 있습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: kqueue fd, kevent filter, EVFILT_READ/WRITE event
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 먼저 말하고, 그다음 `kqueue는 BSD 계열의 kernel event queue이고, socket뿐 아니라 vnode, signal, process 같은 다양한 filter를 같은 이벤트 큐에 올릴 수 있습니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Linux에서 kqueue 설정을 찾거나, macOS에서 epoll fdinfo를 찾는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 kqueue fd, kevent filter, EVFILT_READ/WRITE event이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

커널 이벤트 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`kqueue fd, kevent filter, EVFILT_READ/WRITE event` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

커널 이벤트 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `uname -s` 같은 확인은 `kqueue fd, kevent filter, EVFILT_READ/WRITE event` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `Linux에서 kqueue 설정을 찾거나, macOS에서 epoll fdinfo를 찾는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 kqueue fd, kevent filter, EVFILT_READ/WRITE event 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `macOS에서 Java NIO나 Netty native transport를 볼 때 epoll이 아니라 kqueue/kevent가 관측되는 장면`에서는 kqueue fd, kevent filter, EVFILT_READ/WRITE event 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Linux에서 kqueue 설정을 찾거나, macOS에서 epoll fdinfo를 찾는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

#### 운영 질문으로 다시 압축하기

커널 이벤트 모델을 실제 면접 답변으로 꺼낼 때는 먼저 커널 이벤트 모델에서 readiness 신호와 실제 데이터 처리를 어떻게 나눌지 말해야 합니다. 핵심은 readiness 신호와 실제 데이터 처리 책임을 분리하는 것입니다. 예를 들어 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 커널 이벤트 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`kqueue fd, kevent filter, EVFILT_READ/WRITE event`의 관측 근거는 `strace`, fd 수, event-loop delay, readiness/completion 호출의 빈도와 대기 시간에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `kqueue fd, kevent filter, EVFILT_READ/WRITE event` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 커널 이벤트 모델의 세부 메커니즘으로 내려가면 됩니다.

#### 작은 재현과 반례까지 닫기

커널 이벤트 모델은 작은 재현을 하나 만들어 보면 훨씬 오래 남습니다. 로컬에서 server와 client를 띄운 뒤 요청을 천천히 보내거나, 연결을 짧게 반복하거나, worker thread를 일부러 막거나, proxy buffering을 켜고 끄면 `kqueue fd, kevent filter, EVFILT_READ/WRITE event`가 어느 위치에서 달라지는지 볼 수 있습니다. 이 실험은 production 장애를 그대로 복제하려는 목적이 아닙니다. 각 계층이 어떤 이름의 상태를 남기는지 손으로 확인해 두면, 실제 장애에서 로그 문구 하나에 끌려가지 않습니다.

반례도 함께 붙여야 합니다. 커널 이벤트 모델을 설명할 때 가장 흔한 실패는 한 계층의 개선책이 모든 계층의 문제를 해결한다고 말하는 것입니다. fd 한도를 올려도 DB pool이 작으면 요청은 여전히 기다리고, event loop를 쓰더라도 blocking DB 호출을 같은 loop 위에서 실행하면 지연은 사라지지 않습니다. proxy buffering을 끄면 streaming 체감은 좋아질 수 있지만 느린 client가 만드는 send buffer 압박은 그대로 남을 수 있습니다. 그래서 설정 변경은 항상 "어떤 queue, 어떤 buffer, 어떤 pool, 어떤 timeout을 바꿉니까"라는 문장으로 끝나야 합니다.

### io_uring은 readiness가 아니라 completion을 어떻게 다루는가

io_uring은 fd가 준비됐는지 묻는 대신 수행할 I/O 작업을 제출하고 완료 항목을 회수하는 모델입니다. 예를 들어 애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 SQE, submission queue, kernel operation, CQE를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

이벤트 모델을 이해할 때 가장 먼저 버려야 할 오해는 이벤트 API가 데이터를 대신 읽어 준다는 생각입니다.
여기서 먼저 붙잡을 대상은 SQE, submission queue, kernel operation, CQE이며, 이 대상은 준비 상태 또는 완료 상태를 알려 주는 신호이지 애플리케이션 프로토콜 경계를 대신 판단하는 값이 아닙니다.
readiness 기반 모델에서는 커널이 "지금 읽거나 쓸 수 있다"고 알려 주고, 애플리케이션이 실제 read/write를 수행합니다.
completion 기반 모델에서는 애플리케이션이 "이 read/write를 해 달라"고 작업을 제출하고, 나중에 완료 결과를 받습니다.
두 모델의 차이는 API 이름보다 소유권 차이에 있습니다.
누가 작업을 시작하고, 누가 반복하며, 누가 완료와 오류를 책임지는지가 달라집니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: SQE, submission queue, kernel operation, CQE
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `SQE, submission queue, kernel operation, CQE`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "io_uring을 epoll의 빠른 버전으로만 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `SQE, submission queue, kernel operation, CQE` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 커널 이벤트 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `SQE, submission queue, kernel operation, CQE`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 커널 이벤트 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `SQE, submission queue, kernel operation, CQE`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `io_uring을 epoll의 빠른 버전으로만 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

io_uring은 fd가 준비됐는지 묻는 대신 수행할 I/O 작업을 제출하고 완료 항목을 회수하는 모델입니다.
이 문장을 실제 서버로 옮기면 먼저 "애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 SQE, submission queue, kernel operation, CQE이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
커널 이벤트 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `SQE, submission queue, kernel operation, CQE`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux select(2)](https://man7.org/linux/man-pages/man2/select.2.html), [Linux poll(2)](https://man7.org/linux/man-pages/man2/poll.2.html), [Linux epoll(7)](https://www.man7.org/linux/man-pages/man7/epoll.7.html), [Linux io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html), [FreeBSD kqueue(2)](https://man.freebsd.org/kqueue%282%29)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `SQE, submission queue, kernel operation, CQE`의 소유권을 확인하는 근거로 사용합니다.
커널 이벤트 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `SQE, submission queue, kernel operation, CQE`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 커널 이벤트 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`SQE, submission queue, kernel operation, CQE`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `SQE, submission queue, kernel operation, CQE`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
커널 이벤트 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

커널 이벤트 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -e io_uring_setup,io_uring_enter -p <pid>
grep -w io_uring /proc/kallsyms 2>/dev/null | head
uname -r
```

`SQE, submission queue, kernel operation, CQE`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 커널 이벤트 모델을 확인할 때는 `strace`로 `select`, `poll`, `epoll_wait`, `io_uring_enter` 호출을 보고, fd 수가 늘 때 대기 비용이 어디서 발생하는지 함께 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `SQE, submission queue, kernel operation, CQE`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `io_uring을 epoll의 빠른 버전으로만 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
io_uring을 epoll의 빠른 버전으로만 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`io_uring은 fd가 준비됐는지 묻는 대신 수행할 I/O 작업을 제출하고 완료 항목을 회수하는 모델입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `SQE, submission queue, kernel operation, CQE`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `SQE, submission queue, kernel operation, CQE`를 단일 로그에서만 찾는 것입니다.
커널 이벤트 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 커널 이벤트 모델의 커널 상태 vs 애플리케이션 상태 | `SQE, submission queue, kernel operation, CQE` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 커널 이벤트 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 커널 이벤트 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 커널 이벤트 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 커널 이벤트 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `io_uring을 epoll의 빠른 버전으로만 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`SQE, submission queue, kernel operation, CQE`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 커널 이벤트 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`SQE, submission queue, kernel operation, CQE`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "io_uring은 fd가 준비됐는지 묻는 대신 수행할 I/O 작업을 제출하고 완료 항목을 회수하는 모델입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: SQE, submission queue, kernel operation, CQE
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `SQE, submission queue, kernel operation, CQE`를 먼저 말하고, 그다음 `io_uring은 fd가 준비됐는지 묻는 대신 수행할 I/O 작업을 제출하고 완료 항목을 회수하는 모델입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `io_uring을 epoll의 빠른 버전으로만 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

커널 이벤트 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`SQE, submission queue, kernel operation, CQE` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

커널 이벤트 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `strace -e io_uring_setup,io_uring_enter -p <pid>` 같은 확인은 `SQE, submission queue, kernel operation, CQE` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `io_uring을 epoll의 빠른 버전으로만 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 SQE, submission queue, kernel operation, CQE 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서는 SQE, submission queue, kernel operation, CQE 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `io_uring을 epoll의 빠른 버전으로만 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`애플리케이션이 read 요청 자체를 큐에 제출하고 완료 결과를 나중에 받는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 SQE, submission queue, kernel operation, CQE이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

#### 운영 질문으로 다시 압축하기

커널 이벤트 모델을 실제 면접 답변으로 꺼낼 때는 먼저 커널 이벤트 모델에서 readiness 신호와 실제 데이터 처리를 어떻게 나눌지 말해야 합니다. 핵심은 readiness 신호와 실제 데이터 처리 책임을 분리하는 것입니다. 예를 들어 `SQE, submission queue, kernel operation, CQE`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 커널 이벤트 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `SQE, submission queue, kernel operation, CQE`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`SQE, submission queue, kernel operation, CQE`의 관측 근거는 `strace`, fd 수, event-loop delay, readiness/completion 호출의 빈도와 대기 시간에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `SQE, submission queue, kernel operation, CQE` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 커널 이벤트 모델의 세부 메커니즘으로 내려가면 됩니다.

### readiness 기반 모델과 completion 기반 모델은 서버 설계를 어떻게 바꾸는가

readiness 모델은 '지금 해도 되는가'를 받고 사용자가 직접 수행하며, completion 모델은 '이 일을 해 달라'고 맡기고 완료를 받습니다. 예를 들어 같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 ready fd, user read loop, submitted operation, completion result를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

이벤트 모델을 이해할 때 가장 먼저 버려야 할 오해는 이벤트 API가 데이터를 대신 읽어 준다는 생각입니다.
여기서 먼저 붙잡을 대상은 ready fd, user read loop, submitted operation, completion result이며, 이 대상은 준비 상태 또는 완료 상태를 알려 주는 신호이지 애플리케이션 프로토콜 경계를 대신 판단하는 값이 아닙니다.
readiness 기반 모델에서는 커널이 "지금 읽거나 쓸 수 있다"고 알려 주고, 애플리케이션이 실제 read/write를 수행합니다.
completion 기반 모델에서는 애플리케이션이 "이 read/write를 해 달라"고 작업을 제출하고, 나중에 완료 결과를 받습니다.
두 모델의 차이는 API 이름보다 소유권 차이에 있습니다.
누가 작업을 시작하고, 누가 반복하며, 누가 완료와 오류를 책임지는지가 달라집니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: ready fd, user read loop, submitted operation, completion result
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `ready fd, user read loop, submitted operation, completion result`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `ready fd, user read loop, submitted operation, completion result` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 커널 이벤트 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `ready fd, user read loop, submitted operation, completion result`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 커널 이벤트 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `ready fd, user read loop, submitted operation, completion result`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

readiness 모델은 '지금 해도 되는가'를 받고 사용자가 직접 수행하며, completion 모델은 '이 일을 해 달라'고 맡기고 완료를 받습니다.
이 문장을 실제 서버로 옮기면 먼저 "같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 ready fd, user read loop, submitted operation, completion result이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
커널 이벤트 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `ready fd, user read loop, submitted operation, completion result`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux select(2)](https://man7.org/linux/man-pages/man2/select.2.html), [Linux poll(2)](https://man7.org/linux/man-pages/man2/poll.2.html), [Linux epoll(7)](https://www.man7.org/linux/man-pages/man7/epoll.7.html), [Linux io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html), [FreeBSD kqueue(2)](https://man.freebsd.org/kqueue%282%29)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `ready fd, user read loop, submitted operation, completion result`의 소유권을 확인하는 근거로 사용합니다.
커널 이벤트 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `ready fd, user read loop, submitted operation, completion result`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 커널 이벤트 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`ready fd, user read loop, submitted operation, completion result`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `ready fd, user read loop, submitted operation, completion result`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
커널 이벤트 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

커널 이벤트 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -f -e epoll_wait,read,write,io_uring_enter -p <pid>
perf trace -p <pid>
```

`ready fd, user read loop, submitted operation, completion result`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 커널 이벤트 모델을 확인할 때는 `strace`로 `select`, `poll`, `epoll_wait`, `io_uring_enter` 호출을 보고, fd 수가 늘 때 대기 비용이 어디서 발생하는지 함께 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `ready fd, user read loop, submitted operation, completion result`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`readiness 모델은 '지금 해도 되는가'를 받고 사용자가 직접 수행하며, completion 모델은 '이 일을 해 달라'고 맡기고 완료를 받습니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `ready fd, user read loop, submitted operation, completion result`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `ready fd, user read loop, submitted operation, completion result`를 단일 로그에서만 찾는 것입니다.
커널 이벤트 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 커널 이벤트 모델의 커널 상태 vs 애플리케이션 상태 | `ready fd, user read loop, submitted operation, completion result` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 커널 이벤트 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 커널 이벤트 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 커널 이벤트 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 커널 이벤트 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`ready fd, user read loop, submitted operation, completion result`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 커널 이벤트 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`ready fd, user read loop, submitted operation, completion result`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "readiness 모델은 '지금 해도 되는가'를 받고 사용자가 직접 수행하며, completion 모델은 '이 일을 해 달라'고 맡기고 완료를 받습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: ready fd, user read loop, submitted operation, completion result
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `ready fd, user read loop, submitted operation, completion result`를 먼저 말하고, 그다음 `readiness 모델은 '지금 해도 되는가'를 받고 사용자가 직접 수행하며, completion 모델은 '이 일을 해 달라'고 맡기고 완료를 받습니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

커널 이벤트 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `strace -f -e epoll_wait,read,write,io_uring_enter -p <pid>` 같은 확인은 `ready fd, user read loop, submitted operation, completion result` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 ready fd, user read loop, submitted operation, completion result 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서는 ready fd, user read loop, submitted operation, completion result 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `모델 차이를 성능 숫자 하나로만 비교하고 코드 소유권과 backpressure 지점을 놓치는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 ready fd, user read loop, submitted operation, completion result이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

커널 이벤트 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`ready fd, user read loop, submitted operation, completion result` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`같은 파일 전송 서버를 epoll 스타일과 io_uring 스타일로 설계할 때 코드가 달라지는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

#### 운영 질문으로 다시 압축하기

커널 이벤트 모델을 실제 면접 답변으로 꺼낼 때는 먼저 커널 이벤트 모델에서 readiness 신호와 실제 데이터 처리를 어떻게 나눌지 말해야 합니다. 핵심은 readiness 신호와 실제 데이터 처리 책임을 분리하는 것입니다. 예를 들어 `ready fd, user read loop, submitted operation, completion result`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 커널 이벤트 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `ready fd, user read loop, submitted operation, completion result`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`ready fd, user read loop, submitted operation, completion result`의 관측 근거는 `strace`, fd 수, event-loop delay, readiness/completion 호출의 빈도와 대기 시간에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `ready fd, user read loop, submitted operation, completion result` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 커널 이벤트 모델의 세부 메커니즘으로 내려가면 됩니다.

## JVM 네트워크 런타임과 비동기 처리

이 장은 Linux I/O 모델이 Java NIO, Tomcat, Netty, Reactor, WebFlux 안에서 어떤 추상화로 보이는지 연결합니다. 핵심은 비동기라는 단어가 아니라 스레드 점유와 backpressure 책임 경계습니다.

### Java NIO는 blocking socket 모델과 무엇이 다른가

Java NIO는 socket을 non-blocking channel로 바꾸고 selector에 등록해 여러 연결의 준비 상태를 한 흐름에서 다룰 수 있게 합니다. 예를 들어 스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 SocketChannel, Selector, SelectionKey, ByteBuffer를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

JVM 런타임 계층은 Linux fd readiness를 Java 객체와 callback, signal, worker thread로 번역합니다.
여기서 먼저 붙잡을 대상은 SocketChannel, Selector, SelectionKey, ByteBuffer이며, 이 대상은 커널에서 직접 WebFlux handler로 순간 이동하지 않습니다.
Java NIO Selector가 준비된 channel을 알려 주고, Netty EventLoop가 그 channel event를 처리하며, ChannelPipeline이 byte를 HTTP object나 application message로 바꾸고, Reactor는 그 흐름을 Publisher/Subscriber 계약으로 감싼습니다.
Spring WebFlux는 이 계약 위에 controller와 handler adapter를 올립니다.
이 경로를 알면 "WebFlux니까 빠르다"가 아니라 "event loop를 막지 않고, upstream/downstream도 backpressure를 존중할 때 연결 수와 I/O 대기 비용을 줄입니다"고 말할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: SocketChannel, Selector, SelectionKey, ByteBuffer
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `SocketChannel, Selector, SelectionKey, ByteBuffer`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `SocketChannel, Selector, SelectionKey, ByteBuffer` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 JVM 런타임 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `SocketChannel, Selector, SelectionKey, ByteBuffer`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 JVM 런타임 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `SocketChannel, Selector, SelectionKey, ByteBuffer`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Java NIO는 socket을 non-blocking channel로 바꾸고 selector에 등록해 여러 연결의 준비 상태를 한 흐름에서 다룰 수 있게 합니다.
이 문장을 실제 서버로 옮기면 먼저 "스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 SocketChannel, Selector, SelectionKey, ByteBuffer이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
JVM 런타임 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `SocketChannel, Selector, SelectionKey, ByteBuffer`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html), [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `SocketChannel, Selector, SelectionKey, ByteBuffer`의 소유권을 확인하는 근거로 사용합니다.
JVM 런타임 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `SocketChannel, Selector, SelectionKey, ByteBuffer`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 JVM 런타임 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`SocketChannel, Selector, SelectionKey, ByteBuffer`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `SocketChannel, Selector, SelectionKey, ByteBuffer`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
JVM 런타임 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

JVM 런타임 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
jcmd <pid> Thread.print | grep -E 'Selector|Poller|nio'
strace -f -e epoll_wait -p <pid>
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 JVM 런타임 모델을 확인할 때는 thread dump, Netty event-loop thread 이름, Reactor operator log, connection pool 지표를 커널 fd 상태와 함께 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `SocketChannel, Selector, SelectionKey, ByteBuffer`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Java NIO는 socket을 non-blocking channel로 바꾸고 selector에 등록해 여러 연결의 준비 상태를 한 흐름에서 다룰 수 있게 합니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `SocketChannel, Selector, SelectionKey, ByteBuffer`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
JVM 런타임 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| JVM 런타임 모델의 커널 상태 vs 애플리케이션 상태 | `SocketChannel, Selector, SelectionKey, ByteBuffer` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| JVM 런타임 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| JVM 런타임 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| JVM 런타임 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| JVM 런타임 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 JVM 런타임 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`SocketChannel, Selector, SelectionKey, ByteBuffer`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Java NIO는 socket을 non-blocking channel로 바꾸고 selector에 등록해 여러 연결의 준비 상태를 한 흐름에서 다룰 수 있게 합니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: SocketChannel, Selector, SelectionKey, ByteBuffer
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `SocketChannel, Selector, SelectionKey, ByteBuffer`를 먼저 말하고, 그다음 `Java NIO는 socket을 non-blocking channel로 바꾸고 selector에 등록해 여러 연결의 준비 상태를 한 흐름에서 다룰 수 있게 합니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 SocketChannel, Selector, SelectionKey, ByteBuffer 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서는 SocketChannel, Selector, SelectionKey, ByteBuffer 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 SocketChannel, Selector, SelectionKey, ByteBuffer이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

JVM 런타임 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`SocketChannel, Selector, SelectionKey, ByteBuffer` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`스레드 하나가 여러 SocketChannel의 OP_READ readiness를 처리하는 echo server 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

JVM 런타임 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `jcmd <pid> Thread.print | grep -E 'Selector|Poller|nio'` 같은 확인은 `SocketChannel, Selector, SelectionKey, ByteBuffer` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `NIO를 Java가 자동으로 비동기 작업을 끝까지 수행해 주는 모델로 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

#### 운영 질문으로 다시 압축하기

JVM 런타임 모델을 실제 면접 답변으로 꺼낼 때는 먼저 JVM 런타임 경계에서 thread-per-request, event loop, Publisher/Subscriber signal을 어떻게 나눌지 말해야 합니다. 핵심은 thread-per-request, event loop, Publisher/Subscriber signal을 같은 비동기라는 말로 섞지 않는 것입니다. 예를 들어 `SocketChannel, Selector, SelectionKey, ByteBuffer`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 JVM 런타임 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `SocketChannel, Selector, SelectionKey, ByteBuffer`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 thread dump, Netty channel pipeline, Reactor signal log, connection pool 지표에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `SocketChannel, Selector, SelectionKey, ByteBuffer` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 JVM 런타임 모델의 세부 메커니즘으로 내려가면 됩니다.

### Tomcat thread-per-request 모델과 Netty event loop 모델은 어디서 갈라지는가

전통적 Servlet MVC는 요청 처리 흐름이 worker thread를 점유하는 쪽이고, Netty는 event loop가 readiness event를 받아 handler chain으로 흘려보내는 쪽입니다. 예를 들어 같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 Tomcat worker thread, request object, Netty EventLoop, Channel event를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

JVM 런타임 계층은 Linux fd readiness를 Java 객체와 callback, signal, worker thread로 번역합니다.
여기서 먼저 붙잡을 대상은 Tomcat worker thread, request object, Netty EventLoop, Channel event이며, 이 대상은 커널에서 직접 WebFlux handler로 순간 이동하지 않습니다.
Java NIO Selector가 준비된 channel을 알려 주고, Netty EventLoop가 그 channel event를 처리하며, ChannelPipeline이 byte를 HTTP object나 application message로 바꾸고, Reactor는 그 흐름을 Publisher/Subscriber 계약으로 감싼습니다.
Spring WebFlux는 이 계약 위에 controller와 handler adapter를 올립니다.
이 경로를 알면 "WebFlux니까 빠르다"가 아니라 "event loop를 막지 않고, upstream/downstream도 backpressure를 존중할 때 연결 수와 I/O 대기 비용을 줄입니다"고 말할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: Tomcat worker thread, request object, Netty EventLoop, Channel event
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `Tomcat worker thread, request object, Netty EventLoop, Channel event`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `Tomcat worker thread, request object, Netty EventLoop, Channel event` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 JVM 런타임 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `Tomcat worker thread, request object, Netty EventLoop, Channel event`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 JVM 런타임 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `Tomcat worker thread, request object, Netty EventLoop, Channel event`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

전통적 Servlet MVC는 요청 처리 흐름이 worker thread를 점유하는 쪽이고, Netty는 event loop가 readiness event를 받아 handler chain으로 흘려보내는 쪽입니다.
이 문장을 실제 서버로 옮기면 먼저 "같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 Tomcat worker thread, request object, Netty EventLoop, Channel event이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
JVM 런타임 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `Tomcat worker thread, request object, Netty EventLoop, Channel event`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html), [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `Tomcat worker thread, request object, Netty EventLoop, Channel event`의 소유권을 확인하는 근거로 사용합니다.
JVM 런타임 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `Tomcat worker thread, request object, Netty EventLoop, Channel event`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 JVM 런타임 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`Tomcat worker thread, request object, Netty EventLoop, Channel event`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `Tomcat worker thread, request object, Netty EventLoop, Channel event`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
JVM 런타임 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

JVM 런타임 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
jcmd <pid> Thread.print | grep -E 'http-nio|reactor-http|Netty|worker'
ss -antp | grep <pid>
```

`Tomcat worker thread, request object, Netty EventLoop, Channel event`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 JVM 런타임 모델을 확인할 때는 thread dump, Netty event-loop thread 이름, Reactor operator log, connection pool 지표를 커널 fd 상태와 함께 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `Tomcat worker thread, request object, Netty EventLoop, Channel event`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`전통적 Servlet MVC는 요청 처리 흐름이 worker thread를 점유하는 쪽이고, Netty는 event loop가 readiness event를 받아 handler chain으로 흘려보내는 쪽입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `Tomcat worker thread, request object, Netty EventLoop, Channel event`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `Tomcat worker thread, request object, Netty EventLoop, Channel event`를 단일 로그에서만 찾는 것입니다.
JVM 런타임 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| JVM 런타임 모델의 커널 상태 vs 애플리케이션 상태 | `Tomcat worker thread, request object, Netty EventLoop, Channel event` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| JVM 런타임 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| JVM 런타임 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| JVM 런타임 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| JVM 런타임 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`Tomcat worker thread, request object, Netty EventLoop, Channel event`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 JVM 런타임 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`Tomcat worker thread, request object, Netty EventLoop, Channel event`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "전통적 Servlet MVC는 요청 처리 흐름이 worker thread를 점유하는 쪽이고, Netty는 event loop가 readiness event를 받아 handler chain으로 흘려보내는 쪽입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: Tomcat worker thread, request object, Netty EventLoop, Channel event
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `Tomcat worker thread, request object, Netty EventLoop, Channel event`를 먼저 말하고, 그다음 `전통적 Servlet MVC는 요청 처리 흐름이 worker thread를 점유하는 쪽이고, Netty는 event loop가 readiness event를 받아 handler chain으로 흘려보내는 쪽입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서는 Tomcat worker thread, request object, Netty EventLoop, Channel event 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 Tomcat worker thread, request object, Netty EventLoop, Channel event이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

JVM 런타임 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`Tomcat worker thread, request object, Netty EventLoop, Channel event` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

JVM 런타임 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `jcmd <pid> Thread.print | grep -E 'http-nio|reactor-http|Netty|worker'` 같은 확인은 `Tomcat worker thread, request object, Netty EventLoop, Channel event` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `Netty를 쓰면 비즈니스 로직이 자동으로 병렬 처리된다고 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`같은 HTTP 요청을 Spring MVC와 WebFlux/Reactor Netty로 처리할 때 스레드 점유 방식이 갈라지는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 Tomcat worker thread, request object, Netty EventLoop, Channel event 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

### Netty는 selector, event loop, channel, pipeline을 어떻게 조립하는가

Netty는 Java NIO/네이티브 transport의 준비 상태 처리를 EventLoop와 Channel 추상화로 묶고, 애플리케이션 프로토콜 처리는 Pipeline의 handler에게 맡깁니다. 예를 들어 boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

JVM 런타임 계층은 Linux fd readiness를 Java 객체와 callback, signal, worker thread로 번역합니다.
여기서 먼저 붙잡을 대상은 ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event이며, 이 대상은 커널에서 직접 WebFlux handler로 순간 이동하지 않습니다.
Java NIO Selector가 준비된 channel을 알려 주고, Netty EventLoop가 그 channel event를 처리하며, ChannelPipeline이 byte를 HTTP object나 application message로 바꾸고, Reactor는 그 흐름을 Publisher/Subscriber 계약으로 감싼습니다.
Spring WebFlux는 이 계약 위에 controller와 handler adapter를 올립니다.
이 경로를 알면 "WebFlux니까 빠르다"가 아니라 "event loop를 막지 않고, upstream/downstream도 backpressure를 존중할 때 연결 수와 I/O 대기 비용을 줄입니다"고 말할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 JVM 런타임 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 JVM 런타임 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Netty는 Java NIO/네이티브 transport의 준비 상태 처리를 EventLoop와 Channel 추상화로 묶고, 애플리케이션 프로토콜 처리는 Pipeline의 handler에게 맡긴습니다.
이 문장을 실제 서버로 옮기면 먼저 "boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
JVM 런타임 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html), [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`의 소유권을 확인하는 근거로 사용합니다.
JVM 런타임 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 JVM 런타임 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
JVM 런타임 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

JVM 런타임 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
jcmd <pid> Thread.print | grep -E 'nioEventLoop|epollEventLoop'
# Netty 앱이면 LoggingHandler 또는 wiretap 로그 확인
```

`ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 JVM 런타임 모델을 확인할 때는 thread dump, Netty event-loop thread 이름, Reactor operator log, connection pool 지표를 커널 fd 상태와 함께 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Netty는 Java NIO/네이티브 transport의 준비 상태 처리를 EventLoop와 Channel 추상화로 묶고, 애플리케이션 프로토콜 처리는 Pipeline의 handler에게 맡깁니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 단일 로그에서만 찾는 것입니다.
JVM 런타임 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| JVM 런타임 모델의 커널 상태 vs 애플리케이션 상태 | `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| JVM 런타임 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| JVM 런타임 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| JVM 런타임 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| JVM 런타임 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 JVM 런타임 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Netty는 Java NIO/네이티브 transport의 준비 상태 처리를 EventLoop와 Channel 추상화로 묶고, 애플리케이션 프로토콜 처리는 Pipeline의 handler에게 맡긴습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event`를 먼저 말하고, 그다음 `Netty는 Java NIO/네이티브 transport의 준비 상태 처리를 EventLoop와 Channel 추상화로 묶고, 애플리케이션 프로토콜 처리는 Pipeline의 handler에게 맡깁니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

JVM 런타임 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

JVM 런타임 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `jcmd <pid> Thread.print | grep -E 'nioEventLoop|epollEventLoop'` 같은 확인은 `ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `boss event loop가 accept하고 worker event loop가 channelRead를 pipeline으로 전달하는 장면`에서는 ServerChannel, child Channel, EventLoop, ByteBuf, ChannelPipeline event 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `ChannelPipeline을 단순 interceptor 목록으로만 보고 ByteBuf 생애와 framing 책임을 놓치는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

### Reactor는 Publisher, Subscriber, backpressure를 어떤 계약으로 묶는가

Reactor의 핵심은 데이터가 마음대로 밀려오는 것이 아니라 subscription과 request 신호로 생산량을 조절하는 계약입니다. 예를 들어 느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 Subscription.request(n), onNext signal, demand counter를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

JVM 런타임 계층은 Linux fd readiness를 Java 객체와 callback, signal, worker thread로 번역합니다.
여기서 따라갈 대상은 Subscription.request(n), onNext signal, demand counter이며, 이 대상은 커널에서 직접 WebFlux handler로 순간 이동하지 않습니다.
Java NIO Selector가 준비된 channel을 알려 주고, Netty EventLoop가 그 channel event를 처리하며, ChannelPipeline이 byte를 HTTP object나 application message로 바꾸고, Reactor는 그 흐름을 Publisher/Subscriber 계약으로 감싼습니다.
Spring WebFlux는 이 계약 위에 controller와 handler adapter를 올립니다.
이 경로를 알면 "WebFlux니까 빠르다"가 아니라 "event loop를 막지 않고, upstream/downstream도 backpressure를 존중할 때 연결 수와 I/O 대기 비용을 줄입니다"고 말할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: Subscription.request(n), onNext signal, demand counter
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `fd, 큐, 버퍼, 스레드, 연결 상태`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `Subscription.request(n), onNext signal, demand counter` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 JVM 런타임 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `fd, 큐, 버퍼, 스레드, 연결 상태`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 JVM 런타임 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `Subscription.request(n), onNext signal, demand counter`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Reactor의 핵심은 데이터가 마음대로 밀려오는 것이 아니라 subscription과 request 신호로 생산량을 조절하는 계약입니다.
이 문장을 실제 서버로 옮기면 먼저 "느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 Subscription.request(n), onNext signal, demand counter이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
JVM 런타임 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `fd, 큐, 버퍼, 스레드, 연결 상태`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html), [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `Subscription.request(n), onNext signal, demand counter`의 소유권을 확인하는 근거로 사용합니다.
JVM 런타임 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `fd, 큐, 버퍼, 스레드, 연결 상태`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 JVM 런타임 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`Subscription.request(n), onNext signal, demand counter`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `fd, 큐, 버퍼, 스레드, 연결 상태`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
JVM 런타임 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

JVM 런타임 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
Hooks.onOperatorDebug()
# StepVerifier로 request/cancel 흐름 검증
jcmd <pid> Thread.print | grep reactor
```

`Subscription.request(n), onNext signal, demand counter`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 JVM 런타임 모델을 확인할 때는 thread dump, Netty event-loop thread 이름, Reactor operator log, connection pool 지표를 커널 fd 상태와 함께 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `Subscription.request(n), onNext signal, demand counter`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Reactor의 핵심은 데이터가 마음대로 밀려오는 것이 아니라 subscription과 request 신호로 생산량을 조절하는 계약입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `Subscription.request(n), onNext signal, demand counter`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `Subscription.request(n), onNext signal, demand counter`를 단일 로그에서만 찾는 것입니다.
JVM 런타임 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| JVM 런타임 모델의 커널 상태 vs 애플리케이션 상태 | `Subscription.request(n), onNext signal, demand counter` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| JVM 런타임 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| JVM 런타임 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| JVM 런타임 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| JVM 런타임 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`Subscription.request(n), onNext signal, demand counter`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 JVM 런타임 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`Subscription.request(n), onNext signal, demand counter`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Reactor의 핵심은 데이터가 마음대로 밀려오는 것이 아니라 subscription과 request 신호로 생산량을 조절하는 계약입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: Subscription.request(n), onNext signal, demand counter
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `Subscription.request(n), onNext signal, demand counter`를 먼저 말하고, 그다음 `Reactor의 핵심은 데이터가 마음대로 밀려오는 것이 아니라 subscription과 request 신호로 생산량을 조절하는 계약입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

JVM 런타임 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`Subscription.request(n), onNext signal, demand counter` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

JVM 런타임 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `Hooks.onOperatorDebug()` 같은 확인은 `Subscription.request(n), onNext signal, demand counter` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 Subscription.request(n), onNext signal, demand counter 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서는 Subscription.request(n), onNext signal, demand counter이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Flux를 반환하면 TCP와 DB까지 모두 자동으로 속도 조절된다고 말하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`느린 downstream subscriber가 upstream에게 한 번에 1개만 요청하는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 Subscription.request(n), onNext signal, demand counter이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

#### 운영 질문으로 다시 압축하기

JVM 런타임 모델을 실제 면접 답변으로 꺼낼 때는 먼저 JVM 런타임 경계에서 thread-per-request, event loop, Publisher/Subscriber signal을 어떻게 나눌지 말해야 합니다. 핵심은 thread-per-request, event loop, Publisher/Subscriber signal을 같은 비동기라는 말로 섞지 않는 것입니다. 예를 들어 `fd, 큐, 버퍼, 스레드, 연결 상태`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 JVM 런타임 모델을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `fd, 큐, 버퍼, 스레드, 연결 상태`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`Subscription.request(n), onNext signal, demand counter`의 관측 근거는 thread dump, Netty channel pipeline, Reactor signal log, connection pool 지표에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `fd, 큐, 버퍼, 스레드, 연결 상태` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 JVM 런타임 모델의 세부 메커니즘으로 내려가면 됩니다.

### Spring WebFlux는 Servlet MVC와 요청 처리 모델이 어떻게 다른가

Spring WebFlux는 Servlet API 전용 MVC와 달리 reactive adapter와 non-blocking I/O 경계를 중심으로 요청 처리를 구성합니다. 예를 들어 컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

JVM 런타임 계층은 Linux fd readiness를 Java 객체와 callback, signal, worker thread로 번역합니다.
여기서 먼저 붙잡을 대상은 DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal이며, 이 대상은 커널에서 직접 WebFlux handler로 순간 이동하지 않습니다.
Java NIO Selector가 준비된 channel을 알려 주고, Netty EventLoop가 그 channel event를 처리하며, ChannelPipeline이 byte를 HTTP object나 application message로 바꾸고, Reactor는 그 흐름을 Publisher/Subscriber 계약으로 감싼습니다.
Spring WebFlux는 이 계약 위에 controller와 handler adapter를 올립니다.
이 경로를 알면 "WebFlux니까 빠르다"가 아니라 "event loop를 막지 않고, upstream/downstream도 backpressure를 존중할 때 연결 수와 I/O 대기 비용을 줄입니다"고 말할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 JVM 런타임 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 JVM 런타임 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Spring WebFlux는 Servlet API 전용 MVC와 달리 reactive adapter와 non-blocking I/O 경계를 중심으로 요청 처리를 구성합니다.
이 문장을 실제 서버로 옮기면 먼저 "컨트롤러가 `Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
JVM 런타임 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html), [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`의 소유권을 확인하는 근거로 사용합니다.
JVM 런타임 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 JVM 런타임 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
JVM 런타임 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

JVM 런타임 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
jcmd <pid> Thread.print | grep -E 'reactor-http|http-nio'
curl -N http://localhost:8080/stream
# Spring actuator metrics 확인
```

`DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 JVM 런타임 모델을 확인할 때는 thread dump, Netty event-loop thread 이름, Reactor operator log, connection pool 지표를 커널 fd 상태와 함께 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Spring WebFlux는 Servlet API 전용 MVC와 달리 reactive adapter와 non-blocking I/O 경계를 중심으로 요청 처리를 구성합니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 단일 로그에서만 찾는 것입니다.
JVM 런타임 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| JVM 런타임 모델의 커널 상태 vs 애플리케이션 상태 | `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| JVM 런타임 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| JVM 런타임 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| JVM 런타임 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| JVM 런타임 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 JVM 런타임 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Spring WebFlux는 Servlet API 전용 MVC와 달리 reactive adapter와 non-blocking I/O 경계를 중심으로 요청 처리를 구성합니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal`를 먼저 말하고, 그다음 `Spring WebFlux는 Servlet API 전용 MVC와 달리 reactive adapter와 non-blocking I/O 경계를 중심으로 요청 처리를 구성합니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

JVM 런타임 모델을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `jcmd <pid> Thread.print | grep -E 'reactor-http|http-nio'` 같은 확인은 `DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서는 DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `WebFlux 컨트롤러 안에서 blocking JDBC를 호출하면서 전체를 non-blocking이라고 부르는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

JVM 런타임 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`DispatcherServlet request, DispatcherHandler exchange, Mono/Flux signal` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`컨트롤러가`Mono<Response>`를 반환하고 실제 응답 쓰기가 subscriber demand와 event loop 위에서 이어지는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

### WebFlux, Reactor, Netty는 Linux I/O 모델 위에서 어떻게 이어지는가

Linux는 fd readiness를 알려 주고, Netty는 그 이벤트를 channel/pipeline으로 바꾸며, Reactor는 signal과 demand 계약을 제공하고, WebFlux는 HTTP handler 모델을 올립니다. 예를 들어 client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

JVM 런타임 계층은 Linux fd readiness를 Java 객체와 callback, signal, worker thread로 번역합니다.
여기서 먼저 붙잡을 대상은 Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result이며, 이 대상은 커널에서 직접 WebFlux handler로 순간 이동하지 않습니다.
Java NIO Selector가 준비된 channel을 알려 주고, Netty EventLoop가 그 channel event를 처리하며, ChannelPipeline이 byte를 HTTP object나 application message로 바꾸고, Reactor는 그 흐름을 Publisher/Subscriber 계약으로 감싼습니다.
Spring WebFlux는 이 계약 위에 controller와 handler adapter를 올립니다.
이 경로를 알면 "WebFlux니까 빠르다"가 아니라 "event loop를 막지 않고, upstream/downstream도 backpressure를 존중할 때 연결 수와 I/O 대기 비용을 줄입니다"고 말할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 JVM 런타임 모델의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 JVM 런타임 모델을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

Linux는 fd readiness를 알려 주고, Netty는 그 이벤트를 channel/pipeline으로 바꾸며, Reactor는 signal과 demand 계약을 제공하고, WebFlux는 HTTP handler 모델을 올립니다.
이 문장을 실제 서버로 옮기면 먼저 "client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
JVM 런타임 모델을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html), [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`의 소유권을 확인하는 근거로 사용합니다.
JVM 런타임 모델의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 JVM 런타임 모델을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
JVM 런타임 모델을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

JVM 런타임 모델의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
strace -f -e epoll_wait -p <pid>
jcmd <pid> Thread.print | grep reactor-http
curl -v http://localhost:8080/
```

`Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 JVM 런타임 모델을 확인할 때는 thread dump, Netty event-loop thread 이름, Reactor operator log, connection pool 지표를 커널 fd 상태와 함께 맞춰 봅니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`Linux는 fd readiness를 알려 주고, Netty는 그 이벤트를 channel/pipeline으로 바꾸며, Reactor는 signal과 demand 계약을 제공하고, WebFlux는 HTTP handler 모델을 올립니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 단일 로그에서만 찾는 것입니다.
JVM 런타임 모델을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| JVM 런타임 모델의 커널 상태 vs 애플리케이션 상태 | `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| JVM 런타임 모델의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| JVM 런타임 모델의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| JVM 런타임 모델의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| JVM 런타임 모델의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 JVM 런타임 모델을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "Linux는 fd readiness를 알려 주고, Netty는 그 이벤트를 channel/pipeline으로 바꾸며, Reactor는 signal과 demand 계약을 제공하고, WebFlux는 HTTP handler 모델을 올립니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result`를 먼저 말하고, 그다음 `Linux는 fd readiness를 알려 주고, Netty는 그 이벤트를 channel/pipeline으로 바꾸며, Reactor는 signal과 demand 계약을 제공하고, WebFlux는 HTTP handler 모델을 올립니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서는 Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Reactor가 epoll을 직접 호출한다고 설명하거나 WebFlux가 커널 버퍼를 직접 제어한다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

JVM 런타임 모델의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`Linux epoll event, Netty Channel event, Reactor signal, WebFlux handler result` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`client socket에 읽을 byte가 생기고, epoll_wait 결과가 Netty event loop를 깨우며, WebFlux handler가 Mono/Flux로 응답을 만드는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

## Streaming과 backpressure

이 장은 응답을 한 번에 만들지 않는 흐름을 다룹니다. streaming은 HTTP API 모양만의 문제가 아니라 page cache, proxy buffering, TCP window, 느린 client까지 함께 봐야 합니다.

### HTTP streaming은 응답을 한 번에 만들지 않고 어떻게 흘려보내는가

streaming은 응답 객체 전체를 완성한 뒤 한 번에 쓰는 대신, header를 먼저 보내고 body 조각을 생산되는 순서대로 흘려보내는 방식입니다. 예를 들어 서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 response header, chunk, flush, TCP send buffer, client read를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

스트리밍과 backpressure는 "응답을 조금씩 보냅니다"는 표면만으로는 부족합니다.
여기서 먼저 붙잡을 대상은 response header, chunk, flush, TCP send buffer, client read이며, 이 대상은 애플리케이션 buffer, proxy buffer, kernel send buffer, TCP window, client read 속도 사이에서 계속 압력을 주고받습니다.
서버가 조각을 만들었다고 client가 곧바로 본다는 보장은 없습니다.
프록시가 응답을 모을 수 있고, 커널 send buffer가 찰 수 있으며, 느린 client의 receive window가 줄면 송신 쪽 전송이 멈출 수 있습니다.
따라서 streaming 설명은 HTTP API 모양과 운영 경로를 함께 봐야 합니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: response header, chunk, flush, TCP send buffer, client read
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `response header, chunk, flush, TCP send buffer, client read`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `response header, chunk, flush, TCP send buffer, client read` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 스트리밍 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `response header, chunk, flush, TCP send buffer, client read`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 스트리밍 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `response header, chunk, flush, TCP send buffer, client read`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

streaming은 응답 객체 전체를 완성한 뒤 한 번에 쓰는 대신, header를 먼저 보내고 body 조각을 생산되는 순서대로 흘려보내는 방식입니다.
이 문장을 실제 서버로 옮기면 먼저 "서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 response header, chunk, flush, TCP send buffer, client read이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
스트리밍 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `response header, chunk, flush, TCP send buffer, client read`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9110 HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [Nginx proxy_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `response header, chunk, flush, TCP send buffer, client read`의 소유권을 확인하는 근거로 사용합니다.
스트리밍 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `response header, chunk, flush, TCP send buffer, client read`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 스트리밍 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`response header, chunk, flush, TCP send buffer, client read`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `response header, chunk, flush, TCP send buffer, client read`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
스트리밍 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

스트리밍 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
curl -N -v http://localhost:8080/stream
tcpdump -i lo -nn -A tcp port 8080
nginx -T | grep proxy_buffering
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 스트리밍 경로를 확인할 때는 `curl --limit-rate`, `ss -tin`, Nginx buffering 설정, 서버 heap/direct buffer 지표를 함께 보아 데이터가 어디에 쌓이는지 추적합니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `response header, chunk, flush, TCP send buffer, client read`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`streaming은 응답 객체 전체를 완성한 뒤 한 번에 쓰는 대신, header를 먼저 보내고 body 조각을 생산되는 순서대로 흘려보내는 방식입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `response header, chunk, flush, TCP send buffer, client read`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
스트리밍 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 스트리밍 경로의 커널 상태 vs 애플리케이션 상태 | `response header, chunk, flush, TCP send buffer, client read` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 스트리밍 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 스트리밍 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 스트리밍 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 스트리밍 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 스트리밍 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`response header, chunk, flush, TCP send buffer, client read`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "streaming은 응답 객체 전체를 완성한 뒤 한 번에 쓰는 대신, header를 먼저 보내고 body 조각을 생산되는 순서대로 흘려보내는 방식입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: response header, chunk, flush, TCP send buffer, client read
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `response header, chunk, flush, TCP send buffer, client read`를 먼저 말하고, 그다음 `streaming은 응답 객체 전체를 완성한 뒤 한 번에 쓰는 대신, header를 먼저 보내고 body 조각을 생산되는 순서대로 흘려보내는 방식입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서는 response header, chunk, flush, TCP send buffer, client read 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 response header, chunk, flush, TCP send buffer, client read이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

스트리밍 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`response header, chunk, flush, TCP send buffer, client read` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

스트리밍 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `curl -N -v http://localhost:8080/stream` 같은 확인은 `response header, chunk, flush, TCP send buffer, client read` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `Flux를 반환했다는 이유만으로 client가 즉시 조각을 받는다고 단정하는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`서버가 SSE나 NDJSON을 한 줄씩 만들어 client에게 계속 보내는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 response header, chunk, flush, TCP send buffer, client read 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

#### 운영 질문으로 다시 압축하기

스트리밍 경로를 실제 면접 답변으로 꺼낼 때는 먼저 streaming 경로에서 데이터 생성과 중간 buffering을 어떻게 나눌지 말해야 합니다. 핵심은 데이터를 조금씩 만든다는 사실과 중간 계층이 buffering한다는 사실을 함께 보는 것입니다. 예를 들어 `response header, chunk, flush, TCP send buffer, client read`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 스트리밍 경로를 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `response header, chunk, flush, TCP send buffer, client read`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 `curl --limit-rate`, `ss -tin`, proxy buffering 설정, heap/direct buffer, TCP window에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `response header, chunk, flush, TCP send buffer, client read` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 스트리밍 경로의 세부 메커니즘으로 내려가면 됩니다.

### 100GB 파일 스트리밍은 page cache, TCP window, backpressure를 어떻게 지나간다

큰 파일 전송의 핵심은 파일 전체를 heap에 올리지 않고, 커널 page cache와 socket send buffer, TCP window의 압력에 맞춰 조금씩 보내는 것입니다. 예를 들어 서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 file page, page cache entry, send buffer, TCP receive window를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

스트리밍과 backpressure는 "응답을 조금씩 보냅니다"는 표면만으로는 부족합니다.
여기서 먼저 붙잡을 대상은 file page, page cache entry, send buffer, TCP receive window이며, 이 대상은 애플리케이션 buffer, proxy buffer, kernel send buffer, TCP window, client read 속도 사이에서 계속 압력을 주고받습니다.
서버가 조각을 만들었다고 client가 곧바로 본다는 보장은 없습니다.
프록시가 응답을 모을 수 있고, 커널 send buffer가 찰 수 있으며, 느린 client의 receive window가 줄면 송신 쪽 전송이 멈출 수 있습니다.
따라서 streaming 설명은 HTTP API 모양과 운영 경로를 함께 봐야 합니다.

100GB 파일을 다룰 때 가장 먼저 버려야 할 생각은 "응답 크기만큼 애플리케이션 메모리가 필요하다"는 단순한 계산입니다.
파일을 읽는 과정에서 커널 page cache가 개입하면 파일의 page는 JVM heap이 아니라 커널이 관리하는 캐시와 파일 시스템 경로에 먼저 놓입니다.
그 뒤 애플리케이션이나 서버 런타임은 그 page를 HTTP body로 내보내는 일을 시작하지만, 실제 송신 속도는 client가 읽는 속도와 TCP window, 중간 proxy buffering 정책에 의해 계속 조절됩니다.
그래서 운영에서 봐야 하는 숫자도 heap 사용량 하나가 아니라 page cache, process RSS, direct buffer, socket send queue, client receive window, proxy temporary file 여부로 나뉩니다.
이 구분이 있어야 "스트리밍으로 바꿨는데도 메모리가 오른다"는 증상을 heap leak, page cache 증가, direct buffer 압박, proxy buffering 중 어디로 좁힐지 결정할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: file page, page cache entry, send buffer, TCP receive window
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `file page, page cache entry, send buffer, TCP receive window`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `file page, page cache entry, send buffer, TCP receive window` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 스트리밍 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `file page, page cache entry, send buffer, TCP receive window`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 스트리밍 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `file page, page cache entry, send buffer, TCP receive window`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

큰 파일 전송의 핵심은 파일 전체를 heap에 올리지 않고, 커널 page cache와 socket send buffer, TCP window의 압력에 맞춰 조금씩 보내는 것입니다.
이 문장을 실제 서버로 옮기면 먼저 "서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 file page, page cache entry, send buffer, TCP receive window이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
스트리밍 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `file page, page cache entry, send buffer, TCP receive window`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9110 HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [Nginx proxy_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `file page, page cache entry, send buffer, TCP receive window`의 소유권을 확인하는 근거로 사용합니다.
스트리밍 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `file page, page cache entry, send buffer, TCP receive window`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 스트리밍 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`file page, page cache entry, send buffer, TCP receive window`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `file page, page cache entry, send buffer, TCP receive window`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
스트리밍 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

스트리밍 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
curl -O --limit-rate 1m http://host/bigfile
ss -tinp 'sport = :8080'
vmstat 1
cat /proc/meminfo | grep -E 'Cached|Dirty|Writeback'
```

`file page, page cache entry, send buffer, TCP receive window`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 스트리밍 경로를 확인할 때는 `curl --limit-rate`, `ss -tin`, Nginx buffering 설정, 서버 heap/direct buffer 지표를 함께 보아 데이터가 어디에 쌓이는지 추적합니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `file page, page cache entry, send buffer, TCP receive window`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`큰 파일 전송의 핵심은 파일 전체를 heap에 올리지 않고, 커널 page cache와 socket send buffer, TCP window의 압력에 맞춰 조금씩 보내는 것입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `file page, page cache entry, send buffer, TCP receive window`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `file page, page cache entry, send buffer, TCP receive window`를 단일 로그에서만 찾는 것입니다.
스트리밍 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 스트리밍 경로의 커널 상태 vs 애플리케이션 상태 | `file page, page cache entry, send buffer, TCP receive window` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 스트리밍 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 스트리밍 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 스트리밍 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 스트리밍 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`file page, page cache entry, send buffer, TCP receive window`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 스트리밍 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`file page, page cache entry, send buffer, TCP receive window`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "큰 파일 전송의 핵심은 파일 전체를 heap에 올리지 않고, 커널 page cache와 socket send buffer, TCP window의 압력에 맞춰 조금씩 보내는 것입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: file page, page cache entry, send buffer, TCP receive window
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `file page, page cache entry, send buffer, TCP receive window`를 먼저 말하고, 그다음 `큰 파일 전송의 핵심은 파일 전체를 heap에 올리지 않고, 커널 page cache와 socket send buffer, TCP window의 압력에 맞춰 조금씩 보내는 것입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 file page, page cache entry, send buffer, TCP receive window이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

스트리밍 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`file page, page cache entry, send buffer, TCP receive window` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

스트리밍 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `curl -O --limit-rate 1m http://host/bigfile` 같은 확인은 `file page, page cache entry, send buffer, TCP receive window` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 file page, page cache entry, send buffer, TCP receive window 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `서버가 100GB 백업 파일을 HTTP로 내려 주고 client가 느리게 받는 장면`에서는 file page, page cache entry, send buffer, TCP receive window 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `streaming이면 메모리를 전혀 쓰지 않는다고 설명하거나, page cache를 애플리케이션 heap과 섞는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

### proxy buffering은 streaming 응답의 의미를 어떻게 바꿀 수 있는가

proxy buffering이 켜져 있으면 upstream이 보낸 조각이 client에게 즉시 보이지 않을 수 있고, 서버 관측과 client 관측이 갈라집니다. 예를 들어 Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

스트리밍과 backpressure는 "응답을 조금씩 보냅니다"는 표면만으로는 부족합니다.
여기서 먼저 붙잡을 대상은 upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk이며, 이 대상은 애플리케이션 buffer, proxy buffer, kernel send buffer, TCP window, client read 속도 사이에서 계속 압력을 주고받습니다.
서버가 조각을 만들었다고 client가 곧바로 본다는 보장은 없습니다.
프록시가 응답을 모을 수 있고, 커널 send buffer가 찰 수 있으며, 느린 client의 receive window가 줄면 송신 쪽 전송이 멈출 수 있습니다.
따라서 streaming 설명은 HTTP API 모양과 운영 경로를 함께 봐야 합니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 스트리밍 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 스트리밍 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

proxy buffering이 켜져 있으면 upstream이 보낸 조각이 client에게 즉시 보이지 않을 수 있고, 서버 관측과 client 관측이 갈라진습니다.
이 문장을 실제 서버로 옮기면 먼저 "Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
스트리밍 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9110 HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [Nginx proxy_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`의 소유권을 확인하는 근거로 사용합니다.
스트리밍 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 스트리밍 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
스트리밍 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

스트리밍 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
nginx -T | grep -E 'proxy_buffering|proxy_request_buffering|X-Accel-Buffering'
curl -N -i http://host/stream
sudo tail -f /var/log/nginx/access.log
```

`upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 스트리밍 경로를 확인할 때는 `curl --limit-rate`, `ss -tin`, Nginx buffering 설정, 서버 heap/direct buffer 지표를 함께 보아 데이터가 어디에 쌓이는지 추적합니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`proxy buffering이 켜져 있으면 upstream이 보낸 조각이 client에게 즉시 보이지 않을 수 있고, 서버 관측과 client 관측이 갈라집니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 단일 로그에서만 찾는 것입니다.
스트리밍 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 스트리밍 경로의 커널 상태 vs 애플리케이션 상태 | `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 스트리밍 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 스트리밍 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 스트리밍 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 스트리밍 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 스트리밍 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "proxy buffering이 켜져 있으면 upstream이 보낸 조각이 client에게 즉시 보이지 않을 수 있고, 서버 관측과 client 관측이 갈라진습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk`를 먼저 말하고, 그다음 `proxy buffering이 켜져 있으면 upstream이 보낸 조각이 client에게 즉시 보이지 않을 수 있고, 서버 관측과 client 관측이 갈라집니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

스트리밍 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

스트리밍 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `nginx -T | grep -E 'proxy_buffering|proxy_request_buffering|X-Accel-Buffering'` 같은 확인은 `upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서는 upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `애플리케이션 로그의 emit 시각을 client 수신 시각으로 착각하는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Spring WebFlux SSE는 1초마다 emit하지만 Nginx 뒤 client는 한참 뒤 묶어서 받는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 upstream response chunk, Nginx proxy buffer, temporary file, client-facing chunk이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

### 느린 클라이언트는 서버 스레드, 메모리, 소켓 버퍼에 어떤 압력을 주는가

느린 client는 송신 쪽 버퍼와 연결 점유 시간을 늘리고, 모델에 따라 worker thread, event loop task queue, proxy buffer, 메모리 사용량을 압박합니다. 예를 들어 모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

스트리밍과 backpressure는 "응답을 조금씩 보냅니다"는 표면만으로는 부족합니다.
여기서 먼저 붙잡을 대상은 slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer이며, 이 대상은 애플리케이션 buffer, proxy buffer, kernel send buffer, TCP window, client read 속도 사이에서 계속 압력을 주고받습니다.
서버가 조각을 만들었다고 client가 곧바로 본다는 보장은 없습니다.
프록시가 응답을 모을 수 있고, 커널 send buffer가 찰 수 있으며, 느린 client의 receive window가 줄면 송신 쪽 전송이 멈출 수 있습니다.
따라서 streaming 설명은 HTTP API 모양과 운영 경로를 함께 봐야 합니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 스트리밍 경로의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 스트리밍 경로를 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

느린 client는 송신 쪽 버퍼와 연결 점유 시간을 늘리고, 모델에 따라 worker thread, event loop task queue, proxy buffer, 메모리 사용량을 압박합니다.
이 문장을 실제 서버로 옮기면 먼저 "모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
스트리밍 경로를 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9110 HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110), [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy), [Nginx proxy_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering), [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html), [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`의 소유권을 확인하는 근거로 사용합니다.
스트리밍 경로의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 스트리밍 경로를 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
스트리밍 경로를 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

스트리밍 경로의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
curl --limit-rate 10k http://host/large
ss -tinp 'sport = :8080'
jcmd <pid> Thread.print
nginx -T | grep send_timeout
```

`slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 스트리밍 경로를 확인할 때는 `curl --limit-rate`, `ss -tin`, Nginx buffering 설정, 서버 heap/direct buffer 지표를 함께 보아 데이터가 어디에 쌓이는지 추적합니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`느린 client는 송신 쪽 버퍼와 연결 점유 시간을 늘리고, 모델에 따라 worker thread, event loop task queue, proxy buffer, 메모리 사용량을 압박합니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 단일 로그에서만 찾는 것입니다.
스트리밍 경로를 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 스트리밍 경로의 커널 상태 vs 애플리케이션 상태 | `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 스트리밍 경로의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 스트리밍 경로의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 스트리밍 경로의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 스트리밍 경로의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 스트리밍 경로를 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "느린 client는 송신 쪽 버퍼와 연결 점유 시간을 늘리고, 모델에 따라 worker thread, event loop task queue, proxy buffer, 메모리 사용량을 압박합니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer`를 먼저 말하고, 그다음 `느린 client는 송신 쪽 버퍼와 연결 점유 시간을 늘리고, 모델에 따라 worker thread, event loop task queue, proxy buffer, 메모리 사용량을 압박합니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

스트리밍 경로를 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `curl --limit-rate 10k http://host/large` 같은 확인은 `slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서는 slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `서버가 빠르게 응답을 생성하면 느린 client가 서버 자원을 더 쓰지 않는다고 보는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

스트리밍 경로의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`slow client receive window, server send buffer, worker/event-loop scheduling, heap buffer` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`모바일 client가 매우 느린 네트워크에서 대용량 응답을 받는 동안 서버 연결이 오래 유지되는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

## 장애 상황으로 묶어 말하기

이 장은 앞에서 배운 개념을 장애 증상으로 다시 묶습니다. 어떤 증상이 어느 계층의 상태로 관측되는지 좁히는 연습이 문서 전체의 최종 목적입니다.

### 장애 분석은 증상에서 커널, 런타임, 애플리케이션, 외부 시스템 중 어디로 좁혀 가는가

장애 분석은 가장 그럴듯한 원인을 찍는 일이 아니라, 증상을 계층별 신호로 나누고 반증 가능한 순서로 줄여 가는 일입니다. 예를 들어 API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 증상, 관측 지표, 계층별 가설, 검증 결과를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 증상, 관측 지표, 계층별 가설, 검증 결과입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 4-tuple, 포트 범위, 재전송 상태를 남기고, 프록시는 keep-alive 재사용 여부와 upstream 연결 churn을 남깁니다.
애플리케이션과 런타임은 "누가 먼저 연결을 닫도록 만들었는가"를 추적하는 데 필요한 요청 로그, client abort 로그, timeout 로그를 남깁니다.
좋은 장애 분석은 TIME_WAIT 개수 하나를 원인으로 삼지 않고, 짧은 연결이 만들어지는 속도와 닫히는 주체, 포트 재사용 압박이 같은 시간대에 맞물리는지 확인하면서 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: 증상, 관측 지표, 계층별 가설, 검증 결과
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `증상, 관측 지표, 계층별 가설, 검증 결과`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `증상, 관측 지표, 계층별 가설, 검증 결과` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `증상, 관측 지표, 계층별 가설, 검증 결과`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `증상, 관측 지표, 계층별 가설, 검증 결과`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

장애 분석은 가장 그럴듯한 원인을 찍는 일이 아니라, 증상을 계층별 신호로 나누고 반증 가능한 순서로 줄여 가는 일입니다.
이 문장을 실제 서버로 옮기면 먼저 "API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 증상, 관측 지표, 계층별 가설, 검증 결과이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `증상, 관측 지표, 계층별 가설, 검증 결과`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `증상, 관측 지표, 계층별 가설, 검증 결과`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `증상, 관측 지표, 계층별 가설, 검증 결과`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`증상, 관측 지표, 계층별 가설, 검증 결과`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `증상, 관측 지표, 계층별 가설, 검증 결과`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
curl -w '%{time_namelookup} %{time_connect} %{time_appconnect} %{time_starttransfer} %{time_total}\n' -o /dev/null -s https://host
ss -antp
jcmd <pid> Thread.print
vmstat 1
```

여기서 확인할 수 있는 신호는 명령마다 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `증상, 관측 지표, 계층별 가설, 검증 결과`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`장애 분석은 가장 그럴듯한 원인을 찍는 일이 아니라, 증상을 계층별 신호로 나누고 반증 가능한 순서로 줄여 가는 일입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `증상, 관측 지표, 계층별 가설, 검증 결과`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 관측 범위를 단일 로그로 좁히는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `증상, 관측 지표, 계층별 가설, 검증 결과` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

이 비교를 쓰면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`증상, 관측 지표, 계층별 가설, 검증 결과`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "장애 분석은 가장 그럴듯한 원인을 찍는 일이 아니라, 증상을 계층별 신호로 나누고 반증 가능한 순서로 줄여 가는 일입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: 증상, 관측 지표, 계층별 가설, 검증 결과
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `증상, 관측 지표, 계층별 가설, 검증 결과`를 먼저 말하고, 그다음 `장애 분석은 가장 그럴듯한 원인을 찍는 일이 아니라, 증상을 계층별 신호로 나누고 반증 가능한 순서로 줄여 가는 일입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 증상, 관측 지표, 계층별 가설, 검증 결과 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서는 증상, 관측 지표, 계층별 가설, 검증 결과 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 증상, 관측 지표, 계층별 가설, 검증 결과이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`증상, 관측 지표, 계층별 가설, 검증 결과` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`API timeout 알림 하나에서 네트워크, JVM, DB, 외부 API 중 원인을 좁히는 장애 대응 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `curl -w '%{time_namelookup} %{time_connect} %{time_appconnect} %{time_starttransfer} %{time_total}\n' -o /dev/null -s https://host` 같은 확인은 `증상, 관측 지표, 계층별 가설, 검증 결과` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `p99 지연을 애플리케이션 로그 하나만 보고 결론 내리는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

#### 운영 질문으로 다시 압축하기

장애 분석을 실제 면접 답변으로 꺼낼 때는 먼저 장애 분석 경로에서 증상 문구와 실제 원인을 어떻게 나눌지 말해야 합니다. 핵심은 증상 문구를 원인으로 착각하지 않고 계층별로 반증하는 것입니다. 예를 들어 `증상, 관측 지표, 계층별 가설, 검증 결과`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 장애 분석을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `증상, 관측 지표, 계층별 가설, 검증 결과`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

여기서 관측 근거는 커널 로그, `ss`, JVM thread/GC, proxy timing, DB wait, 애플리케이션 요청 로그에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `증상, 관측 지표, 계층별 가설, 검증 결과` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 장애 분석의 세부 메커니즘으로 내려가면 됩니다.

### connection refused와 timeout은 어디가 다른가

refused는 연결 시도에 대해 명시적 거절 신호가 돌아온 쪽이고, timeout은 연결 시도에 대한 충분한 응답을 받지 못한 쪽입니다. 예를 들어 새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 SYN, RST, dropped SYN, retry timer를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 SYN, RST, dropped SYN, retry timer입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: SYN, RST, dropped SYN, retry timer
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `SYN, RST, dropped SYN, retry timer`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "둘을 모두 서버 다운으로만 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `SYN, RST, dropped SYN, retry timer` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `SYN, RST, dropped SYN, retry timer`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `SYN, RST, dropped SYN, retry timer`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `둘을 모두 서버 다운으로만 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

refused는 연결 시도에 대해 명시적 거절 신호가 돌아온 쪽이고, timeout은 연결 시도에 대한 충분한 응답을 받지 못한 쪽입니다.
이 문장을 실제 서버로 옮기면 먼저 "새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 SYN, RST, dropped SYN, retry timer이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `SYN, RST, dropped SYN, retry timer`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `SYN, RST, dropped SYN, retry timer`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `SYN, RST, dropped SYN, retry timer`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`SYN, RST, dropped SYN, retry timer`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `SYN, RST, dropped SYN, retry timer`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
nc -vz host 8080
curl -v --connect-timeout 3 http://host:8080
sudo tcpdump -nn host <host> and tcp
```

`SYN, RST, dropped SYN, retry timer`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `SYN, RST, dropped SYN, retry timer`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `둘을 모두 서버 다운으로만 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
둘을 모두 서버 다운으로만 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`refused는 연결 시도에 대해 명시적 거절 신호가 돌아온 쪽이고, timeout은 연결 시도에 대한 충분한 응답을 받지 못한 쪽입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `SYN, RST, dropped SYN, retry timer`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `SYN, RST, dropped SYN, retry timer`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `SYN, RST, dropped SYN, retry timer` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `둘을 모두 서버 다운으로만 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`SYN, RST, dropped SYN, retry timer`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`SYN, RST, dropped SYN, retry timer`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "refused는 연결 시도에 대해 명시적 거절 신호가 돌아온 쪽이고, timeout은 연결 시도에 대한 충분한 응답을 받지 못한 쪽입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: SYN, RST, dropped SYN, retry timer
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `SYN, RST, dropped SYN, retry timer`를 먼저 말하고, 그다음 `refused는 연결 시도에 대해 명시적 거절 신호가 돌아온 쪽이고, timeout은 연결 시도에 대한 충분한 응답을 받지 못한 쪽입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `둘을 모두 서버 다운으로만 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서는 SYN, RST, dropped SYN, retry timer 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `둘을 모두 서버 다운으로만 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 SYN, RST, dropped SYN, retry timer이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`SYN, RST, dropped SYN, retry timer` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `nc -vz host 8080` 같은 확인은 `SYN, RST, dropped SYN, retry timer` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `둘을 모두 서버 다운으로만 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`새 인스턴스 배포 후 health check가 refused를 보거나 timeout을 보는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 SYN, RST, dropped SYN, retry timer 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

#### 운영 질문으로 다시 압축하기

장애 분석을 실제 면접 답변으로 꺼낼 때는 먼저 장애 분석 경로에서 증상 문구와 실제 원인을 어떻게 나눌지 말해야 합니다. 핵심은 증상 문구를 원인으로 착각하지 않고 계층별로 반증하는 것입니다. 예를 들어 `SYN, RST, dropped SYN, retry timer`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 장애 분석을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `SYN, RST, dropped SYN, retry timer`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`SYN, RST, dropped SYN, retry timer`의 관측 근거는 커널 로그, `ss`, JVM thread/GC, proxy timing, DB wait, 애플리케이션 요청 로그에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `SYN, RST, dropped SYN, retry timer` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 장애 분석의 세부 메커니즘으로 내려가면 됩니다.

### CLOSE_WAIT가 쌓이면 무엇을 의심하는가

CLOSE_WAIT는 상대의 FIN을 받은 뒤 로컬 애플리케이션이 아직 socket을 닫지 않은 상태이므로, 로컬 close 경로와 resource release를 먼저 봅니다. 예를 들어 상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 peer FIN, local ACK, application close 누락, fd retention를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 peer FIN, local ACK, application close 누락, fd retention입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: peer FIN, local ACK, application close 누락, fd retention
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `peer FIN, local ACK, application close 누락, fd retention`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `peer FIN, local ACK, application close 누락, fd retention` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `peer FIN, local ACK, application close 누락, fd retention`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `peer FIN, local ACK, application close 누락, fd retention`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

CLOSE_WAIT는 상대의 FIN을 받은 뒤 로컬 애플리케이션이 아직 socket을 닫지 않은 상태이므로, 로컬 close 경로와 resource release를 먼저 봅니다.
이 문장을 실제 서버로 옮기면 먼저 "상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 peer FIN, local ACK, application close 누락, fd retention이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `peer FIN, local ACK, application close 누락, fd retention`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `peer FIN, local ACK, application close 누락, fd retention`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `peer FIN, local ACK, application close 누락, fd retention`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`peer FIN, local ACK, application close 누락, fd retention`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `peer FIN, local ACK, application close 누락, fd retention`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -antp state close-wait
lsof -p <pid> | grep CLOSE_WAIT
jcmd <pid> Thread.print
```

`peer FIN, local ACK, application close 누락, fd retention`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `peer FIN, local ACK, application close 누락, fd retention`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`CLOSE_WAIT는 상대의 FIN을 받은 뒤 로컬 애플리케이션이 아직 socket을 닫지 않은 상태이므로, 로컬 close 경로와 resource release를 먼저 봅니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `peer FIN, local ACK, application close 누락, fd retention`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `peer FIN, local ACK, application close 누락, fd retention`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `peer FIN, local ACK, application close 누락, fd retention` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`peer FIN, local ACK, application close 누락, fd retention`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`peer FIN, local ACK, application close 누락, fd retention`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "CLOSE_WAIT는 상대의 FIN을 받은 뒤 로컬 애플리케이션이 아직 socket을 닫지 않은 상태이므로, 로컬 close 경로와 resource release를 먼저 봅니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: peer FIN, local ACK, application close 누락, fd retention
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `peer FIN, local ACK, application close 누락, fd retention`를 먼저 말하고, 그다음 `CLOSE_WAIT는 상대의 FIN을 받은 뒤 로컬 애플리케이션이 아직 socket을 닫지 않은 상태이므로, 로컬 close 경로와 resource release를 먼저 봅니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 peer FIN, local ACK, application close 누락, fd retention이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`peer FIN, local ACK, application close 누락, fd retention` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -antp state close-wait` 같은 확인은 `peer FIN, local ACK, application close 누락, fd retention` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 peer FIN, local ACK, application close 누락, fd retention 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `상대가 연결을 닫았는데 Java 프로세스에 CLOSE_WAIT socket이 계속 남는 장면`에서는 peer FIN, local ACK, application close 누락, fd retention 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `CLOSE_WAIT를 네트워크 장비가 계속 붙잡고 있는 상태로 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

#### 운영 질문으로 다시 압축하기

장애 분석을 실제 면접 답변으로 꺼낼 때는 먼저 장애 분석 경로에서 증상 문구와 실제 원인을 어떻게 나눌지 말해야 합니다. 핵심은 증상 문구를 원인으로 착각하지 않고 계층별로 반증하는 것입니다. 예를 들어 `peer FIN, local ACK, application close 누락, fd retention`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 장애 분석을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `peer FIN, local ACK, application close 누락, fd retention`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`peer FIN, local ACK, application close 누락, fd retention`의 관측 근거는 커널 로그, `ss`, JVM thread/GC, proxy timing, DB wait, 애플리케이션 요청 로그에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `peer FIN, local ACK, application close 누락, fd retention` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 장애 분석의 세부 메커니즘으로 내려가면 됩니다.

### TIME_WAIT가 많으면 항상 문제인가

TIME_WAIT는 정상 종료 뒤 늦게 도착한 packet을 걸러 내고 마지막 ACK를 다시 보낼 수 있게 잠시 남는 TCP 안전 상태입니다. `ss`에 TIME_WAIT가 많이 보인다는 사실만으로 장애라고 단정할 수 없고, 짧은 HTTP 연결이 자주 만들어지고 닫히는지, ephemeral port가 부족한지, active close를 어느 쪽이 수행하는지까지 함께 봐야 합니다. 여기서는 active close, final ACK, 2MSL 대기, ephemeral port 재사용 가능성을 시간 순서로 따라갑니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 active close, final ACK, 2MSL wait, ephemeral port reuse입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 짧은 HTTP 연결이 많이 생긴 뒤 로컬 쪽이 먼저 close를 수행해 TIME_WAIT가 쌓이는 장면을 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 active close, 마지막 ACK, 2MSL 대기, local port 재사용 가능성은 운영 환경에서도 직접 확인할 수 있습니다.

```text
짧은 HTTP 요청이 끝납니다
    |
    | 1. 어느 한쪽이 먼저 close를 시작합니다. 먼저 닫은 쪽이 active closer입니다
    v
FIN / ACK 교환
    |
    | 2. 마지막 ACK를 보낸 active closer 쪽 커널이 TIME_WAIT 상태를 유지합니다
    v
2MSL 대기
    |
    | 3. 늦게 도착한 segment를 걸러 내고 마지막 ACK 재전송 가능성을 남깁니다
    v
local address:port 재사용 판단
    |
    | 4. ephemeral port 범위, destination 4-tuple, keep-alive 재사용 여부가 다음 연결 가능성을 좌우합니다
```

여기서 중요한 것은 `active close, final ACK, 2MSL wait, ephemeral port reuse`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
TIME_WAIT 상태 자체는 커널 TCP 상태이고, Java heap 객체나 DB connection pool 항목이 아닙니다.
다만 짧은 연결을 많이 만드는 원인은 HTTP keep-alive 비활성화, 프록시 upstream pool 설정, client timeout, 서버의 close 정책, health check 방식처럼 상위 계층에서 생길 수 있습니다.
그래서 "TIME_WAIT를 leak으로만 보고 무조건 커널 파라미터부터 줄이는 설명"은 위험합니다. 커널 상태만 보고 현상을 발견할 수는 있지만, 연결 churn을 만든 주체까지 곧바로 확정할 수는 없기 때문입니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | TCP close 상태 전이를 관리하고 TIME_WAIT를 유지합니다. | active closer, final ACK, 2MSL timer, local/remote 4-tuple, ephemeral port 범위 | 어떤 HTTP 요청이 짧은 연결을 많이 만들었는지 판단 |
| 프록시 / 로드밸런서 | client-facing 연결과 upstream 연결을 따로 관리하고 keep-alive 재사용 여부를 결정합니다. | upstream keep-alive pool, timeout, health check, connection close header | TCP 상태의 2MSL 의미를 바꾸기 |
| JVM / 서버 런타임 | Tomcat connector나 Netty channel close 이벤트를 애플리케이션 로그와 연결합니다. | channel close, client abort, timeout, access log timing | 커널의 TIME_WAIT timer를 직접 제거 |
| 애플리케이션 | 응답 완료, timeout, 예외, streaming 중단처럼 연결 종료를 유발한 상위 이유를 남깁니다. | request id, response status, timeout, client cancel, retry 정책 | TCP port 재사용 정책 결정 |
| 외부 시스템 / 클라이언트 | 짧은 요청 반복, aggressive reconnect, idle timeout mismatch로 연결 churn을 만들 수 있습니다. | 호출 빈도, retry, keep-alive 설정, upstream close 정책 | 서버 커널의 socket 상태를 직접 소유 |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `ss`에 TIME_WAIT가 많이 보일 때 실제로 봐야 할 질문은 "정상 종료의 흔적이 많은가, 아니면 새 연결을 만들 local port나 conntrack entry가 부족한가"입니다.
같은 숫자라도 초당 요청 수가 높고 keep-alive가 꺼진 서비스에서는 자연스러운 결과일 수 있고, outbound 호출을 많이 하는 서버에서는 ephemeral port 고갈이나 NAT table 압박의 전조일 수 있습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. TIME_WAIT가 많은 쪽은 inbound 서버 포트입니까, outbound ephemeral port입니까.
2. active close를 주로 수행한 쪽은 서버입니까, 프록시입니까, 클라이언트입니까.
3. 같은 시간대에 새 연결 실패, `EADDRNOTAVAIL`, NAT/conntrack 포화, upstream keep-alive 미사용 같은 신호가 있습니까.
4. keep-alive 재사용률과 초당 새 연결 수가 TIME_WAIT 증가 속도와 맞물립니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

TIME_WAIT는 정상 종료 뒤 늦은 packet과 마지막 ACK 재전송을 위해 남는 안전 상태라서, 많다는 사실만으로 장애는 아닙니다.
이 문장을 실제 서버로 옮기면 먼저 "짧은 HTTP 연결이 매우 많아 `ss`에 TIME_WAIT가 많이 보이는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 active close, final ACK, 2MSL wait, ephemeral port reuse이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `active close, final ACK, 2MSL wait, ephemeral port reuse`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293), [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux ip(7)](https://man7.org/linux/man-pages/man7/ip.7.html), Nginx keepalive/upstream 문서, Tomcat connector 문서에서 잡을 수 있습니다.
RFC와 Linux man page는 TIME_WAIT가 왜 남는지와 TCP 상태 이름을 확인하게 해 주고, 프록시와 Tomcat 문서는 짧은 HTTP 연결이 왜 반복되는지 확인하게 해 줍니다.
즉 이 절의 근거는 "커널 상태 이름"과 "상위 계층이 연결을 재사용하거나 닫는 정책"을 분리해서 잡아야 합니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "`ss`에 TIME_WAIT가 많다"와 "장애다" 사이를 곧바로 잇지 않는 편이 정확합니다.
먼저 TIME_WAIT가 정상 종료 뒤 남는 커널 상태임을 말하고, 그다음 짧은 연결 반복, active closer, local port 범위, keep-alive 재사용률을 확인해야 합니다.
프록시 upstream 연결이 계속 새로 만들어지는지, 클라이언트가 매 요청마다 연결을 닫는지, 서버가 timeout 정책으로 먼저 닫는지에 따라 수리 위치가 달라집니다.

이 구분은 운영에서 커널 파라미터를 만지기 전에 특히 중요합니다.
TIME_WAIT가 많아도 새 연결 실패가 없고 local port 여유가 충분하며 keep-alive 정책상 예상 가능한 수준이면 문제로 보지 않을 수 있습니다.
반대로 `EADDRNOTAVAIL`, connect timeout 증가, NAT/conntrack 포화, upstream keep-alive 미사용이 같이 보이면 단순히 숫자를 줄이는 것이 아니라 연결 재사용과 닫는 주체를 함께 고쳐야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ss -ant state time-wait | wc -l
ss -s
cat /proc/sys/net/ipv4/ip_local_port_range
```

`active close, final ACK, 2MSL wait, ephemeral port reuse`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 `ss -ant state time-wait`는 현재 TIME_WAIT socket 수를 보여 주고, `ss -s`는 전체 TCP 상태 분포를 빠르게 보여 줍니다.
`cat /proc/sys/net/ipv4/ip_local_port_range`는 outbound 연결에서 사용할 수 있는 local port 범위를 확인하게 해 줍니다.
Nginx access log와 upstream keepalive 설정은 프록시가 매 요청마다 새 upstream 연결을 만드는지, Tomcat connector 로그는 서버가 먼저 닫는 흐름이 있는지 확인하게 해 줍니다.

어긋나는 신호는 `active close, final ACK, 2MSL wait, ephemeral port reuse`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `TIME_WAIT를 leak으로만 보고 무조건 커널 파라미터부터 줄이는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
새 연결 실패가 함께 보이면 local port 범위, destination 4-tuple, NAT/conntrack, upstream keep-alive 설정을 먼저 봐야 합니다.
처리 지연만 있고 새 연결 실패가 없다면 TIME_WAIT 숫자보다 요청 처리 시간, client timeout, 프록시 idle timeout을 따로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
TIME_WAIT를 leak으로만 보고 무조건 커널 파라미터부터 줄이는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`TIME_WAIT는 정상 종료 뒤 늦은 packet과 마지막 ACK 재전송을 위해 남는 안전 상태라서, 많다는 사실만으로 장애는 아닙니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `ss`에 TIME_WAIT 상태가 많이 보일 정도로 짧은 HTTP 연결이 자주 만들어지고 닫히는 장면에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `active close, final ACK, 2MSL wait, ephemeral port reuse`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `active close, final ACK, 2MSL wait, ephemeral port reuse`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| TIME_WAIT vs CLOSE_WAIT | 정상 종료 뒤 active closer 쪽 커널에 잠시 남는 상태를 봅니다 | 상대 FIN을 받은 뒤 로컬 애플리케이션이 아직 close하지 않은 상태를 봅니다 | 문제는 정상 종료의 잔여 상태입니까, 로컬 close 누락입니까 |
| 많은 TIME_WAIT vs 포트 고갈 | 숫자는 많지만 새 연결 실패가 없는 상태를 봅니다 | `EADDRNOTAVAIL`, connect timeout, NAT table 포화가 같이 보이는 상태를 봅니다 | 숫자가 큰 것입니까, 실제로 새 연결을 만들 자원이 부족합니까 |
| active close vs passive close | 로컬 쪽이 먼저 FIN을 보내 TIME_WAIT를 소유합니다 | 상대가 먼저 닫고 로컬은 다른 종료 상태를 거칩니다 | TIME_WAIT가 쌓이는 쪽이 왜 먼저 닫고 있습니까 |
| keep-alive 재사용 vs 짧은 연결 반복 | 같은 TCP 연결 위에 여러 HTTP 요청을 싣습니다 | 요청마다 connect와 close가 반복됩니다 | 프록시와 client가 연결을 재사용하고 있습니까 |
| 커널 파라미터 조정 vs 상위 정책 조정 | port range, TIME_WAIT 재사용 정책 같은 OS 한도를 봅니다 | upstream keepalive, timeout, retry, health check 주기를 봅니다 | 한도를 늘릴 문제입니까, 연결 churn을 줄일 문제입니까 |

이 비교를 쓰면 "TIME_WAIT가 많다"라는 문장을 바로 튜닝 처방으로 바꾸지 않게 됩니다.
먼저 많은 숫자가 예상 가능한 정상 상태인지, 실제 연결 실패와 함께 나타나는 압박인지 분리합니다.
그다음 active close를 만든 주체가 커널이 아니라 HTTP keep-alive 정책, 프록시 upstream 설정, client retry 방식, 서버 timeout 정책 중 어디에 가까운지 확인합니다.
커널 한도를 조정해야 하는 경우도 있지만, 연결 churn 자체를 줄이면 더 작은 설정 변경으로 문제를 닫을 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "TIME_WAIT는 정상 종료 뒤 늦은 packet과 마지막 ACK 재전송을 위해 남는 안전 상태라서, 많다는 사실만으로 장애는 아닙니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `누가 active close를 했는지 -> local port 여유가 있는지 -> keep-alive가 재사용되는지 -> 실제 connect 실패가 있는지` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 `ss`, access log, 프록시 설정, kernel parameter를 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: active close, final ACK, 2MSL wait, ephemeral port reuse
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: TIME_WAIT count, active closer 추정, port reuse 가능성, keep-alive 재사용 여부
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `active close, final ACK, 2MSL wait, ephemeral port reuse`를 먼저 말하고, 그다음 `TIME_WAIT는 정상 종료 뒤 늦은 packet과 마지막 ACK 재전송을 위해 남는 안전 상태라서, 많다는 사실만으로 장애는 아닙니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `ss`에 TIME_WAIT 상태가 많이 보일 정도로 짧은 HTTP 연결이 자주 만들어지고 닫히는 장면을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `TIME_WAIT를 leak으로만 보고 무조건 커널 파라미터부터 줄이는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

이 상태를 설명할 때는 "어디에 byte가 쌓였는가"보다 "어떤 연결 식별자가 얼마나 오래 재사용 대기 중인가"를 먼저 말해야 합니다.
TIME_WAIT는 대용량 payload가 버퍼에 쌓인 상태가 아니라, 특정 local/remote address와 port 조합을 잠시 안전하게 보류하는 TCP 상태입니다.
따라서 메모리나 지연이 늘어난다면 TIME_WAIT 자체가 byte buffer를 채운 것인지가 아니라, 짧은 연결 반복 때문에 connect 비용, NAT/conntrack 비용, proxy upstream 연결 비용이 같이 늘었는지 확인해야 합니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ss -ant state time-wait | wc -l` 같은 확인은 `active close, final ACK, 2MSL wait, ephemeral port reuse` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `TIME_WAIT를 leak으로만 보고 무조건 커널 파라미터부터 줄이는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 한도를 바꾸는지 분명히 말해야 합니다.
`ip_local_port_range`는 outbound 연결의 local port 후보 범위를 바꾸고, keep-alive 설정은 새 연결 생성 속도를 줄이며, 프록시 upstream pool은 backend와의 연결 재사용 방식을 바꿉니다.
커널 TIME_WAIT 관련 설정은 플랫폼과 커널 버전에 따라 의미가 달라질 수 있으므로, 공식 문서와 현재 커널 값을 확인하지 않고 관성적으로 줄이면 안 됩니다.
설정값을 바꾸기 전에 실제로 부족한 것이 local port인지, NAT/conntrack entry인지, upstream keep-alive인지, client retry 폭주인지 확인해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `ss`에 TIME_WAIT 상태가 많이 보일 정도로 짧은 HTTP 연결이 자주 만들어지고 닫히는 장면에서는 active close, final ACK, 2MSL wait, ephemeral port reuse 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `TIME_WAIT를 leak으로만 보고 무조건 커널 파라미터부터 줄이는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 "애플리케이션 thread가 기다리는가"와 "커널이 TIME_WAIT timer를 유지하는가"를 분리합니다.
TIME_WAIT는 보통 thread가 붙잡고 기다리는 상태가 아니라 커널이 일정 시간 보존하는 TCP 상태입니다.
따라서 thread dump에서 대기 thread를 찾는 일보다, `ss`, 연결 생성 실패 로그, 프록시 upstream timing, port range, keep-alive 재사용률을 먼저 맞추는 편이 자연스럽습니다.
같은 timeout이라도 connect timeout은 새 연결 생성 압박을 가리킬 수 있고, upstream read timeout은 이미 만들어진 연결에서 응답을 기다리는 문제일 수 있습니다.

#### 운영 질문으로 다시 압축하기

운영 답변으로 압축하면 먼저 "TIME_WAIT는 정상 종료 뒤 active closer 쪽에 남는 커널 TCP 상태"라고 말해야 합니다. 그다음 숫자가 많은 것과 장애를 분리합니다. 새 연결 실패가 없고 keep-alive 정책상 예상 가능한 churn이면 관찰 대상일 수 있고, local port 부족이나 NAT/conntrack 압박이 함께 보이면 수리 대상이 됩니다.

실무에서는 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 예를 들어 프록시가 backend upstream 연결을 재사용하지 못해 매 요청마다 새 TCP 연결을 만들고 닫는다고 가정해 보겠습니다. 이때 `ss`에는 TIME_WAIT가 늘고, Nginx upstream log에는 짧은 연결이 반복되며, backend에는 요청 처리 지연보다 connect/close 비용이 먼저 보일 수 있습니다. 이 장면에서는 controller 코드를 먼저 고치기보다 upstream keep-alive, timeout, active close 주체, local port 범위를 확인해야 합니다.

관측 근거는 `ss`, `ss -s`, local port range, 프록시 access/upstream log, connection reuse metric에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 각 명령이 어느 질문을 줄이는지 말하는 것입니다. `ss`는 커널 TCP 상태를 보여 주고, port range는 outbound 연결 후보의 폭을 보여 주며, 프록시 로그는 연결을 새로 만들었는지 재사용했는지 보여 줍니다. 이 신호들이 같은 시간대에 맞물리면 가설이 강해지고, 맞물리지 않으면 TIME_WAIT가 아니라 다른 지연 원인으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 TIME_WAIT가 많다는 사실만으로 장애라고 보지 않습니다. 먼저 active close 주체, 초당 새 연결 수, local port 여유, keep-alive 재사용률, 실제 connect 실패 여부를 같이 봅니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 2MSL과 마지막 ACK 재전송 의미로 내려가면 됩니다.

조금 더 실제적인 예로, 사용자가 보는 API 서버의 8080 포트 주변에 TIME_WAIT가 많다고 해 보겠습니다. 이때 먼저 볼 것은 서버가 매 요청 뒤 연결을 먼저 닫고 있는지, 아니면 클라이언트가 짧은 연결을 반복하면서 서버 쪽에는 상대적으로 다른 상태가 남는지입니다. `ss -antp '( sport = :8080 or dport = :8080 )'`처럼 포트를 좁혀 보면 local/remote 쌍과 상태 분포를 함께 볼 수 있습니다. 서버 local port가 8080인 TIME_WAIT가 많으면 서버가 active close를 수행했을 가능성을 보지만, reverse proxy가 앞에 있으면 실제 active close 주체가 Nginx와 Tomcat 사이의 upstream 연결일 수도 있습니다. 그래서 access log의 `Connection: close`, upstream keepalive 설정, idle timeout 값을 같이 봐야 합니다.

반대로 애플리케이션이 외부 API를 많이 호출하는 outbound 장면에서는 관측 지점이 달라집니다. 이 경우 TIME_WAIT가 많이 보이는 local port는 8080 같은 서비스 포트가 아니라 ephemeral port 범위에서 나옵니다. 짧은 외부 호출을 초당 많이 만들고 keep-alive를 재사용하지 못하면 local port 후보가 빠르게 TIME_WAIT에 묶이고, 어느 순간 `Cannot assign requested address`나 connect timeout이 나타날 수 있습니다. 이때는 controller 처리 시간을 먼저 의심하기보다 HTTP client connection pool, DNS별 destination 분포, NAT gateway 또는 conntrack 한도, `ip_local_port_range`를 같이 봅니다. 같은 TIME_WAIT 숫자라도 inbound 서버 포트 주변에서 보이는 것과 outbound ephemeral port에서 보이는 것은 수리 경로가 다릅니다.

운영에서 설정을 바꿀 때도 순서가 있습니다. 먼저 `ss -s`로 상태 분포를 보고, 특정 포트나 remote destination으로 좁혀 어느 4-tuple 주변에서 TIME_WAIT가 늘어나는지 확인합니다. 그다음 Nginx나 client의 keep-alive 재사용률을 보고, 새 연결 생성 속도와 오류 로그가 같은 시각에 올라오는지 맞춥니다. 여기까지 맞은 뒤에야 port range 확대, upstream keepalive 조정, timeout 조정, retry 간격 조정 같은 선택지가 의미를 갖습니다. 커널 값을 먼저 낮추거나 재사용 정책부터 켜는 식의 접근은 당장은 숫자를 줄여 보일 수 있어도, 늦은 packet 처리와 연결 식별 충돌이라는 TCP의 안전 장치를 약하게 만들 수 있습니다.

면접에서는 이 내용을 한 문단으로 줄여 말할 수 있어야 합니다. "TIME_WAIT는 정상 종료의 흔적일 수 있으므로, 저는 숫자만 보고 장애라고 하지 않습니다. active close를 누가 했는지, 짧은 연결이 왜 반복되는지, outbound라면 ephemeral port가 부족한지, proxy와 HTTP client가 keep-alive를 재사용하는지 확인합니다. 새 연결 실패가 없으면 관찰 대상으로 두고, 실패가 있으면 연결 churn을 줄이는 설정과 port/NAT 한도를 함께 봅니다." 이 정도로 말하면 정의와 운영 판단이 함께 보입니다.

마지막으로, 이 상태를 메모리 누수와 같은 방식으로 설명하지 않는 것이 중요합니다. TIME_WAIT 항목은 보통 "객체가 해제되지 않았다"는 뜻이 아니라 "TCP가 같은 연결 식별자의 재사용을 잠시 늦추고 있다"는 뜻입니다. 물론 항목 수가 많으면 커널이 관리해야 할 상태도 늘지만, 운영 판단의 중심은 누수 여부보다 연결 생성률과 재사용률입니다. 그래서 같은 대시보드에 TIME_WAIT count, 초당 new connection, keep-alive hit ratio, connect error, NAT/conntrack 사용량을 같이 놓으면 훨씬 빨리 좁혀집니다. 이 값들이 따로 움직이면 TIME_WAIT는 배경 신호일 수 있고, 함께 움직이면 연결 churn이 실제 병목 후보가 됩니다. 이 차이를 끝까지 분리해야 합니다.

### too many open files는 왜 네트워크 장애처럼 보이는가

이 항목은 fd 한도 초과가 왜 네트워크 장애처럼 보이는지 설명합니다. 네트워크 연결도 fd를 쓰기 때문에 fd 한도 초과는 새 socket, 파일, epoll 등록, 로그 파일 열기까지 함께 실패시키며 네트워크 장애처럼 보입니다. 예를 들어 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 fd allocation failure, accept failure, DB socket creation failure, log file open failure를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 fd allocation failure, accept failure, DB socket creation failure, log file open failure입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: fd allocation failure, accept failure, DB socket creation failure, log file open failure
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `fd allocation failure, accept failure, DB socket creation failure, log file open failure` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

네트워크 연결도 fd를 쓰기 때문에 fd 한도 초과는 새 socket, 파일, epoll 등록, 로그 파일 열기까지 함께 실패시키며 네트워크 장애처럼 보입니다.
이 문장을 실제 서버로 옮기면 먼저 "새 연결이 들어오는데 서버가 `EMFILE` 때문에 accept하지 못하고 client는 연결 실패를 보는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 fd allocation failure, accept failure, DB socket creation failure, log file open failure이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `fd allocation failure, accept failure, DB socket creation failure, log file open failure`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`fd allocation failure, accept failure, DB socket creation failure, log file open failure`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
ulimit -n
cat /proc/<pid>/limits
ls /proc/<pid>/fd | wc -l
journalctl -u myapp | grep -i 'too many open files\|EMFILE'
```

`fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`네트워크 연결도 fd를 쓰기 때문에 fd 한도 초과는 새 socket, 파일, epoll 등록, 로그 파일 열기까지 함께 실패시키며 네트워크 장애처럼 보입니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `fd allocation failure, accept failure, DB socket creation failure, log file open failure` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "네트워크 연결도 fd를 쓰기 때문에 fd 한도 초과는 새 socket, 파일, epoll 등록, 로그 파일 열기까지 함께 실패시키며 네트워크 장애처럼 보입니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: fd allocation failure, accept failure, DB socket creation failure, log file open failure
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `fd allocation failure, accept failure, DB socket creation failure, log file open failure`를 먼저 말하고, 그다음 `네트워크 연결도 fd를 쓰기 때문에 fd 한도 초과는 새 socket, 파일, epoll 등록, 로그 파일 열기까지 함께 실패시키며 네트워크 장애처럼 보입니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `ulimit -n` 같은 확인은 `fd allocation failure, accept failure, DB socket creation failure, log file open failure` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 fd allocation failure, accept failure, DB socket creation failure, log file open failure 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서는 fd allocation failure, accept failure, DB socket creation failure, log file open failure 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `open files라는 문구 때문에 파일 I/O 문제로만 한정하는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 fd allocation failure, accept failure, DB socket creation failure, log file open failure이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`fd allocation failure, accept failure, DB socket creation failure, log file open failure` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
새 연결이 들어오는데 서버가 `EMFILE` 때문에 `accept()`하지 못하고 client는 연결 실패를 보는 장면에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

### p99 latency가 튈 때 CPU, GC, DB, network를 어떻게 가른다

이 항목은 p99 지연이 튈 때 CPU, GC, DB, network를 어떻게 나누어 확인할지 설명합니다. p99는 소수 요청의 꼬리 지연이므로 CPU 포화, GC stop-the-world, DB lock/wait, network retransmission, upstream timeout을 같은 타임라인에 놓고 봅니다. 기준 장면은 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`입니다. 이 장면을 보면 p99 지연도 추상적인 숫자가 아니라 실제 운영 경로 위에서 생기는 상태라는 점이 드러납니다. 여기서는 latency histogram, thread state, GC pause, DB wait, TCP retransmission를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 latency histogram, thread state, GC pause, DB wait, TCP retransmission입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

p99를 볼 때 평균값이 낮다는 사실은 장애가 없다는 증거가 아닙니다.
대부분의 요청이 30ms에 끝나도, 일부 요청이 3초까지 밀리면 사용자는 그 느린 요청을 실제 장애로 경험합니다.
따라서 p99 분석은 "어느 계층이 가끔 늦어지는가"를 묻는 작업입니다.
CPU가 계속 높으면 계산 또는 run queue를 먼저 보고, GC pause가 같은 시간대에 튀면 JVM 중단 시간을 봅니다.
DB wait가 겹치면 lock, I/O, connection pool 대기를 확인하고, retransmission이나 upstream timeout이 맞물리면 network와 proxy 경로를 봅니다.
이 순서가 있어야 p99 숫자 하나를 보고 무조건 scale-out하거나, 반대로 평균이 낮다는 이유로 장애를 부정하는 실수를 피할 수 있습니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: latency histogram, thread state, GC pause, DB wait, TCP retransmission
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `latency histogram, thread state, GC pause, DB wait, TCP retransmission` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

p99는 소수 요청의 꼬리 지연이므로 CPU 포화, GC stop-the-world, DB lock/wait, network retransmission, upstream timeout을 같은 타임라인에 놓고 봅니다.
이 문장을 실제 서버로 옮기면 먼저 "평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 latency histogram, thread state, GC pause, DB wait, TCP retransmission이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `latency histogram, thread state, GC pause, DB wait, TCP retransmission`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`latency histogram, thread state, GC pause, DB wait, TCP retransmission`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
curl -w '%{time_connect} %{time_starttransfer} %{time_total}\n'
jcmd <pid> GC.heap_info
jstat -gcutil <pid> 1s 10
vmstat 1
sar -n TCP,ETCP 1
```

`latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`p99는 소수 요청의 꼬리 지연이므로 CPU 포화, GC stop-the-world, DB lock/wait, network retransmission, upstream timeout을 같은 타임라인에 놓고 봅니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `latency histogram, thread state, GC pause, DB wait, TCP retransmission` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "p99는 소수 요청의 꼬리 지연이므로 CPU 포화, GC stop-the-world, DB lock/wait, network retransmission, upstream timeout을 같은 타임라인에 놓고 봅니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: latency histogram, thread state, GC pause, DB wait, TCP retransmission
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `latency histogram, thread state, GC pause, DB wait, TCP retransmission`를 먼저 말하고, 그다음 `p99는 소수 요청의 꼬리 지연이므로 CPU 포화, GC stop-the-world, DB lock/wait, network retransmission, upstream timeout을 같은 타임라인에 놓고 봅니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 latency histogram, thread state, GC pause, DB wait, TCP retransmission 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서는 latency histogram, thread state, GC pause, DB wait, TCP retransmission 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 latency histogram, thread state, GC pause, DB wait, TCP retransmission이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`latency histogram, thread state, GC pause, DB wait, TCP retransmission` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`평균은 괜찮은데 p99만 튀고 일부 요청이 timeout에 걸리는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `curl -w '%{time_connect} %{time_starttransfer} %{time_total}\n'` 같은 확인은 `latency histogram, thread state, GC pause, DB wait, TCP retransmission` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `평균 latency가 낮다는 이유로 장애를 부정하거나, p99만 보고 무조건 scale-out하는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

### 컨테이너에서 OOMKilled가 났을 때 Linux 관점에서 무엇을 본다

컨테이너 OOMKilled는 cgroup limit 안에서 커널이 프로세스를 죽인 이벤트(event)일 수 있으므로 JVM heap 로그뿐 아니라 cgroup memory.events와 container runtime 이벤트를 함께 봐야 합니다. 예를 들어 Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 cgroup memory usage, process RSS, heap/direct/native memory, OOM event를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 cgroup memory usage, process RSS, heap/direct/native memory, OOM event입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: cgroup memory usage, process RSS, heap/direct/native memory, OOM event
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `cgroup memory usage, process RSS, heap/direct/native memory, OOM event` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

컨테이너 OOMKilled는 cgroup limit 안에서 커널이 프로세스를 죽인 이벤트(event)일 수 있으므로 JVM heap 로그뿐 아니라 cgroup memory.events와 container runtime 이벤트를 함께 봐야 합니다.
이 문장을 실제 서버로 옮기면 먼저 "Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 cgroup memory usage, process RSS, heap/direct/native memory, OOM event이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`cgroup memory usage, process RSS, heap/direct/native memory, OOM event`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
kubectl describe pod <pod>
kubectl logs <pod> --previous
cat /sys/fs/cgroup/memory.events
cat /proc/<pid>/status | grep -E 'VmRSS|VmHWM|Threads'
```

`cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`컨테이너 OOMKilled는 cgroup limit 안에서 커널이 프로세스를 죽인 이벤트(event)일 수 있으므로 JVM heap 로그뿐 아니라 cgroup memory.events와 container runtime 이벤트를 함께 봐야 합니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `cgroup memory usage, process RSS, heap/direct/native memory, OOM event` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "컨테이너 OOMKilled는 cgroup limit 안에서 커널이 프로세스를 죽인 이벤트(event)일 수 있으므로 JVM heap 로그뿐 아니라 cgroup memory.events와 container runtime 이벤트를 함께 봐야 합니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: cgroup memory usage, process RSS, heap/direct/native memory, OOM event
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event`를 먼저 말하고, 그다음 `컨테이너 OOMKilled는 cgroup limit 안에서 커널이 프로세스를 죽인 이벤트(event)일 수 있으므로 JVM heap 로그뿐 아니라 cgroup memory.events와 container runtime 이벤트를 함께 봐야 합니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서는 cgroup memory usage, process RSS, heap/direct/native memory, OOM event 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 cgroup memory usage, process RSS, heap/direct/native memory, OOM event이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`cgroup memory usage, process RSS, heap/direct/native memory, OOM event` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `kubectl describe pod <pod>` 같은 확인은 `cgroup memory usage, process RSS, heap/direct/native memory, OOM event` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `Java heap이 limit보다 작으면 OOMKilled가 불가능하다고 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Kubernetes Pod가 exit code 137로 재시작되고 Java 로그에는 명확한 OutOfMemoryError가 없는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 cgroup memory usage, process RSS, heap/direct/native memory, OOM event 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

### WebFlux나 Netty로 바꿨는데도 느릴 때 무엇을 의심하는가

비동기 런타임은 I/O 대기 중 스레드 점유를 줄이는 도구이지 CPU 작업, blocking DB, 느린 downstream, 잘못된 buffering을 자동으로 해결하지 않습니다. 예를 들어 Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 event loop task, blocking call, demand signal, DB connection, proxy buffer를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 event loop task, blocking call, demand signal, DB connection, proxy buffer입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: event loop task, blocking call, demand signal, DB connection, proxy buffer
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "WebFlux 전환을 성능 개선 보증으로 설명하는 경우" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `event loop task, blocking call, demand signal, DB connection, proxy buffer` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `event loop task, blocking call, demand signal, DB connection, proxy buffer`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `WebFlux 전환을 성능 개선 보증으로 설명하는 경우`라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

비동기 런타임은 I/O 대기 중 스레드 점유를 줄이는 도구이지 CPU 작업, blocking DB, 느린 downstream, 잘못된 buffering을 자동으로 해결하지 않습니다.
이 문장을 실제 서버로 옮기면 먼저 "Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 event loop task, blocking call, demand signal, DB connection, proxy buffer이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `event loop task, blocking call, demand signal, DB connection, proxy buffer`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `event loop task, blocking call, demand signal, DB connection, proxy buffer`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `event loop task, blocking call, demand signal, DB connection, proxy buffer`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`event loop task, blocking call, demand signal, DB connection, proxy buffer`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
jcmd <pid> Thread.print | grep -A30 'reactor-http'
# BlockHound 적용 테스트
curl -N --limit-rate 1k http://host/stream
nginx -T | grep proxy_buffering
```

`event loop task, blocking call, demand signal, DB connection, proxy buffer`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `WebFlux 전환을 성능 개선 보증으로 설명하는 경우`라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
WebFlux 전환을 성능 개선 보증으로 설명하는 경우는 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`비동기 런타임은 I/O 대기 중 스레드 점유를 줄이는 도구이지 CPU 작업, blocking DB, 느린 downstream, 잘못된 buffering을 자동으로 해결하지 않습니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `event loop task, blocking call, demand signal, DB connection, proxy buffer`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `event loop task, blocking call, demand signal, DB connection, proxy buffer` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `WebFlux 전환을 성능 개선 보증으로 설명하는 경우`라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`event loop task, blocking call, demand signal, DB connection, proxy buffer`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`event loop task, blocking call, demand signal, DB connection, proxy buffer`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "비동기 런타임은 I/O 대기 중 스레드 점유를 줄이는 도구이지 CPU 작업, blocking DB, 느린 downstream, 잘못된 buffering을 자동으로 해결하지 않습니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: event loop task, blocking call, demand signal, DB connection, proxy buffer
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 먼저 말하고, 그다음 `비동기 런타임은 I/O 대기 중 스레드 점유를 줄이는 도구이지 CPU 작업, blocking DB, 느린 downstream, 잘못된 buffering을 자동으로 해결하지 않습니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `WebFlux 전환을 성능 개선 보증으로 설명하는 경우` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 event loop task, blocking call, demand signal, DB connection, proxy buffer이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`event loop task, blocking call, demand signal, DB connection, proxy buffer` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `jcmd <pid> Thread.print | grep -A30 'reactor-http'` 같은 확인은 `event loop task, blocking call, demand signal, DB connection, proxy buffer` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `WebFlux 전환을 성능 개선 보증으로 설명하는 경우` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 event loop task, blocking call, demand signal, DB connection, proxy buffer 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Spring MVC에서 WebFlux로 바꾼 뒤에도 대용량 요청에서 latency가 줄지 않는 장면`에서는 event loop task, blocking call, demand signal, DB connection, proxy buffer 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `WebFlux 전환을 성능 개선 보증으로 설명하는 경우`라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

#### 운영 질문으로 다시 압축하기

장애 분석을 실제 면접 답변으로 꺼낼 때는 먼저 장애 분석 경로에서 증상 문구와 실제 원인을 어떻게 나눌지 말해야 합니다. 핵심은 증상 문구를 원인으로 착각하지 않고 계층별로 반증하는 것입니다. 예를 들어 `event loop task, blocking call, demand signal, DB connection, proxy buffer`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 장애 분석을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `event loop task, blocking call, demand signal, DB connection, proxy buffer`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`event loop task, blocking call, demand signal, DB connection, proxy buffer`의 관측 근거는 커널 로그, `ss`, JVM thread/GC, proxy timing, DB wait, 애플리케이션 요청 로그에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `event loop task, blocking call, demand signal, DB connection, proxy buffer` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 장애 분석의 세부 메커니즘으로 내려가면 됩니다.

### 커널 파라미터와 애플리케이션 설정은 언제 같이 봐야 하는가

커널 파라미터와 애플리케이션 설정은 같은 병목 지점을 양쪽에서 제한할 때 함께 봐야 합니다. 한쪽만 키우면 다음 좁은 목이 드러납니다. 예를 들어 Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면을 보면 이 흐름이 추상적인 그림이 아니라 실제 운영 경로라는 점이 드러납니다. 여기서는 somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout를 따라가며, 각 계층이 어떤 상태를 직접 소유하고 어떤 상태를 다음 동작에 넘기는지 확인합니다.

장애 상황에서는 개념을 따로 외우는 방식이 거의 도움이 되지 않습니다.
여기서 먼저 붙잡을 대상은 somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout입니다.
그 대상을 시간 순서로 놓고 어느 계층에서 멈췄는지 좁혀야 합니다.
커널은 TCP 상태와 fd, cgroup memory, scheduler 상태를 남기고, 런타임은 thread dump와 GC log, event loop delay를 남기며, 애플리케이션은 요청 로그와 business timing을 남깁니다.
DB와 외부 API는 query plan, wait event, connection pool, upstream timeout으로 자기 신호를 남깁니다.
좋은 장애 분석은 이 신호들을 한 타임라인에 맞춰 보고, 반증 가능한 순서로 가설을 줄입니다.

#### 값이 움직이는 순서

아래 흐름은 이 질문을 `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`이라는 실제 요청 흐름 안에 놓고 본 것입니다.
표현은 단순화했지만, 여기서 보이는 fd, 큐, 버퍼, 스레드, 연결 상태는 운영 환경에서도 직접 확인할 수 있습니다.

```text
client / upstream / local trigger
    |
    | 1. 요청 경계로 들어오는 대상: somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout
    v
kernel 또는 OS boundary
    |
    | 2. 커널은 fd, socket state, buffer, queue, route, memory accounting 같은 자기 상태를 갱신합니다
    v
proxy 또는 JVM runtime
    |
    | 3. Nginx, Tomcat, Netty, Reactor 중 해당 계층이 byte를 HTTP 메시지, event, signal, object로 해석합니다
    v
application / DB / external system
    |
    | 4. 애플리케이션은 비즈니스 규칙, DB query, 외부 API 호출로 의미를 만듭니다
    v
response path
    |
    | 5. 결과는 다시 byte가 되어 proxy, kernel send buffer, network를 거쳐 client로 돌아갑니다
```

여기서 중요한 것은 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 한 덩어리로 뭉뚱그리지 않고, 어느 계층이 그 상태를 소유하는지 나누어 보는 것입니다.
커널이 남기는 것은 Java 객체가 아니라 fd, 연결 상태, socket buffer, route 결과이고, 프록시가 남기는 것은 kernel packet이 아니라 upstream 요청, header 조정, buffering 정책입니다.
Tomcat이나 Netty가 만드는 것도 곧바로 domain object가 아니라 request wrapper, channel event, ByteBuf, decoded HTTP object에 가깝습니다.
Spring controller와 service가 business 의미를 만들고, DB driver는 그 의미를 다시 SQL과 database protocol byte로 낮춥니다.
그래서 "설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명" 같은 설명은 중간 변환 하나를 지워 버려 장애 확인 지점까지 함께 지우는 문제가 있습니다.

#### 계층별 책임 경계

| 계층 | 여기서 맡는 일 | 직접 소유하는 상태 | 이 계층이 하지 않는 일 |
| --- | --- | --- | --- |
| 커널 | fd, socket, TCP 상태, buffer, route, cgroup, scheduler 같은 실행 상태를 관리합니다. | `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout` 중 커널 객체와 큐에 해당하는 부분 | HTTP controller 선택, JSON mapping, SQL 의미 해석 |
| 프록시 / 로드밸런서 | client-facing 연결을 받고 upstream 연결이나 HTTP 메시지로 다시 만듭니다. | TLS 종료 여부, upstream pool, proxy buffer, timeout, health check | Java thread dump 해석, DB query plan 선택 |
| JVM / 서버 런타임 | fd readiness나 blocking read를 Java thread, selector, event loop, request object로 올립니다. | worker thread, selector key, channel, direct buffer, heap object | 커널 route 결정, NIC interrupt 처리 |
| 애플리케이션 | request를 business 의미로 해석하고 service/repository 흐름을 실행합니다. | controller argument, transaction boundary, domain object, response DTO | TCP retransmission, cgroup OOM victim 결정 |
| DB / 외부 시스템 | query, transaction, lock, result set, remote API response를 처리합니다. | connection, cursor, lock, execution plan, wait event | 프록시 buffering, JVM event loop scheduling |

이 표는 장애 분석의 원인을 한 계층에 고정하라는 뜻이 아닙니다.
오히려 반대입니다.
한 요청은 계층을 모두 지납니다.
다만 각 계층이 소유한 상태를 분리해야 질문이 선명해집니다.
예를 들어 `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서 response가 느리다면 애플리케이션 method 시간이 길 수도 있고, DB connection pool에서 기다렸을 수도 있으며, Nginx가 upstream response를 buffer에 모으고 있을 수도 있고, client가 느려 send buffer가 차 있을 수도 있습니다.
따라서 여기서 말하는 "느림"은 하나의 원인이 아니라 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`가 멈춘 위치에 따라 다른 수리 경로를 갖습니다.

운영에서는 그래서 장애 분석을 아래 순서로 묻는 편이 안전합니다.

1. 여기서 보이는 증상은 연결 생성 전, 연결 생성 중, 요청 body 수신 중, 애플리케이션 처리 중, DB/외부 호출 중, 응답 송신 중 어디에서 관측됩니까.
2. `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`의 그 지점 주인은 커널, 프록시, JVM 런타임, 애플리케이션, DB 중 누구입니까.
3. 주인이 가진 숫자나 로그는 무엇입니까. `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 `ss`, `lsof`, `/proc`, `journalctl`, `jcmd`, Nginx access log, DB wait view 중 무엇으로 직접 볼 수 있습니까.
4. 바로 위 계층과 바로 아래 계층의 신호가 같은 시간대에 맞물립니까. `설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명`이라는 설명을 반증할 수 있는 인접 계층의 신호는 무엇입니까.

#### 메커니즘을 조금 더 낮은 층에서 보기

커널 파라미터와 애플리케이션 설정은 같은 병목 지점을 양쪽에서 제한할 때 함께 봐야 합니다. 한쪽만 키우면 다음 좁은 목이 드러납니다.
이 문장을 실제 서버로 옮기면 먼저 "Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면"을 붙잡아야 합니다.
여기서 움직이는 대상은 somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout이고, 그 대상은 한 번에 한 계층씩 낮아지거나 높아집니다.
장애 분석을 높은 층의 API 이름으로만 기억하면 호출 순서만 남지만, `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`가 남기는 낮은 층의 상태를 보면 왜 그 API와 설정이 그런 모양인지 이해할 수 있습니다.

여기서 공식 자료로 먼저 확인할 내용은 [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html), [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html), [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/), [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html), [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html), [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)에서 잡을 수 있습니다.
이 자료들은 서로 다른 계층을 설명하지만, 여기서는 모두 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`의 소유권을 확인하는 근거로 사용합니다.
장애 분석의 근거를 되짚을 때 Linux man page와 RFC는 커널과 프로토콜의 상태 이름을 확인하게 해 주고, JVM/Spring/Netty/Reactor 문서는 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`가 Java 객체, callback, signal, handler 흐름으로 올라오는 방식을 확인하게 해 줍니다.
Nginx나 Tomcat 문서는 장애 분석을 애플리케이션 코드만으로 설명할 수 없다는 점도 보여 줍니다. user-space 서버는 커널 socket 위에 자기 timeout, queue, buffer, worker 정책을 따로 얹습니다.

그래서 이 설명을 면접이나 장애 분석에서 사용할 때는 "무엇이 무엇을 대신한다"가 아니라 "`somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`가 어떤 계층의 입력이 되는가"라고 말하는 편이 정확합니다.
여기서 커널 readiness는 런타임 이벤트의 단서가 될 수 있지만, `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 HTTP request나 application object로 해석하는 일까지 대신하지 않습니다.
여기서 프록시 upstream response는 client response의 원재료가 될 수 있지만, proxy buffering이나 header rewrite가 끼면 같은 byte-for-byte tunnel이라고 단정할 수 없습니다.
여기서 DB result set이나 외부 API response도 Java collection으로 순간 이동하지 않습니다. driver buffer, pool state, transaction boundary를 지나야 애플리케이션 의미가 됩니다.

이 구분은 `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면` 같은 성능 설명에서도 중요합니다.
하나의 latency 숫자 안에는 DNS lookup, TCP connect, TLS handshake, proxy upstream wait, runtime scheduling, DB connection wait, query execution, response write, client receive delay가 섞일 수 있습니다.
장애 분석을 제대로 설명하려면 total latency 하나를 크게 말하기보다, 어떤 계층의 시간이 늘었고 어떤 명령이나 로그가 그 판단을 지지하는지 함께 말해야 합니다.

#### 직접 확인할 수 있는 관측 지점

장애 분석의 핵심 내용은 아래 명령으로 작게 재현해 볼 수 있습니다.

실제 운영 서버에서 실행할 때는 권한, 개인정보, 트래픽 영향, 로그 보존 정책을 먼저 확인해야 합니다.

개발 환경이나 staging에서 작은 재현을 만든 뒤 production에서는 읽기 전용 관측부터 시작하는 편이 안전합니다.

```sh
sysctl net.core.somaxconn net.ipv4.tcp_max_syn_backlog
cat /proc/<pid>/limits
nginx -T | grep worker_connections
# Tomcat server.tomcat.* 설정 확인
```

`somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 확인할 때 명령마다 신호는 다르지만, 공통적으로는 "예상한 계층의 상태가 실제로 보인다"는 것입니다.
예를 들어 장애 분석을 확인할 때는 증상이 보인 시각의 `ss`, 커널 로그, JVM thread/GC 자료, proxy timing, DB wait view를 한 줄의 타임라인으로 맞춥니다.
`strace`에서는 Java나 Nginx가 결국 `read`, `write`, `accept4`, `epoll_wait` 같은 시스템 콜로 커널과 만나는 지점을 확인할 수 있습니다.
`jcmd Thread.print`에서는 Tomcat worker, Reactor event loop, DB pool 대기 같은 JVM 내부 대기 지점이 보입니다.
Nginx 설정 출력에서는 proxy buffering, upstream, timeout이 애플리케이션 밖에서 응답 의미를 바꿀 수 있음을 확인합니다.

어긋나는 신호는 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 엉뚱한 계층에서만 찾는 경우입니다.
예를 들어 `설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명`이라는 식으로 설명하면, 여기서 확인해야 할 관측 지점이 사라집니다.
연결 생성 전 문제는 packet trace, listen socket, route, firewall 표면에서 먼저 닫아야 하고, handler 진입 뒤 문제는 thread dump, event-loop delay, DB wait, application timing으로 좁혀야 합니다.
관측 명령은 결론을 대신하지 않습니다.
명령이 보여 준 상태가 어느 계층의 상태인지 설명할 수 있을 때 비로소 이 항목의 검증이 끝납니다.

#### 자주 틀리는 설명과 장애 해석

첫 번째 위험은 계층을 건너뛰는 설명입니다.
설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명은 운영 중에도 자주 보이는 오해입니다.
그럴듯해 보이지만, 한 계층을 지워 버리는 순간 반증할 명령도 사라집니다.
네트워크 요청은 "클라이언트가 Spring controller를 호출한다"로 줄일 수 없습니다.
그 사이에는 DNS, routing, TCP state, TLS endpoint, proxy request, servlet 또는 Netty runtime, application handler, DB protocol이 있습니다.

두 번째 위험은 기술 이름을 해결책으로 착각하는 것입니다.
epoll, Netty, WebFlux, streaming, backpressure, cgroup, systemd 같은 이름은 특정 문제를 줄이기 위한 도구이거나 관측 표면이지 모든 문제를 자동으로 해결하는 해법이 아닙니다.
`커널 파라미터와 애플리케이션 설정은 같은 병목 지점을 양쪽에서 제한할 때 함께 봐야 합니다. 한쪽만 키우면 다음 좁은 목이 드러납니다.`라는 문장은 어느 병목을 줄이는지까지 말할 때만 쓸모가 있습니다.
예를 들어 많은 fd를 효율적으로 기다리는 모델은 CPU가 무거운 JSON 직렬화를 줄여 주지 않고, event loop 모델은 blocking DB 호출을 event loop 위에서 실행하는 순간 장점이 사라집니다.

세 번째 위험은 `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서 정상 상태와 장애 상태를 구분하지 못하는 것입니다.
같은 숫자라도 문맥에 따라 의미가 달라집니다.
TIME_WAIT는 정상 종료의 안전 장치일 수 있고, CLOSE_WAIT는 로컬 close 누락 신호일 수 있으며, page cache 증가는 정상 캐시일 수도 cgroup 압박의 일부일 수도 있습니다.
따라서 여기서는 숫자 하나를 보는 대신 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`가 어느 계층에서 누가 소유한 상태인지 먼저 고정합니다.

네 번째 위험은 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 단일 로그에서만 찾는 것입니다.
장애 분석을 운영 장애에서 확인하려면 access log, application log, GC log, thread dump, `ss`, `lsof`, `/proc`, DB wait view, proxy metrics 중 필요한 표면을 한 타임라인으로 붙여야 합니다.
한 로그에 아무것도 없다는 사실은 그 계층에 증거가 없다는 뜻일 뿐, 전체 시스템에 문제가 없다는 뜻이 아닙니다.

#### 비슷한 개념과 나란히 놓기

헷갈리는 개념은 나란히 놓아야 빨리 정리됩니다.
아래 비교는 암기용 표가 아니라 장애 분석에서 질문을 좁히기 위한 기준입니다.

| 비교축 | 왼쪽을 볼 때 | 오른쪽을 볼 때 | 판단 질문 |
| --- | --- | --- | --- |
| 장애 분석의 커널 상태 vs 애플리케이션 상태 | `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout` 중 fd, TCP state, buffer, route, cgroup처럼 OS가 직접 소유한 상태 | request object, transaction, business object, controller log처럼 앱이 만든 상태 | `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서 증상은 연결이 만들어지기 전입니까, 앱 handler가 시작된 뒤입니까 |
| 장애 분석의 blocking vs non-blocking | 준비될 때까지 호출한 thread가 기다립니다 | 준비 안 됐으면 즉시 돌아오고 readiness나 completion을 따로 기다립니다 | 기다리는 비용을 thread가 냅니까, event loop와 state machine이 냅니까 |
| 장애 분석의 readiness vs completion | fd가 읽기/쓰기 가능하다는 신호를 받습니다 | 제출한 I/O 작업이 끝났다는 결과를 받습니다 | 작업 수행의 반복 책임이 사용자 코드에 남습니까, 커널 제출/완료 모델로 이동합니까 |
| 장애 분석의 streaming vs buffering | 조각을 생산되는 대로 전달하려 합니다 | 중간 계층이 조각을 모아 한 번에 보내거나 임시 파일에 담을 수 있습니다 | client가 보는 시간과 upstream이 보낸 시간이 같습니까 |
| 장애 분석의 TCP backpressure vs Reactive Streams backpressure | receive window와 send buffer가 byte 전송 속도를 조절합니다 | subscriber demand가 Publisher 생산량을 조절합니다 | `설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명`이라는 설명이 어느 계층의 속도 조절을 섞고 있습니까 |

`somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 이 비교에 넣으면 불필요한 단정이 줄고, 어떤 계층의 상태를 확인해야 하는지 좁혀집니다.
예를 들어 장애 분석을 설명하면서 "backpressure가 걸렸다"라고만 말하면, TCP receive window가 줄어든 것인지, Reactor subscriber가 `request(n)`을 적게 보낸 것인지, Nginx가 proxy buffer를 채우고 upstream read를 늦춘 것인지 알 수 없습니다.
각각의 확인 명령도 다릅니다.
TCP는 `ss -tinp`와 packet trace를 보고, Reactor는 signal 로그나 StepVerifier, operator demand를 보며, Nginx는 `proxy_buffering`과 access/upstream timing을 봅니다.

`somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 기준으로 해도 비교의 목적은 선택지의 우열을 한 번에 정하는 것이 아닙니다.
Spring MVC가 항상 낡았고 WebFlux가 항상 우월한 것도 아니며, epoll이 있으니 thread가 필요 없는 것도 아닙니다.
사용자의 요청이 짧고 CPU 작업이 많으며 JDBC/JPA 중심이면 단순한 Servlet MVC가 더 이해하기 쉽고 운영하기 쉬울 수 있습니다.
반대로 많은 idle connection, streaming, 여러 non-blocking upstream 조합이 핵심이면 WebFlux/Reactor/Netty 모델이 thread 점유를 줄이는 데 유리할 수 있습니다.

#### 면접과 실무에서 다시 말하는 방법

면접에서 짧게 답해야 한다면 용어를 먼저 늘어놓기보다 움직이는 대상부터 말하는 편이 좋습니다.
예를 들어 이 질문에는 "커널 파라미터와 애플리케이션 설정은 같은 병목 지점을 양쪽에서 제한할 때 함께 봐야 합니다. 한쪽만 키우면 다음 좁은 목이 드러납니다."라는 판단을 먼저 놓고, 곧바로 어느 계층의 상태를 볼 수 있는지 이어 갑니다.
그 뒤 꼬리 질문이 들어오면 `socket fd -> kernel queue/buffer -> runtime thread/event -> application object -> DB/external I/O` 순서로 내려가면 됩니다.
이 순서는 단순히 문서를 읽기 좋게 만드는 장치가 아니라, 실제 장애 대응에서 명령을 고르는 순서이기도 합니다.

직접 손으로 다시 그려 볼 때는 아래 세 줄이면 충분합니다.

```text
입력: somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout
규칙: 여기서 설명한 계층의 소유 상태와 처리 규칙
출력: 바로 이어지는 계층이 소비할 fd readiness, event, request object, SQL result, response byte
```

이 세 줄을 다른 주제에도 옮길 수 있으면 이해가 남은 것입니다.
여기서는 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 먼저 말하고, 그다음 `커널 파라미터와 애플리케이션 설정은 같은 병목 지점을 양쪽에서 제한할 때 함께 봐야 합니다. 한쪽만 키우면 다음 좁은 목이 드러납니다.`라는 규칙을 붙인 뒤, 마지막으로 바로 이어지는 계층이 실제로 받는 상태를 말하면 됩니다.
예를 들어 `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`을 다시 그릴 때는 "무엇이 들어왔습니까 -> 누가 그 상태를 소유합니까 -> 바로 이어지는 계층은 어떤 형태로 소비합니까" 순서가 무너지면 안 됩니다.
이 순서가 유지되면 `설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명` 같은 설명이 왜 부족한지도 자연스럽게 드러납니다.

읽고 난 뒤에는 한 가지 질문을 스스로 던지면 됩니다.

지금 설명한 상태가 어느 계층에서 관측되는지 먼저 물어보셔야 합니다.

이 질문에 `ss`, `lsof`, `strace`, `jcmd`, Nginx log, DB wait view 중 하나로 답할 수 없으면 아직 가설입니다.

운영에서도 면접에서도 가설과 확인된 사실을 분리해야 합니다.

장애 분석의 버퍼를 설명할 때는 어디에 쌓인 byte인지 먼저 말해야 합니다.
`somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout` 중 일부는 kernel receive buffer나 send buffer에 있고, 일부는 proxy buffer, JVM heap, direct buffer, DB driver buffer에 있을 수 있습니다.
`Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서 메모리나 지연이 늘어난다면 어느 버퍼가 늘었는지 확인해야 합니다.
커널 버퍼가 찬 문제를 controller 코드만 고쳐 해결할 수 없고, proxy buffering 문제를 JVM heap 크기만 키워 해결할 수도 없습니다.

장애 분석을 운영에서 검증할 때 명령은 결론이 아니라 다음 질문을 줄이는 도구입니다.
예를 들어 `sysctl net.core.somaxconn net.ipv4.tcp_max_syn_backlog` 같은 확인은 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout` 중 한 계층의 상태만 보여 줍니다.
그 결과가 PASS처럼 보여도 바로 전체 원인을 확정하지 말고, 바로 위와 아래 계층의 신호가 같은 시간대에 맞는지 봅니다.
이 방식으로 `설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명` 같은 단정은 반증 가능한 가설로 내려오고, 실제 수리 경로도 좁아집니다.

여기서 설정값을 바꿀 때는 어떤 queue, pool, limit의 상한을 바꾸는지 말해야 합니다.
`Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서 한 설정을 키우면 병목이 사라지는 것이 아니라 다음 좁은 지점으로 이동할 수 있습니다.
예를 들어 listen backlog, acceptCount, maxConnections, worker thread, DB pool, proxy timeout, cgroup memory limit은 서로 다른 계층의 상한입니다.
설정값을 바꾸기 전에 somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout 가운데 실제로 한도에 닿은 항목이 무엇인지 확인해야 합니다. 확인 없이 값을 키우면 병목은 다른 큐, pool, timeout, 메모리 한도로 옮겨갈 수 있습니다.

여기서 "연결"이라는 단어를 쓸 때는 먼저 어떤 계층의 연결인지 말해야 합니다.
기준 장면인 `Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서는 somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout 같은 대상이 서로 다른 연결 의미를 만들 수 있습니다.
TCP 연결은 커널이 sequence number, window, retransmission timer로 관리하는 상태이고, HTTP keep-alive는 그 TCP 연결 위에 여러 request/response를 싣는 사용 방식입니다.
DB connection이나 Reactor subscription은 또 다른 계층의 계약입니다.
그래서 `설정 값을 크게 하면 안전하다고 생각하고 downstream 용량과 timeout, pool size를 같이 보지 않는 설명`이라고만 말하면 부족합니다. 어느 fd, 큐, 버퍼, 스레드, pool, signal을 봐야 하는지까지 이어지지 않기 때문입니다.

여기서 대기 지점을 찾을 때는 누가 기다리는지부터 분리합니다.
`Tomcat acceptCount를 키웠는데 여전히 새 연결이 밀리거나, ulimit만 올렸는데 DB가 먼저 죽는 장면`에서는 호출한 thread가 커널 안에서 기다릴 수도 있고, event loop가 readiness를 기다릴 수도 있으며, DB나 외부 시스템이 자기 wait 상태를 남길 수도 있습니다.
대상은 somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout이고, 이 대상이 멈춘 곳에 따라 thread dump, `ss`, proxy timing, DB wait view, cgroup event 중 확인 표면이 달라집니다.
같은 timeout이라도 socket connect timeout, upstream read timeout, pool acquisition timeout, query timeout은 서로 다른 소유자의 대기입니다.

#### 운영 질문으로 다시 압축하기

장애 분석을 실제 면접 답변으로 꺼낼 때는 먼저 장애 분석 경로에서 증상 문구와 실제 원인을 어떻게 나눌지 말해야 합니다. 핵심은 증상 문구를 원인으로 착각하지 않고 계층별로 반증하는 것입니다. 예를 들어 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`를 관찰한다고 해서 곧바로 원인이 확정되는 것은 아닙니다. 이 값이 커널이 소유한 상태인지, 프록시가 만든 HTTP 경계인지, JVM 런타임이 관리하는 thread나 event인지, 애플리케이션이 만든 business 객체인지 먼저 구분해야 합니다. 이 구분이 없으면 같은 timeout, 같은 latency, 같은 연결 실패를 보고도 매번 다른 원인을 찍게 됩니다.

실무에서는 장애 분석을 한 장면으로 다시 좁혀 말하는 편이 좋습니다. 사용자가 요청을 보냈고, 서버가 어느 시점에서 늦거나 실패했다고 가정해 보겠습니다. 이때 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`는 단순한 용어 목록이 아니라 시간 순서에 놓인 관측 대상입니다. 먼저 연결이 만들어졌는지 확인하고, 그다음 요청 byte가 user-space 서버까지 올라왔는지 확인하며, 그 뒤에 런타임 thread나 event loop가 작업을 잡았는지, 애플리케이션이 DB나 외부 시스템을 기다렸는지 봅니다. 응답 쪽에서는 반대로 애플리케이션 결과가 proxy buffer와 kernel send buffer를 지나 client receive window까지 빠져나갔는지 확인합니다.

`somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout`의 관측 근거는 커널 로그, `ss`, JVM thread/GC, proxy timing, DB wait, 애플리케이션 요청 로그에서 잡습니다. 중요한 것은 명령을 많이 나열하는 것이 아니라, 명령 하나가 어느 계층의 상태를 보여 주는지 말하는 것입니다. `ss`가 보여 주는 것은 TCP socket 상태이고, thread dump가 보여 주는 것은 JVM thread의 대기 지점이며, proxy log가 보여 주는 것은 client-facing 연결과 upstream 연결 사이의 시간입니다. DB wait view는 데이터베이스 내부의 lock, I/O, CPU, transaction wait를 보여 줍니다. 같은 시각에 이 신호들이 맞물리면 가설이 강해지고, 맞물리지 않으면 다른 계층으로 돌아가야 합니다.

면접에서는 답을 길게 시작하기보다 판단 순서를 먼저 말하면 됩니다. "저는 먼저 이 증상이 연결 생성 전인지, 요청 처리 중인지, 응답 송신 중인지 나눕니다. 그다음 `somaxconn, acceptCount, maxConnections, ulimit, pool size, timeout` 중 어떤 상태가 어느 계층에 남는지 보고, 커널, 프록시, 런타임, 애플리케이션, DB의 관측값을 같은 시간대에 맞춥니다." 이렇게 말하면 외운 정의보다 강합니다. 이어서 필요한 경우 장애 분석의 세부 메커니즘으로 내려가면 됩니다.

## 면접 질문 세트

아래 질문은 문서 전체를 다시 꺼내 말하기 위한 질문입니다. 짧은 암기 답을 만들기보다, 각 질문을 받았을 때 어느 계층의 상태를 먼저 고정할지 떠올리는 데 사용합니다.

1. DNS가 끝난 뒤 HTTP 요청이 실제로 서버 애플리케이션에 도착하기까지 어떤 연결과 핸드셰이크가 필요합니까?
2. Nginx와 Tomcat 사이의 upstream 연결은 클라이언트와 Nginx 사이의 연결과 왜 별개입니까?
3. `bind`, `listen`, `accept`는 각각 어떤 커널 상태를 만들고, 애플리케이션은 어느 시점부터 요청 byte를 읽을 수 있습니까?
4. `CLOSE_WAIT`와 `TIME_WAIT`가 많이 보일 때 각각 로컬 애플리케이션과 정상 종료 관점에서 무엇을 먼저 확인해야 합니까?
5. epoll readiness와 io_uring completion은 서버 코드의 반복 책임을 어떻게 다르게 만듭니다?
6. Tomcat thread-per-request와 Netty event loop는 많은 연결을 다룰 때 어떤 자원을 서로 다르게 소비합니까?
7. Reactive Streams의 `request(n)`은 TCP receive window와 같은 계층입니까, 아니면 다른 계층의 속도 조절 계약입니까?
8. HTTP streaming에서 서버가 chunk를 만들었는데 client가 늦게 보는 경우, proxy buffering과 TCP backpressure 중 무엇을 확인해야 합니까?
9. p99 latency가 튈 때 CPU, GC, DB, network, proxy 중 어느 계층부터 줄여 갈 것입니까?
10. 컨테이너 OOMKilled와 Java `OutOfMemoryError`는 어떻게 다르며, cgroup 관점에서 어떤 파일을 볼 수 있습니까?

## 직접 확인하는 명령어

아래 명령은 운영 서버에서 바로 실행하라는 처방이 아니라, 각 계층이 실제로 어떤 신호를 남기는지 재현하기 위한 출발점입니다.
production에서는 권한, 개인정보, 부하 영향, 로그 보존 정책을 먼저 확인하고, 가능한 한 읽기 전용 관측부터 시작합니다.

| 계층 | 명령 | 무엇을 보는가 |
| --- | --- | --- |
| DNS | `dig +trace api.example.com` | 이름이 어떤 authoritative server와 record를 거쳐 IP 후보로 바뀌는지 봅니다. |
| routing | `ip route get <ip>` | 특정 목적지 IP로 나갈 interface, source IP, next hop을 봅니다. |
| TCP state | `ss -antp` | ESTABLISHED, SYN-RECV, CLOSE-WAIT, TIME-WAIT 같은 연결 상태를 봅니다. |
| listening socket | `ss -ltnp 'sport = :8080'` | 어떤 프로세스가 어떤 local address와 port에서 listen 중인지 봅니다. |
| fd 사용량 | `ls /proc/<pid>/fd | wc -l`,`cat /proc/<pid>/limits` | too many open files 위험과 실제 fd 사용량을 비교합니다. |
| syscall | `strace -f -e socket,bind,listen,accept4,epoll_wait,read,write -p <pid>` | JVM이나 Nginx가 커널과 만나는 시스템 콜 경계를 봅니다. |
| JVM thread | `jcmd <pid> Thread.print` | Tomcat worker, Reactor event loop, DB pool 대기, lock wait를 봅니다. |
| memory/cgroup | `cat /sys/fs/cgroup/memory.current`, `cat /sys/fs/cgroup/memory.events` | 컨테이너 memory limit 안의 현재 사용량과 OOM event를 봅니다. |
| CPU/scheduler | `vmstat 1`, `pidstat -t -p <pid> 1`, `top -H -p <pid>` | runnable thread, iowait, per-thread CPU 압박을 봅니다. |
| proxy | `nginx -T | grep -E 'proxy_pass|proxy_buffering|proxy_read_timeout'` | upstream, buffering, timeout 설정을 봅니다. |
| streaming | `curl -N -v --limit-rate 10k http://host/stream` | client 수신 속도가 느릴 때 서버/프록시 동작을 봅니다. |
| packet | `tcpdump -i eth0 -nn tcp port 8080` | SYN/RST/ACK, retransmission, payload 흐름을 낮은 층에서 봅니다. |
| DB | `EXPLAIN (ANALYZE, BUFFERS) ...`, `pg_stat_activity` | query plan, wait event, active/idle 상태를 봅니다. |

## 더 깊게 볼 기존 문서 링크

- [웹 요청 하나는 OS, 프록시, 톰캣, 스프링, DB를 어떻게 지나가나요](core-interview-guide.md)
- [파일 디스크립터는 파일이 아니라 무엇을 가리키는가](os-kernel-computer-architecture.md)
- [네트워크 패킷은 NIC에서 애플리케이션 버퍼까지 어떻게 이동하는가](os-kernel-computer-architecture.md)
- [proxy와 reverse proxy](network-web-protocols.md)
- [Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현](concurrency-async-io.md)
- [Nginx가 Tomcat을 업스트림으로 요청을 전달하고, Tomcat이 Spring 애플리케이션의 `@RestController`나 `@Controller`로 요청을 전달하는 과정](spring-backend-frameworks.md)
- [HTTP streaming](../web/http_streaming.md)
- [massive connections](../web/massive_connections.md)
- [Netty](../jvm/netty/netty.md)
- [Netty and Reactor](../jvm/netty/netty_and_reactor.md)
- [Reactor](../jvm/reactor/reactor.md)
- [Reactor Netty](../jvm/reactor/reactor_netty.md)
- [Spring WebFlux](../jvm/spring/spring_webflux.md)
- [Java socket](../jvm/java/java_socket.md)
- [Java socket connection reset](../jvm/java/java_socket_connection_reset.md)

## 근거와 참고 자료

아래 자료는 문서의 하중을 지탱하는 기술 주장에 사용한 1차 또는 공식 자료입니다. 기존 repo 문서는 주제 위치와 학습 맥락을 확인하는 내부 근거로 사용했고, 기술 사실은 가능한 한 man page, RFC, 공식 문서로 닫았습니다.

- [Linux tcp(7)](https://www.man7.org/linux/man-pages/man7/tcp.7.html)
- [Linux bind(2)](https://man7.org/linux/man-pages/man2/bind.2.html)
- [Linux listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html)
- [Linux accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html)
- [Linux select(2)](https://man7.org/linux/man-pages/man2/select.2.html)
- [Linux poll(2)](https://man7.org/linux/man-pages/man2/poll.2.html)
- [Linux epoll(7)](https://www.man7.org/linux/man-pages/man7/epoll.7.html)
- [Linux io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html)
- [Linux cgroup v2](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)
- [Linux signal(7)](https://www.man7.org/linux/man-pages/man7/signal.7.html)
- [Linux getrlimit(2)](https://www.man7.org/linux/man-pages/man2/getrlimit.2.html)
- [systemd.service(5)](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html)
- [RFC 9293 TCP](https://www.rfc-editor.org/rfc/rfc9293)
- [RFC 8446 TLS 1.3](https://www.rfc-editor.org/rfc/rfc8446)
- [RFC 9110 HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110)
- [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy)
- [Nginx proxy_buffering](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffering)
- [Tomcat architecture](https://tomcat.apache.org/tomcat-10.1-doc/architecture/overview.html)
- [Java NIO channels](https://docs.oracle.com/en/java/javase/24/docs/api/java.base/java/nio/channels/package-summary.html)
- [Netty user guide](https://netty.io/wiki/user-guide-for-4.x.html)
- [Reactor backpressure](https://projectreactor.io/docs/core/release/reference/coreFeatures/simple-ways-to-create-a-flux-or-mono-and-subscribe-to-it.html)
- [Reactive Streams JVM](https://github.com/reactive-streams/reactive-streams-jvm)
- [Spring WebFlux](https://docs.spring.io/spring-framework/reference/web/webflux.html)
- [Reactor Netty reference](https://docs.spring.io/projectreactor/reactor-netty/docs/current/reference/html/)
- [FreeBSD kqueue(2)](https://man.freebsd.org/kqueue%282%29)
- [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html)
- [PostgreSQL monitoring stats](https://www.postgresql.org/docs/current/monitoring-stats.html)
