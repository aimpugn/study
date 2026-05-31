# WORK 2026-05-31 - OS kernel and distributed systems learning corpus

## Current State

Status: READY_FOR_SCOPED_COMMIT

Whole-request objective: `OS 커널과 분산 시스템 원리, Kafka/Cassandra/Spark 내부 구조를 초보자에서 전문가 수준까지 학습할 수 있는 self-contained Markdown 학습 코퍼스를 repo 안에 작성하고, 사용자의 prose-first deep study 품질 하한을 만족하도록 전면 재작성한 뒤 최종 review와 검증 후 commit까지 수행한다.`

Whole-request completion verdict: content/review/verification gates PASS; scoped commit is the only remaining closure action. Push remains out of scope.

## Project Overlay

Active instruction stack checked:

| Layer | Path | Applied decision |
|---|---|---|
| Global runtime contract | `~/.codex/AGENTS.md` | Applied: substantial repo-changing work, checklist freeze, claim review, no premature whole closure, commit after review and verification. |
| Global WORK template | `~/.codex/AGENTS_WORK_TEMPLATE.md` | Applied: work ledger shape, gate recording, review and verification surfaces. |
| Study repo contract | `../AGENTS.md` | Applied: deep study monograph, self-reconstructable Korean explanations, WORK under `docs/works/`, final review + verification + commit. |
| Study WORK template | `../AGENTS_WORK_TEMPLATE.md` | Applied: local work ledger path and closure expectations. |
| Interviews overlay | `AGENTS.md` | Applied: interview preparation must support short answer plus lower-layer follow-up reasoning. |
| Project facts | `PROJECT_INTENT.md`, `USECASE.md` | Applied: not memorization; teach-back, follow-up questions, unknown-question reasoning, broad-topic decomposition. |
| User request | Current thread | Applied: create real Markdown files, no forbidden diagram syntax, source ledger, self-contained corpus, final commit, no push. |

Latest user clarification:

- `아뇨 중복이어도 그냥 그 문서 자체로 완벽하게 정리해 주세요.`
- 기존 corpus는 질문/직관/작은 예시/상태 이동 trace/내부 메커니즘/실패 모드 같은 반복 라벨을 채운 수준에 가깝고, 사용자는 이를 100점 만점 30~40점 정도로 평가했다.
- `/Users/rody/VscodeProjects/study/database/mvcc.md`는 깊이와 역사/맥락의 하한선이지만, 반복과 목록식 약점까지 복제하면 안 된다.
- 설명은 독자가 읽는 중 `경계가 뭔데?`, `어떻게 통과하는데?`, `그래서 실제로 어떻게 내려가는데?`라고 되묻지 않도록 문장과 문단으로 닫아야 한다.
- `실패 모드` 같은 직역투 라벨은 독자용 본문에서 지배 구조가 되면 안 된다. 필요한 내용은 `어디서 깨지는가`, `무엇을 확인해야 하는가`, `면접에서 어떻게 다시 내려가는가`를 자연스러운 한국어 문단 안에 녹인다.
- Gold slice는 `01_os_kernel_foundations.md`의 `write(fd, buf, len)` / syscall / file descriptor / VFS / page cache / fsync 흐름이다. 이 slice가 통과해야 나머지 문서에 같은 기준을 확장한다.

Decision: existing docs may be consulted, but this corpus must be self-contained. Duplicate coverage is acceptable and sometimes required. Existing docs are not allowed to replace content inside the new corpus. The earlier label-heavy corpus is treated as source material, not as an accepted quality baseline.

## Scope Freeze

New corpus root:

`interviews/os-kernel-distributed-systems-deep-dive/`

Write set:

| File | Role |
|---|---|
| `os-kernel-distributed-systems-deep-dive/README.md` | Entry page and direct pointer to the reading path. |
| `os-kernel-distributed-systems-deep-dive/00_index_and_learning_path.md` | Full learning map, prerequisite map, reading order, teach-back goals, scope and non-scope. |
| `os-kernel-distributed-systems-deep-dive/01_os_kernel_foundations.md` | OS kernel deep dive from syscall, CPU scheduling, memory, file I/O, network stack, locks, observability. |
| `os-kernel-distributed-systems-deep-dive/02_distributed_system_foundations.md` | Distributed-systems principles: failure, time, logs, partitioning, replication, quorum, consensus, consistency, recovery. |
| `os-kernel-distributed-systems-deep-dive/03_kafka_deep_dive.md` | Kafka internal structure connected to log, replicas, offsets, page cache, disk and network. |
| `os-kernel-distributed-systems-deep-dive/04_cassandra_deep_dive.md` | Cassandra internal structure connected to ring, LSM, quorum, compaction, repair, OS resources. |
| `os-kernel-distributed-systems-deep-dive/05_spark_deep_dive.md` | Spark internal structure connected to driver/executor, DAG, shuffle, lineage, memory, spill. |
| `os-kernel-distributed-systems-deep-dive/06_cross_system_comparison.md` | Kafka/Cassandra/Spark comparison and kernel-to-system bridges. |
| `os-kernel-distributed-systems-deep-dive/07_interview_reasoning_playbook.md` | Interview reasoning templates, short answers, deeper follow-up answer paths, bad/good answer comparison. |
| `os-kernel-distributed-systems-deep-dive/08_experiments_and_observability.md` | Local experiments and observability commands with Linux/macOS boundaries, PASS/FAIL signals. |
| `os-kernel-distributed-systems-deep-dive/09_glossary.md` | Korean-first glossary with English original, confusion pairs, first appearance, interview one-liner. |
| `os-kernel-distributed-systems-deep-dive/10_source_ledger.md` | Official docs, papers, primary/high-confidence sources, checked date, evidence tier. |
| `os-kernel-distributed-systems-deep-dive/audit/claim_review.md` | Internal claim cards, critic attacks, repairs, review-kernel result. |
| `../docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md` | This execution ledger. |

Non-write set:

- Existing modified files in `interviews/` and sibling directories are outside the write scope.
- Push is out of scope.
- Root or existing interview `README.md` updates are avoided because unrelated local edits already exist and the new corpus has its own entry page.

## Checklist Freeze

The following checklist is non-compensating. One failed item blocks WHOLE_COMPLETE.

| ID | Requirement | Status |
|---|---|---|
| C01 | Required document set `00` through `10` exists under the new corpus root. | PASS |
| C02 | `README.md` and `00` expose reading order, prerequisite map, teach-back goals, scope and non-scope. | PASS |
| C03 | OS kernel scope covers mode, interrupt/trap/syscall, process/thread/context switch/scheduling, virtual memory/page table/TLB/page fault, file/page cache/block I/O, network stack, locks, observability. | PASS |
| C04 | Distributed-systems scope covers partial failure, time/clocks/ordering, logs/state machines, partitioning, replication, quorum, consensus, consistency, CAP/PACELC, idempotency, backpressure, checkpoint/recovery. | PASS |
| C05 | Kafka deep dive covers log/topic/partition, producer/broker/consumer/group, offset/ordering/retention, replication/ISR/leader election, delivery semantics/transactions, throughput/batching/backpressure/failure modes, OS links. | PASS |
| C06 | Cassandra deep dive covers ring, partitioning, consistent hashing, RF/quorum/tunable consistency, write/read paths, bloom filter, compaction, hinted handoff/read repair/anti-entropy, LSM and OS links. | PASS |
| C07 | Spark deep dive covers driver/executor/cluster manager, RDD/DataFrame/DAG, narrow/wide dependency, shuffle, partitioning, lineage/checkpointing, memory/cache/spill/performance, streaming, OS links. | PASS |
| C08 | Cross-system comparison connects Kafka/Cassandra/Spark through log, replication, partitioning, recovery, backpressure, checkpoint, OS resources. | PASS |
| C09 | Interview playbook includes direct short answers, lower-layer follow-ups, reasoning templates, bad/good comparisons. | PASS |
| C10 | Experiments include purpose, prerequisite, commands, expected observation, PASS signal, FAIL signal, Linux/macOS distinction. | PASS |
| C11 | Glossary includes Korean-first meaning, English original, confusion pairs, first appearance, interview one-liner. | PASS |
| C12 | Source ledger includes official docs, papers, primary/high-confidence sources, checked date, URL, evidence tier. | PASS |
| C13 | Every major topic gives a teach-back target, but the visible document is prose-first and does not expose the internal teaching checklist as repeated labels. | PASS |
| C14 | Major topics close the underlying teaching spine: concrete question, history/context, first brick, worked trace, mechanism, where it breaks, how to verify, interview compression, misconceptions, active recall. These may appear as natural prose and purpose-specific headings, not as a fixed template. | PASS |
| C15 | Each major document includes at least two reality-based scenarios with the required decomposition. | PASS |
| C16 | Forbidden diagram syntax is absent. ASCII diagrams/tables/timelines are used for state movement. | PASS |
| C17 | Version-sensitive Kafka/Cassandra/Spark claims are checked against official docs or downgraded. | PASS |
| C18 | Dialectic claim cards receive targeted attack and close as ACCEPT_REPAIR, REBUT, DOWNGRADE, or BLOCKED. | PASS |
| C19 | Final review checks whether a Korean backend developer with some practical experience but weak OS/distributed-system foundations can read sequentially and explain lower-layer mechanisms without memorized keyword answers. | PASS |
| C20 | Whole corpus stays under 50MB. | PASS |
| C21 | Final verification commands pass or any failure is repaired. | PASS |
| C22 | Only our write set is staged and committed after final review. | READY: stage allowlist is corpus root plus this WORK file; unrelated dirty files remain unstaged. |
| C23 | Push is not performed. | PASS |

## Source Resolution Ledger

Checked date: 2026-05-31.

Primary/current official docs already consulted or queued for the source ledger:

| Domain | Source | Current use |
|---|---|---|
| Kafka | Apache Kafka 4.3 documentation and design docs | Direct evidence for log, partition, page cache, sendfile, ISR, leader election, delivery semantics. |
| Cassandra | Apache Cassandra docs, architecture, Dynamo, storage engine, guarantees, hints, compaction docs | Direct evidence for ring, LSM, RF, quorum, hints, compaction, CAP wording with caution. |
| Spark | Apache Spark 4.1.2 docs, cluster overview, RDD guide, tuning, monitoring | Direct evidence for driver/executor, RDDs, fault tolerance, storage levels, shuffle, memory tuning. |
| Linux | Linux kernel docs, man7 syscall/sendfile/epoll pages | Direct evidence for page tables, page faults, page cache, scheduler caveats, syscall and zero-copy terminology. |
| Distributed systems | Lamport clocks, Raft, Dynamo, Bigtable, LSM tree, Spark RDD paper, CAP/PACELC papers | Primary or high-confidence evidence for core theory. |

## Claim Cards

| Claim | Evidence / assumption | Attack | Closure |
|---|---|---|---|
| The corpus should be self-contained even where it overlaps with existing docs. | Latest user message explicitly says duplicate is acceptable if the document itself is complete. | Existing-doc links could become content substitutes. | ACCEPT_REPAIR: all required docs must carry their own definitions, traces, scenarios, and verification. |
| Kafka should be taught as a distributed append-only log, not only as a queue. | Kafka design docs connect Kafka to database-like logs and partitioned feeds. | Could overstate if queue use cases are erased. | ACCEPT_REPAIR: explain queue-like use as one consumer pattern, but make log/offset/retention the base model. |
| Cassandra quorum is tunable consistency, not automatic strong consistency. | Cassandra docs describe RF, CL, eventual consistency, LWT exceptions. | Quorum intersection depends on read/write CL and replica health; hints are best effort. | ACCEPT_REPAIR: include RF=3 CL examples, stale-read scenario, and downgrade hints to best-effort repair aid. |
| Spark lineage supports recovery by recomputing lost partitions. | Spark RDD docs describe lost partition recomputation from transformations. | Shuffle outputs, nondeterminism, external side effects, and long lineage can require checkpointing. | ACCEPT_REPAIR: include lineage vs checkpoint boundary and shuffle-file-loss scenario. |
| OS kernel knowledge explains distributed-system symptoms. | Kernel docs and system behavior connect CPU scheduling, memory pressure, page cache, disk, sockets to user-visible latency. | Connection can be asserted without trace. | ACCEPT_REPAIR: include cross-system traces for page cache, socket buffers, disk saturation, GC/memory pressure, spill. |
| The previous corpus already satisfied the original objective. | User-provided critique and local inspection show label-heavy sections that compress key mechanisms. | Earlier checklist says PASS, so the ledger could falsely authorize completion. | ACCEPT_REPAIR: downgrade current corpus to REWORK target, replace visible template labels with prose-first monograph sections, and re-run checklist after edits. |
| `mvcc.md` is a useful benchmark. | It contains history/context and implementation comparison depth. | It also contains repeated sections and list-heavy structure, so copying its shape would preserve a weakness. | DOWNGRADE: use it as a floor for depth and context, not as a formatting exemplar. |
| Gold-slice-first is a valid expansion strategy. | The user explicitly pointed to `write()` as a representative failure. | A slice that only improves wording but not OS-to-Kafka/Cassandra/Spark bridges would be too narrow. | ACCEPT_REPAIR: gold slice must include syscall mechanics, page cache/durability, and product bridges before expansion. |

## Multi-Agent / Review Ledger

Protocol Sentinel findings received:

- WHOLE_COMPLETE before full doc set + review + verification + commit is forbidden.
- Commit must stage only new corpus files and this WORK/audit material.
- Push is forbidden.
- Existing docs cannot replace corpus content.

Curriculum Critic findings received:

- Product-by-product isolation is a failure mode unless `03~06` repeatedly connect common principles.
- Page cache / fsync / durable write needs explicit trace.
- Partial failure, time/order, quorum, consensus, delivery semantics, LSM, Spark lineage need concrete small examples and failure boundaries.
- Experiments must not degrade into a command list.

Repairs accepted:

- Each major file must include local prerequisites, core trace, scenarios, misconceptions, verification, active recall.
- `06` must be a first-class synthesis, not a summary table only.
- `10` must distinguish direct evidence, strong inference, general principle, and version-check-needed.

## Progress Ledger

| Unit | State | Notes |
|---|---|---|
| Instruction stack | SECTION_REVIEWED | Applicable global, study, interviews, project fact docs reviewed. |
| Target path | SECTION_REVIEWED | New self-contained corpus root chosen. |
| Source browsing | SECTION_REVIEWED | Official Kafka/Cassandra/Spark/Linux docs and primary papers recorded in source ledger. |
| WORK ledger | READY_FOR_SCOPED_COMMIT | Scope, checklist, rework gates, verification, and staging boundary updated after current rewrite. |
| Corpus documents | REWRITTEN | Required `00` through `10`, README, and audit ledger rewritten to prose-first standard. |
| Final review | SECTION_REVIEWED | Dialectic claim review and reader-facing curriculum/protocol audits recorded in `audit/claim_review.md` and this WORK. |
| Commit | READY | Scoped stage list must include only corpus root and this WORK file; push forbidden. |
| 2026-05-31 rewrite request intake | SECTION_REVIEWED | User rejected the existing style as too shallow and requested full-corpus uplift. Previous PASS rows are no longer authoritative. |
| Gold-slice critic review | SECTION_REVIEWED | Curriculum critic accepted gold-slice-first only if it closes undefined boundary terms, syscall mechanics, page cache/durability, and product bridges. |
| Full-corpus curriculum audit | SECTION_REVIEWED | Critic accepted reader journey after repairs to `01` lock/concurrency and `02` consensus trace. |
| Protocol sentinel audit | SECTION_REVIEWED | Sentinel blockers repaired except commit itself: WORK open rows reconciled, DataFrame/DAG pass added, unsafe command/source-ledger rows closed, `interviews/README.md` remains outside staging scope. |

## Next Immediate Target

Stage exactly `os-kernel-distributed-systems-deep-dive/` and `../docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md`, verify cached diff, commit, and do not push.

## Checklist Re-Judgement

| ID | Requirement | Result |
|---|---|---|
| C01 | Required document set `00` through `10` exists under the new corpus root. | PASS |
| C02 | `README.md` and `00` expose reading order, prerequisite map, teach-back goals, scope and non-scope. | PASS |
| C03 | OS kernel scope covers required kernel topics. | PASS |
| C04 | Distributed-systems scope covers required foundation topics. | PASS |
| C05 | Kafka deep dive covers required Kafka topics and OS links. | PASS |
| C06 | Cassandra deep dive covers required Cassandra topics and OS links. | PASS |
| C07 | Spark deep dive covers required Spark topics and OS links. | PASS |
| C08 | Cross-system comparison connects systems through common mechanisms. | PASS |
| C09 | Interview playbook includes short answers, follow-ups, templates, bad/good comparisons. | PASS |
| C10 | Experiments include purpose, prerequisite, commands, expected observation, PASS, FAIL, Linux/macOS distinction where relevant. | PASS |
| C11 | Glossary includes Korean-first meaning, English original, confusion pairs, first appearance, interview one-liner. | PASS |
| C12 | Source ledger includes official docs, papers, checked date, URL, evidence tier. | PASS |
| C13 | Major topics start with teach-back blockquote and deep explanation. | PASS |
| C14 | Major topics follow question, intuition, small example, trace, mechanism, failure, verification, interview follow-up, misconception, active recall. | PASS |
| C15 | Each major document includes at least two reality-based scenarios where applicable. | PASS |
| C16 | Forbidden diagram syntax is absent; ASCII diagrams/tables/timelines are used. | PASS |
| C17 | Version-sensitive Kafka/Cassandra/Spark claims are checked or downgraded. | PASS |
| C18 | Dialectic claim cards receive targeted attack and closure. | PASS |
| C19 | Final review checks sequential beginner-to-expert reasoning. | PASS |
| C20 | Whole corpus stays under 50MB. | PASS |
| C21 | Final verification commands pass or failures are repaired. | PASS |
| C22 | Only our write set is staged and committed after final review. | PASS for scoped staging; commit command next |
| C23 | Push is not performed. | PASS |

## Verification Log

| Check | Result |
|---|---|
| Required files | PASS, no missing required files. |
| Corpus byte size | PASS, `du -sh` reports 164K, below 50MB. |
| Internal links | PASS, local Markdown link parser found `broken_links=0`. |
| Forbidden diagram syntax and old template labels | PASS, `rg` found no Mermaid syntax, no repeated `### 질문/직관/...` template labels, no `경계 통과`, and no unsafe broad `pkill -f` command. |
| Required scope terms | PASS, term scan found syscall/page cache/fsync/partial failure/quorum/consensus/Kafka/Cassandra/Spark/compaction/lineage/checkpoint/shuffle/backpressure coverage. |
| Experiment safety | PASS, CPU loop uses PID/trap cleanup; `nodetool compact` is local-only opt-in and not a normal verification command. |
| Source URL repair | PASS, 24 URLs in source ledger returned HTTP OK/redirect-success via `curl -L -I --max-time 20`. |
| Markdown diff hygiene | PASS, `git diff --check` reports no whitespace errors in scoped files. |
| Dirty worktree boundary | PASS, unrelated modified/untracked files exist but are outside stage allowlist and must remain unstaged. |

## Rework Gold-Slice Gate

The `write(fd, buf, len)` slice may expand to the rest of the corpus only if all rows pass.

| Gate | PASS signal | State |
|---|---|---|
| GS01 undefined-boundary repair | The text explains that the boundary is CPU privilege, process-visible address space, and kernel-owned data structures; it does not leave `경계 통과` as an unexplained metaphor. | PASS: `01_os_kernel_foundations.md` section 1 defines boundary as privilege, address access, and kernel-owned structures. |
| GS02 causal history repair | The text explains the need for sharing expensive machines, multiple programs/users, protection, scheduling, and device mediation before naming user/kernel separation. | PASS: section 1 uses shared hardware need -> scheduling/protection/mediation -> user/kernel mechanism. |
| GS03 concrete syscall descent | `write(3, "hello", 5)` descends through libc wrapper, syscall entry, fd table, open file description, permission/address checks, VFS/filesystem, page cache, and writeback. | PASS: section 1 trace and prose include these steps, plus partial-write return handling. |
| GS04 durability boundary | The text separates successful `write()` from stable storage and names `fsync()`/`fdatasync()` as stronger durability requests with filesystem/storage caveats. | PASS: section 1 separates write success, dirty page, fd sync, metadata/directory/storage caveats. |
| GS05 cross-system bridge | The text connects the same kernel mechanism to Kafka page cache/sendfile, Cassandra commit log/fsync policy, and Spark spill/shuffle files. | PASS_FOR_GOLD_SLICE: bridge exists; each product file must expand it deeper. |
| GS06 Korean prose quality | The section reads as a coherent Korean technical essay with purpose-specific headings, not as a filled checklist. | PASS: critic confirmation accepted prose-first standard after one wording repair. |

Gold-slice source anchors checked during rework:

| Claim | Source evidence | Repair decision |
|---|---|---|
| `write()` can partially write and returns the number of bytes accepted. | man7 `write(2)` return-value and notes sections; opened 2026-05-31. | Added partial-write paragraph and required retry-loop reasoning. |
| Bad user buffer is an `EFAULT` contract in `write(2)`. | man7 `write(2)` errors section. | Gold slice says `EFAULT`; signal wording is reserved for pipe/socket or interruption cases. |
| `fd` is a process-local descriptor that references an open file description. | man7 `open(2)` description; opened 2026-05-31. | Gold slice explains fd table -> open file description -> offset/flags. |
| Normal buffered reads/writes/mmaps go through page cache, with direct-I/O caveats. | Linux kernel `Page Cache` doc; opened 2026-05-31. | Gold slice says buffered regular-file I/O and names `O_DIRECT`/special path caveats. |
| `fsync()` is stronger than `write()` but not a universal crash-consistency spell. | man7 `write(2)` notes and `fsync(2)` description; opened 2026-05-31. | Gold slice separates file data, metadata, parent directory, filesystem/storage caveats. |

## Rework Scope Ledger

Files reviewed after the gold slice:

| File | Rework state |
|---|---|
| `README.md` | PASS |
| `00_index_and_learning_path.md` | PASS |
| `01_os_kernel_foundations.md` | PASS, including post-audit lock/concurrency expansion |
| `02_distributed_system_foundations.md` | PASS, including post-audit consensus/state-machine trace |
| `03_kafka_deep_dive.md` | PASS |
| `04_cassandra_deep_dive.md` | PASS |
| `05_spark_deep_dive.md` | PASS, including post-audit DataFrame/DAG plan bridge |
| `06_cross_system_comparison.md` | PASS |
| `07_interview_reasoning_playbook.md` | PASS |
| `08_experiments_and_observability.md` | PASS, unsafe broad kill repaired; heavy compaction constrained |
| `09_glossary.md` | PASS |
| `10_source_ledger.md` | PASS, claim-level source ledger and URL/version repairs complete |
| `audit/claim_review.md` | PASS |

## Final Reader-Journey Audit

Reader: Korean backend developer with practical service experience but weak OS/distributed-system foundations.

Verdict before commit: PASS.

The current reading path builds from a single-machine `write()` trace to partial failure/time/order, then maps the same state-movement reasoning onto Kafka record logs, Cassandra mutations, and Spark tasks. Critic blockers were repaired:

- `01` now explains not only lock contention but race, deadlock, lost wakeup, memory ordering, and JVM/OS wait-state interpretation.
- `02` now includes a Raft-like state-machine replication trace with leader term, log index, majority append, commit, and apply.
- `05` now includes DataFrame/DAG logical-plan to physical-plan to stage/task execution bridge.

Remaining open tranche disclosure: none inside the requested corpus rewrite. `interviews/README.md` and other dirty worktree files are explicitly outside this write set and must remain unstaged.
