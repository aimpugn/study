# 06. Kafka, Cassandra, Spark Cross-System Comparison

## 목차

- [1. 세 시스템의 한 문장 모델](#1-세-시스템의-한-문장-모델)
- [2. Log는 세 시스템에서 서로 다르게 산다](#2-log는-세-시스템에서-서로-다르게-산다)
- [3. Partition은 ordering, ownership, computation의 경계다](#3-partition은-ordering-ownership-computation의-경계다)
- [4. Replication과 recovery는 목적이 다르다](#4-replication과-recovery는-목적이-다르다)
- [5. Consistency 질문은 시스템별 read/write 계약으로 바꿔야 한다](#5-consistency-질문은-시스템별-readwrite-계약으로-바꿔야-한다)
- [6. Backpressure는 각 시스템의 queue 이름으로 나타난다](#6-backpressure는-각-시스템의-queue-이름으로-나타난다)
- [7. OS resource 관점의 공통 지도](#7-os-resource-관점의-공통-지도)
- [현실 시나리오 1: 한 노드의 disk가 느려졌다](#현실-시나리오-1-한-노드의-disk가-느려졌다)
- [현실 시나리오 2: retry가 장애를 회복하지 않고 증폭한다](#현실-시나리오-2-retry가-장애를-회복하지-않고-증폭한다)
- [현실 시나리오 3: network timeout이 세 시스템에서 다르게 보인다](#현실-시나리오-3-network-timeout이-세-시스템에서-다르게-보인다)
- [같은 단어, 다른 소유자](#같은-단어-다른-소유자)
- [Durability boundary 비교](#durability-boundary-비교)
- [Backpressure 비교](#backpressure-비교)
- [Ordering 비교](#ordering-비교)
- [Recovery 비교](#recovery-비교)
- [Resource pressure 비교](#resource-pressure-비교)
- [같은 장애를 세 시스템에 투영하기](#같은-장애를-세-시스템에-투영하기)
- [Interview replay: 비교 답변의 구조](#interview-replay-비교-답변의-구조)
- [State ownership 비교](#state-ownership-비교)
- [Security와 isolation 비교](#security와-isolation-비교)
- [한 문장 비교 훈련](#한-문장-비교-훈련)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [근거와 더 읽을 자료](#근거와-더-읽을-자료)

Kafka, Cassandra, Spark는 모두 partition, log, replication, recovery, backpressure 같은 단어를 씁니다. 하지만 같은 단어가 같은 약속을 뜻하지 않습니다. 이 문서의 목적은 세 시스템을 표로 외우는 것이 아니라, 같은 lower-layer 질문을 던졌을 때 각 시스템이 어디에서 다른 선택을 하는지 보는 것입니다.

가장 중요한 비교 질문은 다음 하나입니다.

```text
What state is moving?
  -> who owns it?
  -> how is order defined?
  -> how is it copied or recomputed?
  -> where can it be lost or delayed?
  -> what observation proves recovery?
```

## 1. 세 시스템의 한 문장 모델

| 시스템 | 중심 상태 | 주된 약속 | 대가 |
| --- | --- | --- | --- |
| Kafka | partition log record와 consumer offset | 순서 있는 append, replay, fan-out consumption | partition 간 전역 순서 없음, lag/retention/replication 관리 필요 |
| Cassandra | partitioned mutation과 replica cell version | 높은 write availability, tunable consistency, scale-out storage | read/repair/compaction 복잡도, stale read 가능성 |
| Spark | partitioned computation과 lineage | 큰 계산을 task로 나누고 실패한 조각을 재계산 | shuffle/spill/skew/driver metadata 비용 |

Kafka는 data를 오래 읽는 log로 둡니다. Cassandra는 latest queryable state를 여러 SSTable과 replica에서 조립합니다. Spark는 data 자체보다 계산 graph와 partition을 중심으로 움직입니다.

## 2. Log는 세 시스템에서 서로 다르게 산다

Kafka의 log는 사용자에게 직접 보이는 중심 데이터 구조입니다. consumer는 offset을 지정해 log를 읽고, retention이 허용하는 동안 다시 읽을 수 있습니다. Kafka에서 log는 "서비스가 제공하는 데이터"입니다.

Cassandra의 commit log는 사용자 query 모델이 아닙니다. crash recovery를 위한 write-ahead log입니다. acknowledged mutation이 memtable에만 있고 node가 죽었을 때 commit log replay로 복원합니다. 실제 read는 memtable과 SSTable을 merge합니다.

Spark의 lineage는 파일 log가 아니라 계산의 기억입니다. 어떤 partition이 어떤 parent partition과 transformation으로 만들어졌는지를 기억해 실패한 partition을 다시 계산합니다. checkpoint를 만들면 lineage의 일부를 끊고 안정적인 저장소를 기준점으로 삼습니다.

```text
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
| --- | --- | --- | --- |
| 같은 partition 안에서 중요한 것 | offset order | row clustering / local read shape | task-local processing |
| partition이 많아지면 | parallelism 증가, metadata/replication 비용 증가 | distribution 개선 가능, repair/operation 복잡도 증가 | parallelism 증가, scheduler/shuffle overhead 증가 |
| 잘못 고르면 | hot partition, consumer 병렬성 한계 | hot key, wide partition, bad query | skew task, tiny task, spill |

## 4. Replication과 recovery는 목적이 다르다

Kafka replication은 partition log를 follower에게 복사해 leader failure에 대비합니다. 중요한 질문은 어떤 offset이 in-sync replica에 있고 consumer에게 노출 가능한가입니다.

Cassandra replication은 같은 partition data를 여러 node에 두어 availability와 latency를 조절합니다. 중요한 질문은 write CL과 read CL이 어떤 replica 집합을 만들고, repair가 불일치를 어떻게 줄이는가입니다.

Spark는 일반적인 data replication보다 recomputation을 먼저 활용합니다. executor가 죽어 cache나 shuffle block을 잃으면 lineage를 따라 다시 계산할 수 있습니다. 단, shuffle output 손실과 긴 lineage, nondeterministic computation은 checkpoint나 external storage가 필요할 수 있습니다.

```text
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
| --- | --- | --- |
| Kafka | consumer lag, request queue, replica lag | broker disk/page cache, network, partition skew, downstream DB |
| Cassandra | pending compaction, write latency, dropped messages, read latency | commit log disk, memtable flush, SSTable count, tombstone, repair traffic |
| Spark | long stage, skewed task, spill size, shuffle fetch wait | executor heap/off-heap, local disk, network, GC, partition skew |

backpressure를 이해하지 못하면 worker 수나 thread 수를 늘리는 처방으로 장애를 키울 수 있습니다. downstream DB가 느린데 Kafka consumer를 늘리면 DB timeout과 retry가 늘 수 있습니다. Cassandra compaction이 disk를 잡아먹는데 read timeout만 늘리면 queue가 더 오래 쌓입니다. Spark skew가 있는데 executor만 늘리면 느린 partition 하나가 여전히 job을 붙잡습니다.

## 7. OS resource 관점의 공통 지도

| OS 자원 | Kafka | Cassandra | Spark |
| --- | --- | --- | --- |
| CPU | compression, request handling, TLS, consumer processing | read merge, compaction, serialization, GC | task execution, serialization, GC, SQL execution |
| Memory | broker buffer, page cache | memtable, cache, heap/off-heap, page cache | executor heap, storage/execution memory, broadcast, shuffle buffer |
| Disk | log segment append/read, retention/compaction | commit log, SSTable, compaction, repair streaming | shuffle spill, cache on disk, checkpoint |
| Network | producer/fetch traffic, replica fetch | replica read/write, repair, gossip | shuffle fetch, driver/executor RPC, data source reads |

이 표는 "어디를 봐야 하는가"를 정하기 위한 지도입니다. 예를 들어 disk saturation이 있으면 Kafka는 leader append와 follower fetch, Cassandra는 compaction과 commit log, Spark는 shuffle spill을 먼저 떠올립니다. 같은 disk라도 시스템마다 만들어내는 증상이 다릅니다.

OS 상세 문서와 연결하면 비교 질문이 더 구체적입니다.

| OS 상세 경로 | Kafka에서 묻는 질문 | Cassandra에서 묻는 질문 | Spark에서 묻는 질문 |
| --- | --- | --- | --- |
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

```text
Kafka:     append/fetch log segment -> replica lag / consumer lag
Cassandra: commit log / SSTable / compaction -> read/write latency
Spark:     spill/shuffle block -> stage skew / task retry
```

## 현실 시나리오 2: retry가 장애를 회복하지 않고 증폭한다

Kafka producer retry가 idempotence 없이 중복 append를 만들 수 있습니다. Cassandra client retry가 timeout된 write의 적용 여부를 모른 채 다시 쓰면 last-write-wins와 timestamp conflict를 복잡하게 만들 수 있습니다. Spark task retry는 pure computation에는 유용하지만 external side effect를 가진 `foreachPartition` 같은 코드에서는 중복 write를 만들 수 있습니다.

retry는 실패를 없애는 기능이 아니라 불확실한 상태에서 다시 시도하는 정책입니다. 그래서 request id, idempotent sink, transaction boundary, rate limit, backoff, circuit breaker가 함께 필요합니다.

## 현실 시나리오 3: network timeout이 세 시스템에서 다르게 보인다

같은 `timeout`이라는 로그라도 내부 상태는 다릅니다.

```text
common lower layer:
  packet loss or delayed processing
    -> TCP retransmission or socket queue backlog
    -> application thread wakeup delayed
    -> request timeout observed
```

Kafka에서는 producer request가 broker ack를 못 받았거나 consumer fetch가 늦은 상태일 수 있습니다. Cassandra에서는 coordinator가 CL을 만족할 replica 응답을 제때 받지 못한 상태일 수 있습니다. Spark에서는 reduce task가 shuffle block fetch를 기다리다 실패한 상태일 수 있습니다. 세 경우 모두 [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md)의 network path를 통과하지만, 위쪽의 recovery 의미는 다릅니다. Kafka는 producer retry/idempotence/offset, Cassandra는 적용 여부 불명확성과 CL/repair, Spark는 task retry와 external side effect를 각각 봐야 합니다.

## 같은 단어, 다른 소유자

Kafka, Cassandra, Spark를 비교할 때 가장 먼저 해야 할 일은 같은 단어가 가리키는 소유자를 나누는 것입니다. Partition이라는 단어는 세 시스템에 모두 나오지만 의미가 다릅니다. Kafka partition은 log ordering과 consumer group assignment의 단위입니다. Cassandra partition은 partition key로 결정되는 storage row group이며 token range와 replica ownership으로 이어집니다. Spark partition은 distributed dataset의 계산 단위이고 task scheduling과 shuffle의 단위입니다. 단어가 같다고 같은 보장을 뜻하지 않습니다.

```text
Kafka partition
  -> ordered append log within a topic
  -> offset sequence and leader/follower replication

Cassandra partition
  -> rows sharing a partition key
  -> replica placement and read/write lookup unit

Spark partition
  -> slice of a dataset
  -> task input and shuffle distribution unit
```

Log도 마찬가지입니다. Kafka log는 사용자가 직접 읽는 data structure입니다. Cassandra commitlog는 crash recovery를 위한 write-ahead log이고, query path의 primary storage는 SSTable/memtable입니다. Spark lineage는 byte log가 아니라 계산 graph입니다. Streaming checkpoint는 source offset과 state를 포함한 recovery boundary입니다. "로그 기반이라 복구된다"는 문장은 각 시스템에서 전혀 다른 의미를 갖습니다.

## Durability boundary 비교

내구성을 비교할 때는 user-visible success, local persistence, replica persistence, recovery visibility를 나눠야 합니다. Kafka produce ack는 `acks`와 ISR 조건에 따라 의미가 달라집니다. Cassandra write success는 consistency level과 replica response, commitlog sync policy에 묶입니다. Spark action 성공은 job output commit protocol과 sink semantics에 묶이고, streaming은 checkpoint와 sink idempotency를 함께 봐야 합니다.

| System | Success boundary | Local OS boundary | Distributed boundary | Caveat |
| --- | --- | --- | --- | --- |
| Kafka | produce ack by `acks` policy | log append/page cache/flush policy | ISR/high watermark/leader epoch | external side effects and unclean election require care |
| Cassandra | coordinator receives enough CL responses | commitlog/memtable/SSTable flush | RF/CL/repair/gossip convergence | QUORUM is not blanket linearizability |
| Spark | job/stage/task or streaming batch completes | shuffle/checkpoint/local disk files | lineage recomputation, checkpoint, sink commit | external sink duplicate unless idempotent |

이 표에서 중요한 점은 어느 system이 더 좋다는 순위가 아닙니다. 각 system이 해결하려는 문제가 다르므로 내구성 경계도 다릅니다. Kafka는 append log와 replay가 중심이고, Cassandra는 always-on replicated storage가 중심이며, Spark는 distributed computation과 recomputation이 중심입니다.

## Backpressure 비교

Backpressure는 세 시스템 모두에 있지만 신호가 다릅니다. Kafka에서는 producer throughput, broker request queue, replica lag, consumer lag, quota가 신호입니다. Cassandra에서는 pending tasks, dropped messages, compaction backlog, coordinator timeout, repair/streaming backlog가 신호입니다. Spark에서는 scheduler delay, shuffle wait, source rate, streaming batch duration, executor backlog가 신호입니다.

```text
common shape
  producer faster than consumer
  -> queue grows
  -> latency grows
  -> timeout/retry begins
  -> useful throughput may drop
```

Kafka에서 consumer lag가 커질 때 producer를 멈추거나 consumer 처리량을 늘릴 수 있습니다. Cassandra에서 coordinator timeout이 늘 때 client retry를 줄이고 admission control을 걸어야 할 수 있습니다. Spark streaming에서 batch duration이 trigger interval보다 길어지면 source rate limit이나 state/shuffle 최적화가 필요합니다. Backpressure를 이해하지 못하면 retry를 "회복"으로 보고 실제로는 overload를 키우게 됩니다.

## Ordering 비교

Kafka는 partition 안 offset order가 중심입니다. Cassandra는 일반 write에서 전역 order를 만들지 않고 timestamp와 reconciliation rule을 사용합니다. Spark는 dataset partition과 stage/task execution order가 있지만, distributed record processing의 전역 order를 기본으로 보장하지 않습니다. Ordering 질문은 항상 "어느 범위 안에서"라고 물어야 합니다.

```text
Kafka
  same partition: ordered offsets
  across partitions: no default total order

Cassandra
  per cell/row reconciliation by timestamp/version rules
  concurrent writes depend on conflict resolution

Spark
  partition-level computation
  shuffle can reorder data
  explicit sort/window required for order-sensitive results
```

이 차이는 interview에서 자주 드러납니다. "Kafka는 순서를 보장하나요?"라는 질문의 답은 "partition 안에서는 offset order를 보장하지만 topic 전체 partition 사이 전역 순서는 기본으로 보장하지 않습니다"입니다. "Cassandra는 최신 값을 보장하나요?"는 RF/CL, LWT, timestamp, repair를 봐야 합니다. "Spark 결과는 항상 같은 순서인가요?"는 operation과 partitioning, sort 여부를 봐야 합니다.

## Recovery 비교

Kafka recovery는 log와 offset, replica state를 중심으로 합니다. Broker는 log segment를 recover하고 follower는 leader를 따라잡으며 consumer는 committed offset부터 다시 읽습니다. Cassandra recovery는 commitlog replay, memtable/SSTable, repair를 중심으로 합니다. Spark recovery는 failed task recomputation, lost executor recovery, lineage, checkpoint를 중심으로 합니다.

```text
Kafka restart
  -> recover local logs
  -> rejoin cluster
  -> leader/follower catch-up
  -> consumers resume by offset

Cassandra restart
  -> replay commitlog
  -> expose SSTables/memtables
  -> gossip membership
  -> repair for divergence

Spark recovery
  -> reschedule failed tasks
  -> recompute lost partitions by lineage
  -> use checkpoint for state/lineage cut
```

OS 관점에서는 모두 process restart, file recovery, network reconnect, scheduler, page cache warmup을 겪습니다. 분산 관점에서는 membership view, ownership, replay boundary가 다릅니다. 그래서 recovery time을 줄이려면 각 system의 기준점을 알아야 합니다. Kafka는 log segment와 replica catch-up, Cassandra는 commitlog/repair/compaction, Spark는 lineage length/checkpoint/shuffle persistence를 봅니다.

## Resource pressure 비교

세 시스템의 resource profile은 다릅니다. Kafka는 disk-friendly append와 network fetch, page cache가 중요합니다. Cassandra는 write path는 빠르게 받지만 compaction과 read amplification, memory/cache 균형이 중요합니다. Spark는 CPU, memory, network, disk를 job shape에 따라 폭발적으로 바꿉니다.

| Resource | Kafka | Cassandra | Spark |
| --- | --- | --- | --- |
| CPU | compression, network thread, request handler | compaction, serialization, read merge, GC | task compute, serialization, compression, GC |
| Memory | page cache, socket buffer, broker heap | memtable, cache, off-heap, page cache | executor heap/off-heap, cache, shuffle buffers |
| Disk | append log, segment read, flush | commitlog, SSTable, compaction | shuffle spill, cache/checkpoint/local dirs |
| Network | producer/fetch/replica | coordinator/replica/gossip/repair | shuffle fetch, driver/executor, remote reads |

이 표는 tuning의 방향을 정해 줍니다. Kafka에서 heap을 무작정 키우면 page cache 여유를 줄일 수 있습니다. Cassandra에서 compaction을 억제하면 당장 disk pressure는 줄어도 read amplification이 커질 수 있습니다. Spark에서 executor memory를 늘리면 spill은 줄지만 GC와 cgroup pressure가 커질 수 있습니다.

## 같은 장애를 세 시스템에 투영하기

Disk latency가 갑자기 증가했다고 합시다. Kafka에서는 produce append, flush, fetch from cold segment, replica catch-up이 느려질 수 있습니다. Cassandra에서는 commitlog sync, SSTable read, compaction, memtable flush가 느려질 수 있습니다. Spark에서는 shuffle spill/write/read, checkpoint, local cache read가 느려질 수 있습니다. 같은 하위 계층 이벤트(event)가 서로 다른 product metric으로 올라옵니다.

Network loss가 증가하면 Kafka producer/fetch timeout, replica lag, consumer lag가 나타날 수 있습니다. Cassandra에서는 coordinator timeout, gossip suspicion, repair/streaming slowdown이 나타납니다. Spark에서는 shuffle fetch failure, executor lost suspicion, task retry가 늘 수 있습니다. Memory pressure가 커지면 Kafka page cache miss와 GC, Cassandra GC/compaction/cache miss, Spark spill/GC/OOM으로 다르게 나타납니다.

이 비교의 목적은 "세 시스템을 한 줄로 요약"하는 것이 아닙니다. 하나의 OS 이벤트(event)가 각 제품의 어떤 queue, log, replica, task로 올라오는지 보는 능력을 만드는 것입니다.

## Interview replay: 비교 답변의 구조

비교 질문을 받으면 먼저 같은 추상화 수준을 맞춥니다. "Kafka와 Cassandra는 둘 다 분산 저장 계층을 갖지만 Kafka는 append log와 stream replay가 중심이고, Cassandra는 partitioned replicated table storage가 중심입니다. Spark는 저장소라기보다 distributed computation runtime이며 lineage와 shuffle, checkpoint가 중심입니다." 이렇게 시작하면 제품 범주가 섞이지 않습니다.

그 다음 같은 축으로 비교합니다. Data placement는 Kafka partition, Cassandra token/replica, Spark partition입니다. Recovery는 Kafka offset/log, Cassandra commitlog/repair, Spark lineage/checkpoint입니다. Durability는 Kafka ack/ISR, Cassandra CL/commitlog, Spark sink/checkpoint입니다. OS pressure는 Kafka page cache/network, Cassandra compaction/disk/cache, Spark shuffle/memory/disk입니다. 이 축을 유지하면 답변이 길어져도 흐트러지지 않습니다.

## State ownership 비교

세 시스템은 상태를 소유하는 방식도 다릅니다. Kafka에서 source of truth는 partition log입니다. Consumer의 local state는 offset과 처리 결과를 기준으로 다시 만들 수 있어야 합니다. Cassandra에서 source of truth는 replica들이 가진 partition data와 그 수렴을 돕는 repair 체계입니다. Spark에서 많은 중간 state는 lineage로 다시 만들 수 있지만, checkpoint나 external sink로 나간 state는 별도 소유자가 생깁니다.

```text
Kafka
  durable stream state -> broker partition logs
  consumer progress -> committed offsets

Cassandra
  table state -> replicated SSTables/memtables per token range
  convergence -> repair/gossip/hints

Spark
  computation state -> lineage and task attempts
  materialized state -> cache/checkpoint/sink
```

State owner를 모르면 복구 책임을 잘못 둡니다. Kafka consumer가 외부 DB에 쓴 뒤 offset commit 전에 죽으면 broker log는 안전해도 sink 중복 문제가 남습니다. Cassandra replica 하나가 오래 내려가면 살아 있는 replica의 상태와 hint/repair 책임을 봐야 합니다. Spark task가 실패하면 lineage로 재계산할 수 있지만, 이미 외부 API를 호출했다면 그 side effect는 Spark lineage가 소유하지 않습니다.

## Security와 isolation 비교

운영 환경에서는 세 시스템 모두 container, namespace, cgroup, network policy, TLS, authentication/authorization 아래에서 실행됩니다. Kafka는 listener, ACL, TLS/SASL, inter-broker auth가 중요합니다. Cassandra는 client/internode encryption, authentication, authorization, network exposure, node identity가 중요합니다. Spark는 driver/executor communication, UI exposure, shuffle service, secret propagation, Kubernetes/YARN permission이 중요합니다.

OS isolation은 security와 performance를 동시에 바꿉니다. cgroup CPU quota는 Kafka network thread와 Cassandra compaction, Spark task를 모두 throttle할 수 있습니다. Network namespace와 policy는 broker/client, coordinator/replica, driver/executor path를 막을 수 있습니다. Seccomp와 capability는 observability 도구를 제한합니다. 보안 설정을 성능 문제에서 분리하면 "왜 tcpdump가 안 되지", "왜 perf가 안 되지", "왜 inter-broker connection이 timeout이지" 같은 질문을 놓칩니다.

## 한 문장 비교 훈련

Kafka는 "순서 있는 partition log를 중심으로 replay와 fan-out을 제공하는 시스템"입니다. Cassandra는 "partition key와 replica quorum을 중심으로 항상 켜져 있는 write/read storage를 제공하는 시스템"입니다. Spark는 "lineage와 task scheduling, shuffle을 중심으로 큰 계산을 분산 실행하는 runtime"입니다. 이 세 문장을 먼저 말하고, 그 뒤 OS 자원과 분산 보장을 붙이면 비교가 안정됩니다.

```text
Kafka answer expands toward:
  log, offset, ISR, page cache, sendfile, consumer group

Cassandra answer expands toward:
  token, replica, CL, commitlog, SSTable, compaction, repair

Spark answer expands toward:
  DAG, stage, task, shuffle, lineage, checkpoint, executor resource
```

마지막 확인은 "같은 해결책을 세 시스템에 기계적으로 적용하지 않는가"입니다. Kafka lag에 consumer 수를 늘리는 감각을 Cassandra read timeout이나 Spark skew에 그대로 가져오면 틀릴 수 있습니다. Cassandra compaction 조절 감각을 Kafka log cleaner나 Spark shuffle cleanup에 그대로 옮겨도 안 됩니다. 비교 문서의 목적은 차이를 외우는 것이 아니라, 같은 하위 계층 이벤트(event)가 각 시스템의 다른 의미 경계로 올라간다는 점을 몸에 익히는 것입니다.

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
