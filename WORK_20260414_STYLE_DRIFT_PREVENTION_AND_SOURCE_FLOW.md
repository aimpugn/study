# WORK_20260414_STYLE_DRIFT_PREVENTION_AND_SOURCE_FLOW

## 0. Meta

- 작업 제목: 스타일 드리프트 방지와 source-to-causality 설명 흐름 강화
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260414_STYLE_DRIFT_PREVENTION_AND_SOURCE_FLOW.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - `이제 JAR 안으로 올라가 보겠습니다` 같은 표현 제거
  - `manifest` 같은 용어도 쉬운 한국어로 같이 설명
  - `이 파일의 정체` 같은 표현 제거
  - 공식 문서를 먼저 말하고, 그 사실이 왜 필요와 구조로 이어지는지 자연스럽게 설명
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
  - style drift를 줄이기 위해 문장 수준 anti-pattern과 preferred flow를 guidance와 skill에 명시하고, current doc에 즉시 반영한다.
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
  - vague stage-direction phrasing ban
  - source -> cause -> necessity -> structure flow
  - plain Korean definition for general nouns like manifest

## 2. Frozen Checklist

- [x] AGENTS에 sentence-level anti-pattern과 preferred flow를 추가한다
- [x] WORK template에 피해야 하는 표현 / 공식 자료 설명 흐름 hooks를 추가한다
- [x] study-explanation skill에 같은 규칙을 반영한다
- [x] `spring_boot_jar_startup.md`의 문제 표현과 흐름을 수정한다
- [x] skill validation을 수행한다
- [x] repo scoped review와 commit으로 닫는다

## 3. Execution Log

- repo guidance update:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `이제 X 안으로 올라가 보겠습니다`, `이 파일의 정체`, `왜 이런 파일이 필요했는지도 같이 봐야 합니다` 같은 문장을 anti-pattern으로 명시
  - `source -> cause -> necessity -> structure` 흐름을 설명 규칙으로 추가
  - `manifest`, `archive`, `classpath` 같은 일반명사를 쉬운 한국어로 먼저 풀어 쓰는 규칙 추가
- WORK template update:
  - 피해야 하는 표현 / 메타 진행 멘트
  - 공식 자료를 어떤 인과 흐름으로 설명할지
  - review 시 anti-pattern 재등장 여부 점검 항목 추가
- external skill update:
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - stage-direction phrasing ban, source-to-causality flow, plain-Korean noun definition rule 추가
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
  - 기존 규칙이 원칙 수준에는 충분했지만, sentence-level anti-pattern과 preferred rewrite flow가 부족해서 매 라운드마다 일부 표현이 되살아났다
- 이번 보강으로 달라진 점:
  - 금지 표현과 선호 흐름이 문장 수준까지 내려왔다
  - 공식 자료를 단순 인용이 아니라 인과 설명으로 잇는 규칙이 생겼다
  - 이후 similar doc review에서 같은 drift를 더 쉽게 잡을 수 있다
