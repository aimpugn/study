# WORK_20260506_INTERVIEW_CORE_GUIDE

## 0. Meta

- 작업 제목: 핵심 인터뷰 정리 문서 생성
- WORK 파일 경로: `docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `explain | audit | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청: 기존 문서는 그대로 두고, 자주 묻는 질문과 꼭 알아야 할 질문을 여러 소주제 연결형 핵심 인터뷰 정리로 새로 만들기
- 대상 경로 / 자산: `interviews/core-interview-guide.md`, `docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
- 시작 일시: `2026-05-06 22:36:49 KST`
- 종료 일시: `2026-05-06 22:48:14 KST`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 면접에서 한 질문이 여러 기술 주제로 이어지는 상황을 대비할 수 있는 핵심 인터뷰 정리 문서를 새로 만든다.
- refs: `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, `interviews/README.md`, `_question-index.md`, 10개 대주제 문서, source reservoir 일부
- scope: 기존 문서는 보존하고 새 root-level 인터뷰 허브 문서만 추가한다.
- mode: `execute`
- run_mode: `normal`
- finish: `test+commit`
- must_keep: 기존 문서 내용은 수정하지 않는다. 질문 수집형 문서의 원문 순서와 문체를 그대로 답습하지 않는다.
- extra_checks: transaction, index, event loop, Spring runtime, messaging, distributed consistency, TLS, 장애 분석처럼 꼬리 질문으로 확장되는 핵심 묶음을 포함한다.

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - 현재 문서들은 그대로 둔다.
  - 자주 묻는 질문 또는 꼭 알고 있어야 하는 질문을 특히 꼽는다.
  - 시간 촉박 여부와 무관하게 개발자가 반드시 알아야 하는 것들을 빠르게 상기할 수 있게 한다.
  - 단일 소주제가 아니라 여러 소주제가 엮여 답변되는 실전형 구조로 만든다.
  - 트랜잭션과 인덱싱처럼 꼬리 질문이 이어지는 대표 묶음을 반영한다.
- 사용자가 명시한 금지 사항:
  - 기존 문서 변경 금지.
- path / naming / format / finish 관련 요구:
  - 명시 파일명은 없으므로 기존 naming 흐름에 맞춰 `interviews/core-interview-guide.md`를 선택.
- 내가 추가한 누락 방지 항목:
  - 면접 첫 30초 답변, 꼬리 질문 지도, 실무 판단 포인트, 확인 경로를 각 핵심 묶음에 둔다.
  - repo 규칙에 따라 WORK 기록과 final review, pathspec-limited commit을 수행한다.

### 1.2 Non-Goals

- 기존 10개 대주제 문서를 deep rewrite하지 않는다.
- source reservoir를 재생성하지 않는다.
- 모든 인터뷰 질문을 전수 답변화하지 않는다. 이번 whole unit은 "핵심 복합 질문 허브 1개 생성"이다.
- push는 요청되지 않았으므로 하지 않는다.

## 2. Root-First Framing

- 근본 문제: 현재 `interviews`에는 세부 원재료와 대주제 문서가 있지만, 면접장에서 한 질문을 여러 계층으로 엮어 빠르게 답변하는 허브가 부족하다.
- 왜 중요한가: 실제 면접은 `ACID가 뭔가요`에서 끝나지 않고 MVCC, phantom read, Spring 전파, 2PC, 보상 트랜잭션으로 이어진다. 개별 문서만 읽으면 답변 연결 순서를 잃을 수 있다.
- 작업 목표: 세부 문서를 보존하면서, 빠른 회상과 꼬리 질문 대비용 핵심 정리 문서를 추가한다.
- 기대 이점: 급한 복습 때 우선순위를 잡고, 긴 꼬리 질문에서도 계층별 답변 경로를 잃지 않는다.
- 성공 정의: 새 문서가 실전 질문 묶음, 30초 직답, 꼬리 질문, 실무 판단, 확인 경로를 포함하고 기존 문서를 수정하지 않는다.
- PARTIAL 조건: 문서는 생겼지만 핵심 묶음이나 확인 경로가 빠짐.
- BLOCKED 조건: 저장소 쓰기/검증/커밋이 환경상 불가능함.

## 3. Reader & Internalization Contract

- 주 독자: 경력 백엔드/서버 개발 기술 인터뷰를 준비하는 미래의 사용자.
- 독자가 나중에 스스로 설명할 수 있어야 하는 것: 작은 기술 단위가 실제 면접 질문에서 어떻게 큰 답변으로 조립되는가.
- 사용자가 내재화해야 할 사고 패턴: 문제 고정 -> 불변식 -> 낮은 계층 메커니즘 -> 프레임워크/서비스 적용 -> 트레이드오프 -> 검증.
- 특히 막아야 하는 오해:
  - 2PC와 2PL 혼동.
  - non-blocking과 async 동일시.
  - B+Tree를 단순히 "빠른 자료구조"로만 설명.
  - `@Transactional`을 annotation만 붙이면 항상 동작한다고 설명.
  - messaging의 broker 전달 보장과 business side effect의 exactly-once를 혼동.
- 목차 필요 여부와 이유: 긴 허브 문서이므로 Markdown 목차 필요.
- 이번 문서의 기본 전개 흐름: 사용법 -> 우선순위 -> 공통 답변 프레임 -> 핵심 복합 질문 묶음 -> 복합 시나리오 연습 -> 마지막 점검 질문.
- 품질 기준 exemplar:
  - primary exemplar: `computer_architecture/threads/threads.md`
  - reference principle: 여러 계층의 기술 단위를 하나의 큰 계보로 연결.
  - trait to avoid: taxonomy breadth만 늘리고 면접 답변 순서를 잃는 구조.
  - target quality lift: 각 묶음마다 30초 답변과 꼬리 질문 지도를 붙여 실전 회상성을 높인다.
  - secondary exemplar: `jvm/java/java_synchronized.md`
  - secondary principle: 작은 실패 상황에서 시작해 런타임/하드웨어 이유까지 연결.
  - trait to avoid: 개별 API 설명으로 축소되는 것.

## 4. Depth Decision

- 선택한 깊이: `full`
- 이유: 인터뷰 하위 프로젝트의 핵심 허브 문서 생성이며, 여러 대주제와 source reservoir를 연결해야 한다.
- 전체 루프 트리거: 새 정식 문서 생성, 설명 품질 핵심, 다중 주제 연결, repo 변경 및 commit 필요.
- 축약 가능한 섹션과 근거: 외부 공식 문서 전수 인용은 이번 범위가 "기존 repo 자산 기반 핵심 허브"이므로 local evidence 중심으로 축약한다. 엔진별 최신 세부는 각 세부 문서 보강 시 별도 확인한다.

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`
  - `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/USECASE.md`
  - `/Users/rody/VscodeProjects/study/interviews/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/interviews/USECASE.md`
  - `/Users/rody/VscodeProjects/study/interviews/README.md`
- 활성화한 프로젝트 규칙:
  - 기존 평균이 아니라 exemplar 수준을 기준으로 한다.
  - 질문형/interview 문서는 `질문 -> 짧은 직답 -> 계층별 설명 -> 구체 예시 -> 꼬리 질문 -> 검증/근거` 구조를 우선한다.
  - source reservoir는 원재료이지 스타일 기준이 아니다.
  - 파일 변경 작업은 final review, verification, commit으로 닫는다.
- 전역 규칙과의 충돌 여부 / 해소: 충돌 없음. 런타임 delegation은 user가 명시하지 않았으므로 developer rule에 따라 `single-agent + separated audit lanes`로 보완한다.

## 6. Topic Analysis

- 현재 이해한 사용자 의도: 기존 세부 문서들은 보존하고, 실제 면접에서 꼬리 질문에 대비할 수 있는 핵심 복합 질문 정리를 원한다.
- 현재 보이는 문제 구조: 10개 대주제 문서는 원문 배치본이고, 질문 index는 raw source 위치 추적에 강하지만 실전 우선순위와 답변 조립 경로는 별도 허브가 필요하다.
- 핵심 경계: 모든 질문 전수 답변이 아니라 핵심 묶음 중심의 빠른 회상 문서.
- 숨은 가정 / 불확실성: "자주 묻는"의 빈도는 외부 면접 데이터가 아니라 기존 source reservoir와 사용자 예시, 백엔드 인터뷰 일반 중요도에 기반한 강한 추론이다.
- 성공을 오판하기 쉬운 지점: 너무 짧은 치트시트로 만들어 메커니즘이 사라지거나, 반대로 세부 문서를 다시 쓰듯 길어져 빠른 회상성이 사라지는 것.

## 7. Analysis Critique + Repair

- Challenge 1: 핵심 질문 선별이 주관적일 수 있다.
  - 보강안: 사용자 예시와 local project intent에 명시된 기술 단위, `_question-index.md`의 반복 출현 항목을 함께 근거로 삼는다.
- Challenge 2: 기존 문서를 그대로 둔다는 요구를 README 업데이트로 어길 수 있다.
  - 보강안: 기존 문서는 수정하지 않고 새 문서와 WORK만 추가한다.
- Challenge 3: 각 항목이 얕은 문답 카드처럼 끝날 수 있다.
  - 보강안: 모든 핵심 묶음에 30초 답변, 이어 말할 순서, 꼬리 질문, 실무 포인트, 확인 경로를 둔다.

## 8. Scope Expansion & Impact Sync

- 시작 키워드: 트랜잭션, 인덱스, B+Tree, MVCC, 2PC, 보상 트랜잭션, epoll, Spring, Kafka, RabbitMQ, TLS
- 확장 키워드: isolation, phantom, repeatable read, HikariCP, outbox, idempotency, event loop, non-blocking, keep-alive, GC, p99, slow query, replication, eventual consistency
- 조사한 경로:
  - `interviews/*.md` headings
  - `_question-index.md`
  - `source/_source-context-and-question-bank.md`
  - 관련 대주제 문서 snippets
- 함께 점검한 자산:
  - `database-storage-search-nosql.md`
  - `distributed-systems-architecture.md`
  - `messaging-event-driven.md`
  - `concurrency-async-io.md`
  - `network-web-protocols.md`
  - `security-cryptography.md`
  - `spring-backend-frameworks.md`
  - `language-runtime.md`
  - `os-kernel-computer-architecture.md`
  - `problem-solving-code-quality.md`
- 함께 움직여야 하는 표면: 새 문서가 기존 원재료를 대체하지 않음을 opening에 명시.
- 제외 표면과 근거: README 링크 추가는 편의상 유용하지만 "현재 문서들은 그대로" 요구를 약화할 수 있어 제외.

## 9. Evidence Gathering

- E-01
  - 주장: `interviews`의 목적은 경력 기술 인터뷰 준비이며, 짧은 직답과 깊은 메커니즘을 함께 요구한다.
  - 근거 유형: repo evidence
  - 자료: `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`
  - 상태: APPLY
- E-02
  - 주장: 기존 대주제 문서는 현재 원문 배치본이고, source reservoir를 style exemplar로 쓰면 안 된다.
  - 근거 유형: repo evidence
  - 자료: `interviews/README.md`, root `AGENTS.md`
  - 상태: APPLY
- E-03
  - 주장: transaction, index, async/event loop, Spring runtime, messaging, TLS, 장애 분석은 existing source에서 반복되며 핵심 묶음 후보로 타당하다.
  - 근거 유형: repo evidence + inference
  - 자료: `_question-index.md`, `rg` expanded search, 관련 대주제 snippets
  - 상태: APPLY
- E-04
  - 주장: 기존 문서 보존 요구를 지키려면 README나 10개 대주제 문서를 수정하지 않는 편이 안전하다.
  - 근거 유형: explicit user requirement + reasoning
  - 자료: 사용자 요청 원문
  - 상태: APPLY

## 10. Evidence Critique + Repair

- 소스 품질 리스크: source reservoir에는 rough note와 일부 부정확하거나 불완전한 문장이 섞여 있다.
- 보강: raw prose를 그대로 옮기지 않고, 질문 cluster와 local project intent를 기준으로 재구성한다.
- 남는 한계: 각 DB/브로커/JDK 버전별 최신 세부는 이 허브에서 전수 확정하지 않는다. 세부 문서 보강 시 official docs로 닫아야 한다.

## 11. Design

- 선택한 접근: `interviews/core-interview-guide.md` 새 파일 추가.
- 고려한 대안:
  - 기존 README에 핵심 링크를 추가: discoverability는 좋아지지만 기존 문서 보존 요구를 건드린다.
  - 각 대주제 문서에 핵심 섹션 추가: 세부 문서가 더 좋아지지만 "빠른 통합 허브" 목적에는 분산된다.
  - 새 파일 하나: 기존 문서 보존과 실전 통합 허브 목적을 동시에 만족한다.
- 최종 채택: 새 파일 하나.
- 문서 뼈대: 사용법, 우선순위, 공통 프레임, 14개 핵심 질문 묶음, 복합 질문 연습, 마지막 점검 질문.
- 검증 경로: pathspec diff, 기존 파일 미수정 확인, Markdown heading/anchor sanity, 핵심 키워드 coverage `rg`, final review, commit.

## 12. Design Critique + Repair

- Architect view 반론: 14개 묶음은 넓어서 얕아질 수 있다.
  - 결정: 각 묶음을 세부 완결 문서가 아니라 "답변 경로 지도"로 제한하고, 확인 경로를 세부 문서로 연결한다.
- Learner view 반론: 빠른 회상 문서가 너무 압축되면 꼬리 질문에서 무너질 수 있다.
  - 결정: 각 묶음에 꼬리 질문 지도와 실무 답변 포인트를 넣어 압축과 연결성을 함께 유지한다.
- Protocol view 반론: 기존 문서를 건드리지 말라는 요구와 WORK 추가가 충돌할 수 있다.
  - 결정: WORK는 repo 규칙상 실행 기록이고, 사용자 학습 문서 본문은 건드리지 않는다. final에서 새로 추가한 파일을 명확히 밝힌다.

## 13. Overall Plan

1. Instruction stack과 project fact docs를 확인한다.
2. Existing interviews structure, source index, 핵심 snippets를 조사한다.
3. 새 핵심 가이드와 WORK를 추가한다.
4. pathspec verification과 final audit를 수행한다.
5. 새 파일만 stage/commit한다.

## 14. Plan Critique + Repair

- 실패 지점: broad `git diff --check`가 unrelated dirty files 때문에 task closure를 흐릴 수 있다.
- 보강안: 새 파일 pathspec으로만 diff/check/stage/commit한다.
- 실패 지점: 너무 많은 source file을 직접 수정할 위험.
- 보강안: `apply_patch` Add File만 사용하고 existing file update hunk를 쓰지 않는다.

## 15. Detailed Task Plan

- 생성 파일:
  - `interviews/core-interview-guide.md`
  - `docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
- 변경하지 않을 파일:
  - existing `interviews/*.md`
  - `interviews/source/*`
  - `interviews/README.md`
- 성공 케이스:
  - S1: 새 문서가 빠른 회상과 실전 꼬리 질문 대비를 지원한다.
  - S2: transaction/index 예시가 직접 반영된다.
  - S3: 기존 문서가 수정되지 않는다.
- 실패 케이스:
  - F1: 기존 문서를 수정함.
  - F2: 단일 소주제 요약만 있고 cross-topic 연결이 없음.
  - F3: 확인 경로 없이 암기 문장만 나열함.

## 16. Detailed Plan Critique + Repair

- 누락된 케이스: 장애 분석, TLS, GC, 알고리즘/코드 품질도 개발자 필수 질문이다.
- 보강안: 14개 핵심 묶음에 포함.
- 최종 상세 계획: 새 파일에 핵심 묶음을 모두 넣고, local doc refs와 점검 질문을 둔다.

## 17. Frozen Checklist

- C-01
  - 출처: 사용자
  - 내용: 기존 문서들은 그대로 둔다.
  - PASS 기준: existing tracked/interviews docs에 update diff가 없다.
- C-02
  - 출처: 사용자
  - 내용: 자주 묻거나 반드시 알아야 할 핵심 인터뷰 질문을 꼽는다.
  - PASS 기준: 시간 없을 때 우선순위와 핵심 묶음 섹션이 있다.
- C-03
  - 출처: 사용자
  - 내용: 여러 소주제가 엮이는 실전형 답변 구조를 제공한다.
  - PASS 기준: 각 묶음이 30초 답변, 이어 말할 순서, 꼬리 질문, 실무 포인트, 확인 경로를 포함한다.
- C-04
  - 출처: 사용자
  - 내용: transaction과 index 예시를 직접 반영한다.
  - PASS 기준: transaction/MVCC/2PC/보상 트랜잭션, B+Tree/covering/row fetch/depth가 별도 핵심 섹션으로 있다.
- C-05
  - 출처: AI-추가
  - 내용: repo closure 규칙을 지킨다.
  - PASS 기준: pathspec verification, final audit, 새 파일만 commit.

### 17.2 Checklist Quality Review

- 각 항목이 목표, 이점, 불변식에 매핑된다: Y
- PASS/FAIL이 관측 가능하다: Y
- 필요한 근거 또는 검증 경로가 있다: Y
- 사용자 요구가 조용히 약화되지 않았다: Y
- 한 항목 실패 시 task가 reopened 되는 구조다: Y
- 판정: PASS

### 17.3 Freeze

- freeze 시각: `2026-05-06 22:36:49 KST`
- freeze 버전: v1

## 18. Execution Log

- 실제 조사한 것:
  - instruction stack, root/local project fact docs, study-explanation/rigorous-task skill references
  - interviews file inventory, headings, `_question-index.md`, source context, selected snippets
- 실제 수정한 것:
  - 새 `interviews/core-interview-guide.md` 추가
  - 새 WORK file 추가
- 실행 중 바뀐 가정:
  - README discoverability update는 제외했다. 기존 문서 보존 요구가 더 강하다.
- 버린 접근과 이유:
  - 기존 대주제 문서별 핵심 섹션 추가: 사용자 보존 요구와 충돌 가능성.

## 19. Verification

### 19.1 Verification Plan

- 실행 / 확인할 명령:
  - `git diff -- --stat interviews/core-interview-guide.md docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
  - `git diff --check -- interviews/core-interview-guide.md docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
  - `rg -n "트랜잭션|MVCC|2PC|보상|B\\+Tree|커버링|epoll|@Transactional|Kafka|RabbitMQ|TLS|GC|p99" interviews/core-interview-guide.md`
  - `git status --short -- interviews/core-interview-guide.md docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
  - `git status --short -- interviews/*.md ':!interviews/core-interview-guide.md'`
- PASS 조건:
  - 새 파일 diff만 존재.
  - whitespace check PASS.
  - 핵심 키워드 coverage PASS.
  - 기존 인터뷰 문서 수정 없음.
- FAIL 조건:
  - existing docs diff 발생, whitespace error, 핵심 사용자 예시 누락.

### 19.2 Verification Result

- whitespace check: PASS
  - `git diff --no-index --check /dev/null interviews/core-interview-guide.md`
  - `git diff --no-index --check /dev/null docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md`
- keyword coverage: PASS
  - `트랜잭션`, `MVCC`, `2PC`, `보상`, `B+Tree`, `커버링`, `epoll`, `@Transactional`, `Kafka`, `RabbitMQ`, `TLS`, `GC`, `p99`, `outbox`, `idempotent`, `phantom`, `REPEATABLE READ` 확인.
- local link existence: PASS
  - guide 안의 상대 Markdown 링크가 현재 저장소 파일로 resolve됨.
- heading / repeated section structure: PASS
  - 14개 핵심 묶음이 `첫 30초 답변`, `이어 말할 순서`, `꼬리 질문 지도`, `실무 답변 포인트`, `확인 경로` 구조를 가진다.
- existing-doc preservation: PASS
  - 새로 추가한 두 파일 외 기존 interview 문서는 이번 작업에서 수정하지 않았다.
- pathspec-limited commit readiness: PASS
  - stage 대상은 `interviews/core-interview-guide.md`, `docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE.md` 두 파일로 제한한다.

## 20. Explanation Quality Review

- 결론이 초반에 드러나는가: PASS
- 질문에 대한 직접 답이 초반에 드러나는가: PASS
- 왜 중요한지가 닫히는가: PASS
- 근거와 제약이 설명에 연결되는가: PASS
- 검증 경로가 보이는가: PASS
- 의미 있는 대안과 트레이드오프가 남는가: PASS
- 구체적인 anchor가 있는가: PASS
- 불확실성이 정직하게 표시되는가: PASS
- 전이 가능한 원리가 남는가: PASS
- 필요한 추상화 계층이 연결되는가: PASS
- 핵심 대조쌍이 대칭적으로 설명되는가: PASS
- 현재 저장소의 낮은 품질 문서를 답습하지 않았는가: PASS
- 선택한 exemplar 수준까지 충분히 끌어올렸는가: PASS for current guide role, not promoted as exemplar.

## 21. Final Audit & Closure

- intent-fit review: PASS. 새 guide가 기존 세부 문서를 대체하지 않고, 실전 꼬리 질문에 필요한 답변 경로를 한 곳에 모은다.
- expert-perspective review: PASS. 핵심 묶음이 DB, 런타임, 네트워크, 분산 시스템, 운영 사고까지 이어져 경력 백엔드 면접의 복합 질문 대비 목적에 맞는다.
- remaining risks: official docs cross-check는 각 세부 문서 deep rewrite 단계의 후속 후보이며, 이번 guide의 closure blocker는 아님.
- 문서 / 예제 / 관련 자산 동기화 상태: existing docs intentionally unchanged.

### 21.1 Checklist Re-Judgement

- C-01: `PASS`
- C-02: `PASS`
- C-03: `PASS`
- C-04: `PASS`
- C-05: `PASS`

### 21.2 Final State

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 커밋 해시 / 미커밋 사유: 파일 내용은 commit 직전 상태로 닫힘. 실제 커밋 해시는 git 기록과 최종 응답에서 확인한다.
