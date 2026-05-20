# Database deep-dive index

> 2026-05-20 정정: 이 디렉터리는 이전 실행에서 만들어진 DB 장문 초안과 coverage 참고 자료입니다.
> 면접 준비용 정식 DB 심화 학습 자산은 `interviews/database-deep-dive/`에서 claim audit와 composition audit를 거쳐 새로 작성합니다.
> 이 파일의 DU 체계는 보존할 수 있는 원재료와 coverage map으로만 참고하고, 정식 산출물 위치나 완료 판정의 기준으로 사용하지 않습니다.

이 디렉터리는 데이터베이스를 “기능 목록”이 아니라 하나의 작동 시스템으로 다시 조립해 이해하기 위한 장문 학습 축입니다.
각 DU는 최소 20,000자 이상이어야 하며, 길이만 채우는 문서가 아니라 배경, 첫 번째 작동 단위, 손으로 따라가는 trace, 실무 장애 함정, 관측과 검증 경로, 자연스러운 한국어 설명 흐름을 모두 갖춰야 합니다.

이전 초안 체계 안에서의 당시 완료 조건은 `database/deep-dive/validate_deep_dive.py`가 옵션 없이 PASS하고, critic review가 DU 전체에 대해 더 이상 material blocker를 내지 않는 것이었습니다.
이 조건은 현재 `interviews/database-deep-dive/` 정식 재구성 작업의 완료 조건이 아니며, 과거 초안의 coverage를 참고할 때만 역사적 기준으로 봅니다.
중간에 일부 DU가 작성되어도 그것은 checkpoint일 뿐이며, whole-complete가 아닙니다.

## 읽는 순서

1. foundations: DB가 해결하는 문제, 관계 모델, SQL 의미론, NULL, collation, prepared statement를 잡습니다.
2. storage-index-optimizer: page, buffer, WAL, index, 실행기, optimizer를 연결합니다.
3. schema-migration-ops: schema design, constraint, migration, online DDL을 운영 관점으로 읽습니다.
4. transactions: ACID, MVCC, isolation, lock, deadlock을 실제 두 세션 trace로 이해합니다.
5. reliability-distribution: replication, backup, PITR, partitioning, sharding, distributed SQL을 다룹니다.
6. mysql/postgresql: 같은 개념이 InnoDB와 PostgreSQL에서 어떻게 다르게 구현되는지 깊게 비교합니다.
7. application-boundaries: Spring/JDBC transaction, 따닥 요청, 금액 계산, 금융 거래 복구, outbox를 DB 경계와 연결합니다.
8. operations/security/search-nosql-newsql: 운영 진단, 접근 제어, 검색 엔진, document DB, NewSQL까지 확장합니다.

## 이전 초안 DU 목록

| DU | 파일 | 섹션 |
|---|---|---|
| DU01 | `foundations/01-database-mental-model.md` | DB는 어떤 문제를 해결하는 시스템인가 |
| DU02 | `foundations/01-database-mental-model.md` | 논리 모델과 물리 실행 모델을 분리해 읽기 |
| DU03 | `foundations/02-relational-model-and-sql.md` | 관계 모델, tuple, relation, key, set/bag semantics |
| DU04 | `foundations/02-relational-model-and-sql.md` | SELECT 문이 논리적으로 처리되는 순서 |
| DU05 | `foundations/03-sql-semantics-types-null-collation.md` | NULL과 3-valued logic |
| DU06 | `foundations/03-sql-semantics-types-null-collation.md` | type, charset, collation, comparison |
| DU07 | `foundations/03-sql-semantics-types-null-collation.md` | placeholder, escaping, prepared statement |
| DU08 | `storage-index-optimizer/04-storage-files-pages-rows.md` | file, page, extent, row, tuple layout |
| DU09 | `storage-index-optimizer/04-storage-files-pages-rows.md` | heap table, clustered table, row movement, vacuum/purge |
| DU10 | `storage-index-optimizer/05-buffer-pool-cache-io.md` | buffer pool과 OS page cache |
| DU11 | `storage-index-optimizer/05-buffer-pool-cache-io.md` | random/sequential I/O, fsync, flush, checkpoint pressure |
| DU12 | `storage-index-optimizer/06-wal-undo-redo-recovery.md` | WAL/redo/undo를 값 변화 trace로 이해하기 |
| DU13 | `storage-index-optimizer/06-wal-undo-redo-recovery.md` | crash recovery, checkpoint, PITR |
| DU14 | `storage-index-optimizer/07-index-structures.md` | B+tree 구조와 탐색/삽입/분할 |
| DU15 | `storage-index-optimizer/07-index-structures.md` | composite, covering, unique, partial/function index |
| DU16 | `storage-index-optimizer/07-index-structures.md` | LSM/hash/bitmap/search index와 RDBMS index 비교 |
| DU17 | `storage-index-optimizer/08-query-execution-operators.md` | scan, filter, projection, sort, aggregate |
| DU18 | `storage-index-optimizer/08-query-execution-operators.md` | nested loop/hash/merge join |
| DU19 | `storage-index-optimizer/08-query-execution-operators.md` | subquery, CTE, window, pagination |
| DU20 | `storage-index-optimizer/09-optimizer-statistics-explain.md` | statistics, cardinality, selectivity, cost |
| DU21 | `storage-index-optimizer/09-optimizer-statistics-explain.md` | EXPLAIN/EXPLAIN ANALYZE 읽기 |
| DU22 | `storage-index-optimizer/09-optimizer-statistics-explain.md` | slow query diagnosis and plan regression |
| DU23 | `schema-migration-ops/10-schema-design-constraints.md` | normalization, denormalization, keys, constraints |
| DU24 | `schema-migration-ops/10-schema-design-constraints.md` | migration, schema diff, online DDL, Flyway |
| DU25 | `transactions/11-transaction-lifecycle-acid.md` | transaction boundary, autocommit, savepoint, ACID |
| DU26 | `transactions/11-transaction-lifecycle-acid.md` | consistency와 도메인 불변식 |
| DU27 | `transactions/12-mvcc-snapshot-visibility.md` | MVCC 역사와 등장 배경 |
| DU28 | `transactions/12-mvcc-snapshot-visibility.md` | PostgreSQL tuple visibility와 vacuum |
| DU29 | `transactions/12-mvcc-snapshot-visibility.md` | InnoDB read view, undo, purge |
| DU30 | `transactions/13-isolation-anomalies.md` | SQL 표준 isolation과 anomaly |
| DU31 | `transactions/13-isolation-anomalies.md` | PostgreSQL과 InnoDB 실제 차이 |
| DU32 | `transactions/14-locks-latches-deadlocks.md` | lock vs latch, row/table/gap/next-key/predicate lock |
| DU33 | `transactions/14-locks-latches-deadlocks.md` | deadlock, wait graph, timeout, retry |
| DU34 | `reliability-distribution/15-replication-backup-recovery.md` | replication log, binlog/WAL shipping, lag |
| DU35 | `reliability-distribution/15-replication-backup-recovery.md` | backup, restore, PITR, failover, consistency |
| DU36 | `reliability-distribution/16-partitioning-sharding-distribution.md` | partitioning and pruning |
| DU37 | `reliability-distribution/16-partitioning-sharding-distribution.md` | sharding, resharding, global transaction trade-off |
| DU38 | `mysql/17-mysql-innodb-deep-dive.md` | InnoDB architecture: clustered index, buffer pool, redo/undo |
| DU39 | `mysql/17-mysql-innodb-deep-dive.md` | InnoDB locking/isolation/online DDL |
| DU40 | `postgresql/18-postgresql-deep-dive.md` | PostgreSQL heap, tuple, xmin/xmax, vacuum |
| DU41 | `postgresql/18-postgresql-deep-dive.md` | PostgreSQL planner, locks, WAL, replication |
| DU42 | `application-boundaries/19-application-boundaries.md` | connection, session, transaction manager, pooling |
| DU43 | `application-boundaries/19-application-boundaries.md` | Spring/JPA/MyBatis transaction boundary |
| DU44 | `application-boundaries/20-idempotency-duplicate-request.md` | duplicate request and idempotency key |
| DU45 | `application-boundaries/21-money-calculation.md` | decimal, rounding, allocation, minor unit |
| DU46 | `application-boundaries/22-financial-transaction-scenarios.md` | ledger, balance, payment state machine |
| DU47 | `application-boundaries/22-financial-transaction-scenarios.md` | timeout, unknown state, settlement, reconciliation |
| DU48 | `reliability-distribution/23-outbox-saga-2pc.md` | 2PC/XA/JTA vs saga/outbox |
| DU49 | `operations/24-operations-observability-troubleshooting.md` | DB troubleshooting by symptom |
| DU50 | `operations/24-operations-observability-troubleshooting.md` | metrics, logs, locks, slow query, capacity |
| DU51 | `security-governance/25-security-access-control.md` | roles, users, privileges, grants, ownership |
| DU52 | `security-governance/25-security-access-control.md` | row-level security, encryption, auditing, secret hygiene |
| DU53 | `search-nosql-newsql/26-search-engine-internals.md` | Elasticsearch/OpenSearch mapping, indexing, query, scoring |
| DU54 | `search-nosql-newsql/26-search-engine-internals.md` | search pagination, reindexing, dump/restore, consistency boundary |
| DU55 | `search-nosql-newsql/27-document-nosql-modeling.md` | Firestore/document modeling, consistency, security rules, cost |
| DU56 | `search-nosql-newsql/28-newsql-distributed-sql.md` | NewSQL/distributed SQL, consensus, global transaction trade-offs |

## 이전 초안 검증

중간 점검에서는 planned source를 허용할 수 있습니다.

```bash
python3 database/deep-dive/validate_deep_dive.py --allow-planned-sources
```

최종 점검은 옵션 없이 실행합니다.

```bash
python3 database/deep-dive/validate_deep_dive.py
```

이전 초안의 최종 명령이 PASS하려면 모든 DU 본문이 존재하고, 각 section body가 20,000자 이상이며, `source-map.tsv`의 모든 source status가 `verified`여야 했습니다.
현재 면접용 정식 DB 심화 문서의 검증 계약은 `interviews/database-deep-dive/validation.md`와 그 하위 audit 파일을 기준으로 봅니다.
