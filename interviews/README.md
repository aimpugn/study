# interviews

이 디렉터리는 경력 기술 인터뷰를 준비하기 위한 공간입니다.
목표는 예상 질문을 많이 모으는 데서 끝나지 않고, 질문이 들어왔을 때 짧은 시간 안에 먼저 핵심을 답하고, 이어서 필요한 만큼 깊게 설명할 수 있는 상태를 만드는 것입니다.

프로젝트 의도는 [PROJECT_INTENT.md](PROJECT_INTENT.md)에서 고정합니다.
대표 사용 장면과 산출물 역할은 [USECASE.md](USECASE.md)에서 고정합니다.

이 저장소에서 좋은 면접 답변은 암기한 정의를 빠르게 말하는 답변이 아닙니다.
면접관의 질문을 받으면 먼저 질문이 묻는 문제를 잡고, 그 문제가 어떤 상태를 움직이는지 보여 준 뒤, 실패하면 어디서 확인할 수 있는지까지 말해야 합니다.

```text
면접 질문
  -> 질문이 묻는 문제
  -> 숨은 상태나 깨지면 안 되는 약속
  -> 런타임 / 운영체제 / DB / 네트워크 / 서비스 실행 경로
  -> 비용이나 실패 신호
  -> 확인할 증거
```

예를 들어 "WebFlux로 바꾸면 빨라지나요?"라는 질문은 WebFlux 이름을 설명하라는 질문이 아니라, 요청이 이벤트 루프, 소켓 큐, blocking call, worker thread, downstream DB 대기 중 어디에서 시간을 쓰는지 설명하라는 질문입니다.
이 디렉터리의 문서들은 그 답변 경로를 만들기 위한 자산입니다.

## 읽기 시작점

| 목적 | 먼저 볼 문서 | 이 문서에서 붙잡을 것 |
| --- | --- | --- |
| 면접 전 빠른 복습 | [핵심 인터뷰 정리](core-interview-guide.md) | 질문을 받았을 때 `문제 -> 불변식 -> 상태 이동 -> 트레이드오프 -> 검증` 순서로 말하는 법 |
| 대주제별 원문 위치 확인 | [_question-index.md](_question-index.md) | 원문 질문이 어느 대분류와 원문 위치 범위(source span)에서 왔는지 찾는 법 |
| OS/분산 시스템 심화 | [OS Kernel And Distributed Systems Deep Dive](os-kernel-distributed-systems-deep-dive/README.md) | `write()`, page cache, socket buffer, log, quorum 같은 낮은 층 상태가 Kafka/Cassandra/Spark로 올라오는 경로 |
| DB 심화 | [Database Deep Dive](database-deep-dive/README.md) | page, buffer pool, WAL, MVCC, lock, replication, query plan이 답변으로 이어지는 경로 |
| 원문 확인 | [source](source/) | 원문 질문 은행과 시나리오를 보존하는 증거 표면 |

읽기 경로는 목적에 따라 나뉩니다.
면접 전 빠른 답변 조립은 [핵심 인터뷰 정리](core-interview-guide.md)에서 시작하고, 원문 위치와 대주제 분류를 확인할 때는 [_question-index.md](_question-index.md)와 아래 대주제 문서를 함께 봅니다.
OS, DB, 분산 시스템처럼 한 층 더 내려가야 하는 주제는 deep dive 문서로 이동합니다.

예를 들어 WebFlux 질문을 받았다면 아래처럼 읽습니다.

```text
WebFlux로 바꾸면 빨라지나요?
  -> core-interview-guide.md 6번에서 blocking / non-blocking / async / event loop 답변 골격을 잡는다
  -> _question-index.md에서 WebFlux, WebClient, Netty, epoll 원문 위치를 확인한다
  -> concurrency-async-io.md와 spring-backend-frameworks.md에서 원문 묶음을 본다
  -> thread-scheduling-java-spring.md나 OS deep dive에서 실제 thread, socket, event loop 상태로 내려간다
```

현재 학습용 기본 진입점은 이 디렉터리 바로 아래의 10개 대주제 문서입니다.
원문 질문 은행과 시나리오는 [source](source/)에 보관하고, 앞으로의 정리와 학습은 아래 대주제 문서를 기준으로 진행합니다.
`검색/NoSQL`은 독립 문서로 분리하지 않고 `데이터베이스, 저장소, 검색/NoSQL` 안에 포함합니다.

현재 대주제 문서는 아직 모든 소주제가 심화 재작성으로 닫힌 완성본이 아닙니다.
원문을 보존한 분류본 앞쪽에 `먼저 기억할 정리`, 숨은 상태 흐름, 확인 방법을 덧붙여 원문을 읽기 전에 비교축을 잡도록 했습니다.
다음 단계에서 각 소주제를 정식 답변 자산으로 승격할 때는 `짧은 직답 -> 깊은 메커니즘 -> 예시 -> 꼬리 질문 -> 검증/근거` 구조로 다시 씁니다.

## 전체 문서 읽기 기준

이 디렉터리의 Markdown은 성격이 서로 다릅니다.
`source/`의 raw 파일은 원문 증거이고, 대주제 문서는 원문 조각을 보존한 분류본이며, `database-deep-dive/`와 `os-kernel-distributed-systems-deep-dive/`는 정식 심화 학습 문서입니다.
따라서 모든 파일에 같은 양식을 강제로 입히지 않고, 각 문서가 맡은 역할에 맞게 아래 기준을 적용합니다.

- 정식 학습 문서는 먼저 기억할 구조를 평서문으로 정리하고, 질문은 그 뒤의 replay 장치로 둡니다.
- 원문 배치본은 SHA-256과 원문 위치 범위를 보존해야 하므로 원문 조각을 직접 고치지 않습니다. 대신 문서 앞쪽에서 읽을 축, 숨은 상태, 확인 방법을 먼저 잡습니다.
- `source/`는 정식 답변 품질로 다듬는 대상이 아니라 원문 확인과 재생성 검증 대상입니다. 여기서 뽑은 질문은 `_question-index.md`와 대주제 문서를 거쳐 정식 문서로 승격합니다.
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
- [실무 엔지니어를 위한 Linux 커널·하드웨어 내부 구조](linux-kernel-hardware-practical-internals.md): Linux/VMware 지표와 장애 분석을 면접 답변으로 연결하는 승격된 통합 교본입니다.

## 보존 규칙

- 원본 `intervie*.md` 파일은 `source/` 아래 원문 저장소로 유지합니다.
- 원문 조각은 source span, duplicate alias, SHA-256으로 추적합니다.
- 대주제 문서에서는 계층을 맞추기 위해 heading depth만 조정합니다.
- `_curriculum_manifest.json`은 모든 고유 원문 조각이 어느 대분류/중분류/소분류에 놓였는지 기록합니다.

## 재생성

```sh
python3 interviews/tools/build_interview_curriculum.py
```

이 명령은 대주제 문서를 다시 생성합니다.
README와 `_question-index.md`의 읽기 안내는 generator 템플릿에도 반영되어 있지만, 개별 대주제 문서 앞쪽의 수동 학습 bridge는 재생성 뒤에 다시 확인해야 합니다.
