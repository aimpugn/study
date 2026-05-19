# 쿼리 실행 연산자를 row stream으로 읽기


이 문서는 인덱스와 실행 계획을 실제 row 이동과 관측값으로 읽기 위한 장문 학습 문서다. 각 절은 정의 암기가 아니라 작은 trace를 따라가며 구조, 비용, 장애 신호를 함께 설명한다.


## 문서 안에서 먼저 잡을 목차


- [scan, filter, projection, sort, aggregate](#scan-filter-projection-sort-aggregate)
- [nested loop/hash/merge join](#nested-loophashmerge-join)
- [subquery, CTE, window, pagination](#subquery-cte-window-pagination)




## scan, filter, projection, sort, aggregate

등장 배경은 SQL이 선언적 언어라는 점에 있다. 사용자는 어떤 결과를 원하는지 쓰지만, DBMS는 그 결과를 만들기 위해 row를 어디서 읽고, 어느 조건에서 버리고, 어떤 열만 들고 올라가고, 언제 정렬하거나 묶을지 실행 경로로 바꿔야 한다. 이 변환이 보이지 않으면 SELECT 문은 마치 한 번에 결과표로 바뀌는 것처럼 느껴진다. 실행 연산자 관점은 그 착각을 깨고, 각 단계가 어떤 입력을 소비하고 어떤 중간 상태를 만드는지 관측하게 해 주려고 등장한 설명 도구다.

독자는 SELECT가 한 번에 결과표로 변하는 것이 아니라, scan이 row 후보를 만들고 filter가 줄이고 projection이 열을 고르고 sort/aggregate가 stream을 재배열하거나 요약한다는 실행기 관점을 설명할 수 있어야 한다. 이 절은 정의 목록이 아니라 작은 입력이 인덱스나 실행기 내부에서 어떤 순서로 움직이는지 따라가는 설명이다. 읽은 뒤에는 문서의 문장을 외우는 것이 아니라, 같은 종류의 쿼리를 보았을 때 어떤 구조가 먼저 일을 시작하고 어떤 관측값을 봐야 하는지 말할 수 있어야 한다.

로컬 seed `database/query.md`는 correlated subquery를 통해 외부 row와 내부 query의 관계를 설명한다. 이 DU는 그보다 더 아래층인 row stream operator를 먼저 고정한다. 기존 seed는 짧거나 특정 상황에 묶여 있으므로 그대로 복사하지 않는다. 대신 seed가 품고 있던 질문을 살리고, 공식 문서가 닫아 주는 사실과 실행 trace가 보여 주는 관측 경로를 합쳐 새 본문으로 확장한다.

공식 근거는 다음 경계 안에서 사용한다.

- PostgreSQL 18 공식 문서의 Using EXPLAIN은 실행 계획이 plan node tree이고, 아래쪽 scan node가 raw row를 만들고 위쪽 node가 join, aggregate, sort 같은 연산을 수행한다고 설명한다.

첫 벽돌은 `orders` 5행에서 실패 주문을 날짜순으로 묶는 작은 SELECT다. 이 작은 장면을 붙잡아야 추상 용어가 실제 비용으로 바뀐다. 데이터베이스의 성능 설명은 보통 `빠르다`, `느리다`, `인덱스를 탄다` 같은 말로 시작하지만, 실제 진단은 어떤 row나 key가 어떤 구조를 거쳐 다음 소비자에게 전달되는지 확인하는 일이다.

```sql
SELECT customer_id, count(*) AS failed_count
FROM orders
WHERE status = 'FAILED'
GROUP BY customer_id
ORDER BY failed_count DESC;
```

```text
input rows
r1 (c1, FAILED, 10)
r2 (c1, PAID,   20)
r3 (c2, FAILED, 30)
r4 (c1, FAILED, 40)
r5 (c3, PAID,   50)

scan       -> r1 r2 r3 r4 r5
filter     -> r1 r3 r4
projection -> (c1) (c2) (c1)
aggregate  -> (c1,2) (c2,1)
sort       -> (c1,2) (c2,1)
output     -> customer_id, failed_count
```


PostgreSQL EXPLAIN 문서는 plan이 node tree라고 설명한다. 아래쪽 scan node는 테이블에서 raw row를 만들고, 위쪽 node는 join, aggregate, sort 같은 작업을 한다. 이 구조를 알아야 `Seq Scan` 한 줄과 `Sort` 한 줄을 비용상 같은 종류로 보지 않는다.

scan은 읽기 시작점이다. sequential scan은 테이블을 넓게 읽고, index scan은 인덱스가 좁힌 순서대로 읽고, bitmap scan은 후보 row 위치를 모은 뒤 heap을 방문한다. 어떤 scan이든 다음 operator에게 row stream을 넘긴다.

filter는 WHERE 조건을 통과하지 못한 row를 버린다. 중요한 함정은 모든 조건이 scan 단계에서 같은 비용으로 처리되는 것이 아니라는 점이다. Index Cond로 내려간 조건과 Filter로 남은 조건은 읽은 row 수와 버린 row 수가 다르다.

projection은 최종 열만 예쁘게 잘라내는 일이 아니라, 위쪽 operator가 필요한 열 폭(width)을 줄이는 역할도 한다. 불필요하게 큰 JSON/TEXT 열을 끌고 올라가면 sort와 hash aggregate의 memory pressure가 달라진다.

sort와 aggregate는 stream을 소비하는 방식이 다르다. sort는 전체 또는 제한된 top-N 후보를 순서대로 재배열하고, aggregate는 group key별 상태를 누적한다. HashAggregate와 GroupAggregate는 입력 정렬 상태와 메모리 조건에 따라 장단점이 갈린다.

실행 연산자 흐름에서는 row stream이 어디서 줄고 어디서 재배열되는지를 계속 추적해야 한다. scan이 많이 읽고 filter가 뒤에서 버리는 plan과, scan 단계에서 후보가 작아지는 plan은 같은 WHERE 문장을 갖고 있어도 비용이 다르다. projection은 들고 올라가는 row 폭을 바꾸고, sort와 aggregate는 그 폭과 row 수를 메모리 압력으로 바꾼다.

### 손으로 다시 읽는 예제

1. Filter pushdown 감각

    ```sql
    WHERE status='FAILED' AND amount > 10000
    ```

    status가 인덱스 조건이면 scan 후보가 줄고, amount가 Filter로 남으면 읽은 뒤 버릴 수 있다.
1. Projection 폭

    ```sql
    SELECT * 대신 SELECT id, customer_id
    ```

    위쪽 sort/hash가 들고 다니는 tuple 폭이 줄어 memory와 spill 가능성이 줄어든다.
1. Aggregate 선택

    ```sql
    GROUP BY customer_id
    ```

    정렬된 입력이면 GroupAggregate가 자연스럽고, 넓은 입력이면 HashAggregate가 빠를 수 있지만 memory 초과시 spill이 생긴다.

위 예제들은 SELECT 절들이 실행기 안에서 서로 다른 압력을 만든다는 점을 보여 준다. filter pushdown은 읽는 row 수를, projection 폭은 위쪽 node가 들고 다니는 byte 수를, aggregate 선택은 누적 상태와 memory 사용을 바꾼다. 같은 결과표라도 중간 stream은 완전히 다를 수 있다.

### 실무에서 먼저 의심해야 하는 함정

- SELECT 문을 한 번에 결과표로 변환한다고 생각하면 scan, filter, sort, aggregate 각각의 비용을 놓친다.
- WHERE 조건이 있으니 적게 읽을 것이라고 믿으면 안 된다. 조건이 인덱스 접근 조건인지, 읽은 뒤 필터인지, join 뒤 필터인지가 다르다.
- ORDER BY와 GROUP BY를 문법 절로만 보면 둘 다 비슷해 보이지만, 실행기에서는 재배열과 상태 누적이라는 전혀 다른 압력을 만든다.

실행 연산자 함정은 WHERE, SELECT, GROUP BY, ORDER BY를 문법 순서로만 볼 때 생긴다. 실제 비용은 scan 후보 수, filter에서 버린 row, projection 폭, sort spill, aggregate batch로 나타난다. 느린 쿼리에서 이 중 어느 단계가 커졌는지 분리하지 않으면 튜닝이 감으로 흐른다.

### 관측과 검증 경로

- `EXPLAIN (ANALYZE, BUFFERS)`에서 plan node tree의 indentation을 따라 아래에서 위로 읽는다. PASS는 각 node의 실제 rows가 예상한 stream 축소/확대와 맞는 것이다.
- Sort Method, Memory, Disk, HashAgg Batches 같은 부가 정보를 본다. PASS는 메모리 안에서 끝나고, FAIL은 disk spill이나 batches 증가가 나타나는 것이다.
- 쿼리를 `SELECT count(*)`, `SELECT key only`, `SELECT *`로 나눠 width와 buffer 사용이 어떻게 달라지는지 비교하면 projection의 의미를 관측할 수 있다.

실행 연산자 검증은 plan tree를 아래에서 위로 읽으며 row 수와 width, memory 정보를 맞추는 일이다. PASS는 scan 후보가 filter에서 얼마나 줄었는지, projection이 tuple 폭을 줄였는지, sort나 hash aggregate가 메모리 안에서 끝났는지 설명할 수 있는 상태다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 1

1-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 1번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 2

2-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 2번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 3

3-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 3번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 4

4-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 4번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 5

5-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 5번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 6

6-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 6번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 7

7-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 7번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### scan, filter, projection, sort, aggregate 현장 판독 노트 8

8-1. `Seq Scan` 상황에서는 먼저 전체 스캔을(를) 하나의 관측 단위로 분리한다. 전체를 읽는 것이 항상 나쁜 것은 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 작은 테이블이나 높은 selectivity에서는 순차 읽기가 더 단순하고 빠를 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-2. `Index Cond vs Filter` 상황에서는 먼저 조건 분리을(를) 하나의 관측 단위로 분리한다. Index Cond는 접근 범위를 줄이고 Filter는 읽은 뒤 버린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 둘을 같은 WHERE 조건으로 뭉뚱그리면 rows removed를 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-3. `Projection width` 상황에서는 먼저 열 폭을(를) 하나의 관측 단위로 분리한다. 위쪽 node가 들고 다니는 tuple 폭이 비용에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. SELECT *는 관측하기 전까지 편하지만 sort/hash를 무겁게 할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-4. `Sort spill` 상황에서는 먼저 정렬을(를) 하나의 관측 단위로 분리한다. 정렬 입력이 메모리를 넘으면 디스크를 쓴다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 시간이 튀면 sort method와 temp file을 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-5. `HashAggregate batches` 상황에서는 먼저 집계을(를) 하나의 관측 단위로 분리한다. 그룹 상태가 메모리에 안 들어가면 batch가 나뉜다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 latency보다 p95/p99에서 더 드러날 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-6. `LIMIT stop` 상황에서는 먼저 상위 node 중단을(를) 하나의 관측 단위로 분리한다. 부모 node가 필요한 만큼만 읽고 멈출 수 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 모든 node가 항상 전체를 끝까지 읽는 것은 아니다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 scan, filter, projection, sort, aggregate에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 8번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, scan, filter, projection, sort, aggregate을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### 오해 수리와 전이 질문

이 절에서 바로잡을 오해는 SELECT 문이 결과표를 한 번에 만든다는 생각이다. 실행기는 row stream을 만들고 줄이고 얇게 만들고 재배열하거나 요약한다. 어느 단계에서 row와 byte가 커지는지 봐야 성능 문제를 실제 위치로 되돌릴 수 있다.

마지막으로 scan, filter, projection, sort, aggregate을(를) 자기 말로 다시 설명하려면 세 문장을 직접 써 본다. 첫 문장은 이 구조가 해결하는 문제, 둘째 문장은 작은 입력이 내부에서 바뀌는 순서, 셋째 문장은 운영에서 깨지는 조건이다. 세 문장 중 하나라도 공식 문서나 trace로 되돌아가지 못하면 아직 외운 말이고, 세 문장이 모두 trace와 검증 명령으로 이어지면 다음 DU의 실행 계획을 읽을 준비가 된 것이다.


## nested loop/hash/merge join

등장 배경은 관계 모델의 join 의미와 물리 장치의 비용 사이에 큰 간격이 있다는 데 있다. SQL은 두 relation을 조건에 맞게 붙이라고 말하지만, 실제 엔진은 그 조합을 전부 비교할지, 한쪽을 hash table로 만들지, 양쪽을 정렬해서 포인터를 밀지 결정해야 한다. 데이터가 작던 시절에는 반복 비교도 견딜 수 있었지만, 업무 테이블이 커지고 index와 memory hierarchy가 복잡해지면서 join algorithm 선택 자체가 성능의 핵심 판단이 되었다.

독자는 SQL의 INNER/LEFT JOIN 같은 논리 join과 nested loop/hash/merge join 같은 물리 실행 알고리즘을 분리하고, 같은 입력 데이터에서 세 알고리즘이 어떤 비용과 전제를 가지는지 손으로 계산할 수 있어야 한다. 이 절은 정의 목록이 아니라 작은 입력이 인덱스나 실행기 내부에서 어떤 순서로 움직이는지 따라가는 설명이다. 읽은 뒤에는 문서의 문장을 외우는 것이 아니라, 같은 종류의 쿼리를 보았을 때 어떤 구조가 먼저 일을 시작하고 어떤 관측값을 봐야 하는지 말할 수 있어야 한다.

로컬 seed `database/join.md`와 `database/mysql/explains/when_join.md`는 JOIN 조건과 planner 선택이 성능을 좌우한다는 감각을 준다. 이 DU는 그 감각을 세 물리 join 알고리즘으로 분해한다. 기존 seed는 짧거나 특정 상황에 묶여 있으므로 그대로 복사하지 않는다. 대신 seed가 품고 있던 질문을 살리고, 공식 문서가 닫아 주는 사실과 실행 trace가 보여 주는 관측 경로를 합쳐 새 본문으로 확장한다.

공식 근거는 다음 경계 안에서 사용한다.

- PostgreSQL 18 공식 문서의 Planner/Optimizer와 planner statistics 문서는 planner가 가능한 plan을 만들고 통계와 비용 추정으로 선택한다는 경계를 준다.
- PostgreSQL 18 공식 문서의 Using EXPLAIN은 실행 계획이 plan node tree이고, 아래쪽 scan node가 raw row를 만들고 위쪽 node가 join, aggregate, sort 같은 연산을 수행한다고 설명한다.

첫 벽돌은 주문 4건과 결제 3건을 `order_id`로 붙이는 작은 데이터다. 이 작은 장면을 붙잡아야 추상 용어가 실제 비용으로 바뀐다. 데이터베이스의 성능 설명은 보통 `빠르다`, `느리다`, `인덱스를 탄다` 같은 말로 시작하지만, 실제 진단은 어떤 row나 key가 어떤 구조를 거쳐 다음 소비자에게 전달되는지 확인하는 일이다.

```text
orders                  payments
o1 id=1 customer=c1     p1 order_id=1 amount=100
o2 id=2 customer=c2     p2 order_id=1 amount=50
o3 id=3 customer=c3     p3 order_id=3 amount=70
o4 id=4 customer=c4

nested loop
o1 -> scan payments p1,p2,p3 -> match p1,p2
o2 -> scan payments p1,p2,p3 -> no match
o3 -> scan payments p1,p2,p3 -> match p3
o4 -> scan payments p1,p2,p3 -> no match

hash join
build hash on payments.order_id: 1->[p1,p2], 3->[p3]
probe orders: o1=>p1,p2 / o2=>empty / o3=>p3 / o4=>empty

merge join, both sorted by id/order_id
orders pointer and payments pointer move forward together
```


논리 JOIN은 어떤 row 조합이 결과에 포함되는지 정한다. 물리 join 알고리즘은 그 조합을 어떤 순서와 자료구조로 찾을지 정한다. `LEFT JOIN`이라고 해서 nested loop라는 뜻이 아니고, `INNER JOIN`이라고 해서 hash join이라는 뜻도 아니다.

nested loop join은 바깥 row마다 안쪽 relation을 찾아본다. 안쪽에 적절한 index가 있고 바깥 row가 작으면 매우 강하다. 하지만 바깥 row가 많고 안쪽 lookup이 비싸면 반복 비용이 폭발한다.

hash join은 한쪽 입력으로 hash table을 만들고 다른 쪽을 probe한다. equality join에 강하고, 정렬이 필요 없지만 build side가 메모리에 들어가지 않으면 batch나 spill 비용이 생긴다.

merge join은 양쪽 입력이 join key 순서로 정렬되어 있을 때 포인터를 앞으로 밀며 합친다. 이미 index order나 sort order가 있으면 좋지만, 정렬이 새로 필요하면 그 비용까지 포함해야 한다.

planner는 cardinality 추정, index 존재, sort order, work memory, join condition 형태를 보고 후보 plan의 비용을 비교한다. 그래서 같은 SQL도 데이터 분포와 통계가 바뀌면 다른 join 알고리즘으로 바뀔 수 있다.

join plan을 읽을 때는 먼저 어떤 입력이 바깥처럼 반복을 만들고, 어떤 입력이 build side나 정렬된 stream이 되는지 본다. nested loop는 반복 횟수와 inner lookup 비용으로, hash join은 build 크기와 skew로, merge join은 정렬 전제와 포인터 이동으로 판단한다. SQL의 JOIN 키워드가 아니라 이 물리 흐름이 병목을 만든다.

### 손으로 다시 읽는 예제

1. Nested loop가 맞는 경우

    ```sql
    최근 주문 10건 각각의 결제 행을 PK/FK index로 찾는다.
    ```

    outer가 작고 inner index lookup이 빠르다.
1. Hash join이 맞는 경우

    ```sql
    하루 주문 100만 건과 결제 100만 건을 order_id equality로 붙인다.
    ```

    정렬보다 hash build/probe가 유리할 수 있다.
1. Merge join이 맞는 경우

    ```sql
    두 입력이 이미 order_id 순서이고 큰 범위를 모두 붙인다.
    ```

    정렬 비용 없이 순차 포인터 이동으로 처리한다.

위 예제들은 join 알고리즘을 선택하는 기준이 SQL 문법 이름이 아니라 입력 크기와 준비 상태라는 점을 보여 준다. 작은 outer와 index가 있으면 nested loop가 강하고, 큰 equality join은 hash build/probe가 유리할 수 있으며, 이미 정렬된 큰 입력은 merge join이 자연스러울 수 있다.

### 실무에서 먼저 의심해야 하는 함정

- join 종류를 SQL 문법의 JOIN 종류와 혼동하면 튜닝 방향이 틀린다. LEFT JOIN을 INNER JOIN으로 바꿀 수 있는지와 nested loop를 hash join으로 바꿀 수 있는지는 다른 문제다.
- MySQL EXPLAIN에서 조인 순서를 보지 않고 key만 보면, 어떤 table이 outer처럼 먼저 줄어드는지 놓친다.
- hash join을 무조건 빠른 알고리즘으로 보면 memory spill과 skew 문제를 놓친다. 특정 key에 row가 몰리면 hash bucket이 커지고 probe 비용이 커진다.

join 함정은 논리 join과 물리 join을 섞을 때 생긴다. LEFT JOIN을 썼다는 사실만으로 nested loop가 된 것도 아니고, hash join이 보인다고 항상 빠른 것도 아니다. outer rows, inner lookup, hash memory, sort 전제, skew를 같이 보아야 한다.

### 관측과 검증 경로

- PostgreSQL에서는 `SET enable_hashjoin`, `enable_mergejoin`, `enable_nestloop`을 실험적으로 조절해 같은 SQL의 후보 plan 차이를 관측할 수 있다. 학습용 실험이지 운영 튜닝 기본값은 아니다.
- `EXPLAIN (ANALYZE, BUFFERS)`에서 Hash, Sort, Nested Loop node의 actual rows와 loops를 읽는다. PASS는 nested loop의 loops가 작은지, hash build rows가 예상과 맞는지, sort가 spill하지 않는지 확인하는 것이다.
- MySQL seed의 JSON EXPLAIN처럼 table 순서, type, key, rows를 함께 보아야 한다. PASS는 const/ref로 먼저 줄어든 table이 실제 join 후보를 작게 만든다는 설명과 plan이 맞는 것이다.

join 검증은 실제 알고리즘 이름보다 반복 횟수와 준비 비용을 본다. Nested Loop는 inner loops와 index lookup을, Hash Join은 build rows와 batch/spill을, Merge Join은 sort 유무와 정렬된 입력을 확인한다. PASS는 chosen join이 cardinality와 input order에 맞는 이유를 설명하는 것이다.

### nested loop/hash/merge join 현장 판독 노트 1

1-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 1번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 2

2-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 2번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 3

3-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 3번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 4

4-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 4번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 5

5-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 5번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 6

6-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 6번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 7

7-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 7번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### nested loop/hash/merge join 현장 판독 노트 8

8-1. `outer 작음` 상황에서는 먼저 작은 outer을(를) 하나의 관측 단위로 분리한다. nested loop는 바깥 row가 작으면 강하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. outer 추정이 틀리면 반복 폭발이 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-2. `inner index` 상황에서는 먼저 inner 탐색을(를) 하나의 관측 단위로 분리한다. nested loop의 성패는 안쪽 lookup 비용이 좌우한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 안쪽이 매번 seq scan이면 작은 착각이 큰 장애가 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-3. `hash build side` 상황에서는 먼저 hash build을(를) 하나의 관측 단위로 분리한다. 작은 쪽을 build하면 memory 압력이 줄어든다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. skew가 있으면 평균 row 수만으로 부족하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-4. `merge order` 상황에서는 먼저 정렬 입력을(를) 하나의 관측 단위로 분리한다. 이미 정렬된 입력은 merge join 비용을 낮춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 정렬을 새로 만들면 sort cost도 포함해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-5. `join filter` 상황에서는 먼저 join filter을(를) 하나의 관측 단위로 분리한다. join condition과 filter condition 위치를 분리해 읽는다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. join 후 버리는 row가 많으면 join order를 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-6. `SQL join vs physical join` 상황에서는 먼저 논리/물리 분리을(를) 하나의 관측 단위로 분리한다. LEFT/INNER는 결과 의미이고 nested/hash/merge는 실행 방식이다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 두 층을 섞으면 튜닝 대화가 어긋난다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 nested loop/hash/merge join에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 8번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, nested loop/hash/merge join을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### 오해 수리와 전이 질문

이 절에서 바로잡을 오해는 SQL JOIN 종류와 물리 join 알고리즘이 같은 층이라는 생각이다. INNER/LEFT는 결과 의미이고 nested loop/hash/merge는 실행 방식이다. 튜닝은 두 층을 분리한 뒤 cardinality와 입력 준비 상태를 보는 일에서 시작한다.

마지막으로 nested loop/hash/merge join을(를) 자기 말로 다시 설명하려면 세 문장을 직접 써 본다. 첫 문장은 이 구조가 해결하는 문제, 둘째 문장은 작은 입력이 내부에서 바뀌는 순서, 셋째 문장은 운영에서 깨지는 조건이다. 세 문장 중 하나라도 공식 문서나 trace로 되돌아가지 못하면 아직 외운 말이고, 세 문장이 모두 trace와 검증 명령으로 이어지면 다음 DU의 실행 계획을 읽을 준비가 된 것이다.


## subquery, CTE, window, pagination

등장 배경은 SQL이 단순한 한 단계 filter만으로 업무 질문을 표현하기 어렵다는 데 있다. 업무 쿼리는 “각 부서 평균보다 큰 직원”, “고객별 최신 주문”, “이전 행과 비교한 변화”, “다음 페이지”처럼 중간 결과와 순서를 계속 필요로 한다. subquery, CTE, window function, pagination은 이 중간 사고를 SQL 안에 표현하기 위해 생겼지만, 표현이 같다고 실행 경계가 같은 것은 아니다. 그래서 materialization, row numbering, 정렬 안정성, 앞 row를 버리는 비용을 분리해 읽어야 한다.

독자는 subquery, CTE, window function, pagination이 모두 중간 결과를 다루지만, materialize 여부와 row numbering, 정렬 안정성, 반복 실행 비용이 다르다는 점을 설명할 수 있어야 한다. 이 절은 정의 목록이 아니라 작은 입력이 인덱스나 실행기 내부에서 어떤 순서로 움직이는지 따라가는 설명이다. 읽은 뒤에는 문서의 문장을 외우는 것이 아니라, 같은 종류의 쿼리를 보았을 때 어떤 구조가 먼저 일을 시작하고 어떤 관측값을 봐야 하는지 말할 수 있어야 한다.

로컬 seed `database/elasticsearch/paginate.md`는 search_after와 PIT로 검색 pagination의 일관성을 다룬다. 이 DU는 RDBMS OFFSET/keyset pagination과 window/CTE의 중간 결과 경계를 비교한다. 기존 seed는 짧거나 특정 상황에 묶여 있으므로 그대로 복사하지 않는다. 대신 seed가 품고 있던 질문을 살리고, 공식 문서가 닫아 주는 사실과 실행 trace가 보여 주는 관측 경로를 합쳐 새 본문으로 확장한다.

공식 근거는 다음 경계 안에서 사용한다.

- PostgreSQL 18 공식 문서의 WITH Queries는 CTE가 보조 statement처럼 쓰이고 materialization/NOT MATERIALIZED 선택이 plan에 영향을 줄 수 있음을 설명한다.
- PostgreSQL 18 공식 tutorial의 Window Functions는 window 함수가 현재 row와 관련된 행 집합 위에서 계산된다는 기준 근거를 준다.

첫 벽돌은 고객별 최근 주문 순번을 붙이고, 다음 페이지를 가져오는 목록 쿼리다. 이 작은 장면을 붙잡아야 추상 용어가 실제 비용으로 바뀐다. 데이터베이스의 성능 설명은 보통 `빠르다`, `느리다`, `인덱스를 탄다` 같은 말로 시작하지만, 실제 진단은 어떤 row나 key가 어떤 구조를 거쳐 다음 소비자에게 전달되는지 확인하는 일이다.

```sql
SELECT id, customer_id, ordered_at,
       row_number() OVER (PARTITION BY customer_id ORDER BY ordered_at DESC, id DESC) AS rn
FROM orders
WHERE ordered_at >= DATE '2026-05-01';
```

```text
partition customer=c1 ordered by ordered_at desc, id desc
row o9  2026-05-05  -> rn=1
row o7  2026-05-03  -> rn=2
row o1  2026-05-01  -> rn=3

OFFSET pagination page 2 size 2
sorted stream: [o9, o7, o1, o8, o5]
OFFSET 2 discards o9,o7 then returns o1,o8

keyset pagination after (2026-05-03, id=7)
predicate: (ordered_at,id) < (2026-05-03,7)
index can seek near o1 and continue
```


subquery는 쿼리 안의 쿼리다. 상관 서브쿼리는 외부 row 값을 참조하므로 외부 row마다 다시 평가될 수 있다. optimizer가 decorrelation하거나 cache할 수는 있지만, SQL을 읽는 사람은 먼저 의존 관계를 분리해야 한다.

CTE는 읽기 좋은 이름 붙은 중간 query로 보이지만, 실행에서는 inline될 수도 있고 materialize될 수도 있다. PostgreSQL 문서는 `MATERIALIZED`, `NOT MATERIALIZED` 선택이 있음을 보여 준다. 따라서 CTE를 쓰면 항상 임시 테이블이 생긴다는 설명도, 항상 inline된다는 설명도 틀리다.

window function은 GROUP BY처럼 row를 줄이지 않고, 현재 row 주변의 window frame이나 partition을 보고 값을 붙인다. 그래서 ranking, running total, lag/lead 같은 계산을 할 수 있지만, 필요한 정렬과 partition memory 비용이 생긴다.

OFFSET pagination은 앞 page를 버리는 방식이다. page 1000을 보려면 앞의 많은 row를 정렬하거나 읽고 버릴 수 있다. keyset pagination은 마지막으로 본 key를 다음 predicate로 바꾸어 seek 가능한 조건을 만든다.

Elasticsearch의 search_after도 keyset과 닮았다. 하지만 검색 엔진은 refresh, PIT, shard, sort value, tie breaker라는 별도 경계를 가진다. RDBMS keyset pagination과 같은 이름으로 합치면 일관성 문제를 놓친다.

subquery, CTE, window, pagination은 모두 중간 결과를 다루지만 중간 결과의 성격이 다르다. 상관 서브쿼리는 외부 row 의존성을 만들고, CTE는 이름 붙은 query 경계를 만들 수 있으며, window는 row를 줄이지 않고 계산 열을 붙이고, pagination은 정렬된 stream의 일부를 소비한다. 이 차이를 놓치면 materialize 비용과 OFFSET 비용을 같은 종류로 착각한다.

### 손으로 다시 읽는 예제

1. 상관 서브쿼리

    ```sql
    WHERE salary > (SELECT avg(salary) FROM employees e2 WHERE e2.dept = e1.dept)
    ```

    외부 row의 dept가 내부 query를 바꾼다.
1. CTE 재사용

    ```sql
    WITH recent AS (...) SELECT ... FROM recent r1 JOIN recent r2 ...
    ```

    읽기 좋지만 materialization 여부가 비용을 좌우한다.
1. window rank

    ```sql
    row_number() over (partition by customer_id order by ordered_at desc)
    ```

    row를 줄이지 않고 순번 열을 붙인다.

위 예제들은 중간 결과를 표현하는 SQL 기능이 실행 비용을 숨길 수 있음을 보여 준다. 상관 서브쿼리는 외부 row 의존성을 만들고, CTE는 materialize 여부를 봐야 하며, window rank는 row 수를 유지한 채 계산 열을 붙인다. pagination은 그 결과 stream을 어디서부터 소비할지의 문제다.

### 실무에서 먼저 의심해야 하는 함정

- OFFSET pagination이 항상 싸다고 믿으면 운영에서 page 뒤쪽 조회가 갑자기 느려진다. 앞 row를 버리는 비용은 사용자에게 보이지 않지만 DB에는 실제 비용이다.
- CTE를 성능 최적화 장치로만 쓰면 planner가 predicate를 밀어 넣지 못하거나 materialize 경계가 생겨 느려질 수 있다.
- window function을 GROUP BY의 다른 표기라고 생각하면 결과 row 수가 유지되는 성질을 놓쳐 중복 row나 잘못된 합계를 만든다.

중간 결과 함정은 읽기 좋은 SQL을 곧 빠른 SQL로 착각할 때 생긴다. CTE는 materialize 경계가 될 수 있고, window는 row 수를 줄이지 않으며, OFFSET은 앞 row를 버리는 비용을 숨긴다. 검색의 search_after도 refresh와 PIT 경계가 있어 RDBMS keyset과 그대로 같지 않다.

### 관측과 검증 경로

- PostgreSQL에서 CTE에 `MATERIALIZED`와 `NOT MATERIALIZED`를 바꿔 `EXPLAIN` plan을 비교한다. PASS는 scan 횟수, filter pushdown, materialize node 차이를 설명할 수 있는 것이다.
- OFFSET과 keyset pagination을 같은 인덱스에서 `EXPLAIN (ANALYZE, BUFFERS)`로 비교한다. PASS는 뒤쪽 page에서 OFFSET의 읽고 버리는 row가 늘어난다는 점이 관측되는 것이다.
- Elasticsearch seed처럼 search_after를 쓸 때는 sort value와 tie breaker, PIT 사용 여부를 확인한다. PASS는 refresh 이후 순서 흔들림을 별도 경계로 설명하는 것이다.

중간 결과 검증은 materialize 여부, window 정렬, pagination 소비 위치를 plan에서 확인한다. PostgreSQL에서는 CTE의 MATERIALIZED/NOT MATERIALIZED 차이와 WindowAgg, Sort, Limit node를 비교한다. PASS는 OFFSET이 읽고 버린 row와 keyset이 seek한 시작점의 차이를 관측하는 것이다.

### subquery, CTE, window, pagination 현장 판독 노트 1

1-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 1번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 2

2-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 2번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 3

3-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 3번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 4

4-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 4번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 5

5-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 5번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 6

6-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 6번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 7

7-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 7번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### subquery, CTE, window, pagination 현장 판독 노트 8

8-1. `correlated subquery` 상황에서는 먼저 외부 참조을(를) 하나의 관측 단위로 분리한다. 외부 row 값이 내부 query를 바꾼다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. row마다 반복될 수 있는지 먼저 의심한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-2. `CTE readability` 상황에서는 먼저 중간 이름을(를) 하나의 관측 단위로 분리한다. CTE는 이름을 주지만 반드시 성능 장치는 아니다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. materialize 경계가 생기면 predicate pushdown이 막힐 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-3. `window partition` 상황에서는 먼저 순번 계산을(를) 하나의 관측 단위로 분리한다. partition별로 순서를 만들고 현재 row에 값을 붙인다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. GROUP BY처럼 row가 줄지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-4. `OFFSET cost` 상황에서는 먼저 뒤쪽 페이지을(를) 하나의 관측 단위로 분리한다. 앞 page를 버리는 비용이 page 번호와 함께 증가한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 뒤쪽 page 장애는 데이터가 커진 뒤에야 보인다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-5. `keyset tie breaker` 상황에서는 먼저 동률 처리을(를) 하나의 관측 단위로 분리한다. 동일 timestamp가 있으면 id 같은 안정 key가 필요하다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. tie breaker 없이는 누락/중복 page가 생긴다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-6. `search_after PIT` 상황에서는 먼저 검색 페이지을(를) 하나의 관측 단위로 분리한다. 검색 pagination은 refresh와 shard order 경계가 있다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. RDBMS keyset과 같은 보장으로 말하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 subquery, CTE, window, pagination에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 8번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, subquery, CTE, window, pagination을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### 오해 수리와 전이 질문

이 절에서 바로잡을 오해는 중간 결과를 쓰면 DB가 알아서 싸게 처리한다는 생각이다. subquery의 외부 의존성, CTE materialization, window 정렬, OFFSET discard, search_after 일관성은 각각 다른 비용과 실패 모드를 가진다.

마지막으로 subquery, CTE, window, pagination을(를) 자기 말로 다시 설명하려면 세 문장을 직접 써 본다. 첫 문장은 이 구조가 해결하는 문제, 둘째 문장은 작은 입력이 내부에서 바뀌는 순서, 셋째 문장은 운영에서 깨지는 조건이다. 세 문장 중 하나라도 공식 문서나 trace로 되돌아가지 못하면 아직 외운 말이고, 세 문장이 모두 trace와 검증 명령으로 이어지면 다음 DU의 실행 계획을 읽을 준비가 된 것이다.
