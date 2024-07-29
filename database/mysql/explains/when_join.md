# When `JOIN`

- [When `JOIN`](#when-join)
    - [`another_table`와  `this_table` 테이블 조인시](#another_table와--this_table-테이블-조인시)
        - [`this_table.id`하고만 `ON`](#this_tableid하고만-on)
        - [`this_table.id` 및 `this_table.user_id` 동시에 `ON`](#this_tableid-및-this_tableuser_id-동시에-on)

## `another_table`와  `this_table` 테이블 조인시

- `JOIN ON` 조건이 다르다고 해도, 결국 query planner가 어떤 인덱스를 고르느냐에 따라 달라진다

### `this_table.id`하고만 `ON`

```sql
EXPLAIN SELECT
    AnotherTable.key_name,
    AnotherTable.key_value
FROM another_table AnotherTable
JOIN this_table This ON This.id = AnotherTable.this_id
WHERE This.another_uniq_id = 'imp_445341175933';
```

```json
[
  {
    "id": 1,
    "select_type": "SIMPLE",
    "table": "This",
    "type": "const",
    "possible_keys": "PRIMARY,another_uniq_id,status_paging_idx,user_created_idx,idx_all_in_one",
    "key": "another_uniq_id",
    "key_len": "99",
    "ref": "const",
    "rows": 1,
    "Extra": ""
  },
  {
    "id": 1,
    "select_type": "SIMPLE",
    "table": "AnotherTable",
    "type": "ref",
    "possible_keys": "idx_This_id",
    "key": "idx_This_id",
    "key_len": "4",
    "ref": "const",
    "rows": 6,
    "Extra": ""
  }
]
```

### `this_table.id` 및 `this_table.user_id` 동시에 `ON`

```sql
EXPLAIN SELECT
    AnotherTable.key_name,
    AnotherTable.key_value
FROM another_table AnotherTable
JOIN this_table This ON This.id = AnotherTable.this_id
    AND This.user_id = 38138
WHERE This.another_uniq_id = 'imp_445341175933';
```

```json
[
  {
    "id": 1,
    "select_type": "SIMPLE",
    "table": "This",
    "type": "const",
    "possible_keys": "PRIMARY,another_uniq_id,status_paging_idx,user_created_idx,idx_all_in_one",
    "key": "another_uniq_id",
    "key_len": "99",
    "ref": "const",
    "rows": 1,
    "Extra": ""
  },
  {
    "id": 1,
    "select_type": "SIMPLE",
    "table": "AnotherTable",
    "type": "ref",
    "possible_keys": "idx_This_id",
    "key": "idx_This_id",
    "key_len": "4",
    "ref": "const",
    "rows": 6,
    "Extra": ""
  }
]
```
