# OpenObserve REST API

- [OpenObserve REST API](#openobserve-rest-api)
    - [액세스 토큰 발급](#액세스-토큰-발급)
    - [API 호출 curl](#api-호출-curl)

## 액세스 토큰 발급

## API 호출 curl

```sql
SELECT 
    column 
FROM something
WHERE name = 'name' 
    AND column = 'this'
```

위와 같은 쿼리로 검색할 때, 바디는 아래와 같이 구성이 됩니다.

```json
{
  "query": {
    "from": 0,
    "size": 100,
    "sql": "SELECT column FROM something WHERE name = 'name' AND column = 'this'",
    "start_time": 1722314671721000,
    "end_time": 1722315571721000,
    "fast_mode": true,
    "sql_mode": "full",
    "track_total_hits": true
  }
}
```

네트워크 패널을 통해 확인하면, 로그인 통해 생성된 Authorization 토큰으로 아래와 같은 요청이 들어가게 됩니다.

```sh
curl -XGET \
  'https://openobserve.domain.co/api/default/_search' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Basic <TOKEN>' \
  -d '{
        "query": {
            "from": 0,
            "size": 100,
            "sql": "SELECT column FROM something WHERE name = 'name' AND column = 'this'",
            "start_time": 1722314671721000,
            "end_time": 1722315571721000,
            "fast_mode": true,
            "sql_mode": "full",
            "track_total_hits": true
        }
    }'
```
