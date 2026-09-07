# 동시성, 비동기, I/O

## 개요

서버가 여러 요청을 처리할 때는 **작업이 어디서 기다리는지**, **누가 다시 실행하는지**, **공유 상태를 누가 보호하는지**를 구분해야 합니다. 비동기 API를 썼다는 사실만으로 스레드 대기나 경쟁 조건이 사라지지는 않습니다.

예를 들어 요청 A가 외부 HTTP 응답을 기다리는 동안 요청 B를 처리할 수 있습니다. 이것은 동시성입니다. 두 요청의 계산을 서로 다른 실행 자원에서 같은 순간 수행하면 병렬성도 있습니다. A를 기다리는 방법은 별개의 선택입니다. 플랫폼 스레드를 붙잡을 수도 있고, 코루틴을 중단할 수도 있고, 콜백을 등록할 수도 있습니다.

이 문서는 짧은 답변 뒤에 실제 동작, 반례, 확인 방법을 배치했습니다. Java의 의미는 JLS와 Java SE 25 API를 기준으로 하고, Linux·libuv·JVM 구현에 따라 달라지는 부분은 따로 표시합니다. 특정 제품의 처리량이나 모든 버전의 내부 구현을 보장하는 문서는 아닙니다.

- [개요](#개요)
- [먼저 기억할 정리](#먼저-기억할-정리)
- [Blocking, non-blocking, async 구분](#blocking-non-blocking-async-구분)
    - [호출의 반환과 작업의 완료를 나누기](#호출의-반환과-작업의-완료를-나누기)
    - [Future를 쓰면 무엇이 달라지는가](#future를-쓰면-무엇이-달라지는가)
    - [HTTP 요청을 Async로 처리했을 때의 이점](#http-요청을-async로-처리했을-때의-이점)
    - [Kotlin의 async HTTP 요청과 코루틴 비교](#kotlin의-async-http-요청과-코루틴-비교)
    - [I/O 서브시스템과 단일 코어의 비동기 처리](#io-서브시스템과-단일-코어의-비동기-처리)
- [Coroutine과 협력형 실행](#coroutine과-협력형-실행)
    - [suspend는 실행 위치를 자동으로 바꾸지 않는다](#suspend는-실행-위치를-자동으로-바꾸지-않는다)
    - [협력형 스케줄링과 취소](#협력형-스케줄링과-취소)
- [Event loop와 네트워크 런타임](#event-loop와-네트워크-런타임)
    - [Netty의 이벤트 루프와 ChannelPipeline](#netty의-이벤트-루프와-channelpipeline)
    - [이벤트 루프에서 긴 작업이 위험한 이유](#이벤트-루프에서-긴-작업이-위험한-이유)
- [Goroutine과 런타임 스케줄링](#goroutine과-런타임-스케줄링)
    - [G, M, P와 blocking의 위치](#g-m-p와-blocking의-위치)
    - [Java 가상 스레드와의 비교](#java-가상-스레드와의-비교)
- [Java monitor와 wait/notify](#java-monitor와-waitnotify)
    - [같은 모니터를 쓰는 구간만 상호 배제된다](#같은-모니터를-쓰는-구간만-상호-배제된다)
    - [조건을 기다릴 때 wait가 필요한 이유](#조건을-기다릴-때-wait가-필요한-이유)
    - [notify 없이도 깨어날 수 있으며 조건은 다시 확인한다](#notify-없이도-깨어날-수-있으며-조건은-다시-확인한다)
    - [작은 실행 예제: 한 번 열리는 gate](#작은-실행-예제-한-번-열리는-gate)
    - [가시성과 JVM 구현을 분리하기](#가시성과-jvm-구현을-분리하기)
- [OS I/O multiplexing](#os-io-multiplexing)
    - [epoll은 준비된 FD를 알려 준다](#epoll은-준비된-fd를-알려-준다)
    - [LT, ET와 실제 읽기의 종료 조건](#lt-et와-실제-읽기의-종료-조건)
    - [일반 파일, AIO, io_uring은 왜 따로 보는가](#일반-파일-aio-io_uring은-왜-따로-보는가)
    - [Node.js와 libuv의 실행 경로](#nodejs와-libuv의-실행-경로)
    - [NIC에서 nginx handler까지](#nic에서-nginx-handler까지)
    - [FD 고갈, backlog, 포트 고갈을 구분하기](#fd-고갈-backlog-포트-고갈을-구분하기)
- [공유 상태와 경쟁 조건](#공유-상태와-경쟁-조건)
    - [경쟁 조건과 원자성을 작은 예로 확인하기](#경쟁-조건과-원자성을-작은-예로-확인하기)
    - [가시성은 공유 메모리에 접근할 수 있다는 것과 다르다](#가시성은-공유-메모리에-접근할-수-있다는-것과-다르다)
    - [synchronized, ReentrantLock, CAS 선택](#synchronized-reentrantlock-cas-선택)
    - [하나의 HTTP 클라이언트를 여러 스레드가 사용하는 이유](#하나의-http-클라이언트를-여러-스레드가-사용하는-이유)
- [교착과 진행 보장](#교착과-진행-보장)
    - [Coffman 조건과 실제 대기 순환](#coffman-조건과-실제-대기-순환)
    - [예방, 회피, 탐지와 복구](#예방-회피-탐지와-복구)
- [스레드 상태와 스케줄링](#스레드-상태와-스케줄링)
    - [Thread.State는 Java의 관측 모델이다](#threadstate는-java의-관측-모델이다)
- [스레드 생성과 스케줄링](#스레드-생성과-스케줄링)
    - [pthread_create 이후의 Linux 실행 흐름](#pthread_create-이후의-linux-실행-흐름)
- [면접에서 시스템 설계로 연결하기](#면접에서-시스템-설계로-연결하기)
- [검증과 원문 추적](#검증과-원문-추적)

## 먼저 기억할 정리

면접에서 먼저 말할 핵심은 다음과 같습니다.

| 질문 | 첫 답변 | 바로 이어질 확인 |
| --- | --- | --- |
| 동시성과 병렬성은? | 여러 작업의 진행 기간이 겹치는 것과 실제 실행 시점이 겹치는 것의 차이입니다. | 단일 실행 스레드에서도 I/O 대기를 겹칠 수 있는가? |
| 비동기면 논블로킹인가? | API의 완료 전달 방식과 특정 호출의 대기 여부를 분리해야 합니다. | 호출자, worker, 이벤트 루프 중 누가 기다리는가? |
| epoll이 데이터를 읽어 주나? | 준비된 FD를 알려 주며, 실제 읽기는 애플리케이션이 수행합니다. | EOF, 부분 읽기, EAGAIN을 처리하는가? |
| suspend면 스레드가 안 막히나? | 중단 가능한 함수라는 뜻이며, 내부의 블로킹 호출까지 바꾸지는 않습니다. | 실제 중단 지점과 dispatcher는 무엇인가? |
| notify하면 바로 실행하나? | 통지받은 스레드는 같은 모니터를 다시 얻고 조건을 재검사해야 합니다. | 통지한 스레드가 아직 락을 가지고 있는가? |
| volatile이면 안전한가? | 가시성·순서 보장과 복합 연산의 원자성은 다릅니다. | count++ 또는 두 필드의 불변식도 보호되는가? |

`동기화(synchronization)`는 공유 상태를 조율한다는 뜻으로도 쓰입니다. I/O의 `동기(synchronous)`와 이름이 비슷하지만 질문하는 대상이 다릅니다. 또한 동기/비동기의 용례는 API와 문헌에 따라 다르므로, 면접에서는 먼저 어느 호출 경계의 의미인지 밝히는 편이 정확합니다.

## Blocking, non-blocking, async 구분

<a id="원문-http-요청을-async로-처리했을-때의-이점"></a>
<a id="http-요청을-async로-처리했을-때의-이점-1"></a>
<a id="블로킹-스레드-감소"></a>
<a id="스레드-효율성-및-확장성-증가"></a>
<a id="시스템-응답성-향상"></a>
<a id="비용-절감"></a>
<a id="코드-가독성과-유지보수성"></a>
<a id="async-처리의-실제-예제"></a>
<a id="블로킹-http-요청"></a>
<a id="비동기-http-요청-webclient"></a>
<a id="코루틴-기반-http-요청"></a>
<a id="io-서브시스템"></a>
<a id="원문-io-서브시스템"></a>
<a id="io-서브시스템-1"></a>
<a id="원문-kotlin의-async-http-요청과-코루틴-비교"></a>
<a id="kotlin의-async-http-요청과-코루틴-비교-1"></a>
<a id="async-http-요청-예-okhttp-retrofit"></a>
<a id="논블로킹-vs-비동기-차이"></a>
<a id="원문-논블로킹-vs-비동기-차이"></a>
<a id="논블로킹-vs-비동기-차이-1"></a>
<a id="블로킹-vs-논블로킹-동기-vs-비동기-동시성-vs-병렬성-개념-정리"></a>
<a id="원문-블로킹-vs-논블로킹-동기-vs-비동기-동시성-vs-병렬성-개념-정리"></a>
<a id="블로킹-vs-논블로킹-동기-vs-비동기-동시성-vs-병렬성-개념-정리-1"></a>
<a id="비동기-프로그래밍"></a>
<a id="원문-비동기-프로그래밍"></a>
<a id="비동기-프로그래밍-1"></a>
<a id="싱글-코어--싱글-스레드-경우-커널에서의-비동기와-멀티플렉싱"></a>
<a id="원문-싱글-코어--싱글-스레드-경우-커널에서의-비동기와-멀티플렉싱"></a>
<a id="싱글-코어--싱글-스레드-경우-커널에서의-비동기와-멀티플렉싱-1"></a>

### 호출의 반환과 작업의 완료를 나누기

**blocking은 호출한 실행 흐름이 조건을 기다리며 멈출 수 있는 성질이고, async는 작업 제출과 완료 관측이 분리된 API 계약입니다.** 다른 스레드를 쓰는지는 이 정의만으로 결정되지 않습니다. 이 문서에서 동기 API는 이번 호출의 결과를 반환값이나 예외로 같은 호출 안에서 전달하는 API를 뜻합니다. non-blocking `recv()`도 이번 읽기의 바이트 수나 EOF를 반환하고, 읽을 데이터가 없으면 -1과 errno로 그 결과를 전달하므로 제출 후 별도로 완료되는 비동기 읽기와 구분됩니다.

소켓에 아직 데이터가 없다고 가정하면 차이가 선명해집니다.

1. blocking `recv()`는 데이터·종료·오류 등 반환 조건을 기다릴 수 있습니다. 그 호출을 수행한 스레드는 다음 문장을 실행하지 못합니다.
2. non-blocking `recv()`는 지금 받을 데이터가 없으면 -1을 반환하고 errno를 `EAGAIN` 또는 `EWOULDBLOCK`으로 설정합니다. 실패한 읽기 요청이 나중에 자동으로 완료되는 것은 아니며, 호출자가 준비 알림 등을 이용해 다시 읽어야 합니다.
3. 비동기 읽기 API는 작업을 제출하고 나중에 완료 결과를 받습니다. 내부 구현은 OS 완료 통지일 수도, blocking I/O를 실행하는 worker일 수도 있습니다.

non-blocking의 “즉시”는 실행 시간이 0이거나 실시간 응답 상한이 보장된다는 뜻이 아닙니다. 해당 I/O 준비 조건을 기다리지 않는다는 의미입니다. 반대로 blocking API도 데이터가 이미 있으면 기다리지 않고 반환할 수 있습니다. [recv(2)](https://man7.org/linux/man-pages/man2/recv.2.html)

### Future를 쓰면 무엇이 달라지는가

다음은 `ExecutorService executor`와 blocking 함수 `fetchData()`가 이미 있다고 가정한 **API 조합 예시**입니다.

```java
// 호출자가 다음 일을 할 수 있도록 작업을 executor에 제출합니다.
Future<String> future = executor.submit(this::fetchData);

// 결과가 필요해진 시점에 기다립니다. 이미 완료되었다면 바로 반환합니다.
String result = future.get();
```

`submit()`은 호출자와 작업 실행을 분리하지만, `fetchData()`를 실행하는 worker는 여전히 I/O에서 막힐 수 있습니다. `get()`을 호출한 스레드도 결과가 미완료라면 대기합니다. 따라서 이 코드는 “비동기 제출 후 blocking 대기”라고 설명하면 충분합니다. 서로 다른 호출을 한데 묶어 하나의 시스템 호출이 동시에 async와 blocking이라고 외울 필요는 없습니다. [Future API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/Future.html)

여러 작업을 **먼저 모두 제출한 뒤** `get()`을 순서대로 호출해도 작업의 동시 실행은 유지됩니다. 예를 들어 A와 B가 이미 실행 중이라면 A의 결과를 기다리는 동안 B도 진행할 수 있습니다. `submit(A) → get(A) → submit(B)`처럼 제출 자체를 늦추면 대기 시간을 겹치기 어렵습니다.

`CompletableFuture.thenApply()`나 `thenAccept()`도 별도 worker 실행을 보장하지 않습니다. 완료시키는 스레드 또는 다른 완료 관련 호출자가 실행할 수 있고, `...Async` 계열은 지정 executor 또는 기본 비동기 실행 시설을 사용합니다. blocking I/O를 기본 common pool에 무제한 제출하는 설계는 다른 작업까지 지연시킬 수 있습니다. [CompletableFuture API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/CompletableFuture.html)

UI 스레드에서 미완료 Future의 `get()`을 호출하면 UI가 멈출 수 있습니다. SwingWorker의 `done()`에서 `get()`을 쓰는 예시는 완료 후 호출되는 콜백이라는 조건 때문에 다릅니다. 또한 `get(timeout)`의 시간 초과는 **기다림의 종료**이며, 실행 중인 작업이 자동으로 취소되었다는 뜻은 아닙니다. [SwingWorker API](https://docs.oracle.com/en/java/javase/25/docs/api/java.desktop/javax/swing/SwingWorker.html)

### HTTP 요청을 Async로 처리했을 때의 이점

**외부 응답을 기다리는 동안 요청별 플랫폼 스레드를 계속 점유하지 않으면, 같은 스레드 수로 더 많은 대기를 관리할 수 있습니다.** 이 효과는 호출 경로의 구현에 달려 있습니다. blocking 클라이언트를 worker로 옮긴 경우에는 대기를 옮긴 것입니다.

WebFlux와 Reactor Netty를 사용하는 다음 조각은 응답을 기다리는 처리 흐름을 `Mono`로 반환합니다. `WebClient webClient`가 주입되어 있고 Spring WebFlux가 구독한다고 가정합니다.

```java
// HTTP 응답이 도착한 뒤 문자열을 만들도록 파이프라인을 반환합니다.
return webClient.get()
        .uri("https://example.com")
        .retrieve()
        .bodyToMono(String.class)
        .map(body -> "Response: " + body);
```

요청 흐름은 `파이프라인 구성 → 구독 → 연결 획득 및 전송 → 응답 도착 → map 실행 → 응답 전달`입니다. `map` 안에 JDBC 호출이나 긴 CPU 계산을 넣으면 해당 콜백의 실행 스레드를 점유합니다. `Mono.just(blockingCall())`은 `just` 호출 전에 이미 blocking 함수를 실행한다는 점도 주의해야 합니다. [Spring WebFlux의 동시성 모델](https://docs.spring.io/spring/reference/web/webflux/new-framework.html#webflux-concurrency-model)

blocking 연동이 필요한 경우에는 지연 실행과 제한된 별도 실행 자원을 함께 사용합니다. 예를 들어 Reactor의 `Mono.fromCallable(...).subscribeOn(Schedulers.boundedElastic())`은 격리 수단이지 JDBC를 논블로킹 드라이버로 바꾸는 수단이 아닙니다. 큐·동시 호출 수·시간 제한은 여전히 필요합니다. [Reactor의 blocking 호출 처리](https://projectreactor.io/docs/core/release/reference/faq.html#faq.wrap-blocking)

비동기 전환만으로 외부 서버의 응답 시간이나 CPU 계산량이 줄지는 않습니다. 확인할 것은 처리량, 지연의 상위 백분위, 이벤트 루프 지연, 대기 큐, 연결 풀 사용량입니다. 스레드 수만 줄었는데 큐가 계속 증가한다면 부하를 감당한 것이 아닙니다.

### Kotlin의 async HTTP 요청과 코루틴 비교

OkHttp의 `enqueue()`는 호출자에게 콜백 API를 제공합니다. 그러나 비동기 호출을 실행하는 dispatcher의 스레드와 연결별 내부 작업은 별도로 존재합니다. “콜백을 썼으니 어느 스레드도 네트워크에서 기다리지 않는다”는 결론은 나오지 않습니다. 공유 클라이언트, 연결 풀, HTTP/2 연결의 읽기 스레드와 호출 스레드 책임은 각각 구분해야 합니다. [OkHttp 동시성 설계](https://github.com/square/okhttp/blob/master/docs/contribute/concurrency.md)

코루틴은 이러한 결과를 기다리는 코드를 순차적인 형태로 표현할 수 있게 합니다. 콜백 기반 비동기 API를 코루틴으로 감싸도 아래쪽 I/O 구현은 그대로입니다. Ktor에서는 선택한 engine의 동작까지 확인해야 합니다. Ktor 3.x에서 응답 본문을 문자열로 읽는 조각은 `client.get(url).bodyAsText()`입니다. `get()`의 결과인 `HttpResponse`를 곧바로 `String`처럼 반환하면 안 됩니다. [Ktor 응답 처리](https://ktor.io/docs/client-responses.html)

### I/O 서브시스템과 단일 코어의 비동기 처리

I/O 서브시스템에는 파일 시스템, 페이지 캐시, 네트워크 스택, 장치 드라이버, 장치와의 전송·완료 처리 등이 포함됩니다. 일반 파일 읽기는 캐시에서 해결될 수도 있고 장치 접근을 기다릴 수도 있습니다. 네트워크 수신은 소켓 버퍼의 상태를 확인해야 합니다.

단일 코어에서도 애플리케이션의 계산과 장치의 진행을 시간상 겹칠 수 있습니다. 장치에 작업을 요청한 뒤 CPU는 다른 코드를 실행하고, 이후 커널과 런타임이 결과를 처리합니다. 커널도 CPU에서 실행되므로 애플리케이션 코드와 커널 코드가 한 CPU 실행 자원에서 동시에 실행된다는 뜻은 아닙니다.

DMA는 장치와 메모리 사이의 데이터 전송을 CPU가 바이트마다 복사하는 부담을 줄입니다. 하지만 버퍼 준비, DMA 매핑, 드라이버, 프로토콜 처리와 완료 후 작업에는 CPU가 관여합니다. 따라서 “비동기 I/O는 CPU를 사용하지 않는다”는 설명은 틀립니다. 네트워크 패킷 처리에는 인터럽트와 polling을 결합하는 Linux NAPI 같은 경로도 있습니다. [DMA API](https://docs.kernel.org/core-api/dma-api-howto.html), [NAPI](https://docs.kernel.org/networking/napi.html)

## Coroutine과 협력형 실행

<a id="kotlin-coroutine"></a>
<a id="원문-kotlin-coroutine"></a>
<a id="kotlin-coroutine-1"></a>
<a id="코루틴-제어의-양보"></a>
<a id="원문-코루틴-제어의-양보"></a>
<a id="코루틴-제어의-양보-1"></a>
<a id="코루틴을-cooperative-multitasking-이라고-하는-이유"></a>
<a id="원문-코루틴을-cooperative-multitasking-이라고-하는-이유"></a>
<a id="코루틴을-cooperative-multitasking-이라고-하는-이유-1"></a>
<a id="kotlin-coroutine-2"></a>
<a id="일반적인-java의-멀티스레딩"></a>

### suspend는 실행 위치를 자동으로 바꾸지 않는다

**Kotlin의 `suspend`는 중단할 수 있는 함수라는 표시입니다. 새 스레드 생성, 백그라운드 실행, 항상 중단됨을 뜻하지 않습니다.**

```kotlin
// 선언은 suspend지만 이 코드를 실행하는 스레드는 실제로 잠듭니다.
suspend fun blockingPause() {
    Thread.sleep(1000)
}

// kotlinx.coroutines.delay는 기다리는 동안 코루틴을 중단할 수 있습니다.
suspend fun suspendingPause() {
    delay(1000)
}
```

위 조각에서 첫 함수는 실행 스레드를 점유한 채 기다립니다. 두 번째 함수는 중단 후 나머지 실행을 이어갈 정보를 남길 수 있습니다. 이미 결과가 준비된 suspend 함수는 중단 없이 반환할 수도 있습니다. [Kotlin 코루틴 기초](https://kotlinlang.org/docs/coroutines-basics.html)

중단과 재개의 흐름을 다음처럼 따라갈 수 있습니다.

1. 코루틴이 현재 스레드에서 코드를 실행합니다.
2. 비동기 결과가 아직 없으면 이어서 실행할 위치와 필요한 지역 상태를 continuation에 보존합니다. continuation은 남은 계산을 나타내는 객체입니다.
3. 현재 호출은 중단을 알리며 돌아가고, 스레드는 다른 실행 가능한 작업을 처리할 수 있습니다.
4. 결과가 준비되면 재개 요청이 발생합니다. dispatcher와 context의 규칙에 따라 실행이 이어지며, 이전과 같은 스레드라는 보장은 없습니다.

Kotlin 컴파일러의 CPS 변환과 상태 머신은 이런 흐름을 구현합니다. 모든 suspend 호출이 반드시 실제 중단을 일으키거나, 소스 코드와 일대일로 별도 상태 객체를 만드는 것으로 외우지는 않습니다. [Kotlin 언어 명세의 비동기 프로그래밍](https://kotlinlang.org/spec/asynchronous-programming-with-coroutines.html)

### 협력형 스케줄링과 취소

한 실행 스레드를 오래 점유하는 코루틴이 양보하지 않으면, 그 스레드에서 실행되어야 하는 다른 코루틴이 늦어집니다. CPU 루프에서는 필요에 따라 `yield()`로 실행 기회를 주거나 `ensureActive()`로 취소를 확인합니다. 둘의 목적은 다릅니다. 취소 확인 자체가 스케줄러에 양보하는 것은 아닙니다.

코루틴의 협력형 취소는 OS가 그 코루틴을 실행 중인 스레드를 선점할 수 없다는 뜻이 아닙니다. OS 스케줄링과 코루틴 스케줄링은 서로 다른 계층입니다. [Kotlin 취소와 타임아웃](https://kotlinlang.org/docs/cancellation-and-timeouts.html)

| 실행 자원 | 일반적인 용도 | 주의점 |
| --- | --- | --- |
| `Dispatchers.Default` | CPU 작업 | 긴 작업의 취소 확인과 공유 상태 보호가 필요 |
| `Dispatchers.IO` | blocking I/O의 실행 | underlying 호출은 여전히 스레드를 점유하며 동시 실행 한도가 필요 |
| `Dispatchers.Main` | UI 접근 | 지원 모듈과 UI 환경이 필요하고 blocking 작업을 넣으면 UI 지연 |

`Dispatchers.IO`와 Default는 구현상 스레드를 공유할 수 있어 `withContext(IO)`가 항상 다른 스레드로 이동하지는 않습니다. 논블로킹 클라이언트의 모든 호출을 IO dispatcher로 감쌀 필요도 없습니다. [Dispatchers.IO API](https://kotlinlang.org/api/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-dispatchers/-i-o.html)

코루틴 두 개를 순서대로 호출한다고 병렬로 시작되지는 않습니다. 독립된 두 작업을 겹치려면 `coroutineScope` 안에서 `async` 등으로 두 작업을 시작한 뒤 결과를 기다리는 구조가 필요합니다. 동시에 시작할 수 있는 작업인지, 실패 시 형제 작업을 취소해야 하는지 먼저 결정합니다. `runBlocking`은 호출 스레드를 막고 코루틴 완료를 기다리는 진입 도구이므로 일반 suspend 함수 내부에서 반복적으로 사용할 도구가 아닙니다. [코루틴 조합](https://kotlinlang.org/docs/composing-suspending-functions.html)

## Event loop와 네트워크 런타임

<a id="4-netty-관련-질문"></a>
<a id="원문-4-netty-관련-질문"></a>
<a id="4-netty-관련-질문-1"></a>
<a id="질문-9-netty는-어떤-원리로-고성능-네트워크를-구현하나요"></a>
<a id="답변"></a>
<a id="질문-10-netty-파이프라인에서-데이터가-인바운드아웃바운드로-흐를-때-예외가-발생하면-어떻게-처리되나요"></a>
<a id="답변-1"></a>
<a id="netty의-이벤트-루프event-loop와-epoll-기반-논블로킹-io-구현"></a>
<a id="원문-netty의-이벤트-루프event-loop와-epoll-기반-논블로킹-io-구현"></a>
<a id="netty의-이벤트-루프event-loop와-epoll-기반-논블로킹-io-구현-1"></a>

### Netty의 이벤트 루프와 ChannelPipeline

**하나의 EventLoop가 여러 Channel의 I/O와 작업을 처리하므로, 연결마다 전용 플랫폼 스레드를 유지할 필요가 없습니다.** EventLoopGroup은 여러 이벤트 루프를 관리합니다. 등록된 Channel은 통상 하나의 EventLoop에 배정되어 이벤트를 처리하지만, 별도 executor를 지정한 handler까지 모두 같은 스레드라고 가정하면 안 됩니다. [EventLoop API](https://netty.io/4.1/api/io/netty/channel/EventLoop.html)

일반적인 서버에서는 accept용 boss group이 새 연결을 수락하고, 연결의 읽기·쓰기 처리를 worker group의 EventLoop에 맡깁니다. 이것은 흔한 구성이지, 비즈니스 로직을 매번 별도 worker pool로 보낸다는 뜻은 아닙니다.

`NioEventLoop`는 Java NIO Selector를 이용하고, native epoll transport는 JNI를 통해 Linux epoll을 이용합니다. 모든 Netty transport가 반드시 Java Selector를 거치는 것은 아닙니다. 사용할 transport와 JDK·OS 조합에 따라 경로를 확인해야 합니다. [Netty native transports](https://netty.io/wiki/native-transports.html)

```text
소켓 준비 알림
  → EventLoop가 소켓 읽기
  → ChannelPipeline의 decoder
  → 메시지 처리 handler
  → encoder 및 write 요청
  → 보낼 수 있는 만큼 쓰기, 남은 바이트는 이후 이어서 쓰기
```

ChannelPipeline은 handler들의 연결입니다. 일반적인 inbound 이벤트는 앞에서 뒤로, outbound 작업은 뒤에서 앞으로 이동합니다. handler가 다음 단계로 전달하지 않으면 그 흐름은 그곳에서 멈춥니다. 데이터가 일부만 왔을 때 decoder는 완전한 메시지를 만들 수 있을 만큼 누적해야 합니다. [ChannelPipeline API](https://netty.io/4.1/api/io/netty/channel/ChannelPipeline.html)

### 이벤트 루프에서 긴 작업이 위험한 이유

이벤트 루프가 DB 호출로 500ms 동안 멈추면 그 루프에 배정된 다른 연결도 콜백 실행이 지연될 수 있습니다. 긴 CPU 계산 역시 동일한 지연을 만들 수 있습니다. 반면 처리할 작업이 없어서 selector에서 기다리는 것은 정상적인 유휴 상태입니다. “이벤트 루프는 어떤 호출에서도 절대 blocking하지 않는다”는 설명은 이 둘을 혼동합니다.

별도 executor로 넘기면 이벤트 루프를 보호할 수 있지만, 작업 순서·공유 handler 상태·큐 크기·거절 정책을 관리해야 합니다. `ChannelPipeline` 자체가 안전하게 관리된다는 사실이 사용자 handler의 가변 필드까지 안전하게 만들지는 않습니다.

inbound 예외는 `exceptionCaught` 흐름으로 처리할 수 있습니다. outbound I/O 실패는 반환된 `ChannelFuture` 또는 promise의 실패도 확인해야 합니다. 모든 실패가 `exceptionCaught` 하나로 해결되거나, 처리되지 않은 예외가 언제나 자동으로 연결을 닫는다고 가정하면 안 됩니다. [ChannelFuture API](https://netty.io/4.1/api/io/netty/channel/ChannelFuture.html)

native transport를 사용한다고 모든 데이터 복사가 사라지지도 않습니다. `sendfile` 등 zero-copy 경로의 적용 가능성은 파일·소켓·TLS 처리 경로에 달려 있습니다. “고성능”이라는 이름보다 실제 실행 경로와 측정 결과로 판단합니다.

## Goroutine과 런타임 스케줄링

<a id="go의-goroutine은-m-대-n-모델-멀티스레딩"></a>
<a id="원문-go의-goroutine은-m-대-n-모델-멀티스레딩"></a>
<a id="go의-goroutine은-m-대-n-모델-멀티스레딩-1"></a>

### G, M, P와 blocking의 위치

**Go는 많은 goroutine을 OS 스레드들에 배치하는 M:N 모델을 사용합니다.** Go 런타임의 용어로 G는 goroutine, M은 OS 스레드, P는 Go 코드를 실행하는 데 필요한 스케줄러 자원입니다. P는 물리 CPU 코어 자체가 아닙니다. `GOMAXPROCS`는 Go 코드를 동시에 실행할 수 있는 P 수를 정하며 전체 OS 스레드 수의 상한은 아닙니다. [Go runtime 설명](https://go.dev/src/runtime/HACKING.md)

지원되는 네트워크 I/O에서 준비가 안 되면 netpoller에 대기를 맡기고 G를 park할 수 있습니다. 이때 M은 다른 G를 실행할 수 있습니다. 반면 blocking syscall에 들어가면 M 자체가 묶일 수 있으므로 런타임은 P를 다른 M에서 사용할 수 있게 조정합니다. 따라서 “goroutine이 막히면 언제나 다른 OS 스레드로 옮긴다”는 설명은 부정확합니다.

Go 1.14부터 비동기 선점이 도입되어 함수 호출 없는 긴 루프도 런타임이 선점할 수 있게 되었습니다. 다만 모든 기계어 위치를 임의로 중단하는 것은 아니며 안전 지점과 런타임 제약이 있습니다. Kotlin 코루틴의 협력형 실행과 비교할 때도 이 차이를 구분합니다. [Go 1.14 릴리스 노트](https://go.dev/doc/go1.14), [선점 구현 설명](https://go.dev/src/runtime/preempt.go)

### Java 가상 스레드와의 비교

Java 플랫폼 스레드는 일반적으로 OS 스레드와 일대일로 매핑되지만, 가상 스레드는 JVM이 carrier 플랫폼 스레드 위에서 스케줄링합니다. 지원되는 blocking I/O에서 가상 스레드를 carrier로부터 분리하면 다른 가상 스레드가 carrier를 사용할 수 있습니다. 호출 코드는 blocking 형태여도 OS 스레드 점유 양상이 달라질 수 있는 사례입니다. [Java 가상 스레드](https://docs.oracle.com/en/java/javase/25/core/virtual-threads.html)

가상 스레드는 CPU 계산 자체를 빠르게 만들지 않으며, DB 연결 수나 외부 API의 수용량도 늘리지 않습니다. 또한 JDK 21과 JDK 24 이후의 `synchronized` pinning 동작은 다릅니다. JDK 24의 JEP 491은 모니터 때문에 carrier에 고정되던 상황을 개선했습니다. 버전을 빼고 “synchronized는 가상 스레드를 항상 pin한다”고 답하지 않습니다. [JDK 24 변경 사항](https://docs.oracle.com/en/java/javase/24/migrate/significant-changes-jdk-24.html)

## Java monitor와 wait/notify

<a id="1-자바에서의-쓰레드와-모니터-개념"></a>
<a id="원문-1-자바에서의-쓰레드와-모니터-개념"></a>
<a id="1-자바에서의-쓰레드와-모니터-개념-1"></a>
<a id="10-핵심-요약-및-결론"></a>
<a id="원문-10-핵심-요약-및-결론"></a>
<a id="10-핵심-요약-및-결론-1"></a>
<a id="3-synchronized-블록과-모니터-잠금"></a>
<a id="원문-3-synchronized-블록과-모니터-잠금"></a>
<a id="3-synchronized-블록과-모니터-잠금-1"></a>
<a id="31-진입-과정"></a>
<a id="32-탈출-과정"></a>
<a id="4-wait와-notify-메서드의-메커니즘"></a>
<a id="원문-4-wait와-notify-메서드의-메커니즘"></a>
<a id="4-wait와-notify-메서드의-메커니즘-1"></a>
<a id="41-wait가-하는-일"></a>
<a id="42-notify와-notifyall"></a>
<a id="5-wait-set의-작동-원리"></a>
<a id="원문-5-wait-set의-작동-원리"></a>
<a id="5-wait-set의-작동-원리-1"></a>
<a id="6-쓰레드-상태-blocked-vs-waiting"></a>
<a id="원문-6-쓰레드-상태-blocked-vs-waiting"></a>
<a id="6-쓰레드-상태-blocked-vs-waiting-1"></a>
<a id="61-blocked-상태"></a>
<a id="62-waiting-상태"></a>
<a id="7-notify가-없으면-wait한-스레드는-어떻게-되는가"></a>
<a id="원문-7-notify가-없으면-wait한-스레드는-어떻게-되는가"></a>
<a id="7-notify가-없으면-wait한-스레드는-어떻게-되는가-1"></a>
<a id="7-결론-정리"></a>
<a id="원문-7-결론-정리"></a>
<a id="7-결론-정리-1"></a>
<a id="8-jvm-레벨에서의-synchronized-작동-방식"></a>
<a id="원문-8-jvm-레벨에서의-synchronized-작동-방식"></a>
<a id="8-jvm-레벨에서의-synchronized-작동-방식-1"></a>
<a id="81-모니터-엔터익시트enterexit"></a>
<a id="82-경량화-기법biased-locking-lightweight-locking"></a>
<a id="9-실제-동작-시나리오-예시"></a>
<a id="원문-9-실제-동작-시나리오-예시"></a>
<a id="9-실제-동작-시나리오-예시-1"></a>
<a id="쓰레드-wait와-notify-예제"></a>
<a id="원문-쓰레드-wait와-notify-예제"></a>
<a id="쓰레드-wait와-notify-예제-1"></a>
<a id="1-왜-notify가-필요하고-누가-언제-호출해야-하는가"></a>
<a id="요약"></a>
<a id="예시-흐름"></a>
<a id="2-현실-세계-프로덕션-코드-예시-taskqueue-시나리오"></a>
<a id="21-taskqueue-구현-프로덕션-코드"></a>
<a id="22-실제-dosomething-메서드-예시"></a>
<a id="3-테스트-코드-tdd-접근"></a>
<a id="4-notify는-명시적으로-호출해야-하며-누가-언제-실행할까"></a>
<a id="5-notify를-호출한-스레드는-어떻게-되나요"></a>
<a id="6-synchronizedlock-사용-시-메모리-장벽memory-barrier-적용-여부"></a>
<a id="자바-쓰레드-상태-그리고-모니터-등"></a>
<a id="원문-자바-쓰레드-상태-그리고-모니터-등"></a>
<a id="자바-쓰레드-상태-그리고-모니터-등-1"></a>
<a id="자바에서-waitnotify-패턴을-조금-더-현실감-있게-적용한프로덕션-코드-예시"></a>
<a id="원문-자바에서-waitnotify-패턴을-조금-더-현실감-있게-적용한프로덕션-코드-예시"></a>
<a id="자바에서-waitnotify-패턴을-조금-더-현실감-있게-적용한-프로덕션-코드-예시"></a>
<a id="1-핵심-개념-요약"></a>
<a id="2-예시-코드-작업-준비producer와-작업-처리consumer"></a>
<a id="21-코드-흐름-설명"></a>
<a id="3-notify는-누가-언제-실행하고-그-결과는-무엇인가"></a>
<a id="4-모든-스레드가-wait만-하고-notify를-호출하는-스레드가-없다면-발생하는-문제"></a>
<a id="5-synchronized락을-사용하면-메모리-베리어가-적용되는가"></a>
<a id="6-마지막-정리"></a>

### 같은 모니터를 쓰는 구간만 상호 배제된다

**`synchronized(lock)`은 같은 객체의 모니터를 사용하는 다른 스레드의 동기화 구간과 상호 배제됩니다.** 객체의 모든 메서드와 필드 접근을 자동으로 봉쇄하는 기능은 아닙니다.

```java
final class Counter {
    private int count;

    /** 같은 인스턴스 모니터를 사용하는 호출 사이에서 증가를 보호합니다. */
    synchronized void increment() {
        count++;
    }

    /** 증가와 같은 모니터를 얻어 현재 값을 반환합니다. */
    synchronized int value() {
        return count;
    }
}
```

인스턴스 synchronized 메서드는 `this`, static synchronized 메서드는 해당 클래스의 `Class` 객체를 잠급니다. 두 static 메서드는 같은 Class 모니터로 경합하지만 반드시 번갈아 실행되지는 않습니다. 인스턴스 모니터와 Class 모니터는 서로 다릅니다. 같은 스레드는 이미 소유한 모니터에 재진입할 수 있습니다. [JLS 17.1](https://docs.oracle.com/javase/specs/jls/se25/html/jls-17.html#jls-17.1)

락을 가진 스레드도 OS에 의해 선점될 수 있습니다. CPU를 계속 독점하는 것과 락 소유권을 유지하는 것은 다릅니다. 보호 범위를 더 깊게 복습하려면 [synchronized 범위 설계](../jvm/java/java_synchronized.md)를 참고합니다.

### 조건을 기다릴 때 wait가 필요한 이유

작업 큐가 비어 있는데 소비자가 락을 잡은 채 반복 확인하면 생산자가 그 락을 얻어 작업을 넣지 못합니다. `wait()`는 **조건 확인과 대기 전환을 같은 모니터로 조율하면서**, 생산자가 상태를 바꿀 수 있도록 그 모니터를 놓습니다.

```text
소비자: lock 획득 → 큐가 비었음 → wait로 lock 해제 및 대기
생산자: lock 획득 → 작업 추가 → notifyAll → lock 해제
소비자: lock 재획득 → 큐 재확인 → 작업 꺼내기 → lock 해제 → 처리
```

`wait`, `notify`, `notifyAll`은 대상 객체의 모니터를 소유한 스레드만 호출할 수 있습니다. 그렇지 않으면 `IllegalMonitorStateException`이 발생합니다. `wait()`는 그 객체의 모니터만 놓으며, 다른 객체에서 얻은 락까지 놓지 않습니다. 재진입으로 여러 번 얻은 해당 모니터는 대기 시 해제되고 복귀 전에 원래 획득 깊이가 복구됩니다. [Object API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/Object.html)

통지는 상태가 아닙니다. 대기자가 없을 때 한 `notify()`가 나중의 대기자를 위해 저장되지는 않습니다. 그래서 `ready`, 큐의 내용처럼 **지속되는 조건**을 같은 락으로 확인하고 바꿉니다. 생산자가 먼저 작업을 넣었다면 소비자는 조건을 보고 기다리지 않아야 합니다.

### notify 없이도 깨어날 수 있으며 조건은 다시 확인한다

`wait()`는 통지 외에 interrupt나 spurious wakeup으로도 대기에서 벗어날 수 있습니다. 시간 제한을 둔 wait에는 timeout도 있습니다. 정상 반환과 `InterruptedException` 전달 모두 모니터를 다시 얻은 뒤 이루어집니다. interrupt는 강제 종료가 아니라 중단 요청이며, 예외를 전파하거나 종료 정책에 따라 처리해야 합니다.

대기 코드를 `if` 대신 `while`로 쓰는 이유는 다음 두 가지입니다.

- 아무도 조건을 만족시키지 않았는데도 깨어날 수 있습니다.
- 정상 통지를 받았더라도 락을 다시 얻기 전에 다른 소비자가 작업을 먼저 가져갈 수 있습니다.

`notify()`는 임의의 대기 스레드 하나를 선택하고, `notifyAll()`은 대기자 모두가 조건을 다시 검사할 기회를 줍니다. 어느 쪽도 락을 즉시 넘기거나 공정한 실행 순서를 보장하지 않습니다. 조건이 여러 종류인 하나의 wait set에서 `notify()`를 쓰면 진행할 수 없는 종류의 대기자만 깨어날 위험도 고려해야 합니다. [JLS 17.2](https://docs.oracle.com/javase/specs/jls/se25/html/jls-17.html#jls-17.2)

통지할 실행 경로가 없고 조건도 계속 거짓이면 무기한 대기할 수 있습니다. 하지만 이것을 “notify 없이는 절대 깨어날 수 없다”거나 “모든 무한 대기는 순환 락 교착이다”라고 바꾸어 말하면 안 됩니다.

### 작은 실행 예제: 한 번 열리는 gate

다음은 **Java 17 이상에서 실행 가능한 학습 예제**입니다. `GateDemo.java`로 저장한 뒤 `javac GateDemo.java && java GateDemo`로 실행합니다. 기대 출력은 `opened`입니다. gate가 먼저 열려도 나중 호출자는 기다리지 않습니다. 반복 작업 큐가 아니라 한 번 열린 뒤 열린 상태를 유지하는 gate입니다.

```java
import java.util.concurrent.FutureTask;
import java.util.concurrent.TimeUnit;

/** 한 번 열리는 gate의 조건 확인과 통지를 보여 줍니다. */
public final class GateDemo {
    /** 열린 상태의 확인과 변경을 같은 모니터로 보호합니다. */
    static final class Gate {
        private boolean open;

        /**
         * gate가 열릴 때까지 기다립니다.
         * @throws InterruptedException 대기 중 중단 요청을 받으면 전달합니다.
         */
        synchronized void awaitOpen() throws InterruptedException {
            // 통지 후에도 조건이 거짓일 수 있으므로 다시 확인합니다.
            while (!open) {
                wait();
            }
        }

        /** gate를 열고, 기존 대기자에게 조건을 재확인할 기회를 줍니다. */
        synchronized void open() {
            open = true;
            notifyAll();
        }
    }

    /**
     * 작업 결과와 종료를 확인합니다.
     * @param args 사용하지 않습니다.
     * @throws Exception 작업 실패나 시간 초과를 실행 실패로 전달합니다.
     */
    public static void main(String[] args) throws Exception {
        Gate gate = new Gate();
        FutureTask<String> result = new FutureTask<>(() -> {
            gate.awaitOpen();
            return "opened";
        });
        Thread worker = new Thread(result, "gate-worker");
        worker.start();
        try {
            // gate 상태를 먼저 바꾸므로 worker가 아직 wait 전이어도 안전합니다.
            gate.open();
            System.out.println(result.get(3, TimeUnit.SECONDS));
        } finally {
            // 실패한 경우에도 예제 스레드가 무기한 남지 않도록 정리합니다.
            worker.interrupt();
            worker.join(3000);
            if (worker.isAlive()) {
                throw new IllegalStateException("worker did not terminate");
            }
        }
    }
}
```

이 예제는 `wait`를 반드시 거쳤음을 증명하지 않습니다. 실행 순서에 따라 열린 상태를 먼저 볼 수도 있습니다. **대기 경로를 검증하려면** worker의 `WAITING` 상태를 제한 시간 안에 확인한 후 gate를 열고 결과를 확인해야 합니다. 단순히 `sleep(500)`을 넣고 “충분히 기다렸으니 성공”이라고 판정하지 않습니다.

생산자·소비자 큐에는 별도의 계약이 필요합니다. 큐에서 작업을 꺼낸 뒤 락 밖에서 처리하고, “큐가 비었다”와 “처리가 끝났다”를 구분합니다. 실무에서는 수용량이 정해진 `BlockingQueue` 같은 표준 도구를 우선 검토합니다. `put/take`는 대기, `offer/poll`은 즉시 실패 또는 시간 제한 같은 선택지를 제공합니다. 종료 신호와 작업 실패 정책은 애플리케이션이 정해야 합니다. [BlockingQueue API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/BlockingQueue.html)

### 가시성과 JVM 구현을 분리하기

같은 모니터에 대한 unlock은 이후 그 모니터의 lock보다 happens-before에 있습니다. 앞선 쓰기를 뒤의 읽기에 보이게 하는 언어 수준의 순서 보장입니다. `notify()` 단독으로 모든 스레드에 값을 퍼뜨린다는 뜻도, CPU 캐시 전체를 비우고 DRAM에 모두 기록한다는 뜻도 아닙니다. 실제 구현은 CPU와 JIT 최적화에 따라 달라집니다.

synchronized **블록**은 일반적으로 `monitorenter/monitorexit` 바이트코드를 사용합니다. synchronized **메서드**는 `ACC_SYNCHRONIZED` 플래그를 통해 JVM이 진입·종료를 처리합니다. `javap -c -v`로 두 경우를 구분할 수 있습니다. [JVMS의 동기화](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-2.html#jvms-2.11.10)

HotSpot의 편향 락은 JDK 15에서 기본 비활성화·사용 중단 예고가 이루어진 과거 최적화입니다. 모든 현대 JVM이 `편향 → 경량 → 중량` 단계를 반드시 밟는다는 도식은 사용하지 않습니다. 언어 계약인 배타성·재진입·가시성과 특정 버전의 객체 헤더·락 구현을 분리해 답합니다. [JEP 374](https://openjdk.org/jeps/374)

## OS I/O multiplexing

<a id="nodejs의-싱글-스레드와-libuv의-멀티-스레딩"></a>
<a id="원문-nodejs의-싱글-스레드와-libuv의-멀티-스레딩"></a>
<a id="nodejs의-싱글-스레드와-libuv의-멀티-스레딩-1"></a>
<a id="epoll-상세"></a>
<a id="원문-epoll-상세"></a>
<a id="epoll-상세-1"></a>
<a id="libuv-aio_-io_uring"></a>
<a id="원문-libuv-aio_-io_uring"></a>
<a id="libuv-aio_-io_uring-1"></a>
<a id="네트워크-통신-시-epoll-동작-방식"></a>
<a id="원문-네트워크-통신-시-epoll-동작-방식"></a>
<a id="네트워크-통신-시-epoll-동작-방식-1"></a>
<a id="동시-요청이-많아-소켓이-몇-만-개-이상-필요한-상황에서-file-descriptor-설정을-수정하지-않은-경우"></a>
<a id="원문-동시-요청이-많아-소켓이-몇-만-개-이상-필요한-상황에서-file-descriptor-설정을-수정하지-않은-경우"></a>
<a id="동시-요청이-많아-소켓이-몇-만-개-이상-필요한-상황에서-file-descriptor-설정을-수정하지-않은-경우-1"></a>
<a id="소켓-epoll-nginx-과다한-요청"></a>
<a id="원문-소켓-epoll-nginx-과다한-요청"></a>
<a id="소켓-epoll-nginx-과다한-요청-1"></a>

### epoll은 준비된 FD를 알려 준다

**I/O multiplexing은 하나의 대기 호출로 여러 FD의 준비 상태를 관찰하는 방식입니다.** FD는 프로세스가 열린 파일·소켓 등을 참조할 때 사용하는 정수 식별자입니다. Linux의 epoll 인스턴스는 관심 목록과 준비 목록을 관리합니다.

```text
epoll_create1 → 감시 인스턴스 생성
    epoll_ctl(ADD) → 관심 FD와 이벤트 등록
    epoll_wait → 준비 이벤트 반환 또는 대기
    recv/send → 실제 I/O 결과 처리
    반복
```

`epoll_wait(..., -1)`은 준비 이벤트가 없다면 호출 스레드를 재울 수 있습니다. CPU 전체가 멈추는 것은 아니며, 다른 실행 가능한 스레드가 CPU를 사용할 수 있습니다. 이벤트가 생긴 스레드는 스케줄링되어 호출에서 돌아오고, 사용자 코드가 handler를 실행합니다. 커널이 Java 또는 JavaScript 콜백을 직접 호출하는 것은 아닙니다.

`select/poll`과 비교한 핵심은 매번 전체 감시 집합을 전달·검사하는 부담을 줄인다는 것입니다. 준비된 이벤트 k개를 반환하고 처리하는 비용까지 O(1)이 되지는 않습니다. 등록 변경, 커널의 준비 목록 관리, 실제 handler 비용도 따로 존재합니다. [epoll(7)](https://man7.org/linux/man-pages/man7/epoll.7.html)

### LT, ET와 실제 읽기의 종료 조건

LT(level-triggered)는 준비 상태가 유지되는 동안 다시 알림을 받을 수 있습니다. 기본 모드입니다. ET(edge-triggered)는 상태 변화에 대한 알림을 사용하므로 일반적으로 non-blocking FD를 사용하고 `EAGAIN`까지 읽는 패턴이 필요합니다. 일부만 읽고 다음 변화만 기다리면 이미 남아 있는 데이터를 처리하지 못할 수 있습니다.

다음은 **연결된 TCP 소켓의 ET 읽기 처리 의사 코드**입니다. 송신, accept, 오류 복구를 포함한 완성 서버가 아닙니다.

```text
반복:
    n = recv(fd, buffer)
    n > 0 이면 받은 바이트를 메시지 decoder에 전달하고 계속 읽기
    n == 0 이면 상대의 송신 종료(EOF)를 처리하고 반복 종료
    EINTR 이면 다시 시도
    EAGAIN 또는 EWOULDBLOCK 이면 지금 읽을 만큼 읽었으므로 반복 종료
    그 밖의 오류이면 오류 정책에 따라 처리하고 반복 종료
```

실제 서버에서는 한 연결만 계속 처리해 다른 연결이 굶지 않도록 처리 예산과 재실행 큐도 고려합니다. ET에서 예산 때문에 읽기를 중단한다면 해당 연결을 다시 처리할 경로가 있어야 합니다. [epoll의 starvation 주의사항](https://man7.org/linux/man-pages/man7/epoll.7.html#NOTES)

`EPOLLIN`은 HTTP 응답 전체의 완료 통지가 아닙니다. 읽기 직전에 다른 실행 흐름이 데이터를 소비할 수도 있고, EOF도 읽기로 확인해야 합니다. `EPOLLOUT`도 모든 송신 데이터가 들어갈 공간이나 상대 애플리케이션의 수신 완료를 보장하지 않습니다. 부분 쓰기와 미전송 버퍼를 관리해야 합니다. [recv(2)](https://man7.org/linux/man-pages/man2/recv.2.html), [send(2)](https://man7.org/linux/man-pages/man2/send.2.html)

일반 LT/ET 등록은 이벤트마다 새로 ADD하지 않습니다. `EPOLLONESHOT`을 사용한 경우에는 이벤트 처리 후 `EPOLL_CTL_MOD`로 재활성화합니다. `EPOLLERR/HUP`도 확인해야 하며, HUP만 보고 버퍼에 남은 데이터를 무조건 버리면 안 됩니다. [epoll_ctl(2)](https://man7.org/linux/man-pages/man2/epoll_ctl.2.html)

### 일반 파일, AIO, io_uring은 왜 따로 보는가

Linux에서 일반 파일을 epoll에 등록하려 하면 보통 `EPERM`으로 거절됩니다. `O_NONBLOCK`도 일반 파일의 저장장치 대기를 제거하는 방법이 아닙니다. 일반 파일이 poll 관점에서 준비되었다는 것과 실제 데이터가 페이지 캐시에 있어 read가 빨리 끝난다는 것은 다릅니다. EOF 역시 읽기의 정상적인 즉시 반환 결과입니다. [epoll_ctl(2)의 EPERM](https://man7.org/linux/man-pages/man2/epoll_ctl.2.html#ERRORS), [open(2)의 O_NONBLOCK](https://man7.org/linux/man-pages/man2/open.2.html)

| 인터페이스 | 애플리케이션이 받는 것 | 구분할 점 |
| --- | --- | --- |
| epoll | FD의 준비 이벤트 | 실제 read/write는 별도 |
| POSIX `aio_*` | 제출한 작업의 상태와 결과 | API 표준이며 Linux glibc는 사용자 공간 스레드 기반 구현 |
| Linux native AIO `io_submit` 등 | 커널에 제출한 작업의 완료 | POSIX `aio_read`와 같은 API가 아님 |
| io_uring | 제출 항목에 대응하는 완료 항목 | 작업 종류·커널·파일 시스템에 따라 실행 경로가 다름 |
| Windows IOCP | 연결된 handle의 I/O 완료 패킷 | epoll과 같은 준비 모델로 단정하면 안 됨 |

POSIX AIO에서는 `aio_error()`로 상태를 확인하고 완료 후 `aio_return()`으로 결과를 얻습니다. `aio_suspend()`는 미완료 작업 집합을 기다릴 수 있으며 단순한 무조건 busy polling 함수가 아닙니다. 요청 구조체와 버퍼는 완료까지 유효해야 합니다. 스택 지역 변수를 비동기 제출에 넘기고 함수가 반환되어 수명이 끝나게 하는 예제는 잘못된 코드입니다. [aio(7)](https://man7.org/linux/man-pages/man7/aio.7.html), [Windows IOCP](https://learn.microsoft.com/en-us/windows/win32/fileio/i-o-completion-ports)

io_uring의 제출 큐(SQ)는 요청을, 완료 큐(CQ)는 결과를 전달하는 공유 메모리 기반 구조입니다. 일반적으로 제출과 필요시 완료 대기에 시스템 호출을 사용합니다. SQPOLL은 커널 스레드가 제출 큐를 polling하는 별도 선택지입니다. 모든 모드에서 커널이 항상 polling하거나, 모든 I/O가 별도 worker 없이 처리된다고 가정하지 않습니다. 여러 제출자가 같은 큐를 임의로 함께 수정해도 안전하다는 의미도 아닙니다.

배치 제출 등으로 호출 오버헤드를 줄일 수 있지만, io_uring이 epoll보다 모든 부하에서 빠른 것은 아닙니다. 동일한 작업·큐 깊이·CPU 사용량·지연으로 비교해야 합니다. 완료의 의미도 작업마다 다릅니다. 예를 들어 일반 write 완료와 영속 저장 보장은 별개이며 필요한 동기화 연산을 확인해야 합니다. [io_uring(7)](https://man7.org/linux/man-pages/man7/io_uring.7.html)

### Node.js와 libuv의 실행 경로

**하나의 Node.js 이벤트 루프에서는 JavaScript 콜백을 한 번에 하나씩 실행하지만, 프로세스 전체가 스레드 하나인 것은 아닙니다.** libuv의 풀과 런타임 보조 스레드가 있으며, `worker_threads`로 다른 JavaScript 실행 스레드를 만들 수도 있습니다.

Unix 계열의 네트워크 I/O에서는 libuv가 non-blocking 소켓과 epoll/kqueue 같은 준비 알림을 조합합니다. 파일 I/O 등 일부 작업은 worker pool로 보내고 완료 콜백을 루프에서 실행합니다. Windows IOCP는 완료 모델이라는 차이가 있습니다. [libuv 설계](https://docs.libuv.org/en/v1.x/design.html)

libuv 파일 I/O의 기본 경로에는 버전 차이가 있습니다. 1.45.0부터 Linux의 일부 파일 작업에 io_uring을 사용했고, 1.49.0부터 기본은 다시 thread pool이며 io_uring SQPOLL은 명시 설정으로 활성화합니다. 실제 Node.js에 포함된 버전은 `process.versions.uv`로 확인합니다. [libuv 파일 시스템 문서](https://docs.libuv.org/en/v1.x/fs.html)

libuv의 thread pool은 기본 4개이고 프로세스의 여러 이벤트 루프에서 공유합니다. 파일 시스템 작업, `getaddrinfo/getnameinfo`, 사용자 작업 등에 사용됩니다. `UV_THREADPOOL_SIZE`는 큐 압력과 CPU·메모리 비용을 보고 정해야 합니다. [libuv thread pool](https://docs.libuv.org/en/v1.x/threadpool.html)

Node.js의 `dns.lookup()`은 시스템 이름 해석 경로를 사용하고 `dns.resolve*()`는 다른 경로를 사용하므로 모든 DNS가 같은 풀을 쓴다고 답하지 않습니다. crypto/zlib의 지원되는 비동기 API도 풀을 사용할 수 있지만 임의의 긴 JavaScript 계산이 자동으로 풀로 이동하는 것은 아닙니다. [Node.js DNS 구현 고려사항](https://nodejs.org/api/dns.html#implementation-considerations), [Node.js 이벤트 루프와 worker pool](https://nodejs.org/en/learn/asynchronous-work/dont-block-the-event-loop)

`async function`의 CPU 루프도 실행 중에는 이벤트 루프를 점유합니다. I/O가 준비되어도 콜백이 실행될 차례가 오지 않으면 응답이 늦습니다. timer는 지정한 시간에 정확히 실행된다는 보장이 아니라 실행 가능한 시점의 제약입니다. Node.js의 버전별 단계 차이까지 포함하려면 해당 버전의 이벤트 루프 문서를 확인합니다. [Node.js 이벤트 루프](https://nodejs.org/en/learn/asynchronous-work/event-loop-timers-and-nexttick)

### NIC에서 nginx handler까지

일반적인 Linux TCP 수신을 단순화하면 다음 순서입니다.

1. NIC가 패킷을 수신하고 드라이버·커널이 수신 처리를 진행합니다. DMA와 인터럽트/NAPI 등의 구현이 관여합니다.
2. 커널 TCP 스택이 연결을 식별하고 바이트의 순서와 수신 상태를 관리합니다.
3. 소켓이 읽기 가능한 상태가 되면 epoll 준비 알림으로 이어질 수 있습니다.
4. nginx worker가 스케줄링되어 `epoll_wait`에서 돌아옵니다.
5. worker가 소켓을 읽고 HTTP 파싱과 후속 처리를 수행합니다. 전체 요청이 아직 없으면 다음 수신을 기다립니다.

이벤트가 도착한다고 nginx가 직접 CPU를 선점하는 것은 아닙니다. 실행 시점은 OS 스케줄러가 결정합니다. 또한 TCP는 바이트 스트림이므로 패킷 하나, read 한 번, HTTP 요청 하나가 일대일 관계가 아닙니다. [nginx 개발 가이드](https://nginx.org/en/docs/dev/development_guide.html), [tcp(7)](https://man7.org/linux/man-pages/man7/tcp.7.html)

### FD 고갈, backlog, 포트 고갈을 구분하기

**프로세스의 FD 한도는 동시 HTTP 요청 수와 같지 않습니다.** 연결 재사용과 HTTP/2 다중화가 가능하고, 로그·파일·파이프도 FD를 사용합니다. FD는 자원 객체 자체가 아니라 그 자원을 가리키는 프로세스의 식별자입니다.

| 관측 | 우선 확인할 원인 | 자동으로 따라오지 않는 결론 |
| --- | --- | --- |
| `open/socket/accept`의 `EMFILE` | 프로세스의 `RLIMIT_NOFILE` | 시스템 메모리 부족이나 기존 read 실패 |
| `ENFILE` | 시스템 전체 열린 파일 한도 | 특정 프로세스의 soft limit만 올리면 해결 |
| 연결 지연·drop·reset | listen 대기열, accept 처리, 네트워크·설정 | FD 고갈이면 반드시 RST 발생 |
| outbound connect의 주소 할당 실패 | 사용 가능한 로컬 주소·임시 포트 등 | 수신 서버는 65,535개 연결이 상한 |

Linux에서는 실제 프로세스의 `/proc/<pid>/limits`와 `/proc/<pid>/fd`를 확인합니다. `ulimit -n`은 현재 셸의 값이며 이미 실행된 서비스의 한도를 증명하지 않습니다. systemd 서비스라면 해당 unit의 `LimitNOFILE`과 실행 프로세스에 적용된 값을 함께 봅니다. [getrlimit(2)](https://man7.org/linux/man-pages/man2/getrlimit.2.html), [systemd.exec](https://raw.githubusercontent.com/systemd/systemd/main/man/systemd.exec.xml)

FD 고갈은 **새 FD 할당**을 막습니다. 이미 열린 소켓의 read/write가 그 이유만으로 모두 실패하거나 기존 FD가 자동으로 닫히지는 않습니다. accept를 못 하면 완료 연결 대기열이 쌓일 수 있지만, 클라이언트가 timeout을 보는지 reset을 보는지는 대기열과 커널 설정 등 추가 조건에 달려 있습니다. [accept(2)](https://man7.org/linux/man-pages/man2/accept.2.html), [listen(2)](https://man7.org/linux/man-pages/man2/listen.2.html)

프로세스가 살아 있는 동안 예외 경로에서 close를 누락하면 FD 누수가 생길 수 있습니다. 프로세스가 종료되면 그 프로세스의 FD는 OS가 닫습니다. 다만 다른 프로세스가 공유·상속한 참조나 TCP의 TIME_WAIT 같은 커널 상태는 별개입니다. [exit(2)](https://man7.org/linux/man-pages/man2/_exit.2.html)

TCP 연결은 로컬·원격 주소/포트 조합으로 구분되므로 같은 서버 포트에 65,535개를 넘는 연결도 가능합니다. outbound 연결에서는 특정 소스 주소와 목적지 조합에 대해 사용할 임시 포트 범위가 제약이 될 수 있습니다. TIME_WAIT은 종료된 연결의 지연 패킷과 종료 처리를 안전하게 다루기 위한 상태이며, 항상 애플리케이션 FD를 계속 점유하는 것은 아닙니다. [tcp(7)](https://man7.org/linux/man-pages/man7/tcp.7.html)

nginx의 `worker_connections`에는 클라이언트 연결뿐 아니라 upstream 등 다른 연결도 포함됩니다. 따라서 단순히 worker 수와 이 값의 곱을 “최대 사용자 수”라고 계산하지 않습니다. FD 한도, upstream 연결, 버퍼 메모리, TLS CPU, 대역폭을 함께 확인해야 합니다. [nginx core 설정](https://nginx.org/en/docs/ngx_core_module.html#worker_connections)

## 공유 상태와 경쟁 조건

<a id="경쟁-조건-race-condition"></a>
<a id="원문-경쟁-조건-race-condition"></a>
<a id="경쟁-조건-race-condition-1"></a>
<a id="멀티스레딩에서-어떻게-읽기-전용-속성을-동시에-사용할-수-있는지"></a>
<a id="원문-멀티스레딩에서-어떻게-읽기-전용-속성을-동시에-사용할-수-있는지"></a>
<a id="멀티스레딩에서-어떻게-읽기-전용-속성을-동시에-사용할-수-있는지-1"></a>
<a id="jvm에서-클래스-로드와-초기화"></a>
<a id="cpu-관점"></a>
<a id="메모리-관점"></a>
<a id="os-관점"></a>
<a id="스레드-관리"></a>
<a id="메모리-매핑"></a>
<a id="jvm-내부-동작"></a>
<a id="http-클라이언트의-내부-동작"></a>
<a id="힙-메모리-관리"></a>
<a id="메모리-가시성"></a>
<a id="비동기-처리"></a>
<a id="실제-요청-처리-흐름"></a>
<a id="요청-처리"></a>
<a id="동시성-처리"></a>
<a id="자바-쓰레드-동기화"></a>
<a id="원문-자바-쓰레드-동기화"></a>
<a id="자바-쓰레드-동기화-1"></a>
<a id="1-자바에서의-쓰레드와-모니터-개념-2"></a>
<a id="2-쓰레드-상태states-개념"></a>
<a id="3-synchronized-블록과-모니터-잠금-2"></a>
<a id="31-진입-과정-1"></a>
<a id="32-탈출-과정-1"></a>
<a id="4-wait와-notify-메서드의-메커니즘-2"></a>
<a id="41-wait가-하는-일-1"></a>
<a id="42-notify와-notifyall-1"></a>
<a id="5-wait-set의-작동-원리-2"></a>
<a id="6-쓰레드-상태-blocked-vs-waiting-2"></a>
<a id="61-blocked-상태-1"></a>
<a id="62-waiting-상태-1"></a>
<a id="7-notify가-없으면-wait한-스레드는-어떻게-되는가-2"></a>
<a id="8-jvm-레벨에서의-synchronized-작동-방식-2"></a>
<a id="81-모니터-엔터익시트enterexit-1"></a>
<a id="82-경량화-기법biased-locking-lightweight-locking-1"></a>
<a id="9-실제-동작-시나리오-예시-2"></a>
<a id="10-핵심-요약-및-결론-2"></a>

### 경쟁 조건과 원자성을 작은 예로 확인하기

**정상 결과가 실행 순서에 의존하면 경쟁 조건을 의심해야 합니다.** Java의 data race는 같은 변수에 대한 충돌 접근 중 적어도 하나가 쓰기이고, happens-before로 순서 지어지지 않은 경우를 뜻합니다. race condition은 더 넓은 논리 오류 개념이므로 개별 접근이 안전해도 발생할 수 있습니다.

```text
초기 count = 0
A: count 읽기 → 0
B: count 읽기 → 0
A: 1 쓰기
B: 1 쓰기
최종 count = 1  (두 번 증가시켰지만 하나가 유실됨)
```

`volatile int count`도 `count++` 전체를 원자적으로 만들지 않습니다. 단일 카운터라면 `AtomicInteger.incrementAndGet()`을, 여러 값의 관계를 함께 지켜야 한다면 같은 락 등 적절한 보호 범위를 선택합니다. 예를 들어 잔액 확인과 차감을 별개의 원자 연산으로 두면 두 요청이 모두 “충분한 잔액”을 보고 차감할 수 있습니다. [AtomicInteger API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/atomic/AtomicInteger.html)

`ConcurrentHashMap`도 `containsKey → put`을 하나의 연산으로 묶어 주지 않습니다. 해당 목적에 맞는 `putIfAbsent`, `compute` 등의 계약을 사용해야 합니다. 단일 이벤트 루프에서도 `읽기 → await → 쓰기` 사이에 다른 작업이 끼어들면 논리적 경쟁이 가능합니다. “스레드가 하나”는 중단을 포함한 업무 전체가 원자적이라는 뜻이 아닙니다. [ConcurrentHashMap API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/ConcurrentHashMap.html)

### 가시성은 공유 메모리에 접근할 수 있다는 것과 다르다

다음은 **Java의 한 번만 게시하는 상황을 설명하는 조각**입니다. writer는 한 번만 실행하고, data는 게시 이후 다시 수정하지 않는다고 가정합니다.

```java
int data = 0;
volatile boolean ready = false;

// writer: 데이터 초기화를 끝낸 사실을 게시합니다.
data = 42;
ready = true;

// reader: 게시를 관측했을 때만 데이터를 사용합니다.
if (ready) {
    System.out.println(data); // 위 게시를 관측했다면 42
}
```

volatile 게시와 그 게시를 관측한 읽기 사이의 순서 보장이 data의 앞선 쓰기까지 연결합니다. ready가 일반 필드라면 같은 결론을 낼 수 없습니다. 이 예는 반복 갱신되는 여러 필드의 일관된 스냅샷이나, ready를 false로 되돌리는 재사용 프로토콜까지 보장하지 않습니다. [JLS 17.4.5](https://docs.oracle.com/javase/specs/jls/se25/html/jls-17.html#jls-17.4.5)

CPU의 캐시 일관성 프로토콜은 언어 수준의 복합 연산이나 게시 규칙을 대신하지 않습니다. MMU와 페이지 테이블도 주소 변환·접근 권한을 담당하며, 잔액 확인과 차감의 경쟁을 막는 장치가 아닙니다.

C/C++에서는 일반 공유 변수에 대한 data race가 undefined behavior가 될 수 있어 Java 예제를 기계적으로 옮기면 안 됩니다. 일반 `int`에 `atomic_store` 호출만 붙여서는 해결되지 않습니다. 공유 플래그를 `_Atomic`으로 선언하고 읽기와 쓰기 모두 그 원자 객체를 통해 수행해야 합니다. C에서는 `_Atomic` 변수의 직접 읽기와 대입도 원자 연산이며 기본적으로 sequentially consistent 순서를 따릅니다. 명시적 memory order를 지정할 때는 그에 맞는 읽기와 쓰기의 계약을 함께 확인합니다. [C11 N1570 §5.1.2.4, §7.17](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n1570.pdf)

### synchronized, ReentrantLock, CAS 선택

| 도구 | 적합한 문제 | 주의점 |
| --- | --- | --- |
| `synchronized` | 같은 불변식을 짧은 범위에서 보호 | 획득 대기를 interrupt나 timeout으로 취소하는 API가 없음 |
| `ReentrantLock` | 시간 제한, interrupt 가능한 획득, 여러 Condition | 성공한 획득마다 finally에서 unlock 필요 |
| Atomic/CAS | 한 상태 전이를 원자적으로 비교·변경 | 여러 원자 연산의 조합은 자동으로 원자적이지 않음 |
| `BlockingQueue` | 생산자·소비자 간 작업 전달 | 용량·종료·실패 정책은 별도 |

재진입은 **같은 스레드가 자신이 소유한 락을 다시 얻을 수 있음**을 뜻합니다. wait 후 재개할 수 있다는 뜻이 아닙니다. `ReentrantLock.lockInterruptibly()`는 획득 대기를 interrupt로 끝낼 수 있고, `tryLock(timeout, unit)`은 시간 제한을 둡니다. 공정 모드는 경합 시 오래 기다린 스레드를 우대하지만 OS 스케줄링의 공정성이나 모든 호출 방식의 순서를 보장하지 않습니다. 특히 인자 없는 `tryLock()`은 공정 설정을 따르지 않습니다. [ReentrantLock API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/locks/ReentrantLock.html)

`Condition.await()`도 조건을 while로 재검사해야 합니다. `notEmpty`와 `notFull`처럼 기다리는 이유를 분리할 수 있다는 점이 Object의 단일 wait set과 다릅니다. [Condition API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/locks/Condition.html)

CAS로 만드는 spin lock은 기다리는 동안 CPU를 쓰며 공정성·취소·재진입·소유권 정책까지 생깁니다. 락 보유자가 오래 멈추면 비효율적입니다. lock-free 알고리즘의 진행 보장도 모든 스레드가 유한 단계 안에 끝나는 wait-free와 다릅니다. 단순히 CAS를 썼다는 이유로 전체 알고리즘의 진행 보장을 선언하지 않습니다.

파일의 `읽기 → 변경 → 쓰기`도 전체 구간을 협력하는 프로세스 사이의 잠금 등으로 보호해야 갱신 유실을 막을 수 있습니다. 임시 파일의 원자적 rename은 부분 파일 노출을 줄일 수 있지만 그 자체로 동시 갱신의 유실을 막지는 않습니다. [rename(2)](https://man7.org/linux/man-pages/man2/rename.2.html)

### 하나의 HTTP 클라이언트를 여러 스레드가 사용하는 이유

**객체 하나에 메서드 실행 슬롯이 하나만 있는 것은 아닙니다.** 같은 클라이언트를 참조하는 여러 스레드는 각자 호출 프레임·인자·지역 상태를 가지고 같은 메서드 코드를 실행할 수 있습니다. 실제 동시 실행과 코어 배치는 OS와 런타임이 결정합니다.

공유 가능한 클라이언트는 설정, 연결 풀 등 공유 상태를 내부 계약에 따라 관리하고 요청별 상태를 분리합니다. `val client` 또는 `final client`는 참조 재대입을 막을 뿐, 연결 풀·캐시 등 객체 내부를 불변으로 만들지 않습니다. 스레드 안전성은 해당 클라이언트와 사용자 확장 코드의 계약을 확인해야 합니다. [OkHttp 클라이언트 공유 권고](https://raw.githubusercontent.com/square/okhttp/master/okhttp/src/commonJvmAndroid/kotlin/okhttp3/OkHttpClient.kt)

```text
스레드 A의 요청 상태 ─┐
스레드 B의 요청 상태 ─┼→ 공유 HttpClient → 안전하게 관리되는 연결/스트림
스레드 C의 요청 상태 ─┘
```

연결도 반드시 요청마다 하나는 아닙니다. HTTP/1.1 연결을 재사용할 수 있고, HTTP/2는 한 연결에 여러 스트림을 다중화할 수 있습니다. 응답을 올바른 요청으로 전달하는 일은 클라이언트의 요청·스트림 관리 책임이지, OS가 Java 요청 객체까지 알아서 연결해 주는 일이 아닙니다. [OkHttp 연결 관리](https://raw.githubusercontent.com/square/okhttp/master/docs/features/connections.md)

안전하게 초기화·게시한 읽기 전용 상태는 동시 읽기에 유리합니다. 그러나 생성 중 `this`가 빠져나오거나 일반 필드에 무동기 게시하면 “나중에는 읽기만 한다”는 이유로 문제가 없어지지 않습니다. JVM 클래스 초기화는 같은 클래스의 초기화를 조율하고 완료 후 사용에 필요한 순서 보장을 제공하지만, 나중의 임의 변경까지 항상 최신으로 보이게 하지는 않습니다. 클래스 정체성에는 정의한 class loader도 포함됩니다. [JLS 클래스 초기화](https://docs.oracle.com/javase/specs/jls/se25/html/jls-12.html#jls-12.4.2)

스레드가 바뀔 수 있는 코루틴·비동기 흐름에서는 요청 상태를 일반 ThreadLocal에만 넣고 언제나 같은 요청이 읽을 것이라고 가정하지 않습니다. 명시적 인자 전달이나 해당 런타임의 context 전달 기능이 필요합니다.

## 교착과 진행 보장

<a id="교착-상태-deadlock"></a>
<a id="원문-교착-상태-deadlock"></a>
<a id="교착-상태-deadlock-1"></a>

### Coffman 조건과 실제 대기 순환

**교착 상태는 작업들이 서로 필요한 자원이나 진행을 기다리느라 더 이상 진행할 수 없는 상태입니다.** 전형적인 재사용 자원의 락 교착은 다음 네 조건으로 분석합니다.

1. 상호 배제: 해당 자원을 한 번에 하나만 소유할 수 있습니다.
2. 점유와 대기: 자원을 가진 채 다른 자원을 기다립니다.
3. 비선점: 보유 자원을 강제로 회수할 수 없습니다.
4. 순환 대기: A는 B의 자원, B는 A의 자원을 기다리는 식의 순환이 있습니다.

자원 종류마다 인스턴스가 하나인 자원 할당 그래프에서는 순환이 교착 판정으로 이어집니다. 여러 인스턴스가 있으면 순환만으로 충분하지 않을 수 있습니다. 네 조건을 가능한 설계 특성으로 나열했다는 사실과 지금 교착이 발생했다는 증거도 구분합니다. [Operating System Concepts 저자 자료의 Deadlocks](https://www.os-book.com/OS10/slide-dir/index.html)

```text
A: lock1 획득 → lock2 대기
B: lock2 획득 → lock1 대기
```

두 경로 모두 `lock1 → lock2` 순서로 얻도록 통일하면 이 순환을 예방합니다. DB에서도 서로 다른 순서로 두 행을 갱신하는 트랜잭션이 같은 형태를 만들 수 있습니다. 교착이 실제 발생하는지는 실행 순서와 DB의 잠금 규칙에 달려 있습니다.

### 예방, 회피, 탐지와 복구

예방은 전역 락 순서처럼 교착의 필요 조건을 깨뜨립니다. 회피는 요청을 허용한 뒤에도 안전한 완료 순서가 있는지 판단하는 방식입니다. 은행가 알고리즘에는 각 작업의 최대 자원 요구량과 가용량 등의 정보가 필요합니다. 무조건 `false`를 반환하는 `wouldBeDeadlocked()` 함수는 알고리즘 구현이 아닙니다.

탐지는 실제 대기 관계를 보고 교착을 찾은 뒤 희생 작업을 중단·롤백하는 방식입니다. DB가 트랜잭션을 롤백할 수 있다는 것과 Java가 모니터 소유 스레드를 안전하게 강제 종료할 수 있다는 것은 다릅니다. PostgreSQL은 교착을 탐지해 한 트랜잭션을 중단하며, 애플리케이션은 실패한 트랜잭션 전체의 재시도를 검토해야 합니다. [PostgreSQL 잠금과 교착](https://www.postgresql.org/docs/current/explicit-locking.html#LOCKING-DEADLOCKS)

시간 제한은 무한 대기를 제한하지만, 실패 시 이미 가진 락을 해제하고 재시도 정책을 정해야 합니다. 모두가 같은 시점에 재시도하면 서로 양보만 반복하는 livelock이 생길 수 있습니다. 락 경쟁에서 특정 작업만 계속 밀리면 starvation입니다. 상태는 바뀌는데 유용한 일이 끝나지 않는 경우도 있으므로 CPU 사용량만으로 정상 진행을 판정하지 않습니다.

Spring의 `@Transactional`은 설정된 트랜잭션 매니저와 연결된 자원에 대한 경계를 제공합니다. 일반 `HashMap` 갱신을 자동으로 롤백하는 소프트웨어 트랜잭션 메모리도 아니며, 낙관적 충돌의 자동 재시도를 의미하지도 않습니다. 프록시 기반 호출 경계 등 실제 적용 조건도 따로 확인합니다. [Spring 선언적 트랜잭션](https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html)

스레드 풀에서도 모든 worker가 같은 풀에 제출한 자식 작업의 결과를 기다리면 자식이 실행되지 못할 수 있습니다. 이 경우 모니터의 순환만 찾는 진단으로는 부족합니다. thread dump의 대기 스택과 executor의 실행 수·큐를 함께 봅니다.

## 스레드 상태와 스케줄링

<a id="2-쓰레드-상태states-개념-1"></a>
<a id="원문-2-쓰레드-상태states-개념"></a>
<a id="2-쓰레드-상태states-개념-2"></a>

### Thread.State는 Java의 관측 모델이다

| 상태 | 의미와 대표 상황 |
| --- | --- |
| `NEW` | 생성되었지만 아직 시작하지 않은 스레드 |
| `RUNNABLE` | JVM에서 실행 중 또는 실행 가능한 상태. OS 자원을 기다리는 경우도 포함될 수 있음 |
| `BLOCKED` | synchronized 모니터 획득 또는 wait 복귀 시 모니터 재획득 대기 |
| `WAITING` | 시간 제한 없는 Object.wait, Thread.join, LockSupport.park 등의 대기 |
| `TIMED_WAITING` | sleep, 시간 제한이 있는 wait/join/park 계열 대기 |
| `TERMINATED` | 실행 종료 |

이 값은 OS 스레드 상태와 일대일 대응하지 않습니다. 예를 들어 blocking 소켓 읽기가 Java dump에서 RUNNABLE로 보일 수 있으므로 RUNNABLE을 곧 CPU 사용 중이라고 판정하지 않습니다. [Thread.State API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/Thread.State.html)

`WAITING`은 Object의 wait set만을 의미하지 않습니다. join은 다른 스레드의 종료, park는 permit 등의 조건을 기다립니다. `ReentrantLock` 경합에서 관측되는 park 기반 대기를 synchronized의 BLOCKED와 혼동하지 않습니다.

`Thread.sleep()`은 가지고 있던 모니터를 놓지 않습니다. `Object.wait()`와의 결정적인 차이입니다. `synchronized` 획득을 기다리는 스레드에 interrupt를 보낸다고 획득 대기가 취소되지는 않지만, wait/sleep/join은 각 API 계약에 따라 `InterruptedException`을 던질 수 있습니다. [Thread API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/Thread.html), [LockSupport API](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/concurrent/locks/LockSupport.html)

## 스레드 생성과 스케줄링

<a id="pthread_create-이후-커널에서-발생하는-작업들"></a>
<a id="원문-pthread_create-이후-커널에서-발생하는-작업들"></a>
<a id="pthread_create-이후-커널에서-발생하는-작업들-1"></a>

### pthread_create 이후의 Linux 실행 흐름

**Linux의 pthread_create는 C 라이브러리 API이며, 일반적인 NPTL 구현에서는 clone 계열 시스템 호출을 통해 커널 태스크를 만듭니다.** libc·커널·아키텍처에 따라 `clone` 또는 `clone3` 사용 경로가 달라질 수 있으므로 항상 하나라고 단정하지 않습니다.

1. 라이브러리가 스레드 속성과 사용자 스택, thread-local storage 등 실행에 필요한 상태를 준비합니다. 사용자 스택은 보통 라이브러리가 할당·재사용하며 호출자가 제공할 수도 있습니다.
2. clone 계열 요청으로 주소 공간·FD 테이블·신호 처리 방식 등을 공유하는 커널 태스크를 만듭니다.
3. 새 태스크는 독립적인 레지스터 문맥과 스케줄링 상태를 가지고 시작 함수를 실행합니다.
4. join 가능한 스레드는 종료 후 join 등으로 관련 자원을 회수하고, detached 스레드는 다른 수명 규칙을 따릅니다.

“부모 PCB를 복사해 TCB로 바꾼다”는 요약은 Linux 커널 태스크와 사용자 공간 스레드 자료 구조를 혼동하게 합니다. 공유 여부는 clone 플래그로 결정되며 커널 자료 구조의 단순 명칭 변경이 아닙니다. [pthread_create(3)](https://man7.org/linux/man-pages/man3/pthread_create.3.html), [clone(2)](https://man7.org/linux/man-pages/man2/clone.2.html)

같은 프로세스의 스레드는 주소 공간을 공유하므로 전환 때 다른 주소 공간으로 바꾸는 부담을 피할 수 있습니다. 그래도 레지스터 저장·복구, 스케줄러, CPU 이동과 캐시 지역성 비용은 남습니다. 다른 프로세스로 전환한다고 CPU 캐시를 모두 지우거나 FD 테이블 전체를 복사하는 것도 아닙니다. 비용 비교에는 CPU·OS·부하와 측정 조건이 필요합니다.

## 면접에서 시스템 설계로 연결하기

대기 구현을 선택했다면 다음으로 **어디까지 일을 받아들일지**를 정해야 합니다. 이벤트 루프·가상 스레드·코루틴 모두 DB 연결이나 원격 서비스의 처리량을 무한대로 만들지 못합니다.

가상의 안정된 시스템에서 초당 1,000개 작업이 완료되고 평균 체류 시간이 0.2초라면, Little의 법칙으로 평균 시스템 내 작업 수는 약 200개입니다. 이 값에는 대기와 실행이 포함되며 필요한 스레드 수를 뜻하지 않습니다. 과부하로 큐가 끝없이 증가하는 상태에 안정 상태의 수치를 그대로 적용하지 않습니다. [Little의 원 논문](https://doi.org/10.1287/opre.9.3.383)

실무 판단은 요청 입구의 동시 실행 제한, executor 큐, 연결 풀, 소켓 버퍼를 연결해서 합니다. timeout은 어디서 시작해 어디까지 적용하는지 정하고, 취소가 하위 작업까지 전달되는지도 확인합니다. backpressure가 있는 스트림이라도 blocking 외부 시스템의 처리량과 모든 중간 버퍼를 자동으로 제어해 주는 것은 아닙니다.

문서를 덮고 다음 질문에 답해 보세요.

1. blocking HTTP 함수를 `supplyAsync`로 감싼 뒤 `get()`하면 어떤 스레드들이 어디서 기다릴 수 있나요?
2. epoll이 읽기 가능하다고 했는데 HTTP 메시지를 아직 만들 수 없는 이유는 무엇인가요?
3. notify보다 먼저 도착한 작업을 소비자가 놓치지 않게 하려면 무엇을 저장해야 하나요?
4. `volatile` 카운터와 atomic 카운터는 어떤 실행 순서에서 결과가 달라지나요?
5. 요청이 느릴 때 CPU, EventLoop, worker queue, 연결 풀 중 무엇을 어떻게 구분하나요?

정답의 중심은 각각 **호출 경계**, **바이트와 메시지의 차이**, **지속되는 조건**, **복합 연산의 원자성**, **대기 위치의 관측**입니다. 이 다섯 가지로 설명할 수 있다면 라이브러리 이름이 바뀌어도 같은 판단을 다시 구성할 수 있습니다.

## 검증과 원문 추적

이 본문은 원문을 그대로 옮긴 자료가 아니라 검토 후 수정한 복습 문서입니다. 아래 보존 주석의 SHA-256은 **원문 묶음의 식별자**이며 현재 본문의 해시나 정확성 인증이 아닙니다. 이전 질문 링크는 대응하는 주제 위치로 연결됩니다. 원문은 `source/`에 남아 있으며 검토 전 표현이 포함되어 있으므로 기술 판단에는 본문의 공식 근거를 우선합니다.

GateDemo를 로컬에서 컴파일·실행했고, 대기 후 통지·먼저 열린 gate·통지 없는 interrupt·조건 불충족 시 작업 미완료를 별도로 확인했습니다. Linux 컨테이너에서는 LT의 부분 읽기 후 재알림, 버퍼 소진 후 EAGAIN, EOF 읽기, 일반 파일 등록 시 EPERM도 확인했습니다. 검증의 정확한 실행 환경과 결과는 [이번 검토 기록](../docs/works/WORK_20260907_CONCURRENCY_ASYNC_IO_REVIEW.md)에 남깁니다. ET/ONESHOT과 io_uring은 공식 계약을 확인했으며 실행 검증하지 않았습니다. 프레임워크 예제 조각과 실제 서버의 성능도 실행 검증한 것으로 간주하지 않습니다.

<!-- 원문 식별자 보존: 현재 설명은 위 본문에서 통합·수정했습니다. -->
<!-- curriculum-chunk: sha256=ba19407a52452a019f15b8d1a639fe4ec588c6cc10c49fe5dcbcfe504cdf57eb major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=HTTP 요청을 Async로 처리했을 때의 이점 sources=source/interview_questions2.md:439-541, source/interviews2.md:439-541 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=91014364d9fab87a6f61ccb6eda97432a5593d7d2479beb11f549912c396daa4 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=I/O 서브시스템 sources=source/interview_questions.md:4524-4582, source/interviews.md:4524-4582 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=0bc1c669012932f489656618f282d0df26cf0193c741b0224f33c6efe970cb91 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=Kotlin의 async HTTP 요청과 코루틴 비교 sources=source/interview_questions2.md:409-438, source/interviews2.md:409-438 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=29a146ab57b3d26f1db3c13d61aa65dd932ef48b76355a34aff2986ea856d01a major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=논블로킹 vs 비동기 차이 sources=source/interview_questions.md:4470-4523, source/interviews.md:4470-4523 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=6f99ae3cef9ed51b575f4112e6885dd1c67496a61e106dc32ae241e52e03c9c3 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리 sources=source/interview_questions.md:4155-4469, source/interviews.md:4155-4469 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=e61baa9c0afee796de80c74db875cd8798270502a2b58dc2f30e15a090b71552 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=비동기 프로그래밍 sources=source/interview_questions.md:3812-3873, source/interviews.md:3812-3873 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=f3cf345881e3dda85aa07e8cf735cc91c2c2a4c2821b187523513ac97d5263c9 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱 sources=source/interview_questions.md:4104-4154, source/interviews.md:4104-4154 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=83b85432cf78a549a38f93d84e6930c9985f9e0630c6075330de38a2c337c8d9 major=concurrency-async-io mid=Coroutine과 협력형 실행 sub=Kotlin Coroutine sources=source/interview_questions.md:3747-3770, source/interviews.md:3747-3770 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=ffbe6742cabd9b009c81d5584ca5541c668bd6e0aac7616a146862a6bc079ae2 major=concurrency-async-io mid=Coroutine과 협력형 실행 sub=코루틴 제어의 양보 sources=source/interview_questions2.md:357-408, source/interviews2.md:357-408 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=21d3928f49c0b7c1820fb38a5130d9de92f36c9f5d35820a269fe916564f3242 major=concurrency-async-io mid=Coroutine과 협력형 실행 sub=코루틴을 Cooperative Multitasking 이라고 하는 이유? sources=source/interview_questions2.md:250-356, source/interviews2.md:250-356 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=f594edbc04d54389d2f0ff4fbb940d7d898d04a74c7e888a06a818910ad5b7c4 major=concurrency-async-io mid=Event loop와 네트워크 런타임 sub=4. Netty 관련 질문 sources=source/interview_questions3.md:189-212 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=649cf073128bff3d4874f03aabb397d6d79635628df319cc500895d3d14f1408 major=concurrency-async-io mid=Event loop와 네트워크 런타임 sub=Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현 sources=source/interview_questions.md:10855-10992, source/interviews.md:10803-10940 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=c6902e7e8ee1488f24a206675e072db5143558634bb85ac1846f259b15ca0fa2 major=concurrency-async-io mid=Goroutine과 런타임 스케줄링 sub=Go의 Goroutine은 M 대 N 모델 멀티스레딩 sources=source/interview_questions.md:3735-3746, source/interviews.md:3735-3746 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=b3d7caa2f7f8d0678fa28e3765b52b5efb83e3f23ca512a3a0d2529c0f3a1633 major=concurrency-async-io mid=Java monitor와 wait/notify sub=1. 자바에서의 쓰레드와 모니터 개념 sources=source/interview_questions3.md:396-405 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=26965e7485369ce237a93404fac05c64e47f232e3eb2dd542d7c14641823122d major=concurrency-async-io mid=Java monitor와 wait/notify sub=10. 핵심 요약 및 결론 sources=source/interview_questions3.md:535-550 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=e8e1b25a1a710a3be97449a3edf42a13868ca90e6cd8f8386e686b99988d4088 major=concurrency-async-io mid=Java monitor와 wait/notify sub=3. synchronized 블록과 모니터 잠금 sources=source/interview_questions3.md:421-440 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=b9855554ec7f8548ecdfec4dec7ea48dba53f891accde28c7f65db90496a97f2 major=concurrency-async-io mid=Java monitor와 wait/notify sub=4. wait()와 notify() 메서드의 메커니즘 sources=source/interview_questions3.md:441-462 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=dc98dc5c9e117401b065ca46145124d4b24f30c8034b0ea23700a1db9a230978 major=concurrency-async-io mid=Java monitor와 wait/notify sub=5. wait set의 작동 원리 sources=source/interview_questions3.md:463-472 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=510429daa623c32b746152808abece85f146922230224c7349ebf7f93a811ab3 major=concurrency-async-io mid=Java monitor와 wait/notify sub=6. 쓰레드 상태: BLOCKED vs WAITING sources=source/interview_questions3.md:473-490 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=5bb4abec82ee2b4472b53da72b32794b50f166bdf26f85b5e1db0623bbb7c809 major=concurrency-async-io mid=Java monitor와 wait/notify sub=7. notify가 없으면 wait한 스레드는 어떻게 되는가? sources=source/interview_questions3.md:491-498 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=d8916525ad5ed1cbf51d60cdbb6bd2469af888eef730d7b00abbfbb1fb8ef8cd major=concurrency-async-io mid=Java monitor와 wait/notify sub=7. 결론 정리 sources=source/interview_questions3.md:933-947 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=06a678eddb2cae922c806faacf62f24c3f926def22401f979690f9191ccb1ad7 major=concurrency-async-io mid=Java monitor와 wait/notify sub=8. JVM 레벨에서의 synchronized 작동 방식 sources=source/interview_questions3.md:499-515 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=f0cc627e4757c059dc91c52d2e7e4557d9eb23b009c2bd6dfbfa35698f124a93 major=concurrency-async-io mid=Java monitor와 wait/notify sub=9. 실제 동작 시나리오 예시 sources=source/interview_questions3.md:516-534 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=81a56e3576b3a49162d40755c09be27585f58f529d702af36e6d3ff954e0df7b major=concurrency-async-io mid=Java monitor와 wait/notify sub=쓰레드 wait와 notify 예제 sources=source/interview_questions3.md:731-932 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=0021bab710687cef9aa4a021f615003eeb3c2129a1665f90deb1f5f0e092509f major=concurrency-async-io mid=Java monitor와 wait/notify sub=자바 쓰레드, 상태, 그리고 모니터 등 sources=source/interview_questions3.md:394-395 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=3f7a46d63875ce1f079336d99f76be43ca8b15c26d2cbb3c43740558cb1fddf1 major=concurrency-async-io mid=Java monitor와 wait/notify sub=자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시 sources=source/interview_questions3.md:551-730 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=f7f95ffe9d57af36d470d8c74dabbcc0a72646d43e7e056faffc0f03cc7c6fac major=concurrency-async-io mid=OS I/O multiplexing sub=Node.js의 싱글 스레드와 libuv의 멀티 스레딩 sources=source/interview_questions.md:8897-9127, source/interviews.md:8845-9075 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=9f6edefaad1c2f587d5d7f19477786fedf9b0c38cca16269960736f4fa108db3 major=concurrency-async-io mid=OS I/O multiplexing sub=epoll 상세 sources=source/interview_questions.md:9128-9213, source/interviews.md:9076-9161 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=3f5711fdf977cc4175636a2dc6393efe4eba9f2754124183105205148b4a0b0e major=concurrency-async-io mid=OS I/O multiplexing sub=libuv, aio_*, io_uring sources=source/interview_questions.md:3874-4103, source/interviews.md:3874-4103 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=280db8c4d50445da1912415592bf4b630881eaab1204fe71cb5d759cbc6bee9b major=concurrency-async-io mid=OS I/O multiplexing sub=네트워크 통신 시 epoll 동작 방식 sources=source/interview_questions.md:6499-6611, source/interviews.md:6499-6611 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=2c1a02421e0d534b966ab604c5e02388f2793d2a5108d0b48362f368be1d9488 major=concurrency-async-io mid=OS I/O multiplexing sub=동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우 sources=source/interview_questions2.md:542-701, source/interviews2.md:542-701 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=0190f89b047fc24a65efa059996d2d9fddd431e24923ec23c1ccbf223c3bc117 major=concurrency-async-io mid=OS I/O multiplexing sub=소켓, epoll, nginx, 과다한 요청 sources=source/interview_questions.md:7294-7330, source/interviews.md:7242-7278 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=1be2b1764a1b77373e661d7765db5573e62ab2cc4fbb4ab1d92f3783d6e69161 major=concurrency-async-io mid=공유 상태와 경쟁 조건 sub=경쟁 조건 (Race Condition) sources=source/interview_questions.md:635-1098, source/interviews.md:635-1098 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=8c0646c3c8f7ac0313a8d98d13373b16f8e4c97169740009ab0ec6bad7d91a0d major=concurrency-async-io mid=공유 상태와 경쟁 조건 sub=멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지 sources=source/interview_questions2.md:102-249, source/interviews2.md:102-249 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=212bbb821bac41121fd8d7e12be8c6fbbd28f163d24f2a056ad19e0718285fcb major=concurrency-async-io mid=공유 상태와 경쟁 조건 sub=자바 쓰레드 동기화 sources=source/interview_questions3.md:237-393 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=1c70931f2f4b3c5973967b55f6d677113322336f6857ee222f40eb43b36dff26 major=concurrency-async-io mid=교착과 진행 보장 sub=교착 상태 (Deadlock) sources=source/interview_questions.md:1099-1288, source/interviews.md:1099-1288 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=6862e633a746c7b7eb9332cad0cc0d7d2f4c50bf50108719ec80a31cd85d72ec major=concurrency-async-io mid=스레드 상태와 스케줄링 sub=2. 쓰레드 상태(States) 개념 sources=source/interview_questions3.md:406-420 -->
<!-- /curriculum-chunk -->
<!-- curriculum-chunk: sha256=1f6c20060905e8f0f49500dac33bc6e7f21447d534cb445e562fa75bffb38fc3 major=concurrency-async-io mid=스레드 생성과 스케줄링 sub=`pthread_create` 이후 커널에서 발생하는 작업들 sources=source/interview_questions.md:3771-3811, source/interviews.md:3771-3811 -->
<!-- /curriculum-chunk -->
