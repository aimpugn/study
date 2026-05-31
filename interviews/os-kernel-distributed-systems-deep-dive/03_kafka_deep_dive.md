# 03. Kafka Deep Dive

Kafka를 queue로만 이해하면 빠르게 막힙니다. Kafka는 record를 broker의 partition log에 append하고, consumer가 자기 offset을 따라가며 다시 읽는 시스템입니다. queue처럼 쓸 수는 있지만, 내부의 중심은 "메시지를 하나씩 건네주고 지우는 통"이 아니라 "순서 있는 log를 오래 보존하고 여러 consumer가 각자 위치를 기억하며 읽는 구조"입니다.

Kafka가 왜 빠른지, 왜 ordering이 partition 안에서만 자연스러운지, 왜 consumer lag가 생기는지, 왜 exactly-once라는 말이 조심스러운지 모두 이 log 모델에서 나옵니다.

## 1. Record는 queue item이 아니라 partition log의 한 entry가 된다

producer가 `send()`를 호출하면 record는 바로 consumer에게 가지 않습니다. producer client는 record를 memory buffer에 모아 batch를 만들고, key가 있으면 partitioner가 key를 기준으로 partition을 고릅니다. broker는 해당 partition leader가 받은 batch를 log segment 파일에 append합니다. append는 file offset 뒤에 byte를 붙이는 방식이라 random update보다 filesystem과 storage에 유리합니다.

```
producer record
  -> producer buffer / batch
  -> partition selection
  -> broker partition leader
  -> append to log segment
  -> page cache / filesystem
  -> replica fetchers copy the log
  -> consumer fetches from offset N
```

topic은 논리적 이름이고, partition은 실제 ordering과 parallelism의 단위입니다. 같은 partition 안에서는 offset이 증가하며 순서가 생깁니다. 그러나 topic 전체에 partition이 여러 개라면 partition 사이의 전역 순서는 기본 약속이 아닙니다. 사용자 ID를 key로 잡으면 같은 사용자의 record가 같은 partition에 모여 순서를 지키기 쉽지만, hot key가 있으면 특정 partition만 바빠집니다.

Kafka의 log는 retention 때문에 queue와 다릅니다. consumer가 읽었다고 record가 즉시 사라지지 않습니다. retention time이나 size 정책에 따라 보존되고, consumer group마다 offset을 따로 관리합니다. 그래서 같은 topic을 fraud detector, billing job, analytics job이 각자 다른 속도로 읽을 수 있습니다.

## 2. Broker는 page cache와 순차 I/O를 적극적으로 빌린다

Kafka broker가 빠르다는 말은 "디스크를 안 쓴다"는 뜻이 아닙니다. 오히려 Kafka는 log segment라는 파일을 중심에 놓고, append-only write와 sequential read가 kernel page cache와 storage에 잘 맞도록 설계합니다. producer가 보낸 record는 broker process의 heap에 영원히 머무르지 않고 log file로 append됩니다. 이때 운영체제는 해당 file page를 page cache에 두고 나중에 writeback할 수 있습니다.

consumer가 최근 record를 읽으면 그 byte는 이미 page cache에 있을 가능성이 높습니다. broker는 file에서 socket으로 보내는 경로에서 불필요한 user-space copy를 줄일 수 있습니다. Apache Kafka design docs는 filesystem, page cache, batching, sendfile 같은 OS 기능을 Kafka 성능 모델의 중요한 일부로 설명합니다.

```
append path:
  network receive buffer
    -> broker request handling
    -> log append
    -> page cache dirty page
    -> later writeback

fetch path:
  consumer fetch request
    -> log segment bytes
    -> page cache if hot
    -> socket send path
    -> consumer
```

여기서 `acks=all`이나 replication은 `fsync()`와 같은 말이 아닙니다. replication ack는 leader와 follower replica가 Kafka protocol의 조건을 만족했다는 뜻이고, local disk flush 정책은 별도의 설정과 filesystem/storage 경계에 기대는 문제입니다. 내구성 질문이 나오면 "page cache에 있음", "broker process가 ack함", "replica가 복제함", "storage device가 flush함"을 분리해야 합니다.

이 구분은 [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md)의 파일 경로와 그대로 연결됩니다. broker가 segment file에 append할 때 사용자 공간의 record byte는 system call을 거쳐 page cache의 dirty page가 되고, background writeback이나 flush 정책을 통해 block layer와 device queue로 내려갑니다. follower가 leader log를 fetch하는 동안 같은 disk와 page cache는 read path로도 쓰입니다. 따라서 produce latency가 높을 때는 Kafka request queue만 볼 것이 아니라 dirty writeback, disk queue, page cache hit, follower fetch lag를 함께 보아야 합니다.

```
producer batch reaches broker
  -> network receive queue
  -> request handler thread
  -> append to segment file
  -> page cache dirty page
  -> replica fetch reads same segment range
  -> socket send buffer to follower/consumer
```

## 3. Replication과 ISR은 어느 replica가 따라오고 있는지 관리한다

Kafka partition에는 leader가 있고 follower가 있습니다. producer와 consumer의 일반 요청은 leader를 중심으로 처리됩니다. follower는 leader의 log를 fetch해 따라옵니다. ISR(in-sync replicas)은 leader와 충분히 동기화되어 있다고 간주되는 replica 집합입니다. producer가 `acks=all`을 요구하면 leader는 필요한 ISR 조건을 만족할 때 성공을 반환합니다.

```
leader log:
  offset 10 11 12 13

follower A:
  offset 10 11 12 13   in sync

follower B:
  offset 10 11         lagging, may fall out of ISR

committed / high watermark:
  replicas that must have the record determine what is safe to expose
```

leader가 죽으면 다른 replica가 leader가 됩니다. 이때 어떤 replica를 leader로 뽑을지, leader가 가진 log가 어디까지 안전한지, consumer에게 어느 offset까지 보여 줄지 모두 correctness와 관련됩니다. 단순히 "replica가 있으니 안전하다"가 아니라 "어떤 offset이 몇 replica에 있으며, leader election 뒤 log가 어떻게 잘리거나 이어지는가"를 봐야 합니다.

Kafka의 modern metadata path는 ZooKeeper에서 KRaft로 이동했습니다. 이 문서는 Kafka 4.x 문서 기준의 KRaft 중심 설명을 우선하며, ZooKeeper-era 상세 동작은 version-sensitive 또는 historical로 다룹니다.

network 관점에서는 leader와 follower의 관계가 단순한 method call이 아닙니다. follower fetch request도 [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md)의 packet path를 지나고, socket receive queue, broker event loop, request handler, page cache read, socket send buffer를 거칩니다. follower가 lagging이라는 말은 follower process가 느리다는 한 문장이 아니라, leader disk read, leader network send, follower network receive, follower append, follower disk write 중 어느 상태가 밀렸는지 다시 나누어야 한다는 뜻입니다.

## 4. Consumer group은 병렬성과 소유권을 offset으로 표현한다

consumer group은 여러 consumer가 topic partition을 나눠 읽는 구조입니다. 같은 group 안에서는 한 partition이 보통 한 consumer에게 assign됩니다. 이렇게 해야 partition 안의 순서와 offset 진행을 단순하게 유지할 수 있습니다. consumer가 늘어도 partition 수보다 많이 늘리면 더 이상 병렬성이 늘지 않는 이유가 여기 있습니다.

consumer offset은 "이 group이 이 partition에서 어디까지 처리했다고 기록했는가"입니다. offset commit 시점은 delivery semantics와 직접 연결됩니다.

```
fetch offset 100
  -> process record 100
  -> write result to DB
  -> commit offset 101
```

DB write 전에 offset을 commit하면 장애 시 record 100을 건너뛸 수 있습니다. DB write 후 offset commit 전에 죽으면 record 100을 다시 처리할 수 있습니다. 그래서 at-least-once는 중복 가능성을 받아들이고 idempotent 처리로 막습니다. at-most-once는 유실 가능성을 받아들입니다. Kafka transactions와 idempotent producer는 Kafka 내부 read-process-write 경로에서 더 강한 exactly-once processing을 만들 수 있지만, 외부 DB나 HTTP side effect까지 자동으로 원자화하지는 않습니다.

## 5. Backpressure는 lag와 queue로 보인다

Kafka에서 lag는 consumer group의 committed offset이 partition log의 end offset을 따라가지 못할 때 생깁니다. lag 자체는 원인이 아니라 결과입니다. 원인은 consumer CPU, downstream DB latency, broker fetch latency, network, partition skew, rebalance, GC, disk I/O 등 다양합니다.

```
producer rate > consumer processing rate
  -> log end offset grows
  -> committed offset moves slowly
  -> lag grows
  -> retention window risk appears
```

lag를 줄이려고 consumer 수만 늘리면 partition 수, key distribution, downstream capacity가 곧 한계가 됩니다. partition이 12개면 같은 group에서 동시에 active하게 읽을 수 있는 consumer도 보통 12개가 상한입니다. downstream DB가 병목이면 consumer를 늘릴수록 DB retry와 timeout만 늘 수 있습니다.

producer 쪽도 backpressure가 있습니다. broker가 느리거나 network가 막히면 producer buffer가 차고, `linger.ms`, `batch.size`, `buffer.memory`, `max.in.flight.requests.per.connection` 같은 설정이 latency와 throughput, ordering에 영향을 줍니다. 설정 이름을 외우기보다 "어느 queue가 자라고 어느 순서 보장이 깨질 수 있는가"를 먼저 봐야 합니다.

OS scheduler도 이 backpressure를 가릴 수 있습니다. [01a_process_scheduling.md](01a_process_scheduling.md)에서 본 것처럼 socket이 readable이 되어도 network thread가 CPU를 받지 못하면 request 처리는 늦어집니다. producer client도 non-blocking network I/O와 selector loop를 사용하므로, user thread가 `send()`를 호출한 시점과 broker가 실제 batch를 받은 시점 사이에는 producer buffer, selector thread, kernel socket send buffer, TCP window가 있습니다.

## 6. Compaction과 retention은 log를 어떻게 오래 살릴지 정한다

retention은 시간이 지나거나 크기가 커지면 오래된 segment를 지우는 정책입니다. log compaction은 key별 최신 값을 남기는 방식으로 log를 줄이는 정책입니다. compaction topic은 changelog나 table-like state 복원에 유용하지만, 모든 event history를 보존하는 용도와는 다릅니다.

예를 들어 key `user-1`에 대해 `email=a`, `email=b`, `delete` record가 들어오면 compacted log는 어느 시점 이후 최신 상태나 tombstone 중심으로 줄어들 수 있습니다. 이것은 "queue에서 소비했으니 삭제"와 다릅니다. Kafka log는 consumer와 독립적으로 보존 정책에 의해 정리됩니다.

## 현실 시나리오 1: "Kafka는 왜 빠른가요?"

짧게 답하면 Kafka는 partition log에 순차 append하고, producer/broker/consumer가 batch로 일하며, OS page cache와 file-to-socket 전송 경로를 활용하고, consumer가 pull 방식으로 자기 속도에 맞게 읽기 때문에 빠릅니다. 하지만 이 답은 "디스크를 안 써서 빠르다"가 아닙니다. 오히려 디스크를 쓰되 append와 sequential read가 잘 맞게 설계한 것입니다.

꼬리 질문이 오면 이렇게 내려갑니다.

```
producer batch
  -> leader append to segment
  -> page cache and replication
  -> high watermark / visibility
  -> consumer fetch
  -> offset commit
```

각 단계에서 비용은 다릅니다. 작은 message를 너무 자주 보내면 syscall과 network overhead가 커집니다. follower가 느리면 ISR과 acks latency가 영향을 받습니다. consumer가 느리면 lag가 늘고 retention 위험이 생깁니다.

## 현실 시나리오 2: "exactly-once면 중복이 절대 없나요?"

아닙니다. Kafka의 exactly-once는 범위가 있습니다. idempotent producer는 producer retry로 생기는 중복 append를 줄입니다. transaction은 Kafka topic에서 읽고 다른 Kafka topic에 쓰며 offset commit까지 transaction에 묶는 경로를 지원합니다. consumer가 외부 DB에 쓰거나 HTTP API를 호출하면 그 외부 시스템까지 Kafka transaction이 자동으로 감싸지는 것은 아닙니다.

따라서 실무 답변은 "Kafka 내부의 지원된 read-process-write 경로에서는 정확히 한 번 처리처럼 보이도록 만들 수 있지만, 외부 side effect에는 idempotency나 별도 transaction/outbox 설계가 필요합니다"가 안전합니다.

## 문서를 덮고 확인할 것

- Kafka topic, partition, segment, offset을 서로 구분해 보세요.
- 같은 topic 안에서도 partition 사이 전역 순서가 없는 이유를 설명해 보세요.
- `acks=all`, ISR, fsync, replication ack의 차이를 말해 보세요.
- consumer offset commit 위치가 유실과 중복 중 무엇을 바꾸는지 trace로 설명해 보세요.
- consumer lag가 생겼을 때 consumer 수 증가가 해법이 아닐 수 있는 이유를 말해 보세요.

## 근거와 더 읽을 자료

- Apache Kafka 4.3 documentation and design docs.
- Kafka original paper, "Kafka: a Distributed Messaging System for Log Processing", for historical motivation.
- Linux `sendfile(2)` and page cache documentation for OS-level transfer/caching model.
