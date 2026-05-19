# Search, NoSQL, NewSQL Lab Notes

이 파일은 DU53-DU56의 실행 실험을 위한 안전한 관측 절차입니다.
검색 엔진, Firestore, distributed SQL은 로컬 환경과 계정 비용이 다르므로, 운영 계정이나 실제 인덱스에 바로 실행하지 않습니다.

## DU53-DU54: Elasticsearch/OpenSearch

확인할 질문은 RDBMS B+tree index와 inverted index가 무엇을 다르게 저장하는지입니다.
작은 synthetic document index를 만들고 `title`, `body`, `created_at`, `tenant_id`를 넣습니다.
그다음 `match`, `term`, `range`, `sort`, `search_after`, PIT(point in time) 또는 OpenSearch의 동등 기능을 비교합니다.

```json
{
  "query": {
    "bool": {
      "must": [{ "match": { "body": "transaction recovery" } }],
      "filter": [{ "term": { "tenant_id": 1 } }]
    }
  },
  "sort": [
    { "created_at": "desc" },
    { "_id": "asc" }
  ],
  "size": 10
}
```

PASS는 검색 결과가 나온다는 뜻이 아닙니다.
`match`가 분석기를 거친 token을 찾고, `term`은 정확한 term을 찾으며, 깊은 `from/size` pagination이 왜 비용과 consistency 문제를 만드는지 설명할 수 있어야 합니다.

## DU55: Firestore/document modeling

확인할 질문은 document DB에서 join 대신 어떤 중복과 읽기 경로를 선택하는지입니다.
예를 들어 `users/{userId}/orders/{orderId}`와 `orders/{orderId}` top-level collection을 비교하고, security rules가 어떤 document path와 request auth를 기준으로 판단하는지 적습니다.
PASS는 “정규화를 버린다”가 아니라, 읽기 경로, rule 경계, 비용, hot spot을 같이 설명하는 것입니다.

## DU56: Distributed SQL

확인할 질문은 SQL을 유지하면서 consensus와 분산 transaction이 들어오면 latency와 failure mode가 어떻게 바뀌는지입니다.
CockroachDB 또는 TiDB의 공개 문서와 로컬 playground가 있다면, 단일 row transaction과 cross-range/cross-region transaction의 latency 차이를 비교합니다.
PASS는 `SERIALIZABLE` 같은 이름을 보는 것이 아니라, transaction이 어떤 key range, raft group, region, timestamp oracle 또는 transaction coordinator를 지나가는지 설명할 수 있는 것입니다.
