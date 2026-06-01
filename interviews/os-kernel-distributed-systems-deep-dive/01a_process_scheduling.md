# 01a. Process, Thread, Scheduling

## 목차

- [왜 process라는 단위가 필요했는가](#왜-process라는-단위가-필요했는가)
- [fork, exec, clone을 "프로그램 실행" 한 단어로 뭉개면 안 된다](#fork-exec-clone을-프로그램-실행-한-단어로-뭉개면-안-된다)
- [scheduler는 CPU를 "공평하게" 주는 기계가 아니라 latency와 throughput을 조절한다](#scheduler는-cpu를-공평하게-주는-기계가-아니라-latency와-throughput을-조절한다)
- [sleep, blocked, runnable은 관측 도구마다 다르게 보인다](#sleep-blocked-runnable은-관측-도구마다-다르게-보인다)
- [signal, wait, zombie는 process lifecycle의 뒷정리다](#signal-wait-zombie는-process-lifecycle의-뒷정리다)
- [NUMA와 CPU affinity는 "CPU가 몇 개냐"보다 아래 질문이다](#numa와-cpu-affinity는-cpu가-몇-개냐보다-아래-질문이다)
- [제품으로 다시 연결하기](#제품으로-다시-연결하기)
- [스케줄링 정책은 공정함, 응답성, 처리량 사이의 선택이다](#스케줄링-정책은-공정함-응답성-처리량-사이의-선택이다)
- [Context switch 비용은 register 저장보다 넓다](#context-switch-비용은-register-저장보다-넓다)
- [Priority inversion, starvation, livelock은 서버에서도 일어난다](#priority-inversion-starvation-livelock은-서버에서도-일어난다)
- [Scheduler 관측은 on-CPU와 off-CPU를 나눠야 한다](#scheduler-관측은-on-cpu와-off-cpu를-나눠야-한다)
- [cgroup CPU quota와 container scheduling](#cgroup-cpu-quota와-container-scheduling)
- [Interview replay: 요청 하나가 CPU를 기다리는 전체 문장](#interview-replay-요청-하나가-cpu를-기다리는-전체-문장)
- [Work stealing, queue ownership, partition affinity를 같이 보기](#work-stealing-queue-ownership-partition-affinity를-같이-보기)
- [Time slice와 tail latency의 관계](#time-slice와-tail-latency의-관계)
- [Signal과 graceful shutdown은 scheduler와 lifecycle의 마지막 시험이다](#signal과-graceful-shutdown은-scheduler와-lifecycle의-마지막-시험이다)
- [문서를 덮고 실행해 볼 작은 관측](#문서를-덮고-실행해-볼-작은-관측)

백엔드 서버에서 요청 하나가 늦어졌다고 할 때, 그 요청은 추상적인 "서버" 안에서 기다린 것이 아닙니다. 어느 시점에는 application queue에 있었고, 어느 시점에는 Java thread나 native thread가 되었고, 어느 시점에는 OS scheduler가 CPU를 줄 때까지 runnable queue에서 기다렸거나, system call 안에서 sleep 상태로 내려갔습니다. process와 thread, scheduler를 모르면 Kafka network thread, Cassandra compaction thread, Spark executor task가 왜 서로 영향을 주는지 설명할 수 없습니다.

이 문서의 목표는 process와 thread라는 단어를 외우는 것이 아닙니다. `accept()`로 들어온 연결이 어떤 실행 흐름 위에서 request handler가 되고, 그 실행 흐름이 언제 CPU를 받지 못하며, 언제 block되어 scheduler에게 CPU를 돌려주는지를 말할 수 있게 되는 것입니다.

## 왜 process라는 단위가 필요했는가

초기의 프로그램 실행 모델은 지금처럼 수십 개 service process와 수백 개 thread가 동시에 떠 있는 모습이 아니었습니다. 한 기계가 한 번에 하나의 작업을 길게 수행하던 시기에는 "내 프로그램이 CPU와 메모리를 쓴다"라는 단순한 모델이 어느 정도 통했습니다. 하지만 비싼 컴퓨터를 여러 사용자가 함께 쓰고, I/O를 기다리는 동안 다른 일을 실행해야 한다는 요구가 커지면서 운영체제는 두 가지를 동시에 해결해야 했습니다.

첫째, 여러 실행 흐름이 한 CPU를 번갈아 쓰게 해야 했습니다. 둘째, 한 프로그램이 다른 프로그램의 메모리와 파일을 함부로 망치지 못하게 해야 했습니다. process는 이 두 요구를 묶는 기본 단위가 되었습니다. process는 실행 중인 프로그램의 코드와 데이터만 뜻하지 않습니다. 주소 공간, 열린 파일 목록, signal 처리 상태, 자원 제한, credential 같은 실행 문맥을 함께 가진 묶음입니다.

```text
process
  address space: code, heap, stack, mmap regions
  file descriptor table: 0, 1, 2, 3 ...
  credentials: uid, gid, capabilities
  signal dispositions
  resource limits
  one or more threads
```

thread는 CPU가 실제로 실행하는 흐름입니다. 같은 process 안의 thread들은 주소 공간과 파일 디스크립터 테이블을 공유하지만, 각 thread는 자기 register 상태와 stack, scheduler 상태를 갖습니다. 그래서 thread 간 통신은 process 간 통신보다 싸지만, 같은 heap을 동시에 만지므로 race와 lock 문제가 생깁니다.

## fork, exec, clone을 "프로그램 실행" 한 단어로 뭉개면 안 된다

Unix 계열 시스템에서 새 프로그램이 실행되는 흐름을 이해하려면 `fork()`와 `execve()`를 분리해야 합니다. `fork()`는 현재 process를 거의 같은 모습의 child process로 복제합니다. 현대 커널은 copy-on-write를 사용해 물리 메모리를 즉시 전부 복사하지 않고, 부모와 자식이 같은 page를 공유하다가 한쪽이 쓰려 할 때 page fault를 통해 분리할 수 있습니다. `execve()`는 현재 process의 주소 공간을 새 프로그램 이미지로 바꿉니다. PID와 일부 process 문맥은 이어지지만, user code와 data는 새 실행 파일에서 온 내용으로 바뀝니다.

```text
shell process
  fork()
    -> child has copied process context
  child execve("/usr/bin/java", ...)
    -> child address space becomes JVM process image
  parent wait()
    -> later observes child's exit status
```

Linux의 `clone()` 계열은 process와 thread의 경계를 더 세밀하게 나눕니다. 주소 공간을 공유할지, 파일 디스크립터 테이블을 공유할지, signal handler를 공유할지 같은 flag 조합으로 process-like 또는 thread-like 실행 단위를 만들 수 있습니다. 그래서 Linux kernel 내부에서는 process와 thread 모두 task로 표현되는 면이 있습니다. 면접에서 이 세부 구조를 말할 때는 "process는 자원 묶음, thread는 scheduling되는 실행 흐름"이라는 개발자 관점의 모델을 먼저 고정하고, Linux에서는 둘 다 scheduler가 다루는 task라는 구현 관점이 이어진다고 말하면 안전합니다.

## scheduler는 CPU를 "공평하게" 주는 기계가 아니라 latency와 throughput을 조절한다

CPU core 하나는 한 순간에 하나의 thread만 실행합니다. runnable thread가 core 수보다 많으면 누군가는 기다려야 합니다. scheduler는 runnable queue에 있는 task 중 어느 task가 다음에 CPU를 받을지 고릅니다. context switch는 현재 task의 register와 kernel bookkeeping 상태를 저장하고, 다음 task의 상태를 복원해 CPU가 다른 실행 흐름을 이어가게 하는 일입니다.

```text
time slice expires or task blocks
  -> enter kernel
  -> save current task CPU state
  -> pick next runnable task
  -> restore next task CPU state
  -> return to user mode or continue kernel path
```

여기서 중요한 점은 "thread가 많다"와 "일을 많이 처리한다"가 같은 말이 아니라는 것입니다. CPU-bound thread가 core 수보다 많으면 runnable queue가 길어지고 context switch 비용이 늘어납니다. I/O-bound thread는 자주 sleep 상태로 내려가 CPU를 양보하지만, 깨어난 뒤에도 run queue에서 다시 기다릴 수 있습니다. lock을 기다리는 thread는 CPU를 못 쓰고, futex나 condition variable을 통해 kernel wait queue에 들어갈 수 있습니다.

```text
request thread timeline

user code parsing
  -> read() on socket would block
  -> scheduler runs another task
  -> packet arrives, socket becomes readable
  -> thread becomes runnable
  -> waits in run queue
  -> gets CPU
  -> continues request handling
```

이 trace는 "network packet이 도착했다"와 "application thread가 바로 request를 처리했다"가 다른 말임을 보여 줍니다. packet 도착은 socket 상태를 바꿉니다. thread가 CPU를 다시 받는 것은 scheduler의 별도 결정입니다.

## sleep, blocked, runnable은 관측 도구마다 다르게 보인다

운영 장애에서 자주 하는 실수는 JVM thread dump의 `RUNNABLE`을 OS CPU running과 같은 뜻으로 읽는 것입니다. Java thread dump에서 `RUNNABLE`은 Java 수준에서 실행 가능하거나 native method/system call 안에 있다는 뜻으로 보일 수 있습니다. 그 thread가 실제 CPU 위에서 계산 중인지, kernel I/O 안에서 기다리는지는 `perf`, `strace`, `top`, `pidstat`, off-CPU profiling 같은 다른 관측과 함께 봐야 합니다.

| 관측 | 성급한 해석 | 더 안전한 질문 |
| --- | --- | --- |
| CPU 사용률 낮음 | 여유가 많다 | thread들이 I/O, lock, futex, GC, scheduler wait 중 어디에 있는가? |
| runnable thread 많음 | CPU만 늘리면 된다 | runnable이 CPU-bound인지, lock owner 하나를 기다리는 fan-in인지 확인했는가? |
| context switch 많음 | 커널이 느리다 | 짧은 blocking call이 많아진 것인지, thread 수가 과한 것인지, lock handoff가 잦은 것인지 보았는가? |
| Java `RUNNABLE` | CPU에서 실행 중이다 | native I/O wait나 kernel path 안의 대기를 포함하는가? |

Kafka broker에서 request handler thread가 많아도 disk flush나 replica fetch가 병목이면 처리량은 늘지 않습니다. Cassandra에서 compaction thread를 늘리면 compaction backlog는 줄 수 있지만 foreground read/write와 disk bandwidth를 더 세게 경쟁할 수 있습니다. Spark executor core 수를 크게 잡으면 task 병렬성은 늘 수 있지만 heap pressure, GC, shuffle spill, remote fetch 경쟁이 같이 커질 수 있습니다.

## signal, wait, zombie는 process lifecycle의 뒷정리다

process는 실행이 끝나도 곧바로 모든 흔적이 사라지지 않습니다. child process가 종료하면 kernel은 exit status 같은 최소 정보를 남겨 parent가 `wait()`로 회수할 수 있게 합니다. parent가 회수하기 전의 child는 zombie입니다. zombie는 CPU를 쓰는 살아 있는 process가 아니라, parent에게 종료 사실을 전달하기 위해 process table에 남아 있는 작은 기록입니다.

signal은 process나 thread에 비동기 사건을 전달하는 Unix식 장치입니다. `SIGTERM`은 종료 요청, `SIGKILL`은 잡을 수 없는 강제 종료, `SIGCHLD`는 child 상태 변화 알림처럼 쓰입니다. 서버 종료 시 graceful shutdown이 어려운 이유는 signal 하나가 곧바로 "모든 request가 안전하게 끝났다"를 뜻하지 않기 때문입니다. accept loop를 멈추고, inflight request를 기다리고, offset이나 transaction을 정리하고, file/socket을 닫는 application-level protocol이 별도로 필요합니다.

```text
SIGTERM received
  -> stop accepting new work
  -> mark service draining
  -> wait for in-flight work or deadline
  -> commit/rollback/flush needed state
  -> close sockets and exit
  -> parent or supervisor observes exit status
```

이 흐름은 Kafka consumer shutdown, Spark executor decommission, Cassandra node drain 같은 제품 동작을 볼 때 그대로 다시 등장합니다. process 종료는 OS 사건이고, 안전한 서비스 종료는 application protocol입니다.

## NUMA와 CPU affinity는 "CPU가 몇 개냐"보다 아래 질문이다

큰 서버에서는 모든 CPU core가 모든 memory에 같은 비용으로 접근하지 않을 수 있습니다. NUMA(non-uniform memory access)는 CPU socket과 memory bank의 물리적 위치에 따라 접근 비용이 달라지는 구조입니다. Linux scheduler와 memory allocator는 locality를 고려하지만, thread migration, interrupt affinity, memory placement가 어긋나면 같은 CPU 사용률에서도 latency가 달라질 수 있습니다.

CPU affinity는 특정 process/thread를 특정 CPU 집합에서 실행되게 묶는 설정입니다. 무조건 좋은 최적화가 아니라, cache locality나 interrupt 처리 위치를 맞추는 대신 scheduler의 자유도를 줄이는 tradeoff입니다. 고성능 network server나 low-latency pipeline에서 NIC queue interrupt, NAPI poll, application worker를 같은 NUMA node에 맞추려는 이유가 여기 있습니다.

## 제품으로 다시 연결하기

Kafka, Cassandra, Spark는 모두 JVM process이지만 scheduler 관점에서는 서로 다른 thread와 kernel wait가 뒤섞인 workload입니다.

- Kafka broker는 network thread, request handler, replica fetcher, log cleaner가 CPU와 disk/network를 나눠 씁니다. producer latency가 늘었을 때 request queue만 보지 말고 해당 thread들이 runnable인지, I/O sleep인지, lock을 기다리는지 봐야 합니다.
- Cassandra는 foreground request thread와 background compaction/repair thread가 같은 CPU와 disk bandwidth를 공유합니다. compaction thread를 늘리는 것은 queue 하나를 줄이는 대신 다른 queue의 latency를 키울 수 있습니다.
- Spark executor는 task thread가 많을수록 병렬성이 늘 수 있지만, executor heap과 local disk, shuffle fetch가 같이 압박을 받습니다. core 수, task 수, partition 수는 scheduler와 resource pressure를 함께 보는 설정입니다.

문서를 덮고 아래 trace를 말할 수 있어야 합니다.

```text
incoming request
  -> socket readable
  -> application thread runnable
  -> scheduler grants CPU
  -> user code runs
  -> blocking syscall or lock wait
  -> scheduler runs someone else
  -> wakeup event
  -> runnable again
```

이 흐름을 말할 수 있으면 "CPU는 낮은데 p99가 높다"는 질문을 감으로 답하지 않고, run queue, off-CPU wait, lock, I/O, GC, downstream queue로 나눠 볼 수 있습니다.

## 스케줄링 정책은 공정함, 응답성, 처리량 사이의 선택이다

스케줄링 알고리즘을 외울 때 흔히 round-robin, priority, multilevel feedback queue 같은 이름부터 잡습니다. 하지만 면접이나 운영 상황에서는 이름보다 선택 기준이 먼저입니다. 어떤 정책은 모든 runnable task에게 비슷한 기회를 주려 하고, 어떤 정책은 interactive task의 짧은 응답을 우선하며, 어떤 정책은 batch throughput을 우선합니다. 서버 workload에서는 평균 처리량뿐 아니라 tail latency도 중요합니다. 하나의 정책이 모든 목표를 동시에 만족시키지는 못합니다.

Round-robin은 task를 돌아가며 일정 시간씩 실행하는 단순한 모델입니다. Time slice가 너무 짧으면 context switch가 잦아지고 cache locality가 깨집니다. 너무 길면 한 task가 CPU를 오래 잡아 interactive responsiveness가 나빠집니다. Priority scheduling은 중요한 task를 먼저 실행할 수 있지만, 낮은 priority task가 계속 밀리는 starvation을 만들 수 있습니다. Multilevel feedback queue는 자주 잠드는 interactive task와 CPU를 오래 쓰는 batch task를 다르게 보며, task의 최근 행동을 기준으로 queue를 이동시켜 응답성과 처리량을 절충하려 합니다.

현대 Linux의 일반 task scheduling은 단순 round-robin으로 설명하기 어렵습니다. CFS 계열의 핵심 사고는 task가 자기 weight에 맞게 공정한 CPU 시간을 받도록 실행 시간을 accounting하는 것입니다. 여기서 weight, nice value, cgroup CPU share, runnable time, wakeup placement, CPU affinity가 함께 작동합니다. Real-time scheduling class는 또 다른 규칙을 갖습니다. 그러니 "리눅스 scheduler는 round-robin인가요?"라는 질문에는 "일반 task에는 단순 round-robin보다 공정한 CPU 시간 배분에 가까운 모델을 쓴다. 다만 scheduling class와 cgroup 설정에 따라 달라진다"라고 답하는 것이 안전합니다.

```text
two CPU-bound tasks, same weight
  A runs and accumulates virtual runtime
  B has smaller virtual runtime
  scheduler tends to pick B

one interactive task
  sleeps on I/O
  wakes up after event
  should not wait behind a long CPU hog forever
```

이 설명이 제품으로 내려오면 thread pool size와 연결됩니다. Kafka request handler를 늘리면 동시에 처리할 수 있는 request 수는 늘 수 있지만, CPU-bound 구간에서는 runnable 경쟁이 커집니다. Cassandra compaction thread를 늘리면 backlog가 줄 수 있지만, foreground read/write가 disk와 CPU를 더 많이 기다릴 수 있습니다. Spark executor core 수를 늘리면 task parallelism은 커지지만, shuffle spill, serialization, GC, remote fetch가 같이 늘 수 있습니다. 스케줄러는 thread 수를 무한히 흡수하는 기계가 아니라, 제한된 CPU 시간을 어떤 runnable 흐름에 줄지 결정하는 마지막 중재자입니다.

## Context switch 비용은 register 저장보다 넓다

Context switch를 "현재 register를 저장하고 다음 register를 복원한다"로만 외우면 비용을 작게 봅니다. 물론 직접 비용으로 register, program counter, stack pointer, kernel bookkeeping을 바꾸는 일이 있습니다. 하지만 실제 비용은 간접 효과까지 포함합니다. 다른 process로 바뀌면 주소 공간이 달라지고, TLB entry가 무효화되거나 새 주소 공간에 맞게 바뀌어야 할 수 있습니다. 같은 CPU cache에 남아 있던 data가 다른 task의 data로 밀려나면, 원래 task가 돌아왔을 때 cache miss가 늘어납니다.

```text
Task A works on hot data
  -> cache lines for A are warm
  -> scheduler switches to Task B
  -> B touches different data
  -> A resumes
  -> A's hot data may no longer be in cache
```

Thread가 많으면 동시에 기다릴 수 있는 작업은 늘지만, context switch와 cache disruption도 늘어납니다. 특히 짧은 CPU 작업과 짧은 blocking 작업이 너무 많은 thread에서 교차하면 scheduler는 바쁘고 실제 useful work는 줄 수 있습니다. Event-loop 기반 서버가 적은 수의 thread로 많은 connection을 다루려는 이유도 여기 있습니다. 모든 request를 thread 하나에 영구히 묶는 대신, readiness notification과 non-blocking I/O로 "지금 실제로 할 일이 있는 fd"만 처리하면 thread 수와 context switch를 줄일 수 있습니다.

반대로 thread-per-request 모델이 항상 나쁜 것도 아닙니다. 코드가 단순하고 blocking I/O가 많으며 thread 수가 bounded되어 있고 kernel scheduler가 충분히 잘 처리하는 workload에서는 좋은 선택일 수 있습니다. 핵심은 모델 자체를 선악으로 나누는 것이 아니라, "현재 workload의 대기 지점이 어디이며 thread 수가 그 대기를 감추는가, 아니면 새로운 contention을 만드는가"를 보는 것입니다.

## Priority inversion, starvation, livelock은 서버에서도 일어난다

우선순위 역전(priority inversion)은 낮은 priority task가 lock을 쥐고 있고 높은 priority task가 그 lock을 기다리는데, 중간 priority task들이 CPU를 계속 받아 lock holder가 실행될 기회를 얻지 못하는 상황입니다. 실시간 시스템의 고전 사례로 자주 배우지만, 백엔드 서버에서도 같은 모양이 나타납니다. Background maintenance thread가 metadata lock을 오래 잡고 있고 request thread가 그 lock을 기다리는데, 다른 CPU worker가 계속 실행되면 중요한 요청은 낮은 priority 작업 하나 때문에 밀립니다.

```text
background thread
  -> holds shared lock
  -> needs CPU to finish and release lock

request thread
  -> waits for that lock

other workers
  -> consume CPU
  -> background thread releases lock late
  -> user-visible latency grows
```

Starvation은 어떤 작업이 계속 자원을 받지 못하는 상태입니다. CPU뿐 아니라 lock, disk queue, connection pool, partition queue에서도 생깁니다. Livelock은 모두가 계속 움직이지만 유용한 진전이 거의 없는 상태입니다. 과한 retry가 대표적입니다. 요청이 timeout되고, client가 즉시 재시도하고, 서버는 원래 요청과 재시도를 모두 처리하려다 더 느려지고, timeout이 더 늘어나는 흐름은 시스템이 바쁘게 움직이지만 성공률을 낮춥니다.

이런 문제를 줄이는 방법은 하나가 아닙니다. Priority inheritance는 lock을 가진 낮은 priority task가 높은 priority task를 막고 있을 때 일시적으로 priority를 올려 주는 방식입니다. Lock hold time을 줄이고, lock을 shard하거나, background job의 concurrency와 I/O rate를 제한하고, queue에 backpressure를 걸고, retry에 exponential backoff와 jitter를 넣는 것도 해결책입니다. Kafka log cleaner, Cassandra compaction, Spark shuffle service 같은 background path는 "남는 시간에만 도는 일"이 아니라 foreground request와 같은 CPU, disk, memory, network를 공유하는 참가자입니다.

## Scheduler 관측은 on-CPU와 off-CPU를 나눠야 한다

프로파일링을 할 때 CPU flame graph만 보면 CPU에서 실행된 시간만 보입니다. 하지만 tail latency의 큰 부분은 CPU 밖에서 기다린 시간일 수 있습니다. Off-CPU profiling은 thread가 실행되지 못한 시간을 봅니다. 왜 sleep했는지, futex wait였는지, disk I/O wait였는지, network read였는지, scheduler delay였는지 분리하면 "코드가 느리다"와 "코드가 기다린다"를 구분할 수 있습니다.

Linux에서는 `perf sched`, eBPF 기반 off-CPU trace, `/proc/schedstat`, `pidstat -w`, `vmstat`의 run queue와 context switch 같은 지표가 도움이 됩니다. JVM에서는 async-profiler의 wall-clock/off-CPU 모드, Java Flight Recorder, thread dump를 함께 볼 수 있습니다. `RUNNABLE` thread가 CPU에서 burning 중인지, native syscall 안에서 기다리는지, safepoint에 묶였는지는 단일 도구로 닫히지 않을 수 있습니다.

```text
high p99 latency
  -> CPU profile shows little user code time
  -> thread dump shows many runnable/native frames
  -> off-CPU trace shows futex wait and disk wait
  -> fix target moves from algorithm to lock/I/O/backpressure
```

이 관측 방식은 면접에서도 좋은 답변이 됩니다. "CPU 사용률이 낮은데 latency가 높으면 무엇을 보겠습니까?"라는 질문에는 "CPU가 남는지보다 request thread가 어디서 시간을 보냈는지 보겠습니다. On-CPU time, run queue delay, lock/futex wait, disk/network I/O wait, GC pause, cgroup throttling을 분리하겠습니다"라고 답할 수 있습니다.

## cgroup CPU quota와 container scheduling

컨테이너에서 CPU 문제가 헷갈리는 이유는 process가 보는 CPU와 실제 사용할 수 있는 CPU 시간이 다를 수 있기 때문입니다. cpuset은 사용할 수 있는 CPU 집합을 제한할 수 있고, CPU quota는 일정 period 안에서 사용할 수 있는 CPU 시간을 제한합니다. Host 전체 CPU가 여유로워도 container가 자기 quota를 다 쓰면 throttle될 수 있습니다. 이때 애플리케이션 안에서는 thread가 runnable인데 주기적으로 멈추는 것처럼 보일 수 있습니다.

```text
cgroup period = 100 ms
quota = 200 ms
  -> container can use roughly 2 CPU cores worth of time per period
  -> many runnable threads consume quota early
  -> remaining period: throttled
  -> request latency gets periodic spikes
```

Kafka broker를 container에 넣고 CPU quota를 낮게 잡으면 network thread, request handler, replica fetcher, log cleaner가 한 cgroup 안에서 quota를 나눠 씁니다. Cassandra에서는 compaction과 foreground read/write가 같은 quota를 태웁니다. Spark executor에서는 task slot 수가 cgroup quota보다 과하면 runnable task가 많아지고 throttling이 tail latency와 stage runtime으로 올라올 수 있습니다. 그래서 containerized workload에서는 OS scheduler뿐 아니라 cgroup accounting을 함께 봐야 합니다.

관측은 `/sys/fs/cgroup`의 CPU stat, Kubernetes metrics, container runtime stats, application-level latency를 나란히 놓고 봅니다. CPU throttling count가 늘고 p99가 period 경계와 비슷한 리듬으로 튄다면, thread 수를 늘리는 해결은 반대 방향일 수 있습니다. CPU request/limit, executor core, Kafka thread pool, Cassandra concurrent settings를 같이 조정해야 합니다.

## Interview replay: 요청 하나가 CPU를 기다리는 전체 문장

이 장을 면접 답변으로 압축하면 다음 흐름이 됩니다. "요청이 들어오면 네트워크 stack이 socket buffer를 채우고, event loop나 worker thread가 깨어나 runnable 상태가 됩니다. 그 thread는 바로 실행되는 것이 아니라 scheduler가 CPU를 줄 때까지 run queue에서 기다릴 수 있습니다. 실행 중에는 사용자 코드 계산뿐 아니라 lock, futex, disk I/O, network I/O, GC, cgroup quota 때문에 다시 sleep하거나 throttled될 수 있습니다. 그래서 CPU 사용률과 latency를 직접 연결하지 않고, on-CPU time과 off-CPU wait를 나눠 봐야 합니다."

이 답변에 꼬리 질문이 오면 다음 기준으로 내려가면 됩니다. Process는 주소 공간과 파일, credential, signal 상태 같은 자원 묶음입니다. Thread는 그 안에서 CPU에 올라가는 실행 흐름입니다. Scheduler는 runnable thread 사이에서 CPU 시간을 배분합니다. Blocking I/O나 futex wait는 thread를 runnable에서 빼고, wakeup event가 오면 다시 runnable로 올립니다. Context switch는 register보다 cache/TLB locality까지 흔듭니다. NUMA와 affinity는 "어느 CPU에서 실행되는가"와 "어느 memory를 주로 읽는가"를 연결합니다. 이 정도를 말하면 단어 암기가 아니라 실제 서버 흐름을 이해하고 있다는 신호가 됩니다.

## Work stealing, queue ownership, partition affinity를 같이 보기

OS scheduler는 CPU를 배분하지만, 많은 runtime과 제품은 그 위에 자기만의 queue와 scheduler를 또 만듭니다. Java executor, Netty event loop, Kafka request queue, Cassandra stage executor, Spark task scheduler는 모두 "할 일을 어떤 worker에게 줄 것인가"를 결정합니다. 이 계층을 OS scheduler와 분리해서 보면 절반만 보는 것입니다. User-space scheduler가 어떤 thread를 깨우고, OS scheduler가 그 thread에 언제 CPU를 주며, thread가 다시 어떤 queue에서 다음 일을 가져오는지가 이어져야 합니다.

```text
product scheduler
  -> chooses work item for worker thread
  -> worker becomes runnable or stays busy

OS scheduler
  -> chooses runnable thread for CPU
  -> CPU executes worker

worker
  -> touches partition/cache/socket/file
  -> may block or enqueue downstream work
```

Work stealing은 idle worker가 다른 worker queue에서 일을 훔쳐 오는 방식입니다. 전체 처리량을 높일 수 있지만, cache locality와 partition ordering을 깨뜨릴 수 있습니다. Partition affinity는 특정 key, partition, shard의 일을 같은 worker에 배치해 locality와 ordering을 얻으려는 방식입니다. Kafka는 partition order와 request 처리 경로가 있고, Cassandra는 token range와 memtable/SSTable locality가 있으며, Spark는 partition별 task와 data locality가 있습니다. 일을 아무 worker에게나 주면 CPU utilization은 좋아 보일 수 있지만, cache miss, lock contention, ordering repair, remote fetch가 늘 수 있습니다.

이 관점은 "thread pool을 늘릴까요?"라는 질문에 더 좋은 답을 줍니다. Queue가 CPU-bound work로 꽉 찼고 run queue도 길다면 thread를 늘려도 scheduler 경쟁만 늘 수 있습니다. Queue는 길지만 worker들이 disk I/O에서 자고 있다면 I/O parallelism과 device queue depth를 봐야 합니다. Queue가 특정 partition에만 몰리면 worker 수보다 partition skew와 routing을 봐야 합니다. Queue가 짧은데 latency가 길면 downstream wait나 lock을 봐야 합니다.

## Time slice와 tail latency의 관계

Time slice는 task가 한 번 CPU를 받았을 때 계속 실행될 수 있는 시간 감각입니다. 실제 Linux scheduling은 단순 고정 time slice만으로 설명되지 않지만, "짧게 번갈아 실행한다"는 비용은 여전히 중요합니다. Tail latency가 중요한 서버에서는 긴 CPU-bound 작업 하나가 event loop나 request thread를 밀어내는 순간 p99가 흔들립니다. 반대로 너무 잦은 preemption은 useful work보다 context switch와 cache miss를 늘립니다.

```text
event loop thread should run briefly and often
  -> read ready sockets
  -> enqueue work
  -> flush responses

CPU-heavy task on same limited cores
  -> event loop wakeup delayed
  -> all connections see extra latency
```

Kafka broker에서 log cleaner나 compression이 CPU를 많이 쓰면 network thread의 wakeup이 늦어질 수 있습니다. Cassandra compaction은 disk뿐 아니라 CPU도 씁니다. Spark executor에서 serialization, compression, user-defined function이 CPU를 오래 쓰면 shuffle fetch나 heartbeat 처리도 늦어질 수 있습니다. 이때 해결은 단순히 "CPU를 더 준다"가 아닐 수 있습니다. Thread class를 분리하고, background concurrency를 줄이고, CPU quota를 조정하고, event loop가 맡는 일을 줄이는 식으로 scheduling surface를 정리해야 합니다.

Tail latency를 볼 때는 평균 runnable wait보다 percentile을 봅니다. 대부분 request는 빠르지만 특정 순간 run queue가 burst로 길어지면 p99만 튈 수 있습니다. `runqlat` 같은 eBPF 도구나 scheduler trace는 runnable이 된 thread가 실제 CPU를 받기까지 기다린 시간을 보여 줄 수 있습니다. 이 값이 튀면 application 코드가 느린 것이 아니라 CPU scheduling delay가 request path에 섞인 것입니다.

## Signal과 graceful shutdown은 scheduler와 lifecycle의 마지막 시험이다

Process lifecycle의 끝도 scheduling과 관련됩니다. Service가 `SIGTERM`을 받으면 handler나 shutdown hook이 실행되어야 하고, 그 코드도 CPU를 받아야 합니다. 종료 deadline이 짧고 run queue가 길거나 GC pause가 겹치면 graceful shutdown이 마무리되기 전에 orchestrator가 `SIGKILL`을 보낼 수 있습니다. 그러면 log flush, offset commit, partition revoke, in-flight request drain이 중간에 끊길 수 있습니다.

```text
orchestrator sends SIGTERM
  -> process records pending signal
  -> user handler or runtime hook runs
  -> service stops accepting
  -> drains in-flight work
  -> flushes/commits necessary state
  -> exits before grace period

if deadline passes
  -> SIGKILL
  -> no user cleanup path
```

Kafka broker shutdown은 controller와 replica state, log flush, client connection close가 얽힙니다. Cassandra node drain은 memtable flush와 gossip state, coordinator behavior를 생각해야 합니다. Spark executor decommission은 running task와 shuffle block, driver heartbeat를 생각해야 합니다. OS signal은 시작일 뿐이고, 실제 안전한 종료는 product protocol과 scheduler가 함께 만들어야 합니다.

## 문서를 덮고 실행해 볼 작은 관측

Linux 환경이라면 간단한 CPU-bound loop와 sleep loop를 동시에 돌려 `pidstat -w`, `top -H`, `perf sched` 또는 eBPF run queue 도구로 차이를 볼 수 있습니다. CPU-bound thread는 runnable 상태로 CPU를 기다리고, sleep loop는 timer나 I/O event 전까지 runnable이 아닙니다. 같은 "느린 프로그램"이어도 하나는 CPU time 경쟁이고, 다른 하나는 wakeup과 wait의 문제입니다.

```text
CPU-bound worker
  -> high on-CPU time
  -> run queue delay if cores are saturated

sleeping/I/O worker
  -> low CPU usage
  -> elapsed time dominated by wait and wakeup
```

이 작은 실험을 제품에 옮기면, Kafka producer latency가 CPU saturation 때문인지 socket/disk wait 때문인지, Cassandra read latency가 compaction CPU 때문인지 SSTable read wait 때문인지, Spark stage delay가 task scheduling 때문인지 shuffle fetch wait 때문인지 구분하는 습관이 생깁니다. Scheduler 공부의 끝은 알고리즘 이름이 아니라, request의 시간을 on-CPU와 off-CPU, runnable wait와 blocked wait로 나눠 설명하는 능력입니다.

마지막으로 한 문장을 더 붙이면, scheduler는 application 바깥의 블랙박스가 아닙니다. 애플리케이션이 thread 수, queue 구조, blocking call, retry, background job, CPU quota를 어떻게 설계하느냐가 scheduler에게 전달되는 입력을 바꿉니다. 좋은 백엔드 설계는 OS scheduler를 이기려 하지 않고, scheduler가 예측 가능한 작은 실행 단위와 bounded queue를 보도록 도와줍니다.
