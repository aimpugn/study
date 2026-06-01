# 00. 전체 학습 지도와 읽는 순서

## 목차

- [먼저 붙잡을 지도](#먼저-붙잡을-지도)
- [1단계: 한 머신을 작은 운영체제 교과서가 아니라 실제 서버 경로로 이해한다](#1단계-한-머신을-작은-운영체제-교과서가-아니라-실제-서버-경로로-이해한다)
- [2단계: 여러 머신에서는 관측이 불완전하다는 사실을 받아들인다](#2단계-여러-머신에서는-관측이-불완전하다는-사실을-받아들인다)
- [3단계: 세 시스템을 제품 기능이 아니라 내부 경로로 읽는다](#3단계-세-시스템을-제품-기능이-아니라-내부-경로로-읽는다)
- [4단계: 비교로 원리를 분리한다](#4단계-비교로-원리를-분리한다)
- [5단계: 짧게 답하고 다시 내려간다](#5단계-짧게-답하고-다시-내려간다)
- [실험과 근거를 함께 읽는 방법](#실험과-근거를-함께-읽는-방법)
- [이 코퍼스가 다루지 않는 것](#이-코퍼스가-다루지-않는-것)
- [마지막 자기 점검](#마지막-자기-점검)

이 코퍼스의 중심 질문은 하나입니다. 백엔드 시스템에서 "느리다", "유실됐다", "중복 처리됐다", "일관성이 깨졌다", "확장하면 될 줄 알았는데 더 느려졌다" 같은 증상이 보일 때, 제품 이름을 외우지 않고 어떤 상태가 어느 층에서 밀렸는지 설명할 수 있는가입니다.

OS 커널은 한 머신 안에서 이 질문을 다룹니다. CPU는 한 순간에 제한된 실행 흐름만 처리할 수 있고, 메모리는 프로세스마다 다른 주소 공간으로 보이며, 디스크와 네트워크 카드는 여러 프로세스가 함께 쓰는 장치입니다. 커널은 user mode와 kernel mode, syscall, scheduler, virtual memory, page cache, socket buffer 같은 장치로 이 공유 문제를 관리합니다.

OS 파트는 특정 교과서의 목차를 베끼지 않습니다. 대신 OSTEP의 virtualization/concurrency/persistence/security, CSAPP의 exceptional control flow와 system-level I/O, Operating Systems: Principles and Practice와 Operating System Concepts 계열의 process/thread/scheduling/synchronization/deadlock/memory/storage/I/O/protection/security/virtualization 주제군을 실제 백엔드 요청 경로에 맞게 다시 묶습니다. 그래서 부팅과 실행 파일, process lifecycle, virtual memory, VFS/page cache/block layer, NIC/TCP/epoll, futex/cgroup/namespace/observability가 모두 `01` 계열에서 독립적으로 닫힙니다.

분산 시스템은 같은 질문이 여러 머신으로 흩어졌을 때 등장합니다. 한 머신 안에서는 커널이 전역 상태를 많이 알고 있지만, 여러 머신 사이에는 완전한 전역 시계도 없고, 네트워크 지연과 노드 crash를 즉시 구분할 수도 없습니다. 그래서 log, partition, replica, quorum, consensus, checkpoint, retry, backpressure 같은 구조가 필요해집니다.

Kafka, Cassandra, Spark는 이 원리를 서로 다른 목적에 맞게 조립한 시스템입니다. Kafka는 record의 순서와 재생을 partition log로 다룹니다. Cassandra는 write 가용성과 scale-out을 위해 mutation을 commit log, memtable, SSTable, repair 흐름으로 다룹니다. Spark는 큰 계산을 DAG, stage, task, shuffle, lineage, checkpoint로 나누어 실패한 조각을 다시 계산합니다.

## 먼저 붙잡을 지도

```text
single machine
  process
    |
    +-- syscall: user request -> kernel decision
    +-- CPU: runnable task -> scheduler -> context switch
    +-- memory: virtual address -> page table -> frame / page fault
    +-- file: user bytes -> VFS -> page cache -> writeback -> block device
    +-- network: socket -> kernel buffer -> TCP/IP -> NIC

multiple machines
  request or record
    |
    +-- partition: key와 routing이 상태 조각의 owner를 정한다
    +-- log/order: 상태 변경은 재생 가능한 순서로 기록된다
    +-- replication: 다른 node가 같은 변경을 복사해 장애에 대비한다
    +-- consistency: read는 어떤 write까지 봐야 하는지 계약을 가진다
    +-- recovery: 실패 뒤 다시 시작할 checkpoint나 log 위치가 필요하다
```

이 지도에서 `상태`라는 말은 추상적인 기분이 아닙니다. 파일 offset, dirty page, socket send queue, Kafka offset, Cassandra timestamped cell, Spark shuffle block처럼 실제로 저장되거나 관찰되는 값입니다. 좋은 면접 답변은 이 상태가 어디에 있고 누가 바꾸며 장애가 나면 어떤 값이 남는지 말할 수 있어야 합니다.

## 1단계: 한 머신을 작은 운영체제 교과서가 아니라 실제 서버 경로로 이해한다

[01_os_kernel_foundations.md](01_os_kernel_foundations.md)는 OS 파트의 허브입니다. 여기서 `write(fd, buf, len)` 하나가 왜 단순한 라이브러리 호출이 아닌지, 그리고 이 질문이 Kafka의 log append, Cassandra의 commit log, Spark의 shuffle spill까지 어떻게 이어지는지 먼저 봅니다.

그 다음에는 다섯 개의 상세 문서로 내려갑니다.

| 순서 | 문서 | 핵심 기억 |
| --- | --- | --- |
| 1 | [01a_process_scheduling.md](01a_process_scheduling.md) | process와 thread는 scheduler 위에서 CPU 시간을 나눠 받고, runnable 상태여도 run queue에서 늦어질 수 있다. |
| 2 | [01b_memory_and_address_space.md](01b_memory_and_address_space.md) | virtual address, page table, TLB, page fault, mmap, OOM은 JVM/DB/process 메모리 증상을 물리 메모리 경쟁으로 연결한다. |
| 3 | [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md) | `write()`, page cache, dirty page, writeback, fsync, block layer는 빠른 반환과 durable write를 서로 다른 시점으로 나눈다. |
| 4 | [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md) | NIC로 들어온 packet은 driver, NAPI, TCP, socket buffer, epoll, scheduler를 지나 application request가 된다. |
| 5 | [01e_concurrency_isolation_observability.md](01e_concurrency_isolation_observability.md) | lock, futex, cgroup, namespace, `/proc`, perf/eBPF 관측은 장애 추론을 공유 상태, 대기 주체, 증거로 나눈다. |

여기서 배우는 습관은 "API 이름을 외우기 전에 그 API가 커널 안에서 어떤 객체와 queue를 건드리는지 본다"입니다. OS 파트는 Kafka/Cassandra/Spark를 이해하기 위한 낮은 하한선이 아니라, 세 시스템이 실제로 기대는 물리적 실행 경로입니다.

각 OS 상세 문서는 20,000자 이상의 독립 장문 문서로 확장되어 있습니다. 읽는 동안 분량 자체보다 trace를 붙잡으세요. 예를 들어 `01d`에서는 packet이 NIC RX ring, NAPI, TCP, socket buffer, epoll, scheduler를 지나 request가 되는지 말할 수 있어야 하고, `01c`에서는 `write()` 성공, dirty page, writeback, `fsync()`, Kafka ack, Cassandra CL success를 서로 다른 계층의 성공으로 나눌 수 있어야 합니다.

이 단계를 지나면 다음 질문에 답할 수 있어야 합니다.

- `fd`는 왜 파일 자체가 아니라 프로세스별 table의 index인가?
- CPU가 바쁘지 않아 보여도 thread가 runnable queue에서 기다릴 수 있는 이유는 무엇인가?
- page fault와 GC pause, disk writeback은 모두 latency로 보일 수 있는데 서로 무엇이 다른가?
- page cache가 성능을 높이는 동시에 durability 오해를 만들 수 있는 이유는 무엇인가?
- blocking/non-blocking, synchronous/asynchronous, readiness/completion은 왜 같은 축이 아닌가?
- NIC가 packet을 받았다는 사실과 application thread가 request를 처리했다는 사실 사이에는 어떤 kernel queue와 scheduling 지점이 있는가?

## 2단계: 여러 머신에서는 관측이 불완전하다는 사실을 받아들인다

[02_distributed_system_foundations.md](02_distributed_system_foundations.md)는 CAP나 quorum이라는 단어를 먼저 외우게 하지 않습니다. 먼저 한 노드가 응답하지 않을 때 그것이 crash인지, network delay인지, partition인지 관측만으로 확정하기 어렵다는 점에서 시작합니다. 그 다음 logical clock, log, state machine, partitioning, replication, quorum, consensus, retry, checkpoint로 올라갑니다.

이 단계를 지나면 다음 문장을 더 정확하게 바꿔 말할 수 있어야 합니다.

- "timeout이면 실패다"가 아니라 "timeout은 실패의 증명이 아니라 실패 가능성에 대한 관측이다."
- "replica가 많으면 안전하다"가 아니라 "어떤 write가 몇 replica에 기록되었고 어떤 read가 그 replica와 교차하는지가 중요하다."
- "retry하면 복구된다"가 아니라 "retry는 idempotency와 backpressure 없이 장애를 증폭할 수 있다."

## 3단계: 세 시스템을 제품 기능이 아니라 내부 경로로 읽는다

Kafka, Cassandra, Spark 문서는 기능 목록이 아닙니다. 각 문서는 하나의 작은 요청이 시스템 안에서 어떻게 바뀌는지 추적합니다.

| 문서 | 핵심 기억 | 내부에서 따라갈 상태 |
| --- | --- | --- |
| [03_kafka_deep_dive.md](03_kafka_deep_dive.md) | producer record는 leader partition의 log에 append되고, high watermark와 consumer offset을 거쳐 다시 읽을 수 있는 기록이 된다. | batch, leader partition, log segment, page cache, replica fetch, high watermark, consumer offset |
| [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md) | 빠른 write path는 commit log와 memtable로 먼저 받아들이고, read와 compaction은 여러 SSTable과 replica state를 다시 조립한다. | token, replica, commit log, memtable, SSTable, Bloom filter, tombstone, repair |
| [05_spark_deep_dive.md](05_spark_deep_dive.md) | 하나의 action은 DAG, stage, task, shuffle file로 쪼개지고, 실패한 조각은 lineage나 checkpoint를 기준으로 다시 계산된다. | DAG, stage boundary, partition, executor memory, spill, shuffle block, lineage, checkpoint |

세 문서의 공통 축은 각 시스템이 어떤 비용을 줄이기 위해 어떤 내부 상태를 중심에 두는지입니다. Kafka는 순서 있는 append와 재생을 싸게 만들기 위해 log를 중심에 둡니다. Cassandra는 여러 노드가 쓰기를 계속 받아야 하므로 write path를 append와 memory buffer 중심으로 만들고, 그 대가로 read repair와 compaction을 받아들입니다. Spark는 대용량 계산을 실패 가능한 작은 조각으로 나누기 위해 data partition과 lineage를 중심에 둡니다.

## 4단계: 비교로 원리를 분리한다

[06_cross_system_comparison.md](06_cross_system_comparison.md)는 세 제품을 표로 외우게 하려는 문서가 아닙니다. 같은 단어가 시스템마다 어떻게 다른 약속을 뜻하는지 분리합니다.

`log`는 Kafka에서 durable ordered record stream에 가깝고, Cassandra에서는 crash recovery를 위한 commit log와 SSTable history로 나뉘며, Spark에서는 lineage와 event/history 관찰 표면으로 나타납니다. `partition`은 Kafka에서는 ordering과 parallelism의 단위이고, Cassandra에서는 data ownership과 replica placement의 단위이며, Spark에서는 task scheduling과 shuffle의 단위입니다. 같은 단어를 같은 뜻으로 밀어붙이면 면접 꼬리 질문에서 바로 무너집니다.

## 5단계: 짧게 답하고 다시 내려간다

[07_interview_reasoning_playbook.md](07_interview_reasoning_playbook.md)는 긴 설명을 그대로 외우게 하지 않습니다. 먼저 20~30초 답변을 만들고, 꼬리 질문이 오면 OS 자원, 분산 상태, 제품 내부 객체 순서로 다시 펼치는 연습을 합니다.

예를 들어 "Kafka는 왜 빠른가요?"라는 질문의 짧은 답은 "partition log에 순차 append하고, batch와 page cache, sendfile 경로로 작은 I/O와 불필요한 copy를 줄이기 때문입니다"입니다. 꼬리 질문이 오면 `producer batch -> leader append -> page cache -> replica fetch -> consumer fetch -> offset commit`으로 내려갑니다.

## 실험과 근거를 함께 읽는 방법

[08_experiments_and_observability.md](08_experiments_and_observability.md)는 운영용 runbook이 아니라 학습 검증용 실험입니다. 명령은 항상 목적, 전제, 예상 관찰, PASS/FAIL 신호와 함께 읽어야 합니다. 특히 CPU 부하, page cache, socket queue, JVM thread dump, Kafka/Cassandra/Spark CLI는 운영 환경에서 함부로 실행하면 부작용이 있을 수 있으므로 local-only와 read-only 경계를 분리합니다.

[09_glossary.md](09_glossary.md)는 처음 보는 영어 용어를 한국어로 붙잡기 위한 문서입니다. 원어를 지우지 않는 이유는 공식 문서를 역추적해야 하기 때문입니다. [10_source_ledger.md](10_source_ledger.md)는 어떤 주장이 공식 문서 직접 근거인지, 논문 근거인지, 일반 원리인지, 버전 확인이 필요한지 구분합니다.

## 이 코퍼스가 다루지 않는 것

이 문서 묶음은 Linux kernel source 전체 해설, Kafka/Cassandra/Spark 운영 인증 준비, Kubernetes/YARN 운영 전체, 보안 설정 전체, production cluster 장애 주입을 목표로 하지 않습니다. 대신 면접과 실무 추론에 필요한 내부 구조의 뼈대를 세웁니다. 특정 설정명과 기본값은 버전마다 달라질 수 있으므로, 운영 적용 전에는 반드시 target version 공식 문서를 다시 확인해야 합니다.

## 마지막 자기 점검

문서를 다 읽은 뒤에는 아래 순서로 스스로 설명해 봅니다.

```text
API call
  -> internal object
  -> buffer/cache/log
  -> ordering or ownership rule
  -> failure point
  -> recovery or verification path
```

이 여섯 칸을 채우지 못하면 아직 키워드를 외운 상태입니다. 채울 수 있다면 Kafka, Cassandra, Spark 중 모르는 설정이 나와도 "이 설정은 어느 상태 이동을 바꾸는가"라는 질문으로 다시 공부할 수 있습니다.
