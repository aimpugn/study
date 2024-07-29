# DSN

- [DSN](#dsn)
    - [Data Source Name?](#data-source-name)
    - [mysql DSN](#mysql-dsn)
    - [기타](#기타)

## Data Source Name?

애플리케이션이 ODBC 데이터 소스에 연결을 요청할 때 사용하는 이름입니다. 다시 말해, ODBC 연결을 대표하는 상징적인 이름입니다.

ODBC에 연결을 할 때, 이것은 다음과 같은 연결 세부 정보를 저장합니다:
- 데이터베이스 이름
- 디렉토리
- 데이터베이스 드라이버
- 사용자 ID
- 비밀번호
- 등등

## mysql DSN

```bash
<USER_NAME>:<PASSWORD>@tcp(<HOST>:<PORT>)/<DATABASE_NAME>?loc=Asia%2FSeoul&parseTime=true&group_concat_max_len=10240
```

## 기타

- [5.3.1 Configuring a Connector/ODBC DSN on Windows with the ODBC Data Source Administrator GUI](https://dev.mysql.com/doc/connector-odbc/en/connector-odbc-configuration-dsn-windows-5-2.html)
- [data source name (DSN)](https://support.microsoft.com/en-us/topic/what-is-a-dsn-data-source-name-ae9a0c76-22fc-8a30-606e-2436fe26e89f#:~:text=More%20Information,%2C%20UserID%2C%20password%2C%20etc.)
- [Data source name](https://en.wikipedia.org/wiki/Data_source_name)
