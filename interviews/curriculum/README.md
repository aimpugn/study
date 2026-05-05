# Interview Curriculum

이 디렉터리는 `intervie*.md` 원재료를 실제 인터뷰 준비용 대분류 구조로 재배치한 곳입니다.
대주제는 10개로 유지하고, `검색/NoSQL`은 `데이터베이스, 저장소, 검색/NoSQL` 안에 포함합니다.

현재 파일은 아직 딥 리라이트 완료본이 아니라 `원문 배치본`입니다.
다음 단계에서 각 소주제를 `짧은 직답 -> 깊은 메커니즘 -> 예시 -> 꼬리 질문 -> 검증/근거` 구조로 승격합니다.

## 대분류

- [01. 언어와 런타임](01-language-runtime.md): 9 source chunks
- [02. 동시성, 비동기, I/O](02-concurrency-async-io.md): 38 source chunks
- [03. OS, 커널, 컴퓨터 구조](03-os-kernel-computer-architecture.md): 10 source chunks
- [04. 네트워크와 웹 프로토콜](04-network-web-protocols.md): 8 source chunks
- [05. 보안과 암호학](05-security-cryptography.md): 6 source chunks
- [06. 데이터베이스, 저장소, 검색/NoSQL](06-database-storage-search-nosql.md): 17 source chunks
- [07. 메시징과 이벤트 기반 구조](07-messaging-event-driven.md): 5 source chunks
- [08. 분산 시스템과 아키텍처](08-distributed-systems-architecture.md): 5 source chunks
- [09. Spring과 백엔드 프레임워크](09-spring-backend-frameworks.md): 22 source chunks
- [10. 문제 해결, 코드 품질, 운영 실천](10-problem-solving-code-quality.md): 5 source chunks
- [Source Context And Question Bank](_source-context-and-question-bank.md): 8 context chunks

## 보존 규칙

- 원본 `intervie*.md` 파일은 source reservoir로 유지합니다.
- 원문 chunk는 source span, duplicate alias, SHA-256으로 추적합니다.
- curriculum 문서에서는 계층을 맞추기 위해 heading depth만 조정합니다.
- `_curriculum_manifest.json`은 모든 unique chunk가 어느 대분류/중분류/소분류에 놓였는지 기록합니다.

## 재생성

```sh
python3 interviews/tools/build_interview_curriculum.py
```
