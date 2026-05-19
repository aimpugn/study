# Relational model and SQL foundations

## 관계 모델, tuple, relation, key, set/bag semantics

관계형 데이터베이스를 “표를 저장하는 DB”라고만 이해하면 중요한 절반을 놓칩니다. 표 모양은 눈에 보이는 표현이고, 관계 모델의 핵심은 어떤 사실을 어떤 속성들의 묶음으로 표현하며, 그 묶음들이 어떤 조건에서 같은 것인지, 어떤 속성이 식별을 맡는지, 중복을 어떻게 다룰지 정하는 데 있습니다. PostgreSQL 공식 튜토리얼은 테이블을 행과 열로 이루어진 구조로 설명합니다. 이 설명은 출발점으로 좋지만, 공부할 때는 행을 단순한 줄이 아니라 tuple, 테이블을 단순한 시트가 아니라 relation으로 읽어야 합니다. tuple은 한 개의 사실을 구성하는 값들의 묶음이고, relation은 같은 속성 구조를 가진 tuple들의 집합에 가까운 모델입니다.

가장 작은 예로 주문 데이터를 봅니다.

```text
orders
+----------+---------+------------+--------+
| order_id | user_id | status     | amount |
+----------+---------+------------+--------+
| 1001     | 10      | READY      | 120000 |
| 1002     | 10      | PAID       |  50000 |
| 1003     | 20      | READY      |  30000 |
+----------+---------+------------+--------+
```

엑셀처럼 보면 이 표는 셀들의 격자입니다. 관계 모델로 보면 각 행은 주문이라는 사실 하나를 표현하는 tuple입니다.

```text
t1 = (order_id=1001, user_id=10, status=READY, amount=120000)
t2 = (order_id=1002, user_id=10, status=PAID,  amount=50000)
t3 = (order_id=1003, user_id=20, status=READY, amount=30000)
```

이 차이가 왜 중요할까요? 엑셀 시트에서는 같은 줄을 두 번 복사해도 사람이 보고 “아, 중복이네”라고 판단할 수 있습니다. DB에서는 그 중복이 허용되는지, 허용된다면 무엇을 뜻하는지, 허용되지 않는다면 어떤 속성으로 막을지 명시해야 합니다. `order_id`가 주문의 식별자라면 같은 `order_id=1001`인 tuple이 두 개 존재하면 안 됩니다. 이 규칙을 key가 맡습니다. key는 보기 좋게 붙인 번호가 아니라, relation 안에서 tuple을 식별하고 중복과 참조를 통제하는 계약입니다.

```sql
CREATE TABLE orders (
    order_id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    status text NOT NULL,
    amount integer NOT NULL
);
```

`PRIMARY KEY(order_id)`는 “이 테이블에서 주문 하나를 식별할 때 `order_id`를 사용하고, 같은 `order_id`는 두 번 들어올 수 없다”는 뜻입니다. 이 제약이 없으면 다음 두 행이 들어와도 DB는 표면상 저장할 수 있습니다.

```text
+----------+---------+--------+
| order_id | user_id | amount |
+----------+---------+--------+
| 1001     | 10      | 120000 |
| 1001     | 10      | 120000 |
+----------+---------+--------+
```

이 두 줄이 완전히 같은 주문을 실수로 두 번 넣은 것인지, 같은 주문의 이력 두 건인지, 일부러 허용한 duplicate event인지 DB는 알 수 없습니다. key가 없으면 애플리케이션이 매번 의미를 추정해야 합니다. key를 두면 DB가 중복을 거부하고, 참조하는 다른 테이블은 그 key를 기준으로 연결할 수 있습니다.

관계 모델에서 relation은 수학적으로 set에 가까운 개념입니다. set은 같은 원소를 두 번 갖지 않습니다. 하지만 SQL 테이블과 SELECT 결과는 실제로 bag, 즉 중복을 허용하는 다중집합처럼 동작하는 경우가 많습니다. 이 차이는 초보자뿐 아니라 실무자도 자주 놓칩니다. SQL에서 `SELECT status FROM orders`를 실행하면 status가 같은 주문이 여러 개일 때 같은 값이 여러 번 나옵니다.

```sql
SELECT status
FROM orders;
```

```text
status
------
READY
PAID
READY
```

관계 대수의 이상적인 set 결과라면 `READY`는 한 번만 나와야 할 것 같지만, SQL은 기본적으로 중복을 보존합니다. 중복을 제거하려면 `DISTINCT`를 명시해야 합니다.

```sql
SELECT DISTINCT status
FROM orders;
```

```text
status
------
PAID
READY
```

이 차이를 set semantics와 bag semantics로 부릅니다. 한국어로는 각각 “중복을 허용하지 않는 집합 의미”와 “중복을 허용하는 다중집합 의미”라고 풀어 쓰는 편이 안전합니다. `semantics`는 여기서 문법이 아니라 “그 표현이 어떤 결과 의미를 갖는가”를 뜻합니다. SQL이 bag처럼 동작하는 이유는 실용적입니다. 실제 테이블에는 같은 값 조합이 여러 번 나올 수 있고, 중복을 매번 제거하려면 비용이 듭니다. 주문이 100만 건 있는데 상태가 `READY`, `PAID`, `CANCELLED` 세 종류뿐이라면, `SELECT status`는 100만 행을 반환할 수 있습니다. `DISTINCT status`는 중복 제거 작업을 해야 합니다.

```text
orders.status 원본 흐름

row1 READY
row2 PAID
row3 READY
row4 READY

SELECT status
  -> READY, PAID, READY, READY

SELECT DISTINCT status
  -> READY, PAID

COUNT(status)
  -> 4

COUNT(DISTINCT status)
  -> 2
```

이 worked trace는 단순해 보이지만 실무 버그를 많이 막습니다. 예를 들어 join 뒤 카운트가 예상보다 커지는 문제는 대부분 “관계가 하나씩만 붙을 것”이라고 믿었지만 실제로는 오른쪽 relation에 여러 tuple이 매칭되어 bag 결과가 증폭되는 경우입니다. 기존 seed `database/join.md`에는 서브쿼리를 join하는 짧은 예가 있습니다. 그 예도 결국 “왼쪽 행 하나에 오른쪽 후보가 몇 개 붙을 수 있는가”를 묻는 문제로 확장해야 합니다.

```sql
CREATE TABLE users (
    user_id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE
);

CREATE TABLE orders (
    order_id bigint PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(user_id),
    amount integer NOT NULL
);
```

```text
users
+---------+----------------+
| user_id | email          |
+---------+----------------+
| 10      | a@example.com  |
| 20      | b@example.com  |
+---------+----------------+

orders
+----------+---------+--------+
| order_id | user_id | amount |
+----------+---------+--------+
| 1001     | 10      | 120000 |
| 1002     | 10      |  50000 |
| 1003     | 20      |  30000 |
+----------+---------+--------+
```

```sql
SELECT u.user_id, u.email, o.order_id, o.amount
FROM users u
JOIN orders o ON o.user_id = u.user_id;
```

```text
결과
+---------+---------------+----------+--------+
| user_id | email         | order_id | amount |
+---------+---------------+----------+--------+
| 10      | a@example.com | 1001     | 120000 |
| 10      | a@example.com | 1002     |  50000 |
| 20      | b@example.com | 1003     |  30000 |
+---------+---------------+----------+--------+
```

`users`의 `user_id=10` 행은 하나였지만 결과에서는 두 번 나옵니다. 이것은 DB가 중복을 잘못 만든 것이 아닙니다. `users`와 `orders`의 관계가 1:N이기 때문에, 사용자 tuple 하나가 주문 tuple 두 개와 각각 결합된 것입니다. relation을 엑셀처럼만 보면 “왜 user가 두 번 나오지?”라고 놀라지만, tuple의 결합으로 보면 자연스럽습니다. join은 두 relation에서 조건을 만족하는 tuple 쌍을 만들어 새 tuple을 구성합니다.

```text
u10 = (user_id=10, email=a@example.com)
o1001 = (order_id=1001, user_id=10, amount=120000)
o1002 = (order_id=1002, user_id=10, amount=50000)

u10 ⋈ o1001 -> (10, a@example.com, 1001, 120000)
u10 ⋈ o1002 -> (10, a@example.com, 1002, 50000)
```

이 trace가 중요한 이유는 집계 버그를 예방하기 때문입니다. 사용자 수를 세고 싶은데 주문과 join한 뒤 `count(*)`를 하면 주문 수를 셀 수 있습니다.

```sql
SELECT count(*)
FROM users u
JOIN orders o ON o.user_id = u.user_id;
```

위 결과는 3입니다. 사용자는 2명이지만 join 결과 tuple은 3개이기 때문입니다. 사용자 수를 세려면 `count(DISTINCT u.user_id)`를 쓰거나, 먼저 사용자 relation에서 세거나, 의도에 맞는 grouping을 해야 합니다. `DISTINCT`를 무조건 붙이면 결과가 맞아 보일 수도 있지만, 왜 중복이 생겼는지 모르면 다른 쿼리에서 다시 깨집니다.

key는 join 결과의 크기를 예측하는 데도 필요합니다. `users.user_id`가 primary key이면 `orders.user_id` 하나는 최대 한 사용자와만 연결됩니다. 반대로 `orders.user_id`는 여러 주문에 반복될 수 있습니다. 따라서 `orders -> users` 방향은 N:1이고, `users -> orders` 방향은 1:N입니다. 이 cardinality, 즉 관계의 개수 감각이 없으면 join이 결과를 줄이는지 늘리는지 예상하기 어렵습니다.

```text
관계 방향별 tuple 수 감각

orders JOIN users ON orders.user_id = users.user_id
  orders 한 행 -> users 최대 한 행
  결과 크기: orders보다 커지지 않는 편

users JOIN orders ON orders.user_id = users.user_id
  users 한 행 -> orders 여러 행 가능
  결과 크기: users보다 커질 수 있음
```

물론 outer join, 누락된 참조, 잘못된 key, 중복 데이터가 있으면 결과는 더 복잡해집니다. 그래서 key는 성능을 위한 인덱스만이 아니라 의미를 위한 제약입니다. primary key와 unique constraint는 relation 안의 tuple 식별을 고정하고, foreign key는 relation 사이의 참조 가능성을 고정합니다. 인덱스는 이 제약을 빠르게 검사하거나 탐색하는 물리 장치가 될 수 있지만, key의 본질은 “무엇이 같은 사실이고 무엇이 다른 사실인가”입니다.

관계 모델의 등장 배경도 이 지점과 연결됩니다. 파일 중심 시스템에서는 프로그램이 파일 구조를 알고 있어야 했습니다. 고객 파일, 주문 파일, 상품 파일이 각각 다른 형식을 갖고, 프로그램이 그 파일의 위치와 레코드 구조를 직접 다뤘습니다. 데이터 의미가 프로그램에 묶이면 같은 고객 정보를 여러 파일에 복사하고, 한쪽만 수정되어 불일치가 생기기 쉽습니다. 관계 모델은 데이터를 값들의 관계로 표현하고, 사용자가 원하는 결과를 선언적으로 질의하게 하려는 방향을 제시했습니다. 물리 저장 방식이 바뀌어도 논리 relation과 질의 의미는 최대한 유지하려는 발상입니다.

이 배경을 알아야 normalization, 즉 정규화도 장식 규칙이 아니라는 점이 보입니다. 정규화는 모든 테이블을 잘게 쪼개라는 종교가 아닙니다. 같은 사실이 여러 곳에 중복 저장되어 수정 이상, 삽입 이상, 삭제 이상이 생기는 문제를 줄이려는 모델링 방법입니다.

```text
나쁜 중복 예

order_lines
+----------+---------+----------------+------------+----------+
| order_id | user_id | user_email     | product_id | quantity |
+----------+---------+----------------+------------+----------+
| 1001     | 10      | a@example.com  | 501        | 1        |
| 1002     | 10      | old@example.com| 502        | 1        |
+----------+---------+----------------+------------+----------+

문제:
- 같은 user_id=10인데 email이 두 값이다.
- 어떤 email이 최신인지 tuple만 보고는 알기 어렵다.
- 사용자의 email 변경이 주문 라인 전체 수정으로 번진다.
```

이 문제를 줄이려면 사용자 사실과 주문 사실을 분리합니다.

```text
users
+---------+---------------+
| user_id | email         |
+---------+---------------+
| 10      | a@example.com |
+---------+---------------+

orders
+----------+---------+
| order_id | user_id |
+----------+---------+
| 1001     | 10      |
| 1002     | 10      |
+----------+---------+
```

이렇게 나누면 join이 필요해지고, join 비용과 모델 복잡도가 생깁니다. 하지만 사용자 이메일이라는 사실의 소유 위치가 명확해집니다. 실무 설계는 정규화와 반정규화 사이의 trade-off입니다. 조회 성능이나 이력 보존 때문에 일부 값을 복사해 둘 수 있습니다. 예를 들어 주문 당시의 배송지나 상품명을 주문 테이블에 스냅샷으로 남기는 것은 중복이지만 의미 있는 중복입니다. 핵심은 중복이 있는지 없는지가 아니라, 그 중복이 어떤 사실을 보존하려는 것인지 명시되어 있는가입니다.

set/bag 의미는 `UNION`에서도 드러납니다.

```sql
SELECT user_id FROM orders WHERE status = 'READY'
UNION
SELECT user_id FROM orders WHERE amount >= 100000;

SELECT user_id FROM orders WHERE status = 'READY'
UNION ALL
SELECT user_id FROM orders WHERE amount >= 100000;
```

`UNION`은 중복을 제거하고, `UNION ALL`은 중복을 보존합니다. 같은 사용자가 READY 주문도 있고 100000원 이상 주문도 있으면 `UNION ALL`에서는 두 번 나올 수 있습니다. 비용도 다릅니다. 중복 제거를 하려면 정렬이나 해시가 필요할 수 있습니다. 이처럼 중복 의미는 결과 의미와 물리 비용을 동시에 바꿉니다.

```text
READY user_id:        [10, 20, 10]
amount>=100000 user_id:[10, 30]

UNION:
  [10, 20, 30]

UNION ALL:
  [10, 20, 10, 10, 30]
```

SQL의 bag 의미를 모르고 `UNION ALL`을 쓰면 중복이 늘고, 중복 제거가 필요한데 `UNION ALL`을 쓰면 잘못된 결과가 나옵니다. 반대로 중복이 의미 있는 이벤트 목록인데 `UNION`을 쓰면 실제 발생 횟수를 잃습니다. 이벤트, 로그, 결제 시도, 재시도 이력처럼 “같은 값 조합이 여러 번 발생했다”는 사실 자체가 중요한 데이터에서는 bag 의미가 필요합니다. 사용자 목록, 상태 목록, 고유 코드 목록처럼 중복이 의미 없는 결과에서는 set 의미가 필요합니다.

NULL도 key와 relation 의미를 복잡하게 만듭니다. primary key는 NULL을 허용하지 않습니다. 식별자가 비어 있으면 tuple을 안정적으로 식별할 수 없기 때문입니다. unique constraint에서 NULL을 어떻게 다루는지는 DB별 차이가 있으므로 공식 문서와 실제 DB에서 확인해야 합니다. 이 섹션의 핵심은 “빈 값도 그냥 값 하나 아닌가?”라는 오해를 버리는 것입니다. 식별에 쓰는 값은 반드시 존재해야 하고, 비교 가능해야 하며, relation 안에서 tuple을 구분할 수 있어야 합니다. NULL과 3-valued logic은 뒤 DU05에서 깊게 다루지만, key를 이해할 때 이미 NULL은 단순한 문자열 `''`나 숫자 0과 다르다는 점을 잡아야 합니다.

실무에서 key를 장식으로 취급하면 어떤 일이 생길까요? 이메일로 사용자를 찾는 서비스가 있는데 DB에는 unique constraint가 없다고 가정합니다.

```text
users
+---------+----------------+
| user_id | email          |
+---------+----------------+
| 10      | a@example.com  |
| 11      | a@example.com  |
+---------+----------------+
```

로그인 쿼리는 이렇게 되어 있습니다.

```sql
SELECT user_id
FROM users
WHERE email = 'a@example.com';
```

애플리케이션은 “한 명만 나올 것”이라고 가정했지만 DB는 두 행을 반환합니다. 어떤 프레임워크는 첫 행만 가져오고, 어떤 코드는 예외를 던지고, 어떤 코드는 마지막 행으로 덮어쓸 수 있습니다. 장애는 DB에서 시작된 것이 아니라 “email이 사용자 식별자로 쓰인다는 사실을 DB 제약으로 고정하지 않은 것”에서 시작됐습니다. unique constraint를 나중에 추가하려고 하면 이미 중복 데이터가 있어서 마이그레이션이 막힐 수 있습니다. key는 초기에 귀찮은 제약이 아니라, 나중의 복구 비용을 줄이는 계약입니다.

```sql
ALTER TABLE users
ADD CONSTRAINT users_email_uk UNIQUE (email);
```

이 명령은 단순히 인덱스를 추가하는 작업이 아닙니다. 앞으로 같은 email이 두 번 들어오면 실패해야 한다는 의미를 데이터 모델에 넣습니다. 운영에서는 기존 중복을 정리하고, 어떤 email을 살릴지 결정하고, 관련 주문·권한·세션을 어떻게 합칠지 검토해야 합니다. 그래서 key 설계는 모델링 초기에 매우 중요합니다.

관계 모델을 제대로 읽으려면 “row가 무엇을 대표하는가”를 먼저 물어야 합니다. `orders`의 row는 주문 헤더인가, 주문 라인인가, 결제 시도인가, 정산 대상인가. 이름이 비슷해도 tuple의 의미가 다르면 key도 달라집니다. 주문 헤더는 `order_id`가 key일 수 있고, 주문 라인은 `(order_id, line_no)`가 key일 수 있고, 결제 시도는 `payment_attempt_id`가 key일 수 있습니다. 잘못된 key를 잡으면 중복을 막아야 할 곳을 못 막거나, 여러 번 발생해야 하는 사건을 부당하게 막습니다.

```text
주문 헤더 tuple:
  (order_id=1001, user_id=10, status=READY)
  key: order_id

주문 라인 tuple:
  (order_id=1001, line_no=1, product_id=501, quantity=1)
  key: (order_id, line_no)

결제 시도 tuple:
  (payment_attempt_id=9001, order_id=1001, pg='KCP', result='TIMEOUT')
  key: payment_attempt_id
  같은 order_id에 여러 attempt가 있을 수 있음
```

이 예에서 `order_id`를 결제 시도 테이블의 unique key로 잡으면 재시도 이력을 저장할 수 없습니다. 반대로 `payment_attempt_id`만 있고 `(order_id, idempotency_key)` 같은 재시도 방지 key가 없으면 같은 외부 요청이 두 번 들어왔을 때 중복 결제가 생길 수 있습니다. 관계 모델은 추상 이론이 아니라 도메인의 사실 단위를 정확히 나누는 실무 도구입니다.

검증 경로는 다음처럼 잡을 수 있습니다.

```sql
-- 중복을 허용하는 bag 결과 확인
SELECT status FROM orders;

-- 중복 제거 확인
SELECT DISTINCT status FROM orders;

-- join 증폭 확인
SELECT u.user_id, count(*) AS joined_rows
FROM users u
JOIN orders o ON o.user_id = u.user_id
GROUP BY u.user_id;

-- key 위반 확인
INSERT INTO users(user_id, email) VALUES (10, 'a@example.com');
```

PASS는 각 쿼리의 결과를 손으로 예측하고, 중복이 의미 있는지 제거되어야 하는지 설명할 수 있는 상태입니다. FAIL은 join 결과가 늘어난 것을 DB 버그로 보거나, `DISTINCT`를 원인 분석 없이 붙이거나, key 제약 위반을 “운영에서만 조심하면 된다”고 넘기는 상태입니다. 특히 운영 장애에서 `DISTINCT`는 증상을 가릴 수 있습니다. 중복된 사용자가 왜 생겼는지, 잘못된 join cardinality가 어디서 시작됐는지, key가 빠진 것인지, 데이터 정합성이 깨진 것인지 먼저 봐야 합니다.

마지막으로 이 섹션의 기억 문장을 정리합니다. relation은 표처럼 보이지만, 실제로는 같은 속성 구조를 가진 tuple들의 의미 있는 모음입니다. key는 tuple을 식별하고 중복과 참조를 통제하는 계약입니다. SQL은 관계 모델의 set 이상을 그대로만 따르지 않고, 실용적으로 bag 의미를 많이 사용합니다. 그래서 `DISTINCT`, `GROUP BY`, `JOIN`, `UNION`, key 제약을 볼 때는 항상 “중복이 지금 의미 있는가, 아니면 제거되어야 하는가”를 물어야 합니다. 이 질문을 놓치면 테이블을 엑셀처럼 다루게 되고, key와 무결성을 장식으로 취급하는 함정에 빠집니다.

key를 더 깊게 보면 candidate key, primary key, alternate key, foreign key라는 구분이 나옵니다. 한국어로는 후보키, 기본키, 대체키, 외래키라고 많이 옮깁니다. 후보키는 tuple을 유일하게 식별할 수 있는 속성 집합 후보입니다. 기본키는 그 후보 중 이 relation에서 대표 식별자로 선택한 key입니다. 대체키는 후보였지만 기본키로 선택하지 않은 key입니다. 외래키는 다른 relation의 key를 참조해 관계를 표현하는 속성입니다. 이 이름들은 시험용 분류처럼 보이지만, 실무에서는 “어떤 값이 진짜 식별자인가”를 정하는 대화의 언어가 됩니다.

```text
users relation

attributes:
  user_id
  email
  phone_number

candidate keys:
  user_id        -- 시스템이 발급한 식별자
  email          -- 서비스 정책상 중복 가입을 막는다면 후보
  phone_number   -- 전화번호 인증 정책상 후보일 수 있음

primary key:
  user_id

alternate keys:
  email
  phone_number
```

후보키를 고를 때 가장 흔한 고민은 자연키와 대리키입니다. 자연키는 도메인에서 이미 의미를 가진 값입니다. 이메일, 주민등록번호, 사업자등록번호, ISBN, 계좌번호 같은 값이 후보가 될 수 있습니다. 대리키는 DB나 애플리케이션이 식별을 위해 만든 값입니다. `user_id`, `order_id`, UUID 같은 값입니다. 자연키는 사람이 이해하기 쉽고 외부 시스템과 맞추기 편하지만, 값이 바뀌거나 재사용되거나 정책이 달라질 수 있습니다. 대리키는 내부 참조가 안정적이지만, 도메인 중복을 막으려면 별도의 unique constraint가 필요합니다.

```text
email을 primary key로 쓰는 설계

장점:
  사용자를 찾는 업무 식별자와 DB 식별자가 같다.
  별도 user_id 없이도 참조가 가능하다.

위험:
  이메일 변경이 모든 참조 관계에 영향을 준다.
  대소문자/정규화/collation 정책이 식별 정책으로 들어온다.
  같은 사람이 여러 이메일을 가질 수 있는 정책 변경에 취약하다.

user_id를 primary key로 쓰고 email unique를 두는 설계

장점:
  내부 참조는 user_id로 안정적이다.
  email 변경이 외래키 전체 변경으로 번지지 않는다.
  email 중복 금지는 unique constraint로 따로 표현한다.

위험:
  email unique를 빼먹으면 도메인 중복을 못 막는다.
  외부 시스템과 매핑하는 식별자 관리가 필요하다.
```

이 비교에서 정답은 항상 `user_id`라는 뜻이 아닙니다. 어떤 시스템에서는 외부 표준 코드가 안정적이고 변하지 않아 자연키가 적절할 수 있습니다. 반대로 사람, 주문, 결제 시도처럼 정책 변화와 외부 입력이 많은 영역에서는 대리키와 도메인 unique key를 함께 두는 편이 안전한 경우가 많습니다. 중요한 것은 key를 성능 튜닝 도구로만 보지 않고, 도메인 사실의 식별 정책으로 보는 것입니다.

복합키도 자주 등장합니다. 주문 라인처럼 하나의 주문 안에서 `line_no`가 의미를 가지는 경우 `(order_id, line_no)`가 자연스러운 key일 수 있습니다.

```sql
CREATE TABLE order_lines (
    order_id bigint NOT NULL REFERENCES orders(order_id),
    line_no integer NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, line_no)
);
```

이 모델은 `line_no=1`이 전역적으로 유일하다는 뜻이 아니라, 같은 `order_id` 안에서만 유일하다는 뜻입니다. 이 구조를 모르고 `line_no`만 foreign key로 참조하려고 하면 의미가 깨집니다. 복합키는 참조하는 쪽에서도 같은 속성 묶음을 가져야 하므로 테이블이 번거로워질 수 있습니다. 그래서 내부 surrogate key를 추가하고 `(order_id, line_no)`에는 unique constraint를 두는 설계도 자주 씁니다.

```sql
CREATE TABLE order_lines (
    order_line_id bigint PRIMARY KEY,
    order_id bigint NOT NULL REFERENCES orders(order_id),
    line_no integer NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL CHECK (quantity > 0),
    UNIQUE (order_id, line_no)
);
```

두 설계는 같은 요구를 다른 방식으로 표현합니다. 첫 번째는 복합키 자체가 식별자입니다. 두 번째는 내부 참조용 식별자를 따로 두고, 도메인 규칙은 unique constraint로 둡니다. 어떤 설계가 좋은지는 참조 빈도, 외부 API 모양, 변경 가능성, ORM 편의, 마이그레이션 비용에 따라 달라집니다. 다만 어느 쪽이든 `(order_id, line_no)`가 주문 안에서 중복되면 안 된다는 규칙은 사라지면 안 됩니다.

관계 모델을 엑셀처럼 보면 NULL과 빈 문자열도 쉽게 섞입니다. 예를 들어 `middle_name`은 모를 수 있고, 없을 수도 있습니다. `''`는 빈 문자열이라는 값이고, `NULL`은 값이 없거나 알 수 없다는 표시입니다. relation의 tuple 안에 어떤 속성이 NULL일 수 있다는 것은 그 속성이 tuple의 식별이나 필수 사실을 맡지 않는다는 뜻일 수 있습니다. 반대로 key 속성이 NULL이면 tuple을 안정적으로 식별할 수 없습니다. 그래서 primary key는 NOT NULL입니다. “빈 칸도 값 하나”라는 엑셀식 감각을 DB에 가져오면 unique, join, group, count에서 예상과 다른 결과가 나옵니다.

```text
customer_profiles
+-------------+-------------+
| customer_id | middle_name |
+-------------+-------------+
| 1           | NULL        |  -- 모름/없음/수집 안 함
| 2           | ''          |  -- 빈 문자열이라는 값
+-------------+-------------+

COUNT(*) = 2
COUNT(middle_name) = 1
```

이 예는 DU05에서 다시 자세히 다루지만, relation 모델 단계에서도 속성의 의미를 정해야 한다는 점을 보여 줍니다. 속성이 선택인지 필수인지, NULL이 허용된다면 무엇을 뜻하는지, 빈 문자열과 구분해야 하는지, unique 비교에서 어떤 정책을 원하는지 모델링 때 결정해야 합니다.

relation 사이의 관계도 1:1, 1:N, N:M으로만 외우면 부족합니다. 중요한 것은 어떤 tuple이 어떤 tuple을 참조할 수 있고, 참조 대상이 없어질 때 어떻게 할지입니다. 사용자를 삭제할 때 주문도 삭제할 것인가, 주문은 보존하고 사용자 참조만 익명화할 것인가, 삭제를 막을 것인가. foreign key에는 `ON DELETE` 정책이 붙을 수 있지만, 그 정책은 비즈니스와 감사 요구를 반영해야 합니다.

```sql
CREATE TABLE orders (
    order_id bigint PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT
);
```

`ON DELETE RESTRICT`는 주문이 있는 사용자를 그냥 삭제하지 못하게 합니다. 결제와 정산 이력이 있는 서비스에서는 이 정책이 자연스러울 수 있습니다. 반대로 임시 장바구니 item은 사용자가 삭제될 때 함께 삭제해도 될 수 있습니다. 관계 모델의 제약은 기술 선택이 아니라 사실의 생명주기를 표현합니다.

N:M 관계는 중간 relation으로 풀어야 합니다. 사용자와 역할이 다대다라면 `user_roles` 같은 relation을 둡니다.

```sql
CREATE TABLE roles (
    role_id bigint PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE user_roles (
    user_id bigint NOT NULL REFERENCES users(user_id),
    role_id bigint NOT NULL REFERENCES roles(role_id),
    granted_at timestamp NOT NULL,
    PRIMARY KEY (user_id, role_id)
);
```

`user_roles`는 단순 연결 테이블이 아니라 “사용자가 어떤 역할을 부여받았다”는 사실의 relation입니다. `granted_at`, `granted_by`, `revoked_at` 같은 속성이 붙기 시작하면 더 분명해집니다. 이 relation의 primary key를 `(user_id, role_id)`로 둘지, `user_role_id`를 따로 두고 이력을 여러 번 남길지 역시 도메인 의미에 달려 있습니다. 같은 역할을 한 번만 부여할 수 있다면 복합키가 맞고, 부여와 회수 이력을 사건으로 남겨야 한다면 별도 event relation이 필요할 수 있습니다.

```text
현재 상태 모델:
  user_roles(user_id, role_id) 하나만 존재
  같은 역할 재부여 이력은 보존하지 않음

이력 모델:
  role_grant_events(event_id, user_id, role_id, action, occurred_at)
  GRANTED/REVOKED 사건을 여러 번 기록
```

상태와 이력을 구분하지 않으면 key 설계가 흔들립니다. 현재 상태 relation에서는 같은 `(user_id, role_id)`가 두 번 있으면 중복입니다. 이력 relation에서는 같은 사용자와 역할 조합이 여러 번 나올 수 있습니다. 대신 각 사건은 `event_id`나 `(user_id, role_id, occurred_at, action)` 같은 식별자가 필요합니다. 관계 모델은 결국 “내가 저장하는 tuple은 상태인가 사건인가”를 묻게 만듭니다.

bag semantics는 이력 모델에서 특히 중요합니다. 같은 사용자가 같은 상품을 같은 금액으로 두 번 결제 시도했다면 값 조합이 비슷해도 두 사건은 둘 다 보존되어야 합니다.

```text
payment_attempts
+------------+----------+--------+---------+
| attempt_id | order_id | amount | result  |
+------------+----------+--------+---------+
| 9001       | 1001     | 50000  | TIMEOUT |
| 9002       | 1001     | 50000  | TIMEOUT |
+------------+----------+--------+---------+
```

여기서 `DISTINCT order_id, amount, result`를 해 버리면 두 번의 timeout 사건이 하나로 사라집니다. 반대로 현재 결제 상태만 보고 싶다면 attempt를 group하거나 최신 attempt를 골라야 합니다. 따라서 중복을 제거할지 보존할지는 데이터가 상태인지 사건인지와 연결됩니다.

실무 관측에서는 DB catalog를 통해 key와 constraint를 확인할 수 있습니다. PostgreSQL에서는 `information_schema`나 `pg_catalog`를 볼 수 있고, MySQL에서는 `information_schema`와 `SHOW CREATE TABLE`을 볼 수 있습니다. 운영에서 “이 테이블은 email unique가 걸려 있나요?”라는 질문을 코드 추측으로 답하면 안 됩니다. 실제 DB 정의를 확인해야 합니다.

```sql
-- MySQL에서 테이블 정의 확인
SHOW CREATE TABLE users;

-- 표준 information_schema 계열에서 key column 확인 예
SELECT constraint_name, column_name
FROM information_schema.key_column_usage
WHERE table_name = 'users';
```

PASS는 문서나 ORM entity만 보는 것이 아니라 실제 DB에 어떤 key와 constraint가 있는지 확인하는 것입니다. FAIL은 “엔티티에 `@Column(unique=true)`가 있으니 DB에도 있겠지”라고 가정하는 것입니다. migration이 누락되었거나, 운영 DB에 수동 변경이 있었거나, 테스트 DB와 운영 DB가 다를 수 있습니다. 관계 모델의 계약은 최종적으로 DB schema에 존재해야 합니다.

관계 모델을 공부하는 목적은 이론 용어를 외우는 것이 아니라, 데이터 불일치가 생기기 전에 질문을 던지는 힘을 얻는 것입니다. 이 tuple은 어떤 사실을 대표하는가. 이 fact가 두 번 들어오면 중복인가 이력인가. 어떤 속성이 이 fact를 식별하는가. 외부에서 들어온 값이 식별자로 안정적인가. 같은 값을 여러 relation에 복사했다면 어떤 시점의 사실을 보존하려는 것인가. join 뒤 결과가 늘어나는 것은 의도인가. count가 세는 것은 entity 수인가 event 수인가 tuple 수인가. 이 질문에 답하면 SQL과 schema가 훨씬 단단해집니다.

관계 모델은 “데이터를 나누는 기술”이 아니라 “사실의 경계를 정하는 기술”입니다. 예를 들어 상품 가격을 어디에 둘지 생각해 봅니다. `products.price`에 현재 가격을 두면 상품의 현재 상태를 표현합니다. `order_lines.unit_price`에 주문 당시 가격을 복사하면 주문이라는 사건이 발생한 시점의 가격 스냅샷을 보존합니다. 두 값이 같을 수도 있지만 의미는 다릅니다.

```text
products
+------------+------------+---------------+
| product_id | name       | current_price |
+------------+------------+---------------+
| 501        | Keyboard   | 70000         |
+------------+------------+---------------+

order_lines
+----------+---------+------------+------------+
| order_id | line_no | product_id | unit_price |
+----------+---------+------------+------------+
| 1001     | 1       | 501        | 65000      |
+----------+---------+------------+------------+
```

`current_price=70000`과 `unit_price=65000`은 불일치가 아니라 서로 다른 사실입니다. 현재 상품 가격과 주문 당시 체결 가격이 다르기 때문입니다. 이 경우 중복처럼 보이는 값 복사는 정당한 반정규화가 아니라 더 정확히는 사건의 속성 보존입니다. 반대로 현재 사용자 이메일을 주문 테이블마다 복사해 두고 “항상 최신이어야 한다”고 기대한다면 수정 이상이 생깁니다. 같은 값 복사라도 어떤 사실을 대표하는지에 따라 안전한 설계와 위험한 설계가 갈립니다.

이 구분은 감사와 정산에서 특히 중요합니다. 결제 금액, 할인 금액, 세금, 환율, 수수료처럼 시간이 지나면 정책이 바뀔 수 있는 값은 “현재 정책으로 다시 계산한 값”과 “당시 거래에 사용한 값”을 분리해야 합니다. relation의 tuple이 상태인지 사건인지, 현재값인지 스냅샷인지, 계산 결과인지 원천 입력인지 명확해야 나중에 재처리와 감사가 가능합니다. 따라서 key와 relation을 설계할 때는 컬럼명만 보지 말고 그 tuple이 어느 시점의 어떤 사실을 증명하는지 적어 보는 습관이 필요합니다. 이것이 관계 모델을 실무 설계 언어로 쓰는 방법입니다.

### 마지막 실무 연결: key는 조회 편의가 아니라 사실의 경계다

관계 모델에서 key를 장식처럼 보면 중복 데이터가 조용히 쌓인다. 예를 들어 같은 결제 이벤트가 두 번 들어왔을 때 event id에 unique key가 없으면 애플리케이션 if문 하나가 빠지는 순간 원장 row가 두 번 생긴다. 반대로 key가 정확하면 중복 입력은 정상적인 충돌로 드러나고, 코드는 그 충돌을 replay나 무시로 처리할 수 있다.

그래서 key 설계는 성능 설계이기 전에 사실의 경계 설계다. 어떤 값 조합이 같은 고객, 같은 주문, 같은 외부 이벤트, 같은 원장 이동을 뜻하는지 먼저 정해야 한다. 이 질문이 닫히지 않으면 index를 아무리 추가해도 데이터 의미는 안전해지지 않는다.

### 추가 판정 질문: 이 테이블은 어떤 사실을 한 번만 말하는가

관계 모델을 실제 설계에 적용할 때 마지막으로 묻는 질문은 이 테이블이 어떤 사실을 한 번만 말하는가다. 주문 테이블은 주문 헤더 사실을 말하고, 주문 라인 테이블은 특정 주문 안의 특정 상품 수량을 말하며, 결제 원장 테이블은 돈이 움직인 이유를 말한다. 이 사실의 단위가 흐려지면 한 테이블이 고객, 주문, 결제, 정산을 동시에 말하게 되고, 중복과 NULL과 불완전한 key가 늘어난다.

좋은 schema 리뷰는 column 목록을 보기 전에 fact grain을 말한다. grain이 닫히면 primary key, foreign key, unique constraint, nullable 여부가 자연스럽게 따라온다. grain이 닫히지 않으면 나중에 index를 아무리 추가해도 같은 사실이 여러 row에 흩어지는 문제를 막기 어렵다.


## SELECT 문이 논리적으로 처리되는 순서

`SELECT` 문은 SQL에서 가장 자주 쓰는 문장이지만, 표면 문법 순서대로 이해하면 금방 막힙니다. 문장은 `SELECT`로 시작하지만, 논리적으로는 보통 `FROM`에서 후보 relation을 만들고, `WHERE`로 행을 줄이고, `GROUP BY`로 묶고, `HAVING`으로 그룹을 줄이고, `SELECT` 목록으로 출력식을 만들고, `ORDER BY`와 `LIMIT`으로 결과 표현을 마무리합니다. PostgreSQL 공식 `SELECT` 문서는 일반 처리 순서를 단계적으로 설명합니다. 이 순서는 실제 실행 계획의 물리 순서가 아니라 SQL 결과가 어떤 의미로 만들어지는지 이해하는 기준입니다.

작은 데이터부터 고정합니다.

```sql
CREATE TABLE orders (
    order_id integer PRIMARY KEY,
    user_id integer NOT NULL,
    status text NOT NULL,
    amount integer NOT NULL,
    created_at date NOT NULL
);

INSERT INTO orders VALUES
  (1, 10, 'READY', 120000, DATE '2026-05-18'),
  (2, 10, 'READY',  80000, DATE '2026-05-18'),
  (3, 20, 'PAID',  200000, DATE '2026-05-18'),
  (4, 20, 'READY',  30000, DATE '2026-05-19'),
  (5, 30, 'READY', 150000, DATE '2026-05-18');
```

다음 쿼리를 손으로 따라갑니다.

```sql
SELECT user_id,
       count(*) AS order_count,
       sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id
HAVING sum(amount) >= 150000
ORDER BY total_amount DESC;
```

논리 trace는 다음과 같습니다.

```text
1. FROM orders
   후보: row 1, 2, 3, 4, 5

2. WHERE status='READY' AND created_at='2026-05-18'
   남음: row 1, 2, 5
   버림: row 3(PAID), row 4(날짜 다름)

3. GROUP BY user_id
   group user_id=10: row 1, row 2
   group user_id=30: row 5

4. aggregate 계산
   user_id=10: count=2, sum=200000
   user_id=30: count=1, sum=150000

5. HAVING sum(amount) >= 150000
   두 그룹 모두 남음

6. SELECT 출력
   (10, 2, 200000)
   (30, 1, 150000)

7. ORDER BY total_amount DESC
   (10, 2, 200000)
   (30, 1, 150000)
```

이 trace에서 `WHERE`와 `HAVING`의 차이를 먼저 잡아야 합니다. `WHERE`는 아직 개별 row를 보는 단계입니다. 그래서 `amount`, `status`, `created_at`처럼 row마다 존재하는 값을 조건으로 쓸 수 있습니다. `HAVING`은 group이 만들어지고 집계값이 계산된 뒤의 단계입니다. 그래서 `sum(amount)` 같은 그룹 결과를 조건으로 쓸 수 있습니다. `WHERE sum(amount) >= 150000`은 논리적으로 맞지 않습니다. group이 아직 없기 때문입니다.

```sql
-- 실패하는 사고방식
SELECT user_id, sum(amount)
FROM orders
WHERE sum(amount) >= 150000
GROUP BY user_id;

-- 맞는 사고방식
SELECT user_id, sum(amount)
FROM orders
GROUP BY user_id
HAVING sum(amount) >= 150000;
```

두 번째로 중요한 함정은 SELECT alias입니다. 다음 쿼리는 자연스럽게 보이지만 `WHERE`에서 `total_amount`를 쓸 수 없습니다.

```sql
SELECT user_id, amount * 2 AS doubled_amount
FROM orders
WHERE doubled_amount >= 200000;
```

논리적으로 `WHERE`가 먼저 행을 줄이고, 그 뒤 `SELECT`가 `doubled_amount`라는 출력 이름을 만들기 때문입니다. 따라서 `WHERE` 단계에서는 아직 `doubled_amount`라는 이름이 없습니다. 이 경우 식을 반복하거나, 서브쿼리로 한 단계 결과를 만든 뒤 바깥에서 필터링합니다.

```sql
SELECT user_id, amount * 2 AS doubled_amount
FROM orders
WHERE amount * 2 >= 200000;

SELECT *
FROM (
    SELECT user_id, amount * 2 AS doubled_amount
    FROM orders
) AS x
WHERE doubled_amount >= 200000;
```

반대로 `ORDER BY doubled_amount`는 허용되는 경우가 많습니다. 정렬은 출력식이 만들어진 뒤 결과 순서를 정하는 단계이기 때문입니다.

```sql
SELECT user_id, amount * 2 AS doubled_amount
FROM orders
ORDER BY doubled_amount DESC;
```

이 차이를 외우는 가장 좋은 방법은 단계별로 이름이 언제 생기는지 보는 것입니다.

```text
FROM/WHERE 단계:
  볼 수 있는 이름: orders의 컬럼 이름
  아직 없는 이름: SELECT에서 붙일 alias

SELECT 단계:
  doubled_amount라는 출력 이름이 생김

ORDER BY 단계:
  출력 alias를 참조할 수 있음
```

`GROUP BY`가 들어오면 더 조심해야 합니다. group 이후에는 개별 row가 그대로 남아 있지 않습니다. `user_id=10` 그룹은 row 1과 row 2를 하나로 묶은 상태입니다. 이때 `amount` 하나를 그냥 출력하면 어떤 row의 amount를 보여 줘야 할지 모호합니다. 그래서 group query의 SELECT 목록에는 group key나 aggregate expression이 와야 합니다.

```sql
-- 모호한 쿼리
SELECT user_id, amount
FROM orders
GROUP BY user_id;

-- 의미가 닫힌 쿼리
SELECT user_id, sum(amount) AS total_amount
FROM orders
GROUP BY user_id;
```

손으로 보면 왜 모호한지 바로 보입니다.

```text
group user_id=10:
  row 1 amount=120000
  row 2 amount=80000

SELECT user_id, amount
  user_id는 10으로 하나다.
  amount는 120000인가 80000인가?
```

일부 DB 모드에서는 이런 쿼리를 허용하고 임의 행의 값을 보여 줄 수 있습니다. 하지만 학습 기준으로는 위험합니다. 결과가 데이터 배치, 실행 계획, DB 설정에 따라 달라질 수 있기 때문입니다. 그룹화된 relation에서 출력하려는 값이 그룹마다 하나로 정해지는지 확인해야 합니다.

`JOIN`이 있는 SELECT도 같은 순서로 읽습니다.

```sql
SELECT u.user_id, u.email, count(o.order_id) AS ready_orders
FROM users u
LEFT JOIN orders o
  ON o.user_id = u.user_id
 AND o.status = 'READY'
GROUP BY u.user_id, u.email
HAVING count(o.order_id) >= 1
ORDER BY ready_orders DESC;
```

논리적으로는 `FROM` 절에서 `users`와 `orders`의 join 결과를 만듭니다. 이때 `ON` 조건은 join 자체의 매칭 조건입니다. `LEFT JOIN`이므로 매칭되는 주문이 없어도 사용자 row는 남고, 주문 쪽 컬럼은 NULL로 채워질 수 있습니다. 그 뒤 `GROUP BY`로 사용자별로 묶고, `HAVING`으로 READY 주문이 하나 이상인 사용자만 남깁니다. 만약 `o.status='READY'`를 `WHERE`로 옮기면 주문이 없는 사용자 row는 WHERE에서 사라져 LEFT JOIN 의미가 달라집니다.

```text
LEFT JOIN 조건 위치 trace

users:
  u10, u20

orders:
  o1001(user_id=10, READY)

조건을 ON에 둔 경우:
  u10 + o1001
  u20 + NULL
  -> 사용자는 둘 다 join 결과에 남음

조건을 WHERE o.status='READY'에 둔 경우:
  u10 + o1001 -> TRUE
  u20 + NULL  -> UNKNOWN/TRUE 아님
  -> u20이 사라짐
```

이 trace는 SELECT 논리 순서가 단순 시험 암기가 아니라는 점을 보여 줍니다. 조건을 어느 단계에 두느냐가 결과 의미를 바꿉니다. 특히 outer join, NULL, aggregate가 섞이면 조건 위치가 장애로 이어집니다. 운영에서 “LEFT JOIN인데 누락된 사용자도 보여야 하는 화면에서 일부 사용자가 사라졌다”는 문제를 만나면, 먼저 오른쪽 테이블 조건이 WHERE에 들어갔는지 봐야 합니다.

`DISTINCT`는 SELECT 출력 뒤 중복을 제거하는 단계로 읽으면 됩니다.

```sql
SELECT DISTINCT user_id
FROM orders
WHERE status = 'READY'
ORDER BY user_id;
```

논리 trace는 간단합니다.

```text
FROM orders -> row 1,2,3,4,5
WHERE READY -> row 1,2,4,5
SELECT user_id -> [10,10,20,30]
DISTINCT -> [10,20,30]
ORDER BY -> [10,20,30]
```

`DISTINCT`를 너무 늦게 떠올리면 join 증폭을 가리는 용도로 쓰게 됩니다. 예를 들어 사용자 목록이 중복되어 나오니 `SELECT DISTINCT u.*`를 붙이는 식입니다. 이 방법이 결과 화면의 중복을 없앨 수는 있지만, 왜 join이 중복을 만들었는지 설명하지 못하면 집계나 pagination에서 다시 문제가 생깁니다. SELECT 논리 순서를 손으로 따라가면 중복이 어디서 생겼는지 볼 수 있습니다. FROM/JOIN 단계에서 tuple 쌍이 늘어난 것인지, SELECT 단계에서 일부 컬럼만 뽑으며 서로 다른 tuple이 같은 모양으로 접힌 것인지, DISTINCT 단계에서 비용을 들여 제거한 것인지 구분합니다.

`ORDER BY`와 `LIMIT`은 결과 표현에 강한 영향을 줍니다. `LIMIT`만 있고 `ORDER BY`가 없으면 “앞의 10개”라는 말이 안정적인 업무 의미를 갖기 어렵습니다. DB는 relation을 본질적으로 순서 없는 결과로 다루며, 물리 저장 순서나 실행 계획에 따라 반환 순서가 바뀔 수 있습니다. 따라서 최신 주문 10개, 가장 큰 금액 10개, 오래된 장애 10개처럼 업무적으로 “앞”이 중요하면 반드시 `ORDER BY`로 순서를 명시해야 합니다.

```sql
-- 불안정한 의도
SELECT order_id
FROM orders
LIMIT 10;

-- 의미가 닫힌 의도
SELECT order_id
FROM orders
ORDER BY created_at DESC, order_id DESC
LIMIT 10;
```

두 번째 쿼리에서 `order_id DESC`를 함께 둔 것도 중요합니다. `created_at`이 같은 주문이 여러 개면 그 안의 순서가 흔들릴 수 있습니다. tie-breaker, 즉 동률일 때 순서를 결정할 추가 기준이 있어야 pagination이 안정됩니다. 그렇지 않으면 다음 페이지를 볼 때 어떤 주문이 중복되거나 빠질 수 있습니다.

```text
created_at이 같은 주문

order_id=1001 created_at=2026-05-18 10:00
order_id=1002 created_at=2026-05-18 10:00

ORDER BY created_at DESC만 있으면:
  1001,1002 또는 1002,1001 모두 가능

ORDER BY created_at DESC, order_id DESC:
  1002,1001로 고정
```

SELECT 논리 순서를 이해하면 `COUNT(*)`, `COUNT(column)`, `COUNT(DISTINCT column)`의 차이도 명확해집니다. `COUNT(*)`는 group 안의 row 수를 셉니다. `COUNT(column)`은 해당 column이 NULL이 아닌 row 수를 셉니다. `COUNT(DISTINCT column)`은 NULL이 아닌 값의 중복을 제거한 뒤 개수를 셉니다. 이 차이는 NULL 단원에서 더 깊게 다루지만, SELECT 순서에서도 이미 중요합니다.

```text
group user_id=10

row1 coupon_code = 'A'
row2 coupon_code = NULL
row3 coupon_code = 'A'

COUNT(*) = 3
COUNT(coupon_code) = 2
COUNT(DISTINCT coupon_code) = 1
```

이런 차이를 모르면 쿠폰 사용 사용자 수, 주문 수, 주문당 라인 수, 결제 성공률 같은 지표가 틀어집니다. 지표 쿼리는 결과가 숫자 하나라서 더 위험합니다. 숫자가 그럴듯하면 오류를 알아차리기 어렵습니다. SELECT 논리 trace를 작은 fixture로 만든 뒤 기대값을 손으로 계산하는 습관이 필요합니다.

실무에서 SELECT 순서를 검증하는 가장 좋은 방법은 작은 fixture를 직접 만드는 것입니다. production 데이터에서 바로 복잡한 쿼리를 돌리기 전에 5~10행짜리 임시 테이블을 만들고, 각 단계에서 중간 결과를 따로 확인합니다.

```sql
-- WHERE 결과만 확인
SELECT *
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18';

-- GROUP 결과 확인
SELECT user_id, count(*), sum(amount)
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id;

-- HAVING 추가
SELECT user_id, count(*), sum(amount)
FROM orders
WHERE status = 'READY'
  AND created_at = DATE '2026-05-18'
GROUP BY user_id
HAVING sum(amount) >= 150000;
```

PASS 기준은 각 단계의 row 수와 group 수를 손으로 예측할 수 있는 것입니다. FAIL 기준은 최종 쿼리만 보고 결과가 맞는지 감으로 판단하는 것입니다. 복잡한 쿼리일수록 중간 relation을 확인해야 합니다. CTE를 일시적으로 사용해 단계별 이름을 붙이는 것도 좋은 방법입니다.

```sql
WITH filtered AS (
    SELECT *
    FROM orders
    WHERE status = 'READY'
      AND created_at = DATE '2026-05-18'
),
grouped AS (
    SELECT user_id, count(*) AS order_count, sum(amount) AS total_amount
    FROM filtered
    GROUP BY user_id
)
SELECT *
FROM grouped
WHERE total_amount >= 150000
ORDER BY total_amount DESC;
```

이 CTE는 설명과 디버깅에는 좋습니다. 다만 물리 실행에서 CTE가 최적화 경계가 되는지 여부는 DB와 버전에 따라 다를 수 있으므로 성능까지 판단하려면 `EXPLAIN`을 봐야 합니다. 여기서 중요한 것은 논리 순서를 사람에게 보이게 만드는 것입니다.

SELECT 논리 순서의 흔한 장애 함정을 정리하면 다음과 같습니다.

| 함정 | 왜 생기는가 | 증상 | 확인 방법 |
|---|---|---|---|
| WHERE에서 SELECT alias 사용 | alias가 SELECT 단계에서 생김 | 컬럼 없음 오류 또는 DB별 비호환 | 식 반복 또는 subquery로 분리 |
| WHERE와 HAVING 혼동 | WHERE는 row, HAVING은 group 조건 | 집계 조건 오류 | 작은 group trace |
| LEFT JOIN 조건을 WHERE에 둠 | NULL-extended row가 WHERE에서 제거 | 누락되어야 하지 않는 왼쪽 row 사라짐 | 조건을 ON에 두었을 때와 비교 |
| GROUP BY 없이 비집계 컬럼 출력 | group 안 여러 값 중 하나를 고를 수 없음 | 오류 또는 임의 값 | group마다 값이 하나인지 확인 |
| ORDER BY 없는 LIMIT | relation 결과 순서를 보장하지 않음 | pagination 흔들림 | tie-breaker 포함 ORDER BY |
| DISTINCT로 join 중복을 가림 | FROM/JOIN 단계의 증폭 원인을 숨김 | count/page/성능 문제 재발 | join cardinality 확인 |

이 표는 암기용이 아니라 디버깅 순서입니다. 쿼리가 이상하면 먼저 FROM/JOIN에서 후보 tuple이 어떻게 만들어졌는지 봅니다. 그다음 WHERE에서 어떤 row가 사라졌는지 봅니다. group이 있다면 group당 어떤 row가 묶였는지 봅니다. HAVING이 그룹을 버렸는지 봅니다. SELECT에서 어떤 이름과 값이 만들어졌는지 봅니다. DISTINCT가 중복을 제거했는지 봅니다. ORDER/LIMIT이 최종 표현을 어떻게 바꿨는지 봅니다. 이 순서로 보면 복잡한 쿼리도 작은 관계 변환들의 연결로 분해됩니다.

SELECT 문은 논리적으로 읽어야 하지만, 실제 DB가 반드시 그 순서로 실행한다는 뜻은 아닙니다. `WHERE` 조건은 인덱스 탐색으로 밀려 들어갈 수 있고, join 순서는 바뀔 수 있고, grouping 방식은 hash나 sort로 달라질 수 있습니다. 이 내용은 DU02에서 본 논리/물리 분리와 이어집니다. 여기서 배우는 순서는 “결과 의미의 순서”입니다. 실행 계획은 그 의미를 보존하면서 비용을 줄이려고 다른 물리 순서를 선택할 수 있습니다. 따라서 SELECT를 공부할 때는 두 문장을 함께 기억합니다. 논리 순서는 결과를 검증하는 기준이고, 실행 계획은 성능을 검증하는 관측 자료입니다.

마지막 replay path는 다음과 같습니다.

```text
1. 쿼리에서 FROM/JOIN만 남기고 후보 row 수를 적는다.
2. WHERE 조건을 하나씩 적용하며 사라지는 row를 표시한다.
3. GROUP BY가 있으면 group마다 포함된 row id를 적는다.
4. aggregate 값을 직접 계산한다.
5. HAVING이 버리는 group을 표시한다.
6. SELECT 출력 이름이 언제 생기는지 표시한다.
7. DISTINCT, ORDER BY, LIMIT이 최종 결과를 어떻게 바꾸는지 적는다.
```

이 replay를 5행짜리 데이터로 할 수 있으면 큰 production 쿼리도 무서움이 줄어듭니다. 쿼리는 길어도 결국 relation을 만들고 줄이고 묶고 출력하는 흐름입니다. SELECT alias 함정, WHERE/HAVING 혼동, LEFT JOIN 누락, DISTINCT 남용은 모두 이 흐름의 어느 단계를 섞었기 때문에 생깁니다. SQL을 잘 읽는다는 것은 문법을 많이 아는 것보다, 각 절이 어느 논리 단계에서 어떤 relation을 입력으로 받아 어떤 relation을 출력하는지 설명할 수 있다는 뜻입니다.

SELECT 논리 순서를 더 깊게 보려면 window function도 좋은 예입니다. window function은 `GROUP BY`처럼 행을 하나로 합치지 않고, 현재 결과 행들 위에 창을 잡아 계산한 값을 붙입니다.

```sql
SELECT order_id,
       user_id,
       amount,
       sum(amount) OVER (PARTITION BY user_id ORDER BY order_id) AS running_amount
FROM orders
WHERE status = 'READY'
ORDER BY user_id, order_id;
```

이 쿼리는 `WHERE`로 READY 주문만 남긴 뒤, 사용자별로 order_id 순서를 잡고 누적 합계를 계산합니다. 결과 행 수는 줄어들지 않습니다. `GROUP BY user_id`가 사용자별 한 행으로 접는 것과 다릅니다.

```text
READY rows:
  (order_id=1, user_id=10, amount=120000)
  (order_id=2, user_id=10, amount=80000)
  (order_id=5, user_id=30, amount=150000)

window PARTITION BY user_id ORDER BY order_id:
  user_id=10:
    row1 running=120000
    row2 running=200000
  user_id=30:
    row5 running=150000

결과 row 수:
  입력 READY row 3개 -> 출력 row 3개
```

window function을 집계와 섞어 이해하면 장애가 생깁니다. `GROUP BY`는 여러 row를 한 group 결과로 접습니다. window function은 row를 유지하면서 주변 row를 참고해 계산합니다. 그래서 “사용자별 총액만 필요하다”면 GROUP BY가 맞고, “각 주문 행을 유지하면서 사용자별 누적 총액을 붙이고 싶다”면 window가 맞습니다. 논리 순서상 window 계산은 WHERE와 GROUP/HAVING 이후, SELECT 출력과 ORDER BY 근처의 의미로 이해하는 편이 좋습니다. DB별 세부 문법과 처리 순서는 공식 문서를 확인해야 하지만, 핵심은 row를 접는지 유지하는지입니다.

서브쿼리는 논리 단계에 이름을 붙이는 도구로 사용할 수 있습니다. 특히 SELECT alias를 WHERE에서 쓰고 싶을 때 유용합니다.

```sql
SELECT *
FROM (
    SELECT order_id,
           amount,
           amount * 2 AS doubled_amount
    FROM orders
) AS priced
WHERE doubled_amount >= 200000;
```

손으로 보면 내부 SELECT가 먼저 `priced`라는 relation을 만듭니다.

```text
priced
+----------+--------+----------------+
| order_id | amount | doubled_amount |
+----------+--------+----------------+
| 1        | 120000 | 240000         |
| 2        |  80000 | 160000         |
| 3        | 200000 | 400000         |
+----------+--------+----------------+

바깥 WHERE doubled_amount >= 200000
  -> order_id 1, 3
```

이렇게 한 단계 relation을 만든 뒤 바깥에서 필터링하면 alias가 실제 입력 컬럼처럼 존재합니다. 단, 물리 실행에서 DB가 내부 relation을 실제로 materialize하는지는 별도 문제입니다. 논리적으로는 이름 붙은 중간 relation이고, 물리적으로는 optimizer가 합칠 수도 있습니다. 논리 의미를 설명하려고 subquery를 쓰는 것과 성능을 위해 materialization을 기대하는 것은 다른 주장입니다.

`CASE` expression도 SELECT 단계의 출력 계산을 이해하는 데 좋습니다.

```sql
SELECT order_id,
       amount,
       CASE
           WHEN amount >= 100000 THEN 'HIGH'
           ELSE 'NORMAL'
       END AS amount_band
FROM orders
WHERE status = 'READY';
```

논리적으로 `WHERE`가 READY 주문만 남긴 뒤, SELECT 목록에서 각 남은 row에 대해 `amount_band`를 계산합니다. `amount_band`로 다시 필터링하려면 식을 반복하거나 subquery가 필요합니다.

```sql
SELECT *
FROM (
    SELECT order_id,
           amount,
           CASE
               WHEN amount >= 100000 THEN 'HIGH'
               ELSE 'NORMAL'
           END AS amount_band
    FROM orders
    WHERE status = 'READY'
) AS banded
WHERE amount_band = 'HIGH';
```

이 구조는 보고서 쿼리에서 자주 나옵니다. 상태 분류, 금액 구간, 날짜 bucket 같은 계산 컬럼을 만들고, 그 계산 결과로 다시 필터링하거나 grouping하고 싶을 때 논리 단계를 분리하면 의도가 선명해집니다. 다만 같은 식을 여러 곳에 반복하면 실수 가능성이 생기므로, view, generated column, CTE, application-level constant 등으로 정책을 어디에 둘지 검토해야 합니다.

SELECT 논리 순서는 권한과 보안에서도 중요합니다. `SELECT *`를 습관적으로 쓰면 relation의 모든 컬럼을 출력 단계로 끌어옵니다. 지금은 문제가 없어 보여도 나중에 `users` 테이블에 `password_hash`, `ssn`, `internal_note` 같은 컬럼이 추가되면 같은 쿼리가 더 많은 민감 정보를 반환할 수 있습니다. 논리적으로 SELECT 목록은 결과 relation의 schema를 정하는 단계입니다. 따라서 필요한 컬럼을 명시하는 것은 성능뿐 아니라 정보 노출 경계입니다.

```sql
-- 약한 형태
SELECT *
FROM users
WHERE user_id = 10;

-- 더 강한 형태
SELECT user_id, email, display_name
FROM users
WHERE user_id = 10;
```

이 차이는 ORM을 쓰더라도 사라지지 않습니다. 엔티티 전체를 로딩한 뒤 JSON serializer가 노출할 필드를 고르는 구조라면 DB SELECT와 API 출력 경계가 멀어집니다. 필요한 relation 모양을 SQL이나 projection에서 먼저 고정하면 downstream으로 넘어가는 데이터가 줄고, 의도하지 않은 컬럼 노출 위험도 줄어듭니다.

정렬과 collation도 SELECT 순서의 일부입니다. `ORDER BY name`은 단순히 문자열 바이트 순서로 정렬하라는 뜻이 아닐 수 있습니다. 컬럼의 collation이 대소문자, 악센트, 언어별 규칙을 반영하면 결과 순서가 달라집니다. 따라서 SELECT의 마지막 표현 단계에서 정렬 결과가 업무 기대와 맞는지 확인해야 합니다. 예를 들어 고객명 검색 화면에서는 대소문자 무시가 자연스러울 수 있지만, 코드값이나 암호화된 token 정렬에서는 binary에 가까운 비교가 필요할 수 있습니다. 이 내용은 DU06에서 자세히 다루지만, SELECT 순서에서도 `ORDER BY`가 어떤 비교 규칙을 쓰는지 묻는 습관이 필요합니다.

```text
ORDER BY 검증 질문

1. 어떤 컬럼 또는 expression으로 정렬하는가?
2. 같은 값일 때 tie-breaker가 있는가?
3. 문자열이면 collation이 업무 기대와 맞는가?
4. LIMIT/OFFSET이나 pagination이 이 순서에 의존하는가?
5. 인덱스가 이 순서를 도와주는가, 아니면 별도 sort가 필요한가?
```

SELECT 논리 순서를 운영 검증으로 연결하면 slow query 분석이 더 선명해집니다. 느린 쿼리를 보면 먼저 어느 논리 단계에서 row 수가 줄어들어야 하는지 예측합니다. WHERE에서 1%만 남아야 하는데 실제 계획에서 대부분을 읽는다면 조건식, 인덱스, 통계를 봅니다. GROUP BY에서 그룹 수가 작아져야 하는데 메모리 사용이 크다면 grouping key와 입력 row 수를 봅니다. ORDER BY/LIMIT에서 상위 20개만 필요하지만 전체 sort가 일어난다면 정렬을 지원하는 인덱스나 keyset pagination을 봅니다. 이렇게 논리 단계별 기대 row 수를 잡으면 실행 계획의 문제 지점도 찾기 쉽습니다.

```text
느린 SELECT 분석 trace

논리 기대:
  FROM orders: 10,000,000 rows
  WHERE tenant_id=7 AND created_at today: 20,000 rows
  GROUP BY user_id: 3,000 groups
  ORDER BY total DESC LIMIT 20: 20 rows

관측:
  Seq Scan reads 10,000,000 rows
  Filter removes 9,980,000 rows
  Sort spills to disk

해석:
  WHERE 조건을 빨리 좁히지 못하고 있다.
  tenant/date 인덱스, 통계, 조건식 타입 변환을 확인한다.
  ORDER BY LIMIT은 최종 결과가 작아도 중간 sort 비용이 클 수 있다.
```

SELECT 순서를 알면 쿼리 리뷰도 좋아집니다. 리뷰어는 “문법 맞음”만 보지 않고, 각 절이 담당하는 의미를 묻습니다. FROM/JOIN이 의도한 cardinality를 만드는가. WHERE가 row 조건만 담는가. GROUP BY 기준이 업무 단위와 맞는가. HAVING은 집계 조건인가. SELECT 목록이 필요한 정보만 노출하는가. DISTINCT가 원인 은폐는 아닌가. ORDER BY가 안정적인가. LIMIT이 업무 의미를 갖는가. 이런 질문은 코드 스타일보다 훨씬 큰 버그를 잡습니다.

```text
SELECT 리뷰 질문

FROM/JOIN:
  왼쪽 한 행에 오른쪽 몇 행이 붙을 수 있는가?
  outer join 조건 위치가 의도와 맞는가?

WHERE:
  개별 row 조건인가?
  NULL이 섞이면 TRUE/FALSE/UNKNOWN 결과가 의도와 맞는가?

GROUP/HAVING:
  group key가 업무 단위와 맞는가?
  비집계 컬럼을 임의로 뽑고 있지 않은가?

SELECT:
  alias가 어디서부터 유효한가?
  필요한 컬럼만 출력하는가?

ORDER/LIMIT:
  정렬 기준과 tie-breaker가 안정적인가?
  LIMIT이 임의 순서의 앞부분을 의미하지 않는가?
```

마지막으로 SELECT 순서는 학습용으로는 고정된 길처럼 보이지만, 실제 SQL dialect마다 세부 차이가 있습니다. PostgreSQL, MySQL, Oracle, SQL Server는 alias 허용 위치, GROUP BY 엄격성, CTE 최적화, NULL 정렬 기본값, LIMIT/OFFSET 문법, window function 세부에서 차이를 보입니다. 그래서 이 섹션은 특정 DBMS의 모든 문법을 확정하는 문서가 아니라, SQL 결과 의미를 읽는 공통 뼈대를 세우는 문서입니다. 실제 현업에서는 사용하는 DB의 공식 문서와 `EXPLAIN`으로 다시 닫아야 합니다.

이 논리 순서가 필요해진 배경은 SQL이 절차형 반복문이 아니라 선언형 질의 언어라는 점에 있습니다. 사용자는 “orders 배열을 앞에서부터 돌며 if를 실행하고 map에 누적하라”고 명령하지 않습니다. 대신 어떤 relation에서 어떤 조건의 row를 남기고 어떤 단위로 묶어 어떤 결과를 원한다고 말합니다. DB는 그 의미를 실제 실행 계획으로 바꿉니다. 그래서 사람에게는 결과 의미를 복원하는 논리 순서가 필요하고, DB에게는 비용을 줄이는 물리 실행 계획이 필요합니다.

이 섹션의 기억 문장은 이렇게 둘 수 있습니다. `SELECT`는 문장 첫 단어이지만 논리 처리의 첫 단계가 아닙니다. SQL은 후보 relation을 만들고, 줄이고, 묶고, 계산하고, 출력하고, 정렬하는 흐름으로 읽어야 합니다. 각 절이 어느 단계의 relation을 입력으로 받는지 알면 alias, HAVING, LEFT JOIN, DISTINCT, LIMIT의 함정을 스스로 설명할 수 있습니다.

SELECT 순서를 실제로 체득하려면 같은 요구를 두세 가지 형태로 바꿔 보는 연습이 좋습니다. 예를 들어 “사용자별 READY 주문 총액이 15만 원 이상인 사용자”를 구한다고 합시다. 첫 번째 형태는 GROUP BY와 HAVING을 직접 씁니다.

```sql
SELECT user_id, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
GROUP BY user_id
HAVING sum(amount) >= 150000;
```

두 번째 형태는 단계별 CTE로 씁니다.

```sql
WITH ready_orders AS (
    SELECT user_id, amount
    FROM orders
    WHERE status = 'READY'
),
user_totals AS (
    SELECT user_id, sum(amount) AS total_amount
    FROM ready_orders
    GROUP BY user_id
)
SELECT user_id, total_amount
FROM user_totals
WHERE total_amount >= 150000;
```

두 쿼리는 논리적으로 같은 결과를 만들 수 있습니다. 첫 번째는 짧고, 두 번째는 중간 relation을 사람에게 보여 줍니다. 학습과 리뷰 단계에서는 두 번째 형태가 이해를 돕고, 최종 코드에서는 첫 번째 형태가 충분할 수 있습니다. 다만 실제 실행 계획은 DB가 어떻게 최적화하느냐에 따라 달라지므로 성능 판단은 `EXPLAIN`으로 해야 합니다. 이 연습의 목적은 특정 문법을 선호하는 것이 아니라, 논리 단계가 어떻게 중간 relation으로 표현되는지 보는 것입니다.

```text
ready_orders:
  user_id=10 amount=120000
  user_id=10 amount=80000
  user_id=20 amount=30000
  user_id=30 amount=150000

user_totals:
  user_id=10 total=200000
  user_id=20 total=30000
  user_id=30 total=150000

final WHERE total >= 150000:
  user_id=10 total=200000
  user_id=30 total=150000
```

이 trace는 `HAVING`과 바깥 `WHERE`가 왜 비슷해 보이는지도 설명합니다. 직접 GROUP BY 쿼리에서는 그룹 결과를 거르는 조건이 `HAVING`입니다. CTE로 그룹 결과를 named relation으로 만든 뒤에는 그 relation의 row를 거르는 조건이 바깥 `WHERE`가 됩니다. 즉 `HAVING`은 마법 절이 아니라 “아직 같은 SELECT 안에서 group 결과를 거르는 위치”입니다. group 결과를 별도 relation으로 만든 뒤에는 일반 WHERE로 다시 읽을 수 있습니다.

또 다른 연습은 잘못된 쿼리를 단계별로 고치는 것입니다.

```sql
-- 요구: READY 주문 총액이 15만 원 이상인 사용자
-- 잘못된 쿼리
SELECT user_id, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
  AND total_amount >= 150000
GROUP BY user_id;
```

수정은 “alias를 못 쓰니 이름만 바꾸자”가 아닙니다. 문제는 `total_amount`가 SELECT 단계에서 생기는 출력 이름이고, `WHERE`는 group 전 row 단계라는 점입니다. 따라서 집계 조건을 HAVING으로 옮겨야 합니다.

```sql
SELECT user_id, sum(amount) AS total_amount
FROM orders
WHERE status = 'READY'
GROUP BY user_id
HAVING sum(amount) >= 150000;
```

이렇게 고치면 논리 단계가 맞아집니다. `WHERE`는 READY row만 남기고, `GROUP BY`가 사용자별로 묶고, `HAVING`이 총액 15만 원 이상 group만 남깁니다. 단순히 에러를 없앤 것이 아니라 조건이 적용될 세계를 바로잡은 것입니다.

SELECT 논리 순서를 익히는 마지막 관측 지점은 row count입니다. 복잡한 쿼리를 최종 결과만 보면 어디서 행이 늘고 줄었는지 모릅니다. 중간 단계별 count를 찍으면 많은 오류가 보입니다.

```sql
SELECT count(*) FROM orders;

SELECT count(*)
FROM orders
WHERE status = 'READY';

SELECT user_id, count(*)
FROM orders
WHERE status = 'READY'
GROUP BY user_id;
```

장애 조사에서는 이 count들이 작은 계단처럼 작동합니다. 전체 주문이 1,000만 건이고 READY가 20만 건이어야 하는데 WHERE 뒤 count가 900만 건이면 status 값이나 조건이 잘못됐을 수 있습니다. 사용자별 group이 3,000개여야 하는데 300만 개라면 group key가 너무 세분화됐거나 join이 증폭됐을 수 있습니다. 최종 결과 20개만 보고는 이런 문제를 찾기 어렵습니다.

```text
단계별 row count 계단

FROM orders                         10,000,000
WHERE status='READY'                   200,000
JOIN users                             200,000
GROUP BY user_id                         3,000
HAVING sum(amount)>=150000                 120
ORDER BY total DESC LIMIT 20                20
```

이 계단에서 갑자기 예상과 다른 변화가 생기는 지점이 디버깅 시작점입니다. SELECT 논리 순서는 그래서 사람이 읽는 실행 일지이기도 합니다. DB가 실제로는 다른 물리 순서로 실행하더라도, 의미 검증과 장애 가설은 이 논리 계단을 기준으로 세울 수 있습니다.

최종적으로 SELECT 문을 읽을 때는 눈으로 보이는 문장 순서를 잠시 내려놓고, relation이 어떻게 변하는지 그립니다. `FROM`은 출발 relation을 만들고, `JOIN`은 tuple을 결합하며, `WHERE`는 row를 버리고, `GROUP BY`는 여러 row를 group으로 접고, `HAVING`은 group을 버리고, `SELECT`는 출력 모양과 이름을 만들고, `DISTINCT`는 중복을 제거하고, `ORDER BY`와 `LIMIT`은 표시 순서와 개수를 정합니다. 이 흐름을 한 줄씩 손으로 적으면 alias와 aggregate와 outer join이 더 이상 예외 규칙처럼 보이지 않습니다. 각 절이 자기 단계에서 볼 수 있는 이름과 값만 볼 수 있다는 원칙으로 설명됩니다.

실무에서는 이 원칙을 작은 쿼리 리뷰 습관으로 바꾸면 좋습니다. PR에서 복잡한 SELECT가 보이면 “결과가 맞나요?”라고만 묻지 말고 “WHERE 뒤에는 몇 행이 남나요?”, “GROUP BY 뒤에는 무엇이 한 행인가요?”, “이 alias는 어느 단계에서 생기나요?”, “ORDER BY가 없는 LIMIT이 안정적인가요?”라고 묻습니다. 이 질문들은 작성자를 공격하는 말이 아니라, 쿼리의 논리 단계를 함께 재생하는 도구입니다.

그 재생이 가능해지면 SQL은 암기 과목이 아니라 작은 관계 변환들의 연결로 보입니다. 이후 실행 계획을 보더라도 먼저 이 논리 계단을 기준으로 의미가 보존되는지 확인한 뒤, scan과 index와 sort 비용을 따로 평가할 수 있습니다.

### 마지막 실무 연결: 논리 순서는 디버깅의 지도다

SELECT 논리 처리 순서는 optimizer가 실제로 그 순서대로 실행한다는 뜻이 아니다. 하지만 디버깅할 때는 매우 좋은 지도다. `WHERE`에서 사라진 row는 `GROUP BY` 뒤에는 돌아오지 않고, `HAVING`은 이미 묶인 그룹을 걸러내며, `ORDER BY`는 결과를 정렬할 뿐 새로운 사실을 만들지 않는다. 이 지도를 갖고 있으면 alias가 왜 특정 위치에서 안 보이는지, 집계 전에 필터링해야 하는지, 집계 후에 필터링해야 하는지 설명할 수 있다.

실무에서는 잘못된 위치의 조건 하나가 성능과 의미를 동시에 망친다. 집계 뒤에 걸러야 할 조건을 WHERE로 내리면 결과가 달라지고, WHERE로 내릴 수 있는 조건을 HAVING에 남기면 불필요한 row를 끝까지 끌고 간다. 논리 순서는 SQL을 외우기 위한 표가 아니라 조건을 어디에 두어야 하는지 판단하는 도구다.
