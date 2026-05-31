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

> blocking/non-blocking은 호출한 thread가 준비될 때까지 잠드는지의 축이고, synchronous/asynchronous는 작업 완료를 호출 흐름에서 직접 받는지 나중에 별도 경로로 받는지의 축입니다. epoll은 보통 완료가 아니라 readiness, 즉 "읽거나 쓸 수 있을 가능성"을 알려 줍니다.

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

> Cassandra는 random in-place update를 피하고 commit log append와 memtable update로 write를 먼저 받아들입니다. 나중에 memtable을 immutable SSTable로 flush하고 compaction으로 정리합니다.

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
