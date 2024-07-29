# Queries

- [Queries](#queries)
    - [특정 필드만 가져오기](#특정-필드만-가져오기)
    - [boolean query](#boolean-query)
        - [with range](#with-range)
    - [bool \& nested](#bool--nested)
    - [nested \&\& 단일 term 쿼리](#nested--단일-term-쿼리)
    - [`minimum_should_match`](#minimum_should_match)

## 특정 필드만 가져오기

```json
GET <index_name>/_search
{
  "_source": [
    "list.field1_of_nested_obejct",
    "list.field2_of_nested_obejct",
    "list.field3_of_nested_obejct"
  ],
  "query": {
    "match": {
      "property": "some value"
    }
  },
  "sort": [
    {
      "created_at": "desc"
    }
  ]
}
```

## boolean query

### with range

```json
GET /restored_port-dev-eks-logger-2023.09.13-000254/_search
{
    "query": {
        "bool": {
            "must": [
                {
                    "term": {
                        "kubernetes.container_name": {
                            "value": "novelist-core"
                        }
                    }
                },
                {
                    "term": {
                        "kubernetes.pod_name": {
                            "value": "novelist-core-586f6b6795-x5xl7"
                        }
                    }
                },
                {
                    "range": {
                        "timestamp": {
                            "gte": "2023-09-13T09:34:00.000Z",
                            "lte": "2023-09-13T09:47:00.000Z"
                        }
                    }
                }
            ]
        }
    }
}
```

## bool & nested

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "nested": {
            "path": "transactions",
            "query": {
              "bool": {
                "minimum_should_match": 1,
                "must": [
                  {
                    "term": {
                      "transactions.version": "V1"
                    }
                  },
                  {
                    "term": {
                      "transactions.is_primary": true
                    }
                  }
                ],
                "should": [
                  {
                    "range": {
                      "transactions.requested_at": {
                        "gte": "2022-08-01T00:00:00+09:00",
                        "lte": "2022-10-30T00:00:00+09:00"
                      }
                    }
                  },
                  {
                    "range": {
                      "transactions.paid_at": {
                        "gte": "2022-08-01T00:00:00+09:00",
                        "lte": "2022-10-30T00:00:00+09:00"
                      }
                    }
                  },
                  {
                    "range": {
                      "transactions.failed_at": {
                        "gte": "2022-08-01T00:00:00+09:00",
                        "lte": "2022-10-30T00:00:00+09:00"
                      }
                    }
                  },
                  {
                    "range": {
                      "transactions.cancelled_at": {
                        "gte": "2022-08-01T00:00:00+09:00",
                        "lte": "2022-10-30T00:00:00+09:00"
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
            "merchant_id": "merchant-e03bdca8-9a9e-4b26-be51-1ca9fe6b4045"
          }
        },
        {
          "term": {
            "store_id": "store-24a44428-eee0-4a9e-9d31-e221b955461a"
          }
        }
      ]
    }
  },
  "size": 10000
}
```

## nested && 단일 term 쿼리

```json
{
    "query": {
        "nested": {
            "path": "transactions",
            "query": {
                "bool": {
                    "must": [
                        {
                            "term": {
                                "transactions.id.keyword": "imp_919437393916"
                            }
                        }
                    ]
                }
            }
        }
    }
}

```

## [`minimum_should_match`](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-minimum-should-match.html)

얼마나 많은 "should"절이 반드시 매치되어야 하는지 명시하는 파라미터로,
"bool" 쿼리의 "should" 절에서 사용한다.
