# Java SQL

- [Java SQL](#java-sql)
    - [`Statement`와 `PreparedStatement`](#statement와-preparedstatement)
        - [`Statement`](#statement)
        - [`PreparedStatement`](#preparedstatement)
        - [Java 애플리케이션 관점](#java-애플리케이션-관점)
        - [MySQL 서버 관점](#mysql-서버-관점)
    - [기타](#기타)

## `Statement`와 `PreparedStatement`

### `Statement`

일반적으로 단순하고 고정적인 SQL을 실행할 때 사용합니다.
'실행 시점마다 매번 SQL 쿼리 문자열을 직접 DB 서버에 전달'하여 처리합니다.

```java
try (var conn = dataSource.getConnection()) {
    try(var stmt = conn.createStatement()) {
        stmt.execute("CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100));");
        stmt.execute("DELETE FROM users");
        stmt.execute("INSERT INTO users(name) VALUES ('hello'), ('world'), ('Alice'), ('Bob');");
    }

    try (var stmt = conn.createStatement()) {
        stmt.executeQuery("SELECT * FROM users WHERE id = 1");
    }
}
```

실행마다 쿼리를 문자열로 생성하기 때문에 문자열 연산, 변수 처리 등으로 인한 오버헤드가 발생합니다.

MySQL 로그를 보면 다음과 같이 `Query`로 남습니다.

```sql
255 Query  SELECT 42
255 Query  SELECT 43
```

[`COM_QUERY`](https://dev.mysql.com/doc/dev/mysql-server/8.4.3/page_protocol_com_query.html) 명령을 서버로 보냅니다.

### `PreparedStatement`

SQL을 미리 '컴파일'하여 재사용할 수 있도록 하는 인터페이스입니다.
파라미터를 바인딩하여 실행 시 매번 쿼리 문자열을 전달하지 않고, 변수의 값만 전달합니다.

```java
try (var pstmt = conn.prepareStatement(
    "SELECT * FROM users WHERE id = ?"
)) {
    pstmt.setInt(1, 1);
    pstmt.executeQuery();
}
```

최초 한 번만 문자열을 생성하고, 이후 실행 시엔 파라미터만 바인딩하여 재사용합니다.
문자열 처리나 SQL Injection (SQL 주입 공격)을 방지하는 데 효과적이고, 문자열 처리의 오버헤드가 덜합니다.

MySQL 로그를 보면 다음과 같이 `Prepare`, `Execute` 등으로 남습니다.

```sql
254 Prepare    SELECT ?
254 Execute    SELECT 42
254 Execute    SELECT 43
```

실행 시점에는 실제 SQL 문자열 대신 Prepared Statement 핸들과 placeholder 값만 전송됩니다.

### Java 애플리케이션 관점

```properties
useServerPrepStmts=false
```

이 설정을 사용하면 클라이언트 측의 Prepared Statement를 사용합니다.
즉 Java 애플리케이션에서 Prepared Statement를 에뮬레이션(emulation, 모방)하는 방식입니다.

실제로는 쿼리를 문자열로 구성해 매번 전송합니다.
클라이언트에서 드라이버가 SQL 문자열을 미리 분석(tokenize)하고, 모든 placeholder(`?`)를 실행 전에 실제 값으로 치환한 뒤 완성된 SQL을 서버로 보내 실행합니다.

이 방식은 매 실행마다 완성된 SQL 쿼리 전체를 DB 서버로 전달합니다.

### MySQL 서버 관점

`Statement` 사용할 경우 쿼리가 동일해도, 쿼리 요청마다 다음 과정을 반복합니다:
1. SQL 문자열 파싱
2. 최적화(쿼리 최적화기 실행)
3. 실행 계획 생성
4. 실행

'쿼리 파싱'과 '실행 계획을 생성'하는 작업이 비교적 무거운 연산이기 때문에, `PreparedStatement`를 사용할 경우 성능 향상이 가능합니다:
- 최초 실행 시:
    1. SQL 문자열 파싱
    2. 최적화(쿼리 최적화기 실행)
    3. 실행 계획 생성 및 캐시

- 이후 실행 시:
    1. 캐시된 실행 계획 재사용
    2. 변수 바인딩
    3. 즉시 실행

최초 한 번만 [`COM_STMT_PREPARE`](https://dev.mysql.com/doc/dev/mysql-server/8.4.3/page_protocol_com_stmt_prepare.html) 명령으로 'SQL 쿼리 문자열'이 서버에 전달됩니다.
이후 실행은 Prepared Statement 핸들(Statement Handle)과 placeholder에 치환할 실제 값만을 포함한 [`COM_STMT_EXECUTE`](https://dev.mysql.com/doc/dev/mysql-server/8.4.3/page_protocol_com_stmt_execute.html) 명령이 전달됩니다.

참고로 `PreparedStatement`를 설명할 때 흔히 "컴파일된다"는 표현을 사용합니다.
이는 데이터베이스 서버에서 SQL 쿼리를 파싱하고, 실행 가능한 '실행 계획'으로 변환하고, 최적화하고, 캐시하는 과정을 의미합니다.
다음 실행 시에는 서버가 실행 계획을 바로 재사용하여 즉시 실행할 수 있습니다.

데이터베이스 서버의 리소스 사용량이 증가할 수 있고, [MySQL 경우 개수 제한](https://dev.mysql.com/doc/refman/8.4/en/server-system-variables.html#sysvar_max_prepared_stmt_count)을 설정할 수 있습니다.

```ini
[mysqld]
max_prepared_stmt_count=16382
```

JDBC 통해 다음과 같은 설정들도 가능합니다.

```properties
jdbc:mysql://host:3306/db
  ?useServerPrepStmts=true
  &cachePrepStmts=true
  &prepStmtCacheSize=250
  &prepStmtCacheSqlLimit=2048
```

- [useServerPrepStmts](https://dev.mysql.com/doc/connector-j/en/connector-j-connp-props-prepared-statements.html#cj-conn-prop_useServerPrepStmts)
- [cachePrepStmts](https://dev.mysql.com/doc/connector-j/en/connector-j-connp-props-performance-extensions.html#cj-conn-prop_cachePrepStmts)

    ['prepared statement'를 캐시하고 같은 SQL 문장에 대해 같은 인스턴스를 사용하도록 하는데, '준비 가능성'(preparability)도 캐시](https://stackoverflow.com/a/32645365)합니다. 어떤 문장은 데이터베이스 서버에서 준비가능하지 않을 수 있기 대문입니다.

    드라이버는 우선 데이터베이스 서버에서 준비를 시도하고, 만약 실패하면 클라이언트의 prepared statement로 돌아갑니다. 이 확인은 서버 라운드 트립(왕복)이 필요하므로 비용이 많이 듭니다.

    또한 이 체크 결과를 캐싱합니다.

## 기타

- [Prepared statement](https://en.wikipedia.org/wiki/Prepared_statement)
- [Using Prepared Statements](https://docs.oracle.com/javase/tutorial/jdbc/basics/prepared.html)
- [Table 6.7 Prepared Statements Properties](https://dev.mysql.com/doc/connector-j/en/connector-j-reference-configuration-properties.html#:~:text=Table%C2%A06.7%C2%A0Prepared,3.0.2)
