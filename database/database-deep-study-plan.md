# 데이터베이스 완전 분해 학습 계획

이 문서는 `database/` 아래에 흩어져 있는 기존 정리와 앞으로 작성할 장문 학습 문서를 하나의 DB 심화 학습 체계로 묶기 위한 계획서입니다.
이전 계획은 DB 트랜잭션, MVCC, 고립 수준, 따닥 이슈, 금액 계산, 은행/결제 시나리오를 중심으로 세웠습니다.
이번 요청의 보호 의도는 더 큽니다.
목표는 특정 이슈 몇 개를 정리하는 것이 아니라, 데이터베이스를 저장 구조, SQL 실행, 인덱스, 옵티마이저, 트랜잭션, 동시성, 복구, 복제, 분산, 운영, 애플리케이션 경계, 금융 무결성까지 쪼개어 머릿속에서 다시 조립할 수 있게 만드는 것입니다.

여기서 `완벽하고 완전한 내용`은 무근거로 모든 DBMS를 다 안다고 주장하는 뜻이 아닙니다.
이 계획에서의 완전성은 다음 조건을 뜻합니다.
먼저 DB의 주요 작동 축을 빠뜨리지 않는 유한한 레지스트리를 둡니다.
그다음 각 축마다 공식 문서, 로컬 기존 정리, 실험, 반례, 관측 명령을 연결합니다.
마지막으로 작성 완료를 말하기 전에 각 주요 학습 단위가 최소 20,000자 이상인지, 근거가 있는지, 반복으로 분량을 채우지 않았는지, 다른 문서와 충돌하지 않는지 기계적으로 검증합니다.

## 1. review-kernel 판정

review target:
기존 `database/db-transactions-deep-study-plan.md` 계열 계획과 사용자의 최신 요구입니다.

protected purpose:
DB 전체를 실제로 이해 가능한 내부 모델로 분해하고, 기존 `database/` 자산을 살리면서, 장문 학습 문서 단위마다 깊이와 검증을 보장하는 것입니다.

verdict:
`REWORK`.
기존 계획은 트랜잭션과 금융 시나리오에는 강하지만, DB 전체 학습 체계로는 범위가 좁습니다.
SQL 언어 모델, 관계 모델, 저장 포맷, 페이지와 버퍼, 인덱스, 옵티마이저, 실행기, 통계, 조인, 정렬, 스키마 설계, 백업/복구, 복제, 파티셔닝, 보안, 운영 관측, MySQL/PostgreSQL 엔진별 심화, 검색/NoSQL 비교가 독립 작성 단위로 고정되어 있지 않습니다.

repair disposition:
이 파일을 `database/database-deep-study-plan.md`로 승격하고, 기존 12개 트랜잭션 중심 단위는 전체 DB 체계의 `transaction/concurrency/application-integrity` 구간으로 흡수합니다.
작성 위치는 `database/deep-dive/`를 기본값으로 두어 기존 짧은 메모와 새 정식 monograph가 섞이지 않게 합니다.
기존 `database/*.md` 파일은 삭제하거나 덮어쓰는 대상이 아니라 seed, legacy note, 또는 확장 대상입니다.

evidence boundary:
이 계획의 구조 판단은 현재 repo evidence와 사용자의 최신 요구에 근거합니다.
DBMS별 세부 동작, 표준 SQL, PostgreSQL/MySQL/InnoDB/Oracle/SQL Server 차이는 실제 본문 작성 단계에서 공식 문서와 실험으로 다시 닫아야 합니다.

## 2. 요청 정규화

- goal: DB를 기초 개념부터 내부 실행과 운영 시나리오까지 해체해, 시간이 지나도 다시 복원 가능한 장문 지식 자산을 만든다.
- scope: `/Users/rody/VscodeProjects/study/database` 전체와 연결되는 `interviews`, `jvm/spring`, `knowledge/cards`, `domains/payment`, `domains/firmbanking`, `math` 자료.
- mode: planning, review, design, then staged execute.
- finish for current whole request: 모든 등록된 학습 단위 작성, 각 단위 최소 20,000자 검증, 근거/실험/링크/문장 흐름 검증, 최종 리뷰, commit.
- must_keep: 기존 `database/mvcc.md` 같은 사용자 정리 내용을 최대한 살린다. 내용 누락, 열화, 타협을 허용하지 않는다. 단, `완벽`을 근거 없는 과장으로 쓰지 않고 fail-closed 검증 조건으로 바꾼다.
- extra_checks: 각 major learning unit은 최소 20,000자 이상이어야 한다. 분량은 반복 문장으로 채우지 않고 first brick, worked trace, failure mode, observability, 공식 근거, 실험으로 채운다.

## 3. canonical 배치

새 정식 문서 체계는 `database/deep-dive/` 아래에 둡니다.
`database/` 루트에는 이미 `mvcc.md`, `collation.md`, `placeholder.md`, `prepared_statements.md`, `replication.md`, `lock.md`, `join.md`, `query.md` 같은 학습 메모가 있습니다.
이 파일들을 무작정 한 곳에 합치면 기존 흔적이 깨지고, 새 장문 문서와 짧은 메모가 뒤섞입니다.
따라서 새 체계는 평면 번호 목록이 아니라, 학습자가 어디를 공부하고 있는지 바로 알 수 있는 주제별 하위 디렉터리로 둡니다.

```text
database/
  database-deep-study-plan.md
  deep-dive/
    00-index.md
    foundations/
      01-database-mental-model.md
      02-relational-model-and-sql.md
      03-sql-semantics-types-null-collation.md
    storage-index-optimizer/
      04-storage-files-pages-rows.md
      05-buffer-pool-cache-io.md
      06-wal-undo-redo-recovery.md
      07-index-structures.md
      08-query-execution-operators.md
      09-optimizer-statistics-explain.md
    schema-migration-ops/
      10-schema-design-constraints.md
    transactions/
      11-transaction-lifecycle-acid.md
      12-mvcc-snapshot-visibility.md
      13-isolation-anomalies.md
      14-locks-latches-deadlocks.md
    reliability-distribution/
      15-replication-backup-recovery.md
      16-partitioning-sharding-distribution.md
      23-outbox-saga-2pc.md
    mysql/
      17-mysql-innodb-deep-dive.md
    postgresql/
      18-postgresql-deep-dive.md
    application-boundaries/
      19-application-boundaries.md
      20-idempotency-duplicate-request.md
      21-money-calculation.md
      22-financial-transaction-scenarios.md
    operations/
      24-operations-observability-troubleshooting.md
    security-governance/
      25-security-access-control.md
    search-nosql-newsql/
      26-search-engine-internals.md
      27-document-nosql-modeling.md
      28-newsql-distributed-sql.md
    labs/
      postgres-two-session/
      mysql-two-session/
      spring-transaction/
      money-calculation/
      outbox-race/
    source-map.md
    source-map.tsv
    validation.md
```

이 구조는 `database/`를 새로 만드는 것이 아니라, 이미 있는 `database/`를 정식 지식 축으로 인정하고 그 아래에 장문 교재용 하위 디렉터리를 둡니다.
기존 루트 파일은 곧바로 삭제하지 않습니다.
새 문서 작성 시 사람이 읽는 보존 원칙은 `source-map.md`에 남기고, 기계가 검증할 DU별 출처·실험·보존 처분은 `source-map.tsv`에 기록합니다.
각 기존 파일의 내용은 `preserve`, `merge`, `expand`, `supersede-with-pointer`, `archive-candidate`, `sensitive-source-do-not-promote`, `raw-source-sanitize-first` 중 하나로 처리하고, 실제 이동이나 축약은 별도 review slice에서 diff를 보며 결정합니다.

## 4. 기존 자료 보존 정책

기존 자료는 낮은 품질이라는 이유만으로 버리지 않습니다.
중복, 부정확한 단정, 근거 없는 역사 설명, 비유 과다, 엔진 차이 혼합이 있더라도, 그것은 삭제할 쓰레기가 아니라 더 좋은 문서로 승격할 원재료입니다.

현재 확인된 주요 seed는 다음입니다.

| 기존 파일 | 현재 역할 | 향후 처리 |
|---|---|---|
| `database/mvcc.md` | MVCC 역사, MGA/rollback segment, Oracle/PostgreSQL/MySQL 비교, isolation/lock 연결을 담은 가장 큰 원재료 | `deep-dive/transactions/12-mvcc-snapshot-visibility.md`로 보존형 재구성. 원문에서 살릴 내용과 공식 근거 재검증 대상을 ledger화 |
| `database/collation.md` | 문자열 정렬과 비교 원재료 | `deep-dive/foundations/03-sql-semantics-types-null-collation.md`에 흡수하거나 확장 |
| `database/placeholder.md`, `database/prepared_statements.md`, `database/quote_and_escape.md` | SQL 파라미터, escaping, prepared statement 원재료 | `deep-dive/foundations/02-relational-model-and-sql.md`, `deep-dive/application-boundaries/19-application-boundaries.md`, 보안/SQL injection 설명과 연결 |
| `database/replication.md`, `database/replication_lag.md` | replication과 lag 원재료 | `deep-dive/reliability-distribution/15-replication-backup-recovery.md`로 확장 |
| `database/lock.md`, `database/postgresql/lock.md` | lock 원재료가 있으나 매우 짧음 | `deep-dive/transactions/14-locks-latches-deadlocks.md`로 확장 |
| `database/join.md`, `database/query.md`, `database/mysql/explains/when_join.md` | query/join 원재료 | `deep-dive/storage-index-optimizer/08-query-execution-operators.md`, `deep-dive/storage-index-optimizer/09-optimizer-statistics-explain.md`로 확장 |
| `database/db_diff.md`, `database/examples/db_diff/*`, `database/migration/flyway/flyway.md` | schema diff와 migration 원재료 | `deep-dive/schema-migration-ops/10-schema-design-constraints.md`와 migration/online DDL 절로 확장 |
| `database/mysql/*` | MySQL/InnoDB 조각 지식 | `deep-dive/mysql/17-mysql-innodb-deep-dive.md`에 연결 |
| `database/elasticsearch/*`, `database/opensearch/*` | 검색 엔진 원재료 | `deep-dive/search-nosql-newsql/26-search-engine-internals.md`로 정리 |
| `database/firebase/*` | document/NoSQL 원재료 | `deep-dive/search-nosql-newsql/27-document-nosql-modeling.md`로 정리 |
| `database/newsql.md` | distributed SQL/NewSQL 원재료 | `deep-dive/search-nosql-newsql/28-newsql-distributed-sql.md`로 정리 |

새 문서는 기존 내용을 복사해 양만 늘리는 방식으로 만들지 않습니다.
각 seed는 `원문 내용 -> 보존할 핵심 -> 근거 재검증 -> 새 설명 위치 -> 남는 legacy 파일 처리` 순서로 다룹니다.

## 4.1 current database inventory disposition

아래 표는 현재 `find database -maxdepth 6 -type f`로 확인한 파일 단위 inventory입니다.
이 표는 “전부 읽고 이미 승격했다”는 뜻이 아니라, `source-map.md`가 빠뜨리면 안 되는 최소 coverage ledger입니다.
특히 `auth.ini`, `auth.ini.bak`, query JSON, log 파일은 학습 원자료일 수 있어도 본문에 그대로 복사하지 않습니다.
민감 값, 내부 endpoint, 토큰, 계정 정보가 없는지 먼저 검사하고, 필요하면 redaction 또는 synthetic sample로 바꿔야 합니다.

| current file | disposition | first target |
|---|---|---|
| `database/.DS_Store` | archive-ignore system noise | source-map에서 content 대상 제외 |
| `database/collation.md` | expand | `deep-dive/foundations/03-sql-semantics-types-null-collation.md` |
| `database/database-deep-study-plan.md` | active plan | planning artifact, 본문 seed 아님 |
| `database/db_diff.md` | expand | `deep-dive/schema-migration-ops/10-schema-design-constraints.md` |
| `database/elasticsearch/corret_json.py` | inspect-then-preserve | search tooling appendix 또는 cleanup proposal |
| `database/elasticsearch/elasticdump.md` | expand | `deep-dive/search-nosql-newsql/26-search-engine-internals.md` |
| `database/elasticsearch/es_via_curl.md` | expand | search operations subsection |
| `database/elasticsearch/mapping.md` | expand | search schema/mapping subsection |
| `database/elasticsearch/paginate.md` | compare | RDBMS pagination and search pagination contrast |
| `database/elasticsearch/queries/queries.md` | expand | search query model subsection |
| `database/elasticsearch/queries/range.md` | expand | search range query and type semantics subsection |
| `database/elasticsearch/response.md` | preserve | search response reading subsection |
| `database/elasticsearch/tools/esdump/auth.ini` | sensitive-source-do-not-promote | redact or replace before any quotation |
| `database/elasticsearch/tools/esdump/auth.ini.bak` | sensitive-source-do-not-promote | redact or archive only after explicit review |
| `database/elasticsearch/tools/esdump/query1.json` | raw-source-sanitize-first | synthetic search lab input if safe |
| `database/elasticsearch/tools/esdump/query2.json` | raw-source-sanitize-first | synthetic search lab input if safe |
| `database/elasticsearch/tools/esdump/query3.json` | raw-source-sanitize-first | synthetic search lab input if safe |
| `database/elasticsearch/update.md` | expand | search write/update semantics subsection |
| `database/examples/.DS_Store` | archive-ignore system noise | source-map에서 content 대상 제외 |
| `database/examples/db_diff/.DS_Store` | archive-ignore system noise | source-map에서 content 대상 제외 |
| `database/examples/db_diff/README.md` | preserve | schema diff lab guide |
| `database/examples/db_diff/diff_db.sh` | preserve | schema diff lab executable |
| `database/examples/db_diff/docker-compose.yaml` | preserve | schema diff lab environment |
| `database/examples/db_diff/mysql-connector-j/mysql-connector-j-9.2.0.jar` | binary-lab-dependency | license/size check before promotion |
| `database/examples/db_diff/start.sh` | preserve | schema diff lab executable |
| `database/examples/db_diff/whendiff.result` | preserve | schema diff expected output |
| `database/examples/db_diff/whendiff/dev/init.sql` | preserve | schema diff lab fixture |
| `database/examples/db_diff/whendiff/prod/init.sql` | preserve | schema diff lab fixture |
| `database/examples/db_diff/whensame.result` | preserve | schema diff expected output |
| `database/examples/db_diff/whensame/dev/init.sql` | preserve | schema diff lab fixture |
| `database/examples/db_diff/whensame/prod/init.sql` | preserve | schema diff lab fixture |
| `database/firebase/firebase_model.md` | compare | `deep-dive/search-nosql-newsql/27-document-nosql-modeling.md` |
| `database/firebase/firebase_price.md` | compare | document DB operational cost subsection |
| `database/firebase/firestore.md` | expand | document DB modeling and consistency subsection |
| `database/firebase/firestore_sharded_timestamps.md` | expand | document DB hot-spot/sharding subsection |
| `database/join.md` | expand | `deep-dive/storage-index-optimizer/08-query-execution-operators.md` |
| `database/lock.md` | expand | `deep-dive/transactions/14-locks-latches-deadlocks.md` |
| `database/migration/flyway/flyway.md` | expand | migration/Flyway subsection |
| `database/mvcc.md` | preserve-and-rebuild | `deep-dive/transactions/12-mvcc-snapshot-visibility.md` |
| `database/mysql/charset.sql` | preserve | charset/collation worked example |
| `database/mysql/dsn.md` | expand | application connection boundary subsection |
| `database/mysql/explains/when_join.md` | expand | MySQL EXPLAIN join subsection |
| `database/mysql/information.md` | preserve | MySQL metadata/introspection subsection |
| `database/mysql/order.md` | expand | ordering/index/collation subsection |
| `database/mysql/problems/mysql5_online_large_indexing.md` | expand | online DDL and operational risk subsection |
| `database/mysql/procedure.md` | compare | stored procedure boundary subsection |
| `database/mysql/types/mysql_varchar.md` | expand | MySQL type semantics subsection |
| `database/mysql/update.md` | expand | update/locking/binlog implications subsection |
| `database/newsql.md` | expand | `deep-dive/search-nosql-newsql/28-newsql-distributed-sql.md` |
| `database/ogg.md` | inspect-then-expand | replication/CDC comparison subsection |
| `database/opensearch/log.schema.json` | raw-source-sanitize-first | synthetic schema example if safe |
| `database/opensearch/queries.md` | expand | OpenSearch query comparison subsection |
| `database/placeholder.md` | expand | placeholder and parameter binding subsection |
| `database/postgresql/introspection.log` | raw-source-sanitize-first | PostgreSQL introspection example if safe |
| `database/postgresql/lock.md` | expand | PostgreSQL lock subsection |
| `database/prepared_statements.md` | expand | prepared statement subsection |
| `database/query.md` | expand | logical SQL and execution subsection |
| `database/quote_and_escape.md` | expand | quoting/escaping and injection boundary subsection |
| `database/replication.md` | expand | replication subsection |
| `database/replication_lag.md` | expand | lag diagnosis subsection |
| `database/tmp/diff_db.md` | inspect-then-merge | tmp duplicate candidate, do not delete without diff |

## 5. 전체 학습 경로

DB를 머릿속에 넣으려면 “트랜잭션부터” 시작하면 오히려 중간층이 비어 버립니다.
학습 순서는 아래처럼 아래층에서 위층으로 쌓습니다.

1. DB가 왜 파일, 페이지, row, index, log를 동시에 다루는 시스템인지 잡습니다.
2. 관계 모델과 SQL이 어떤 논리적 언어인지 잡습니다.
3. SQL의 실제 의미인 `NULL`, 3-valued logic, collation, type conversion, prepared statement를 닫습니다.
4. row가 디스크 파일과 page 안에 어떻게 놓이는지 봅니다.
5. buffer pool, OS page cache, fsync, random/sequential I/O 차이를 봅니다.
6. WAL, undo, redo, checkpoint, crash recovery를 봅니다.
7. B+tree, hash, bitmap, LSM, covering index, composite index를 봅니다.
8. scan, filter, join, sort, aggregate, window 같은 실행 연산자를 봅니다.
9. optimizer가 통계와 cost로 plan을 고르는 과정을 봅니다.
10. schema, key, constraint, normalization, migration으로 데이터 모양을 고정합니다.
11. transaction lifecycle과 ACID를 봅니다.
12. MVCC와 snapshot visibility를 봅니다.
13. isolation level과 anomaly를 봅니다.
14. lock, latch, deadlock, wait를 봅니다.
15. replication, backup, PITR, failover를 봅니다.
16. partitioning, sharding, distributed SQL, consensus, NewSQL을 봅니다.
17. MySQL/InnoDB와 PostgreSQL을 각각 실제 엔진으로 다시 읽습니다.
18. 애플리케이션 경계에서 connection pool, driver, ORM, Spring transaction을 봅니다.
19. idempotency, money, financial workflow, outbox/saga로 실무 실패를 닫습니다.
20. 운영 관측과 troubleshooting으로 실제 장애를 다시 따라갑니다.
21. 권한, 역할, 행 단위 보안, 암호화, 감사 로그처럼 DB가 데이터를 보호하는 경계를 봅니다.
22. 검색 엔진, document/NoSQL, NewSQL은 RDBMS의 대체재가 아니라 다른 trade-off로 비교합니다.

## 6. major learning unit registry

아래 `DU`는 deep-study unit입니다.
각 `DU`는 해당 파일의 `##` major section이거나 파일 전체 본문이며, 실제 작성 시 최소 20,000자 이상이어야 합니다.
한 파일에 여러 `DU`가 들어가면 각 `##` 본문을 따로 검증합니다.
`DU`는 내부 실행 단위이지 완료 조건이 아닙니다.
whole request는 모든 `DU`가 작성되고 검증될 때만 닫힙니다.

| id | target | major section | 최소 길이 | 기존 seed |
|---|---|---|---:|---|
| DU01 | `foundations/01-database-mental-model.md` | DB는 어떤 문제를 해결하는 시스템인가 | 20,000자 이상 | `database/query.md`, `interviews/database-storage-search-nosql.md` |
| DU02 | `foundations/01-database-mental-model.md` | 논리 모델과 물리 실행 모델을 분리해 읽기 | 20,000자 이상 | 전체 DB corpus |
| DU03 | `foundations/02-relational-model-and-sql.md` | 관계 모델, tuple, relation, key, set/bag semantics | 20,000자 이상 | `database/join.md` |
| DU04 | `foundations/02-relational-model-and-sql.md` | SELECT 문이 논리적으로 처리되는 순서 | 20,000자 이상 | `database/query.md` |
| DU05 | `foundations/03-sql-semantics-types-null-collation.md` | NULL과 3-valued logic | 20,000자 이상 | existing SQL notes |
| DU06 | `foundations/03-sql-semantics-types-null-collation.md` | type, charset, collation, comparison | 20,000자 이상 | `database/collation.md`, `database/mysql/types/mysql_varchar.md`, `database/mysql/charset.sql` |
| DU07 | `foundations/03-sql-semantics-types-null-collation.md` | placeholder, escaping, prepared statement | 20,000자 이상 | `database/placeholder.md`, `database/prepared_statements.md`, `database/quote_and_escape.md` |
| DU08 | `storage-index-optimizer/04-storage-files-pages-rows.md` | file, page, extent, row, tuple layout | 20,000자 이상 | `database/mvcc.md` partial |
| DU09 | `storage-index-optimizer/04-storage-files-pages-rows.md` | heap table, clustered table, row movement, vacuum/purge | 20,000자 이상 | `database/mvcc.md`, MySQL/PostgreSQL sources |
| DU10 | `storage-index-optimizer/05-buffer-pool-cache-io.md` | buffer pool과 OS page cache | 20,000자 이상 | storage notes to add |
| DU11 | `storage-index-optimizer/05-buffer-pool-cache-io.md` | random/sequential I/O, fsync, flush, checkpoint pressure | 20,000자 이상 | `database/replication.md` partial |
| DU12 | `storage-index-optimizer/06-wal-undo-redo-recovery.md` | WAL/redo/undo를 값 변화 trace로 이해하기 | 20,000자 이상 | `database/mvcc.md`, `database/replication.md` |
| DU13 | `storage-index-optimizer/06-wal-undo-redo-recovery.md` | crash recovery, checkpoint, PITR | 20,000자 이상 | official PostgreSQL/MySQL docs required |
| DU14 | `storage-index-optimizer/07-index-structures.md` | B+tree 구조와 탐색/삽입/분할 | 20,000자 이상 | MySQL/PostgreSQL docs required |
| DU15 | `storage-index-optimizer/07-index-structures.md` | composite, covering, unique, partial/function index | 20,000자 이상 | `database/mysql/order.md` partial |
| DU16 | `storage-index-optimizer/07-index-structures.md` | LSM/hash/bitmap/search index와 RDBMS index 비교 | 20,000자 이상 | `database/elasticsearch/*`, `database/opensearch/*` |
| DU17 | `storage-index-optimizer/08-query-execution-operators.md` | scan, filter, projection, sort, aggregate | 20,000자 이상 | `database/query.md` |
| DU18 | `storage-index-optimizer/08-query-execution-operators.md` | nested loop/hash/merge join | 20,000자 이상 | `database/join.md`, `database/mysql/explains/when_join.md` |
| DU19 | `storage-index-optimizer/08-query-execution-operators.md` | subquery, CTE, window, pagination | 20,000자 이상 | `database/elasticsearch/paginate.md` for comparison |
| DU20 | `storage-index-optimizer/09-optimizer-statistics-explain.md` | statistics, cardinality, selectivity, cost | 20,000자 이상 | explain docs required |
| DU21 | `storage-index-optimizer/09-optimizer-statistics-explain.md` | EXPLAIN/EXPLAIN ANALYZE 읽기 | 20,000자 이상 | `database/mysql/explains/when_join.md`, memory DB perf preference |
| DU22 | `storage-index-optimizer/09-optimizer-statistics-explain.md` | slow query diagnosis and plan regression | 20,000자 이상 | local DB perf memories |
| DU23 | `schema-migration-ops/10-schema-design-constraints.md` | normalization, denormalization, keys, constraints | 20,000자 이상 | DB design docs to add |
| DU24 | `schema-migration-ops/10-schema-design-constraints.md` | migration, schema diff, online DDL, Flyway | 20,000자 이상 | `database/db_diff.md`, `database/examples/db_diff/*`, `database/migration/flyway/flyway.md` |
| DU25 | `transactions/11-transaction-lifecycle-acid.md` | transaction boundary, autocommit, savepoint, ACID | 20,000자 이상 | existing transaction plan |
| DU26 | `transactions/11-transaction-lifecycle-acid.md` | consistency와 도메인 불변식 | 20,000자 이상 | existing transaction plan |
| DU27 | `transactions/12-mvcc-snapshot-visibility.md` | MVCC 역사와 등장 배경 | 20,000자 이상 | `database/mvcc.md` |
| DU28 | `transactions/12-mvcc-snapshot-visibility.md` | PostgreSQL tuple visibility와 vacuum | 20,000자 이상 | `database/mvcc.md`, official PostgreSQL docs |
| DU29 | `transactions/12-mvcc-snapshot-visibility.md` | InnoDB read view, undo, purge | 20,000자 이상 | `database/mvcc.md`, official MySQL docs |
| DU30 | `transactions/13-isolation-anomalies.md` | SQL 표준 isolation과 anomaly | 20,000자 이상 | `database/mvcc.md` |
| DU31 | `transactions/13-isolation-anomalies.md` | PostgreSQL과 InnoDB 실제 차이 | 20,000자 이상 | official docs and two-session labs |
| DU32 | `transactions/14-locks-latches-deadlocks.md` | lock vs latch, row/table/gap/next-key/predicate lock | 20,000자 이상 | `database/lock.md`, `database/postgresql/lock.md` |
| DU33 | `transactions/14-locks-latches-deadlocks.md` | deadlock, wait graph, timeout, retry | 20,000자 이상 | two-session labs |
| DU34 | `reliability-distribution/15-replication-backup-recovery.md` | replication log, binlog/WAL shipping, lag | 20,000자 이상 | `database/replication.md`, `database/replication_lag.md` |
| DU35 | `reliability-distribution/15-replication-backup-recovery.md` | backup, restore, PITR, failover, consistency | 20,000자 이상 | official docs required |
| DU36 | `reliability-distribution/16-partitioning-sharding-distribution.md` | partitioning and pruning | 20,000자 이상 | docs to add |
| DU37 | `reliability-distribution/16-partitioning-sharding-distribution.md` | sharding, resharding, global transaction trade-off | 20,000자 이상 | `database/newsql.md` |
| DU38 | `mysql/17-mysql-innodb-deep-dive.md` | InnoDB architecture: clustered index, buffer pool, redo/undo | 20,000자 이상 | `database/mysql/*`, `database/mvcc.md` |
| DU39 | `mysql/17-mysql-innodb-deep-dive.md` | InnoDB locking/isolation/online DDL | 20,000자 이상 | `database/mysql/problems/mysql5_online_large_indexing.md` |
| DU40 | `postgresql/18-postgresql-deep-dive.md` | PostgreSQL heap, tuple, xmin/xmax, vacuum | 20,000자 이상 | `database/mvcc.md`, `database/postgresql/*` |
| DU41 | `postgresql/18-postgresql-deep-dive.md` | PostgreSQL planner, locks, WAL, replication | 20,000자 이상 | `database/postgresql/introspection.log` |
| DU42 | `application-boundaries/19-application-boundaries.md` | connection, session, transaction manager, pooling | 20,000자 이상 | `jvm/spring/spring_transactional.md` |
| DU43 | `application-boundaries/19-application-boundaries.md` | Spring/JPA/MyBatis transaction boundary | 20,000자 이상 | `jvm/spring/spring_transactional.md`, `interviews/spring-backend-frameworks.md` |
| DU44 | `application-boundaries/20-idempotency-duplicate-request.md` | duplicate request and idempotency key | 20,000자 이상 | `knowledge/cards/K-IDEMPOTENT-EXECUTION-WINNER-ELECTION-REPLAY.md` |
| DU45 | `application-boundaries/21-money-calculation.md` | decimal, rounding, allocation, minor unit | 20,000자 이상 | `math/*`, `domains/payment/*` |
| DU46 | `application-boundaries/22-financial-transaction-scenarios.md` | ledger, balance, payment state machine | 20,000자 이상 | `domains/payment/*`, `domains/firmbanking/*` |
| DU47 | `application-boundaries/22-financial-transaction-scenarios.md` | timeout, unknown state, settlement, reconciliation | 20,000자 이상 | `domains/payment/*`, `domains/firmbanking/*` |
| DU48 | `reliability-distribution/23-outbox-saga-2pc.md` | 2PC/XA/JTA vs saga/outbox | 20,000자 이상 | `knowledge/cards/K-ASYNC-EVENTS-OUTBOX-INBOX-CLAIM-LEASE.md` |
| DU49 | `operations/24-operations-observability-troubleshooting.md` | DB troubleshooting by symptom | 20,000자 이상 | `database/replication_lag.md`, DB perf memories |
| DU50 | `operations/24-operations-observability-troubleshooting.md` | metrics, logs, locks, slow query, capacity | 20,000자 이상 | DB perf memories and official docs |
| DU51 | `security-governance/25-security-access-control.md` | roles, users, privileges, grants, ownership | 20,000자 이상 | official PostgreSQL/MySQL docs required |
| DU52 | `security-governance/25-security-access-control.md` | row-level security, encryption, auditing, secret hygiene | 20,000자 이상 | `database/elasticsearch/tools/esdump/auth.ini`, sensitive-source rules |
| DU53 | `search-nosql-newsql/26-search-engine-internals.md` | Elasticsearch/OpenSearch mapping, indexing, query, scoring | 20,000자 이상 | `database/elasticsearch/*`, `database/opensearch/*` |
| DU54 | `search-nosql-newsql/26-search-engine-internals.md` | search pagination, reindexing, dump/restore, consistency boundary | 20,000자 이상 | `database/elasticsearch/paginate.md`, `database/elasticsearch/elasticdump.md`, `database/opensearch/*` |
| DU55 | `search-nosql-newsql/27-document-nosql-modeling.md` | Firestore/document modeling, consistency, security rules, cost | 20,000자 이상 | `database/firebase/*`, `interviews/database-storage-search-nosql.md` |
| DU56 | `search-nosql-newsql/28-newsql-distributed-sql.md` | NewSQL/distributed SQL, consensus, global transaction trade-offs | 20,000자 이상 | `database/newsql.md`, `interviews/database-storage-search-nosql.md` |

이 56개 DU는 시작 레지스트리입니다.
작성 중 공식 근거와 기존 파일 조사에서 빠진 축이 발견되면 DU를 추가할 수는 있지만, 이미 등록된 DU를 조용히 삭제하거나 축약할 수 없습니다.
만약 어떤 DU가 실제로 독립 20,000자 이상 단위로 부적합하다고 드러나면, 삭제가 아니라 `merge proposal`을 만들고 무엇을 어느 DU가 흡수하는지 기록해야 합니다.

## 7. 각 DU의 최소 내부 구조

각 DU는 최소한 아래 흐름을 가져야 합니다.
표면 라벨은 문서마다 자연스럽게 바꿀 수 있지만, 내용은 빠지면 안 됩니다.

1. 직접 진술:
   이 DU가 답하는 질문을 먼저 말합니다.
2. first brick:
   가장 작은 데이터, SQL, page, row, log, schedule, request trace 중 하나로 시작합니다.
3. mechanism:
   입력, 내부 처리, 상태 변화, 산출물, 소비자를 순서대로 설명합니다.
4. worked trace:
   실제 값이 변하는 표, 두 세션 schedule, EXPLAIN 출력, 로그 timeline, schema 변화 중 하나 이상을 넣습니다.
5. counterexample:
   흔한 오해와 그 오해가 깨지는 상황을 보여 줍니다.
6. engine boundary:
   표준 SQL, PostgreSQL, MySQL/InnoDB, 검색/NoSQL, Spring 같은 경계가 섞이면 반드시 분리합니다.
7. observability:
   어떤 명령, 쿼리, 로그, metric, 테스트로 확인할 수 있는지 적습니다.
8. source boundary:
   공식 문서, 로컬 seed, 실험, 추론을 구분합니다.
9. transfer:
   이 원리가 다른 DBMS나 애플리케이션 설계에서 어떻게 다시 나타나는지 연결합니다.

### 7.1 20,000자 기준을 채우는 방식

20,000자는 분량 목표가 아니라 설명 밀도의 하한선입니다.
각 DU는 같은 말을 길게 반복해서 길이를 맞추면 FAIL입니다.
본문은 최소한 아래 요소를 모두 자연스럽게 포함해야 합니다.

1. 직접 진술과 읽은 뒤의 teach-back 목표
    독자가 이 DU를 읽고 자기 말로 무엇을 설명할 수 있어야 하는지 첫머리에서 고정합니다.
    예를 들어 MVCC라면 “여러 버전이 있으니 읽기와 쓰기가 덜 막힌다”가 아니라, “어떤 트랜잭션이 어떤 row version을 볼 수 있는지 snapshot과 version metadata로 판정할 수 있다”까지 말할 수 있어야 합니다.
2. 등장 배경과 역사적 압력
    개념이 왜 생겼는지 설명합니다.
    단순 연표가 아니라, 이전 방식이 어떤 장애나 성능 한계를 만들었고 그 압력이 어떤 구조를 낳았는지 연결합니다.
3. first brick
    가장 작은 row, page, SQL, transaction schedule, log record, request trace 중 하나로 시작합니다.
    독자가 추상 용어를 만나기 전에 손으로 따라갈 수 있는 첫 상태를 가져야 합니다.
4. ASCII diagram 또는 trace
    구조나 시간 흐름이 있는 DU는 `text` code fence, 표, timeline, before/after block 중 최소 하나를 포함합니다.
    장식용 그림은 통과하지 않습니다.
    그림은 “무엇이 바뀌고, 어떤 규칙이 바꾸고, 다음 소비자가 무엇을 읽는지”를 보여 줘야 합니다.
5. worked example
    SQL, row version, index lookup, EXPLAIN, lock wait, 금액 배분, outbox replay처럼 실제 값을 넣은 예시를 둡니다.
6. senior practical failure traps
    20~30년차 실무자가 후배에게 미리 짚어 주듯, 어떤 지뢰를 밟으면 장애가 나는지 설명합니다.
    “주의해야 한다”가 아니라, 어떤 운영 증상으로 나타나고 어떤 잘못된 조치가 상황을 악화시키는지까지 적습니다.
7. 반례와 오해 수리
    그럴듯하지만 틀린 단축 이해를 하나 이상 공격하고, 왜 틀렸는지 작은 반례로 보여 줍니다.
8. 관측과 검증
    어떤 명령, 쿼리, 로그, metric, lab으로 확인할 수 있는지 PASS/FAIL 신호를 적습니다.
9. 자연스러운 한국어 흐름
    문장은 쉬워야 하지만 얕으면 안 됩니다.
    문단은 `직접 진술 -> 이유 -> 작은 예 -> 메커니즘 -> 실패 지점 -> 검증`으로 이어져야 하며, 독자가 문장 사이를 다시 조립해야 하면 FAIL입니다.
10. source boundary
    공식 문서, 로컬 seed, 실험, 추론을 분리합니다.
    vendor별 차이가 있는 내용은 PostgreSQL, MySQL/InnoDB, 검색/NoSQL, Spring 같은 경계를 섞지 않습니다.

기계 validator는 길이와 일부 구조 신호를 검사하고, critic/humanize-korean/study-explanation review는 논리 흐름과 문장 품질을 검사합니다.
두 축 중 하나라도 실패하면 해당 DU는 complete가 아닙니다.

### 7.2 DU registry와 per-DU teaching spine

canonical registry는 `database/deep-dive/du-registry.tsv`입니다.
Markdown 표는 사람이 빠르게 보는 지도이고, TSV는 validator와 작업자가 읽는 실행 계약입니다.
각 행은 `id`, `target`, `section`, `min_chars`, `source_requirement`, `teaching_spine`, `required_trap`을 가집니다.
작성자는 DU 본문을 쓰기 전에 해당 행을 읽고, `source-map.tsv`의 local seed, official source, lab 또는 observability path, preservation disposition이 본문과 같은 목표를 닫는지 확인해야 합니다.
whole-complete 시점에는 각 DU의 `source_status`가 `verified`여야 합니다.

## 8. source strategy

공식 자료 없이 DB domain truth를 확정하지 않습니다.
본문 작성 중 최소한 아래 source pack을 확인합니다.

- PostgreSQL: SQL command reference, MVCC/concurrency control, transaction isolation, explicit locking, WAL, vacuum, planner statistics, EXPLAIN, replication, backup/PITR.
- MySQL/InnoDB: InnoDB architecture, clustered index, buffer pool, redo/undo, transaction isolation, consistent read, locking read, gap/next-key lock, deadlock, online DDL, replication/binlog.
- Security/access control: PostgreSQL roles/privileges/row-level security, MySQL account/privilege model, TLS/encryption/audit guidance, vendor docs for backup and secret handling.
- SQL standard or vendor-neutral references: SQL logical processing, NULL/3-valued logic, isolation phenomena. 표준 원문 접근이 어려우면 PostgreSQL/MySQL 공식 문서와 신뢰 가능한 DB 교재를 구분해 사용합니다.
- Java/Spring: JDBC transaction, `BigDecimal`, Spring declarative transaction, transaction propagation, rollback rules, reactive transaction.
- HTTP/API: RFC 9110 idempotent methods, idempotency-key draft 또는 결제 API 공식 문서.
- 금융/결제: 공개 가능한 payment/firmbanking 자료와 로컬 `domains/payment/*`, `domains/firmbanking/*`.

## 9. labs and verification

문서만으로 닫히지 않는 내용은 실험을 붙입니다.
실험은 학습자가 직접 실패를 재현하고, 왜 그 설명이 맞는지 확인하는 장치입니다.

필수 lab 후보:

- PostgreSQL two-session lab: isolation, MVCC snapshot, lock wait, deadlock, serializable retry.
- MySQL/InnoDB two-session lab: repeatable read, read committed, gap lock, next-key lock, deadlock, consistent read vs locking read.
- EXPLAIN lab: 같은 쿼리가 통계/인덱스/조건에 따라 다른 plan을 고르는 예.
- WAL/recovery lab: 가능한 범위에서 commit/rollback/checkpoint 관측. 실제 crash는 선택 실험으로 둡니다.
- security lab: least-privilege role, denied query, row-level security policy, audit/log observation을 synthetic schema로 검증.
- Spring transaction lab: proxy, self-invocation, rollback rule, `REQUIRES_NEW`, async boundary.
- money lab: double 실패, `BigDecimal(String)`, rounding mode, allocation remainder.
- idempotency/outbox lab: concurrent request winner election, response replay, worker claim race.

검증 명령은 `database/deep-dive/validation.md`와 `database/deep-dive/validate_deep_dive.py`에 고정합니다.
검증은 출력만 찍고 사람이 알아서 판단하는 방식이 아니라, 실패하면 non-zero로 종료하는 fail-closed validator여야 합니다.

```bash
python3 database/deep-dive/validate_deep_dive.py
```

PASS 기준:

- 등록된 DU 전부 존재.
- 각 DU 본문 20,000자 이상.
- 각 DU에 ASCII/code/table trace가 있음.
- 각 DU에 worked example, senior practical failure trap, 등장 배경, 관측/검증 경로, 자연스러운 문단 흐름을 확인할 수 있는 신호가 있음.
- `source-map.tsv`에서 각 DU의 로컬 seed, 공식 source, 실험 또는 검증 경로, preservation disposition, source status가 비어 있지 않음.
- whole-complete 검증에서 `source-map.tsv`의 모든 `source_status`가 `verified`.
- duplicate long paragraph scan 0.
- fenced code block 균형 PASS.
- local link sanity PASS.
- `git diff --check` PASS.
- active plan과 WORK에 이전 길이 기준이 남아 있지 않음.

FAIL 기준:

- DU 하나라도 누락.
- 20,000자 미만.
- 반복 문단으로 분량을 채움.
- ASCII trace, worked example, senior failure trap, history/origin, observability path 중 하나라도 빠짐.
- 공식 근거가 필요한 엔진별 사실을 로컬 추론만으로 확정.
- 기존 `database/` 내용을 살린다고 해 놓고 실제로는 원문을 추적 없이 덮어씀.
- 전체 완료가 아닌 실행 slice나 checkpoint를 whole complete처럼 표현.

## 10. dialectic claim cards

### C1. 범위 claim

- claim: 기존 transaction plan은 DB 전체 분해 계획으로는 부족하므로, 56개 DU 이상의 DB-wide registry로 확장해야 한다.
- premises: 사용자 최신 요구는 `MVCC뿐만 아니라 DB를 진짜 완전 해체 분해`하는 것이다. 현재 plan은 transaction/MVCC/isolation/idempotency/money/financial에 집중한다.
- strongest attack: 너무 넓은 registry는 실행 불가능한 계획표가 될 수 있고, `완전`이라는 말에 휘둘려 불필요한 주제까지 끌어올 수 있다.
- response lane: ACCEPT_REPAIR.
- repaired claim: DBMS 작동을 설명하는 핵심 축만 DU로 고정한다. 검색/NoSQL/NewSQL은 RDBMS를 이해하기 위한 비교 축으로 포함하고, 제품별 전 기능 나열은 제외한다.
- support tier: T1 for user request and repo scope, T2 for exact DU split.
- admission lane: APPLY for planning.

### C2. 배치 claim

- claim: 새 장문 체계는 `database/deep-dive/` 아래 주제별 하위 디렉터리에 둔다.
- premises: `database/` 루트에는 이미 짧은 메모와 큰 원재료가 섞여 있다. 새 장문 체계를 루트에 대량 생성하면 기존 구조가 더 혼잡해진다.
- strongest attack: 하위 디렉터리를 만들면 기존 파일과 새 문서가 분리되어 사용자가 기존 정리를 놓칠 수 있다.
- response lane: ACCEPT_REPAIR.
- repaired claim: `source-map.md`와 `00-index.md`를 필수로 두어 기존 파일과 새 deep-dive 문서의 연결을 추적하고, `foundations`, `storage-index-optimizer`, `transactions`, `application-boundaries` 같은 디렉터리로 학습 위치를 분명히 한다.
- support tier: T2 repo-structure inference.
- admission lane: APPLY for planning.

### C3. 20,000자 DU gate claim

- claim: 최소 20,000자는 파일이 아니라 major DU 단위에 적용해야 한다.
- premises: 사용자 표현은 `각 섹션`이고, 한 파일에 여러 큰 주제가 들어갈 수 있다. 파일 단위 검증은 특정 섹션이 얕아지는 것을 숨길 수 있다.
- strongest attack: 모든 `##`에 20,000자를 강제하면 작은 연결 섹션까지 과도하게 커질 수 있다.
- response lane: ACCEPT_REPAIR.
- repaired claim: `DU registry`에 등록된 major learning section만 20,000자 hard gate로 둔다. 안내, 색인, source-map, validation 같은 운영 섹션은 길이 대상에서 제외한다.
- support tier: T1 for user requirement, T2 for section-unit design.
- admission lane: APPLY for planning.

### C4. 완전성 claim

- claim: `완벽하고 완전한 내용`은 모든 사실을 단번에 확정한다는 뜻이 아니라, 누락과 열화를 막는 fail-closed 작성/검증 체계를 뜻한다.
- premises: DB는 범위가 넓고 vendor별 차이가 커서 무근거 전체 단정은 위험하다.
- strongest attack: 이렇게 말하면 사용자의 강한 품질 요구를 약화하는 것처럼 보일 수 있다.
- response lane: REBUT with clarification.
- repaired claim: 품질 요구는 약화하지 않는다. 오히려 `완전`을 말로 주장하지 못하게 하고, DU registry, source-map, 공식 근거, lab, length proof, critic review가 모두 PASS해야만 완료로 인정한다.
- support tier: T1 from user requirement and rigorous-task closure rules.
- admission lane: APPLY.

### C5. 기존 자료 보존 claim

- claim: 기존 `database/` 파일은 삭제/대체보다 보존형 재구성을 우선한다.
- premises: 사용자는 기존 정리 내용을 최대한 살리고 싶다고 명시했다. `database/mvcc.md` 같은 큰 파일은 학습 흔적과 원재료를 가진다.
- strongest attack: 기존 문서에 중복과 부정확한 단정이 있으면 그대로 살리는 것이 품질을 해칠 수 있다.
- response lane: ACCEPT_REPAIR.
- repaired claim: 문장 그대로 보존이 아니라 의미와 좋은 설명 재료를 보존한다. 부정확하거나 중복된 문장은 source-map에서 `repair`, `merge`, `supersede`로 추적한 뒤 고친다.
- support tier: T1 user feedback and repo evidence.
- admission lane: APPLY.

## 11. multi-agent council record

이번 whole-complete 실행의 topology:

- Orchestrator: Codex main. 요청 재정의, write scope, synthesis, patch, verification, commit을 담당한다.
- DB Curriculum Architect: DB 전체 topic map과 학습 순서의 충분성을 본다.
- Critic: 누락, 과장, 목적 불일치, `완벽` 표현의 위험, 검증 공백을 공격한다.
- Protocol Sentinel: multi-agent/dialectic/review 기록, dirty tree, commit scope, whole-complete 과장 여부를 본다.

최소 debate rounds:

| round | topic | current synthesis |
|---|---|---|
| R1 | problem definition | transaction plan이 아니라 DB-wide deconstruction plan으로 재정의 |
| R2 | evidence breadth | `database/` corpus inventory와 existing seed를 source-map 대상으로 고정 |
| R3 | registry stress | 12개 transaction unit을 56개 DB-wide DU registry로 확장 |
| R4 | validation repair | per-file 검증을 major DU 검증으로 강화 |
| R5 | closure audit | planning checkpoint는 historical artifact일 뿐이고, 현재 요청은 remaining count 0인 WHOLE_COMPLETE만 허용 |

sub-agent 결과가 material하게 다른 누락을 찾으면 이 section은 최종 patch 전에 갱신해야 합니다.

## 12. execution slices

전체 요청은 모든 등록 DU와 비-DU 운영 산출물이 닫힐 때만 완료됩니다.
내부 slice는 작업을 나누기 위한 실행 단위일 뿐이며, verified slice 하나로 멈출 수 없습니다.

권장 순서:

1. Tranche A: `source-map.md`, `00-index.md`, validation script 작성.
2. Tranche B: DU01-DU13, mental model부터 WAL/recovery까지 작성.
3. Tranche C: DU14-DU24, index/query/optimizer/schema/migration 작성.
4. Tranche D: DU25-DU33, transaction/MVCC/isolation/lock 작성. 기존 `database/mvcc.md` 보존형 재구성 포함.
5. Tranche E: DU34-DU41, replication/distribution/MySQL/PostgreSQL deep dive 작성.
6. Tranche F: DU42-DU48, application boundary/idempotency/money/financial/outbox 작성.
7. Tranche G: DU49-DU56, operations/security/search/NoSQL/NewSQL 작성.
8. Tranche H: whole-corpus reverse audit, source-map closure, duplicate scan, local link sanity, final review, commit set 정리.

각 DU 또는 아주 작은 DU batch가 끝날 때마다 남은 DU count와 next immediate target을 갱신합니다.
사용자가 `전체`, `완전`, `누락 없이`를 요구했으므로, 외부 blocker가 없다면 남은 DU가 있는 상태에서 whole complete를 말하지 않습니다.

## 13. downstream impact gate

이 계획은 후속 작성자가 대량 파일 생성과 기존 파일 재구성을 하게 만드는 action surface입니다.
예상 downstream actor는 현재 Codex 실행, 다음 continuation, 또는 사용자입니다.
예상 행동은 `database/deep-dive/` 생성, 기존 `database/` 파일 source-map 작성, 56개 DU 작성, lab 추가, 검증, commit입니다.

되돌림 가능성은 중간 정도입니다.
새 파일 생성은 되돌리기 쉽지만, 기존 `database/mvcc.md` 같은 파일을 직접 고치면 사용자의 기존 정리 흔적이 손상될 수 있습니다.
따라서 기존 파일 수정은 반드시 보존/이동/삭제 ledger를 먼저 만든 뒤 진행합니다.

safer path:

1. source-map을 먼저 작성한다.
2. 새 deep-dive 파일을 먼저 만든다.
3. 기존 파일은 원문 보존 상태에서 새 문서로 의미를 옮긴다.
4. 기존 파일을 축약하거나 pointer로 바꾸는 일은 별도 review와 commit으로 분리한다.

## 14. closure control

- requested whole objective: DB 전체를 완전 분해해 학습 가능한 장문 corpus로 정리한다.
- current execution scope: DB-wide whole-complete writing run.
- achieved closure scope: DU01-DU56 본문, index/source-map/source-map.tsv, validator, validation contract, category lab scripts가 작성되었고 `python3 database/deep-dive/validate_deep_dive.py`가 PASS했다.
- whole-request completion verdict: WHOLE_COMPLETE.
- remaining executable count: 0.
- next immediate target: none inside the requested corpus; final response records the resulting commit hash.
- current-vs-historical evidence: 이전 transaction plan과 Linux/network monograph validation은 precedent일 뿐, DB domain truth는 작성 tranche에서 공식 자료와 실험으로 새로 닫아야 한다.
