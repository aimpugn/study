# MySQL InnoDB와 PostgreSQL 엔진 Deep Dive

MySQL과 PostgreSQL은 둘 다 SQL을 처리하고 B-tree 인덱스를 쓰며 로그 기반 장애 복구를 합니다. 이 정도만 보면 두 DBMS는 비슷해 보입니다. 하지만 면접에서 중요한 질문은 "둘 다 MVCC를 쓴다"가 아니라, 같은 UPDATE와 SELECT가 내부에서 어느 저장 구조를 바꾸고, 어떤 로그를 남기며, 오래된 버전을 어디에 보관하고, 어느 시점에 정리 비용을 치르는지입니다.

이 문서는 InnoDB와 PostgreSQL을 같은 추상어로 뭉개지 않습니다. InnoDB는 clustered index가 테이블의 본체이고 undo log가 과거 버전을 붙잡습니다. PostgreSQL은 heap table의 tuple version이 테이블 본체이고 transaction id와 vacuum이 생명주기를 관리합니다. 이 차이는 primary key 설계, 보조 인덱스 크기, long transaction 장애, lock 대기, 장애 복구, query plan 해석까지 계속 번집니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [InnoDB: clustered index가 테이블의 중심이다](#innodb-clustered-index가-테이블의-중심이다)
    - [InnoDB undo log: rollback만이 아니라 consistent read의 재료다](#innodb-undo-log-rollback만이-아니라-consistent-read의-재료다)
    - [InnoDB redo log와 buffer pool: dirty page보다 commit log를 먼저 믿는다](#innodb-redo-log와-buffer-pool-dirty-page보다-commit-log를-먼저-믿는다)
    - [DB 프로세스 아래에는 커널, page cache, block I/O가 있다](#db-프로세스-아래에는-커널-page-cache-block-io가-있다)
    - [InnoDB lock: record, gap, next-key를 인덱스와 함께 본다](#innodb-lock-record-gap-next-key를-인덱스와-함께-본다)
    - [PostgreSQL: heap tuple version과 snapshot visibility](#postgresql-heap-tuple-version과-snapshot-visibility)
    - [PostgreSQL vacuum: 삭제가 아니라 생명주기 정리다](#postgresql-vacuum-삭제가-아니라-생명주기-정리다)
    - [PostgreSQL WAL, checkpoint, replication](#postgresql-wal-checkpoint-replication)
    - [PostgreSQL planner와 통계: 좋은 SQL도 잘못된 추정에서 느려진다](#postgresql-planner와-통계-좋은-sql도-잘못된-추정에서-느려진다)
    - [PostgreSQL lock: MVCC가 lock을 없애지는 않는다](#postgresql-lock-mvcc가-lock을-없애지는-않는다)
- [DBMS별 경계](#dbms별-경계)
    - ["InnoDB는 clustered, PostgreSQL은 heap"이 만드는 설계 차이](#innodb는-clustered-postgresql은-heap이-만드는-설계-차이)
    - ["둘 다 WAL"이라는 말의 안전한 범위](#둘-다-wal이라는-말의-안전한-범위)
    - [isolation level을 같은 이름으로만 비교하면 위험하다](#isolation-level을-같은-이름으로만-비교하면-위험하다)
    - [explain을 읽는 관점도 다르다](#explain을-읽는-관점도-다르다)
- [직접 재생해 보기](#직접-재생해-보기)
    - [InnoDB secondary index가 primary key를 품는지 확인하기](#innodb-secondary-index가-primary-key를-품는지-확인하기)
    - [PostgreSQL tuple version과 vacuum 감각 보기](#postgresql-tuple-version과-vacuum-감각-보기)
    - [InnoDB lock 범위 실험](#innodb-lock-범위-실험)
    - [PostgreSQL blocker/waiter 확인](#postgresql-blockerwaiter-확인)
- [면접 꼬리 질문](#면접-꼬리-질문)
    - [InnoDB에서 primary key가 중요한 이유는 무엇인가요?](#innodb에서-primary-key가-중요한-이유는-무엇인가요)
    - [PostgreSQL에서 UPDATE가 많으면 왜 vacuum 이야기가 나오나요?](#postgresql에서-update가-많으면-왜-vacuum-이야기가-나오나요)
    - [InnoDB undo log와 PostgreSQL dead tuple은 같은 것인가요?](#innodb-undo-log와-postgresql-dead-tuple은-같은-것인가요)
    - [MySQL과 PostgreSQL의 WAL/redo 차이를 어떻게 설명하나요?](#mysql과-postgresql의-walredo-차이를-어떻게-설명하나요)
    - [PostgreSQL에서 index-only scan이 항상 heap을 안 보나요?](#postgresql에서-index-only-scan이-항상-heap을-안-보나요)
    - [InnoDB gap lock은 언제 설명해야 하나요?](#innodb-gap-lock은-언제-설명해야-하나요)
    - [PostgreSQL planner가 sequential scan을 골랐으면 무조건 나쁜가요?](#postgresql-planner가-sequential-scan을-골랐으면-무조건-나쁜가요)
- [함정 질문](#함정-질문)
    - ["MVCC면 read lock이 없으니 lock 문제도 거의 없죠?"](#mvcc면-read-lock이-없으니-lock-문제도-거의-없죠)
    - ["InnoDB secondary index는 행 전체를 가지고 있으니 covering index가 자동 아닌가요?"](#innodb-secondary-index는-행-전체를-가지고-있으니-covering-index가-자동-아닌가요)
    - ["PostgreSQL은 heap이라서 primary key 순서 locality가 전혀 의미 없나요?"](#postgresql은-heap이라서-primary-key-순서-locality가-전혀-의미-없나요)
    - ["VACUUM은 DELETE한 공간을 운영체제에 바로 돌려주는 작업인가요?"](#vacuum은-delete한-공간을-운영체제에-바로-돌려주는-작업인가요)
    - ["redo log가 있으면 data page flush는 성능만의 문제인가요?"](#redo-log가-있으면-data-page-flush는-성능만의-문제인가요)
    - ["PostgreSQL SERIALIZABLE이면 모든 충돌을 대기시켜 순서대로 처리하나요?"](#postgresql-serializable이면-모든-충돌을-대기시켜-순서대로-처리하나요)
    - [전이 질문: 같은 SQL 이름 아래 다른 저장 엔진 계약](#전이-질문-같은-sql-이름-아래-다른-저장-엔진-계약)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

InnoDB를 가장 짧게 설명하면 "기본키 B-tree가 곧 테이블이고, 보조 인덱스는 기본키를 다시 따라가며, 과거 버전은 undo log를 통해 읽는다"입니다. InnoDB table은 clustered index를 중심으로 저장됩니다. clustered index의 leaf page에는 행 데이터가 함께 들어 있고, secondary index의 leaf entry에는 보조 키와 함께 clustered index key가 들어갑니다. 그래서 InnoDB에서 기본키는 단순 식별자가 아니라 물리적 접근 경로의 중심입니다. 기본키가 길거나 자주 바뀌면 보조 인덱스와 페이지 재배치 비용까지 흔듭니다.

PostgreSQL을 가장 짧게 설명하면 "table heap에 tuple version을 계속 쌓고, 각 tuple의 xmin/xmax 같은 transaction id로 보이는 버전을 고르며, vacuum이 더 이상 필요 없는 과거 버전을 정리한다"입니다. PostgreSQL의 인덱스는 보통 heap tuple 위치를 가리킵니다. UPDATE는 기존 tuple을 제자리에서 덮어쓰는 동작이라기보다 새 tuple version을 만들고 예전 version의 가시성을 닫는 동작에 가깝습니다. 이 모델은 rollback과 snapshot isolation 설명을 단순하게 해 주지만, dead tuple과 vacuum 지연이라는 운영 비용을 만듭니다.

두 엔진 모두 장애 복구에는 로그가 핵심입니다. InnoDB redo log는 dirty page가 아직 data file에 쓰이지 않았더라도 committed change를 crash recovery 때 되살릴 수 있게 합니다. PostgreSQL WAL도 data page를 먼저 믿지 않고, commit record와 page change log를 기준으로 재생합니다. 다만 "undo/redo/WAL"이라는 이름만 보고 같은 계층으로 대응시키면 틀리기 쉽습니다. InnoDB undo는 MVCC consistent read와 rollback에 깊게 연결되고, PostgreSQL의 오래된 row version은 heap tuple 자체에 남아 있습니다. PostgreSQL WAL은 redo 성격의 로그이며, rollback을 위해 별도 undo segment를 재생하는 구조가 아닙니다. 로그와 복구의 공통 원리는 [WAL, redo, undo, crash recovery, PITR](03-wal-redo-undo-crash-recovery-pitr.md)에서 더 낮은 층위로 이어집니다.

잠금도 겉보기보다 다릅니다. InnoDB는 record lock, gap lock, next-key lock 같은 범위 잠금이 중요합니다. 특히 REPEATABLE READ에서 phantom을 막기 위해 인덱스 record와 그 사이 gap을 함께 잠그는 경우가 있습니다. PostgreSQL은 MVCC 때문에 읽기와 쓰기가 많이 분리되지만, row-level lock, table-level lock, predicate lock, advisory lock이 있으며, isolation level과 index access path에 따라 대기 양상이 달라집니다. "PostgreSQL은 lock이 없다"도 틀리고, "InnoDB는 row lock만 건다"도 틀립니다.

실무 판단은 결국 엔진 경계를 따라야 합니다. InnoDB에서 "PK를 UUID 문자열로 아무렇게나 둬도 괜찮다"는 말은 secondary index leaf에 PK가 반복 저장된다는 사실을 놓친 말일 수 있습니다. PostgreSQL에서 "UPDATE가 적당히 빠르겠지"라고 보면 dead tuple, HOT update 가능 여부, autovacuum 지연, visibility map과 index-only scan의 관계를 놓칠 수 있습니다. 같은 SQL을 보고도 어떤 엔진에서는 page split과 secondary lookup을, 다른 엔진에서는 tuple churn과 vacuum debt를 먼저 의심해야 합니다.

## 먼저 잡아야 할 작은 모델

하나의 은행 계좌 테이블을 기준으로 잡겠습니다.

```sql
CREATE TABLE account (
    id BIGINT PRIMARY KEY,
    owner_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    balance DECIMAL(19, 2) NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_account_status ON account(status);
```

이 테이블에서 `id = 10`인 행을 읽고, `status = 'ACTIVE'` 조건으로 목록을 조회하고, `balance`를 갱신한다고 하자. SQL은 같지만 엔진의 mental model은 다릅니다.

InnoDB에서는 `PRIMARY KEY(id)` B-tree가 테이블의 본체입니다. `id = 10`을 찾으면 clustered index를 내려가 leaf page에서 실제 행을 만납니다. `idx_account_status`는 `status` 값으로 정렬된 별도 B-tree지만 leaf에는 전체 행이 아니라 `status`와 그 행의 primary key가 들어갑니다. 그래서 `status = 'ACTIVE'`로 보조 인덱스를 타면 먼저 matching secondary entry를 찾고, 각 entry가 가진 primary key로 clustered index를 다시 찾아 실제 행을 읽습니다. 이 과정을 보통 back to primary key lookup이라고 이해하면 됩니다.

PostgreSQL에서는 table heap이 행 version을 담는 본체이고, primary key index도 secondary index도 heap tuple 위치를 가리키는 접근 경로입니다. `id = 10`을 찾으면 primary key B-tree가 heap tuple 위치를 알려 주고, PostgreSQL은 해당 heap page의 tuple을 읽은 뒤 현재 transaction snapshot에서 보이는지 판단합니다. `status = 'ACTIVE'`도 index entry에서 heap tuple 위치를 얻고 heap을 확인합니다. index-only scan이 가능하려면 visibility map 같은 추가 조건이 맞아야 합니다. 단순히 "인덱스에 컬럼이 다 있으니 heap을 안 본다"로 끝나지 않습니다.

UPDATE를 작은 그림으로 보면 차이가 더 선명합니다.

```text
InnoDB
  clustered record(id=10, balance=100)
      |
      | UPDATE balance = 90
      v
  clustered record는 새 값으로 바뀌고,
  이전 값은 undo log chain을 통해 consistent read와 rollback에 쓰입니다.

PostgreSQL
  heap tuple A(id=10, balance=100, xmin=old)
      |
      | UPDATE balance = 90
      v
  heap tuple B(id=10, balance=90, xmin=new)가 생기고,
  tuple A는 새 transaction 이후에는 더 이상 보이지 않도록 표시됩니다.
```

같은 `UPDATE account SET balance = 90 WHERE id = 10`을 조금 더 작게 쪼개면, 두 엔진이 비용을 미루는 위치가 다릅니다.

| 질문 | InnoDB 쪽에서 먼저 떠올릴 것 | PostgreSQL 쪽에서 먼저 떠올릴 것 |
| --- | --- | --- |
| 현재 row는 어디에 있는가 | clustered index leaf page의 record | heap page의 tuple version |
| 이전 값은 어디서 읽는가 | undo log chain을 따라 과거 값을 재구성 | heap에 남은 과거 tuple version을 snapshot으로 판정 |
| 나중에 무엇을 정리하는가 | purge가 더 이상 필요 없는 undo/history를 정리 | vacuum이 dead tuple과 visibility/freeze 정보를 정리 |
| 오래 열린 transaction이 막는 것 | undo purge와 history list 축소 | dead tuple 제거와 xid freeze 진행 |

이 표는 "둘 다 MVCC"라는 말을 안전한 출발점으로 바꿔 줍니다. 같은 기능 이름 아래에서도 과거 버전이 저장되는 장소가 다르면, 장애 증상과 튜닝 지표도 달라집니다.

이 그림은 실제 구현을 모두 담지는 않지만, 면접에서 거의 모든 후속 질문의 기준점이 됩니다. InnoDB는 "현재 row record + undo history" 쪽으로 생각하고, PostgreSQL은 "heap 안의 여러 tuple version + snapshot visibility + vacuum" 쪽으로 생각하면 됩니다.

이 차이를 잡으면 buffer/cache도 다르게 보입니다. InnoDB buffer pool은 data page와 index page를 캐시하고, dirty page를 나중에 flush합니다. PostgreSQL은 shared buffers가 database page를 들고 있고, OS page cache와 함께 작동합니다. 둘 다 "메모리에 올려 빠르게 읽는다"는 말은 맞지만, 튜닝 질문에서 `innodb_buffer_pool_size`와 PostgreSQL `shared_buffers`, `effective_cache_size`, checkpoint, WAL flush를 같은 knob처럼 다루면 위험합니다.

transaction id도 작은 모델에서 잡아 두는 편이 좋습니다. InnoDB의 read view는 "내 snapshot에서 어떤 transaction이 active였고, 어떤 version까지 보이는가"를 판단하는 데 쓰입니다. PostgreSQL은 tuple header의 `xmin`, `xmax`와 transaction commit status를 바탕으로 현재 snapshot에서 보이는 tuple인지 판단합니다. 둘 다 MVCC지만, 과거 버전이 저장되는 장소와 정리되는 방식이 다릅니다. snapshot과 가시성 판단 자체가 헷갈리면 먼저 [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md)를 읽고 돌아오는 편이 좋습니다.

잠금은 인덱스와 분리해서 생각하면 자주 틀립니다. InnoDB row lock은 index record에 걸린다고 이해하는 편이 안전합니다. 적절한 index 없이 범위 조건을 갱신하면 더 넓은 범위를 잠그거나 많은 record를 스캔하며 대기를 만들 수 있습니다. PostgreSQL row lock은 heap tuple에 걸리는 동시성 제어와 연결되고, table-level lock은 DDL이나 특정 명령에서 별도로 중요합니다. 두 DBMS 모두 "SQL 조건이 좁다"와 "잠금 범위가 좁다"가 항상 같은 말은 아닙니다. 실제로 어떤 access path를 탔는지가 중요합니다. 이 부분은 [인덱스와 optimizer](04-index-query-optimizer.md), [격리 수준, lock, deadlock](08-isolation-lock-deadlock.md)과 바로 연결됩니다.

## 깊은 메커니즘

### InnoDB: clustered index가 테이블의 중심이다

InnoDB table은 clustered index 중심으로 저장됩니다. primary key가 있으면 그 key로 clustered index를 만들고, primary key가 없으면 unique not null key를 찾고, 그것도 없으면 숨은 row id를 사용합니다. 중요한 점은 clustered index가 "행을 찾기 위한 인덱스 하나"가 아니라 테이블 데이터를 담는 구조라는 것입니다. leaf page에 row data가 함께 들어 있으므로 primary key lookup은 곧 행 접근입니다.

이 구조는 MySQL이 여러 storage engine을 붙일 수 있는 SQL layer를 갖고 있고, InnoDB가 그 아래에서 트랜잭션과 row 저장을 책임지는 엔진이라는 배경과도 맞닿아 있습니다. SQL layer 입장에서는 `SELECT ... WHERE id = ?`라는 논리 질문을 던지지만, InnoDB 입장에서는 그 질문을 B-tree page를 따라 내려가 leaf record를 찾는 물리 작업으로 바꿉니다. clustered index는 "primary key로 자주 찾는 행을 한 번에 만나게 한다"는 장점이 있는 대신, primary key 선택이 모든 보조 인덱스와 page 배치에 영향을 주는 저장 계약이 됩니다.

```text
SQL layer question
  WHERE id = 10
        |
        v
InnoDB clustered tree
  root page -> branch page -> leaf page(id=10 + row payload)
        |
        v
row data returned
```

이 흐름을 보면 primary key가 단순 논리 제약이 아니라 "행이 어느 B-tree leaf에 살 것인가"를 정하는 주소 체계에 가깝다는 점이 드러납니다.

이 구조는 primary key 설계를 운영 비용으로 끌어올립니다. 예를 들어 `BIGINT AUTO_INCREMENT`는 대체로 짧고 증가 방향이 일정해 clustered page append 성격이 강합니다. 반대로 길고 랜덤한 문자열 primary key는 clustered index page를 더 많이 차지하고 page split 가능성을 높이며, 모든 secondary index entry에도 반복 저장됩니다. UUID를 쓰면 무조건 나쁘다는 뜻은 아닙니다. 다만 InnoDB에서 UUID primary key는 "식별자 선택"이 아니라 "모든 보조 인덱스 leaf entry의 payload 선택"이기도 합니다.

secondary index lookup은 두 단계가 될 수 있습니다. `idx_account_status(status)`로 `status = 'ACTIVE'` entry를 찾으면 leaf에는 status와 primary key가 있습니다. 필요한 컬럼이 secondary index 안에 모두 있으면 covering index처럼 clustered lookup을 줄일 수 있지만, `balance`, `owner_name` 같은 컬럼이 필요하면 primary key로 clustered index를 다시 찾아야 합니다. 그래서 InnoDB explain에서 secondary index를 탄다고 항상 row data access가 없다고 생각하면 안 됩니다.

작은 예로 보면 보조 인덱스 leaf가 무엇을 들고 있는지 더 분명합니다.

```text
clustered index PRIMARY(id)
  leaf:
    id=10, owner_name='Kim', status='ACTIVE', balance=90
    id=20, owner_name='Lee', status='LOCKED', balance=30

secondary index idx_account_status(status)
  leaf:
    status='ACTIVE', primary key=10
    status='LOCKED', primary key=20
```

`SELECT id, status FROM account WHERE status = 'ACTIVE'`는 보조 인덱스 leaf만으로 답할 수 있을 가능성이 있습니다. 하지만 `SELECT owner_name, balance FROM account WHERE status = 'ACTIVE'`는 보조 인덱스에서 `id=10`을 얻은 뒤 clustered index를 다시 내려가야 합니다. 이 두 번째 탐색이 많아지면 random I/O와 buffer pool pressure가 커질 수 있습니다. 그래서 InnoDB에서 covering index는 "인덱스를 탔다"보다 한 단계 더 좁은 판단입니다.

### InnoDB undo log: rollback만이 아니라 consistent read의 재료다

InnoDB undo log는 transaction rollback에 쓰이지만, 그것만으로 설명하면 절반만 맞습니다. MVCC consistent read도 undo chain을 통해 과거 version을 재구성합니다. 어떤 transaction이 오래 열려 있으면 purge가 더 이상 필요 없는 undo record를 제거하지 못하고 history list가 길어질 수 있습니다. 이때 새 쿼리 자체가 단순 SELECT라도 오래된 snapshot을 유지하는 connection 때문에 undo 정리가 밀리고, storage와 성능에 영향을 줄 수 있습니다.

예를 들어 transaction A가 `SELECT`를 실행한 뒤 commit하지 않고 오래 대기합니다. 그 사이 transaction B, C, D가 같은 table을 계속 update합니다. InnoDB는 최신 record를 바꿔 가면서도 A가 처음 본 snapshot을 재구성할 수 있도록 undo history를 남겨 둬야 합니다. A가 끝나기 전에는 일부 undo를 지우기 어렵습니다. 면접에서 "long transaction은 왜 나쁜가"를 묻는다면, 단순히 lock을 오래 잡아서만이 아니라 MVCC history 정리와 purge 지연까지 말해야 합니다.

### InnoDB redo log와 buffer pool: dirty page보다 commit log를 먼저 믿는다

InnoDB는 buffer pool에 data/index page를 올리고, 변경된 page를 즉시 data file에 모두 쓰지 않습니다. 대신 redo log에 변경 내용을 남기고 commit durability를 확보합니다. crash가 나면 data file에는 예전 page가 남아 있을 수 있지만, redo log를 재생해 committed change를 복구할 수 있습니다. 이 원리가 write-ahead logging의 핵심입니다. data page보다 log가 먼저 durable해야 합니다.

buffer pool은 단순 LRU 캐시가 아닙니다. InnoDB는 midpoint insertion 같은 정책으로 대량 scan이 hot page를 밀어내는 문제를 줄입니다. 운영에서 buffer pool hit ratio만 보며 안심하면 부족합니다. dirty page 비율, checkpoint age, redo log pressure, flush rate, read-ahead, adaptive hash index 사용 여부처럼 write path와 read cache가 만나는 지점도 함께 봐야 합니다.

Dirty page와 redo log의 시간축도 따로 그려야 합니다.

```text
T1 update
  buffer pool page(id=10)가 dirty가 됨
  redo log buffer에 변경 기록 생성

T2 commit
  redo log가 write/flush 정책에 따라 durable boundary로 내려감
  data page는 아직 tablespace에 안 내려갔을 수 있음

T3 background flush/checkpoint
  dirty data page가 나중에 tablespace로 내려감
  checkpoint가 recovery 시작점을 앞으로 당김

T4 crash recovery
  tablespace page가 낡았으면 redo log를 재생해 committed change를 복원
```

이 trace에서 commit의 핵심은 data page 전체를 즉시 제자리 파일에 쓰는 것이 아니라, 장애 뒤 재생할 수 있는 로그를 먼저 믿는다는 점입니다. 그래서 redo log pressure가 높으면 단순 로그 파일 문제가 아니라 dirty page flush와 checkpoint가 따라오지 못한다는 신호일 수 있습니다.

장애 복구 질문에서 undo와 redo를 섞지 않는 것이 중요합니다. redo는 committed change를 되살리는 방향이고, undo는 transaction rollback과 old version reconstruction에 쓰입니다. crash recovery 때 committed transaction의 page change는 redo로 반영되고, incomplete transaction은 undo 쪽 정보로 정리됩니다. "redo는 다시 하고 undo는 되돌린다"는 말은 출발점으로는 좋지만, InnoDB의 undo가 평상시 consistent read에도 쓰인다는 점까지 붙여야 실제 엔진 설명이 됩니다.

### DB 프로세스 아래에는 커널, page cache, block I/O가 있다

InnoDB buffer pool과 PostgreSQL shared buffers는 DBMS 내부 cache입니다. 그 아래에는 OS page cache, filesystem, block layer, device driver, storage device cache가 이어집니다. 그래서 "DB가 page를 flush했다"는 문장을 읽을 때는 어떤 계층의 flush인지 확인해야 합니다. DB가 자기 buffer에서 OS로 write했는지, OS가 page cache의 dirty page를 writeback했는지, storage device의 volatile cache까지 비웠는지는 서로 다릅니다.

```text
DB process
  -> InnoDB buffer pool / PostgreSQL shared buffers
  -> write() 또는 pwrite()로 kernel에 전달
  -> kernel page cache의 page가 dirty 상태가 됨
  -> filesystem writeback이 block I/O request를 만듦
  -> block layer가 scheduler/driver queue로 request를 보냄
  -> SSD/HDD controller cache 또는 device media
  -> fsync/fdatasync/flush/FUA가 durable boundary를 끌어올림
```

Buffered I/O에서는 `write()`가 끝나도 데이터가 아직 page cache에 있을 수 있습니다. Linux VFS 문서는 page가 더러워지면 `PG_Dirty` 상태가 되고, writeback 중에는 `PG_Writeback` 상태로 관리된다고 설명합니다. Block layer에서는 BIO가 request로 묶여 software queue와 hardware dispatch queue를 지나 device driver로 내려갑니다. 이 경로가 막히면 DB 관점에서는 WAL write, log file sync, checkpoint latency, dirty page flush 지연처럼 보이고, OS 관점에서는 device utilization, await, queue depth, flush request 증가처럼 보일 수 있습니다.

이 층을 알면 InnoDB와 PostgreSQL의 cache 튜닝도 더 안전해집니다. InnoDB는 buffer pool을 크게 잡아 data/index page를 DBMS가 직접 관리하려는 성격이 강하고, PostgreSQL은 shared buffers와 OS page cache가 함께 읽기 cache를 형성합니다. `O_DIRECT`, `fsync`, `fdatasync`, `wal_sync_method`, `innodb_flush_method` 같은 설정은 "빠른가 느린가"만의 knob가 아니라 어떤 cache를 우회하거나 어떤 durable boundary를 기다리는지 결정하는 설정입니다.

Background worker도 이 경계에서 같이 봐야 합니다.

| 계층 | InnoDB에서 흔히 보는 흐름 | PostgreSQL에서 흔히 보는 흐름 | 운영에서 보이는 증상 |
| --- | --- | --- | --- |
| log write | redo log write/flush | WAL write/flush, WAL writer | commit latency, log sync wait |
| dirty page flush | page cleaner, checkpoint pressure | bgwriter/checkpointer, checkpoint | write burst, dirty buffer 증가 |
| old version cleanup | purge thread가 undo history 정리 | autovacuum이 dead tuple/freeze 정리 | long transaction 뒤 storage 증가 |
| OS/device | filesystem writeback, device flush | filesystem writeback, device flush | `await`, queue depth, flush latency |

제품마다 내부 스레드 이름과 세부 구현은 다르지만, 면접에서는 "commit이 기다리는 로그 경계", "나중에 내려가는 data page 경계", "오래된 버전 정리 경계"를 분리해 말하는 것이 중요합니다. 이 세 경계를 한 단어 `flush`로 합치면 어떤 지표를 봐야 하는지 흐려집니다.

### InnoDB lock: record, gap, next-key를 인덱스와 함께 본다

InnoDB row-level lock은 보통 index record를 기준으로 설명합니다. unique index로 정확히 한 row를 찾는 update와, range condition으로 여러 index record를 훑는 update는 잠금 범위가 다릅니다. REPEATABLE READ에서 range scan은 phantom을 막기 위해 next-key lock, 즉 record lock과 gap lock이 결합된 형태를 사용할 수 있습니다. gap lock은 실제 존재하는 row가 아니라 index record 사이의 빈 공간에 대한 insert를 막습니다.

예를 들어 `WHERE amount BETWEEN 100 AND 200 FOR UPDATE` 같은 범위 잠금은 이미 존재하는 record만 막는 것이 아니라, 그 범위에 새 row가 끼어드는 것도 막아야 phantom을 막을 수 있습니다. 이때 적절한 index가 없으면 엔진은 더 넓게 scan하고 더 많은 lock을 잡을 수 있습니다. 그래서 InnoDB lock troubleshooting에서는 `SHOW ENGINE INNODB STATUS`, performance schema lock table, 실행 계획, isolation level을 함께 봅니다.

deadlock은 "코드가 잘못됐다"라는 단순 결론으로 끝내면 안 됩니다. InnoDB는 deadlock을 감지하고 한 transaction을 victim으로 rollback할 수 있습니다. 애플리케이션은 deadlock을 일시적 충돌로 보고 재시도할 수 있어야 하며, 동시에 lock acquisition order를 안정화하고 index를 맞춰 conflict set을 줄여야 합니다. 재시도만 넣고 원인 범위를 줄이지 않으면 부하가 높아질수록 같은 deadlock이 반복됩니다.

### PostgreSQL: heap tuple version과 snapshot visibility

PostgreSQL table은 heap file에 tuple을 저장합니다. UPDATE는 많은 경우 기존 tuple을 덮어쓰기보다 새 tuple version을 만들고, 이전 tuple에는 더 이상 새 snapshot에서 보이지 않도록 `xmax` 같은 정보를 남깁니다. INSERT tuple에는 그 tuple을 만든 transaction id가 `xmin`으로 남고, DELETE나 UPDATE로 사라지는 쪽에는 종료 transaction id가 남습니다. SELECT는 자신의 snapshot과 transaction commit 상태를 보고 어떤 tuple이 보이는지 판단합니다.

PostgreSQL의 heap 중심 모델은 "읽는 transaction이 자기 snapshot에서 보던 row를 계속 볼 수 있어야 한다"는 MVCC 요구를 매우 직접적인 방식으로 풀어냅니다. 현재 row를 제자리에서 덮어쓰면 오래된 snapshot이 볼 과거 값이 사라집니다. PostgreSQL은 과거 값을 별도 undo segment에서 재구성하기보다, heap 안에 새 tuple version을 추가하고 예전 tuple의 가시성만 닫는 쪽을 택합니다. 이 선택은 reader/writer 충돌을 줄이는 대신, 언젠가 필요 없어진 tuple version을 찾아 치워야 한다는 vacuum 비용을 남깁니다.

```text
old snapshot S1 starts
  sees tuple A(balance=100)

writer commits update
  tuple A: old version, old snapshot에는 아직 보일 수 있음
  tuple B: new version(balance=90), new snapshot에 보임

after S1 ends
  tuple A는 더 이상 필요하지 않음
  vacuum이 회수 후보로 볼 수 있음
```

따라서 PostgreSQL에서 UPDATE가 많은 테이블을 볼 때는 "쓰기 성공"만 보지 않고, old snapshot이 얼마나 오래 남는지와 vacuum이 그 뒤처리를 따라잡는지를 같이 봐야 합니다.

이 모델의 장점은 reader와 writer가 서로 덜 막힌다는 것입니다. 오래된 snapshot을 가진 SELECT는 자신에게 보이는 과거 tuple version을 읽으면 됩니다. writer도 새 tuple version을 만들 수 있습니다. 하지만 비용은 dead tuple입니다. 더 이상 어떤 snapshot에서도 필요 없는 tuple version은 vacuum이 정리해야 합니다. vacuum이 늦어지면 table bloat, index bloat, visibility map 지연, autovacuum 부하가 생깁니다.

PostgreSQL을 설명할 때 HOT update도 중요한 경계입니다. HOT은 Heap-Only Tuple의 약자로, 인덱스 컬럼이 바뀌지 않고 같은 page 안에 새 tuple version을 둘 수 있을 때 인덱스 새 entry를 만들지 않고 update chain을 이어 성능 비용을 줄이는 최적화입니다. 하지만 page 여유 공간이 없거나 인덱스에 포함된 컬럼을 바꾸면 HOT이 어려워집니다. 그래서 PostgreSQL update-heavy table에서는 fillfactor, 인덱스 설계, vacuum 상태가 같이 중요해집니다.

### PostgreSQL vacuum: 삭제가 아니라 생명주기 정리다

PostgreSQL vacuum은 단순히 "삭제된 행을 지운다"가 아닙니다. 더 이상 보이지 않는 dead tuple을 회수하고, visibility map을 갱신해 index-only scan이 가능하도록 돕고, transaction id wraparound를 막기 위해 freeze를 수행합니다. autovacuum이 보이지 않게 돌지만, long transaction이나 replication slot, 오래된 snapshot이 있으면 vacuum이 제거할 수 있는 범위가 막힙니다.

면접에서 "PostgreSQL에서 오래 열린 transaction이 왜 문제인가"를 받으면, lock만 말하면 부족합니다. 오래 열린 snapshot은 vacuum이 old tuple을 지우지 못하게 만들고, dead tuple이 쌓여 table scan과 index scan 비용을 올립니다. 심하면 xid wraparound 방지를 위한 aggressive vacuum이 필요해지고, 운영 부하가 갑자기 커질 수 있습니다.

visibility map은 PostgreSQL의 성능 설명에서 자주 빠집니다. index-only scan은 인덱스에 필요한 컬럼이 있다는 조건만으로 끝나지 않습니다. heap page의 모든 tuple이 현재 snapshot에서 보인다는 표시가 있어야 heap 확인을 생략할 수 있습니다. vacuum은 이 표시를 관리합니다. 따라서 vacuum이 지연되면 같은 query가 index-only scan처럼 보여도 실제 heap fetch가 많아질 수 있습니다.

Vacuum을 "삭제 파일 정리"가 아니라 tuple 생명주기 정리로 보면 다음 순서가 보입니다.

```text
T1 INSERT
  tuple A: xmin=10, xmax=0

T2 UPDATE
  tuple A: xmax=20
  tuple B: xmin=20, xmax=0

T3 오래된 snapshot 존재
  어떤 transaction은 아직 tuple A를 볼 수 있음
  vacuum은 tuple A를 마음대로 제거할 수 없음

T4 모든 관련 snapshot 종료
  tuple A는 더 이상 필요 없음
  vacuum이 공간 재사용 가능 표시, visibility map/freeze 정보 갱신
```

이 순서를 알면 `VACUUM`이 곧 파일 크기를 줄이는 명령이 아니라는 점도 자연스럽게 이해됩니다. 일반 vacuum은 보통 table 안의 공간을 재사용 가능하게 만들고, 운영체제 파일 크기를 줄이려면 더 강한 작업이 필요할 수 있습니다. 그 작업은 잠금과 rewrite 비용이 커질 수 있으므로 장애 대응 중 즉흥적으로 선택하면 위험합니다.

### PostgreSQL WAL, checkpoint, replication

PostgreSQL WAL은 data file 변경보다 먼저 durable하게 기록되는 로그입니다. crash 후에는 checkpoint 이후 WAL을 재생해 data directory를 consistent state로 되돌립니다. commit durability는 WAL flush와 관련이 깊고, `synchronous_commit` 설정은 commit 응답 시점과 WAL flush/replica 확인 경계를 바꿉니다. "commit됐는데 서버가 죽으면 어떻게 되나"라는 질문은 WAL flush 정책과 함께 답해야 합니다.

WAL이 필요한 이유는 data page를 매 commit마다 완벽하게 제자리 파일에 써 내려가면 쓰기 비용이 너무 커지고, 장애 순간에는 data file 일부만 바뀐 애매한 상태가 생길 수 있기 때문입니다. 로그를 먼저 순차적으로 남기면 DB는 "어떤 변경이 committed였는가"라는 사실을 작고 재생 가능한 기록으로 보존합니다. 장애 뒤에는 data file을 그대로 믿지 않고, 마지막 checkpoint 이후 WAL record를 따라가며 page를 다시 앞으로 밀어 올립니다. 이 구조는 성능 최적화이면서 복구 절차의 기준선입니다.

```text
commit path
  row/page changes in memory
  -> WAL record durable
  -> client commit ack
  -> data page flush는 나중에 가능

crash recovery path
  last checkpoint에서 시작
  -> WAL record replay
  -> committed change를 data page에 다시 반영
```

이 흐름 때문에 checkpoint는 "모든 것을 디스크에 쓰는 순간"이라기보다, crash recovery가 어디서부터 WAL을 다시 읽으면 되는지 줄여 주는 회계 지점으로 이해하는 편이 안전합니다.

PostgreSQL physical replication도 WAL을 중심으로 움직입니다. primary에서 생성된 WAL record를 standby가 받아 재생합니다. replication lag는 하나의 숫자처럼 보이지만 실제로는 send, write, flush, replay 지연을 나눠 볼 수 있습니다. primary에서는 commit이 끝났는데 standby query에서는 아직 안 보이는 상황이 생길 수 있습니다. 읽기 replica를 쓰는 애플리케이션에서는 read-your-writes 보장을 별도로 설계해야 합니다.

checkpoint는 WAL과 data page 사이의 회계 정리 지점입니다. checkpoint가 너무 자주 일어나면 write burst가 커질 수 있고, 너무 늦으면 crash recovery 시간이 길어질 수 있습니다. PostgreSQL 운영 질문에서 checkpoint warning, WAL volume, bgwriter/checkpointer activity, `pg_stat_wal`, `pg_stat_bgwriter`를 같이 보는 이유가 여기에 있습니다.

### PostgreSQL planner와 통계: 좋은 SQL도 잘못된 추정에서 느려진다

PostgreSQL planner는 table 통계, column 통계, correlation, selectivity 추정, cost parameter를 사용해 plan을 고릅니다. 같은 SQL이라도 데이터 분포가 바뀌거나 통계가 오래되면 nested loop, hash join, merge join, index scan, sequential scan 선택이 달라집니다. 그래서 PostgreSQL slow query 분석은 `EXPLAIN (ANALYZE, BUFFERS)`에서 estimated rows와 actual rows의 차이를 먼저 봅니다.

MySQL도 optimizer와 statistics가 있지만, PostgreSQL은 planner 설명과 통계 튜닝이 면접에서 특히 자주 등장합니다. `ANALYZE`가 통계를 갱신하고, extended statistics가 다중 컬럼 상관관계를 도울 수 있으며, `work_mem`은 sort/hash 작업의 memory 경계를 바꿉니다. 단순히 "인덱스를 추가한다"보다 "왜 planner가 이 plan을 선택했는지"를 설명하는 능력이 더 중요합니다.

### PostgreSQL lock: MVCC가 lock을 없애지는 않는다

PostgreSQL에서는 일반 SELECT가 UPDATE를 막지 않고 UPDATE도 SELECT를 많이 막지 않습니다. 하지만 row lock은 분명히 있고, `SELECT ... FOR UPDATE`, UPDATE/DELETE, foreign key check, DDL은 lock 대기를 만들 수 있습니다. table-level lock mode도 다양하고, `ALTER TABLE` 같은 DDL은 강한 lock을 요구할 수 있습니다. isolation level을 SERIALIZABLE로 올리면 predicate lock과 serialization failure를 이해해야 합니다.

PostgreSQL lock troubleshooting에서는 `pg_stat_activity`, `pg_locks`, blocker/waiter 관계, wait_event, transaction age를 함께 봅니다. 한 session이 "idle in transaction" 상태로 오래 남으면 lock뿐 아니라 vacuum 지연까지 만들 수 있습니다. 이것이 PostgreSQL 운영에서 application boundary가 중요한 이유입니다. DBMS 내부 문제처럼 보이는 현상이 사실은 connection을 반납하지 않거나 transaction을 닫지 않은 애플리케이션 문제일 수 있습니다.

## DBMS별 경계

이 절은 이 문서의 핵심입니다. 같은 용어를 쓰더라도 두 DBMS에서 어디까지 같은지, 어디서부터 다르게 판단해야 하는지 분리해야 합니다.

| 주제 | MySQL InnoDB | PostgreSQL | 면접에서 안전한 말 |
| --- | --- | --- | --- |
| 테이블 본체 | clustered index leaf가 row data를 담는다 | heap table이 tuple version을 담는다 | InnoDB는 PK 접근 경로가 저장 구조의 중심이고, PostgreSQL은 heap tuple 생명주기가 중심이다 |
| 보조 인덱스 | secondary index leaf에 clustered key가 들어간다 | index entry가 heap tuple 위치를 가리킨다 | InnoDB는 secondary lookup 뒤 clustered lookup을, PostgreSQL은 heap visibility 확인을 생각한다 |
| UPDATE | current record 변경과 undo history가 연결된다 | 새 tuple version 생성과 old tuple 종료 표시가 중심이다 | InnoDB는 undo/purge, PostgreSQL은 dead tuple/vacuum 비용을 본다 |
| MVCC 과거 버전 | undo log chain으로 재구성한다 | heap에 남은 tuple version과 transaction id로 판단한다 | "둘 다 MVCC" 뒤에 과거 버전 위치를 반드시 말한다 |
| 정리 작업 | purge가 불필요한 undo/history를 정리한다 | vacuum/autovacuum이 dead tuple과 xid freeze를 관리한다 | long transaction은 두 엔진 모두 나쁘지만 증상과 지표가 다르다 |
| 장애 복구 | redo log, doublewrite buffer, checkpoint, dirty page flush가 중요하다 | WAL, checkpoint, full page write, replay가 중요하다 | data page보다 log를 먼저 믿는 원리는 같지만 구현 경계는 다르다 |
| 잠금 | record/gap/next-key lock과 index access path가 중요하다 | row/table/predicate/advisory lock과 snapshot conflict가 중요하다 | WHERE 조건이 아니라 실제 access path와 isolation level로 잠금 범위를 본다 |
| 읽기 성능 | buffer pool, covering index, clustered locality가 중요하다 | shared buffers, OS cache, visibility map, planner statistics가 중요하다 | cache hit만 보지 말고 heap fetch, dirty page, 통계 추정을 같이 본다 |
| PK 설계 | 모든 secondary index에 PK가 붙으므로 길이와 순서성이 비용이다 | primary key index는 heap 접근 경로 중 하나다 | InnoDB PK는 저장 구조 비용이고, PostgreSQL PK는 강한 논리 제약과 접근 경로다 |
| bloat 성격 | page split, undo history, fragmentation, purge 지연이 중요하다 | dead tuple, index bloat, autovacuum 지연이 중요하다 | "공간이 커졌다"를 같은 원인으로 단정하지 않는다 |

### "InnoDB는 clustered, PostgreSQL은 heap"이 만드는 설계 차이

InnoDB에서 primary key를 설계할 때는 짧고 안정적이며 lookup locality가 좋은지를 봅니다. 물론 비즈니스 natural key가 항상 나쁜 것은 아니지만, 길고 변경 가능한 natural key를 clustered primary key로 쓰면 모든 secondary index가 그 key를 품습니다. 운영 중 key 변경이 필요하면 clustered index에서 row 이동과 secondary index 갱신 비용이 커질 수 있습니다. 그래서 InnoDB에서는 surrogate numeric key를 primary key로 두고 business unique constraint를 별도로 두는 설계가 자주 합리적입니다.

PostgreSQL에서는 primary key가 heap의 물리 정렬을 자동으로 지배하지 않습니다. `CLUSTER` 명령이나 BRIN, fillfactor, partitioning 같은 별도 선택지가 있습니다. PostgreSQL에서 랜덤 UUID primary key가 비용이 없는 것은 아니지만, InnoDB처럼 모든 secondary index leaf가 clustered key를 payload로 들고 다니는 구조와는 다릅니다. 대신 PostgreSQL은 update churn과 vacuum, index bloat를 더 민감하게 봅니다.

### "둘 다 WAL"이라는 말의 안전한 범위

WAL은 write-ahead logging, 즉 data page보다 log를 먼저 durable하게 남기는 원리입니다. 이 원리 수준에서는 MySQL InnoDB redo log와 PostgreSQL WAL을 비교할 수 있습니다. 하지만 PostgreSQL WAL이 곧 InnoDB undo까지 포함한다고 말하면 틀립니다. PostgreSQL은 heap tuple version과 transaction status로 visibility를 판단하고, WAL은 변경 재생과 replication의 중심입니다. InnoDB는 redo log와 undo log가 서로 다른 목적을 갖고, undo는 평상시 consistent read에도 쓰입니다.

장애 복구 질문에서 좋은 답은 "둘 다 WAL 계열 로그가 있으니 복구됩니다"가 아닙니다. 좋은 답은 먼저 commit 시점에 어떤 log가 durable해야 하는지 말하고, crash 후 data page와 log가 불일치할 수 있음을 인정하고, recovery가 log를 기준으로 page를 앞으로 재생하거나 미완료 transaction을 정리한다고 설명합니다. 그 다음 엔진별로 InnoDB redo/undo와 PostgreSQL WAL/checkpoint/replay의 경계를 나눕니다.

### isolation level을 같은 이름으로만 비교하면 위험하다

MySQL InnoDB의 기본 isolation level은 흔히 REPEATABLE READ로 운영되고, PostgreSQL의 기본은 READ COMMITTED입니다. 이름만 비교해 "MySQL이 더 강하다"라고 끝내면 위험합니다. InnoDB REPEATABLE READ는 consistent read snapshot과 locking read의 차이, gap/next-key lock 동작을 같이 봐야 합니다. PostgreSQL READ COMMITTED는 statement마다 snapshot이 바뀌므로 같은 transaction 안에서도 두 SELECT 결과가 달라질 수 있습니다. PostgreSQL REPEATABLE READ는 transaction-level snapshot을 제공하지만, phantom 처리와 serialization anomaly 경계는 SERIALIZABLE과 구분해야 합니다.

면접에서 isolation 질문을 받으면 세 문장으로 정리하는 편이 좋습니다. 첫째, read snapshot이 transaction 단위인지 statement 단위인지 말합니다. 둘째, locking read와 plain select를 구분합니다. 셋째, phantom, write skew, serialization failure를 DBMS가 어떤 방식으로 다루는지 예시로 말합니다. 이렇게 답하면 "격리 수준 이름 암기"가 아니라 실제 동작 경계를 설명할 수 있습니다.

### explain을 읽는 관점도 다르다

InnoDB explain에서 secondary index를 탔다면 다음 질문은 "covering인가, clustered lookup이 얼마나 발생하는가, range scan이 lock 범위를 넓히는가"입니다. rows estimate, filtered, Extra의 `Using index`, `Using where`, `Using temporary`, `Using filesort`를 보고 access path와 후속 비용을 봅니다. 단, explain은 추정이며 실제 runtime 지표와 같이 봐야 합니다.

PostgreSQL explain에서는 estimated rows와 actual rows 차이, buffers hit/read/dirtied, heap fetches, sort/hash memory spill, loop count를 봅니다. PostgreSQL의 planner는 통계 기반 선택을 많이 하므로 "왜 sequential scan을 했나"를 무조건 실패로 보지 않습니다. table이 작거나 selectivity가 낮거나 correlation이 맞지 않으면 sequential scan이 합리적일 수 있습니다. 좋은 답은 plan node 이름보다 planner가 본 비용 모델과 실제 관측 차이를 설명합니다.

## 직접 재생해 보기

아래 실험은 운영 DB에서 실행하지 않습니다. 로컬 disposable database에서 작은 table로 확인하는 용도입니다. 목적은 숫자 하나를 외우는 것이 아니라, 같은 SQL이 엔진별로 다른 내부 비용을 만든다는 감각을 얻는 것입니다.

### InnoDB secondary index가 primary key를 품는지 확인하기

MySQL 8.4 또는 호환 환경에서 InnoDB table을 만듭니다.

```sql
CREATE DATABASE IF NOT EXISTS engine_lab;
USE engine_lab;

DROP TABLE IF EXISTS account_innodb;
CREATE TABLE account_innodb (
    id BIGINT NOT NULL,
    owner_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    balance DECIMAL(19, 2) NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    KEY idx_status (status)
) ENGINE=InnoDB;

INSERT INTO account_innodb
SELECT seq, CONCAT('user-', seq), IF(seq % 10 = 0, 'LOCKED', 'ACTIVE'), seq * 10, NOW()
FROM (
    SELECT 1 seq UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5
) s;

EXPLAIN SELECT id, status FROM account_innodb WHERE status = 'ACTIVE';
EXPLAIN SELECT id, status, balance FROM account_innodb WHERE status = 'ACTIVE';
```

첫 번째 query는 secondary index만으로 필요한 컬럼을 만족할 가능성이 큽니다. 두 번째 query는 `balance`가 secondary index에 없으므로 clustered index lookup이 필요할 수 있습니다. PASS 신호는 explain에서 access key가 `idx_status`로 잡히고, 필요한 컬럼에 따라 Extra와 접근 비용이 달라지는 것을 관찰하는 것입니다. FAIL 신호는 index를 타지 않거나 table이 너무 작아 차이가 보이지 않는 경우입니다. 그때는 row 수를 늘리고 `EXPLAIN ANALYZE`가 가능한 버전이면 실제 실행 통계까지 봅니다.

### PostgreSQL tuple version과 vacuum 감각 보기

현재 사용하는 PostgreSQL 버전 또는 로컬 disposable PostgreSQL에서 실험합니다.

```sql
CREATE SCHEMA IF NOT EXISTS engine_lab;
SET search_path = engine_lab;

DROP TABLE IF EXISTS account_pg;
CREATE TABLE account_pg (
    id BIGINT PRIMARY KEY,
    owner_name TEXT NOT NULL,
    status TEXT NOT NULL,
    balance NUMERIC(19, 2) NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

INSERT INTO account_pg VALUES
(1, 'user-1', 'ACTIVE', 100.00, now());

SELECT xmin, xmax, ctid, * FROM account_pg WHERE id = 1;

UPDATE account_pg
SET balance = balance - 10, updated_at = now()
WHERE id = 1;

SELECT xmin, xmax, ctid, * FROM account_pg WHERE id = 1;
```

`xmin`, `xmax`, `ctid`는 내부 설명을 위한 관찰 도구입니다. application query가 이 값에 의존해서는 안 됩니다. UPDATE 뒤에 `xmin`이나 `ctid`가 바뀌는 것을 보면 "같은 logical row가 새 tuple version으로 표현될 수 있다"는 감각을 얻을 수 있습니다. 이어서 `VACUUM (VERBOSE, ANALYZE) account_pg;`를 실행하면 vacuum이 dead tuple과 통계를 다루는 운영 작업임을 확인할 수 있습니다.

더 확실히 보려면 두 세션을 사용합니다.

```sql
-- session A
BEGIN;
SELECT * FROM engine_lab.account_pg WHERE id = 1;

-- session B
UPDATE engine_lab.account_pg SET balance = balance + 1 WHERE id = 1;
VACUUM (VERBOSE) engine_lab.account_pg;

-- session A
COMMIT;
```

session A가 오래 열린 상태에서는 session B의 vacuum이 모든 과거 version을 마음대로 제거할 수 없습니다. PASS 신호는 오래 열린 transaction이 vacuum 정리에 영향을 준다는 점을 통계와 verbose output에서 확인하는 것입니다. FAIL 신호는 autocommit 상태로만 실험해서 snapshot 유지 효과가 사라지는 것입니다.

### InnoDB lock 범위 실험

두 connection에서 range locking read를 실행합니다.

```sql
-- session A
START TRANSACTION;
SELECT * FROM account_innodb
WHERE id BETWEEN 10 AND 20
FOR UPDATE;

-- session B
START TRANSACTION;
INSERT INTO account_innodb(id, owner_name, status, balance, updated_at)
VALUES (15, 'user-15', 'ACTIVE', 150.00, NOW());
```

이미 row가 없거나 데이터가 작으면 원하는 대기가 보이지 않을 수 있습니다. 핵심은 InnoDB가 isolation level과 index range에 따라 record뿐 아니라 gap까지 보호할 수 있다는 점입니다. 실험을 제대로 보려면 해당 범위 주변 데이터를 넣고, session B가 대기하는 동안 `SHOW ENGINE INNODB STATUS` 또는 performance schema lock view를 확인합니다.

### PostgreSQL blocker/waiter 확인

PostgreSQL에서는 row lock 대기를 다음처럼 볼 수 있습니다.

```sql
-- session A
BEGIN;
UPDATE engine_lab.account_pg
SET balance = balance + 10
WHERE id = 1;

-- session B
BEGIN;
UPDATE engine_lab.account_pg
SET balance = balance - 10
WHERE id = 1;

-- session C
SELECT pid, state, wait_event_type, wait_event, query
FROM pg_stat_activity
WHERE datname = current_database();
```

PASS 신호는 session B가 lock 대기 상태로 보이고, blocker session을 찾을 수 있는 것입니다. 여기서 중요한 결론은 PostgreSQL이 MVCC를 쓰더라도 같은 row를 동시에 갱신하는 write-write conflict는 사라지지 않는다는 점입니다.

## 면접 꼬리 질문

### InnoDB에서 primary key가 중요한 이유는 무엇인가요?

InnoDB의 primary key는 단순한 unique constraint가 아니라 clustered index의 key입니다. clustered index leaf에는 row data가 들어 있고, secondary index leaf에는 secondary key와 clustered key가 들어갑니다. 그래서 primary key가 길면 secondary index도 같이 커지고, random key는 clustered page split과 cache locality에 영향을 줄 수 있습니다. PostgreSQL에서도 primary key 설계가 중요하지만, InnoDB처럼 table storage의 중심이라는 점이 특히 강합니다.

### PostgreSQL에서 UPDATE가 많으면 왜 vacuum 이야기가 나오나요?

PostgreSQL UPDATE는 logical row를 제자리에서 단순 덮어쓰기보다 새 tuple version을 만들고 예전 version을 dead tuple로 남기는 방식에 가깝습니다. 더 이상 어떤 snapshot에서도 필요 없는 dead tuple은 vacuum이 정리합니다. 오래 열린 transaction이나 replication slot 때문에 vacuum이 old tuple을 제거하지 못하면 table과 index가 커지고, query가 더 많은 page를 읽으며, xid wraparound 방지 작업까지 부담이 커질 수 있습니다.

### InnoDB undo log와 PostgreSQL dead tuple은 같은 것인가요?

같은 문제, 즉 MVCC에서 과거 version을 어떻게 보여 줄 것인가를 해결하지만 같은 구조는 아닙니다. InnoDB는 current clustered record와 undo log chain을 통해 과거 version을 재구성합니다. PostgreSQL은 heap 안에 여러 tuple version이 남고, tuple header의 transaction id와 snapshot으로 visibility를 판단합니다. 그래서 정리 작업도 InnoDB purge와 PostgreSQL vacuum으로 다르게 나타납니다.

### MySQL과 PostgreSQL의 WAL/redo 차이를 어떻게 설명하나요?

공통 원리는 data page보다 log를 먼저 durable하게 기록해 crash recovery를 가능하게 한다는 것입니다. InnoDB redo log는 dirty page가 아직 tablespace에 flush되지 않아도 committed change를 복구할 수 있게 합니다. PostgreSQL WAL은 heap/index/page change와 commit record를 남기고 checkpoint 이후 재생합니다. 다만 InnoDB undo는 WAL과 별개로 rollback 및 consistent read에 쓰이고, PostgreSQL은 undo segment를 별도로 두는 방식이 아니라 heap tuple version과 WAL을 함께 사용합니다.

### PostgreSQL에서 index-only scan이 항상 heap을 안 보나요?

아닙니다. index에 필요한 컬럼이 있어도 heap page의 tuple들이 현재 snapshot에서 모두 보인다는 정보가 있어야 합니다. PostgreSQL은 visibility map을 사용해 heap 확인을 생략할 수 있는지 판단합니다. vacuum이 visibility map을 갱신하므로, vacuum 상태가 좋지 않거나 page가 all-visible 상태가 아니면 index-only scan에서도 heap fetch가 발생할 수 있습니다.

### InnoDB gap lock은 언제 설명해야 하나요?

범위 조건, phantom 방지, REPEATABLE READ, locking read가 함께 나오면 gap lock을 떠올려야 합니다. InnoDB는 index record뿐 아니라 record 사이 gap을 잠가 그 범위에 새 row가 삽입되는 것을 막을 수 있습니다. unique index로 정확히 한 row를 찾는 경우와 range scan은 다르게 봐야 하며, 적절한 index가 없으면 lock 범위가 넓어질 수 있습니다.

### PostgreSQL planner가 sequential scan을 골랐으면 무조건 나쁜가요?

아닙니다. table이 작거나 조건 selectivity가 낮거나 랜덤 I/O 비용이 크다고 판단되면 sequential scan이 합리적일 수 있습니다. 문제는 estimated rows와 actual rows가 크게 어긋나는 경우, 통계가 낡은 경우, correlation이 반영되지 않는 경우, memory spill이 발생하는 경우입니다. `EXPLAIN (ANALYZE, BUFFERS)`로 실제 row 수와 buffer 사용을 함께 봐야 합니다.

## 함정 질문

### "MVCC면 read lock이 없으니 lock 문제도 거의 없죠?"

틀린 방향입니다. MVCC는 reader와 writer 충돌을 줄여 주지만 write-write conflict, DDL lock, foreign key 관련 lock, explicit locking read, predicate lock, metadata lock 같은 문제를 없애지 않습니다. InnoDB에서는 next-key lock이 insert를 막을 수 있고, PostgreSQL에서는 row lock 대기나 idle transaction이 vacuum 지연을 만들 수 있습니다.

### "InnoDB secondary index는 행 전체를 가지고 있으니 covering index가 자동 아닌가요?"

아닙니다. InnoDB secondary index leaf에는 secondary key와 primary key가 들어갑니다. 필요한 컬럼이 secondary index에 모두 포함되어 있으면 covering이 될 수 있지만, 그렇지 않으면 primary key로 clustered index를 다시 찾아야 합니다. "secondary index에 row 전체가 있다"는 말은 InnoDB 구조를 잘못 이해한 것입니다.

### "PostgreSQL은 heap이라서 primary key 순서 locality가 전혀 의미 없나요?"

전혀 의미 없다고 하면 과합니다. PostgreSQL primary key가 InnoDB clustered index처럼 table physical order를 자동으로 결정하지는 않습니다. 하지만 index locality, correlation, CLUSTER 명령, BRIN index, partitioning, insert pattern은 여전히 성능에 영향을 줍니다. 안전한 답은 "InnoDB만큼 primary key가 table body를 직접 구성하지는 않지만, 접근 패턴과 인덱스 구조에는 여전히 중요하다"입니다.

### "VACUUM은 DELETE한 공간을 운영체제에 바로 돌려주는 작업인가요?"

보통은 아닙니다. PostgreSQL vacuum은 dead tuple을 재사용 가능한 공간으로 표시하고 visibility/freeze 정보를 관리합니다. 파일을 운영체제에 바로 줄이는 것은 `VACUUM FULL`처럼 더 강한 작업이 필요할 수 있고, 이는 table rewrite와 lock 비용을 동반합니다. 일반 vacuum을 "디스크 파일 축소 명령"으로 이해하면 운영 판단을 잘못할 수 있습니다.

### "redo log가 있으면 data page flush는 성능만의 문제인가요?"

성능만의 문제가 아닙니다. log와 page flush의 균형은 crash recovery 시간, checkpoint pressure, write burst, durability 설정과 연결됩니다. data page가 늦게 flush되어도 redo로 복구할 수 있지만, redo가 무한히 커질 수는 없고 checkpoint와 flush가 따라와야 합니다. 운영에서는 redo/WAL volume과 checkpoint 상태를 함께 봅니다.

### "PostgreSQL SERIALIZABLE이면 모든 충돌을 대기시켜 순서대로 처리하나요?"

PostgreSQL SERIALIZABLE은 Serializable Snapshot Isolation 방식으로 동작하며, 모든 것을 단순 대기로 직렬화하는 모델이 아닙니다. 위험한 read-write dependency가 감지되면 serialization failure가 발생할 수 있고, 애플리케이션은 transaction retry를 준비해야 합니다. "강한 격리 수준이면 그냥 느려질 뿐 실패는 없다"는 답은 위험합니다.

### 전이 질문: 같은 SQL 이름 아래 다른 저장 엔진 계약

새 DBMS를 만나면 `SQL을 지원한다`보다 저장 구조, version 관리, 복구 로그, lock 모델, planner 통계를 먼저 확인합니다. 이 다섯 축을 잡으면 익숙한 용어와 낯선 구현을 안전하게 분리할 수 있습니다. 일반 원리는 출발점이고, 운영 판단은 엔진별 관측값으로 닫아야 합니다.

## 더 깊게 볼 자료

- [MySQL 8.4 Reference Manual - Clustered and Secondary Indexes](https://dev.mysql.com/doc/refman/8.4/en/innodb-index-types.html): InnoDB clustered index와 secondary index의 공식 설명을 확인합니다.
- [MySQL 8.4 Reference Manual - The InnoDB Buffer Pool](https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html): buffer pool, page caching, LRU 변형을 볼 때 기준 자료로 삼습니다.
- [MySQL 8.4 Reference Manual - InnoDB Redo Log](https://dev.mysql.com/doc/refman/8.4/en/innodb-redo-log.html): redo log와 crash recovery의 경계를 확인합니다.
- [MySQL 8.4 Reference Manual - `innodb_flush_log_at_trx_commit`](https://dev.mysql.com/doc/refman/8.4/en/innodb-parameters.html#sysvar_innodb_flush_log_at_trx_commit): commit log write/flush 설정과 ACID trade-off를 확인합니다.
- [MySQL 8.4 Reference Manual - InnoDB Undo Logs](https://dev.mysql.com/doc/refman/8.4/en/innodb-undo-logs.html): undo log가 rollback과 MVCC consistent read에 어떻게 연결되는지 확인합니다.
- [MySQL 8.4 Reference Manual - InnoDB Locking](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html): record, gap, next-key lock을 공식 용어로 다시 정리합니다.
- [PostgreSQL Current Documentation - Multiversion Concurrency Control](https://www.postgresql.org/docs/current/mvcc.html): PostgreSQL snapshot과 tuple visibility의 출발점입니다.
- [PostgreSQL Current Documentation - Database Page Layout](https://www.postgresql.org/docs/current/storage-page-layout.html): heap page와 tuple header를 더 낮은 층위에서 확인합니다.
- [PostgreSQL Current Documentation - Routine Vacuuming](https://www.postgresql.org/docs/current/routine-vacuuming.html): vacuum, autovacuum, freeze, 통계 갱신의 운영 의미를 확인합니다.
- [PostgreSQL Current Documentation - Write-Ahead Logging](https://www.postgresql.org/docs/current/wal-intro.html): WAL, checkpoint, crash recovery 설명의 공식 기준입니다.
- [PostgreSQL Current Documentation - WAL Reliability](https://www.postgresql.org/docs/16/wal-reliability.html): OS buffer cache, controller cache, disk cache와 full page write 위험을 확인합니다.
- [PostgreSQL Current Documentation - Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html): PostgreSQL row/table lock mode와 대기 상황을 정리합니다.
- [PostgreSQL Current Documentation - Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html): planner와 실행 계획을 실제 관측으로 검증하는 방법을 확인합니다.
- [Linux kernel 문서: VFS writeback](https://docs.kernel.org/filesystems/vfs.html): page cache dirty/writeback 상태와 fsync 오류 보고 경계를 확인합니다.
- [Linux kernel 문서: Multi-Queue Block I/O Queueing](https://docs.kernel.org/block/blk-mq.html): block I/O request가 software queue와 hardware dispatch queue를 거쳐 driver로 내려가는 흐름을 확인합니다.
