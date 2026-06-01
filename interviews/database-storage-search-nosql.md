# 데이터베이스, 저장소, 검색/NoSQL

- [데이터베이스, 저장소, 검색/NoSQL](#데이터베이스-저장소-검색nosql)
    - [먼저 기억할 정리](#먼저-기억할-정리)
    - [DB 접근과 저장소 운영](#db-접근과-저장소-운영)
        - [3. MySQL 관련 질문](#3-mysql-관련-질문)
            - [원문: 3. MySQL 관련 질문](#원문-3-mysql-관련-질문)
                - [3. MySQL 관련 질문](#3-mysql-관련-질문-1)
                    - [질문 7. MySQL에서 인덱스는 어떤 구조이며, 쿼리 성능을 어떻게 높일 수 있나요?](#질문-7-mysql에서-인덱스는-어떤-구조이며-쿼리-성능을-어떻게-높일-수-있나요)
                    - [**답변**](#답변)
                    - [질문 8. MySQL 트랜잭션 격리 수준과 InnoDB의 MVCC 방식은 어떻게 동작하나요?](#질문-8-mysql-트랜잭션-격리-수준과-innodb의-mvcc-방식은-어떻게-동작하나요)
                    - [**답변**](#답변-1)
        - [DB와 레플리케이션과 고가용성](#db와-레플리케이션과-고가용성)
            - [원문: DB와 레플리케이션과 고가용성](#원문-db와-레플리케이션과-고가용성)
                - [DB와 레플리케이션과 고가용성](#db와-레플리케이션과-고가용성-1)
        - [PDO가 데이터베이스에서 데이터 가져오는 원리](#pdo가-데이터베이스에서-데이터-가져오는-원리)
            - [원문: PDO가 데이터베이스에서 데이터 가져오는 원리](#원문-pdo가-데이터베이스에서-데이터-가져오는-원리)
                - [PDO가 데이터베이스에서 데이터 가져오는 원리](#pdo가-데이터베이스에서-데이터-가져오는-원리-1)
    - [검색 엔진과 샤딩](#검색-엔진과-샤딩)
        - [ElasticSearch 쿼리 과정 예시](#elasticsearch-쿼리-과정-예시)
            - [원문: ElasticSearch 쿼리 과정 예시](#원문-elasticsearch-쿼리-과정-예시)
                - [ElasticSearch 쿼리 과정 예시](#elasticsearch-쿼리-과정-예시-1)
        - [ElasticSearch에서 OOM(Out of Memory) 문제](#elasticsearch에서-oomout-of-memory-문제)
            - [원문: ElasticSearch에서 OOM(Out of Memory) 문제](#원문-elasticsearch에서-oomout-of-memory-문제)
                - [ElasticSearch에서 OOM(Out of Memory) 문제](#elasticsearch에서-oomout-of-memory-문제-1)
        - [Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명](#elasticsearch-oom-문제-hotwarm-아키텍처-설명)
            - [원문: Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명](#원문-elasticsearch-oom-문제-hotwarm-아키텍처-설명)
                - [Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명](#elasticsearch-oom-문제-hotwarm-아키텍처-설명-1)
        - [Elasticsearch의 샤드와 레플리카 샤드](#elasticsearch의-샤드와-레플리카-샤드)
            - [원문: Elasticsearch의 샤드와 레플리카 샤드](#원문-elasticsearch의-샤드와-레플리카-샤드)
                - [Elasticsearch의 샤드와 레플리카 샤드](#elasticsearch의-샤드와-레플리카-샤드-1)
        - [ElasticSearch Hot/Warm 아키텍처](#elasticsearch-hotwarm-아키텍처)
            - [원문: ElasticSearch Hot/Warm 아키텍처](#원문-elasticsearch-hotwarm-아키텍처)
                - [ElasticSearch Hot/Warm 아키텍처](#elasticsearch-hotwarm-아키텍처-1)
    - [검색/NoSQL 저장소](#검색nosql-저장소)
        - [카우치베이스](#카우치베이스)
            - [원문: 카우치베이스](#원문-카우치베이스)
                - [카우치베이스](#카우치베이스-1)
    - [대용량 테이블 운영](#대용량-테이블-운영)
        - [수억건의 데이터가 존재하는 테이블 처리 방법](#수억건의-데이터가-존재하는-테이블-처리-방법)
            - [원문: 수억건의 데이터가 존재하는 테이블 처리 방법](#원문-수억건의-데이터가-존재하는-테이블-처리-방법)
                - [수억건의 데이터가 존재하는 테이블 처리 방법](#수억건의-데이터가-존재하는-테이블-처리-방법-1)
                    - [문제 상황](#문제-상황)
                    - [핵심 제약](#핵심-제약)
                    - [테이블 파티셔닝 (수동 구현)](#테이블-파티셔닝-수동-구현)
                    - [수동 파티셔닝 설계](#수동-파티셔닝-설계)
                    - [조회 시 파티션 통합](#조회-시-파티션-통합)
                    - [데이터 아카이빙](#데이터-아카이빙)
                    - [아카이브 테이블 설계](#아카이브-테이블-설계)
                    - [아카이브 데이터 통합 조회](#아카이브-데이터-통합-조회)
                    - [인덱스 최적화 및 테이블 정리](#인덱스-최적화-및-테이블-정리)
                    - [3.1 중복 인덱스 제거](#31-중복-인덱스-제거)
                    - [3.2 복합 인덱스 추가](#32-복합-인덱스-추가)
                    - [점진적 테이블 분리](#점진적-테이블-분리)
                    - [배치로 데이터 옮기기](#배치로-데이터-옮기기)
                    - [점진적 데이터 분리 전략](#점진적-데이터-분리-전략)
    - [인덱스와 조회 성능](#인덱스와-조회-성능)
        - [3억건의 데이터의 B+Tree 깊이 계산](#3억건의-데이터의-btree-깊이-계산)
            - [원문: 3억건의 데이터의 B+Tree 깊이 계산](#원문-3억건의-데이터의-btree-깊이-계산)
                - [3억건의 데이터의 B+Tree 깊이 계산](#3억건의-데이터의-btree-깊이-계산-1)
                    - [B+ Tree의 주요 특징](#b-tree의-주요-특징)
                    - [B+ Tree의 최대 용량](#b-tree의-최대-용량)
                    - [인덱싱 예시](#인덱싱-예시)
                    - [깊이 계산 방법](#깊이-계산-방법)
                    - [B+ Tree 깊이와 데이터 접근 시간](#b-tree-깊이와-데이터-접근-시간)
                    - [추가 고려 사항](#추가-고려-사항)
        - [B Tree & B+ Tree](#b-tree--b-tree)
            - [원문: B Tree & B+ Tree](#원문-b-tree--b-tree)
                - [B Tree & B+ Tree](#b-tree--b-tree-1)
    - [트랜잭션, 락, 격리](#트랜잭션-락-격리)
        - [2단계커밋과 2단계 잠금](#2단계커밋과-2단계-잠금)
            - [원문: 2단계커밋과 2단계 잠금](#원문-2단계커밋과-2단계-잠금)
                - [2단계커밋과 2단계 잠금](#2단계커밋과-2단계-잠금-1)
        - [ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)](#acidatomicity-consistency-isolation-durability와-mvccmulti-version-concurrency-control-다중-버전-동시성-제어)
            - [원문: ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)](#원문-acidatomicity-consistency-isolation-durability와-mvccmulti-version-concurrency-control-다중-버전-동시성-제어)
                - [ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)](#acidatomicity-consistency-isolation-durability와-mvccmulti-version-concurrency-control-다중-버전-동시성-제어-1)
        - [MVCC와 스냅숏 격리](#mvcc와-스냅숏-격리)
            - [원문: MVCC와 스냅숏 격리](#원문-mvcc와-스냅숏-격리)
                - [MVCC와 스냅숏 격리](#mvcc와-스냅숏-격리-1)
        - [DB - ACID](#db---acid)
            - [원문: DB - ACID](#원문-db---acid)
                - [DB - ACID](#db---acid-1)
        - [데이터베이스 락(Lock)과 격리 수준(Isolation Level)](#데이터베이스-락lock과-격리-수준isolation-level)
            - [원문: 데이터베이스 락(Lock)과 격리 수준(Isolation Level)](#원문-데이터베이스-락lock과-격리-수준isolation-level)
                - [데이터베이스 락(Lock)과 격리 수준(Isolation Level)](#데이터베이스-락lock과-격리-수준isolation-level-1)

트랜잭션, 인덱스, 락, 복제, 파티셔닝, Elasticsearch, Couchbase처럼 데이터를 저장하고 찾는 전체 축을 함께 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## 먼저 기억할 정리

DB와 검색 문서는 SQL 문장이나 제품 이름보다 "row가 어떤 저장 상태와 실행 경로를 지나 결과가 되는가"를 먼저 잡아야 합니다.

```text
query / mutation
  -> parser / planner
  -> index or shard routing
  -> buffer pool / page cache
  -> lock or MVCC visibility check
  -> WAL / transaction log / replication stream
  -> result row, document, or ack
```

비교축은 저장, 조회, 동시성, 복구입니다. 인덱스는 단순히 빠른 자료구조가 아니라 읽어야 할 row stream을 줄이는 물리 구조이고, WAL/redo/undo는 장애 뒤에 어떤 변경을 다시 만들거나 되돌릴지 남기는 기록입니다. Elasticsearch나 Couchbase를 볼 때도 shard, replica, segment, heap, cache, query fan-out처럼 보이지 않는 상태가 어디에 있는지 확인해야 합니다.

검증 anchor는 `EXPLAIN`, slow query log, lock wait/deadlock log, buffer pool 지표, WAL/replication lag, shard allocation, JVM heap/GC 지표입니다. "빠르다", "일관적이다", "고가용이다" 같은 말은 어떤 상태와 어떤 관측값으로 확인되는지까지 내려가야 면접 답변이 단단해집니다.

## DB 접근과 저장소 운영

### 3. MySQL 관련 질문

#### 원문: 3. MySQL 관련 질문

<!-- curriculum-chunk: sha256=6460b6776fe2899a33d8a67993374a814f47c68e83af6056f64776f8c47dacdd major=database-storage-search-nosql mid=DB 접근과 저장소 운영 sub=3. MySQL 관련 질문 sources=source/interview_questions3.md:165-188 -->

> Source: `source/interview_questions3.md:165-188`
> Classification reason: database/storage

##### 3. MySQL 관련 질문

###### 질문 7. MySQL에서 인덱스는 어떤 구조이며, 쿼리 성능을 어떻게 높일 수 있나요?

###### **답변**

1. MySQL(InnoDB 엔진 기준)에서 대부분의 인덱스는 **B+Tree** 구조로 되어 있습니다. B+Tree는 트리 높이를 낮게 유지하도록 설계되어, 많은 데이터를 저장하면서도 빠른 검색을 가능케 합니다.
2. InnoDB에서 **PRIMARY KEY**(기본키) 인덱스는 “클러스터링 인덱스”로서, 실제 테이블 레코드 자체가 B+Tree leaf에 저장됩니다. 반면, 보조 인덱스(secondary index)는 leaf 노드에 PK 값을 별도로 저장하고, 그 PK를 통해 테이블 데이터를 찾아가는 방식입니다.
3. 쿼리가 느릴 때는 **쿼리 실행 계획**(`EXPLAIN`)을 통해 인덱스 사용 여부, 풀 스캔 여부, 조인 방식 등을 점검해야 합니다. 필요한 열(WHERE절, JOIN에 사용되는 열)에 적절한 인덱스를 추가하면 성능이 개선될 수 있습니다.
4. 하지만 지나친 인덱스 생성은 INSERT/UPDATE 시 성능을 떨어뜨릴 수 있으므로, 빈도와 활용도를 고려해 인덱스를 설계해야 합니다.

---

###### 질문 8. MySQL 트랜잭션 격리 수준과 InnoDB의 MVCC 방식은 어떻게 동작하나요?

###### **답변**

1. 트랜잭션 격리 수준은 한 트랜잭션이 다른 트랜잭션의 중간 변경 내용을 얼마나 볼 수 있는지를 정의합니다. MySQL의 기본값은 **REPEATABLE READ**이며, 다른 선택지로 READ COMMITTED, SERIALIZABLE 등이 있습니다.
2. REPEATABLE READ는 같은 트랜잭션 내에서 같은 SELECT 쿼리를 여러 번 실행해도 결과가 동일하게 보이도록 보장합니다.
3. InnoDB는 **MVCC(Multi-Version Concurrency Control)**를 통해 이 격리 수준을 구현합니다. 구체적으로, 각 행에는 트랜잭션 ID 정보가 기록되며, Undo Log에 이전 버전이 보관됩니다. SELECT 시점의 트랜잭션 ID보다 작은 버전 중에서 최신 것을 읽어 **일관된 스냅샷**을 제공합니다.
4. 이렇게 MVCC를 사용하면 대부분의 읽기 작업이 락 없이도 동작 가능해지고, 동시에 쓰기가 일어날 때도 Undo Log를 참고해 충돌을 최소화합니다.

---

<!-- /curriculum-chunk -->

### DB와 레플리케이션과 고가용성

#### 원문: DB와 레플리케이션과 고가용성

<!-- curriculum-chunk: sha256=b771bbde329427fd1fef84e21d961a0a1b502d12b1c9ea3c4e5040bc04d6d575 major=database-storage-search-nosql mid=DB 접근과 저장소 운영 sub=DB와 레플리케이션과 고가용성 sources=source/interview_questions.md:1735-2084, source/interviews.md:1735-2084 -->

> Source: `source/interview_questions.md:1735-2084`
> Classification reason: database/storage
> Duplicate source aliases: `source/interview_questions.md:1735-2084, source/interviews.md:1735-2084`

##### DB와 레플리케이션과 고가용성

데이터베이스 레플리케이션은 단순히 마스터의 모든 변경사항을 그대로 복제하는 것이 아니라, 다음과 같은 보호 메커니즘을 포함합니다:

1. 지연된 복제(Delayed Replication)
2. 백업 시스템
3. 변경 로그(Change Log) 관리
4. Point-in-Time Recovery (PITR)

구체적으로 데이터 손실 방지 메커니즘은 다음과 같습니다.

1. 지연된 복제

    ```mermaid
    sequenceDiagram
        participant Master
        participant Buffer as Replication Buffer
        participant Slave

        Master->>Buffer: 1. 변경사항 기록
        Note over Buffer: 2. 설정된 시간만큼 대기
        Buffer->>Slave: 3. 지연 후 복제
    ```

   작동 방식:
    - 슬레이브는 마스터의 변경사항을 즉시 적용하지 않고 일정 시간(예: 1시간) 지연
    - 실수로 데이터가 삭제되었을 때 지연 시간 내에 복구 가능

    ```sql
    -- MySQL에서 복제 지연 설정 예시
    CHANGE MASTER TO MASTER_DELAY = 3600; -- 1시간 지연
    ```

2. 바이너리 로그(Binary Log)

    ```mermaid
    graph LR
        A[트랜잭션 시작] --> B[바이너리 로그 기록]
        B --> C[데이터 변경]
        C --> D[트랜잭션 커밋]
        D --> E[레플리케이션 전송]
    ```

    - 목적: 모든 데이터 변경사항을 시간 순서대로 기록
    - 활용: 특정 시점으로 복구 가능
    - 보관: 일정 기간(예: 7일) 동안 보관

    ```sql
    -- MySQL 바이너리 로그 설정
    SET GLOBAL binlog_retention_period = 604800; -- 7일간 보관
    ```

   단, 디스크 용량 관리에 대한 고려가 필요합니다.
   바이너리 로그는 용량을 많이 차지할 수 있으므로, 백업과 삭제 주기를 정하고 디스크 용량이 부족해지지 않도록 모니터링하는 것이 중요합니다.

    ```sql
    -- 바이너리 로그 정리 (용량 관리)
    PURGE BINARY LOGS BEFORE 'YYYY-MM-DD HH:MM:SS';
    ```

3. Point-in-Time Recovery (PITR)

   전체 백업과 바이너리 로그를 조합하여 특정 시점까지 데이터를 복구하는 방식입니다.

    ```mermaid
    graph TD
        A[전체 백업] --> B[바이너리 로그]
        B --> C[특정 시점 선택]
        C --> D[복구 실행]
    ```

   작동 방식:
    1. 전체 백업본 사용
    2. 바이너리 로그로 특정 시점까지 복구
    3. 원하는 시점 이후의 변경사항은 무시

    ```sql
    -- MySQL에서 특정 시점으로 복구 예시
    RESTORE FROM BACKUP;
    APPLY BINARY LOGS UNTIL '2024-01-20 15:30:00';
    ```

   PITR을 실제로 사용할 때 데이터 무결성을 확보하는 것도 중요합니다.
   예를 들어, 복구 시 중단된 트랜잭션이나 불완전한 데이터 업데이트가 없도록 트랜잭션 일관성을 유지해야 합니다.

   또한, 복구하는 과정에서 데이터가 많을 경우 복구 시간(RTO)이 길어질 수 있으므로, 복구 속도를 고려하여 백업 주기를 설정하는 것이 좋습니다.

실제 상황에서의 데이터 손실 대응은 다음과 같이 이뤄질 수 있습니다:

1. 시나리오 1: 실수로 데이터 삭제

   지연된 복제를 통해 실수로 데이터가 삭제된 경우 복구할 수 있습니다.

    ```mermaid
    sequenceDiagram
        participant Admin
        participant Master
        participant Slave

        Admin->>Master: DELETE 실수 발생
        Note over Master,Slave: 복제 지연 시간 내
        Admin->>Master: 복제 중지 명령
        Admin->>Slave: 데이터 복구
        Admin->>Master: Slave에서 데이터 복사
    ```

    - 슬레이브를 마스터로 승격하는 방식 (Failover 방식):
      실수로 삭제된 데이터가 슬레이브에 남아 있다면, 슬레이브를 새로운 마스터로 승격하여 복구할 수 있습니다.

        1. 데이터 삭제 확인

           마스터에서 실수로 데이터를 삭제했음을 확인합니다.
           삭제된 시점과 데이터를 기록해 두어야 나중에 복구 시 필요한 정보를 확보할 수 있습니다.

        2. 복제 중지 명령 실행 (슬레이브에서 삭제 반영 방지)
           지연된 복제 상태에서 데이터가 실수로 삭제된 경우 복제를 중지하지 않으면, 지연된 슬레이브에도 삭제가 반영될 수 있습니다.

            ```sql
            -- 복제 중지 명령
            STOP SLAVE;
            ```

        3. 슬레이브 상태 확인

           슬레이브가 어느 시점까지 마스터의 변경 사항을 반영했는지 확인합니다.
           이를 통해 슬레이브에 복구할 수 있는 데이터가 있는지 확인해야 합니다.

            ```sql
            -- 슬레이브 상태 확인
            SHOW SLAVE STATUS;
            ```

           이를 통해 슬레이브가 마스터의 변경 사항을 얼마나 반영했는지 확인하고, 복제 지연이 올바르게 동작했는지를 점검해야 합니다.

        4. 슬레이브를 새로운 마스터로 승격: 슬레이브에서 데이터를 직접 복구하는 대신, 슬레이브를 새로운 마스터로 승격시킵니다.

            ```sql
            RESET MASTER;
            ```

        5. 애플리케이션 연결 전환:
           애플리케이션이 새로운 마스터로 정상적으로 작동하도록 설정을 업데이트합니다.
           로드밸런서를 사용하거나 직접 DB 연결 설정을 변경하여 새로운 마스터로 연결되도록 합니다.

        6. 기존 마스터를 슬레이브로 전환: 기존 마스터 DB가 복구된 후에는 기존 마스터를 새로운 마스터의 슬레이브로 전환할 수 있습니다.

            ```sql
            CHANGE MASTER TO MASTER_HOST='new_master_ip', MASTER_LOG_FILE='new_master_log_file', MASTER_LOG_POS=new_master_log_position;
            START SLAVE;
            ```

    - 복제 시스템을 사용한 복구:

        1. 데이터 삭제 확인

           마스터에서 실수로 데이터를 삭제했음을 확인합니다.
           삭제된 시점과 데이터를 기록해 두어야 나중에 복구 시 필요한 정보를 확보할 수 있습니다.

        2. 복제 중지 명령 실행 (슬레이브에서 삭제 반영 방지)
           지연된 복제 상태에서 데이터가 실수로 삭제된 경우 복제를 중지하지 않으면, 지연된 슬레이브에도 삭제가 반영될 수 있습니다.

            ```sql
            -- 복제 중지 명령
            STOP SLAVE;
            ```

        3. 슬레이브 상태 확인

           슬레이브가 어느 시점까지 마스터의 변경 사항을 반영했는지 확인합니다.
           이를 통해 슬레이브에 복구할 수 있는 데이터가 있는지 확인해야 합니다.

            ```sql
            -- 슬레이브 상태 확인
            SHOW SLAVE STATUS;
            ```

           이를 통해 슬레이브가 마스터의 변경 사항을 얼마나 반영했는지 확인하고, 복제 지연이 올바르게 동작했는지를 점검해야 합니다.

        4. 슬레이브의 데이터 덤프: 슬레이브에서 데이터를 mysqldump와 같은 도구를 사용하여 전체 데이터를 덤프합니다.

            ```sql
            mysqldump -u root -p --all-databases > slave_backup.sql
            ```

        5. 마스터에서 데이터 복구: 덤프된 데이터를 마스터에 복구합니다. 이 방식은 모든 데이터를 일관되게 복구할 수 있어 쿼리를 수동으로 추출하는 방식보다 안전합니다.

            ```sql
            mysql -u root -p < slave_backup.sql
            ```

        6. 복제 재설정: 복구가 완료되면, 복제를 다시 설정하고 슬레이브로부터 최신 상태의 데이터를 마스터로 복제할 수 있도록 합니다.

2. 시나리오 2: 마스터 DB 장애

    ```mermaid
    sequenceDiagram
        participant Master
        participant Slave
        participant Monitor

        Monitor->>Master: 장애 감지
        Monitor->>Slave: Promote to Master
        Note over Slave: 새로운 마스터로 승격
        Note over Master: 복구 후 Slave로 전환
    ```

    1. 슬레이브를 새로운 마스터로 승격 (Failover)

       슬레이브를 새로운 마스터로 승격시키기 위해, 필요한 설정을 변경합니다.
       이 과정에서 GTID(Globally Unique Transaction ID)가 설정되어 있으면, 데이터 불일치 문제를 최소화할 수 있습니다.

        ```sql
        -- 복제 중지 명령
        STOP SLAVE;
        -- 슬레이브의 새로운 로그 파일을 생성하고, 현재 슬레이브가 새로운 마스터로 동작하도록 설정합니다.
        RESET MASTER;
        ```

    2. 슬레이브의 데이터 일관성 확인

       슬레이브를 마스터로 승격한 후, 데이터가 일관성 있는지 확인합니다.
       이를 위해 SHOW SLAVE STATUS를 통해 확인하고, 장애 시점 전까지 모든 트랜잭션이 제대로 반영되었는지 확인합니다.

        ```sql
        -- 슬레이브 상태 및 데이터 일관성 확인
        SHOW SLAVE STATUS;
        ```

        ```log
        # 슬레이브가 마스터로부터 얼마나 뒤처져 있는지를 초 단위로 나타냅니다.
        # 0에 가까울수록 슬레이브가 마스터와 거의 일치함을 의미합니다.
        Seconds_Behind_Master: 0

        # 두 필드 모두 Yes여야 슬레이브가 정상적으로 복제 작업을 수행하고 있다는 것을 의미합니다.
        # 만약 하나라도 No인 경우, 복제가 중단되었거나 장애가 발생한 것입니다.
        Slave_IO_Running: Yes
        Slave_SQL_Running: Yes

        # 슬레이브의 복제가 정상적으로 진행되고 있다면, Master_Log_File와 Relay_Master_Log_File이 동일하거나 근접해야 합니다.
        # 두 파일 간의 차이가 크다면 슬레이브가 복제 작업을 늦게 처리하고 있다는 의미일 수 있습니다.
        # 마스터에서 현재 슬레이브가 읽고 있는 마스터의 바이너리 로그 파일입니다.
        Master_Log_File: mysql-bin.000123
        # 슬레이브에서 현재 처리 중인 마스터의 바이너리 로그 파일입니다.
        Relay_Master_Log_File: mysql-bin.000123

        # 아래 두 값이 가까울수록 슬레이브가 마스터의 바이너리 로그를 빠르게 처리하고 있음을 나타냅니다.
        # 두 값이 많이 차이가 난다면 슬레이브의 SQL 실행이 느리거나 병목이 발생했을 가능성이 있습니다.
        # 슬레이브가 마스터에서 바이너리 로그를 읽어온 로그 위치입니다.
        Read_Master_Log_Pos: 345678
        # 슬레이브가 받은 바이너리 로그 중에서 현재 SQL로 실행된 로그 위치입니다.
        Exec_Master_Log_Pos: 345678

        # Last_Errno가 0이고 Last_Error가 비어 있어야 슬레이브 복제가 정상적으로 진행되었음을 의미합니다.
        # 오류가 발생한 경우 해당 번호와 메시지를 참고하여 문제를 해결해야 합니다.
        Last_Errno: 0
        Last_Error:

        # 슬레이브가 현재 읽고 있는 릴레이 로그 파일 및 로그 위치를 나타냅니다.
        # 릴레이 로그는 마스터에서 받은 바이너리 로그를 슬레이브에서 재생하는 용도로 사용됩니다.
        Relay_Log_File: relay-bin.000456
        Relay_Log_Pos: 789012
        ```

    3. 애플리케이션 연결 전환

       슬레이브가 새로운 마스터로 승격되었으므로, 애플리케이션을 새로운 마스터로 연결합니다.
       애플리케이션의 DB 연결 설정을 변경하거나, 로드밸런서를 사용하는 경우 로드밸런서를 통해 새로운 마스터로 연결 경로를 변경합니다.

    4. 장애 복구 후 기존 마스터를 슬레이브로 전환: 마스터 장애 복구가 완료되면, 기존 마스터를 슬레이브로 재구성합니다.

        ```sql
        -- 기존 마스터를 슬레이브로 설정
        CHANGE MASTER TO MASTER_HOST='새로운 마스터 주소', MASTER_USER='replication_user', MASTER_PASSWORD='password', MASTER_LOG_FILE='마스터 로그 파일', MASTER_LOG_POS=마스터 로그 위치;
        START SLAVE;
        ```

    5. 모니터링 및 복제 상태 확인

       모든 설정이 완료된 후, 복제 상태와 데이터 일관성을 다시 한 번 확인하고 시스템을 정상적으로 모니터링합니다.

        ```sql
        -- 복제 상태 확인
        SHOW SLAVE STATUS;
        ```

고가용성 보장을 위한 모범 사례는 다음과 같습니다.

1. 다중 레플리케이션

    ```plaintext
    Master  -> Slave1 (즉시 복제)
            -> Slave2 (1시간 지연)
            -> Slave3 (24시간 지연)
    ```

2. 정기적인 백업
    - 전체 백업(Full Backup)
    - 증분 백업(Incremental Backup)
    - 차등 백업(Differential Backup)

3. 모니터링 및 자동화

    ```python
    def monitor_replication_lag():
        if get_replication_lag() > threshold:
            alert_admin()
            if is_critical():
                pause_replication()
    ```

4. 복구 계획 및 테스트
    - 정기적인 복구 테스트 수행
    - 복구 시간 목표(RTO) 및 복구 시점 목표(RPO) 설정

```sql
-- 레플리케이션 상태 확인
SHOW
SLAVE STATUS;

-- 복제 중지
STOP
SLAVE;

-- 특정 시점으로 복구
-- 1. 바이너리 로그 위치 확인
SHOW
BINARY LOGS;

-- 2. 특정 시점까지 복구
CHANGE
MASTER TO MASTER_LOG_FILE='mysql-bin.000123',
                 MASTER_LOG_POS=456789;

-- 3. 복제 재시작
START
SLAVE;
```

권장 사항:

1. 레플리케이션 설정
    - 최소 하나의 지연된 복제본 유지
    - 바이너리 로그 충분한 기간 보관

2. 모니터링
    - 복제 지연 모니터링
    - 데이터 불일치 감지

3. 백업 전략
    - 정기적인 전체 백업
    - 실시간 바이너리 로그 백업

4. 문서화
    - 복구 절차 문서화
    - 긴급 연락망 유지

<!-- /curriculum-chunk -->

### PDO가 데이터베이스에서 데이터 가져오는 원리

#### 원문: PDO가 데이터베이스에서 데이터 가져오는 원리

<!-- curriculum-chunk: sha256=0af0591a647d746228909c6db6be849f7ba928e65d70942cc12a1a18e414f00d major=database-storage-search-nosql mid=DB 접근과 저장소 운영 sub=PDO가 데이터베이스에서 데이터 가져오는 원리 sources=source/interview_questions.md:9216-9329, source/interviews.md:9164-9277 -->

> Source: `source/interview_questions.md:9216-9329`
> Classification reason: database/storage
> Duplicate source aliases: `source/interview_questions.md:9216-9329, source/interviews.md:9164-9277`

##### PDO가 데이터베이스에서 데이터 가져오는 원리

`PDO`가 MySQL과 같은 데이터베이스와 통신할 때, 네트워크를 통해 데이터를 주고받는 방식은 내부적으로 MySQL 프로토콜에 의존합니다.
특히 데이터베이스에서 많은 양의 데이터를 가져올 때는 이를 효율적으로 처리하기 위해 데이터가 청크(chunk) 단위로 전송되며, PDO는 이를 관리하고 파싱하는 역할을 수행합니다.

1. MySQL 서버와의 통신은 MySQL 프로토콜을 통해 이루어지며, SQL 쿼리 결과는 청크 단위로 전송됩니다.
2. PDO는 네트워크를 통해 전송된 청크 데이터를 수신하고 파싱하여 PHP에서 사용할 수 있는 데이터 구조(배열, 객체)로 변환합니다.
3. 대량 데이터 처리를 최적화하려면 페이징이나 스트리밍 방식을 사용하여 메모리 사용을 줄일 수 있습니다.
4. MySQL 패킷과 청크 처리 과정은 자동으로 PDO가 관리하므로, 개발자는 주로 메모리 최적화와 결과셋 처리 전략에 집중할 수 있습니다.

PDO는 내부적으로 MySQL 프로토콜을 사용하여 데이터를 주고받으며, 이를 파싱하고 관리하는 역할을 효율적으로 수행합니다.

- MySQL 프로토콜

  `mysql:host=localhost;dbname=testdb`와 같은 방식으로 PDO가 MySQL에 연결을 설정하면,
  내부적으로 MySQL 프로토콜을 사용하여 통신이 이루어집니다.

  MySQL 클라이언트와 서버 간의 통신은 다음 단계를 거칩니다:

    1. 요청 전송: 애플리케이션에서 SQL 쿼리를 작성하고, 이를 PDO 객체를 통해 MySQL 서버에 전송합니다. 이때, MySQL 클라이언트는 SQL 쿼리를 패킷(packet)으로 변환하여 전송합니다.
    2. MySQL 서버에서 쿼리 실행: MySQL 서버는 쿼리를 실행하고, 그 결과를 결과셋(result set) 형태로 클라이언트에 반환합니다. 이때, 결과가 매우 크면 MySQL은 이를 청크(chunk) 단위로 나누어 전송합니다.
    3. 결과 반환: 클라이언트(PHP 애플리케이션)는 MySQL 서버로부터 청크 단위로 데이터를 수신하여 처리합니다.

- PDO의 역할

  `PDO`는 MySQL과의 통신에서 추상화 레이어 역할을 합니다. 다음은 PDO가 수행하는 주요 작업입니다:

    - 데이터베이스 연결 관리: PDO는 MySQL 서버와의 연결을 관리하고, 네트워크를 통해 주고받는 데이터를 처리합니다.
    - SQL 쿼리 전송: SQL 쿼리를 MySQL 서버에 전송하며, 이 과정에서 SQL 명령어를 네트워크 프로토콜 패킷으로 변환하여 서버에 전달합니다.
    - 결과셋 수신 및 파싱: MySQL 서버에서 반환된 청크 데이터를 수신하고, 이를 파싱하여 PHP에서 사용할 수 있는 형태로 변환합니다. PDO는 이 과정을 자동으로 처리하며, 결과는 배열, 객체 등의 형태로 반환됩니다.

데이터 청크 전송과 수신 과정은 다음과 같습니다.

1. MySQL 데이터 전송

   MySQL 서버는 쿼리 결과를 반환할 때 단일 패킷으로 데이터를 전송하는 것이 아니라, 청크(chunk) 단위로 데이터를 나누어 전송합니다.
   이는 데이터의 크기가 클 경우 네트워크 효율성을 높이기 위한 메커니즘입니다.

    - 패킷 단위:

      MySQL은 클라이언트로 데이터를 전송할 때, 네트워크 패킷의 크기를 제한하여 데이터를 청크 단위로 나눕니다.
      일반적으로 한 패킷은 최대 16MB 크기일 수 있으며, 데이터가 그 이상일 경우 여러 패킷으로 나누어 전송됩니다.

    - 청크 처리:

      PDO는 이러한 청크를 연속적으로 수신하며, 데이터를 하나의 결과셋으로 합칩니다.

      예를 들어, 100MB 크기의 데이터를 가져오는 경우 MySQL은 이를 여러 청크로 나누어 전송하고, PDO는 이를 차례로 수신하여 파싱합니다.

2. PDO의 청크 수신 및 파싱

   PDO는 MySQL에서 전송된 청크 데이터를 수신한 후 PHP 데이터 구조로 변환하는 작업을 합니다.

    - 결과셋 처리:

      쿼리 결과는 MySQL 서버에서 전송된 바이너리 포맷의 데이터를 기반으로 합니다.
      PDO는 이 데이터를 수신하여 PHP에서 사용할 수 있는 배열 또는 객체 형태로 변환합니다.
      이는 `fetch()`나 `fetchAll()` 등의 메서드를 사용할 때 내부적으로 수행됩니다.

    - 메모리 관리:

      PDO는 대량의 데이터를 한 번에 가져올 수 있지만, 이로 인해 메모리 사용이 급증할 수 있습니다.
      이를 방지하기 위해 페이징 처리나 스트리밍 방식으로 데이터를 부분적으로 가져오는 전략을 사용할 수 있습니다.

      예를 들어, `PDOStatement::fetch()` 메서드를 사용하여 한 행씩 데이터를 처리하는 방식은 메모리 사용량을 줄일 수 있습니다.

3. PDO 내부에서 데이터 파싱 및 메모리 사용

    1. 파싱 및 데이터 구조 변환

       PDO는 MySQL 서버에서 받은 바이너리 포맷의 결과를 파싱하여 PHP의 데이터 구조로 변환합니다.
       파싱 과정은 MySQL 프로토콜에 따라 수행되며, 주로 다음과 같은 데이터 변환 작업을 포함합니다:

        - 숫자형 데이터: MySQL에서 숫자형 데이터는 바이너리로 전송되며, PDO는 이를 PHP의 숫자형 데이터로 변환합니다.
        - 문자열: 문자열은 UTF-8 또는 기타 인코딩 방식으로 전송되며, PDO는 이를 적절한 PHP 문자열로 변환합니다.
        - NULL 값: MySQL에서 NULL 값은 PDO에서 PHP의 `null`로 변환됩니다.

    2. 메모리 사용 최적화

       PDO는 데이터를 메모리 내에서 처리하므로, 대량의 데이터를 한 번에 가져올 때 메모리 사용량이 급증할 수 있습니다.
       이를 최적화하기 위해 다음과 같은 전략을 사용할 수 있습니다:

        - 페이징 처리: 한 번에 전체 데이터를 가져오는 대신, LIMIT과 OFFSET을 사용하여 데이터베이스에서 결과를 나눠서 가져옵니다.

            ```php
            $stmt = $pdo->prepare("SELECT * FROM large_table LIMIT :limit OFFSET :offset");
            $limit = 1000;
            $offset = 0;
            while (true) {
                $stmt->bindValue(':limit', $limit, PDO::PARAM_INT);
                $stmt->bindValue(':offset', $offset, PDO::PARAM_INT);
                $stmt->execute();
                $rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
                if (empty($rows)) {
                    break;
                }
                // 데이터 처리
                $offset += $limit;
            }
            ```

        - 스트리밍 방식: `fetch()`를 사용하여 한 번에 한 행씩 가져오면, 메모리 사용량을 줄이면서도 데이터를 효율적으로 처리할 수 있습니다.

            ```php
            $stmt = $pdo->prepare("SELECT * FROM large_table");
            $stmt->execute();
            // fetch()를 호출할 때마다, PDO는 MySQL 서버에서 다음 행을 요청하고, 그 행만을 메모리에 로드하여 반환합니다.
            // 이 과정은 결과셋을 모두 한꺼번에 가져와서 메모리에 저장한 후 하나씩 리턴하는 것이 아니라,
            // 필요할 때마다 다음 데이터를 서버에서 가져오는 방식으로 동작합니다.
            while ($row = $stmt->fetch(PDO::FETCH_ASSOC)) {
                // 각 행에 대해 처리
            }
            ```

<!-- /curriculum-chunk -->

## 검색 엔진과 샤딩

### ElasticSearch 쿼리 과정 예시

#### 원문: ElasticSearch 쿼리 과정 예시

<!-- curriculum-chunk: sha256=c6bae0aff340f5cbc1c780c9d70b4357ed2a599dc2d2e7f9e895c267dd582339 major=database-storage-search-nosql mid=검색 엔진과 샤딩 sub=ElasticSearch 쿼리 과정 예시 sources=source/interview_questions.md:7606-7755, source/interviews.md:7554-7703 -->

> Source: `source/interview_questions.md:7606-7755`
> Classification reason: search engine
> Duplicate source aliases: `source/interview_questions.md:7606-7755, source/interviews.md:7554-7703`

##### ElasticSearch 쿼리 과정 예시

> Elasticsearch에서 데이터 노드 A, B, C 3대, `number_of_shards: 3`, `number_of_replicas: 2`로 설정된 경우

위의 설정을 기반으로 샤드 분배 상황은 다음과 같다고 가정합니다. 총 9개의 샤드가 생성됩니다.

- 3개의 프라이머리 샤드 (각각 P1, P2, P3)
- 각 프라이머리 샤드의 2개의 레플리카 샤드 (R1_1, R1_2, R2_1, R2_2, R3_1, R3_2)

각 샤드는 최대한 균등하게 데이터 노드 A, B, C에 분배됩니다. 샤드의 분배는 다음과 같이 구성될 수 있습니다:

- 노드 A: P1, R2_1, R3_2
- 노드 B: P2, R1_1, R3_1
- 노드 C: P3, R1_2, R2_2

이렇게 구성하면 프라이머리 샤드와 해당 레플리카 샤드가 같은 노드에 위치하지 않도록 보장됩니다.

쿼리 수행 방식은 다음과 같습니다.

- 쓰기 쿼리 (Indexing)

  쓰기 작업(인덱싱)은 프라이머리 샤드에서 먼저 수행된 후, 해당 레플리카 샤드로 복제됩니다.

  예를 들어, 데이터가 프라이머리 샤드 P1에 저장될 경우, 노드 A에 있는 P1에 쓰기 작업이 수행됩니다.
  그 후 노드 B의 R1_1과 노드 C의 R1_2로 복제됩니다.

  쓰기 작업이 프라이머리 샤드에 완료된 후에야 레플리카 샤드로 복제가 진행됩니다.
  프라이머리 샤드가 주도권을 가지고 쓰기 작업을 처리합니다.

- 읽기 쿼리 (Search)

  읽기 작업(검색)의 경우, Elasticsearch는 프라이머리 샤드와 레플리카 샤드 중에서 선택적으로 쿼리를 수행할 수 있습니다.
  이때, Elasticsearch는 프라이머리 샤드와 레플리카 샤드에 대해 부하를 분산시켜 성능을 최적화합니다.

  프라이머리 샤드 또는 레플리카 샤드 모두 동일한 데이터를 가지고 있습니다.
  따라서 Elasticsearch는 프라이머리 샤드와 레플리카 샤드 중 아무 샤드에서나 쿼리를 처리할 수 있습니다.

  예를 들어, P1에 해당하는 데이터를 검색한다고 가정해 봅니다.
  노드 A의 P1, 노드 B의 R1_1, 또는 노드 C의 R1_2 중 아무 노드에서나 쿼리를 처리할 수 있습니다.
  Elasticsearch는 부하를 분산하기 위해 자동으로 결정합니다.

  사용자가 Elasticsearch에 검색 쿼리를 요청하면, 쿼리는 다음과 같은 방식으로 처리됩니다.

    1. 쿼리 전달:

       클라이언트는 마스터 노드나 데이터 노드 중 하나에 쿼리를 보냅니다.
       쿼리를 받은 노드는 전체 인덱스의 프라이머리 샤드와 레플리카 샤드 중 어디서 데이터를 가져올지 결정합니다.

    2. 샤드 선택:

       예를 들어, 쿼리가 P1에 해당하는 데이터를 요청한다고 가정합니다.
       Elasticsearch는 P1(프라이머리 샤드) 또는 R1_1, R1_2(레플리카 샤드) 중 하나를 선택하여 데이터를 검색합니다.

    3. 쿼리 분배:

       여러 샤드에서 동시에 검색 작업을 수행합니다.
       P1, P2, P3에서 각각 검색 쿼리가 실행될 수 있고, 레플리카 샤드에서 처리될 수도 있습니다.
       Elasticsearch는 부하를 분산하여 처리합니다.

    4. 결과 집계: 각 샤드에서 검색된 결과는 하나로 통합되어 클라이언트에게 반환됩니다.

- 장애 발생 시 처리

  만약 특정 프라이머리 샤드가 있는 노드가 다운되면, Elasticsearch는 해당 샤드의 레플리카 샤드 중 하나를 프라이머리 샤드로 승격하여 가용성을 유지합니다.

  예를 들어, 노드 A의 P1이 있는 노드가 다운되면, 노드 B의 R1_1이나 노드 C의 R1_2가 프라이머리 샤드로 승격됩니다.

  이렇게 하면 읽기 및 쓰기 작업이 중단되지 않고 계속 진행될 수 있습니다.

Elasticsearch에서 코디네이터 노드는 클러스터 내에서 쿼리 분배 및 결과 통합을 담당합니다.
코디네이터 노드는 데이터를 저장하거나 색인하지 않지만, 클러스터의 모든 프라이머리 샤드와 레플리카 샤드의 정보를 알고 있습니다.
따라서 검색 쿼리를 여러 데이터 노드로 분배하고, 그 결과를 다시 클라이언트에 반환하는 역할을 합니다.

코디네이터 노드가 존재하는 경우 쿼리 처리 단계는 다음과 같습니다.

1. 쿼리 수신

   클라이언트가 Elasticsearch에 검색 쿼리를 보내면, 이 쿼리는 코디네이터 노드에 도착합니다.
   이때 코디네이터 노드는 검색할 인덱스와 해당 인덱스에 포함된 샤드의 위치 정보를 알고 있습니다.

2. 쿼리 분배

   코디네이터 노드는 프라이머리 샤드와 레플리카 샤드 중에서 해당 쿼리를 처리할 수 있는 샤드를 선택합니다.

   예를 들어, 프라이머리 샤드(P1, P2, P3)와 레플리카 샤드(R1_1, R2_1, R3_1)가 노드 A, B, C에 각각 분산되어 있는 경우, 코디네이터 노드는 모든 프라이머리 또는 레플리카 샤드에 쿼리를 분배합니다.

3. 샤드에서 검색 실행

   각 데이터 노드에서 해당 쿼리를 처리합니다.

   예를 들어, 노드 A에서 P1이, 노드 B에서 R2_1이, 노드 C에서 R3_1이 쿼리를 수행할 수 있습니다. 코디네이터는 이러한 노드들에게 검색 작업을 분배하여 병렬로 처리하게 합니다.

   검색 작업은 각각의 샤드에서 이루어지므로 데이터가 분산된 여러 샤드에서 동시에 처리됩니다.

4. 결과 수집

   각 노드에서 수행된 검색 결과는 코디네이터 노드로 다시 전송됩니다.
   코디네이터 노드는 여러 노드에서 반환된 검색 결과를 병합합니다.

   예를 들어, 각 샤드가 반환한 검색 결과가 다수의 문서일 수 있으며, 이 결과를 통합하여 정렬합니다.

5. 결과 정렬 및 집계

   코디네이터는 모든 샤드에서 가져온 결과를 집계합니다.
   만약 쿼리가 집계(aggregation) 작업을 포함하고 있다면, 이를 수행합니다.
   Elasticsearch는 스코어링(relevance score)을 사용해 문서를 정렬할 수도 있고, 집계 결과(예: 평균, 합계)를 계산할 수도 있습니다.

6. 코디네이터 노드는 최종적으로 병합된 결과를 클라이언트에게 반환합니다.

코디네이터 노드가 없는 경우 직접 데이터 노드로 쿼리 수행할 수 있습니다.
이 방식은 클라이언트가 여러 데이터 노드로 직접 쿼리를 보내는 방식으로 작동합니다.
각 데이터 노드는 클러스터 내 샤드의 위치 정보를 알고 있어 적절한 샤드가 있는 노드로 쿼리를 라우팅할 수 있습니다.

1. 노드 목록 제공 및 쿼리 전송

    - 클라이언트는 Elasticsearch 클러스터의 여러 데이터 노드의 주소 목록을 알고 있습니다.
        - 이 주소는 각 데이터 노드의 IP 주소나 호스트 이름으로 구성됩니다.
        - 클라이언트는 쿼리를 임의의 데이터 노드(예: 노드 A)에 전송합니다.

    - 데이터 노드 A는 클러스터 상태와 샤드 위치 정보를 가지고 있습니다.

      이 정보는 마스터 노드가 관리하며, 모든 데이터 노드에 메타데이터로 배포됩니다.
      이를 통해 각 데이터 노드는 어떤 노드에 어떤 프라이머리 샤드 또는 레플리카 샤드가 있는지를 알고 있습니다.

2. 해당 노드에 샤드가 없을 경우

   클라이언트가 보낸 쿼리에 해당하는 샤드가 노드 A에 없을 경우, 노드 A는 해당 샤드가 어느 노드에 위치하는지 알고 있습니다.
   노드 A는 쿼리를 적절한 노드로 라우팅합니다.

   예를 들어, 클라이언트가 P1에 해당하는 데이터를 쿼리했지만, 노드 A에 P1 샤드가 없다면, 노드 A는 노드 B에 P1이 있다는 것을 알고 쿼리를 노드 B로 라우팅합니다.

   쿼리 라우팅은 노드 A가 직접 처리하며, 노드 간 통신은 Elasticsearch가 내부적으로 관리합니다.

3. 쿼리 분배 및 검색 수행

   쿼리를 받은 노드 B는 자신이 보유한 샤드(P1)에 대해 쿼리를 수행합니다.
   노드 B는 자신이 관리하는 샤드에서 검색을 수행하고, 결과를 준비합니다.

   만약 프라이머리 샤드가 바쁘거나 레플리카 샤드를 활용할 수 있는 경우, 레플리카 샤드에서 쿼리를 수행할 수도 있습니다.

   예를 들어, 노드 B에 P1 프라이머리 샤드가 있지만, 노드 C에 있는 P1 레플리카 샤드가 부하가 적다면, 쿼리가 노드 C에서 처리될 수 있습니다.

4. 결과 통합

   클라이언트는 쿼리를 보낸 노드(노드 A)를 통해 결과를 수신합니다.
    - 쿼리를 보낸 노드가 결과를 받으면 직접 클라이언트로 반환합니다.
    - 클라이언트는 여러 데이터 노드로부터 직접 결과를 받지 않습니다. 클라이언트는 쿼리를 보낸 노드가 모든 처리 과정을 중계하여, 최종 결과를 반환합니다.

   클라이언트는 단일 노드로부터 병합된 결과를 받습니다.
    - 클라이언트는 검색 결과를 받아 결과를 정렬하거나, 집계 작업을 처리할 수 있습니다.
    - 대량의 데이터를 처리할 때는 scroll API나 search_after 기능을 통해 페이지별로 데이터를 가져올 수 있습니다.

<!-- /curriculum-chunk -->

### ElasticSearch에서 OOM(Out of Memory) 문제

#### 원문: ElasticSearch에서 OOM(Out of Memory) 문제

<!-- curriculum-chunk: sha256=6904e1c9711bd9b04ede509af231ada6b9ae5dc0a2d29bcdc0ad78ee9146055c major=database-storage-search-nosql mid=검색 엔진과 샤딩 sub=ElasticSearch에서 OOM(Out of Memory) 문제 sources=source/interview_questions.md:7436-7500, source/interviews.md:7384-7448 -->

> Source: `source/interview_questions.md:7436-7500`
> Classification reason: search engine
> Duplicate source aliases: `source/interview_questions.md:7436-7500, source/interviews.md:7384-7448`

##### ElasticSearch에서 OOM(Out of Memory) 문제

OOM(Out of Memory) 문제는 프로그램이 실행 중에 사용할 수 있는 메모리가 부족해질 때 발생합니다. 이는 주로 다음과 같은 이유로 발생할 수 있습니다:

- 메모리 누수: 프로그램에서 사용한 메모리가 제대로 해제되지 않는 경우.
- 과도한 데이터 처리: 프로그램이 처리해야 할 데이터 양이 메모리 용량을 초과하는 경우.
- Garbage Collection 문제: Java 기반의 애플리케이션(예: Elasticsearch)에서 GC(Garbage Collection)가 제때 실행되지 않으면, 메모리가 효율적으로 회수되지 않아 OOM 문제가 발생할 수 있습니다.

Elasticsearch에서는 대량의 데이터를 처리하다 보면 heap 메모리가 꽉 차는 경우가 있습니다.
특히 실시간으로 대규모 검색과 색인 작업이 이루어지는 Hot 노드에서 이러한 문제가 빈번하게 발생할 수 있습니다.

Elasticsearch에서 OOM 문제는 메모리가 부족하여 검색 또는 색인 작업이 실패할 때 발생합니다.
이를 해결하기 위해서는 여러 가지 방법이 있습니다.

1. Heap 메모리 크기 조정

   Elasticsearch는 JVM(Java Virtual Machine) 기반으로 동작하므로,
   JVM heap 메모리 크기를 적절하게 설정하는 것이 중요합니다.
   Elasticsearch의 heap 크기는 다음 설정을 통해 조정할 수 있습니다.

    - 설정 파일: `jvm.options`
    - 권장 사항: 전체 시스템 메모리의 50% 정도를 Elasticsearch의 heap으로 할당하고, 최대 32GB를 넘지 않도록 설정하는 것이 좋습니다.

        ```bash
        -Xms16g
        -Xmx16g
        ```

        - Xms와 Xmx 옵션은 JVM의 초기 및 최대 heap 크기를 설정합니다.
        - 너무 작은 heap 크기를 설정하면 OOM 문제가 발생하고, 너무 큰 크기를 설정하면 Garbage Collection이 느려질 수 있습니다.

2. Garbage Collection 최적화

   Garbage Collection(GC)은 Elasticsearch에서 메모리를 정리하는 중요한 과정입니다.
   GC 설정을 통해 메모리 사용 효율성을 높일 수 있습니다.

   JVM에서 GC 로깅을 활성화하면 메모리 누수를 추적하거나, 불필요한 메모리 할당 문제를 확인할 수 있습니다.

    ```bash
    -XX:+PrintGCDetails
    -XX:+PrintGCDateStamps
    -XX:+UseGCLogFileRotation
    ```

   필요한 경우 Garbage Collector를 G1GC 또는 CMS(Concurrent Mark-Sweep)로 변경하여 GC 동작 방식을 최적화할 수 있습니다.

3. 인덱스 크기 최적화

   메모리 사용을 줄이기 위해서는 인덱스 크기를 적절하게 관리하는 것이 중요합니다.
   너무 많은 샤드와 인덱스를 유지하면 메모리 사용량이 급격히 증가할 수 있습니다.

    - 샤드 크기 관리:
      샤드 크기를 적절하게 조정하여 메모리 사용을 줄일 수 있습니다.
      일반적으로 30GB~50GB의 샤드 크기가 권장됩니다.

    - 복제본 설정 최적화:
      필요 이상의 복제본을 생성하면 메모리 사용량이 급증하므로, 복제본 수를 적절히 조정합니다.

4. 노드 간 데이터 분산

   Hot/Warm 아키텍처에서 Hot 노드와 Warm 노드 간에 데이터를 적절하게 분산시켜 메모리 부담을 줄일 수 있습니다.

    - Index Lifecycle Management(ILM) 정책을 사용하여 오래된 데이터를 Warm 또는 Cold 노드로 자동으로 이동시켜 메모리와 CPU 자원을 최적화합니다.
    - Hot 노드에서는 자주 액세스되는 데이터만 유지하고, 오래된 데이터는 자동으로 Warm 노드로 이동하게끔 설정할 수 있습니다.

<!-- /curriculum-chunk -->

### Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명

#### 원문: Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명

<!-- curriculum-chunk: sha256=70bee7f4dfca1a539eb44e380e43934d3c32312b1c012565f31240f9879dc82c major=database-storage-search-nosql mid=검색 엔진과 샤딩 sub=Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명 sources=source/interview_questions.md:7425-7435, source/interviews.md:7373-7383 -->

> Source: `source/interview_questions.md:7425-7435`
> Classification reason: search engine
> Duplicate source aliases: `source/interview_questions.md:7425-7435, source/interviews.md:7373-7383`

##### Elasticsearch, OOM 문제, Hot/Warm 아키텍처 설명

Elasticsearch는 대량의 데이터를 실시간으로 검색하고 분석할 수 있는 분산형 검색 엔진입니다.
주로 로그나 데이터를 빠르게 검색하고 저장할 수 있는 NoSQL 데이터베이스로 사용됩니다.
Elasticsearch는 분산 아키텍처를 기반으로 다수의 노드로 구성된 클러스터 형태로 운영되며,
각 노드는 여러 개의 샤드(shard)와 인덱스(index)로 나뉘어 데이터를 저장하고 처리합니다.

- 노드(Node): Elasticsearch 클러스터에 속한 하나의 서버입니다. 각 노드는 데이터를 저장하고, 검색 쿼리를 처리합니다.
- 샤드(Shard): 인덱스를 쪼개어 분산된 데이터의 작은 부분입니다. 검색 성능을 높이기 위해 데이터를 여러 샤드에 분산하여 저장합니다.
- 인덱스(Index): 데이터를 논리적으로 분류한 단위입니다. 데이터베이스의 테이블과 비슷한 개념으로, 데이터가 검색되고 저장되는 기본 단위입니다.

<!-- /curriculum-chunk -->

### Elasticsearch의 샤드와 레플리카 샤드

#### 원문: Elasticsearch의 샤드와 레플리카 샤드

<!-- curriculum-chunk: sha256=b0379a154e23d32c86878d0850a82946be3db4f865da408f77619b79191d93bf major=database-storage-search-nosql mid=검색 엔진과 샤딩 sub=Elasticsearch의 샤드와 레플리카 샤드 sources=source/interview_questions.md:7538-7605, source/interviews.md:7486-7553 -->

> Source: `source/interview_questions.md:7538-7605`
> Classification reason: search engine
> Duplicate source aliases: `source/interview_questions.md:7538-7605, source/interviews.md:7486-7553`

##### Elasticsearch의 샤드와 레플리카 샤드

- 샤드(Shard):

  Elasticsearch 인덱스는 기본적으로 프라이머리 샤드(Primary Shard)와 레플리카 샤드(Replica Shard)로 나누어집니다.
  하나의 샤드는 인덱스 데이터를 담는 독립적인 데이터 단위입니다.

- 프라이머리 샤드:

  원본 데이터를 저장하는 샤드입니다.
  인덱스에 데이터를 쓸 때 이 프라이머리 샤드에 기록됩니다.

  "number_of_shards: 3" 설정은 인덱스의 전체 데이터를 3개의 프라이머리 샤드에 분산하여 저장하겠다는 의미입니다.

- 레플리카 샤드:

  프라이머리 샤드의 복사본입니다.
  레플리카 샤드는 고가용성을 보장하기 위해 사용되며, 프라이머리 샤드가 장애가 발생할 경우 이를 대체할 수 있습니다.
  또한 검색 쿼리에서 부하 분산 역할도 합니다.

  "number_of_replicas: 1" 설정은 각 프라이머리 샤드마다 1개의 레플리카 샤드를 생성하겠다는 의미입니다.
  프라이머리 샤드가 3개인 경우, 각 프라이머리 샤드에 대해 1개의 레플리카 샤드가 생성되므로, 레플리카 샤드도 3개가 만들어집니다.
  따라서 프라이머리 샤드 3개 + 레플리카 샤드 3개 = 총 6개의 샤드가 존재하게 됩니다.

Elasticsearch 클러스터는 노드들이 모여 하나의 클러스터를 이루고, 각각의 노드는 샤드를 보유합니다.
샤드 분배는 아래의 원칙을 따릅니다.

- 프라이머리 샤드의 분배 원칙

  Elasticsearch는 클러스터에 있는 여러 노드에 프라이머리 샤드를 최대한 고르게 분산시킵니다.
  이로 인해 데이터가 한 노드에 집중되지 않게 하고, 시스템 리소스(CPU, 메모리, 디스크)의 효율적인 사용을 보장합니다.

  예를 들어, 인덱스에 5개의 프라이머리 샤드를 설정하고 3개의 노드가 있다면, 각 노드에 최대한 균등하게 샤드를 분배합니다.
  한 노드가 2개의 프라이머리 샤드를 가지고, 나머지 노드는 1개의 프라이머리 샤드를 가지는 방식입니다.

    - 샤드 이동 원칙 (Rebalancing)

      새로운 노드를 클러스터에 추가하거나 기존 노드를 제거하면 Elasticsearch는 자동으로 샤드를 재분배합니다.
      새로운 노드가 추가되면 기존에 과부하가 걸릴 수 있는 노드에서 샤드를 이동시켜 부하를 균등화하려 합니다.

    - 샤드의 성능 고려

      샤드 하나의 크기는 성능에 영향을 미칩니다.
      일반적으로 하나의 샤드 크기는 30GB~50GB로 설정하는 것이 적당합니다.
      너무 작은 샤드가 많으면 오버헤드가 커지고, 너무 큰 샤드가 있으면 메모리 및 CPU 성능이 저하될 수 있습니다.

- 레플리카 샤드 분배 원칙

  레플리카 샤드는 프라이머리 샤드의 복사본이므로, 고가용성과 부하 분산을 위해 여러 노드에 위치하게 됩니다.
  레플리카 샤드가 분배되는 원칙은 다음과 같습니다.

    - 프라이머리 샤드와 다른 노드에 위치

      레플리카 샤드는 절대 해당 프라이머리 샤드와 같은 노드에 위치하지 않습니다.
      이는 한 노드가 다운되더라도 해당 노드에 있던 프라이머리 샤드를 레플리카 샤드를 통해 복구할 수 있게 하기 위한 것입니다.
      만약 프라이머리와 레플리카가 같은 노드에 있다면 그 노드가 다운될 경우 데이터를 잃게 될 가능성이 있기 때문입니다.

    - 레플리카 샤드의 부하 분산

      Elasticsearch는 검색 쿼리를 프라이머리 샤드뿐만 아니라 레플리카 샤드에서도 처리할 수 있습니다.
      이렇게 함으로써 여러 샤드에 부하를 분산하여 성능을 향상시킵니다.
      따라서 레플리카 샤드는 최대한 고르게 여러 노드에 분산됩니다.

    - 레플리카 샤드의 재분배

      노드가 추가되거나 제거되면, Elasticsearch는 레플리카 샤드를 적절히 이동시켜 각 노드가 균형 있게 부하를 분산하도록 조정합니다.
      마찬가지로, 한 노드가 다운되면 Elasticsearch는 자동으로 레플리카 샤드를 이용해 프라이머리 샤드를 다시 복구합니다.

<!-- /curriculum-chunk -->

### [ElasticSearch Hot/Warm 아키텍처](https://www.elastic.co/kr/blog/elasticsearch-data-lifecycle-management-with-data-tiers)

#### 원문: [ElasticSearch Hot/Warm 아키텍처](https://www.elastic.co/kr/blog/elasticsearch-data-lifecycle-management-with-data-tiers)

<!-- curriculum-chunk: sha256=f89728d54bec0e1f9ff887b532d59a401ac1dd5e56b3cf055ec32e062e4c612a major=database-storage-search-nosql mid=검색 엔진과 샤딩 sub=[ElasticSearch Hot/Warm 아키텍처](https://www.elastic.co/kr/blog/elasticsearch-data-lifecycle-management-with-data-tiers) sources=source/interview_questions.md:7501-7537, source/interviews.md:7449-7485 -->

> Source: `source/interview_questions.md:7501-7537`
> Classification reason: search engine
> Duplicate source aliases: `source/interview_questions.md:7501-7537, source/interviews.md:7449-7485`

##### [ElasticSearch Hot/Warm 아키텍처](https://www.elastic.co/kr/blog/elasticsearch-data-lifecycle-management-with-data-tiers)

Hot/Warm 아키텍처는 Elasticsearch에서 데이터 처리 성능을 최적화하기 위해 사용하는 데이터 분류 방식입니다.
이 구조는 다양한 데이터 사용 패턴에 맞추어 서버 자원을 효율적으로 관리할 수 있게 도와줍니다.

- Hot 노드:

  실시간으로 자주 액세스되는 데이터를 처리합니다.
  최신 데이터를 색인하고 검색하는 역할을 하며, 성능이 중요한 작업이기 때문에 SSD와 같은 고성능 스토리지와 많은 메모리가 필요합니다.
  하지만 많은 메모리를 사용하기 때문에 OOM 문제가 종종 발생할 수 있습니다.

- Warm 노드:

  Hot 노드보다 덜 자주 조회되는 데이터(오래된 로그 등)를 처리합니다.
  성능이 덜 중요하므로 상대적으로 저렴한 스토리지(예: HDD)를 사용해도 됩니다.
  Warm 노드는 Hot 노드처럼 메모리 문제가 심각하지 않지만, 디스크 I/O 성능이 중요할 수 있습니다.

이 구조의 목적은 Hot 노드에서 자주 사용되는 최신 데이터를 빠르게 처리하고, Warm 노드에서 덜 빈번하게 조회되는 데이터를 처리하여 시스템 자원의 효율을 극대화하는 것입니다.

```plaintext
+-----------------+            +-----------------+
|   Hot Node 1    |   <---->   |   Warm Node 1    |
+-----------------+            +-----------------+
|   Hot Node 2    |            |   Warm Node 2    |
+-----------------+            +-----------------+
        |                             |
        +---------+     +-------------+
                  |     |
           +-----------------------+
           |  Cold Node (Optional) |
           +-----------------------+
```

- Hot Node: 주로 실시간 데이터 처리.
- Warm Node: 오래된 데이터 처리.
- Cold Node(Optional): 거의 접근하지 않는 데이터를 저장, 주로 매우 오래된 로그 데이터.

<!-- /curriculum-chunk -->

## 검색/NoSQL 저장소

### 카우치베이스

#### 원문: 카우치베이스

<!-- curriculum-chunk: sha256=35e4fc51b5487e51496e077e81c1666a5a290bb415bee95b1998c1fe3912b790 major=database-storage-search-nosql mid=검색/NoSQL 저장소 sub=카우치베이스 sources=source/interview_questions.md:7816-7975, source/interviews.md:7764-7923 -->

> Source: `source/interview_questions.md:7816-7975`
> Classification reason: NoSQL storage
> Duplicate source aliases: `source/interview_questions.md:7816-7975, source/interviews.md:7764-7923`

##### 카우치베이스

Couchbase는 NoSQL 분산형 데이터베이스로, 고성능과 확장성을 목표로 설계된 시스템입니다.
이를 통해 대규모 트래픽을 처리하는 애플리케이션에서 많이 사용됩니다.
메모리 중심 구조를 기반으로 하여 빠른 응답 속도를 제공하며, 다양한 데이터 모델링 방식과 강력한 관리 도구를 갖추고 있습니다.

1. 하이브리드 구조: 키-값과 문서 데이터베이스

   Couchbase는 키-값(Key-Value) 스토어와 문서 지향형(Document-Oriented) 데이터베이스의 기능을 결합한 하이브리드 NoSQL 데이터베이스입니다.
   내부적으로 JSON 문서 형식을 사용하여 데이터를 저장하며, 각 문서는 고유한 키(Key)로 식별됩니다.

    - 키-값 스토어: 빠른 읽기/쓰기 성능을 제공하는 캐시 기반의 저장 방식으로, 각 키는 고유하며, 키에 대한 조회는 메모리에서 직접 처리되므로 매우 빠릅니다.
    - 문서 지향 데이터베이스: Couchbase는 JSON 형식의 문서를 저장하며, 데이터가 스키마리스(Schemaless) 방식으로 저장되므로 유연한 구조 변경이 가능합니다. JSON 데이터는 네이티브로 저장되고 쿼리됩니다.

2. 메모리 우선 구조

   Couchbase는 메모리 중심 데이터베이스로, 모든 데이터가 메모리에 우선 저장된 후, 일정 주기나 트래픽 변화에 따라 디스크로 플러시(Flush)됩니다.
   이를 통해 빠른 데이터 읽기 및 쓰기 속도를 보장합니다.

    - 메모리 캐싱: 데이터를 읽고 쓸 때 캐시에서 먼저 처리하며, 캐시에 없을 때만 디스크에서 읽습니다. 또한, 메모리 내 데이터는 백그라운드 플러시 작업을 통해 디스크에 기록되므로 빠른 성능을 유지할 수 있습니다.
    - ejection: 메모리가 부족할 경우에는 Ejection 메커니즘을 사용해 데이터를 디스크로 밀어내고, 더 자주 사용하는 데이터를 메모리에 남겨둡니다.

3. 분산 아키텍처 및 클러스터링

   Couchbase는 분산 시스템으로 설계되어 있으며, 데이터를 클러스터 전체에 분산하여 저장합니다.
   클러스터는 여러 노드(Node)로 구성되며, 각 노드는 독립적으로 데이터를 처리하고 저장합니다.

    - 데이터 파티셔닝:
      Couchbase는 데이터를 vBucket이라는 단위로 나눕니다.
      데이터는 각 vBucket에 할당된 노드에 분배되며, 이 방식으로 클러스터 간에 데이터를 고르게 분산하여 수평 확장성을 제공합니다.

    - 복제:
      각 데이터는 하나 이상의 리플리카(Replica)로 복제됩니다.
      이를 통해 노드 장애 시 데이터를 다른 노드에서 복구할 수 있으며, 고가용성을 보장합니다.

    - Failover 및 장애 복구:
      Couchbase는 자동 장애 복구(Failover) 기능을 제공하여, 한 노드에 장애가 발생할 경우 다른 노드에서 복구가 이루어집니다.
      수동 또는 자동 복구 설정이 가능합니다.

4. N1QL (SQL-like 쿼리 언어)

   Couchbase는 N1QL이라는 SQL 기반의 쿼리 언어를 제공합니다.
   N1QL은 SQL과 매우 유사하며, JSON 문서를 대상으로 복잡한 쿼리를 실행할 수 있는 기능을 제공합니다.
   NoSQL임에도 불구하고 SQL처럼 JOIN, GROUP BY, WHERE 등의 연산을 사용할 수 있습니다.

   SQL의 유연성과 JSON의 확장성을 결합하여, 개발자들이 친숙한 SQL 문법을 사용하면서도 NoSQL의 스키마리스 데이터를 다룰 수 있도록 해줍니다.

5. 인덱싱

   Couchbase는 빠른 데이터 조회를 위해 인덱싱 기능을 제공합니다.
   주 인덱스(primary index)와 보조 인덱스(secondary index)를 사용할 수 있습니다.

    - 주 인덱스: 모든 문서에 대해 기본적으로 제공되는 인덱스입니다.
    - 보조 인덱스: 특정 필드를 기반으로 인덱스를 생성할 수 있으며, 이를 통해 효율적으로 데이터를 조회할 수 있습니다.

6. 어드민 콘솔과 관리 도구

   Couchbase는 웹 기반 GUI 어드민 콘솔을 제공합니다. 이 콘솔을 통해 다음과 같은 관리 작업을 쉽게 할 수 있습니다.

    - 클러스터 모니터링: 각 노드의 상태, 데이터 분포, 리소스 사용량, 트래픽 등을 실시간으로 모니터링할 수 있습니다.
    - 데이터 관리: 버킷 생성, 인덱스 관리, 쿼리 실행, 복제본 설정 등을 인터페이스에서 직접 관리할 수 있습니다.
    - 클러스터 확장: 노드 추가, 제거, 업그레이드 등의 클러스터 관리 작업을 웹 콘솔을 통해 쉽게 할 수 있습니다.
    - 백업 및 복원: Couchbase는 데이터를 백업하고 복원할 수 있는 기능도 제공합니다.

7. 라이브 업그레이드 및 무중단 운영

   Couchbase는 라이브 업그레이드를 지원하여, 서비스를 중단하지 않고 실시간으로 노드를 추가하거나 업그레이드할 수 있습니다. 이를 통해 무중단 서비스가 가능합니다.

Couchbase의 장단점:

- 장점
    1. 고성능 메모리 중심 구조:
        - Couchbase는 메모리 우선 아키텍처를 사용하여 매우 빠른 읽기/쓰기 성능을 제공합니다. 이는 실시간 처리가 중요한 애플리케이션에 적합합니다.
        - 데이터를 메모리에 캐싱하고, 백그라운드에서 디스크로 플러시하는 방식으로 설계되어 성능 최적화가 가능합니다.

    2. 수평 확장성 (Horizontal Scalability):
        - Couchbase는 분산형 데이터베이스로 설계되어 있어, 노드를 쉽게 추가하여 클러스터를 확장할 수 있습니다. 클러스터 내의 노드는 데이터를 고르게 분산해 처리합니다.
        - 확장할 때 데이터 리밸런싱이 자동으로 이루어지며, 노드 간 장애가 발생해도 Failover를 통해 자동 복구됩니다.

    3. N1QL 쿼리:
        - Couchbase는 SQL과 유사한 N1QL 쿼리 언어를 지원하므로, 기존의 SQL 경험이 있는 개발자들이 NoSQL 환경에서도 쉽게 적응할 수 있습니다. 이를 통해 복잡한 데이터 조회 및 필터링을 유연하게 처리할 수 있습니다.

    4. 어드민 콘솔 및 관리 기능:
        - Couchbase는 웹 기반 관리 도구를 통해 클러스터 모니터링, 버킷 관리, 인덱스 관리 등 다양한 기능을 제공합니다.
        - 장애 복구, 백업 및 복원, 모니터링 등 관리 작업을 GUI로 쉽게 수행할 수 있습니다.

    5. 고가용성 및 자동 장애 복구:
        - Couchbase는 복제본을 사용해 데이터를 여러 노드에 복제하며, 한 노드에 장애가 발생해도 다른 노드에서 자동으로 데이터를 복구할 수 있습니다.
        - 이는 클러스터의 무중단 운영을 가능하게 하며, 노드 추가/삭제 또는 업그레이드 시 서비스 중단 없이 작업할 수 있습니다.

    6. 노드 추가시 자동 리밸런싱

    7. [ACID](https://www.couchbase.com/blog/acid-properties-couchbase-part-1/)

- 단점:
    1. 메모리 사용량:
        - Couchbase는 메모리 중심 아키텍처이기 때문에, 대규모 데이터 처리 시 메모리 사용량이 매우 큽니다. 메모리가 부족할 경우 성능이 저하될 수 있으며, 메모리 튜닝이 필요할 수 있습니다.
        - 데이터가 메모리에 적재되지 않으면 디스크 액세스가 빈번하게 발생할 수 있어 성능이 저하될 수 있습니다.

    2. 복잡한 설정:
        - Couchbase의 분산 클러스터 설정 및 확장 작업은 다른 NoSQL 솔루션에 비해 비교적 복잡할 수 있습니다.
        - 특히, 대규모 클러스터를 운영할 때 노드 간의 데이터 분배 및 복구 설정에 신경을 써야 합니다.

    3. N1QL 성능 문제:
        - N1QL은 SQL과 유사한 기능을 제공하지만, 매우 복잡한 쿼리에서 성능 저하가 발생할 수 있습니다. 잘못된 인덱스 설정이나 쿼리 최적화가 부족하면 속도가 느려질 수 있습니다.

    4. 거버넌스 및 모니터링의 추가 비용:
        - Couchbase의 많은 기능(특히 고급 모니터링, 백업 및 보안 기능)은 엔터프라이즈 버전에서만 제공되며, 이러한 기능을 사용하기 위해 추가 비용이 발생할 수 있습니다.

Couchbase는 빠른 성능, 대규모 확장성, 고가용성을 요구하는 다양한 분야에서 사용됩니다.

- 실시간 애플리케이션: 온라인 광고 플랫폼, 소셜 네트워크, 실시간 게임 등에서 빠른 데이터 조회와 업데이트가 중요한 경우.
- 분산 캐시: 대규모 웹 애플리케이션에서 메모리 중심 캐시로 사용해 대용량 트래픽을 처리합니다.
- IoT 데이터 처리: 대규모 사물인터넷(IoT) 데이터를 처리하고 관리할 때 분산 데이터베이스로 유용합니다.
- 모바일 백엔드: Couchbase는 오프라인 동기화 기능을 지원하는 Couchbase Lite와 연동할 수 있어, 모바일 애플리케이션의 백엔드로도 자주 사용됩니다.

카우치베이스(Couchbase)가 Erlang을 사용하고 있고, 그로 인해 메모리 사용에 대한 이슈가 있을 수 있다는 말은 일면 맞는 부분도 있지만, 오해의 소지가 있을 수 있습니다.
이를 정확히 이해하려면 Couchbase의 구조와 Erlang의 역할, 그리고 메모리 사용 방식에 대해 자세히 설명하는 것이 필요합니다.

1. Couchbase와 Erlang의 관계

   Couchbase는 분산 데이터베이스 시스템으로, 기본적으로 Couchbase 서버는 여러 프로세스 간의 통신과 분산 처리를 효율적으로 관리하기 위해 Erlang을 사용합니다.
   Erlang은 높은 병렬성과 분산 시스템에 최적화된 언어로, 분산 환경에서의 안정적인 운영을 위해 Couchbase의 클러스터 관리와 노드 간 통신 등을 처리하는 데 사용됩니다.

2. Couchbase의 메모리 사용 특성

   Couchbase는 메모리 사용을 상당히 중시하는 메모리 중심 데이터베이스로 설계되었습니다.
   즉, 데이터를 빠르게 처리하기 위해 대부분의 데이터가 메모리에서 캐싱되고, 필요시 디스크에 플러시(flush)되는 방식으로 동작합니다.
   이러한 설계는 고성능을 목표로 하기 때문에, 메모리를 많이 사용할 수밖에 없는 구조입니다.

Couchbase의 메모리 사용 방식:

1. 캐시 메모리:
   Couchbase는 메모리 우선(memory-first) 구조로 동작하며, 가능한 한 많은 데이터를 메모리에 캐싱합니다.
   이를 통해 읽기와 쓰기 작업에서 높은 성능을 제공합니다.

2. 메모리 풀 관리:
   Couchbase는 데이터를 처리하는 동안 메모리 풀(pool)을 사용해 관리하고, 메모리에서 데이터를 처리한 후 백그라운드로 디스크에 기록하는 구조입니다.
   이는 대규모 데이터를 메모리에 많이 의존하게 만들고, 메모리 부족 시 성능 저하가 발생할 수 있습니다.

3. Erlang의 역할:
   Erlang은 Couchbase의 클러스터 통신과 노드 간의 분산 처리에 사용되며, 노드 간 안정성을 위해 중요한 역할을 합니다.
   Erlang 자체는 가벼운 프로세스를 사용하여 병렬 처리를 효율적으로 수행하지만, Couchbase의 주요 데이터 처리는 C++로 작성된 데이터베이스 엔진이 담당합니다.

Couchbase가 메모리 부담이 크다는 말은, Couchbase의 메모리 중심 설계에서 기인한 것입니다.
Couchbase는 성능을 극대화하기 위해 데이터 캐시를 가능한 한 메모리에 저장하려 하며, 그로 인해 메모리를 많이 사용하게 됩니다.

- 대규모 메모리 캐싱:
  Couchbase는 성능을 높이기 위해 데이터를 메모리에 캐시하고 관리합니다.
  이를 통해 매우 빠른 응답성을 제공하지만, 메모리 부족 상황이 오면 성능 저하가 발생할 수 있습니다.

- 메모리 튜닝:
  Couchbase는 기본적으로 메모리 사용량을 설정할 수 있으며, 이를 적절히 튜닝하여 메모리 사용량을 조절할 수 있습니다.

Erlang은 주로 분산 시스템의 안정성을 보장하고, 클러스터 관리에서 사용되기 때문에 메모리 사용량 자체에 직접적인 영향을 미치는 것은 아닙니다.
다만, Couchbase의 설계 자체가 메모리 중심적이기 때문에, Erlang의 사용 여부와는 별개로 많은 메모리를 사용하는 구조입니다.

- Erlang은 Couchbase에서 클러스터 관리와 노드 간 통신에 사용되며, Couchbase의 분산 처리 시스템의 안정성을 높여줍니다.
- Couchbase의 메모리 사용 부담은 Couchbase가 메모리 중심 구조로 설계되었기 때문에 발생하며, 이는 성능을 높이기 위한 설계 선택입니다.
- Couchbase는 대규모 메모리 캐싱을 통해 빠른 성능을 제공하지만, 메모리 설정을 잘못하거나 메모리가 부족한 환경에서는 성능 저하가 발생할 수 있습니다.

<!-- /curriculum-chunk -->

## 대용량 테이블 운영

### 수억건의 데이터가 존재하는 테이블 처리 방법

#### 원문: 수억건의 데이터가 존재하는 테이블 처리 방법

<!-- curriculum-chunk: sha256=544da3e54cefe9923c6c09d44d83b774d5874f064a659d2381363b9d6c11760b major=database-storage-search-nosql mid=대용량 테이블 운영 sub=수억건의 데이터가 존재하는 테이블 처리 방법 sources=source/interview_questions2.md:1120-1408, source/interviews2.md:1120-1408 -->

> Source: `source/interview_questions2.md:1120-1408`
> Classification reason: large table operation
> Duplicate source aliases: `source/interview_questions2.md:1120-1408, source/interviews2.md:1120-1408`

##### 수억건의 데이터가 존재하는 테이블 처리 방법

###### 문제 상황

- MySQL 5.5버전으로 3억건의 데이터가 저장된 테이블이 존재합니다.
- 하나의 테이블이 너무 커져서 몇 백 기가에 달하는 데이터가 존재합니다.
- 결제 내역을 저장하는 테이블이어서 서비스 중간은 불가능한 상태입니다.
- 또한 DB 버전 업그레이드도 현재로서는 어려운 상황입니다.
- 별도의 reader 서버나 replica 서버 등도 있지만, 결국 하나의 큰 테이블을 어떻게 해야 하는 상황입니다.
- incremental integer id가 있어서 이 아이디가 조회에 유의미하게 사용되고 있습니다. 따라서 데이터가 쌓이는 순서도 중요합니다.
- 조회도 실시간으로 이뤄지고 있습니다. 따라서 파티셔닝을 해도 조회는 일관되게 제공해야 합니다.

이런 상황에서 현재 테이블을 어떻게 해야 할지 한번 논해보세요.

###### 핵심 제약

1. MySQL 5.5의 한계:
    - 최신 MySQL 버전에서 제공하는 `InnoDB`의 native partitioning 및 온라인 작업 기능을 사용할 수 없습니다.
2. 실시간 데이터 일관성:
    - 조회와 삽입 작업 모두 서비스 중단 없이 이루어져야 하며, 모든 데이터가 일관된 인터페이스로 접근 가능해야 합니다.
3. 데이터 크기와 성능 문제:
    - 테이블 크기가 크기 때문에 스캔 속도, 인덱스 유지 비용, 쿼리 응답 시간이 점점 느려질 가능성이 높습니다.

###### 테이블 파티셔닝 (수동 구현)

MySQL 5.5에서 제공되는 기본적인 파티셔닝 기능은 제한적입니다.
대신, 수동으로 테이블을 분리하고 애플리케이션 레벨에서 이를 관리하는 방식을 사용할 수 있습니다.

###### 수동 파티셔닝 설계

- 기존 테이블(`payments`)을 기간 기반 또는 범위 기반으로 나누어 별도의 테이블로 분리합니다.
- ID 범위 기준으로 파티션 분리:
    - ID는 증가형 정수이므로, 특정 범위를 기준으로 물리적 테이블을 분리할 수 있습니다.
    - 예를 들어:

        ```sql
        CREATE TABLE payments_1_to_100M LIKE payments;
        CREATE TABLE payments_100M_to_200M LIKE payments;
        CREATE TABLE payments_200M_to_300M LIKE payments;
        ```

    - 기존 데이터를 새로운 테이블로 옮깁니다:

        ```sql
        INSERT INTO payments_1_to_100M SELECT * FROM payments WHERE id BETWEEN 1 AND 100000000;
        DELETE FROM payments WHERE id BETWEEN 1 AND 100000000;
        ```

- 데이터 삽입:
    - 데이터는 새로운 `payments` 테이블에 삽입하며, 일정 기준에 도달하면 다음 파티션 테이블로 데이터를 옮깁니다.

###### 조회 시 파티션 통합

조회 작업은 뷰(View)를 사용하여 통합 인터페이스를 제공합니다:

```sql
CREATE VIEW all_payments AS
SELECT *
FROM payments_1_to_100M
UNION ALL
SELECT *
FROM payments_100M_to_200M
UNION ALL
SELECT *
FROM payments_200M_to_300M;
```

- 실시간 쿼리에서 `all_payments`를 사용하면 기존 단일 테이블과 동일한 인터페이스를 제공합니다.
- 쿼리는 MySQL의 최적화 엔진에 따라 각 파티션 테이블에 병렬로 접근합니다.

하지만 뷰를 사용하는 경우, 뷰 자체에는 인덱스를 생성할 수 없기 때문에 기존 인덱스의 기능을 유지하기 위한 별도의 접근 방식이 필요합니다.

뷰를 사용하는 경우, 뷰 자체에는 인덱스를 생성할 수 없기 때문에 기존 인덱스의 기능을 유지하기 위한 별도의 접근 방식이 필요합니다. 각 파티션(분리된 테이블)에 존재하는 인덱스를 활용하면서, 통합 조회 성능을 유지할 수 있는 몇 가지 전략을 아래와 같이 제안합니다.

1. 개별 파티션 테이블에 동일한 인덱스 유지

   뷰가 여러 파티션 테이블을 병합(`UNION ALL`)하므로, 각 파티션 테이블에 필요한 인덱스를 동일하게 설정해야 합니다.

   파티션 테이블에 인덱스 생성:
    - 기존 단일 테이블에 존재하던 주요 인덱스를 분리된 테이블에도 동일하게 생성합니다.
    - 예를 들어, 단일 테이블의 인덱스들을 각 파티션 테이블에도 동일하게 생성합니다.:

        ```sql
        -- 기존 단일 테이블의 인덱스
        ALTER TABLE payments ADD INDEX idx_created (created);
        ALTER TABLE payments ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments ADD INDEX idx_name_created (name, created);

        -- 파티션된 테이블의 인덱스
        ALTER TABLE payments_1_to_100M ADD INDEX idx_created (created);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_created (created);
        ALTER TABLE payments_1_to_100M ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments_1_to_100M ADD INDEX idx_name_created (name, created);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_name_created (name, created);
        ```

      이렇게 하면 MySQL은 뷰를 통한 조회 시 각 파티션 테이블에 존재하는 인덱스를 활용하여 조건에 맞는 데이터를 효율적으로 검색할 수 있습니다.

2. 쿼리 최적화: 조건절로 파티션 테이블 선택

   MySQL의 쿼리 최적화 엔진은 `UNION ALL`이 포함된 뷰를 사용할 때, `WHERE` 조건을 기반으로 쿼리를 최적화하여 불필요한 테이블에 접근하지 않도록 설계되어 있습니다.

    - 특정 조건이 있는 쿼리

      뷰에서 특정 조건(예: `id`, `created`, `uuid`)이 있는 경우, MySQL은 해당 조건이 충족되는 테이블만 스캔합니다.
      예를 들어, 다음과 같은 쿼리를 실행한다고 가정합니다:

        ```sql
        SELECT * FROM all_payments WHERE created >= '2023-01-01' AND created < '2023-02-01';
        ```

      MySQL은 `payments_1_to_100M`과 `payments_100M_to_200M`의 `created` 인덱스를 활용하여 쿼리를 최적화할 수 있습니다.

    - MySQL 실행 계획 확인

      뷰가 제대로 작동하는지 확인하려면 `EXPLAIN`을 사용하여 실행 계획을 점검합니다:

        ```sql
        EXPLAIN SELECT * FROM all_payments WHERE created >= '2023-01-01' AND created < '2023-02-01';
        ```

      결과에서 MySQL이 조건에 맞는 테이블만 스캔하고, 해당 테이블의 인덱스를 활용하는지 확인할 수 있습니다.

3. 필요 시 애플리케이션 레벨에서 파티션 관리

   뷰가 모든 테이블을 병합하는 방식은 간단하지만, 데이터 규모가 매우 클 경우 비효율적일 수 있습니다.
   애플리케이션 로직에서 조건에 따라 적절한 파티션 테이블에 직접 접근하도록 설계하면 성능을 극대화할 수 있습니다.

    - 애플리케이션 레벨에서 파티션 테이블을 직접 선택하면 불필요한 테이블 스캔을 완전히 제거할 수 있습니다.
    - 조건이 명확한 경우 뷰를 사용하지 않아도 되므로 성능이 크게 개선됩니다.

   예를 들어, ID나 날짜를 기준으로 각 파티션 테이블을 선택하는 애플리케이션 레벨 로직을 추가할 수 있습니다:

    ```python
    def get_partition_table(created_date):
        if created_date < '2023-01-01':
            return 'payments_1_to_100M'
        elif created_date < '2024-01-01':
            return 'payments_100M_to_200M'
        else:
            return 'payments_200M_to_300M'

    query = f"SELECT * FROM {get_partition_table(requested_date)} WHERE created = '{requested_date}'"
    ```

4. 인덱스 중복 제거 및 통합 테이블 조회 성능 최적화

   뷰와 파티션 테이블 접근을 병행하는 경우, 다음과 같은 최적화 전략을 추가로 사용할 수 있습니다.

    - 중복 인덱스 최소화
        - 뷰를 사용하면 각 테이블에 동일한 인덱스가 있어야 하므로, 필요한 최소 인덱스만 유지합니다.
        - 조회 패턴 분석을 통해 자주 사용되지 않는 인덱스를 제거합니다.

    - 데이터 분할 기준 검토
        - ID 기반 분할 외에, 데이터 접근 패턴을 분석하여 `created` 날짜 또는 `uuid`와 같은 다른 기준으로 분할을 고려할 수 있습니다.
        - 예: 날짜 기반 테이블 분리:

            ```sql
            CREATE TABLE payments_2023 LIKE payments;
            CREATE TABLE payments_2024 LIKE payments;
            ```

5. 대규모 뷰 최적화가 어려운 경우

   뷰를 통한 통합 조회가 특정 쿼리에서 비효율적이라면, 머티리얼라이즈드 뷰를 사용하거나 주기적으로 통합 테이블을 생성하는 방식을 도입할 수 있습니다.

   > 머티리얼라이즈드 뷰(Materialized View)는 데이터베이스에서 뷰(View)의 일종으로, 질의 결과를 디스크에 저장하여 실제 데이터처럼 사용할 수 있는 데이터베이스 객체입니다.
   >
   > ```sql
    > CREATE MATERIALIZED VIEW mv_sales_summary AS
    > SELECT region, SUM(sales) AS total_sales
    > FROM sales
    > GROUP BY region;
    >
    > -- 머티리얼라이즈드 뷰를 최신 상태로 유지하려면 새로 고침을 수행해야 합니다:
    > -- Oracle 예제: Complete Refresh
    > BEGIN
    > DBMS_MVIEW.REFRESH('MV_SALES_SUMMARY', 'COMPLETE');
    > END;
    >
    > -- PostgreSQL 예제: Refresh Materialized View
    > REFRESH MATERIALIZED VIEW mv_sales_summary;
    > ```
   >
   > 일반적인 뷰(View)는 질의 실행 시점에 데이터를 실시간으로 조회하는 반면, 머티리얼라이즈드 뷰는 질의 결과를 물리적으로 저장하여 나중에 필요할 때 빠르게 조회할 수 있습니다.

    - 통합 테이블 생성
        - 일정 주기로 각 파티션 테이블의 데이터를 하나의 통합 테이블로 병합합니다.
        - 통합 테이블은 읽기 전용으로 사용하며, 정기적으로 업데이트합니다.

            ```sql
            CREATE TABLE consolidated_payments LIKE payments;
            INSERT INTO consolidated_payments
            SELECT * FROM payments_1_to_100M
            UNION ALL
            SELECT * FROM payments_100M_to_200M
            UNION ALL
            SELECT * FROM payments_200M_to_300M;
            ```

    - 주기적인 업데이트
        - Cron 스케줄러나 배치 작업을 통해 통합 테이블을 갱신합니다.
        - 장점: 복잡한 조회 쿼리를 단순화할 수 있음.
        - 단점: 통합 데이터가 실시간으로 업데이트되지 않을 수 있음.

###### 데이터 아카이빙

과거 데이터를 아카이브 테이블로 이동하여 현재 테이블 크기를 줄입니다.
이 방법은 오래된 결제 데이터를 조회할 필요가 적은 경우에 적합합니다.

###### 아카이브 테이블 설계

- 과거 데이터를 별도의 테이블로 이동:

    ```sql
    CREATE TABLE payments_archive LIKE payments;
    INSERT INTO payments_archive SELECT * FROM payments WHERE created_at < '2022-01-01';
    DELETE FROM payments WHERE created_at < '2022-01-01';
    ```

- 정기적인 스케줄링:
    - 데이터를 특정 기간 동안 유지하고, 그 이후에 아카이빙 작업을 수행합니다.
    - 스케줄링 도구(Cron + Shell Script 또는 MySQL Event)를 사용하여 배치 작업으로 구현.

###### 아카이브 데이터 통합 조회

- 아카이브 테이블을 포함하여 조회할 수 있도록 뷰를 제공합니다:

    ```sql
    CREATE VIEW all_payments AS
    SELECT * FROM payments
    UNION ALL
    SELECT * FROM payments_archive;
    ```

- 최적화: 과거 데이터에 접근이 적다면, 아카이브 테이블을 읽기 전용 서버(레플리카)로 옮겨 읽기 부하를 분산시킵니다.

###### 인덱스 최적화 및 테이블 정리

대규모 테이블에서는 인덱스 최적화가 성능에 큰 영향을 미칩니다.

###### 3.1 중복 인덱스 제거

- `SHOW INDEX FROM payments`로 중복 인덱스를 확인합니다.
- 불필요한 인덱스를 삭제하여 인덱스 유지 비용을 줄입니다:

    ```sql
    DROP INDEX idx_column_name ON payments;
    ```

###### 3.2 복합 인덱스 추가

- 조회 패턴을 분석하여 복합 인덱스를 추가합니다.
- 예: `id`와 `created_at`을 함께 사용하는 쿼리 최적화를 위해 복합 인덱스를 생성:

    ```sql
    ALTER TABLE payments ADD INDEX idx_id_created_at (id, created_at);
    ```

###### 점진적 테이블 분리

대규모 테이블을 분리하기 위해 서비스 중단 없이 데이터 마이그레이션을 수행합니다.

###### 배치로 데이터 옮기기

- 데이터를 분리하는 동안 트랜잭션을 사용하여 무결성을 유지:

    ```sql
    INSERT INTO payments_1_to_100M SELECT * FROM payments WHERE id BETWEEN 1 AND 100000000;
    DELETE FROM payments WHERE id BETWEEN 1 AND 100000000;
    ```

- 한 번에 모든 데이터를 옮기지 말고, 작은 청크 단위로 배치 처리합니다:

    ```sql
    SET @start = 1, @batch_size = 100000;
    WHILE (@start <= 100000000) DO
        INSERT INTO payments_1_to_100M
        SELECT * FROM payments
        WHERE id BETWEEN @start AND @start + @batch_size - 1;
        DELETE FROM payments
        WHERE id BETWEEN @start AND @start + @batch_size - 1;
        SET @start = @start + @batch_size;
    END WHILE;
    ```

###### 점진적 데이터 분리 전략

- 새로운 데이터를 추가할 때부터 파티션 테이블로 직접 저장.
- 예: ID 생성 규칙에 따라 `payments_200M_to_300M` 테이블로 삽입.

<!-- /curriculum-chunk -->

## 인덱스와 조회 성능

### 3억건의 데이터의 B+Tree 깊이 계산

#### 원문: 3억건의 데이터의 B+Tree 깊이 계산

<!-- curriculum-chunk: sha256=ff783a17d8347f66fc6f37b3b83e6a083d5ab87fe00eb81c404c52d3fdb9d978 major=database-storage-search-nosql mid=인덱스와 조회 성능 sub=3억건의 데이터의 B+Tree 깊이 계산 sources=source/interview_questions2.md:1409-1498, source/interviews2.md:1409-1498 -->

> Source: `source/interview_questions2.md:1409-1498`
> Classification reason: index/query performance
> Duplicate source aliases: `source/interview_questions2.md:1409-1498, source/interviews2.md:1409-1498`

##### 3억건의 데이터의 B+Tree 깊이 계산

B+ Tree에서 데이터의 깊이를 계산하려면 트리의 구조와 노드에 저장할 수 있는 데이터의 양을 기반으로 깊이를 추정할 수 있습니다.
B+ Tree는 모든 리프 노드가 동일한 깊이에 있고, 내부 노드가 많은 키를 저장할 수 있으므로 깊이가 비교적 얕은 것이 특징입니다.

###### B+ Tree의 주요 특징

- 차수(Order, `m`): 각 노드는 최대 `m-1`개의 키를 저장할 수 있고, 최대 `m`개의 자식 노드를 가질 수 있습니다.
- 데이터 저장 위치: 모든 데이터는 리프 노드에 저장되며, 내부 노드는 검색을 위한 키만 저장합니다.
- 균형 트리: 트리는 항상 균형 상태를 유지하며, 모든 리프 노드는 동일한 깊이를 가집니다.

###### B+ Tree의 최대 용량

- 한 레벨의 노드 수가 증가할수록 저장 가능한 키의 수는 기하급수적으로 증가합니다.
- 루트 노드는 최대 `m-1`개의 키를 저장합니다.
- 각 레벨의 노드는 최대 `m^h`개의 키를 관리합니다(`h`는 깊이).

###### 인덱싱 예시

MySQL에서 일반적으로 사용되는 InnoDB의 B+ Tree는 16KB 페이지 크기를 사용하며, 페이지마다 약 200~400개의 키를 저장할 수 있습니다(각 키의 크기에 따라 다름).

###### 깊이 계산 방법

1. 노드당 키 수 추정: 데이터가 3억 건이고, 각 노드(페이지)가 최대 300개의 키를 저장할 수 있다고 가정합니다.

2. 트리 깊이 공식

    - 리프 노드의 개수를 `L`이라 하고, 데이터 개수를 `N`, 노드의 키 수를 `m`이라 할 때:
      $$
      L = \lceil \frac{N}{m} \rceil
      $$
      $$
      h = \lceil \log_{m} L \rceil
      $$
      여기서:
    - $N = 300,000,000$ (데이터 수).
    - $m = 300$ (노드당 최대 키 수).

3. 구체적인 계산

    - 리프 노드의 수 $L$:
      $$
      L = \lceil \frac{300,000,000}{300} \rceil = 1,000,000 \, \text{(리프 노드 개수)}.
      $$

    - 트리 깊이 계산

      내부 노드의 자식 수(차수)를 $m = 300$으로 가정했을 때:
      $$
      h = \lceil \log_{300} 1,000,000 \rceil
      $$

      이를 계산하면:
        1. $\log_{300} 1,000,000 = \frac{\log_{10} 1,000,000}{\log_{10} 300}$
        2. $\log_{10} 1,000,000 = 6$ (10의 6제곱).
        3. $\log_{10} 300 \approx 2.477$.
        4. $\frac{6}{2.477} \approx 2.42$.

      즉, $h \approx 3$.
      트리의 깊이는 약 3~4 레벨로 유지됩니다.

      이러한 얕은 깊이 덕분에 데이터 접근 시 최대 3~4번의 I/O만으로도 원하는 데이터를 찾을 수 있습니다.

###### B+ Tree 깊이와 데이터 접근 시간

- 접근 시간
    - B+ Tree는 깊이가 얕기 때문에 검색 성능이 매우 우수합니다.
    - 3억 건의 데이터에서도 3~4번의 디스크 페이지 읽기만으로 원하는 데이터를 찾을 수 있습니다.

- 다양한 차수와 키 수 변화

  차수가 클수록(페이지당 저장 가능한 키 수가 많을수록), 트리의 깊이는 더 얕아집니다:
    - $m = 400$일 경우:
        - 리프 노드 수 $L = 750,000$.
        - 깊이 $h = \lceil \log_{400} 750,000 \rceil \approx 3$.

###### 추가 고려 사항

- InnoDB의 실제 구현
    - InnoDB는 리프 노드에 데이터 레코드를 직접 저장하고, 내부 노드는 검색 키만 저장합니다.
    - 내부 노드는 더 많은 키를 저장할 수 있으므로, 실제 트리의 깊이는 이론적으로 계산된 값보다 조금 더 얕을 수 있습니다.

- 다양한 데이터 분포
    - 데이터가 균등 분포되지 않고 특정 범위에 집중되어 있으면, 일부 리프 노드가 비대해질 수 있습니다.
    - 그러나 B+ Tree는 자동으로 균형을 유지하므로 깊이는 크게 변하지 않습니다.

- 데이터 추가 시 깊이 변화
    - 데이터가 지속적으로 추가되면 리프 노드와 내부 노드가 분할되며 트리의 깊이가 증가할 수 있습니다.
    - 그러나 실질적으로 데이터 수가 수십 배로 증가하지 않는 이상 깊이는 일정하게 유지됩니다.

<!-- /curriculum-chunk -->

### B Tree & B+ Tree

#### 원문: B Tree & B+ Tree

<!-- curriculum-chunk: sha256=fea2e1bafed0b5b9afea28c93665a0d2590debac93641cd0ce8ffa3c4c17a784 major=database-storage-search-nosql mid=인덱스와 조회 성능 sub=B Tree & B+ Tree sources=source/interview_questions.md:2924-3024, source/interviews.md:2924-3024 -->

> Source: `source/interview_questions.md:2924-3024`
> Classification reason: index/query performance
> Duplicate source aliases: `source/interview_questions.md:2924-3024, source/interviews.md:2924-3024`

##### B Tree & B+ Tree

B Tree와 B+ Tree는 모두 균형 잡힌 트리 구조로, 대용량 데이터의 효율적인 검색, 삽입, 삭제를 위해 설계된 자료구조입니다.
B Tree와 B+ Tree는 모두 다중 경로 검색 트리이지만, 데이터 저장 방식, 리프 노드의 구조, 검색 효율성 등에서 차이가 있습니다.

- B Tree는 *모든 노드에서 데이터를 찾을 수 있어 특정 검색에서 더 효율적*일 수 있습니다.
- B+ Tree는 *순차 접근과 범위 검색에서 뛰어난 성능*을 보이며, *더 낮은 트리 높이로 인해 대량의 데이터를 다룰 때 유리*할 수 있습니다.

실제 적용 시에는 사용 사례, 데이터의 특성, 접근 패턴 등을 고려하여 적절한 구조를 선택해야 합니다.
데이터베이스 시스템에서는 주로 B+ Tree가 선호되는 경향이 있지만, 특정 상황에서는 B Tree가 더 적합할 수 있습니다.

1. 데이터 저장 위치
    - B Tree
        - 모든 노드(내부 노드와 리프 노드)에 키와 데이터를 저장합니다.
        - 각 노드는 키와 해당 키에 연관된 데이터, 그리고 자식 노드에 대한 포인터를 포함합니다.

    - B+ Tree
        - 내부 노드에는 키만 저장하고, 실제 데이터는 리프 노드에만 저장합니다.
        - 내부 노드는 키와 자식 노드에 대한 포인터만을 포함합니다.

2. 리프 노드 구조
    - B Tree
        - 리프 노드와 내부 노드의 구조가 동일합니다.
        - 리프 노드들 사이에 특별한 연결이 없습니다.

    - B+ Tree
        - 리프 노드들이 연결 리스트로 연결되어 있습니다.
        - 이 연결을 통해 순차적 접근이 가능합니다.

3. 검색 효율성
    - B Tree
        - 검색 시 루트에서 리프까지 내려가지 않아도 중간에 데이터를 찾을 수 있습니다.
        - 최악의 경우 루트에서 리프까지 탐색해야 합니다.

    - B+ Tree
        - 항상 리프 노드까지 내려가야 데이터를 찾을 수 있습니다.
        - 그러나 내부 노드에 키만 저장하므로 더 많은 키를 저장할 수 있어, 트리의 높이가 낮아질 수 있습니다.

4. 범위 검색 (Range Query) 성능
    - B Tree
        - 범위 검색 시 여러 노드를 오가며 검색해야 할 수 있습니다.

    - B+ Tree
        - 리프 노드가 연결되어 있어 범위 검색이 매우 효율적입니다.
        - 시작 지점을 찾은 후 연결된 리프 노드를 따라 순차적으로 접근할 수 있습니다.

5. 트리 높이와 균형
    - B Tree
        - 데이터가 내부 노드에도 저장되므로 B+ Tree에 비해 트리의 높이가 높을 수 있습니다.

    - B+ Tree
        - 내부 노드에 키만 저장하므로 더 많은 키를 저장할 수 있어 트리의 높이가 낮아질 수 있습니다.
        - 이는 검색 시 더 적은 I/O 연산을 필요로 할 수 있음을 의미합니다.

6. 노드 크기와 분기 요소 (Fanout)
    - B Tree
        - 각 노드가 키와 데이터를 모두 저장하므로 노드의 크기가 더 클 수 있습니다.
        - 이로 인해 각 노드의 자식 수(분기 요소)가 B+ Tree에 비해 적을 수 있습니다.

    - B+ Tree
        - 내부 노드가 키만 저장하므로 더 많은 키를 포함할 수 있습니다.
        - 이는 더 높은 분기 요소를 가능하게 하여, 트리의 높이를 낮출 수 있습니다.

7. 구현 복잡성
    - B Tree
        - 모든 노드가 동일한 구조를 가지므로 구현이 상대적으로 간단할 수 있습니다.

    - B+ Tree
        - 내부 노드와 리프 노드의 구조가 다르고, 리프 노드 간의 연결을 관리해야 하므로 구현이 더 복잡할 수 있습니다.

다음은 B Tree와 B+ Tree의 노드 구조를 Java로 간단히 표현한 예시입니다:

```java
// B Tree 노드는 키, 데이터, 자식 노드 참조를 모두 포함합니다.
public class BTreeNode {
    private int[] keys;  // 키 배열
    private Object[] data;  // 데이터 배열
    private BTreeNode[] children;  // 자식 노드 배열

    // 생성자, getter, setter 등 메서드...
}

// B+ Tree는 내부 노드와 리프 노드가 다른 구조를 가집니다.
// B+ Tree 내부 노드는 키와 자식 노드 참조만을 포함합니다.
public class BPlusTreeInternalNode {
    private int[] keys;  // 키 배열
    private BPlusTreeNode[] children;  // 자식 노드 배열

    // 생성자, getter, setter 등 메서드...
}

// B+ Tree 리프 노드는 키, 데이터, 그리고 다음 리프 노드에 대한 참조를 포함합니다.
public class BPlusTreeLeafNode {
    private int[] keys;  // 키 배열
    private Object[] data;  // 데이터 배열
    private BPlusTreeLeafNode next;  // 다음 리프 노드에 대한 참조

    // 생성자, getter, setter 등 메서드...
}
```

<!-- /curriculum-chunk -->

## 트랜잭션, 락, 격리

### 2단계커밋과 2단계 잠금

#### 원문: 2단계커밋과 2단계 잠금

<!-- curriculum-chunk: sha256=680eb04c6ec09929ee4ccdef10556c4587f66e75907eb0e07782f8a41489cb81 major=database-storage-search-nosql mid=트랜잭션, 락, 격리 sub=2단계커밋과 2단계 잠금 sources=source/interview_questions.md:7980-8056, source/interviews.md:7928-8004 -->

> Source: `source/interview_questions.md:7980-8056`
> Classification reason: transaction/isolation
> Duplicate source aliases: `source/interview_questions.md:7980-8056, source/interviews.md:7928-8004`

##### 2단계커밋과 2단계 잠금

2단계 커밋(Two-Phase Commit, 2PC)은 분산 시스템에서 트랜잭션의 원자성(Atomicity)을 보장하기 위한 프로토콜입니다.
주로 데이터베이스 시스템이나 분산 트랜잭션 관리에서 사용됩니다.
여러 시스템이나 노드가 하나의 트랜잭션에 참여할 때, 트랜잭션의 모든 참여자가 성공적으로 트랜잭션을 수행했는지 확인하고, 모든 참여자가 동의했을 때만 최종적으로 커밋하는 방식입니다.

1. 준비(Prepare) 단계
    - 코디네이터(또는 트랜잭션 관리자)는 각 참여 노드(서버나 데이터베이스)에 트랜잭션을 준비할 수 있는지 물어봅니다.
    - 각 노드는 트랜잭션을 수행할 준비가 되면, "준비 완료(Prepared)" 상태가 됩니다.
    - 모든 노드가 준비 상태라는 응답을 보내면, 코디네이터는 트랜잭션을 커밋할 준비가 된 것으로 간주합니다.

2. 커밋(Commit) 단계
    - 코디네이터는 모든 참여 노드가 준비되었다는 응답을 받으면, 각 노드에 트랜잭션을 커밋하라고 지시합니다.
    - 각 노드는 트랜잭션을 실제로 커밋하고, 성공 여부를 코디네이터에게 보고합니다.
    - 만약 어떤 노드라도 준비되지 않았다고 응답하거나, 중간에 실패하면 코디네이터는 트랜잭션을 롤백(Rollback) 하라고 지시합니다.

2단계 커밋의 특징:

- 원자성 보장: 모든 참여 노드가 트랜잭션에 성공하거나, 실패하면 트랜잭션 전체가 롤백됩니다. 중간에 일부 노드만 성공하고 다른 노드가 실패하는 경우는 없습니다.
- 동기화 비용: 2PC는 각 노드와 코디네이터 간의 통신을 여러 번 요구하기 때문에, 성능 상의 오버헤드가 발생할 수 있습니다.
- 문제점: 코디네이터가 장애가 나거나 노드 중 하나가 응답하지 않으면 트랜잭션이 오랫동안 대기 상태에 머물 수 있는 블록킹 문제가 발생할 수 있습니다.

예를 들어, 두 개의 데이터베이스 시스템이 동일한 트랜잭션에 참여하는 경우, 하나의 데이터베이스에서 작업이 완료되었지만 다른 데이터베이스에서 작업이 실패할 수 있습니다.
이때 2단계 커밋 프로토콜은 두 데이터베이스가 모두 성공했는지 확인한 후, 성공하면 트랜잭션을 커밋하고, 하나라도 실패하면 롤백하여 데이터의 일관성을 보장합니다.

2단계 잠금(Two-Phase Locking, 2PL)은 데이터베이스 트랜잭션에서 일관성을 보장하기 위해 사용되는 잠금(lock) 프로토콜입니다.
2PL은 직렬 가능성(Serializability), 즉 트랜잭션이 동시에 수행될 때 발생할 수 있는 경쟁 조건(Race Condition)이나 데이터 불일치 문제를 방지하는 데 사용됩니다.

1. 잠금 획득 단계 (Growing Phase)
    - 트랜잭션은 필요한 모든 자원에 대해 잠금을 요청합니다.
    - 트랜잭션이 필요한 자원에 대한 모든 잠금을 획득할 때까지 잠금만 요청할 수 있으며, 이 동안 잠금을 해제할 수 없습니다.

2. 잠금 해제 단계 (Shrinking Phase)
    - 트랜잭션이 더 이상 자원에 대한 잠금을 획득하지 않고, 잠금을 해제하는 단계입니다.
    - 일단 잠금을 해제하기 시작하면, 그 이후에는 더 이상 새로운 자원에 대한 잠금을 요청할 수 없습니다.

여러 트랜잭션들이 동시에 어떤 자원에 대해 잠금을 요청하는 경우, 잠금 관리자(lock manager)는 누가 먼저 요청했는지를 기준으로 우선순위를 결정합니다.
일반적으로 요청 시점에 따라 우선순위가 정해지며, 먼저 요청한 트랜잭션이 우선권을 갖습니다.

2단계 잠금의 특징:

- 직렬 가능성 보장: 2PL은 트랜잭션 간의 직렬 가능성을 보장합니다. 즉, 동시에 여러 트랜잭션이 실행되더라도 그 결과는 순차적으로 실행한 것과 동일하게 보장됩니다.
- 교착 상태 가능성: 2PL은 교착 상태(Deadlock)를 발생시킬 가능성이 있습니다. 트랜잭션이 자원을 획득하기 위해 서로 대기하는 상황이 발생할 수 있기 때문입니다.
- 엄격 2단계 잠금 (Strict 2PL): 모든 자원의 잠금을 트랜잭션이 완료될 때까지 유지하다가 트랜잭션이 끝나면 한 번에 모든 잠금을 해제하는 방식으로, 교착 상태를 방지할 수 있지만 더 많은 잠금을 오랫동안 유지하게 됩니다.

예를 들어, 트랜잭션 A가 먼저 자원 X에 대해 잠금을 요청하고, 자원 Y에 대한 잠금을 요청하는 동안, 트랜잭션 B가 자원 Y에 대해 잠금을 요청할 수 있습니다.
2단계 잠금 프로토콜에서는 트랜잭션이 필요한 자원에 대한 잠금을 요청할 때, 먼저 요청한 트랜잭션이 우선적으로 처리됩니다.
따라서 2단계 잠금 프로토콜을 통해 트랜잭션 A가 자원 Y에 대한 잠금을 성공적으로 획득하고 작업을 완료한 후 자원을 해제하면, 그제서야 트랜잭션 B는 자원 Y에 접근할 수 있습니다.
이를 통해 트랜잭션 간 충돌을 피할 수 있습니다.

만약 트랜잭션 A와 B가 서로 다른 자원에 대해 잠금을 요청하고, 동시에 다른 트랜잭션이 이미 소유하고 있는 자원에 대해 잠금을 요청하는 상황이 발생하면 *교착 상태(Deadlock)*가 발생할 수 있습니다.
이를 방지하기 위해 여러 가지 기법이 사용됩니다:

- 타임아웃(Timeouts):

  특정 시간 동안 잠금을 획득하지 못하면 트랜잭션이 실패하고 롤백되는 방식으로, 교착 상태를 피할 수 있습니다.

- 교착 상태 탐지(Deadlock Detection):

  시스템이 교착 상태를 감지하고, 특정 트랜잭션을 중단시켜 교착 상태를 해결합니다.

- 락 순서 규칙(Lock Ordering):

  자원에 대한 잠금을 요청할 때 항상 일정한 순서를 따르게 함으로써 교착 상태를 방지하는 방법입니다.

  예를 들어, 자원 X를 먼저 잠근 트랜잭션이 자원 Y에 대한 잠금을 요청할 때, 자원 Y를 먼저 잠근 트랜잭션이 자원 X에 대한 잠금을 요청하지 않도록 규칙을 설정하는 방식입니다.

2단계 커밋과 2단계 잠금의 차이점:

- 2단계 커밋:

  분산 시스템에서 트랜잭션의 원자성(Atomicity)을 보장하기 위한 프로토콜입니다.
  트랜잭션이 여러 노드에서 동시에 성공했는지 확인하고 최종적으로 커밋 또는 롤백을 결정합니다.

- 2단계 잠금:

  데이터베이스에서 트랜잭션 간의 일관성(Consistency)과 직렬 가능성을 보장하는 프로토콜입니다.
  자원에 대한 잠금을 두 단계로 나누어 트랜잭션이 완료될 때까지 필요한 자원을 독점적으로 사용할 수 있게 합니다.

이 두 가지 개념은 분산 트랜잭션이나 데이터베이스의 일관성을 보장하는 중요한 메커니즘이지만, 적용되는 대상과 해결하고자 하는 문제의 관점에서 다릅니다.

<!-- /curriculum-chunk -->

### ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)

#### 원문: ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)

<!-- curriculum-chunk: sha256=c4aa600316cb96440ad050f2e5eaeaae6fc1c58fa559b1f8a514eb8c00daf1bf major=database-storage-search-nosql mid=트랜잭션, 락, 격리 sub=ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어) sources=source/interview_questions.md:2522-2576, source/interviews.md:2522-2576 -->

> Source: `source/interview_questions.md:2522-2576`
> Classification reason: transaction/isolation
> Duplicate source aliases: `source/interview_questions.md:2522-2576, source/interviews.md:2522-2576`

##### ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)

ACID(Atomicity, Consistency, Isolation, Durability)와 MVCC(Multi-Version Concurrency Control, 다중 버전 동시성 제어)는 데이터베이스 시스템에서 트랜잭션의 일관성과 동시성을 관리하기 위한 중요한 개념들입니다. 두 개념은 서로 보완적이며, ACID 트랜잭션의 격리성(Isolation)을 보장하기 위해 MVCC가 많이 사용됩니다.

MVCC는 데이터베이스에서 다중 버전의 데이터를 저장함으로써 동시성 문제를 해결하는 기법입니다.
MVCC는 각 트랜잭션이 데이터에 접근할 때마다 데이터의 버전을 기준으로 작업을 수행하여, 트랜잭션 간의 충돌을 줄이고 동시성 제어를 효율적으로 처리합니다.
이를 통해 읽기 작업과 쓰기 작업이 서로 충돌하지 않게 만듭니다.

MVCC의 주요 특징:

- 읽기 작업은 항상 일관된 데이터 버전을 읽음: 읽기 트랜잭션이 실행되는 동안 다른 트랜잭션에서 쓰기 작업이 이루어져도 읽기 작업은 자신이 시작할 때의 데이터 버전을 유지합니다.
- 쓰기 작업은 새로운 버전을 생성: 쓰기 트랜잭션이 완료되면 기존 데이터를 수정하지 않고, 새로운 버전의 데이터를 생성합니다. 그 결과 읽기 작업에 영향을 주지 않으면서 쓰기가 가능해집니다.
- 낙관적 동시성 제어(Optimistic Concurrency Control)에 가깝습니다. 쓰기 충돌이 발생할 가능성을 낮게 보고, 충돌이 발생하면 나중에 해결하는 방식입니다.

ACID와 MVCC의 연관성:

1. Atomicity (원자성):

   MVCC는 원자성을 유지하는 데 도움이 됩니다.
   MVCC에서 각 트랜잭션은 고유한 데이터 버전을 사용하며, 트랜잭션이 성공하면 새로운 데이터 버전이 커밋되고 실패하면 롤백됩니다.
   이로써 모든 변경사항이 한 번에 적용되거나 전혀 적용되지 않는 원자성을 보장할 수 있습니다.

2. Consistency (일관성):

   MVCC는 일관성도 유지합니다.
   트랜잭션이 시작되면, 트랜잭션은 고정된 시점의 데이터 버전에 기반해 작업을 수행하며, 무결성을 위반하는 작업은 거부됩니다.
   트랜잭션이 성공적으로 커밋되면, 모든 작업이 일관성 있게 반영되며 새로운 데이터 버전이 생성됩니다.

3. Isolation (격리성):

   MVCC는 주로 격리성을 보장하기 위해 설계된 기법입니다.
   여러 트랜잭션이 동시에 실행될 때, MVCC는 각각의 트랜잭션에 별도의 데이터 버전을 제공하여 트랜잭션 간의 간섭을 최소화합니다.

   예를 들어, 트랜잭션 1이 데이터를 읽고 있는 동안 트랜잭션 2가 해당 데이터를 수정할 수 있습니다.
   이 경우, 트랜잭션 1은 수정 전의 데이터 버전을 계속 읽을 수 있으며, 트랜잭션 2는 새로운 버전을 생성합니다.
   이 방식으로 Non-repeatable Read와 같은 문제를 해결할 수 있습니다.

   격리 수준이 낮을 때도 읽기 작업이 쓰기 작업에 방해받지 않기 때문에 동시성 처리 성능이 향상됩니다.
   이는 특히 READ COMMITTED나 REPEATABLE READ 격리 수준에서 많이 사용됩니다.

4. Durability (지속성):

   MVCC는 트랜잭션이 커밋되면, 새로운 데이터 버전이 영구적으로 저장됩니다.
   이때 데이터는 로그와 같은 메커니즘에 의해 안전하게 보존되며, 시스템 장애가 발생해도 데이터를 복구할 수 있습니다.
   데이터 버전 관리 덕분에 지속성이 보장됩니다.

MVCC는 PostgreSQL, MySQL(InnoDB 스토리지 엔진)과 같은 현대의 많은 관계형 데이터베이스에서 동시성 제어를 위해 채택하고 있는 기법입니다.
이러한 데이터베이스에서는 ACID 트랜잭션과 MVCC가 결합되어, 높은 동시성을 지원하면서도 트랜잭션의 일관성을 보장합니다.

- PostgreSQL:
  트랜잭션을 실행할 때, PostgreSQL은 각 트랜잭션이 시작된 시점의 데이터 스냅샷을 제공하며, 트랜잭션이 완료될 때까지 해당 스냅샷이 유지됩니다.
  트랜잭션이 커밋되면 새로운 데이터 버전이 생성됩니다.

- MySQL (InnoDB):
  InnoDB 스토리지 엔진도 MVCC를 통해 트랜잭션 격리성을 보장하며, 주로 `REPEATABLE READ` 격리 수준에서 사용됩니다.
  InnoDB는 쓰기 작업 시 새로운 데이터 버전을 생성하여 읽기 작업과 충돌을 방지합니다.

<!-- /curriculum-chunk -->

### MVCC와 스냅숏 격리

#### 원문: MVCC와 스냅숏 격리

<!-- curriculum-chunk: sha256=2e97940a6ed53cea83d7bf96dfe959c9151ae38e7a968ed7758f10129c81015c major=database-storage-search-nosql mid=트랜잭션, 락, 격리 sub=MVCC와 스냅숏 격리 sources=source/interview_questions.md:7806-7811, source/interviews.md:7754-7759 -->

> Source: `source/interview_questions.md:7806-7811`
> Classification reason: transaction/isolation
> Duplicate source aliases: `source/interview_questions.md:7806-7811, source/interviews.md:7754-7759`

##### MVCC와 스냅숏 격리

데이터베이스는 객체마다 커밋된 버전 여러 개를 유지할 수 있어야 한다.
진행중인 여러 트랜잭션에서 서로 다른 시점의 데이터베이스 상태를 봐야 할 수도 있기 때문이다.
데이터베이스가 객체의 여러 버전을 함께 유지하므로 이 기법은 다중 버전 동시성 제어(multi-version concurrency control, MVCC)라고 한다.

<!-- /curriculum-chunk -->

### [DB - ACID](https://postgresql.kr/blog/pg_phantom_read.html)

#### 원문: [DB - ACID](https://postgresql.kr/blog/pg_phantom_read.html)

<!-- curriculum-chunk: sha256=88fd4a6f201c36299537a989e42bb8daf227bf7993412ca987a0b6b1972a6613 major=database-storage-search-nosql mid=트랜잭션, 락, 격리 sub=[DB - ACID](https://postgresql.kr/blog/pg_phantom_read.html) sources=source/interview_questions.md:2323-2521, source/interviews.md:2323-2521 -->

> Source: `source/interview_questions.md:2323-2521`
> Classification reason: transaction/isolation
> Duplicate source aliases: `source/interview_questions.md:2323-2521, source/interviews.md:2323-2521`

##### [DB - ACID](https://postgresql.kr/blog/pg_phantom_read.html)

ACID는 데이터베이스 트랜잭션의 네 가지 핵심 속성을 나타내는 약어로, 데이터베이스의 무결성과 일관성을 보장하는 기본 원칙입니다.

각 속성은 특정한 문제를 해결하며, 이들이 결합하여 신뢰할 수 있는 데이터베이스 시스템을 구현합니다:

- Atomicity(원자성): 트랜잭션의 모든 연산이 성공하거나 모두 실패해야 함
- Consistency(일관성): 트랜잭션 실행 전후의 데이터베이스가 일관된 상태를 유지
- Isolation(격리성): 동시에 실행되는 트랜잭션들이 서로 영향을 미치지 않음
- Durability(지속성): 성공적으로 완료된 트랜잭션의 결과는 영구적으로 보존

  "내구성"이란 일반적으로 작업이 성공적으로 완료되면 디스크가 작업으로 인한 변경 사항을 저장하는 것을 의미합니다.
  분산 데이터베이스에서 내구성이란 디스크 및/또는 다른 노드의 메모리가 변경 사항을 저장하는 것을 의미할 수 있습니다.
  Couchbase에서 네트워크가 디스크보다 훨씬 빠르기 때문에 다른 노드로의 복제는 내구성을 위해 선호되는 메커니즘입니다.
  (카우치베이스는 메모리에 상주시키다가 자체적인 로직에 따라 나주에 디스크로 flush 하기 때문)
  궁극적으로 개발자에게는 시스템 장애가 발생하더라도 변경 사항이 계속 유지된다는 의미입니다.

ACID별로 문제되는 상황과 어떻게 이를 해결하는지를 확인해보겠습니다.
ACID는 데이터베이스 트랜잭션에서 반드시 지켜야 하는 4가지 중요한 특성을 의미합니다.
이는 트랜잭션이 데이터를 안전하고 일관되게 관리할 수 있도록 보장합니다.
각 특성은 원자성(Atomicity), 일관성(Consistency), 격리성(Isolation), 지속성(Durability)으로 나뉩니다.
아래에서는 각 특성에 대한 설명과 구체적인 예시를 제공합니다.

1. Atomicity (원자성)

   원자성은 트랜잭션이 완전히 실행되거나 전혀 실행되지 않은 상태를 보장합니다.
   즉, 트랜잭션 중 일부 작업만 성공하고 나머지가 실패하는 상황은 허용되지 않으며, 중간에 문제가 발생하면 모든 작업이 롤백되어야 합니다.

    - 고객 A의 계좌에서 1000원을 출금하고, 고객 B의 계좌에 1000원을 입금하는 트랜잭션을 수행하는 중, 시스템 장애로 인해 출금은 성공했지만 입금이 실패하는 상황.

   이 시나리오에서는 트랜잭션이 출금과 입금이 모두 성공했을 때만 커밋되고, 중간에 오류가 발생하면 모든 작업이 롤백됩니다.

    ```sql
    -- 트랜잭션 시작
    BEGIN;
        -- 출금 계좌에서 금액 차감
        UPDATE accounts
        SET balance = balance - 1000
        WHERE account_id = 'A';

        -- 장애 발생 또는 논리적 오류
        -- 입금 계좌에 금액 추가
        UPDATE accounts
        SET balance = balance + 1000
        WHERE account_id = 'B';

    -- 트랜잭션 완료
    COMMIT;
    ```

   Java 코드:

    ```java
    try {
        connection.setAutoCommit(false); // 트랜잭션 시작
        withdrawStatement.executeUpdate(); // 출금 처리
        depositStatement.executeUpdate();  // 입금 처리
        connection.commit(); // 트랜잭션 성공 시 커밋
    } catch (SQLException e) {
        connection.rollback();  // 실패 시 롤백
    }
    ```

   핵심 원리:
    1. Undo 로그: 트랜잭션이 진행되는 동안 이전 상태를 기록하여, 트랜잭션 실패 시 이 로그를 사용해 데이터를 복원합니다.
    2. 이중 쓰기 버퍼: 데이터 손상 방지를 위해 데이터가 실제로 디스크에 쓰이기 전에 이중 버퍼에 기록하여 안전성을 보장합니다.

2. Consistency (일관성)

   일관성은 트랜잭션이 데이터베이스의 무결성 제약 조건을 유지해야 함을 의미합니다.
   트랜잭션이 성공적으로 완료된 후 데이터는 항상 일관된 상태에 있어야 하며, 데이터 무결성을 위반하는 트랜잭션은 실패해야 합니다.

    - 제품 A의 재고가 10개 남아 있는 상황에서, 주문을 통해 15개를 요청하는 경우. 이때 재고가 음수가 되어서는 안 되며, 트랜잭션이 실패해야 합니다.

    ```sql
    -- 재고 음수 방지를 위한 제약조건 설정
    CREATE TABLE products (
        id INT PRIMARY KEY,
        name VARCHAR(100),
        stock INT CHECK (stock >= 0), -- 재고 음수 방지
        price DECIMAL(10,2)
    );

    -- 주문 트랜잭션 실행
    BEGIN;
        -- 재고 확인
        SELECT stock FROM products WHERE id = 1;

        -- 재고 감소 (음수 발생 시 트랜잭션 실패)
        UPDATE products
        SET stock = stock - 15
        WHERE id = 1;
    ROLLBACK; -- 트랜잭션 실패
    ```

   트리거 사용: 재고가 부족한 상황에서 주문이 발생하지 않도록 트리거를 사용할 수 있습니다.

    ```sql
    CREATE TRIGGER check_stock
    BEFORE INSERT ON orders
    FOR EACH ROW
    BEGIN
        DECLARE current_stock INT;

        SELECT stock INTO current_stock
        FROM products
        WHERE id = NEW.product_id;

        IF NEW.quantity > current_stock THEN
            SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Insufficient stock';
        END IF;
    END;
    ```

   핵심 원리:
    - 제약조건은 데이터 무결성을 보장하는 데 중요한 역할을 합니다. 제약조건을 위반하는 트랜잭션은 자동으로 롤백됩니다.
    - 트리거는 트랜잭션 실행 전 추가적인 무결성 검사를 수행하여 데이터의 일관성을 보장합니다.

3. Isolation (격리성)

   격리성은 여러 트랜잭션이 동시에 실행될 때, 서로 간섭하지 않고 독립적으로 실행되도록 보장합니다.
   트랜잭션이 완료되기 전까지는 다른 트랜잭션이 그 결과를 볼 수 없습니다.

    - 사용자 A와 B가 동시에 제품 A를 주문하는 경우, 트랜잭션이 서로 간섭하지 않도록 해야 합니다. 트랜잭션 1이 제품 가격을 변경하는 동안 트랜잭션 2가 Dirty Read를 통해 변경된 값을 읽으면 안 됩니다.

   Dirty Read 문제:

    ```sql
    -- 트랜잭션 1: 가격 인상
    BEGIN;
        UPDATE products
        SET price = price * 1.1
        WHERE id = 1;
        -- 아직 커밋되지 않음

    -- 트랜잭션 2: 주문 처리
    BEGIN;
        SELECT price FROM products WHERE id = 1;
        -- 커밋되지 않은 가격을 읽음 (Dirty Read)
    COMMIT;
    ```

   해결 방법:
    - 격리 수준을 SERIALIZABLE로 설정하면 트랜잭션이 독립적으로 처리되어 동시성 문제를 방지할 수 있습니다.

    ```sql
    SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;

    BEGIN;
        SELECT * FROM products WHERE id = 1
        FOR UPDATE;  -- 행 잠금
        -- 안전한 주문 처리
    COMMIT;
    ```

   낙관적 락(Optimistic Locking):
   낙관적 락은 트랜잭션 간에 충돌이 발생하지 않을 것이라 가정하고 작업을 진행하다가, 커밋 시점에 충돌을 감지하여 처리합니다.

    ```sql
    UPDATE products
    SET price = NEW_PRICE,
        version = version + 1
    WHERE id = 1
    AND version = CURRENT_VERSION;
    ```

   핵심 원리:
    - 락 메커니즘을 사용하여 여러 트랜잭션이 동시에 같은 데이터를 수정하지 못하게 합니다.
    - 낙관적 락은 데이터 충돌 가능성을 줄이기 위해 버전 번호 등을 사용해 동시성을 제어합니다.

4. Durability (지속성)

   지속성은 트랜잭션이 커밋된 후에는 시스템 장애가 발생하더라도 변경 사항이 영구적으로 유지되어야 함을 의미합니다.
   트랜잭션이 커밋된 데이터는 항상 안전하게 저장됩니다.

    - 고객이 주문을 제출하고 트랜잭션이 커밋된 후 시스템 장애가 발생해도, 주문 데이터가 손실되지 않아야 합니다.

   WAL (Write-Ahead Logging) 사용:
   트랜잭션에서 변경된 사항을 데이터베이스에 적용하기 전에 로그에 먼저 기록합니다. 시스템 장애 발생 시 이 로그를 바탕으로 데이터를 복구할 수 있습니다.

    ```plaintext
    Log Entry {
        LSN: 1234,              // 로그 순번
        TransactionID: T123,     // 트랜잭션 ID
        PageID: P456,            // 수정된 페이지 ID
        Operation: UPDATE,       // 수행된 연산
        OldValue: 10,            // 이전 값
        NewValue: 5,             // 새로운 값
        Timestamp: 2024-01-20 10:15:00
    }
    ```

   이중 쓰기 버퍼(Double Write Buffer):
   장애 시 데이터 손상을 방지하기 위해 데이터를 실제 디스크에 기록하기 전에 버퍼에 먼저 기록하고, 이후 데이터베이스에 적용합니다.

   핵심 원리:
    - WAL은 트랜잭션이 커밋되기 전에 로그에 기록하여 장애 발생 시 복구가 가능하도록 합니다.
    - 이중 쓰기 버퍼는 디스크 손상이나 시스템 중단에도 데이터를 안전하게 유지하는 방법입니다.

<!-- /curriculum-chunk -->

### 데이터베이스 락(Lock)과 격리 수준(Isolation Level)

#### 원문: 데이터베이스 락(Lock)과 격리 수준(Isolation Level)

<!-- curriculum-chunk: sha256=b5b2c64cf37ef19189b8b6b6b77f0114610335009fa46c9426704bdc02ab0fa9 major=database-storage-search-nosql mid=트랜잭션, 락, 격리 sub=데이터베이스 락(Lock)과 격리 수준(Isolation Level) sources=source/interview_questions.md:7760-7805, source/interviews.md:7708-7753 -->

> Source: `source/interview_questions.md:7760-7805`
> Classification reason: transaction/isolation
> Duplicate source aliases: `source/interview_questions.md:7760-7805, source/interviews.md:7708-7753`

##### 데이터베이스 락(Lock)과 격리 수준(Isolation Level)

데이터베이스 락과 격리 수준은 *동시성 제어*와 *데이터 일관성을 유지*하는 데 중요한 역할을 합니다.

- 락(Lock): 데이터베이스 리소스에 대한 동시 접근을 제어하는 메커니즘
- 격리 수준(Isolation Level): 동시에 실행되는 트랜잭션들 사이의 상호작용 정도를 정의
- 트랜잭션(Transaction): 데이터베이스의 상태를 변화시키는 하나의 논리적 작업 단위
- 동시성 제어(Concurrency Control): 동시에 실행되는 트랜잭션들이 데이터베이스의 일관성을 해치지 않도록 보장하는 기법
- 데드락(Deadlock): 두 개 이상의 트랜잭션이 서로가 점유하고 있는 자원을 요구하며 무한정 기다리는 상황

낙관적 잠금과 비관적 잠금

- 낙관적 잠금:
  "낙관적" 잠금은 CAS(compare-and-swap)라는 Couchbase의 값에 달려 있습니다.
  모든 문서에는 불투명한 값인 CAS 값이 있습니다.
  해당 문서가 변경될 때마다 새로운 CAS 값을 얻습니다.
  문서를 업데이트하려고 할 때 작업의 일부로 CAS 값을 전달합니다.
  CAS 값이 일치하면 Couchbase는 작업을 허용합니다.
  일치하지 않으면 작업이 허용되지 않고 대신 오류를 반환합니다.

  예를 들어 두 개의 프로세스가 있다고 가정해 보겠습니다: A와 B.
  A와 B는 모두 문서를 가져오기 위해 Couchbase에 "get" 요청을 합니다.
  Couchbase는 CAS 값과 함께 문서를 반환합니다.
  그런 다음 A와 B는 모두 이전에 받은 CAS 값을 전달하면서 Couchbase에 "set" 요청을 보냅니다.
  처리는 경쟁을 통해 둘 중 하나에서 먼저 발생합니다.
  B의 차례가 되면 문서의 CAS 값이 이미 변경되었으므로 작업이 실패한다고 가정해 보겠습니다.

  이 프로세스가 끝나면 B는 실패하고 플레이어의 검은 레벨 2에 머물러 있습니다.
  B가 성공하도록 하려면 문서를 다시 가져와서 최신 CAS 값을 가져온 다음 다시 시도하는 것이 한 가지 해결책입니다.

  물론 이 솔루션도 다시 실패하고 또 다시 실패할 수 있습니다.
  하지만 이것이 바로 "낙관적"이라고 불리는 이유입니다.
  이 방법은 문서가 심한 경합 상태에 있지 않을 것이고 *결국에는 성공할 것이라고 가정*합니다.
  서버에 의한 실제 잠금은 필요하지 않으며 값만 확인하면 됩니다.

- 비관적 잠금

  비관적 잠금을 사용하여 실제로 잠금을 설정할 수 있습니다.
  이 기능은 여러 문서를 변경하기 위해 문서 그래프를 잠그고 싶을 때 유용할 수 있습니다.

  카우치베이스에는 "GetAndLock"이라는 원자 연산이 있습니다.
  이 연산은 문서와 CAS 값을 반환합니다.
  이 시점에서 문서는 "잠긴" 것으로 간주됩니다.
  다른 프로세스에서 더 이상 잠글 수 없으며 CAS 값만 문서의 잠금을 해제할 수 있습니다.

  또한 GetAndLock을 사용할 때는 시간 제한 기간을 설정해야 합니다.
  타임아웃 기간이 지나면 Couchbase는 자동으로 잠금을 해제합니다.

<!-- /curriculum-chunk -->
