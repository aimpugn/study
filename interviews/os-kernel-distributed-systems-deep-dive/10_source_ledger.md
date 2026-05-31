# 10. Source Ledger

> 이 코퍼스의 근거는 공식 문서, 논문, 1차 자료, 고신뢰 자료를 우선합니다.
> 버전별 동작이나 기본값은 확인 날짜와 문서 버전을 함께 남기고, 공식 문서로 닫지 못한 주장은 일반 원리나 강한 추론으로 낮춥니다.
> 본문은 자료 요약본이 아니라 학습 문서이며, 이 ledger는 어떤 설명이 어느 근거에 기대는지 되짚기 위한 지도입니다.

확인 날짜: 2026-05-31

Evidence tier:

- Direct evidence: 공식 문서, 논문, man page가 해당 claim을 직접 지지합니다.
- Strong inference: 직접 문구는 아니지만 여러 직접 근거와 시스템 구조에서 강하게 따라옵니다.
- General principle: 특정 제품 버전보다 넓은 OS/분산 시스템 일반 원리입니다.
- Version-check-needed: 버전, 설정, 배포 모드에 따라 달라질 수 있어 현재 운영 환경에서 재확인이 필요합니다.

## Official Documentation

| Domain | Source | Checked version / page signal | URL | Used for | Tier |
|---|---|---|---|---|---|
| Linux kernel | Linux Kernel documentation top | `7.1.0-rc5` documentation page observed | https://docs.kernel.org/ | Kernel subsystem references, tracing, memory, networking entry points | Direct evidence |
| Linux memory | Page Tables | `7.1.0-rc5` page tables page observed | https://docs.kernel.org/mm/page_tables.html | page tables, MMU, TLB, page fault, lazy allocation, COW, OOM caveats | Direct evidence |
| Linux page cache | Page Cache | `next-20260527` page observed | https://docs.kernel.org/next/mm/page_cache.html | normal file reads/writes/mmaps go through page cache, O_DIRECT caveat | Direct evidence, version-check-needed for implementation details |
| Linux scheduler | CFS Scheduler / EEVDF docs | version-specific scheduler docs observed | https://docs.kernel.org/scheduler/sched-eevdf.html | scheduler fairness and modern EEVDF transition caution | Direct evidence, version-check-needed |
| Linux filesystem | Filesystems in the Linux kernel | current docs page observed | https://docs.kernel.org/filesystems/ | VFS, filesystem support, page cache relation | Direct evidence |
| Linux networking | Networking docs | current docs page observed | https://docs.kernel.org/networking/index.html | network stack and observability entry points | Direct evidence |
| Linux syscall API | Linux man-pages syscalls(2) | man7 page observed | https://man7.org/linux/man-pages/man2/syscalls.2.html | syscall vocabulary and Linux syscall surface | Direct evidence |
| Linux zero-copy | sendfile(2) | man7 page observed | https://man7.org/linux/man-pages/man2/sendfile.2.html | file-to-socket transfer and Kafka zero-copy discussion | Direct evidence |
| Linux readiness I/O | epoll_wait(2) | man7 page observed | https://man7.org/linux/man-pages/man2/epoll_wait.2.html | evented socket observation and server I/O experiments | Direct evidence |
| Kafka | Apache Kafka documentation | releases list showed AK 4.3.x as current docs branch | https://kafka.apache.org/documentation/ | top-level Kafka docs and current branch selection | Direct evidence |
| Kafka design | Apache Kafka 4.3 Design | AK 4.3.x design page observed | https://kafka.apache.org/43/design/design/ | log design, filesystem/page cache, batching, sendfile, consumer position, replication, ISR, leader election, compaction | Direct evidence |
| Kafka operations | Apache Kafka 4.3 Operations | AK 4.3.x operations pages listed | https://kafka.apache.org/43/documentation.html | monitoring, KRaft, transaction protocol, consumer rebalance protocol, eligible leader replicas | Direct evidence, version-check-needed |
| Cassandra | Apache Cassandra 5.0.8 docs | fixed released docs branch identified by source review | https://cassandra.apache.org/doc/5.0.8/index.html | released Cassandra documentation branch for version-sensitive claims | Direct evidence |
| Cassandra latest | Apache Cassandra docs | `/doc/latest/` page may show prerelease signals | https://cassandra.apache.org/doc/latest/ | architecture browsing and cross-checking only when fixed release URL is unavailable | Direct evidence, version-check-needed for defaults |
| Cassandra architecture | Overview | fixed 5.0.8 release docs preferred | https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/overview.html | Cassandra goals, partitioned wide-column model, no cross-partition transactions/joins | Direct evidence |
| Cassandra Dynamo | Dynamo | fixed 5.0.8 release docs preferred | https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/dynamo.html | consistent hashing, token ring, vnodes, RF, replication strategy, tunable consistency | Direct evidence |
| Cassandra storage | Storage Engine | fixed 5.0.8 release docs preferred; URL checked via HTTP 200 | https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/storage-engine.html | commit log, memtable, SSTable components, bloom filter, index, compaction trigger | Direct evidence |
| Cassandra guarantees | Guarantees | fixed 5.0.8 release docs preferred | https://cassandra.apache.org/doc/5.0.8/cassandra/architecture/guarantees.html | CAP framing, availability, eventual consistency, LWT caveat | Direct evidence with caution |
| Cassandra hints | Hints | fixed 5.0.8 release docs preferred | https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/hints.html | hinted handoff, best-effort boundary, hint window default claim | Direct evidence, version-check-needed for defaults |
| Cassandra compaction | Compaction overview | fixed 5.0.8 release docs preferred | https://cassandra.apache.org/doc/5.0.8/cassandra/managing/operating/compaction/overview.html | immutable SSTables, tombstones, read amplification, compaction merge | Direct evidence |
| Spark | Apache Spark docs | Spark 4.1.2 docs observed | https://spark.apache.org/docs/latest/ | Spark version, supported runtime, docs entry | Direct evidence |
| Spark cluster | Cluster Mode Overview | Spark 4.1.2 page observed | https://spark.apache.org/docs/latest/cluster-overview.html | driver, SparkContext, cluster manager, executor, task, web UI | Direct evidence |
| Spark RDD | RDD Programming Guide | Spark 4.1.2 page observed | https://spark.apache.org/docs/latest/rdd-programming-guide.html | RDD, partition, transformations/actions, shuffle, persistence, recomputation | Direct evidence |
| Spark tuning | Tuning Spark | Spark 4.1.2 page observed | https://spark.apache.org/docs/latest/tuning.html | serialization, memory tuning, execution/storage memory, GC, level of parallelism, locality | Direct evidence |
| Spark monitoring | Monitoring | Spark 4.1.2 page observed | https://spark.apache.org/docs/latest/monitoring.html | web UI and monitoring surfaces | Direct evidence |
| Spark streaming | Structured Streaming guide | Spark 4.1.2 page observed, guide split after Spark 4.0 | https://spark.apache.org/docs/latest/streaming/index.html | streaming scope, checkpointing, sink/source semantics, version-sensitive caution | Direct evidence, version-check-needed |

## Papers and Primary References

| Topic | Source | URL | Used for | Tier |
|---|---|---|---|---|
| Logical time | Leslie Lamport, "Time, Clocks, and the Ordering of Events in a Distributed System" | https://lamport.org/pubs/time-clocks.pdf | happens-before, logical clocks, total order as construction rather than wall-clock truth | Direct evidence |
| Consensus | Diego Ongaro and John Ousterhout, "In Search of an Understandable Consensus Algorithm" | https://raft.github.io/raft.pdf | leader election, log replication, safety, consensus teaching model | Direct evidence |
| CAP proof | Gilbert and Lynch, "Brewer's conjecture and the feasibility of consistent, available, partition-tolerant web services" | https://groups.csail.mit.edu/tds/papers/Gilbert/Brewer2.pdf | CAP theorem boundary and partition-time tradeoff | Direct evidence |
| PACELC | Daniel Abadi, "Consistency Tradeoffs in Modern Distributed Database System Design" | https://www.cs.umd.edu/~abadi/papers/abadi-pacelc.pdf | latency/consistency tradeoff outside partitions | Direct evidence |
| Dynamo | DeCandia et al., "Dynamo: Amazon's Highly Available Key-value Store" | https://www.amazon.science/publications/dynamo-amazons-highly-available-key-value-store | consistent hashing, sloppy quorum, hinted handoff, vector clocks, always-on design pressure | Direct evidence |
| Bigtable | Chang et al., "Bigtable: A Distributed Storage System for Structured Data" | https://research.google/pubs/pub27898/ | wide-column storage lineage and SSTable/data model background | Direct evidence |
| LSM tree | O'Neil et al., "The Log-Structured Merge-Tree" | https://dsf.berkeley.edu/cs286/papers/lsm-acta1996.pdf | write-optimized disk data structure and read/write amplification tradeoff | Direct evidence |
| Spark RDD | Zaharia et al., "Resilient Distributed Datasets" | https://people.csail.mit.edu/matei/papers/2012/nsdi_spark.pdf | lineage-based fault tolerance, in-memory cluster computing abstraction | Direct evidence |
| Kafka paper | Kreps, Narkhede, Rao, "Kafka: a Distributed Messaging System for Log Processing" | https://notes.stephenholiday.com/Kafka.pdf | Kafka's original log-processing motivation and design background | Strong inference if not official current behavior |
| MapReduce | Dean and Ghemawat, "MapReduce: Simplified Data Processing on Large Clusters" | https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf | Spark DAG/shuffle lineage as a response to MapReduce-style batch processing | Direct evidence for background |

## Version-Sensitive Claims To Treat Carefully

| Claim area | Safe wording used in corpus |
|---|---|
| Kafka metadata quorum and ZooKeeper/KRaft | Kafka 4.x documentation is KRaft-centered. When explaining older ZooKeeper-era behavior, this corpus labels it historical or version-dependent. |
| Kafka transactions and consumer group protocol | The corpus explains the state movement and marks exact protocol details as version-sensitive unless directly tied to current official docs. |
| Cassandra default values | Defaults such as hint window, vnode count, compaction strategy, and read repair settings can differ by version and deployment. The corpus teaches mechanism first and asks readers to verify defaults in the target version. |
| Cassandra consistency guarantees | Cassandra usually favors availability and tunable consistency, but LWT and newer transaction/consensus features are separate paths. The corpus avoids saying "Cassandra never has strong consistency." |
| Spark current version | Spark docs observed were 4.1.2. Runtime support, Java/Scala/Python versions, streaming guide structure, and default GC details are version-sensitive. |
| Linux scheduler | Modern Linux scheduler behavior is version- and configuration-sensitive. The corpus teaches runnable queue, fairness, latency, and context switch as stable concepts and marks CFS/EEVDF specifics as version-sensitive. |
| Linux page cache internals | The normal page-cache path is stable as a conceptual model, but folios and implementation details are version-specific. |
| macOS observability | macOS tools differ from Linux; `strace`, `/proc`, `ss`, `iostat` options, and eBPF availability are platform-specific. |

## Claim Wording Repairs From Source Review

| Risky wording | Corpus-safe wording |
|---|---|
| Kafka guarantees exactly-once. | Kafka can build exactly-once processing for supported Kafka-internal read-process-write paths when idempotent producer, transactions, offset commits in the transaction, and read isolation are configured correctly. External systems require their own idempotency or transaction boundary. |
| Kafka barely uses disk. | Kafka lowers disk and network cost with append-only logs, page cache, batching, and sendfile. Catch-up reads, cache misses, retention, compaction, TLS, and flush settings can still make disk I/O visible. |
| Cassandra hints guarantee eventual consistency. | Hints are best-effort and reduce inconsistency windows; anti-entropy repair is the stronger convergence mechanism. |
| Cassandra QUORUM is always latest. | Quorum reads depend on RF, write CL, read CL, timestamp rules, failures, and repair state. Quorum intersection improves recency but is not a blanket linearizable-read claim. |
| Spark shuffle is network transfer. | Spark shuffle combines repartitioning, serialization, disk files, network transfer, memory pressure, and possible spill. |
| Page cache write means durable write. | A successful `write()` can mean bytes reached kernel memory, not the physical device. Durability depends on `fsync()`, filesystem, storage stack, and application-specific replication/failure model. |

## Evidence Boundaries

This corpus does not claim to cover every implementation detail of Linux, Kafka, Cassandra, or Spark. It aims to build a lower-layer reasoning model that remains useful when exact settings or versions differ.

When a statement says "일반적으로", read it as a mechanism-level default expectation, not a guarantee for every release. When a statement names a config or current default, check the target version's official documentation before using it in production.
