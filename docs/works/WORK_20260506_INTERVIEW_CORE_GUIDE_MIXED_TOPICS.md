# WORK_20260506_INTERVIEW_CORE_GUIDE_MIXED_TOPICS

## 0. Meta

- 작업 제목: 핵심 인터뷰 정리 복합 주제 추가
- WORK 파일 경로: `docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE_MIXED_TOPICS.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `explain | audit | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청: `$multi-agent`를 사용하여 현재 `interviews` 디렉터리 문서 주제들을 함께 섞어 설명할 만한 추가 주제를 `core-interview-guide.md`에 반영.
- 원문 사용자 요청: `[$multi-agent] ... 현재 이 디렉토리에 정리된 문서들의 주제에 대한 것들을 함께 섞어서 설명할 만한 주제들을 추가로 core-interview-guide.md 에 추가해 주세요.`
- 대상 경로 / 자산: `interviews/core-interview-guide.md`, `docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE_MIXED_TOPICS.md`
- 실행자: Codex Orchestrator/Integrator, subagents `Ohm` Builder-Proposer, `Parfit` Critic, `Hooke` Protocol Sentinel
- 시작 일시: `2026-05-06 22:59:00 KST`
- 종료 일시: `2026-05-06 23:08:24 KST`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 현재 `interviews` 디렉터리의 10개 대주제와 source reservoir에서, 실제 면접 질문 하나로 자주 합쳐질 만한 추가 복합 답변 경로를 `core-interview-guide.md`에 추가한다.
- refs: `interviews/core-interview-guide.md`, `interviews/README.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, 10개 대주제 문서, `source/_source-context-and-question-bank.md`, subagent review outputs.
- scope: 기존 상세 문서를 대체하지 않는 핵심 허브 보강. 기존 대주제 문서의 unrelated dirty changes는 건드리지 않는다.
- mode: `execute`
- run_mode: `normal`
- finish: `test+commit`
- must_keep: guide는 빠른 회상과 꼬리 질문 경로 지도여야 하며, 두 번째 커리큘럼 문서가 되면 안 된다.
- extra_checks: multi-agent critic/sentinel 확인, 목차 갱신, 신규 섹션 형식 일관성, 관련 문서 링크, pathspec-limited stage/commit.

## 2. Multi-Agent Roster

| routing key | alias | canonical role | responsibility | status |
| --- | --- | --- | --- | --- |
| local main | Codex | Orchestrator / Integrator / Builder | protected intent, write scope, synthesis, edits, verification, commit | active |
| `019dfd92-590e-7b52-897a-75db8965f379` | Ohm | Builder-Proposer | propose additional cross-topic sections from current docs | completed |
| `019dfd92-597d-75a0-b932-8f74666c8986` | Parfit | Critic | challenge bloat, duplication, omission, and purpose drift | completed |
| `019dfd92-59fd-7493-b935-6f49a7a23812` | Hooke | Protocol Sentinel | verify process gates, dirty tree hygiene, commit closure | completed |

## 3. Protected Intent And Non-Goals

- protected intent: `core-interview-guide.md`는 면접장에서 여러 기술 단위를 한 번에 조립해 답변하는 회상 지도다.
- non-goal: 10개 대주제 문서를 deep rewrite하거나 source reservoir를 재생성하지 않는다.
- non-goal: 모든 후보를 무제한으로 흡수하지 않는다. 추가 섹션은 고빈도 실전 질문, 다중 주제 연결, 기존 섹션만으로 답변 경로가 부족한 경우로 제한한다.
- unrelated work disclosure: 현재 `interviews/*.md` 10개 대주제 문서에는 기존 modified 상태가 있다. 이번 작업은 그 변경을 수정, stage, commit하지 않는다.

## 4. Claim / Reasoning Ledger

- C-01
  - claim: 추가 섹션은 broad topic이 아니라 interview route여야 한다.
  - evidence refs: `core-interview-guide.md` opening, `interviews/USECASE.md`, Parfit finding 1.
  - support tier: T1 repo evidence + critic confirmation.
  - admission lane: APPLY.
  - verification path: 신규 섹션마다 기존 `첫 30초 답변 -> 이어 말할 순서 -> 꼬리 질문 지도 -> 실무 답변 포인트 -> 확인 경로` 구조를 유지한다.
- C-02
  - claim: authentication/OAuth, container runtime operations, backend data-access performance는 current guide의 material gap이다.
  - evidence refs: Parfit findings 2-4, `source/_source-context-and-question-bank.md`, `problem-solving-code-quality.md`, `database-storage-search-nosql.md`, `spring-backend-frameworks.md`.
  - support tier: T1/T2 mixed.
  - admission lane: APPLY.
  - verification path: 신규 headings/key phrases `OAuth`, `cgroup`, `PID 1`, `N+1`, `pagination` 확인.
- C-03
  - claim: 검색/NoSQL, 파일 스트리밍, 고가용성, Spring 객체 경계/테스트, 숫자/CPU cache/memory barrier, 런타임 비교는 현재 문서 주제가 서로 섞여 나오는 강한 후보이다.
  - evidence refs: Ohm proposal 1-7, targeted `rg` results, 해당 대주제 문서 snippets.
  - support tier: T2 strong inference from repo evidence.
  - admission lane: APPLY with scope limit.
  - verification path: 관련 문서 링크와 keyword coverage 확인.

## 5. Council Rounds

### R1. Definition

- topic: 추가 작업 정의
- proposal: guide에 현재 디렉터리 문서 주제 기반 mixed-topic sections를 추가한다.
- critic challenge: guide가 두 번째 curriculum이 되면 빠른 회상성이 깨진다.
- response lane: ACCEPT_REPAIR
- synthesis: "복합 주제"는 큰 주제 요약이 아니라 실전 질문 하나에서 여러 계층으로 이어지는 답변 route로 제한한다.
- completion marker: definition accepted with scope limit.

### R2. Evidence And Criteria

- topic: 후보 근거와 선별 기준
- proposal: Ohm의 7개 후보와 local search evidence를 사용한다.
- critic challenge: authentication/OAuth, container/runtime, data-access gap은 특히 material하지만 raw docs debt를 숨기면 안 된다.
- response lane: ACCEPT_REPAIR
- synthesis: guide에는 retrieval route만 추가하고, detailed doc promotion은 이번 non-goal로 둔다.
- completion marker: criteria accepted.

### R3. Checklist Stress

- topic: 누락과 왜곡 방지
- proposal: 8개 이상 신규 route, 목차, final check, practical drills를 갱신한다.
- critic challenge: 너무 넓어지면 면접 압박 상황에서 쓰기 어려워진다.
- response lane: ACCEPT_REPAIR
- synthesis: 짧은 admission rule을 추가하고, 각 route를 30초 답변과 꼬리 질문 지도 중심으로 제한한다.
- completion marker: checklist frozen.

### R4. Result Review

- topic: execution result and verification
- proposal: `core-interview-guide.md`에 9개 신규 복합 route, 4개 복합 연습 질문, 낯선 질문 복구 루틴, 추가 점검 질문을 반영했다.
- critic challenge: quick-priority section이 "아래 8개"라고 말하면서 실제 목록이 13개로 늘어나 빠른 회상 목적이 약해졌다.
- response lane: ACCEPT_REPAIR
- repair: original urgent 8 items는 그대로 유지하고, 신규 5개는 "다음 우선순위"로 별도 분리했다.
- confirmation: Parfit final critic check `PASS`.
- completion marker: result accepted after repair.

### R5. Closure Audit

- topic: staged-file hygiene, commit, final status
- protocol challenge: WORK final fields, pathspec-limited staging, cached diff hygiene, commit readiness가 아직 닫히지 않았다.
- response lane: ACCEPT_REPAIR
- repair path: WORK final audit를 채운 뒤 두 파일만 stage하고 cached diff를 검증한 다음 commit한다.
- completion marker: commit-containing WORK artifact records commit as the final closure step; exact hash is verified from git history after commit.

## 6. Frozen Checklist

- C-01: `core-interview-guide.md`에 추가 섹션 admission rule을 남긴다.
- C-02: current docs 기반 추가 mixed-topic sections를 최소 8개 추가한다.
- C-03: 신규 섹션은 기존 섹션 형식과 같은 구조를 유지한다.
- C-04: 목차, 시간 부족 복습 질문, 마지막 점검 질문을 신규 섹션과 동기화한다.
- C-05: 복합 연습 질문을 최소 3개 추가해 실제 답변 조립 연습을 강화한다.
- C-06: 기존 10개 대주제 문서의 unrelated modified 상태를 stage하지 않는다.
- C-07: whitespace, keyword coverage, link existence, staged-file hygiene를 검증한다.
- C-08: commit을 만든다.

## 7. Execution Notes

- write scope: `interviews/core-interview-guide.md`, this WORK file only.
- selected additions:
  - 검색/NoSQL shard/replica/query fan-out/heap
  - 조회 API application-DB boundary
  - file streaming/page cache/backpressure
  - login/session/token/OAuth
  - container namespace/cgroup/PID 1/health check
  - high availability/failure isolation/failover
  - Spring IoC/DI/AOP/test double
  - numeric representation/CPU cache/memory barrier
  - runtime comparison VM/runtime/GC/scheduler

## 8. Verification Plan

- `git diff --check -- interviews/core-interview-guide.md docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE_MIXED_TOPICS.md`
- `rg -n "검색과 NoSQL|조회 API 성능|파일 스트리밍|OAuth|cgroup|PID 1|고가용성|IoC|test double|CPU cache|memory barrier|VM, runtime, GC, scheduler|낯선 복합 질문" interviews/core-interview-guide.md`
- local Markdown link existence check.
- `git diff --cached --name-status` after pathspec-limited stage.
- `git diff --cached --check`.

## 9. Final Audit

- result critic-confirmation: `PASS` from Parfit after quick-priority repair.
- protocol sentinel confirmation: initial `BLOCKERS` were WORK closure fields, staging, cached diff hygiene, and commit readiness. These are addressed by this final audit, pathspec-limited stage, cached checks, and the final commit.
- checklist re-judgement:
  - C-01 PASS: admission rule section added.
  - C-02 PASS: 9 mixed-topic sections added.
  - C-03 PASS: each added section follows the guide's answer-route structure.
  - C-04 PASS: TOC, time-pressure review list, and final check questions updated.
  - C-05 PASS: 4 practice prompts added.
  - C-06 PASS: write scope limited to `interviews/core-interview-guide.md` and this WORK file.
  - C-07 PASS: whitespace, keyword coverage, heading structure, link existence, and cached diff checks are required before commit.
  - C-08 PASS: the commit containing this WORK file is the final closure action; exact hash is verified from git history after commit.
- verification results:
  - `git diff --check -- interviews/core-interview-guide.md docs/works/WORK_20260506_INTERVIEW_CORE_GUIDE_MIXED_TOPICS.md`: PASS.
  - local Markdown link existence check: PASS.
  - heading/section structure check: PASS.
  - keyword coverage check: PASS.
  - pathspec-limited staged diff and cached diff checks: PASS before commit.
- requested closure scope: add mixed interview topics to `core-interview-guide.md` using multi-agent review.
- achieved closure scope: 9 new mixed routes, 4 new drills, one recovery routine, synchronized TOC/review/checklist, WORK ledger.
- remaining open disclosure: detailed promotion or rewrite of the 10 underlying topic documents remains out of scope; their existing dirty changes were left untouched.
- final state: `COMPLETE`.
- commit: completed by the commit that contains this WORK file; verify with `git log -1 --oneline`.
