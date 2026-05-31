# 07. Interview Reasoning Playbook

면접에서 좋은 답변은 길이가 긴 답변이 아닙니다. 첫 20~30초에는 질문의 핵심을 짧게 잡고, 꼬리 질문이 오면 어디서부터 다시 내려갈지 잃지 않는 답변입니다. 이 문서는 Kafka, Cassandra, Spark 이름을 외우는 대신 다음 여섯 칸으로 질문을 다시 쓰는 연습을 합니다.

```
state
  -> owner
  -> order
  -> copy or recomputation
  -> failure point
  -> verification
```

## 1. 모르는 질문을 받으면 상태부터 찾는다

질문이 "Kafka가 왜 빠른가요?"든 "Spark job이 느리면 어디를 보나요?"든 먼저 상태를 찾습니다. Kafka라면 record와 offset, Cassandra라면 mutation과 replica cell, Spark라면 partition과 task입니다. 상태를 찾으면 다음 질문은 자연스럽게 이어집니다. 누가 그 상태를 소유하는가? 어떤 순서가 보장되는가? 복제되거나 다시 계산되는가? 실패하면 어떤 상태까지 남는가?

예를 들어 "consumer lag가 늘면 어떻게 보나요?"라는 질문에 곧바로 "consumer를 늘립니다"라고 답하면 위험합니다. lag는 원인이 아니라 log end offset과 committed offset의 차이입니다. 원인은 consumer CPU일 수도 있고, downstream DB일 수도 있고, broker fetch latency일 수도 있고, partition skew일 수도 있습니다.

짧은 답은 이렇게 시작할 수 있습니다.

> Kafka lag는 consumer group이 partition log를 따라가지 못하는 상태입니다. 먼저 어느 partition의 end offset과 committed offset 차이가 커지는지 보고, 그 다음 consumer 처리 시간, downstream latency, broker fetch latency, partition skew를 나눠 봅니다.

꼬리 질문이 오면 다음 trace로 내려갑니다.

```
producer append rate
  -> broker partition log end offset
  -> consumer fetch
  -> processing / downstream write
  -> offset commit
  -> lag
```

## 2. 장애 질문은 "무엇이 죽었나"보다 "무엇을 관측했나"로 시작한다

분산 시스템에서 timeout은 실패의 증명이 아닙니다. 장애 질문을 받으면 먼저 관측을 말해야 합니다. 응답이 없었다, latency가 늘었다, replica lag가 생겼다, leader election이 일어났다, task가 retry됐다처럼 관측 가능한 신호를 분리합니다.

좋은 답변은 관측에서 가능한 원인을 나눕니다.

| 관측 | 가능한 원인 | 다음 확인 |
|---|---|---|
| Kafka produce timeout | broker disk, network, ISR shrink, producer buffer full | broker request metrics, ISR, network, disk latency |
| Cassandra read timeout | slow replica, tombstone, compaction, network, CL too high | tracing, replica latency, pending compaction, tombstone warnings |
| Spark task retry | executor lost, fetch failure, OOM, skew, transient network | Spark UI, executor logs, shuffle fetch metrics, GC |

짧은 답변은 "어떤 상태 전이가 멈췄고, crash/slow/partition 중 어떤 관측과 양립 가능한지부터 나눕니다"입니다.

## 3. 성능 질문은 queue와 병목을 찾는 문제다

throughput과 latency는 평균 숫자로만 보면 속기 쉽습니다. 성능 질문에서는 producer와 consumer가 누구인지, 어느 queue가 자라는지 찾아야 합니다.

Kafka에서 queue는 producer buffer, broker request queue, replica lag, consumer lag로 나타납니다. Cassandra에서는 memtable flush queue, compaction backlog, request stage pending, disk queue로 나타납니다. Spark에서는 scheduler pending task, shuffle fetch wait, spill file, skewed task로 나타납니다.

답변을 만들 때는 "CPU, memory, disk, network 중 무엇이 병목인가요?"로 바로 뛰지 말고, 먼저 요청 경로를 그립니다.

```
input rate
  -> in-memory queue
  -> CPU work
  -> disk or network
  -> downstream queue
  -> ack or commit point
```

그 다음 각 칸에 metric을 붙입니다. metric 없이 "느린 것 같습니다"라고 말하면 추론이 아니라 감상입니다.

## 4. 일관성 질문은 read/write 계약으로 번역한다

"Cassandra는 consistent한가요?", "Kafka exactly-once는 완벽한가요?" 같은 질문은 범위를 좁혀야 답할 수 있습니다. 먼저 어떤 write가 어떤 read에 보여야 하는지 묻습니다.

Cassandra 답변 예시는 이렇습니다.

> Cassandra는 RF와 read/write consistency level로 관측 가능한 최신성의 확률과 비용을 조절합니다. RF=3에서 write QUORUM, read QUORUM이면 read/write replica 집합이 교차하지만, clock skew, failed write, repair 상태, LWT 여부 때문에 이를 blanket linearizability로 말하면 안 됩니다.

Kafka 답변 예시는 이렇습니다.

> Kafka의 exactly-once는 지원된 Kafka 내부 read-process-write 경로에서 producer idempotence, transaction, offset commit을 묶어 중복 처리를 줄이는 의미입니다. 외부 DB나 HTTP side effect까지 자동으로 exactly-once가 되는 것은 아니므로 idempotent sink나 transaction boundary가 필요합니다.

## 5. 확장성 질문은 partitioning과 coordination 비용으로 답한다

"node를 늘리면 해결되나요?"라는 질문은 항상 partitioning과 coordination을 같이 봐야 합니다. Kafka partition이 충분하지 않으면 consumer를 늘려도 병렬성이 늘지 않습니다. Cassandra는 node를 늘려도 hot partition이 있으면 특정 replica 집합이 계속 병목입니다. Spark는 executor를 늘려도 skewed partition 하나가 느리면 stage completion이 늦습니다.

scale-out 답변의 기본 흐름은 다음과 같습니다.

```
what is partitioned?
  -> is load evenly distributed?
  -> does adding workers increase parallel consumers/tasks?
  -> what coordination or shuffle/repair/rebalance cost appears?
  -> what metric proves improvement?
```

## 자주 나오는 질문에 대한 압축 답변과 꼬리 경로

### blocking, non-blocking, synchronous, asynchronous는 어떻게 다르나요?

짧은 답변:

> blocking/non-blocking은 호출한 thread가 준비될 때까지 잠드는지의 축이고, synchronous/asynchronous는 작업 완료를 호출 흐름에서 직접 받는지 별도 경로로 받는지의 축입니다. epoll은 보통 완료가 아니라 readiness, 즉 "읽거나 쓸 수 있을 가능성"을 알려 줍니다.

꼬리 경로:

```
blocking read
  -> no data
  -> thread sleeps in kernel wait queue
  -> packet arrives
  -> thread runnable

non-blocking + epoll
  -> fd registered
  -> epoll_wait returns readable
  -> application read() drains until EAGAIN
```

조심할 말: "`epoll`은 async I/O입니다"라고 말하면 범위가 흐립니다. epoll은 readiness multiplexing이고, application이 실제 read/write와 partial 처리 책임을 가집니다.

### 패킷 하나는 어떻게 request가 되나요?

짧은 답변:

> NIC가 받은 packet은 DMA ring, interrupt/NAPI, kernel TCP/IP stack, socket receive queue를 거친 뒤 epoll/read를 통해 application buffer로 올라오고, 그 byte를 application protocol parser가 request로 해석합니다.

꼬리 경로:

```
NIC RX ring
  -> NAPI poll
  -> TCP reassembly
  -> socket receive queue
  -> epoll readiness
  -> application read buffer
  -> protocol frame parser
  -> request object
```

조심할 말: packet 도착과 handler 실행을 같은 사건으로 말하면 안 됩니다. socket이 readable이어도 thread scheduling, event loop, parser, downstream I/O가 남아 있습니다.

### Kafka는 왜 빠른가요?

짧은 답변:

> Kafka는 partition log에 순차 append하고, producer/broker/consumer가 batch로 일하며, OS page cache와 file-to-socket 전송 경로를 활용해 작은 I/O와 불필요한 copy를 줄이기 때문에 빠릅니다.

꼬리 경로:

```
producer batch
  -> leader partition append
  -> page cache / log segment
  -> replica fetch / high watermark
  -> consumer fetch
  -> offset commit
```

조심할 말: "디스크를 안 써서 빠르다"는 틀립니다. Kafka는 디스크를 쓰되 append와 sequential read가 잘 맞게 설계합니다.

### Cassandra는 왜 write가 빠른가요?

짧은 답변:

> Cassandra는 random in-place update를 피하고 commit log append와 memtable update로 write를 먼저 받아들입니다. 이후 memtable을 immutable SSTable로 flush하고 compaction으로 정리합니다.

꼬리 경로:

```
mutation
  -> coordinator
  -> replicas
  -> commit log + memtable
  -> CL ack
  -> SSTable flush
  -> compaction / repair
```

조심할 말: write가 빠른 대가로 read amplification, tombstone, compaction, repair 비용이 생깁니다.

### Spark shuffle은 왜 비싼가요?

짧은 답변:

> Spark shuffle은 data를 key나 partitioner 기준으로 다시 배치하므로 serialization, local disk spill, network fetch, memory pressure, skew가 함께 생깁니다.

꼬리 경로:

```
map task output
  -> partition by reducer
  -> spill/write shuffle blocks
  -> reduce task fetches blocks
  -> deserialize / merge / aggregate
```

조심할 말: "network가 비싸서"만 말하면 부족합니다. shuffle은 disk와 memory, GC 문제이기도 합니다.

### quorum과 consensus는 무엇이 다른가요?

짧은 답변:

> quorum은 몇 replica 응답을 성공으로 볼지 정하는 응답 조건이고, consensus는 여러 node가 같은 값이나 log 순서에 안전하게 합의하는 protocol입니다.

꼬리 경로:

```
Cassandra QUORUM:
  read/write response intersection

Raft-like consensus:
  leader term
  log index
  majority replication
  commit rule
```

조심할 말: Cassandra 일반 QUORUM을 consensus라고 부르면 안 됩니다.

### CPU는 낮은데 latency가 높으면 어디를 보나요?

짧은 답변:

> CPU 사용률이 낮아도 요청 thread가 socket I/O, disk I/O, lock/futex, GC, scheduler wait, cgroup throttle에서 기다릴 수 있습니다. 먼저 요청 경로의 어느 queue나 wait state가 커졌는지 봅니다.

꼬리 경로:

```
request arrives
  -> socket queue
  -> runnable thread
  -> lock or futex wait
  -> disk/network syscall
  -> downstream response
```

조심할 말: "CPU가 낮으니 서버는 여유 있습니다"는 위험합니다. off-CPU time, run queue, thread dump, `ss`, `iostat`, GC log를 함께 봐야 합니다.

## 나쁜 답변을 고치는 방법

| 나쁜 답변 | 왜 약한가 | 고친 답변 방향 |
|---|---|---|
| Kafka는 메모리라 빠릅니다. | disk/page cache/log 구조를 놓침 | page cache, append-only log, batch, sendfile로 설명 |
| Cassandra는 eventual이라 최신을 못 봅니다. | CL/LWT/repair 경계를 지움 | RF/CL/read-write trace와 caveat로 설명 |
| Spark는 executor 늘리면 됩니다. | partition/skew/shuffle 병목을 무시 | stage/task/shuffle metric으로 병목 분리 |
| timeout이면 실패입니다. | 적용 여부 불명확성을 무시 | timeout은 관측이며 idempotency/retry 설계 필요 |
| CAP 때문에 선택해야 합니다. | CAP를 평상시 분류표로 오용 | partition 시 consistency/availability tradeoff와 PACELC로 좁힘 |

## 연습 문제

각 질문에 30초 답변과 2분 꼬리 답변을 따로 만들어 보세요.

1. Kafka consumer lag가 특정 partition에서만 늘면 어떻게 보나요?
2. Cassandra에서 QUORUM read인데 오래된 값이 보였다는 제보를 받으면 무엇을 확인하나요?
3. Spark job에서 한 stage의 task 하나만 오래 걸리면 어떤 원인을 의심하나요?
4. `write()`가 성공했는데 crash 후 데이터가 사라질 수 있다는 말을 Kafka/Cassandra/Spark에 각각 연결해 보세요.
5. retry를 넣었더니 장애가 더 커진 사례를 세 시스템 중 하나로 설명해 보세요.

## 면접 전 마지막 점검

좋은 답변은 항상 다음 네 문장을 포함할 수 있습니다.

- "먼저 상태와 소유자를 나누겠습니다."
- "이 지점의 순서 보장은 여기까지입니다."
- "실패하면 이 상태가 불명확해지고, 그래서 이 검증이 필요합니다."
- "이 설정은 성능을 높일 수 있지만, 대신 이 consistency/durability/recovery 비용을 냅니다."

## 답변을 시작하는 네 가지 안전한 문장

면접에서 깊은 주제를 받으면 바로 모든 지식을 쏟아내기보다, 먼저 범위를 좁히는 문장을 두는 편이 좋습니다. 첫째, "어느 계층의 성공인지 나누겠습니다." 둘째, "이 보장은 어느 범위에서만 성립하는지 보겠습니다." 셋째, "timeout은 실패 증명이 아니라 불확실한 관측으로 보겠습니다." 넷째, "성능 문제는 on-CPU와 off-CPU, queue와 backpressure로 나누겠습니다." 이 네 문장은 OS와 분산 시스템, 제품 질문에 모두 쓸 수 있는 시작점입니다.

예를 들어 `write()` 질문을 받으면 이렇게 답을 엽니다. "`write()` 성공은 일반적으로 kernel/file path가 bytes를 받아들였다는 뜻이지, 항상 durable storage를 뜻하지 않습니다. 더 강한 내구성이 필요하면 `fsync()`와 filesystem, storage, application-level replication boundary를 봐야 합니다." 이 문장은 짧지만 page cache, dirty writeback, product durability로 내려갈 길을 열어 둡니다.

Kafka 질문에는 "partition 안 offset order와 topic 전체 order를 나누겠습니다"라고 시작합니다. Cassandra 질문에는 "RF와 CL, 그리고 LWT 같은 더 강한 path를 분리하겠습니다"라고 시작합니다. Spark 질문에는 "lineage로 재계산 가능한 순수 계산과 외부 sink side effect를 분리하겠습니다"라고 시작합니다. 좋은 답변은 첫 문장부터 경계를 만듭니다.

## 30초 답변과 2분 답변을 분리하는 법

30초 답변은 정의보다 구조가 중요합니다. 질문자가 듣고 싶은 것은 모든 세부가 아니라 "이 사람이 올바른 축으로 생각하는가"입니다. 2분 답변은 trace와 반례가 중요합니다. 예를 들어 Kafka lag 질문의 30초 답변은 "lag는 consumer group이 log end를 따라가지 못한다는 결과 지표이고, 원인은 broker fetch, consumer processing, downstream sink, partition skew, rebalance로 나눠 봅니다"입니다.

2분 답변은 이렇게 확장합니다.

```
lag grows
  -> check whether all partitions or one partition
  -> compare broker fetch latency and consumer processing time
  -> inspect downstream sink latency
  -> look for rebalance/heartbeat/max poll interval
  -> only then choose consumer count, partitioning, or backpressure change
```

Cassandra stale read 질문의 30초 답변은 "QUORUM은 read/write 응답 집합 교차를 만들지만 blanket 최신성 보장은 아니므로, RF/CL trace, timestamp conflict, failed write, repair/LWT 여부를 봅니다"입니다. 2분 답변은 replica A/B/C에 어떤 version이 있고 read가 어느 replica를 봤는지 그립니다. Spark straggler 질문의 30초 답변은 "stage 하나가 느리면 skew, slow executor, shuffle fetch, GC, disk, cgroup throttle을 나눠 봅니다"입니다. 2분 답변은 Spark UI metric과 OS metric을 연결합니다.

## 반례를 말할 수 있어야 깊은 답변이다

면접에서 강한 인상을 주는 답변은 "무엇이다"보다 "언제 그 말이 깨지는가"를 압니다. "Kafka는 순서를 보장한다"는 말은 partition 안에서는 맞지만 partition 사이에서는 깨집니다. "Cassandra QUORUM은 최신 값을 준다"는 말은 정상적인 교차 조건에서는 최신성 가능성을 높이지만 clock skew, conflict resolution, failed write, repair 상태에 따라 깨집니다. "Spark는 lineage로 복구한다"는 말은 deterministic computation에는 강하지만 외부 side effect는 자동으로 되돌리지 못합니다.

반례를 말할 때는 비관적으로 끝내지 말고 수리 방법을 붙입니다. Kafka partition 사이 order가 필요하면 key design을 조정하거나 single partition의 tradeoff를 받아들이거나 application-level ordering을 둡니다. Cassandra에서 stronger semantics가 필요하면 LWT나 transaction feature, data model 변경, read/write CL 조정, idempotency를 검토합니다. Spark external sink에는 idempotent key, batch id, commit protocol을 둡니다.

## 관측 질문으로 답변을 마무리하기

좋은 답변의 끝에는 관측 경로가 있습니다. "그럴 수 있습니다"에서 끝나면 추론입니다. "그래서 무엇을 보겠습니다"까지 가면 실무 답변입니다. CPU가 낮은데 latency가 높다면 run queue, off-CPU wait, thread dump, `ss`, `iostat`, GC log를 봅니다. Kafka produce latency가 높다면 request queue time, local time, remote time, disk/network/GC를 봅니다. Cassandra timeout이면 coordinator와 replica log, pending tasks, GC, compaction, network를 봅니다. Spark stage가 느리면 task distribution, shuffle read/write, spill, GC, executor loss를 봅니다.

```
claim
  -> possible mechanism
  -> counterexample
  -> observation that would distinguish causes
  -> safe action after evidence
```

이 구조를 외우면 모르는 질문에도 무너지지 않습니다. 바로 정답을 모르더라도 어떤 상태를 나눠야 하고, 어떤 증거가 있으면 판단이 바뀌는지 말할 수 있기 때문입니다.

## 제품별 꼬리 질문 대응

Kafka에서 "`acks=all`이면 안전한가요?"라고 물으면, "ISR 조건을 만족하는 replication ack를 기다린다는 뜻이지, 모든 외부 side effect나 모든 storage caveat가 닫힌다는 뜻은 아닙니다"라고 시작합니다. 이어서 `min.insync.replicas`, high watermark, leader election, flush policy, idempotent producer를 말합니다. "consumer lag가 있으면 consumer를 늘리면 되나요?"에는 partition 수, skew, downstream sink, rebalance, fetch latency를 먼저 봅니다.

Cassandra에서 "eventual consistency니까 최신성은 포기한 건가요?"라고 물으면, "ordinary read/write는 RF와 CL로 latency/availability/recency를 조절하고, 더 강한 조건에는 LWT나 별도 transaction path가 있습니다"라고 답합니다. 이어서 QUORUM trace와 stale 가능성을 말합니다. "Compaction을 줄이면 빨라지나요?"에는 foreground I/O는 줄 수 있지만 read amplification과 tombstone, space 문제가 커질 수 있다고 말합니다.

Spark에서 "executor를 늘리면 빨라지나요?"라고 물으면, "병렬화 가능한 partition이 충분하고 병목이 resource 부족일 때는 도움이 되지만, skew, single reducer, shuffle fetch, external sink, driver bottleneck이면 제한적입니다"라고 답합니다. "checkpoint를 자주 하면 안전한가요?"에는 recovery boundary는 좋아질 수 있지만 I/O 비용과 state size, sink semantics를 함께 봐야 한다고 말합니다.

## 마지막 연습: 하나의 현상을 세 층으로 답하기

현상: p99 latency가 올랐습니다. OS 답변은 thread가 CPU에서 실행 중인지, run queue에서 기다리는지, lock/futex, disk, network, GC, cgroup throttle에서 기다리는지 나누는 것입니다. Distributed 답변은 timeout, retry, queue growth, partial failure, replica lag, membership change를 보는 것입니다. Product 답변은 Kafka request/consumer lag, Cassandra pending/compaction/repair, Spark stage/shuffle/skew를 보는 것입니다.

```
p99 latency
  OS layer: CPU, memory, disk, network, lock, scheduler
  distributed layer: timeout, retry, quorum, membership, recovery
  product layer: log/replica, SSTable/repair, shuffle/checkpoint
```

이 세 층을 동시에 말할 수 있으면 면접 답변이 단어 암기에서 reasoning으로 바뀝니다. 답변이 길어져도 항상 현재 상태, 소유자, 순서 보장, 불확실성, 관측 경로를 잃지 않는 것이 이 playbook의 목적입니다.

## 꼬리 질문 1: `write()`와 durability

30초 답변은 이렇게 시작합니다. "`write()` 성공은 일반적으로 커널의 파일 쓰기 경로가 bytes를 받아들였다는 뜻이고, durable storage까지 내려갔다는 뜻은 아닐 수 있습니다. Buffered I/O에서는 page cache dirty page가 생기고, 더 강한 동기화가 필요하면 `fsync()`와 filesystem/storage caveat를 봐야 합니다." 이 답변은 OS 기초를 짧게 닫습니다.

2분 답변에서는 제품으로 연결합니다. Kafka produce ack는 broker replication 조건이고, Cassandra write success는 CL response와 commitlog/memtable path이며, Spark checkpoint success는 sink/storage commit semantics와 묶입니다. 세 시스템 모두 local `write()`와 product-level success가 다릅니다. 면접관이 "그러면 어떻게 확인하나요?"라고 물으면 `strace`로 syscall, `iostat`으로 block I/O, product metric으로 ack/flush/replication/checkpoint를 나눠 보겠다고 말합니다.

## 꼬리 질문 2: `epoll`과 async

30초 답변은 "epoll은 readiness notification입니다. fd가 읽거나 쓸 준비가 되었을 가능성을 알려 주지만, 실제 read/write의 완료나 전체 message boundary를 보장하지 않습니다"입니다. 이어서 blocking/non-blocking과 sync/async는 다른 축이라고 말합니다. Non-blocking fd에서 `read()`가 `EAGAIN`을 줄 수 있고, partial read/write를 application protocol이 처리해야 합니다.

2분 답변에서는 event loop를 말합니다. Event loop는 많은 connection을 적은 thread로 기다리게 해 주지만, loop 안에서 blocking call이나 긴 CPU 작업을 하면 그 loop가 담당하는 모든 connection이 늦어집니다. Kafka network thread, Cassandra native transport, Spark RPC/shuffle service 모두 이 문제를 갖습니다. 관측은 event loop CPU, socket queue, request queue, scheduler delay를 함께 봅니다.

## 꼬리 질문 3: QUORUM과 consensus

30초 답변은 "QUORUM은 replica 응답 수 정책이고 consensus는 값이나 log 순서에 합의하는 protocol입니다. 둘 다 majority를 쓸 수 있지만 해결하는 문제가 다릅니다"입니다. Cassandra RF=3, R=2, W=2 trace를 그려 read/write set intersection을 설명합니다. 하지만 이것이 blanket linearizability는 아니라고 덧붙입니다.

2분 답변에서는 Raft류 consensus와 비교합니다. Consensus는 leader, term, log index, commit rule로 같은 log prefix를 보장하려 합니다. Cassandra ordinary write/read는 availability와 tunable consistency를 위해 quorum 정책을 쓰고, LWT나 transaction feature는 별도 path입니다. Kafka partition replication과 metadata quorum도 서로 다른 계층입니다. 면접관이 "그럼 어떤 걸 쓰나요?"라고 물으면 operation requirement와 failure model에 따라 선택한다고 답합니다.

## 꼬리 질문 4: GC와 OS 메모리

30초 답변은 "JVM heap과 process 전체 memory는 다릅니다. Heap, off-heap, direct buffer, thread stack, metaspace, page cache, socket buffer, cgroup limit을 나눠 봐야 합니다"입니다. Heap이 남아도 container OOM이 날 수 있고, heap을 키우면 page cache가 줄어 disk I/O가 늘 수 있습니다.

2분 답변에서는 Kafka/Cassandra/Spark로 연결합니다. Kafka는 page cache가 throughput에 중요하므로 heap을 무작정 키우면 오히려 해로울 수 있습니다. Cassandra는 memtable/cache/off-heap/page cache/compaction이 함께 움직입니다. Spark executor는 heap, overhead, off-heap, Python worker, shuffle spill, page cache가 cgroup limit 안에서 경쟁합니다. 관측은 GC log, native memory tracking, `/proc`, cgroup memory events, OOM reason을 함께 봅니다.

## 꼬리 질문 5: retry와 idempotency

30초 답변은 "Retry는 실패를 없애는 기능이 아니라 불확실한 요청을 다시 보내는 정책입니다. Timeout 뒤 원래 요청이 적용되었을 수 있으므로 idempotency key, deduplication, transaction boundary, backoff가 필요합니다"입니다. 이 답변은 분산 시스템의 partial failure를 정확히 짚습니다.

2분 답변에서는 제품별 boundary를 말합니다. Kafka idempotent producer와 transaction은 Kafka 내부 특정 경로를 보호하지만 외부 DB side effect를 자동으로 감싸지 않습니다. Cassandra write retry는 primary key와 timestamp semantics에 따라 안전할 수도 위험할 수도 있습니다. Spark task retry는 순수 계산에는 좋지만 외부 sink write에는 중복을 만들 수 있습니다. 관측은 request id, sink record, offset/task attempt, retry count를 봅니다.

## 꼬리 질문 6: CAP와 PACELC

30초 답변은 "CAP는 network partition이 있을 때 consistency와 availability를 동시에 완벽히 만족할 수 없다는 불가능성 결과이지, 평상시 제품 분류표가 아닙니다"입니다. Partition이 없을 때도 latency와 consistency tradeoff가 있다는 점에서 PACELC가 더 넓은 사고를 줍니다.

2분 답변에서는 operation 단위로 말합니다. Cassandra는 CL 설정에 따라 read/write behavior가 달라지고, Kafka는 partition log와 replication/election policy가 보장 범위를 정하며, Spark는 storage system이 아니라 computation runtime이므로 CAP 표에 억지로 넣으면 오해가 큽니다. 좋은 답변은 제품 이름을 C/A/P 칸에 넣는 대신, 어떤 operation이 어떤 failure에서 어떤 관측 결과를 보이는지 말합니다.

## 답변을 망치는 습관과 수리법

첫 번째 나쁜 습관은 강한 단어를 범위 없이 쓰는 것입니다. "항상", "무조건", "정확히 한 번", "최신", "안전" 같은 단어는 범위가 없으면 위험합니다. 수리법은 "어느 범위에서"를 붙이는 것입니다. Partition 안에서, Kafka transaction이 지원하는 경로에서, RF/CL 조건과 repair 상태에서, deterministic Spark transformation 안에서처럼 범위를 적습니다.

두 번째 나쁜 습관은 단일 metric을 원인으로 읽는 것입니다. CPU 낮음, lag 증가, timeout, GC pause, disk busy는 모두 결과 지표일 수 있습니다. 수리법은 queue와 owner를 찾는 것입니다. 누가 생산하고 누가 소비하며, 어느 queue가 커지고, 그 queue의 downstream은 무엇인지 묻습니다.

세 번째 나쁜 습관은 해결책을 너무 빨리 말하는 것입니다. "thread를 늘리겠습니다", "executor를 늘리겠습니다", "CL을 올리겠습니다", "retry를 넣겠습니다"는 모두 tradeoff가 있습니다. 수리법은 조치 전 관측을 말하는 것입니다. 어떤 지표가 PASS면 이 조치를 하고, 어떤 지표가 나오면 다른 조치를 하겠다고 조건을 붙입니다.

## 마지막 replay: 긴 답변을 짧게 접는 법

긴 설명을 마친 뒤에는 항상 한 문장으로 접어야 합니다. OS 질문은 "자원은 process-local handle처럼 보이지만 kernel-global object와 queue를 지난다"로 접습니다. 분산 질문은 "timeout은 불확실성을 만들고 log/quorum/consensus/retry/recovery가 그 불확실성을 줄인다"로 접습니다. Kafka 질문은 "partition log와 offset, ISR/high watermark가 중심이다"로 접습니다. Cassandra 질문은 "partition key, replica, CL, commitlog/SSTable/repair가 중심이다"로 접습니다. Spark 질문은 "DAG/stage/task, shuffle, lineage/checkpoint가 중심이다"로 접습니다.

이 접는 문장이 있어야 답변이 길어져도 면접관이 핵심을 기억합니다. 그리고 본인이 길을 잃었을 때도 다시 돌아올 수 있습니다.

## 압박 질문을 받았을 때의 방어선

면접관이 "그래서 결국 뭐가 원인인가요?"라고 재촉하면, 근거 없이 하나를 찍지 않는 것이 중요합니다. 대신 "현재 증거만으로는 A와 B가 모두 가능합니다. A라면 이 지표가 보일 것이고, B라면 저 지표가 보일 것입니다"라고 답합니다. 이 태도는 회피가 아니라 실무적인 가설 분리입니다. 분산 시스템과 OS 문제는 같은 증상이 여러 원인과 양립하기 때문입니다.

예를 들어 Cassandra read timeout에서 disk await와 GC pause 정보가 없다면, replica가 느린 이유를 확정할 수 없습니다. Kafka produce latency에서 request queue time과 remote time을 모르면 broker local append가 느린지 follower replication이 느린지 모릅니다. Spark straggler에서 task input size와 executor host metric을 모르면 skew인지 noisy node인지 모릅니다. 좋은 답변은 "지금은 확정하지 않고, 다음 관측으로 가르겠습니다"라고 말할 수 있어야 합니다.

```
symptom
  -> possible cause A
  -> possible cause B
  -> discriminating evidence
  -> safe next action
```

이 구조는 신뢰를 만듭니다. 모르는 것을 아는 척하지 않고, 그렇다고 멈추지도 않기 때문입니다.

## 시스템 설계 질문으로 확장될 때

OS/분산 시스템 질문은 종종 설계 질문으로 이어집니다. "대량 주문 이벤트를 어떻게 처리하겠습니까?"라는 질문을 받으면 Kafka topic만 말하지 말고, partition key, idempotency key, consumer group, offset commit, downstream DB transaction, retry/backpressure, observability를 함께 말해야 합니다. Cassandra를 쓴다면 partition key와 CL, TTL/tombstone, compaction, repair를 함께 봅니다. Spark를 쓴다면 batch/streaming, checkpoint, sink idempotency, shuffle/skew를 함께 봅니다.

```
design answer skeleton
  data ownership
  ordering boundary
  durability boundary
  retry/idempotency
  backpressure
  recovery
  observability
```

이 skeleton은 외워서 쓰는 template이 아니라 생각 순서입니다. 먼저 data가 어디에 저장되고 누가 소유하는지 말합니다. 그 다음 어떤 순서 보장이 필요한지, 어느 지점까지 성공으로 볼지, timeout 뒤 retry가 안전한지, queue가 커질 때 어떻게 속도를 줄일지, 장애 뒤 어디서 재시작할지, 어떤 metric으로 판단할지 말합니다. 이 순서를 지키면 설계 답변도 낮은 층과 높은 층이 이어집니다.

## 마지막 자기 점검

답변을 끝내기 전 마음속으로 네 가지를 확인합니다. 첫째, 같은 단어를 다른 시스템에서 같은 뜻으로 써 버리지 않았는가. 둘째, 성공과 내구성, visibility와 recovery를 섞지 않았는가. 셋째, timeout을 실패 확정으로 말하지 않았는가. 넷째, 관측 없이 해결책을 먼저 제안하지 않았는가. 하나라도 걸리면 답변을 한 문장 고쳐야 합니다.

이 playbook의 목적은 완벽한 암기가 아닙니다. 낯선 질문을 받아도 상태, 소유자, 순서, 불확실성, 관측이라는 다섯 축으로 다시 세우는 것입니다. 그 다섯 축을 지키면 OS kernel, Kafka, Cassandra, Spark 어느 쪽으로 질문이 휘어도 답변의 중심을 잃지 않습니다.

마지막으로 답변은 짧게 시작하고 깊게 내려가야 합니다. 첫 문장은 경계를 세우고, 두 번째 문장은 trace를 열고, 세 번째 문장은 반례와 관측을 붙입니다. 이 세 박자만 지켜도 긴장한 자리에서 단어를 흩뿌리는 답변을 피할 수 있습니다.

이 기준이 면접장에서 답변을 다시 세우는 최소한의 안전장치입니다.
