# Database interview deep dive

이 디렉터리는 데이터베이스 면접 주제를 번호 순서대로 읽고 복습하는 학습 경로입니다.
처음에는 SQL 한 줄이 DBMS 내부에서 어떤 계층을 지나 결과가 되는지 잡고, 그다음 저장 조각인 페이지(page), 변경 기록인 로그(log), 빠른 접근 경로인 인덱스(index), 실행 단위를 묶는 트랜잭션(transaction), 여러 시점의 행을 판단하는 MVCC, 동시 접근을 조율하는 잠금(lock), 변경을 다른 서버로 보내는 복제(replication), 데이터를 나누는 partition과 sharding, 엔진별 차이, 애플리케이션 경계, 운영, 검색/NoSQL로 넓혀 갑니다.

기존 `study/database` 자료와 `database/deep-dive`의 이전 장문 초안은 버릴 자료가 아니라 원자료와 학습 기록입니다. 다만 이 디렉터리의 문서는 그 문장들을 그대로 옮긴 것이 아닙니다. 작은 주장들을 다시 검토하고, 면접 답변으로 말할 수 있는 작은 모델, 상태 trace, 제품별 경계, 함정 질문으로 재구성한 본문입니다.

정식 문서는 한 곳에 모읍니다. DB 면접 준비에서 바로 참고할 본문은 이 디렉터리의 문서들을 보면 되고, 기존 `study/database` 자료는 더 오래된 학습 기록이나 원자료를 추적할 때 사용합니다.

처음 읽을 때 붙잡을 흐름은 아래 하나입니다.

```text
애플리케이션의 SQL 요청
  -> SQL 구조와 값 바인딩
  -> 논리적인 결과 의미
  -> 실행 계획과 행 흐름(row stream)
  -> 페이지(page) / 버퍼 풀(buffer pool) / 인덱스(index) 접근
  -> 트랜잭션이 보는 시점(snapshot) / 잠금(lock) / 먼저 남기는 로그(WAL)
  -> commit, crash recovery, replication, backup, failover
  -> 면접 답변에서 말할 비용과 확인 신호
```

이 흐름을 잡으면 "인덱스가 있으면 빠른가요?", "트랜잭션 격리 수준은 무엇인가요?", "복제 지연은 왜 생기나요?" 같은 질문이 서로 끊어진 암기 문제가 아니라 같은 상태 이동 위에 놓입니다.
특히 [01](01-database-system-mental-model.md)의 UPDATE timeline은 SQL 하나가 잠금, 버전 가시성, 페이지 수정, 로그 안정화, 복제 적용 경계까지 어떻게 이어지는지 보여 주는 공통 기준선입니다.

## 권장 읽기 순서

파일명 앞의 번호가 기본 읽기 순서입니다. 처음에는 DBMS 전체를 설명하는 기본 모델에서 시작해 저장/복구/조회/동시성의 내부 구조를 잡고, 그 뒤 운영·분산·애플리케이션 경계와 검색/document store로 확장합니다.

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

처음 읽을 때는 `01-database-system-mental-model.md`에서 DBMS를 하나의 시스템으로 잡고, `02-storage-pages-buffer-io.md`와 `03-wal-redo-undo-crash-recovery-pitr.md`로 page와 log의 시간축을 잡는 편이 좋습니다.
그 뒤 `04-index-query-optimizer.md`, `06-transaction-acid-boundary.md`, `07-mvcc-snapshot-visibility.md`, `08-isolation-lock-deadlock.md`를 읽으면 성능과 동시성 질문을 하나의 흐름으로 연결할 수 있습니다.

운영과 실무 면접 질문은 뒤쪽 문서에서 이어집니다. 복제와 백업, partition과 sharding, MySQL/PostgreSQL 엔진 차이, 애플리케이션의 멱등성/돈/outbox, 운영 troubleshooting, 검색/NoSQL 저장소를 차례대로 보면 됩니다.

면접 준비 시간이 짧다면 아래 묶음으로 읽습니다.

| 목적 | 먼저 읽을 문서 | 이때 붙잡을 질문 |
| --- | --- | --- |
| DBMS 전체 그림 | [01](01-database-system-mental-model.md), [02](02-storage-pages-buffer-io.md), [03](03-wal-redo-undo-crash-recovery-pitr.md) | SQL 한 줄이 어떤 논리 의미, page 접근, log 기록으로 바뀌는가 |
| 조회 성능 | [04](04-index-query-optimizer.md), [02](02-storage-pages-buffer-io.md) | 같은 결과를 만들 때 왜 어떤 실행 계획은 page를 적게 읽고 어떤 계획은 많이 읽는가 |
| 동시성과 트랜잭션 | [06](06-transaction-acid-boundary.md), [07](07-mvcc-snapshot-visibility.md), [08](08-isolation-lock-deadlock.md) | 동시에 읽고 쓰는 요청이 어떤 snapshot, lock, retry 경계에서 안전해지는가 |
| 운영과 장애 | [09](09-replication-lag-backup-failover.md), [13](13-operations-security-troubleshooting.md) | 느린 쿼리, lock wait, replication lag, failover를 어떤 증거로 좁히는가 |
| 실무 설계 확장 | [10](10-partition-sharding-distributed-sql.md), [12](12-application-boundaries-idempotency-money-outbox.md), [14](14-search-document-nosql-engine.md) | DB 보장만으로 부족한 부분을 application, message, search/document store가 어떻게 나눠 맡는가 |

## 품질 기준

이 문서 묶음은 구조 검사만 통과하면 끝나는 자료가 아닙니다. 좋은 문서인지의 최종 기준은 독자가 작은 모델을 다시 그리고, trace를 따라가며, 제품별 경계와 함정 질문을 자기 말로 설명할 수 있는가입니다.

검증 스크립트는 목차, 필수 섹션, 링크, 기본 구조가 무너지지 않았는지 확인하는 보조 장치입니다. 실제 학습 품질은 본문 안에서 구체 예시, 상태 변화, 실패 신호, DBMS별 차이가 함께 닫히는지로 다시 봐야 합니다.

소스 경계도 함께 읽어야 합니다. 제품별 사실은 PostgreSQL, MySQL, Elasticsearch, Firestore 같은 공식 문서가 기준이고, 이 저장소의 `study/database` 자료와 이전 `database/deep-dive` 초안은 학습 기록과 설명 seed입니다. `audit/*.tsv`와 `validation.md`는 어떤 주장과 구성이 어떤 근거에 기대는지 추적하는 보조 장부이지, 독자가 암기할 본문은 아닙니다.
