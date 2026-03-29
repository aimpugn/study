# WORK_YYYYMMDD_TITLE

> study 저장소용 기본 WORK 템플릿입니다.
> SSOT는 항상 [`AGENTS.md`](/Users/rody/VscodeProjects/study/AGENTS.md)와 전역 `~/.codex/AGENTS.md`입니다.
> 이 템플릿은 특히 Markdown 문서 작성, 정리, 검증, 리팩토링 작업에 맞춰져 있습니다.
> 작업이 코드 구현까지 크게 확장되면, 필요한 섹션을 전역 `AGENTS_WORK_TEMPLATE.md` 수준으로 추가해서 사용합니다.

## 0. Meta

- 작업 제목:
- WORK 파일 경로:
- 저장소:
- 작업 유형: `new_doc | refine_doc | fact_check | restructure_docs | add_examples | explain_only`
- 대상 경로:
- 관련 요청 / 질문:
- 원문 사용자 요청:
- 실행자:
- 시작 일시:
- 종료 일시:
- 현재 상태: `IN_PROGRESS | PARTIAL | BLOCKED | COMPLETE`
- 완료 게이트: `PENDING | ALLOW_COMPLETE | BLOCK_COMPLETE`
- finish 조건: `report | verify | verify+commit`

## 1. Request Normalization

- 이 작업을 한 문장으로 다시 쓰면:
- 사용자가 진짜로 원하는 결과물:
- 사용자가 직접 준 필수 요구:
- 사용자가 직접 준 금지 사항:
- 내가 추가한 누락 방지 항목:
- 지금 수정해야 하는가, 새로 만들어야 하는가:
- 새 파일이라면 왜 기존 문서 보강이 아닌가:

## 2. Project Overlay

- 적용한 로컬 AGENTS 경로:
- 로컬 규칙을 읽고 적용했는가: `Y | N`
- 이번 작업에서 특히 중요한 규칙:
- 전역 규칙과 충돌 가능성 / 해소 방식:

### 2.1 Project Rule Ledger

- STUDY-01 미래의 내가 다시 이해할 수 있는 설명:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-02 이해와 실전을 함께 남기기:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-03 근거와 추론 분리:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-04 기존 문서와의 관계 선확인:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-05 문서 유형에 맞는 구조:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-06 사용자 이해 기준과 톤 보존:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-07 관련 문서와 탐색 경로 동기화:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`
- STUDY-08 문서 작업 검증과 솔직한 마감:
  - Activated: `Y | N`
  - 적용 이유:
  - 검증 방법:
  - Final: `PASS | FAIL | N/A`

## 3. Reader & Learning Contract

- 이 문서의 주 독자:
- 독자가 이미 알고 있다고 가정하는 것:
- 이 문서를 읽고 나면 바로 할 수 있어야 하는 것:
- 시간이 지나도 다시 떠올릴 수 있게 꼭 남겨야 하는 것:
- 특히 막아야 하는 오해 / 혼동:
- 설명 톤 메모:

## 4. Scope & Existing Material Survey

- 대상 문서 / 디렉터리:
- 함께 확인한 기존 문서:
- 검색한 키워드:
- 동의어 / 원어 / 약어 / 구표현:
- 중복 가능성 점검 결과:
- 이번 작업에서 함께 맞춰야 하는 관련 문서:
- 제외한 범위와 이유:

## 5. Doc Archetype

- 문서 유형: `concept | command | troubleshooting | standard | knowledge_card | mixed`
- 이 유형을 선택한 이유:
- 이 문서에서 반드시 닫아야 하는 제어점 2~4개:
  - C1:
  - C2:
  - C3:
  - C4:

### 5.1 이 유형에서 필요한 최소 요소

- `concept`: 질문, 왜 중요한가, 핵심 메커니즘, 예제, 흔한 오해, 확인 방법, 근거
- `command`: 무엇을 하는가, 입력/옵션 감각, 안전한 예제, 실무 예제, 실패/주의점, 확인 방법, 근거
- `troubleshooting`: 증상, 실제 원인, 왜 그런가, 해결, 재발 방지 또는 확인, 근거
- `standard`: 규칙 범위, 원 규칙의 의미, 구현 해석 포인트, 예시, 주의점, 근거
- `knowledge_card`: [`knowledge/KNOWLEDGE_TEMPLATE.md`](/Users/rody/VscodeProjects/study/knowledge/KNOWLEDGE_TEMPLATE.md) 우선

## 6. Evidence Ledger

- 핵심 주장 1:
  - 주장:
  - 근거 종류: `공식 문서 | 표준 | repo 문서 | 실험 | 추론`
  - 자료:
  - 이 자료로 확인한 범위:
  - 아직 비어 있는 것:
- 핵심 주장 2:
  - 주장:
  - 근거 종류: `공식 문서 | 표준 | repo 문서 | 실험 | 추론`
  - 자료:
  - 이 자료로 확인한 범위:
  - 아직 비어 있는 것:
- 핵심 주장 3:
  - 주장:
  - 근거 종류: `공식 문서 | 표준 | repo 문서 | 실험 | 추론`
  - 자료:
  - 이 자료로 확인한 범위:
  - 아직 비어 있는 것:

### 6.1 추론 / 미확인 항목

- 추론으로만 남는 내용:
- 왜 아직 직접 확인하지 못했는가:
- 최종 문서에서 어떻게 표시할 것인가:

## 7. Frozen Checklist

> 사용자 항목은 최소 바닥선입니다. 삭제하거나 완화하지 않습니다.

- [ ] 이 문서가 답하는 질문이 초반에 드러난다.
- [ ] 왜 중요한지 또는 어디서 헷갈리는지 설명한다.
- [ ] 핵심 메커니즘이 닫혀 있다.
- [ ] 손에 잡히는 예제나 사용 예가 있다.
- [ ] 실패 모드 / 흔한 오해 / 주의점 중 해당하는 것을 담는다.
- [ ] 확인 방법, 최소 실험, 또는 검증 경로가 있다.
- [ ] 근거 / 참고 자료가 남아 있다.
- [ ] 기존 관련 문서와의 역할 관계가 정리되어 있다.
- [ ] 실행하지 않은 내용은 실행한 것처럼 쓰지 않는다.
- [ ] 사용자 요청의 명시 항목이 빠지지 않았다.

### 7.1 Task-Specific Checklist

- C-01:
  - 출처: `사용자 | AI-추가`
  - 내용:
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
- C-02:
  - 출처: `사용자 | AI-추가`
  - 내용:
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:
- C-03:
  - 출처: `사용자 | AI-추가`
  - 내용:
  - PASS 기준:
  - FAIL 기준:
  - 필요한 증거:

### 7.2 Freeze

- Freeze 시각:
- Freeze 버전:
- Freeze 이후 추가된 항목과 이유:

## 8. Structure Plan

- 문서 첫 문장 / 첫 문단이 답할 것:
- 메인 섹션 흐름:
- 넣을 예제:
- 넣을 반례 / 오해 / 주의점:
- 넣을 확인 방법:
- 근거 / 참고 자료 섹션 구성:
- 관련 문서 링크 계획:

## 9. Execution Log

- 조사한 것:
- 실제로 수정한 것:
- 버린 접근과 이유:
- 사용자의 이해 기준에 맞추기 위해 조정한 점:

## 10. Verification & Final Audit

- 수행한 검증 명령:
- 검증 결과:
- 직접 실행/확인한 예제 범위:
- 실행하지 못한 것과 이유:
- 관련 문서 동기화 확인:
- 남은 불확실성:

### 10.1 Checklist Re-Judgement

- C-01: `PASS | FAIL | N/A`
- C-02: `PASS | FAIL | N/A`
- C-03: `PASS | FAIL | N/A`
- 공통 체크리스트 판정:

### 10.2 Final Status

- 최종 상태: `COMPLETE | PARTIAL | BLOCKED`
- 완료 게이트: `ALLOW_COMPLETE | BLOCK_COMPLETE`
- COMPLETE가 아니라면 남은 blocker:
- 다음 재진입 조건:
- 커밋 해시 / 미커밋 사유:
