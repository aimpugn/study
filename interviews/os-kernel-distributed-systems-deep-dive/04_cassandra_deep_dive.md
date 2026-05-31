# 04. Cassandra Deep Dive

Cassandra는 "NoSQL이라 빠르다"로 설명하면 위험합니다. Cassandra의 핵심은 많은 node가 계속 write를 받을 수 있게 data ownership, replication, write path, repair를 설계한 것입니다. 빠른 write는 공짜가 아닙니다. commit log와 memtable로 random update를 피하는 대신, read path와 compaction, repair가 복잡해집니다.

이 문서는 Cassandra를 기능 목록이 아니라 mutation 하나가 token ring, replica, commit log, memtable, SSTable, compaction, repair를 지나가는 흐름으로 읽습니다.

## 1. Ring과 token은 어느 node가 데이터를 맡는지 정한다

Cassandra는 key를 hash해 token으로 바꾸고, token range를 node들에게 나누어 줍니다. 이 구조를 ring이라고 부릅니다. ring은 원형 그림으로 자주 나오지만, 핵심은 그림이 아니라 ownership입니다. 어떤 partition key가 들어오면 그 key의 token이 어느 range에 속하는지 보고 replica node를 고릅니다.

```
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

```
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

이 구조가 write에 유리한 이유는 disk에서 random page를 찾아 in-place update하지 않기 때문입니다. commit log append는 순차 쓰기이고, memtable은 memory write입니다. SSTable은 flush 후 immutable입니다. 하지만 같은 partition의 최신 값이 여러 SSTable에 흩어질 수 있고, delete는 tombstone이라는 marker로 남습니다. read는 memtable과 여러 SSTable을 함께 봐야 하며, compaction은 그 흩어진 history를 나중에 정리합니다.

commit log sync 정책은 `write()`와 `fsync()`의 차이를 Cassandra 버전으로 다시 만나는 지점입니다. `batch` 모드에서는 write ack 전에 commit log fsync를 기다리는 쪽으로 더 강한 내구성을 선택합니다. `periodic` 모드에서는 정해진 주기로 sync하면서 latency를 낮춥니다. 어떤 모드가 맞는지는 workload와 손실 허용 범위에 달려 있습니다.

이 write path는 [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md)의 durable write 구분을 그대로 사용합니다. commit log append가 빠른 이유는 random update보다 순차 write에 가깝기 때문입니다. 하지만 ack 시점이 commit log file의 page cache 반영인지, fsync 완료인지, replica CL 충족인지에 따라 crash 후 믿을 수 있는 상태가 달라집니다.

```
mutation reaches replica
  -> append commit log record
  -> update memtable
  -> maybe wait for commitlog sync policy
  -> send ack to coordinator
  -> later memtable flush creates SSTable
```

## 3. Read path는 여러 후보에서 최신 값을 조립한다

read는 write보다 단순하지 않습니다. coordinator는 replica에 read를 보내고, replica는 먼저 memtable을 보고, Bloom filter로 SSTable에 key가 없을 가능성을 빠르게 걸러내며, partition index와 summary를 통해 SSTable 안의 위치를 찾습니다. 여러 SSTable에 같은 key의 다른 version이 있으면 timestamp와 tombstone 규칙에 따라 결과를 merge합니다.

```
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

```
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

replica가 잠시 내려가 있으면 coordinator는 hint를 저장했다가 나중에 해당 replica가 돌아왔을 때 전달할 수 있습니다. hinted handoff는 불일치 시간을 줄이는 best-effort 장치입니다. 영원히 보장되는 복구 mechanism으로 말하면 안 됩니다.

read repair는 read 과정에서 replica 간 불일치가 발견될 때 수리하는 경로입니다. anti-entropy repair는 Merkle tree 같은 구조로 replica 간 data 차이를 비교하고 동기화하는 더 체계적인 운영 작업입니다. repair를 하지 않으면 오래된 replica와 tombstone 처리 문제가 쌓여 stale read나 deleted data resurrection 위험이 커질 수 있습니다.

```
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

```
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

```
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
