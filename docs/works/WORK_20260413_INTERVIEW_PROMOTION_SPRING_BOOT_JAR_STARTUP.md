# WORK_20260413_INTERVIEW_PROMOTION_SPRING_BOOT_JAR_STARTUP

## 0. Meta

- 작업 제목: `.tmp/interviews2.md`의 `java -jar SpringBootApp.jar` 실행 과정을 정식 학습 문서로 승격
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260413_INTERVIEW_PROMOTION_SPRING_BOOT_JAR_STARTUP.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `research | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청: `.tmp/interview*.md 중 하나를 골라 정식 학습 문서로 승격`
- 원문 사용자 요청: `.tmp/interview*.md 중 하나를 골라 정식 학습 문서로 승격`
- 대상 경로 / 자산:
  - source reservoir: `/Users/rody/VscodeProjects/study/.tmp/interviews2.md`
  - promoted doc: `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - nearby entrypoint: `/Users/rody/VscodeProjects/study/jvm/spring/spring_init.md`
- 실행자: Codex
- 시작 일시: 2026-04-13
- 종료 일시: 2026-04-13
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal: interview raw note 한 cluster를 정식 학습 문서로 승격해, 나중에 다시 읽었을 때 `java -jar`부터 Spring Boot ready 상태까지를 스스로 설명할 수 있게 만든다.
- refs:
  - `/Users/rody/VscodeProjects/study/.tmp/interviews2.md`
  - `/Users/rody/VscodeProjects/study/jvm/spring/spring_init.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/HowWorks/app/src/main/kotlin/spring/SpringMain.kt`
  - official docs and primary sources about `execve`, Java launcher, Spring Boot executable jar, SpringApplication lifecycle
- scope:
  - source cluster 선택 및 정제
  - 정식 문서 작성
  - `jvm/spring` 내 진입 링크 추가
  - 문서형 검증 및 commit
- mode: `execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - raw note의 핵심 질문은 유지
  - shallow summary가 아니라 deep study monograph로 승격
  - 작은 확인 경로와 실제 관측 anchor 포함
- extra_checks:
  - 기존 `spring_init.md`와 역할 중복을 피할 것
  - `.tmp` 문체를 그대로 답습하지 말 것

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - `.tmp/interview*.md` 계열에서 하나를 골라 정식 학습 문서로 승격
- 사용자가 명시한 금지 사항:
  - 없음
- path / naming / format / finish 관련 요구:
  - study 저장소 스타일의 깊은 학습 문서
  - repo 변경 작업이므로 기본 closure는 검수 + 검증 + commit
- 내가 추가한 누락 방지 항목:
  - 승격 source 선택 근거 기록
  - 기존 spring 문서와의 연결 동선 추가
  - 사실 / 추론 / 검증 경로 구분

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - `.tmp/interviews2.md` 전체 정리
  - `interviews/` 전체 통합 정리
  - Spring Boot 전체 아키텍처 문서 재구성
- 지금 하지 않는 이유:
  - 이번 요청은 raw material에서 한 cluster를 골라 정식 자산으로 승격하는 1회 promotion이기 때문

## 2. Root-First Framing

- 근본 문제:
  - `.tmp/interviews2.md`에는 좋은 재료가 있지만 raw note 상태라, 나중에 다시 이해를 복원하거나 다른 문서에서 참조하기에 너무 넓고 거칠다.
- 왜 이 문제가 지금 중요한가:
  - `study-explanation`을 실제 작업에 투입해 승격 루프를 검증하려면, raw note -> 정식 문서 promotion의 첫 사례가 필요하다.
- 작업 목표:
  - `java -jar Spring Boot` 실행 과정을 운영체제 -> Java launcher -> Boot loader -> SpringApplication -> ready state까지 자연스럽게 이어지는 정식 학습 문서로 만든다.
- 기대 이점:
  - 미래의 내가 `왜 Spring Boot jar가 그냥 실행되는가`, `실제로 어디서 main 이 결정되는가`, `언제 ready 라고 볼 수 있는가`를 다시 복원할 수 있다.
- 이점이 닫혔다고 판단할 확인 기준:
  - 새 문서가 source note보다 질문 범위가 더 선명하고, 기존 `spring_init.md`와 역할이 분리되며, 작은 검증 경로를 제공한다.
- 하드 제약 / 호환성 경계:
  - 이 문서는 일반 Spring 입문서가 아니라 `java -jar` 기반 실행 경로 설명 문서여야 한다.
  - Spring Boot 버전 세부 구현 차이는 공식 문서와 repo evidence 수준에서만 단정한다.
- 성공 정의:
  - 새 문서 1개 + 관련 진입 링크 + WORK ledger + 검증 + commit
- PARTIAL 조건:
  - 문서를 썼지만 공식 근거나 진입 링크, 검증 경로가 비어 있으면 PARTIAL
- BLOCKED 조건:
  - load-bearing claim에 필요한 공식 근거를 확보할 수 없거나, target 구조를 안전하게 정할 수 없으면 BLOCKED

## 3. Reader & Internalization Contract

- 주 독자: 미래의 나, 그리고 Spring Boot jar 실행 과정을 단순 “그냥 main 호출” 이상으로 이해하고 싶은 개발자
- 독자가 이미 알고 있다고 가정하는 것:
  - Java / JAR / Spring Boot를 이름 수준으로는 알고 있음
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - `java -jar app.jar` 입력 이후 어떤 계층을 지나 어떤 준비 상태가 되는지
- 사용자가 내재화해야 할 사고 패턴:
  - 겉으로 단순한 명령도 OS, launcher, packaging, runtime, framework 계층이 연결되어 있음을 본다
- 특히 막아야 하는 오해:
  - `java -jar`가 곧바로 app main 클래스를 실행한다고 생각하는 오해
  - Spring Boot fat jar와 일반 executable jar를 동일한 구조로 보는 오해
  - context refresh와 ready 상태를 같은 시점으로 보는 오해
- 기억 anchor 후보:
  - `셸 -> execve -> java launcher -> Main-Class -> JarLauncher -> main -> SpringApplication.run -> ready`
- 반드시 거쳐야 하는 추상화 계층:
  - shell / OS / Java launcher / JAR manifest / Spring Boot loader / SpringApplication lifecycle
- 핵심 대조쌍 / 혼동쌍:
  - `fork` vs `execve`
  - `Main-Class` vs 애플리케이션 main class
  - `context refreshed` vs `ready to accept traffic`
- 질문형으로 먼저 답해야 하는 핵심 질문:
  - `java -jar SpringBootApp.jar`를 치면 실제로 누가 무엇을 읽고 어디서부터 애플리케이션이 뜨는가
- 이번 작업의 품질 기준 exemplar:
  - `/Users/rody/VscodeProjects/study/git/git_rebase.md`
  - `/Users/rody/VscodeProjects/study/computer_architecture/threads/threads.md`
  - `/Users/rody/VscodeProjects/study/jvm/java/java_synchronized.md`
- 현재 저장소 평균 품질을 기준으로 삼으면 안 되는 이유:
  - raw note나 얕은 문서는 기준선이 아니라 개선 대상이기 때문

## 4. Depth Decision

- 선택한 깊이: `full`
- 왜 이 깊이가 맞는가:
  - raw material conversion이고, 여러 계층을 연결해야 하며, 잘못 요약하면 장기 학습 비용이 커진다.
- 전체 루프를 켜야 하는 트리거:
  - 공식 근거 + repo evidence + explanation contract를 함께 닫아야 함
- 축약 가능한 섹션과 그 근거:
  - code change나 test surface는 없으므로 문서형 검증 중심으로 축약 가능

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - study 저장소 목적은 deep reusable learning asset 생성
  - `.tmp`는 raw material이며 exemplar가 아님
  - repo-changing doc work는 검수 + 검증 + commit으로 닫힘
- 특히 중요한 규칙:
  - prose-first
  - replay / verification path
  - fact / inference separation
  - current repo average를 기준으로 삼지 않음
- 전역 규칙과의 충돌 여부 / 해소:
  - 충돌 없음. 전역 rigor와 로컬 학습 품질 규칙이 같은 방향

## 6. Topic Analysis

- 현재 이해한 사용자 의도:
  - raw interview note의 일부를 진짜 학습 자산으로 승격하면서, `study-explanation`의 실전 품질을 높이고 싶다.
- 현재 보이는 문제 구조:
  - source note는 범위가 넓고 반복도 있으며, OS/JVM 설명 일부는 과도하게 일반론적이다.
- 핵심 경계:
  - 문서 중심은 `Spring Boot jar startup path`
  - JVM 전체 메모리 모델의 세부는 필요한 만큼만 넣는다
- 숨은 가정 / 불확실성:
  - Java launcher와 Spring Boot loader 세부 구현은 버전에 따라 다를 수 있으므로, 문서에는 안정적인 경계 위주로 적어야 한다.
- 성공을 오판하기 쉬운 지점:
  - source 내용을 많이 옮기는 것을 깊이로 착각하는 것

## 7. Analysis Critique + Repair

### 7.1 Challenge 1

- 어떤 점이 틀릴 수 있는가:
  - `fork()` 설명을 너무 크게 전면에 두면 실제 `java -jar` 경로의 핵심보다 process theory 설명이 더 커질 수 있다.
- 보강안:
  - POSIX shell 관점에서는 보통 `fork + execve`, 이미 다른 프로세스에서 프로그램을 직접 실행하는 관점에서는 `execve`가 핵심이라고 구분한다.
- 왜 이 보강안이 더 강한가:
  - 실제 실행 경로의 핵심을 흐리지 않는다.

### 7.2 Challenge 2

- 어떤 점이 좁거나 누락될 수 있는가:
  - Spring Boot loader와 ready state 사이의 차이를 충분히 닫지 못할 수 있다.
- 보강안:
  - `Main-Class=JarLauncher`와 `Start-Class` 역할을 분리해 설명하고, `ApplicationStartedEvent`와 `ApplicationReadyEvent` 차이를 넣는다.
- 왜 이 보강안이 더 강한가:
  - 사용자가 실제 로그 시점과 내부 준비 시점을 더 정확히 연결할 수 있다.

### 7.3 Challenge 3

- 어떤 점이 설명/학습 목표를 놓칠 수 있는가:
  - 문서가 단순 절차 나열로 끝날 수 있다.
- 보강안:
  - 작은 확인 실험과 `jar tf`, `unzip -p`, startup log 관측 경로를 포함한다.
- 왜 이 보강안이 더 강한가:
  - replay 가능성이 생긴다.

### 7.4 Retained Framing

- 최종 채택한 문제 정의:
  - `java -jar Spring Boot executable jar`의 실행 경로를 최소한의 OS 배경부터 Spring Boot ready 상태까지 연결하는 deep monograph
- 폐기한 문제 정의와 이유:
  - `JVM 시작과 메모리 구조 전부 설명`: 범위 과대, 이번 핵심 질문보다 넓음

## 8. Scope Expansion & Impact Sync

- 시작 키워드:
  - `java -jar`, `Spring Boot`, `JarLauncher`, `SpringApplication.run`
- 확장 키워드(동의어 / 약어 / 원어 / 구명칭 / 관련 에러 / 표준 번호):
  - `execve`, `Main-Class`, `Start-Class`, `BOOT-INF`, `ApplicationReadyEvent`, `ContextRefreshedEvent`, `fat jar`, `executable jar`
- 조사한 경로:
  - `/Users/rody/VscodeProjects/study/.tmp/interviews2.md`
  - `/Users/rody/VscodeProjects/study/jvm/spring/*.md`
  - `/Users/rody/VscodeProjects/study/jvm/examples/HowWorks/.../SpringMain.kt`
  - official docs and man page
- 함께 점검한 자산:
  - existing `spring_init.md`
  - existing `java_jar.md`
  - repo example code comments
- 함께 움직여야 하는 표면:
  - promoted doc
  - spring entrypoint doc link
  - WORK ledger
- 한쪽만 바꾸면 깨질 부분:
  - 새 문서만 생기고 진입 링크가 없으면 discoverability가 약해짐
- 제외 표면과 근거:
  - `knowledge/cards`: narrow decision memo가 아니라 general study doc이므로 제외

## 9. Evidence Gathering

### 9.1 Evidence Ledger

- E-01
  - 주장: source raw note 중 `java -jar SpringBootApp.jar 실행 과정` cluster는 정식 문서 하나로 수렴 가능한 질문을 이미 갖고 있다.
  - 근거 유형: `repo evidence`
  - 자료: `/Users/rody/VscodeProjects/study/.tmp/interviews2.md`
  - 이 자료로 닫힌 것: promotion source 선정 근거
  - 아직 비어 있는 것: raw note의 과한 일반론 정제
- E-02
  - 주장: Spring Boot executable jar는 일반 JAR처럼 `Main-Class`만 앱 main으로 두지 않고 Boot loader 구조를 사용한다.
  - 근거 유형: `official doc`
  - 자료: Spring Boot executable jar format docs
  - 이 자료로 닫힌 것: `BOOT-INF` 구조와 loader 설명
  - 아직 비어 있는 것: ready state 시점
- E-03
  - 주장: `java -jar` 경로에서 새 프로그램 이미지를 실행하는 핵심 시스템 콜은 `execve`다.
  - 근거 유형: `official doc`
  - 자료: Linux `execve(2)` man page
  - 이 자료로 닫힌 것: OS level replacement semantics
  - 아직 비어 있는 것: shell이 실제로 어떤 방식으로 launch하는지의 구현별 차이
- E-04
  - 주장: Spring Boot는 context refreshed 이후 started/ready 이벤트를 분리한다.
  - 근거 유형: `official doc`
  - 자료: SpringApplication reference docs
  - 이 자료로 닫힌 것: started vs ready distinction
  - 아직 비어 있는 것: 세부 이벤트 전체를 문서에 얼마나 넣을지 선택
- E-05
  - 주장: repo 안에 `SpringApplication.run()` 이후 웹 서버/DispatcherServlet 준비 흐름을 해설한 보조 자산이 이미 있다.
  - 근거 유형: `repo evidence`
  - 자료: `/Users/rody/VscodeProjects/study/jvm/examples/HowWorks/app/src/main/kotlin/spring/SpringMain.kt`
  - 이 자료로 닫힌 것: repo-local explanation anchor
  - 아직 비어 있는 것: 새 문서에 어느 정도까지 녹일지

### 9.2 Source Conflicts / Gaps

- 충돌하는 근거:
  - 없음
- 아직 부족한 근거:
  - Java launcher 공식 문서 line reference 추가 필요
- 추론으로만 남는 항목:
  - 사용 중인 shell이 내부적으로 `fork + execve`를 어떻게 조합하는지의 세부는 shell 구현 차이로 남는다

## 10. Evidence Critique + Repair

- 소스 품질 리스크:
  - raw note는 그대로 믿으면 과도한 일반론을 가져올 수 있다
- 오래되었을 가능성이 있는 가정:
  - Spring Boot loader package/class details
- 빠진 대안 또는 빠진 근거:
  - `java -jar`의 launcher 규칙
- 근거 세트를 어떻게 보강했는가:
  - official docs / man page / repo example을 함께 사용
- 보강 후에도 남는 한계:
  - 버전별 세부 구현 전체를 다루지는 않음

## 11. Design

- 선택한 접근:
  - 새 standalone 문서를 `jvm/spring` 아래 추가하고, `spring_init.md`에서 심화 문서로 링크
- 왜 이 접근이 근본 문제와 이점에 가장 잘 맞는가:
  - `spring_init.md`의 역할을 보존하면서 startup path를 깊게 설명할 수 있다
- 고려한 대안:
  - `spring_init.md` 안에 그대로 확장
  - `jvm/java/java_jar.md` 쪽으로 이동
- 대안을 채택하지 않은 이유:
  - 첫 대안은 문서가 과도하게 비대해지고 역할이 섞인다
  - 둘째 대안은 Spring Boot 고유 loader/lifecycle 맥락이 약해진다
- 문서 / 예제 / 자산 구조:
  - 질문형 직답
  - 작은 mental model
  - 단계별 실행 경로
  - 확인 명령
  - 흔한 오해
  - 추가 읽기 링크
- 설명 뼈대: `질문형 | 계층형 | 시나리오형 | 혼합`
- 계층별 설명 순서:
  - shell / OS -> java launcher -> executable jar structure -> Boot loader -> app main -> SpringApplication -> ready state
- 넣을 구체 예시 / 관측 anchor:
  - `jar tf`
  - `unzip -p ... META-INF/MANIFEST.MF`
  - startup log
- 이 문서를 끌어올릴 목표 수준:
  - `git_rebase.md`의 시나리오성 + `threads.md`의 계층 연결 + `java_synchronized.md`의 first-principles
- 실패 모드:
  - too broad JVM theory
  - version-fragile details as facts
  - existing spring doc duplication
- 검증 경로:
  - heading / link structure review
  - factual cross-check against primary sources
  - repo diff review

## 12. Design Critique + Repair

### 12.1 Architect View

- 반론:
  - standalone 문서가 너무 잘게 쪼개질 수 있다
- 보강 또는 유지 결정:
  - 유지
- 이유:
  - 질문 단위가 충분히 독립적이고, `spring_init.md`에서 허브처럼 연결하면 분산 비용보다 검색성과 집중도가 더 좋다

### 12.2 Domain / API Consumer View

- 반론:
  - Spring Boot 전용으로 좁히면 일반 `java -jar` 학습과 분리될 수 있다
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - 문서 초반에 일반 executable jar와 Spring Boot executable jar의 차이를 비교해 다리 놓기

### 12.3 Newcomer / Learner View

- 반론:
  - OS, manifest, loader, lifecycle이 한 번에 나오면 너무 많다
- 보강 또는 유지 결정:
  - 보강
- 이유:
  - 가장 작은 mental model과 한 줄 흐름도부터 시작하도록 설계

### 12.4 Final Design Decision

- 최종 채택:
  - `spring_boot_jar_startup.md` 신설 + `spring_init.md` 연결
- 트레이드오프:
  - 파일 수는 늘지만 문서 역할 분리가 더 명확해진다

## 13. Overall Plan

- 작업 순서:
  - WORK freeze
  - new doc 작성
  - entrypoint link 추가
  - review / verify
  - commit
- 선행 의존성:
  - source cluster / official evidence 확인
- validation order:
  - 내용 자체 점검 -> link/diff 점검 -> git status
- rollback / retry / staging 필요 여부와 이유:
  - 문서 작업이므로 staging 불필요

## 14. Plan Critique + Repair

- 계획이 실패할 수 있는 지점:
  - 문서가 너무 길어져 핵심 질문을 잃을 수 있다
- 순서상 위험:
  - 링크부터 수정하면 target title/shape가 바뀔 수 있다
- 빠진 prerequisite:
  - Java launcher rule check
- 보강안:
  - 새 문서 본문 확정 후에만 링크 추가
- 왜 보강된 계획이 더 나은가:
  - dead link나 제목 mismatch를 줄인다

## 15. Detailed Task Plan

- 수정 / 생성 / 검토할 파일:
  - create: `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md`
  - update: `/Users/rody/VscodeProjects/study/jvm/spring/spring_init.md`
  - create: `/Users/rody/VscodeProjects/study/WORK_20260413_INTERVIEW_PROMOTION_SPRING_BOOT_JAR_STARTUP.md`
- 각 파일에서 바꿀 논리 또는 구조:
  - 새 문서: raw cluster를 deep monograph로 정제
  - `spring_init.md`: jar build section에서 심화 학습 링크 제공
- 관련 문서 동기화 계획:
  - role overlap 최소화
- 예제 추가 / 보강 계획:
  - manifest, jar structure, ready log 확인 명령 추가
- 근거 섹션 반영 계획:
  - 공식 문서 링크와 repo example anchor를 마지막에 정리

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: 독자가 `Main-Class`와 실제 app main class의 차이를 설명할 수 있다
  - S2: 독자가 `context refreshed`와 `ready` 차이를 설명할 수 있다
  - S3: 독자가 jar 내부에서 무엇을 확인할지 직접 따라 할 수 있다
- 실패 케이스 최소 3개:
  - F1: 문서가 JVM 메모리 일반론으로 퍼져 핵심 질문이 흐려진다
  - F2: Spring Boot loader를 일반 jar와 동일하게 설명한다
  - F3: 실행 명령은 있는데 PASS / FAIL 관측이 없다
- 회귀 위험:
  - `spring_init.md`와의 역할 중복
- 회귀 방지 확인 경로:
  - 두 문서 역할을 diff와 본문으로 교차 검토

### 15.2 Code / Doc Quality Review Points

- 단순성:
  - 작은 흐름도에서 시작할 것
- 응집도:
  - 문서 질문은 하나로 유지
- 확장 여지:
  - future doc `java -cp` / exploded jar / native image 비교로 확장 가능
- 과한 일반화 여부:
  - shell 구현 차이, Boot version details는 안정적 경계까지만
- 설명 누락 위험:
  - Main-Class / Start-Class / ready state

## 16. Detailed Plan Critique + Repair

- 누락된 케이스:
  - `fork`를 항상 발생한다고 단정하지 않기
- fuzzy success criteria:
  - “깊다”는 표현이 추상적일 수 있음
- scope overreach / under-specification:
  - JVM memory details 과다 확장 위험
- 보강안:
  - smallest question -> minimal model -> layered expansion을 유지
- 최종 상세 계획:
  - new doc를 중심으로 쓰되, raw note의 OS/JVM 설명은 핵심 질문을 닫는 범위만 남긴다

## 17. Frozen Checklist

- [x] source raw note cluster를 하나로 선택하고 선정 근거를 남긴다
- [x] target path와 역할을 기존 문서 구조 안에서 정당화한다
- [x] `.tmp` 문체를 그대로 복사하지 않고 deep study monograph로 재구성한다
- [x] 작은 확인 경로 또는 관측 anchor를 포함한다
- [x] load-bearing claim을 official doc / repo evidence / explicit inference로 구분한다
- [x] nearby entrypoint 문서에서 discoverability를 보강한다
- [x] 최종 self-review와 검증을 수행한다
- [x] commit으로 closure를 닫는다

## 18. Execution Log

- `.tmp/interviews2.md`의 `java -jar SpringBootApp.jar` cluster를 promotion target으로 선택
- `jvm/spring` 아래 standalone deep doc이 가장 자연스럽다고 판단
- `/Users/rody/VscodeProjects/study/jvm/spring/spring_boot_jar_startup.md` 신설
- `/Users/rody/VscodeProjects/study/jvm/spring/spring_init.md`에 심화 문서 진입 링크 추가
- raw note의 과도한 JVM 일반론은 줄이고, `Main-Class -> JarLauncher -> Start-Class -> SpringApplication.run -> readiness` 흐름으로 재구성
- 저장소 샘플 jar(`/Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar`)를 실제로 열어 manifest와 startup 로그를 관측

## 19. Verification Plan

- 문서 구조 검토:
  - 질문이 하나로 수렴하는지
  - 기존 문서와 역할 충돌이 없는지
- 사실 검토:
  - `execve`
  - Spring Boot executable jar structure
  - `ApplicationStartedEvent` / `ApplicationReadyEvent`
- repo 검토:
  - `git diff --stat`
  - 관련 파일 diff review

## 20. Final Audit

- 상태:
  - `COMPLETE`
- 검증 결과:
  - `java --help | sed -n '1,80p'`
    - `java [options] -jar <jarfile>.jar` 확인
  - `jar tf /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar | sed -n '1,80p'`
    - `org/springframework/boot/loader/...`와 `BOOT-INF/...` 구조 확인
  - `unzip -p /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar META-INF/MANIFEST.MF`
    - `Main-Class: org.springframework.boot.loader.launch.JarLauncher`
    - `Start-Class: com.example.portable.FullyPortableApplication`
  - `java -jar /Users/rody/VscodeProjects/study/jvm/examples/fully-portable-spring-boot/target/fully-portable-spring-boot-0.0.1-SNAPSHOT.jar --server.port=0`
    - `Tomcat started on port ...`
    - `Started FullyPortableApplication ...`
- 품질 판정:
  - 새 문서는 raw note를 정식 학습 문서로 승격했고, 기존 `spring_init.md`와 역할이 겹치지 않는다.
  - 일반 JAR과 Spring Boot executable jar의 차이, `Main-Class`와 `Start-Class`의 차이, started와 readiness의 차이를 모두 닫았다.
  - commit을 포함한 기본 closure까지 닫았다.
