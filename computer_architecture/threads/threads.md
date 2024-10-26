# Thread

- [Thread](#thread)
    - [스레드](#스레드)
    - [스레드의 종류](#스레드의-종류)
        - [하드웨어 스레드](#하드웨어-스레드)
            - [Threaded Processor](#threaded-processor)
            - [Temporal multithreading](#temporal-multithreading)
                - [Coarse-Grained Multithreading(block or cooperative multithreading)](#coarse-grained-multithreadingblock-or-cooperative-multithreading)
                - [Fine-Grained Multithreading(Interleaved, preemptive, fine-grained or time-sliced multithreading)](#fine-grained-multithreadinginterleaved-preemptive-fine-grained-or-time-sliced-multithreading)
            - [Simultaneous Multithreading (SMT, 동시 멀티스레딩)](#simultaneous-multithreading-smt-동시-멀티스레딩)
                - [하이퍼스레딩 (Hyper-Threading)](#하이퍼스레딩-hyper-threading)
        - [소프트웨어 스레드](#소프트웨어-스레드)
            - [커널 수준 스레드 (Kernel-Level Threads, KLTs)](#커널-수준-스레드-kernel-level-threads-klts)
            - [사용자 수준 스레드 (User-Level Threads, ULTs)](#사용자-수준-스레드-user-level-threads-ults)
            - [하이브리드 스레딩](#하이브리드-스레딩)
            - [그린 스레드(Green Threads)](#그린-스레드green-threads)
            - [파이버 (Fibers)](#파이버-fibers)
    - [k8s와 멀티스레딩](#k8s와-멀티스레딩)
        - [k8s의 CPU 자원 할당](#k8s의-cpu-자원-할당)
        - [호스트 노드와 컨테이너](#호스트-노드와-컨테이너)
        - [Kubernetes의 CPU 할당과 커널 스레드 간의 상호작용](#kubernetes의-cpu-할당과-커널-스레드-간의-상호작용)
    - [Thread와 Join](#thread와-join)
    - [Thread detach](#thread-detach)
    - [쓰레드와 프로세스](#쓰레드와-프로세스)
        - [멀티 프로세스와 멀티 스레딩](#멀티-프로세스와-멀티-스레딩)
    - [각 언어별 스레드](#각-언어별-스레드)
        - [Java Virtual Thread, Kotlin coroutine, and Go goroutine](#java-virtual-thread-kotlin-coroutine-and-go-goroutine)
            - [Java 21 Virtual Threads](#java-21-virtual-threads)
            - [Kotlin Coroutines](#kotlin-coroutines)
            - [Go Goroutines](#go-goroutines)
            - [주요 차이점](#주요-차이점)
    - [멀티태스킹 운영체제와 GC `ParallelGCThreads` 설정](#멀티태스킹-운영체제와-gc-parallelgcthreads-설정)
        - [컨텍스트 스위칭 비용](#컨텍스트-스위칭-비용)
        - [자원 경합](#자원-경합)
        - [가비지 컬렉션 최적화](#가비지-컬렉션-최적화)
        - [스레드와 코어의 관계](#스레드와-코어의-관계)

## [스레드](https://en.wikipedia.org/wiki/Thread_(computing))

스레드(Thread)는 프로세스 내에서 실행되는 가장 작은 실행 단위입니다.
프로세스가 여러 작업을 동시에 수행할 수 있게 해주는 경량 프로세스(Lightweight Process)라고 볼 수 있습니다.

## 스레드의 종류

### 하드웨어 스레드

하드웨어 스레드는 CPU의 물리적인 구성 요소입니다.
- CPU가 *물리적으로 지원하는 명령어 실행의 단위*
- 또는 *동시에 수행할 수 있는 명령어 스트림의 수*

현대의 CPU는 다수의 하드웨어 스레드를 동시에 처리하거나, 각 스레드를 번갈아 실행하면서 자원을 효율적으로 사용합니다.

[하드웨어에서 스레드를 처리하는 방식](https://en.wikipedia.org/wiki/Multithreading_(computer_architecture))에는 크게 Temporal Multithreading(시간 기반 멀티스레딩)과 Simultaneous Multithreading(SMT, 동시 멀티스레딩) 두 가지가 있습니다.
Temporal Multithreading(시간 기반 멀티스레딩)은 다시 Coarse-grained와 Fine-grained 멀티쓰레딩으로 나뉩니다.
- [Temporal multithreading](https://en.wikipedia.org/wiki/Temporal_multithreading)
    - [Coarse-grained](https://en.wikipedia.org/wiki/Granularity_(parallel_computing))
    - [Fine-grained](https://en.wikipedia.org/wiki/Granularity_(parallel_computing)) (or interleaved)
- [Simultaneous multithreading](https://en.wikipedia.org/wiki/Simultaneous_multithreading)

#### Threaded Processor

Threaded processor는 일반적으로 *멀티스레딩(multithreading)을 지원하는 CPU*를 가리킵니다.

- Threaded Processor (스레드 지원 프로세서)란?

    멀티스레딩을 지원하는 CPU입니다. 멀티스레딩을 지원한다는 것은 한 번에 여러 스레드를 실행하거나, 스레드 간 빠르게 전환할 수 있는 능력을 갖춘 프로세서를 의미합니다.

    CPU가 여러 스레드를 처리할 수 있는 방법은 여러 가지가 있습니다. 대표적인 방식으로는 아래와 같은 것이 있습니다.

    - Coarse-Grained Multithreading: 한 스레드가 큰 지연(예: 메모리 접근)으로 인해 대기할 때, 다른 스레드로 전환하는 방식.
    - Fine-Grained Multithreading: 매 사이클마다 스레드 전환을 빠르게 수행하는 방식.
    - Simultaneous Multithreading (SMT): 여러 스레드가 동시에 프로세서의 여러 자원을 나누어 사용하는 방식(예: 하이퍼스레딩).

- Non-threaded Processor

    멀티스레딩을 지원하지 않는 프로세서는 단일 스레드(single-threaded)만을 실행할 수 있습니다.
    예전의 대부분의 RISC(Reduced Instruction Set Computer) 프로세서와 `CISC` 프로세서들이 이러한 구조를 따랐습니다.

    스레드를 지원하지 않는 프로세서에서는, 한 번에 하나의 스레드만 처리할 수 있습니다.
    따라서 하나의 스레드가 메모리 접근과 같은 긴 대기 상태에 들어가면 프로세서의 자원이 유휴 상태로 남게 되고, CPU 자원을 비효율적으로 사용할 수밖에 없습니다.

    예를 들어, 구형 프로세서나 특정 임베디드 시스템에서 사용되는 단일 스레드 프로세서가 이에 해당합니다.

#### Temporal multithreading

##### Coarse-Grained Multithreading(block or cooperative multithreading)

가장 간단한 유형의 멀티스레딩입니다.
하나의 스레드가 긴 지연 시간을 유발하는 이벤트(캐시 미스 등으로 오프칩 메모리에 접근 등)에 의해 차단될 때까지 실행될 때 발생합니다. 스레드 전환 기능을 지원하는 프로세서(threaded processor)는 멈춤이 해결될 때까지 기다리는 대신 실행 준비가 완료된 다른 스레드로 실행을 전환합니다.
즉, 한 스레드가 대기 상태에 빠져 있을 때만 전환이 이루어지며, 전환이 비교적 적게 발생합니다.

이전 스레드의 데이터가 도착했을 때만 이전 스레드가 [ready-to-run](https://en.wikipedia.org/wiki/Process_state#Ready) 스레드 목록에 다시 추가됩니다.

예를 들어,
1. Cycle i: A 쓰레드에서 instruction j 발생.
2. Cycle i + 1: A 쓰레드에서 instruction j + 1 발생.
3. Cycle i + 2: A 쓰레드에서 모든 캐시 미스되는 load instruction j + 2 발생.
4. Cycle i + 3: 쓰레드 스케쥴러가 호출되고 B 쓰레드로 전환
5. Cycle i + 4: B 쓰레드에서 instruction k 발생
6. Cycle i + 5: B 쓰레드에서 instruction k + 1 발생

개념적으로는 [실시간 운영 체제(RTOS)](https://en.wikipedia.org/wiki/Real-time_operating_system)에서 사용되는 협력적 멀티태스킹과 유사하며, 어떤 유형의 이벤트를 기다려야 할 때 작업들이 자발적으로 실행 시간을 포기하는 방식입니다.

전환 빈도가 비교적 낮고, 비용이 적으며, 단순한 구조를 가지고 있습니다.
스레드 전환은 주로 하드웨어의 긴 대기시간에 의존하기 때문에, 메모리 접근이 많은 작업에서 스레드가 자주 전환됩니다.

Coarse-Grained 방식은 CPU가 오랫동안 대기해야 할 때 자원을 낭비하지 않도록 하여,
지연 시간이 길거나 입출력(I/O) 대기 시간이 많은 작업에서 유용합니다.

##### Fine-Grained Multithreading(Interleaved, preemptive, fine-grained or time-sliced multithreading)

실행 [파이프라인](https://en.wikipedia.org/wiki/Pipeline_(computing))에서 모든 [데이터 종속성](https://en.wikipedia.org/wiki/Data_dependency) 지연(stall)을 제거하는 것이 목적입니다.
한 스레드는 다른 스레드로부터 비교적 독립적이기 때문에 파이프라인의 한 단계에서 처리되는 명령어가 이전 명령어의 출력을 필요로 하는 경우가 적어집니다.
개념적으로는 운영 체제에서 사용되는 선제적([preemptive](https://en.wikipedia.org/wiki/Preemption_(computing))) 멀티태스킹과 유사하며, 각 활성 스레드에 주어진 time slice가 하나의 CPU 사이클이라고 비유할 수 있습니다.

예를 들어,
1. Cycle i + 1: B 쓰레드로부터 instruction 발생.
2. Cycle i + 2: C 쓰레드로부터 instruction 발생.

이러한 유형의 멀티스레딩은 처음에 배럴 처리(barrel processing)라고 불렀는데, 배럴의 stave는 파이프라인 단계와 해당 실행 스레드를 나타냅니다.

![barrel-parts](./resources/barrel-parts.jpg)

블록 유형의 멀티스레딩에서 설명한 하드웨어 비용 외에도 각 파이프라인 단계에서 처리 중인 명령어의 스레드 ID를 추적하는 추가 비용이 있습니다.
또한 파이프라인에서 동시에 실행되는 스레드가 많아지기 때문에 캐시나 TLB와 같은 공유 리소스가 더 커야 서로 다른 스레드 간의 thrashing을 ​​방지할 수 있습니다.

전환이 매우 자주 일어나며, 매 사이클마다 스레드가 교체됩니다.
CPU의 자원 활용률을 극대화할 수 있기 때문에, 계산 작업이 많고 파이프라인 스톨을 줄여야 하는 경우에 사용됩니다.

#### Simultaneous Multithreading (SMT, 동시 멀티스레딩)

여러 스레드의 명령어가 동일한 CPU 사이클에서 병렬로 실행되는 것처럼 보이게 만듭니다.
논리적으로 동시에 실행되는 것으로 보이지만, 실질적으로는 CPU의 리소스가 나누어져 각 스레드가 병렬적으로 처리되는 것처럼 보이는 것입니다.

##### [하이퍼스레딩 (Hyper-Threading)](https://en.wikipedia.org/wiki/Hyper-threading)

Intel에 의해 개발된 기술로, 동시 멀티스레딩의 한 형태입니다.
진정한 물리적 동시성(parallel execution)은 아니며, 동시에 작업을 수행하는 것처럼 보이게 합니다.
실제로는 두 스레드가 CPU의 ALU(Arithmetic Logic Unit), FPU(Floating Point Unit), 캐시 메모리 등의 실행 유닛을 공유하면서 시분할(time-sharing) 방식으로 매우 빠른 속도로 스레드 간 전환이 이루어집니다.
따라서 두 개의 독립적인 물리적 코어만큼 성능 향상이 되지는 않습니다.

CPU 자원의 활용도를 개선하지만, 작업 종류에 따라 성능 향상 정도가 달라집니다.
- CPU를 논리적으로 두 개로 나누기 때문에 CPU 바운드 작업의 경우 하이퍼스레딩의 이점이 제한적일 수 있습니다.
- I/O 바운드 또는 다양한 유형의 작업이 혼합된 경우 하이퍼스레딩이 더 효과적일 수 있습니다.

예를 들어, 4코어 8스레드 CPU의 경우:
- 4개의 물리적 코어, 각 코어당 2개의 하드웨어 스레드로 총 8개의 스레드가 됩니다.
- 운영체제는 이를 8개의 논리적 프로세서로 인식합니다.

운영체제의 스레드 스케줄러가 소프트웨어 스레드를 하드웨어 스레드에 할당합니다.
예를 들어, 1000개의 소프트웨어 스레드가 있는 경우, 스케줄러가 이를 가용한 하드웨어 스레드에 동적으로 할당하고 관리합니다.

### 소프트웨어 스레드

소프트웨어 스레드는 운영체제가 관리하는 프로그램 실행의 단위로 운영체제의 스케줄러에 의해 직접 관리되고 스케줄링됩니다.
프로그램이나 프로세스 내에서 동시에 여러 작업을 수행하기 위해 생성됩니다.

각 스레드는 자체 스택을 가집니다.
힙 영역, 코드 영역, 데이터 영역과 같은 다른 자원은 프로세스 내에서 공유합니다.

#### 커널 수준 스레드 (Kernel-Level Threads, KLTs)

네이티브 스레드라고도 불립니다. 운영체제 커널에 의해 직접 관리되며 커널의 스케줄링 대상이 됩니다.
커널에 의해 관리되기 때문에, 멀티코어 환경에서 진정한 병렬성을 활용할 수 있습니다.
즉, 멀티코어 CPU에서는 여러 KLT가 각각의 물리적 코어에서 동시에 실행될 수 있습니다.

장점:
- 커널이 직접 스레드를 스케줄링하므로 효율적인 CPU 사용이 가능합니다.
- 한 스레드가 블로킹되어도 다른 스레드가 실행될 수 있습니다.

단점:
- 생성과 컨텍스트 스위칭에 더 많은 오버헤드가 발생합니다.

#### 사용자 수준 스레드 (User-Level Threads, ULTs)

사용자 공간에서 라이브러리에 의해 관리됩니다.
커널은 이 스레드의 존재를 알지 못합니다.

장점:

- 생성과 관리가 빠르고 효율적입니다.
- 사용자 공간에서 스케줄링되므로 컨텍스트 스위칭이 빠릅니다.

단점:

- 하나의 ULT가 블로킹되면 전체 프로세스가 블로킹될 수 있습니다.

    예를 들어, 사용자 수준에서 스케줄링되기 때문에, 커널이 이를 인식하지 못합니다.
    하나의 ULT가 I/O와 같은 블로킹 연산을 수행할 때 프로세스 전체가 대기 상태에 들어갈 수 있습니다.

- 멀티코어 시스템에서 진정한 병렬성을 활용하기 어렵습니다.

#### [하이브리드 스레딩](https://en.wikipedia.org/wiki/Thread_(computing)#M:N_(hybrid_threading))

'다대다(Many-to-Many) 스레딩 모델'이라고도 합니다.
M개의 사용자 수준 스레드를 N개의 커널 수준 스레드에 매핑합니다.
OS 스레드보다 더 가볍고, 병렬성을 쉽게 활용할 수 있습니다.
현실 세계의 시스템에서는 이 모델을 많이 사용합니다.

에를 들어, Go의 스케줄러는 고루틴을 M:N 모델로 관리합니다.
- M: 고루틴 수
- N: 커널 수준의 스레드(Kernel-Level Threads, KLT) 수

즉, 다수의 고루틴(M)을 소수의 커널 수준 스레드(N)에 매핑하여 실행합니다.

#### 그린 스레드(Green Threads)

가상 머신(예: JVM)에 의해 관리되는 스레드입니다.
사용자 수준 스레드의 한 형태로, 운영체제 스레드 대신 사용자 공간에서 직접 스케줄링되는 방식입니다.
초기 Java 버전에서 사용되었지만, 현재는 대부분의 JVM이 네이티브 스레드를 사용합니다.

#### 파이버 (Fibers)

협력적 멀티태스킹을 위한 매우 가벼운 스레드입니다.
운영 체제의 스케줄러가 아닌, 애플리케이션 코드가 명시적으로 스레드 간의 제어권을 양보(yield)하는 구조입니다.
이 때문에 컨텍스트 스위칭 비용이 매우 적고, 효율적입니다.
Windows OS와 일부 프로그래밍 언어에서 지원합니다.

## k8s와 멀티스레딩

### k8s의 CPU 자원 할당

k8s는 Linux 커널의 CFS(Completely Fair Scheduler) 스케쥴링 알고리즘을 사용하여 millicore 단위로 CPU 자원을 공정하게 분배합니다.

여기서 millicore는 컨테이너가 사용할 수 있는 *CPU 사용 비율*을 나타냅니다.
즉, *CPU 사용 시간을 기준*으로 한 *CPU 자원의 비율적 할당*을 의미합니다.
millicore 값은 *실제로 사용할 수 있는 CPU 시간을 제한*하는 방식으로 작동합니다.
전체 CPU 사용 시간에서 해당 컨테이너가 사용할 수 있는 비율을 의미하며, 컨테이너에 할당된 CPU 시간이 설정된 millicore 값을 넘지 않도록 보장합니다.

이러한 CPU 제한은 *스케줄러*와 *Linux cgroup*을 통해 적용됩니다.
*cgroup*은 리눅스 커널에서 리소스 제어를 제공하는 메커니즘으로, CPU, 메모리, I/O와 같은 자원을 제한하거나 우선순위를 설정할 수 있습니다.

예를 들어, `1000mil`는 다음과 같은 의미입니다.
- CPU의 100% 사용, 즉 하나의 CPU 코어에서 사용할 수 있는 전체 시간
- CPU 코어 하나를 풀 타임 사용할 수 있는 자원
- 1개의 CPU 코어 전체를 사용하는 것과 동일한 시간

`500mil`은 CPU 50% 사용을 의미하며, 컨테이너가 하나의 코어에서 전체 시간 중 절반만 사용할 수 있음을 나타냅니다.
`200mil`은 CPU 20% 사용을 의미하며, 컨테이너가 하나의 코어에서 20% 시간을 사용할 수 있음을 의미합니다.

이때 주의할 점은 CFS는 CPU 시간을 할당하는 것이지, 물리적 코어나 스레드의 수를 직접 제어하는 것은 아닙니다.
이 리소스 제한은 어떤 물리적 코어에서 실행되는지에 대한 직접적인 제어를 제공하지 않습니다.
따라서 컨테이너에 할당된 CPU 시간이 여러 물리적 코어에 걸쳐 분산될 수 있습니다.

즉, `1000mil`로 설정된 경우, 1개의 물리적 코어에서 100%를 사용하거나, 여러 코어에서 나누어 사용하더라도 *총합은 1000mil(1개 코어)*에 해당하는 만큼의 CPU 자원을 사용할 수 있게 됩니다.
이 제한은 컨테이너 내에서 실행되는 스레드나 프로세스들이 *시분할 방식(time-sharing)*으로 실행되면서 달성됩니다.

### 호스트 노드와 컨테이너

컨테이너는 단지 리눅스 네임스페이스(Linux namespaces), cgroup(control groups) 등의 커널 기능을 사용하여 리소스 사용을 격리하고 제한할 뿐이며, 스케줄링 자체는 여전히 호스트인 노드 서버의 커널 스케줄러가 담당합니다.

컨테이너 안에서 운영되는 애플리케이션은 컨테이너 자체가 마치 독립적인 운영체제처럼 동작하는 것처럼 보이지만,
제로는 호스트 노드(즉, 컨테이너가 실행되는 물리적/가상 머신)의 Linux 커널에 의해 관리됩니다.

- 호스트 물리 코어 10개 & `1000mil` 설정된 컨테이너 & 10개의 고루틴 경우:

    이때 10개의 고루틴은 1개의 코어에서 시분할로 병렬로 처리되는 것처럼 보일 수 있습니다.
    반면 10개의 고루틴이 10개의 코어에서 물리적으로 병렬로 처리될 수도 있습니다.

    - Pod:

        Host Node의 커널 스케줄러가 전체 CPU 스케줄링을 담당합니다.
        그리고 컨테이너들은 독립된 프로세스처럼 취급되어 커널의 프로세스 스케줄러에 의해 CPU 자원을 분배받습니다.

        `GOMAXPROCS=1`로 설정하여 Go 프로그램 내에서 동시에 실행될 수 있는 OS 스레드의 최대 수를 제어합니다.
        즉, 한 번에 하나의 OS 스레드만을 사용하여 고루틴을 실행한다는 것을 의미합니다.
        실제로는 추가적인 시스템 스레드(예: 가비지 컬렉션, 네트워크 폴링 등을 위한)가 존재할 수 있습니다.

    - Node(Host):

        호스트 시스템(노드)의 CFS는 컨테이너에 할당된 `1000mil`(1코어, 전체 10개 CPU 시간의 10%)의 CPU 시간을 관리합니다.
        CFS는 Go 프로그램의 OS 스레드를 10개의 물리적 코어 중 어느 것에서든 실행할 수 있습니다.
        실행 위치는 시스템의 전반적인 부하와 스케줄링 결정에 따라 동적으로 변경될 수 있습니다.

        즉, 하나의 OS 스레드는 노드의 여러 물리적 코어에서 시분할 방식으로 실행될 수 있습니다.
        예를 들어, 10ms 동안은 코어 1에서 실행되고, 다음 10ms는 코어 5에서 실행되고, 그 다음 10ms 동안은 코어 3에서 실행될 수 있습니다.

        또한 OS는 필요에 따라 Go 프로그램의 스레드를 다른 CPU 코어로 이동시킬 수 있습니다.

### Kubernetes의 CPU 할당과 커널 스레드 간의 상호작용

1. CPU 할당과 cgroup 사용

    k8s는 각 컨테이너가 사용할 수 있는 CPU 자원을 cgroup을 통해 제한합니다.
    cgroup은 컨테이너의 프로세스 및 스레드가 사용할 수 있는 CPU 자원을 제한하고, 커널이 스케줄링을 통해 CPU 시간을 조절합니다.

    즉, cgroup은 CPU 시간을 제한할 뿐, 스레드 수는 리눅스 커널의 프로세스 스케줄러 CFS(Completely Fair Scheduler)가 관리합니다.

    예를 들어, `500mil`로 설정된 컨테이너는 컨테이너 전체에 50%의 CPU 시간이 제한이 걸립니다.

2. 스레드 스케줄링

    컨테이너 내에서 생성된 스레드는 리눅스 커널의 CFS(Completely Fair Scheduler) 스케줄러에 의해 관리됩니다.
    k8s에서 설정한 CPU 제한은 컨테이너의 전체 CPU 사용 시간만 제한하며, 스레드의 개수와 직접적인 연관은 없습니다.

    즉, 여러 스레드가 생성되더라도, 스레드들이 사용할 수 있는 전체 CPU 시간은 cgroup에 의해 제한된 범위 내에서 공정하게 분배됩니다.

   예를 들어, 컨테이너가 여러 스레드를 생성하더라도, 스레드가 실행되는 CPU 시간은 cgroup에 의해 제한됩니다.
   컨테이너 내부에서 생성된 여러 스레드는 컨테이너에 할당된 전체 CPU 시간을 나누어 사용합니다.

3. 스레드와 M:N 맵핑

    k8s는 컨테이너 수준에서 리소스를 제한할 뿐, 애플리케이션의 스레드 관리는 전적으로 애플리케이션 자체나 OS 커널에 의존합니다.
    컨테이너 내부에서 애플리케이션은 전통적인 1:1 스레드 모델을 따릅니다.
    즉, 애플리케이션에서 생성한 사용자 스레드는 하나의 커널 스레드에 직접 매핑됩니다.

    고루틴과 같은 M:N 스레딩 모델을 사용하는 경우에는, M:N 스케줄링은 애플리케이션 레벨에서 이루어지므로, k8s의 리소스 관리와는 별개입니다. CPU 제한은 이러한 애플리케이션의 실행 시간에만 영향을 미칩니다.

    따라서 k8s의 CPU 제한은 고루틴처럼 애플리케이션 레벨의 M:N 스케줄링에 영향을 주지 않습니다.
    하지만 컨테이너에 할당된 CPU 시간이 부족하다면, 커널의 CFS 스케줄러에 의해 대기 상태에 들어갈 수 있습니다.

## Thread와 Join

```mermaid
graph TD
    A[메인 스레드 시작] --> B[스레드 생성]
    B --> C[메인 스레드 계속 실행]
    B --> D[새 스레드 실행]
    C --> E[Join 호출]
    D --> F[새 스레드 완료]
    E --> F
    F --> G[메인 스레드 계속]
```

```mermaid
sequenceDiagram
    participant Main as 메인 스레드
    participant New as 새 스레드
    participant OS as 운영 체제

    Main->>OS: 스레드 생성 요청
    OS->>New: 새 스레드 생성
    OS-->>Main: JoinHandle 반환
    Main->>Main: 다른 작업 수행
    Main->>+New: 새 스레드 시작
    Main->>Main: join() 호출
    Main->>OS: 대기 상태로 전환
    New->>-OS: 종료 신호
    OS->>Main: 깨우기
    Main->>Main: 실행 재개
```

## Thread detach

```mermaid
sequenceDiagram
    participant Main as 메인 스레드
    participant Original as 원본 스레드
    participant Detached as 분리된 스레드

    Main->>Original: spawn
    Original->>Detached: spawn
    Original-->>Main: 종료
    Note over Main,Original: 원본 스레드 JoinHandle drop
    Main->>Main: sleep
    Detached->>Detached: 계속 실행
    Note over Detached: "♫ Still alive ♫" 출력
    Main->>Main: 프로그램 종료
```

## 쓰레드와 프로세스

`pthread_create`는 POSIX 스레드를 생성하기 위한 함수로, 주로 리눅스에서는 GNU C 라이브러리(glibc)를 통해 구현됩니다.

glibc 소스 코드 저장소인 [GNU C Library (glibc)](https://sourceware.org/git/?p=glibc.git)에서 [`nptl/pthread_create.c` 파일](https://sourceware.org/git/?p=glibc.git;a=blob;f=nptl/pthread_create.c;h=1d3665d5edb684e3de0e070763914a90b47545c7;hb=HEAD)에 구현되어 있습니다.

`pthread_create`의 동작은 사용자 수준에서 스레드를 설정하고 커널 수준에서 이를 관리합니다.
리눅스에서는 주로 `clone()` syscall을 사용해 새로운 스레드를 생성합니다.

`clone()`은 새로운 실행 흐름을 만들며, 이를 통해 새로 만들어진 스레드는 같은 주소 공간을 공유하게 됩니다.
이로 인해 여러 스레드가 같은 프로세스 내에서 데이터를 공유할 수 있습니다.

1. 사용자 모드에서 스레드 생성 요청:

    `pthread_create`가 호출되면 glibc는 새 스레드에 필요한 속성(예: 스택, 우선순위 등)을 설정합니다.
    이러한 정보는 스레드 생성 시에 전달될 매개변수들로 구성됩니다.

2. 커널 모드로 전환 및 시스템 호출:

    설정이 완료되면, glibc는 `clone()` 시스템 호출을 통해 커널에 새로운 실행 흐름을 요청합니다.
    `clone()` 호출은 *기존 프로세스의 주소 공간을 공유하는 새로운 스레드를 생성*하는 핵심 시스템 호출입니다.

3. 커널에서 스레드 관리:

    커널은 `clone()` 호출을 처리하여 새 스레드를 생성합니다.
    이 새 스레드는 기존 프로세스의 메모리 공간을 공유하며, 스케줄러에 의해 다른 스레드와 함께 관리되고 실행됩니다.
    스레드 간의 동기화, 스케줄링, 종료 등은 커널이 관리하며, 각 스레드는 독립적으로 실행되면서도 프로세스 자원을 공유합니다.

4. 스레드 실행 및 종료:

    새로 생성된 스레드는 커널의 관리 하에 실행됩니다.
    스레드가 종료되면, 스레드의 리소스는 커널에 의해 정리됩니다.
    이는 프로세스와 비슷하게 커널 수준에서 처리되며, `pthread_join`과 같은 함수로 스레드의 종료를 기다릴 수 있습니다.

리눅스에서 프로세스는 주로 `fork()` syscall을 통해 생성됩니다.
`fork()`는 호출한 프로세스(부모 프로세스)를 복사하여 새로운 프로세스(자식 프로세스)를 만듭니다.
새로 생성된 자식 프로세스는 부모와 동일한 주소 공간을 복사하지만, *각자 독립적으로 실행*됩니다.
또한, `exec()` 계열 함수들은 새로 생성된 프로세스(자식 프로세스)가 다른 프로그램을 실행하도록 할 수 있습니다.

1. `fork()` 호출:

    부모 프로세스가 `fork()`를 호출하면, 커널은 부모 프로세스의 거의 모든 것을 복사하여 자식 프로세스를 생성합니다.
    이때 부모와 자식은 같은 프로그램 코드를 실행하지만, 각자 독립적인 프로세스로 관리됩니다.

2. 메모리 및 자원 공유:

    자식 프로세스는 부모와 독립된 주소 공간을 가지며, 프로세스 간 데이터는 공유되지 않습니다.
    그러나 파일 디스크립터와 같은 일부 자원은 부모와 자식이 공유할 수 있습니다.

3. `exec()`를 통한 프로그램 실행:

    자식 프로세스는 기본적으로 부모와 같은 프로그램을 실행하지만, `exec()` 함수를 사용하면 자식 프로세스가 다른 프로그램을 실행할 수 있습니다.
    이를 통해 자식 프로세스는 완전히 새로운 프로그램으로 전환됩니다.

4. 스케줄링:

    자식 프로세스는 부모와 동일한 우선순위로 스케줄링됩니다.
    그러나 자식 프로세스는 별도의 실행 흐름을 가지고 있습니다.

### 멀티 프로세스와 멀티 스레딩

- 멀티 프로세스

    하나의 애플리케이션이 여러 개의 프로세스로 나누어져 병렬로 작업을 수행하는 방식입니다.
    각 프로세스는 독립적인 주소 공간을 가지고 있어, 메모리를 공유하지 않습니다.

    프로세스 간 통신(IPC)은 파이프, 메시지 큐, 공유 메모리, 소켓 등을 사용해야 합니다.

    프로세스 간의 문맥 전환(Context Switch)이 스레드보다 더 무겁습니다.
    각 프로세스는 독립된 메모리 공간을 가지고 있어, 메모리 복사가 필요할 수 있습니다.

    - 장점
        - 안정성: 한 프로세스가 실패해도 다른 프로세스에 영향을 주지 않습니다.
        - 보안성: 독립된 메모리 공간을 가지기 때문에 프로세스 간의 데이터 접근이 제한됩니다.

    - 단점
        - 메모리 사용량: 프로세스마다 메모리를 할당해야 하므로 메모리 사용량이 더 많습니다.
        - 통신의 복잡성: 프로세스 간 통신은 스레드 간 통신보다 복잡합니다.

- 멀티 스레딩

    하나의 프로세스 내에서 여러 스레드를 생성하여 병렬로 작업을 수행하는 방식입니다.
    스레드는 같은 프로세스 내에서 주소 공간을 공유합니다.
    즉, 같은 데이터를 쉽게 공유하고 접근할 수 있습니다.

    스레드는 동일한 프로세스 내에서 메모리를 공유하므로 별도의 통신 메커니즘이 필요 없습니다.

    스레드 간 문맥 전환은 비교적 가볍고, 스레드는 같은 메모리 공간을 공유하기 때문에 프로세스 간 전환보다 효율적입니다.

    - 장점
        - 효율성: 메모리와 자원을 공유하기 때문에 프로세스보다 자원 사용이 적고, 문맥 전환도 더 빠릅니다.
        - 데이터 공유: 같은 프로세스 내에서 데이터를 쉽게 공유할 수 있어 IPC가 필요하지 않습니다.

    - 단점
        - 안정성: 하나의 스레드에서 발생한 오류가 다른 스레드나 전체 프로세스에 영향을 미칠 수 있습니다.
        - 동기화 문제: 스레드가 같은 메모리 공간을 공유하기 때문에 동기화 문제가 발생할 수 있습니다. 동기화 처리를 제대로 하지 않으면 데이터 경합이나 교착 상태(데드락) 같은 문제가 발생할 수 있습니다.

## 각 언어별 스레드

- [Java](./examples/java/src/main/ThreadExample.java)
    - 커널 수준 스레드(또는 네이티브 스레드)
    - Virtual Thread(사용자 수준 스레드)

- [Kotlin](./examples/kotlin/src/main/kotlin/Main.kt)
    - 커널 수준 스레드(또는 네이티브 스레드)
    - Coroutine

        경량 스레드로, Kotlin 런타임에 의해 관리되며 주로 이벤트 루프 기반으로 동작합니다.
        일반적으로 스레드 풀 위에서 실행됩니다.
        비동기 프로그래밍을 위한 추상화로, 실제 OS 스레드를 직접 사용하지 않습니다.

- [Rust](./examples/rust/src/main.rs)

    커널 수준 스레드(또는 네이티브 스레드) 사용합니다.

- Go

    M:N 하이브리드 멀티스레딩 모델입니다.
    경량 스레드로, Go 런타임에 의해 관리됩니다.

### Java Virtual Thread, Kotlin coroutine, and Go goroutine

Virtual Threads, Coroutines, Goroutines는 모두 효율적인 동시성(concurrency)을 지원하기 위해 설계된 경량 스레드 시스템입니다.
운영체제의 커널 스레드보다 훨씬 가벼운 경량 스레드를 제공하여 수많은 비동기 작업을 처리할 수 있습니다.

> - 동시성(concurrency):
>
>   여러 작업이 *논리적으로 동시에 실행*될 수 있는 개념입니다.
>   실제로 작업들이 정확히 동시에 실행되지 않더라도, 프로세스들이 번갈아가며 실행되면서 동시성이 구현될 수 있습니다.
>   ex: 커널 수준의 스레드, Java의 가상 스레드, Kotlin의 코루틴, Go의 고루틴 등
>
> - 비동기(asynchronous) 프로그래밍:
>
>   작업이 완료되기를 기다리지 않고, 완료되면 특정 시점에 처리를 이어가는 방식입니다.
>   동시성과 관련이 깊지만, 비동기 프로그래밍은 작업의 완료를 기다리지 않으면서 실행을 진행한다는 점에서 특정한 목적을 가집니다.
>   코루틴이나 고루틴, 가상 스레드는 비동기 작업을 쉽게 처리할 수 있게 해줍니다.

모두 높은 수준의 동시성을 지원하며, 블로킹 작업이 전체 시스템 성능에 미치는 영향을 최소화합니다.
세 시스템 모두 블로킹 I/O를 비동기식으로 처리하여 효율성을 높입니다.

#### Java 21 Virtual Threads

Java 21에서 도입된 Virtual Threads는 JVM 레벨에서 구현된 경량 스레드로, OS 스레드와 1:1 매핑되지 않습니다.
이는 JVM 레벨에서 구현된 경량 스레드로, OS 스레드와 1:1 매핑되지 않습니다.

기존 Java의 `Thread`와 인터페이스는 동일하지만, 가벼우며, OS 스레드 자원을 많이 소모하지 않고도 많은 스레드를 실행할 수 있습니다.

Virtual Thread는 Carrier Thread라고 불리는 커널 수준의 스레드(또는 Platform Thread) 위에서 실행됩니다.
블로킹 연산 시 Virtual Thread는 Carrier Thread에서 분리되어 다른 Virtual Thread가 실행될 수 있게 합니다.

```java
// Virtual Thread 생성 예시
Runnable task = () -> {
    Thread.sleep(1000); // 블로킹 I/O 작업 시뮬레이션
    System.out.println("Do I/O things");
};
Thread vThread = Thread.ofVirtual().start(task);
vThread.join();
```

1. 실행 시작: Virtual Thread가 Carrier Thread에 "마운트"되어 실행됩니다.
2. 블로킹 연산 발생: I/O 작업 등의 블로킹 연산이 발생합니다.
3. 분리: Virtual Thread는 Carrier Thread에서 "언마운트"됩니다.
4. 상태 저장: Virtual Thread의 실행 상태(스택 등)가 힙 메모리에 저장됩니다.
5. 다른 작업 실행: Carrier Thread는 다른 Virtual Thread를 실행할 수 있게 됩니다.
6. 블로킹 해제: 블로킹 연산이 완료되면, Virtual Thread는 다시 Carrier Thread에 "마운트"되어 실행을 재개합니다.

자바의 가상 쓰레드는 다음과 같은 특징이 있습니다.

- Thread와 동일한 API:

    Virtual Threads는 기존의 Java `Thread` 클래스와 동일한 API를 사용합니다.
    코드를 비동기적으로 작성하는데 별도의 API를 배우지 않아도 됩니다.

    ```java
    public static void main(String[] args) throws InterruptedException {
        Thread virtualThread = Thread.ofVirtual().start(() -> {
            System.out.println("Hello from a Virtual Thread!");
        });
        virtualThread.join();
    }
    ```

- 커널 스레드와 독립적:

    Virtual Threads는 운영체제의 커널 스레드에 직접 매핑되지 않습니다.
    따라서 수천 개의 Virtual Threads를 생성해도 성능 문제가 거의 없습니다.

- Blocking과 Non-blocking:

    Virtual Threads는 블로킹 I/O와도 잘 동작합니다.
    I/O가 블로킹되면 Virtual Thread는 커널 스레드를 차지하지 않고 다시 스케줄링됩니다.

- 간단한 동기식 프로그래밍 모델:

    동시성 문제를 해결하기 위해 기존에 사용했던 복잡한 비동기 콜백이나 Future, CompletableFuture 대신, 동기식 프로그래밍 모델을 유지하면서도 가볍고 많은 수의 스레드를 생성할 수 있습니다.

#### Kotlin Coroutines

Kotlin의 코루틴은 비동기, 논블로킹 코드를 마치 동기 코드처럼 작성할 수 있게 해주는 동시성 모델입니다.
코루틴은 경량의 스레드와 유사한 작업 단위입니다.
코루틴은 함수 실행을 중단하고 재개할 수 있는 능력이 있으며, 이를 통해 비동기 작업을 처리합니다.
Kotlin Coroutines는 suspend 함수와 코루틴 빌더를 사용해 명시적으로 비동기 코드를 작성하게 됩니다.

```kotlin
import kotlinx.coroutines.*

// Coroutine 사용 예시
suspend fun fetchData(): String {
    delay(1000)     // 비동기 대기(일시 중단)합니다.
                    // 코루틴의 상태가 저장되고, 실행 중이던 스레드는 해제됩니다.
    return "데이터"   // 1초 후, 코루틴이 재개되어 "데이터"를 반환합니다.
                    // 이 과정에서 스레드는 블록되지 않고, 다른 작업을 수행할 수 있습니다.
}

fun main() = runBlocking {
    val result = fetchData()
    println(result)

    // 코루틴 실행
    launch {
        delay(1000L) // 논블로킹 딜레이
        println("코루틴 종료!")
    }
    println("메인 스레드 종료!")
}
```

1. `suspend` 함수 호출 시 일시 중단됩니다.
2. 코루틴의 현재 실행 상태(로컬 변수, 실행 지점 등)가 객체로 저장됩니다.
3. 코루틴이 실행 중이던 스레드는 다른 작업을 수행할 수 있게 됩니다.
4. 재개: 일시 중단된 작업이 완료되면, 저장된 상태를 기반으로 실행이 재개됩니다.

Kotlin의 coroutine은 다음과 같은 특징을 갖습니다:
- 콜백 대신 일시 중단:

    전통적인 콜백 대신 일시 중단 메커니즘을 사용합니다.

- 명시적 비동기:

    `suspend` 키워드를 통해 함수가 비동기로 동작함을 명시적으로 나타냅니다.
    코루틴은 일시 중단 지점에서 실행을 멈추고 나중에 재개할 수 있습니다.
    이 중단은 스레드를 블로킹하지 않으므로 자원을 아낄 수 있습니다.

- Coroutine Scope:

    코루틴은 `CoroutineScope` 내에서 실행되며, 이 스코프는 코루틴의 생명 주기를 관리합니다.

- Structured Concurrency:

    코루틴은 부모-자식 관계로 실행되며, 자식 코루틴의 생명 주기가 부모 코루틴과 연결되어 있습니다.
    이를 통해 예외 처리와 자원 관리를 쉽게 할 수 있습니다.
    즉, 모든 코루틴은 특정 스코프 내에서 실행되며, 부모 스코프가 종료되면 자식 코루틴도 정리됩니다.

- Non-blocking:

    코루틴은 블로킹되지 않고, 비동기적으로 작업이 완료되면 다시 재개됩니다.

- Dispatchers:

    코루틴은 다양한 `Dispatcher`를 통해 특정 스레드 풀 또는 스레드에서 실행될 수 있습니다 (예: `Dispatchers.IO`, `Dispatchers.Default`).

#### Go Goroutines

Go의 Goroutines는 Go에서 제공하는 경량 스레드입니다.
Goroutines는 매우 효율적으로 동시성 작업을 처리하기 위해 설계되었으며, 수백만 개의 Goroutines를 생성하더라도 최소한의 메모리만 소비합니다.
가령 [goroutine은 기본적으로 2kb의 적은 공간을 차지하는 반면, java thread 같은 경우 만드는 경우에 256kb~2mb 사이즈](https://baeji77.github.io/dev/golang-with-gorotine(2)/)로 쓰레드를 만들 수 있습니다.
Go는 Goroutines와 채널을 사용해 동시성을 처리합니다.

```go
package main
import (
    "fmt"
    "time"
)

func main() {
    go func() {
        time.Sleep(1 * time.Second)
        fmt.Println("고루틴 종료!")
    }()
    fmt.Println("메인 스레드 종료!")
    time.Sleep(2 * time.Second)
}
```

Golang의 goroutine은 다음과 같은 특징을 갖습니다:
- 비동기 작업의 기본 단위:

    Goroutines는 Go에서 비동기 작업을 실행하는 기본 단위입니다.
    Go의 표준 `go` 키워드를 통해 쉽게 실행할 수 있습니다.

- 채널 기반 통신:

    Goroutines 간의 통신은 채널을 통해 이루어지며, 이는 안전한 동시성을 제공하는데 중요한 역할을 합니다.

    ```go
    // Goroutine 사용 예시
    func fetchData() string {
        time.Sleep(time.Second) // 시뮬레이션된 작업
        return "데이터"
    }

    func main() {
        ch := make(chan string)
        go func() {
            result := fetchData()
            ch <- result
        }()
        fmt.Println(<-ch)
    }
    ```

- 스택 공간 자동 관리:

    Goroutines는 필요한 만큼의 스택 메모리만을 동적으로 할당하고, 필요에 따라 자동으로 조절됩니다.

- OS 스레드와 분리된 경량 스레드:

    Goroutines는 OS 스레드와는 독립적으로 운영되며, Go 런타임이 이를 효율적으로 관리합니다.
    Go 런타임이 OS 스레드에 고루틴을 다중화하여 실행할 수 있어 많은 고루틴을 동시에 처리할 수 있습니다.

- 간결한 API:

    Goroutines는 매우 간단하게 사용할 수 있으며, 비동기적 코드와 동기적 코드를 자연스럽게 통합할 수 있습니다.

#### 주요 차이점

- 동시성 모델:
    - Java 가상 스레드:

        전통적인 동기 스레드 모델을 사용하되, 스레드 풀에 의존하지 않습니다.
        대신 Virtual Threads를 대규모로 사용하여 동시성을 처리합니다.
        각 가상 스레드는 JVM이 관리하는 실제 스레드에 매핑되며, 유저 모드에서 효율적인 스케줄링을 제공합니다.

    - Kotlin 코루틴:

        비동기 작업을 `suspend` 키워드를 통해 명시적으로 표시하고, structured concurrency를 강조하여 스코프를 통해 코루틴을 관리합니다.
        코루틴은 비동기 실행을 중단하고 재개하는 협력적 멀티태스킹 모델을 사용합니다.

    - Go 고루틴

        고루틴은 Go 런타임이 관리하며, 다수의 고루틴이 소수의 OS 스레드에 매핑됩니다.
        Go에서는 `go` 키워드와 채널을 사용한 CSP(Communicating Sequential Processes) 모델을 사용하여 간결하게 동시성을 처리합니다.

- 블로킹과 논블로킹 작업:

    - Java 가상 스레드:

        가상 스레드는 블로킹 작업을 효율적으로 처리하며, 커널 레벨에서 블로킹되지 않도록 관리됩니다.

    - Kotlin 코루틴:

        코루틴은 기본적으로 논블로킹 방식으로 설계되었으며, 지연 함수를 통해 비동기 작업을 처리합니다.

    - Go 고루틴:

        고루틴은 Go 런타임의 스케줄러를 통해 블로킹 작업을 처리하며, 채널을 통해 동기화됩니다.

- 메모리 관리 및 성능

    - Java Virtual Threads:

        운영체제 스레드보다 메모리를 덜 사용하지만, Virtual Thread가 커널 스레드에 의해 스케줄링됩니다.
        따라서 성능은 JVM의 런타임에 달려 있습니다.

    - Kotlin Coroutines:

        코루틴은 중단점에서 제어권을 반환하고 다시 재개될 수 있기 때문에 메모리와 CPU 사용량을 효율적으로 관리합니다.

    - Go Goroutines:

        매우 적은 메모리에서 실행되며, 런타임이 Goroutines와 OS 스레드를 매핑해 최적화합니다.
        Goroutines는 스택 공간을 필요에 따라 동적으로 할당합니다.

## 멀티태스킹 운영체제와 GC `ParallelGCThreads` 설정

맥북과 같은 멀티태스킹 운영체제에서 실행되는 수천 개의 스레드와 `ParallelGCThreads` 설정 간의 관계

### 컨텍스트 스위칭 비용

운영체제는 여러 프로세스와 스레드 간에 CPU 시간을 할당한다.
스레드 수가 코어 수를 초과하면, 운영체제는 컨텍스트 스위칭(Context Switching)을 자주 수행해야 한다.
컨텍스트 스위칭은 CPU가 한 작업에서 다른 작업으로 전환하는 과정에서 발생하는 오버헤드.
이 오버헤드가 많아지면 시스템 성능이 저하될 수 있다

### 자원 경합

하드웨어 코어 수보다 많은 스레드가 동시에 실행되면 자원 경합(Resource Contention)이 발생할 수 있다.
이는 특히 CPU 및 메모리 자원에 대한 경합이 심할 수 있으며, 이로 인해 시스템의 전반적인 성능이 저하될 수 있다.

### 가비지 컬렉션 최적화

`ParallelGCThreads`는 가비지 컬렉션 작업에 사용될 병렬 스레드의 수를 지정한다.
이 값이 너무 높으면 GC 작업 동안 시스템의 다른 프로세스나 애플리케이션에 필요한 자원을 과도하게 점유할 수 있으며, GC의 효율성이 저하될 수 있다.

### 스레드와 코어의 관계

하드웨어 코어 하나당 동시에 하나의 스레드만 실행할 수 있다.
따라서 코어 수를 초과하는 스레드는 대기 상태에 있거나 시간을 분할하여 실행된다.
`ParallelGCThreads`를 너무 높게 설정하면 이론적으로는 더 빠른 처리가 가능해 보일 수 있지만,
실제로는 컨텍스트 스위칭과 자원 경합으로 인해 GC 작업의 효율성이 저하될 수 있다.
