# 01b. Memory, Address Space, Page Fault

## 목차

- [virtual address는 process에게 주는 보호된 착시다](#virtual-address는-process에게-주는-보호된-착시다)
- [TLB는 주소 번역의 page cache에 가깝다](#tlb는-주소-번역의-page-cache에-가깝다)
- [page fault는 오류일 수도 있고 정상 실행 경로일 수도 있다](#page-fault는-오류일-수도-있고-정상-실행-경로일-수도-있다)
- [mmap은 파일을 메모리처럼 보이게 하지만 I/O를 없애지 않는다](#mmap은-파일을-메모리처럼-보이게-하지만-io를-없애지-않는다)
- [heap, off-heap, page cache는 같은 물리 메모리 위에서 경쟁한다](#heap-off-heap-page-cache는-같은-물리-메모리-위에서-경쟁한다)
- [allocator와 slab은 작은 객체를 빠르게 주고받기 위한 계층이다](#allocator와-slab은-작은-객체를-빠르게-주고받기-위한-계층이다)
- [OOM은 "메모리가 0이 됐다"보다 복잡하다](#oom은-메모리가-0이-됐다보다-복잡하다)
- [제품으로 다시 연결하기](#제품으로-다시-연결하기)
- [Page replacement는 어떤 기억을 버릴지 고르는 정책이다](#page-replacement는-어떤-기억을-버릴지-고르는-정책이다)
- [Huge page, TLB pressure, memory locality](#huge-page-tlb-pressure-memory-locality)
- [False sharing과 memory ordering은 같은 변수만의 문제가 아니다](#false-sharing과-memory-ordering은-같은-변수만의-문제가-아니다)
- [mmap과 direct I/O를 cache 우회로만 이해하면 위험하다](#mmap과-direct-io를-cache-우회로만-이해하면-위험하다)
- [Cgroup memory와 JVM memory를 같이 읽는 법](#cgroup-memory와-jvm-memory를-같이-읽는-법)
- [Interview replay: 메모리 문제를 한 문장으로 시작하고 아래로 내려가기](#interview-replay-메모리-문제를-한-문장으로-시작하고-아래로-내려가기)
- [Page table 자체도 메모리를 쓴다](#page-table-자체도-메모리를-쓴다)
- [Reclaim이 latency로 보이는 순간](#reclaim이-latency로-보이는-순간)
- [Garbage collection과 OS memory는 서로 독립적이지 않다](#garbage-collection과-os-memory는-서로-독립적이지-않다)
- [Interview replay: 관측 지표를 계층별로 배치하기](#interview-replay-관측-지표를-계층별로-배치하기)
- [마지막 연결: 메모리의 정답은 한 숫자가 아니다](#마지막-연결-메모리의-정답은-한-숫자가-아니다)

메모리 문제는 "heap을 늘리면 된다"로 끝나지 않습니다. Kafka broker, Cassandra node, Spark executor는 모두 process이고, process는 자기만의 가상 주소 공간을 보는 것처럼 실행됩니다. 하지만 실제 physical memory는 OS page cache, JVM heap, off-heap buffer, mmap file, kernel socket buffer, filesystem metadata, 다른 process와 함께 나눠 씁니다. 이 층을 모르면 GC, OOM, page cache miss, mmap, swap, container memory limit이 서로 다른 단어처럼 보입니다.

이 문서에서는 virtual address가 physical frame으로 번역되는 흐름에서 시작해, page fault, copy-on-write, mmap, allocator, OOM, page cache가 backend system의 증상으로 어떻게 올라오는지 설명합니다.

## virtual address는 process에게 주는 보호된 착시다

process가 포인터 `0x1000`을 읽는다고 해도 CPU가 곧바로 물리 메모리 주소 0x1000을 읽는 것은 아닙니다. CPU의 MMU(memory management unit)는 virtual address를 page table을 통해 physical frame으로 번역합니다. page table은 "이 process의 virtual page가 어느 physical frame에 연결되는가, 읽기/쓰기/실행 권한은 무엇인가"를 담은 구조입니다.

```text
process A virtual address 0x4000
  -> page table A
  -> physical frame 0x9a000

process B virtual address 0x4000
  -> page table B
  -> physical frame 0x3f000
```

같은 숫자 주소가 process마다 다른 물리 page를 가리킬 수 있기 때문에 process isolation이 가능합니다. 한 process가 다른 process의 heap을 그냥 포인터로 읽을 수 없는 이유가 여기 있습니다. user/kernel 구분도 이 주소 변환과 권한 검사 위에 올라갑니다. user process가 kernel-only mapping이나 권한 없는 page를 접근하면 CPU는 instruction을 계속 실행하지 않고 page fault를 발생시켜 커널로 들어갑니다.

## TLB는 주소 번역의 page cache에 가깝다

page table walk는 비용이 큽니다. CPU는 최근 virtual-to-physical 번역 결과를 TLB(translation lookaside buffer)에 캐시합니다. TLB hit가 나면 page table을 다시 걷지 않고 바로 physical frame을 찾을 수 있습니다. TLB miss가 많아지면 같은 CPU 사용률에서도 memory access latency가 커질 수 있습니다.

```text
load [virtual X]
  -> TLB lookup
       hit: physical frame found quickly
       miss: page table walk
  -> permission check
  -> memory access or page fault
```

large page 또는 huge page가 등장하는 이유도 여기서 이해할 수 있습니다. page 크기가 커지면 같은 memory range를 더 적은 TLB entry로 덮을 수 있어 TLB pressure를 줄일 수 있습니다. 대신 memory fragmentation, allocation 실패, NUMA placement, 운영 복잡도가 생깁니다. Cassandra나 JVM 튜닝에서 huge page를 볼 때는 "마법처럼 빨라진다"가 아니라 TLB와 memory mapping 비용을 줄이는 선택으로 봐야 합니다.

## page fault는 오류일 수도 있고 정상 실행 경로일 수도 있다

page fault라는 이름 때문에 많은 사람이 곧바로 crash를 떠올립니다. 하지만 page fault는 "현재 instruction이 접근한 virtual address를 지금 당장 사용할 수 없다"는 CPU의 알림입니다. 커널은 그 이유를 보고 정상 처리할지, process를 죽일지 결정합니다.

```text
user instruction touches virtual address X
  -> TLB/page table cannot provide valid writable/readable mapping
  -> page fault trap
  -> kernel checks VMA and permissions
       stack growth? allocate page
       mmap file page? read or map file-backed page
       copy-on-write? allocate and copy page
       invalid address? send SIGSEGV
```

stack이 커질 때 page를 늦게 할당하는 것도 page fault로 처리될 수 있습니다. `mmap()`으로 파일을 매핑한 뒤 첫 접근에서 파일 page를 가져오는 것도 page fault입니다. `fork()` 직후 부모와 자식이 같은 physical page를 공유하다가 한쪽이 쓰려 할 때 copy-on-write가 발생하는 것도 page fault입니다. 반대로 권한 없는 주소나 해제된 mapping을 접근하면 segmentation fault가 됩니다.

## mmap은 파일을 메모리처럼 보이게 하지만 I/O를 없애지 않는다

`mmap()`은 파일이나 anonymous memory를 process의 address space에 매핑합니다. 파일을 `read()`로 buffer에 복사하는 대신, 파일의 byte range를 virtual address range로 보이게 만들 수 있습니다. 그러나 "mmap은 디스크 I/O가 없다"는 뜻이 아닙니다. 파일-backed mapping은 해당 page를 처음 만질 때 page fault를 통해 page cache와 storage path를 만날 수 있습니다. dirty mmap page도 이후 writeback되어야 합니다.

```text
mmap(file)
  -> virtual range is created
  -> no file byte may be read yet

first load from mapped page
  -> page fault
  -> kernel locates file page
  -> page cache / disk read if needed
  -> map physical page
  -> instruction retries
```

Kafka의 log segment, Cassandra SSTable, Lucene/OpenSearch index, Spark spill file 같은 구조를 볼 때 mmap과 page cache의 관계를 이해하면 "파일인데 왜 memory pressure와 관련 있지?"라는 질문이 풀립니다. file-backed page도 physical memory를 씁니다. 메모리가 부족하면 커널은 clean page cache를 버리거나 dirty page를 writeback하고, anonymous memory는 swap 후보가 될 수 있습니다.

## heap, off-heap, page cache는 같은 물리 메모리 위에서 경쟁한다

JVM process를 볼 때 `-Xmx`만 보면 부족합니다. JVM heap은 process memory의 일부입니다. direct buffer, native library allocation, thread stack, code cache, mmap file, kernel socket buffer, page cache는 heap 밖에서 memory를 씁니다. container memory limit 안에서는 이 모든 것이 같이 계산될 수 있습니다.

```text
physical memory / cgroup memory limit
  -> JVM heap
  -> off-heap direct buffers
  -> thread stacks
  -> mmap regions
  -> kernel memory and socket buffers
  -> page cache
```

Spark executor memory를 크게 잡으면 execution/storage memory는 늘 수 있지만 OS page cache와 off-heap, container headroom이 줄어들 수 있습니다. Cassandra에서 heap을 너무 크게 잡으면 GC pause가 길어지고, 너무 작으면 memtable/cache pressure가 커질 수 있습니다. Kafka broker heap이 record data를 오래 붙잡지 않도록 설계되어도 page cache와 network buffer, compression buffer는 여전히 memory pressure와 관련됩니다.

## allocator와 slab은 작은 객체를 빠르게 주고받기 위한 계층이다

커널과 user-space runtime은 매번 page 단위로만 memory를 할당하지 않습니다. user space에는 malloc 계열 allocator가 있고, 커널에는 slab/slub 같은 allocator가 있습니다. allocator는 작은 객체를 빠르게 할당/해제하기 위해 memory를 미리 나누고 재사용합니다. 이 구조는 빠르지만 fragmentation과 cache pressure를 만들 수 있습니다.

예를 들어 kernel은 socket buffer, inode, dentry, task 구조 같은 객체를 계속 만들고 없앱니다. network connection이 폭증하면 application heap만 늘어나는 것이 아니라 kernel object와 socket buffer, slab cache도 늘 수 있습니다. `slabtop`, `/proc/meminfo`, `/proc/slabinfo` 같은 관측이 필요한 이유입니다.

## OOM은 "메모리가 0이 됐다"보다 복잡하다

Linux에서 memory pressure가 커지면 커널은 먼저 reclaim을 시도합니다. clean page cache는 버릴 수 있고, dirty page는 writeback해야 하며, anonymous page는 swap이 있다면 swap out할 수 있습니다. 그래도 필요한 memory를 만들 수 없으면 OOM killer가 process를 죽일 수 있습니다. cgroup memory limit 안에서는 시스템 전체 memory가 남아 있어도 해당 container/process group이 OOM을 맞을 수 있습니다.

```text
allocation request
  -> free memory insufficient
  -> reclaim clean page cache
  -> writeback dirty page
  -> swap anonymous pages if possible
  -> if still impossible, OOM selection
  -> kill a process or fail allocation path
```

운영에서 OOM을 보면 `free` 숫자 하나로 판단하지 않습니다. page cache가 많은지, reclaim이 가능한지, cgroup limit이 어디인지, kernel memory가 늘었는지, swap이 쓰였는지, OOM killer log가 어떤 process를 골랐는지 봐야 합니다.

## 제품으로 다시 연결하기

- Kafka는 broker heap에 모든 message를 오래 두는 시스템이 아닙니다. 빠른 path는 log segment와 page cache, network buffer를 적극 활용합니다. 그래서 heap만 보지 말고 page cache와 dirty writeback, socket buffer를 같이 봐야 합니다.
- Cassandra는 memtable, Bloom filter, key cache, row cache, off-heap, OS page cache가 함께 움직입니다. SSTable read가 page cache hit를 받으면 빠르지만, compaction과 repair가 memory/cache locality를 흔들 수 있습니다.
- Spark는 executor heap, unified memory, off-heap, broadcast, shuffle buffer, spill file, page cache가 얽힙니다. executor memory를 늘렸는데 더 느려지는 경우에는 GC와 page cache 축소, container overhead를 같이 의심해야 합니다.

문서를 덮고 아래 질문에 답해 보세요.

```text
virtual address
  -> page table / TLB
  -> physical frame or page fault
  -> anonymous/file-backed page
  -> reclaim/writeback/swap/OOM decision
```

이 trace를 설명할 수 있으면 "JVM heap은 남는데 OOM이 났다", "page cache가 커서 free memory가 작다", "mmap 파일을 읽었는데 왜 page fault가 늘었나" 같은 질문을 한 층 아래에서 다시 볼 수 있습니다.

## Page replacement는 어떤 기억을 버릴지 고르는 정책이다

메모리는 CPU register나 cache보다 크지만 storage보다 훨씬 작습니다. 운영체제는 모든 process와 kernel subsystem이 원하는 page를 물리 메모리에 계속 둘 수 없습니다. 그래서 memory pressure가 생기면 어떤 page를 계속 resident로 둘지, 어떤 page를 버리거나 writeback하거나 swap으로 내보낼지 결정해야 합니다. 이 결정이 page replacement입니다. 교과서에서는 LRU, clock, working set 같은 정책을 배우지만, 실무에서 중요한 것은 "자주 쓰는 working set이 memory에 남아 있는가"입니다.

Clean file-backed page는 비교적 쉽게 버릴 수 있습니다. 파일에서 다시 읽으면 되기 때문입니다. Dirty file-backed page는 storage로 내려보낸 뒤 버릴 수 있습니다. Anonymous page는 파일 원본이 없으므로 swap이 있으면 swap out할 수 있고, swap이 없으면 reclaim이 어렵습니다. Kernel slab cache나 socket buffer도 상황에 따라 줄일 수 있지만, 모든 kernel memory가 쉽게 회수되는 것은 아닙니다.

```text
memory pressure
  -> scan candidate pages
  -> keep hot pages if possible
  -> drop clean file cache
  -> write back dirty file pages
  -> swap anonymous pages if configured
  -> shrink reclaimable kernel caches
```

Kafka, Cassandra, Spark는 모두 working set 문제를 만납니다. Kafka consumer가 최근 segment를 반복해서 읽으면 page cache hit가 많아지고 disk I/O가 줄어듭니다. Consumer가 오래된 segment를 넓게 scan하면 hot cache가 밀려날 수 있습니다. Cassandra read path가 Bloom filter와 index summary, key cache, OS page cache를 잘 활용하면 random read 비용이 줄지만, compaction이 큰 SSTable들을 읽고 쓰면서 cache를 흔들 수 있습니다. Spark shuffle job이 spill file을 대량으로 만들면 page cache와 executor memory가 함께 압박받고, 다음 stage가 같은 데이터를 다시 읽을 때 hit/miss가 job runtime을 바꿀 수 있습니다.

Page replacement를 성능 문제로 볼 때는 "free memory가 작다"보다 "reclaim이 누구를 밀어냈는가"를 묻는 편이 낫습니다. Linux는 남는 메모리를 page cache로 적극 활용하므로 free가 작아도 정상일 수 있습니다. 반대로 available memory가 줄고 major fault, swap in/out, direct reclaim, OOM event가 늘면 실제 pressure입니다. JVM heap graph만 보면 heap이 안정적인데도 page cache가 줄고 disk I/O가 늘어 전체 latency가 오르는 경우를 놓칠 수 있습니다.

## Huge page, TLB pressure, memory locality

TLB는 주소 번역 결과를 cache하지만 entry 수가 제한되어 있습니다. Working set이 큰데 page 크기가 작으면 많은 virtual page를 번역해야 하고, TLB miss가 늘 수 있습니다. Huge page는 더 큰 page 크기로 넓은 memory range를 적은 TLB entry로 덮어 TLB pressure를 줄입니다. 그러나 huge page는 공짜가 아닙니다. 큰 contiguous memory가 필요하고, fragmentation과 NUMA placement를 신경 써야 하며, transparent huge page는 workload에 따라 pause나 latency 변동을 만들 수 있습니다.

```text
4 KiB pages
  1 GiB memory range -> many page table entries and TLB entries

2 MiB huge pages
  same range -> fewer entries
  but allocation and placement become more sensitive
```

JVM과 database tuning에서 huge page를 볼 때는 "켜면 빨라진다"가 아니라 "주소 번역 비용을 줄이는 대신 운영 복잡도와 allocation 제약을 받는다"로 이해해야 합니다. Cassandra나 large heap JVM은 TLB miss 감소 이점을 볼 수 있지만, container memory, NUMA, GC, page cache와 함께 검증해야 합니다. Kafka처럼 page cache를 적극 활용하는 workload에서는 heap huge page보다 OS file cache와 disk path가 더 중요한 경우도 많습니다.

NUMA locality도 memory 성능의 일부입니다. Thread가 CPU socket 0에서 실행되는데 주로 socket 1에 붙은 memory를 읽으면 remote access가 늘 수 있습니다. Scheduler와 allocator가 자동으로 locality를 맞추려 하지만, thread migration, container placement, memory policy, long-running process의 heap growth가 어긋나면 latency가 튈 수 있습니다. Spark executor를 큰 heap으로 띄우거나 Cassandra node가 큰 off-heap/cache를 쓰는 경우, NUMA와 GC, compaction thread placement가 함께 영향을 줄 수 있습니다.

## False sharing과 memory ordering은 같은 변수만의 문제가 아니다

동시성 문서에서 더 깊게 다루지만, 메모리 장에서는 cache line을 먼저 붙잡아야 합니다. CPU cache는 byte 하나가 아니라 cache line 단위로 데이터를 가져오고 일관성을 유지합니다. 서로 다른 thread가 서로 다른 변수를 쓰더라도 그 변수가 같은 cache line에 있으면, 두 core가 같은 line ownership을 계속 주고받을 수 있습니다. 이것이 false sharing입니다.

```text
cache line
  [ counterA ][ counterB ][ padding ... ]

core 0 writes counterA
core 1 writes counterB
  -> each write invalidates or transfers the same line
  -> both variables are independent, but hardware treats the line as shared
```

이 문제는 Kafka request metric counter, Cassandra per-table counter, Spark task metric 같은 곳에서도 이론적으로 나타날 수 있습니다. 대부분의 개발자는 source-level variable만 보지만, CPU는 cache line을 봅니다. 그래서 고성능 라이브러리나 runtime은 padding, striped counter, per-thread accumulator를 사용해 contention을 줄입니다. 단, padding을 아무 곳에나 넣는 것은 답이 아닙니다. 실제 contention이 있는지 관측하고, memory footprint 증가와 locality 손실까지 함께 봐야 합니다.

Memory ordering은 CPU와 compiler가 성능을 위해 memory operation의 관측 순서를 최적화할 수 있다는 사실에서 시작합니다. 한 thread가 `data`를 쓰고 `ready=true`를 썼다고 해서, 다른 core에서 항상 그 순서로 보인다고 가정하면 안 되는 환경이 있습니다. Lock, atomic, volatile, memory barrier는 "이 지점 앞뒤의 읽기/쓰기 순서가 다른 thread에게 어떻게 보여야 하는가"를 정하는 장치입니다. Java의 `volatile`과 `synchronized`, C/C++ atomics, kernel memory barrier는 층은 다르지만 같은 질문을 다룹니다.

## mmap과 direct I/O를 cache 우회로만 이해하면 위험하다

`mmap()`은 파일을 process address space에 연결합니다. 애플리케이션은 pointer로 파일 내용을 읽는 것처럼 보지만, page fault와 page cache, writeback이 뒤에서 작동합니다. Direct I/O는 page cache를 우회해 user buffer와 storage 사이를 더 직접 연결하려는 경로입니다. 두 선택 모두 "더 빠른 I/O"라는 단순한 꼬리표로 이해하면 안 됩니다.

`mmap()`은 read/write system call 횟수와 copy를 줄일 수 있고, random access를 pointer access처럼 만들 수 있습니다. 하지만 page fault가 요청 처리 중간에 튈 수 있고, SIGBUS 같은 예외 상황이 생길 수 있으며, dirty page writeback 시점을 application이 직관적으로 보기 어렵습니다. Direct I/O는 page cache pollution을 줄이고 application이 자체 cache를 관리하는 database에는 유리할 수 있지만, alignment 제약과 작은 I/O 비용, readahead/writeback 이점 손실이 있습니다.

```text
buffered read
  storage -> page cache -> copy to user buffer

mmap read
  page fault -> page cache page mapped -> user load reads mapped memory

direct I/O
  user buffer <-> block layer/device path
  page cache mostly bypassed
```

Kafka는 일반적으로 OS page cache와 sequential I/O를 잘 활용하는 방향으로 설계되어 있습니다. Cassandra는 commitlog와 SSTable, key cache, row cache, chunk cache, OS page cache가 조합됩니다. Spark는 shuffle spill과 cache/persist가 workload에 따라 memory와 disk 사이를 오갑니다. 그러므로 "page cache를 쓰는 시스템"과 "자체 cache를 쓰는 시스템"을 이분법으로 나누지 말고, 어떤 데이터가 어느 cache에 있고, 누가 eviction을 결정하며, write durability는 어느 계층에서 닫히는지 봐야 합니다.

## Cgroup memory와 JVM memory를 같이 읽는 법

컨테이너 환경에서 memory 사고는 process RSS 하나로 끝나지 않습니다. cgroup은 process group의 memory usage와 limit을 account합니다. 이 안에는 anonymous memory, file cache, kernel memory accounting 일부, tmpfs, page table 등이 포함될 수 있습니다. JVM의 `-Xmx`는 Java heap 상한일 뿐입니다. Direct buffer, metaspace, code cache, thread stack, native allocation, mmap, page cache, socket buffer는 별도 공간을 씁니다.

```text
container memory limit
  -> Java heap
  -> metaspace / code cache
  -> thread stacks
  -> direct buffers and native memory
  -> mmap regions
  -> page cache charged to cgroup
  -> kernel/socket memory depending on accounting
```

Spark executor에서 흔히 보는 문제는 executor heap은 limit 안에 있는데 container가 killed 되는 경우입니다. Shuffle buffer, off-heap memory, Python worker, Arrow buffer, thread stack, page cache, memory overhead가 합쳐져 limit을 넘을 수 있습니다. Cassandra는 off-heap memtable이나 cache, compression buffer, file cache가 heap 밖에서 커질 수 있습니다. Kafka는 page cache와 network buffer, log segment access가 heap 밖 memory를 중요하게 만듭니다.

관측할 때는 JVM GC log, native memory tracking, `/proc/<pid>/smaps`, cgroup memory events, container runtime OOM reason, kernel log를 함께 봅니다. Java `OutOfMemoryError`는 JVM 내부 allocation 실패이고, Linux OOM kill은 kernel이 process를 죽인 사건입니다. 같은 "메모리 부족"이라도 복구와 증거가 다릅니다. 전자는 heap dump나 exception stack이 남을 수 있고, 후자는 process가 바로 사라져 supervisor log와 kernel/cgroup event를 봐야 할 수 있습니다.

## Interview replay: 메모리 문제를 한 문장으로 시작하고 아래로 내려가기

면접에서 "가상 메모리를 설명해 보세요"라는 질문에는 이렇게 시작할 수 있습니다. "프로세스가 보는 주소는 물리 주소가 아니라 page table을 통해 physical frame으로 번역되는 virtual address입니다. 이 덕분에 process isolation과 lazy allocation이 가능하고, 주소가 아직 실제 page와 연결되지 않았거나 권한이 맞지 않으면 page fault로 kernel에 들어갑니다." 여기까지가 첫 문장입니다.

꼬리 질문이 오면 TLB, page table, permission, page fault 종류로 내려갑니다. 성능 질문으로 바뀌면 TLB miss, major fault, page cache hit/miss, reclaim, swap, NUMA, false sharing을 말합니다. 운영 질문으로 바뀌면 heap, off-heap, page cache, socket buffer, cgroup limit, OOM kill을 구분합니다. 제품 질문으로 바뀌면 Kafka segment page cache, Cassandra memtable/SSTable/cache/compaction, Spark executor memory/shuffle spill/checkpoint를 연결합니다.

이 장의 마지막 확인은 trace 하나입니다.

```text
application allocates or maps memory
  -> virtual address range exists
  -> first access may fault
  -> kernel attaches anonymous or file-backed page
  -> CPU caches translation in TLB
  -> memory pressure later reclaims, writes back, swaps, or kills
  -> product symptom appears as latency, OOM, GC, cache miss, spill, timeout
```

이 trace를 자기 말로 말할 수 있으면, 메모리를 "heap 크기"로만 보던 단계에서 벗어나 OS와 runtime, distributed product가 같은 physical resource를 공유한다는 관점으로 올라올 수 있습니다.

## Page table 자체도 메모리를 쓴다

가상 주소 공간을 넓게 잡으면 실제 data page만 문제 되는 것이 아닙니다. Page table도 kernel이 관리하는 memory를 씁니다. Sparse mapping은 모든 address range에 physical page를 붙이지 않더라도, mapping metadata와 page table level을 필요로 할 수 있습니다. Process가 수많은 작은 mapping을 만들거나, 많은 thread stack과 mmap region을 만들면 kernel metadata와 TLB pressure가 함께 늘 수 있습니다.

```text
many mappings
  -> VMA metadata grows
  -> page tables grow as pages are populated
  -> TLB/cache locality can suffer
  -> page fault handling has more metadata to inspect
```

JVM, memory-mapped index, large off-heap allocator, many classloader/native library mappings은 모두 이 표면을 건드립니다. Cassandra가 많은 SSTable 파일을 열고 index/filter를 mapping하거나, Spark executor가 많은 spill file과 library mapping을 갖거나, Kafka broker가 segment 파일을 많이 다룰 때 file descriptor뿐 아니라 mapping count와 kernel memory도 같이 볼 수 있습니다. `vm.max_map_count` 같은 설정이 등장하는 이유도 이 때문입니다.

## Reclaim이 latency로 보이는 순간

Memory reclaim은 background에서만 조용히 도는 일이 아닙니다. Allocation을 해야 하는 thread가 free page를 빨리 얻지 못하면 직접 reclaim 경로에 들어갈 수 있고, 이때 application thread가 latency를 겪습니다. Dirty page가 많으면 writeback을 기다릴 수 있고, anonymous page를 swap으로 내보내면 이후 다시 접근할 때 swap-in 비용이 생깁니다. 메모리 압박은 CPU 사용률보다 먼저 p99 latency로 보일 수 있습니다.

```text
request thread allocates buffer
  -> free pages low
  -> enters reclaim path
  -> scans page lists
  -> may trigger/write wait for dirty pages
  -> allocation finally succeeds
  -> request latency includes reclaim time
```

Kafka에서 큰 produce/fetch buffer가 늘고 page cache가 흔들리면 broker latency가 튈 수 있습니다. Cassandra compaction이 큰 I/O와 allocation을 만들면 foreground read가 reclaim과 cache miss를 만날 수 있습니다. Spark shuffle이 많은 buffer와 spill을 만들면 executor memory pressure가 stage runtime을 늘리고, container limit에서는 OOM으로 바로 이어질 수 있습니다. "메모리가 부족하면 죽는다"보다 "메모리가 부족해지기 전부터 reclaim과 cache miss가 latency를 만든다"가 더 실전적인 문장입니다.

## Garbage collection과 OS memory는 서로 독립적이지 않다

JVM GC는 heap 안의 객체 생명주기를 관리합니다. OS는 process 전체의 virtual memory와 physical memory, page cache, kernel memory를 관리합니다. 둘은 다른 계층이지만 독립적이지 않습니다. Heap을 크게 잡으면 GC 빈도는 줄 수 있지만, OS page cache와 다른 native memory 여유가 줄어 disk read와 reclaim이 늘 수 있습니다. Heap을 작게 잡으면 GC가 자주 돌고 allocation pressure가 커질 수 있습니다. Direct buffer나 mmap은 GC heap 밖에 있지만, reference가 남아 있으면 해제 timing이 늦어질 수 있습니다.

```text
increase JVM heap
  -> fewer collections maybe
  -> less memory for page cache/native buffers
  -> more disk I/O or cgroup pressure maybe

decrease JVM heap
  -> more page cache headroom maybe
  -> more GC pressure maybe
```

Kafka는 heap보다 page cache를 중시하는 경향이 강합니다. Cassandra는 heap, off-heap, memtable, cache, page cache의 균형이 중요합니다. Spark는 executor heap과 memory overhead, off-heap, Python worker, shuffle spill이 함께 움직입니다. 그러므로 memory tuning은 `-Xmx` 하나를 움직이는 것이 아니라 "어떤 memory consumer가 줄고 어떤 consumer가 늘어나는가"를 묻는 tradeoff입니다.

## Interview replay: 관측 지표를 계층별로 배치하기

메모리 장애를 받으면 다음 순서로 질문을 세울 수 있습니다. 첫째, process가 보는 virtual memory와 실제 resident memory를 나눕니다. 둘째, heap과 heap 밖 memory를 나눕니다. 셋째, anonymous memory와 file-backed page cache를 나눕니다. 넷째, cgroup limit과 host 전체 memory를 나눕니다. 다섯째, latency가 allocation, reclaim, page fault, GC, I/O 중 어디서 생겼는지 봅니다.

```text
observability map
  JVM GC log / heap dump
  -> heap object pressure

  native memory tracking / smaps
  -> off-heap, metaspace, thread stack, mapping

  /proc/meminfo / cgroup memory.events
  -> reclaim, OOM, cache, limit

  major faults / swap in-out / iostat
  -> memory pressure reaching storage
```

이 map을 제품별로 바꾸면 Kafka는 page cache와 network buffer, Cassandra는 memtable/cache/SSTable/compaction, Spark는 executor memory/shuffle spill/checkpoint를 중심으로 봅니다. 좋은 답변은 "메모리 부족입니다"에서 끝나지 않고, 어떤 memory 종류가 어떤 lower-layer event를 만들고 어떤 product symptom으로 올라왔는지 보여 줍니다.

## 마지막 연결: 메모리의 정답은 한 숫자가 아니다

운영에서 메모리를 볼 때 `free`, RSS, heap usage, cgroup usage, page cache, swap, major fault, GC pause가 서로 다른 이야기를 합니다. 이 숫자들은 서로 모순되는 것이 아니라 다른 층을 보고 있습니다. Heap usage가 낮아도 cgroup OOM이 날 수 있고, free memory가 낮아도 page cache로 재사용 가능한 상태일 수 있으며, RSS가 높아도 shared mapping 때문에 process별 합산이 실제 물리 사용량보다 커 보일 수 있습니다.

```text
one symptom: process killed
  -> JVM log has OutOfMemoryError? maybe heap
  -> kernel log has OOM kill? maybe cgroup/host memory
  -> cgroup memory.events increments? container limit
  -> major faults/swap high? reclaim pressure
```

이런 이유로 면접이나 장애 대응에서 좋은 답은 항상 "어느 메모리입니까?"로 되묻는 데서 시작합니다. Java heap인지, native memory인지, page cache인지, kernel socket buffer인지, cgroup limit인지, NUMA locality인지에 따라 관측과 해결이 달라집니다. Kafka의 처리량을 위해 page cache를 남길지, Cassandra의 read path를 위해 cache와 compaction을 조절할지, Spark의 executor memory overhead를 얼마나 줄지 모두 이 질문에서 갈라집니다.

메모리 장을 끝내며 가장 중요한 기준은 단순합니다. Virtual memory는 process에게 독립된 주소 공간을 주지만, physical memory는 모든 계층이 공유하는 실제 자원입니다. 이 둘을 동시에 볼 수 있어야 "주소는 넉넉한데 page fault가 난다", "heap은 남는데 container가 죽는다", "disk가 느린 줄 알았는데 page cache가 밀렸다" 같은 문장을 정확히 해석할 수 있습니다.

이 기준은 작은 튜닝에도 적용됩니다. Heap을 키우기 전에 page cache가 줄어드는지 보고, direct buffer를 늘리기 전에 cgroup memory overhead를 보며, mmap을 쓰기 전에 page fault와 mapping 수를 봅니다. 메모리 튜닝은 한 숫자를 크게 만드는 일이 아니라, 여러 계층의 memory consumer가 같은 physical memory를 어떤 시간 순서로 빌려 쓰는지 조정하는 일입니다.

그래서 좋은 관측 보고서는 "메모리 사용률 90%"라고만 쓰지 않습니다. 어떤 메모리가 90%인지, 회수 가능한지, 회수하면 어떤 I/O가 늘어나는지, process가 죽기 전에 어떤 latency 신호가 먼저 나타났는지까지 적어야 다음 조치가 안전해집니다.

그 차이가 실제 장애 복구 시간을 줄입니다.
