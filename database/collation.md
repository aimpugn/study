# Collation

- [Collation](#collation)
    - [`collation`이란?](#collation이란)
    - [Collation이 필요한 이유](#collation이-필요한-이유)
    - [Collation 이름 패턴](#collation-이름-패턴)
    - [collation 설정 방법](#collation-설정-방법)
        - [서버 수준](#서버-수준)
        - [데이터베이스 수준](#데이터베이스-수준)
        - [테이블 및 컬럼 수준](#테이블-및-컬럼-수준)
        - [쿼리 수준에서의 콜레이션](#쿼리-수준에서의-콜레이션)
    - [Collation 예시](#collation-예시)
        - [INSERT](#insert)
        - [WHERE 절](#where-절)
        - [LIKE 검색](#like-검색)
        - [ORDER BY 절](#order-by-절)
        - [인덱스(B-Tree) 탐색 기준](#인덱스b-tree-탐색-기준)
    - [raw 포맷 및 collation 확인하기](#raw-포맷-및-collation-확인하기)
    - [MySQL에서 Collation](#mysql에서-collation)
    - [Citations](#citations)

## `collation`이란?

Collation이라는 용어는 "모으다" 또는 "정리하다"라는 의미의 라틴어 "collatio"에서 유래되었습니다.

> late 14c., "act of bringing together and comparing,"

'책의 순서 정리'나 '데이터베이스에서의 정렬' 등 다양한 맥락에서 사용됩니다.

문자 데이터는 대소문자 구분, 악센트 표시('é', 'è', 'ê' 등 악센트가 있는 문자와 없는 문자를 구분할지 여부를 결정), 공백 처리, 문자 너비 등 특정 규칙에 따라 저장되고 비교될 필요가 있습니다.

데이터베이스에서 Collation은 *데이터가 어떻게 저장, 정렬, 비교되는지를 결정하는 규칙 집합*을 의미하며, 특정 문자셋에 속합니다.
- 문자셋: 문자를 어떻게 바이트로 인코딩할지 결정
- collation: 문자셋에 따라 '인코딩된 문자를 어떻게 비교할지' 결정

## Collation이 필요한 이유

문자열을 저장하거나 검색할 때 "같은 문자인지"를 판단하는 것은 복잡한 문제입니다.
특히 비교할 때 '같다/다르다'를 판단해야 하는데, 단순히 '문자가 일치하느냐'를 판단하는 것보다 더 복잡한 문제입니다.
언어나 지역에 따라 비슷하지만 다르게 인식되는 글자가 존재하기도 하고, 같은 문자라도 인코딩 방식이나 Collation 규칙에 따라 같거나 다를 수 있기 때문입니다.

예를 들어, 컴퓨터 입장에서 악센트가 있는 `'é'`와 그냥 `'e'`는 명백히 다릅니다.
하지만 현실적으로 악센트를 중요시 여기는 지역에서는 서로 다르게 처리하는 게 중요할 수 있고, 그렇지 않다면 같게 처리해야 할 수 있습니다.

```sql
-- 'José' vs 'Jose' 비교 예시
SELECT * FROM users WHERE name = 'Jose';

-- utf8_general_ci: José ≈ Jose → 매칭됨
-- utf8_bin: José ≠ Jose → 매칭 안 됨
```

그리고 특이하게도 독일어의 [`'ß'`](https://www.compart.com/en/unicode/U+00DF)([에스체트](https://namu.wiki/w/%C3%9F))의 경우 '원칙적으로 소문자이며 ss 또는 sz로 나타냅니다. 대문자로 바꿀 때나 에스체트를 쓸 수 없는 경우 SS로 풀어쓰는 게 원칙'이라고 합니다.
그렇다면 독일어에 대한 검색이나 비교를 지원해야 하는 경우 `ß = ss`로 처리되어야 하고, `'ß'`를 대문자로 만들면 `SS`가 되어야 합니다.

그리고, ASCII 기준으로 본다면 `'Z' = 90`, `'a' = 97`, `'Zebra' < 'apple'` 등으로 비교가 이뤄질 수 있습니다.
하지만 누구도 `'Z'`와 90을 같다고 하지 않듯이, 이는 일반적인 언어의 사용례와 거리가 있습니다.

따라서 이 규칙들은 다음과 같은 경우에 사용됩니다.
- 저장할 때
- `WHERE name = 'José'` 같이 문자열을 비교할 때
- `ORDER BY`로 정렬할 때
- `UNIQUE`, `GROUP BY`, `DISTINCT` 등 동등성 판단할 때
- 문자열 인덱스 탐색 성능 결정

데이터의 일관성을 유지, 정렬 정확성, 검색 정밀도, 인덱스 성능, 다국어 호환성, 성능 최적화 등에 영향을 줄 수 있습니다.

## Collation 이름 패턴

```plaintext
{문자셋}_{언어의 지역(locale)을 기준으로 한 비교 규칙}_{비교 민감도}
```

- `{문자셋}`
    - latin1
    - utf8
    - utf8mb4
    - usc2
    - ascii 등
- `{언어의 지역(locale)을 기준으로 한 비교 알고리즘}`
    - general: 일반적인, 단순하고 빠른 규칙 기반 비교
    - unicode: [Unicode 표준 DUCET](https://www.unicode.org/reports/tr10/)에 기반한 비교
    - 0900: Unicode 9.0 이상을 기반으로 한 정렬
    - turkish, german2, swedish, ja 등 특정 언어 맞춤

        e.g. `utf8mb4_ja_0900_as_cs`는 일본어(ja)를 위한 Unicode 9.0 기반 정렬/비교

- `{비교 민감도}`
    - ci: Case Insensitive
    - cs: Case Sensitive
    - ai: Accent Insensitive
    - as: Accent Sensitive
    - ws: Width sensitive

        전각(full width) 및 반각(half width) 문자가 존재하는지 구별합니다.
        `_ws`가 없는 경우 너비에 민감하지 않으므로, 전각 및 반각 문자는 동일한 것으로 간주됩니다.

    - bin: Binary

예를 들어, `utf8_general_ci`는 다음과 같은 의미를 갖습니다.
- utf8: utf-8 문자셋
- general: 단순하고 빠른 규칙 기반 비교
- ci: 대소문자 구분 무시

`utf8mb4_0900_ai_ci` 경우 다음과 같습니다.
- utf8mb4: 'U+10FFFF'까지 utf-8 전체 범위 지원하는 문자셋

    > utf8 vs utf8mb4
    >
    > [utf8](https://dev.mysql.com/doc/refman/8.4/en/charset-unicode-utf8.html)은 [utf8mb3](https://dev.mysql.com/doc/refman/8.4/en/charset-unicode-utf8mb3.html)과 같습니다.
    > 즉 [*3바이트 유니코드*](https://en.wikipedia.org/wiki/UTF-8#Description)까지 지원하기 때문에 U+010000~U+10FFFF 범위의 유니코드를 저장할 수 없습니다.
    >
    > ```sql
    > -- '😄'은 U+1F604(0xF0 0x9F 0x98 0x84)이므로, 3바이트 경우 저장할 수 없습니다.
    > -- ERROR 1366 (HY000): Incorrect string value
    > INSERT INTO messages (text) VALUES ('😄');
    > ```
    >
    > 반면 utf8mb4은 4바이트까지 지원하기 때문에 U+010000~U+10FFFF 범위의 유니코드까지 저장할 수 있습니다.

- 0900: Unicode 9.0
- ai: 악센트 무시
- ci: 대소문자 구분 무시

`utf8_bin`은 이진 비교를 수행합니다.

다음과 같이 Collation을 읽을 수 있습니다.

- `utf8mb4_general_ci`
    - `utf8mb4`: 문자셋
    - `general`: 언어의 지역(locale)을 기준으로 한 비교 알고리즘
    - `ci`: 대소문자를 구분하지 않는다(case-insensitive)는 것

- `utf8_general_ci`
    - `utf8`: 문자셋
    - `general`: 언어의 지역(locale)을 기준으로 한 비교 알고리즘
    - `ci`: 대소문자를 구분하지 않는다(case-insensitive)는 것

- `latin1_general_cs`
    - `latin1`: 문자셋
    - `general`: 언어의 지역(locale)을 기준으로 한 비교 알고리즘
    - `cs`: 대소문자를 구별한다(case-sensitive)는 것

- `latin1_swedish_ci`
    - `latin1`: 문자셋
    - `swedish`: 언어의 지역(locale)을 기준으로 한 비교 알고리즘
    - `cs`: 대소문자를 구별한다(case-sensitive)는 것

## collation 설정 방법

collation은 데이터베이스 서버 수준, 데이터베이스 수준, 테이블 수준, 그리고 컬럼 수준에서 설정할 수 있습니다.

### 서버 수준

SQL 서버 설치 시 collation을 설정할 수 있습니다.
이는 기본적으로 모든 데이터베이스와 테이블에 상속됩니다.

### 데이터베이스 수준

데이터베이스를 생성할 때 collation을 설정할 수 있습니다.
이는 해당 데이터베이스 내의 모든 테이블과 컬럼에 기본적으로 적용됩니다.

```sql
CREATE DATABASE ExampleDatabase COLLATE utf8_unicode_ci;
```

### 테이블 및 컬럼 수준

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

### 쿼리 수준에서의 콜레이션

쿼리에서 collation을 명시적으로 지정하여 비교할 수 있습니다.

```sql
SELECT *
FROM Example
WHERE ExampleName COLLATE utf8_unicode_ci = 'example';
```

```sql
col_name COLLATE latin1_general_cs LIKE 'a%'
col_name LIKE 'a%' COLLATE latin1_general_cs
col_name COLLATE latin1_bin LIKE 'a%'
col_name LIKE 'a%' COLLATE latin1_bin
```

## Collation 예시

### INSERT

예를 들어, 이메일의 경우 [RFC 5321](https://www.rfc-editor.org/rfc/rfc5321#section-4.1.1)에 따르면 'local-part@domain'에서 로컬 부분은 수신하는 사이트의 규칙과 [Section 4.1.2](https://www.rfc-editor.org/rfc/rfc5321#section-4.1.2)의 규칙을 따라야 합니다.

> The syntax of the local part of a mailbox MUST conform to receiver site conventions and the syntax specified in [Section 4.1.2](https://www.rfc-editor.org/rfc/rfc5321#section-4.1.2).

하지만 [Gmail, Outlook, Yahoo 등 여러 메일 서비스에서는 대소문자를 구별하지 않는다](https://mailmeteor.com/glossary/email-case-sensitive)고 해서, 다음과 같은 이메일 주소들을 같다고 판단합니다.

```plaintext
mary.pesquet@example.org
Mary.Pesquet@Example.Org
MARY.PESQUET@EXAMPLE.ORG
MaRy.PeSqUeT@eXaMpLe.OrG
```

하지만 `utf8_bin`를 사용할 경우, `'User@example.com'`와 `'user@example.com'`를 다른 것으로 보게되어 둘 다 등록될 수 있습니다.

```sql
CREATE TABLE users (
  email VARCHAR(100) UNIQUE
) COLLATE utf8_bin;

INSERT INTO users VALUES ('User@example.com');
INSERT INTO users VALUES ('user@example.com');
```

### WHERE 절

```sql
SELECT * FROM users WHERE name = 'café';
```

- `utf8_general_ci` 경우

    `'CAFE'`, `'Café'`, `'cafe'` 모두 매칭됩니다.

- `utf8_bin` 경우

    `name = 'café'`만 매칭됩니다.

내부적으로는 각 문자를 *정규화된 [비교 코드(weight string)](https://mariadb.com/kb/en/weight_string/)*로 변환 후 비교합니다.

### LIKE 검색

저장값이 `'café'`인 경우, 검색 기능이 언어 정규화를 반영하지 않으면 아래와 같은 쿼리를 실행했을 때 결과가 달라질 수 있습니다.

```sql
SELECT * FROM products WHERE name LIKE '%cafe%';
```

- `utf8_general_ci`: 검색 성공 (악센트 무시)
- `utf8_bin`: 검색 실패

### ORDER BY 절

```sql
SELECT name FROM products ORDER BY name;
```

- `utf8_unicode_ci`: 언어학적 기준에 맞게 정렬합니다.
- `utf8_bin`: 바이트값 순서(대문자 먼저)로 정렬합니다.

### 인덱스(B-Tree) 탐색 기준

인덱스는 문자열의 정렬 기준에 따라 노드가 배치되므로, collation이 다르면 인덱스 구성 자체가 달라집니다.
따라서 사용례에 맞지 않는 Collation을 사용하면 풀 스캔이 발생할 수 있습니다.

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
    AND TABLE_NAME IN ('this_values', 'other_values');

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

## MySQL에서 Collation

- [Case Sensitivity in String Searches](https://dev.mysql.com/doc/refman/5.7/en/case-sensitivity.html)

    nonbinary strings(`CHAR`, `VARCHAR`, `TEXT`) 경우 문자열 탐색은 비교 피연산자(comparison operands)의 콜레이션을 사용합니다.

    binary strings(`BINARY`, `VARBINARY`, `BLOB`) 경우 피연산자 바이트의 숫자 값을 비교에 사용합니다. 이는 알파벳 문자열에 대해 대소문자를 구별하여 비교함을 의미합니다.

    nonbinary strings(`CHAR`, `VARCHAR`, `TEXT`)과 binary strings(`BINARY`, `VARBINARY`, `BLOB`)의 비교는 binary strings 비교로 다뤄집니다.

    간단한 비교 연산(`>=`, `>`, `=`, `<`, `<=`, `sorting`, and `grouping`)은 각 문자열의 정렬 값(sort value)에 기반합니다. 같은 정렬 값의 문자들은 같은 문자로 취급됩니다.

    예를 들어 'e'와 'é'가 주어진 콜레이션에서 같은 정렬 값을 갖는다면, 같다고 비교됩니다.

    기본 문자셋과 콜레이션은 `latin1`과 `latin1_swedish_ci`이고, 따라서 nonbinary string 비교는 기본적으로 대소문자를 구별하지 않습니다.

    이 경우 만약 아래와 같이 검색한다면, 'A' 또는 'a'로 시작하는 모든 칼럼의 값들을 조회하게 됩니다.

    ```sql
    WHERE col_name LIKE 'a%'
    ```

    '대소문자를 구분'하여 검색하고 싶다면, 피연산자중 하나가 '대소문자 구분'하거나 이진인 콜레이션이어야 합니다.
    예를 들어, 둘 다 `latin1` 문자셋인 칼럼과 문자열을 비교한다면, `COLLATE` 연산자를 사용하여 피연산자가 `latin1_general_cs` 또는 `latin1_bin` 콜레이션이 되도록 할 수 있습니다.

    ```sql
    col_name COLLATE latin1_general_cs LIKE 'a%'
    col_name LIKE 'a%' COLLATE latin1_general_cs
    col_name COLLATE latin1_bin LIKE 'a%'
    col_name LIKE 'a%' COLLATE latin1_bin
    ```

- [Cast Function and Operator Descriptions](https://dev.mysql.com/doc/refman/5.7/en/cast-functions.html#operator_binary)

    ```sql
    mysql> SELECT 'a' = 'A';
            -> 1
    mysql> SELECT BINARY 'a' = 'A';
            -> 0
    mysql> SELECT 'a' = 'a ';
            -> 1
    mysql> SELECT BINARY 'a' = 'a ';
            -> 0
    ```

## Citations

- [Collation](https://en.wikipedia.org/wiki/Collation)
- [Collations](https://manifold.net/doc/mfd9/index.htm#collations.htm)
- [Collations and Case Sensitivity](https://learn.microsoft.com/en-us/ef/core/miscellaneous/collations-and-case-sensitivity)
- [What Is Database Collation?](https://fivetran.com/docs/connectors/databases/troubleshooting/database-collation)
- [What is Collation in Database?](https://prerna7692.medium.com/what-is-collation-in-database-720dd92f6a57)
- [What does "collation" mean?](https://dba.stackexchange.com/a/196541)
- [Understanding Database Character Sets and Collations](https://blog.fourninecloud.com/database-character-set-charset-collation-and-their-relationship-explained-227bd5155c48)
- [Character sets and collations in MySQL](https://planetscale.com/blog/mysql-charsets-collations)
- [Character Sets vs. Collations in a MySQL Database Infrastructure](https://www.dbvis.com/thetable/character-sets-vs-collations-in-a-mysql-database-infrastructure/)
- [Unicode Collation Algorithm](https://www.unicode.org/reports/tr10/#Introduction)
- ICU
    - [ICU-TC Home Page](https://icu.unicode.org/home)
    - [ICU Collation Demo](https://icu4c-demos.unicode.org/icu-bin/collation.html)
    - [ICU - github](https://github.com/unicode-org/icu)
