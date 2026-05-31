# 03. Kafka Deep Dive

> Kafka는 메시지를 임시로 전달하는 큐라기보다, 여러 소비자가 각자 읽은 위치를 기억하며 따라갈 수 있는 분산 append-only log로 이해해야 합니다.
> partition은 처리량을 늘리는 단위이면서 동시에 순서 보장의 경계입니다.
> Kafka의 장애와 성능 문제는 대개 log, replica, offset, consumer lag, network/disk backpressure 중 어디에서 상태가 밀렸는지를 따라가며 해석합니다.

## 1. Append-only Log, Topic, Partition

> topic은 record가 모이는 이름이고, partition은 topic을 나눈 append-only log 조각입니다.
> Kafka의 순서 보장은 topic 전체가 아니라 한 partition 안에서 성립합니다.
> partition 수를 늘리면 병렬성은 늘 수 있지만, key ordering, consumer parallelism, rebalance, storage layout도 함께 바뀝니다.

### 질문

Kafka를 queue라고만 부르면 어떤 중요한 구조를 놓칠까요?

### 직관

queue는 보통 한 소비자가 가져가면 사라지는 줄로 생각하기 쉽습니다. Kafka partition은 도서관의 원본 장부에 가깝습니다. 여러 독자가 각자 책갈피(offset)를 들고 같은 장부를 다른 속도로 읽습니다.

### 작은 예시

`orders` topic에 partition 3개가 있습니다.

```
orders-0: offset 0,1,2,3...
orders-1: offset 0,1,2,3...
orders-2: offset 0,1,2,3...
```

`orderId=100`의 record가 `orders-1`에 들어갔다면, 같은 key를 쓰는 다음 record도 같은 partition으로 가야 그 key의 순서를 읽을 수 있습니다.

### 상태 이동 trace

```
producer record(key=order-100, value=paid)
  |
  v
partitioner chooses orders-1
  |
  v
broker leader appends at offset 42
  |
  v
consumer group member reads orders-1 offset 42
  |
  v
consumer commits next offset 43
```

### 내부 메커니즘

Kafka broker는 partition별 log segment file을 관리합니다. record는 기존 위치를 수정하기보다 log 끝에 append됩니다. offset은 partition 안의 position입니다. retention은 "소비자가 읽었는가"와 별개로 시간, 크기, compaction 정책에 따라 오래된 record를 지웁니다. 그래서 느린 consumer도 retention 안에서는 과거 record를 다시 읽을 수 있습니다.

log compaction topic에서는 같은 key의 최신값을 남기는 방식으로 log를 줄입니다. 이것은 event history를 모두 보존하는 retention과 목적이 다릅니다.

### 실패 모드

- topic 전체 ordering을 기대하면 partition 간 순서가 뒤섞인 것처럼 보입니다.
- retention을 consumer ack처럼 이해하면 읽지 않은 record가 삭제될 수 있습니다.
- partition 수를 늘리면 key-to-partition mapping이 바뀔 수 있어 ordering과 locality가 흔들립니다.

### 검증 방법

로컬 Kafka가 있다면:

```bash
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic orders --partitions 3 --replication-factor 1
kafka-console-producer.sh --bootstrap-server localhost:9092 --topic orders --property parse.key=true --property key.separator=:
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic orders --from-beginning --property print.key=true --property print.offset=true --property print.partition=true
```

PASS 신호:

- 같은 key가 같은 partition에 들어가는 것을 볼 수 있습니다.
- offset은 partition별로 증가합니다.

FAIL 신호:

- partition이 다르면 offset 숫자만으로 전역 순서를 만들 수 있다고 해석합니다.

### 면접식 되묻기

"Kafka에서 순서는 어떻게 보장되나요?"에는 "한 partition 안에서 append 순서와 offset으로 보장됩니다. topic 전체나 여러 partition 사이의 전역 순서는 기본 보장이 아닙니다"라고 답합니다. 꼬리 질문이 오면 key partitioning, producer ordering, retry/idempotence, consumer group assignment로 확장합니다.

### 흔한 오해와 반례

오해: "Kafka topic은 하나의 큰 queue다."

반례: topic은 여러 partition log의 묶음입니다. consumer group 안에서는 partition이 group member에게 나뉘지만, 서로 다른 consumer group은 같은 log를 독립적으로 읽습니다.

### Active recall

- offset은 topic 전체 번호인가요, partition 안 번호인가요?
- retention과 consumer offset commit의 차이를 설명해 보세요.

## 2. Producer, Broker, Consumer, Consumer Group

> producer는 record를 어느 partition leader에게 보낼지 결정하고, broker는 partition log와 replica를 관리하며, consumer는 자신의 group에서 할당받은 partition을 읽습니다.
> consumer group은 "같은 작업을 나눠 처리하는 소비자 집합"이고, group마다 offset이 따로 관리됩니다.
> Kafka scaling은 producer batching, partition count, broker I/O, consumer group parallelism이 함께 맞아야 합니다.

### 질문

consumer를 10개 띄우면 항상 처리량이 10배가 될까요?

### 직관

계산대가 10개 있어도 줄이 3개뿐이면 동시에 일할 수 있는 계산원은 최대 3명입니다. Kafka consumer group에서 병렬 처리의 상한은 보통 할당 가능한 partition 수와 처리 병목입니다.

### 작은 예시

topic partition이 3개이고 consumer group member가 5개입니다.

```
p0 -> c1
p1 -> c2
p2 -> c3
c4 -> idle
c5 -> idle
```

같은 group 안에서 한 partition은 한 번에 하나의 member에게만 할당되는 것이 기본 모델입니다.

### 상태 이동 trace

```
producer
  | metadata request: who leads p1?
  v
broker B is leader for p1
  |
  | ProduceRequest batch
  v
broker appends records
  |
  | FetchRequest from consumer c2
  v
consumer processes batch
  |
  | OffsetCommit next offset
  v
__consumer_offsets internal topic
```

### 내부 메커니즘

producer는 metadata를 통해 partition leader를 알고 batch를 모아 전송합니다. `acks` 설정은 leader 또는 replica ack를 어디까지 기다릴지 정합니다. broker는 network thread, request queue, log append, replica fetch, response path를 처리합니다. consumer는 poll loop로 fetch하고 처리한 뒤 offset을 commit합니다.

consumer group은 group coordinator와 rebalance protocol을 통해 partition ownership을 조정합니다. rebalance 중에는 일부 partition 처리가 멈추거나 중복 처리 가능성이 생길 수 있으므로 processing과 offset commit 순서가 중요합니다.

### 실패 모드

- processing 후 offset commit 전에 consumer가 죽으면 같은 record를 다시 처리할 수 있습니다.
- offset commit 후 processing 전에 죽으면 record가 유실된 것처럼 보일 수 있습니다.
- partition 수보다 consumer 수가 많으면 추가 consumer는 idle 상태가 됩니다.

### 검증 방법

로컬 Kafka:

```bash
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group orders-worker
```

PASS 신호:

- partition별 current offset, log end offset, lag, owner consumer를 볼 수 있습니다.

FAIL 신호:

- group lag만 보고 모든 partition이 균등하게 밀렸다고 가정합니다. partition별 lag 분포가 중요합니다.

### 면접식 되묻기

"consumer lag가 생기면 consumer만 늘리면 되나요?"에는 partition 수, partition별 lag, processing 병목, downstream 병목, rebalance, max poll interval, fetch size를 나눠 봐야 한다고 답합니다.

### 흔한 오해와 반례

오해: "Kafka가 record를 consumer에게 push한다."

반례: Kafka consumer는 pull 방식으로 fetch합니다. 이 구조는 consumer가 자기 처리 속도에 맞게 읽을 수 있고, broker가 느린 consumer 하나 때문에 전체 push를 막지 않게 합니다.

### Active recall

- 같은 consumer group과 다른 consumer group의 offset ownership 차이를 설명해 보세요.
- offset commit 순서가 중복/유실에 어떻게 영향을 주나요?

## 3. Replication, ISR, Leader Election

> Kafka partition은 leader replica가 append 순서를 정하고 follower replica가 leader log를 따라가는 구조입니다.
> ISR(in-sync replicas)은 leader를 충분히 따라오고 있는 replica 집합이며, committed record의 내구성 판단에 중요합니다.
> leader election은 availability와 durability tradeoff를 가지며, unclean leader election과 `min.insync.replicas`는 이 tradeoff를 드러내는 설정입니다.

### 질문

Kafka에서 record가 leader에 쓰이면 곧바로 안전하다고 말할 수 있을까요?

### 직관

한 장부 담당자가 기록을 적었지만 복사본 담당자들이 아직 따라 쓰지 못했다면 담당자 crash 때 기록을 잃을 수 있습니다. 복사본이 얼마나 따라왔는지를 봐야 합니다.

### 작은 예시

replication factor 3, replicas `B1, B2, B3`, leader `B1`, ISR `{B1, B2}`입니다.

```
producer acks=all
min.insync.replicas=2
record offset 100
B1 append ok
B2 append ok
B3 lagging
ack success
```

### 상태 이동 trace

```
producer -> leader B1 append offset 100
              |
              v
          follower B2 fetches offset 100
              |
              v
          high watermark advances
              |
              v
          record visible as committed
```

### 내부 메커니즘

Kafka follower는 leader에게 fetch request를 보내 log를 복제합니다. leader는 ISR 상태와 high watermark를 관리합니다. producer의 `acks=all`은 ISR 관련 ack를 기다립니다. `min.insync.replicas`는 충분한 ISR 수가 없을 때 write를 거절해 durability를 availability보다 우선할 수 있게 합니다.

leader가 죽으면 controller는 ISR 중 하나를 새 leader로 선택합니다. ISR 밖 replica를 leader로 세우는 unclean election은 availability를 높일 수 있지만 committed되지 않은 data loss 위험을 만듭니다. Kafka 4.x의 metadata 관리와 controller 세부는 KRaft 중심이며 버전 민감합니다.

### 실패 모드

- follower lag가 커지면 ISR에서 빠지고 write availability가 줄어들 수 있습니다.
- `acks=1`은 leader ack만 보고 성공하므로 leader crash 때 손실 위험이 더 큽니다.
- unclean leader election을 허용하면 더 오래 살아남을 수 있지만 log truncation/data loss 가능성이 커집니다.

### 검증 방법

로컬 multi-broker cluster가 있을 때:

```bash
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic orders
```

PASS 신호:

- partition별 leader, replicas, ISR을 확인합니다.

FAIL 신호:

- replicas 목록과 ISR 목록을 같은 의미로 봅니다. replicas는 배치된 복사본이고 ISR은 현재 따라오고 있는 복사본입니다.

### 면접식 되묻기

"Kafka는 consensus를 쓰나요?"에는 데이터 partition replication path와 metadata/controller quorum을 구분합니다. partition data는 leader-follower log replication과 ISR 규칙이 핵심이고, Kafka 4.x metadata 관리는 KRaft 기반 quorum이 중요합니다. 정확한 프로토콜 세부는 버전 문서를 확인해야 합니다.

### 흔한 오해와 반례

오해: "replication factor 3이면 항상 3개에 쓰인 뒤 성공한다."

반례: producer `acks`, ISR 상태, `min.insync.replicas`, broker failure에 따라 성공 조건이 달라집니다.

### Active recall

- ISR과 replicas의 차이를 설명해 보세요.
- `acks=all`과 `min.insync.replicas=2`가 함께 필요한 이유는 무엇인가요?

## 4. Delivery Semantics and Transactions

> Kafka delivery semantics는 producer retry, broker append, consumer processing, offset commit이 어떤 순서로 성공하거나 실패하는지에 달려 있습니다.
> at-least-once는 중복 가능성을 받아들이고, at-most-once는 유실 가능성을 받아들이며, exactly-once processing은 제한된 조건의 transaction/idempotence 조합으로 구성됩니다.
> 외부 DB나 API까지 자동으로 exactly-once가 되는 것은 아니므로 idempotency key와 outbox 같은 설계가 필요합니다.

### 질문

Kafka에서 "exactly-once"라고 하면 무엇이 exactly once인가요?

### 직관

우체국이 편지를 한 번만 배달한다고 해도, 편지를 받은 사람이 은행 송금을 두 번 실행하면 결과는 중복입니다. delivery와 processing side effect를 분리해야 합니다.

### 작은 예시

consumer가 input topic에서 offset 10을 읽고 output topic에 결과를 쓴 뒤 offset 11을 commit합니다.

```
read input offset 10
process
produce output record
commit consumer offset 11
```

output write와 offset commit이 따로 성공하면 crash 지점에 따라 중복이나 유실이 생깁니다.

### 상태 이동 trace

At-least-once:

```
read offset 10
process side effect
crash before offset commit
restart reads offset 10 again
side effect may repeat
```

Transactional Kafka-internal path:

```
begin transaction
read input offset 10
produce output
send offsets to transaction
commit transaction
read_committed consumers see committed output
```

### 내부 메커니즘

idempotent producer는 producer id와 sequence number를 사용해 broker가 중복 append를 거부할 수 있게 합니다. transaction은 여러 partition에 produce한 record와 consumer group offset commit을 하나의 atomic boundary로 묶는 데 쓰입니다. consumer가 `read_committed`를 사용하면 aborted transaction record를 숨길 수 있습니다.

이 보장은 Kafka 내부의 read-process-write에 강합니다. 외부 database write, email send, REST call 같은 side effect는 Kafka transaction과 같은 atomic boundary 안에 자동으로 들어오지 않습니다.

### 실패 모드

- auto commit이 processing보다 먼저 되면 crash 시 record가 건너뛰어질 수 있습니다.
- processing 후 commit 전 crash는 중복 처리를 만듭니다.
- producer retry와 broker append 사이에서 idempotence가 없으면 중복 record가 생길 수 있습니다.
- transaction timeout이나 fencing을 이해하지 못하면 zombie producer 문제가 생깁니다.

### 검증 방법

로컬에서는 crash 지점을 손으로 나누어 봅니다.

```text
Case A: process -> crash -> no offset commit
Expected: record re-read, duplicate-safe processing needed

Case B: offset commit -> crash -> no process
Expected: record skipped from this group perspective
```

PASS 신호:

- 중복/유실 가능성을 offset commit 위치로 설명합니다.

FAIL 신호:

- "Kafka는 exactly-once니까 application은 아무것도 안 해도 된다"고 결론 냅니다.

### 면접식 되묻기

"at-least-once로도 안전하게 만들 수 있나요?"에는 "가능하지만 side effect가 idempotent해야 합니다"라고 답합니다. payment request id, database unique key, outbox table, deduplication store를 예로 듭니다.

### 흔한 오해와 반례

오해: "Kafka transaction을 쓰면 외부 DB write도 exactly once다."

반례: 외부 DB가 Kafka transaction coordinator와 같은 atomic commit에 참여하지 않으면 crash 경계가 남습니다. outbox/inbox, idempotency key, two-phase commit 같은 별도 설계가 필요합니다.

### Active recall

- offset commit이 processing 전/후 어디에 있는지에 따라 어떤 실패가 생기나요?
- Kafka exactly-once processing의 안전한 범위를 말해 보세요.

## 5. Throughput, Batching, Backpressure, OS Links

> Kafka의 높은 처리량은 작은 요청을 크게 묶고, log append를 sequential I/O로 만들며, OS page cache와 sendfile 같은 경로를 활용하는 데서 나옵니다.
> 빠른 broker는 disk를 안 쓰는 broker가 아니라, disk와 network를 예측 가능한 큰 흐름으로 쓰는 broker입니다.
> 성능 문제는 producer buffer, broker request queue, page cache, disk, replica fetch, consumer processing, downstream sink 중 어디에서 밀리는지 찾아야 합니다.

### 질문

Kafka는 왜 disk에 저장하면서도 빠를 수 있을까요?

### 직관

흩어진 쪽지를 매번 창고 구석구석에 넣는 일은 느립니다. 같은 방향으로 계속 들어오는 상자를 컨베이어 벨트에 올리는 일은 빠릅니다. Kafka는 random update보다 append와 batch를 선택했습니다.

### 작은 예시

producer가 1 record씩 10,000번 보내는 대신 100개씩 100 batch를 보냅니다.

```
single-record: 10000 network round trips, many small writes
batched:        100 network round trips, larger sequential appends
```

### 상태 이동 trace

```
producer accumulator
  |
  | batch.size / linger.ms
  v
compressed record batch
  |
  v
broker network thread -> request queue
  |
  v
append to partition log file via page cache
  |
  v
follower fetch / consumer fetch
  |
  v
sendfile path from page cache to socket when eligible
```

### 내부 메커니즘

Kafka producer는 record를 batch로 모으고 compression을 적용할 수 있습니다. broker는 producer가 쓰는 format과 consumer가 읽는 format을 최대한 유지해 불필요한 변환을 줄입니다. log segment는 append 중심이라 disk seek를 줄입니다. OS page cache는 최근 log를 memory에 두고, catch-up이 아닌 consumer는 disk read 없이 page cache에서 fetch될 수 있습니다. Linux `sendfile()` 경로는 file page cache에서 socket으로 copy를 줄일 수 있습니다. 다만 TLS/SSL 경로에서는 이 최적화가 제한될 수 있습니다.

### 실패 모드

- producer가 너무 큰 batch를 기다리면 latency가 증가합니다.
- broker page cache가 밀리면 catch-up consumer가 disk를 강하게 읽습니다.
- follower lag가 커지면 ISR이 줄고 write availability가 낮아집니다.
- consumer processing이나 downstream DB가 느리면 lag가 증가합니다.

### 검증 방법

관찰 지점:

```bash
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group <group>
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic <topic>
iostat -xz 1 5
vmstat 1 5
ss -tin
```

PASS 신호:

- consumer lag 증가와 broker disk/network/CPU 중 어느 지표가 함께 움직이는지 연결합니다.

FAIL 신호:

- "Kafka가 느리다"를 단일 원인으로 처리합니다. producer, broker, replica, consumer, downstream 중 어느 queue가 자라는지 확인해야 합니다.

### 면접식 되묻기

"Kafka tuning에서 batch.size를 키우면 항상 좋은가요?"에는 throughput과 latency tradeoff를 말합니다. batch가 크면 syscall/network overhead와 compression ratio가 좋아질 수 있지만, linger와 memory 사용, tail latency가 늘 수 있습니다.

### 흔한 오해와 반례

오해: "Kafka가 page cache를 쓰므로 broker heap은 크게 잡을수록 좋다."

반례: heap을 과하게 키우면 OS page cache에 남는 memory가 줄고 GC 비용이 커질 수 있습니다. Kafka는 message cache를 JVM heap에 크게 두기보다 filesystem/page cache를 활용하는 설계입니다.

### Active recall

- Kafka의 fast path를 producer batch에서 consumer fetch까지 그려 보세요.
- TLS 사용 시 sendfile 최적화가 왜 제한될 수 있나요?

## 현실 시나리오 1: Kafka consumer lag가 계속 증가한다

> consumer lag는 "log end offset과 consumer committed offset 사이 거리"입니다.
> lag 증가는 broker가 느리다는 뜻일 수도 있지만, consumer processing, downstream sink, partition skew, rebalance, fetch 설정 문제일 수도 있습니다.

1. 관측된 증상

    `orders-worker` group lag가 0에서 3,000,000으로 증가합니다.

2. 가능한 원인 후보

    consumer CPU 부족, downstream DB latency, hot partition, broker fetch latency, rebalance 반복, max poll interval 초과, record 처리 오류 retry loop.

3. OS/kernel 관점에서 볼 지점

    consumer host의 CPU run queue, GC, socket queue, disk if local state store exists, broker host의 disk/network.

4. distributed-system 관점에서 볼 지점

    input rate와 processing rate, partition별 lag 분포, rebalance event, retry/idempotency, downstream backpressure를 봅니다.

5. Kafka 내부 구조와 연결되는 지점

    consumer group assignment, committed offset, log end offset, fetch batch, partition ordering boundary.

6. 확인 명령 또는 로그

```bash
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group orders-worker
jcmd <consumer-pid> Thread.print
vmstat 1 5
ss -tin
```

7. 잘못된 결론의 예

    "broker를 늘리면 lag가 줄 것이다."

8. 더 나은 추론 과정

    partition별 lag를 먼저 봅니다. 한 partition만 밀리면 hot key나 poison record를 의심합니다. 모든 partition이 비슷하게 밀리면 consumer fleet, downstream, broker fetch path를 비교합니다. 처리 후 offset commit 순서도 중복 가능성과 함께 확인합니다.

## 현실 시나리오 2: broker disk I/O가 갑자기 증가한다

> Kafka가 빠를 수 있는 이유는 disk를 쓰지 않아서가 아니라 sequential I/O와 page cache를 잘 활용하기 때문입니다.
> catch-up consumer, page cache miss, retention cleanup, log compaction, follower lag는 disk I/O를 다시 foreground 문제로 끌어올립니다.

1. 관측된 증상

    broker `iostat`에서 read await가 증가하고 consumer fetch latency가 늘었습니다.

2. 가능한 원인 후보

    오래 멈춘 consumer가 과거 segment를 읽음, page cache가 다른 workload에 밀림, tiered storage/remote read, log compaction, disk saturation, filesystem issue.

3. OS/kernel 관점에서 볼 지점

    page cache pressure, device await/util, major page fault, read throughput, writeback 상태.

4. distributed-system 관점에서 볼 지점

    consumer catch-up 위치, replica fetch lag, partition leadership distribution, retention/compaction policy.

5. Kafka 내부 구조와 연결되는 지점

    segment file, page cache hit/miss, consumer fetch, replica fetch, sendfile eligibility.

6. 확인 명령 또는 로그

```bash
iostat -xz 1 5
vmstat 1 5
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic <topic>
```

7. 잘못된 결론의 예

    "Kafka는 disk I/O가 없어야 하므로 버그다."

8. 더 나은 추론 과정

    어떤 fetch가 오래된 segment를 읽는지 확인합니다. page cache miss와 catch-up read가 맞다면 consumer lag 해소와 disk capacity를 함께 봅니다. compaction topic이면 cleaner I/O와 tombstone/key distribution을 확인합니다.

## 근거와 더 읽을 자료

- [10_source_ledger.md](10_source_ledger.md)의 Kafka 4.3 design, hardware/OS, man7 sendfile, Linux page cache 항목
- Apache Kafka 4.3 Design: https://kafka.apache.org/43/design/design/
- Kafka hardware and OS operations: https://kafka.apache.org/43/operations/hardware-and-os/
- Linux sendfile: https://man7.org/linux/man-pages/man2/sendfile.2.html

