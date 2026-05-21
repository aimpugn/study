# Database interview deep dive

이 디렉터리는 데이터베이스 면접 주제를 번호 순서대로 읽고 복습하는 학습 경로입니다. 처음에는 SQL 한 줄이 DBMS 내부에서 어떤 계층을 지나 결과가 되는지 잡고, 그다음 page, log, index, transaction, MVCC, lock, replication, partition, engine 차이, 애플리케이션 경계, 운영, 검색/NoSQL로 넓혀 갑니다.

기존 `study/database` 자료와 `database/deep-dive`의 이전 장문 초안은 버릴 자료가 아니라 원자료와 학습 기록입니다. 다만 이 디렉터리의 문서는 그 문장들을 그대로 옮긴 것이 아닙니다. 작은 주장들을 다시 검토하고, 면접 답변으로 말할 수 있는 작은 모델, 상태 trace, 제품별 경계, 함정 질문으로 재구성한 본문입니다.

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
8. [트랜잭션 전파, 격리 수준, lock, deadlock](08-isolation-lock-deadlock.md)
9. [복제, 지연, 백업, failover](09-replication-lag-backup-failover.md)
10. [파티셔닝, 샤딩, 분산 SQL](10-partition-sharding-distributed-sql.md)
11. [MySQL/InnoDB와 PostgreSQL 엔진 비교](11-mysql-postgresql-engine-deep-dive.md)
12. [애플리케이션 경계, 멱등성, 금액 처리, 아웃박스](12-application-boundaries-idempotency-money-outbox.md)
13. [운영, 보안, 장애 분석](13-operations-security-troubleshooting.md)
14. [검색 엔진과 문서형 NoSQL](14-search-document-nosql-engine.md)

## 읽는 순서

처음 읽을 때는 `01-database-system-mental-model.md`에서 DBMS를 하나의 시스템으로 잡고, `02-storage-pages-buffer-io.md`와 `03-wal-redo-undo-crash-recovery-pitr.md`로 page와 log의 시간축을 잡는 편이 좋습니다. 그 뒤 `04-index-query-optimizer.md`, `06-transaction-acid-boundary.md`, `07-mvcc-snapshot-visibility.md`, `08-isolation-lock-deadlock.md`를 읽으면 성능과 동시성 질문을 하나의 흐름으로 연결할 수 있습니다.

운영과 실무 면접 질문은 뒤쪽 문서에서 이어집니다. 복제와 백업, partition과 sharding, MySQL/PostgreSQL 엔진 차이, 애플리케이션의 멱등성/돈/outbox, 운영 troubleshooting, 검색/NoSQL 저장소를 차례대로 보면 됩니다.

## 품질 기준

이 문서 묶음은 구조 검사만 통과하면 끝나는 자료가 아닙니다. 좋은 문서인지의 최종 기준은 독자가 작은 모델을 다시 그리고, trace를 따라가며, 제품별 경계와 함정 질문을 자기 말로 설명할 수 있는가입니다.

검증 스크립트는 목차, 필수 섹션, 링크, 기본 구조가 무너지지 않았는지 확인하는 보조 장치입니다. 실제 학습 품질은 본문 안에서 구체 예시, 상태 변화, 실패 신호, DBMS별 차이가 함께 닫히는지로 다시 봐야 합니다.
