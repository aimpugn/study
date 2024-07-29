# Queries

## [search nested field matching](https://opensearch.org/docs/1.3/opensearch/supported-field-types/nested/)

```json
// GET _search 
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "kubernetes.container_name": "tx-gateway-service"
          }
        }
      ]
    }
  }
}
```

## keyword 검색

검색하고자 하는 필드 마지막에 [`keyword`를 붙이면](https://stackoverflow.com/a/48875105) 된다. 물론 당연히 mapping 생성 시에 keyword 속성이 선언되어 있어야 한다

```json
// GET _search
{
  "query": {
    "bool": {
      "must": [
        {
          "term": {
            "kubernetes.container_name.keyword": "tx-gateway-service"
          }
        },
        {
          "match": {
            "message": "x509"
          }
        }
      ]
    }
  }
}
```

## [Simple term query not working with elastic while match works](https://stackoverflow.com/questions/52412359/simple-term-query-not-working-with-elastic-while-match-works)
