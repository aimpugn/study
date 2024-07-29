# [update](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html)

- [update](#update)
    - [Update part of a document](#update-part-of-a-document)
    - [특정 필드 업데이트 예제](#특정-필드-업데이트-예제)
        - [history 배열에서 id 4 제거](#history-배열에서-id-4-제거)
        - [history 배열을 새 값으로 덮어쓰기](#history-배열을-새-값으로-덮어쓰기)
        - [이름 수정](#이름-수정)

## [Update part of a document](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-update.html#_update_part_of_a_document)

```shell
curl -X POST "localhost:9200/test/_update/1?pretty" -H 'Content-Type: application/json' -d'
{
  "doc": {
    "name": "new_name"
  }
}
'
```

## 특정 필드 업데이트 예제

- Elasticsearch에서 `nested` 객체의 특정 필드를 업데이트하는 것은 nested 객체의 복잡성 때문에 쉽지 않음
- `nested` 객체는 **배열의 각 객체를 독립된 하위 문서로 취급하여 저장**하므로, 배열 내의 특정 객체만 업데이트하는 것이 어려움
- 그래도 스크립트를 사용하여 `nested` 객체의 특정 필드를 업데이트 할 수 있다

가령 아래와 같은 맵핑의 인덱스가 있다고 보고 정리

```json
// PUT /user-index
{
  "mappings": {
    "properties": {
      "name": {"type": "text"},
      "age": {"type": "integer"},
      "history": {
        "type": "nested",
        "properties": {
          "id": {"type": "integer"},
          "activity": {"type": "keyword"}
        }
      }
    }
  }
}
```

문서 삽입:

```json
// POST /user-index/_doc/1
{
  "name": "user name",
  "age": 30,
  "history": [
    {"id": 1, "activity": "loggedin"},
    {"id": 2, "activity": "loggedout"},
    {"id": 3, "activity": "loggedin"},
    {"id": 4, "activity": "posted"}
  ]
}
```

### history 배열에서 id 4 제거

```json
// POST /user-index/_doc/1/_update
{
  "script" : {
    "source": "ctx._source.history.removeIf(item -> item.id == params.id)",
    "params": {
      "id": 4
    }
  }
}
```

### history 배열을 새 값으로 덮어쓰기

```json
// POST /user-index/_doc/1/_update
{
  "doc": {
    "history": [
      {"id": 5, "activity": "liked"},
      {"id": 6, "activity": "shared"},
      {"id": 7, "activity": "commented"},
      {"id": 8, "activity": "loggedout"}
    ]
  }
}
```

### 이름 수정

```json
// POST /user-index/_doc/1/_update
{
  "doc": {
    "name": "new user name"
  }
}
```
