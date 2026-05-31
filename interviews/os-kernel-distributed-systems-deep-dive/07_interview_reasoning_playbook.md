# 07. Interview Reasoning Playbook

> 좋은 면접 답변은 길게 외운 문장이 아니라, 짧은 직답 뒤에 lower layer로 내려갈 수 있는 구조입니다.
> 장애·성능·일관성·확장성 질문은 먼저 상태, 순서, 병목, 실패 모델을 분리하면 제품명이 바뀌어도 같은 방식으로 풀 수 있습니다.
> 답변의 목표는 "정답 키워드"를 맞히는 것이 아니라, 꼬리 질문이 와도 어디서부터 다시 trace할지 잃지 않는 것입니다.

## 1. 공통 Reasoning Template

> 어떤 질문이 오든 먼저 "어떤 상태가 어디에 있고, 누가 그 상태를 바꾸며, 실패하면 어디서 다시 시작하는가"를 묻습니다.
> 이 네 가지를 잡으면 Kafka, Cassandra, Spark의 용어가 달라도 같은 원리로 답할 수 있습니다.

### 질문

모르는 시스템 질문을 받았을 때 어떻게 당황하지 않고 구조화할 수 있을까요?

### 직관

처음 가는 건물에서도 비상구, 전기실, 엘리베이터, 안내 데스크를 찾으면 길을 잡을 수 있습니다. 시스템 질문에서도 상태, 순서, 병목, 회복 지점을 찾으면 됩니다.

### 작은 예시

질문: "Kafka consumer lag가 늘면 어떻게 보나요?"

첫 구조:

```
state: broker log end offset, consumer committed offset
order: partition offset
bottleneck: consumer processing? broker fetch? downstream?
recovery: offset commit 이후 어디서 재개?
```

### 상태 이동 trace

```
question
  |
  +-- state: what data/task/log/checkpoint exists?
  +-- ownership: who can mutate it?
  +-- ordering: how is next state decided?
  +-- failure: what if actor pauses/crashes?
  +-- resource: CPU/memory/disk/network/lock?
  +-- verification: which metric/log/command proves it?
```

### 내부 메커니즘

이 template은 제품 기능을 나열하지 않게 막습니다. Kafka 질문이면 offset/log/replica로, Cassandra 질문이면 RF/CL/SSTable/repair로, Spark 질문이면 DAG/stage/task/shuffle/checkpoint로 내려갑니다. 그 다음 OS 자원과 연결합니다.

### 실패 모드

- 키워드만 말하면 꼬리 질문에서 collapse됩니다.
- product docs의 설정명만 말하면 원리 질문에서 약합니다.
- OS 지표만 말하면 제품 내부 상태를 놓칩니다.

### 검증 방법

답변 후 스스로 물어봅니다.

```text
1. 내가 말한 상태는 실제로 어디에 저장되는가?
2. 그 상태를 바꾸는 주체는 누구인가?
3. 실패하면 같은 작업이 다시 실행되는가, 건너뛰는가, 중복되는가?
4. 무엇을 보면 내 설명이 맞는지 확인할 수 있는가?
```

PASS 신호:

- 답변에 상태 저장 위치와 검증 신호가 있습니다.

FAIL 신호:

- "Kafka는 빠릅니다", "Cassandra는 AP입니다", "Spark는 in-memory입니다" 같은 라벨로 끝납니다.

### 면접식 되묻기

면접관이 "그럼 실제 장애 때 어떤 명령을 보나요?"라고 물으면 이론에서 관측으로 내려갑니다. Kafka라면 consumer group describe와 broker I/O, Cassandra라면 tracing/nodetool/iostat, Spark라면 Spark UI stage metrics와 executor GC/spill을 말합니다.

### 흔한 오해와 반례

오해: "짧게 답하라면 깊게 알 필요가 없다."

반례: 짧은 답변은 깊은 trace를 압축한 결과여야 합니다. 압축 원본이 없으면 꼬리 질문에서 무너집니다.

### Active recall

- 임의의 시스템 질문 하나를 골라 state, ownership, ordering, failure, resource, verification으로 나눠 보세요.
- "빠르다"라는 답변을 병목 trace로 바꿔 보세요.

## 2. 장애 질문 Template

> 장애 질문은 "무엇이 죽었나"보다 "어떤 상태 전이가 멈췄고, 그 관측이 crash/slow/partition 중 무엇과 양립 가능한가"로 시작합니다.
> timeout은 결론이 아니라 단서입니다.

짧은 답변:

```text
먼저 단일 timeout을 node crash로 단정하지 않고 partial failure로 봅니다.
client, network, remote process, disk/GC, distributed membership view를 나눠서 관측합니다.
그 다음 retry가 안전한지 idempotency와 checkpoint/offset 경계를 확인합니다.
```

깊은 꼬리 답변:

```
symptom: timeout
  |
  +-- caller local problem? run queue, GC, socket?
  +-- network? retransmit, packet loss, partition?
  +-- callee process? thread blocked, GC, deadlock?
  +-- callee storage? disk await, commit log, compaction, spill?
  +-- distributed state? ISR, quorum, executor lost, membership?
```

나쁜 답변:

> 서버가 죽은 것 같으면 재시작하고, 안 되면 replica를 늘립니다.

좋은 답변:

> timeout은 crash와 slow를 구분하지 못하므로 먼저 관측을 나눕니다. 예를 들어 Cassandra write timeout이면 coordinator, replica latency, CL, GC, commit log disk, network를 확인합니다. retry는 idempotency와 timestamp 경계를 확인한 뒤 조심해서 적용합니다.

Active recall:

- "응답 없음"과 "죽음"의 차이를 설명해 보세요.
- retry를 늘리기 전에 확인해야 할 것은 무엇인가요?

## 3. 성능 질문 Template

> 성능 질문은 throughput 숫자보다 queue가 어디서 자라는지 보는 문제입니다.
> CPU, memory, disk, network, lock, GC, partition skew, downstream sink 중 하나가 처리률을 제한하면 상위 시스템에는 lag, timeout, spill, pending queue로 나타납니다.

짧은 답변:

```text
먼저 병목이 어느 queue인지 봅니다.
입력률과 처리률, partition별 분포, OS 자원, runtime GC, downstream latency를 나누고,
해당 병목을 실제로 줄이는 scaling이나 batching, partitioning, data model 변경을 선택합니다.
```

깊은 꼬리 답변:

| 증상 | 먼저 보는 것 | 다음 질문 |
|---|---|---|
| Kafka lag | partition별 lag, processing rate, downstream | consumer를 늘려도 partition이 충분한가? |
| Cassandra latency | read/write latency, compactionstats, iostat | SSTable/compaction/disk가 병목인가? |
| Spark slow stage | task duration distribution, shuffle, spill | skew인가, memory인가, network인가? |

나쁜 답변:

> 서버를 늘리거나 memory를 키웁니다.

좋은 답변:

> 먼저 scale-out이 병목 queue를 나눌 수 있는지 확인합니다. Spark에서 slow task 하나가 skew로 900MB를 처리한다면 executor를 늘려도 그 task는 빨라지지 않을 수 있습니다. partitioning이나 salting 같은 data distribution 처방이 필요합니다.

Active recall:

- 낮은 CPU와 높은 latency가 함께 나타나는 원인 세 가지를 말해 보세요.
- Kafka/Cassandra/Spark에서 "queue"가 각각 어떤 metric으로 보이나요?

## 4. 일관성 질문 Template

> 일관성 질문은 "강하다/약하다"가 아니라 어떤 read가 어떤 write를 보아야 하는지로 구체화해야 합니다.
> quorum, consensus, transaction, idempotency, checkpoint는 모두 상태 순서를 제어하지만 보장 범위가 다릅니다.

짧은 답변:

```text
먼저 필요한 consistency model을 정합니다.
read-your-writes인지, linearizable read인지, eventual convergence로 충분한지 구분하고,
그 다음 시스템의 partition/replica/offset/checkpoint 경계로 내려갑니다.
```

깊은 꼬리 답변:

```
write accepted?
  |
  +-- accepted by one node only?
  +-- accepted by quorum?
  +-- committed in replicated log?
  +-- visible to read isolation level?
  +-- external side effect included?
```

나쁜 답변:

> Cassandra는 eventual consistency라 최신값을 못 믿습니다.

좋은 답변:

> Cassandra는 RF와 read/write CL로 tradeoff를 조절합니다. `R + W > RF`이면 intersection이 생기지만, timestamp reconciliation, failed write, repair state, LWT 여부를 함께 봐야 하므로 "QUORUM이면 항상 linearizable"이라고 하면 틀립니다.

Active recall:

- Kafka transaction이 외부 DB write까지 자동으로 묶지 못하는 이유는 무엇인가요?
- CAP과 PACELC를 한 문장씩 구분해 보세요.

## 5. 확장성 질문 Template

> 확장성은 노드 수가 아니라 work가 얼마나 균등하고 독립적으로 나뉘는지의 문제입니다.
> partition이 병렬성의 단위인 동시에 ordering, transaction, skew, recovery의 경계라는 점을 놓치면 잘못된 scale-out을 선택합니다.

짧은 답변:

```text
먼저 병목 작업이 partition 가능한지 봅니다.
partition 수, key distribution, coordination 비용, failure domain, OS 자원 병목을 확인한 뒤 scale-out이 실제 처리률을 늘릴지 판단합니다.
```

깊은 꼬리 답변:

| 시스템 | scale-out 질문 |
|---|---|
| Kafka | partition 수와 consumer group parallelism이 충분한가? leader가 균형 있게 분산됐나? |
| Cassandra | partition key가 균등한가? RF와 rack/datacenter placement가 맞나? compaction/repair 비용은 감당 가능한가? |
| Spark | task 수와 partition size가 core에 맞나? skew와 shuffle이 scale-out을 막나? |

나쁜 답변:

> Kafka는 broker를 늘리면 됩니다.

좋은 답변:

> broker를 늘려도 topic partition leadership이 이동하지 않거나 consumer group parallelism이 partition 수에 막히면 처리량은 안 늘 수 있습니다. partition, leader distribution, producer/consumer 병목을 함께 봐야 합니다.

Active recall:

- hot key는 왜 노드 수 증가로 해결되지 않을 수 있나요?
- Spark executor를 늘려도 stage가 빨라지지 않는 경우를 설명해 보세요.

## 6. 자주 나오는 질문과 답변 골격

> 아래 답변은 외울 문장이 아니라 trace를 여는 첫 문장입니다.
> 면접에서는 먼저 20~30초로 답하고, 꼬리 질문이 오면 바로 아래층으로 내려갑니다.

### Kafka는 왜 빠른가요?

짧은 답변:

> Kafka는 record를 partition별 append-only log에 순차적으로 쓰고, producer batching과 compression으로 작은 요청을 묶으며, broker는 OS page cache와 가능한 경우 sendfile 같은 경로를 활용해 copy와 disk seek를 줄입니다. 다만 cache miss, TLS, compaction, follower lag, slow consumer가 있으면 disk와 network 병목이 드러납니다.

꼬리 질문 trace:

```
producer batch -> broker append -> page cache -> replica fetch -> consumer fetch
```

### Cassandra는 왜 write가 빠른가요?

짧은 답변:

> Cassandra는 기존 row를 제자리에서 수정하지 않고 commit log에 append하고 memtable에 반영한 뒤 immutable SSTable로 flush합니다. 이 LSM 구조는 random write를 줄이지만, 나중에 read amplification과 compaction pressure를 비용으로 냅니다.

꼬리 질문 trace:

```
mutation -> commit log -> memtable -> SSTable -> compaction
```

### Spark shuffle은 왜 비싼가요?

짧은 답변:

> shuffle은 partition 간 data를 재분배하므로 network 전송뿐 아니라 map output file, serialization, reduce fetch, sort/merge, memory pressure, disk spill을 함께 발생시킵니다. skew가 있으면 가장 큰 task 하나가 stage 전체를 지연시킬 수 있습니다.

꼬리 질문 trace:

```
wide dependency -> map-side shuffle write -> reduce fetch -> spill/merge
```

### Kafka의 exactly-once는 어떤 의미인가요?

짧은 답변:

> Kafka는 idempotent producer와 transaction, consumer offset을 transaction에 포함하는 방식, `read_committed` 같은 조건이 맞을 때 Kafka 내부 read-process-write 경로에서 exactly-once processing을 구성할 수 있습니다. 외부 DB나 API side effect는 별도 idempotency나 transaction 설계가 필요합니다.

꼬리 질문 trace:

```
read offset -> produce output -> send offsets to transaction -> commit
```

### quorum과 consensus는 무엇이 다른가요?

짧은 답변:

> quorum은 여러 replica 중 몇 개 응답을 성공으로 볼지 정하고 read/write 집합을 겹치게 만드는 기법입니다. consensus는 실패 상황에서도 여러 노드가 같은 값이나 log 순서에 안전하게 동의하도록 하는 더 큰 프로토콜입니다. quorum은 consensus 안에서 쓰일 수 있지만 둘은 동의어가 아닙니다.

꼬리 질문 trace:

```
quorum: response set size/intersection
consensus: leader/election/log matching/commit safety
```

## 현실 시나리오 1: "Kafka lag가 늘었는데 어떻게 설명할래요?"

> 좋은 답변은 lag 정의에서 시작해 partition별 원인과 OS/downstream 병목으로 내려갑니다.

1. 관측된 증상

    consumer group lag가 증가합니다.

2. 가능한 원인 후보

    consumer processing slow, downstream DB slow, hot partition, rebalance, broker fetch latency, network issue.

3. OS/kernel 관점에서 볼 지점

    consumer CPU/GC, socket queue, broker disk/page cache, network retransmission.

4. distributed-system 관점에서 볼 지점

    input rate vs processing rate, partition ownership, offset commit, retry/idempotency.

5. Kafka 내부 구조와 연결되는 지점

    log end offset, committed offset, partition assignment, fetch path.

6. 확인 명령 또는 로그

    `kafka-consumer-groups.sh --describe`, consumer logs, `vmstat`, `iostat`, `jstack`.

7. 잘못된 결론의 예

    "consumer를 늘리면 됩니다."

8. 더 나은 추론 과정

    partition별 lag와 processing rate를 먼저 봅니다. partition 수와 downstream bottleneck 때문에 consumer 추가가 효과 없을 수 있음을 말합니다.

## 현실 시나리오 2: "Spark job이 느린데 어디를 볼래요?"

> 좋은 답변은 stage와 task 분포에서 시작해 shuffle, skew, memory, spill, OS disk/network로 내려갑니다.

1. 관측된 증상

    job 전체 시간이 길고 특정 stage만 매우 느립니다.

2. 가능한 원인 후보

    wide dependency shuffle, skewed key, too few partitions, executor memory pressure, slow disk, GC, data locality.

3. OS/kernel 관점에서 볼 지점

    executor host disk await, network, memory pressure, GC, CPU run queue.

4. distributed-system 관점에서 볼 지점

    slow task tail, retry, executor lost, shuffle fetch failure.

5. Spark 내부 구조와 연결되는 지점

    DAG, stage boundary, task metrics, spill, shuffle read/write.

6. 확인 명령 또는 로그

    Spark UI Stage tab, Executors tab, executor logs, `iostat`, `jcmd`.

7. 잘못된 결론의 예

    "executor memory만 늘리면 됩니다."

8. 더 나은 추론 과정

    max task duration과 median을 비교해 skew를 확인합니다. spill/GC/disk/network 중 어느 지표와 함께 움직이는지 본 뒤 처방을 고릅니다.

## Active Recall Drill

문서를 덮고 아래 질문에 30초 답변과 2분 꼬리 답변을 각각 해 보세요.

1. Kafka partition을 늘리면 어떤 이점과 위험이 있나요?
2. Cassandra에서 `LOCAL_QUORUM` read/write를 쓰면 어떤 보장을 기대할 수 있고, 무엇은 기대하면 안 되나요?
3. Spark lineage와 checkpoint의 차이는 무엇인가요?
4. page cache와 fsync 차이가 Kafka/Cassandra 운영에서 왜 중요한가요?
5. backpressure를 제품 metric과 OS metric으로 동시에 설명해 보세요.

## 근거와 더 읽을 자료

- [01_os_kernel_foundations.md](01_os_kernel_foundations.md)
- [02_distributed_system_foundations.md](02_distributed_system_foundations.md)
- [03_kafka_deep_dive.md](03_kafka_deep_dive.md)
- [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md)
- [05_spark_deep_dive.md](05_spark_deep_dive.md)
- [10_source_ledger.md](10_source_ledger.md)

