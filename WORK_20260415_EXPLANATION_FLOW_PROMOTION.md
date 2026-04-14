# WORK_20260415_EXPLANATION_FLOW_PROMOTION

## 0. Meta

- 작업 제목: 선택된 설명 흐름을 AGENTS와 study-explanation skill에 승격
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260415_EXPLANATION_FLOW_PROMOTION.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - 선택된 설명 흐름을 guidance와 skill에 반영
- 원문 사용자 요청:
  - 네 진행해주세요.
- 대상 경로 / 자산:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - `/Users/rody/VscodeProjects/study/documentation/study_explanation_style_candidates.md`
- 실행자: Codex
- 시작 일시: 2026-04-15
- 종료 일시: 2026-04-15
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - 사용자가 선택한 설명 전개 순서 `A+B -> 필요시 E -> C`를 로컬 AGENTS, WORK 템플릿, study-explanation skill의 기본 계약으로 승격한다.
- refs:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/USECASE.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - `/Users/rody/VscodeProjects/study/documentation/study_explanation_style_candidates.md`
- scope:
  - promotion target contract 고정
  - repo AGENTS 반영
  - WORK 템플릿 반영
  - skill 본문과 reference 반영
  - 검토 / 검증 / commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - blacklist가 아니라 원칙/판정 질문/흐름 계약으로 승격
  - 일반 deep study 문서의 기본 흐름을 명시
  - 후보 D와 E의 역할은 과하게 승격하지 않고 현재 선택에 맞게 재배치
- extra_checks:
  - repo guidance와 skill guidance가 어긋나지 않아야 한다
  - future WORK ledger가 이 흐름을 실제로 기록할 수 있어야 한다

## 17. Frozen Checklist

### 17.1 Checklist Draft

- C-01
  - 출처: `사용자`
  - 내용:
    - 선택된 설명 흐름을 AGENTS에 반영한다
  - PASS 기준:
    - 로컬 AGENTS 설명 계약에 `A+B -> 필요시 E -> C`가 구조 규칙으로 반영된다
  - FAIL 기준:
    - 여전히 후보 비교 수준에 머문다
  - 필요한 증거:
    - AGENTS diff
- C-02
  - 출처: `AI-추가`
  - 내용:
    - WORK 템플릿이 같은 흐름을 기록하고 점검할 수 있게 바뀐다
  - PASS 기준:
    - Reader/Internalization Contract나 Design 쪽에 해당 흐름을 적을 자리가 생긴다
  - FAIL 기준:
    - future ledger에서 다시 즉흥적으로 적어야 한다
  - 필요한 증거:
    - AGENTS_WORK_TEMPLATE diff
- C-03
  - 출처: `사용자`
  - 내용:
    - study-explanation skill이 같은 흐름을 기본값으로 사용한다
  - PASS 기준:
    - SKILL.md와 deep-study-monograph reference에 동일한 흐름이 반영된다
  - FAIL 기준:
    - repo guidance와 skill guidance가 어긋난다
  - 필요한 증거:
    - skill diff
- C-04
  - 출처: `AI-추가`
  - 내용:
    - 검토, 가능한 검증, commit으로 닫는다
  - PASS 기준:
    - targeted review와 skill file syntax-level validation 또는 equivalent check가 있다
  - FAIL 기준:
    - 수정만 하고 닫는다
  - 필요한 증거:
    - verification log + commit

### 17.2 Checklist Quality Review

- [x] 각 항목이 목표, 이점, 불변식에 매핑된다.
- [x] PASS/FAIL이 관측 가능하다.
- [x] 필요한 근거 또는 검증 경로가 있다.
- [x] 사용자 요구가 조용히 약화되지 않았다.
- [x] 한 항목 실패 시 task가 reopened 되는 구조다.
- 판정:
  - PASS
- 보완 사항:
  - 없음

### 17.3 Freeze

- freeze 시각:
  - 2026-04-15
- freeze 버전:
  - v1
- freeze 이후 추가된 항목과 이유:
  - 없음

## 2. Root-First Framing

- 근본 문제:
  - 설명 스타일 후보는 좁혀졌지만, 아직 active guidance와 skill에는 기본 계약으로 승격되지 않았다.
- 왜 이 문제가 지금 중요한가:
  - 다음 문서부터 같은 흐름이 자동으로 재사용되지 않으면 다시 문서별 수동 피드백 루프로 돌아가게 된다.
- 작업 목표:
  - 선택된 설명 흐름을 repo와 skill의 상시 규칙으로 고정한다.
- 기대 이점:
  - future 문서가 `직접 진술 -> 등장 배경과 이유 -> 공통 구조/메타데이터 -> 필요시 짧은 실무 연결 -> 실행 경로 추적` 순서를 기본값으로 따르게 된다.
- 이점이 닫혔다고 판단할 확인 기준:
  - AGENTS, WORK 템플릿, skill, reference가 같은 흐름을 가리킨다.
- 하드 제약 / 호환성 경계:
  - 문구 blacklist로 후퇴하면 안 된다.
  - 후보 이름보다 실제 설명 순서와 판정 질문이 우선이어야 한다.
- 성공 정의:
  - 대상 파일 반영 + 검토 + 검증 + commit
- PARTIAL 조건:
  - 일부 파일만 반영되어 guidance가 분리되면 PARTIAL
- BLOCKED 조건:
  - skill syntax 또는 repo 규칙과 충돌해 승격을 완료할 수 없으면 BLOCKED

## 3. Reader & Internalization Contract

- 주 독자:
  - future Codex
  - future me
- 독자가 이미 알고 있다고 가정하는 것:
  - style candidate 문서에서 현재 선택 방향이 정리되었다는 것
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - 일반 학습 문서를 어떤 순서로 열고, 어디서부터 실행 경로 추적으로 내려가야 하는지
- 사용자가 내재화해야 할 사고 패턴:
  - 좋은 설명은 문장 취향보다 `어떻게 시작하고 무엇을 먼저 고정한 뒤 어디서부터 깊게 들어가는가`로 판단한다
- 특히 막아야 하는 오해:
  - A/B/C/E가 서로 배타적인 개별 문체라고 보는 오해
  - E가 모든 문서의 기본 시작점이라고 보는 오해
- 기억 anchor 후보:
  - `직접 진술 -> 등장 배경과 이유 -> 공통 구조/메타데이터 -> 필요시 짧은 실무 연결 -> 실행 경로 추적`
- 반드시 거쳐야 하는 추상화 계층:
  - repo guidance
  - WORK ledger
  - skill main instructions
  - skill reference
- 핵심 대조쌍 / 혼동쌍:
  - 후보 비교 문서 vs active rule
  - 기본 전개 흐름 vs 보조 overlay
- 질문형으로 먼저 답해야 하는 핵심 질문:
  - 일반 deep study 문서는 어떤 순서로 전개하는 것이 이 저장소의 현재 기본값인가
- 목차 필요 여부와 이유:
  - 불필요. 이번 작업은 guideline promotion이므로 section-oriented update면 충분하다
- `개요` 또는 서두에서 먼저 고정할 문장:
  - 구조와 실행 경로가 함께 중요한 일반 학습 문서는 `직접 진술 -> 등장 배경과 이유 -> 공통 구조/메타데이터 -> 필요시 짧은 실무 연결 -> 실행 경로 추적`을 기본값으로 둔다
- 사용자 / 저장소의 말투, 표기, 목록 형식 선호:
  - direct prose
  - 구조 규칙은 natural Korean으로
  - 내부 shorthand는 보조
- 이번 문서에서 특히 지켜야 할 설명 원칙:
  - 후보 이름보다 실제 설명 순서를 우선 적는다
  - rule promotion과 rationale을 함께 남긴다
- 이번 문서에서 문장 단위로 점검할 판정 질문:
  - 이 문장은 active rule을 직접 말하는가
  - 이 문장은 다음 문서 작성자가 실제로 적용할 수 있는가
- bad -> better 예시 페어(선택):
  - `A를 기본으로 하고 B를 섞는다` -> `먼저 직접 진술하고, 그 구조가 왜 생겼는지와 공통 메타데이터를 고정한 뒤, 필요하면 실무 중요성을 짧게 연결하고, 그다음 실행 경로를 추적한다`
- 바로 풀어써야 하는 지시어 / 애매한 연결 표현:
  - `이 흐름`
  - `이 조합`
  - `보조 축`
- 초반에 먼저 보여줄 공식 규약 / 실제 구조 / 예시:
  - style candidate doc의 선택 결과
- 독자가 자연스럽게 묻게 될 한 단계 아래 질문:
  - 실무 시나리오는 언제 짧게만 쓰고 언제 전면에 둘 것인가
- 독자가 자연스럽게 묻게 될 역사 / 등장 맥락 / 이전 방식 질문:
  - 왜 blacklist식 규칙보다 구조 승격이 더 강한가
- 현재 머신 관측과 일반 규칙을 구분해야 하는 지점:
  - 없음
- 먼저 분해해서 설명할 구성요소:
  - opening flow
  - optional practical connection
  - execution-path descent
- 그 구성요소들이 어떻게 엮여 상위 메커니즘이 되는가:
  - opening flow가 구조 이해를 잡아 주고, optional practical connection이 중요도를 연결하며, execution-path descent가 깊이를 만든다
- 번호 목록으로 써야 하는 순차 흐름:
  - selected opening flow
- 불릿 목록으로 써야 하는 동등 비교축:
  - affected files
- 한 줄로 닫을 항목과 들여쓴 prose가 필요한 항목:
  - rule statement는 한 줄
  - rationale은 prose
- 공식 자료를 어떤 인과 흐름으로 설명에 연결할 것인가:
  - 이번 작업의 primary source는 repo candidate doc과 user feedback이다
- 공식 자료 인용이 멈추면 FAIL로 볼 판정 지점:
  - 해당 없음
- 이번 작업의 품질 기준 exemplar:
  - `/Users/rody/VscodeProjects/study/documentation/study_explanation_style_candidates.md`
- 현재 저장소 평균 품질을 기준으로 삼으면 안 되는 이유:
  - 이번 작업은 average style이 아니라 selected best-fit flow를 승격하는 작업이다
- 속도를 이유로 줄이면 안 되는 조사 / 비교 / 검증 / 설명 단계:
  - AGENTS/skill consistency check
  - syntax-level skill validation
  - final reread

## 4. Depth Decision

- 선택한 깊이:
  - `full`
- 왜 이 깊이가 맞는가:
  - repo guidance, work template, external skill reference가 함께 움직인다
- 전체 루프를 켜야 하는 트리거:
  - active rule promotion
- 축약 가능한 섹션과 그 근거:
  - 새로운 external research는 불필요. current repo decisions and user feedback are the source of truth

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - deep study monograph
  - direct prose
  - quality over speed
  - active style contract promotion
- 특히 중요한 규칙:
  - explanation rules should be generalized into reusable structure
- 전역 규칙과의 충돌 여부 / 해소:
  - 충돌 없음

## 6. Topic Analysis

- 현재 이해한 사용자 의도:
  - 후보 비교는 끝났고, 이제 실제 집행 규칙으로 고정하고 싶다
- 현재 보이는 문제 구조:
  - AGENTS에는 설명 계약이 많지만, 이번에 선택된 opening flow가 아직 explicit default는 아니다
  - WORK 템플릿은 이 흐름을 future ledger에서 바로 적게 만들지 못한다
  - skill은 deep study monograph를 설명하지만 selected flow가 explicit하지 않다
- 핵심 경계:
  - 후보 이름 자체는 보조이고, active rule은 실제 설명 순서여야 한다
- 숨은 가정 / 불확실성:
  - 모든 문서가 동일한 정도로 E를 필요로 하지는 않는다
- 성공을 오판하기 쉬운 지점:
  - A+B+C+E 이름만 적고 실제 적용 문장을 안 적는 것

## 7. Analysis Critique + Repair

### 7.1 Challenge 1

- 어떤 점이 틀릴 수 있는가:
  - E를 지나치게 올리면 기초 설명 문서가 실무 시나리오로 시작하는 습관이 생길 수 있다
- 보강안:
  - E는 `필요하면 짧게 연결`하는 보조 축으로만 적는다
- 왜 이 보강안이 더 강한가:
  - 현재 사용자 피드백과 가장 정확히 맞는다

### 7.2 Challenge 2

- 어떤 점이 좁거나 누락될 수 있는가:
  - candidate names만 남기면 future writer가 실제 문장 흐름으로 번역하지 못할 수 있다
- 보강안:
  - 이름보다 `직접 진술 -> 등장 배경과 이유 -> 공통 구조/메타데이터 -> 필요시 짧은 실무 연결 -> 실행 경로 추적`을 먼저 적는다
- 왜 이 보강안이 더 강한가:
  - active rule로 쓰기 좋다

### 7.3 Challenge 3

- 어떤 점이 설명/학습 목표를 놓칠 수 있는가:
  - skill만 바꾸고 WORK 템플릿을 안 바꾸면 future execution drift가 남을 수 있다
- 보강안:
  - WORK 템플릿에 selected flow를 직접 적는 항목과 review 항목을 넣는다
- 왜 이 보강안이 더 강한가:
  - future task execution까지 닫힌다

### 7.4 Retained Framing

- 최종 채택한 문제 정의:
  - selected explanation flow를 repo guidance + work template + skill stack에 동시에 승격한다
- 폐기한 문제 정의와 이유:
  - skill만 먼저 수정: repo SSOT와 WORK path가 비게 됨

## 8. Scope Expansion & Impact Sync

- 시작 키워드:
  - selected explanation flow
  - direct statement
  - why/history
  - shared metadata
  - execution path
- 확장 키워드(동의어 / 약어 / 원어 / 구명칭 / 관련 에러 / 표준 번호):
  - opening flow
  - practical stake
  - runtime path
  - deep study monograph
- 조사한 경로:
  - AGENTS
  - WORK template
  - skill main doc
  - skill reference
  - candidate doc
- 함께 점검한 자산:
  - repo guidance
  - skill guidance
- 함께 움직여야 하는 표면:
  - AGENTS
  - AGENTS_WORK_TEMPLATE
  - SKILL.md
  - deep-study-monograph.md
- 한쪽만 바꾸면 깨질 부분:
  - repo rule과 skill rule drift
  - future WORK ledger drift
- 제외 표면과 근거:
  - canonical-exemplars: exemplar set 자체는 이번 승격과 직접 관련 없음

## 9. Evidence Gathering

### 9.1 Evidence Ledger

- E-01
  - 주장:
    - 현재 선택된 기본 흐름은 `A+B -> 필요시 E -> C`다
  - 근거 유형: `repo evidence`
  - 자료:
    - `/Users/rody/VscodeProjects/study/documentation/study_explanation_style_candidates.md`
  - 이 자료로 닫힌 것:
    - promotion target
  - 아직 비어 있는 것:
    - 없음
- E-02
  - 주장:
    - local AGENTS already prefers direct prose, artifact-first, history, and execution-path explanation
  - 근거 유형: `repo evidence`
  - 자료:
    - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - 이 자료로 닫힌 것:
    - promotion can be additive/refining rather than disruptive
  - 아직 비어 있는 것:
    - explicit opening flow
- E-03
  - 주장:
    - skill and reference still describe deep study monograph generally, but not the selected opening flow explicitly
  - 근거 유형: `repo evidence`
  - 자료:
    - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
    - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - 이 자료로 닫힌 것:
    - need for promotion
  - 아직 비어 있는 것:
    - validation after edit

### 9.2 Source Conflicts / Gaps

- 충돌하는 근거:
  - 없음
- 아직 부족한 근거:
  - skill semantic validation path
- 추론으로만 남는 항목:
  - 없음

## 10. Design

- 선택한 접근:
  - existing rule stack를 유지하면서 selected flow를 explicit default로 끼워 넣는다
- 왜 이 접근이 근본 문제와 이점에 가장 잘 맞는가:
  - 기존 설명 계약을 유지하면서 current choice를 실제 집행 규칙으로 승격할 수 있다
- 고려한 대안:
  - AGENTS only
  - skill only
  - brand new playbook
- 대안을 채택하지 않은 이유:
  - AGENTS only는 execution layer가 약하고, skill only는 SSOT가 약하며, new playbook은 문서가 하나 더 늘어난다
- 문서 / 예제 / 자산 구조:
  - repo AGENTS = SSOT refinement
  - WORK template = execution prompts
  - skill docs = execution layer
- 설명 뼈대:
  - `혼합`
- 계층별 설명 순서:
  - direct statement
  - why/history
  - shared structure/metadata
  - optional practical stake
  - execution path
- 넣을 구체 예시 / 관측 anchor:
  - `META-INF/MANIFEST.MF`-style structure/mechanism docs
- 이 문서를 끌어올릴 목표 수준:
  - future tasks can follow without re-deriving the style
- 실패 모드:
  - names without actual rule
  - E overpromotion
  - repo/skill drift
- 검증 경로:
  - targeted reread
  - diff review
  - syntax-level check for YAML/skill files

## 11. Execution Log

- 실제 조사한 것:
  - AGENTS, WORK template, SKILL, deep-study-monograph, candidate doc 재확인
- 실제 수정한 것:
  - AGENTS에 selected opening flow 반영
  - WORK template에 selected flow planning/review 항목 추가
  - SKILL.md와 deep-study-monograph.md에 selected flow 반영
- 실행 중 바뀐 가정:
  - 없음
- earliest affected phase로 되돌아간 이력:
  - deep-study-monograph patch context mismatch로 해당 파일만 다시 읽고 안전하게 재패치
- 버린 접근과 이유:
  - candidate names만 올리는 접근: active rule로 약함

## 12. Verification

### 12.1 Verification Plan

- 실행 / 확인할 명령:
  - targeted `git diff --check`
  - targeted reread
  - YAML parse for `agents/openai.yaml`
  - skill doc/reference presence + syntax sanity checks
- 확인 경로:
  - repo files and skill files align
  - no formatting errors
- PASS 조건:
  - targeted diffs clean
  - selected flow is explicit in all intended files
- FAIL 조건:
  - drift or malformed skill-related files

### 12.2 Verification Result

- 실제 실행한 검증:
-  repo 대상 파일에 대해 `git diff --check -- AGENTS.md AGENTS_WORK_TEMPLATE.md WORK_20260415_EXPLANATION_FLOW_PROMOTION.md`
-  selected flow 핵심 문구가 모든 대상 파일에 들어갔는지 `rg`로 확인
-  skill frontmatter와 `agents/openai.yaml`를 Ruby `YAML.safe_load`로 parse
- 결과:
-  PASS
- 실행하지 못한 검증과 이유:
-  dedicated skill validator는 현재 찾지 못했다. 대신 frontmatter/YAML parse와 content alignment check로 닫았다
- 예제 / 명령 검증 범위:
-  repo guideline formatting
-  skill metadata syntax
-  selected flow alignment
- 소스 검증 범위:
-  AGENTS
-  AGENTS_WORK_TEMPLATE
-  SKILL.md
-  deep-study-monograph.md
-  study_explanation_style_candidates.md

## 13. Final Audit & Closure

- intent-fit review:
-  사용자가 고른 혼합 흐름이 실제 active rule로 승격되었다
- expert-perspective review:
-  후보 이름을 그대로 나열하지 않고, 실제 설명 순서를 active rule로 번역한 점이 핵심 개선이다
- remaining risks:
-  이후 실제 문서 몇 개에 적용해 보면 E를 어느 정도 길이로 허용할지 더 미세 조정할 수 있다
- 문서 / 예제 / 관련 자산 동기화 상태:
-  repo AGENTS와 WORK template는 동기화되었다
-  external skill main doc와 reference도 같은 흐름으로 동기화되었다

### 13.1 Checklist Re-Judgement

- C-01: `PASS`
- C-02: `PASS`
- C-03: `PASS`
- C-04: `PASS`

### 13.2 Final State

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- COMPLETE 승격 조건:
-  selected flow reflected in AGENTS, WORK template, SKILL, reference; verification passed; repo commit created
- 커밋 해시 / 미커밋 사유:
-  이 WORK를 포함한 동일 커밋의 Git history 참조
