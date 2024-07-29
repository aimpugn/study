# Placeholder

## Placeholder란?

"Placeholder" 또는 "Parameter Marker"는 데이터베이스 쿼리에서 사용되는, 특정 값이 나중에 지정될 위치를 표시하는 심볼 또는 문자다.
이러한 메커니즘은 SQL 쿼리를 더 안전하고, 유연하며, 재사용 가능하게 만드는 데 도와준다.
주로 SQL Injection과 같은 보안 취약점을 방지하고, 쿼리의 성능을 향상시키는 데에 기여한다.

## 역할 및 사용법

Placeholder는 값을 직접 쿼리 문자열에 포함시키는 대신, 쿼리를 실행할 때 바인딩될 값의 위치를 지정합니다. 이 방식을 사용함으로써, 데이터베이스 엔진은 쿼리를 컴파일한 뒤, 실행 시점에 실제 값으로 placeholder를 대체하여 실행합니다. 이는 다음과 같은 이점을 제공합니다:

- **보안 향상:** 사용자 입력을 직접 쿼리에 포함시키지 않기 때문에, SQL 인젝션 공격을 방지할 수 있습니다.
- **성능 최적화:** 쿼리 구조가 변경되지 않고 값만 변경될 때, 데이터베이스는 쿼리 계획을 재사용할 수 있어, 쿼리 실행 속도가 향상됩니다.
- **코드 가독성 및 유지보수:** 쿼리와 데이터를 분리함으로써 코드의 가독성이 향상되고, 쿼리 수정이 용이해집니다.

## 처리 과정

1. **쿼리 작성:** 개발자는 쿼리를 작성하면서, 데이터가 들어갈 위치에 placeholder를 사용합니다. SQL에서는 주로 `?`가 placeholder로 사용됩니다.
2. **쿼리 준비:** 데이터베이스는 준비된 쿼리(prepared statement)를 생성하고, 쿼리의 구조를 분석하여 실행 계획을 준비합니다.
3. **값 바인딩:** 쿼리 실행 전, 개발자는 쿼리에 바인딩할 실제 데이터를 제공합니다. 데이터베이스는 이 데이터를 placeholder의 위치에 삽입합니다.
4. **쿼리 실행:** 데이터가 삽입된 쿼리가 데이터베이스에 의해 실행됩니다.

## 종류

주로 사용하는 데이터베이스 및 데이터베이스 접근 라이브러리에 따라 달라진다.

### 명명된 Placeholder (Named Placeholders)

일부 데이터베이스 라이브러리나 프레임워크는 명명된 placeholder를 지원하여, 쿼리에서 변수를 더 명확하게 지정할 수 있게 합니다. 이 경우, 각 placeholder는 고유한 이름을 가집니다. 예를 들어, Python의 `sqlite3` 라이브러리 또는 Go의 `sqlx` 패키지에서 사용할 수 있습니다.

- Python `sqlite3` 예시:

    ```python
    cursor.execute("SELECT * FROM users WHERE username=:username", {"username": "exampleUser"})
    ```

- Go `sqlx` 예시:

    ```go
    db.NamedQuery("SELECT * FROM users WHERE username=:username", map[string]interface{}{"username": "exampleUser"})
    ```

### 숫자형 Placeholder (Numeric Placeholders)

일부 SQL 인터페이스는 placeholder로 숫자를 사용하여, 쿼리의 각 바인딩 위치를 숫자로 지정합니다. PostgreSQL의 `pq` 드라이버에서는 `$1`, `$2` 등의 형식을 사용합니다.

- PostgreSQL 예시:

    ```sql
    SELECT * FROM users WHERE username = $1 AND password = $2
    ```

### `@` 기호를 사용하는 Placeholder

MySQL이나 SQL Server와 같은 일부 데이터베이스 시스템에서는 `@` 기호를 사용하는 placeholder를 지원합니다. 이는 주로 저장 프로시저나 특정 쿼리에서 변수를 지정하는 데 사용됩니다.

- MySQL 예시:

    ```sql
    SELECT * FROM users WHERE username = @username AND password = @password
    ```

### 데이터베이스 및 라이브러리에 따른 차이

Placeholder의 종류와 사용법은 사용하는 데이터베이스 시스템, 프로그래밍 언어 및 라이브러리에 따라 크게 달라집니다. 따라서 특정 데이터베이스 또는 라이브러리를 사용할 때는 해당 문서를 참조하여 올바른 placeholder 문법을 확인하는 것이 중요합니다.

Placeholder를 사용하는 주된 이유는 SQL 인젝션 공격을 방지하고, 데이터베이스 쿼리의 가독성과 유지보수성을 높이며, 쿼리 실행 성능을 최적화하기 위함입니다. 다양한 유형의 placeholder를 적절히 활용함으로써, 보다 안전하고 효율적인 데이터베이스 애플리케이션을 개발할 수 있습니다.

## 예시

SQL 쿼리에서 `?`를 placeholder로 사용하는 예시입니다:

```sql
SELECT * FROM users WHERE username = ? AND password = ?
```

이 쿼리는 `username`과 `password` 필드에 대한 값을 나중에 지정할 수 있게 합니다. 프로그래밍 언어에서 이 쿼리를 사용할 때는 다음과 같이 값을 바인딩합니다:

```go
// Go의 database/sql 패키지 사용 예
stmt, err := db.Prepare("SELECT * FROM users WHERE username = ? AND password = ?")
if err != nil {
    // 오류 처리
}
defer stmt.Close()

rows, err := stmt.Query("exampleUser", "examplePass")
if err != nil {
    // 오류 처리
}
defer rows.Close()
```

이처럼 placeholder를 사용하는 것은 데이터베이스 프로그래밍에서 중요한 보안 관행이며, 코드의 유지보수성과 성능 최적화에 기여합니다.
