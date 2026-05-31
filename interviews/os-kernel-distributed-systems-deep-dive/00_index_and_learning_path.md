# 00. 전체 학습 지도와 읽는 순서

> OS 커널은 한 머신 안에서 CPU 시간, 메모리 주소, 파일의 byte, 네트워크 packet을 공정하고 안전하게 나누는 계층입니다.
> 분산 시스템은 이 자원들이 여러 머신으로 흩어졌을 때 실패, 시간, 순서, 복제, 회복을 다루는 계층입니다.
> Kafka, Cassandra, Spark는 모두 partition, log, replica, scheduler, checkpoint 같은 공통 원리를 쓰지만, 각각 지키려는 약속이 다릅니다.
> 따라서 면접 답변은 제품명을 먼저 외우기보다 "어떤 상태가 어디에 있고, 누가 다음 상태를 결정하며, 실패하면 어떤 관측값이 남는가"를 따라가야 깊어집니다.

## 이 코퍼스가 답하는 질문

이 문서 묶음은 다음 질문에 순서대로 답합니다.

1. 한 머신 안에서 프로그램의 요청은 커널을 거쳐 CPU, 메모리, 디스크, 네트워크로 어떻게 이동하는가?
2. 여러 머신으로 나뉜 시스템에서는 왜 "한 번 쓰면 모두가 바로 같은 값을 본다"는 기대가 깨지는가?
3. Kafka는 왜 queue보다 log로 이해해야 하고, partition과 offset은 왜 성능과 순서의 경계인가?
4. Cassandra는 왜 B-tree 기반 데이터베이스처럼 update-in-place 하지 않고 commit log, memtable, SSTable, compaction을 쓰는가?
5. Spark는 왜 하나의 큰 작업을 DAG, stage, task, shuffle, lineage로 나누는가?
6. 장애, 성능, 일관성, 확장성 질문이 나왔을 때 어떻게 lower layer부터 다시 설명할 수 있는가?

## 전제 지도

이 코퍼스는 독자를 전문가로 가정하지 않습니다. 다만 백엔드 개발자가 자주 만나는 HTTP 서버, 데이터베이스, JVM 프로세스, 로그 파일, 네트워크 지연 같은 경험은 발판으로 사용합니다.

```
첫 발판
  |
  +-- 프로그램은 syscall로 커널에 요청한다.
  |     |
  |     +-- 파일 I/O: byte -> page cache -> block I/O -> device
  |     +-- 네트워크 I/O: byte -> socket buffer -> TCP/IP -> NIC
  |     +-- 메모리: virtual address -> page table -> physical frame
  |     +-- CPU: runnable task -> scheduler -> context switch
  |
  +-- 여러 머신은 같은 순간을 공유하지 않는다.
        |
        +-- message delay, crash, partition, retry
        +-- log order, quorum, leader, replica, checkpoint
        +-- consistency, availability, latency tradeoff
```

Kafka, Cassandra, Spark는 위 두 층을 서로 다른 방식으로 조합합니다.

| 시스템 | 기본 상태 | 확장 단위 | 회복 방식 | OS와 만나는 지점 |
|---|---|---|---|---|
| Kafka | append-only log record | topic partition | replica log와 consumer offset | page cache, sequential disk I/O, sendfile, socket buffer |
| Cassandra | partitioned row mutation | token range, vnode, replica | hinted handoff, read repair, anti-entropy repair, compaction | commit log, SSTable, page cache, disk bandwidth, heap/off-heap memory |
| Spark | partitioned dataset transformation | partition, task, stage | lineage recomputation, checkpoint | JVM process, executor heap, CPU scheduling, network shuffle, disk spill |

## 초보자에서 전문가 사고방식까지의 경로

### 1단계: 이름을 안다

목표는 용어를 외우는 것이 아니라 말을 걸 수 있는 정도의 첫 모델을 만드는 것입니다. 이 단계에서는 user mode와 kernel mode, process와 thread, partition과 replica, offset과 checkpoint를 서로 구분할 수 있으면 충분합니다.

읽을 문서:

- [01_os_kernel_foundations.md](01_os_kernel_foundations.md)
- [02_distributed_system_foundations.md](02_distributed_system_foundations.md)
- [09_glossary.md](09_glossary.md)

Teach-back 목표:

- "시스템 콜은 왜 함수 호출과 다른가?"
- "분산 시스템에서 partial failure가 왜 어렵나?"
- "partition은 왜 성능 단위이면서 correctness 경계가 되는가?"

### 2단계: 상태 이동을 그린다

이 단계에서는 값이 실제로 어디서 어디로 이동하는지 그릴 수 있어야 합니다. 예를 들어 Kafka producer가 보낸 record가 broker의 page cache와 replica fetch를 거쳐 consumer offset으로 이어지는 trace를 설명합니다.

읽을 문서:

- [03_kafka_deep_dive.md](03_kafka_deep_dive.md)
- [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md)
- [05_spark_deep_dive.md](05_spark_deep_dive.md)

Teach-back 목표:

- "Kafka consumer lag는 어떤 상태가 밀린 것인가?"
- "Cassandra write는 commit log와 memtable 중 어디에 먼저 남는가?"
- "Spark shuffle은 왜 네트워크와 디스크 병목을 동시에 만들 수 있는가?"

### 3단계: 실패를 추론한다

이 단계에서는 증상에서 원인 후보를 나눕니다. `latency 증가`라는 하나의 증상도 CPU run queue, GC pause, page cache writeback, disk saturation, network retransmission, replica lag, compaction pressure, shuffle spill 중 어디서 왔는지 구분해야 합니다.

읽을 문서:

- [06_cross_system_comparison.md](06_cross_system_comparison.md)
- [08_experiments_and_observability.md](08_experiments_and_observability.md)

Teach-back 목표:

- "네트워크 지연과 노드 crash는 왜 관측상 구분하기 어려운가?"
- "page cache 덕분에 write가 빨라 보여도 데이터가 아직 device에 없을 수 있다는 말은 무슨 뜻인가?"
- "backpressure는 어느 queue가 길어지는 현상인가?"

### 4단계: 면접 답변으로 압축한다

전문가처럼 답한다는 것은 길게 말하는 능력이 아닙니다. 처음에는 짧게 답하고, 꼬리 질문이 오면 아래층부터 다시 펼칠 수 있어야 합니다.

읽을 문서:

- [07_interview_reasoning_playbook.md](07_interview_reasoning_playbook.md)

Teach-back 목표:

- "Kafka는 왜 빠른가?"를 page cache, sequential I/O, batching, zero-copy, consumer pull로 설명할 수 있다.
- "Cassandra는 왜 eventual consistency를 택했나?"를 quorum, availability, LSM, repair로 설명할 수 있다.
- "Spark job이 느릴 때 어디를 보나?"를 DAG, skew, shuffle, memory, spill, locality로 설명할 수 있다.

## 문서별 읽은 뒤 설명할 수 있어야 하는 것

| 문서 | 읽은 뒤 설명할 수 있어야 하는 것 |
|---|---|
| `01_os_kernel_foundations.md` | syscall, context switch, page fault, page cache, socket buffer, lock contention이 실제 시스템 증상으로 어떻게 나타나는지 |
| `02_distributed_system_foundations.md` | partial failure, happens-before, quorum, consensus, CAP/PACELC, idempotency, backpressure, checkpointing을 상태 전이로 설명하는 법 |
| `03_kafka_deep_dive.md` | Kafka record가 producer에서 broker log, replica, consumer offset까지 이동하는 과정과 장애 지점 |
| `04_cassandra_deep_dive.md` | Cassandra mutation이 commit log, memtable, SSTable, compaction, repair로 이동하는 과정과 stale read 추론 |
| `05_spark_deep_dive.md` | Spark action이 DAG, stage, task, shuffle, executor memory, spill, lineage recovery로 바뀌는 과정 |
| `06_cross_system_comparison.md` | 세 시스템의 공통 원리와 차이를 log, partition, replication, recovery, backpressure, checkpoint 관점에서 비교하는 법 |
| `07_interview_reasoning_playbook.md` | 짧은 답변 뒤 꼬리 질문을 lower-layer로 내려가며 확장하는 법 |
| `08_experiments_and_observability.md` | 로컬에서 OS와 분산 시스템 증상을 관찰하고 PASS/FAIL 신호를 분리하는 법 |
| `09_glossary.md` | 핵심 용어를 쉬운 한국어와 공식 영어 원어로 함께 설명하는 법 |
| `10_source_ledger.md` | 어떤 주장이 공식 문서 직접 근거인지, 논문 기반 일반 원리인지, 버전 확인이 필요한지 구분하는 법 |

## 이 코퍼스가 다루는 범위

다루는 범위:

- Linux 중심 OS 커널 원리와 macOS 관찰 차이
- 일반 분산 시스템 원리
- Apache Kafka, Apache Cassandra, Apache Spark의 핵심 내부 구조
- 장애, 성능, 일관성, 확장성 질문을 원리로 추론하는 방법
- 로컬에서 가능한 작은 실험과 관찰 명령
- 면접에서 쓸 수 있는 짧은 답변과 꼬리 질문 대응

다루지 않는 범위:

- Linux kernel source 전체 해설
- Kafka/Cassandra/Spark 운영 인증, 보안, 세부 설정 전체 목록
- Kubernetes, YARN, ZooKeeper/KRaft migration, Cassandra Accord 같은 별도 대형 주제의 완전한 운영 가이드
- Spark MLlib, GraphX, Catalyst optimizer 전체 구현
- production cluster를 실제로 띄우고 장애를 주입하는 고위험 실험

버전별로 달라지는 설정명과 기본값은 [10_source_ledger.md](10_source_ledger.md)에 확인 날짜와 함께 남깁니다. 본문에서는 원리와 상태 이동을 우선하고, 특정 버전의 기본값은 공식 문서로 확인한 경우에만 단정합니다.

## Cross-link 구조

핵심 흐름은 다음과 같습니다.

```
00 학습 지도
  |
  +--> 01 OS kernel foundations
  |       +--> 03 Kafka page cache / sendfile
  |       +--> 04 Cassandra commit log / SSTable / compaction
  |       +--> 05 Spark executor memory / spill / network shuffle
  |
  +--> 02 Distributed systems foundations
          +--> 03 Kafka replication / ISR / offset
          +--> 04 Cassandra RF / quorum / repair
          +--> 05 Spark lineage / checkpoint / task retry

03 + 04 + 05 --> 06 comparison --> 07 interview reasoning --> 08 experiments
                                                |
                                                +--> 09 glossary
                                                +--> 10 source ledger
```

## 공부 방법

각 문서에서 `active recall` 질문을 먼저 보고 답하지 못하면 바로 앞의 trace로 돌아갑니다. 답을 외우지 말고 그림을 다시 그리면 됩니다.

면접 대비에서는 `07`의 짧은 답변을 먼저 말한 뒤, 같은 질문을 `01`의 kernel 자원, `02`의 분산 상태, `03~05`의 시스템 내부 구조로 다시 펼쳐 봅니다. 답변이 길어지는 것이 목표가 아니라, 질문이 깊어져도 어느 층에서 상태가 깨졌는지 잃지 않는 것이 목표입니다.

