# 02. Distributed System Foundations

## 목차

- [1. Partial Failure는 왜 제일 먼저 배워야 하는가](#1-partial-failure는-왜-제일-먼저-배워야-하는가)
- [2. 시간과 순서는 wall clock만으로 닫히지 않는다](#2-시간과-순서는-wall-clock만으로-닫히지-않는다)
- [3. Log는 상태를 다시 만들기 위한 기억 장치다](#3-log는-상태를-다시-만들기-위한-기억-장치다)
- [4. Partitioning은 성능 단위이면서 correctness 경계다](#4-partitioning은-성능-단위이면서-correctness-경계다)
- [5. Replication, Quorum, Consensus는 같은 말이 아니다](#5-replication-quorum-consensus는-같은-말이-아니다)
- [6. Consistency는 "강함/약함"이 아니라 read/write 계약이다](#6-consistency는-강함약함이-아니라-readwrite-계약이다)
- [7. Recovery는 실패 후 다시 시작할 기준점을 남기는 일이다](#7-recovery는-실패-후-다시-시작할-기준점을-남기는-일이다)
- [8. Backpressure는 queue가 길어지는 것을 인정하는 설계다](#8-backpressure는-queue가-길어지는-것을-인정하는-설계다)
- [현실 시나리오 1: timeout이 났지만 write가 되었을 수도 있다](#현실-시나리오-1-timeout이-났지만-write가-되었을-수도-있다)
- [현실 시나리오 2: QUORUM인데 stale read가 의심된다](#현실-시나리오-2-quorum인데-stale-read가-의심된다)
- [Membership과 failure detector는 "누가 살아 있는가"를 추정하는 장치다](#membership과-failure-detector는-누가-살아-있는가를-추정하는-장치다)
- [Leader, follower, epoch는 순서를 고정하기 위한 이름표다](#leader-follower-epoch는-순서를-고정하기-위한-이름표다)
- [Quorum과 consensus를 더 선명하게 가르기](#quorum과-consensus를-더-선명하게-가르기)
- [Recovery는 replay, snapshot, checkpoint를 섞어 시간을 줄인다](#recovery는-replay-snapshot-checkpoint를-섞어-시간을-줄인다)
- [Idempotency, deduplication, outbox는 timeout의 불확실성을 다루는 도구다](#idempotency-deduplication-outbox는-timeout의-불확실성을-다루는-도구다)
- [Overload와 backpressure는 correctness 문제이기도 하다](#overload와-backpressure는-correctness-문제이기도-하다)
- [Failure injection과 Jepsen식 사고는 반례를 실험으로 바꾼다](#failure-injection과-jepsen식-사고는-반례를-실험으로-바꾼다)
- [Interview replay: 분산 시스템 질문의 첫 답변](#interview-replay-분산-시스템-질문의-첫-답변)
- [Consistency model 이름은 관측 가능한 약속으로 번역해야 한다](#consistency-model-이름은-관측-가능한-약속으로-번역해야-한다)
- [Two-phase commit과 saga, outbox는 외부 side effect의 경계를 드러낸다](#two-phase-commit과-saga-outbox는-외부-side-effect의-경계를-드러낸다)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [근거와 더 읽을 자료](#근거와-더-읽을-자료)

분산 시스템은 "서버가 여러 대인 시스템"이라는 말로는 부족합니다. 핵심은 여러 머신이 같은 메모리, 같은 시계, 같은 실패 신호를 공유하지 않는다는 점입니다. 한 머신 안에서는 커널이 많은 전역 상태를 알고 있지만, 여러 머신 사이에서는 message가 늦게 도착할 수 있고, node가 멈췄는지 network가 막혔는지 구분하기 어렵고, 어느 write가 먼저였는지 모든 참여자가 동시에 알 수 없습니다.

그래서 분산 시스템의 질문은 대개 다음 모양으로 바뀝니다.

```text
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

```text
client sends write W
  -> network delay
  -> server applies W
  -> response delayed or lost
  -> client timeout
  -> client does not know whether W happened
```

이 흐름에서 retry가 안전하려면 요청이 idempotent해야 합니다. idempotent는 같은 요청을 여러 번 적용해도 결과가 한 번 적용한 것과 같다는 뜻입니다. 결제 승인처럼 중복 적용이 치명적인 작업은 request id, deduplication table, transactional outbox 같은 장치가 필요합니다. Kafka producer idempotence, Cassandra timestamp conflict resolution, Spark task retry가 모두 이 문제의 다른 형태입니다.

## 2. 시간과 순서는 wall clock만으로 닫히지 않는다

여러 머신에는 각자의 clock이 있습니다. NTP로 맞추더라도 완전히 같은 순간을 공유한다고 볼 수 없습니다. clock skew가 있고, network delay가 있고, VM pause나 GC pause로 한 process가 한동안 멈출 수 있습니다. 그래서 "timestamp가 더 크니까 반드시 나중 이벤트(event)"라고 단정하면 위험합니다.

Lamport의 happens-before는 이 문제를 다루기 위한 사고방식입니다. 한 process 안에서 앞선 이벤트는 뒤 이벤트보다 먼저입니다. message send는 그 message receive보다 먼저입니다. 이 관계를 따라갈 수 있으면 일부 이벤트의 순서를 말할 수 있습니다. 하지만 관계가 이어지지 않는 두 이벤트는 concurrent일 수 있습니다. 즉 둘 중 무엇이 먼저인지 시스템 관점에서 정하지 못할 수 있습니다.

```text
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

```text
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

```text
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

```text
RF=3 replicas: A, B, C

write W reaches A, B
  -> CL=QUORUM success

read R asks B, C
  -> intersects with B
  -> can observe W if B's value wins and response path is healthy
```

consensus는 quorum보다 더 엄격합니다. Raft 같은 consensus protocol은 leader election, term, log index, majority replication, commit rule을 통해 참여자가 같은 log prefix에 합의하도록 합니다. Kafka의 metadata quorum이나 transactional protocol을 설명할 때는 consensus 계열 사고가 필요하지만, Cassandra의 일반 read/write quorum을 곧바로 consensus라고 부르면 틀립니다. quorum은 응답 개수 정책이고, consensus는 어떤 값과 순서에 대해 안전성을 보장하는 protocol입니다.

state machine replication을 작은 trace로 보면 차이가 더 선명합니다. 목표는 모든 node가 같은 명령을 같은 순서로 적용해 같은 상태에 도달하게 하는 것입니다. leader는 client 명령을 log entry로 만들고, follower에게 같은 index와 term으로 append해 달라고 보냅니다. majority가 받아들이면 leader는 그 entry를 committed로 보고 state machine에 적용합니다. follower는 leader의 commit index를 따라 같은 entry를 같은 순서로 적용합니다.

```text
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

```text
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

## Membership과 failure detector는 "누가 살아 있는가"를 추정하는 장치다

분산 시스템에서 가장 난감한 질문은 node가 죽었는지, network가 막혔는지, 잠깐 느린지 즉시 알 수 없다는 점입니다. 그래서 membership은 cluster에 어떤 node가 참여하고 있는지에 대한 view를 관리하고, failure detector는 어떤 node가 의심스러운지 추정합니다. 여기서 중요한 단어는 추정입니다. Timeout은 node death의 증명이 아니라 "정해진 시간 안에 기대한 신호를 받지 못했다"는 관측입니다.

```text
Node A expects heartbeat from Node B
  -> heartbeat delayed by network
  -> or B paused by GC
  -> or B overloaded by disk I/O
  -> or B actually crashed
  -> A only observes missing heartbeat before timeout
```

Failure detector는 빠르게 의심하면 장애 대응이 빨라지지만 false positive가 늘고, 느리게 의심하면 정상 node를 잘못 제거할 가능성은 줄지만 실제 장애 대응이 늦습니다. Cassandra의 gossip과 failure detection, Kafka broker/controller membership, Spark executor heartbeat는 모두 이 tradeoff를 갖습니다. "heartbeat가 끊겼으니 죽었다"가 아니라, "heartbeat를 받지 못해 membership view에서 의심하거나 제외하는 정책이 작동했다"라고 말해야 정확합니다.

Membership change는 단순 목록 수정이 아닙니다. Kafka에서 broker가 빠지면 partition leadership과 replica placement가 바뀔 수 있습니다. Cassandra에서 node가 들어오거나 나가면 token range ownership과 streaming, repair가 움직입니다. Spark에서 executor가 사라지면 그 executor에서 돌던 task와 shuffle block, cached data를 다시 계산하거나 가져와야 합니다. Node 하나의 생사 판단은 data ownership과 recovery plan으로 이어집니다.

## Leader, follower, epoch는 순서를 고정하기 위한 이름표다

여러 node가 동시에 같은 결정을 내리려 하면 충돌이 생깁니다. 그래서 많은 시스템은 특정 범위에서 leader를 둡니다. Leader는 write를 받아 log 순서를 정하거나 metadata 변경을 조정합니다. Follower는 leader의 log나 상태를 따라가며, leader가 사라졌다고 판단되면 election을 통해 새 leader를 고릅니다. 이때 epoch, term, generation 같은 번호는 "어느 leader 시대의 결정인가"를 구분하는 이름표입니다.

```text
term 7 leader L1
  -> appends entries 10, 11
  -> network partition

term 8 leader L2
  -> must not accept stale decisions from term 7 as current authority
```

Epoch가 없으면 오래된 leader가 뒤늦게 응답했을 때 새 leader의 결정을 덮어쓸 수 있습니다. Kafka의 leader epoch, consumer group generation, Raft의 term은 모두 stale actor를 구분하는 장치입니다. Cassandra의 일반 read/write path는 leader 기반 consensus와 다르지만, gossip state와 schema/version, repair session 같은 곳에서는 "어느 view와 어느 상태를 기준으로 말하는가"가 중요합니다. Spark에서도 stage attempt, task attempt, executor heartbeat, streaming batch id가 비슷한 역할을 합니다.

Leader가 있다고 모든 것이 strong consistency가 되는 것은 아닙니다. Leader가 어떤 범위의 순서만 정하는지 봐야 합니다. Kafka partition leader는 해당 partition log의 append order를 정하지만, 여러 partition 사이의 전역 순서를 기본으로 만들지 않습니다. Spark driver는 job/stage/task scheduling을 조정하지만, 외부 sink side effect의 exactly-once를 자동으로 보장하지 않습니다. Cassandra 일반 write는 leader 하나가 전역 순서를 정하는 모델이 아닙니다.

## Quorum과 consensus를 더 선명하게 가르기

Quorum은 replica 집합 중 몇 개의 응답을 성공으로 볼지 정하는 방식입니다. Consensus는 참여자들이 하나의 값이나 log 순서에 안전하게 합의하는 protocol입니다. 둘 다 majority라는 숫자를 쓸 수 있어 헷갈리지만, 같은 문제가 아닙니다. Quorum read/write는 응답 집합이 겹치게 만들어 최신성을 높이거나 availability와 latency를 조절합니다. Consensus는 leader election, term, log matching, commit rule 같은 규칙으로 서로 다른 node가 같은 결정 순서를 갖게 합니다.

```text
quorum read/write
  write succeeds after W replicas
  read succeeds after R replicas
  if R + W > N, sets intersect
  -> intersection helps observe a write

consensus log
  leader proposes entry at index i, term t
  majority accepts
  commit rule marks entry committed
  state machines apply same prefix
```

Quorum이 consensus보다 약하다는 말도 조심해야 합니다. "약하다"는 비난이 아니라 해결하는 문제가 다르다는 뜻입니다. Cassandra는 availability와 tunable consistency를 위해 일반 read/write path에서 quorum 정책을 사용합니다. LWT나 Accord 같은 더 강한 transaction/consensus 계열 기능은 별도 path입니다. Kafka는 partition replication과 leader/follower log, ISR, high watermark를 통해 partition 단위의 durability와 visibility를 관리하고, metadata quorum은 또 다른 consensus 계층을 가질 수 있습니다. 같은 제품 안에도 여러 consistency mechanism이 공존합니다.

## Recovery는 replay, snapshot, checkpoint를 섞어 시간을 줄인다

장애 뒤에는 "어디서 다시 시작할 것인가"가 필요합니다. Log replay는 가장 이해하기 쉬운 방법입니다. 변경 기록을 순서대로 다시 적용해 상태를 복원합니다. 하지만 log가 너무 길면 복구 시간이 길어집니다. Snapshot은 특정 시점의 상태를 저장해 그 이후 log만 replay하게 합니다. Checkpoint는 계산이나 stream processing에서 "여기까지 처리했다"는 기준점을 남깁니다.

```text
state at snapshot S
  -> load snapshot
  -> replay log entries after S
  -> recover current state

stream processing checkpoint
  -> source offsets
  -> operator state
  -> sink commit/progress metadata
  -> restart from recorded boundary
```

Kafka consumer offset은 consumer가 어디까지 처리했는지 나타냅니다. Kafka broker log는 record 자체의 source of truth입니다. Cassandra commitlog는 memtable에 있던 write를 crash 뒤 되살리는 write-ahead log이고, SSTable은 flush된 immutable state입니다. Spark lineage는 partition을 어떻게 다시 만들 수 있는지 알려 주고, checkpoint는 긴 lineage를 끊거나 streaming state를 복구하는 기준점이 됩니다. 이름은 다르지만 모두 "현재 상태만으로는 복구 경로가 부족하므로, 다시 시작할 발판을 남긴다"는 같은 이유에서 나옵니다.

복구 장치는 항상 비용을 갖습니다. Log를 자주 sync하면 runtime latency가 늘 수 있습니다. Snapshot은 저장 비용과 생성 pause를 만들 수 있습니다. Checkpoint는 외부 storage I/O와 consistency 문제를 만듭니다. 너무 드문 checkpoint는 장애 뒤 replay 시간이 길고, 너무 잦은 checkpoint는 평상시 throughput을 낮춥니다. 그래서 복구 설계는 정상 경로 비용과 장애 시 복구 시간의 tradeoff입니다.

## Idempotency, deduplication, outbox는 timeout의 불확실성을 다루는 도구다

분산 시스템에서 retry는 필요하지만 위험합니다. Timeout 뒤 retry를 하면 원래 요청이 적용되었는지 모르는 상태에서 같은 일을 다시 시도하는 것입니다. Idempotency는 같은 요청을 여러 번 적용해도 결과가 한 번 적용한 것과 같게 만드는 성질입니다. Deduplication은 request id나 operation id를 기록해 중복 처리를 막는 방식입니다. Outbox는 local transaction 안에 외부로 보낼 event를 함께 기록한 뒤 별도 publisher가 안전하게 내보내는 pattern입니다.

```text
client sends CreateOrder(requestId=R)
  -> server writes order but response lost
  -> client retries R
  -> server sees R already applied
  -> returns previous result or no-op
```

Kafka idempotent producer는 producer retry로 같은 record가 중복 append되는 문제를 줄입니다. Transactional producer는 Kafka read-process-write와 offset commit을 묶을 수 있습니다. 하지만 외부 DB, 결제 API, 이메일 발송까지 자동으로 포함하지는 않습니다. Cassandra write는 primary key와 timestamp semantics에 따라 중복 write가 같은 값으로 collapse될 수 있지만, counter나 non-idempotent side effect는 조심해야 합니다. Spark task retry는 같은 output을 두 번 쓸 수 있으므로 sink가 idempotent하거나 commit protocol이 안전해야 합니다.

Exactly-once라는 표현은 항상 범위를 붙여 말해야 합니다. 어느 source에서 어느 state를 거쳐 어느 sink까지, 어떤 failure와 retry 안에서 정확히 한 번처럼 보이는가가 중요합니다. 범위 밖 side effect는 별도 idempotency나 transaction boundary가 필요합니다.

## Overload와 backpressure는 correctness 문제이기도 하다

과부하는 단순 성능 저하가 아닙니다. 큐(queue)가 커지면 timeout이 늘고, timeout은 retry를 만들고, retry는 더 많은 load를 만듭니다. 이 feedback loop가 닫히지 않으면 시스템은 실제 처리보다 재시도와 큐 관리에 더 많은 자원을 씁니다. Backpressure는 producer가 consumer의 처리 능력에 맞게 속도를 줄이게 하는 장치입니다. 진입 제어(admission control)는 애초에 처리할 수 없는 요청을 일찍 거절해 내부 queue를 보호합니다.

```text
downstream slows
  -> upstream queue grows
  -> timeout increases
  -> clients retry
  -> load grows further
  -> useful throughput drops

with backpressure
  -> queue limit or rate limit triggers
  -> caller slows or gets early error
  -> internal latency stays bounded
```

Kafka에서는 producer quota, broker request queue, consumer lag, replica lag가 backpressure 신호입니다. Cassandra에서는 pending compaction, dropped mutations, coordinator timeout, hinted handoff backlog가 신호입니다. Spark에서는 streaming source rate, shuffle fetch wait, executor backlog, task retry가 신호입니다. 좋은 설계는 timeout을 마지막 관측으로 두지 않고, queue가 커지는 초기에 속도를 조절합니다.

## Failure injection과 Jepsen식 사고는 반례를 실험으로 바꾼다

분산 시스템 설명은 정상 path만으로는 부족합니다. Network partition, clock skew, process pause, disk full, packet loss, duplicate request, delayed message, node restart를 넣었을 때 어떤 보장이 깨지는지 봐야 합니다. Jepsen식 테스트는 시스템에 장애를 주입하고, history를 수집한 뒤 consistency model과 실제 결과가 맞는지 분석합니다. 모든 프로젝트가 Jepsen을 직접 돌릴 수는 없지만, 사고방식은 가져올 수 있습니다.

```text
claim: read-after-write is guaranteed for this operation
  -> inject timeout after write reaches one replica
  -> retry and read from another replica
  -> inspect observed histories
  -> decide whether claim still holds
```

작은 팀에서는 더 간단한 실험도 가능합니다. Kafka broker 하나를 늦추거나 죽이고 producer ack와 consumer lag를 봅니다. Cassandra node를 stop하고 CL별 read/write 결과와 hinted handoff/repair를 봅니다. Spark executor를 kill하고 task retry, lineage recomputation, checkpoint recovery를 봅니다. 중요한 것은 "장애를 넣었다"가 아니라, 넣기 전에 어떤 claim을 반증하려는지 정하고, 결과를 consistency/availability/latency 관점에서 읽는 것입니다.

## Interview replay: 분산 시스템 질문의 첫 답변

분산 시스템 면접에서 가장 안전한 첫 문장은 이것입니다. "여러 node는 같은 메모리와 같은 시계와 완벽한 실패 신호를 공유하지 않습니다. 그래서 timeout은 실패의 증명이 아니라 불확실한 관측이고, 시스템은 순서, 복제, membership, recovery, idempotency, backpressure를 명시적으로 설계해야 합니다." 이 문장 뒤에 partial failure, logical time, quorum/consensus, log replay, checkpoint, retry를 연결하면 됩니다.

제품 질문으로 오면 Kafka는 partition log와 offset, ISR/high watermark, consumer group generation으로 답합니다. Cassandra는 RF/CL, gossip/failure detector, commitlog/SSTable/repair로 답합니다. Spark는 driver/executor heartbeat, task retry, lineage/checkpoint, shuffle materialization으로 답합니다. 세 시스템은 다르지만, 모두 "어디까지 적용되었고, 누가 보았고, 어디서 다시 시작할 수 있는가"라는 같은 질문을 해결합니다.

## Consistency model 이름은 관측 가능한 약속으로 번역해야 한다

Linearizability, serializability, eventual consistency, causal consistency 같은 이름은 면접에서 자주 나오지만, 이름만 외우면 위험합니다. Linearizability는 각 operation이 실제 시간 순서와 양립하는 한 지점에서 즉시 일어난 것처럼 보이는 강한 단일 객체 약속입니다. Serializability는 transaction들이 어떤 순차 실행과 같은 결과를 내면 된다는 약속이며, 실제 시간 순서까지 항상 보존하는 것은 아닙니다. Eventual consistency는 update가 더 이상 들어오지 않고 통신이 회복되면 replica들이 결국 수렴한다는 약한 방향의 약속입니다.

```text
client A writes x=1 and gets success
client B later reads x

linearizable expectation
  -> B must see x=1 or later value

eventual expectation
  -> B may see old value for a while
  -> convergence expected after propagation/repair
```

Kafka partition log는 partition 안에서 offset order를 강하게 다루지만, 여러 partition 사이 전역 linearizability를 제공한다고 말하면 틀립니다. Cassandra ordinary quorum read/write는 tunable consistency를 제공하지만 모든 operation을 linearizable transaction으로 만들지는 않습니다. Spark job은 deterministic transformation이라면 lineage로 같은 결과를 재계산할 수 있지만, 외부 sink side effect의 serializability는 별도 문제입니다. Consistency model은 제품 이름이 아니라 operation 범위와 관측 가능한 read/write history로 말해야 합니다.

## Two-phase commit과 saga, outbox는 외부 side effect의 경계를 드러낸다

여러 시스템에 걸친 transaction은 어렵습니다. Two-phase commit은 coordinator가 prepare와 commit 단계를 통해 참여자들이 함께 commit/abort하도록 조정하지만, coordinator failure와 blocking, participant capability 문제가 있습니다. Saga는 긴 business transaction을 여러 local transaction과 보상 action으로 나눕니다. Outbox는 local DB transaction 안에 event 발행 의도를 기록해, 이후 publisher가 message broker로 안전하게 전달하게 합니다.

이 패턴들은 Kafka transaction이나 Cassandra write, Spark sink와 자주 만납니다. Kafka topic에는 write했지만 DB update가 실패하거나, DB에는 썼지만 Kafka event 발행 전 process가 죽는 상황을 생각해야 합니다. Spark streaming sink가 batch를 두 번 실행할 수 있으면 sink idempotency가 필요합니다. 분산 시스템에서 "정확히 한 번"은 항상 어느 boundary 안의 이야기인지 먼저 밝혀야 합니다.

## 문서를 덮고 확인할 것

- timeout이 실패 증명이 아닌 이유를 request trace로 설명해 보세요.
- happens-before로 순서를 말할 수 있는 이벤트와 말할 수 없는 이벤트를 하나씩 만들어 보세요.
- replication, quorum, consensus를 각각 한 문장으로 구분해 보세요.
- CAP를 평상시 분류표로 쓰면 왜 위험한지 설명해 보세요.
- Kafka offset, Cassandra commit log, Spark checkpoint가 모두 "다시 시작할 기준점"이라는 점을 비교해 보세요.

## 근거와 더 읽을 자료

- Lamport, "Time, Clocks, and the Ordering of Events in a Distributed System".
- Ongaro and Ousterhout, "In Search of an Understandable Consensus Algorithm".
- Gilbert and Lynch, "Brewer's conjecture and the feasibility of consistent, available, partition-tolerant web services".
- Abadi, "Consistency Tradeoffs in Modern Distributed Database System Design".
- Dynamo, Bigtable, Kafka, Cassandra, Spark primary papers and official docs listed in [10_source_ledger.md](10_source_ledger.md).
