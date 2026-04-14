# WORK_20260414_STYLE_DRIFT_PREVENTION_AND_SOURCE_FLOW

## 0. Meta

- 작업 제목: 스타일 드리프트 방지와 principle-judgment-example 설명 규칙 리팩토링
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260414_STYLE_DRIFT_PREVENTION_AND_SOURCE_FLOW.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - `이제 JAR 안으로 올라가 보겠습니다` 같은 표현 제거
  - `manifest` 같은 용어도 쉬운 한국어로 같이 설명
  - `이 파일의 정체` 같은 표현 제거
  - 공식 문서를 먼저 말하고, 그 사실이 왜 필요와 구조로 이어지는지 자연스럽게 설명
  - 특정 문자열을 하드코딩해 걸러내는 방식이 맞는지 점검
  - 가능하면 `원칙 -> 판정 질문 -> 예시` 구조로 guidance / template / skill을 리팩토링
  - 이런 피드백이 순간 반영으로 끝나지 않게 guidance / template / skill에 고정
- 원문 사용자 요청:
  - 제 스타일에 맞게 설명을 작성하게 지침을 설정하고 템플릿을 구성하는 것이 어려운가에 대한 개선
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
  - style drift를 줄이되, 문장 수준 규칙이 특정 문자열 blacklist에 과하게 기대지 않도록 guidance / template / skill을 `원칙 -> 판정 질문 -> 예시` 구조로 다시 정리한다.
- refs:
  - 사용자 피드백
  - current guidance / template / skill / doc
  - already gathered official sources for JAR history/context
- scope:
  - repo guidance update
  - template update
  - current doc refactor
  - external skill update
  - skill validation
  - repo commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - 메타 진행 멘트보다 직접 진술형 문장을 우선한다는 원칙
  - source -> cause -> necessity -> structure 흐름을 판정 질문과 함께 고정
  - `manifest` 같은 load-bearing 일반명사를 쉬운 한국어로 먼저 풀어 쓰는 원칙
  - specific string는 calibration example로만 남기고 rule 본체로 두지 않음

## 2. Frozen Checklist

- [x] AGENTS에 `원칙 -> 판정 질문 -> 예시` 구조를 추가한다
- [x] WORK template을 blacklist 중심이 아니라 판정 질문 중심으로 바꾼다
- [x] study-explanation skill에 같은 구조를 반영한다
- [x] `spring_boot_jar_startup.md`의 문제 표현과 흐름을 수정한다
- [x] skill validation을 수행한다
- [x] repo scoped review와 commit으로 닫는다

## 3. Execution Log

- repo guidance update:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - 특정 문장을 blacklist 본체로 두지 않고 `원칙 -> 판정 질문 -> 예시` 순서로 재구성
  - `source -> cause -> necessity -> structure` 흐름을 판정 질문과 함께 설명 규칙으로 추가
  - `manifest`, `archive`, `classpath` 같은 일반명사를 쉬운 한국어로 먼저 풀어 쓰는 규칙을 판정 질문과 함께 고정
- WORK template update:
  - `설명 원칙`, `문장 단위 판정 질문`, `bad -> better 예시 페어` 항목 추가
  - 공식 자료 연결도 `어떤 인과 흐름인가`, `어디서 멈추면 FAIL인가`를 적게 변경
  - review 시 문자열 재등장보다 "문장이 서술자의 동선/감상만 말하는가"를 점검하도록 변경
- external skill update:
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - style rule을 blacklist가 아니라 principle / judgment / example 구조로 재정의
  - source-to-causality flow, plain-Korean noun definition rule을 판정 질문 중심으로 재구성
- current doc refactor:
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `이제 JAR 안으로 올라가 보겠습니다` 제거
  - `manifest`와 `archive`를 쉬운 한국어로 정의
  - `이 파일의 정체` 문장 제거
  - Oracle JAR overview 내용을 `왜 JAR가 생겼는가 -> 그래서 왜 manifest가 필요한가` 흐름으로 재작성
- validation:
  - 임시 venv + `PyYAML`로 `quick_validate.py` 실행
  - 결과: `Skill is valid!`

## 4. Final Audit

- 상태 정직성:
  - COMPLETE 가능. guidance, template, skill, doc, validation, commit까지 닫힘
- 왜 drift가 있었는가:
  - 기존 규칙이 있었지만, 일부 수정은 특정 문구를 막는 방향으로만 내려와 있었다
  - 그래서 비슷한 의미의 다른 문장이 다시 생기면 같은 문제를 반복할 위험이 있었다
- 이번 보강으로 달라진 점:
  - 규칙 본체가 `원칙 -> 판정 질문 -> 예시` 구조로 올라갔다
  - specific string는 calibration example로만 남고, 더 넓은 패턴을 같은 기준으로 잡을 수 있게 됐다
  - 공식 자료를 단순 인용이 아니라 인과 설명으로 잇는 규칙이 판정 질문 형태로 고정됐다
