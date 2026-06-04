# 04. Cassandra Deep Dive

## 목차

- [1. Ring과 token은 어느 node가 데이터를 맡는지 정한다](#1-ring과-token은-어느-node가-데이터를-맡는지-정한다)
- [2. Write path는 append와 memory buffer로 빠르게 받는다](#2-write-path는-append와-memory-buffer로-빠르게-받는다)
- [3. Read path는 여러 후보에서 최신 값을 조립한다](#3-read-path는-여러-후보에서-최신-값을-조립한다)
- [4. Tunable consistency는 quorum이라는 단어 하나로 끝나지 않는다](#4-tunable-consistency는-quorum이라는-단어-하나로-끝나지-않는다)
- [5. Hinted handoff, read repair, anti-entropy repair는 서로 다른 회복 장치다](#5-hinted-handoff-read-repair-anti-entropy-repair는-서로-다른-회복-장치다)
- [6. OS resource 관점에서 Cassandra를 읽는다](#6-os-resource-관점에서-cassandra를-읽는다)
- [현실 시나리오 1: stale read가 의심된다](#현실-시나리오-1-stale-read가-의심된다)
- [현실 시나리오 2: compaction 때문에 cluster가 느려진다](#현실-시나리오-2-compaction-때문에-cluster가-느려진다)
- [Coordinator write path를 network, log, memory로 나눠 보기](#coordinator-write-path를-network-log-memory로-나눠-보기)
- [Read path는 Bloom filter, index, page cache, repair를 지난다](#read-path는-bloom-filter-index-page-cache-repair를-지난다)
- [Compaction은 storage 청소가 아니라 read/write tradeoff 조절이다](#compaction은-storage-청소가-아니라-readwrite-tradeoff-조절이다)
- [Gossip, failure detection, hinted handoff는 수렴을 돕지만 마법이 아니다](#gossip-failure-detection-hinted-handoff는-수렴을-돕지만-마법이-아니다)
- [Tunable consistency는 availability, latency, recency의 선택이다](#tunable-consistency는-availability-latency-recency의-선택이다)
- [Cassandra의 OS pressure 읽기](#cassandra의-os-pressure-읽기)
- [Interview replay: Cassandra를 한 번에 설명하기](#interview-replay-cassandra를-한-번에-설명하기)
- [Tombstone은 삭제가 아니라 이후 정리될 삭제 표시다](#tombstone은-삭제가-아니라-이후-정리될-삭제-표시다)
- [Partition key 설계는 분산성과 local read path를 동시에 정한다](#partition-key-설계는-분산성과-local-read-path를-동시에-정한다)
- [Multi-DC와 LOCAL_QUORUM은 latency와 장애 영역을 나누는 선택이다](#multi-dc와-local_quorum은-latency와-장애-영역을-나누는-선택이다)
- [Cassandra 작은 관측 map](#cassandra-작은-관측-map)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [근거와 더 읽을 자료](#근거와-더-읽을-자료)

Cassandra는 "NoSQL이라 빠르다"로 설명하면 위험합니다. Cassandra의 핵심은 많은 node가 계속 write를 받을 수 있게 data ownership, replication, write path, repair를 설계한 것입니다. 빠른 write는 공짜가 아닙니다. commit log와 memtable로 random update를 피하는 대신, read path와 compaction, repair가 복잡해집니다.

이 문서는 Cassandra를 기능 목록이 아니라 mutation 하나가 token ring, replica, commit log, memtable, SSTable, compaction, repair를 지나가는 흐름으로 읽습니다.

## 1. Ring과 token은 어느 node가 데이터를 맡는지 정한다

Cassandra는 key를 hash해 token으로 바꾸고, token range를 node들에게 나누어 줍니다. 이 구조를 ring이라고 부릅니다. ring은 원형 그림으로 자주 나오지만, 핵심은 그림이 아니라 ownership입니다. 어떤 partition key가 들어오면 그 key의 token이 어느 range에 속하는지 보고 replica node를 고릅니다.

```text
partition key "user-42"
  -> hash
  -> token T
  -> token range owner
  -> replica placement by replication factor and strategy
```

replication factor(RF)는 같은 data를 몇 replica에 둘지 정합니다. RF=3이면 보통 서로 다른 node, 가능하면 서로 다른 rack에 복사본을 둡니다. vnode는 한 물리 node가 여러 token range를 맡게 해 data distribution과 rebalance를 부드럽게 만들지만, repair와 운영 복잡도도 함께 가져옵니다.

partition key 설계는 correctness와 performance를 같이 바꿉니다. hot key가 있으면 특정 replica 집합에 traffic이 몰립니다. 너무 큰 partition은 read와 compaction 비용을 키웁니다. 너무 잘게 나누면 query가 여러 partition을 건너야 해 비싸집니다. Cassandra에서 data modeling이 중요한 이유는 query planner가 모든 것을 해결해 주는 전통적 RDBMS 모델이 아니기 때문입니다.

## 2. Write path는 append와 memory buffer로 빠르게 받는다

Cassandra write는 update-in-place를 피합니다. mutation이 들어오면 coordinator node는 해당 key의 replica들에게 write를 보냅니다. 각 replica는 mutation을 commit log에 append하고 memtable에 반영합니다. commit log는 crash recovery를 위한 write-ahead log이고, memtable은 memory 안의 정렬된 write buffer입니다. memtable이 threshold에 도달하면 disk의 immutable SSTable로 flush됩니다.

```text
client mutation
  -> coordinator
  -> replica nodes
       -> append to commit log
       -> update memtable
       -> acknowledge according to consistency level
  -> later memtable flush
       -> SSTable files on disk
  -> later compaction
       -> merge SSTables, discard obsolete/tombstoned data
```

이 구조가 write에 유리한 이유는 disk에서 random page를 찾아 in-place update하지 않기 때문입니다. commit log append는 순차 쓰기이고, memtable은 memory write입니다. SSTable은 flush 후 immutable입니다. 하지만 같은 partition의 최신 값이 여러 SSTable에 흩어질 수 있고, delete는 tombstone이라는 marker로 남습니다. read는 memtable과 여러 SSTable을 함께 봐야 하며, compaction은 그 흩어진 history를 이후 정리합니다.

commit log sync 정책은 `write()`와 `fsync()`의 차이를 Cassandra 버전으로 다시 만나는 지점입니다. `batch` 모드에서는 write ack 전에 commit log fsync를 기다리는 쪽으로 더 강한 내구성을 선택합니다. `periodic` 모드에서는 정해진 주기로 sync하면서 latency를 낮춥니다. 어떤 모드가 맞는지는 workload와 손실 허용 범위에 달려 있습니다.

이 write path는 [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md)의 durable write 구분을 그대로 사용합니다. commit log append가 빠른 이유는 random update보다 순차 write에 가깝기 때문입니다. 하지만 ack 시점이 commit log file의 page cache 반영인지, fsync 완료인지, replica CL 충족인지에 따라 crash 후 믿을 수 있는 상태가 달라집니다.

```text
mutation reaches replica
  -> append commit log record
  -> update memtable
  -> maybe wait for commitlog sync policy
  -> send ack to coordinator
  -> later memtable flush creates SSTable
```

## 3. Read path는 여러 후보에서 최신 값을 조립한다

read는 write보다 단순하지 않습니다. coordinator는 replica에 read를 보내고, replica는 먼저 memtable을 보고, Bloom filter로 SSTable에 key가 없을 가능성을 빠르게 걸러내며, partition index와 summary를 통해 SSTable 안의 위치를 찾습니다. 여러 SSTable에 같은 key의 다른 version이 있으면 timestamp와 tombstone 규칙에 따라 결과를 merge합니다.

```text
read key K
  -> coordinator chooses replicas by consistency level
  -> replica checks memtable
  -> check Bloom filters for SSTables
  -> read candidate rows/cells
  -> merge by timestamp/tombstone
  -> coordinator compares replica responses
  -> maybe trigger read repair or digest mismatch handling
```

Bloom filter는 "있다"를 보장하는 도구가 아닙니다. false positive가 있을 수 있지만 false negative를 피하도록 설계된 확률적 자료구조입니다. 즉 Bloom filter가 "없다"고 말하면 해당 SSTable을 건너뛰고, "있을 수 있다"고 말하면 실제 disk 구조를 더 봅니다.

compaction은 read 비용을 줄이고 space를 회수하지만 background I/O를 크게 씁니다. compaction이 밀리면 SSTable 수가 늘고 read amplification이 커집니다. tombstone이 많으면 read가 오래된 delete marker까지 검사해야 해서 latency가 커질 수 있습니다. 그래서 Cassandra 성능 문제는 write latency, read latency, compaction throughput, tombstone, disk bandwidth를 함께 봐야 합니다.

block I/O 관점에서 compaction은 foreground 요청과 같은 disk queue를 두고 경쟁합니다. 여러 SSTable을 sequential하게 읽고 새 SSTable을 쓰더라도, 동시에 commit log append와 normal read가 들어오면 page cache, device queue, CPU decompression/merge 비용이 겹칩니다. [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md)의 block layer trace를 Cassandra에 대입하면 pending compaction이 왜 read latency와 write latency를 동시에 흔드는지 보입니다.

## 4. Tunable consistency는 quorum이라는 단어 하나로 끝나지 않는다

Cassandra는 요청마다 consistency level(CL)을 고를 수 있습니다. write CL은 몇 replica가 성공 응답해야 write 성공으로 볼지 정합니다. read CL은 몇 replica에서 응답을 받아 read 성공으로 볼지 정합니다. RF=3에서 `QUORUM`은 보통 2개 replica 응답을 뜻합니다.

```text
RF=3 replicas: A, B, C

write CL=QUORUM:
  A and B acknowledge
  -> client sees success

read CL=QUORUM:
  B and C respond
  -> intersects at B
```

이 교차 때문에 quorum read/write는 recency를 높입니다. 하지만 "항상 최신"이라고 말하면 안 됩니다. clock skew가 있는 timestamp conflict, failed write, repair되지 않은 replica, hinted handoff window, tombstone, LWT 사용 여부가 결과에 영향을 줍니다. Cassandra의 일반 write는 consensus commit처럼 모든 replica가 같은 log 순서에 합의하는 방식이 아닙니다. Lightweight transaction(LWT) 같은 별도 경로가 필요한 이유가 여기 있습니다.

면접에서는 "Cassandra는 eventual consistency인가요?"라는 질문에 "기본 모델은 availability와 tunable consistency를 중시하고, CL 조합으로 read/write 관측을 조절합니다. 다만 LWT 같은 강한 경로는 별도 비용을 내고 사용합니다"라고 답한 뒤, RF=3 CL=QUORUM trace로 설명하면 좋습니다.

## 5. Hinted handoff, read repair, anti-entropy repair는 서로 다른 회복 장치다

replica가 잠시 내려가 있으면 coordinator는 hint를 저장했다가 해당 replica가 돌아왔을 때 전달할 수 있습니다. hinted handoff는 불일치 시간을 줄이는 best-effort 장치입니다. 영원히 보장되는 복구 mechanism으로 말하면 안 됩니다.

read repair는 read 과정에서 replica 간 불일치가 발견될 때 수리하는 경로입니다. anti-entropy repair는 Merkle tree 같은 구조로 replica 간 data 차이를 비교하고 동기화하는 더 체계적인 운영 작업입니다. repair를 하지 않으면 오래된 replica와 tombstone 처리 문제가 쌓여 stale read나 deleted data resurrection 위험이 커질 수 있습니다.

```text
replica C down during write
  -> A, B accept write
  -> hint may be stored for C
  -> C returns
  -> hint delivery or repair brings C closer
  -> anti-entropy repair is still needed for long-term convergence
```

여기서도 OS와 연결됩니다. repair와 compaction은 disk와 network를 많이 씁니다. 운영 중 heavy repair나 manual compaction을 함부로 실행하면 정상 traffic의 latency를 올릴 수 있습니다. 실험 문서에서 `nodetool compact`를 local-only, opt-in으로 제한해야 하는 이유도 이 때문입니다.

## 6. OS resource 관점에서 Cassandra를 읽는다

Cassandra node는 JVM process입니다. heap에는 object와 일부 cache가 있고, off-heap memory와 OS page cache도 중요합니다. commit log와 SSTable은 filesystem 위의 file입니다. compaction은 많은 file을 읽고 새 file을 쓰는 background I/O입니다. network는 replica read/write와 repair streaming을 운반합니다.

```text
write pressure
  -> commit log append
  -> memtable grows
  -> flush to SSTable
  -> compaction backlog
  -> disk bandwidth and page cache pressure
  -> read/write latency changes
```

따라서 Cassandra를 볼 때 JVM heap만 보면 부족합니다. memtable flush가 밀리는지, commit log disk가 느린지, SSTable 수가 많은지, compaction pending이 늘었는지, page cache hit가 줄었는지, network repair traffic이 많은지 함께 봐야 합니다.

network도 같은 방식으로 내려가야 합니다. coordinator가 replica에게 read/write를 보내고 repair가 streaming을 수행하는 동안 [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md)의 socket receive queue와 TCP send buffer가 모두 관여합니다. read timeout은 replica process crash, compaction으로 인한 scheduler delay, GC pause, socket queue backlog, packet loss/retransmission과 모두 양립할 수 있습니다. "timeout이면 replica 실패"라고 말하면 분산 시스템의 partial failure와 OS network path를 동시에 놓친 답이 됩니다.

container 환경에서는 [01e_concurrency_isolation_observability.md](01e_concurrency_isolation_observability.md)의 cgroup 질문도 필요합니다. Cassandra가 heap 밖에서 page cache와 off-heap memory를 쓰고, compaction과 repair가 disk/network를 강하게 쓰기 때문에, container memory/I/O limit은 단순한 배포 옵션이 아니라 storage engine의 실제 동작 조건입니다.

## 현실 시나리오 1: stale read가 의심된다

사용자가 방금 프로필을 바꿨는데 바로 조회하면 예전 값이 보인다고 합시다. 첫 질문은 "Cassandra가 eventual이라서 그렇다"가 아닙니다. 어떤 key에 어떤 timestamp의 mutation이 들어갔고, write CL은 무엇이었고, read CL은 무엇이었고, read가 어느 replica에서 응답을 받았는지 봐야 합니다.

```text
write email=b with CL=ONE
  -> replica A accepted
  -> B, C not yet updated

read CL=ONE
  -> replica C replies old email=a
  -> stale read visible
```

CL을 올리면 이런 가능성을 줄일 수 있지만 latency와 availability tradeoff가 생깁니다. LWT를 쓰면 더 강한 조건을 만들 수 있지만 비용이 큽니다. repair는 장기 수렴을 돕지만 read-after-write를 즉시 보장하는 만능 버튼은 아닙니다.

## 현실 시나리오 2: compaction 때문에 cluster가 느려진다

write traffic이 높고 SSTable이 많이 쌓이면 compaction이 바빠집니다. compaction은 작은 SSTable을 읽어 merge하고 새 SSTable을 씁니다. 이때 disk read/write bandwidth와 CPU, page cache가 정상 read/write traffic과 경쟁합니다. compaction을 너무 늦추면 read amplification이 커지고, 너무 공격적으로 돌리면 foreground latency가 올라갑니다.

해결은 "node를 늘린다"로 끝나지 않습니다. partition 설계, tombstone 양, compaction strategy, disk throughput, heap/off-heap memory, repair schedule을 함께 봐야 합니다.

## Coordinator write path를 network, log, memory로 나눠 보기

Cassandra write는 client가 아무 replica에나 data를 던지는 단순한 흐름이 아닙니다. Client request를 받은 node는 coordinator가 되고, partition key를 token으로 해석해 replica set을 찾고, 필요한 replica에 mutation을 보냅니다. Replica는 mutation을 commitlog에 append하고 memtable에 반영합니다. Coordinator는 consistency level에 필요한 replica response를 받으면 client에게 성공을 돌려줍니다.

```text
client
  -> sends write to coordinator over TCP

coordinator
  -> hashes partition key to token
  -> finds natural replicas
  -> sends mutation to replicas

replica
  -> appends mutation to commitlog
  -> applies mutation to memtable
  -> acknowledges according to local durability path

coordinator
  -> waits for CL responses
  -> returns success or timeout
```

이 trace에서 OS와 분산 시스템이 동시에 보입니다. Network path는 socket buffer, scheduler, TCP timeout을 지납니다. Commitlog append는 filesystem page cache, dirty writeback, fsync policy, block queue를 지납니다. Memtable은 heap/off-heap memory와 allocator, GC pressure를 만납니다. Coordinator의 timeout은 replica가 죽었다는 증명이 아니라, coordinator가 deadline 안에 충분한 response를 받지 못했다는 관측입니다. 일부 replica에는 이미 write가 들어갔을 수 있습니다.

Commitlog sync mode는 durability와 latency의 중요한 경계입니다. Batch sync는 더 자주 강한 동기화를 요구할 수 있고, periodic sync는 latency를 줄이는 대신 crash window를 가질 수 있습니다. 정확한 설정과 version별 의미는 공식 문서와 운영 설정을 확인해야 합니다. 중요한 것은 "commitlog에 썼다"와 "storage에 안정적으로 내려갔다"와 "CL 성공을 받았다"를 같은 이벤트(event)로 합치지 않는 것입니다.

## Read path는 Bloom filter, index, page cache, repair를 지난다

Cassandra read는 partition key로 replica를 찾아 data를 가져오는 흐름입니다. Replica 내부에서는 memtable과 SSTable을 함께 봅니다. Memtable에는 최근 write가 있고, SSTable에는 flush된 immutable data가 있습니다. 여러 SSTable에 같은 partition이나 row fragment가 있을 수 있으므로, read path는 Bloom filter로 "이 SSTable에 key가 없을 가능성"을 빠르게 배제하고, partition index와 compression offset map을 사용해 필요한 data 위치로 이동합니다.

```text
read request
  -> coordinator selects replicas
  -> replica checks memtable
  -> Bloom filter skips irrelevant SSTables
  -> partition index locates candidate rows
  -> OS page cache or disk reads SSTable blocks
  -> results merge by timestamp/tombstone rules
  -> coordinator reconciles responses
```

Bloom filter는 false positive가 있을 수 있지만 false negative가 없어야 하는 구조입니다. "없다"고 말한 key가 실제로 있으면 correctness가 깨집니다. False positive는 불필요한 disk/index lookup을 만들 뿐입니다. SSTable이 immutable이기 때문에 update는 기존 row를 제자리에서 바꾸지 않고 새 version을 씁니다. Delete도 tombstone으로 기록되어 compaction이 적절한 시점에 정리합니다. 이 구조는 write를 빠르게 만들지만 read amplification과 compaction 비용을 만듭니다.

Read repair, digest read, speculative retry, consistency level은 read path를 더 복잡하게 만듭니다. Coordinator는 여러 replica에서 data나 digest를 받아 차이를 발견할 수 있고, background repair나 read repair가 replica 간 divergence를 줄입니다. 그러나 repair는 즉시 모든 stale read를 없애는 버튼이 아닙니다. Tombstone과 clock skew, failed repair, long GC pause, network partition이 섞이면 read 결과를 설명할 때 더 많은 evidence가 필요합니다.

## Compaction은 storage 청소가 아니라 read/write tradeoff 조절이다

LSM 계열 storage는 write를 빠르게 받기 위해 memtable과 immutable SSTable을 사용합니다. 하지만 SSTable이 계속 쌓이면 read는 여러 파일을 봐야 하고, tombstone과 old version이 남아 disk space와 read latency를 키웁니다. Compaction은 여러 SSTable을 읽어 merge하고 새 SSTable을 쓰며 오래된 version이나 tombstone을 정리합니다. 즉 compaction은 청소 작업이 아니라 write-optimized 구조가 뒤로 미룬 비용을 갚는 과정입니다.

```text
many small SSTables
  -> reads check more files
  -> tombstones remain visible
  -> disk space grows

compaction
  -> read old SSTables
  -> merge rows and tombstones
  -> write new SSTable
  -> drop obsolete files
```

Compaction strategy에 따라 비용이 달라집니다. Size-tiered, leveled, time-window, unified 같은 전략은 write amplification, read amplification, space amplification을 다르게 조절합니다. Time-series workload와 random update workload는 같은 전략으로 최적화되지 않습니다. Tombstone이 많으면 compaction이 더 중요하지만, tombstone grace와 repair window를 무시하고 빨리 지우면 삭제가 다시 살아나는 위험이 생길 수 있습니다.

OS 관점에서 compaction은 큰 sequential read/write, page cache churn, CPU compression/checksum, disk queue pressure를 만듭니다. Foreground read/write와 같은 disk를 공유하므로 compaction concurrency와 throughput을 조절해야 합니다. 너무 늦추면 read amplification이 커지고, 너무 빠르게 돌리면 foreground latency가 나빠집니다. 이 tradeoff를 이해해야 "compaction 때문에 느리다"를 단순히 compaction을 끄는 문제로 보지 않습니다.

## Gossip, failure detection, hinted handoff는 수렴을 돕지만 마법이 아니다

Cassandra는 gossip으로 node 상태와 membership 정보를 퍼뜨립니다. 각 node는 다른 node의 상태를 주기적으로 교환하고, failure detector는 heartbeat 지연을 바탕으로 의심 정도를 계산합니다. 이 정보는 coordinator가 replica availability를 판단하고 cluster가 topology를 이해하는 데 쓰입니다. 하지만 gossip은 즉시 전역 진실을 만드는 protocol이 아닙니다. Node마다 잠시 다른 view를 가질 수 있습니다.

Hinted handoff는 replica가 일시적으로 unavailable할 때 coordinator가 hint를 저장했다가 복귀 후 전달하는 best-effort 보조 장치입니다. 이것은 repair를 대체하지 않습니다. Hint window를 넘거나 coordinator가 hint를 잃거나 long outage가 생기면 divergence가 남을 수 있습니다. Anti-entropy repair는 replica data를 비교하고 차이를 고쳐 장기 수렴을 돕습니다. Merkle tree 같은 구조는 전체 data를 모두 전송하지 않고 차이를 찾는 데 사용될 수 있습니다.

```text
write to replicas A, B, C
  -> C unavailable
  -> coordinator stores hint for C
  -> write CL may still succeed with A, B
  -> later hint is delivered to C if within policy
  -> repair still needed for full convergence assurance
```

Read repair는 read path에서 발견한 inconsistency를 고칠 수 있지만, table setting과 consistency tradeoff가 있습니다. Repair는 network, disk, CPU를 많이 쓰므로 운영 schedule이 중요합니다. Repair를 너무 늦추면 stale data와 tombstone 문제를 키울 수 있고, 너무 공격적으로 돌리면 foreground traffic을 밀어낼 수 있습니다.

## Tunable consistency는 availability, latency, recency의 선택이다

Replication factor는 data가 몇 replica에 놓이는지입니다. Consistency level은 read/write operation이 몇 replica의 응답을 기다릴지입니다. RF=3에서 write CL=QUORUM, read CL=QUORUM이면 정상적인 조건에서 read/write 응답 집합이 겹칩니다. 하지만 이것은 blanket linearizability가 아닙니다. Conflict resolution, timestamp, failed write, read repair, LWT 사용 여부, clock skew, network partition이 결과에 영향을 줍니다.

```text
RF=3: A, B, C
write CL=QUORUM succeeds on A, B
read CL=QUORUM reads B, C
  -> intersection at B
  -> can see latest if B has winning version and reconciliation works
```

CL=ONE은 latency와 availability에 유리하지만 stale read 가능성이 커집니다. CL=ALL은 강하지만 한 replica만 불안정해도 availability가 낮아집니다. LOCAL_QUORUM은 multi-DC 환경에서 local datacenter latency와 cross-DC tradeoff를 조절합니다. LWT는 compare-and-set 같은 조건부 update에 더 강한 보장을 제공하지만 비용이 큽니다. Accord나 newer transaction features는 ordinary quorum path와 분리해서 version-sensitive하게 설명해야 합니다.

## Cassandra의 OS pressure 읽기

Cassandra node가 느릴 때는 제품 metric과 OS metric을 함께 봐야 합니다. Write latency가 늘면 coordinator pending task, mutation stage, commitlog sync, memtable allocation, GC, disk queue를 봅니다. Read latency가 늘면 Bloom filter false positive, key cache hit, SSTable count, tombstone scan, page cache miss, disk read latency를 봅니다. Compaction backlog가 늘면 disk bandwidth, CPU, page cache churn, tombstone volume, compaction strategy를 봅니다.

```text
read timeout
  -> coordinator waited for replica responses
  -> replica may be slow due to disk read
  -> or GC pause
  -> or compaction pressure
  -> or network retransmission
  -> or scheduler/cgroup throttle
```

JVM heap만 보면 부족합니다. Off-heap memory, direct buffer, memtable allocation, page cache, file descriptors, mmap count, thread stacks가 함께 움직입니다. 컨테이너(container) 환경에서는 cgroup memory limit과 disk I/O throttle이 추가됩니다. Cassandra는 OS와 매우 가까운 database입니다. Storage engine을 이해하는 것은 SSTable 이름을 외우는 것이 아니라, immutable file과 page cache, compaction I/O, quorum response가 어떻게 request latency로 이어지는지 보는 일입니다.

## Interview replay: Cassandra를 한 번에 설명하기

Cassandra를 짧게 설명하면 "partition key로 data ownership을 나누고, replica set에 write/read를 보내며, tunable consistency level로 latency와 availability, recency를 조절하는 wide-column store"입니다. 내부 path를 붙이면 "write는 commitlog와 memtable에 들어가고, memtable flush가 immutable SSTable을 만들며, read는 memtable과 여러 SSTable을 Bloom filter/index/page cache를 통해 찾아 merge한다"가 됩니다. 운영 path를 붙이면 "compaction과 repair가 뒤로 미룬 비용과 replica divergence를 정리하지만 disk/network/CPU를 foreground와 공유한다"가 됩니다.

꼬리 질문에는 네 가지를 분리합니다. Commitlog 성공과 CL 성공은 다릅니다. QUORUM과 linearizability는 다릅니다. Hinted handoff와 repair는 다릅니다. Compaction을 줄이면 당장 I/O는 줄 수 있지만 read amplification과 space/tombstone 문제가 커질 수 있습니다. 이 네 구분을 말하면 Cassandra를 단순 NoSQL 키워드가 아니라 OS와 분산 storage의 결합으로 이해하고 있다는 신호가 됩니다.

## Tombstone은 삭제가 아니라 이후 정리될 삭제 표시다

Cassandra에서 delete는 즉시 모든 replica와 SSTable에서 data를 지우는 일이 아닙니다. 삭제는 tombstone으로 기록되고, replica들이 이 삭제 표시를 보고 오래된 값을 이겨야 합니다. Tombstone은 compaction과 repair window를 지나 안전하다고 판단될 때 정리될 수 있습니다. 너무 빨리 지우면 한동안 내려가 있던 replica가 오래된 값을 들고 돌아와 삭제된 data를 되살릴 수 있습니다.

```text
delete key K
  -> write tombstone with timestamp
  -> reads merge tombstone against older values
  -> repair propagates delete marker
  -> compaction later drops old data and tombstone when safe
```

Tombstone이 많으면 read path가 느려질 수 있습니다. Query는 많은 tombstone을 scan하고 merge해야 하며, compaction도 더 많은 metadata를 처리합니다. TTL을 많이 쓰는 time-series workload에서는 tombstone pattern과 compaction strategy가 특히 중요합니다. "삭제했는데 왜 느려졌나요?"라는 질문은 Cassandra에서 매우 현실적인 질문입니다. 삭제는 storage 비용을 미래로 미루는 write입니다.

## Partition key 설계는 분산성과 local read path를 동시에 정한다

Partition key는 data가 어떤 token range와 replica에 놓이는지 정합니다. 좋은 partition key는 load를 넓게 퍼뜨리고, query가 필요한 data를 한 partition 또는 예측 가능한 partition 집합에서 찾게 합니다. Hot partition은 특정 replica와 coordinator path를 과부하시키고, 너무 큰 partition은 read와 compaction, repair를 어렵게 만듭니다. 너무 잘게 쪼개면 query가 많은 partition을 fan-out해야 할 수 있습니다.

```text
bad key shape
  all writes for one tenant/day -> one huge hot partition

better shape maybe
  tenant + day + shard
  -> spreads writes
  -> query must know shard merge cost
```

Partition 설계는 OS와도 연결됩니다. Hot partition은 특정 SSTable range와 page cache, disk I/O를 반복해서 건드립니다. Large partition은 index와 row scan, memory allocation, network response size를 키웁니다. Repair와 streaming도 partition distribution에 영향을 받습니다. Cassandra는 query-first modeling을 요구한다는 말은, storage layout과 distributed ownership을 query shape에 맞춰 설계해야 한다는 뜻입니다.

## Multi-DC와 LOCAL_QUORUM은 latency와 장애 영역을 나누는 선택이다

여러 datacenter에 replica를 두면 지역 장애에 강해질 수 있지만, cross-DC latency와 consistency tradeoff가 생깁니다. LOCAL_QUORUM은 local datacenter 안에서 quorum을 만족시켜 latency를 줄이고 remote DC 장애의 영향을 제한하려는 선택입니다. EACH_QUORUM이나 QUORUM의 의미는 replication strategy와 DC별 RF를 함께 봐야 합니다.

```text
DC1 replicas: A, B, C
DC2 replicas: D, E, F

LOCAL_QUORUM in DC1
  -> wait for two of A/B/C
  -> remote propagation may lag
```

Multi-DC에서 stale read와 conflict, repair, hinted handoff, network partition은 더 복잡해집니다. Clock skew와 last-write-wins conflict resolution도 더 조심해야 합니다. Latency를 줄이려고 local만 보면 remote recency가 늦어질 수 있고, 강한 cross-DC coordination을 걸면 availability와 latency가 나빠집니다. 운영 답변은 CL 이름만 말하지 말고 DC별 replica와 request path를 그려야 합니다.

## Cassandra 작은 관측 map

Write latency가 늘면 `nodetool tpstats`, client-side latency, coordinator pending, commitlog sync, GC log, disk await를 같이 봅니다. Read latency가 늘면 table별 read latency, SSTable per read, Bloom filter false positive, key cache hit rate, tombstone warning, compaction backlog를 봅니다. Repair가 느리면 network streaming, merkle tree build, disk read, compaction aftermath를 봅니다.

```text
Cassandra symptom
  write timeout
    -> coordinator waiting for CL responses
    -> replica commitlog/memtable/GC/network

  read timeout
    -> SSTable scan/tombstone/page cache/disk
    -> replica or coordinator overload

  high pending compaction
    -> write amplification, disk queue, tombstones, strategy mismatch
```

이 map을 OS 관측과 연결하면 Cassandra 장애 대응이 더 안정됩니다. JVM log와 Cassandra metric이 한쪽이고, `iostat`, `ss`, cgroup, `/proc`, profiler가 다른 한쪽입니다. 둘을 같이 놓아야 "Cassandra가 느리다"를 "commitlog fsync가 느리다", "compaction이 page cache를 흔든다", "coordinator timeout은 network와 GC가 섞였다"처럼 실행 가능한 가설로 바꿀 수 있습니다.

## 문서를 덮고 확인할 것

- Cassandra write가 commit log와 memtable 두 곳에 들어가는 이유를 설명해 보세요.
- SSTable이 immutable인 것이 write와 read에 각각 어떤 영향을 주는지 말해 보세요.
- RF와 CL을 구분하고, RF=3 CL=QUORUM read/write trace를 그려 보세요.
- hinted handoff, read repair, anti-entropy repair를 서로 구분해 보세요.
- compaction이 read latency와 disk bandwidth에 동시에 영향을 주는 이유를 설명해 보세요.

## 근거와 더 읽을 자료

- Apache Cassandra 5.0.x documentation: architecture overview, Dynamo, storage engine, guarantees, hints, compaction.
- Dynamo paper for always-on key-value store lineage and hinted handoff/quorum context.
- LSM-tree paper for write-optimized storage and compaction tradeoff.
- Linux page cache and `fsync(2)` documentation for commit-log durability boundary.
