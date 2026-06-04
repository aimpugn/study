# OS Kernel and Distributed Systems Deep Dive

## 목차

- [읽는 순서](#읽는-순서)
- [통과 기준](#통과-기준)

이 코퍼스는 OS 커널, 분산 시스템 원리, Kafka, Cassandra, Spark를 따로 외우는 문서가 아닙니다.
목표는 한 백엔드 개발자가 `write()` 하나에서 시작해 page cache, `fsync()`, replication, quorum, shuffle, checkpoint까지 이어지는 상태 이동을 자기 말로 설명하게 만드는 것입니다.
제품 이름은 뒤에 붙습니다.
먼저 붙잡아야 할 것은 요청이 어느 계층에서 어떤 상태로 바뀌고, 그 상태가 실패 뒤 어디에 남는가입니다.

가장 작은 첫 장면은 아래처럼 볼 수 있습니다.

```text
application calls write(fd, bytes)
  -> kernel finds open file state from fd
  -> bytes become dirty page in page cache
  -> writeback or fsync pushes the page toward storage
  -> product protocol later decides ack, retry, replay, or recovery
```

이 한 줄의 요청이 파일, 메모리, 장치 큐(queue), 제품 내부 log를 차례로 만나기 때문에 OS와 분산 시스템을 따로 외우면 실제 장애 질문에서 설명이 끊깁니다.

첫 번째 기준은 self-contained입니다.
다른 문서를 이미 읽었다고 가정하지 않습니다.
같은 개념이 앞뒤에서 다시 나오더라도, 그 문서 안에서 필요한 만큼 다시 설명합니다.
두 번째 기준은 prose-first입니다.
`질문/직관/깨지는 지점` 같은 내부 작성 체크리스트를 본문 목차로 반복하지 않고, 독자가 읽는 순서대로 원인, 역사, 작은 상태, 내부 경로, 깨지는 지점, 확인 방법이 자연스럽게 이어지게 씁니다.

문서의 강도는 글자 수가 아니라 다시 설명 가능한 경로로 판단합니다.
각 문서는 최소 하나 이상의 실제 상태 이동 trace를 중심으로 OS 객체, 분산 상태, 제품 내부 queue/log/buffer를 연결하고, 관측 경로와 과장하면 안 되는 지점을 함께 남깁니다.
읽는 동안에는 아래 모양을 계속 다시 그리면 됩니다.

```text
external request or record
  -> kernel/runtime object
  -> queue/cache/buffer/log
  -> ordering or ownership rule
  -> failure point
  -> recovery or verification path
```

## 읽는 순서

1. [00_index_and_learning_path.md](00_index_and_learning_path.md)에서 전체 지도를 잡습니다.
2. [01_os_kernel_foundations.md](01_os_kernel_foundations.md)에서 커널 학습의 허브를 먼저 읽습니다.
3. OS 상세 문서인 [01a_process_scheduling.md](01a_process_scheduling.md), [01b_memory_and_address_space.md](01b_memory_and_address_space.md), [01c_filesystem_page_cache_block_io.md](01c_filesystem_page_cache_block_io.md), [01d_network_stack_and_io_multiplexing.md](01d_network_stack_and_io_multiplexing.md), [01e_concurrency_isolation_observability.md](01e_concurrency_isolation_observability.md)를 순서대로 읽습니다. 여기서 CPU, 메모리, 디스크, 네트워크, 동시성, 컨테이너 제한, 관측 도구를 실제 요청 경로로 묶습니다.
4. [02_distributed_system_foundations.md](02_distributed_system_foundations.md)에서 여러 머신 사이의 실패, 시간, 순서, 복제, 일관성, 회복을 배웁니다.
5. [03_kafka_deep_dive.md](03_kafka_deep_dive.md), [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md), [05_spark_deep_dive.md](05_spark_deep_dive.md)를 읽으며 같은 원리가 제품 내부 구조에서 어떻게 다르게 나타나는지 봅니다.
6. [06_cross_system_comparison.md](06_cross_system_comparison.md)과 [07_interview_reasoning_playbook.md](07_interview_reasoning_playbook.md)에서 세 시스템을 비교하고 면접식 짧은 답변으로 압축합니다.
7. [08_experiments_and_observability.md](08_experiments_and_observability.md), [09_glossary.md](09_glossary.md), [10_source_ledger.md](10_source_ledger.md)는 실험, 용어, 근거를 확인할 때 함께 봅니다.

## 통과 기준

이 문서 묶음을 제대로 읽었다면 다음 질문에 외운 키워드가 아니라 경로로 답할 수 있어야 합니다.

- `write()`가 성공해도 왜 아직 durable write가 아닐 수 있는가?
- Kafka가 빠르다는 말은 page cache, sequential append, batch, sendfile 중 어떤 상태 이동을 줄였다는 뜻인가?
- Cassandra의 quorum read/write는 왜 "항상 최신"이라는 말로 끝나지 않는가?
- Spark shuffle이 느릴 때 왜 executor heap, spill file, network, skew를 같이 봐야 하는가?
- 장애 질문이 오면 어떤 상태가 어디에 기록되었고, 실패 후 어디서 다시 시작할 수 있는지 어떻게 추론하는가?

이 질문들은 모두 같은 구조로 돌아갑니다.
먼저 움직이는 대상을 고르고, 그 대상이 어느 queue, cache, buffer, log에 들어가는지 말한 뒤, 어떤 성공 조건이 어느 계층에서 닫히는지 나누어야 합니다.
예를 들어 Kafka의 `acks=all`은 replica protocol의 성공 조건이고, `fsync()`는 파일 변경을 storage 쪽에 더 강하게 반영해 달라는 OS 요청입니다.
둘은 모두 "성공"처럼 보이지만 같은 증거를 남기지 않습니다.

`audit/claim_review.md`는 이 코퍼스 자체에 대한 검수 기록입니다. 학습 본문이 아니라, 어떤 claim이 어떤 공격을 받았고 어떤 수리를 거쳐 살아남았는지 추적하기 위한 장부입니다.

OS 교과서 주제군은 chapter-by-chapter 요약이 아니라 topic-family coverage로 반영했습니다. OSTEP의 virtualization/concurrency/persistence/security 축, CSAPP의 exceptional control flow/virtual memory/system I/O/network/concurrency 축, OSPP와 공룡책 계열의 process/thread/scheduling/synchronization/deadlock/memory/storage/I/O/filesystem/protection/security/virtualization 관점을 Linux man-pages, kernel docs, 공식 제품 문서, 논문 근거로 다시 썼습니다.
