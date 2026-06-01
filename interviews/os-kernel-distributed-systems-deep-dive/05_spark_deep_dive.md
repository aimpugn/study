# 05. Spark Deep Dive

## 목차

- [1. Driver와 executor는 역할이 다르다](#1-driver와-executor는-역할이-다르다)
- [2. Transformation은 계획이고 action은 실행을 시작한다](#2-transformation은-계획이고-action은-실행을-시작한다)
- [3. Partition과 task는 Spark 병렬성의 기본 단위다](#3-partition과-task는-spark-병렬성의-기본-단위다)
- [4. Narrow dependency와 wide dependency가 stage를 가른다](#4-narrow-dependency와-wide-dependency가-stage를-가른다)
- [5. Shuffle은 network만이 아니라 serialization, disk, memory 문제다](#5-shuffle은-network만이-아니라-serialization-disk-memory-문제다)
- [6. Cache와 persist는 recomputation을 줄이지만 memory 압력을 만든다](#6-cache와-persist는-recomputation을-줄이지만-memory-압력을-만든다)
- [7. Lineage와 checkpoint는 실패 후 어디서 다시 시작할지 정한다](#7-lineage와-checkpoint는-실패-후-어디서-다시-시작할지-정한다)
- [8. Structured Streaming은 checkpoint와 sink semantics가 핵심이다](#8-structured-streaming은-checkpoint와-sink-semantics가-핵심이다)
- [현실 시나리오 1: join이 갑자기 느려졌다](#현실-시나리오-1-join이-갑자기-느려졌다)
- [현실 시나리오 2: executor memory를 늘렸는데 더 느려졌다](#현실-시나리오-2-executor-memory를-늘렸는데-더-느려졌다)
- [문서를 덮고 확인할 것](#문서를-덮고-확인할-것)
- [근거와 더 읽을 자료](#근거와-더-읽을-자료)
- [Driver, executor, task를 OS process와 thread로 다시 보기](#driver-executor-task를-os-process와-thread로-다시-보기)
- [Lazy evaluation과 stage boundary는 lineage를 실행 가능한 조각으로 자른다](#lazy-evaluation과-stage-boundary는-lineage를-실행-가능한-조각으로-자른다)
- [Shuffle은 network가 아니라 disk, memory, serialization, skew가 만나는 지점이다](#shuffle은-network가-아니라-disk-memory-serialization-skew가-만나는-지점이다)
- [Executor memory는 heap 하나가 아니라 여러 영역이다](#executor-memory는-heap-하나가-아니라-여러-영역이다)
- [Structured Streaming은 offset, state, sink commit이 함께 안전해야 한다](#structured-streaming은-offset-state-sink-commit이-함께-안전해야-한다)
- [Scheduler delay, straggler, speculative execution](#scheduler-delay-straggler-speculative-execution)
- [OS와 Spark UI를 함께 읽는 map](#os와-spark-ui를-함께-읽는-map)
- [Interview replay: Spark를 계산 graph와 OS resource로 설명하기](#interview-replay-spark를-계산-graph와-os-resource로-설명하기)
- [Data locality와 remote fetch는 scheduler 판단을 바꾼다](#data-locality와-remote-fetch는-scheduler-판단을-바꾼다)
- [External sink와 side effect는 task retry와 충돌한다](#external-sink와-side-effect는-task-retry와-충돌한다)
- [Spark와 운영체제의 마지막 연결](#spark와-운영체제의-마지막-연결)

Spark는 "메모리에서 돌아서 빠른 MapReduce 대체재" 정도로 외우면 shuffle, spill, skew, checkpoint 질문에서 막힙니다. Spark의 중심은 큰 계산을 partition 단위의 task로 나누고, dependency graph를 보고 어느 부분을 병렬 실행할지, 어디서 shuffle이 필요한지, 실패하면 어떤 partition을 다시 계산할지 결정하는 구조입니다.

이 문서는 `count()` 같은 action 하나가 driver, DAG, stage, task, executor memory, shuffle file, lineage, checkpoint로 바뀌는 흐름을 따라갑니다.

## 1. Driver와 executor는 역할이 다르다

Spark application은 driver program과 executor process들의 집합입니다. driver는 사용자의 main function을 실행하고 SparkContext/SparkSession을 만들며, job을 stage와 task로 나누어 scheduling합니다. executor는 worker node에서 떠 있는 process로, task를 실행하고 data를 memory나 disk에 저장합니다.

```text
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

```text
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

```text
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

```text
dataset
  partition 0 -> task 0 on executor A
  partition 1 -> task 1 on executor B
  partition 2 -> task 2 on executor A
```

partition은 data locality와도 연결됩니다. 입력 data가 HDFS나 object storage, local disk에 있을 때 Spark는 가능하면 data가 가까운 executor에서 task를 실행하려 합니다. 하지만 cluster resource, executor availability, dynamic allocation 때문에 항상 local하게 실행되지는 않습니다.

Kafka와 Cassandra에서 partition이 data ownership과 ordering의 경계였다면, Spark에서 partition은 계산 병렬성과 shuffle 비용의 경계입니다. 같은 단어라도 시스템마다 지키는 약속이 다릅니다.

## 4. Narrow dependency와 wide dependency가 stage를 가른다

dependency는 어떤 partition을 만들기 위해 어떤 parent partition이 필요한지를 말합니다. narrow dependency는 child partition 하나가 적은 수의 parent partition에만 의존하는 경우입니다. `map`과 `filter`가 대표적입니다. wide dependency는 여러 parent partition의 data를 다시 섞어야 하는 경우입니다. `groupByKey`, `reduceByKey`, `join`, `repartition`이 대표적이며 보통 shuffle을 만듭니다.

```text
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

```text
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

[01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md)의 network path를 Spark에 대입하면 reduce task의 shuffle fetch는 remote executor의 file read와 socket send, local executor의 socket receive queue와 TCP reassembly, deserialization, memory allocation을 모두 포함합니다. fetch failure가 "네트워크가 끊겼다" 하나로 닫히지 않는 이유입니다.

```text
reduce task fetches shuffle block
  -> remote executor reads local shuffle file
  -> page cache or disk read
  -> socket send buffer
  -> TCP network path
  -> local socket receive queue
  -> deserialize and merge
  -> spill again if memory is insufficient
```

## 6. Cache와 persist는 recomputation을 줄이지만 memory 압력을 만든다

Spark lineage 덕분에 RDD/DataFrame partition은 필요할 때 다시 계산될 수 있습니다. 같은 dataset을 여러 action에서 반복 쓰면 매번 다시 계산하지 않도록 cache/persist를 사용할 수 있습니다. memory에 올리면 빠르지만, memory가 부족하면 eviction이나 spill, GC pressure가 생깁니다. disk persistence를 선택하면 recomputation은 줄지만 disk I/O가 늘어납니다.

```text
without cache:
  action 1 -> read -> transform -> result
  action 2 -> read -> transform -> result again

with cache:
  action 1 -> read -> transform -> store partitions
  action 2 -> reuse stored partitions if still available
```

cache는 correctness 도구가 아니라 performance 도구입니다. executor가 죽어 cache가 사라지면 Spark는 lineage로 다시 계산할 수 있습니다. 그러나 source가 nondeterministic하거나 외부 side effect가 있으면 재계산 결과가 달라질 수 있습니다. 이때 checkpoint가 필요합니다.

메모리 관점에서는 [01b_memory_and_address_space.md](01b_memory_and_address_space.md)의 구분이 그대로 필요합니다. Spark executor heap 안의 storage/execution memory, off-heap buffer, JVM thread stack, Python worker memory, mmap/page cache, container overhead가 같은 memory limit 안에서 경쟁합니다. heap을 늘렸는데 오히려 느려지는 이유는 GC pause 증가, page cache 축소, spill pattern 변화, cgroup OOM risk가 같이 움직이기 때문입니다.

## 7. Lineage와 checkpoint는 실패 후 어디서 다시 시작할지 정한다

Spark의 RDD가 resilient하다고 말하는 이유는 partition을 만드는 transformation lineage를 알고 있기 때문입니다. executor가 죽어 partition 하나를 잃으면 Spark는 그 partition을 만드는 데 필요한 parent partition부터 다시 계산할 수 있습니다.

```text
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

```text
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

```text
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

container에서 executor를 실행한다면 [01e_concurrency_isolation_observability.md](01e_concurrency_isolation_observability.md)의 cgroup memory도 같이 확인해야 합니다. Spark 설정의 executor memory는 전체 process resident memory를 뜻하지 않습니다. overhead memory, off-heap, Python worker, native library, page cache까지 더해 cgroup limit을 넘으면 JVM OutOfMemoryError가 아니라 container OOM kill로 보일 수 있습니다.

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

## Driver, executor, task를 OS process와 thread로 다시 보기

Spark application은 driver와 executor로 구성됩니다. Driver는 user program의 control flow를 실행하고 job, stage, task를 만들며 cluster manager와 통신합니다. Executor는 worker node에서 task를 실행하고 data를 cache하거나 shuffle block을 저장합니다. 이 구조를 Spark 용어로만 보면 추상적이지만, OS 관점으로 보면 driver와 executor는 process이고, executor 안의 task는 thread pool 위에서 실행되는 work item입니다.

```text
driver process
  -> builds logical/physical plan
  -> submits jobs and stages
  -> tracks task attempts and executor heartbeats

executor process
  -> JVM heap/off-heap/thread stacks
  -> task threads
  -> shuffle files and cache blocks
  -> network connections to driver and other executors
```

Driver가 느리면 scheduling delay가 생기고, executor heartbeat를 제때 처리하지 못하면 executor loss로 오판할 수 있습니다. Executor가 느리면 task가 길어지고, shuffle block을 늦게 제공하고, heartbeat가 지연됩니다. Cluster manager는 resource allocation과 process lifecycle을 담당하지만, task 내부의 CPU, memory, disk, network 경쟁은 OS와 JVM 위에서 일어납니다. 따라서 Spark tuning은 Spark config만의 문제가 아니라 process scheduling, memory accounting, local disk, network socket, cgroup boundary와 연결됩니다.

## Lazy evaluation과 stage boundary는 lineage를 실행 가능한 조각으로 자른다

Spark transformation은 즉시 실행되지 않고 lineage graph를 만듭니다. Action이 호출되면 Spark는 필요한 computation graph를 job으로 만들고, shuffle이 필요한 wide dependency를 기준으로 stage를 나눕니다. Narrow dependency는 parent partition 하나 또는 소수에서 child partition을 만들 수 있어 pipeline으로 이어질 수 있습니다. Wide dependency는 data를 key 기준으로 다시 나누어 shuffle을 만들고 stage boundary가 됩니다.

```text
read input
  -> map
  -> filter
  -> reduceByKey
  -> write output

stage 1
  read/map/filter
  -> shuffle write

stage 2
  shuffle read
  -> reduce/write
```

이 구조는 recovery와도 연결됩니다. 어떤 partition이 실패하면 lineage를 따라 필요한 parent partition을 다시 계산할 수 있습니다. 하지만 lineage가 너무 길거나 source가 비결정적이거나 shuffle output이 사라지면 재계산 비용이 커집니다. Checkpoint는 lineage를 끊고 안정된 storage에 중간 상태를 남깁니다. Cache/persist는 반복 계산을 줄이지만 executor memory를 사용하고 eviction될 수 있습니다. "Spark는 memory computation이라 빠르다"가 아니라, "계산 graph와 materialization boundary를 선택해 재계산, memory, disk I/O를 조절한다"가 더 정확합니다.

## Shuffle은 network가 아니라 disk, memory, serialization, skew가 만나는 지점이다

Shuffle은 data를 partition 사이에서 재분배하는 과정입니다. Map-side task는 shuffle data를 local disk에 쓰고, reduce-side task는 필요한 block을 여러 executor에서 fetch합니다. 이때 serialization, compression, spill, file consolidation, network fetch, disk read/write, memory buffer가 모두 관여합니다. 그래서 shuffle이 느리다는 말은 "네트워크가 느리다"로 끝나지 않습니다.

```text
map task
  -> builds shuffle buffers
  -> spills to local disk if memory pressure
  -> writes shuffle files and index

reduce task
  -> fetches remote/local blocks
  -> stores buffers in memory or spills
  -> merges and aggregates
```

Skew는 특정 key나 partition에 data가 몰리는 현상입니다. 평균 partition size가 좋아 보여도 한 partition이 매우 크면 그 task가 stage 전체를 붙잡습니다. Executor를 더 늘려도 single hot partition은 한 task가 처리해야 할 수 있습니다. 해결은 salting, repartitioning, skew join optimization, broadcast join, data modeling 조정 등 workload에 따라 달라집니다. 단순히 executor 수를 늘리는 것은 병렬화 가능한 조각이 충분할 때만 효과가 있습니다.

Shuffle file은 OS local filesystem과 page cache, disk queue를 사용합니다. Executor가 container 안에서 실행되면 local directory가 ephemeral storage인지, network volume인지, SSD인지, quota가 있는지에 따라 성능과 안정성이 달라집니다. Disk full은 stage failure로 올라오고, slow disk는 fetch wait와 task runtime으로 올라옵니다. Page cache가 충분하면 repeated read에 도움이 될 수 있지만, executor heap을 너무 크게 잡으면 OS cache 여유가 줄 수 있습니다.

## Executor memory는 heap 하나가 아니라 여러 영역이다

Spark executor memory를 볼 때 `spark.executor.memory`만 보면 부족합니다. Execution memory는 shuffle, join, aggregation 같은 계산 중간 구조에 쓰이고, storage memory는 cache/persist된 block에 쓰입니다. Unified memory manager는 이 둘 사이를 조절합니다. 그 밖에 off-heap memory, direct buffer, serializer buffer, Python worker memory, JVM metaspace, thread stack, code cache, OS page cache, container overhead가 있습니다.

```text
executor container limit
  -> JVM heap
     -> execution memory
     -> storage memory
     -> user objects
  -> memory overhead
     -> off-heap/direct buffers
     -> Python worker/native memory
     -> thread stacks/metaspace
  -> local disk page cache and kernel memory
```

Executor heap을 늘리면 spill이 줄 수 있지만 GC pause가 늘고 page cache가 줄 수 있습니다. Heap을 줄이면 GC는 가벼워질 수 있지만 spill이 늘어 disk I/O가 커질 수 있습니다. Off-heap을 켜면 GC pressure를 줄일 수 있지만 cgroup memory와 native allocation을 더 잘 관측해야 합니다. PySpark는 Python worker process memory를 별도로 고려해야 합니다. Container OOM kill은 JVM heap exception과 다르게 process가 갑자기 사라질 수 있습니다.

Spark UI에서 memory spill과 disk spill, GC time, executor lost reason, task deserialization time, shuffle read blocked time을 함께 봐야 합니다. OS에서는 cgroup memory events, disk I/O, network socket, GC logs, executor process RSS를 봅니다. "메모리를 늘렸는데 더 느려졌다"는 질문은 이 tradeoff를 이해하고 있을 때만 답할 수 있습니다.

## Structured Streaming은 offset, state, sink commit이 함께 안전해야 한다

Structured Streaming은 source에서 data를 읽고, stateful operator를 처리하고, sink에 결과를 씁니다. 장애 뒤 정확한 재시작을 하려면 checkpoint에 source offset, state store, progress metadata가 남아야 합니다. 하지만 sink side effect가 idempotent하지 않거나 transaction boundary가 맞지 않으면 중복 output이 생길 수 있습니다. Kafka source와 Kafka sink를 쓰는 경우와 외부 DB나 object storage sink를 쓰는 경우의 보장이 다릅니다.

```text
micro-batch
  -> read source offsets [a, b)
  -> update state store
  -> write sink output
  -> record progress in checkpoint

crash between sink write and checkpoint
  -> restart may reprocess batch
  -> sink must handle duplicate or commit protocol must prevent it
```

Checkpoint는 단순한 임시 파일이 아닙니다. Streaming query identity와 recovery boundary입니다. Checkpoint directory를 지우면 query는 과거 진행 상태를 잃습니다. Object storage 위 checkpoint는 listing/rename/commit semantics와 latency를 고려해야 합니다. State store가 커지면 memory와 local disk, checkpoint I/O가 늘고, watermark와 state eviction 정책이 중요해집니다.

## Scheduler delay, straggler, speculative execution

Spark UI에서 task time을 볼 때 scheduler delay, task deserialization, executor run time, GC time, result serialization, getting result time 같은 항목을 나눠 봐야 합니다. Scheduler delay가 크면 resource 부족, task locality wait, driver bottleneck, executor launch delay를 의심합니다. Executor run time이 길면 skew, CPU, I/O, GC, remote fetch를 봅니다. GC time이 크면 object allocation과 memory layout을 봅니다.

Straggler는 일부 task가 다른 task보다 훨씬 늦는 현상입니다. 원인은 data skew, bad executor, slow disk, network congestion, GC, noisy neighbor, node-level throttling 등 다양합니다. Speculative execution은 느린 task의 복제본을 다른 executor에서 실행해 먼저 끝나는 것을 사용하려는 방법입니다. 하지만 side effect가 있는 task나 shuffle/storage pressure가 큰 workload에서는 추가 load를 만들 수 있습니다. 느린 이유를 모른 채 speculation을 켜면 cluster를 더 바쁘게 만들 수 있습니다.

```text
stage has 100 tasks
  -> 99 tasks finish in 20s
  -> 1 task runs for 5min
  -> job waits for the last task

possible causes
  skewed partition
  executor GC
  slow disk
  network fetch wait
  CPU throttling
```

## OS와 Spark UI를 함께 읽는 map

Spark UI는 Spark 내부의 job/stage/task 관측을 제공합니다. 하지만 OS 자원 원인을 직접 확정하지는 않습니다. Shuffle read blocked time이 높으면 network fetch와 remote executor, local disk, connection pool을 봅니다. Disk spill이 크면 executor memory와 aggregation/join strategy, local disk 성능을 봅니다. GC time이 높으면 object allocation, serializer, cache format, heap size를 봅니다. Executor lost가 OOM이면 cgroup memory와 overhead를 봅니다.

```text
Spark symptom
  high scheduler delay
    -> driver/cluster resource/locality

  high shuffle read wait
    -> network, remote block, disk, skew

  high spill
    -> memory pressure, aggregation/join, partition size

  executor lost
    -> OOM, heartbeat timeout, node/preemption, disk failure
```

Kafka와 Cassandra와 비교하면 Spark는 long-running storage service라기보다 distributed compute runtime입니다. 하지만 OS 자원은 똑같이 중요합니다. Executor는 process이고, task는 thread이며, shuffle file은 filesystem과 disk를 쓰고, network fetch는 TCP와 socket buffer를 지납니다. Driver와 executor 사이의 heartbeat도 partial failure와 timeout 불확실성을 갖습니다.

## Interview replay: Spark를 계산 graph와 OS resource로 설명하기

Spark를 짧게 설명하면 "driver가 lazy transformation graph를 action 시점에 job/stage/task로 나누고, executor가 partition 단위 task를 실행하며, wide dependency에서 shuffle을 통해 data를 재분배하는 분산 계산 엔진"입니다. OS와 연결하면 "executor는 JVM process이고 task는 thread pool 위에서 CPU, heap/off-heap memory, local disk spill, TCP shuffle fetch, cgroup limit을 공유한다"가 됩니다.

꼬리 질문에는 네 경계로 답합니다. Lineage는 재계산 경로이고 checkpoint는 재시작 기준점입니다. Cache/persist는 반복 계산을 줄이지만 memory와 eviction 문제를 만듭니다. Shuffle은 network뿐 아니라 disk, serialization, memory, skew입니다. Structured Streaming은 source offset, state checkpoint, sink idempotency가 함께 맞아야 안전합니다. 이 경계를 말할 수 있으면 Spark를 "빠른 빅데이터 도구"가 아니라 OS와 분산 recovery 위에 올라간 runtime으로 설명할 수 있습니다.

## Data locality와 remote fetch는 scheduler 판단을 바꾼다

Spark scheduler는 task를 아무 executor에나 던지지 않습니다. Input data가 있는 곳, cached block이 있는 곳, executor availability, locality wait 같은 요소를 고려합니다. Data locality가 맞으면 network transfer를 줄일 수 있지만, locality를 너무 오래 기다리면 cluster resource가 놀 수 있습니다. Locality wait은 "가까운 data를 기다릴 것인가, 멀리서 가져오더라도 지금 실행할 것인가"의 tradeoff입니다.

```text
task needs partition P
  -> executor E1 has cached block
  -> E1 busy
  -> wait for locality?
  -> or launch on E2 and fetch remotely?
```

이 질문은 OS scheduler와 닮았습니다. 둘 다 locality와 utilization 사이를 고릅니다. CPU scheduler는 cache locality와 CPU 사용률을 조절하고, Spark scheduler는 data locality와 cluster utilization을 조절합니다. Kafka partition leader placement나 Cassandra token range ownership도 비슷한 질문을 갖습니다. "어디서 실행할 것인가"는 항상 data movement와 queue waiting을 동시에 만듭니다.

## External sink와 side effect는 task retry와 충돌한다

Spark task는 실패하면 재시도될 수 있습니다. Speculative execution이 켜져 있으면 느린 task의 복제본이 동시에 실행될 수 있습니다. 이때 task 안에서 외부 DB update, HTTP call, file append 같은 side effect를 직접 수행하면 중복이 생길 수 있습니다. Spark가 RDD/DataFrame lineage를 재계산할 수 있다는 것은 순수 계산에 강하다는 뜻이지, 외부 세계의 side effect를 자동으로 되돌린다는 뜻이 아닙니다.

```text
task attempt 1
  -> writes row to external DB
  -> executor dies before Spark records success

task attempt 2
  -> writes same row again

safe only if
  -> idempotent key
  -> transaction/commit protocol
  -> exactly-once capable sink boundary
```

Structured Streaming에서도 같은 문제가 batch 단위로 나타납니다. Source offset을 다시 읽을 수 있고 state checkpoint가 있어도 sink가 중복 write를 허용하지 않으면 결과가 깨질 수 있습니다. 그래서 foreachBatch나 custom sink를 쓸 때 idempotency key, batch id, transaction, merge/upsert semantics를 명시해야 합니다.

## Spark와 운영체제의 마지막 연결

Spark는 분산 계산 엔진이지만, 실제 병목은 매우 OS적입니다. Driver는 process이고 executor도 process입니다. Task는 thread에서 실행됩니다. Shuffle은 TCP socket과 local disk file입니다. Cache는 heap/off-heap memory입니다. Spill은 filesystem과 page cache를 씁니다. Heartbeat와 fetch timeout은 partial failure의 관측입니다. Container 안에서는 cgroup CPU/memory limit이 executor behavior를 바꿉니다.

```text
Spark symptom
  slow stage
  -> skew or scheduler delay?
  -> executor CPU/GC?
  -> shuffle network?
  -> local disk spill?
  -> cgroup throttle/OOM?
  -> lost executor and recomputation?
```

이 흐름을 말할 수 있으면 Spark tuning이 설정 표 암기가 아니라 trace가 됩니다. Input partition이 task가 되고, task가 thread에서 실행되고, shuffle이 disk와 network를 지나고, checkpoint가 recovery boundary가 되며, sink idempotency가 외부 correctness를 닫습니다. Spark를 OS와 분산 시스템의 bridge로 읽는 이유가 여기에 있습니다.
