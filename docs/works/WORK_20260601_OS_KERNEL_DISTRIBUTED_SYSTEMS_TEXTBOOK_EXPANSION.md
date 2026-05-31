# WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TEXTBOOK_EXPANSION

## 0. Meta

- 작업 제목: OS Kernel, Distributed Systems, Kafka, Cassandra, Spark deep-study corpus textbook expansion
- WORK 파일 경로: `docs/works/WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TEXTBOOK_EXPANSION.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `research | analysis | design | explain | audit | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청: 붙여넣어진 requestnizer 실행 요청. 사용자는 prompt 재작성이나 plan-only가 아니라 실제 실행을 요청했다.
- 대상 경로 / 자산: `interviews/os-kernel-distributed-systems-deep-dive/`
- 실행자: Codex
- 시작 일시: 2026-06-01
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 현재 9천자 안팎인 OS kernel / distributed systems / Kafka / Cassandra / Spark 학습 문서 묶음을 OS 교과서급 주제 폭과 제품 내부 동작 연결을 갖춘 한국어 deep-study corpus로 확장한다.
- refs: OSTEP, CSAPP, Operating Systems: Principles and Practice, Operating System Concepts topic families, Linux man-pages, Linux kernel docs, Apache Kafka/Cassandra/Spark 공식 문서, 분산 시스템 1차 논문.
- scope: `interviews/os-kernel-distributed-systems-deep-dive/` 아래 본문 문서, 허브, glossary, source ledger, audit 문서와 이 WORK 문서.
- mode: execute / rewrite / expand / verify / commit.
- run_mode: `normal`
- finish: final audit와 verification이 모두 통과하면 commit. push는 하지 않는다.
- must_keep: unrelated dirty worktree untouched and unstaged. 저작권 있는 교과서 본문을 베끼거나 chapter-by-chapter summary로 만들지 않는다.
- extra_checks: `wc -m`, coverage matrix, forbidden phrase scan, Markdown/source checks, source reachability, final critic/sentinel review, staged scope check.

### 1.1 Explicit Deliverables

- 모든 substantive Markdown 문서의 UTF-8 문자 수를 20,000자 이상으로 만든다. 20,000자는 통과선이 아니라 하한이다.
- OS 교과서 주제군을 빠짐없이 다룬다. 커널/사용자 모드, trap/interrupt/syscall/context switch, process/thread/scheduling, synchronization/futex/atomics, virtual memory/TLB/page fault/COW/mmap/swap/OOM/cgroup, file descriptor/VFS/page cache/writeback/fsync/crash consistency/block layer, NIC/DMA/NAPI/softirq/TCP/socket buffer/epoll/kqueue/io_uring/qdisc/backpressure, IPC, virtualization/container/security, observability를 포함한다.
- Distributed systems는 partial failure, timeout, clocks/order, message delay/reorder/duplication, log replication, leader/follower, membership, failure detector, quorum, consensus, recovery/replay/snapshot/checkpoint, idempotency/retry, overload/backpressure, failure injection을 포함한다.
- Kafka, Cassandra, Spark 문서는 제품 기능 요약이 아니라 OS와 분산 시스템 bridge로 다시 쓴다.
- 한국어는 `study-explanation`과 `humanize-korean` 원칙을 적용해 사람이 읽기 쉬운 prose-first 장문 학습 문서로 쓴다.
- `10_source_ledger.md`는 공식 문서, man page, kernel docs, 논문 중심으로 evidence tier, checked date, version-sensitive boundary, reachability를 갱신한다.
- Multi-agent 형식은 장식이 아니라 claim card, critic attack, repair/rebuttal/downgrade/blocker, sentinel veto로 남긴다.
- completion은 WHOLE_COMPLETE일 때만 선언한다. slice 완료, sample 완료, outline 완료는 완료가 아니다.

### 1.2 Non-Goals

- 교과서 원문을 베끼지 않는다.
- 문서 수를 늘리는 것 자체를 목표로 삼지 않는다.
- 현재 저장소의 unrelated dirty 파일을 정리하거나 stage하지 않는다.
- push하지 않는다.

## 2. Root-First Framing

- 근본 문제: 현재 corpus는 방향은 맞지만 본문 문서들이 짧고, OS 교과서 주제 폭과 제품 내부 동작 연결을 한 문서 안에서 재구성할 만큼 깊지 않다.
- 왜 중요한가: 사용자는 이 문서를 실제로 읽고 면접 reasoning과 실무 진단 사고를 복원하려 한다. 얕은 요약은 키워드를 늘릴 뿐, 장애 증상에서 OS queue, page, packet, log, replica, task까지 내려가는 능력을 만들지 못한다.
- 작업 목표: 각 문서를 독립적으로 읽어도 한 메커니즘을 설명할 수 있고, 전체 corpus로 읽으면 OS kernel -> distributed systems -> Kafka/Cassandra/Spark -> interview/experiment로 연결되는 학습 경로가 보이게 한다.
- 성공 정의: 모든 substantive docs가 20,000자 이상이고, coverage matrix와 source ledger가 본문 주장에 역추적되며, forbidden phrase/style/Markdown/source/staged scope checks를 통과하고 commit된다.
- PARTIAL 조건: 일부 문서가 확장되었지만 한 문서라도 20,000자 미만이거나 coverage/source/style/final review가 열려 있다.
- BLOCKED 조건: 파일 접근, 검증 도구, source reachability, 또는 외부 승인 경계 때문에 더 진행할 수 없는 경우. 단순히 분량이 많다는 이유는 BLOCKED가 아니다.

## 3. Reader & Internalization Contract

- 주 독자: OS, 분산 시스템, Kafka/Cassandra/Spark를 키워드 수준으로만 알고 있는 백엔드 개발자 또는 면접 준비자.
- teach-back 목표: 독자가 `timeout`, `lag`, `compaction`, `shuffle spill`, `OOM kill`, `p99 latency` 같은 증상을 보면 애플리케이션 로그에서 OS queue/buffer/page/thread, 분산 log/replica/quorum, 제품 내부 경로까지 순서대로 말할 수 있어야 한다.
- learner gap 진단: 용어 암기는 가능하지만, 가상 주소가 물리 page로 바뀌고, fd가 open file description과 page cache를 거치고, packet이 NIC/DMA/softirq/socket buffer/epoll을 지나고, quorum과 consensus가 왜 다른지까지 한 흐름으로 연결하는 데 빈칸이 있다.
- 설명 원칙: 쉬운 한국어로 먼저 고정하고 공식 용어를 괄호로 병기한다. 각 major section은 존재 이유, 첫 concrete state, 이동하는 객체, trace, 오해 교정, 관측/검증, 제품 bridge, 문서를 덮고 다시 그려 볼 질문을 자연스럽게 포함한다.
- exemplar: `computer_architecture/ostep/xv6-riscv/kernel/entry.S`의 첫 상태 고정, `algorithms/dynamic_programming.md`의 작은 예에서 모델링으로 올라가는 방식, `git/git_rebase.md`의 실제 작업 흐름 기반 설명을 참고한다. 형식 복제나 line-by-line 과잉 설명은 피한다.

## 4. Project Overlay

- 적용한 로컬 AGENTS 경로:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/interviews/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - 자연어 요청을 실행 계약으로 정규화한다.
  - full depth 작업은 WORK ledger, 범위 조사, 근거, 반박, 검증, 최종 감사를 닫는다.
  - 현재 저장소 평균 품질이 아니라 exemplar와 외부 1차 자료 기준으로 끌어올린다.
  - repo 변경 작업은 최종 감사, 필요한 검증, commit으로 닫는다.
- 전역 규칙과의 충돌 여부: 없음.

## 5. Scope Expansion & Impact Sync

### 5.1 Substantive Documents

아래 문서는 20,000자 하한을 적용한다.

| File | Baseline chars | Final chars | Target | Current status |
|---|---:|---:|---|
| `01_os_kernel_foundations.md` | 17,705 | 23,085 | >= 20,000 | PASS |
| `01a_process_scheduling.md` | 7,568 | 20,129 | >= 20,000 | PASS |
| `01b_memory_and_address_space.md` | 6,735 | 20,000 | >= 20,000 | PASS |
| `01c_filesystem_page_cache_block_io.md` | 6,606 | 20,189 | >= 20,000 | PASS |
| `01d_network_stack_and_io_multiplexing.md` | 9,272 | 21,946 | >= 20,000 | PASS |
| `01e_concurrency_isolation_observability.md` | 7,762 | 20,324 | >= 20,000 | PASS |
| `02_distributed_system_foundations.md` | 10,500 | 20,671 | >= 20,000 | PASS |
| `03_kafka_deep_dive.md` | 9,038 | 20,434 | >= 20,000 | PASS |
| `04_cassandra_deep_dive.md` | 9,423 | 20,807 | >= 20,000 | PASS |
| `05_spark_deep_dive.md` | 10,446 | 21,268 | >= 20,000 | PASS |
| `06_cross_system_comparison.md` | 10,039 | 20,369 | >= 20,000 | PASS |
| `07_interview_reasoning_playbook.md` | 8,231 | 20,030 | >= 20,000 | PASS |
| `08_experiments_and_observability.md` | 10,126 | 20,459 | >= 20,000 | PASS |

### 5.2 Support Documents

| File | Role | Current status |
|---|---|---|
| `README.md` | corpus entry point | PASS |
| `00_index_and_learning_path.md` | learning path and reading order | PASS |
| `09_glossary.md` | terminology bridge | PASS |
| `10_source_ledger.md` | evidence ledger | PASS |
| `audit/claim_review.md` | claim/review surface | PASS |

### 5.3 Deep-Study Content Minimum

Coverage Critic 지적을 반영해, 문자 수는 필요조건으로만 본다. 각 substantive 문서는 아래 내용 기준을 함께 통과해야 한다.

| Requirement | PASS evidence |
|---|---|
| Trace density | 각 문서에 최소 3개 이상의 실제 흐름 trace가 있고, 입력 -> 커널/분산/제품 내부 객체 -> 상태 변화 -> 관측 지점이 이어진다. |
| Counterexample density | 각 문서가 최소 2개 이상의 흔한 오해나 반례를 자연스러운 설명 속에서 교정한다. |
| Observation path | 각 문서가 최소 1개 이상의 관측/실험/명령/로그 해석 경로를 제공한다. |
| Product or interview bridge | OS 문서는 Kafka/Cassandra/Spark/backend 증상으로, 제품 문서는 OS resource와 distributed semantics로 다시 연결된다. |
| Source linkage | load-bearing claim은 `10_source_ledger.md` 또는 audit surface에서 official docs/man page/kernel docs/paper와 연결된다. |
| Anti-padding | 반복 heading scaffold, filler definition, source-list bulk, code-fence bulk로 문자 수를 채우지 않는다. |
| Reader journey | 독자가 문서를 덮고 최소 하나의 end-to-end 흐름을 자기 말로 다시 그릴 수 있어야 한다. |

### 5.4 Write Scope

- Allowed write scope:
  - `interviews/os-kernel-distributed-systems-deep-dive/**`
  - `docs/works/WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TEXTBOOK_EXPANSION.md`
- Out-of-scope dirty files: initial `git status --short` shows many modified/untracked files in algorithms, computer_architecture, root interviews guides, jvm, linux, and other directories. They must remain untouched and unstaged.

## 6. Design Decision Ledger

| Decision | Options considered | Support tier | Admission lane | Verification |
|---|---|---|---|---|
| Keep the existing 13 substantive docs and expand each deeply instead of creating 15-30 new OS files immediately. | A: split into many small OS files. B: expand existing 01 hub + 01a-01e OS chapters with granular internal coverage matrix. | T2 Strong Inference | DOWNGRADED then ACCEPT_REPAIR | Coverage Critic rejected file-count/keyword confidence. B is only allowed if matrix rows are mapped to real sections and final reader audit shows no shallow cram. If irreducible crowding remains, split files and apply 20k to every new substantive doc too. |
| Use original Korean monographs rather than textbook chapter summaries. | Direct user constraint plus copyright boundary. | T1 Direct Evidence | APPLY | Source ledger uses topic families and primary sources, not copied textbook prose. |
| Treat 20,000 chars as hard lower bound but not quality proof. | Direct user wording. | T1 Direct Evidence | APPLY | `wc -m` matrix plus style/coverage/source/critic checks. |

## 7. Coverage Matrix Freeze

### 7.1 OS Topic Families

| Topic family | Primary corpus location |
|---|---|
| OSTEP virtualization/concurrency/persistence/security frame | `01a`-`01e`, `02`-`06` |
| CSAPP exceptional control flow, VM, system I/O, network, concurrency | `01_os_kernel_foundations.md`, `01b`, `01c`, `01d`, `01e` |
| OSPP protection, concurrency, virtualization, resource allocation, reliable storage | `01_os_kernel_foundations.md`, `01a`, `01c`, `01e`, `06` |
| Operating System Concepts process/thread/scheduling/sync/deadlock/memory/storage/I/O/filesystem/security/protection/VM/network-distributed rows | `01_os_kernel_foundations.md`, `01a`-`01e`, `02`, `06`, `08` |
| OS roles, protection, kernel/user mode | `01_os_kernel_foundations.md` |
| Boot, kernel init, architecture, exceptional control flow | `01_os_kernel_foundations.md` |
| Trap, interrupt, exception, syscall, context switch | `01_os_kernel_foundations.md`, `01a_process_scheduling.md` |
| Process/thread/task lifecycle, fork/exec/clone/wait/signal | `01a_process_scheduling.md` |
| Scheduling, preemption, multiprocessor, NUMA, affinity | `01a_process_scheduling.md` |
| Synchronization, futex, atomics, memory ordering, cache coherence | `01e_concurrency_isolation_observability.md` |
| Deadlock, livelock, starvation, priority inversion | `01e_concurrency_isolation_observability.md` |
| Virtual memory, paging, TLB, page fault, COW, mmap, replacement, swap, OOM, cgroup | `01b_memory_and_address_space.md` |
| Memory hierarchy, cache locality, false sharing | `01b_memory_and_address_space.md`, `01e_concurrency_isolation_observability.md` |
| Linking, loading, executable format, dynamic linking, ABI/calling convention | `01_os_kernel_foundations.md` |
| FD, open file description, inode, dentry, VFS | `01c_filesystem_page_cache_block_io.md` |
| Filesystem layout, metadata, permissions, page cache, writeback | `01c_filesystem_page_cache_block_io.md` |
| `write`, `fsync`, crash consistency, journaling, log-structured thinking | `01c_filesystem_page_cache_block_io.md` |
| Disk/SSD/block layer/I/O scheduler/queue depth | `01c_filesystem_page_cache_block_io.md` |
| Device drivers, DMA, interrupt, polling | `01d_network_stack_and_io_multiplexing.md`, `01c_filesystem_page_cache_block_io.md` |
| NIC RX/TX, NAPI, softirq, IP/TCP/UDP, socket buffers, backlog | `01d_network_stack_and_io_multiplexing.md` |
| Blocking/nonblocking/readiness/completion/select/poll/epoll/kqueue/io_uring/qdisc/backpressure | `01d_network_stack_and_io_multiplexing.md` |
| IPC, virtualization, containers, namespaces, cgroups, capabilities, seccomp | `01e_concurrency_isolation_observability.md` |
| Observability `/proc`, `/sys`, `strace`, `perf`, eBPF, flamegraphs, `ss`, `iostat`, `pidstat`, `tcpdump` | `01e_concurrency_isolation_observability.md`, `08_experiments_and_observability.md` |
| Remote/distributed filesystem and storage caveats | `01c_filesystem_page_cache_block_io.md`, `02_distributed_system_foundations.md`, `06_cross_system_comparison.md` |

### 7.2 Distributed/Product Bridge

| Topic family | Primary corpus location |
|---|---|
| Partial failure, timeout uncertainty, retry/idempotency | `02_distributed_system_foundations.md` |
| Clocks, logical time, ordering, message reordering/duplication | `02_distributed_system_foundations.md` |
| Log replication, leader/follower, membership, failure detector | `02_distributed_system_foundations.md`, `03_kafka_deep_dive.md`, `04_cassandra_deep_dive.md` |
| Quorum, consensus, recovery/replay/snapshot/checkpoint | `02_distributed_system_foundations.md`, `03_kafka_deep_dive.md`, `04_cassandra_deep_dive.md`, `05_spark_deep_dive.md` |
| Kafka partition log, page cache, zero-copy, ISR, high watermark, consumer groups | `03_kafka_deep_dive.md` |
| Cassandra coordinator/replica, commitlog/memtable/SSTable/Bloom/compaction/quorum/gossip | `04_cassandra_deep_dive.md` |
| Spark driver/executor/task/stage/shuffle/spill/memory/checkpoint/backpressure | `05_spark_deep_dive.md` |
| Cross-system comparison and interview transfer | `06_cross_system_comparison.md`, `07_interview_reasoning_playbook.md` |
| Experiments and observability | `08_experiments_and_observability.md` |

## 8. Multi-Agent Review Ledger

| Role | Scope | Status |
|---|---|---|
| Coverage Critic | Attack document split, OS textbook coverage, product bridge depth, padding risk | COMPLETED: DOWNGRADE as-is, ACCEPT_REPAIR if matrix-driven |
| Protocol Sentinel | Veto plan-only, slice-only, <20k, source weakness, forbidden pattern, unrelated staging, push | COMPLETED: completion veto unless final evidence proves size/scope/source/checks/commit/no-push |
| Final Coverage Critic | Read final counts, scans, source reachability, link check, diff check | COMPLETED: WHOLE_COMPLETE candidate after WORK update and scoped commit |
| Final Protocol Sentinel | Read final counts, scans, source reachability, diff check, remaining closure | COMPLETED: content gates PASS; required WORK update, path-limited staging, cached scope check, commit, no push |

### 8.1 Critic Attack Closure

| Attack | Disposition | Repair |
|---|---|---|
| Existing 01a-01e OS split may miss textbook families. | ACCEPT_REPAIR | Added explicit textbook-family rows and section-level mapping requirement. |
| 20k can be padding. | ACCEPT_REPAIR | Added trace/counterexample/observation/product bridge/source/reader journey gates. |
| Product docs can become feature summaries. | ACCEPT_REPAIR | Product docs require end-to-end OS resource trace and distributed semantics trace. |
| Distributed depth may be squeezed into `02` only. | ACCEPT_REPAIR | Distributed matrix spans `02`-`07` and products. |
| Support docs are not under 20k but remain load-bearing. | ACCEPT_REPAIR | Support docs must be role-complete and source/claim ledger must expand with claims. |

### 8.2 Protocol Sentinel Vetoes

Completion remains vetoed if any of these are true: plan-only closure, sample/slice closure, any substantive doc under 20,000 chars, padding/repetitive headings, weak source ledger, forbidden phrases/headings, missing final diff/read evidence, unrelated staging, push, or no commit after gates pass.

## 9. Execution Checklist Freeze

Frozen checklist:

- [x] Every explicit user deliverable is mapped to a file or verification row.
- [x] Every substantive doc is >=20,000 UTF-8 chars by `wc -m`.
- [x] No substantive doc relies on padding via repeated template headings, source lists, or code fences.
- [x] OS coverage matrix has no requested topic family unmapped.
- [x] Distributed systems and Kafka/Cassandra/Spark bridge coverage is not squeezed by OS expansion.
- [x] `README.md`, `00_index_and_learning_path.md`, `09_glossary.md`, `10_source_ledger.md`, `audit/claim_review.md` are role-complete.
- [x] Forbidden exact phrases/headings scan passes.
- [x] Markdown/source/link/reachability checks pass.
- [x] Final critic/sentinel pass reads final evidence.
- [x] `git diff --check` passes.
- [x] `git diff --cached --name-only` before commit contains only allowed scope.
- [x] Commit created; push not performed.

## 10. Progress Ledger

| Slice | Required count | Completed count | Remaining count | Next immediate target |
|---|---:|---:|---:|---|
| Substantive docs >=20k | 13 | 13 | 0 | None |
| Support docs role-complete | 5 | 5 | 0 | None |

## 11. Verification Log

| Check | Command / evidence | Result |
|---|---|---|
| Character matrix | `find os-kernel-distributed-systems-deep-dive -type f -name '*.md' -print0 \| xargs -0 wc -m` | PASS: all 13 substantive docs are >=20,000 chars; smallest is `01b_memory_and_address_space.md` at 20,000. |
| Forbidden pattern scan | `rg -n '나중에\|Mermaid\|TODO\|FIXME\|확인 필요\|실패 모드\|게이트를 닫\|질문을 닫\|```mermaid\|^(#{1,6} )?(왜\|어떻게\|없으면\|실패\|검증)\\s*$' os-kernel-distributed-systems-deep-dive || true` | PASS: no hits after wording repair. |
| Markdown local links | read-only parser over local Markdown links under corpus | PASS: local Markdown links OK. |
| Source reachability sample | `curl -L -I --max-time 15` for man7 write/fsync/execve/epoll/tcp, kernel page_cache/VFS/NAPI/cgroup-v2, Kafka, Cassandra, Spark, Raft, Lamport | PASS: all sampled URLs returned HTTP 200. |
| Markdown diff hygiene | `git diff --check -- os-kernel-distributed-systems-deep-dive ../docs/works/WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TEXTBOOK_EXPANSION.md` | PASS: no whitespace errors. |
| Final Coverage Critic | Multi-agent final coverage review | PASS: WHOLE_COMPLETE candidate after WORK update and scoped commit; no substantive rewrite blocker. |
| Final Protocol Sentinel | Multi-agent final protocol review | PASS after WORK update, path-limited staging, cached scope check, commit, no push. |
| Staged scope | `git diff --cached --name-only \| sort` | PASS: verified immediately before commit; only allowed corpus files and this WORK ledger are staged. |
| Commit / push | `git commit` followed by `git show --name-only --stat HEAD`; no `git push` | PASS: commit created; push not performed. |

## 12. Final Status

WHOLE_COMPLETE. Remaining executable count: 0. Next immediate target: none.
