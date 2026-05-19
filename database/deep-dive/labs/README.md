# DB deep-dive labs

이 디렉터리는 본문 설명을 직접 확인하기 위한 작은 실험 모음입니다.
모든 DU가 별도 실행 파일을 갖는 것은 아니지만, `source-map.tsv`의 `lab_or_observability`가 가리키는 관측 경로는 아래 파일 중 하나로 재현하거나 변형할 수 있어야 합니다.

실험은 운영 DB가 아니라 로컬 throwaway schema에서만 실행합니다.
특히 lock, DDL, replication, security 실험은 실제 서비스 계정이나 운영 데이터베이스에서 실행하면 안 됩니다.

## Lab Map

| 범위 | 파일 | 다루는 DU |
|---|---|---|
| SQL 의미론과 기본 모델 | `foundations/sql-semantics-lab.sql` | DU01-DU07 |
| 저장 구조, WAL, 실행 계획 | `storage-index-optimizer/storage-index-explain-lab.sql` | DU08-DU22 |
| 트랜잭션, MVCC, lock, deadlock | `transactions/two-session-transaction-lab.sql` | DU25-DU33 |
| 복제, 백업, 파티션, outbox | `reliability-distribution/reliability-outbox-lab.sql` | DU34-DU37, DU48 |
| InnoDB 관측 | `mysql/innodb-observability-lab.sql` | DU38-DU39 |
| PostgreSQL 관측 | `postgresql/postgresql-observability-lab.sql` | DU40-DU41 |
| 애플리케이션 경계, 금액, 금융 복구 | `application-boundaries/application-money-idempotency-lab.sql` | DU42-DU47 |
| 운영 진단과 보안 | `operations/operations-security-lab.sql` | DU49-DU52 |
| 검색, document DB, distributed SQL | `search-nosql-newsql/search-nosql-newsql-lab.md` | DU53-DU56 |

## 사용 방식

1. 먼저 해당 DU 본문을 읽고, 본문 안의 trace가 어떤 상태 전이를 보여 주는지 말로 설명합니다.
2. lab 파일을 열어 로컬 DBMS에 맞는 부분만 실행합니다.
3. `PASS` 조건은 값 하나가 맞는지가 아니라, 본문에서 설명한 관측 포인트가 실제로 보이는 것입니다.
4. `FAIL` 조건은 SQL이 실패하는 것만이 아닙니다. 기대한 lock wait, snapshot 차이, plan 차이, 권한 거부, rounding 차이가 보이지 않으면 설명 또는 실험 전제가 틀린 것입니다.
