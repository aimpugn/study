# Database interview deep dive

이 디렉터리는 데이터베이스 면접 주제를 정식 deep-dive 학습 자산으로 모아 둔 canonical 위치입니다. `study/database` 아래의 기존 자료와 `database/deep-dive`의 이전 장문 초안은 버릴 자료가 아니라 source와 coverage 참고 자료입니다. 다만 그 문장들을 그대로 옮기지는 않습니다. 각 문서는 source의 작은 claim을 다시 검토하고, 그 claim들이 하나의 큰 설명 안에서 올바른 지식으로 이어지는지 재구성한 결과입니다.

정식 문서는 한 곳에 모읍니다. DB 면접 준비에서 바로 참고할 본문은 이 디렉터리의 문서들을 보면 되고, 기존 `study/database` 자료는 더 오래된 학습 기록이나 원자료를 추적할 때 사용합니다.

## 권장 읽기 순서

파일명 앞의 번호가 기본 읽기 순서입니다. 처음에는 DBMS 전체 mental model에서 시작해 저장/복구/조회/동시성의 내부 구조를 잡고, 그 뒤 운영·분산·애플리케이션 경계와 검색/document store로 확장합니다.

1. [데이터베이스 시스템 mental model](01-database-system-mental-model.md)
2. [저장소, page, buffer pool, I/O](02-storage-pages-buffer-io.md)
3. [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)
4. [인덱스, 실행 계획, optimizer](04-index-query-optimizer.md)
5. [스키마, 제약, 마이그레이션](05-schema-constraints-migration.md)
6. [트랜잭션과 ACID 경계](06-transaction-acid-boundary.md)
7. [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md)
8. [격리 수준, lock, deadlock](08-isolation-lock-deadlock.md)
9. [복제, 지연, 백업, failover](09-replication-lag-backup-failover.md)
10. [파티셔닝, 샤딩, 분산 SQL](10-partition-sharding-distributed-sql.md)
11. [MySQL/InnoDB와 PostgreSQL 엔진 비교](11-mysql-postgresql-engine-deep-dive.md)
12. [애플리케이션 경계, 멱등성, 돈, outbox](12-application-boundaries-idempotency-money-outbox.md)
13. [운영 관측, 보안, troubleshooting](13-operations-security-troubleshooting.md)
14. [검색 엔진과 document NoSQL](14-search-document-nosql-engine.md)

## 읽는 순서

처음 읽을 때는 `01-database-system-mental-model.md`에서 DBMS를 하나의 시스템으로 잡고, `02-storage-pages-buffer-io.md`와 `03-wal-redo-undo-crash-recovery-pitr.md`로 page와 log의 시간축을 잡는 편이 좋습니다. 그 뒤 `04-index-query-optimizer.md`, `06-transaction-acid-boundary.md`, `07-mvcc-snapshot-visibility.md`, `08-isolation-lock-deadlock.md`를 읽으면 성능과 동시성 질문을 하나의 흐름으로 연결할 수 있습니다.

운영과 실무 면접 질문은 뒤쪽 문서에서 이어집니다. 복제와 백업, partition과 sharding, MySQL/PostgreSQL 엔진 차이, 애플리케이션의 멱등성/돈/outbox, 운영 troubleshooting, 검색/NoSQL 저장소를 차례대로 보면 됩니다.

## 검증 방식

기계 검증은 구조적 필요조건입니다. `tools/validate_interview_database_deep_dive.py`는 모든 정식 문서가 필수 섹션, 최소 설명량, source 링크, audit 연결, reader-facing 금지 패턴을 지키는지 확인합니다. 하지만 이 검증만으로 좋은 문서라고 판단하지 않습니다. `audit/claim-audit.tsv`와 `audit/composition-audit.tsv`는 작은 claim의 근거와 큰 문서 흐름의 적합성을 함께 확인하기 위한 표면입니다.

좋은 문서는 면접장에서 먼저 짧게 말할 수 있고, 꼬리 질문을 받으면 작은 모델에서 실제 메커니즘과 DBMS별 경계까지 내려갈 수 있어야 합니다. 이 디렉터리의 문서는 그 기준을 만족하도록 작성되었습니다.
