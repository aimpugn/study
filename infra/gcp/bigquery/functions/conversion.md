# [conversion](https://cloud.google.com/bigquery/docs/reference/standard-sql/conversion_functions)

- [conversion](#conversion)
    - [`SAFE_CAST`](#safe_cast)
        - [숫자인지 확인하기](#숫자인지-확인하기)

## [`SAFE_CAST`](https://cloud.google.com/bigquery/docs/reference/standard-sql/conversion_functions#safe_casting)

### 숫자인지 확인하기

```sql
WITH `project.dataset.table` AS (
    SELECT '1234.56' col UNION ALL
    SELECT '1234.' col UNION ALL
    SELECT '1234' col UNION ALL
    SELECT '.56' col UNION ALL
    SELECT '1234..56' col UNION ALL
    SELECT 'a1234.56' 
)
SELECT
    col,
    if(SAFE_CAST(col AS FLOAT64) is null,'Not a number', 'A number')
FROM `project.dataset.table`
```
