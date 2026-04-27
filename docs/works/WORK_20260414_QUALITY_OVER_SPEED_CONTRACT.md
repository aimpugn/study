# WORK_20260414_QUALITY_OVER_SPEED_CONTRACT

## 0. Meta

- 작업 제목: 품질 우선 / 속도 종속 계약 반영
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260414_QUALITY_OVER_SPEED_CONTRACT.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | execute | refactor_docs`
- 작업 깊이: `standard`
- 관련 요청:
  - `빠르게`는 원하는 바가 아니다
  - 하루가 걸려도 좋으니 가장 정확하고, 가장 올바르고, 가장 탁월하고, 가장 논리적이고, 가장 이성적이고, 가장 근거가 튼튼한 작업을 원한다
- 원문 사용자 요청:
  - 속도보다 품질을 우선하는 계약으로 이해하고 반영
- 대상 경로 / 자산:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
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
  - 속도는 부차 목표이고, 정확성·근거·논리·완결성이 최우선이라는 계약을 repo guidance와 study-explanation skill에 반영한다.
- refs:
  - 사용자 피드백
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- scope:
  - repo guidance update
  - WORK template update
  - external skill update
  - skill validation
  - repo commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - 사용자의 품질 우선 계약을 직접 문장으로 반영
  - 속도를 장점처럼 강조하지 않도록 guidance 수정

## 2. Root-First Framing

- 근본 문제:
  - 현재 규범은 이미 품질 중심이지만, 속도보다 탁월함을 더 명시적으로 우선한다는 문장이 약하다.
- 왜 이 문제가 지금 중요한가:
  - 집행자나 skill이 "빠르게"라는 말을 습관적으로 쓰면, 사용자가 원하는 품질 기준과 실제 작업 태도가 어긋날 수 있다.
- 작업 목표:
  - 품질 우선 / 속도 종속 원칙을 명시 규칙으로 승격한다.
- 성공 정의:
  - AGENTS, WORK template, skill에 같은 원칙이 반영되고 validator 확인까지 끝난다.

## 3. Frozen Checklist

- [x] AGENTS에 품질 우선 / 속도 종속 원칙을 명시한다
- [x] AGENTS에서 `빠르게`가 장점처럼 읽히는 문장을 수정한다
- [x] WORK template에 속도를 이유로 품질을 낮추지 않는 검토 지점을 추가한다
- [x] study-explanation skill에 같은 원칙을 반영한다
- [x] skill validation을 수행한다
- [x] repo 변경을 commit으로 닫는다

## 4. Execution Log

- repo guidance update:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - 속도는 1차 목표가 아니며, 품질이 떨어지는 빠른 경로는 채택하지 않는다는 원칙 추가
  - `knowledge/cards` 설명의 `빠르게` 표현을 `안정적으로 다시 참고`로 수정
- WORK template update:
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - 속도를 이유로 줄이면 안 되는 단계, 더 빠른 경로를 왜 채택하지 않았는지, 속도로 품질을 낮추지 않았는지 점검 항목 추가
- external skill update:
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - speed-only shortcut 금지, quality-over-speed 원칙 반영
- validation:
  - 임시 venv + `PyYAML`로 `quick_validate.py` 실행
  - 결과: `Skill is valid!`

## 5. Final Audit

- 상태 정직성:
  - COMPLETE 가능. guidance, template, skill, validation, commit 조건을 닫음
- 사용자 요구 반영:
  - 속도보다 정확성, 논리, 근거, 탁월함을 우선한다는 계약을 명시적으로 반영
- 남은 리스크:
  - repo 밖 skill 변경은 git commit에 포함되지 않으므로, 로컬 Codex 환경 자산으로 유지됨
