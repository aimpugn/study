# 실무 엔지니어를 위한 Linux 커널·하드웨어 내부 구조

이 문서는 인프라 엔지니어, DevOps, 개발자 플랫폼 엔지니어, 백엔드 개발자가 장애와 면접에서 같은 언어로 설명할 수 있어야 하는 Linux 커널·하드웨어 내부 구조를 한 파일에 묶은 교본이다. VMware나 특정 모니터링 스크립트 사용법이 중심이 아니다. 특정 환경에서 수집한 지표는 출발점일 뿐이고, 진짜 목표는 애플리케이션 요청이 CPU, 메모리, 파일 시스템, 블록 장치, 네트워크, DB 엔진, JVM 런타임을 통과할 때 어떤 커널 객체와 하드웨어 자원을 쓰는지 이해하는 것이다.

이 문서를 읽을 때 가장 먼저 붙잡을 단어는 `queue`, `cache`, `buffer`, `wait`다. 운영 장애는 대개 "자원이 없다"라는 말보다 더 구체적으로, 어느 층의 queue가 길어졌는지, 어떤 cache가 비었거나 너무 커졌는지, 어떤 buffer가 막혔는지, 어떤 thread나 process가 어떤 wait 상태에 들어갔는지로 설명된다. `vmstat`, `iostat`, `ss`, `pidstat`, `dmesg` 같은 명령은 그 상태의 그림자를 보여 줄 뿐이다. 따라서 명령을 외우기 전에, 그 명령이 어떤 커널 상태를 비추는지 먼저 알아야 한다.

## 목차

- [짧은 직답](#짧은-직답)
- [한 요청이 지나가는 실제 경로](#한-요청이-지나가는-실제-경로)
- [사고의 기본 단위: queue, cache, buffer, wait](#사고의-기본-단위-queue-cache-buffer-wait)
- [하드웨어를 OS 관점에서 다시 보기](#하드웨어를-os-관점에서-다시-보기)
- [CPU와 scheduler](#cpu와-scheduler)
- [메모리, page cache, reclaim, swap, OOM, cgroup](#메모리-page-cache-reclaim-swap-oom-cgroup)
- [파일 시스템, block I/O, fsync, writeback, device queue](#파일-시스템-block-io-fsync-writeback-device-queue)
- [네트워크, NIC, TCP, socket buffer, backlog, retransmit](#네트워크-nic-tcp-socket-buffer-backlog-retransmit)
- [명령과 지표를 커널 객체로 해석하기](#명령과-지표를-커널-객체로-해석하기)
- [DB는 OS 위에서 가장 강한 애플리케이션이다](#db는-os-위에서-가장-강한-애플리케이션이다)
- [JVM, Spring, Kafka, Cassandra, Spark가 OS 자원을 쓰는 방식](#jvm-spring-kafka-cassandra-spark가-os-자원을-쓰는-방식)
- [장애 대응 학습 구조](#장애-대응-학습-구조)
- [면접 답변 구조](#면접-답변-구조)
- [실전 분기 예시](#실전-분기-예시)
- [기존 문서 owner와 drift map](#기존-문서-owner와-drift-map)
- [반복 학습 질문](#반복-학습-질문)

## 짧은 직답

Linux 서버에서 애플리케이션 성능이나 장애를 본다는 것은 CPU 사용률만 보는 일이 아니다. 한 요청이 들어오면 네트워크 카드가 패킷을 받고, 커널은 socket buffer와 TCP 상태를 갱신하고, scheduler는 애플리케이션 thread에 CPU 시간을 배정하고, 그 thread는 syscall을 통해 파일, DB, 네트워크, 메모리를 사용한다. DB commit은 WAL 파일에 쓰고 `fsync`로 저장 장치까지 내리려 하고, 이때 page cache, dirty page, writeback, block layer queue, 장치 내부 queue가 이어진다. JVM 애플리케이션은 heap, GC, thread pool, native memory, socket, file descriptor, cgroup 제한을 통해 OS 자원을 쓴다.

장애가 나면 먼저 "어느 자원인가"보다 "어느 queue에서 기다리는가"를 묻는다. CPU run queue에서 기다리면 `vmstat r`, load average, `pidstat`, `perf`가 실마리를 준다. 디스크나 원격 스토리지에서 기다리면 `iostat await`, `aqu-sz`, `%util`, process state `D`, dirty/writeback page가 중요해진다. 메모리 압박이면 `MemAvailable`, `pgscan`, `pgsteal`, swap in/out, OOM log, cgroup memory event를 본다. 네트워크는 `ss -tin`, retransmit, send/receive queue, listen backlog, conntrack, NIC error/drop을 본다. DB는 lock wait, MVCC dead tuple, WAL fsync, checkpoint, compaction, buffer pool miss가 OS 지표와 만나 증상을 만든다.

면접에서 좋은 답변은 "CPU, 메모리, 디스크, 네트워크를 봅니다"에서 끝나지 않는다. "요청이 어디서 기다리는지부터 좁힙니다. CPU run queue인지, page reclaim인지, block device queue인지, TCP retransmit인지, DB lock인지 확인하고, 각 지표가 어떤 커널 객체와 wait state를 반영하는지 연결해 설명합니다"라고 말할 수 있어야 한다.

## 한 요청이 지나가는 실제 경로

먼저 가장 흔한 장면부터 보자. 사용자가 결제 API를 호출했고, Spring 애플리케이션이 PostgreSQL에 주문을 저장한 뒤 Kafka에 이벤트를 발행한다고 하자. 이 요청은 "애플리케이션 코드" 안에서만 처리되지 않는다.

```text
client
  -> NIC interrupt / NAPI polling
  -> kernel TCP stack
  -> socket receive buffer
  -> epoll/kqueue style readiness wakeup
  -> application worker thread
  -> JVM heap allocation and lock
  -> DB client socket write/read
  -> DB process backend thread or process
  -> DB lock table / MVCC snapshot / buffer pool
  -> WAL write
  -> filesystem page cache
  -> dirty page and writeback
  -> block layer queue
  -> storage controller / SSD / NVMe queue
  -> interrupt or completion queue
  -> DB response
  -> application response
  -> socket send buffer
  -> kernel TCP stack
  -> NIC transmit queue
```

이 흐름에서 "느리다"는 말은 너무 넓다. 느린 위치는 여러 곳일 수 있다.

- CPU가 부족하면 애플리케이션 worker thread나 DB backend process가 runnable 상태인데 CPU를 못 받는다.
- lock 때문에 막히면 thread는 CPU를 태우는 것이 아니라 mutex, futex, DB lock, transaction wait에서 깨어나기를 기다린다.
- 메모리가 부족하면 새 객체 할당, page fault, page reclaim, swap I/O, cgroup OOM으로 지연이 생긴다.
- WAL `fsync`가 늦으면 DB commit이 storage flush 완료를 기다린다.
- checkpoint나 compaction이 몰리면 background writer가 block I/O queue를 채우고 foreground 요청까지 늦어진다.
- TCP retransmit이 늘면 애플리케이션은 상대가 느린 것처럼 보지만 실제로는 패킷 손실, 혼잡, NIC/스위치/방화벽/커널 buffer 문제가 섞였을 수 있다.

따라서 실무자는 각 계층을 따로 외우기보다, 요청이 어떤 "대기 가능한 지점"을 지나가는지 기억해야 한다. 장애 대응은 이 대기 지점을 빠르게 지우거나 좁히는 일이다.

## 사고의 기본 단위: queue, cache, buffer, wait

### Queue

Queue는 처리 순서를 기다리는 줄이다. CPU run queue, block device queue, NIC transmit queue, listen backlog, DB lock wait queue, Kafka partition backlog가 모두 queue다. queue가 길어지는 이유는 보통 두 가지다. 들어오는 속도가 처리 속도보다 크거나, 처리자가 어떤 이유로 멈췄다.

실무에서 queue를 볼 때는 길이만 보지 않는다. 처리 시간이 같이 늘었는지, 처리량이 줄었는지, 어느 소비자가 느려졌는지를 같이 본다. Little's law는 단순하지만 강력한 감각을 준다.

```text
queue length ~= arrival rate * wait time
```

예를 들어 `iostat`의 `aqu-sz`는 block I/O 평균 queue 깊이에 가깝게 해석할 수 있고, `await`가 올라가면 같은 IOPS에서도 queue가 길어질 수 있다. 하지만 `%util` 하나만 보고 "디스크 100%라서 끝"이라고 말하면 부족하다. SSD/NVMe는 내부 병렬성이 있고, `%util`은 장치가 바빴던 시간 비율이지 사용자 요청이 느린 이유 전체를 뜻하지 않는다.

### Cache

Cache는 느린 계층에 덜 내려가기 위해 가까운 계층에 둔 복사본이다. CPU cache, TLB, page cache, DB buffer pool, Kafka page cache, Cassandra key cache와 row cache가 모두 cache다. cache가 잘 맞으면 latency가 줄고, 잘못 맞으면 eviction, reclaim, writeback, double caching, cache pollution이 생긴다.

Linux page cache는 파일 내용을 메모리에 올려 두는 커널 cache다. DB buffer pool은 DB가 page 단위로 관리하는 자체 cache다. 둘은 서로 다른 소유자가 관리한다. DB가 `O_DIRECT`를 쓰면 page cache를 우회할 수 있고, 일반 buffered I/O를 쓰면 DB buffer pool과 OS page cache가 동시에 같은 데이터를 품을 수 있다. 어느 쪽이 좋은지는 DB 엔진, workload, memory pressure, 운영 정책에 따라 달라진다.

### Buffer

Buffer는 생산자와 소비자의 속도 차이를 흡수하는 임시 저장 공간이다. socket receive buffer, socket send buffer, pipe buffer, page cache dirty page, DB WAL buffer, JVM direct buffer가 있다. buffer는 일시적 파동을 흡수하지만 무한하지 않다. buffer가 꽉 차면 backpressure가 생기거나 drop, blocking, timeout이 발생한다.

네트워크에서 receive buffer가 꽉 차면 커널은 더 이상 애플리케이션이 읽지 않는 데이터를 받기 어렵다. TCP window가 줄고, 상대는 전송을 늦춘다. 애플리케이션 입장에서는 "상대가 느리다"로 보일 수 있지만, 실제 원인은 내 프로세스가 socket을 빨리 읽지 못하는 것일 수 있다.

### Wait

Wait는 process나 thread가 CPU를 쓰지 않고 어떤 조건을 기다리는 상태다. Linux process state에서 `R`은 실행 중이거나 실행 가능하다는 뜻이고, `S`는 interruptible sleep, `D`는 uninterruptible sleep이다. `D` 상태는 흔히 디스크나 네트워크 파일 시스템 같은 커널 I/O 완료를 기다릴 때 보인다. Java thread dump의 `RUNNABLE`은 OS CPU 실행 가능 상태와 정확히 같지 않다. Java에서 native socket read 안에 들어가 있어도 JVM thread state가 `RUNNABLE`로 보일 수 있다. 그래서 JVM thread dump만 보고 CPU 병목이라고 단정하면 안 된다.

## 하드웨어를 OS 관점에서 다시 보기

OS는 하드웨어를 추상화하지만, 하드웨어의 물리적 제약을 없애지는 못한다. 커널과 애플리케이션 지표는 결국 하드웨어의 처리량, latency, 병렬성, locality 제약이 드러난 결과다.

### CPU core, cache, TLB

CPU core는 instruction을 실행한다. core가 많아도 모든 thread가 동시에 빠르게 도는 것은 아니다. scheduler는 runnable task를 core에 배치하고, task가 core를 옮기면 CPU cache locality가 깨질 수 있다. 같은 data를 여러 core가 자주 수정하면 cache coherence 비용이 생긴다. lock 경합, false sharing, atomic operation 남용은 CPU 사용률을 높이면서도 실제 처리량을 낮출 수 있다.

TLB는 virtual address를 physical address로 바꾸는 page table translation을 cache한다. memory access가 많고 working set이 크면 cache miss와 TLB miss가 늘 수 있다. huge page는 TLB pressure를 줄일 수 있지만, memory fragmentation, allocation 정책, DB/JVM 설정과 함께 봐야 한다.

### RAM, NUMA

RAM은 CPU가 직접 load/store하는 작업 공간이다. NUMA는 여러 CPU socket이나 memory controller가 있을 때, 어떤 CPU에서 어떤 memory bank에 접근하느냐에 따라 latency가 달라지는 구조다. DB buffer pool이나 JVM heap이 큰 서버에서 thread가 한 NUMA node에 있고 memory가 다른 node에 몰리면 remote memory access가 늘 수 있다. 이 문제는 단순 `free -h`로 보이지 않는다. `numactl`, `numastat`, CPU affinity, cgroup cpuset까지 같이 봐야 한다.

### Disk, SSD, NVMe, storage controller

디스크와 SSD는 block 단위 I/O를 처리한다. HDD는 seek와 rotation 때문에 random I/O에 약하고, SSD/NVMe는 내부 병렬성과 queue depth가 중요하다. NVMe는 여러 submission/completion queue를 통해 높은 병렬성을 낼 수 있다. 하지만 병렬성이 높다는 말은 queue가 길어져도 latency가 항상 안전하다는 뜻이 아니다. write amplification, garbage collection, flush/FUA, controller cache, RAID/HBA, SAN, NAS, hypervisor layer가 latency를 키울 수 있다.

DB commit의 `fsync`는 "파일에 썼다"가 아니라 "내구성 보장을 위해 저장 계층까지 밀어 넣었다"에 가깝다. 실제 보장 범위는 파일 시스템, mount option, storage cache, controller battery, device flush 구현에 따라 달라진다.

### NIC, HBA, DMA, interrupt

NIC는 network packet을 주고받는 장치이고, HBA는 storage network나 disk controller와 연결되는 adapter다. 고속 장치는 CPU가 매 byte를 직접 옮기게 하지 않는다. DMA는 장치가 memory에 직접 데이터를 쓰거나 읽게 하는 방식이다. 장치가 일을 마치면 interrupt나 polling 방식으로 CPU에 알린다. Linux 네트워크는 interrupt 폭주를 줄이기 위해 NAPI처럼 interrupt와 polling을 섞어 쓴다.

이 때문에 높은 network traffic에서는 단순 user CPU보다 system CPU, softirq, packet drop, NIC queue, driver 통계가 중요해질 수 있다.

## CPU와 scheduler

CPU 병목을 이해하려면 `CPU 사용률`보다 먼저 task 상태를 봐야 한다. Linux scheduler가 다루는 기본 대상은 task다. process와 thread는 구현상 모두 task로 취급된다. runnable task는 CPU를 받을 수 있는 상태이고, sleeping task는 어떤 event를 기다리는 상태다.

### scheduler가 실제로 하는 일

애플리케이션 thread가 CPU에서 코드를 실행하다가 socket read, disk read, futex wait, sleep, lock wait 같은 지점에 도달하면 kernel로 들어간다. 기다릴 수밖에 없는 조건이면 scheduler는 그 task를 sleep 상태로 두고 다른 runnable task를 CPU에 올린다. 반대로 I/O 완료 interrupt, timer, wakeup event, lock release가 발생하면 task는 다시 runnable이 된다.

```text
running on CPU
  -> syscall or fault or interrupt
  -> cannot proceed now
  -> sleep/wait queue
  -> wakeup event
  -> runnable queue
  -> scheduler picks it
  -> running on CPU
```

이 흐름에서 latency는 두 곳에서 생긴다. 하나는 기다리던 조건 자체가 늦게 만족되는 시간이다. 다른 하나는 조건은 만족됐는데 runnable queue에서 CPU를 받기까지 기다리는 시간이다.

### CPU 지표를 읽는 방법

- `load average`
    - runnable task와 uninterruptible sleep task가 섞여 반영된다. Linux load가 높다고 항상 CPU가 바쁜 것은 아니다. storage I/O 때문에 `D` 상태가 많아도 load가 높아질 수 있다.
- `vmstat r`
    - 실행 중이거나 실행 가능한 task 수를 본다. CPU core 수보다 지속적으로 크면 CPU run queue 대기를 의심한다.
- `vmstat us/sy/id/wa/st`
    - `us`는 user space 실행, `sy`는 kernel 실행, `id`는 idle, `wa`는 I/O wait, `st`는 hypervisor가 vCPU에서 CPU 시간을 빼앗은 비율이다. `wa`는 CPU가 idle하면서 outstanding I/O를 기다린 시간이라, 모든 I/O 지연을 완전히 대표하지 않는다.
- `pidstat -u -t`
    - process나 thread별 CPU 사용을 본다. 전체 CPU는 높지 않아도 특정 thread가 hot loop를 돌 수 있다.
- `perf top`, `perf record`
    - CPU sample이 어느 함수에 몰리는지 본다. 권한과 overhead가 있으므로 1분 초기 대응보다 다음 단계에 가깝다.
- `ps -eo state,pid,ppid,comm,wchan:32`
    - task state와 kernel wait channel을 본다. `D` 상태의 `wchan`은 어떤 kernel path에서 기다리는지 실마리를 준다.

### CPU 병목의 흔한 실전 형태

- runnable queue가 길다.
    - core 수 대비 runnable task가 많고 context switch가 증가한다. thread pool이 너무 크거나, GC thread, DB backend, Kafka network thread, application worker가 동시에 CPU를 요구할 수 있다.
- system CPU가 높다.
    - syscall, network processing, page fault, lock contention, context switch, kernel copy 비용이 커졌을 수 있다.
- CPU는 낮은데 latency가 높다.
    - thread가 CPU를 기다리지 않고 DB lock, socket read, disk `fsync`, page reclaim, cgroup throttle에서 기다릴 가능성이 크다.
- cgroup CPU throttle이 있다.
    - Kubernetes나 container 환경에서는 host CPU가 남아도 container CPU quota 때문에 `nr_throttled`, `throttled_time`이 늘 수 있다. 애플리케이션은 "CPU가 부족한 것처럼" 느리지만 host `idle`은 남아 있을 수 있다.

## 메모리, page cache, reclaim, swap, OOM, cgroup

메모리 문제는 "free가 작다"로 판단하면 자주 틀린다. Linux는 남는 memory를 page cache로 적극 사용한다. 중요한 질문은 "정말 할당할 memory가 부족한가", "page cache가 유용한가", "reclaim이 애플리케이션을 늦추는가", "swap이 들어갔는가", "cgroup 제한에 걸렸는가"다.

### virtual memory와 page fault

프로세스는 virtual address를 본다. CPU와 MMU는 page table을 통해 virtual address를 physical frame으로 바꾼다. 프로세스가 아직 실제 physical memory에 매핑되지 않은 page에 접근하면 page fault가 난다. page fault는 항상 나쁜 것이 아니다. 파일을 `mmap`한 뒤 처음 접근할 때 page cache에서 mapping을 만들거나, copy-on-write page를 만들 때도 page fault가 발생한다.

하지만 major page fault는 disk I/O가 필요할 수 있어 latency를 만든다. DB나 JVM에서 working set이 memory보다 커지고 page cache miss가 많아지면 major fault와 I/O가 늘 수 있다.

### page cache와 dirty page

파일 read는 보통 page cache를 먼저 본다. cache hit이면 disk까지 가지 않는다. write도 많은 경우 page cache에 먼저 dirty page를 만들고 나중에 writeback으로 장치에 내려간다.

```text
write(fd, user_buffer, len)
  -> kernel copies data into page cache page
  -> page becomes dirty
  -> write() may return before device persistence
  -> background writeback flushes dirty pages
  -> fsync()/fdatasync() waits for relevant dirty data and metadata
```

이 차이를 모르면 DB commit 지연을 설명할 수 없다. 일반 `write()`가 빠르다고 데이터가 장치에 안전하게 내려간 것은 아니다. DB가 transaction commit에서 WAL `fsync`를 기다리는 이유는 crash 후 복구 가능한 순서를 만들기 위해서다.

### reclaim

memory pressure가 생기면 kernel은 page cache나 anonymous page를 회수하려 한다. file-backed page cache는 깨끗하면 버리기 쉽고, dirty면 먼저 writeback이 필요하다. anonymous page는 swap이 있으면 swap out될 수 있고, 없으면 reclaim이 어려워진다.

`kswapd`는 background reclaim을 담당하고, 압박이 심하면 애플리케이션 thread가 직접 reclaim에 참여하는 direct reclaim이 생긴다. direct reclaim은 요청 처리 thread가 memory를 얻기 위해 reclaim 작업을 하느라 멈추는 것이므로 tail latency를 크게 키울 수 있다.

### swap

swap은 memory 부족 상황에서 anonymous page를 disk로 내보내는 장치다. swap이 있다는 것은 즉시 장애라는 뜻은 아니지만, swap in/out이 지속되면 애플리케이션 latency는 크게 흔들린다. 특히 JVM heap 일부가 swap out되면 GC나 application access 때 큰 지연이 생길 수 있다. DB buffer pool이 swap에 밀려도 치명적이다. 운영 DB에서는 swap을 완전히 끄는지, 낮은 `swappiness`로 두는지, OOM 정책을 어떻게 잡는지 환경별로 명확히 정해야 한다.

### OOM과 cgroup

시스템 전체 memory가 부족하면 kernel OOM killer가 process를 죽일 수 있다. container 환경에서는 host 전체 memory가 남아 있어도 해당 cgroup memory limit에 걸리면 cgroup OOM이 발생할 수 있다. Kubernetes에서 pod가 OOMKilled 되었는데 node memory가 남아 있는 장면은 흔하다.

cgroup v2 기준으로는 `memory.current`, `memory.max`, `memory.events`, `memory.stat` 같은 파일이 중요하다. `memory.events`의 `oom`, `oom_kill`은 제한 안에서 OOM이 있었는지 알려 준다. `memory.stat`의 file/anon, workingset, pgscan/pgsteal 계열은 page cache와 reclaim 압박을 읽는 데 도움이 된다.

### 메모리 지표를 읽는 방법

- `free -h`
    - `available`을 먼저 본다. `free`가 작아도 page cache가 회수 가능하면 정상일 수 있다.
- `/proc/meminfo`
    - `MemAvailable`, `Cached`, `Buffers`, `Dirty`, `Writeback`, `SwapFree`, `AnonPages`, `Slab`을 본다.
- `vmstat 1`
    - `si/so`는 swap in/out, `free`, `buff`, `cache`, `wa`, `r`, `b`를 함께 본다.
- `sar -B`, `/proc/vmstat`
    - page fault, pgscan, pgsteal, allocstall 같은 reclaim 신호를 본다.
- `dmesg -T`
    - OOM kill, filesystem error, device reset, memory allocation failure를 확인한다.
- cgroup files
    - container별 memory limit, current usage, OOM event, reclaim pressure를 본다.

## 파일 시스템, block I/O, fsync, writeback, device queue

파일 I/O는 애플리케이션이 파일 descriptor에 읽고 쓰는 일처럼 보이지만, 내부에서는 VFS, 파일 시스템, page cache, block layer, device driver, storage controller를 지난다.

### read path

```text
read(fd)
  -> VFS resolves file object and offset
  -> page cache lookup
  -> cache hit: copy data to user buffer
  -> cache miss: filesystem maps file offset to block
  -> block layer submits bio/request
  -> device queue and controller
  -> DMA fills memory
  -> completion interrupt
  -> task wakes and data is copied/visible
```

read latency가 높을 때는 page cache miss, storage queue, remote filesystem, device reset, filesystem lock, cgroup I/O throttle을 모두 생각해야 한다.

### write path와 fsync

```text
write(fd)
  -> copy user data into page cache
  -> mark page dirty
  -> return may happen before device write

fsync(fd)
  -> find dirty data and required metadata
  -> submit writeback to filesystem/block layer
  -> wait for completion and required flush
  -> return after durability boundary accepted
```

DB에서 WAL은 transaction commit 순서를 보장하는 핵심이다. WAL write가 page cache에만 있고 storage에 내려가지 않았는데 commit을 성공으로 말하면 crash recovery가 깨진다. 그래서 DB는 commit 시점에 `fsync`, group commit, `fdatasync`, storage flush 정책을 조합한다.

### writeback과 checkpoint

Linux는 dirty page를 계속 쌓지 않는다. background writeback은 dirty page를 장치에 내려보낸다. dirty page가 너무 많으면 foreground write가 throttle될 수 있다. DB checkpoint는 DB 내부 dirty page를 data file로 내리는 작업이고, OS writeback은 kernel dirty page를 storage로 내리는 작업이다. 둘은 다르지만 서로 겹친다. DB checkpoint가 많은 data file write를 만들고, 그 write가 OS page cache dirty page와 block queue를 채우면 WAL `fsync` 같은 foreground I/O도 늦어질 수 있다.

### block layer와 device queue

block layer는 filesystem 요청을 device가 처리할 block I/O request로 만든다. scheduler와 queue는 request를 병합하거나 순서를 조정할 수 있다. NVMe 같은 장치는 여러 hardware queue를 갖고 높은 병렬성을 처리한다. 하지만 queue depth가 커지면 throughput은 높아질 수 있어도 개별 latency는 커질 수 있다.

### I/O 지표를 읽는 방법

- `iostat -xz 1`
    - `r/s`, `w/s`는 초당 read/write 요청 수다.
    - `rkB/s`, `wkB/s`는 throughput이다.
    - `await`는 요청이 완료되기까지의 평균 시간이다. queue 대기와 service 시간이 함께 섞인다.
    - `r_await`, `w_await`가 있으면 read/write를 나눠 본다.
    - `aqu-sz`는 평균 queue 크기에 가깝다.
    - `%util`은 장치가 바빴던 시간 비율이다. HDD에서는 포화 신호로 강하지만, SSD/NVMe에서는 병렬성과 함께 해석해야 한다.
- `/proc/diskstats`
    - 장치별 누적 sector, I/O count, time, queue time을 제공한다. `iostat` 같은 도구의 원천 중 하나다.
- `vmstat b/wa`
    - `b`는 uninterruptible sleep task 수, `wa`는 CPU가 I/O wait로 idle한 비율이다. 둘 다 I/O 실마리지만 단독 결론은 아니다.
- `dmesg -T`
    - I/O error, device reset, filesystem journal error, path failover 같은 결정적 evidence가 나올 수 있다.
- `iotop`, `pidstat -d`
    - process별 I/O 경향을 본다. `iotop`은 권한이 필요할 수 있다.

## 네트워크, NIC, TCP, socket buffer, backlog, retransmit

네트워크는 애플리케이션에서는 `read`, `write`, `connect`, `accept`처럼 보이지만, 내부에는 NIC queue, interrupt/polling, kernel TCP state machine, socket buffer, backlog, routing, firewall, conntrack, TLS, application thread가 있다.

### receive path

```text
packet arrives at NIC
  -> NIC writes packet to memory by DMA
  -> interrupt or NAPI polling
  -> driver hands packet to network stack
  -> IP/TCP validation and state update
  -> payload enters socket receive buffer
  -> epoll/select wakes application
  -> application reads from socket
```

애플리케이션이 socket을 늦게 읽으면 receive buffer가 차고 TCP window가 줄어든다. 상대는 전송 속도를 낮추거나 멈춘다. 이 경우 네트워크가 느린 것처럼 보이지만, 원인은 애플리케이션 thread pool, GC pause, lock, CPU quota일 수 있다.

### send path

```text
application write()
  -> copy or reference data into socket send buffer
  -> TCP segmentation and congestion control
  -> qdisc / NIC transmit queue
  -> NIC sends frames
  -> ACKs arrive
  -> send buffer space is freed
```

send buffer가 차면 `write()`가 block되거나 non-blocking socket에서는 `EAGAIN`이 날 수 있다. 원인은 상대 receive가 느린 것, 네트워크 손실, congestion window 축소, local NIC queue, kernel send buffer 제한 등이다.

### listen backlog와 accept

서버 socket은 새 연결을 처리할 때 SYN backlog와 accept queue를 사용한다. SYN flood, application accept 지연, thread pool 포화, TLS handshake 지연이 있으면 새 연결이 밀릴 수 있다. `ss -ltn`에서 listen socket의 queue를 보고, kernel log나 TCP stats에서 overflow/drop을 확인한다.

### retransmit

TCP retransmit은 보낸 segment에 대한 ACK를 제때 받지 못해 다시 보내는 것이다. retransmit은 단순 "네트워크가 나쁘다"가 아니라 packet loss, congestion, MTU 문제, 방화벽, NIC/driver drop, receiver overload, asymmetric routing, middlebox 문제를 모두 열어 둔다. retransmit이 늘면 tail latency가 커지고, DB replication, Kafka produce/fetch, HTTP 요청이 모두 흔들릴 수 있다.

### 네트워크 지표를 읽는 방법

- `ss -s`
    - socket 상태 개요를 본다. established, time-wait, orphaned, retransmission 관련 단서를 빠르게 얻는다.
- `ss -tin`
    - TCP connection별 send/receive queue, retransmit, rtt, cwnd 같은 세부를 본다.
- `ss -ltn`
    - listen socket의 queue/backlog 상태를 본다.
- `ip -s link`
    - NIC 수준 RX/TX packet, error, drop을 본다.
- `/proc/net/snmp`, `/proc/net/netstat`
    - TCP retransmit, reset, listen overflow, receive error 같은 누적 counter를 본다.
- `dmesg -T`
    - NIC driver reset, firmware error, link flap, MTU 문제 같은 낮은 레이어 로그가 나올 수 있다.
- `tcpdump`
    - packet 증거를 직접 본다. 다만 권한, overhead, 개인정보 노출 위험이 있어 escalation 단계에서 범위를 좁혀 사용한다.

## 명령과 지표를 커널 객체로 해석하기

명령은 목적 없이 나열하면 금방 잊힌다. 아래 표는 "무엇을 보는가"보다 "어떤 커널 객체와 대기 상태를 비추는가"에 맞춘다.

| 명령/지표 | 비추는 상태 | 커널 객체/누적값 | 실무 해석 |
| --- | --- | --- | --- |
| `uptime`, load average | runnable + uninterruptible task 압력 | scheduler load accounting | CPU 병목인지 I/O wait인지 분리 전까지는 결론이 아니다. |
| `vmstat -w 1`의 `r` | CPU를 받을 수 있는 task 수 | run queue 근처의 scheduler 상태 | core 수보다 지속적으로 크면 CPU 대기 가능성이 있다. |
| `vmstat -w 1`의 `b` | uninterruptible sleep task | 주로 I/O 완료 wait task | storage, filesystem, remote FS, kernel wait를 의심한다. |
| `vmstat`의 `si/so` | swap in/out | VM subsystem counter | 지속되면 memory pressure와 latency 흔들림을 의심한다. |
| `free -h`의 `available` | 회수 가능성을 고려한 여유 memory | `/proc/meminfo` 계산 | `free`보다 먼저 본다. page cache 사용 자체는 문제 아니다. |
| `/proc/meminfo`의 `Dirty/Writeback` | 아직 장치에 내려가지 않은 page | page cache dirty/writeback accounting | writeback backlog, fsync 지연, checkpoint 영향과 연결한다. |
| `iostat -xz`의 `await` | block I/O 완료 평균 시간 | `/proc/diskstats` 기반 누적 시간 | queue 대기와 device service가 섞인다. read/write 분리 필요. |
| `iostat -xz`의 `aqu-sz` | 평균 block queue 깊이 | diskstats queue time | IOPS와 latency가 만나 queue가 커졌는지 본다. |
| `pidstat -u -d -r` | process별 CPU/I/O/memory fault | scheduler, block I/O, VM counter | 전체 지표를 어느 process가 만들었는지 좁힌다. |
| `ps ... state,wchan` | task state와 kernel wait 지점 | task state, wait channel | `D`와 `wchan`은 낮은 레이어 wait를 좁히는 강한 단서다. |
| `ss -tin` | TCP connection 세부 상태 | socket, TCP control block | send/recv queue, retransmit, RTT, cwnd를 본다. |
| `ip -s link` | NIC packet/error/drop | network device stats | host 내부 TCP 문제와 NIC/driver/link 문제를 나눈다. |
| `dmesg -T` | kernel event log | kernel ring buffer | OOM, device reset, filesystem error, driver issue는 결정적 증거가 될 수 있다. |
| cgroup `cpu.stat` | quota throttle | cgroup CPU controller | host CPU가 남아도 container가 느릴 수 있다. |
| cgroup `memory.events` | cgroup OOM/reclaim event | cgroup memory controller | pod/container 단위 memory 제한 문제를 확인한다. |

이 표를 실제 장애에서 쓰는 방식은 간단하다. 먼저 전체 압력을 보고, 그 다음 process별로 좁히고, 마지막으로 kernel log와 application/DB 지표로 같은 방향의 증거를 맞춘다. 하나의 지표만으로 결론을 내리지 않는다.

## DB는 OS 위에서 가장 강한 애플리케이션이다

DB는 OS 관점에서 보면 process다. 하지만 일반 애플리케이션보다 OS와 hardware 자원을 훨씬 직접적이고 지속적으로 사용한다. DB는 자체 buffer pool, lock manager, transaction log, checkpoint, background writer, replication, compaction을 가진다. 이 내부 구조가 CPU, memory, disk, network 지표를 만든다.

### Lock

DB lock은 logical data consistency를 지키기 위해 transaction 사이의 접근 순서를 제어한다. lock wait가 생기면 DB session은 CPU를 계속 태우는 것이 아니라 lock holder가 끝나기를 기다린다. 애플리케이션에서는 query latency 증가, connection pool 고갈, thread pool 점유로 보인다. OS에서는 CPU가 낮고 socket은 열려 있고 DB process는 sleeping처럼 보일 수 있다.

실무에서는 lock wait를 볼 때 holder를 같이 봐야 한다. holder가 CPU를 태우는지, WAL `fsync`를 기다리는지, 외부 API를 호출한 채 transaction을 열어 둔 것인지에 따라 해결책이 완전히 다르다.

### MVCC

MVCC는 reader와 writer가 서로 덜 막히도록 version을 유지하는 방식이다. PostgreSQL의 dead tuple, InnoDB undo log, Cassandra tombstone처럼 엔진마다 형태는 다르지만, 공통 원리는 "현재 transaction이 볼 수 있는 version"을 판단한다는 것이다. MVCC는 lock을 줄여 주지만 공짜가 아니다. 오래 열린 transaction은 정리할 수 없는 version을 붙잡고, vacuum이나 purge가 밀리며, table/index bloat, page cache pressure, disk I/O 증가로 이어진다.

### Buffer pool과 page cache

DB buffer pool은 DB page를 담는 cache다. OS page cache와 다르다. DB가 buffer pool miss를 내면 data file read를 요청하고, 이 요청은 OS page cache hit이면 storage까지 가지 않을 수 있다. 하지만 DB가 direct I/O를 쓰면 OS page cache를 우회한다. 어떤 DB는 WAL은 buffered I/O, data file은 direct I/O처럼 다르게 설정하기도 한다.

따라서 DB memory 문제를 볼 때는 OS `free`만 보지 말고 DB buffer hit ratio, dirty page, checkpoint age, eviction, page read/write, cgroup memory limit을 같이 본다.

### WAL, fsync, group commit

WAL은 data page를 직접 바꾸기 전에 변경 의도를 log로 남기는 구조다. commit이 성공하려면 crash recovery가 그 transaction을 재현하거나 되돌릴 수 있어야 한다. 그래서 DB는 WAL을 순서 있게 쓰고, commit 시점에 durability boundary를 닫는다.

```text
transaction updates row
  -> DB changes page in buffer pool
  -> DB appends WAL record
  -> commit waits for WAL flush/fsync
  -> response can be returned
  -> data page write may happen later by checkpoint/background writer
```

WAL `fsync` 지연은 commit latency로 바로 보인다. storage latency가 튀거나, checkpoint가 data file write를 밀어 넣거나, filesystem journal이 바쁘거나, controller flush가 느리면 commit p99가 오른다. group commit은 여러 transaction의 WAL flush를 묶어 throughput을 높일 수 있지만, flush latency 자체가 나빠지면 tail latency는 여전히 흔들린다.

### Checkpoint

Checkpoint는 DB가 "여기까지는 data file에 반영되어 있고, crash recovery는 이 지점 이후 WAL을 보면 된다"는 기준점을 만드는 작업이다. checkpoint가 너무 자주 또는 너무 크게 일어나면 background write가 몰리고, block queue와 writeback을 압박한다. 반대로 checkpoint가 너무 늦으면 recovery 시간이 길어진다.

### Compaction

LSM 기반 저장소나 Cassandra 같은 시스템은 memtable을 SSTable로 flush하고, 여러 SSTable을 compaction으로 합친다. Compaction은 read amplification을 줄이고 tombstone을 정리하지만, CPU, disk read/write, page cache, network streaming까지 크게 쓴다. Foreground read/write와 같은 storage를 공유하므로, compaction이 몰리면 애플리케이션 latency가 올라간다.

### Replication

Replication은 network와 disk를 동시에 쓴다. primary는 WAL/binlog/commit log를 만들고, replica는 network로 받아 disk에 쓰고 적용한다. replication lag는 network 지연, replica apply CPU, disk `fsync`, lock conflict, large transaction, checkpoint/compaction과 연결된다. "lag가 늘었다"는 말은 원인이 아니라 증상이다.

## JVM, Spring, Kafka, Cassandra, Spark가 OS 자원을 쓰는 방식

### JVM

JVM은 OS process다. Java heap만 memory가 아니다. thread stack, metaspace, code cache, direct buffer, mmap file, native library, JIT compiler, GC metadata도 memory를 쓴다. container에서 `-Xmx`만 보고 memory limit을 맞추면 native memory 때문에 cgroup OOM이 날 수 있다.

JVM thread는 OS thread와 연결된다. thread가 많으면 context switch, stack memory, scheduler overhead가 늘어난다. GC는 CPU를 쓰고, stop-the-world pause 동안 애플리케이션 thread가 멈출 수 있다. 이때 socket receive buffer가 차고, DB connection을 오래 잡고, Kafka consumer poll이 늦어지고, upstream timeout이 생길 수 있다.

### Spring

Spring MVC/Tomcat 같은 thread-per-request 모델은 요청마다 worker thread가 DB, Redis, 외부 API, file I/O를 기다릴 수 있다. thread pool이 모두 blocking wait에 묶이면 CPU는 낮아도 새 요청이 기다린다. WebFlux나 Reactor 기반 모델은 적은 event loop thread로 많은 connection을 다루지만, event loop에서 blocking call을 실행하면 전체 connection 처리 흐름이 막힌다.

따라서 Spring 장애에서는 CPU뿐 아니라 thread pool active/queue, DB connection pool active/wait, HTTP client pool, GC pause, socket 상태, cgroup throttle을 같이 본다.

### Kafka

Kafka broker는 page cache를 적극 활용한다. log segment는 file이고, produce는 append write, consume은 sequential read에 가깝다. OS page cache hit가 높으면 consumer fetch가 빠르다. Broker는 network thread, I/O thread, request queue, page cache, disk flush, replication을 함께 쓴다.

Consumer lag가 늘었다고 곧바로 Kafka broker 문제는 아니다. Consumer 애플리케이션 CPU, GC, DB sink latency, network retransmit, partition assignment, rebalance, broker disk I/O가 모두 원인일 수 있다. Broker에서 under-replicated partition, request queue time, produce/fetch latency, disk await, page cache pressure를 같이 봐야 한다.

### Cassandra

Cassandra는 commit log, memtable, SSTable, compaction이 핵심이다. Write는 commit log append와 memtable update로 시작하고, memtable flush가 SSTable을 만든다. Read는 memtable, cache, bloom filter, index, SSTable을 지나고, 여러 SSTable을 많이 봐야 하면 read amplification이 생긴다. Tombstone이 많으면 read와 compaction이 비싸진다.

Cassandra 장애는 JVM GC, heap/off-heap, compaction backlog, pending task, disk I/O, network streaming, repair, hinted handoff가 OS 지표와 겹친다. `iostat await`가 높고 compaction이 몰리면 read/write p99가 함께 올라갈 수 있다.

### Spark

Spark executor도 JVM process다. Task는 CPU를 쓰고, shuffle은 disk와 network를 크게 쓴다. Executor memory는 heap, off-heap, storage memory, execution memory, Python worker memory까지 나뉠 수 있다. Shuffle spill이 늘면 local disk I/O가 증가하고, remote fetch가 늘면 network와 TCP retransmit이 중요해진다. Container memory overhead를 작게 잡으면 JVM heap은 남아도 native/off-heap 때문에 pod가 죽을 수 있다.

## 장애 대응 학습 구조

### 1초 사고 지도

장애 알람을 보자마자 외울 문장은 이것이다.

```text
요청은 지금 CPU를 기다리는가,
memory를 얻거나 회수하느라 기다리는가,
storage flush/read를 기다리는가,
network ACK/receive/send를 기다리는가,
DB lock/WAL/checkpoint/compaction을 기다리는가,
아니면 cgroup이나 platform limit에 막혔는가?
```

이 문장은 추상적 분류가 아니라 queue 후보 목록이다.

- CPU run queue
- scheduler throttle queue
- page reclaim/swap/OOM
- page cache dirty/writeback
- block device queue
- socket receive/send buffer
- listen backlog
- TCP retransmit/congestion window
- DB lock wait queue
- DB WAL flush queue
- DB checkpoint/compaction backlog
- JVM thread pool/GC safepoint
- application connection pool queue

### 1분 read-only 명령 세트

아래 명령은 초기 1분에 "파괴적이지 않은 관측"으로 큰 방향을 잡기 위한 것이다. 환경에 따라 일부 명령은 설치되어 있지 않거나 권한이 필요할 수 있다. 권한이 필요한 명령은 무리해서 실행하지 말고 가능한 대체 지표를 먼저 본다.

```sh
date
hostname
uptime
vmstat -w 1 5
free -h
cat /proc/meminfo | egrep 'MemAvailable|Cached|Dirty|Writeback|Swap|AnonPages|Slab'
iostat -xz 1 5
pidstat -urd -h 1 5
ps -eo state,pid,ppid,comm,wchan:32 --sort=state | head -80
ss -s
ss -tin | head -80
ss -ltn
ip -s link
dmesg -T | tail -120
```

Container나 Kubernetes 환경이면 cgroup도 같이 본다. 경로는 cgroup v1/v2와 runtime에 따라 다르다.

```sh
cat /sys/fs/cgroup/cpu.stat 2>/dev/null
cat /sys/fs/cgroup/memory.current 2>/dev/null
cat /sys/fs/cgroup/memory.max 2>/dev/null
cat /sys/fs/cgroup/memory.events 2>/dev/null
cat /sys/fs/cgroup/memory.stat 2>/dev/null | egrep 'anon|file|slab|workingset|pgscan|pgsteal|oom'
```

DB와 JVM은 process 내부 지표가 필요하다. 초기 1분에는 read-only로 현재 상태만 본다.

```sh
jcmd <pid> Thread.print
jcmd <pid> GC.heap_info
jcmd <pid> VM.native_memory summary
```

DB별로는 lock wait, active query, checkpoint/WAL, replication lag, buffer/cache 지표를 본다. SQL은 운영 DB에서 부하가 낮은 조회만 사용해야 한다.

### 10분 원인 분기

1분 관측으로 방향을 잡았으면 10분 안에 "증상 -> 후보 queue -> 확인 evidence -> 임시 완화 또는 escalation"으로 좁힌다.

#### CPU 쪽으로 갈 때

- `vmstat r`이 core 수보다 지속적으로 크고 `us/sy`가 높다.
- `pidstat -u -t`에서 특정 process/thread가 CPU를 태운다.
- Java면 thread dump에서 hot thread가 CPU 작업인지, lock spin인지, GC인지 본다.
- `perf top`이나 async-profiler 같은 profiler는 권한과 overhead를 판단한 뒤 짧게 쓴다.
- cgroup `cpu.stat`의 throttle이 늘면 quota 문제를 platform evidence로 모은다.

#### Memory 쪽으로 갈 때

- `MemAvailable`이 낮고 swap in/out, major fault, reclaim counter가 증가한다.
- `Dirty/Writeback`이 높으면 writeback과 storage를 같이 본다.
- cgroup `memory.events`에 OOM이나 high event가 있다.
- JVM이면 heap, native memory, direct buffer, thread count, GC pause를 같이 본다.
- DB면 buffer pool, shared buffer, temp file, sort/hash spill, connection 수를 본다.

#### Storage 쪽으로 갈 때

- `iostat await`, `aqu-sz`, write await가 올라간다.
- `vmstat b`가 늘거나 process `D` 상태가 보인다.
- `dmesg`에 device reset, I/O error, filesystem error가 있다.
- DB WAL fsync, checkpoint, compaction, backup, snapshot, log rotation, replication apply와 시간대를 맞춘다.
- Storage controller, SAN/NAS, hypervisor, noisy neighbor 가능성은 host evidence로 escalation한다.

#### Network 쪽으로 갈 때

- `ss -tin`에서 retransmit, 높은 send/receive queue, RTT 증가가 보인다.
- `ip -s link`에서 drop/error가 늘어난다.
- listen queue가 차거나 connection reset/time-wait가 급증한다.
- 애플리케이션 thread가 socket read/write에 묶였는지 thread dump와 같이 본다.
- 특정 peer만 문제인지, node 전체인지, AZ/IDC/network segment 문제인지 분리한다.

#### DB/application 쪽으로 갈 때

- CPU, memory, storage, network의 host 지표가 애매한데 application p99가 크다.
- DB lock wait, connection pool wait, slow query, transaction age, replication lag, checkpoint time, compaction backlog를 본다.
- JVM GC pause, thread pool queue, event loop blocking, HTTP client pool, Kafka consumer lag를 본다.
- "DB가 느리다"와 "DB를 기다리는 애플리케이션 thread가 많다"를 분리한다. 둘은 다르다.

### Escalation evidence

인프라 팀, DBA, 네트워크 팀, 플랫폼 팀에 escalation할 때는 "느립니다"가 아니라 재현 가능한 evidence 묶음을 준다.

- 공통
    - 장애 시작/종료 시각, 영향 서비스, p50/p95/p99, RPS, error rate, 변경/배포/스케일 이벤트
    - 같은 시간대의 host, container, application, DB, network 지표
- CPU
    - load, `vmstat r`, process/thread CPU, cgroup throttle, hot function 또는 thread dump
- Memory
    - `MemAvailable`, swap, reclaim, OOM log, cgroup memory event, JVM heap/native memory, DB memory 지표
- Storage
    - `iostat -xz`, diskstats 변화, `dmesg`, DB WAL/checkpoint/compaction 시간대, device/path 정보
- Network
    - `ss -tin`, retransmit counter, NIC drop/error, peer별 영향 범위, packet capture가 있다면 제한된 범위의 요약
- DB
    - lock holder/waiter, query id, transaction age, WAL/checkpoint, replication lag, buffer hit/miss, compaction/vacuum 상태

## 면접 답변 구조

### 30초 답변

장애 상황에서 Linux 서버를 본다면 먼저 요청이 어느 대기열에서 밀리는지 좁힙니다. CPU run queue인지, memory reclaim/swap인지, block I/O queue와 fsync인지, TCP socket buffer와 retransmit인지, DB lock/WAL/checkpoint인지 분리합니다. 그래서 `vmstat`, `iostat`, `pidstat`, `ss`, `dmesg`, cgroup 파일, DB/JVM 내부 지표를 같이 봅니다. 중요한 것은 명령어를 많이 치는 것이 아니라, 각 지표가 어떤 커널 객체와 wait state를 반영하는지 연결해서 원인을 줄이는 것입니다.

### 2분 꼬리 답변

예를 들어 API p99가 갑자기 늘었는데 CPU 사용률이 낮다고 해서 서버가 여유롭다고 말하지 않습니다. `vmstat r`이 낮고 `b`가 높으면 CPU가 아니라 uninterruptible sleep, 주로 I/O wait 쪽을 봅니다. `iostat -xz`에서 `await`와 `aqu-sz`가 같이 오르면 block device queue가 밀릴 수 있고, DB commit latency가 같이 오르면 WAL `fsync`나 checkpoint 영향을 의심합니다. 반대로 `ss -tin`에서 retransmit과 send queue가 늘면 TCP ACK나 상대 receive 문제일 수 있습니다. JVM thread dump에서 많은 thread가 DB socket read에 있다면 DB lock이나 slow query, storage fsync까지 내려가 봅니다.

메모리는 `free`가 작다는 이유만으로 장애라고 하지 않습니다. Linux는 page cache를 쓰기 때문에 `MemAvailable`, dirty/writeback, swap in/out, reclaim counter, cgroup memory event를 봅니다. Container에서는 host memory가 남아도 cgroup OOM이 날 수 있습니다. DB는 OS 위의 process지만 buffer pool, WAL, MVCC, lock, checkpoint 같은 내부 구조 때문에 OS page cache, block I/O, CPU, network를 강하게 사용합니다. 그래서 애플리케이션, DB, OS, 하드웨어 지표를 같은 시간축으로 맞춰 원인을 좁힙니다.

### 명령/관측으로 검증하는 답변

면접에서 "어떻게 확인하겠습니까?"를 받으면 아래처럼 답한다.

- CPU 의심
    - `vmstat -w 1`, `pidstat -u -t 1`, cgroup `cpu.stat`을 보고 run queue, per-thread CPU, throttle을 확인한다.
    - PASS는 core 수 대비 `r`이 지속적으로 높고 특정 thread/process가 CPU를 태우는 증거가 같이 나오는 것이다.
    - FAIL은 CPU가 낮고 thread가 I/O나 lock wait에 묶인 증거가 나오는 것이다.
- Memory 의심
    - `free -h`, `/proc/meminfo`, `vmstat si/so`, cgroup `memory.events`, JVM heap/native memory를 본다.
    - PASS는 reclaim/swap/OOM/cgroup event가 latency와 같은 시간대에 증가하는 것이다.
    - FAIL은 memory pressure가 없고 다른 queue에서 병목 증거가 나오는 것이다.
- Storage 의심
    - `iostat -xz 1`, `vmstat b/wa`, `ps state,wchan`, `dmesg`, DB WAL/checkpoint 지표를 본다.
    - PASS는 `await/aqu-sz`, `D` state, DB fsync/checkpoint 지연이 같은 시간대에 맞는 것이다.
    - FAIL은 storage 지표가 안정적이고 lock/network/application queue 증거가 강한 것이다.
- Network 의심
    - `ss -tin`, `ss -s`, `ip -s link`, TCP stats, application socket wait를 본다.
    - PASS는 retransmit, send/receive queue, NIC drop/error, peer별 지연이 같은 방향으로 나오는 것이다.
    - FAIL은 TCP 상태가 안정적이고 DB lock이나 CPU/memory/storage evidence가 강한 것이다.

## 실전 분기 예시

### CPU는 낮은데 API p99가 높다

CPU가 낮으면 "할 일이 없다"가 아니라 "CPU 말고 다른 것을 기다린다"일 수 있다. 먼저 thread dump와 `pidstat`, `vmstat`를 같이 본다. Java thread가 DB socket read에 많으면 DB 응답을 기다린다. DB에서 lock wait가 있으면 holder를 찾는다. holder가 WAL `fsync`를 기다리면 storage로 내려간다. holder가 외부 API를 transaction 안에서 호출 중이면 application 설계 문제다.

### `iostat await`가 튄다

`await`는 block I/O 완료까지 걸린 평균 시간이다. 이것만으로 device가 고장이라고 말하지 않는다. `aqu-sz`, read/write 분리, throughput, `%util`, process별 I/O, DB checkpoint/WAL/compaction, `dmesg`를 함께 본다. 같은 시간대에 DB commit latency가 오르면 WAL flush를 의심한다. Backup이나 snapshot이 있었다면 read/write 경쟁을 본다. Hypervisor나 SAN 환경이면 storage path evidence를 모아 escalation한다.

### load average가 높다

load가 높을 때 `vmstat r`이 높으면 CPU runnable queue를 본다. `b`가 높고 process `D`가 많으면 I/O wait나 kernel wait를 본다. Linux load에는 uninterruptible sleep이 포함되므로, load만 보고 CPU scale-out을 결정하면 틀릴 수 있다.

### Kafka consumer lag가 늘었다

Consumer lag는 Kafka broker 문제일 수도 있고 consumer 애플리케이션 문제일 수도 있다. Consumer process CPU/GC/thread pool/DB sink latency를 먼저 본다. Broker 쪽에서는 produce/fetch latency, request queue, under-replicated partition, disk await, page cache pressure, network retransmit을 본다. Lag는 queue 길이이고, 원인은 생산 증가, 소비 감소, broker/network/storage 지연 중 하나다.

### Cassandra p99가 튄다

Cassandra는 compaction, GC, disk I/O, network streaming이 p99를 크게 흔든다. Pending compaction, tombstone, SSTable 수, read repair, heap/off-heap, GC pause, `iostat await`, NIC retransmit을 같은 시간대에 본다. Compaction은 장기적으로 필요하지만 foreground 요청과 같은 자원을 쓰므로 throttle 정책과 window를 봐야 한다.

### Spark job이 느려졌다

Spark는 CPU 계산만 하는 시스템이 아니다. Shuffle spill이 늘면 local disk가 바빠지고, remote fetch가 많으면 network와 TCP가 중요해진다. Executor OOM은 heap 부족뿐 아니라 off-heap, direct memory, Python worker, container overhead 부족으로 날 수 있다. Stage별 task time, GC time, spill size, shuffle read/write, executor lost reason, cgroup memory event를 같이 본다.

## 기존 문서 owner와 drift map

이 파일은 링크 허브가 아니라 완결형 통합 교본이다. 다만 기존 본진 문서가 이미 깊게 다루는 주제를 모두 여기서 계속 최신화하면 중복과 drift가 생긴다. 따라서 이 파일의 역할과 각 본진 문서의 역할을 나눈다.

| 주제 | 이 파일의 역할 | 본진 owner | 중복 허용 범위 | drift 방지 규칙 |
| --- | --- | --- | --- | --- |
| OS/컴퓨터 구조 전체 map | 하드웨어 -> 커널 -> 애플리케이션 연결을 한 번에 설명 | `os-kernel-computer-architecture.md` | 핵심 mental model과 면접 답변은 중복 허용 | root curriculum이 바뀌면 이 파일의 짧은 직답과 사고 지도를 맞춘다. |
| scheduler/process/thread | 장애 분기와 커널 wait 관점으로 요약 | `os-kernel-distributed-systems-deep-dive/01a_process_scheduling.md` | run queue, process state, Java thread caveat는 중복 허용 | 세부 CFS, priority, scheduling class 설명은 본진에서 보강한다. |
| page cache/filesystem/block I/O | DB WAL/fsync와 장애 지표까지 연결 | `os-kernel-distributed-systems-deep-dive/01c_filesystem_page_cache_block_io.md` | page cache, dirty/writeback, fsync, `iostat` 해석은 중복 허용 | filesystem별 세부와 kernel version 차이는 본진과 source ledger를 따른다. |
| cgroup/observability/eBPF | cgroup throttle/OOM과 read-only command 중심 | `os-kernel-distributed-systems-deep-dive/01e_concurrency_isolation_observability.md` | cgroup CPU/memory event는 중복 허용 | 이 파일은 관측 철학이 아니라 내부 구조와 사고 분기를 중심으로 유지한다. |
| DB storage/buffer/WAL | DB를 OS application으로 연결 | `database-deep-dive/02-storage-pages-buffer-io.md` | WAL, buffer pool, checkpoint 설명은 중복 허용 | DB 엔진별 세부 차이는 DB deep-dive에서 확장한다. |
| DB troubleshooting | lock/MVCC/replication/운영 분기 연결 | `database-deep-dive/13-operations-security-troubleshooting.md` | escalation evidence와 장애 분기는 중복 허용 | SQL 예제와 엔진별 운영 절차는 본진에서 관리한다. |
| Kafka | page cache, network, lag를 OS 관점으로 연결 | `database-deep-dive/03_kafka_deep_dive.md` | Kafka lag와 OS 자원 연결은 중복 허용 | broker 내부 설정과 protocol 세부는 본진에서 관리한다. |
| Cassandra | commit log/memtable/SSTable/compaction을 OS 지표로 연결 | `database-deep-dive/04_cassandra_deep_dive.md` | compaction과 disk/CPU pressure 설명은 중복 허용 | Cassandra tuning과 repair 세부는 본진에서 관리한다. |
| Spark | executor/shuffle/off-heap/cgroup 연결 | `database-deep-dive/05_spark_deep_dive.md` | shuffle disk/network와 container memory 설명은 중복 허용 | Spark UI 지표와 cluster manager별 차이는 본진에서 관리한다. |
| VMware/metric scripts | Linux I/O와 VM host 환경 사례로 일반화 | `linux/commands/metrics/*` | `iostat`, diskstats, dirty/writeback 해석만 일반화해 중복 허용 | VMware 자체 절차나 특정 script 사용법은 이 파일의 중심이 아니다. |

후속 보강 목록은 아래처럼 둔다. 이번 단계에서는 기존 본진 문서를 대규모 개편하지 않는다.

- OS deep-dive에는 cgroup v2 파일 해석, PSI, NUMA, block scheduler/NVMe queue depth를 더 깊게 보강한다.
- DB deep-dive에는 PostgreSQL/MySQL/Cassandra/Kafka별 `metric -> OS state -> user symptom` 표를 엔진별로 추가한다.
- JVM/Spring 문서에는 thread dump state와 OS process state가 어긋나는 사례, direct buffer/native memory/cgroup OOM 사례를 보강한다.
- 운영 실천 문서에는 1분 read-only command set을 실제 incident report template과 연결한다.
- Linux metric script 문서는 VMware 중심 문구를 줄이고, Linux kernel counter와 queue 해석을 먼저 설명하도록 정리한다.

## 반복 학습 질문

문서를 읽은 뒤 아래 질문에 답할 수 있어야 한다.

- Linux load average가 높을 때 왜 CPU 병목이라고 바로 말하면 안 되는가?
- Java thread dump의 `RUNNABLE`과 Linux task state `R`은 왜 같은 말이 아닌가?
- `write()`가 성공했는데도 DB가 `fsync`를 기다리는 이유는 무엇인가?
- page cache와 DB buffer pool은 무엇이 같고 무엇이 다른가?
- checkpoint가 foreground query latency를 높일 수 있는 경로를 설명해 보라.
- `iostat await`와 `aqu-sz`가 같이 오를 때 어떤 queue를 의심해야 하는가?
- `free`가 낮고 `available`이 충분한 서버와, `available`이 낮고 swap in/out이 있는 서버의 차이는 무엇인가?
- TCP retransmit이 늘 때 애플리케이션, 커널, 네트워크 장비 중 어디를 어떤 evidence로 나눌 수 있는가?
- Kafka consumer lag를 생산 증가, 소비 감소, broker/storage/network 문제로 어떻게 분리할 것인가?
- Container에서 host CPU와 memory가 남아 있는데도 애플리케이션이 느리거나 죽을 수 있는 이유는 무엇인가?

이 질문에 답할 때는 항상 `증상 -> queue/cache/buffer/wait -> 커널 객체/카운터 -> 명령/지표 -> 애플리케이션 또는 DB 내부 상태 -> 완화/확인` 순서로 말한다. 이 순서가 몸에 들어오면, 면접 답변과 장애 대응이 같은 사고에서 나온다.
