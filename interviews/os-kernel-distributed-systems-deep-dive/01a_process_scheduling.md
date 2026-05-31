# 01a. Process, Thread, Scheduling

백엔드 서버에서 요청 하나가 늦어졌다고 할 때, 그 요청은 추상적인 "서버" 안에서 기다린 것이 아닙니다. 어느 시점에는 application queue에 있었고, 어느 시점에는 Java thread나 native thread가 되었고, 어느 시점에는 OS scheduler가 CPU를 줄 때까지 runnable queue에서 기다렸거나, system call 안에서 sleep 상태로 내려갔습니다. process와 thread, scheduler를 모르면 Kafka network thread, Cassandra compaction thread, Spark executor task가 왜 서로 영향을 주는지 설명할 수 없습니다.

이 문서의 목표는 process와 thread라는 단어를 외우는 것이 아닙니다. `accept()`로 들어온 연결이 어떤 실행 흐름 위에서 request handler가 되고, 그 실행 흐름이 언제 CPU를 받지 못하며, 언제 block되어 scheduler에게 CPU를 돌려주는지를 말할 수 있게 되는 것입니다.

## 왜 process라는 단위가 필요했는가

초기의 프로그램 실행 모델은 지금처럼 수십 개 service process와 수백 개 thread가 동시에 떠 있는 모습이 아니었습니다. 한 기계가 한 번에 하나의 작업을 길게 수행하던 시기에는 "내 프로그램이 CPU와 메모리를 쓴다"라는 단순한 모델이 어느 정도 통했습니다. 하지만 비싼 컴퓨터를 여러 사용자가 함께 쓰고, I/O를 기다리는 동안 다른 일을 실행해야 한다는 요구가 커지면서 운영체제는 두 가지를 동시에 해결해야 했습니다.

첫째, 여러 실행 흐름이 한 CPU를 번갈아 쓰게 해야 했습니다. 둘째, 한 프로그램이 다른 프로그램의 메모리와 파일을 함부로 망치지 못하게 해야 했습니다. process는 이 두 요구를 묶는 기본 단위가 되었습니다. process는 실행 중인 프로그램의 코드와 데이터만 뜻하지 않습니다. 주소 공간, 열린 파일 목록, signal 처리 상태, 자원 제한, credential 같은 실행 문맥을 함께 가진 묶음입니다.

```
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

```
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

```
time slice expires or task blocks
  -> enter kernel
  -> save current task CPU state
  -> pick next runnable task
  -> restore next task CPU state
  -> return to user mode or continue kernel path
```

여기서 중요한 점은 "thread가 많다"와 "일을 많이 처리한다"가 같은 말이 아니라는 것입니다. CPU-bound thread가 core 수보다 많으면 runnable queue가 길어지고 context switch 비용이 늘어납니다. I/O-bound thread는 자주 sleep 상태로 내려가 CPU를 양보하지만, 깨어난 뒤에도 run queue에서 다시 기다릴 수 있습니다. lock을 기다리는 thread는 CPU를 못 쓰고, futex나 condition variable을 통해 kernel wait queue에 들어갈 수 있습니다.

```
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
|---|---|---|
| CPU 사용률 낮음 | 여유가 많다 | thread들이 I/O, lock, futex, GC, scheduler wait 중 어디에 있는가? |
| runnable thread 많음 | CPU만 늘리면 된다 | runnable이 CPU-bound인지, lock owner 하나를 기다리는 fan-in인지 확인했는가? |
| context switch 많음 | 커널이 느리다 | 짧은 blocking call이 많아진 것인지, thread 수가 과한 것인지, lock handoff가 잦은 것인지 보았는가? |
| Java `RUNNABLE` | CPU에서 실행 중이다 | native I/O wait나 kernel path 안의 대기를 포함하는가? |

Kafka broker에서 request handler thread가 많아도 disk flush나 replica fetch가 병목이면 처리량은 늘지 않습니다. Cassandra에서 compaction thread를 늘리면 compaction backlog는 줄 수 있지만 foreground read/write와 disk bandwidth를 더 세게 경쟁할 수 있습니다. Spark executor core 수를 크게 잡으면 task 병렬성은 늘 수 있지만 heap pressure, GC, shuffle spill, remote fetch 경쟁이 같이 커질 수 있습니다.

## signal, wait, zombie는 process lifecycle의 뒷정리다

process는 실행이 끝나도 곧바로 모든 흔적이 사라지지 않습니다. child process가 종료하면 kernel은 exit status 같은 최소 정보를 남겨 parent가 `wait()`로 회수할 수 있게 합니다. parent가 회수하기 전의 child는 zombie입니다. zombie는 CPU를 쓰는 살아 있는 process가 아니라, parent에게 종료 사실을 전달하기 위해 process table에 남아 있는 작은 기록입니다.

signal은 process나 thread에 비동기 사건을 전달하는 Unix식 장치입니다. `SIGTERM`은 종료 요청, `SIGKILL`은 잡을 수 없는 강제 종료, `SIGCHLD`는 child 상태 변화 알림처럼 쓰입니다. 서버 종료 시 graceful shutdown이 어려운 이유는 signal 하나가 곧바로 "모든 request가 안전하게 끝났다"를 뜻하지 않기 때문입니다. accept loop를 멈추고, inflight request를 기다리고, offset이나 transaction을 정리하고, file/socket을 닫는 application-level protocol이 별도로 필요합니다.

```
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

```
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
