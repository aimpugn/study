# OS Kernel and Distributed Systems Deep Dive

> 이 코퍼스는 OS 커널과 분산 시스템을 따로 외우는 대신, 한 머신 안의 자원 관리가 여러 머신 사이의 실패·순서·복제·회복 문제로 어떻게 확장되는지 따라가도록 만든 학습 경로입니다.
> Kafka, Cassandra, Spark는 제품 기능 목록이 아니라 서로 다른 목적을 가진 분산 상태 기계로 읽습니다.
> 각 문서는 다른 기존 문서를 읽지 않아도 핵심 질문, 상태 이동, 실패 모드, 검증 방법, 면접식 되묻기까지 복원할 수 있게 작성합니다.

이 디렉터리는 기존 interview 문서와 내용이 겹칠 수 있습니다. 최신 작업 기준은 중복 제거가 아니라 이 문서 묶음 자체의 학습 완결성입니다.

## 읽는 순서

1. [00_index_and_learning_path.md](00_index_and_learning_path.md)
2. [01_os_kernel_foundations.md](01_os_kernel_foundations.md)
3. [02_distributed_system_foundations.md](02_distributed_system_foundations.md)
4. [03_kafka_deep_dive.md](03_kafka_deep_dive.md)
5. [04_cassandra_deep_dive.md](04_cassandra_deep_dive.md)
6. [05_spark_deep_dive.md](05_spark_deep_dive.md)
7. [06_cross_system_comparison.md](06_cross_system_comparison.md)
8. [07_interview_reasoning_playbook.md](07_interview_reasoning_playbook.md)
9. [08_experiments_and_observability.md](08_experiments_and_observability.md)
10. [09_glossary.md](09_glossary.md)
11. [10_source_ledger.md](10_source_ledger.md)

## 사용법

먼저 `00`을 읽고 전체 지도를 잡은 뒤 `01`과 `02`를 천천히 읽습니다. `03~05`는 제품별 심화 문서지만, 각 문서가 계속 OS와 분산 시스템 원리로 되돌아오도록 구성되어 있습니다. `06`은 세 시스템을 비교하며 지식을 다시 묶고, `07`은 면접 답변을 암기가 아니라 추론으로 바꾸는 연습장입니다. `08`은 로컬에서 관찰 가능한 실험을 제공합니다.

각 큰 주제는 먼저 `>`로 시작하는 핵심 요약 문장을 둡니다. 이 문장은 장식 문구가 아니라 문서를 덮고도 다시 말할 수 있어야 하는 직답입니다. 그 다음에는 질문, 직관, 작은 예시, 상태 이동 trace, 내부 메커니즘, 실패 모드, 검증, 면접식 되묻기, 흔한 오해, active recall 순서로 내려갑니다.

