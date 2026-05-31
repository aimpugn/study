# 02. Distributed System Foundations

> 분산 시스템은 여러 머신이 하나의 상태를 다루는 것처럼 보이게 만들지만, 실제로는 메시지 지연, 노드 정지, clock 차이, packet loss를 완전히 제거할 수 없습니다.
> 그래서 핵심 질문은 "누가 최신인가"보다 "어떤 상태 변화가 어떤 순서로 기록되었고, 몇 replica가 그것을 보았으며, 실패 후 어디서 다시 시작할 수 있는가"입니다.
> Kafka, Cassandra, Spark의 내부 구조는 모두 log, partition, replica, quorum, checkpoint, backpressure 같은 공통 원리를 자기 목적에 맞게 변형한 결과입니다.

## 1. Partial Failure

> 부분 실패(partial failure)는 시스템 전체가 죽지 않았는데 일부 노드, 링크, disk, process, thread만 느려지거나 응답하지 않는 상태입니다.
> 분산 시스템이 어려운 이유는 crash, network delay, GC pause, overloaded disk가 관측상 모두 timeout으로 보일 수 있기 때문입니다.
> 좋은 설계는 실패를 없애지 못하므로 timeout, retry, idempotency, replica, recovery log로 실패를 다룹니다.

### 질문

서버 A가 서버 B에게 요청했는데 응답이 없으면, B가 죽었다고 확신할 수 있을까요?

### 직관

멀리 있는 사람이 전화를 받지 않는 이유는 많습니다. 사람이 쓰러졌을 수도 있고, 터널에 들어갔을 수도 있고, 내 전화기가 멈췄을 수도 있습니다. 분산 시스템의 timeout도 "모른다"는 관측이지 "죽었다"는 증명이 아닙니다.

### 작은 예시

3개 노드 `N1, N2, N3`가 있습니다. `N1`이 coordinator이고 `N2`로 write를 보냈습니다. 200ms 안에 응답이 없습니다.

가능한 사실:

- `N2` process crash
- `N2` JVM GC pause
- `N2` disk queue saturation
- `N1 -> N2` network packet loss
- `N1` scheduler가 response 처리 thread를 늦게 실행

### 상태 이동 trace

```
t0  N1 sends write(x=1) to N2
t1  packet is delayed OR N2 is paused OR N2 disk is slow
t2  N1 timeout fires
t3  N1 marks request failed, maybe retries to N3

Observation at N1: "no response before timeout"
Hidden truth: one of several partial failures
```

### 내부 메커니즘

분산 시스템은 보통 failure detector를 씁니다. heartbeat, TCP connection, request timeout, gossip, lease 같은 방식으로 상대 상태를 추정합니다. 하지만 비동기 네트워크에서는 지연의 상한을 항상 알 수 없으므로 완벽한 failure detector는 어렵습니다. 실제 시스템은 오탐과 미탐 사이에서 timeout 값을 선택합니다.

Kafka는 broker/controller session, replica lag, ISR 변화를 통해 실패를 반영합니다. Cassandra는 gossip과 failure detector, hinted handoff, repair로 노드 상태와 불일치를 다룹니다. Spark는 executor lost, task retry, lineage recomputation으로 일부 실패를 흡수합니다.

### 실패 모드

- timeout이 너무 짧으면 느린 노드를 죽은 것으로 오탐합니다.
- timeout이 너무 길면 실제 장애 감지가 늦어집니다.
- retry가 idempotent하지 않으면 중복 write가 생깁니다.
- 한 노드의 disk saturation이 cluster-wide latency로 번질 수 있습니다.

### 검증 방법

로컬에서 완전한 network partition을 안전하게 재현하기는 어렵습니다. 대신 timeout과 sleep으로 부분 실패를 모사합니다.

```bash
python3 - <<'PY'
import time
start = time.time()
try:
    time.sleep(2)
finally:
    print("observed delay:", round(time.time() - start, 2))
PY
```

PASS 신호:

- caller 관점에서는 "응답 지연"만 보입니다.
- 원인이 sleep인지 crash인지 network인지 추가 관측 없이는 모릅니다.

FAIL 신호:

- 단일 로그만 보고 crash로 단정합니다.

### 면접식 되묻기

"분산 시스템에서 실패 처리가 어려운 이유는?"이라는 질문에는 "부분 실패와 불완전한 관측 때문"이라고 답합니다. 꼬리 질문이 오면 timeout은 추정이며, retry는 idempotency와 결합되어야 하고, replica와 log는 실패 후 상태를 재구성하기 위한 장치라고 설명합니다.

### 흔한 오해와 반례

오해: "TCP connection이 끊기지 않았으면 노드는 살아 있다."

반례: process가 GC pause 중이거나 application thread가 deadlock에 빠져도 TCP connection은 한동안 살아 있을 수 있습니다.

### Active recall

- crash, slow, partition을 timeout만으로 구분하기 어려운 이유를 말해 보세요.
- Kafka ISR 축소와 Cassandra hinted handoff는 partial failure에 각각 어떻게 반응하나요?

## 2. Time, Clocks, Ordering

> 분산 시스템에서 wall-clock 시간은 편리하지만, 여러 노드의 사건 순서를 완벽히 증명하지 못합니다.
> 논리적 시간(logical time)은 "무엇이 무엇보다 먼저 영향을 주었는가"를 추적하기 위한 도구입니다.
> log offset, timestamp, vector clock, checkpoint는 모두 순서를 표현하지만, 서로 보장하는 의미가 다릅니다.

### 질문

두 서버 로그에 찍힌 timestamp가 `10:00:01`과 `10:00:02`이면 첫 번째 사건이 반드시 먼저 일어났다고 말할 수 있을까요?

### 직관

시계가 조금씩 다른 사람 둘이 각자 시간을 적으면 종이의 숫자만으로 실제 순서를 확정하기 어렵습니다. 반면 "A가 보낸 메시지를 B가 받은 뒤 처리했다"는 관계는 시계보다 강한 순서 증거입니다.

### 작은 예시

```
P1: write order=1 at local clock 10:00:02
P2: read order=1  at local clock 10:00:01
```

P2의 시간이 더 작아도, P2가 P1의 메시지를 받은 뒤 읽었다면 causality는 `P1 write -> P2 read`입니다.

### 상태 이동 trace

```
P1 event a: create message M, Lamport=5
P1 sends M(clock=5) --------->
                             P2 receives M
                             P2 clock=max(P2_clock, 5)+1 = 6
                             P2 event b: apply M, Lamport=6

happens-before: a -> b
```

### 내부 메커니즘

Lamport clock은 각 process가 counter를 증가시키고 message에 counter를 실어 보냅니다. 수신자는 자기 counter와 message counter의 max에 1을 더합니다. 이 방식은 `a happens-before b`이면 `clock(a) < clock(b)`가 되도록 만듭니다. 그러나 clock 숫자가 작다고 항상 causality가 있다는 뜻은 아닙니다. 동시 사건도 임의 tie-break로 total order처럼 정렬할 수 있지만, 그것은 실제 원인 관계가 아니라 결정적 처리 순서를 만들기 위한 장치입니다.

Kafka offset은 한 partition 안의 log 순서입니다. Cassandra timestamp는 LWW 충돌 해결에 사용되지만 clock skew에 민감할 수 있습니다. Spark stage/task 순서는 DAG dependency와 scheduler 결정의 조합입니다.

### 실패 모드

- wall-clock만 믿으면 clock skew 때문에 잘못된 conflict resolution이나 로그 분석이 생깁니다.
- Kafka의 partition offset을 전역 순서로 오해하면 multi-partition event ordering을 잘못 설계합니다.
- Cassandra timestamp 기반 LWW는 동시에 가까운 write에서 예상과 다른 값이 이길 수 있습니다.

### 검증 방법

작은 논리시계 모사:

```bash
python3 - <<'PY'
p1 = 0
p2 = 0
p1 += 1
msg = p1
p2 = max(p2, msg) + 1
print({"p1_send": msg, "p2_receive": p2})
PY
```

PASS 신호:

- receive timestamp가 send timestamp보다 큽니다.

FAIL 신호:

- 이 숫자를 실제 wall-clock timestamp로 해석합니다. 논리시계는 원인 순서 추적용입니다.

### 면접식 되묻기

"Kafka offset과 timestamp 중 무엇으로 순서를 보장하나요?"에는 partition 안에서는 offset이 log 순서의 기준이고, timestamp는 event-time 처리나 retention/logging에 쓰일 수 있지만 partition log ordering과 같은 보장은 아니라고 답합니다.

### 흔한 오해와 반례

오해: "timestamp가 있으면 순서를 알 수 있다."

반례: clock skew, NTP 조정, leap, process pause, batching 때문에 timestamp는 관측과 해석이 섞입니다. message causality나 log position이 더 강한 순서 근거일 때가 많습니다.

### Active recall

- Lamport clock이 보장하는 것과 보장하지 않는 것을 구분해 보세요.
- Kafka offset, Cassandra timestamp, Spark checkpoint offset은 각각 어떤 순서 의미를 갖나요?

## 3. Logs and State Machines

> log는 상태 변경 명령을 순서대로 남긴 기록이고, state machine은 같은 초기 상태에 같은 명령 순서를 적용하면 같은 결과가 나오는 규칙입니다.
> 복제된 log는 여러 replica가 같은 순서로 같은 명령을 적용하게 만들어 회복과 failover의 기준이 됩니다.
> Kafka의 partition log, Raft의 replicated log, Spark lineage, Cassandra commit log는 모두 "실패 후 어디서 다시 시작할 것인가"에 답합니다.

### 질문

분산 시스템은 왜 현재 상태만 저장하지 않고 log를 남길까요?

### 직관

은행 잔액만 있으면 왜 그 잔액이 됐는지 알 수 없습니다. 거래 내역이 있으면 중간에 서버가 죽어도 마지막으로 반영된 거래 이후부터 다시 적용할 수 있습니다.

### 작은 예시

초기 상태 `balance=0`에 log가 있습니다.

```
1: deposit +100
2: withdraw -30
3: deposit +20
```

순서대로 적용하면 `balance=90`입니다. replica도 같은 순서를 적용하면 같은 상태가 됩니다.

### 상태 이동 trace

```
client command
  |
  v
append log record at index 3
  |
  v
replicate record
  |
  v
commit decision
  |
  v
apply to state machine
  |
  v
snapshot/checkpoint may compact old log
```

### 내부 메커니즘

state machine replication은 "명령 순서"를 합의하고 각 replica가 같은 순서로 적용하는 방식입니다. consensus는 여러 node가 log의 다음 entry를 어떤 값으로 할지 동의하는 문제를 다룹니다. 모든 시스템이 Raft 같은 consensus를 데이터 path에 쓰지는 않습니다. Kafka는 partition leader가 append 순서를 정하고 replica가 leader log를 따라갑니다. Cassandra는 모든 replica가 독립적으로 mutation을 받을 수 있고 timestamp/repair로 수렴합니다. Spark는 명령 log보다 transformation lineage로 partition을 재계산합니다.

### 실패 모드

- log append는 됐지만 commit 전 crash가 나면 적용 여부가 애매합니다.
- snapshot/checkpoint가 없으면 log replay가 너무 오래 걸립니다.
- nondeterministic state machine은 같은 log를 적용해도 replica 상태가 달라질 수 있습니다.

### 검증 방법

간단한 replay:

```bash
python3 - <<'PY'
log = [("+", 100), ("-", 30), ("+", 20)]
state = 0
for op, n in log:
    state = state + n if op == "+" else state - n
print(state)
PY
```

PASS 신호:

- 같은 log를 같은 순서로 replay하면 같은 state가 나옵니다.

FAIL 신호:

- command가 현재 시간, random, 외부 API 응답에 의존하면 deterministic replay가 깨질 수 있습니다.

### 면접식 되묻기

"Kafka log와 database transaction log는 같은가요?"에는 둘 다 ordered record를 남기는 점은 같지만 목적과 소비자가 다르다고 답합니다. Kafka log는 여러 consumer가 독립 offset으로 읽는 distributed log이고, DB transaction log는 주로 recovery와 replication을 위한 내부 기록입니다.

### 흔한 오해와 반례

오해: "log가 있으면 consensus가 있는 것이다."

반례: log는 순서 있는 기록이고, consensus는 여러 노드가 그 순서와 값을 안전하게 정하는 프로토콜입니다. Kafka, Cassandra, Spark는 log를 쓰지만 각자의 합의/복제 모델은 다릅니다.

### Active recall

- log와 state machine이 함께 있을 때 recovery가 쉬워지는 이유를 말해 보세요.
- Kafka partition log와 Spark lineage는 어떤 점에서 비슷하고 어떤 점에서 다른가요?

## 4. Partitioning and Sharding

> partitioning은 큰 상태나 일을 여러 조각으로 나눠 다른 node, core, disk에 배치하는 방법입니다.
> partition은 처리량을 늘리는 단위이지만, 순서 보장, transaction 범위, skew, hot key 같은 correctness와 성능 경계도 됩니다.
> Kafka partition, Cassandra token range, Spark partition은 모두 "어떤 key나 record가 어디에서 처리되는가"를 결정합니다.

### 질문

scale-out하려면 왜 데이터를 나눠야 할까요?

### 직관

책 한 권을 여러 사람이 동시에 정리하려면 장 단위로 나눠 맡겨야 합니다. 그런데 어떤 장이 너무 길면 그 사람만 계속 바빠지고 전체 작업이 느려집니다. 이것이 hot partition이나 skew입니다.

### 작은 예시

사용자 id를 partition key로 쓰고 4개 partition에 hash합니다.

```
user-10 -> p0
user-11 -> p3
user-12 -> p1
celebrity-user -> p2  (but 80% traffic)
```

hash는 균등을 기대하지만 key popularity가 균등하지 않으면 p2가 hot spot이 됩니다.

### 상태 이동 trace

```
record(key)
  |
  v
partitioner: hash(key) % partition_count
  |
  +-- p0 -> node A
  +-- p1 -> node B
  +-- p2 -> node C
  +-- p3 -> node D
```

### 내부 메커니즘

partitioning은 key space를 나누는 규칙입니다. range partitioning은 정렬과 range query에 유리하지만 hot range에 취약합니다. hash partitioning은 분산에 유리하지만 range query가 어렵습니다. consistent hashing은 node 수가 바뀔 때 key 이동을 줄입니다. Spark는 data partition과 task partition을 통해 병렬 실행 단위를 만듭니다.

### 실패 모드

- hot key가 한 partition에 몰리면 cluster 전체가 아니라 한 node가 병목이 됩니다.
- partition 수를 나중에 바꾸면 ordering, key mapping, rebalance 비용이 생깁니다.
- cross-partition transaction이나 join은 partition-local 작업보다 훨씬 비쌉니다.

### 검증 방법

hash 분포 확인:

```bash
python3 - <<'PY'
import hashlib, collections
keys = [f"user-{i}" for i in range(1000)] + ["celebrity"] * 500
def part(k): return int(hashlib.md5(k.encode()).hexdigest(), 16) % 8
print(collections.Counter(part(k) for k in keys))
PY
```

PASS 신호:

- 단순 key 수는 어느 정도 분산되지만, hot key 반복은 특정 partition count를 크게 만듭니다.

FAIL 신호:

- "hash를 쓰면 항상 균등하다"로 결론 냅니다. traffic distribution은 key distribution과 다를 수 있습니다.

### 면접식 되묻기

"partition을 늘리면 성능이 선형으로 좋아지나요?"에는 "아니요"라고 답합니다. hot key, coordination, network, disk, metadata, consumer/task parallelism, rebalance 비용이 함께 변합니다.

### 흔한 오해와 반례

오해: "partition은 storage만 나누는 개념이다."

반례: Kafka partition은 ordering과 consumer parallelism 경계이고, Cassandra token range는 replica placement와 query path 경계이며, Spark partition은 task scheduling과 shuffle 경계입니다.

### Active recall

- hash partitioning과 range partitioning의 tradeoff를 말해 보세요.
- Kafka, Cassandra, Spark에서 partition이 각각 무엇의 경계인지 비교해 보세요.

## 5. Replication, Quorum, Consensus

> replication은 같은 데이터를 여러 node에 두어 durability와 availability를 높이는 방법입니다.
> quorum은 여러 replica 중 몇 개의 응답을 성공으로 볼지 정해 읽기/쓰기 집합이 겹치게 만드는 기법입니다.
> consensus는 여러 node가 하나의 값이나 log 순서에 안전하게 동의하는 문제이며, quorum보다 강한 상태 전이 규칙을 포함합니다.

### 질문

replica가 3개이고 2개가 응답하면 항상 최신 값을 읽는다고 말할 수 있을까요?

### 직관

세 명이 같은 장부 복사본을 들고 있습니다. 쓰기 때 두 명에게 적고, 읽기 때 두 명에게 물으면 최소 한 명은 쓰기를 본 사람과 겹칩니다. 그러나 clock, conflict resolution, failed write, repair 상태에 따라 "최신"의 의미는 더 조심해야 합니다.

### 작은 예시

RF=3, `W=2`, `R=2`이면 `R + W > RF`라서 read set과 write set이 적어도 하나 겹칩니다.

```
replicas: A B C
write x=1 ack from A,B
read asks B,C
intersection: B
```

### 상태 이동 trace

```
client write x=1
  |
  v
coordinator sends to A,B,C
  |
  +-- A ack
  +-- B ack
  +-- C timeout
  |
  v
success if W=2

later read R=2:
  B returns x=1
  C returns old
  coordinator reconciles by timestamp/version/rule
```

### 내부 메커니즘

quorum은 intersection을 이용해 최신 write를 볼 가능성을 높입니다. 그러나 quorum protocol 자체가 모든 상황에서 linearizability를 자동 제공하지는 않습니다. Cassandra는 tunable consistency와 timestamp reconciliation을 사용합니다. Kafka는 partition leader가 write order를 정하고 ISR replica가 leader log를 따라오는 방식입니다. Raft 같은 consensus는 leader election, log matching, commit index, term 같은 규칙을 통해 더 강한 safety를 제공합니다.

### 실패 모드

- quorum을 majority vote로만 이해하면 read repair, failed write, timestamp conflict, stale replica를 놓칩니다.
- consensus를 모든 replication의 동의어로 쓰면 Kafka ISR, Cassandra RF/CL, Spark task retry의 차이를 흐립니다.
- replica가 많아도 같은 rack이나 disk failure domain에 있으면 내구성이 생각보다 낮습니다.

### 검증 방법

quorum intersection 계산:

```bash
python3 - <<'PY'
RF = 3
for R in range(1, RF+1):
    for W in range(1, RF+1):
        print({"R": R, "W": W, "intersects": R + W > RF})
PY
```

PASS 신호:

- `R + W > RF`일 때 intersection이 생깁니다.

FAIL 신호:

- intersection을 곧바로 "항상 최신"이나 "항상 linearizable"로 과장합니다.

### 면접식 되묻기

"quorum과 consensus 차이는?"에는 quorum은 응답 집합 크기와 겹침을 이용하는 기법이고, consensus는 실패가 있어도 여러 node가 같은 값/순서에 안전하게 동의하도록 하는 프로토콜이라고 답합니다.

### 흔한 오해와 반례

오해: "quorum이면 consensus다."

반례: quorum은 consensus 프로토콜 안에서 쓰이기도 하지만, Cassandra의 tunable consistency 같은 quorum read/write는 Raft log agreement와 같은 상태 전이 규칙을 제공하지 않습니다.

### Active recall

- RF=5에서 R=2, W=3이면 intersection이 보장되나요?
- Kafka ISR replication과 Cassandra quorum replication의 차이를 설명해 보세요.

## 6. Consistency Models, CAP, PACELC

> 일관성(consistency)은 모든 시스템이 하나의 의미를 갖지 않습니다. linearizable, sequential, causal, eventual consistency는 각각 다른 읽기 경험을 말합니다.
> CAP은 partition 상황에서 consistency와 availability를 동시에 완전하게 만족할 수 없다는 제한입니다.
> PACELC는 partition이 없을 때도 latency와 consistency 사이 tradeoff가 남는다는 점을 강조합니다.

### 질문

"이 시스템은 consistent한가요?"라는 질문은 왜 불완전할까요?

### 직관

"빠른가요?"라고만 묻는 것과 비슷합니다. 평균이 빠른지, p99가 빠른지, 읽기가 빠른지, 쓰기가 빠른지 물어야 합니다. consistency도 어떤 관측자가 어떤 순서로 무엇을 읽는지 정해야 의미가 생깁니다.

### 작은 예시

```
t1: client A writes x=1
t2: client B reads x
```

linearizable read라면 t2가 t1 이후라면 `x=1`을 보거나 오류를 내야 합니다. eventual consistency에서는 잠깐 `x=0`을 볼 수 있지만 시간이 지나면 수렴해야 합니다.

### 상태 이동 trace

```
write x=1 accepted at replica A
  |
  +-- replica B receives later
  +-- replica C receives later

read from B before propagation:
  eventual: may see old x=0
  quorum with reconciliation: may see x=1 if read set intersects and rule wins
  linearizable: must coordinate enough to respect real-time order
```

### 내부 메커니즘

linearizability는 실제 시간 순서와 일치하는 단일 객체처럼 보이는 강한 모델입니다. causal consistency는 원인 관계가 있는 update 순서를 보존합니다. eventual consistency는 update가 더 이상 없으면 replica가 결국 수렴한다는 약속입니다. CAP은 네트워크 partition이 생겼을 때 request에 계속 응답할지, 아니면 일관성을 위해 일부 요청을 거절할지의 선택을 말합니다. PACELC는 partition이 없을 때도 더 강한 consistency를 위해 더 많은 coordination을 하면 latency가 늘어난다고 설명합니다.

### 실패 모드

- CAP을 "C/A/P 중 둘 고르기"로 외우면 평상시 latency tradeoff와 partition-time 선택을 혼동합니다.
- eventual consistency를 "아무 값이나 나와도 된다"로 오해하면 repair와 convergence 조건을 놓칩니다.
- strong consistency를 요구하면서 cross-region low latency와 high availability를 동시에 기대하면 물리 지연을 무시하게 됩니다.

### 검증 방법

소규모 사고 실험:

```text
RF=3
write x=1 reaches only A before partition
client reads from B during partition

Question:
- availability를 지키려면 B가 응답할 수 있는가?
- consistency를 지키려면 B가 무엇을 해야 하는가?
```

PASS 신호:

- B가 응답하면 stale 가능성이 있고, stale을 막으려면 오류나 대기가 필요하다는 tradeoff를 설명합니다.

FAIL 신호:

- "좋은 DB면 둘 다 된다"처럼 물리적 partition을 무시합니다.

### 면접식 되묻기

"CAP에서 Cassandra는 AP인가요?"에는 "Cassandra는 availability와 partition tolerance를 우선하는 설계 요소가 강하지만, consistency level과 LWT 같은 경로가 있어 단순 라벨로 끝내면 안 됩니다"라고 답합니다.

### 흔한 오해와 반례

오해: "CAP은 평상시에도 세 가지 중 두 개만 가능하다는 말이다."

반례: CAP의 핵심은 partition이 발생했을 때입니다. partition이 없을 때의 latency-consistency tradeoff는 PACELC가 더 직접적인 설명입니다.

### Active recall

- eventual consistency와 linearizability를 read trace로 비교해 보세요.
- PACELC에서 `ELC`가 말하는 평상시 tradeoff는 무엇인가요?

## 7. Idempotency and Retry

> idempotency는 같은 요청을 여러 번 보내도 최종 상태가 한 번 보낸 것과 같게 만드는 성질입니다.
> 분산 시스템에서 retry는 필요하지만, idempotency 없이 retry하면 중복 결제, 중복 메시지, 중복 task side effect가 생깁니다.
> Kafka producer idempotence, consumer offset 처리, Cassandra LWT, Spark output commit은 모두 중복 실행 문제를 다루는 방법입니다.

### 질문

timeout이 났을 때 같은 요청을 다시 보내면 안전할까요?

### 직관

"문 닫기"는 여러 번 해도 문이 닫혀 있으면 상태가 같습니다. "1만 원 입금하기"는 두 번 하면 2만 원이 됩니다. retry 안전성은 요청의 의미에 달려 있습니다.

### 작은 예시

```
POST /payments {requestId=abc, amount=10000}
```

server가 결제는 처리했지만 response가 lost됐습니다. client가 같은 `requestId`로 retry하면 server는 중복 처리 대신 기존 결과를 반환할 수 있습니다.

### 상태 이동 trace

```
client sends requestId=abc
  |
server applies side effect
  |
response lost
  |
client retries requestId=abc
  |
server detects duplicate key abc
  |
returns previous result without new side effect
```

### 내부 메커니즘

idempotency key, unique constraint, deduplication log, producer sequence number, transaction boundary가 중복 방지에 쓰입니다. 설계의 기준은 "요청이 한 번만 도착한다"가 아니라 "여러 번 도착해도 상태 전이가 한 번처럼 보인다"입니다.

### 실패 모드

- retry가 side effect 뒤에 재실행되면 중복 write가 됩니다.
- idempotency key TTL이 너무 짧으면 늦은 retry가 새 요청처럼 처리됩니다.
- deduplication 저장소 자체가 장애 나면 중복 방지가 깨집니다.

### 검증 방법

로컬 사고 실험:

```bash
python3 - <<'PY'
seen = set()
balance = 0
for request_id in ["abc", "abc"]:
    if request_id not in seen:
        balance += 100
        seen.add(request_id)
print(balance)
PY
```

PASS 신호:

- 같은 request id를 두 번 처리해도 balance가 100입니다.

FAIL 신호:

- request id 없이 retry 횟수만 믿습니다.

### 면접식 되묻기

"at-least-once와 exactly-once 차이는?"에는 delivery와 processing을 분리합니다. 적어도 한 번 전달은 중복 가능성을 포함하고, exactly-once processing은 source offset, output write, deduplication/transaction boundary가 함께 맞아야 합니다.

### 흔한 오해와 반례

오해: "HTTP PUT은 항상 idempotent다."

반례: method semantics는 의도일 뿐 구현이 side effect를 어떻게 만들었는지가 중요합니다. `PUT /balance/add/100`처럼 구현하면 idempotent하지 않을 수 있습니다.

### Active recall

- retry가 필요한 이유와 retry가 위험한 이유를 동시에 설명해 보세요.
- Kafka consumer가 처리 후 offset commit 전에 죽으면 어떤 중복이 생기나요?

## 8. Backpressure, Checkpointing, Recovery

> backpressure는 생산 속도와 소비 속도 사이의 불균형이 queue, lag, buffer, spill, timeout으로 드러나는 현상입니다.
> checkpoint는 긴 log나 lineage를 줄여 실패 후 재시작 지점을 만드는 기록입니다.
> recovery는 "어디까지 확정됐고, 어디부터 다시 실행해야 하며, 다시 실행해도 안전한가"를 결정하는 과정입니다.

### 질문

consumer lag가 늘거나 Spark job이 spill을 시작하면 단순히 worker를 늘리면 될까요?

### 직관

배수구가 막힌 욕조에 물을 더 빨리 붓는다고 문제가 해결되지 않습니다. 어디가 좁은지 찾아야 합니다. 좁은 곳이 CPU인지, network인지, disk인지, downstream DB인지, partition skew인지에 따라 처방이 다릅니다.

### 작은 예시

Kafka topic에 초당 10,000 records가 들어오고 consumer는 초당 6,000 records만 처리합니다. lag는 초당 4,000씩 증가합니다.

### 상태 이동 trace

```
producer rate: 10k/s
  |
  v
broker partition log
  |
  | consumer read rate: 6k/s
  v
consumer offset lags behind log end

lag growth = input rate - processing rate
```

Spark shuffle spill:

```
task input partition too large
  |
  v
executor memory fills execution region
  |
  v
spill sorted/hash data to local disk
  |
  v
stage duration and disk I/O rise
```

### 내부 메커니즘

backpressure는 queueing theory의 기본 현상입니다. 도착률이 처리률보다 지속적으로 높으면 queue는 자랍니다. checkpoint는 recovery 비용을 줄입니다. Kafka consumer offset은 "어디까지 읽고 처리했다고 볼 것인가"의 checkpoint입니다. Spark checkpoint는 lineage가 너무 길거나 streaming state recovery가 필요할 때 reliable storage에 상태를 남깁니다. Cassandra repair와 SSTable metadata도 replica convergence와 recovery에 관여합니다.

### 실패 모드

- lag를 보고 consumer 수만 늘렸는데 partition 수가 부족하면 병렬성이 늘지 않습니다.
- checkpoint가 local disk에 있으면 node loss 때 recovery가 안 됩니다.
- backpressure를 무시하고 retry를 늘리면 thundering herd가 생깁니다.

### 검증 방법

간단한 queue simulation:

```bash
python3 - <<'PY'
queue = 0
for sec in range(1, 6):
    queue += 10000
    queue -= 6000
    print(sec, queue)
PY
```

PASS 신호:

- 처리률이 도착률보다 낮으면 queue가 선형 증가합니다.

FAIL 신호:

- 순간 queue 크기만 보고 안정/불안정을 판단합니다. 추세가 더 중요합니다.

### 면접식 되묻기

"consumer lag가 늘 때 어떻게 대응하나요?"에는 lag의 증가율, partition별 분포, processing latency, downstream latency, broker fetch latency, rebalance 여부를 나눕니다. 그 다음 scale-out, partition 재설계, batch 조정, downstream 최적화 중 병목에 맞는 처방을 고릅니다.

### 흔한 오해와 반례

오해: "backpressure는 consumer가 느린 문제다."

반례: consumer가 느려 보이는 원인이 downstream DB, network, broker throttling, disk read, GC, partition skew일 수 있습니다.

### Active recall

- checkpoint와 offset의 공통점과 차이를 말해 보세요.
- Spark spill과 Kafka lag를 같은 queueing 모델로 설명해 보세요.

## 현실 시나리오 1: 네트워크 지연과 노드 crash를 구분할 수 없다

> timeout은 실패의 증거가 아니라 관측의 한계입니다.
> 더 나은 추론은 단일 timeout을 crash로 단정하지 않고, 여러 노드의 관측과 OS/application metric을 교차 확인합니다.

1. 관측된 증상

    client가 replica B에 요청했지만 500ms timeout이 발생했습니다.

2. 가능한 원인 후보

    B crash, B GC pause, B disk saturation, packet loss, client scheduler delay, firewall/drop, overloaded coordinator.

3. OS/kernel 관점에서 볼 지점

    `ss -tin`, retransmission, run queue, GC log, disk await, process state, kernel drops.

4. distributed-system 관점에서 볼 지점

    다른 노드도 B를 의심하는지, gossip/heartbeat가 어떻게 변했는지, quorum이 어떤 replica로 구성됐는지 봅니다.

5. Kafka/Cassandra/Spark 내부 구조와 연결되는 지점

    Kafka ISR shrink, Cassandra failure detector conviction, Spark executor lost/task retry가 같은 timeout을 서로 다른 정책으로 해석합니다.

6. 확인 명령 또는 로그

    `ss -tin`, `vmstat 1`, application heartbeat log, Kafka controller log, Cassandra system.log, Spark driver event log.

7. 잘못된 결론의 예

    "timeout이 났으니 B는 죽었다."

8. 더 나은 추론 과정

    local host, network path, remote host, application runtime, distributed membership view를 나눠 관측합니다. 모든 증거가 remote process death를 가리킬 때 crash로 좁힙니다.

## 현실 시나리오 2: QUORUM인데 stale read가 의심된다

> quorum은 읽기와 쓰기 집합이 겹치게 만드는 도구지만, 모든 일관성 의심을 자동으로 제거하지 않습니다.
> timestamp, failed write, repair, read path, consistency level, clock skew를 함께 봐야 합니다.

1. 관측된 증상

    Cassandra에서 `LOCAL_QUORUM` read를 했는데 직전 update가 보이지 않는 사례가 보고됐습니다.

2. 가능한 원인 후보

    write가 실제로 quorum ack를 받지 못함, 다른 datacenter CL 혼동, client timestamp 충돌, clock skew, read repair 경계, 잘못된 partition key, application cache stale.

3. OS/kernel 관점에서 볼 지점

    node time sync, disk latency, GC pause, network latency, process health를 확인합니다.

4. distributed-system 관점에서 볼 지점

    RF, read CL, write CL, quorum intersection, failed write visibility, reconciliation rule을 확인합니다.

5. Kafka/Cassandra/Spark 내부 구조와 연결되는 지점

    Cassandra timestamp 기반 LWW와 repair path를 봅니다. Kafka라면 offset commit과 isolation level, Spark라면 checkpoint와 sink commit 경계를 봤을 것입니다.

6. 확인 명령 또는 로그

    CQL consistency settings, client logs, Cassandra `system.log`, `nodetool status`, tracing, table schema timestamp policy.

7. 잘못된 결론의 예

    "QUORUM이면 무조건 최신이어야 하니 Cassandra가 버그다."

8. 더 나은 추론 과정

    "최신"의 기준을 먼저 정합니다. write success boundary, replica set, timestamp winner, read coordinator reconciliation을 확인한 뒤 Cassandra 보장과 애플리케이션 기대가 맞는지 봅니다.

## 근거와 더 읽을 자료

- [10_source_ledger.md](10_source_ledger.md)의 Lamport clocks, Raft, CAP, PACELC, Dynamo, Kafka, Cassandra, Spark RDD 항목
- Lamport clocks: https://lamport.org/pubs/time-clocks.pdf
- Raft paper: https://raft.github.io/raft.pdf
- Dynamo paper: https://www.amazon.science/publications/dynamo-amazons-highly-available-key-value-store
- CAP proof: https://groups.csail.mit.edu/tds/papers/Gilbert/Brewer2.pdf
- PACELC: https://www.cs.umd.edu/~abadi/papers/abadi-pacelc.pdf
