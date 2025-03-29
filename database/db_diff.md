# DB diff

- [DB diff](#db-diff)
    - [스키마 비교 툴](#스키마-비교-툴)
    - [직접 diff 스크립트로 비교하기](#직접-diff-스크립트로-비교하기)
    - [Liquibase 사용한 비교](#liquibase-사용한-비교)

## 스키마 비교 툴

- [Liquibase](https://docs.liquibase.com/home.html)
    - Java 기반으로 XML, YAML, JSON, SQL 형식의 체인지셋을 지원합니다.
    - 두 데이터베이스 간의 구조(schema) 차이를 간단히 비교하고 마이그레이션 스크립트를 생성할 수 있습니다.

- [Flyway](https://flywaydb.org/documentation/)
    - 데이터베이스 마이그레이션 전문 도구로, SQL 스크립트를 버전별로 관리합니다.
    - 비교(diff) 기능보다는 명시적으로 정의된 마이그레이션을 선호합니다.

- [SchemaCrawler](https://www.schemacrawler.com/)
    - 데이터베이스의 모든 객체를 세부적으로 분석하여 차이를 추적할 수 있습니다.

- MySQL Workbench, DBeaver, DataGrip 등의 GUI 도구도 스키마 비교 기능을 내장하고 있습니다.

## 직접 diff 스크립트로 비교하기

비교하려는 DB들의 테이블 정의를 덤프하여 `diff` 명령어나 텍스트 비교 도구를 사용하여 비교할 수 있습니다.

```shell
mysqldump --no-data -u user -p database > dev_schema.sql
mysqldump --no-data -u user -p production_db > prod_schema.sql

diff dev_schema.sql prod_schema.sql
```

## Liquibase 사용한 비교

[db_diff](./examples/db_diff/)를 참고합니다.
