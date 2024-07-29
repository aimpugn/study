# update

## `ON DUPLICATE KEY UPDATE`

MySQL에서 중복된 키 값이 삽입되려 할 때 업데이트를 수행하는 기능으로 `INSERT ... ON DUPLICATE KEY UPDATE` 구문을 사용할 수 있습니다.
이 구문은 테이블에 레코드를 삽입하려고 할 때 이미 같은 기본 키를 가진 레코드가 존재하는 경우, 새로 삽입하는 대신 해당 레코드를 업데이트하도록 설계되었습니다.

예를 들어, `users` 테이블에서 `id`가 기본 키라고 할 때, `id` 값이 중복되는 경우 해당 `id`의 다른 컬럼을 업데이트하는 SQL 문을 작성할 수 있습니다:

```sql
INSERT INTO users (id, name, email) VALUES (1, 'John Doe', 'john@example.com')
ON DUPLICATE KEY UPDATE name = VALUES(name), email = VALUES(email);
```

이 예제에서는 `id`가 1인 사용자를 삽입하려고 시도합니다. 만약 `id` 1이 이미 존재한다면, `name`과 `email` 필드를 새로운 값으로 업데이트합니다. `VALUES()` 함수는 `INSERT` 문에서 사용된 값을 참조하는 데 사용됩니다.

이 구문은 데이터의 일관성을 유지하고 데이터베이스 오류를 방지하는 데 유용하며, 특히 데이터를 자주 갱신하거나 중복 입력 가능성이 높은 경우에 적합합니다.

MySQL에서는 `INSERT INTO ... ON DUPLICATE KEY UPDATE` 구문을 사용하여 중복 키가 발견될 때 업데이트를 수행할 수 있습니다. 이 구문은 `WHERE` 절을 사용하여 업데이트 조건을 지정할 수 없지만, 업데이트할 값 자체에 조건을 포함시킬 수 있습니다.

### 기본 구문

`INSERT INTO ... ON DUPLICATE KEY UPDATE` 구문은 다음과 같이 사용됩니다:

```sql
INSERT INTO 테이블명 (컬럼1, 컬럼2, ...)
VALUES (값1, 값2, ...)
ON DUPLICATE KEY UPDATE 컬럼1 = 값1, 컬럼2 = 값2, ...;
```

만약 테이블에 이미 같은 키를 가진 레코드가 존재한다면, `UPDATE` 절이 실행되어 해당 레코드가 새로운 값으로 업데이트됩니다.

### 예제

예를 들어, `users` 테이블에 `id`가 기본 키이고, 사용자의 `email`을 업데이트하려는 경우 다음과 같이 쿼리를 작성할 수 있습니다:

```sql
INSERT INTO users (id, name, email)
VALUES (1, 'John Doe', 'john@example.com')
ON DUPLICATE KEY UPDATE email = VALUES(email);
```

이 쿼리는 `id`가 1인 레코드가 이미 존재한다면, `email`을 '<john@example.com>'으로 업데이트합니다. 만약 `id`가 1인 레코드가 존재하지 않는다면, 새로운 레코드를 삽입합니다.

### 조건적 업데이트

`WHERE` 절을 사용할 수 없으므로, 업데이트할 값에 직접 조건을 포함시켜야 합니다.

예를 들어, 특정 조건에 따라 `email`을 업데이트하려면 다음과 같이 할 수 있습니다:

```sql
INSERT INTO users (id, name, email)
VALUES (1, 'John Doe', 'john@example.com')
ON DUPLICATE KEY UPDATE email = IF(조건, VALUES(email), email);
```

여기서 `조건`은 특정 조건식이며, 이 조건이 참이면 `VALUES(email)`로 업데이트하고, 거짓이면 기존의 `email` 값을 유지합니다.

이 방법을 통해 `INSERT ... ON DUPLICATE KEY UPDATE` 구문을 사용하여 중복 키가 있는 경우에만 선택적으로 업데이트를 수행할 수 있습니다.
