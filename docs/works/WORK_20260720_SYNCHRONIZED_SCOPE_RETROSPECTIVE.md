# WORK_20260720_SYNCHRONIZED_SCOPE_RETROSPECTIVE

## 0. Meta

- 작업 제목: `synchronized` 범위 오용과 JVM 동작 회고
- WORK 파일 경로: `docs/works/WORK_20260720_SYNCHRONIZED_SCOPE_RETROSPECTIVE.md`
- 저장소: `C:\Users\rody\WorkspacePrivate\study`
- 작업 유형: `research | analysis | explain | refactor_docs`
- 작업 깊이: `full`
- 원문 사용자 요청: 특정 프로젝트와 무관하게 넓은 `synchronized` 사용의 문제, JVM 수준의 이유,
  대안 선택 기준과 베스트 프랙티스를 사람이 이해하기 쉬운 회고 문서로 정리한다.
- 대상 경로 / 자산: `jvm/java/java_synchronized.md`, 최소 바이트코드 예제, `jvm/java/java.md`
- 실행자: Codex
- 시작 일시: 2026-07-20
- 종료 일시: 2026-07-20
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: `synchronized`가 실제로 무엇을 잠그며, 왜 넓은 인스턴스 monitor가 불필요한 직렬화와 잘못된
  안전감각을 동시에 만들 수 있는지 설명한다.
- refs: Java SE 25 JLS/JVMS, Java 25 API, `javac`·`javap` 실험.
- scope: intrinsic monitor, happens-before, reentrancy, bytecode, contention, compound invariant,
  `volatile`·Atomic·Lock·ReadWriteLock 선택 기준, 네트워크 서비스 shutdown 사례의 일반화.
- mode: 문서 신설과 검증 가능한 예제 추가.
- run_mode: `normal`
- finish: `test+commit`
- must_keep: 특정 회사·프로젝트·클래스 이름에 의존하지 않는 중립적 설명.
- extra_checks: 링크 유효성, Java 25 컴파일·`javap`, Markdown 공백 검사, 관련 index 연결.

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구: 잘못된 사용과 그 이유, JVM 수준 해설, 개선 방법과 선택 이유, 회고 맥락,
  STOPPING 이후 요청이 정상 경로에서는 도달 불가능하지만 방어가 필요한 조건.
- 사용자가 명시한 금지 사항: 특정 작업 프로젝트에 종속된 문서.
- path / naming / format / finish 관련 요구: `C:\Users\rody\WorkspacePrivate\study\jvm` 아래 Markdown.
- 내가 추가한 누락 방지 항목: 직접 재현 가능한 바이트코드 예제와 검증 명령.

### 1.2 Non-Goals

- HotSpot의 특정 버전별 lock word 비트 배치를 영구 규칙처럼 설명하지 않는다.
- 모든 `java.util.concurrent` 자료구조를 망라하지 않는다.

## 2. Root-First Framing

- 근본 문제: 공유 상태의 실제 불변식보다 넓은 객체 monitor를 선택하면 정확성과 성능을 동시에 오판한다.
- 작업 목표: 독자가 `synchronized`, `volatile`, Atomic, 명시적 Lock 중 하나를 근거 있게 선택하게 한다.
- 하드 제약 / 호환성 경계: Java 언어·VM 규격과 특정 JVM 구현 최적화를 분리한다.
- 성공 정의: 소스 문법부터 bytecode, monitor, JMM, 애플리케이션 경합까지 인과 관계가 이어지고 직접 검증 가능하다.
- PARTIAL 조건: 개념 설명은 있으나 실험 또는 선택 기준이 빠짐.
- BLOCKED 조건: 1차 규격 근거를 확인하지 못함.

## 3. Reader & Internalization Contract

- 주 독자: Java 동시성의 기본 문법은 보았지만 monitor와 JMM 연결이 아직 선명하지 않은 개발자.
- teach-back 목표: 인스턴스 `synchronized` 메서드들이 왜 같은 monitor에서 경쟁하는지, Atomic 하나가 왜
  여러 필드의 불변식을 자동으로 보호하지 않는지 자기 말로 설명할 수 있다.
- 특히 막아야 하는 오해: `synchronized`가 객체 전체의 모든 메서드를 멈춘다, CAS가 여러 공유 값을 한꺼번에
  안전하게 만든다, `volatile`이 `count++`를 원자화한다, lock-free가 항상 더 빠르다.
- 기억 anchor 후보: `무엇을 잠갔는가`, `어떤 상태를 함께 바꿔야 하는가`, `경쟁 중 누가 기다려야 하는가`.
- active recall 질문 / 작은 실험: `javap -c -v`로 method flag와 `monitorenter` 차이를 확인한다.
- 반드시 거쳐야 하는 추상화 계층: Java source -> class file -> JVM monitor -> JMM happens-before -> OS 대기 가능성
  -> 서비스 수준 불변식.
- 핵심 대조쌍 / 혼동쌍: method monitor vs block monitor, visibility vs atomicity, single variable vs compound invariant,
  mutual exclusion vs admission control.
- 목차 필요 여부와 이유: 장문 학습 문서이므로 질문과 실행 흐름을 복원할 수 있는 목차가 필요하다.
- 이번 작업의 품질 기준 exemplar: `computer_architecture/threads/threads.md`, `git/git_rebase.md`.
- primary exemplar와 참고할 설명 원리: 전자는 스레드의 작은 실행 단위에서 운영체제·하드웨어 층으로 연결하는
  방식을 참고한다. 특정 플랫폼 설명을 그대로 복제하지 않는다.
- secondary exemplar와 참고할 설명 원리: 후자는 오해하기 쉬운 상태 전이를 구체 예와 복구 가능한 명령으로
  설명하는 방식을 참고한다. Git 고유의 장문 명령 나열은 따르지 않는다.

## 4. Evidence & Design

- 1차 근거: JLS 8.4.3.6, 14.19, 17.1, 17.4.5, 17.5; JVMS 2.11.10과 6장의
  `monitorenter`·`monitorexit`; Java SE 25 Atomic·Lock API.
- 문서 구조: 작은 코드 예 -> 실제 monitor 식별 -> bytecode -> happens-before -> 오용 사례 -> 복합 경합 반례
  -> 대안 선택표 -> shutdown 경계 사례 -> 검증과 회고 체크리스트.
- 선택한 방향: 기존에 파일이 없지만 `AGENTS.md`가 canonical exemplar 경로로 지정한
  `jvm/java/java_synchronized.md`를 생성하고 상위 Java 문서에서 연결한다.

## 5. Frozen Checklist

- [x] 저장소 규칙과 기존 문서 구조를 확인했다.
- [x] 1차 규격과 Java 25 API 근거를 확보했다.
- [x] 중립적이고 자족적인 `java_synchronized.md`를 작성한다.
- [x] 소스와 bytecode 차이를 재현하는 최소 예제를 추가한다.
- [x] 정상적으로 도달 불가능한 상태를 방어해야 하는 이유와 조건을 일반화한다.
- [x] 상위 Java 문서에 상세 문서 링크를 추가한다.
- [x] Java 25 컴파일·`javap`, 링크·공백·diff 검사를 통과한다.
- [x] 최종 감사 후 study 저장소에 한정해 커밋한다.
- freeze 시각: 2026-07-20
- freeze 버전: v1

## 6. Execution Log

- 실제 조사한 것: study 규칙, jvm 자산, canonical exemplar 경로의 파일 부재, Java SE 25 규격·API.
- 실제 수정한 것: `java_synchronized.md`, 최소 바이트코드 예제, Java 상위 색인, 본 ledger.
- 버린 접근과 이유: 별도 회고 파일 추가는 같은 질문을 다루는 canonical 경로와 중복되므로 버렸다.

## 7. Verification

- 실행 / 확인할 명령: Java 25 `javac`, `javap -c -v`, Markdown 링크 대상 검사, trailing whitespace,
  `git diff --check`.
- PASS 조건: 예제가 컴파일되고 문서 설명과 bytecode가 일치하며 내부 링크와 정적 검사가 성공한다.
- 결과:
  - Java 25 `javac`: exit 0.
  - `javap -c -v`: instance·static method의 `ACC_SYNCHRONIZED`와 block의
    `monitorenter`·정상/예외 `monitorexit` 확인, exit 0.
  - 문서 내부 로컬 링크 검사: PASS.
  - trailing whitespace: 0건.
  - `git diff --check`: exit 0.

## 8. Final Audit & Closure

- intent-fit review: 프로젝트 고유 이름 없이 monitor 범위, JMM, Atomic 한계, 대안 선택과 종료 경계를 연결했다.
- expert-perspective review: Java 규격과 HotSpot 구현을 분리하고, JDK 24 이후 virtual thread pinning 변화를 반영했다.
- remaining risks: 실제 contention 비용은 workload와 JVM에 따라 달라지므로 측정 없이 성능 수치로 단정하지 않았다.
- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 커밋 해시 / 미커밋 사유: 이 ledger와 문서를 포함하는 최종 study 커밋으로 확정한다.
