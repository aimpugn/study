# 01. OS Kernel Foundations

OS 커널을 공부하는 이유는 "운영체제 면접 키워드"를 외우기 위해서가 아닙니다. Kafka broker, Cassandra node, Spark executor는 결국 하나의 OS process이고, 그 process는 CPU 시간, 가상 메모리, 파일, page cache, socket buffer, disk queue를 커널에게 빌려 씁니다. 애플리케이션 로그에는 `timeout`, `GC pause`, `request latency`, `consumer lag`, `shuffle spill`처럼 보이지만, 많은 경우 아래층에서는 runnable task가 CPU를 기다리거나, dirty page가 writeback을 기다리거나, socket buffer가 비거나 차기를 기다리고 있습니다.

이 문서는 한 머신 안에서 커널이 무엇을 관리하는지 lower layer부터 쌓습니다. 목적은 `syscall`, `context switch`, `page fault`, `page cache`, `socket buffer`라는 말을 아는 것이 아니라, 한 요청이 실제로 어디서 어디로 이동하는지 설명하는 것입니다.

## 1. `write()`는 왜 커널에 부탁해야 하는가

> `write(fd, buf, len)`은 문법상 C 함수 호출처럼 보이지만, 의미상으로는 프로세스가 운영체제 커널에게 "내가 가진 파일 핸들에 이 바이트들을 반영해 달라"고 요청하는 일입니다.
> 커널이 필요한 이유는 디스크, 네트워크 카드, 메모리 매핑, 파일 offset 같은 자원이 한 프로세스의 소유물이 아니라 여러 프로그램이 함께 쓰는 전역 자원이기 때문입니다.
> 이 흐름을 설명할 수 있으면 Kafka의 page cache와 `sendfile`, Cassandra의 commit log와 `fsync`, Spark의 shuffle spill이 왜 애플리케이션 코드만의 문제가 아닌지 같이 설명할 수 있습니다.

`write(fd, buf, len)`을 처음 보면 평범한 라이브러리 함수처럼 느껴집니다. 같은 프로세스 안에서 `strlen()`이나 `memcpy()`를 부르면 CPU는 현재 프로그램의 명령을 그대로 실행하고, 함수가 끝나면 원래 코드로 돌아옵니다. 그런데 파일 쓰기는 그렇게 끝나지 않습니다. `fd`가 가리키는 파일은 현재 프로세스의 지역 변수 안에 들어 있는 물건이 아니고, 그 파일이 놓인 파일시스템과 디스크 장치도 현재 프로세스 혼자 쓰는 물건이 아닙니다. 같은 순간에 다른 프로세스가 같은 파일을 열 수도 있고, 데이터베이스가 다른 파일을 쓰고 있을 수도 있으며, 커널은 메모리 압박 때문에 page cache의 일부를 밀어내야 할 수도 있습니다.

초기 컴퓨터가 점점 비싸고 큰 공용 장비가 되면서 운영체제가 해결해야 할 질문은 단순했습니다. 여러 사람이 한 기계에 접속하고, 여러 프로그램이 동시에 실행되면 누가 CPU를 얼마나 쓰는지, 누가 어느 메모리를 볼 수 있는지, 누가 어떤 파일과 장치를 건드릴 수 있는지를 한곳에서 조정해야 합니다. 이 필요에서 CPU 시간을 나누는 시분할, 사용자별 권한, 프로세스별 주소 공간, 장치 접근 중재가 함께 발전했습니다. 그러니 "시분할 시스템이 등장해서 경계가 생겼다"라고 외우면 순서가 흐려집니다. 더 아래의 원인은 공유되는 하드웨어를 안전하게 나눠 써야 한다는 요구이고, 사용자 모드와 커널 모드는 그 요구를 CPU와 운영체제 구조로 구현한 방식입니다.

여기서 말하는 사용자 모드와 커널 모드의 차이는 분위기나 폴더 구분이 아닙니다. CPU가 실행 중인 코드에 부여하는 권한 수준이 다릅니다. 사용자 모드(user mode)에서 실행되는 애플리케이션은 자기에게 매핑된 가상 주소를 읽고 쓰고, 보통의 산술 명령이나 분기 명령을 실행할 수 있습니다. 하지만 page table을 마음대로 바꾸거나, 디스크 컨트롤러에 직접 명령을 보내거나, 다른 프로세스의 file descriptor table을 열어 보는 일은 할 수 없습니다. 커널 모드(kernel mode)는 그런 권한 있는 명령과 커널 내부 자료구조에 접근할 수 있는 실행 상태입니다. `경계`라는 말은 바로 이 권한 수준, 접근 가능한 주소 범위, 접근 가능한 자료구조가 바뀌는 지점을 뜻합니다.

그렇다고 애플리케이션이 커널 함수를 그냥 호출할 수 있는 것은 아닙니다. 커널 코드와 자료구조가 어떤 가상 주소 배치로 매핑되는지는 운영체제와 CPU 아키텍처에 따라 달라질 수 있지만, 사용자 코드가 접근 가능한 라이브러리 함수처럼 노출되는 것은 아닙니다. 사용자 모드 코드는 CPU가 허용한 system call 또는 trap 진입점으로만 커널 실행 경로에 들어갑니다. 리눅스에서 보통 C 라이브러리의 `write()` wrapper는 인자를 정해진 레지스터나 ABI 규칙에 맞게 놓고, 해당 아키텍처의 system call 진입 명령을 실행합니다. 이 명령이 실행되면 CPU는 현재 실행 흐름을 멈추고 커널이 등록해 둔 진입점으로 이동합니다. 이때 권한 수준이 커널 모드로 바뀌고, 커널은 현재 프로세스가 요청한 system call 번호와 인자를 읽어 `write` 처리 경로로 보냅니다.

```
process memory
  buf -> "hello"
  fd  -> 3
       |
       | libc write wrapper
       |   system call number = write
       |   arguments = fd, user buffer address, length
       v
CPU syscall entry
  user mode -> kernel mode
       |
       v
kernel syscall dispatch
  fd 3 -> current process fd table
       -> open file description
       -> file offset, flags, filesystem operations
       |
       v
VFS and filesystem write path
  check: fd is valid and writable
  check: user buffer address can be read
  copy bytes from user memory into kernel-managed file cache path
       |
       v
page cache
  file page containing the target offset becomes dirty
  write() may return after bytes are accepted by this path
       |
       v
later writeback / fsync-triggered flush
  filesystem -> block layer -> driver -> storage device
```

이 trace에서 중요한 점은 "문자열이 디스크로 바로 날아간다"가 아니라 "요청이 커널의 파일 객체와 파일시스템 경로로 번역된다"입니다. 사용자 프로세스가 넘긴 `buf`는 사용자 주소입니다. 커널은 그 주소가 현재 프로세스가 읽을 수 있는 메모리인지 확인해야 합니다. 잘못된 포인터라면 `EFAULT` 같은 오류로 실패할 수 있습니다. `fd`도 확인해야 합니다. 이미 닫힌 번호이거나 쓰기 권한 없이 열린 파일이면 `EBADF` 같은 오류가 날 수 있습니다. 파일이 일반 파일인지, pipe인지, socket인지에 따라 뒤쪽의 처리 함수도 달라집니다.

반환값도 조심해야 합니다. `write(fd, buf, len)`이 성공하면 `len`이 아니라 실제로 받아들인 byte 수를 반환합니다. 반환값이 `len`보다 작으면 partial write입니다. 디스크 공간 부족, resource limit, signal, socket/pipe의 상태 같은 이유로 일부만 기록될 수 있고, 이때 호출자는 남은 byte를 다시 쓰는 loop를 가져야 합니다. 견고한 파일 쓰기 코드는 `write()`를 한 번 호출하고 끝내지 않고, `ret < len`인 경우 `buf + ret`부터 남은 길이를 다시 요청합니다.

일반적인 buffered regular-file I/O에서는 파일의 해당 위치에 대응하는 page cache가 핵심 중간 지점이 됩니다. page cache는 파일 내용을 메모리 page 단위로 들고 있는 커널의 캐시입니다. 읽기에서는 디스크에서 가져온 파일 page를 재사용하게 해 주고, 쓰기에서는 애플리케이션이 준 바이트를 먼저 메모리 안의 파일 page에 반영한 뒤 그 page를 dirty 상태로 표시할 수 있습니다. dirty page는 "파일의 최신 내용이 메모리에는 있지만 아직 저장 장치에 완전히 반영되지 않았을 수 있다"는 뜻입니다. 단, `O_DIRECT`처럼 page cache를 우회하려는 경로, 특수 파일, 일부 장치/파일시스템 경로는 이 설명의 예외가 될 수 있습니다.

더 강한 내구성이 필요하면 애플리케이션은 `fsync()`나 `fdatasync()` 같은 별도 system call로 dirty data를 저장 장치 쪽으로 밀어 넣어 달라고 요청합니다. `fsync()`는 해당 파일의 데이터와 필요한 metadata를 동기화하려는 fd 단위 요청이지만, 새 파일을 만들거나 rename한 경우 부모 directory entry의 내구성까지 별도로 고려해야 할 수 있습니다. 파일시스템, 디스크 캐시, 컨트롤러, 네트워크 파일시스템 같은 층도 관계하므로 정확한 보장은 환경에 따라 달라집니다.

성능도 같은 자리에서 이해해야 합니다. system call 하나에는 CPU 권한 전환, 커널 진입/복귀, 인자 검증, 파일 객체 lookup, lock 획득, 사용자 메모리에서 커널 쪽으로의 바이트 복사, scheduler 개입 가능성이 붙습니다. 한 번의 비용은 작을 수 있지만, 1 byte씩 백만 번 쓰면 애플리케이션은 실제 디스크 처리보다 system call과 lock, copy 비용에 더 많은 시간을 쓰게 됩니다. 그래서 고성능 서버는 작은 요청을 모아 batch로 보내고, 사용자 공간과 커널 공간 사이의 불필요한 복사를 줄이려 합니다. Kafka가 큰 record batch, 순차 append, OS page cache, file-to-socket 전송 경로를 중요하게 보는 이유도 여기서 시작합니다.

면접에서 "`write()`가 성공했으면 디스크에 저장된 건가요?"라고 물으면 첫 문장은 짧게 가면 됩니다. "아닐 수 있습니다. 일반적인 buffered I/O에서는 `write()` 성공이 커널의 파일 쓰기 경로와 page cache 반영을 뜻할 수 있고, 저장 장치에 안정적으로 내려갔는지는 `fsync()` 같은 별도 요청과 파일시스템/스토리지 보장까지 봐야 합니다." 꼬리 질문이 오면 `fd -> open file description -> VFS -> page cache -> dirty writeback -> fsync` 순서로 다시 내려가면 됩니다.

## 2. Interrupt, Trap, Syscall은 CPU 흐름을 어떻게 바꾸는가

프로세스가 실행 중이라는 말은 CPU가 그 프로세스의 instruction pointer와 register 값을 기준으로 명령을 실행하고 있다는 뜻입니다. 그런데 CPU는 항상 사용자 코드만 순서대로 실행하지 않습니다. 네트워크 카드가 packet을 받으면 interrupt가 오고, 잘못된 메모리 주소를 접근하면 page fault라는 trap이 생기며, `write()` wrapper는 의도적으로 syscall trap을 발생시킵니다.

interrupt는 외부 장치가 CPU에게 "처리할 일이 있다"고 알리는 비동기 사건입니다. 현재 프로세스가 `for` loop를 돌고 있든 JVM bytecode를 실행하고 있든, 장치 interrupt가 오면 CPU와 커널은 정해진 handler로 들어갑니다. handler는 보통 긴 일을 다 하지 않습니다. packet을 아주 간단히 받아 queue에 올리고, 더 무거운 network stack 처리는 softirq나 kernel thread 같은 뒤쪽 경로로 넘길 수 있습니다.

trap은 현재 실행 흐름 안에서 생긴 진입입니다. page fault는 CPU가 현재 instruction을 실행하다가 "이 가상 주소를 물리 메모리로 번역할 수 없다"는 사실을 발견해 커널로 들어가는 사건입니다. syscall도 trap의 한 형태로 볼 수 있습니다. 차이는 의도입니다. page fault는 코드가 메모리를 접근하다가 만난 예외이고, syscall은 애플리케이션이 커널 기능을 요청하기 위해 의도적으로 만든 진입입니다.

```
NIC receives packet
  -> interrupt handler
  -> kernel network stack / socket receive buffer
  -> waiting thread becomes runnable
  -> scheduler later chooses the thread
  -> user code returns from epoll_wait()
```

이 흐름 때문에 "packet이 도착했다"와 "서버 thread가 바로 사용자 코드를 실행했다"는 같은 말이 아닙니다. packet은 먼저 커널 network stack과 socket receive buffer에 도착합니다. 기다리던 thread가 runnable 상태가 되어도, CPU가 다른 일을 하고 있거나 run queue가 길거나 lock을 기다리면 애플리케이션 관점의 latency는 늘어납니다.

면접에서 interrupt와 syscall을 비교하라는 질문을 받으면 원인을 기준으로 나누면 됩니다. interrupt는 장치가 비동기적으로 CPU의 흐름을 끊습니다. syscall은 현재 thread가 동기적으로 커널에게 요청합니다. page fault는 현재 instruction이 만든 예외입니다. 셋 모두 커널 코드로 들어갈 수 있지만, 누가 왜 그 진입을 만들었는지가 다릅니다.

## 3. Process, Thread, Scheduler는 CPU 시간을 어떻게 나누는가

process는 주소 공간과 file descriptor table 같은 자원 묶음을 가진 실행 단위입니다. thread는 그 process 안에서 CPU가 실제로 실행하는 흐름입니다. 같은 process의 thread들은 heap과 열린 파일을 공유하지만, 각 thread는 자기 register 상태와 stack, scheduling 상태를 가집니다.

CPU core보다 runnable thread가 많으면 모든 thread가 동시에 실행될 수 없습니다. scheduler는 run queue에 있는 실행 가능한 task 중 다음에 CPU를 받을 대상을 고릅니다. context switch는 현재 thread의 register와 커널 bookkeeping 상태를 저장하고, 다른 thread의 상태를 복원해 CPU가 다른 실행 흐름을 이어가게 하는 일입니다.

```
Thread A running
  -> timer interrupt or blocking syscall
  -> kernel saves A's CPU state
  -> scheduler selects Thread B
  -> kernel restores B's CPU state
  -> Thread B returns to user mode
```

thread를 늘리면 처리량이 늘 것 같지만 항상 그렇지 않습니다. CPU-bound 작업이 core 수보다 많으면 run queue가 길어지고 context switch가 늘어납니다. I/O-bound 작업도 lock, memory pressure, socket buffer, disk wait 때문에 runnable과 sleep 사이를 오갑니다. 백엔드에서 CPU 사용률이 낮은데 p99 latency가 높다면 "CPU가 남는다"라고 단정하기 전에 thread dump, run queue, blocking syscall, GC, lock wait를 같이 봐야 합니다.

Kafka는 network thread, I/O thread, request handler가 서로 다른 queue를 지나며 일합니다. Cassandra는 request 처리 thread와 compaction thread가 같은 disk bandwidth와 CPU를 나눠 씁니다. Spark executor는 task thread가 많아질수록 CPU parallelism은 늘 수 있지만, shuffle spill과 GC pressure도 같이 늘 수 있습니다. 그래서 thread 수는 "많을수록 좋다"가 아니라 "어느 자원이 병목인지"와 함께 정해야 합니다.

## 4. Virtual Memory, Page Table, Page Fault는 주소를 어떻게 속이는가

프로세스가 보는 주소는 물리 메모리 주소가 아니라 virtual address입니다. 같은 숫자 `0x1000`이라도 process A와 process B에서 다른 physical frame을 가리킬 수 있습니다. 이 덕분에 한 process가 다른 process의 메모리를 쉽게 훔쳐보지 못하고, 커널은 각 process에게 독립된 메모리 공간을 가진 것처럼 보이게 할 수 있습니다.

CPU의 MMU(memory management unit)는 virtual address를 page table을 통해 physical frame으로 번역합니다. 자주 쓰는 번역 결과는 TLB(translation lookaside buffer)에 캐시됩니다. page table에 mapping이 없거나 권한이 맞지 않으면 page fault가 발생하고, 커널이 그 fault를 처리합니다.

page fault는 항상 오류가 아닙니다. 프로그램이 처음 접근한 stack page를 커널이 그때 할당할 수 있고, `mmap()`으로 매핑한 파일 page를 첫 접근 시점에 읽어 올 수도 있습니다. fork 후 copy-on-write에서는 부모와 자식이 같은 physical page를 공유하다가 한쪽이 쓰려고 할 때 fault가 나고, 커널이 새 page를 복사해 분리합니다. 반대로 정말 잘못된 주소 접근이면 process는 segmentation fault로 종료될 수 있습니다.

```
user load [virtual address X]
  -> MMU checks TLB
  -> TLB miss: walk page table
  -> no valid mapping or permission mismatch
  -> page fault trap
  -> kernel decides:
       allocate page / load file page / copy-on-write / kill process
```

분산 시스템 문서에서 memory를 말할 때도 이 층을 잊으면 안 됩니다. Spark executor heap이 커지고 off-heap buffer와 page cache가 같이 메모리를 쓰면, 애플리케이션 heap만 봐서는 부족합니다. Cassandra의 memtable, Bloom filter, index cache, OS page cache는 서로 다른 이름을 갖지만 결국 같은 physical memory를 나눠 씁니다. 메모리 압박은 page cache 축소, swap, GC 증가, OOM kill 같은 전혀 다른 증상으로 나타날 수 있습니다.

## 5. File I/O와 Page Cache는 성능과 내구성을 동시에 만든다

파일 I/O에서 page cache는 성능의 친구이자 내구성 오해의 원인입니다. 같은 파일을 두 번 읽을 때 두 번째가 빠른 이유는 애플리케이션 코드가 갑자기 똑똑해져서가 아니라, 첫 번째 read가 파일 page를 커널 메모리에 올려 두었기 때문일 수 있습니다. 반대로 write가 빠르게 반환되는 이유도 디스크가 즉시 모든 byte를 물리적으로 기록해서가 아니라, 커널이 dirty page로 받아 둔 뒤 나중에 writeback할 수 있기 때문일 수 있습니다.

VFS(virtual filesystem)는 사용자에게 공통 파일 API를 보여 주는 커널 계층입니다. VFS 자체가 데이터를 영구 저장하는 것은 아닙니다. VFS는 `read`, `write`, `fsync` 같은 요청을 해당 파일의 filesystem/file operation으로 dispatch하고, 그 뒤에서 ext4, XFS, tmpfs, socket, pipe 같은 대상별 구현이 동작합니다. 일반 regular file write는 page cache와 filesystem writeback을 통과하지만, socket write는 network stack으로 가고, pipe write는 pipe buffer로 갑니다.

```
regular file write
  -> VFS
  -> filesystem
  -> page cache dirty page
  -> writeback
  -> block layer queue
  -> device driver
  -> storage

socket write
  -> VFS-like fd dispatch
  -> socket send buffer
  -> TCP congestion/flow control
  -> NIC queue
```

Kafka가 segment file append와 page cache를 잘 활용하는 이유는 sequential write와 sequential read가 kernel/filesystem/storage에 친화적이기 때문입니다. Cassandra는 commit log를 append하고 memtable을 memory에 두었다가 SSTable로 flush합니다. Spark는 shuffle 중간 결과를 memory에 유지하지 못하면 local disk spill file로 밀어냅니다. 세 경우 모두 "파일을 쓴다"라는 상위 표현 아래에 page cache와 writeback, disk bandwidth, fsync 정책이 숨어 있습니다.

## 6. Network Stack과 Socket Buffer는 distributed latency의 첫 관문이다

분산 시스템에서 네트워크는 단순한 선이 아닙니다. 애플리케이션이 `send()`를 호출하면 byte는 바로 상대 process의 변수로 들어가지 않습니다. 먼저 현재 process의 system call을 거쳐 kernel socket send buffer로 들어가고, TCP는 congestion window와 receiver window를 보며 보낼 수 있는 양을 정합니다. NIC queue로 내려간 packet은 switch와 router, 상대 NIC, 상대 kernel network stack, 상대 socket receive buffer를 거쳐야 합니다.

```
sender process
  -> send()
  -> kernel socket send buffer
  -> TCP segmentation / retransmission state
  -> NIC queue
  -> network
  -> receiver NIC
  -> kernel socket receive buffer
  -> receiver process read()
```

이 구조 때문에 "네트워크가 느리다"라는 말은 너무 큽니다. sender가 socket buffer가 비기를 기다릴 수도 있고, receiver가 application thread scheduling 때문에 read를 늦게 할 수도 있으며, packet loss로 retransmission이 늘 수도 있고, TLS 처리로 CPU가 병목일 수도 있습니다. Kafka consumer lag, Cassandra read timeout, Spark shuffle fetch failure를 볼 때도 application metric만 보지 말고 socket queue, retransmission, bandwidth, CPU, GC를 함께 봐야 합니다.

## 7. Lock, Queue, Observability는 증상을 해석하는 언어다

커널과 런타임은 공유 상태를 보호하기 위해 lock을 씁니다. lock은 correctness를 지키지만, contention이 생기면 thread는 CPU를 쓰고 싶어도 기다립니다. queue는 burst를 흡수하지만, consumer가 producer보다 느리면 길어집니다. 관찰은 이 둘을 추적하는 일입니다.

lock을 이해할 때 첫 질문은 "왜 느린가"가 아니라 "무엇을 동시에 바꾸면 안 되는가"입니다. 두 thread가 같은 counter를 읽고 각각 1을 더해 다시 쓰면, 둘 다 같은 old value를 읽고 같은 new value를 저장해 update 하나가 사라질 수 있습니다. 이것이 race입니다. lock은 한 번에 한 thread만 critical section에 들어가게 해 이 문제를 막습니다.

```
without lock:
  T1 read count=10
  T2 read count=10
  T1 write count=11
  T2 write count=11
  lost update

with lock:
  T1 acquire lock
  T1 read/write count=11
  T1 release lock
  T2 acquire lock
  T2 read/write count=12
```

하지만 lock은 기다림을 만듭니다. thread A가 lock을 잡고 disk I/O나 network I/O를 기다리면 thread B는 CPU가 비어 있어도 같은 critical section에 들어가지 못합니다. 두 lock을 서로 반대 순서로 잡으면 deadlock이 생길 수 있습니다. thread A는 lock X를 잡고 Y를 기다리고, thread B는 Y를 잡고 X를 기다리는 식입니다. lost wakeup도 중요한 버그입니다. 어떤 thread가 "조건이 만족되면 깨워 달라"고 등록하기 전에 다른 thread가 이미 signal을 보내면, 기다리는 thread가 영원히 잠들 수 있습니다. 그래서 condition variable, futex, monitor 같은 primitive는 조건 확인과 sleep 등록을 원자적으로 묶어야 합니다.

memory ordering도 여기서 이어집니다. 여러 CPU core는 store buffer와 cache를 갖고 있고, compiler와 CPU는 성능을 위해 일부 memory operation의 관측 순서를 바꿀 수 있습니다. lock, volatile/atomic, memory barrier는 "다른 thread가 어떤 순서로 값을 보아야 하는가"를 정하는 장치입니다. Java의 `synchronized`, `volatile`, `Lock`도 결국 OS thread scheduling과 CPU memory model 위에 올라갑니다. Kafka, Cassandra, Spark가 JVM 위에서 동작한다는 말은 이 runtime-level lock과 wait가 kernel scheduler, futex, park/unpark, I/O wait와 만난다는 뜻입니다.

```
application thread
  -> tries to enter synchronized block
  -> JVM monitor / lock state
  -> may spin briefly or park
  -> OS scheduler may put thread to sleep
  -> another thread unlocks and wakes it
  -> runnable queue -> CPU
```

이 흐름 때문에 thread dump에서 `BLOCKED`, `WAITING`, `TIMED_WAITING`, `RUNNABLE`을 구분하는 일이 중요합니다. `BLOCKED`는 Java monitor 진입을 기다리는 상태일 수 있고, `WAITING`은 condition이나 park를 기다릴 수 있습니다. `RUNNABLE`도 항상 CPU에서 계산 중이라는 뜻은 아닙니다. native I/O wait나 kernel syscall 안의 대기도 JVM에서는 runnable처럼 보일 수 있습니다.

백엔드 장애를 볼 때는 다음 순서가 실용적입니다.

| 관측 | 바로 단정하면 안 되는 것 | 같이 볼 lower layer |
|---|---|---|
| CPU 사용률 낮음 | 여유가 많다 | run queue, blocking I/O, lock wait, GC pause |
| disk write latency 증가 | 디스크만 문제다 | page cache dirty ratio, fsync 빈도, compaction/spill |
| network timeout | 상대 노드가 죽었다 | retransmission, socket queue, GC, scheduler delay |
| memory free 감소 | leak이다 | page cache, heap, off-heap, mmap, file cache |
| Kafka/Cassandra/Spark queue 증가 | worker 수만 늘리면 된다 | downstream backpressure, disk/network, partition skew |

Linux에서는 `strace`, `perf`, `vmstat`, `iostat`, `ss`, `/proc`, eBPF 계열 도구가 도움이 됩니다. macOS는 `/proc`과 Linux tracing 도구가 없거나 다르므로 `dtrace`, `dtruss`, Activity Monitor, `netstat`, Instruments 같은 대체 표면을 봐야 합니다. 도구 이름보다 중요한 것은 "어느 상태를 보려고 하는가"입니다.

## 현실 시나리오 1: `write()`는 빨랐는데 장애 후 로그가 사라졌다

애플리케이션이 로그 파일에 쓴 뒤 성공을 받았고, 몇 초 뒤 전원이 나갔다고 합시다. 재부팅 후 마지막 로그 일부가 없습니다. 이때 "write가 거짓말했다"가 아니라, 어떤 성공을 말했는지 분리해야 합니다.

```
app write()
  -> page cache accepted bytes
  -> write() returned count
  -> dirty page not yet on stable storage
  -> crash
  -> after reboot: last bytes may be absent
```

해결은 목적에 따라 다릅니다. 모든 record를 강하게 보존해야 하면 `fsync()` 빈도를 늘리거나, log record를 replication/quorum으로 보완하거나, storage와 filesystem 보장을 확인해야 합니다. 하지만 매 요청마다 `fsync()`하면 latency와 throughput이 크게 나빠질 수 있습니다. Kafka와 Cassandra가 durability 설정을 제공하는 이유는 이 tradeoff를 workload에 맞게 선택해야 하기 때문입니다.

## 현실 시나리오 2: CPU는 낮은데 p99 latency가 오른다

CPU 사용률이 낮으면 많은 사람이 "CPU 병목은 아니다"라고 말합니다. 하지만 thread가 runnable queue에서 짧게짧게 기다리거나, socket read에서 block되거나, disk writeback과 fsync를 기다리거나, JVM lock을 기다리면 CPU 사용률은 낮아도 p99 latency는 커질 수 있습니다.

이때 질문은 "CPU가 몇 퍼센트인가"가 아니라 "요청 thread가 시간을 어디서 보냈는가"입니다. thread dump로 Java thread 상태를 보고, `vmstat`로 runnable queue와 context switch를 보고, `iostat`으로 disk wait를 보고, `ss`로 socket queue를 봅니다. 하나의 숫자로 결론을 내지 않고, 요청이 user code, kernel syscall, I/O wait, scheduler wait 중 어디에 있었는지 좁혀 가야 합니다.

## 문서를 덮고 확인할 것

- `write()`의 반환값, page cache dirty 상태, `fsync()` 완료, Kafka replica ack는 각각 어떤 성공인가?
- interrupt, trap, syscall을 "누가 왜 커널 진입을 만들었는가" 기준으로 구분해 보세요.
- thread를 늘렸는데 latency가 늘어나는 경로를 CPU, lock, memory, I/O로 나눠 설명해 보세요.
- Cassandra compaction과 Spark shuffle spill이 OS page cache와 disk bandwidth를 두고 경쟁하는 이유를 설명해 보세요.
- macOS에서 Linux 실험 명령이 실패했을 때, 개념 실패와 관찰 도구 실패를 어떻게 구분할지 말해 보세요.

## 근거와 더 읽을 자료

- Linux man-pages: `write(2)`, `open(2)`, `fsync(2)`, `sendfile(2)`, `epoll_wait(2)`.
- Linux kernel documentation: page cache, VFS/filesystem, scheduler, networking.
- Apache Kafka design docs: filesystem/page cache, batching, zero-copy transfer.
- Apache Cassandra storage-engine docs: commit log, memtable, SSTable, commitlog sync.
- Apache Spark cluster and RDD programming guides: driver/executor, task, partition, shuffle, persistence.
