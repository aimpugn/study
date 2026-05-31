# 06. Kafka, Cassandra, Spark Cross-System Comparison

Kafka, Cassandra, Spark는 모두 partition, log, replication, recovery, backpressure 같은 단어를 씁니다. 하지만 같은 단어가 같은 약속을 뜻하지 않습니다. 이 문서의 목적은 세 시스템을 표로 외우는 것이 아니라, 같은 lower-layer 질문을 던졌을 때 각 시스템이 어디에서 다른 선택을 하는지 보는 것입니다.

가장 중요한 비교 질문은 다음 하나입니다.

```
What state is moving?
  -> who owns it?
  -> how is order defined?
  -> how is it copied or recomputed?
  -> where can it be lost or delayed?
  -> what observation proves recovery?
```

## 1. 세 시스템의 한 문장 모델

| 시스템 | 중심 상태 | 주된 약속 | 대가 |
|---|---|---|---|
| Kafka | partition log record와 consumer offset | 순서 있는 append, replay, fan-out consumption | partition 간 전역 순서 없음, lag/retention/replication 관리 필요 |
| Cassandra | partitioned mutation과 replica cell version | 높은 write availability, tunable consistency, scale-out storage | read/repair/compaction 복잡도, stale read 가능성 |
| Spark | partitioned computation과 lineage | 큰 계산을 task로 나누고 실패한 조각을 재계산 | shuffle/spill/skew/driver metadata 비용 |

Kafka는 data를 오래 읽는 log로 둡니다. Cassandra는 latest queryable state를 여러 SSTable과 replica에서 조립합니다. Spark는 data 자체보다 계산 graph와 partition을 중심으로 움직입니다.

## 2. Log는 세 시스템에서 서로 다르게 산다

Kafka의 log는 사용자에게 직접 보이는 중심 데이터 구조입니다. consumer는 offset을 지정해 log를 읽고, retention이 허용하는 동안 다시 읽을 수 있습니다. Kafka에서 log는 "서비스가 제공하는 데이터"입니다.

Cassandra의 commit log는 사용자 query 모델이 아닙니다. crash recovery를 위한 write-ahead log입니다. acknowledged mutation이 memtable에만 있고 node가 죽었을 때 commit log replay로 복원합니다. 실제 read는 memtable과 SSTable을 merge합니다.

Spark의 lineage는 파일 log가 아니라 계산의 기억입니다. 어떤 partition이 어떤 parent partition과 transformation으로 만들어졌는지를 기억해 실패한 partition을 다시 계산합니다. checkpoint를 만들면 lineage의 일부를 끊고 안정적인 저장소를 기준점으로 삼습니다.

```
Kafka log:
  record stream -> consumer offset -> replay

Cassandra commit log:
  mutation durability -> crash replay -> memtable restore

Spark lineage:
  transformation graph -> lost partition recompute -> checkpoint cut
```

면접에서 "세 시스템 모두 log를 쓰나요?"라고 물으면 "네"로 끝내면 안 됩니다. Kafka log는 사용자 데이터 모델, Cassandra commit log는 recovery mechanism, Spark lineage는 recomputation plan이라고 구분해야 합니다.

## 3. Partition은 ordering, ownership, computation의 경계다

Kafka partition은 ordering의 경계입니다. 같은 partition 안에서는 offset 순서가 있지만, 여러 partition 사이에는 전역 순서가 없습니다. partition 수는 consumer group 병렬성의 상한을 만듭니다.

Cassandra partition은 data ownership과 query 경계입니다. partition key가 token을 만들고 replica placement를 결정합니다. query가 partition key에 맞지 않으면 여러 node와 SSTable을 넓게 훑는 비용이 생깁니다.

Spark partition은 task의 입력 단위입니다. partition 수와 크기는 병렬성, task overhead, shuffle block 수, skew를 바꿉니다.

| 질문 | Kafka | Cassandra | Spark |
|---|---|---|---|
| 같은 partition 안에서 중요한 것 | offset order | row clustering / local read shape | task-local processing |
| partition이 많아지면 | parallelism 증가, metadata/replication 비용 증가 | distribution 개선 가능, repair/operation 복잡도 증가 | parallelism 증가, scheduler/shuffle overhead 증가 |
| 잘못 고르면 | hot partition, consumer 병렬성 한계 | hot key, wide partition, bad query | skew task, tiny task, spill |

## 4. Replication과 recovery는 목적이 다르다

Kafka replication은 partition log를 follower에게 복사해 leader failure에 대비합니다. 중요한 질문은 어떤 offset이 in-sync replica에 있고 consumer에게 노출 가능한가입니다.

Cassandra replication은 같은 partition data를 여러 node에 두어 availability와 latency를 조절합니다. 중요한 질문은 write CL과 read CL이 어떤 replica 집합을 만들고, repair가 불일치를 어떻게 줄이는가입니다.

Spark는 일반적인 data replication보다 recomputation을 먼저 활용합니다. executor가 죽어 cache나 shuffle block을 잃으면 lineage를 따라 다시 계산할 수 있습니다. 단, shuffle output 손실과 긴 lineage, nondeterministic computation은 checkpoint나 external storage가 필요할 수 있습니다.

```
failure of one node
  Kafka:
    elect replica leader if log is safe enough

  Cassandra:
    route around failed replica, later hints/repair converge

  Spark:
    reschedule task or recompute lost partition from lineage
```

## 5. Consistency 질문은 시스템별 read/write 계약으로 바꿔야 한다

Kafka에서 consistency 질문은 보통 "어느 offset까지 committed로 볼 수 있는가", "producer retry가 중복을 만들 수 있는가", "transactional read-process-write의 범위가 어디까지인가"로 바뀝니다.

Cassandra에서 consistency 질문은 "RF, write CL, read CL, timestamp conflict, LWT 여부, repair 상태가 어떤 read를 만들었는가"로 바뀝니다.

Spark에서 consistency 질문은 "input snapshot, task retry, cache/checkpoint, sink write가 같은 결과를 재현하는가"로 바뀝니다. Spark는 storage system이 아니라 computation engine이므로, 외부 source/sink consistency가 결과에 크게 들어옵니다.

따라서 "Kafka는 CP인가요?", "Cassandra는 AP인가요?", "Spark는 consistent한가요?" 같은 질문은 너무 거칠게 시작한 것입니다. 더 좋은 답은 각 시스템의 read/write 또는 compute/retry 계약으로 질문을 좁히는 것입니다.

## 6. Backpressure는 각 시스템의 queue 이름으로 나타난다

세 시스템 모두 producer와 consumer 속도가 어긋나면 queue가 자랍니다. 다만 queue의 이름과 위치가 다릅니다.

| 시스템 | 자라는 신호 | 실제로 볼 lower layer |
|---|---|---|
| Kafka | consumer lag, request queue, replica lag | broker disk/page cache, network, partition skew, downstream DB |
| Cassandra | pending compaction, write latency, dropped messages, read latency | commit log disk, memtable flush, SSTable count, tombstone, repair traffic |
| Spark | long stage, skewed task, spill size, shuffle fetch wait | executor heap/off-heap, local disk, network, GC, partition skew |

backpressure를 이해하지 못하면 worker 수나 thread 수를 늘리는 처방으로 장애를 키울 수 있습니다. downstream DB가 느린데 Kafka consumer를 늘리면 DB timeout과 retry가 늘 수 있습니다. Cassandra compaction이 disk를 잡아먹는데 read timeout만 늘리면 queue가 더 오래 쌓입니다. Spark skew가 있는데 executor만 늘리면 느린 partition 하나가 여전히 job을 붙잡습니다.

## 7. OS resource 관점의 공통 지도

| OS 자원 | Kafka | Cassandra | Spark |
|---|---|---|---|
| CPU | compression, request handling, TLS, consumer processing | read merge, compaction, serialization, GC | task execution, serialization, GC, SQL execution |
| Memory | broker buffer, page cache | memtable, cache, heap/off-heap, page cache | executor heap, storage/execution memory, broadcast, shuffle buffer |
| Disk | log segment append/read, retention/compaction | commit log, SSTable, compaction, repair streaming | shuffle spill, cache on disk, checkpoint |
| Network | producer/fetch traffic, replica fetch | replica read/write, repair, gossip | shuffle fetch, driver/executor RPC, data source reads |

이 표는 "어디를 봐야 하는가"를 정하기 위한 지도입니다. 예를 들어 disk saturation이 있으면 Kafka는 leader append와 follower fetch, Cassandra는 compaction과 commit log, Spark는 shuffle spill을 먼저 떠올립니다. 같은 disk라도 시스템마다 만들어내는 증상이 다릅니다.

OS 상세 문서와 연결하면 비교 질문이 더 구체적입니다.

| OS 상세 경로 | Kafka에서 묻는 질문 | Cassandra에서 묻는 질문 | Spark에서 묻는 질문 |
|---|---|---|---|
| [process/scheduler](01a_process_scheduling.md) | network/request/log cleaner thread가 CPU를 받는가? | request/compaction/repair thread가 서로 굶기지 않는가? | executor task thread와 driver scheduling이 어디서 기다리는가? |
| [memory/address](01b_memory_and_address_space.md) | heap 밖 page cache와 network buffer가 충분한가? | memtable, off-heap, page cache, heap이 어떻게 경쟁하는가? | heap, overhead, off-heap, spill buffer, page cache가 cgroup limit 안에 맞는가? |
| [file/block I/O](01c_filesystem_page_cache_block_io.md) | segment append/read, writeback, fsync 정책이 latency를 바꾸는가? | commit log, SSTable read, compaction write가 device queue를 채우는가? | shuffle spill/checkpoint/local disk read가 stage를 붙잡는가? |
| [network/multiplexing](01d_network_stack_and_io_multiplexing.md) | produce/fetch/replica traffic의 socket queue가 밀리는가? | coordinator-replica, gossip, repair streaming 중 어디가 막히는가? | shuffle fetch와 driver/executor RPC가 TCP/backpressure를 겪는가? |
| [concurrency/isolation/observability](01e_concurrency_isolation_observability.md) | lock/futex/thread wait와 cgroup throttle이 보이는가? | compaction/repaired data/GC/cgroup OOM을 분리했는가? | task skew, GC, off-CPU, container OOM을 증거로 구분했는가? |

## 현실 시나리오 1: 한 노드의 disk가 느려졌다

Kafka에서는 해당 broker가 leader인 partition의 produce latency가 늘고 follower fetch가 밀리며 ISR 변화가 생길 수 있습니다. consumer lag도 늘 수 있지만, lag의 직접 원인이 consumer인지 broker disk인지 분리해야 합니다.

Cassandra에서는 commit log fsync, memtable flush, compaction이 모두 영향을 받습니다. write latency와 pending compaction이 올라가고, read는 여러 SSTable을 더 오래 읽을 수 있습니다.

Spark에서는 shuffle spill과 shuffle read가 느려지고, 그 node의 task만 길어질 수 있습니다. Spark UI에서 task duration distribution을 보면 특정 executor에 긴 task가 몰릴 수 있습니다.

같은 "disk 느림"이라도 다음 trace가 다릅니다.

```
Kafka:     append/fetch log segment -> replica lag / consumer lag
Cassandra: commit log / SSTable / compaction -> read/write latency
Spark:     spill/shuffle block -> stage skew / task retry
```

## 현실 시나리오 2: retry가 장애를 회복하지 않고 증폭한다

Kafka producer retry가 idempotence 없이 중복 append를 만들 수 있습니다. Cassandra client retry가 timeout된 write의 적용 여부를 모른 채 다시 쓰면 last-write-wins와 timestamp conflict를 복잡하게 만들 수 있습니다. Spark task retry는 pure computation에는 유용하지만 external side effect를 가진 `foreachPartition` 같은 코드에서는 중복 write를 만들 수 있습니다.

retry는 실패를 없애는 기능이 아니라 불확실한 상태에서 다시 시도하는 정책입니다. 그래서 request id, idempotent sink, transaction boundary, rate limit, backoff, circuit breaker가 함께 필요합니다.

## 현실 시나리오 3: network timeout이 세 시스템에서 다르게 보인다

같은 `timeout`이라는 로그라도 내부 상태는 다릅니다.

```
common lower layer:
  packet loss or delayed processing
    -> TCP retransmission or socket queue backlog
    -> application thread wakeup delayed
    -> request timeout observed
```

Kafka에서는 producer request가 broker ack를 못 받았거나 consumer fetch가 늦은 상태일 수 있습니다. Cassandra에서는 coordinator가 CL을 만족할 replica 응답을 제때 받지 못한 상태일 수 있습니다. Spark에서는 reduce task가 shuffle block fetch를 기다리다 실패한 상태일 수 있습니다. 세 경우 모두 [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md)의 network path를 통과하지만, 위쪽의 recovery 의미는 다릅니다. Kafka는 producer retry/idempotence/offset, Cassandra는 적용 여부 불명확성과 CL/repair, Spark는 task retry와 external side effect를 각각 봐야 합니다.

## 문서를 덮고 확인할 것

- Kafka log, Cassandra commit log, Spark lineage를 같은 점과 다른 점으로 설명해 보세요.
- partition이 세 시스템에서 각각 어떤 약속의 경계인지 말해 보세요.
- "replication이 있으니 안전하다"라는 말을 시스템별로 반박해 보세요.
- disk saturation이 세 시스템에서 어떤 metric과 증상으로 다르게 나타나는지 설명해 보세요.
- retry가 중복과 backpressure를 만들 수 있는 경로를 세 시스템에서 하나씩 들어 보세요.

## 근거와 더 읽을 자료

- Kafka design docs: log, partition, replication, consumer offset, filesystem/page cache.
- Cassandra architecture/storage/guarantees docs: ring, RF/CL, commit log, SSTable, repair.
- Spark cluster/RDD/tuning docs: driver/executor, partition, shuffle, lineage, memory.
- Linux page cache, `write(2)`, `sendfile(2)`, networking docs for shared OS layer.
