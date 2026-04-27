# WORK_20260414_STUDY_EXPLANATION_STYLE_FEEDBACK_ROUND2

## 0. Meta

- 작업 제목: 설명 스타일 피드백 2차 반영 및 `spring_boot_jar_startup.md` 추가 리팩터링
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260414_STUDY_EXPLANATION_STYLE_FEEDBACK_ROUND2.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - 셸이 명령을 어떻게 해석하는지 더 설명
  - 운영 체제가 어떻게 `java`를 실행하는지 더 설명
  - executable format(ELF 등)과 `libjvm` / 동적 로딩 감각 보강
  - `META-INF/MANIFEST.MF` 규약과 공식 링크를 더 이른 시점에 보여주기
  - Spring Boot jar 구조 예제를 더 이른 시점에 보여주기
  - 피드백을 guidance / template / skill에 일반화
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
- 시작 일시: 2026-04-14
- 종료 일시: 2026-04-14
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - 사용자의 2차 피드백을 explanation contract로 승격하고, 현재 문서를 더 아래층부터 친절하게 이해되도록 재구성한다.
- refs:
  - 사용자 피드백 전문
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - local command evidence about `java` launcher binary and JVM libs
- scope:
  - style feedback 분석
  - repo guidance update
  - current doc refactor
  - external skill update
  - skill validation
  - repo commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - 새 피드백을 최소 바닥선으로 유지
  - 현재 문서의 사실 기반은 유지
  - 실제 관측 명령과 공식 링크를 함께 제공
- extra_checks:
  - 현재 로컬이 macOS라는 점을 반영하되, Linux 독자에게는 ELF도 함께 설명

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - shell/OS 실행 과정을 더 자세히 설명
  - `ELF` 같은 executable format 설명
  - `libjvm`, 동적 로딩에 대한 설명
  - `META-INF/MANIFEST.MF`의 의미와 공식 링크
  - 공식 설명과 실제 예제를 더 앞쪽에 배치
  - guidance/template 반영
  - current doc refactor
- 사용자가 명시한 금지 사항:
  - 설명을 뒤로 미루는 방식
- path / naming / format / finish 관련 요구:
  - repo guidance and template update
  - repo changes must commit
- 내가 추가한 누락 방지 항목:
  - local environment evidence (`file`, `otool`, `java_home`)
  - skill validation

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - JVM 부팅 전체 구현을 소스 레벨로 끝까지 추적
  - Spring Boot loader 내부 코드를 line-by-line으로 모두 해설
- 지금 하지 않는 이유:
  - 이번 문서의 중심은 `java -jar Spring Boot executable jar startup path`이며, 더 깊은 launcher source walkthrough는 별도 문서로 나누는 편이 좋기 때문

## 2. Root-First Framing

- 근본 문제:
  - 현재 문서는 구조는 좋아졌지만, "셸/OS는 실제로 뭘 하는가", "spec artifact를 왜 믿을 수 있는가"를 독자가 스스로 붙잡을 수 있는 발판이 아직 부족하다.
- 왜 이 문제가 지금 중요한가:
  - 이 저장소의 목표는 읽는 사람이 다시 올라가서 스펙/예제/구조를 반복 상기하며 자기 지식으로 만드는 것이다.
- 작업 목표:
  - lower-layer explanation, visible spec artifacts, early concrete examples를 더 앞에 배치하고 guidance로 일반화한다.
- 기대 이점:
  - 독자가 `java`, executable format, manifest, Boot jar layout을 눈으로 먼저 보고 그 다음 설명을 따라갈 수 있다.
- 이점이 닫혔다고 판단할 확인 기준:
  - current doc에서 shell/OS/spec/real jar evidence가 더 앞에 나오고, guidance에 같은 원리가 반영된다.
- 하드 제약 / 호환성 경계:
  - local machine is macOS, so local observed executable format is Mach-O. Linux readers still need ELF explanation as neighboring case.
- 성공 정의:
  - repo guidance update + current doc refactor + external skill update + validation + repo commit
- PARTIAL 조건:
  - current doc만 고치고 guidance/skill 반영이 빠지면 PARTIAL
- BLOCKED 조건:
  - external skill validation failure or repo write blocker

## 3. Reader & Internalization Contract

- 주 독자:
  - 미래의 나
  - low-level runtime과 packaging 규약을 처음부터 이해하고 싶은 개발자
- 독자가 이미 알고 있다고 가정하는 것:
  - Java와 Spring Boot라는 이름 정도
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - 셸이 명령을 받아 실행 파일을 찾고, OS가 platform executable을 시작하고, launcher가 manifest를 읽고, Boot loader가 classpath를 구성하는 흐름
- 사용자가 내재화해야 할 사고 패턴:
  - 공식 규약, 실제 파일 구조, 실제 명령 결과를 눈으로 먼저 보고 설명을 따라간다
- 특히 막아야 하는 오해:
  - OS가 JAR 내부를 직접 이해한다고 생각하는 오해
  - `META-INF/MANIFEST.MF`가 임의 텍스트라고 보는 오해
  - 예제는 마지막에 확인하면 된다고 생각하는 오해
- 기억 anchor 후보:
  - `shell parsing -> PATH lookup -> platform executable(Mach-O/ELF) -> launcher -> manifest -> JarLauncher -> Start-Class -> SpringApplication.run`
- 반드시 거쳐야 하는 추상화 계층:
  - shell -> OS executable loader -> JVM launcher -> JAR spec -> Spring Boot loader -> SpringApplication
- 핵심 대조쌍 / 혼동쌍:
  - Mach-O vs ELF
  - launcher binary vs JVM runtime library
  - `Main-Class` vs `Start-Class`
  - spec artifact vs implementation detail
- 질문형으로 먼저 답해야 하는 핵심 질문:
  - 기계인 운영 체제가 어떻게 `java`를 실행하고, 왜 JAR 규약과 Spring Boot 규약을 그 뒤 계층에서 처리하는가
- 목차 필요 여부와 이유:
  - 필요. lower-layer부터 올라가는 구조를 독자가 한눈에 봐야 함
- `개요` 또는 서두에서 먼저 고정할 문장:
  - `java -jar SpringBootApp.jar`를 입력했을 때 shell, OS, launcher, manifest, Boot loader, SpringApplication이 각각 무엇을 하는지 정리한다
- 사용자 / 저장소의 말투, 표기, 목록 형식 선호:
  - direct prose
  - closed contrast는 번호 목록 + 4 space prose
  - 애매한 지시어 금지
- 바로 풀어써야 하는 지시어 / 애매한 연결 표현:
  - `그 사이`
  - `이 과정`
  - `이 구조`
  - `OS가 실행`
- 이번 작업의 품질 기준 exemplar:
  - 사용자 피드백 자체
- 현재 저장소 평균 품질을 기준으로 삼으면 안 되는 이유:
  - 이번 작업은 explanation floor를 더 올리는 작업이기 때문

## 4. Depth Decision

- 선택한 깊이: `full`
- 왜 이 깊이가 맞는가:
  - repo guidance와 skill까지 sync해야 하며, current doc 구조도 앞쪽을 재배치해야 한다
- 전체 루프를 켜야 하는 트리거:
  - lower-layer explanation and spec-first anchors are load-bearing
- 축약 가능한 섹션과 그 근거:
  - large external research unnecessary; local evidence + official links suffice

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - prose-first
  - immediate explanation
  - replay and verification path
  - direct prose
- 특히 중요한 규칙:
  - 사용자 피드백은 active style contract
- 전역 규칙과의 충돌 여부 / 해소:
  - 없음

## 6. Topic Analysis

- 현재 이해한 사용자 의도:
  - 단순히 읽히는 문서보다, 공식 규약과 실제 파일/구조를 눈으로 계속 상기할 수 있는 친절한 학습 문서를 원한다.
- 현재 보이는 문제 구조:
  - shell/OS 층 설명이 아직 얕다
  - spec artifact가 늦게 나온다
  - real example timing이 늦다
- 핵심 경계:
  - lower-layer explanation은 현재 문서의 질문을 닫는 범위까지만
- 숨은 가정 / 불확실성:
  - local environment evidence is macOS-specific, so doc must distinguish local observation from cross-platform principle
- 성공을 오판하기 쉬운 지점:
  - 설명을 더 길게 쓰는 것과 더 친절하게 쓰는 것을 혼동하는 것

## 7. Analysis Critique + Repair

### 7.1 Challenge 1

- 어떤 점이 틀릴 수 있는가:
  - ELF를 전면에 두면 현재 macOS 관측과 충돌할 수 있다
- 보강안:
  - executable format general concept를 먼저 설명하고, Linux는 ELF, macOS는 Mach-O라고 함께 적는다
- 왜 이 보강안이 더 강한가:
  - 원리와 현재 환경을 동시에 닫는다

### 7.2 Challenge 2

- 어떤 점이 좁거나 누락될 수 있는가:
  - `libjvm` 언급만 하고 dynamic loading이 무슨 뜻인지 안 풀 수 있다
- 보강안:
  - launcher binary와 runtime library를 구분하고, "실행 중 필요한 라이브러리를 프로세스 안으로 연결해 사용하는 것" 정도로 직접 풀어 쓴다
- 왜 이 보강안이 더 강한가:
  - jargon repetition을 줄인다

### 7.3 Challenge 3

- 어떤 점이 설명/학습 목표를 놓칠 수 있는가:
  - 공식 링크를 넣더라도 예제가 뒤에 나오면 학습 anchor가 늦다
- 보강안:
  - manifest sample과 jar layout sample을 앞쪽에 옮긴다
- 왜 이 보강안이 더 강한가:
  - 독자가 먼저 구조를 눈으로 보고 뒤 설명을 따라갈 수 있다

### 7.4 Retained Framing

- 최종 채택한 문제 정의:
  - lower-layer 친절성과 spec/structure visibility를 높이는 방향으로 guidance와 current doc를 함께 보강
- 폐기한 문제 정의와 이유:
  - current doc 안에서만 몇 문장 추가: root cause closure 부족

## 8. Scope Expansion & Impact Sync

- 시작 키워드:
  - shell parsing, PATH lookup, executable format, Mach-O, ELF, manifest spec, libjvm, dynamic loading
- 확장 키워드(동의어 / 약어 / 원어 / 구명칭 / 관련 에러 / 표준 번호):
  - argv, launcher, JAR File Specification, Main-Class, META-INF/MANIFEST.MF, Mach-O, PE, shared library
- 조사한 경로:
  - repo guidance
  - current doc
  - skill docs
  - local command evidence
- 함께 점검한 자산:
  - `/usr/bin/java`
  - `java.home`
  - `libjli.dylib`
  - `libjvm.dylib`
- 함께 움직여야 하는 표면:
  - AGENTS
  - WORK template
  - current doc
  - external skill
- 한쪽만 바꾸면 깨질 부분:
  - current doc만 바꾸면 later docs repeat the same issue
- 제외 표면과 근거:
  - unrelated repo docs: this is a style-contract round, not a repo-wide rewrite

## 9. Evidence Gathering

### 9.1 Evidence Ledger

- E-01
  - 주장: local environment uses a platform-native launcher binary, not a JAR-aware OS loader
  - 근거 유형: `command result`
  - 자료:
    - `command -v java`
    - `file /usr/bin/java`
    - `otool -L /usr/bin/java`
  - 이 자료로 닫힌 것:
    - local `java` is a Mach-O executable on macOS
  - 아직 비어 있는 것:
    - Linux ELF case is still conceptual / external-reference-level
- E-02
  - 주장: JVM runtime libraries are separate loadable libraries under the JDK
  - 근거 유형: `command result`
  - 자료:
    - `/usr/libexec/java_home`
    - `find $(/usr/libexec/java_home) -name 'libjvm.dylib' -o -name 'libjli.dylib'`
  - 이 자료로 닫힌 것:
    - local runtime library presence
  - 아직 비어 있는 것:
    - deep implementation details beyond doc scope
- E-03
  - 주장: user wants spec artifact and real structure shown earlier as learning anchors
  - 근거 유형: `repo evidence`
  - 자료:
    - user feedback
    - current doc ordering
  - 이 자료로 닫힌 것:
    - new structural priority
  - 아직 비어 있는 것:
    - exact wording and ordering

## 10. Evidence Critique + Repair

- 소스 품질 리스크:
  - local evidence is platform-specific
- 오래되었을 가능성이 있는 가정:
  - none material
- 빠진 대안 또는 빠진 근거:
  - official spec links for manifest and executable-jar structure should be explicit
- 근거 세트를 어떻게 보강했는가:
  - local commands + existing official links + new doc restructuring
- 보강 후에도 남는 한계:
  - launcher internals are still summarized, not source-traced

## 11. Design

- 선택한 접근:
  - repo guidance에 `spec artifact를 앞에 보여주기`, `lower-layer why/how 닫기`, `local observation vs platform-general rule 구분` 추가
  - current doc에 `셸이 명령을 해석하는 과정`, `OS executable format`, `java launcher and runtime libs`, `manifest spec`, `real jar structure early` 삽입
  - external skill에도 same rules 반영
- 왜 이 접근이 근본 문제와 이점에 가장 잘 맞는가:
  - 친절함을 말투가 아니라 설명 장치와 순서로 구현한다
- 고려한 대안:
  - current doc only
  - add footnotes only
- 대안을 채택하지 않은 이유:
  - both are too weak for recurrence prevention
- 문서 / 예제 / 자산 구조:
  - 개요
  - 실제로 먼저 볼 구조
  - shell/OS
  - manifest spec
  - Boot layout
  - launcher and SpringApplication
  - ready state
- 설명 뼈대: `질문형 | 계층형 | 비교형 | 시나리오형 | 혼합`
- 계층별 설명 순서:
  - visible artifact -> shell -> OS executable loader -> launcher -> manifest spec -> Boot loader -> SpringApplication
- 넣을 구체 예시 / 관측 anchor:
  - `file /usr/bin/java`
  - `otool -L /usr/bin/java`
  - `libjvm.dylib`
  - manifest sample
  - jar layout sample
- 이 문서를 끌어올릴 목표 수준:
  - 읽는 사람이 다시 위로 올라가 구조를 상기하며 따라갈 수 있는 문서
- 실패 모드:
  - too much platform detail
  - structure becomes scattered
- 검증 경로:
  - diff review
  - skill validation

## 12. Design Critique + Repair

### 12.1 Architect View

- 반론:
  - current doc may become too broad
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - only include lower-layer detail that directly explains the startup path

### 12.2 Domain / API Consumer View

- 반론:
  - platform-specific local evidence may confuse Linux readers
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - label local observation explicitly and pair it with Linux ELF note

### 12.3 Newcomer / Learner View

- 반론:
  - too many new nouns at once
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - show the artifact first, then define each term at first mention

### 12.4 Final Design Decision

- 최종 채택:
  - guidance + current doc + skill synchronized refactor
- 트레이드오프:
  - more up-front context, but much stronger re-derivability

## 13. Overall Plan

- 작업 순서:
  - WORK freeze
  - guidance patch
  - current doc reorder and rewrite
  - skill patch
  - validation / commit
- 선행 의존성:
  - none
- validation order:
  - doc review
  - skill validate
  - scoped git review
- rollback / retry / staging 필요 여부와 이유:
  - single commit after scoped review

## 14. Plan Critique + Repair

- 계획이 실패할 수 있는 지점:
  - shell/OS section can overwhelm the doc
- 순서상 위험:
  - moving examples too early may expose terms before definition
- 빠진 prerequisite:
  - none
- 보강안:
  - introduce structure first, then define its terms immediately
- 왜 보강된 계획이 더 나은가:
  - keeps both visibility and comprehension

## 15. Detailed Task Plan

- 수정 / 생성 / 검토할 파일:
  - create: this WORK file
  - update: `/Users/rody/VscodeProjects/study/AGENTS.md`
  - update: `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - update: `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - update: `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - update: `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- 각 파일에서 바꿀 논리 또는 구조:
  - AGENTS: spec-first / visible-artifact-first / lower-layer clarification rules
  - template: hooks for spec artifacts and earlier examples
  - current doc: reorder and expand lower layers
  - skill: same rules
- 관련 문서 동기화 계획:
  - current doc should satisfy the new rules immediately
- 예제 추가 / 보강 계획:
  - move manifest and jar structure evidence earlier
- 근거 섹션 반영 계획:
  - local observations and official links together

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: shell/OS execution path is materially clearer
  - S2: manifest and Boot jar structure appear early enough to anchor later reading
  - S3: guidance now tells future docs to show load-bearing artifacts and official specs early
- 실패 케이스 최소 3개:
  - F1: only more words, no better ordering
  - F2: local platform details are stated as universal facts
  - F3: official spec artifacts remain too late
- 회귀 위험:
  - doc bloat
  - structure fragmentation
- 회귀 방지 확인 경로:
  - scoped diff review
  - skill validation

### 15.2 Code / Doc Quality Review Points

- 단순성:
  - only keep lower-layer detail that closes the current question
- 응집도:
  - examples should serve the startup path, not branch into separate topics
- 확장 여지:
  - later deeper docs can branch from launcher internals or jar spec
- 과한 일반화 여부:
  - distinguish macOS local observation from Linux ELF explanation
- 설명 누락 위험:
  - shell tokenization / executable format / manifest spec / example timing

## 16. Detailed Plan Critique + Repair

- 누락된 케이스:
  - numbering without prose indentation
- fuzzy success criteria:
  - `친절함` is abstract
- scope overreach / under-specification:
  - too much JVM implementation would exceed scope
- 보강안:
  - keep a strict rule: if a lower-layer detail does not help explain the startup path, leave it out
- 최종 상세 계획:
  - rewrite current doc to show artifacts and specs earlier, while keeping the main question coherent

## 17. Frozen Checklist

- [x] 새 피드백을 최소 바닥선으로 유지한다
- [x] repo guidance에 visible artifact / spec-first / lower-layer clarification rules를 추가한다
- [x] WORK template에 같은 review hooks를 추가한다
- [x] `spring_boot_jar_startup.md`에서 shell/OS/spec/examples ordering을 개선한다
- [x] local observation과 platform-general rule을 구분한다
- [x] external skill에도 같은 guidance를 반영한다
- [x] skill validation을 수행한다
- [x] repo scoped review와 commit으로 closure를 닫는다

## 18. Execution Log

- local command evidence 수집:
  - `command -v java` -> `/usr/bin/java`
  - `file /usr/bin/java` -> Mach-O executable 확인
  - `otool -L /usr/bin/java` -> JavaLaunching.framework 의존성 확인
  - `/usr/libexec/java_home` -> JDK home 확인
  - `find $(/usr/libexec/java_home) -name 'libjli.dylib' -o -name 'libjvm.dylib'` -> runtime library 위치 확인
- repo guidance update:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
- current doc refactor:
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - actual jar structure와 `MANIFEST.MF`를 앞쪽으로 이동
  - shell parsing / executable format / dynamic loading / manifest spec 설명 보강
- external skill update:
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
- actual run verification:
  - sample jar 실행 후 `Tomcat initialized`, `Tomcat started`, `Started FullyPortableApplication` 로그 확인

## 19. Verification Plan

- current doc structure review
- guidance diff review
- skill validation
- scoped git review

## 20. Final Audit

- 상태 정직성:
  - COMPLETE로 표기 가능. guidance, template, doc, external skill, validation, commit까지 닫힘
- 근거 우선:
  - official docs + local command evidence + actual sample jar run으로 주요 주장 뒷받침
- 설명 계약:
  - load-bearing artifact를 초반에 배치했고, shell/OS 질문을 한 단계 아래까지 내려가 닫음
- local vs general:
  - 현재 머신의 Mach-O 관측과 Linux ELF 일반 설명을 분리함
- 남은 리스크:
  - JVM launcher 내부 구현을 source-level로 끝까지 추적하는 문서는 아직 별도 작업으로 남음
