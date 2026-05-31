# 02. Distributed System Foundations

분산 시스템은 "서버가 여러 대인 시스템"이라는 말로는 부족합니다. 핵심은 여러 머신이 같은 메모리, 같은 시계, 같은 실패 신호를 공유하지 않는다는 점입니다. 한 머신 안에서는 커널이 많은 전역 상태를 알고 있지만, 여러 머신 사이에서는 message가 늦게 도착할 수 있고, node가 멈췄는지 network가 막혔는지 구분하기 어렵고, 어느 write가 먼저였는지 모든 참여자가 동시에 알 수 없습니다.

그래서 분산 시스템의 질문은 대개 다음 모양으로 바뀝니다.

```
state changed somewhere
  -> who observed it?
  -> in what order?
  -> how many replicas copied it?
  -> what can a read safely assume?
  -> after failure, which record is the restart point?
```

Kafka, Cassandra, Spark는 모두 이 질문을 다룹니다. Kafka는 partition log와 offset으로 순서와 재생을 다룹니다. Cassandra는 replica와 consistency level로 availability와 recency를 조절합니다. Spark는 lineage와 checkpoint로 실패한 계산 조각을 되살립니다.

## 1. Partial Failure는 왜 제일 먼저 배워야 하는가

단일 process 안에서는 함수 호출이 실패하면 대개 예외나 반환값으로 실패를 압니다. 분산 시스템에서는 요청을 보냈는데 응답이 오지 않는 상황이 흔합니다. 이때 응답 없음은 세 가지와 모두 양립할 수 있습니다.

- 상대 node가 죽었다.
- 상대 node는 살아 있지만 network가 막혔다.
- 상대 node가 처리하고 응답을 보냈지만 응답 packet이 늦거나 사라졌다.

이 차이를 즉시 알 수 없기 때문에 timeout은 실패의 증명이 아니라 실패 가능성에 대한 관측입니다. timeout을 너무 짧게 잡으면 정상 요청을 실패로 오판하고 retry storm을 만들 수 있습니다. 너무 길게 잡으면 실제 장애를 늦게 감지해 queue가 쌓입니다.

```
client sends write W
  -> network delay
  -> server applies W
  -> response delayed or lost
  -> client timeout
  -> client does not know whether W happened
```

이 흐름에서 retry가 안전하려면 요청이 idempotent해야 합니다. idempotent는 같은 요청을 여러 번 적용해도 결과가 한 번 적용한 것과 같다는 뜻입니다. 결제 승인처럼 중복 적용이 치명적인 작업은 request id, deduplication table, transactional outbox 같은 장치가 필요합니다. Kafka producer idempotence, Cassandra timestamp conflict resolution, Spark task retry가 모두 이 문제의 다른 형태입니다.

## 2. 시간과 순서는 wall clock만으로 닫히지 않는다

여러 머신에는 각자의 clock이 있습니다. NTP로 맞추더라도 완전히 같은 순간을 공유한다고 볼 수 없습니다. clock skew가 있고, network delay가 있고, VM pause나 GC pause로 한 process가 한동안 멈출 수 있습니다. 그래서 "timestamp가 더 크니까 반드시 나중 사건"이라고 단정하면 위험합니다.

Lamport의 happens-before는 이 문제를 다루기 위한 사고방식입니다. 한 process 안에서 앞선 사건은 뒤 사건보다 먼저입니다. message send는 그 message receive보다 먼저입니다. 이 관계를 따라갈 수 있으면 일부 사건의 순서를 말할 수 있습니다. 하지만 관계가 이어지지 않는 두 사건은 concurrent일 수 있습니다. 즉 둘 중 무엇이 먼저인지 시스템 관점에서 정하지 못할 수 있습니다.

```
Node A: write x=1 at local time 10
        send message m -------------------->

Node B: local time 8                 receive m
                                      read x?

happens-before:
  A write before A send
  A send before B receive
  therefore A write before B receive

not guaranteed:
  local time 10 on A vs local time 8 on B as absolute order
```

Kafka는 partition 안에서 append order를 정합니다. Cassandra는 cell timestamp와 conflict resolution을 쓰지만 clock skew와 last-write-wins의 위험을 이해해야 합니다. Spark는 dataset partition 사이의 전역 record 순서를 기본으로 보장하지 않습니다. 각 시스템은 "모든 것의 전역 순서" 대신 필요한 범위의 순서만 만듭니다.

## 3. Log는 상태를 다시 만들기 위한 기억 장치다

분산 시스템이 log를 좋아하는 이유는 단순합니다. 현재 상태만 있으면 장애 후 왜 그 상태가 되었는지 복원하기 어렵습니다. 반대로 상태 변경 명령이 순서대로 기록되어 있으면, 어느 지점까지 적용했는지 알고 다시 재생할 수 있습니다.

```
commands:
  1. put a=1
  2. put b=2
  3. delete a

state after replay:
  b=2
```

Kafka의 log는 사용자가 읽고 재생하는 중심 데이터 구조입니다. consumer offset은 "어느 log 위치까지 처리했는가"를 나타냅니다. Cassandra의 commit log는 사용자가 직접 읽는 데이터 모델은 아니지만 crash recovery를 위한 write-ahead log입니다. Spark의 lineage는 byte log는 아니지만 "이 partition은 어떤 transformation으로 만들어졌는가"라는 재계산 기록입니다. 셋 모두 장애 후 어디서 다시 시작할지 남긴다는 공통점이 있습니다.

log가 있다고 모든 문제가 사라지지는 않습니다. log를 어느 replica까지 복제했는지, log entry가 commit되었다고 볼 조건은 무엇인지, 오래된 log를 언제 지울지, replay가 idempotent한지 같은 질문이 따라옵니다.

## 4. Partitioning은 성능 단위이면서 correctness 경계다

데이터나 일을 여러 조각으로 나누면 parallelism이 생깁니다. 하지만 나누는 순간 "같은 조각 안에서만 쉽게 보장되는 것"과 "조각 사이에서는 비싸거나 보장되지 않는 것"이 갈라집니다.

Kafka에서 partition은 ordering의 경계입니다. 같은 partition 안에서는 offset 순서가 있지만, topic 전체의 여러 partition 사이에는 기본 전역 순서가 없습니다. Cassandra에서 partition key는 어느 token range와 replica가 데이터를 맡을지 정합니다. partition을 잘못 고르면 hot partition이 생기고 특정 node가 과부하됩니다. Spark에서 partition은 task의 입력 단위입니다. partition 수가 너무 적으면 병렬성이 부족하고, 너무 많으면 scheduler overhead와 shuffle metadata가 커집니다.

```
same key / same partition
  -> local order or ownership can be reasoned about

different partitions
  -> coordination is required for global order or transaction
  -> coordination adds latency and failure cases
```

분산 시스템에서 scale-out은 단순히 node를 늘리는 일이 아닙니다. 어떤 key가 어디로 가고, rebalancing 때 어떤 데이터가 이동하며, 이동 중 read/write가 어떤 값을 보는지까지 설계해야 합니다.

## 5. Replication, Quorum, Consensus는 같은 말이 아니다

replication은 같은 상태를 여러 곳에 복사하는 일입니다. quorum은 그 복사본 중 몇 개의 응답을 성공 조건으로 볼지 정하는 방식입니다. consensus는 여러 node가 하나의 값이나 log 순서에 합의하는 더 강한 문제입니다. 세 단어를 섞으면 Cassandra와 Kafka, Raft를 모두 잘못 설명하게 됩니다.

replication factor 3인 저장소에서 write consistency level이 2이고 read consistency level이 2라면, 정상적인 조건에서는 read와 write의 replica 집합이 최소 하나 겹칩니다. 이 교차 때문에 최신 write를 볼 가능성이 커집니다. 하지만 "항상 최신"이라고 단정하면 안 됩니다. clock conflict, failed write, hinted handoff, read repair 상태, lightweight transaction 사용 여부, network partition이 모두 영향을 줍니다.

```
RF=3 replicas: A, B, C

write W reaches A, B
  -> CL=QUORUM success

read R asks B, C
  -> intersects with B
  -> can observe W if B's value wins and response path is healthy
```

consensus는 quorum보다 더 엄격합니다. Raft 같은 consensus protocol은 leader election, term, log index, majority replication, commit rule을 통해 참여자가 같은 log prefix에 합의하도록 합니다. Kafka의 metadata quorum이나 transactional protocol을 설명할 때는 consensus 계열 사고가 필요하지만, Cassandra의 일반 read/write quorum을 곧바로 consensus라고 부르면 틀립니다. quorum은 응답 개수 정책이고, consensus는 어떤 값과 순서에 대해 안전성을 보장하는 protocol입니다.

state machine replication을 작은 trace로 보면 차이가 더 선명합니다. 목표는 모든 node가 같은 명령을 같은 순서로 적용해 같은 상태에 도달하게 하는 것입니다. leader는 client 명령을 log entry로 만들고, follower에게 같은 index와 term으로 append해 달라고 보냅니다. majority가 받아들이면 leader는 그 entry를 committed로 보고 state machine에 적용합니다. follower는 leader의 commit index를 따라 같은 entry를 같은 순서로 적용합니다.

```
client command: set x=1

leader term 7:
  log[42] = (term 7, set x=1)
  append log[42] to followers A, B, C

followers:
  A stores log[42]
  B stores log[42]
  C is slow

majority reached: leader + A + B
  -> entry 42 is committed
  -> leader applies set x=1 to state machine
  -> followers learn commit index and apply in order
```

leader가 entry를 자기 log에만 적고 죽으면 그 entry는 committed가 아닐 수 있습니다. 새 leader는 majority가 안전하게 가진 log prefix를 기준으로 이어 갑니다. 이 trace가 quorum read/write와 다른 점은 "응답을 몇 개 받았는가"에서 끝나지 않고, term과 log index, leader election, commit rule을 통해 log 순서의 안전성을 지킨다는 데 있습니다.

Kafka의 KRaft metadata quorum처럼 cluster metadata의 순서를 안전하게 정해야 하는 영역은 이 사고방식과 가깝습니다. 반대로 Kafka data partition의 producer append/replica fetch나 Cassandra의 일반 CL read/write는 각각의 protocol 약속을 따르며, 모든 것을 Raft로 설명하면 안 됩니다.

## 6. Consistency는 "강함/약함"이 아니라 read/write 계약이다

"이 시스템은 consistent한가요?"라는 질문은 불완전합니다. 더 정확한 질문은 "어떤 read가 어떤 write를 반드시 보아야 하는가?"입니다. linearizability는 한 객체에 대한 작업이 실제 시간 순서를 지키는 하나의 원자적 순서로 보이는 성질입니다. sequential consistency는 실제 시간 순서까지는 요구하지 않지만 모든 process가 같은 순서를 본다는 성질입니다. eventual consistency는 update가 더 이상 없고 통신이 회복되면 replica들이 결국 수렴한다는 성질입니다.

CAP도 이 맥락에서 읽어야 합니다. CAP는 평상시 모든 시스템을 세 글자로 분류하는 도구가 아닙니다. network partition이 발생했을 때, 시스템이 더 이상 통신할 수 없는 replica 사이에서 consistency와 availability를 동시에 완전히 만족할 수 없다는 한계를 말합니다. PACELC는 partition이 없을 때도 latency와 consistency 사이의 tradeoff가 있다는 점을 강조합니다.

Kafka는 partition leader를 중심으로 ordering과 committed offset을 다룹니다. Cassandra는 tunable consistency로 read/write 응답 조건을 고르게 합니다. Spark는 저장소 consistency보다 계산 lineage와 checkpoint의 재실행 가능성이 더 중요합니다. 각 시스템이 어떤 consistency를 약속하는지 보려면 문서의 기능 이름보다 read/write의 관측 계약을 따라가야 합니다.

## 7. Recovery는 실패 후 다시 시작할 기준점을 남기는 일이다

장애 처리에서 가장 중요한 질문은 "다시 시작할 때 무엇을 믿을 것인가"입니다. log, checkpoint, offset, commit marker, SSTable, lineage는 모두 이 질문에 답하기 위한 기준점입니다.

Kafka consumer가 죽었다가 다시 시작하면 committed offset부터 다시 읽습니다. offset을 너무 일찍 commit하면 처리하지 않은 record를 건너뛸 수 있고, 너무 늦게 commit하면 중복 처리할 수 있습니다. Cassandra node가 crash하면 commit log를 replay해 memtable에 있던 acknowledged mutation을 복원합니다. Spark task가 실패하면 lineage를 따라 partition을 다시 계산하고, lineage가 너무 길거나 외부 side effect가 있으면 checkpoint가 필요해집니다.

```
failure happens
  -> find durable or reconstructable boundary
  -> discard uncertain in-memory state
  -> replay log / recompute lineage / reload checkpoint
  -> resume from recorded offset or commit point
```

회복은 공짜가 아닙니다. log replay가 길면 재시작이 느려지고, checkpoint가 잦으면 정상 처리 비용이 늘어납니다. replication을 늘리면 read/write latency와 storage cost가 늘 수 있습니다. 좋은 설계는 장애를 없애는 것이 아니라, 장애 후 어느 상태까지 잃을 수 있고 어느 지점부터 다시 시작할지 명확히 합니다.

## 8. Backpressure는 queue가 길어지는 것을 인정하는 설계다

분산 시스템은 producer와 consumer 속도가 항상 같지 않습니다. producer가 더 빠르면 queue가 생깁니다. queue가 작으면 빠르게 실패하고, queue가 크면 실패를 늦게 보지만 memory와 latency가 커집니다. backpressure는 downstream이 느릴 때 upstream의 속도를 줄이거나, 명시적으로 실패하거나, load shedding을 하는 구조입니다.

Kafka에서는 broker disk/network, consumer processing, downstream DB 중 하나가 느리면 consumer lag가 증가합니다. Cassandra에서는 memtable flush나 compaction이 밀리면 write latency가 올라가거나 write가 block될 수 있습니다. Spark에서는 shuffle read가 느리거나 skew partition이 있으면 일부 task가 전체 job을 붙잡습니다. 현상은 다르지만 질문은 같습니다. 어느 queue가 자라고, 그 queue의 producer와 consumer는 누구인가?

## 현실 시나리오 1: timeout이 났지만 write가 되었을 수도 있다

사용자가 주문 생성 API를 호출했고, 서비스가 Cassandra에 write를 보낸 뒤 timeout을 받았다고 합시다. client 입장에서는 실패처럼 보입니다. 하지만 replica 일부 또는 전부에는 write가 이미 들어갔을 수 있습니다. 이때 같은 주문 생성 요청을 아무 보호 없이 retry하면 중복 주문이 생길 수 있습니다.

해결은 "timeout이면 rollback"이 아닙니다. 요청 id를 두고 idempotent하게 처리하거나, write/read consistency를 조절하거나, 상태 조회 API로 이미 처리된 요청인지 확인해야 합니다. 면접에서는 timeout을 실패 증명으로 말하지 말고, "응답 없음은 적용 여부 불명확 상태이므로 idempotency와 관측 경로가 필요하다"고 답하는 편이 안전합니다.

## 현실 시나리오 2: QUORUM인데 stale read가 의심된다

RF=3, write CL=QUORUM, read CL=QUORUM이면 read/write 집합이 겹칩니다. 그래서 많은 경우 최신 값을 볼 가능성이 높습니다. 그러나 replica가 서로 다른 timestamp의 값을 갖고 있고 clock skew가 있었거나, 실패한 write와 repair 상태가 섞였거나, LWT가 아닌 일반 write로 경쟁 update가 들어오면 "QUORUM이면 무조건 최신"이라고 말할 수 없습니다.

여기서 해야 할 일은 quorum이라는 단어를 반복하는 것이 아니라 write가 어느 replica에 성공했는지, read가 어느 replica에서 어떤 version을 받았는지, conflict resolution과 repair가 어떻게 동작하는지 확인하는 것입니다.

## 문서를 덮고 확인할 것

- timeout이 실패 증명이 아닌 이유를 request trace로 설명해 보세요.
- happens-before로 순서를 말할 수 있는 사건과 말할 수 없는 사건을 하나씩 만들어 보세요.
- replication, quorum, consensus를 각각 한 문장으로 구분해 보세요.
- CAP를 평상시 분류표로 쓰면 왜 위험한지 설명해 보세요.
- Kafka offset, Cassandra commit log, Spark checkpoint가 모두 "다시 시작할 기준점"이라는 점을 비교해 보세요.

## 근거와 더 읽을 자료

- Lamport, "Time, Clocks, and the Ordering of Events in a Distributed System".
- Ongaro and Ousterhout, "In Search of an Understandable Consensus Algorithm".
- Gilbert and Lynch, "Brewer's conjecture and the feasibility of consistent, available, partition-tolerant web services".
- Abadi, "Consistency Tradeoffs in Modern Distributed Database System Design".
- Dynamo, Bigtable, Kafka, Cassandra, Spark primary papers and official docs listed in [10_source_ledger.md](10_source_ledger.md).
