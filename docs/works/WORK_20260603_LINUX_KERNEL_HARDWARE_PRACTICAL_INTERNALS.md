# WORK 20260603 Linux Kernel Hardware Practical Internals

## 상태

- Status: COMPLETE after verification
- Work root: `/Users/rody/VscodeProjects/study`
- Primary output: `/Users/rody/VscodeProjects/study/interviews/linux-kernel-hardware-practical-internals.md`
- Minimal index update: `/Users/rody/VscodeProjects/study/interviews/README.md`
- Finish contract: final review, Markdown/diff verification, commit

## 요청 정규화

사용자는 VMware나 observability playbook이 아니라 Linux 커널, 하드웨어, OS 내부 구조, 애플리케이션 자원 사용, 실무 장애 대응, 면접 답변을 한 파일에서 닫는 완결형 교본을 요청했다. 기존 본진 문서는 이번 단계에서 대규모 개편하지 않고, 새 교본 안에 owner/drift map과 후속 보강 목록을 남기는 것이 범위다.

## Instruction Stack

- Global AGENTS: read/apply. Substantial repo-changing study document work requires stack resolution, checklist, evidence, verification, and commit unless excluded.
- Global WORK template: read/apply by compressed ledger. Full template fields were condensed because this work has a narrow write set and no external side effect.
- Study AGENTS: read/apply. Deep-study monograph quality, exemplar floor, source grounding, Markdown style, WORK path, final verification and commit are active.
- Interviews AGENTS: read/apply. Interview output must support short answer, deep mechanism, examples, tail questions, and verification.
- `PROJECT_INTENT.md`: read/apply. The document must train 30-second answer plus deeper OS/runtime/framework descent.
- `USECASE.md`: read/apply. The document must help unknown-question reasoning and interview reconstruction.

## Exemplar Floor

- Primary exemplar: `computer_architecture/threads/threads.md`
    - Used for layered explanation from hardware/OS/runtime/language rather than flat definitions.
- Secondary exemplar: `git/git_rebase.md`
    - Used for concept plus practical command/procedure plus failure-mode structure.
- Supporting pattern: `jvm/java/java_synchronized.md`
    - Used for linking application-level behavior to lower-level runtime and hardware constraints.

## Frozen Checklist

- [x] Create `/Users/rody/VscodeProjects/study/interviews/linux-kernel-hardware-practical-internals.md`.
- [x] Make it a self-contained monograph, not a link hub.
- [x] Keep the center on Linux kernel and hardware internals, not VMware or metric cheat sheets.
- [x] Explain each major command/metric as a view of a kernel object, counter, queue, wait state, or resource contention.
- [x] Include CPU/scheduler.
- [x] Include memory/page cache/reclaim/swap/OOM/cgroup.
- [x] Include filesystem/block I/O/fsync/writeback/device queue.
- [x] Include network/NIC/TCP/socket buffer/backlog/retransmit.
- [x] Include DB as OS application: lock, MVCC, WAL, checkpoint, compaction, buffer pool.
- [x] Include JVM/Spring/Kafka/Cassandra/Spark resource usage.
- [x] Include 1-second incident map, 1-minute read-only commands, 10-minute branching, escalation evidence.
- [x] Include 30-second answer, 2-minute follow-up answer, and command/observation-backed interview answers.
- [x] Identify existing document owners, duplicate allowance, drift prevention, and follow-up reinforcement list.
- [x] Add only minimal README/index update.
- [x] Run final review and Markdown/diff checks.

## Evidence And Scope Ledger

- Read project README and confirmed root curriculum, source reservoir, and promoted deep-dive roles.
- Read OS/kernel deep-dive source ledger and existing OS sections through prior analysis; key reusable surfaces were scheduler, page cache/block I/O, cgroup/observability, TCP/socket, and DB/Kafka/Cassandra/Spark source rows.
- Read database deep-dive surfaces through prior analysis; key reusable surfaces were DB process as OS application, buffer pool/page cache, WAL/fsync, checkpoint, lock wait, compaction, Kafka page cache, Cassandra commit log/SSTable, Spark shuffle.
- Read Linux metric script/material references through prior analysis; key reusable surfaces were `iostat`, `/proc/diskstats`, dirty/writeback, VMware as environment-specific case rather than center.
- Confirmed dirty worktree before editing and kept this work's write set narrow.

## Claim Ledger

| Claim | Support | Critic attack | Final status |
| --- | --- | --- | --- |
| A complete integrated capstone is needed even though topic docs exist. | User preference for one-file completeness plus existing topic docs having specialized ownership. | Risk: duplicate and drift with OS/DB deep dives. | ACCEPT_REPAIR: include owner/drift map and keep broad integration here, detailed evolution in owners. |
| Metrics must be taught as shadows of kernel objects and queues. | User explicitly rejected command/metric cheat sheet and asked what each indicator means. | Risk: still becoming observability playbook. | ACCEPT_REPAIR: document starts from request path, hardware, kernel objects, then maps commands. |
| DB must be explained as an OS application. | User explicitly named lock, MVCC, WAL, checkpoint, compaction, buffer pool. Existing DB docs already use this axis. | Risk: DB section could become DB-only and detach from OS. | ACCEPT_REPAIR: DB section ties each mechanism to CPU, memory, block I/O, network, and wait. |
| Existing topic docs should not be rewritten in this tranche. | User allowed only minimal README/source ledger updates and requested drift map/follow-up list. | Risk: new capstone may hide required owner updates. | ACCEPT_REPAIR: follow-up list names owner docs without changing them now. |

## Verification Log

- `npx --no-install markdownlint-cli2 interviews/linux-kernel-hardware-practical-internals.md interviews/README.md docs/works/WORK_20260603_LINUX_KERNEL_HARDWARE_PRACTICAL_INTERNALS.md`: PASS, 0 errors.
- `git diff --cached --check -- interviews/linux-kernel-hardware-practical-internals.md interviews/README.md docs/works/WORK_20260603_LINUX_KERNEL_HARDWARE_PRACTICAL_INTERNALS.md`: PASS after staging.
- Required-term coverage by `rg`: PASS for CPU/scheduler, memory/page cache/reclaim/swap/OOM/cgroup, block I/O/fsync/writeback/device queue, NIC/TCP/socket buffer/backlog/retransmit, DB mechanisms, JVM/Spring/Kafka/Cassandra/Spark, 1초/1분/10분/30초/2분, owner/drift.
- Final critic review: PASS. The document is Linux kernel/hardware centered and uses VMware only as a drift-map boundary, not the teaching center.

## Remaining Open Work

- Follow-up owner-document improvements are intentionally listed inside the new monograph and not executed in this tranche.
- Existing unrelated dirty/untracked files in the repository were not modified, staged, or committed by this work.
