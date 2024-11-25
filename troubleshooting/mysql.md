# mysql

- [mysql](#mysql)
    - [SQL Error \[1267\] \[HY000\]: Illegal mix of collations (utf8\_unicode\_ci,IMPLICIT) and (utf8\_general\_ci,IMPLICIT) for operation '='](#sql-error-1267-hy000-illegal-mix-of-collations-utf8_unicode_ciimplicit-and-utf8_general_ciimplicit-for-operation-)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction](#serialization-failure-1213-deadlock-found-when-trying-to-get-lock-try-restarting-transaction)
        - [원인](#원인-1)
    - [Can not read response from server. Expected to read 4 bytes, read 0 bytes before connection was unexpectedly lost](#can-not-read-response-from-server-expected-to-read-4-bytes-read-0-bytes-before-connection-was-unexpectedly-lost)
        - [문제](#문제-1)
    - [SQLSTATE\[HY000\]: General error: 1267 Illegal mix of collations (utf8\_general\_ci,IMPLICIT) and (utf8\_unicode\_ci,IMPLICIT) for operation '='](#sqlstatehy000-general-error-1267-illegal-mix-of-collations-utf8_general_ciimplicit-and-utf8_unicode_ciimplicit-for-operation-)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-1)
        - [예시](#예시)
            - [테이블 및 컬럼의 collation 변경](#테이블-및-컬럼의-collation-변경)
            - [쿼리에서 collation 명시](#쿼리에서-collation-명시)
            - [데이터베이스 전체의 collation 변경](#데이터베이스-전체의-collation-변경)
        - [결론](#결론)

## SQL Error [1267] [HY000]: Illegal mix of collations (utf8_unicode_ci,IMPLICIT) and (utf8_general_ci,IMPLICIT) for operation '='

### 문제

```sql
CREATE PROCEDURE insert_demo_payments_by_rody(IN dup_merchant_uid VARCHAR(40), IN loop_cnt INT)
BEGIN
    DECLARE payments_cnt INT;
    DECLARE i INT DEFAULT 0;
    SELECT COUNT(1) INTO payments_cnt FROM payments p WHERE p.merchant_uid = dup_merchant_uid;
                                                            ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
```

표시한 부분에서 에러가 발생하는 것으로 보임

### 원인

collation 확인

```sql
SELECT 
    TABLE_SCHEMA, 
    TABLE_NAME, 
    COLUMN_NAME, 
    COLLATION_NAME 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = 'service_name' 
AND TABLE_NAME = 'payments';
```

`merchant_uid`의 collation은 `utf8_unicode_ci`

### 해결

## Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction

```json
{
  "_index": "log-v1-2023.10.11-000022",
  "_id": "HQNLH4sBqFTWdmBOIdBT",
  "_version": 1,
  "_score": null,
  "_source": {
    "app": {
      "_raw": "{\"@timestamp\":\"2023-10-12T00:11:38+09:00\",\"app\":\"core-pay\",\"context\":{\"parent_span_id\":null,\"request_id\":\"Root=1-6526bb29-3385cad6748db8cf7fe462f6\",\"span_id\":\"16a785dec697b60398c64ed5e4cc4d70\",\"trace_id\":\"4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4\"},\"etc\":{\"cpu\":\"0.2(1m)\",\"memory\":\"6(MB)\",\"pid\":17510,\"processing_time\":\"0.22138214111328(s)\",\"stack_trace\":[{\"class\":\"PortLogger\",\"file\":\"/var/www/service/app/Controller/AppController.php\",\"function\":\"write\",\"line\":71,\"type\":\"::\"},{\"class\":\"AppController\",\"file\":\"/var/www/service/app/Controller/CertificatesController.php\",\"function\":\"log\",\"line\":339,\"type\":\"->\"},{\"class\":\"CertificatesController\",\"file\":\"/var/www/service/app/Controller/CertificatesController.php\",\"function\":\"generateUID\",\"line\":294,\"type\":\"->\"}]},\"file\":\"/var/www/service/app/tmp/logs/application.log\",\"host\":\"ip-10-0-3-158\",\"level\":\"error\",\"logger_name\":\"core-pay\",\"msg\":\"SQLSTATE[40001]: Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction\"}",
      "code_location": "[{\"class\":\"PortLogger\",\"file\":\"/var/www/service/app/Controller/AppController.php\",\"function\":\"write\",\"line\":71,\"type\":\"::\"},{\"class\":\"AppController\",\"file\":\"/var/www/service/app/Controller/CertificatesController.php\",\"function\":\"log\",\"line\":339,\"type\":\"->\"},{\"class\":\"CertificatesController\",\"file\":\"/var/www/service/app/Controller/CertificatesController.php\",\"function\":\"generateUID\",\"line\":294,\"type\":\"->\"}]",
      "context": {
        "_raw": "{\"parent_span_id\":null,\"request_id\":\"Root=1-6526bb29-3385cad6748db8cf7fe462f6\",\"span_id\":\"16a785dec697b60398c64ed5e4cc4d70\",\"trace_id\":\"4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4\"}",
        "b3": {
          "_raw": "{\"parent_span_id\":null,\"request_id\":\"Root=1-6526bb29-3385cad6748db8cf7fe462f6\",\"span_id\":\"16a785dec697b60398c64ed5e4cc4d70\",\"trace_id\":\"4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4\"}",
          "request_id": "Root=1-6526bb29-3385cad6748db8cf7fe462f6",
          "span_id": "16a785dec697b60398c64ed5e4cc4d70",
          "trace_id": "4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4"
        }
      },
      "level": "error",
      "message": "SQLSTATE[40001]: Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction",
      "name": "core-pay",
      "timestamp": "2023-10-11T15:11:38.582578723Z"
    },
    "b3": {
      "_raw": "{\"parent_span_id\":null,\"request_id\":\"Root=1-6526bb29-3385cad6748db8cf7fe462f6\",\"span_id\":\"16a785dec697b60398c64ed5e4cc4d70\",\"trace_id\":\"4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4\"}",
      "request_id": "Root=1-6526bb29-3385cad6748db8cf7fe462f6",
      "span_id": "16a785dec697b60398c64ed5e4cc4d70",
      "trace_id": "4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4"
    },
    "file": "/var/www/service/app/tmp/logs/application.log",
    "hostname": "ip-10-0-3-158",
    "level": "ERROR",
    "timestamp": "2023-10-11T15:11:38.582549462Z"
  },
  "fields": {
    "timestamp": [
      "2023-10-11T15:11:38.582Z"
    ]
  },
  "highlight": {
    "app._raw": [
      "{\"@timestamp\":\"2023-10-12T00:11:38+09:00\",\"app\":\"core-pay\",\"context\":{\"parent_span_id\":null,\"request_id\":\"Root=1-6526bb29-3385cad6748db8cf7fe462f6\",\"span_id\":\"16a785dec697b60398c64ed5e4cc4d70\",\"trace_id\":\"4e95c9a05a9fbc2fc340d8480e882359692edfcf39655063abae2b73f0c73ec4\"},\"etc\":{\"cpu\":\"0.2(1m)\",\"memory\":\"6(MB)\",\"pid\":17510,\"processing_time\":\"0.22138214111328(s)\",\"stack_trace\":[{\"class\":\"PortLogger\",\"file\":\"/var/www/service/app/Controller/AppController.php\",\"function\":\"write\",\"line\":71,\"type\":\"::\"},{\"class\":\"AppController\",\"file\":\"/var/www/service/app/Controller/CertificatesController.php\",\"function\":\"log\",\"line\":339,\"type\":\"->\"},{\"class\":\"CertificatesController\",\"file\":\"/var/www/service/app/Controller/CertificatesController.php\",\"function\":\"generateUID\",\"line\":294,\"type\":\"->\"}]},\"file\":\"/var/www/service/app/tmp/logs/application.log\",\"host\":\"ip-10-0-3-158\",\"level\":\"error\",\"logger_name\":\"core-pay\",\"msg\":\"SQLSTATE[40001]: Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction\"}"
    ]
  },
  "sort": [
    1697037098582
  ]
}
```

### 원인

- [[Messenger] SQLSTATE[40001]: Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction #43339](https://github.com/symfony/symfony/issues/43339)

```sql
SHOW ENGINE INNODB STATUS;
```

- [Serialization failure: 1213 Deadlock found when trying to get lock; try restarting transaction](https://laracasts.com/discuss/channels/laravel/serialization-failure-1213-deadlock-found-when-trying-to-get-lock-try-restarting-transaction)

```log
Deadlock means that another process is locking the record you are trying to update in the database.

I take it you are using database transactions.

They work something like this

Select the record(s) that you need to update and lock them
Update the record(s)
Commit or rollback the changes
Release the lock
While the record is locked another process or user can't update it until the process or user locking it does a rollback or a commit.

There are a few ways to solve this but it requires a bit more insight into your application.
```

## Can not read response from server. Expected to read 4 bytes, read 0 bytes before connection was unexpectedly lost

### 문제

```log
Can not read response from server. Expected to read 4 bytes, read 0 bytes before connection was unexpectedly lost.
```

## SQLSTATE[HY000]: General error: 1267 Illegal mix of collations (utf8_general_ci,IMPLICIT) and (utf8_unicode_ci,IMPLICIT) for operation '='

### 문제

```bash
Error: SQLSTATE[HY000]: General error: 1267 Illegal mix of collations (utf8_general_ci,IMPLICIT) and (utf8_unicode_ci,IMPLICIT) for operation '='
#0 /var/www/pay/lib/Cake/Model/Datasource/DboSource.php(459): PDOStatement->execute(Array)
#1 /var/www/pay/lib/Cake/Model/Datasource/DboSource.php(424): DboSource->_execute('SELECT `Billing...', Array)
#2 /var/www/pay/lib/Cake/Model/Datasource/DboSource.php(667): DboSource->execute('SELECT `Billing...', Array, Array)
#3 /var/www/pay/lib/Cake/Model/Datasource/DboSource.php(1107): DboSource->fetchAll('SELECT `Billing...', false)
... 생략 ...
```

### 원인

`SQLSTATE[HY000]: General error: 1267 Illegal mix of collations (utf8_general_ci,IMPLICIT) and (utf8_unicode_ci,IMPLICIT) for operation '='` 에러는 MySQL에서 서로 다른 collation을 사용하는 두 문자열을 비교할 때 발생하는 에러입니다. 이 에러는 주로 다음과 같은 상황에서 발생합니다:

1. **서로 다른 collation을 사용하는 두 컬럼을 비교할 때**:
   - 예를 들어, 하나의 컬럼이 `utf8_general_ci` collation을 사용하고 다른 컬럼이 `utf8_unicode_ci` collation을 사용하는 경우.

2. **쿼리에서 문자열 리터럴과 컬럼을 비교할 때**:
   - 문자열 리터럴이 특정 collation을 가지고 있고, 컬럼이 다른 collation을 가지고 있는 경우.

### 해결

이 문제를 해결하기 위해서는 collation을 일치시키는 것이 중요합니다. 다음은 몇 가지 해결 방법입니다:

1. **테이블 및 컬럼의 collation을 일치시키기**:
   - 테이블과 컬럼의 collation을 동일하게 설정합니다.
   - 예를 들어, 모든 테이블과 컬럼을 `utf8_unicode_ci`로 설정할 수 있습니다.

   ```sql
   ALTER TABLE table_name CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
   ALTER TABLE table_name MODIFY column_name VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci;
   ```

2. **쿼리에서 collation을 명시적으로 지정하기**:
   - 쿼리에서 비교할 때 collation을 명시적으로 지정하여 일치시킵니다.

   ```sql
   SELECT * FROM table_name WHERE column_name COLLATE utf8_unicode_ci = 'value' COLLATE utf8_unicode_ci;
   ```

3. **데이터베이스 전체의 collation을 변경하기**:
   - 데이터베이스 전체의 collation을 변경하여 일관성을 유지합니다.

   ```sql
   ALTER DATABASE database_name CHARACTER SET utf8 COLLATE utf8_unicode_ci;
   ```

### 예시

#### 테이블 및 컬럼의 collation 변경

```sql
-- 테이블의 모든 컬럼을 utf8_unicode_ci로 변경
ALTER TABLE users CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE products CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
ALTER TABLE productUsers CONVERT TO CHARACTER SET utf8 COLLATE utf8_unicode_ci;
```

#### 쿼리에서 collation 명시

```sql
-- 쿼리에서 collation을 명시적으로 지정
SELECT * FROM users WHERE username COLLATE utf8_unicode_ci = 'john_doe' COLLATE utf8_unicode_ci;
```

#### 데이터베이스 전체의 collation 변경

```sql
-- 데이터베이스의 collation을 utf8_unicode_ci로 변경
ALTER DATABASE my_database CHARACTER SET utf8 COLLATE utf8_unicode_ci;
```

### 결론

`Illegal mix of collations` 에러는 서로 다른 collation을 사용하는 문자열을 비교할 때 발생합니다. 이를 해결하기 위해서는 테이블과 컬럼의 collation을 일치시키거나, 쿼리에서 collation을 명시적으로 지정하는 방법을 사용할 수 있습니다. 데이터베이스 전체의 collation을 일관되게 설정하는 것도 좋은 방법입니다.
