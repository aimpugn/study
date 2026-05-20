# Orchestrator findings

## Verdict

`REWORK -> REPAIRED -> READY_FOR_FINAL_COMMIT`.

정식 경로는 `interviews/database-deep-dive/`다. `database/deep-dive/`는 이전 실행에서 만들어진 장문 초안과 coverage 자료로 보존하되, 정식 인터뷰 학습 산출물로 그대로 승격하지 않는다.

## Multi-agent synthesis

- Source Architect: source와 generated draft의 경계를 먼저 나누지 않으면 잘못된 위치 문제와 T1 근거 과장을 반복한다. `database/deep-dive/**`는 generated draft/noncanonical로만 쓴다.
- Composition Critic: 작은 claim이 각각 맞아도 반복, 순서, DBMS 경계, log 소비자 경계가 흐리면 큰 설명은 틀린다. 반복 filler 제거와 topic-specific composition risk가 필요하다.
- Protocol Sentinel: validator PASS, WORK ledger, staged path, commit/push 상태가 서로 맞아야 한다. dirty working tree가 크므로 path-limited staging이 closure 조건이다.
- Builders: 14개 canonical 문서를 병렬로 작성했다. 최종 오케스트레이션에서 반복 문구, sensitive source unit, audit/validator mismatch를 수리했다.

## Final corpus

- `01-database-system-mental-model.md`
- `02-storage-pages-buffer-io.md`
- `03-wal-redo-undo-crash-recovery-pitr.md`
- `04-index-query-optimizer.md`
- `05-schema-constraints-migration.md`
- `06-transaction-acid-boundary.md`
- `07-mvcc-snapshot-visibility.md`
- `08-isolation-lock-deadlock.md`
- `09-replication-lag-backup-failover.md`
- `10-partition-sharding-distributed-sql.md`
- `11-mysql-postgresql-engine-deep-dive.md`
- `12-application-boundaries-idempotency-money-outbox.md`
- `13-operations-security-troubleshooting.md`
- `14-search-document-nosql-engine.md`

## Material findings and repairs

### Source boundary finding

- finding: `DBI-OPS-01` listed a sensitive source file as a completed topic source unit.
- repair: removed the sensitive source from `topic-registry.tsv`; source boundary still classifies it as `sensitive-source-do-not-promote` so future audits can see why it is excluded.
- validator: completed topics now fail if any source unit is classified as sensitive.

### Repetition finding

- finding: several generated sections previously used repeated generic transfer, pass/fail, tail-answer, and trap-answer paragraphs.
- repair: removed repeated generic paragraphs and added topic-specific mechanism and scenario sections where length dropped below the deep-dive floor.
- validator: reader docs now fail on known generic repeated phrases, duplicate H3 headings, and repeated long paragraphs.

### Composition audit finding

- finding: `composition-audit.tsv` repeated the same large-unit risk across many sections and therefore did not prove topic-specific composition review.
- repair: regenerated composition audit rows with topic-specific large-unit risks, counterexamples, missing context checks, and repair actions for all 14 topics.
- validator: composition audit now fails if the same large-unit risk is repeated across too many rows.

### Evidence finding

- finding: a few claim rows used official references that were too broad for the exact claim.
- repair: added narrower official evidence rows for PostgreSQL constraints, MySQL online DDL, Flyway migrations, Spring transactions, and Java BigDecimal; downgraded application/reliability synthesis claims where they combine multiple official references and design inference.
- validator: claim verification refs must resolve in `audit/evidence-refs.tsv`.

### Protocol finding

- finding: WORK still described a pilot/partial state while the current corpus had advanced to 14 complete topics.
- repair: WORK now records the whole-request objective, finite corpus, repaired critic findings, validator result, remaining count 0 inside the requested canonical corpus, and path-limited final stage requirement.

## Verification result

- structural/source/audit validator: `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py` -> PASS.
- repeated generic phrase scan -> PASS.
- duplicate H3 scan -> PASS.
- repeated long paragraph scan -> PASS.
- required heading order scan -> PASS for all 14 canonical documents.
- reader-facing generation meta scan -> PASS for canonical reader docs.

## Residual risk

- Official docs are linked as current primary sources, but actual production DB versions can differ. The final documents state or imply current-doc grounding and keep version-specific implementation details as DBMS-specific boundaries.
- `interviews/database-storage-search-nosql.md` remains a source reservoir and currently has unrelated dirty changes in the worktree. It is intentionally not part of the final canonical corpus commit.
