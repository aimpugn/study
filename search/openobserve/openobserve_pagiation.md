# OpenObserve Pagination

## `from` 속성 사용

```json
// POST https://openobserve.domain.co/api/default/_search?type=logs
{
    "sql": "SELECT.. query",
    "start_time": 1723504880811000,
    "end_time": 1723526480811000,
    "from": 250,
    "size": 248,
    "fast_mode": true,
    "sql_mode": "full"
}
```
