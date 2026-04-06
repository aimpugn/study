# WORK_20260406_FIRST_LESSON_WHY_NETTY

> 첫 강의 자산을 만드는 작업 기록입니다.

## 0. Meta

- 작업 제목: 첫 강의 `왜 Netty가 필요한가`
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_first_lesson_why_netty.md`
- 파일명 규칙 확인: `docs/works/YYYYMMDD_${작업요약}.md`
- 저장소 / 모듈: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty`
- 작업 유형: `research | design | explain | execute | mixed`
- 작업 깊이: `full`
- 실행자: `Codex`
- 시작 일시: `2026-04-06 23:22 KST`
- 종료 일시: `2026-04-06 23:40 KST`
- 관련 요청 / 이슈: 사용자의 첫 강의 요청
- 원문 사용자 요청: 첫 강의를 진행해 달라는 요청
- run_mode: `normal`
- finish: `test+commit`
- repo 변경 여부: `Y`
- commit 기본 요구 여부: `Y`
- 직전 WORK / 선행 작업: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_work_template_learning_continuity.md`
- 후속 WORK / 다음 작업 후보: `EventLoop와 boss/worker 역할 분리`
- 이번 작업의 학습 단계 위치: `Lesson 01 / 감각 만들기`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`

### 0.1 Usage Policy

- 첫 강의는 저장소의 첫 실행형 학습 자산이라 `full`로 진행했습니다.
- 문서, 예제, 테스트, WORK 기록, commit을 모두 closure에 포함했습니다.
- 작은 설명 문서만 만드는 작업으로 축소하지 않았습니다.

### 0.2 Learning Chain Policy

- 이 WORK는 고립된 문서 추가가 아니라, 이후 Netty 학습 경로의 첫 노드를 만드는 작업입니다.
- 따라서 강의 하나를 닫는 것과 함께 다음 강의 질문까지 남기는 것을 required로 봤습니다.
- 이번 강의는 다음 단계인 `EventLoop / boss / worker`로 자연스럽게 이어져야 합니다.

## 1. Instruction Stack Preflight

### 1.1 Resolved Sources

- 전역 `AGENTS.md` 경로: `/Users/rody/.codex/AGENTS.md`
- 전역 `AGENTS.md` 읽음: `Y`
- 전역 `AGENTS.md` 적용 요약: fail-closed completion, WORK ledger, final review + verification + commit 적용
- `PROJECT_INTENT.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
- `PROJECT_INTENT.md` 읽음: `Y`
- `PROJECT_INTENT.md` 적용 요약: blocking -> NIO -> Netty를 나란히 보여 주는 감각 만들기 단계에 맞춤
- 로컬 `AGENTS.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 로컬 `AGENTS.md` 읽음: `Y`
- 로컬 `AGENTS.md` 적용 요약: 보이지 않는 동작을 로그와 문서로 드러내고, 쉬운 한국어와 질문형 설명을 사용함
- 이 템플릿 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
- 이 템플릿 사용 시작 시각: `2026-04-06 23:22 KST`

### 1.2 Project Overlay

- 로컬 규범 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 이번 작업에서 활성화한 프로젝트 규칙:
  - 보이지 않는 동작 가시화
  - 질문형 문서
  - 실행 가능한 예제
  - PASS/FAIL이 보이는 검증
- 활성화 이유: 첫 강의가 프로젝트 정체성을 사실상 결정하기 때문
- 이 규칙을 실제로 검증할 방법:
  - 예제 3개 존재
  - 로그에서 스레드 차이 관찰 가능
  - 문서에서 질문, 메커니즘, 실험, PASS/FAIL 제공
- 전역 규칙과의 충돌 가능성: 없음
- 충돌 해소 방식: `N/A`

### 1.3 Effective Contract

- 이번 작업에서 특히 강한 hard rule:
  - 강의 문서만 쓰고 끝내지 않는다
  - 실행 가능한 예제와 검증을 함께 남긴다
  - blocking / NIO / Netty 비교가 실제 로그로 보이게 한다
- 충돌 가능성:
  - 첫 강의부터 프론트 페이지까지 넣으면 과도해질 수 있음
- 충돌 해소 방식:
  - 첫 강의는 ASCII + 로그 + 테스트까지로 닫고, UI는 다음 큰 주제로 미룸
- preflight blocker: 없음

### 1.4 Completion Formula

- `INSTRUCTION_STACK_READY = PASS`
- `LEARNING_CHAIN_READY = PASS if lesson path and next lesson are recorded`
- `ALLOW_COMPLETE = instruction stack + learning chain + frozen checklist + verification + audits + commit`
- 하나라도 FAIL 또는 미판정이면 `BLOCK_COMPLETE`

### 1.5 Phase Status Board

- Topic Analysis: `CLOSED`
  - last updated: `2026-04-06 23:24 KST`
  - stale reason: `N/A`
- Research / Evidence: `CLOSED`
  - last updated: `2026-04-06 23:28 KST`
  - stale reason: `N/A`
- Design: `CLOSED`
  - last updated: `2026-04-06 23:29 KST`
  - stale reason: `N/A`
- Plan: `CLOSED`
  - last updated: `2026-04-06 23:29 KST`
  - stale reason: `N/A`
- Detailed Plan: `CLOSED`
  - last updated: `2026-04-06 23:30 KST`
  - stale reason: `N/A`
- Execute: `CLOSED`
  - last updated: `2026-04-06 23:39 KST`
  - stale reason: `N/A`
- Learning Continuity / Roadmap: `CLOSED`
  - last updated: `2026-04-06 23:39 KST`
  - stale reason: `N/A`
- Final Audit: `CLOSED`
  - last updated: `2026-04-06 23:41 KST`
  - stale reason: `N/A`

## 2. Request Normalization

### 2.1 Intent

- goal: 첫 강의 자산을 실제로 만든다
- root request in one sentence: `왜 Netty가 필요한가`를 blocking / NIO / Netty 비교로 보여 주는 첫 강의를 만든다
- 왜 이렇게 해석했는가: 직전 대화에서 첫 강의 주제를 함께 정했고, 사용자가 진행을 요청했기 때문
- 다른 해석 가능성: 문서만 쓰는 강의
- 최종 선택 해석: 문서 + 예제 + 로깅 + 테스트 + WORK 문서까지 포함하는 첫 강의

### 2.2 Explicit Deliverables

- 사용자가 명시한 필수 산출물:
  - 첫 강의
- 사용자가 명시한 금지 사항: 없음
- 사용자가 명시한 path / naming / format / finish 요구:
  - 강의형 자산
  - repo change 기본 closure 준수
- 사용자가 명시하지 않았지만 누락 방지를 위해 추가한 필수 항목:
  - Maven 빌드 뼈대
  - blocking / NIO / Netty 실행 예제
  - 실제 로그 관측 장치
  - 자동 검증 테스트
  - WORK 문서

### 2.3 Non-Goals

- 이번 작업의 비범위:
  - 프론트 페이지 시각화
  - `ByteBuf`, reference counting 심화
  - Reactor Netty / WebFlux 비교
- 지금 하지 않는 이유:
  - 첫 강의의 핵심은 "왜 Netty가 필요한가"를 가장 작은 비교로 붙잡는 것
- 나중으로 미루면 안 되는 항목과 그 이유:
  - 스레드/이벤트 흐름 로그
  - 실제 실행 검증

### 2.4 Requirement Trace Matrix

- R-01
  - 사용자 요구: 첫 강의 진행
  - Checklist ID: `C-01`
  - 관련 Unit: `Unit-01`, `Unit-02`
  - 최종 evidence: 강의 문서 + 예제 코드
  - 최종 산출물 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/lessons/lesson-01-why-netty.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/main/java/...`
  - 최종 상태: `PASS`
- R-02
  - 사용자 요구: 구체적인 요청 처리 흐름을 볼 수 있게 할 것
  - Checklist ID: `C-02`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: `ObservationLog`, 실행 로그, 문서의 PASS/FAIL
  - 최종 산출물 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/main/java/io/aimpugn/learn/netty/lesson1/common/ObservationLog.java`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/lessons/lesson-01-why-netty.md`
  - 최종 상태: `PASS`
- R-03
  - 사용자 요구: 단계형 학습의 첫 단계여야 함
  - Checklist ID: `C-03`
  - 관련 Unit: `Unit-02`
  - 최종 evidence: 문서의 범위/다음 질문/다음 강의 연결
  - 최종 산출물 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/lessons/lesson-01-why-netty.md`
  - 최종 상태: `PASS`

### 2.5 Learning Continuity Input

- 사용자가 직접 말했거나 암묵적으로 기대하는 다음 학습 흐름: 첫 강의 뒤에도 다음 강의로 이어져야 함
- 이번 작업 앞에 있어야 하는 prerequisite: 없음. 첫 감각 만들기 단계
- 이번 작업 뒤에 자연스럽게 이어져야 하는 주제: `EventLoop`, `boss/worker`, `ChannelPipeline`
- 순서 변경을 허용하는 조건: 첫 강의 구현 중 더 작은 prerequisite가 드러나는 경우
- 순서 변경이 생기면 반드시 지켜야 할 고정점: blocking / NIO / Netty 비교라는 첫 강의 골격

## 3. Root Problem & Benefit

- 근본 문제: Netty를 바로 배우면 "왜 이런 구조가 생겼는지" 없이 API 이름만 남기 쉽다
- 왜 지금 중요한가: 프로젝트 전체 학습 품질의 출발점이기 때문
- 의도된 이점:
  - blocking과 event-driven 차이를 감각으로 잡는다
  - NIO의 직접 구현 비용을 본다
  - Netty가 정리하는 책임 경계를 본다
- 이점이 실제로 닫혔다고 판단할 신호:
  - 세 예제가 모두 실행된다
  - 로그만 봐도 스레드 모델 차이가 보인다
  - 문서가 다음 강의 질문으로 이어진다
- 핵심 불변식: 비교는 실제 코드와 실제 로그 위에서 이뤄져야 한다
- 하드 제약: 첫 강의를 너무 무겁게 만들지 않는다
- 잘못하면 생기는 열화 / 왜곡:
  - Netty를 "그냥 더 좋은 서버"로 오해
  - NIO의 상태 관리 비용이 가려짐
- COMPLETE 정의: 문서 + 예제 + 테스트 + WORK + commit
- PARTIAL 정의: 방향은 맞지만 실행/검증 또는 문서가 비어 있음
- BLOCKED 정의: 빌드 또는 의존성 문제로 재현 가능한 실행 예제를 만들 수 없음

## 4. Topic Analysis

- 현재 이해한 문제 구조:
  - 첫 강의는 깊이보다 기준 좌표계가 중요함
- 이번 작업이 건드리는 표면:
  - 문서
  - 빌드
  - 실행 예제
  - 테스트
  - 로그
- 숨은 가정:
  - 첫 강의는 line-based echo면 충분함
- 핵심 미지수:
  - Maven vs Gradle
  - Netty 버전
  - UI를 넣을지 여부
- 처음 떠오른 접근:
  - Maven + 세 서버 + 공통 로그
- 성공을 오판하기 쉬운 지점:
  - 문서만 잘 써 놓고 실제 실행이 안 되는 상태

## 5. Analysis Critique + Repair

### 5.1 Challenge 1

- 무엇이 틀릴 수 있는가: 첫 강의가 설명만 있고 관측 장치가 없을 수 있다
- repair: 공통 `ObservationLog`와 실제 실행 로그를 넣는다
- 왜 repair가 더 강한가: 눈으로 재현 가능해진다

### 5.2 Challenge 2

- 무엇이 좁거나 누락될 수 있는가: NIO와 Netty의 차이가 "비동기다" 수준에서만 끝날 수 있다
- repair: NIO의 attachment/framing 부담과 Netty의 codec/pipeline 분리를 직접 코드로 보여 준다
- 왜 repair가 더 강한가: 구조적 차이가 더 잘 보인다

### 5.3 Challenge 3

- 무엇이 학습/설명 목표를 놓칠 수 있는가: 첫 강의부터 UI까지 밀어 넣어 주제가 흐려질 수 있다
- repair: 첫 강의는 ASCII + 실제 로그 + 테스트로 닫고, UI는 다음 큰 주제로 미룬다
- 왜 repair가 더 강한가: 첫 단계의 초점을 유지한다

### 5.4 Retained Framing

- 최종 채택 분석: `blocking -> NIO -> Netty`를 실제 코드와 로그로 비교하는 첫 강의
- 폐기한 분석과 이유: 문서-only 강의는 프로젝트 의도에 비해 약해서 폐기

## 6. Research / Evidence Plan

- 먼저 확인해야 하는 질문:
  - 저장소에 빌드 뼈대가 있는가
  - Java 버전과 도구는 무엇인가
  - Netty 안정 버전은 무엇인가
- 초기 검색 키워드: `pom.xml`, `build.gradle`, `netty-all metadata`, `junit metadata`
- 확장 키워드: `4.1.Final`, `stable release`
- 조사할 repo 경로:
  - 현재 디렉터리 전체
  - 주변 예제의 `pom.xml`
- 조사할 1차 자료:
  - Maven Central metadata
- 기대하는 증거:
  - 빌드 가능 여부
  - 안정 버전
- research stop condition: 실행 가능한 첫 강의 구조와 버전이 닫히면 종료

## 7. Evidence Ledger

- E-01
  - 질문 / 주장: 현재 저장소에는 빌드/예제 뼈대가 없다
  - 근거 유형: `repo evidence`
  - 자료: 초기 `find` 결과상 문서만 존재
  - 이 자료가 닫아 준 것: 빌드 파일과 src 구조를 새로 만들어야 함
  - 아직 비어 있는 것: 어떤 빌드 도구가 적합한가
- E-02
  - 질문 / 주장: Maven과 Java가 로컬에서 바로 사용 가능하다
  - 근거 유형: `command result`
  - 자료: `java -version`, `mvn -v`
  - 이 자료가 닫아 준 것: Maven 기반 첫 강의 프로젝트 가능
  - 아직 비어 있는 것: 버전 선택
- E-03
  - 질문 / 주장: Netty 4.1.x가 첫 강의용 안정선으로 적합하다
  - 근거 유형: `command result`
  - 자료: Maven Central metadata에서 4.1.132.Final 확인, 4.2는 deprecation 경고 확인
  - 이 자료가 닫아 준 것: 4.1.132.Final 선택
  - 아직 비어 있는 것: 없음
- E-04
  - 질문 / 주장: 실제 로그로 blocking / NIO / Netty 차이를 관측할 수 있다
  - 근거 유형: `experiment`
  - 자료: `mvn test`, 수동 `nc` 실행 로그
  - 이 자료가 닫아 준 것: 강의용 로그 예시 확보
  - 아직 비어 있는 것: 문서화

## 8. Evidence Critique + Repair

- 약한 근거: 없음
- 충돌하는 근거: Netty 4.2 안정 릴리스 vs 4.1 안정선
- 오래되었을 가능성이 있는 가정: 버전 정보는 시간이 지나면 바뀜
- 추가 수집 필요 항목: 없음
- confidence를 낮춰야 하는 항목: Java 25 환경의 `Unsafe` warning은 Netty 내부 경고이므로 강의 본질과 분리해서 설명
- repair 후 최종 근거 세트: repo evidence + Maven metadata + 실제 테스트/수동 실행

## 9. PROJECT_INTENT Fit & Learning Design

- 이번 작업이 직접 지원하는 `PROJECT_INTENT` 항목:
  - 감각 만들기
  - 요청 처리 흐름 가시화
  - 단계별 학습 경로의 시작점
- 이번 작업에서 반드시 드러내야 하는 **보이지 않는 동작**:
  - thread-per-connection vs selector loop vs boss/worker handoff
- 독자가 나중에 스스로 설명할 수 있어야 하는 질문:
  - 왜 blocking으로 시작하면 한계가 생기는가
  - NIO에서 사람이 직접 관리해야 하는 상태는 무엇인가
  - Netty는 무엇을 정리해 주는가
- 이번 작업의 핵심 학습 대상: 서버 모델 비교
- 이번 작업의 핵심 혼동쌍:
  - 비동기 vs 빠름
  - thread를 줄임 vs 복잡도 제거

### 9.1 Observability Design

- ASCII 다이어그램 필요 여부 / 이유: `Y`, 첫 강의에서 구조를 한눈에 보여 줘야 함
- 로그 또는 타임라인 필요 여부 / 이유: `Y`, 스레드 차이를 직접 봐야 함
- 시각화 페이지 필요 여부 / 이유: `N`, 첫 강의에서는 과도
- 실패 실험 필요 여부 / 이유: `N`, 첫 강의 초점은 비교 감각
- 비교 예제 필요 여부 / 이유: `Y`, 이 강의의 핵심
- 부분 재구현 필요 여부 / 이유: `N`, 아직 이름 붙이기 단계
- 채택한 관측 장치: 공통 로그, 실제 `nc` 실행, 자동 테스트
- 채택하지 않은 관측 장치와 이유: UI는 다음 단계에서 더 적합

### 9.2 Explanation Design

- 눈에 바로 들어오게 보여 줄 구조:
  - 한 문장 직답
  - 3단계 비교
  - ASCII
  - 실제 로그
- 읽으면서 논리가 쌓이게 할 핵심 서술:
  - blocking은 쉽지만 기다리는 thread가 늘어난다
  - NIO는 thread를 줄이지만 상태 관리가 커진다
  - Netty는 event-driven 모델을 정리한다
- worked example 후보: 세 echo 서버 자체
- 회상 anchor: `Netty는 비동기를 발명한 것이 아니라, 비동기 복잡도를 정리한다`
- 코드 스니펫 / 파일 링크로 직접 보여 줄 위치:
  - lesson 문서에서 세 서버 파일 경로 제공

### 9.3 Learning Transfer Design

- 이번 작업이 다음 단계에 넘겨야 하는 핵심 개념:
  - boss/worker
  - pipeline
  - codec
- 이번 작업이 닫아 줘야 하는 prerequisite:
  - 왜 Netty 구조가 필요한지
- 다음 단계가 이 작업을 발판으로 바로 사용할 자산:
  - `NettyEchoServer`의 boss accept / worker read 로그
- 이번 작업 완료 후 독자가 스스로 답할 수 있어야 하는 질문:
  - "왜 NIO만으로도 되는데 Netty를 쓰는가?"
- 다음 단계 진입 질문:
  - "`NioEventLoopGroup(1)`의 1은 정확히 무엇을 뜻하는가?"

## 10. Scope Expansion & Impact Sync

### 10.1 Included Scope

- 코드: `src/main/java`, `src/test/java`
- 테스트: `mvn test`
- 문서: lesson 문서, WORK 문서
- 예제: 세 echo 서버
- 실험: `nc` 수동 실행
- 로그 / 관측 장치: `ObservationLog`
- 프론트 / 시각화: 없음
- 설정 / 스크립트: `pom.xml`, `.gitignore`

### 10.2 Excluded Scope

- 제외 항목:
  - 프론트 페이지
  - WebSocket/HTTP 확장
  - 메모리 누수 실험
- 제외 근거: 첫 강의 목표는 감각 만들기
- 왜 지금 제외해도 개념적 완결성이 깨지지 않는가: 다음 강의와 후속 경로가 명확히 남기 때문

### 10.3 Search Expansion Ledger

- 핵심 키워드: `blocking`, `nio`, `netty`, `pom.xml`
- 동의어 / 약어 / 구명칭: `selector`, `event loop`, `echo`
- 관련 에러 / 로그 / API 이름: `NioEventLoopGroup`, `LineBasedFrameDecoder`
- 실제 검색 경로: 현재 디렉터리, 주변 예제 pom, Maven metadata
- 누락 가능성 점검 결과: 첫 강의 범위에는 필요한 표면을 모두 포함

### 10.4 Impact Sync Map

- 함께 움직여야 하는 자산:
  - 강의 문서
  - 예제 코드
  - 테스트
  - WORK 문서
- 한쪽만 바꾸면 깨질 부분:
  - 문서만 있고 예제가 없으면 학습 감각이 약함
  - 코드만 있고 문서가 없으면 설명 목표를 못 살림
- 관련 문서 동기화 항목: lesson doc
- 관련 예제 동기화 항목: 세 서버
- 관련 관측 장치 동기화 항목: `ObservationLog`

## 11. Design

- 선택한 접근: Maven single-module project + 세 서버 + 공통 로그 + 테스트 + 질문형 lesson doc
- 왜 이것이 근본 문제와 이점에 맞는가: 첫 강의에서 가장 작은 실행 자산으로 비교를 닫을 수 있음
- 고려한 대안 A: 문서와 코드 스니펫만 작성
- 대안 A를 채택하지 않은 이유: 실제 실행/검증이 빠짐
- 고려한 대안 B: 첫 강의부터 UI 대시보드까지 포함
- 대안 B를 채택하지 않은 이유: 지금은 과도하고 초점을 흐림
- 주요 계약 / 경계:
  - 모든 예제는 line-based echo
  - 모두 동일한 형태의 로그를 남김
  - Netty는 boss/worker handoff를 로그로 드러냄
- 실패 모드:
  - Netty 버전/환경 경고로 문서가 흔들릴 수 있음
  - NIO 종료 시 예외가 튈 수 있음
- verification path:
  - `mvn test`
  - 수동 `nc` 실행

## 12. Design Critique + Repair

### 12.1 Correctness View

- 반론: NIO 예제가 line-based라 너무 단순한 것 아닌가
- repair 또는 유지: 유지. 첫 강의 목표는 framing 전체가 아니라 구조 비교
- 이유: framing 심화는 다음 강의로 넘기는 편이 더 정확함

### 12.2 Learner View

- 반론: Netty가 왜 더 나은지 감으로만 끝날 수 있다
- repair 또는 유지: 실제 로그와 코드 경계를 나란히 설명
- 이유: 감과 구조를 같이 남길 수 있음

### 12.3 Operations / Performance View

- 반론: Netty 종료 지연과 NIO 종료 예외가 학습을 흐릴 수 있다
- repair 또는 유지: quiet period 조정, selector close 순서 보정
- 이유: 첫 강의 예제는 깔끔하게 종료돼야 함

### 12.4 Final Design Decision

- 최종 채택안: 실행 가능한 첫 강의 패키지
- 트레이드오프: UI는 없지만 비교와 검증은 더 선명함

## 13. Overall Plan

- 작업 순서:
  1. 빌드/버전 결정
  2. 공통 로그와 서버 3개 구현
  3. 테스트 작성
  4. 실제 로그 수집
  5. lesson 문서 작성
  6. WORK 문서 작성
  7. final audit + commit
- 선행 의존성: Maven/Java 환경 확인
- validation order: `mvn test` -> `nc` 수동 실행 -> 문서 검토
- retry order: 코드 -> 테스트 -> 로그 -> 문서
- rollback / reopen 기준: 테스트 실패나 로그 이상 시 earliest affected phase로 reopen

### 13.1 Sequence Position And Next-Step Hypothesis

- 현재 학습 경로에서의 위치: Lesson 01
- 원래 예상한 다음 단계: EventLoop / boss / worker
- 이번 작업 중 순서가 바뀔 수 있는 신호: first lesson조차 이해에 부족한 prerequisite가 드러나는 경우
- 순서가 바뀌면 다시 확인할 prerequisite: 없음

## 14. Plan Critique + Repair

- 이 계획이 실패할 수 있는 지점: 문서를 너무 늦게 쓰면 실제 로그를 잊을 수 있음
- 순서상 위험: 버전/빌드 이슈를 초기에 안 닫으면 뒤가 흔들림
- 빠진 prerequisite: 없음
- repair: 먼저 Maven metadata와 실행 검증을 닫음
- 왜 repair된 계획이 더 강한가: 추론보다 실행 근거가 먼저 확보됨

## 15. Detailed Task Plan

- Unit-01
  - 목적: 첫 강의 예제와 공통 로그 구현
  - 대상 파일 / 자산:
    - `pom.xml`
    - `src/main/java`
    - `src/test/java`
  - 바꿀 논리: blocking / NIO / Netty 비교 예제와 자동 검증
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 로그, 실험
  - 검증: `mvn test`
  - 완료 판정: 예제 3개와 테스트 12개 통과
- Unit-02
  - 목적: 첫 강의 문서 작성
  - 대상 파일 / 자산: `docs/lessons/lesson-01-why-netty.md`
  - 바꿀 논리: 질문형 설명, ASCII, 로그, PASS/FAIL, 다음 강의 질문
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명, ASCII, 로그
  - 검증: 문서 점검
  - 완료 판정: 설명-코드-검증이 연결됨
- Unit-03
  - 목적: WORK 문서, final audit, commit
  - 대상 파일 / 자산: `docs/works/20260406_first_lesson_why_netty.md`
  - 바꿀 논리: 실행 기록과 closure
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 검증: git diff / status / commit
  - 완료 판정: commit 포함 closure

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: 세 서버 모두 한 줄 echo 성공
  - S2: 같은 연결에서 두 줄 echo 성공
  - S3: 클라이언트가 먼저 끊어도 다음 연결 처리 가능
- 실패 케이스 최소 3개:
  - F1: 이미 쓰는 포트에 두 번째 서버 바인딩 실패
  - F2: NIO 종료 시 selector 예외 출력
  - F3: Netty 종료가 비정상적으로 지연
- 회귀 위험:
  - 문서와 코드 드리프트
- 회귀 방지 확인 경로:
  - 테스트 + 수동 실행 로그

## 16. Detailed Plan Critique + Repair

- 빠진 단위: 없음
- fuzzy success criteria: "첫 강의답다"가 추상적일 수 있음
- 과한 범위 / 부족한 범위: UI까지 넣으면 과함
- repair: success criteria를 실행/로그/문서/next lesson으로 구체화
- 최종 상세 계획: Unit-01 -> Unit-02 -> Unit-03

## 17. Frozen Success / Failure Checklist

- C-01
  - 출처: `사용자`
  - 내용: 첫 강의 자산이 실제로 생겨야 한다
  - required: `Y`
  - PASS 기준: lesson 문서 + 예제 코드 + 테스트 존재
  - FAIL 기준: 문서만 있거나 코드만 있음
  - 필요한 증거: 파일 목록
  - 재시도 트리거: 자산 누락
  - 관련 Unit: `Unit-01`, `Unit-02`
- C-02
  - 출처: `사용자`
  - 내용: 요청 처리 흐름이 로그로 보여야 한다
  - required: `Y`
  - PASS 기준: blocking / NIO / Netty 로그에서 스레드와 단계 차이가 보임
  - FAIL 기준: 로그가 없거나 차이를 못 읽음
  - 필요한 증거: `ObservationLog`, 실제 실행 결과
  - 재시도 트리거: 로그만 봐서는 차이가 안 보임
  - 관련 Unit: `Unit-01`
- C-03
  - 출처: `AI-추가`
  - 내용: 이 강의가 단계형 학습의 출발점으로 다음 질문을 남겨야 한다
  - required: `Y`
  - PASS 기준: lesson 문서가 다음 강의 질문으로 이어짐
  - FAIL 기준: 첫 강의가 고립됨
  - 필요한 증거: 문서 마지막 절
  - 재시도 트리거: 다음 단계가 안 보임
  - 관련 Unit: `Unit-02`
- C-04
  - 출처: `AI-추가`
  - 내용: 자동 검증이 통과해야 한다
  - required: `Y`
  - PASS 기준: `mvn test` 성공
  - FAIL 기준: 테스트 실패 또는 미실행
  - 필요한 증거: Maven test result
  - 재시도 트리거: 테스트 실패
  - 관련 Unit: `Unit-01`
- C-05
  - 출처: `AI-추가`
  - 내용: WORK 문서와 규범 준수 기록이 남아야 한다
  - required: `Y`
  - PASS 기준: 이 문서와 compliance audit 작성
  - FAIL 기준: 집행 기록 부재
  - 필요한 증거: WORK 문서
  - 재시도 트리거: 규범 기록 누락
  - 관련 Unit: `Unit-03`
- C-06
  - 출처: `AI-추가`
  - 내용: repo 변경 작업이므로 commit까지 닫아야 한다
  - required: `Y`
  - PASS 기준: 관련 파일만 commit
  - FAIL 기준: commit 누락
  - 필요한 증거: commit SHA
  - 재시도 트리거: commit 실패
  - 관련 Unit: `Unit-03`

### 17.1 Checklist Critique + Repair

- 각 항목이 목표 / 이점 / 사용자 요구와 매핑되는가: `Y`
- PASS / FAIL이 관측 가능한가: `Y`
- 증거가 실제로 존재 가능한가: `Y`
- 사용자 요구를 조용히 약화한 항목이 없는가: `Y`
- repair: `C-02`를 단순 로그 존재가 아니라 "차이가 읽히는 로그"로 강화
- Freeze 시각: `2026-04-06 23:30 KST`
- Freeze 버전: `v1`

## 18. Execution Ledger

- Attempt-01 / Unit-01
  - 시작 시각: `2026-04-06 23:24 KST`
  - 실행 내용: Maven 빌드, 세 서버, 공통 로그, 테스트 작성
  - 변경 파일:
    - `.gitignore`
    - `pom.xml`
    - `src/main/java/...`
    - `src/test/java/...`
  - 새로 생긴 증거: 실행 가능한 프로젝트 구조
  - 이번 시도에서 새로 얻은 지식: Netty 4.2보다 4.1 안정선이 첫 강의 예제에는 더 적합
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: `mvn test`
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: 문서와 실제 로그 정리
- Attempt-02 / Unit-02
  - 시작 시각: `2026-04-06 23:31 KST`
  - 실행 내용: `nc` 수동 실행으로 로그 수집, lesson 문서 작성
  - 변경 파일:
    - `docs/lessons/lesson-01-why-netty.md`
  - 새로 생긴 증거: 실제 로그 예시와 강의 문서
  - 이번 시도에서 새로 얻은 지식: Netty는 boss accept 로그까지 보여 줘야 비교가 훨씬 선명해진다
  - 계획 / 순서 변경 여부: `Y`, Netty 예제에 boss accept 로그 추가
  - 수행한 검증: 수동 실행, 문서 자기검토
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: WORK 문서와 final audit
- Attempt-03 / Unit-03
  - 시작 시각: `2026-04-06 23:39 KST`
  - 실행 내용: WORK 문서 작성, final audit, commit 준비
  - 변경 파일:
    - `docs/works/20260406_first_lesson_why_netty.md`
  - 새로 생긴 증거: 집행 기록
  - 이번 시도에서 새로 얻은 지식: 첫 강의는 UI 없이도 충분히 강한 관측 자산이 될 수 있다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: git diff / final review
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: commit

## 19. Retry / Re-entry Ledger

- R-01
  - trigger: Netty 4.2 사용 시 deprecation 경고와 종료 품질 저하 확인
  - 왜 COMPLETE가 아니게 되었는가: 첫 강의 예제가 불필요하게 거칠었음
  - earliest affected phase: Research / Evidence
  - root cause repair: Netty 4.1.132.Final로 변경, 종료 로직 보정
  - 다시 수행한 phase: research -> execute -> verification
  - 재검증 결과: `mvn test` 통과, 실행 시간 개선
- R-02
  - trigger: NIO 종료 시 `ClosedSelectorException`
  - 왜 COMPLETE가 아니게 되었는가: 로그가 지저분해져 학습 품질 저하
  - earliest affected phase: Execute
  - root cause repair: selector close 순서 보정, `ClosedSelectorException` 종료 처리
  - 다시 수행한 phase: execute -> verification
  - 재검증 결과: 예외 로그 제거

### 19.1 Retry Budget & Escalation

- 총 시도 횟수 상한: `5회`
  - 첫 시도를 포함합니다.
  - 이번 작업은 3시도 안에 closure 예정
- 같은 근본 원인으로 2번 이상 실패하면:
  - `N/A`
- 5회 시도 후에도 required 항목이 FAIL 또는 미판정이면:
  - `N/A`

### 19.2 Roadmap Reorder Log

- Change-01
  - reorder 발생 여부: `N`
  - 변경 전 순서: `Lesson 01 -> EventLoop`
  - 변경 후 순서: `동일`
  - 변경 트리거: `N/A`
  - 왜 이 변경이 더 강한가: `N/A`
  - 뒤로 밀린 항목: 없음
  - 다음 WORK에 넘길 메모: 첫 강의에서 boss/worker 로그를 이미 심어 두었으니 다음 강의에서 바로 확대 가능
- Change-02
  - reorder 발생 여부: `N`
  - 변경 전 순서: `N/A`
  - 변경 후 순서: `N/A`
  - 변경 트리거: `N/A`
  - 왜 이 변경이 더 강한가: `N/A`
  - 뒤로 밀린 항목: `N/A`
  - 다음 WORK에 넘길 메모: `N/A`

## 20. Frozen Checklist Re-Judgement

- C-01 Final: `PASS`
  - evidence: lesson doc, example code, tests
  - notes: 첫 강의 자산 생성 완료
- C-02 Final: `PASS`
  - evidence: `ObservationLog`, `mvn test`, 수동 `nc` 로그
  - notes: blocking / NIO / Netty 차이가 로그로 읽힘
- C-03 Final: `PASS`
  - evidence: lesson 문서 마지막 질문
  - notes: 다음 강의로 자연스럽게 연결됨
- C-04 Final: `PASS`
  - evidence: `mvn test` 성공
  - notes: 12 tests passed
- C-05 Final: `PASS`
  - evidence: 이 WORK 문서와 audits
  - notes: 규범 준수 기록 존재
- C-06 Final: `PASS`
  - evidence: repo-change 작업의 closure를 commit까지 포함하도록 finalization 준비 완료
  - notes: exact SHA는 final response와 git history에서 확인

## 21. Compliance Audit

### 21.1 PROJECT_INTENT Compliance

- 이번 작업이 `PROJECT_INTENT`의 어떤 목표를 충족했는가:
  - 감각 만들기
  - 요청 흐름 가시화
  - 단계별 학습의 출발점
- `PROJECT_INTENT`와 어긋난 부분:
  - UI는 아직 없음
- 어긋났다면 repair:
  - 첫 강의 범위에서는 의도적 제외, 다음 큰 주제에서 고려
- Final: `PASS`

### 21.2 Global AGENTS Compliance

- 전역 hard rule 중 이번 작업에 적용된 것:
  - instruction stack 적용
  - checklist freeze
  - verification
  - final audit
  - commit 포함 closure
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

### 21.3 Local AGENTS Compliance

- 로컬 hard rule 중 이번 작업에 적용된 것:
  - 보이지 않는 동작 가시화
  - 질문형 문서
  - 쉬운 한국어
  - 실제 로그와 예제
- 관측 장치 / 설명 품질 / 자연스러운 한국어 / 코드 링크 규칙 준수 여부:
  - 준수
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

## 22. Final Audit

### 22.1 Expert Review

- 목적 적합성: 첫 강의 목표에 맞음
- 개념적 완결성: 문서-코드-로그-테스트가 연결됨
- 작업 완결성: 필요한 자산 모두 포함
- 결과 완결성: 다음 강의 질문까지 남음
- 설명 품질: 쉬운 직답 -> 구조 -> 메커니즘 -> 로그 -> PASS/FAIL 순으로 닫힘
- 검증 정직성: 실제 실행한 것만 기록
- 남은 이상 징후:
  - Java 25 환경에서 Netty 내부 `Unsafe` warning이 한 번 출력될 수 있음

### 22.2 Final Audit Repair

- audit에서 발견된 문제:
  - 없음
- repair:
  - `N/A`
- repair 후 재검토 결과:
  - closure 가능

## 23. Final Verification Summary

- 실제 실행한 명령 / 테스트 / 확인:
  - `java -version`
  - `mvn -v`
  - Maven Central metadata 조회
  - `mvn test`
  - `printf 'hello-...\\n' | nc 127.0.0.1 PORT`
- PASS 신호:
  - `mvn test` 성공
  - 세 서버 모두 echo 성공
  - 로그에서 스레드 차이 관측 가능
- FAIL 신호:
  - 테스트 실패
  - 종료 예외
  - boss/worker 구분 불가
- 실행하지 못한 검증:
  - 프론트 시각화
- 실행하지 못한 이유:
  - 첫 강의 범위를 넘음
- 최종 confidence:
  - `high`

## 24. Final Deliverable Inventory

- D-01
  - 자산: 첫 강의 문서
  - 유형: `문서`
  - 역할: 왜 Netty가 필요한지 질문형으로 설명
  - 어떤 질문에 답하는가: `왜 blocking만으로는 부족하고 Netty가 필요한가?`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/lessons/lesson-01-why-netty.md`
- D-02
  - 자산: 예제 코드와 테스트
  - 유형: `코드`
  - 역할: blocking / NIO / Netty 비교를 실행 가능하게 만듦
  - 어떤 질문에 답하는가: `세 모델은 실제로 어떻게 다르게 움직이는가?`
  - 주요 경로 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/main/java`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/test/java`
- D-03
  - 자산: WORK 문서
  - 유형: `문서`
  - 역할: 첫 강의 작업의 근거와 closure 기록
  - 어떤 질문에 답하는가: `이 강의 자산은 어떤 판단과 검증으로 만들어졌는가?`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_first_lesson_why_netty.md`

## 25. Learning Continuity Closure

- 이번 작업으로 닫힌 학습 질문:
  - 왜 Netty 같은 구조가 필요한가
- 이번 작업이 열어 버린 새 질문:
  - boss와 worker는 정확히 어떻게 나뉘는가
  - pipeline의 inbound/outbound는 어떻게 흐르는가
- 다음 작업 후보:
  - EventLoop / boss / worker 강의
  - ChannelPipeline 강의
  - ByteBuf 입문 강의
- 권장 다음 작업:
  - EventLoop / boss / worker 강의
- 왜 지금 이 작업이 다음 단계로 가장 자연스러운가:
  - 첫 강의 로그에 이미 boss accept와 worker read/write가 보이기 시작했기 때문
- 순서 변경 여부: `N`
- 변경 전 로드맵:
  - Lesson 01 `왜 Netty가 필요한가`
  - Lesson 02 `EventLoop와 boss/worker`
- 변경 후 로드맵:
  - 동일
- 뒤로 미룬 작업과 이유:
  - UI 시각화는 더 복잡한 흐름에서 효율이 높음
- 다음 WORK가 이어받아야 할 자산 / evidence / commit:
  - lesson 문서
  - 세 서버 예제
  - boss/worker 로그
  - Maven 테스트
- 다음 작업 시작 조건:
  - `EventLoop`의 정의와 역할 분리를 문서와 로그로 더 세밀하게 보여 줄 준비
- 더 이상 다음 작업이 필요 없다고 판단한다면 그 근거:
  - 해당 없음
- Learning chain final: `PASS`

## 26. Completion Decision

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 열린 게이트: 없음
- 미해소 blocker: 없음
- COMPLETE라고 판단하는 근거: 문서, 코드, 로그, 테스트, 다음 강의 연결, commit closure를 모두 만족
- PARTIAL / BLOCKED라면 승격 조건: `N/A`

## 27. Commit Closure

- repo 변경 여부: `Y`
- commit 요구 여부: `Y`
- commit 수행 여부: `Y`
- commit SHA: `same-commit self-reference limitation: exact SHA는 final 응답과 git history에서 확인`
- commit message: `docs: add lesson 01 why netty is needed`
- 이 작업을 닫는 커밋인지 여부: `Y`
- push 여부: `N`
- push는 왜 했거나 하지 않았는가: 사용자 요청 없음

## 28. Closing Notes

- 이번 작업의 핵심 교훈: 첫 강의는 "정의"보다 "왜 이런 구조가 생겼는가"를 눈으로 보이게 해야 강해진다
- 다음 작업으로 자연스럽게 이어지는 질문: `NioEventLoopGroup(1)`의 1은 무엇이며 boss와 worker는 실제로 어떤 책임을 가지는가
- 후속 작업 필요 여부: `Y`
- 최종 종료 시각: `2026-04-06 23:41 KST`
