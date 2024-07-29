# Query

## [correlated subquery](https://en.wikipedia.org/wiki/Correlated_subquery)

상관 서브쿼리(correlated subquery)는 SQL 데이터베이스 쿼리에서 사용되는 서브쿼리의 한 유형으로, 외부 쿼리(outer query)의 값들을 사용합니다.
이 서브쿼리는 외부 쿼리의 각 행을 처리할 때마다 평가될 수 있으며, 이로 인해 실행 속도가 느려질 수 있습니다.
상관 서브쿼리는 외부 쿼리의 컬럼을 내부 쿼리에서 참조하므로, 외부 쿼리의 각 행에 대해 내부 쿼리가 다시 실행되어야 합니다.
즉, 내부 쿼리(서브쿼리)가 외부 쿼리의 결과에 의존적인 관계를 가지고 있습니다.

상관 서브쿼리의 유래는 SQL 언어의 발전과 함께 데이터베이스 시스템에서 복잡한 데이터 관계를 표현하기 위해 발전해왔습니다.
이는 데이터베이스에서 보다 세밀한 데이터 추출과 조건에 따른 데이터 처리를 가능하게 해줍니다.

예를 들어, 특정 부서의 평균 급여보다 많은 급여를 받는 직원을 찾는 쿼리에서 상관 서브쿼리를 사용할 수 있습니다:

```sql
SELECT employee_number, name
FROM employees emp
WHERE salary > (
    SELECT AVG(salary)
    FROM employees
    WHERE department = emp.department
);
```

이 예에서 내부 쿼리는 외부 쿼리의 `department` 컬럼을 참조하여 각 직원의 부서에 대한 평균 급여를 계산합니다.
이러한 유형의 쿼리는 각 부서에 대해 한 번씩 내부 쿼리를 실행할 수 있으며, 데이터베이스에 따라 내부 쿼리 결과를 캐싱하여 성능을 최적화할 수 있습니다.
