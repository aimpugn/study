# WORK 20260520 Database Interview Deep Dive Reconstruction

## Request Normalization

- whole-request objective: 기존 `study/database` 자료를 source로 삼아, 면접 준비용 DB 심화 학습 자산을 `interviews/database-deep-dive` 아래에 새로 구성한다.
- protected intent: 단순 복붙, 위치 이동, 길이 충족용 확장이 아니라, 문장/claim 단위 진위와 문서 전체 구성의 진위를 함께 검토해 독자가 더 높은 수준으로 이해하도록 돕는다.
- current tranche: canonical root, audit protocol, validator, first pilot monograph.
- explicit non-complete boundary: 전체 DB 면접 심화 코퍼스는 아직 complete가 아니다.

## Project Overlay

- global runtime rules: user-provided AGENTS contract, multi-agent, dialectic-kernel, review-kernel.
- repo overlay: `/Users/rody/VscodeProjects/study/AGENTS.md`.
- project facts: `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/USECASE.md`, `/Users/rody/VscodeProjects/study/interviews/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/interviews/USECASE.md`.
- applied reason: this is substantial study-doc reconstruction with repo changes and completion-risk.

## Roster

| alias | canonical role | responsibility | status |
|---|---|---|---|
| Orchestrator/Builder | main Codex | normalize request, implement bounded files, synthesize reviews | active |
| Peirce | Source Audit Architect | classify source/draft boundary and claim audit structure | completed first pass |
| Sartre | Composition Critic | attack whole-document truth, structure, and Korean learning flow | completed first pass |
| Kant | Protocol Sentinel | stop premature complete, path-limited commit, closure honesty | completed first pass |

## Dialectic Rounds

### R1. Problem framing

- claim: canonical output belongs under `interviews/database-deep-dive`.
- attack: old generated corpus under `database/deep-dive` already exists and passed its old validator.
- response: ACCEPT_REPAIR. Old corpus is retained as noncanonical draft/coverage evidence, not promoted.
- completion marker: `README.md` and old-path notices added.

### R2. Evidence breadth

- claim: `study/database` is source, not target.
- attack: file-level inventory can become ceremony without improving prose.
- response: REBUT. Boundary classification prevents exactly the path drift the user caught.
- completion marker: generated `audit/source-boundary.tsv`.

### R3. Checklist stress

- claim: claim audit plus composition audit is the minimum useful protocol.
- attack: sentence audit alone may look more concrete.
- response: REBUT. Sentence audit detects local truth; composition audit detects false aggregation, ordering drift, and DBMS boundary collapse.
- completion marker: `claim-audit.tsv` and `composition-audit.tsv`.

### R4. Execution result

- claim: a WAL/redo/undo/PITR pilot can falsify the protocol because it contains many true-small/false-large risks.
- attack: writing one pilot may be mistaken for whole completion.
- response: DOWNGRADE. Marked as pilot-complete and PARTIAL overall.
- completion marker: topic registry distinguishes `pilot-complete` and `planned`.
- critic repair: generated draft material was downgraded from domain-truth `T1` evidence, official source rows were added, and the validator now rejects generated draft `T1` domain evidence.

### R5. Closure

- claim: corrective commit is allowed after structural validation and review.
- attack: dirty working tree makes broad staging unsafe.
- response: ACCEPT_REPAIR. Use path-limited staging only.
- completion marker: pending path-limited commit after final staged-file check.

## Frozen Checklist

- [x] active instruction stack and project facts checked.
- [x] multi-agent/review/dialectic topology activated.
- [x] canonical path created under `interviews/database-deep-dive`.
- [x] old `database/deep-dive` path marked noncanonical.
- [x] claim audit and composition audit files created.
- [x] structural validator created.
- [x] first pilot document written.
- [x] validator run.
- [x] critic/sentinel final review and repair.
- [ ] path-limited commit.
- [ ] push only if explicitly requested for this corrective commit.

## Closure Scope

- requested closure scope: proceed with corrected multi-agent reconstruction.
- achieved closure scope: pending validation; planned as corrective structural plus one pilot document.
- achieved closure scope after repair: corrective structural plus one pilot document, with partial validator PASS and intentional whole-validator FAIL because planned topics remain.
- whole-request completion verdict: NOT_WHOLE_COMPLETE.
- remaining count: at least 2 planned registry topics plus the rest of DB interview corpus decomposition.
- next immediate target: validate and review `wal-redo-undo-crash-recovery-pitr.md`.
