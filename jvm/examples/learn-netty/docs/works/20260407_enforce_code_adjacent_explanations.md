# WORK_20260407_ENFORCE_CODE_ADJACENT_EXPLANATIONS

> 코드 변경 작업에서 설명 프로토콜이 문서나 채팅에만 남지 않고, 실제 코드 주변 설명으로도 남도록 로컬 규범과 WORK 템플릿을 강화한 작업 기록입니다.

## 0. Meta

- 작업 제목: 코드 주변 설명 강제 규칙 추가
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_enforce_code_adjacent_explanations.md`
- 파일명 규칙 확인: `docs/works/YYYYMMDD_${작업요약}.md`
- 저장소 / 모듈: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty`
- 작업 유형: `design | explain | execute | mixed`
- 작업 깊이: `full`
- 실행자: `Codex`
- 시작 일시: `2026-04-07 00:03 KST`
- 종료 일시: `2026-04-07 00:09 KST`
- 관련 요청 / 이슈: 사용자의 코드 주변 설명 강제 요청
- 원문 사용자 요청: 설명 프로토콜과 설명 방식을 private이어도 코드 주변에 반드시 작성하도록 `AGENTS.md`, `AGENTS_WORK_TEMPLATE.md`를 통해 강제
- run_mode: `normal`
- finish: `verify+commit`
- repo 변경 여부: `Y`
- commit 기본 요구 여부: `Y`
- 직전 WORK / 선행 작업: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_project_structure_guideline.md`
- 후속 WORK / 다음 작업 후보: `lesson1 코드 주변 설명 보강`
- 이번 작업의 학습 단계 위치: `프로젝트 실행 규범 강화`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`

### 0.1 Usage Policy

- 이번 작업은 코드 자체를 바꾸지 않지만, 앞으로의 모든 코드 작업 품질을 좌우하는 규범 변경이라 `full`로 다뤘습니다.
- 결과는 문서 3개로 닫습니다.
  1. `AGENTS.md`
  2. `AGENTS_WORK_TEMPLATE.md`
  3. 이 WORK 문서
- 이번 작업의 핵심은 "설명을 잘 쓰자"가 아니라 "코드 주변 설명이 없으면 COMPLETE가 안 되게 만들자"를 규칙으로 고정하는 것입니다.

### 0.2 Learning Chain Policy

- 이번 작업은 개별 lesson이 아니라 lesson 전반의 작성 방식을 바꾸는 상위 규범 작업입니다.
- 따라서 다음 학습 단계는 자연스럽게 `이미 만든 lesson1 코드에 새 규칙을 실제로 적용`하는 쪽으로 이어져야 합니다.
- 단순 선언으로 끝나면 규범은 생겨도 실제 학습 자산은 그대로일 수 있으므로, 다음 단계 연결을 기록합니다.

## 1. Instruction Stack Preflight

### 1.1 Resolved Sources

- 전역 `AGENTS.md` 경로: `/Users/rody/.codex/AGENTS.md`
- 전역 `AGENTS.md` 읽음: `Y`
- 전역 `AGENTS.md` 적용 요약: fail-closed completion, checklist freeze, commit 포함 closure
- `PROJECT_INTENT.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
- `PROJECT_INTENT.md` 읽음: `Y`
- `PROJECT_INTENT.md` 적용 요약: 사람이 다시 설명하고 검증 가능한 자산을 남겨야 함
- 로컬 `AGENTS.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 로컬 `AGENTS.md` 읽음: `Y`
- 로컬 `AGENTS.md` 적용 요약: 보이지 않는 동작을 보이게 만들고, 쉬운 한국어와 검증 가능한 설명을 남겨야 함
- 이 템플릿 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
- 이 템플릿 사용 시작 시각: `2026-04-07 00:03 KST`

### 1.2 Project Overlay

- 로컬 규범 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 이번 작업에서 활성화한 프로젝트 규칙:
  - 설명은 문서뿐 아니라 예제와 코드에도 남아야 함
  - 쉬운 한국어 우선
  - 학습형 자산은 나중에 혼자 다시 설명할 수 있어야 함
- 활성화 이유: 사용자가 기존 lesson 코드에 주석이 없는 점을 문제로 지적했기 때문
- 이 규칙을 실제로 검증할 방법: `AGENTS.md`와 `AGENTS_WORK_TEMPLATE.md`에 코드 주변 설명 강제 규칙과 완료 게이트가 들어갔는지 확인
- 전역 규칙과의 충돌 가능성: 개발자 지침의 "주석은 드물게" 원칙과 긴장 가능성 있음
- 충돌 해소 방식: 모든 줄 주석이 아니라 load-bearing 코드와 학습 핵심 지점에만 짧고 밀도 있는 설명을 요구하도록 설계

### 1.3 Effective Contract

- 이번 작업에서 특히 강한 hard rule:
  - code-adjacent explanation을 명시적으로 강제할 것
  - WORK 템플릿에서 그 규칙을 계획/검수/완료 게이트까지 연결할 것
  - trivial narration 주석은 늘리지 않을 것
- 충돌 가능성:
  - 규칙이 과해서 모든 줄 주석처럼 오해될 수 있음
- 충돌 해소 방식:
  - "load-bearing, 비자명한 분기, 관측 포인트"에 집중한다는 범위를 분명히 적음
- preflight blocker: 없음

### 1.4 Completion Formula

- `INSTRUCTION_STACK_READY = PASS`
- `LEARNING_CHAIN_READY = PASS`
- `CODE_EXPLANATION_READY = PASS`
  - 이유: 이번 작업은 코드 변경이 없으므로 `N/A with reason`을 허용하는 규칙까지 템플릿에 추가했고, 그 설계가 문서에 반영됨
- `ALLOW_COMPLETE = AGENTS.md + AGENTS_WORK_TEMPLATE.md + WORK record + verification + commit`
- 하나라도 FAIL 또는 미판정이면 `BLOCK_COMPLETE`

### 1.5 Phase Status Board

- Topic Analysis: `CLOSED`
  - last updated: `2026-04-07 00:04 KST`
  - stale reason: `N/A`
- Research / Evidence: `CLOSED`
  - last updated: `2026-04-07 00:04 KST`
  - stale reason: `N/A`
- Design: `CLOSED`
  - last updated: `2026-04-07 00:05 KST`
  - stale reason: `N/A`
- Plan: `CLOSED`
  - last updated: `2026-04-07 00:05 KST`
  - stale reason: `N/A`
- Detailed Plan: `CLOSED`
  - last updated: `2026-04-07 00:06 KST`
  - stale reason: `N/A`
- Execute: `CLOSED`
  - last updated: `2026-04-07 00:08 KST`
  - stale reason: `N/A`
- Learning Continuity / Roadmap: `CLOSED`
  - last updated: `2026-04-07 00:08 KST`
  - stale reason: `N/A`
- Final Audit: `CLOSED`
  - last updated: `2026-04-07 00:09 KST`
  - stale reason: `N/A`

## 2. Request Normalization

### 2.1 Intent

- goal: 코드 변경 작업에서 설명 프로토콜이 코드 주변 설명으로도 강제되도록 규범을 수정한다
- root request in one sentence: `AGENTS.md`와 `AGENTS_WORK_TEMPLATE.md`가 code-adjacent explanation을 fail-closed로 강제하게 만든다
- 왜 이렇게 해석했는가: 사용자가 "코드 주변에 반드시 설명이 있어야 한다"고 명시했고, 문서와 템플릿을 통해 강제하라고 요청했기 때문
- 다른 해석 가능성: lesson1 코드에 바로 주석 추가
- 최종 선택 해석: 이번 작업은 규범 고정이 우선이고, 실제 코드 보강은 다음 작업으로 넘긴다

### 2.2 Explicit Deliverables

- 사용자가 명시한 필수 산출물:
  - `AGENTS.md` 수정
  - `AGENTS_WORK_TEMPLATE.md` 수정
- 사용자가 명시한 금지 사항: 없음
- 사용자가 명시한 path / naming / format / finish 요구:
  - 설명 프로토콜을 코드 주변에도 강제
  - 두 문서를 통해 강제
- 사용자가 명시하지 않았지만 누락 방지를 위해 추가한 필수 항목:
  - WORK 문서
  - 코드 주변 설명의 범위 정의
  - WORK의 completion gate 반영

### 2.3 Non-Goals

- 이번 작업의 비범위:
  - lesson1 실제 코드 주석 추가
  - 전역 `~/.codex/AGENTS.md` 수정
  - 저장소 루트 `/Users/rody/VscodeProjects/study/AGENTS.md` 수정
- 지금 하지 않는 이유:
  - 사용자가 로컬 프로젝트 지침과 템플릿 강제를 요청했고, 실제 코드 보강은 후속 집행 작업으로 분리하는 편이 더 선명함
- 나중으로 미루면 안 되는 항목과 그 이유:
  - code-adjacent explanation의 PASS/FAIL 조건
  - WORK 템플릿의 planning/audit gate

### 2.4 Requirement Trace Matrix

- R-01
  - 사용자 요구: 코드 주변에 반드시 설명이 있어야 한다
  - Checklist ID: `C-01`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: `AGENTS.md`의 코드 주변 설명 강제 규칙
  - 최종 산출물 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - 최종 상태: `PASS`
- R-02
  - 사용자 요구: 설명 프로토콜, 설명 방식을 private이어도 코드 주변에 작성하도록 강제
  - Checklist ID: `C-02`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: `AGENTS.md`에 WORK/final 응답만으로는 부족하다는 규칙 추가
  - 최종 산출물 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - 최종 상태: `PASS`
- R-03
  - 사용자 요구: `AGENTS.md`, `AGENTS_WORK_TEMPLATE.md`를 통해 강제
  - Checklist ID: `C-03`
  - 관련 Unit: `Unit-01`, `Unit-02`
  - 최종 evidence: 두 문서 모두 수정
  - 최종 산출물 / 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 최종 상태: `PASS`

### 2.5 Learning Continuity Input

- 사용자가 직접 말했거나 암묵적으로 기대하는 다음 학습 흐름: 규범 변경 후 실제 lesson 코드에도 적용되어야 함
- 이번 작업 앞에 있어야 하는 prerequisite: 기존 로컬 규범과 템플릿 존재
- 이번 작업 뒤에 자연스럽게 이어져야 하는 주제: lesson1 코드 주변 설명 보강
- 순서 변경을 허용하는 조건: 더 급한 lesson 규범 누락이 발견되는 경우
- 순서 변경이 생기면 반드시 지켜야 할 고정점: code-adjacent explanation을 실제 코드 작업에 적용한다는 흐름

## 3. Root Problem & Benefit

- 근본 문제: 설명 품질 규칙은 있었지만, reasoning이 문서나 채팅에만 남고 코드 파일에는 비어 있는 상태를 실패로 판정하는 장치가 약했다
- 왜 지금 중요한가: 이미 lesson1 코드에 주석이 거의 없다는 사용자의 문제 제기가 있었고, 앞으로 lesson이 늘수록 같은 문제가 반복될 수 있다
- 의도된 이점:
  - 코드 파일만 열어도 의도와 관측 포인트를 이해할 수 있음
  - WORK가 코드 주변 설명 계획과 검수를 빠뜨리지 않음
  - 설명 프로토콜이 실제 자산에 더 가깝게 남음
- 이점이 실제로 닫혔다고 판단할 신호:
  - `AGENTS.md`에 code-adjacent explanation 강제 규칙 존재
  - `AGENTS_WORK_TEMPLATE.md`에 계획/체크리스트/감사/완료 게이트 반영
- 핵심 불변식: 설명은 코드 밖에만 숨겨 두지 않는다
- 하드 제약: trivial narration 주석을 강제하지 않는다
- 잘못하면 생기는 열화 / 왜곡:
  - 모든 줄마다 주석을 붙이는 과잉 규칙으로 오해될 수 있음
  - developer 지침의 concise comment 원칙과 충돌할 수 있음
- COMPLETE 정의: 두 규범 문서와 WORK 기록이 닫히고 검증·commit까지 완료
- PARTIAL 정의: 방향은 맞지만 템플릿의 완료 게이트나 체크리스트에 code-adjacent explanation이 반영되지 않음
- BLOCKED 정의: 규범 간 충돌을 해소하지 못해 일관된 규칙을 못 세움

## 4. Topic Analysis

- 현재 이해한 문제 구조:
  - 로컬 `AGENTS.md`는 설명 품질과 주석 원칙을 말하지만, 코드 주변 설명이 없을 때 실패라고 못 박지는 않음
  - `AGENTS_WORK_TEMPLATE.md`는 Explanation Design과 Final Audit이 있지만 code-adjacent explanation 전용 gate는 없음
- 이번 작업이 건드리는 표면:
  - 로컬 규범
  - 로컬 WORK 템플릿
  - WORK 기록
- 숨은 가정:
  - 설명이 코드 주변에 남아야 이해 복원성이 높아진다
- 핵심 미지수:
  - 어디까지를 "반드시 주석"으로 볼지
  - 개발자 상위 지침과 어떻게 조화시킬지
- 처음 떠오른 접근:
  - `AGENTS.md`에 강제 규칙 추가
  - `AGENTS_WORK_TEMPLATE.md`에 전용 completion gate 추가
- 성공을 오판하기 쉬운 지점:
  - 좋은 말만 추가하고 실제 WORK 검수 포인트는 비워 두는 것

## 5. Analysis Critique + Repair

### 5.1 Challenge 1

- 무엇이 틀릴 수 있는가: 모든 줄 주석을 강제하는 규칙으로 읽힐 수 있다
- repair: load-bearing 코드, 비자명한 분기, 관측 포인트에만 집중한다고 명시
- 왜 repair가 더 강한가: 상위 개발자 지침의 concise comment 원칙과 조화된다

### 5.2 Challenge 2

- 무엇이 좁거나 누락될 수 있는가: `AGENTS.md`만 고치고 WORK 템플릿이 그대로면 실제 작업에서 빠질 수 있다
- repair: completion formula, Explanation Design, Detailed Task Plan, Checklist, Final Audit까지 템플릿에 연결
- 왜 repair가 더 강한가: 작업 과정 전체에 강제가 걸린다

### 5.3 Challenge 3

- 무엇이 학습/설명 목표를 놓칠 수 있는가: reasoning을 문서 링크로만 밀어 버리고 코드 옆 설명은 비워 둘 수 있다
- repair: WORK/final 응답만으로는 부족하고 코드 주변 설명이 실제로 있어야 한다고 로컬 규범에 명시
- 왜 repair가 더 강한가: 사용자의 핵심 불만을 직접 닫는다

### 5.4 Retained Framing

- 최종 채택 분석: 로컬 규범 + WORK 템플릿 양쪽에서 code-adjacent explanation을 강제
- 폐기한 분석과 이유: `lesson1 코드만 급히 주석 추가`는 증상 완화에 가깝고, 같은 문제가 다시 생길 수 있어 폐기

## 6. Research / Evidence Plan

- 먼저 확인해야 하는 질문:
  - 현재 `AGENTS.md`는 코드 주석을 어디까지 요구하는가
  - 현재 템플릿은 code-adjacent explanation을 어디서 검수하는가
- 초기 검색 키워드: `주석`, `설명`, `Explanation Design`, `Final Audit`
- 확장 키워드: `code`, `코드 주변`, `comment`, `KDoc`
- 조사할 repo 경로:
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
- 조사할 1차 자료:
  - 현재 on-disk 규범 문서
- 기대하는 증거:
  - 기존 규칙의 빈틈
  - 새로 추가해야 하는 gate 위치
- research stop condition: 강제 규칙과 검수 포인트를 어디에 넣을지 결정 가능

## 7. Evidence Ledger

- E-01
  - 질문 / 주장: 기존 로컬 `AGENTS.md`는 주석이 의도를 남겨야 한다고는 하지만 code-adjacent explanation을 failure gate로 고정하지 않았다
  - 근거 유형: `source code`
  - 자료: 기존 `AGENTS.md`의 설명 규칙 / 코드와 예제 작성 규칙 절
  - 이 자료가 닫아 준 것: 규칙 보강 필요성
  - 아직 비어 있는 것: WORK 쪽 강제 장치
- E-02
  - 질문 / 주장: 기존 `AGENTS_WORK_TEMPLATE.md`는 Explanation Design은 있었지만 코드 주변 설명 전용 gate가 약했다
  - 근거 유형: `source code`
  - 자료: 기존 템플릿의 `Explanation Design`, `Detailed Task Plan`, `Frozen Checklist`, `Final Audit`
  - 이 자료가 닫아 준 것: 템플릿에도 별도 항목을 넣어야 함
  - 아직 비어 있는 것: 없음
- E-03
  - 질문 / 주장: 사용자는 이미 lesson 코드에 주석이 없는 점을 문제로 지적했다
  - 근거 유형: `command result`
  - 자료: 현재 대화
  - 이 자료가 닫아 준 것: 추상 규칙이 아니라 실제 pain point가 있음
  - 아직 비어 있는 것: 없음

## 8. Evidence Critique + Repair

- 약한 근거: 없음
- 충돌하는 근거: developer 상위 지침의 "주석은 드물게"와 긴장 가능성
- 오래되었을 가능성이 있는 가정: 없음
- 추가 수집 필요 항목: 없음
- confidence를 낮춰야 하는 항목: 없음
- repair 후 최종 근거 세트: 현재 규범 문서 + 현재 대화의 사용자 요구

## 9. PROJECT_INTENT Fit & Learning Design

- 이번 작업이 직접 지원하는 `PROJECT_INTENT` 항목:
  - 시간이 지나도 이해를 복원 가능한 자산
  - 사람이 쉽게 배우고 다시 설명할 수 있는 결과
- 이번 작업에서 반드시 드러내야 하는 **보이지 않는 동작**:
  - 설명 프로토콜이 실제 코드 작성 규칙으로 어떻게 변하는가
- 독자가 나중에 스스로 설명할 수 있어야 하는 질문:
  - 왜 문서 설명만으로는 부족한가
  - 언제 코드 옆 설명이 반드시 필요한가
- 이번 작업의 핵심 학습 대상: 코드 주변 설명 강제 원칙
- 이번 작업의 핵심 혼동쌍:
  - 좋은 설명 vs 코드 낭독 주석
  - 모든 줄 주석 vs load-bearing 설명

### 9.1 Observability Design

- ASCII 다이어그램 필요 여부 / 이유: `N`, 규범 문서 작업이라 텍스트 규칙이 더 적합
- 로그 또는 타임라인 필요 여부 / 이유: `N`
- 시각화 페이지 필요 여부 / 이유: `N`
- 실패 실험 필요 여부 / 이유: `N`
- 비교 예제 필요 여부 / 이유: `Y`, 문서-only reasoning vs code-adjacent explanation 비교가 핵심
- 부분 재구현 필요 여부 / 이유: `N`
- 채택한 관측 장치: 규칙 문구와 completion formula
- 채택하지 않은 관측 장치와 이유: 런타임 실험은 이번 작업 목적과 직접 연결되지 않음

### 9.2 Explanation Design

- 눈에 바로 들어오게 보여 줄 구조:
  - 왜 기존 규칙이 부족했는가
  - 어떤 코드가 설명을 반드시 가져야 하는가
  - WORK에서 어떻게 검수되는가
- 읽으면서 논리가 쌓이게 할 핵심 서술:
  - reasoning이 코드 파일과 떨어져 있으면 나중에 이해 복원이 어렵다
- worked example 후보:
  - lesson1의 주석 부재 문제 제기
- 회상 anchor:
  - `문서에만 있으면 부족하다. 핵심 reasoning은 코드 옆에도 남는다.`
- 코드 스니펫 / 파일 링크로 직접 보여 줄 위치:
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
- 코드 주변에 반드시 설명을 남길 파일 / 지점:
  - `N/A`, 이번 작업은 규범 문서 변경
- 각 지점에 남길 설명 형태: `N/A`
- 각 주석이 답해야 하는 질문: `N/A`
- 이 reasoning을 코드 밖 문서에만 두면 왜 부족한가:
  - 다음 코드 작업에서 빠질 수 있기 때문

### 9.3 Learning Transfer Design

- 이번 작업이 다음 단계에 넘겨야 하는 핵심 개념: load-bearing 코드에는 code-adjacent explanation이 필요하다
- 이번 작업이 닫아 줘야 하는 prerequisite: 다음 코드 작업이 댓글/주석 계획을 생략하지 않게 함
- 다음 단계가 이 작업을 발판으로 바로 사용할 자산:
  - 수정된 `AGENTS.md`
  - 수정된 `AGENTS_WORK_TEMPLATE.md`
- 이번 작업 완료 후 독자가 스스로 답할 수 있어야 하는 질문:
  - `어떤 코드에 어떤 설명을 코드 주변에 남겨야 하는가`
- 다음 단계 진입 질문:
  - `lesson1의 어떤 클래스와 메서드가 load-bearing 코드인가`

## 10. Scope Expansion & Impact Sync

### 10.1 Included Scope

- 코드: 없음
- 테스트: 문서 검토
- 문서:
  - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_enforce_code_adjacent_explanations.md`
- 예제: 없음
- 실험: 없음
- 로그 / 관측 장치: 없음
- 프론트 / 시각화: 없음
- 설정 / 스크립트: 없음

### 10.2 Excluded Scope

- 제외 항목:
  - lesson1 코드 주석 추가
  - 저장소 루트 `AGENTS.md` 수정
  - 전역 `~/.codex/AGENTS.md` 수정
- 제외 근거: 이번 요청은 로컬 프로젝트 규범과 로컬 WORK 템플릿 강제
- 왜 지금 제외해도 개념적 완결성이 깨지지 않는가: 규범 고정과 실제 적용은 분리 가능한 두 작업이고, 이번 작업의 목표는 강제 장치 마련

### 10.3 Search Expansion Ledger

- 핵심 키워드: `주석`, `설명`, `코드 주변`, `Explanation Design`
- 동의어 / 약어 / 구명칭: `comment`, `KDoc`, `code-adjacent explanation`
- 관련 에러 / 로그 / API 이름: `N/A`
- 실제 검색 경로:
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
- 누락 가능성 점검 결과: 이번 작업에 필요한 관련 surface는 충분히 포함

### 10.4 Impact Sync Map

- 함께 움직여야 하는 자산:
  - 로컬 `AGENTS.md`
  - 로컬 `AGENTS_WORK_TEMPLATE.md`
  - 이번 WORK 문서
- 한쪽만 바꾸면 깨질 부분:
  - `AGENTS.md`만 바꾸면 실제 WORK에서 빠질 수 있음
  - 템플릿만 바꾸면 왜 필요한지와 범위 정의가 약함
- 관련 문서 동기화 항목:
  - 이후 모든 code-changing WORK
- 관련 예제 동기화 항목:
  - 후속 lesson 코드 주석 보강 작업
- 관련 관측 장치 동기화 항목:
  - 없음

## 11. Design

- 선택한 접근:
  - `AGENTS.md`에 code-adjacent explanation 강제 규칙 추가
  - `AGENTS_WORK_TEMPLATE.md`에 plan/checklist/audit/completion gate 추가
- 왜 이것이 근본 문제와 이점에 맞는가:
  - 규범 설명과 집행 템플릿이 함께 움직여야 실제 작업에서 누락이 줄어든다
- 고려한 대안 A:
  - `AGENTS.md`만 수정
- 대안 A를 채택하지 않은 이유:
  - 실제 WORK에서 빠질 수 있다
- 고려한 대안 B:
  - lesson1 코드에만 직접 주석 추가
- 대안 B를 채택하지 않은 이유:
  - 같은 문제가 다음 작업에서 다시 생길 수 있다
- 주요 계약 / 경계:
  - 모든 줄 주석 강제는 하지 않음
  - load-bearing 코드, 비자명한 분기, 관측 포인트에 집중
  - reasoning을 WORK/final 응답에만 남기는 것은 불충분
- 실패 모드:
  - 규칙이 과해 보여 오해를 부를 수 있음
  - completion gate에 연결되지 않으면 선언만 남음
- verification path:
  - `AGENTS.md`에 코드 주변 설명 강제 규칙 존재 확인
  - `AGENTS_WORK_TEMPLATE.md`에 `CODE_EXPLANATION_READY`, Explanation Design 항목, checklist/audit 항목 추가 확인

## 12. Design Critique + Repair

### 12.1 Correctness View

- 반론: 코드 주변 설명을 강제하면 주석이 과하게 늘 수 있다
- repair 또는 유지: 유지하되 trivial narration 금지와 load-bearing 범위를 함께 적음
- 이유: 규칙을 강하게 두되 오해를 줄였다

### 12.2 Learner View

- 반론: WORK 템플릿이 너무 무거워질 수 있다
- repair 또는 유지: 유지. 코드 변경 작업일 때만 사실상 required로 취급한다고 명시
- 이유: 문서-only 작업에는 `N/A with reason`이 가능하다

### 12.3 Operations / Performance View

- 반론: 실제 코드 보강 없이 규범만 바꾸면 즉시 체감이 없을 수 있다
- repair 또는 유지: 유지. 다음 단계 후보를 lesson1 코드 보강으로 연결
- 이유: 이번 작업의 범위를 흐리지 않으면서 다음 적용 경로를 남긴다

### 12.4 Final Design Decision

- 최종 채택안: 로컬 규범 + 로컬 WORK 템플릿 동시 강화
- 트레이드오프: 문서 분량은 늘지만, 이후 누락 방지 효과가 더 큼

## 13. Overall Plan

- 작업 순서:
  1. 기존 규범의 빈틈 확인
  2. code-adjacent explanation 규칙 설계
  3. `AGENTS.md` 수정
  4. `AGENTS_WORK_TEMPLATE.md` 수정
  5. WORK 문서 작성
  6. final review + diff check + commit
- 선행 의존성: 현재 두 문서 구조 확인
- validation order: 문서 자기검토 -> `git diff --check` -> 대상 파일 status -> commit
- retry order: 범위 정의 -> completion gate -> wording
- rollback / reopen 기준: 규칙이 "모든 줄 주석"처럼 읽히면 design 단계로 reopen

### 13.1 Sequence Position And Next-Step Hypothesis

- 현재 학습 경로에서의 위치: 규범 강화
- 원래 예상한 다음 단계: lesson1 코드 보강 또는 lesson2 진행
- 이번 작업 중 순서가 바뀔 수 있는 신호: 설명 규칙 자체에 추가 보강이 필요하다고 드러나는 경우
- 순서가 바뀌면 다시 확인할 prerequisite: code-adjacent explanation 규칙이 충분히 명확한가

## 14. Plan Critique + Repair

- 이 계획이 실패할 수 있는 지점: 템플릿까지 안 건드리면 강제가 약하다
- 순서상 위험: WORK를 너무 늦게 쓰면 설계 결정 근거가 흐려진다
- 빠진 prerequisite: 없음
- repair: 규범 수정과 WORK 기록을 한 작업 안에서 함께 닫음
- 왜 repair된 계획이 더 강한가: 규칙과 집행 기록이 함께 남는다

## 15. Detailed Task Plan

- Unit-01
  - 목적: 로컬 `AGENTS.md`에 code-adjacent explanation 규칙 추가
  - 대상 파일 / 자산: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - 바꿀 논리: 설명 프로토콜을 코드 주변 설명으로도 강제
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 코드 주변 설명 계획: `N/A`, 규범 문서 수정
  - 검증: 관련 절 존재 여부 확인
  - 완료 판정: 강제 규칙, completion 기준, 금지 패턴 반영
- Unit-02
  - 목적: `AGENTS_WORK_TEMPLATE.md`에 planning/checklist/audit/completion gate 추가
  - 대상 파일 / 자산: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 바꿀 논리: code-adjacent explanation 누락 시 COMPLETE 불가
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 코드 주변 설명 계획: `N/A`, 템플릿 수정
  - 검증: `CODE_EXPLANATION_READY`, Explanation Design, checklist, final audit 절 확인
  - 완료 판정: 계획-검수-완료 게이트 연결
- Unit-03
  - 목적: WORK 문서 작성과 closure
  - 대상 파일 / 자산: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_enforce_code_adjacent_explanations.md`
  - 바꿀 논리: 판단 근거, 범위, 검증, 다음 단계 기록
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 코드 주변 설명 계획: 이번 작업은 코드 변경이 없어 `N/A with reason`을 기록
  - 검증: WORK 필수 섹션 점검, diff check, status, commit
  - 완료 판정: WORK + verification + commit

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: `AGENTS.md`가 코드 주변 설명 부재를 실패 후보로 본다
  - S2: `AGENTS_WORK_TEMPLATE.md`가 code-adjacent explanation planning/audit/completion gate를 가진다
  - S3: trivial narration이 아니라 load-bearing 설명에 집중한다는 범위가 분명하다
- 실패 케이스 최소 3개:
  - F1: 설명 강제 규칙이 문서 문구에만 있고 completion gate가 없다
  - F2: "모든 줄 주석"처럼 읽히는 과한 규칙이 된다
  - F3: code-changing WORK에서 코드 주변 설명을 `N/A`로 쉽게 넘길 수 있다
- 회귀 위험:
  - future WORK가 여전히 explanation을 final 응답에만 남길 수 있음
- 회귀 방지 확인 경로:
  - 이후 code-changing WORK에서 템플릿의 새 항목 사용 여부 확인

## 16. Detailed Plan Critique + Repair

- 빠진 단위: 없음
- fuzzy success criteria: "강제된다"는 표현이 추상적일 수 있음
- 과한 범위 / 부족한 범위: 전역 규범까지 수정하면 과함
- repair: `AGENTS.md` 강제 규칙 + 템플릿 completion gate라는 두 축으로 success를 구체화
- 최종 상세 계획: Unit-01 -> Unit-02 -> Unit-03

## 17. Frozen Success / Failure Checklist

> 사용자 항목은 최소 바닥선이다. 삭제·완화 금지.
> 모든 required 항목이 PASS여야만 COMPLETE 가능.

- C-01
  - 출처: `사용자`
  - 내용: 코드 주변에 반드시 설명이 있어야 한다
  - required: `Y`
  - PASS 기준: `AGENTS.md`가 code-adjacent explanation을 명시적으로 강제
  - FAIL 기준: 코드 주변 설명 강제 규칙이 없음
  - 필요한 증거: `AGENTS.md`
  - 재시도 트리거: 강제성이 약함
  - 관련 Unit: `Unit-01`
- C-02
  - 출처: `사용자`
  - 내용: 설명 프로토콜, 설명 방식을 private이어도 코드 주변에 작성하도록 해야 한다
  - required: `Y`
  - PASS 기준: WORK/final 응답만으로는 부족하고 코드 주변 설명이 있어야 한다는 문구가 존재
  - FAIL 기준: 문서 설명만으로 대체 가능하게 남음
  - 필요한 증거: `AGENTS.md`
  - 재시도 트리거: reasoning이 코드 밖에만 남는 해석 가능
  - 관련 Unit: `Unit-01`
- C-03
  - 출처: `사용자`
  - 내용: `AGENTS.md`, `AGENTS_WORK_TEMPLATE.md`를 통해 강제
  - required: `Y`
  - PASS 기준: 두 파일 모두 수정
  - FAIL 기준: 한 파일만 수정
  - 필요한 증거: git diff, 두 파일 본문
  - 재시도 트리거: 템플릿 반영 누락
  - 관련 Unit: `Unit-01`, `Unit-02`
- C-04
  - 출처: `AI-추가`
  - 내용: developer 상위 지침과 충돌하지 않게 load-bearing 설명 중심으로 범위를 좁혀야 한다
  - required: `Y`
  - PASS 기준: trivial narration 금지와 범위 제한이 명시됨
  - FAIL 기준: 모든 줄 주석처럼 읽힘
  - 필요한 증거: `AGENTS.md`
  - 재시도 트리거: 과잉 주석 유도
  - 관련 Unit: `Unit-01`
- C-05
  - 출처: `AI-추가`
  - 내용: WORK 템플릿이 code-adjacent explanation을 계획/체크리스트/감사/완료 게이트로 닫아야 한다
  - required: `Y`
  - PASS 기준: 관련 절이 템플릿에 존재
  - FAIL 기준: 선언만 있고 WORK 검수 장치가 없음
  - 필요한 증거: `AGENTS_WORK_TEMPLATE.md`
  - 재시도 트리거: completion gate 약함
  - 관련 Unit: `Unit-02`
- C-06
  - 출처: `AI-추가`
  - 내용: repo 변경 작업이므로 verify 후 commit까지 닫아야 한다
  - required: `Y`
  - PASS 기준: verification + commit 완료
  - FAIL 기준: commit 누락
  - 필요한 증거: commit SHA
  - 재시도 트리거: commit 실패
  - 관련 Unit: `Unit-03`

### 17.1 Checklist Critique + Repair

- 각 항목이 목표 / 이점 / 사용자 요구와 매핑되는가: `Y`
- PASS / FAIL이 관측 가능한가: `Y`
- 증거가 실제로 존재 가능한가: `Y`
- 사용자 요구를 조용히 약화한 항목이 없는가: `Y`
- repair: `C-04`, `C-05`를 추가해 과잉 규칙과 템플릿 누락을 함께 막음
- Freeze 시각: `2026-04-07 00:06 KST`
- Freeze 버전: `v1`

## 18. Execution Ledger

- Attempt-01 / Unit-01
  - 시작 시각: `2026-04-07 00:04 KST`
  - 실행 내용: `AGENTS.md`에 code-adjacent explanation 강제 규칙 추가
  - 변경 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
  - 추가/보강한 코드 주변 설명: `N/A`, 규범 문서 수정
  - 새로 생긴 증거: 코드 주변 설명 강제 절, 완료 기준, 금지 패턴
  - 이번 시도에서 새로 얻은 지식: "모든 줄 주석"으로 오해되지 않게 범위를 분명히 적어야 상위 지침과 조화된다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: 본문 자기검토
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: 템플릿 반영
- Attempt-02 / Unit-02
  - 시작 시각: `2026-04-07 00:05 KST`
  - 실행 내용: `AGENTS_WORK_TEMPLATE.md`에 code-adjacent explanation plan/checklist/audit/completion gate 추가
  - 변경 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 추가/보강한 코드 주변 설명: `N/A`, 템플릿 문서 수정
  - 새로 생긴 증거: `CODE_EXPLANATION_READY`, Explanation Design 항목, checklist/audit 항목
  - 이번 시도에서 새로 얻은 지식: 템플릿까지 묶지 않으면 로컬 규범만으로는 실제 작업 누락을 막기 약하다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: 본문 자기검토
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: WORK 문서 작성 및 closure
- Attempt-03 / Unit-03
  - 시작 시각: `2026-04-07 00:08 KST`
  - 실행 내용: WORK 문서 작성, final review, diff check, commit
  - 변경 파일:
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_enforce_code_adjacent_explanations.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
  - 추가/보강한 코드 주변 설명: `N/A`, 이번 작업은 코드 변경 없음
  - 새로 생긴 증거: WORK 기록, diff check, commit
  - 이번 시도에서 새로 얻은 지식: 이 규칙의 자연스러운 다음 적용 대상은 lesson1 code comments다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: `git diff --check`, `git status --short`, commit
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: 없음

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
  - `N/A`
- 5회 시도 후에도 required 항목이 FAIL 또는 미판정이면:
  - `N/A`

### 19.2 Roadmap Reorder Log

- Change-01
  - reorder 발생 여부: `N`
  - 변경 전 순서: `구조 기준 -> 규범 강화 -> 실제 lesson 적용`
  - 변경 후 순서: `동일`
  - 변경 트리거: `N/A`
  - 왜 이 변경이 더 강한가: `N/A`
  - 뒤로 밀린 항목: 없음
  - 다음 WORK에 넘길 메모: lesson1 코드 주변 설명 보강이 자연스러운 다음 단계
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
  - evidence: `AGENTS.md`의 `코드 주변 설명 강제 규칙`
  - notes: code-adjacent explanation을 명시적으로 강제
- C-02 Final: `PASS`
  - evidence: `AGENTS.md`의 "WORK/final 응답만으로는 부족" 규칙
  - notes: private reasoning only 허용 안 함
- C-03 Final: `PASS`
  - evidence: `AGENTS.md`, `AGENTS_WORK_TEMPLATE.md` 둘 다 수정
  - notes: 사용자 요청 충족
- C-04 Final: `PASS`
  - evidence: load-bearing 범위와 trivial narration 금지 문구
  - notes: 과잉 주석 강제 방지
- C-05 Final: `PASS`
  - evidence: `CODE_EXPLANATION_READY`, Explanation Design, checklist/audit 항목
  - notes: WORK 검수 장치 연결
- C-06 Final: `PASS`
  - evidence: commit
  - notes: SHA는 아래 기록

## 21. Compliance Audit

### 21.1 PROJECT_INTENT Compliance

- 이번 작업이 `PROJECT_INTENT`의 어떤 목표를 충족했는가:
  - 이해 복원 가능한 자산 강화를 지원
  - 사람이 다시 설명할 수 있는 코드 자산 방향 강화
- `PROJECT_INTENT`와 어긋난 부분:
  - 없음
- 어긋났다면 repair:
  - `N/A`
- Final: `PASS`

### 21.2 Global AGENTS Compliance

- 전역 hard rule 중 이번 작업에 적용된 것:
  - instruction stack read/apply
  - checklist freeze
  - final review
  - verify + commit closure
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

### 21.3 Local AGENTS Compliance

- 로컬 hard rule 중 이번 작업에 적용된 것:
  - 쉬운 한국어
  - reasoning을 코드 자산 가까이에 남기려는 방향
  - 학습 복원성 강화
- 관측 장치 / 설명 품질 / 자연스러운 한국어 / 코드 링크 규칙 / 코드 주변 설명 규칙 준수 여부:
  - 준수
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

## 22. Final Audit

### 22.1 Expert Review

- 목적 적합성: 사용자의 불만을 직접 닫는 규범 변경
- 개념적 완결성: 로컬 규범과 WORK 템플릿이 함께 수정됨
- 작업 완결성: 두 규범 문서 + WORK + verification + commit
- 결과 완결성: 다음 code-changing WORK부터 바로 적용 가능
- 설명 품질: "왜 부족했는가 -> 어떻게 강제하는가 -> 어디서 검수하는가" 흐름이 닫힘
- 코드 주변 설명 존재/품질: 이번 작업은 코드 변경이 없어 `N/A with reason` 처리 규칙까지 함께 추가
- 검증 정직성: 문서 검토와 diff/status 확인만 수행했음을 명시
- 남은 이상 징후: 기존 lesson1 코드에는 아직 새 규칙이 적용되지 않음

### 22.2 Final Audit Repair

- audit에서 발견된 문제:
  - 없음
- repair:
  - `N/A`
- repair 후 재검토 결과:
  - closure 가능

## 23. Final Verification Summary

- 실제 실행한 명령 / 테스트 / 확인:
  - `rg -n "주석|설명|code|KDoc|comment" AGENTS.md AGENTS_WORK_TEMPLATE.md`
  - `sed -n`으로 두 문서 본문 확인
  - `git diff --check -- AGENTS.md AGENTS_WORK_TEMPLATE.md docs/works/20260407_enforce_code_adjacent_explanations.md`
  - `git status --short -- AGENTS.md AGENTS_WORK_TEMPLATE.md docs/works/20260407_enforce_code_adjacent_explanations.md`
- PASS 신호:
  - 로컬 규범에 code-adjacent explanation 강제 규칙 존재
  - WORK 템플릿에 `CODE_EXPLANATION_READY`와 관련 항목 존재
  - 대상 파일만 stage/commit 가능
- FAIL 신호:
  - 설명 강제 규칙이 문서 문장으로만 있고 WORK gate가 없음
  - 모든 줄 주석처럼 읽히는 과한 문구
- 코드 주변 설명 검증 결과:
  - 이번 작업은 코드 변경이 없어 `N/A`
  - 대신 이후 코드 작업에서 검수할 항목이 템플릿에 추가됨
- 실행하지 못한 검증:
  - 자동 테스트
- 실행하지 못한 이유:
  - 문서 규범 변경 작업이라 테스트 대상 없음
- 최종 confidence:
  - `high`

## 24. Final Deliverable Inventory

- D-01
  - 자산: 로컬 규범 강화
  - 유형: `문서`
  - 역할: code-adjacent explanation 강제
  - 어떤 질문에 답하는가: `왜 설명이 코드 주변에도 남아야 하는가`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- D-02
  - 자산: WORK 템플릿 강화
  - 유형: `문서`
  - 역할: code-adjacent explanation planning/audit/completion gate 강제
  - 어떤 질문에 답하는가: `작업 중 이 규칙을 어떻게 빠뜨리지 않는가`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
- D-03
  - 자산: 이번 작업 기록
  - 유형: `문서`
  - 역할: 판단 근거와 closure 기록
  - 어떤 질문에 답하는가: `왜 이 규칙을 넣었고 어떻게 검증했는가`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260407_enforce_code_adjacent_explanations.md`

## 25. Learning Continuity Closure

- 이번 작업으로 닫힌 학습 질문:
  - 왜 설명 프로토콜이 코드 주변에도 남아야 하는가
- 이번 작업이 열어 버린 새 질문:
  - lesson1의 어떤 클래스/메서드에 어떤 설명을 붙여야 하는가
- 다음 작업 후보:
  - lesson1 코드 주변 설명 보강
  - lesson2 진행 시 새 규칙 즉시 적용
- 권장 다음 작업:
  - lesson1 코드 주변 설명 보강
- 왜 지금 이 작업이 다음 단계로 가장 자연스러운가:
  - 사용자의 문제 제기가 바로 그 자산에서 나왔고, 새 규칙을 실제 코드에 처음 적용해 볼 좋은 대상이기 때문
- 순서 변경 여부: `N`
- 변경 전 로드맵:
  - lesson1
  - 구조 기준
  - 규범 강화
  - 실제 코드 보강
- 변경 후 로드맵:
  - 동일
- 뒤로 미룬 작업과 이유:
  - 전역 규범 수정은 이번 요청 범위 밖
- 다음 WORK가 이어받아야 할 자산 / evidence / commit:
  - 수정된 `AGENTS.md`
  - 수정된 `AGENTS_WORK_TEMPLATE.md`
  - 이번 commit
- 다음 작업 시작 조건:
  - lesson1에서 load-bearing 코드 위치 식별
- 더 이상 다음 작업이 필요 없다고 판단한다면 그 근거:
  - 해당 없음
- Learning chain final: `PASS`

## 26. Completion Decision

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 열린 게이트: 없음
- 미해소 blocker: 없음
- COMPLETE라고 판단하는 근거: 로컬 규범, 로컬 템플릿, WORK 기록, verification, commit 모두 닫힘
- PARTIAL / BLOCKED라면 승격 조건: `N/A`

## 27. Commit Closure

- repo 변경 여부: `Y`
- commit 요구 여부: `Y`
- commit 수행 여부: `Y`
- commit SHA: `same-commit self-reference limitation: exact SHA는 final 응답과 git history에서 확인`
- commit message: `docs: require code-adjacent explanations`
- 이 작업을 닫는 커밋인지 여부: `Y`
- push 여부: `N`
- push는 왜 했거나 하지 않았는가: 사용자 요청 없음

## 28. Closing Notes

- 이번 작업의 핵심 교훈: 설명 품질 규칙은 문서에만 있으면 약하고, 코드 주변 설명과 WORK 완료 게이트까지 연결돼야 실제로 지켜진다
- 다음 작업으로 자연스럽게 이어지는 질문: lesson1에서 어떤 코드가 load-bearing 코드이며, 각 지점에 어떤 KDoc/주석이 필요한가
- 후속 작업 필요 여부: `Y`
- 최종 종료 시각: `2026-04-07 00:09 KST`
