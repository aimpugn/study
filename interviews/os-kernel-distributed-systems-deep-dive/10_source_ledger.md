# 10. Source Ledger

확인 날짜: 2026-06-01

이 문서는 참고문헌 목록이 아니라 claim-level 근거 장부입니다. 본문은 공식 문서와 논문을 그대로 요약한 문서가 아니며, 여기서는 어떤 설명이 직접 근거인지, 강한 추론인지, 일반 원리인지, 버전 확인이 필요한지 분리합니다.

Evidence tier:

- Direct evidence: 공식 문서, man page, 논문이 해당 claim을 직접 지지합니다.
- Strong inference: 여러 직접 근거와 시스템 구조에서 강하게 따라오지만, 해당 문장이 그대로 있는 것은 아닙니다.
- General principle: 특정 제품 버전보다 넓은 OS/분산 시스템 일반 원리입니다.
- Version-check-needed: 버전, 설정, 배포 모드에 따라 달라질 수 있어 target environment에서 재확인해야 합니다.

## Claim-Level Ledger

| Claim | Primary source | Tier | Boundary |
|---|---|---|---|
| `write()` can return fewer bytes than requested and bad user buffer can return `EFAULT`. | Linux man-pages `write(2)` https://man7.org/linux/man-pages/man2/write.2.html | Direct evidence | POSIX/Linux API claim; wrapper details vary by language/runtime. |
| `fd` is a process-local descriptor referring to an open file description that records offset/status flags. | Linux man-pages `open(2)` https://man7.org/linux/man-pages/man2/open.2.html | Direct evidence | Linux/POSIX model; exact kernel struct names are implementation details. |
| `write()` success is not a durable-storage guarantee; `fsync()` is a stronger synchronization request. | `write(2)` and `fsync(2)` https://man7.org/linux/man-pages/man2/fsync.2.html | Direct evidence | `fsync()` still has filesystem, directory entry, device cache, network filesystem caveats. |
| Normal Linux buffered file reads/writes/mmaps go through page cache; direct I/O and special files are caveats. | Linux kernel Page Cache docs https://docs.kernel.org/mm/page_cache.html | Direct evidence | Kernel implementation details such as folios are version-sensitive. |
| VFS provides common filesystem interface and dispatches to filesystem operations. | Linux kernel filesystem/VFS docs https://docs.kernel.org/filesystems/vfs.html | Direct evidence | VFS does not by itself define application-level durability. |
| `fork()`, `execve()`, `wait()`, signals, and process lifecycle are distinct process-control mechanisms. | Linux man-pages `fork(2)`, `execve(2)`, `wait(2)`, `signal(7)` https://man7.org/linux/man-pages/ | Direct evidence | POSIX/Linux process model; `clone(2)` details are Linux-specific. |
| Linux `clone()` can create task relationships with selected shared resources. | Linux man-pages `clone(2)` https://man7.org/linux/man-pages/man2/clone.2.html | Direct evidence | Exact flags and semantics are Linux-specific and version-sensitive. |
| `mmap()` maps files/devices/anonymous memory into process address space; access can fault pages in later. | Linux man-pages `mmap(2)` https://man7.org/linux/man-pages/man2/mmap.2.html and kernel memory docs | Direct evidence + Strong inference | Page-fault behavior depends on mapping type, filesystem, kernel, and memory pressure. |
| `epoll` is a readiness notification interface, not a blanket async-completion guarantee. | Linux man-pages `epoll(7)`, `epoll_wait(2)` https://man7.org/linux/man-pages/man7/epoll.7.html | Direct evidence | Edge-triggered/level-triggered behavior and fd type details must be checked. |
| TCP sockets expose byte-stream semantics with buffers, windowing, retransmission, and no message boundary preservation. | Linux man-pages `socket(2)`, `tcp(7)` https://man7.org/linux/man-pages/man7/tcp.7.html | Direct evidence | Kernel TCP behavior and sysctl defaults are version/config sensitive. |
| Linux NAPI combines interrupt notification and polling/budgeted packet processing in the network stack. | Linux kernel NAPI docs https://docs.kernel.org/networking/napi.html | Direct evidence | Driver mapping, threaded NAPI, busy polling, and IRQ moderation are hardware/kernel/config sensitive. |
| Futexes let user-space synchronization use kernel wait/wake only when needed. | Linux man-pages `futex(2)` https://man7.org/linux/man-pages/man2/futex.2.html | Direct evidence | Higher-level runtimes may wrap futex differently by version and platform. |
| cgroups limit/account process groups while namespaces isolate what a process sees. | Linux man-pages `cgroups(7)`, `namespaces(7)` https://man7.org/linux/man-pages/man7/cgroups.7.html https://man7.org/linux/man-pages/man7/namespaces.7.html | Direct evidence | cgroup v1/v2 and container runtime behavior differ. |
| `/proc` and `/sys` expose kernel/process/device state as pseudo-filesystems. | Linux man-pages `proc(5)` and kernel sysfs docs https://man7.org/linux/man-pages/man5/proc.5.html | Direct evidence | Linux-specific; macOS uses different observability surfaces. |
| eBPF provides a kernel runtime for instrumentation and extension under verifier/permission constraints. | Linux kernel eBPF userspace API docs https://docs.kernel.org/userspace-api/ebpf/index.html | Direct evidence | Program type, helper availability, and permissions are kernel-version sensitive. |
| Kafka is best understood as partitioned append-only logs with offsets and retention. | Apache Kafka Design docs https://kafka.apache.org/43/design/design/ | Direct evidence | URL is Kafka 4.3 docs path; older releases may differ in metadata/control-plane details. |
| Kafka performance relies on batching, sequential I/O, filesystem/page cache, and transfer optimizations. | Apache Kafka Design docs and Linux `sendfile(2)` https://man7.org/linux/man-pages/man2/sendfile.2.html | Direct evidence + Strong inference | TLS, compression, cache misses, flush settings, storage, and broker version affect actual performance. |
| Kafka exactly-once processing has a supported scope and does not automatically cover arbitrary external side effects. | Apache Kafka documentation https://kafka.apache.org/documentation/ | Direct evidence | Transaction protocol details are version-sensitive; external sinks need their own idempotency/transaction design. |
| Cassandra storage engine uses commit log, memtable, immutable SSTables, Bloom filters, and compaction. | Cassandra 5.0.8 Storage Engine docs https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/storage-engine.html | Direct evidence | Fixed 5.0.8 URL returned HTTP 200 on 2026-05-31; defaults still deployment-sensitive. |
| Cassandra commitlog sync mode affects write acknowledgement and fsync boundary. | Cassandra 5.0.8 Storage Engine / config refs https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/storage-engine.html | Direct evidence | Exact defaults and config names must be checked for target version. |
| Cassandra tunable consistency depends on RF, read CL, write CL, and repair state; QUORUM is not blanket linearizability. | Cassandra architecture/guarantees docs https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/guarantees.html | Direct evidence + Strong inference | LWT/Accord/transaction features are separate paths; version-sensitive. |
| Cassandra hinted handoff is a best-effort convergence aid, not a complete repair substitute; read repair has table-level consistency tradeoffs. | Cassandra Hints docs https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/hints.html and Read Repair docs https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/read_repair.html | Direct evidence | Hint window/defaults and read repair behavior are version/config sensitive. |
| Cassandra compaction merges immutable SSTables and can create background I/O/write amplification. | Cassandra compaction docs https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/compaction/overview.html and LSM paper | Direct evidence | Strategy-specific behavior differs by STCS/LCS/TWCS/UCS. |
| Spark applications have driver programs, cluster managers, executors, jobs, stages, and tasks. | Spark Cluster Mode Overview https://spark.apache.org/docs/latest/cluster-overview.html | Direct evidence | `/docs/latest/` observed as Spark latest docs on 2026-05-31; production version may differ. |
| Spark RDD/DataFrame work is partitioned; transformations are lazy; shuffle redistributes data and can spill. | Spark RDD Programming Guide https://spark.apache.org/docs/latest/rdd-programming-guide.html | Direct evidence | SQL/DataFrame physical execution can add optimizer-specific details not fully covered here. |
| Spark lineage enables recomputation; checkpoint cuts recovery lineage at extra runtime/storage cost. | Spark RDD guide and RDD paper https://people.csail.mit.edu/matei/papers/2012/nsdi_spark.pdf | Direct evidence | External side effects and nondeterministic computations need special care. |
| Partial failure, logical time, and happens-before require reasoning beyond wall-clock timestamps. | Lamport paper https://lamport.org/pubs/time-clocks.pdf | Direct evidence | Wall-clock synchronized systems can add bounded-time assumptions, but not assumed here. |
| Consensus is stronger than quorum response counting. | Raft paper https://raft.github.io/raft.pdf | Direct evidence | Raft is an explanatory consensus protocol, not every system's implementation. |
| CAP is a partition-time impossibility result, not a complete product taxonomy. | Gilbert/Lynch CAP proof https://groups.csail.mit.edu/tds/papers/Gilbert/Brewer2.pdf and PACELC paper https://www.cs.umd.edu/~abadi/papers/abadi-pacelc.pdf | Direct evidence | Product behavior depends on operation, configuration, and failure mode. |
| Executable loading and dynamic linking must be explained as OS/runtime startup boundaries rather than JVM magic. | Linux man-pages `execve(2)` https://man7.org/linux/man-pages/man2/execve.2.html and System V ABI/ELF references | Direct evidence + General principle | Exact executable format and ABI differ by OS/architecture. |
| Dirty writeback, directory fsync, and crash consistency require filesystem/storage-specific caveats. | Linux man-pages `fsync(2)`, kernel filesystem docs, filesystem documentation | Direct evidence + Version-check-needed | Filesystem, mount option, device cache, and network filesystem semantics differ. |
| Failure detectors and membership are suspicion mechanisms, not perfect crash detectors. | Cassandra architecture/gossip docs and distributed systems failure detector literature | Direct evidence + Strong inference | Product algorithms and timeout defaults are version/config sensitive. |
| Kafka consumer lag is a result metric that can come from broker fetch, consumer processing, downstream sink, skew, or rebalance. | Kafka consumer docs and design docs https://kafka.apache.org/documentation/ | Direct evidence + Strong inference | Exact metric names and group protocol details differ by release. |
| Cassandra tombstones preserve deletion semantics until compaction/repair windows make removal safe. | Cassandra tombstone/compaction docs https://cassandra.apache.org/doc/latest/ | Direct evidence | Version and compaction strategy affect details. |
| Spark task retry/speculation can duplicate external side effects unless the sink/commit protocol is idempotent or transactional. | Spark programming guide and structured streaming docs https://spark.apache.org/docs/latest/ | Direct evidence + Strong inference | Sink-specific semantics must be checked for target connector. |
| Container cgroup limits can kill a process even if host-level memory appears available. | Linux man-pages `cgroups(7)` and kernel cgroup docs https://docs.kernel.org/admin-guide/cgroup-v2.html | Direct evidence | cgroup v1/v2 and runtime accounting differ. |
| `io_uring` is a completion-oriented Linux I/O interface with operation/version/security boundaries. | Linux man-pages `io_uring_setup(2)` and kernel docs https://man7.org/linux/man-pages/man2/io_uring_setup.2.html | Direct evidence | Kernel version, operation support, and security policy matter. |

## Primary Sources

| Domain | Source | URL | Used for |
|---|---|---|---|
| Linux syscall | `write(2)`, `open(2)`, `fsync(2)`, `sendfile(2)`, `epoll_wait(2)` | https://man7.org/linux/man-pages/ | syscall contracts, fd model, durability caveats, zero-copy transfer, readiness I/O |
| Linux kernel | Page Cache, VFS, networking, scheduler docs | https://docs.kernel.org/ | OS mechanism boundaries and version-sensitive implementation notes |
| Kafka | Apache Kafka 4.3 docs and design | https://kafka.apache.org/43/design/design/ | log, partition, offset, replication, page cache, batching, sendfile |
| Kafka operations | Apache Kafka documentation | https://kafka.apache.org/documentation/ | operational/version-sensitive protocol and configuration claims |
| Cassandra | Apache Cassandra 5.0.8 docs | https://cassandra.apache.org/doc/5.0.8/ | ring, Dynamo model, storage engine, guarantees, hints, compaction |
| Spark | Apache Spark latest docs | https://spark.apache.org/docs/latest/ | driver/executor, RDD, shuffle, tuning, monitoring, streaming |
| Distributed theory | Lamport, Raft, CAP, PACELC, Dynamo, Bigtable, LSM, RDD papers | paper URLs in rows above | time/order, consensus, consistency tradeoff, storage and compute lineage |

## Reachability Sampling

2026-06-01 작업에서는 final verification에서 official source URL의 reachability를 표본 확인한다. URL이 reachable하더라도 본문 claim의 충분조건은 아니다. 이 ledger는 source가 official/primary인지, claim이 direct인지 inference인지, version boundary가 있는지를 함께 본다.

| URL | Expected role |
|---|---|
| https://man7.org/linux/man-pages/man2/write.2.html | Linux `write(2)` contract |
| https://man7.org/linux/man-pages/man2/fsync.2.html | Linux `fsync(2)` contract |
| https://docs.kernel.org/mm/page_cache.html | Linux page cache mechanism |
| https://docs.kernel.org/networking/napi.html | Linux NAPI network processing |
| https://kafka.apache.org/documentation/ | Kafka official documentation |
| https://cassandra.apache.org/doc/latest/ | Cassandra official documentation |
| https://spark.apache.org/docs/latest/ | Spark official documentation |

## Version-Sensitive Boundaries

- Kafka 4.x documentation is KRaft-centered. Older ZooKeeper-era behavior should be labeled historical or version-dependent.
- Kafka transactions and consumer group protocol details can change by release; mechanism-level explanations are safer than exact internal state names unless tied to current official docs.
- Cassandra defaults such as commitlog sync period, hint window, vnode count, compaction strategy, and read repair behavior must be verified against the target version and configuration.
- Cassandra LWT, Accord, and transaction-related features are separate from the ordinary RF/CL read/write path and must not be collapsed into "Cassandra consistency".
- Spark `/docs/latest/` can move as releases change. Runtime support, PySpark/Scala version, streaming guide structure, and SQL optimizer behavior are version-sensitive.
- Linux scheduler details, page cache internals, and direct-I/O behavior are kernel-version and filesystem dependent. The corpus teaches stable mechanism-level reasoning and marks implementation details as check-before-use.
- macOS observability commands differ from Linux; failed `strace`/`/proc` commands on macOS are tool-boundary failures, not concept refutations.

## Claim Wording Repairs

| Risky wording | Safer wording used in this corpus |
|---|---|
| `write()` means data is on disk. | `write()` may only mean bytes were accepted into the kernel/filesystem path; durability depends on sync, filesystem, storage, and replication model. |
| Kafka is fast because it does not use disk. | Kafka uses disk-friendly append-only logs, batching, page cache, and transfer optimizations; disk can still be the bottleneck. |
| Cassandra QUORUM is always latest. | QUORUM improves read/write intersection but does not by itself imply blanket linearizability. |
| Hinted handoff guarantees convergence. | Hints are best-effort; repair is required for long-term convergence. |
| Spark shuffle is just network. | Spark shuffle includes repartitioning, serialization, local disk files, memory pressure, network transfer, and skew. |
| Retry fixes transient failure. | Retry can duplicate side effects and amplify load unless idempotency, backoff, and backpressure exist. |
