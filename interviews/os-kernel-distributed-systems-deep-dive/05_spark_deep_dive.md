# 05. Spark Deep Dive

Spark는 "메모리에서 돌아서 빠른 MapReduce 대체재" 정도로 외우면 shuffle, spill, skew, checkpoint 질문에서 막힙니다. Spark의 중심은 큰 계산을 partition 단위의 task로 나누고, dependency graph를 보고 어느 부분을 병렬 실행할지, 어디서 shuffle이 필요한지, 실패하면 어떤 partition을 다시 계산할지 결정하는 구조입니다.

이 문서는 `count()` 같은 action 하나가 driver, DAG, stage, task, executor memory, shuffle file, lineage, checkpoint로 바뀌는 흐름을 따라갑니다.

## 1. Driver와 executor는 역할이 다르다

Spark application은 driver program과 executor process들의 집합입니다. driver는 사용자의 main function을 실행하고 SparkContext/SparkSession을 만들며, job을 stage와 task로 나누어 scheduling합니다. executor는 worker node에서 떠 있는 process로, task를 실행하고 data를 memory나 disk에 저장합니다.

```
user code in driver
  -> builds logical transformations
  -> action triggers a job
  -> driver creates stages and tasks
  -> cluster manager provides executors
  -> executors run tasks
  -> results / shuffle blocks / metrics return
```

driver는 단순한 client가 아닙니다. task scheduling, metadata, executor heartbeat, result collection을 담당하므로 driver가 느리거나 멀리 있으면 전체 job이 영향을 받습니다. executor는 단순 worker thread가 아닙니다. JVM process로서 heap, off-heap, thread pool, local disk, network connection을 갖고, task 실행과 cache, shuffle block storage를 담당합니다.

## 2. Transformation은 계획이고 action은 실행을 시작한다

Spark에서 `map`, `filter`, `join` 같은 transformation은 즉시 cluster에서 실행되지 않습니다. dependency graph를 쌓습니다. `count`, `collect`, `write` 같은 action이 호출될 때 driver가 이 graph를 job으로 materialize합니다. 이 lazy evaluation 덕분에 Spark는 여러 transformation을 합치고 stage를 나눌 수 있습니다.

```
val a = input.filter(...)
val b = a.map(...)
val c = b.reduceByKey(...)

until action:
  driver holds lineage graph

on action:
  driver analyzes dependencies
  -> stage 1: read/filter/map
  -> shuffle boundary
  -> stage 2: reduceByKey result
```

이 구조를 모르면 "왜 내 코드가 여기서 실행되지 않지?"라는 혼란이 생깁니다. transformation 안의 함수는 driver에서 한 번 실행되는 것이 아니라 serialized closure로 executor에 보내져 task 안에서 실행됩니다. closure가 너무 크거나 non-serializable object를 잡으면 task 시작 전부터 문제가 생길 수 있습니다.

DataFrame을 쓴다고 이 원리가 사라지는 것은 아닙니다. RDD는 partitioned data와 transformation lineage를 개발자가 비교적 직접 다루는 API이고, DataFrame/Dataset은 row와 column, schema, expression을 중심으로 더 높은 수준의 logical plan을 만듭니다. Spark SQL의 Catalyst optimizer는 이 logical plan을 분석하고 더 나은 physical plan 후보로 바꿉니다. 하지만 action이 호출되어 job이 생기고, physical plan이 stage와 task로 쪼개지고, executor가 partition을 처리하며, shuffle boundary에서 data가 다시 배치된다는 lower-layer 흐름은 그대로 남습니다.

```
DataFrame code
  -> unresolved logical plan
  -> analyzed logical plan with schema
  -> optimized logical plan
  -> physical plan
  -> jobs / stages / tasks
  -> executor CPU, memory, shuffle, spill
```

따라서 DataFrame 성능 질문도 "Catalyst가 알아서 해요"로 끝나지 않습니다. optimizer가 join order, predicate pushdown, projection pruning, aggregation strategy를 바꿀 수는 있지만, 결국 partition 수, data skew, shuffle size, memory/spill, source/sink I/O가 실제 실행 시간을 결정합니다. explain plan은 이 bridge를 보는 도구입니다. 논리 계획이 어떤 physical operator와 exchange(shuffle)로 바뀌었는지 보아야 DAG와 stage가 왜 그렇게 생겼는지 설명할 수 있습니다.

## 3. Partition과 task는 Spark 병렬성의 기본 단위다

RDD나 DataFrame은 partition으로 나뉩니다. 각 partition은 보통 하나의 task가 처리합니다. partition 수가 적으면 cluster resource를 충분히 쓰지 못합니다. partition 수가 너무 많으면 task scheduling overhead와 metadata가 커지고 작은 file 문제가 생길 수 있습니다.

```
dataset
  partition 0 -> task 0 on executor A
  partition 1 -> task 1 on executor B
  partition 2 -> task 2 on executor A
```

partition은 data locality와도 연결됩니다. 입력 data가 HDFS나 object storage, local disk에 있을 때 Spark는 가능하면 data가 가까운 executor에서 task를 실행하려 합니다. 하지만 cluster resource, executor availability, dynamic allocation 때문에 항상 local하게 실행되지는 않습니다.

Kafka와 Cassandra에서 partition이 data ownership과 ordering의 경계였다면, Spark에서 partition은 계산 병렬성과 shuffle 비용의 경계입니다. 같은 단어라도 시스템마다 지키는 약속이 다릅니다.

## 4. Narrow dependency와 wide dependency가 stage를 가른다

dependency는 어떤 partition을 만들기 위해 어떤 parent partition이 필요한지를 말합니다. narrow dependency는 child partition 하나가 적은 수의 parent partition에만 의존하는 경우입니다. `map`과 `filter`가 대표적입니다. wide dependency는 여러 parent partition의 data를 다시 섞어야 하는 경우입니다. `groupByKey`, `reduceByKey`, `join`, `repartition`이 대표적이며 보통 shuffle을 만듭니다.

```
narrow:
  parent p0 -> child p0
  parent p1 -> child p1

wide / shuffle:
  parent p0, p1, p2
        \  |  /
       hash by key
        /  |  \
  child q0, q1, q2
```

stage는 shuffle boundary를 기준으로 나뉩니다. shuffle 전 stage는 map output을 만들고, shuffle 후 stage는 key나 partitioner에 따라 다시 모인 data를 읽습니다. 그래서 Spark job이 느릴 때 "stage 몇 번이 느린가", "shuffle read/write가 많은가", "특정 task만 오래 걸리는가"를 먼저 봅니다.

## 5. Shuffle은 network만이 아니라 serialization, disk, memory 문제다

shuffle은 data를 key나 partitioner 기준으로 다시 배치하는 일입니다. 이 과정에서 executor는 map output을 serialization해 local disk에 쓰고, reduce task는 필요한 block을 network로 fetch합니다. memory가 부족하면 sort/aggregate 중간 data가 spill file로 내려갑니다.

```
map task
  -> compute records
  -> serialize
  -> partition by target reducer
  -> maybe spill to local disk
  -> write shuffle blocks

reduce task
  -> fetch remote/local shuffle blocks
  -> deserialize
  -> merge / aggregate / sort
  -> maybe spill again
```

따라서 "shuffle이 느리다"는 말은 network가 느리다는 뜻만이 아닙니다. key skew로 한 reduce partition만 커질 수 있고, serialization 비용이 클 수 있으며, executor memory가 부족해 spill이 많을 수 있고, local disk가 느릴 수 있으며, GC가 늘어날 수 있습니다. OS 관점에서는 executor JVM heap, off-heap buffer, page cache, local disk I/O, TCP connection이 모두 관련됩니다.

## 6. Cache와 persist는 recomputation을 줄이지만 memory 압력을 만든다

Spark lineage 덕분에 RDD/DataFrame partition은 필요할 때 다시 계산될 수 있습니다. 같은 dataset을 여러 action에서 반복 쓰면 매번 다시 계산하지 않도록 cache/persist를 사용할 수 있습니다. memory에 올리면 빠르지만, memory가 부족하면 eviction이나 spill, GC pressure가 생깁니다. disk persistence를 선택하면 recomputation은 줄지만 disk I/O가 늘어납니다.

```
without cache:
  action 1 -> read -> transform -> result
  action 2 -> read -> transform -> result again

with cache:
  action 1 -> read -> transform -> store partitions
  action 2 -> reuse stored partitions if still available
```

cache는 correctness 도구가 아니라 performance 도구입니다. executor가 죽어 cache가 사라지면 Spark는 lineage로 다시 계산할 수 있습니다. 그러나 source가 nondeterministic하거나 외부 side effect가 있으면 재계산 결과가 달라질 수 있습니다. 이때 checkpoint가 필요합니다.

## 7. Lineage와 checkpoint는 실패 후 어디서 다시 시작할지 정한다

Spark의 RDD가 resilient하다고 말하는 이유는 partition을 만드는 transformation lineage를 알고 있기 때문입니다. executor가 죽어 partition 하나를 잃으면 Spark는 그 partition을 만드는 데 필요한 parent partition부터 다시 계산할 수 있습니다.

```
lost partition q1
  -> driver looks at lineage
  -> q1 came from shuffle output of stage 1
  -> if shuffle output exists, refetch
  -> if not, recompute needed map partitions
```

하지만 lineage가 길어지면 failure recovery 비용이 커지고 driver metadata도 늘어납니다. streaming job이나 iterative algorithm에서는 checkpoint로 lineage를 끊고 안정적인 storage에 중간 상태를 저장합니다. checkpoint는 정상 실행 비용을 늘리지만 장애 회복 경로를 짧게 만듭니다.

Spark checkpoint는 Kafka offset이나 Cassandra commit log와 같은 물건은 아니지만, "실패 후 다시 시작할 기준점"이라는 질문에 답한다는 점에서 같은 계열입니다.

## 8. Structured Streaming은 checkpoint와 sink semantics가 핵심이다

Structured Streaming은 무한히 들어오는 data를 micro-batch 또는 continuous 처리 모델로 다룹니다. 여기서 중요한 것은 source offset, state store, checkpoint, sink write의 관계입니다. Kafka에서 읽어 aggregation하고 외부 sink에 쓰는 경우, Spark가 source offset과 state를 어디까지 checkpoint했는지, sink가 idempotent한지에 따라 중복과 유실 경계가 달라집니다.

```
read source offsets
  -> process batch
  -> update state
  -> write sink
  -> checkpoint progress
```

장애가 이 네 단계 사이 어디서 나느냐에 따라 같은 batch가 다시 처리될 수 있습니다. 그래서 streaming exactly-once를 말할 때는 Spark 내부 state와 checkpoint뿐 아니라 sink의 idempotency나 transaction support를 함께 말해야 합니다.

## 현실 시나리오 1: join이 갑자기 느려졌다

큰 table과 작은 table을 join한다고 합시다. 작은 table이 충분히 작으면 broadcast join으로 각 executor에 복사해 network shuffle을 줄일 수 있습니다. 하지만 작은 table이 예상보다 크거나 statistics가 틀리면 broadcast가 memory pressure를 만들 수 있고, 일반 shuffle join은 key skew 때문에 특정 reducer가 오래 걸릴 수 있습니다.

보는 순서는 다음과 같습니다.

```
Spark UI stage
  -> task duration distribution
  -> shuffle read/write size
  -> spill memory/disk
  -> GC time
  -> skewed partition or executor loss
```

해결도 하나가 아닙니다. partitioning 조정, salting, broadcast threshold 조정, input file size 개선, executor memory 조정, serializer 개선이 모두 후보입니다. "executor를 늘린다"는 partition skew나 single reducer 병목에는 효과가 제한적입니다.

## 현실 시나리오 2: executor memory를 늘렸는데 더 느려졌다

memory를 늘리면 spill이 줄 수 있지만 GC pause가 커질 수도 있습니다. executor heap이 커지고 object가 많아지면 GC가 오래 걸립니다. off-heap memory, shuffle buffer, OS page cache, container memory limit까지 함께 봐야 합니다.

Spark tuning에서 memory는 storage와 execution이 나눠 쓰는 자원입니다. cache가 너무 많은 memory를 차지하면 execution memory가 부족해 spill이 늘 수 있고, execution이 memory를 많이 쓰면 cache가 evict될 수 있습니다. OS 관점에서는 local disk spill과 page cache도 같이 움직입니다.

## 문서를 덮고 확인할 것

- transformation과 action의 차이를 lazy evaluation 관점에서 설명해 보세요.
- narrow dependency와 wide dependency가 stage boundary를 만드는 이유를 말해 보세요.
- shuffle이 network뿐 아니라 disk, memory, serialization, skew 문제인 이유를 trace로 설명해 보세요.
- lineage와 checkpoint의 차이를 "재계산 가능성"과 "재시작 기준점"으로 설명해 보세요.
- Spark streaming에서 source offset, checkpoint, sink idempotency가 왜 함께 필요한지 말해 보세요.

## 근거와 더 읽을 자료

- Apache Spark cluster mode overview: driver, executor, task, stage.
- Apache Spark RDD programming guide: partition, transformation/action, shuffle, persistence.
- Apache Spark tuning guide: serialization, memory management, GC, parallelism.
- Spark RDD paper: lineage-based fault tolerance.
