# WORK_20260414_SPRING_BOOT_JAR_STARTUP_REFACTOR

## 0. Meta

- 작업 제목: `spring_boot_jar_startup.md` 구조/문장/근거 리팩토링
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260414_SPRING_BOOT_JAR_STARTUP_REFACTOR.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | research | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - `spring_boot_jar_startup.md` 문서를 리팩토링
- 원문 사용자 요청:
  - `spring_boot_jar_startup.md` 문서를 리팩토링
- 대상 경로 / 자산:
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_init.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar`
  - official primary sources about `java -jar`, JAR spec, Spring Boot executable jar, Spring application events
- 실행자: Codex
- 시작 일시: 2026-04-14
- 종료 일시: 2026-04-14
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - 현재 문서를 더 직접적이고 더 촘촘한 deep study monograph로 다시 써서, 독자가 `java -jar SpringBootApp.jar` 입력부터 ready 상태까지를 스스로 복원할 수 있게 만든다.
- refs:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/USECASE.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - previous WORK ledgers for the same document
  - local repo evidence and primary sources
- scope:
  - current document 구조 재검토
  - 필요한 사실 재검증
  - 본문 리팩토링
  - repo-scoped review
  - verify + commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - artifact를 초반에 먼저 보여 주는 구조
  - `OS -> java launcher -> manifest -> JarLauncher -> Start-Class -> SpringApplication.run -> readiness` 흐름
  - 사용자의 direct prose 선호
  - history / why / previous way 맥락
- extra_checks:
  - `/usr/bin/java`와 실제 JDK `bin/java`를 혼동하지 않도록 문장을 더 정확히 잡을 것
  - manifest 항목 설명 뒤 바로 연결 메커니즘이 보이게 만들 것

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - `spring_boot_jar_startup.md` 리팩토링
- 사용자가 명시한 금지 사항:
  - 없음
- path / naming / format / finish 관련 요구:
  - repo 변경 작업이므로 기본 closure는 검수 + 검증 + commit
- 내가 추가한 누락 방지 항목:
  - current-machine observation과 platform-general rule 구분 점검
  - actual sample jar evidence 재확인

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - Spring Boot 전체 문서군 전면 리팩토링
  - AGENTS/skill 재수정
  - `.tmp` 원본 재정리
- 지금 하지 않는 이유:
  - 이번 요청은 current document의 품질을 끌어올리는 single-document loop이기 때문

## 2. Root-First Framing

- 근본 문제:
  - 현재 문서는 사실과 구조가 많이 좋아졌지만, 일부 구간은 여전히 설명자의 진행 멘트나 중간 요약에 기대고 있고, artifact -> component -> connection -> mechanism 연결이 완전히 촘촘하지 않다.
- 왜 이 문제가 지금 중요한가:
  - 이 문서는 앞으로 `study-explanation`의 실전 표본 중 하나가 될 수 있으므로, 현재 수준에서 한 번 더 끌어올릴 가치가 크다.
- 작업 목표:
  - 문서의 구조, 문장, 인과 연결, lower-layer closure를 더 강하게 만들어 재독성과 teach-back 품질을 높인다.
- 기대 이점:
  - 독자가 작은 구성요소부터 큰 실행 흐름까지 빈틈 없이 따라갈 수 있다.
- 이점이 닫혔다고 판단할 확인 기준:
  - TOC와 개요가 범위를 고정하고
  - artifact 설명 후 곧바로 역할과 연결이 나오며
  - OS / launcher / manifest / JarLauncher / readiness 경계가 서로 섞이지 않고
  - 검증 예시와 공식 근거가 다시 맞아 있다.
- 하드 제약 / 호환성 경계:
  - 직접 확인하지 않은 사실은 관측처럼 쓰지 않는다.
  - 현재 머신 관측은 macOS 기준으로 적고, Linux 일반 규칙은 따로 분리한다.
- 성공 정의:
  - 문서 리팩토링 + 필요한 재검증 + WORK ledger + commit
- PARTIAL 조건:
  - 문서를 고쳤지만 핵심 근거나 검증이 다시 닫히지 않으면 PARTIAL
- BLOCKED 조건:
  - load-bearing claim을 지지하는 primary source 또는 repo evidence를 다시 확보하지 못하면 BLOCKED

## 3. Reader & Internalization Contract

- 주 독자:
  - 미래의 나
  - `java -jar`와 Spring Boot executable jar를 뿌리부터 이해하고 싶은 개발자
- 독자가 이미 알고 있다고 가정하는 것:
  - Java, JAR, Spring Boot를 이름 수준으로는 알고 있음
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - `java -jar`를 쳤을 때 누가 무엇을 이해하고, 어느 지점부터 Spring Boot 영역이 시작되는지
- 사용자가 내재화해야 할 사고 패턴:
  - 작은 artifact와 규약을 먼저 붙잡고, 그것들이 어떻게 이어져 상위 메커니즘을 만드는지 본다
- 특히 막아야 하는 오해:
  - 운영 체제가 JAR 내부 구조를 직접 이해한다고 생각하는 오해
  - `Main-Class`와 `Start-Class`가 같은 계층의 규칙이라고 생각하는 오해
  - `Started ...` 로그와 readiness를 같은 말로 생각하는 오해
- 기억 anchor 후보:
  - `셸 -> 운영 체제 -> java 실행 파일 -> Java launcher -> manifest -> JarLauncher -> Start-Class -> SpringApplication.run -> ApplicationReadyEvent`
- 반드시 거쳐야 하는 추상화 계층:
  - terminal/shell
  - OS executable loading
  - Java launcher
  - JAR / manifest spec
  - Spring Boot executable jar layout
  - Spring application startup events
- 핵심 대조쌍 / 혼동쌍:
  - `/usr/bin/java` vs 실제 JDK `bin/java`
  - executable file format vs JAR format
  - `Main-Class` vs `Start-Class`
  - `ApplicationStartedEvent` vs `ApplicationReadyEvent`
- 질문형으로 먼저 답해야 하는 핵심 질문:
  - `java -jar SpringBootApp.jar`를 쳤을 때 실제로 누가 무엇을 읽고, 언제부터 애플리케이션이 떴다고 볼 수 있는가
- 목차 필요 여부와 이유:
  - 필요. 여러 계층과 artifact를 넘나드는 문서라 독자가 진행 경로를 먼저 봐야 한다.
- `개요` 또는 서두에서 먼저 고정할 문장:
  - 이 문서는 `java -jar SpringBootApp.jar` 입력 이후 셸, 운영 체제, Java launcher, manifest, Spring Boot loader, Spring application startup이 어떻게 이어지는지 정리한다.
- 사용자 / 저장소의 말투, 표기, 목록 형식 선호:
  - 직접 진술형
  - 순차 흐름은 번호
  - 동등 비교는 불릿
  - 짧은 항목은 한 줄
- 이번 문서에서 특히 지켜야 할 설명 원칙:
  - artifact를 먼저 보여 주고, 각 요소를 설명한 뒤, 그 요소들이 어떻게 엮이는지 설명한다
  - 출처 인용에서 멈추지 않고 인과와 구조로 잇는다
  - 필요 설명을 뒤로 미루지 않는다
- 이번 문서에서 문장 단위로 점검할 판정 질문:
  - 이 문장은 현재 주제를 설명하는가, 아니면 문서의 동선만 말하는가
  - 이 문장은 독자의 이해를 실제로 한 걸음 전진시키는가
  - 이 인용 다음 문장이 현재 주제의 문제와 구조를 이어 주는가
- bad -> better 예시 페어(선택):
  - `이제 JAR 안으로 올라가 보겠습니다` -> `여기서는 JAR 안의 manifest와 Spring Boot loader 구조를 설명합니다`
  - `이 다섯 줄만 봐도 구조가 드러납니다` -> `각 항목이 무엇이고, 실행에서 어떤 역할을 하는지부터 보겠습니다`
- 바로 풀어써야 하는 지시어 / 애매한 연결 표현:
  - `이 사이`
  - `이 구조`
  - `이 단계`
  - `여기서`
- 초반에 먼저 보여줄 공식 규약 / 실제 구조 / 예시:
  - sample jar file list
  - manifest actual content
  - actual `java` binary observations
- 독자가 자연스럽게 묻게 될 한 단계 아래 질문:
  - 셸은 입력 줄을 실제로 어떻게 프로그램과 인자로 바꾸는가
  - 운영 체제는 어떤 파일 포맷을 읽고 실행하는가
  - Java launcher는 어디까지 알고 어디부터 Spring Boot loader에 넘기는가
- 독자가 자연스럽게 묻게 될 역사 / 등장 맥락 / 이전 방식 질문:
  - JAR는 왜 생겼는가
  - manifest는 왜 필요해졌는가
  - `java -jar`가 생기기 전 감각은 어땠는가
- 현재 머신 관측과 일반 규칙을 구분해야 하는 지점:
  - macOS Mach-O vs Linux ELF
  - `/usr/bin/java` vs current JDK home `bin/java`
- 먼저 분해해서 설명할 구성요소:
  - `META-INF/MANIFEST.MF`
  - `JarLauncher.class`
  - `BOOT-INF/classes/...Application.class`
  - `BOOT-INF/lib/spring-boot-*.jar`
  - `BOOT-INF/lib/spring-boot-autoconfigure-*.jar`
- 그 구성요소들이 어떻게 엮여 상위 메커니즘이 되는가:
  - Java launcher는 manifest의 `Main-Class`를 읽고
  - `Main-Class`가 `JarLauncher`이므로 Spring Boot loader가 먼저 시작되며
  - loader가 `BOOT-INF/classes`와 `BOOT-INF/lib`를 실행 가능한 classpath로 바꾸고
  - 그 뒤에 `Start-Class`의 `main()`이 호출되어 Spring Boot framework startup으로 넘어간다
- 번호 목록으로 써야 하는 순차 흐름:
  - shell input steps
  - manifest -> launcher -> loader flow
  - readiness event order
- 불릿 목록으로 써야 하는 동등 비교축:
  - Linux vs macOS executable format
  - 일반 실행형 JAR vs Spring Boot executable jar
- 한 줄로 닫을 항목과 들여쓴 prose가 필요한 항목:
  - shell 3-step summary는 한 줄
  - manifest / JarLauncher / readiness distinction은 들여쓴 prose 필요
- 공식 자료를 어떤 인과 흐름으로 설명에 연결할 것인가:
  - spec/source fact -> 왜 그 규약이 필요했는가 -> 현재 artifact에 어떤 필드/구조로 남는가 -> 현재 실행 흐름에서 어디에 쓰이는가
- 공식 자료 인용이 멈추면 FAIL로 볼 판정 지점:
  - 출처를 말하고도 현재 artifact나 실행 단계와 연결되지 않으면 FAIL
- 이번 작업의 품질 기준 exemplar:
  - `/Users/rody/VscodeProjects/study/algorithms/dynamic_programming.md`
  - `/Users/rody/VscodeProjects/study/computer_architecture/threads/threads.md`
  - `/Users/rody/VscodeProjects/study/git/git_rebase.md`
- 현재 저장소 평균 품질을 기준으로 삼으면 안 되는 이유:
  - 이 문서는 기준선을 끌어올리는 exemplar 후보이기 때문
- 속도를 이유로 줄이면 안 되는 조사 / 비교 / 검증 / 설명 단계:
  - local command verification
  - official source recheck
  - component -> mechanism rewrite
  - final reread

## 4. Depth Decision

- 선택한 깊이: `full`
- 왜 이 깊이가 맞는가:
  - 여러 계층과 primary source를 다시 맞춰야 하는 문서형 리팩토링이다.
- 전체 루프를 켜야 하는 트리거:
  - 문장 수정이 아니라 구조와 메커니즘 closure가 핵심이다.
- 축약 가능한 섹션과 그 근거:
  - guidance/skill 수정은 비범위이므로 현재 문서와 WORK 중심으로 닫는다.

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - deep study monograph
  - current average보다 exemplar 기준
  - repo-changing doc work는 verify + commit closure
- 특히 중요한 규칙:
  - direct prose
  - artifact early
  - history/context closure
  - current observation과 general rule separation
- 전역 규칙과의 충돌 여부 / 해소:
  - 충돌 없음

## 6. Topic Analysis

- 현재 이해한 사용자 의도:
  - 이미 좋아진 문서를 한 번 더 밀어 올려, 설명이 더 술술 읽히면서도 더 정확하고 단단해지게 하고 싶다.
- 현재 보이는 문제 구조:
  - 앞부분과 중간 일부가 아직 summary/restate 성격을 띤다.
  - `/usr/bin/java`와 실제 JDK `bin/java` 구분은 더 또렷하게 적을 여지가 있다.
  - `JarLauncher`와 `SpringApplication.run()` 경계 설명은 더 세밀하게 엮을 수 있다.
- 핵심 경계:
  - OS가 이해하는 것
  - Java launcher가 이해하는 것
  - Spring Boot loader가 이해하는 것
  - Spring framework가 시작되는 것
- 숨은 가정 / 불확실성:
  - macOS의 `/usr/bin/java` 동작 세부는 JDK/OS 버전에 따라 달라질 수 있으므로 관측 범위와 일반 원리를 분리해서 써야 한다.
- 성공을 오판하기 쉬운 지점:
  - 문장을 매끄럽게만 만들고 인과 연결은 그대로 두는 것

## 7. Analysis Critique + Repair

### 7.1 Challenge 1

- 어떤 점이 틀릴 수 있는가:
  - `/usr/bin/java`와 JDK `bin/java`를 같은 것으로 설명하면 현재 머신 관측이 부정확해질 수 있다.
- 보강안:
  - `/usr/bin/java`는 셸이 찾는 진입점, JDK home `bin/java`는 실제 JDK 실행 파일로 분리해서 적는다.
- 왜 이 보강안이 더 강한가:
  - 관측과 일반 원리를 섞지 않는다.

### 7.2 Challenge 2

- 어떤 점이 좁거나 누락될 수 있는가:
  - manifest 항목 설명이 개별 정의에 머물고 상위 연결이 약할 수 있다.
- 보강안:
  - manifest 항목 직후 `누가 읽고 -> 무엇을 결정하고 -> 다음 단계로 어떻게 넘어가는가`를 더 또렷한 번호 흐름으로 재작성한다.
- 왜 이 보강안이 더 강한가:
  - component explanation이 mechanism explanation으로 자연스럽게 이어진다.

### 7.3 Challenge 3

- 어떤 점이 설명/학습 목표를 놓칠 수 있는가:
  - `JarLauncher`와 `SpringApplication.run()`이 모두 “부팅”으로 뭉개질 수 있다.
- 보강안:
  - file structure bootstrap과 framework bootstrap을 나눠서 대조한다.
- 왜 이 보강안이 더 강한가:
  - reader가 계층 경계를 기억하기 쉬워진다.

### 7.4 Retained Framing

- 최종 채택한 문제 정의:
  - sample jar와 실제 실행 관측을 기반으로 `java -jar` 입력 이후의 계층별 bootstrap 경로를 뿌리부터 설명하는 문서로 다시 다듬는다.
- 폐기한 문제 정의와 이유:
  - 단순 문장 polish 중심 리팩토링: 이번 요청의 기대 수준보다 약함

## 8. Scope Expansion & Impact Sync

- 시작 키워드:
  - `java -jar`
  - `META-INF/MANIFEST.MF`
  - `JarLauncher`
  - `Start-Class`
  - `ApplicationReadyEvent`
- 확장 키워드(동의어 / 약어 / 원어 / 구명칭 / 관련 에러 / 표준 번호):
  - `execve`
  - `ELF`
  - `Mach-O`
  - `launcher`
  - `nested jar`
  - `BOOT-INF`
  - `ReadinessState.ACCEPTING_TRAFFIC`
- 조사한 경로:
  - target doc
  - `spring_init.md`
  - previous WORK ledgers
  - sample jar
  - current java binary observations
  - official docs
- 함께 점검한 자산:
  - nearby spring hub doc
  - sample executable jar
- 함께 움직여야 하는 표면:
  - target doc
  - WORK ledger
- 한쪽만 바꾸면 깨질 부분:
  - 문서만 바뀌고 근거 확인이 약하면 다시 흔들림
- 제외 표면과 근거:
  - AGENTS/skill은 이번 요청 비범위

## 9. Evidence Gathering

### 9.1 Evidence Ledger

- E-01
  - 주장:
    - sample jar manifest와 파일 구조는 `Main-Class`, `Start-Class`, `BOOT-INF/...`를 실제로 담고 있다.
  - 근거 유형: `repo evidence | command result`
  - 자료:
    - `jar tf ... | rg ...`
    - `unzip -p ... META-INF/MANIFEST.MF`
  - 이 자료로 닫힌 것:
    - artifact early section과 manifest section의 concrete example
  - 아직 비어 있는 것:
    - 없음
- E-02
  - 주장:
    - 현재 머신에서 셸이 찾는 진입점은 `/usr/bin/java`이고, 실제 JDK home에는 별도의 `bin/java`가 있다.
  - 근거 유형: `command result`
  - 자료:
    - `command -v java`
    - `file /usr/bin/java`
    - `/usr/libexec/java_home`
    - `file $JAVA_HOME/bin/java`
    - `otool -L $JAVA_HOME/bin/java`
  - 이 자료로 닫힌 것:
    - current observation vs general rule distinction
  - 아직 비어 있는 것:
    - `/usr/bin/java` 내부 경로 전체 메커니즘은 공식 애플 문서로 닫지 못했으므로 관측 범위만 적어야 함
- E-03
  - 주장:
    - JAR/manifest와 Spring Boot executable jar, SpringApplication event order는 공식 문서로 설명 가능하다.
  - 근거 유형: `official doc | standard`
  - 자료:
    - Oracle JAR spec / java command / JAR overview
    - Spring Boot executable jar spec / launching executable jars / SpringApplication reference
    - `execve(2)` / `elf(5)`
  - 이 자료로 닫힌 것:
    - why/history/spec/mechanism/readiness sections
  - 아직 비어 있는 것:
    - 없음

### 9.2 Source Conflicts / Gaps

- 충돌하는 근거:
  - 없음
- 아직 부족한 근거:
  - macOS `/usr/bin/java` stub/launcher의 OS-level 상세 메커니즘은 현재 primary source를 추가로 찾지 못했다.
- 추론으로만 남는 항목:
  - `/usr/bin/java`에서 JDK `bin/java`로 이어지는 세부 내부 경로 전부는 추론을 얹지 않고 관측 범위만 적는다.

## 10. Evidence Critique + Repair

- 소스 품질 리스크:
  - software docs는 버전별 차이가 있을 수 있다.
- 오래되었을 가능성이 있는 가정:
  - Spring Boot 3.3.4 sample jar에 맞는 구조를 너무 일반화하면 안 된다.
- 빠진 대안 또는 빠진 근거:
  - actual startup event order official doc 재확인 필요
- 근거 세트를 어떻게 보강했는가:
  - current machine command output + official primary sources를 함께 사용한다.
- 보강 후에도 남는 한계:
  - macOS launcher internals는 현재 필요한 범위까지만 적는다.

## 11. Design

- 선택한 접근:
  - 현재 문서의 큰 축은 유지하되, 앞부분과 중간의 summary/restate 구간을 줄이고 component -> connection -> mechanism 흐름을 강화한다.
- 왜 이 접근이 근본 문제와 이점에 가장 잘 맞는가:
  - 기존 장점은 보존하면서 읽힘과 정확성을 동시에 끌어올릴 수 있다.
- 고려한 대안:
  - 전면 구조 재구성
  - 부분 문장 polish만 수행
- 대안을 채택하지 않은 이유:
  - 전면 재구성은 이미 좋아진 흐름을 불필요하게 흔들 수 있고, 부분 polish만으로는 구조 문제가 남는다.
- 문서 / 예제 / 자산 구조:
  - current single doc 유지
  - sample jar evidence 유지
  - official links 유지/정리
- 설명 뼈대: `질문형 | 계층형 | 비교형 | 혼합`
- 계층별 설명 순서:
  - overview
  - visible artifacts
  - shell
  - OS executable loading
  - JAR/manifest history and spec
  - plain executable JAR vs Boot executable jar
  - JarLauncher bootstrap
  - SpringApplication bootstrap
  - readiness
- 넣을 구체 예시 / 관측 anchor:
  - sample jar file list
  - manifest actual output
  - current java binary observations
  - startup logs
- 이 문서를 끌어올릴 목표 수준:
  - study 저장소의 strong exemplar 후보
- 실패 모드:
  - 과도한 반복
  - 관측과 일반 규칙 혼합
  - 구조 설명 뒤 연결 메커니즘 약화
- 검증 경로:
  - reread
  - command evidence spot-check
  - startup run

## 12. Design Critique + Repair

### 12.1 Architect View

- 반론:
  - 여전히 section 수가 많아 장황해질 수 있다.
- 보강 또는 유지 결정:
  - 유지하되 각 section이 distinct question을 닫도록 문장을 압축한다.
- 이유:
  - 계층이 많은 주제라 section 자체는 필요하다.

### 12.2 Domain / API Consumer View

- 반론:
  - reader가 실제 실무 판단에 어떻게 쓰는지 약할 수 있다.
- 보강 또는 유지 결정:
  - ready section에서 `startup hook`, `health check`, `readiness probe`로 실무 연결을 유지한다.
- 이유:
  - 학습 문서이면서 재사용성도 보여 줘야 한다.

### 12.3 Newcomer / Learner View

- 반론:
  - lower-layer explanation이 여전히 빨라 보일 수 있다.
- 보강 또는 유지 결정:
  - `/usr/bin/java`와 JDK `bin/java`, executable format, shared library loading 설명을 더 또렷하게 쓴다.
- 이유:
  - 이 구간이 초심자에게 가장 막히기 쉽다.

### 12.4 Final Design Decision

- 최종 채택:
  - 중간 재구성형 리팩토링
- 트레이드오프:
  - 완전 재작성보다 안정적이지만, section 간 연결을 수동으로 더 세심하게 다듬어야 한다.

## 13. Overall Plan

- 작업 순서:
  - evidence recheck
  - rewrite plan
  - target doc patch
  - review
  - verify
  - commit
- 선행 의존성:
  - evidence recheck 먼저
- validation order:
  - document reread
  - command rerun as needed
  - sample jar startup
  - git diff review
- rollback / retry / staging 필요 여부와 이유:
  - targeted staging 필요. repo가 dirty할 가능성이 높다.

## 14. Plan Critique + Repair

- 계획이 실패할 수 있는 지점:
  - 문서 전체를 한 번에 고치다 보면 일부 좋은 문장을 해칠 수 있다.
- 순서상 위험:
  - 증거 확인 없이 먼저 rewrite하면 다시 factual wobble이 생길 수 있다.
- 빠진 prerequisite:
  - current startup log rerun
- 보강안:
  - patch 전 startup command도 다시 실행한다.
- 왜 보강된 계획이 더 나은가:
  - final evidence freshness가 높아진다.

## 15. Detailed Task Plan

- 수정 / 생성 / 검토할 파일:
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - `/Users/rody/VscodeProjects/study/WORK_20260414_SPRING_BOOT_JAR_STARTUP_REFACTOR.md`
- 각 파일에서 바꿀 논리 또는 구조:
  - target doc:
    - overview/direct prose 재정리
    - artifact explanation and connection 강화
    - `/usr/bin/java` vs JDK `bin/java` 명료화
    - JarLauncher / SpringApplication 경계 명료화
    - readiness 설명 밀도 보강
  - WORK:
    - execution / verification / audit 기록
- 관련 문서 동기화 계획:
  - `spring_init.md` 링크는 현재 상태 유지
- 예제 추가 / 보강 계획:
  - manifest actual output과 current java binary observation 유지
- 근거 섹션 반영 계획:
  - existing official links 유지하고 필요하면 보강

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: 독자가 `Main-Class`와 `Start-Class` 차이를 바로 설명할 수 있다.
  - S2: 독자가 OS가 이해하는 것과 Spring Boot loader가 이해하는 것을 구분할 수 있다.
  - S3: 독자가 `ApplicationStartedEvent`와 `ApplicationReadyEvent` 차이를 startup/readiness 관점에서 말할 수 있다.
- 실패 케이스 최소 3개:
  - F1: `/usr/bin/java`와 JDK `bin/java`가 같은 것처럼 읽힌다.
  - F2: manifest 항목 설명 뒤 연결 메커니즘이 약하다.
  - F3: `JarLauncher`와 `SpringApplication.run()`의 경계가 다시 흐려진다.
- 회귀 위험:
  - 기존에 좋아진 front matter 구조를 해칠 수 있음
- 회귀 방지 확인 경로:
  - before/after reread

## 16. Execution Log

- applicable stack 확인:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/USECASE.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
- target re-analysis:
  - current doc 전체 재독
  - previous WORK ledgers 재검토
  - `spring_init.md` 연계 확인
- local evidence refresh:
  - `jar tf ... | rg ...`
  - `unzip -p ... META-INF/MANIFEST.MF`
  - `command -v java`
  - `file /usr/bin/java`
  - `otool -L /usr/bin/java`
  - `file "$JAVA_HOME/bin/java"`
  - `otool -L "$JAVA_HOME/bin/java"`
  - sample jar actual startup run
- official source refresh:
  - Oracle JAR spec
  - Oracle `java` command docs
  - Oracle JAR overview
  - Spring Boot executable jar spec
  - Spring Boot launching docs
  - Spring Boot `SpringApplication` reference
  - `execve(2)` / `elf(5)`
- rewrite highlights:
  - 개요에 핵심 질문 2개를 직접 고정
  - artifact section에서 file list와 manifest를 모두 초반에 제시
  - `META-INF/MANIFEST.MF` / `JarLauncher` / `BOOT-INF/...` 역할 설명 뒤 곧바로 연결 메커니즘으로 전개
  - `/usr/bin/java`와 JDK home `bin/java`를 구분해 current-machine observation을 더 엄밀하게 정리
  - `JarLauncher`와 `SpringApplication.run()` 경계를 더 직접적으로 구분
  - readiness 구간에 Spring Boot event order를 더 세밀하게 추가
  - 실제 startup log를 최신 실행 결과로 갱신
  - 오래된 TOC anchor mismatch 수정

## 17. Verification

- 문서형 검수:
  - TOC, 개요, artifact-first 구조, component -> mechanism 연결, current observation vs general rule 구분을 재검토
- 명령 검증:
  - `command -v java` -> `/usr/bin/java`
  - `file /usr/bin/java` -> Mach-O
  - current JDK home `bin/java` -> Mach-O + `@rpath/libjli.dylib`
  - sample manifest -> `Main-Class`, `Start-Class`, `Spring-Boot-Classes`, `Spring-Boot-Lib`
  - sample jar file list -> `JarLauncher.class`, application class, boot libs 존재 확인
- 실행 검증:
  - `java -jar /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar --server.port=0`
  - `Tomcat initialized`
  - `Tomcat started on port ...`
  - `Started FullyPortableApplication ...`
- diff 검수:
  - `git diff --check -- ...` PASS
  - stale anti-pattern string search PASS

## 18. Final Audit

- 상태 정직성:
  - COMPLETE 가능. target doc rewrite, evidence refresh, verification, commit closure까지 닫을 수 있다.
- 사용자 요구 반영:
  - `spring_boot_jar_startup.md` 리팩토링 완료
- 이번 리팩토링에서 특히 강해진 점:
  - artifact를 더 앞에 두고 바로 풀어 쓴 점
  - `/usr/bin/java`와 JDK `bin/java`를 분리해 current observation을 더 정확하게 쓴 점
  - `JarLauncher`와 `SpringApplication.run()`의 담당 계층을 더 또렷하게 구분한 점
  - readiness를 이벤트 순서까지 내려가 설명한 점
- 남은 리스크:
  - macOS `/usr/bin/java` 내부 연결 메커니즘의 더 세세한 구현 설명은 현재 문서 범위 밖으로 남겨 두었다
