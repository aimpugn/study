# information

- [information](#information)
    - [테이블 사이즈 확인하기](#테이블-사이즈-확인하기)
        - [`TABLE_ROWS`가 실제 `COUNT(1)`과 다르다\\](#table_rows가-실제-count1과-다르다)

## 테이블 사이즈 확인하기

```sql
SELECT
    TABLE_NAME,
    CONCAT(ROUND(DATA_LENGTH / 1024 / 1024 / 1024, 2), ' GB') AS 'DATA_LENGTH_IN_GB',
    CONCAT(ROUND(INDEX_LENGTH / 1024 / 1024 / 1024, 2), ' GB') AS 'INDEX_LENGTH_IN_GB',
    CONCAT(
        ROUND(((DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024 / 1024), 2), ' GB'
    ) AS 'TOTAL_SIZE_IN_GB',
    TABLE_ROWS
FROM
    information_schema.TABLES
WHERE
    TABLE_SCHEMA = 'some_schema'
    AND TABLE_NAME IN ('some_table_in_some_schema');
```

### `TABLE_ROWS`가 실제 `COUNT(1)`과 다르다\

- 이러한 차이는 MySQL의 내부 최적화 메커니즘과 메타데이터 관리 방식에 기인하는 것으로, 일반적으로 큰 문제가 되지 않는다고 한다.
- 그러나 정확한 데이터가 필요한 경우에는 실제 테이블을 직접 쿼리하는 것이 좋다

1. 비동기화된 메타데이터:
   - `information_schema`의 테이블 행 수는 항상 정확하게 최신 상태를 반영하지 않을 수 있습니다. 특히, 큰 테이블의 경우, MySQL은 행 수를 정확하게 계산하기보다는 추정치를 제공할 때가 있습니다.
   - `information_schema`는 데이터베이스의 메타데이터를 저장하는데, 이 메타데이터가 실제 데이터와 완전히 동기화되지 않은 상태일 수 있습니다.
2. 최적화를 위한 추정치 사용:
   - MySQL은 성능 최적화를 위해 `information_schema`에서 행 수를 정확하게 계산하지 않고 추정치를 제공할 수 있습니다.
   - 이 추정치는 인덱스 통계나 다른 내부 메커니즘에 기반하여 계산될 수 있으며, 항상 최신 상태를 반영하지 않습니다.
3. 인덱스 통계의 불일치:
   - `information_schema`의 행 수 정보는 종종 테이블의 인덱스 통계에 기반합니다. 테이블의 인덱스가 최근에 업데이트되지 않았다면, 이러한 통계가 실제 데이터와 일치하지 않을 수 있습니다.
   - `ANALYZE TABLE` 명령을 실행하여 인덱스 통계를 업데이트할 수 있습니다.
4. 트랜잭션과 격리 수준:
   - 다른 트랜잭션에서 행을 추가하거나 삭제하는 작업이 진행 중일 수 있습니다. 트랜잭션 격리 수준에 따라, `COUNT(1)` 쿼리가 실행되는 시점에는 이러한 변경사항이 반영되지 않았을 수도 있습니다.

따라서,

- 실제 행 수를 얻으려면 `COUNT(1)`을 사용하는 것이 가장 정확하다
- `information_schema`에서의 데이터는 참조용으로만 사용하고, 정확한 통계가 필요한 경우는 테이블 자체를 직접 쿼리하는 것이 좋다
- 인덱스 통계를 정기적으로 업데이트하여 `information_schema`의 정확도를 향상시킬 수 있다.
