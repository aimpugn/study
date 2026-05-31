# 06. Cross-System Comparison

> Kafka, Cassandra, Spark는 모두 partition, log, recovery, backpressure를 쓰지만 같은 문제를 푸는 시스템은 아닙니다.
> Kafka는 record stream을 오래 읽을 수 있는 log로 만들고, Cassandra는 key-value/table mutation을 여러 replica에 수렴시키며, Spark는 큰 계산을 partitioned task graph로 실행합니다.
> 공통 원리를 비교할 때는 "무엇이 상태인가", "순서를 누가 정하는가", "어디까지 실패 후 다시 시작할 수 있는가", "어떤 OS 자원이 병목인가"를 함께 봐야 합니다.

## 1. 한눈에 보는 비교표

> 같은 용어라도 Kafka partition, Cassandra partition, Spark partition은 서로 다른 상태와 실행 경계를 뜻합니다.
> 표는 암기용이 아니라 꼬리 질문에서 어떤 lower layer로 내려갈지 정하는 지도입니다.

| 관점 | Kafka | Cassandra | Spark |
|---|---|---|---|
| 주된 목적 | event/log stream 저장과 전달 | partitioned wide-column data 저장 | 분산 data processing |
| 기본 상태 | partition log records | partitioned rows/cells | partitioned datasets and execution plan |
| partition 의미 | 순서와 consumer parallelism 경계 | token range/partition key data placement 경계 | task parallelism and data split 경계 |
| log 의미 | user-visible topic partition log | commit log and SSTable history | lineage/event logs/checkpoints |
| 순서 기준 | partition offset | timestamp/cell reconciliation, clustering order | DAG dependency, stage/task order |
| 복제 방식 | leader-follower replica, ISR | RF, multi-replica write/read, repair | 기본 data replica보다 recomputation/cache/checkpoint 중심 |
| 일관성 모델 | per-partition ordered log, producer/transaction semantics | tunable consistency, eventual convergence, LWT exception | deterministic recomputation and output/sink semantics |
| 회복 단위 | replica log, consumer offset | commit log replay, SSTable/repair | lineage recomputation, checkpoint |
| backpressure 신호 | consumer lag, producer buffer, request queue | pending compaction, read/write latency, dropped messages | slow stage, spill, executor lost, input/processed rows gap |
| OS 병목 | page cache, sequential disk I/O, socket, TLS CPU | commit log disk, SSTable read/write, compaction, GC, page cache | executor heap, GC, network shuffle, local disk spill |

## 2. Kernel 지식이 세 시스템으로 연결되는 방식

> 커널 지식은 "OS 면접 과목"이 아니라 분산 시스템 증상을 해석하는 하부 언어입니다.
> page cache, scheduler, virtual memory, socket buffer, disk queue를 모르면 Kafka lag, Cassandra stale read, Spark spill의 원인 후보를 제대로 나눌 수 없습니다.

### 질문

왜 Kafka/Cassandra/Spark를 공부하는데 syscall, page cache, scheduling을 알아야 할까요?

### 직관

분산 시스템은 여러 건물의 엘리베이터를 연결한 것과 같습니다. 건물 사이 교통도 중요하지만 각 건물 안의 엘리베이터, 전기, 창고가 막히면 전체 물류가 느려집니다.

### 작은 예시

같은 "latency 증가" 증상도 시스템별로 아래처럼 내려갑니다.

```
Kafka fetch latency up
  -> broker page cache miss?
  -> disk read await?
  -> socket send buffer?

Cassandra read latency up
  -> SSTable count?
  -> compaction disk pressure?
  -> GC pause?

Spark stage duration up
  -> shuffle spill?
  -> executor GC?
  -> skewed task reading remote data?
```

### 상태 이동 trace

```
application metric
  |
  v
system internal state
  |
  v
OS resource queue
  |
  v
hardware / network / storage behavior

Example:
Kafka consumer lag
  -> consumer processing rate below input rate
  -> maybe downstream DB socket wait
  -> maybe kernel send/recv queue and remote disk wait
```

### 내부 메커니즘

Kafka의 log segment는 file이고 normal I/O는 page cache를 거칩니다. Cassandra의 SSTable과 commit log도 file이며 compaction은 disk read/write와 CPU를 씁니다. Spark executor는 JVM process이며 shuffle은 local disk와 network socket을 씁니다. 따라서 제품 metric은 OS metric으로 내려가고, OS metric은 다시 제품 상태로 해석되어야 합니다.

### 실패 모드

- 제품 dashboard만 보면 lower-layer saturation을 놓칩니다.
- OS metric만 보면 어느 partition, table, stage가 문제인지 모릅니다.
- 평균 CPU와 평균 latency만 보면 tail latency와 skew를 놓칩니다.

### 검증 방법

공통 triage:

```bash
vmstat 1 5
iostat -xz 1 5
ss -tin
jcmd <pid> Thread.print
```

PASS 신호:

- 제품 metric과 OS metric의 시간대가 맞습니다.

FAIL 신호:

- 서로 맞지 않는 관측을 억지로 한 원인에 끼워 맞춥니다.

### 면접식 되묻기

"Kafka가 느린데 OS에서 뭘 보나요?"라는 질문에는 Kafka 내부의 어느 path인지 먼저 묻고, broker disk/page cache/socket, consumer CPU/GC/downstream, replica lag를 나눠 본다고 답합니다.

### 흔한 오해와 반례

오해: "분산 시스템 문제는 cluster 설정 문제다."

반례: cluster 설정이 맞아도 한 node의 disk, GC, network queue가 분산 상태를 밀리게 만들 수 있습니다.

### Active recall

- page cache가 Kafka와 Cassandra에 각각 어떤 방식으로 영향을 주나요?
- Spark shuffle spill과 Cassandra compaction은 OS 관점에서 어떤 공통점이 있나요?

## 3. Log, Partitioning, Replication, Recovery 비교

> 세 시스템의 공통 단어는 같아도 상태 전이의 주체가 다릅니다.
> Kafka는 log를 사용자-visible stream으로 노출하고, Cassandra는 mutation과 SSTable history로 수렴하며, Spark는 계산 lineage로 recovery합니다.

### 질문

Kafka log, Cassandra commit log, Spark lineage는 모두 log라고 불러도 될까요?

### 직관

세 가지 모두 "과거를 남겨 미래를 복원한다"는 공통점이 있습니다. 그러나 Kafka log는 독자가 직접 따라가는 원본 장부이고, Cassandra commit log는 crash recovery용 내부 장부이며, Spark lineage는 계산 recipe입니다.

### 작은 예시

| 질문 | Kafka | Cassandra | Spark |
|---|---|---|---|
| 어디까지 처리했나? | consumer offset | CL ack, commit log replay point, repair state | completed task/stage, checkpoint |
| 다시 시작하면? | committed offset부터 fetch | commit log replay + SSTable state + repair | lineage recompute or checkpoint |
| 순서가 중요한 곳? | partition offset | per-cell timestamp, clustering order | DAG dependency, shuffle stage |

### 상태 이동 trace

```
Kafka:
produce -> append offset -> replicate -> consume -> commit offset

Cassandra:
mutation -> commit log/memtable -> SSTable -> compaction/repair

Spark:
transformation plan -> task -> shuffle/cache -> lineage/checkpoint recovery
```

### 내부 메커니즘

Kafka recovery는 log offset과 replica high watermark를 중심으로 합니다. Cassandra recovery는 commit log replay와 SSTable immutable state, replica repair를 중심으로 합니다. Spark recovery는 lost partition을 parent partition과 transformation으로 다시 계산하는 방식을 중심으로 합니다. 세 시스템 모두 checkpoint나 compaction으로 무한히 긴 history를 줄입니다.

### 실패 모드

- Kafka offset을 DB commit처럼 오해하면 processing side effect 경계를 놓칩니다.
- Cassandra commit log를 user query history로 오해하면 read model을 잘못 이해합니다.
- Spark lineage를 data durability로 오해하면 external sink와 checkpoint 필요성을 놓칩니다.

### 검증 방법

각 시스템에서 "재시작 지점"을 찾아봅니다.

```text
Kafka: consumer group committed offset
Cassandra: commit log replay + SSTable files + repair state
Spark: checkpoint directory + event log + lineage/stage retry
```

PASS 신호:

- 각 시스템의 restart point가 서로 다름을 설명합니다.

FAIL 신호:

- "log가 있으니 모두 같은 방식으로 복구된다"고 말합니다.

### 면접식 되묻기

"왜 세 시스템 모두 log를 쓰나요?"에는 failure 후 재구성, 순서 보존, append 효율 때문이라고 답합니다. 이어 "하지만 각 log의 소비자와 보장 범위가 다르다"고 정리합니다.

### 흔한 오해와 반례

오해: "log는 항상 사용자가 읽는 event stream이다."

반례: DB commit log는 사용자에게 직접 노출되지 않는 internal recovery structure인 경우가 많습니다.

### Active recall

- Kafka log와 Cassandra commit log의 소비자는 누구인가요?
- Spark lineage는 어떤 의미에서 log와 비슷하고 어떤 의미에서 다른가요?

## 4. 장애, 성능, 일관성, 확장성 관점 비교

> 장애는 "죽음"보다 "어느 상태가 더 이상 진행되지 않는가"로 봐야 합니다.
> 성능은 처리량만이 아니라 queue growth와 tail latency를 봐야 합니다.
> 일관성과 확장성은 서로 독립이 아니라 partition, replication, coordination 비용을 통해 연결됩니다.

### 질문

세 시스템에서 "scale-out"은 같은 뜻인가요?

### 직관

의자를 늘리는 것과 계산대를 늘리는 것과 창고를 늘리는 것은 모두 "늘린다"지만 병목이 다릅니다. Kafka는 partition/broker/consumer group, Cassandra는 token range/node/RF, Spark는 executor/core/partition이 함께 맞아야 합니다.

### 작은 예시

| 증상 | Kafka에서 의심 | Cassandra에서 의심 | Spark에서 의심 |
|---|---|---|---|
| latency up | broker disk, ISR, consumer lag | compaction, quorum wait, GC | skew, shuffle, GC, spill |
| stale/duplicate | offset commit, transaction boundary | CL/timestamp/repair | checkpoint/sink idempotency |
| scale-out 안 됨 | partition 부족/hot partition | hot partition/wide partition | partition 부족/skew |

### 상태 이동 trace

```
scale-out decision
  |
  +-- is work partitionable?
  +-- is partition distribution balanced?
  +-- does coordination increase?
  +-- does OS resource move to another bottleneck?
  +-- does consistency expectation still hold?
```

### 내부 메커니즘

Kafka에서 broker를 늘려도 topic partition leadership과 consumer parallelism이 안 맞으면 효과가 제한됩니다. Cassandra에서 node를 늘려도 hot partition query는 한 replica set에 몰릴 수 있습니다. Spark에서 executor를 늘려도 stage task 수가 적거나 skewed task 하나가 크면 전체 시간이 줄지 않습니다.

### 실패 모드

- scale-out을 storage 용량 문제와 processing 병렬성 문제에 같은 처방으로 씁니다.
- consistency 요구가 강해질수록 coordination 비용이 커지는 점을 놓칩니다.
- 장애 대응으로 retry를 늘려 backpressure를 악화시킵니다.

### 검증 방법

처방 전 질문:

```text
1. 어떤 partition/key/stage가 병목인가?
2. 병목이 CPU, memory, disk, network, coordination 중 무엇인가?
3. 새 node/consumer/executor가 그 병목을 실제로 나눌 수 있는가?
4. ordering/consistency/recovery 경계가 바뀌는가?
```

PASS 신호:

- scaling action이 병목 queue를 직접 줄입니다.

FAIL 신호:

- 병목 partition은 그대로인데 node 수만 늘립니다.

### 면접식 되묻기

"Kafka/Cassandra/Spark 모두 partition으로 scale-out하나요?"에는 "그렇지만 partition이 나누는 대상이 다르다"고 답합니다. Kafka는 log/order/consumer parallelism, Cassandra는 data placement/query locality, Spark는 task/data execution parallelism입니다.

### 흔한 오해와 반례

오해: "분산 시스템은 노드를 늘리면 선형 확장한다."

반례: skew, coordination, network shuffle, compaction, leader bottleneck, failure domain 때문에 선형 확장은 목표일 뿐 자동 결과가 아닙니다.

### Active recall

- Kafka partition 부족과 Spark partition 부족은 각각 어떤 증상으로 나타나나요?
- Cassandra node 추가가 hot partition을 해결하지 못하는 이유를 설명해 보세요.

## 현실 시나리오 1: 한 노드의 disk saturation이 세 시스템에 다르게 나타난다

> 같은 disk saturation도 Kafka에서는 follower/consumer lag, Cassandra에서는 compaction/read-write latency, Spark에서는 spill/shuffle fetch failure로 보일 수 있습니다.
> 증상 이름은 다르지만 lower-layer queue는 block device에 있을 수 있습니다.

1. 관측된 증상

    한 worker/broker/db node의 disk await가 급증합니다.

2. 가능한 원인 후보

    Kafka catch-up read, Cassandra compaction backlog, Spark spill, noisy neighbor, filesystem/device issue.

3. OS/kernel 관점에서 볼 지점

    `iostat -xz`, `vmstat`, disk queue, page cache pressure, process별 I/O.

4. distributed-system 관점에서 볼 지점

    replica lag, quorum wait, task retry, backpressure propagation.

5. 시스템 내부 구조와 연결되는 지점

    Kafka segment read/write, Cassandra SSTable compaction, Spark shuffle/spill local files.

6. 확인 명령 또는 로그

    `iostat -xz 1`, Kafka topic/group describe, `nodetool compactionstats`, Spark UI spill metrics.

7. 잘못된 결론의 예

    "각 제품에서 서로 다른 장애가 동시에 났다."

8. 더 나은 추론 과정

    시간대를 맞춰 OS disk saturation이 먼저인지 제품 내부 queue 증가가 먼저인지 봅니다. 하나의 disk 병목이 여러 distributed symptom으로 번졌을 수 있습니다.

## 현실 시나리오 2: retry가 장애를 회복하지 않고 증폭한다

> retry는 partial failure에 필요한 도구지만, backpressure와 idempotency 없이 쓰면 장애를 키웁니다.
> 세 시스템 모두 retry가 queue를 늘리고 중복 side effect를 만들 수 있는 경계가 있습니다.

1. 관측된 증상

    timeout 후 client retry가 늘고 cluster latency가 더 나빠집니다.

2. 가능한 원인 후보

    overloaded broker/coordinator/executor, downstream sink delay, network loss, too aggressive timeout, non-idempotent operation.

3. OS/kernel 관점에서 볼 지점

    socket queue, CPU run queue, disk await, GC pause, thread pool saturation.

4. distributed-system 관점에서 볼 지점

    retry storm, duplicate processing, idempotency key, quorum wait, task retry policy.

5. 시스템 내부 구조와 연결되는 지점

    Kafka producer retry/idempotence, Cassandra write retry and LWW/timestamp, Spark task retry and external side effects.

6. 확인 명령 또는 로그

    client retry logs, broker/coordinator request rate, Spark task attempt count, DB write duplicate keys.

7. 잘못된 결론의 예

    "timeout이 많으니 retry 횟수를 더 늘리자."

8. 더 나은 추론 과정

    retry가 성공률을 높이는지 queue를 키우는지 봅니다. idempotent boundary를 먼저 확인하고, exponential backoff, jitter, circuit breaker, rate limiting, downstream capacity를 함께 조정합니다.

## 근거와 더 읽을 자료

- [01_os_kernel_foundations.md](01_os_kernel_foundations.md)
- [02_distributed_system_foundations.md](02_distributed_system_foundations.md)
- [03_kafka_deep_dive.md](03_kafka_deep_dive.md)
- [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md)
- [05_spark_deep_dive.md](05_spark_deep_dive.md)
- [10_source_ledger.md](10_source_ledger.md)

