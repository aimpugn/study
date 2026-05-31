# Claim Review Audit

검수 날짜: 2026-05-31

이 문서는 학습 본문이 아니라 이번 corpus rewrite의 review ledger입니다. 이전 버전의 label-heavy PASS 판정은 superseded되었습니다. 또한 이전 `READY_FOR_SCOPED_COMMIT` 판정도 OS 세부 범위 누락 공격을 받은 뒤 superseded되었습니다. 현재 기준은 prose-first deep study, lower-layer trace, source boundary, unsafe command repair, claim-level evidence, 그리고 OS 상세 문서와 제품 문서의 실제 연결입니다.

## Activated Roles

| Role | Responsibility | Result |
|---|---|---|
| Curriculum Critic | gold slice가 독자의 반문을 닫는지 검수 | ACCEPT_REPAIR after one wording fix |
| Evidence Auditor | `write()`, fd, page cache, fsync, Cassandra/Spark/Kafka source boundary 공격 | ACCEPT_REPAIR / DOWNGRADE rows incorporated |
| Protocol Sentinel | premature WHOLE_COMPLETE, unsafe command, source ledger weakness, staging boundary 감시 | HOLD before repairs, incorporated into rework gates |
| Main Orchestrator | repairs, corpus rewrite, verification, commit boundary | in progress until final verification and commit |
| Coverage Critic | current corpus over-closed OS scope, network path, multiplexing, glossary/source/experiments gaps | REWORK, then ACCEPT_REPAIR after OS detail expansion |

## Claim Cards

| Claim | Attack | Closure | Repair |
|---|---|---|---|
| The previous corpus already met the original intent. | User explicitly rated it around 30-40/100 and objected to checklist-like labels, vague boundary wording, and shallow mechanisms. | DOWNGRADE | Treat previous corpus as source material, not accepted baseline. |
| Gold-slice-first is a valid rewrite strategy. | If the slice only gets longer but does not connect syscall/page cache to Kafka/Cassandra/Spark, it will not generalize. | ACCEPT_REPAIR | `01` section 1 now includes boundary definition, historical causality, syscall descent, partial write, page cache/durability, and product bridges. |
| `경계` can be used in explanations. | The word is harmful if left abstract; reader asks "what boundary and how?" | ACCEPT_REPAIR | Define boundary as CPU privilege, address access, and kernel-owned data structure access before using it. |
| `write()` success means requested bytes reached the file. | `write(2)` can return partial byte count; durable storage is not guaranteed. | ACCEPT_REPAIR | Add partial-write and `fsync()` separation; source ledger maps to man7 pages. |
| Kafka is fast because it avoids disk. | Kafka docs emphasize filesystem/page cache and log design; "no disk" is false. | ACCEPT_REPAIR | Rewrite Kafka as disk-friendly append/page-cache/batch/sendfile design. |
| Cassandra QUORUM is always latest. | RF/CL intersection helps but is not blanket linearizability; repair/timestamp/LWT matter. | ACCEPT_REPAIR | Rewrite Cassandra consistency as read/write CL trace with caveats. |
| Spark shuffle is network transfer. | Shuffle includes serialization, local disk, spill, memory, GC, skew. | ACCEPT_REPAIR | Rewrite Spark shuffle as multi-resource state movement. |
| Experiments are safe enough if they mention warnings. | Broad process-kill patterns can kill unrelated processes; `nodetool compact` is heavy. | ACCEPT_REPAIR | Replace with PID/trap cleanup; mark compaction as local-only opt-in and not normal verification. |
| Source ledger can be a bibliography. | User requested evidence/source boundaries and version-sensitive claim handling. | ACCEPT_REPAIR | Rewrite source ledger as claim-level mapping with official URLs, tier, boundary. |
| Whole corpus can pass by keyword/file-count checks. | User quality floor is reader journey and lower-layer explanation; structural checks are not sufficient. | REBUT | Verification must include prose-label scan, unsafe command scan, source URL checks, and reader journey audit. |
| OS lock/concurrency coverage is enough if lock contention is mentioned. | Lock contention alone does not teach race, deadlock, lost wakeup, memory ordering, or JVM wait states. | ACCEPT_REPAIR | Expanded `01` lock section with shared-state protection, lost update, deadlock, lost wakeup, memory ordering, JVM/OS wait trace. |
| Consensus coverage is enough if quorum is contrasted with consensus. | A reader still needs a state-machine replication trace with leader term, log index, majority append, commit, apply. | ACCEPT_REPAIR | Expanded `02` with Raft-like log entry trace and KRaft/Cassandra boundary. |
| Spark DataFrame coverage can be satisfied by naming DataFrame next to RDD. | The frozen scope asks for RDD/DataFrame/DAG; a reader needs the logical plan -> physical plan -> stage/task bridge. | ACCEPT_REPAIR | Expanded `05` with DataFrame/Dataset, Catalyst plan, physical plan, exchange/shuffle, and explain-plan reasoning boundary. |
| The OS kernel scope can remain inside one 255-line overview file. | The user explicitly rejected a low OS lower bound; critic found process lifecycle, VM, block I/O, network stack, multiplexing, cgroups, namespaces, and observability either missing or too thin. | ACCEPT_REPAIR | Demoted `01_os_kernel_foundations.md` to hub and added `01a` through `01e` detailed OS modules. |
| Network coverage is enough if it mentions socket buffers. | Required packet path includes NIC DMA ring, interrupt/NAPI, softirq, TCP queues, accept/listen queues, epoll readiness, request parsing, send buffer, qdisc/NIC transmit queue. | ACCEPT_REPAIR | Added [01d_network_stack_and_io_multiplexing.md](../01d_network_stack_and_io_multiplexing.md) with inbound/outbound trace and product bridges. |
| Blocking/non-blocking and sync/async are small terminology details. | User called these out as foundational; conflating readiness with async completion causes wrong Netty/Kafka/Spark reasoning. | ACCEPT_REPAIR | Added blocking/non-blocking, synchronous/asynchronous, readiness/completion, select/poll/epoll/kqueue/io_uring discussion in `01d` and interview answer in `07`. |
| Glossary and source ledger already cover the expanded OS scope. | They lacked DMA, NAPI, futex, cgroup, namespace, mmap, OOM, dentry/inode, qdisc, epoll readiness and related primary-source rows. | ACCEPT_REPAIR | Expanded `09` with OS detailed terms and related trace/experiment references; expanded `10` with man7/kernel docs rows. |
| Experiments are sufficient with `write`, page cache, socket queue, JVM dump, product CLI commands. | Missing experiments for epoll readiness, mmap/page fault, cgroup memory, perf/off-CPU entry, and packet path overclaim boundary. | ACCEPT_REPAIR | Expanded `08` with those experiments and explicit PASS/FAIL/overclaim warnings. |

## Gold Slice Confirmation

The `write(fd, buf, len)` gold slice in [01_os_kernel_foundations.md](../01_os_kernel_foundations.md) is accepted as the expansion standard after this repair:

- Original risky wording: kernel code is not in process address space like a library.
- Critic attack: some kernels map kernel memory into process virtual address space; the important claim is not address layout but accessibility and entry path.
- Repair: wording now says mapping differs by OS/architecture, but user code cannot call kernel like an exposed library function and must enter through syscall/trap entry.
- Closure: ACCEPT_REPAIR.

The expansion standard is not "copy this section's length." It is:

- abstract terms are grounded immediately;
- history explains need before naming mechanism;
- API calls are translated into internal objects and buffers;
- success return, durability, consistency, and recovery are separated;
- product sections connect back to OS/distributed-system mechanisms.

## Repairs That Changed The Corpus

| Area | Repair |
|---|---|
| README / 00 | Reframed corpus as self-contained prose-first learning path, not fixed label template. |
| 01 OS | Rewrote entire file around syscall, scheduler, memory, page cache, socket, observability traces. |
| 01a-01e OS detail modules | Added process/scheduling, memory/address, filesystem/block I/O, network/multiplexing, concurrency/isolation/observability as first-class reading path. |
| 02 Distributed | Replaced CAP/quorum keyword flow with partial failure, time/order, log, partition, replication, consistency, recovery, backpressure. |
| 03 Kafka | Rewrote as partition log, page cache, replication/ISR, consumer offset, delivery semantics, compaction; then reconnected broker append/fetch/replica path to OS file/network/scheduler modules. |
| 04 Cassandra | Rewrote as token ownership, commit log/memtable/SSTable, read path, CL/RF, repair, compaction; then reconnected commit log, compaction, repair streaming, cgroup resource limits to OS modules. |
| 05 Spark | Rewrote as driver/executor, lazy graph, DataFrame logical/physical plan, partition/task, dependency/stage, shuffle/spill, lineage/checkpoint; then reconnected shuffle fetch, executor memory, cgroup OOM to OS modules. |
| 06 Comparison | Rebuilt around same words with different promises: log, partition, replication, consistency, recovery, OS resource; then added explicit OS module comparison map. |
| 07 Interview | Rebuilt around state/owner/order/failure/verification and short answer plus tail trace; then added multiplexing, packet path, low-CPU/high-latency answers. |
| 08 Experiments | Repaired unsafe process kill; constrained heavy Cassandra compaction; clarified local-only/pass-fail boundaries; added epoll, mmap/page fault, cgroup, perf/off-CPU, packet-path boundary experiments. |
| 09 Glossary | Added Korean-first meaning, English original, confusion pair, first appearance, interview one-liner; then added OS detailed terminology table with related trace/experiment field. |
| 10 Source Ledger | Replaced source list with claim-level ledger and version-sensitive boundaries; then added official source rows for process lifecycle, mmap, epoll, TCP/socket, NAPI, futex, cgroup, namespace, proc, eBPF. |

## Remaining Verification Before Completion

The corpus is not complete merely because these repairs are written. Before `WHOLE_COMPLETE`, the main workflow must verify and record:

- required target file set exists;
- forbidden Mermaid syntax is absent;
- old repeated labels no longer dominate major topic files;
- unsafe broad process-kill command is absent;
- `nodetool compact` is explicitly local-only/opt-in;
- source URLs that were changed return reachable official pages where possible;
- WORK ledger re-judges every row after current edits;
- staged files are restricted to corpus root and the WORK ledger;
- no push is performed.

## Post-Expansion Critic Confirmation

After the REWORK finding, the repair surface changed materially:

- `01_os_kernel_foundations.md` is no longer treated as the whole OS course. It is a hub.
- `01a` through `01e` now carry the previously missing OS detail.
- `03`, `04`, `05`, and `06` link those OS details back to product internals.
- `08`, `09`, and `10` were expanded so the experiments, glossary, and source ledger no longer describe the old smaller scope.

Closure remains conditional on final verification commands and scoped commit. If any final check finds stale links, forbidden format, source URL failure without boundary note, or unrelated staging, this audit returns to REWORK.
