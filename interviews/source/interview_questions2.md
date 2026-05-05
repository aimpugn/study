# Interviews2

- [Interviews2](#interviews2)
    - [멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지](#멀티스레딩에서-어떻게-읽기-전용-속성을-동시에-사용할-수-있는지)
        - [JVM에서 클래스 로드와 초기화](#jvm에서-클래스-로드와-초기화)
        - [CPU 관점](#cpu-관점)
        - [메모리 관점](#메모리-관점)
        - [OS 관점](#os-관점)
            - [스레드 관리](#스레드-관리)
            - [메모리 매핑](#메모리-매핑)
        - [JVM 내부 동작](#jvm-내부-동작)
            - [HTTP 클라이언트의 내부 동작](#http-클라이언트의-내부-동작)
            - [힙 메모리 관리](#힙-메모리-관리)
            - [메모리 가시성](#메모리-가시성)
            - [비동기 처리](#비동기-처리)
        - [실제 요청 처리 흐름](#실제-요청-처리-흐름)
            - [요청 처리](#요청-처리)
            - [동시성 처리](#동시성-처리)
    - [코루틴을 Cooperative Multitasking 이라고 하는 이유?](#코루틴을-cooperative-multitasking-이라고-하는-이유)
        - [Kotlin Coroutine](#kotlin-coroutine)
        - [일반적인 Java의 멀티스레딩](#일반적인-java의-멀티스레딩)
    - [코루틴 제어의 양보](#코루틴-제어의-양보)
    - [Kotlin의 async HTTP 요청과 코루틴 비교](#kotlin의-async-http-요청과-코루틴-비교)
        - [Async HTTP 요청 (예: OkHttp, Retrofit)](#async-http-요청-예-okhttp-retrofit)
    - [HTTP 요청을 Async로 처리했을 때의 이점](#http-요청을-async로-처리했을-때의-이점)
        - [블로킹 스레드 감소](#블로킹-스레드-감소)
        - [스레드 효율성 및 확장성 증가](#스레드-효율성-및-확장성-증가)
        - [시스템 응답성 향상](#시스템-응답성-향상)
        - [비용 절감](#비용-절감)
        - [코드 가독성과 유지보수성](#코드-가독성과-유지보수성)
        - [Async 처리의 실제 예제](#async-처리의-실제-예제)
            - [블로킹 HTTP 요청](#블로킹-http-요청)
            - [비동기 HTTP 요청 (WebClient)](#비동기-http-요청-webclient)
        - [코루틴 기반 HTTP 요청](#코루틴-기반-http-요청)
    - [동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우](#동시-요청이-많아-소켓이-몇-만-개-이상-필요한-상황에서-file-descriptor-설정을-수정하지-않은-경우)
    - [`java -jar SpringBootApp.jar` 명령어의 실행 과정](#java--jar-springbootappjar-명령어의-실행-과정)
        - [바이너리를 프로세스로 만드는 과정](#바이너리를-프로세스로-만드는-과정)
        - [Java Virtual Machine (JVM) 시작](#java-virtual-machine-jvm-시작)
        - [JVM 내부 메모리 초기화](#jvm-내부-메모리-초기화)
        - [JAR 파일 로드](#jar-파일-로드)
        - [클래스 로더에 의해 클래스 로드](#클래스-로더에-의해-클래스-로드)
        - [`main()` 메서드 실행](#main-메서드-실행)
        - [Spring Boot의 초기화 단계](#spring-boot의-초기화-단계)
            - [(a) 스프링 애플리케이션 컨텍스트 초기화](#a-스프링-애플리케이션-컨텍스트-초기화)
            - [(b) 내장 웹 서버(Tomcat, Jetty 등) 시작](#b-내장-웹-서버tomcat-jetty-등-시작)
            - [(c) DispatcherServlet 등록](#c-dispatcherservlet-등록)
            - [(d) 애플리케이션 로직 실행 준비 완료](#d-애플리케이션-로직-실행-준비-완료)
        - [6. 운영 상태로 전환](#6-운영-상태로-전환)
        - [추가 사항](#추가-사항)
    - [수억건의 데이터가 존재하는 테이블 처리 방법](#수억건의-데이터가-존재하는-테이블-처리-방법)
        - [문제 상황](#문제-상황)
        - [핵심 제약](#핵심-제약)
        - [테이블 파티셔닝 (수동 구현)](#테이블-파티셔닝-수동-구현)
            - [수동 파티셔닝 설계](#수동-파티셔닝-설계)
            - [조회 시 파티션 통합](#조회-시-파티션-통합)
        - [데이터 아카이빙](#데이터-아카이빙)
            - [아카이브 테이블 설계](#아카이브-테이블-설계)
            - [아카이브 데이터 통합 조회](#아카이브-데이터-통합-조회)
        - [인덱스 최적화 및 테이블 정리](#인덱스-최적화-및-테이블-정리)
            - [3.1 중복 인덱스 제거](#31-중복-인덱스-제거)
        - [3.2 복합 인덱스 추가](#32-복합-인덱스-추가)
        - [점진적 테이블 분리](#점진적-테이블-분리)
            - [배치로 데이터 옮기기](#배치로-데이터-옮기기)
            - [점진적 데이터 분리 전략](#점진적-데이터-분리-전략)
    - [3억건의 데이터의 B+Tree 깊이 계산](#3억건의-데이터의-btree-깊이-계산)
        - [B+ Tree의 주요 특징](#b-tree의-주요-특징)
        - [B+ Tree의 최대 용량](#b-tree의-최대-용량)
        - [인덱싱 예시](#인덱싱-예시)
        - [깊이 계산 방법](#깊이-계산-방법)
        - [B+ Tree 깊이와 데이터 접근 시간](#b-tree-깊이와-데이터-접근-시간)
        - [추가 고려 사항](#추가-고려-사항)
    - [infix, inline, noinline, crossinline](#infix-inline-noinline-crossinline)
        - [Infix](#infix)
        - [Inline](#inline)
        - [Noinline](#noinline)
        - [Crossinline](#crossinline)
    - [스프링 부트 시작과 빈 라이프 사이클](#스프링-부트-시작과-빈-라이프-사이클)
        - [SpringApplication.run()](#springapplicationrun)
        - [스프링 빈의 라이프사이클](#스프링-빈의-라이프사이클)
    - [함수형](#함수형)
        - [함수형 프로그래밍의 이론적 개념](#함수형-프로그래밍의-이론적-개념)
        - [3.6 Either를 통한 에러 처리](#36-either를-통한-에러-처리)
            - [이론 적용](#이론-적용)
            - [실무 예제](#실무-예제)
        - [3.7 I/O 작업 제어](#37-io-작업-제어)
            - [이론 적용](#이론-적용-1)
            - [Arrow-kt에서의 I/O 모델링](#arrow-kt에서의-io-모델링)
        - [순수 함수 (Pure Function)](#순수-함수-pure-function)
        - [불변성 (Immutability)](#불변성-immutability)
        - [고차 함수 (Higher-Order Functions)](#고차-함수-higher-order-functions)
        - [일급 객체 (First-Class Citizens)](#일급-객체-first-class-citizens)
        - [Functor](#functor)
        - [Monad](#monad)
        - [Task (Deferred in Kotlin)](#task-deferred-in-kotlin)
        - [Either](#either)
        - [Arrow-kt 주요 데이터 타입 및 함수](#arrow-kt-주요-데이터-타입-및-함수)
    - [HTTP 요청이 일반적인 TCP 연결 방식과 Keep-Alive TCP 연결 방식 이루어지는 경우](#http-요청이-일반적인-tcp-연결-방식과-keep-alive-tcp-연결-방식-이루어지는-경우)
        - [서버 A에 HTTP 요청 2회 (일반적인 TCP 연결) 경우](#서버-a에-http-요청-2회-일반적인-tcp-연결-경우)
        - [서버 B에 HTTP 요청 2회 (Keep-Alive TCP 연결)](#서버-b에-http-요청-2회-keep-alive-tcp-연결)
        - [Keep-Alive 타임아웃과 클라이언트가 FIN 패킷을 보내지 않을 경우의 처리 (OS 및 소켓 관점)](#keep-alive-타임아웃과-클라이언트가-fin-패킷을-보내지-않을-경우의-처리-os-및-소켓-관점)

## 멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지

> 멀티스레딩 애플리케이션에서 공유 자원의 쓰기는 동기화가 필요한 반면, 읽기만 이뤄지는 속성에 대해서는 동기화가 필요없습니다. 가령 kotlin에서 http 클라이언트를 한번 초기화하고 여러 스레드에서 재사용하는 경우가 많은데요. jvm이 실행되고 클래스가 로드될 겁니다. 그리고 jvm은 컴파일된 .class 바이트코드로부터 클래스로드를 하게 됩니다. 그러면 어떻게 여러 스레드가 하나의 http 클라이언트 인스턴스를 공유하며 여러 요청을 처리할 수 있는지 cpu, 메모리, os, jvm 관점을 아울러서 설명해보세요.
> jvm 런타임이 프로세스에 실행되면 해당 jvm 내에서 바이트코드는 JIT로 컴파일되어 이때 실제로 기계어로 컴파일되고 실행됩니다. 그리고 http 클라이언트는 힙에 하나 존재한다고 가정하겠습니다. 그리고 이 하나의 인스턴스를 3개의 스레드들이 각각 google.com, naver.com, facebook.com에 GET 요청을 한다고 가정하겠습니다. 해당 애플리케이션이 실행되는 서버의 코어는 4개고 메모리도 충분합니다. 그러면 스레드 3개가 3개의 코어에서 실제로 병렬로 실행될 겁니다. jvm 힙 메모리에 인스턴스는 하나인데 어떻게 세 개의 코어에서 세 개의 스레드가 동시에 하나의 인스턴스를 사용하요 http 요청을 하고 응답을 처리할 수 있는지 설명해보세요.

### JVM에서 클래스 로드와 초기화

JVM은 애플리케이션 실행 시 클래스 로드 과정을 시작합니다.
이는 `.class` 파일을 읽고 메모리에 로드하는 과정을 말합니다.

클래스 로더가 클래스 파일을 찾고 로드하며, 메서드 영역(Method Area)에 클래스 정보를 저장합니다.
메서드 영역은 JVM 내부의 메모리 영역 중 하나로, 클래스 메타데이터(예: 메서드, 필드, 정적 변수, 상수 풀)를 저장합니다.

일반적으로 HTTP 클라이언트는 정적 변수로 선언되어 싱글턴 패턴을 따릅니다.
초기화 시점에는 클래스 초기화와 함께 정적 변수에 인스턴스를 생성합니다.

```kotlin
object HttpClientProvider {
    val client: HttpClient = HttpClient()
}
```

`HttpClientProvider` 클래스가 처음 참조될 때 `client` 인스턴스가 생성됩니다.
이는 JVM 클래스 로드와 초기화 과정에서 단 한 번만 실행됩니다.

JVM의 클래스 초기화는 ClassLoader에 의해 동기화되므로, 여러 스레드가 동시에 접근하더라도 클래스 초기화는 단 한 번만 안전하게 수행됩니다.

### CPU 관점

여러 스레드는 각 CPU 코어에서 병렬로 실행됩니다.

예를 들어 세 개의 스레드가 각각 http 클라이언트를 사용한다고 했을 때, 각자 다른 CPU 코어에서 실행됩니다.
각 코어는 HTTP 클라이언트 객체의 메모리를 참조하며, 필요한 데이터를 CPU 캐시에 로드합니다.

CPU는 컨텍스트 스위칭을 통해 스레드의 실행 상태를 저장하고 복원하며, 이를 기반으로 스케줄링합니다.

HTTP 클라이언트 인스턴스는 읽기 전용 속성을 가진 경우가 많습니다.
즉, 스레드가 이 인스턴스를 공유할 때, CPU 캐시에 올려진 객체를 각 스레드가 읽습니다.

CPU는 MESI(Modified, Exclusive, Shared, Invalid) 프로토콜을 사용하여 공유 메모리의 일관성을 유지합니다.
따라서 여러 스레드가 동일한 HTTP 클라이언트를 읽는 경우 캐시 일관성이 보장됩니다.
가령 세 개의 스레드가 동일한 객체를 읽는 경우, 읽기 작업은 캐시된 데이터를 사용해 빠르게 수행됩니다.

### 메모리 관점

`HttpClient` 인스턴스는 힙 메모리에 저장됩니다.
이는 JVM에서 관리하며, GC(Garbage Collector)의 대상이 됩니다.

여러 스레드가 힙 메모리에 있는 동일한 HTTP 클라이언트 인스턴스를 참조합니다.
각 스레드 스택(Stack)은 인스턴스의 참조를 유지합니다.

읽기 작업은 동기화가 필요 없으므로, 메모리 락이나 다른 동기화 메커니즘이 없어도 안전하게 작동합니다.

### OS 관점

OS는 각 스레드가 독립적으로 실행되도록 스케줄링합니다.
4개의 CPU 코어 중 3개가 스레드를 처리하며, 각 스레드는 고유한 컨텍스트를 유지합니다.

네트워크 요청은 OS의 네트워크 스택을 통해 처리됩니다.
OS는 소켓을 통해 데이터를 송수신하며, 각 스레드가 요청에 대한 응답을 독립적으로 처리하도록 보장합니다.

#### 스레드 관리

OS는 각 JVM 스레드를 시스템 스레드로 매핑합니다.
JVM의 스레드는 POSIX 스레드(pthread)와 같은 시스템 스레드로 구현됩니다.

OS 스케줄러는 각 스레드를 CPU 코어에 할당하고, 컨텍스트 스위칭을 통해 병렬성을 제공합니다.

#### 메모리 매핑

OS는 가상 메모리를 통해 JVM의 힙 메모리를 관리합니다.
HTTP 클라이언트 인스턴스는 물리 메모리에 매핑되어 여러 스레드에서 공유됩니다.

OS는 메모리 접근 충돌을 방지하기 위해 페이지 테이블과 MMU(Memory Management Unit)를 활용하며, 공유 메모리에 대한 동기화는 하드웨어와 CPU 캐시 프로토콜에 의존합니다.

### JVM 내부 동작

`HttpClient` 객체는 JVM의 힙에 저장됩니다.
이 힙 메모리는 JVM 프로세스 내에서 모든 스레드가 접근할 수 있는 공유 영역입니다.

JVM의 스레드는 OS 스레드로 매핑되며, 각각의 스레드는 자체 스택(Stack)을 가지고 있습니다.

HTTP 클라이언트를 호출하는 로직은 각 스레드 스택에서 실행되지만, 힙에 위치한 객체를 참조합니다.
예를 들어 세 개의 스레드가 `google.com`, `naver.com`, `facebook.com` 세 곳에 GET 요청을 할 때, 동일한 HTTP 클라이언트 객체를 참조할 수 있습니다.

JIT 컴파일러는 HTTP 요청과 관련된 바이트코드를 최적화하여 네이티브 기계어로 변환합니다.
각 스레드는 최적화된 기계어 코드를 실행하며, 동일한 객체를 참조하지만 객체 접근은 메모리 주소를 통해 이루어집니다.
컴파일된 코드가 *스레드 안전*하도록 설계되어 있다면, 각 요청이 독립적으로 처리됩니다.

HTTP 요청은 일반적으로 I/O 작업입니다.
`HttpClient`는 내부적으로 비동기 처리 모델을 활용하거나, 요청별로 독립적인 리소스(예: 네트워크 소켓, 버퍼)를 사용합니다.
예를 들어, 각 스레드는 자신만의 네트워크 연결을 사용하므로, 서로의 작업에 영향을 주지 않습니다.

#### HTTP 클라이언트의 내부 동작

대부분의 HTTP 클라이언트는 커넥션 풀을 사용합니다.
이는 네트워크 연결을 재사용하여 성능을 최적화하는 메커니즘입니다.

가령 세 개의 스레드는 동일한 클라이언트를 사용하지만, 요청을 처리하는 소켓 연결은 각각 독립적입니다.

커넥션 풀은 동시성을 안전하게 관리하도록 설계되어 있으며, 동기화된 데이터 구조(예: `ConcurrentLinkedQueue`)를 통해 여러 스레드에서의 접근을 처리합니다.

일부 HTTP 클라이언트는 비동기 요청 모델을 지원합니다.
이 경우, 각 요청은 이벤트 루프와 워커 스레드를 통해 병렬로 처리됩니다.
요청과 응답 간의 상태 관리는 비동기 콜백 또는 `CompletableFuture`와 같은 구조로 이루어집니다.

#### 힙 메모리 관리

HTTP 클라이언트는 힙에 위치하며 여러 스레드가 참조합니다.
힙 메모리는 Young Generation과 Old Generation으로 나뉩니다.

HTTP 클라이언트는 보통 싱글턴으로 한 번만 생성되므로 오래 유지되며 Old Generation으로 이동합니다.
이는 GC의 영향을 덜 받게 합니다.

#### 메모리 가시성

JVM은 힙 메모리 접근이 안전하게 이루어지도록 Java Memory Model (JMM) 규칙을 따릅니다.
JMM은 쓰기 작업에서 volatile 또는 synchronized와 같은 메커니즘을 통해 메모리 가시성을 보장합니다.
그러나 HTTP 클라이언트는 주로 읽기 작업이 이루어지므로, 동기화 오버헤드가 필요하지 않습니다.

HTTP 클라이언트 객체는 힙 메모리에 존재하며, 이는 여러 스레드가 공유할 수 있습니다.

`HttpClient`와 같은 정적 변수는 클래스 초기화 과정에서 안전하게 초기화되며, happens-before 관계에 의해 다른 스레드에서 항상 최신 상태로 보입니다.

정적 변수는 JVM의 메서드 영역(Method Area)에서 관리되며, 이는 모든 스레드가 접근할 수 있습니다.

#### 비동기 처리

HTTP 클라이언트는 종종 비동기 요청을 처리하도록 설계됩니다.
예를 들어, 비동기 요청을 처리할 때 이벤트 루프와 워커 스레드를 사용할 수 있습니다.

JVM은 비동기 작업을 위해 ForkJoinPool이나 사용자 정의 Executor를 활용합니다.

### 실제 요청 처리 흐름

#### 요청 처리

1. 스레드는 HTTP 클라이언트의 메서드를 호출하여 요청을 보냅니다.
2. HTTP 클라이언트는 내부적으로 연결 풀(Connection Pool)을 활용하여 네트워크 연결을 관리합니다.
3. 각 요청은 비동기로 처리되며, 네트워크 I/O는 OS의 커널 네트워크 스택에서 이루어집니다.
4. OS는 요청을 적절한 네트워크 인터페이스로 전달하며, 응답은 다시 JVM으로 반환됩니다.

#### 동시성 처리

HTTP 클라이언트는 내부적으로 스레드 안전성을 보장하는 구조(예: 동기화된 데이터 구조 또는 CAS 연산)를 사용합니다.

요청 간 상태는 분리되어야 하므로, 상태 저장이 필요한 경우 `ThreadLocal` 또는 명시적 동기화를 통해 처리합니다.

## 코루틴을 Cooperative Multitasking 이라고 하는 이유?

### Kotlin Coroutine

코루틴은 경량 쓰레드(lightweight thread)로, *협력적(cooperative)*으로 작업을 스케줄링합니다.
- 코루틴은 실제 OS 스레드가 아니며, JVM의 단일 스레드에서 여러 코루틴이 협력하여 실행됩니다.
- 명시적으로 `suspend` 키워드나 특정 지점에서 작업을 양보하며 다른 코루틴이 실행되도록 합니다.
- 컨텍스트 스위칭이 스레드 수준이 아니라 코루틴 수준에서 이루어집니다.

코루틴은 명시적으로 자신의 실행을 중단(`suspend`)하고, 다른 작업이 실행되도록 협력합니다.
이는 OS 스케줄러에 의한 선점형(Preemptive) 스레드 스케줄링과 달리, 프로그램 코드 수준에서 명시적으로 중단을 관리하는 방식입니다.
- 중단 지점(Suspension Point): `suspend` 키워드가 있는 함수나 작업은 실행 중간에 상태를 저장하고, 다음 실행 흐름으로 제어권을 넘깁니다.
- 상태 머신(State Machine):
    Kotlin 컴파일러는 코루틴을 상태 머신으로 변환하여 실행 흐름을 관리합니다.
    각 `suspend` 호출은 상태 전이를 나타냅니다.

    코루틴은 컴파일러가 코드를 변환하여 상태 머신으로 구현됩니다.
    이는 코루틴의 각 `suspend` 지점이 상태(state)로 변환되고, 실행 흐름이 상태 간 전환으로 표현되는 것을 의미합니다.

    상태 머신은 코루틴의 실행 흐름을 관리하며, 각 `suspend` 호출은 상태 전환을 유발합니다.

    ```kotlin
    suspend fun exampleCoroutine() {
        println("Step 1")
        delay(1000) // Suspension Point
        println("Step 2")
    }
    // 위 코드는
    // 아래와 같이 상태 머신으로 변환 됩니다.
    class ExampleCoroutine : Continuation<Unit> {
        var state = 0
        override fun resumeWith(result: Result<Unit>) {
            when (state) {
                0 -> {
                    println("Step 1")
                    state = 1
                    delay(1000, this) // Suspension Point
                }
                1 -> {
                    println("Step 2")
                }
            }
        }
    }
    ```

    `delay`는 코루틴을 일시 중단(suspend)하고, 나중에 재개(resume)될 수 있도록 `Continuation` 객체에 상태를 저장합니다.

    Kotlin의 코루틴은 Continuation-passing style (CPS)을 활용합니다.
    모든 suspend 함수는 컴파일러에 의해 암묵적으로 `Continuation` 객체를 받는 형태로 변환됩니다.
    `Continuation` 객체는 다음 실행 지점(상태)을 기억하고, 나중에 `resumeWith` 메서드를 호출하여 실행을 이어갑니다.

- 비동기 처리: 코루틴은 CPU를 점유하지 않고, 비동기적으로 작업을 대기하거나 다른 작업으로 전환됩니다.

코루틴은 스레드 위에서 실행됩니다.
즉, 코루틴의 실행은 하나 이상의 스레드에서 처리됩니다.

코루틴의 상태는 `Continuation` 객체에 저장됩니다.
`suspend` 함수는 호출된 지점에서 상태를 저장하고, 다음 작업을 스케줄링합니다.
작업 간 컨텍스트 스위칭은 매우 가볍고, 특정 지점에서 명시적으로 이루어집니다.

```kotlin
suspend fun performTask() {
    println("Task start on ${Thread.currentThread().name}")
    delay(1000) // Suspend here and allow other tasks to run
    println("Task end on ${Thread.currentThread().name}")
}
```

코루틴은 작업을 협력적으로 양보해야 다른 코루틴이 실행됩니다.
명시적으로 `delay`, `yield`, `withContext` 등을 사용해 CPU를 다른 작업으로 넘깁니다.
Dispatcher를 통해 코루틴의 실행 스레드를 관리합니다:
- `Dispatchers.Default`: 백그라운드 스레드 풀에서 실행.
- `Dispatchers.Main`: UI 스레드에서 실행 (Android).
- `Dispatchers.IO`: I/O 최적화 스레드 풀에서 실행.

### 일반적인 Java의 멀티스레딩

일반적인 Java의 멀티스레딩은 OS에서 제공하는 물리적 스레드를 사용하며, *선점형(preemptive)*으로 스케줄링됩니다.
- JVM 스레드는 OS 스레드로 매핑되며, 스레드는 OS 커널에 의해 스케줄링됩니다.
- 작업 간의 컨텍스트 스위칭은 OS가 자동으로 관리합니다.
- 각 스레드는 자체 스택과 메모리를 가지며, 독립적으로 실행됩니다.

Java 스레드는 OS에서 직접 관리되며, 각 스레드는 독립적으로 실행됩니다.
작업 간의 컨텍스트 스위칭은 OS 커널에 의해 자동으로 이루어집니다.
작업은 특정 시점에서 중단되고 다른 스레드가 실행될 수 있습니다.

```java
public class Task extends Thread {
    @Override
    public void run() {
        System.out.println("Task start on " + Thread.currentThread().getName());
        try {
            Thread.sleep(1000); // Pause thread
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        System.out.println("Task end on " + Thread.currentThread().getName());
    }
}
```

OS 커널은 작업을 선점적으로 스케줄링합니다.
실행 중인 스레드는 중단되고 다른 스레드가 실행될 수 있습니다.
Thread Scheduler는 CPU 시간을 각 스레드에 할당합니다.
사용자는 스레드 우선순위(`Thread.setPriority`)를 설정할 수 있지만, OS 스케줄러의 결정에 의존합니다.

## 코루틴 제어의 양보

코루틴은 상태 머신(State Machine)을 통해 비동기 실행을 효율적으로 관리하며, 실행 중인 제어를 다른 코루틴에게 양보하는 것이 가능합니다.
- 코루틴은 단일 스레드 내에서 실행 중인 작업을 중단하고, 다른 작업(코루틴)이 실행될 수 있도록 제어권을 넘기는 방식으로 동작합니다.
- 제어의 양보는 비선점형 협력(cooperative multitasking)의 한 형태로, 명시적인 지점에서 실행 흐름이 중단됩니다.

제어 양보의 원리는 다음과 같습니다:
1. 코루틴의 상태 머신 설계:
    - 컴파일러는 코루틴을 상태 머신으로 변환합니다.
    - 상태 머신은 각 실행 지점(중단점)을 상태로 모델링하며, 다음 상태로 이동할 수 있는 로직을 포함합니다.
    - 이 구조는 `suspend` 키워드가 사용된 지점을 상태 변경의 기준으로 삼습니다.

    예제 (단순 코루틴):

    ```kotlin
    suspend fun fetchData() {
        println("Fetching data...")
        delay(1000)  // 중단점 (suspend)
        println("Data fetched!")
    }
    ```

    - `fetchData()`는 내부적으로 상태 머신으로 변환됩니다.
    - `delay(1000)`은 제어권을 현재 코루틴에서 반환하도록 지시하고, 특정 상태로 복귀 가능한 지점을 기록합니다.

2. 코루틴 디스패처와 실행 컨텍스트:
    - 코루틴 디스패처는 코루틴의 실행을 관리합니다. 주요 디스패처:
        - `Dispatchers.Default`: 백그라운드에서 CPU 집약적인 작업 실행.
        - `Dispatchers.IO`: I/O 작업 처리.
        - `Dispatchers.Main`: UI 업데이트 작업.
    - 실행 중단(`suspend`) 시, 디스패처가 제어권을 받아 다른 대기 중인 작업을 실행하거나 리소스를 다른 작업에 재할당합니다.

3. `suspend`와 컨티뉴에이션(Continuation):
    - `suspend` 함수는 중단 가능한 지점을 정의하며, 호출 스택 대신 컨티뉴에이션 객체를 생성해 호출 상태를 저장합니다.
    - 컨티뉴에이션은 "어디에서부터 다시 실행을 시작할지"를 나타냅니다.

    컨티뉴에이션 구조:

    ```kotlin
    suspend fun example() {
        println("Start")
        delay(1000) // `delay` 호출 시
                    // - 현재 상태를 저장하고,
                    // - 디스패처에게 제어권을 넘깁니다.
        println("End")
    }
    ```

4. 실제 동작 예:
    - `delay`나 `yield`와 같은 함수 호출은 스레드를 차단(blocking)하지 않습니다.
    - 대신, 현재 상태를 기록하고 다른 코루틴에게 실행 기회를 제공합니다.

## Kotlin의 async HTTP 요청과 코루틴 비교

### Async HTTP 요청 (예: OkHttp, Retrofit)

1. 구조:
    - 전통적인 async HTTP 요청은 스레드 풀을 활용하여 비동기 작업을 수행합니다.
    - 콜백 기반 구조로 결과를 처리하며, 스레드가 응답을 기다리는 동안 다른 작업을 처리할 수 있습니다.

    예제 (OkHttp):

    ```kotlin
    val client = OkHttpClient()
    val request = Request.Builder().url("https://example.com").build()

    client.newCall(request).enqueue(object : Callback {
        override fun onResponse(call: Call, response: Response) {
            println("Response: ${response.body()?.string()}")
        }

        override fun onFailure(call: Call, e: IOException) {
            println("Error: ${e.message}")
        }
    })
    ```

2. 특징:
    - 요청마다 별도의 스레드를 생성하거나 스레드 풀에서 작업을 처리.
    - 결과는 콜백 함수로 전달됨.
    - 콜백 지옥(Callback Hell) 문제로 코드가 복잡해질 수 있음.

## HTTP 요청을 Async로 처리했을 때의 이점

### 블로킹 스레드 감소

- 블로킹 모델:
    - 스레드 기반 요청 처리에서는 HTTP 요청 응답을 기다리기 위해 스레드가 블로킹 상태로 대기합니다.
    - 이는 스레드가 다른 작업을 수행하지 못하고, 리소스를 비효율적으로 사용하게 만듭니다.

- 비동기 모델:
    - HTTP 요청을 Async로 처리하면, 스레드가 블로킹 상태에 머물지 않고, 다른 요청을 처리하거나 반환될 수 있습니다.
    - 실제 HTTP 요청이 완료되면 콜백이나 코루틴을 통해 결과를 받아 후속 작업을 이어갑니다.
    - 이점:
        - 더 적은 스레드로 더 많은 요청을 처리 가능.
        - I/O 대기 시간이 긴 애플리케이션에서 특히 효과적.

### 스레드 효율성 및 확장성 증가

- 스레드 기반 처리는 요청 수가 많아질수록 스레드 풀이 포화 상태에 이르러 성능이 급격히 저하될 수 있습니다.
- 비동기 처리에서는 요청 대기 시간을 비동기 이벤트 루프(예: Netty, Project Reactor)로 처리하므로 스레드 수 증가 없이 높은 요청 처리량을 유지할 수 있습니다.
- 예: 100개의 요청을 처리할 때, 10개의 스레드만으로도 충분히 처리 가능.

### 시스템 응답성 향상

- 블로킹 작업이 줄어들면서 스레드 풀이 더 빠르게 반환되어 새로운 요청에 대해 더 빠르게 반응할 수 있습니다.
- 예: 사용자 A의 요청이 비동기로 처리되어 스레드를 반환하면, 사용자 B의 요청에 즉시 대응 가능.

### 비용 절감

- 스레드는 메모리와 CPU를 소비하는 고비용 리소스입니다.
- 비동기 처리는 적은 수의 스레드로 더 많은 요청을 처리할 수 있으므로, 하드웨어 요구사항이 줄어들어 인프라 비용을 절감할 수 있습니다.

### 코드 가독성과 유지보수성

- 비동기 처리를 코루틴이나 `CompletableFuture`와 같은 고수준 API로 구현하면, 코드가 읽기 쉽고 유지보수하기 쉬워집니다.
- 스프링 웹플럭스(Spring WebFlux)와 같은 프레임워크는 코루틴과의 통합으로 더욱 간결한 비동기 코드를 제공합니다.

### Async 처리의 실제 예제

#### 블로킹 HTTP 요청

```java
@RestController
public class BlockingController {

    @GetMapping("/blocking")
    public String blockingHttpRequest() {
        RestTemplate restTemplate = new RestTemplate();
        // `restTemplate`은 블로킹 방식으로 HTTP 요청을 처리합니다.
        // 스레드는 응답이 완료될 때까지 대기.
        String response = restTemplate.getForObject("https://example.com", String.class);
        return "Response: " + response;
    }
}
```

#### 비동기 HTTP 요청 (WebClient)

```java
@RestController
public class AsyncController {

    private final WebClient webClient = WebClient.create();

    @GetMapping("/async")
    public Mono<String> asyncHttpRequest() {
        // `WebClient`는 비동기 논블로킹 HTTP 클라이언트로 작동.
        // 요청 대기 동안 스레드는 반환되어 다른 작업을 처리할 수 있음.
        // 응답이 준비되면 비동기로 후속 작업 수행.
        return webClient.get()
            .uri("https://example.com")
            .retrieve()
            .bodyToMono(String.class)
            .map(response -> "Response: " + response);
    }
}
```

### 코루틴 기반 HTTP 요청

1. 구조:
    - 코루틴은 `suspend` 키워드를 통해 비동기 작업을 처리하며, 스레드 차단 없이 실행 중단과 재개를 관리합니다.
    - 코루틴은 명시적 상태 머신으로 작동하며, 내부적으로 스레드를 효율적으로 재활용합니다.

    예제 (Ktor HTTP Client):

    ```kotlin
    val client = HttpClient()

    suspend fun fetchContent(): String {
        return client.get("https://example.com")
    }

    runBlocking {
        val content = fetchContent()
        println("Response: $content")
    }
    ```

2. 특징:
    - 비동기 작업을 선형적으로 표현 가능하여 코드 가독성이 뛰어남.
    - 호출이 중단되더라도 작업 상태를 유지하여 중단점 이후에 다시 실행.
    - 콜백 대신 `try-catch`로 에러 처리.

## 동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우

Linux에서 기본 사용자별 파일 디스크립터 제한(Soft Limit)은 보통 1024개 또는 4096개로 설정됩니다.
한 프로세스에서 열 수 있는 파일 디스크립터의 수를 초과하는 요청(예: 10,000개 이상)이 들어오면, 추가적인 파일 디스크립터 할당이 불가능해집니다.

1. 새로운 소켓 생성 실패:
    - 네트워크 요청은 각각 하나의 소켓을 필요로 하며, 소켓은 파일 디스크립터로 관리됩니다.
    - 요청 수가 파일 디스크립터 한도를 초과하면, `socket()` 시스템 호출이 실패합니다.
    - 오류 코드: `EMFILE` (Too many open files).

2. 요청 처리 실패:
    - 서버 애플리케이션은 더 이상 새로운 요청을 수락하지 못합니다.
    - 클라이언트는 연결 시도에서 타임아웃 또는 "연결 거부(Connection Refused)" 오류를 경험합니다.

3. 부분적인 요청 처리:
    - 초과된 요청이 대기 상태로 남아 처리되지 않고, 서버의 응답 속도가 점점 느려지거나 특정 요청은 영구적으로 무시됩니다.

자원의 고갈에 따라 다음과 같은 시스템 동작이 발생합니다.
- CPU 과부하
    - 새로운 네트워크 요청이 들어오면, 네트워크 스택(Network Stack)에서 이를 처리하기 위해 TCP/IP 프로토콜 스택을 실행합니다.
    - SYN 패킷을 받아들인 후 SYN 큐에 연결 상태를 기록하며 클라이언트에게 SYN-ACK를 전송합니다.
    - SYN 큐는 고정된 크기로, 기본값은 보통 1024개 정도로 설정됩니다.
    - 한계 초과한 요청으로 네트워크 큐(예: SYN 큐)가 가득 차고 새로운 SYN 패킷을 추가할 공간이 없습니다.
        - 이를 무시하거나 패킷을 드롭(drop)합니다.
        - 클라이언트는 "연결 거부(Connection Refused)" 또는 연결 타임아웃을 경험합니다.
    - 클라이언트는 연결 실패를 감지하고 재시도 요청을 보냅니다.
        - 이 재시도 요청으로 서버는 동일한 SYN 패킷을 계속 처리해야 하며, SYN 큐는 가득 차 있으므로 이 작업이 반복됩니다.
        - 결과적으로 TCP/IP 스택에서의 작업이 폭증하고, CPU 컨텍스트 스위칭이 급격히 증가합니다.
        - CPU 코어는 다른 중요한 작업을 처리하지 못하고, 네트워크 요청 처리에만 몰두하게 됩니다.

- 메모리 부족
    - 각 파일 디스크립터는 OS가 관리하는 작은 커널 객체입니다. 이를 통해 파일, 소켓, 파이프 등 다양한 리소스를 식별합니다.
    - 파일 디스크립터와 소켓 구조체는 커널 메모리 내에서 관리됩니다. TCP 연결 하나마다:
        - 소켓 구조체(struct socket)를 생성하고,
        - 네트워크 상태를 추적하기 위한 추가적인 메모리를 할당하고,
        - 송수신 데이터 버퍼를 생성합니다.
    - 기본 설정(예: 사용자별 1024개의 파일 디스크립터 제한)을 초과하는 요청이 들어오면, OS는 더 이상 새로운 소켓 구조체를 생성할 수 없습니다.
        - 동시 열리는 파일 디스크립터가 많아지면 커널 메모리 사용량이 증가하고, 다른 작업에 필요한 메모리가 부족해질 수 있습니다.
        - 이미 생성된 소켓 구조체와 데이터 버퍼가 커널 메모리를 차지하며, 메모리가 부족해집니다.
            - Cannot allocate memory 오류 발생.
            - 메모리 부족으로 인해 기존 작업들도 실패하기 시작.

- 네트워크 큐의 포화
    - TCP 소켓 대기열(backlog)
        - 새로운 연결 요청은 소켓의 대기열(예: `backlog`)에 추가됩니다.
        - 파일 디스크립터 한계를 초과하면 대기열이 가득 찹니다. 커널은 더 이상 요청을 대기열에 넣지 못하고 새 연결 요청을 보류합니다.
        - 요청 처리 속도가 느려지고 대기열이 가득 차면:
            - 새로운 요청은 대기열에 추가되지 못하고, 클라이언트에게 RST(Reset) 패킷을 전송하여 연결을 종료시킵니다.
            - 클라이언트는 "연결 거부(Connection Refused)"를 받습니다.

    - 대기열 초과 시 OS의 동작
        - 기존 연결 중 일부는 여전히 활성화 상태이나, 새로운 연결을 추가할 공간이 없습니다.
        - OS는 대기열 초과 상태를 로그에 기록합니다

            ```plaintext
            TCP: drop open requests due to syn backlog full
            ```

    - 클라이언트 측 영향
        - 클라이언트는 TCP 연결을 시도하지만, 서버는 `SYN-ACK`를 응답하지 않거나, "연결 거부(Connection Refused)"를 반환합니다.
        - 클라이언트는 타임아웃으로 인해 지연을 겪습니다.

- 파일 디스크립터 누수 (FD Leak)
    - 애플리케이션이 비정상 종료하거나, 요청 처리 중 예외가 발생했을 때,
        - 열린 파일 디스크립터를 닫지 못할 가능성이 있습니다.(FD 누수)
        - OS는 이러한 파일 디스크립터를 "사용 중"으로 간주하고 리소스를 계속 점유합니다.
        - 사용 가능한 파일 디스크립터의 수가 더 빨리 줄어듭니다.
        - 최종적으로, OS 커널 리소스가 모두 소진되어:
            - 새로운 파일이나 소켓을 열 수 없게 됩니다.
            - 애플리케이션에서 "Too many open files" 오류 발생합니다.
    - 이로 인해 파일 디스크립터 한도에 더 빨리 도달하며, 서비스가 중단될 가능성이 커집니다.

- 애플리케이션의 동작 문제
    - 오류 발생 및 중단
        - 파일 디스크립터 제한 초과로 인해 다음과 같은 시스템 호출이 실패:
            - `accept()`:
                - 요청이 들어올 때마다 애플리케이션은 `accept()` 호출로 소켓을 생성하려고 합니다.
                - OS가 허용할 수 있는 파일 디스크립터의 수를 초과하면 `accept()` 호출이 실패하고, 애플리케이션은 이 예외를 처리하지 못해 비정상 종료될 수 있습니다.

                    ```plaintext
                    accept: Too many open files
                    ```

            - `read()`, `write()`:
                - 기존 연결에서의 데이터 송수신 실패.
                - 열려 있는 소켓 중 일부가 비정상적으로 닫히거나, 파일 디스크립터가 고갈되면 데이터 송수신이 중단됩니다.
                - OS는 EIO(Input/Output error) 또는 ENOMEM(Out of Memory) 오류를 반환합니다.
            - `open()`: 로그 파일, 리소스 파일 등 추가 파일 열기 실패.
        - 결과적으로 애플리케이션은 예외를 던지거나 중단될 수 있습니다.

    - 서비스 장애
        - 요청 처리 중단은 전체 서비스에 영향을 미칩니다. 중요한 요청도 처리되지 않으며, 클라이언트는 서비스 불안정으로 인식하게 됩니다.

- TCP 포트 고갈
    - TCP 연결은 `(소스 IP:포트, 대상 IP:포트)`로 식별됩니다.
    - TCP 연결이 종료된 후에도, OS는 해당 포트를 `TIME_WAIT` 상태로 유지합니다(TCP 연결 재사용을 위해).
    - 동시 연결 요청이 급증하면 `TIME_WAIT` 포트가 가득 차 포트가 고갈되고, 새 연결에 사용할 포트가 없어집니다.

- OS 커널 로그에서의 오류 메시지

    파일 디스크립터 제한 초과나 네트워크 자원 부족 상황이 발생하면, 커널 로그(`/var/log/syslog` 또는 `dmesg`)에 다음과 같은 오류가 기록될 수 있습니다:
    - 파일 디스크립터 초과:

        ```plaintext
        Too many open files
        ```

    - 메모리 부족

        ```plaintext
        Cannot allocate memory
        ```

    - 네트워크 큐 초과

        ```plaintext
        TCP: drop open requests due to syn backlog full
        ```

    - 포트 고갈:

        ```plaintext
        Out of sockets: no buffer space available
        ```

    - `Connection refused`

다음과 같은 방법으로 문제를 해결할 수 있습니다:

1. 파일 디스크립터 제한 증가

    - `/etc/security/limits.conf` 또는 `ulimit`로 파일 디스크립터 한도를 늘립니다:

        ```plaintext
        *   hard    nofile   65535
        *   soft    nofile   65535
        ```

    - `sysctl`로 시스템 전체 파일 디스크립터 제한을 증가시킵니다:

        ```bash
        sysctl -w fs.file-max=2097152
        ```

2. 네트워크 큐 크기 조정

    - `backlog` 크기를 늘려 네트워크 대기열의 용량을 증가시킵니다:

        ```bash
        sysctl -w net.core.somaxconn=1024
        ```

3. 비동기 및 논블로킹 I/O 사용

    - 동기 방식 대신 비동기 I/O 모델(예: Netty, Kotlin Coroutine)을 사용하여 스레드 소비를 줄입니다.

4. 애플리케이션 로직 최적화

    - 불필요한 파일 디스크립터 사용을 줄이고, 요청 처리 시 자원 해제를 철저히 관리합니다.

## `java -jar SpringBootApp.jar` 명령어의 실행 과정

### 바이너리를 프로세스로 만드는 과정

사용자가 쉘에서 명령어(`java -jar`)를 입력하거나 프로그램이 `fork()` 또는 `exec()` 시스템 콜을 호출하면, 운영 체제에 새로운 프로세스를 생성하라는 요청이 전달됩니다.

`fork()` 시스템 콜은 부모 프로세스를 복제하여 자식 프로세스를 생성하는 기능을 합니다.

- PCB 생성

    부모 프로세스의 PCB를 기반으로 자식 프로세스의 PCB가 새롭게 만들어지고, 일부 정보가 복사됩니다.
    자식 프로세스는 고유한 PID를 가지며, 이를 포함한 정보를 새 PCB에 저장합니다

    부모 프로세스의 PCB에서 복사되는 주요 정보:
    - 프로세스 상태: 새 PCB는 Ready 상태로 초기화됩니다.
    - 프로세스 계층 구조: 부모-자식 관계가 기록되며, 부모 PID(PPID)가 저장됩니다.
    - 파일 디스크립터:
        - 부모 프로세스의 열린 파일 디스크립터가 복사됩니다.
        - 파일 디스크립터 테이블은 공유되지 않고 별도로 복사되어, 자식은 독립적으로 파일에 접근합니다.
    - 스케줄링 정보: 스케줄링 우선순위와 같은 정보가 복사됩니다.
    - 환경 변수 포인터: 환경 변수는 복사됩니다. 자식이 이를 수정하더라도 부모에게 영향을 미치지 않습니다.

    PCB에서 새롭게 할당되는 정보:
    - 새로운 PID: 자식 프로세스는 고유의 PID를 부여받습니다.
    - 프로세스 통계: CPU 사용 시간, 페이지 결함(page fault) 등은 초기화됩니다.

- 프로세스 메모리 공간 복제

    Copy-On-Write(CoW) 방식으로 복제가 이뤄집니다.
    부모 프로세스의 메모리 공간은 초기에는 복사되지 않습니다.

    자식 프로세스는 부모와 동일한 메모리 페이지를 공유하지만,
    어느 한쪽이 해당 페이지에 쓰기 작업을 시도하면, 운영 체제는 해당 페이지를 복사합니다.

    이를 통해 메모리 사용량을 줄이고, 프로세스 생성 속도를 크게 향상시킬 수 있습니다.

    복제되는 메모리 영역:
    - 코드 영역 (Code Segment): 부모와 자식이 읽기 전용으로 공유합니다. 일반적으로 변경되지 않으므로, 복사가 필요 없습니다.
    - 데이터 영역 (Data Segment): 초기에는 부모와 공유하지만, 쓰기 작업 시 복사됩니다.
    - 힙 영역 (Heap):
        - 부모와 공유하며, 쓰기 작업 시 복사됩니다.
        - 동적 메모리(malloc 등)가 여기에 포함됩니다.
    - 스택 영역 (Stack): 쓰기 작업이 많으므로, Copy-On-Write 시 가장 먼저 복사되는 경우가 많습니다.

    페이지 테이블 복사:
    - 부모 프로세스의 페이지 테이블은 자식 프로세스에 복사됩니다.
    - 페이지 테이블은 가상 주소에서 물리 주소로의 매핑 정보를 저장하며, 초기에 부모와 동일한 페이지를 참조하도록 설정됩니다.

- 파일 디스크립터 복제

    파일 디스크립터 테이블은 복제되며, 자식은 부모와 독립적으로 파일 작업을 수행할 수 있습니다.

    - 부모 프로세스에서 열려 있는 모든 파일 디스크립터가 자식 프로세스로 복사됩니다.
    - 파일 디스크립터 테이블은 복제되지만, 파일 오프셋은 공유됩니다.
        - 예를 들어, 부모가 파일을 읽거나 쓰면 자식이 동일한 오프셋에서 작업을 이어갑니다.
        - 필요시 `lseek` 등을 통해 독립적으로 설정 가능합니다.

새롭게 생성된 프로세스는 자식 프로세스가 되며, 부모와 거의 동일한 상태로 시작합니다.

그리고 복제된 프로세스는 고유한 프로세스 ID(PID)를 부여받습니다.
운영 체제는 고유한 PID를 관리하기 위한 PID 테이블을 유지합니다.
- 테이블은 사용 가능한 PID와 현재 사용 중인 PID를 추적합니다.
- PID는 일반적으로 1부터 시작하며, 사용 가능한 최대값은 시스템 설정에 따라 다릅니다(예: 32767 또는 4194304).
새로운 프로세스가 생성되면, PID 테이블에서 사용 가능한 최소값을 선택하여 새 프로세스에 할당합니다.
프로세스 종료 시 해당 PID는 다시 사용 가능 상태로 돌아갑니다.

이 시점에서 부모와 자식 프로세스는 동일한 프로그램을 실행하고 있습니다.

자식 프로세스는 `exec()` 시스템 콜을 호출하여 특정 프로그램 바이너리를 로드합니다.
이 과정에서 자식 프로세스의 메모리 공간은 새롭게 로드된 프로그램으로 덮어씌워집니다.
- 실행 파일의 코드 세그먼트 (명령어 집합).
- 초기 데이터 세그먼트 (전역 변수, 초기화된 데이터 등).
- 힙과 스택 영역 초기화.

이 과정에서 기존 메모리는 삭제되며, 로드된 프로그램이 새로운 프로세스의 실행 주체가 됩니다.

프로세스 생성 중에는 메모리가 할당됩니다.
운영 체제는 프로세스에 대해 가상 메모리 공간을 설정합니다.
가상 메모리는 페이징(Paging) 또는 세그먼테이션(Segmentation) 기법을 사용하여 실제 물리 메모리를 매핑합니다.
초기 설정은 다음과 같이 이뤄집니다:
- 코드 세그먼트: 실행 바이너리의 명령어가 로드됩니다.
- 데이터 세그먼트: 초기화된 전역 변수 및 정적 변수가 여기에 저장됩니다.
- 힙 영역: 런타임 동적 메모리 할당(`malloc`, `new`)에 사용됩니다.
- 스택 영역: 함수 호출 및 지역 변수 저장에 사용됩니다.

프로세스의 가상 주소를 물리적 메모리로 매핑하기 위한 페이지 테이블이 생성됩니다.
이 테이블은 프로세스 실행 중 메모리 접근 요청을 관리합니다.

부모 프로세스의 파일 디스크립터가 복제되며, 자식 프로세스는 동일한 파일에 접근할 수 있습니다.

프로세스가 생성되면, 운영 체제는 프로세스를 준비 상태(Ready State)로 설정합니다.
준비 상태에서 CPU 스케줄러가 프로세스를 선택하여 실행하게 됩니다.

CPU 스케줄러가 프로세스에 CPU를 할당하면, 프로세스는 실행 상태(Running State)로 전환됩니다.

프로세스가 I/O 작업을 기다리는 동안 대기 상태(Waiting State)로 전환됩니다.

마지막으로 프로세스가 작업을 완료하거나 오류로 인해 종료되면, 운영 체제는 해당 프로세스를 종료 상태로 설정하고 PID를 해제합니다.

### Java Virtual Machine (JVM) 시작

사용자가 `java -jar` 명령을 실행하면 운영 체제의 쉘은 `java` 바이너리 프로그램을 실행합니다.
`java` 바이너리는 Java Virtual Machine의 실행 환경을 초기화하며, 이 바이너리는 네이티브 코드로 작성된 프로그램으로, JVM 실행의 시작점입니다.

운영 체제는 JVM을 새로운 프로세스로 실행합니다.
새로운 프로세스에는 PID(Process ID)가 부여되며, 운영 체제가 제공하는 기본적인 메모리 공간(코드, 데이터, 힙, 스택 영역)을 할당받습니다.

JVM은 바이트코드로 컴파일된 Java 애플리케이션을 실행하기 위한 런타임 환경을 제공합니다.
`-jar` 옵션은 실행할 JAR 파일을 명시하며, 해당 JAR 파일의 엔트리 포인트(Entry Point)를 자동으로 찾습니다.

운영 체제에서 JVM 프로세스에 할당된 메모리 공간은 다음과 같이 구성됩니다:

- Code Segment: 네이티브 JVM 실행 코드와 관련된 명령어가 저장됩니다.
- Data Segment: 정적 데이터와 초기화된 전역 변수가 저장됩니다.
- Heap: JVM의 Heap 영역에 사용됩니다.
- Stack: 각 스레드의 Java 스택이 포함됩니다.
- OS Kernel 영역: 운영 체제가 관리하는 메모리 영역으로, 파일 I/O 및 네트워크 작업에 사용됩니다.

### JVM 내부 메모리 초기화

JVM은 실행될 프로그램에 필요한 메모리 공간을 운영 체제로부터 요청합니다.
JVM은 이 메모리 공간을 다음과 같이 주요 영역으로 나눕니다:
- Method Area (클래스 메타데이터 저장 공간)

    모든 스레드에서 공유되는 영역으로, JVM이 클래스 로더에 의해 로드한 클래스의 메타데이터를 저장하는 공간입니다.
    메서드 코드(바이트코드), 상수 풀(Constant Pool), 정적 변수, 클래스 레벨 데이터가 여기에 저장됩니다.

    Java 8부터는 HotSpot JVM에서 이 영역이 메타스페이스(Metaspace)로 대체되었습니다.
    Java 8 이전에는 `Permanent Generation`이라는 영역에서 클래스 메타데이터를 관리했으나, 이는 JVM 힙 공간의 일부로 관리되었습니다.
    Java 8부터는 `Metaspace`가 도입되었으며, 이는 네이티브 메모리를 활용합니다.
    - 동적으로 크기를 확장할 수 있어 메모리 부족 문제를 해결합니다.
    - `-XX:MetaspaceSize`: 초기 크기 설정.
    - `-XX:MaxMetaspaceSize`: 최대 크기 제한.

    저장되는 데이터:
    1. 클래스 정보:
        - 클래스 이름, 접근 제한자(`public`/`private`), 부모 클래스 정보, 구현된 인터페이스 목록.
        - 클래스와 관련된 메서드, 필드, 정적 변수의 구조 정보.

    2. 상수 풀(Constant Pool):
        - 클래스 파일의 심볼릭 참조와 리터럴 데이터(문자열 상수, 숫자 리터럴 등)가 저장됩니다.
        - 상수 풀은 런타임 중에 사용되며, JVM이 바이트코드를 실행하면서 참조를 실제 메모리 주소로 변환하는 데 활용됩니다.

    3. 정적 변수:
        - `static` 키워드로 선언된 클래스 변수는 Method Area에 저장되며, 애플리케이션 종료 시까지 유지됩니다.

        ```java
        class Example {
            static int sharedCounter = 0;
        }
        ```

    4. 메서드 코드:
        - 메서드 바이트코드가 포함됩니다.
        - 각 메서드의 시그니처, 반환 타입, 로컬 변수와 같은 실행에 필요한 정보도 여기에 포함됩니다.

- Heap (객체 저장 공간)

    Heap 메모리는 JVM에서 가장 중요한 영역 중 하나로, 런타임 동안 생성되는 모든 객체와 배열이 저장되는 공간입니다.
    이 영역은 Garbage Collector(GC)에 의해 관리되며, Java의 메모리 누수를 방지하고 JVM에 의해 크기가 동적으로 조정될 수 있습니다.

    JVM 실행 시 초기 Heap 크기(`-Xms`)와 최대 Heap 크기(`-Xmx`)가 설정됩니다.
    기본적으로 JVM은 시스템 메모리의 일부를 사용하며, 개발자가 옵션을 통해 조정할 수 있습니다.
    Heap 메모리는 필요에 따라 확장되며, 최대 크기(`-Xmx`)를 초과할 경우 `OutOfMemoryError`가 발생합니다.

    객체 생성 과정은 다음과 같습니다.
    - `new` 키워드로 객체가 생성되면 JVM은 Heap에서 메모리를 할당합니다.

        ```java
        // `obj` 참조는 스택에, 객체 데이터는 Heap에 저장됩니다.
        MyClass obj = new MyClass();

        // 런타임 중 크기가 가변적인 데이터 구조(예: `ArrayList`, `HashMap`)가 Heap에 저장됩니다.
        List<String> data = new ArrayList<>();
        data.add("item1");
        ```

    - 생성된 객체의 참조는 스택(Frame Data)에 저장됩니다.

    Generation 모델:
    - Young Generation:
        - 새로운 객체가 생성되고, 짧은 생명 주기의 객체를 관리합니다.
        - 대부분의 객체는 단명하며, Young Generation에서 생성되었다가 곧 GC에 의해 제거됩니다.

        1. Eden Space:
            - 객체가 처음 생성되는 영역.
            - 생성된 객체 대부분은 이곳에서 사라지며, 일정 기준을 충족한 객체만 Survivor Space로 이동합니다.
        2. Survivor Spaces (S0, S1):
            - Eden Space에서 살아남은 객체가 이동하는 영역.
            - Young Generation의 Copying GC에서 한 Survivor Space에서 다른 공간으로 객체를 이동합니다.

    - Old Generation:
        - Young Generation에서 오래 생존한 객체가 이동하며, 장수 객체를 관리합니다.
        - 예를 들어, 대규모 데이터 구조 또는 애플리케이션 전반에 걸쳐 유지되는 객체가 여기에 저장됩니다.
        - Garbage Collection:
            - Old Generation의 객체는 Full GC에서 관리됩니다.
            - Minor GC에 비해 더 많은 리소스를 소비합니다.

    - Metaspace:
        - 클래스 메타데이터가 저장되는 공간으로, Java 8부터 Heap 메모리에서 분리되어 네이티브 메모리를 사용합니다.

- Stack (스레드별 스택)

    Java Stack은 JVM에서 각 스레드마다 독립적으로 생성되는 메모리 공간으로, 메서드 호출과 반환, 지역 변수, 연산 중간 값 등을 관리합니다.
    메모리 관리는 LIFO(Last-In-First-Out) 방식으로 이루어지며, 메서드 호출 시 새로운 스택 프레임(Stack Frame)이 추가되고, 메서드가 반환되면 해당 프레임이 제거됩니다.

    JVM은 스레드가 생성될 때 고유한 스택을 할당합니다.
    스택 크기는 JVM 옵션(`-Xss`)으로 설정할 수 있습니다.

    다음과 같은 역할들을 수행합니다:
    - 메서드 호출 관리(Frame Data):
        - 메서드 호출과 관련된 모든 데이터를 저장합니다.
        - 각 스레드가 고유의 스택을 가지므로, 스레드 간 데이터 충돌이 없습니다.
        - 예외 처리를 위한 데이터도 여기에 포함됩니다.

    - 지역 변수 및 매개변수 저장:
        - 메서드의 지역 변수와 매개변수가 저장됩니다.

        ```java
        // `a`와 `b`는 Local Variables에 저장됩니다.
        int sum(int a, int b) {
            return a + b;
        }
        ```

    - 연산 스택(Operand Stack) 관리:
        - 바이트코드 실행 중 발생하는 연산 중간 값이 저장됩니다.
        - 예: `a + b` 연산에서 `a`와 `b`가 Operand Stack에 쌓이고, 덧셈 연산 후 결과가 다시 스택에 저장됩니다.

- Program Counter Register (PC Register)

    PC Register는 JVM 명령어 해석기(Interpreter)가 다음에 실행할 명령어를 추적하는 역할을 합니다.
    모든 스레드는 독립적인 PC Register를 가지며, 스레드 간 충돌을 방지합니다.

    스레드 생성 시 PC Register가 초기화됩니다.
    메서드 호출 시 해당 메서드의 첫 번째 명령어 주소가 PC Register에 저장됩니다.

    명령어 해석기가 PC Register를 읽어 현재 실행 중인 명령을 확인하고, 다음 명령을 실행합니다.
    네이티브 메서드 실행 중일 때는 PC Register 값이 정의되지 않을 수 있습니다.

    다음과 같은 역할들을 수행합니다.
    1. 명령어 위치 추적:
        - 현재 실행 중인 JVM 명령어(바이트코드)의 주소를 저장합니다.
    2. 스레드 안전성 제공:
        - 각 스레드가 독립적인 PC Register를 가지므로, 동시 실행 환경에서도 안전하게 명령어를 추적할 수 있습니다.

    JVM의 명령어 해석기(Interpreter)가 다음에 실행할 명령을 결정하는 데 사용됩니다.

- Native Method Stack (네이티브 메서드 호출용 스택)

    Native Method Stack은 Java 애플리케이션이 네이티브 코드(C/C++)를 실행할 때 사용하는 메모리 영역입니다.
    JVM은 Java Native Interface(JNI)를 통해 외부 라이브러리와 상호작용하며, 이 과정에서 네이티브 메서드 호출 정보를 관리합니다.

    네이티브 메서드 호출 시 Native Method Stack이 초기화됩니다.
    호출이 완료되면 스택은 정리됩니다.

    예를 들어, `System.loadLibrary()`로 외부 네이티브 라이브러리를 로드하거나 파일 I/O, 네트워크 소켓 접근 등의 경우에 사용됩니다.

    다음과 같은 역할들을 수행합니다.
    1. 네이티브 라이브러리 호출:
        - 네이티브 코드에서 정의된 함수 호출을 처리합니다.
    2. 네이티브 데이터 관리:
        - 네이티브 함수 호출과 관련된 데이터(매개변수, 반환 값 등)를 저장합니다.

    네이티브 메서드를 호출하기 위한 스택으로, 네이티브 라이브러리 함수와의 상호작용을 처리합니다.

JVM의 메모리 할당 및 관리는 다음과 같이 이뤄집니다.
- 클래스 로드 단계
    1. Loading: 클래스 로더가 `.class` 파일을 읽어 JVM에 로드.
        - 클래스 로더(ClassLoader)가 `.class` 파일을 읽고 메타데이터를 Method Area에 적재합니다.
        - 이는 애플리케이션이 처음으로 해당 클래스를 참조할 때 발생합니다.

    2. Linking:
        - Verification: JVM은 바이트코드의 유효성을 확인하여, 잘못된 바이트코드로 인해 런타임 에러가 발생하는 것을 방지합니다.

        - Preparation:
            - 클래스의 정적 변수와 기본값이 설정됩니다.
            - 예: `static int x;`는 0으로 초기화됩니다.

        - Resolution:
            - 심볼릭 참조를 실제 메모리 주소로 변환합니다.
            - 클래스 내부의 메서드 참조, 필드 참조 등이 이 단계에서 해결됩니다.

    3. Initialization: `static` 초기화 블록과 정적 변수가 실행됩니다.

        ```java
        static int x = 10;
        static {
            x = 20;
        }
        ```

- 객체 생성과 Heap 관리
    - `new` 키워드를 사용하면 JVM은 Heap에 객체를 생성하고, 해당 객체의 참조를 반환합니다.
    - GC가 주기적으로 Heap을 스캔하여 더 이상 참조되지 않는 객체를 회수합니다.

- 스레드 스택 관리
    - 각 스레드는 고유의 스택을 가지고 있으며, 메서드 호출 시마다 스택 프레임을 추가합니다.
    - 메서드 종료 시 해당 스택 프레임은 제거됩니다.

- Garbage Collection

    Garbage Collection은 더 이상 참조되지 않는 객체를 자동으로 회수하여 메모리 누수를 방지합니다.

    GC 알고리즘은 다음과 같습니다:
    - Mark-and-Sweep:
        - 사용 중인 객체를 표시하고, 사용되지 않는 객체를 삭제.
        - Mark 단계: 모든 객체를 스캔하여 참조 중인 객체를 마킹.
        - Sweep 단계: 참조되지 않는 객체를 제거하고, 빈 메모리를 회수.
    - Copying:
        - Young Generation에서 사용.
        - 사용 중인 객체를 복사하여 새로운 공간에 할당.
        - 살아남은 객체를 Eden Space에서 Survivor Space로 복사하고, Eden Space를 비웁니다.
    - Generational GC:
        - Heap을 Young Generation과 Old Generation을 나눠 효율적으로 관리.
        - Young Generation에서 Minor GC, Old Generation에서 Major GC 또는 Full GC를 수행.

    Young과 Old 영역별로 서로 다른 GC가 수행됩니다.
    1. Minor GC:
        - Young Generation에서 발생하며, Eden Space와 Survivor Space의 객체를 정리합니다.
        - 대부분의 객체는 Eden Space에서 제거되며, 살아남은 객체는 Survivor Space로 이동합니다.

    2. Major GC (Full GC):
        - Old Generation과 Young Generation을 모두 정리합니다.
        - 더 많은 시간이 소요되며, 애플리케이션 성능에 영향을 미칠 수 있습니다.

    JVM은 다양한 GC 알고리즘을 제공하며, 애플리케이션의 요구에 따라 선택할 수 있습니다:

    - Serial GC:
        - 단일 스레드 환경에 적합.
        - `-XX:+UseSerialGC` 옵션으로 활성화.
    - Parallel GC:
        - 다중 스레드 환경에 적합.
        - `-XX:+UseParallelGC` 옵션으로 활성화.
    - G1 GC (Garbage First):
        - 대규모 애플리케이션에 적합하며, Full GC 발생 빈도를 줄임.
        - `-XX:+UseG1GC` 옵션으로 활성화.

### JAR 파일 로드

JVM은 지정된 JAR 파일을 로드합니다.
이 과정에서 JAR 파일 내부의 `META-INF/MANIFEST.MF` 파일을 읽습니다.
이 파일에는 `Main-Class`라는 속성이 있는데, 이는 JVM이 실행해야 할 클래스의 FQCN(Fully Qualified Class Name)을 지정합니다.

```plaintext
Main-Class: com.example.SpringBootApp
```

메인 클래스를 확인하고, JVM 내부에서 `main()` 메서드의 시작점을 결정합니다.

Spring Boot는 이 속성을 사용하여 메인 애플리케이션 클래스를 지정합니다.

### 클래스 로더에 의해 클래스 로드

JVM은 클래스 로더(Class Loader)를 사용해 애플리케이션이 실행되는 데 필요한 클래스와 의존성을 로드합니다.

Spring Boot JAR 파일은 Fat JAR로, 모든 종속 라이브러리가 포함되어 있습니다.
이를 통해 실행 시 추가적인 라이브러리 설치 없이 독립적으로 실행 가능합니다.

Spring Boot의 `org.springframework.boot.loader.JarLauncher`가 이를 처리하며, Fat JAR의 구조를 분석하고 적절히 클래스를 로드합니다.

### `main()` 메서드 실행

Manifest 파일에 명시된 메인 클래스의 `public static void main(String[] args)` 메서드가 호출됩니다.

Spring Boot 애플리케이션의 경우, 이 메서드는 일반적으로 `SpringApplication.run()`을 호출하여 애플리케이션 컨텍스트를 초기화합니다.

```java
public class SpringBootApp {
    public static void main(String[] args) {
        SpringApplication.run(SpringBootApp.class, args);
    }
}
```

### Spring Boot의 초기화 단계

Spring Boot 애플리케이션의 초기화 단계는 다음과 같습니다.

#### (a) 스프링 애플리케이션 컨텍스트 초기화

- SpringApplication이 실행되면서 애플리케이션 컨텍스트(Application Context)가 생성됩니다.
- 설정된 `@Configuration` 클래스와 `@ComponentScan`을 통해 애플리케이션의 구성 요소가 로드됩니다.
- Bean Factory가 필요한 Bean들을 생성 및 초기화합니다.

#### (b) 내장 웹 서버(Tomcat, Jetty 등) 시작

- Spring Boot는 내장 웹 서버를 포함하고 있으므로, 애플리케이션 컨텍스트를 초기화한 후 내장 서버를 실행합니다.
- 기본적으로 Spring Boot는 Tomcat 서버를 사용하며, 포트 8080에서 수신 대기합니다(설정에 따라 다를 수 있음).

#### (c) DispatcherServlet 등록

- Spring MVC 애플리케이션인 경우, `DispatcherServlet`이 서블릿 컨테이너에 등록됩니다.
- 요청 매핑, 핸들러 설정 등이 완료됩니다.

#### (d) 애플리케이션 로직 실행 준비 완료

- 애플리케이션의 주요 로직(예: REST API, 서비스)이 요청을 처리할 준비가 됩니다.

### 6. 운영 상태로 전환

- Spring Boot 애플리케이션이 정상적으로 실행되었다면, 콘솔에 다음과 비슷한 로그가 표시됩니다.

     ```plaintext
     Started SpringBootApp in 3.456 seconds (JVM running for 3.890)
     ```

- 내장 서버가 시작된 상태에서 HTTP 요청을 수신하거나, 비동기 작업을 처리할 준비가 완료됩니다.

### 추가 사항

1. 환경 변수 및 프로파일
   - 실행 시 `application.properties` 또는 `application.yml` 파일을 로드하며, 환경 변수와 프로파일 설정을 반영합니다.
   - `-Dspring.profiles.active=dev`와 같은 옵션을 사용해 프로파일을 지정할 수도 있습니다.

2. 종료 처리
   - 애플리케이션이 종료될 때 JVM은 모든 리소스를 해제하며, Spring Boot는 등록된 `@PreDestroy` 메서드를 호출하여 필요한 종료 작업을 수행합니다.

## 수억건의 데이터가 존재하는 테이블 처리 방법

### 문제 상황

- MySQL 5.5버전으로 3억건의 데이터가 저장된 테이블이 존재합니다.
- 하나의 테이블이 너무 커져서 몇 백 기가에 달하는 데이터가 존재합니다.
- 결제 내역을 저장하는 테이블이어서 서비스 중간은 불가능한 상태입니다.
- 또한 DB 버전 업그레이드도 현재로서는 어려운 상황입니다.
- 별도의 reader 서버나 replica 서버 등도 있지만, 결국 하나의 큰 테이블을 어떻게 해야 하는 상황입니다.
- incremental integer id가 있어서 이 아이디가 조회에 유의미하게 사용되고 있습니다. 따라서 데이터가 쌓이는 순서도 중요합니다.
- 조회도 실시간으로 이뤄지고 있습니다. 따라서 파티셔닝을 해도 조회는 일관되게 제공해야 합니다.

이런 상황에서 현재 테이블을 어떻게 해야 할지 한번 논해보세요.

### 핵심 제약

1. MySQL 5.5의 한계:
    - 최신 MySQL 버전에서 제공하는 `InnoDB`의 native partitioning 및 온라인 작업 기능을 사용할 수 없습니다.
2. 실시간 데이터 일관성:
    - 조회와 삽입 작업 모두 서비스 중단 없이 이루어져야 하며, 모든 데이터가 일관된 인터페이스로 접근 가능해야 합니다.
3. 데이터 크기와 성능 문제:
    - 테이블 크기가 크기 때문에 스캔 속도, 인덱스 유지 비용, 쿼리 응답 시간이 점점 느려질 가능성이 높습니다.

### 테이블 파티셔닝 (수동 구현)

MySQL 5.5에서 제공되는 기본적인 파티셔닝 기능은 제한적입니다.
대신, 수동으로 테이블을 분리하고 애플리케이션 레벨에서 이를 관리하는 방식을 사용할 수 있습니다.

#### 수동 파티셔닝 설계

- 기존 테이블(`payments`)을 기간 기반 또는 범위 기반으로 나누어 별도의 테이블로 분리합니다.
- ID 범위 기준으로 파티션 분리:
    - ID는 증가형 정수이므로, 특정 범위를 기준으로 물리적 테이블을 분리할 수 있습니다.
    - 예를 들어:

        ```sql
        CREATE TABLE payments_1_to_100M LIKE payments;
        CREATE TABLE payments_100M_to_200M LIKE payments;
        CREATE TABLE payments_200M_to_300M LIKE payments;
        ```

    - 기존 데이터를 새로운 테이블로 옮깁니다:

        ```sql
        INSERT INTO payments_1_to_100M SELECT * FROM payments WHERE id BETWEEN 1 AND 100000000;
        DELETE FROM payments WHERE id BETWEEN 1 AND 100000000;
        ```

- 데이터 삽입:
    - 데이터는 새로운 `payments` 테이블에 삽입하며, 일정 기준에 도달하면 다음 파티션 테이블로 데이터를 옮깁니다.

#### 조회 시 파티션 통합

조회 작업은 뷰(View)를 사용하여 통합 인터페이스를 제공합니다:

```sql
CREATE VIEW all_payments AS
SELECT * FROM payments_1_to_100M
UNION ALL
SELECT * FROM payments_100M_to_200M
UNION ALL
SELECT * FROM payments_200M_to_300M;
```

- 실시간 쿼리에서 `all_payments`를 사용하면 기존 단일 테이블과 동일한 인터페이스를 제공합니다.
- 쿼리는 MySQL의 최적화 엔진에 따라 각 파티션 테이블에 병렬로 접근합니다.

하지만 뷰를 사용하는 경우, 뷰 자체에는 인덱스를 생성할 수 없기 때문에 기존 인덱스의 기능을 유지하기 위한 별도의 접근 방식이 필요합니다.

뷰를 사용하는 경우, 뷰 자체에는 인덱스를 생성할 수 없기 때문에 기존 인덱스의 기능을 유지하기 위한 별도의 접근 방식이 필요합니다. 각 파티션(분리된 테이블)에 존재하는 인덱스를 활용하면서, 통합 조회 성능을 유지할 수 있는 몇 가지 전략을 아래와 같이 제안합니다.

1. 개별 파티션 테이블에 동일한 인덱스 유지

    뷰가 여러 파티션 테이블을 병합(`UNION ALL`)하므로, 각 파티션 테이블에 필요한 인덱스를 동일하게 설정해야 합니다.

    파티션 테이블에 인덱스 생성:
    - 기존 단일 테이블에 존재하던 주요 인덱스를 분리된 테이블에도 동일하게 생성합니다.
    - 예를 들어, 단일 테이블의 인덱스들을 각 파티션 테이블에도 동일하게 생성합니다.:

        ```sql
        -- 기존 단일 테이블의 인덱스
        ALTER TABLE payments ADD INDEX idx_created (created);
        ALTER TABLE payments ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments ADD INDEX idx_name_created (name, created);

        -- 파티션된 테이블의 인덱스
        ALTER TABLE payments_1_to_100M ADD INDEX idx_created (created);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_created (created);
        ALTER TABLE payments_1_to_100M ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments_1_to_100M ADD INDEX idx_name_created (name, created);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_name_created (name, created);
        ```

        이렇게 하면 MySQL은 뷰를 통한 조회 시 각 파티션 테이블에 존재하는 인덱스를 활용하여 조건에 맞는 데이터를 효율적으로 검색할 수 있습니다.

2. 쿼리 최적화: 조건절로 파티션 테이블 선택

    MySQL의 쿼리 최적화 엔진은 `UNION ALL`이 포함된 뷰를 사용할 때, `WHERE` 조건을 기반으로 쿼리를 최적화하여 불필요한 테이블에 접근하지 않도록 설계되어 있습니다.

    - 특정 조건이 있는 쿼리

        뷰에서 특정 조건(예: `id`, `created`, `uuid`)이 있는 경우, MySQL은 해당 조건이 충족되는 테이블만 스캔합니다.
        예를 들어, 다음과 같은 쿼리를 실행한다고 가정합니다:

        ```sql
        SELECT * FROM all_payments WHERE created >= '2023-01-01' AND created < '2023-02-01';
        ```

        MySQL은 `payments_1_to_100M`과 `payments_100M_to_200M`의 `created` 인덱스를 활용하여 쿼리를 최적화할 수 있습니다.

    - MySQL 실행 계획 확인

        뷰가 제대로 작동하는지 확인하려면 `EXPLAIN`을 사용하여 실행 계획을 점검합니다:

        ```sql
        EXPLAIN SELECT * FROM all_payments WHERE created >= '2023-01-01' AND created < '2023-02-01';
        ```

        결과에서 MySQL이 조건에 맞는 테이블만 스캔하고, 해당 테이블의 인덱스를 활용하는지 확인할 수 있습니다.

3. 필요 시 애플리케이션 레벨에서 파티션 관리

    뷰가 모든 테이블을 병합하는 방식은 간단하지만, 데이터 규모가 매우 클 경우 비효율적일 수 있습니다.
    애플리케이션 로직에서 조건에 따라 적절한 파티션 테이블에 직접 접근하도록 설계하면 성능을 극대화할 수 있습니다.

    - 애플리케이션 레벨에서 파티션 테이블을 직접 선택하면 불필요한 테이블 스캔을 완전히 제거할 수 있습니다.
    - 조건이 명확한 경우 뷰를 사용하지 않아도 되므로 성능이 크게 개선됩니다.

    예를 들어, ID나 날짜를 기준으로 각 파티션 테이블을 선택하는 애플리케이션 레벨 로직을 추가할 수 있습니다:

    ```python
    def get_partition_table(created_date):
        if created_date < '2023-01-01':
            return 'payments_1_to_100M'
        elif created_date < '2024-01-01':
            return 'payments_100M_to_200M'
        else:
            return 'payments_200M_to_300M'

    query = f"SELECT * FROM {get_partition_table(requested_date)} WHERE created = '{requested_date}'"
    ```

4. 인덱스 중복 제거 및 통합 테이블 조회 성능 최적화

    뷰와 파티션 테이블 접근을 병행하는 경우, 다음과 같은 최적화 전략을 추가로 사용할 수 있습니다.

    - 중복 인덱스 최소화
        - 뷰를 사용하면 각 테이블에 동일한 인덱스가 있어야 하므로, 필요한 최소 인덱스만 유지합니다.
        - 조회 패턴 분석을 통해 자주 사용되지 않는 인덱스를 제거합니다.

    - 데이터 분할 기준 검토
        - ID 기반 분할 외에, 데이터 접근 패턴을 분석하여 `created` 날짜 또는 `uuid`와 같은 다른 기준으로 분할을 고려할 수 있습니다.
        - 예: 날짜 기반 테이블 분리:

            ```sql
            CREATE TABLE payments_2023 LIKE payments;
            CREATE TABLE payments_2024 LIKE payments;
            ```

5. 대규모 뷰 최적화가 어려운 경우

    뷰를 통한 통합 조회가 특정 쿼리에서 비효율적이라면, 머티리얼라이즈드 뷰를 사용하거나 주기적으로 통합 테이블을 생성하는 방식을 도입할 수 있습니다.

    > 머티리얼라이즈드 뷰(Materialized View)는 데이터베이스에서 뷰(View)의 일종으로, 질의 결과를 디스크에 저장하여 실제 데이터처럼 사용할 수 있는 데이터베이스 객체입니다.
    >
    > ```sql
    > CREATE MATERIALIZED VIEW mv_sales_summary AS
    > SELECT region, SUM(sales) AS total_sales
    > FROM sales
    > GROUP BY region;
    >
    > -- 머티리얼라이즈드 뷰를 최신 상태로 유지하려면 새로 고침을 수행해야 합니다:
    > -- Oracle 예제: Complete Refresh
    > BEGIN
    > DBMS_MVIEW.REFRESH('MV_SALES_SUMMARY', 'COMPLETE');
    > END;
    >
    > -- PostgreSQL 예제: Refresh Materialized View
    > REFRESH MATERIALIZED VIEW mv_sales_summary;
    > ```
    >
    > 일반적인 뷰(View)는 질의 실행 시점에 데이터를 실시간으로 조회하는 반면, 머티리얼라이즈드 뷰는 질의 결과를 물리적으로 저장하여 나중에 필요할 때 빠르게 조회할 수 있습니다.

    - 통합 테이블 생성
        - 일정 주기로 각 파티션 테이블의 데이터를 하나의 통합 테이블로 병합합니다.
        - 통합 테이블은 읽기 전용으로 사용하며, 정기적으로 업데이트합니다.

            ```sql
            CREATE TABLE consolidated_payments LIKE payments;
            INSERT INTO consolidated_payments
            SELECT * FROM payments_1_to_100M
            UNION ALL
            SELECT * FROM payments_100M_to_200M
            UNION ALL
            SELECT * FROM payments_200M_to_300M;
            ```

    - 주기적인 업데이트
        - Cron 스케줄러나 배치 작업을 통해 통합 테이블을 갱신합니다.
        - 장점: 복잡한 조회 쿼리를 단순화할 수 있음.
        - 단점: 통합 데이터가 실시간으로 업데이트되지 않을 수 있음.

### 데이터 아카이빙

과거 데이터를 아카이브 테이블로 이동하여 현재 테이블 크기를 줄입니다.
이 방법은 오래된 결제 데이터를 조회할 필요가 적은 경우에 적합합니다.

#### 아카이브 테이블 설계

- 과거 데이터를 별도의 테이블로 이동:

    ```sql
    CREATE TABLE payments_archive LIKE payments;
    INSERT INTO payments_archive SELECT * FROM payments WHERE created_at < '2022-01-01';
    DELETE FROM payments WHERE created_at < '2022-01-01';
    ```

- 정기적인 스케줄링:
    - 데이터를 특정 기간 동안 유지하고, 그 이후에 아카이빙 작업을 수행합니다.
    - 스케줄링 도구(Cron + Shell Script 또는 MySQL Event)를 사용하여 배치 작업으로 구현.

#### 아카이브 데이터 통합 조회

- 아카이브 테이블을 포함하여 조회할 수 있도록 뷰를 제공합니다:

    ```sql
    CREATE VIEW all_payments AS
    SELECT * FROM payments
    UNION ALL
    SELECT * FROM payments_archive;
    ```

- 최적화: 과거 데이터에 접근이 적다면, 아카이브 테이블을 읽기 전용 서버(레플리카)로 옮겨 읽기 부하를 분산시킵니다.

### 인덱스 최적화 및 테이블 정리

대규모 테이블에서는 인덱스 최적화가 성능에 큰 영향을 미칩니다.

#### 3.1 중복 인덱스 제거

- `SHOW INDEX FROM payments`로 중복 인덱스를 확인합니다.
- 불필요한 인덱스를 삭제하여 인덱스 유지 비용을 줄입니다:

    ```sql
    DROP INDEX idx_column_name ON payments;
    ```

### 3.2 복합 인덱스 추가

- 조회 패턴을 분석하여 복합 인덱스를 추가합니다.
- 예: `id`와 `created_at`을 함께 사용하는 쿼리 최적화를 위해 복합 인덱스를 생성:

    ```sql
    ALTER TABLE payments ADD INDEX idx_id_created_at (id, created_at);
    ```

### 점진적 테이블 분리

대규모 테이블을 분리하기 위해 서비스 중단 없이 데이터 마이그레이션을 수행합니다.

#### 배치로 데이터 옮기기

- 데이터를 분리하는 동안 트랜잭션을 사용하여 무결성을 유지:

    ```sql
    INSERT INTO payments_1_to_100M SELECT * FROM payments WHERE id BETWEEN 1 AND 100000000;
    DELETE FROM payments WHERE id BETWEEN 1 AND 100000000;
    ```

- 한 번에 모든 데이터를 옮기지 말고, 작은 청크 단위로 배치 처리합니다:

    ```sql
    SET @start = 1, @batch_size = 100000;
    WHILE (@start <= 100000000) DO
        INSERT INTO payments_1_to_100M
        SELECT * FROM payments
        WHERE id BETWEEN @start AND @start + @batch_size - 1;
        DELETE FROM payments
        WHERE id BETWEEN @start AND @start + @batch_size - 1;
        SET @start = @start + @batch_size;
    END WHILE;
    ```

#### 점진적 데이터 분리 전략

- 새로운 데이터를 추가할 때부터 파티션 테이블로 직접 저장.
- 예: ID 생성 규칙에 따라 `payments_200M_to_300M` 테이블로 삽입.

## 3억건의 데이터의 B+Tree 깊이 계산

B+ Tree에서 데이터의 깊이를 계산하려면 트리의 구조와 노드에 저장할 수 있는 데이터의 양을 기반으로 깊이를 추정할 수 있습니다.
B+ Tree는 모든 리프 노드가 동일한 깊이에 있고, 내부 노드가 많은 키를 저장할 수 있으므로 깊이가 비교적 얕은 것이 특징입니다.

### B+ Tree의 주요 특징

- 차수(Order, `m`): 각 노드는 최대 `m-1`개의 키를 저장할 수 있고, 최대 `m`개의 자식 노드를 가질 수 있습니다.
- 데이터 저장 위치: 모든 데이터는 리프 노드에 저장되며, 내부 노드는 검색을 위한 키만 저장합니다.
- 균형 트리: 트리는 항상 균형 상태를 유지하며, 모든 리프 노드는 동일한 깊이를 가집니다.

### B+ Tree의 최대 용량

- 한 레벨의 노드 수가 증가할수록 저장 가능한 키의 수는 기하급수적으로 증가합니다.
- 루트 노드는 최대 `m-1`개의 키를 저장합니다.
- 각 레벨의 노드는 최대 `m^h`개의 키를 관리합니다(`h`는 깊이).

### 인덱싱 예시

MySQL에서 일반적으로 사용되는 InnoDB의 B+ Tree는 16KB 페이지 크기를 사용하며, 페이지마다 약 200~400개의 키를 저장할 수 있습니다(각 키의 크기에 따라 다름).

### 깊이 계산 방법

1. 노드당 키 수 추정: 데이터가 3억 건이고, 각 노드(페이지)가 최대 300개의 키를 저장할 수 있다고 가정합니다.

2. 트리 깊이 공식

    - 리프 노드의 개수를 `L`이라 하고, 데이터 개수를 `N`, 노드의 키 수를 `m`이라 할 때:
        $$
        L = \lceil \frac{N}{m} \rceil
        $$
        $$
        h = \lceil \log_{m} L \rceil
        $$
    여기서:
    - $N = 300,000,000$ (데이터 수).
    - $m = 300$ (노드당 최대 키 수).

3. 구체적인 계산

    - 리프 노드의 수 $L$:
        $$
        L = \lceil \frac{300,000,000}{300} \rceil = 1,000,000 \, \text{(리프 노드 개수)}.
        $$

    - 트리 깊이 계산

        내부 노드의 자식 수(차수)를 $m = 300$으로 가정했을 때:
        $$
        h = \lceil \log_{300} 1,000,000 \rceil
        $$

        이를 계산하면:
        1. $\log_{300} 1,000,000 = \frac{\log_{10} 1,000,000}{\log_{10} 300}$
        2. $\log_{10} 1,000,000 = 6$ (10의 6제곱).
        3. $\log_{10} 300 \approx 2.477$.
        4. $\frac{6}{2.477} \approx 2.42$.

        즉, $h \approx 3$.
        트리의 깊이는 약 3~4 레벨로 유지됩니다.

        이러한 얕은 깊이 덕분에 데이터 접근 시 최대 3~4번의 I/O만으로도 원하는 데이터를 찾을 수 있습니다.

### B+ Tree 깊이와 데이터 접근 시간

- 접근 시간
    - B+ Tree는 깊이가 얕기 때문에 검색 성능이 매우 우수합니다.
    - 3억 건의 데이터에서도 3~4번의 디스크 페이지 읽기만으로 원하는 데이터를 찾을 수 있습니다.

- 다양한 차수와 키 수 변화

    차수가 클수록(페이지당 저장 가능한 키 수가 많을수록), 트리의 깊이는 더 얕아집니다:
    - $m = 400$일 경우:
        - 리프 노드 수 $L = 750,000$.
        - 깊이 $h = \lceil \log_{400} 750,000 \rceil \approx 3$.

### 추가 고려 사항

- InnoDB의 실제 구현
    - InnoDB는 리프 노드에 데이터 레코드를 직접 저장하고, 내부 노드는 검색 키만 저장합니다.
    - 내부 노드는 더 많은 키를 저장할 수 있으므로, 실제 트리의 깊이는 이론적으로 계산된 값보다 조금 더 얕을 수 있습니다.

- 다양한 데이터 분포
    - 데이터가 균등 분포되지 않고 특정 범위에 집중되어 있으면, 일부 리프 노드가 비대해질 수 있습니다.
    - 그러나 B+ Tree는 자동으로 균형을 유지하므로 깊이는 크게 변하지 않습니다.

- 데이터 추가 시 깊이 변화
    - 데이터가 지속적으로 추가되면 리프 노드와 내부 노드가 분할되며 트리의 깊이가 증가할 수 있습니다.
    - 그러나 실질적으로 데이터 수가 수십 배로 증가하지 않는 이상 깊이는 일정하게 유지됩니다.

## infix, inline, noinline, crossinline

Kotlin에서 제공하는 `infix`, `inline`, `noinline`, `crossinline` 키워드는 각각 특정한 기능과 의도를 반영하여 설계된 특수한 언어 구성 요소입니다.

### Infix

`infix`는 함수 호출을 보다 직관적이고 읽기 쉽게 하기 위해 제공되는 기능입니다.
"중간(infix)"에 위치한다는 의미에서 이름이 붙었습니다. 이는 기존의 전위(prefix), 후위(postfix) 표현과 구분됩니다.
일반적으로 객체 메서드는 `object.method(argument)` 형식으로 호출되지만, `infix`를 사용하면 중위 표현식(`object method argument`)으로 사용할 수 있습니다.
주로 DSL(Domain-Specific Language)나 간결한 표현식이 필요한 경우에 유용합니다.

사용할 수 있는 조건은 다음과 같습니다:
- 멤버 함수나 확장 함수여야 합니다.
- 함수는 하나의 매개변수만 받아야 합니다.
- `infix` 키워드로 선언되어야 합니다.

`infix`는 JVM에서 특별한 바이트코드를 생성하지 않습니다.
단순히 호출 구문을 간결하게 바꿀 수 있는 문법적 설탕(Syntactic Sugar)입니다.
JVM에서는 여전히 일반 메서드 호출로 처리됩니다.

사용례:
- 집합 연산 DSL:

    ```kotlin
    infix fun String.append(other: String): String = this + other

    val result = "Hello" append "World"
    println(result)  // "HelloWorld"
    ```

- Kotlin 표준 라이브러리의 `to` 함수:

    ```kotlin
    val map = mapOf(1 to "one", 2 to "two")
    ```

### Inline

`inline` 함수는 호출 시점에 함수의 바디를 호출자 코드로 대체(inline)하여 런타임 오버헤드를 줄입니다.
"라인 안에 삽입"한다는 뜻에서 붙여졌습니다.
이는 JVM의 `method call overhead`를 줄이는 것을 목표로 합니다.
특히 람다식을 매개변수로 전달할 때 유용합니다.
고차 함수를 사용하는 코틀린 표준 라이브러리(`forEach`, `filter`, `map`)에서 성능 최적화를 위해 널리 사용됩니다.

컴파일러는 `inline`으로 표시된 함수를 호출하는 모든 지점에서 함수의 실제 코드를 삽입합니다.
람다식이 매개변수로 전달될 경우, 람다 표현식도 인라인됩니다.

장점:
- 함수 호출 비용을 제거하여 성능을 개선합니다.
- 람다 캡처로 인해 생성되는 추가 객체를 줄입니다.

단점:
- 코드 크기가 증가할 수 있습니다(Code Bloat).
- 지나치게 많은 인라인 함수는 성능에 부정적인 영향을 미칠 수 있습니다.

사용례:

- 고차 함수 최적화:

    ```kotlin
    inline fun performAction(action: () -> Unit) {
        println("Starting action")
        action()
        println("Ending action")
    }

    performAction {
        println("Executing action")
    }
    ```

    위 코드는 `performAction` 함수 바디를 호출자로 대체하여 추가 함수 호출을 방지합니다.

### Noinline

`noinline`은 `inline` 함수의 매개변수 중 특정 람다식을 인라인 처리하지 않도록 지정합니다.
"인라인되지 않는다(no-inline)"는 것을 명확히 하기 위해 붙여졌습니다.
`inline` 함수에서 모든 람다가 인라인되는 기본 동작을 제어할 때 사용됩니다.
람다식을 동적으로 조작해야 하는 프레임워크에서 사용됩니다.

JVM은 다음과 같이 동작합니다.
- `noinline` 람다 매개변수는 일반적인 객체로 처리되며, 실제 호출 시점에 생성됩니다.
- 호출자는 해당 람다를 별도의 객체로 전달받아 사용합니다.

사용례:
- 동작 분리:

    ```kotlin
    inline fun example(block1: () -> Unit, noinline block2: () -> Unit) {
        block1()
        block2()
    }

    example({
        println("Inline block")
    }, {
        println("Noinline block")
    })
    ```

    위 예제에서 `block2`는 일반 객체로 처리됩니다.

### Crossinline

`crossinline`은 `inline` 함수에서 람다식이 비지역(non-local) 리턴을 수행하지 못하도록 제한합니다.
"교차 컨텍스트에서의 인라인 제한"이라는 의미에서 이름이 붙었습니다.
보통 `inline` 함수 내부에서 람다를 다른 컨텍스트로 전달할 때 사용됩니다.
코루틴과 같은 비동기 처리에서 비지역 리턴을 방지하여 안전성을 높입니다.

JVM의 동작 방식은 다음과 같습니다:
- 람다의 리턴 동작을 제한하기 위해 추가적인 검사를 삽입합니다.
- `crossinline` 람다식은 단순히 값으로 캡처되며 비지역 리턴이 허용되지 않습니다.

사용례:
- 안전한 리턴 제한:

    ```kotlin
    inline fun safeExecution(crossinline action: () -> Unit) {
        println("Starting safe execution")
        val runnable = Runnable {
            action()  // 비지역 리턴 불가
        }
        runnable.run()
    }

    safeExecution {
        println("Executing safely")
        // return  // 컴파일 에러
    }
    ```

## 스프링 부트 시작과 빈 라이프 사이클

스프링 부트 애플리케이션은 일반적으로 `SpringApplication.run()` 메서드를 호출하여 시작됩니다.
이 과정은 크게 SpringApplication 초기화, 환경 설정, 애플리케이션 컨텍스트 생성 및 부트스트랩, 그리고 빈 초기화 및 실행으로 나뉩니다.

### SpringApplication.run()

- `SpringApplication.run()` 호출

    1. `SpringApplication` 객체 생성:
        - 애플리케이션이 실행되면 `SpringApplication` 클래스의 정적 메서드 `run()`이 호출됩니다.
        - 내부적으로 `SpringApplication` 인스턴스가 생성되며, 애플리케이션의 실행 환경을 설정합니다.

    2. 스프링 부트 설정 초기화:
        - 기본적으로 실행될 애플리케이션 타입을 결정합니다.
            - `REACTIVE`: 웹플럭스 기반 애플리케이션.
            - `SERVLET`: 일반적인 웹 MVC 기반 애플리케이션.
            - `NONE`: 콘솔 기반 애플리케이션.
        - 실행 환경(`SpringApplicationEnvironment`)과 컨텍스트 타입(`ApplicationContext`)을 초기화.

- `prepareEnvironment()`

    1. 프로퍼티 및 설정 로드:
        - `application.properties` 또는 `application.yml` 파일을 읽어 Spring 환경에 추가.
        - 운영 체제 환경 변수, JVM 시스템 속성, 외부 프로퍼티 소스 등을 결합하여 스프링 환경(`Environment`) 객체에 저장.

    2. EnvironmentPostProcessor 호출:
        - 환경 설정을 커스터마이징할 수 있는 확장 포인트를 실행.

- `prepareContext()`

    1. ApplicationContext 생성:
        - 스프링 컨텍스트(ApplicationContext)가 생성됩니다. 기본적으로 다음 중 하나를 선택:
            - `AnnotationConfigServletWebServerApplicationContext`: 서블릿 기반 애플리케이션.
            - `ReactiveWebServerApplicationContext`: 웹플럭스 기반 애플리케이션.
            - `GenericApplicationContext`: 일반 콘솔 애플리케이션.

    2. BeanDefinition 로드:
        - `@ComponentScan` 및 `@EnableAutoConfiguration`에 의해 클래스 경로에서 빈 정의가 로드됩니다.

    3. ApplicationContextInitializer 호출:
        - `SpringApplication`에 등록된 초기화기(ApplicationContextInitializer)가 실행됩니다.
        - 컨텍스트를 추가적으로 초기화할 수 있는 확장 포인트.

- `refreshContext()`

    1. ApplicationContext의 `refresh()` 호출:
        - `ApplicationContext`의 핵심 메서드로, 빈 팩토리 초기화와 빈 라이프사이클의 시작점이 됩니다.
        - 아래에서 자세히 설명할 빈 라이프사이클의 대부분이 이 단계에서 진행됩니다.

    2. CommandLineRunner 및 ApplicationRunner 실행:
        - `refresh()` 이후, `CommandLineRunner`와 `ApplicationRunner` 인터페이스를 구현한 빈들이 실행됩니다.
        - 이 단계는 애플리케이션 로직을 실행할 수 있는 지점을 제공합니다.

- `run()` 종료
    - 스프링 부트 애플리케이션이 완전히 초기화되고 실행 준비가 완료됩니다.
    - 애플리케이션이 HTTP 요청을 수신 대기하거나 백그라운드 작업을 시작합니다.

### 스프링 빈의 라이프사이클

스프링 컨테이너는 빈을 생성, 초기화, 사용, 소멸시키는 일련의 과정을 관리합니다. 이 과정은 다음과 같은 단계로 구성됩니다.

- 빈 정의 등록 (Bean Definition)
    1. 클래스 경로 스캔:
        - `@Component`, `@Configuration`, `@Service`, `@Repository`, `@Controller` 등이 선언된 클래스를 탐색.
        - 이러한 클래스는 빈 정의(Bean Definition)으로 컨테이너에 등록됩니다.

    2. 빈 팩토리 초기화:
        - 스프링 컨테이너는 빈 정의를 기반으로 빈 팩토리(BeanFactory)를 초기화합니다.
        - 빈 정의에는 빈의 클래스 타입, 생성 방법, 의존성 정보 등이 포함됩니다.

- 빈 생성 및 초기화

    빈 라이프사이클의 주요 단계는 다음과 같습니다:

    1. 빈 인스턴스 생성
        - 스프링 컨테이너가 빈 정의에 따라 빈 인스턴스를 생성합니다.
        - 기본적으로 리플렉션을 사용하여 객체를 생성하며, 생성자 주입 시 필요한 의존성을 해결합니다.

    2. 의존성 주입 (Dependency Injection)
        - 생성된 빈에 `@Autowired`, `@Value` 등을 통해 의존성을 주입합니다.
        - 의존성은 생성자, 세터 메서드, 또는 필드를 통해 주입될 수 있습니다.

    3. 빈 초기화 (Initialization)
        1. `Aware` 인터페이스 처리:
            - 빈이 특정 스프링 컨텍스트 정보에 접근해야 하는 경우 `Aware` 인터페이스를 구현합니다.
            - 예: `BeanNameAware`, `ApplicationContextAware`.
            - 스프링이 빈에 컨텍스트 정보를 주입합니다.

        2. `BeanPostProcessor`의 `postProcessBeforeInitialization` 호출:
            - 빈 초기화 전에 호출되는 사용자 정의 로직 실행.

        3. 초기화 메서드 실행:
            - `@PostConstruct`로 지정된 메서드 실행.
            - `InitializingBean` 인터페이스의 `afterPropertiesSet()` 호출.
            - XML이나 Java Config에서 `init-method`로 설정된 메서드 실행.

        4. `BeanPostProcessor`의 `postProcessAfterInitialization` 호출:
            - 초기화 후 추가 처리 작업을 수행.

- 빈 사용
    - 애플리케이션에서 빈이 사용되는 단계입니다.
    - 의존성 주입이 완료된 빈은 다른 빈이나 컴포넌트에 의해 사용됩니다.

- 빈 소멸 (Destruction)
    1. 소멸 전 작업:
        - `@PreDestroy`로 지정된 메서드 실행.
        - `DisposableBean` 인터페이스의 `destroy()` 호출.
        - XML이나 Java Config에서 `destroy-method`로 설정된 메서드 실행.

    2. 빈 제거:
        - 스프링 컨테이너가 빈에 대한 참조를 제거하여, 해당 객체가 GC(Garbage Collection) 대상이 되도록 만듭니다.

## 함수형

함수형 프로그래밍(Functional Programming, FP)은 수학적 함수의 개념에서 영감을 받은 프로그래밍 패러다임으로, 코드의 명확성, 재사용성, 안전성을 강조합니다.

순수 함수(Pure Function), 불변성(Immutability), 고차 함수(Higher-Order Functions), 일급 객체(First-Class Citizens) 등의 원칙을 중심으로 하는 프로그래밍 패러다임입니다.

1. 순수 함수(Pure Function):
    - 동일한 입력에 대해 항상 동일한 결과를 반환하며, 외부 상태를 변경하지 않음.
    - 부수효과(Side Effects)를 피함으로써 프로그램의 예측 가능성을 증가시킴.

2. 불변성(Immutability):
    - 데이터와 상태는 변경할 수 없으며, 변경하려면 복사본을 생성.
    - 이는 데이터 경쟁(Data Race)과 같은 멀티스레드 이슈를 방지.

3. 고차 함수(Higher-Order Function):
    - 함수를 값처럼 취급하며, 이를 다른 함수의 인자로 전달하거나 반환값으로 사용할 수 있음.

4. 일급 객체(First-Class Citizen):
    - 함수가 변수처럼 할당되고, 다른 함수의 인자나 반환값으로 사용될 수 있음.

5. 참조 투명성(Referential Transparency):
    - 표현식이 항상 동일한 값을 반환하여, 호출을 그 결과값으로 대체해도 프로그램의 의미가 유지됨.

6. 부수효과의 최소화(Side Effects Minimization):
    - 함수의 결과에 영향을 미치지 않는 외부 상태 변경을 최소화.

### 함수형 프로그래밍의 이론적 개념

- Functor
    - 컨테이너 안의 값을 매핑(map)하는 추상화.
    - 값을 직접 변경하지 않고, 컨테이너의 구조를 유지하면서 값을 변환.

    이론적 정의:
    - `map` 연산을 지원하며, 다음 법칙을 만족:
        - 항등 법칙: `F.map(id) == F`
        - 합성 법칙: `F.map(f).map(g) == F.map(f ∘ g)`

- Monad
    - 연속적인 계산을 모델링하는 추상화.
    - Functor를 확장하여, 컨테이너 안의 값을 연결(bind)하고 중첩을 평평하게(flatten) 만듦.

    이론적 정의:
    - `flatMap`(또는 `bind`) 연산을 지원하며, 다음 법칙을 만족:
        - 좌항등 법칙: `M.pure(a).flatMap(f) == f(a)`
        - 우항등 법칙: `M.flatMap(M.pure) == M`
        - 결합 법칙: `M.flatMap(f).flatMap(g) == M.flatMap { x -> f(x).flatMap(g) }`

- 부수효과의 제어
    - Effect Handling: 부수효과를 함수형으로 모델링하여 안전하게 관리.
        - 예: I/O 작업, 네트워크 요청.

### 3.6 Either를 통한 에러 처리

#### 이론 적용

`Either`는 성공과 실패를 명시적으로 처리하는 함수형 대안입니다.

#### 실무 예제

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

fun divide(a: Int, b: Int): Either<String, Int> =
    if (b == 0) "Division by zero".left()
    else (a / b).right()

val result = divide(10, 2) // Right(5)
val error = divide(10, 0) // Left("Division by zero")
```

### 3.7 I/O 작업 제어

#### 이론 적용

부수효과를 함수형으로 모델링하여 안전하게 처리.

#### Arrow-kt에서의 I/O 모델링

```kotlin
import arrow.fx.IO

val program = IO {
    println("Performing side-effect")
    "Result"
}

val result = program.unsafeRunSync() // "Performing side-effect" 출력
```

### 순수 함수 (Pure Function)

동일한 입력에 대해 항상 동일한 출력을 반환하며, 외부 상태를 변경하지 않는 함수.

```kotlin
// 순수 함수는 외부 상태에 의존하지 않으며, 동일한 입력에 대해 항상 동일한 결과를 반환합니다.
fun calculateDiscount(price: Double, discountRate: Double): Double =
    price * discountRate

// 외부 상태에 의존하지 않음
val result = calculateDiscount(100.0, 0.1) // 항상 10.0 반환

fun add(a: Int, b: Int): Int = a + b
```

### 불변성 (Immutability)

함수형 프로그래밍에서는 상태를 변경하지 않고 새로운 값을 생성합니다.

```kotlin
// 데이터를 변경하지 않고, 복사본을 생성하여 상태 변경을 구현.
val originalList = listOf(1, 2, 3)
val updatedList = originalList + 4 // [1, 2, 3, 4]

// 원본은 변경되지 않음
println(originalList) // [1, 2, 3]

val list = listOf(1, 2, 3)
val newList = list.map { it * 2 } // 원래 list는 변경되지 않음
```

### 고차 함수 (Higher-Order Functions)

다른 함수를 인자로 받거나 반환하는 함수.

```kotlin
// 고차 함수는 함수를 인자로 받거나 반환값으로 사용하여 동작을 추상화합니다.
fun applyOperation(numbers: List<Int>, operation: (Int) -> Int): List<Int> {
    return numbers.map(operation)
}

// 두 배로 만드는 함수 전달
val doubled = applyOperation(listOf(1, 2, 3)) { it * 2 } // [2, 4, 6]

fun applyTwice(f: (Int) -> Int, x: Int): Int = f(f(x))

val result = applyTwice({ it * 2 }, 5) // 20
```

### 일급 객체 (First-Class Citizens)

함수가 변수에 할당되거나, 다른 함수의 인자로 전달될 수 있음.

```kotlin
val double: (Int) -> Int = { it * 2 }
println(double(4)) // 8
```

### Functor

컨테이너 안의 값을 매핑할 수 있는 추상화.

Kotlin의 `map` 함수는 Functor의 예입니다.

```kotlin
val list = listOf(1, 2, 3)
val doubled = list.map { it * 2 } // [2, 4, 6]
```

Arrow는 `Option`, `Either`와 같은 컨테이너 타입에 Functor 연산을 제공합니다.

```kotlin
import arrow.core.Option
import arrow.core.some

val option: Option<Int> = 3.some()
val result = option.map { it * 2 } // Option(6)
```

### Monad

컨테이너 안의 값을 매핑할 뿐 아니라 컨테이너의 중첩을 평평하게(flatten) 하는 추상화.
`bind` 또는 `flatMap`을 통해 중첩된 컨테이너를 단일 컨테이너로 변환합니다.

```kotlin
// 컨테이너 안의 중첩된 값을 평평하게(flatten) 하여 연결.
val nested = listOf(listOf(1, 2), listOf(3, 4))
val flattened = nested.flatMap { it } // [1, 2, 3, 4]
```

Arrow-kt에서의 Monad:

```kotlin
import arrow.core.Option
import arrow.core.some

val opt1: Option<Int> = 10.some()
val result = opt1.flatMap { x -> Option.just(x * 2) } // Option(20)
```

```kotlin
import arrow.core.Option
import arrow.core.none
import arrow.core.some

val opt1: Option<Int> = 10.some()
val opt2: Option<Int> = none()

val result = opt1.flatMap { x -> opt2.map { y -> x + y } } // None
```

### Task (Deferred in Kotlin)

비동기 작업을 표현하며, 결과를 지연 평가(Lazy Evaluation)로 처리.

Kotlin의 예:

```kotlin
import kotlinx.coroutines.*

suspend fun fetchUser(): String {
    delay(1000) // Simulating long-running task
    return "User"
}

fun main() = runBlocking {
    val task = async { fetchUser() }
    println(task.await()) // "User"
}
```

Arrow의 `IO` 모나드 또는 `Task`는 비동기 작업을 함수형 스타일로 처리하는 데 사용됩니다.

```kotlin
import arrow.fx.coroutines.*

suspend fun main() {
    val task = IO.invoke { "Result" }
    println(task.unsafeRunSync()) // "Result"
}
```

### Either

실패 또는 성공을 표현하는 양자 타입.
`Either.Left`는 오류를, `Either.Right`는 성공 값을 나타냅니다.

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

fun divide(a: Int, b: Int): Either<String, Int> =
    if (b == 0) "Division by zero".left()
    else (a / b).right()

val result = divide(10, 2)
println(result) // Right(5)
```

`bind()`는 Arrow의 `Monad` 컴퓨테이션을 처리하는 DSL(do notation)에서 사용되는 메서드로, Monad 컨텍스트에서 값을 꺼내고 연속적인 계산을 처리합니다.
Monad 컨텍스트란, 예를 들어, `Either`, `Option`, `IO`와 같은 Monad 타입에서 값을 직접 꺼낼 수 없는 상황에서, `bind()`를 통해 안전하게 값을 꺼내 계산을 이어갑니다.

`Either`는 성공(`Right`)과 실패(`Left`)를 표현하는 타입입니다.

```kotlin
import arrow.core.Either
import arrow.core.continuations.either
import arrow.core.left
import arrow.core.right

fun fetchUserData(userId: String): Either<String, String> =
    if (userId == "123") "User data".right() else "User not found".left()

fun fetchUserPosts(userId: String): Either<String, List<String>> =
    if (userId == "123") listOf("Post1", "Post2").right() else "No posts".left()

suspend fun fetchUserWithPosts(userId: String): Either<String, Pair<String, List<String>>> = either {
    val userData = fetchUserData(userId).bind() // 값을 안전하게 꺼냄
                                                // `bind()`는 Monad(`Either`) 내부에서 값을 꺼내 다음 계산에 전달.
                                                // 만약 `Left`가 발생하면 즉시 실행을 중단하고 에러를 반환.
    val userPosts = fetchUserPosts(userId).bind()
    userData to userPosts
}

fun main() {
    val result = fetchUserWithPosts("123")
    println(result) // Right((User data, [Post1, Post2]))
}
```

Arrow를 사용하여 HTTP 통신 결과를 처리하는 경우도 유사합니다.
예를 들어, HTTP 요청이 `Either` 타입으로 성공과 실패를 반환할 때:

```kotlin
import arrow.core.Either
import arrow.core.continuations.either

suspend fun fetchHttpResponse(url: String): Either<String, String> =
    if (url.startsWith("http")) "Response from $url".right()
    else "Invalid URL".left()

suspend fun fetchData(): Either<String, String> = either {
    val response = fetchHttpResponse("http://example.com").bind() // HTTP 응답을 처리
                                                                  // `bind()`는 성공(`Right`) 값을 꺼내 계산을 이어가며, 실패(`Left`)가 발생하면 즉시 반환.
    "Processed: $response"
}

suspend fun main() {
    val result = fetchData()
    println(result) // Right(Processed: Response from http://example.com)
}
```

`bind()`가 HTTP 통신에서 사용된다면, 이는 Arrow와 같은 함수형 라이브러리를 활용하여 안전한 에러 처리와 Monad 값을 관리하는 방식으로 구현되었을 가능성이 큽니다. 이 방식은 다음과 같은 이점을 제공합니다:

1. 명시적 에러 처리:
    - `Either`와 같은 Monad 타입을 통해, 성공과 실패를 명확히 처리.

2. 안전한 값 추출:
    - `bind()`는 Monad 내부 값을 안전하게 꺼내고, 실패가 발생하면 자동으로 에러를 반환.

3. 연속적인 계산 모델링:
    - 여러 HTTP 요청이 연속적으로 연결될 때, `bind()`를 사용해 간결하고 명확하게 코드 작성 가능.

### Arrow-kt 주요 데이터 타입 및 함수

Arrow-kt는 함수형 프로그래밍의 개념을 구현하는 여러 데이터 타입과 도구를 제공합니다.

- `Option`
    - 정의: 값이 있을 수도, 없을 수도 있는 경우를 표현.
    - `Option.Some`은 값을 나타내고, `Option.None`은 값이 없음을 나타냅니다.

    ```kotlin
    import arrow.core.Option
    import arrow.core.some
    import arrow.core.none

    val someValue: Option<Int> = 5.some()
    val noValue: Option<Int> = none()

    val result = someValue.getOrElse { 0 } // 5
    val fallback = noValue.getOrElse { 0 } // 0
    ```

- `Validated`

    실패 또는 성공을 표현하며, 실패를 누적할 수 있음.

    ```kotlin
    import arrow.core.Validated
    import arrow.core.invalid
    import arrow.core.valid

    val success = 42.valid()
    val failure = "Error".invalid()

    println(success) // Valid(42)
    println(failure) // Invalid(Error)
    ```

- `IO`

    효과(effect)를 표현하며, 부수 효과(side-effect)를 안전하게 처리.

    ```kotlin
    import arrow.fx.IO

    val io = IO { println("Running IO") }
    io.unsafeRunSync() // 실제로 실행
    ```

- Applicative

    독립적인 계산을 결합하여 병렬 실행 지원.

    ```kotlin
    import arrow.core.Tuple2
    import arrow.core.extensions.list.apply.map

    val listA = listOf(1, 2, 3)
    val listB = listOf("A", "B", "C")

    val combined = listA.map(listB) { a, b -> "$a$b" }
    println(combined) // [1A, 1B, 1C, 2A, 2B, 2C, 3A, 3B, 3C]
    ```

## HTTP 요청이 일반적인 TCP 연결 방식과 Keep-Alive TCP 연결 방식 이루어지는 경우

### 서버 A에 HTTP 요청 2회 (일반적인 TCP 연결) 경우

1. TCP 연결 시작
    1. 클라이언트는 서버 A에 첫 번째 요청을 보냅니다.
        - TCP 3-way handshake 시작:
            1. 클라이언트 -> 서버 A: `SYN` 패킷 전송 (소스 IP:포트, 대상 IP:포트로 식별되는 연결 요청).
            2. 서버 A -> 클라이언트: `SYN-ACK` 응답.
            3. 클라이언트 -> 서버 A: `ACK` 전송.
        - 연결 상태가 `ESTABLISHED`로 변경.

    2. 클라이언트가 HTTP 요청 데이터를 전송합니다.
        - TCP는 HTTP 요청 데이터를 소켓의 송신 버퍼에 넣고, 이를 네트워크를 통해 전송.
        - 서버 A는 데이터를 수신하고 HTTP 응답을 반환.

    3. HTTP 요청 완료 후, 클라이언트는 연결을 종료합니다.
        - TCP 4-way handshake 시작:
            1. 클라이언트 -> 서버 A: `FIN` 패킷 전송.
            2. 서버 A -> 클라이언트: `ACK` 응답.
            3. 서버 A -> 클라이언트: `FIN` 전송.
            4. 클라이언트 -> 서버 A: `ACK` 응답.
        - 클라이언트는 연결 종료 후 `TIME_WAIT` 상태로 들어갑니다.

    4. 두 번째 HTTP 요청을 위해 동일한 과정을 반복:
        - 새로운 TCP 연결이 설정되고, HTTP 요청/응답이 처리된 후 종료됩니다.

    `LISTEN` -> `SYN_RECV` -> `ESTABLISHED` 반복

2. 네트워크 큐
    - 첫 번째 요청:
        - 클라이언트의 SYN 패킷은 서버 A의 SYN 큐에 추가됩니다.
        - 서버 A가 SYN-ACK를 응답하고 클라이언트로부터 ACK를 수신하면, 연결은 SYN 큐에서 제거됩니다.
    - 두 번째 요청:
        - 동일한 흐름으로 별도의 SYN 큐 항목이 생성됩니다.
    - 서버 A는 두 요청을 각각 독립적으로 처리하며, 두 요청 사이에 네트워크 큐의 상태는 초기화됩니다.

3. 커널 소켓 상태 (서버 A와 클라이언트 모두)
    1. `LISTEN` (서버 A):
        - 서버 A가 수신 대기 상태.
        - 서버 A는 클라이언트의 연결 요청을 대기하며 `LISTEN` 상태에서 요청을 대기.
    2. `SYN_RECV` (서버 A):
        - 클라이언트의 연결 요청을 처리 중.
        - 첫 번째 요청이 들어오면, 서버 A는 `SYN`을 처리하며 `SYN_RECV` 상태로 전환.
    3. `ESTABLISHED` (양쪽):
        - 연결이 성공적으로 설정되어 데이터 송수신 가능.
        - 클라이언트와 서버 A가 데이터 송수신을 시작하면, 양쪽 소켓은 `ESTABLISHED` 상태.
    4. `CLOSE_WAIT` (서버 A):
        - 연결 종료 과정은 `FIN_WAIT_1` -> `CLOSE_WAIT` -> `TIME_WAIT`
        - 클라이언트가 연결 종료를 요청하면, 서버 A는 이를 수락하고 `CLOSE_WAIT` 상태.
    5. `TIME_WAIT` (클라이언트):
        - 연결 종료 후, 클라이언트는 `TIME_WAIT` 상태로 지정된 시간 동안 포트를 유지.

    HTTP 요청마다 독립적인 연결을 생성하고 종료하므로, 각 요청에 대해 상태 전환이 반복됩니다.
    `TIME_WAIT` 상태는 네트워크 패킷이 중복되거나 지연될 가능성을 방지하기 위해 존재.
    두 요청 각각에 대해 동일한 소켓 상태 전환이 발생.

### 서버 B에 HTTP 요청 2회 (Keep-Alive TCP 연결)

1. TCP 연결 시작
    1. 클라이언트는 서버 B에 첫 번째 요청을 보냅니다.
        - TCP 3-way handshake로 연결 설정.
        - 연결 상태가 `ESTABLISHED`로 변경.

    2. HTTP 요청 데이터가 전송됩니다.
        - 클라이언트는 소켓의 송신 버퍼에 데이터를 작성.
        - 서버 B는 데이터를 수신하고 응답을 반환.

    3. 클라이언트는 HTTP 요청 완료 후 연결을 유지합니다.
        - Keep-Alive 설정으로 TCP 연결이 종료되지 않음.
        - 두 번째 요청은 기존 TCP 연결을 재사용.

    `LISTEN` -> `SYN_RECV` -> `ESTABLISHED` 유지

2. 네트워크 큐
    - 첫 번째 요청:
        - 클라이언트의 SYN 패킷은 서버 B의 SYN 큐에 추가.
        - 연결이 설정된 후 큐에서 제거.
    - 두 번째 요청:
        - 네트워크 큐가 관여하지 않음. 기존 TCP 연결을 통해 데이터 송수신.

3. 커널 소켓 상태 (서버 B와 클라이언트 모두)
    1. `LISTEN` (서버 B):
        - 서버 B가 첫 요청을 대기.
        - 서버 B는 클라이언트 연결 요청을 수락하기 위해 대기.
    2. `SYN_RECV` (서버 B):
        - 클라이언트 연결 요청 처리 중.
        - 첫 번째 요청이 들어오면, 서버 B는 `SYN`을 처리하며 `SYN_RECV` 상태.
    3. `ESTABLISHED` (양쪽):
        - 연결 설정 및 데이터 송수신 상태 유지.
        - 클라이언트와 서버 B가 데이터를 송수신하는 동안 연결 유지.
    4. `ESTABLISHED` 유지:
        - 두 번째 요청은 동일한 연결에서 처리되므로 추가 상태 변화 없음.
    5. `TIME_WAIT` (클라이언트):
        - 모든 요청 처리 완료 후 연결 종료.
        - 클라이언트가 연결 종료 시점에서만 `TIME_WAIT` 상태에 들어감.

    Keep-Alive 연결은 소켓을 재사용하여 효율성을 극대화.
    연결 종료 시점에만 상태 전환이 발생.

### Keep-Alive 타임아웃과 클라이언트가 FIN 패킷을 보내지 않을 경우의 처리 (OS 및 소켓 관점)

Keep-Alive는 TCP 연결을 일정 시간 동안 유지하면서 비활성 상태인지 확인하는 메커니즘입니다.
클라이언트가 `FIN` 패킷을 보내지 않고, Keep-Alive 타임아웃이 발생하는 경우, 서버와 클라이언트 모두 TCP 소켓과 운영 체제(OS) 수준에서 각각의 상태를 관리하게 됩니다.

TCP Keep-Alive의 기본 동작:
1. Keep-Alive 메시지:
    - TCP 연결은 `Keep-Alive` 설정이 활성화되었을 때, 비활성 상태 동안 정기적으로 `Keep-Alive` 패킷을 보냅니다.
    - 이 패킷은 연결이 여전히 유효한지 확인하며, 클라이언트 또는 서버가 응답하지 않을 경우 연결 종료를 트리거합니다.

2. Keep-Alive 타이머:
    - OS는 연결이 비활성 상태로 유지되는 시간을 측정하기 위해 타이머를 사용합니다.
    - `Keep-Alive` 타이머는 두 가지 주요 시간 설정으로 구성됩니다:
        - Keep-Alive Time: 마지막 데이터 전송 후 Keep-Alive 패킷을 보내기까지의 대기 시간.
        - Keep-Alive Interval: Keep-Alive 패킷에 응답이 없을 경우, 다음 패킷을 보내기 전 대기 시간.
    - Keep-Alive Retries: 특정 횟수 동안 응답이 없으면 연결을 종료.

클라이언트가 FIN 패킷을 보내지 않을 경우:
- 클라이언트와 서버 간 상태
    - 클라이언트는 명시적으로 연결을 닫지 않고(`FIN` 없음), 서버와의 연결은 여전히 유지된 상태.
    - 서버는 Keep-Alive 설정에 따라 비활성 연결을 모니터링.

- OS 및 소켓 관점

    - Keep-Alive 타이머 만료 전 (정상 상태)
        1. Keep-Alive 패킷 전송:

            Keep-Alive 설정은 TCP 연결이 비활성 상태로 유지될 때 연결 상태를 확인하는 목적으로 사용됩니다.
            Keep-Alive 메시지를 보내는 책임은 TCP 스택 설정에 따라 달라집니다.
            - 서버 책임:
                서버는 연결이 오래된 경우, 클라이언트가 응답하는지 확인하기 위해 Keep-Alive 메시지를 보냅니다.
                서버가 Keep-Alive를 설정하고, 클라이언트의 응답이 없으면 연결을 닫습니다.
            - 클라이언트 책임:
                클라이언트도 Keep-Alive를 설정하고 주기적으로 서버에 메시지를 보낼 수 있습니다.

            - 서버는 Keep-Alive 타이머에 따라 클라이언트로 TCP Keep-Alive 패킷(TCP 헤더의 ACK 플래그가 설정된 빈 패킷)을 보냅니다.
            - 클라이언트가 응답(`ACK`)하면 연결 상태를 유지합니다.

        2. 클라이언트 응답 없음:
            - 클라이언트가 다운되었거나 네트워크 장애로 인해 응답하지 않을 경우, 서버는 Keep-Alive 패킷을 반복적으로 보냅니다.
            - 이 과정에서 Keep-Alive Interval과 Retries 설정이 사용됩니다.

- Keep-Alive 타이머 만료 후 (연결 종료)
    1. 서버 측 동작:
        - 서버는 일정 횟수의 Keep-Alive 패킷 전송 후 응답이 없으면 연결을 닫습니다.
        - TCP 소켓은 `TIME_WAIT` 상태로 전환되어 일정 시간 동안 대기(기본적으로 2MSL, Maximum Segment Lifetime)합니다.
        - `TIME_WAIT` 동안 서버는 해당 연결에 대한 재전송된 패킷을 처리하지 않고 무시.

    2. 소켓 및 리소스 해제:
        - `TIME_WAIT` 상태가 끝나면 소켓이 완전히 닫히고 커널에서 관련 리소스를 해제합니다.

    3. 클라이언트 측 동작:
        - 클라이언트는 `FIN`을 보내지 않았기 때문에, 클라이언트의 OS는 여전히 연결 상태를 유지합니다.
        - 하지만 Keep-Alive 메시지가 서버에서 오지 않으므로, 연결이 끊긴 것으로 간주.
        - 클라이언트 소켓은 `CLOSE_WAIT` 상태에 머물거나, 일정 시간 후 타임아웃이 발생하여 리소스를 해제합니다.

아래는 TCP 상태 전이와 Keep-Alive 관련 상태를 요약한 다이어그램입니다:

```plaintext
    ESTABLISHED
        |
        v
[Keep-Alive 패킷 전송]
        |
    응답 있음 -> 연결 유지
        |
    응답 없음
        |
[Keep-Alive Retries 초과]
        |
        v
    TIME_WAIT (서버)
        |
        v
 연결 종료 및 리소스 해제
```
