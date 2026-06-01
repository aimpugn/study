# interviews

이 디렉터리는 경력 기술 인터뷰를 준비하기 위한 공간입니다.
목표는 예상 질문을 많이 모으는 데서 끝나지 않고, 질문이 들어왔을 때 짧은 시간 안에 먼저 핵심을 답하고, 이어서 필요한 만큼 깊게 설명할 수 있는 상태를 만드는 것입니다.

프로젝트 의도는 [PROJECT_INTENT.md](PROJECT_INTENT.md)에서 고정합니다.
대표 사용 장면과 산출물 역할은 [USECASE.md](USECASE.md)에서 고정합니다.

현재 학습용 기본 진입점은 이 디렉터리 바로 아래의 10개 대주제 문서입니다.
원문 질문 은행과 시나리오는 [source](source/)에 보관하고, 앞으로의 정리와 학습은 아래 새 curriculum 문서를 기준으로 진행합니다.
`검색/NoSQL`은 독립 문서로 분리하지 않고 `데이터베이스, 저장소, 검색/NoSQL` 안에 포함합니다.

현재 root curriculum 파일은 아직 딥 리라이트 완료본이 아니라 `원문 배치본` 위에 수동 학습 bridge를 덧댄 상태입니다.
각 root 문서는 원문 chunk를 보존하되, 앞부분에 주제별 `먼저 기억할 정리`, 숨은 상태 흐름, 검증 anchor를 추가해 원문을 읽기 전에 비교축을 잡도록 했습니다.
다음 단계에서 각 소주제를 정식 답변 자산으로 승격할 때는 `짧은 직답 -> 깊은 메커니즘 -> 예시 -> 꼬리 질문 -> 검증/근거` 구조로 다시 씁니다.

## 전체 문서 읽기 기준

이 디렉터리의 Markdown은 성격이 서로 다릅니다. `source/`의 raw 파일은 원문 증거이고, root curriculum 문서는 원문 chunk를 보존한 분류본이며, `database-deep-dive/`와 `os-kernel-distributed-systems-deep-dive/`는 정식 심화 학습 문서입니다. 따라서 모든 파일에 같은 양식을 강제로 입히지 않고, 각 문서가 맡은 역할에 맞게 아래 기준을 적용합니다.

- 정식 학습 문서는 먼저 기억할 구조를 평서문으로 정리하고, 질문은 그 뒤의 replay 장치로 둡니다.
- 원문 배치본은 SHA-256과 source span을 보존해야 하므로 원문 chunk를 직접 고치지 않습니다. 대신 문서 앞쪽에서 읽을 축, 숨은 상태, 검증 anchor를 먼저 잡습니다.
- source reservoir는 정식 답변 품질로 다듬는 대상이 아니라 원문 확인과 재생성 검증 대상입니다. 여기서 뽑은 질문은 `_question-index.md`와 root curriculum을 거쳐 정식 문서로 승격합니다.
- 시스템 내부의 핵심이 queue, cache, buffer, table, lock, transaction log, scheduler, broker, coordinator 같은 보이지 않는 상태라면, 문서는 그 상태가 어디에 남고 누가 소비하는지 드러내야 합니다.

## 대분류

- [언어와 런타임](language-runtime.md): 9 source chunks
- [동시성, 비동기, I/O](concurrency-async-io.md): 38 source chunks
- [OS, 커널, 컴퓨터 구조](os-kernel-computer-architecture.md): 10 source chunks
- [네트워크와 웹 프로토콜](network-web-protocols.md): 8 source chunks
- [보안과 암호학](security-cryptography.md): 6 source chunks
- [데이터베이스, 저장소, 검색/NoSQL](database-storage-search-nosql.md): 17 source chunks
- [메시징과 이벤트 기반 구조](messaging-event-driven.md): 5 source chunks
- [분산 시스템과 아키텍처](distributed-systems-architecture.md): 5 source chunks
- [Spring과 백엔드 프레임워크](spring-backend-frameworks.md): 22 source chunks
- [문제 해결, 코드 품질, 운영 실천](problem-solving-code-quality.md): 5 source chunks
- [Source Context And Question Bank](source/_source-context-and-question-bank.md): 8 context chunks

## 승격된 심화 문서

- [소켓이란 무엇이고 소켓 프로그래밍이란 무엇인가](socket-programming.md)
- [OS 스레드, Java 스레드, Spring 스케줄링 실행 모델](thread-scheduling-java-spring.md)

## 보존 규칙

- 원본 `intervie*.md` 파일은 `source/` 아래 source reservoir로 유지합니다.
- 원문 chunk는 source span, duplicate alias, SHA-256으로 추적합니다.
- root curriculum 문서에서는 계층을 맞추기 위해 heading depth만 조정합니다.
- `_curriculum_manifest.json`은 모든 unique chunk가 어느 대분류/중분류/소분류에 놓였는지 기록합니다.

## 재생성

```sh
python3 interviews/tools/build_interview_curriculum.py
```

이 명령은 root curriculum 문서를 다시 생성합니다. 현재 root 문서 앞쪽의 수동 학습 bridge는 generator 산출물이 아니므로, 재생성 후에는 이 README의 읽기 기준에 맞춰 다시 적용해야 합니다.
