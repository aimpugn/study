# SQL semantics, types, NULL, collation, and binding

## NULL과 3-valued logic

SQL의 `NULL`을 처음 볼 때 가장 흔한 오해는 `NULL`을 “빈 문자열”, “0”, “false”, “없음이라는 값” 중 하나로 생각하는 것입니다. 이 오해가 생기면 `WHERE`, `JOIN`, `NOT IN`, `COUNT`, unique 제약, outer join 결과를 읽을 때 계속 틀립니다. SQL에서 `NULL`은 일반 값과 같은 방식으로 비교되지 않습니다. PostgreSQL 공식 비교 함수 문서는 일반 비교 연산자가 NULL 입력을 받으면 NULL, 즉 알 수 없음에 해당하는 결과를 낼 수 있음을 설명하고, `IS NULL`, `IS NOT NULL`, `IS DISTINCT FROM` 같은 별도 판단 연산을 제공합니다. MySQL 공식 비교 연산자 문서도 `<=>` 같은 NULL-safe equality와 일반 비교의 차이를 따로 설명합니다. 이 말은 `NULL`이 그냥 값 하나였다면 필요 없었을 장치가 SQL에는 있다는 뜻입니다.

가장 작은 테이블부터 시작합니다.

```sql
CREATE TABLE users (
    user_id integer PRIMARY KEY,
    email text,
    marketing_agreed boolean
);

INSERT INTO users VALUES
  (1, 'a@example.com', true),
  (2, 'b@example.com', false),
  (3, NULL,            NULL);
```

이 데이터에서 `marketing_agreed`는 세 상태를 가질 수 있습니다. `true`는 동의했습니다. `false`는 동의하지 않았습니다. `NULL`은 아직 알 수 없거나 수집하지 않았거나 적용 대상이 아닐 수 있습니다. 이 셋은 서로 다릅니다. 특히 `NULL`은 `false`가 아닙니다. 동의하지 않았다는 사실과 아직 모른다는 상태는 정책적으로도 다릅니다.

```text
marketing_agreed 상태

true   -> 동의함
false  -> 동의하지 않음
NULL   -> 모름/미수집/해당 없음, 구체 의미는 schema와 정책이 정해야 함
```

SQL의 비교는 TRUE/FALSE 두 값만으로 끝나지 않습니다. `NULL`이 끼면 UNKNOWN, 즉 참인지 거짓인지 알 수 없는 결과가 나옵니다. 그래서 SQL의 논리는 3-valued logic, 한국어로는 3값 논리라고 부릅니다. 이 용어는 어렵게 보이지만 핵심은 간단합니다. `WHERE`는 TRUE인 행만 통과시키고, FALSE와 UNKNOWN은 통과시키지 않습니다.

```sql
SELECT user_id
FROM users
WHERE marketing_agreed = true;
```

```text
row 1: true = true   -> TRUE    -> 남음
row 2: false = true  -> FALSE   -> 버림
row 3: NULL = true   -> UNKNOWN -> 버림
```

`NULL = true`가 FALSE가 아니라 UNKNOWN이라는 점이 중요합니다. row 3은 “동의하지 않았다”라서 버려진 것이 아니라, 동의 여부를 알 수 없기 때문에 TRUE 조건을 만족하지 못해 버려진 것입니다. 결과는 둘 다 안 보이는 것으로 같지만 의미가 다릅니다. 이 차이를 놓치면 나중에 `NOT`을 붙였을 때 더 크게 틀립니다.

```sql
SELECT user_id
FROM users
WHERE NOT (marketing_agreed = true);
```

```text
row 1: NOT TRUE    -> FALSE   -> 버림
row 2: NOT FALSE   -> TRUE    -> 남음
row 3: NOT UNKNOWN -> UNKNOWN -> 버림
```

많은 사람이 `NOT (marketing_agreed = true)`이면 동의하지 않은 사람과 모르는 사람이 모두 나올 것이라고 생각합니다. 하지만 UNKNOWN에 NOT을 붙여도 UNKNOWN입니다. `WHERE`는 TRUE만 통과시키므로 NULL row는 여전히 나오지 않습니다. 동의하지 않았거나 모르는 사람을 모두 찾고 싶다면 조건을 명시해야 합니다.

```sql
SELECT user_id
FROM users
WHERE marketing_agreed = false
   OR marketing_agreed IS NULL;
```

3값 논리를 표로 보면 더 잘 보입니다.

| expression | TRUE | FALSE | UNKNOWN |
|---|---:|---:|---:|
| `NOT expression` | FALSE | TRUE | UNKNOWN |

`AND`와 `OR`도 UNKNOWN을 포함합니다.

| A | B | A AND B | A OR B |
|---|---|---|---|
| TRUE | TRUE | TRUE | TRUE |
| TRUE | FALSE | FALSE | TRUE |
| TRUE | UNKNOWN | UNKNOWN | TRUE |
| FALSE | TRUE | FALSE | TRUE |
| FALSE | FALSE | FALSE | FALSE |
| FALSE | UNKNOWN | FALSE | UNKNOWN |
| UNKNOWN | TRUE | UNKNOWN | TRUE |
| UNKNOWN | FALSE | FALSE | UNKNOWN |
| UNKNOWN | UNKNOWN | UNKNOWN | UNKNOWN |

이 표에서 기억할 기준은 “결과가 확정될 수 있으면 확정되고, 확정할 수 없으면 UNKNOWN으로 남는다”입니다. `FALSE AND UNKNOWN`은 이미 왼쪽이 FALSE라 전체 AND가 FALSE로 확정됩니다. `TRUE OR UNKNOWN`은 이미 왼쪽이 TRUE라 전체 OR가 TRUE로 확정됩니다. 하지만 `TRUE AND UNKNOWN`은 오른쪽이 TRUE인지 FALSE인지 몰라 전체가 UNKNOWN입니다.

실무에서 가장 유명한 함정은 `NOT IN`과 NULL입니다.

```sql
CREATE TABLE blocked_users (
    user_id integer
);

INSERT INTO blocked_users VALUES
  (2),
  (NULL);

SELECT user_id
FROM users
WHERE user_id NOT IN (SELECT user_id FROM blocked_users);
```

많은 사람은 blocked_users에 2번만 있으니 1번과 3번이 나올 것이라고 생각합니다. 하지만 `NOT IN`은 내부적으로 여러 비교의 조합처럼 동작합니다. 오른쪽 목록에 NULL이 있으면 “왼쪽 값이 NULL과 다른가?”라는 질문이 UNKNOWN이 됩니다.

```text
user_id = 1

1 NOT IN (2, NULL)
  == NOT (1 = 2 OR 1 = NULL)
  == NOT (FALSE OR UNKNOWN)
  == NOT UNKNOWN
  == UNKNOWN
  -> WHERE 통과 못 함

user_id = 2

2 NOT IN (2, NULL)
  == NOT (TRUE OR UNKNOWN)
  == NOT TRUE
  == FALSE
  -> WHERE 통과 못 함
```

결과가 아예 비어 버릴 수 있습니다. 이것은 SQL 엔진이 이상한 것이 아니라 3값 논리의 결과입니다. 안전한 대안은 NULL을 제거하거나 `NOT EXISTS`를 쓰는 것입니다.

```sql
SELECT u.user_id
FROM users u
WHERE NOT EXISTS (
    SELECT 1
    FROM blocked_users b
    WHERE b.user_id = u.user_id
);
```

`NOT EXISTS`는 “매칭되는 row가 존재하는가”를 묻습니다. blocked_users의 NULL row는 `b.user_id = u.user_id`를 TRUE로 만들지 못하므로 매칭 row가 아닙니다. 따라서 NULL 하나 때문에 전체 결과가 사라지는 문제가 줄어듭니다. 물론 의도에 따라 blocked_users.user_id를 `NOT NULL`로 만드는 것이 더 근본적인 해결일 수 있습니다. blacklist 테이블의 user_id가 NULL이면 의미 있는 차단 대상이 아니므로 schema 제약으로 막는 편이 안전합니다.

```sql
CREATE TABLE blocked_users (
    user_id integer NOT NULL
);
```

NULL은 outer join에서도 자주 나타납니다.

```sql
CREATE TABLE orders (
    order_id integer PRIMARY KEY,
    user_id integer NOT NULL
);

INSERT INTO orders VALUES
  (1001, 1),
  (1002, 1);

SELECT u.user_id, o.order_id
FROM users u
LEFT JOIN orders o ON o.user_id = u.user_id;
```

주문이 없는 사용자는 오른쪽 order 컬럼들이 NULL로 채워진 row를 얻습니다.

```text
u1 + o1001 -> (1, 1001)
u1 + o1002 -> (1, 1002)
u2 + no order -> (2, NULL)
u3 + no order -> (3, NULL)
```

여기서 `WHERE o.order_id IS NULL`을 붙이면 주문이 없는 사용자를 찾는 쿼리가 됩니다.

```sql
SELECT u.user_id
FROM users u
LEFT JOIN orders o ON o.user_id = u.user_id
WHERE o.order_id IS NULL;
```

하지만 오른쪽 테이블 조건을 잘못 WHERE에 두면 LEFT JOIN이 사실상 INNER JOIN처럼 변합니다.

```sql
SELECT u.user_id, o.order_id
FROM users u
LEFT JOIN orders o ON o.user_id = u.user_id
WHERE o.order_id >= 1000;
```

주문이 없는 사용자의 `o.order_id`는 NULL이고, `NULL >= 1000`은 UNKNOWN입니다. WHERE는 TRUE만 통과시키므로 주문이 없는 사용자가 사라집니다. 조건이 join 매칭 조건인지, join 뒤 결과 row 필터인지 구분해야 합니다.

집계에서도 NULL은 특별합니다.

```sql
CREATE TABLE survey (
    user_id integer PRIMARY KEY,
    score integer
);

INSERT INTO survey VALUES
  (1, 5),
  (2, NULL),
  (3, 3);
```

```text
COUNT(*)      = 3  -- row 수
COUNT(score)  = 2  -- score가 NULL이 아닌 row 수
SUM(score)    = 8  -- NULL은 더하지 않음
AVG(score)    = 4  -- NULL 제외한 5,3의 평균
```

이 결과를 보고 “NULL을 0으로 처리했네”라고 생각하면 틀립니다. `AVG(score)`가 `(5+0+3)/3 = 2.66`이 아니라 `(5+3)/2 = 4`인 것을 보면 NULL이 0이 아니라 제외된다는 점이 보입니다. 만약 미응답을 0점으로 보고 싶다면 `COALESCE(score, 0)`처럼 정책을 명시해야 합니다.

```sql
SELECT avg(coalesce(score, 0)) AS avg_treat_null_as_zero
FROM survey;
```

이때도 함정이 있습니다. `COALESCE`는 기술적 처리이지만, NULL을 0으로 보는 정책이 정말 맞는지 먼저 결정해야 합니다. 시험 점수에서 미응시를 0점으로 볼 수도 있지만, 만족도 설문에서 미응답을 0점으로 보면 왜곡이 큽니다. SQL은 NULL 처리 함수를 제공하지만, 의미 결정은 도메인과 분석 목적이 해야 합니다.

NULL 비교에는 `IS NULL`과 `IS NOT NULL`을 씁니다.

```sql
SELECT user_id
FROM users
WHERE email IS NULL;

SELECT user_id
FROM users
WHERE email IS NOT NULL;
```

`email = NULL`은 원하는 결과를 주지 않습니다. `email = NULL`은 email이 NULL인 row에서도 UNKNOWN이기 때문입니다. NULL인지 판단하는 별도 연산이 필요한 이유가 여기에 있습니다.

```text
email = NULL
  row email='a@example.com' -> UNKNOWN
  row email=NULL            -> UNKNOWN

email IS NULL
  row email='a@example.com' -> FALSE
  row email=NULL            -> TRUE
```

PostgreSQL의 `IS DISTINCT FROM` 계열이나 MySQL의 `<=>` 같은 NULL-safe 비교는 “NULL끼리는 같게 보고, NULL과 값은 다르게 보겠다”는 필요에서 나옵니다.

```text
일반 비교:
  NULL = NULL -> UNKNOWN

NULL-safe equality:
  NULL <=> NULL -> TRUE   -- MySQL
  NULL IS NOT DISTINCT FROM NULL -> TRUE -- PostgreSQL식 표현
```

이런 연산은 변경 감지나 upsert, nullable 컬럼 비교에서 유용합니다. 예를 들어 이전 값과 새 값이 모두 NULL이면 “변경 없음”으로 보고 싶은 경우 일반 `=`로는 UNKNOWN이 나와 판단이 흔들립니다. DB별 문법이 다르므로 공식 문서를 확인해야 합니다.

unique constraint와 NULL도 조심해야 합니다. DBMS마다 NULL을 unique에서 어떻게 다루는지 차이가 있습니다. PostgreSQL은 기본 unique constraint에서 NULL 값들이 서로 같지 않은 것으로 취급되어 여러 NULL이 들어갈 수 있습니다. MySQL도 unique index에서 NULL을 여러 개 허용하는 동작이 일반적입니다. 하지만 세부는 버전과 설정, index 옵션에 따라 확인해야 합니다. 중요한 것은 “unique니까 NULL도 하나만 되겠지”라고 가정하지 않는 것입니다. nullable 컬럼의 uniqueness가 업무적으로 중요하면 partial index, generated column, 별도 제약, 애플리케이션 정책을 검토해야 합니다.

```text
nullable email unique 정책 질문

1. email을 모르는 사용자를 여러 명 허용하는가?
2. email이 NULL인 row는 아직 가입 미완료인가?
3. email이 생긴 뒤에는 중복을 막아야 하는가?
4. 빈 문자열과 NULL을 같은 미입력으로 볼 것인가?
5. 대소문자 차이는 같은 email인가?
```

이 질문을 닫지 않으면 SQL 제약이 업무 의미와 어긋납니다. 특히 email은 collation, lower-case normalization, 국제화 주소 정책이 함께 들어오므로 단순 unique만으로 충분하지 않을 수 있습니다.

NULL은 정렬에서도 DB별 기본값이 다를 수 있습니다. 어떤 DB는 오름차순에서 NULL을 먼저 둘 수 있고, 어떤 DB는 나중에 둘 수 있습니다. PostgreSQL은 `NULLS FIRST`와 `NULLS LAST`를 명시할 수 있습니다. 정렬 결과가 UI나 pagination에 중요하면 NULL 위치를 명시하는 편이 안전합니다.

```sql
SELECT user_id, last_login_at
FROM users
ORDER BY last_login_at DESC NULLS LAST, user_id;
```

이 쿼리는 로그인한 사용자를 최신순으로 보여 주고, 로그인 기록이 없는 사용자는 뒤로 보냅니다. 마지막 tie-breaker로 `user_id`를 두어 같은 시각의 순서도 안정화합니다. NULL 위치를 정하지 않으면 DB별 기본값이나 쿼리 변경에 따라 화면 순서가 바뀔 수 있습니다.

NULL을 제대로 다루려면 schema 단계에서 의미를 줄이는 것도 중요합니다. 반드시 있어야 하는 값에는 `NOT NULL`을 둡니다. 알 수 없음이 업무적으로 의미 없는 컬럼에 NULL을 허용하면 모든 쿼리가 불필요하게 복잡해집니다. 반대로 정말 모름/미수집/해당 없음이 의미 있는 값에는 NULL을 허용하되, 그 의미를 문서와 코드에서 명확히 해야 합니다.

```sql
CREATE TABLE user_marketing_preferences (
    user_id integer PRIMARY KEY,
    marketing_agreed boolean,
    decided_at timestamp
);
```

여기서 `marketing_agreed NULL`은 “아직 결정하지 않음”일 수 있습니다. 그 경우 `decided_at`도 NULL일 가능성이 큽니다. 하지만 동의/거부가 반드시 있어야만 row를 만들기로 했다면 `marketing_agreed boolean NOT NULL`이 더 맞을 수 있습니다. 모델링 선택이 쿼리 복잡도를 결정합니다.

운영에서 NULL 관련 장애를 찾는 관측 쿼리는 단순합니다.

```sql
SELECT count(*) AS total,
       count(email) AS email_present,
       count(*) - count(email) AS email_null
FROM users;

SELECT marketing_agreed, count(*)
FROM users
GROUP BY marketing_agreed
ORDER BY marketing_agreed;
```

PASS는 nullable 컬럼마다 NULL 비율과 의미를 설명할 수 있는 상태입니다. FAIL은 NULL이 많이 보이는데 “원래 그런가 보다” 하고 넘어가거나, NULL을 0/false/빈 문자열로 일괄 치환해 보고서 숫자를 만든 뒤 의미를 검증하지 않는 상태입니다.

이 섹션의 핵심 함정은 `NOT IN`입니다. 오른쪽 목록이나 subquery에 NULL이 들어오면 결과가 통째로 사라질 수 있습니다. 안전한 습관은 다음 세 가지입니다. 첫째, 오른쪽 컬럼이 정말 NOT NULL인지 schema로 확인합니다. 둘째, NULL 가능성이 있으면 `WHERE user_id IS NOT NULL`로 subquery에서 제거하거나 `NOT EXISTS`로 바꿉니다. 셋째, 작은 fixture에서 NULL 한 줄을 일부러 넣고 결과를 확인합니다.

```sql
-- 의도 확인용 fixture
WITH users(user_id) AS (
    VALUES (1), (2), (3)
),
blocked(user_id) AS (
    VALUES (2), (NULL)
)
SELECT user_id
FROM users
WHERE user_id NOT IN (SELECT user_id FROM blocked);
```

PASS 기준은 이 쿼리가 빈 결과를 낼 수 있음을 예상하고 설명하는 것입니다. FAIL 기준은 `1`과 `3`이 나와야 한다고 생각하는 것입니다. 이 작은 실험 하나가 3값 논리의 많은 오류를 잡아 줍니다.

NULL은 불편하지만, 없애야 할 실수만은 아닙니다. 현실 세계에는 모르는 값, 아직 수집하지 않은 값, 적용 대상이 아닌 값이 존재합니다. 문제는 그 상태를 `0`, `''`, `'UNKNOWN'`, `false` 같은 임의 값으로 숨길 때 생깁니다. SQL의 NULL은 그 불확실성을 드러내고, 3값 논리는 그 불확실성이 비교와 필터에 어떻게 전파되는지 보여 줍니다. 좋은 모델링은 NULL을 무조건 피하는 것이 아니라, 반드시 필요한 곳에만 허용하고, 허용한 곳에서는 `IS NULL`, `COALESCE`, NULL-safe 비교, NOT NULL 제약, 작은 truth table 실험으로 의미를 검증하는 것입니다.

NULL을 더 현실적으로 보려면 “모름”, “없음”, “해당 없음”을 구분해야 합니다. 같은 NULL이라도 업무 의미는 하나가 아닐 수 있습니다. 예를 들어 배송 완료 시각 `delivered_at`이 NULL이면 아직 배송되지 않았다는 뜻일 수 있습니다. 반면 사용자 생년월일 `birth_date`가 NULL이면 사용자가 입력하지 않았거나 수집하지 않는 정책일 수 있습니다. 탈퇴 시각 `withdrawn_at`이 NULL이면 아직 탈퇴하지 않았다는 현재 상태일 수 있습니다. DB 값은 모두 NULL이지만, 쿼리와 화면과 통계에서 해석은 다릅니다.

```text
컬럼별 NULL 의미 예

orders.delivered_at = NULL
  -> 아직 배송 완료 사건이 없음

users.birth_date = NULL
  -> 모름/미수집, 생일이 없다는 뜻은 아님

users.withdrawn_at = NULL
  -> 아직 탈퇴하지 않은 현재 상태

payments.pg_approved_at = NULL
  -> PG 승인 성공 시각이 아직 없음, 실패/대기/미요청 상태와 구분 필요
```

이 구분이 중요한 이유는 NULL을 제거하려고 임의 기본값을 넣을 때 드러납니다. `delivered_at`에 `1970-01-01`을 넣으면 정렬과 기간 통계가 왜곡됩니다. `birth_date`에 `1900-01-01`을 넣으면 나이 계산이 틀립니다. `withdrawn_at`에 빈 문자열을 넣으면 날짜 타입을 잃습니다. `pg_approved_at`에 현재 시각을 넣으면 승인되지 않은 결제를 승인된 것처럼 보이게 할 수 있습니다. NULL을 불편하다고 가짜 값으로 숨기면 3값 논리 대신 더 위험한 거짓 사실을 저장하게 됩니다.

```text
가짜 기본값의 위험 trace

실제 상태:
  user.birth_date = 모름

가짜 저장:
  birth_date = DATE '1900-01-01'

나이 계산:
  126세로 계산됨

마케팅 조건:
  age >= 65 조건에 포함됨

결과:
  모르는 값을 고령 사용자라는 거짓 사실로 바꿔 버림
```

반대로 NULL을 허용하지 않아야 하는 곳도 많습니다. 주문 금액, 주문 라인 수량, ledger entry의 debit/credit 금액, 상태 코드, 생성 시각처럼 핵심 불변식을 구성하는 값은 NULL이 되면 대부분의 쿼리와 검증이 흔들립니다. 이 경우 `NOT NULL`은 귀찮은 제약이 아니라 데이터 의미의 일부입니다.

```sql
CREATE TABLE ledger_entries (
    entry_id bigint PRIMARY KEY,
    account_id bigint NOT NULL,
    amount numeric(18, 2) NOT NULL,
    direction text NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    occurred_at timestamp NOT NULL
);
```

이 테이블에서 `amount`가 NULL이면 합계가 조용히 누락될 수 있습니다. `direction`이 NULL이면 차변인지 대변인지 모릅니다. `occurred_at`이 NULL이면 시간 순서 재구성이 어렵습니다. 금융·정산 영역에서 이런 NULL은 단순 미입력이 아니라 감사 불가능성을 만듭니다.

3값 논리는 CHECK 제약에서도 주의해야 합니다. SQL 표준과 DB 구현에서는 CHECK 조건이 FALSE일 때 위반으로 보고, TRUE 또는 UNKNOWN은 통과하는 식의 동작을 하는 경우가 많습니다. 즉 nullable 컬럼에 `CHECK (amount > 0)`만 두면 `amount=NULL`이 통과할 수 있습니다. 양수이면서 반드시 있어야 한다면 `NOT NULL`과 CHECK를 함께 둬야 합니다.

```sql
-- amount NULL을 허용할 수 있음
CREATE TABLE weak_payments (
    amount numeric CHECK (amount > 0)
);

-- amount가 반드시 있고 양수여야 함
CREATE TABLE strong_payments (
    amount numeric NOT NULL CHECK (amount > 0)
);
```

이 예는 NULL이 제약 조건에서도 UNKNOWN을 만든다는 점을 보여 줍니다. “양수여야 한다”와 “값이 있어야 한다”는 서로 다른 조건입니다. 둘 다 필요하면 둘 다 써야 합니다. 운영에서 `amount`가 NULL인 row가 발견되면 “CHECK가 있는데 왜 들어왔지?”라고 놀라기 전에 CHECK가 UNKNOWN을 어떻게 다루는지 확인해야 합니다.

NULL은 `IN`에서도 조심해야 합니다.

```sql
SELECT user_id
FROM users
WHERE email IN ('a@example.com', NULL);
```

`email='a@example.com'`인 row는 TRUE입니다. 하지만 `email`이 NULL인 row가 이 조건으로 잡히는 것은 아닙니다. `NULL IN (...)` 역시 UNKNOWN이 됩니다. NULL을 포함하고 싶으면 `email IS NULL`을 별도로 써야 합니다.

```sql
SELECT user_id
FROM users
WHERE email IN ('a@example.com')
   OR email IS NULL;
```

`NOT IN`은 더 위험하지만, `IN`도 NULL을 “목록에 NULL이 있으니 NULL 값도 포함”으로 해석하지 않는다는 점을 기억해야 합니다. SQL에서 NULL 판단은 일반 equality가 아니라 `IS NULL` 계열로 해야 합니다.

NULL과 `OR` 조건은 화면 필터에서도 문제를 만듭니다. 예를 들어 “마케팅 동의하지 않은 사용자”를 찾는 화면이 있다고 합시다.

```sql
SELECT user_id
FROM users
WHERE marketing_agreed = false;
```

이 쿼리는 명시적으로 거부한 사용자만 보여 줍니다. 아직 선택하지 않은 사용자를 포함하려면 다음처럼 써야 합니다.

```sql
SELECT user_id
FROM users
WHERE marketing_agreed = false
   OR marketing_agreed IS NULL;
```

두 결과는 정책적으로 다릅니다. “거부한 사용자”와 “동의하지 않은 사용자”가 같은 말인지 먼저 정해야 합니다. 법적 동의가 필요한 영역에서는 NULL을 false처럼 취급할 수 있습니다. 하지만 데이터 분석에서는 “거부”와 “미응답”을 분리해야 할 수 있습니다. SQL 조건은 정책 문장을 그대로 반영해야 합니다.

```text
정책 문장 -> SQL 조건

"명시적으로 거부한 사용자"
  marketing_agreed = false

"동의 상태가 아니어서 발송하면 안 되는 사용자"
  marketing_agreed = false OR marketing_agreed IS NULL

"동의 여부를 아직 받지 않은 사용자"
  marketing_agreed IS NULL
```

이 매핑을 문서나 코드 주석에 남기면 나중에 조건을 바꾸는 사람이 의미를 잃지 않습니다. 특히 Boolean nullable 컬럼은 이름만으로 의미가 잘 드러나지 않습니다. `is_active`, `enabled`, `verified`, `agreed` 같은 컬럼이 NULL을 가질 수 있다면 NULL의 의미를 반드시 고정해야 합니다.

NULL은 `JOIN` 조건에서도 equality를 흐립니다. 두 테이블에서 nullable 컬럼을 기준으로 join한다고 가정합니다.

```sql
SELECT *
FROM a
JOIN b ON a.optional_code = b.optional_code;
```

`a.optional_code`와 `b.optional_code`가 둘 다 NULL인 row끼리는 매칭되지 않습니다. `NULL = NULL`이 TRUE가 아니기 때문입니다. 둘 다 NULL이면 같은 그룹으로 보고 싶다면 NULL-safe 비교를 쓰거나, NULL을 명시적으로 같은 bucket으로 치환하는 정책이 필요합니다.

```sql
-- PostgreSQL식 예
SELECT *
FROM a
JOIN b ON a.optional_code IS NOT DISTINCT FROM b.optional_code;

-- 정책적으로 sentinel을 쓰는 예, 단 sentinel 충돌 위험을 검토해야 함
SELECT *
FROM a
JOIN b ON coalesce(a.optional_code, '__NULL__') = coalesce(b.optional_code, '__NULL__');
```

두 번째 방식은 조심해야 합니다. 실제 값으로 `__NULL__`이 들어올 수 있으면 충돌합니다. 타입과 도메인에 맞는 안전한 sentinel이 없으면 더 위험합니다. DB가 제공하는 NULL-safe 비교가 있으면 그것을 우선 검토하는 편이 좋습니다.

통계와 옵티마이저도 NULL 분포를 봅니다. 컬럼의 NULL 비율이 높으면 조건 선택도 추정에 영향을 줍니다. `WHERE deleted_at IS NULL`은 많은 서비스에서 “삭제되지 않은 현재 row”를 찾는 매우 흔한 조건입니다. 만약 대부분 row가 삭제되지 않았다면 이 조건은 별로 선택적이지 않습니다. 반대로 archive 테이블에서 NULL이 드물면 선택적일 수 있습니다. 인덱스를 만들 때도 NULL 조건을 포함한 partial index나 composite index를 고민할 수 있습니다.

```sql
-- PostgreSQL partial index 예
CREATE INDEX orders_active_created_at_idx
ON orders(created_at)
WHERE deleted_at IS NULL;
```

이 인덱스는 삭제되지 않은 row만 대상으로 날짜 탐색을 돕습니다. 모든 DB가 같은 방식의 partial index를 지원하는 것은 아니므로 DB별 공식 문서를 확인해야 합니다. 핵심은 NULL이 단순 값 부재가 아니라 데이터 분포와 실행 계획에도 영향을 준다는 점입니다.

NULL 관련 장애의 재현 실험은 일부러 NULL 한 줄을 넣어야 합니다. 테스트 데이터가 모두 NOT NULL이면 3값 논리 버그는 드러나지 않습니다.

```sql
WITH t(value) AS (
    VALUES (1), (2), (NULL)
)
SELECT
    value,
    value = 1 AS eq_one,
    NOT (value = 1) AS not_eq_one,
    value IS NULL AS is_null
FROM t;
```

이 결과를 눈으로 보면 NULL row의 `eq_one`과 `not_eq_one`이 모두 TRUE가 아니라는 사실을 확인할 수 있습니다. DB별 클라이언트가 UNKNOWN을 화면에 NULL처럼 표시할 수 있으므로, 표시 결과도 주의해서 봐야 합니다.

마지막으로 NULL을 다루는 실무 기준을 정리합니다.

| 질문 | 좋은 판단 |
|---|---|
| 이 컬럼의 NULL은 무엇을 뜻하는가 | 모름, 미수집, 해당 없음, 아직 사건 없음 중 하나를 문서화 |
| 반드시 있어야 하는가 | `NOT NULL`과 적절한 default/check를 둠 |
| 집계에서 제외해야 하는가 | `COUNT(*)`, `COUNT(col)`, `COALESCE` 정책을 구분 |
| 비교에서 NULL끼리 같게 볼 것인가 | DB별 NULL-safe 비교 또는 명시 정책 사용 |
| `NOT IN`을 쓰는가 | 오른쪽 값이 NOT NULL인지 확인하거나 `NOT EXISTS` 검토 |
| 정렬에서 NULL 위치가 중요한가 | `NULLS FIRST/LAST` 또는 DB별 표현 명시 |

NULL과 3값 논리가 SQL에서 중요해진 배경은 데이터베이스가 현실의 불완전한 정보를 다루기 때문입니다. 업무 데이터에는 아직 모르는 값, 나중에 채워질 값, 사건이 발생하지 않아 비어 있는 값이 계속 등장합니다. SQL은 이 상태를 억지로 TRUE/FALSE 중 하나로 밀어 넣지 않고 UNKNOWN을 남깁니다. 그 결과 쿼리는 불편해지지만, 모르는 사실을 거짓 사실로 바꾸는 위험은 줄어듭니다.

이 기준은 NULL을 무서워하라는 뜻이 아닙니다. NULL을 의미 있는 상태로 받아들이되, 그 상태가 비교와 필터와 집계와 정렬을 지날 때 어떻게 전파되는지 추적하라는 뜻입니다. SQL의 3값 논리는 처음에는 불편하지만, 현실의 “아직 모름”을 거짓 값으로 바꾸지 않기 위한 장치입니다.

NULL을 애플리케이션 언어의 `null`과 그대로 같다고 보는 것도 조심해야 합니다. Java나 Kotlin에서 `null`은 참조가 없다는 런타임 값입니다. SQL의 NULL은 relation 안의 attribute 값이 unknown/not applicable 상태라는 표시이고, 비교 결과를 UNKNOWN으로 전파합니다. ORM은 이 둘을 매핑하지만 의미가 완전히 같아지는 것은 아닙니다. Kotlin의 nullable type이 `String?`라고 해서 DB 컬럼이 nullable이어야 하는 것은 아니고, DB 컬럼이 nullable이라고 해서 애플리케이션에서 아무 곳에서나 null을 허용해야 하는 것도 아닙니다. 경계에서 의미를 다시 정해야 합니다.

```text
DB -> 애플리케이션 경계

DB column:
  users.withdrawn_at timestamp NULL
  의미: 탈퇴하지 않았으면 NULL

Kotlin type:
  withdrawnAt: Instant?
  의미: null이면 active user로 볼 수 있음

위험:
  다른 문맥에서 null을 "정보 로딩 안 됨"으로도 쓰면 의미 충돌
```

이런 충돌을 막으려면 nullable 값을 도메인 상태로 승격하는 것이 좋습니다. 예를 들어 `withdrawn_at IS NULL`을 매번 직접 해석하기보다, 조회 결과를 `ACTIVE`/`WITHDRAWN` 상태로 변환해 서비스 계층에서 쓰는 방식입니다. DB는 원천 사실을 저장하고, 애플리케이션은 그 사실을 도메인 판단으로 바꿉니다. 단, 이 변환이 DB 쿼리와 코드 여러 곳에 흩어지면 다시 drift가 생기므로 view, query helper, domain mapper 같은 경계를 정해야 합니다.

NULL은 JSON 컬럼에서도 혼동을 만듭니다. JSON 안의 `null`, JSON key 부재, SQL NULL은 서로 다를 수 있습니다.

```text
SQL row examples

payload = SQL NULL
  -> payload 자체가 없음

payload = '{}'
  -> JSON object는 있지만 key가 없음

payload = '{"nickname": null}'
  -> nickname key는 있고 JSON value가 null
```

JSON을 다루는 DB 함수들은 이 차이를 DB별로 다르게 표현할 수 있습니다. 따라서 JSON 컬럼에서 “값이 없다”를 찾을 때는 SQL NULL인지, key 부재인지, JSON null인지 구분해야 합니다. 이 차이를 무시하면 API payload migration이나 검색 조건에서 버그가 생깁니다.

NULL 처리의 마지막 practical trap은 default와 backfill입니다. 새 컬럼을 추가할 때 nullable로 열어 두고 나중에 채운 뒤 NOT NULL로 바꾸는 migration은 흔합니다. 이때 중간 기간의 애플리케이션은 NULL을 볼 수 있습니다. 코드가 새 컬럼을 항상 non-null로 가정하면 배포 순서 사이에서 장애가 납니다.

```text
안전한 NOT NULL migration 흐름

1. nullable 컬럼 추가
2. 애플리케이션이 NULL을 처리할 수 있게 배포
3. 기존 row backfill
4. 새 write path가 값을 항상 채우는지 확인
5. NULL 잔여 row count = 0 검증
6. NOT NULL constraint 추가
```

이 흐름에서 검증 쿼리는 단순하지만 중요합니다.

```sql
SELECT count(*)
FROM users
WHERE new_required_column IS NULL;
```

PASS는 constraint를 추가하기 전에 NULL 잔여 row가 0임을 확인하고, 새 write path가 값을 채우는지 테스트한 상태입니다. FAIL은 로컬 fixture에는 값이 있으니 운영에서도 괜찮다고 생각하고 바로 NOT NULL을 추가하는 것입니다. 큰 테이블에서는 constraint 추가 자체가 lock이나 table scan을 유발할 수 있으므로 DB별 online DDL 지원도 확인해야 합니다.

NULL은 단일 문법 항목이 아니라 데이터 생명주기 전체에 걸친 의미입니다. 입력 시점에는 값이 없을 수 있고, 저장 시점에는 nullable 제약이 작동하고, 조회 시점에는 3값 논리가 전파되고, 집계 시점에는 제외될 수 있고, migration 시점에는 backfill 대상이 됩니다. 이 흐름을 한 번에 보아야 “NULL 때문에 결과가 이상하다”는 막연한 말 대신 어느 단계에서 의미가 흐려졌는지 찾을 수 있습니다.

NULL을 다루는 문서나 코드에서는 “없음”이라는 말을 너무 쉽게 쓰지 않는 편이 좋습니다. 없음은 세 가지 이상의 의미를 뭉갭니다. 아직 수집하지 않음, 실제로 존재하지 않음, 정책상 표시하지 않음, 권한이 없어 숨김, 외부 시스템이 응답하지 않음이 모두 “없음”으로 보일 수 있습니다. DB에 NULL 하나로 저장하더라도 애플리케이션이 사용자에게 보여 줄 메시지는 달라야 합니다.

```text
같은 NULL, 다른 사용자 메시지

birth_date NULL:
  "생년월일을 입력하지 않았습니다"

delivered_at NULL:
  "아직 배송 완료 전입니다"

pg_approved_at NULL + payment_status='FAILED':
  "결제 승인이 실패했습니다"

pg_approved_at NULL + payment_status='PENDING':
  "결제 승인 결과를 기다리는 중입니다"
```

이 예에서 `pg_approved_at`만 보면 두 상태를 구분할 수 없습니다. 그래서 nullable timestamp와 상태 enum이 함께 필요할 수 있습니다. NULL 하나로 모든 상태를 표현하면 쿼리는 짧아지지만 의미가 빈약해집니다. 반대로 상태 enum만 있고 timestamp가 없으면 언제 그 상태가 되었는지 알 수 없습니다. 좋은 모델은 NULL을 의미 있는 다른 컬럼들과 함께 배치합니다.

```sql
CREATE TABLE payments (
    payment_id bigint PRIMARY KEY,
    payment_status text NOT NULL CHECK (payment_status IN ('PENDING', 'APPROVED', 'FAILED')),
    pg_approved_at timestamp,
    pg_failed_at timestamp,
    CHECK (
        (payment_status = 'APPROVED' AND pg_approved_at IS NOT NULL)
        OR
        (payment_status <> 'APPROVED')
    )
);
```

위 CHECK는 단순 예시이며 모든 상태 조합을 완벽히 닫지는 않습니다. 하지만 중요한 방향을 보여 줍니다. NULL을 허용하는 컬럼이 다른 상태 컬럼과 어떤 관계를 가져야 하는지 제약으로 표현할 수 있습니다. 복잡한 상태 머신은 애플리케이션과 DB 제약을 함께 써야 하지만, “승인 상태인데 승인 시각이 없다” 같은 기본 모순은 DB에서 막는 편이 안전합니다.

NULL 관련 쿼리 리뷰에서는 다음 질문을 쓰면 좋습니다.

```text
1. 이 컬럼이 nullable인 이유가 문서화되어 있는가?
2. WHERE 조건에서 UNKNOWN이 생기면 그 row를 버리는 것이 맞는가?
3. NOT 조건이 NULL row를 포함한다고 착각하지 않았는가?
4. aggregate가 NULL을 제외하는 것이 지표 의도와 맞는가?
5. outer join으로 생긴 NULL과 원래 데이터의 NULL을 구분해야 하는가?
6. migration 중 일시적 NULL을 애플리케이션이 견딜 수 있는가?
```

outer join에서 생긴 NULL과 원래 데이터의 NULL도 다릅니다. 오른쪽 테이블에 매칭 row가 없어 생긴 NULL은 “join 결과에서 채워진 빈 자리”입니다. 오른쪽 row는 있었지만 그 컬럼 값이 NULL인 경우와 구분해야 할 수 있습니다.

```text
LEFT JOIN 결과

case A: 주문 row 없음
  o.order_id = NULL
  o.memo = NULL

case B: 주문 row 있음, memo만 NULL
  o.order_id = 1001
  o.memo = NULL

구분 기준:
  o.order_id IS NULL이면 row 없음
  o.order_id IS NOT NULL AND o.memo IS NULL이면 row는 있으나 memo 없음
```

그래서 “오른쪽 값이 없는 사용자”를 찾을 때는 nullable할 수 있는 일반 컬럼보다 오른쪽 테이블의 NOT NULL key 컬럼을 기준으로 보는 편이 안전합니다. 이 작은 습관이 outer join NULL 해석 오류를 줄입니다.

### 마지막 실무 연결: NULL은 값이 아니라 판단 보류를 만든다

NULL을 빈 문자열이나 0처럼 다루면 조건식이 조용히 무너진다. 특히 `NOT IN`과 NULL 조합은 실무 장애로 자주 이어진다. 제외 목록에 NULL이 섞이는 순간 비교 결과가 UNKNOWN으로 번지고, 개발자는 분명히 제외 조건을 썼는데 row가 사라지는 상황을 만난다. 이때 문제는 DB가 이상한 것이 아니라, 모름을 값처럼 취급한 설계에 있다.

검증 습관은 간단하다. nullable column이 조건식에 들어가면 `IS NULL`, `IS NOT NULL`, `IS DISTINCT FROM`, `NOT EXISTS` 같은 명시적 표현을 먼저 검토한다. 그리고 테스트에는 항상 NULL row를 넣는다. NULL row가 없는 테스트는 SQL의 절반만 검증한 것이다.


## type, charset, collation, comparison

문자열 비교를 단순히 “두 글자가 같은가”로 보면 DB에서 자주 틀립니다. DB가 문자열을 저장하고 비교할 때는 적어도 세 층이 있습니다. 첫째, 타입은 컬럼이 어떤 종류의 값을 담는지 정합니다. `varchar(50)`, `text`, `char(10)` 같은 선언이 여기에 해당합니다. 둘째, character set, 한국어로 문자 집합 또는 문자 인코딩 계층은 문자를 어떤 코드와 바이트 표현으로 저장할지 정합니다. MySQL 공식 문서는 character set을 symbol과 encoding의 집합으로 설명합니다. 셋째, collation, 한국어로는 보통 콜레이션 또는 정렬/비교 규칙이라고 부르는 것은 같은 문자 집합 안에서 문자열을 어떻게 비교하고 정렬할지 정합니다. MySQL 공식 문서는 character set마다 하나 이상의 collation이 있고, collation은 문자 비교 규칙이라고 설명합니다.

작은 예부터 봅니다.

```sql
CREATE TABLE people (
    name varchar(50)
);

INSERT INTO people VALUES
  ('Jose'),
  ('José'),
  ('jose'),
  ('김철수'),
  ('김영희');
```

사람은 `Jose`, `José`, `jose`를 상황에 따라 같게 보거나 다르게 볼 수 있습니다. 로그인 ID라면 대소문자를 구분할 수도 있고, 검색 화면이라면 대소문자를 무시할 수도 있습니다. 스페인어 이름 검색에서는 악센트를 무시하고 찾고 싶을 수도 있고, 언어학적 정밀 비교에서는 구분해야 할 수도 있습니다. DB는 이 비교 정책을 collation으로 표현합니다.

```text
비교 정책 예

case-sensitive:
  'Jose' != 'jose'

case-insensitive:
  'Jose' = 'jose'

accent-sensitive:
  'Jose' != 'José'

accent-insensitive:
  'Jose' = 'José'

binary-like:
  코드 포인트 또는 바이트 표현에 가깝게 비교
```

MySQL collation 이름에는 이런 정책이 드러나는 경우가 많습니다. 예를 들어 `utf8mb4_0900_ai_ci`는 `utf8mb4` 문자 집합, Unicode 9.0 기반 계열, accent-insensitive, case-insensitive 비교 규칙이라는 의미로 읽을 수 있습니다. `_as_cs`는 accent-sensitive, case-sensitive를 뜻할 수 있고, `_bin` 계열은 binary 비교에 가깝습니다. 이름 규칙은 DB 버전과 collation 종류에 따라 세부가 있으므로 공식 문서를 확인해야 하지만, 이름을 보면 비교 민감도를 어느 정도 추적할 수 있습니다.

기존 seed `database/collation.md`가 잘 짚은 것처럼 문자셋과 collation은 다릅니다. 문자셋은 문자를 어떤 바이트로 표현할지에 가깝고, collation은 그 문자를 어떻게 비교할지에 가깝습니다.

```text
문자열 'A'

character set / encoding:
  'A'를 어떤 코드/바이트로 저장하는가

collation:
  'A'와 'a'를 같은 것으로 볼 것인가
  'é'와 'e'를 같은 것으로 볼 것인가
  정렬할 때 'Z'와 'a'의 순서를 어떻게 둘 것인가
```

`VARCHAR(1)`이 1 byte와 같은지 묻는 기존 seed도 같은 층위 문제입니다. `VARCHAR(1)`의 1은 보통 “문자 수”에 가까운 의미이지, 항상 1 byte라는 뜻이 아닙니다. UTF-8 계열에서는 영문자 `A`는 1 byte일 수 있지만, 한글 `가`는 보통 3 byte이고, 이모지는 4 byte일 수 있습니다. MySQL에서는 row format과 길이에 따라 길이 저장 overhead도 있습니다. 따라서 타입 길이, 문자 수, 바이트 수를 섞으면 저장 크기와 index prefix 제한을 잘못 계산합니다.

```text
대표 문자열의 UTF-8 바이트 감각

'A'  -> 1 byte
'가' -> 3 bytes
'😄' -> 4 bytes

VARCHAR(1):
  문자 하나를 담을 수 있다는 뜻에 가깝다.
  실제 저장 바이트는 문자와 인코딩에 따라 달라진다.
```

비교 trace를 보겠습니다. 같은 데이터를 두 collation으로 비교한다고 가정합니다.

```text
rows:
  1: 'Jose'
  2: 'José'
  3: 'jose'

case/accent insensitive collation:
  WHERE name = 'jose'
  -> 1, 2, 3 모두 매칭될 수 있음

case/accent sensitive 또는 binary-like collation:
  WHERE name = 'jose'
  -> 3만 매칭될 수 있음
```

이 결과는 단순 검색 UI 문제로 끝나지 않습니다. unique constraint에도 영향을 줍니다. case-insensitive collation에서 `email`에 unique index가 있으면 `User@example.com`과 `user@example.com`을 같은 값으로 보고 중복을 막을 수 있습니다. binary-like collation에서는 서로 다른 값으로 들어갈 수 있습니다. 어느 쪽이 맞는지는 업무 정책에 달려 있습니다. 이메일 local-part는 표준적으로는 수신자 사이트 규칙에 의존하는 복잡한 영역이지만, 많은 서비스는 실무적으로 대소문자를 무시합니다. 따라서 서비스 정책을 정하고, DB collation이나 normalized column으로 그 정책을 구현해야 합니다.

```sql
CREATE TABLE accounts (
    account_id bigint PRIMARY KEY,
    email varchar(255) NOT NULL
);

-- 정책 예:
-- 1. 저장은 사용자가 입력한 원문 보존
-- 2. 비교용 normalized_email은 lower-case로 관리
-- 3. normalized_email에 unique constraint
```

이런 설계는 DB collation만 믿는 방식과 애플리케이션 normalization을 함께 쓰는 방식 사이의 선택입니다. DB collation으로 대소문자 무시 unique를 구현하면 간단할 수 있습니다. 하지만 서비스가 여러 DB를 쓰거나, 일부 비교는 case-sensitive여야 하거나, Unicode normalization까지 통제해야 하면 별도 normalized column이 더 명확할 수 있습니다. 중요한 것은 비교 정책을 암묵적으로 두지 않는 것입니다.

collation은 ORDER BY에도 영향을 줍니다.

```text
값:
  'Zebra'
  'apple'
  'Álvaro'
  'alvaro'

binary-like 정렬:
  코드 값 또는 바이트 표현에 가까운 순서가 나올 수 있음

언어/대소문자 무시 정렬:
  사람이 기대하는 알파벳 순서에 더 가까울 수 있음
```

정렬 결과는 pagination과도 연결됩니다. `ORDER BY name LIMIT 20`이 collation에 따라 다른 20명을 보여 줄 수 있습니다. replica나 다른 환경의 collation 설정이 다르면 같은 쿼리의 순서가 달라질 수 있습니다. migration 중 character set이나 collation을 바꾸는 작업이 위험한 이유입니다. 단순히 문자가 깨지지 않게 하는 문제가 아니라, equality, unique, group, distinct, order, index 탐색 의미가 달라질 수 있습니다.

MySQL에서는 서버, 데이터베이스, 테이블, 컬럼, expression 수준에서 character set과 collation이 정해질 수 있습니다. 기존 seed의 `SHOW` 명령은 이런 관측에 유용합니다.

```sql
-- database 기본 문자셋/콜레이션 확인
SELECT
    SCHEMA_NAME AS database_name,
    DEFAULT_CHARACTER_SET_NAME AS character_set,
    DEFAULT_COLLATION_NAME AS collation
FROM information_schema.SCHEMATA;

-- 현재 세션 변수 확인
SHOW VARIABLES
WHERE VARIABLE_NAME LIKE '%coll%'
   OR VARIABLE_NAME LIKE '%char%';

-- 컬럼별 collation 확인
SHOW FULL COLUMNS FROM people;
```

PASS는 비교 결과가 이상할 때 실제 컬럼 collation과 connection/session character set을 확인하는 것입니다. FAIL은 “로컬에서는 됐는데 운영에서는 왜 다르지?”라고 말하면서 환경별 collation 차이를 보지 않는 것입니다. 특히 MySQL에서는 connection character set도 중요합니다. 클라이언트가 보낸 바이트를 서버가 어떤 문자셋으로 해석하는지에 따라 저장 전부터 값이 깨질 수 있습니다.

타입 변환도 비교 의미를 바꿉니다.

```sql
SELECT '10' = 10;
```

DBMS마다 문자열과 숫자의 암묵 변환 규칙이 다릅니다. MySQL은 비교 문맥에서 타입 변환을 수행하는 경우가 많고, PostgreSQL은 더 엄격한 타입 해석을 요구하는 경우가 많습니다. 세부는 공식 문서로 확인해야 합니다. 실무 원칙은 단순합니다. 비교하려는 값은 같은 의미의 타입으로 맞추고, 암묵 변환에 기대지 않습니다. 숫자 컬럼을 문자열 파라미터와 비교하거나, 날짜 컬럼을 문자열 함수로 감싸 비교하면 인덱스 사용과 결과 의미가 모두 흔들릴 수 있습니다.

```sql
-- 약한 형태
WHERE to_char(created_at, 'YYYY-MM-DD') = '2026-05-18'

-- 더 강한 형태
WHERE created_at >= DATE '2026-05-18'
  AND created_at <  DATE '2026-05-19'
```

두 쿼리는 비슷한 결과를 줄 수 있지만 물리 실행과 타입 의미가 다릅니다. 첫 번째는 컬럼 값을 문자열로 변환해 비교합니다. 두 번째는 date/timestamp 범위로 비교합니다. 인덱스와 통계가 날짜 타입에 맞게 쓰일 가능성이 커집니다. 논리와 물리 실행 모델이 다시 만나는 지점입니다.

collation 차이는 index에도 영향을 줍니다. B-tree index는 정렬된 key를 기준으로 탐색합니다. 문자열 key의 정렬과 동등성은 collation이 정합니다. 따라서 collation이 대소문자를 무시하면 index도 그 비교 규칙에 따라 key를 정렬하고 찾습니다. 같은 문자열 컬럼이라도 collation이 다르면 index 의미가 달라집니다. 어떤 DB에서는 쿼리에서 컬럼과 다른 collation을 강제로 적용하면 기존 index를 못 쓰거나 다른 계획을 선택할 수 있습니다. 세부는 DB별로 다르므로 `EXPLAIN`으로 확인해야 합니다.

```sql
EXPLAIN
SELECT *
FROM people
WHERE name = 'jose';

EXPLAIN
SELECT *
FROM people
WHERE name COLLATE utf8mb4_0900_as_cs = 'jose';
```

위 두 쿼리는 비교 정책이 다를 수 있고, index 사용 여부도 달라질 수 있습니다. PASS는 결과 행과 실행 계획을 모두 확인하는 것입니다. FAIL은 collation을 쿼리에 임시로 붙여 결과만 맞추고, index와 unique 정책이 어떻게 달라지는지 보지 않는 것입니다.

문자열 LIKE 검색도 collation 영향을 받습니다.

```sql
SELECT *
FROM people
WHERE name LIKE 'jo%';
```

case-insensitive collation에서는 `Jose`, `jose`가 모두 매칭될 수 있습니다. accent-insensitive라면 `José`도 매칭될 수 있습니다. binary-like collation에서는 다를 수 있습니다. 검색 정책이 “사용자 친화적 이름 검색”인지 “정확한 코드 prefix 검색”인지에 따라 collation 선택이 달라집니다. 코드, 토큰, 해시, 외부 식별자 같은 값은 사람이 읽는 문자열처럼 보이더라도 보통 binary에 가까운 정확 비교가 필요합니다.

```text
컬럼 의미별 비교 정책 예

사람 이름:
  언어/대소문자/악센트 정책을 고려한 검색 필요

이메일:
  서비스 정책에 따른 normalization과 unique 필요

외부 거래 id:
  정확 비교 필요, 대소문자 무시하면 위험할 수 있음

해시/token:
  binary-like 정확 비교 필요
```

collation migration은 특히 위험합니다. 이미 unique index가 있는 컬럼의 collation을 case-sensitive에서 case-insensitive로 바꾸면 이전에는 서로 달랐던 값들이 새 규칙에서는 같아질 수 있습니다.

```text
기존 collation:
  'User@example.com' != 'user@example.com'
  두 row 모두 존재 가능

새 collation:
  'User@example.com' = 'user@example.com'
  unique constraint 충돌 가능
```

따라서 migration 전에는 충돌 후보를 미리 찾아야 합니다.

```sql
SELECT lower(email) AS normalized_email, count(*)
FROM accounts
GROUP BY lower(email)
HAVING count(*) > 1;
```

이 쿼리는 단순 예시입니다. 실제 normalization은 collation과 Unicode 규칙을 더 고려해야 할 수 있습니다. 하지만 핵심은 migration 전에 새 비교 정책에서 같은 것으로 접힐 값들을 찾는 것입니다. 검증 없이 collation을 바꾸면 배포 중 DDL 실패, unique 충돌, 조회 결과 변화가 생길 수 있습니다.

type, charset, collation을 구분하는 replay path는 다음과 같습니다.

```text
값: 'José'

1. type
   varchar(50)에 저장 가능한가?
   길이는 문자 기준인가 바이트 기준인가?

2. character set
   이 문자를 저장할 수 있는가?
   어떤 바이트 표현이 되는가?

3. collation
   'Jose'와 같게 비교되는가?
   'jose'와 같게 비교되는가?
   ORDER BY에서 어디에 놓이는가?

4. index/unique
   같은 비교 규칙으로 unique가 적용되는가?
   WHERE name='jose'가 index를 사용할 수 있는가?
```

이 네 질문을 통과하면 문자열 문제를 훨씬 안정적으로 다룰 수 있습니다. 문자열은 사람이 보기에는 글자지만, DB에는 타입 선언, 인코딩, 비교 규칙, 물리 index가 함께 붙은 값입니다.

타입을 더 넓게 보면 숫자와 날짜도 comparison의 일부입니다. `INTEGER`, `BIGINT`, `NUMERIC`, `DECIMAL`, `FLOAT`는 모두 숫자처럼 보이지만 저장 방식과 비교·계산 의미가 다릅니다. 금액에는 부동소수점 타입을 쓰지 않는 것이 일반적입니다. 이유는 단순히 “float가 부정확하다”가 아니라, 10진 소수 금액을 2진 부동소수점으로 표현하면 사람이 기대하는 0.1, 0.2 같은 값이 정확히 표현되지 않을 수 있고, 합계·반올림·비교에서 미세한 오차가 누적되기 때문입니다. 이 내용은 money 단원에서 더 깊게 다루지만, type comparison 단계에서도 “같은 숫자처럼 보이는 값이 같은 의미의 타입인가”를 물어야 합니다.

```sql
-- 금액 비교에서는 고정 소수/정밀 decimal 계열을 검토한다.
CREATE TABLE payments (
    payment_id bigint PRIMARY KEY,
    amount numeric(18, 2) NOT NULL
);
```

날짜와 시간도 문자열로 비교하면 위험합니다. ISO 형식의 문자열은 겉으로는 정렬이 맞아 보일 수 있지만, timezone, precision, locale, invalid value, index 사용 문제가 생깁니다. `created_at`이 timestamp라면 timestamp 값으로 비교해야 합니다.

```sql
-- 약한 형태
WHERE created_at_text >= '2026-05-19'

-- 더 강한 형태
WHERE created_at >= TIMESTAMP '2026-05-19 00:00:00'
  AND created_at <  TIMESTAMP '2026-05-20 00:00:00'
```

타입은 DB와 애플리케이션 사이에서도 맞아야 합니다. Java에서 `BigDecimal`로 다루는 금액을 DB에서는 `numeric`으로 저장하는지, 시간은 `Instant`, `LocalDateTime`, `OffsetDateTime` 중 무엇으로 표현하고 DB에는 어떤 timezone 정책으로 저장하는지 정해야 합니다. 타입 경계가 흐리면 prepared statement를 써도 잘못된 값이 들어갑니다. bind는 SQL injection을 막아 주지만, 잘못된 타입 정책을 자동으로 고쳐 주지는 않습니다.

character set은 문자 저장 가능성과 직접 연결됩니다. MySQL의 `utf8`이 과거에는 `utf8mb3` 별칭으로 3바이트 UTF-8만 지원하고, 전체 Unicode 범위를 담으려면 `utf8mb4`를 써야 한다는 점은 널리 알려진 함정입니다. 이모지나 일부 보조 평면 문자가 들어오는 서비스에서 `utf8mb3` 계열을 쓰면 저장 실패나 문자 손상이 생길 수 있습니다. 기존 seed도 `😄` 같은 예를 통해 이 차이를 짚었습니다.

```text
문자 저장 실패 trace

입력:
  "배송 좋아요 😄"

컬럼 character set:
  utf8mb3 계열

문제:
  😄는 U+1F604로 4바이트 UTF-8 표현이 필요
  저장 실패 또는 잘못된 문자 처리 가능

대응:
  utf8mb4 지원 여부와 connection/session charset 확인
```

저장 가능한 문자셋을 쓰더라도 connection 설정이 틀리면 문제가 생길 수 있습니다. 애플리케이션은 UTF-8 바이트를 보냈는데 DB 세션이 다른 문자셋으로 해석하면 저장 전부터 깨질 수 있습니다. 그래서 문자 깨짐을 볼 때는 컬럼 character set만 보지 말고 client, connection, database, table, column 수준을 함께 봅니다.

```sql
SHOW VARIABLES LIKE 'character_set_client';
SHOW VARIABLES LIKE 'character_set_connection';
SHOW VARIABLES LIKE 'character_set_results';
SHOW FULL COLUMNS FROM people;
```

collation은 equality뿐 아니라 grouping에도 영향을 줍니다.

```text
값:
  'Jose'
  'José'
  'jose'

accent/case insensitive grouping:
  GROUP BY name -> 하나의 그룹으로 접힐 수 있음

accent/case sensitive grouping:
  GROUP BY name -> 세 그룹으로 남을 수 있음
```

이 차이는 보고서 숫자를 바꿉니다. 고객 이름별 주문 수, 태그별 게시글 수, 코드별 집계에서 collation이 의도와 다르면 그룹 수가 달라집니다. `DISTINCT name`도 마찬가지입니다. 문자열 중복 제거는 바이트 중복 제거가 아니라 collation에 따른 동등성 판단일 수 있습니다.

```sql
SELECT name, count(*)
FROM people
GROUP BY name
ORDER BY name;
```

이 쿼리의 결과가 몇 그룹인지 보려면 `name` 컬럼의 collation을 알아야 합니다. 같은 SQL이라도 DB와 컬럼 설정이 다르면 그룹이 달라질 수 있습니다.

문자열 타입 중 `CHAR`와 `VARCHAR`의 차이도 비교에 영향을 줄 수 있습니다. 일부 DB에서는 `CHAR`가 고정 길이라 뒤쪽 공백을 채우고, 비교에서 trailing space를 특별하게 다룰 수 있습니다. MySQL도 문자열 타입과 collation에 따라 trailing spaces 비교 규칙이 세부적으로 다릅니다. 코드를 저장하는 컬럼에 `CHAR(10)`을 쓰고 실제 값은 `'ABC'`라면 저장과 비교에서 공백이 어떻게 처리되는지 확인해야 합니다.

```text
코드값 비교 함정

저장 의도:
  code = 'ABC'

CHAR(10) 저장/표시:
  'ABC       '처럼 공백 padding 가능

비교 정책:
  trailing spaces를 무시하는지, binary 비교인지 DB별 확인 필요
```

그래서 외부 전문 코드, token, hash, idempotency key 같은 값에는 타입과 collation을 더 엄격하게 봅니다. 사람이 읽는 자연어 이름과 기계가 비교하는 식별자는 같은 문자열 타입으로 보이지만 비교 정책이 달라야 할 수 있습니다.

```text
값의 성격별 권장 질문

사람이 읽는 이름:
  검색 편의를 위해 case/accent insensitive가 필요한가?

로그인 id/email:
  서비스 정책상 normalization이 필요한가?
  원문 보존과 비교용 값을 분리할 것인가?

외부 거래 id:
  대소문자 하나가 다른 id인가?
  binary-like collation이 필요한가?

JSON key/path:
  DB JSON 타입을 쓸 것인가, 문자열로 저장할 것인가?
  비교와 index 정책은 무엇인가?
```

collation은 cross-database migration에서도 문제를 만듭니다. MySQL에서 대소문자 무시 collation으로 운영하던 컬럼을 PostgreSQL로 옮길 때 같은 unique 정책이 자동으로 따라오지 않을 수 있습니다. PostgreSQL에서는 collation, `citext` extension, expression index on `lower(email)` 같은 대안을 검토할 수 있습니다. 반대로 PostgreSQL에서 case-sensitive unique였던 컬럼을 MySQL의 case-insensitive collation으로 옮기면 기존 데이터가 unique 충돌을 낼 수 있습니다. DB 제품을 바꾸는 작업은 SQL 문법 변환보다 비교 의미 보존이 더 어렵습니다.

```text
migration comparison audit

source:
  MySQL email varchar(255) COLLATE utf8mb4_0900_ai_ci UNIQUE

target:
  PostgreSQL text UNIQUE

질문:
  target UNIQUE가 대소문자/악센트 무시 정책을 보존하는가?
  보존하지 않는다면 lower(email) index나 citext 같은 대안이 필요한가?
  기존 데이터 중 target/source 정책 차이로 충돌하거나 분리되는 값이 있는가?
```

타입과 collation 문제는 테스트 데이터가 영어 소문자만 있으면 드러나지 않습니다. 테스트 fixture에 대소문자, 악센트, 한글, 이모지, trailing space, 숫자 문자열을 일부러 넣어야 합니다.

```sql
INSERT INTO people(name) VALUES
  ('Jose'),
  ('José'),
  ('jose'),
  ('김철수'),
  ('김 철수'),
  ('A'),
  ('a'),
  ('😄');

SELECT name FROM people ORDER BY name;
SELECT name, count(*) FROM people GROUP BY name ORDER BY name;
SELECT * FROM people WHERE name = 'jose';
```

PASS는 이 세 쿼리의 결과를 현재 collation 기준으로 설명할 수 있는 것입니다. FAIL은 결과가 예상과 달랐을 때 “DB가 이상하다”고 말하면서 collation을 확인하지 않는 것입니다.

observability 측면에서는 schema와 query plan을 같이 봐야 합니다.

```sql
SHOW CREATE TABLE people;

EXPLAIN
SELECT *
FROM people
WHERE name = 'jose';
```

`SHOW CREATE TABLE`은 컬럼의 type, character set, collation, index를 확인하게 해 줍니다. `EXPLAIN`은 그 비교가 어떤 접근 경로를 탔는지 보여 줍니다. 결과 의미와 성능을 동시에 닫으려면 둘 다 필요합니다.

문자셋과 collation이 별도 개념으로 등장한 배경은 문자가 단순한 ASCII 범위를 넘어 여러 언어와 지역 규칙을 담게 되었기 때문입니다. 바이트 표현만 정하면 저장은 가능하지만, 사람이 기대하는 같음과 정렬은 언어마다 다릅니다. 그래서 DB는 문자를 저장하는 규칙과 문자를 비교하는 규칙을 분리해 둡니다.

마지막으로 type/charset/collation 사고의 핵심 문장을 남깁니다. DB의 문자열은 그냥 글자가 아닙니다. 타입이 저장 가능한 형태와 길이를 정하고, 문자셋이 문자와 바이트의 대응을 정하고, collation이 같음과 순서를 정하고, index가 그 규칙을 물리 탐색 구조로 바꿉니다. 이 네 층을 분리해 읽으면 “왜 같은 문자열인데 unique가 걸리지?”, “왜 정렬이 이상하지?”, “왜 인덱스를 안 타지?”, “왜 이모지가 저장되지 않지?” 같은 질문을 하나씩 검증할 수 있습니다.

타입 변환은 join에서도 위험합니다. 한 테이블의 `user_id`는 integer이고 다른 테이블의 `user_id`는 varchar라고 가정합니다.

```sql
SELECT *
FROM users u
JOIN external_events e ON e.user_id = u.user_id;
```

DB가 암묵 변환을 허용하면 결과가 나올 수 있습니다. 하지만 어느 쪽을 어느 타입으로 변환하는지에 따라 index 사용이 달라지고, 변환 실패나 예기치 않은 매칭이 생길 수 있습니다. 문자열 `'0010'`과 숫자 `10`을 같은 사용자로 볼 것인지도 정책 문제입니다.

```text
값 비교

DB users.user_id:
  10

external_events.user_id:
  '0010'

질문:
  '0010'은 10과 같은 외부 사용자 id인가?
  아니면 앞자리 0이 의미 있는 별도 코드인가?
```

외부 시스템 id는 숫자처럼 보여도 문자열인 경우가 많습니다. 앞자리 0, 대소문자, prefix, check digit이 의미를 가질 수 있습니다. 이런 값을 DB에 integer로 저장하면 원래 의미를 잃습니다. 타입은 저장 효율만이 아니라 도메인 의미를 보존하는 선택입니다.

```sql
-- 외부 코드가 앞자리 0을 보존해야 한다면 문자열이 더 안전할 수 있다.
CREATE TABLE external_accounts (
    provider text NOT NULL,
    external_account_id varchar(64) NOT NULL,
    user_id bigint NOT NULL,
    PRIMARY KEY (provider, external_account_id)
);
```

숫자 문자열을 숫자로 바꿔 저장하면 정렬도 바뀝니다.

```text
문자열 정렬:
  '1', '10', '2'

숫자 정렬:
  1, 2, 10
```

어느 쪽이 맞는지는 값의 의미에 달려 있습니다. 상품 옵션 코드나 우편번호는 숫자처럼 보일 수 있지만 산술 대상이 아닙니다. 금액과 수량은 산술 대상입니다. 타입 선택은 “어떤 연산이 의미 있는가”를 기준으로 해야 합니다. 더하기/빼기가 의미 없고 앞자리 0이 의미 있으면 숫자 타입이 아닐 가능성이 큽니다.

문자열 길이도 byte와 character와 grapheme cluster가 다를 수 있습니다. 사용자는 이모지 하나를 한 글자로 보지만, Unicode 관점에서는 여러 code point가 결합된 grapheme일 수 있습니다. DB의 `varchar(10)`이 사용자가 보는 글자 10개를 항상 의미하지 않을 수 있습니다. MySQL의 길이 제한, PostgreSQL의 character varying 의미, 애플리케이션 UI의 글자 수 제한이 서로 다를 수 있습니다. 사용자 입력 길이 제한을 엄격히 맞춰야 하면 DB와 애플리케이션 양쪽에서 같은 기준을 쓰는지 확인해야 합니다.

```text
길이 기준

byte length:
  저장 공간과 index prefix 제한에 중요

character/code point length:
  DB 문자열 길이 함수가 보는 기준일 수 있음

grapheme cluster:
  사용자가 화면에서 한 글자로 느끼는 단위
```

이 차이는 닉네임, 메시지, SMS 길이, 결제 영수증 출력에서 실제 문제가 됩니다. DB에 들어간다고 해서 UI 제한이 맞는 것은 아니고, UI에서 10자로 보인다고 DB index prefix가 충분한 것도 아닙니다.

collation coercion도 알아야 합니다. SQL expression에서 서로 다른 collation을 가진 두 문자열을 비교하거나 concat하면 DB가 어떤 collation을 적용할지 결정해야 합니다. MySQL에는 collation coercibility 규칙이 있고, 충돌하면 “illegal mix of collations” 같은 오류가 날 수 있습니다. 운영에서 갑자기 문자열 비교 오류가 나면 테이블 정의뿐 아니라 literal, connection collation, expression collation을 함께 봐야 합니다.

```sql
-- 쿼리 수준에서 collation을 명시하는 예
SELECT *
FROM people
WHERE name COLLATE utf8mb4_0900_as_cs = 'Jose';
```

이 명시는 문제를 임시로 해결할 수 있지만, 전체 schema 정책과 다르면 쿼리마다 다른 비교 의미가 생깁니다. 한 화면에서는 `Jose`와 `José`가 같고, 다른 화면에서는 다르면 사용자도 운영자도 혼란스럽습니다. query-level collation은 진단이나 특수 요구에 쓰되, 일반 정책은 컬럼과 index 설계에 반영하는 편이 좋습니다.

unique와 collation의 관계는 장애 비용이 큽니다. 사용자명이 case-insensitive unique여야 하는데 case-sensitive collation으로 만들면 `Rody`와 `rody`가 둘 다 가입할 수 있습니다. 나중에 정책을 바꾸면 둘 중 누가 원래 계정을 가져갈지 결정해야 합니다.

```text
뒤늦은 unique 정책 변경

현재 데이터:
  username='Rody'
  username='rody'

새 정책:
  대소문자 무시 unique

충돌:
  두 row가 같은 logical username으로 접힘

필요한 결정:
  merge?
  하나 rename?
  둘 다 유지하되 로그인 정책만 변경?
```

이런 결정은 DB migration만으로 닫히지 않습니다. 고객 안내, 로그인 충돌 처리, 세션/권한 이관, 외부 연동까지 이어질 수 있습니다. 그래서 collation과 normalization 정책은 가입 기능 초기에 고정하는 편이 훨씬 싸게 먹힙니다.

collation은 prefix index나 range scan에도 연결됩니다. `WHERE name >= 'a' AND name < 'b'` 같은 조건은 collation이 정한 순서에 의존합니다. 사람이 생각하는 알파벳 범위와 DB collation의 sort weight가 다르면 예상과 다른 값이 포함되거나 빠질 수 있습니다. 검색 prefix는 보통 `LIKE 'abc%'`나 specialized index를 쓰지만, 어떤 방식이든 비교 규칙을 확인해야 합니다.

```text
range comparison 질문

1. 문자열 범위가 업무적으로 의미 있는가?
2. collation 정렬 순서가 그 범위를 뒷받침하는가?
3. 대소문자/악센트 무시가 prefix 검색에서 기대와 맞는가?
4. index가 같은 collation으로 만들어졌는가?
```

한국어 문자열에서는 초성 검색, 자모 분해, 정규화 문제가 추가됩니다. 단순 collation만으로 “ㄱ으로 시작하는 이름” 같은 검색 요구를 만족하지 못할 수 있습니다. 이 경우 별도 검색 컬럼, ngram/full-text/search engine을 검토해야 합니다. RDBMS collation은 문자열 비교와 정렬의 기본 규칙이지 모든 검색 UX를 해결하는 도구는 아닙니다.

타입과 collation 관측을 자동화하는 작은 체크도 만들 수 있습니다.

```sql
-- MySQL 예: 특정 schema의 문자열 컬럼 정책 훑기
SELECT table_name, column_name, data_type, character_set_name, collation_name
FROM information_schema.columns
WHERE table_schema = DATABASE()
  AND character_set_name IS NOT NULL
ORDER BY table_name, ordinal_position;
```

이 쿼리는 schema 안의 문자열 컬럼이 어떤 character set과 collation을 쓰는지 보여 줍니다. 같은 서비스의 email 컬럼들이 서로 다른 collation을 쓰고 있으면 drift 신호일 수 있습니다. 물론 모든 컬럼이 같아야 한다는 뜻은 아닙니다. 이름 검색 컬럼과 token 컬럼은 정책이 다를 수 있습니다. 중요한 것은 차이가 의도된 것인지 확인하는 것입니다.

```text
schema drift 예

users.email:
  utf8mb4_0900_ai_ci

admins.email:
  utf8mb4_bin

질문:
  관리자 email은 대소문자를 구분해야 하는가?
  아니면 테이블 생성 시 기본값이 달라져 생긴 우연인가?
```

이런 질문을 정기적으로 던지면 collation drift로 인한 장애를 줄일 수 있습니다.

마지막 실무 검증은 migration 전후 결과 비교입니다.

```sql
-- migration 전후 같은 fixture를 넣고 비교한다.
SELECT name FROM people ORDER BY name;
SELECT name, count(*) FROM people GROUP BY name ORDER BY name;
SELECT count(DISTINCT name) FROM people;
```

PASS는 migration 전후에 의도한 차이와 의도하지 않은 차이를 분리해 설명하는 것입니다. 예를 들어 대소문자 무시 정책으로 바꾸면서 distinct count가 줄어드는 것은 의도일 수 있습니다. 하지만 외부 거래 id distinct count가 줄어들면 위험입니다. FAIL은 character set 변경을 “문자 깨짐 방지”로만 보고 equality와 unique와 order 변화 검증을 생략하는 것입니다.

type과 collation은 API contract와도 연결됩니다. DB 컬럼은 `varchar(20)`인데 API 문서에는 “최대 20자”라고 적혀 있을 수 있습니다. 이때 20자가 DB 기준의 character length인지, UI 기준의 글자 수인지, 외부 기관 전문의 byte length인지 확인해야 합니다. 특히 금융·공공·레거시 연동에서는 “20자리”가 byte length를 뜻하는 경우도 있습니다. 한글이 들어가면 byte 수가 늘어나므로 DB에는 들어가도 외부 전문에는 못 보낼 수 있습니다.

```text
외부 전문 길이 함정

DB:
  customer_name varchar(20)
  값: '홍길동홍길동홍길동홍길동' 12글자

UTF-8 byte:
  한글 12글자 * 3 bytes = 36 bytes

외부 전문:
  이름 필드 20 bytes 제한

결과:
  DB에는 들어가지만 외부 전송은 실패하거나 잘림
```

따라서 type 검증은 DB 안에서만 끝나지 않습니다. 값이 어느 시스템으로 나가고, 그 시스템은 byte를 세는지 문자를 세는지, encoding이 무엇인지 확인해야 합니다. DB deep-dive에서 charset을 배우는 이유도 여기에 있습니다. 저장은 성공했는데 외부 연동에서 깨지는 문제를 막으려면 DB, 애플리케이션, wire format의 문자 기준을 함께 봐야 합니다.

문자 정규화도 중요합니다. Unicode에는 사람이 보기에는 같은 글자처럼 보이지만 내부 표현이 다른 경우가 있습니다. 예를 들어 `é`는 하나의 precomposed code point일 수도 있고, `e`와 combining accent의 조합일 수도 있습니다. collation이나 normalization 정책에 따라 같게 비교될 수도 있고 다르게 남을 수도 있습니다. DB collation이 어느 수준까지 정규화를 고려하는지는 제품과 설정에 따라 다릅니다. 사용자 ID나 검색 키에서 이런 차이가 중요하면 애플리케이션에서 normalization을 수행하고 DB에는 normalized value를 저장하는 방식을 검토해야 합니다.

```text
Unicode normalization 질문

보이는 값:
  é

가능한 표현:
  U+00E9
  U+0065 + U+0301

질문:
  DB collation은 둘을 같게 보는가?
  unique constraint는 둘을 막는가?
  검색 input도 같은 normalization을 거치는가?
```

이 문제는 한국어 자모 조합, 일본어 전각/반각, 이모지 variation selector에서도 비슷하게 나타납니다. 모든 서비스를 완벽한 Unicode 처리 시스템으로 만들 필요는 없지만, 로그인 ID, 닉네임, 외부 코드처럼 중복과 검색이 중요한 값에서는 정책을 정해야 합니다.

타입 비교는 `BETWEEN`에서도 함정을 만듭니다.

```sql
-- timestamp 컬럼에서 하루 조회를 이렇게 쓰면 끝 경계가 애매하다.
WHERE created_at BETWEEN TIMESTAMP '2026-05-19 00:00:00'
                     AND TIMESTAMP '2026-05-19 23:59:59'
```

timestamp precision이 microsecond나 nanosecond까지 있으면 `23:59:59.500` 같은 값이 빠질 수 있습니다. 더 안전한 패턴은 half-open range입니다.

```sql
WHERE created_at >= TIMESTAMP '2026-05-19 00:00:00'
  AND created_at <  TIMESTAMP '2026-05-20 00:00:00'
```

이 예는 타입과 comparison이 실무 장애로 바로 이어지는 지점입니다. 날짜를 문자열처럼 생각하면 하루의 끝을 문자열로 만들고 싶어집니다. timestamp를 연속적인 시간 값으로 보면 시작 포함, 다음 날 시작 미만이 더 자연스럽습니다.

collation과 timezone은 서로 다른 주제지만 “표시와 비교의 기준”이라는 점에서 비슷한 사고가 필요합니다. 화면에 보이는 문자열과 DB가 비교하는 문자열이 다를 수 있고, 화면에 보이는 날짜와 DB가 저장한 timestamp 기준이 다를 수 있습니다. 사용자가 보는 표현을 그대로 DB 비교 기준이라고 믿으면 장애가 납니다. DB에는 비교 가능한 canonical value를 저장하고, 화면에는 locale/timezone에 맞게 표현하는 경계를 둬야 합니다.

```text
canonical value와 display value

저장:
  created_at = 2026-05-19T00:30:00Z

한국 사용자 표시:
  2026-05-19 09:30 KST

미국 사용자 표시:
  2026-05-18 17:30 PDT

조회 조건:
  사용자의 local day를 UTC 범위로 변환한 뒤 timestamp 비교
```

이와 같은 경계 감각은 문자열에도 적용됩니다. 원문 표시값과 비교용 정규화값을 분리하면 정책이 명확해집니다.

```sql
CREATE TABLE users (
    user_id bigint PRIMARY KEY,
    email_original varchar(255) NOT NULL,
    email_normalized varchar(255) NOT NULL UNIQUE
);
```

`email_original`은 사용자가 입력한 표시값이고, `email_normalized`는 로그인과 중복 검사용 값입니다. 이 설계는 저장 공간을 더 쓰지만, 원문 보존과 비교 정책을 분리합니다. 모든 서비스에 필요한 것은 아니지만, 비교 정책이 중요할 때 강력한 선택지입니다.

마지막으로 DU06의 검증 루프를 하나로 묶으면 다음과 같습니다.

```text
1. schema에서 type/charset/collation을 확인한다.
2. fixture에 대소문자, 악센트, 한글, 이모지, trailing space, 숫자 문자열을 넣는다.
3. equality, DISTINCT, GROUP BY, ORDER BY, LIKE 결과를 확인한다.
4. unique 충돌 후보를 새 정책 기준으로 미리 찾는다.
5. EXPLAIN으로 비교 조건이 index를 쓰는지 확인한다.
6. 외부 API나 전문이 있으면 byte/character 길이 기준을 따로 확인한다.
```

이 여섯 단계가 닫히면 문자열 비교 문제를 감으로 다루지 않게 됩니다. DB가 어떤 글자를 저장할 수 있는지, 어떤 값을 같다고 보는지, 어떤 순서로 정렬하는지, 그 규칙이 index와 unique와 외부 연동에 어떤 영향을 주는지 재현할 수 있습니다.

## placeholder, escaping, prepared statement

SQL injection을 설명할 때 “작은따옴표를 escape하면 된다”라고 말하는 경우가 많습니다. 이 말은 위험하게 반만 맞습니다. 문자열 literal 안의 작은따옴표를 올바르게 escape하는 것은 SQL 문법을 깨지 않게 하는 데 필요할 수 있습니다. 하지만 prepared statement의 parameter binding은 단순 escape와 다른 경계입니다. JDBC `PreparedStatement` 공식 API는 SQL 문장이 precompiled되어 `?` placeholder를 가질 수 있고, setter 메서드로 parameter 값을 전달한다고 설명합니다. MySQL 공식 문서도 server-side prepared statement가 prepare 단계와 execute 단계로 나뉘며 parameter marker를 사용할 수 있음을 설명합니다. 이 구조의 핵심은 SQL 코드와 값이 같은 문자열로 이어붙지 않는다는 점입니다.

취약한 문자열 결합부터 봅니다.

```java
String email = request.getParameter("email");
String sql = "SELECT user_id FROM users WHERE email = '" + email + "'";
```

정상 입력이면 SQL은 이렇게 됩니다.

```sql
SELECT user_id FROM users WHERE email = 'a@example.com'
```

공격 입력이 들어오면 SQL 구조가 바뀝니다.

```text
email 입력:
  ' OR '1'='1

만들어진 SQL:
  SELECT user_id FROM users WHERE email = '' OR '1'='1'

의미:
  email이 빈 문자열이거나 항상 참인 조건
```

여기서 문제는 email 값이 문자열 literal로 남지 않고 SQL 문법 일부가 되었다는 점입니다. 작은따옴표가 닫히고, `OR` 연산자가 생기고, 새로운 조건이 붙었습니다. escape는 이 작은따옴표를 literal 내부 문자로 바꾸려는 방법입니다.

```text
입력:
  O'Reilly

올바른 문자열 literal 예:
  'O''Reilly'
```

하지만 escape를 직접 구현하는 것은 어렵습니다. DB별 문자열 literal 문법, character set, backslash escape 모드, 클라이언트 라이브러리, multi-byte 문자, identifier quoting과 literal quoting 차이, LIKE wildcard escape가 모두 얽힙니다. 그리고 escape가 값 literal을 안전하게 만드는 데 도움이 되더라도, 테이블명, 컬럼명, ORDER BY 방향 같은 SQL 구조 조각은 parameter로 bind할 수 없는 경우가 많습니다. 이 영역은 whitelist로 별도 처리해야 합니다.

prepared statement는 경계를 다르게 만듭니다.

```java
PreparedStatement ps = connection.prepareStatement(
    "SELECT user_id FROM users WHERE email = ?"
);
ps.setString(1, email);
ResultSet rs = ps.executeQuery();
```

이 흐름을 trace로 보면 차이가 분명합니다.

```text
문자열 결합 방식:
  Java string + user input
    -> SQL text 하나 생성
    -> DB parser가 전체 text를 SQL 코드로 파싱
    -> 공격 입력이 SQL 구조로 해석될 수 있음

prepared statement 방식:
  SQL template: SELECT ... WHERE email = ?
    -> DB/driver가 SQL 구조를 prepare
  parameter value: "' OR '1'='1"
    -> 값으로 bind
    -> parser가 parameter 값을 SQL 코드로 다시 파싱하지 않음
```

중요한 문장은 “placeholder는 SQL 구조의 구멍이고, bind 값은 그 구멍에 들어가는 데이터”입니다. bind 값은 테이블명이나 컬럼명이나 연산자로 변하지 않습니다. 그래서 prepared statement는 SQL injection 방어의 기본입니다. 물론 driver가 실제 서버 prepared statement를 쓰는지, client-side emulation을 쓰는지, statement cache를 어떻게 관리하는지는 환경마다 다를 수 있습니다. 그래도 안전한 API는 SQL 코드와 값을 분리하는 계약을 제공합니다.

parameter binding은 타입 정보와도 연결됩니다.

```java
ps.setLong(1, userId);
ps.setString(2, email);
ps.setBigDecimal(3, amount);
```

이 setter들은 단지 문자열 치환을 하는 것이 아닙니다. driver는 각 값을 DB에 맞는 타입으로 전달합니다. 숫자는 숫자로, 문자열은 문자열로, 날짜는 날짜/시간 타입으로 전달되어야 합니다. 모든 값을 문자열로 이어붙이면 DB가 암묵 변환을 하거나, locale/format 문제가 생기거나, index 사용이 흔들릴 수 있습니다. 특히 날짜와 금액은 문자열 결합으로 처리하면 장애가 자주 생깁니다.

```java
// 약한 형태
String sql = "SELECT * FROM orders WHERE created_at >= '" + fromDate + "'";

// 더 강한 형태
PreparedStatement ps = connection.prepareStatement(
    "SELECT * FROM orders WHERE created_at >= ?"
);
ps.setDate(1, java.sql.Date.valueOf(fromDate));
```

prepared statement에는 성능 이점도 있을 수 있습니다. SQL 구조가 같고 값만 바뀌면 DB나 driver가 parse/plan 일부를 재사용할 수 있습니다. 하지만 “prepared statement를 쓰면 항상 빠르다”는 단정은 위험합니다. DBMS는 parameter 값에 따라 generic plan과 custom plan을 고를 수 있고, 분포가 치우친 컬럼에서는 parameter 값에 따라 좋은 계획이 달라질 수 있습니다. MySQL, PostgreSQL, Oracle 등은 prepared statement와 plan cache 전략이 다릅니다. 이 섹션의 핵심은 보안 경계이고, 성능은 공식 문서와 `EXPLAIN`, 실제 workload로 확인해야 합니다.

placeholder는 값 위치에만 쓸 수 있는 경우가 대부분입니다. 다음은 작동하지 않거나 위험한 형태입니다.

```java
// 테이블명은 값 parameter로 bind할 수 없다.
PreparedStatement ps = connection.prepareStatement(
    "SELECT * FROM ? WHERE user_id = ?"
);
```

동적으로 테이블이나 정렬 컬럼을 선택해야 한다면 whitelist가 필요합니다.

```java
Set<String> allowedSortColumns = Set.of("created_at", "amount", "order_id");
String sort = request.getParameter("sort");
if (!allowedSortColumns.contains(sort)) {
    throw new IllegalArgumentException("unsupported sort");
}

String sql = "SELECT order_id, amount FROM orders ORDER BY " + sort + " DESC";
PreparedStatement ps = connection.prepareStatement(sql);
```

여기서 `sort`는 값이 아니라 SQL identifier 조각입니다. parameter binding으로 해결할 수 없으므로 허용 목록으로 구조를 제한합니다. 방향도 `ASC`/`DESC` whitelist로 처리해야 합니다. 사용자가 준 문자열을 그대로 붙이면 prepared statement를 일부 써도 injection 경계가 열립니다.

LIKE 검색에서는 wildcard도 별도 정책입니다.

```java
PreparedStatement ps = connection.prepareStatement(
    "SELECT * FROM users WHERE name LIKE ? ESCAPE '\\\\'"
);
ps.setString(1, "%" + escapeLike(userInput) + "%");
```

여기서 parameter binding은 SQL 구조와 값을 분리합니다. 하지만 `%`와 `_`는 LIKE pattern 안에서 wildcard 의미를 갖습니다. 사용자가 입력한 `%`를 문자 그대로 찾을지 wildcard로 허용할지는 별도 정책입니다. 따라서 LIKE pattern escape는 SQL injection 방어와 다른 문제입니다. injection 방어는 binding이 맡고, 검색 wildcard 정책은 pattern escaping이 맡습니다. 이 둘을 섞으면 “prepared statement를 쓰니 wildcard도 안전하겠지”라는 오해가 생깁니다.

quote와 escape의 차이를 정리하면 다음과 같습니다.

| 구분 | 목적 | 위험 |
|---|---|---|
| quoting | 값을 SQL literal로 표현하기 위해 따옴표로 감쌈 | identifier quoting과 혼동하기 쉬움 |
| escaping | literal 내부의 특수 문자를 literal 문자로 남김 | DB별 문법과 charset에 취약 |
| parameter binding | SQL 구조와 값을 분리해 전달 | 구조 조각에는 사용할 수 없음 |
| whitelist | 컬럼명, 방향, 테이블명 같은 구조 선택 제한 | 목록 누락/과허용 관리 필요 |

prepared statement를 쓰더라도 transaction과 권한은 별도입니다. injection을 막았다고 해서 사용자가 접근하면 안 되는 row를 볼 수 없어지는 것은 아닙니다. `WHERE tenant_id = ?` 같은 tenant 경계를 빠뜨리면 bind를 잘해도 정보 유출이 생깁니다. DB 계정 권한이 과도하면 injection이 뚫렸을 때 피해가 커집니다. 따라서 parameter binding은 필수 방어선이지만, 최소 권한, row-level 정책, query review, audit log와 함께 봐야 합니다.

```sql
-- tenant 경계가 빠진 쿼리
SELECT order_id, amount
FROM orders
WHERE order_id = ?;

-- tenant 경계를 함께 둔 쿼리
SELECT order_id, amount
FROM orders
WHERE tenant_id = ?
  AND order_id = ?;
```

prepared statement의 observability도 필요합니다. 운영 로그에는 parameter 값이 마스킹되어야 할 수 있습니다. 하지만 SQL template, 실행 시간, affected rows, error code, slow query plan은 관측해야 합니다. parameter를 모두 포함한 SQL을 로그로 남기면 개인정보와 secret이 유출될 수 있습니다. 반대로 아무 정보도 없으면 장애를 분석할 수 없습니다. 좋은 로그는 구조와 성능 신호를 남기고 민감 값은 숨깁니다.

```text
좋은 관측 예

sql_template:
  SELECT order_id FROM orders WHERE tenant_id = ? AND order_id = ?

params:
  tenant_id = 17
  order_id = [masked or classified]

duration_ms:
  123

affected_rows:
  1

error_code:
  duplicate_key / timeout / deadlock 등 의미 있는 분류
```

JDBC에서는 `PreparedStatement`를 닫는 것도 중요합니다. statement와 result set은 DB cursor, server resource, network resource와 연결될 수 있습니다. try-with-resources를 사용하면 누수를 줄일 수 있습니다.

```java
String sql = "SELECT user_id FROM users WHERE email = ?";

try (PreparedStatement ps = connection.prepareStatement(sql)) {
    ps.setString(1, email);
    try (ResultSet rs = ps.executeQuery()) {
        while (rs.next()) {
            long userId = rs.getLong("user_id");
            // use userId
        }
    }
}
```

이 코드는 보안만이 아니라 자원 생명주기도 보여 줍니다. prepare, bind, execute, consume, close가 하나의 흐름입니다. connection pool을 쓰는 서비스에서는 statement/result set 누수가 connection 반환 지연, cursor 증가, DB resource 압박으로 이어질 수 있습니다.

prepared statement와 batch도 구분해야 합니다.

```java
PreparedStatement ps = connection.prepareStatement(
    "INSERT INTO events(event_id, payload) VALUES (?, ?)"
);

for (Event event : events) {
    ps.setLong(1, event.id());
    ps.setString(2, event.payload());
    ps.addBatch();
}
ps.executeBatch();
```

batch는 같은 SQL 구조에 여러 parameter set을 묶어 보내는 방식입니다. 네트워크 왕복을 줄일 수 있지만, 실패 시 어느 row까지 성공했는지, transaction으로 묶을지, duplicate key를 어떻게 처리할지 결정해야 합니다. prepared statement를 쓴다고 batch 실패 semantics가 자동으로 쉬워지는 것은 아닙니다. 보안 경계, 성능 경계, 실패 경계는 각각 확인해야 합니다.

실무 함정은 escaping과 binding을 같은 보안 장치로 착각하는 것입니다. escaping은 문자열 literal을 SQL 문법 안에 안전하게 표현하려는 기술입니다. binding은 SQL 구조와 값을 별도 채널로 전달하는 계약입니다. 가능하면 값은 binding으로 전달하고, SQL 구조 조각은 whitelist로 제한합니다. 직접 escape 함수를 만들어 모든 입력에 적용하는 방식은 DB별 문법과 예외를 놓치기 쉽습니다.

검증 path는 작은 injection contrast로 만들 수 있습니다.

```text
입력:
  email = "' OR '1'='1"

문자열 결합 SQL:
  SELECT user_id FROM users WHERE email = '' OR '1'='1'
  -> 모든 user가 나올 수 있음

prepared statement:
  template = SELECT user_id FROM users WHERE email = ?
  value = "' OR '1'='1"
  -> email 값이 정확히 그 문자열인 row만 찾음
```

PASS는 공격 입력이 SQL 구조를 바꾸지 못한다는 점을 설명하고, 동적 ORDER BY 같은 구조 조각은 whitelist가 필요하다고 말할 수 있는 상태입니다. FAIL은 작은따옴표 escape만 하면 prepared statement와 같은 안전성을 얻는다고 믿거나, prepared statement를 쓰면서도 컬럼명·정렬 방향·테이블명을 사용자 입력으로 붙이는 상태입니다.

마지막으로 prepared statement를 DB 내부 흐름으로 읽으면 다음과 같습니다.

```text
1. prepare
   SQL template을 보낸다.
   DB/driver가 문법을 분석하고 parameter marker 위치를 안다.

2. bind
   각 marker에 타입이 있는 값을 전달한다.
   값은 SQL 코드로 다시 파싱되지 않는다.

3. execute
   DB가 현재 parameter set으로 실행한다.

4. fetch/update count
   SELECT면 ResultSet을 읽고, DML이면 affected rows를 확인한다.

5. close/reuse
   statement resource를 닫거나 cache 정책에 따라 재사용한다.
```

이 흐름에서 보안의 핵심은 1과 2 사이의 경계입니다. SQL 구조는 prepare 단계에서 고정되고, 값은 bind 단계에서 데이터로 들어갑니다. 이 경계가 무너지면 injection이 생깁니다. 이 경계를 유지하면 사용자 입력이 아무리 SQL처럼 생겼어도 SQL 코드가 아니라 값으로 남습니다. 그래서 DB 애플리케이션의 기본 원칙은 간단합니다. 값은 bind하고, 구조는 whitelist로 고르고, 권한은 최소화하고, 실행 결과와 실패 신호는 관측합니다.

prepared statement를 더 정확히 보려면 “placeholder가 대체된다”라는 표현도 조심해야 합니다. 많은 설명에서 “DB가 `?`를 실제 값으로 대체한다”고 말합니다. 초보 설명으로는 이해하기 쉽지만, 이 표현 때문에 문자열 치환처럼 오해할 수 있습니다. 더 안전한 표현은 “SQL 구조 안에 parameter marker가 있고, 실행 시점에 그 marker에 값이 bind된다”입니다. 값이 SQL text 안으로 이어붙어 다시 parse되는 것이 아니라, 별도 parameter로 전달된다고 이해해야 합니다.

```text
나쁜 mental model:
  "SELECT ... WHERE email = ?" 문자열에서 ?를 "' OR '1'='1"로 치환한다.

좋은 mental model:
  SQL 구조:
    SELECT ... WHERE email = [parameter 1]

  parameter 1:
    type=VARCHAR
    value="' OR '1'='1"

  executor:
    email 컬럼 값과 parameter 1의 문자열 값을 비교한다.
```

이 차이는 로그를 볼 때도 중요합니다. 어떤 로깅 라이브러리는 편의를 위해 parameter가 들어간 완성 SQL처럼 문자열을 보여 줍니다. 그 문자열은 사람이 읽기 위한 렌더링일 수 있고, 실제 wire protocol이나 server prepare 흐름과 다를 수 있습니다. 따라서 “로그에 최종 SQL이 이렇게 찍혔다”는 사실만으로 driver가 문자열 결합을 했다고 단정하면 안 됩니다. 반대로 로깅이 보기 좋게 찍힌다고 해서 안전하다고도 단정하면 안 됩니다. 실제 코드가 `PreparedStatement`와 bind API를 쓰는지 확인해야 합니다.

JDBC에서는 parameter index가 1부터 시작합니다. 이 작은 규칙도 실무 버그를 만듭니다.

```java
PreparedStatement ps = connection.prepareStatement(
    "UPDATE users SET email = ? WHERE user_id = ?"
);
ps.setString(1, email);
ps.setLong(2, userId);
```

순서를 바꾸면 타입이 맞지 않아 오류가 나거나, 둘 다 문자열이면 잘못된 값이 들어갈 수 있습니다. named parameter를 지원하는 프레임워크를 쓰면 가독성이 좋아질 수 있지만, 최종적으로 driver가 어떤 positional parameter로 변환하는지 이해해야 합니다. MyBatis, JPA, Spring JDBC 같은 계층도 내부에서 prepared statement를 사용하더라도 동적 SQL 조각을 문자열로 만들 수 있으므로 경계를 확인해야 합니다.

```text
동적 SQL 경계

안전한 값 바인딩:
  WHERE email = #{email}
  -> prepared statement parameter

위험한 문자열 치환:
  ORDER BY ${sort}
  -> SQL text 조각으로 삽입
  -> whitelist 필요
```

위 예는 MyBatis 계열에서 자주 설명되는 차이와 닮았습니다. 구체 문법은 프레임워크마다 다르지만 원리는 같습니다. 값 위치는 bind하고, SQL 구조 위치는 허용 목록으로 제한합니다.

identifier quoting도 escaping과 다릅니다. SQL에서 문자열 literal은 작은따옴표로 감싸는 경우가 많고, identifier는 DB별로 큰따옴표, backtick, bracket 등을 씁니다.

```sql
-- 문자열 literal
WHERE name = 'Alice'

-- identifier quoting 예
SELECT "order"
FROM "user";
```

사용자 입력으로 identifier를 만들 때 quote 함수를 쓰면 안전할 것 같지만, 그래도 구조 선택 문제는 남습니다. 사용자가 아무 컬럼명이나 선택할 수 있으면 민감 컬럼을 정렬하거나 노출할 수 있습니다. quote는 문법 깨짐을 줄일 수 있지만, 업무적으로 허용된 구조인지 판단하지 않습니다. 따라서 identifier는 quote보다 whitelist가 먼저입니다.

prepared statement가 막지 못하는 injection도 있습니다. 예를 들어 SQL template 자체가 사용자 입력으로 선택되는 경우입니다.

```java
String condition = request.getParameter("condition");
String sql = "SELECT * FROM users WHERE " + condition + " AND tenant_id = ?";
PreparedStatement ps = connection.prepareStatement(sql);
ps.setLong(1, tenantId);
```

여기서는 tenant_id 값은 bind되지만, `condition`은 SQL 구조입니다. 공격자는 `1=1 --` 같은 조건을 넣어 tenant 조건을 주석 처리할 수 있습니다. prepared statement를 일부 썼다는 사실이 전체 쿼리를 안전하게 만들지 않습니다. SQL text를 구성하는 모든 조각을 분류해야 합니다.

```text
SQL 구성 조각 분류

값:
  email, user_id, amount, from_date
  -> bind

식별자:
  sort column, selected column, table suffix
  -> whitelist + quote if needed

키워드/방향:
  ASC/DESC, NULLS FIRST/LAST
  -> enum/whitelist

연산자:
  =, <, >, LIKE
  -> 허용된 검색 조건 DSL로 제한

전체 조건문:
  사용자 입력으로 직접 받지 않음
```

검색 조건이 복잡해지면 작은 query builder나 criteria API가 필요할 수 있습니다. 목적은 멋진 abstraction이 아니라, 사용자 입력이 SQL 구조로 직접 흐르지 않게 하는 것입니다. 각 필터는 허용된 컬럼, 허용된 연산자, bind 값으로 분해되어야 합니다.

```text
검색 필터 DSL 예

입력 JSON:
  {"field":"amount","op":">=","value":100000}

검증:
  field in [amount, created_at, status]
  op allowed for amount: [=, >=, <=]
  value type: numeric

생성:
  amount >= ?
  bind: 100000
```

이 방식은 SQL injection뿐 아니라 타입 오류와 인덱스 정책도 함께 관리합니다. `amount`에는 숫자 연산만 허용하고, `status`에는 enum equality만 허용하고, `created_at`에는 range 조건을 허용하는 식입니다.

prepared statement는 DDL에는 제한적으로만 맞습니다. `CREATE TABLE ?`처럼 테이블명을 parameter로 넘길 수 없고, migration 도구는 SQL 구조를 생성해야 합니다. 그래서 migration SQL은 더 강한 review와 고정된 source control이 필요합니다. 사용자 입력으로 DDL을 만들면 훨씬 위험합니다. 운영 도구에서 tenant별 table을 만들거나 partition을 만들 때도 identifier whitelist, naming convention, transaction/rollback 계획을 둬야 합니다.

escaping은 여전히 필요한 곳이 있습니다. 예를 들어 SQL literal이 아니라 LIKE pattern 내부에서 `%`와 `_`를 문자 그대로 찾고 싶을 때는 pattern escaping이 필요합니다. CSV나 JSON으로 SQL을 생성하는 migration script에서도 DB 공식 quote 함수나 driver 기능을 써야 할 수 있습니다. 하지만 이때도 escape는 “마지막 수단”에 가깝고, 일반 애플리케이션 값 전달은 bind가 기본입니다.

```text
LIKE pattern 정책 trace

사용자 입력:
  "100% cotton"

의도 A: %를 wildcard로 허용
  pattern = "%100% cotton%"
  -> 100 뒤 아무 문자열 허용

의도 B: %를 문자 그대로 검색
  pattern = "%100\\% cotton%" ESCAPE '\\'
  -> 실제 % 문자를 포함한 문자열 검색

공통:
  pattern 문자열 자체는 prepared statement parameter로 bind
```

이 trace가 보여 주듯 wildcard escaping과 SQL injection 방어는 층위가 다릅니다. wildcard 정책을 정한 뒤, 최종 pattern 값은 bind합니다.

prepared statement의 성능 관측은 세 부분으로 나눕니다. 첫째, parse/prepare 비용이 줄었는지. 둘째, plan이 재사용되는지 또는 parameter별로 적절히 다시 계획되는지. 셋째, statement cache가 connection pool과 맞게 동작하는지. 같은 SQL을 자주 실행하는 OLTP 서비스에서는 이점이 있을 수 있지만, 아주 다양한 동적 SQL을 무한히 prepare하면 cache pressure가 생길 수 있습니다.

```text
성능 관측 질문

1. 같은 SQL template이 반복 실행되는가?
2. DB/driver/server가 prepared statement를 cache하는가?
3. parameter 값 분포가 plan 선택에 큰 영향을 주는가?
4. generic plan이 특정 parameter에서 느려지지 않는가?
5. connection pool의 connection마다 statement cache가 분산되는가?
```

이 질문의 답은 DB와 driver마다 다릅니다. 따라서 prepared statement의 보안 이점은 기본으로 가져가되, 성능 주장은 실제 metric과 plan으로 닫아야 합니다.

affected rows 확인도 prepared statement 흐름의 일부입니다.

```java
PreparedStatement ps = connection.prepareStatement(
    "UPDATE stocks SET quantity = quantity - ? " +
    "WHERE product_id = ? AND quantity >= ?"
);
ps.setInt(1, requested);
ps.setLong(2, productId);
ps.setInt(3, requested);
int updated = ps.executeUpdate();
```

`updated`가 1이면 재고 차감이 성공했습니다. 0이면 상품이 없거나 재고가 부족했을 수 있습니다. prepared statement를 안전하게 썼다는 사실만으로 비즈니스 성공이 아닙니다. 실행 결과를 읽어야 합니다. SQL injection 방어, 타입 바인딩, 도메인 상태 전이는 서로 다른 게이트입니다.

```text
UPDATE 실행 trace

template:
  UPDATE stocks SET quantity = quantity - ? WHERE product_id = ? AND quantity >= ?

bind:
  requested=2, product_id=10, requested=2

execute result:
  affected_rows=0

business decision:
  재고 부족 또는 대상 없음
  주문 확정하면 안 됨
```

보안 검증에서는 공격 문자열을 실제 fixture로 넣어야 합니다.

```java
String attack = "' OR '1'='1";

PreparedStatement ps = connection.prepareStatement(
    "SELECT count(*) FROM users WHERE email = ?"
);
ps.setString(1, attack);
```

PASS는 count가 전체 사용자 수가 아니라 그 문자열을 email로 가진 row 수라는 점입니다. 보통 0입니다. FAIL은 전체 row가 반환되는 것입니다. 문자열 결합 버전과 prepared statement 버전을 나란히 테스트하면 차이가 눈에 보입니다.

마지막으로 권한 최소화까지 연결합니다. injection이 완전히 불가능하다고 장담하는 대신, 뚫렸을 때 피해를 줄이는 설계를 합니다. 읽기 API는 읽기 전용 계정을 쓰고, 쓰기 API도 필요한 table과 statement만 허용하고, migration 계정은 별도로 둡니다. prepared statement는 공격 입력이 SQL 구조로 변하는 것을 막는 강한 방어선이지만, 계정 권한이 모든 table drop 권한을 갖고 있으면 방어선 하나가 뚫렸을 때 피해가 큽니다.

```text
계층 방어

1. 값은 parameter binding
2. 구조 조각은 whitelist
3. DB 계정은 최소 권한
4. 민감 값은 로그에서 masking
5. affected rows와 error code를 확인
6. slow query와 실패율을 관측
```

이 여섯 줄이 prepared statement를 실무 보안/운영 문맥에 놓는 기준입니다. escaping은 이 중 일부 상황에서 필요한 보조 도구일 뿐, parameter binding의 대체물이 아닙니다.

prepared statement와 stored prepared statement도 구분할 필요가 있습니다. JDBC의 `PreparedStatement` API를 쓴다고 해서 항상 서버에 명시적인 `PREPARE` 문이 생성되는 것은 아닐 수 있습니다. driver 설정에 따라 client-side emulation을 하거나, server-side prepare를 쓰거나, threshold 이후 prepare를 쓰는 식으로 다를 수 있습니다. 하지만 애플리케이션 코드 관점의 중요한 계약은 동일합니다. 개발자는 SQL 구조와 parameter 값을 분리된 API로 전달합니다. 성능과 server resource 관점의 세부는 driver와 DB 공식 문서로 확인해야 합니다.

```text
두 층의 질문

보안/코드 계약:
  값을 문자열 결합하지 않고 parameter API로 전달하는가?

성능/리소스 계약:
  driver가 server-side prepare를 쓰는가?
  statement cache가 어떻게 동작하는가?
  DB 서버에 prepared statement resource가 쌓이지 않는가?
```

이 두 질문을 섞으면 논의가 흐려집니다. “우리 driver는 client-side prepare라서 prepared statement가 의미 없다”라고 말하는 것은 성급할 수 있습니다. 안전한 escaping과 protocol 처리를 driver가 맡는 구조라면 문자열 결합보다는 여전히 강합니다. 반대로 “PreparedStatement를 쓰니 plan cache 성능이 항상 좋다”도 성급합니다. 성능은 별도 관측이 필요합니다.

parameter type을 잘못 고르면 index 사용이 흔들릴 수 있습니다. 예를 들어 DB 컬럼은 numeric인데 문자열로 bind하거나, timestamp 컬럼에 문자열을 bind하면 DB가 암묵 변환을 해야 할 수 있습니다. JDBC setter를 구체적으로 쓰는 이유는 단지 컴파일 편의가 아니라 값의 타입 의미를 보존하기 위해서입니다.

```java
// 약한 형태: 모든 값을 문자열로 전달
ps.setString(1, "100000");

// 더 강한 형태: 금액 정책에 맞는 타입으로 전달
ps.setBigDecimal(1, new BigDecimal("100000.00"));
```

DB가 알아서 변환해 줄 수 있지만, 변환 규칙은 DB별로 다르고 인덱스와 plan에 영향을 줄 수 있습니다. 특히 날짜 문자열은 locale과 timezone 문제가 섞입니다. 애플리케이션 경계에서 타입을 명확히 만들고 bind해야 합니다.

prepared statement는 SQL injection을 줄이지만 business injection을 막지는 않습니다. 사용자가 `amount=-100`을 보내면 prepared statement는 그 값을 안전하게 숫자로 전달할 수 있습니다. 하지만 음수 금액이 허용되지 않는다는 도메인 검증은 별도입니다. DB `CHECK(amount > 0)`와 애플리케이션 validation이 필요합니다.

```text
보안 경계 분리

SQL injection:
  입력이 SQL 구조가 되는 문제
  -> parameter binding

business invalid input:
  입력 값 자체가 도메인 규칙을 어기는 문제
  -> validation + DB constraint

authorization failure:
  사용자가 접근하면 안 되는 row를 요청
  -> tenant/user scope condition + 권한 검사
```

세 문제를 모두 “입력 검증”이라고 부르면 해결책이 섞입니다. prepared statement는 첫 번째 문제의 핵심 방어선입니다. 두 번째와 세 번째는 다른 장치가 필요합니다.

동적 `IN` 목록도 자주 실수합니다.

```java
// 위험하거나 어색한 형태
String ids = request.getParameter("ids"); // "1,2,3"
String sql = "SELECT * FROM orders WHERE order_id IN (" + ids + ")";
```

목록 길이가 동적이면 placeholder를 개수만큼 만들어야 합니다.

```java
List<Long> ids = List.of(1L, 2L, 3L);
String placeholders = ids.stream()
    .map(id -> "?")
    .collect(Collectors.joining(", "));

PreparedStatement ps = connection.prepareStatement(
    "SELECT * FROM orders WHERE order_id IN (" + placeholders + ")"
);
for (int i = 0; i < ids.size(); i++) {
    ps.setLong(i + 1, ids.get(i));
}
```

또는 DB/driver가 array parameter를 지원하면 그 방식을 쓸 수 있습니다. 핵심은 comma-separated string을 SQL 조각으로 붙이지 않는 것입니다. 숫자만 온다고 믿어도 공백, 주석, overflow, 빈 목록, 타입 변환 문제가 남습니다. 빈 목록은 `WHERE false`로 처리할지, 쿼리를 실행하지 않을지 별도 정책을 둬야 합니다.

```text
동적 IN 검증 질문

1. 목록 원소 타입이 검증됐는가?
2. 원소마다 placeholder가 있는가?
3. 빈 목록은 어떤 SQL로 바뀌는가?
4. 목록 길이가 너무 길 때 plan/cache/packet 제한은 없는가?
5. 순서가 의미 있으면 결과 ORDER를 어떻게 보존하는가?
```

대량 입력에서는 temporary table이나 bulk insert 후 join이 더 나을 수 있습니다. prepared statement 하나로 모든 문제를 해결하려고 하지 말고, 데이터 양과 DB 기능에 맞게 선택해야 합니다.

error handling도 prepared statement 사용의 일부입니다. unique violation, foreign key violation, deadlock, timeout은 모두 SQLException으로 올 수 있지만 의미가 다릅니다. SQLState나 vendor error code를 분류해 재시도 가능한 오류와 사용자 입력 오류와 시스템 장애를 구분해야 합니다.

```text
오류 분류 예

duplicate key:
  이미 존재하는 값
  -> 사용자 안내 또는 idempotent success 검토

deadlock / serialization failure:
  동시성 충돌
  -> transaction 재시도 가능성 검토

lock wait timeout:
  경합 또는 장기 transaction
  -> 재시도와 원인 관측 분리

syntax error:
  SQL template 버그
  -> 재시도해도 해결 안 됨
```

prepared statement는 SQL template을 미리 고정하므로 syntax error는 보통 코드/배포 버그입니다. 반면 duplicate key나 deadlock은 정상 운영 중에도 발생 가능한 domain/concurrency 신호일 수 있습니다. 에러를 모두 같은 500으로 처리하면 사용자 경험과 장애 대응이 나빠집니다.

마지막으로 안전한 SQL 작성 습관을 한 흐름으로 묶어 봅니다.

```text
요구:
  사용자가 선택한 정렬 기준으로 자기 tenant의 주문을 검색한다.

입력:
  tenantId = 17
  status = "READY"
  minAmount = 100000
  sort = "created_at"
  direction = "DESC"

처리:
  tenantId/status/minAmount -> bind value
  sort/direction -> whitelist enum

SQL:
  SELECT order_id, amount, created_at
  FROM orders
  WHERE tenant_id = ?
    AND status = ?
    AND amount >= ?
  ORDER BY created_at DESC, order_id DESC

bind:
  [17, "READY", 100000]

검증:
  tenant 조건 누락 없음
  sort는 허용된 컬럼
  tie-breaker 있음
  affected/result row count 관측
```

이 trace는 prepared statement를 “보안 API 하나”가 아니라 SQL 작성 전체 흐름 안에 놓습니다. 값, 구조, 권한, 정렬 안정성, 관측이 모두 닫혀야 안전합니다. escaping과 prepared statement를 같은 것으로 보는 함정은 이 전체 흐름을 `문자열을 안전하게 만들기` 하나로 축소하기 때문에 생깁니다. 실제 목표는 문자열을 예쁘게 만드는 것이 아니라, 사용자 입력이 SQL 구조와 권한 경계를 넘지 못하게 하는 것입니다.

prepared statement가 등장한 배경은 애플리케이션이 사용자 입력을 SQL 문자열에 직접 섞기 시작하면서 코드와 데이터의 경계가 쉽게 무너졌기 때문입니다. SQL template과 parameter value를 분리하면 DB는 어떤 부분이 명령 구조이고 어떤 부분이 값인지 더 안정적으로 알 수 있습니다. 이 분리는 보안뿐 아니라 타입 전달, 반복 실행, 관측 기준에도 영향을 줍니다.

따라서 코드 리뷰에서 볼 문장은 짧습니다. “값인가, 구조인가?” 값이면 bind합니다. 구조이면 사용자 입력을 그대로 붙이지 않고 허용된 선택지로 바꿉니다. 그다음 “이 SQL이 어떤 권한 범위에서 실행되는가?”를 봅니다. 마지막으로 “성공/실패 신호를 읽는가?”를 봅니다. 이 네 질문이 닫혀야 prepared statement 사용이 실제 안전으로 이어집니다.

### 마지막 실무 연결: 바인딩은 보안과 의미 경계를 동시에 만든다

escaping은 문자열 안의 위험 문자를 SQL 문법 안에서 덜 위험하게 만들려는 처리이고, parameter binding은 SQL 구조와 값을 애초에 분리해 전달하는 방식이다. 둘은 비슷해 보이지만 경계가 다르다. 동적 ORDER BY나 table name처럼 SQL 구조 자체를 바꿔야 하는 곳에는 parameter binding을 그대로 쓸 수 없고, 허용 목록으로 구조를 선택해야 한다. 반대로 값 자리에는 문자열 결합 대신 binding을 써야 한다.

실무 검증은 공격 문자열 하나로 끝나지 않는다. 작은따옴표, backslash, multibyte 문자, LIKE wildcard, numeric-looking string, NULL, 빈 문자열을 모두 넣어 본다. 그리고 SQL log에서 값이 SQL 구조로 섞이지 않고 parameter로 남는지 확인한다. 이 습관이 있어야 prepared statement를 보안 기능이자 의미 경계로 이해할 수 있다.
