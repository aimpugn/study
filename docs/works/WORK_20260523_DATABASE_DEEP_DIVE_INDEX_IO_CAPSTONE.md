# WORK 2026-05-23 - Database deep dive index I/O capstone

## Request Normalization

- Surface request: 세션 `019e401f-1c8f-7eb0-b724-639d9e807a73`의 DB deep-dive 정리 흐름을 최신 파일 상태 기준으로 재개합니다.
- Protected intent: `interviews/database-deep-dive`를 canonical 면접 학습 표면으로 두고, DB를 page/buffer I/O, index/optimizer, transaction, replication, operations까지 머릿속에서 재생 가능한 형태로 설명합니다.
- Whole objective: 전체 DB corpus를 계속 개선 가능한 canonical 학습 자산으로 유지합니다.
- This tranche: 사용자 benchmark floor 중 `인덱스는 항상 빠르지 않다`를 `04-index-query-optimizer.md`에서 page I/O, random/sequential access, covering/index-only scan, optimizer 선택 기준이 함께 보이는 직접 재생 절로 보강합니다.
- Non-scope: 전체 DB corpus 완료 선언, `database/` source reservoir 재작성, unrelated dirty file stage/revert, push.

## Instruction Stack

- Applied: `/Users/rody/VscodeProjects/study/AGENTS.md`, `/Users/rody/VscodeProjects/study/interviews/AGENTS.md`, `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md`, `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/USECASE.md`, `/Users/rody/VscodeProjects/study/interviews/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/interviews/USECASE.md`.
- Applied skills: `multi-agent`, `dialectic-kernel`, `review-kernel`, with `rigorous-task`, `study-explanation`, and `humanize-korean` as supporting execution contracts.
- Markdown rule: nested list indent remains 4 spaces per `interviews/.markdownlint.json`.

## Current-State Evidence

- Repo boundary: `/Users/rody/VscodeProjects/study`, cwd `/Users/rody/VscodeProjects/study/interviews`.
- Canonical target: `interviews/database-deep-dive/`.
- Source reservoir: `database/` and historical `database/deep-dive/`; not canonical output.
- Existing structural validator: `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py`.
- Existing status note: `docs/works/WORK_20260522_DATABASE_DEEP_DIVE_CHECKPOINT_RECOVERY.md` records checkpoint recovery, but this tranche re-judges current files instead of assuming that WORK is current completion proof.

## Frozen Checklist

- Latest files and validator state are checked before edits.
- Previous session is used only for structure signals, not copied transcript.
- Next executable tranche is bounded and separated from whole objective.
- Patch is path-limited to canonical index/optimizer learning surface and its audit ledger.
- Explanation is prose-first Korean, with concrete SQL/EXPLAIN trace and PASS/FAIL reading criteria.
- Claim card and composition audit make the new benchmark-floor repair traceable.
- TOC is regenerated.
- Validator, `git diff --check`, and path-limited final review pass before commit.
- Commit excludes unrelated dirty files and staged unrelated WIP.

## Debate And Claim Cards

### Claim Card DBI-INDEX-01-C04

- Claim: 인덱스가 빠른지는 인덱스 사용 여부가 아니라 후보 row 수, page 방문 수, random/sequential 접근, buffer hit/read, heap fetch, sort/temp, 통계 추정 정확도가 함께 결정합니다.
- Support: existing canonical `04-index-query-optimizer.md`, `02-storage-pages-buffer-io.md`, PostgreSQL index/EXPLAIN documentation references already present in the corpus, and Linux page-cache boundary rows in the audit ledger.
- Critic attack: `Index Scan`, `Seq Scan`, `Index Only Scan` 같은 node 이름만으로 판단하면 실제 page locality, heap visibility, clustered lookup, cache state, 데이터 분포를 지울 수 있습니다.
- Resolution: `ACCEPT_REPAIR`. 직접 재생 절에 선택도 높은 index-only path, 낮은 선택도 sequential/bitmap path, heap fetch 반례를 함께 넣어 node-name shortcut을 막습니다.

### Composition Card DBI-INDEX-01-S05A

- Composition risk: 작은 단위 설명은 이미 있었지만, 사용자 질문처럼 "개별 데이터를 각각 디스크에서 가져오는 경우와 뭉치로 읽는 경우"를 면접장에서 바로 재생할 수 있는 한 덩어리 trace가 부족할 수 있습니다.
- Critic attack: 기존 본문을 충분하다고만 판단하면 사용자가 지적한 설명 수준 실패가 반복됩니다.
- Resolution: `ACCEPT_REPAIR`. 기존 내용을 덧칠하지 않고 `직접 재생해 보기` 안에 capstone을 추가해 기존 deep mechanism과 실험 경로를 연결합니다.

## Execution Log

- Added `### 인덱스가 항상 빠르지 않은 이유를 EXPLAIN으로 재생하기` to `04-index-query-optimizer.md`.
- Added `DBI-INDEX-01-C04` to `audit/claim-audit.tsv`.
- Added `DBI-INDEX-01-S05A` to `audit/composition-audit.tsv`.
- Regenerated the TOC with `update_markdown_tocs.py`.

## Verification Log

- PASS: `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py`.
- PASS: `git diff --check -- interviews/database-deep-dive/04-index-query-optimizer.md interviews/database-deep-dive/audit/claim-audit.tsv interviews/database-deep-dive/audit/composition-audit.tsv docs/works/WORK_20260523_DATABASE_DEEP_DIVE_INDEX_IO_CAPSTONE.md`.
- PASS: path-limited final review found and repaired two issues before closure:
    - The first draft repeated the existing `orders_plan_lab` DDL, so the capstone now reuses the existing table and only adds distribution/index setup.
    - The first draft used inconsistent status-index names, so the SQL and plan examples now consistently use `orders_status_only_idx`.
- READY: path-limited commit. The final commit hash is recorded by Git after this WORK is written.

## Closure

- Requested closure scope: resume current DB deep-dive flow and, if safe, close the next executable tranche through commit without push.
- Achieved closure scope: index/page-I/O capstone and audit rows are implemented, validator-clean, and ready for path-limited commit.
- Whole-request verdict: not complete. This tranche improves one index/page-I/O benchmark-floor gap and does not claim the entire DB corpus is finished.
- Remaining open tranche disclosure: future DB corpus improvements must be selected from current evidence and benchmark gaps, not from stale session assumptions.
