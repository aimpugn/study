# WORK 2026-05-31 - OS kernel and distributed systems learning corpus

## Current State

Status: PARTIAL_PROGRESS

Whole-request objective: `OS 커널과 분산 시스템 원리, Kafka/Cassandra/Spark 내부 구조를 초보자에서 전문가 수준까지 학습할 수 있는 self-contained Markdown 학습 코퍼스를 repo 안에 작성하고, 최종 review와 검증 후 commit까지 수행한다.`

Whole-request completion verdict: not yet WHOLE_COMPLETE.

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

Latest user clarification: `아뇨 중복이어도 그냥 그 문서 자체로 완벽하게 정리해 주세요.`

Decision: existing docs may be consulted, but this corpus must be self-contained. Duplicate coverage is acceptable and sometimes required. Existing docs are not allowed to replace content inside the new corpus.

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
| C01 | Required document set `00` through `10` exists under the new corpus root. | open |
| C02 | `README.md` and `00` expose reading order, prerequisite map, teach-back goals, scope and non-scope. | open |
| C03 | OS kernel scope covers mode, interrupt/trap/syscall, process/thread/context switch/scheduling, virtual memory/page table/TLB/page fault, file/page cache/block I/O, network stack, locks, observability. | open |
| C04 | Distributed-systems scope covers partial failure, time/clocks/ordering, logs/state machines, partitioning, replication, quorum, consensus, consistency, CAP/PACELC, idempotency, backpressure, checkpoint/recovery. | open |
| C05 | Kafka deep dive covers log/topic/partition, producer/broker/consumer/group, offset/ordering/retention, replication/ISR/leader election, delivery semantics/transactions, throughput/batching/backpressure/failure modes, OS links. | open |
| C06 | Cassandra deep dive covers ring, partitioning, consistent hashing, RF/quorum/tunable consistency, write/read paths, bloom filter, compaction, hinted handoff/read repair/anti-entropy, LSM and OS links. | open |
| C07 | Spark deep dive covers driver/executor/cluster manager, RDD/DataFrame/DAG, narrow/wide dependency, shuffle, partitioning, lineage/checkpointing, memory/cache/spill/performance, streaming, OS links. | open |
| C08 | Cross-system comparison connects Kafka/Cassandra/Spark through log, replication, partitioning, recovery, backpressure, checkpoint, OS resources. | open |
| C09 | Interview playbook includes direct short answers, lower-layer follow-ups, reasoning templates, bad/good comparisons. | open |
| C10 | Experiments include purpose, prerequisite, commands, expected observation, PASS signal, FAIL signal, Linux/macOS distinction. | open |
| C11 | Glossary includes Korean-first meaning, English original, confusion pairs, first appearance, interview one-liner. | open |
| C12 | Source ledger includes official docs, papers, primary/high-confidence sources, checked date, URL, evidence tier. | open |
| C13 | Every major topic starts with teach-back blockquote and then gives deep explanation. | open |
| C14 | Major topics follow question -> intuition -> small example -> trace -> mechanism -> failure -> verification -> interview follow-up -> misconception -> active recall. | open |
| C15 | Each major document includes at least two reality-based scenarios with the required decomposition. | open |
| C16 | Forbidden diagram syntax is absent. ASCII diagrams/tables/timelines are used for state movement. | open |
| C17 | Version-sensitive Kafka/Cassandra/Spark claims are checked against official docs or downgraded. | open |
| C18 | Dialectic claim cards receive targeted attack and close as ACCEPT_REPAIR, REBUT, DOWNGRADE, or BLOCKED. | open |
| C19 | Final review checks whether a Korean backend developer can read sequentially from beginner to expert reasoning. | open |
| C20 | Whole corpus stays under 50MB. | open |
| C21 | Final verification commands pass or any failure is repaired. | open |
| C22 | Only our write set is staged and committed after final review. | open |
| C23 | Push is not performed. | open |

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
| WORK ledger | DOCUMENT_DRAFTED | This file freezes scope and checklist. |
| Corpus documents | DOCUMENT_DRAFTED | Required `00` through `10`, README, and audit ledger written. |
| Final review | SECTION_REVIEWED | Dialectic claim review and final reader-facing review recorded in `audit/claim_review.md`. |
| Commit | SECTION_REVIEWED | Scoped stage list checked; final commit command is the next action. |

## Next Immediate Target

Stage only the new corpus directory and this WORK file, commit after confirming no unrelated dirty files are included, and do not push.

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
| Corpus byte size | PASS, about 238KB. |
| Internal links | PASS, no broken local Markdown links. |
| Forbidden diagram syntax ban | PASS, no occurrences. |
| Required scope terms | PASS, no missing required keyword coverage. |
| Scenario/PASS/FAIL/active recall coverage | PASS after adding active recall to `08`. |
| Source URL repair | PASS, Cassandra 5.0.8 storage-engine URL checked and source ledger updated. |
