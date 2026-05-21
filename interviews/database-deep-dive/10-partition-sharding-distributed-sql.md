# 파티셔닝, 샤딩, 분산 SQL은 어떤 확장 문제를 서로 다르게 푸는가?

대용량 DB 질문에서 "테이블을 나누면 됩니다"라고 답하면 바로 다음 질문이 따라옵니다. 어떤 행이 어느 조각에 들어가나요. 쿼리는 그중 필요한 조각만 읽을 수 있나요. 조각 하나가 너무 커지거나 특정 조각에 쓰기가 몰리면 어떻게 다시 나누나요. 여러 조각을 한 transaction이 함께 바꾸면 누가 commit을 조율하나요. 파티션 pruning, sharding, resharding, hot shard, consensus 같은 용어는 모두 이 질문들에 붙는 공식 이름입니다.

이 문서는 네 가지 선택지를 분리합니다. Native partition은 한 DBMS 안에서 큰 table을 더 작은 물리 단위로 나누는 방법이고, 수동 파티셔닝은 여러 table과 애플리케이션 분기로 비슷한 효과를 직접 만드는 방식입니다. Sharding은 여러 노드에 데이터 소유권을 나누는 방법입니다. Distributed SQL은 SQL 표면과 transaction 경험을 유지하려고 분산 저장과 합의 프로토콜을 DBMS 내부에 넣은 계열입니다. 모두 "큰 데이터를 나눈다"는 공통점은 있지만, query planner, routing, transaction, 운영 책임의 위치가 다릅니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [Native partition pruning과 lifecycle](#native-partition-pruning과-lifecycle)
    - [수동 파티셔닝과 drift](#수동-파티셔닝과-drift)
    - [Sharding: key, routing table, resharding](#sharding-key-routing-table-resharding)
    - [Cross-shard query와 transaction](#cross-shard-query와-transaction)
    - [Distributed SQL: range, consensus, retry, locality](#distributed-sql-range-consensus-retry-locality)
    - [분산 DB write도 결국 커널과 디스크를 지난다](#분산-db-write도-결국-커널과-디스크를-지난다)
- [DBMS별 경계](#dbms별-경계)
- [직접 재생해 보기](#직접-재생해-보기)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
    - [마지막 연결: 데이터를 나눴을 때 query와 transaction이 어디까지 한 조각에 머무는가](#마지막-연결-데이터를-나눴을-때-query와-transaction이-어디까지-한-조각에-머무는가)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

면접에서 먼저 답한다면 이렇게 말할 수 있습니다. 파티셔닝은 보통 하나의 논리 table을 날짜, tenant, hash 같은 기준으로 여러 partition에 나누어 저장하는 방법입니다. 좋은 partition key가 WHERE 조건에 들어오면 DB optimizer가 불필요한 partition을 읽지 않는 partition pruning을 할 수 있습니다. 하지만 partition key가 query predicate에 없거나, 함수로 감싸져 planner가 범위를 판단하지 못하거나, 너무 많은 partition이 생기면 오히려 planning과 maintenance 비용이 커질 수 있습니다.

수동 파티셔닝은 DB native partition 기능을 쓰지 않고 `orders_2026_05`, `orders_2026_06` 같은 여러 table을 애플리케이션이나 view로 묶는 방식입니다. 특정 환경에서는 migration이나 archiving을 쉽게 만들 수 있지만, query routing, union view, index 일관성, constraint, cross-partition aggregation, application code drift를 직접 책임져야 합니다. 따라서 수동 파티셔닝은 임시 절약책이 아니라 명시적인 운영 설계여야 합니다.

샤딩은 데이터를 여러 DB 노드에 분산합니다. 예를 들어 `user_id % 16`으로 shard를 고르면 단일 DB의 storage와 write 한계를 넘을 수 있습니다. 하지만 shard key를 잘못 고르면 hot shard가 생기고, cross-shard join이나 cross-shard transaction이 비싸지며, resharding은 데이터 이동과 dual-write 또는 routing table 변경을 동반합니다. 샤딩의 본질은 성능 옵션이 아니라 데이터 소유권과 query boundary를 애플리케이션 또는 middleware가 감당하는 구조입니다.

Distributed SQL 또는 NewSQL은 SQL과 ACID transaction 경험을 유지하면서 내부적으로 데이터를 range나 tablet으로 나누고, replica 간 consensus로 durability와 failover를 처리하려는 계열입니다. CockroachDB, YugabyteDB, Google Spanner 같은 제품은 세부가 다르지만 공통적으로 분산 합의, timestamp ordering, transaction retry, locality 설계를 비용으로 가집니다. "RDBMS처럼 쓰면서 무한 확장"이 아니라 "일부 관계형 경험을 보존하되 network round trip과 consensus 비용을 관리하는 시스템"으로 봐야 합니다.

좋은 답변은 선택 기준으로 닫힙니다. 단일 노드 table이 너무 커져 관리와 pruning이 문제라면 partition을 먼저 봅니다. 특정 tenant나 user key로 부하를 여러 노드에 나누어야 하면 sharding을 봅니다. Cross-region availability, horizontal write scale, SQL compatibility를 동시에 원하면 distributed SQL을 검토하되 latency, transaction retry, locality, 운영 복잡도를 함께 계산합니다. 어떤 선택이든 query shape와 transaction boundary가 key 설계를 따라가야 합니다.

## 먼저 잡아야 할 작은 모델

작은 모델은 주문 table이 매달 1억 행씩 늘어나는 상황입니다.

```sql
CREATE TABLE orders (
order_id bigint PRIMARY KEY,
user_id bigint NOT NULL,
created_at timestamp NOT NULL,
status text NOT NULL,
amount numeric(12,2) NOT NULL
);
```

가장 흔한 query는 최근 한 달 주문 목록입니다.

```sql
SELECT order_id, status, amount
FROM orders
WHERE created_at >= timestamp '2026-05-01'
  AND created_at <  timestamp '2026-06-01'
ORDER BY created_at DESC
LIMIT 100;
```

Partition이 없으면 DB는 전체 table의 index나 heap에서 조건에 맞는 범위를 찾아야 합니다. `created_at` index가 있으면 상당히 줄일 수 있지만, 오래된 데이터를 drop/archive하거나 vacuum/analyze/manage하는 비용은 여전히 큰 table 하나에 몰릴 수 있습니다. 월별 range partition을 쓰면 `created_at` 조건으로 2026년 5월 partition만 읽게 할 수 있습니다.

```text
logical table: orders

physical partitions:
  orders_2026_03  created_at >= 2026-03-01 and < 2026-04-01
  orders_2026_04  created_at >= 2026-04-01 and < 2026-05-01
  orders_2026_05  created_at >= 2026-05-01 and < 2026-06-01
  orders_2026_06  created_at >= 2026-06-01 and < 2026-07-01

query predicate:
  created_at >= 2026-05-01 and < 2026-06-01

planner can prune:
  scan only orders_2026_05
```

이 예제에서 pruning은 "5월 데이터가 더 적으니까 빠르다"가 아니라 planner가 물리 조각 후보를 먼저 줄이는 일입니다. 같은 주문 100건을 읽더라도, 후보 partition을 줄이는 단계와 partition 안에서 index를 타는 단계는 서로 다릅니다.

| 단계 | DBMS가 보는 질문 | 5월 조회에서 남는 후보 |
| --- | --- | --- |
| 1. partition bound 비교 | `created_at` 조건이 어느 partition 범위와 겹치는가 | `orders_2026_05` |
| 2. partition 안 access path 선택 | 남은 partition 안에서 index scan, sequential scan, sort 중 무엇을 쓸 것인가 | 5월 partition의 `created_at` index 또는 heap |
| 3. row visibility와 정렬 | 조건에 맞고 현재 transaction에서 보이는 row를 어떤 순서로 돌려줄 것인가 | 최신 100건 |

이 표가 중요한 이유는 partition이 index를 대체하지 않기 때문입니다. Partition은 "어느 큰 조각을 볼 것인가"를 줄이고, index는 "그 조각 안에서 어느 row로 내려갈 것인가"를 줄입니다. 그래서 월별 partition만 있고 5월 partition 안에 적절한 index가 없으면, 5월 partition 전체를 훑는 계획이 나올 수 있습니다.

Sharding은 다른 모델입니다. 월별 partition이 같은 DB instance 안에 있을 수 있다면, shard는 여러 DB node에 나뉩니다.

```text
shard key: user_id % 4

shard 0: user_id 4, 8, 12, ...
shard 1: user_id 1, 5, 9, ...
shard 2: user_id 2, 6, 10, ...
shard 3: user_id 3, 7, 11, ...

query by user_id=1001
  router computes 1001 % 4 = 1
  send query only to shard 1

query by created_at last hour without user_id
  router cannot pick one shard
  scatter to all shards and merge results
```

같은 query라도 `WHERE user_id = ?`가 들어오느냐에 따라 움직이는 범위가 달라집니다.

```text
single-shard query
  request(user_id=1001)
    -> router: shard 1
    -> shard 1 local index scan
    -> response

scatter-gather query
  request(created_at >= now() - 1 hour)
    -> router: shard를 하나로 고를 수 없음
    -> shard 0, 1, 2, 3에 각각 query
    -> 각 shard가 local top N 반환
    -> coordinator가 merge/sort/limit
    -> response
```

여기서 coordinator는 단순 중계자가 아닙니다. 전체 정렬, limit, pagination, timeout, partial failure 처리를 맡습니다. 한 shard가 느리면 전체 응답이 늦어질 수 있고, 어떤 shard 응답을 포기하면 결과의 완전성이 달라집니다. 그래서 샤딩 설계에서는 "데이터를 나눌 수 있는가"보다 "실제 query가 한 조각에 머무는가"가 먼저입니다.

이 작은 모델에서 key 질문은 바로 보입니다. Partition key나 shard key가 query predicate에 있으면 범위를 줄일 수 있습니다. 없으면 여러 partition이나 shard를 다 건드립니다. 따라서 "데이터를 나누면 빨라진다"가 아니라 "나눈 기준으로 자주 묻는 질문을 좁힐 수 있을 때 빨라진다"가 정확합니다. 이 판단은 [인덱스와 optimizer](04-index-query-optimizer.md)에서 다룬 access path 판단과 이어집니다. partition이나 shard는 더 큰 단위의 access path를 만드는 선택이기 때문입니다.

Distributed SQL은 이 routing과 replication을 DBMS 내부로 옮기려는 모델입니다.

```text
SQL table: orders

internal key ranges:
  range A: order_id 1-1,000,000
  range B: order_id 1,000,001-2,000,000
  range C: order_id 2,000,001-3,000,000

each range has replicas:
  range A replicas: node1, node2, node3
  consensus leader: node1

write to key in range A:
  SQL layer routes to range A leader
  leader proposes change through consensus
  quorum acknowledges
  transaction commits if all required ranges agree
```

이 모델은 애플리케이션이 shard router를 덜 직접 관리하게 해 줄 수 있습니다. 대신 DB 내부에서 consensus round trip, transaction coordinator, timestamp conflict, range split/merge, leaseholder/leader locality를 관리합니다. 비용이 사라지는 것이 아니라 위치가 바뀝니다.

## 깊은 메커니즘

### Native partition pruning과 lifecycle

Partition pruning은 partition 설계의 첫 번째 관문입니다. PostgreSQL 문서는 declarative partitioning에서 query 조건이 partition constraint와 맞으면 planner가 관련 없는 partition을 제외할 수 있다고 설명합니다. MySQL도 partition pruning을 통해 WHERE 조건에 맞지 않는 partition을 읽지 않을 수 있습니다. 하지만 pruning은 마법이 아닙니다. Query predicate가 partition key와 호환되어야 하고, planner가 조건을 정적으로 또는 실행 시점에 해석할 수 있어야 합니다.

예를 들어 `created_at` range partition인데 query가 `WHERE date(created_at) = '2026-05-20'`처럼 partition key를 함수로 감싸면 planner가 partition bounds와 바로 연결하지 못할 수 있습니다. DBMS별 expression handling과 generated column, functional index, constraint exclusion이 다르므로 실제 `EXPLAIN`으로 확인해야 합니다. 안전한 predicate는 가능하면 half-open range입니다.

```sql
-- pruning에 유리한 형태
WHERE created_at >= timestamp '2026-05-20'
  AND created_at <  timestamp '2026-05-21'

-- planner가 partition bound와 직접 연결하기 어려울 수 있는 형태
WHERE date(created_at) = date '2026-05-20'
```

Partition key 선택은 query와 maintenance를 동시에 봅니다. 날짜 partition은 time-series cleanup과 recent query에 좋습니다. Tenant partition은 특정 tenant의 데이터 격리와 이동에 좋을 수 있습니다. Hash partition은 write 분산에 좋을 수 있지만 범위 삭제나 특정 기간 pruning에는 약합니다. Composite partition은 두 장점을 일부 결합할 수 있지만 partition 수와 index 관리가 복잡해집니다. Key 선택은 "무엇으로 나눌 수 있는가"가 아니라 "무엇으로 자주 좁히는가, 무엇을 독립적으로 보관/삭제/이동하는가"로 판단합니다.

Partitioning이 계속 쓰이는 이유는 단순히 쿼리를 빠르게 만들기 위해서가 아닙니다. 데이터가 계속 쌓이는 업무에서는 "어제까지는 빠르던 테이블"이 어느 순간 백업, 통계 갱신, vacuum, index rebuild, 오래된 데이터 삭제의 단위로 너무 커집니다. 예전 데이터를 `DELETE FROM orders WHERE created_at < ...`로 조금씩 지우면 undo/WAL/binlog, lock, vacuum, replication lag가 함께 커질 수 있습니다. 반대로 월별 partition을 독립 조각으로 두면 오래된 조각을 detach/drop/archive하는 운영 행위가 row 단위 삭제보다 예측 가능한 작업이 됩니다. 그래서 partition key는 조회 조건이면서 보관 정책의 손잡이입니다.

```text
row-by-row cleanup
  3월 row 1억 건 DELETE
    -> log 폭증, vacuum/purge 지연, replica lag 위험

partition lifecycle cleanup
  orders_2026_03 detach
    -> archive 검증
    -> drop 또는 cold storage 이동
```

이 흐름은 partitioning을 "큰 테이블을 쪼개는 기능"보다 넓게 보게 해 줍니다. partition은 optimizer가 읽을 후보를 줄이는 장치이면서, 운영자가 데이터 생명주기를 다루는 더 큰 단위입니다.

Partition은 index 설계도 바꿉니다. PostgreSQL에서는 partitioned table의 index가 각 partition의 index로 구성됩니다. Global unique index 제약에는 제품별 한계가 있습니다. 일반적으로 partition key가 unique constraint에 포함되어야 전체 partition에서 uniqueness를 보장하기 쉽습니다. 예를 들어 `order_id`만 전역 유니크해야 하는데 월별 partition을 쓰면, DBMS가 전역 unique index를 지원하는지 확인해야 합니다. 지원하지 않으면 id generator나 key design이 별도 책임이 됩니다.

Partition lifecycle도 중요합니다. 새 달이 시작되기 전에 새 partition을 만들어야 하고, 오래된 partition을 detach/drop/archive해야 합니다. Partition이 너무 많으면 planner overhead, catalog bloat, migration complexity가 생길 수 있습니다. 월별 partition이 좋은지 일별 partition이 좋은지는 데이터량, query 범위, retention, backup/restore 단위로 판단합니다. 너무 작은 partition은 관리 오버헤드가 커지고, 너무 큰 partition은 pruning과 archiving 이점이 줄어듭니다.

Lifecycle을 손으로 그리면 운영 책임이 더 분명해집니다.

```text
2026-05-31 23:50  새 partition 준비
  orders_2026_06 생성, index와 constraint 확인

2026-06-01 00:00  write 경로 전환
  새 주문은 orders_2026_06으로 들어감

2026-07-01 이후  보관 정책 적용
  orders_2026_05 detach -> archive backup -> 검증 후 drop
```

이 순서에서 `partition 생성`, `write 경로`, `archive 검증`, `drop`은 서로 다른 checkpoint입니다. 새 partition을 만들지 못하면 insert가 실패할 수 있고, archive 검증 없이 drop하면 복구 가능한 보관본이 사라질 수 있습니다. Partitioning은 쿼리 최적화 기능이면서 동시에 시간 단위 운영 절차입니다.

이 축에서 깨지는 신호는 `EXPLAIN`에 예상보다 많은 partition이 나오거나, 새 기간 partition이 없어 insert가 실패하거나, 오래된 partition을 drop했는데 복구 drill에서 대표 row를 찾지 못하는 경우입니다. Pruning 실패는 query predicate와 partition bound의 연결이 끊긴 것이고, lifecycle 실패는 운영 절차의 checkpoint가 닫히지 않은 것입니다.

### 수동 파티셔닝과 drift

수동 파티셔닝은 native partition 기능이 부족하거나 legacy schema를 크게 바꾸기 어려울 때 등장합니다. 예를 들어 `orders_current`, `orders_archive`, `orders_2026_05`처럼 table을 나누고 애플리케이션이 query를 분기하거나 view가 `UNION ALL`로 묶습니다. 이 방식은 특정 table을 독립적으로 백업하거나 drop하기 쉬울 수 있습니다. 하지만 optimizer가 native partition만큼 정보를 얻지 못할 수 있고, 모든 table에 같은 schema, index, constraint, trigger를 유지해야 합니다.

```text
manual partitioning risk

orders_2026_05
  columns: order_id, user_id, created_at, status, amount
  indexes: user_id, created_at

orders_2026_06
  columns: order_id, user_id, created_at, status, amount, coupon_id
  indexes: created_at only

symptom:
  May query is fast, June query is slow.
  Application code has month-specific branches.

root:
  schema/index drift across manually managed tables
```

Sharding은 partition보다 더 큰 경계 이동입니다. Partition은 보통 한 DBMS 안의 storage organization이고, shard는 여러 독립 database node 또는 cluster에 데이터 소유권을 나눕니다. Shard key가 `tenant_id`이면 한 tenant의 데이터가 한 shard에 모여 tenant-local transaction과 query가 쉬워집니다. 하지만 큰 tenant 하나가 전체 부하의 대부분이면 hot shard가 생깁니다. Shard key가 hash(user_id)이면 write 분산은 좋지만 tenant 단위 이동이나 range query는 어려워집니다.

샤딩은 대개 단일 장비를 계속 키우는 방식이 가격, 장애 반경, 쓰기 한계에서 막힐 때 등장합니다. CPU와 메모리를 더 큰 서버로 올리는 vertical scaling은 단순하고 SQL 의미를 많이 보존하지만, 어느 지점부터는 장비 교체와 장애 복구가 부담이 됩니다. 여러 작은 노드에 데이터를 나누는 horizontal scaling은 그 부담을 나누지만, DB가 한 곳에 있을 때 공짜처럼 보이던 join, transaction, unique constraint, backup, schema migration의 책임을 shard 경계마다 다시 묻습니다. 샤딩은 "성능을 더 얻는다"가 아니라 "단일 DB가 맡던 일관성 비용 일부를 routing과 운영 절차로 옮긴다"에 가깝습니다.

| 단일 DB에서 숨겨져 있던 질문 | 샤딩 뒤 드러나는 질문 |
| --- | --- |
| unique key는 DB가 전체 테이블에서 검사한다 | 전역 unique가 필요한가, shard-local unique로 충분한가 |
| join은 optimizer가 table 사이에서 계획한다 | 두 table이 같은 shard key로 모이는가 |
| backup은 한 instance 단위로 생각한다 | shard별 snapshot 시점과 복구 순서를 어떻게 맞추는가 |
| schema migration은 한 catalog에서 수행한다 | 모든 shard에 같은 순서와 버전으로 적용됐는가 |

이 표 때문에 shard key는 해시 함수 선택보다 먼저 업무 경계 선택입니다. 한 주문, 한 사용자, 한 tenant, 한 지갑처럼 같이 읽고 같이 commit되어야 하는 aggregate가 어느 key 주변에 모이는지 먼저 정해야 합니다.

### Sharding: key, routing table, resharding

Routing table은 샤딩의 핵심 운영 자산입니다. 단순 modulo sharding은 `user_id % N`으로 shard를 고르기 쉬우나 shard 수 N이 바뀌면 많은 key가 이동합니다. Consistent hashing이나 virtual bucket을 쓰면 이동 범위를 줄일 수 있습니다. Range sharding은 범위 query에 좋지만 hot range가 생길 수 있습니다. Directory-based sharding은 key-to-shard mapping table을 두어 유연하지만 routing table 일관성과 cache invalidation을 관리해야 합니다.

```text
resharding with virtual buckets

logical bucket count: 1024
physical shards:
  shard A owns buckets 0-255
  shard B owns buckets 256-511
  shard C owns buckets 512-767
  shard D owns buckets 768-1023

add shard E:
  move some buckets from A/B/C/D to E
  router mapping changes bucket ownership
  data copy and dual-read/dual-write window must be controlled
```

Resharding은 단순 data copy가 아닙니다. 이동 중에는 old shard와 new shard 중 어디가 source of truth인지 정해야 합니다. Write를 잠시 막을 것인지, dual-write할 것인지, change log를 따라잡을 것인지, cutover 시점의 idempotency와 rollback은 어떻게 할 것인지 정해야 합니다. Copy가 끝났다는 것과 routing이 바뀌었다는 것, old shard에 더 이상 write가 없다는 것, 검증 checksum이 맞는다는 것은 서로 다른 체크포인트입니다.

작은 bucket 하나를 옮기는 장면도 네 상태로 나눠 보면 실수가 줄어듭니다.

```text
T0 stable
  bucket 17 owner = shard A
  read/write -> shard A

T1 copy
  bucket 17 owner = shard A
  snapshot copy A -> E
  write는 아직 A가 받음

T2 catch-up
  bucket 17 owner = shard A
  copy 이후 변경분을 E에 반영
  checksum(A bucket 17, E bucket 17) 비교

T3 cutover
  bucket 17 owner = shard E
  router cache 갱신
  old shard A에는 새 write가 없어야 함
```

이 trace에서 가장 위험한 구간은 T2와 T3 사이입니다. Data copy는 끝났지만 router 일부가 아직 A를 보고, 일부는 E를 볼 수 있습니다. 그래서 routing table 배포, router cache 만료, idempotent write, rollback window를 함께 설계해야 합니다. "버킷을 옮겼다"는 말은 이 네 상태가 모두 닫혔다는 뜻이어야 합니다.

이 축에서 깨지는 신호는 특정 shard의 QPS와 lock wait만 튀거나, bucket 이동 뒤 row count와 checksum이 맞지 않거나, router cache가 old shard와 new shard를 동시에 바라보는 경우입니다. Sharding 장애는 대개 "데이터가 여러 곳에 있다" 자체보다, 어떤 key가 어느 shard의 소유인지에 대한 정보가 늦게 바뀌거나 서로 다르게 전파될 때 생깁니다.

### Cross-shard query와 transaction

Cross-shard query는 비용이 큽니다. User profile과 user orders가 같은 shard key를 쓰면 local join이 가능할 수 있습니다. 그러나 order by created_at global feed처럼 shard key와 다른 축으로 묻는 query는 모든 shard에 scatter하고 결과를 merge해야 합니다. 이때 pagination, sorting, limit pushdown, duplicate handling이 어려워집니다. 그래서 샤딩 설계에서는 query shape를 먼저 inventory해야 합니다. 나중에 "어떤 shard를 봐야 하는지 모르는 query"가 많으면 scale-out 이점이 사라집니다.

Cross-shard transaction도 어렵습니다. 두 shard에 걸친 원자성을 보장하려면 2PC 같은 coordination이 필요하고, coordinator failure, lock hold time, participant timeout, heuristic outcome 같은 복잡도가 생깁니다. 많은 시스템은 cross-shard transaction을 피하기 위해 aggregate boundary를 shard key와 맞추거나, saga/outbox로 eventual consistency를 선택합니다. 금융 원장처럼 강한 원자성이 필요한 경우에는 shard key 설계가 더 중요해집니다. 이 판단은 [트랜잭션과 ACID 경계](06-transaction-acid-boundary.md)와 [애플리케이션 경계, 멱등성, 돈, outbox](12-application-boundaries-idempotency-money-outbox.md)를 함께 봐야 안전합니다.

이 축에서 깨지는 신호는 single-shard query라고 생각했던 요청이 실제로 모든 shard에 흩어지거나, global sort와 pagination이 coordinator memory를 크게 쓰거나, 두 shard를 건드리는 transaction의 timeout과 retry가 급증하는 경우입니다. `LIMIT 20`이 각 shard의 `LIMIT 20`을 모아 다시 정렬해야 한다면 전체 의미는 이미 한 shard 안에 머무르지 않습니다.

### Distributed SQL: range, consensus, retry, locality

Distributed SQL은 이런 문제 일부를 DBMS 내부로 가져옵니다. 먼저 데이터를 key 범위나 tablet 같은 작은 단위로 나누고, 각 조각을 여러 노드에 복제합니다. 쓰기는 보통 그 조각의 leader가 받아 복제본 다수의 확인(quorum)을 얻은 뒤 commit 가능한 상태로 올립니다. 동시에 transaction이 어떤 시간 순서로 보일지 정하고, 충돌이 나면 transaction을 다시 시도하게 만듭니다. CockroachDB는 SQL layer 아래에 key-value ranges와 Raft replication을 두고, range leaseholder/leader를 통해 읽기와 쓰기를 조율합니다. YugabyteDB도 DocDB, tablet, Raft consensus 같은 구조를 갖습니다. Google Spanner는 TrueTime 기반 timestamp와 Paxos replication으로 외부 일관성(external consistency)을 제공합니다. 제품마다 세부 구현은 다르지만, 공통 질문은 "데이터 조각이 어디에 있고, 누가 leader이며, 몇 복제본이 확인해야 하고, 충돌이 나면 누가 retry하는가"입니다.

Distributed SQL이 등장한 배경에는 두 요구가 같이 있습니다. 하나는 RDBMS의 SQL, schema, transaction 의미를 버리기 어렵다는 요구입니다. 다른 하나는 단일 노드나 단일 region에 모든 write와 장애 복구를 맡기기 어렵다는 요구입니다. 한쪽 극단은 전통적인 단일 노드 RDBMS처럼 강한 관계형 의미를 보존하지만 확장과 region 장애 대응이 어렵고, 다른 극단은 단순 key-value나 document store처럼 분산은 쉬워도 join과 transaction 의미를 애플리케이션이 크게 떠안습니다. Distributed SQL은 이 둘 사이에서 SQL 표면을 유지하려고 하지만, 내부적으로는 range split, replica placement, consensus, timestamp ordering이라는 분산 시스템 비용을 치릅니다.

```text
old comfort
  SQL transaction 하나가 table 여러 개를 바꿈
  DBMS가 같은 서버 안에서 lock/log/recovery를 조율

distributed pressure
  데이터와 사용자가 여러 region에 흩어짐
  한 서버 장애가 서비스 전체 장애가 되면 안 됨

distributed SQL answer
  SQL 표면은 최대한 유지
  내부에서는 range, replica, consensus, retry로 비용을 재배치
```

그래서 distributed SQL을 평가할 때는 "SQL을 지원한다"에서 멈추지 않고, 어떤 transaction이 한 range에 머무는지, 어떤 write가 여러 region quorum을 지나야 하는지, retry가 application contract에 들어와도 안전한지까지 내려가야 합니다.

Consensus는 durability와 failover를 강하게 만들지만 latency를 만듭니다. 단일 노드 DB의 commit은 로컬 WAL flush가 핵심일 수 있습니다. 분산 SQL의 write는 leader가 replica quorum에 log를 복제해야 할 수 있습니다. Cross-region quorum이면 network round trip이 commit latency에 직접 들어옵니다. Multi-region availability를 얻으려면 쓰기 지연과 locality를 설계해야 합니다. 읽기도 follower read, bounded staleness, leaseholder locality 같은 선택에 따라 최신성과 latency가 달라집니다. 복제와 장애 전환의 기본 감각은 [복제, 지연, 백업, failover](09-replication-lag-backup-failover.md)에서 먼저 잡아 두면 distributed SQL의 비용을 더 정확히 읽을 수 있습니다.

Global transaction은 편리하지만 공짜가 아닙니다. Transaction이 한 range 또는 한 shard 안에서 끝나면 비교적 쌉니다. 여러 range와 region을 건드리면 coordinator, timestamp uncertainty, write intent cleanup, conflict detection, retry가 필요해집니다. Distributed SQL 제품은 SQL 표면에서 transaction을 제공하지만, 애플리케이션은 transaction retry를 정상 경로로 받아들여야 합니다. Serializable retry가 자주 나면 query shape, key locality, hot key, transaction size를 의심해야 합니다.

두 row를 갱신하는 transaction도 key 배치에 따라 전혀 다른 작업이 됩니다.

```text
case A: 같은 range 안
  update account:1001
  update ledger:1001:seq10
  -> 한 leader가 대부분 조율
  -> conflict 범위가 작음

case B: 서로 다른 region/range
  update account:1001 in ap-northeast
  update settlement:global-counter in us-central
  -> 여러 leader와 timestamp/commit 조율
  -> network round trip과 retry 가능성이 커짐
```

그래서 distributed SQL에서 transaction이 "지원된다"는 말은 "항상 같은 비용으로 실행된다"는 뜻이 아닙니다. 같은 SQL transaction이라도 key locality가 맞으면 단일 조각에 가까운 비용으로 끝나고, global hot key를 건드리면 합의와 충돌 비용이 커집니다.

Locality 설계는 분산 SQL의 성패를 크게 좌우합니다. 사용자와 가까운 region에서 읽고 쓸 수 있으면 latency가 줄지만, 강한 일관성을 요구하는 global write는 멀리 있는 replica와 합의를 해야 할 수 있습니다. Tenant나 region을 key에 포함해 data placement를 맞추는 전략이 중요합니다. 반대로 모든 region에서 같은 global counter를 갱신하면 hot range와 cross-region consensus가 겹쳐 병목이 됩니다.

이 축에서 깨지는 신호는 `SQL은 성공하지만 p99 commit latency가 특정 key에서만 튀는 경우`, `40001 계열 serialization/retry error가 hot key 주변에서 늘어나는 경우`, `멀리 있는 region에서 strong read/write가 모두 느린 경우`입니다. Distributed SQL은 shard router를 덜 직접 쓰게 해 줄 수 있지만, hot key와 locality, retry contract를 없애지는 않습니다.

### 분산 DB write도 결국 커널과 디스크를 지난다

분산 SQL을 설명할 때 Raft나 timestamp에서 멈추면 실제 commit 비용의 마지막 층이 빠집니다. 각 replica는 결국 하나의 서버 OS 위에서 도는 프로세스입니다. Leader가 quorum을 받기 전에 각 replica가 log entry를 어느 수준까지 durable하게 만들었는지, 그 durable의 의미가 process memory인지, OS page cache인지, storage device의 비휘발 영역인지가 중요합니다.

대표 흐름은 다음처럼 그릴 수 있습니다.

```text
client COMMIT
  -> SQL layer가 row 변경을 KV/range write로 낮춘다
  -> range leader가 Raft log entry를 만든다
  -> follower replica들이 log append를 받는다
  -> 각 DB process가 log file write를 호출한다
  -> kernel page cache에 dirty page가 생긴다
  -> filesystem이 writeback/fsync 시점에 block I/O를 만든다
  -> block layer가 request를 scheduler/driver queue로 넘긴다
  -> device cache와 storage media에 기록된다
  -> 필요한 flush/FUA가 끝난 뒤 durable ack가 올라온다
```

이 그림에서 `write()` 성공과 `fsync()` 성공은 같은 말이 아닙니다. Buffered I/O에서는 데이터가 먼저 커널 page cache에 들어가고, 나중에 writeback으로 storage에 내려갈 수 있습니다. Linux kernel 문서는 storage device가 volatile write-back cache를 가질 수 있고, data integrity operation에서는 filesystem이 forced cache flush나 FUA 같은 block layer 기능으로 비휘발 저장소까지 밀어 넣어야 한다고 설명합니다. PostgreSQL 문서도 committed transaction이 전원 손실 뒤에도 남으려면 OS buffer cache, controller cache, disk cache 경계를 모두 봐야 한다고 설명합니다.

분산 commit의 응답은 여러 층의 ack가 합쳐진 결과입니다.

| Ack 층 | 무엇을 의미하는가 | 면접에서 조심할 말 |
| --- | --- | --- |
| process memory ack | DB process가 log record를 만들었거나 buffer에 올렸다 | durable하다고 말하면 안 됩니다 |
| kernel write ack | kernel이 write 요청을 받아 page cache나 block I/O로 넘겼다 | 전원 손실 뒤 보존을 보장한다고 단정하면 안 됩니다 |
| fsync/flush ack | filesystem과 장치 경계까지 sync 요구가 닫혔다 | 장치와 설정의 신뢰 계약을 확인해야 합니다 |
| quorum ack | 필요한 replica 수가 같은 log entry를 받아 commit 조건을 만족했다 | 각 replica의 local durability 설정과 함께 봐야 합니다 |

이 구분은 성능 튜닝에서도 중요합니다. Sync를 덜 기다리면 latency는 낮아질 수 있지만 장애 뒤 보존성 계약이 약해질 수 있습니다. 반대로 모든 replica와 장치 flush를 강하게 기다리면 보존성은 강해지지만 cross-region write latency가 커질 수 있습니다.

분산 DB의 면접 답변에서는 그래서 두 층을 같이 말하는 편이 안전합니다. 첫째, quorum은 여러 replica가 같은 log entry에 합의했다는 분산 시스템 층의 조건입니다. 둘째, 각 replica의 "log를 썼다"는 말은 로컬 OS와 storage stack에서 어디까지 내려갔는지에 따라 durability 의미가 달라집니다. 네트워크 합의가 끝났어도 각 노드의 fsync/flush 설정과 storage 신뢰성이 약하면 commit latency는 낮아질 수 있지만 장애 뒤 보존성은 약해질 수 있습니다.

## DBMS별 경계

PostgreSQL은 declarative partitioning을 제공합니다. Range, list, hash partitioning을 사용할 수 있고, planner는 partition pruning을 통해 관련 없는 partition을 제외할 수 있습니다. Partition-wise join/aggregate 같은 최적화도 있지만 조건이 있습니다. Unique constraint와 primary key는 partition key 포함 여부 등 제약이 있으므로, 전역 uniqueness 요구를 설계할 때 확인해야 합니다. 또한 너무 많은 partition은 planning과 maintenance 비용을 만듭니다.

MySQL 8.4도 partitioning과 pruning을 제공합니다. RANGE, LIST, HASH, KEY partitioning 계열이 있고, query optimizer가 WHERE 조건을 이용해 필요한 partition만 선택할 수 있습니다. 하지만 InnoDB에서는 foreign key와 user-defined partitioning의 경계가 특히 강한 설계 제약입니다. 공식 문서 기준으로 partitioned InnoDB table은 foreign key를 가질 수 없고, 다른 table의 foreign key가 partitioned InnoDB table의 column을 참조할 수도 없습니다. 따라서 FK-heavy schema에서 MySQL partitioning을 제안한다면 무결성 강제 책임이 DBMS에서 application이나 다른 설계로 밀리는지 먼저 판단해야 합니다. MySQL partitioning은 table 하나의 물리 배치를 나누는 기능이지 여러 server에 자동 분산하는 sharding이 아닙니다.

Vitess는 MySQL sharding middleware 계열로 볼 수 있습니다. VSchema와 vindex를 통해 query routing과 resharding을 돕지만, 모든 MySQL query가 아무 비용 없이 그대로 분산되는 것은 아닙니다. Cross-shard query와 transaction에는 제약과 비용이 있습니다. 이 문서에서는 vendor-specific 세부를 깊게 다루지 않지만, "MySQL을 쓰면서 sharding 운영을 middleware로 보조한다"는 선택지로 분리해 볼 수 있습니다.

CockroachDB는 PostgreSQL wire protocol과 SQL 호환성을 제공하려는 distributed SQL 제품입니다. 내부적으로 range 단위 분산과 Raft 기반 복제를 사용합니다. Strong consistency와 survivability 설정을 제공하지만, transaction retry와 locality 설계를 애플리케이션이 이해해야 합니다. PostgreSQL처럼 접속한다고 해서 단일 노드 PostgreSQL과 같은 latency/profile을 가진다는 뜻은 아닙니다.

YugabyteDB는 PostgreSQL 호환 YSQL layer와 분산 DocDB storage를 결합한 제품입니다. Tablet 단위 분산과 Raft replication을 사용합니다. 분산 transaction과 colocated table, tablespace/placement 같은 locality 도구가 있습니다. Vendor 문서는 제품 기능을 설명하므로, 면접 답변에서는 vendor-specific이라고 분리하고 공통 원리인 range/tablet, consensus, transaction retry, locality를 중심으로 말하는 편이 안전합니다.

Google Spanner는 distributed SQL 논의에서 중요한 기준점입니다. TrueTime API를 사용해 timestamp uncertainty를 다루고, Paxos replication으로 여러 replica의 합의를 구성합니다. Spanner의 강점은 global consistency와 managed operation이지만, 그 대가로 schema design, interleaving/locality, commit latency, vendor lock-in, cost를 봐야 합니다. Spanner 설명을 모든 distributed SQL 제품의 일반 구현처럼 말하면 안 됩니다.

## 직접 재생해 보기

Partition pruning은 `EXPLAIN`으로 확인합니다. PostgreSQL 예시는 다음과 같습니다.

```sql
DROP TABLE IF EXISTS orders_partitioned CASCADE;

CREATE TABLE orders_partitioned (
order_id bigint NOT NULL,
user_id bigint NOT NULL,
created_at timestamp NOT NULL,
status text NOT NULL,
amount numeric(12,2) NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE orders_2026_05 PARTITION OF orders_partitioned
FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE orders_2026_06 PARTITION OF orders_partitioned
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

EXPLAIN
SELECT *
FROM orders_partitioned
WHERE created_at >= timestamp '2026-05-10'
  AND created_at <  timestamp '2026-05-11';
```

PASS 신호는 plan이 5월 partition만 읽는 것입니다. FAIL 신호는 모든 partition을 scan하거나, query predicate가 partition key와 맞지 않아 pruning이 되지 않는 것입니다. 같은 query를 `WHERE date(created_at) = date '2026-05-10'` 형태로 바꾸어 plan이 어떻게 달라지는지도 확인합니다.

MySQL에서는 partition pruning을 `EXPLAIN PARTITIONS`로 볼 수 있습니다.

```sql
CREATE TABLE orders_partitioned (
order_id bigint NOT NULL,
user_id bigint NOT NULL,
created_at date NOT NULL,
status varchar(20) NOT NULL,
amount decimal(12,2) NOT NULL,
PRIMARY KEY (order_id, created_at)
)
PARTITION BY RANGE COLUMNS(created_at) (
PARTITION p202605 VALUES LESS THAN ('2026-06-01'),
PARTITION p202606 VALUES LESS THAN ('2026-07-01'),
PARTITION pmax VALUES LESS THAN (MAXVALUE)
);

EXPLAIN PARTITIONS
SELECT *
FROM orders_partitioned
WHERE created_at >= '2026-05-10'
  AND created_at <  '2026-05-11';
```

PASS는 `partitions` 열에 필요한 partition만 나오는 것입니다. FAIL은 partition key 조건이 없어서 모든 partition을 읽는 것입니다.

Sharding routing은 손으로도 재생할 수 있습니다.

```text
bucket_count = 16
shard_count = 4
bucket = hash(user_id) % 16
shard = routing_table[bucket]

user_id 1001 -> bucket 7 -> shard B
user_id 1002 -> bucket 11 -> shard C

query with user_id:
  route to one shard

query without user_id:
  scatter to all shards
  merge, sort, limit at coordinator
```

PASS는 query가 shard key를 포함할 때 routing이 좁아지고, 포함하지 않을 때 scatter-gather가 필요하다는 것을 설명하는 것입니다. FAIL은 shard 수만 늘리면 모든 query가 자동으로 빨라진다고 말하는 것입니다.

Resharding drill은 실제 데이터 이동 순서로 점검합니다.

```text
1. 현재 bucket-to-shard mapping을 snapshot으로 저장합니다.
2. 이동할 bucket 목록을 고릅니다.
3. old shard에서 new shard로 data copy를 시작합니다.
4. copy 중 생긴 변경을 change log나 dual-write로 따라잡습니다.
5. checksum과 row count를 비교합니다.
6. router mapping을 new shard로 바꿉니다.
7. old shard write가 더 이상 없는지 확인합니다.
8. rollback plan이 유효한 기간을 정합니다.
```

PASS는 cutover 시점에 source of truth가 하나로 정해지고, checksum과 routing이 맞는 것입니다. FAIL은 data copy 완료만 보고 router를 바꾸거나, dual-write 실패를 검증하지 않는 것입니다.

Distributed SQL은 제품 하나를 고정해 로컬 demo cluster에서 재생해야 실제 감각이 생깁니다. 아래 예시는 CockroachDB를 기준으로 한 최소 절차입니다. CockroachDB 공식 문서의 `cockroach demo --nodes=3`는 임시 multi-node demo cluster를 만들고, transaction retry 문서는 `SERIALIZABLE` 환경에서 동시 write 충돌이 client-visible retry error로 올라올 수 있다고 설명합니다.

```text
terminal 1
  cockroach demo --nodes=3 --no-example-database
  \demo ls
  -- 표시된 sql URL을 terminal 2, 3에서 각각 사용합니다.

session A/B 공통 준비
  CREATE DATABASE IF NOT EXISTS lab;
  USE lab;
  CREATE TABLE IF NOT EXISTS account_balance (
    account_id INT PRIMARY KEY,
    balance INT NOT NULL
  );
  UPSERT INTO account_balance VALUES (1, 1000), (2, 1000);
```

Hot key 충돌은 두 session에서 같은 row를 동시에 갱신해 봅니다.

```sql
-- session A
BEGIN;
UPDATE account_balance
SET balance = balance + 1
WHERE account_id = 1;
-- 아직 COMMIT하지 않습니다.

-- session B
BEGIN;
UPDATE account_balance
SET balance = balance - 1
WHERE account_id = 1;
COMMIT;

-- session A
COMMIT;
```

PASS 신호는 둘 중 한 transaction이 대기하거나, conflict가 커질 때 retry 가능한 오류가 관측되고, 애플리케이션이 transaction 전체를 다시 실행해야 한다는 결론까지 이어지는 것입니다. FAIL 신호는 오류 코드를 한 번 보고 "DB 장애"라고만 판단하거나, `UPDATE` 한 문장만 다시 보내도 된다고 생각하는 것입니다. Transaction 안에 `잔액 조회 -> 한도 판단 -> 변경 -> outbox 기록` 같은 여러 단계가 있으면, retry 단위는 SQL 한 줄이 아니라 전체 판단 흐름이어야 합니다.

Locality와 stale read는 multi-region demo에서 별도로 봅니다.

```text
1. cockroach demo --global --nodes=9 로 region locality가 있는 임시 cluster를 엽니다.
2. SHOW REGIONS FROM CLUSTER; 로 cluster region을 확인합니다.
3. read-mostly table과 자주 갱신되는 table을 나누어 table locality를 다르게 설정합니다.
4. 가까운 region에서 읽을 때와 먼 region의 leader/leaseholder를 거칠 때 latency를 비교합니다.
5. stale follower read를 허용한 read와 최신값이 필요한 strong read를 같은 의미로 말하지 않습니다.
```

PASS 신호는 `이 화면은 약간 오래된 값을 읽어도 된다`, `이 결제/잔액 경로는 최신 write를 읽어야 한다`처럼 업무 의미별로 read policy가 나뉘는 것입니다. CockroachDB의 follower read 문서는 가까운 replica에서 과거의 일관된 시점을 읽어 latency를 줄일 수 있지만 최신성 trade-off를 가진다고 설명합니다. 따라서 distributed SQL 검증은 `SQL이 실행된다`가 아니라 `retry, locality, stale read, global write latency가 업무 계약 안에 들어오는가`로 닫아야 합니다.

## 면접 꼬리 질문

1. Partitioning과 sharding의 차이는 무엇인가요?

Partitioning은 보통 하나의 논리 table을 DBMS 내부의 여러 물리 partition으로 나누는 것입니다. Sharding은 데이터를 여러 DB node나 cluster에 나누어 저장합니다. Partitioning은 pruning과 maintenance에 강하고, sharding은 단일 노드 한계를 넘는 scale-out에 강하지만 routing과 cross-shard 비용을 만듭니다.

1. Partition pruning은 언제 실패하나요?

Query predicate가 partition key와 맞지 않거나, key를 함수로 감싸 planner가 bounds를 해석하지 못하거나, runtime parameter 처리 한계가 있거나, 통계와 constraint가 부족하면 실패할 수 있습니다. 실제로는 `EXPLAIN`으로 어떤 partition을 읽는지 확인해야 합니다.

1. Shard key는 어떻게 고르나요?

자주 쓰는 query가 한 shard로 좁혀지는지, transaction boundary가 key 안에 들어오는지, hot key가 생기지 않는지, resharding이 가능한지, tenant 이동이나 지역 배치 요구가 있는지 봅니다. 단순 hash는 분산에는 좋지만 range query와 tenant 이동에는 약할 수 있습니다.

1. Cross-shard transaction은 왜 어려운가요?

여러 shard가 하나의 commit 결과에 합의해야 하므로 coordinator, participant, prepare/commit, timeout, failure recovery가 필요합니다. Lock 보유 시간과 failure mode가 커집니다. 가능하면 aggregate boundary를 shard key와 맞추거나, saga/outbox로 eventual consistency를 선택합니다.

1. Distributed SQL은 sharding 문제를 완전히 없애나요?

애플리케이션이 직접 shard router를 덜 관리하게 해 줄 수는 있지만, 비용이 사라지지는 않습니다. 내부 range/tablet, consensus, transaction retry, locality, hot key 문제가 남습니다. SQL 표면이 익숙해도 physical distribution을 이해해야 합니다.

1. Global transaction을 쓰면 항상 더 안전한가요?

강한 일관성을 줄 수 있지만 latency와 availability 비용을 만듭니다. 모든 업무가 global serializable transaction을 필요로 하지는 않습니다. 지역별 독립성이 있는 데이터는 locality를 살리고, 꼭 필요한 invariant에만 global coordination을 쓰는 편이 낫습니다.

## 함정 질문

1. Partition을 많이 만들수록 항상 빨라지나요?

아닙니다. Pruning이 잘 되면 읽을 데이터가 줄지만, partition 수가 너무 많으면 planning, catalog, maintenance, index 관리 비용이 커집니다. Query가 partition key를 쓰지 않으면 많은 partition을 다 읽을 수 있습니다.

1. 월별 table로 나누면 native partition과 같은가요?

아닙니다. 수동 table 분리는 routing, schema/index drift, union view, constraint, migration을 직접 관리해야 합니다. Native partition은 DBMS planner와 DDL이 partition 구조를 이해합니다. 수동 방식은 운영 책임이 더 애플리케이션 쪽에 있습니다.

1. Hash sharding을 쓰면 hot spot이 사라지나요?

Key 분포가 균등하면 도움이 되지만 hot user, hot tenant, global counter, 특정 이벤트성 key가 있으면 여전히 hot shard나 hot range가 생깁니다. Write pattern과 business skew를 함께 봐야 합니다.

1. NewSQL은 CAP를 우회하나요?

아닙니다. 제품은 consistency, availability, partition tolerance 사이의 trade-off를 특정 방식으로 선택합니다. Consensus와 replication으로 강한 consistency를 제공할 수 있지만 network partition에서는 availability나 latency 비용이 생깁니다. CAP를 없애는 것이 아니라 선택을 제품과 설정 안에 담는 것입니다.

1. PostgreSQL 호환 distributed SQL이면 기존 PostgreSQL 앱을 그대로 옮기면 되나요?

SQL 문법과 wire protocol 호환은 출발점일 뿐입니다. 기존 PostgreSQL 앱을 옮길 때는 "접속된다"가 아니라 "기존 업무 계약이 분산 실행에서도 유지된다"를 증명해야 합니다.

| 검증 축 | 이전 전에 만들어야 하는 작은 시험 | 깨지는 신호 |
| --- | --- | --- |
| Transaction retry | 결제 승인, 재고 차감, 포인트 적립처럼 여러 SQL과 판단이 묶인 transaction을 일부러 충돌시킨 뒤 전체 transaction retry가 가능한지 봅니다 | `40001`류 retry error 뒤 일부 side effect만 남거나, retry가 SQL 한 줄 단위로만 구현됩니다 |
| Sequence/identity | 주문 번호, 전표 번호처럼 단조 증가를 기대하는 값을 insert concurrency와 rollback 상황에서 확인합니다 | 값이 건너뛰는 것을 장애로 오해하거나, commit 순서와 번호 순서를 같은 뜻으로 사용합니다 |
| Lock/savepoint 의미 | `SELECT ... FOR UPDATE`, savepoint rollback, advisory lock에 기대는 코드를 분리해 제품 문서와 실험으로 확인합니다 | PostgreSQL에서 되던 row lock release나 wait 패턴을 그대로 가정합니다 |
| Query plan과 latency | 가장 중요한 10개 query를 실제 데이터 분포와 비슷한 샘플에서 `EXPLAIN`과 p95/p99 latency로 비교합니다 | 단일 노드에서는 local index scan이던 query가 distributed lookup, remote scan, fan-out으로 바뀝니다 |
| Extension과 DDL | 사용하는 extension, trigger, partial index, generated column, online DDL을 목록화하고 지원 여부와 migration 절차를 확인합니다 | 배포 당일 unsupported feature나 긴 schema change job이 발견됩니다 |
| 운영 도구 | backup/restore, PITR, metric, alert, slow query 수집, 권한 모델, 장애 전환 runbook을 새 제품 기준으로 재작성합니다 | 장애 때 기존 PostgreSQL runbook을 실행했지만 관측 지표와 복구 명령이 맞지 않습니다 |

좋은 답변은 "PostgreSQL compatible이니까 대부분 됩니다"가 아니라, compatibility를 문법, driver, transaction, lock, plan, 운영 절차로 나누어 증명한다고 말하는 것입니다. 특히 transaction retry와 외부 side effect의 경계를 [애플리케이션 경계, 멱등성, 돈, outbox](12-application-boundaries-idempotency-money-outbox.md)에서 다시 확인해야 합니다.

### 마지막 연결: 데이터를 나눴을 때 query와 transaction이 어디까지 한 조각에 머무는가

이 주제의 답변은 `어떤 key로 나누는가 -> 그 key로 query가 좁혀지는가 -> 좁혀지지 않을 때 fan-out과 transaction 비용을 누가 감당하는가`로 이어져야 합니다. `partition key, shard key, routing, pruning, rebalance, consensus`는 암기용 키워드가 아니라 같은 row가 어느 물리 조각으로 이동하고, 어느 coordinator가 그 조각을 찾아가며, 어떤 합의나 검증을 거쳐 결과를 닫는지 보여 주는 경로입니다. 면접에서는 이 경로를 말해야 "큰 table을 나눴다"는 말이 실제 설계 판단으로 바뀝니다.

반례를 하나 놓고 보면 경계가 선명해집니다. `created_at` partition은 월별 조회에는 좋지만 user_id 전체 이력 조회에는 모든 partition을 읽을 수 있습니다. 반대로 `user_id` hash shard는 사용자별 point lookup에는 좋지만, "최근 1시간 전체 주문 TOP 100"에는 모든 shard의 local TOP N을 모아 다시 정렬해야 합니다. 그래서 답변은 "나누었다"에서 끝나면 안 되고, 실제 query가 한 조각을 고를 수 있는지까지 내려가야 합니다.

DBMS별로는 PostgreSQL/MySQL partitioning은 주로 한 시스템 안의 pruning과 관리 단위이고, sharding과 distributed SQL은 node 간 routing, replication, consensus, cross-shard transaction 문제를 만듭니다. 이 차이는 장애 대응도 바꿉니다. PostgreSQL partition에서 오래된 데이터를 지우는 문제라면 `DETACH/DROP partition, archive 검증, autovacuum 영향`을 봅니다. Sharding에서 같은 증상이 나오면 `어떤 shard가 뜨거운지, routing table이 맞는지, rebalance 중 dual-write가 성공했는지`를 봅니다. Distributed SQL에서는 `hot range, leaseholder locality, transaction retry, quorum latency`를 함께 봅니다.

현장에서 이 설명을 검증하려면 `partition pruning 여부, shard별 QPS와 p99 latency, hot key 충돌률, cross-shard query 비율, rebalance 소요 시간, distributed transaction retry율`을 봅니다. 예를 들어 "sharding했는데 느립니다"라는 말은 `특정 shard p99만 높은가`, `모든 shard가 같이 높은가`, `coordinator merge 시간이 높은가`에 따라 원인이 달라집니다. 지표가 예상과 다르면 처음 세운 key 선택과 query boundary를 다시 열어야 합니다.

```text
incoming query -> choose partition/shard from key -> local scan or fan-out -> aggregate/commit -> rebalance or archive operation
```

이 trace를 손으로 다시 그리면, partition과 sharding을 단순 분할 기법으로 외웠는지, 아니면 query와 transaction이 실제로 어느 조각까지 머무는지 이해했는지가 드러납니다. 각 화살표마다 `이 key로 한 조각을 고를 수 있는가`, `fan-out이 필요하면 누가 merge하는가`, `commit이 여러 조각을 건드리면 누가 재시도와 검증을 맡는가`를 붙여 봅니다. 답이 막히는 화살표가 있으면 그 부분이 설계와 운영에서 가장 먼저 깨질 지점입니다.

면접에서는 최종적으로 이렇게 압축할 수 있습니다. Partition은 한 DBMS 안에서 읽을 물리 조각과 보관 단위를 줄이는 방법이고, sharding은 데이터 소유권을 여러 노드에 나누어 query와 transaction 경계를 key 주변으로 모으는 방법입니다. Distributed SQL은 이 분산과 합의 비용을 DBMS 내부로 옮겨 SQL 표면을 유지하려는 시도입니다. 좋은 답변은 이 세 단어를 분리한 뒤, `이 query는 한 조각에 머무는가`, `이 transaction은 몇 조각을 commit해야 하는가`, `깨졌을 때 어떤 지표로 확인하는가`까지 내려갑니다.

## 더 깊게 볼 자료

- [PostgreSQL 18 문서: Table Partitioning](https://www.postgresql.org/docs/current/ddl-partitioning.html) - declarative partitioning, pruning, maintenance 경계를 확인할 수 있습니다.
- [PostgreSQL 18 문서: Partition Pruning](https://www.postgresql.org/docs/current/ddl-partitioning.html#DDL-PARTITION-PRUNING) - planner가 partition을 제외하는 조건을 볼 수 있습니다.
- [MySQL 8.4 문서: Partitioning](https://dev.mysql.com/doc/refman/8.4/en/partitioning.html) - MySQL partitioning 기능과 제약의 공식 입구입니다.
- [MySQL 8.4 문서: Partition Pruning](https://dev.mysql.com/doc/refman/8.4/en/partitioning-pruning.html) - `EXPLAIN PARTITIONS`로 pruning을 확인하는 방법을 볼 수 있습니다.
- [MySQL 8.4 문서: Partitioning Limitations Relating to Storage Engines](https://dev.mysql.com/doc/refman/8.4/en/partitioning-limitations-storage-engines.html) - InnoDB partitioning과 foreign key의 hard caveat를 확인할 수 있습니다.
- [CockroachDB 문서: Architecture Overview](https://www.cockroachlabs.com/docs/stable/architecture/overview) - range, replica, leaseholder, Raft 등 CockroachDB-specific 구조를 확인할 수 있습니다.
- [CockroachDB 문서: cockroach demo](https://www.cockroachlabs.com/docs/stable/cockroach-demo) - multi-node와 multi-region demo cluster를 로컬에서 시작하는 방법을 확인할 수 있습니다.
- [CockroachDB 문서: Transaction Retry Error Reference](https://www.cockroachlabs.com/docs/stable/transaction-retry-error-reference) - `40001` retry error와 client-side retry 처리의 공식 기준입니다.
- [CockroachDB 문서: Follower Reads](https://www.cockroachlabs.com/docs/stable/follower-reads) - 가까운 replica에서 과거의 일관된 시점을 읽는 선택과 최신성 trade-off를 확인할 수 있습니다.
- [YugabyteDB 문서: Architecture](https://docs.yugabyte.com/preview/architecture/) - tablet, DocDB, Raft replication 등 YugabyteDB-specific 구조를 확인할 수 있습니다.
- [Google Spanner paper](https://research.google/pubs/spanner-googles-globally-distributed-database/) - TrueTime과 external consistency를 다룬 원 논문입니다.
- [Linux kernel 문서: Explicit volatile write back cache control](https://www.kernel.org/doc/html/v6.6/block/writeback_cache_control.html) - fsync/sync/unmount 같은 data integrity operation에서 block layer flush와 FUA가 왜 필요한지 확인합니다.
- [PostgreSQL 문서: Reliability](https://www.postgresql.org/docs/16/wal-reliability.html) - OS buffer cache, controller cache, disk cache가 DB durability 설명에 왜 들어가는지 확인합니다.
- `database/deep-dive/reliability-distribution/16-partitioning-sharding-distribution.md`와 `database/deep-dive/search-nosql-newsql/28-newsql-distributed-sql.md` - 기존 장문 정리 자료입니다. 파티션, 샤딩, NewSQL을 더 넓은 맥락에서 이어 읽을 때 참고합니다.
