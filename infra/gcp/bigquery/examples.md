# Examples

- [Examples](#examples)
    - [`DATETIME_TRUNC`](#datetime_trunc)
        - [`MONTH`](#month)
    - [`EXTRACT`](#extract)
        - [`ISOYEAR`](#isoyear)
    - [`FORMAT_DATETIME`](#format_datetime)

## `DATETIME_TRUNC`

### `MONTH`

```sql
SELECT (COUNT(1) / (SELECT COUNT(1) FROM `schema_for_biqeury.some_table`)) * 100
FROM (
  SELECT DATETIME_TRUNC(p.created, MONTH), p.user_id, p.merchant_uid
  FROM `schema_for_biqeury.some_table` p
  GROUP BY DATETIME_TRUNC(p.created, MONTH), p.user_id, p.merchant_uid HAVING COUNT(1) > 15
) p;
```

## `EXTRACT`

### `ISOYEAR`

```sql
SELECT CONCAT(EXTRACT(ISOYEAR FROM p.created), '-', EXTRACT(MONTH FROM p.created)) AS ym, p.user_id, p.merchant_uid, COUNT(1) AS merchant_uid_cnt
FROM `schema_for_biqeury.some_table` p
GROUP BY ym, p.user_id, p.merchant_uid HAVING COUNT(1) > 15
ORDER BY ym DESC;
```

## `FORMAT_DATETIME`

```sql
SELECT
  large_retries_per_month.yearmonth,
  large_retries_per_month.count AS large_retry_count,
  count_per_month.count AS total_count,
  large_retries_per_month.count / count_per_month.count * 100 AS percentage
FROM (
  SELECT yearmonth, COUNT(1) AS count
  FROM (
    SELECT FORMAT_DATETIME("%Y-%m", created) AS yearmonth, user_id, merchant_uid, COUNT(1) AS count
    FROM `schema_for_biqeury.some_table`
    GROUP BY yearmonth, user_id, merchant_uid HAVING COUNT(1) > 15
  )
  GROUP BY yearmonth
) AS large_retries_per_month
INNER JOIN (
  SELECT FORMAT_DATETIME("%Y-%m", created) AS yearmonth, COUNT(1) AS count
  FROM `schema_for_biqeury.some_table`
  GROUP BY yearmonth
  ORDER BY yearmonth DESC
) AS count_per_month
ON large_retries_per_month.yearmonth = count_per_month.yearmonth
ORDER BY large_retries_per_month.yearmonth DESC;
```
