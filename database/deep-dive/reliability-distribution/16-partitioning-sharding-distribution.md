# Partitioning, Sharding, Distribution

## partitioning and pruning

partitioning은 큰 테이블을 작은 관리 단위로 나누는 기능이지만, 자동 성능 향상 버튼은 아니다. partition key가 query predicate와 만나면 DB는 읽지 않아도 되는 partition을 제외할 수 있고, 오래된 partition을 통째로 drop하거나 archive할 수도 있다. 반대로 query가 partition key를 쓰지 않거나 partition 수가 지나치게 늘면, 작은 구조가 많이 생긴 만큼 plan, statistics, lock, maintenance 비용이 늘어난다.

이 절에서 공식 근거로 삼은 자료는 다음과 같다. 링크는 단순 참고가 아니라 본문 판단의 경계다. PostgreSQL과 MySQL은 같은 단어를 쓰더라도 로그 이름, 보존 책임, failover 절차, 관측 뷰가 다르므로, 공통 원리와 제품별 실행 방법을 분리해 읽어야 한다.

- PostgreSQL 18 Table Partitioning: https://www.postgresql.org/docs/current/ddl-partitioning.html
- MySQL 8.4 Partitioning: https://dev.mysql.com/doc/refman/8.4/en/partitioning.html
- MySQL 8.4 Partition Pruning: https://dev.mysql.com/doc/refman/8.4/en/partitioning-pruning.html
- 로컬 seed: database/database-deep-study-plan.md

partitioning이 등장한 배경은 단순히 row 수가 많아서가 아니다. 운영자는 오래된 데이터를 빨리 지우고 싶고, optimizer는 조건에 맞지 않는 범위를 읽지 않고 싶고, DBA는 큰 index rebuild나 vacuum/analyze를 더 작은 단위로 나누고 싶다. 이 세 요구가 같은 partition key에 맞춰질 때 partitioning은 강해진다. 예를 들어 로그 테이블에서 `created_at`은 월별 보존 삭제, 최근 기간 조회, 월별 archive라는 요구를 동시에 만족시킬 수 있다.

가장 작은 모델은 range partition이다. 하나의 논리 테이블 `orders`가 있고, 내부에는 `orders_2026_05`, `orders_2026_06` 같은 물리 partition이 있다. 애플리케이션은 보통 부모 테이블에 insert/select를 실행하지만, DB는 partition bound를 보고 어떤 자식 partition이 대상인지 결정한다. pruning은 이 bound와 query predicate를 비교해 읽을 필요가 없는 partition을 plan에서 제외하는 일이다.

```sql
CREATE TABLE orders (
    order_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at date NOT NULL,
    amount numeric(12,2) NOT NULL,
    status text NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE orders_2026_05 PARTITION OF orders
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE orders_2026_06 PARTITION OF orders
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

EXPLAIN SELECT sum(amount)
FROM orders
WHERE created_at >= DATE '2026-06-10'
  AND created_at <  DATE '2026-06-11';
```

이 쿼리에서 optimizer가 `orders_2026_05`를 제외할 수 있는 이유는 5월 partition의 bound가 `2026-06-10 <= created_at < 2026-06-11` 조건과 겹치지 않는다는 것을 증명할 수 있기 때문이다. PostgreSQL 공식 문서는 partition pruning을 declarative partitioned table의 성능을 개선하는 query optimization이라고 설명하며, planner가 각 partition definition을 보고 WHERE 절을 만족할 row가 없음을 증명하면 plan에서 제외한다고 말한다. 핵심 단어는 '증명'이다. DB는 사람이 의도한 기간 조회를 읽는 것이 아니라 조건식과 partition bound의 수학적 관계를 본다.

| query shape | partition key와 만나는 방식 | pruning 기대 | 주의점 |
| --- | --- | --- | --- |
| created_at = 2026-06-10 | range bound와 직접 비교 | 강함 | 날짜 타입과 timezone 변환 주의 |
| created_at BETWEEN 2026-06-01 AND 2026-06-30 | 연속 range | 강함 | 상한 inclusive/exclusive 실수 주의 |
| date_trunc(month, created_at)=2026-06-01 | partition column에 함수 적용 | 약하거나 불가 | expression partition이 아니면 증명 어려움 |
| user_id = 10 | partition key와 무관 | 없음 | 모든 월 partition scan 가능 |
| created_at parameter | 실행 시 parameter 확정 | DBMS별 plan/runtime pruning 확인 | EXPLAIN ANALYZE 필요 |

MySQL pruning도 같은 사고방식이다. 공식 문서는 optimizer가 WHERE 조건을 `partition_column = constant` 또는 `partition_column IN (...)` 같은 형태로 줄일 수 있으면 대상 partition만 scan한다고 설명한다. range 조건도 작은 값 목록으로 줄일 수 있으면 pruning될 수 있다. 그러나 HASH/KEY partition에서 range가 partition 수보다 넓거나 partition expression과 맞지 않으면 기대만큼 줄지 않는다. PostgreSQL과 MySQL 모두 'partitioned table이다'가 아니라 '이 쿼리가 partition definition과 만난다'가 성능 조건이다.

```text
pruning trace

partitions:
  p202605: 2026-05-01 <= created_at < 2026-06-01
  p202606: 2026-06-01 <= created_at < 2026-07-01
  p202607: 2026-07-01 <= created_at < 2026-08-01

query predicate:
  2026-06-10 <= created_at < 2026-06-11

proof:
  p202605 upper bound 2026-06-01 <= query lower bound 2026-06-10 -> cannot match
  p202606 overlaps query range -> scan
  p202607 lower bound 2026-07-01 >= query upper bound 2026-06-11 -> cannot match

scan set:
  p202606 only
```

partitioning은 maintenance boundary로도 중요하다. 월별 partition이 있으면 오래된 데이터를 `DELETE FROM orders WHERE created_at < ...`로 row마다 지우지 않고 partition detach/drop으로 처리할 수 있다. 이 방식은 undo/redo/WAL/binlog 생성량과 lock 시간을 크게 줄일 수 있다. 하지만 foreign key, backup, replication, archive, application query가 partition 단위를 알고 있어야 한다. 예를 들어 drop한 partition이 아직 법적 보존 대상이면 빠른 삭제가 오히려 사고다.

unique constraint는 partitioning의 실전 함정이다. 업무적으로 `order_no`가 전체 테이블에서 유일해야 하는데 partition key가 `created_at`이면, DBMS는 partition별 local uniqueness만 쉽게 보장할 수 있고 전체 유일성을 보장하려면 partition key를 unique key에 포함하라는 제한을 둘 수 있다. 이 제한은 불편한 문법이 아니라 물리적으로 다른 partition에 같은 order_no가 들어오는지 한 번에 확인하기 어렵다는 문제에서 나온다. 따라서 partition 설계는 query pruning뿐 아니라 업무 invariant까지 같이 봐야 한다.

```sql
-- 관측 루틴 예시: PostgreSQL
SET enable_partition_pruning = on;
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders
WHERE created_at >= DATE '2026-06-10'
  AND created_at <  DATE '2026-06-11';

SET enable_partition_pruning = off;
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders
WHERE created_at >= DATE '2026-06-10'
  AND created_at <  DATE '2026-06-11';

-- 관측 루틴 예시: MySQL
EXPLAIN PARTITIONS
SELECT * FROM orders
WHERE created_at >= '2026-06-10'
  AND created_at <  '2026-06-11';

SELECT TABLE_NAME, PARTITION_NAME, TABLE_ROWS
FROM INFORMATION_SCHEMA.PARTITIONS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'orders';
```

PASS 신호는 plan에서 읽는 partition이 predicate와 일치하고, pruning을 껐을 때보다 scan 범위와 buffer read가 줄어드는 것이다. FAIL 신호는 partitioned table인데 plan이 모든 partition을 읽거나, 함수/형 변환 때문에 partition key 조건이 증명되지 않거나, pruning은 되지만 hot partition 하나가 모든 write를 받아 병목이 되는 것이다. 특히 최신 데이터만 쓰는 workload에서 월별 partition은 현재 월 partition 하나를 뜨겁게 만들 수 있다.

partitioning과 indexing은 대체 관계가 아니다. partition pruning은 '어느 partition을 볼 것인가'를 줄이고, index는 '그 partition 안에서 어느 row를 빨리 찾을 것인가'를 줄인다. 월별 partition이 있어도 특정 사용자 주문을 찾는 쿼리는 각 월 partition 안의 `user_id` index가 필요할 수 있다. 반대로 좋은 composite index가 있어도 보존 삭제와 대용량 maintenance 경계를 위해 partitioning이 필요할 수 있다.

문서를 덮고 다시 설명할 때는 이렇게 말할 수 있어야 한다. partitioning은 데이터를 물리적으로 나누는 기능이고, pruning은 query 조건과 partition bound가 만나 읽지 않아도 되는 조각을 제외하는 최적화다. partition key는 query, retention, maintenance, uniqueness, write hotspot 사이의 trade-off를 고정한다. 그래서 설계 순서는 '큰 테이블이니까 나눈다'가 아니라 '어떤 조건으로 읽고, 어떤 단위로 지우고, 어떤 invariant를 지키며, 어떤 관측으로 검증할 것인가'다.

### 설계 리허설 카드: partition key가 만드는 이득과 비용

1. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

2. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

3. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

4. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

5. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

6. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

7. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

8. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

9. hot partition은 partitioning이 만든 새로운 병목이다. 최신 주문만 계속 쓰는 월별 partition은 현재 월 partition 하나에 쓰기가 몰린다. 이 문제는 과거 partition drop을 쉽게 만드는 장점과 동시에 발생한다. write hot spot을 줄이려면 시간 key만으로 충분한지, hash subpartition이나 shard key가 필요한지 별도로 판단해야 한다.

10. partitioning과 sharding을 섞어 말하면 설계가 흐려진다. partitioning은 보통 같은 DB instance 안에서 table storage와 plan 범위를 나누는 기능이고, sharding은 여러 노드나 cluster에 데이터를 분산해 ownership 자체를 나누는 설계다. 전자는 optimizer와 DDL 관리가 중심이고, 후자는 routing, transaction, resharding, global index 문제가 중심이다.

11. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

12. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

13. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

14. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

15. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

16. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

17. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

18. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

19. hot partition은 partitioning이 만든 새로운 병목이다. 최신 주문만 계속 쓰는 월별 partition은 현재 월 partition 하나에 쓰기가 몰린다. 이 문제는 과거 partition drop을 쉽게 만드는 장점과 동시에 발생한다. write hot spot을 줄이려면 시간 key만으로 충분한지, hash subpartition이나 shard key가 필요한지 별도로 판단해야 한다.

20. partitioning과 sharding을 섞어 말하면 설계가 흐려진다. partitioning은 보통 같은 DB instance 안에서 table storage와 plan 범위를 나누는 기능이고, sharding은 여러 노드나 cluster에 데이터를 분산해 ownership 자체를 나누는 설계다. 전자는 optimizer와 DDL 관리가 중심이고, 후자는 routing, transaction, resharding, global index 문제가 중심이다.

21. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

22. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

23. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

24. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

25. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

26. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

27. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

28. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

29. hot partition은 partitioning이 만든 새로운 병목이다. 최신 주문만 계속 쓰는 월별 partition은 현재 월 partition 하나에 쓰기가 몰린다. 이 문제는 과거 partition drop을 쉽게 만드는 장점과 동시에 발생한다. write hot spot을 줄이려면 시간 key만으로 충분한지, hash subpartition이나 shard key가 필요한지 별도로 판단해야 한다.

30. partitioning과 sharding을 섞어 말하면 설계가 흐려진다. partitioning은 보통 같은 DB instance 안에서 table storage와 plan 범위를 나누는 기능이고, sharding은 여러 노드나 cluster에 데이터를 분산해 ownership 자체를 나누는 설계다. 전자는 optimizer와 DDL 관리가 중심이고, 후자는 routing, transaction, resharding, global index 문제가 중심이다.

31. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

32. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

33. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

34. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

35. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

36. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

37. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

38. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

39. hot partition은 partitioning이 만든 새로운 병목이다. 최신 주문만 계속 쓰는 월별 partition은 현재 월 partition 하나에 쓰기가 몰린다. 이 문제는 과거 partition drop을 쉽게 만드는 장점과 동시에 발생한다. write hot spot을 줄이려면 시간 key만으로 충분한지, hash subpartition이나 shard key가 필요한지 별도로 판단해야 한다.

40. partitioning과 sharding을 섞어 말하면 설계가 흐려진다. partitioning은 보통 같은 DB instance 안에서 table storage와 plan 범위를 나누는 기능이고, sharding은 여러 노드나 cluster에 데이터를 분산해 ownership 자체를 나누는 설계다. 전자는 optimizer와 DDL 관리가 중심이고, 후자는 routing, transaction, resharding, global index 문제가 중심이다.

41. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

42. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

43. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

44. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

45. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

46. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

47. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

48. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

49. hot partition은 partitioning이 만든 새로운 병목이다. 최신 주문만 계속 쓰는 월별 partition은 현재 월 partition 하나에 쓰기가 몰린다. 이 문제는 과거 partition drop을 쉽게 만드는 장점과 동시에 발생한다. write hot spot을 줄이려면 시간 key만으로 충분한지, hash subpartition이나 shard key가 필요한지 별도로 판단해야 한다.

50. partitioning과 sharding을 섞어 말하면 설계가 흐려진다. partitioning은 보통 같은 DB instance 안에서 table storage와 plan 범위를 나누는 기능이고, sharding은 여러 노드나 cluster에 데이터를 분산해 ownership 자체를 나누는 설계다. 전자는 optimizer와 DDL 관리가 중심이고, 후자는 routing, transaction, resharding, global index 문제가 중심이다.

51. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

52. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

53. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

54. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

55. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

56. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

57. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

58. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

59. hot partition은 partitioning이 만든 새로운 병목이다. 최신 주문만 계속 쓰는 월별 partition은 현재 월 partition 하나에 쓰기가 몰린다. 이 문제는 과거 partition drop을 쉽게 만드는 장점과 동시에 발생한다. write hot spot을 줄이려면 시간 key만으로 충분한지, hash subpartition이나 shard key가 필요한지 별도로 판단해야 한다.

60. partitioning과 sharding을 섞어 말하면 설계가 흐려진다. partitioning은 보통 같은 DB instance 안에서 table storage와 plan 범위를 나누는 기능이고, sharding은 여러 노드나 cluster에 데이터를 분산해 ownership 자체를 나누는 설계다. 전자는 optimizer와 DDL 관리가 중심이고, 후자는 routing, transaction, resharding, global index 문제가 중심이다.

61. partition key는 데이터를 나누는 이름표이면서 optimizer가 일부 조각을 건너뛰기 위한 증명 재료다. WHERE 절이 partition bound와 맞물리지 않으면 DB는 많은 partition을 그대로 본다. 그래서 partitioning은 자동 성능 버튼이 아니라 query shape, retention 정책, maintenance 경계를 partition key에 맞추는 설계다.

62. 월별 range partition은 오래된 데이터를 drop하기 쉽지만, 사용자별 조회가 대부분이면 특정 사용자의 데이터가 여러 달 partition에 흩어진다. 반대로 user_id hash partition은 사용자별 부하를 나누기 좋지만, 월별 보존 삭제는 여러 partition을 훑어야 한다. partition key는 성능과 운영 중 무엇을 먼저 고정할지의 선택이다.

63. PostgreSQL partition pruning은 planner가 partition definition을 보고 WHERE 조건과 겹치지 않는 partition을 제외하는 최적화다. plan time뿐 아니라 실행 중 parameter 값이 확정될 때 pruning될 수도 있다. 그래서 prepared statement나 nested loop 내부 parameter에서는 EXPLAIN ANALYZE의 Subplans Removed, loops, never executed 표시까지 봐야 한다.

64. MySQL pruning은 partition expression이 상수 비교나 작은 IN 목록으로 줄어들 때 강하다. 하지만 함수를 잘못 감싸거나 range가 partition 수보다 커지는 모양이 되면 기대한 만큼 줄지 않는다. MySQL 문서의 예처럼 partitioning expression과 WHERE 조건이 같은 계산 모델로 만나는지 확인해야 한다.

65. partition 수가 많아지면 각각의 index, statistics, metadata, lock, plan 후보도 늘어난다. 하나의 큰 테이블을 여러 작은 물리 구조로 나누는 일은 관리 단위를 늘리는 일이다. pruning이 되지 않는 쿼리는 오히려 작은 테이블을 많이 여는 비용을 추가로 낸다.

66. unique constraint와 primary key는 partition key와 충돌할 수 있다. 전체 테이블에서 유일해야 하는 값이 partition 내부에서만 유일하면 업무 invariant가 깨진다. DBMS마다 제한은 다르지만, partitioned table의 unique index가 partition key를 포함해야 하는 이유를 업무 키 설계와 함께 읽어야 한다.

67. partition maintenance는 DDL이지만 운영 traffic과 만난다. 새 partition을 미리 만들지 않으면 insert가 실패하거나 default partition으로 몰리고, 오래된 partition drop은 빠른 보존 삭제 수단이지만 foreign key, replication, backup, archiving 정책과 엮인다. 자동화는 날짜 계산보다 검증 쿼리가 중요하다.

68. partition pruning을 확인할 때는 쿼리 결과가 빠르다는 체감보다 plan이 실제로 어떤 partition을 읽었는지 봐야 한다. PostgreSQL에서는 EXPLAIN에 남은 child scan, Subplans Removed, execution loop를 보고, MySQL에서는 EXPLAIN PARTITIONS 또는 optimizer trace로 접근 partition을 확인한다. 관측 없는 partitioning은 추측이다.

## sharding, resharding, global transaction trade-off

sharding은 데이터를 여러 노드로 나누어 저장하는 설계지만, 단순히 table을 여러 개로 쪼개는 작업이 아니다. 한 row가 어느 노드의 책임인지 정하는 순간 routing, transaction, index, join, sequence, rebalancing, 장애 관측이 모두 바뀐다. 이 절은 로컬 seed `database/newsql.md`의 NewSQL 감각을 살리되, sharding과 distributed SQL이 어떤 문제를 시스템 내부 또는 애플리케이션 설계로 옮기는지 설명한다.

이 절에서 공식 근거로 삼은 자료는 다음과 같다. 링크는 단순 참고가 아니라 본문 판단의 경계다. PostgreSQL과 MySQL은 같은 단어를 쓰더라도 로그 이름, 보존 책임, failover 절차, 관측 뷰가 다르므로, 공통 원리와 제품별 실행 방법을 분리해 읽어야 한다.

- CockroachDB Architecture Overview: https://www.cockroachlabs.com/docs/stable/architecture/overview
- TiDB Architecture: https://docs.pingcap.com/tidb/stable/tidb-architecture/
- 로컬 seed: database/newsql.md

sharding이 등장한 배경은 단일 노드의 저장 용량, write throughput, 장애 반경, 지역 latency 한계다. partitioning이 같은 DB 안에서 물리 조각을 나누는 일이라면, sharding은 여러 노드나 cluster에 데이터 ownership을 나누는 일이다. 그래서 shard key는 optimizer hint가 아니라 데이터의 주소가 된다. 이 주소를 잘못 고르면 특정 shard가 뜨거워지고, cross-shard join이 늘어나고, resharding 때 서비스가 흔들린다.

```text
naive single-node model

  app -> DB node A
         orders table
         users table
         payments table

sharded ownership model by user_id

  app -> router -> shard 0: user_id % 4 = 0
                -> shard 1: user_id % 4 = 1
                -> shard 2: user_id % 4 = 2
                -> shard 3: user_id % 4 = 3

changed responsibility:
  DB no longer owns every row in one place.
  router and shard map become part of the correctness path.
```

가장 작은 예로 `user_id % 4` sharding을 보자. 사용자 10의 주문은 shard 2에 저장된다. 사용자 11의 주문은 shard 3에 저장된다. 사용자별 주문 조회는 route가 단순하다. 하지만 전체 매출 집계는 네 shard를 모두 읽어 합쳐야 한다. 두 사용자의 계좌를 동시에 바꾸는 transaction은 shard 2와 shard 3을 함께 바꿔야 한다. 단일 DB에서는 평범한 transaction이었지만, shard가 나뉘면 coordinator가 필요한 분산 transaction이 된다.

| 설계 요소 | 단일 DB에서의 감각 | sharding 후 바뀌는 점 | 실무 함정 |
| --- | --- | --- | --- |
| transaction | 한 connection, 한 transaction manager | 여러 shard 참여와 coordinator 필요 | 부분 commit/abort 복구 |
| secondary index | 같은 DB 안 index lookup | global index 또는 fan-out 필요 | index와 원본 row 불일치 |
| join | optimizer가 plan 선택 | shard 간 data 이동 또는 application join | 네트워크 비용 폭증 |
| sequence | auto increment 사용 | 전역 ID 전략 필요 | 순서와 유일성 혼동 |
| backup | 한 cluster 기준 | shard별 backup과 일관 시점 필요 | shard 간 time skew |
| observability | DB instance 단위 | shard별 latency/hot key/rebalance 관측 | 평균 지표가 장애를 숨김 |

resharding은 shard 수를 바꾸는 일인데, 가장 위험한 단계는 data copy가 아니라 route 전환이다. old shard에서 new shard로 row를 복사하는 동안 새 write가 계속 들어온다. 이 write를 old shard에만 쓰면 copy가 뒤처지고, new shard에도 같이 쓰면 dual-write 일관성 문제가 생긴다. change stream으로 따라잡는 방식도 cutover 순간에는 old route를 막고 new route를 열어야 한다. 그래서 resharding 계획에는 snapshot, catch-up, validation, cutover, rollback이 모두 있어야 한다.

```text
resharding timeline from 4 shards to 8 shards

T0 freeze plan:
   shard_map_v1: user_id % 4
   shard_map_v2: user_id % 8

T1 snapshot copy:
   copy rows from old shard 2 to new shard 2 and 6

T2 change capture:
   replay writes that happened after snapshot start

T3 dual-read validation:
   count, checksum, sample aggregate compare old vs new

T4 cutover:
   router switches user_id route to shard_map_v2

T5 guard window:
   old shards reject writes or forward to new owner

T6 cleanup:
   remove old duplicated ranges only after rollback window closes
```

global transaction trade-off는 sharding의 핵심이다. 두 shard 이상을 한 transaction으로 묶으려면 참여자들이 prepare 단계에서 commit 가능 상태를 기록하고, coordinator가 최종 commit 또는 rollback 결정을 내려야 한다. 이 방식은 원자성을 높이지만 latency와 recovery 책임을 키운다. 반대로 transaction을 local shard 안으로 제한하면 성능과 단순성은 좋아지지만, 모델링이 shard key에 묶인다. 좋은 설계는 가능한 많은 aggregate 변경을 한 shard 안에 두고, cross-shard 업무는 saga, outbox, reconciliation 같은 별도 패턴으로 명시한다.

distributed SQL은 이 어려움을 시스템 내부로 상당 부분 가져간다. CockroachDB 공식 문서는 SQL statement를 KV operation으로 바꾸고, keyspace를 range라는 연속 구간으로 나누며, range replica들이 consensus로 commit에 동의한다고 설명한다. TiDB 공식 문서는 TiDB server가 stateless SQL layer이고, TiKV가 storage를 맡으며, PD가 metadata와 scheduling, transaction ID 할당을 맡는다고 설명한다. 두 구조 모두 애플리케이션이 직접 shard 위치를 덜 다루게 해 주지만, 물리적 분산 비용이 사라지는 것은 아니다.

| 제품/모델 | 데이터 분산 단위 | transaction 감각 | 애플리케이션이 여전히 봐야 할 것 |
| --- | --- | --- | --- |
| 수동 sharding | 직접 정한 shard key와 shard map | local 우선, cross-shard는 별도 설계 | route map, resharding, global index |
| CockroachDB | KV keyspace range | range replica consensus와 transactional layer | hot range, follower/leaseholder, region latency |
| TiDB | TiKV region과 PD scheduling | 분산 transaction과 timestamp/metadata 서비스 | PD/TiKV/TiDB 병목, region split, hotspot |
| Vitess류 middleware | keyspace/shard와 routing layer | query routing과 일부 scatter/gather | unsupported query shape, vindex 설계 |

global index는 sharding을 처음 배우는 사람이 자주 놓치는 함정이다. `order_id`로 shard를 나누면 `user_id`로 주문 목록을 찾을 때 모든 shard를 뒤져야 할 수 있다. 이를 피하려고 `user_id -> order_id/shard` lookup index를 별도로 만들면, 주문 원본 row와 index row를 함께 갱신해야 한다. 둘을 같은 shard에 둘 수 없으면 원자성 문제가 생긴다. 즉 읽기를 빠르게 만드는 index가 쓰기 경로를 분산 transaction으로 끌고 갈 수 있다.

```sql
-- sharding 설계 검토용 pseudo SQL: 실제 제품 문법이 아니라 질문을 고정하기 위한 예시
-- 1. 단일 shard로 닫히는 쿼리인가?
SELECT * FROM orders WHERE user_id = 42 AND created_at >= CURRENT_DATE - INTERVAL '7 days';

-- 2. scatter/gather가 필요한 쿼리인가?
SELECT status, count(*) FROM orders WHERE created_at = CURRENT_DATE GROUP BY status;

-- 3. global index가 필요한 접근인가?
SELECT * FROM orders WHERE external_payment_id = 'pay_20260519_001';

-- 4. cross-shard transaction이 되는 업무인가?
UPDATE wallet SET balance = balance - 100 WHERE user_id = 42;
UPDATE wallet SET balance = balance + 100 WHERE user_id = 77;
```

관측은 평균 latency가 아니라 shard별 분포로 봐야 한다. 전체 p95가 괜찮아도 shard 6의 p99가 폭발하면 특정 사용자군만 장애를 겪는다. hot key, range split backlog, transaction retry rate, coordinator abort, cross-region round trip, shard map version mismatch 같은 지표가 필요하다. distributed SQL에서도 node alive만 보면 부족하다. range leaseholder가 한 node에 몰렸는지, PD scheduling이 밀리는지, TiKV region이 split/merge 중인지, consensus quorum이 느린 region에 걸렸는지를 봐야 한다.

sharding의 오해를 고치려면 이렇게 기억하면 된다. table을 나누는 것은 결과이고, 먼저 바뀌는 것은 책임 경계다. 어느 component가 row의 위치를 알고, 어느 component가 두 위치를 함께 바꾸며, 어느 component가 실패 후 다시 reconcile하는지 정해야 한다. 이 질문에 답하지 못하면 shard 수를 늘린 순간 성능 문제가 correctness 문제로 변한다.

### 분산 설계 리허설 카드: shard key 이후에 따라오는 책임

1. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

2. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

3. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

4. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

5. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

6. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

7. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

8. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.

9. multi-tenant 시스템에서 tenant_id shard는 이해하기 쉽고 blast radius를 줄인다. 하지만 큰 tenant 하나가 shard를 독점하면 균등 분산이 깨지고, tenant 간 join이나 global report가 어려워진다. tenant별 이동 가능성을 처음부터 route table과 migration 절차에 넣어 두지 않으면 큰 고객이 들어온 뒤 구조를 바꾸기 어렵다.

10. shard 장애 관측은 노드 alive만으로 부족하다. 특정 shard의 queue depth, p95 latency, replica quorum, split/merge backlog, hot key, cross-shard transaction abort rate가 필요하다. 사용자는 전체 서비스가 느리다고 말하지만 실제 원인은 shard 17 하나의 range split 지연일 수 있다.

11. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

12. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

13. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

14. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

15. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

16. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

17. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

18. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.

19. multi-tenant 시스템에서 tenant_id shard는 이해하기 쉽고 blast radius를 줄인다. 하지만 큰 tenant 하나가 shard를 독점하면 균등 분산이 깨지고, tenant 간 join이나 global report가 어려워진다. tenant별 이동 가능성을 처음부터 route table과 migration 절차에 넣어 두지 않으면 큰 고객이 들어온 뒤 구조를 바꾸기 어렵다.

20. shard 장애 관측은 노드 alive만으로 부족하다. 특정 shard의 queue depth, p95 latency, replica quorum, split/merge backlog, hot key, cross-shard transaction abort rate가 필요하다. 사용자는 전체 서비스가 느리다고 말하지만 실제 원인은 shard 17 하나의 range split 지연일 수 있다.

21. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

22. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

23. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

24. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

25. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

26. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

27. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

28. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.

29. multi-tenant 시스템에서 tenant_id shard는 이해하기 쉽고 blast radius를 줄인다. 하지만 큰 tenant 하나가 shard를 독점하면 균등 분산이 깨지고, tenant 간 join이나 global report가 어려워진다. tenant별 이동 가능성을 처음부터 route table과 migration 절차에 넣어 두지 않으면 큰 고객이 들어온 뒤 구조를 바꾸기 어렵다.

30. shard 장애 관측은 노드 alive만으로 부족하다. 특정 shard의 queue depth, p95 latency, replica quorum, split/merge backlog, hot key, cross-shard transaction abort rate가 필요하다. 사용자는 전체 서비스가 느리다고 말하지만 실제 원인은 shard 17 하나의 range split 지연일 수 있다.

31. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

32. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

33. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

34. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

35. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

36. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

37. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

38. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.

39. multi-tenant 시스템에서 tenant_id shard는 이해하기 쉽고 blast radius를 줄인다. 하지만 큰 tenant 하나가 shard를 독점하면 균등 분산이 깨지고, tenant 간 join이나 global report가 어려워진다. tenant별 이동 가능성을 처음부터 route table과 migration 절차에 넣어 두지 않으면 큰 고객이 들어온 뒤 구조를 바꾸기 어렵다.

40. shard 장애 관측은 노드 alive만으로 부족하다. 특정 shard의 queue depth, p95 latency, replica quorum, split/merge backlog, hot key, cross-shard transaction abort rate가 필요하다. 사용자는 전체 서비스가 느리다고 말하지만 실제 원인은 shard 17 하나의 range split 지연일 수 있다.

41. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

42. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

43. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

44. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

45. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

46. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

47. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

48. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.

49. multi-tenant 시스템에서 tenant_id shard는 이해하기 쉽고 blast radius를 줄인다. 하지만 큰 tenant 하나가 shard를 독점하면 균등 분산이 깨지고, tenant 간 join이나 global report가 어려워진다. tenant별 이동 가능성을 처음부터 route table과 migration 절차에 넣어 두지 않으면 큰 고객이 들어온 뒤 구조를 바꾸기 어렵다.

50. shard 장애 관측은 노드 alive만으로 부족하다. 특정 shard의 queue depth, p95 latency, replica quorum, split/merge backlog, hot key, cross-shard transaction abort rate가 필요하다. 사용자는 전체 서비스가 느리다고 말하지만 실제 원인은 shard 17 하나의 range split 지연일 수 있다.

51. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

52. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

53. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

54. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

55. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

56. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

57. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

58. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.

59. multi-tenant 시스템에서 tenant_id shard는 이해하기 쉽고 blast radius를 줄인다. 하지만 큰 tenant 하나가 shard를 독점하면 균등 분산이 깨지고, tenant 간 join이나 global report가 어려워진다. tenant별 이동 가능성을 처음부터 route table과 migration 절차에 넣어 두지 않으면 큰 고객이 들어온 뒤 구조를 바꾸기 어렵다.

60. shard 장애 관측은 노드 alive만으로 부족하다. 특정 shard의 queue depth, p95 latency, replica quorum, split/merge backlog, hot key, cross-shard transaction abort rate가 필요하다. 사용자는 전체 서비스가 느리다고 말하지만 실제 원인은 shard 17 하나의 range split 지연일 수 있다.

61. sharding은 table을 여러 개로 나누는 작업이 아니라 데이터 ownership을 여러 노드에 배치하는 일이다. ownership이 나뉘면 transaction coordinator, query router, global index, sequence generator, rebalance worker, observability가 같이 필요해진다. 단순한 modulo 분산도 운영 시간이 지나면 shard 추가와 키 재배치 문제를 피할 수 없다.

62. range shard는 지역성 있는 조회와 range scan에 유리하지만 특정 range가 뜨거워질 수 있다. hash shard는 분산이 쉽지만 range query와 운영자가 직관적으로 범위를 이해하는 능력을 잃는다. directory shard는 유연하지만 routing table이 새로운 SSOT가 된다. shard key는 단순 컬럼 선택이 아니라 장애와 운영 방식의 선택이다.

63. resharding은 새 shard를 만들고 데이터를 복사하는 작업만이 아니다. 복사 중 들어오는 새 write를 어디에 쓸지, dual-write를 할지, change stream을 따라잡을지, cutover 순간 old route와 new route를 어떻게 동시에 막을지 정해야 한다. 이 과정을 문서화하지 않으면 데이터가 양쪽에 갈라진다.

64. global transaction은 shard가 많아질수록 비용이 커진다. 한 shard 안의 local transaction은 DB가 익숙하게 처리하지만, 두 shard 이상을 바꾸려면 coordinator가 prepare, commit, abort, recovery log를 관리해야 한다. 그래서 분산 시스템은 가능하면 한 aggregate가 한 shard 안에 머물도록 모델링하려고 한다.

65. global secondary index는 읽기를 편하게 만들지만 쓰기 경로를 늘린다. 주문을 user_id shard에 저장하면서 email이나 order_no로 전역 조회하려면 별도 index shard나 lookup table이 필요하고, 그 index와 원본 row의 원자성 문제가 생긴다. 조회 편의는 쓰기 일관성 비용으로 돌아온다.

66. sequence와 auto increment는 단일 DB에서는 사소하지만 shard에서는 전역 순서와 충돌한다. shard별 sequence는 중복을 피할 수 있어도 전역 시간순을 보장하지 않고, 중앙 sequence는 병목이 된다. Snowflake류 ID, UUID, hi/lo allocation은 이 문제를 다른 trade-off로 푸는 방식이다.

67. CockroachDB와 TiDB 같은 distributed SQL은 sharding의 많은 운영 부담을 시스템 내부로 가져간다. CockroachDB는 SQL을 KV range로 바꾸고 range를 consensus로 복제한다. TiDB는 stateless SQL layer와 TiKV storage, PD metadata/scheduling으로 나눈다. 그러나 추상화가 사라지는 것은 아니며, cross-range transaction과 hotspot은 여전히 설계 주제가 된다.

68. NewSQL을 쓰면 애플리케이션이 shard 위치를 덜 의식할 수 있지만, latency가 단일 노드 DB처럼 행동한다고 가정하면 실패한다. consensus quorum, timestamp allocation, distributed execution, region placement가 query latency에 들어온다. 분산 SQL의 장점은 투명성이고, 비용은 물리 거리와 합의 과정을 완전히 숨길 수 없다는 점이다.
