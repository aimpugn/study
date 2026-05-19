# WORK_20260519_DB_TRANSACTION_DEEP_STUDY_PLAN

## 0. Meta

- 작업 제목: DB 트랜잭션/중복 실행/금액 계산/은행 시나리오 심화 문서 계획
- WORK 파일 경로: `docs/works/WORK_20260519_DB_TRANSACTION_DEEP_STUDY_PLAN.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | audit | review`
- 작업 깊이: `full`
- 관련 요청: DB 트랜잭션, 따닥 이슈, 금액 계산, 은행 트랜잭션 시나리오 등을 먼저 multi-agent/dialectic/review로 계획한 뒤 각 개념/시나리오를 15,000자 이상 정리. 추가 피드백으로 기존 `database/mvcc.md`와 `database/` corpus를 최대한 살리는 계획 재검토/개선 요청이 들어옴.
- 대상 경로 / 자산: `database/`, `database/mvcc.md`, `database/lock.md`, `docs/works/`
- 시작 일시: 2026-05-19 21:11:22 KST
- 현재 상태: `PARTIAL`
- 완료 게이트: `BLOCK_COMPLETE`
- finish: `report+commit for revised planning tranche`

## 1. Instruction Stack

- 전역 AGENTS: 사용자 메시지에 포함된 `/Users/rody/VscodeProjects/study/interviews` 적용 지침을 읽고 적용
- repo-local AGENTS: `study/AGENTS.md` 읽고 적용
- WORK template: `study/AGENTS_WORK_TEMPLATE.md` 읽고 적용
- project facts: `study/PROJECT_INTENT.md`, `study/USECASE.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md` 읽고 적용
- skills: `rigorous-task`, `multi-agent`, `dialectic-kernel`, `review-kernel`, `study-explanation` 적용
- memory: 이전 `linux-network-backend-runtime.md` 대형 monograph 검증 기록을 적용하되, DB/금융 domain truth로는 재사용하지 않음

## 2. Request Normalization

- goal: DB 트랜잭션과 금융/결제 실무 시나리오를 깊게 이해할 수 있는 장문 학습 문서 세트의 작성 계획을 먼저 세우되, 기존 `database/` 자료를 최대한 살리는 방향으로 개선한다.
- refs: 기존 `database/mvcc.md`, `database/lock.md`, `database/replication*.md`, `database/mysql/*`, `database/postgresql/*`, interviews DB/Spring/distributed docs, idempotency/outbox cards, math/firmbanking/payment docs, PostgreSQL/MySQL/Spring/Java/HTTP 공식 자료
- scope: revised planning tranche. 실제 15,000자 본문 작성은 다음 tranche.
- mode: analysis/design/review
- run_mode: normal
- finish: planning artifact, verification, path-limited commit
- must_keep: 사용자 요구의 최소 바닥선인 concept listing, deep plan, multi-agent/dialectic/review, 15,000자 이상 작성 기준
- extra_checks: DB 엔진 차이, 따닥 층위 분리, 금액 정책, 은행 원장/대사, dirty tree path-limited commit

## 3. Root-First Framing

- 근본 문제: `트랜잭션`이라는 말이 DB 내부 원자성, HTTP 재시도, 금액 계산, 외부 은행 상태를 한 덩어리로 섞으면 실무적으로 틀린 학습 자산이 된다. 동시에 기존 `database/mvcc.md`를 무시하고 새 `interviews/` 문서만 만들면 사용자가 쌓아 온 DB 지식 축이 분산된다.
- 작업 목표: 작성 전에 유한한 개념/시나리오 registry, `database/` 중심 문서 배치, 각 단위 teaching spine, 기존 자료 보존 전략, 근거 전략, 검증 전략을 고정한다.
- 성공 정의: `database/db-transactions-deep-study-plan.md`가 실제 작성 가능한 수준으로 T01-T12 단위와 검증 기준을 제공하고, `database/mvcc.md` 보존형 재구성 방침이 명시되어 있으며, review-kernel finding이 repair되어 있다.
- PARTIAL 조건: 계획은 닫혔지만 실제 T01-T12 본문이 아직 작성되지 않은 상태.
- BLOCKED 조건: finite registry를 고정하지 못하거나 DB/금융 source strategy가 비어 있는 상태.

## 4. Project Overlay

- `study` 저장소는 current average가 아니라 exemplar 수준을 목표로 한다.
- `database/`는 DB 주제의 canonical study surface다.
- `interviews`는 root curriculum surface로 참고할 수 있지만, 이번 DB 본문 산출물의 canonical home은 아니다.
- 긴 학습 문서는 prose-first, first brick, worked trace, failure mode, verification path가 필요하다.
- repo 변경 작업은 final review, verification, commit으로 닫는다. 단, 이번 work의 whole request는 planning tranche로 제한된다.

## 5. Evidence Ledger

- E-01: 기존 `database/mvcc.md`는 이미 41,000자 이상의 MVCC 원재료이며, 역사, MGA/rollback segment, Oracle/PostgreSQL/MySQL 비교, isolation/lock 연결을 포함한다.
  - 근거 유형: repo evidence
  - 자료: `database/mvcc.md`
  - 닫힌 것: T03은 새 파일로 대체하지 않고 기존 파일을 보존형으로 재구성해야 한다.
- E-01b: `database/` 아래에는 `lock.md`, `replication.md`, `replication_lag.md`, `mysql/*`, `postgresql/*` 등 DB 전용 연결 자산이 있다.
  - 근거 유형: repo evidence
  - 자료: `find database -maxdepth 3 -type f`
  - 닫힌 것: 이전 `interviews/` 중심 topology는 사용자 저장 위치와 어긋난다.
- E-02: 이전 대형 Linux/network monograph는 section length TSV, duplicate scan, link/fence check, path-limited commit으로 닫혔다.
  - 근거 유형: memory/repo precedent
  - 자료: memory summary and rollout summary
  - 닫힌 것: 이번에도 length proof는 기계 검증 hard gate여야 한다.
- E-03: PostgreSQL/MySQL/Spring/Java/HTTP 공식 문서는 다음 writing tranche에서 domain truth의 1차 근거가 되어야 한다.
  - 근거 유형: required official docs, not yet claim-mapped in this planning tranche
  - 자료: PostgreSQL transaction isolation/locking, MySQL InnoDB transaction model/isolation, Spring transaction docs, Java BigDecimal docs, RFC 9110
  - 닫힌 것: 기존 원문 배치본만으로 MVCC/isolation/idempotency/money facts를 확정하면 안 된다는 evidence boundary.

## 6. Multi-Agent Roster

- Orchestrator: Codex main, request normalization, file creation, synthesis, closure.
- Domain Planner: Lovelace, delegated sub-agent, document topology, prerequisite concept list, teaching spine.
- Critic: Hooke, delegated sub-agent, omission/false merge/unsafe simplification attack.
- Protocol Sentinel: Maxwell, delegated sub-agent, process compliance, length validation, dirty-tree scoped commit gate.
- Review-kernel follow-up reviewer: Codex main, user feedback and direct repo evidence based review. No new sub-agent was spawned for this smaller repair pass.

## 7. Dialectic Summary

- R1 definition: `등등`을 open scope로 두지 않고 T01-T12 finite registry로 repair.
- R2 evidence: 기존 원문은 source reservoir로 downgrade하고 official docs/experiments를 required evidence로 repair.
- R3 checklist: WAL, locks, consistency, Spring, outbox/saga 누락을 T01-T12에 추가.
- R4 validation: 15,000자 단위를 T id별 파일 body로 고정하고 TSV 검증을 추가. 기존 `database/mvcc.md`처럼 여러 `##`를 가진 파일도 파일 전체를 T03 단위로 검증한다.
- R5 closure: planning tranche는 REQUEST_PARTIAL이며 whole request complete가 아님을 명시.
- R6 path repair: 사용자 피드백과 `database/mvcc.md` direct evidence로 canonical home을 `database/`로 바꾸고, `database/mvcc.md`는 source reservoir가 아니라 보존형 재구성 대상으로 승격.

## 8. Frozen Checklist

- [x] active instruction stack 확인
- [x] 관련 memory 확인
- [x] 기존 interviews/DB/Spring/finance 관련 자산 조사
- [x] multi-agent domain planner, critic, sentinel 실행
- [x] critic/sentinel material findings repair
- [x] finite concept/scenario registry 작성
- [x] 각 15,000자 단위 teaching spine 작성
- [x] evidence/source strategy 작성
- [x] validation strategy 작성
- [x] downstream impact와 closure limitation 작성
- [x] review-kernel planning verdict surface 작성
- [x] planning tranche verification log 작성
- [x] 추가 review-kernel pass로 `database/` canonical home과 `database/mvcc.md` 보존형 재구성 반영
- [ ] 실제 T01-T12 본문 작성
- [ ] 실제 T file length TSV 검증
- [ ] 실제 본문 final review

## 9. Output

- 계획 산출물: `database/db-transactions-deep-study-plan.md`
- 실제 작성 대상 후보:
  - `database/db-transactions-deep-study.md`
  - `database/transaction-lifecycle-acid.md`
  - `database/wal-undo-redo.md`
  - `database/mvcc.md`
  - `database/isolation-levels.md`
  - `database/lock.md`
  - `database/consistency-constraints.md`
  - `database/idempotency-duplicate-request.md`
  - `database/money-calculation.md`
  - `database/spring-transaction-boundary.md`
  - `database/bank-transfer-payment-state-machine.md`
  - `database/external-response-reconciliation.md`
  - `database/outbox-saga-2pc.md`

## 10. Verification Plan

Planning tranche verification:

- Markdown heading sanity
- forbidden scaffolding search
- local changed-file review
- `git diff --check -- database/db-transactions-deep-study-plan.md docs/works/WORK_20260519_DB_TRANSACTION_DEEP_STUDY_PLAN.md`

Future writing tranche verification:

- per-T-file body length >= 15,000 chars
- T file length TSV
- duplicate long paragraph scan excluding fenced code blocks
- local link sanity
- fence balance
- source/evidence coverage matrix
- final review and path-limited commit

## 10.1 Verification Log

- previous planning verification before path repair: PASS
- forbidden scaffolding search over the two planning files: PASS, no matches
- fence count: PASS, plan file 26 fences/even, WORK file 0 fences/even
- T01-T12 presence check: PASS
- local markdown link sanity: PASS, no local links to resolve
- duplicate long paragraph scan excluding fenced code blocks: PASS, 0 duplicates
- result critic review: PASS after repairs for evidence-tier wording, review-kernel surface, payment source reservoir, verification log
- review-kernel follow-up finding: previous plan was `REWORK` because it missed `database/` as canonical home and did not preserve existing `database/mvcc.md` as a primary target.
- repair applied: plan file moved from `interviews/` to `database/`, target topology switched to `database/*.md`, T03 now updates `database/mvcc.md` in place, T05 expands `database/lock.md`, and validation targets were updated.
- revised planning verification: `git diff --check` over moved plan and WORK PASS.
- revised forbidden scaffolding search: PASS, no matches.
- revised fence count: PASS, plan file 26 fences/even, WORK file 0 fences/even.
- revised T01-T12 and target path presence: PASS.
- revised duplicate long paragraph scan excluding fenced code blocks: PASS, 0 duplicates.
- changed path review: PASS, only plan move `interviews/db-transactions-deep-study-plan.md -> database/db-transactions-deep-study-plan.md` and this WORK file are in scope; `database/mvcc.md` and `database/lock.md` were read but not edited in this tranche.

## 11. Closure

- requested closure scope: whole DB/금융 트랜잭션 deep-study writing request
- achieved closure scope: revised planning tranche only
- whole-request completion verdict: `REQUEST_PARTIAL`
- remaining executable count: 12 primary writing units plus hub creation
- next immediate target: `database/mvcc.md` preservation/restructure ledger, then T01-T06 database core concepts writing
- unrelated open work: repo has many pre-existing modified/untracked files; this work must stage only its own files
