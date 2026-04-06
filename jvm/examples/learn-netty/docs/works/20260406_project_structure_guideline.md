# WORK_20260406_PROJECT_STRUCTURE_GUIDELINE

> 학습 자산 구조 기준 문서를 만드는 작업 기록입니다.

## 0. Meta

- 작업 제목: 학습 자산 구조 기준 문서 작성
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_project_structure_guideline.md`
- 파일명 규칙 확인: `docs/works/YYYYMMDD_${작업요약}.md`
- 저장소 / 모듈: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty`
- 작업 유형: `analysis | design | explain | mixed`
- 작업 깊이: `standard`
- 실행자: `Codex`
- 시작 일시: `2026-04-06 23:34 KST`
- 종료 일시: `2026-04-06 23:48 KST`
- 관련 요청 / 이슈: 사용자의 구조 기준 문서화 요청
- 원문 사용자 요청: 단계별로 발전하는 과정을 비교하면서 보기 위한 구조를 패키지/서브프로젝트/프론트 관점에서 먼저 문서로 고정
- run_mode: `normal`
- finish: `verify+commit`
- repo 변경 여부: `Y`
- commit 기본 요구 여부: `Y`
- 직전 WORK / 선행 작업: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_first_lesson_why_netty.md`
- 후속 WORK / 다음 작업 후보: `Lesson 02 EventLoop / boss / worker`
- 이번 작업의 학습 단계 위치: `프로젝트 구조 기준 고정`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`

### 0.1 Usage Policy

- 이번 작업은 문서 작업이지만, 프로젝트 전체의 기준을 고정하는 작업이라 `standard`보다 강하게 검토했습니다.
- 문서 본문만 쓰고 끝내지 않고, WORK 기록과 commit까지 closure에 포함했습니다.
- 코드 리팩토링은 이번 작업 범위에 넣지 않았습니다.

### 0.2 Learning Chain Policy

- 이 문서는 개별 lesson보다 상위의 구조 기준이지만, 결국 각 lesson이 어디에 놓이는지와 다음 lesson이 어떻게 이어지는지 돕기 위해 작성했습니다.
- 따라서 현재 구조 기준과 다음 lesson 진행 방향이 함께 보이도록 기록했습니다.
- 이 문서는 템플릿이 아니라 실제 프로젝트 구조 판단 기준입니다.

## 1. Instruction Stack Preflight

### 1.1 Resolved Sources

- 전역 `AGENTS.md` 경로: `/Users/rody/.codex/AGENTS.md`
- 전역 `AGENTS.md` 읽음: `Y`
- 전역 `AGENTS.md` 적용 요약: fail-closed completion, WORK ledger, commit 포함 closure 적용
- `PROJECT_INTENT.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md`
- `PROJECT_INTENT.md` 읽음: `Y`
- `PROJECT_INTENT.md` 적용 요약: 단계별 학습 경로, 다음 학습 연결, 보이는 비교 자산을 우선
- 로컬 `AGENTS.md` 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 로컬 `AGENTS.md` 읽음: `Y`
- 로컬 `AGENTS.md` 적용 요약: 대부분 요청을 full로 해석하지만, 이번 작업은 기준 문서 성격에 맞게 structure-first로 진행
- 이 템플릿 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md`
- 이 템플릿 사용 시작 시각: `2026-04-06 23:34 KST`

### 1.2 Project Overlay

- 로컬 규범 경로: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md`
- 이번 작업에서 활성화한 프로젝트 규칙:
  - 단계형 학습 경로 우선
  - 보이지 않는 동작을 드러내는 자산 구조
  - 다음 lesson으로의 연결
- 활성화 이유: 구조 기준이 이후 lesson 배치와 공통 support 승격 규칙을 좌우하기 때문
- 이 규칙을 실제로 검증할 방법: 문서에서 lesson 중심 구조, support 기준, subproject 승격 조건이 모두 명시되는지 확인
- 전역 규칙과의 충돌 가능성: 없음
- 충돌 해소 방식: `N/A`

### 1.3 Effective Contract

- 이번 작업에서 특히 강한 hard rule:
  - 지금 기본값과 나중 승격 기준을 분리해서 적을 것
  - "무조건 패키지"나 "무조건 서브프로젝트"처럼 성급히 고정하지 않을 것
- 충돌 가능성:
  - 문서가 추상론으로 흐를 수 있음
- 충돌 해소 방식:
  - 현재 구조, 판단 규칙, 예시 구조, 승격 조건을 모두 적음
- preflight blocker: 없음

### 1.4 Completion Formula

- `INSTRUCTION_STACK_READY = PASS`
- `LEARNING_CHAIN_READY = PASS`
- `ALLOW_COMPLETE = structure guideline + WORK record + audits + commit`
- 하나라도 FAIL 또는 미판정이면 `BLOCK_COMPLETE`

### 1.5 Phase Status Board

- Topic Analysis: `CLOSED`
  - last updated: `2026-04-06 23:36 KST`
  - stale reason: `N/A`
- Research / Evidence: `CLOSED`
  - last updated: `2026-04-06 23:38 KST`
  - stale reason: `N/A`
- Design: `CLOSED`
  - last updated: `2026-04-06 23:40 KST`
  - stale reason: `N/A`
- Plan: `CLOSED`
  - last updated: `2026-04-06 23:40 KST`
  - stale reason: `N/A`
- Detailed Plan: `CLOSED`
  - last updated: `2026-04-06 23:41 KST`
  - stale reason: `N/A`
- Execute: `CLOSED`
  - last updated: `2026-04-06 23:46 KST`
  - stale reason: `N/A`
- Learning Continuity / Roadmap: `CLOSED`
  - last updated: `2026-04-06 23:46 KST`
  - stale reason: `N/A`
- Final Audit: `CLOSED`
  - last updated: `2026-04-06 23:48 KST`
  - stale reason: `N/A`

## 2. Request Normalization

### 2.1 Intent

- goal: 앞으로 lesson을 어떻게 배치할지 기준 문서를 먼저 고정한다
- root request in one sentence: lesson 중심 패키지와 서브프로젝트 승격 기준을 문서로 정한다
- 왜 이렇게 해석했는가: 직전 대화에서 사용자가 "먼저 기준 문서"를 원했고 그 방향에 동의했기 때문
- 다른 해석 가능성: 곧바로 코드 구조 리팩토링
- 최종 선택 해석: 이번에는 기준 문서만 작성하고, 코드 리팩토링은 하지 않음

### 2.2 Explicit Deliverables

- 사용자가 명시한 필수 산출물:
  - 구조 기준 문서
- 사용자가 명시한 금지 사항: 없음
- 사용자가 명시한 path / naming / format / finish 요구:
  - 문서 우선
  - repo change 기본 closure 준수
- 사용자가 명시하지 않았지만 누락 방지를 위해 추가한 필수 항목:
  - WORK 문서
  - lesson/package/support/subproject/viewer 기준 모두 포함

### 2.3 Non-Goals

- 이번 작업의 비범위:
  - 실제 패키지 이동
  - 멀티모듈 전환
  - lesson1 코드 리팩토링
- 지금 하지 않는 이유:
  - 먼저 기준을 고정한 뒤 적용하는 편이 더 안전
- 나중으로 미루면 안 되는 항목과 그 이유:
  - support와 subproject의 경계 기준
  - viewer를 나중에 붙일 때의 준비 방향

### 2.4 Requirement Trace Matrix

- R-01
  - 사용자 요구: 단계별로 발전 과정을 비교하면서 보기 위한 구조 고민
  - Checklist ID: `C-01`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: lesson 중심 구조 기준 문서
  - 최종 산출물 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/project-structure.md`
  - 최종 상태: `PASS`
- R-02
  - 사용자 요구: 패키지 / 서브프로젝트 / 프론트 시각화를 함께 고려
  - Checklist ID: `C-02`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: 문서에 세 축 모두 반영
  - 최종 산출물 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/project-structure.md`
  - 최종 상태: `PASS`
- R-03
  - 사용자 요구: 문서부터 만들기
  - Checklist ID: `C-03`
  - 관련 Unit: `Unit-01`
  - 최종 evidence: 문서 생성, 코드 변경 없음
  - 최종 산출물 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/project-structure.md`
  - 최종 상태: `PASS`

### 2.5 Learning Continuity Input

- 사용자가 직접 말했거나 암묵적으로 기대하는 다음 학습 흐름: lesson이 단계적으로 쌓여야 함
- 이번 작업 앞에 있어야 하는 prerequisite: lesson1 존재
- 이번 작업 뒤에 자연스럽게 이어져야 하는 주제: lesson2 진행
- 순서 변경을 허용하는 조건: viewer나 benchmark가 실제 artifact로 빨리 생기는 경우
- 순서 변경이 생기면 반드시 지켜야 할 고정점: lesson 발전 과정이 눈에 보이는 구조

## 3. Root Problem & Benefit

- 근본 문제: 지금 구조를 고정하지 않으면, lesson이 늘어날수록 패키지/서브프로젝트 기준이 흔들릴 수 있다
- 왜 지금 중요한가: lesson1을 만든 직후라 앞으로의 자산이 빠르게 늘어날 가능성이 크다
- 의도된 이점:
  - lesson 중심 배치 기준 고정
  - support 승격 기준 고정
  - viewer/subproject 승격 기준 고정
- 이점이 실제로 닫혔다고 판단할 신호:
  - 현재 기본값과 미래 승격 조건이 분리되어 적힘
- 핵심 불변식: 구조는 학습 경로를 더 잘 보이게 해야 한다
- 하드 제약: 현재 lesson1을 무리하게 리팩토링하지 않음
- 잘못하면 생기는 열화 / 왜곡:
  - 제품 아키텍처 사고가 학습 구조를 압도
  - 공통화가 lesson 맥락을 가림
- COMPLETE 정의: 기준 문서 + WORK + commit
- PARTIAL 정의: 방향은 맞지만 승격 조건이나 support 기준이 빠짐
- BLOCKED 정의: 현재 구조를 판단할 근거 부족

## 4. Topic Analysis

- 현재 이해한 문제 구조:
  - 지금은 single-module, lesson1 존재
  - 앞으로 lesson과 viewer가 늘 수 있음
- 이번 작업이 건드리는 표면:
  - `docs` 구조 기준
  - lesson 배치 기준
  - support 기준
  - subproject 승격 조건
- 숨은 가정:
  - 학습 프로젝트는 배포 구조보다 학습 경로가 우선
- 핵심 미지수:
  - 기준 문서 위치
  - 얼마나 강하게 “지금은 패키지”를 못 박을지
- 처음 떠오른 접근:
  - `docs/project-structure.md`
- 성공을 오판하기 쉬운 지점:
  - 단순 의견문으로 끝나고 운영 규칙이 비는 것

## 5. Analysis Critique + Repair

### 5.1 Challenge 1

- 무엇이 틀릴 수 있는가: 문서가 추상적 취향 논쟁으로 흐를 수 있다
- repair: 현재 기본값, support 기준, subproject 승격 조건을 규칙으로 씀
- 왜 repair가 더 강한가: 실제 판단 기준이 남는다

### 5.2 Challenge 2

- 무엇이 좁거나 누락될 수 있는가: 프론트 시각화는 언급만 하고 구조 연결을 안 남길 수 있다
- repair: event 모델/console sink/viewer 확장 방향을 명시
- 왜 repair가 더 강한가: 지금과 나중이 연결된다

### 5.3 Challenge 3

- 무엇이 학습/설명 목표를 놓칠 수 있는가: 서브프로젝트 장점만 보고 지금도 그렇게 가자고 결론낼 수 있다
- repair: 왜 지금은 lesson 패키지가 맞는지 학습 관점으로 설명
- 왜 repair가 더 강한가: 프로젝트 의도와 더 일치한다

### 5.4 Retained Framing

- 최종 채택 분석: 현재 기본값은 lesson 중심 single-module
- 폐기한 분석과 이유: 즉시 멀티모듈 전환은 학습 경로 가시성에 비해 이득이 작아 폐기

## 6. Research / Evidence Plan

- 먼저 확인해야 하는 질문:
  - 현재 디렉터리 구조는 어떤가
  - lesson1은 어떤 형태로 존재하는가
  - viewer artifact는 이미 있는가
- 초기 검색 키워드: `docs/lessons`, `src/main/java`, `target`, `lesson1`
- 확장 키워드: `support`, `viewer`, `subproject`
- 조사할 repo 경로:
  - 현재 프로젝트 전체
- 조사할 1차 자료:
  - 현재 on-disk repo 구조
  - lesson1 문서
- 기대하는 증거:
  - 지금은 single-module 기준이 자연스럽다는 근거
- research stop condition: 구조 기준을 정하기에 충분한 현재 상태 파악 완료

## 7. Evidence Ledger

- E-01
  - 질문 / 주장: 현재 프로젝트는 single-module이고 lesson1이 이미 존재한다
  - 근거 유형: `repo evidence`
  - 자료: `find` 결과
  - 이 자료가 닫아 준 것: 문서는 `docs`, 코드는 `src`에 있고 멀티모듈은 아직 없음
  - 아직 비어 있는 것: 없음
- E-02
  - 질문 / 주장: lesson1은 비교형 강의로 구성돼 있다
  - 근거 유형: `repo evidence`
  - 자료: `docs/lessons/lesson-01-why-netty.md`
  - 이 자료가 닫아 준 것: lesson 중심 구조가 현재 자산과 잘 맞음
  - 아직 비어 있는 것: 없음
- E-03
  - 질문 / 주장: viewer는 아직 실제 artifact가 아니다
  - 근거 유형: `repo evidence`
  - 자료: 현재 디렉터리에 프론트 앱/별도 모듈 없음
  - 이 자료가 닫아 준 것: 지금 바로 subproject로 갈 이유가 약함
  - 아직 비어 있는 것: 없음

## 8. Evidence Critique + Repair

- 약한 근거: 없음
- 충돌하는 근거: 없음
- 오래되었을 가능성이 있는 가정: viewer가 빠르게 생기면 기준을 다시 볼 수 있음
- 추가 수집 필요 항목: 없음
- confidence를 낮춰야 하는 항목: 미래 확장 속도
- repair 후 최종 근거 세트: 현재 repo 구조 + lesson1 자산

## 9. PROJECT_INTENT Fit & Learning Design

- 이번 작업이 직접 지원하는 `PROJECT_INTENT` 항목:
  - 단계별 학습 경로
  - 다음 학습 경로 가시성
  - 장기 복원 가능한 자산 구조
- 이번 작업에서 반드시 드러내야 하는 **보이지 않는 동작**:
  - 학습 자산이 어떻게 누적되고 support로 승격되는지
- 독자가 나중에 스스로 설명할 수 있어야 하는 질문:
  - 왜 지금은 패키지 중심이 맞는가
  - 언제 서브프로젝트로 가는가
- 이번 작업의 핵심 학습 대상: 프로젝트 구조 판단 기준
- 이번 작업의 핵심 혼동쌍:
  - 학습 구조 vs 제품 구조
  - 공통화 vs lesson 가시성

### 9.1 Observability Design

- ASCII 다이어그램 필요 여부 / 이유: `Y`, 구조 예시가 바로 보여야 함
- 로그 또는 타임라인 필요 여부 / 이유: `N`, 구조 기준 문서라 불필요
- 시각화 페이지 필요 여부 / 이유: `N`
- 실패 실험 필요 여부 / 이유: `N`
- 비교 예제 필요 여부 / 이유: `Y`, 패키지 vs 서브프로젝트의 판단 축 비교
- 부분 재구현 필요 여부 / 이유: `N`
- 채택한 관측 장치: directory tree 예시
- 채택하지 않은 관측 장치와 이유: 런타임 실험은 문서 목적에 비해 과함

### 9.2 Explanation Design

- 눈에 바로 들어오게 보여 줄 구조:
  - 결론
  - 왜 지금은 lesson 패키지인가
  - 현재 기본 구조
  - support / subproject 승격 기준
- 읽으면서 논리가 쌓이게 할 핵심 서술:
  - 지금 중요한 것은 artifact 경계보다 lesson 경계다
- worked example 후보: lesson1
- 회상 anchor: `지금은 lesson 중심, artifact가 자라면 subproject`
- 코드 스니펫 / 파일 링크로 직접 보여 줄 위치:
  - structure doc 자체
  - lesson1 문서

### 9.3 Learning Transfer Design

- 이번 작업이 다음 단계에 넘겨야 하는 핵심 개념: lesson 최상위, support 최소 승격, artifact가 자라면 subproject
- 이번 작업이 닫아 줘야 하는 prerequisite: 다음 lesson을 어디에 둘지 헷갈리지 않게 함
- 다음 단계가 이 작업을 발판으로 바로 사용할 자산: structure guideline 문서
- 이번 작업 완료 후 독자가 스스로 답할 수 있어야 하는 질문: `lesson2는 어디에 두고, 공통 코드는 언제 support로 올리는가`
- 다음 단계 진입 질문: `lesson2 EventLoop 자산은 lesson2 아래에서 어떤 하위 축으로 나눌 것인가`

## 10. Scope Expansion & Impact Sync

### 10.1 Included Scope

- 코드: 없음
- 테스트: 수동 검토
- 문서:
  - `docs/project-structure.md`
  - `docs/works/20260406_project_structure_guideline.md`
- 예제: lesson1 문서 참조
- 실험: 없음
- 로그 / 관측 장치: 없음
- 프론트 / 시각화: 문서에서만 고려
- 설정 / 스크립트: 없음

### 10.2 Excluded Scope

- 제외 항목:
  - lesson1 코드 이동
  - pom.xml 변경
  - viewer 모듈 생성
- 제외 근거: 이번 요청은 기준 문서 우선
- 왜 지금 제외해도 개념적 완결성이 깨지지 않는가: 기준을 먼저 고정하는 것이 선행 단계

### 10.3 Search Expansion Ledger

- 핵심 키워드: `lesson`, `src`, `docs`, `support`
- 동의어 / 약어 / 구명칭: `module`, `viewer`, `shared`
- 관련 에러 / 로그 / API 이름: `N/A`
- 실제 검색 경로: 현재 프로젝트 전체
- 누락 가능성 점검 결과: 기준 문서 작성에는 충분

### 10.4 Impact Sync Map

- 함께 움직여야 하는 자산:
  - 구조 기준 문서
  - WORK 문서
- 한쪽만 바꾸면 깨질 부분:
  - 문서만 있고 기록이 없으면 판단 근거가 흐려짐
- 관련 문서 동기화 항목:
  - future lesson들이 이 기준을 따르게 됨
- 관련 예제 동기화 항목:
  - 현재는 없음
- 관련 관측 장치 동기화 항목:
  - 없음

## 11. Design

- 선택한 접근: `docs/project-structure.md`에 현재 기본값과 미래 승격 기준을 함께 명시
- 왜 이것이 근본 문제와 이점에 맞는가: 기준을 한 문서에 모아 두면 이후 lesson마다 반복 설명이 줄어든다
- 고려한 대안 A: lesson1 문서 끝에 짧은 메모 추가
- 대안 A를 채택하지 않은 이유: 프로젝트 전체 기준으로는 너무 약함
- 고려한 대안 B: 바로 코드 리팩토링
- 대안 B를 채택하지 않은 이유: 기준 없이 움직이면 다시 흔들릴 수 있음
- 주요 계약 / 경계:
  - 지금은 lesson 중심 single-module
  - 여러 lesson 공유분만 support
  - viewer/benchmark/shared artifact가 자라면 subproject
- 실패 모드:
  - 너무 추상적이어서 실제 판단에 못 씀
- verification path:
  - 문서에 결론/현재 구조/승격 기준/예시 구조/판정 규칙이 모두 존재하는지 점검

## 12. Design Critique + Repair

### 12.1 Correctness View

- 반론: 구조 기준은 코드로 강제되지 않으니 약할 수 있다
- repair 또는 유지: 유지. 지금 단계에서는 문서 기준 고정이 먼저
- 이유: 아직 자산 규모가 작아 코드 강제보다 기준 합의가 더 중요

### 12.2 Learner View

- 반론: support와 lesson 경계가 헷갈릴 수 있다
- repair 또는 유지: `둘 이상의 lesson에서 함께 쓰는 것만 support`로 명시
- 이유: 가장 실용적인 1차 판단 규칙

### 12.3 Operations / Performance View

- 반론: 나중에 viewer나 benchmark가 붙으면 문서가 금방 낡을 수 있다
- repair 또는 유지: 승격 조건을 명시해 future-proofing
- 이유: 현재 기준과 변경 기준을 함께 적어 둠

### 12.4 Final Design Decision

- 최종 채택안: 현재 기본값 + 미래 승격 조건을 함께 적는 구조 기준 문서
- 트레이드오프: 실제 코드 강제력은 약하지만, 지금 단계에서는 가장 단순하고 명확함

## 13. Overall Plan

- 작업 순서:
  1. 현재 구조 확인
  2. lesson1과 미래 viewer 관점 정리
  3. structure guideline 작성
  4. WORK 문서 작성
  5. final audit + commit
- 선행 의존성: 현재 디렉터리 구조 파악
- validation order: 문서 자기검토 -> git diff check -> commit
- retry order: structure decision -> wording -> file location
- rollback / reopen 기준: 기준이 추상적이면 design 단계로 reopen

### 13.1 Sequence Position And Next-Step Hypothesis

- 현재 학습 경로에서의 위치: structure guideline
- 원래 예상한 다음 단계: lesson2 EventLoop
- 이번 작업 중 순서가 바뀔 수 있는 신호: viewer artifact를 즉시 만들기로 바뀌는 경우
- 순서가 바뀌면 다시 확인할 prerequisite: actual artifact separation need

## 14. Plan Critique + Repair

- 이 계획이 실패할 수 있는 지점: 문서 위치가 lesson 문서와 섞여 보일 수 있음
- 순서상 위험: 기준 문서 없이 lesson2가 먼저 생기면 흔들릴 수 있음
- 빠진 prerequisite: 없음
- repair: docs root에 프로젝트 기준 문서로 배치
- 왜 repair된 계획이 더 강한가: lesson 문서와 WORK 문서 사이에서 역할이 선명해짐

## 15. Detailed Task Plan

- Unit-01
  - 목적: 구조 기준 문서 작성
  - 대상 파일 / 자산: `docs/project-structure.md`
  - 바꿀 논리: 패키지/lesson/support/subproject 기준 고정
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명, ASCII 구조
  - 검증: 문서 자기검토
  - 완료 판정: 현재 기본값과 승격 조건이 모두 존재
- Unit-02
  - 목적: WORK 문서 작성
  - 대상 파일 / 자산: `docs/works/20260406_project_structure_guideline.md`
  - 바꿀 논리: 판단 근거와 closure 기록
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 검증: 필수 섹션 점검
  - 완료 판정: 기록 완결
- Unit-03
  - 목적: final audit와 commit
  - 대상 파일 / 자산: 두 문서
  - 바꿀 논리: wording 보정, closure
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것: 설명
  - 검증: git diff check, status, commit
  - 완료 판정: commit 완료

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: 현재 기본값이 lesson 중심 single-module로 명시됨
  - S2: support 승격 기준이 있음
  - S3: subproject 승격 조건이 있음
- 실패 케이스 최소 3개:
  - F1: 패키지냐 서브프로젝트냐 의견만 있고 판단 규칙이 없음
  - F2: viewer를 고려한다고 했지만 현재 구조와의 연결이 없음
  - F3: 현재 기본값과 미래 승격 기준이 섞여 혼란스러움
- 회귀 위험: future lesson이 다른 기준으로 흘러갈 수 있음
- 회귀 방지 확인 경로: 이후 WORK에서 이 문서 참조

## 16. Detailed Plan Critique + Repair

- 빠진 단위: 없음
- fuzzy success criteria: "구조 기준이 정리됐다"가 추상적일 수 있음
- 과한 범위 / 부족한 범위: code refactor까지 가면 과함
- repair: success criteria를 current default / support / subproject / viewer 네 축으로 구체화
- 최종 상세 계획: Unit-01 -> Unit-02 -> Unit-03

## 17. Frozen Success / Failure Checklist

- C-01
  - 출처: `사용자`
  - 내용: 단계별 발전 과정을 비교하면서 보기 위한 구조 기준이 있어야 한다
  - required: `Y`
  - PASS 기준: lesson 중심 배치 기준이 명시됨
  - FAIL 기준: 비교 관점이 비어 있음
  - 필요한 증거: `docs/project-structure.md`
  - 재시도 트리거: lesson 축이 안 보임
  - 관련 Unit: `Unit-01`
- C-02
  - 출처: `사용자`
  - 내용: 패키지 / 서브프로젝트 / 프론트를 함께 고려해야 한다
  - required: `Y`
  - PASS 기준: 세 축 모두 문서에 존재
  - FAIL 기준: 하나라도 빠짐
  - 필요한 증거: 문서 본문
  - 재시도 트리거: 고려 축 누락
  - 관련 Unit: `Unit-01`
- C-03
  - 출처: `사용자`
  - 내용: 문서부터 만들기
  - required: `Y`
  - PASS 기준: 구조 문서 생성, 코드 이동 없음
  - FAIL 기준: 기준 문서 없이 리팩토링 진행
  - 필요한 증거: 새 문서, git diff
  - 재시도 트리거: 문서 없이 구현으로 새나감
  - 관련 Unit: `Unit-01`
- C-04
  - 출처: `AI-추가`
  - 내용: 현재 기본값과 미래 승격 조건이 분리돼 적혀야 한다
  - required: `Y`
  - PASS 기준: now vs later가 문서에서 구분됨
  - FAIL 기준: 혼재돼 판단 규칙이 흐림
  - 필요한 증거: 문서 본문
  - 재시도 트리거: 구조 기준이 모호함
  - 관련 Unit: `Unit-01`
- C-05
  - 출처: `AI-추가`
  - 내용: WORK 문서와 준수 기록이 남아야 한다
  - required: `Y`
  - PASS 기준: WORK 문서 완성
  - FAIL 기준: 기록 누락
  - 필요한 증거: WORK 문서
  - 재시도 트리거: preflight/audit 누락
  - 관련 Unit: `Unit-02`
- C-06
  - 출처: `AI-추가`
  - 내용: repo 변경 작업이므로 commit까지 닫아야 한다
  - required: `Y`
  - PASS 기준: commit 완료
  - FAIL 기준: commit 누락
  - 필요한 증거: commit SHA
  - 재시도 트리거: commit 실패
  - 관련 Unit: `Unit-03`

### 17.1 Checklist Critique + Repair

- 각 항목이 목표 / 이점 / 사용자 요구와 매핑되는가: `Y`
- PASS / FAIL이 관측 가능한가: `Y`
- 증거가 실제로 존재 가능한가: `Y`
- 사용자 요구를 조용히 약화한 항목이 없는가: `Y`
- repair: `C-04` 추가로 now/later 분리 강제
- Freeze 시각: `2026-04-06 23:41 KST`
- Freeze 버전: `v1`

## 18. Execution Ledger

- Attempt-01 / Unit-01
  - 시작 시각: `2026-04-06 23:36 KST`
  - 실행 내용: 현재 구조와 lesson1 확인 후 structure guideline 작성
  - 변경 파일: `docs/project-structure.md`
  - 새로 생긴 증거: 구조 기준 문서
  - 이번 시도에서 새로 얻은 지식: lesson1이 이미 비교형 구조를 갖고 있어 lesson 중심 기준이 더 자연스럽다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: 문서 자기검토
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: WORK 문서 작성
- Attempt-02 / Unit-02
  - 시작 시각: `2026-04-06 23:44 KST`
  - 실행 내용: WORK 문서 작성
  - 변경 파일: `docs/works/20260406_project_structure_guideline.md`
  - 새로 생긴 증거: 집행 기록
  - 이번 시도에서 새로 얻은 지식: structure 기준 문서는 docs root가 lesson 문서와 가장 잘 분리된다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: 필수 섹션 점검
  - 결과: `PASS`
  - 실패 시 근본 원인: `N/A`
  - 실패 시 earliest affected phase: `N/A`
  - 다음 조치: final audit, commit
- Attempt-03 / Unit-03
  - 시작 시각: `2026-04-06 23:47 KST`
  - 실행 내용: diff check, final review, commit
  - 변경 파일:
    - `docs/project-structure.md`
    - `docs/works/20260406_project_structure_guideline.md`
  - 새로 생긴 증거: git diff/check, commit
  - 이번 시도에서 새로 얻은 지식: 기준 문서를 먼저 두면 다음 lesson 설계 대화가 짧고 선명해진다
  - 계획 / 순서 변경 여부: `N`
  - 수행한 검증: `git diff --check`, git status, commit
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
  - 변경 전 순서: `lesson1 -> structure guideline -> lesson2`
  - 변경 후 순서: `동일`
  - 변경 트리거: `N/A`
  - 왜 이 변경이 더 강한가: `N/A`
  - 뒤로 밀린 항목: 없음
  - 다음 WORK에 넘길 메모: lesson2는 이 기준 문서를 직접 따라가면 됨
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
  - evidence: structure doc의 lesson 중심 구조 절
  - notes: 비교형 발전 과정을 최상위 축으로 고정
- C-02 Final: `PASS`
  - evidence: package/support/subproject/viewer 절
  - notes: 세 축 모두 반영
- C-03 Final: `PASS`
  - evidence: 문서 생성, 코드 변경 없음
  - notes: 요청 충족
- C-04 Final: `PASS`
  - evidence: 문서의 `현재 기본 구조`, `언제 서브프로젝트로 가는가`
  - notes: now/later 분리됨
- C-05 Final: `PASS`
  - evidence: 이 WORK 문서
  - notes: 기록 완결
- C-06 Final: `PASS`
  - evidence: commit
  - notes: SHA는 아래 기록

## 21. Compliance Audit

### 21.1 PROJECT_INTENT Compliance

- 이번 작업이 `PROJECT_INTENT`의 어떤 목표를 충족했는가:
  - 단계별 학습 경로를 더 명확히 함
  - 다음 학습 연결성을 구조 수준에서 강화
- `PROJECT_INTENT`와 어긋난 부분:
  - 없음
- 어긋났다면 repair:
  - `N/A`
- Final: `PASS`

### 21.2 Global AGENTS Compliance

- 전역 hard rule 중 이번 작업에 적용된 것:
  - instruction stack read/apply
  - checklist freeze
  - WORK record
  - final review
  - commit closure
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

### 21.3 Local AGENTS Compliance

- 로컬 hard rule 중 이번 작업에 적용된 것:
  - 단계형 학습 경로 우선
  - 쉬운 한국어
  - 다음 단계 연결
- 관측 장치 / 설명 품질 / 자연스러운 한국어 / 코드 링크 규칙 준수 여부:
  - 준수
- 누락 또는 위반 가능성:
  - 없음
- repair:
  - `N/A`
- Final: `PASS`

## 22. Final Audit

### 22.1 Expert Review

- 목적 적합성: 기준 문서 목적에 맞음
- 개념적 완결성: 현재 기본값, support, viewer, subproject 승격이 연결됨
- 작업 완결성: 문서, WORK, commit 모두 존재
- 결과 완결성: 다음 lesson 설계에 바로 쓸 수 있음
- 설명 품질: 추상론보다 판단 규칙 중심
- 검증 정직성: 문서 검토와 diff check만 수행했음을 명시
- 남은 이상 징후: 실제 자산이 늘면 기준 조정 가능성은 있음

### 22.2 Final Audit Repair

- audit에서 발견된 문제:
  - 없음
- repair:
  - `N/A`
- repair 후 재검토 결과:
  - closure 가능

## 23. Final Verification Summary

- 실제 실행한 명령 / 테스트 / 확인:
  - `find . -maxdepth 3 \( -path './docs/*' -o -path './src/*' \) | sort | sed -n '1,160p'`
  - `sed -n`으로 lesson1 문서 확인
  - `git diff --check -- docs/project-structure.md docs/works/20260406_project_structure_guideline.md`
  - `git status --short -- docs/project-structure.md docs/works/20260406_project_structure_guideline.md`
- PASS 신호:
  - 문서에 current default / support / viewer / subproject 기준 존재
  - 문서 파일과 WORK 파일이 생성됨
- FAIL 신호:
  - 판단 규칙 누락
  - now/later 기준 혼재
- 실행하지 못한 검증:
  - 자동 테스트
- 실행하지 못한 이유:
  - 문서 작업이라 별도 테스트 불필요
- 최종 confidence:
  - `high`

## 24. Final Deliverable Inventory

- D-01
  - 자산: 구조 기준 문서
  - 유형: `문서`
  - 역할: 학습 자산 배치 기준 고정
  - 어떤 질문에 답하는가: `지금은 패키지인가, 서브프로젝트인가?`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/project-structure.md`
- D-02
  - 자산: WORK 문서
  - 유형: `문서`
  - 역할: 판단 근거와 closure 기록
  - 어떤 질문에 답하는가: `이 기준은 어떤 근거로 결정됐는가?`
  - 주요 경로 / 파일: `/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/docs/works/20260406_project_structure_guideline.md`
- D-03
  - 자산: commit
  - 유형: `기타`
  - 역할: 이번 기준 문서 작업 마감
  - 어떤 질문에 답하는가: `이 작업은 어떤 커밋으로 닫혔는가?`
  - 주요 경로 / 파일: `git history`

## 25. Learning Continuity Closure

- 이번 작업으로 닫힌 학습 질문:
  - 지금 단계에서 기본 구조는 무엇인가
- 이번 작업이 열어 버린 새 질문:
  - lesson2의 하위 패키지는 어떤 축으로 나눌 것인가
- 다음 작업 후보:
  - lesson2 EventLoop / boss / worker
  - support/observability 구조 보강
- 권장 다음 작업:
  - lesson2 EventLoop / boss / worker
- 왜 지금 이 작업이 다음 단계로 가장 자연스러운가:
  - 기준이 생겼으니 이제 lesson2를 그 기준에 맞게 바로 만들 수 있음
- 순서 변경 여부: `N`
- 변경 전 로드맵:
  - lesson1
  - structure guideline
  - lesson2
- 변경 후 로드맵:
  - 동일
- 뒤로 미룬 작업과 이유:
  - viewer artifact는 아직 실제 필요가 없음
- 다음 WORK가 이어받아야 할 자산 / evidence / commit:
  - structure guideline 문서
  - 이 WORK 문서
  - 이번 commit
- 다음 작업 시작 조건:
  - lesson2의 핵심 질문 확정
- 더 이상 다음 작업이 필요 없다고 판단한다면 그 근거:
  - 해당 없음
- Learning chain final: `PASS`

## 26. Completion Decision

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 열린 게이트: 없음
- 미해소 blocker: 없음
- COMPLETE라고 판단하는 근거: 기준 문서, WORK, 검토, commit 모두 완료
- PARTIAL / BLOCKED라면 승격 조건: `N/A`

## 27. Commit Closure

- repo 변경 여부: `Y`
- commit 요구 여부: `Y`
- commit 수행 여부: `Y`
- commit SHA: `same-commit self-reference limitation: exact SHA는 final 응답과 git history에서 확인`
- commit message: `docs: add project structure guideline`
- 이 작업을 닫는 커밋인지 여부: `Y`
- push 여부: `N`
- push는 왜 했거나 하지 않았는가: 사용자 요청 없음

## 28. Closing Notes

- 이번 작업의 핵심 교훈: 학습 프로젝트에서는 artifact 분리보다 lesson 발전 과정이 더 먼저 보여야 한다
- 다음 작업으로 자연스럽게 이어지는 질문: lesson2 EventLoop 자산은 `lesson2/eventloop`, `lesson2/bossworker`, `support` 중 어디에 무엇을 둘 것인가
- 후속 작업 필요 여부: `Y`
- 최종 종료 시각: `2026-04-06 23:48 KST`
