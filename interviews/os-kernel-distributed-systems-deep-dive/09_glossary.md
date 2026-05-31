# 09. Glossary

> 용어집의 목적은 영어 단어를 외우는 것이 아니라, 쉬운 한국어 의미와 공식 원어를 연결해 문서와 공식 자료를 오갈 수 있게 하는 것입니다.
> 헷갈리는 개념은 함께 적어 두어 면접에서 비슷한 단어를 섞어 쓰지 않게 합니다.
> 한 문장 답변은 직답용이며, 깊은 꼬리 답변은 각 본문 문서의 trace로 돌아가야 합니다.

| 쉬운 한국어 의미 | 영어 원어 | 헷갈리는 개념 | 처음 등장 | 면접 한 문장 |
|---|---|---|---|---|
| 애플리케이션이 실행되는 제한된 CPU 권한 영역 | user mode | kernel mode | [01](01_os_kernel_foundations.md) | 사용자 모드는 프로그램이 직접 하드웨어와 전역 커널 상태를 만지지 못하게 제한된 실행 영역입니다. |
| 커널이 하드웨어와 전역 자원을 다룰 수 있는 권한 영역 | kernel mode | user mode | [01](01_os_kernel_foundations.md) | 커널 모드는 syscall, interrupt, fault 처리처럼 보호된 자원 접근이 필요한 코드를 실행하는 영역입니다. |
| 사용자 코드가 커널 기능을 요청하는 공식 진입 | system call | library call, trap | [01](01_os_kernel_foundations.md) | 시스템 콜은 파일, socket, process 같은 OS 자원을 쓰기 위해 사용자 모드에서 커널 모드로 넘어가는 경계입니다. |
| 외부 장치나 timer가 CPU 처리를 요청하는 사건 | interrupt | trap, syscall | [01](01_os_kernel_foundations.md) | 인터럽트는 장치나 timer가 현재 CPU 흐름을 잠시 멈추고 커널 handler를 실행하게 만드는 비동기 사건입니다. |
| 현재 명령 실행 중 예외나 의도적 커널 진입 | trap | interrupt | [01](01_os_kernel_foundations.md) | trap은 page fault나 syscall처럼 현재 실행 흐름 안에서 커널 처리가 필요해지는 진입입니다. |
| 독립 주소 공간과 자원 테이블을 가진 실행 단위 | process | thread | [01](01_os_kernel_foundations.md) | 프로세스는 자기 주소 공간과 fd 같은 자원 테이블을 가진 OS 실행 단위입니다. |
| 같은 process 안에서 CPU 실행 흐름을 나누는 단위 | thread | process, task | [01](01_os_kernel_foundations.md) | thread는 같은 주소 공간을 공유하면서 독립적으로 scheduling될 수 있는 실행 흐름입니다. |
| CPU가 실행 흐름을 바꾸기 위해 register와 상태를 저장/복원하는 일 | context switch | mode switch | [01](01_os_kernel_foundations.md) | context switch는 CPU가 한 thread/process에서 다른 실행 흐름으로 넘어가기 위해 실행 상태를 교체하는 비용입니다. |
| process마다 자기 메모리처럼 보이는 주소 공간 | virtual memory | physical memory | [01](01_os_kernel_foundations.md) | 가상 메모리는 process isolation과 보호, lazy allocation, file mapping을 가능하게 하는 주소 추상화입니다. |
| 가상 주소를 물리 frame으로 연결하는 자료구조 | page table | TLB | [01](01_os_kernel_foundations.md) | page table은 CPU가 보는 가상 주소를 실제 메모리 frame으로 바꾸는 mapping table입니다. |
| 주소 변환 cache | TLB | page cache | [01](01_os_kernel_foundations.md) | TLB는 page table walk를 매번 하지 않도록 자주 쓰는 주소 변환을 CPU 가까이에 캐시합니다. |
| 주소 변환이나 접근 권한 문제로 커널 handler가 실행되는 사건 | page fault | TLB miss, segmentation fault | [01](01_os_kernel_foundations.md) | page fault는 lazy allocation처럼 정상일 수도 있고 권한 오류처럼 process 종료로 이어질 수도 있는 memory exception입니다. |
| 파일 내용을 memory page 단위로 보관하는 커널 cache | page cache | application cache, TLB | [01](01_os_kernel_foundations.md) | page cache는 일반 파일 read/write가 거치는 kernel memory cache이며 `write()` 성공과 device durability를 분리합니다. |
| file descriptor에서 socket으로 copy를 줄여 전송하는 kernel 경로 | sendfile | zero-copy 일반 | [01](01_os_kernel_foundations.md), [03](03_kafka_deep_dive.md) | sendfile은 Kafka 같은 log reader가 page cache의 file data를 socket으로 더 적은 copy로 보낼 수 있게 해 주는 syscall입니다. |
| 여러 실행 흐름이 같은 상태를 동시에 바꾸지 못하게 하는 도구 | lock | semaphore, consensus | [01](01_os_kernel_foundations.md) | lock은 local shared state의 변경 순서를 만들어 correctness를 지키지만 contention 비용을 만들 수 있습니다. |
| 일부 노드나 링크만 실패하는 상태 | partial failure | total failure | [02](02_distributed_system_foundations.md) | partial failure는 timeout만으로 crash, delay, pause, partition을 확정할 수 없게 만드는 분산 시스템의 핵심 난점입니다. |
| 사건 사이의 원인 순서 | happens-before | wall-clock order | [02](02_distributed_system_foundations.md) | happens-before는 물리 시각보다 메시지 송수신과 local event 순서로 결정되는 원인 관계입니다. |
| 물리 시계 없이 사건 순서를 추적하는 counter | logical clock | wall clock, timestamp | [02](02_distributed_system_foundations.md) | logical clock은 실제 시간을 재는 장치가 아니라 분산 사건의 순서 관계를 추적하는 도구입니다. |
| 상태 변경 기록을 순서대로 남긴 것 | log | queue, table | [02](02_distributed_system_foundations.md) | log는 실패 후 replay하거나 여러 replica가 같은 순서로 상태를 적용하게 만드는 기반 기록입니다. |
| 같은 명령 순서에 같은 결과를 내는 상태 처리 규칙 | state machine | process state | [02](02_distributed_system_foundations.md) | state machine은 같은 초기 상태와 같은 입력 순서가 주어지면 같은 결과를 내는 deterministic processing model입니다. |
| 데이터를 여러 조각으로 나누는 일 | partitioning | replication, sharding | [02](02_distributed_system_foundations.md) | partitioning은 scale-out을 위해 data나 일을 나누지만 ordering과 transaction 경계도 함께 만듭니다. |
| partitioning을 storage shard 관점에서 부르는 말 | sharding | partitioning | [02](02_distributed_system_foundations.md) | sharding은 큰 data set을 여러 shard에 나눠 저장해 용량과 처리량을 확장하는 방식입니다. |
| 같은 데이터를 여러 node에 두는 일 | replication | partitioning | [02](02_distributed_system_foundations.md) | replication은 한 node 실패에도 data를 읽거나 복구할 수 있도록 복사본을 두는 방식입니다. |
| 성공 판단에 필요한 replica 응답 집합 | quorum | consensus | [02](02_distributed_system_foundations.md) | quorum은 read/write 집합이 겹치도록 응답 수를 정해 stale 가능성을 줄이는 기법입니다. |
| 여러 node가 같은 값이나 순서에 안전하게 동의하는 문제 | consensus | quorum, leader election | [02](02_distributed_system_foundations.md) | consensus는 실패가 있어도 replica들이 같은 log entry나 결정에 동의하도록 만드는 프로토콜 문제입니다. |
| 실제 시간 순서와 같은 단일 객체처럼 보이는 강한 일관성 | linearizability | sequential consistency, eventual consistency | [02](02_distributed_system_foundations.md) | linearizability는 write가 끝난 뒤 시작한 read가 그 write를 보거나 오류를 내야 하는 강한 모델입니다. |
| 시간이 지나면 replica가 수렴한다는 약속 | eventual consistency | weak consistency | [02](02_distributed_system_foundations.md) | eventual consistency는 update가 멈추면 replica가 결국 같은 상태로 수렴한다는 모델입니다. |
| partition 상황에서 consistency와 availability tradeoff를 말하는 정리 | CAP theorem | PACELC | [02](02_distributed_system_foundations.md) | CAP은 네트워크 partition이 있을 때 consistency와 availability를 동시에 완전하게 보장할 수 없다는 제한입니다. |
| partition이 없을 때도 latency와 consistency tradeoff가 남는다는 관점 | PACELC | CAP theorem | [02](02_distributed_system_foundations.md) | PACELC는 partition 중에는 A/C, 평상시에는 latency/consistency 선택이 남는다고 설명합니다. |
| 같은 요청을 여러 번 처리해도 최종 상태가 한 번과 같은 성질 | idempotency | exactly-once | [02](02_distributed_system_foundations.md) | idempotency는 retry가 필요한 분산 시스템에서 중복 side effect를 막는 핵심 성질입니다. |
| 생산 속도보다 소비 속도가 낮아 압력이 거꾸로 전파되는 현상 | backpressure | throttling | [02](02_distributed_system_foundations.md) | backpressure는 queue, lag, buffer, spill, timeout으로 드러나는 처리률 불균형입니다. |
| 실패 후 다시 시작할 위치나 중간 상태 기록 | checkpoint | snapshot, offset | [02](02_distributed_system_foundations.md) | checkpoint는 긴 log나 lineage를 줄여 recovery 시작점을 만드는 기록입니다. |
| Kafka의 record 묶음 이름 | topic | queue | [03](03_kafka_deep_dive.md) | topic은 record stream을 담는 logical name이고 내부적으로 여러 partition log로 나뉠 수 있습니다. |
| Kafka topic을 나눈 append-only log 조각 | partition | shard, Spark partition | [03](03_kafka_deep_dive.md) | Kafka partition은 처리량과 consumer 병렬성의 단위이자 순서 보장의 경계입니다. |
| Kafka partition 안 record 위치 | offset | timestamp, checkpoint | [03](03_kafka_deep_dive.md) | offset은 partition 안에서 record 위치를 나타내며 consumer가 어디까지 읽었는지 기록하는 기준입니다. |
| Kafka partition에서 write/read 순서를 정하는 replica | leader replica | controller | [03](03_kafka_deep_dive.md) | leader replica는 해당 partition의 append 순서를 정하고 follower들이 그 log를 따라옵니다. |
| leader를 충분히 따라오는 Kafka replica 집합 | ISR | replicas | [03](03_kafka_deep_dive.md) | ISR은 단순 복사본 목록이 아니라 leader log를 충분히 따라오는 in-sync replica 집합입니다. |
| Kafka consumer들이 같은 일을 나눠 처리하는 집합 | consumer group | topic subscriber | [03](03_kafka_deep_dive.md) | consumer group은 partition을 group member에게 나눠 할당하고 group별 offset을 관리합니다. |
| 소비자가 log 끝보다 뒤처진 거리 | consumer lag | network lag | [03](03_kafka_deep_dive.md) | consumer lag는 partition log end offset과 group committed offset의 차이입니다. |
| Kafka에서 key별 최신 record를 남기도록 log를 줄이는 정책 | log compaction | Cassandra compaction | [03](03_kafka_deep_dive.md) | Kafka log compaction은 key의 최신 값을 유지해 state rebuild를 돕는 retention 방식입니다. |
| Cassandra에서 node/key range를 배치하는 원형 hash 공간 | token ring | Kafka partition | [04](04_cassandra_deep_dive.md) | token ring은 partition key hash가 어느 replica set에 배치될지 정하는 Cassandra의 keyspace 지도입니다. |
| Cassandra에서 한 physical node가 가진 여러 ring 위치 | vnode | partition | [04](04_cassandra_deep_dive.md) | vnode는 data balance와 scale-out을 돕지만 repair와 failure 조합을 더 복잡하게 만들 수 있습니다. |
| Cassandra partition 복사본 개수 | replication factor | consistency level | [04](04_cassandra_deep_dive.md) | replication factor는 각 partition을 몇 replica에 저장할지 정합니다. |
| Cassandra read/write가 기다리는 응답 수준 | consistency level | replication factor | [04](04_cassandra_deep_dive.md) | consistency level은 성공으로 인정할 replica 응답 수를 정해 latency와 consistency를 조절합니다. |
| Cassandra crash recovery용 append-only 기록 | commit log | Kafka topic log | [04](04_cassandra_deep_dive.md) | commit log는 mutation을 memtable과 함께 안전하게 복구하기 위한 내부 append 기록입니다. |
| Cassandra memory write buffer | memtable | page cache | [04](04_cassandra_deep_dive.md) | memtable은 mutation을 memory에 정렬된 형태로 보관하다가 SSTable로 flush되는 구조입니다. |
| Cassandra의 immutable disk table file | SSTable | B-tree page | [04](04_cassandra_deep_dive.md) | SSTable은 flush된 data를 불변 파일로 저장하고 read/compaction의 기본 단위가 됩니다. |
| "없다"를 빠르게 알려 주는 확률적 자료구조 | Bloom filter | cache | [04](04_cassandra_deep_dive.md) | Bloom filter는 어떤 SSTable에 key가 없음을 빠르게 판단해 불필요한 disk read를 줄입니다. |
| 여러 SSTable을 병합해 최신 값과 tombstone을 정리하는 작업 | compaction | compression, Kafka compaction | [04](04_cassandra_deep_dive.md) | Cassandra compaction은 read amplification을 줄이지만 disk/CPU 압력을 만드는 background merge입니다. |
| 삭제를 표시하는 record | tombstone | null value | [04](04_cassandra_deep_dive.md) | tombstone은 delete를 replica와 compaction이 이해할 수 있게 남기는 삭제 표시입니다. |
| 내려간 replica를 위해 coordinator가 임시로 저장한 mutation | hinted handoff | anti-entropy repair | [04](04_cassandra_deep_dive.md) | hinted handoff는 unavailable replica가 돌아왔을 때 missed write를 전달하려는 best-effort 복구 기법입니다. |
| replica data 차이를 비교해 수렴시키는 repair | anti-entropy repair | read repair | [04](04_cassandra_deep_dive.md) | anti-entropy repair는 replica 간 data divergence를 체계적으로 줄이는 수렴 작업입니다. |
| Spark 사용자 프로그램의 조정 process | driver | executor | [05](05_spark_deep_dive.md) | driver는 DAG를 만들고 task scheduling을 조정하는 Spark application의 중심 process입니다. |
| Spark task를 실행하는 worker-side process | executor | worker node | [05](05_spark_deep_dive.md) | executor는 worker node에서 task를 실행하고 cache와 shuffle data를 보관하는 JVM process입니다. |
| Spark의 낮은 수준 partitioned data abstraction | RDD | DataFrame | [05](05_spark_deep_dive.md) | RDD는 partitioned collection과 transformation lineage를 가진 Spark의 기본 abstraction입니다. |
| schema와 optimizer를 활용하는 Spark data abstraction | DataFrame | RDD, table | [05](05_spark_deep_dive.md) | DataFrame은 schema를 가진 distributed data abstraction으로 optimizer가 실행 계획을 개선할 수 있습니다. |
| 순환 없는 작업 의존 그래프 | DAG | stage | [05](05_spark_deep_dive.md) | DAG는 transformation dependency를 나타내며 Spark가 stage와 task로 나누는 실행 계획입니다. |
| parent partition 일부만 필요한 dependency | narrow dependency | wide dependency | [05](05_spark_deep_dive.md) | narrow dependency는 data 이동 없이 pipeline으로 처리하기 쉬운 dependency입니다. |
| 여러 parent partition의 data 재분배가 필요한 dependency | wide dependency | narrow dependency | [05](05_spark_deep_dive.md) | wide dependency는 shuffle을 만들고 stage 경계가 되는 dependency입니다. |
| partition 간 data를 재분배하는 작업 | shuffle | repartition only | [05](05_spark_deep_dive.md) | shuffle은 network, disk, serialization, memory pressure가 함께 드는 Spark의 비싼 data movement입니다. |
| executor memory 부족 시 중간 data를 disk에 내리는 일 | spill | checkpoint | [05](05_spark_deep_dive.md) | spill은 OOM을 피하지만 local disk I/O와 merge 비용으로 task를 느리게 만듭니다. |
| lost partition을 다시 계산하기 위한 transformation 기록 | lineage | checkpoint | [05](05_spark_deep_dive.md) | lineage는 Spark가 data replica 없이도 partition을 재계산할 수 있게 하는 recipe입니다. |
| 어느 partition/key/task가 유난히 커지는 불균형 | skew | load average | [05](05_spark_deep_dive.md) | skew는 평균 자원이 충분해도 가장 큰 partition이나 task가 전체 stage를 붙잡는 현상입니다. |

## Active Recall

1. page cache, TLB, application cache를 한 문장씩 구분해 보세요.
2. Kafka partition, Cassandra partition, Spark partition이 각각 나누는 것은 무엇인가요?
3. quorum과 consensus가 같은 말이 아닌 이유를 설명해 보세요.
4. compaction이라는 단어가 Kafka와 Cassandra에서 어떻게 다른가요?
5. checkpoint, offset, lineage의 공통점과 차이를 말해 보세요.

