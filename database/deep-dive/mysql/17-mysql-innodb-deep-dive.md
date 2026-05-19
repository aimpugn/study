# MySQL InnoDB Deep Dive

이 문서는 MySQL InnoDB를 저장 구조, MVCC, lock, online DDL, 운영 관측 흐름으로 읽기 위한 장문 학습 문서다. 각 절은 서로 이어지지만 독립적으로도 다시 읽을 수 있게 구성한다.

## InnoDB architecture: clustered index, buffer pool, redo/undo

InnoDB를 이해할 때 가장 먼저 버려야 하는 생각은 “테이블은 그냥 행이 모여 있는 파일이고, 인덱스는 그 옆에 붙은 빠른 검색표”라는 그림이다. InnoDB 테이블에서 행은 기본적으로 clustered index, 즉 행 데이터까지 품고 있는 기본 B+Tree 안에 놓인다. MySQL 공식 문서는 InnoDB 테이블마다 행 데이터를 저장하는 특별한 인덱스가 있고, 보통 이 clustered index가 primary key와 같다고 설명한다. Primary key가 없으면 모든 컬럼이 `NOT NULL`인 첫 번째 unique index를 고르고, 그것도 없으면 InnoDB가 숨은 `GEN_CLUST_INDEX`를 만든다는 규칙도 같은 문서에 나온다. 그래서 primary key는 단순한 논리 식별자가 아니라, 행이 어느 순서의 leaf page에 놓이고 secondary index가 어떤 값을 들고 다시 본문 행을 찾아갈지까지 바꾸는 물리 저장 구조의 핵심 입력이다. 이 절의 목표는 InnoDB가 clustered index를 중심으로 row, page, buffer pool, redo log, undo log를 어떻게 묶는지 설명하는 것이다. 여기서 놓치면 운영에서 “PK는 아무거나 unique하면 된다”, “buffer pool은 단순 캐시다”, “redo와 undo는 둘 다 복구 로그니까 비슷하다” 같은 위험한 shortcut을 갖게 된다.

이 구조가 등장한 배경에는 MySQL이 오랫동안 여러 storage engine을 품어 왔다는 역사도 있다. SQL 계층은 같은 `SELECT`, `UPDATE`, `COMMIT` 문장을 받지만, 실제 행을 저장하고 잠그고 복구하는 책임은 엔진마다 달랐다. InnoDB는 transaction, crash recovery, row-level locking, foreign key 같은 기능을 강하게 제공하는 엔진으로 자리 잡으면서, 행을 clustered index에 모으고, buffer pool과 redo/undo를 긴밀하게 묶는 방향으로 발전했다. 그래서 InnoDB를 볼 때는 “MySQL의 문법”만이 아니라 “이 SQL이 InnoDB라는 엔진의 page, log, version 구조로 내려가면 무엇이 되는가”를 함께 읽어야 한다. 이 배경을 놓치면 MySQL을 쓰면서도 MyISAM식 파일 감각이나 일반적인 heap table 감각으로 운영 결정을 내리게 된다.

### 왜 InnoDB는 clustered index 중심으로 생각해야 하는가

전통적인 설명에서는 테이블 파일과 인덱스 파일을 분리해서 떠올리기 쉽다. 어떤 DBMS나 어떤 저장 엔진에서는 이런 그림이 어느 정도 맞는다. 하지만 InnoDB의 기본 테이블 접근 모델은 row를 별도 heap에 두고 index가 row 위치만 가리키는 형태가 아니라, clustered index leaf page가 row 자체를 품는 형태다. 공식 MySQL 8.4 문서의 “Clustered and Secondary Indexes” 절은 clustered index search가 row data를 포함하는 page로 직접 이어지기 때문에 큰 테이블에서 별도 row page를 찾아가는 디스크 I/O를 줄일 수 있다고 설명한다. 반대로 secondary index record에는 secondary key뿐 아니라 해당 row의 primary key column이 들어 있고, InnoDB는 그 primary key 값을 사용해 clustered index에서 row를 다시 찾는다. 이 한 문장이 운영 설계에서 매우 크다. secondary index는 “row 주소”를 들고 있는 것이 아니라 “clustered index로 다시 들어갈 primary key”를 들고 있다. 따라서 primary key가 길면 모든 secondary index leaf record가 비대해진다. PK를 UUID 문자열로 둘지, 짧은 bigint로 둘지, 복합 natural key로 둘지는 단순 취향이 아니라 secondary index 전체 크기, buffer pool 효율, redo 양, page split 빈도에 영향을 준다.

InnoDB architecture 문서는 memory 구조와 disk 구조를 함께 보여준다. memory 쪽에는 buffer pool, adaptive hash index, change buffer, log buffer가 있고, disk 쪽에는 tablespace, redo log, doublewrite buffer 같은 구조가 있다. 이 그림을 “구성요소 목록”으로만 보면 학습 효과가 약하다. 실제로는 한 행의 변경이 clustered index page를 읽고, buffer pool에서 dirty page가 되고, redo log record가 먼저 durable한 흔적을 남기고, undo log record가 과거 버전을 복원할 길을 남기며, 나중에 checkpoint와 purge가 뒤처리를 하는 시간 순서가 중요하다. 다시 말해 InnoDB의 구조는 파일 목록이 아니라 “읽기와 쓰기를 동시에 많이 처리하면서도 crash 후 일관성을 되찾기 위한 상태 전이 장치”다.

### 첫 번째 벽돌: 한 건의 UPDATE가 지나가는 길

다음 예시는 `orders` 테이블의 primary key가 `order_id`이고, `status`에 secondary index가 있다고 가정한다. SQL 하나가 실행될 때 InnoDB가 실제로는 어떤 구조를 차례로 건드리는지 손으로 따라가면 clustered index, buffer pool, redo, undo의 역할이 분리된다.

```sql
CREATE TABLE orders (
  order_id BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  PRIMARY KEY (order_id),
  KEY idx_status (status)
) ENGINE=InnoDB;

UPDATE orders
SET status = 'PAID'
WHERE order_id = 1001;
```

```text
1. SQL layer
   UPDATE orders SET status='PAID' WHERE order_id=1001
          |
          v
2. clustered index search
   PRIMARY KEY(order_id) B+Tree -> leaf page P42 -> row(order_id=1001)
          |
          v
3. buffer pool
   P42가 memory에 없으면 disk에서 읽어 old sublist midpoint 근처로 들어온다.
   P42가 이미 있으면 latch와 row lock을 거쳐 page 안 record를 수정한다.
          |
          v
4. undo log
   이전 status='READY'를 다시 만들 수 있는 undo record를 남긴다.
          |
          v
5. redo log
   page P42와 secondary index page 변경을 다시 적용할 수 있는 redo record를 남긴다.
          |
          v
6. dirty page
   buffer pool의 P42와 idx_status 관련 page는 memory에서 바뀌었지만,
   실제 tablespace file에는 나중에 flush될 수 있다.
```

이 trace에서 row lock과 page latch를 구분해야 한다. row lock은 트랜잭션 사이의 논리적 동시성 제어를 위한 잠금이다. page latch는 page 안 구조를 동시에 고치다 망가뜨리지 않도록 아주 짧게 잡는 내부 보호 장치에 가깝다. 이 절의 주제는 lock 깊은 설명이 아니라 storage architecture이므로 latch를 길게 다루지는 않지만, “row가 clustered index page 안에 있다”는 사실 때문에 record 수정은 page 구조 변경이기도 하다는 점은 꼭 붙잡아야 한다.

### Primary key가 물리 구조를 바꾸는 방식

InnoDB에서 primary key는 세 층위에 동시에 영향을 준다.

| 층위 | primary key가 미치는 영향 | 운영에서 보이는 증상 |
|---|---|---|
| clustered index leaf | 행 자체가 primary key 순서로 배치된다 | 삽입 패턴에 따라 page split, fragment, hot page가 달라진다 |
| secondary index leaf | secondary key와 함께 primary key column을 저장한다 | PK가 길면 secondary index 전체가 커지고 buffer pool에 덜 들어간다 |
| lookup path | secondary index hit 후 primary key로 clustered index를 다시 찾는다 | covering index가 아니면 random page 접근이 늘 수 있다 |

예를 들어 `PRIMARY KEY(order_id)`가 8바이트 정수라면 `idx_status`의 leaf record는 대략 `status` 값과 8바이트 primary key를 들고 row 본문으로 돌아간다. 반대로 `PRIMARY KEY(country_code, tenant_id, external_order_no)`처럼 길고 가변적인 복합 key라면 모든 secondary index가 이 복합 key를 들고 다닌다. `idx_status`, `idx_customer_id`, `idx_created_at`이 각각 같은 긴 PK를 반복 저장한다. 그래서 “secondary index 하나 추가했을 뿐인데 디스크가 생각보다 많이 늘었다”는 장애 감각은 InnoDB에서는 이상한 일이 아니다. row 본문이 clustered index에 있고 secondary index가 clustered index key를 품는다는 구조를 모르고 있으면, 인덱스 크기 예측과 buffer pool 압박을 과소평가한다.

또 하나의 함정은 UUID 같은 랜덤 primary key다. InnoDB clustered index leaf page는 key order를 유지해야 한다. 순차 증가 key는 보통 B+Tree 오른쪽 끝으로 들어가므로 마지막 page hot spot과 auto-increment 관련 병목은 따로 고려해야 하지만, page split은 비교적 예측 가능하다. 랜덤 key는 기존 tree 곳곳에 삽입되므로 많은 leaf page를 건드리고, page split과 dirty page 분산이 커질 수 있다. 이 말이 “랜덤 UUID는 항상 금지”라는 뜻은 아니다. 분산 write hot spot을 줄이는 장점이 필요한 시스템도 있다. 다만 InnoDB에서 PK 선택은 애플리케이션 모델링만의 문제가 아니라 storage write pattern 선택이라는 사실을 먼저 인정해야 한다.

### Buffer pool은 단순 캐시가 아니라 page life-cycle 관리자다

MySQL 공식 buffer pool 문서는 대량 read 효율을 위해 buffer pool이 page 단위로 나뉘고, cache 관리를 위해 page linked list와 LRU 변형 알고리즘을 사용한다고 설명한다. 특히 새로 읽은 page를 list head가 아니라 midpoint에 넣어 old sublist와 young sublist를 나누는 전략이 중요하다. 단순 LRU라면 큰 table scan 하나가 자주 쓰던 hot page를 몰아낼 수 있다. InnoDB의 midpoint insertion은 이런 scan pollution을 줄이려는 장치다. 공식 문서도 `mysqldump`나 `WHERE` 없는 `SELECT` 같은 table scan이 많은 데이터를 buffer pool로 끌어와 오래된 page를 밀어낼 수 있다고 경고한다.

Buffer pool을 “메모리에 올라온 데이터” 정도로만 보면 tuning 판단이 얕아진다. 실제 운영에서는 다음 지표를 함께 읽어야 한다.

```text
SHOW ENGINE INNODB STATUS\G

----------------------
BUFFER POOL AND MEMORY
----------------------
Buffer pool size   131072
Free buffers       124908
Database pages     5720
Old database pages 2071
Modified db pages  910
Pending reads 0
Pending writes: LRU 0, flush list 0, single page 0
Pages read 197, created 5523, written 5060
Buffer pool hit rate 999 / 1000
```

이 출력은 단순히 “hit rate가 높다/낮다”로 끝나지 않는다. `Database pages`는 현재 LRU list에 있는 database page 규모를 보여 주고, `Old database pages`는 old sublist 쪽 page 규모를 보여 준다. `Modified db pages`는 memory 안에서 바뀌었지만 아직 tablespace file에 flush되지 않은 dirty page 수다. `Pending writes: flush list`가 늘고 `Modified db pages`가 계속 쌓이면 redo checkpoint와 page cleaner가 뒤처지는지 봐야 한다. `Pending reads`가 늘고 hit rate가 떨어지면 working set이 buffer pool보다 큰지, 갑작스러운 scan이 hot page를 밀어냈는지, secondary index 설계 때문에 random lookup이 늘었는지로 좁혀 간다.

여기서 “buffer pool이 크면 모든 게 해결된다”는 오해도 고쳐야 한다. buffer pool이 커지면 더 많은 page를 memory에 둘 수 있지만, dirty page를 언제 flush할지, redo log capacity가 checkpoint를 얼마나 여유 있게 밀어 줄지, long transaction 때문에 undo purge가 밀리는지까지 같이 봐야 한다. memory가 충분해도 redo log가 너무 빨리 차거나 checkpoint age가 커지면 flush 압력이 커진다. 반대로 buffer pool이 작아도 working set이 작고 access pattern이 안정적이면 hit rate는 높을 수 있다. 시니어가 보는 것은 절대값 하나가 아니라 page life-cycle의 병목 지점이다.

### Redo log는 “변경을 다시 적용하는 기록”이다

공식 MySQL redo log 문서는 redo log를 crash recovery 동안 incomplete transaction이 쓴 data를 바로잡는 disk-based data structure로 설명한다. 정상 동작 중에는 SQL 문이나 low-level API call에서 나온 table data 변경 요청이 redo log record로 encoding된다. 갑작스러운 shutdown 전에 data file 갱신이 끝나지 못한 변경은 initialization 중 connection을 받기 전에 자동으로 replay된다. 핵심은 redo가 “논리 SQL을 다시 실행한다”가 아니라 page 변경을 다시 적용할 수 있게 하는 기록이라는 점이다.

다음 timeline을 보자.

```text
T0  UPDATE orders SET status='PAID' WHERE order_id=1001
T1  buffer pool page P42 modified
T2  redo log buffer에 P42 변경 record append
T3  COMMIT path에서 redo가 durability policy에 맞게 disk로 flush
T4  mysqld crash before dirty page P42 reaches tablespace file
T5  restart
T6  recovery reads redo from last checkpoint LSN
T7  redo replay makes tablespace page catch up to committed change
```

이 흐름 때문에 write-ahead logging의 감각이 필요하다. data page를 먼저 안정적으로 쓰고 log는 나중에 쓰는 모델이 아니다. crash recovery가 가능하려면, page flush보다 변경 설명이 먼저 durable해야 한다. InnoDB redo log는 LSN, 즉 계속 증가하는 log sequence number 흐름으로 checkpoint와 연결된다. 오래된 redo는 checkpoint가 진행되면서 truncate될 수 있고, 아직 checkpoint가 따라잡지 못한 구간은 recovery에 필요하다. 운영에서 redo log capacity를 너무 작게 잡으면 checkpoint가 자주 몰아치고 dirty page flush가 급해질 수 있다. 너무 크게 잡으면 burst write에는 여유가 생기지만 recovery time과 disk 사용, 관측 기준이 달라진다. 그래서 redo tuning은 “크게 하면 성능 좋다”가 아니라 workload burst, recovery objective, dirty page flush 능력을 함께 보는 문제다.

### Undo log는 “과거 버전을 보여 주고 rollback할 수 있게 하는 기록”이다

공식 MySQL undo log 문서는 undo log record가 read-write transaction의 최신 변경을 되돌리는 정보를 담고, 다른 transaction이 consistent read를 위해 original data를 봐야 하면 unmodified data를 undo log record에서 가져온다고 설명한다. InnoDB multi-versioning 문서도 InnoDB가 old version 정보를 undo tablespace의 rollback segment에 저장해 concurrency와 rollback을 지원한다고 설명한다. 이 때문에 undo는 redo와 완전히 다르다. redo는 crash 후 committed page change를 다시 적용하는 방향이고, undo는 transaction rollback이나 snapshot read가 과거 값을 재구성하는 방향이다.

다음 예시는 같은 row가 두 번 바뀌는 동안 reader가 무엇을 보는지 보여 준다.

```text
초기 clustered record
  order_id=1001, status='READY', DB_TRX_ID=80, roll_ptr=NULL

T1 updates status='PAID'
  clustered record
    order_id=1001, status='PAID', DB_TRX_ID=101, roll_ptr=U1
  undo U1
    before image: status='READY', previous DB_TRX_ID=80

T2 updates status='SHIPPED'
  clustered record
    order_id=1001, status='SHIPPED', DB_TRX_ID=102, roll_ptr=U2
  undo U2
    before image: status='PAID', previous roll_ptr=U1

reader snapshot sees only trx_id <= 100
  current record DB_TRX_ID=102 is too new
      -> follow U2 gives status='PAID', trx_id=101, still too new
      -> follow U1 gives status='READY', trx_id=80, visible
```

이 trace는 단순화한 모델이지만, 중요한 학습 효과가 있다. MVCC에서 “여러 버전이 테이블에 별도 row로 모두 저장된다”는 그림은 InnoDB에는 정확하지 않다. 현재 row는 clustered index record에 있고, 과거 버전 재구성에 필요한 정보는 undo chain을 통해 따라간다. 이 구조 때문에 long-running transaction은 undo purge를 막을 수 있다. 오래된 snapshot을 가진 transaction이 있으면 purge thread가 “이 undo는 더 이상 누구도 필요로 하지 않는다”고 단정할 수 없다. 그래서 오래 열린 read transaction 하나가 undo tablespace 증가, purge lag, history list length 증가, secondary index 정리 지연 같은 운영 증상으로 번질 수 있다.

### Consistent read는 lock-free가 아니라 snapshot을 기준으로 다른 길을 탄다

InnoDB consistent read 문서는 READ COMMITTED와 REPEATABLE READ에서 일반 `SELECT`를 처리하는 기본 모드이며, 접근하는 table에 lock을 set하지 않으므로 다른 session이 동시에 modify할 수 있다고 설명한다. 기본 REPEATABLE READ에서 일반 `SELECT`를 수행하면 InnoDB는 transaction에 timepoint를 주고, 그 timepoint 뒤에 다른 transaction이 delete, insert, update한 내용은 보이지 않게 한다. 여기서 오해가 자주 생긴다. “SELECT가 lock을 잡지 않는다”는 말은 “아무 비용이 없다”가 아니다. 최신 record가 snapshot에 보이지 않으면 undo chain을 따라 과거 version을 재구성해야 한다. long transaction이 많으면 과거 version을 오래 보존해야 한다. 또한 locking read, `UPDATE`, `DELETE`, foreign key check, duplicate key check는 일반 consistent read와 다른 lock 경로를 탄다.

InnoDB architecture를 실무적으로 설명하면 다음 네 문장으로 압축할 수 있다.

1. 읽기와 쓰기의 중심 단위는 row가 아니라 page이며, row는 clustered index page 안에 있다.
2. buffer pool은 page를 memory에 머물게 하고, dirty page와 LRU 상태를 관리한다.
3. redo log는 crash 후 page 변경을 다시 적용할 수 있게 하고, checkpoint와 flush 압력을 만든다.
4. undo log는 rollback과 snapshot read가 과거 값을 볼 수 있게 하지만, 오래된 snapshot이 남으면 purge와 공간 회수의 발목을 잡는다.

### 운영 관측: architecture를 숫자로 되짚는 순서

운영에서 InnoDB 내부를 볼 때 하나의 명령만 신뢰하면 위험하다. 먼저 `SHOW ENGINE INNODB STATUS\G`로 큰 구조를 본 뒤, Performance Schema와 Information Schema로 lock, metadata, table/index 규모를 좁혀 가는 편이 안전하다. DU38의 관측 anchor는 다음처럼 잡을 수 있다.

```sql
-- buffer pool과 dirty page, checkpoint 압력의 큰 그림
SHOW ENGINE INNODB STATUS\G

-- table/index 크기와 row 추정치
SELECT table_schema, table_name, engine, table_rows, data_length, index_length
FROM information_schema.tables
WHERE table_schema = 'app'
ORDER BY data_length + index_length DESC
LIMIT 20;

-- buffer pool page 분포를 더 자세히 볼 수 있는 환경이라면
SELECT pool_id, page_type, count(*) AS pages
FROM information_schema.innodb_buffer_page
GROUP BY pool_id, page_type
ORDER BY pages DESC;
```

PASS 신호는 하나의 숫자가 아니라 원인 가설과 관측값이 맞물리는 것이다. 예를 들어 secondary index를 추가한 뒤 `index_length`가 예상보다 크게 늘고 buffer pool hit rate가 흔들리며 random read가 늘었다면, primary key 길이와 covering index 여부를 함께 본다. 반대로 `Modified db pages`가 계속 쌓이고 flush list pending write가 증가한다면, read path보다 write flush와 redo checkpoint 압력을 먼저 본다. FAIL 신호는 관측값이 서로 맞지 않는데도 하나의 설명으로 밀어붙이는 것이다. hit rate가 높다고 I/O 문제가 없다고 단정하거나, dirty page가 많다고 무조건 buffer pool을 줄이는 판단은 architecture를 숫자 하나로 납작하게 만든다.

### Local seed와 기존 문서에서 이어받을 것

로컬 seed `database/mysql/*`는 아직 InnoDB 내부를 체계적으로 설명하는 deep-dive 문서는 아니지만, 몇 가지 관측 습관을 제공한다. `database/mysql/information.md`류의 자료는 `information_schema`를 통해 table과 server metadata를 확인하는 방향을 준다. `database/mysql/explains/when_join.md`는 index와 execution plan이 실제 query path에 영향을 준다는 감각을 준다. `database/mysql/problems/mysql5_online_large_indexing.md`는 수억 건 테이블에서 index 추가가 단순 DDL이 아니라 memory, disk, lock, service availability 문제라는 운영 감각을 준다. 이 절은 그 seed들을 “InnoDB는 왜 그런 운영 증상을 보이는가”라는 구조 설명으로 승격한다.

특히 `mysql5_online_large_indexing.md`에 나온 `TABLE_ROWS`, `DATA_LENGTH`, `INDEX_LENGTH` 숫자는 DU38과 DU39의 좋은 연결점이다. 수억 row table에서 index를 추가하면 DU38 관점에서는 secondary index leaf가 primary key를 반복 저장하고, index build가 많은 page read/write와 redo/temporary space를 유발하며, buffer pool working set을 흔든다. DU39 관점에서는 online DDL이라 해도 metadata lock과 final commit phase가 남는다. 따라서 “인덱스 하나 추가”라는 요청은 storage architecture, concurrency, observability를 함께 읽어야 한다.

### 흔한 오해를 짧은 반례로 고치기

첫 번째 오해는 “primary key는 논리적으로만 중요하다”는 말이다. InnoDB에서는 primary key가 clustered index이고 secondary index record에 들어간다. 긴 PK는 모든 secondary index 크기를 키우며, 랜덤 PK는 page split과 dirty page 분산을 바꿀 수 있다. 논리 모델에서 좋은 key가 물리 모델에서도 항상 좋은 key는 아니다.

두 번째 오해는 “buffer pool hit rate가 높으면 storage 문제는 없다”는 말이다. hit rate가 높아도 dirty page flush가 밀리거나 redo checkpoint 압력이 높을 수 있다. 반대로 hit rate가 낮은 원인이 buffer pool 크기 부족이 아니라 대량 scan, 비효율 secondary lookup, covering index 부족일 수도 있다. 관측은 한 줄 숫자가 아니라 page life-cycle 전체로 읽어야 한다.

세 번째 오해는 “redo와 undo는 둘 다 transaction log다”라는 말이다. redo는 committed change를 crash 후 다시 적용하기 위한 forward 기록이고, undo는 rollback과 consistent read가 이전 값을 재구성하기 위한 backward 기록이다. 둘을 섞어 생각하면 recovery, purge, snapshot, disk pressure를 잘못 진단한다.

네 번째 오해는 “MVCC 덕분에 읽기는 쓰기와 무관하다”는 말이다. 일반 consistent read는 table lock을 set하지 않을 수 있지만, 최신 record가 snapshot보다 새로우면 undo chain을 따라가야 하고, 오래된 snapshot은 purge를 늦춘다. 읽기가 쓰기를 직접 block하지 않아도, 오래 열린 읽기는 내부 청소와 공간 회수에 영향을 줄 수 있다.

### 직접 replay할 작은 실험

실제 MySQL이 있는 환경에서는 다음 실험을 해 보면 이 절의 핵심을 눈으로 확인할 수 있다. 테이블을 만들고 primary key 폭과 secondary index 크기의 관계를 비교하는 것이 가장 작고 안전한 실험이다.

```sql
CREATE DATABASE IF NOT EXISTS innodb_lab;
USE innodb_lab;

CREATE TABLE short_pk_orders (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  external_id CHAR(36) NOT NULL,
  status VARCHAR(20) NOT NULL,
  payload VARCHAR(200) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_status (status)
) ENGINE=InnoDB;

CREATE TABLE wide_pk_orders (
  tenant_id BIGINT NOT NULL,
  external_id CHAR(36) NOT NULL,
  id BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL,
  payload VARCHAR(200) NOT NULL,
  PRIMARY KEY (tenant_id, external_id),
  KEY idx_status (status)
) ENGINE=InnoDB;

SELECT table_name, table_rows, data_length, index_length
FROM information_schema.tables
WHERE table_schema = 'innodb_lab'
ORDER BY table_name;
```

충분한 row를 넣은 뒤 PASS 신호는 `wide_pk_orders`의 secondary index 쪽 공간이 더 커지는 경향을 관측하는 것이다. 정확한 byte 수는 row format, page fill, stats freshness, sample size에 따라 달라질 수 있으므로 숫자를 절대값으로 외우면 안 된다. FAIL 신호는 row 수가 너무 적어 page 단위 overhead만 보이는데도 결론을 일반화하는 것이다. 이 실험의 목적은 “PK 폭이 secondary index에 반복 반영된다”는 구조를 확인하는 것이지, 모든 환경의 index size 공식을 도출하는 것이 아니다.

### Insert, secondary index, redo/undo를 함께 보는 worked trace

InnoDB architecture를 더 실무적으로 만들려면 `UPDATE`만이 아니라 `INSERT`도 따라가야 한다. Insert는 undo가 필요 없다고 생각하기 쉽지만, transaction rollback을 생각하면 insert한 row를 사라지게 하는 undo 정보가 필요하다. 또한 secondary index가 있는 table에서는 clustered index page 하나만 바뀌지 않는다. `orders`에 `idx_status(status)`와 `idx_customer_id(customer_id)`가 있으면 하나의 insert가 최소 세 개의 B+Tree를 건드린다. 각 B+Tree page 변경은 buffer pool에 dirty page를 만들고 redo log record를 만든다.

```text
INSERT INTO orders(order_id, customer_id, status, amount)
VALUES (1002, 77, 'READY', 31000.00);

logical row:
  order_id=1002, customer_id=77, status='READY', amount=31000.00

physical-ish path:
  clustered PRIMARY tree
    key=1002 -> leaf page Pc
    insert full row into Pc
    dirty Pc, redo record Rc

  secondary idx_status
    key=('READY', 1002) -> leaf page Ps
    insert secondary entry containing status + PK
    dirty Ps, redo record Rs

  secondary idx_customer_id
    key=(77, 1002) -> leaf page Pu
    insert secondary entry containing customer_id + PK
    dirty Pu, redo record Ru

  undo
    rollback information says inserted clustered record and secondary entries
    can be removed if transaction aborts
```

이 trace는 “row 하나가 page 하나”라는 모델을 깨뜨린다. Secondary index가 많을수록 insert/update/delete는 더 많은 B+Tree page를 건드린다. `status` 값이 바뀌는 update라면 clustered record의 `status` column만 바뀌는 것이 아니라 `idx_status`에서 old key entry를 제거하고 new key entry를 넣는 작업도 생긴다. 이런 page 변경은 redo를 늘리고, dirty page를 늘리고, buffer pool eviction과 flush 압력에 영향을 준다. 그래서 write-heavy table에 index를 추가할 때는 read query 하나가 빨라지는 이점만 보지 말고 write amplification, 즉 한 logical write가 몇 개의 물리 구조 변경으로 증폭되는지 함께 봐야 한다. 이 표현이 영어라면 “쓰기 증폭(write amplification)” 정도로 병기하고, 본문에서는 “논리적 변경 하나가 여러 page와 log 변경으로 불어나는 현상”으로 이해하면 된다.

### Change buffer와 adaptive hash index는 중심이 아니라 주변 최적화다

InnoDB architecture diagram에는 change buffer와 adaptive hash index도 나온다. 이 둘을 모르면 그림이 비어 보이지만, 처음부터 중심으로 잡으면 오히려 길을 잃는다. 중심축은 clustered index, buffer pool, redo, undo다. Change buffer는 secondary index page가 buffer pool에 없을 때 특정 secondary index 변경을 나중에 merge할 수 있도록 buffering하는 최적화로 이해할 수 있다. 즉 write path의 일부 비용을 지금 당장 random read로 내지 않고 뒤로 미루는 장치다. Adaptive hash index는 특정 B+Tree access pattern 위에 memory hash 형태의 빠른 경로를 얹는 최적화로 볼 수 있다. 둘 다 workload와 version/config에 따라 효과와 부작용이 다르며, primary architecture의 의미를 바꾸지는 않는다.

운영에서 이 차이를 아는 것이 중요하다. 어떤 장애에서 “change buffer가 있으니 secondary index write 비용은 걱정하지 않아도 된다”고 말하면 틀린다. Buffering은 비용을 없애는 것이 아니라 시점과 형태를 바꾼다. Merge가 필요하고, redo와 buffer pool pressure는 여전히 남는다. “adaptive hash index가 있으니 B+Tree cost는 무시해도 된다”도 틀린다. AHI는 특정 repeated lookup에서 도움을 줄 수 있지만, primary key 폭, secondary index 크기, page split, redo/undo 구조를 대체하지 않는다. 최적화 장치는 중심 구조를 이해한 뒤에 읽어야 한다.

### Page split과 fill factor 감각

InnoDB B+Tree leaf page는 key order를 유지한다. 새 key가 들어왔는데 leaf page에 공간이 부족하면 page split이 일어난다. 이때 하나의 logical insert가 page allocation, record redistribution, parent page update로 커질 수 있다. 특히 random primary key는 tree 곳곳의 leaf page에 insert하므로 많은 page가 dirty해지고 split 위치가 분산된다. 순차 key는 오른쪽 끝 page에 집중되므로 split이 더 예측 가능하지만, 마지막 page와 auto-increment lock 또는 insert hot spot을 따로 고려해야 한다. 즉 둘 중 하나가 언제나 정답이 아니라, workload가 원하는 tradeoff를 선택해야 한다.

```text
순차 key insert
  leaf pages:
    [1..100] [101..200] [201..300*]
                              ^
                              new key 301 arrives near right edge

랜덤 key insert
  leaf pages:
    [1..100*] [101..200] [201..300*] [301..400*]
       ^                       ^          ^
       inserts arrive across many pages
```

이 그림은 정확한 page split algorithm을 재현하려는 것이 아니라, 운영 감각을 만들기 위한 것이다. 순차 key는 locality가 좋아 buffer pool에 같은 leaf page가 머물 가능성이 높지만, 쓰기 집중이 생길 수 있다. 랜덤 key는 집중을 분산시킬 수 있지만, 더 많은 leaf page를 건드리고 secondary index까지 합쳐 dirty page 수를 키울 수 있다. PK 설계 회의에서 “UUID가 편하다”, “BIGINT가 빠르다” 같은 단정 대신, “우리 workload는 insert locality, sharding, secondary index size, read path, replication, application id generation 중 무엇을 우선하는가”를 물어야 한다.

### Crash recovery를 두 단계로 생각하기

Redo와 undo를 구분했으면 crash recovery도 더 잘 보인다. Crash 직전에는 committed transaction의 dirty page가 data file에 아직 flush되지 않았을 수 있고, uncommitted transaction의 일부 page 변경이 data file에 flush되었을 수도 있다. Recovery는 단순히 “redo를 다 적용한다”가 아니라, committed change를 반영하고 uncommitted change를 되돌리는 방향이 함께 필요하다. Redo는 page를 forward로 catch up시키고, undo는 committed되지 않은 transaction의 효과를 제거하거나 rollback 상태를 완성하는 데 필요하다.

```text
crash moment:
  Transaction A committed
    redo durable
    dirty page not yet flushed

  Transaction B not committed
    some dirty page may have reached disk
    undo information exists

restart recovery:
  redo phase
    replay A and B page changes up to log point so pages become structurally consistent

  undo/rollback phase
    remove effects of B because B did not commit

result:
  A visible, B not visible
```

이 trace가 중요한 이유는 “commit은 data page가 모두 disk에 쓰였다는 뜻”이라는 오해를 고치기 때문이다. DBMS가 매 commit마다 관련 data page를 모두 flush한다면 write throughput은 크게 떨어진다. WAL/redo 구조는 data page flush를 늦출 수 있게 해 주되, crash 후 log를 통해 복구할 수 있게 한다. 운영에서 `innodb_flush_log_at_trx_commit` 같은 durability 관련 설정을 다룰 때도 이 원리를 모르면 성능 옵션처럼만 보인다. 실제로는 commit durability, crash loss window, fsync 비용 사이의 계약이다.

### Interview 관점의 압축 답변

면접이나 설계 리뷰에서 InnoDB architecture를 짧게 설명해야 한다면 다음 정도로 말할 수 있어야 한다. InnoDB table은 clustered index 중심으로 저장되고, primary key가 clustered index를 결정한다. Secondary index는 row pointer가 아니라 primary key 값을 품고 clustered index로 돌아간다. Buffer pool은 page cache이면서 dirty page와 LRU 상태를 관리한다. Redo log는 crash 후 page 변경을 다시 적용하기 위한 forward log이고, undo log는 rollback과 consistent read가 과거 version을 재구성하기 위한 backward 정보다. 그래서 PK 선택, secondary index 수, long transaction, redo capacity, buffer pool size는 서로 독립된 튜닝 항목이 아니라 같은 row/page/log life-cycle의 다른 표면이다.

이 압축 답변 뒤에 바로 따라와야 하는 practical trap은 primary key 선택이다. “PK는 짧고 stable하면 좋다”라는 조언은 맞을 때가 많지만, 이유가 빠지면 외운 말이 된다. 이유는 secondary index가 PK를 반복 저장하고, clustered index insertion pattern이 page locality와 split을 바꾸며, 모든 lookup과 write path가 그 구조 위에 놓이기 때문이다.

### 출처와 검증 경계

이 절의 공식 근거는 MySQL 8.4 Reference Manual의 [InnoDB Architecture](https://dev.mysql.com/doc/refman/8.4/en/innodb-architecture.html), [Buffer Pool](https://dev.mysql.com/doc/refman/8.4/en/innodb-buffer-pool.html), [Clustered and Secondary Indexes](https://dev.mysql.com/doc/refman/8.4/en/innodb-index-types.html), [Redo Log](https://dev.mysql.com/doc/refman/8.4/en/innodb-redo-log.html), [Undo Logs](https://dev.mysql.com/doc/refman/8.4/en/innodb-undo-logs.html), [InnoDB Multi-Versioning](https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html), [Consistent Nonlocking Reads](https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html)다. 로컬 seed는 `database/mysql/*` 아래의 MySQL/InnoDB 메모와 실험 기록이다. 이 절의 trace는 이해를 위한 단순화 모델이며, 실제 record header, undo record format, latch 내부 구현까지 완전하게 재현한 물리 포맷 설명은 아니다. 다만 “clustered index가 row를 품고, secondary index가 primary key를 품고, buffer pool과 redo/undo가 page 변경과 snapshot을 관리한다”는 판단은 공식 문서의 구조와 관측 명령으로 다시 확인할 수 있다.

## InnoDB locking/isolation/online DDL

InnoDB의 locking과 isolation을 운영 관점에서 이해하려면 “MVCC니까 reader와 writer는 서로 안 막는다”와 “online DDL이니까 서비스 중에도 안전하다”라는 두 문장을 먼저 의심해야 한다. 둘 다 일부 상황에서는 맞지만, 그대로 운영 규칙으로 쓰면 위험하다. InnoDB는 일반 `SELECT`에 consistent nonlocking read를 제공하지만, `UPDATE`, `DELETE`, locking read, unique check, foreign key check, gap/next-key lock은 다른 경로를 탄다. Online DDL은 많은 작업에서 concurrent DML을 허용할 수 있지만, metadata lock을 아예 없애는 기능이 아니다. MySQL 공식 Online DDL 문서는 초기화, 실행, commit table definition의 세 단계에서 metadata lock이 쓰이고, 마지막 definition commit 단계에서 exclusive metadata lock으로 upgrade된다고 설명한다. 더 중요한 운영 함정은 pending exclusive metadata lock이다. 오래 열린 transaction 때문에 online DDL이 exclusive metadata lock을 기다리면, 그 뒤에 들어온 평범한 `SELECT`까지 DDL 뒤에서 줄을 설 수 있다.

이 주제가 별도 절로 필요한 배경은 InnoDB가 “빠른 읽기”와 “범위 정합성”을 동시에 만족시키려다 여러 종류의 읽기와 잠금을 만들어 냈기 때문이다. 일반 consistent read는 undo와 read view로 과거 committed version을 보여 주어 reader/writer 충돌을 줄인다. 그러나 재고 차감, 예약 범위 검사, unique/foreign key 검증, `SELECT ... FOR UPDATE` 같은 경로에서는 단순히 과거 snapshot을 보여 주는 것만으로는 충분하지 않다. 누군가는 현재 record나 index gap에 대한 권리를 잡아야 하고, DDL은 table definition 자체가 바뀌는 순간을 metadata lock으로 보호해야 한다. online DDL의 등장은 이 보호를 없애기 위한 것이 아니라, 가능한 작업에서 rebuild와 copy 시간을 길게 가져가더라도 짧고 강한 잠금 구간을 줄이기 위한 운영상의 타협으로 읽어야 한다.

### 격리 수준은 이름보다 읽기 종류가 중요하다

MySQL 공식 transaction isolation 문서는 InnoDB가 각 isolation level을 서로 다른 locking strategy로 지원한다고 설명한다. 기본값인 REPEATABLE READ는 중요한 데이터 작업에서 높은 일관성을 제공하고, READ COMMITTED는 locking overhead를 낮추는 방향의 선택지다. 하지만 실무에서 더 자주 터지는 문제는 isolation level 이름만 보고 쿼리의 읽기 종류를 구분하지 않는 것이다. 같은 transaction 안에서도 일반 `SELECT`는 consistent read이고, `SELECT ... FOR UPDATE`는 locking read이며, `UPDATE`는 대상 row와 index range에 lock을 건다. 일반 `SELECT`가 snapshot을 보는 동안 `UPDATE`는 현재 version과 lock conflict를 본다. 그래서 “REPEATABLE READ니까 이 transaction 안에서는 모든 쿼리가 같은 snapshot만 본다”는 말은 부정확하다.

다음 schedule을 보자.

```text
Session A                                  Session B
----------------------------------------   --------------------------------
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
START TRANSACTION;
SELECT stock FROM item WHERE id = 10; -- sees 5
                                           UPDATE item SET stock = 4 WHERE id = 10;
                                           COMMIT;
SELECT stock FROM item WHERE id = 10; -- still sees 5 as consistent read
SELECT stock FROM item WHERE id = 10 FOR UPDATE;
-- locking read waits or reads the current committed version after locking path
COMMIT;
```

이 예시에서 일반 `SELECT`와 locking read를 같은 “읽기”로 묶으면 판단이 깨진다. 일반 `SELECT`는 transaction timepoint에 맞는 snapshot을 보지만, `FOR UPDATE`는 앞으로 update할 현재 row를 잠가야 한다. 이 차이를 모르면 애플리케이션에서 재고, 잔액, 상태 전이 같은 값을 “읽고 판단한 뒤 업데이트”하는 코드가 snapshot read와 current read를 섞어 버린다. MySQL 공식 문서도 REPEATABLE READ transaction 안에서 locking statement와 nonlocking `SELECT`를 섞는 것을 권장하지 않는다고 경고한다. 보통 그런 경우에는 SERIALIZABLE이나 더 명시적인 lock 전략을 검토해야 한다.

### InnoDB row lock, intention lock, gap lock을 한 장면으로 보기

InnoDB locking 문서는 row-level locking의 기본 lock으로 shared lock과 exclusive lock을 설명한다. Shared lock은 row를 읽는 것을 허용하고, exclusive lock은 row를 update/delete하는 것을 허용한다. 같은 row에 shared lock은 함께 잡힐 수 있지만, 다른 transaction이 exclusive lock을 요청하면 기다려야 한다. 그런데 InnoDB lock은 row만으로 끝나지 않는다. table-level intention lock은 row lock과 table lock이 공존할 수 있게 “이 transaction은 이 table 안 row에 S 또는 X lock을 잡을 예정”이라는 의도를 표시한다. gap lock은 index record 사이의 빈 공간에 insert가 들어오는 것을 막는 억제용 lock이다. 공식 문서는 unique index로 unique row 하나를 찾는 경우에는 gap lock이 필요하지 않지만, index가 없거나 nonunique index면 preceding gap을 잠글 수 있다고 설명한다.

다음은 gap lock을 이해하기 위한 단순 trace다.

```sql
CREATE TABLE seat_hold (
  show_id BIGINT NOT NULL,
  seat_no INT NOT NULL,
  user_id BIGINT,
  PRIMARY KEY (show_id, seat_no),
  KEY idx_user (user_id)
) ENGINE=InnoDB;
```

```text
현재 clustered index key:
  (10, 1)   (10, 2)   (10, 5)

Session A:
  START TRANSACTION;
  SELECT * FROM seat_hold
  WHERE show_id = 10 AND seat_no BETWEEN 2 AND 5
  FOR UPDATE;

잠금 감각:
  record (10,2) locked
  gap between (10,2) and (10,5) protected
  record (10,5) locked

Session B:
  INSERT INTO seat_hold(show_id, seat_no, user_id)
  VALUES (10, 3, 200);
  -> may wait because the gap is protected
```

이 lock은 “존재하는 row만 보호한다”는 모델로는 설명되지 않는다. 범위 조건의 의미를 보존하려면 아직 존재하지 않는 `(10, 3)`이 끼어드는 것도 막아야 한다. 이것이 phantom 방지와 next-key locking을 이해하는 첫 발판이다. 하지만 모든 상황에서 gap lock이 같은 방식으로 걸린다고 외우면 또 틀린다. isolation level, index shape, unique lookup 여부, search condition이 composite index 전체를 쓰는지 일부만 쓰는지에 따라 달라진다. 운영에서 중요한 것은 “왜 이 statement가 이 row만이 아니라 이 range를 막았는가”를 execution plan과 index shape로 다시 확인하는 습관이다.

### Local seed: MySQL 5.5 대형 인덱스 추가 문제에서 배울 것

`database/mysql/problems/mysql5_online_large_indexing.md`의 seed는 아주 현실적인 질문에서 시작한다. 한 table은 약 4억 4천만 row, `DATA_LENGTH` 약 63.91GB, `INDEX_LENGTH` 약 29.3GB이고, 다른 table은 약 1억 5천만 row, `DATA_LENGTH` 약 15.79GB, `INDEX_LENGTH` 약 16.86GB다. 서비스 핵심 DB에서 index 추가를 문의받았고, 문제가 생기면 서비스 전체 장애가 날 수 있는 상황이다. 이 seed는 오래된 MySQL 5.5를 언급하므로 MySQL 8.4 Online DDL 기능을 그대로 적용해 단정하면 안 된다. 하지만 운영 판단의 핵심은 여전히 살아 있다. 대형 table index build는 CPU, disk I/O, buffer pool, redo/temporary space, metadata lock, replication lag, rollback/kill 비용을 모두 건드린다.

이 seed를 DU39 관점으로 승격하면 질문은 “인덱스를 온라인으로 추가할 수 있는가”가 아니라 다음 다섯 질문으로 바뀐다.

| 질문 | 왜 중요한가 | 관측 또는 검증 |
|---|---|---|
| 어떤 MySQL 버전과 engine인가 | Online DDL 지원 범위가 버전과 engine에 따라 다르다 | `SELECT VERSION()`, `SHOW TABLE STATUS` |
| 어떤 algorithm/lock이 실제 선택되는가 | `LOCK=NONE` 의도와 실제 실행이 다를 수 있다 | `ALTER TABLE ... ALGORITHM=..., LOCK=...` dry-run 성격의 syntax 확인, staging |
| metadata lock 대기가 있는가 | pending DDL이 뒤 query까지 막을 수 있다 | `performance_schema.metadata_locks`, `SHOW PROCESSLIST` |
| index build I/O가 감당 가능한가 | 큰 table scan/sort/write가 buffer pool과 disk를 흔든다 | I/O, redo, temp, buffer pool metrics |
| 실패 시 rollback/재시도 비용은 얼마인가 | kill이 즉시 원상복구가 아닐 수 있다 | staging elapsed time, replica test, backup/restore plan |

이렇게 질문을 바꾸면 “트래픽 낮은 시간에 하자”라는 단일 조언보다 훨씬 안전하다. 낮은 시간대는 하나의 완화책일 뿐이고, 실제 안전성은 lock 관측, build 비용 예측, 실패 시 되돌림 경로, replication 영향까지 닫혀야 한다.

### Online DDL은 lock-free가 아니다

MySQL 공식 Online DDL Operations 표는 secondary index 생성이 `INPLACE` 가능하고 table rebuild는 하지 않으며 concurrent DML을 허용할 수 있다고 보여준다. 이 문장만 보면 “그럼 서비스 중에 index 추가해도 괜찮다”고 착각하기 쉽다. 그러나 Online DDL Performance and Concurrency 문서는 online DDL이 세 단계로 볼 수 있다고 설명한다. 초기화 단계에서는 허용 가능한 concurrency를 판단하고 shared upgradeable metadata lock을 잡는다. 실행 단계에서는 작업을 준비하고 수행한다. 마지막 table definition commit 단계에서는 old table definition을 내보내고 새 definition을 commit하기 위해 metadata lock을 exclusive로 upgrade한다. exclusive metadata lock이 필요한 시간이 짧더라도, 그 lock을 얻기 위해 오래 기다릴 수 있다는 점이 운영 함정이다.

다음 schedule은 공식 문서의 설명을 운영자가 보는 형태로 재구성한 것이다.

```text
Session 1                                  Session 2                                      Session 3
---------------------------------------    --------------------------------------------   --------------------------
START TRANSACTION;
SELECT * FROM big_orders WHERE id=1;
-- shared metadata lock held until trx end
                                           ALTER TABLE big_orders
                                           ADD INDEX idx_created_at(created_at),
                                           ALGORITHM=INPLACE, LOCK=NONE;
                                           -- waits for exclusive MDL at commit phase
                                                                                         SELECT * FROM big_orders
                                                                                         WHERE id=2;
                                                                                         -- blocked behind pending exclusive MDL
COMMIT;
                                           -- obtains exclusive MDL briefly, commits DDL
                                                                                         -- then proceeds
```

여기서 가장 무서운 점은 Session 3이다. Session 3은 DDL을 실행하지 않았다. 평범한 조회일 수 있다. 그런데 pending exclusive metadata lock이 queue 앞에 있으면 뒤에서 기다릴 수 있다. 사용자는 “online DDL이라 DML 허용이라고 했는데 왜 SELECT가 멈췄지?”라고 느낀다. 답은 online DDL이 “전체 작업 동안 table을 계속 독점한다”는 뜻은 아니지만, “metadata lock이 전혀 없고 후속 query를 막지 않는다”는 뜻도 아니라는 데 있다.

### Metadata lock은 performance_schema로 관계를 봐야 한다

MySQL metadata locking 문서는 metadata lock이 database object에 대한 concurrent access를 관리하고 consistency를 보장하기 위한 장치라고 설명한다. table뿐 아니라 schema, stored program, tablespace, user lock에도 적용된다. Performance Schema의 `metadata_locks` table은 granted lock, pending lock, deadlock victim, timeout 같은 상태를 보여 줄 수 있고, 어떤 session이 어떤 lock을 들고 있으며 어떤 session이 무엇을 기다리는지 이해하는 데 도움을 준다. 이 관측면을 쓰지 않고 `SHOW PROCESSLIST`의 `Waiting for table metadata lock` 한 줄만 보면 원인 session을 놓치기 쉽다.

운영에서 쓸 수 있는 최소 관측 query는 다음과 같다.

```sql
SELECT
  ml.OBJECT_SCHEMA,
  ml.OBJECT_NAME,
  ml.LOCK_TYPE,
  ml.LOCK_DURATION,
  ml.LOCK_STATUS,
  th.PROCESSLIST_ID,
  th.PROCESSLIST_USER,
  th.PROCESSLIST_HOST,
  th.PROCESSLIST_DB,
  th.PROCESSLIST_COMMAND,
  th.PROCESSLIST_TIME,
  th.PROCESSLIST_STATE,
  th.PROCESSLIST_INFO
FROM performance_schema.metadata_locks ml
JOIN performance_schema.threads th
  ON ml.OWNER_THREAD_ID = th.THREAD_ID
WHERE ml.OBJECT_SCHEMA = 'app'
  AND ml.OBJECT_NAME = 'big_orders'
ORDER BY FIELD(ml.LOCK_STATUS, 'PENDING', 'GRANTED'), th.PROCESSLIST_TIME DESC;
```

PASS 신호는 `PENDING` lock과 그 앞의 `GRANTED` lock을 같은 object 기준으로 연결하는 것이다. 예를 들어 long transaction의 `SELECT`가 table definition에 shared metadata lock을 들고 있고, `ALTER TABLE`이 exclusive metadata lock을 기다리며, 뒤 query들이 줄줄이 대기한다면 장애 원인은 “SELECT가 느리다”가 아니라 “오래 열린 transaction이 DDL commit을 막고, pending DDL이 뒤 요청을 막는 queue 구조”다. FAIL 신호는 processlist에서 가장 오래 걸린 query만 죽이거나, DDL session만 kill하고 원인 transaction을 보지 않는 것이다. DDL kill 자체도 비용이 있을 수 있으므로, 어떤 session을 종료할지 결정하기 전에 transaction age, binlog/replication 영향, application retry 가능성을 같이 봐야 한다.

### Locking read와 write path를 구분한 장애 분석

InnoDB lock 장애에서 흔한 실수는 “어떤 query가 느리다”에서 바로 index 부족으로 가는 것이다. 물론 index 부족은 매우 흔한 원인이다. 하지만 lock wait에서는 query plan과 lock footprint를 함께 봐야 한다. 예를 들어 다음 query는 같은 business 의미처럼 보여도 lock footprint가 다르다.

```sql
-- unique lookup, PK 전체를 사용한다.
SELECT * FROM payment WHERE payment_id = 100 FOR UPDATE;

-- tenant_id만 사용하고 status 범위로 훑는다.
SELECT * FROM payment
WHERE tenant_id = 10 AND status = 'READY'
FOR UPDATE;

-- index가 없거나 leading column을 놓치면 더 넓은 scan과 lock footprint가 생긴다.
UPDATE payment
SET status = 'PROCESSING'
WHERE status = 'READY'
ORDER BY created_at
LIMIT 100;
```

첫 번째 query는 unique key 전체를 사용해 하나의 record lock에 가까운 형태로 좁혀질 가능성이 높다. 두 번째 query는 `(tenant_id, status, created_at)` 같은 적절한 index가 있으면 범위를 줄일 수 있지만, index가 없거나 leading column 순서가 맞지 않으면 훨씬 넓은 범위를 읽고 잠글 수 있다. 세 번째 query는 batch claim 패턴에서 자주 보인다. 이때 `ORDER BY created_at LIMIT 100`이 있더라도 index가 맞지 않으면 많은 row를 훑고, lock wait과 deadlock 가능성이 커진다. 따라서 lock 분석은 `SHOW ENGINE INNODB STATUS`, `performance_schema.data_locks`, `data_lock_waits`, `EXPLAIN`, 실제 index definition을 함께 놓고 본다.

```sql
SELECT
  dl.ENGINE_TRANSACTION_ID,
  dl.OBJECT_SCHEMA,
  dl.OBJECT_NAME,
  dl.INDEX_NAME,
  dl.LOCK_TYPE,
  dl.LOCK_MODE,
  dl.LOCK_STATUS,
  dl.LOCK_DATA
FROM performance_schema.data_locks dl
WHERE dl.OBJECT_SCHEMA = 'app'
  AND dl.OBJECT_NAME = 'payment'
ORDER BY dl.ENGINE_TRANSACTION_ID, dl.INDEX_NAME, dl.LOCK_DATA;

SELECT *
FROM performance_schema.data_lock_waits;
```

이 query들의 목적은 “잠금 목록을 예쁘게 출력”하는 것이 아니라 lock이 어느 index와 어느 key range에 걸렸는지 확인하는 것이다. `INDEX_NAME`이 기대한 index와 다르거나, `LOCK_DATA`가 넓은 range를 암시하거나, wait graph에서 같은 table과 index가 반복되면 query shape와 index 설계를 함께 고쳐야 한다. 반대로 lock wait가 metadata lock이라면 row lock query만 계속 봐도 답이 나오지 않는다. 이 구분이 실무 장애 대응의 첫 번째 분기다.

### Isolation level 변경은 치료제가 아니라 계약 변경이다

Gap lock 때문에 insert가 막히면 READ COMMITTED로 바꾸고 싶어진다. MySQL 공식 locking 문서는 READ COMMITTED에서는 search와 index scan에 대한 gap locking이 비활성화되고 foreign-key constraint checking과 duplicate-key checking에만 쓰인다고 설명한다. 이 정보는 유용하지만, isolation level 변경은 단순 성능 옵션이 아니다. REPEATABLE READ에서 기대하던 repeatable snapshot, phantom 방지 성질, statement-based replication 안전성, 애플리케이션의 재시도 정책이 함께 바뀐다. 어떤 batch processor가 “READY row를 읽어서 PROCESSING으로 바꾼다”는 계약을 갖고 있다면 READ COMMITTED에서 같은 row를 두 worker가 잡지 않도록 `FOR UPDATE SKIP LOCKED` 같은 명시적 claim 전략이나 unique state transition 조건이 필요할 수 있다.

따라서 isolation 변경을 검토할 때는 두 가지 후보를 같은 수준에서 비교해야 한다.

| 후보 | 장점 | 위험 | 검증 |
|---|---|---|---|
| REPEATABLE READ 유지 + index/claim query 개선 | 기존 일관성 계약을 덜 흔든다 | gap/next-key lock footprint가 여전히 남을 수 있다 | wait graph, EXPLAIN, deadlock 재현 |
| READ COMMITTED 전환 + 명시적 claim/재시도 설계 | gap lock 경합을 줄일 수 있다 | snapshot/phantom/반복 읽기 계약이 달라진다 | 동시 worker 테스트, 중복 claim 방지, 재시도 정책 |

이 비교를 하지 않고 isolation level을 낮추면 장애는 줄어든 것처럼 보이다가 정산, 재고, 쿠폰 발급 같은 invariant에서 중복 처리나 누락으로 돌아올 수 있다. 반대로 무조건 REPEATABLE READ를 유지하겠다고 고집하면 hot range insert와 batch update가 불필요하게 막힐 수 있다. 좋은 판단은 “어떤 invariant를 보호해야 하는가”를 먼저 말하고, 그 invariant를 lock으로 지킬지, unique constraint와 idempotent state transition으로 지킬지, retry 가능한 transaction으로 지킬지 비교하는 것이다.

### Online index build 실행 전 체크리스트

대형 table에 index를 추가하기 전에는 SQL 한 줄보다 실행 전제와 관측 계획이 중요하다. 다음 체크리스트는 DU39의 local seed를 실제 운영 작업으로 바꾸기 위한 최소 뼈대다.

```text
대상:
  app.big_orders
규모:
  rows ~= 439,912,614
  data_length ~= 63.91 GB
  index_length ~= 29.3 GB
DDL:
  ALTER TABLE app.big_orders ADD INDEX idx_created_at(created_at), ALGORITHM=INPLACE, LOCK=NONE;

사전 확인:
  [ ] MySQL version과 InnoDB engine 확인
  [ ] staging에서 같은 cardinality 또는 축소 비율 실험
  [ ] EXPLAIN으로 새 index가 실제로 쿼리에 필요한지 확인
  [ ] backup/restore 또는 replica rebuild 경로 확인
  [ ] disk free space, redo capacity, temp space 확인
  [ ] long transaction / idle transaction 확인
  [ ] metadata_locks, data_locks, processlist 관측 query 준비
  [ ] replication lag와 replica apply capacity 확인
  [ ] kill/rollback 시나리오와 의사결정자 지정
```

실행 중에는 다음 순서로 본다.

```sql
SHOW PROCESSLIST;

SELECT *
FROM performance_schema.metadata_locks
WHERE OBJECT_SCHEMA = 'app'
  AND OBJECT_NAME = 'big_orders';

SELECT EVENT_NAME, WORK_COMPLETED, WORK_ESTIMATED
FROM performance_schema.events_stages_current
WHERE EVENT_NAME LIKE 'stage/innodb/alter table%';

SHOW ENGINE INNODB STATUS\G
```

PASS 신호는 DDL이 예상한 algorithm과 lock mode로 진행되고, metadata lock 대기가 짧으며, replication lag와 I/O가 사전에 정한 threshold 안에 있고, application error rate가 변하지 않는 것이다. FAIL 신호는 metadata lock pending이 생긴 뒤 신규 query가 queue 뒤에 쌓이거나, replication lag가 회복 불가능하게 벌어지거나, disk/redo/temp 공간이 임계치에 접근하거나, DDL progress가 멈췄는데 원인 session을 식별하지 못하는 것이다. 이 경우 “조금만 더 기다리면 되겠지”가 아니라 사전에 정한 rollback/kill/traffic drain 기준으로 움직여야 한다.

### Deadlock과 lock wait를 학습용으로 재현하기

운영 장애를 처음 보는 순간에 deadlock graph를 읽으려 하면 늦다. 작은 테이블로 먼저 재현해 보면 InnoDB의 wait graph 감각이 생긴다.

```sql
CREATE TABLE account_balance (
  account_id BIGINT NOT NULL,
  balance BIGINT NOT NULL,
  PRIMARY KEY (account_id)
) ENGINE=InnoDB;

INSERT INTO account_balance VALUES (1, 1000), (2, 1000);
```

```text
Session A                                  Session B
---------------------------------------    ---------------------------------------
START TRANSACTION;                         START TRANSACTION;
UPDATE account_balance
SET balance = balance - 100
WHERE account_id = 1;
                                           UPDATE account_balance
                                           SET balance = balance - 100
                                           WHERE account_id = 2;
UPDATE account_balance
SET balance = balance + 100
WHERE account_id = 2;
-- waits for B
                                           UPDATE account_balance
                                           SET balance = balance + 100
                                           WHERE account_id = 1;
                                           -- deadlock detected, one transaction rolls back
```

이 실험의 핵심은 두 transaction이 같은 두 row를 반대 순서로 잡았다는 것이다. 해결은 “timeout을 늘린다”가 아니라 lock acquisition order를 통일하거나, single statement update로 바꾸거나, application retry를 transaction 단위로 설계하는 것이다. `SHOW ENGINE INNODB STATUS\G`의 latest detected deadlock에는 어떤 transaction이 어떤 lock을 기다렸는지, 어떤 index record가 관련되었는지 단서가 나온다. 이 단서를 query와 index definition에 연결해야 한다. deadlock은 InnoDB가 망가졌다는 뜻이 아니라 동시성 제어가 둘 중 하나를 희생해 cycle을 끊었다는 뜻이고, 애플리케이션은 retry 가능한 단위로 설계되어야 한다.

### Misconception repair: online, nonlocking, concurrent의 경계

`nonlocking read`는 일반 `SELECT`가 table에 lock을 set하지 않는다는 뜻이지, 읽기가 내부 비용 없이 공짜라는 뜻이 아니다. Snapshot보다 최신 record를 만나면 undo를 따라 과거 version을 재구성해야 하고, 오래된 snapshot은 purge를 늦춘다.

`LOCK=NONE`은 DDL 실행 중 concurrent DML을 허용하려는 요청이지, metadata lock이 전혀 없다는 보장이 아니다. 초기화와 commit definition 단계의 metadata lock, exclusive upgrade 대기, pending DDL 뒤 query blocking을 반드시 고려해야 한다.

`INPLACE`는 항상 instant라는 뜻이 아니다. MySQL 공식 Online DDL 성능 문서는 DDL 성능이 instant인지, in-place인지, table rebuild를 하는지에 크게 좌우된다고 설명한다. Secondary index 추가는 table copy를 하지 않을 수 있지만 시간이 오래 걸리고 physical read/write와 progress monitoring이 필요하다.

`row-level locking`은 table-level 영향이 없다는 뜻이 아니다. Intention lock, metadata lock, gap lock, foreign key check, duplicate key check가 table과 range 차원에서 같이 작동한다. 장애 대응에서는 row lock과 metadata lock을 먼저 나누고, 그 다음 index range와 transaction age를 본다.

### 실무 설계로 가져가는 원리

InnoDB locking과 online DDL을 안전하게 다루는 팀은 보통 세 가지 원칙을 갖고 있다. 첫째, business invariant를 먼저 말한다. 예를 들어 “한 쿠폰은 한 번만 지급된다”가 invariant라면 lock으로 막을지, unique key로 막을지, idempotency key로 replay를 안전하게 만들지 비교한다. 둘째, query가 어떤 index range를 잠그는지 사전에 본다. `WHERE status='READY' LIMIT 100` 같은 query는 보기에는 작아도 index가 없으면 큰 lock footprint가 된다. 셋째, DDL은 배포 작업처럼 다룬다. migration SQL, 관측 query, abort 기준, replication 확인, backup/restore 경로, application retry 정책이 함께 있어야 한다.

이 원리는 MySQL 5.5 같은 오래된 환경에서는 더 강해진다. 공식 8.4 문서의 online DDL 지원 범위가 넓어졌다고 해서 legacy version에 같은 결론을 이식하면 안 된다. 버전 차이가 있으면 “이 operation이 instant/inplace/copy 중 무엇인가”, “concurrent DML을 허용하는가”, “metadata lock이 어느 단계에서 필요한가”를 해당 버전 문서나 staging 실험으로 다시 닫아야 한다. 이 절에서 8.4 문서를 근거로 든 이유는 현재 공식 reference의 개념과 관측 surface를 잡기 위해서이고, legacy 운영 결론은 local seed의 MySQL 5.5 맥락에 맞게 낮춰 말해야 한다.

### Application code에서 자주 생기는 lock footprint 확대

DB lock 문제는 SQL 한 줄보다 application transaction boundary에서 커지는 경우가 많다. 예를 들어 Spring service가 transaction을 시작한 뒤 외부 API를 호출하고, 그 뒤에 `UPDATE payment SET status='DONE'`을 실행한다면 row lock을 잡는 시점은 뒤쪽일 수 있지만 transaction은 이미 오래 열려 있다. 반대로 먼저 `SELECT ... FOR UPDATE`로 row를 잡고 외부 API를 기다리면 lock hold time이 외부 network latency에 묶인다. 이 경우 InnoDB 문제가 아니라 transaction 설계 문제가 lock wait로 보인다.

```text
나쁜 흐름:
  BEGIN
    SELECT payment WHERE id=10 FOR UPDATE  -> row lock acquired
    call external PSP API                  -> 800 ms or timeout
    UPDATE payment SET status='DONE'
  COMMIT

더 안전한 흐름 후보:
  BEGIN
    claim payment with short UPDATE condition
  COMMIT
  call external PSP API outside DB lock
  BEGIN
    finalize with idempotency/result check
  COMMIT
```

두 번째 흐름이 항상 정답은 아니다. 외부 호출과 DB 상태 전이를 강하게 묶어야 하는 도메인도 있다. 하지만 lock hold time을 외부 latency와 분리할 수 있는지 먼저 검토해야 한다. InnoDB의 row lock은 짧은 critical section을 보호할수록 강하다. Business transaction 전체를 DB transaction 하나에 밀어 넣으면 gap lock, row lock, metadata lock, undo retention이 모두 길어진다.

### DDL 작업의 dry-run은 무엇을 확인해야 하는가

MySQL에는 모든 DDL을 완전히 무해하게 dry-run하는 단일 명령이 있는 것은 아니다. 그래서 운영자는 “실행 전 확인”을 여러 층으로 나눠야 한다. 첫째, 문법과 지원 algorithm을 version 문서로 확인한다. 둘째, staging에서 같은 schema와 비슷한 row distribution으로 elapsed time, rows affected, progress instrumentation을 본다. 셋째, production에서는 실제 실행 전에 long transaction과 metadata lock 후보를 비운다. 넷째, DDL 시작 뒤에는 progress와 wait를 분리해 본다.

```sql
-- 실행 전 long transaction 후보
SELECT
  trx_id,
  trx_started,
  TIMESTAMPDIFF(SECOND, trx_started, NOW()) AS trx_age_seconds,
  trx_state,
  trx_query
FROM information_schema.innodb_trx
ORDER BY trx_started;

-- metadata lock instrumentation이 켜져 있는지 확인
SELECT NAME, ENABLED, TIMED
FROM performance_schema.setup_instruments
WHERE NAME = 'wait/lock/metadata/sql/mdl';
```

이 확인의 PASS는 “문제가 없을 것 같다”가 아니라 “문제가 생기면 어떤 query로 어떤 상태를 볼지 준비되었다”다. DDL이 기다릴 때 row lock인지 metadata lock인지, DDL 자체가 진행 중인지 commit definition phase에서 막혔는지, 뒤 요청이 pending exclusive lock 뒤에 막히는지 분리할 수 있어야 한다. 이 분리가 안 되면 장애 중에 가장 위험한 행동인 blind kill, blind retry, blind scale-out을 하게 된다.

### 출처와 검증 경계

이 절의 공식 근거는 MySQL 8.4 Reference Manual의 [InnoDB Locking](https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html), [Transaction Isolation Levels](https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html), [Consistent Nonlocking Reads](https://dev.mysql.com/doc/refman/8.4/en/innodb-consistent-read.html), [Online DDL Operations](https://dev.mysql.com/doc/refman/8.4/en/innodb-online-ddl-operations.html), [Online DDL Performance and Concurrency](https://dev.mysql.com/doc/refman/8.4/en/innodb-online-ddl-performance.html), [Metadata Locking](https://dev.mysql.com/doc/refman/8.4/en/metadata-locking.html), [Performance Schema metadata_locks](https://dev.mysql.com/doc/refman/8.4/en/performance-schema-metadata-locks-table.html)다. 로컬 seed는 `database/mysql/problems/mysql5_online_large_indexing.md`와 기존 MySQL 운영 메모다. 이 절의 schedule과 lock trace는 학습용 단순화이며, 실제 lock mode 표기와 record 내용은 MySQL version, index shape, isolation level, optimizer plan에 따라 달라진다. 검증은 staging에서 같은 DDL과 대표 transaction을 실행하고, `SHOW PROCESSLIST`, `performance_schema.metadata_locks`, `performance_schema.data_locks`, `SHOW ENGINE INNODB STATUS`, replication lag 지표가 예상한 queue 구조와 일치하는지 확인하는 방식으로 닫아야 한다.
