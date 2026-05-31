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
