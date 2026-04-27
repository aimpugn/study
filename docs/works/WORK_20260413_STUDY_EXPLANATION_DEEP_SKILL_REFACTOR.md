# WORK_20260413_STUDY_EXPLANATION_DEEP_SKILL_REFACTOR

## 0. Meta

- 작업 제목: study-explanation 스킬을 deep study monograph 지향으로 리팩터링
- WORK 파일 경로: `WORK_20260413_STUDY_EXPLANATION_DEEP_SKILL_REFACTOR.md`
- 저장소: `study`
- 작업 유형: `analysis | design | execute`
- 작업 깊이: `full`
- 관련 요청:
  - study-explanation에 대해 기대하는 것은 아주 깊고, 전문적이며, 뿌리부터 하나씩 파고들고 작은 예제로 확장해 나가는 학습 자료를 만드는 것
  - 목적과 바라는 방향에 맞게 스킬을 리팩토링해 달라
- 원문 사용자 요청:
  - exemplar 파일들을 기준으로 skill을 고도화
  - `.tmp/interview*.md`, `interviews/*` 같은 대형 원재료도 앞으로 정리 대상이 될 수 있음을 반영
- 대상 경로 / 자산:
  - `AGENTS.md`
  - external skill: `~/.codex/skills/study-explanation/*`
  - exemplar refs:
    - `computer_architecture/ostep/xv6-riscv/kernel/entry.S`
    - `algorithms/dynamic_programming.md`
    - `algorithms/optimal_solution.md`
    - `git/git_rebase.md`
    - `computer_architecture/threads/threads.md`
    - `jvm/java/java_synchronized.md`
- 실행자: Codex
- 시작 일시: `2026-04-13`
- 종료 일시: `2026-04-13`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal:
  - `study-explanation`의 기본 산출물을 얕은 요약이 아니라 `deep study monograph`로 재정의하고, 그 기준을 repo SSOT와 external skill에 함께 반영한다.
- refs:
  - `AGENTS.md`
  - `AGENTS_WORK_TEMPLATE.md`
  - `PROJECT_INTENT.md`
  - `USECASE.md`
  - current `~/.codex/skills/study-explanation/SKILL.md`
  - user-selected exemplar files
- scope:
  - repo 설명 규칙 보강
  - skill workflow / metadata / references 리팩터링
- mode:
  - `analysis + design + execute`
- run_mode:
  - `normal`
- finish:
  - `test+commit`
- must_keep:
  - repo AGENTS가 SSOT
  - skill은 집행기
  - cards는 기본 출력이 아님
  - 깊고, 작은 예제로 시작하고, 계층적으로 확장하는 설명
- extra_checks:
  - exemplar file 기반 규칙 추출
  - `.tmp/interview*.md` / `interviews/*`는 원재료로 취급하는 규칙
  - skill validation

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - 깊고 전문적이며, 뿌리부터 단계적으로 이해되게 만드는 skill
  - 작은 예제로 동작을 먼저 만들고 큰 것으로 확장하는 설명
  - 맥락과 역사, 왜/어떻게를 논리적으로 잇는 설명
  - study 목적에 맞는 방향으로 skill 리팩터링
- 사용자가 명시한 금지 사항:
  - 없음
- path / naming / format / finish 관련 요구:
  - `study-explanation` skill을 계속 개선 가능한 형태로 유지
  - repo 변경은 commit으로 닫기
- 내가 추가한 누락 방지 항목:
  - AGENTS에 deep study 계약 반영
  - exemplar notes를 skill references로 분리
  - openai metadata 갱신
  - validator 통과

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - exemplar 문서 본문 전체 리라이트
  - `.tmp/interview*.md` 실질 정리 작업
  - 전역 CS 용어집 작성
- 지금 하지 않는 이유:
  - 이번 작업의 핵심은 앞으로의 설명 품질을 만드는 기준과 집행기 리팩터링이다.

## 2. Root-First Framing

- 근본 문제:
  - 현재 `study-explanation`은 일반 설명 품질 가이드로는 충분하지만, 사용자가 원하는 "뿌리부터 쌓는 깊은 학습 자료"를 기본 출력 계약으로 강하게 고정하지 못한다.
- 왜 이 문제가 지금 중요한가:
  - skill이 기본적으로 얕은 설명이나 일반 정리로 미끄러지면 저장소의 핵심 목적(study, 체화, 복기 가능성)을 달성하기 어렵다.
- 작업 목표:
  - skill의 기본 산출물을 `deep study monograph`로 바꾸고, repo SSOT에도 그 기준을 명시한다.
- 기대 이점:
  - 다양한 주제에서도 일관되게 깊은 학습 자료 생성 가능
  - 작은 예제 -> 메커니즘 -> 확장 구조의 재사용
  - raw material와 exemplar의 구분 강화
- 이점이 닫혔다고 판단할 확인 기준:
  - `AGENTS.md`에 deep study 계약이 명시된다.
  - skill이 exemplar references를 읽고 그 계약을 집행하도록 바뀐다.
  - skill validation 통과
- 하드 제약 / 호환성 경계:
  - repo SSOT를 external skill이 덮어쓰지 않는다.
  - cards와 general docs의 역할은 유지한다.
- 성공 정의:
  - repo rule 보강 + skill refactor + validation + commit
- PARTIAL 조건:
  - skill만 바뀌고 repo SSOT 반영이 비어 있음
- BLOCKED 조건:
  - skill 구조나 validator가 현재 환경에서 신뢰성 있게 검증되지 않음

## 3. Reader & Internalization Contract

- 주 독자:
  - 미래의 저장소 소유자
  - 이 저장소 문서 작업을 수행할 AI
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - 이 저장소의 기본 산출물이 어떤 학습 문서인가
  - skill이 무엇을 강제하고 무엇을 보조하는가
  - raw material와 exemplar를 어떻게 구분하는가
- 사용자가 내재화해야 할 사고 패턴:
  - 설명은 기초를 먼저 닫고 확장한다
  - 작은 예제가 mental model의 시작점이다
  - 역사/맥락/왜/어떻게/실패/검증이 이어져야 한다
- 특히 막아야 하는 오해:
  - 긴 문서면 자동으로 깊다는 오해
  - 작은 예제 없이도 체화가 된다는 오해
  - `.tmp/interview*.md` 같은 원재료를 바로 exemplar로 삼아도 된다는 오해

## 4. Project Overlay

- 적용한 로컬 AGENTS 경로:
  - `AGENTS.md`
- 활성화한 프로젝트 규칙:
  - exemplar보다 평균 문서 수준을 따르지 않는다
  - 나중에 다시 설명하고 검증 가능한 문서를 만든다
  - repo 변경 작업은 검수/검증/커밋으로 닫는다
- 전역 규칙과의 충돌 여부 / 해소:
  - 충돌 없음

## 5. Topic Analysis

- 현재 이해한 사용자 의도:
  - AI가 일반적인 "정리"를 넘어서, 작은 단위부터 벽돌처럼 쌓아 올리는 학습 자료를 만들도록 하고 싶다.
- 현재 보이는 문제 구조:
  - 기존 skill은 long-form exemplar 지향은 있지만, "first principles -> minimal example -> layered expansion -> real scale" 구조를 default output으로 명시하지 않는다.
  - exemplar는 디렉터리 단위보다 파일 단위가 더 명확하다.
  - `.tmp/interview*.md` / `interviews/*`는 큰 원재료이므로 style exemplar로 삼기보다 source reservoir로 다루는 편이 맞다.

## 6. Scope Expansion & Impact Sync

- 시작 키워드:
  - `study-explanation`, `deep explanation`, `first principles`, `minimal example`, `history`, `why`, `how`
- 확장 키워드:
  - `entry.S`, `dynamic_programming`, `optimal_solution`, `git_rebase`, `threads`, `java_synchronized`, `interviews`
- 조사한 경로:
  - exemplar files
  - `AGENTS.md`
  - current skill files
  - `.tmp` / `interviews` 파일 목록과 규모
- 함께 움직여야 하는 표면:
  - repo SSOT
  - external skill body
  - skill references
  - skill UI metadata

## 7. Evidence Ledger

- E-01
  - 주장: 사용자가 원하는 exemplar는 directory family보다 file-level로 더 구체적이다.
  - 근거 유형: `repo evidence`
  - 자료:
    - user가 직접 6개 exemplar file 지정
  - 이 자료로 닫힌 것:
    - skill reference는 exact file exemplar를 가져야 함
- E-02
  - 주장: exemplar 문서들은 모두 "기초 -> 메커니즘 -> 예제/확장" 성격을 강하게 가진다.
  - 근거 유형: `repo evidence`
  - 자료:
    - `dynamic_programming.md`
    - `optimal_solution.md`
    - `threads.md`
    - `java_synchronized.md`
    - `git_rebase.md`
    - `entry.S`
  - 이 자료로 닫힌 것:
    - deep study contract의 공통 패턴
- E-03
  - 주장: `.tmp/interview*.md` / `interviews/*`는 대형 원재료로서 다루는 규칙이 필요하다.
  - 근거 유형: `repo evidence`
  - 자료:
    - `.tmp/interviews.md` ~ 67709 words
    - `interviews/interview_questions.md` ~ 68147 words
  - 이 자료로 닫힌 것:
    - raw material handling rule 필요

## 8. Design

- 선택한 접근:
  - `AGENTS.md`에 deep study monograph 계약을 추가하고, `study-explanation` skill은 그 계약을 집행하도록 references 기반으로 재구성한다.
- 고려한 대안:
  - A. skill만 수정
  - B. AGENTS만 수정
  - C. AGENTS + skill + exemplar references
- 대안을 채택하지 않은 이유:
  - A는 SSOT drift 위험
  - B는 자동 실행 루프가 약함
  - C가 가장 균형이 좋음
- 문서 / 자산 구조:
  - `AGENTS.md`: 설명 계약 보강
  - `~/.codex/skills/study-explanation/SKILL.md`: concise workflow
  - `~/.codex/skills/study-explanation/references/*.md`: exemplar와 deep contract
  - `~/.codex/skills/study-explanation/agents/openai.yaml`: UI metadata refresh

## 9. Frozen Checklist

- C-01: `AGENTS.md`가 deep study monograph 기본값을 명시한다.
- C-02: skill이 exact exemplar files를 reference로 삼도록 바뀐다.
- C-03: skill이 raw material(`.tmp/interview*.md`, `interviews/*`) 처리 규칙을 가진다.
- C-04: skill metadata와 validator가 통과한다.
- C-05: repo 변경분이 WORK와 함께 commit 된다.

## 10. Verification Plan

- repo diff review
- skill file / references inspection
- `generate_openai_yaml.py`로 metadata 갱신
- `quick_validate.py`로 skill validation
- repo 변경 commit

## 11. Final Audit & Closure

- 최종 상태:
  - `COMPLETE`
- 검증 결과:
  - `AGENTS.md`에 deep study monograph 계약, canonical exemplar, raw material rule 반영
  - external skill에 `references/deep-study-monograph.md`, `references/canonical-exemplars.md` 추가
  - `generate_openai_yaml.py`로 `agents/openai.yaml` 재생성
  - `quick_validate.py` 결과: `Skill is valid!`
- 실제 실행한 검증:
  - exemplar file 구조 검토
  - `.tmp` / `interviews` 파일 목록과 규모 확인
  - `git diff -- AGENTS.md WORK_20260413_STUDY_EXPLANATION_DEEP_SKILL_REFACTOR.md`
  - `/tmp/study-explanation-skill-validate/bin/python .../generate_openai_yaml.py ...`
  - `/tmp/study-explanation-skill-validate/bin/python .../quick_validate.py /Users/rody/.codex/skills/study-explanation`
- 최종 감사:
  - skill의 기본 산출물이 "좋은 설명" 수준이 아니라 "기초부터 쌓는 학습 문서"로 더 강하게 고정되었고, repo SSOT와 external executor의 역할 분리도 유지했다.
- 체크리스트 재판정:
  - C-01: `PASS`
  - C-02: `PASS`
  - C-03: `PASS`
  - C-04: `PASS`
  - C-05: `PASS`
- 커밋 해시 / 미커밋 사유:
  - 최종 repo commit 후 git history와 최종 보고에 기록한다.
