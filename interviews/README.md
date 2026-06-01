# interviews

이 디렉터리는 경력 기술 인터뷰를 준비하기 위한 공간입니다.
목표는 예상 질문을 많이 모으는 데서 끝나지 않고, 질문이 들어왔을 때 짧은 시간 안에 먼저 핵심을 답하고, 이어서 필요한 만큼 깊게 설명할 수 있는 상태를 만드는 것입니다.

프로젝트 의도는 [PROJECT_INTENT.md](PROJECT_INTENT.md)에서 고정합니다.
대표 사용 장면과 산출물 역할은 [USECASE.md](USECASE.md)에서 고정합니다.

현재 학습용 기본 진입점은 이 디렉터리 바로 아래의 10개 대주제 문서입니다.
원문 질문 은행과 시나리오는 [source](source/)에 보관하고, 앞으로의 정리와 학습은 아래 새 curriculum 문서를 기준으로 진행합니다.
`검색/NoSQL`은 독립 문서로 분리하지 않고 `데이터베이스, 저장소, 검색/NoSQL` 안에 포함합니다.

현재 파일은 아직 딥 리라이트 완료본이 아니라 `원문 배치본`입니다.
다음 단계에서 각 소주제를 `짧은 직답 -> 깊은 메커니즘 -> 예시 -> 꼬리 질문 -> 검증/근거` 구조로 승격합니다.

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
