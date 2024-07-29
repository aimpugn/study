# Expectation

## `ExpectPrepare`와 `ExpectQuery` / `ExpectExec`의 차이

- **`ExpectPrepare`**: 특정 SQL 문이 준비될 것을 기대하며, 준비된 문에 대한 참조를 반환한다. 이를 통해 해당 준비된 문에 대해 실행될 쿼리나 명령어에 대한 기대를 설정할 수 있다.

- **`ExpectQuery`와 `ExpectExec`**: `ExpectPrepare`에 의해 반환된 객체를 통해 호출되며, 준비된 SQL 문에 대해 실행될 쿼리(`ExpectQuery`)나 명령어(`ExpectExec`)가 어떤 인자를 받고, 어떤 결과를 반환할지 등의 구체적인 기대를 설정한다.

## `ExpectPrepare`

```go
result := sqlMock.ExpectPrepare(strings.Replace(userByAccessTokenSQL, "?", "\\?", -1))
result.ExpectQuery().
    WithArgs(sqlmock.AnyArg()).
    WillReturnRows(rows)
```
