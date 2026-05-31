# OS Kernel and Distributed Systems Deep Dive

이 코퍼스는 OS 커널, 분산 시스템 원리, Kafka, Cassandra, Spark를 따로 외우는 문서가 아닙니다. 목표는 한 백엔드 개발자가 `write()` 하나에서 시작해 page cache, fsync, replication, quorum, shuffle, checkpoint까지 이어지는 상태 이동을 자기 말로 설명하게 만드는 것입니다.

첫 번째 기준은 self-contained입니다. 다른 문서를 이미 읽었다고 가정하지 않습니다. 같은 개념이 앞뒤에서 다시 나오더라도, 그 문서 안에서 필요한 만큼 다시 설명합니다. 두 번째 기준은 prose-first입니다. `질문/직관/실패 모드` 같은 내부 작성 체크리스트를 본문 목차로 반복하지 않고, 독자가 읽는 순서대로 원인, 역사, 작은 상태, 내부 경로, 깨지는 지점, 확인 방법이 자연스럽게 이어지게 씁니다.

## 읽는 순서

1. [00_index_and_learning_path.md](00_index_and_learning_path.md)에서 전체 지도를 잡습니다.
2. [01_os_kernel_foundations.md](01_os_kernel_foundations.md)에서 한 머신 안의 CPU, 메모리, 파일, 네트워크가 커널을 통해 어떻게 관리되는지 배웁니다.
3. [02_distributed_system_foundations.md](02_distributed_system_foundations.md)에서 여러 머신 사이의 실패, 시간, 순서, 복제, 일관성, 회복을 배웁니다.
4. [03_kafka_deep_dive.md](03_kafka_deep_dive.md), [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md), [05_spark_deep_dive.md](05_spark_deep_dive.md)를 읽으며 같은 원리가 제품 내부 구조에서 어떻게 다르게 나타나는지 봅니다.
5. [06_cross_system_comparison.md](06_cross_system_comparison.md)과 [07_interview_reasoning_playbook.md](07_interview_reasoning_playbook.md)에서 세 시스템을 비교하고 면접식 짧은 답변으로 압축합니다.
6. [08_experiments_and_observability.md](08_experiments_and_observability.md), [09_glossary.md](09_glossary.md), [10_source_ledger.md](10_source_ledger.md)는 실험, 용어, 근거를 확인할 때 함께 봅니다.

## 통과 기준

이 문서 묶음을 제대로 읽었다면 다음 질문에 외운 키워드가 아니라 경로로 답할 수 있어야 합니다.

- `write()`가 성공해도 왜 아직 durable write가 아닐 수 있는가?
- Kafka가 빠르다는 말은 page cache, sequential append, batch, sendfile 중 어떤 상태 이동을 줄였다는 뜻인가?
- Cassandra의 quorum read/write는 왜 "항상 최신"이라는 말로 끝나지 않는가?
- Spark shuffle이 느릴 때 왜 executor heap, spill file, network, skew를 같이 봐야 하는가?
- 장애 질문이 오면 어떤 상태가 어디에 기록되었고, 실패 후 어디서 다시 시작할 수 있는지 어떻게 추론하는가?

`audit/claim_review.md`는 이 코퍼스 자체에 대한 검수 기록입니다. 학습 본문이 아니라, 어떤 claim이 어떤 공격을 받았고 어떤 수리를 거쳐 살아남았는지 추적하기 위한 장부입니다.
