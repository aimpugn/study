# Collation

- [Collation](#collation)
    - [`collation`이란?](#collation이란)
    - [Collation이 필요한 이유](#collation이-필요한-이유)
    - [비교는 어떻게 이뤄지는가](#비교는-어떻게-이뤄지는가)
        - [문자, 코드포인트, 가중치](#문자-코드포인트-가중치)
        - [UCA와 DUCET](#uca와-ducet)
        - [다층 가중치](#다층-가중치)
        - [기호의 가변 가중치](#기호의-가변-가중치)
        - [식별자에서의 함의](#식별자에서의-함의)
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

문자열을 저장하거나 검색할 때 두 문자열이 "같은가"를 판단하는 일은 생각보다 까다롭습니다. 단순히 바이트가 일치하는지 보는 것으로는 부족하기 때문입니다. 언어나 지역에 따라 비슷하지만 다르게 인식되는 글자가 있고, 같은 문자라도 인코딩 방식이나 Collation 규칙에 따라 같다고도 다르다고도 볼 수 있습니다.

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
- `UNIQUE`, `GROUP BY`, `DISTINCT` 등으로 동등성을 판단할 때
- 문자열 인덱스 탐색 성능을 결정할 때

결국 데이터 일관성, 정렬 정확성, 검색 정밀도, 인덱스 성능, 다국어 호환성까지 두루 영향을 받습니다.

## 비교는 어떻게 이뤄지는가

콜레이션이 "두 문자열이 같은가, 어느 쪽이 앞인가"를 정한다고 했는데, 그 판단은 바이트를 직접 견주는 것이 아닙니다. 각 문자에 비교용 숫자(가중치, weight)를 매기고 그 숫자열을 견줍니다. 이 절은 그 가중치가 어디서 오는지(유니코드와 UCA), 어떻게 층을 이루는지, 그리고 기호·구두점이 왜 까다로운지를 가장 작은 예부터 쌓습니다.

### 문자, 코드포인트, 가중치

먼저 세 가지를 구분합니다.

1. 문자(character): 사람이 읽는 글자. 예 `A`, `a`, `é`.
2. 코드포인트(code point): 유니코드가 각 문자에 붙인 번호. 예 `A`=U+0041, `a`=U+0061, `é`=U+00E9. 유니코드는 "문자와 번호"를 잇는 사전입니다.
3. 인코딩(encoding): 코드포인트를 실제 바이트로 적는 방법. UTF-8은 `A`(U+0041)를 `0x41` 한 바이트로, `é`(U+00E9)를 `0xC3 0xA9` 두 바이트로 적습니다.

저장(인코딩, UTF-8)은 문자셋의 일입니다. 콜레이션은 그 위에서 "코드포인트를 어떤 가중치로 비교할까"를 정합니다. 같은 바이트라도 콜레이션이 다르면 비교 결과가 다릅니다.

작은 예로, `'a'`와 `'A'`는 코드포인트가 U+0061, U+0041로 분명히 다릅니다. 그런데 대소문자 무시(ci) 콜레이션에서는 같다고 봅니다. 바이트나 코드포인트가 아니라 가중치로 비교하는데, ci 콜레이션이 둘에 같은 가중치를 주도록 설계돼 있기 때문입니다.

### UCA와 DUCET

그 가중치를 표준으로 정한 것이 UCA(Unicode Collation Algorithm, 유니코드 기술표준 [UTS #10](https://www.unicode.org/reports/tr10/))입니다. UCA는 사람이 사전에서 기대하는 순서로 문자열을 정렬·비교하는 방법을 규정합니다. UCA가 참조하는 기본 가중치 표가 DUCET(Default Unicode Collation Element Table)이고, DUCET은 각 문자에 collation element(가중치 묶음)를 부여합니다.

MySQL에서 콜레이션 이름의 가운데 토막이 이 알고리즘을 가리킵니다.

- `..._unicode_ci`: UCA 기반(MySQL에서 UCA 4.0.0의 가중치 표). DUCET을 따르므로 언어학적으로 자연스럽습니다.
- `..._0900_ai_ci`: 같은 UCA 기반이되 더 최신인 9.0.0 가중치 표(MySQL 8.0). 표가 새로워진 만큼 더 많은 문자를 정확히 다룹니다.
- `..._general_ci`: UCA 비기반의 간이 비교. 빠르지만 표가 단순해, 서로 다른 문자를 같다고 보는 경우가 UCA 기반보다 많습니다.

같은 `utf8mb4`라도 `_unicode_ci`와 `_0900_ai_ci`는 가중치 표 버전이 달라, 어떤 문자들이 같다고 접히는지가 다릅니다. 그래서 "콜레이션 버전"이 동작을 바꿉니다.

### 다층 가중치

UCA의 핵심은 가중치가 한 겹이 아니라 여러 층이라는 점입니다.

| 레벨 | 구분하는 것 | 예 |
|---|---|---|
| 1차(primary) | 글자의 본체 | `a`·`A`·`á`는 1차가 같다(모두 "a") |
| 2차(secondary) | 악센트·분음부호 | `a`와 `á`는 2차에서 갈린다 |
| 3차(tertiary) | 대소문자 | `a`와 `A`는 3차에서 갈린다 |

비교는 위에서 아래로 내려갑니다. 두 문자열의 1차 가중치열을 끝까지 견주고, 1차가 모두 같을 때만 2차를, 그것도 같으면 3차를 봅니다.

민감도 접미사가 어느 층까지 보는지를 정합니다.

- `ci`(case-insensitive): 3차(대소문자)를 무시. `a`=`A`.
- `ai`(accent-insensitive): 2차(악센트)를 무시. `a`=`á`.
- `cs`/`as`: 각각 대소문자/악센트를 본다.
- `bin`: 층 개념 없이 코드포인트(바이트)를 그대로 견준다. 무엇도 접지 않는다.

UCA 기반 `_ci`는 보통 1차 위주로 비교해 대소문자와, 많은 경우 악센트까지 함께 접습니다. 그래서 `'café' = 'cafe' = 'CAFE'`가 됩니다. 정확한 동작은 콜레이션마다 다르므로 측정으로 확인합니다.

MySQL의 `WEIGHT_STRING()`은 콜레이션이 실제로 쓰는 가중치를 그대로 돌려줍니다.

```sql
-- 같은 가중치면 같은 16진수가 나온다
SELECT HEX(WEIGHT_STRING('A' COLLATE utf8mb4_unicode_ci));
SELECT HEX(WEIGHT_STRING('a' COLLATE utf8mb4_unicode_ci));   -- 위와 같게 나오면 1차에서 같음
SELECT 'a' = 'A' COLLATE utf8mb4_unicode_ci;                 -- 1 (같음)
SELECT 'a' = 'A' COLLATE utf8mb4_bin;                        -- 0 (다름)
```

기대대로라면 ci에서는 `'a'='A'`가 1이고 두 `WEIGHT_STRING` 값이 같게 나오며, bin에서는 0이 나옵니다.

### 기호의 가변 가중치

글자는 위 규칙이 깔끔합니다. 문제는 `#` `@` `^` `*` `:` 같은 기호입니다. UCA는 기호·구두점에 가변 가중치(variable weighting)라는 별도 처리를 둡니다. 모드가 둘입니다.

1. 유의미(non-ignorable): 기호도 1차 가중치를 갖고 비교에 참여한다. `'a#b'` ≠ `'ab'`.
2. 무시(ignorable, 흔히 shifted): 기호를 1차에서 빼고 맨 마지막 층으로 미룬다. 1차만 보면 `'a#b'` = `'ab'`가 될 수 있다.

MySQL의 UCA 기반 콜레이션은 기본적으로 유의미(non-ignorable) 쪽이라 `#` 같은 기호가 비교에서 살아 있습니다. 그래서 구분자로 `#`을 쓰면 동작합니다. 다만 두 가지를 측정으로 확인해야 합니다. 어떤 기호가 서로 구별되는지, 그리고 어떤 기호 쌍이 우연히 같은 가중치로 접히지는 않는지입니다. 이것은 콜레이션과 그 버전의 가중치 표에 달렸으므로, 기억이 아니라 `WEIGHT_STRING`으로 확인합니다.

```sql
-- 두 기호가 구별되면 서로 다른 가중치가 나온다
SELECT HEX(WEIGHT_STRING('A#B' COLLATE utf8mb4_unicode_ci));
SELECT HEX(WEIGHT_STRING('A*B' COLLATE utf8mb4_unicode_ci));
SELECT 'A#B' = 'A*B' COLLATE utf8mb4_unicode_ci;   -- 0이어야 # 와 * 가 구별됨
```

### 식별자에서의 함의

이 이론이 실무에서 닿는 곳은 식별자입니다. 로그인 ID, 외부 키, 사용자 코드처럼 정확히 구별돼야 하는 문자열을 콜레이션이 접어 버리면 사고가 납니다(뒤의 INSERT·WHERE·인덱스 예시가 그 구체입니다). 그래서 식별자 설계의 안전한 기본값은 이렇습니다.

- 비교가 코드포인트 단위로 정확해야 하면 `_bin`이나 `_cs`/`_as`를 고려한다.
- 대소문자·악센트를 의도적으로 접고 싶으면(이메일 등) `_ci`/`_ai`가 맞다. 단 그 접힘에 의존하지 말고 입력을 미리 정규화(소문자 폴딩 등)해 앱이 동등성을 통제한다.
- 구분 기호는 기억으로 고르지 말고 `WEIGHT_STRING`과 실제 `UNIQUE` INSERT로 검증한 문자(또는 단순 ASCII)를 쓴다.
- `general_ci`는 식별자에 신중히 쓴다. 간이 표라 의외의 동등이 더 많다.

흔한 오해를 정리하면 이렇습니다.

- "콜레이션이 저장 형식을 바꾼다"는 틀립니다. 저장(인코딩)은 문자셋의 일이고, 콜레이션은 비교·정렬만 바꿉니다.
- "ci는 대소문자만 무시한다"도 정확하지 않습니다. UCA 기반 `_ci`는 대개 악센트도 함께 접습니다. 정확한 범위는 콜레이션마다 다릅니다.
- "기호는 비교에서 무시된다"는 콜레이션 나름입니다. MySQL UCA 기반은 기본적으로 기호를 살립니다.

스스로 확인하려면, 콜레이션을 바꿔 가며 위 `WEIGHT_STRING`·`=` 쿼리를 돌려 같다고 보는 문자 쌍이 콜레이션마다 어떻게 달라지는지 직접 봅니다.

## Collation 이름 패턴

```plaintext
{문자셋}_{언어의 지역(locale)을 기준으로 한 비교 규칙}_{비교 민감도}
```

- `{문자셋}`
    - latin1
    - utf8
    - utf8mb4
    - ucs2
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
    - `swedish`: 스웨덴어 기준 비교 알고리즘
    - `ci`: 대소문자를 구분하지 않는다(case-insensitive)는 것

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

LIKE에는 콜레이션과 별개로 *와일드카드 위치*라는 축이 하나 더 있고, 이것이 성능을 가릅니다. 와일드카드가 앞에 있느냐 뒤에 있느냐가 인덱스 사용 여부를 바꿉니다.

- `LIKE 'abc%'` (뒤 와일드카드, 접두 고정): 인덱스 범위 스캔이 가능합니다. `abc`로 시작하는 값은 정렬 순서에서 한 덩어리로 모여 있기 때문입니다.
- `LIKE '%abc'` 또는 `LIKE '%abc%'` (선두 와일드카드): 일반 B-Tree 인덱스로 못 좁힙니다. "어디에든 abc를 포함"은 정렬 순서에서 흩어져 있어, 엔진이 모든 행을 봐야 합니다(풀 스캔).

콜레이션은 여기서 접두 매칭의 *범위 경계*를 정하는 역할을 합니다. ci 콜레이션이면 `LIKE 'abc%'`가 `ABC...`·`Abc...`도 포함하도록 범위를 잡습니다. 왜 접두는 인덱스를 타고 선두 와일드카드는 못 타는지는 아래 인덱스 절에서 정렬 구조로 설명합니다.

### ORDER BY 절

```sql
SELECT name FROM products ORDER BY name;
```

- `utf8_unicode_ci`: 언어학적 기준에 맞게 정렬합니다.
- `utf8_bin`: 바이트값 순서(대문자 먼저)로 정렬합니다.

### 인덱스(B-Tree) 탐색 기준

B-Tree 인덱스는 키를 *정렬 순서*로 저장합니다. 그래서 두 가지를 잘합니다. 한 값으로 바로 찾는 동등 탐색(seek), 그리고 정렬상 연속한 구간을 훑는 범위 스캔(range scan)입니다. 정렬 기준은 콜레이션이 정하므로, 콜레이션이 다르면 인덱스에 담기는 순서 자체가 달라집니다.

이 구조가 LIKE의 두 경우를 가릅니다. 인덱스에 다음 값들이 정렬돼 있다고 합시다.

```text
apple
applet
application
banana
band
```

- `LIKE 'app%'`: 엔진은 `app`로 seek한 뒤 접두가 깨질 때까지 앞으로 읽습니다. `apple`·`applet`·`application`이 연속으로 나오고 `banana`에서 멈춥니다. 범위 `[app, apq)`만 읽으므로 인덱스를 탑니다. 사실상 `WHERE x >= 'app' AND x < 'apq'`와 같습니다.
- `LIKE '%and%'`: "and를 포함"하는 값(`band`)은 정렬 순서에서 한 덩어리로 모이지 않습니다. B-Tree는 앞에서부터 정렬돼 있어 "중간에 무엇을 포함"으로는 시작점을 잡지 못합니다. 결국 전부 훑습니다(풀 스캔).

정리하면, 접두가 고정된 `LIKE '접두%'`만 B-Tree 범위 스캔을 타고, 선두 와일드카드(`%...`)는 타지 못합니다. 선두 와일드카드를 인덱스로 처리하려면 다른 구조가 필요합니다. 예를 들어 문자열을 뒤집어 저장한 컬럼에 인덱스를 걸거나(접미사 검색을 접두 검색으로 바꿈), n-gram·전문검색(full-text) 인덱스를 씁니다.

콜레이션이 어긋나면 두 가지로 샙니다. 첫째, 정렬 순서가 기대와 달라 범위 경계가 틀어집니다. 둘째, 컬럼 콜레이션과 비교 상대(또는 함수 결과)의 콜레이션이 달라 암묵 변환이 끼면 인덱스를 못 쓰고 풀 스캔이 됩니다. 그래서 인덱스를 태우려면 비교 양쪽의 콜레이션을 일치시키고 접두 매칭 형태를 유지합니다.

`EXPLAIN`으로 확인합니다. 접두 LIKE는 보통 `type=range`에 `key`가 그 인덱스로 잡히고, 선두 와일드카드는 `type=ALL`(풀 스캔)로 나옵니다.

```sql
EXPLAIN SELECT * FROM products WHERE name LIKE 'app%';   -- type=range 기대
EXPLAIN SELECT * FROM products WHERE name LIKE '%app';   -- type=ALL 기대
```

접두 쿼리가 `range`로, 선두 와일드카드 쿼리가 `ALL`로 나오면 기대대로다. B-Tree 자체의 노드·페이지 배치는 [`deep-dive/storage-index-optimizer/07-index-structures.md`](deep-dive/storage-index-optimizer/07-index-structures.md)에서 더 깊게 다룹니다.

실무 함의가 하나 있습니다. 식별자를 `{접두}#{값}` 같은 구조 문자열로 만들고 "이 접두에 속한 것"을 `LIKE '접두#%'`로 찾으면, 선두 와일드카드가 아니므로 인덱스 범위 스캔을 탑니다. 비용은 그 접두에 속한 행 수에 비례하고, 추가 조건은 그 좁혀진 행에 대한 잔여 필터입니다. 즉 접두를 어떻게 설계하느냐가 곧 인덱스 친화성을 줍니다.

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
