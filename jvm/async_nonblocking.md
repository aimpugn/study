# Async nonblocking

## 스케쥴러, 런큐, 그리고 대기큐

런큐(run queue)는 운영체제가 "지금 당장 CPU에 태울 수 있는 `task_struct`들의 대기실"을 의미합니다.
이 태스크는 "메모리 공간과 PID를 혼자 차지한다"는 점에서 프로세스일 수도 있고,
"같은 메모리 공간을 공유하면서 PID만 다른 스레드 그룹의 일원"이라는 점에서 스레드일 수도 있습니다.
리눅스는 `clone()` 시스템콜의 플래그 조합으로 "무엇을 공유할지"만 결정할 뿐, 결국 스케줄링 관점에서는 전부 동일한 `task_struct`입니다.

런큐는 한 개의 CPU 코어마다 하나씩 존재하므로, 네 개 코어 시스템이라면 곧 네 개의 런큐가 따로 돌아갑니다.
Completely Fair Scheduler(CFS)는 주기적으로 각 런큐 길이를 비교해 코어 간 태스크를 이동시키는 load balancing을 수행하지만, 기본 스케줄링 결정을 내리는 단위는 여전히 CPU별 런큐입니다.

리눅스의 기본 스케줄러인 Completely Fair Scheduler(CFS)는 이 런큐를 가상 실행시간(`vruntime`) 순으로 정렬한 레드블랙 트리로 구현합니다.
트리의 정렬 키는 '매 스케줄링 틱마다 태스크가 실제로 사용한 나노초 단위를 계속 합산해 두는 값을 우선순위 가중치로 나눈 가상 실행 시간'인 `se.vruntime`입니다.
상위 `nice` 값을 가진 태스크는 weight가 작아서 실제로 같은 시간을 달려도 `vruntime`이 빨리 늘기 때문에 "앞으로 CPU를 덜 받아도 된다"는 신호가 됩니다.

각각의 런큐는 지금 이 코어에서 언제든 실행될 수 있는 `task_struct`들을 한데 모아 둡니다.
런큐에 엔트리가 존재한다는 사실이 곧 `TASK_RUNNING`(state `R`)를 의미합니다.
`TASK_RUNNING`은 CPU 위에서 실행 중이거나 런큐에 올라가 실행 차례를 기다리는(runnable) 태스크 모두를 포괄합니다.

```c
// 리눅스 커널의 실제 태스크 상태들
#define TASK_RUNNING         0x00000000
#define TASK_INTERRUPTIBLE   0x00000001
#define TASK_UNINTERRUPTIBLE 0x00000002
#define TASK_STOPPED         0x00000004
#define TASK_TRACED          0x00000008
// ... 기타 상태들
```

스케줄러가 타임슬라이스가 끝날 때마다 트리의 맨 왼쪽 노드, 즉 `vruntime`이 가장 작은 태스크(즉 "가장 덜 달린" 태스크)를 꺼내 CPU 레지스터에 적재합니다.
커널은 컨텍스트 스위치를 수행해 방금 고른 태스크를 실제 CPU 코어 위에 올립니다.
컨텍스트 스위치가 일어날 때 커널은 먼저 지금 돌고 있던 태스크의 레지스터 값을 PCB(Process Control Block) 영역에 저장한 뒤, 선택된 태스크의 PCB에서 레지스터 세트를 꺼내 *CPU 하드웨어 레지스터에 적재*합니다.
이어서 프로그램 카운터가 새 태스크의 명령주소를 가리키도록 갱신되면 곧바로 그 코드가 실행을 재개합니다.
그래서 *"레지스터 집합을 적재한다"*는 표현은 "선택된 태스크 안에 보관돼 있던 이전 실행 맥락을 CPU 레지스터에 다시 불러 넣는다"는 것을 뜻합니다.

즉, 런큐는 '다음에 CPU를 쓸 차례를 기다리는 태스크들이 CPU별로 줄서는 목록'이라고 볼 수 있습니다.
런큐 관점에서 본다면, 블로킹은 태스크가 그 줄에서 계속 버티는 것이고, 논블로킹은 잠시 빠졌다가 이벤트 발생 시점에 다시 끼어드는 것입니다.
이러한 관점에서 `run queue length` 같은 지표가 서비스 지연과 연결되고, 논블로킹 I/O가 런큐를 꾸준히 채우면 CPU 활용률과 처리량을 동시에 끌어올릴 수 있습니다.

만약 태스크가 `read()` 같은 블로킹 시스템 콜로 들어가 데이터를 기다리다가 I/O 자원이 아직 준비되지 않았다는 사실을 확인하면, 태스크는 자신의 상태를 `TASK_UNINTERRUPTIBLE` 또는 `TASK_INTERRUPTIBLE`으로 변경합니다.
- `TASK_INTERRUPTIBLE`이면 시그널을 받을 때도 깨어납니다.
- `TASK_UNINTERRUPTIBLE`이면 시그널을 무시한 채 오직 디바이스 완료 이벤트에 의해서만 깨어납니다.

그리고 커널 함수 `schedule()`를 호출하여 제어권을 넘깁니다(자발적 CPU 양보, voluntary context switch).
`schedule()`은 현재 태스크를 런큐에서 삭제하고 'wait queue'(슬립 큐)에 추가합니다.
이때부터는 CPU가 아무리 많아도 그 태스크는 런큐에 없으므로 선택될 일이 없습니다.
가령 CPU가 두 개인데 태스크 다섯 개가 모두 잠들어 버리면 두 런큐가 비게 되고, 커널은 두 아이들 태스크를 태우며 "CPU가 한가하다"는 상태로 빠지게 됩니다.
런큐를 다시 채우기 전까지는 계산 작업도, 다른 요청 처리도 진행되지 못합니다.

```plaintext
# 시스템콜이 호출되면 사용자 모드에서 커널 모드로 전환되지만, 실행 주체는 여전히 동일한 태스크입니다.

User Space                    Kernel Space
----------                    ------------
[Java Thread]  --syscall-->  [Same Thread in Kernel Mode]
   |                                     |
   |                                     v
read() 호출                      [sys_read() 함수 실행]
   |                                     |
   |                                     v
   |                             [VFS 계층 함수들 실행]
   |                                     |
   |                                     v
   |                             [드라이버 함수들 실행]
   |                                     |
   |                                     v
   |                          [현재 태스크가 자신의 상태 변경]
   |                                     |
   |                                     v
   |                          [schedule() 호출로 CPU 양보]
```

Wait queue는 "특정 이벤트가 발생하기를 기다리는 태스크들의 대기열"입니다.
특정 조건이 만족될 때까지 잠들어야 하는 태스크들의 집합입니다.
이 wait queue는 커널 오브젝트(자원)에 속합니다. 즉, 파일, 소켓, blk I/O 요청, 파이프, 의사 잠금 등 거의 모든 커널 자원은 자신만의 wait queue를 가집니다.
이는 태스크가 기다리는 이벤트의 종류가 다르기 때문입니다.
'소켓에서 데이터를 기다리는 태스크'와 '파일 읽기를 기다리는 태스크'는 서로 다른 조건에 의해 깨어나야 하므로 별도의 대기열이 필요합니다.

런큐 길이가 텅 비면 커널은 아예 스페셜 아이들(idle) 태스크를 태워 "CPU가 놀고 있다"는 신호를 전력 관리 서브시스템에 보냅니다.
반대로 런큐 길이가 코어 수를 훨씬 넘어서면, 사용자 공간에서는 CPU 사용률이 높음에도 응답 지연이 길어지는 현상을 체감하게 됩니다.

`vmstat`, `sar -q`, `top` 의 load average 열을 통해 런큐 길이와 평균 대기 시간을 실시간으로 확인할 수 있습니다. 이 값은 결국 각 코어 런큐를 전부 합친 크기를 지수평균한 결과입니다.

### JDBC 호출 케이스

런큐 관점에서 JDBC 호출은 자바 애플리케이션에서 데이터베이스 호출 시 해당 스레드가 런큐에서 웨이트 큐로 이동하여 실질적인 동시 처리 능력이 제한되는 현상을 의미합니다.

가령 톰캣에서 200개 스레드를 설정했다고 해서 200개 요청을 동시에 처리할 수 있는 것이 아니라,
실제로는 런큐에 머물러 있는 스레드 수만큼만 동시 처리가 가능합니다.

```java
// 예시: 톰캣 스레드 풀 상황
maxThreads=200, 현재 활성 스레드=200
├── 런큐에 있는 스레드: 5개 (실제 CPU에서 실행 중이거나 대기 중)
├── DB I/O 대기 중: 150개 (각각 다른 DB 연결의 wait queue에서 대기)
├── 외부 API 대기 중: 40개 (HTTP 소켓의 wait queue에서 대기)
└── 기타 I/O 대기 중: 5개

// 실제 동시 처리량 = 5개 (런큐에 있는 스레드만 CPU를 받을 수 있음)
```

데이터베이스는 네트워크 너머에 있는 별도의 시스템입니다. JDBC 호출이 일어나면 다음과 같은 물리적 과정을 거쳐야 합니다.

1. SQL 쿼리를 TCP 패킷으로 변환하여 네트워크 전송
2. 데이터베이스 서버에서 쿼리 파싱, 실행, 결과 생성
3. 결과를 TCP 패킷으로 변환하여 네트워크 응답
4. 애플리케이션 서버에서 패킷 수신 및 `ResultSet` 변환

이 모든 과정은 수 밀리초에서 수백 밀리초가 걸리며, CPU 관점에서는 "엄청나게 긴 시간"입니다.
만약 스레드가 이 시간 동안 런큐에 계속 머물러 있다면, CPU 사이클을 낭비하면서 아무 일도 하지 않고 기다리는 상황이 됩니다.

이러한 낭비를 피하기 위해 태스크가 네트워크 I/O를 기다리는 동안, 해당 태스크를 런큐에서 제거하여 다른 태스크가 CPU를 사용할 수 있도록 하여 전체 시스템의 처리량을 극대화합니다.

```java
// Java 애플리케이션
ResultSet rs = statement.executeQuery("SELECT * FROM users");
```

이 코드는 JDBC 드라이버가 SQL을 TCP 소켓을 통해 전송하려 하지만, 네트워크 전송은 즉시 완료되지 않습니다.
특히 소켓 송신 버퍼가 가득 차 있거나, 응답 데이터가 아직 도착하지 않은 상황에서는 대기가 필요합니다.

현재 스레드가 소켓에서 데이터를 기다려야 한다고 판단되면, 스레드 자신이 자발적으로 다음 과정을 수행합니다:
1. 자신의 상태를 `TASK_INTERRUPTIBLE`(시그널로 깨어날 수 있는 대기 상태)로 변경
2. 해당 소켓의 웨이트 큐에 자신을 등록
3. `schedule()` 함수를 호출하여 CPU 제어권을 다른 태스크에게 양보

이때 런큐에서는 해당 스레드 엔트리가 완전히 사라집니다.
CFS 스케줄러의 레드블랙 트리에서 제거되어, 다음 순위의 태스크가 자동으로 트리의 가장 좌측(다음 실행 대상)으로 이동합니다.

```plaintext
[Java Thread] → [JDBC Driver] → [TCP Socket] → [Network Driver] → [Database Server]
      ↓              ↓              ↓                ↓
[Runtime Stack] [Native Call] [Socket Wait Queue] [NIC Queue] → [DB Processing]
                                    ↑
                          여기서 태스크가 잠들고
                          네트워크 응답까지 대기합니다.
```

런큐에 남아있는 `TASK_RUNNING` 상태의 태스크 수가 곧 시스템의 실질적 동시 처리 능력이 됩니다. CPU가 아무리 빠르고 많아도, 런큐가 비어있으면 처리할 작업이 없는 상태입니다.

네트워크 인터럽트가 발생하여 DB 응답 패킷이 도착하면:
1. 인터럽트 핸들러가 해당 소켓의 웨이트 큐에서 대기 중인 태스크들을 깨웁니다.
2. 깨어난 태스크가 `TASK_RUNNING` 상태로 변경되어 런큐에 다시 추가됩니다.
3. CFS 스케줄러가 `vruntime`에 따라 해당 태스크를 적절한 위치에 배치합니다.

많은 경우 "동시 접속자가 많으니 스레드 풀을 늘려야겠다"* 생각하지만, 이는 근본적 해결책이 아닙니다.
스레드를 아무리 늘려도 대부분이 웨이트 큐에서 잠들어 있다면 실제 처리량은 향상되지 않습니다.

오히려 스레드가 너무 많으면 오히려 다음과 같은 문제들이 발생합니다:
- 메모리 사용량 증가 (스레드당 약 1MB의 스택 공간)
- 컨텍스트 스위치 오버헤드 증가
- 더 많은 스레드가 동시에 웨이트 큐에서 대기하는 상황

진짜 최적화는 블로킹 I/O 자체를 줄이는 것이 핵심입니다:
- 연결 풀링: DB 연결 생성/해제 시간 단축
- 쿼리 최적화: DB 응답 시간 단축
- 캐싱: I/O 호출 빈도 감소
- 배치 처리: 여러 쿼리를 한 번에 처리
- 논블로킹 아키텍처: WebFlux + R2DBC로 아예 블로킹 제거

### HTTP 클라이언트 호출 케이스

```java
// RestTemplate을 사용한 외부 API 호출
RestTemplate restTemplate = new RestTemplate();
String result = restTemplate.getForObject("https://api.external.com/data", String.class);
```

```c
// DNS 해석용 UDP 소켓
struct socket dns_sock = {
    .wait = dns_wait_queue,     // DNS 응답 전용 wait queue
    .type = SOCK_DGRAM,
    .dest_addr = dns_server_ip
};

// HTTP 통신용 TCP 소켓
struct socket http_sock = {
    .wait = http_wait_queue,    // HTTP 응답 전용 wait queue
    .type = SOCK_STREAM,
    .dest_addr = api_server_ip
};

// 1. DNS 해석 단계
struct socket *dns_socket = create_udp_socket();
// DNS 서버 응답을 위한 wait queue

// 2. TCP 연결 설정 단계
struct socket *http_socket = create_tcp_socket();
connect(http_socket, server_addr);
// TCP 3-way handshake 완료를 위한 wait queue

// 3. HTTP 요청 전송 단계
write(http_socket, http_request, request_size);
// 송신 버퍼 여유 공간을 위한 wait queue

// 4. HTTP 응답 수신 단계
read(http_socket, response_buffer, buffer_size);
// 응답 데이터 도착을 위한 wait queue
```

DNS 응답과 HTTP 응답은 *서로 다른 소켓*, *다른 서버*, *다른 타이밍*에 도착하므로 독립적인 이벤트 관리가 필요하므로 각각 별도의 wait queue가 필요합니다.

### 순차적 의존성이 있는 블로킹 I/O 체인에서 WebFlux 사용의 이점

WebFlux를 사용하는 이점은 존재하지만 제한적입니다.
하지만 "어떻게 사용하느냐"에 따라 그 이점의 크기가 달라집니다.

```java
DB조회결과=DB_조회();
외부_호출_파라미터=DB조회결과.toParameters();
결과 = 외부_서비스_호출();
```

이 코드는 논리적으로는 순서가 중요합니다.
DB 조회 결과가 있어야 외부 서비스 호출이 가능하므로 비즈니스 로직상 대기가 필수입니다.
하지만 **스레드가 런큐에서 웨이트 큐로 이동하는 물리적 블로킹은 별개의 문제**입니다.
WebFlux의 진짜 이점은 **논리적 순서는 유지하면서 물리적 블로킹을 제거**하는 데 있습니다.

전통적인 Spring MVC 방식의 코드는 다음과 같습니다.

```java
@RestController
public class MvcController {

    @GetMapping("/transfer")
    public TransferResponse transfer(@RequestBody TransferRequest request) {
        // 1. 스레드가 DB 조회를 위해 DB 웨이트 큐로 이동
        Account account = accountRepository.findById(request.getAccountId());
        // DB 응답 후 런큐로 복귀

        // 2. 스레드가 외부 API 호출을 위해 HTTP 웨이트 큐로 이동
        ExternalResult result = externalApiClient.call(account.toParameters());
        // 외부 서비스 응답 후 런큐로 복귀

        // 3. 스레드가 DB 업데이트를 위해 DB 웨이트 큐로 이동
        transactionRepository.save(result.toDBData());
        // DB 완료 후 런큐로 복귀

        return new TransferResponse(result);
    }
}
```

메모리 및 런큐 관점에서 다음과 같은 문제점이 있습니다:
- 각 I/O 단계마다 스레드가 웨이트 큐로 이동
- 전체 처리 시간 = 각 I/O 대기 시간의 합
- 스레드가 대부분의 시간을 웨이트 큐에서 소비
- 각 스레드는 약 1MB의 스택 공간을 사용하므로 1000개 동시 요청 = 1000개 스레드 = 1GB 메모리
- MVC에서는 각 I/O마다 상당한 CPU 오버헤드를 발생시키는 컨텍스트 스위칭 반복
    1. 스레드 상태를 `TASK_INTERRUPTIBLE`로 변경
    2. 레지스터 값을 PCB에 저장
    3. 런큐에서 제거하고 웨이트 큐에 추가
    4. 다른 스레드로 컨텍스트 스위치
    5. I/O 완료 시 웨이트 큐에서 런큐로 복귀
    6. 레지스터 값을 PCB에서 복원

이때 하나의 요청이 스레드를 독점하면서 대부분의 시간을 wait queue에서 보내게 됩니다.

```plaintext
Timeline: |---DB 대기----|--외부API 대기---|--DB 업데이트 대기--|
RunQueue: [    빈상태    ][    빈상태     ][      빈상태      ]
Thread:   [Wait Queue A][ Wait Queue B ][  Wait Queue A  ]
```

반면 WebFlux를 사용하면 다음과 같습니다.

```java
@RestController
public class WebFluxController {
    @GetMapping("/transfer")
    public Mono transfer(@RequestBody TransferRequest request) {
        return accountRepository.findById(request.getAccountId())
            .flatMap(account ->
                externalApiClient.call(account.toParameters())
                    .flatMap(result ->
                        transactionRepository.save(result.toDBData())
                            .map(saved -> new TransferResponse(result))
                    )
            );
    }
}
```

메모리 및 런큐 관점에서 다음과 같은 이점이 있습니다.
- 이벤트 루프 스레드가 런큐에 상주하면서 콜백 체인을 처리
- 각 I/O는 콜백 체인으로 연결되어 비동기 처리
- 스레드는 웨이트 큐가 아닌 런큐에서 다른 요청 처리 가능
- 1000개 동시 요청 = 소수의 이벤트 루프 스레드 (보통 CPU 코어 수만큼)
- 이벤트 루프 스레드는 런큐에서 계속 실행되면서 콜백 체인만 관리하므로 컨텍스트 스위칭 대폭 감소

즉, WebFlux의 경우 소수의 이벤트 루프 스레드가 런큐에서 지속 실행되면서 수많은 요청을 동시 처리합니다.

```plaintext
Event Loop: [런큐에서 지속 실행]
├── Request 1의 DB 호출 콜백 등록 → 즉시 다음 작업
├── Request 2 처리 시작 → API 호출 콜백 등록 → 즉시 다음 작업
├── Request 3 처리 시작 → ...
├── Request 1의 DB 응답 도착 → 콜백 실행 → 외부 API 호출 등록
└── Request 2의 API 응답 도착 → 콜백 실행 → ...
```

WebFlux는 I/O 바운드 작업에 특히 효과적입니다.
I/O 바운드 작업*(WebFlux 유리):
- DB 조회, 외부 API 호출, 파일 읽기
- 대기 시간이 길고 CPU 사용률이 낮음
- WebFlux로 런큐 활용률 극대화 가능

반면 CPU 바운드 작업의 경우 차이가 적습니다.
- 복잡한 계산, 암호화, 이미지 처리
- CPU를 집약적으로 사용
- 어차피 CPU가 busy하므로 WebFlux 이점 제한적

WebFlux에서도 blocking operation이 필요한 경우 별도 스케줄러로 분리할 수 있습니다:

```java
@GetMapping("/process")
public Mono process() {
    return Mono.fromCallable(() -> {
        // 블로킹 작업들을 별도 스레드 풀에서 실행
        User user = jdbcUserRepository.findById(1L);
        ExternalResponse resp = blockingExternalService.call(user.toRequest());
        jdbcUserRepository.upsert(resp.toDbData());
        return resp.toResult();
    })
    .subscribeOn(Schedulers.boundedElastic())  // 블로킹 전용 스레드 풀
    .publishOn(Schedulers.parallel());         // 결과 처리는 메인 스레드 풀
}
```

이 코드의 경우 다음과 같은 이점이 있습니다.

```java
Main Event Loop Threads (보통 CPU 코어 수만큼):
├── 런큐에 상주하여 HTTP 요청 처리, 콜백 실행
├── 블로킹 I/O에 빠지지 않음
└── 높은 동시성 유지

Bounded Elastic Pool (별도 스레드 풀):
├── 블로킹 작업만 전담 처리
├── wait queue 이동은 이 풀에서만 발생
└── 메인 이벤트 루프와 격리
```

또한 메모리 효율성과 컨텍스트 스위치가 감소합니다.

검색 결과 에서 보여주듯이 WebFlux는 더 적은 메모리를 사용합니다:

Spring MVC:
- 요청당 1개 스레드 필요 (약 1MB 스택)
- 1000개 동시 요청 = 1000개 스레드 = 1GB 메모리

WebFlux:
- 소수의 이벤트 루프 스레드 (CPU 코어 수만큼)
- 1000개 동시 요청도 동일한 스레드 수로 처리
- 현저히 적은 메모리 사용량

컨텍스트 스위치 관점:

```c
// MVC: 매 요청마다 새로운 task_struct가 런큐-웨이트큐 이동
Task-1: RunQueue → DB WaitQueue → RunQueue → HTTP WaitQueue → ...
Task-2: RunQueue → DB WaitQueue → RunQueue → HTTP WaitQueue → ...
...
Task-1000: RunQueue → DB WaitQueue → RunQueue → HTTP WaitQueue → ...
```

```c
// WebFlux: 소수의 이벤트 루프가 런큐에 상주
EventLoop-1: 런큐 상주 (콜백 체인 처리)
EventLoop-2: 런큐 상주 (콜백 체인 처리)
...
EventLoop-N: 런큐 상주 (콜백 체인 처리)
```

또한 순차적 시나리오에서도 일부 최적화가 가능합니다:

```java
// 개선된 WebFlux 방식
@GetMapping("/process")
public Mono process() {
    return userRepository.findById(1L)
        .flatMap(user -> {
            ExternalRequest req = user.toExternalRequest();

            // 외부 서비스 호출과 동시에 다른 사전 작업 수행
            Mono externalCall = externalService.call(req);
            Mono parallelWork = someOtherService.prepareData();

            return Mono.zip(externalCall, parallelWork)
                .flatMap(tuple -> {
                    ExternalResponse response = tuple.getT1();
                    SomeData preparedData = tuple.getT2();

                    // DB 업데이트와 로그 기록을 병렬로
                    Mono dbUpdate = userRepository.upsert(response.toDbData());
                    Mono logging = logService.logTransaction(response, preparedData);

                    return Mono.zip(dbUpdate, logging)
                        .thenReturn(response.toResult());
                });
        });
}
```

MVC의 문제점은 다음과 같습니다.

```java
// 모든 톰캣 스레드가 블로킹 I/O에 빠진 상황
maxThreads=200
현재 상황:
├── DB 대기: 150개 스레드 (각각 서로 다른 DB wait queue)
├── 외부 API 대기: 45개 스레드 (HTTP wait queue)
└── 기타 대기: 5개 스레드

// 새로운 요청 도착 시 → 처리할 스레드 없음 → 큐에서 대기
```

WebFlux의 장점은 다음과 같습니다.

```java
// 이벤트 루프는 항상 런큐에 있어서 새 요청 수락 가능
EventLoop 스레드: 런큐 상주 (새 요청 즉시 수락)
BoundedElastic 풀: 블로킹 작업만 처리 (격리됨)

// 시스템이 완전히 멈추는 상황 방지
```

블로킹 호출이 많은 환경에서도 reactive approach의 이점이 있습니다:

1. 높은 동시성 요구사항 (1000+ 동시 요청)
2. I/O 대기 시간이 긴 경우 (외부 API 응답 지연 등)
3. 메모리 제약이 있는 환경
4. 부분적 병렬화 가능성이 있는 비즈니스 로직

단, 다음과 같은 경우에는 WebFlux를 피해야 합니다.

1. 팀의 reactive programming 경험 부족
2. 완전히 순차적이고 단순한 CRUD
3. 낮은 동시성 (100 미만 동시 요청)
4. 디버깅과 유지보수가 더 중요한 경우

순차적 블로킹 시나리오에서도 WebFlux 사용 이점이 있습니다:

1. 런큐 효율성: 이벤트 루프가 런큐에 상주하여 새 요청 처리
2. 메모리 효율성: 적은 스레드로 높은 동시성 처리
3. 시스템 안정성: 스레드 풀 고갈 방지
4. 부분적 병렬화: 일부 작업의 동시 실행 가능
5. 확장성: 향후 논블로킹 컴포넌트 도입 시 자연스러운 전환

## 동기/비동기, 블로킹/논블로킹 정리

**동기와 비동기는 "작업의 완료를 어떻게 처리하는가?"에 대한 개념입니다:**

- 동기(Synchronous):
    - 요청한 작업이 완료될 때까지 요청한 스레드가 작업을 직접 처리합니다.
    - 호출한 함수(caller)가 호출된 함수(callee)의 작업 완료 여부를 직접 확인하며, 작업이 완료될 때까지 기다립니다.
    - 즉, 작업의 완료 여부를 호출한 함수가 신경 쓰며, 순차적으로 작업을 처리합니다.

    ```java
    void syncMethod() {
        String result = fetchData();  // 데이터 요청 후 직접 기다림
        System.out.println("Result: " + result);
    }

    String fetchData() {
        return "Data";  // 데이터 반환 후 다음 코드 실행
    }
    ```

- 비동기(Asynchronous):
    - 요청한 작업을 별도의 스레드가 처리하고, 요청한 스레드는 즉시 반환됩니다.
    - 호출한 함수(caller)가 호출된 함수(callee)의 작업 완료 여부를 신경 쓰지 않고, 호출된 함수(callee)가 작업 완료 후 결과를 전달하는 방식입니다.
    - 일반적으로 콜백 함수나 이벤트 리스너를 통해 결과를 전달받으며, 호출한 함수(caller)는 다른 작업을 계속 수행할 수 있습니다.

    ```java
    void asyncMethod() {
        CompletableFuture.supplyAsync(this::fetchData)
            .thenAccept(result -> System.out.println("Result: " + result));
        System.out.println("Request sent, doing other work...");
    }

    ```java
    CompletableFuture.supplyAsync(() -> {
        return readDataFromSocket(); // 백그라운드 스레드에서 실행됩니다.
        // 요청한 스레드는 즉시 반환되며, 작업이 완료되면 `thenAccept()`에서 결과를 처리합니다.
    }).thenAccept(data -> {
        // 비동기는 요청한 스레드가 직접 데이터를 확인할 필요 없이,
        // 작업이 완료되면 자동으로 콜백을 실행함.
        System.out.println("데이터 수신 완료: " + data);
    });
    ```

즉, 동기는 "작업이 완료될 때까지 대기하는 방식"이고, 비동기는 "작업을 요청하고 결과가 오면 처리하는 방식"이다.

**블로킹(Blocking)과 논블로킹(Non-blocking)은 "작업을 요청한 스레드가 결과를 기다리는 방식"에 대한 개념입니다:**

- 블로킹(Blocking):
    - 블로킹 호출은 요청한 작업이 완료될 때까지 스레드가 CPU 타임슬라이스를 소진하거나 `park()` 되어 대기합니다.
    - 호출된 함수(callee)가 자신의 작업을 완료할 때까지 호출한 함수(caller)의 실행을 멈추게 합니다.
    - 즉, 제어권을 호출된 함수(callee)가 가지고 있으며, 작업이 끝날 때까지 호출한 함수(caller)는 대기 상태에 놓입니다.
    - 자바 쓰레드가 커널에 들어가 `read()` 같은 시스템 콜을 수행할 때, 커널이 "아직 읽을 데이터가 없다"고 판단하면 그 쓰레드는 런큐에서 빠져나가 잠들고, 깨어날 때까지 CPU를 전혀 얻지 못합니다. 이 상태가 바로 블로킹입니다.

    ```java
    // 블로킹 방식: read()가 완료될 때까지 현재 스레드는 멈춰 있음
    InputStream input = socket.getInputStream();
    int data = input.read();  // 데이터가 도착할 때까지 스레드가 대기
    ```

- 논블로킹(Non-Blocking):
    - 요청한 작업이 즉시 반환되며, 완료 여부와 관계없이 스레드는 계속 실행됩니다.
    - 논블로킹 호출은 커널/라이브러리에게 맡기고 즉시 리턴합니다.
    - 호출된 함수(callee)가 즉시 제어권을 호출한 함수(caller)에게 반환하여, 호출한 함수(caller)가 다른 작업을 계속 수행할 수 있도록 합니다.
    - 호출된 함수(callee)는 자신의 작업 완료 여부와 상관없이 즉시 반환합니다.

    ```java
    InputStream input = socket.getInputStream();
    socket.configureBlocking(false);  // 논블로킹 모드 설정
    int data = input.read();  // 데이터가 없으면 -1 반환, 스레드는 대기하지 않음
    ```

즉, 블로킹은 "작업이 끝날 때까지 기다리는 방식"이고, 논블로킹은 "기다리지 않고 즉시 반환하는 방식"입니다.

1. 블로킹 + 동기 (Blocking + Synchronous):  `Thread.sleep()`, `read()`, `Future.get()` 등
    - 요청한 스레드가 결과를 기다림(블로킹) + 요청한 스레드가 작업을 직접 처리(동기)
    - 호출한 함수가 호출된 함수의 작업이 완료될 때까지 기다립니다.
    - 그동안 다른 작업을 수행하지 못합니다.

    ```java
    ResultSet rs = statement.executeQuery("SELECT * FROM users");
    ```

2. ~~블로킹 + 비동기 (Blocking + Asynchronous)~~: 불가
    - 블로킹의 경우 현재 스레드가 결과를 기다리며 멈추고, 비동기는 다른 실행 주체가 작업을 처리하고 완료 시 알리는데, 이 둘은 모순적 개념입니다.
    - 블로킹이면 현재 스레드가 계속 기다리고 있으므로 "비동기적으로 알림받는다"는 것이 의미가 없습니다.

3. 논블로킹 + 동기 (Non-Blocking + Synchronous): Java NIO, `poll()`, `tryLock()`, 즉시 반환하는 함수들
    - 요청한 스레드가 결과를 기다리지 않음(논블로킹) + 요청한 스레드가 작업을 직접 처리(동기)
    - 호출된 함수가 즉시 제어권을 반환하여 호출한 함수가 다른 작업을 수행할 수 있습니다.
    - 하지만 호출한 함수는 주기적으로 호출된 함수의 작업 완료 여부를 확인해야 합니다.

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

    `read()`가 데이터를 읽을 준비가 되어 있지 않으면 즉시 -1을 반환합니다.
    이 경우, 스레드는 직접 반복해서 확인해야 합니다.(Polling).
    논블로킹은 작업을 직접 확인해야 하므로, 효율적이지 않을 수 있습니다.

4. 논블로킹 + 비동기 (Non-Blocking + Asynchronous): `CompletableFuture.supplyAsync()`, Node.js 콜백
    - 요청한 스레드가 결과를 기다리지 않음(논블로킹) + 요청한 작업을 별도의 스레드가 처리(비동기)
    - 호출된 함수가 즉시 제어권을 반환합니다.
    - 호출한 함수는 호출된 함수의 작업 완료 여부를 신경 쓰지 않습니다.
    - 호출된 함수는 작업 완료 후 콜백 등을 통해 결과를 전달합니다.

    ```js
    console.log('Start');
    // 1. fs.readFile() 호출 즉시 반환 (논블로킹)
    fs.readFile(
        'data.txt',
        // 3. I/O 작업은 libuv가 별도로 처리 (비동기)
        (err, data) => {  // 4. I/O 작업 완료 시 콜백이 이벤트 큐에 추가되어 이벤트 루프가 처리
            console.log('File read complete');
        }
    );
    // 2. 이벤트 루프 스레드는 런큐에서 계속 실행되어 다른 작업 처리
    console.log('End');  // 파일 읽기를 기다리지 않고 즉시 실행
    ```

    ```java
    // 또는 WebFlux
    return webClient.get()
        .uri("/api/data")
        .retrieve()
        .bodyToMono(String.class)
        .flatMap(result -> processResult(result));
    ```

    메인 스레드가 런큐에서 지속 실행합니다.
    I/O 작업은 별도 스레드/프로세스가 처리하고 완료 시 콜백/이벤트로 결과 전달합니다.

## 싱글 코어 & 싱글 스레드 환경에서도 비동기 I/O

싱글 코어 & 싱글 스레드 환경에서도 비동기 I/O가 가능한 이유는 CPU가 비동기적으로 I/O 작업을 요청한 후, *커널이 하드웨어 인터럽트 및 비동기 대기 메커니즘을 사용해 I/O 작업을 처리*하기 때문입니다.
애플리케이션은 I/O 작업을 요청한 후 다른 작업을 처리할 수 있으며, I/O 작업 완료 여부는 커널이 적절히 관리하고, 애플리케이션은 이를 비동기적으로 처리합니다.

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

비동기 I/O의 처리 과정은 다음과 같습니다:
1. (애플리케이션 -> 커널) 비동기 I/O 요청 전달:
    - 애플리케이션이 `aio_*` 함수를 호출하면, 해당 I/O 요청은 커널로 전달됩니다.
    - 커널은 해당 요청을 작업 큐(Work Queue)에 추가하고, 바로 애플리케이션에 제어를 반환합니다.
    - 이때 I/O 요청은 비동기로 처리되므로, 애플리케이션은 블로킹되지 않고 다른 작업을 수행할 수 있습니다.

2. (커널) 커널의 I/O 처리:
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

## Tomcat 8/9에서 NIO 커넥터

Tomcat NIO 커넥터는 기존 BIO(Blocking I/O) 커넥터의 "연결당 스레드" 비효율성을 부분적으로 해결한 아키텍처입니다.
Tomcat에서 NIO 커넥터는 세 종류의 내부 스레드로 이루어집니다.

- Acceptor - `ServerSocketChannel`을 *accept* 해 새 소켓을 등록합니다.
- Poller - 자바 NIO `Selector`를 "폴링(poll)"하며, 읽기·쓰기 등 이벤트가 준비된 소켓을 찾아 Processor(Worker) 풀에 넘깁니다.
- Worker/Processor - HTTP 파서를 돌리고, 서블릿 -> 필터 -> 디스패처 -> 컨트롤러 -> 서비스까지의 사용자 코드를 실행합니다.

하지만 NIO 가 *비동기*를 보장해 주지는 않습니다.

Poller가 non-blocking I/O로 이벤트를 모으더라도, 한 요청을 끝까지 처리하는 Worker는 여전히 "스레드 당 요청 1개" 모델입니다.
Worker가 JDBC, Redis, REST 호출처럼 블로킹 I/O 를 만나면 해당 스레드는 `park()`/`epoll_wait()` 상태로 잠이 듭니다.

CPU 코어가 하나라면 커널 스케줄러가 즉시 컨텍스트 스위치를 일으켜 다른 runnable 스레드를 실행하겠지만,

- Worker 풀이 가득 차 있다면 새 요청은 큐에 쌓이고,
- runnable 스레드도 없다면 CPU는 *idle* 로 들어갑니다.

따라서 Tomcat이 NIO를 쓰니 CPU가 쉬지 않는 것은 아닙니다.
NIO는 이벤트 처리 단계인 `accept`/`poll` 단계에서만 논블로킹일 뿐, *사용자-레벨 로직이 블로킹*이면 결국 동시성은 스레드 수 만큼으로 제한됩니다

### Acceptor - 연결 수락의 논블로킹화

```java
// Acceptor 스레드의 핵심 역할
while (running) {
    SocketChannel socket = serverSocket.accept();  // 새로운 연결 수락
    if (socket != null) {
        setSocketOptions(socket);  // 소켓 옵션 설정
        getPoller().register(socket);  // Poller에 등록
    }
}
```

기존 BIO 경우 각 연결마다 새 스레드 생성합니다. 이 때문에 수천 개 연결 시 수천 개 스레드가 런큐/웨이트 큐를 차지합니다.
반면 NIO Acceptor 경우 단일 스레드가 런큐에서 지속 실행되며 모든 연결 수락 처리하고 Poller에게 연결을 위임합니다.

### Poller - 이벤트 모니터링의 효율화

```java
// Poller 스레드의 핵심 동작
Selector selector = Selector.open();
while (running) {
    int readyChannels = selector.select(timeout);  // 논블로킹 이벤트 대기

    if (readyChannels > 0) {
        Set selectedKeys = selector.selectedKeys();
        for (SelectionKey key : selectedKeys) {
            if (key.isReadable()) {
                // 읽기 가능한 소켓을 Worker Pool에 할당
                assignToWorker(key.channel());
            }
        }
    }
}
```

기존 BIO 경우 각 연결마다 스레드가 `read()` 블로킹 호출에서 웨이트 큐 대기했습니다.
하지만 NIO Poller 경우 소수의 Poller 스레드가 런큐에서 지속 실행되며 수천 개 연결을 동시 모니터링합니다.
Poller라는 전용 스레드가 Selector를 이용해 수천 개의 커넥션을 감시하여 런큐 효율성을 극대화합니다.

### Worker - 여전한 블로킹의 한계

```java
// Worker 스레드에서의 요청 처리
public void processRequest(HttpServletRequest request, HttpServletResponse response) {
    // 1. HTTP 파싱은 논블로킹으로 완료됨

    // 2. 하지만 사용자 코드는 여전히 블로킹
    @GetMapping("/users/{id}")
    public User getUser(@PathVariable Long id) {
        User user = userRepository.findById(id);  // JDBC 블로킹 호출
        // 이 순간 Worker 스레드가 웨이트 큐로 이동

        Profile profile = externalApi.getProfile(user.getExtId());  // HTTP 블로킹 호출
        // 다시 웨이트 큐로 이동

        return new User(user, profile);
    }
}
```

"NIO connector is built with the assumption that your app will block somewhere"입니다.
NIO는 연결 관리는 효율화했지만 비즈니스 로직 처리는 여전히 블로킹입니다.

Worker 스레드 풀의 제약:

```xml
<Connector
    maxThreads="200"          // Worker 스레드 최대 개수
    maxConnections="10000"    // 최대 동시 연결 수
    acceptCount="100"         // 대기 큐 크기
    protocol="org.apache.coyote.http11.Http11NioProtocol" />
```

런큐 관점의 분석:
- maxConnections=10000: Poller가 런큐에서 동시 관리 가능한 연결 수
- maxThreads=200: 실제 비즈니스 로직을 처리하는 Worker 스레드 수
- 실질적 동시 처리량 = 200개 (런큐에 있는 Worker 스레드 수에 의해 제한)

따라서 NIO의 Poller는 런큐에서 효율적으로 실행되지만, Worker 스레드들은 여전히 블로킹 I/O로 인해 대부분의 시간을 웨이트 큐에서 소비합니다.

```java
// NIO 환경에서도 발생하는 블로킹 시나리오
@RestController
public class UserController {
    @GetMapping("/transfer")
    public TransferResponse transfer(@RequestBody TransferRequest request) {
        // Worker 스레드가 런큐에서 실행 시작

        Account account = accountRepository.findById(request.getAccountId());
        // → Worker 스레드가 DB 소켓의 웨이트 큐로 이동
        // → 해당 스레드는 런큐에서 제거됨

        ExternalResult result = externalApiClient.call(account.getParams());
        // → Worker 스레드가 HTTP 소켓의 웨이트 큐로 이동
        // → 또다시 런큐에서 제거됨

        transactionRepository.save(result.toEntity());
        // → Worker 스레드가 DB 소켓의 웨이트 큐로 이동

        return new TransferResponse(result);
        // 최종적으로 Worker 스레드가 스레드 풀로 반환
    }
}
```

[Tomcat NIO connector with a blocking application](https://stackoverflow.com/a/28816417):

> "While there are performance differences between Tomcat Connectors, the difference in raw request/response time is pretty small when the servlet itself blocks. However, the difference in the number of simultaneous requests that Tomcat can handle is vastly different when you use non-blocking I/O."

개별 요청의 응답 시간은 NIO와 BIO 차이 미미하지만, 동시 처리 가능한 요청 수에서 NIO가 압도적으로 유리합니다.

WebFlux를 사용하여 전체 요청 처리 과정을 논블로킹으로 변경하면 근본적으로 해결할 수 있습니다.

```java
// WebFlux: 완전한 논블로킹
@GetMapping("/transfer")
public Mono transfer(@RequestBody TransferRequest request) {
    return accountRepository.findById(request.getAccountId())    // 논블로킹 R2DBC
        .flatMap(account ->
            externalApiClient.call(account.getParams())          // 논블로킹 WebClient
                .flatMap(result ->
                    transactionRepository.save(result.toEntity()) // 논블로킹 R2DBC
                        .map(saved -> new TransferResponse(result))
                )
        );
    // 전 과정에서 이벤트 루프 스레드가 런큐에 머무름
}
```

단계적으로 개선할 수 있다고 합니다.
1. NIO 커넥터로 연결 관리 최적화
2. 비동기 서블릿 (Servlet 3.0)으로 Worker 해제
3. WebFlux + R2DBC로 완전한 논블로킹 달성

```java
// 2단계: 비동기 서블릿으로 Worker 해제
@GetMapping("/transfer")
public DeferredResult transfer(@RequestBody TransferRequest request) {
    DeferredResult deferredResult = new DeferredResult<>();

    // Worker 스레드를 즉시 해제하고 별도 스레드에서 처리
    CompletableFuture.supplyAsync(() -> {
        // 블로킹 작업들...
        return processTransfer(request);
    }).whenComplete((result, ex) -> {
        if (ex != null) {
            deferredResult.setErrorResult(ex);
        } else {
            deferredResult.setResult(result);
        }
    });

    return deferredResult;  // Worker 스레드 즉시 반환
}
```

## 기타

- [서버 부하 테스트 및 최적화(2)](https://velog.io/@kts5927/%EC%88%98%EC%A0%95%EC%A4%91%EC%84%9C%EB%B2%84-%EB%B6%80%ED%95%98-%ED%85%8C%EC%8A%A4%ED%8A%B8-%EB%B0%8F-%EC%B5%9C%EC%A0%81%ED%99%942)
- [Connecting Tomcat to the World](http://events17.linuxfoundation.org/sites/events/files/slides/Tomcat%20Connectors.pdf)
