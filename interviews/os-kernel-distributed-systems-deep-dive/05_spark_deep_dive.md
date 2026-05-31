# 05. Spark Deep Dive

> Spark는 큰 데이터를 driver가 세운 실행 계획(DAG)에 따라 executor process들이 partition 단위 task로 처리하는 분산 실행 엔진입니다.
> Spark의 fault tolerance는 모든 데이터를 복제하는 방식보다, 데이터 partition을 어떻게 다시 계산할 수 있는지 lineage와 checkpoint로 기록하는 방식에 가깝습니다.
> Spark 성능 문제는 DAG stage, shuffle, partition skew, executor memory, GC, disk spill, network fetch 중 어디에서 비용이 커졌는지 따라가야 합니다.

## 1. Driver, Executor, Cluster Manager

> driver는 사용자 프로그램의 main 흐름과 SparkContext/SparkSession을 가진 조정자이고, executor는 worker node에서 task를 실행하고 cache/shuffle data를 보관하는 JVM process입니다.
> cluster manager는 executor를 띄울 자원을 배정하며, Spark 자체의 작업 순서를 결정하는 주체와는 역할이 다릅니다.
> driver가 executor와 통신할 수 없거나 executor가 memory/disk/network 압력을 받으면 Spark job은 논리적으로 맞아도 실행에서 실패합니다.

### 질문

Spark job은 하나의 프로그램처럼 보이는데 실제로는 어디서 실행될까요?

### 직관

driver는 지휘자이고 executor는 악기 연주자입니다. 지휘자가 악보를 나누고 순서를 지시하지만, 실제 소리는 여러 연주자가 각자 자기 파트를 연주할 때 납니다.

### 작은 예시

```python
df = spark.read.parquet("/data/orders")
df.groupBy("country").count().show()
```

driver는 이 코드를 받아 logical/physical plan을 만들고 job을 제출합니다. executor는 partition별 scan, shuffle, aggregation task를 실행합니다.

### 상태 이동 trace

```
user code in driver
  |
  v
SparkContext builds job
  |
  v
DAG scheduler creates stages
  |
  v
cluster manager allocates executors
  |
  v
task scheduler sends tasks to executors
  |
  v
executors read data, compute, shuffle, cache/spill
  |
  v
driver collects status/result metadata
```

### 내부 메커니즘

Spark application은 driver와 executor process 집합으로 구성됩니다. driver는 task scheduling과 metadata를 들고 있으므로 driver가 죽으면 application이 실패하는 경우가 많습니다. executor는 task를 여러 thread로 실행하고, cache와 shuffle file을 local memory/disk에 둡니다. cluster manager는 Standalone, YARN, Kubernetes 같은 자원 관리자입니다.

### 실패 모드

- driver가 remote network에 있어 executor가 driver에 연결하지 못하면 job이 시작되지 않거나 중간에 실패합니다.
- executor memory가 부족하면 OOM, GC pause, spill 증가가 생깁니다.
- executor lost는 cached data와 shuffle output 손실로 이어질 수 있습니다.

### 검증 방법

local Spark:

```bash
spark-submit --master local[2] examples/src/main/python/pi.py 10
```

Spark UI:

```text
http://localhost:4040
```

PASS 신호:

- driver log, executor/task count, stage timeline을 확인합니다.

FAIL 신호:

- local mode에서 잘 되었다고 cluster network와 executor placement 문제가 없다고 단정합니다.

### 면접식 되묻기

"driver와 executor 차이는?"에는 driver는 job을 만들고 scheduling을 조정하며, executor는 worker에서 task를 실행하고 data/cache/shuffle을 관리한다고 답합니다. 꼬리 질문에서는 driver failure, executor lost, cluster manager 역할을 분리합니다.

### 흔한 오해와 반례

오해: "cluster manager가 Spark의 DAG를 최적화한다."

반례: cluster manager는 자원 배정을 담당합니다. Spark의 DAG scheduling과 query planning은 Spark driver 쪽 역할입니다.

### Active recall

- driver가 멀리 있는 네트워크에 있으면 어떤 문제가 생기나요?
- executor process가 OS 관점에서 어떤 자원을 많이 쓰나요?

## 2. RDD, DataFrame, DAG

> RDD는 partitioned collection과 transformation lineage를 가진 Spark의 기본 추상화이고, DataFrame은 schema와 optimizer를 활용하는 더 높은 수준의 추상화입니다.
> Spark는 action이 호출될 때 transformation graph를 DAG로 만들고 실행합니다.
> DAG를 이해하면 "코드 한 줄"이 실제로 몇 stage와 task, shuffle로 나뉘는지 설명할 수 있습니다.

### 질문

Spark 코드는 왜 transformation을 쓸 때 바로 실행되지 않고 action에서 실행될까요?

### 직관

요리사가 "양파 썰기, 고기 굽기, 소스 만들기" 주문서를 받아도 손님이 실제 주문을 확정하기 전에는 전체 동선을 최적화할 수 있습니다. Spark도 action 전까지 계획을 쌓아 둡니다.

### 작은 예시

```python
rdd = sc.textFile("logs")
errors = rdd.filter(lambda line: "ERROR" in line)
counts = errors.map(lambda line: (line.split()[0], 1)).reduceByKey(lambda a, b: a + b)
counts.collect()
```

`collect()`가 action이므로 그때 job이 실행됩니다.

### 상태 이동 trace

```
textFile
  |
  v
filter        narrow dependency
  |
  v
map          narrow dependency
  |
  v
reduceByKey  wide dependency -> shuffle
  |
  v
collect      action triggers job
```

### 내부 메커니즘

transformation은 새로운 RDD/DataFrame plan을 만듭니다. action은 결과가 필요하므로 scheduler가 DAG를 stage로 나눕니다. DataFrame은 Catalyst optimizer와 Tungsten execution 같은 내부 최적화 경로를 가질 수 있습니다. RDD API는 더 직접적이지만 optimizer가 적고, DataFrame은 구조 정보를 사용해 더 많은 최적화를 할 수 있습니다.

### 실패 모드

- action이 여러 번 있으면 같은 lineage가 반복 실행될 수 있습니다.
- `collect()`로 큰 데이터를 driver에 가져오면 driver OOM이 발생합니다.
- closure에 큰 object를 캡처하면 task serialization 비용이 커집니다.

### 검증 방법

PySpark shell:

```python
rdd = sc.parallelize(range(10), 2)
mapped = rdd.map(lambda x: x * 2)
mapped.toDebugString()
mapped.collect()
```

PASS 신호:

- action 전에는 결과가 계산되지 않고, lineage/debug string으로 dependency를 볼 수 있습니다.

FAIL 신호:

- transformation 호출이 곧바로 cluster computation을 완료했다고 생각합니다.

### 면접식 되묻기

"RDD와 DataFrame 차이는?"에는 RDD는 typed/low-level partitioned data abstraction이고 DataFrame은 schema 기반 optimizer를 활용하는 higher-level API라고 답합니다. 꼬리 질문에서는 optimizer, serialization, user-defined function 비용을 설명합니다.

### 흔한 오해와 반례

오해: "DataFrame은 RDD와 전혀 다른 실행 시스템이다."

반례: API와 optimizer 계층은 다르지만 결국 Spark execution engine에서 partition, task, shuffle, memory, disk를 사용합니다.

### Active recall

- action이 없는 transformation chain은 왜 실행되지 않나요?
- DataFrame optimizer가 도와줄 수 없는 경우는 어떤 경우인가요?

## 3. Narrow/Wide Dependency, Stage, Shuffle

> narrow dependency는 child partition이 소수의 parent partition만 필요로 하므로 pipeline으로 처리하기 쉽습니다.
> wide dependency는 여러 parent partition의 data를 재분배해야 하므로 shuffle이 발생하고 stage 경계가 됩니다.
> shuffle은 network 전송만이 아니라 serialization, memory, local disk file, fetch, merge를 포함하는 비싼 상태 이동입니다.

### 질문

`map()`은 빠른데 `groupByKey()`는 왜 갑자기 느려질 수 있을까요?

### 직관

각자 자기 책상에서 줄을 긋는 일은 빠릅니다. 하지만 같은 나라별로 모든 종이를 다시 모아야 하면 사람들이 서로 종이를 주고받고, 임시 상자에 담고, 다시 정렬해야 합니다. 이것이 shuffle입니다.

### 작은 예시

```
input partitions:
p0: (KR,1), (US,1)
p1: (KR,1), (JP,1)
p2: (US,1)

groupBy country:
KR must collect from p0,p1
US must collect from p0,p2
JP from p1
```

### 상태 이동 trace

```
map-side task
  |
  v
write shuffle blocks by target partition
  |
  v
local disk / memory buffer
  |
  v
reduce-side task fetches blocks over network
  |
  v
merge/sort/aggregate
```

### 내부 메커니즘

Spark는 wide dependency 앞뒤를 stage로 나눕니다. map-side stage는 shuffle output을 만들고, reduce-side stage는 필요한 block을 fetch합니다. shuffle data는 executor local disk에 파일로 남을 수 있고, executor loss나 disk loss는 downstream stage 재시도와 parent recomputation을 유발할 수 있습니다.

### 실패 모드

- key skew로 한 reduce partition이 다른 partition보다 훨씬 커집니다.
- shuffle file fetch failure가 반복되면 stage retry가 늘어납니다.
- disk spill과 GC가 섞이면 CPU보다 I/O가 병목이 됩니다.

### 검증 방법

Spark UI:

```text
Stages tab -> Shuffle Read / Shuffle Write / Spill / Task Duration
```

PASS 신호:

- stage 경계가 shuffle 연산에서 생기고, task별 shuffle size와 duration 분포를 확인합니다.

FAIL 신호:

- job duration 평균만 보고 skew를 놓칩니다. task max/min과 percentile이 중요합니다.

### 면접식 되묻기

"wide dependency가 왜 비싼가요?"에는 data가 partition 경계를 넘어 재분배되고, network와 disk와 serialization을 함께 쓰기 때문이라고 답합니다. 꼬리 질문에서는 map-side combine, salting, partitioner 조정을 설명합니다.

### 흔한 오해와 반례

오해: "shuffle은 network 비용이다."

반례: shuffle은 network뿐 아니라 map output file, reduce fetch, memory buffer, spill, sort/merge, GC를 포함합니다.

### Active recall

- narrow dependency와 wide dependency를 partition trace로 비교해 보세요.
- Spark UI에서 shuffle bottleneck을 확인할 때 어떤 열을 보나요?

## 4. Partitioning and Skew

> Spark partition은 task 병렬성의 단위이고, partition 크기와 key 분포는 executor 자원 사용을 결정합니다.
> partition이 너무 적으면 core가 놀고, 너무 많으면 scheduling overhead가 늘며, skew가 있으면 가장 큰 partition 하나가 stage 전체를 붙잡습니다.
> Kafka partition과 Cassandra token range처럼 Spark partition도 scale-out의 단위이면서 성능 실패의 경계입니다.

### 질문

cluster core가 100개인데 task가 10개뿐이면 왜 느릴까요?

### 직관

일꾼이 100명인데 상자가 10개뿐이면 90명은 일할 수 없습니다. 반대로 상자가 100,000개이면 상자 나눠 주고 기록하는 관리 비용이 커집니다.

### 작은 예시

```
partition sizes:
p0 10MB
p1 11MB
p2 9MB
p3 900MB
```

p3 하나가 끝나지 않으면 stage가 끝나지 않습니다.

### 상태 이동 trace

```
input data
  |
  v
partitioner / file splits / shuffle partition count
  |
  v
tasks assigned to executors
  |
  +-- small tasks finish quickly
  +-- skewed task spills and runs long
  |
  v
stage waits for slowest task
```

### 내부 메커니즘

Spark의 task는 partition 하나 또는 일부를 처리합니다. input file split, `repartition`, `coalesce`, `spark.sql.shuffle.partitions`, custom partitioner 등이 partition count와 distribution을 바꿉니다. SQL/DataFrame에서는 adaptive query execution(AQE)이 일부 skew와 partition 조정을 도와줄 수 있지만 모든 workload를 자동 해결하지는 않습니다.

### 실패 모드

- `groupByKey`로 큰 iterable을 한 task에 모으면 OOM이 날 수 있습니다.
- `repartition(1)`은 모든 output을 한 task로 모아 병목을 만듭니다.
- 너무 작은 파일이 많으면 task scheduling과 listing overhead가 커집니다.

### 검증 방법

Spark UI:

```text
Stage detail -> Summary Metrics for Completed Tasks
Check min/median/max input size, duration, spill
```

PASS 신호:

- max task duration이 median보다 훨씬 크면 skew를 의심합니다.

FAIL 신호:

- executor 수만 늘리면 skewed partition도 빨라질 것이라고 가정합니다.

### 면접식 되묻기

"Spark skew를 어떻게 해결하나요?"에는 먼저 skew key를 확인하고, salting, pre-aggregation, broadcast join, partitioner 조정, AQE 설정, data model 변경을 병목에 맞춰 고릅니다.

### 흔한 오해와 반례

오해: "partition 수는 많을수록 좋다."

반례: partition 수가 너무 많으면 task overhead, scheduler delay, 작은 file 문제가 생깁니다. 데이터 크기와 cluster core, shuffle pattern에 맞춰야 합니다.

### Active recall

- 한 partition만 큰 경우 왜 stage 전체가 늦어지나요?
- Kafka hot partition과 Spark skewed partition의 공통점을 말해 보세요.

## 5. Lineage, Checkpointing, Fault Tolerance

> Spark는 lost partition을 원본 data와 transformation lineage로 다시 계산할 수 있어 fault tolerance를 제공합니다.
> 그러나 lineage가 너무 길거나 shuffle output이 반복 손실되거나 stateful streaming을 다룰 때 checkpoint가 필요합니다.
> checkpoint는 recovery 시작점을 줄이는 대신 reliable storage 비용과 일관성 경계를 요구합니다.

### 질문

Spark는 replica를 항상 여러 개 저장하지 않는데도 executor failure를 어떻게 견딜까요?

### 직관

요리 결과물을 복사해 두지 않아도 recipe와 재료 위치를 알고 있으면 다시 만들 수 있습니다. Spark lineage는 recipe에 가깝고, checkpoint는 중간 완성품을 냉장고에 넣어 두는 것에 가깝습니다.

### 작은 예시

```
RDD3 = RDD1.map(...).filter(...)
```

RDD3 partition 하나가 사라지면 Spark는 RDD1의 해당 partition에서 map/filter를 다시 실행할 수 있습니다.

### 상태 이동 trace

```
executor E1 cached partition P
  |
  x E1 lost
  |
  v
scheduler detects lost partition
  |
  v
trace lineage to parent partitions
  |
  v
reschedule tasks on another executor
  |
  v
recompute P
```

Checkpoint:

```
long lineage
  |
  v
write checkpoint to reliable storage
  |
  v
future recovery starts from checkpoint, not original lineage
```

### 내부 메커니즘

RDD는 parent RDD와 transformation 정보를 들고 있습니다. narrow dependency는 필요한 parent partition이 적어 recomputation이 쉽습니다. wide dependency는 shuffle output이 필요하고, shuffle file loss가 downstream recomputation을 유발합니다. Structured Streaming은 offset, state, sink progress를 checkpoint location에 남겨 restart할 수 있게 합니다. sink semantics에 따라 exactly-once 표현은 제한됩니다.

### 실패 모드

- nondeterministic transformation이나 외부 side effect가 있으면 recomputation이 같은 결과를 보장하지 않을 수 있습니다.
- checkpoint location이 local disk이면 executor/node failure에서 쓸 수 없습니다.
- 너무 긴 lineage는 failure recovery와 scheduler overhead를 키웁니다.

### 검증 방법

Spark UI와 logs:

```text
Jobs/Stages tab -> failed tasks, recomputation, locality, shuffle fetch failures
Streaming query -> checkpoint directory contents
```

PASS 신호:

- failed task가 retry되고 lost partition이 lineage로 recomputed됩니다.

FAIL 신호:

- recomputation이 외부 API side effect를 중복 실행해도 안전하다고 가정합니다.

### 면접식 되묻기

"Spark checkpoint와 cache 차이는?"에는 cache는 성능을 위해 partition을 memory/disk에 보관하는 것이고, checkpoint는 lineage를 끊고 reliable recovery point를 만들기 위한 것이라고 답합니다.

### 흔한 오해와 반례

오해: "lineage가 있으니 checkpoint는 필요 없다."

반례: long lineage, iterative algorithm, streaming state, nondeterminism, external source/sink에서는 checkpoint가 recovery 비용과 correctness boundary를 줄입니다.

### Active recall

- lost cached partition이 lineage로 복구되는 trace를 그려 보세요.
- checkpoint location이 reliable storage여야 하는 이유는 무엇인가요?

## 6. Memory, Caching, Spill, Performance Bottlenecks

> Spark executor memory는 execution memory와 storage memory가 경쟁하는 공간이고, JVM heap, off-heap, OS page cache, local disk spill이 함께 성능을 결정합니다.
> cache는 재사용을 빠르게 하지만 memory를 차지하고, spill은 OOM을 피하지만 disk I/O와 serialization 비용을 만듭니다.
> Spark 성능 튜닝은 "메모리를 늘린다"가 아니라 data layout, partition size, serialization, GC, shuffle, locality를 함께 맞추는 일입니다.

### 질문

Spark job이 spill을 시작하면 왜 갑자기 느려질까요?

### 직관

책상 위에서 정리하던 종이가 너무 많아 바닥과 창고에 내려놓기 시작하면, 매번 다시 꺼내고 정리하는 시간이 듭니다. spill은 memory 부족을 disk 왕복으로 바꾸는 신호입니다.

### 작은 예시

reduce task가 한 key group을 처리하려고 hash table을 만듭니다. task memory가 부족해 중간 data를 disk에 spill합니다.

### 상태 이동 trace

```
task reads shuffle blocks
  |
  v
deserialize records into memory
  |
  v
build aggregation/sort structure
  |
  | memory threshold exceeded
  v
spill batch to local disk
  |
  v
later merge spilled files
```

### 내부 메커니즘

Spark memory는 execution과 storage가 공유합니다. cache된 block은 storage memory를 쓰고, shuffle/sort/join은 execution memory를 씁니다. execution이 storage를 밀어낼 수 있고, storage level에 따라 memory에 못 들어간 partition은 recompute되거나 disk에 저장됩니다. Java object overhead와 GC도 중요합니다. serialized caching과 Kryo는 memory footprint와 network serialization 비용을 줄일 수 있습니다.

### 실패 모드

- `MEMORY_ONLY` cache가 memory에 못 들어가면 일부 partition이 매번 recompute됩니다.
- 큰 object graph는 GC pause를 키웁니다.
- spill disk가 느리면 task duration이 길어지고 stage tail이 늘어납니다.

### 검증 방법

Spark UI:

```text
Executors tab -> GC Time, Storage Memory, Shuffle Read/Write
Stage detail -> Spill (memory), Spill (disk)
Storage tab -> cached RDD/DataFrame size
```

PASS 신호:

- spill 증가, GC time 증가, slow task, disk I/O 증가가 같은 stage에서 연결됩니다.

FAIL 신호:

- executor memory만 무작정 키워 GC pause와 cluster utilization을 악화시킵니다.

### 면접식 되묻기

"Spark job이 느릴 때 어디를 보나요?"에는 DAG stage, task skew, shuffle size, spill, GC, data locality, input file count, executor lost를 순서대로 봅니다. OS 지표로 disk/network/CPU를 교차 확인합니다.

### 흔한 오해와 반례

오해: "Spark는 in-memory engine이므로 disk는 중요하지 않다."

반례: shuffle, spill, cache storage level, checkpoint, input/output 모두 disk나 remote storage를 쓸 수 있습니다. in-memory는 중요한 최적화 방향이지 disk-free 보장이 아닙니다.

### Active recall

- execution memory와 storage memory가 각각 어떤 작업에 쓰이나요?
- spill이 생긴 stage를 OS 관점에서 어떻게 확인하나요?

## 7. Streaming

> Spark Structured Streaming은 streaming input을 계속 들어오는 table처럼 보고, micro-batch 또는 지원되는 execution mode로 incremental query를 실행합니다.
> streaming correctness는 source offset, checkpoint, state store, sink commit semantics가 함께 맞아야 합니다.
> "exactly-once"는 source와 sink 조건에 따라 달라지므로, checkpoint만 있다고 외부 side effect까지 자동 보장된다고 말하면 안 됩니다.

### 질문

Spark streaming job이 restart되면 어디서부터 다시 읽을까요?

### 직관

책을 읽다가 책갈피와 필기장을 남겨 두면 다시 시작할 위치와 중간 계산 상태를 알 수 있습니다. Spark streaming checkpoint는 이 책갈피와 필기장 역할을 합니다.

### 작은 예시

Kafka source에서 offset 100까지 처리하고 checkpoint에 남깁니다. restart 후 Spark는 checkpoint를 읽고 다음 offset부터 처리를 재개하려고 합니다.

### 상태 이동 trace

```
micro-batch reads source offsets [90, 100]
  |
  v
process stateful aggregation
  |
  v
write sink output
  |
  v
commit progress/checkpoint
  |
  v
restart reads checkpoint and resumes
```

### 내부 메커니즘

Structured Streaming은 query plan과 state를 관리하고 checkpoint directory에 progress와 state store metadata를 남깁니다. Kafka source의 offset, stateful operator의 state, sink commit semantics가 맞아야 restart 후 중복/유실을 줄일 수 있습니다. 외부 sink가 idempotent하지 않으면 재시도 시 중복 side effect가 생길 수 있습니다.

### 실패 모드

- checkpoint directory를 삭제하면 streaming query가 처음부터 또는 잘못된 위치에서 시작할 수 있습니다.
- sink가 중복 write에 취약하면 retry/restart가 중복 결과를 만듭니다.
- state store가 커지면 checkpoint I/O와 memory pressure가 커집니다.

### 검증 방법

Structured Streaming local query를 실행할 때:

```text
checkpointLocation is set to durable path
query progress logs show inputRowsPerSecond, processedRowsPerSecond, batchId
```

PASS 신호:

- restart 후 batchId/offset progress가 checkpoint 기준으로 이어집니다.

FAIL 신호:

- checkpoint 없이 restart하면서 exactly-once를 기대합니다.

### 면접식 되묻기

"Structured Streaming은 exactly-once인가요?"에는 source, checkpoint, state, sink가 조건을 만족할 때 의미 있는 처리 보장을 구성할 수 있지만, 외부 sink와 side effect는 별도 idempotency가 필요하다고 답합니다.

### 흔한 오해와 반례

오해: "streaming은 batch와 완전히 다른 엔진이다."

반례: Structured Streaming은 Spark SQL engine 위에서 incremental query를 실행합니다. DAG, task, shuffle, state, checkpoint 문제는 여전히 중요합니다.

### Active recall

- streaming checkpoint에는 어떤 종류의 상태가 필요할까요?
- Kafka source + external DB sink에서 중복을 막으려면 무엇이 필요하나요?

## 현실 시나리오 1: Spark shuffle 때문에 작업이 느려진다

> Spark shuffle 병목은 네트워크 하나로 설명되지 않습니다.
> shuffle은 partition 재분배, local disk file, network fetch, memory pressure, spill, serialization, skew가 함께 만드는 비용입니다.

1. 관측된 증상

    특정 stage가 오래 걸리고, Spark UI에서 shuffle read와 disk spill이 큽니다.

2. 가능한 원인 후보

    key skew, partition 수 부족/과다, large shuffle block, slow disk, executor memory 부족, GC, network fetch failure.

3. OS/kernel 관점에서 볼 지점

    executor host `iostat`, `vmstat`, network retransmission, disk space, CPU steal/context switch.

4. distributed-system 관점에서 볼 지점

    data redistribution, slowest task tail, executor failure retry, locality.

5. Spark 내부 구조와 연결되는 지점

    wide dependency, stage boundary, shuffle map output, reduce fetch, spill metrics.

6. 확인 명령 또는 로그

    Spark UI stage detail, executor logs, `iostat -xz 1`, `jcmd <executor-pid> GC.heap_info`.

7. 잘못된 결론의 예

    "network bandwidth를 늘리면 해결된다."

8. 더 나은 추론 과정

    stage detail에서 task duration 분포와 shuffle read size를 봅니다. skew면 key salting이나 pre-aggregation, memory/spill이면 partition 조정과 serialization/cache 전략, disk면 local disk와 spill path를 봅니다.

## 현실 시나리오 2: executor memory pressure와 GC가 cluster 증상으로 나타난다

> Spark executor는 JVM process이므로 heap object 수, serialized data, GC, OS memory pressure가 task latency와 executor lost로 보입니다.
> GC pause는 분산 시스템 관점에서는 느린 노드나 사라진 노드처럼 관측될 수 있습니다.

1. 관측된 증상

    executor lost가 늘고 task retry가 증가하며 일부 stage에서 GC time이 높습니다.

2. 가능한 원인 후보

    cache 과다, large shuffle aggregation, skewed partition, inefficient serialization, too many cores per executor, insufficient overhead memory.

3. OS/kernel 관점에서 볼 지점

    process RSS, swap, page fault, CPU run queue, disk spill, cgroup memory limit.

4. distributed-system 관점에서 볼 지점

    executor heartbeat timeout, task retry, lost shuffle output, recomputation.

5. Spark 내부 구조와 연결되는 지점

    execution/storage memory competition, GC logs, spill, lineage recomputation, checkpoint.

6. 확인 명령 또는 로그

    Spark UI Executors tab, GC logs, `jcmd <pid> GC.heap_info`, `vmstat 1`, container memory events.

7. 잘못된 결론의 예

    "executor memory를 크게 늘리면 무조건 해결된다."

8. 더 나은 추론 과정

    memory가 부족한지 object가 너무 많은지 구분합니다. partition size, serialization, cache storage level, executor cores, GC pressure를 같이 조정합니다.

## 근거와 더 읽을 자료

- [10_source_ledger.md](10_source_ledger.md)의 Spark 4.1.2 docs, RDD guide, cluster overview, tuning, monitoring, RDD paper 항목
- Spark docs: https://spark.apache.org/docs/latest/
- Spark cluster overview: https://spark.apache.org/docs/latest/cluster-overview.html
- Spark RDD programming guide: https://spark.apache.org/docs/latest/rdd-programming-guide.html
- Spark tuning guide: https://spark.apache.org/docs/latest/tuning.html
- Spark RDD paper: https://people.csail.mit.edu/matei/papers/2012/nsdi_spark.pdf

