# Collation

- [Collation](#collation)
    - [`collation`이란?](#collation이란)
    - [Collation의 중요성](#collation의-중요성)
    - [collation 설정 방법](#collation-설정-방법)
        - [**서버 수준**](#서버-수준)
        - [**데이터베이스 수준**](#데이터베이스-수준)
        - [**테이블 및 컬럼 수준**](#테이블-및-컬럼-수준)
        - [**쿼리 수준**](#쿼리-수준)
    - [Collation의 명명 규칙](#collation의-명명-규칙)
    - [raw 포맷 및 collation 확인하기](#raw-포맷-및-collation-확인하기)
    - [Citations](#citations)

## `collation`이란?

Collation이라는 용어는 "모으다" 또는 "정리하다"라는 의미의 라틴어 "collatio"에서 유래되었습니다.
이 용어는 600년 이상 사용되어 왔으며, 책의 순서 정리나 데이터베이스에서의 정렬 등 다양한 맥락에서 사용됩니다.

데이터베이스에서 Collation은 *데이터가 어떻게 정렬되고 비교되는지를 결정하는 규칙 집합*을 의미합니다.
Collation은 특정 문자셋에 속합니다.
문자셋이 문자를 어떻게 바이트로 인코딩할지 결정한다면, collation은 이러한 인코딩된 문자를 어떻게 비교할지 결정합니다.

문자 데이터는 특정 규칙에 따라 정렬되며, 이 규칙은 다음 같은 항목들을 지정할 수 있습니다
- 대소문자 구분
- 악센트 표시(é, è, ê 등 악센트가 있는 문자와 없는 문자를 구분할지 여부를 결정)
- 공백 처리
- 일본어 카나 문자 유형
- 문자 너비 등

이 규칙들은 다음과 같은 경우들에 사용됩니다.
- 저장
- 비교
- 정렬: 어떤 문자가 다른 문자보다 먼저 오는지를 결정(알파벳 순, 사전식 순서 등)
- 검색

이를 통해 데이터의 일관성을 유지하고, 다국어 지원을 가능하게 하며, 성능을 최적화할 수 있습니다.

예를 들어, `utf8_general_ci`는 대소문자를 구분하지 않는(ci는 case-insensitive를 의미) 비교를 수행하며, `utf8_bin`은 이진 비교를 수행합니다.

## Collation의 중요성

- **정렬 및 비교**:

   데이터베이스에서 문자열을 정렬하거나 비교할 때 collation은 중요한 역할을 합니다.

   예를 들어, `utf8_general_ci`와 `utf8_unicode_ci`는 서로 다른 정렬 및 비교 규칙을 가지고 있습니다.

- **다국어 지원**:

   다양한 언어와 문자 집합을 지원하기 위해 collation은 필수적입니다.

   예를 들어, 터키어에서는 'i'와 'İ'가 다른 문자로 취급되므로, 이를 올바르게 처리하기 위해 적절한 collation이 필요합니다.

- **성능**:

   적절한 collation을 선택하면 데이터베이스의 성능을 최적화할 수 있습니다.
   잘못된 collation은 불필요한 인덱스 스캔을 초래할 수 있습니다.

## collation 설정 방법

collation은 데이터베이스 서버 수준, 데이터베이스 수준, 테이블 수준, 그리고 컬럼 수준에서 설정할 수 있습니다.

### **서버 수준**

SQL 서버 설치 시 collation을 설정할 수 있습니다.
이는 기본적으로 모든 데이터베이스와 테이블에 상속됩니다.

### **데이터베이스 수준**

데이터베이스를 생성할 때 collation을 설정할 수 있습니다.
이는 해당 데이터베이스 내의 모든 테이블과 컬럼에 기본적으로 적용됩니다.

```sql
CREATE DATABASE ExampleDatabase COLLATE utf8_unicode_ci;
```

### **테이블 및 컬럼 수준**

테이블을 생성할 때 특정 컬럼에 대해 collation을 설정할 수 있습니다.

```sql
CREATE TABLE Example (
    ExampleId int,
    ExampleName varchar(255) COLLATE utf8_unicode_ci
);
```

이미 생성된 테이블에 칼럼을 추가할 경우

```sql
-- mysql 5.5
ALTER TABLE some_table
ADD column_to_add VARCHAR(128)
CHARACTER SET utf8 COLLATE utf8_unicode_ci
AFTER another_column;
```

### **쿼리 수준**

쿼리에서 collation을 명시적으로 지정하여 비교할 수 있습니다.

```sql
SELECT * FROM Example WHERE ExampleName COLLATE utf8_unicode_ci = 'example';
```

## Collation의 명명 규칙

MySQL의 collation 이름에는 보통 세 가지 부분이 있다.

- `utf8mb4_general_ci`

    - `utf8mb4`: 문자셋
    - `general`: 비교 알고리즘
    - `ci`: 대소문자를 구분하지 않는다(case-insensitive)는 것

- `utf8_general_ci`

    - `utf8`: 문자셋
    - `general`: 비교 알고리즘
    - `ci`: 대소문자를 구분하지 않는다(case-insensitive)는 것

## raw 포맷 및 collation 확인하기

```sql
-- MySQL
-- 파일 포맷 보기
SHOW VARIABLES LIKE "%innodb_file%";
-- row format 보기
SELECT
    TABLE_NAME,
    TABLE_TYPE,
    ENGINE,
    ROW_FORMAT,
    TABLE_COLLATION
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA='some_table'
    AND TABLE_NAME IN ('some_value_histories', 'other_values');

SHOW TABLE STATUS WHERE NAME LIKE 'table_name';

SHOW FULL COLUMNS FROM table_name;

-- column들의 collation 확인하기
SELECT
    TABLE_NAME,
    COLUMN_NAME,
    COLLATION_NAME
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'schema_name'
AND TABLE_NAME IN ('table_name');
```

## Citations

[1] <https://fivetran.com/docs/connectors/databases/troubleshooting/database-collation#:~:text=Collation%20refers%20to%20a%20set,character%20types%2C%20and%20character%20width>.
[2] <https://prerna7692.medium.com/what-is-collation-in-database-720dd92f6a57>
[3] <https://en.wikipedia.org/wiki/Collation>
[4] <https://manifold.net/doc/mfd9/collations.htm>
[5] <https://dba.stackexchange.com/questions/4270/what-does-collation-mean>
[6] <https://blog.fourninecloud.com/database-character-set-charset-collation-and-their-relationship-explained-227bd5155c48>
[7] <https://learn.microsoft.com/en-us/ef/core/miscellaneous/collations-and-case-sensitivity>
[8] <https://www.dbvis.com/thetable/character-sets-vs-collations-in-a-mysql-database-infrastructure/>
[9] <https://planetscale.com/blog/mysql-charsets-collations>
[10] <https://docs.intersystems.com/irislatest/csp/docbook/DocBook.UI.Page.cls?KEY=GSQL_collation>
