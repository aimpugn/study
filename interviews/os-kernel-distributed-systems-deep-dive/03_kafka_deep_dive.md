# 03. Kafka Deep Dive

## 목차

- [1. Record는 queue item이 아니라 partition log의 한 entry가 된다](#1-record는-queue-item이-아니라-partition-log의-한-entry가-된다)
- [2. Broker는 page cache와 순차 I/O를 적극적으로 빌린다](#2-broker는-page-cache와-순차-io를-적극적으로-빌린다)
- [3. Replication과 ISR은 어느 replica가 따라오고 있는지 관리한다](#3-replication과-isr은-어느-replica가-따라오고-있는지-관리한다)
- [4. Consumer group은 병렬성과 소유권을 offset으로 표현한다](#4-consumer-group은-병렬성과-소유권을-offset으로-표현한다)
- [5. Backpressure는 lag와 queue로 보인다](#5-backpressure는-lag와-queue로-보인다)
- [6. Compaction과 retention은 log를 어떻게 오래 살릴지 정한다](#6-compaction과-retention은-log를-어떻게-오래-살릴지-정한다)
- [현실 시나리오 1: "Kafka는 왜 빠른가요?"](#현실-시나리오-1-kafka는-왜-빠른가요)
- [현실 시나리오 2: "exactly-once면 중복이 절대 없나요?"](#현실-시나리오-2-exactly-once면-중복이-절대-없나요)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [근거와 더 읽을 자료](#근거와-더-읽을-자료)
- [Producer append path를 OS와 replication path로 동시에 읽기](#producer-append-path를-os와-replication-path로-동시에-읽기)
- [Segment, index, retention은 log를 파일 묶음으로 관리하는 방식이다](#segment-index-retention은-log를-파일-묶음으로-관리하는-방식이다)
- [Zero-copy와 sendfile은 "복사를 줄인다"이지 모든 비용을 없애지 않는다](#zero-copy와-sendfile은-복사를-줄인다이지-모든-비용을-없애지-않는다)
- [ISR, high watermark, leader epoch는 보이는 log의 경계를 만든다](#isr-high-watermark-leader-epoch는-보이는-log의-경계를-만든다)
- [Consumer group, offset commit, rebalance를 처리 의미로 읽기](#consumer-group-offset-commit-rebalance를-처리-의미로-읽기)
- [Kafka transaction과 outbox 경계](#kafka-transaction과-outbox-경계)
- [OS pressure가 Kafka metric으로 올라오는 경로](#os-pressure가-kafka-metric으로-올라오는-경로)
- [Interview replay: Kafka를 log, OS, distributed boundary로 설명하기](#interview-replay-kafka를-log-os-distributed-boundary로-설명하기)
- [Consumer lag를 원인별로 쪼개는 법](#consumer-lag를-원인별로-쪼개는-법)
- [Controller와 metadata quorum은 data log와 다른 control path다](#controller와-metadata-quorum은-data-log와-다른-control-path다)
- [Retention과 compaction은 log의 의미를 바꾼다](#retention과-compaction은-log의-의미를-바꾼다)

Kafka를 queue로만 이해하면 빠르게 막힙니다. Kafka는 record를 broker의 partition log에 append하고, consumer가 자기 offset을 따라가며 다시 읽는 시스템입니다. queue처럼 쓸 수는 있지만, 내부의 중심은 "메시지를 하나씩 건네주고 지우는 통"이 아니라 "순서 있는 log를 오래 보존하고 여러 consumer가 각자 위치를 기억하며 읽는 구조"입니다.

Kafka가 왜 빠른지, 왜 ordering이 partition 안에서만 자연스러운지, 왜 consumer lag가 생기는지, 왜 exactly-once라는 말이 조심스러운지 모두 이 log 모델에서 나옵니다.

## 1. Record는 queue item이 아니라 partition log의 한 entry가 된다

producer가 `send()`를 호출하면 record는 바로 consumer에게 가지 않습니다. producer client는 record를 memory buffer에 모아 batch를 만들고, key가 있으면 partitioner가 key를 기준으로 partition을 고릅니다. broker는 해당 partition leader가 받은 batch를 log segment 파일에 append합니다. append는 file offset 뒤에 byte를 붙이는 방식이라 random update보다 filesystem과 storage에 유리합니다.

```text
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

Kafka broker가 빠르다는 말은 "디스크를 안 쓴다"는 뜻이 아닙니다. 오히려 Kafka는 log segment라는 파일을 중심에 놓고, append-only write와 sequential read가 kernel page cache와 storage에 잘 맞도록 설계합니다. producer가 보낸 record는 broker process의 heap에 영원히 머무르지 않고 log file로 append됩니다. 이때 운영체제는 해당 file page를 page cache에 두고 이후 writeback할 수 있습니다.

consumer가 최근 record를 읽으면 그 byte는 이미 page cache에 있을 가능성이 높습니다. broker는 file에서 socket으로 보내는 경로에서 불필요한 user-space copy를 줄일 수 있습니다. Apache Kafka design docs는 filesystem, page cache, batching, sendfile 같은 OS 기능을 Kafka 성능 모델의 중요한 일부로 설명합니다.

```text
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

```text
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

```text
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

```text
fetch offset 100
  -> process record 100
  -> write result to DB
  -> commit offset 101
```

DB write 전에 offset을 commit하면 장애 시 record 100을 건너뛸 수 있습니다. DB write 후 offset commit 전에 죽으면 record 100을 다시 처리할 수 있습니다. 그래서 at-least-once는 중복 가능성을 받아들이고 idempotent 처리로 막습니다. at-most-once는 유실 가능성을 받아들입니다. Kafka transactions와 idempotent producer는 Kafka 내부 read-process-write 경로에서 더 강한 exactly-once processing을 만들 수 있지만, 외부 DB나 HTTP side effect까지 자동으로 원자화하지는 않습니다.

## 5. Backpressure는 lag와 queue로 보인다

Kafka에서 lag는 consumer group의 committed offset이 partition log의 end offset을 따라가지 못할 때 생깁니다. lag 자체는 원인이 아니라 결과입니다. 원인은 consumer CPU, downstream DB latency, broker fetch latency, network, partition skew, rebalance, GC, disk I/O 등 다양합니다.

```text
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

```text
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

## Producer append path를 OS와 replication path로 동시에 읽기

Kafka produce request는 "broker에 message를 보낸다"로 끝나지 않습니다. Producer는 record를 batch로 묶고, partitioner는 topic partition을 고르며, broker leader는 request를 받아 partition log에 append합니다. 이 append는 user-space Kafka code, JVM memory, OS page cache, filesystem, block I/O를 지나고, 동시에 follower replication과 ISR 조건을 기다릴 수 있습니다. `acks=all`이라는 설정은 이 전체 중 어느 지점까지 기다릴지를 바꾸는 분산 의미입니다.

```text
producer
  -> batches records
  -> sends ProduceRequest over TCP

leader broker
  -> network thread reads request
  -> request handler validates and appends to log buffer/file path
  -> bytes enter OS page cache as dirty pages
  -> follower fetchers copy records
  -> ISR condition and high watermark move
  -> broker returns ack according to acks/min.insync.replicas
```

여기서 page cache와 replication ack를 분리해야 합니다. Leader가 bytes를 page cache에 반영했다는 것과 storage device에 sync되었다는 것, follower가 받았다는 것, high watermark가 이동했다는 것, consumer에게 보일 수 있다는 것은 서로 다른 사건입니다. Kafka는 일반적으로 OS page cache와 sequential append를 적극 활용해 throughput을 얻습니다. 그래서 broker heap에 모든 record를 보관하는 시스템이 아닙니다. 하지만 page cache에 있다는 말은 crash 뒤 항상 durable하다는 뜻이 아니며, durability는 replication, flush policy, election policy, storage behavior를 함께 봐야 합니다.

`acks=0`은 broker ack를 기다리지 않습니다. `acks=1`은 leader 응답을 기다립니다. `acks=all`은 in-sync replica 조건을 만족할 때까지 기다립니다. `min.insync.replicas`가 충분히 높으면 leader 하나만 받는 상황을 막을 수 있지만, replica가 느리거나 빠지면 availability와 latency가 영향을 받습니다. Unclean leader election 같은 정책은 availability와 data loss 위험 사이의 tradeoff입니다. 따라서 "acks=all이면 디스크에 안전하다"가 아니라 "Kafka replication protocol의 특정 조건을 기다린다"가 정확합니다.

## Segment, index, retention은 log를 파일 묶음으로 관리하는 방식이다

Kafka partition log는 하나의 추상 log처럼 보이지만, 실제로는 segment file과 index file의 묶음입니다. Segment는 일정 크기나 시간 기준으로 roll되고, retention 정책에 따라 오래된 segment가 삭제됩니다. Offset index와 time index는 특정 offset이나 timestamp 근처의 file position을 빠르게 찾게 도와줍니다. 이 구조는 OS filesystem과 page cache 위에서 동작합니다.

```text
partition directory
  00000000000000000000.log
  00000000000000000000.index
  00000000000000000000.timeindex
  00000000000000123456.log
  ...
```

Log append는 sequential write에 가깝고, consumer fetch도 많은 경우 sequential read입니다. 이 access pattern은 HDD와 SSD, page cache, readahead에 유리합니다. 하지만 partition 수가 너무 많고 segment가 잘게 쪼개지면 file descriptor, page cache locality, filesystem metadata, recovery time, controller metadata가 부담이 됩니다. Retention이 길면 storage 사용량이 늘고, retention이 짧으면 느린 consumer가 필요한 offset을 잃을 수 있습니다.

Kafka startup이나 recovery에서 broker는 log directory를 scan하고 segment의 끝과 index를 확인해야 합니다. Crash 뒤 partial record나 index mismatch가 있으면 복구 경로가 작동합니다. 이 부분은 filesystem crash consistency와 application-level log recovery가 만나는 자리입니다. Filesystem journal이 Kafka record boundary를 이해하는 것은 아니므로, Kafka는 record format, checksum, offset, index rebuild로 자기 의미를 복구합니다.

## Zero-copy와 sendfile은 "복사를 줄인다"이지 모든 비용을 없애지 않는다

Consumer fetch에서 broker가 file에 있는 log data를 client socket으로 보내려면 원래는 storage/page cache에서 user buffer로 copy하고 다시 kernel socket path로 copy할 수 있습니다. `sendfile()` 같은 file-to-socket 경로는 file page cache에서 socket으로 data를 보내는 과정에서 user-space copy를 줄입니다. Kafka design에서 page cache와 zero-copy가 자주 함께 등장하는 이유입니다.

```text
traditional path
  disk/page cache -> kernel buffer -> user buffer -> kernel socket buffer -> NIC

sendfile-like path
  page cache file pages -> socket send path -> NIC
  fewer user/kernel copies
```

하지만 zero-copy는 조건부 이점입니다. TLS encryption이 broker process에서 일어나면 bytes를 암호화해야 하므로 data path가 바뀝니다. Compression된 record batch는 이미 compressed bytes로 저장되지만, fetch path에서 protocol header나 filtering, authorization, quota, throttling이 붙습니다. Page cache hit가 낮으면 storage I/O가 병목입니다. NIC와 TCP congestion window가 작으면 network가 병목입니다. Follower가 느리면 replication lag가 병목입니다. "Kafka는 zero-copy라 빠르다"가 아니라 "Kafka는 log segment와 page cache, sendfile 계열 경로가 잘 맞는 access pattern을 설계했다"라고 말해야 합니다.

## ISR, high watermark, leader epoch는 보이는 log의 경계를 만든다

Kafka replication에서 leader는 partition의 append authority입니다. Follower는 leader로부터 records를 fetch해 따라옵니다. ISR(in-sync replicas)은 leader를 충분히 따라오고 있다고 간주되는 replica 집합입니다. High watermark는 모든 ISR이 복제한 것으로 간주되는 log 위치이며, consumer에게 노출 가능한 committed boundary로 작동합니다. Leader epoch는 leader 시대를 구분해 stale leader나 truncation 문제를 다루는 이름표입니다.

```text
leader log:   offsets 0 1 2 3 4 5
follower A:   offsets 0 1 2 3 4
follower B:   offsets 0 1 2 3

high watermark may be 3
  -> records above boundary are not yet fully committed for visibility
```

이 경계가 필요한 이유는 follower가 뒤처질 수 있기 때문입니다. Leader가 offset 5까지 받았더라도 ISR replica들이 거기까지 복제하지 못했다면, leader crash 뒤 새 leader가 될 replica에는 그 record가 없을 수 있습니다. High watermark를 통해 consumer visibility와 durability expectation을 조절합니다. 물론 정확한 behavior는 Kafka version과 configuration, metadata quorum, election policy에 따라 달라집니다. Source ledger에 version boundary를 둬야 하는 이유입니다.

Leader election은 단순히 "다른 broker가 leader가 된다"가 아닙니다. 새 leader가 어떤 log end를 갖고 있고, follower가 어떤 위치로 truncate하거나 catch up해야 하며, producer가 어떤 epoch의 response를 받았는지 모두 중요합니다. Producer idempotence와 sequence number, transactions는 이 불확실성 안에서 duplicate와 ordering을 줄이기 위한 protocol입니다.

## Consumer group, offset commit, rebalance를 처리 의미로 읽기

Consumer group은 topic partition을 group member들에게 나눠 맡기는 coordination mechanism입니다. 한 partition은 한 group 안에서 보통 한 consumer member에게 할당됩니다. Consumer는 fetch한 records를 처리하고, 어디까지 처리했는지 offset을 commit합니다. Offset commit 위치는 장애 뒤 중복 처리와 유실 사이의 경계를 바꿉니다.

```text
fetch records offsets 100..199
  -> process records
  -> commit offset 200

commit before processing
  -> crash can lose records

commit after processing
  -> crash can duplicate records
```

Rebalance는 group membership이나 partition assignment가 바뀔 때 발생합니다. Rebalance 중에는 일부 partition 처리가 멈추거나 지연될 수 있고, cooperative rebalance 같은 protocol은 이를 줄이려 합니다. Generation은 consumer group의 membership 시대를 구분합니다. 오래된 generation의 member가 늦게 commit하거나 heartbeat하면 coordinator가 거부할 수 있습니다. 이 역시 distributed systems의 epoch pattern입니다.

Lag는 단순히 consumer가 느리다는 뜻이 아닙니다. Broker fetch latency, network, consumer processing, downstream sink, partition skew, rebalance, GC, offset commit 전략이 모두 lag로 올라올 수 있습니다. Lag를 줄이려고 consumer 수를 늘려도 partition 수보다 많으면 병렬성이 늘지 않고, 특정 partition hot key가 원인이면 전체 consumer 수 증가가 효과가 작습니다. Downstream DB가 느린데 consumer를 늘리면 DB 과부하만 키울 수 있습니다.

## Kafka transaction과 outbox 경계

Kafka transaction은 Kafka 안의 read-process-write와 offset commit을 묶는 강한 도구입니다. Idempotent producer는 producer retry가 duplicate append를 만들지 않도록 producer id, sequence number를 사용합니다. Transactional producer는 여러 partition write와 consumed offset commit을 하나의 transaction으로 묶어 consumer가 committed data만 읽도록 할 수 있습니다. 하지만 외부 DB나 HTTP side effect까지 자동으로 transaction에 들어가는 것은 아닙니다.

```text
supported Kafka transaction shape
  read input topic
  -> process
  -> produce output topic records
  -> commit consumed offsets in transaction

outside boundary
  write to external DB
  call payment API
  send email
  mutate object storage
```

외부 side effect가 있으면 outbox, idempotency key, sink transaction, two-phase commit connector, exactly-once capable sink 같은 별도 설계가 필요합니다. "exactly-once"라는 표현은 사용 범위를 말하지 않으면 위험합니다. Kafka 내부에서 지원되는 처리 경로와 바깥 side effect를 분리해야 실무 답변이 안전합니다.

## OS pressure가 Kafka metric으로 올라오는 경로

Kafka metric은 제품 언어로 보이지만 OS 자원 경로와 자주 연결됩니다. Request queue time이 늘면 network thread와 request handler, scheduler, CPU quota를 봅니다. Local time이 늘면 log append, page cache, filesystem, disk queue를 봅니다. Remote time이 늘면 follower replication과 network를 봅니다. Consumer lag가 늘면 broker fetch path와 consumer processing, downstream sink를 함께 봅니다.

```text
produce latency
  -> network receive
  -> request queue
  -> append to page cache/log
  -> replica fetch/ISR ack
  -> response send

possible OS wait
  scheduler delay
  lock/futex
  dirty page writeback
  disk fsync or queue
  socket buffer/backpressure
  cgroup CPU/memory throttle
```

Kafka troubleshooting은 broker log만으로 닫히지 않습니다. `iostat`으로 disk await와 utilization을 보고, `ss`로 socket queue를 보고, GC log로 pause를 보고, `perf`나 async-profiler로 CPU를 보고, cgroup stats로 quota/throttle을 봅니다. Page cache hit/miss를 직접 보기 어렵더라도 disk read pattern과 cache pressure, segment access를 통해 추론할 수 있습니다.

## Interview replay: Kafka를 log, OS, distributed boundary로 설명하기

Kafka를 한 문장으로 설명하면 "partitioned append log를 중심으로 producer와 consumer가 offset을 통해 진행 상태를 나누고, broker replication과 high watermark로 visibility와 durability boundary를 관리하는 system"입니다. 여기서 OS 연결을 붙이면 "log segment는 filesystem과 page cache 위에 있고, fetch path는 page cache와 file-to-socket 전송 최적화의 이점을 받으며, disk/network/scheduler/backpressure가 latency로 올라온다"가 됩니다.

꼬리 질문에는 네 단계로 내려가면 됩니다. Produce는 batch -> leader append -> page cache/log segment -> follower replication -> ack입니다. Consume은 fetch -> socket receive -> processing -> offset commit입니다. Durability는 `acks`, ISR, high watermark, flush/election policy를 분리해야 합니다. Exactly-once는 Kafka 내부 transaction boundary와 외부 side effect를 분리해야 합니다. 이 네 단계를 말할 수 있으면 Kafka를 단순 message queue가 아니라 OS와 분산 log가 만나는 시스템으로 설명할 수 있습니다.

## Consumer lag를 원인별로 쪼개는 법

Lag는 consumer group이 log end를 따라가지 못한다는 결과 지표입니다. 원인은 여러 층에 있습니다. Broker가 fetch를 늦게 응답할 수도 있고, network가 막힐 수도 있고, consumer processing이 느릴 수도 있으며, downstream DB가 병목일 수도 있습니다. Partition skew가 있으면 group 전체 consumer 수가 충분해도 특정 partition lag만 계속 늘 수 있습니다. Rebalance가 잦으면 처리보다 assignment 변화에 시간을 쓸 수 있습니다.

```text
lag grows
  -> broker log end moves fast
  -> consumer fetch returns slowly?
  -> consumer processes slowly?
  -> offset commit delayed?
  -> one partition hot?
  -> downstream sink backpressures?
```

해결책은 원인별로 다릅니다. Consumer processing이 CPU-bound이면 worker parallelism이나 code path를 봅니다. Downstream sink가 느리면 consumer를 늘리는 것이 오히려 sink overload를 키울 수 있습니다. Partition skew면 key design이나 partition 수를 봐야 합니다. Broker fetch latency가 문제면 broker disk/page cache/network/replication을 봅니다. Rebalance가 문제면 heartbeat, session timeout, max poll interval, cooperative protocol, processing batch size를 봅니다. Lag 하나를 보고 consumer 수부터 늘리는 것은 깊은 답변이 아닙니다.

## Controller와 metadata quorum은 data log와 다른 control path다

Kafka에는 record data path뿐 아니라 metadata/control path가 있습니다. Topic, partition, leader, ISR, broker membership 같은 metadata를 누가 관리하고 어떻게 합의하는지가 cluster stability를 좌우합니다. ZooKeeper 기반 과거 구조와 KRaft 기반 구조는 다르므로 version을 붙여 말해야 합니다. 중요한 것은 metadata quorum이나 controller가 data log append와 같은 path가 아니라, cluster의 authority와 epoch를 관리하는 별도 control plane이라는 점입니다.

```text
data path
  producer/consumer records
  -> partition leader log
  -> follower replication

control path
  broker registration
  partition leadership
  topic metadata
  -> controller / metadata quorum
```

Controller가 불안정하면 produce/fetch 자체의 disk path가 정상이어도 leader election, metadata propagation, client metadata refresh가 흔들릴 수 있습니다. Broker가 많고 partition이 매우 많으면 metadata update와 controller load도 커집니다. 운영에서 Kafka를 볼 때 data plane metric과 control plane metric을 구분해야 합니다.

## Retention과 compaction은 log의 의미를 바꾼다

Kafka retention은 오래된 record를 삭제해 storage를 관리합니다. Log compaction은 key별 최신 value를 남기는 방식으로 topic을 compact할 수 있습니다. Retention topic은 event stream에 가깝고, compacted topic은 changelog나 table snapshot 성격을 가질 수 있습니다. 두 정책은 consumer replay 가능성과 storage 사용량, tombstone 처리에 영향을 줍니다.

Consumer가 오래 멈춰 retention window를 지나면 필요한 offset이 사라질 수 있습니다. Compacted topic에서는 모든 과거 event가 남아 있지 않고 key별 최신 상태 중심으로 남을 수 있습니다. Kafka Streams state store changelog, compacted metadata topic, 일반 event topic은 읽는 의미가 다릅니다. Interview에서는 "Kafka는 log라서 영원히 재생할 수 있다"라고 답하지 말고, retention과 compaction policy 안에서 재생 가능하다고 말해야 합니다.
