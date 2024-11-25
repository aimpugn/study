# BigQuery

## Access Denied: BigQuery BigQuery: User has neither fine-grained reader nor masked get permission to get data protected by policy tag "pii_taxonomy : default" on column

### 문제

```bash
Access Denied: BigQuery BigQuery: User has neither fine-grained reader nor masked get permission to get data protected by policy tag "pii_taxonomy : default" on column schema_for_biqeury.payments.apply_num. at [43:1]
```

### 원인

해당 칼럼을 조회할 권한이 없어서 발생

### 해결

해당 칼럼 조회하지 않도록 수정

## Query error: Correlated subqueries that reference other tables are not supported

### 문제

```bash
Query error: Correlated subqueries that reference other tables are not supported unless they can be de-correlated, such as by transforming them into an efficient JOIN. at [43:1]
```

### 원인

BigQuery에서 발생하는 이 에러 메시지는 쿼리 내에 다른 테이블을 참조하는 상관 서브쿼리(correlated subqueries)가 포함되어 있기 때문입니다.
BigQuery는 특정 유형의 상관 서브쿼리를 지원하지 않습니다.
상관 서브쿼리는 외부 쿼리의 컬럼을 내부 쿼리에서 참조하는 것을 말합니다.

### 해결

이 문제를 해결하기 위해서는 상관 서브쿼리를 비상관 서브쿼리(non-correlated subquery)로 변환하거나, JOIN을 사용하여 효율적으로 쿼리를 재작성해야 합니다.

JOIN을 사용하면 서브쿼리가 외부 쿼리의 컬럼에 직접적으로 의존하지 않게 되므로, BigQuery에서 처리할 수 있습니다.

예를 들어, 다음과 같은 상관 서브쿼리가 있다고 가정해 보겠습니다:

```sql
SELECT a.id, 
       (SELECT b.value FROM b WHERE b.id = a.b_id) as value
FROM a
```

이 쿼리는 `a` 테이블의 각 행에 대해 `b` 테이블에서 `b.id`가 `a.b_id`와 같은 `b.value`를 찾습니다. 이는 상관 서브쿼리입니다.

이를 JOIN을 사용하여 다음과 같이 재작성할 수 있습니다:

```sql
SELECT a.id, b.value
FROM a
JOIN b ON b.id = a.b_id
```

이렇게 수정하면, 각 `a` 행에 대해 `b` 테이블을 조인하여 `b.id`가 `a.b_id`와 일치하는 경우에만 `b.value`를 가져오게 됩니다. 이 방식은 BigQuery에서 지원하며, 성능도 더 효율적일 수 있습니다.
