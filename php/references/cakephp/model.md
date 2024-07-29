# Model

- [Model](#model)
  - [update field validation](#update-field-validation)

## update field validation

- lib/Cake/Model/Datasource/DboSource.php
  - 업데이트 시에 `DboSource::_prepareUpdateFields` 통해서 필드 타입 검증을 거친다
  - 가령 `card_type`은 여기서 string인데, tinyint 타입이 맞지 않기 때문에 NULL로 변환이 된다

```log
Array
(
    [$fields] => Array
        (
            [card_number] => 123456****1234
            [emb_pg_provider] => gateway
            [card_type] =>
            [card_issue_code] =>
            [str_col_7] =>
            [id] => 32785325
            [modified] => 2022-12-15 16:23:52
        )

    [type] => string
)
Array
(
    [0] => `card_number` = '123456****1234'
    [1] => `emb_pg_provider` = 'gateway'
    [2] => `card_type` = NULL
    [3] => `card_issue_code` = NULL
    [4] => `str_col_7` = NULL
    [5] => `id` = 32785325
    [6] => `modified` = '2022-12-15 16:23:52'
)
```
