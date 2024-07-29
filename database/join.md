# JOIN

## 서브쿼리를 조인

```sql
FROM aaaa
LEFT JOIN bbbb ON bbbb.foreign_id = aaaa.id
LEFT JOIN (
    SELECT cccc.*
    FROM cccc
    JOIN dddd ON cccc.some_uid = dddd.some_uid
    WHERE cccc.foreign_id = aaaa.id
)
```
