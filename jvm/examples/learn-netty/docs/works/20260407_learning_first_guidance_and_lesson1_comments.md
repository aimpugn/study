# WORK_20260407_LEARNING_FIRST_GUIDANCE_AND_LESSON1_COMMENTS

> 학습 목적과 비교형 설명 규칙을 지침에 더 강하게 반영하고, 그 규칙을 기준으로 lesson1 코드 주변 설명을 실제로 보강한 작업 기록입니다.

## 0. Meta

- 작업 제목: 학습 중심 지침 보강 및 lesson1 주석 적용
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_learning_first_guidance_and_lesson1_comments.md`
- 파일명 규칙 확인: `docs/works/YYYYMMDD_${작업요약}.md`
- 저장소 / 모듈: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty`
- 작업 유형: `design | explain | execute | mixed`
- 작업 깊이: `full`
- 실행자: `Codex`
- 시작 일시: `2026-04-07 00:10 KST`
- 종료 일시: `2026-04-07 00:23 KST`
- 관련 요청 / 이슈: 사용자의 학습/비교 강조 요청
- 원문 사용자 요청: 학습 목적과 비교의 중요성을 지침에 더 강하게 반영하고, 그 후 개선된 지침 기반으로 주석 설명을 적용
- run_mode: `normal`
- finish: `test+commit`
- repo 변경 여부: `Y`
- commit 기본 요구 여부: `Y`
- 직전 WORK / 선행 작업: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_enforce_code_adjacent_explanations.md`
- 후속 WORK / 다음 작업 후보: `lesson1 문서/주석 비교 보강 후 lesson2 EventLoop`
- 이번 작업의 학습 단계 위치: `첫 강의 비교 학습 품질 보강`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`

### 0.1 Usage Policy

- 이번 작업은 규범 보강과 실제 code-adjacent explanation 적용이 함께 있으므로 `full`로 다룹니다.
- 결과는 아래 세 축이 모두 닫혀야 합니다.
  1. `PROJECT_INTENT.md`, `AGENTS.md`, `AGENTS_WORK_TEMPLATE.md`의 학습/비교 강조
  2. lesson1 코드의 실제 주석/KDoc 반영
  3. 검증과 commit

### 0.2 Learning Chain Policy

- 이번 작업은 "코드 주변 설명 강제" 다음 단계로, 그 규칙을 실제 첫 강의 자산에 적용하는 작업입니다.
- 비교형 학습 규칙은 앞으로 lesson2 이후에도 계속 쓰이므로, 이번 작업은 일회성 주석 추가가 아니라 첫 exemplar를 만드는 성격을 가집니다.
- 다음 단계는 이 비교형 설명 방식을 lesson1 문서 보강이나 lesson2 설계에 그대로 이어붙이는 쪽이 자연스럽습니다.

## 1. Instruction Stack Preflight

### 1.1 Resolved Sources

- 전역 `AGENTS.md` 경로: `/Users/rody/.codex/AGENTS.md`
- 전역 `AGENTS.md` 읽음: `Y`
- 전역 `AGENTS.md` 적용 요약: fail-closed completion, final review, test+commit closure
- `PROJECT_INTENT.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
- `PROJECT_INTENT.md` 읽음: `Y`
- `PROJECT_INTENT.md` 적용 요약: 학습용 실험실, 비교 예제, 단계적 학습 경로
- 로컬 `AGENTS.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 로컬 `AGENTS.md` 읽음: `Y`
- 로컬 `AGENTS.md` 적용 요약: 학습 친화 설명, 비교형 설명, code-adjacent explanation
- 이 템플릿 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
- 이 템플릿 사용 시작 시각: `2026-04-07 00:10 KST`

### 1.2 Project Overlay

- 로컬 규범 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 이번 작업에서 활성화한 프로젝트 규칙:
  - 보이지 않는 동작을 보이게 만들기
  - 학습 목적 우선
  - 비교가 핵심인 주제는 차이와 이유까지 설명
  - 코드 주변 설명 강제
- 활성화 이유: 사용자가 "학습"과 "비교"를 지침에 더 강하게 강조해 달라고 요청했기 때문
- 이 규칙을 실제로 검증할 방법: 지침 문서에 비교 학습 규칙이 반영되고, lesson1 코드 주석이 blocking/NIO/Netty 차이를 직접 설명하는지 확인
- 전역 규칙과의 충돌 가능성: 주석을 과하게 늘리는 방향으로 오해될 수 있음
- 충돌 해소 방식: load-bearing 코드와 비교 핵심 지점에만 집중하는 주석으로 제한

### 1.3 Effective Contract

- 이번 작업에서 특히 강한 hard rule:
  - 학습 목적과 비교형 학습 규칙을 지침에 더 또렷하게 반영
  - lesson1 코드 주석이 `같은 문제 / 다른 메커니즘 / 차이 이유 / 관측 포인트`를 보여 줘야 함
  - 테스트 후 commit까지 closure
- 충돌 가능성:
  - 지침 분량이 늘어 실사용성이 떨어질 수 있음
  - 비교 설명이 길어져 코드 가독성을 해칠 수 있음
- 충돌 해소 방식:
  - 문서는 핵심 규칙만 추가
  - 코드는 class-level KDoc와 핵심 블록 설명에만 집중
- preflight blocker: 없음

### 1.4 Completion Formula

- `INSTRUCTION_STACK_READY = PASS`
- `LEARNING_CHAIN_READY = PASS`
- `CODE_EXPLANATION_READY = PASS`
- `COMPARISON_LEARNING_READY = PASS`
- `ALLOW_COMPLETE = docs updated + code comments applied + tests pass + commit`
- 하나라도 FAIL 또는 미판정이면 `BLOCK_COMPLETE`

### 1.5 Phase Status Board

- Topic Analysis: `CLOSED`
  - last updated: `2026-04-07 00:11 KST`
  - stale reason: `N/A`
- Research / Evidence: `CLOSED`
  - last updated: `2026-04-07 00:12 KST`
  - stale reason: `N/A`
- Design: `CLOSED`
  - last updated: `2026-04-07 00:13 KST`
  - stale reason: `N/A`
- Plan: `CLOSED`
  - last updated: `2026-04-07 00:13 KST`
  - stale reason: `N/A`
- Detailed Plan: `CLOSED`
  - last updated: `2026-04-07 00:14 KST`
  - stale reason: `N/A`
- Execute: `CLOSED`
  - last updated: `2026-04-07 00:19 KST`
  - stale reason: `N/A`
- Learning Continuity / Roadmap: `CLOSED`
  - last updated: `2026-04-07 00:16 KST`
  - stale reason: `N/A`
- Final Audit: `CLOSED`
  - last updated: `2026-04-07 00:20 KST`
  - stale reason: `N/A`

## 2. Request Normalization

### 2.1 Intent

- goal: 학습과 비교를 지침에 더 강하게 반영하고, 개선된 지침을 기준으로 lesson1 코드 주석을 보강한다
- root request in one sentence: 학습형·비교형 설명 규칙을 문서에 고정한 뒤, 첫 강의 코드에 그 규칙을 실제로 적용한다
- 왜 이렇게 해석했는가: 사용자가 먼저 지침 개선을 원했고, 그 다음 개선된 지침 기반으로 주석을 적용하라고 분명히 순서를 제시했기 때문
- 다른 해석 가능성: lesson1 코드에 바로 주석만 추가
- 최종 선택 해석: 의도/규범 보강 후 code-adjacent explanation 적용

### 2.2 Explicit Deliverables

- 사용자가 명시한 필수 산출물:
  - 학습 목적 강조가 더 강한 지침
  - 비교 설명 방식이 더 또렷한 지침
  - 개선된 지침을 기반으로 한 lesson1 주석 설명
- 사용자가 명시한 금지 사항: 없음
- 사용자가 명시한 path / naming / format / finish 요구:
  - 지침 먼저
  - 그 다음 주석 적용
- 사용자가 명시하지 않았지만 누락 방지를 위해 추가한 필수 항목:
  - WORK 문서
  - 테스트
  - commit

### 2.3 Non-Goals

- 이번 작업의 비범위:
  - lesson2 신규 구현
  - lesson1 문서 전면 개정
  - 프론트 시각화 추가
- 지금 하지 않는 이유:
  - 이번 요청은 지침 강화와 code-adjacent explanation 적용이 핵심
- 나중으로 미루면 안 되는 항목과 그 이유:
  - 비교형 학습 규칙 문서화
  - lesson1 exemplar 주석 보강

### 2.4 Requirement Trace Matrix

- R-01
  - 사용자 요구: 학습 목적임을 지침에 강조
  - Checklist ID: `C-01`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: `PROJECT_INTENT.md`, `AGENTS.md`
  - 최종 산출물 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - 최종 상태: `PASS`
- R-02
  - 사용자 요구: blocking/non-blocking 같은 비교에서 왜 그렇게 되고 뭐가 차이인지 설명
  - Checklist ID: `C-02`
  - 관련 Unit: `Unit-01`, `Unit-02`
  - 최종 evidence: 지침의 비교 학습 규칙 + lesson1 주석
  - 최종 산출물 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
    - `lesson1 code files`
  - 최종 상태: `PASS`
- R-03
  - 사용자 요구: 개선된 지침 기반으로 주석 설명 적용
  - Checklist ID: `C-03`
  - 관련 Unit: `Unit-02`
  - 최종 evidence: lesson1 class/block comments
  - 최종 산출물 / 파일: `lesson1 code files`
  - 최종 상태: `PASS`

### 2.5 Learning Continuity Input

- 사용자가 직접 말했거나 암묵적으로 기대하는 다음 학습 흐름: lesson1 비교 학습을 더 선명하게 만든 뒤 lesson2로 이어짐
- 이번 작업 앞에 있어야 하는 prerequisite: lesson1 example trio 존재
- 이번 작업 뒤에 자연스럽게 이어져야 하는 주제: lesson1 문서 보강 또는 lesson2 EventLoop
- 순서 변경을 허용하는 조건: lesson1에서 추가 설명 누락이 더 크게 발견되는 경우
- 순서 변경이 생기면 반드시 지켜야 할 고정점: 비교형 학습 규칙은 유지

## 3. Root Problem & Benefit

- 근본 문제: 기존 문서에는 학습과 비교가 이미 있었지만, "비교가 핵심일 때 무엇을 꼭 닫아야 하는가"와 "그 규칙이 코드 주석에 어떻게 반영되어야 하는가"가 더 또렷할 수 있었다
- 왜 지금 중요한가: lesson1이 blocking/NIO/Netty 비교를 처음 보여 주는 exemplar이기 때문
- 의도된 이점:
  - 지침이 학습형 저장소라는 성격을 더 강하게 드러냄
  - 비교형 설명 규칙이 WORK와 코드에 연결됨
  - lesson1 코드만 읽어도 각 구현 차이를 이해하기 쉬워짐
- 이점이 실제로 닫혔다고 판단할 신호:
  - 지침에 comparison-heavy topic 규칙 존재
  - lesson1 코드 주석이 blocking/NIO/Netty 차이와 관측 포인트를 설명
- 핵심 불변식: 이 저장소는 제품 코드보다 학습 복원성을 우선
- 하드 제약: 주석은 load-bearing 비교 포인트에 집중
- 잘못하면 생기는 열화 / 왜곡:
  - 지침은 좋아졌는데 코드가 그대로일 수 있음
  - 반대로 코드 주석만 늘고 비교 원칙은 흐릴 수 있음
- COMPLETE 정의: 지침 보강 + lesson1 주석 + 테스트 + commit
- PARTIAL 정의: 문서 또는 코드 한쪽만 닫힘
- BLOCKED 정의: 설명 규칙과 실제 코드 적용을 조화롭게 닫지 못함

## 4. Topic Analysis

- 현재 이해한 문제 구조:
  - lesson1은 비교형 강의인데 코드 주변 설명이 아직 얕음
  - AGENTS와 템플릿은 설명 품질이 강하지만 comparison-heavy topic 전용 규칙을 더 명시할 여지가 있음
- 이번 작업이 건드리는 표면:
  - `PROJECT_INTENT.md`
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
  - `lesson1` Java sources
- 숨은 가정:
  - 비교형 학습은 같은 질문 위에 다른 구현을 올려 두는 방식이 효과적임
- 핵심 미지수:
  - 문서를 얼마나 더 강화해야 충분한가
  - lesson1에서 어떤 지점이 주석 우선순위가 높은가
- 처음 떠오른 접근:
  - 문서에 비교 학습 규칙 명시
  - 각 서버에 class-level KDoc + 핵심 block comments 추가
- 성공을 오판하기 쉬운 지점:
  - "비교가 중요하다"는 선언만 있고 실제 코드 주석은 일반론에 머무는 것

## 5. Analysis Critique + Repair

### 5.1 Challenge 1

- 무엇이 틀릴 수 있는가: 기존 문서도 이미 학습과 비교를 말하므로 과한 반복이 될 수 있다
- repair: 새 문구는 선언이 아니라 `같은 질문 / 차이 / 이유 / 관측 포인트`라는 실행 규칙으로 추가
- 왜 repair가 더 강한가: 실제 작업 판단에 바로 쓸 수 있다

### 5.2 Challenge 2

- 무엇이 좁거나 누락될 수 있는가: 코드 주석이 단순 구현 설명으로 끝날 수 있다
- repair: 각 서버가 "다른 비교 대상과 무엇이 다른가"를 직접 설명하게 함
- 왜 repair가 더 강한가: 비교형 학습 목적과 바로 연결된다

### 5.3 Challenge 3

- 무엇이 학습/설명 목표를 놓칠 수 있는가: blocking/NIO/Netty를 따로따로만 설명하고 공통 질문이 안 보일 수 있다
- repair: 같은 echo 문제 위에서 어떤 메커니즘 차이가 생기는지 주석과 지침 양쪽에 명시
- 왜 repair가 더 강한가: 독자가 비교축을 잃지 않는다

### 5.4 Retained Framing

- 최종 채택 분석: 문서 규범 + lesson1 code-adjacent explanation exemplar
- 폐기한 분석과 이유: lesson1 문서만 고치는 방식은 code-adjacent explanation 요구를 충분히 닫지 못함

## 6. Research / Evidence Plan

- 먼저 확인해야 하는 질문:
  - 현재 문서에 비교 학습 규칙이 어느 정도 들어 있는가
  - lesson1 코드에서 어떤 지점이 load-bearing 비교 포인트인가
- 초기 검색 키워드: `학습`, `비교`, `blocking`, `non-blocking`, `주석`
- 확장 키워드: `comparison`, `same question`, `observability`
- 조사할 repo 경로:
  - `PROJECT_INTENT.md`
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
  - `src/main/java/io/aimpugn/learn/netty/lesson1`
- 조사할 1차 자료:
  - 현재 on-disk 문서와 코드
- 기대하는 증거:
  - 비교 학습 규칙의 빈틈
  - lesson1 주석 보강 포인트
- research stop condition: 문서 패치와 code comment target이 확정됨

## 7. Evidence Ledger

- E-01
  - 질문 / 주장: `PROJECT_INTENT.md`는 이미 학습용 실험실과 비교 예제를 말하지만, comparison-heavy topic의 닫힘 조건을 더 분명히 적을 수 있다
  - 근거 유형: `source code`
  - 자료: `PROJECT_INTENT.md`
  - 이 자료가 닫아 준 것: 의도 문서 보강 필요성
  - 아직 비어 있는 것: 없음
- E-02
  - 질문 / 주장: `AGENTS.md`는 설명 품질은 강하지만 비교형 topic 전용 강제 규칙이 더 명시될 수 있다
  - 근거 유형: `source code`
  - 자료: `AGENTS.md`
  - 이 자료가 닫아 준 것: comparison rule 추가 위치
  - 아직 비어 있는 것: 없음
- E-03
  - 질문 / 주장: lesson1 code는 KDoc이 적고, 각 구현이 왜 다른 모양인지 설명이 충분하지 않다
  - 근거 유형: `source code`
  - 자료: `lesson1` Java files
  - 이 자료가 닫아 준 것: class/block comments 보강 필요성
  - 아직 비어 있는 것: 없음

## 8. Evidence Critique + Repair

- 약한 근거: 없음
- 충돌하는 근거: 없음
- 오래되었을 가능성이 있는 가정: 없음
- 추가 수집 필요 항목: 없음
- confidence를 낮춰야 하는 항목: 없음
- repair 후 최종 근거 세트: 현재 지침 문서 + lesson1 source

## 9. PROJECT_INTENT Fit & Learning Design

- 이번 작업이 직접 지원하는 `PROJECT_INTENT` 항목:
  - 학습이 우선인 프로젝트 정체성
  - 비교 예제 중심 학습 방식
  - 질문/실험/관측으로 닫는 산출물
- 이번 작업에서 반드시 드러내야 하는 **보이지 않는 동작**:
  - blocking/NIO/Netty가 같은 echo 문제를 서로 다른 모델로 푼다는 점
- 독자가 나중에 스스로 설명할 수 있어야 하는 질문:
  - 왜 blocking은 client마다 worker가 필요한가
  - 왜 selector는 line 경계를 직접 조립해야 하는가
  - 왜 Netty는 pipeline과 boss/worker로 보이는가
- 이번 작업의 핵심 학습 대상: comparison-heavy explanation
- 이번 작업의 핵심 혼동쌍:
  - 구현 소개 vs 비교 학습
  - code comment vs code narration

### 9.1 Observability Design

- ASCII 다이어그램 필요 여부 / 이유: `N`, 이번 작업은 지침과 코드 주석 보강이 핵심
- 로그 또는 타임라인 필요 여부 / 이유: `Y`, 기존 ObservationLog를 관측 포인트로 활용
- 시각화 페이지 필요 여부 / 이유: `N`
- 실패 실험 필요 여부 / 이유: `N`
- 비교 예제 필요 여부 / 이유: `Y`, lesson1 trio 자체가 비교 예제
- 부분 재구현 필요 여부 / 이유: `N`
- 채택한 관측 장치: thread 이름과 phase 로그
- 채택하지 않은 관측 장치와 이유: UI는 이번 작업 범위 밖

### 9.2 Explanation Design

- 눈에 바로 들어오게 보여 줄 구조:
  - 학습 목적 강조
  - 비교형 topic 전용 규칙
  - 각 서버가 다른 점
- 읽으면서 논리가 쌓이게 할 핵심 서술:
  - 같은 echo 문제를 서로 다른 동시성/추상화 모델로 푼다
- worked example 후보:
  - blocking vs NIO vs Netty
- 회상 anchor:
  - `같은 문제, 다른 모델, 다른 관측 포인트`
- 비교형 학습 작업인지 여부: `Y`
- 비교 대상:
  - `BlockingEchoServer`
  - `NioSelectorEchoServer`
  - `NettyEchoServer`
- 같은 질문 또는 동일 조건:
  - 같은 echo 요청을 서버가 어떻게 받고 다시 돌려주는가
- 차이가 생기는 핵심 지점:
  - thread 배치
  - line 경계 처리 위치
  - write 처리 방식
  - 역할 분리 수준
- 왜 그 차이가 생기는가:
  - blocking은 단순성을 위해 client별 대기를 택함
  - NIO는 selector 기반으로 readiness를 묶어 처리함
  - Netty는 그 차이를 EventLoop와 pipeline 추상화로 정리함
- 차이를 직접 볼 수 있는 관측 포인트:
  - 로그의 thread 이름
  - accept/read/write phase
  - decoder/handler 경계 주석
- 코드 스니펫 / 파일 링크로 직접 보여 줄 위치:
  - `lesson1` server source files
- 코드 주변에 반드시 설명을 남길 파일 / 지점:
  - 각 server class KDoc
  - blocking accept/handle loop
  - NIO selector/read/write blocks
  - Netty boss/worker/pipeline/handler blocks
- 각 지점에 남길 설명 형태: `KDoc | 짧은 근접 주석`
- 각 주석이 답해야 하는 질문:
  - 왜 이 구현은 옆 구현과 다른 모양인가
  - 어디서 차이를 볼 수 있는가
- 이 reasoning을 코드 밖 문서에만 두면 왜 부족한가:
  - 코드 파일만 읽는 순간 비교축을 잃기 쉽기 때문

### 9.3 Learning Transfer Design

- 이번 작업이 다음 단계에 넘겨야 하는 핵심 개념: 비교형 학습은 같은 질문과 차이의 이유를 함께 닫아야 한다
- 이번 작업이 닫아 줘야 하는 prerequisite: lesson1을 exemplar로 삼을 수 있게 함
- 다음 단계가 이 작업을 발판으로 바로 사용할 자산:
  - 강화된 지침
  - lesson1 code comments
- 이번 작업 완료 후 독자가 스스로 답할 수 있어야 하는 질문:
  - `blocking/NIO/Netty는 왜 각각 이런 구조를 가지는가`
- 다음 단계 진입 질문:
  - `lesson2 EventLoop 설명에서는 어떤 비교축이 핵심인가`

## 10. Scope Expansion & Impact Sync

### 10.1 Included Scope

- 코드:
  - `lesson1` source files
- 테스트:
  - `mvn test`
- 문서:
  - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 이 WORK 문서
- 예제:
  - lesson1 echo servers
- 실험:
  - 없음
- 로그 / 관측 장치:
  - `ObservationLog`
- 프론트 / 시각화:
  - 없음
- 설정 / 스크립트:
  - 없음

### 10.2 Excluded Scope

- 제외 항목:
  - lesson2 신규 예제
  - lesson1 문서 대폭 개정
  - front viewer
- 제외 근거: 이번 요청은 지침 강화와 code-adjacent explanation 적용
- 왜 지금 제외해도 개념적 완결성이 깨지지 않는가: 첫 exemplar와 규칙을 먼저 닫는 편이 안전

### 10.3 Search Expansion Ledger

- 핵심 키워드: `학습`, `비교`, `blocking`, `non-blocking`, `주석`
- 동의어 / 약어 / 구명칭: `comparison`, `selector`, `EventLoop`, `comment`, `KDoc`
- 관련 에러 / 로그 / API 이름: `readLine`, `selector`, `writeAndFlush`
- 실제 검색 경로:
  - `PROJECT_INTENT.md`
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
  - `src/main/java/io/aimpugn/learn/netty/lesson1`
- 누락 가능성 점검 결과: 이번 작업 범위에는 충분

### 10.4 Impact Sync Map

- 함께 움직여야 하는 자산:
  - 의도 문서
  - 로컬 규범
  - WORK 템플릿
  - lesson1 code comments
- 한쪽만 바꾸면 깨질 부분:
  - 지침만 바뀌고 exemplar 코드가 그대로면 설득력이 약함
  - 코드만 바뀌고 지침이 그대로면 다음 작업에서 반복 누락 가능
- 관련 문서 동기화 항목:
  - 향후 lesson 문서/WORK
- 관련 예제 동기화 항목:
  - lesson1 trio
- 관련 관측 장치 동기화 항목:
  - `ObservationLog`

## 11. Design

- 선택한 접근:
  - `PROJECT_INTENT.md`에 학습과 비교를 더 명시
  - `AGENTS.md`에 comparison-heavy topic 규칙 추가
  - `AGENTS_WORK_TEMPLATE.md`에 comparison closure gate 추가
  - lesson1 classes와 핵심 blocks에 비교형 주석 추가
- 왜 이것이 근본 문제와 이점에 맞는가:
  - 규범과 exemplar를 함께 바꾸어 다음 작업에 재사용 가능
- 고려한 대안 A:
  - code comments만 추가
- 대안 A를 채택하지 않은 이유:
  - 지침이 흐리면 다음 작업에서 다시 빠질 수 있음
- 고려한 대안 B:
  - lesson1 문서만 보강
- 대안 B를 채택하지 않은 이유:
  - 사용자가 code-adjacent explanation을 원했고, 코드 비교 포인트가 직접 보여야 함
- 주요 계약 / 경계:
  - 학습과 비교를 지침에 강조
  - 주석은 비교 핵심에 집중
  - 기존 동작은 바꾸지 않음
- 실패 모드:
  - 비교를 말하지만 실제 코드 주석은 일반론으로 끝남
  - 테스트 누락
- verification path:
  - 지침 문서에서 comparison rule 존재 확인
  - lesson1 source에 class/block comments 존재 확인
  - `mvn test`

## 12. Design Critique + Repair

### 12.1 Correctness View

- 반론: 주석 추가는 동작을 안 바꾸니 테스트가 약할 수 있다
- repair 또는 유지: 유지. 그래도 build/test는 regression guard로 필요
- 이유: 주석 추가 중 문법 실수나 import drift를 막는다

### 12.2 Learner View

- 반론: 비교 설명이 길어져 코드 읽기가 무거워질 수 있다
- repair 또는 유지: 유지하되 class-level KDoc와 핵심 block comment만 추가
- 이유: 오해 비용이 큰 지점에만 밀도 있게 남기는 편이 더 좋다

### 12.3 Operations / Performance View

- 반론: lesson1 문서도 함께 손봐야 완전할 수 있다
- repair 또는 유지: 유지. 이번 작업은 exemplar code comments까지 우선 닫음
- 이유: 사용자의 직접 지적 대상이 코드 주석 부재였기 때문

### 12.4 Final Design Decision

- 최종 채택안: docs rule upgrade + lesson1 code comments
- 트레이드오프: 문서와 코드를 함께 바꾸므로 작업량은 늘지만, 학습 일관성이 훨씬 좋아진다

## 13. Overall Plan

- 작업 순서:
  1. 관련 문서와 lesson1 code 읽기
  2. comparison-heavy rule 설계
  3. 문서 패치
  4. lesson1 comments 패치
  5. WORK 문서 작성
  6. test + audit + commit
- 선행 의존성: current guidance와 lesson1 구조 파악
- validation order: 문서/코드 자기검토 -> `git diff --check` -> `mvn test` -> commit
- retry order: rule wording -> comment target -> test
- rollback / reopen 기준: comparison rule이 실제 comment 전략과 어긋나면 design 단계 reopen

### 13.1 Sequence Position And Next-Step Hypothesis

- 현재 학습 경로에서의 위치: lesson1 품질 보강
- 원래 예상한 다음 단계: lesson2 EventLoop 또는 lesson1 code comment follow-up
- 이번 작업 중 순서가 바뀔 수 있는 신호: lesson1 문서 쪽 설명 공백이 크게 보이는 경우
- 순서가 바뀌면 다시 확인할 prerequisite: lesson1 exemplar가 충분히 닫혔는가

## 14. Plan Critique + Repair

- 이 계획이 실패할 수 있는 지점: 문서만 좋아지고 code comments가 약할 수 있다
- 순서상 위험: code comments 먼저 붙이면 비교 규칙이 흔들릴 수 있다
- 빠진 prerequisite: 없음
- repair: guidance first, code second
- 왜 repair된 계획이 더 강한가: 사용자가 요청한 순서를 따르고 기준 흔들림을 줄인다

## 15. Detailed Task Plan

- Unit-01
  - 목적: 학습/비교 중심 지침 보강
  - 대상 파일 / 자산:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 바꿀 논리: comparison-heavy topic closure rules
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 코드 주변 설명 계획: `N/A`, 문서 작업
  - 검증: 관련 규칙 존재 여부 확인
  - 완료 판정: comparison rule과 learning purpose 강화
- Unit-02
  - 목적: lesson1 code-adjacent explanation 적용
  - 대상 파일 / 자산:
    - `lesson1` Java files
  - 바꿀 논리: 각 구현이 다른 구현과 무엇이 다른지와 관측 포인트 설명
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명, 기존 로그 관측
  - 코드 주변 설명 계획:
    - class-level KDoc
    - 핵심 loop/block 설명
  - 검증: source review
  - 완료 판정: comparison comments visible
- Unit-03
  - 목적: WORK, test, audit, commit
  - 대상 파일 / 자산:
    - 이 WORK 문서
    - 변경 전체
  - 바꿀 논리: closure 기록
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 코드 주변 설명 계획:
    - comment target과 actual added comments 기록
  - 검증: `git diff --check`, `mvn test`, commit
  - 완료 판정: verification + commit

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: 지침이 학습 목적과 comparison-heavy topic closure를 더 분명히 요구
  - S2: lesson1 code comments가 blocking/NIO/Netty 차이와 이유를 설명
  - S3: 테스트가 그대로 통과
- 실패 케이스 최소 3개:
  - F1: 문서가 여전히 일반론만 말하고 비교형 주제 규칙이 흐림
  - F2: code comments가 "무슨 코드인가"만 말하고 왜 다른지 안 보임
  - F3: 주석 추가 후 build/test가 깨짐
- 회귀 위험:
  - lesson1 설명 밀도가 아직 문서와 코드 사이에 완전히 균형 잡히지 않을 수 있음
- 회귀 방지 확인 경로:
  - 이후 lesson work에서 새 comparison gate 사용

## 16. Detailed Plan Critique + Repair

- 빠진 단위: 없음
- fuzzy success criteria: "학습 목적 강조"가 추상적일 수 있음
- 과한 범위 / 부족한 범위: lesson1 문서 대규모 개정은 과함
- repair: comparison-heavy closure와 actual comment targets로 success를 구체화
- 최종 상세 계획: Unit-01 -> Unit-02 -> Unit-03

## 17. Frozen Success / Failure Checklist

> 사용자 항목은 최소 바닥선이다. 삭제·완화 금지.
> 모든 required 항목이 PASS여야만 COMPLETE 가능.

- C-01
  - 출처: `사용자`
  - 내용: 학습 목적임을 지침에 더 강조
  - required: `Y`
  - PASS 기준: `PROJECT_INTENT.md`, `AGENTS.md`에서 학습과 비교 중심 성격이 더 또렷하게 드러남
  - FAIL 기준: 기존 수준과 크게 다르지 않음
  - 필요한 증거: 문서 본문
  - 재시도 트리거: 학습 강조 부족
  - 관련 Unit: `Unit-01`
- C-02
  - 출처: `사용자`
  - 내용: blocking/non-blocking 같은 비교에서 차이와 이유 설명
  - required: `Y`
  - PASS 기준: comparison-heavy rule과 lesson1 comments가 same question / difference / why / observability를 닫음
  - FAIL 기준: 비교 대상 소개만 있고 차이 이유가 없음
  - 필요한 증거: 지침 문서, code comments
  - 재시도 트리거: 차이 이유 부재
  - 관련 Unit: `Unit-01`, `Unit-02`
- C-03
  - 출처: `사용자`
  - 내용: 개선된 지침 기반으로 주석 설명 적용
  - required: `Y`
  - PASS 기준: lesson1 source에 새 비교형 KDoc/comment 존재
  - FAIL 기준: 문서만 바뀌고 code comment 적용 없음
  - 필요한 증거: lesson1 source files
  - 재시도 트리거: code-adjacent explanation 누락
  - 관련 Unit: `Unit-02`
- C-04
  - 출처: `AI-추가`
  - 내용: comparison-heavy topic gate를 WORK 템플릿에 반영
  - required: `Y`
  - PASS 기준: `COMPARISON_LEARNING_READY`와 관련 Explanation Design 항목 존재
  - FAIL 기준: 템플릿 반영 없음
  - 필요한 증거: `AGENTS_WORK_TEMPLATE.md`
  - 재시도 트리거: 다음 작업에서 누락 가능
  - 관련 Unit: `Unit-01`
- C-05
  - 출처: `AI-추가`
  - 내용: 코드 주변 설명은 load-bearing 지점에 집중해 developer 지침과 조화
  - required: `Y`
  - PASS 기준: class/block-level comments 중심, line-by-line narration 없음
  - FAIL 기준: 과잉 주석
  - 필요한 증거: diff review
  - 재시도 트리거: 주석 과다
  - 관련 Unit: `Unit-02`
- C-06
  - 출처: `AI-추가`
  - 내용: test + commit closure
  - required: `Y`
  - PASS 기준: `mvn test` 통과 후 commit
  - FAIL 기준: test 또는 commit 누락
  - 필요한 증거: command result, commit SHA
  - 재시도 트리거: test failure or commit failure
  - 관련 Unit: `Unit-03`

### 17.1 Checklist Critique + Repair

- 각 항목이 목표 / 이점 / 사용자 요구와 매핑되는가: `Y`
- PASS / FAIL이 관측 가능한가: `Y`
- 증거가 실제로 존재 가능한가: `Y`
- 사용자 요구를 조용히 약화한 항목이 없는가: `Y`
- repair: `C-04`, `C-05`, `C-06`으로 지속성/과잉 주석/closure를 함께 묶음
- Freeze 시각: `2026-04-07 00:14 KST`
- Freeze 버전: `v1`

## 18. Execution Ledger

- Attempt-01 / Unit-01
  - 시작 시각: `2026-04-07 00:11 KST`
  - 실행 내용: 의도 문서와 지침/템플릿에 comparison-heavy topic rules 추가
  - 변경 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 추가/보강한 코드 주변 설명: `N/A`, 문서 작업
  - 새로 생긴 증거: comparison-heavy closure rules
  - 이번 시도에서 새로 얻은 지식: 의도 문서와 로컬 규범을 같이 바꿔야 학습 정체성과 실행 규칙이 함께 닫힌다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: 본문 자기검토
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: code comments apply
- Attempt-02 / Unit-02
  - 시작 시각: `2026-04-07 00:13 KST`
  - 실행 내용: lesson1 sources에 comparison-focused KDoc/comments 추가
  - 변경 파일:
    - `BlockingEchoServer.java`
    - `NioSelectorEchoServer.java`
    - `NettyEchoServer.java`
    - `ObservationLog.java`
    - `Lesson1Support.java`
    - `LessonServer.java`
  - 추가/보강한 코드 주변 설명:
    - 각 서버 class-level KDoc
    - blocking accept/handle loop 설명
    - NIO selector/read/write 설명
    - Netty boss/worker/pipeline/handler 설명
  - 새로 생긴 증거: source comments
  - 이번 시도에서 새로 얻은 지식: `ObservationLog`와 `LessonServer`에도 비교 학습의 공통 축을 남겨 두면 첫 강의 전체가 더 잘 묶인다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: source review
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: verification and commit
- Attempt-03 / Unit-03
  - 시작 시각: `2026-04-07 00:15 KST`
  - 실행 내용: WORK 문서 작성, test, audit, commit
  - 변경 파일:
    - 이 WORK 문서
    - 위 문서/코드 전부
  - 추가/보강한 코드 주변 설명:
    - 실제 comment target과 added comments를 WORK에 기록
  - 새로 생긴 증거: verification results, commit
  - 이번 시도에서 새로 얻은 지식: comparison-heavy lesson은 code comments와 logs를 함께 봐야 이해가 훨씬 빨라진다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: `git diff --check`, `mvn test`, commit
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: verification results 반영

## 19. Retry / Re-entry Ledger

- R-01
  - trigger: `N/A`
  - 왜 COMPLETE가 아니게 되었는가: `N/A`
  - earliest affected phase: `N/A`
  - root cause repair: `N/A`
  - 다시 수행한 phase: `N/A`
  - 재검증 결과: `N/A`
- R-02
  - trigger: `N/A`
  - 왜 COMPLETE가 아니게 되었는가: `N/A`
  - earliest affected phase: `N/A`
  - root cause repair: `N/A`
  - 다시 수행한 phase: `N/A`
  - 재검증 결과: `N/A`

### 19.1 Retry Budget & Escalation

- 총 시도 횟수 상한: `5회`
  - 첫 시도를 포함합니다.
  - 이번 작업은 3시도 안에 닫힘
- 같은 근본 원인으로 2번 이상 실패하면:
  - earliest affected phase reopen
- 5회 시도 후에도 required 항목이 FAIL 또는 미판정이면:
  - `PARTIAL` 또는 `BLOCKED`

### 19.2 Roadmap Reorder Log

- Change-01
  - reorder 발생 여부: `N`
  - 변경 전 순서: `지침 강화 -> lesson1 comments -> lesson2`
  - 변경 후 순서: `동일`
  - 변경 트리거: `N/A`
  - 왜 이 변경이 더 강한가: `N/A`
  - 뒤로 밀린 항목: 없음
  - 다음 WORK에 넘길 메모: lesson1 문서 또는 lesson2에서 같은 comparison rule 사용
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
  - evidence: `PROJECT_INTENT.md`, `AGENTS.md`의 학습/비교 강조 문구
  - notes: 학습 목적과 비교 중심성이 더 분명해짐
- C-02 Final: `PASS`
  - evidence: 비교 학습 규칙 문서 + lesson1 comments
  - notes: same question / difference / why / observability가 함께 보임
- C-03 Final: `PASS`
  - evidence: lesson1 source KDoc/comment diff
  - notes: 개선된 지침 기반으로 code-adjacent explanation 적용
- C-04 Final: `PASS`
  - evidence: `AGENTS_WORK_TEMPLATE.md`의 `COMPARISON_LEARNING_READY`와 Explanation Design 항목
  - notes: comparison-heavy WORK 강제 장치 추가
- C-05 Final: `PASS`
  - evidence: source diff review
  - notes: class/block-level comments 중심, 과잉 주석 없음
- C-06 Final: `PASS`
  - evidence: `mvn test` 성공, commit
  - notes: tests run 12, failures 0

## 21. Compliance Audit

### 21.1 PROJECT_INTENT Compliance

- 이번 작업이 `PROJECT_INTENT`의 어떤 목표를 충족했는가:
  - 학습 목적 강화
  - 비교 예제의 학습 규칙 강화
- `PROJECT_INTENT`와 어긋난 부분:
  - 없음
- 어긋났다면 repair:
  - `N/A`
- Final: `PASS`

### 21.2 Global AGENTS Compliance

- 전역 hard rule 중 이번 작업에 적용된 것:
  - instruction stack
  - checklist freeze
  - test+commit closure
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

### 21.3 Local AGENTS Compliance

- 로컬 hard rule 중 이번 작업에 적용된 것:
  - 학습 친화 설명
  - comparison-heavy rule
  - code-adjacent explanation
- 관측 장치 / 설명 품질 / 자연스러운 한국어 / 코드 링크 규칙 / 코드 주변 설명 규칙 준수 여부:
  - 준수
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

## 22. Final Audit

### 22.1 Expert Review

- 목적 적합성: 사용자의 "학습 우선, 비교 설명 강화, 그 뒤 code comments 적용" 요청을 정확히 따름
- 개념적 완결성: 의도 문서, 로컬 규범, WORK 템플릿, exemplar 코드가 함께 움직임
- 작업 완결성: 문서 보강, code comments, WORK, test, commit까지 닫힘
- 결과 완결성: lesson1이 comparison-heavy lesson exemplar로 더 선명해짐
- 설명 품질: class-level KDoc와 핵심 block comment가 차이와 이유를 바로 설명
- 코드 주변 설명 존재/품질: blocking/NIO/Netty와 공통 support 코드에 load-bearing 설명 존재
- 비교 학습 closure: 같은 echo 문제, 다른 모델, 차이 이유, 관측 포인트가 문서와 코드 모두에 존재
- 검증 정직성: `git diff --check`와 `mvn test`만 실제 실행했다고 명시
- 남은 이상 징후: lesson1 문서 본문은 아직 이번 code comment 강화까지 반영되지는 않음

### 22.2 Final Audit Repair

- audit에서 발견된 문제:
  - 없음
- repair:
  - `N/A`
- repair 후 재검토 결과:
  - closure 가능

## 23. Final Verification Summary

- 실제 실행한 명령 / 테스트 / 확인:
  - `git diff --check -- PROJECT_INTENT.md AGENTS.md AGENTS_WORK_TEMPLATE.md src/main/java/io/aimpugn/learn/netty/lesson1 docs/works/20260407_learning_first_guidance_and_lesson1_comments.md`
  - `mvn test`
  - `git status --short -- PROJECT_INTENT.md AGENTS.md AGENTS_WORK_TEMPLATE.md src/main/java/io/aimpugn/learn/netty/lesson1 docs/works/20260407_learning_first_guidance_and_lesson1_comments.md`
- PASS 신호:
  - comparison-heavy rule이 지침과 템플릿에 반영
  - lesson1 source에 새 KDoc/comment 반영
  - `mvn test` 12개 테스트 통과
- FAIL 신호:
  - 지침은 바뀌었는데 code comments가 일반론에 머무름
  - code comments 추가 후 컴파일/테스트 실패
- 코드 주변 설명 검증 결과:
  - `BlockingEchoServer`, `NioSelectorEchoServer`, `NettyEchoServer`에 비교형 KDoc/comment 추가
  - `ObservationLog`, `Lesson1Support`, `LessonServer`에도 첫 강의 공통 학습 축 설명 추가
- 비교 학습 검증 결과:
  - 문서와 코드에서 모두 `같은 문제 / 다른 메커니즘 / 이유 / 관측 포인트` 확인
- 참고 메모:
  - Netty 테스트 중 `sun.misc.Unsafe` 경고가 보이지만, 현재 JDK와 Netty 내부 구현 조합에서 나오는 런타임 경고이고 테스트 실패는 아님
- 실행하지 못한 검증:
  - 없음
- 실행하지 못한 이유:
  - `N/A`
- 최종 confidence:
  - `high`

## 24. Final Deliverable Inventory

- D-01
  - 자산: 학습/비교 중심 규범 보강
  - 유형: `문서`
  - 역할: comparison-heavy topic의 지침 고정
  - 어떤 질문에 답하는가: `이 프로젝트는 무엇을 어떻게 비교하며 배우는가`
  - 주요 경로 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
- D-02
  - 자산: lesson1 comparison-focused code comments
  - 유형: `코드`
  - 역할: blocking/NIO/Netty 차이와 이유를 코드 주변에서 바로 보이게 함
  - 어떤 질문에 답하는가: `같은 echo 문제를 세 모델이 왜 다르게 푸는가`
  - 주요 경로 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/main/java/io/aimpugn/learn/netty/lesson1/blocking/BlockingEchoServer.java`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/main/java/io/aimpugn/learn/netty/lesson1/nio/NioSelectorEchoServer.java`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/src/main/java/io/aimpugn/learn/netty/lesson1/netty/NettyEchoServer.java`
- D-03
  - 자산: 이번 작업 기록
  - 유형: `문서`
  - 역할: 결정, 검증, closure 기록
  - 어떤 질문에 답하는가: `왜 이 지침과 주석을 이렇게 바꿨는가`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_learning_first_guidance_and_lesson1_comments.md`

## 25. Learning Continuity Closure

- 이번 작업으로 닫힌 학습 질문:
  - 첫 강의 비교형 설명에서 무엇을 꼭 닫아야 하는가
- 이번 작업이 열어 버린 새 질문:
  - lesson1 문서 본문도 이번 code comment 강화에 맞춰 더 다듬을 것인가
  - lesson2 EventLoop 설명에서는 어떤 비교축을 쓸 것인가
- 다음 작업 후보:
  - lesson1 문서 보강
  - lesson2 EventLoop / boss / worker
- 권장 다음 작업:
  - lesson2 EventLoop / boss / worker
- 왜 지금 이 작업이 다음 단계로 가장 자연스러운가:
  - lesson1의 비교축이 정리됐으니 이제 그 다음 개념을 같은 방식으로 확장하기 좋음
- 순서 변경 여부: `N`
- 변경 전 로드맵:
  - 지침 강화
  - lesson1 comments
  - lesson2
- 변경 후 로드맵:
  - 동일
- 뒤로 미룬 작업과 이유:
  - lesson1 문서 대폭 개정은 이번 요청 범위를 넘음
- 다음 WORK가 이어받아야 할 자산 / evidence / commit:
  - comparison-heavy guidance
  - lesson1 code comments
  - 이번 commit
- 다음 작업 시작 조건:
  - lesson2 핵심 비교 질문 확정
- 더 이상 다음 작업이 필요 없다고 판단한다면 그 근거:
  - 해당 없음
- Learning chain final: `PASS`

## 26. Completion Decision

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 열린 게이트: 없음
- 미해소 blocker: 없음
- COMPLETE라고 판단하는 근거: docs rule upgrade, code comments, tests, commit 모두 닫힘
- PARTIAL / BLOCKED라면 승격 조건: `N/A`

## 27. Commit Closure

- repo 변경 여부: `Y`
- commit 요구 여부: `Y`
- commit 수행 여부: `Y`
- commit SHA: `same-commit self-reference limitation: exact SHA는 final 응답과 git history에서 확인`
- commit message: `docs: strengthen learning-first lesson guidance`
- 이 작업을 닫는 커밋인지 여부: `Y`
- push 여부: `N`
- push는 왜 했거나 하지 않았는가: 사용자 요청 없음

## 28. Closing Notes

- 이번 작업의 핵심 교훈: 비교가 핵심인 lesson은 문서 선언만으로는 부족하고, 코드 옆에서도 같은 질문과 차이의 이유가 보여야 학습 효과가 커진다
- 다음 작업으로 자연스럽게 이어지는 질문: EventLoop를 본격적으로 다룰 때는 boss/worker, single-thread affinity, offload를 어떤 비교축으로 보여 줄 것인가
- 후속 작업 필요 여부: `Y`
- 최종 종료 시각: `2026-04-07 00:23 KST`
