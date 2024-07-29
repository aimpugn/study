# DSN

- [DSN](#dsn)
    - [DSN (Data Source Name)?](#dsn-data-source-name)
    - [DSN의 유래와 역사](#dsn의-유래와-역사)
    - [DSN의 사용](#dsn의-사용)
    - [DSN 예시](#dsn-예시)
        - [mysql DSN](#mysql-dsn)
        - [PostgreSQL DSN](#postgresql-dsn)
        - [SQLite DSN](#sqlite-dsn)
        - [SQL Server DSN](#sql-server-dsn)
        - [Oracle DSN](#oracle-dsn)
        - [MongoDB DSN](#mongodb-dsn)

## DSN (Data Source Name)?

DSN은 다양한 데이터베이스 시스템과 어플리케이션에서 널리 사용되는 개념으로, 데이터베이스 서버에 연결하기 위해 필요한 정보를 포함하는 문자열 또는 구성 요소를 설명합니다.

이는 ODBC (Open Database Connectivity) 표준의 일부로도 널리 알려져 있으며, 다양한 데이터베이스 및 프로그래밍 환경에서 활용됩니다.

## DSN의 유래와 역사

**ODBC와 DSN**:
DSN의 개념은 1990년대 초반 Microsoft가 ODBC를 개발하면서 더욱 명확하게 정립되었습니다.
ODBC는 다양한 데이터베이스 시스템 간의 상호 운용성을 제공하기 위해 만들어진 API 세트입니다.
이 표준은 어플리케이션에서 데이터베이스에 독립적으로 접근할 수 있게 하며, DSN은 이러한 접근을 설정하기 위한 수단으로 사용됩니다.

**DSN의 기능**:
DSN은 특정 데이터베이스 서버에 연결하기 위한 정보를 정의합니다.
이 정보에는 서버의 위치, 데이터베이스 이름, 사용자 이름과 비밀번호, 그리고 종종 추가적인 연결 옵션들이 포함될 수 있습니다.
DSN 설정은 파일 DSN, 사용자 DSN, 시스템 DSN 등 여러 형태로 저장될 수 있습니다.

## DSN의 사용

**데이터베이스 관리 시스템(DBMS)에서의 DSN**:
- **MySQL, PostgreSQL 등**: 이 데이터베이스 시스템들에서는 DSN 문자열이 각각의 연결 라이브러리나 드라이버를 통해 사용됩니다. 이 문자열은 호스트, 포트, 사용자 이름, 비밀번호 등을 포함할 수 있으며, 특정 드라이버나 라이브러리에서 요구하는 형식을 따릅니다.
- **SQL Server**: SQL Server에서는 ODBC 또는 ADO.NET과 같은 라이브러리를 통해 DSN 또는 DSN-less 연결이 사용될 수 있습니다. DSN-less 연결은 모든 필요한 정보를 프로그램 코드 내에서 직접 지정합니다.

**프로그래밍 언어에서의 DSN**:
- **Python, PHP, Go 등**: 이 언어들은 데이터베이스 연결을 위해 DSN 형식을 사용하는 라이브러리를 제공합니다. 예를 들어, Python의 SQLAlchemy나 PHP의 PDO, Go의 database/sql 패키지 등이 있습니다.

## DSN 예시

### mysql DSN

```bash
<USER_NAME>:<PASSWORD>@tcp(<HOST>:<PORT>)/<DATABASE_NAME>?loc=Asia%2FSeoul&parseTime=true&group_concat_max_len=10240
```

### PostgreSQL DSN

PostgreSQL에서의 DSN은 보통 다음과 같은 형식을 취합니다:

```bash
host=<HOST> port=<PORT> dbname=<DATABASE_NAME> user=<USER_NAME> password=<PASSWORD>
```

또는 URL 형식을 사용할 수도 있습니다:

```bash
postgresql://<USER_NAME>:<PASSWORD>@<HOST>:<PORT>/<DATABASE_NAME>?sslmode=disable
```

이 형식에서는 옵션으로 SSL 모드 설정 등 추가 파라미터를 포함할 수 있습니다.

### SQLite DSN

SQLite는 파일 기반의 데이터베이스로, DSN은 간단히 데이터베이스 파일의 경로를 지정합니다:

```bash
/path/to/database.db
```

SQLite는 서버가 없기 때문에 호스트나 포트는 필요 없습니다.

### SQL Server DSN

Microsoft SQL Server를 사용할 때는 다음과 같은 DSN 형식이 일반적입니다:

```bash
server=<HOST>;port=<PORT>;database=<DATABASE_NAME>;user id=<USER_NAME>;password=<PASSWORD>
```

ODBC 연결을 위해서는 다음과 같이 설정할 수 있습니다:

```bash
Driver={SQL Server};Server=<HOST>,<PORT>;Database=<DATABASE_NAME>;Uid=<USER_NAME>;Pwd=<PASSWORD>;
```

### Oracle DSN

Oracle 데이터베이스의 경우 다음과 같은 형식을 사용할 수 있습니다:

```bash
<USER_NAME>/<PASSWORD>@<HOST>:<PORT>/<DB_SERVICE_NAME>
```

Oracle에서는 SID 또는 서비스 이름을 사용하여 특정 데이터베이스 인스턴스에 연결합니다.

### MongoDB DSN

MongoDB의 경우 URL 형식의 DSN을 사용합니다:

```bash
mongodb://<USER_NAME>:<PASSWORD>@<HOST>:<PORT>/<DATABASE_NAME>
```

MongoDB에서는 또한 복제 세트에 연결하기 위해 여러 호스트와 포트를 지정할 수 있습니다.

```bash
mongodb://<USER_NAME>:<PASSWORD>@<HOST1>:<PORT1>,<HOST2>:<PORT2>/<DATABASE_NAME>?replicaSet=<REPLICA_SET_NAME>
```
