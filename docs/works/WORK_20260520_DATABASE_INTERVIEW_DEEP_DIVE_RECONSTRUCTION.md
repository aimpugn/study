# WORK 20260520 Database Interview Deep Dive Reconstruction

## Request Normalization

- whole-request objective: 기존 `study/database` 자료와 면접 source를 원자료로 삼아, DB 면접 심화 학습 자산을 `interviews/database-deep-dive` 한 곳에 완결된 정식 문서 세트로 구성한다.
- protected intent: 단순 복붙, 위치 이동, 길이 충족용 확장이 아니라, 작은 claim의 진위와 문서 전체 구성의 진위를 함께 검토해 독자가 면접장에서 짧게 답하고 꼬리 질문에서 깊게 내려갈 수 있게 한다.
- current tranche: full canonical DB interview deep-dive corpus, audit protocol, validator hardening, final closure.
- explicit source boundary: `study/database/**`, `interviews/source/**`, `interviews/database-storage-search-nosql.md`, 이전 `database/deep-dive/**`는 source 또는 generated draft reservoir이다. 정식 본문 위치는 `interviews/database-deep-dive/**`다.

## Project Overlay

- global runtime rules: user-provided AGENTS contract, multi-agent, dialectic-kernel, review-kernel.
- repo overlay: `/Users/rody/VscodeProjects/study/AGENTS.md`.
- project facts: `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/USECASE.md`, `/Users/rody/VscodeProjects/study/interviews/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/interviews/USECASE.md`.
- applied reason: this is substantial study-doc reconstruction with repo changes, whole-request completion risk, and explicit multi-agent review requirement.

## Roster

| alias | canonical role | responsibility | status |
|---|---|---|---|
| Orchestrator | main Codex | normalize request, integrate worker commits, repair audit/validator, path-limited closure | active |
| Peirce | Source Audit Architect | classify source/draft boundary and source coverage risk | reviewed; material findings repaired or bounded |
| Sartre | Composition Critic | attack whole-document truth, repetition, DBMS boundary, validator adequacy | reviewed; material findings repaired or validator-backed |
| Kant | Protocol Sentinel | stop premature complete, WORK/current mismatch, path-limited commit/push | reviewed; material findings repaired or disclosed |
| Faraday | Builder A | model/storage/index/schema documents | completed commit `2216cfc` |
| Jason | Builder B | transaction/MVCC/lock/replication/partition documents | completed commit `1a56b96` |
| Turing | Builder C | engine/application/operations/search documents | completed commit `902526b` |

## Dialectic Rounds

### R1. Problem framing

- claim: canonical output belongs under `interviews/database-deep-dive`.
- attack: old generated corpus under `database/deep-dive` already exists and passed its old validator.
- response: ACCEPT_REPAIR. Old corpus is retained as noncanonical draft/coverage evidence, not promoted.
- completion marker: `README.md`, `topic-registry.tsv`, and source boundary audit identify the canonical target and source/draft reservoirs.

### R2. Evidence breadth

- claim: `study/database` is source, not target.
- attack: file-level inventory can become ceremony without improving prose.
- response: REBUT. Boundary classification prevents path drift and prevents generated drafts from becoming T1 domain truth.
- completion marker: `audit/source-boundary.tsv`, `audit/claim-audit.tsv`, `audit/evidence-refs.tsv`.

### R3. Checklist stress

- claim: claim audit plus composition audit is the minimum useful protocol.
- attack: sentence audit alone may look more concrete.
- response: REBUT. Sentence audit detects local truth; composition audit detects false aggregation, ordering drift, DBMS boundary collapse, and source overgeneralization.
- completion marker: topic-specific `large_unit_risk` rows in `audit/composition-audit.tsv` and validator cross-checks.

### R4. Execution result

- claim: 14 canonical documents close the DB interview deep-dive corpus for this request.
- attack: structural validator could pass even with repeated filler or sensitive source leakage.
- response: ACCEPT_REPAIR. Repeated generic reader phrases, duplicate H3 headings, repeated long paragraphs, repeated audit risk rows, and sensitive source units are validator failures.
- completion marker: `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py` passes after repair.

### R5. Closure

- claim: final commit and push are allowed after review and validation.
- attack: dirty working tree has many unrelated files, and `interviews/database-storage-search-nosql.md` remains a dirty source file outside canonical target.
- response: ACCEPT_REPAIR. Use path-limited staging for `interviews/database-deep-dive/**` and this WORK file only. Do not stage unrelated dirty files or the source reservoir file.
- completion marker: pending path-limited stage, `git diff --cached --name-status`, commit, push.

## Frozen Checklist

- [x] active instruction stack and project facts checked.
- [x] multi-agent/review/dialectic topology activated.
- [x] canonical path fixed under `interviews/database-deep-dive`.
- [x] source/draft boundary recorded.
- [x] 14 finite DB interview deep-dive topics fixed and marked complete.
- [x] 14 canonical documents written in the canonical directory.
- [x] each canonical document has required 8 learning sections.
- [x] each canonical document is at least 20,000 characters excluding README/validation.
- [x] old generated draft material is not used as T1 domain truth.
- [x] sensitive source files are not listed as completed topic source units.
- [x] repeated generic filler, duplicate H3 headings, and repeated long paragraphs are validator-blocked.
- [x] composition audit uses topic-specific large-unit risks.
- [x] validator run without `--allow-planned`.
- [x] critic/sentinel findings reviewed and repaired.
- [ ] path-limited final stage.
- [ ] path-limited final commit.
- [ ] push to `origin/main`.

## Verification Log

- `python3 interviews/database-deep-dive/tools/build_source_boundary.py`
- `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py`
- `wc -m interviews/database-deep-dive/*.md`
- repeated generic phrase scan over canonical reader docs.
- duplicate H3 and repeated long paragraph scan.
- section ordering scan for the 8 required headings.
- `rg` scan for reader-facing generation meta in canonical docs.

Latest validator verdict before final stage: PASS.

## Closure Scope

- requested closure scope: whole DB interview deep-dive reconstruction, including each document's complete cleanup, multi-agent review, final verification, commit, and push.
- achieved closure scope before final stage: 14 canonical documents complete; planned topic count 0; source boundary, claim audit, composition audit, and validator hardened; material critic findings repaired.
- whole-request completion verdict before final stage: READY_FOR_COMMIT_AND_PUSH.
- remaining count inside requested canonical corpus: 0.
- remaining open disclosure: unrelated dirty files remain in the repository and are intentionally excluded from the final canonical corpus commit. `interviews/database-storage-search-nosql.md` remains source-reservoir dirty state and is not staged for this closure unless explicitly requested later.
- next immediate target: path-limited stage, staged diff audit, commit, push.
