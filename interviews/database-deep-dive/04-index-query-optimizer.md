# 인덱스와 옵티마이저는 "빨리 찾기"가 아니라 row stream을 싸게 만드는 협상이다

인덱스를 설명할 때 "B-tree 계열이라서 검색이 O(log N)입니다"라고 말하면 시작은 했지만 아직 면접 답변이 아닙니다. 실제 query는 하나의 key를 찾는 문제만이 아닙니다. scan, filter, sort, aggregate, join, pagination이 함께 움직이고, 옵티마이저는 통계를 바탕으로 여러 실행 경로의 비용을 추정합니다. 인덱스는 그 경로 후보를 늘려 주지만, 항상 선택되는 것도 아니고 항상 빠른 것도 아닙니다.

이 문서는 B-tree/B+tree 계열 구조를 중심으로 composite, covering, partial, function index를 설명하고, hash, bitmap, LSM 계열과의 차이를 경계로 둡니다. 이어서 scan/filter/sort/aggregate/join이 row stream을 어떻게 바꾸는지, statistics와 `EXPLAIN`이 왜 plan regression 진단의 언어가 되는지, pagination에서 offset과 keyset이 왜 다른 비용 구조를 가지는지 설명합니다. Index lookup이 page와 buffer를 어떻게 움직이는지는 [storage page와 buffer I/O](02-storage-pages-buffer-io.md), index-only scan이 visibility map에 기대는 이유는 [MVCC와 snapshot visibility](07-mvcc-snapshot-visibility.md), index를 추가하거나 바꾸는 운영 위험은 [schema와 migration](05-schema-constraints-migration.md)으로 이어서 보면 됩니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [B-tree/B+tree 계열은 정렬된 leaf를 이용해 범위를 줄인다](#b-treebtree-계열은-정렬된-leaf를-이용해-범위를-줄인다)
    - [composite index는 왼쪽부터 정렬된 tuple이다](#composite-index는-왼쪽부터-정렬된-tuple이다)
    - [covering index와 index-only scan은 table 방문을 줄이는 전략이다](#covering-index와-index-only-scan은-table-방문을-줄이는-전략이다)
    - [partial index와 function index는 index의 적용 범위를 바꾼다](#partial-index와-function-index는-index의-적용-범위를-바꾼다)
    - [hash, bitmap, LSM은 같은 index라는 이름 아래 다른 질문에 답한다](#hash-bitmap-lsm은-같은-index라는-이름-아래-다른-질문에-답한다)
    - [scan, filter, sort, aggregate는 row stream을 단계적으로 바꾼다](#scan-filter-sort-aggregate는-row-stream을-단계적으로-바꾼다)
    - [statistics는 optimizer의 눈이다](#statistics는-optimizer의-눈이다)
    - [optimizer cost는 운영체제 I/O 비용을 간접적으로 본다](#optimizer-cost는-운영체제-io-비용을-간접적으로-본다)
    - [EXPLAIN은 정답표가 아니라 가설 검사 도구다](#explain은-정답표가-아니라-가설-검사-도구다)
    - [pagination은 index order와 cursor 안정성 문제다](#pagination은-index-order와-cursor-안정성-문제다)
    - [plan regression은 SQL 문제가 아니라 환경과 데이터 문제일 수 있다](#plan-regression은-sql-문제가-아니라-환경과-데이터-문제일-수-있다)
- [DBMS별 경계](#dbms별-경계)
- [직접 재생해 보기](#직접-재생해-보기)
    - [Composite index와 leftmost prefix 확인](#composite-index와-leftmost-prefix-확인)
    - [PostgreSQL에서 index-only scan 경계 확인](#postgresql에서-index-only-scan-경계-확인)
    - [Pagination offset과 keyset 비교](#pagination-offset과-keyset-비교)
- [면접 꼬리 질문](#면접-꼬리-질문)
- [함정 질문](#함정-질문)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

인덱스는 table 전체를 읽지 않고도 후보 row 위치를 좁히기 위한 물리 구조입니다. PostgreSQL 문서는 index가 row를 더 빠르게 찾게 해 주지만, 전체 DBMS에 overhead를 추가하므로 신중히 써야 한다고 설명합니다. MySQL InnoDB 문서는 spatial index를 제외한 InnoDB index가 B-tree structure이고, clustered index와 secondary index가 서로 다른 역할을 가진다고 설명합니다.

B+tree 계열 인덱스는 key 순서로 정렬된 leaf를 따라 equality, range, order 요구를 처리하기 좋습니다. Composite index는 여러 column 값을 이어 붙인 정렬 구조로 이해하면 됩니다. MySQL 문서도 multiple-column index를 indexed column 값들을 concatenate한 sorted array처럼 볼 수 있다고 설명하고, leftmost prefix가 중요하다고 설명합니다. PostgreSQL multicolumn B-tree도 leading column equality와 첫 inequality 조건이 scan 범위를 줄이는 데 핵심입니다.

Covering index는 query가 필요한 column을 index에서 얻어 table 또는 heap 방문을 줄이는 설계입니다. PostgreSQL의 index-only scan은 index type과 query column 조건만으로 끝나지 않고, MVCC visibility를 확인해야 하므로 visibility map이 중요합니다. InnoDB에서는 secondary index가 primary key를 포함하고, query가 secondary index에 있는 column만 요구하면 clustered lookup을 피할 수 있습니다. 같은 "covering"이라도 DBMS별 세부 경계가 다릅니다.

Partial index와 function index는 "모든 row의 모든 원본 값을 같은 방식으로 정렬한다"는 기본 B-tree 감각을 좁히거나 바꾸는 방법입니다. PostgreSQL partial index는 predicate를 만족하는 subset만 담고, planner가 query condition이 index predicate를 함의한다고 알아야 사용할 수 있습니다. Expression index는 `lower(email)`처럼 계산 결과를 index key로 저장해 검색을 빠르게 합니다. MySQL은 generated column과 functional key parts, prefix index 등으로 비슷한 문제를 풀 수 있지만 문법과 제약이 다릅니다.

옵티마이저는 index 존재 여부만 보지 않습니다. table과 index page 수, row 수, distinct count, most common values, histogram 같은 통계를 바탕으로 selectivity와 cost를 추정합니다. PostgreSQL `EXPLAIN`은 plan node tree와 cost, rows estimate를 보여 주고, MySQL `EXPLAIN`은 table read order, chosen key, estimated rows, filtered ratio 등을 보여 줍니다. plan regression은 SQL이 바뀌지 않았는데 데이터 분포, 통계, parameter, 설정, DBMS version 변화 때문에 더 나쁜 plan이 선택되는 현상으로 볼 수 있습니다.

Pagination은 query cost를 드러내는 좋은 예입니다. `OFFSET 100000 LIMIT 20`은 앞의 100000개를 "반환하지 않을 뿐" 지나가야 할 수 있습니다. Keyset pagination은 마지막으로 본 key를 조건으로 넣어 다음 범위를 바로 좁힙니다. 이때 `(created_at, id)` 같은 composite index와 정렬 방향이 맞아야 안정적인 성능과 중복/누락 없는 페이지 이동을 만들 수 있습니다. 정렬 key가 nullable이면 `NULLS FIRST/LAST`나 DBMS별 NULL 정렬 규칙까지 cursor 조건에 포함해야 합니다.

## 먼저 잡아야 할 작은 모델

가장 작은 모델은 주문 목록 페이지입니다.

```sql
SELECT id, created_at, total_amount
FROM orders
WHERE user_id = ?
  AND status = 'PAID'
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

후보 index를 세 개 생각해 보자.

```text
index A: (user_id)
index B: (user_id, status)
index C: (user_id, status, created_at DESC, id DESC) INCLUDE(total_amount)
```

같은 SQL이라도 각 index가 만드는 row stream이 다릅니다.

```text
index A path
  user_id = 42인 모든 주문 후보를 찾습니다.
  status = 'PAID'는 table/index filter로 더 걸러야 합니다.
  created_at DESC 정렬이 필요할 수 있습니다.
  total_amount를 얻기 위해 table/heap 방문이 필요할 수 있습니다.

index B path
  user_id = 42이고 status = 'PAID'인 후보를 더 좁힙니다.
  created_at DESC 순서는 아직 보장되지 않을 수 있습니다.
  sort 또는 다른 order path가 필요합니다.

index C path
  user_id/status equality로 범위를 좁힙니다.
  그 범위 안에서 created_at/id 순서로 leaf를 읽습니다.
  필요한 column이 index에 있으면 table 방문을 줄일 수 있습니다.
```

여기서 움직이는 대상은 "row stream"입니다.

```text
raw table rows
  [user=1 paid], [user=42 canceled], [user=42 paid old], [user=42 paid new], ...

after access path
  index C jumps to user=42/status=PAID range

after order
  [newest paid], [next paid], [next paid], ...

after limit
  first 20 rows
```

이 작은 모델은 인덱스 설계의 핵심을 보여 줍니다. 좋은 index는 조건을 줄이고, 정렬을 피하고, table lookup을 줄이는 방향으로 row stream을 만들어 줍니다. 하지만 index가 넓어질수록 write overhead와 storage cost가 늘어납니다. `INCLUDE(total_amount)` 같은 payload column은 index-only scan에는 도움이 되지만 index size를 키웁니다. 그래서 index 설계는 "query 하나를 빠르게"가 아니라 "read path 절감과 write/storage overhead의 교환"입니다.

## 깊은 메커니즘

### B-tree/B+tree 계열은 정렬된 leaf를 이용해 범위를 줄인다

B-tree/B+tree 계열 index의 핵심은 key를 정렬된 tree에 넣고, root에서 branch를 따라 leaf까지 내려가 후보 범위를 찾는 것입니다. InnoDB 문서는 spatial index를 제외한 index가 B-tree structure이고, index record가 leaf page에 저장된다고 설명합니다. PostgreSQL도 기본 index type으로 B-tree를 제공하고 equality와 range, ordering에 널리 씁니다.

```text
root
  key < M  -> left branch
  key >= M -> right branch

leaf pages
  ... 38, 39, 40, 41, 42 | 43, 44, 45 ...
                  ^
              search key
```

single key lookup에서는 root에서 leaf까지 내려가고, leaf에서 matching entry를 찾습니다. range query에서는 시작 key를 찾은 뒤 leaf를 순서대로 읽습니다. 이 구조는 equality와 range에 강하지만, key의 정렬 규칙에 맞지 않는 조건에는 약합니다. `WHERE lower(email) = 'kim@example.com'`은 `email` 원본 key와 다른 표현식이므로 일반 `email` index를 그대로 쓰기 어렵습니다. expression/function index가 필요한 이유가 여기에 있습니다.

삽입과 update도 공짜가 아닙니다. random key를 계속 넣으면 leaf page split이 생기고, tree가 커지고, buffer와 WAL/redo pressure가 생깁니다. InnoDB physical index 문서는 sequential insert면 page가 대략 15/16까지 채워질 수 있지만 random order insert에서는 page가 1/2부터 15/16 사이로 남을 수 있다고 설명합니다. 즉 key 분포와 insertion order는 storage density와 write cost에 영향을 줍니다.

### composite index는 왼쪽부터 정렬된 tuple이다

Composite index는 `(a, b, c)`를 따로따로 세 개의 index로 갖는 것이 아니라, `(a 값, b 값, c 값)` tuple을 하나의 정렬 key로 삼는 구조입니다. MySQL 문서는 multiple-column index를 indexed column 값을 concatenate해 만든 sorted array로 볼 수 있다고 설명합니다. 그래서 leftmost prefix가 중요합니다.

```text
index (user_id, status, created_at)

sorted keys:
  (1, PAID, 2026-01-01)
  (1, PAID, 2026-01-02)
  (1, REFUND, 2026-01-01)
  (2, PAID, 2026-01-01)
  (42, CANCELED, 2026-01-03)
  (42, PAID, 2026-01-01)
  (42, PAID, 2026-01-05)
```

`WHERE user_id = 42 AND status = 'PAID'`는 하나의 연속 범위를 만듭니다. 그 범위 안에서 `created_at` 순서도 이미 정렬되어 있습니다. 반대로 `WHERE status = 'PAID'`만 있으면 `user_id`별로 흩어진 여러 범위를 모두 봐야 합니다. MySQL의 기본 설명은 leftmost prefix를 기준으로 잡아야 합니다. 다만 MySQL도 조건이 맞으면 skip scan range access를 선택해 leading column의 distinct value를 건너뛰며 여러 subrange를 읽을 수 있고, 이때 `EXPLAIN`의 `Extra`에 `Using index for skip scan`이 보일 수 있습니다. PostgreSQL도 version과 조건에 따라 skip scan 같은 optimization을 고려할 수 있습니다. 그렇다고 "앞 column 조건이 없어도 항상 같은 비용으로 쓴다"는 뜻은 아닙니다. Skip scan은 optimizer가 조건, 통계, 비용을 보고 고르는 예외적 접근이고, 설계의 기본 감각은 여전히 leading column 조건이 scan 범위를 가장 직접적으로 줄인다는 쪽에 둬야 합니다.

Composite index 설계는 equality, range, order의 순서를 같이 봐야 합니다.

```text
좋은 후보
  WHERE user_id = ?
    AND status = ?
  ORDER BY created_at DESC, id DESC

index
  (user_id, status, created_at DESC, id DESC)

이유
  equality columns narrow the range
  order columns keep rows in desired order inside the range
```

하지만 모든 query를 하나의 composite index로 해결하려 하면 write overhead와 index sprawl이 생깁니다. 같은 table에 `(user_id, status, created_at)`, `(status, user_id, created_at)`, `(created_at, user_id)`를 계속 추가하면 insert/update/delete마다 여러 index를 유지해야 합니다. 실무에서는 query 빈도, selectivity, order 요구, write rate, index size를 함께 봅니다.

### covering index와 index-only scan은 table 방문을 줄이는 전략이다

Index scan에서 비싼 부분은 key 찾기만이 아닙니다. key를 찾은 뒤 full row를 얻기 위해 table 또는 heap을 방문해야 할 수 있습니다. PostgreSQL 문서는 일반 index scan에서 index와 heap을 모두 fetch해야 하며, heap access가 random access가 될 수 있다고 설명합니다. Index-only scan은 query에 필요한 값을 index entry에서 직접 반환해 heap access를 줄이는 방법입니다.

PostgreSQL에서는 index-only scan이 가능하려면 index가 필요한 값을 저장하거나 재구성할 수 있어야 하고, query가 index에 저장된 column만 참조해야 합니다. 그런데 MVCC visibility 정보는 index entry에 없기 때문에 visibility map도 봐야 합니다. heap page가 all-visible이면 heap visit 없이 row를 반환할 수 있지만, 최근 변경된 table에서는 heap visit이 필요해질 수 있습니다. 그래서 PostgreSQL에서 covering index를 만들었다고 항상 heap access가 사라지는 것은 아닙니다.

InnoDB에서는 secondary index entry가 primary key를 포함하고, query가 secondary index leaf에 있는 column만 필요로 하면 clustered index lookup을 피할 수 있습니다. 하지만 full row column이 필요하면 primary key로 clustered index를 다시 찾아야 합니다. 같은 "covering"이라는 말을 쓰더라도 PostgreSQL은 visibility map 경계가 중요하고, InnoDB는 secondary index에 어떤 column이 들어 있는지와 clustered lookup 경계가 중요합니다.

```text
query
  SELECT created_at, total_amount
  FROM orders
  WHERE user_id = 42 AND status = 'PAID'
  ORDER BY created_at DESC
  LIMIT 20

covering candidate
  index(user_id, status, created_at DESC) INCLUDE(total_amount)

without covering
  index leaf -> table row lookup -> return columns

with covering
  index leaf -> return columns from index
  PostgreSQL: visibility map condition matters
  InnoDB: secondary index contains needed columns and primary key; full row lookup may be skipped
```

### partial index와 function index는 index의 적용 범위를 바꾼다

Partial index는 table 전체가 아니라 predicate를 만족하는 subset만 index에 넣습니다. PostgreSQL 문서는 partial index가 conditional expression으로 정의된 subset 위에 만들어지고, common value를 index에서 제외해 index size와 update 비용을 줄일 수 있다고 설명합니다. 예를 들어 `WHERE deleted_at IS NULL`인 active row만 자주 찾는다면 active row subset index가 유리할 수 있습니다.

```sql
CREATE INDEX orders_active_user_created_idx
ON orders (user_id, created_at DESC)
WHERE deleted_at IS NULL;
```

하지만 partial index는 query predicate가 index predicate를 함의한다고 planner가 알아야 씁니다. PostgreSQL 문서는 일반 theorem prover가 아니므로, query의 WHERE 조건이 partial index predicate와 맞아야 한다고 설명합니다. parameterized query clause가 partial index와 잘 맞지 않는 경우도 있습니다. 즉 partial index는 storage를 줄이는 강력한 도구지만, query 형태와 predicate 안정성을 함께 관리해야 합니다.

Function 또는 expression index는 원본 column이 아니라 계산 결과를 key로 저장합니다. PostgreSQL expression index 문서는 `lower(col1)` 같은 계산 결과에 index를 만들면 `WHERE lower(col1) = 'value'`가 index를 쓸 수 있다고 설명합니다. 이 방식은 case-insensitive search, normalized phone number, date truncation 같은 문제에 유용합니다. 대가도 있습니다. insert와 non-HOT update 때 expression을 계산하고 index를 갱신해야 합니다.

```sql
CREATE INDEX users_email_lower_idx
ON users (lower(email));

SELECT id FROM users
WHERE lower(email) = lower('Kim@Example.com');
```

여기서 함정은 application에서 `lower(email)`을 쓰는 query와 index expression이 정확히 맞아야 한다는 점입니다. collation, function volatility, generated column, DBMS별 expression syntax 차이도 봅니다. MySQL에서는 functional key parts나 generated column index로 비슷한 문제를 다룰 수 있지만, PostgreSQL partial/expression index 문법을 그대로 옮길 수는 없습니다.

### hash, bitmap, LSM은 같은 index라는 이름 아래 다른 질문에 답한다

Hash index는 equality lookup에 강한 구조입니다. key를 hash 값으로 바꿔 bucket을 찾기 때문에 order와 range를 자연스럽게 제공하지 않습니다. PostgreSQL은 hash index type을 제공하지만 일반적인 범위/정렬 query에는 B-tree가 중심입니다. MySQL에서는 InnoDB adaptive hash index처럼 engine 내부 optimization으로 hash 성격이 보일 수 있고, MEMORY engine hash index 같은 별도 경계도 있습니다. 중요한 것은 hash를 "B+tree보다 빠른 index"로 일반화하지 않는 것입니다. hash는 equality 질문에 맞고, range/order 질문에는 맞지 않습니다.

Bitmap은 두 가지를 구분해야 합니다. PostgreSQL의 bitmap index scan/bitmap heap scan은 여러 index 결과를 memory bitmap으로 합쳐 table row 위치를 물리 순서로 방문하는 실행 전략입니다. 이것은 Oracle의 bitmap index처럼 저장된 bitmap index 구조와 같은 말이 아닙니다. PostgreSQL 문서는 여러 index scan 결과를 bitmap으로 만들고 AND/OR로 결합한 뒤 table row를 physical order로 방문한다고 설명합니다. 이때 원래 index order는 사라지므로 `ORDER BY`가 있으면 별도 sort가 필요할 수 있습니다.

LSM(Log-Structured Merge) tree는 write를 memory와 sequential log/segment로 모은 뒤 compaction으로 정리하는 계열입니다. PostgreSQL과 MySQL InnoDB의 기본 row-store B-tree 모델과는 다른 write/read trade-off를 갖습니다. LSM은 write-heavy workload에 강할 수 있지만, read amplification, compaction pressure, range query behavior, tombstone 처리 같은 비용이 생깁니다. RDBMS 면접에서 LSM을 언급할 때는 "PostgreSQL/MySQL 기본 인덱스가 LSM이다"라고 말하지 말고, RocksDB 계열 storage engine이나 search/NoSQL 계열과 대비하는 구조로 말하는 편이 안전합니다.

### scan, filter, sort, aggregate는 row stream을 단계적으로 바꾼다

실행 plan은 row stream을 만드는 tree입니다. Scan node는 raw row 후보를 만듭니다. Filter는 조건을 통과한 row만 내보냅니다. Sort는 row stream의 순서를 바꿉니다. Aggregate는 여러 row를 grouping key별로 합쳐 더 적은 row를 만듭니다. Join은 두 row stream을 연결합니다.

```text
orders table
  Scan: user_id = 42 후보 rows
  Filter: status = 'PAID'
  Sort: created_at DESC
  Limit: first 20

payments table join
  Join: orders.id = payments.order_id
  Aggregate: sum(payment.amount) by order
```

각 node는 memory와 I/O 조건을 가집니다. Sort와 hash aggregate는 memory가 부족하면 temp file로 spill할 수 있습니다. Nested loop join은 outer row가 적고 inner lookup이 싸면 좋지만, outer row가 많고 inner lookup이 비싸면 폭발합니다. Hash join은 build side를 memory에 올릴 수 있으면 좋지만, build side가 크면 spill이 생깁니다. Merge join은 양쪽 input이 정렬되어 있거나 정렬 비용이 감당될 때 강합니다.

면접에서 join algorithm을 외우듯 나열하기보다, "row stream을 어떤 순서로 만들고, 어느 쪽을 반복해서 읽고, 어떤 memory 구조를 만들며, 언제 spill 또는 random lookup이 생기는가"로 말하는 편이 실무적입니다.

### statistics는 optimizer의 눈이다

Optimizer는 실제로 query를 끝까지 실행해 보고 고르는 것이 아닙니다. 통계와 cost model로 추정합니다. PostgreSQL 문서는 planner가 좋은 plan 선택을 위해 query가 가져올 row 수를 estimate해야 하고, table/index entry 수와 disk block 수, selectivity 정보를 통계에서 사용한다고 설명합니다. `reltuples`, `relpages`, `pg_stats`, most common values, histogram 같은 값이 여기에 들어갑니다. MySQL InnoDB도 persistent optimizer statistics를 disk에 저장해 restart 뒤에도 plan stability를 높이고, 자동 재계산이 지연될 수 있어 즉시 최신 통계가 필요하면 `ANALYZE TABLE`을 실행할 수 있다고 설명합니다.

통계가 틀리면 plan도 틀릴 수 있습니다.

```text
planner estimate
  status='PAID' matches 1% rows
  choose index lookup + nested loop

actual data
  status='PAID' matches 80% rows
  index lookup visits too many rows
  sequential scan or different join might have been cheaper
```

Plan regression은 SQL text가 그대로여도 발생합니다. 데이터가 늘거나 skew가 생기거나, 통계가 낡거나, parameter 값이 바뀌거나, DBMS version과 cost setting이 바뀌면 optimizer의 선택이 달라집니다. 그래서 성능 장애를 볼 때는 query text, bind value, table cardinality, statistics freshness, plan diff, actual rows vs estimated rows를 함께 봅니다.

### optimizer cost는 운영체제 I/O 비용을 간접적으로 본다

옵티마이저의 cost는 실제 Linux block queue를 실시간으로 읽어 계산한 값이 아닙니다. 하지만 cost model이 비교하려는 비용의 뿌리는 결국 운영체제와 storage 경로에 있습니다. Index scan은 후보 row를 빠르게 좁힐 수 있지만, heap이나 clustered page를 여러 곳에서 가져와야 하면 DBMS buffer miss와 OS page cache miss가 늘 수 있습니다. Sequential scan은 더 많은 row를 읽어도 page를 연속으로 읽으므로 OS read-ahead, filesystem layout, storage queue가 더 잘 맞을 수 있습니다.

```text
optimizer compares paths

Path A: index lookup
  index leaf pages are close
  data pages may be scattered
  many buffer misses can become many OS reads

Path B: sequential scan
  reads many pages
  access pattern is predictable
  OS/filesystem/storage can stream or prefetch more easily

Path C: bitmap heap scan
  gathers tuple locations first
  visits heap pages in a more physical order
  may trade memory for fewer random visits
```

이 관점은 "index가 있는데 왜 sequential scan인가요?"라는 질문을 더 정확하게 만듭니다. DBMS는 SQL 의미만이 아니라 page 접근 비용을 추정합니다. 그 page 접근 비용은 DBMS buffer, OS page cache, filesystem, block layer, device 특성의 영향을 받습니다. PostgreSQL의 `seq_page_cost`와 `random_page_cost` 같은 설정은 이 차이를 추상화한 값이고, MySQL도 range optimizer와 InnoDB statistics로 access path의 비용을 추정합니다. 다만 이 값들은 현재 장치 queue 길이, 순간적인 writeback, 다른 process의 I/O 경쟁을 완벽히 반영하지 못합니다.

그래서 plan 진단은 두 층으로 나눠야 합니다. 먼저 `EXPLAIN`으로 optimizer가 어떤 row 수와 access path를 상상했는지 봅니다. 그다음 `EXPLAIN (ANALYZE, BUFFERS)`, MySQL `EXPLAIN ANALYZE`, slow query log, performance schema, OS의 `iostat` 같은 관측으로 실제 page read, cache hit, read latency, writeback pressure를 확인합니다. 추정과 실제가 크게 다르면 통계, parameter, index 설계, cache 상태, OS I/O 압력 중 어디가 어긋났는지 다시 좁혀야 합니다.

### EXPLAIN은 정답표가 아니라 가설 검사 도구다

PostgreSQL `EXPLAIN`은 plan node tree, cost, estimated rows, width를 보여 줍니다. `EXPLAIN ANALYZE`는 실제 실행까지 하며 actual time과 rows를 볼 수 있습니다. MySQL `EXPLAIN`은 table read order, possible keys, chosen key, estimated rows, filtered percentage, Extra 정보를 보여 줍니다. 두 제품의 출력 형식은 다르지만, 공통 질문은 같습니다.

```text
읽어야 할 질문
  어떤 table이 먼저 읽히는가?
  어떤 access path를 쓰는가?
  index condition과 residual filter가 나뉘는가?
  estimated rows와 actual rows 차이가 큰가?
  sort, temp table, materialization, filesort, bitmap heap scan이 있는가?
  LIMIT 때문에 child node를 끝까지 안 읽는가?
```

`EXPLAIN`만 보고도 많은 것을 알 수 있지만, 실제 병목을 보려면 runtime 관측이 필요합니다. PostgreSQL에서는 `EXPLAIN (ANALYZE, BUFFERS)`가 buffer hit/read와 actual rows를 보여 줍니다. MySQL에서는 `EXPLAIN ANALYZE`, optimizer trace, performance schema, slow query log를 함께 볼 수 있습니다. `ANALYZE` 계열은 실제 실행을 동반하므로 운영 query에 붙일 때는 매우 조심해야 합니다. 특히 DML이나 함수 호출처럼 side effect가 있는 statement, 오래 걸리는 SELECT, 큰 sort/hash가 예상되는 query는 격리 환경이나 transaction rollback 경계에서 먼저 확인해야 합니다.

### pagination은 index order와 cursor 안정성 문제다

Offset pagination은 간단합니다.

```sql
SELECT id, created_at
FROM orders
WHERE user_id = 42
ORDER BY created_at DESC, id DESC
LIMIT 20 OFFSET 100000;
```

하지만 offset이 커지면 앞의 100000개를 건너뛰어야 합니다. DBMS는 그 row들을 client에게 반환하지 않을 뿐, order를 만들고 count하며 지나가야 할 수 있습니다. insert/delete가 동시에 일어나면 page 사이에서 중복이나 누락도 생길 수 있습니다.

Keyset pagination은 마지막으로 본 key를 다음 query 조건으로 사용합니다.

```sql
SELECT id, created_at
FROM orders
WHERE user_id = 42
  AND (created_at, id) < (?, ?)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

이 방식은 `(user_id, created_at DESC, id DESC)` index와 잘 맞으면 바로 다음 range로 들어갈 수 있습니다. `id`를 tie-breaker로 넣는 이유는 `created_at`이 같은 row들이 있을 때 순서를 안정적으로 만들기 위해서입니다. `created_at`이 nullable이면 NULL 정렬 위치까지 cursor 조건과 index 설계에 포함해야 합니다. keyset pagination은 arbitrary page number jump에는 불편하지만, 무한 scroll이나 next page에는 훨씬 안정적입니다.

### plan regression은 SQL 문제가 아니라 환경과 데이터 문제일 수 있다

Plan regression을 볼 때는 먼저 "SQL이 느려졌다"를 "어떤 실행 경로가 바뀌었거나, 같은 경로의 실제 비용이 바뀌었다"로 바꿔 말해야 합니다. 같은 SQL이라도 데이터가 100배 늘면 nested loop가 폭발할 수 있고, 통계가 낡으면 optimizer가 small table처럼 착각할 수 있습니다. bind parameter가 평소에는 selective하지만 특정 고객에서는 대부분의 row를 매칭할 수도 있습니다. collation이나 function을 바꾸면 기존 index를 못 쓸 수도 있습니다. 이런 schema와 expression 변경은 [schema migration](05-schema-constraints-migration.md)에서도 배포 전 검증 대상으로 다시 등장합니다.

```text
diagnosis trace

old plan
  Index Scan orders_user_status_created_idx
  estimated rows: 20
  actual rows: 22
  no sort

new plan
  Seq Scan orders
  estimated rows: 500000
  actual rows: 510000
  Sort spilled to disk

possible causes
  status='PAID' selectivity changed
  statistics updated after data skew
  index was dropped or made invisible
  ORDER BY expression no longer matches index order
```

그래서 plan regression 대응은 index 추가만이 아닙니다. `ANALYZE`, extended statistics, query rewrite, composite index order 조정, pagination 방식 변경, DBMS version change 검토, problematic parameter 분리, stale prepared generic plan 확인이 모두 후보가 됩니다.

## DBMS별 경계

PostgreSQL은 다양한 index type을 제공합니다. B-tree, hash, GiST, SP-GiST, GIN, BRIN이 있고, multicolumn, partial, expression, covering index를 지원합니다. B-tree multicolumn index는 leading column 조건이 scan 범위를 줄이는 핵심입니다. Partial index는 query predicate가 index predicate를 함의해야 쓰입니다. Index-only scan은 visibility map이라는 PostgreSQL 특유의 MVCC 경계가 중요합니다. Bitmap scan은 여러 index scan 결과를 memory bitmap으로 결합한 뒤 heap을 물리 순서로 방문하는 실행 전략입니다.

MySQL/InnoDB는 clustered index와 secondary index 경계를 먼저 잡아야 합니다. Primary key가 clustered index가 되며, secondary index record는 primary key columns를 포함합니다. Multiple-column index는 leftmost prefix가 중요합니다. MySQL `EXPLAIN`의 `type`, `key`, `rows`, `filtered`, `Extra`는 access path와 join order를 읽는 기본 단서입니다. InnoDB persistent optimizer statistics는 plan stability를 높이지만, 자동 재계산 지연과 sampling 때문에 실제 분포와 어긋날 수 있습니다.

정렬도 제품별 세부가 다릅니다. PostgreSQL은 index order와 `NULLS FIRST/LAST`, collation, operator class가 plan에 영향을 줍니다. MySQL은 `filesort`라는 Extra가 보일 수 있는데, 이것은 반드시 OS file sort라는 뜻이 아니라 MySQL의 sort algorithm을 넓게 가리키는 용어입니다. `ORDER BY` 없는 결과 순서는 어느 제품에서도 application contract로 삼으면 안 됩니다.

Partial index는 PostgreSQL에서는 강력한 1급 기능이지만, MySQL에서는 같은 문법으로 제공되지 않습니다. MySQL에서는 generated column, functional index, prefix index, invisible index, partitioning, application predicate 분리 등 다른 도구로 비슷한 목적을 달성해야 할 수 있습니다. 따라서 "partial index를 만들면 됩니다"라는 답은 PostgreSQL 문맥인지 먼저 확인해야 합니다.

LSM은 PostgreSQL/MySQL 기본 인덱스 설명에서 직접 적용하지 않습니다. RocksDB 기반 MyRocks 같은 MySQL storage engine이나 Cassandra/RocksDB/LevelDB 계열을 비교할 때 꺼내는 편이 정확합니다. RDBMS row-store의 B-tree 비용과 LSM compaction 비용을 섞어 말하면 면접에서 바로 반례가 나옵니다.

## 직접 재생해 보기

### Composite index와 leftmost prefix 확인

PostgreSQL 또는 MySQL에서 작은 table을 만듭니다.

```sql
CREATE TABLE orders_plan_lab (
  id bigint PRIMARY KEY,
  user_id bigint NOT NULL,
  status varchar(20) NOT NULL,
  created_at timestamp NOT NULL,
  total_amount numeric(12, 2) NOT NULL
);

CREATE INDEX orders_user_status_created_idx
ON orders_plan_lab (user_id, status, created_at DESC, id DESC);
```

이후 두 query를 비교합니다.

```sql
EXPLAIN
SELECT id
FROM orders_plan_lab
WHERE user_id = 42 AND status = 'PAID'
ORDER BY created_at DESC, id DESC
LIMIT 20;

EXPLAIN
SELECT id
FROM orders_plan_lab
WHERE status = 'PAID'
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

PASS 신호는 첫 query가 composite index의 leading equality 조건과 order를 더 잘 활용할 가능성이 크다는 점을 보는 것입니다. 두 번째 query는 leading column `user_id` 조건이 없으므로 같은 방식으로 좁히기 어렵습니다. 실제 plan은 데이터 양과 DBMS에 따라 달라질 수 있으므로, 핵심은 leftmost prefix가 왜 중요한지 설명하는 것입니다.

### PostgreSQL에서 index-only scan 경계 확인

```sql
CREATE TABLE idx_only_lab (
  id integer PRIMARY KEY,
  user_id integer NOT NULL,
  amount integer NOT NULL
);

INSERT INTO idx_only_lab
SELECT g, g % 10, g * 10
FROM generate_series(1, 10000) AS g;

CREATE INDEX idx_only_lab_user_amount_idx
ON idx_only_lab (user_id) INCLUDE (amount);

VACUUM idx_only_lab;

EXPLAIN (ANALYZE, BUFFERS)
SELECT amount
FROM idx_only_lab
WHERE user_id = 3;
```

PASS 신호는 plan에서 index-only scan이 나올 수 있고, heap fetch 여부가 visibility map 상태와 연결된다는 점을 보는 것입니다. FAIL 신호는 include column을 넣었으니 항상 heap을 보지 않는다고 단정하는 것입니다.

### Pagination offset과 keyset 비교

```sql
-- offset 방식
EXPLAIN
SELECT id, created_at
FROM orders_plan_lab
WHERE user_id = 42
ORDER BY created_at DESC, id DESC
LIMIT 20 OFFSET 100000;

-- keyset 방식
EXPLAIN
SELECT id, created_at
FROM orders_plan_lab
WHERE user_id = 42
  AND (created_at, id) < (TIMESTAMP '2026-01-01 00:00:00', 1000000)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

PASS 신호는 offset이 앞 row를 건너뛰는 비용을 만들 수 있고, keyset은 마지막 key 이후 range를 바로 좁힌다는 점을 plan에서 확인하는 것입니다. MySQL에서는 row constructor 비교와 index 사용 가능성을 version과 조건 형태에 맞게 확인해야 합니다. DBMS별 문법 차이는 실험 환경에서 조정합니다.

## 면접 꼬리 질문

1. B+tree index가 있는데도 sequential scan이 선택되는 이유는 무엇인가?

    index lookup이 후보 row를 줄이지 못하거나, 많은 row를 가져와야 해서 random table access가 비싸면 sequential scan이 더 싸게 추정될 수 있습니다. 통계, selectivity, page cost, table size, caching 상태가 영향을 줍니다.

2. composite index에서 column 순서는 어떻게 정하나?

    query의 equality 조건, range 조건, order 요구, selectivity, 재사용성을 함께 봅니다. 보통 leading equality 조건으로 범위를 좁히고, 그 뒤 order column을 배치해 sort를 줄이는 방향을 검토합니다. 다만 write overhead와 다른 query와의 균형도 봐야 합니다.

3. covering index가 항상 좋은가?

    아닙니다. table 방문을 줄일 수 있지만 index가 커지고 write 비용이 늘어납니다. PostgreSQL에서는 visibility map 때문에 index-only scan이 항상 heap access를 없애지 못합니다. InnoDB에서는 secondary index가 covering이면 clustered lookup을 줄일 수 있지만 secondary index size가 커집니다.

4. partial index는 언제 위험한가?

    query predicate가 index predicate와 맞지 않으면 쓰이지 않습니다. parameterized query, predicate 표현 차이, workload 변화 때문에 planner가 함의를 알아보지 못할 수 있습니다. PostgreSQL 기능으로 강력하지만 DBMS 이식성은 낮습니다.

5. `EXPLAIN`에서 가장 먼저 볼 것은 무엇인가?

    access path, join order, estimated rows와 actual rows 차이, filter 위치, sort/temp/spill 단서, chosen key와 possible keys를 봅니다. 단순히 "index를 탔는가"보다 "몇 row를 읽고 몇 row를 버리는가"가 더 중요합니다.

6. OFFSET pagination이 왜 느려질 수 있나?

    앞 row를 반환하지 않을 뿐 plan이 그 row들을 지나가며 order를 유지해야 할 수 있습니다. offset이 커질수록 버리는 작업이 많아집니다. Keyset pagination은 마지막 key를 조건으로 넣어 다음 범위를 바로 좁히므로 큰 offset보다 안정적입니다.

## 함정 질문

1. "B+tree는 O(log N)이므로 table 크기가 커져도 별문제 없나요?"

    tree 높이는 천천히 커지지만, 실제 query 비용은 leaf range length, table lookup, buffer miss, visibility check, sort, join, network transfer에 영향을 받습니다. O(log N) 하나로 운영 성능을 설명하면 부족합니다.

2. "index가 많을수록 조회가 빨라지나요?"

    조회 후보 경로는 늘지만 write 비용, storage, buffer pressure, optimizer 선택 복잡도가 늘어납니다. 중복 index와 거의 쓰지 않는 index는 write-heavy table에서 직접적인 비용이 됩니다.

3. "hash index가 B+tree보다 빠르니 equality 검색에는 무조건 hash가 좋은가요?"

    DBMS별 구현과 WAL/recovery 지원, concurrency, range/order 요구, optimizer maturity를 봐야 합니다. PostgreSQL과 InnoDB 일반 workload에서는 B-tree가 기본 선택인 경우가 많습니다. hash는 equality 중심 질문에 맞지만 범위와 정렬에는 맞지 않습니다.

4. "bitmap index scan은 bitmap index가 있다는 뜻인가요?"

    PostgreSQL에서는 보통 실행 전략으로서 bitmap scan을 뜻합니다. 여러 index scan 결과를 memory bitmap으로 합쳐 heap page를 물리 순서로 방문합니다. 저장 구조로서의 bitmap index와 섞으면 안 됩니다.

5. "EXPLAIN 결과에 cost가 낮으면 실제로도 빠른가요?"

    cost는 optimizer 내부 단위의 추정값입니다. 실제 시간은 cache, concurrent load, I/O latency, lock wait, parameter, 통계 정확도에 영향을 받습니다. `EXPLAIN ANALYZE`나 runtime metrics로 actual row/time/buffer를 확인해야 합니다.

6. "ORDER BY 없는 query가 어제는 id 순서로 나왔는데 오늘 바뀌면 DB 버그인가요?"

    아닙니다. `ORDER BY` 없는 결과 순서는 SQL contract가 아닙니다. access path가 바뀌거나 table이 vacuum/rebuild되거나 page layout이 바뀌면 순서가 달라질 수 있습니다.

## 더 깊게 볼 자료

공식 자료는 index type과 optimizer 동작의 1차 근거입니다. 이 저장소의 기존 `database/` 문서는 query와 interview 흐름을 재구성하는 재료로 보되, 제품별 기능은 공식 문서로 다시 확인합니다.

- PostgreSQL current docs
    - [Indexes](https://www.postgresql.org/docs/current/indexes.html)
    - [Multicolumn Indexes](https://www.postgresql.org/docs/current/indexes-multicolumn.html)
    - [Index-Only Scans and Covering Indexes](https://www.postgresql.org/docs/current/indexes-index-only-scans.html)
    - [Partial Indexes](https://www.postgresql.org/docs/current/indexes-partial.html)
    - [Indexes on Expressions](https://www.postgresql.org/docs/current/indexes-expressional.html)
    - [Combining Multiple Indexes](https://www.postgresql.org/docs/current/indexes-bitmap-scans.html)
    - [Using EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html)
    - [Planner Statistics](https://www.postgresql.org/docs/current/planner-stats.html)
- MySQL 8.4 Reference Manual
    - [Optimization and Indexes](https://dev.mysql.com/doc/refman/8.4/en/optimization-indexes.html)
    - [Multiple-Column Indexes](https://dev.mysql.com/doc/refman/8.4/en/multiple-column-indexes.html)
    - [Clustered and Secondary Indexes](https://dev.mysql.com/doc/refman/8.4/en/innodb-index-types.html)
    - [EXPLAIN Output Format](https://dev.mysql.com/doc/refman/8.4/en/explain-output.html)
    - [Range Optimization](https://dev.mysql.com/doc/refman/8.4/en/range-optimization.html)
    - [Optimizer Statistics](https://dev.mysql.com/doc/refman/8.4/en/optimizer-statistics.html)
    - [ORDER BY Optimization](https://dev.mysql.com/doc/refman/8.4/en/order-by-optimization.html)
    - [LIMIT Query Optimization](https://dev.mysql.com/doc/refman/8.4/en/limit-optimization.html)
- Linux kernel and man-pages
    - [Linux Page Cache](https://docs.kernel.org/next/mm/page_cache.html)
    - [Linux block layer writeback cache control](https://docs.kernel.org/block/writeback_cache_control.html)
    - [fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html)
- Repo study material
    - `database/deep-dive/storage-index-optimizer/07-index-structures.md`
    - `database/deep-dive/storage-index-optimizer/08-query-execution-operators.md`
    - `database/deep-dive/storage-index-optimizer/09-optimizer-statistics-explain.md`
    - `database/mysql/explains/when_join.md`
    - `database/mysql/order.md`
    - `database/join.md`
    - `interviews/database-storage-search-nosql.md`
