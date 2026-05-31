# 01. OS Kernel Foundations

> OS 커널은 애플리케이션이 CPU, 메모리, 디스크, 네트워크를 직접 만지지 않게 막고, 제한된 자원을 안전하게 나눠 주는 계층입니다.
> 백엔드 시스템의 지연, 처리량, 장애는 결국 runnable task, virtual memory, page cache, block I/O, socket buffer 중 어디에서 상태가 밀렸는지로 내려갑니다.
> Kafka의 page cache, Cassandra의 commit log와 compaction, Spark의 executor heap과 spill은 모두 커널이 관리하는 자원 위에 세워진 구조입니다.

## 1. 사용자 모드와 커널 모드

> 사용자 모드(user mode)는 애플리케이션 코드가 실행되는 제한된 영역이고, 커널 모드(kernel mode)는 하드웨어와 전역 자원에 접근할 수 있는 보호 영역입니다.
> 시스템 콜(system call)은 사용자 코드가 커널에 "대신 해 달라"고 요청하는 공식 통로입니다.
> 이 경계를 이해하면 파일 쓰기, 네트워크 송신, process 생성이 단순 함수 호출보다 비싼 이유를 설명할 수 있습니다.

### 질문

`write(fd, buf, len)` 같은 함수는 왜 그냥 라이브러리 함수처럼 실행되지 않고 커널을 거쳐야 할까요?

### 직관

여러 프로그램이 한 컴퓨터의 디스크와 네트워크 카드를 마음대로 만질 수 있다면 한 프로세스가 다른 프로세스의 파일을 덮거나, 메모리를 훔쳐보거나, 네트워크 장치를 독점할 수 있습니다. 커널은 건물의 관리실처럼 공용 설비 사용을 중재합니다.

### 작은 예시

프로세스 A가 `log.txt`에 100 byte를 쓰고, 프로세스 B가 같은 디스크에서 DB 파일을 읽습니다. 두 프로세스 모두 "내 코드가 디스크를 쓴다"고 느끼지만 실제 장치 명령은 커널과 드라이버가 순서를 정합니다.

### 상태 이동 trace

```
user process
  fd=3, buf="hello"
      |
      | write(3, buf, 5)
      v
syscall entry
      |
      | CPU mode switch, argument validation
      v
kernel VFS / filesystem / page cache
      |
      | dirty page, later writeback
      v
block layer / device driver / storage
```

이동하는 것은 `hello`라는 byte와 "fd 3에 쓰라"는 요청입니다. 상태가 바뀌는 지점은 사용자 메모리, 커널 page cache, 파일시스템 metadata, block device queue입니다.

### 내부 메커니즘

사용자 모드 코드는 일반 CPU 명령을 실행하지만, 장치 제어나 커널 자료구조 접근 같은 권한 있는 동작은 할 수 없습니다. 시스템 콜 명령이 실행되면 CPU는 정해진 진입점으로 이동하고 권한 수준을 바꿉니다. 커널은 fd가 유효한지, 사용자 주소가 읽을 수 있는지, 파일 권한이 맞는지 확인한 뒤 VFS(가상 파일시스템 계층)를 통해 실제 파일시스템으로 요청을 넘깁니다.

이 경계는 성능에도 영향을 줍니다. mode switch 자체, 인자 검증, 사용자 공간과 커널 공간 사이의 copy, scheduler 개입 가능성이 비용입니다. Kafka가 batch와 sendfile을 중시하는 이유도 이 비용을 줄이기 위해서입니다.

### 실패 모드

- 작은 `write()`를 너무 자주 호출하면 syscall 횟수와 lock 경쟁이 늘어납니다.
- 사용자 buffer가 잘못된 주소면 커널은 오류를 반환하거나 process에 signal을 보냅니다.
- `write()`가 성공해도 device에 영속화됐다는 뜻은 아닐 수 있습니다. 정상 파일 I/O는 page cache에 먼저 들어갈 수 있습니다.

### 검증 방법

Linux:

```bash
strace -e trace=write,fsync -o /tmp/syscalls.log sh -c 'printf hello > /tmp/os-demo.txt; sync'
sed -n '1,20p' /tmp/syscalls.log
```

macOS:

```bash
sudo dtruss -t write -o /tmp/syscalls.log sh -c 'printf hello > /tmp/os-demo.txt'
```

PASS 신호:

- `write` 계열 syscall이 관찰됩니다.
- `printf` 같은 쉘 명령도 결국 파일 쓰기 syscall로 내려갑니다.

FAIL 신호:

- 권한 문제로 tracing 자체가 실패합니다. macOS에서는 SIP(System Integrity Protection) 때문에 일부 프로세스 추적이 제한될 수 있습니다.

### 면접식 되묻기

면접관이 "`write()`가 성공했으면 디스크에 저장된 건가요?"라고 물으면 먼저 "아닐 수 있습니다"라고 짧게 답합니다. 그 다음 user buffer에서 kernel page cache로 이동한 것과 storage device에 flush된 것을 분리합니다. 이어서 `fsync()`나 데이터베이스의 commit log sync 정책, Kafka replication 같은 상위 내구성 모델로 연결합니다.

### 흔한 오해와 반례

오해: "시스템 콜은 느리니까 항상 피해야 한다."

반례: 시스템 콜을 완전히 피할 수는 없습니다. 핵심은 syscall 자체를 악으로 보는 것이 아니라, 작은 요청을 너무 잘게 쪼개지 않고 batch, buffer, zero-copy, async I/O처럼 경계 통과 비용을 줄이는 것입니다.

### Active recall

- 사용자 모드가 디스크 장치를 직접 제어하지 못하게 하는 이유는 무엇인가요?
- `write()` 성공과 durable write를 구분하는 trace를 그려 보세요.
- Kafka의 batch와 sendfile은 어떤 kernel boundary 비용을 줄이나요?

## 2. Interrupt, Trap, Syscall

> interrupt는 외부 장치가 CPU의 주의를 끄는 사건이고, trap은 현재 실행 흐름 안에서 발생한 예외나 의도적 진입입니다.
> syscall은 애플리케이션이 커널 기능을 요청하기 위해 의도적으로 일으키는 trap의 한 종류로 볼 수 있습니다.
> 커널은 이런 사건을 처리하면서 지금 실행 중이던 thread의 흐름을 잠시 멈추고, handler를 실행한 뒤 돌아올지 다른 task로 바꿀지 결정합니다.

### 질문

네트워크 packet이 도착하거나 page fault가 발생하면 CPU는 원래 실행하던 코드를 어떻게 멈추고 커널 코드를 실행할까요?

### 직관

interrupt는 초인종에 가깝습니다. 집 안에서 다른 일을 하고 있어도 초인종이 울리면 잠깐 문으로 갑니다. trap은 집 안에서 스스로 "관리실에 문의해야 한다"고 판단해 전화를 거는 상황에 가깝습니다.

### 작은 예시

웹 서버가 `epoll_wait()`로 socket event를 기다립니다. NIC가 packet을 받으면 interrupt가 발생하고, 커널 network stack은 socket receive buffer에 데이터를 놓습니다. `epoll_wait()`는 그 socket이 읽기 가능하다고 사용자 process를 깨웁니다.

### 상태 이동 trace

```
NIC receives packet
  |
  | interrupt
  v
kernel interrupt handler / softirq
  |
  | packet parsed, socket buffer updated
  v
waiting thread becomes runnable
  |
  | scheduler decision
  v
user process returns from epoll_wait()
```

### 내부 메커니즘

interrupt handler는 가능한 짧게 처리하고 나머지 작업을 softirq, tasklet, kernel thread 같은 하위 경로로 넘기는 경우가 많습니다. trap은 page fault처럼 CPU가 현재 명령을 실행하다가 더 이상 진행할 수 없을 때 생기거나, syscall처럼 의도적으로 커널 진입을 요청할 때 생깁니다.

이 구조 때문에 "CPU가 내 thread를 실행 중이었다"와 "내 thread가 계속 진행 가능하다"는 다릅니다. packet이 도착해도 scheduler가 해당 thread를 즉시 실행하지 않으면 애플리케이션 관점의 latency가 생깁니다.

### 실패 모드

- interrupt가 특정 CPU에 몰리면 softirq 처리와 user task 실행이 경쟁합니다.
- page fault가 많으면 애플리케이션 CPU 시간이 줄고 kernel time이 늘어납니다.
- syscall이 blocking I/O에 들어가면 thread가 runnable 상태에서 sleep 상태로 내려갈 수 있습니다.

### 검증 방법

Linux:

```bash
cat /proc/interrupts | head
vmstat 1 5
```

PASS 신호:

- `/proc/interrupts`에서 device별 interrupt count가 증가합니다.
- `vmstat`에서 `in` 필드는 interrupt rate 추세를 보여 줍니다.

FAIL 신호:

- 컨테이너나 권한 제한 환경에서는 `/proc/interrupts`가 제한될 수 있습니다.

### 면접식 되묻기

"interrupt와 syscall은 둘 다 kernel mode로 들어가는데 뭐가 다른가요?"라는 질문에는 원인을 구분합니다. interrupt는 장치나 외부 event가 비동기적으로 발생시키고, syscall은 현재 process가 동기적으로 요청합니다. page fault는 현재 명령이 만든 예외라 trap으로 설명합니다.

### 흔한 오해와 반례

오해: "packet이 오면 서버 thread가 바로 실행된다."

반례: packet은 먼저 kernel network stack과 socket buffer에 도착합니다. thread가 runnable이 되어도 CPU 경쟁, scheduler, lock contention 때문에 사용자 코드 실행은 늦어질 수 있습니다.

### Active recall

- interrupt, trap, syscall을 발생 원인 기준으로 구분해 보세요.
- packet이 NIC에서 user buffer까지 오는 중 어떤 queue를 지날 수 있나요?

## 3. Process, Thread, Context Switch, Scheduling

> process는 독립된 주소 공간과 자원 테이블을 가진 실행 단위이고, thread는 같은 process 안에서 CPU 실행 흐름을 나누는 단위입니다.
> context switch는 CPU가 한 실행 흐름의 register와 kernel bookkeeping 상태를 저장하고 다른 실행 흐름으로 바꾸는 일입니다.
> scheduler는 누가 다음 CPU 시간을 받을지 정하며, 백엔드 latency는 run queue와 blocking 상태를 함께 봐야 해석됩니다.

### 질문

thread를 늘리면 왜 항상 처리량이 늘지 않고 오히려 latency가 커질 수 있을까요?

### 직관

CPU core는 계산대입니다. 손님(thread)이 너무 적으면 계산대가 놀고, 너무 많으면 줄이 길어지고 계산원이 손님을 바꿀 때마다 장부를 다시 펼칩니다. thread 수는 병렬성뿐 아니라 대기열 길이와 전환 비용도 만듭니다.

### 작은 예시

4 core 머신에서 runnable Java thread가 200개입니다. 각 thread는 10ms CPU 작업을 하고 network I/O를 기다립니다. I/O 대기 thread는 잠들 수 있지만, 동시에 runnable인 thread가 많으면 각 thread가 CPU를 다시 받기까지 시간이 길어집니다.

### 상태 이동 trace

```
Thread A: RUNNING -> syscall read() -> SLEEPING
Thread B: RUNNABLE -----------------> RUNNING
Thread C: RUNNABLE -----------------> RUNNABLE

CPU core:
  save registers of A
  load registers of B
  update scheduler accounting
```

### 내부 메커니즘

커널은 task 상태를 runnable, running, sleeping 등으로 관리합니다. scheduler는 runnable task 중 하나를 고릅니다. context switch에서는 register, stack pointer, memory context, scheduling statistics가 바뀝니다. process 간 전환은 주소 공간 전환과 TLB 영향이 더 클 수 있고, thread 간 전환은 같은 process 주소 공간을 공유하지만 여전히 scheduler와 cache locality 비용이 있습니다.

현대 Linux 스케줄러 세부 구현은 CFS에서 EEVDF로 변화하는 등 버전 민감합니다. 그러나 안정적인 학습 모델은 같습니다. CPU는 동시에 실행 가능한 흐름이 core 수보다 많으면 시간을 나눠야 하고, blocking I/O나 lock wait은 thread를 CPU 실행에서 빼냅니다.

### 실패 모드

- thread pool이 너무 크면 context switch와 cache miss가 늘어납니다.
- CPU-bound 작업이 event-loop thread를 점유하면 network event 처리 latency가 커집니다.
- lock contention은 thread를 runnable로 보이게 하거나 sleep 상태로 밀어 CPU 사용률 해석을 어렵게 만듭니다.

### 검증 방법

Linux:

```bash
vmstat 1 5
ps -L -p <pid> -o pid,tid,stat,pcpu,comm
```

PASS 신호:

- `vmstat`의 `r` 값이 core 수보다 지속적으로 높으면 runnable queue pressure를 의심할 수 있습니다.
- `ps -L`에서 특정 process의 thread 상태가 보입니다.

FAIL 신호:

- CPU 사용률이 낮아도 latency가 높다면 runnable queue보다 lock, disk, network, GC pause를 함께 봐야 합니다.

### 면접식 되묻기

"thread와 process 차이가 뭔가요?"에는 주소 공간과 자원 테이블 공유 여부로 짧게 답합니다. 꼬리 질문이 오면 context switch 비용, TLB/cache 영향, JVM thread와 OS thread mapping, thread pool sizing으로 확장합니다.

### 흔한 오해와 반례

오해: "I/O-bound 서버는 thread를 많이 만들수록 좋다."

반례: blocking I/O 모델에서는 어느 정도 thread가 필요하지만, 너무 많으면 scheduling과 memory stack 비용이 커집니다. non-blocking I/O 모델도 event loop를 CPU-bound 작업으로 막으면 망가집니다.

### Active recall

- runnable thread 수가 core 수보다 많을 때 latency가 커지는 이유를 설명해 보세요.
- Spark executor에서 task 수가 너무 많으면 어떤 OS 자원에 압력이 생기나요?

## 4. Virtual Memory, Page Table, TLB, Page Fault

> 가상 메모리(virtual memory)는 process마다 자기만의 연속된 주소 공간을 가진 것처럼 보이게 만드는 추상화입니다.
> page table은 가상 주소를 물리 메모리 frame으로 매핑하고, TLB는 이 변환을 빠르게 하기 위한 CPU 쪽 cache입니다.
> page fault는 주소 변환이나 접근 권한 문제가 생겼을 때 CPU가 커널에 처리를 넘기는 사건이며, 정상적인 lazy allocation부터 진짜 오류까지 여러 의미를 가질 수 있습니다.

### 질문

프로그램이 `0x1000` 주소를 읽는다고 할 때, 그 주소는 실제 RAM의 `0x1000`일까요?

### 직관

가상 주소는 건물 안의 "내 사무실 101호" 같은 번호입니다. 다른 회사도 자기 층에서는 101호를 가질 수 있습니다. 실제 건물 전체의 좌표는 관리 시스템(page table)이 연결합니다.

### 작은 예시

두 process가 모두 `0x7ffee000` 같은 stack 주소를 사용합니다. 겉보기 주소는 같아도 각 process의 page table이 다른 physical frame을 가리키므로 서로의 stack을 보지 못합니다.

### 상태 이동 trace

```
CPU instruction: load [virtual 0x4000]
  |
  v
MMU checks TLB
  | hit: physical frame known
  | miss
  v
page table walk
  | present + allowed: fill TLB, continue
  | not present or not allowed
  v
page fault trap -> kernel handler
  | lazy allocate / load from file / COW / SIGSEGV
```

### 내부 메커니즘

page table은 계층 구조입니다. 모든 가상 주소에 대해 거대한 단일 배열을 두면 낭비가 크기 때문에 여러 단계의 table을 둡니다. CPU의 MMU(메모리 관리 장치)는 주소 변환을 수행하고, 자주 쓰는 변환은 TLB(주소 변환 cache)에 둡니다.

page fault는 항상 나쁜 일이 아닙니다. 처음 stack page를 만졌을 때 lazy allocation을 수행하거나, `mmap()` 파일의 page를 처음 읽을 때 disk에서 가져오거나, copy-on-write page를 쓸 때 private page를 만들기 위해 발생할 수 있습니다. 그러나 권한 없는 주소 접근이면 `SIGSEGV`로 끝납니다.

### 실패 모드

- memory pressure가 심하면 page reclaim과 swap 때문에 tail latency가 커집니다.
- 큰 heap과 많은 object는 TLB/cache locality와 GC 비용을 악화시킬 수 있습니다.
- page fault가 급증하면 CPU 사용률만 봐서는 원인을 놓칠 수 있습니다.

### 검증 방법

Linux:

```bash
time python3 - <<'PY'
data = bytearray(512 * 1024 * 1024)
for i in range(0, len(data), 4096):
    data[i] = 1
PY
/usr/bin/time -v python3 - <<'PY'
data = bytearray(128 * 1024 * 1024)
for i in range(0, len(data), 4096):
    data[i] = 1
PY
```

macOS:

```bash
/usr/bin/time -l python3 - <<'PY'
data = bytearray(128 * 1024 * 1024)
for i in range(0, len(data), 4096):
    data[i] = 1
PY
```

PASS 신호:

- page fault count 또는 memory statistics가 증가합니다.
- 첫 접근이 두 번째 접근보다 느릴 수 있습니다.

FAIL 신호:

- 시스템 memory pressure가 너무 커져 swap이나 OOM이 발생합니다. 큰 크기는 머신 상황에 맞게 낮춥니다.

### 면접식 되묻기

"page fault는 장애인가요?"에는 "항상 장애는 아닙니다"라고 답합니다. 그 다음 lazy allocation, mmap file load, copy-on-write 같은 정상 page fault와 illegal access, OOM, swap thrash 같은 문제성 page fault를 구분합니다.

### 흔한 오해와 반례

오해: "virtual memory는 RAM보다 큰 메모리를 쓰기 위한 기술이다."

반례: 그 기능도 있지만 더 근본적으로는 process isolation, address translation, protection, lazy allocation, file mapping을 가능하게 하는 추상화입니다.

### Active recall

- TLB miss와 page fault의 차이는 무엇인가요?
- Cassandra의 off-heap memory와 OS page cache를 같은 "남는 메모리"로 보면 왜 위험한가요?

## 5. File System, Page Cache, Block I/O

> 일반 파일 I/O는 대부분 사용자 buffer에서 곧장 disk platter나 SSD cell로 가지 않고, 커널 page cache를 거칩니다.
> page cache는 읽기를 빠르게 하고 작은 쓰기를 모아 주지만, `write()` 성공과 장치 영속화를 분리합니다.
> Kafka와 Cassandra는 이 차이를 이용하거나 보완하기 위해 append-only log, commit log, replication, fsync 정책을 설계합니다.

### 질문

프로그램이 파일에 쓴 byte는 언제 "진짜 디스크에 저장됐다"고 말할 수 있을까요?

### 직관

page cache는 택배 접수 창구와 비슷합니다. 접수 창구에 맡기면 사용자는 처리됐다고 느끼지만, 실제 물류 트럭에 실려 목적지에 도착하는 시점은 별도입니다. 빠른 접수와 실제 도착을 구분해야 합니다.

### 작은 예시

로그 파일에 1KB를 씁니다. `write()`는 빠르게 반환합니다. 직후 전원이 나가면 파일시스템, storage cache, flush 여부에 따라 그 byte가 남을 수도 있고 사라질 수도 있습니다. `fsync()`는 해당 파일의 dirty data와 metadata를 더 강하게 밀어 넣는 요청입니다.

### 상태 이동 trace

```
user buffer "event=1"
  |
  | write()
  v
kernel page cache: dirty page
  |
  | background writeback or fsync()
  v
block layer request queue
  |
  v
device cache / persistent media
```

### 내부 메커니즘

파일시스템은 path, inode, file offset, page cache, block mapping을 연결합니다. page cache는 파일 내용을 memory page 단위로 캐시합니다. 읽기에서는 disk에서 가져온 page를 재사용하고, 쓰기에서는 dirty page로 표시한 뒤 나중에 writeback합니다. block layer는 여러 I/O를 합치거나 재정렬하고 device driver로 넘깁니다.

Kafka는 topic partition log를 append-only file로 두고 OS page cache와 sequential I/O를 적극 활용합니다. Cassandra는 commit log에 먼저 append하고 memtable에 반영한 뒤, memtable flush로 immutable SSTable을 만듭니다. 두 시스템 모두 "빠른 append"와 "영속성 확인"의 경계를 운영 설정과 replication 모델로 조절합니다.

### 실패 모드

- page cache hit 덕분에 read latency가 낮다가 memory pressure로 cache가 밀리면 갑자기 disk read가 증가합니다.
- dirty page writeback이 몰리면 foreground write latency가 흔들릴 수 있습니다.
- `fsync()`를 너무 자주 호출하면 durability는 좋아지지만 throughput이 급격히 낮아질 수 있습니다.

### 검증 방법

Linux:

```bash
dd if=/dev/zero of=/tmp/page-cache-demo.bin bs=1m count=256
sync
time dd if=/tmp/page-cache-demo.bin of=/dev/null bs=1m
time dd if=/tmp/page-cache-demo.bin of=/dev/null bs=1m
```

PASS 신호:

- 두 번째 read가 첫 번째보다 빠르거나 disk read 통계가 덜 증가할 수 있습니다.
- `vmstat 1`에서 writeback과 I/O wait 추세를 함께 볼 수 있습니다.

FAIL 신호:

- `/tmp`가 tmpfs이거나 APFS/SSD/cache 특성 때문에 차이가 작을 수 있습니다. 실험 결과는 "개념 확인"이지 device benchmark가 아닙니다.

### 면접식 되묻기

"Kafka는 왜 disk를 써도 빠른가요?"라는 질문에는 sequential append, page cache, batch, sendfile, consumer pull을 연결합니다. 단, TLS, cache miss, follower lag, retention cleanup, compaction, fsync 설정에서는 disk I/O가 병목이 될 수 있다고 경계를 둡니다.

### 흔한 오해와 반례

오해: "page cache는 애플리케이션 cache와 같으니 둘 중 하나만 보면 된다."

반례: JVM heap cache와 OS page cache는 같은 RAM을 놓고 경쟁합니다. Kafka가 heap에 큰 message cache를 두지 않는 이유는 OS page cache와 GC 비용을 함께 고려하기 때문입니다.

### Active recall

- `write()`, `fsync()`, background writeback의 차이를 설명해 보세요.
- Cassandra compaction이 disk I/O와 page cache에 압력을 주는 이유는 무엇인가요?

## 6. Network Stack

> 네트워크 I/O는 애플리케이션 byte가 socket buffer, TCP/IP stack, qdisc, NIC queue를 지나 packet으로 나가는 과정입니다.
> 분산 시스템의 timeout은 상대 노드의 crash만 뜻하지 않습니다. 내 process scheduling, GC, socket buffer, packet loss, congestion, receiver backlog도 같은 증상으로 보일 수 있습니다.
> Kafka broker, Cassandra coordinator, Spark executor는 모두 network stack 위에서 서로의 상태를 추정합니다.

### 질문

상대 서버가 응답하지 않을 때, 정말 상대 서버가 죽었다고 말할 수 있을까요?

### 직관

전화가 안 받는다고 상대가 사라졌다고 단정할 수는 없습니다. 상대가 바쁠 수도 있고, 통신망이 막혔을 수도 있고, 내 전화기가 멈췄을 수도 있습니다. 분산 시스템의 timeout도 같은 문제입니다.

### 작은 예시

Kafka consumer가 fetch request를 보냈지만 응답이 늦습니다. 원인은 broker disk read, page cache miss, network congestion, TLS CPU overhead, consumer-side GC pause, group rebalance 중 하나일 수 있습니다.

### 상태 이동 trace

```
user send()
  |
  v
kernel socket send buffer
  |
  v
TCP segmentation / congestion window
  |
  v
qdisc / NIC queue
  |
  v
wire
  |
  v
receiver NIC -> kernel receive buffer -> user recv()
```

### 내부 메커니즘

socket은 file descriptor처럼 보이지만 내부에는 send buffer와 receive buffer, protocol state, retransmission state가 있습니다. TCP는 순서, 재전송, 흐름 제어, 혼잡 제어를 제공합니다. 애플리케이션이 `send()`를 호출해도 bytes가 즉시 상대 process의 user buffer에 도착하는 것은 아닙니다. 커널 buffer에 들어간 뒤 네트워크 상태와 receiver 처리 속도에 따라 이동합니다.

backpressure는 여기서 자연스럽게 생깁니다. receiver가 느리면 receive window가 줄고, sender의 socket buffer가 차고, `send()`가 block되거나 async send completion이 늦어집니다. Kafka producer buffer, Spark shuffle fetch, Cassandra internode messaging은 모두 이 압력을 상위 queue로 드러냅니다.

### 실패 모드

- packet loss는 retransmission과 tail latency를 만듭니다.
- receiver process가 GC pause에 들어가면 network는 살아 있어도 application-level timeout이 날 수 있습니다.
- socket buffer가 가득 차면 producer/worker가 느려지고 상위 queue가 쌓입니다.

### 검증 방법

Linux:

```bash
ss -tin
netstat -s | head -40
```

macOS:

```bash
netstat -an
netstat -s | head -40
```

PASS 신호:

- TCP connection state, send/receive queue, retransmission 통계를 확인할 수 있습니다.

FAIL 신호:

- 권한이나 OS 차이로 queue 세부 필드가 보이지 않을 수 있습니다. 이 경우 애플리케이션 metric과 packet capture를 함께 봅니다.

### 면접식 되묻기

"네트워크 partition과 node crash는 어떻게 구분하나요?"에는 "완벽히 구분할 수 없고 관측 증거를 쌓아 확률적으로 판단한다"고 답합니다. heartbeat, TCP state, application log, peer observations, kernel network counters, orchestrator event를 함께 봅니다.

### 흔한 오해와 반례

오해: "TCP는 reliable하니까 application timeout은 네트워크 문제가 아니다."

반례: TCP는 byte stream 전달을 재시도하지만 무한히 빠르게 보장하지 않습니다. 재전송, flow control, congestion, receiver stall은 application timeout으로 나타날 수 있습니다.

### Active recall

- `send()` 성공과 상대 process의 `recv()` 성공 사이에는 어떤 queue가 있나요?
- Spark shuffle fetch timeout을 OS/network 관점에서 분해해 보세요.

## 7. Locks, Synchronization, Concurrency Bugs

> lock은 여러 실행 흐름이 같은 상태를 동시에 바꾸지 못하게 순서를 만드는 도구입니다.
> 동기화는 correctness를 지키지만, lock contention은 CPU, scheduler, cache coherence, blocking wait 비용을 만듭니다.
> 분산 시스템의 leader, quorum, transaction도 더 큰 범위에서 "누가 어떤 순서로 상태를 바꿀 권한이 있는가"를 정하는 동기화 문제입니다.

### 질문

공유 counter 하나를 여러 thread가 증가시키는 일이 왜 어려울까요?

### 직관

은행 창구 두 명이 같은 장부의 잔액을 동시에 읽고 각자 더한 뒤 다시 쓰면 한 번의 입금이 사라질 수 있습니다. lock은 장부를 한 명씩만 만지게 합니다.

### 작은 예시

`count++`는 하나의 동작처럼 보이지만 CPU 수준에서는 load, add, store로 나뉩니다.

```
Thread A: load count=0
Thread B: load count=0
Thread A: store count=1
Thread B: store count=1
```

결과는 2가 아니라 1입니다.

### 상태 이동 trace

```
without lock:
  count=0
  A reads 0
  B reads 0
  A writes 1
  B writes 1

with lock:
  A lock -> read 0 -> write 1 -> unlock
  B lock -> read 1 -> write 2 -> unlock
```

### 내부 메커니즘

동기화는 atomic instruction, memory barrier, futex, mutex, semaphore, condition variable 같은 계층을 포함합니다. user-space에서 빠르게 lock을 잡을 수 있으면 kernel 진입 없이 끝나지만, 경쟁이 있으면 futex 같은 syscall로 sleep/wakeup을 처리할 수 있습니다. memory ordering도 중요합니다. CPU와 compiler는 성능을 위해 load/store 순서를 바꿀 수 있으므로 동기화 primitive는 순서 보장을 함께 제공합니다.

### 실패 모드

- data race는 간헐적이고 재현이 어렵습니다.
- deadlock은 서로의 lock을 기다리며 멈춥니다.
- lock convoy는 한 lock을 기다리는 thread들이 줄지어 scheduler와 cache 비용을 키웁니다.
- distributed lock을 local mutex처럼 생각하면 network partition에서 split-brain 문제가 생깁니다.

### 검증 방법

JVM:

```bash
jstack <pid> | sed -n '1,120p'
jcmd <pid> Thread.print | sed -n '1,120p'
```

Linux:

```bash
perf top
```

PASS 신호:

- thread dump에서 `BLOCKED`, `WAITING`, monitor owner, lock wait stack을 볼 수 있습니다.

FAIL 신호:

- sampling 순간에만 보이는 stack이므로 짧은 lock contention은 놓칠 수 있습니다. 반복 sampling과 애플리케이션 metric이 필요합니다.

### 면접식 되묻기

"mutex와 semaphore 차이는 무엇인가요?"라는 질문에는 mutex는 보통 하나의 critical section 소유권을 보호하고, semaphore는 제한된 permit 수로 동시 진입 개수를 제어한다고 답합니다. 꼬리 질문에서는 futex와 kernel sleep/wakeup 비용, 분산 lock의 failure model 차이로 내려갑니다.

### 흔한 오해와 반례

오해: "lock을 쓰면 thread-safe하니 성능 문제만 남는다."

반례: lock 범위가 너무 넓으면 성능이 무너지고, lock 순서가 일관되지 않으면 deadlock이 생기며, lock이 보호해야 할 모든 상태를 포함하지 않으면 여전히 data race가 남습니다.

### Active recall

- `count++`가 atomic하지 않은 이유를 trace로 설명해 보세요.
- local lock과 distributed consensus의 공통점과 차이를 말해 보세요.

## 8. Observability and Performance Debugging

> 성능 디버깅은 "느리다"를 CPU, memory, disk, network, lock, runtime, distributed state 중 어디의 queue가 길어졌는지 분해하는 작업입니다.
> 한 도구의 숫자는 결론이 아니라 관측값입니다. `top`, `vmstat`, `iostat`, `ss`, `lsof`, `jstack`, `perf`를 서로 연결해야 원인 후보가 줄어듭니다.
> 분산 시스템 장애에서 OS 관측값은 제품 metric을 해석하는 lower-layer 근거가 됩니다.

### 질문

서버 latency가 증가했을 때 어디부터 봐야 할까요?

### 직관

병원 triage처럼 먼저 생명 신호를 봅니다. CPU가 꽉 찼는지, memory pressure가 있는지, disk가 막혔는지, network가 재전송 중인지, thread가 lock을 기다리는지 확인합니다. 그 다음 Kafka lag, Cassandra pending compaction, Spark stage duration 같은 제품 metric으로 올라갑니다.

### 작은 예시

API latency p99가 200ms에서 3s로 늘었습니다. CPU는 40%입니다. 이때 "CPU가 남으니 애플리케이션 버그"라고 단정하면 위험합니다. disk I/O wait, GC pause, socket retransmission, DB quorum wait가 숨어 있을 수 있습니다.

### 상태 이동 trace

```
symptom: p99 latency up
  |
  +-- CPU: run queue high? user/sys time?
  +-- memory: page fault, swap, OOM, GC?
  +-- disk: await/util/writeback/compaction/spill?
  +-- network: retransmit, send-Q/recv-Q, timeout?
  +-- locks: thread dump BLOCKED/WAITING?
  +-- distributed: replica lag, consumer lag, shuffle skew?
```

### 내부 메커니즘

관측 도구는 각기 다른 계층을 봅니다. `vmstat`는 runnable queue, memory, swap, I/O, interrupt, context switch 추세를 줍니다. `iostat`는 block device latency와 utilization을 봅니다. `ss`는 socket 상태를 보여 줍니다. `jstack`은 JVM thread 상태를 보여 줍니다. `perf`와 ftrace는 kernel/user CPU hot path를 더 깊게 봅니다.

숫자는 항상 workload와 함께 읽어야 합니다. `iowait`가 높아도 실제 병목은 remote storage일 수 있고, CPU가 낮아도 thread가 lock이나 network를 기다릴 수 있습니다.

### 실패 모드

- 단일 metric만 보고 결론을 내리면 잘못된 scaling을 할 수 있습니다.
- 평균 latency만 보면 tail latency와 queueing을 놓칩니다.
- 컨테이너 내부 metric과 host metric이 다를 수 있습니다.

### 검증 방법

Linux quick triage:

```bash
date
uptime
vmstat 1 5
iostat -xz 1 5
ss -s
```

JVM:

```bash
jcmd <pid> Thread.print
jcmd <pid> GC.heap_info
```

PASS 신호:

- 최소 두 계층의 증거가 같은 방향을 가리킵니다. 예를 들어 Spark stage가 느리고, executor log에 spill이 늘며, host `iostat`의 write latency도 증가합니다.

FAIL 신호:

- 제품 metric과 OS metric이 맞지 않습니다. 예를 들어 Kafka consumer lag는 증가하지만 broker disk/network는 한가하면 consumer processing, group rebalance, downstream sink를 봐야 합니다.

### 면접식 되묻기

"장애 상황에서 어떤 명령을 보나요?"에는 명령어 이름만 나열하지 말고 각 명령이 어느 queue를 보는지 말합니다. `vmstat`는 CPU runnable과 memory/I/O 추세, `iostat`는 device latency, `ss`는 socket queue, `jstack`은 JVM thread state, 제품 metric은 replica/log/task 상태를 봅니다.

### 흔한 오해와 반례

오해: "CPU 사용률이 낮으면 서버는 여유롭다."

반례: 서버는 disk, network, lock, remote quorum, GC pause 때문에 CPU를 못 쓰고 기다릴 수 있습니다. 낮은 CPU는 여유가 아니라 대기일 수 있습니다.

### Active recall

- "느림"을 queue 관점으로 나누면 어떤 후보가 나오나요?
- Kafka broker latency와 OS page cache pressure를 함께 확인하려면 무엇을 보나요?

## 현실 시나리오 1: `write()`는 빨랐는데 장애 후 로그가 사라졌다

> `write()` 성공은 사용자 byte가 커널에 받아들여졌다는 신호일 수 있고, device에 영속화됐다는 신호가 아닐 수 있습니다.
> durable write를 논하려면 page cache, filesystem, storage cache, `fsync()`, application replication을 분리해야 합니다.

1. 관측된 증상

    애플리케이션 로그에는 "commit written"이 찍혔지만 crash 후 일부 record가 복구되지 않았습니다.

2. 가능한 원인 후보

    `fsync()` 누락, filesystem mount option, storage write cache, process가 success log를 너무 일찍 찍음, replica ack 의미 오해.

3. OS/kernel 관점에서 볼 지점

    `write()`가 dirty page 생성까지만 했는지, `fsync()`가 호출됐는지, writeback과 device flush가 있었는지 봅니다.

4. distributed-system 관점에서 볼 지점

    단일 node durability와 replica durability를 분리합니다. replication ack가 있었다면 어느 replica set에서 어떤 quorum 조건을 만족했는지 봅니다.

5. Kafka/Cassandra/Spark 내부 구조와 연결되는 지점

    Kafka는 page cache와 replication을 함께 씁니다. Cassandra는 commit log sync policy와 replica CL이 중요합니다. Spark는 checkpoint location이 local disk인지 reliable storage인지가 중요합니다.

6. 확인 명령 또는 로그

    `strace -e trace=write,fsync`, application log, Kafka broker config, Cassandra commitlog sync config, Spark checkpoint path.

7. 잘못된 결론의 예

    "`write()`가 성공했으니 OS가 데이터를 잃었다."

8. 더 나은 추론 과정

    먼저 "성공 로그가 어떤 경계 뒤에 찍혔는가"를 찾습니다. user buffer에서 page cache로 간 성공인지, `fsync()` 이후인지, replica quorum ack 이후인지 나누면 장애 모델이 좁아집니다.

## 현실 시나리오 2: CPU는 낮은데 p99 latency가 오른다

> 낮은 CPU 사용률은 항상 여유를 뜻하지 않습니다.
> thread가 disk, network, lock, GC, remote quorum을 기다리면 CPU는 비어 보여도 요청 latency는 커질 수 있습니다.

1. 관측된 증상

    API p99가 2초로 증가했지만 host CPU는 35%입니다.

2. 가능한 원인 후보

    disk I/O wait, socket retransmission, DB replica lag, JVM GC pause, lock contention, thread pool starvation, downstream backpressure.

3. OS/kernel 관점에서 볼 지점

    `vmstat`의 run queue와 block, `iostat`의 await/util, `ss`의 retransmission/queue, `jstack`의 thread state를 봅니다.

4. distributed-system 관점에서 볼 지점

    timeout은 crash와 slow를 구분하지 못합니다. client retry와 idempotency, quorum wait, queue growth를 함께 봅니다.

5. Kafka/Cassandra/Spark 내부 구조와 연결되는 지점

    Kafka consumer lag, Cassandra pending compaction/read latency, Spark shuffle spill/stage skew가 OS 대기와 연결될 수 있습니다.

6. 확인 명령 또는 로그

    `vmstat 1`, `iostat -xz 1`, `ss -tin`, `jcmd <pid> Thread.print`, 제품별 metric.

7. 잘못된 결론의 예

    "CPU가 낮으니 서버를 더 띄우면 해결된다."

8. 더 나은 추론 과정

    CPU가 왜 낮은지 먼저 묻습니다. runnable이 적은지, sleep이 많은지, I/O wait인지, remote wait인지 분리한 뒤 scale-out, config 변경, query/data layout 변경 중 하나를 선택합니다.

## 근거와 더 읽을 자료

- [10_source_ledger.md](10_source_ledger.md)의 Linux kernel, man-pages, Kafka hardware/OS, Cassandra storage, Spark tuning 항목
- Linux kernel page tables: https://docs.kernel.org/mm/page_tables.html
- Linux page cache: https://docs.kernel.org/next/mm/page_cache.html
- man7 syscalls: https://man7.org/linux/man-pages/man2/syscalls.2.html
- man7 sendfile: https://man7.org/linux/man-pages/man2/sendfile.2.html

