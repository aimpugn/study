# 02. 동시성, 비동기, I/O

스레드, 락, 대기, 이벤트 루프, 논블로킹 I/O처럼 많은 요청을 안전하고 효율적으로 처리하는 실행 모델을 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## Blocking, non-blocking, async 구분

### HTTP 요청을 Async로 처리했을 때의 이점

#### 원문: HTTP 요청을 Async로 처리했을 때의 이점

<!-- curriculum-chunk: sha256=ba19407a52452a019f15b8d1a639fe4ec588c6cc10c49fe5dcbcfe504cdf57eb major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=HTTP 요청을 Async로 처리했을 때의 이점 sources=interview_questions2.md:439-541, interviews2.md:439-541 -->

> Source: `interview_questions2.md:439-541`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions2.md:439-541, interviews2.md:439-541`

##### HTTP 요청을 Async로 처리했을 때의 이점

###### 블로킹 스레드 감소

- 블로킹 모델:
    - 스레드 기반 요청 처리에서는 HTTP 요청 응답을 기다리기 위해 스레드가 블로킹 상태로 대기합니다.
    - 이는 스레드가 다른 작업을 수행하지 못하고, 리소스를 비효율적으로 사용하게 만듭니다.

- 비동기 모델:
    - HTTP 요청을 Async로 처리하면, 스레드가 블로킹 상태에 머물지 않고, 다른 요청을 처리하거나 반환될 수 있습니다.
    - 실제 HTTP 요청이 완료되면 콜백이나 코루틴을 통해 결과를 받아 후속 작업을 이어갑니다.
    - 이점:
        - 더 적은 스레드로 더 많은 요청을 처리 가능.
        - I/O 대기 시간이 긴 애플리케이션에서 특히 효과적.

###### 스레드 효율성 및 확장성 증가

- 스레드 기반 처리는 요청 수가 많아질수록 스레드 풀이 포화 상태에 이르러 성능이 급격히 저하될 수 있습니다.
- 비동기 처리에서는 요청 대기 시간을 비동기 이벤트 루프(예: Netty, Project Reactor)로 처리하므로 스레드 수 증가 없이 높은 요청 처리량을 유지할 수 있습니다.
- 예: 100개의 요청을 처리할 때, 10개의 스레드만으로도 충분히 처리 가능.

###### 시스템 응답성 향상

- 블로킹 작업이 줄어들면서 스레드 풀이 더 빠르게 반환되어 새로운 요청에 대해 더 빠르게 반응할 수 있습니다.
- 예: 사용자 A의 요청이 비동기로 처리되어 스레드를 반환하면, 사용자 B의 요청에 즉시 대응 가능.

###### 비용 절감

- 스레드는 메모리와 CPU를 소비하는 고비용 리소스입니다.
- 비동기 처리는 적은 수의 스레드로 더 많은 요청을 처리할 수 있으므로, 하드웨어 요구사항이 줄어들어 인프라 비용을 절감할 수 있습니다.

###### 코드 가독성과 유지보수성

- 비동기 처리를 코루틴이나 `CompletableFuture`와 같은 고수준 API로 구현하면, 코드가 읽기 쉽고 유지보수하기 쉬워집니다.
- 스프링 웹플럭스(Spring WebFlux)와 같은 프레임워크는 코루틴과의 통합으로 더욱 간결한 비동기 코드를 제공합니다.

###### Async 처리의 실제 예제

###### 블로킹 HTTP 요청

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

###### 비동기 HTTP 요청 (WebClient)

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

###### 코루틴 기반 HTTP 요청

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

<!-- /curriculum-chunk -->

### I/O 서브시스템

#### 원문: I/O 서브시스템

<!-- curriculum-chunk: sha256=91014364d9fab87a6f61ccb6eda97432a5593d7d2479beb11f549912c396daa4 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=I/O 서브시스템 sources=interview_questions.md:4524-4582, interviews.md:4524-4582 -->

> Source: `interview_questions.md:4524-4582`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions.md:4524-4582, interviews.md:4524-4582`

##### I/O 서브시스템

I/O 서브시스템은 컴퓨터 시스템에서 입출력(Input/Output) 작업을 처리하는 전체적인 구성 요소들의 집합을 의미합니다.
이 시스템은 애플리케이션과 하드웨어(디스크, 네트워크, 기타 I/O 장치) 사이에서 입출력 요청을 효율적으로 관리합니다.
일반적으로 운영체제의 커널이 관리하며, I/O 작업을 처리하고 성능을 최적화하는 역할을 합니다.

I/O 서브시스템은 여러 부분으로 나뉘며, 주요 구성 요소들은 다음과 같습니다:

- 파일 시스템 (File System)

    I/O 작업이 디스크에서 읽고 쓰는 데이터를 처리할 때 사용되는 계층입니다.
    파일 시스템은 데이터를 파일로 관리하며, 파일 읽기, 쓰기 등의 I/O 요청을 처리합니다.

- 버퍼 캐시 (Buffer Cache)

    입출력 성능을 향상시키기 위한 메모리 영역입니다.
    데이터를 디스크로부터 직접 읽거나 쓰는 대신, 자주 사용하는 데이터를 메모리에 캐싱하여 성능을 향상시킵니다.

- 장치 드라이버 (Device Driver)

    장치 드라이버는 하드웨어와 커널 사이에서 하드웨어 장치에 접근하고 제어할 수 있는 인터페이스를 제공합니다.

    예를 들어, 네트워크 카드, 디스크 드라이브, USB 장치 등이 해당됩니다.
    장치 드라이버는 I/O 요청을 하드웨어 명령으로 변환하여 실제 I/O 작업이 이루어지게 합니다.

- I/O 스케줄러

    커널의 I/O 서브시스템은 여러 I/O 요청이 들어왔을 때, 이를 최적화하여 처리하기 위해 스케줄링합니다.
    I/O 스케줄러는 성능을 극대화하기 위해 어떤 순서로 I/O 작업을 처리할지를 결정합니다.

    예를 들어, 디스크에서 데이터를 읽을 때는 디스크 헤드 이동을 최소화하는 방향으로 최적화합니다.

- 인터럽트 처리 (Interrupt Handling)

    I/O 작업이 완료되면 하드웨어는 인터럽트를 발생시켜 운영체제에 알립니다.
    커널은 이 인터럽트를 처리하여 작업이 완료되었음을 애플리케이션에 알리며, 필요한 후속 작업을 수행합니다.

- DMA (Direct Memory Access)

    I/O 서브시스템의 중요한 부분 중 하나는 DMA입니다.
    이는 데이터를 CPU의 개입 없이 메모리와 I/O 장치 사이에서 직접 전송할 수 있게 해주는 기술로,
    CPU 자원을 절약하고 입출력 성능을 크게 향상시킵니다.

    Direct memory access (DMA)는 메모리와 I/O 장치 사이에 채널을 설정하여 프로세서를 우회하기 때문에 CPU 없이도 가능합니다.

    DMA 컨트롤러는 시스템 버스를 제어하여 CPU 없이 메모리에 직접 액세스할 수 있습니다.
    그런 다음 컨트롤러는 특정 메모리 주소에 데이터를 읽거나 씁니다.
    DMA를 사용하면 메모리 위치와 주변 장치 간에 효율적이고 빠른 데이터 전송이 가능합니다. 또한 CPU의 오버헤드도 줄여줍니다.

    버스 마스터링 시스템에서는 CPU와 주변 장치가 모두 메모리 버스를 제어할 수 있습니다.
    주변 장치가 버스 마스터가 되어 CPU의 개입 없이 시스템 메모리에 직접 쓸 수 있습니다.

- I/O 서브시스템의 동작 예시
    - 애플리케이션이 파일을 읽으려는 요청을 하면, 이 요청은 파일 시스템을 거쳐 버퍼 캐시에서 먼저 데이터를 확인합니다.
    - 캐시에 데이터가 없으면, 요청은 장치 드라이버를 통해 디스크로 전달되고, 디스크에서 데이터를 읽은 후 인터럽트를 통해 커널에 작업 완료를 알립니다.
    - 이 과정에서 DMA는 CPU 없이도 데이터를 메모리에 적재하고, I/O 스케줄러는 여러 요청을 최적화된 순서로 처리합니다.

더 자세한 내용은 [Linux I/O Architecture](https://www.kernel.org/doc/html/latest/block/index.html) 등을 참고합니다.

<!-- /curriculum-chunk -->

### Kotlin의 async HTTP 요청과 코루틴 비교

#### 원문: Kotlin의 async HTTP 요청과 코루틴 비교

<!-- curriculum-chunk: sha256=0bc1c669012932f489656618f282d0df26cf0193c741b0224f33c6efe970cb91 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=Kotlin의 async HTTP 요청과 코루틴 비교 sources=interview_questions2.md:409-438, interviews2.md:409-438 -->

> Source: `interview_questions2.md:409-438`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions2.md:409-438, interviews2.md:409-438`

##### Kotlin의 async HTTP 요청과 코루틴 비교

###### Async HTTP 요청 (예: OkHttp, Retrofit)

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

<!-- /curriculum-chunk -->

### 논블로킹 vs 비동기 차이

#### 원문: 논블로킹 vs 비동기 차이

<!-- curriculum-chunk: sha256=29a146ab57b3d26f1db3c13d61aa65dd932ef48b76355a34aff2986ea856d01a major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=논블로킹 vs 비동기 차이 sources=interview_questions.md:4470-4523, interviews.md:4470-4523 -->

> Source: `interview_questions.md:4470-4523`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions.md:4470-4523, interviews.md:4470-4523`

##### 논블로킹 vs 비동기 차이

논블로킹(Non-blocking)과 비동기(Asynchronous)는 비슷해 보이지만, 같은 개념이 아닙니다.
둘 다 "즉시 반환된다"는 공통점이 있지만, 결정적인 차이는 "작업이 어디에서 실행되는가?"입니다.

| 개념 | 설명 |
|------|---------------------------------------------------|
| 논블로킹 (Non-blocking) | 요청한 작업이 즉시 반환되지만, 요청한 스레드가 계속 작업을 직접 확인해야 함. |
| 비동기 (Asynchronous) | 요청한 작업을 별도의 스레드 또는 이벤트 루프가 처리하고, 요청한 스레드는 즉시 반환됨. |

📌 즉, 논블로킹은 요청한 스레드가 작업을 직접 확인해야 하지만, 비동기는 작업이 별도의 실행 컨텍스트에서 실행됨.

- 논블로킹(Non-blocking) 예제: 논블로킹은 "결과를 직접 확인해야 하는 방식"
    - 요청한 스레드가 직접 결과를 확인해야 함.
    - 데이터를 사용할 수 없는 경우, 반복해서 확인해야 함 (Polling 방식).

    ✅ 예제: 논블로킹 소켓 읽기

    ```java
    socket.configureBlocking(false);  // 논블로킹 모드 설정
    ByteBuffer buffer = ByteBuffer.allocate(1024);
    int bytesRead = socket.read(buffer);  // 데이터가 없으면 즉시 -1 반환

    if (bytesRead > 0) {
        System.out.println("데이터 수신 완료");
    } else {
        System.out.println("데이터 없음, 다른 작업 수행");
    }
    ```

    - `read()`가 데이터를 읽을 준비가 되어 있지 않으면 즉시 -1을 반환함.
    - 즉, 스레드는 직접 반복해서 확인해야 함 (Polling).

    ➡ 즉, 논블로킹은 작업을 직접 확인해야 하므로, 효율적이지 않을 수 있음.

- 비동기(Asynchronous) 예제: 비동기는 "작업을 별도의 스레드에서 실행하고, 요청한 스레드는 바로 반환되는 방식"
    - 작업을 별도의 스레드에서 실행하며, 요청한 스레드는 즉시 반환됨.
    - 작업이 완료되면 자동으로 결과를 전달받아 처리 가능.

    ✅ 예제: 비동기 소켓 읽기

    ```java
    CompletableFuture.supplyAsync(() -> {
        return readDataFromSocket();
    }).thenAccept(data -> {
        System.out.println("데이터 수신 완료: " + data);
    });
    ```

    - `readDataFromSocket()`이 백그라운드 스레드에서 실행됨.
    - 요청한 스레드는 즉시 반환되며, 작업이 완료되면 `thenAccept()`에서 결과를 처리.

    ➡ 즉, 비동기는 요청한 스레드가 직접 데이터를 확인할 필요 없이, 작업이 완료되면 자동으로 콜백을 실행함.

<!-- /curriculum-chunk -->

### 블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리

#### 원문: 블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리

<!-- curriculum-chunk: sha256=6f99ae3cef9ed51b575f4112e6885dd1c67496a61e106dc32ae241e52e03c9c3 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리 sources=interview_questions.md:4155-4469, interviews.md:4155-4469 -->

> Source: `interview_questions.md:4155-4469`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions.md:4155-4469, interviews.md:4155-4469`

##### 블로킹 vs 논블로킹, 동기 vs 비동기, 동시성 vs 병렬성 개념 정리

- 블로킹(Blocking):
    - 호출된 함수가 자신의 작업을 완료할 때까지 호출한 함수의 실행을 멈추게 합니다.
    - 즉, 제어권을 호출된 함수가 가지고 있으며, 작업이 끝날 때까지 호출한 함수는 대기 상태에 놓입니다.

- 논블로킹(Non-Blocking):
    - 호출된 함수가 즉시 제어권을 호출한 함수에게 반환하여, 호출한 함수가 다른 작업을 계속 수행할 수 있도록 합니다.
    - 호출된 함수는 자신의 작업 완료 여부와 상관없이 즉시 반환합니다.

- 동기(Synchronous):
    - 호출한 함수가 호출된 함수의 작업 완료 여부를 직접 확인하며, 작업이 완료될 때까지 기다립니다.
    - 즉, 작업의 완료 여부를 호출한 함수가 신경 쓰며, 순차적으로 작업을 처리합니다.

- 비동기(Asynchronous):
    - 호출한 함수가 호출된 함수의 작업 완료 여부를 신경 쓰지 않고, 호출된 함수가 작업 완료 후 결과를 전달하는 방식입니다.
    - 일반적으로 콜백 함수나 이벤트 리스너를 통해 결과를 전달받으며, 호출한 함수는 다른 작업을 계속 수행할 수 있습니다.

- 블로킹(Blocking) vs 논블로킹(Non-blocking): "작업을 요청한 스레드가 결과를 기다리는 방식"

    📌 블로킹과 논블로킹은 "작업을 요청한 스레드가 결과를 기다리는 방식"에 대한 개념입니다.

    | 개념 | 설명 |
    |------|----------------------------------------------------|
    | 블로킹(Blocking) | 요청한 작업이 완료될 때까지 스레드가 대기함. |
    | 논블로킹(Non-blocking) | 요청한 작업이 즉시 반환되며, 완료 여부와 관계없이 스레드는 계속 실행됨. |

    ✅ 블로킹 I/O 예제

    ```java
    // 블로킹 방식: read()가 완료될 때까지 현재 스레드는 멈춰 있음
    InputStream input = socket.getInputStream();
    int data = input.read();  // 데이터가 도착할 때까지 스레드가 대기
    ```

    ➡ 입력이 없으면 `read()`가 블로킹되어, 현재 스레드는 대기 상태가 됨.

    ✅ 논블로킹 I/O 예제

    ```java
    socket.configureBlocking(false);  // 논블로킹 모드 설정
    int data = input.read();  // 데이터가 없으면 -1 반환, 스레드는 대기하지 않음
    ```

    ➡ 데이터가 없으면 즉시 반환하므로, 스레드가 계속 실행될 수 있음.

    📌 즉, 블로킹은 "작업이 끝날 때까지 기다리는 방식"이고, 논블로킹은 "기다리지 않고 즉시 반환하는 방식"이다.

- 동기(Synchronous) vs 비동기(Asynchronous): "작업의 완료를 어떻게 처리하는가?"

    📌 동기와 비동기는 "작업의 완료를 어떻게 처리하는가?"에 대한 개념입니다.

    | 개념 | 설명 |
    |------|---------------------------------------------------|
    | 동기(Synchronous) | 요청한 작업이 완료될 때까지 요청한 스레드가 작업을 직접 처리함. |
    | 비동기(Asynchronous) | 요청한 작업을 별도의 스레드가 처리하고, 요청한 스레드는 즉시 반환됨. |

    ✅ 동기 예제

    ```java
    void syncMethod() {
        String result = fetchData();  // 데이터 요청 후 직접 기다림
        System.out.println("Result: " + result);
    }

    String fetchData() {
        return "Data";  // 데이터 반환 후 다음 코드 실행
    }
    ```

    ➡ `fetchData()`가 완료될 때까지 `syncMethod()`는 대기함.

    ✅ 비동기 예제

    ```java
    void asyncMethod() {
        CompletableFuture.supplyAsync(this::fetchData)
            .thenAccept(result -> System.out.println("Result: " + result));
        System.out.println("Request sent, doing other work...");
    }
    ```

    ➡ 데이터를 요청한 후 즉시 반환되고, 나중에 결과가 오면 처리됨.

    📌 즉, 동기는 "작업이 완료될 때까지 대기하는 방식"이고, 비동기는 "작업을 요청하고 결과가 오면 처리하는 방식"이다.

- 동시성(Concurrency) vs 병렬성(Parallelism)

    📌 동시성과 병렬성은 "여러 작업을 어떻게 실행하는가?"에 대한 개념입니다.

    | 개념 | 설명 |
    |------|--------------------------------------------------|
    | 동시성(Concurrency) | 여러 작업이 논리적으로 동시에 실행되지만, 실제로는 한 개의 CPU 코어에서 교차 실행됨. |
    | 병렬성(Parallelism) | 여러 작업이 물리적으로 동시에 실행됨 (멀티코어 CPU 필요). |

    ✅ 동시성 예제 (Concurrency)

    ```java
    ExecutorService executor = Executors.newFixedThreadPool(2);
    executor.submit(() -> { processTask(1); });
    executor.submit(() -> { processTask(2); });
    executor.shutdown();
    ```

    ➡ CPU가 하나의 코어에서 빠르게 스위칭하면서 두 작업을 동시에 실행하는 것처럼 보이게 만듦.

    ✅ 병렬성 예제 (Parallelism)

    ```java
    IntStream.range(1, 5).parallel().forEach(i -> processTask(i));
    ```

    ➡ 멀티코어에서 여러 개의 작업이 실제로 동시에 실행됨.

    📌 즉, 동시성은 "한 개의 CPU에서 여러 작업을 빠르게 전환하며 실행하는 방식"이고, 병렬성은 "여러 개의 CPU에서 실제로 동시에 실행하는 방식"이다.

- 블로킹, 논블로킹, 동기, 비동기의 조합

    📌 이제 개념들을 조합하여 네 가지 패턴을 정리해 보겠습니다.

    | 유형 | 블로킹/논블로킹 | 동기/비동기 | 예제 |
    |------|---------------|----------|------------------------------|
    | 블로킹 동기 | 블로킹 | 동기 | `Thread.sleep()`, `read()` |
    | 블로킹 비동기 | 블로킹 | 비동기 | `Future.get()` (결과 대기) |
    | 논블로킹 동기 | 논블로킹 | 동기 | `poll()`, `tryLock()` |
    | 논블로킹 비동기 | 논블로킹 | 비동기 | `CompletableFuture.supplyAsync()` |

    1. 동기 + 블로킹 (Synchronous + Blocking):
        - 요청한 스레드가 작업을 직접 처리 + 요청한 스레드가 결과를 기다림
        - 호출한 함수가 호출된 함수의 작업이 완료될 때까지 기다립니다.
        - 그동안 다른 작업을 수행하지 못합니다.
    2. 동기 + 논블로킹 (Synchronous + Non-Blocking):
        - 요청한 스레드가 작업을 직접 처리 + 요청한 스레드가 결과를 기다리지 않음
        - 호출된 함수가 즉시 제어권을 반환하여 호출한 함수가 다른 작업을 수행할 수 있습니다.
        - 하지만 호출한 함수는 주기적으로 호출된 함수의 작업 완료 여부를 확인해야 합니다.
    3. 비동기 + 블로킹 (Asynchronous + Blocking):
        - 요청한 작업을 별도의 스레드가 처리 + 요청한 스레드가 결과를 기다림
        - 호출한 함수가 호출된 함수의 작업 완료 여부를 신경 쓰지 않습니다.
        - 하지만 호출된 함수가 작업을 완료할 때까지 제어권을 반환하지 않아 호출한 함수의 실행이 멈춥니다.
    4. 비동기 + 논블로킹 (Asynchronous + Non-Blocking):
        - 요청한 작업을 별도의 스레드가 처리 + 요청한 스레드가 결과를 기다리지 않음
        - 호출된 함수가 즉시 제어권을 반환합니다.
        - 호출한 함수는 호출된 함수의 작업 완료 여부를 신경 쓰지 않습니다.
        - 호출된 함수는 작업 완료 후 콜백 등을 통해 결과를 전달합니다.

    ✅ 예제 코드

    ```java
    // 블로킹 동기 (Blocking Synchronous)
    String result = fetchData(); // 작업 완료될 때까지 대기

    // 블로킹 비동기 (Blocking Asynchronous)
    // 비동기 작업을 실행하지만, 최종적으로 결과를 가져올 때 블로킹하는 패턴입니다.
    // 즉, 작업을 실행하는 동안은 비동기적으로 수행되지만, 결과를 가져올 때는 블로킹되는 방식입니다.
    Future<String> future = executor.submit(this::fetchData);
    String result = future.get(); // get()을 호출하면 대기함

    // 논블로킹 동기 (Non-blocking Synchronous)
    Optional<String> result = queue.poll(); // 데이터가 없으면 즉시 반환

    // 논블로킹 비동기 (Non-blocking Asynchronous)
    CompletableFuture.supplyAsync(this::fetchData)
        .thenAccept(result -> System.out.println("Result: " + result));
    ```

    📌 즉, "동기/비동기"는 "작업 완료를 어떻게 처리할 것인가?", "블로킹/논블로킹"은 "작업 중에 대기할 것인가?"의 차이이다.

- 블로킹 비동기 추가 정리

    - 동기식 API가 필요하지만 내부적으로 비동기 실행이 필요한 경우

        예를 들어, 라이브러리나 프레임워크가 동기적인 API만 지원하지만, 내부적으로는 비동기로 실행할 필요가 있을 때 유용함.

        ```java
        // fetchData()는 동기 API지만, 내부적으로는 비동기 asyncDatabaseCall() 실행
        // → 동기 API를 유지하면서 비동기의 장점을 활용.
        public String fetchData() throws Exception {
            Future<String> future = asyncDatabaseCall();
            return future.get(); // 최종적으로 블로킹되지만, 내부적으로 비동기 실행됨
        }
        ```

    - 멀티스레딩 환경에서 조인(join) 필요: 멀티스레드 환경에서 하나의 스레드가 다른 스레드의 작업을 기다릴 때 유용함.

        ```java
        // 예제: 여러 개의 비동기 작업을 실행한 후, 모든 작업이 완료될 때까지 대기
        // 여러 개의 비동기 작업을 동시에 실행한 후, 모든 작업이 완료될 때까지 기다려야 하는 경우 유용함.
        import java.util.concurrent.*;

        public class ParallelAPICalls {
            public static void main(String[] args) throws Exception {
                ExecutorService executor = Executors.newFixedThreadPool(3);

                Future<String> future1 = executor.submit(() -> fetchDataFromAPI("Service1"));
                Future<String> future2 = executor.submit(() -> fetchDataFromAPI("Service2"));
                Future<String> future3 = executor.submit(() -> fetchDataFromAPI("Service3"));

                // 다른 작업 수행 가능
                System.out.println("다른 로직 실행 중...");

                // 모든 결과가 필요할 때 블로킹하여 최종 결과 취합
                String result1 = future1.get();
                String result2 = future2.get();
                String result3 = future3.get();

                System.out.println("모든 데이터 수집 완료: " + result1 + ", " + result2 + ", " + result3);
                executor.shutdown();
            }

            public static String fetchDataFromAPI(String service) throws InterruptedException {
                Thread.sleep(3000);  // 3초 지연 (API 호출 시뮬레이션)
                return service + " 응답 완료";
            }
        }
        ```

    - CPU 작업과 I/O 작업을 효율적으로 분리

        비동기적으로 실행한 후, 특정 시점에서만 블로킹하여 결과를 사용할 수도 있음.

        ```java
        public class CPUAndIOParallel {
            public static void main(String[] args) throws Exception {
                ExecutorService executor = Executors.newSingleThreadExecutor();

                Future<String> future = executor.submit(() -> {
                    return fetchDataFromDB();  // DB에서 데이터를 가져오는 비동기 작업
                });

                performCPUIntensiveTask();  // I/O 작업 중 CPU 연산 수행

                // I/O 작업이 끝났다면 블로킹하여 데이터 가져오기
                String data = future.get();
                System.out.println("최종 데이터: " + data);

                executor.shutdown();
            }

            public static String fetchDataFromDB() throws InterruptedException {
                Thread.sleep(5000);  // DB 쿼리 시뮬레이션 (5초 지연)
                return "DB 데이터 로드 완료";
            }

            public static void performCPUIntensiveTask() {
                System.out.println("CPU 연산 수행 중...");
                for (int i = 0; i < 1_000_000_000; i++) { /* 연산 */ }
                System.out.println("CPU 연산 완료");
            }
        }
        ```

    - UI 애플리케이션에서 네트워크 호출을 백그라운드에서 실행
        - UI 애플리케이션(JavaFX, Android, Swing 등)에서 네트워크 요청을 직접 실행하면 UI가 멈추는 문제 발생.
        - 따라서 네트워크 요청을 비동기적으로 실행한 후, UI 스레드에서 get()으로 결과를 받아 업데이트.
        - UI 스레드는 네트워크 요청을 직접 실행하지 않으므로, UI가 멈추지 않음.
        - 백그라운드에서 실행한 후, 결과가 필요할 때 get()을 호출하여 UI를 업데이트.

        ```java
        SwingWorker<String, Void> worker = new SwingWorker<>() {
            @Override
            protected String doInBackground() throws Exception {
                return fetchDataFromAPI();  // 네트워크 요청 (비동기 실행)
            }

            @Override
            protected void done() {
                try {
                    String result = get();  // 결과 가져오기 (UI 스레드에서 블로킹)
                    label.setText(result);  // UI 업데이트
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }
        };

        worker.execute();  // 비동기 실행
        ```

    - 타임아웃을 설정하여 일정 시간 동안만 결과를 기다림

        어떤 비동기 작업이 너무 오래 걸릴 경우, 특정 시간까지만 기다리고 이후에는 다른 처리를 수행해야 할 수도 있음.
        - 일정 시간 동안만 기다리고, 제한 시간을 초과하면 타임아웃 처리 가능.
        - 시스템이 멈추지 않고, 적절한 예외 처리를 수행할 수 있음.

        ```java
        ExecutorService executor = Executors.newSingleThreadExecutor();
        Future<String> future = executor.submit(() -> {
            Thread.sleep(5000);  // 5초 동안 실행
            return "완료된 작업";
        });

        try {
            String result = future.get(3, TimeUnit.SECONDS);  // 최대 3초 대기 후 결과 가져오기
            System.out.println("결과: " + result);
        } catch (TimeoutException e) {
            System.out.println("타임아웃 발생: 작업이 너무 오래 걸림");
        }

        executor.shutdown();
        ```

    블로킹 비동기의 단점:
    - 비동기의 장점(응답 속도 향상, 동시성 개선)을 완전히 살리지 못할 수 있음.
        - future.get()을 호출하면 결과를 받을 때까지 블로킹되므로, 결국 동기적인 방식과 비슷해짐.
        - 여러 개의 Future.get()을 순차적으로 호출하면 사실상 동기 실행과 다를 바 없음.

        ```java
        // 결과를 기다리지 않고, 비동기적으로 처리할 수 있음.
        CompletableFuture.supplyAsync(() -> {
            return fetchDataFromDB();
        }).thenAccept(result -> {
            System.out.println("비동기적으로 결과 처리: " + result);
        });
        ```

<!-- /curriculum-chunk -->

### 비동기 프로그래밍

#### 원문: 비동기 프로그래밍

<!-- curriculum-chunk: sha256=e61baa9c0afee796de80c74db875cd8798270502a2b58dc2f30e15a090b71552 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=비동기 프로그래밍 sources=interview_questions.md:3812-3873, interviews.md:3812-3873 -->

> Source: `interview_questions.md:3812-3873`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions.md:3812-3873, interviews.md:3812-3873`

##### 비동기 프로그래밍

비동기 프로그래밍은 I/O 작업을 효율적으로 처리하고 시스템 리소스를 최적화하기 위한 모델로, 동시성 구현에 도움을 줄 수 있습니다.
주요 기반 기술은 *이벤트 기반 아키텍처*와 *논 블로킹 I/O* 시스템입니다.
시스템 호출이나 커널의 도움을 최소화하는 방식으로 동작합니다.

- 논 블로킹 I/O (Non-blocking I/O)

    비동기 프로그래밍의 핵심 요소는 논 블로킹 I/O입니다.
    일반적인 블로킹 I/O는 I/O 작업이 완료될 때까지 프로그램의 실행이 멈춥니다.
    반면 논 블로킹 I/O는 I/O 작업이 즉시 리턴되어 프로그램의 실행이 계속되고, 실제 I/O 작업은 백그라운드에서 처리됩니다.
    논 블로킹 I/O가 가능하게 하는 기술들은 다음과 같습니다.

    - `select`와 `poll`:

        오래된 이벤트 대기 메커니즘으로, 파일 디스크립터의 상태 변화를 감지하여 비동기 I/O를 구현합니다.
        그러나 대규모 파일 디스크립터 관리에 비효율적입니다.

    - `epoll` (Linux):

        `select`와 `poll`의 성능 문제를 해결하기 위해 도입된 Linux의 고성능 이벤트 감지 시스템입니다.
        `epoll`은 파일 디스크립터의 변경을 감지하고, 비동기적으로 I/O 작업을 처리할 수 있게 합니다.

    - `kqueue` (BSD, macOS):

        BSD와 macOS에서 제공하는 유사한 이벤트 대기 메커니즘입니다.
        다수의 파일 디스크립터에 대한 상태 변화를 효율적으로 감지하여 비동기 I/O를 처리합니다.

- 이벤트 루프 (Event Loop)

    이벤트 루프는 비동기 프로그래밍의 중요한 구성 요소입니다.
    프로그램이 논 블로킹 방식으로 I/O 작업을 처리할 때, 이벤트 루프는 이벤트 발생을 감지하고, 적절한 콜백 함수나 핸들러를 실행합니다.

    - Node.js와 같은 비동기 런타임은 `libuv`라는 이벤트 루프 라이브러리를 기반으로 하여 논 블로킹 I/O와 타이머, 비동기 작업 스케줄링을 관리합니다.
    - JavaScript의 `Promise`, `async`/`await` 역시 이벤트 루프와 함께 작동하여, 비동기 작업의 결과를 받을 때까지 대기하지 않고 다른 작업을 계속 수행할 수 있도록 돕습니다.

- 시스템 호출 (System Calls)

    비동기 프로그래밍은 특정 시스템 호출을 통해 커널의 도움을 받기도 합니다.
    대표적인 비동기 I/O 관련 시스템 호출들은 다음과 같습니다:

    - `aio_*`:

        POSIX에서 제공하는 비동기 I/O(AIO) 시스템 호출입니다.
        `aio_read()`, `aio_write()` 같은 호출을 통해 비동기적으로 파일을 읽고 쓸 수 있습니다.

    - `io_uring`: 최신 Linux 커널에서 도입된 io_uring은 비동기 I/O를 위한 고성능 API입니다. 이는 기존의 `epoll`보다 훨씬 더 효율적이고 확장 가능한 비동기 I/O 시스템을 제공합니다.

비동기 프로그래밍을 쉽게 구현할 수 있도록 다양한 이벤트 드리븐 프레임워크들이 존재합니다.
이들은 비동기 I/O와 이벤트 루프를 추상화하여 개발자가 쉽게 사용할 수 있게 해줍니다.
- Python의 asyncio:
    Python에서 비동기 프로그래밍을 지원하는 모듈입니다.
    이벤트 루프와 `async/await` 구문을 통해 비동기 작업을 관리합니다.

- Go의 goroutine
    Go 언어는 자체 런타임에서 비동기 처리를 위한 goroutine을 제공합니다.
    이는 가볍고 효율적인 비동기 처리 방식입니다.

비동기 프로그래밍은 멀티스레딩과는 다르게 하나의 스레드에서 여러 작업을 동시적으로 처리할 수 있습니다.
전통적인 멀티스레딩 모델은 OS 스레드에 의존해 작업을 병렬로 처리하지만,
비동기 프로그래밍은 이벤트 기반 아키텍처 덕분에 스레드 간의 전환이 없이도 높은 동시성을 제공할 수 있습니다.

<!-- /curriculum-chunk -->

### 싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱

#### 원문: 싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱

<!-- curriculum-chunk: sha256=f3cf345881e3dda85aa07e8cf735cc91c2c2a4c2821b187523513ac97d5263c9 major=concurrency-async-io mid=Blocking, non-blocking, async 구분 sub=싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱 sources=interview_questions.md:4104-4154, interviews.md:4104-4154 -->

> Source: `interview_questions.md:4104-4154`
> Classification reason: async/non-blocking
> Duplicate source aliases: `interview_questions.md:4104-4154, interviews.md:4104-4154`

##### 싱글 코어 & 싱글 스레드 경우 커널에서의 비동기와 멀티플렉싱

싱글 코어 & 싱글 스레드 환경에서도 비동기 I/O가 가능한 이유는 CPU가 비동기적으로 I/O 작업을 요청한 후, 커널이 하드웨어 인터럽트 및 비동기 대기 메커니즘을 사용해 I/O 작업을 처리하기 때문입니다.
애플리케이션은 I/O 작업을 요청한 후 다른 작업을 처리할 수 있으며, I/O 작업 완료 여부는 커널이 적절히 관리하고, 애플리케이션은 이를 비동기적으로 처리합니다.

비동기 I/O의 처리 과정은 다음과 같습니다:
1. 비동기 I/O 요청 전달 (애플리케이션 → 커널):
    - 애플리케이션이 `aio_*` 함수를 호출하면, 해당 I/O 요청은 커널로 전달됩니다.
    - 커널은 해당 요청을 작업 큐(Work Queue)에 추가하고, 바로 애플리케이션에 제어를 반환합니다.
    - 이때 I/O 요청은 비동기로 처리되므로, 애플리케이션은 블로킹되지 않고 다른 작업을 수행할 수 있습니다.

2. 커널의 I/O 처리 (커널 내에서):
    - 커널은 I/O 장치(디스크, 네트워크 등)에 직접 접근하여 데이터를 읽거나 쓰는 작업을 수행합니다.
    - 커널은 I/O 작업이 CPU 자원을 많이 요구하지 않기 때문에(디스크 접근, 네트워크 전송 등의 작업은 주로 I/O 장치에서 이루어지므로), CPU는 다른 작업을 계속 수행할 수 있습니다.
    - 커널은 인터럽트(Interrupt) 기반으로 I/O 작업을 관리합니다. I/O 작업이 완료되면 하드웨어 인터럽트가 발생하여 커널에 작업 완료를 알립니다.

3. I/O 완료 시점까지의 애플리케이션 상태:
    - 애플리케이션은 커널로 I/O 요청을 보낸 후, 다른 작업을 계속 수행하거나 대기 상태로 전환될 수 있습니다. 만약 애플리케이션이 다른 작업을 할 것이 없을 경우, 대기 상태로 들어가 이벤트 루프(Event Loop)를 통해 I/O 완료 여부를 체크할 수 있습니다.
    - 이때 애플리케이션이 I/O 요청의 완료 여부를 확인하는 방법은 `aio_suspend` 또는 `epoll`, `select`와 같은 이벤트 기반 I/O 대기 메커니즘을 사용합니다.

4. 커널에서 I/O 완료 알림:
   - 커널이 I/O 작업을 완료하면, 하드웨어 인터럽트를 통해 CPU에 완료 상태를 알립니다.
   - 커널은 해당 I/O 작업이 완료되었음을 애플리케이션에 알릴 준비를 합니다.
   - 애플리케이션이 대기 상태일 경우, 커널은 작업 큐에서 대기 중인 프로세스나 스레드를 깨워 I/O가 완료되었음을 알립니다.

5. 싱글 코어 & 싱글 스레드 환경에서의 비동기 처리:
   - 코어 하나와 스레드 하나만 있는 경우, 애플리케이션이 I/O 요청을 보내면 CPU는 애플리케이션의 다른 코드를 실행하거나 대기 상태로 전환합니다.
   - I/O 작업은 CPU가 아닌 I/O 장치에서 주로 처리되므로, CPU는 다른 작업을 처리할 수 있습니다. CPU는 주기적으로 인터럽트를 통해 I/O 완료 여부를 확인하거나, 애플리케이션이 명시적으로 `aio_suspend`나 `select`를 호출하여 I/O 완료 여부를 확인하게 됩니다.

6. I/O 완료 후 애플리케이션 처리:
   - I/O 작업이 완료되면, 커널은 애플리케이션에 I/O 완료 알림을 전달합니다.
   - 그 후, 애플리케이션은 I/O 작업 결과를 처리하고, 해당 데이터를 기반으로 추가 작업을 수행할 수 있습니다.
   - 이 전체 과정에서 CPU는 계속해서 애플리케이션 코드와 커널 간의 스위칭을 관리하며, 필요한 작업이 완료될 때마다 애플리케이션이 적절히 반응할 수 있도록 합니다.

비록 싱글 코어에서 하나의 스레드만 존재하더라도, 비동기 I/O가 가능한 이유는 다음과 같습니다:

- 비동기 I/O 작업은 CPU를 점유하지 않음:

    비동기 I/O는 하드웨어 장치(디스크, 네트워크 인터페이스)에서 이루어지기 때문에 CPU가 관여하지 않습니다.
    CPU는 I/O 작업의 완료 여부를 추적할 뿐, 실제로는 I/O 장치가 작업을 처리합니다.

- 커널의 인터럽트 처리:

    커널은 I/O 작업이 완료되면 하드웨어 인터럽트를 통해 작업이 완료되었음을 알립니다.
    이 인터럽트는 CPU가 애플리케이션 코드와 다른 작업을 수행하고 있는 중에도 처리됩니다.

- 비동기 요청 후 애플리케이션의 대기 방식:

    애플리케이션이 I/O 완료를 기다리는 동안 CPU는 다른 작업을 하거나 대기 상태로 들어갈 수 있습니다.
    필요한 경우 `aio_suspend`, `select`, `epoll` 등의 비동기 대기 메커니즘을 통해 비동기적으로 완료를 확인할 수 있습니다.

<!-- /curriculum-chunk -->

## Coroutine과 협력형 실행

### Kotlin Coroutine

#### 원문: Kotlin Coroutine

<!-- curriculum-chunk: sha256=83b85432cf78a549a38f93d84e6930c9985f9e0630c6075330de38a2c337c8d9 major=concurrency-async-io mid=Coroutine과 협력형 실행 sub=Kotlin Coroutine sources=interview_questions.md:3747-3770, interviews.md:3747-3770 -->

> Source: `interview_questions.md:3747-3770`
> Classification reason: coroutine
> Duplicate source aliases: `interview_questions.md:3747-3770, interviews.md:3747-3770`

##### Kotlin Coroutine

Kotlin의 Coroutine은 협업적 멀티 스레딩(Cooperative Multithreading)을 기반으로 합니다.

> 협업적 멀티태스킹 (Cooperative Multitasking)
>
> 협업적 멀티태스킹 모델에서 작업은 스스로 적절한 지점에서 중단하고, 다른 작업이 실행되도록 양보합니다.
> 이 과정에서 커널의 스케줄러 개입 없이도 효율적인 비동기 동작을 수행할 수 있습니다.
>
> Kotlin Coroutines은 협업적 멀티태스킹의 대표적인 예입니다.
> 코루틴은 명시적으로 일시 중단(`suspend`)되고, 필요한 시점에 다시 실행되어 비동기 작업을 처리합니다.

협업적 멀티 스레딩은 스레드가 명시적으로 스케줄링 제어를 양보하는 방식입니다.
이를 통해 제어권을 명시적으로 반환하여 다른 Coroutine이 실행되도록 합니다.
즉, Coroutine은 자신이 직접 제어하여 언제 일시 중단하고 언제 재개할지 결정합니다.

이는 코루틴을 명시적으로 일시 중단하는 `suspend` 키워드나 `yield`와 같은 함수를 통해 이루어집니다.

비교적 적은 리소스를 사용하며, *여러 코루틴이 하나의 스레드에서 순차적으로 실행*됩니다.
Kotlin의 Dispatchers를 통해 다양한 스레드 풀에서 코루틴이 실행되도록 제어할 수 있지만, 기본적으로는 협업적 방식으로 실행됩니다.

- 코루틴은 하나의 스레드에서 여러 코루틴을 관리할 수 있지만, 협업적 모델이므로 코루틴은 강제로 멈춰지지 않으며, 명시적으로 멈춰져야 합니다.
- Kotlin 코루틴은 경량 쓰레드로 작동하며, JVM의 `Thread`와 비교했을 때 훨씬 적은 메모리 오버헤드로 수천 개의 코루틴을 동시에 실행할 수 있습니다.

<!-- /curriculum-chunk -->

### 코루틴 제어의 양보

#### 원문: 코루틴 제어의 양보

<!-- curriculum-chunk: sha256=ffbe6742cabd9b009c81d5584ca5541c668bd6e0aac7616a146862a6bc079ae2 major=concurrency-async-io mid=Coroutine과 협력형 실행 sub=코루틴 제어의 양보 sources=interview_questions2.md:357-408, interviews2.md:357-408 -->

> Source: `interview_questions2.md:357-408`
> Classification reason: coroutine
> Duplicate source aliases: `interview_questions2.md:357-408, interviews2.md:357-408`

##### 코루틴 제어의 양보

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

<!-- /curriculum-chunk -->

### 코루틴을 Cooperative Multitasking 이라고 하는 이유?

#### 원문: 코루틴을 Cooperative Multitasking 이라고 하는 이유?

<!-- curriculum-chunk: sha256=21d3928f49c0b7c1820fb38a5130d9de92f36c9f5d35820a269fe916564f3242 major=concurrency-async-io mid=Coroutine과 협력형 실행 sub=코루틴을 Cooperative Multitasking 이라고 하는 이유? sources=interview_questions2.md:250-356, interviews2.md:250-356 -->

> Source: `interview_questions2.md:250-356`
> Classification reason: coroutine
> Duplicate source aliases: `interview_questions2.md:250-356, interviews2.md:250-356`

##### 코루틴을 Cooperative Multitasking 이라고 하는 이유?

###### Kotlin Coroutine

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

###### 일반적인 Java의 멀티스레딩

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

<!-- /curriculum-chunk -->

## Event loop와 네트워크 런타임

### 4. Netty 관련 질문

#### 원문: 4. Netty 관련 질문

<!-- curriculum-chunk: sha256=f594edbc04d54389d2f0ff4fbb940d7d898d04a74c7e888a06a818910ad5b7c4 major=concurrency-async-io mid=Event loop와 네트워크 런타임 sub=4. Netty 관련 질문 sources=interview_questions3.md:189-212 -->

> Source: `interview_questions3.md:189-212`
> Classification reason: event loop

##### 4. Netty 관련 질문

###### 질문 9. Netty는 어떤 원리로 고성능 네트워크를 구현하나요?

###### **답변**

1. Netty는 **NIO(Non-blocking I/O)** 기반의 **이벤트 루프(EventLoop)** 모델을 채택하고 있습니다. 이는 기존의 Thread-per-Connection 방식을 지양하고, 소수의 쓰레드가 다수의 소켓 채널을 관리하게 만듭니다.
2. 예를 들어, 이벤트 루프는 Selector를 사용해 I/O 이벤트(읽기, 쓰기, 연결 요청 등)를 감지하고, 그 이벤트가 발생한 채널에 대해 처리 로직을 호출합니다.
3. Netty에서는 **ChannelPipeline** 개념을 통해, 인바운드 및 아웃바운드 데이터를 단계별 핸들러로 넘깁니다. 인코더, 디코더, 비즈니스 로직 핸들러를 각각 체인으로 구성할 수 있어, 유지보수와 확장성이 뛰어납니다.
4. 또한 Netty는 OS별로 지원되는 **Native Transport**(예: Epoll on Linux, KQueue on macOS)와 **Zero-Copy** 기법 등을 통해 커널/유저 간 데이터 복사를 최소화하여 성능을 끌어올립니다.

---

###### 질문 10. Netty 파이프라인에서 데이터가 인바운드/아웃바운드로 흐를 때, 예외가 발생하면 어떻게 처리되나요?

###### **답변**

1. Netty는 파이프라인을 **ChannelHandler**들의 체인으로 표현합니다. 데이터가 들어오는 방향(인바운드)과 나가는 방향(아웃바운드) 이벤트가 이 체인 위를 흐릅니다.
2. 특정 핸들러에서 예외가 발생하면, Netty는 그 예외를 **exceptionCaught** 메서드로 전파하여 처리하도록 합니다.
3. exceptionCaught가 오버라이딩되지 않았거나, 처리되지 않고 위로 전파되면 파이프라인 상위 레벨(또는 마지막 단계)까지 예외가 전달됩니다. 최종적으로도 처리하지 않으면 채널을 닫을 수도 있습니다.
4. 개발자는 예외 상황에 따라 로그를 남기고 연결을 정상 종료할지, 재시도할지, 특별한 예외 처리를 할지 결정해야 합니다.

---

<!-- /curriculum-chunk -->

### Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현

#### 원문: Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현

<!-- curriculum-chunk: sha256=649cf073128bff3d4874f03aabb397d6d79635628df319cc500895d3d14f1408 major=concurrency-async-io mid=Event loop와 네트워크 런타임 sub=Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현 sources=interview_questions.md:10855-10992, interviews.md:10803-10940 -->

> Source: `interview_questions.md:10855-10992`
> Classification reason: event loop
> Duplicate source aliases: `interview_questions.md:10855-10992, interviews.md:10803-10940`

##### Netty의 이벤트 루프(Event Loop)와 epoll 기반 논블로킹 I/O 구현

- `epoll`, `kqueue` 등의 커널 기반 멀티플렉싱 기법 이해

    `epoll`은 Linux 커널에서 제공하는 I/O 멀티플렉싱(multiplexing) 기법으로,
    여러 개의 파일 디스크립터(File Descriptor, FD)의 변화를 감지하고,
    이벤트가 발생한 파일 디스크립터만 애플리케이션에 전달하는 방식입니다.
    이벤트가 발생한 파일 디스크립터만 반환하므로, O(1) 시간 복잡도로 동작합니다.

    ➡ 즉, `epoll`은 애플리케이션이 많은 소켓을 효율적으로 감시할 수 있도록 하는 시스템 호출입니다.

    ✅ epoll의 기본 동작 흐름:
    1. `epoll_create()` → `epoll` 인스턴스 생성 (`epoll_fd`)
    2. `epoll_ctl()` → 특정 소켓(파일 디스크립터)을 감시 대상으로 등록
    3. `epoll_wait()` → 감시 중인 소켓에서 이벤트가 발생할 때까지 대기 (Blocking). 이벤트가 발생하지 않으면 CPU는 대기 상태에 들어가며, 불필요한 리소스 사용을 줄입니다.
    4. 이벤트 발생 시 → 커널이 해당 이벤트를 애플리케이션에 전달
    5. 이벤트 핸들링 후 → 다시 epoll에 등록하여 감시 지속

    ```c
    // epoll을 이용한 I/O 멀티플렉싱
    int epoll_fd = epoll_create1(0);
    struct epoll_event event;
    event.events = EPOLLIN;
    event.data.fd = server_socket;
    epoll_ctl(epoll_fd, EPOLL_CTL_ADD, server_socket, &event);

    while (1) {
        struct epoll_event events[MAX_EVENTS];
        int num_events = epoll_wait(epoll_fd, events, MAX_EVENTS, -1);

        for (int i = 0; i < num_events; i++) {
            if (events[i].events & EPOLLIN) {
                // 읽기 가능한 소켓 처리
            }
        }
    }
    ```

    - `epoll_wait()`을 호출하면, 이벤트가 발생한 파일 디스크립터만 반환.
    - 즉, 매번 모든 소켓을 확인하는 것이 아니라, 변경된 소켓만 확인하므로 O(1) 시간 복잡도로 동작.
    - Blocking 방식과 Non-blocking 방식 지원:
        - `epoll_wait()`을 블로킹으로 호출하면 이벤트가 발생할 때까지 대기
        - 논블로킹 모드에서는 즉시 반환

- Netty가 커널의 도움 없이 select/epoll 같은 기능을 제공하는가?
    Netty는 Java NIO의 `Selector` API를 사용하여 기본적으로 OS 커널에서 제공하는 I/O 멀티플렉싱(`epoll`, `kqueue`)을 추상화한 방식으로 구현합니다.

    Java NIO의 Selector는 플랫폼에 따라 다르게 구현됩니다:
    - Linux: `epoll` 기반 (`EPollSelectorProvider`)
    - macOS: `kqueue` 기반 (`KQueueSelectorProvider`)
    - Windows: `select` 기반

    Linux 환경에서는 Netty가 `netty-transport-native-epoll` 모듈을 통해 `epoll` 기반의 네이티브 I/O를 지원하며, 이를 통해 성능을 최적화할 수 있습니다

    Netty는 Java의 NIO(Non-blocking I/O) 기반의 이벤트 루프 모델을 사용하여,
    OS 커널의 I/O 멀티플렉싱(`epoll`, `kqueue`)을 Java의 `Selector` API로 추상화하여 관리합니다.

    ➡ 즉, Netty는 커널에서 제공하는 `epoll`, `kqueue` 등의 멀티플렉싱 기능을 `Selector` API로 감싸서 사용합니다.

- Java NIO의 `Selector`는 무엇인가?

    - Java의 `Selector`는 OS 커널이 제공하는 I/O 멀티플렉싱 기능(`epoll`, `kqueue`)을 추상화한 API입니다.
    - Netty는 `Selector`를 활용하여 여러 개의 소켓을 동시에 감시할 수 있습니다.

    ```java
    Selector selector = Selector.open();
    ServerSocketChannel serverSocket = ServerSocketChannel.open();
    serverSocket.configureBlocking(false);
    serverSocket.register(selector, SelectionKey.OP_ACCEPT);

    while (true) {
        selector.select();  // 이벤트가 발생할 때까지 대기
        Set<SelectionKey> selectedKeys = selector.selectedKeys();
        for (SelectionKey key : selectedKeys) {
            if (key.isAcceptable()) {
                SocketChannel client = serverSocket.accept();
                client.configureBlocking(false);
                client.register(selector, SelectionKey.OP_READ);
            } else if (key.isReadable()) {
                ByteBuffer buffer = ByteBuffer.allocate(1024);
                ((SocketChannel) key.channel()).read(buffer);
            }
        }
    }
    ```

    ➡ Netty는 내부적으로 `Selector`를 사용하여 `epoll`을 추상화한 이벤트 루프를 구현합니다.

Netty의 이벤트 루프는 멀티스레드 기반으로 실행되며, `NioEventLoopGroup`을 사용하여 스레드를 관리합니다.

| 구성 요소 | 설명 |
|------------|----------|
| Boss Group | 클라이언트의 연결을 처리하는 이벤트 루프 |
| Worker Group | 실제 데이터 송수신을 담당하는 이벤트 루프 |

- Netty의 이벤트 루프 흐름

    1. `BossGroup`이 클라이언트 연결을 감지하고, `WorkerGroup`으로 전달
    2. WorkerGroup이 `Selector`를 통해 이벤트를 감시
    3. 이벤트가 발생하면 해당 이벤트를 처리하는 핸들러 실행

    ```java
    EventLoopGroup bossGroup = new NioEventLoopGroup(1);
    EventLoopGroup workerGroup = new NioEventLoopGroup();
    ServerBootstrap bootstrap = new ServerBootstrap();
    bootstrap.group(bossGroup, workerGroup)
            .channel(NioServerSocketChannel.class)
            .childHandler(new ChannelInitializer<SocketChannel>() {
                @Override
                protected void initChannel(SocketChannel ch) {
                    ch.pipeline().addLast(new MyChannelHandler());
                }
            });
    bootstrap.bind(8080).sync();
    ```

    ➡ Netty의 이벤트 루프는 내부적으로 `Selector`를 사용하여 논블로킹 방식으로 동작함.

- CPU가 epoll을 어떻게 실행하는가?
    - CPU는 주기적으로 epoll을 실행하는 것이 아니라, 이벤트 기반으로 epoll을 트리거함.
    - `epoll_wait()`을 호출하면, 커널은 이벤트가 발생하기 전까지 CPU를 사용하지 않음.
    - 즉, CPU가 지속적으로 `epoll`을 실행하는 것이 아니라 "이벤트 발생 시 콜백 방식으로 처리".

    ➡ 결과적으로, CPU는 epoll이 동작할 때만 개입하며, 불필요한 CPU 사용을 줄임.

- Netty의 이벤트 루프는 어떻게 변경 사항을 감지하는가?

    ✅ Netty는 `Selector`를 사용하여 OS 커널이 제공하는 이벤트를 감지하고, 이를 비동기적으로 처리합니다.
    ✅ 즉, Netty의 이벤트 루프(Thread)가 하나 상주하면서 계속 변경 사항을 감시하는 것이 아니라, "이벤트가 발생했을 때만 처리"하는 구조입니다.

    Netty가 변경 사항을 감지하는 방식:
    1. `Selector.select()` 또는 `epoll_wait()`을 호출하여 변경 사항 감지
    2. 이벤트가 발생한 소켓만 감지하여 처리 (불필요한 스캔 없음)
    3. 데이터가 준비되면, 이벤트 루프에서 등록된 핸들러 실행
    4. 처리 후, 다시 `Selector.select()`를 호출하여 다음 이벤트를 기다림

    ➡ 즉, Netty는 "변경된 것만 감지하는 방식"을 채택하여 불필요한 CPU 사용을 줄이고, 효율적으로 이벤트를 관리함.

<!-- /curriculum-chunk -->

## Goroutine과 런타임 스케줄링

### Go의 Goroutine은 M 대 N 모델 멀티스레딩

#### 원문: Go의 Goroutine은 M 대 N 모델 멀티스레딩

<!-- curriculum-chunk: sha256=c6902e7e8ee1488f24a206675e072db5143558634bb85ac1846f259b15ca0fa2 major=concurrency-async-io mid=Goroutine과 런타임 스케줄링 sub=Go의 Goroutine은 M 대 N 모델 멀티스레딩 sources=interview_questions.md:3735-3746, interviews.md:3735-3746 -->

> Source: `interview_questions.md:3735-3746`
> Classification reason: goroutine
> Duplicate source aliases: `interview_questions.md:3735-3746, interviews.md:3735-3746`

##### Go의 Goroutine은 M 대 N 모델 멀티스레딩

Go 언어에서 goroutine은 M:N 스레드 모델을 기반으로 동작합니다.
이 모델에서 수많은 goroutine이 소수의 OS-level 스레드에 매핑됩니다.
M:N 모델이란 많은 사용자 레벨 스레드가 소수의 커널 스레드 위에서 실행됨을 의미합니다.

구체적으로 Go는 내부에서 `G` (goroutine), `M` (machine, OS-level 스레드), 그리고 `P` (processor, 스케줄러 역할)를 사용하여 goroutine을 관리합니다.
Goroutine이 블록될 때, Go 런타임은 해당 goroutine을 중지하고 다른 OS 스레드에 있는 다른 goroutine을 실행시키면서 동시성을 유지합니다.

- Go는 이를 프리엠티브 멀티태스킹 방식을 사용하여 구현합니다. 즉, goroutine은 Go 런타임에 의해 강제로 멈춰지고 다른 goroutine에게 CPU 자원을 할당할 수 있습니다.
- Goroutine은 OS 스레드와 독립적으로 Go 런타임에 의해 관리되며, 스케줄링이 이루어집니다.

<!-- /curriculum-chunk -->

## Java monitor와 wait/notify

### 1. 자바에서의 쓰레드와 모니터 개념

#### 원문: 1. 자바에서의 쓰레드와 모니터 개념

<!-- curriculum-chunk: sha256=b3d7caa2f7f8d0678fa28e3765b52b5efb83e3f23ca512a3a0d2529c0f3a1633 major=concurrency-async-io mid=Java monitor와 wait/notify sub=1. 자바에서의 쓰레드와 모니터 개념 sources=interview_questions3.md:396-405 -->

> Source: `interview_questions3.md:396-405`
> Classification reason: monitor/wait-notify

##### 1. 자바에서의 쓰레드와 모니터 개념

1. 모든 동물은 생을 마친다(일반 원칙).
2. 자바의 `synchronized` 블록 또는 메서드는 “모니터(Monitor)”를 기반으로 동작한다(특정 원칙).
3. 따라서 `synchronized(obj)` 구문을 만나면, 쓰레드는 `obj`라는 객체의 모니터를 **획득**해야만 블록 내부 코드를 실행할 수 있다(결론).

먼저, “**모니터**”란, 자바에서 **동기화(synchronization)**를 지원하기 위해 객체마다 존재하는 구조체 혹은 **잠금(락) 기법**을 의미합니다. `synchronized (obj)`라고 하면, 해당 obj에 대한 **모니터 잠금**을 쓰레드가 얻어야만(획득해야만) 코드 실행이 가능합니다.

---

<!-- /curriculum-chunk -->

### 10. 핵심 요약 및 결론

#### 원문: 10. 핵심 요약 및 결론

<!-- curriculum-chunk: sha256=26965e7485369ce237a93404fac05c64e47f232e3eb2dd542d7c14641823122d major=concurrency-async-io mid=Java monitor와 wait/notify sub=10. 핵심 요약 및 결론 sources=interview_questions3.md:535-550 -->

> Source: `interview_questions3.md:535-550`
> Classification reason: monitor/wait-notify

##### 10. 핵심 요약 및 결론

1. **Java 스레드 관리**에서 `synchronized`는 객체 모니터를 기반으로 임계영역 접근을 직렬화한다. 이때 모니터가 이미 점유 중이면 새로 진입하려는 쓰레드는 BLOCKED가 된다.
2. **wait()**는 모니터를 임의로 반납하고 객체의 wait set에 들어가게 하여 쓰레드를 WAITING 상태로 만든다. 이때 **notify()** 혹은 **notifyAll()**이 호출되기 전까지는 깨어나지 못한다.
3. notify가 없으면 해당 쓰레드는 영구 대기할 수 있으므로, 반드시 올바른 로직으로 wait/notify 쌍을 구성해야 한다.
4. **BLOCKED** 상태는 모니터를 아직 얻지 못해 문 앞에서 기다리는 상태이고, **WAITING** 상태는 이미 모니터를 얻은 뒤 wait()를 호출해 놓은 상태다. 깨어날 때도 모니터를 다시 얻어야 하므로, 일시적으로 BLOCKED로 돌아갈 수 있다.
5. JVM 레벨에서는 `monitorenter`/`monitorexit`, wait set, 락 최적화 등 다양한 내부 메커니즘으로 이 과정을 관리한다.

이처럼 자바의 **동기화 모델**은 모니터 잠금과 wait/notify를 결합하여, 스레드 간 통신(Condition Wait)과 임계영역 보호(Mutex Lock)를 동시에 지원합니다. 이것이 자바가 멀티쓰레드 프로그래밍을 강력하고 안전하게 제공하는 기저가 됩니다.

> **결론**: synchronized, wait, notify를 올바르게 사용하기 위해서는
> 1) 모니터 잠금의 획득/해제 시점
> 2) BLOCKED vs WAITING 상태 구분
> 3) notify/notifyAll을 통한 대기 해제 원리
> 등을 숙지해야 하며, 이는 JVM의 모니터 구조와 자바 스레드 상태 모델을 이해해야 명확히 파악할 수 있습니다.

<!-- /curriculum-chunk -->

### 3. synchronized 블록과 모니터 잠금

#### 원문: 3. synchronized 블록과 모니터 잠금

<!-- curriculum-chunk: sha256=e8e1b25a1a710a3be97449a3edf42a13868ca90e6cd8f8386e686b99988d4088 major=concurrency-async-io mid=Java monitor와 wait/notify sub=3. synchronized 블록과 모니터 잠금 sources=interview_questions3.md:421-440 -->

> Source: `interview_questions3.md:421-440`
> Classification reason: monitor/wait-notify

##### 3. synchronized 블록과 모니터 잠금

###### 3.1 진입 과정

1. 만약 A 쓰레드가 `synchronized(LockObject)` 구문을 실행하려고 시도한다(원인).
2. JVM은 **LockObject**라는 객체의 모니터를 확인한다(작동).
3. 객체 모니터가 사용 중(이미 다른 쓰레드가 갖고 있음)이면, A 쓰레드는 **BLOCKED** 상태로 전환되어 대기한다(결론 1).
4. 만약 모니터가 비어있다면, A 쓰레드는 **LockObject 모니터를 획득**하고 **RUNNABLE** 상태로 블록 내부 코드를 실행한다(결론 2).

즉, `BLOCKED` 상태는 “**모니터 잠금을 얻지 못해서** 대기”인 상태를 의미합니다. `synchronized(obj)`가 끝난 뒤(모니터가 release된 뒤)에야 다시 재시도할 수 있습니다.

###### 3.2 탈출 과정

1. A 쓰레드가 `synchronized(LockObject)` 블록의 끝에 도달하거나, 예외가 발생해 블록을 빠져나가면,
2. 잠금이 풀리고(= 모니터 반납), 대기 중인 다른 쓰레드(만약 있다면)가 모니터를 획득할 기회를 얻는다.

이로써 단일 객체 모니터를 통한 직렬화(One-at-a-time)가 달성됩니다.

---

<!-- /curriculum-chunk -->

### 4. wait()와 notify() 메서드의 메커니즘

#### 원문: 4. wait()와 notify() 메서드의 메커니즘

<!-- curriculum-chunk: sha256=b9855554ec7f8548ecdfec4dec7ea48dba53f891accde28c7f65db90496a97f2 major=concurrency-async-io mid=Java monitor와 wait/notify sub=4. wait()와 notify() 메서드의 메커니즘 sources=interview_questions3.md:441-462 -->

> Source: `interview_questions3.md:441-462`
> Classification reason: monitor/wait-notify

##### 4. wait()와 notify() 메서드의 메커니즘

###### 4.1 wait()가 하는 일

1. 자바에서 `obj.wait()`는 “**이 객체(obj)에 대한 모니터**”를 사용 중인 쓰레드가 임시로 모니터 **점유권을 포기**하고, 해당 객체의 **wait set**에 들어가 기다리게 하는 메서드이다(일반 원칙).
2. 즉, 쓰레드가 synchronized(obj) 블록 내부에서 `obj.wait()`를 호출하면, 자발적으로 모니터를 놓고(wait/release), JVM 내부의 **wait set**에 들어가 **WAITING** 상태로 전환된다(작동).
3. 그 결과, 모니터가 비워지므로 다른 쓰레드가 이 모니터를 획득할 수 있게 된다(결론).

`wait()`가 호출되어 WAITING 상태가 된 쓰레드는, **notify() / notifyAll()** 등이 호출되기 전까지는 깨어나지 않는다. 깨어난 뒤에는 모니터를 다시 획득해야만(`BLOCKED` 과정을 거칠 수 있음) 원래 코드의 이어지는 부분(`doSomething()`)을 실행한다.

###### 4.2 notify()와 notifyAll()

1. `obj.notify()`는 “**obj의 wait set**에서 대기 중인 쓰레드들 중 한 쓰레드**”를 깨우는(awakening) 역할이다(원칙).
2. 깨워진 쓰레드는 **모니터를 다시 획득하려고 시도**하지만, 그 순간 모니터가 이미 다른 쓰레드에 의해 사용 중이면, `BLOCKED` 상태로 잠시 대기해야 한다(작동).
3. 만약 모니터가 비어있다면, 바로 획득한 뒤 **RUNNABLE** 상태가 되어 `wait()` 호출 직후 코드를 이어서 수행한다(결론).

차이점:
- `notify()`는 한 개 쓰레드만 깨운다.
- `notifyAll()`는 wait set의 **모든 쓰레드**를 깨우지만, 결국 모니터는 한 쓰레드만 먼저 획득할 수 있으므로, 나머지 쓰레드는 `BLOCKED` → `RUNNABLE` 순으로 경합한다.

---

<!-- /curriculum-chunk -->

### 5. wait set의 작동 원리

#### 원문: 5. wait set의 작동 원리

<!-- curriculum-chunk: sha256=dc98dc5c9e117401b065ca46145124d4b24f30c8034b0ea23700a1db9a230978 major=concurrency-async-io mid=Java monitor와 wait/notify sub=5. wait set의 작동 원리 sources=interview_questions3.md:463-472 -->

> Source: `interview_questions3.md:463-472`
> Classification reason: monitor/wait-notify

##### 5. wait set의 작동 원리

1. 어떤 객체 LockObject에 대해 `LockObject.wait()`가 호출되면, 해당 쓰레드는 “LockObject의 **wait set**”에 추가된다(원칙).
2. 그동안 이 쓰레드는 모니터를 포기했으므로, **WAITING** 상태가 된다(작동).
3. 다른 쓰레드가 `LockObject.notify()`나 `LockObject.notifyAll()`을 호출하면, wait set에 있던 쓰레드(들)이 깨운 신호를 받고, 모니터 재획득을 시도한다(결론).

만약 **notify**나 **notifyAll**이 절대 호출되지 않는다면(즉, 락 오브젝트에 대한 신호가 없으면), **해당 쓰레드는 영원히 wait set에서 WAITING 상태**로 남을 수 있습니다. 따라서 wait/notify 구문은 반드시 한 쌍으로 설계해야 합니다.

---

<!-- /curriculum-chunk -->

### 6. 쓰레드 상태: BLOCKED vs WAITING

#### 원문: 6. 쓰레드 상태: BLOCKED vs WAITING

<!-- curriculum-chunk: sha256=510429daa623c32b746152808abece85f146922230224c7349ebf7f93a811ab3 major=concurrency-async-io mid=Java monitor와 wait/notify sub=6. 쓰레드 상태: BLOCKED vs WAITING sources=interview_questions3.md:473-490 -->

> Source: `interview_questions3.md:473-490`
> Classification reason: monitor/wait-notify

##### 6. 쓰레드 상태: BLOCKED vs WAITING

###### 6.1 BLOCKED 상태

1. A 쓰레드가 “**아직 모니터를 획득하지 못했음**”에도 불구하고, `synchronized(obj)` 구문에 진입하려 시도하면, 모니터가 다른 쓰레드에 점유된 경우 BLOCKED 상태가 된다(원칙).
2. 즉, **mutex 락**을 얻지 못해서 대기 중인 상태라고도 할 수 있다(작동).
3. 다른 쓰레드가 모니터를 해제할 때까지 스케줄링 자격이 없다가, 해제된 순간 JVM이 내부적으로 다시 락 획득을 시도해, 성공하면 RUNNABLE이 된다(결론).

###### 6.2 WAITING 상태

1. 쓰레드가 이미 모니터를 확보한 상황에서 `obj.wait()`를 호출하면, 모니터를 자발적으로 반납하고 wait set에 들어가게 된다(원칙).
2. 이는 “**명시적 신호(notify)**”가 오기 전까지는 깨어날 의사가 없음”을 표현하는 것이다(작동).
3. notify() 또는 notifyAll()을 통해 신호가 오면, 쓰레드는 모니터 재획득을 위해 경쟁하다가(그 사이 `BLOCKED`가 될 수도 있음) 성공시 RUNNABLE이 되고, 결국 wait() 호출 직후 코드를 계속 실행한다(결론).

**정리**: BLOCKED는 “모니터를 얻지 못해 문 앞에서 기다림”이고, WAITING은 “모니터를 얻었다가, wait() 호출로 임시 반납 후 notify를 기다리는 상태”입니다.

---

<!-- /curriculum-chunk -->

### 7. notify가 없으면 wait한 스레드는 어떻게 되는가?

#### 원문: 7. notify가 없으면 wait한 스레드는 어떻게 되는가?

<!-- curriculum-chunk: sha256=5bb4abec82ee2b4472b53da72b32794b50f166bdf26f85b5e1db0623bbb7c809 major=concurrency-async-io mid=Java monitor와 wait/notify sub=7. notify가 없으면 wait한 스레드는 어떻게 되는가? sources=interview_questions3.md:491-498 -->

> Source: `interview_questions3.md:491-498`
> Classification reason: monitor/wait-notify

##### 7. notify가 없으면 wait한 스레드는 어떻게 되는가?

1. 만약 어떤 스레드가 `obj.wait()`로 WAITING 상태에 진입했는데, 이후 어떤 스레드도 `obj.notify()`나 `obj.notifyAll()`을 호출하지 않는다면(가정),
2. 그 쓰레드는 영원히 wait set에서 깨어나지 못하고, **영구적으로 WAITING 상태**로 남습니다(결론).
3. 실제 시스템에서도, wait/notify를 잘못 구현하면 데드락(deadlock)이나 **무한 대기**가 발생할 수 있으므로, 반드시 논리적으로 짝을 맞춰야 합니다(교훈).

---

<!-- /curriculum-chunk -->

### 7. 결론 정리

#### 원문: 7. 결론 정리

<!-- curriculum-chunk: sha256=d8916525ad5ed1cbf51d60cdbb6bd2469af888eef730d7b00abbfbb1fb8ef8cd major=concurrency-async-io mid=Java monitor와 wait/notify sub=7. 결론 정리 sources=interview_questions3.md:933-947 -->

> Source: `interview_questions3.md:933-947`
> Classification reason: monitor/wait-notify

##### 7. 결론 정리

1. `synchronized(LockObject)` 구문을 만나면, 스레드는 LockObject의 **모니터**를 획득해야 합니다. 만약 누군가 이미 락을 갖고 있으면 해당 스레드는 **BLOCKED** 상태가 됩니다.
2. `LockObject.wait()`는 “현재 쓰레드가 모니터를 자발적으로 놓고, wait set에서 WAITING 상태가 된다”는 의미입니다. 이때, doSomething() 같은 다음 로직은 잠시 멈추게 되고, **notify**나 **notifyAll**이 호출되어야만 깨어날 수 있습니다.
3. 깨어난 쓰레드는 다시 모니터를 획득하기 위해 **BLOCKED** 상태를 거칠 수 있고, 획득에 성공하면 RUNNABLE 상태로 돌아가 **doSomething()**을 실행합니다.
4. `notify()`와 `notifyAll()`은 자바가 자동으로 호출해주지 않으므로, **명시적으로** 호출해야 합니다. 보통 “상태 변화를 일으키는 쪽(예: 작업 추가, 특정 플래그 변경)”에서 알맞은 시점에 `notify()`를 호출합니다.
5. `synchronized`는 **메모리 장벽**을 제공하여, 락을 잡고 있는 동안 이뤄진 변경사항이 락 해제 후 다른 스레드에게 확실히 보이도록 합니다.
6. **실무/프로덕션 코드**에서는 “**Producer-Consumer** 패턴” 같은 큐 기반 구조에서 `wait()`/`notify()`를 자주 활용하며, TDD로 이 로직을 검증할 때는 **Thread + Sleep + assert** 같은 방식을 사용하거나, **CountDownLatch**, **CyclicBarrier**, **Awaitility** 등의 테스트 유틸을 함께 쓰기도 합니다.

**최종 결론**:
- 모든 쓰레드가 wait만 하고 notify를 아무도 안 해주면, 영원히 깨어나지 못하므로 **교착(deadlock) 비슷한 상태**가 생깁니다.
- `doSomething()`은 “대기(wait) → notify → 모니터 재획득” 순서가 성립해야 실행됩니다.
- `synchronized(락)`을 사용하면, 해당 락(모니터)에 대해 “상호 배타적 접근(mutex) + 쓰기/읽기의 가시성(메모리 장벽)”이 보장됩니다.

이상으로, **“모든 스레드가 wait를 반복할 것 같은 상황에서 언제 doSomething이 실행되는지, notify는 누가 어떻게 호출해야 하고, 그 호출 후의 흐름이 어떻게 이어지는지, 실전 프로덕션 코드 예시와 TDD 방식을 곁들여”** 자세히 살펴보았습니다.

<!-- /curriculum-chunk -->

### 8. JVM 레벨에서의 synchronized 작동 방식

#### 원문: 8. JVM 레벨에서의 synchronized 작동 방식

<!-- curriculum-chunk: sha256=06a678eddb2cae922c806faacf62f24c3f926def22401f979690f9191ccb1ad7 major=concurrency-async-io mid=Java monitor와 wait/notify sub=8. JVM 레벨에서의 synchronized 작동 방식 sources=interview_questions3.md:499-515 -->

> Source: `interview_questions3.md:499-515`
> Classification reason: monitor/wait-notify

##### 8. JVM 레벨에서의 synchronized 작동 방식

###### 8.1 모니터 엔터/익시트(enter/exit)

1. 자바 바이트코드 상에서 `synchronized`는 `monitorenter`, `monitorexit` 명령어로 번역된다(원칙).
2. 실제 실행 시, JVM은 객체 헤더(mark word)와 모니터 구조를 이용해 락 상태, 소유 스레드, 대기열 등을 관리한다(작동).
3. `monitorenter`는 락 획득에 성공할 때까지 BLOCKED로 대기할 수도 있고, 성공하면 RUNNABLE로 진행한다(결론).
4. 블록 끝, 예외 발생 시 등에서 `monitorexit`가 호출되어 모니터를 해제한다.

###### 8.2 경량화 기법(Biased Locking, Lightweight Locking)

1. HotSpot JVM은 락 오버헤드를 줄이기 위해 **Biased Locking**이나 **Lightweight Locking** 같은 최적화를 한다(원칙).
2. 단일 스레드가 계속해서 같은 락을 잡는다면 편향 모드(biased)로 전환해 락 재획득 비용을 제거(작동).
3. 스레드 간 실질적 경합이 일어나기 전까지, 락 연산이 매우 가볍게 처리된다(결론).

---

<!-- /curriculum-chunk -->

### 9. 실제 동작 시나리오 예시

#### 원문: 9. 실제 동작 시나리오 예시

<!-- curriculum-chunk: sha256=f0cc627e4757c059dc91c52d2e7e4557d9eb23b009c2bd6dfbfa35698f124a93 major=concurrency-async-io mid=Java monitor와 wait/notify sub=9. 실제 동작 시나리오 예시 sources=interview_questions3.md:516-534 -->

> Source: `interview_questions3.md:516-534`
> Classification reason: monitor/wait-notify scenario

##### 9. 실제 동작 시나리오 예시

이제 위의 개념을 토대로, 질문에서 주어진 예시 코드를 분석해봅니다:

```java
synchronized(LockObject) {
    LockObject.wait();
    doSomething();
}
```

1. 이 구문이 실행될 때, 우선 `synchronized(LockObject)`에 진입하려면, 현재 쓰레드는 **LockObject 모니터**를 획득해야 합니다. 이미 다른 쓰레드가 이 모니터를 잡고 있으면, 이 쓰레드는 **BLOCKED 상태**가 되어 모니터 획득을 대기합니다.
2. 만약 성공적으로 모니터를 얻었다면, `LockObject.wait()`를 호출합니다. 이 순간, 쓰레드는 **(a) 모니터를 release**하고, **(b) LockObject의 wait set**에 들어가며, 쓰레드 상태는 **WAITING**으로 바뀝니다.
3. 이제 다른 쓰레드가 `LockObject.notify()` 또는 `notifyAll()`을 호출해주기 전까지는, 이 쓰레드는 깨어나지 못합니다.
4. notify가 도착하면, 대기 중인 쓰레드(혹은 여러 쓰레드) 중 하나(또는 전부)가 **“모니터 재획득”**을 시도합니다. 락을 획득하지 못하면 **BLOCKED**, 획득에 성공하면 **RUNNABLE** 상태로 돌아가 `doSomething()`을 실행합니다.
5. `doSomething()` 실행 후, `synchronized(LockObject)` 블록 끝에 도달해 **monitorexit**가 일어나고, 쓰레드는 모니터를 해제합니다.

---

<!-- /curriculum-chunk -->

### 쓰레드 wait와 notify 예제

#### 원문: 쓰레드 wait와 notify 예제

<!-- curriculum-chunk: sha256=81a56e3576b3a49162d40755c09be27585f58f529d702af36e6d3ff954e0df7b major=concurrency-async-io mid=Java monitor와 wait/notify sub=쓰레드 wait와 notify 예제 sources=interview_questions3.md:731-932 -->

> Source: `interview_questions3.md:731-932`
> Classification reason: monitor/wait-notify

##### 쓰레드 wait와 notify 예제

1. `synchronized` 블록 내부에서 `wait()`가 반복해서 호출되면 도대체 언제 `doSomething()`이 실행되는가?
2. `notify()`는 누가, 어떻게 호출해야 하는가? 자동 호출되는가?
3. 현실 세계의 **프로덕션 코드**에서 이런 로직을 어떻게 작성하며, **TDD(Test-Driven Development)** 원칙에 따라 테스트 코드를 어떻게 설계할 수 있는가?
4. `synchronized`가 메모리 장벽(메모리 배리어)을 제공하므로, 쓰레드 간에 공유 객체에 대한 접근을 어떻게 제어하는지?

아래 코드는 “**생산자-소비자(Producer-Consumer)**” 시나리오로 구현해 보았습니다. “생산자 스레드”가 작업(Tasks)을 큐에 추가하면 `notify()`를 호출하여, 대기 중인 “소비자 스레드”들을 깨워서 `doSomething()`을 수행하게 하는 구조입니다.

---

###### 1. 왜 `notify()`가 필요하고, 누가 언제 호출해야 하는가?

###### 요약

- `wait()`로 **WAITING 상태**가 된 스레드는, **반드시** 다른 쓰레드가 `notify()`(또는 `notifyAll()`)를 호출해야만 깨어날 수 있습니다.
- `notify()`는 자바에서 **자동으로** 호출되는 것이 아니라, **코드 상에서 명시적으로** 호출해야 합니다.
- `notify()`를 호출하는 스레드도 `synchronized(lockObj)` 블록 안에서 모니터를 잡은 상태여야 합니다(왜냐하면 wait/notify 모두 모니터를 필요로 하기 때문입니다).
- `notify()`를 호출하고 블록을 빠져나가야(=모니터 해제) 다른 쓰레드가 모니터를 재획득할 수 있으므로, 그때 비로소 깨어난 스레드가 `doSomething()`을 실행할 기회를 얻게 됩니다.

###### 예시 흐름

1. 스레드 A가 `synchronized(lockObj)`로 모니터 획득
2. `lockObj.wait()` 호출 → **모니터 해제** & **WAITING 상태** & **lockObj의 wait set**에 들어감
3. 스레드 B가 `synchronized(lockObj)`로 진입 (A가 해제했으므로 B가 락 획득 가능)
4. 스레드 B가 어떤 조건을 달성하면 `lockObj.notify()` 호출 후, `synchronized` 블록을 빠져나옴
5. 대기중이던 A 스레드가 깨어나서(=notify 신호) 모니터 재획득 시도 → 획득 성공 → `doSomething()` 실행

**중요**: 모든 스레드가 들어가자마자 `wait()`만 호출한다면, **아무도 notify()를 호출하지 않는 상황**이 되어 영원히 깨어나지 못할 수 있습니다. 실제 프로덕션 코드에서는 “이벤트가 발생했을 때 notify()를 호출하는” 쪽 로직이 반드시 필요합니다.

---

###### 2. 현실 세계 프로덕션 코드 예시: “TaskQueue” 시나리오

**시나리오**:
- `TaskQueue` 클래스에 작업(Task)을 추가(생산자)하고, 작업이 없으면 대기(소비자)가 발생.
- 소비자는 `wait()`를 통해 **대기**하고, 작업이 들어오면 `notify()`를 통해 **깨운다**.

###### 2.1 TaskQueue 구현 (프로덕션 코드)

```java
public class TaskQueue {
    private final Object lock = new Object();
    private final LinkedList<String> tasks = new LinkedList<>();

    /**
     * 작업을 큐에 추가하고, 대기 중인 스레드를 깨운다.
     */
    public void addTask(String task) {
        synchronized(lock) {
            tasks.add(task);
            // 작업이 추가되었으니, wait()로 대기 중인 스레드를 깨운다.
            lock.notify();
        }
    }

    /**
     * 큐에서 작업을 꺼낸다.
     * 만약 큐가 비어 있으면 wait()로 대기한다.
     */
    public String getTask() throws InterruptedException {
        synchronized(lock) {
            while (tasks.isEmpty()) {
                // 여기서 락을 포기하고 wait set에 들어가 WAITING 상태
                lock.wait();
            }
            // 깨어난 뒤 모니터를 다시 획득해야 함 (획득 시 BLOCKED → RUNNABLE)
            return tasks.removeFirst();
        }
    }

    /**
     * 큐에 남아있는 작업 수를 반환.
     */
    public int size() {
        synchronized(lock) {
            return tasks.size();
        }
    }
}
```

1. `addTask()` 메서드는 새 작업을 추가한 후, `lock.notify()`를 호출합니다. 이렇게 함으로써 **대기(wait) 중인 스레드**가 있다면 하나를 깨울 수 있습니다.
2. `getTask()`는 큐가 비어있으면 `lock.wait()`를 호출하여, 모니터를 즉시 릴리즈하고 **WAITING 상태**로 들어갑니다. 여기서 `while`문을 사용하는 이유는, 깨어난 뒤(=notify)에도 여전히 큐가 비어있을 수 있기 때문입니다(“spurious wakeup” 처리).
3. `size()` 등 다른 메서드도 `synchronized(lock)`를 통해 동시성을 제어합니다.

---

###### 2.2 실제 doSomething() 메서드 예시

이제 `doSomething()` 로직을 “**작업을 실제 처리**”하는 것으로 가정해봅시다. 예를 들어, “큐에서 가져온 task를 처리”하는 스레드 코드를 작성할 수 있습니다.

```java
public class TaskProcessor {
    private final TaskQueue queue;

    public TaskProcessor(TaskQueue queue) {
        this.queue = queue;
    }

    /**
     * 무한 루프 형태의 소비자 스레드 역할
     * - 큐에서 작업을 꺼내 처리(doSomething()).
     * - 큐가 비어있으면 wait() 상태로 대기.
     */
    public void processLoop() {
        while (true) {
            try {
                String task = queue.getTask();
                // 여기서 getTask()는 synchronized(lock) + wait()를 통한 블록
                doSomething(task);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break; // 스레드를 종료할 수도 있음
            }
        }
    }

    private void doSomething(String task) {
        // 실제 로직: 작업 처리
        System.out.println("Processing: " + task);
        // 여기서 DB 저장, 파일 쓰기 등등 수행할 수 있음
    }
}
```

- `TaskProcessor`는 “소비자(Consumer) 스레드”로, `queue.getTask()`를 통해 작업이 들어올 때까지 대기(`wait()`)하다가, 작업이 있을 때만 깨어나서 `doSomething()`을 수행합니다.
- `getTask()` 메서드는 내부적으로 `synchronized`+`wait()`+`notify()`를 통해 동작하므로, 안전하게 큐를 공유합니다.

---

###### 3. 테스트 코드 (TDD 접근)

JUnit을 사용하여, **멀티쓰레드 환경**에서 정상 동작하는지 간단한 테스트를 보여주겠습니다.

```java
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertTrue;

public class TaskQueueTest {

    @Test
    public void testProducerConsumer() throws InterruptedException {
        TaskQueue queue = new TaskQueue();
        TaskProcessor processor = new TaskProcessor(queue);

        // 1. 소비자 스레드(Worker) 생성
        Thread worker = new Thread(processor::processLoop);
        worker.start();

        // 2. 작업을 여러 개 추가 (생산자 역할)
        queue.addTask("Task1");
        queue.addTask("Task2");

        // 3. 소비자가 처리할 시간을 조금 주고...
        Thread.sleep(500);

        // 4. 아직 queue가 비어 있기를 기대
        int size = queue.size();
        assertTrue(size == 0, "All tasks should have been consumed by now.");

        // 5. 스레드 종료를 위해 인터럽트
        worker.interrupt();
        worker.join();
    }
}
```

1. **`TaskQueue queue`**를 만들고, **`TaskProcessor processor`**로 소비자 스레드를 구성합니다.
2. `processor::processLoop`는 무한 루프에서 `queue.getTask()` → `doSomething(task)`를 반복합니다. 큐가 비어있을 경우 자동으로 `wait()`를 통해 대기한다.
3. 생산자 역할(`main` 스레드)이 `queue.addTask("TaskX")`를 호출하면, 내부에서 `notify()`가 일어나고 소비자가 깨어나 `doSomething()`를 수행합니다.
4. 좀 기다린 후(queue가 비어있을 시간 간격) `assertTrue(size == 0)`로 테스트합니다.

---

###### 4. notify는 명시적으로 호출해야 하며, 누가 언제 실행할까?

1. **명시적 호출**: 자바 언어에서는 **notify/notifyAll**을 자동으로 호출해주지 않습니다. 프로그래머가 적절한 시점에 `lockObj.notify()`를 해야만 **WAITING** 스레드를 깨울 수 있습니다.
2. **누가 언제**: 일반적으로 “**상태 변화**”가 일어날 때(예: 큐가 비어 있다가 새 작업이 들어온 경우, 어떤 조건이 충족되었을 때) `notify()`를 호출합니다. 위 예시의 `addTask()`처럼, “중요한 상태 변화를 일으킨 쪽”이 다른 스레드(대기 중인 소비자)에게 알려주는 개념입니다.
3. **notify() 호출 스레드**는 `synchronized(obj)` 블록 안에서 모니터를 점유 중이어야 `obj.notify()`를 실행할 수 있습니다. 이 스레드가 블록을 빠져나와야 모니터가 해제되고, 깨어난 스레드가 락을 획득할 수 있습니다.

---

###### 5. notify()를 호출한 스레드는 어떻게 되나요?

1. notify()를 호출한 스레드는, 일반적으로 **`notify()` 코드가 속한 synchronized 블록**이 끝날 때까지 모니터를 계속 점유합니다.
2. 블록이 끝나는 시점(`}`)에 `monitorexit`가 실행되며, 모니터가 해제됩니다.
3. 그제서야 WAITING 중이던 스레드가 모니터 획득을 시도(BLOCKED → RUNNABLE)하여, 성공하면 `wait()` 다음 코드를 진행(`doSomething()`)합니다.

즉, **notify()**를 호출하는 순간 곧바로 모니터를 뺏기는 것이 아니라, “**synchronized 블록이 끝날 때까지**”는 계속 모니터를 유지합니다. 이는 **“모니터를 여러 스레드가 동시에 점유할 수 없다”**는 점과 일관됩니다.

---

###### 6. `synchronized(lock)` 사용 시 메모리 장벽(Memory Barrier) 적용 여부

1. 자바에서 **synchronized** 블록은 **모니터 락**을 통해, 쓰레드가 블록에 진입할 때와 빠져나올 때 **메모리 장벽**(happens-before 관계)을 설정합니다.
2. 구체적으로, 블록에 진입하기 전까지 쓰레드가 캐시에 가지고 있던 값들이 모두 비워지고, 블록을 빠져나오기 전까지 수행된 writes가 메인 메모리에 반영됩니다.
3. 그 결과, `synchronized`가 보장하는 것은 “**락 해제 후 다른 스레드가 락을 잡았을 때, 전 스레드가 한 변경 사항을 볼 수 있다**”는 것입니다.
4. 따라서 `synchronized(lock)`로 감싼 로직은, 락을 통해 “원자적(atomic) + 가시성(visibility) 보장”이 됩니다. 이는 흔히 “**mutex + 메모리 배리어**”라고 말할 수 있습니다.

---

<!-- /curriculum-chunk -->

### 자바 쓰레드, 상태, 그리고 모니터 등

#### 원문: 자바 쓰레드, 상태, 그리고 모니터 등

<!-- curriculum-chunk: sha256=0021bab710687cef9aa4a021f615003eeb3c2129a1665f90deb1f5f0e092509f major=concurrency-async-io mid=Java monitor와 wait/notify sub=자바 쓰레드, 상태, 그리고 모니터 등 sources=interview_questions3.md:394-395 -->

> Source: `interview_questions3.md:394-395`
> Classification reason: monitor/wait-notify

##### 자바 쓰레드, 상태, 그리고 모니터 등

<!-- /curriculum-chunk -->

### 자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시

#### 원문: 자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시

<!-- curriculum-chunk: sha256=3f7a46d63875ce1f079336d99f76be43ca8b15c26d2cbb3c43740558cb1fddf1 major=concurrency-async-io mid=Java monitor와 wait/notify sub=자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시 sources=interview_questions3.md:551-730 -->

> Source: `interview_questions3.md:551-730`
> Classification reason: monitor/wait-notify

##### **자바에서 wait/notify 패턴**을 조금 더 현실감 있게 적용한 **프로덕션 코드 예시**

여러 스레드가 동시에 `wait()`를 호출할 수 있고, 어떤 스레드는 `notify()` 또는 `notifyAll()`을 호출해서 대기 중인 스레드를 깨우는 구조입니다.

###### 1. 핵심 개념 요약

1. **wait()**:
   - 현재 **모니터(동기화 락)**를 가진 스레드가 스스로 락을 해제하고, 해당 객체의 **wait set**에서 대기(WAITING 상태)로 들어감.
   - 추후 **notify** 또는 **notifyAll**이 호출되어야만 다시 모니터 획득을 재시도할 수 있음.

2. **notify() / notifyAll()**:
   - 같은 모니터를 사용하는 객체에서, **wait set**에 대기 중인 스레드들을 깨움.
   - 깨운 스레드는 모니터를 다시 획득해야만(RUNNABLE로 복귀 전) `wait()` 다음 코드를 실행할 수 있음.

3. **synchronized**(LockObject) 블록:
   - 해당 블록에 진입하기 위해서는 `LockObject`의 모니터를 획득해야 함.
   - 모니터는 “한 순간에 오직 한 스레드만 접근 가능”을 보장하여 상호배타(Mutual Exclusion)를 실현.
   - synchronized 블록 안에서는 **메모리 베리어**(happens-before)가 보장되어, 해당 블록 내부에서 이루어진 변경은 블록을 빠져나갈 때 외부에 반영되고, 반대로 외부 변경사항을 읽을 때 재주문(reordering)이 억제됨.

4. **"모든 스레드가 wait만 하고 notify를 안 하면?"**
   - 영원히 깨어나지 못함(Deadlock 또는 영구 대기).
   - 따라서 반드시 **notify()** 또는 **notifyAll()**을 호출해야 대기 중인 스레드들이 다시 동작을 이어갈 수 있음.

---

###### 2. 예시 코드: 작업 준비(Producer)와 작업 처리(Consumer)

아래 시나리오에서:
- 다수의 Worker 스레드가 `wait()`를 호출해 대기 상태에 들어갑니다.
- Main 스레드(또는 다른 관리 스레드)가 특정 타이밍에 `notifyAll()`을 통해 Worker 스레드들을 깨웁니다.
- 깨어난 Worker 스레드는 모니터를 다시 획득한 후 `doSomething()` 로직을 실행합니다.
- 실제 프로덕션 환경에서는 “job 큐”를 사용하거나, “조건변수(Condition)”를 사용하는 것이 더 직관적이지만, 여기서는 `wait/notify`의 기본 구조를 보여주기 위해 단순화했습니다.

```java
public class WaitNotifyDemo {

    private final Object lockObject = new Object();
    private boolean canProcess = false;
    // 이 변수는 'notify' 시점에 true로 바뀌며, 다른 스레드가 doSomething 해도 좋다는 신호라고 가정.

    /**
     * Worker 스레드가 호출하는 메서드.
     * 이 메서드는 'canProcess'가 false인 상태에서는 wait()를 통해 대기한다.
     * 이후 notifyAll()로 깨어나면 doSomething()을 수행한다.
     */
    public void workerMethod() {
        synchronized (lockObject) {
            while (!canProcess) {
                try {
                    // [1] 모니터 획득 중
                    // [2] wait()를 호출 → 모니터 반납 + wait set 대기 (WAITING 상태)
                    lockObject.wait();
                    // 깨어난 시점에 모니터 재획득 필요 → 만약 다른 스레드가 락을 점유 중이면 BLOCKED로 대기
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    System.out.println("Worker interrupted during wait");
                    return;
                }
            }
            // 이 시점: canProcess == true 이며, 모니터를 다시 획득한 상태
            doSomething();
            // synchronized 블록 끝 → monitorexit
        }
    }

    /**
     * 다른 스레드나 컨트롤러가 호출해서 'canProcess'를 true로 만들고,
     * 대기 중인 Worker 스레드들을 깨우는 역할을 한다.
     */
    public void enableProcessingAndNotifyAll() {
        synchronized (lockObject) {
            // [1] 모니터 획득
            // [2] canProcess 값을 true로 설정
            canProcess = true;
            // [3] wait set에 있는 모든 스레드 깨우기
            lockObject.notifyAll();
            // notifyAll()은 wait set에 있는 스레드들을 깨운다.
            // 하지만 깨어난 스레드들이 모니터 재획득을 위해 경합 → 한 번에 한 스레드씩 들어와서 doSomething() 실행
        }
    }

    /**
     * 실제로 처리해야 할 로직 (예: DB 저장, API 호출 등)
     */
    private void doSomething() {
        System.out.println(Thread.currentThread().getName() + " is doing something...");
        // 좀 더 복잡한 로직이라고 가정해도 됨
        try {
            Thread.sleep(500); // 시뮬레이션
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * 테스트를 위한 메인: 여러 Worker 스레드를 띄우고, 일정 시간 뒤 enableProcessingAndNotifyAll()을 호출.
     */
    public static void main(String[] args) throws InterruptedException {
        WaitNotifyDemo demo = new WaitNotifyDemo();

        // 스레드 5개 생성 → 모두 workerMethod()를 수행
        for (int i = 1; i <= 5; i++) {
            Thread worker = new Thread(demo::workerMethod, "Worker-" + i);
            worker.start();
        }

        // 이 시점에 canProcess = false 이므로, workerMethod 내에서 모두 wait() 상태로 진입할 것임
        System.out.println("All workers are likely waiting now...");

        // 3초 후에 notifyAll()을 호출 (canProcess = true)
        Thread.sleep(3000);
        System.out.println("Now let's enable processing and notify all!");
        demo.enableProcessingAndNotifyAll();

        // 이후 main 스레드가 종료될 때까지 대기
        // 실제로 Worker 스레드들이 깨어나서 doSomething()을 실행할 것임
    }
}
```

###### 2.1 코드 흐름 설명

1. **Worker 스레드** (`workerMethod()`):
   - `synchronized (lockObject)`로 진입
   - `while(!canProcess)` 조건 체크 → `canProcess`가 `false`면 `wait()` 호출
   - `wait()`를 호출하면 모니터를 반납하고 `WAITING` 상태로 들어감.
   - (나중에) `notifyAll()`로 깨워진 뒤 → 모니터 재획득(이때 BLOCKED 일 수 있음) → `doSomething()` 실행.

2. **Notifier 스레드** 또는 **Main 스레드** (`enableProcessingAndNotifyAll()`):
   - `synchronized (lockObject)`로 진입
   - `canProcess = true;` 설정 → 이제 Worker가 doSomething을 수행해도 된다는 조건 충족
   - `lockObject.notifyAll()` 호출 → wait set에서 WAITING 중인 스레드들을 깨움
   - 깨어난 스레드가 모니터를 얻으면 `while(!canProcess)`를 탈출하고 `doSomething()` 호출

3. **테스트(main)**:
   - 5개의 Worker 스레드가 `workerMethod()`를 시작 → 모두 `wait()`로 대기
   - 3초 뒤 `enableProcessingAndNotifyAll()`이 호출 → 모든 worker가 깨어나 `doSomething()` 진행

이 시나리오에서는 “모든 스레드가 락을 얻자마자 `wait()`해버리면 `notify()`는 어떻게 호출되냐?”라는 의문에 대해, 실제로는 **다른 스레드(혹은 동일 스레드의 다른 시점)**가 같은 락을 잡고 `notify()`를 실행해야만 된다는 사실을 보여줍니다.

---

###### 3. notify는 누가 언제 실행하고, 그 결과는 무엇인가?

1. notify/notifyAll은 **동일한 모니터 객체**에 대해 `synchronized(obj)` 블록 내부에서만 유효합니다. 즉, 락을 획득한 스레드만 `obj.notify()`를 호출할 수 있습니다.
2. 위 예시에서 `enableProcessingAndNotifyAll()` 메서드가 바로 그 "notify를 하는 로직"입니다. 이 메서드를 실행하는 스레드(메인 스레드)는 **모니터 획득** → `canProcess = true`로 변경 → `notifyAll()`을 호출하고 → 블록을 빠져나갑니다.
3. notifyAll()이 호출되면, **모니터의 wait set**에 있던 스레드가 깨어납니다. 그러나 모니터는 여전히 메인 스레드가 잡고 있으므로, Worker 스레드가 깨어나도 **우선은 BLOCKED**가 될 수 있습니다.
4. 메인 스레드가 모니터를 release(블록 끝)하면, 깨어난 Worker 스레드 중 하나가 모니터를 획득 → `while(!canProcess)` 조건을 재확인 → 이제 true이므로 탈출 → `doSomething()` 실행.
5. 그 다음 모니터를 해제하면 또 다른 Worker 스레드가 모니터를 획득하고 로직을 수행합니다.

---

###### 4. "모든 스레드가 wait만 하고 notify를 호출하는 스레드가 없다면?" 발생하는 문제

1. 모든 스레드가 `wait()`로 들어가면, **WAITING 상태**에서 멈춰있습니다.
2. **notify**나 **notifyAll**을 호출하는 스레드가 없다면, 깨어날 방법이 없습니다.
3. 따라서 “**무한 대기**” 혹은 “데드락”이 발생합니다.
4. 이를 방지하려면, **로직상 반드시** 특정 시점에(조건을 만족하면) `notify()`나 `notifyAll()`을 호출해야 합니다.

---

###### 5. synchronized(락)을 사용하면 메모리 베리어가 적용되는가?

1. 자바 `synchronized` 구문은 **락을 획득할 때**와 **락을 해제할 때** 메모리 베리어(정확히는 happens-before 관계)를 부여합니다.
2. 즉, 한 스레드가 synchronized 블록 안에서 변경한 데이터는 블록을 빠져나오기 전에 **메모리 가시화**가 이루어집니다. 그리고 다음에 모니터를 획득한 스레드는 그 변경된 상태를 볼 수 있게 됩니다.
3. 따라서 `synchronized`를 적절히 사용하면, 다른 스레드가 임의로 바꾼 값을 못 읽거나, reorder로 인해 예기치 않은 동작이 벌어지는 상황을 완화할 수 있습니다.
4. wait/notify 또한 synchronized 블록 내부에서만 허용되므로, 이로 인한 상태 변경도 자연스럽게 메모리 베리어 효과가 반영됩니다.

---

###### 6. 마지막 정리

1. **모니터**(= 객체 락, synchronized)는 “한 순간에 오직 한 스레드만 해당 구간을 실행”하도록 보장합니다.
2. 모든 스레드가 `synchronized(LockObject)`에 들어오자마자 `wait()`만 호출한다면, 실제로는 누군가가 그 모니터를 잡고 `notify()`를 호출하지 않는 이상, 깨어날 스레드가 없게 됩니다.
3. 따라서 “누군가가 `notify()`나 `notifyAll()`을 명시적으로 호출”해야 합니다. 자동 호출되는 일이 절대 없으므로, 프로그래머가 필요한 시점에서 로직을 설계해야 합니다.
4. 위 예시 코드처럼 **조건변수** 성격으로 `wait/notify`를 쓰고자 한다면, 보통 `while(!condition) wait()` 패턴을 씁니다. 이것이 자바에서 표준적인 접근(“guarded block”)입니다.
5. 메모리 베리어 관점에서 `synchronized`는 쓰레드 간 메모리 일관성을 높여주며, wait/notify로 멀티쓰레드 간 통신(이벤트/신호)을 구현합니다.

이상으로, 질문에서 제기된 **“A스레드, B스레드 모두 `wait()`만 반복하면 `doSomething()`은 언제 실행되는가?”, “notify는 명시적으로 호출해야 하는가?”, “메모리 장벽은 어떻게 보장되는가?”** 등 내용을 모두 포함하여, **프로덕션 코드 예시**를 통해 논리적으로 설명했습니다.

<!-- /curriculum-chunk -->

## OS I/O multiplexing

### Node.js의 싱글 스레드와 libuv의 멀티 스레딩

#### 원문: Node.js의 싱글 스레드와 libuv의 멀티 스레딩

<!-- curriculum-chunk: sha256=f7f95ffe9d57af36d470d8c74dabbcc0a72646d43e7e056faffc0f03cc7c6fac major=concurrency-async-io mid=OS I/O multiplexing sub=Node.js의 싱글 스레드와 libuv의 멀티 스레딩 sources=interview_questions.md:8897-9127, interviews.md:8845-9075 -->

> Source: `interview_questions.md:8897-9127`
> Classification reason: io multiplexing
> Duplicate source aliases: `interview_questions.md:8897-9127, interviews.md:8845-9075`

##### Node.js의 싱글 스레드와 libuv의 멀티 스레딩

Node.js는 흔히 싱글 스레드(single-threaded)라고 알려져 있지만, 내부적으로는 libuv를 통해 멀티 스레드 작업을 수행합니다.

Node.js는 자바스크립트 런타임이기 때문에, 기본적으로 자바스크립트 코드를 실행하는 이벤트 루프(Event Loop)는 하나의 스레드에서 실행됩니다.
이 싱글 스레드 이벤트 루프가 모든 비동기 작업을 처리하는 것처럼 보이지만, 실제로는 비동기 작업의 종류에 따라 libuv가 동작을 관리하며, 때에 따라 멀티 스레딩을 사용합니다.

Node.js에서 비동기 I/O 작업을 처리하는 방식은 크게 두 가지로 나눌 수 있습니다:
1. I/O 작업: 네트워크 요청, 파일 읽기/쓰기, DNS 조회 등의 비동기 작업.
2. CPU 바운드 작업: 암호화, 압축 등의 CPU 집약적인 작업.

libuv는 Node.js의 비동기 I/O 작업을 처리하는 핵심 라이브러리입니다.
libuv는 *이벤트 루프를 관리*하고, *비동기 작업의 결과가 준비되면 Node.js에 알려주는 역할*을 합니다.

- libuv에서의 I/O 작업 처리

    대부분의 I/O 작업(파일 읽기/쓰기, 네트워크 요청 등)은 실제로 운영 체제에서 비동기 처리가 가능합니다.

    예를 들어, 리눅스에서는 커널의 비동기 I/O 메커니즘을 사용하여 작업을 처리합니다.
    이 경우, libuv는 운영 체제의 비동기 기능을 사용하여 작업을 비동기로 처리하고, 그 결과가 완료되면 이벤트 루프를 통해 처리합니다.
    리눅스 경우에는 다음과 같은 비동기 I/O 방식들을 제공하며, libuv 경우 주로 `epoll`을 사용합니다.

    - Non-blocking I/O (논블로킹 I/O)

        시스템 콜이 즉시 리턴되도록 하여, 작업이 완료될 때까지 기다리지 않고 다른 작업을 수행할 수 있도록 하는 방식입니다.
        일반적으로 파일 디스크립터에 대해 non-blocking 모드로 설정한 후, 데이터를 읽거나 쓸 때 I/O 작업이 바로 완료되지 않더라도 시스템 콜이 즉시 반환됩니다.

        `O_NONBLOCK` 플래그로 소켓이나 파일을 논블로킹 모드로 설정합니다.
        데이터가 준비되지 않았을 때 시스템 콜은 바로 실패를 리턴하고, 나중에 다시 시도할 수 있습니다.

        그러나 논블로킹 I/O는 비동기 I/O와는 다르게, 결과를 바로 받을 수 없기 때문에 루프를 돌면서 확인해야 하므로 CPU 리소스를 많이 소비할 수 있습니다.

    - `epoll` 또는 `select` (이벤트 기반 I/O)

        리눅스에서는 `epoll`(이전의 `select()` 또는 `poll()`)을 사용하여 다수의 파일 디스크립터(소켓, 파일, 파이프 등)의 상태 변화를 이벤트 기반으로 감시할 수 있습니다.
        `epoll`은 Linux에서 제공하는 고성능 이벤트 감지 시스템 호출로, 주어진 파일 디스크립터(FD)들의 상태 변화를 감시하는 기능을 합니다.

        `epoll`은 논블로킹 I/O의 단점을 보완하기 위한 메커니즘입니다.
        `epoll`은 다수의 파일 디스크립터를 등록한 후, 해당 디스크립터들에 읽기, 쓰기, 연결 등 이벤트가 발생했을 때 이를 감지하여 알림을 받습니다.

        이 방법은 여러 I/O 작업이 동시에 진행될 때, 각 I/O 작업이 완료되면 이를 감지하여 적절한 콜백을 호출합니다.

        FD 유형별 감지 가능 여부:
        - 소켓(Socket): TCP/UDP 네트워크 소켓(✅ 가능)
        - 파이프(Pipe): 익명/이름 있는 파이프 (✅ 가능)
        - 터미널(TTY, PTS): 가상 터미널 장치 (✅ 가능)
        - 이벤트 파일(inotify, signalfd 등): 파일 시스템 이벤트 모니터링 (✅ 가능)
        - 일반 파일: 디스크에 저장된 파일 (❌ 불가능)

        Linux에서 일반 파일(Regular File, HDD/SSD에 저장된 파일)의 FD는 "항상 읽기 가능(ready-to-read)" 상태이기 때문에 `epoll`로 감지할 수 없습니다. `epoll`은 "파일 디스크립터의 상태 변화를 감지"하는 시스템입니다. 하지만 일반 파일은 비어 있거나, 파일 포인터가 끝에 도달하지 않는 한 항상 읽기 가능합니다. 따라서 이벤트 루프에서 일반 파일의 "읽기 가능" 상태를 감지하는 것은 의미가 없습니다.
        - 파일을 읽기 위해서는 디스크에서 데이터를 가져와야 하므로 물리적인 지연(latency)이 발생.
        - 파일은 이벤트 루프에서 기다릴 필요 없이 즉시 읽을 수 있으므로, 이벤트 기반 방식과 맞지 않음.

        ```c
        int file_fd = open("test.txt", O_RDONLY);
        epoll_ctl(epoll_fd, EPOLL_CTL_ADD, file_fd, &event);  // ❌ 실패: 일반 파일은 epoll 감지가 불가능
        ```

        네트워크 소켓은 클라이언트가 데이터를 보낼 때만 "읽기 가능" 상태가 되므로 감지가 가능하지만, 파일은 항상 읽기 가능 상태이므로 감지가 불가능합니다.

        소켓(Socket)의 상태 변화는 네트워크 I/O 이벤트가 발생할 때마다 달라집니다.
        - `EPOLLIN`: 소켓에 읽을 수 있는 데이터가 있음
        - `EPOLLOUT`: 소켓이 쓰기 가능 상태
        - `EPOLLERR`: 소켓에 오류 발생
        - `EPOLLHUP`: 연결이 종료됨
        소켓은 연결이 없으면 읽을 수 없고, 데이터가 들어오면 읽을 수 있는 상태가 됩니다. 따라서 이벤트 루프에서 epoll을 사용하여 "데이터가 도착했을 때만 실행"하도록 최적화할 수 있습니다.

        libuv는 이러한 `epoll` 인터페이스를 이용해 이벤트 루프에서 I/O 작업이 준비되었을 때 처리할 수 있도록 합니다.
        `epoll`은 커널이 파일 디스크립터의 상태 변화를 감지하는 효율적인 메커니즘이므로, 다수의 클라이언트로부터 들어오는 네트워크 요청을 처리할 때 유용합니다.

    - AIO (Asynchronous I/O)

        AIO(비동기 I/O)는 리눅스에서 제공하는 고급 I/O 처리 방식으로, 커널 레벨에서 직접 비동기 작업을 처리할 수 있습니다.
        AIO는 작업이 완료되기를 기다리지 않고, 작업이 백그라운드에서 실행된 후 완료되면 그 결과를 알리는 방식입니다.

        `aio_read()`, `aio_write()` 등의 함수는 커널이 파일 읽기/쓰기 작업을 비동기적으로 수행하도록 요청합니다.
        작업이 끝나면, 커널은 시그널 또는 콜백을 통해 완료를 알리거나, `aio_suspend()`를 통해 완료 여부를 폴링할 수 있습니다.

        그러나 *AIO는 제한적인 파일 시스템에서만 효율적으로 작동*하며, 주로 네트워크가 아닌 디스크 I/O에서 사용됩니다.
        또한 AIO의 구현은 다소 복잡하고 성능 이슈가 있어, 실무에서는 자주 사용되지는 않습니다.

    - io_uring

        io_uring은 최근 리눅스 커널에서 도입된 I/O 인터페이스로, 비동기 I/O를 더 효율적으로 처리할 수 있도록 설계되었습니다.
        io_uring은 사용자 공간과 커널 공간 간의 데이터를 주고받을 때 발생하는 시스템 콜 오버헤드를 줄이고, 비동기 작업을 더욱 빠르고 효과적으로 처리할 수 있습니다.

        io_uring은 큐 기반 메커니즘을 사용합니다.
        커널과 사용자 공간 간에 공유 메모리 영역을 두고, 사용자는 I/O 요청을 큐에 넣으면 커널이 비동기적으로 이를 처리합니다.
        작업이 완료되면 사용자 공간에서 결과를 확인할 수 있습니다.
        이 방식은 기존의 epoll이나 AIO 방식보다 더 효율적입니다.

        libuv는 아직 io_uring을 기본적으로 사용하지 않지만, io_uring의 도입으로 비동기 I/O 성능이 크게 향상될 수 있습니다.

    이 과정에서 싱글 스레드에서 이벤트 루프가 해당 작업이 완료되었음을 감지하고, 콜백을 실행합니다.
    CPU 작업 없이 I/O만 처리하는 작업들은 일반적으로 멀티 스레드를 사용하지 않습니다.

- libuv에서의 멀티 스레드 활용 (Thread Pool)

    libuv는 기본적으로 이벤트 루프 기반의 논블로킹(Non-blocking) I/O 모델을 따릅니다.
    하지만 파일 읽기, 쓰기와 같은 비동기 파일 I/O의 경우, 네트워크 통신에서 사용하는 소켓과 달리,
    파일 디스크립터(File Descriptor)를 epoll과 같은 I/O 다중화 메커니즘에 직접 등록하지는 않습니다.
    파일 시스템의 I/O는 네트워크 소켓 I/O와 다르게 비동기 처리를 지원하지 않는 경우가 많기 때문입니다.

    리눅스 및 유닉스 계열 시스템에서 네트워크 소켓은 epoll, kqueue, IOCP 같은 비동기 I/O 멀티플렉싱 기술을 사용할 수 있습니다.

    - `epoll`, `kqueue`, `IOCP`는 이벤트 기반으로 동작하므로, 네트워크 소켓의 상태 변화를 이벤트 루프에서 감지하고 처리할 수 있음.
    - 따라서 네트워크 소켓의 읽기/쓰기 이벤트가 발생하면, 이벤트 루프가 이를 감지하여 논블로킹 방식으로 처리 가능.

    ```c
    // epoll을 사용한 네트워크 I/O 처리 예제
    int epoll_fd = epoll_create1(0);
    struct epoll_event event;
    event.events = EPOLLIN;
    event.data.fd = server_socket;
    epoll_ctl(epoll_fd, EPOLL_CTL_ADD, server_socket, &event);

    while (1) {
        struct epoll_event events[MAX_EVENTS];
        int num_events = epoll_wait(epoll_fd, events, MAX_EVENTS, -1);  // 이벤트 발생 대기

        for (int i = 0; i < num_events; i++) {
            if (events[i].events & EPOLLIN) {
                // 비동기적으로 데이터 읽기 처리
            }
        }
    }
    ```

    즉, 네트워크 소켓은 OS 차원에서 비동기 처리를 지원하므로 이벤트 루프에서 효율적으로 관리할 수 있습니다.

    반면 epoll, kqueue, IOCP 등의 이벤트 기반 감지 시스템은 파일 디스크립터의 상태 변화를 감지하지 못합니다.즉, 파일이 읽기 가능해질 때까지 기다리는 Polling 방식이 불가능하며, 파일 읽기/쓰기를 시작하면 OS가 즉시 블로킹됩니다.
    - 이벤트 루프에서 파일 I/O 작업을 처리하려면, 파일 읽기 요청이 들어올 때마다 이벤트 루프가 멈춰야 함.
    - 이는 네트워크 I/O와는 다르게 비효율적인 구조를 만들게 됨.

    ```c
    // epoll을 사용하여 파일 I/O를 처리할 수 없음
    int fd = open("file.txt", O_RDONLY);
    struct epoll_event event;
    event.events = EPOLLIN;
    event.data.fd = fd;
    epoll_ctl(epoll_fd, EPOLL_CTL_ADD, fd, &event);  // 실패: 파일 디스크립터는 epoll로 감시할 수 없음
    ```

    즉, 파일 I/O는 epoll을 사용할 수 없으므로, 논블로킹 이벤트 루프에서 다루기에 적절하지 않습니다.

    대신 스레드 풀(thread pool)을 사용하여 일부 비동기 작업을 병렬로 처리합니다.
    스레드 풀의 작업이 완료되면, 그 결과는 이벤트 루프로 다시 전달되어 해당 작업의 콜백이 실행됩니다.

    1. 파일 디스크립터 생성 및 사용:
        파일도 소켓과 마찬가지로 파일 디스크립터(File Descriptor, FD)를 사용합니다.
        파일을 열 때 OS는 파일 디스크립터를 할당하며, 이 디스크립터를 통해 파일에 접근하게 됩니다.
        하지만 *네트워크 소켓과는 달리, 파일 시스템은 기본적으로 비동기 I/O를 지원하지 않습니다*.
        이를 해결하기 위해, 리눅스에서는 주로 `스레드 풀`이나 `AIO`(Asynchronous I/O) 같은 방식을 사용합니다.

        애플리케이션이 `uv_fs_read()` 또는 `uv_fs_write()`를 호출하여 파일 읽기/쓰기 비동기 작업을 요청합니다.

        ```c
        void on_read(uv_fs_t* req) {
            printf("File Read Completed: %s\n", (char*)req->bufs->base);
            uv_fs_req_cleanup(req);
        }

        void read_file(uv_loop_t* loop) {
            uv_fs_t req;
            uv_buf_t buffer = uv_buf_init((char*)malloc(1024), 1024);

            uv_fs_open(loop, &req, "example.txt", O_RDONLY, 0, NULL);
            uv_fs_read(loop, &req, req.result, &buffer, 1, 0, on_read);  // 비동기 파일 I/O
        }
        ```

        - `uv_fs_read()`는 이벤트 루프를 블로킹하지 않으며, 스레드 풀에서 파일을 읽음.
        - 파일 읽기가 완료되면 `on_read()` 콜백이 실행됨.

        즉, libuv는 파일 I/O를 이벤트 루프에서 직접 처리하는 것이 아니라, 별도의 스레드에서 실행한 후 콜백을 통해 결과를 반환합니다.

    2. 스레드 풀 기반 비동기 I/O (libuv의 기본 방식):
        libuv는 네트워크 소켓과 달리 파일 시스템에 대한 비동기 요청을 처리할 때 스레드 풀(thread pool)을 활용합니다.
        이는 기본적으로 동기 방식으로 처리되는 파일 I/O를 백그라운드 스레드에서 비동기적으로 처리하는 방식입니다.
        따라서 libuv는 파일 읽기 또는 쓰기 작업을 수행할 때 파일을 열어 파일 디스크립터를 생성하지만, 이 디스크립터는 네트워크 소켓처럼 `epoll`에 등록되지 않습니다.

        - 스레드 풀에서 파일 작업 실행:
            libuv는 내부적으로 스레드 풀을 사용하여 파일 읽기/쓰기와 같은 블로킹 작업을 백그라운드 스레드에서 실행합니다.
            이를 통해 메인 이벤트 루프가 블로킹되지 않고 다른 작업을 계속 처리할 수 있게 됩니다.

        - 작업 완료 후 콜백 실행:
            스레드 풀에서 작업이 완료되면, libuv는 메인 이벤트 루프에 이를 통지하고, 등록된 콜백 함수(`on_read`)를 실행하여 완료된 작업에 대한 후속 처리를 수행합니다.
            파일 읽기 또는 쓰기 작업이 백그라운드 스레드에서 완료되면, 메인 이벤트 루프로 돌아와 콜백을 실행합니다.

    3. Linux AIO(Asynchronous I/O):
        리눅스에서는 AIO라는 비동기 파일 I/O 메커니즘을 제공합니다.
        AIO는 파일 I/O 작업을 비동기적으로 처리하며, 이를 사용하면 별도의 스레드 풀 없이 비동기 파일 I/O를 수행할 수 있습니다.
        하지만 AIO는 일반적으로 네트워크 소켓보다는 복잡하고, 모든 파일 시스템에서 지원하지 않기 때문에 널리 사용되지는 않습니다.

        기본적으로 libuv는 스레드 풀 방식을 사용하지만, 일부 시스템에서 AIO를 활용할 수도 있습니다.
        그러나 대부분의 경우 스레드 풀 방식이 더 널리 사용됩니다.

    결국 비동기 파일 I/O에서 epoll을 사용하지 않는 이유는 파일 시스템의 I/O 작업은 네트워크 소켓과 달리 블로킹 방식으로 동작하기 때문입니다.

    Node.js의 싱글 스레드 이벤트 루프와는 별개로, libuv는 네 가지 작업 유형에서 스레드 풀을 사용하여 비동기 처리를 합니다.
    - 파일 시스템 작업:
        파일 읽기/쓰기와 같은 I/O 작업은 운영 체제에서 비동기 처리가 불가능한 경우가 많습니다.
        이때, libuv는 스레드 풀을 사용하여 이러한 작업을 백그라운드에서 처리합니다.
        파일을 읽거나 쓸 때, libuv는 이를 스레드 풀의 스레드 중 하나에 할당하여 동시 작업을 처리합니다.

    - DNS 조회:
        일부 DNS 조회 작업은 운영 체제에서 비동기 처리가 되지 않습니다.
        libuv는 스레드 풀을 이용해 이러한 작업도 비동기적으로 처리할 수 있습니다.

    - 암호화 및 압축:
        CPU 집약적인 작업(예: 암호화, 압축, 해싱)도 스레드 풀에서 실행됩니다.
        이 작업들은 이벤트 루프에서 직접 처리되지 않고, 스레드 풀에서 처리한 후 결과가 준비되면 다시 이벤트 루프로 전달됩니다.

    - 사용자 정의 비동기 작업:
        C++ 애드온을 사용하는 경우, libuv의 스레드 풀을 사용하여 사용자 정의 비동기 작업을 처리할 수 있습니다.

    libuv는 기본적으로 4개의 스레드를 갖춘 스레드 풀을 사용합니다.
    이는 libuv의 기본 설정이지만, `UV_THREADPOOL_SIZE` 환경 변수를 통해 스레드 풀 크기를 조정할 수 있습니다.
    스레드 풀의 크기를 늘리면 더 많은 비동기 작업을 병렬로 처리할 수 있지만, 너무 많은 스레드가 생성되면 오히려 성능이 저하될 수 있습니다.

결국 전체적인 작업 흐름은 다음과 같습니다.
1. I/O 작업:

    네트워크 요청, 파일 읽기/쓰기 등은 운영 체제에서 비동기로 처리할 수 있는 경우, libuv는 이를 커널에 위임합니다.
    커널에서 작업이 완료되면 이벤트 루프에서 그 사실을 감지하여 처리합니다.

2. 스레드 풀 사용 작업:

    파일 시스템 작업이나 CPU 집약적인 작업은 libuv의 스레드 풀에서 처리됩니다.
    작업이 완료되면 그 결과를 이벤트 루프가 처리하고, 콜백 함수를 호출합니다.

<!-- /curriculum-chunk -->

### epoll 상세

#### 원문: epoll 상세

<!-- curriculum-chunk: sha256=9f6edefaad1c2f587d5d7f19477786fedf9b0c38cca16269960736f4fa108db3 major=concurrency-async-io mid=OS I/O multiplexing sub=epoll 상세 sources=interview_questions.md:9128-9213, interviews.md:9076-9161 -->

> Source: `interview_questions.md:9128-9213`
> Classification reason: io multiplexing
> Duplicate source aliases: `interview_questions.md:9128-9213, interviews.md:9076-9161`

##### epoll 상세

`epoll`은 리눅스에서 제공하는 시스템 콜이며, C 언어나 다른 고수준 언어에서 이를 호출하여 사용합니다.
시스템 콜은 일반적으로 운영 체제의 커널 기능을 호출하는 방식으로, 애플리케이션 레벨에서 커널 레벨로 작업을 위임합니다.

- `epoll`과 C 언어

    `epoll`은 리눅스에서 제공하는 이벤트 기반 비동기 I/O 처리를 위한 시스템 콜 중 하나입니다.
    C 언어에서 직접 호출할 수 있습니다.
    파일 디스크립터(file descriptor)를 등록하고, 등록된 파일 디스크립터에서 이벤트가 발생했을 때 알림을 받는 방식으로 동작합니다.

    C 언어에서 `epoll` 관련 함수는 다음과 같습니다:
    - `epoll_create`: `epoll` 인스턴스를 생성하는 함수.
    - `epoll_ctl`: 파일 디스크립터를 `epoll` 인스턴스에 등록하거나, 수정하거나, 삭제하는 함수.
    - `epoll_wait`: 등록된 파일 디스크립터에 이벤트가 발생했을 때 대기하고, 그 이벤트를 반환하는 함수.

    C에서 `epoll` 사용하는 코드는 다음과 같습니다.

    ```c
    #include <sys/epoll.h>
    #include <fcntl.h>
    #include <unistd.h>
    #include <stdio.h>

    int main() {
        int epfd = epoll_create(1); // epoll 인스턴스 생성
        struct epoll_event event;
        struct epoll_event events[10];

        int fd = open("somefile.txt", O_RDONLY); // 파일 열기
        event.events = EPOLLIN; // 읽기 이벤트를 감지
        event.data.fd = fd;

        epoll_ctl(epfd, EPOLL_CTL_ADD, fd, &event); // epoll에 파일 디스크립터 추가

        // 이벤트 대기
        int n = epoll_wait(epfd, events, 10, -1); // 이벤트 발생 시까지 대기

        // 처리
        for (int i = 0; i < n; i++) {
            if (events[i].events & EPOLLIN) {
                printf("Data available to read\n");
            }
        }

        close(fd);
        close(epfd);
        return 0;
    }
    ```

    위의 예제에서는 `epoll_create`를 통해 `epoll` 인스턴스를 만듭니다.
    `epoll_ctl`을 통해 파일 디스크립터를 등록합니다.
    그 후 `epoll_wait`를 사용하여 이벤트가 발생할 때까지 대기합니다.

- 시스템 콜과 어셈블리

    `epoll`은 시스템 콜이므로, 결국 커널 레벨의 작업을 호출하는 명령입니다.

    시스템 콜은 일반적으로 사용자 공간에서 커널 공간으로 제어를 넘기고 커널에서 동작하는 기능을 호출하는 것입니다.
    C에서 호출될 수 있지만, 실제로 시스템 콜을 트리거하는 어셈블리 레벨의 명령어인 `syscall` 또는 `int 0x80` 등을 통해 커널과 통신합니다.
    시스템 콜을 호출하는 방식은 CPU 아키텍처에 따라 다르지만, 리눅스 x86-64 시스템에서는 `syscall` 명령어가 주로 사용됩니다.

    C 언어에서 `epoll`을 호출하면, C 컴파일러가 시스템 콜 번호와 매개변수를 설정하고, 어셈블리 명령을 사용하여 커널에 제어를 넘깁니다.
    1. 사용자 공간에서 시스템 콜 호출:
        C 코드에서 `epoll_wait`와 같은 시스템 콜을 호출하면, 이 호출은 실제로 시스템 콜 트랩을 통해 커널로 전달됩니다.

    2. 커널 모드 전환:
        CPU는 사용자 모드에서 커널 모드로 전환되며, 커널은 요청된 시스템 콜을 처리합니다.

    3. 커널에서 작업 수행:
        커널은 시스템 콜의 매개변수를 확인하고, 그에 맞는 작업을 수행합니다.
        `epoll`의 경우, 파일 디스크립터의 상태를 확인하거나 대기하는 작업을 합니다.

    4. 결과 반환:
        작업이 완료되면 커널은 결과를 다시 사용자 공간으로 반환하고, CPU는 사용자 모드로 돌아갑니다.

    ```assembly
    ; 시스템 콜의 어셈블리 예시 (x86-64 기준)
    mov rax, 시스템콜_번호  ; epoll_wait의 경우, 시스템 콜 번호가 rax 레지스터에 들어갑니다.
    mov rdi, 첫번째_인자     ; 첫 번째 인자는 rdi 레지스터에
    mov rsi, 두번째_인자     ; 두 번째 인자는 rsi 레지스터에
    mov rdx, 세번째_인자     ; 세 번째 인자는 rdx 레지스터에
    syscall                 ; syscall 명령을 사용하여 커널에 요청
    ```

<!-- /curriculum-chunk -->

### libuv, aio_*, io_uring

#### 원문: libuv, aio_*, io_uring

<!-- curriculum-chunk: sha256=3f5711fdf977cc4175636a2dc6393efe4eba9f2754124183105205148b4a0b0e major=concurrency-async-io mid=OS I/O multiplexing sub=libuv, aio_*, io_uring sources=interview_questions.md:3874-4103, interviews.md:3874-4103 -->

> Source: `interview_questions.md:3874-4103`
> Classification reason: io multiplexing
> Duplicate source aliases: `interview_questions.md:3874-4103, interviews.md:3874-4103`

##### libuv, aio_*, io_uring

- libuv

    libuv는 Node.js와 같은 비동기 애플리케이션을 위한 이벤트 기반 비동기 I/O 라이브러리입니다.

    1. 이벤트 루프(Event Loop) 초기화

        libuv의 핵심은 이벤트 루프입니다.
        이벤트 루프는 작업을 관리하며, I/O 이벤트가 발생할 때 적절한 콜백 함수를 호출합니다.
        이 루프는 애플리케이션이 계속해서 동작하면서도 비동기적으로 많은 작업을 처리할 수 있게 해줍니다.

        - 애플리케이션이 시작되면 libuv는 이벤트 루프를 초기화합니다.
        - 플랫폼별 I/O 다중화 메커니즘(예: Linux의 `epoll`, macOS의 `kqueue`, Windows의 `IOCP`)을 설정합니다.
            - 리눅스의 `epoll`은 이벤트 기반으로 대규모 파일 디스크립터를 효율적으로 처리할 수 있으며, macOS의 `kqueue`는 비슷한 방식으로 동작하지만 macOS와 BSD에서 사용됩니다.
            - Windows의 경우, I/O Completion Ports(IOCP)를 사용해 완료 기반 모델로 비동기 작업을 처리하며, 운영체제가 작업 완료 시 이를 비동기적으로 통지하는 방식입니다.

        libuv는 이와 같은 추상화 계층을 통해 운영체제별 차이를 숨기고, 모든 플랫폼에서 동일한 비동기 인터페이스를 제공합니다.

    2. 논 블로킹 I/O 요청

        애플리케이션은 논 블로킹 I/O 작업(파일 읽기/쓰기, 네트워크 요청 등)을 요청합니다.
        libuv는 이 요청을 내부 작업 큐에 추가하고, 이를 OS의 I/O 서브시스템에 전달합니다.
        이때 애플리케이션은 블로킹되지 않고, 작업이 완료될 때까지 기다리지 않습니다.

        libuv는 요청을 큐에 추가한 뒤 OS에 작업을 전달하며, 요청이 완료되면 콜백 함수가 실행됩니다.
        - 콜백 함수는 작업이 완료된 후에만 실행되며, 작업이 완료되기 전에는 호출되지 않습니다.
        - 작업 큐는 요청이 비동기적으로 처리되기 전까지 해당 요청을 관리하며, OS의 비동기 메커니즘을 통해 요청이 처리될 때 이를 콜백으로 연결합니다.

            - 비동기 메커니즘:
                비동기 I/O 메커니즘은 운영체제(OS)가 제공하는 I/O 다중화 기능을 사용하여 여러 I/O 작업을 병렬로 관리하고,
                그 작업이 완료되었을 때만 애플리케이션에 알림을 주는 방식입니다.

                각 플랫폼은 각기 다른 비동기 메커니즘을 제공합니다:

                - Linux의 `epoll`: 다수의 파일 디스크립터에서 이벤트가 발생했을 때 이를 감시하고, 발생한 이벤트만 비동기적으로 애플리케이션에 전달합니다.
                - macOS의 `kqueue`: 유사하게 다수의 파일 디스크립터나 소켓에서 이벤트를 감시하고, 이벤트가 발생했을 때 이를 처리합니다.
                - Windows의 `IOCP`: 비동기 I/O 완료 시 완료 포트를 통해 완료된 이벤트를 전달하고 처리합니다.

                이 비동기 메커니즘을 사용하면, CPU가 I/O 작업을 직접 기다리지 않고도 많은 요청을 병렬로 처리할 수 있게 됩니다.
                요청이 완료될 때까지 블로킹되지 않으며, 커널은 I/O가 완료되면 이를 애플리케이션에 통보합니다.

            - 요청 처리

                libuv에서 I/O 요청이 처리되는 과정은 다음과 같이 이루어집니다:

                1. 애플리케이션이 비동기 I/O 요청을 발생시키면, libuv는 이 요청을 OS에 전달합니다.
                    예를 들어, 파일을 읽거나 네트워크 데이터를 수신하는 요청일 수 있습니다.

                2. 이 요청은 libuv의 작업 큐에 먼저 저장되며, 비동기적으로 처리됩니다.
                    libuv는 OS의 I/O 다중화 메커니즘(`epoll`, `kqueue`, `IOCP`)을 사용하여 이러한 요청이 완료되기를 기다립니다.

                3. 요청이 커널로 전달되면, 커널은 이 작업을 처리하기 위해 파일 시스템 또는 네트워크 스택과 같은 하위 시스템을 통해 실제 작업을 수행합니다.
                    여기서 중요한 점은 비동기적 방식으로 요청을 처리한다는 것입니다.
                    즉, 요청을 OS가 처리하는 동안 애플리케이션은 다른 작업을 계속 수행할 수 있습니다.

                4. I/O 작업이 완료되면, 커널은 이를 비동기 메커니즘(예: `epoll`, `kqueue`, `IOCP`)을 통해 통지합니다.

            - 콜백으로 연결

                I/O 요청이 완료되었을 때, 커널은 libuv에 이를 알리고, libuv는 콜백 함수를 호출하여 해당 I/O 작업이 완료되었음을 애플리케이션에 전달합니다.
                이 과정은 다음과 같이 구체적으로 이루어집니다:

                1. I/O 완료 이벤트 발생:
                    OS의 비동기 I/O 메커니즘은 I/O 작업이 완료되었음을 감지하고, 해당 이벤트를 libuv의 이벤트 루프로 전달합니다.
                    예를 들어, `epoll`은 수신된 소켓에 데이터가 있음을 libuv에 알립니다.

                2. 이벤트 루프에서 콜백 실행:
                    libuv의 이벤트 루프는 OS에서 전달받은 이벤트 큐를 감시합니다.
                    I/O 작업이 완료되었다는 이벤트가 발생하면, 등록된 콜백 함수를 호출하여 해당 요청이 완료되었음을 애플리케이션에 알립니다.

                3. 콜백 함수 실행:
                    콜백 함수는 이벤트 루프의 다음 순환에서 실행되며, 이때 애플리케이션은 해당 I/O 작업에 대한 후속 작업을 수행할 수 있게 됩니다.
                    예를 들어, 파일을 모두 읽은 후 그 내용을 처리하거나, 네트워크로부터 수신한 데이터를 처리하는 작업이 이루어집니다.

                이 콜백 구조는 비동기 I/O 모델의 핵심입니다. 비동기 방식에서는 요청이 완료될 때까지 기다리지 않고, OS가 작업을 완료하면 콜백을 통해 작업 결과를 전달합니다. libuv는 이 구조를 통해 높은 동시성을 제공하며, 애플리케이션이 블로킹되지 않고 많은 작업을 처리할 수 있게 합니다.

    3. 이벤트 루프 시작

        이벤트 루프는 다음 단계들을 반복적으로 실행합니다:

        1. 타이머 확인: 등록된 타이머가 만료되었는지 확인하고, 만료된 타이머의 콜백을 실행합니다.
        2. 보류 중인 콜백 실행: 완료된 비동기 작업의 콜백을 처리합니다.
        3. idle, prepare 핸들러 실행: CPU가 유휴 상태일 때 실행되는 핸들러로, 필요한 작업을 처리합니다.
        4. I/O 폴링 (핵심 단계)

            이벤트 루프는 OS의 *I/O 다중화 메커니즘*을 사용하여 I/O 이벤트를 폴링합니다.
            이를 통해 파일 디스크립터와 네트워크 소켓에서 이벤트가 발생했는지 확인합니다.
            - Linux에서는 `epoll`, macOS에서는 `kqueue`, Windows에서는 `IOCP`(I/O Completion Ports)와 같은 시스템 호출을 통해 커널이 I/O 작업의 완료 상태를 알립니다.
            - libuv는 이 메커니즘을 통해 각 플랫폼에서 발생한 이벤트를 감지하고, 이를 처리합니다. 이는 CPU를 블로킹하지 않고 비동기적으로 다수의 작업을 처리할 수 있게 합니다.

        5. check 핸들러 실행: 이전 단계에서 완료된 I/O 작업 후 추가적인 처리를 진행합니다.
        6. close 콜백 실행: 종료된 핸들러에 대해 후속 처리를 진행합니다.

    4. 하드웨어와 OS의 상호작용

        libuv는 OS의 논블로킹 I/O 메커니즘을 사용하여 하드웨어와 상호작용합니다.
        I/O 작업이 발생하면, 디스크 컨트롤러, 네트워크 인터페이스 카드(NIC)와 같은 하드웨어가 실제로 I/O 작업을 수행하고, 완료되면 인터럽트를 통해 OS에 알립니다.

        1. I/O 요청이 발생하면, libuv는 이를 `epoll`, `kqueue`, 또는 `IOCP` 같은 메커니즘을 통해 커널에 전달합니다.
        2. OS는 이 작업을 관리하고, DMA(Direct Memory Access)를 사용해 메모리 버퍼에 데이터를 적재합니다.
           이 과정은 CPU의 개입 없이 진행되며, 하드웨어가 완료되었음을 인터럽트를 통해 OS에 알립니다.
        3. OS는 I/O 작업의 완료를 기록하며, I/O 작업이 완료되었음을 이벤트 큐에 기록합니다. 이때 I/O가 완료되었음을 libuv에 통지합니다.

    5. 이벤트 처리 및 콜백 실행

        I/O 작업이 완료되면 OS는 libuv가 관리하는 이벤트 큐에 작업 완료 신호를 보냅니다.
        libuv의 이벤트 루프는 이 큐를 확인하고, 작업이 완료된 I/O 요청에 대한 콜백 함수를 실행합니다.

        - 이벤트 큐:
            libuv는 완료된 I/O 작업을 처리하기 위해 이벤트 큐를 사용합니다.
            이 큐에 완료된 작업이 기록되면, 해당 작업에 대한 콜백이 호출됩니다.

        - 콜백 함수의 실행 시점:
            작업이 완료된 후에만 콜백이 실행되며, 비동기 모델에 따라 이벤트 루프가 처리 가능한 시점에서 콜백이 호출됩니다.

        추가적으로, libuv는 타이머 이벤트도 처리합니다.
        타이머가 만료되면 해당 타이머에 설정된 콜백이 실행됩니다.
        또한, 네트워크 I/O 작업의 경우 소켓에서 데이터가 수신되면 이를 이벤트 루프를 통해 처리하고, 적절한 콜백 함수가 실행됩니다.

    이와 같은 구조를 통해 libuv는 비동기 작업을 효율적으로 처리하며, OS, 하드웨어와 애플리케이션 간의 원활한 상호작용을 가능하게 합니다.

- [POSIX `aio_*`](https://pubs.opengroup.org/onlinepubs/009695399/basedefs/aio.h.html)

    POSIX에서 비동기 I/O 시스템 호출인 `aio_*` 함수들을 제공합니다.
    이를 통해 사용자가 비동기적으로 파일 I/O 작업을 수행할 수 있게 해줍니다.
    `aio_read`, `aio_write` 등은 파일 디스크립터를 비동기적으로 읽고 쓸 수 있도록 설계되었습니다.

    1. AIO 서브시스템을 초기화
        - 애플리케이션이 `aio_init()`을 호출하여 AIO 서브시스템을 초기화합니다 (선택적).
        - AIO 컨트롤 블록(`struct aiocb`)을 생성하고 초기화합니다.

    2. 비동기 I/O 요청

        애플리케이션이 `aio_read()` 또는 `aio_write()`와 같은 시스템 호출을 통해 비동기 I/O 작업을 요청합니다.
        이때 작업을 설명하는 `aiocb`(Asynchronous I/O Control Block) 구조체에 작업에 대한 세부 정보(파일 디스크립터, 오프셋, 버퍼, 바이트 수 등)를 인자로 받습니다.

        커널의 AIO 서브시스템이 I/O 요청을 관리합니다.

    3. I/O 요청 커널 전달

        `aio_*` 함수는 내부적으로 커널에 비동기 I/O 요청을 전달합니다.
        커널은 I/O 요청을 작업 큐에 추가하고, 즉시 제어를 애플리케이션에 반환합니다.
        이때 블로킹 없이 요청을 처리하므로, 커널 내부에서 I/O 작업이 완료될 때까지 사용자 프로세스는 다른 작업을 계속할 수 있습니다.

    4. 하드웨어와 OS의 상호작용

        이 요청은 하드웨어 I/O 서브시스템에 전달되어 실제 I/O 작업이 진행됩니다.
        하드웨어는 디스크 I/O 또는 네트워크 I/O 작업을 처리하고, 작업이 완료되면 OS에 완료 신호를 보냅니다.
        이 과정은 애플리케이션의 메인 실행 흐름과 독립적으로 진행됩니다.

        I/O 스케줄러는 하드웨어와의 상호작용을 최적화하여, 필요한 디스크 또는 네트워크 I/O 작업을 수행합니다.
        I/O 작업은 Direct Memory Access (DMA) 등을 통해 CPU의 개입 없이 메모리와 디바이스 간 데이터를 전송할 수 있습니다.

    5. I/O 작업 완료 알림

        I/O 작업이 완료되면, 하드웨어가 인터럽트를 통해 커널에 작업 완료를 알립니다.
        커널은 완료 상태를 AIO 컨트롤 블록에 업데이트하고, 시그널 또는 이벤트를 사용하여 애플리케이션에 이 사실을 알립니다.
        애플리케이션은 미리 등록한 콜백 함수를 통해 작업이 완료되었음을 알 수 있습니다.

        통지 방식에 따라 다음 중 하나가 발생합니다:
        - 시그널 발송 (`SIGEV_SIGNAL`)
        - 스레드 생성 및 콜백 함수 호출 (`SIGEV_THREAD`)
        - 아무 동작 없음 (`SIGEV_NONE`)

    6. 결과 처리

        - `aio_error()`: I/O 작업이 완료되었는지, 오류가 발생했는지 여부를 확인할 수 있게 합니다.
        - `aio_return()`: 완료된 I/O 작업의 결과(전송된 바이트 수 등)를 반환합니다.

    POSIX 비동기 I/O는 주로 C와 C++ 언어에서 사용되며, 특히 고성능 네트워크 애플리케이션이나 파일 입출력 시스템에서 많이 사용됩니다.
    - GlusterFS: 분산 파일 시스템으로, 높은 성능과 비동기 I/O 처리를 위해 aio_*를 사용합니다.
    - Ceph: 고성능 분산 스토리지 시스템으로, 비동기 I/O 작업을 최적화하기 위해 aio_*를 사용하는 경우가 많습니다.
    - Nginx: 고성능 웹 서버로서 비동기 I/O 처리에 aio_*를 사용하여 요청 처리 성능을 향상시킵니다.
    - MySQL: 데이터베이스 서버에서도 대규모 파일 입출력을 최적화하기 위해 aio_* 시스템 호출을 사용하는 경우가 있습니다.

- `io_uring`

    io_uring은 최신 Linux 커널 5.1부터 제공하는 고성능 비동기 I/O API입니다.
    기존 비동기 I/O 메커니즘에 비해 더욱 효율적이고 강력한 성능을 제공합니다.
    io_uring은 *링 버퍼*를 사용하여 사용자 공간과 커널 공간 간의 상호작용을 개선하였습니다.

    1. 초기화
        - 애플리케이션이 `io_uring_setup()`을 호출하여 io_uring 인스턴스를 생성합니다.
        - 제출 큐(SQ)와 완료 큐(CQ)라는 두 개의 락프리 링 버퍼가 생성됩니다.
            - 서브미션 큐(submission queue, SQ)

                애플리케이션은 I/O 요청을 서브미션 큐에 기록합니다. 사용자 공간에 존재합니다.

            - 컴플리션 큐(completion queue, CQ)

                커널이 서브미션 큐의 I/O 작업을 완료하면 컴플리션 큐에 완료된 작업을 기록합니다.

    2. 메모리 매핑

        애플리케이션은 `mmap()`을 사용하여 SQ와 CQ를 자신의 주소 공간에 매핑합니다.
        이를 통해 커널과 사용자 공간 간의 직접적인 데이터 교환이 가능해집니다.

    3. I/O 요청

        애플리케이션은 SQ 엔트리(SQE)를 준비합니다.
        각 SQE는 수행할 I/O 작업의 세부 정보를 포함합니다.

        준비된 SQE들을 `io_uring_enter()` 시스템 콜을 통해 커널에 제출합니다.
        또는 `io_uring_submit()` 래퍼 함수를 사용할 수 있습니다.

        애플리케이션이 비동기 I/O 작업을 요청하면, 서브미션 큐에 작업을 추가합니다.

        커널은 SQ를 폴링하여 새로운 요청을 확인합니다.
        각 요청에 대해 비동기 I/O 작업을 시작합니다.

        이 큐는 사용자 공간에 존재하며, 시스템 호출을 하지 않고도 I/O 요청을 기록할 수 있기 때문에 매우 빠르게 요청을 제출할 수 있습니다.

    4. 커널과 하드웨어 상호작용

        커널은 서브미션 큐에 기록된 요청을 처리하여 하드웨어 I/O 서브시스템에 전달합니다.
        하드웨어는 요청된 I/O 작업을 처리하고, 작업이 완료되면 결과를 커널에 반환합니다.

    5. 컴플리션 큐에서 결과 확인

        I/O 작업이 완료되면 커널은 컴플리션 큐에 작업 결과를 기록합니다.
        이 큐는 사용자 공간에 노출되어 있으므로, 애플리케이션은 시스템 호출을 하지 않고도 작업 완료 여부를 확인할 수 있습니다.

        애플리케이션은 CQ를 폴링하여 완료된 작업을 확인합니다.

    6. 결과 처리

        애플리케이션은 컴플리션 큐에서 작업 결과를 확인하고, 완료된 작업에 대해 후속 처리를 수행합니다.
        io_uring은 이러한 작업을 매우 효율적으로 수행할 수 있도록 설계되었으며, 기존 메커니즘보다 낮은 오버헤드로 비동기 I/O 작업을 처리합니다.

<!-- /curriculum-chunk -->

### 네트워크 통신 시 epoll 동작 방식

#### 원문: 네트워크 통신 시 epoll 동작 방식

<!-- curriculum-chunk: sha256=280db8c4d50445da1912415592bf4b630881eaab1204fe71cb5d759cbc6bee9b major=concurrency-async-io mid=OS I/O multiplexing sub=네트워크 통신 시 epoll 동작 방식 sources=interview_questions.md:6499-6611, interviews.md:6499-6611 -->

> Source: `interview_questions.md:6499-6611`
> Classification reason: io multiplexing
> Duplicate source aliases: `interview_questions.md:6499-6611, interviews.md:6499-6611`

##### 네트워크 통신 시 epoll 동작 방식

1. 외부 네트워크 요청의 시작

    사용자가 웹 브라우저나 HTTP 클라이언트(예: `curl`)를 통해 웹 서버에 요청을 보냅니다.
    이 요청은 일반적으로 TCP 연결을 통해 이루어지며, HTTP, HTTPS, 또는 다른 프로토콜을 통해 전달됩니다.

    요청이 도달하면 인터넷을 통해 NGINX가 동작하는 서버의 네트워크 인터페이스 카드(NIC)로 전달됩니다.

2. 커널 네트워크 스택에서 TCP 세그먼트 처리

    서버 운영체제의 커널은 네트워크 스택을 관리하며, 네트워크 인터페이스를 통해 들어오는 패킷을 처리합니다.

    1. NIC에서 커널로 전달
        - 클라이언트가 보내는 패킷이 네트워크 인터페이스 카드(NIC)를 통해 수신됩니다.
        - NIC는 이 패킷을 커널 메모리 공간의 수신 버퍼로 전달합니다. 이 과정은 일반적으로 DMA(Direct Memory Access) 기술을 사용해 CPU 개입 없이 이루어집니다.

    2. 커널의 TCP/IP 스택에서 패킷 처리
        - 커널의 TCP/IP 네트워크 스택은 수신 버퍼에 도착한 패킷을 분석하고, 해당 패킷이 TCP 패킷인 경우 TCP 세그먼트로 재조립합니다.
        - TCP 세그먼트가 올바르게 조립되면, 이를 소켓과 연결된 수신 버퍼로 전달합니다.

3. epoll과 nginx 상호작용

    1. epoll이 nginx에 이벤트 전달

        `epoll`은 리눅스에서 많은 수의 파일 디스크립터를 효율적으로 감시하기 위한 인터페이스입니다.
        클라이언트의 요청이 도착하면, *커널은 해당 소켓에 대한 읽기 이벤트를 `epoll`을 통해 NGINX에 전달*합니다.

        `epoll`은 소켓의 이벤트(예: 데이터 수신)를 감지하면, 이를 nginx에 전달합니다.
        구체적인 과정은 다음과 같습니다.

        1. nginx가 `epoll_wait()` 호출:
            nginx는 비동기 I/O로 동작하며, 새로운 이벤트가 발생하기 전까지 `epoll_wait()`로 대기 상태에 있습니다.

        2. 이벤트 발생 시 epoll이 알림:
            수신 버퍼에 데이터가 도착하면 epoll은 그 이벤트를 nginx 프로세스에 알립니다.
            epoll은 발생한 이벤트들만 nginx로 전달하므로, nginx는 수천 개의 소켓 중에서 데이터가 수신된 소켓만 처리합니다.

    2. nginx의 처리 방식

        이벤트가 발생하면, nginx는 이를 처리하기 위해 대기 중인 작업 큐에 해당 소켓 작업을 추가합니다.
        nginx는 논블로킹 방식으로 동작하며, 하나의 요청을 처리하는 동안 다른 작업을 막지 않습니다.
        이렇게 비동기적으로 여러 연결을 처리하는 방식 덕분에 nginx는 매우 효율적으로 대규모 트래픽을 처리할 수 있습니다.

        각 이벤트가 발생할 때 NGINX는 소켓에 데이터를 읽거나 쓰는 작업을 수행합니다.
        이 작업은 매우 빠르게 이루어지며, 메모리에서 몇 사이클의 CPU 작업으로 데이터를 전송하거나 수신하는 정도의 작업만 필요할 수 있습니다.

    `epoll`은 리눅스 커널에서 다수의 파일 디스크립터(소켓 포함)를 효율적으로 감시하는 비동기 이벤트 감시 메커니즘입니다.
    특히 nginx와 같은 서버 애플리케이션이 대규모 연결을 처리할 수 있게 도와줍니다.

    - epoll의 구조

        `epoll`은 기존의 `select()`나 `poll()`과 비교해 매우 효율적인 비동기 감시 메커니즘입니다.
        수천~수만 개의 파일 디스크립터를 감시할 때도 성능을 크게 저하시키지 않고 처리할 수 있습니다.
        이는 다음과 같은 구조 덕분입니다.

        - 파일 디스크립터 등록:
            먼저 서버 애플리케이션(예: nginx)은 `epoll_ctl()` 시스템 호출을 사용해 파일 디스크립터(소켓 등)를 epoll에 등록합니다.
            등록된 소켓은 epoll이 감시하는 대상이 됩니다.

        - 이벤트 기반 감시:
            `epoll`은 이벤트 기반으로 동작합니다.
            즉, 소켓에 변화가 생겼을 때(데이터가 수신되었거나, 오류가 발생했거나 등)만 이벤트가 발생하여, 이를 epoll이 감지합니다.
            이 방식은 `poll()`처럼 매번 모든 파일 디스크립터를 반복문으로 순차 확인하는 방식과 달리, 변화가 있는 디스크립터만 처리하므로 매우 효율적입니다.

    - epoll이 수천 개의 소켓을 감시하는 방법

        `epoll`은 커널의 파일 디스크립터 테이블과 연동되어 효율적으로 파일 디스크립터의 상태를 감시합니다.
        특히 O(1)의 시간 복잡도로 이벤트를 감지하는 방식으로 설계되어, 수천~수만 개의 소켓을 효율적으로 감시할 수 있습니다.
        구체적인 방식은 다음과 같습니다.

        1. 소켓에 변화가 생기면 이벤트가 발생: 각 소켓의 수신 버퍼에 데이터가 도착하면 커널은 이 변화를 감지하고, 이벤트 발생을 기록합니다.
        2. 이벤트 대기 상태: `epoll_wait()`는 소켓에 변화가 생길 때까지 대기 상태에 있으며, 데이터가 도착하거나 오류가 발생할 경우 이벤트가 발생합니다.
        3. 변화가 있는 소켓만 처리: epoll은 변화가 감지된 파일 디스크립터들만을 반환하므로, 매번 수천 개의 소켓을 확인하지 않고, 변화가 발생한 디스크립터만 처리합니다.

    - 버퍼의 변화 감지

        epoll이 감시하는 것은 소켓의 수신 버퍼에서 일어나는 변화입니다.
        즉, 수신 버퍼에 새로운 데이터가 도착하거나, 에러가 발생하는 등의 이벤트를 감지합니다.

        - 수신 버퍼의 데이터 도착:
            클라이언트로부터 패킷이 도착하면 커널은 TCP 세그먼트를 재조립하여 소켓의 수신 버퍼에 데이터를 적재합니다.
            이때 커널은 소켓에 읽을 수 있는 데이터가 있다는 이벤트를 epoll에 알립니다.

        - 수신 버퍼의 크기 변화:
            수신 버퍼의 크기가 변동되면(데이터가 추가되거나, 읽혔을 때) 이벤트가 발생하며, epoll이 이를 감지하여 nginx 같은 애플리케이션으로 전달합니다.

4. nginx가 CPU를 선점할 수 있는 방법

    `nginx`와 같은 서버 애플리케이션은 이벤트 기반 아키텍처로 설계되어, CPU 리소스를 효율적으로 사용합니다.

    1. 이벤트 기반 처리 방식

        nginx는 이벤트 기반 아키텍처로 동작합니다.
        이는 폴링 방식과 다르게, 이벤트가 발생했을 때만 작업을 처리하므로 CPU 리소스를 절약할 수 있습니다.

        1. CPU 대기 상태:
            nginx는 `epoll_wait()`로 대기 상태에 있다가, 이벤트가 발생하면 이를 처리하기 위해 활성화됩니다.

        2. CPU 선점:
            이벤트가 발생하면, 커널은 nginx 프로세스에 이벤트를 전달합니다.
            이때 nginx는 CPU를 선점하여 해당 이벤트를 처리합니다.
            이후 작업이 완료되면 다시 `epoll_wait()` 상태로 돌아가 대기합니다.

    2. 스레드 또는 프로세스 풀

        nginx는 멀티스레드 또는 멀티프로세스 방식으로 동작할 수 있습니다.
        이 경우, 각 스레드 또는 프로세스가 별도의 이벤트를 처리하게 되며, 각각의 스레드 또는 프로세스는 비동기적으로 작업을 처리합니다.
        이는 대규모 연결에서 매우 효율적입니다.

        nginx는 각 연결을 별도의 비동기 작업으로 처리하며, CPU를 효율적으로 사용하여 각 연결을 관리합니다.
        이를 통해 nginx는 많은 클라이언트의 요청을 동시에 처리할 수 있게 됩니다.

<!-- /curriculum-chunk -->

### 동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우

#### 원문: 동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우

<!-- curriculum-chunk: sha256=2c1a02421e0d534b966ab604c5e02388f2793d2a5108d0b48362f368be1d9488 major=concurrency-async-io mid=OS I/O multiplexing sub=동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우 sources=interview_questions2.md:542-701, interviews2.md:542-701 -->

> Source: `interview_questions2.md:542-701`
> Classification reason: io multiplexing
> Duplicate source aliases: `interview_questions2.md:542-701, interviews2.md:542-701`

##### 동시 요청이 많아 소켓이 몇 만 개 이상 필요한 상황에서 File Descriptor 설정을 수정하지 않은 경우

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

<!-- /curriculum-chunk -->

### 소켓, epoll, nginx, 과다한 요청

#### 원문: 소켓, epoll, nginx, 과다한 요청

<!-- curriculum-chunk: sha256=0190f89b047fc24a65efa059996d2d9fddd431e24923ec23c1ccbf223c3bc117 major=concurrency-async-io mid=OS I/O multiplexing sub=소켓, epoll, nginx, 과다한 요청 sources=interview_questions.md:7294-7330, interviews.md:7242-7278 -->

> Source: `interview_questions.md:7294-7330`
> Classification reason: io multiplexing
> Duplicate source aliases: `interview_questions.md:7294-7330, interviews.md:7242-7278`

##### 소켓, epoll, nginx, 과다한 요청

1. 멀티플렉싱(Multiplexing):

    멀티플렉싱은 여러 개의 입력 신호나 데이터 스트림을 하나의 채널을 통해 전송하는 기술입니다. I/O 멀티플렉싱의 경우:

    - 하나의 프로세스가 여러 I/O 작업을 동시에 모니터링할 수 있게 합니다.
    - 블로킹 없이 여러 소켓이나 파일 디스크립터의 상태를 효율적으로 관찰할 수 있습니다.

2. epoll과 멀티플렉싱 메커니즘의 동작:

    `epoll`은 Linux 커널의 I/O 이벤트 통지 기능입니다. 작동 방식은 다음과 같습니다:

    - NGINX는 시작 시 `epoll` 인스턴스를 생성하고, 관심 있는 파일 디스크립터(소켓 포함)를 등록합니다.
    - 커널은 이 파일 디스크립터들의 상태를 모니터링합니다.
    - 데이터가 도착하면 커널은 이를 감지하고, `epoll` 인스턴스에 이벤트를 기록합니다.
    - NGINX는 주기적으로 또는 필요할 때 `epoll_wait()` 시스템 콜을 호출하여 이벤트를 확인합니다.
    - 이벤트가 있으면 NGINX는 해당 이벤트에 대응하는 처리를 수행합니다.

    이 메커니즘을 통해 NGINX는 효율적으로 여러 연결을 동시에 관리할 수 있습니다.

3. 대량의 동시 연결 처리:

   10만 개의 동시 연결을 처리하는 방법은 다음과 같습니다:

   - 포트 제한 극복: 서버 측에서는 65535개의 포트 제한이 클라이언트-서버 쌍에 적용됩니다. 즉, (클라이언트 IP, 클라이언트 Port, 서버 IP, 서버 Port)의 고유한 조합으로 식별됩니다. 따라서 다른 클라이언트로부터의 연결이라면 같은 서버 포트를 재사용할 수 있습니다.

   - NGINX의 이벤트 기반 모델: NGINX는 비동기, 이벤트 기반 아키텍처를 사용합니다. 이를 통해 적은 수의 워커 프로세스로 많은 연결을 처리할 수 있습니다.

   - epoll의 확장성: epoll은 수십만 개의 파일 디스크립터를 효율적으로 모니터링할 수 있도록 설계되었습니다.

   - 커널 튜닝: net.ipv4.ip_local_port_range, net.core.somaxconn 등의 커널 파라미터를 조정하여 더 많은 동시 연결을 지원할 수 있습니다.

   - 연결 풀링: NGINX는 업스트림 서버(예: Spring 애플리케이션)와의 연결을 풀링하여 재사용함으로써 효율성을 높입니다.

이러한 기술과 최적화를 통해 현대적인 웹 서버는 수십만, 심지어 수백만 개의 동시 연결을 처리할 수 있습니다. 물론 하드웨어 리소스, 네트워크 대역폭 등의 물리적 제한도 고려해야 합니다.

<!-- /curriculum-chunk -->

## 공유 상태와 경쟁 조건

### 경쟁 조건 (Race Condition)

#### 원문: 경쟁 조건 (Race Condition)

<!-- curriculum-chunk: sha256=1be2b1764a1b77373e661d7765db5573e62ab2cc4fbb4ab1d92f3783d6e69161 major=concurrency-async-io mid=공유 상태와 경쟁 조건 sub=경쟁 조건 (Race Condition) sources=interview_questions.md:635-1098, interviews.md:635-1098 -->

> Source: `interview_questions.md:635-1098`
> Classification reason: race condition
> Duplicate source aliases: `interview_questions.md:635-1098, interviews.md:635-1098`

##### 경쟁 조건 (Race Condition)

경쟁 조건(Race Condition)이란 멀티스레드 환경에서 여러 스레드가 공유 자원에 동시 접근할 때, 실행 순서에 따라 예측할 수 없는 결과가 발생하는 문제를 의미합니다. 즉, 코드의 실행 순서가 의도한 대로 보장되지 않아서 논리적 오류가 발생하는 상황을 의미합니다.

경쟁 조건은 다음과 같은 상황에서 발생합니다:

1. 타이밍 이슈(Timing Issues)와 인터리빙 실행

    다중 스레드나 프로세스가 공유 자원을 사용할 때 실행 순서에 따라 결과가 달라지는 상황을 의미합니다.
    특히 비원자적 연산을 통해 공유 자원에 접근할 때 발생합니다.
    인터리빙(interleaving) 실행으로 인해 예측할 수 없는 결과가 나타나며, 경쟁 상태가 발생하게 됩니다.

    인터리빙(Interleaving)이란 멀티스레드 환경에서 여러 스레드가 동시에 실행될 때, 각 스레드의 명령어가 번갈아(interleave) 실행되는 방식을 의미합니다.
    단일 스레드에서는 순차적으로 실행되지만, 멀티스레드에서는 OS의 스케줄러가 실행 순서를 결정하므로 실행 순서가 달라질 수 있습니다. 그 결과, 공유 자원에 대한 접근 순서가 변하면서 예측할 수 없는 결과가 발생할 수 있습니다.

2. 메모리 일관성 이슈 (Memory Consistency Issues)

    현대의 멀티코어 시스템에서는 각 코어가 로컬 캐시를 가지고 있기 때문에, 캐시와 메인 메모리 간 동기화 문제가 발생할 수 있습니다.
    또한, 메모리 재배치(memory reordering)는 명령어 실행 순서를 변경하여 메모리 가시성 문제를 유발할 수 있습니다.
    이런 경우, 스레드가 업데이트된 데이터를 보지 못하거나, 부분적으로 업데이트된 상태의 데이터를 읽게 되는 문제가 생깁니다.

    메모리 재배치는 CPU와 컴파일러가 최적화를 위해 명령어의 순서를 변경할 수 있는 상황을 말합니다.
    이로 인해 프로그램의 논리적인 실행 순서와 다르게 명령어를 실행할 수 있습니다.
    - 성능 최적화: CPU는 연산 속도를 극대화하기 위해 병렬 실행 및 명령어 재배치를 수행.
    - 파이프라이닝(Pipelining): 명령어를 병렬로 실행하여 성능을 향상.

    ```java
    class ReorderingExample {
        int a = 0, b = 0;
        boolean flag = false;
        // volatile boolean flag = false; //  `volatile` 사용 시, 이전 명령어가 완료된 후 실행됨

        public void writer() {
            a = 1;  // (1)
            flag = true; // (2) CPU가 실행 최적화를 위해 `flag = true;` 를 먼저 실행할 수도 있음.
        }

        public void reader() {
            if (flag) {
                b = a * 2; // (3) 그러면 `b = a * 2;` 가 실행될 때 a의 값이 여전히 0일 수도 있음.
            }
        }
    }
    ```

    이로 인해 동기화된 코드가 아닌 경우, 데이터의 가시성에 문제가 생깁니다.
    메모리 가시성이란 "한 스레드에서 변경한 데이터가 다른 스레드에서 즉시 볼 수 있는가?"에 대한 문제입니다.
    현대 CPU는 캐시 계층 구조를 사용하므로, 각 코어의 캐시와 메인 메모리 간 데이터 동기화 문제가 발생할 수 있습니다.

    ```java
    class VisibilityExample {
        // CPU가 flag를 로컬 캐시에 저장할 경우, 다른 스레드는 이 값을 즉시 확인하지 못할 수 있음.
        boolean flag = false;
        // volatile boolean flag = false; // flag 값이 변경되면 CPU 캐시가 즉시 동기화되어, 모든 스레드가 최신 값을 볼 수 있음.

        public void writer() {
            // 즉, writer()가 flag = true;를 실행했어도,
            // reader() 스레드는 여전히 flag = false로 보일 수 있음.
            flag = true;
        }

        public void reader() {
            while (!flag) {
                // flag가 true가 될 때까지 대기 (busy-waiting)
            }
            System.out.println("Flag is true!");
        }
    }
    ```

    또한, 캐시 일관성 문제는 캐시를 사용하는 모든 프로세서 코어들이 동일한 데이터를 보장하지 못할 때 발생할 수 있습니다.

```java
public class Counter {
    private int count = 0;

    // 잘못된 구현
    public void increment() {
        // 다음 세 단계가 원자적이지 않음
        int temp = count;      // READ
        temp = temp + 1;       // MODIFY
        count = temp;          // WRITE
    }
}

/*
스레드 1과 2가 동시에 실행할 때 가능한 시나리오:

시작: count = 0

스레드 1: READ count = 0
스레드 2: READ count = 0
스레드 1: MODIFY temp = 1
스레드 2: MODIFY temp = 1
스레드 1: WRITE count = 1
스레드 2: WRITE count = 1

최종: count = 1 (기대값: 2)
*/
```

다양한 레벨에서 경쟁 조건이 발생할 수 있습니다.

1. CPU 레벨

    ```nasm
    ; x86 어셈블리에서의 경쟁 조건
    mov eax, [count]   ; load
    inc eax            ; increment
    mov [count], eax   ; store

    ; 원자적 연산 사용
    lock inc dword [count]  ; atomic increment
    ```

    이 경우 경쟁 조건은 비원자적 연산에서 비롯됩니다.
    `mov`, `inc`, `mov` 연산은 각각 별개의 명령어로 CPU에서 실행되므로,
    이들 사이에 다른 스레드가 count 값을 변경할 수 있습니다.

    `lock` 프리픽스를 사용하여 `inc` 명령을 원자적으로 만들어 경쟁 상태를 방지할 수 있습니다.

    추가적으로, 원자적 연산을 사용할 때 메모리 장벽(memory barrier)은 명령어 재배치나 캐시 일관성 문제를 방지하는 데 도움을 줍니다.

2. 메모리 레벨

    ```c
    // Memory Barrier 없는 경우
    int ready = 0;
    int data = 0;

    // Thread 1
    data = 42;    // Store
    ready = 1;    // Store

    // Thread 2
    while (!ready) {}  // Load
    use(data);        // Load

    // Memory Barrier 사용
    atomic_store(&data, 42);
    atomic_store(&ready, 1);
    ```

    메모리 장벽이 없는 경우, CPU는 명령어 최적화를 위해 명령어 순서를 변경할 수 있습니다(메모리 재배치).
    `ready` 변수가 1로 설정되기 전에 `data`가 먼저 사용될 수 있습니다.
    이를 방지하기 위해 메모리 장벽(memory barrier) 또는 원자적 연산을 사용합니다.
    `atomic_store`는 이 연산이 완료되기 전까지 메모리 장벽을 삽입하여, 다른 스레드가 이를 올바르게 읽을 수 있게 만듭니다.

    `atomic_store`는 메모리 일관성을 보장하는 Happens-Before 관계를 설정하여,
    `ready = 1`이 이루어진 후에만 `data` 값이 읽히도록 보장합니다.

3. 파일 시스템 레벨

    ```python
    # 파일 시스템 레벨 경쟁 조건
    def update_config(key, value):
        config = read_config_file()
        config[key] = value
        write_config_file(config)

    # 해결: 파일 락 사용
    def safe_update_config(key, value):
        with FileLock("config.lock"):
            update_config(key, value)
    ```

    파일 시스템에서도 경쟁 조건이 발생할 수 있습니다.
    여러 프로세스가 동시에 동일한 파일에 접근하여 내용을 수정하려 하면, 데이터 손실이나 불일치가 발생할 수 있습니다.
    이를 방지하기 위해 파일 락(file lock)을 사용하여, 파일이 안전하게 수정되도록 합니다.
    FileLock을 사용하여 파일에 대한 독점적인 접근을 보장할 수 있습니다.

    추가로, 파일 잠금이 적절히 해제되지 않을 경우 발생할 수 있는 교착 상태(deadlock) 문제도 고려해야 합니다.
    이를 해결하기 위해 타임아웃 설정을 추가할 수도 있습니다.

이를 해결 메커니즘은 다음과 같습니다:

1. 원자성 보장

    ```java
    public class SafeCounter {
        private AtomicInteger count = new AtomicInteger(0);

        public void increment() {
            count.incrementAndGet(); // 원자적 연산
        }
    }
    ```

2. 스핀락

    ```java
    public class SpinLockExample {
        private AtomicBoolean locked = new AtomicBoolean(false);

        public void lock() {
            while (!locked.compareAndSet(false, true)) {
                // 계속 CPU를 사용하면서 확인
                // 다른 스레드에게 CPU를 양보하지 않음
            }
        }
    }
    ```

    - CPU를 계속 사용 (busy waiting)
    - 락이 곧 해제될 것으로 예상될 때 유용
    - 컨텍스트 스위칭 비용 회피 가능
    - 하지만 CPU 자원 낭비가 심함

3. synchronized

    `synchronized`는 자바의 기본적인 동기화 기법으로, 여러 스레드가 공유 자원에 동시 접근할 때 경쟁 조건(Race Condition)을 방지하는 데 사용됩니다.
    `synchronized` 블록이나 메서드 내부에 진입한 스레드는 해당 객체의 락을 획득해야만 실행 가능합니다.
    객체(또는 클래스) 단위로 Monitor Lock(모니터 락)을 사용하여 하나의 스레드만 특정 코드 블록을 실행할 수 있도록 보장합니다.

    `synchronized` 블럭으로 동기화를 하면 자동적으로 lock이 잠기고 풀립니다.
    `synchronized` 블럭 내에서 예외가 발생해도 lock은 자동적으로 풀립니다.
    그러나 같은 메소드 내에서만 lock을 걸 수 있다는 제약이 존재합니다.
    WAITING 상태인 스레드는 interrupt가 불가능합니다.

    1. Biased Locking (편향 락): 락이 자주 사용되는 단일 스레드에 편향됨

        JVM은 하나의 스레드가 반복적으로 같은 객체를 락을 걸고 해제하는 패턴이 많다고 가정하고, 락을 "편향(Biased)"되도록 설정합니다.
        락 획득 비용을 최소화하기 위해, 객체의 mark word에 "어느 스레드가 이 락을 사용하는지" 기록합니다.

        ```java
        class Example {
            // 대부분의 경우 같은 스레드가 반복적으로 락을 획득
            private synchronized void method() {
                // 최초 락 획득시 스레드 ID를 객체 헤더에 기록
                // 이후 같은 스레드는 실제 락 획득 없이 진입
                // ...
            }
        }
        ```

        편향 락의 동작 과정:
        1. 처음 락을 획득할 때, JVM 내부적으로는 객체의 헤더(Object Header)에 있는 mark word에 스레드 ID를 기록하여 락을 관리.
        2. 같은 스레드가 다시 해당 락을 획득하려고 하면, 추가적인 동기화 없이 바로 실행 가능 (거의 비용 없음).
        3. 다른 스레드가 이 락을 획득하려고 하면, 편향 락을 해제하고 경량 락으로 전환.

        편향 락이 해제되는 조건:
        - 다른 스레드가 같은 락을 획득하려고 시도할 때
        - `System.identityHashCode(obj)` 호출 시 (mark word를 해시코드 저장용으로 변경해야 함)
        - JVM 내부에서 GC(Garbage Collection)가 편향 락을 해제해야 할 필요가 있을 때

        만약 두 개의 메서드중 하나에만 synchronized가 존재할 경우 결과는 순차적으로 실행되지 않습니다.
        synchronized 키워드가 블럭 or 메소드에 있다고 해서 무조건 해당 쓰레드가 계속 점유를 하는 것은 아닙니다.
        synchronized 메소드에 들어갔더라도 해당 메소드에 *여러 쓰레드가 동시에 접근이 불가능*할 뿐이지, 문맥 교환은 발생합니다.

        그리고 Lock의 범위는 객체 단위라는 것을 하나 더 알 수 있습니다.
        즉, 객체당 Lock을 하나만 가질 수 있습니다.
        그렇기 때문에 synchronized가 메소드 둘 다 붙어있다면 하나의 메소드가 이미 lock을 쥐고 있기 때문에 나머지 메소드는 기다리게 되는 것입니다.

    2. static 메소드에서 synchronized lock을 사용하면

        ```java
        class Example {
            // 대부분의 경우 같은 스레드가 반복적으로 락을 획득
            private static synchronized void method1() {
                // 최초 락 획득시 스레드 ID를 객체 헤더에 기록
                // 이후 같은 스레드는 실제 락 획득 없이 진입
                // ...
            }

            // 대부분의 경우 같은 스레드가 반복적으로 락을 획득
            private static synchronized void method2() {
                // 최초 락 획득시 스레드 ID를 객체 헤더에 기록
                // 이후 같은 스레드는 실제 락 획득 없이 진입
                // ...
            }
        }
        ```

        `method1`와 `method2`가 번갈아 가면서 실행됩니다.
        `static` 메소드는 객체의 것이 아니라 클래스 메소드 입니다.
        그렇기 때문에 Lock을 클래스도 하나를 가지고, 객체들 마다 하나씩 가지게 되는 것입니다.
        그래서 둘 다 락을 가지고 메소드에 접근을 할 수 있는 것입니다.

    3. Thin Lock (경량 락)

        경량 락은 CAS(Compare-And-Swap) 연산을 활용하여 락 경쟁을 최소화합니다.

        경량 락의 핵심 원리:
        1. 스레드가 락을 시도하면, 객체의 mark word를 "락 레코드(Lock Record) 포인터"로 변경.
        2. 다른 스레드가 동시에 락을 시도하면 CAS 연산을 통해 락 획득 여부를 결정.
        3. 스핀 락(Spin Lock)을 사용하여 일부 경쟁 상황에서 스레드 대기 시간을 줄임.

        ```plaintext
        객체 헤더 (mark word) 구조:

        일반 상태:
        [해시코드 | 나이 | 락 정보 | 01]  // 마지막 비트: 일반 상태

        편향 락 상태:
        [스레드 ID | 에포크 | 나이 | 01]  // 특정 스레드에 편향

        경량 락 상태:
        [락 레코드 포인터 | 00]  // 스택의 락 레코드 가리킴

        중량 락 상태:
        [모니터 포인터 | 10]  // 힙의 모니터 객체 가리킴
        ```

    4. Fat Lock (중량 락)

        ```java
        // JVM 내부 모니터 구현 (의사 코드)
        class ObjectMonitor {
            void* owner;                  // 락 보유 스레드
            intptr_t recursions;         // 재진입 카운트
            ObjectWaiter* waitSet;       // wait() 중인 스레드들
            ObjectWaiter* entryList;     // 진입 대기 스레드들
            int contentions;            // 경쟁 카운트

            void enter(Thread* thread) {
                if (owner == thread) {
                    recursions++;            // 재진입
                    return;
                }

                if (contentions++ > 0) {    // 경쟁 발생
                    enqueue_and_wait(thread);
                } else {
                    set_owner(thread);      // 락 획득
                }
            }
        }
        ```

    5. 락 단계 상승(Lock Inflation)

        ```plaintext
        편향 락 → 경량 락 → 중량 락
        ```

        1. 편향 락 해제 조건:
            - 다른 스레드가 락 획득 시도
            - `System.identityHashCode()` 호출
            - 편향 락 해제 요청

        2. 경량 락 → 중량 락 전환 조건:
            - 락 획득 실패 후 스핀 횟수 초과
            - 너무 많은 스레드가 대기
            - 특정 시간 이상 락 획득 실패

    실제 예시로 확인해보면:

    ```java
    public class LockEscalationExample {
        private static final int THREADS = 4;
        private final Object lock = new Object();
        private int counter = 0;

        public void increment() {
            synchronized(lock) {  // 처음엔 편향 락으로 시작
                counter++;       // 경쟁 발생시 경량 락으로 전환
            }                    // 심한 경쟁시 중량 락으로 전환
        }

        public static void main(String[] args) {
            LockEscalationExample example = new LockEscalationExample();

            // 단일 스레드: 편향 락 사용
            example.increment();

            // 여러 스레드: 락 단계 상승 발생
            for (int i = 0; i < THREADS; i++) {
                new Thread(() -> {
                    for (int j = 0; j < 1000000; j++) {
                        example.increment();
                    }
                }).start();
            }
        }
    }
    ```

    이 락 단계 상승 메커니즘 때문에:
    - 경쟁이 없는 경우 거의 무비용 동기화 (편향 락)
    - 약한 경쟁의 경우 스핀으로 해결 (경량 락)
    - 심한 경쟁의 경우 OS 수준 동기화 (중량 락)

    이렇게 상황에 따라 최적의 성능을 발휘할 수 있습니다.

    이 메커니즘을 이해하면 synchronized 사용 시 더 나은 성능을 위한 최적화가 가능합니다:
    - 가능한 한 락의 범위를 좁게 유지
    - 불필요한 락 경쟁 피하기
    - 동일 스레드의 반복적인 락 사용 패턴 활용

4. ReentrantLock 락 기반 동기화

    ReentrantLock에서 Reentrant라는 이름은 재진입 가능성(reentrancy)을 의미합니다.
    재진입 가능성은 다음을 의미합니다:
    - 동일한 스레드가 이미 소유한 락을 다시 요청할 수 있고, 이를 성공적으로 획득할 수 있음.
    - 락이 여러 번 획득되었더라도, 락을 해제(unlock)할 때는 획득한 횟수만큼 해제해야 완전히 락이 풀림.

    이는 동일한 스레드가 이미 소유한 락을 다시 획득하려고 시도할 때, 교착 상태 없이 성공적으로 획득할 수 있음을 보장합니다.

    ReentrantLock이 경쟁 조건을 방지하는 원리:
    1. CAS(Compare-And-Swap) 기반의 락 획득
        - ReentrantLock은 내부적으로 CAS 연산을 활용하여 락을 관리.
        - CAS를 통해 다른 스레드가 현재 락을 획득했는지 확인하고, 획득하지 않았다면 현재 스레드가 락을 획득.

    2. 락이 해제될 때만 다른 스레드가 락을 획득 가능
        - 락을 보유한 스레드가 `lock.unlock()`을 호출해야만 다른 스레드가 실행 가능.

    3. 재진입 가능
        - 같은 스레드가 여러 번 `lock.lock()`을 호출해도, Deadlock이 발생하지 않음.
        - 같은 스레드가 여러 번 `lock()`을 호출하면, 락 횟수(recursion count)가 증가하며, 같은 횟수만큼 `unlock()`을 호출해야 락이 해제됨.
    synchronized와 달리 수동으로 lock을 잠그고 해제해야 합니다.
    `lockInterruptably()` 함수를 통해 WAITING 상태의 스레드를 interrupt할 수 있습니다.

    ```java
    public class ReentrantLockExample {
        // 재 진입할 수 있는(Reentrant) 이라는 단어가 붙어 있는 이유는
        // wait(), notify()와 같이 특정 조건에서 lock을 풀고
        // 나중에 다시 lock을 얻고 임계영역으로 들어와서 이후의 작업을 수행할 수 있기 때문입니다.
        //
        // 생성자의 매개변수를 true를 주면 lock이 풀렸을 때 가장 오래 기다린 쓰레드가 lock을 획득할 수 있게 공정(fair)하게 처리합니다.
        // 하지만 공정하게 처리하려면 어떤  쓰레드가 가장 오래 기다렸는지 확인하는 과정을 거칠 수 밖에 없으므로 성능은 떨어질 수 밖에 없습니다.
        private final ReentrantLock lock = new ReentrantLock(true); // true: 공정성 활성화

        public void process() {
            // 인터럽트 가능한 락 획득
            try {
                // 락 획득 시도 방법 2: 타임아웃
                if (lock.tryLock(1, TimeUnit.SECONDS)) {
                    try {
                        // 임계 영역
                    } finally {
                        lock.unlock();
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }

        // 조건 변수 사용
        private final Condition notFull = lock.newCondition();
        private final Condition notEmpty = lock.newCondition();

        public void produce() throws InterruptedException {
            // 락 획득 시도 방법 1: 계속 대기
            lock.lock(); // 락을 얻을 때까지 WAITING 상태
            try {
                while (isFull()) {
                    notFull.await();
                }
                // 생산
                notEmpty.signal();
            } finally {
                lock.unlock();
            }
        }
    }
    ```

    - 더 유연한 대기 정책 제공
    - lock() 사용시 `WAITING` 상태로 전환 (CPU 사용 안함)
    - tryLock() 사용시 즉시 리턴 가능
    - 인터럽트 가능
    - 공정성 정책 설정 가능

<!-- /curriculum-chunk -->

### 멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지

#### 원문: 멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지

<!-- curriculum-chunk: sha256=8c0646c3c8f7ac0313a8d98d13373b16f8e4c97169740009ab0ec6bad7d91a0d major=concurrency-async-io mid=공유 상태와 경쟁 조건 sub=멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지 sources=interview_questions2.md:102-249, interviews2.md:102-249 -->

> Source: `interview_questions2.md:102-249`
> Classification reason: shared state
> Duplicate source aliases: `interview_questions2.md:102-249, interviews2.md:102-249`

##### 멀티스레딩에서 어떻게 읽기 전용 속성을 동시에 사용할 수 있는지

> 멀티스레딩 애플리케이션에서 공유 자원의 쓰기는 동기화가 필요한 반면, 읽기만 이뤄지는 속성에 대해서는 동기화가 필요없습니다. 가령 kotlin에서 http 클라이언트를 한번 초기화하고 여러 스레드에서 재사용하는 경우가 많은데요. jvm이 실행되고 클래스가 로드될 겁니다. 그리고 jvm은 컴파일된 .class 바이트코드로부터 클래스로드를 하게 됩니다. 그러면 어떻게 여러 스레드가 하나의 http 클라이언트 인스턴스를 공유하며 여러 요청을 처리할 수 있는지 cpu, 메모리, os, jvm 관점을 아울러서 설명해보세요.
> jvm 런타임이 프로세스에 실행되면 해당 jvm 내에서 바이트코드는 JIT로 컴파일되어 이때 실제로 기계어로 컴파일되고 실행됩니다. 그리고 http 클라이언트는 힙에 하나 존재한다고 가정하겠습니다. 그리고 이 하나의 인스턴스를 3개의 스레드들이 각각 google.com, naver.com, facebook.com에 GET 요청을 한다고 가정하겠습니다. 해당 애플리케이션이 실행되는 서버의 코어는 4개고 메모리도 충분합니다. 그러면 스레드 3개가 3개의 코어에서 실제로 병렬로 실행될 겁니다. jvm 힙 메모리에 인스턴스는 하나인데 어떻게 세 개의 코어에서 세 개의 스레드가 동시에 하나의 인스턴스를 사용하요 http 요청을 하고 응답을 처리할 수 있는지 설명해보세요.

###### JVM에서 클래스 로드와 초기화

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

###### CPU 관점

여러 스레드는 각 CPU 코어에서 병렬로 실행됩니다.

예를 들어 세 개의 스레드가 각각 http 클라이언트를 사용한다고 했을 때, 각자 다른 CPU 코어에서 실행됩니다.
각 코어는 HTTP 클라이언트 객체의 메모리를 참조하며, 필요한 데이터를 CPU 캐시에 로드합니다.

CPU는 컨텍스트 스위칭을 통해 스레드의 실행 상태를 저장하고 복원하며, 이를 기반으로 스케줄링합니다.

HTTP 클라이언트 인스턴스는 읽기 전용 속성을 가진 경우가 많습니다.
즉, 스레드가 이 인스턴스를 공유할 때, CPU 캐시에 올려진 객체를 각 스레드가 읽습니다.

CPU는 MESI(Modified, Exclusive, Shared, Invalid) 프로토콜을 사용하여 공유 메모리의 일관성을 유지합니다.
따라서 여러 스레드가 동일한 HTTP 클라이언트를 읽는 경우 캐시 일관성이 보장됩니다.
가령 세 개의 스레드가 동일한 객체를 읽는 경우, 읽기 작업은 캐시된 데이터를 사용해 빠르게 수행됩니다.

###### 메모리 관점

`HttpClient` 인스턴스는 힙 메모리에 저장됩니다.
이는 JVM에서 관리하며, GC(Garbage Collector)의 대상이 됩니다.

여러 스레드가 힙 메모리에 있는 동일한 HTTP 클라이언트 인스턴스를 참조합니다.
각 스레드 스택(Stack)은 인스턴스의 참조를 유지합니다.

읽기 작업은 동기화가 필요 없으므로, 메모리 락이나 다른 동기화 메커니즘이 없어도 안전하게 작동합니다.

###### OS 관점

OS는 각 스레드가 독립적으로 실행되도록 스케줄링합니다.
4개의 CPU 코어 중 3개가 스레드를 처리하며, 각 스레드는 고유한 컨텍스트를 유지합니다.

네트워크 요청은 OS의 네트워크 스택을 통해 처리됩니다.
OS는 소켓을 통해 데이터를 송수신하며, 각 스레드가 요청에 대한 응답을 독립적으로 처리하도록 보장합니다.

###### 스레드 관리

OS는 각 JVM 스레드를 시스템 스레드로 매핑합니다.
JVM의 스레드는 POSIX 스레드(pthread)와 같은 시스템 스레드로 구현됩니다.

OS 스케줄러는 각 스레드를 CPU 코어에 할당하고, 컨텍스트 스위칭을 통해 병렬성을 제공합니다.

###### 메모리 매핑

OS는 가상 메모리를 통해 JVM의 힙 메모리를 관리합니다.
HTTP 클라이언트 인스턴스는 물리 메모리에 매핑되어 여러 스레드에서 공유됩니다.

OS는 메모리 접근 충돌을 방지하기 위해 페이지 테이블과 MMU(Memory Management Unit)를 활용하며, 공유 메모리에 대한 동기화는 하드웨어와 CPU 캐시 프로토콜에 의존합니다.

###### JVM 내부 동작

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

###### HTTP 클라이언트의 내부 동작

대부분의 HTTP 클라이언트는 커넥션 풀을 사용합니다.
이는 네트워크 연결을 재사용하여 성능을 최적화하는 메커니즘입니다.

가령 세 개의 스레드는 동일한 클라이언트를 사용하지만, 요청을 처리하는 소켓 연결은 각각 독립적입니다.

커넥션 풀은 동시성을 안전하게 관리하도록 설계되어 있으며, 동기화된 데이터 구조(예: `ConcurrentLinkedQueue`)를 통해 여러 스레드에서의 접근을 처리합니다.

일부 HTTP 클라이언트는 비동기 요청 모델을 지원합니다.
이 경우, 각 요청은 이벤트 루프와 워커 스레드를 통해 병렬로 처리됩니다.
요청과 응답 간의 상태 관리는 비동기 콜백 또는 `CompletableFuture`와 같은 구조로 이루어집니다.

###### 힙 메모리 관리

HTTP 클라이언트는 힙에 위치하며 여러 스레드가 참조합니다.
힙 메모리는 Young Generation과 Old Generation으로 나뉩니다.

HTTP 클라이언트는 보통 싱글턴으로 한 번만 생성되므로 오래 유지되며 Old Generation으로 이동합니다.
이는 GC의 영향을 덜 받게 합니다.

###### 메모리 가시성

JVM은 힙 메모리 접근이 안전하게 이루어지도록 Java Memory Model (JMM) 규칙을 따릅니다.
JMM은 쓰기 작업에서 volatile 또는 synchronized와 같은 메커니즘을 통해 메모리 가시성을 보장합니다.
그러나 HTTP 클라이언트는 주로 읽기 작업이 이루어지므로, 동기화 오버헤드가 필요하지 않습니다.

HTTP 클라이언트 객체는 힙 메모리에 존재하며, 이는 여러 스레드가 공유할 수 있습니다.

`HttpClient`와 같은 정적 변수는 클래스 초기화 과정에서 안전하게 초기화되며, happens-before 관계에 의해 다른 스레드에서 항상 최신 상태로 보입니다.

정적 변수는 JVM의 메서드 영역(Method Area)에서 관리되며, 이는 모든 스레드가 접근할 수 있습니다.

###### 비동기 처리

HTTP 클라이언트는 종종 비동기 요청을 처리하도록 설계됩니다.
예를 들어, 비동기 요청을 처리할 때 이벤트 루프와 워커 스레드를 사용할 수 있습니다.

JVM은 비동기 작업을 위해 ForkJoinPool이나 사용자 정의 Executor를 활용합니다.

###### 실제 요청 처리 흐름

###### 요청 처리

1. 스레드는 HTTP 클라이언트의 메서드를 호출하여 요청을 보냅니다.
2. HTTP 클라이언트는 내부적으로 연결 풀(Connection Pool)을 활용하여 네트워크 연결을 관리합니다.
3. 각 요청은 비동기로 처리되며, 네트워크 I/O는 OS의 커널 네트워크 스택에서 이루어집니다.
4. OS는 요청을 적절한 네트워크 인터페이스로 전달하며, 응답은 다시 JVM으로 반환됩니다.

###### 동시성 처리

HTTP 클라이언트는 내부적으로 스레드 안전성을 보장하는 구조(예: 동기화된 데이터 구조 또는 CAS 연산)를 사용합니다.

요청 간 상태는 분리되어야 하므로, 상태 저장이 필요한 경우 `ThreadLocal` 또는 명시적 동기화를 통해 처리합니다.

<!-- /curriculum-chunk -->

### 자바 쓰레드 동기화

#### 원문: 자바 쓰레드 동기화

<!-- curriculum-chunk: sha256=212bbb821bac41121fd8d7e12be8c6fbbd28f163d24f2a056ad19e0718285fcb major=concurrency-async-io mid=공유 상태와 경쟁 조건 sub=자바 쓰레드 동기화 sources=interview_questions3.md:237-393 -->

> Source: `interview_questions3.md:237-393`
> Classification reason: shared state

##### 자바 쓰레드 동기화

###### 1. 자바에서의 쓰레드와 모니터 개념

1. 모든 동물은 생을 마친다(일반 원칙).
2. 자바의 `synchronized` 블록 또는 메서드는 “모니터(Monitor)”를 기반으로 동작한다(특정 원칙).
3. 따라서 `synchronized(obj)` 구문을 만나면, 쓰레드는 `obj`라는 객체의 모니터를 **획득**해야만 블록 내부 코드를 실행할 수 있다(결론).

먼저, “**모니터**”란, 자바에서 **동기화(synchronization)**를 지원하기 위해 객체마다 존재하는 구조체 혹은 **잠금(락) 기법**을 의미합니다. `synchronized (obj)`라고 하면, 해당 obj에 대한 **모니터 잠금**을 쓰레드가 얻어야만(획득해야만) 코드 실행이 가능합니다.

---

###### 2. 쓰레드 상태(States) 개념

1. 자바 쓰레드는 **Thread.State** 열거형에 의해 여러 상태로 분류된다(일반 원칙).
2. 그 상태는 크게 **RUNNABLE, BLOCKED, WAITING, TIMED_WAITING, TERMINATED** 등이 있다(특정 원칙).
3. 따라서 어떤 특정 상황에서 쓰레드가 동작 중이거나(waiting, blocked), 종료되었는지 등을 구분할 수 있다(결론).

구체적으로:
- **RUNNABLE**: JVM 내부적으로 “실행 가능 상태” 또는 OS 스케줄러가 CPU에 올릴 수 있는 상태.
- **BLOCKED**: `synchronized`를 위해 모니터 잠금을 **획득하려고 시도**하지만, 잠금이 이미 다른 쓰레드에 의해 점유되어 있어서 **대기 중**인 상태.
- **WAITING**: 특정 조건(예: `wait()`, `join()`, `LockSupport.park()`)을 만족할 때까지 **“(notify 같은) 신호”**를 기다리는 상태.
- **TIMED_WAITING**: `wait(timeout)`, `sleep(timeout)` 처럼 유한 시간 동안만 대기하는 상태.
- **TERMINATED**: 쓰레드 실행이 끝난 상태.

---

###### 3. synchronized 블록과 모니터 잠금

###### 3.1 진입 과정

1. 만약 A 쓰레드가 `synchronized(LockObject)` 구문을 실행하려고 시도한다(원인).
2. JVM은 **LockObject**라는 객체의 모니터를 확인한다(작동).
3. 객체 모니터가 사용 중(이미 다른 쓰레드가 갖고 있음)이면, A 쓰레드는 **BLOCKED** 상태로 전환되어 대기한다(결론 1).
4. 만약 모니터가 비어있다면, A 쓰레드는 **LockObject 모니터를 획득**하고 **RUNNABLE** 상태로 블록 내부 코드를 실행한다(결론 2).

즉, `BLOCKED` 상태는 “**모니터 잠금을 얻지 못해서** 대기”인 상태를 의미합니다. `synchronized(obj)`가 끝난 뒤(모니터가 release된 뒤)에야 다시 재시도할 수 있습니다.

###### 3.2 탈출 과정

1. A 쓰레드가 `synchronized(LockObject)` 블록의 끝에 도달하거나, 예외가 발생해 블록을 빠져나가면,
2. 잠금이 풀리고(= 모니터 반납), 대기 중인 다른 쓰레드(만약 있다면)가 모니터를 획득할 기회를 얻는다.

이로써 단일 객체 모니터를 통한 직렬화(One-at-a-time)가 달성됩니다.

---

###### 4. wait()와 notify() 메서드의 메커니즘

###### 4.1 wait()가 하는 일

1. 자바에서 `obj.wait()`는 “**이 객체(obj)에 대한 모니터**”를 사용 중인 쓰레드가 임시로 모니터 **점유권을 포기**하고, 해당 객체의 **wait set**에 들어가 기다리게 하는 메서드이다(일반 원칙).
2. 즉, 쓰레드가 synchronized(obj) 블록 내부에서 `obj.wait()`를 호출하면, 자발적으로 모니터를 놓고(wait/release), JVM 내부의 **wait set**에 들어가 **WAITING** 상태로 전환된다(작동).
3. 그 결과, 모니터가 비워지므로 다른 쓰레드가 이 모니터를 획득할 수 있게 된다(결론).

`wait()`가 호출되어 WAITING 상태가 된 쓰레드는, **notify() / notifyAll()** 등이 호출되기 전까지는 깨어나지 않는다. 깨어난 뒤에는 모니터를 다시 획득해야만(`BLOCKED` 과정을 거칠 수 있음) 원래 코드의 이어지는 부분(`doSomething()`)을 실행한다.

###### 4.2 notify()와 notifyAll()

1. `obj.notify()`는 “**obj의 wait set**에서 대기 중인 쓰레드들 중 한 쓰레드**”를 깨우는(awakening) 역할이다(원칙).
2. 깨워진 쓰레드는 **모니터를 다시 획득하려고 시도**하지만, 그 순간 모니터가 이미 다른 쓰레드에 의해 사용 중이면, `BLOCKED` 상태로 잠시 대기해야 한다(작동).
3. 만약 모니터가 비어있다면, 바로 획득한 뒤 **RUNNABLE** 상태가 되어 `wait()` 호출 직후 코드를 이어서 수행한다(결론).

차이점:
- `notify()`는 한 개 쓰레드만 깨운다.
- `notifyAll()`는 wait set의 **모든 쓰레드**를 깨우지만, 결국 모니터는 한 쓰레드만 먼저 획득할 수 있으므로, 나머지 쓰레드는 `BLOCKED` → `RUNNABLE` 순으로 경합한다.

---

###### 5. wait set의 작동 원리

1. 어떤 객체 LockObject에 대해 `LockObject.wait()`가 호출되면, 해당 쓰레드는 “LockObject의 **wait set**”에 추가된다(원칙).
2. 그동안 이 쓰레드는 모니터를 포기했으므로, **WAITING** 상태가 된다(작동).
3. 다른 쓰레드가 `LockObject.notify()`나 `LockObject.notifyAll()`을 호출하면, wait set에 있던 쓰레드(들)이 깨운 신호를 받고, 모니터 재획득을 시도한다(결론).

만약 **notify**나 **notifyAll**이 절대 호출되지 않는다면(즉, 락 오브젝트에 대한 신호가 없으면), **해당 쓰레드는 영원히 wait set에서 WAITING 상태**로 남을 수 있습니다. 따라서 wait/notify 구문은 반드시 한 쌍으로 설계해야 합니다.

---

###### 6. 쓰레드 상태: BLOCKED vs WAITING

###### 6.1 BLOCKED 상태

1. A 쓰레드가 “**아직 모니터를 획득하지 못했음**”에도 불구하고, `synchronized(obj)` 구문에 진입하려 시도하면, 모니터가 다른 쓰레드에 점유된 경우 BLOCKED 상태가 된다(원칙).
2. 즉, **mutex 락**을 얻지 못해서 대기 중인 상태라고도 할 수 있다(작동).
3. 다른 쓰레드가 모니터를 해제할 때까지 스케줄링 자격이 없다가, 해제된 순간 JVM이 내부적으로 다시 락 획득을 시도해, 성공하면 RUNNABLE이 된다(결론).

###### 6.2 WAITING 상태

1. 쓰레드가 이미 모니터를 확보한 상황에서 `obj.wait()`를 호출하면, 모니터를 자발적으로 반납하고 wait set에 들어가게 된다(원칙).
2. 이는 “**명시적 신호(notify)**”가 오기 전까지는 깨어날 의사가 없음”을 표현하는 것이다(작동).
3. notify() 또는 notifyAll()을 통해 신호가 오면, 쓰레드는 모니터 재획득을 위해 경쟁하다가(그 사이 `BLOCKED`가 될 수도 있음) 성공시 RUNNABLE이 되고, 결국 wait() 호출 직후 코드를 계속 실행한다(결론).

**정리**: BLOCKED는 “모니터를 얻지 못해 문 앞에서 기다림”이고, WAITING은 “모니터를 얻었다가, wait() 호출로 임시 반납 후 notify를 기다리는 상태”입니다.

---

###### 7. notify가 없으면 wait한 스레드는 어떻게 되는가?

1. 만약 어떤 스레드가 `obj.wait()`로 WAITING 상태에 진입했는데, 이후 어떤 스레드도 `obj.notify()`나 `obj.notifyAll()`을 호출하지 않는다면(가정),
2. 그 쓰레드는 영원히 wait set에서 깨어나지 못하고, **영구적으로 WAITING 상태**로 남습니다(결론).
3. 실제 시스템에서도, wait/notify를 잘못 구현하면 데드락(deadlock)이나 **무한 대기**가 발생할 수 있으므로, 반드시 논리적으로 짝을 맞춰야 합니다(교훈).

---

###### 8. JVM 레벨에서의 synchronized 작동 방식

###### 8.1 모니터 엔터/익시트(enter/exit)

1. 자바 바이트코드 상에서 `synchronized`는 `monitorenter`, `monitorexit` 명령어로 번역된다(원칙).
2. 실제 실행 시, JVM은 객체 헤더(mark word)와 모니터 구조를 이용해 락 상태, 소유 스레드, 대기열 등을 관리한다(작동).
3. `monitorenter`는 락 획득에 성공할 때까지 BLOCKED로 대기할 수도 있고, 성공하면 RUNNABLE로 진행한다(결론).
4. 블록 끝, 예외 발생 시 등에서 `monitorexit`가 호출되어 모니터를 해제한다.

###### 8.2 경량화 기법(Biased Locking, Lightweight Locking)

1. HotSpot JVM은 락 오버헤드를 줄이기 위해 **Biased Locking**이나 **Lightweight Locking** 같은 최적화를 한다(원칙).
2. 단일 스레드가 계속해서 같은 락을 잡는다면 편향 모드(biased)로 전환해 락 재획득 비용을 제거(작동).
3. 스레드 간 실질적 경합이 일어나기 전까지, 락 연산이 매우 가볍게 처리된다(결론).

---

###### 9. 실제 동작 시나리오 예시

이제 위의 개념을 토대로, 질문에서 주어진 예시 코드를 분석해봅니다:

```java
synchronized(LockObject) {
    LockObject.wait();
    doSomething();
}
```

1. 이 구문이 실행될 때, 우선 `synchronized(LockObject)`에 진입하려면, 현재 쓰레드는 **LockObject 모니터**를 획득해야 합니다. 이미 다른 쓰레드가 이 모니터를 잡고 있으면, 이 쓰레드는 **BLOCKED 상태**가 되어 모니터 획득을 대기합니다.
2. 만약 성공적으로 모니터를 얻었다면, `LockObject.wait()`를 호출합니다. 이 순간, 쓰레드는 **(a) 모니터를 release**하고, **(b) LockObject의 wait set**에 들어가며, 쓰레드 상태는 **WAITING**으로 바뀝니다.
3. 이제 다른 쓰레드가 `LockObject.notify()` 또는 `notifyAll()`을 호출해주기 전까지는, 이 쓰레드는 깨어나지 못합니다.
4. notify가 도착하면, 대기 중인 쓰레드(혹은 여러 쓰레드) 중 하나(또는 전부)가 **“모니터 재획득”**을 시도합니다. 락을 획득하지 못하면 **BLOCKED**, 획득에 성공하면 **RUNNABLE** 상태로 돌아가 `doSomething()`을 실행합니다.
5. `doSomething()` 실행 후, `synchronized(LockObject)` 블록 끝에 도달해 **monitorexit**가 일어나고, 쓰레드는 모니터를 해제합니다.

---

###### 10. 핵심 요약 및 결론

1. **Java 스레드 관리**에서 `synchronized`는 객체 모니터를 기반으로 임계영역 접근을 직렬화한다. 이때 모니터가 이미 점유 중이면 새로 진입하려는 쓰레드는 BLOCKED가 된다.
2. **wait()**는 모니터를 임의로 반납하고 객체의 wait set에 들어가게 하여 쓰레드를 WAITING 상태로 만든다. 이때 **notify()** 혹은 **notifyAll()**이 호출되기 전까지는 깨어나지 못한다.
3. notify가 없으면 해당 쓰레드는 영구 대기할 수 있으므로, 반드시 올바른 로직으로 wait/notify 쌍을 구성해야 한다.
4. **BLOCKED** 상태는 모니터를 아직 얻지 못해 문 앞에서 기다리는 상태이고, **WAITING** 상태는 이미 모니터를 얻은 뒤 wait()를 호출해 놓은 상태다. 깨어날 때도 모니터를 다시 얻어야 하므로, 일시적으로 BLOCKED로 돌아갈 수 있다.
5. JVM 레벨에서는 `monitorenter`/`monitorexit`, wait set, 락 최적화 등 다양한 내부 메커니즘으로 이 과정을 관리한다.

이처럼 자바의 **동기화 모델**은 모니터 잠금과 wait/notify를 결합하여, 스레드 간 통신(Condition Wait)과 임계영역 보호(Mutex Lock)를 동시에 지원합니다. 이것이 자바가 멀티쓰레드 프로그래밍을 강력하고 안전하게 제공하는 기저가 됩니다.

> **결론**: synchronized, wait, notify를 올바르게 사용하기 위해서는
> 1) 모니터 잠금의 획득/해제 시점
> 2) BLOCKED vs WAITING 상태 구분
> 3) notify/notifyAll을 통한 대기 해제 원리
> 등을 숙지해야 하며, 이는 JVM의 모니터 구조와 자바 스레드 상태 모델을 이해해야 명확히 파악할 수 있습니다.

<!-- /curriculum-chunk -->

## 교착과 진행 보장

### 교착 상태 (Deadlock)

#### 원문: 교착 상태 (Deadlock)

<!-- curriculum-chunk: sha256=1c70931f2f4b3c5973967b55f6d677113322336f6857ee222f40eb43b36dff26 major=concurrency-async-io mid=교착과 진행 보장 sub=교착 상태 (Deadlock) sources=interview_questions.md:1099-1288, interviews.md:1099-1288 -->

> Source: `interview_questions.md:1099-1288`
> Classification reason: deadlock
> Duplicate source aliases: `interview_questions.md:1099-1288, interviews.md:1099-1288`

##### 교착 상태 (Deadlock)

발생 필수 조건 (Coffman 조건):

1. 상호 배제 (Mutual Exclusion)
    - 자원은 한 번에 하나의 프로세스만 사용 가능
    - 다른 프로세스의 자원 사용 완료를 기다려야 함

2. 점유와 대기 (Hold and Wait)
    - 프로세스가 이미 자원을 보유한 상태에서
    - 다른 자원을 추가로 요청하고 대기

3. 비선점 (No Preemption)
    - 다른 프로세스가 사용 중인 자원을 강제로 빼앗을 수 없음
    - 자원을 보유한 프로세스가 자발적으로 반환해야 함

4. 순환 대기 (Circular Wait)
    - 프로세스들이 순환적으로 서로의 자원을 기다림
    - P1→P2→P3→P1 형태의 대기 사이클 형성

심층 예시와 분석:

1. 데이터베이스 트랜잭션 데드락

    ```sql
    -- Transaction 1
    BEGIN;
    UPDATE accounts SET balance = balance - 100 WHERE id = 1;
    -- Lock acquired on account 1
    UPDATE accounts SET balance = balance + 100 WHERE id = 2;
    COMMIT;

    -- Transaction 2 (동시 실행)
    BEGIN;
    UPDATE accounts SET balance = balance - 100 WHERE id = 2;
    -- Lock acquired on account 2
    UPDATE accounts SET balance = balance + 100 WHERE id = 1;
    COMMIT;
    ```

2. 자바 스레드 데드락

    ```java
    public class ResourceManager {
        private final Object resource1 = new Object();
        private final Object resource2 = new Object();

        public void method1() {
            synchronized(resource1) {
                processResource1();
                synchronized(resource2) {
                    processResource2();
                }
            }
        }

        public void method2() {
            synchronized(resource2) {
                processResource2();
                synchronized(resource1) {
                    processResource1();
                }
            }
        }
    }

    // 데드락 감지를 위한 스레드 덤프 분석
    /*
    "Thread-1" Id=12 BLOCKED
        at ResourceManager.method1()
        - waiting to lock resource2
        - locked resource1

    "Thread-2" Id=13 BLOCKED
        at ResourceManager.method2()
        - waiting to lock resource1
        - locked resource2
    */
    ```

해결 및 예방 전략:

1. 예방 (Prevention)

    ```java
    public class DeadlockFreeResourceManager {
        private final Object resource1 = new Object();
        private final Object resource2 = new Object();

        // 자원에 전역 순서 부여
        public void method1() {
            synchronized(resource1) {  // 항상 먼저 획득
                synchronized(resource2) {
                    process();
                }
            }
        }

        public void method2() {
            synchronized(resource1) {  // 동일한 순서 유지
                synchronized(resource2) {
                    process();
                }
            }
        }
    }
    ```

2. 회피 (Avoidance)

    ```java
    public class ResourceAllocator {
        private final Set<Resource> available = new HashSet<>();
        private final Map<Process, Set<Resource>> allocated = new HashMap<>();

        public boolean allocate(Process p, Resource r) {
            // 은행가 알고리즘 구현
            if (wouldBeDeadlocked(p, r)) {
                return false;
            }
            allocateResource(p, r);
            return true;
        }

        private boolean wouldBeDeadlocked(Process p, Resource r) {
            // 시스템 상태 검사
            // 안전 상태 유지 가능성 확인
            return false;
        }
    }
    ```

3. 탐지 및 복구 (Detection & Recovery)

    ```java
    public class DeadlockDetector {
        private final DirectedGraph resourceGraph = new DirectedGraph();

        public void detect() {
            // 자원 할당 그래프 분석
            if (hasCycle()) {
                recoverFromDeadlock();
            }
        }

        private void recoverFromDeadlock() {
            // 1. 프로세스 종료
            // 2. 자원 선점
            // 3. 프로세스 롤백
        }
    }
    ```

고급 해결 기법:

1. 타임아웃 기반

    ```java
    public class TimeoutBasedLocking {
        private final Lock lock = new ReentrantLock();

        public void process() throws TimeoutException {
            if (!lock.tryLock(1, TimeUnit.SECONDS)) {
                throw new TimeoutException("Lock acquisition timed out");
            }
            try {
                // 임계 영역
            } finally {
                lock.unlock();
            }
        }
    }
    ```

2. 트랜잭션 메모리

    ```java
    @Transactional
    public class OptimisticResourceManager {
        private Map<String, Integer> resources;

        public void updateResources(String r1, String r2) {
            // 낙관적 동시성 제어
            // 충돌 시 자동 재시도
            resources.compute(r1, (k, v) -> v + 1);
            resources.compute(r2, (k, v) -> v - 1);
        }
    }
    ```

<!-- /curriculum-chunk -->

## 스레드 상태와 스케줄링

### 2. 쓰레드 상태(States) 개념

#### 원문: 2. 쓰레드 상태(States) 개념

<!-- curriculum-chunk: sha256=6862e633a746c7b7eb9332cad0cc0d7d2f4c50bf50108719ec80a31cd85d72ec major=concurrency-async-io mid=스레드 상태와 스케줄링 sub=2. 쓰레드 상태(States) 개념 sources=interview_questions3.md:406-420 -->

> Source: `interview_questions3.md:406-420`
> Classification reason: thread state

##### 2. 쓰레드 상태(States) 개념

1. 자바 쓰레드는 **Thread.State** 열거형에 의해 여러 상태로 분류된다(일반 원칙).
2. 그 상태는 크게 **RUNNABLE, BLOCKED, WAITING, TIMED_WAITING, TERMINATED** 등이 있다(특정 원칙).
3. 따라서 어떤 특정 상황에서 쓰레드가 동작 중이거나(waiting, blocked), 종료되었는지 등을 구분할 수 있다(결론).

구체적으로:
- **RUNNABLE**: JVM 내부적으로 “실행 가능 상태” 또는 OS 스케줄러가 CPU에 올릴 수 있는 상태.
- **BLOCKED**: `synchronized`를 위해 모니터 잠금을 **획득하려고 시도**하지만, 잠금이 이미 다른 쓰레드에 의해 점유되어 있어서 **대기 중**인 상태.
- **WAITING**: 특정 조건(예: `wait()`, `join()`, `LockSupport.park()`)을 만족할 때까지 **“(notify 같은) 신호”**를 기다리는 상태.
- **TIMED_WAITING**: `wait(timeout)`, `sleep(timeout)` 처럼 유한 시간 동안만 대기하는 상태.
- **TERMINATED**: 쓰레드 실행이 끝난 상태.

---

<!-- /curriculum-chunk -->

## 스레드 생성과 스케줄링

### `pthread_create` 이후 커널에서 발생하는 작업들

#### 원문: `pthread_create` 이후 커널에서 발생하는 작업들

<!-- curriculum-chunk: sha256=1f6c20060905e8f0f49500dac33bc6e7f21447d534cb445e562fa75bffb38fc3 major=concurrency-async-io mid=스레드 생성과 스케줄링 sub=`pthread_create` 이후 커널에서 발생하는 작업들 sources=interview_questions.md:3771-3811, interviews.md:3771-3811 -->

> Source: `interview_questions.md:3771-3811`
> Classification reason: thread creation
> Duplicate source aliases: `interview_questions.md:3771-3811, interviews.md:3771-3811`

##### `pthread_create` 이후 커널에서 발생하는 작업들

Linux에서 스레드는 실제로 프로세스의 경량 버전입니다.
스레드는 프로세스 내에서 실행되는 가벼운 실행 단위입니다.
여러 스레드는 같은 프로세스 내에서 코드, 데이터, 힙을 공유하면서 독립적인 실행 흐름(스택, 레지스터)을 유지합니다.
스레드는 병렬 처리를 위해 사용되며, 각 스레드는 프로세스의 자원을 공유하지만, 독립적인 실행 컨텍스트(레지스터 상태 및 스택)를 가집니다.

쓰레드 생성은 커널에 직접적인 개입을 필요로 하며, 이를 위해 `pthread_create` 함수가 사용됩니다.
`pthread_create`는 결국 시스템 호출인 `clone()` 함수를 사용하여 새로운 쓰레드를 생성합니다.

clone()은 스레드 생성뿐만 아니라 프로세스 생성도 담당할 수 있는 다목적 시스템 호출입니다.
clone()은 인자로 전달된 플래그에 따라 프로세스나 스레드가 공유할 자원을 결정합니다.

스레드 생성 시, 부모 프로세스의 PCB가 복사되어 스레드 제어 블록(TCB, Thread Control Block)이 생성됩니다.
PCB와 TCB는 구조적으로 유사하지만, 스레드는 프로세스의 메모리 자원을 공유합니다.
커널은 clone() 플래그를 참조하여 스레드가 공유할 자원을 결정합니다.
일반적으로 스레드는 부모와 가상 메모리 공간, 파일 디스크립터 테이블, 신호 처리 정보를 공유합니다.

스레드는 독립적인 스택을 가집니다.
clone() 호출 시, 새로운 스레드를 위한 스택 공간이 설정되고, 새 스레드는 독립적인 레지스터 상태 및 프로그램 카운터를 가지고 실행을 시작합니다.

새로 생성된 스레드는 스케줄링 큐에 추가되어 독립적으로 실행됩니다.
스레드는 같은 프로세스 내의 다른 스레드들과 병렬로 실행될 수 있습니다.

1. 커널은 새로운 쓰레드가 기존 프로세스와 공유할 리소스(메모리, 파일 디스크립터, 시그널 핸들러 등)를 설정합니다.
2. 쓰레드마다 별도의 스택이 필요하므로, 쓰레드용 스택이 할당됩니다.
3. 부모 프로세스의 레지스터와 문맥 정보를 복사하여 새 쓰레드가 동일한 환경에서 실행을 시작할 수 있도록 합니다.
4. 커널은 `clone()` 시스템 호출을 통해 이 작업을 처리하며, 이를 통해 쓰레드는 같은 주소 공간을 공유하게 됩니다
5. 새로 생성된 스레드는 커널의 스케줄러에 의해 스케줄링 큐에 등록되어, 시스템이 언제 해당 스레드를 실행할지 결정합니다​​.

프로세스는 독립적인 가상 메모리 공간을 가집니다(복사 혹은 COW 적용).
스레드는 가상 메모리 공간을 공유하며, 같은 힙과 전역 데이터를 사용할 수 있습니다.

프로세스 간의 컨텍스트 스위칭은 더 많은 작업을 요구합니다.
프로세스는 독립적인 메모리 공간을 사용하기 때문에 페이지 테이블, PCB(프로세스 제어 블록), 캐시 등을 모두 교체해야 합니다.
이로 인해 프로세스 간의 스위칭은 더 많은 비용을 소모합니다.

스레드 간의 컨텍스트 스위칭은 비교적 비용이 적게 듭니다.
왜냐하면 스레드는 동일한 프로세스 내에서 실행되므로 가상 메모리, 파일 디스크립터 테이블 등 대부분의 자원을 공유하기 때문입니다.
스택, 레지스터, 프로그램 카운터만을 변경하면 되기 때문에 오버헤드가 상대적으로 적습니다.

<!-- /curriculum-chunk -->
