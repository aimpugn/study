# response

## `hits`

- [Track total hits](https://www.elastic.co/guide/en/elasticsearch/reference/7.5/search-request-body.html#request-body-search-track-total-hits)

### `hits.total`

```json
{
    "took": 406,
    "timed_out": false,
    "_shards": {
        // 생략 
    },
    "hits": {
        "total": {
            "value": 8,
            "relation": "eq"
        }
    }
}
```

- 일반적으로 전체 개수는 모든 매치되는 결과를 방문하지 않으면 정확하게 계산할 수 없으며, 이는 많은 문서가 매치되는 쿼리의 경우 비용이 많이 든다
- "there are at least 10000 hits"와 같이 조회수의 하한을 설정하는 것으로 충분할 때가 많으므로, 기본값은 `10,000` 으로 설정
- 즉, 요청에 대한 전체 히트 수는 10,000까지 정확하게 카운트됨을 의미
