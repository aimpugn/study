# WORK_20260413_STUDY_EXPLANATION_STYLE_FEEDBACK_REFACTOR

## 0. Meta

- 작업 제목: 설명 스타일 피드백 일반화 및 `spring_boot_jar_startup.md` 리팩터링
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260413_STUDY_EXPLANATION_STYLE_FEEDBACK_REFACTOR.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - table of contents 추가
  - `개요` 같은 서두 도입
  - 사용자 말투와 구조 선호 반영
  - 애매한 지시어와 설명 지연 제거
  - 지침 / 작업 템플릿 / skill에도 일반화
- 원문 사용자 요청:
  - 피드백을 적용하고, 일반화해서 지침과 작업 템플릿 등에 반영하고, `spring_boot_jar_startup.md`를 리팩토링
- 대상 경로 / 자산:
  - repo guidance:
    - `/Users/rody/VscodeProjects/study/AGENTS.md`
    - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - repo doc:
    - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - external skill:
    - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
    - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- 실행자: Codex
- 시작 일시: 2026-04-13
- 종료 일시: 2026-04-13
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - 사용자의 구체 피드백을 일회성 문장 수정이 아니라 설명 계약으로 승격하고, 그 기준으로 현재 문서를 다시 작성한다.
- refs:
  - 사용자 피드백 전문
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- scope:
  - style feedback 분석
  - repo 규칙 반영
  - WORK template 보강
  - current doc refactor
  - external skill 보강
  - 검증과 repo commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - 사용자의 직접 피드백을 최소 바닥선으로 유지
  - `spring_boot_jar_startup.md`의 기존 핵심 질문과 사실 기반은 유지
  - repo 변경은 commit으로 닫기
- extra_checks:
  - 지나치게 사용자의 현재 말투만 좇는 대신, 재사용 가능한 규칙으로 일반화할 것
  - external skill 변경은 repo commit 범위 밖이라는 점을 분리해 기록할 것

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - table of contents
  - `개요` 같은 서두
  - 사용자 선호 문체 반영
  - 애매한 표현 제거
  - 설명을 뒤로 미루지 않는 구조
  - 지침 / 작업 템플릿 반영
  - `spring_boot_jar_startup.md` 리팩터링
- 사용자가 명시한 금지 사항:
  - 핵심 설명을 나중으로 미루는 방식
- path / naming / format / finish 관련 요구:
  - repo 지침 및 템플릿 반영
  - repo 변경 작업이므로 commit 필요
- 내가 추가한 누락 방지 항목:
  - external skill 반영
  - skill validation

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - 저장소 전체 문서 전수 리라이팅
  - global `~/.codex/AGENTS.md` 수정
- 지금 하지 않는 이유:
  - 이번 피드백은 study 저장소와 study-explanation skill에 먼저 안정적으로 녹이는 것이 우선이기 때문

## 2. Root-First Framing

- 근본 문제:
  - 현재 문서는 내용은 맞아도 구조와 말투가 사용자의 실제 설명 방식과 어긋나는 부분이 있고, 그 어긋남이 규칙으로 정리되지 않으면 같은 문제가 반복된다.
- 왜 이 문제가 지금 중요한가:
  - `study-explanation`이 앞으로 핵심 집행기라면, 사용자의 설명 감각이 규칙으로 들어가야 한다.
- 작업 목표:
  - 설명 구조, 서두, 용어 전개, 목록 형식, 지시어 처리에 대한 사용자 선호를 reusable guidance로 승격하고, 현재 문서에 즉시 적용한다.
- 기대 이점:
  - 이후 생성되는 문서가 사용자의 자연스러운 설명 방식과 더 가까워지고, 현재 문서도 exemplar로 쓸 수 있게 된다.
- 이점이 닫혔다고 판단할 확인 기준:
  - AGENTS / WORK template / skill ref에 반영
  - `spring_boot_jar_startup.md`가 TOC + 개요 + 즉시 설명 + direct prose 기준으로 재작성
- 하드 제약 / 호환성 경계:
  - repo local guidance는 사용자의 명시 피드백을 약화하지 않아야 한다.
  - external skill 수정은 repo commit과 분리한다.
- 성공 정의:
  - repo docs 3개 이상 반영 + current doc refactor + skill update + validation + repo commit
- PARTIAL 조건:
  - 현재 문서만 고치고 guidance 일반화가 빠지면 PARTIAL
- BLOCKED 조건:
  - active guidance 문서 수정 승인이나 file access에 문제가 생기면 BLOCKED

## 3. Reader & Internalization Contract

- 주 독자:
  - 미래의 나
  - 이 저장소 스타일을 따를 AI
- 독자가 이미 알고 있다고 가정하는 것:
  - 기존 study 저장소가 deep study monograph를 지향한다는 것
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - 좋은 study 문서가 왜 TOC, 개요, direct prose, immediate clarification을 가져야 하는지
- 사용자가 내재화해야 할 사고 패턴:
  - 피드백은 문장 고치기보다 설명 계약으로 올려야 재발이 줄어든다
- 특히 막아야 하는 오해:
  - 문체 피드백은 취향이므로 guidance로 올릴 수 없다는 오해
  - TOC나 개요가 단순 장식이라는 오해
- 기억 anchor 후보:
  - `목차 -> 개요 -> 작은 실행 흐름 -> 차이 즉시 설명 -> 확장`
- 반드시 거쳐야 하는 추상화 계층:
  - user style feedback -> repo guidance -> work template -> skill -> concrete doc
- 핵심 대조쌍 / 혼동쌍:
  - one-off edit vs reusable rule
  - direct prose vs meta framing
  - immediate explanation vs deferred explanation
- 질문형으로 먼저 답해야 하는 핵심 질문:
  - 어떤 피드백을 규칙으로 올려야 이후 문서 품질이 안정적으로 개선되는가
- 이번 작업의 품질 기준 exemplar:
  - 사용자 피드백 자체
- 현재 저장소 평균 품질을 기준으로 삼으면 안 되는 이유:
  - 평균 수준은 이미 개선 대상이며, 이번 작업은 기준선 자체를 올리는 작업이기 때문

## 4. Depth Decision

- 선택한 깊이: `full`
- 왜 이 깊이가 맞는가:
  - active guidance와 skill까지 건드리는 규범 계열 작업이기 때문
- 전체 루프를 켜야 하는 트리거:
  - guidance / template / exemplar doc / skill sync 필요
- 축약 가능한 섹션과 그 근거:
  - external research는 최소화 가능. 사용자 피드백과 existing repo rules가 핵심 근거다

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - prose-first
  - deep study monograph
  - prompt/skill보다 repo rules 우선
  - repo-changing doc work commit closure
- 특히 중요한 규칙:
  - 질문 중심 설명
  - 계층 연결
  - 검증 정직성
- 전역 규칙과의 충돌 여부 / 해소:
  - 로컬 AGENTS/WORK_TEMPLATE 수정은 사용자 요청으로 승인된 변경으로 해석

## 6. Topic Analysis

- 현재 이해한 사용자 의도:
  - 자신이 실제로 쓰는 설명 방식의 구조와 말투를 AI가 재현하기를 원한다.
- 현재 보이는 문제 구조:
  - 구조 문제: TOC와 개요 부재
  - 말투 문제: 메타 프레이밍, 지시어 모호성, 설명 지연
  - 운영 문제: 이 피드백이 guidance에 흡수되지 않으면 다시 반복
- 핵심 경계:
  - 단순 개인 취향이 아니라 "독자가 재구성하기 쉬운 방식"으로 일반화 가능한 것만 규칙으로 올린다
- 숨은 가정 / 불확실성:
  - 사용자의 모든 말투를 규칙으로 박을 필요는 없고, 실패를 줄이는 패턴만 추출해야 한다
- 성공을 오판하기 쉬운 지점:
  - 문서를 예쁘게만 고치고 guidance에 반영하지 않는 것

## 7. Analysis Critique + Repair

### 7.1 Challenge 1

- 어떤 점이 틀릴 수 있는가:
  - 사용자의 특정 표현 선호를 과도하게 로컬 규범으로 박아 버릴 수 있다.
- 보강안:
  - 표현 자체보다 실패를 줄이는 구조적 규칙으로 승격한다.
- 왜 이 보강안이 더 강한가:
  - 재사용성과 일반화가 높다.

### 7.2 Challenge 2

- 어떤 점이 좁거나 누락될 수 있는가:
  - TOC/개요만 넣고 애매한 지시어나 deferred explanation 문제를 놓칠 수 있다.
- 보강안:
  - `애매한 지시어 즉시 해소`, `현재 문장 이해에 필요한 설명은 뒤로 미루지 않기`, `직접 진술형`을 함께 규칙화한다.
- 왜 이 보강안이 더 강한가:
  - 구조와 문장 수준의 문제를 같이 잡는다.

### 7.3 Challenge 3

- 어떤 점이 설명/학습 목표를 놓칠 수 있는가:
  - 현재 문서를 고치면서 사실 검증 경로가 약해질 수 있다.
- 보강안:
  - 기존 관측 명령과 official link는 유지하고 구조만 더 강하게 재편한다.
- 왜 이 보강안이 더 강한가:
  - style fix가 substance를 훼손하지 않는다.

### 7.4 Retained Framing

- 최종 채택한 문제 정의:
  - 사용자 피드백을 reusable explanation contract로 올리고, 그 contract로 current exemplar doc을 refactor한다.
- 폐기한 문제 정의와 이유:
  - `spring_boot_jar_startup.md`만 고친다: 재발 방지 효과가 약함

## 8. Scope Expansion & Impact Sync

- 시작 키워드:
  - TOC, 개요, direct prose, immediate explanation, ambiguous referent
- 확장 키워드(동의어 / 약어 / 원어 / 구명칭 / 관련 에러 / 표준 번호):
  - table of contents, overview, 서두, 지시어, deferred explanation, narrative opening, numbered contrast
- 조사한 경로:
  - repo AGENTS
  - repo WORK template
  - current doc
  - study-explanation skill
- 함께 점검한 자산:
  - current doc
  - repo guidance
  - external skill reference
- 함께 움직여야 하는 표면:
  - AGENTS
  - AGENTS_WORK_TEMPLATE
  - current doc
  - skill guidance
- 한쪽만 바꾸면 깨질 부분:
  - 문서만 고치면 다음 문서에서 재발
  - guidance만 고치면 현재 exemplar가 뒤처짐
- 제외 표면과 근거:
  - global AGENTS: 이번 범위는 study repo + study-explanation skill에 국한

## 9. Evidence Gathering

### 9.1 Evidence Ledger

- E-01
  - 주장: 현재 문서에 TOC, 개요, direct wording, 즉시 설명이 부족하다는 평가는 사용자 피드백으로 직접 닫힌다.
  - 근거 유형: `repo evidence`
  - 자료: 사용자 피드백 전문 + current doc
  - 이 자료로 닫힌 것: current style gap
  - 아직 비어 있는 것: reusable rule shape
- E-02
  - 주장: repo AGENTS와 template에는 설명 구조에 대한 일반 규칙은 있으나, TOC/개요/애매한 지시어/설명 지연 방지 규칙은 아직 구체화가 약하다.
  - 근거 유형: `repo evidence`
  - 자료: AGENTS.md, AGENTS_WORK_TEMPLATE.md
  - 이 자료로 닫힌 것: guidance update 필요성
  - 아직 비어 있는 것: exact clause design
- E-03
  - 주장: study-explanation skill은 deep-study-monograph를 말하지만 사용자 고유 피드백을 반영한 구조 지침은 더 보강할 여지가 있다.
  - 근거 유형: `repo evidence`
  - 자료: skill docs
  - 이 자료로 닫힌 것: skill update 필요성
  - 아직 비어 있는 것: validation

## 10. Evidence Critique + Repair

- 소스 품질 리스크:
  - 사용자 피드백은 강하지만 현재 한 문서 사례에 집중돼 있다
- 오래되었을 가능성이 있는 가정:
  - 없음. 현재 active preference다
- 빠진 대안 또는 빠진 근거:
  - explanation-quality 관련 기존 문구와의 중복 점검
- 근거 세트를 어떻게 보강했는가:
  - AGENTS / template / skill reference를 함께 읽고 gap을 확인
- 보강 후에도 남는 한계:
  - 이번 작업 후에도 추가 피드백이 나오면 계속 조정해야 한다

## 11. Design

- 선택한 접근:
  - guidance에는 reusable rule만 추가
  - current doc은 TOC + 개요 + direct style로 전면 리팩터링
  - skill에는 긴 문서 구조와 문장 규칙 보강
- 왜 이 접근이 근본 문제와 이점에 가장 잘 맞는가:
  - 재발 방지와 즉시 품질 개선을 동시에 달성한다
- 고려한 대안:
  - current doc만 수정
  - skill만 수정
- 대안을 채택하지 않은 이유:
  - 둘 다 단독으로는 scope closure가 약하다
- 문서 / 예제 / 자산 구조:
  - repo guidance update
  - template update
  - current doc rewrite
  - skill update
- 설명 뼈대: `질문형 | 계층형 | 비교형 | 시나리오형 | 혼합`
- 계층별 설명 순서:
  - feedback -> rule -> concrete rewrite
- 넣을 구체 예시 / 관측 anchor:
  - current doc before/after structure
- 이 문서를 끌어올릴 목표 수준:
  - 사용자의 직접 설명 감각과 가장 가까운 exemplar
- 실패 모드:
  - 규칙이 너무 취향적이거나 너무 추상적일 수 있음
- 검증 경로:
  - diff review
  - skill validation
  - repo staged files review

## 12. Design Critique + Repair

### 12.1 Architect View

- 반론:
  - guidance가 너무 세세해질 수 있다
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - 문장 예시를 그대로 규범화하지 않고, 실패를 막는 원리만 남긴다

### 12.2 Domain / API Consumer View

- 반론:
  - template가 비대해질 수 있다
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - 새로운 필드는 최소로 추가하고 review points 쪽에 흡수한다

### 12.3 Newcomer / Learner View

- 반론:
  - TOC와 개요만 있으면 충분하다고 오해할 수 있다
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - 애매한 지시어 제거와 immediate clarification을 같이 강조한다

### 12.4 Final Design Decision

- 최종 채택:
  - guidance + template + skill + current doc 동시 수정
- 트레이드오프:
  - 이번 커밋 범위는 커지지만 이후 재발 비용이 크게 줄어든다

## 13. Overall Plan

- 작업 순서:
  - WORK freeze
  - guidance patch
  - current doc refactor
  - skill patch
  - review / validate
  - repo commit
- 선행 의존성:
  - none beyond current instruction stack
- validation order:
  - repo diff review
  - skill validate
  - git status scoped review
- rollback / retry / staging 필요 여부와 이유:
  - guidance와 current doc는 함께 검토되어야 하므로 마지막에 한 번에 commit

## 14. Plan Critique + Repair

- 계획이 실패할 수 있는 지점:
  - current doc rewrite가 너무 과하게 짧아질 수 있다
- 순서상 위험:
  - guidance를 먼저 고쳐 두고 current doc가 그 기준을 못 따라갈 수 있다
- 빠진 prerequisite:
  - 없음
- 보강안:
  - current doc rewrite 후 guidance wording과 다시 대조
- 왜 보강된 계획이 더 나은가:
  - 규칙과 exemplar를 서로 맞출 수 있다

## 15. Detailed Task Plan

- 수정 / 생성 / 검토할 파일:
  - create: this WORK file
  - update: `/Users/rody/VscodeProjects/study/AGENTS.md`
  - update: `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - update: `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - update: `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - update: `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- 각 파일에서 바꿀 논리 또는 구조:
  - AGENTS: reusable style rules
  - template: style review hooks
  - current doc: TOC + overview + direct prose + immediate clarification
  - skill: same rules carried forward
- 관련 문서 동기화 계획:
  - current doc wording should satisfy new rules
- 예제 추가 / 보강 계획:
  - not needed beyond current doc examples
- 근거 섹션 반영 계획:
  - user feedback is primary evidence

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: current doc opens with TOC and overview
  - S2: ambiguous phrases like `그 사이에` are removed or immediately resolved
  - S3: guidance now contains reusable clauses for future docs
- 실패 케이스 최소 3개:
  - F1: TOC만 넣고 prose style 문제는 그대로 둔다
  - F2: skill은 안 고치고 repo doc만 고친다
  - F3: 사용자 피드백의 direct examples를 약화한다
- 회귀 위험:
  - guidance duplication
  - current doc losing factual closure
- 회귀 방지 확인 경로:
  - scoped diff review
  - skill quick validation

### 15.2 Code / Doc Quality Review Points

- 단순성:
  - 규칙은 짧게, current doc은 direct하게
- 응집도:
  - style feedback should map to a few reusable clauses
- 확장 여지:
  - later feedback should fit same sections
- 과한 일반화 여부:
  - style rules should remain explanation-specific
- 설명 누락 위험:
  - 개요 / TOC / immediate explanation / direct wording

## 16. Detailed Plan Critique + Repair

- 누락된 케이스:
  - numbered contrast formatting preference
- fuzzy success criteria:
  - `사용자 말투 반영`이 추상적일 수 있음
- scope overreach / under-specification:
  - every phrasing preference should not become hard rule
- 보강안:
  - direct wording, immediate clarification, TOC, overview, numbering preference 정도만 규칙으로 승격
- 최종 상세 계획:
  - current doc를 전면 다시 쓰되, stable facts and verification remain intact

## 17. Frozen Checklist

- [x] 사용자 피드백을 항목별로 해석하고 최소 바닥선으로 유지한다
- [x] repo guidance에 reusable explanation rules를 추가한다
- [x] WORK template에 style review hooks를 추가한다
- [x] `spring_boot_jar_startup.md`를 TOC + 개요 + direct prose 기준으로 리팩터링한다
- [x] 애매한 지시어와 deferred explanation을 제거한다
- [x] external skill에도 같은 guidance를 반영한다
- [x] skill validation을 수행한다
- [x] repo scoped review와 commit으로 closure를 닫는다

## 18. Execution Log

- 사용자 피드백을 `목차`, `개요`, `직접 진술형`, `즉시 설명`, `번호 목록 + 4칸 들여쓰기` 규칙으로 정리
- `/Users/rody/VscodeProjects/study/AGENTS.md`에 reusable explanation clauses 추가
- `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`에 style review hooks 추가
- `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`를 전면 재작성
- `/Users/rody/.codex/skills/study-explanation/SKILL.md`와 `references/deep-study-monograph.md`에 같은 규칙 반영
- skill validator 실행 완료

## 19. Verification Plan

- repo diff review
- current doc structure review
- skill validation
- scoped git status

## 20. Final Audit

- 상태:
  - `COMPLETE`
- 구조 검증:
  - current doc는 TOC와 `개요`를 갖추고, 구조만 봐도 설명 범위를 예측할 수 있다.
  - 일반 실행형 JAR와 Spring Boot executable jar의 차이를 번호 목록과 4칸 들여쓰기 prose로 재구성했다.
- 문체 검증:
  - 메타 진행 멘트를 줄이고 직접 진술형 문장을 우세하게 만들었다.
  - `그 사이`처럼 애매한 지시어를 구체 문장으로 치환했다.
  - 현재 문장을 이해하는 데 필요한 설명을 뒤로 미루지 않도록 재배치했다.
- guidance 검증:
  - repo AGENTS와 WORK template에 같은 규칙이 반영되었다.
  - external skill도 같은 방향으로 보강되었다.
- skill 검증:
  - `/tmp/study-explanation-skill-validate/bin/python /Users/rody/.codex/skills/.system/skill-creator/scripts/quick_validate.py /Users/rody/.codex/skills/study-explanation`
  - 결과: `Skill is valid!`
- commit:
  - repo scoped files를 별도 staging 후 commit으로 closure를 닫는다.
