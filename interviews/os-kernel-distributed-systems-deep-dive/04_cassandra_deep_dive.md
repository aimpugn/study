# 04. Cassandra Deep Dive

> Cassandra는 여러 노드가 같은 key range의 replica를 나눠 갖고, coordinator가 consistency level에 맞춰 read/write 응답 수를 모으는 분산 wide-column database입니다.
> 쓰기는 commit log와 memtable에 빠르게 반영한 뒤 immutable SSTable로 흘러가고, 읽기는 memtable, bloom filter, index, SSTable, compaction 상태를 함께 거칩니다.
> Cassandra의 성능과 일관성 문제는 token ring, replication factor, consistency level, timestamp reconciliation, compaction pressure, repair 상태를 함께 봐야 해석됩니다.

## 1. Ring, Partitioning, Consistent Hashing

> Cassandra는 partition key를 hash하여 token ring 위 위치로 바꾸고, 그 token range를 소유한 replica들에게 data를 배치합니다.
> consistent hashing은 노드 추가/제거 때 전체 key를 다시 배치하지 않고 일부 range만 움직이게 합니다.
> vnode는 균형과 운영 유연성을 주지만 repair, failure combination, token ownership 추론을 더 복잡하게 만들 수 있습니다.

### 질문

Cassandra는 어떤 노드가 어떤 row를 저장할지 어떻게 정할까요?

### 직관

원형 트랙 위에 노드들이 표지판처럼 서 있고, key를 hash한 값이 트랙의 한 지점에 떨어집니다. 그 지점에서 시계 방향으로 걸어가며 만나는 노드들이 replica가 됩니다.

### 작은 예시

replication factor 3, token ring에 `A, B, C, D`가 있습니다.

```
hash(user-10) = token 37

ring order:
  A(0) -> B(25) -> C(50) -> D(75) -> A(100)

token 37 belongs after B before C
replicas: C, D, A
```

### 상태 이동 trace

```
partition key: user-10
  |
  v
Murmur3 hash -> token 37
  |
  v
walk token ring
  |
  v
replica set = [C, D, A]
  |
  v
coordinator sends mutation/read to replicas according to CL
```

### 내부 메커니즘

Cassandra table은 partition key를 기준으로 row를 나눕니다. 빠른 query는 partition key로 시작합니다. partition key hash는 token을 만들고, token map은 token range와 node endpoint를 연결합니다. `NetworkTopologyStrategy`는 datacenter와 rack 같은 failure domain을 고려해 replica를 고릅니다.

vnode는 한 physical node가 여러 token을 갖게 합니다. 이로써 새 노드가 들어올 때 여러 작은 range를 조금씩 가져가 균형을 맞출 수 있습니다. 그러나 token이 많아질수록 repair 단위와 failure 조합이 늘어날 수 있습니다.

### 실패 모드

- partition key가 잘못 설계되면 hot partition이나 wide partition이 생깁니다.
- node 추가/제거 때 streaming과 compaction이 발생해 disk/network에 압력이 생깁니다.
- rack 설정이 잘못되면 replica가 원하는 failure domain에 분산되지 않습니다.

### 검증 방법

local Cassandra가 있다면:

```bash
nodetool status
nodetool ring
```

PASS 신호:

- node별 token ownership과 상태를 확인합니다.
- data distribution imbalance를 의심할 지표를 볼 수 있습니다.

FAIL 신호:

- `nodetool status`의 `UN`만 보고 data balance와 query hot spot이 괜찮다고 결론 냅니다.

### 면접식 되묻기

"consistent hashing은 왜 쓰나요?"에는 node 수가 변할 때 key 이동을 줄여 scale-out과 decommission 비용을 낮춘다고 답합니다. 꼬리 질문에서는 vnode tradeoff, RF, rack-aware placement로 확장합니다.

### 흔한 오해와 반례

오해: "hash partitioning이면 hot spot이 없다."

반례: key count가 균등해도 traffic이 특정 key에 몰리면 hot partition이 됩니다. time bucket이나 synthetic shard를 설계해야 할 수 있습니다.

### Active recall

- partition key에서 replica set까지의 trace를 그려 보세요.
- vnode의 장점과 비용을 하나씩 말해 보세요.

## 2. Replication Factor, Quorum, Tunable Consistency

> replication factor(RF)는 한 partition이 몇 replica에 저장될지 정하고, consistency level(CL)은 read/write가 몇 replica 응답을 기다릴지 정합니다.
> `R + W > RF`이면 read set과 write set이 겹치는 quorum intersection이 생기지만, 이것만으로 모든 read가 linearizable해지는 것은 아닙니다.
> Cassandra는 availability와 latency를 위해 consistency를 조절할 수 있게 만들었고, 그 대가로 timestamp, repair, failed write 경계를 이해해야 합니다.

### 질문

Cassandra에서 `QUORUM`으로 읽으면 언제나 가장 최신 값을 읽는다고 해도 될까요?

### 직관

세 명 중 두 명에게 쓰고, 나중에 두 명에게 물어보면 적어도 한 명은 새 값을 본 사람입니다. 하지만 그 사람이 어떤 값을 "이긴 값"으로 내놓는지는 timestamp와 reconciliation 규칙, write 성공 여부, clock 상태에 따라 달라질 수 있습니다.

### 작은 예시

RF=3, replicas `A, B, C`.

```
write CL=QUORUM:
  A <- x=1 ack
  B <- x=1 ack
  C <- timeout

read CL=QUORUM:
  B -> x=1
  C -> x=0
coordinator reconciles -> x=1 if timestamp/version wins
```

### 상태 이동 trace

```
client write
  |
  v
coordinator chooses replicas
  |
  +-- A write + ack
  +-- B write + ack
  +-- C missed
  |
  v
write success at CL=QUORUM
  |
  v
later read asks B,C
  |
  v
compare cells by timestamp and tombstone rules
  |
  v
return winning value, maybe trigger repair path
```

### 내부 메커니즘

Cassandra coordinator는 client가 접속한 노드일 뿐, data owner와 같을 필요가 없습니다. write는 replica들에게 mutation을 보내고 CL에 필요한 ack 수를 기다립니다. read는 digest request와 data request를 조합해 replica 값 차이를 감지할 수 있습니다. 충돌은 cell timestamp, tombstone, TTL 같은 metadata를 기반으로 조정됩니다.

Lightweight transaction(LWT) 같은 별도 경로는 linearizable compare-and-set semantics를 제공합니다. 일반 write/read CL과 같은 비용 모델로 보면 안 됩니다.

### 실패 모드

- CL=ONE write 후 QUORUM read를 하면 이전 write가 충분히 전파되지 않았을 수 있습니다.
- client timestamp가 꼬이면 나중에 쓴 값이 더 오래된 timestamp로 져서 보이지 않을 수 있습니다.
- failed write가 일부 replica에 남아 이후 read에서 나타나는 edge case가 생길 수 있습니다.

### 검증 방법

local multi-node cluster가 있다면 CQL tracing을 켭니다.

```sql
CONSISTENCY QUORUM;
TRACING ON;
SELECT * FROM orders WHERE user_id = 'u1';
```

PASS 신호:

- coordinator가 어떤 replica에 read를 보내고 어떤 latency가 나오는지 trace에서 확인합니다.

FAIL 신호:

- cqlsh에서 보인 값 하나만 보고 모든 replica가 같은 상태라고 가정합니다.

### 면접식 되묻기

"Cassandra는 consistency를 어떻게 조절하나요?"에는 RF와 CL을 분리해 답합니다. RF는 복사본 개수, CL은 성공으로 인정할 응답 개수입니다. 꼬리 질문이 오면 quorum intersection, LWT 예외, stale read 원인, repair path를 설명합니다.

### 흔한 오해와 반례

오해: "eventual consistency라서 Cassandra는 일관성을 포기한 DB다."

반례: Cassandra는 CL과 LWT로 tradeoff를 조정합니다. 다만 기본 설계는 cross-partition coordination을 피하고 availability와 scale-out을 우선합니다.

### Active recall

- RF=3에서 `W=1`, `R=1`과 `W=2`, `R=2`의 차이를 trace로 설명해 보세요.
- QUORUM과 linearizable read를 혼동하면 어떤 답변이 틀리나요?

## 3. Write Path: Commit Log, Memtable, SSTable

> Cassandra write는 기존 disk row를 제자리에서 고치지 않고, commit log에 append하고 memtable에 반영한 뒤 나중에 immutable SSTable로 flush합니다.
> 이 구조는 random write를 줄여 write throughput을 높이지만, 시간이 지나면 많은 SSTable과 compaction 비용을 만듭니다.
> commit log durability, memtable flush, SSTable creation은 OS page cache와 disk flush 정책의 영향을 받습니다.

### 질문

Cassandra는 update가 많은데 왜 disk의 기존 row를 바로 덮어쓰지 않을까요?

### 직관

노트의 중간 페이지를 계속 지우고 다시 쓰는 것보다, 변경 내역을 맨 뒤에 계속 적고 나중에 정리하는 편이 빠릅니다. Cassandra의 LSM 구조는 이 방식을 disk에 맞게 만든 것입니다.

### 작은 예시

`user_id=u1, status=ACTIVE`를 쓰면:

```
commit log: append mutation
memtable:   in-memory sorted structure update
SSTable:    later flush to immutable disk files
```

### 상태 이동 trace

```
client mutation
  |
  v
coordinator sends to replicas
  |
  v
replica write path
  |
  +-- append to commit log
  +-- apply to memtable
  |
  v
ack according to CL
  |
  v
memtable threshold reached
  |
  v
flush immutable SSTable to disk
```

### 내부 메커니즘

commit log는 crash recovery를 위한 append-only 기록입니다. memtable은 memory에 있는 sorted write buffer입니다. memtable이 커지면 flush되어 SSTable이 됩니다. SSTable은 immutable이므로 update와 delete는 새 version과 tombstone으로 기록됩니다. 이 때문에 write path는 빠르지만 read path는 여러 SSTable을 봐야 할 수 있습니다.

commit log sync mode와 flush 정책은 버전과 설정에 따라 달라집니다. 본문에서는 "commit log가 durability boundary의 일부"라고 설명하되, 특정 기본값은 target version 문서로 확인해야 합니다.

### 실패 모드

- commit log disk가 느리면 write latency가 증가합니다.
- memtable flush가 compaction과 겹치면 disk write pressure가 커집니다.
- tombstone이 많으면 read와 compaction 비용이 커지고, 오래된 delete가 부활하는 위험도 repair 설정과 함께 고려해야 합니다.

### 검증 방법

local Cassandra:

```bash
nodetool tablestats <keyspace>.<table>
nodetool tpstats
```

PASS 신호:

- memtable, SSTable count, write latency, flush/compaction 관련 지표를 확인합니다.

FAIL 신호:

- write latency만 보고 network 문제로 단정합니다. commit log disk와 memtable/flush도 확인해야 합니다.

### 면접식 되묻기

"LSM tree가 write에 유리한 이유는?"에는 random in-place update를 줄이고 sequential append/flush 중심으로 만들기 때문이라고 답합니다. 꼬리 질문에서는 read amplification, space amplification, compaction pressure를 반드시 붙입니다.

### 흔한 오해와 반례

오해: "Cassandra write는 memory에만 쓰니까 빠르다."

반례: commit log append가 durability를 위해 disk path와 연결됩니다. memory만 보고 write가 완료되는 구조로 설명하면 crash recovery를 놓칩니다.

### Active recall

- mutation이 commit log, memtable, SSTable로 가는 순서를 말해 보세요.
- immutable SSTable이 update/delete 비용을 어떻게 바꾸나요?

## 4. Read Path, Bloom Filter, Compaction

> Cassandra read는 memtable과 여러 SSTable 후보를 확인하고, bloom filter와 index를 이용해 불필요한 disk access를 줄입니다.
> SSTable이 많아질수록 read amplification이 생기며, compaction은 여러 SSTable을 병합해 최신 값과 tombstone을 정리합니다.
> compaction은 읽기를 도와주지만, 동시에 강한 disk I/O와 CPU 압력을 만들어 운영 병목이 될 수 있습니다.

### 질문

write가 빠른 LSM 구조에서 read는 왜 복잡해질까요?

### 직관

변경 내역을 여러 노트에 계속 적었다면, 현재 값을 알려면 최신 노트뿐 아니라 예전 노트의 삭제 표시와 수정 내역도 확인해야 합니다. compaction은 여러 노트를 합쳐 정리하는 일입니다.

### 작은 예시

같은 partition의 값이 세 SSTable에 흩어져 있습니다.

```
SSTable-1: status=PENDING, ts=10
SSTable-2: status=ACTIVE,  ts=20
SSTable-3: tombstone,      ts=30
```

read는 timestamp와 tombstone 규칙으로 winning state를 결정해야 합니다.

### 상태 이동 trace

```
read partition key u1
  |
  +-- check memtable
  +-- check row cache/key cache if enabled
  +-- SSTable bloom filter: maybe contains u1?
  |       |
  |       +-- no: skip SSTable
  |       +-- maybe: index lookup -> data read
  |
  v
merge versions by timestamp/tombstone
  |
  v
return row
```

Compaction:

```
SSTable A + SSTable B + SSTable C
  |
  v
merge sorted rows, discard obsolete versions when safe
  |
  v
new SSTable D
  |
  v
old SSTables removed after safe point
```

### 내부 메커니즘

bloom filter는 "없다"를 빠르게 말해 주는 probabilistic structure입니다. false positive는 있을 수 있지만 false negative는 없어야 합니다. 그래서 bloom filter가 "없다"고 하면 SSTable을 건너뜁니다. "있을 수 있다"고 하면 index와 data file을 봅니다.

compaction은 SSTable 수를 줄이고 오래된 value와 tombstone을 정리합니다. 그러나 compaction 자체는 기존 SSTable을 읽고 새 SSTable을 쓰는 작업이라 read/write amplification과 disk space temporary overhead를 만듭니다.

### 실패 모드

- compaction backlog가 쌓이면 read amplification이 증가합니다.
- tombstone이 많으면 read가 느려지고 timeout이 발생할 수 있습니다.
- compaction이 disk bandwidth를 잡아먹어 foreground read/write latency를 악화시킬 수 있습니다.

### 검증 방법

local Cassandra:

```bash
nodetool compactionstats
nodetool tablestats <keyspace>.<table>
```

PASS 신호:

- pending compactions, SSTable count, read latency, tombstone warning을 함께 봅니다.

FAIL 신호:

- compaction을 단순 cleanup 작업으로만 보고 latency 영향이 없다고 생각합니다.

### 면접식 되묻기

"Bloom filter는 Cassandra read에서 무엇을 줄이나요?"에는 특정 SSTable에 partition key가 없다는 것을 빠르게 판단해 disk/index read를 줄인다고 답합니다. false positive는 가능하지만 false negative가 없어야 한다는 점을 붙입니다.

### 흔한 오해와 반례

오해: "compaction은 disk 공간을 줄이는 작업이다."

반례: 공간 회수도 하지만 read amplification 감소, tombstone 처리, SSTable version upgrade, repair-related compaction 등 성능과 correctness에 영향을 줍니다.

### Active recall

- read path에서 bloom filter가 있는 위치를 trace로 그려 보세요.
- compaction이 read latency를 좋게도 나쁘게도 만들 수 있는 이유는 무엇인가요?

## 5. Hinted Handoff, Read Repair, Anti-entropy Repair

> Cassandra replica는 일시적으로 불일치할 수 있고, 이를 줄이기 위해 hints, read repair, anti-entropy repair를 사용합니다.
> hinted handoff는 unavailable replica가 돌아왔을 때 missed mutation을 전달하려는 best-effort 경로입니다.
> read repair와 anti-entropy repair는 replica 수렴을 돕지만, 서로 비용과 보장 범위가 다릅니다.

### 질문

replica 하나가 잠깐 내려갔다 올라오면 Cassandra는 어떻게 빠진 write를 맞출까요?

### 직관

회의에 빠진 사람에게 회의록을 임시로 맡겨 뒀다가 돌아오면 전달합니다. 하지만 회의록 전달만 믿고 장기 결석자의 모든 상태가 완벽히 맞는다고 보면 안 됩니다. 정기적인 대조와 정리가 필요합니다.

### 작은 예시

RF=3, CL=QUORUM write 중 replica C가 down입니다.

```
A ack
B ack
C down
coordinator stores hint for C
later C up
coordinator replays hint
```

### 상태 이동 trace

```
write x=1
  |
  +-- replica A applies
  +-- replica B applies
  +-- replica C unavailable
          |
          v
      coordinator stores hint
          |
          v
      C rejoins
          |
          v
      hint replay to C
```

### 내부 메커니즘

hints는 unavailable replica를 위한 임시 mutation 기록입니다. Cassandra 공식 문서도 hints를 best-effort로 설명하며 anti-entropy repair 대체가 아니라고 경계를 둡니다. read repair는 read 과정에서 replica 간 차이를 발견했을 때 조정하는 경로입니다. anti-entropy repair는 Merkle tree 같은 방식으로 replica data 차이를 비교해 수렴을 강화합니다.

### 실패 모드

- hint window를 넘긴 downtime은 hints로 복구되지 않습니다.
- repair를 오래 돌리지 않으면 tombstone과 missed update 문제가 커질 수 있습니다.
- read repair 설정과 consistency level을 오해하면 monotonic read와 write atomicity tradeoff를 놓칩니다.

### 검증 방법

local cluster:

```bash
nodetool netstats
nodetool repair --preview <keyspace> <table>
```

PASS 신호:

- streaming, repair preview, pending hints 관련 지표를 확인합니다.

FAIL 신호:

- hints가 있으니 repair가 필요 없다고 결론 냅니다.

### 면접식 되묻기

"hinted handoff가 eventual consistency를 보장하나요?"에는 "아니요. hints는 불일치 시간을 줄이는 best-effort이고, anti-entropy repair가 더 강한 수렴 장치입니다"라고 답합니다.

### 흔한 오해와 반례

오해: "read repair는 항상 좋은 것이다."

반례: read repair는 foreground read latency와 write atomicity tradeoff에 영향을 줄 수 있습니다. 버전과 설정에 따라 동작이 달라지므로 target version 문서를 확인해야 합니다.

### Active recall

- hints, read repair, anti-entropy repair의 역할을 한 문장씩 구분해 보세요.
- replica down 시간이 길어질 때 어떤 recovery 경계가 열리나요?

## 6. OS Resource Links: Disk, Memory, Page Cache, Compaction Pressure

> Cassandra는 JVM process이지만 성능의 많은 부분은 OS disk, page cache, memory pressure, file descriptor, network queue와 맞닿아 있습니다.
> compaction은 CPU와 disk를 쓰는 background 작업이지만 foreground read/write latency에 직접 영향을 줄 수 있습니다.
> Cassandra 장애를 볼 때는 CQL latency와 `nodetool`만 보지 말고 `iostat`, `vmstat`, GC log, network stats를 함께 봐야 합니다.

### 질문

Cassandra read/write latency가 증가할 때 왜 OS 지표를 같이 봐야 할까요?

### 직관

식당 주문 시스템이 느릴 때 주문 앱만 볼 수 없습니다. 주방 조리대, 창고, 배달 동선이 막혔는지도 봐야 합니다. Cassandra의 주방은 disk와 memory입니다.

### 작은 예시

compaction이 큰 SSTable 세 개를 읽고 새 SSTable 하나를 씁니다. 동시에 client read가 여러 SSTable을 읽어야 합니다. disk queue가 길어지고 read latency가 증가합니다.

### 상태 이동 trace

```
foreground read
  |
  +-- SSTable reads

background compaction
  |
  +-- read old SSTables
  +-- write new SSTable

both share:
  disk bandwidth, page cache, CPU, file handles
```

### 내부 메커니즘

Cassandra heap에는 memtable, metadata, runtime object가 있고 off-heap/native memory와 OS page cache도 중요합니다. heap을 너무 크게 잡으면 GC 비용과 page cache 감소가 문제가 될 수 있습니다. disk는 commit log, SSTable read/write, compaction, repair streaming을 동시에 처리합니다.

### 실패 모드

- disk utilization이 높아 commit log sync와 SSTable read가 같이 느려집니다.
- memory pressure가 page cache를 밀어내 read miss가 증가합니다.
- GC pause가 coordinator timeout과 replica failure suspicion으로 번집니다.

### 검증 방법

```bash
nodetool tpstats
nodetool compactionstats
iostat -xz 1 5
vmstat 1 5
jcmd <pid> GC.heap_info
```

PASS 신호:

- Cassandra internal queue와 OS disk/memory 지표가 같은 시간대에 악화되는지 확인합니다.

FAIL 신호:

- `nodetool status`가 모두 `UN`이면 성능도 정상이라고 판단합니다.

### 면접식 되묻기

"Cassandra에서 compaction이 왜 위험할 수 있나요?"에는 읽기 성능 개선을 위한 병합 작업이 동시에 disk read/write와 CPU를 많이 써 foreground latency를 해칠 수 있다고 답합니다.

### 흔한 오해와 반례

오해: "Cassandra는 write-optimized라 disk가 느려도 괜찮다."

반례: commit log, flush, compaction, repair, read path 모두 disk에 닿습니다. write-optimized는 disk 비용이 사라진다는 뜻이 아니라 비용을 append와 background merge로 재배치한다는 뜻입니다.

### Active recall

- Cassandra heap을 무작정 키우면 OS page cache에 어떤 일이 생기나요?
- compaction backlog와 read amplification은 어떻게 연결되나요?

## 현실 시나리오 1: stale read가 의심된다

> Cassandra에서 stale read 의심은 "Cassandra가 eventual consistency라서 그렇다"로 끝내면 안 됩니다.
> RF, read/write CL, timestamp, failed write, repair, client cache, datacenter 경계를 순서대로 확인해야 합니다.

1. 관측된 증상

    사용자가 방금 변경한 profile 값이 일부 요청에서 이전 값으로 보입니다.

2. 가능한 원인 후보

    write CL이 낮음, read CL이 낮음, application cache, client가 다른 datacenter를 읽음, timestamp 충돌, replica repair 지연, failed write residue.

3. OS/kernel 관점에서 볼 지점

    node clock sync, GC pause, disk latency, network timeout, process health.

4. distributed-system 관점에서 볼 지점

    quorum intersection, partial failure, read-your-writes 필요성, consistency model 기대.

5. Cassandra 내부 구조와 연결되는 지점

    coordinator read reconciliation, timestamp LWW, hints, read repair, anti-entropy repair.

6. 확인 명령 또는 로그

```sql
CONSISTENCY;
TRACING ON;
SELECT * FROM profile WHERE user_id='u1';
```

```bash
nodetool status
nodetool netstats
```

7. 잘못된 결론의 예

    "Cassandra는 원래 오래된 값을 반환하니 해결할 수 없다."

8. 더 나은 추론 과정

    먼저 필요한 consistency 모델을 정합니다. 같은 user가 쓴 뒤 바로 읽어야 하는지, cross-region latency를 어디까지 허용하는지, LWT가 필요한지, CL 조정으로 충분한지 판단합니다.

## 현실 시나리오 2: compaction 때문에 cluster가 느려진다

> compaction은 background 작업이지만 foreground workload와 같은 disk, CPU, page cache를 공유합니다.
> compaction pressure는 단순 maintenance가 아니라 latency와 timeout을 만드는 운영 사건입니다.

1. 관측된 증상

    read latency p99가 증가하고 `nodetool compactionstats`에 pending compaction이 쌓입니다.

2. 가능한 원인 후보

    high write rate, tombstone-heavy workload, wrong compaction strategy, undersized disk bandwidth, repair streaming, wide partition.

3. OS/kernel 관점에서 볼 지점

    `iostat` await/util, disk queue, `vmstat` I/O wait, page cache pressure, CPU sys time.

4. distributed-system 관점에서 볼 지점

    replica별 compaction imbalance, read quorum이 느린 replica를 기다리는지, coordinator timeout.

5. Cassandra 내부 구조와 연결되는 지점

    SSTable count, read amplification, tombstone scan, compaction throughput, memtable flush.

6. 확인 명령 또는 로그

```bash
nodetool compactionstats
nodetool tablestats <ks>.<table>
iostat -xz 1 5
```

7. 잘못된 결론의 예

    "compaction을 끄면 latency가 회복된다."

8. 더 나은 추론 과정

    compaction을 무작정 끄면 SSTable 수와 read amplification이 더 나빠질 수 있습니다. compaction strategy, tombstone pattern, disk capacity, throttling, data model을 같이 봅니다.

## 근거와 더 읽을 자료

- [10_source_ledger.md](10_source_ledger.md)의 Cassandra architecture, Dynamo, storage engine, hints, compaction, LSM-tree 항목
- Cassandra architecture overview: https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/overview.html
- Cassandra Dynamo architecture: https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/dynamo.html
- Cassandra storage engine: https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/storage-engine.html
- Cassandra hints: https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/hints.html
- Cassandra compaction overview: https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/compaction/overview.html
- LSM-tree paper: https://dsf.berkeley.edu/cs286/papers/lsm-acta1996.pdf
