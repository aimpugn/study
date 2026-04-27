# WORK_20260414_COMPONENT_TO_COMPOSITION_EXPLANATION_FEEDBACK

## 0. Meta

- 작업 제목: 구성요소별 설명 후 조합 메커니즘으로 올라가는 설명 계약 반영
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260414_COMPONENT_TO_COMPOSITION_EXPLANATION_FEEDBACK.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - 구조를 보여 준 뒤 감상형 문장으로 넘기지 말 것
  - 작은 단위를 하나씩 설명하고, 그 뒤에 그것들이 어떻게 엮여 큰 동작을 만드는지 설명할 것
  - 이 피드백을 guidance / template / skill에 일반화할 것
  - `spring_boot_jar_startup.md`를 리팩토링할 것
- 원문 사용자 요청:
  - 피드백 적용 + 일반화 + 문서 리팩토링
- 대상 경로 / 자산:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- 실행자: Codex
- 시작 일시: 2026-04-14
- 종료 일시: 2026-04-14
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - 구조 예시를 보여 준 뒤 `구성요소 분해 -> 각 요소 역할 -> 요소 간 연결 -> 상위 메커니즘` 순서로 설명하는 계약을 명시적으로 반영한다.
- refs:
  - 사용자 피드백
  - 현재 `spring_boot_jar_startup.md`
  - repo guidance / template
  - external skill files
- scope:
  - guidance update
  - template update
  - current doc refactor
  - external skill update
  - skill validation
  - repo commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - 작은 단위에서 큰 메커니즘으로 올라가는 설명 구조
  - 메타 감상문 제거
  - 실제 문서와 guidance 동시 반영

## 2. Root-First Framing

- 근본 문제:
  - 현재 문서는 실제 구조를 빨리 보여 주는 장점은 있지만, 그 구조를 본 직후 각 구성요소를 분해해 설명하고 다시 조합하는 단계가 약하다.
- 왜 이 문제가 지금 중요한가:
  - 사용자가 원하는 학습 방식은 작은 단위를 먼저 이해하고, 그 단위들이 엮여 큰 동작을 만드는 과정을 따라가는 방식이기 때문이다.
- 작업 목표:
  - 분해와 조합을 설명 계약으로 승격하고, 현재 문서에 즉시 반영한다.
- 성공 정의:
  - guidance, template, skill, current doc에 같은 원칙이 보이고 validator와 commit까지 닫힌다.

## 3. Frozen Checklist

- [x] AGENTS에 `구조 예시 -> 요소별 역할 -> 조합 메커니즘` 규칙을 추가한다
- [x] WORK template에 같은 planning / review hooks를 추가한다
- [x] `spring_boot_jar_startup.md` 앞부분을 구성요소 중심으로 재구성한다
- [x] `spring_boot_jar_startup.md`에서 각 요소가 어떻게 엮여 실행 경로가 되는지 직접 설명한다
- [x] external skill에 같은 규칙을 반영한다
- [x] skill validation을 수행한다
- [x] repo scoped review와 commit으로 닫는다

## 4. Execution Log

- repo guidance update:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - 구조 예시를 보여 준 뒤 감상형 문장으로 넘기지 않고, 구성요소별 설명 후 상위 메커니즘으로 조립하는 규칙 추가
- WORK template update:
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - 먼저 분해해서 설명할 구성요소, 상위 메커니즘 조립 경로, 관련 review 항목 추가
- current doc refactor:
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `예를 들어` 도입으로 직접 진입형 표현 사용
  - manifest, loader, app class, library jar를 각각 설명
  - 그 요소들이 실제 실행 경로를 어떻게 만드는지 순서대로 설명
- external skill update:
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - 구조 예시 뒤 `구성요소 -> 역할 -> 연결 -> 상위 메커니즘` 규칙 반영
- validation:
  - 임시 venv + `PyYAML`로 `quick_validate.py` 실행
  - 결과: `Skill is valid!`

## 5. Final Audit

- 상태 정직성:
  - COMPLETE 가능. guidance, template, doc, external skill, validation, commit 조건을 닫음
- 사용자 요구 반영:
  - 구조 예시 뒤 감상형 문장을 제거하고, 각 요소 설명 후 조합 메커니즘으로 올라가는 방식으로 반영
- 남은 리스크:
  - 같은 피드백을 다른 기존 문서들에도 순차적으로 적용할 필요가 있음
