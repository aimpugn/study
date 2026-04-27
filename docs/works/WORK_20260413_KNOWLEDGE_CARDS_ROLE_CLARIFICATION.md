# WORK_20260413_KNOWLEDGE_CARDS_ROLE_CLARIFICATION

## 0. Meta

- 작업 제목: knowledge cards 역할 재정의와 catalog 정합성 복구
- WORK 파일 경로: `WORK_20260413_KNOWLEDGE_CARDS_ROLE_CLARIFICATION.md`
- 저장소: `study`
- 작업 유형: `analysis | design | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청: "knowledge card는 ... 제거하는 게 오히려 ... 의견이 궁금합니다." 이후 "좋습니다... 진행해주세요."
- 원문 사용자 요청:
  - knowledge cards를 유지할지, 일반 디렉터리로 흡수할지 의견을 달라
  - 역할과 목적이 다르다면 그 기준으로 실제 정리해 달라
- 대상 경로 / 자산:
  - `USECASE.md`
  - `knowledge/KNOWLEDGE_TEMPLATE.md`
  - `knowledge/CATALOG.yaml`
- 실행자: Codex
- 시작 일시: `2026-04-13`
- 종료 일시: `2026-04-13`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal:
  - `knowledge/cards`를 "짧은 카드"가 아니라 "한 질문 또는 한 판단을 좁고 깊게 닫는 decision memo"로 재정의하고, 관련 문서와 catalog를 그 기준에 맞춘다.
- refs:
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
  - `PROJECT_INTENT.md`
  - `USECASE.md`
  - `knowledge/KNOWLEDGE_TEMPLATE.md`
  - `knowledge/CATALOG.yaml`
- scope:
  - 카드의 역할과 일반 장문 문서와의 차이 명문화
  - `knowledge/CATALOG.yaml`의 stale entry 정리
- mode:
  - `analysis + design + execute`
- run_mode:
  - `normal`
- finish:
  - `test+commit`
- must_keep:
  - cards 존치
  - 일반 장문 문서와 역할 분리
  - 과도한 문서 증식 금지
- extra_checks:
  - catalog path 정합성
  - 실제 카드 수와 catalog entry 수 정합성

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - cards를 유지할지 없앨지 판단 근거를 바탕으로 실제 정리
- 사용자가 명시한 금지 사항:
  - 없음
- path / naming / format / finish 관련 요구:
  - 현재 저장소 구조를 존중
  - 설명 중심 자산답게 문구를 자연스럽게 정리
- 내가 추가한 누락 방지 항목:
  - stale catalog entry 제거 또는 통합
  - cards의 목적이 "짧은 답변"으로 오해되지 않도록 문구 보강

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - 개별 카드 본문 전면 리라이트
  - 일반 디렉터리 문서 구조 대이동
  - knowledge 시스템 전면 폐기
- 지금 하지 않는 이유:
  - 이번 작업의 핵심은 역할 정의와 정합성 복구이지, 카드 컨텐츠 자체 재작성은 아님

## 2. Root-First Framing

- 근본 문제:
  - `knowledge/cards`가 일반 장문 문서와 어떤 점에서 다른지 문서상으로 충분히 닫혀 있지 않고, catalog도 실제 파일과 어긋나 드리프트 비용이 발생한다.
- 왜 이 문제가 지금 중요한가:
  - cards를 유지하더라도 역할 정의가 흐리면 다시 "중복 분기"처럼 보이게 되고, stale index가 남으면 신뢰가 무너진다.
- 작업 목표:
  - cards를 "좁고 깊은 판단 메모"로 고정하고, 일반 문서와의 차이를 분명히 하며, catalog를 실제 상태에 맞춘다.
- 기대 이점:
  - 자산 유형 간 역할 혼동 감소
  - cards 존치의 정당성 명문화
  - AI/사람 모두 카드 검색 표면을 더 신뢰할 수 있음
- 이점이 닫혔다고 판단할 확인 기준:
  - `USECASE.md`와 `KNOWLEDGE_TEMPLATE.md`에 역할 차이가 드러난다.
  - `CATALOG.yaml` entry가 실제 파일과 맞는다.
- 하드 제약 / 호환성 경계:
  - cards는 삭제하지 않는다.
  - 일반 문서를 cards 형식으로 강제하지 않는다.
- 성공 정의:
  - 역할 문구 정리 + catalog 정합성 PASS + commit
- PARTIAL 조건:
  - 역할 문구는 정리됐지만 catalog 정합성이 남아 있음
- BLOCKED 조건:
  - actual card/file mapping을 확정할 근거가 부족함

## 3. Reader & Internalization Contract

- 주 독자:
  - 미래의 저장소 소유자
  - 이 저장소를 읽는 AI
- 독자가 이미 알고 있다고 가정하는 것:
  - cards와 일반 문서가 모두 존재한다는 사실
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - 왜 cards를 별도 계층으로 유지하는가
  - cards와 일반 장문 문서의 역할 차이
  - 왜 catalog 정합성이 중요한가
- 사용자가 내재화해야 할 사고 패턴:
  - 길이보다 질문 범위와 판단 범위로 자산 유형을 나눈다
  - 인덱스는 실제 파일과 맞아야 한다
- 특히 막아야 하는 오해:
  - cards는 짧은 치트시트라는 오해
  - cards는 일반 장문 문서와 중복이라 지워야 한다는 오해
  - catalog 드리프트를 무시해도 된다는 오해
- 기억 anchor 후보:
  - `일반 문서 = 넓고 길게`
  - `cards = 좁고 깊게`
  - `CATALOG = 카드 검색 인덱스`
- 핵심 대조쌍 / 혼동쌍:
  - 장문 개념 문서 vs decision memo
  - 본문 자산 vs 중앙 인덱스

## 4. Project Overlay

- 적용한 로컬 AGENTS 경로:
  - `AGENTS.md`
- 활성화한 프로젝트 규칙:
  - exemplar보다 평균 품질을 따르지 않는다
  - 문서와 예제의 역할 차이를 분명히 한다
  - repo 변경 작업은 검수/검증/커밋으로 닫는다
- 특히 중요한 규칙:
  - 자산 유형별 역할 분리
  - explanation quality 유지
- 전역 규칙과의 충돌 여부 / 해소:
  - 충돌 없음

## 5. Topic Analysis

- 현재 이해한 사용자 의도:
  - cards가 정말 별도 계층으로 남을 가치가 있는지 판단했고, 가치가 있다면 그 이유를 문서와 구조에 반영하길 원한다.
- 현재 보이는 문제 구조:
  - `USECASE.md`는 cards를 "짧지만 강하게"라고 설명해 실제 자산 성격과 약간 어긋난다.
  - `KNOWLEDGE_TEMPLATE.md`는 card의 밀도와 완결성을 잘 설명하지만 일반 문서와의 관계는 더 분명히 할 여지가 있다.
  - `CATALOG.yaml`에는 실제 파일이 없는 stale entry가 2개 있다.
- 핵심 경계:
  - cards는 범용 주제 장문 문서가 아니라 재사용 가능한 판단 자산이어야 한다.
  - 일반 문서는 주제 자체를 넓게 배우는 장문 자산이다.
  - catalog는 knowledge 계층 내부 인덱스이며 실제 파일과 맞아야 한다.

## 6. Scope Expansion & Impact Sync

- 시작 키워드:
  - `knowledge/cards`, `decision memo`, `catalog`, `USECASE`
- 확장 키워드:
  - `짧지만 강하게`, `질문`, `판단`, `default_decision`, `problem_signature`
- 조사한 경로:
  - `USECASE.md`
  - `knowledge/KNOWLEDGE_TEMPLATE.md`
  - `knowledge/CATALOG.yaml`
  - `knowledge/cards/*.md`
- 함께 점검한 자산:
  - `AGENTS.md`
  - `PROJECT_INTENT.md`
- 한쪽만 바꾸면 깨질 부분:
  - usecase만 고치고 catalog를 안 고치면 실제 검색 표면은 여전히 불신 상태
  - catalog만 고치고 role 문구를 안 고치면 cards 존재 이유가 다시 흐려짐

## 7. Evidence Ledger

- E-01
  - 주장: 현재 cards는 짧은 메모가 아니라 밀도 높은 판단 문서다.
  - 근거 유형: `repo evidence`
  - 자료:
    - `knowledge/cards/*.md` line count 281~553
    - 카드들이 모두 `이 카드가 답하는 질문` 중심으로 시작
  - 이 자료로 닫힌 것:
    - cards는 짧은 치트시트와 다르다
  - 아직 비어 있는 것:
    - 문서상 역할 정의 보강
- E-02
  - 주장: 현재 `USECASE.md` 표현은 실제 card 성격과 완전히 일치하지 않는다.
  - 근거 유형: `repo evidence`
  - 자료:
    - `USECASE.md`의 `짧지만 강하게` 표현
    - 실제 card 분량과 구조
  - 이 자료로 닫힌 것:
    - usecase 문구 보정 필요성
  - 아직 비어 있는 것:
    - 보정 후 wording
- E-03
  - 주장: `CATALOG.yaml`에는 stale entry가 존재한다.
  - 근거 유형: `repo evidence`
  - 자료:
    - `K-DRY-WRONG-ABSTRACTION-AND-UTIL-PLACEMENT`
    - `K-HEXAGONAL-PORT-TYPES-AND-MAPPING-OWNERSHIP`
    - 실제 파일 부재
  - 이 자료로 닫힌 것:
    - catalog repair 필요성
  - 아직 비어 있는 것:
    - 최종 정리 방식

## 8. Design

- 선택한 접근:
  - cards는 유지하되, `좁고 깊은 판단 메모`로 역할을 명시하고, catalog는 실제 card 집합에 맞게 정리한다.
- 고려한 대안:
  - A. cards를 각 일반 디렉터리로 흡수
  - B. cards는 유지하되 role/cross-cutting 기준을 강화
- 대안을 채택하지 않은 이유:
  - A는 cross-cutting 판단 축을 여러 디렉터리로 흩어 재발견 비용을 키운다.
  - B가 역할 분리와 검색 재사용성 모두를 더 잘 지킨다.
- 문서 / 자산 구조:
  - `USECASE.md`: 자산 역할 정의 수정
  - `KNOWLEDGE_TEMPLATE.md`: card purpose/relationship 명시
  - `CATALOG.yaml`: 실제 card 집합 반영
- 설명 뼈대:
  - `비교형 | 질문형`
- 실패 모드:
  - cards가 다시 "짧은 문서"로 오해됨
  - catalog entry/file mismatch 재발

## 9. Frozen Checklist

- C-01: `USECASE.md`가 cards를 일반 문서와 구분되는 좁고 깊은 판단 메모로 설명한다.
- C-02: `KNOWLEDGE_TEMPLATE.md`가 cards의 목적과 일반 문서와의 관계를 더 분명히 말한다.
- C-03: `knowledge/CATALOG.yaml` entry가 실제 card 파일과 일치한다.
- C-04: 변경분이 검토되고 commit 된다.

## 10. Verification Plan

- `git diff -- <files>`로 내용 검토
- `ruby -e 'require "yaml"; ...'` 로 `knowledge/CATALOG.yaml` 파싱 확인
- catalog path와 실제 files 정합성 확인

### 10.1 Verification Result

- 실제 실행한 검증:
  - `git diff -- USECASE.md knowledge/KNOWLEDGE_TEMPLATE.md knowledge/CATALOG.yaml WORK_20260413_KNOWLEDGE_CARDS_ROLE_CLARIFICATION.md`
  - `git diff --stat -- USECASE.md knowledge/KNOWLEDGE_TEMPLATE.md knowledge/CATALOG.yaml WORK_20260413_KNOWLEDGE_CARDS_ROLE_CLARIFICATION.md`
  - `ruby -e 'require "yaml"; ...'`
  - `find knowledge/cards -maxdepth 1 -type f -name "*.md" | wc -l`
  - `git status --short -- USECASE.md knowledge/KNOWLEDGE_TEMPLATE.md knowledge/CATALOG.yaml WORK_20260413_KNOWLEDGE_CARDS_ROLE_CLARIFICATION.md`
- 결과:
  - `USECASE.md`는 cards를 길이 기준이 아니라 판단 범위 기준으로 설명하도록 보강됨
  - `KNOWLEDGE_TEMPLATE.md`는 cards와 일반 장문 문서의 관계를 분명히 설명함
  - `knowledge/CATALOG.yaml`는 `cards=7`, `missing=0` 확인
  - actual card file 수와 catalog entry 수가 모두 `7`로 일치함
- 검증 중 나온 이슈:
  - 첫 Ruby 검증 명령에서 `filter_map`이 현재 Ruby 버전에서 지원되지 않아 실패
- 이슈 처리:
  - 호환되는 `each` 기반 검증 명령으로 즉시 재실행했고, 구조 문제는 없음을 확인
- 실행하지 못한 검증과 이유:
  - 없음

## 11. Execution Log

- 실제 조사한 것:
  - `USECASE.md`의 cards 역할 문구
  - `knowledge/KNOWLEDGE_TEMPLATE.md`의 카드 목적과 품질 기준
  - `knowledge/CATALOG.yaml`의 entry와 실제 card file 정합성
  - `knowledge/cards/*.md`의 실제 질문 중심 구조와 분량
- 실제 수정한 것:
  - `USECASE.md`에서 cards를 좁고 깊은 판단 메모로 재정의
  - `knowledge/KNOWLEDGE_TEMPLATE.md`에서 cards와 일반 장문 문서의 관계 명시
  - `knowledge/CATALOG.yaml`에서 stale entry 제거 및 실제 card entry 반영
- 버린 접근과 이유:
  - cards 전체를 각 디렉터리로 분산: cross-cutting 판단 축 재발견 비용 증가
  - stale entry를 그대로 두고 role 문구만 수정: 검색 표면 신뢰 회복 실패

## 12. Final Audit & Closure

- 최종 상태:
  - `COMPLETE`
- 완료 게이트:
  - `ALLOW_COMPLETE`
- 체크리스트 재판정:
  - C-01: `PASS`
  - C-02: `PASS`
  - C-03: `PASS`
  - C-04: `PASS`
- 최종 감사:
  - cards를 삭제하지 않고도 역할 충돌을 줄이는 방향으로 정리되었고, stale catalog까지 같이 복구해 "정의만 있고 실행 표면은 틀린 상태"를 피했다.
- 커밋 해시 / 미커밋 사유:
  - 최종 repo commit 후 git history와 최종 보고에 기록한다.
