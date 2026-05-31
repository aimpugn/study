# 01e. Concurrency, Isolation, Observability

운영 장애를 제대로 설명하려면 "느리다"는 증상을 lock, futex, wait queue, cgroup limit, namespace, `/proc` 값, perf/eBPF trace 같은 관측 가능한 상태로 바꿔야 합니다. OS는 process를 격리하고 자원을 나눠 주지만, 동시에 같은 머신 안의 모든 process가 CPU, memory, disk, network를 경쟁하게 만듭니다. 커널 동시성과 관측 도구를 모르면 Kafka/Cassandra/Spark의 병목을 제품 metric만으로 오판합니다.

이 문서에서는 lock과 memory ordering, futex, container resource isolation, 관측 표면을 하나의 질문으로 묶습니다. "어느 공유 상태를 누가 기다리고 있으며, 그 증거는 어디서 볼 수 있는가?"

## race는 두 실행 흐름이 같은 상태를 다른 순서로 본다는 문제다

두 thread가 같은 counter를 읽고 각각 1을 더한다고 합시다. 둘 다 `10`을 읽고 둘 다 `11`을 쓰면 update 하나가 사라집니다. 이것이 lost update이고, race의 가장 작은 예입니다.

```
without lock
  T1 read count=10
  T2 read count=10
  T1 write count=11
  T2 write count=11
  result: 11, but expected 12
```

lock은 critical section을 한 번에 하나의 실행 흐름만 지나가게 만듭니다. 하지만 lock은 공짜가 아닙니다. lock을 잡은 thread가 오래 걸리면 다른 thread는 기다립니다. lock이 너무 굵으면 parallelism이 줄고, 너무 잘게 나누면 ordering과 deadlock 문제가 어려워집니다.

```
with lock
  T1 acquire
  T1 read/write count=11
  T1 release
  T2 acquire
  T2 read/write count=12
  T2 release
```

Kafka의 producer accumulator, broker request queue, Cassandra memtable/update path, Spark scheduler metadata는 모두 공유 상태를 가집니다. 제품 내부 구현은 더 복잡하지만, 질문은 같습니다. 무엇을 동시에 바꾸면 안 되며, 그 보호 때문에 누가 기다리는가?

## futex는 user-space lock과 kernel sleep 사이의 다리다

고성능 lock은 매번 kernel로 들어가면 너무 비쌉니다. uncontended lock은 user space atomic instruction으로 빠르게 잡고 풀 수 있어야 합니다. 하지만 이미 다른 thread가 lock을 잡고 오래 기다려야 한다면 CPU를 계속 태우기보다 kernel에게 "이 주소의 값이 바뀔 때까지 재워 달라"고 부탁하는 편이 낫습니다. Linux futex(fast userspace mutex)는 이 다리를 제공합니다.

```
lock attempt
  -> atomic compare-and-swap succeeds
       stay in user space
  -> atomic compare-and-swap fails
       futex wait on memory word
       kernel puts thread to sleep
  -> unlock path changes memory word
       futex wake
       waiting thread becomes runnable
```

Java monitor, `LockSupport.park()`, pthread mutex 같은 higher-level primitive는 runtime과 구현에 따라 futex 계열 wait/wake를 사용할 수 있습니다. 그래서 JVM thread dump의 `WAITING`이나 `BLOCKED`는 결국 OS scheduler와 futex wait queue, CPU runnable 상태로 이어질 수 있습니다.

## lost wakeup은 조건 확인과 잠들기를 따로 하면 생긴다

condition variable을 쓸 때 "조건이 아직 false면 sleep"이라는 흐름이 필요합니다. 그런데 조건 확인과 sleep 등록 사이가 원자적이지 않으면 signal이 그 틈에 지나가 버릴 수 있습니다.

```
bad sequence
  T1 checks queue empty
  T2 enqueues item
  T2 sends signal
  T1 goes to sleep after signal already passed
  T1 may sleep forever
```

그래서 condition variable은 mutex와 함께 사용하고, 보통 loop로 조건을 다시 확인합니다. signal은 "조건이 참이다"라는 영구 상태가 아니라 "다시 확인해 보라"는 알림에 가깝기 때문입니다.

```
while queue is empty:
  wait(cond, mutex)
dequeue item
```

이 구조는 blocking queue, executor work queue, network event loop, Spark task scheduling에서 모두 중요합니다. 한 번 깨어났다고 원하는 일이 반드시 가능하다고 단정하면 race가 생깁니다.

## memory ordering은 "썼다"와 "다른 core가 그 순서로 보았다"를 분리한다

여러 CPU core는 cache와 store buffer를 갖고 있고, compiler와 CPU는 성능을 위해 일부 memory operation의 순서를 바꿀 수 있습니다. 한 thread가 `data`를 쓰고 `ready=true`를 썼다고 해서, 다른 core가 반드시 그 순서로 관측한다고 단정하면 안 됩니다. lock, atomic, volatile, memory barrier는 이런 관측 순서를 제한합니다.

```
producer thread:
  data = 42
  ready = true

consumer thread:
  if ready:
     read data

without ordering rule:
  consumer may observe ready before data is safely visible
```

Java의 `volatile`, `synchronized`, `java.util.concurrent` primitive는 Java Memory Model이 정의한 happens-before 관계를 만듭니다. 이 관계는 JVM 수준의 약속이지만, 결국 CPU memory ordering과 OS scheduling 위에서 구현됩니다. Kafka/Cassandra/Spark가 JVM 위에서 동작한다는 말은 application-level consistency와 runtime-level memory visibility를 모두 고려해야 한다는 뜻입니다.

## cgroup과 namespace는 container를 process처럼 보이게 만드는 장치다

container는 가벼운 VM이 아닙니다. 보통 Linux namespace와 cgroup을 조합해 process에게 독립된 hostname, mount, network, PID view를 보여 주고, CPU/memory/I/O 같은 자원 사용량을 제한합니다. namespace는 "무엇이 보이는가"를 나누고, cgroup은 "얼마나 쓸 수 있는가"를 제한하고 계측합니다.

```
containerized process
  namespace:
    sees its own pid tree, mounts, network view
  cgroup:
    limited CPU weight/quota
    limited memory
    limited I/O depending on setup
```

이 구분은 장애 분석에서 중요합니다. container 안에서 PID 1로 보이는 process가 host에서는 다른 PID일 수 있습니다. container 안의 `/proc` 값이 namespace/cgroup view에 맞춰 보일 수 있고, host 전체 자원과 container limit이 다를 수 있습니다. memory limit을 넘으면 host 전체 memory가 남아 있어도 cgroup OOM이 날 수 있습니다. CPU quota가 낮으면 process는 runnable인데 quota가 소진되어 다음 period까지 기다릴 수 있습니다.

Kafka broker를 container에서 돌릴 때 page cache와 cgroup memory accounting을 잘못 이해하면 heap을 안전하게 잡았다고 생각해도 OOM을 맞을 수 있습니다. Cassandra는 disk and memory heavy workload라 cgroup I/O와 memory limit이 compaction/flush에 직접 영향을 줍니다. Spark executor container는 heap, overhead memory, off-heap, Python worker, shuffle buffer를 함께 제한받습니다.

## observability는 tool 이름이 아니라 상태 질문이다

도구를 외우는 순서는 약합니다. 먼저 무엇을 보려는지 정해야 합니다.

| 질문 | Linux에서 볼 표면 | 볼 때 조심할 점 |
|---|---|---|
| 어떤 syscall에서 시간을 쓰는가? | `strace`, `perf trace`, eBPF tracepoint | tracing overhead와 권한 |
| CPU에서 실행 중인가, 기다리는가? | `top`, `pidstat`, `perf sched`, off-CPU profiling | Java `RUNNABLE`과 OS running 구분 |
| memory pressure가 어디서 오나? | `/proc/meminfo`, `/proc/<pid>/smaps`, cgroup memory files | page cache와 anonymous memory 구분 |
| socket queue가 차는가? | `ss -tin`, `/proc/net/*` | loopback과 실제 network 차이 |
| disk I/O가 밀리는가? | `iostat -xz`, `pidstat -d`, block trace | device, filesystem, page cache 구분 |
| kernel path를 동적으로 보고 싶은가? | eBPF/bpftrace, `perf` | kernel version, symbol, production safety |

`/proc`은 kernel이 process와 system 상태를 file처럼 보여 주는 pseudo filesystem입니다. 실제 disk file이 아니라, 읽을 때 kernel이 현재 상태를 만들어 줍니다. `/sys`는 device, driver, kernel object와 설정을 더 구조적으로 노출합니다. 이 둘은 Linux 중심 도구입니다. macOS에서는 같은 경로가 없으므로 `dtrace`, `dtruss`, Activity Monitor, Instruments, `netstat`, `fs_usage` 같은 다른 도구를 사용해야 합니다.

## perf와 eBPF는 "왜 느린지"를 함수와 event로 좁힌다

`perf`는 CPU profiling, scheduler event, tracepoint, hardware counter 같은 표면을 제공합니다. flame graph는 stack sample을 모아 어떤 함수 경로가 CPU 시간을 쓰는지 보여 줍니다. 하지만 CPU flame graph만으로 off-CPU wait는 보이지 않습니다. lock, I/O, sleep 대기는 off-CPU profiling이 필요합니다.

eBPF는 제한된 프로그램을 kernel hook에 붙여 runtime 관측을 할 수 있는 mechanism입니다. system call, kprobe, tracepoint, network event 등에 붙어 "어떤 pid가 어떤 fd에 얼마나 오래 write했나", "어떤 TCP retransmission이 늘었나" 같은 질문을 더 세밀하게 볼 수 있습니다. 다만 production에서 eBPF를 쓸 때는 권한, overhead, kernel version, 개인 정보/민감 데이터 노출을 조심해야 합니다.

## 현실 시나리오: CPU는 낮고 Kafka consumer lag는 증가한다

이때 바로 consumer 수를 늘리면 안 됩니다. 아래처럼 상태를 좁혀 갑니다.

```
lag grows
  -> committed offset moves slowly
  -> consumer thread timeline?
       CPU running?
       waiting on DB write?
       waiting on socket read?
       blocked on lock?
       paused by GC?
  -> downstream queue?
  -> broker fetch latency?
  -> partition skew?
```

관측은 여러 층을 같이 봅니다. JVM thread dump로 thread 상태를 보고, application metric으로 processing time과 downstream latency를 보고, `ss`로 socket queue를 보고, `iostat`으로 disk wait를 보고, GC log로 pause를 봅니다. 하나의 metric이 답을 주는 것이 아니라, 요청이 어디서 시간을 보냈는지 trace를 좁힙니다.

## 현실 시나리오: container memory limit 안에서 Spark executor가 죽는다

executor heap을 `4g`로 잡고 container memory limit을 `4g`에 가깝게 주면 위험합니다. JVM heap 밖에도 thread stack, direct buffer, native memory, Python worker, shuffle buffer, page cache가 있습니다. OS는 cgroup limit 안에서 이 전체를 봅니다.

```
container memory limit
  -> JVM heap
  -> executor overhead
  -> off-heap/direct buffer
  -> thread stacks
  -> Python worker/native memory
  -> page cache / mmap
  -> cgroup OOM if limit exceeded
```

이때 해결은 heap만 줄이거나 늘리는 문제가 아닙니다. executor memory overhead, off-heap 사용, shuffle spill, partition 수, local disk, Kubernetes/YARN memory accounting을 함께 봐야 합니다.

## 문서를 덮고 확인할 것

- lock이 correctness를 높이면서 latency를 키우는 경로를 race 예제로 설명해 보세요.
- futex가 user-space atomic lock과 kernel sleep을 어떻게 연결하는지 말해 보세요.
- cgroup과 namespace를 "보이는 세계"와 "쓸 수 있는 자원"으로 구분해 보세요.
- `/proc`, `ss`, `iostat`, `perf`, eBPF가 각각 어떤 상태 질문에 답하는지 말해 보세요.
- Kafka/Cassandra/Spark metric 하나를 골라 OS 관측 표면 세 개와 연결해 보세요.

## Deadlock, livelock, starvation은 같은 정지가 아니다

동시성 문제를 모두 "락 문제"라고 부르면 해결 방향이 흐려집니다. Deadlock은 서로가 가진 자원을 기다리며 아무도 आगे가지 못하는 상태입니다. 전형적인 조건은 mutual exclusion, hold and wait, no preemption, circular wait입니다. Thread A가 lock X를 잡고 Y를 기다리고, thread B가 Y를 잡고 X를 기다리면 두 thread 모두 진행하지 못합니다.

```
Thread A
  lock X
  wait for Y

Thread B
  lock Y
  wait for X

result
  both sleep or spin forever unless timeout/recovery exists
```

Livelock은 조금 다릅니다. 참여자들이 계속 움직이고 상태를 바꾸지만 실질적인 진전이 없습니다. 두 worker가 충돌을 피하려고 계속 양보만 하거나, 여러 client가 timeout 뒤 즉시 retry하면서 서버를 더 과부하시키는 상황을 떠올리면 됩니다. Starvation은 어떤 작업이 계속 자원을 받지 못하는 상태입니다. CPU priority, unfair lock, connection pool, partition queue, disk queue 어디서든 생길 수 있습니다.

해결책도 다릅니다. Deadlock은 lock ordering, timeout, try-lock 후 rollback, 자원 계층화로 줄입니다. Livelock은 backoff, jitter, admission control, coordination을 넣어 "움직임"이 아니라 "진전"을 만들게 해야 합니다. Starvation은 fairness, aging, queue 분리, priority inheritance, quota를 봅니다. Kafka consumer group rebalance가 너무 자주 일어나 계속 처리 시간이 줄어드는 상황, Cassandra repair/compaction이 foreground request를 밀어내는 상황, Spark speculative execution과 retry가 cluster를 더 바쁘게 만드는 상황은 각각 다른 진전 문제를 갖습니다.

## Futex와 park/unpark는 user-space와 kernel wait를 잇는다

Futex는 fast userspace mutex의 줄임말로, uncontended path는 user-space atomic operation으로 빠르게 처리하고, 실제로 기다려야 할 때만 kernel wait queue로 내려가는 원리입니다. Lock이 비어 있으면 user-space에서 owner를 바꾸고 끝납니다. 이미 누가 lock을 갖고 있으면 thread는 futex wait로 kernel에 들어가 잠들고, unlock하는 thread가 futex wake로 깨웁니다.

```
lock fast path
  atomic compare-and-swap succeeds
  -> no syscall

lock contended path
  atomic compare-and-swap fails
  -> futex wait
  -> scheduler sleeps thread

unlock
  atomic state change
  -> futex wake if waiters exist
```

JVM의 monitor, `LockSupport.park/unpark`, pthread mutex, condition variable은 구현과 정책은 다르지만 같은 큰 질문을 다룹니다. 언제 spin하고, 언제 잠들며, 누가 깨우는가입니다. Spin은 짧은 대기에 유리할 수 있지만 CPU를 태웁니다. Sleep은 CPU를 양보하지만 syscall과 context switch, wakeup latency가 붙습니다. 따라서 lock 설계는 correctness뿐 아니라 contention duration, CPU pressure, scheduler latency를 함께 봅니다.

Kafka request path에서 shared queue lock이 길게 잡히면 network thread나 handler thread가 futex wait로 내려갈 수 있습니다. Cassandra memtable이나 schema metadata lock, Spark scheduler/driver lock도 같은 방식으로 thread 상태를 바꿉니다. Thread dump에는 Java 상태만 보이고, kernel에는 futex wait가 보일 수 있습니다. `perf`, eBPF, `strace -f -e futex`, async-profiler lock/off-CPU mode가 필요한 이유입니다.

## Memory ordering과 cache coherence는 lock 밖의 정확성을 다룬다

CPU core마다 cache와 store buffer가 있고, compiler와 CPU는 성능을 위해 일부 memory operation을 재배치하거나 지연해 보이게 만들 수 있습니다. Cache coherence는 같은 memory location에 대한 일관된 관측을 유지하려는 hardware protocol입니다. 하지만 coherence가 곧 모든 변수의 순서 보장을 뜻하지는 않습니다. "A를 쓰고 B를 썼다"가 다른 core에서 항상 같은 순서로 보인다고 가정하려면 memory ordering 규칙이 필요합니다.

```
Thread 1
  data = 42
  ready = true

Thread 2
  if ready:
    read data

without ordering rule
  Thread 2 may observe ready but not the intended data order on weak models
```

Java에서는 `volatile`, `synchronized`, atomic classes, happens-before 규칙이 이 문제를 다룹니다. C/C++에서는 memory order를 명시할 수 있습니다. Kernel code에서는 memory barrier와 atomic primitive를 사용합니다. OS 문서에서 memory ordering을 다루는 이유는 JVM이나 database가 높은 수준의 lock을 쓰더라도 결국 CPU와 kernel scheduler, futex, cache coherence 위에 있기 때문입니다.

False sharing은 correctness가 아니라 성능을 흔드는 예입니다. 두 thread가 서로 다른 counter를 갱신해도 같은 cache line에 있으면 line ownership이 계속 이동합니다. Lock contention이 없어도 throughput이 떨어질 수 있습니다. High-throughput metric counter, ring buffer head/tail, queue index, per-partition accumulator를 설계할 때 padding이나 striped design이 등장하는 이유입니다.

## IPC는 process isolation을 넘는 의도적인 통로다

Process isolation은 한 process가 다른 process의 memory를 마음대로 읽지 못하게 합니다. 하지만 실제 시스템은 process 사이 통신이 필요합니다. IPC(inter-process communication)는 이 격리를 유지하면서 data나 signal을 주고받는 통로입니다. Pipe, Unix domain socket, TCP socket, shared memory, signal, file lock, message queue가 모두 IPC에 속합니다. 각각 copy 비용, ordering, backpressure, security boundary, lifetime이 다릅니다.

```
pipe
  process A writes bytes
  -> kernel pipe buffer
  -> process B reads bytes

shared memory
  processes map same physical pages
  -> data copy reduced
  -> synchronization must be explicit

Unix domain socket
  local socket semantics
  -> can pass file descriptors
  -> kernel mediates permissions and buffers
```

Kafka/Cassandra/Spark는 주로 network socket을 통해 process와 node 사이를 통신하지만, local IPC도 중요할 수 있습니다. Spark executor와 Python worker 사이의 pipe/socket, sidecar와 application 사이의 Unix socket, metrics agent와 process 사이의 shared file or socket이 있습니다. IPC를 보면 "왜 process를 나눴는데도 서로 영향을 주는가"를 설명할 수 있습니다. Kernel buffer와 scheduler, backpressure, file descriptor limit, namespace permission이 통신의 실제 성격을 정합니다.

## Namespace, cgroup, capability, seccomp는 container의 보이는 세계와 가능한 행동을 나눈다

Container는 작은 VM이 아닙니다. 일반적으로 같은 kernel을 공유하면서 namespace와 cgroup, capability, seccomp, filesystem mount, network setup을 조합해 process의 세계를 제한합니다. Namespace는 process가 보는 이름 공간을 분리합니다. PID namespace는 process id의 보이는 범위를 바꾸고, mount namespace는 filesystem tree를 다르게 보이게 하며, network namespace는 interface와 routing table을 분리합니다. Cgroup은 CPU, memory, I/O 같은 자원 사용을 account하고 제한합니다.

Capability는 root 권한을 더 작은 권한 조각으로 나눕니다. 예를 들어 raw socket을 열거나 network setting을 바꾸거나 특정 mount 작업을 하려면 별도 capability가 필요할 수 있습니다. Seccomp는 process가 호출할 수 있는 system call을 제한하는 필터입니다. 컨테이너 안에서 `perf`, eBPF, `tcpdump`, `strace`가 실패하는 이유는 도구 문제가 아니라 capability, seccomp, namespace, host kernel policy 때문일 수 있습니다.

```
containerized process
  sees: its PID namespace, mount namespace, network namespace
  uses: cgroup CPU/memory/I/O budget
  may lack: CAP_SYS_ADMIN, CAP_NET_ADMIN, perf_event permission
  may be blocked: seccomp-denied syscalls
```

Kafka broker가 container 안에서 file descriptor limit이나 mmap count에 걸리고, Cassandra가 locked memory나 perf counter permission 문제를 만나고, Spark executor가 cgroup memory limit으로 kill되는 것은 모두 isolation과 resource control이 application behavior로 올라온 사례입니다. Container를 설명할 때 "격리한다"에서 멈추지 말고, "무엇이 다르게 보이고, 무엇이 제한되며, 어떤 system call이나 kernel feature가 막히는가"를 말해야 합니다.

Virtual machine은 다른 boundary를 만듭니다. Hypervisor는 guest OS에게 virtual CPU, memory, device를 제공하고, guest kernel은 자기 안의 process를 관리합니다. Container는 host kernel을 공유하지만 namespace/cgroup으로 process를 제한합니다. VM은 kernel boundary가 더 강하지만 overhead와 운영 단위가 다르고, container는 가볍지만 host kernel attack surface와 resource accounting을 더 조심해야 합니다. Kubernetes cluster에서 둘은 함께 쓰이기도 합니다.

## Observability는 증상을 kernel 질문으로 바꾸는 일이다

관측 도구를 많이 아는 것보다 중요한 것은 질문을 잘 세우는 것입니다. `/proc`은 process와 kernel state를 pseudo-file로 보여 줍니다. `/sys`는 device와 kernel object의 속성을 노출합니다. `strace`는 system call 경로를 봅니다. `perf`는 CPU sampling, hardware counter, scheduler/block/network event를 볼 수 있습니다. eBPF는 kernel/user event에 작은 program을 붙여 더 구체적인 telemetry를 만들 수 있습니다. `ss`는 socket state와 queue를 보여 주고, `iostat`은 block device I/O 지표를 보여 주며, `pidstat`은 process별 CPU, memory, I/O, context switch를 봅니다. `tcpdump`는 wire-level packet을 잡습니다.

```
symptom: p99 latency increased
  -> application histogram: which endpoint/partition/stage?
  -> thread state: on CPU, lock, I/O, GC?
  -> syscall trace: read/write/fsync/futex/connect?
  -> socket queue: receive/send backlog?
  -> block I/O: await, queue depth, utilization?
  -> scheduler: run queue, throttling, context switch?
  -> product metric: lag, compaction, shuffle, replica delay?
```

도구는 권한과 환경을 탑니다. macOS에는 Linux `/proc`이 없고, container 안에서는 `perf`와 eBPF 권한이 막힐 수 있습니다. Cloud managed service에서는 host-level trace를 볼 수 없을 수도 있습니다. 이런 경우 도구 실패를 개념 실패로 착각하면 안 됩니다. 같은 질문을 application metric, runtime log, cloud provider metric, sampling profiler, packet capture 가능한 지점으로 바꿔야 합니다.

## Flame graph와 tail latency를 함께 읽기

Flame graph는 sampling profile을 시각화해 어디서 CPU time이 쓰였는지 보여 줍니다. 넓은 frame은 많은 sample이 그 stack에 있었음을 뜻합니다. 하지만 CPU flame graph는 기다린 시간을 직접 보여 주지 않을 수 있습니다. Request가 lock이나 I/O에서 오래 기다렸다면 CPU flame graph에는 짧게 나타날 수 있습니다. Wall-clock profiler나 off-CPU flame graph를 함께 봐야 tail latency의 대기 시간을 읽을 수 있습니다.

```
CPU flame graph
  -> where running CPU time went

off-CPU or wall-clock profile
  -> where elapsed request time waited

latency histogram
  -> which percentile is affected
```

Kafka consumer lag가 늘었는데 CPU flame graph가 parse code를 크게 보여 준다면 parsing/serialization을 봅니다. CPU graph가 작고 off-CPU가 futex와 fsync를 보여 준다면 lock이나 disk sync를 봅니다. Cassandra read latency가 늘었는데 block I/O wait가 크면 SSTable read/cache/compaction을 봅니다. Spark stage가 느린데 CPU는 낮고 network wait가 크면 shuffle fetch, remote executor, skew를 봅니다.

## Interview replay: 동시성과 격리, 관측을 한 번에 묶기

이 장의 답변은 이렇게 시작할 수 있습니다. "동시성은 여러 실행 흐름이 공유 상태를 어떤 순서로 보고 바꾸는가의 문제입니다. Lock과 atomic은 correctness를 지키지만, contention이 생기면 futex wait와 scheduler sleep으로 latency를 만들 수 있습니다. Memory ordering과 cache coherence는 CPU core 사이에서 값과 순서가 어떻게 보이는지를 다룹니다. Process isolation은 namespace, cgroup, capability, seccomp, VM/container boundary와 만나 운영상의 권한과 자원 제한으로 드러납니다."

꼬리 질문이 관측으로 오면 "증상을 kernel 질문으로 바꿉니다"라고 말하면 됩니다. CPU 문제인지 보려면 on-CPU profile과 run queue를 봅니다. Lock 문제인지 보려면 thread dump, futex wait, monitor contention을 봅니다. Disk 문제인지 보려면 page cache, dirty writeback, block I/O를 봅니다. Network 문제인지 보려면 socket queue, retransmission, packet capture를 봅니다. Container 문제인지 보려면 cgroup event와 namespace/capability/seccomp boundary를 봅니다. 제품 문제인지 보려면 Kafka lag/ISR, Cassandra compaction/repair, Spark shuffle/stage metric을 OS 관측과 나란히 놓습니다.

문서를 덮고 다시 그려 볼 trace는 다음과 같습니다.

```
shared state or external resource
  -> threads coordinate with lock/atomic/futex
  -> scheduler decides who runs
  -> CPU cache and memory ordering affect visibility
  -> container boundary limits what process sees and uses
  -> observability tools sample syscall, CPU, I/O, socket, cgroup events
  -> product symptom becomes explainable as a lower-layer wait or contention
```

이 흐름을 말할 수 있으면 "동시성"을 Java keyword 암기로만 보지 않고, OS scheduler와 CPU memory model, container isolation, kernel observability까지 연결할 수 있습니다.

## 보안과 보호는 성능과 별개가 아니다

Protection은 한 process가 다른 process나 kernel을 망치지 못하게 하는 운영체제의 핵심 기능입니다. User/kernel mode, page permission, file permission, capability, seccomp, namespace는 모두 protection의 다른 얼굴입니다. 이 장치들은 안전을 주지만, 관측과 성능에도 영향을 줍니다. 예를 들어 seccomp가 특정 syscall을 막으면 profiling 도구가 실패할 수 있고, capability가 없으면 `tcpdump`나 eBPF program load가 안 될 수 있습니다. Namespace가 다르면 process가 보는 network interface와 filesystem이 host와 다릅니다.

```
debug command fails in container
  -> missing binary?
  -> missing permission/capability?
  -> seccomp denied syscall?
  -> namespace hides target?
  -> host policy blocks perf/eBPF?
```

보안 제한을 우회하라는 뜻이 아닙니다. 오히려 반대입니다. 관측을 설계할 때 어떤 권한이 필요한지, production에서 허용 가능한지, 대체 metric은 무엇인지 알아야 합니다. Kafka/Cassandra/Spark 장애를 볼 때 container 안에서 host network packet을 못 잡는다면, sidecar나 node-level agent, application metric, cloud flow log 같은 다른 관측 경로를 준비해야 합니다.

## Small experiment: lock wait와 CPU 사용률을 분리하기

작은 실험으로 두 thread가 같은 lock을 두고 경쟁하게 만들면 CPU와 elapsed time의 차이를 볼 수 있습니다. Thread A가 lock을 잡고 sleep하거나 I/O를 기다리고, Thread B가 같은 lock을 기다리면 B는 CPU를 많이 쓰지 않지만 요청 시간은 늘어납니다. Spin lock이라면 B가 CPU를 태울 수 있고, blocking mutex라면 futex wait로 내려갈 수 있습니다.

```
blocking mutex case
  Thread A locks and sleeps
  Thread B tries lock
  -> B sleeps in futex wait
  -> low CPU, high elapsed latency

spin case
  Thread B loops checking lock
  -> high CPU, still no progress
```

이 차이는 운영에서 매우 중요합니다. CPU가 낮은 lock wait와 CPU가 높은 spin contention은 해결책이 다릅니다. 전자는 lock hold time과 I/O-in-critical-section을 줄이는 것이 먼저이고, 후자는 spin policy, contention reduction, data partitioning을 봐야 합니다. Thread dump와 CPU profile, futex trace를 함께 보아야 구분됩니다.

## Product bridge: 하나의 증상을 세 갈래로 쪼개기

Kafka consumer lag가 늘었다고 해 보겠습니다. 동시성 관점에서는 consumer application의 worker lock, offset commit serialization, downstream DB connection pool을 봅니다. OS 관점에서는 consumer process의 scheduler delay, socket receive queue, disk write, GC pause를 봅니다. Distributed 관점에서는 broker fetch latency, partition skew, rebalance, group coordinator 상태를 봅니다. 같은 metric 하나가 세 층의 원인과 양립합니다.

```
consumer lag
  application: worker queue, lock, downstream sink
  OS: CPU quota, socket queue, GC, disk I/O
  distributed: partition skew, rebalance, broker/replica lag
```

Cassandra read timeout도 마찬가지입니다. 동시성으로는 request executor와 shared lock, memtable/SSTable access contention을 봅니다. OS로는 disk queue, page cache miss, network retransmission, cgroup throttle을 봅니다. 분산으로는 coordinator/replica path, consistency level, hinted handoff, repair 상태를 봅니다. Spark straggler는 task skew, shuffle fetch, executor GC, local disk spill, scheduler delay가 함께 후보입니다.

이렇게 쪼개는 습관이 있으면 "느리다"라는 문장을 바로 설정 변경으로 연결하지 않습니다. 먼저 증상이 어느 층의 대기인지 좁히고, 관측 도구를 선택하고, 제품의 semantics와 다시 연결합니다. 동시성, isolation, observability는 서로 다른 장처럼 보이지만 실제 장애 대응에서는 한 번에 필요합니다.
