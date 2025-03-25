# Example

- [Example](#example)
    - [Download](#download)
    - [Command](#command)
    - [Result](#result)
        - [스키마가 다른 경우](#스키마가-다른-경우)
        - [스키마가 같은 경우](#스키마가-같은-경우)

## [Download](https://www.liquibase.com/download)

CLI를 다운로드 받습니다.

또는 MacOS 경우 brew로 설치합니다.

```sh
brew install liquibase
```

## [Command](https://docs.liquibase.com/commands/inspection/diff.html)

```sh
liquibase \
  --classpath=./mysql-connector-j/mysql-connector-j-9.2.0.jar \
  diff \
  --output-file=example \
  --diff-types=tables,columns,indexes,views \
  --url=jdbc:mysql://localhost:3307/some_service \
  --username=dev_user \
  --password=dev \
  --referenceUrl=jdbc:mysql://localhost:3308/some_service \
  --referenceUsername=prod_user \
  --referencePassword=prod
```

- `--referenceUrl`: 비교의 기준이 되는 데이터베이스를 지정합니다.
- `--url`: 비교의 대상이 되는 데이터베이스입니다.

## Result

### [스키마가 다른 경우](./whendiff/)

[스키마가 다른 경우](./whendiff/) 다음과 같은 결과가 출력됩니다.

```log
Reference Database: prod_user@192.168.97.1 @ jdbc:mysql://localhost:3308/some_service (Default Schema: some_service)
Comparison Database: dev_user@192.168.97.1 @ jdbc:mysql://localhost:3307/some_service (Default Schema: some_service)
Compared Schemas: some_service
Product Name: EQUAL
Product Version: EQUAL
Missing Column(s):
     some_service.posts.only_at_prod
Unexpected Column(s):
     some_service.only_at_dev.created_at
     some_service.users.email
     some_service.only_at_dev.extra
     some_service.only_at_dev.id
     some_service.only_at_dev.ref_id
     some_service.comments.updated_at
     some_service.posts.updated_at
Changed Column(s):
     some_service.posts.created_at
          order changed from '6' to '5'
     some_service.users.created_at
          order changed from '4' to '5'
Missing Index(s): NONE
Unexpected Index(s):
     PRIMARY UNIQUE  ON some_service.only_at_dev(id)
Changed Index(s): NONE
Missing Table(s): NONE
Unexpected Table(s):
     only_at_dev
Changed Table(s): NONE
Missing View(s): NONE
Unexpected View(s): NONE
Changed View(s): NONE
```

- `Missing`: 기준이 되는 데이터베이스에만 존재하는 경우입니다.
- `Unexpected`: 비교의 대상이 되는 데이터베이스에만 존재하는 경우입니다.

### [스키마가 같은 경우](./whensame/)

[스키마가 같은 경우](./whensame/) 같은 경우 다음과 같이 출력됩니다.

```log
Reference Database: prod_user@192.168.97.1 @ jdbc:mysql://localhost:3308/some_service (Default Schema: some_service)
Comparison Database: dev_user@192.168.97.1 @ jdbc:mysql://localhost:3307/some_service (Default Schema: some_service)
Compared Schemas: some_service
Product Name: EQUAL
Product Version: EQUAL
Missing Column(s): NONE
Unexpected Column(s): NONE
Changed Column(s): NONE
Missing Index(s): NONE
Unexpected Index(s): NONE
Changed Index(s): NONE
Missing Table(s): NONE
Unexpected Table(s): NONE
Changed Table(s): NONE
Missing View(s): NONE
Unexpected View(s): NONE
Changed View(s): NONE
```
