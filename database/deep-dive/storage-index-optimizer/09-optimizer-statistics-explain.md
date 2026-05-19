# 옵티마이저, 통계, EXPLAIN을 운영 진단 언어로 읽기


이 문서는 인덱스와 실행 계획을 실제 row 이동과 관측값으로 읽기 위한 장문 학습 문서다. 각 절은 정의 암기가 아니라 작은 trace를 따라가며 구조, 비용, 장애 신호를 함께 설명한다.


## 문서 안에서 먼저 잡을 목차


- [statistics, cardinality, selectivity, cost](#statistics-cardinality-selectivity-cost)
- [EXPLAIN/EXPLAIN ANALYZE 읽기](#explainexplain-analyze-읽기)
- [slow query diagnosis and plan regression](#slow-query-diagnosis-and-plan-regression)




## statistics, cardinality, selectivity, cost

등장 배경은 optimizer가 실제 데이터를 모두 실행해 본 뒤 plan을 고를 수 없다는 제약이다. 쿼리를 실행하기 전에 이미 어떤 plan을 쓸지 정해야 하므로, DBMS는 샘플과 통계로 row 수를 추정하고 그 추정값을 비용 모델에 넣는다. 이 구조 덕분에 매번 모든 후보 plan을 실제 실행하지 않아도 되지만, 통계가 낡거나 분포가 치우치면 planner는 과거 세계나 평균 세계를 보고 잘못된 선택을 할 수 있다.

독자는 optimizer가 실제 row 수를 미리 아는 것이 아니라 통계에서 cardinality와 selectivity를 추정하고, 그 추정으로 비용을 계산해 plan을 고른다는 점을 설명할 수 있어야 한다. 이 절은 정의 목록이 아니라 작은 입력이 인덱스나 실행기 내부에서 어떤 순서로 움직이는지 따라가는 설명이다. 읽은 뒤에는 문서의 문장을 외우는 것이 아니라, 같은 종류의 쿼리를 보았을 때 어떤 구조가 먼저 일을 시작하고 어떤 관측값을 봐야 하는지 말할 수 있어야 한다.

로컬 seed는 `database/database-deep-study-plan.md` 자체다. 이 DU는 DB deep-dive 전체에서 EXPLAIN을 읽기 전에 통계와 추정이라는 공통 언어를 고정한다. 기존 seed는 짧거나 특정 상황에 묶여 있으므로 그대로 복사하지 않는다. 대신 seed가 품고 있던 질문을 살리고, 공식 문서가 닫아 주는 사실과 실행 trace가 보여 주는 관측 경로를 합쳐 새 본문으로 확장한다.

공식 근거는 다음 경계 안에서 사용한다.

- PostgreSQL 18 공식 문서의 Planner/Optimizer와 planner statistics 문서는 planner가 가능한 plan을 만들고 통계와 비용 추정으로 선택한다는 경계를 준다.
- MySQL 8.4 공식 문서의 Optimizer Statistics는 table/index 통계와 histogram이 optimizer 판단에 들어간다는 점을 설명하는 기준 근거다.

첫 벽돌은 1,000,000건 주문 중 status가 대부분 PAID이고 FAILED는 1,000건뿐인 분포다. 이 작은 장면을 붙잡아야 추상 용어가 실제 비용으로 바뀐다. 데이터베이스의 성능 설명은 보통 `빠르다`, `느리다`, `인덱스를 탄다` 같은 말로 시작하지만, 실제 진단은 어떤 row나 key가 어떤 구조를 거쳐 다음 소비자에게 전달되는지 확인하는 일이다.

```text
table orders: 1,000,000 rows
status distribution from sampled statistics
PAID     970,000  selectivity about 0.970
FAILED     1,000  selectivity about 0.001
CANCELLED 29,000  selectivity about 0.029

query: WHERE status='FAILED' AND merchant_id='m-7'

rough planner flow
statistics -> estimate status selectivity -> combine with merchant_id estimate
   -> estimated rows maybe 10
   -> compare Seq Scan cost vs Index Scan cost
   -> choose plan
```


cardinality는 어떤 단계가 몇 row를 낼지에 대한 수량 감각이다. selectivity는 조건이 전체 중 얼마를 남기는지에 대한 비율 감각이다. optimizer는 이 둘을 통계에서 추정한다.

PostgreSQL의 planner statistics 문서는 ANALYZE가 표본을 기반으로 통계를 만들고, `pg_stats` 같은 view로 null fraction, distinct count, most common values, histogram bounds를 볼 수 있게 한다. 표본 기반이라는 말은 추정이 정확한 계측값이 아니라는 뜻이다.

MySQL optimizer statistics도 table/index cardinality와 histogram 같은 통계 정보를 사용한다. InnoDB persistent statistics와 histogram은 plan 선택의 입력이지만, 데이터 분포가 바뀌면 최신성이 중요해진다.

cost는 시간 그 자체가 아니라 planner가 plan을 비교하기 위해 쓰는 단위다. PostgreSQL EXPLAIN 문서는 cost가 arbitrary unit이고 전통적으로 page fetch 비용을 기준으로 삼는다고 설명한다. 그래서 cost 100이 100ms라는 뜻은 아니다.

잘못된 cardinality 추정은 join order, join algorithm, index 선택, sort/aggregate memory 예측을 모두 흔든다. 통계 문제가 느린 쿼리의 시작점인 이유가 여기에 있다.

통계와 비용 설명에서는 추정이 실제를 대신한다는 전제를 놓치면 안 된다. planner는 표본과 histogram, 고빈도 값, distinct 추정으로 row 수를 예상하고 그 예상값을 비용 비교에 넣는다. 따라서 튜닝의 첫 질문은 “왜 이 plan인가”보다 “이 plan을 고르게 만든 row 수 추정이 실제와 얼마나 맞는가”가 되어야 한다.

### 손으로 다시 읽는 예제

1. 균등 분포 착각

    ```sql
    WHERE status='FAILED'
    ```

    status 값이 세 개라고 1/3씩 있다고 가정하면 실제 0.1% 분포를 놓친다.
1. 상관관계 누락

    ```sql
    WHERE country='KR' AND currency='KRW'
    ```

    두 조건이 독립이 아니면 단순 곱셈 추정이 틀릴 수 있다.
1. stale statistics

    ```sql
    대량 적재 직후 ANALYZE 전 쿼리
    ```

    통계가 과거 분포를 보고 있어 새 데이터의 plan을 잘못 고를 수 있다.

위 예제들은 통계가 평균적인 분포 설명에 머물면 plan이 쉽게 빗나간다는 점을 보여 준다. 균등 분포 착각은 인기 값과 희귀 값을 섞고, 상관관계 누락은 조건 곱셈을 망가뜨리며, stale statistics는 이미 바뀐 테이블을 과거 모습으로 보게 한다.

### 실무에서 먼저 의심해야 하는 함정

- optimizer가 실제 row 수를 미리 안다고 믿으면 EXPLAIN의 rows와 실제 rows 차이를 장애 신호로 읽지 못한다.
- cost를 wall-clock time으로 읽으면 서로 다른 서버나 설정에서 숫자를 잘못 비교한다. cost는 같은 planner 설정 안에서 plan 후보를 비교하는 내부 점수에 가깝다.
- index를 추가했는데도 plan이 바뀌지 않을 때 통계 최신성, 분포 skew, 상관관계, selectivity를 보지 않고 optimizer를 탓하면 원인 추적이 막힌다.

통계 함정은 optimizer를 전지적인 실행자로 상상할 때 생긴다. planner는 실제 row 수를 모르고 추정한다. 따라서 통계가 낡거나 분포가 치우치면 index가 있어도 plan이 이상해질 수 있으며, cost 숫자를 ms처럼 읽으면 원인 판단이 틀어진다.

### 관측과 검증 경로

- PostgreSQL에서는 `ANALYZE`, `SELECT * FROM pg_stats WHERE tablename='orders'`, `EXPLAIN (ANALYZE, BUFFERS)`를 함께 본다. PASS는 estimated rows와 actual rows 차이를 조건별로 설명하는 것이다.
- MySQL에서는 `ANALYZE TABLE`, histogram 관련 명령, `EXPLAIN FORMAT=JSON`의 rows/filtered를 함께 본다. PASS는 통계 갱신 전후 plan 차이를 설명할 수 있는 것이다.
- 실패 신호는 estimated rows 10인데 actual rows 100000처럼 차이가 커지고, 그 차이 때문에 nested loop 반복이나 sort spill이 발생하는 것이다.

통계 검증은 추정값의 근거와 실제값의 차이를 나란히 놓는 일이다. PostgreSQL의 pg_stats와 EXPLAIN ANALYZE, MySQL의 ANALYZE TABLE과 EXPLAIN rows/filtered를 함께 본다. PASS는 통계 갱신 전후 또는 분포 변화 전후에 estimated rows가 왜 달라지는지 설명하는 것이다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 1

1-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 1번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 2

2-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 2번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 3

3-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 3번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 4

4-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 4번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 5

5-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 5번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 6

6-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 6번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 7

7-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 7번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### statistics, cardinality, selectivity, cost 현장 판독 노트 8

8-1. `MCV` 상황에서는 먼저 인기 값을(를) 하나의 관측 단위로 분리한다. most common values는 인기 값 skew를 planner에 알려 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인기 값과 희귀 값을 평균 하나로 보면 추정이 깨진다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-2. `histogram` 상황에서는 먼저 범위 분포을(를) 하나의 관측 단위로 분리한다. 범위 조건은 histogram boundary로 추정된다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 급격한 분포 변화는 boundary 사이에서 숨을 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-3. `distinct count` 상황에서는 먼저 고유값을(를) 하나의 관측 단위로 분리한다. 고유값 추정은 group/join 크기에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 실제 tenant 수가 바뀌면 join cardinality도 바뀐다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-4. `correlation` 상황에서는 먼저 물리 상관을(를) 하나의 관측 단위로 분리한다. 물리 순서와 값의 상관관계는 index scan 비용 감각에 영향을 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. append-only timestamp는 random key와 다르게 읽힌다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-5. `stale stats` 상황에서는 먼저 통계 최신성을(를) 하나의 관측 단위로 분리한다. 대량 변경 뒤 통계가 늦으면 과거 세계를 보고 plan을 고른다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. ANALYZE는 튜닝이 아니라 관측 갱신이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-6. `cost parameter` 상황에서는 먼저 비용 단위을(를) 하나의 관측 단위로 분리한다. cost 단위는 DB 설정의 비교 점수다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 숫자를 ms로 번역하면 안 된다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 statistics, cardinality, selectivity, cost에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 8번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, statistics, cardinality, selectivity, cost을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### 오해 수리와 전이 질문

이 절에서 바로잡을 오해는 optimizer가 실제 row 수를 미리 알고 최선의 plan을 고른다는 생각이다. planner는 통계에서 추정하고 비용을 비교한다. 그래서 추정이 틀리면 좋은 인덱스가 있어도 나쁜 plan이 합리적인 선택처럼 보일 수 있다.

마지막으로 statistics, cardinality, selectivity, cost을(를) 자기 말로 다시 설명하려면 세 문장을 직접 써 본다. 첫 문장은 이 구조가 해결하는 문제, 둘째 문장은 작은 입력이 내부에서 바뀌는 순서, 셋째 문장은 운영에서 깨지는 조건이다. 세 문장 중 하나라도 공식 문서나 trace로 되돌아가지 못하면 아직 외운 말이고, 세 문장이 모두 trace와 검증 명령으로 이어지면 다음 DU의 실행 계획을 읽을 준비가 된 것이다.


## EXPLAIN/EXPLAIN ANALYZE 읽기

등장 배경은 query plan이 눈에 보이지 않으면 튜닝이 감으로 흐른다는 데 있다. SQL 텍스트만 보면 조건과 join은 보이지만, 실제로 어느 table을 먼저 읽었는지, 몇 row를 예상했는지, 몇 번 반복했는지, 정렬이나 hash가 메모리를 넘었는지는 보이지 않는다. EXPLAIN은 이 보이지 않는 실행 계획을 사람이 읽을 수 있는 형태로 드러내기 위해 필요해졌고, EXPLAIN ANALYZE는 추정과 실제 실행의 차이를 비교하는 관측 장치가 된다.

독자는 EXPLAIN 한 줄의 type/key만 보고 성능을 판정하지 않고, plan tree, estimated rows, actual rows, loops, buffer, sort/hash 부가 정보를 함께 읽어 병목을 진단할 수 있어야 한다. 이 절은 정의 목록이 아니라 작은 입력이 인덱스나 실행기 내부에서 어떤 순서로 움직이는지 따라가는 설명이다. 읽은 뒤에는 문서의 문장을 외우는 것이 아니라, 같은 종류의 쿼리를 보았을 때 어떤 구조가 먼저 일을 시작하고 어떤 관측값을 봐야 하는지 말할 수 있어야 한다.

로컬 seed `database/mysql/explains/when_join.md`는 MySQL JSON EXPLAIN에서 const/ref/key/rows를 보여 준다. 이 DU는 그 예시를 PostgreSQL의 EXPLAIN ANALYZE 판독법과 연결한다. 기존 seed는 짧거나 특정 상황에 묶여 있으므로 그대로 복사하지 않는다. 대신 seed가 품고 있던 질문을 살리고, 공식 문서가 닫아 주는 사실과 실행 trace가 보여 주는 관측 경로를 합쳐 새 본문으로 확장한다.

공식 근거는 다음 경계 안에서 사용한다.

- PostgreSQL 18 공식 문서의 Using EXPLAIN은 실행 계획이 plan node tree이고, 아래쪽 scan node가 raw row를 만들고 위쪽 node가 join, aggregate, sort 같은 연산을 수행한다고 설명한다.
- MySQL 8.4 공식 문서의 EXPLAIN Output Format과 Slow Query Log는 MySQL에서 plan 행의 type/key/rows/Extra, 느린 쿼리 기록의 역할을 읽는 기준 근거다.

첫 벽돌은 같은 JOIN 쿼리를 실행 전 계획과 실행 후 계측으로 나눠 읽는 것이다. 이 작은 장면을 붙잡아야 추상 용어가 실제 비용으로 바뀐다. 데이터베이스의 성능 설명은 보통 `빠르다`, `느리다`, `인덱스를 탄다` 같은 말로 시작하지만, 실제 진단은 어떤 row나 key가 어떤 구조를 거쳐 다음 소비자에게 전달되는지 확인하는 일이다.

```text
PostgreSQL style simplified

Hash Join  (cost=35..80 rows=100) (actual time=2..30 rows=500 loops=1)
  Hash Cond: payments.order_id = orders.id
  -> Seq Scan on payments (cost=0..40 rows=1000) (actual rows=1000 loops=1)
  -> Hash
     -> Seq Scan on orders (cost=0..20 rows=100) (actual rows=500 loops=1)

read order for diagnosis
1. bottom scans: payments actual 1000, orders actual 500
2. estimate gap: orders estimated 100 but actual 500
3. join output: estimated 100 but actual 500
4. consequence: if this were nested loop, loops could multiply the mistake
```

```text
MySQL style simplified
table=This        type=const key=another_uniq_id rows=1
table=Another     type=ref   key=idx_This_id     rows=6

good sign: first table becomes const and second table probes by ref
missing question: are rows estimates close to actual runtime observation?
```


PostgreSQL의 EXPLAIN은 실행하지 않고 planner가 만든 계획을 보여 준다. EXPLAIN ANALYZE는 실제 실행까지 해서 actual time, actual rows, loops를 붙인다. 실제 DML이라면 부작용이 생길 수 있으므로 트랜잭션 rollback 같은 안전 경계를 잡아야 한다.

plan tree는 위에서 아래로 출력되지만, 데이터는 보통 아래 scan node에서 만들어져 위로 올라간다. 진단은 아래쪽 actual rows부터 읽고, 어디서 예상과 실제가 벌어졌는지 찾는 편이 안전하다.

estimated rows와 actual rows의 차이는 optimizer가 어떤 통계 가정을 했는지 보여 준다. 차이가 작으면 plan 선택의 근거가 비교적 안정적이고, 차이가 크면 join order나 알고리즘 선택이 흔들릴 수 있다.

loops는 node가 몇 번 반복 실행되었는지 보여 준다. nested loop 안쪽 index scan이 1ms라도 loops가 100000이면 총비용은 커진다. 한 줄의 node 시간만 보고 병목을 판단하면 반복 비용을 놓친다.

MySQL EXPLAIN의 type, key, rows, filtered, Extra는 접근 방식과 추정 후보 수를 보여 준다. `Using index`는 covering 가능성을, `Using filesort`는 별도 정렬을, `Using temporary`는 중간 임시 구조를 의심하게 한다.

EXPLAIN 판독에서는 출력 한 줄의 멋진 키워드보다 plan tree의 데이터 흐름이 중요하다. 아래 scan node에서 row가 만들어지고, 중간 node에서 반복되거나 정렬되고, 위 node에서 최종 결과가 나온다. estimated rows와 actual rows, loops, buffers를 함께 읽으면 plan이 느린 이유를 access method 이름이 아니라 증폭 지점으로 설명할 수 있다.

### 손으로 다시 읽는 예제

1. estimated vs actual

    ```sql
    rows=100, actual rows=50000
    ```

    통계나 조건 상관관계가 틀렸을 가능성을 먼저 본다.
1. loops 함정

    ```sql
    Index Scan actual time=0.02..0.03 rows=1 loops=200000
    ```

    한 번은 작지만 반복 총량이 병목이다.
1. MySQL key 함정

    ```sql
    key=idx_status rows=500000 Extra=Using filesort
    ```

    인덱스를 탔다고 빠른 것이 아니라 너무 많이 읽고 정렬할 수 있다.

위 예제들은 EXPLAIN 판독이 access type 암기가 아니라 추정과 실제의 차이를 찾는 일이라는 점을 보여 준다. estimated/actual 차이는 통계 문제를, loops는 반복 증폭을, MySQL key와 Extra의 조합은 인덱스 사용 뒤에 남은 정렬이나 임시 작업을 드러낸다.

### 실무에서 먼저 의심해야 하는 함정

- EXPLAIN 한 줄의 type/key만 보고 성능을 판정하면 실제 rows, loops, sort spill, buffer hit를 놓친다.
- EXPLAIN ANALYZE를 운영 쓰기 쿼리에 그대로 실행하면 실제 변경이 일어날 수 있다. 진단 도구도 action surface라는 점을 잊으면 사고가 난다.
- PostgreSQL cost와 MySQL rows를 같은 숫자 의미로 비교하면 안 된다. 각 DBMS의 EXPLAIN 출력은 자기 optimizer의 언어다.

EXPLAIN 함정은 한 줄의 key/type/node 이름을 성능 판정으로 착각할 때 생긴다. 실제 병목은 estimated/actual rows 차이, loops 증폭, buffer read, sort/hash spill, temporary 작업에 숨어 있다. 특히 EXPLAIN ANALYZE는 실제 실행이므로 쓰기 쿼리에서는 안전 경계도 진단의 일부다.

### 관측과 검증 경로

- PostgreSQL에서는 `EXPLAIN (ANALYZE, BUFFERS, VERBOSE)`를 읽고 estimated rows와 actual rows 비율이 10배 이상 벌어지는 첫 node를 표시한다. PASS는 그 node가 위쪽 비용을 어떻게 증폭하는지 설명하는 것이다.
- MySQL에서는 `EXPLAIN FORMAT=JSON`과 가능하면 `EXPLAIN ANALYZE` 계열 런타임 정보를 같이 본다. PASS는 access type, chosen key, rows examined, filesort/temporary 여부를 함께 설명하는 것이다.
- 로컬 seed의 두 JOIN 예시처럼 조건을 하나 추가해도 같은 key와 rows가 유지되는 경우, 왜 planner가 같은 plan을 선택했는지 const/ref 관계로 설명할 수 있어야 한다.

EXPLAIN 검증은 plan tree에서 첫 추정 오류 지점을 찾고 그 오류가 위쪽 node에서 어떻게 증폭되는지 설명하는 일이다. PostgreSQL은 actual rows, loops, buffers를 보고, MySQL은 type, key, rows, filtered, Extra를 함께 읽는다. PASS는 access method 이름이 아니라 병목의 위치와 이유가 닫히는 것이다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 1

1-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 1번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 2

2-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 2번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 3

3-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 3번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 4

4-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 4번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 5

5-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 5번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 6

6-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 6번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 7

7-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 7번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### EXPLAIN/EXPLAIN ANALYZE 읽기 현장 판독 노트 8

8-1. `plan tree` 상황에서는 먼저 트리 판독을(를) 하나의 관측 단위로 분리한다. 출력은 위에서 보이지만 데이터는 아래서 올라간다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 진단은 scan node에서 시작하는 편이 안정적이다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-2. `actual rows` 상황에서는 먼저 추정 비교을(를) 하나의 관측 단위로 분리한다. 실행 후 row 수는 추정 오류를 보여 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 한 node의 오류가 위쪽 join에서 증폭될 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-3. `loops` 상황에서는 먼저 반복 횟수을(를) 하나의 관측 단위로 분리한다. 짧은 inner scan도 반복되면 비싸다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. actual time 한 번만 보면 총량을 놓친다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-4. `buffers` 상황에서는 먼저 버퍼 관측을(를) 하나의 관측 단위로 분리한다. CPU 문제인지 I/O 문제인지 힌트를 준다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 읽은 block 수와 cache hit를 같이 본다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-5. `MySQL Extra` 상황에서는 먼저 추가 작업을(를) 하나의 관측 단위로 분리한다. Using filesort/temporary/index는 접근 후 추가 작업을 말한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. key가 잡혔다는 사실만으로 끝내지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-6. `safe analyze` 상황에서는 먼저 부작용 경계을(를) 하나의 관측 단위로 분리한다. 실제 실행 계측은 부작용 경계를 확인해야 한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쓰기 쿼리는 transaction rollback 등 안전 장치가 필요하다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 EXPLAIN/EXPLAIN ANALYZE 읽기에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 8번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### 오해 수리와 전이 질문

이 절에서 바로잡을 오해는 EXPLAIN의 key나 node 이름 하나가 성능 판정을 끝낸다는 생각이다. plan은 tree이고, 병목은 추정 오류와 반복, buffer, sort/hash 상태가 결합된 곳에서 생긴다. 한 줄 판독은 그 결합을 숨긴다.

마지막으로 EXPLAIN/EXPLAIN ANALYZE 읽기을(를) 자기 말로 다시 설명하려면 세 문장을 직접 써 본다. 첫 문장은 이 구조가 해결하는 문제, 둘째 문장은 작은 입력이 내부에서 바뀌는 순서, 셋째 문장은 운영에서 깨지는 조건이다. 세 문장 중 하나라도 공식 문서나 trace로 되돌아가지 못하면 아직 외운 말이고, 세 문장이 모두 trace와 검증 명령으로 이어지면 다음 DU의 실행 계획을 읽을 준비가 된 것이다.


## slow query diagnosis and plan regression

등장 배경은 느린 쿼리 장애가 SQL 한 줄의 문법 문제가 아니라 시간에 따라 바뀌는 운영 상태의 문제인 경우가 많다는 데 있다. 어제 빠르던 쿼리가 오늘 느려지는 이유는 데이터 증가, tenant skew, 통계 stale, lock wait, cache 상태, plan 변화, 배포 후 parameter 변화처럼 여러 층에서 생긴다. slow query log와 plan regression 추적은 이 시간 축을 복원하기 위해 필요하며, 인덱스 추가라는 빠른 처방 전에 무엇이 실제로 바뀌었는지 묻도록 만든다.

독자는 느린 쿼리를 곧바로 인덱스 추가 문제로 좁히지 않고, 느린 쿼리 로그, plan 변화, 통계 최신성, parameter 분포, cache 상태, 데이터 증가, plan regression을 순서대로 추적할 수 있어야 한다. 이 절은 정의 목록이 아니라 작은 입력이 인덱스나 실행기 내부에서 어떤 순서로 움직이는지 따라가는 설명이다. 읽은 뒤에는 문서의 문장을 외우는 것이 아니라, 같은 종류의 쿼리를 보았을 때 어떤 구조가 먼저 일을 시작하고 어떤 관측값을 봐야 하는지 말할 수 있어야 한다.

로컬 seed `database/mysql/explains/when_join.md`는 작은 JOIN plan 예시를 준다. 이 DU는 그 plan이 운영 데이터 분포와 시간이 지나며 어떻게 느린 쿼리 문제로 바뀌는지 추적한다. 기존 seed는 짧거나 특정 상황에 묶여 있으므로 그대로 복사하지 않는다. 대신 seed가 품고 있던 질문을 살리고, 공식 문서가 닫아 주는 사실과 실행 trace가 보여 주는 관측 경로를 합쳐 새 본문으로 확장한다.

공식 근거는 다음 경계 안에서 사용한다.

- PostgreSQL 18 공식 문서의 Using EXPLAIN은 실행 계획이 plan node tree이고, 아래쪽 scan node가 raw row를 만들고 위쪽 node가 join, aggregate, sort 같은 연산을 수행한다고 설명한다.
- MySQL 8.4 공식 문서의 EXPLAIN Output Format과 Slow Query Log는 MySQL에서 plan 행의 type/key/rows/Extra, 느린 쿼리 기록의 역할을 읽는 기준 근거다.

첫 벽돌은 어제 30ms였던 결제 목록 쿼리가 오늘 4초로 느려졌다는 장애 티켓이다. 이 작은 장면을 붙잡아야 추상 용어가 실제 비용으로 바뀐다. 데이터베이스의 성능 설명은 보통 `빠르다`, `느리다`, `인덱스를 탄다` 같은 말로 시작하지만, 실제 진단은 어떤 row나 key가 어떤 구조를 거쳐 다음 소비자에게 전달되는지 확인하는 일이다.

```text
incident timeline

09:00 deploy none
  -> baseline plan still healthy
10:00 merchant m-big imports 5,000,000 historical rows
  -> data distribution changes before SQL text changes
10:10 slow query log starts showing list query > 3s
  -> symptom points to the query family, not yet to the root cause
10:15 EXPLAIN still shows idx_merchant_status but rows estimate=120, actual rows=900000
  -> estimate/actual gap explains why the old plan no longer behaves like yesterday
10:20 ANALYZE updates statistics
  -> planner sees the new distribution
10:25 plan changes to index on (merchant_id,status,requested_at), latency drops to 80ms
  -> verification closes only after plan, rows, and latency move together

wrong shortcut -> add another single-column status index
better path -> compare old/new plan, estimate/actual gap, data distribution, freshness of stats
```


느린 쿼리 진단은 증상 수집에서 시작한다. MySQL slow query log는 오래 걸린 SQL, 검사한 row, lock time 같은 단서를 남길 수 있다. PostgreSQL에서는 EXPLAIN ANALYZE와 로그 설정, pg_stat_statements 같은 관측 표면을 함께 볼 수 있다.

plan regression은 SQL 문자가 같아도 plan이 나빠지거나, plan은 같아 보이지만 데이터 분포가 바뀌어 실제 비용이 커지는 상황이다. 원인은 통계 stale, parameter skew, index bloat, row width 증가, cache miss, 설정 변경, version upgrade 등 다양하다.

첫 대응으로 인덱스를 추가하면 운 좋게 해결될 수 있지만, 원인을 닫지 못하면 write cost와 lock/DDL risk만 늘어난다. 먼저 느려진 시점, 데이터 변화, plan 변화, estimated/actual gap, rows examined, sort/temp/spill을 맞춰 봐야 한다.

parameter-sensitive한 쿼리는 어떤 값에서는 작고 어떤 값에서는 거대하다. `merchant_id=small`은 10행이지만 `merchant_id=m-big`은 900000행이면 같은 prepared statement나 plan cache 감각으로는 안전하지 않다.

운영 데이터 분포를 보지 않고 로컬 plan만 믿으면 실패한다. 로컬 fixture는 균등하고 작으며 cache가 따뜻할 수 있다. 운영은 skew, cold page, vacuum/purge 지연, replication lag, lock wait를 함께 가진다.

느린 쿼리 진단에서는 단일 실행 시간이 아니라 시간선이 핵심이다. slow log가 언제 늘었고, 그 사이 데이터 분포와 통계, plan, lock wait, cache 상태, 배포가 어떻게 바뀌었는지 맞춰야 한다. 이 시간선을 복원해야 인덱스 추가가 원인 해결인지, 통계 갱신이나 parameter 분리 또는 lock 진단이 먼저인지 판단할 수 있다.

### 손으로 다시 읽는 예제

1. 통계 stale

    ```sql
    대량 import 직후 rows estimate가 작게 남아 nested loop를 고른다.
    ```

    ANALYZE 이후 hash/merge 또는 다른 index plan으로 바뀔 수 있다.
1. parameter skew

    ```sql
    merchant_id 작은 값과 큰 값의 rows가 100000배 차이난다.
    ```

    평균 selectivity가 대표값을 설명하지 못한다.
1. plan은 같지만 느림

    ```sql
    같은 index scan인데 buffer read가 급증한다.
    ```

    cache, bloat, row width, storage latency를 함께 본다.

위 예제들은 느린 쿼리 원인이 SQL 텍스트 하나에 고정되지 않는다는 점을 보여 준다. 통계 stale은 plan 선택을, parameter skew는 tenant별 비용을, 같은 plan의 buffer read 증가는 storage나 cache 상태를 의심하게 한다. 그래서 장애 분석은 쿼리 모양과 운영 시간선을 함께 읽어야 한다.

### 실무에서 먼저 의심해야 하는 함정

- 운영 데이터 분포 변화를 보지 않고 로컬 plan만 믿으면 장애를 재현하지 못한다. 로컬에서 빠른 쿼리가 운영의 큰 tenant에서도 빠르다는 보장은 없다.
- 느린 쿼리 하나마다 인덱스를 추가하면 쓰기 지연, DDL 위험, optimizer 선택 혼란, 디스크 사용량 증가가 누적된다.
- slow query log만 보고 lock wait와 CPU query time을 구분하지 않으면, 인덱스가 아니라 lock contention이나 connection pool 포화가 원인인 장애에 잘못 대응한다.

slow query 함정은 느린 SQL을 보면 즉시 인덱스부터 떠올릴 때 생긴다. 원인이 통계 stale, 큰 tenant, lock wait, cache miss, 데이터 폭 증가라면 인덱스 추가는 문제를 가리거나 쓰기 비용을 늘릴 수 있다. 느린 시점과 바뀐 조건을 먼저 맞춰야 한다.

### 관측과 검증 경로

- MySQL에서는 slow query log, EXPLAIN, performance_schema 또는 rows examined 지표를 연결한다. PASS는 느린 SQL의 발생 시점과 plan/row 변화가 같은 timeline 안에서 설명되는 것이다.
- PostgreSQL에서는 `EXPLAIN (ANALYZE, BUFFERS)`, 통계 갱신 전후 plan, `pg_stat_statements`의 mean/max time과 calls를 함께 본다. PASS는 한 번의 느림과 반복적인 regression을 구분하는 것이다.
- 실패 신호는 동일 SQL의 parameter별 latency 분산이 큰데 평균만 보거나, estimated/actual gap이 큰데 통계와 분포를 확인하지 않는 것이다.

slow query 검증은 느려진 SQL의 시간선과 plan/통계/대기 변화를 연결하는 일이다. MySQL slow query log와 EXPLAIN, PostgreSQL EXPLAIN ANALYZE와 pg_stat_statements 같은 관측 표면을 맞춘다. PASS는 인덱스 추가 전에도 원인이 데이터 분포인지 plan regression인지 lock wait인지 분리되는 것이다.

### slow query diagnosis and plan regression 현장 판독 노트 1

1-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

1-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 1번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 2

2-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

2-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 2번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 3

3-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

3-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 3번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 4

4-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

4-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 4번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 5

5-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

5-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 5번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 6

6-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

6-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 6번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 7

7-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

7-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 7번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### slow query diagnosis and plan regression 현장 판독 노트 8

8-1. `slow log timeline` 상황에서는 먼저 시간선을(를) 하나의 관측 단위로 분리한다. 느린 SQL 발생 시점과 배포/데이터 변경 시점을 맞춘다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 쿼리 텍스트만 보면 원인이 좁혀지지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-2. `plan regression` 상황에서는 먼저 계획 변화을(를) 하나의 관측 단위로 분리한다. 같은 SQL의 plan이나 실제 비용이 변했는지 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 이전 plan baseline이 없으면 감으로 대응한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-3. `tenant skew` 상황에서는 먼저 큰 tenant을(를) 하나의 관측 단위로 분리한다. 큰 tenant와 작은 tenant의 selectivity가 다르다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 평균 plan 하나로 모두를 설명할 수 없다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-4. `stats refresh` 상황에서는 먼저 통계 갱신을(를) 하나의 관측 단위로 분리한다. ANALYZE 전후 rows estimate를 비교한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 통계 문제를 인덱스 부족으로 오해하지 않는다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-5. `lock wait` 상황에서는 먼저 대기 구분을(를) 하나의 관측 단위로 분리한다. 느림이 CPU/query work인지 lock wait인지 분리한다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 인덱스 추가가 lock wait를 해결하지 못할 수 있다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

8-6. `index debt` 상황에서는 먼저 인덱스 부채을(를) 하나의 관측 단위로 분리한다. 인덱스는 읽기만 돕고 쓰기 비용을 늘린다. 여기서 바로 결론을 내리지 않고 입력 row 수, 후보 row 수, 다음 소비자, 실패 신호를 같은 줄에 적어 본다. 장애 때 만든 임시 인덱스도 나중에 소유자를 정해야 한다. 실무에서 이 점이 중요한 이유는 장애 티켓이 보통 "느리다" 한 단어로 들어오기 때문이다. 그 한 단어를 page, row, term, plan node, 통계, 로그 중 어느 층의 문제인지 나누지 않으면 가장 먼저 떠오른 인덱스 추가나 SQL 재작성으로 기울어진다. 반대로 이 판독 단계를 거치면 slow query diagnosis and plan regression에서 설명한 구조가 단순 암기가 아니라 재현 가능한 진단 절차가 된다.

이 8번째 판독 묶음의 핵심은 같은 설명을 반복하는 것이 아니라, slow query diagnosis and plan regression을(를) 운영 관측 단위로 잘게 나누어 보는 연습이다. 한 쿼리의 평균 시간만 보지 말고, 어떤 입력이 어떤 내부 구조를 지나 어떤 산출물을 만들었는지 작은 trace로 되돌려야 한다. 그 trace가 닫히면 공식 문서의 용어와 로컬 seed의 짧은 메모가 서로 연결되고, 닫히지 않으면 아직 원인을 안다고 말하면 안 된다.

### 오해 수리와 전이 질문

이 절에서 바로잡을 오해는 느린 쿼리의 답이 항상 새 인덱스라는 생각이다. 운영에서는 데이터 분포, 통계, lock, cache, parameter skew, plan 변화가 함께 움직인다. 원인을 닫지 않은 인덱스 추가는 다음 장애의 비용으로 남을 수 있다.

마지막으로 slow query diagnosis and plan regression을(를) 자기 말로 다시 설명하려면 세 문장을 직접 써 본다. 첫 문장은 이 구조가 해결하는 문제, 둘째 문장은 작은 입력이 내부에서 바뀌는 순서, 셋째 문장은 운영에서 깨지는 조건이다. 세 문장 중 하나라도 공식 문서나 trace로 되돌아가지 못하면 아직 외운 말이고, 세 문장이 모두 trace와 검증 명령으로 이어지면 다음 DU의 실행 계획을 읽을 준비가 된 것이다.
