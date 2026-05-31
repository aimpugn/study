# 01b. Memory, Address Space, Page Fault

메모리 문제는 "heap을 늘리면 된다"로 끝나지 않습니다. Kafka broker, Cassandra node, Spark executor는 모두 process이고, process는 자기만의 가상 주소 공간을 보는 것처럼 실행됩니다. 하지만 실제 physical memory는 OS page cache, JVM heap, off-heap buffer, mmap file, kernel socket buffer, filesystem metadata, 다른 process와 함께 나눠 씁니다. 이 층을 모르면 GC, OOM, page cache miss, mmap, swap, container memory limit이 서로 다른 단어처럼 보입니다.

이 문서에서는 virtual address가 physical frame으로 번역되는 흐름에서 시작해, page fault, copy-on-write, mmap, allocator, OOM, page cache가 backend system의 증상으로 어떻게 올라오는지 설명합니다.

## virtual address는 process에게 주는 보호된 착시다

process가 포인터 `0x1000`을 읽는다고 해도 CPU가 곧바로 물리 메모리 주소 0x1000을 읽는 것은 아닙니다. CPU의 MMU(memory management unit)는 virtual address를 page table을 통해 physical frame으로 번역합니다. page table은 "이 process의 virtual page가 어느 physical frame에 연결되는가, 읽기/쓰기/실행 권한은 무엇인가"를 담은 구조입니다.

```
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

```
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

```
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

`mmap()`은 파일이나 anonymous memory를 process의 address space에 매핑합니다. 파일을 `read()`로 buffer에 복사하는 대신, 파일의 byte range를 virtual address range로 보이게 만들 수 있습니다. 그러나 "mmap은 디스크 I/O가 없다"는 뜻이 아닙니다. 파일-backed mapping은 해당 page를 처음 만질 때 page fault를 통해 page cache와 storage path를 만날 수 있습니다. dirty mmap page도 나중에 writeback되어야 합니다.

```
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

```
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

```
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

```
virtual address
  -> page table / TLB
  -> physical frame or page fault
  -> anonymous/file-backed page
  -> reclaim/writeback/swap/OOM decision
```

이 trace를 설명할 수 있으면 "JVM heap은 남는데 OOM이 났다", "page cache가 커서 free memory가 작다", "mmap 파일을 읽었는데 왜 page fault가 늘었나" 같은 질문을 한 층 아래에서 다시 볼 수 있습니다.
