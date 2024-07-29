# Paginate

- [Paginate](#paginate)
    - [`search_after`?](#search_after)
        - [사용법](#사용법)
        - [`sort` 필드가 `date`이지만 다른 타겟에 `date_nanos` 있는 경우](#sort-필드가-date이지만-다른-타겟에-date_nanos-있는-경우)
        - [예제](#예제)
    - [scroll API?](#scroll-api)

## [`search_after`](https://www.elastic.co/guide/en/elasticsearch/reference/current/paginate-search-results.html#search-after)?

- You can use the `search_after` parameter to retrieve the next page of hits using a set of sort values from the previous page.
- 동일한 `query`와 `sort` 값으로 여러번 검색 요청하는 것 필요

### 사용법

- 쿼리로 최초 요청

    ```json
    GET twitter/_search
    {
        "query": {
            "match": {
                "title": "elasticsearch"
            }
        },
        "sort": [{ "date": "asc" }, { "tie_breaker_id": "asc" }]
    }
    ```

- 응답은 각 hit에 대한 `sort` 배열을 포함

    ```json
    {
        "took" : 17,
        "timed_out" : false,
        "_shards" : ...,
        "hits" : {
            "total" : ...,
            "max_score" : null,
            "hits" : [
                ...
                {
                    "_index" : "twitter",
                    "_id" : "654322",
                    "_score" : null,
                    "_source" : ...,
                    "sort" : [ <---------- 
                        1463538855,
                        "654322"
                    ]
                },
                {
                    "_index" : "twitter",
                    "_id" : "654323",
                    "_score" : null,
                    "_source" : ...,
                    "sort" : [ <----------                          
                        1463538857,
                        "654323"
                    ]
                }
            ]
        }
    }
    ```

- 다음 페이지를 가져오려면 마지막 hit의 `sort` 배열을 `search_after`에 넣는다

    ```json
    GET twitter/_search
    {
        "query": {
            "match": {
                "title": "elasticsearch"
            }
        },
        "search_after": [1463538857, "654323"],
        "sort": [
            {"date": "asc"},
            {"tie_breaker_id": "asc"}
        ]
    }
    ```

- 만약 이 여러 요청간에 [`refresh`](https://www.elastic.co/guide/en/elasticsearch/reference/current/near-real-time.html)가 발생하면, 결과의 순서가 달라질 수 있고, 페이지들에 걸쳐서 일관되지 않은 결과가 발생할 수 있다. 이를 방지하기 위해 [point in time(PIT)](https://www.elastic.co/guide/en/elasticsearch/reference/current/point-in-time-api.html)를 생성해서 현재 인덱스 상태를 유지할 수 있다.

### `sort` 필드가 `date`이지만 다른 타겟에 `date_nanos` 있는 경우

> If the [`sort`](https://www.elastic.co/guide/en/elasticsearch/reference/current/date.html) field is a date in some target data streams or indices but a [`date_nanos`](https://www.elastic.co/guide/en/elasticsearch/reference/current/date_nanos.html) field in other targets, use the `numeric_type` parameter to convert the values to a single resolution and the `format` parameter to specify a [date format](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-date-format.html) for the `sort` field. Otherwise, Elasticsearch won’t interpret the search after parameter correctly in each request.

위의 설명은 Elasticsearch에서 `search_after`를 사용시 검색 대상이 되는 데이터 스트림이나 인덱스에서 정렬(`sort`) 필드의 데이터 타입이 일부는 `date`이고, 일부는 `date_nanos`인 경우에 대한 처리 방법을 설명한다

여기서 `date`와 `date_nanos`는 Elasticsearch에서 시간 데이터를 나타내는 두 가지 다른 데이터 타입

1. `date`: 밀리초 단위의 정밀도를 가지는 일반적인 날짜/시간 데이터 타입입니다.
2. `date_nanos`: 나노초 단위의 더 높은 정밀도를 가지는 날짜/시간 데이터 타입입니다.

이 문장은 다음과 같은 상황을 설명한다

- 당신이 검색하는 데이터에서, 일부는 `date` 타입으로, 다른 일부는 `date_nanos` 타입으로 시간 데이터가 저장되어 있다
- 이러한 상황에서 `search_after`를 사용하여 정렬을 할 때, 두 타입의 데이터가 서로 호환되도록 처리해야 한다

이를 위해, Elasticsearch는 다음 두 가지 파라미터를 제공한다:

1. `numeric_type` 파라미터: 이를 사용하여 `date`와 `date_nanos` 필드의 값을 하나의 해상도(single resolution, 예: 밀리초)로 변환한다. 이는 두 타입 간의 정렬 호환성을 보장한다.
2. `format` 파라미터: `sort` 필드에 대한 날짜 포맷을 지정한다. 이것은 `sort` 필드가 일관된 형식으로 처리되도록 보장합니다.

이렇게 하면 `search_after` 파라미터를 사용할 때 Elasticsearch가 각 요청에서 정렬 필드를 올바르게 해석할 수 있다. 즉, `date`와 `date_nanos` 필드가 혼재된 환경에서도 일관된 정렬 및 검색 결과를 얻을 수 있게 된다

### 예제

```json
// GET payments-v2/_search
{
  "_source": [
    "transactions.id",
    "transactions.requested_at",
    "transactions.failed_at",
    "transactions.paid_at",
    "transactions.cancelled_at",
    "transactions.updated_at",
    "transactions.status",
    "transactions.version"
  ],
  "size": 50000,
  "query": {
    "bool": {
      "must": [
        {
          "nested": {
            "path": "transactions",
            "query": {
              "bool": {
                "minimum_should_match": 1,
                "must": [{ "term": { "transactions.version": "V1" } }],
                "should": [
                  {
                    "range": {
                      "transactions.requested_at": {
                        "gte": "2023-12-28T12:28:54+09:00",
                        "lte": "2023-12-28T12:28:55+09:00"
                      }
                    }
                  },
                  {
                    "range": {
                      "transactions.paid_at": {
                        "gte": "2023-12-28T12:28:54+09:00",
                        "lte": "2023-12-28T12:28:55+09:00"
                      }
                    }
                  },
                  {
                    "range": {
                      "transactions.failed_at": {
                        "gte": "2023-12-28T12:28:54+09:00",
                        "lte": "2023-12-28T12:28:55+09:00"
                      }
                    }
                  },
                  {
                    "range": {
                      "transactions.cancelled_at": {
                        "gte": "2023-12-28T12:28:54+09:00",
                        "lte": "2023-12-28T12:28:55+09:00"
                      }
                    }
                  }
                ]
              }
            }
          }
        },
        {
          "term": {
            "merchant_id": "merchant-6f36a161-c4bb-4f4a-8729-d67beb958460"
          }
        },
        { "term": { "store_id": "store-921f4004-d467-4695-ab33-87f31d6949b6" } }
      ]
    }
  },
  "sort": [
    {
      "transactions.requested_at": {
        "nested_path": "transactions",
        "order": "desc"
      }
    }
  ]
}
```

호출 결과

```json
{
    "took": 4110,
    "timed_out": false,
    "_shards": {
        "total": 52,
        "successful": 52,
        "skipped": 0,
        "failed": 0
    },
    "hits": {
        "total": {
            "value": 8,
            "relation": "eq"
        },
        "max_score": null,
        "hits": [
            {
                "_index": "payments-v2",
                "_id": "store-921f4004-d467-4695-ab33-87f31d6949b6-billing-6574012-20231228032854475833",
                "_score": null,
                "_source": {
                    "transactions": [
                        {
                            "cancelled_at": null,
                            "paid_at": "2023-12-28T03:38:44Z",
                            "updated_at": "2023-12-28T03:38:44Z",
                            "id": "imp_011135316559",
                            "failed_at": null,
                            "version": "V1",
                            "requested_at": "2023-12-28T03:28:55Z",
                            "status": "PAID"
                        }
                    ]
                },
                "sort": [1703734135000]
            },
            {
                "_index": "payments-v2",
                "_id": "store-921f4004-d467-4695-ab33-87f31d6949b6-bid-66886998-20231228032845837166",
                "_score": null,
                "_source": {
                    "transactions": [
                        {
                            "cancelled_at": null,
                            "paid_at": "2023-12-28T03:29:20Z",
                            "updated_at": "2023-12-28T03:29:21Z",
                            "id": "imp_297135207231",
                            "failed_at": null,
                            "version": "V1",
                            "requested_at": "2023-12-28T03:28:55Z",
                            "status": "PAID"
                        }
                    ]
                },
                "sort": [1703734135000]
            }
        ]
    }
}
```

여기서 `"sort": [1703734135000]` 이 값을 재사용해야 한다

## scroll API?

> NOTE:
>
> We no longer recommend using the scroll API for deep pagination.
> If you need to preserve the index state while paging through more than 10,000 hits, use the `search_after` parameter with a point in time (`PIT`).

- 컨텍스트를 유지하면서 페이지네이션을 제공하는 API
- deep pagination(깊은 페이지네이션)?
    - 많은 양의 데이터를 여러 페이지에 걸쳐 조회할 때, 사용자가 상당히 뒤쪽 페이지(예: 1000번째 페이지 이상)에 접근하는 것을 의미
    - 이는 일반적으로 매우 많은 데이터 레코드를 갖는 검색에서 발생하며, 각 페이지에 대한 요청이 데이터의 깊은 부분으로 접근할 때까지 계속된다
- shallow pagination(얕은 페이지네이션)?
    - 사용자가 상대적으로 앞쪽 페이지(예: 처음 몇 페이지)에만 접근하는 경우
    - 이 경우, 페이지네이션은 데이터의 시작 부분에서만 발생하며, 전체 데이터셋의 극히 일부분만 조회

Elasticsearch의 문맥에서, `scroll` API는 이전에는 deep pagination을 수행하는 데 적합한 방법으로 사용되었습니다. 이 API는 한 번의 검색 요청으로 시작하여, 검색 컨텍스트를 유지하면서 계속해서 다음 데이터 세트를 가져옵니다. 그러나 이 방법은 많은 양의 데이터와 장시간에 걸친 검색 상태 유지로 인해 성능상의 문제를 일으킬 수 있습니다.

따라서 Elasticsearch는 이제 `search_after` 파라미터와 함께 Point In Time (PIT) 검색을 사용하여 deep pagination을 수행하는 것을 권장합니다. `search_after`를 사용하면 이전 검색 결과의 마지막 문서 이후로 검색을 계속할 수 있으며, PIT는 검색 중에 인덱스 상태가 변경되지 않도록 보장합니다. 이 방법은 더 효율적이며 클러스터에 덜 부담을 줍니다.
