# WORK_20260520_DATABASE_DEEP_DIVE_POLITE_OS_VISUAL_REWRITE

## 0. Meta

- 작업 제목: database-deep-dive 존댓말, OS 관점, 시각화, TOC 보강
- WORK 파일 경로: `docs/works/WORK_20260520_DATABASE_DEEP_DIVE_POLITE_OS_VISUAL_REWRITE.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs | explain | audit`
- 작업 깊이: `full`
- 관련 요청: `interviews/database-deep-dive` 전체 문서를 면접 준비용 deep-dive 자산으로 완결도 있게 보강
- 대상 경로 / 자산: `interviews/database-deep-dive/01-*.md`부터 `14-*.md`, `README.md`, `topic-registry.tsv`, `audit/*.tsv`, `tools/*.py`, 필요 시 `assets/*`
- 시작 일시: 2026-05-20 Asia/Seoul
- 현재 상태: `COMPLETE_PENDING_PATH_LIMITED_COMMIT`
- 완료 게이트: `PASS_AFTER_VALIDATION`
- finish: `test+commit`

## 1. Request Normalization

- goal: 기존 database deep-dive 문서를 존댓말과 자연스러운 한국어로 다듬고, 기술적 설명과 OS 관점, 시각 자료, 목차 품질을 함께 끌어올립니다.
- refs: `database-deep-dive`, `.markdownlint.json`, repo `AGENTS.md`, `PROJECT_INTENT.md`, `USECASE.md`, `AGENTS_WORK_TEMPLATE.md`, 명시된 `humanize-korean`, `multi-agent`, `dialectic-kernel`, `review-kernel`
- scope: 14개 numbered deep-dive 문서 전체와 이를 검증하는 도구/감사 표면
- mode: `execute`
- run_mode: `normal`
- finish: `final review + validation + path-limited commit`
- must_keep: 기존 사실, 코드, SQL, 명령, 공식 링크, 문서 간 링크, 최근 transaction propagation 보강 내용, 읽기 순서 번호
- extra_checks: H3 heading TOC 포함, TOC nested indent 4 spaces, OS/kernel/scheduler/memory/page-cache/block-layer/fsync 설명, 학습자 관점 시각화, 별도 staged GIF 보존

### 1.1 Explicit Deliverables

- 모든 reader-facing 문장을 가능한 한 존댓말로 바꾸되, 기술 용어와 식별자는 보존합니다.
- `humanize-korean` 관점으로 번역투, 딱딱한 관리 문체, 부자연스러운 평서형을 줄입니다.
- 기술 설명을 보강하고, 필요한 곳에는 ASCII diagram 또는 이미지 자산을 추가합니다.
- DB가 서버 OS 위에서 실행된다는 관점을 깊게 설명합니다. 최소 범위는 kernel, scheduler, memory/page cache, filesystem/block layer, driver, disk read/write, flush/fsync입니다.
- `###` 소제목을 TOC에 포함하고, `.markdownlint.json` 규칙에 맞춰 nested list는 4칸 들여씁니다.
- `multi-agent + dialectic-kernel + review-kernel` 방식으로 builder/critic/sentinel 역할을 분리하고, 결과에 대해 학습자 관점의 비판과 수리를 거칩니다.
- 검증 후 path-limited commit을 만듭니다. push는 이번 요청에 명시되지 않았으므로 하지 않습니다.

### 1.2 Non-Goals

- `study/database`의 기존 학습 자료를 직접 재작성하지 않습니다. 필요하면 source로만 씁니다.
- unrelated dirty files와 기존 staged GIF는 건드리지 않습니다.
- 모든 설명에 이미지를 억지로 넣지 않습니다. 이해를 실제로 돕는 시각화만 추가합니다.

## 2. Root-First Framing

- 근본 문제: 현재 문서가 정보량은 충분해도 문장 register, TOC 탐색성, OS 하부 계층 설명, 시각적 state movement가 약하면 면접 준비자가 머릿속에 실행 경로를 재생하기 어렵습니다.
- 작업 목표: 짧은 답변과 깊은 본문이 모두 살아 있고, DBMS 내부와 OS 계층을 함께 설명할 수 있는 학습 자산으로 승격합니다.
- 성공 정의:
    - 14개 numbered 문서 전부에서 H3 TOC 검증이 통과합니다.
    - 존댓말/자연스러운 한국어 전환이 reader-facing prose에 적용됩니다.
    - OS 관점 설명과 시각화가 최소한 foundation, storage, WAL/recovery, operations 축에 들어갑니다.
    - 공식/1차 근거가 필요한 새 기술 주장은 근거 링크나 명시적 추론 경계를 가집니다.
    - 구조 validator, diff check, 문서 링크/자산 검증, 최종 critic/sentinel 검수가 통과합니다.
    - unrelated staged/dirty 변경을 섞지 않고 경로 제한 commit을 만듭니다.
- PARTIAL 조건: 일부 문서의 존댓말 전환이나 OS 설명 보강이 열려 있거나, 검증이 통과하지 못했을 때입니다.
- BLOCKED 조건: 파일 접근, 검증 도구 실행, 또는 공식 근거 확인이 불가능해 신뢰 가능한 완료 판정을 할 수 없을 때입니다.

## 3. Multi-Agent Roster And Dialectic Claims

- Orchestrator: Codex parent. 범위, WORK, 공통 검증, 최종 통합, commit 담당.
- Builder A / Curie: `01-04` 문서 보강.
- Builder B / Russell: `05-09` 문서 보강.
- Builder C / Bohr: `10-14` 문서 보강.
- Critic + Protocol Sentinel / Huygens: 전체 문서, TOC, OS 설명, 시각화, 검증 경계 read-only 검수.

### Claim Cards

- C-01: `###` heading을 TOC에 포함하지 않으면 긴 문서의 학습 탐색성이 약해집니다.
    - support tier: `T1 Direct Evidence`
    - evidence refs: `.markdownlint.json`, 현재 14개 문서의 H3 count와 H2-only TOC 관측
    - falsifier: 모든 H3가 TOC에 이미 4칸 indent로 들어 있다면 기각됩니다.
    - admission lane: `APPLY`
- C-02: DB deep-dive에서 OS 계층 설명이 약하면 durability와 latency 설명이 반쪽이 됩니다.
    - support tier: `T1 Direct Evidence + T2 Strong Inference`
    - evidence refs: PostgreSQL WAL/fsync docs, Linux `write(2)`/`fsync(2)`, Linux blk-mq docs, MySQL InnoDB I/O docs
    - falsifier: 문서가 이미 write/fsync/page cache/block layer/scheduler 상호작용을 충분히 설명한다면 범위를 축소합니다.
    - admission lane: `APPLY`
- C-03: 전면 존댓말 전환은 사실 보존과 문장 자연스러움이 함께 닫힐 때만 성공입니다.
    - support tier: `T2 Strong Inference`
    - evidence refs: `humanize-korean` strict workflow, repo 설명 계약
    - falsifier: 존댓말 변환이 코드/SQL/용어 의미를 바꾸거나 과윤문으로 기술 주장을 훼손하면 재작업합니다.
    - admission lane: `APPLY with same-loop review`

## 4. Evidence Ledger

- E-01: Linux `write(2)`는 성공 반환만으로 데이터가 디스크에 커밋되었음을 보장하지 않으며, 확실히 하려면 쓰기 이후 `fsync(2)`가 필요하다고 설명합니다.
- E-02: Linux `fsync(2)`는 수정된 in-core data와 파일 metadata를 디스크 또는 영구 저장 장치로 flush하고, 장치가 완료를 보고할 때까지 block합니다.
- E-03: PostgreSQL WAL 설정 문서는 `fsync`가 OS/hardware crash 뒤 일관된 상태 복구를 위해 물리 기록을 보장하려고 쓰이며, `wal_sync_method`별로 kernel cache와 sync 경계가 달라진다고 설명합니다.
- E-04: PostgreSQL WAL configuration은 checkpoint가 쓴 OS page가 page cache에 남아 있다가 checkpoint 끝의 `fsync`에서 stall을 만들 수 있다고 설명합니다.
- E-05: Linux blk-mq 문서는 userspace I/O가 block layer에서 software staging queue와 hardware dispatch queue를 거쳐 device driver로 내려간다고 설명합니다.
- E-06: MySQL InnoDB doublewrite buffer 문서는 buffer pool에서 flush된 page를 data file 위치에 쓰기 전에 doublewrite area에 쓰고 `fsync()`를 사용해 torn page recovery 안전성을 높인다고 설명합니다.
- E-07: PostgreSQL monitoring stats는 `pg_stat_io`, `pg_stat_wal`, checkpoint write/sync time처럼 DB 내부 지표가 OS I/O 경계와 연결되는 관측점을 제공합니다.
- E-08: Linux scheduler docs는 runnable task 선택, vruntime, enqueue/dequeue/pick_next_task 같은 개념을 제공하여 DB process/thread가 CPU를 얻거나 기다리는 경계를 설명할 수 있게 합니다.

## 5. Frozen Checklist

- [x] 14개 numbered 문서가 모두 존댓말 중심의 reader-facing prose로 정리되었습니다.
- [x] 새 기술 주장은 공식/1차 자료 또는 명시적 추론 경계로 지지됩니다.
- [x] OS/kernel/scheduler/memory/page-cache/filesystem/block-layer/driver/disk/flush/fsync 설명이 foundation, storage, WAL/recovery, operations 축에 반영되었습니다.
- [x] 이해를 돕는 ASCII diagram 또는 이미지가 추가되었고, 장식이 아니라 state movement를 설명합니다.
- [x] 모든 numbered 문서 TOC에 H2와 H3가 들어가며, H3 항목은 4 spaces로 들여쓰기됩니다.
- [x] `.markdownlint.json`의 4-space indent 규칙을 훼손하지 않습니다.
- [x] 기존 문서 간 cross-link와 읽기 순서 numbering이 유지됩니다.
- [x] 최근 08 문서의 transaction propagation/race condition 보강을 보존합니다.
- [x] validator가 H3 TOC 누락과 nested indent 위반을 잡습니다.
- [x] structural validator, link/asset check, `git diff --check`가 통과합니다.
- [x] Critic/Sentinel이 definition, criteria, checklist, result를 최소 1회 확인합니다.
- [x] unrelated dirty files와 기존 staged GIF를 commit에 섞지 않습니다.
- [x] path-limited commit을 만듭니다.

## 6. Council Rounds

- R1 problem framing: C-01, C-02, C-03을 적용 claim으로 고정했고, critic/sentinel이 작업 정의와 성공 기준을 확인했습니다.
- R2 evidence breadth: Linux `write(2)`, `fsync(2)`, block writeback cache, blk-mq, CFS, page cache, PostgreSQL WAL/monitoring, MySQL InnoDB I/O/doublewrite 근거를 audit surface에 추가했습니다.
- R3 checklist stress: H3 TOC, 4-space indent, image link existence, evidence ref resolution을 validator와 markdownlint로 닫았습니다.
- R4 execution result: 14개 numbered 문서 전체에 존댓말 전환, H3 TOC, OS 관점 설명, cross-link, ASCII/이미지 보강을 적용했습니다.
- R5 closure: final validation은 통과했고, unrelated staged GIF는 path-limited staging/commit으로 제외합니다.

## 7. Verification Plan

- `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py`
- H3 TOC coverage validator 확장 또는 별도 검증 스크립트
- Markdown image/link asset existence check
- `git diff --check -- interviews/database-deep-dive docs/works/WORK_20260520_DATABASE_DEEP_DIVE_POLITE_OS_VISUAL_REWRITE.md`
- `git diff --cached --name-status`로 unrelated staged file 보존 확인

## 8. Verification Results

- `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py` -> PASS. H2/H3 TOC, 4-space H3 indentation, local image links, audit refs를 함께 확인했습니다.
- `git diff --check -- interviews/database-deep-dive docs/works/WORK_20260520_DATABASE_DEEP_DIVE_POLITE_OS_VISUAL_REWRITE.md` -> PASS.
- `npx --yes markdownlint-cli@0.44.0 --config interviews/.markdownlint.json "interviews/database-deep-dive/**/*.md" "docs/works/WORK_20260520_DATABASE_DEEP_DIVE_POLITE_OS_VISUAL_REWRITE.md"` -> PASS.
- Mixed Korean/ASCII broad residual scan for impolite sentence-final `~다.` candidates across numbered docs, including trace/code blocks -> PASS, 0 candidates.
- `rg -n '^ {1,3}- ' interviews/database-deep-dive/*.md docs/works/WORK_20260520_DATABASE_DEEP_DIVE_POLITE_OS_VISUAL_REWRITE.md` -> PASS, no 1-3 space nested list candidates.

## 9. Critic Repair Ledger

- Critic/Sentinel first result: `REWORK`. Findings were H3 TOC coverage gap, partial 존댓말 conversion, missing OS breadth in several documents, unlinked visuals, and missing OS claim audit rows.
- Repair: added `update_markdown_tocs.py`, extended validator, regenerated TOCs, linked `db-os-io-stack.svg` and `wal-fsync-durability-path.svg`, added OS evidence and claim audit rows, and revised all numbered documents.
- Critic/Sentinel second result: `PARTIAL`. Remaining blockers were a small set of reader-facing plain endings and stale WORK state.
- Repair: fixed the cited prose lines and ran a broader residual scan that includes trace/code-block prose. Updated this WORK ledger from pending state to validation-bound closure state.
- Critic/Sentinel final result: `COMPLETE`. Checklist/result critic-confirmation is closed; no blocker remains before path-limited commit.

## 10. Closure Scope

- requested closure scope: `interviews/database-deep-dive` numbered corpus, validation/audit surface, and this WORK record.
- achieved closure scope: 14/14 numbered documents revised and validated; new visual assets linked; validator updated; audit evidence extended.
- remaining count in requested corpus: `0`.
- next immediate target: path-limited staging and commit of only `interviews/database-deep-dive` plus this WORK file.
- remaining open tranche disclosure: no requested-corpus blocker remains. Existing unrelated dirty files and the pre-existing staged programmers GIF remain outside this task and must not be included in this commit.
