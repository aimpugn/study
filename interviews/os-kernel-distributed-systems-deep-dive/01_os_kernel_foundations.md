# 01. OS Kernel Foundations

## 이 문서가 답하는 질문

운영체제 커널을 공부할 때 가장 먼저 잡아야 하는 질문은 "`write()`가 어떻게 동작하는가"가 아닙니다.
`write()`는 좋은 입구일 뿐입니다.
이 문서가 답하려는 질문은 더 큽니다.

**OS kernel은 여러 프로그램이 CPU, 메모리, 파일, 네트워크 장치를 안전하게 함께 쓰도록 보호하고, 애플리케이션의 요청을 커널 내부 상태 변화와 하드웨어 작업으로 바꾸는 실행 기반**입니다.

이 문서를 읽고 나면 다음을 설명할 수 있어야 합니다.

- 왜 애플리케이션은 하드웨어와 커널 자료구조를 직접 만지지 못하는가
- user mode와 kernel mode의 경계가 무엇을 막고 무엇을 허용하는가
- process가 실행된다는 말이 CPU register, 주소 공간, 열린 파일, scheduler 상태와 어떻게 연결되는가
- syscall, trap, interrupt가 CPU 실행 흐름을 어떻게 커널로 바꾸는가
- virtual memory, page cache, socket buffer, run queue 같은 보이지 않는 상태가 왜 latency와 장애 증상으로 올라오는가
- Kafka, Cassandra, Spark의 증상을 제품 이름이 아니라 OS 자원 이동으로 어떻게 다시 읽는가

이 파일에서 `write(fd, buf, len)`은 중심 주제가 아니라 대표 관찰 예시입니다.
`write()` 하나만 따라가도 커널이 맡는 일을 한 번에 볼 수 있습니다.

- 권한 경계: user mode code가 kernel mode 작업을 직접 하지 못합니다.
- 객체 변환: `fd` 숫자는 커널의 open file state로 해석됩니다.
- 상태 이동: user buffer의 byte는 page cache, dirty page, writeback 경로로 이동할 수 있습니다.
- 완료 경계: `write()` 성공, `fsync()` 완료, storage 내구성은 서로 다른 지점입니다.

## 목차

- [이 문서가 답하는 질문](#이-문서가-답하는-질문)
- [커널을 읽을 때 붙잡을 네 가지 축](#커널을-읽을-때-붙잡을-네-가지-축)
- [1. 왜 커널이라는 중재자가 필요한가](#1-왜-커널이라는-중재자가-필요한가)
- [2. Process의 첫 상태는 어떻게 만들어지는가](#2-process의-첫-상태는-어떻게-만들어지는가)
- [3. User Mode와 Kernel Mode는 이름이 아니라 권한 경계다](#3-user-mode와-kernel-mode는-이름이-아니라-권한-경계다)
- [4. Interrupt, Trap, Syscall은 CPU 흐름을 어떻게 꺾는가](#4-interrupt-trap-syscall은-cpu-흐름을-어떻게-꺾는가)
- [5. `write()`는 커널의 객체 변환을 보여 주는 작은 trace다](#5-write는-커널의-객체-변환을-보여-주는-작은-trace다)
- [6. Scheduler는 CPU 시간을 나누고 기다림을 만든다](#6-scheduler는-cpu-시간을-나누고-기다림을-만든다)
- [7. Virtual Memory는 보호된 주소 공간을 만든다](#7-virtual-memory는-보호된-주소-공간을-만든다)
- [8. File I/O와 Page Cache는 빠른 반환과 내구성 경계를 나눈다](#8-file-io와-page-cache는-빠른-반환과-내구성-경계를-나눈다)
- [9. Network Stack과 Socket Buffer는 request의 첫 대기 큐(queue)다](#9-network-stack과-socket-buffer는-request의-첫-대기-큐queue다)
- [10. Lock, Queue, Observability는 증상을 kernel 질문으로 바꾼다](#10-lock-queue-observability는-증상을-kernel-질문으로-바꾼다)
- [11. Kafka, Cassandra, Spark 증상을 OS 자원 이동으로 다시 읽기](#11-kafka-cassandra-spark-증상을-os-자원-이동으로-다시-읽기)
- [면접에서 바로 꺼내는 짧은 답변](#면접에서-바로-꺼내는-짧은-답변)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [더 깊게 읽을 상세 문서](#더-깊게-읽을-상세-문서)
- [근거와 더 읽을 자료](#근거와-더-읽을-자료)

## 커널을 읽을 때 붙잡을 네 가지 축

커널 설명은 용어가 많아서 금방 흩어집니다.
처음에는 아래 네 가지 축으로 모든 개념을 다시 배치하면 됩니다.

- 보호: 한 프로세스가 다른 프로세스, 커널, 장치를 마음대로 망치지 못하게 막습니다.
- 중재: CPU 시간, 물리 메모리, 파일, 네트워크 장치처럼 공유되는 자원을 나눕니다.
- 변환: 애플리케이션의 함수 호출과 주소, 파일 번호, byte를 커널 객체와 queue, cache, buffer 상태로 바꿉니다.
- 관측: "느리다", "유실됐다", "CPU는 낮다" 같은 증상을 run queue, page fault, dirty page, socket buffer, lock wait 같은 더 낮은 상태로 좁힙니다.

```text
application process
  user code
    |
    +-- ordinary function call
    |     same process memory and registers
    |
    +-- syscall / trap / fault
          |
          v
kernel
  process table / thread state
  address space / page table
  file descriptor table / open file state
  page cache / socket buffer / device queue
  scheduler run queue / wait queue
          |
          v
hardware
  CPU / MMU / memory / storage / NIC
```

이 그림에서 중요한 점은 애플리케이션이 "OS 위에서 실행된다"는 말의 실제 의미입니다.
애플리케이션은 혼자 CPU를 소유하지 않고, 물리 메모리를 직접 나눠 갖지 않으며, 파일과 네트워크 장치를 자기 변수처럼 조작하지 않습니다.
프로세스는 커널이 정한 규칙 안에서 실행되고, 커널은 그 요청을 내부 객체와 하드웨어 작업으로 바꿉니다.

## 1. 왜 커널이라는 중재자가 필요한가

초기 컴퓨터는 비싼 공용 장비였고, 여러 사용자가 같은 기계를 함께 써야 했습니다.
현대 서버도 같은 문제를 더 큰 규모로 반복합니다.
한 머신 안에는 수십 개의 process와 수백 개의 thread가 있고, JVM, 데이터베이스, 로그 수집기, sidecar, kernel thread가 같은 CPU와 메모리, 파일시스템, 네트워크 장치를 함께 씁니다.

여기서 운영체제가 풀어야 하는 문제는 단순합니다.
다만 각 문제에는 이미 커널이 사용하는 대표 해법이 붙어 있습니다.

- 누가 CPU를 얼마나 오래 쓸 수 있는가
    - scheduler: 실행 가능한 thread 중 다음에 CPU를 받을 대상을 고릅니다.
    - timer interrupt: 현재 실행 중인 thread를 주기적으로 끊어 scheduler가 개입할 기회를 만듭니다.
    - run queue: CPU를 받을 준비가 된 task들이 기다리는 큐(queue)입니다.
    - context switch: 현재 task의 실행 상태를 저장하고 다른 task의 실행 상태를 CPU에 올립니다.
- 어느 프로세스가 어느 메모리를 볼 수 있는가
    - virtual memory: process마다 독립된 주소 공간을 가진 것처럼 보이게 합니다.
    - page table: virtual address가 어느 physical page로 이어지는지 기록합니다.
    - MMU: CPU가 memory address를 사용할 때 page table을 기준으로 주소 변환과 권한 검사를 수행합니다.
    - page fault: 필요한 mapping이 없거나 권한이 맞지 않을 때 커널로 들어가게 만드는 trap입니다.
- 누가 어떤 파일과 장치에 접근할 수 있는가
    - user mode/kernel mode 경계: 애플리케이션이 장치와 커널 자료구조를 직접 만지지 못하게 막습니다.
    - permission: user, group, capability, file mode를 기준으로 접근 가능 여부를 판단합니다.
    - file descriptor table: process 안의 작은 정수 `fd`를 커널의 열린 파일 상태로 연결합니다.
    - VFS: 여러 filesystem과 file-like object를 공통 파일 API로 다루게 합니다.
    - device driver: 특정 장치와 실제로 통신하는 커널 코드를 제공합니다.
- 장치에서 이벤트가 발생하면 어떤 실행 흐름을 깨워야 하는가
    - interrupt: 장치가 CPU에게 처리할 일이 생겼다고 알립니다.
    - NAPI/softirq: network packet 처리를 interrupt handler 안에서 모두 끝내지 않고 뒤쪽 경로에서 나누어 처리합니다.
    - wait queue: 특정 조건을 기다리는 task들을 커널 안에 묶어 둡니다.
    - scheduler wakeup: 기다리던 조건이 만족된 task를 다시 runnable 상태로 바꿉니다.
- 빠르게 반환된 작업이 실제로는 어디에 남아 있는가
    - page cache: file page를 kernel memory에 들고 있어 read/write를 빠르게 만듭니다.
    - dirty page: memory에는 최신 내용이 있지만 storage에는 아직 완전히 반영되지 않았을 수 있는 page입니다.
    - socket buffer: socket으로 오가야 할 byte를 kernel 안에서 잠시 기다리게 하는 queue입니다.
    - device queue: storage나 NIC 같은 장치가 처리할 작업이 줄 서 있는 자리입니다.
    - writeback: dirty page를 나중에 storage 쪽으로 밀어내는 커널 작업입니다.
    - `fsync()`: 애플리케이션이 page cache에 남은 변경 data를 storage 쪽에 더 강하게 반영해 달라고 기다리는 요청입니다.

이 문제 때문에 kernel은 application 옆의 helper library가 아니라 더 높은 권한을 가진 중재자가 됩니다.
라이브러리 함수는 보통 현재 process 안에서 실행됩니다.
커널은 CPU 권한, 주소 공간, 장치 접근, process lifecycle을 바꿀 수 있는 별도의 실행 영역입니다.

**커널 경계는 함수 이름의 경계가 아니라 권한, 주소 접근, 커널 자료구조 접근이 바뀌는 지점**입니다.
이 경계를 이해하면 `syscall`, `page fault`, `context switch`, `page cache`, `socket buffer` 같은 단어가 서로 따로 놀지 않습니다.
모두 애플리케이션 요청이 커널의 공유 자원 관리와 만나는 자리입니다.

## 2. Process의 첫 상태는 어떻게 만들어지는가

**process는 실행 중인 프로그램을 커널이 관리하기 위해 만든 자원 묶음**입니다.
프로세스에는 보통 아래 상태가 붙습니다.

- 주소 공간: code, heap, stack, mmap 영역이 놓이는 가상 주소 공간
- 실행 상태: instruction pointer, register, stack pointer처럼 CPU가 이어서 실행할 값
- 열린 객체: file descriptor table, socket, pipe 같은 handle
- 권한과 소속: user id, group, namespace, cgroup, capability
- scheduling 상태: running, runnable, sleeping, stopped 같은 실행 가능성

`java -jar broker.jar` 같은 명령을 실행하면 갑자기 JVM 객체가 생기는 것이 아닙니다.
먼저 OS process가 만들어지고, 그 process 안에서 JVM runtime이 시작됩니다.

```text
power on
  -> firmware prepares minimal hardware state
  -> boot loader loads kernel
  -> kernel initializes memory, interrupt, scheduler, drivers
  -> first user-space process starts
  -> service manager or shell starts a program
  -> execve("/usr/bin/java", argv, envp)
  -> kernel builds a new process image
  -> dynamic linker and runtime startup
  -> JVM starts application code
```

**`execve()`는 현재 process의 사용자 주소 공간을 새 program image로 바꾸는 system call**입니다.
Linux의 ELF 같은 실행 파일 포맷은 기계어 byte만 담지 않습니다.
어느 segment를 메모리에 올릴지, entry point가 어디인지, dynamic linker가 필요한지 같은 정보를 함께 담습니다.
커널은 이 정보를 읽어 새 주소 공간을 만들고, stack에 argument와 environment를 놓고, CPU가 시작할 instruction pointer를 맞춘 뒤 user mode로 돌아갑니다.

이 지식은 JVM, Kafka, Cassandra, Spark에도 그대로 연결됩니다.
JVM은 OS를 우회하는 마법이 아니라 하나의 native process 안에서 실행되는 runtime입니다.
그래서 file descriptor limit, signal 처리, cgroup memory limit, native library loading, process working directory 같은 OS 상태가 애플리케이션 시작과 장애 대응에 직접 영향을 줍니다.

## 3. User Mode와 Kernel Mode는 이름이 아니라 권한 경계다

CPU는 실행 중인 코드에 권한 수준을 부여합니다.
운영체제마다 세부 이름과 아키텍처 표현은 다를 수 있지만, 일반적인 학습 모델에서는 user mode와 kernel mode를 먼저 잡으면 됩니다.

**user mode는 애플리케이션 코드가 제한된 권한으로 실행되는 상태**입니다.
user mode에서 실행되는 애플리케이션이

- 할 수 있는 일
    - 자기에게 매핑된 가상 주소 읽기와 쓰기
    - 산술 명령, 분기 명령, 일반 함수 호출 실행
    - 커널이 허용한 system call 진입점으로 요청 보내기
- 할 수 없는 일
    - page table을 마음대로 바꾸기
    - 디스크 컨트롤러나 NIC에 직접 명령 보내기
    - 다른 프로세스의 file descriptor table 열어 보기
    - 커널 메모리와 커널 내부 자료구조 직접 수정하기

**kernel mode는 커널 코드가 더 높은 권한으로 CPU, 메모리 관리 구조, 장치, 커널 자료구조에 접근할 수 있는 실행 상태**입니다.
커널 모드가 강력한 이유는 단순히 "OS 코드라서"가 아닙니다.
CPU가 그 코드에게 더 넓은 명령과 주소 접근을 허용하기 때문입니다.

이 경계가 없으면 한 프로세스의 버그가 곧바로 전체 머신의 버그가 됩니다.
잘못된 pointer 하나가 다른 프로세스의 memory를 덮어쓰고, 평범한 애플리케이션이 디스크 장치 queue를 망가뜨리고, 네트워크 카드 설정을 바꾸는 일이 가능해집니다.
커널은 이런 위험을 막기 위해 모든 공유 자원 접근을 자신의 진입점으로 모읍니다.

## 4. Interrupt, Trap, Syscall은 CPU 흐름을 어떻게 꺾는가

프로세스가 실행 중이라는 말은 CPU가 그 process의 instruction pointer와 register 값을 기준으로 명령을 실행하고 있다는 뜻입니다.
그런데 CPU는 사용자 코드를 항상 다음 줄로만 실행하지 않습니다.
다음처럼 실행 흐름을 바꾸는 이벤트가 생길 수 있습니다.

- 네트워크 카드가 packet을 받으면 interrupt 발생
- 잘못된 메모리 주소를 접근하면 page fault라는 trap 발생
- `write()` wrapper는 의도적으로 syscall trap 발생
- timer interrupt 뒤에 scheduler가 다른 thread를 고를 수 있음
- signal이 pending 상태였다가 user mode 복귀 시점에 handler로 이어질 수 있음

**interrupt는 외부 장치가 CPU에게 "처리할 일이 있다"고 알리는 비동기 이벤트**입니다.
현재 process가 계산 중이든, JVM bytecode를 실행 중이든, timer나 NIC 같은 장치가 interrupt를 만들 수 있습니다.
커널은 interrupt handler로 들어가 최소한의 처리를 하고, 더 무거운 일은 softirq, kernel thread, scheduler wakeup 같은 뒤쪽 경로로 넘길 수 있습니다.

**trap은 현재 실행 중인 instruction 때문에 생기는 동기적 커널 진입**입니다.
page fault는 대표적인 trap입니다.
CPU가 어떤 가상 주소를 실제 물리 메모리로 번역하려고 했는데 page table에 mapping이 없거나 권한이 맞지 않으면, 커널이 fault handler에서 처리합니다.
정상적인 lazy allocation일 수도 있고, 잘못된 주소 접근이라 process를 종료해야 할 수도 있습니다.

**syscall은 애플리케이션이 커널 기능을 쓰기 위해 의도적으로 만든 trap**입니다.
파일을 열고, socket을 만들고, memory를 mapping하고, process를 만들고, clock을 읽는 많은 작업이 syscall을 통해 커널에 요청됩니다.

```text
ordinary function call
  user function -> user function -> return

syscall
  user code -> syscall entry -> kernel handler -> return to user code

page fault
  load/store instruction -> fault handler -> map page or deliver signal

device interrupt
  NIC/timer/storage -> interrupt handler -> kernel work -> maybe wake a task

timer interrupt
  timer -> interrupt handler -> scheduler may switch task
```

이 이벤트들은 모두 "실행 흐름이 직선이 아니다"라는 점을 보여 줍니다.
하지만 원인은 다릅니다.
interrupt는 외부 장치가 만들고, page fault는 현재 instruction이 만들며, syscall은 현재 thread가 요청합니다.
면접에서 비교할 때는 "누가 왜 커널 진입을 만들었는가"를 기준으로 나누면 흔들리지 않습니다.

## 5. `write()`는 커널의 객체 변환을 보여 주는 작은 trace다

`write(fd, buf, len)`은 커널 경계와 객체 변환을 한 번에 보여 주는 대표 trace입니다.
이 trace를 따라가면 애플리케이션의 작은 함수 호출처럼 보이는 일이 어떻게 커널의 권한 경계, descriptor table, file/socket object, cache/buffer 상태 변화로 바뀌는지 볼 수 있습니다.

먼저 비교축을 고정해야 합니다.
핵심 질문은 "호출 뒤 실제로 누가 어떤 상태를 바꾸는가"입니다.

- `strlen()`이나 `memcpy()`
    - 현재 process 안의 memory와 register를 CPU가 읽거나 복사합니다.
    - 보통 kernel-owned shared resource를 바꾸지 않습니다.
- `write(fd, buf, len)`
    - 현재 process는 요청을 만들고, 커널은 `fd`를 kernel object로 해석합니다.
    - 대상이 regular file이면 VFS, filesystem, page cache, writeback 경로가 이어질 수 있습니다.
    - 대상이 socket이면 socket send buffer와 TCP/network stack 경로가 이어집니다.

**file descriptor는 process별 file descriptor table에서 kernel open-file state를 찾기 위한 작은 정수 handle**입니다.
`fd` 자체가 파일도 아니고, 파일 내용도 아닙니다.
커널은 현재 process의 table에서 `fd`를 lookup하고, 그 항목이 가리키는 open file description과 file operation을 확인합니다.

```text
process memory
  fd  = 3
  buf = address of "hello"
  len = 5
       |
       | libc write wrapper arranges syscall number and arguments
       v
CPU syscall entry
  user mode -> kernel mode
       |
       v
kernel syscall dispatch
  current process fd table
    fd 3 -> open file description
            offset / flags / file operations
       |
       v
target-specific kernel path
  regular file -> VFS -> filesystem -> page cache
  socket       -> socket layer -> TCP -> send buffer
  pipe         -> pipe buffer
```

regular file write라면 흐름은 더 구체적으로 이렇게 볼 수 있습니다.

```text
user buffer bytes
  -> syscall boundary
  -> validate fd and user address
  -> copy bytes into kernel-managed file path
  -> page cache page becomes dirty
  -> write() returns accepted byte count
  -> later writeback or fsync pushes data toward storage
```

이 trace에서 커널의 역할은 세 가지입니다.

- 검증: `fd`가 유효한지, 쓰기 권한이 있는지, user buffer 주소를 읽을 수 있는지 확인합니다.
- 변환: 숫자 `fd`를 open file state로, user address를 kernel copy path로, byte sequence를 page cache update로 바꿉니다.
- 지연 관리: `write()` 반환 뒤에도 dirty page와 device queue에 일이 남을 수 있습니다.

`write()`의 반환값은 요청한 길이와 같을 수도 있고 더 작을 수도 있습니다.
반환값이 `len`보다 작으면 partial write입니다.
견고한 코드는 남은 byte를 다시 써야 합니다.
또 잘못된 user buffer는 `EFAULT`, 닫힌 fd나 잘못된 권한은 `EBADF` 같은 오류로 이어질 수 있습니다.

> `write()`가 성공했으면 디스크에 저장된 건가요?
>
> 아닐 수 있습니다. 일반적인 buffered regular-file I/O에서 `write()` 성공은 보통 커널이 해당 `fd`의 쓰기 요청을 받아들이고, 파일의 해당 부분을 page cache의 dirty page로 반영했다는 뜻에 가깝습니다.
>
> 여기서 파일 쓰기 경로는 `fd -> open file description -> VFS -> filesystem -> page cache`로 이어지는 커널 내부 처리 흐름입니다. **page cache는 파일 내용을 메모리 page 단위로 들고 있는 커널의 캐시**입니다. 따라서 "파일 쓰기 경로를 지났다"와 "page cache에 반영됐다"는 서로 동떨어진 단계가 아니라, 같은 buffered file write 경로 안의 앞뒤 처리로 보면 됩니다.
>
> 저장 장치에 더 강하게 반영하고 싶으면 애플리케이션은 `fsync(fd)`를 따로 호출합니다. **`fsync(fd)`는 그 파일의 dirty data와 파일을 다시 읽는 데 필요한 metadata를 저장 장치 쪽으로 내려보내고, 장치가 완료를 보고할 때까지 기다리는 요청**입니다. 새 파일 생성이나 `rename()`처럼 파일 이름표가 바뀌는 작업은 부모 directory entry 내구성도 따로 봐야 할 수 있습니다.

## 6. Scheduler는 CPU 시간을 나누고 기다림을 만든다

**scheduler는 runnable 상태의 task 중 다음에 CPU를 받을 대상을 고르는 커널 구성요소**입니다.
CPU core 하나는 한 순간에 하나의 실행 흐름만 실제로 실행합니다.
core 수보다 runnable thread가 많으면 누군가는 기다립니다.

**thread는 CPU가 실제로 실행하는 흐름**입니다.
같은 process의 thread들은 주소 공간과 file descriptor table을 공유할 수 있지만, 각 thread는 자기 register 상태, stack, scheduling state를 따로 가집니다.

```text
Thread A running
  -> timer interrupt or blocking syscall
  -> kernel saves A's CPU state
  -> scheduler selects Thread B
  -> kernel restores B's CPU state
  -> Thread B runs
```

여기서 흔한 오해가 생깁니다.
CPU 사용률이 낮으면 CPU가 항상 여유롭다고 생각하기 쉽습니다.
하지만 thread가 runnable queue에서 기다리거나, futex wait에서 잠들거나, disk I/O와 network I/O를 기다리거나, GC safepoint에 묶이면 CPU 사용률은 낮아도 request latency는 커질 수 있습니다.

요청 하나를 CPU 관점으로 보면 다음 흐름이 됩니다.

```text
packet or event arrives
  -> kernel changes socket or wait state
  -> application thread becomes runnable
  -> thread waits in run queue
  -> scheduler gives CPU
  -> user code runs
  -> thread may block again on lock, disk, network, timer, GC
```

Kafka broker의 network thread, Cassandra의 compaction thread, Spark executor의 task thread는 모두 OS scheduler 위에서 CPU를 나눠 씁니다.
그래서 thread 수를 늘리는 것은 항상 정답이 아닙니다.
CPU-bound 작업이 core 수보다 많으면 context switch와 cache disruption이 늘고, I/O-bound 작업이 많으면 run queue보다 device queue와 lock wait가 중요할 수 있습니다.

## 7. Virtual Memory는 보호된 주소 공간을 만든다

**virtual memory는 process가 보는 주소를 물리 메모리 주소와 분리해, 각 process에게 독립된 주소 공간처럼 보이게 하는 메커니즘**입니다.
같은 `0x1000`이라는 virtual address라도 process A와 process B에서는 서로 다른 physical frame을 가리킬 수 있습니다.

CPU 안의 MMU(memory management unit)는 virtual address를 page table을 통해 physical frame으로 번역합니다.
자주 쓰는 번역은 TLB(translation lookaside buffer)에 cache될 수 있습니다.
page table에 mapping이 없거나 권한이 맞지 않으면 page fault가 발생합니다.

```text
user instruction reads virtual address X
  -> MMU checks TLB
  -> if needed, walks page table
  -> mapping exists and permission allows access
       -> physical memory access
  -> mapping missing or permission mismatch
       -> page fault trap
       -> kernel decides what to do
```

page fault는 항상 버그가 아닙니다.

- 아직 실제 page를 할당하지 않았던 stack이나 heap page를 커널이 그때 할당할 수 있습니다.
- `mmap()`으로 mapping한 file page를 첫 접근 시점에 읽어 올 수 있습니다.
- fork 뒤 copy-on-write page를 한쪽 process가 쓰려고 할 때 새 page로 복사할 수 있습니다.
- 정말 잘못된 주소나 권한 위반이면 signal을 보내 process를 종료할 수 있습니다.

분산 시스템을 공부할 때도 이 층은 계속 등장합니다.
Spark executor heap, off-heap buffer, Python worker, shuffle spill page cache는 결국 같은 cgroup memory limit 안에서 경쟁할 수 있습니다.
Cassandra의 memtable, Bloom filter, key cache, OS page cache도 같은 physical memory를 나눠 씁니다.
Kafka는 heap을 무작정 키우면 log segment를 cache할 page cache 여유가 줄어들 수 있습니다.

메모리 질문을 받으면 "어느 메모리입니까?"라고 다시 나눠야 합니다.
Java heap인지, native memory인지, page cache인지, socket buffer인지, page table인지, cgroup limit인지에 따라 관측과 해결이 달라집니다.

## 8. File I/O와 Page Cache는 빠른 반환과 내구성 경계를 나눈다

파일 I/O에서 커널은 path, file descriptor, inode, page cache, block device queue 같은 여러 객체를 연결합니다.
`open("a.log")`는 path string을 파일 객체로 바꾸고, `write(fd, ...)`는 fd table의 항목을 따라 열린 파일 상태와 file operation을 찾습니다.

**VFS는 다양한 filesystem과 file-like object를 공통 파일 API로 다루게 해 주는 커널 계층**입니다.
VFS 자체가 데이터를 영구 저장하는 것은 아닙니다.
VFS는 regular file, socket, pipe, device file 같은 대상에 맞는 구현으로 요청을 넘깁니다.

regular file write의 기본 흐름은 아래처럼 볼 수 있습니다.

```text
regular file write
  -> fd lookup
  -> VFS
  -> filesystem
  -> page cache dirty page
  -> writeback
  -> block layer queue
  -> device driver
  -> storage
```

**dirty page는 메모리 안에는 최신 파일 내용이 있지만 저장 장치에는 아직 완전히 반영되지 않았을 수 있는 page**입니다.
이 상태는 성능에는 유리합니다.
애플리케이션은 매번 물리 장치가 끝날 때까지 기다리지 않고 빠르게 다음 일을 할 수 있습니다.
하지만 내구성 질문에서는 위험한 오해를 만듭니다.

파일 내용과 파일 이름표도 분리해야 합니다.

| 바꾸는 것 | 예 | 보통 확인해야 하는 동기화 대상 |
| --- | --- | --- |
| 파일 안의 byte | 기존 파일에 내용 추가 | file `fd`에 `fsync(fd)` 또는 `fdatasync(fd)` |
| 파일을 찾는 이름표 | 새 파일 생성, `rename("tmp", "log")` | 부모 directory의 file descriptor에 `fsync(dir_fd)` |

**directory entry는 디렉터리 안의 이름이 어떤 파일을 가리키는지 기록하는 이름표**입니다.
`fsync(file_fd)`는 파일 내용과 필요한 파일 metadata를 더 강하게 밀어내는 요청이지, 항상 부모 디렉터리의 이름표 변경까지 보장한다는 뜻은 아닙니다.

```text
write tmp file
  -> fsync(tmp_file_fd)       # file content and required file metadata
  -> rename("tmp", "log")     # directory entry change
  -> fsync(parent_dir_fd)     # name-to-file durability
```

`fsync()` 아래에는 filesystem journal, device write cache, controller, virtualized storage, network filesystem이 있을 수 있습니다.
그래서 운영 환경에서는 filesystem, mount option, storage, replication layer까지 확인해야 합니다.
하지만 학습 단계에서는 우선 `write()` 성공, dirty page, writeback, `fsync(file_fd)`, `fsync(parent_dir_fd)`를 서로 다른 지점으로 분리하면 됩니다.

## 9. Network Stack과 Socket Buffer는 request의 첫 대기 큐(queue)다

분산 시스템에서 request는 application method 안에 갑자기 생기지 않습니다.
NIC가 frame을 받고, driver와 kernel network stack이 packet을 처리하고, TCP가 byte stream을 조립하고, socket receive buffer에 data를 넣고, application thread가 CPU를 받은 뒤 `read()`를 호출해야 합니다.

**socket buffer는 kernel이 socket마다 들고 있는 송수신 byte 큐(queue)**입니다.
sender의 `send()`가 성공해도 byte가 상대 process의 변수에 들어갔다는 뜻은 아닙니다.
receiver의 NIC가 packet을 받아도 application thread가 바로 request handler를 실행했다는 뜻도 아닙니다.

```text
inbound request
  NIC receives frame
    -> interrupt or NAPI polling
    -> kernel network stack
    -> TCP reassembles byte stream
    -> socket receive buffer
    -> epoll/select/poll reports readiness
    -> scheduler runs application thread
    -> read() copies bytes to user buffer
    -> application parser builds request

outbound response
  application write()/send()
    -> socket send buffer
    -> TCP flow control and congestion control
    -> qdisc / NIC TX queue
    -> network
```

이 구조 때문에 "네트워크가 느리다"는 말은 너무 넓습니다.
실제로는 아래 중 하나일 수 있습니다.

- NIC queue나 softirq 처리가 밀립니다.
- socket receive buffer는 찼지만 application thread가 CPU를 못 받습니다.
- sender의 socket send buffer가 receiver window나 congestion 때문에 비지 않습니다.
- TLS나 compression이 network 문제가 아니라 CPU 문제로 바뀝니다.
- event loop가 blocking 작업에 막혀 readiness event를 늦게 처리합니다.

Kafka consumer lag, Cassandra read timeout, Spark shuffle fetch failure는 모두 network metric처럼 보일 수 있습니다.
하지만 원인은 socket queue, scheduler delay, GC pause, disk wait, partition skew, downstream backpressure와 함께 봐야 합니다.

## 10. Lock, Queue, Observability는 증상을 kernel 질문으로 바꾼다

커널과 runtime은 공유 상태를 보호하기 위해 lock과 wait queue를 씁니다.
애플리케이션도 같은 문제를 반복합니다.
Kafka request queue, Cassandra memtable update path, Spark scheduler metadata는 여러 thread가 동시에 접근할 수 있는 상태를 가집니다.

**race는 여러 실행 흐름이 같은 상태를 예상하지 못한 순서로 읽고 써서 결과가 달라지는 문제**입니다.
lock은 이런 문제를 막지만, contention이 생기면 기다림을 만듭니다.

```text
without lock
  T1 read count = 10
  T2 read count = 10
  T1 write count = 11
  T2 write count = 11
  result: lost update

with lock
  T1 acquire lock
  T1 read/write count = 11
  T1 release lock
  T2 acquire lock
  T2 read/write count = 12
  T2 release lock
```

기다림은 여러 모양으로 보입니다.

- Java thread dump에서는 `BLOCKED`, `WAITING`, `TIMED_WAITING`, `RUNNABLE`로 보일 수 있습니다.
- Linux 관점에서는 futex wait, disk sleep, socket wait, scheduler delay로 보일 수 있습니다.
- CPU flame graph에는 CPU에서 실행한 시간만 보이고, off-CPU wait는 따로 봐야 할 수 있습니다.

관측 도구를 외우는 것보다 중요한 질문은 "어떤 상태를 보려는가"입니다.

| 증상 | 성급한 결론 | kernel 질문 |
| --- | --- | --- |
| CPU 사용률 낮음 | CPU는 문제가 아니다 | thread가 run queue, lock, I/O, GC, cgroup throttle 중 어디서 기다리는가 |
| disk latency 증가 | 디스크만 느리다 | dirty page, writeback, fsync, block queue, compaction/spill 중 어디가 밀리는가 |
| network timeout | 상대 노드가 죽었다 | retransmission, socket queue, scheduler delay, GC pause, packet loss 중 무엇인가 |
| memory free 감소 | memory leak이다 | heap, off-heap, page cache, mmap, socket buffer, cgroup accounting 중 무엇인가 |
| queue 증가 | worker만 늘리면 된다 | producer가 빠른가, consumer가 느린가, downstream이나 partition skew가 있는가 |

Linux에서는 `/proc`, `/sys`, `strace`, `perf`, eBPF, `vmstat`, `iostat`, `ss`, `tcpdump` 같은 도구를 쓸 수 있습니다.
macOS는 `/proc`이나 Linux eBPF 도구가 그대로 없을 수 있으므로 `dtrace`, `dtruss`, Instruments, Activity Monitor, `netstat` 같은 다른 표면을 봐야 합니다.
도구 차이는 개념 실패가 아니라 관측 표면 차이입니다.

## 11. Kafka, Cassandra, Spark 증상을 OS 자원 이동으로 다시 읽기

이 문서의 목적은 OS 용어를 외우는 데서 끝나지 않습니다.
Kafka, Cassandra, Spark는 모두 OS process이고, 각 product는 OS 자원을 자기 내부 상태와 protocol 의미로 다시 포장합니다.

| 제품 증상 | 제품 내부 표현 | 아래 OS kernel 관점 |
| --- | --- | --- |
| Kafka consumer lag | consumer가 offset을 따라가지 못함 | broker disk/page cache, network socket, consumer CPU, downstream wait |
| Kafka replica lag | follower fetch가 leader log를 따라가지 못함 | disk read/writeback, network send buffer, scheduler delay |
| Cassandra write latency | mutation이 coordinator/replica path에서 늦음 | commit log append/fsync, memtable lock, compaction I/O, network |
| Cassandra read timeout | replica read가 deadline 안에 오지 않음 | page cache miss, SSTable disk I/O, CPU, repair/compaction pressure |
| Spark shuffle spill | memory에 못 담은 shuffle data가 local disk로 밀림 | executor memory pressure, page cache, local disk queue, network fetch |
| Spark straggler | 일부 task가 유독 늦음 | scheduler delay, data skew, disk spill, GC, network wait |

Kafka가 page cache와 batch, sequential append, file-to-socket 전송 경로를 활용한다는 말도 OS foundation 위에서 읽어야 합니다.
각 항목은 "Kafka가 중요하게 본다"로 끝나면 부족합니다.

| Kafka 쪽 표현 | 무엇인가 | 왜 OS kernel foundation과 연결되는가 |
| --- | --- | --- |
| record batch | 여러 record를 한 덩어리로 묶은 단위 | 작은 syscall과 copy를 많이 반복하는 비용을 줄입니다. |
| sequential append | log segment 뒤에 이어 붙이는 쓰기 패턴 | filesystem, page cache, storage가 처리하기 쉬운 흐름을 만듭니다. |
| page cache 활용 | file page를 kernel memory에 올려 재사용하는 방식 | broker heap 밖에서 segment read/write 성능을 좌우합니다. |
| file-to-socket path | file byte를 socket으로 보내는 kernel 경로 | user-space copy를 줄일 수 있지만 TLS, compression, transformation은 경계를 바꿀 수 있습니다. |

이 관점은 세 시스템에 모두 적용됩니다.
Cassandra의 commit log와 SSTable은 file I/O와 compaction pressure로 내려가고, Spark의 shuffle은 memory, local disk, network socket, scheduler로 내려갑니다.
좋은 면접 답변은 product metric을 말한 뒤 "그 아래에서는 어떤 kernel resource가 queue를 만들었는가"로 다시 내려갑니다.

## 면접에서 바로 꺼내는 짧은 답변

> OS kernel은 왜 필요한가요?
>
> 한 머신 안의 CPU, 메모리, 파일, 네트워크 장치는 여러 process가 함께 쓰는 공유 자원입니다. 커널은 user mode와 kernel mode 경계로 잘못된 접근을 막고, system call과 trap을 통해 애플리케이션 요청을 검증한 뒤, page table, scheduler, page cache, socket buffer, device queue 같은 내부 상태로 바꿉니다. 그래서 커널은 단순한 장벽이 아니라 보호자이자 중재자이고, 백엔드 latency와 장애 증상은 이 내부 상태가 어디서 밀렸는지로 다시 해석해야 합니다.
>
> `write()`는 OS kernel foundation에서 어떤 의미인가요?
>
> `write()`는 커널 경계와 객체 변환을 보여 주는 대표 예시입니다. 애플리케이션은 `fd`, user buffer address, length를 넘기고, 커널은 `fd`를 현재 process의 descriptor table에서 open file state로 해석합니다. regular file이면 VFS와 filesystem, page cache dirty page, writeback 경로로 이어지고, socket이면 socket buffer와 TCP 경로로 이어집니다. 이 trace 하나로 syscall, 권한 경계, 커널 객체, cache/buffer, 빠른 반환과 실제 완료의 차이를 볼 수 있습니다.
>
> CPU 사용률이 낮은데 p99 latency가 높으면 무엇을 보겠습니까?
>
> CPU 사용률 하나로 결론을 내리지 않겠습니다. 요청 thread가 CPU에서 실행된 시간과 CPU 밖에서 기다린 시간을 나눠 보겠습니다. run queue delay, lock/futex wait, disk I/O, socket buffer, GC pause, cgroup throttling, downstream queue를 함께 봐야 합니다. CPU가 낮다는 말은 "할 일이 없다"일 수도 있지만, "thread가 CPU를 못 쓰고 다른 조건을 기다린다"일 수도 있습니다.
>
> interrupt, trap, syscall은 어떻게 구분하나요?
>
> 누가 왜 커널 진입을 만들었는지로 구분합니다. interrupt는 NIC나 timer 같은 외부 장치가 비동기적으로 만듭니다. trap은 현재 instruction 실행 중 생긴 이벤트이고, page fault가 대표적입니다. syscall은 애플리케이션이 커널 기능을 쓰려고 의도적으로 만든 trap입니다. 셋 모두 커널 코드로 들어갈 수 있지만 원인과 복귀 흐름이 다릅니다.

## 문서를 덮고 확인할 것

아래 질문에 답할 수 있으면 OS kernel foundation을 단어가 아니라 경로로 이해한 것입니다.

- user mode에서 할 수 있는 일과 할 수 없는 일을 각각 세 가지씩 말해 보세요.
- process가 실행된다는 말을 instruction pointer, register, 주소 공간, file descriptor table, scheduler state로 풀어 보세요.
- `write(fd, buf, len)`에서 `fd`, `buf`, `len`이 각각 kernel 안에서 무엇으로 해석되는지 trace로 그려 보세요.
- `write()` 성공, dirty page, writeback, `fsync(file_fd)`, `fsync(parent_dir_fd)`를 서로 다른 완료 지점으로 구분해 보세요.
- packet 도착부터 application `read()`까지 NIC, interrupt/NAPI, TCP, socket buffer, scheduler를 넣어 그려 보세요.
- Kafka consumer lag, Cassandra read timeout, Spark shuffle spill 중 하나를 골라 OS 자원 대기 후보를 세 가지 이상 말해 보세요.

## 더 깊게 읽을 상세 문서

이 파일은 OS kernel foundation의 허브입니다.
각 항목의 깊이는 아래 문서에서 닫습니다.

| 다음 문서 | 이 파일에서 이어지는 질문 |
| --- | --- |
| [01a_process_scheduling.md](01a_process_scheduling.md) | process, thread, scheduler, run queue, signal, cgroup CPU를 어떻게 봐야 하는가 |
| [01b_memory_and_address_space.md](01b_memory_and_address_space.md) | virtual address, page table, TLB, page fault, mmap, OOM, cgroup memory가 어떻게 연결되는가 |
| [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md) | fd, inode, VFS, page cache, writeback, fsync, block queue, crash consistency를 어떻게 나눌 것인가 |
| [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md) | NIC, interrupt/NAPI, TCP, socket buffer, epoll, send path, qdisc를 어떻게 request latency로 읽을 것인가 |
| [01e_concurrency_isolation_observability.md](01e_concurrency_isolation_observability.md) | race, futex, lock wait, cgroup, namespace, `/proc`, perf/eBPF 관측을 어떻게 장애 추론으로 바꿀 것인가 |

## 근거와 더 읽을 자료

이 문서의 세부 claim은 [10_source_ledger.md](10_source_ledger.md)의 claim-level source ledger와 함께 읽어야 합니다.
운영 적용 전에는 target OS, kernel version, filesystem, storage, container runtime, Kafka/Cassandra/Spark version을 다시 확인해야 합니다.

- Linux man-pages: `write(2)`, `open(2)`, `fsync(2)`, `execve(2)`, `fork(2)`, `clone(2)`, `mmap(2)`, `epoll(7)`, `socket(2)`, `tcp(7)`, `cgroups(7)`, `namespaces(7)`, `proc(5)`
- Linux kernel documentation: page cache, VFS, scheduler, networking/NAPI, cgroup v2, eBPF userspace API
- POSIX.1-2024: file I/O, synchronization, directory modification durability
- Apache Kafka design and documentation: log, partition, batching, page cache, zero-copy transfer, consumer offset
- Apache Cassandra documentation: storage engine, commit log, memtable, SSTable, consistency, hints, compaction
- Apache Spark documentation: driver/executor, RDD/DataFrame, task/stage, shuffle, persistence, checkpoint
