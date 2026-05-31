# WORK 2026-05-31 - OS kernel and distributed systems learning corpus

## 0. Meta

- 작업 제목: OS kernel / distributed systems / Kafka / Cassandra / Spark deep-study corpus repair
- WORK 파일 경로: `docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs | explain | audit`
- 작업 깊이: `full`
- 시작 일시: 2026-05-31
- 현재 상태: `COMPLETE`
- 완료 게이트: `WHOLE_COMPLETE` after scoped verification and commit; push remains forbidden
- finish: `test+commit`, no push

## 1. Request Normalization

Goal: `/Users/rody/VscodeProjects/study/interviews/os-kernel-distributed-systems-deep-dive/` 코퍼스를 OS 커널, 분산 시스템, Kafka, Cassandra, Spark의 lower-layer 메커니즘을 순서대로 공부할 수 있는 self-contained 학습 자산으로 재작성/보강한다.

Protected intent:

- 독자는 한국어로 공부하는 백엔드 개발자다.
- 거의 모르는 상태에서 순서대로 읽어도 OS kernel -> distributed systems -> Kafka/Cassandra/Spark 내부 구조를 lower-layer부터 설명할 수 있어야 한다.
- 얕은 정의 모음, 면접 키워드 암기장, 제품 기능 나열, 공식 문서 요약본, 카드형 Q/A 템플릿 채우기는 실패다.
- OS 파트는 "Kafka 이해를 위한 낮은 하한선"이 아니라, CPU/메모리/디스크/네트워크/동시성/컨테이너/관측을 실제 backend request path로 설명하는 기반이어야 한다.

Benchmark floor:

- `/Users/rody/VscodeProjects/study/database/mvcc.md`는 역사/맥락/구현 비교/깊이의 하한선으로 사용한다.
- 다만 반복과 목록식 약점은 복제하지 않는다. 이번 corpus는 prose-first, trace-first, product-bridge-first로 더 강해야 한다.

Explicit constraints:

- 파일을 실제 수정한다.
- 최종 검수와 검증 후 commit한다.
- push하지 않는다.
- unrelated dirty worktree는 수정/stage하지 않는다.
- Mermaid 금지, ASCII/table/timeline trace 사용.
- source/evidence boundary와 version-sensitive claim boundary를 남긴다.
- context compaction이나 재개가 있어도 WORK와 git status로 남은 항목을 복원해 계속한다.

## 2. Project Overlay

| Layer | Path | Applied decision |
|---|---|---|
| Runtime/global instructions | visible AGENTS stack in current thread | substantial repo-changing work, checklist freeze, claim review, final verification, commit after pass |
| Study repo contract | `/Users/rody/VscodeProjects/study/AGENTS.md` via user-provided overlay | deep study monograph, self-reconstructable Korean docs, final review + verification + commit |
| Interviews project facts | `PROJECT_INTENT.md`, `USECASE.md` | interview assets must support short direct answer plus lower-layer follow-up reasoning |
| Skills | `$multi-agent`, `$dialectic-kernel`, `$review-kernel`, `$study-explanation`, `$humanize-korean` | role separation, claim-card attacks, prose-first deep-study, Korean technical prose |

## 3. Scope Freeze

Write set:

- `interviews/os-kernel-distributed-systems-deep-dive/**`
- `docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md`

Non-write set:

- Existing unrelated dirty files in `study/` and `interviews/` remain outside scope.
- Push is out of scope.

## 4. Required Document Set

| File | Role | Current state |
|---|---|---|
| `README.md` | Entry page and reading order | Updated with OS detail sequence |
| `00_index_and_learning_path.md` | Full learning map | Updated with OS detail map |
| `01_os_kernel_foundations.md` | OS hub and gold slice | Updated as hub, not final OS boundary |
| `01a_process_scheduling.md` | process/thread/scheduler/lifecycle | Added |
| `01b_memory_and_address_space.md` | virtual memory, page fault, mmap, OOM | Added |
| `01c_filesystem_page_cache_block_io.md` | fd/inode/VFS/page cache/block I/O/fsync | Added |
| `01d_network_stack_and_io_multiplexing.md` | NIC/DMA/NAPI/TCP/socket/epoll/sync-async | Added |
| `01e_concurrency_isolation_observability.md` | lock/futex/cgroup/namespace/proc/perf/eBPF | Added |
| `02_distributed_system_foundations.md` | failure/time/order/replication/recovery | Reviewed; retained |
| `03_kafka_deep_dive.md` | Kafka internals | Reconnected to new OS file/network/scheduler traces |
| `04_cassandra_deep_dive.md` | Cassandra internals | Reconnected to commitlog/compaction/network/cgroup traces |
| `05_spark_deep_dive.md` | Spark internals | Reconnected to shuffle fetch, memory, cgroup OOM traces |
| `06_cross_system_comparison.md` | Cross-system comparison | Added OS detail comparison map and timeout scenario |
| `07_interview_reasoning_playbook.md` | Short answers and tail paths | Added multiplexing, packet path, low-CPU/high-latency answers |
| `08_experiments_and_observability.md` | Experiments | Added epoll, mmap/page fault, cgroup, perf, packet-path boundary experiments |
| `09_glossary.md` | Glossary | Added OS detail terms and trace/experiment references |
| `10_source_ledger.md` | Source/evidence ledger | Added official rows for missing OS claims |
| `audit/claim_review.md` | Claim-card audit | Updated with REWORK finding and accepted repairs |

## 5. Checklist Freeze And Re-Judgement

| ID | Requirement | Result |
|---|---|---|
| C01 | Required document set exists, including new OS detail files. | PASS |
| C02 | README and `00` link the OS detail reading order. | PASS |
| C03 | OS scope covers syscall, process/thread/lifecycle, scheduler, VM/TLB/page fault, mmap, filesystem/page cache/block I/O, network stack, blocking/non-blocking, sync/async, multiplexing, futex/concurrency, cgroups/namespaces, observability. | PASS |
| C04 | Network path includes NIC DMA ring, interrupt/NAPI/softirq, TCP/IP, socket queues, epoll/readiness, app parsing, response send, qdisc/NIC transmit. | PASS |
| C05 | Kafka/Cassandra/Spark reconnect to the new OS traces rather than only naming page cache/disk/network. | PASS |
| C06 | `06` compares log, partitioning, replication, consistency, recovery, backpressure, checkpoint, storage, network, failure handling. | PASS |
| C07 | `07` supports 30-second answers and lower-layer follow-up paths. | PASS |
| C08 | `08` experiments include purpose, prerequisite, command, expected observation, PASS/FAIL, safety/overclaim boundary. | PASS |
| C09 | `09` includes Korean-first meaning, English original, confusion pair, first appearance, interview one-liner, and related trace/experiment for newly added OS terms. | PASS |
| C10 | `10` includes official/primary sources, checked date, evidence tier, version-sensitive boundary for new OS/product claims. | PASS |
| C11 | Prose-first style remains; no old repeated card template dominates. | PASS |
| C12 | Mermaid is absent. | PASS |
| C13 | Local Markdown links are valid. | PASS |
| C14 | Scoped `git diff --check` and cached `git diff --cached --check` pass. | PASS |
| C15 | Source URLs are reachable or failures are explicitly bounded. | PASS |
| C16 | Reviewer lifecycle has no unresolved blocking status before commit. | PASS |
| C17 | Only write set is staged for the commit. | PASS |
| C18 | Push is not performed. | PASS_BY_FINAL_BOUNDARY |

## 6. Claim / Reasoning Ledger

| Claim | Attack | Closure | Repair |
|---|---|---|---|
| Previous `READY_FOR_SCOPED_COMMIT` was sufficient. | Critic found OS scope over-closed and missing network/multiplexing/container/source/experiment coverage. | ACCEPT_REPAIR | WORK status reset to `IN_PROGRESS`; added OS detail modules and synchronized product docs, experiments, glossary, source ledger, audit. |
| One OS overview file can carry the kernel foundation. | User explicitly said the line was far too low; critic found many topics as omissions. | ACCEPT_REPAIR | `01` is now hub/gold slice; `01a`-`01e` carry deep OS detail. |
| Network stack is sufficiently explained by socket buffer sketches. | Required path includes DMA/NAPI/softirq/TCP/epoll/request parsing/send/qdisc. | ACCEPT_REPAIR | `01d` added full inbound/outbound trace and readiness/completion explanation. |
| Blocking/non-blocking and sync/async can be brief glossary entries. | User named them as foundational; confusion changes server architecture reasoning. | ACCEPT_REPAIR | `01d`, `07`, `08`, `09`, `10` now cover the distinction and evidence boundary. |
| Product files are enough if they mention OS terms. | Product docs must use OS mechanisms to explain why the systems are shaped that way. | ACCEPT_REPAIR | `03`-`06` now point to OS traces for append/fetch, compaction, shuffle, timeout, memory/cgroup. |
| Source ledger proves explanation quality. | Sentinel flagged source ledger as necessary but not sufficient. | REBUT with boundary | Source ledger is treated as evidence boundary only; mechanism closure remains in prose and trace sections. |

## 7. Multi-Agent Reviewer Lifecycle

| Reviewer | Role | Assigned artifact | Observed status | Result ref | Closure impact |
|---|---|---|---|---|---|
| Darwin | Coverage Critic | current corpus before OS expansion | completed | subagent notification `019e7e75-d3cf-77c0-9520-ded476e80cc6` | REWORK accepted; drove OS expansion |
| Plato | Protocol Sentinel | closure checklist and false-complete risks | completed | subagent notification `019e7e75-e574-7482-b7ef-af551380dd46` | checklist adopted for final verification |
| Main orchestrator separated review | Builder/Integrator | final changed corpus | completed | final verification section | passed after scoped checks and sentinel repair |

Positive completion is allowed only as a scoped commit of the staged write set. Push remains out of scope.

## 8. Evidence Ledger

Checked date: 2026-05-31.

| Evidence | Type | Used for | Boundary |
|---|---|---|---|
| Linux man-pages `write(2)`, `open(2)`, `fsync(2)` | Direct evidence | syscall/fd/durability boundary | POSIX/Linux API; filesystem/device caveats remain |
| Linux man-pages `fork(2)`, `execve(2)`, `wait(2)`, `signal(7)`, `clone(2)` | Direct evidence | process lifecycle | Linux/POSIX boundary differs by call |
| Linux man-pages `mmap(2)` and kernel memory docs | Direct evidence + strong inference | VM/mmap/page fault explanation | mapping type and kernel version sensitive |
| Linux man-pages `epoll(7)`, `tcp(7)`, `socket(2)` | Direct evidence | readiness, TCP/socket semantics | fd type and kernel config sensitive |
| Linux kernel NAPI docs | Direct evidence | interrupt/polling network path | driver/hardware/kernel sensitive |
| Linux man-pages `futex(2)`, `cgroups(7)`, `namespaces(7)`, `proc(5)` | Direct evidence | futex/container/observability | cgroup v1/v2 and runtime sensitive |
| Linux kernel eBPF userspace API docs | Direct evidence | eBPF observability boundary | verifier/permission/kernel version sensitive |
| Apache Kafka/Cassandra/Spark official docs and primary papers | Direct evidence | product internals | version-sensitive claims marked in `10_source_ledger.md` |

## 9. Verification Plan

Run before commit:

```bash
git status --short
git diff --name-only
git diff --check -- os-kernel-distributed-systems-deep-dive ../docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md
rg -n '```mermaid|^graph (TD|TB|BT|LR|RL)|^flowchart|^sequenceDiagram|^classDiagram|^stateDiagram' os-kernel-distributed-systems-deep-dive
rg -n '^### (질문|직관|작은 예시|상태 이동 trace|실패 모드|면접식 되묻기|오해|Active recall|active recall)' os-kernel-distributed-systems-deep-dive
rg -n '경계 통과입니다|게이트를 닫|실패 모드' os-kernel-distributed-systems-deep-dive
```

Also run an unfinished-marker scan over the corpus prose, local Markdown link verification, and source URL checks. If no repo checker exists, use a bounded script/check and record it.

## 10. Final Verification Results

Checked date: 2026-05-31.

| Check | Result |
|---|---|
| `git diff --check -- os-kernel-distributed-systems-deep-dive ../docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md` | PASS |
| `git diff --cached --check` after scoped staging | PASS |
| Mermaid / graph syntax scan | PASS, no matches |
| Old card-heading scan for `### 질문`, `### 직관`, `### 작은 예시`, `### 상태 이동 trace`, interview-drill headings | PASS, no matches |
| Forbidden wording scan for `경계 통과입니다`, `게이트를 닫`, `실패 모드` | PASS, no matches |
| Unfinished-marker scan over the corpus and WORK | PASS, no actionable markers |
| Local Markdown link verification | PASS, `19` Markdown files checked |
| Source URL reachability check from `10_source_ledger.md` | PASS, all extracted URLs returned HTTP 200 |
| Scoped cached diff inspection | PASS, staged files are limited to `interviews/os-kernel-distributed-systems-deep-dive/**` and `docs/works/WORK_20260531_OS_KERNEL_DISTRIBUTED_SYSTEMS_CORPUS.md` |
| Coverage critic final pass | ACCEPT_REPAIR for content; one minor typo in `00_index_and_learning_path.md` was repaired before commit |
| Protocol sentinel pass | BLOCKED before repair because WORK still had stale pending states and cached scope proof was missing; required repairs were applied before commit |

Repo-wide `git diff --check` was not claimed because unrelated dirty worktree files are outside this task. The proof used for this task is scoped diff-check plus cached diff-check.

## 11. Current Progress Ledger

| Unit | State |
|---|---|
| Instruction stack and skill load | PASS |
| Initial corpus inspection | PASS |
| Critic/sentinel REWORK intake | PASS |
| OS detail docs `01a`-`01e` | WRITTEN |
| README/00/01 hub links | UPDATED |
| Product bridge docs `03`-`06` | UPDATED |
| Interview/experiments/glossary/source/audit | UPDATED |
| Final verification | PASS |
| Scoped staging | PASS |
| Commit | PASS |
| Push | FORBIDDEN |

## 12. Remaining Open Items

Remaining executable count before final response: `0`.

No executable item remains in this workstream. The final response must report the commit hash and no-push boundary.

WHOLE_COMPLETE is allowed for the corpus repair scope only. Unrelated dirty worktree entries remain outside this WORK.
