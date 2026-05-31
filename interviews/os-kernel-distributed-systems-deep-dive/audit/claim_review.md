# Claim Review Ledger

Status: SECTION_REVIEWED

This file records dialectic and review-kernel evidence for the corpus. It is not a reader-facing tutorial. The reader-facing learning material is in `00` through `10`.

## Activated Review Roles

| Role | Responsibility | Result |
|---|---|---|
| Source Evidence Researcher | Primary/official source selection and version-sensitive claim warnings | Cassandra fixed-version source, Kafka exactly-once boundary, Spark latest docs, Linux version-sensitive boundaries integrated. |
| Curriculum Critic | Attack shallow-summary failure modes and missing traces | Repair accepted: self-contained docs, per-system traces, scenarios, OS/distributed links. |
| Protocol Sentinel | Guard whole-request completion, commit/push boundary, no premature closure | Repair accepted: no intermediate commit, only new corpus and WORK can be staged, push forbidden. |

## Claim Cards

| ID | Claim | Attack focus | Closure |
|---|---|---|---|
| CC01 | The new corpus may duplicate existing docs. | Latest user message says duplicate is acceptable, but duplication must not become shallow repetition. | ACCEPT_REPAIR: new corpus is self-contained; existing docs are not linked as substitutes. |
| CC02 | OS kernel knowledge explains Kafka/Cassandra/Spark symptoms. | Could become a slogan without concrete trace. | ACCEPT_REPAIR: `01`, `03`, `04`, `05`, `06`, `08` include page cache, disk, socket, scheduling, GC/spill/compaction traces. |
| CC03 | Kafka should be taught as append-only log. | Risk of erasing queue use cases or overstating topic-wide order. | ACCEPT_REPAIR: `03` explains queue-like consumer group pattern but grounds model in partition log, offset, retention, per-partition order. |
| CC04 | Kafka exactly-once can be described in the corpus. | Dangerous overclaim if external systems are included. | ACCEPT_REPAIR: `03` and `07` restrict exactly-once processing to supported Kafka-internal transaction/idempotence/read isolation conditions and call out external side effects. |
| CC05 | Kafka durability follows from replication. | Overclaim if `acks`, ISR, `min.insync.replicas`, unclean election are omitted. | ACCEPT_REPAIR: `03` includes leader/follower, ISR, high watermark, `acks`, `min.insync.replicas`, unclean election tradeoff. |
| CC06 | Cassandra quorum improves consistency. | Could be misread as automatic linearizability. | ACCEPT_REPAIR: `02`, `04`, `07` distinguish quorum intersection from linearizable read, include timestamp/failed write/repair cautions. |
| CC07 | Cassandra hints repair missed writes. | Hints are best-effort and not anti-entropy repair. | ACCEPT_REPAIR: `04`, `09`, `10` explicitly say hints do not replace anti-entropy repair. |
| CC08 | LSM tree is write-optimized. | Half-truth if read amplification and compaction are omitted. | ACCEPT_REPAIR: `04` includes write path, read path, bloom filter, SSTable count, tombstone, compaction pressure, OS disk impact. |
| CC09 | Spark lineage provides fault tolerance. | Incomplete if shuffle output loss, nondeterminism, checkpoint boundaries are omitted. | ACCEPT_REPAIR: `05` includes lineage, shuffle file loss, nondeterministic side effects, checkpoint reliable storage. |
| CC10 | Spark is in-memory. | Dangerous simplification if shuffle/spill/checkpoint/input/output disk use is omitted. | ACCEPT_REPAIR: `05`, `06`, `08`, `09` say in-memory is an optimization direction, not disk-free execution. |
| CC11 | CAP/PACELC are useful teaching frames. | Risk of label memorization. | ACCEPT_REPAIR: `02` and `06` teach partition-time and normal-time tradeoff through read/write traces. |
| CC12 | Experiments can support the corpus. | Risk of becoming a command list or implying production-safe operations. | ACCEPT_REPAIR: `08` uses purpose/prerequisite/commands/expected/PASS/FAIL, and warns about production-side effects like `nodetool compact`. |
| CC13 | Official docs and papers are enough source coverage. | Version-sensitive docs and defaults can drift. | ACCEPT_REPAIR: `10` distinguishes Direct, Strong inference, General principle, Version-check-needed and records checked date. |
| CC14 | Each document can be read independently. | Cross-links might hide missing definitions. | ACCEPT_REPAIR: `01` through `05` include local prerequisites, traces, failure modes, verification, scenarios, active recall. |
| CC15 | The corpus can be closed by file count. | File count is a proxy, not quality evidence. | REBUT: final closure requires coverage checks, forbidden diagram syntax check, link/source review, final audit, size check, commit, no push. |

## Repairs That Changed The Corpus

| Critique | Repair in documents |
|---|---|
| Kafka might be reduced to "message queue." | `03` begins with distributed append-only log model, explains consumer groups as one reading pattern. |
| Page cache/durable write could be vague. | `01` and `08` include `write()` vs `fsync()` traces; `03` and `04` connect this to Kafka/Cassandra. |
| Quorum could be overstated. | `02`, `04`, `07`, `09`, `10` all distinguish quorum intersection from automatic linearizability. |
| Spark could become a feature list. | `05` follows driver -> DAG -> stage -> shuffle -> spill -> lineage/checkpoint state movement. |
| Experiments could be unsafe command dumps. | `08` separates PASS/FAIL and marks side-effect risks, especially cache drop and Cassandra compaction. |
| Existing docs could replace new content. | `README`, `00`, WORK ledger state self-contained scope; no root README update or existing-doc substitution. |

## Final Review Kernel

Review question: Can a Korean backend developer who starts with little OS/distributed-systems knowledge read this corpus sequentially and answer lower-layer follow-up questions about Kafka, Cassandra, and Spark?

Findings after automated coverage and author-side review:

| Finding | Severity | Status |
|---|---|---|
| Required files, forbidden diagram syntax check, internal links, source ledger, size. | material | PASS by verification log. |
| Every big deep-dive or synthesis doc has at least two scenarios. | material | PASS for `01` through `08`; `00`, `09`, `10` are index/glossary/source roles. |
| Major scope terms appear in corpus. | material | PASS by keyword coverage check. |
| Kafka/Cassandra/Spark are connected back to OS and distributed-systems principles. | material | PASS by review of `03`, `04`, `05`, `06`, `07`, `08`. |
| Commit only scoped files after final checks. | material | PASS pending final git stage/commit command. |

## Verification Log

Commands run from `interviews/`:

```bash
find os-kernel-distributed-systems-deep-dive -type f -maxdepth 3 | sort
du -sh os-kernel-distributed-systems-deep-dive
wc -c os-kernel-distributed-systems-deep-dive/*.md os-kernel-distributed-systems-deep-dive/audit/*.md
rg <forbidden diagram syntax keyword check> os-kernel-distributed-systems-deep-dive ../docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md
python3 <coverage-check: required files, quote/scenario/PASS/FAIL/active recall counts, internal links, total bytes>
python3 <scope-term-check>
rg <progress-filler phrase check> os-kernel-distributed-systems-deep-dive
curl <Cassandra storage-engine fixed/latest URL check>
```

Observed results:

| Check | Result |
|---|---|
| Required `00` through `10` files | PASS, no missing files. |
| Corpus size | PASS, about 238KB, far below 50MB. |
| Internal Markdown links | PASS, no missing local targets. |
| Forbidden diagram syntax | PASS, no occurrences. |
| Scope terms | PASS, no missing required terms in corpus text. |
| Scenario count | PASS for `01`-`08`; each has at least two scenarios or reality scenarios. |
| Experiment PASS/FAIL signals | PASS, `08` has 12 PASS/FAIL blocks plus active recall. |
| Korean progress filler phrases | PASS after replacing the only material phrase hit in `02`. |
| Cassandra storage engine URL | PASS, fixed 5.0.8 `storage-engine.html` returned HTTP 200. |

## Final Reader-Facing Quality Review

Verdict before git commit: PASS.

Reasoning:

- The corpus starts from OS boundary concepts and repeatedly turns them into byte/page/packet/log/task traces.
- Distributed-system concepts are not treated as CAP/quorum vocabulary only; the docs show timeout ambiguity, ordering, log/state machine, quorum intersection, retry, backpressure, checkpoint and recovery traces.
- Kafka, Cassandra, and Spark each include their own local prerequisites, state movement diagrams, failure modes, verification methods, common misconceptions, and interview follow-up paths.
- Cross-system comparison and interview playbook force the reader to reuse lower-layer mechanisms rather than memorize isolated product facts.
- Experiments are framed as observation exercises with PASS/FAIL signals, not production prescriptions.

Residual risk:

- This corpus is not an operations manual for every version and deployment topology. The source ledger marks version-sensitive areas that must be rechecked against the target production version.
