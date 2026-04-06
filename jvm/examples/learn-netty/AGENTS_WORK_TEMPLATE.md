# WORK_YYYYMMDD_TASK_SUMMARY

> 이 템플릿으로 만든 개별 작업 문서는 반드시 `docs/works/YYYYMMDD_${작업요약}.md` 형식으로 저장합니다.
> 이 템플릿의 목적은 메모를 남기는 것이 아니라, 작업의 개념적 완결성, 작업 과정의 완결성, 결과의 완결성을 fail-closed 방식으로 강제하는 것입니다.
> 이 템플릿은 매 작업마다 최소한 다음 문서들을 실제로 읽고 적용했다는 기록을 남겨야 합니다.
> - 전역 `~/.codex/AGENTS.md`
> - 프로젝트 [`PROJECT_INTENT.md`](/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/PROJECT_INTENT.md)
> - 프로젝트 [`AGENTS.md`](/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS.md)
> - 현재 작업 템플릿 [`AGENTS_WORK_TEMPLATE.md`](/Users/rody/VscodeProjects/study/jvm/examples/learn-netty/AGENTS_WORK_TEMPLATE.md)
> 섹션은 임의로 삭제하지 않습니다. 정말 해당 없음이면 `N/A`와 이유를 적습니다.
> 한 Unit은 해당 Unit에 연결된 체크리스트와 검증이 닫히기 전에는 다음 Unit으로 넘어가지 않습니다.

## 0. Meta

- 작업 제목:
- WORK 파일 경로:
- 파일명 규칙 확인: `docs/works/YYYYMMDD_${작업요약}.md`
- 저장소 / 모듈:
- 작업 유형: `research | analysis | design | explain | execute | refactor | audit | mixed`
- 작업 깊이: `minimal | standard | full`
- 실행자:
- 시작 일시:
- 종료 일시:
- 관련 요청 / 이슈:
- 원문 사용자 요청:
- run_mode: `normal | dry-run`
- finish: `report | verify | test | test+commit | test+commit+push`
- repo 변경 여부: `Y | N`
- commit 기본 요구 여부: `Y | N`
- 현재 상태: `IN_PROGRESS | PARTIAL | BLOCKED | COMPLETE`
- 완료 게이트: `PENDING | ALLOW_COMPLETE | BLOCK_COMPLETE`

## 1. Instruction Stack Preflight

### 1.1 Resolved Sources

- 전역 `AGENTS.md` 경로:
- 전역 `AGENTS.md` 읽음: `Y | N`
- 전역 `AGENTS.md` 적용 요약:
- `PROJECT_INTENT.md` 경로:
- `PROJECT_INTENT.md` 읽음: `Y | N`
- `PROJECT_INTENT.md` 적용 요약:
- 로컬 `AGENTS.md` 경로:
- 로컬 `AGENTS.md` 읽음: `Y | N`
- 로컬 `AGENTS.md` 적용 요약:
- 이 템플릿 경로:
- 이 템플릿 사용 시작 시각:

### 1.2 Effective Contract

- 이번 작업에서 특히 강한 hard rule:
- 충돌 가능성:
- 충돌 해소 방식:
- preflight blocker:

### 1.3 Completion Formula

- `INSTRUCTION_STACK_READY = global AGENTS read+applied AND PROJECT_INTENT read+applied AND local AGENTS read+applied AND this template instantiated`
- `ALLOW_COMPLETE = INSTRUCTION_STACK_READY AND frozen checklist all PASS AND required verification all PASS AND PROJECT_INTENT compliance PASS AND global/local AGENTS compliance PASS AND final audit PASS AND unresolved blocker = 0 AND (repo change -> commit recorded when commit is in scope)`
- 하나라도 FAIL 또는 미판정이면 `BLOCK_COMPLETE`

## 2. Request Normalization

### 2.1 Intent

- goal:
- root request in one sentence:
- 왜 이렇게 해석했는가:
- 다른 해석 가능성:
- 최종 선택 해석:

### 2.2 Explicit Deliverables

- 사용자가 명시한 필수 산출물:
- 사용자가 명시한 금지 사항:
- 사용자가 명시한 path / naming / format / finish 요구:
- 사용자가 명시하지 않았지만 누락 방지를 위해 추가한 필수 항목:

### 2.3 Non-Goals

- 이번 작업의 비범위:
- 지금 하지 않는 이유:
- 나중으로 미루면 안 되는 항목과 그 이유:

## 3. Root Problem & Benefit

- 근본 문제:
- 왜 지금 중요한가:
- 의도된 이점:
- 이점이 실제로 닫혔다고 판단할 신호:
- 핵심 불변식:
- 하드 제약:
- 잘못하면 생기는 열화 / 왜곡:
- COMPLETE 정의:
- PARTIAL 정의:
- BLOCKED 정의:

## 4. Topic Analysis

- 현재 이해한 문제 구조:
- 이번 작업이 건드리는 표면:
- 숨은 가정:
- 핵심 미지수:
- 처음 떠오른 접근:
- 성공을 오판하기 쉬운 지점:

## 5. Analysis Critique + Repair

### 5.1 Challenge 1

- 무엇이 틀릴 수 있는가:
- repair:
- 왜 repair가 더 강한가:

### 5.2 Challenge 2

- 무엇이 좁거나 누락될 수 있는가:
- repair:
- 왜 repair가 더 강한가:

### 5.3 Challenge 3

- 무엇이 학습/설명 목표를 놓칠 수 있는가:
- repair:
- 왜 repair가 더 강한가:

### 5.4 Retained Framing

- 최종 채택 분석:
- 폐기한 분석과 이유:

## 6. Research / Evidence Plan

- 먼저 확인해야 하는 질문:
- 초기 검색 키워드:
- 확장 키워드:
- 조사할 repo 경로:
- 조사할 1차 자료:
- 기대하는 증거:
- research stop condition:

## 7. Evidence Ledger

- E-01
  - 질문 / 주장:
  - 근거 유형: `repo evidence | experiment | command result | official doc | standard | source code | inference`
  - 자료:
  - 이 자료가 닫아 준 것:
  - 아직 비어 있는 것:
- E-02
  - 질문 / 주장:
  - 근거 유형:
  - 자료:
  - 이 자료가 닫아 준 것:
  - 아직 비어 있는 것:
- E-03
  - 질문 / 주장:
  - 근거 유형:
  - 자료:
  - 이 자료가 닫아 준 것:
  - 아직 비어 있는 것:
- E-04
  - 질문 / 주장:
  - 근거 유형:
  - 자료:
  - 이 자료가 닫아 준 것:
  - 아직 비어 있는 것:

## 8. Evidence Critique + Repair

- 약한 근거:
- 충돌하는 근거:
- 오래되었을 가능성이 있는 가정:
- 추가 수집 필요 항목:
- confidence를 낮춰야 하는 항목:
- repair 후 최종 근거 세트:

## 9. PROJECT_INTENT Fit & Learning Design

- 이번 작업이 직접 지원하는 `PROJECT_INTENT` 항목:
- 이번 작업에서 반드시 드러내야 하는 **보이지 않는 동작**:
- 독자가 나중에 스스로 설명할 수 있어야 하는 질문:
- 이번 작업의 핵심 학습 대상:
- 이번 작업의 핵심 혼동쌍:

### 9.1 Observability Design

- ASCII 다이어그램 필요 여부 / 이유:
- 로그 또는 타임라인 필요 여부 / 이유:
- 시각화 페이지 필요 여부 / 이유:
- 실패 실험 필요 여부 / 이유:
- 비교 예제 필요 여부 / 이유:
- 부분 재구현 필요 여부 / 이유:
- 채택한 관측 장치:
- 채택하지 않은 관측 장치와 이유:

### 9.2 Explanation Design

- 눈에 바로 들어오게 보여 줄 구조:
- 읽으면서 논리가 쌓이게 할 핵심 서술:
- worked example 후보:
- 회상 anchor:
- 코드 스니펫 / 파일 링크로 직접 보여 줄 위치:

## 10. Scope Expansion & Impact Sync

### 10.1 Included Scope

- 코드:
- 테스트:
- 문서:
- 예제:
- 실험:
- 로그 / 관측 장치:
- 프론트 / 시각화:
- 설정 / 스크립트:

### 10.2 Excluded Scope

- 제외 항목:
- 제외 근거:
- 왜 지금 제외해도 개념적 완결성이 깨지지 않는가:

### 10.3 Search Expansion Ledger

- 핵심 키워드:
- 동의어 / 약어 / 구명칭:
- 관련 에러 / 로그 / API 이름:
- 실제 검색 경로:
- 누락 가능성 점검 결과:

### 10.4 Impact Sync Map

- 함께 움직여야 하는 자산:
- 한쪽만 바꾸면 깨질 부분:
- 관련 문서 동기화 항목:
- 관련 예제 동기화 항목:
- 관련 관측 장치 동기화 항목:

## 11. Design

- 선택한 접근:
- 왜 이것이 근본 문제와 이점에 맞는가:
- 고려한 대안 A:
- 대안 A를 채택하지 않은 이유:
- 고려한 대안 B:
- 대안 B를 채택하지 않은 이유:
- 주요 계약 / 경계:
- 실패 모드:
- verification path:

## 12. Design Critique + Repair

### 12.1 Correctness View

- 반론:
- repair 또는 유지:
- 이유:

### 12.2 Learner View

- 반론:
- repair 또는 유지:
- 이유:

### 12.3 Operations / Performance View

- 반론:
- repair 또는 유지:
- 이유:

### 12.4 Final Design Decision

- 최종 채택안:
- 트레이드오프:

## 13. Overall Plan

- 작업 순서:
- 선행 의존성:
- validation order:
- retry order:
- rollback / reopen 기준:

## 14. Plan Critique + Repair

- 이 계획이 실패할 수 있는 지점:
- 순서상 위험:
- 빠진 prerequisite:
- repair:
- 왜 repair된 계획이 더 강한가:

## 15. Detailed Task Plan

- Unit-01
  - 목적:
  - 대상 파일 / 자산:
  - 바꿀 논리:
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것:
  - 검증:
  - 완료 판정:
- Unit-02
  - 목적:
  - 대상 파일 / 자산:
  - 바꿀 논리:
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것:
  - 검증:
  - 완료 판정:
- Unit-03
  - 목적:
  - 대상 파일 / 자산:
  - 바꿀 논리:
  - 설명 / 로그 / 시각화 / 실험 중 포함할 것:
  - 검증:
  - 완료 판정:

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1:
  - S2:
  - S3:
- 실패 케이스 최소 3개:
  - F1:
  - F2:
  - F3:
- 회귀 위험:
- 회귀 방지 확인 경로:

## 16. Detailed Plan Critique + Repair

- 빠진 단위:
- fuzzy success criteria:
- 과한 범위 / 부족한 범위:
- repair:
- 최종 상세 계획:

## 17. Frozen Success / Failure Checklist

> 사용자 항목은 최소 바닥선이다. 삭제·완화 금지.
> 모든 required 항목이 PASS여야만 COMPLETE 가능.

- C-01
  - 출처: `사용자 | AI-추가`
  - 내용:
  - required: `Y | N`
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
  - 재시도 트리거:
  - 관련 Unit:
- C-02
  - 출처:
  - 내용:
  - required: `Y | N`
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
  - 재시도 트리거:
  - 관련 Unit:
- C-03
  - 출처:
  - 내용:
  - required: `Y | N`
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
  - 재시도 트리거:
  - 관련 Unit:
- C-04
  - 출처:
  - 내용:
  - required: `Y | N`
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
  - 재시도 트리거:
  - 관련 Unit:
- C-05
  - 출처:
  - 내용:
  - required: `Y | N`
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
  - 재시도 트리거:
  - 관련 Unit:

### 17.1 Checklist Critique + Repair

- 각 항목이 목표 / 이점 / 사용자 요구와 매핑되는가:
- PASS / FAIL이 관측 가능한가:
- 증거가 실제로 존재 가능한가:
- 사용자 요구를 조용히 약화한 항목이 없는가:
- repair:
- Freeze 시각:
- Freeze 버전:

## 18. Execution Ledger

> 현재 Unit이 FAIL 또는 PARTIAL이면, root cause를 고치고 관련 phase를 다시 연 뒤에만 다음 Unit으로 진행합니다.

- Attempt-01 / Unit-01
  - 시작 시각:
  - 실행 내용:
  - 변경 파일:
  - 새로 생긴 증거:
  - 수행한 검증:
  - 결과: `PASS | FAIL | PARTIAL`
  - 실패 시 근본 원인:
  - 실패 시 earliest affected phase:
  - 다음 조치:
- Attempt-02 / Unit-02
  - 시작 시각:
  - 실행 내용:
  - 변경 파일:
  - 새로 생긴 증거:
  - 수행한 검증:
  - 결과: `PASS | FAIL | PARTIAL`
  - 실패 시 근본 원인:
  - 실패 시 earliest affected phase:
  - 다음 조치:
- Attempt-03 / Unit-03
  - 시작 시각:
  - 실행 내용:
  - 변경 파일:
  - 새로 생긴 증거:
  - 수행한 검증:
  - 결과: `PASS | FAIL | PARTIAL`
  - 실패 시 근본 원인:
  - 실패 시 earliest affected phase:
  - 다음 조치:

## 19. Retry / Re-entry Ledger

- R-01
  - trigger:
  - 왜 COMPLETE가 아니게 되었는가:
  - earliest affected phase:
  - root cause repair:
  - 다시 수행한 phase:
  - 재검증 결과:
- R-02
  - trigger:
  - 왜 COMPLETE가 아니게 되었는가:
  - earliest affected phase:
  - root cause repair:
  - 다시 수행한 phase:
  - 재검증 결과:

## 20. Frozen Checklist Re-Judgement

- C-01 Final: `PASS | FAIL | N/A`
  - evidence:
  - notes:
- C-02 Final: `PASS | FAIL | N/A`
  - evidence:
  - notes:
- C-03 Final: `PASS | FAIL | N/A`
  - evidence:
  - notes:
- C-04 Final: `PASS | FAIL | N/A`
  - evidence:
  - notes:
- C-05 Final: `PASS | FAIL | N/A`
  - evidence:
  - notes:

## 21. Compliance Audit

### 21.1 PROJECT_INTENT Compliance

- 이번 작업이 `PROJECT_INTENT`의 어떤 목표를 충족했는가:
- `PROJECT_INTENT`와 어긋난 부분:
- 어긋났다면 repair:
- Final: `PASS | FAIL`

### 21.2 Global AGENTS Compliance

- 전역 hard rule 중 이번 작업에 적용된 것:
- 누락 또는 위반 가능성:
- repair:
- Final: `PASS | FAIL`

### 21.3 Local AGENTS Compliance

- 로컬 hard rule 중 이번 작업에 적용된 것:
- 관측 장치 / 설명 품질 / 자연스러운 한국어 / 코드 링크 규칙 준수 여부:
- 누락 또는 위반 가능성:
- repair:
- Final: `PASS | FAIL`

## 22. Final Audit

### 22.1 Expert Review

- 목적 적합성:
- 개념적 완결성:
- 작업 완결성:
- 결과 완결성:
- 설명 품질:
- 검증 정직성:
- 남은 이상 징후:

### 22.2 Final Audit Repair

- audit에서 발견된 문제:
- repair:
- repair 후 재검토 결과:

## 23. Final Verification Summary

- 실제 실행한 명령 / 테스트 / 확인:
- PASS 신호:
- FAIL 신호:
- 실행하지 못한 검증:
- 실행하지 못한 이유:
- 최종 confidence:

## 24. Completion Decision

- 최종 상태: `COMPLETE | PARTIAL | BLOCKED`
- 완료 게이트: `ALLOW_COMPLETE | BLOCK_COMPLETE`
- 열린 게이트:
- 미해소 blocker:
- COMPLETE라고 판단하는 근거:
- PARTIAL / BLOCKED라면 승격 조건:

## 25. Commit Closure

- repo 변경 여부: `Y | N`
- commit 요구 여부: `Y | N`
- commit 수행 여부: `Y | N`
- commit SHA:
- commit message:
- 이 작업을 닫는 커밋인지 여부: `Y | N`
- push 여부: `Y | N`
- push는 왜 했거나 하지 않았는가:

## 26. Closing Notes

- 이번 작업의 핵심 교훈:
- 다음 작업으로 자연스럽게 이어지는 질문:
- 후속 작업 필요 여부:
- 최종 종료 시각:
