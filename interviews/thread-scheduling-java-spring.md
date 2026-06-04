# OS 스레드, Java 스레드, Spring 스케줄링 실행 모델

- [OS 스레드, Java 스레드, Spring 스케줄링 실행 모델](#os-스레드-java-스레드-spring-스케줄링-실행-모델)
    - [면접에서 먼저 말할 답](#면접에서-먼저-말할-답)
    - [이 질문이 헷갈리는 이유](#이-질문이-헷갈리는-이유)
    - [전체 지도: OS에서 Spring까지](#전체-지도-os에서-spring까지)
    - [OS와 커널은 스레드를 어떻게 본다](#os와-커널은-스레드를-어떻게-본다)
    - [Java 스레드는 OS 스레드와 어떻게 이어지는가](#java-스레드는-os-스레드와-어떻게-이어지는가)
    - [`Thread.sleep()`은 정확히 무엇을 멈추는가](#threadsleep은-정확히-무엇을-멈추는가)
    - [Java의 `Executor`와 스케줄러는 스레드와 무엇이 다른가](#java의-executor와-스케줄러는-스레드와-무엇이-다른가)
    - [Spring에서 스레드는 어디서 생기고 어디서 쓰이는가](#spring에서-스레드는-어디서-생기고-어디서-쓰이는가)
    - [`@Scheduled`는 어떤 스레드에서 메서드를 호출하는가](#scheduled는-어떤-스레드에서-메서드를-호출하는가)
    - [`fixedDelay`, `fixedRate`, `cron`을 시간선으로 이해하기](#fixeddelay-fixedrate-cron을-시간선으로-이해하기)
    - [시나리오 해체](#시나리오-해체)
    - [설계 판단: `while + sleep`을 어디에 둘 것인가](#설계-판단-while--sleep을-어디에-둘-것인가)
    - [관측과 검증 방법](#관측과-검증-방법)
    - [면접 꼬리 질문 대비](#면접-꼬리-질문-대비)
    - [정리](#정리)
    - [근거와 더 읽을 자료](#근거와-더-읽을-자료)

## 면접에서 먼저 말할 답

`Thread.sleep()`은 현재 실행 중인 Java 스레드 하나를 일정 시간 쉬게 하는 호출입니다. 프로세스 전체가 멈추는 것도 아니고, CPU 코어 전체가 잠기는 것도 아닙니다. 다만 그 스레드가 맡고 있던 Java 호출 흐름은 아직 끝나지 않습니다.

Spring `@Scheduled` 메서드 안에서 `Thread.sleep()`을 호출하면, 잠드는 쪽은 그 `@Scheduled` 메서드를 실행하던 스케줄러 스레드입니다. 메서드가 반환되지 않았으므로 Spring 입장에서는 이번 실행이 아직 진행 중입니다. 기본 Spring Boot 설정에서는 스케줄러 풀이 보통 스레드 1개로 시작하므로, 한 scheduled 작업이 오래 자거나 무한 `while` 안에 머물면 같은 스케줄러를 쓰는 다른 scheduled 작업도 밀릴 수 있습니다. 다만 애플리케이션이 직접 `TaskScheduler`, `ScheduledExecutorService`, `SchedulingConfigurer`, `@Scheduled(scheduler = "...")`를 지정했다면 그 설정을 먼저 봐야 합니다.

면접에서 더 짧게 말하면 이렇게 답할 수 있습니다.

> `Thread.sleep()`은 현재 스레드만 `TIMED_WAITING` 상태로 보내고, 그 메서드 호출은 아직 끝나지 않습니다. `@Scheduled` 안에서 sleep하면 스케줄러 스레드가 그 호출 흐름 안에 머물기 때문에 다음 실행은 설정에 따라 늦어질 수 있습니다. 멀티코어 서버라고 해서 같은 scheduled 작업이 자동으로 새 스레드에서 겹쳐 실행되지는 않습니다. 반복 주기는 Spring 스케줄러와 그 스레드 풀 설정, 그리고 `fixedDelay`인지 `fixedRate`인지에 따라 해석해야 합니다.

## 이 질문이 헷갈리는 이유

이 주제는 단어 하나가 여러 계층에서 조금씩 다른 뜻으로 쓰이기 때문에 헷갈립니다.

| 말 | 이 문서에서의 의미 |
| --- | --- |
| 스레드 | 실행 흐름입니다. OS가 스케줄링하는 스레드일 수도 있고, Java 런타임이 관리하는 가상 스레드일 수도 있습니다. |
| 작업 | 실행할 코드 조각입니다. Java에서는 보통 `Runnable`, `Callable`, Spring scheduled method 같은 형태로 나타납니다. |
| 스레드 풀 | 여러 작업이 빌려 쓰는 스레드 묶음입니다. 작업 수와 스레드 수는 같지 않습니다. |
| 스케줄러 | 지금 실행할 작업을 시간 규칙에 따라 제출하거나 실행시키는 주체입니다. |
| `sleep` | 현재 스레드가 일정 시간 CPU를 쓰지 않고 기다리게 하는 동작입니다. |
| `@Scheduled` | Spring이 특정 메서드를 시간 규칙에 따라 호출하도록 등록하는 애노테이션입니다. |

질문에서 자주 섞이는 층위는 아래와 같습니다.

```text
CPU core
  ^
  | OS scheduler가 실행할 커널 스레드를 고름
  |
OS / 커널 스레드
  ^
  | JVM 플랫폼 스레드가 보통 1:1로 올라탐
  |
Java Thread
  ^
  | Executor / Scheduler가 작업을 어느 Java Thread에서 실행할지 정함
  |
Spring TaskScheduler
  ^
  | @Scheduled 메서드를 주기 규칙에 맞춰 호출
  |
애플리케이션 코드
  |
  +-- @Autowired 된 컴포넌트 메서드
  +-- while
  +-- Thread.sleep()
```

멀티코어 서버라는 말은 OS가 동시에 실행할 수 있는 CPU 실행 자원이 여러 개 있다는 뜻입니다. 하지만 어떤 Java 작업이 새 스레드로 제출되는지, 같은 scheduled 작업이 겹쳐 실행될 수 있는지, 한 스케줄러 풀이 몇 개의 스레드를 갖는지는 별도의 문제입니다. 하드웨어에 코어가 많다고 해서 Spring이 같은 메서드를 자동으로 여러 개 복제해서 실행하지는 않습니다.

## 전체 지도: OS에서 Spring까지

가장 작은 장면부터 보겠습니다. 아래 코드는 단순하지만, 질문의 핵심이 모두 들어 있습니다.

```java
@Component
class PollingJob {
    private final Worker worker;

    PollingJob(Worker worker) {
        this.worker = worker;
    }

    @Scheduled(fixedDelay = 5000)
    public void run() throws InterruptedException {
        worker.loop();
    }
}

@Component
class Worker {
    public void loop() throws InterruptedException {
        while (true) {
            doOnce();
            Thread.sleep(10_000);
        }
    }
}
```

이 코드는 아래처럼 실행됩니다.

```text
Spring scheduler thread
  -> PollingJob.run()
      -> Worker.loop()
          -> while
              -> doOnce()
              -> Thread.sleep(10초)
              -> doOnce()
              -> Thread.sleep(10초)
              -> ...
```

`Worker`가 `@Autowired`로 주입됐다는 사실은 여기서 새 스레드를 만들지 않습니다. `worker.loop()`는 평범한 Java 메서드 호출입니다. 호출한 스레드는 `loop()`가 끝날 때까지 돌아오지 못합니다. `loop()`가 무한 루프라면 `PollingJob.run()`도 끝나지 않습니다.

한 번 더 낮은 계층에서 보면 이렇게 됩니다.

```text
Spring TaskScheduler
  |
  | scheduled Runnable 실행
  v
Java 플랫폼 스레드: "scheduling-1"
  |
  | Worker.loop() 호출
  v
Thread.sleep(10_000)
  |
  | JVM이 현재 Java thread를 timed waiting 상태로 둠
  v
OS scheduler
  |
  | 이 커널 스레드를 당분간 CPU에 올리지 않음
  | 다른 runnable thread는 다른 코어 또는 나중의 time slice에서 실행 가능
  v
10초 뒤 깨울 수 있는 상태가 됨
```

여기서 `sleep`은 CPU를 계속 붙잡고 도는 바쁜 대기가 아닙니다. 자는 동안 그 스레드는 실행되지 않으며, 다른 스레드가 실행될 수 있습니다. 그러나 "그 스레드가 맡은 작업이 끝났다"는 뜻도 아닙니다. 잠든 작업은 아직 호출 스택 안에 남아 있습니다.

## OS와 커널은 스레드를 어떻게 본다

운영체제는 실행 가능한 작업을 CPU에 올렸다 내리며 시스템 전체를 움직입니다. 일반적인 서버 OS에서는 여러 프로세스와 스레드가 모두 동시에 실행되는 것처럼 보이지만, 실제로는 커널 스케줄러가 짧은 시간 단위로 실행 대상을 고릅니다. Linux `sched(7)`도 scheduler를 다음에 CPU에서 실행할 runnable thread를 결정하는 커널 컴포넌트로 설명합니다.[^linux-sched]

스레드를 OS 관점에서 보면 중요한 질문은 세 가지입니다.

1. 지금 실행 가능한가

    CPU를 받을 수 있는 스레드는 실행 가능 상태입니다. CPU 코어 수보다 실행 가능한 스레드가 많으면, 일부는 실행 대기 큐(queue)에서 자기 차례를 기다립니다. Linux CFS 문서는 runnable task를 실행 시간 순서의 자료구조에 두고 다음 실행 대상을 고르는 방식으로 설명합니다.[^linux-cfs]

2. 무엇을 기다리는가

    잠든 스레드, I/O 완료를 기다리는 스레드, 락을 기다리는 스레드는 당장 CPU에 올릴 필요가 없습니다. 커널이나 런타임은 조건이 만족될 때 다시 실행 가능한 상태로 옮깁니다.

3. 깨어난 즉시 실행되는가

    시간이 지났거나 I/O가 끝났다고 해서 바로 CPU에서 실행된다는 뜻은 아닙니다. 깨어난 스레드는 실행 후보가 되고, 실제 실행 시점은 OS 스케줄러와 시스템 부하에 따라 정해집니다. Linux `nanosleep(2)` 문서도 sleep 시간이 끝난 뒤에도 CPU가 다시 비어 호출 스레드를 실행하기까지 지연이 있을 수 있다고 설명합니다.[^linux-nanosleep]

멀티코어 환경에서는 여러 OS 스레드가 서로 다른 코어에서 동시에 실행될 수 있습니다. 그렇다고 해서 하나의 스레드가 여러 코어에서 동시에 실행되는 것은 아닙니다. 스레드 하나의 호출 스택은 한 번에 한 실행 흐름으로 진행됩니다.

```text
4-core server

core-1: Tomcat worker thread
core-2: GC thread
core-3: scheduling-1 thread  <- 지금 sleep 중이면 CPU에 없음
core-4: DB driver callback thread

sleep 중인 scheduling-1은 CPU를 태우지 않습니다.
하지만 scheduling-1이 맡은 @Scheduled 호출은 아직 끝난 것이 아닙니다.
```

## Java 스레드는 OS 스레드와 어떻게 이어지는가

Java의 플랫폼 스레드(platform thread)는 보통 OS가 스케줄링하는 커널 스레드(kernel thread)와 1:1로 이어집니다. Oracle의 Java `Thread` 문서도 플랫폼 스레드가 일반적으로 OS가 스케줄링하는 커널 스레드에 1:1로 매핑된다고 설명합니다.[^java-thread]

```text
Java Thread object
  |
  | JVM 내부 구조
  v
플랫폼 스레드
  |
  | 보통 1:1 대응
  v
OS 커널 스레드
  |
  | kernel scheduler가 CPU에 올림
  v
CPU core
```

Java 코드에서 `new Thread(runnable).start()`를 호출하면 새 실행 흐름이 시작됩니다. 반대로 `runnable.run()`을 직접 호출하면 새 스레드가 생기지 않고 현재 스레드에서 그냥 메서드가 실행됩니다. 이 차이는 Spring에서도 그대로 중요합니다.

```java
Runnable task = () -> work();

task.run();              // 현재 스레드에서 바로 실행
new Thread(task).start(); // 새 스레드에서 실행
executor.submit(task);    // executor가 가진 스레드 중 하나에서 실행
```

`@Autowired`된 빈의 메서드를 부르는 것은 기본적으로 첫 번째와 같습니다. 프록시, AOP, 트랜잭션 같은 Spring 기능이 끼더라도, `@Async`나 별도 executor 제출이 없다면 호출 스레드가 그 메서드를 끝까지 실행합니다.

Java 21 이후에는 가상 스레드(virtual thread)도 있습니다. 가상 스레드는 Java 런타임이 관리하는 가벼운 스레드이고, 많은 대기 작업을 적은 플랫폼 스레드 위에서 처리하기 위해 설계됐습니다. 하지만 가상 스레드를 쓴다고 해서 "동시 실행 정책"이 사라지는 것은 아닙니다. 어떤 작업을 몇 개 만들지, fixed delay를 어떻게 지킬지, 락을 잡은 채 오래 기다리는지 같은 문제는 여전히 설계해야 합니다.

## `Thread.sleep()`은 정확히 무엇을 멈추는가

`Thread.sleep()`은 현재 실행 중인 스레드를 지정한 시간 동안 쉬게 합니다. Java 문서는 `sleep`이 현재 실행 중인 스레드를 잠시 실행 중지시키며, 정확한 시간은 시스템 타이머와 스케줄러의 정밀도에 영향을 받는다고 설명합니다.[^java-thread-sleep] 또한 `sleep` 중에도 그 스레드가 가진 monitor lock은 풀리지 않습니다.[^java-thread-sleep]

Java 스레드 상태로 보면 `Thread.sleep()`에 들어간 스레드는 `TIMED_WAITING`입니다. `Thread.State` 문서는 `TIMED_WAITING`을 지정된 대기 시간이 있는 waiting 상태로 설명하고, 그 원인 중 하나로 `Thread.sleep`을 듭니다.[^java-state]

```text
RUNNABLE
  |
  | Thread.sleep(10_000)
  v
TIMED_WAITING
  |
  | 10초 경과 또는 interrupt
  v
RUNNABLE 후보
  |
  | OS/JVM 스케줄러가 실제 실행 기회를 줌
  v
다음 줄 실행
```

여기서 자주 틀리는 지점이 세 가지 있습니다.

1. `sleep`은 프로세스를 멈추지 않습니다.

    같은 JVM 안의 다른 스레드는 계속 실행될 수 있습니다. Tomcat 요청 스레드, DB pool housekeeping thread, GC thread, 다른 scheduler thread는 별개의 실행 흐름입니다.

2. `sleep`은 호출 스택을 끝내지 않습니다.

    `sleep`에서 깨어나면 그 다음 줄부터 계속 실행합니다. `@Scheduled` 메서드 안에서 잤다면, 이번 scheduled 메서드 실행은 아직 살아 있습니다.

3. `sleep`은 monitor lock을 풀지 않습니다.

    `synchronized` 안에서 `Thread.sleep()`을 호출하면 그 스레드는 잠들지만 lock은 계속 들고 있습니다. 다른 스레드가 같은 lock을 필요로 하면 `BLOCKED`가 될 수 있습니다. 조건을 기다리며 lock을 놓아야 하는 상황이면 `wait`, `Condition`, `BlockingQueue`, `Semaphore`, `CountDownLatch` 같은 도구를 검토해야 합니다.

```java
synchronized (lock) {
    Thread.sleep(10_000); // lock을 잡은 채 잠듭니다.
}
```

위 코드는 "CPU를 안 쓰니 안전하다"가 아닙니다. CPU는 안 쓰지만 lock은 잡고 있으므로 다른 스레드의 진행을 막을 수 있습니다.

하나 더 조심할 점이 있습니다. `sleep`은 메모리 가시성을 보장하는 동기화 도구가 아닙니다. 다른 스레드가 바꾼 값을 이 스레드가 반드시 다시 읽게 해 주는 장치가 아니라는 뜻입니다. Java 언어 명세도 `Thread.sleep`이나 `Thread.yield` 호출 전후로 컴파일러가 공유 값을 메모리에 내보내거나 다시 읽어야 하는 것은 아니라고 설명합니다.[^jls-sleep-yield]

```java
class Worker {
    private boolean running = true;

    void stop() {
        running = false;
    }

    void loop() throws InterruptedException {
        while (running) {
            doOnce();
            Thread.sleep(1000);
        }
    }
}
```

위 코드는 겉으로는 1초마다 `running`을 다시 볼 것 같지만, 스레드 사이의 변경을 안전하게 전달하는 계약이 없습니다. 이런 종료 신호는 최소한 `volatile boolean`, `AtomicBoolean`, interrupt, queue 종료 메시지처럼 스레드 사이 전달 규칙이 있는 방식으로 표현해야 합니다.

## Java의 `Executor`와 스케줄러는 스레드와 무엇이 다른가

Java에서 직접 스레드를 만들 수도 있지만, 서버 애플리케이션에서는 보통 `ExecutorService`나 스레드 풀을 씁니다. `ExecutorService`는 비동기 작업의 진행을 추적할 `Future`를 만들고, 종료를 관리하는 API를 제공합니다.[^executor-service]

핵심 구분은 이렇습니다.

```text
Thread
  실행 흐름 자체

Runnable / Callable
  실행할 일의 내용

ExecutorService
  작업을 받아서 보유한 스레드에서 실행하는 관리자

ScheduledExecutorService / ScheduledThreadPoolExecutor
  작업을 시간 규칙에 따라 실행하는 관리자
```

스레드 풀 크기가 1이면 한 번에 실행할 수 있는 작업은 하나입니다. 작업 하나가 오래 자거나 I/O에서 오래 기다리면, 그 스레드는 당분간 다른 작업을 처리하지 못합니다. 스레드 풀 크기가 4이면 서로 다른 작업 네 개까지는 동시에 실행될 수 있습니다. 그래도 "같은 주기 작업의 이전 실행과 다음 실행이 자동으로 겹친다"는 결론으로 바로 가면 안 됩니다.

Java `ScheduledThreadPoolExecutor` 문서는 `scheduleAtFixedRate`나 `scheduleWithFixedDelay`로 등록된 주기 작업의 연속 실행이 서로 겹치지 않는다고 설명합니다.[^scheduled-executor] 실행이 늦어질 수는 있지만, 같은 periodic task의 다음 실행이 이전 실행 위에 자동으로 포개지는 구조는 아닙니다.

이 규칙은 사용자의 직관을 조정하는 데 중요합니다.

```text
"서버가 멀티코어다"
  -> 여러 스레드가 동시에 실행될 수 있다.

"스케줄러 풀이 여러 스레드다"
  -> 서로 다른 scheduled 작업이 동시에 실행될 수 있다.

"같은 @Scheduled 메서드가 자동으로 겹쳐 돈다"
  -> 일반적인 단일 periodic task 모델에서는 아니다.
```

## Spring에서 스레드는 어디서 생기고 어디서 쓰이는가

Spring 애플리케이션에는 여러 종류의 스레드가 함께 살아 있습니다. 모두 Spring 빈 메서드를 실행할 수 있지만, 출발점과 관리자가 다릅니다.

| 실행 경로 | 누가 스레드를 관리하는가 | 대표 장면 |
| --- | --- | --- |
| HTTP 요청 처리 | Tomcat, Jetty, Undertow, Netty 같은 웹 서버 런타임 | `@RestController` 호출 |
| `@Async` | Spring `TaskExecutor` 또는 애플리케이션이 등록한 executor | 오래 걸리는 일을 별도 작업으로 넘김 |
| `@Scheduled` | Spring `TaskScheduler` | 주기 작업, 배치성 polling |
| 직접 만든 executor | 애플리케이션 코드 | 특정 외부 API 호출, 파일 처리, CPU 작업 분리 |
| JVM 내부 스레드 | JVM | GC, JIT, reference 처리 |

이 구분은 장애 분석에서 특히 중요합니다. 같은 `Thread.sleep()`이라도 Tomcat 요청 스레드에서 자면 HTTP 응답이 늦어지고, scheduler thread에서 자면 scheduled 작업이 밀리며, 직접 만든 worker pool에서 자면 그 풀의 처리량이 줄어듭니다.

```text
HTTP request thread에서 sleep
  -> 해당 요청이 늦어짐
  -> Tomcat worker pool이 고갈될 수 있음

@Scheduled scheduler thread에서 sleep
  -> 해당 scheduled 실행이 끝나지 않음
  -> scheduler pool이 작으면 다른 scheduled 작업도 밀림

별도 Executor worker에서 sleep
  -> 그 worker 하나가 묶임
  -> scheduler는 빨리 반환될 수 있지만 worker pool 큐가 쌓일 수 있음
```

## `@Scheduled`는 어떤 스레드에서 메서드를 호출하는가

Spring의 scheduling 기능은 `@EnableScheduling` 또는 관련 설정을 통해 켜집니다. Spring 문서는 `@Scheduled`와 `@Async`를 활성화하는 예시를 보여 주며, `@Async`에는 executor가, `@Scheduled`에는 scheduler가 연결된다는 점을 분리해 설명합니다.[^spring-scheduling]

`@Scheduled` 메서드는 보통 Spring이 등록한 scheduler가 시간 조건에 맞춰 호출합니다. Spring Boot에서는 scheduled task execution에 필요한 scheduler를 자동 구성할 수 있고, 가상 스레드를 켜지 않은 일반 설정에서는 `ThreadPoolTaskScheduler`가 기본적으로 사용되며 기본 스레드 수는 1개입니다. 이 값은 `spring.task.scheduling.pool.size`로 조정할 수 있습니다.[^spring-boot-scheduling]

```yaml
spring:
  task:
    scheduling:
      thread-name-prefix: "scheduling-"
      pool:
        size: 2
```

Spring Framework XML 예시에서도 scheduler pool size를 지정할 수 있고, 지정하지 않으면 기본 scheduler thread pool은 단일 스레드라고 설명합니다.[^spring-scheduling-pool]

따라서 기본 감각은 아래처럼 잡으면 됩니다.

```text
@Scheduled 메서드 등록
  -> Spring scheduler가 trigger를 관리
  -> trigger가 "실행할 때"라고 판단
  -> scheduler thread가 메서드 호출
  -> 메서드가 return하면 그 실행이 끝남
```

`@Autowired`된 컴포넌트는 여기서 새 실행 경계가 아닙니다.

```java
@Scheduled(fixedRate = 1000)
public void job() {
    service.work(); // 기본적으로 같은 scheduler thread에서 이어서 실행
}
```

`service.work()`가 30초 걸리면 `job()`도 30초 걸립니다. `service.work()`가 무한 루프면 `job()`도 반환되지 않습니다.

## `fixedDelay`, `fixedRate`, `cron`을 시간선으로 이해하기

Spring 문서는 `fixedDelay`를 이전 실행이 끝난 시점부터 다음 실행까지의 간격으로 설명하고, `fixedRate`를 연속 실행의 시작 시점 사이 간격으로 설명합니다.[^spring-scheduling] 이 차이는 `sleep`이나 긴 작업이 들어갔을 때 바로 드러납니다.

### `fixedDelay`

```java
@Scheduled(fixedDelay = 5000)
void poll() {
    work(); // 12초 걸림
}
```

```text
time ->
0s        12s       17s       29s       34s
| work 12 | delay 5 | work 12 | delay 5 |
```

`fixedDelay`는 "이번 실행이 끝난 뒤 5초 후"를 뜻합니다. `work()`가 오래 걸리면 전체 주기도 같이 길어집니다.

### `fixedRate`

```java
@Scheduled(fixedRate = 5000)
void poll() {
    work(); // 12초 걸림
}
```

```text
원래 기대 시작 시점:
0s   5s   10s  15s  20s ...

실제 실행:
0s        12s       24s
| work 12 | work 12 | work 12 ...
```

`fixedRate`는 시작 시점 기준입니다. 하지만 실행 시간이 period보다 길면 다음 실행은 제시간에 시작하지 못하고 늦어집니다. 일반적인 `ScheduledThreadPoolExecutor` 기반 주기 작업에서는 같은 작업의 연속 실행이 겹치지 않습니다.[^scheduled-executor]

### `cron`

`cron`은 벽시계 시간 규칙에 맞춰 다음 실행 시점을 계산합니다. 예를 들어 매분 0초에 실행하도록 걸면 기준은 "이전 실행이 끝난 뒤 몇 초"가 아니라 "벽시계상 매분 0초"입니다. 다만 놓친 tick이 무조건 큐에 쌓인다고 이해하면 안 됩니다. 이전 실행이 아직 끝나지 않았을 때 실제로 겹치는지는 scheduler 구현, pool size, 같은 메서드에 여러 schedule이 붙었는지, `@Async`로 넘겼는지, JVM 인스턴스가 여러 개인지에 따라 나눠 봐야 합니다.

Spring `@Scheduled` 문서는 같은 메서드에 여러 scheduled 선언이 있으면 각각 독립적으로 처리되며, 서로 겹치거나 병렬 실행될 수 있다고 설명합니다.[^spring-scheduled] 이 경우는 단일 periodic task가 "자동으로 겹치지 않는다"는 Java scheduler 감각과 다른 장면입니다. 같은 메서드에 schedule을 여러 개 붙였기 때문에 별도 trigger들이 생기는 것입니다.

## 시나리오 해체

### 1. `@Scheduled` 메서드가 한 번 일하고 바로 반환한다

```java
@Scheduled(fixedDelay = 10_000)
public void poll() {
    worker.doOneBatch();
}
```

```text
scheduling-1
  -> poll()
      -> doOneBatch()
  <- return

10초 뒤 다시 실행 후보
```

이 구조가 가장 단순합니다. 반복은 Spring scheduler가 맡고, 애플리케이션 메서드는 한 번 처리하고 끝납니다. 종료, 지연, 예외, 관측이 모두 쉽습니다.

### 2. `@Scheduled` 메서드 안에서 `Thread.sleep()`을 한 번 호출한다

```java
@Scheduled(fixedDelay = 10_000)
public void poll() throws InterruptedException {
    worker.doSomething();
    Thread.sleep(30_000);
}
```

```text
scheduling-1
  -> poll()
      -> doSomething()
      -> sleep 30초
      -> return
  -> fixedDelay 10초 대기
  -> 다음 poll()
```

`fixedDelay`에서는 sleep 시간도 이번 실행 시간에 포함됩니다. 위 코드는 대략 `doSomething 시간 + 30초 + 10초`마다 돕니다.

`fixedRate`라면 다음 시작 시점은 더 일찍 예약될 수 있지만, 이전 실행이 끝나지 않았으면 실제 실행은 밀립니다.

### 3. `@Autowired`된 컴포넌트 메서드 안에서 `while + sleep`을 돈다

```java
@Scheduled(fixedDelay = 5000)
public void run() throws InterruptedException {
    worker.loop();
}

public void loop() throws InterruptedException {
    while (true) {
        doOnce();
        Thread.sleep(10_000);
    }
}
```

이 장면이 사용자가 물은 핵심입니다.

```text
scheduling-1
  -> run()
      -> worker.loop()
          -> while(true)
              -> doOnce()
              -> sleep
              -> doOnce()
              -> sleep
              -> ...

run()이 return하지 않음
```

`worker`가 Spring 빈이어도 새 스레드가 생기지 않습니다. `worker.loop()`는 현재 scheduler thread에서 계속 실행됩니다. `while(true)`가 끝나지 않으면 `@Scheduled`의 첫 실행도 끝나지 않습니다. 다음 `fixedDelay` 실행은 "이전 실행 완료"를 볼 수 없고, 같은 scheduled 작업은 다시 시작되지 않는 것처럼 보입니다.

멀티코어 서버에서도 이 결론은 바뀌지 않습니다. 멀티코어는 다른 스레드가 실행될 가능성을 높일 뿐, 이미 `worker.loop()` 안에 들어간 호출을 자동으로 끊어서 다른 scheduled invocation을 만들지 않습니다.

### 4. scheduler pool size가 1이고 scheduled 작업이 여러 개 있다

```java
@Scheduled(fixedRate = 1000)
void jobA() throws InterruptedException {
    Thread.sleep(60_000);
}

@Scheduled(fixedRate = 1000)
void jobB() {
    log.info("B");
}
```

기본 scheduler thread가 하나라면 이렇게 볼 수 있습니다.

```text
scheduling-1
  -> jobA()
      -> sleep 60초
  -> jobB() 실행 기회가 뒤로 밀림
```

`jobA`가 자는 동안 CPU를 계속 쓰는 것은 아니지만, scheduler thread 하나가 `jobA` 호출 스택 안에 묶여 있습니다. 그 스레드가 `jobB`를 실행할 수 없습니다.

### 5. scheduler pool size가 2 이상이다

```yaml
spring:
  task:
    scheduling:
      pool:
        size: 2
```

```text
scheduling-1 -> jobA() sleep
scheduling-2 -> jobB() 실행 가능
```

풀을 키우면 서로 다른 scheduled 작업이 동시에 실행될 수 있습니다. 하지만 이 말은 "같은 jobA가 계속 겹쳐 실행된다"와 다릅니다. 일반적인 주기 작업 모델에서는 jobA의 이전 실행과 다음 실행은 겹치지 않는 쪽으로 이해해야 합니다.[^scheduled-executor]

다만 아래 경우에는 겹침을 따로 경계해야 합니다.

- 같은 메서드에 여러 `@Scheduled` 선언이 붙어 독립 trigger가 여러 개 생긴 경우
- `@Scheduled` 메서드가 내부에서 `@Async` 또는 별도 executor로 작업을 넘기고 바로 반환하는 경우
- 애플리케이션 인스턴스가 여러 개 떠 있어 각 JVM이 같은 scheduled job을 실행하는 경우
- 직접 `TaskScheduler`나 `Executor`에 같은 작업을 여러 번 등록한 경우

### 6. `@Scheduled`에서 `@Async` 또는 별도 executor로 넘긴다

```java
@Scheduled(fixedRate = 1000)
public void tick() {
    worker.runAsync(); // @Async가 실제로 적용된 public method라고 가정
}

@Async
public void runAsync() throws InterruptedException {
    Thread.sleep(10_000);
}
```

이 구조에서는 scheduler thread가 빨리 반환될 수 있습니다.

```text
scheduling-1
  -> tick()
      -> async executor에 작업 제출
  <- return

async-1 -> runAsync() sleep 10초
async-2 -> 다음 runAsync() sleep 10초
async-3 -> 다음 runAsync() sleep 10초
...
```

이제 문제는 scheduler가 아니라 async executor로 옮겨갑니다. 매초 작업을 제출하고 각 작업이 10초 걸리면 동시에 10개 안팎의 작업이 쌓일 수 있습니다. executor의 큐, 스레드 수, 외부 시스템 부하, 중복 실행 가능성, 그리고 여러 번 실행돼도 한 번 실행된 것과 같은 결과가 되는지, 즉 멱등성을 반드시 봐야 합니다.

그리고 `@Async`는 프록시를 통해 호출될 때 적용됩니다. 같은 클래스 안에서 자기 메서드를 직접 부르면 프록시를 거치지 않아 비동기 실행이 되지 않을 수 있습니다. Spring 문서도 기본 proxy 모드에서는 같은 클래스 내부 호출이 intercept되지 않는다고 설명합니다.[^spring-scheduling]

### 7. 여러 서버 인스턴스가 같은 `@Scheduled`를 실행한다

```text
pod-1 -> @Scheduled cleanup()
pod-2 -> @Scheduled cleanup()
pod-3 -> @Scheduled cleanup()
```

Spring scheduler는 JVM 내부의 스케줄러입니다. Kubernetes, systemd, 어떤 agent가 Spring Boot 프로세스를 시작했는지는 보통 JVM 내부의 단일 scheduler 동작을 바꾸지 않습니다. 하지만 프로세스가 여러 개면 이야기가 달라집니다. 각 프로세스가 자기 scheduler를 갖기 때문에 같은 작업이 인스턴스 수만큼 실행될 수 있습니다.

```text
단일 JVM 안에서의 중복 실행 문제
  -> scheduler pool, fixedRate/fixedDelay, @Async, lock으로 분석

여러 JVM 사이의 중복 실행 문제
  -> 분산 락, leader election, job 전용 인스턴스, DB unique constraint로 분석
```

면접에서 "agent가 Spring Boot 프로세스를 실행한다"는 말이 나오면 먼저 물어볼 것은 agent 종류보다 프로세스 수입니다. 하나의 JVM 안에서 sleep이 문제인지, 여러 JVM이 같은 job을 각각 실행하는 문제인지 분리해야 합니다.

### 8. `synchronized` 안에서 `sleep`한다

```java
public void update() throws InterruptedException {
    synchronized (lock) {
        Thread.sleep(10_000);
        state++;
    }
}
```

```text
thread-A
  -> lock 획득
  -> sleep 10초, lock은 계속 보유

thread-B
  -> synchronized(lock) 진입 시도
  -> BLOCKED
```

이 시나리오는 "sleep은 CPU를 안 쓰니까 괜찮다"는 오해를 깨 줍니다. CPU는 안 쓰지만 lock을 오래 들고 있으면 다른 스레드의 진행을 막습니다. 조건을 기다려야 한다면 lock을 잡은 채 자는 대신, 조건 대기 도구를 써서 lock과 wake-up 정책을 명확히 해야 합니다.

### 9. sleep 중 shutdown 또는 interrupt가 들어온다

`Thread.sleep()` 중인 스레드는 interrupt를 받으면 `InterruptedException`을 던지고 깨어납니다. Java 문서는 이때 interrupted status가 cleared 된다고 설명합니다.[^java-thread-sleep] 그래서 catch 블록에서 예외를 삼키고 루프를 계속 돌면, 종료 요청을 무시하는 worker가 될 수 있습니다.

```java
while (running) {
    try {
        doOnce();
        Thread.sleep(10_000);
    } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
        return;
    }
}
```

`@Scheduled` 메서드 안에서 무한 루프를 돌고 있다면 shutdown 때도 같은 문제가 드러납니다. scheduler가 "다음 실행을 하지 않는다"와 "이미 실행 중인 긴 루프를 안전하게 끝낸다"는 다른 문제입니다. 오래 살아 있는 loop를 둬야 한다면 `running` 플래그, interrupt 처리, executor 종료 대기, 외부 I/O timeout을 함께 설계해야 합니다.

### 10. 가상 스레드를 켜면 해결되는가

Spring Boot는 Java 21 이상에서 가상 스레드를 켜면 scheduled task execution에 `SimpleAsyncTaskScheduler`를 사용할 수 있다고 설명합니다. 이 scheduler는 가상 스레드와 맞춰져 있으며, fixed-delay 작업은 단일 scheduler thread에서 동작한다는 예외도 함께 설명됩니다.[^spring-boot-scheduling][^spring-simple-async]

가상 스레드는 많은 대기 작업을 더 적은 플랫폼 스레드 자원으로 감당하게 해 줄 수 있습니다. 하지만 아래 문제를 자동으로 해결하지는 않습니다.

- 같은 job이 동시에 여러 번 실행되어도 되는가
- 외부 API에 중복 요청을 보내도 되는가
- DB row를 중복 처리해도 되는가
- lock을 잡은 채 오래 기다리는가
- fixed delay 의미상 이전 작업 완료를 기다려야 하는가

가상 스레드는 대기 비용을 낮추는 도구이지, 작업의 중복 실행 정책을 대신 정해 주는 도구가 아닙니다.

## 설계 판단: `while + sleep`을 어디에 둘 것인가

### 대부분의 scheduled 작업

반복은 scheduler에게 맡기고, 메서드는 한 번 일하고 반환하는 편이 좋습니다.

```java
@Scheduled(fixedDelayString = "${jobs.poll.delay:10s}")
public void poll() {
    worker.doOneBatch();
}
```

이 구조의 장점은 단순합니다.

- 실행 한 번의 시작과 끝이 분명합니다.
- `fixedDelay`와 `fixedRate` 의미가 그대로 살아 있습니다.
- 예외, 로그, 지표, thread dump 해석이 쉽습니다.
- 정상 종료에서 현재 실행만 마무리하면 됩니다.

### 정말 오래 살아 있는 loop가 필요한 경우

정말로 while loop가 필요하다면 `@Scheduled`보다는 별도 생명주기 컴포넌트나 executor를 검토합니다.

```java
@Component
class PollingWorker implements SmartLifecycle {
    private final ExecutorService executor = Executors.newSingleThreadExecutor();
    private volatile boolean running;

    @Override
    public void start() {
        running = true;
        executor.submit(this::loop);
    }

    private void loop() {
        while (running && !Thread.currentThread().isInterrupted()) {
            try {
                doOnce();
                Thread.sleep(10_000);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }
    }

    @Override
    public void stop() {
        running = false;
        executor.shutdownNow();
    }

    @Override
    public boolean isRunning() {
        return running;
    }
}
```

이 예시는 구조를 보여 주기 위한 최소형입니다. 실제 운영 코드에서는 executor를 직접 만들지 않고 Spring bean으로 관리하거나, 종료 대기, 예외 처리, 재시도 간격 조절, 지표, 중복 실행 방지, 외부 시스템 timeout을 함께 둡니다.

판단 기준은 아래처럼 잡으면 됩니다.

| 상황 | 더 자연스러운 구조 |
| --- | --- |
| N초마다 한 번 확인한다 | `@Scheduled`가 반복을 맡고 메서드는 한 번 처리 후 반환 |
| 계속 살아 있는 consumer가 필요하다 | 메시지 컨슈머, listener container, 별도 lifecycle worker |
| 오래 걸리는 작업을 주기적으로 시작한다 | `@Scheduled`는 트리거만 맡고 큐와 스레드 수가 제한된 executor로 위임 |
| 같은 작업이 겹치면 안 된다 | 단일 scheduler, lock, 상태 플래그, DB unique key, 분산 락 중 요구에 맞게 선택 |
| 여러 인스턴스 중 하나만 실행해야 한다 | leader election 또는 분산 락 |

## 관측과 검증 방법

이 주제는 로그 몇 줄만 넣어도 직접 확인할 수 있습니다.

```java
@Scheduled(fixedRate = 1000)
public void job() throws InterruptedException {
    log.info("start thread={}", Thread.currentThread().getName());
    Thread.sleep(3000);
    log.info("end thread={}", Thread.currentThread().getName());
}
```

기본 scheduler pool size가 1이면 로그는 대체로 겹치지 않고 밀립니다.

```text
12:00:00 start thread=scheduling-1
12:00:03 end   thread=scheduling-1
12:00:03 start thread=scheduling-1
12:00:06 end   thread=scheduling-1
```

`fixedDelay`로 바꾸면 종료 뒤 delay가 붙습니다.

```text
12:00:00 start
12:00:03 end
12:00:04 start   # fixedDelay=1000이면 종료 1초 뒤
```

`jcmd`로 JVM thread dump를 보면 sleep 중인 스레드가 `TIMED_WAITING`으로 보일 수 있습니다.

```sh
jcmd <pid> Thread.print
```

확인할 지점은 세 가지입니다.

1. 스레드 이름

    `scheduling-`, `task-`, `http-nio-`, `reactor-`, `ForkJoinPool`처럼 이름을 보면 어느 실행 경로인지 추정할 수 있습니다.

2. Java thread state

    `TIMED_WAITING`이면 `sleep`, timeout 있는 `wait`, timeout 있는 `join`, `parkNanos` 같은 대기일 수 있습니다. stack trace의 맨 위 호출을 함께 봐야 합니다.

3. 호출 스택

    `@Scheduled` 메서드 아래에 `Worker.loop()`와 `Thread.sleep()`이 보이면 scheduler thread가 그 루프 안에 잡혀 있는 것입니다.

설정을 바꿔 비교하는 것도 좋은 실험입니다.

```yaml
spring:
  task:
    scheduling:
      pool:
        size: 1
```

```yaml
spring:
  task:
    scheduling:
      pool:
        size: 2
```

pool size를 1에서 2로 바꿨을 때 다른 scheduled 작업이 살아나면 scheduler thread 고갈이 원인일 가능성이 큽니다. 반대로 같은 작업의 중복 처리, DB 중복 갱신, 외부 API 중복 호출이 보이면 pool을 키운 것이 문제를 가린 채 동시성 위험을 키웠을 수 있습니다.

## 면접 꼬리 질문 대비

### `Thread.sleep()`하면 CPU를 쓰나요?

자는 동안 그 스레드는 CPU에서 실행되지 않습니다. 바쁜 루프처럼 CPU를 태우는 대기가 아닙니다. 다만 깨어난 뒤 바로 실행된다는 보장은 없고, 스케줄러가 실행 기회를 줘야 다음 줄이 실행됩니다.

### `Thread.sleep()`하면 lock도 놓나요?

아닙니다. Java 문서는 `sleep` 중에도 monitor 소유권을 잃지 않는다고 설명합니다.[^java-thread-sleep] `synchronized` 안에서 sleep하면 lock을 잡은 채 쉽니다.

### `wait()`와 `sleep()`은 무엇이 다른가요?

`sleep()`은 단순히 현재 스레드를 시간 동안 쉬게 합니다. `wait()`는 객체 monitor와 연결된 조건 대기입니다. `wait()`는 monitor를 놓고 wait set에 들어가며, 다른 스레드의 `notify`/`notifyAll` 또는 timeout/interruption으로 깨어납니다. 따라서 조건이 바뀔 때까지 협력해야 하는 코드에서는 `sleep`으로 추측성 polling을 하기보다 조건 대기 도구를 쓰는 편이 맞습니다.

### 서버가 16코어면 scheduled 작업도 16개가 동시에 도나요?

아닙니다. 코어 수는 OS가 동시에 실행할 수 있는 하드웨어 자원입니다. scheduled 작업의 동시성은 Spring scheduler pool size, 등록된 trigger 수, `@Async` 사용 여부, 여러 JVM 인스턴스 여부가 결정합니다.

### `fixedRate`면 이전 실행이 끝나기 전에 다음 실행이 시작되나요?

일반적인 단일 periodic task 모델에서는 이전 실행과 다음 실행이 겹치지 않습니다. Java `ScheduledThreadPoolExecutor` 문서는 연속 실행이 겹치지 않는다고 설명합니다.[^scheduled-executor] 다만 같은 메서드에 여러 `@Scheduled`를 붙였거나, 내부에서 async 작업을 계속 제출하거나, 애플리케이션 인스턴스가 여러 개면 겹침이 생길 수 있습니다.

### `@Autowired`된 컴포넌트 메서드는 별도 스레드에서 실행되나요?

기본적으로 아닙니다. 의존성 주입은 객체 참조를 연결해 줄 뿐입니다. 호출한 스레드가 그 메서드를 실행합니다. 별도 스레드가 필요하면 `@Async`, executor 제출, 메시지 컨슈머, 별도 worker lifecycle처럼 실행 경계를 명시해야 합니다.

### `@Scheduled` 안에서 무한 루프를 돌면 왜 안 좋나요?

스케줄러가 반복 주기를 관리할 기회를 잃습니다. 실행 한 번이 끝나지 않으므로 `fixedDelay`의 의미도 사라지고, 종료도 어려워지며, scheduler thread가 묶입니다. 반복이 필요하면 `@Scheduled`가 반복을 맡게 하거나, 아예 오래 사는 worker로 설계하고 종료 신호를 분리하는 편이 낫습니다.

### 여러 서버에서 같은 `@Scheduled`가 도는 것은 어떻게 막나요?

JVM 하나 안의 scheduler 설정으로는 여러 서버 인스턴스 사이의 중복 실행을 막을 수 없습니다. 여러 인스턴스 중 하나만 실행해야 한다면 DB row lock, Redis lock, ShedLock 같은 분산 락 도구, leader election, job 전용 인스턴스 분리 같은 설계가 필요합니다. 그리고 lock만 믿지 말고 작업 자체가 중복 실행돼도 안전한지, 적어도 중복 실행을 탐지할 수 있는지도 함께 봐야 합니다.

## 정리

이 주제의 핵심은 "어느 계층의 실행 단위가 붙잡혀 있는가"입니다.

```text
Thread.sleep()
  -> 현재 Java thread 하나가 쉼
  -> CPU 전체나 JVM 전체가 멈추지는 않음
  -> 호출 스택은 끝나지 않음

@Scheduled
  -> Spring scheduler가 메서드를 시간 규칙에 맞춰 호출
  -> 메서드가 return해야 이번 실행이 끝남
  -> 기본 scheduler pool이 작으면 한 작업의 sleep이 다른 scheduled 작업을 밀 수 있음

while + sleep inside @Scheduled
  -> scheduler thread가 루프 안에 계속 머묾
  -> 다음 fixedDelay 실행은 사실상 오지 않음
  -> 멀티코어라고 자동 해결되지 않음

pool size 증가 또는 @Async
  -> 다른 작업 실행 가능성을 늘림
  -> 동시에 중복 작업, 큐 적체, 외부 시스템 부하라는 새 문제를 만들 수 있음
```

면접에서는 아래 순서로 답하면 안정적입니다.

1. `sleep`은 현재 스레드 하나만 멈춘다고 말합니다.
2. 그 스레드가 어떤 pool의 스레드인지 밝힙니다.
3. `@Scheduled` 메서드가 반환됐는지 확인합니다.
4. `fixedDelay`, `fixedRate`, scheduler pool size, `@Async`, 다중 인스턴스를 나눕니다.
5. 설계 대안으로 "한 번 일하고 반환하는 scheduled method"와 "별도 lifecycle worker"를 비교합니다.

## 근거와 더 읽을 자료

- [Oracle Java `Thread` API](https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/lang/Thread.html): 플랫폼 스레드, 가상 스레드, `sleep`, interruption의 공식 설명입니다.
- [Oracle Java `Thread.State` API](https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/lang/Thread.State.html): `RUNNABLE`, `BLOCKED`, `WAITING`, `TIMED_WAITING` 상태의 의미를 확인할 수 있습니다.
- [Oracle Java `ExecutorService` API](https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/util/concurrent/ExecutorService.html): 작업 제출, `Future`, shutdown, awaitTermination의 공식 설명입니다.
- [Oracle Java `ScheduledThreadPoolExecutor` API](https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/util/concurrent/ScheduledThreadPoolExecutor.html): periodic task의 연속 실행이 겹치지 않는다는 규칙을 확인할 수 있습니다.
- [Java Language Specification 17장](https://docs.oracle.com/javase/specs/jls/se21/html/jls-17.html): `sleep`이 monitor를 놓지 않는다는 점과 `sleep`/`yield`가 메모리 가시성을 보장하지 않는다는 점을 확인할 수 있습니다.
- [Linux `sched(7)` manual page](https://man7.org/linux/man-pages/man7/sched.7.html): Linux CPU scheduling의 runnable thread, scheduling policy, priority 감각을 확인할 수 있습니다.
- [Linux `nanosleep(2)` manual page](https://man7.org/linux/man-pages/man2/nanosleep.2.html): OS 수준의 sleep 계열 대기에서 시간 경과 후에도 실제 CPU 재실행까지 지연될 수 있음을 확인할 수 있습니다.
- [Linux Kernel CFS Scheduler 문서](https://docs.kernel.org/scheduler/sched-design-CFS.html): 일반 task scheduler가 runnable task를 어떤 기준으로 고르는지 더 깊게 볼 수 있습니다.
- [Spring Framework Task Execution and Scheduling](https://docs.spring.io/spring-framework/reference/integration/scheduling.html): `TaskScheduler`, `@Scheduled`, `@Async`, `fixedDelay`, `fixedRate`, scheduler pool 설정의 공식 설명입니다.
- [Spring `@Scheduled` Javadoc](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/scheduling/annotation/Scheduled.html): `@Scheduled` 메서드 제약, reactive support, repeatable schedule의 overlap 가능성을 확인할 수 있습니다.
- [Spring Boot Task Execution and Scheduling](https://docs.spring.io/spring-boot/reference/features/task-execution-and-scheduling.html): Spring Boot의 task execution/scheduling 자동 설정과 `spring.task.scheduling.pool.size`를 확인할 수 있습니다.

[^java-thread]: Oracle Java SE 26 `Thread` 문서는 플랫폼 스레드가 일반적으로 OS가 스케줄링하는 커널 스레드와 1:1로 매핑된다고 설명합니다. <https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/lang/Thread.html>
[^java-thread-sleep]: Oracle Java SE 26 `Thread.sleep` 문서는 현재 실행 중인 스레드가 잠시 실행을 멈추며, 시스템 타이머와 스케줄러 정밀도의 영향을 받고, monitor 소유권을 잃지 않는다고 설명합니다. <https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/lang/Thread.html>
[^java-state]: Oracle Java SE 26 `Thread.State` 문서는 `TIMED_WAITING`의 원인 중 하나로 `Thread.sleep`을 듭니다. 또한 Java thread state는 VM 상태이며 OS thread state를 그대로 반영하지 않는다고 설명합니다. <https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/lang/Thread.State.html>
[^executor-service]: Oracle Java SE 26 `ExecutorService` 문서는 `ExecutorService`가 종료 관리와 비동기 작업 진행 추적을 위한 `Future` 생성 기능을 제공한다고 설명합니다. <https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/util/concurrent/ExecutorService.html>
[^scheduled-executor]: Oracle Java SE 26 `ScheduledThreadPoolExecutor` 문서는 `scheduleAtFixedRate` 또는 `scheduleWithFixedDelay`로 등록된 periodic task의 연속 실행이 서로 겹치지 않는다고 설명합니다. <https://docs.oracle.com/en/java/javase/26/docs/api/java.base/java/util/concurrent/ScheduledThreadPoolExecutor.html>
[^jls-sleep-yield]: Java Language Specification 17장은 `Thread.sleep`이나 `Thread.yield` 호출 전후로 컴파일러가 공유 값을 메모리에 flush하거나 다시 읽어야 하는 것은 아니라고 설명합니다. <https://docs.oracle.com/javase/specs/jls/se21/html/jls-17.html>
[^linux-sched]: Linux `sched(7)` manual page는 scheduler를 다음에 CPU에서 실행할 runnable thread를 결정하는 커널 컴포넌트로 설명합니다. <https://man7.org/linux/man-pages/man7/sched.7.html>
[^linux-cfs]: Linux Kernel CFS Scheduler 문서는 runnable task를 가상 실행 시간 기준으로 정렬하고, 그중 다음 실행 대상을 고르는 설계를 설명합니다. <https://docs.kernel.org/scheduler/sched-design-CFS.html>
[^linux-nanosleep]: Linux `nanosleep(2)` manual page는 sleep 시간이 끝난 뒤에도 CPU가 호출 스레드를 다시 실행하기까지 지연이 있을 수 있다고 설명합니다. <https://man7.org/linux/man-pages/man2/nanosleep.2.html>
[^spring-scheduling]: Spring Framework "Task Execution and Scheduling" 문서는 `@Scheduled`, `@Async`, `fixedDelay`, `fixedRate`, scheduler/executor 설정을 설명합니다. <https://docs.spring.io/spring-framework/reference/integration/scheduling.html>
[^spring-scheduling-pool]: 같은 Spring Framework 문서는 `<task:scheduler>`의 `pool-size`를 지정할 수 있으며, 지정하지 않으면 기본 scheduler thread pool이 단일 스레드라고 설명합니다. <https://docs.spring.io/spring-framework/reference/integration/scheduling.html>
[^spring-scheduled]: Spring `@Scheduled` Javadoc은 같은 메서드에 여러 scheduled declaration이 있으면 각각 독립적으로 처리되고 겹치거나 병렬 실행될 수 있다고 설명합니다. <https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/scheduling/annotation/Scheduled.html>
[^spring-boot-scheduling]: Spring Boot "Task Execution and Scheduling" 문서는 가상 스레드 설정이 없을 때 `ThreadPoolTaskScheduler`가 sensible defaults로 자동 구성되고, 기본적으로 한 스레드를 쓰며, `spring.task.scheduling.pool.size`로 조정할 수 있다고 설명합니다. <https://docs.spring.io/spring-boot/reference/features/task-execution-and-scheduling.html>
[^spring-simple-async]: Spring Framework 문서는 `SimpleAsyncTaskScheduler`가 JDK 21 가상 스레드와 맞춰져 있으며, fixed-delay 작업은 단일 scheduler thread에서 동작한다고 설명합니다. <https://docs.spring.io/spring-framework/reference/integration/scheduling.html>
