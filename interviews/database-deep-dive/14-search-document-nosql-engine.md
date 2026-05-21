# 검색 엔진과 문서형 NoSQL은 RDBMS와 어떤 다른 계약을 가지는가?

Elasticsearch, OpenSearch, Firestore는 모두 전통적인 RDBMS와 다른 저장·조회 모델을 제공합니다. 하지만 "NoSQL"이라는 한 단어로 묶으면 가장 중요한 차이를 잃습니다. Elasticsearch와 OpenSearch는 문서를 저장하지만 핵심 역할은 검색 엔진입니다. text를 token으로 쪼개 inverted index를 만들고, relevance score로 정렬하며, shard와 replica 위에서 분산 검색을 수행합니다. Firestore는 document database입니다. collection/document 경로에 JSON-like document를 저장하고, index 기반 query와 security rules, pricing model, hotspot 제약을 중심으로 설계합니다.

검색 엔진은 "문서를 넣으면 SQL처럼 찾아준다"가 아닙니다. mapping이 field type을 정하고, analyzer가 text를 token으로 바꾸며, inverted index가 term에서 document 목록으로 가는 길을 만듭니다. 검색 결과 정렬은 단순 최신순이 아니라 BM25 같은 scoring 모델, filter context, sort, pagination 방식에 따라 달라집니다. Firestore도 "JSON을 저장하는 DB"로만 보면 안 됩니다. 문서 크기, index fanout, transaction/batch limit, security rules evaluation, read/write 과금, sequential key hotspot이 설계 제약이 됩니다.

이 문서는 검색 엔진과 문서형 데이터베이스를 RDBMS와 대비해 설명합니다. 핵심은 "왜 검색 index는 RDBMS B-tree와 다르게 생겼는가", "분산 shard에서 검색 결과가 어떻게 만들어지는가", "deep pagination과 reindex가 왜 어려운가", "Firestore document modeling이 왜 query와 security/cost에서 시작해야 하는가"를 독자가 다시 설명할 수 있게 만드는 것입니다.
- [2-5분 개요](#2-5분-개요)
- [먼저 잡아야 할 작은 모델](#먼저-잡아야-할-작은-모델)
- [깊은 메커니즘](#깊은-메커니즘)
    - [Inverted index와 analyzer: 검색은 저장할 때 이미 시작된다](#inverted-index와-analyzer-검색은-저장할-때-이미-시작된다)
    - [Mapping은 나중에 쉽게 바뀌지 않는 계약이다](#mapping은-나중에-쉽게-바뀌지-않는-계약이다)
    - [Segment, refresh, merge: index는 즉시 하나의 파일이 되지 않는다](#segment-refresh-merge-index는-즉시-하나의-파일이-되지-않는다)
    - [Segment flush와 OS cache는 visibility와 durability를 분리한다](#segment-flush와-os-cache는-visibility와-durability를-분리한다)
    - [Shard와 replica: 분산 검색은 fan-out/fan-in이다](#shard와-replica-분산-검색은-fan-outfan-in이다)
    - [Query와 scoring: filter와 query context를 나눈다](#query와-scoring-filter와-query-context를-나눈다)
    - [Pagination: from/size, search_after, PIT의 경계](#pagination-fromsize-search_after-pit의-경계)
    - [Reindex와 alias cutover: 검색 index는 재생 가능한 파생물이어야 한다](#reindex와-alias-cutover-검색-index는-재생-가능한-파생물이어야-한다)
    - [Snapshot, dump, restore: backup과 export의 목적을 분리한다](#snapshot-dump-restore-backup과-export의-목적을-분리한다)
    - [Firestore document modeling: query에서 거꾸로 설계한다](#firestore-document-modeling-query에서-거꾸로-설계한다)
    - [Firestore security rules, cost, hotspot](#firestore-security-rules-cost-hotspot)
- [DBMS별 경계](#dbms별-경계)
- [직접 재생해 보기](#직접-재생해-보기)
    - [text와 keyword 차이 확인](#text와-keyword-차이-확인)
    - [analyzer 결과 보기](#analyzer-결과-보기)
    - [search_after와 PIT 흐름](#search_after와-pit-흐름)
    - [alias 기반 reindex cutover](#alias-기반-reindex-cutover)
    - [Firestore document modeling 점검](#firestore-document-modeling-점검)
    - [Firestore security rules 사고 실험](#firestore-security-rules-사고-실험)
- [면접 꼬리 질문](#면접-꼬리-질문)
    - [Inverted index를 RDBMS B-tree와 비교하면 어떻게 설명하나요?](#inverted-index를-rdbms-b-tree와-비교하면-어떻게-설명하나요)
    - [`text`와 `keyword` field 차이는 무엇인가요?](#text와-keyword-field-차이는-무엇인가요)
    - [Mapping을 잘못 잡으면 왜 reindex가 필요한가요?](#mapping을-잘못-잡으면-왜-reindex가-필요한가요)
    - [shard 수는 많을수록 좋은가요?](#shard-수는-많을수록-좋은가요)
    - [`from/size` pagination은 왜 깊은 페이지에서 문제가 되나요?](#fromsize-pagination은-왜-깊은-페이지에서-문제가-되나요)
    - [Firestore에서 데이터 모델은 어떻게 시작하나요?](#firestore에서-데이터-모델은-어떻게-시작하나요)
    - [Firestore security rules가 있으면 server authorization은 필요 없나요?](#firestore-security-rules가-있으면-server-authorization은-필요-없나요)
- [함정 질문](#함정-질문)
    - ["Elasticsearch에 넣었는데 바로 검색이 안 되면 저장 실패인가요?"](#elasticsearch에-넣었는데-바로-검색이-안-되면-저장-실패인가요)
    - ["Replica가 있으니 snapshot backup은 없어도 되죠?"](#replica가-있으니-snapshot-backup은-없어도-되죠)
    - ["OpenSearch는 Elasticsearch와 완전히 호환되죠?"](#opensearch는-elasticsearch와-완전히-호환되죠)
    - ["Firestore는 document DB니까 아무 JSON이나 넣고 나중에 query하면 되죠?"](#firestore는-document-db니까-아무-json이나-넣고-나중에-query하면-되죠)
    - ["Offset pagination은 Firestore에서도 검색 엔진에서도 UI만의 문제 아닌가요?"](#offset-pagination은-firestore에서도-검색-엔진에서도-ui만의-문제-아닌가요)
    - ["검색 index를 primary DB로 쓰면 RDBMS가 필요 없나요?"](#검색-index를-primary-db로-쓰면-rdbms가-필요-없나요)
- [더 깊게 볼 자료](#더-깊게-볼-자료)

## 2-5분 개요

Elasticsearch와 OpenSearch의 가장 작은 모델은 inverted index입니다. RDBMS B-tree가 key에서 row로 좁혀 들어간다면, inverted index는 term에서 그 term을 포함한 document 목록으로 갑니다. "database engine"이라는 단어보다 "검색 가능한 색인 파일을 만들고, query time에 term posting list를 조합한다"는 그림이 먼저입니다. `"quick brown fox"`라는 text를 analyzer가 `quick`, `brown`, `fox` 같은 token으로 만들면, inverted index는 각 token이 어떤 document에 나타났는지 저장합니다. 검색할 때 query도 같은 analyzer 계열 처리를 거쳐 term이 되고, matching document가 score와 함께 나옵니다. B-tree와 cost model의 기본 감각은 [인덱스와 optimizer](04-index-query-optimizer.md)를 먼저 떠올리면 차이가 더 선명합니다.

Mapping은 field의 의미를 고정합니다. `text` field는 analyzer를 거쳐 full-text search 대상이 되고, `keyword` field는 정확한 값 match, aggregation, sorting에 적합합니다. 숫자, 날짜, boolean, object, nested field도 서로 다른 indexing 방식을 가집니다. 한번 잘못 mapping된 field는 나중에 간단히 바꾸기 어렵습니다. 대개 새 index를 만들고 reindex해야 합니다. 그래서 검색 엔진 설계에서는 document를 넣기 전에 "어떤 field를 어떤 방식으로 검색, filter, sort, aggregate할 것인가"를 먼저 정해야 합니다.

Shard와 replica는 분산 검색의 기본 단위입니다. Primary shard는 data를 나눠 갖고, replica shard는 고가용성과 read capacity를 돕습니다. Query는 coordinating node가 여러 shard에 요청을 보내고, 각 shard의 top result를 모아 전체 top result를 만듭니다. 이 구조 때문에 shard 수는 단순 병렬성 knob가 아닙니다. shard가 너무 많으면 cluster metadata와 query fan-out 비용이 커지고, 너무 적으면 scale-out과 recovery가 어려워집니다.

Pagination은 검색 엔진의 대표 함정입니다. `from + size`는 앞 페이지 result를 건너뛰기 위해 각 shard가 더 많은 candidate를 유지해야 하므로 깊은 페이지에서 비싸집니다. 안정적인 깊은 pagination에는 `search_after`와 point-in-time(PIT) 같은 방식을 씁니다. PIT는 검색 기준 snapshot을 일정 시간 유지해, refresh와 segment merge 사이에서도 같은 logical view를 따라가게 돕습니다. 단, PIT도 resource를 잡으므로 무한히 열어 두면 안 됩니다.

Reindex, dump, restore는 consistency 질문의 중심입니다. 검색 index는 보통 primary source of truth가 아니라 RDBMS나 event log에서 파생된 read model인 경우가 많습니다. Reindex할 때는 old index와 new index의 mapping 차이, alias cutover, write 중 변경분 반영, backfill 중 문서 삭제 처리, refresh timing을 설계해야 합니다. Snapshot/restore는 cluster state와 index data를 특정 시점으로 보존하지만, application source DB와 같은 시점으로 맞추려면 별도 절차가 필요합니다. 파생 read model을 안전하게 유지하는 문제는 [애플리케이션 경계, 멱등성, 돈, outbox](12-application-boundaries-idempotency-money-outbox.md)의 outbox/멱등 처리와 직접 연결됩니다.

Firestore의 작은 모델은 collection 안의 document와 자동/복합 index입니다. Firestore query는 대부분 index를 필요로 하고, 쿼리 모양이 곧 index와 비용을 결정합니다. Security rules는 client가 직접 Firestore에 접근하는 모델에서 권한의 핵심 경계입니다. 읽기/쓰기 과금은 "요청 한 번"이 아니라 document read/write/delete와 index 작업에 영향을 받습니다. Sequential document id나 단일 field에 write가 몰리면 hotspot이 생길 수 있으므로, key 설계와 sharding counter 같은 패턴을 알아야 합니다.

## 먼저 잡아야 할 작은 모델

검색 엔진의 작은 모델은 책 뒤의 색인과 비슷합니다. 책의 본문을 처음부터 끝까지 훑지 않고, 색인에서 단어를 찾아 그 단어가 나온 페이지 목록을 봅니다. Elasticsearch/OpenSearch의 inverted index도 term에서 document id 목록으로 갑니다.

```text
document 1: "spring transaction rollback"
document 2: "database transaction isolation"
document 3: "spring boot database connection"

analyzer 결과:
  spring      -> doc 1, doc 3
  transaction -> doc 1, doc 2
  database   -> doc 2, doc 3
  rollback   -> doc 1
  isolation  -> doc 2
  connection -> doc 3
```

사용자가 `spring transaction`을 검색하면 검색 엔진은 `spring` posting list와 `transaction` posting list를 조합합니다. 두 term을 모두 가진 doc 1은 높은 점수를 받을 수 있고, 하나만 가진 doc 2, doc 3은 query 종류에 따라 제외되거나 낮은 점수를 받을 수 있습니다. 여기서 score는 단순 포함 여부가 아닙니다. term frequency, inverse document frequency, field length normalization 같은 요소가 들어갑니다. Elasticsearch와 OpenSearch는 Lucene 기반이므로 BM25 계열 scoring을 기본으로 이해하면 됩니다.

검색이 실행될 때의 작은 trace를 더 붙이면 B-tree와 차이가 선명합니다.

```text
query text: "spring transaction"
  -> analyzer: spring, transaction

posting lists:
  spring       -> [doc1, doc3]
  transaction  -> [doc1, doc2]

bool must query라면:
  intersection -> [doc1]

bool should query라면:
  union candidates -> [doc1, doc2, doc3]
  score 계산 -> doc1이 두 term을 모두 가져 더 위에 올 수 있음
```

이 흐름에서 inverted index는 row를 정렬된 key 순서로 내려가는 구조가 아니라 term이 가리키는 후보 document 집합을 조합하는 구조입니다. 그래서 검색 엔진의 성능 질문은 "어떤 term 후보가 얼마나 커지는가", "filter로 후보를 얼마나 줄이는가", "score와 sort가 shard마다 얼마나 많은 후보를 들고 있어야 하는가"로 이어집니다.

Mapping은 이 색인을 어떻게 만들지 정합니다.

```json
{
  "mappings": {
    "properties": {
      "title": { "type": "text", "analyzer": "standard" },
      "title_keyword": { "type": "keyword" },
      "created_at": { "type": "date" },
      "price": { "type": "scaled_float", "scaling_factor": 100 },
      "tags": { "type": "keyword" }
    }
  }
}
```

`title`은 full-text search에 적합합니다. `title_keyword`는 exact match, sort, aggregation에 적합합니다. `created_at`은 range query와 sort에 쓰이고, `price`는 decimal-like 가격 검색을 위해 scale을 고정할 수 있습니다. `tags`는 tokenization 없이 값 단위 match에 적합합니다. 같은 문자열 field라도 text와 keyword는 완전히 다른 질문에 답합니다.

같은 값도 field type에 따라 index에 남는 모양이 달라집니다.

| 원본 값 | field type | 검색 가능한 모양 | 잘 맞는 질문 |
| --- | --- | --- | --- |
| `"Spring Transaction Rollback"` | `text` | `spring`, `transaction`, `rollback` token | "rollback 관련 글을 찾아줘" |
| `"Spring Transaction Rollback"` | `keyword` | 전체 문자열 하나 | 제목이 정확히 이 값인가 |
| `"2026-05-20T00:00:00Z"` | `date` | 날짜 range/sort용 값 | 특정 기간 이후 문서인가 |
| `10000` | numeric 계열 | range/sort/aggregation용 값 | 가격이 1만 원 이상인가 |

Mapping을 잘못 고르면 검색 품질 문제와 운영 문제가 같이 생깁니다. 예를 들어 주문 상태 `PAID`를 `text`로만 두면 exact aggregation이나 sort에서 불필요한 비용을 만들 수 있고, 상품 설명을 `keyword`로만 두면 부분 단어 검색이 거의 되지 않습니다.

Firestore의 작은 모델은 다음처럼 다릅니다.

```text
/users/{userId}
  name: "Rody"
  status: "ACTIVE"

/users/{userId}/orders/{orderId}
  amount: 10000
  createdAt: timestamp
  state: "PAID"
```

Firestore는 join 중심으로 생각하지 않습니다. 자주 같이 읽는 데이터를 문서에 중복해서 넣거나, subcollection으로 분리하거나, query별 collection group을 설계합니다. 문서 하나가 너무 커지면 read/write 비용과 contention이 커지고, 너무 잘게 쪼개면 읽기 횟수와 consistency 관리가 어려워집니다. Query는 "어떤 collection에서 어떤 field 조건과 orderBy를 쓰는가"로 제한되고, 필요한 composite index를 만들어야 합니다.

작은 경계는 이렇게 정리할 수 있습니다.

```text
Elasticsearch/OpenSearch
  source document -> analyzer/mapping -> inverted index/columnar doc values
  query -> shard fan-out -> score/sort -> top results

Firestore
  application document -> collection/document path -> automatic/composite index
  query -> index scan -> document reads -> security rules check/cost
```

검색 엔진은 "관련도 높은 결과를 빠르게 찾는 read model"로 보는 편이 안전하고, Firestore는 "document 단위로 저장하고 index된 query를 수행하는 operational database"로 보는 편이 안전합니다. 둘 다 JSON 비슷한 document를 다루지만, 설계 질문의 중심축이 다릅니다.

## 깊은 메커니즘

### Inverted index와 analyzer: 검색은 저장할 때 이미 시작된다

Elasticsearch/OpenSearch에서 document를 index하면 `_source`만 저장되는 것이 아닙니다. Field type과 analyzer에 따라 검색용 index structure가 만들어집니다. `text` field는 character filter, tokenizer, token filter를 거쳐 term stream이 됩니다. standard analyzer는 대체로 punctuation과 whitespace를 기준으로 token을 만들고 소문자화 같은 처리를 합니다. Korean, Japanese, Chinese처럼 띄어쓰기와 형태소가 중요한 언어에서는 기본 analyzer만으로 좋은 검색 품질을 얻기 어렵습니다. ngram, edge ngram, 형태소 analyzer, synonym filter를 검토해야 합니다.

Inverted index가 필요한 이유는 사람의 검색어가 row key처럼 정확히 정렬되어 들어오지 않기 때문입니다. RDBMS B-tree는 `user_id = 10`, `created_at BETWEEN ...`처럼 비교 가능한 key를 좁혀 들어가는 데 강합니다. 하지만 사용자는 "스프링 트랜잭션 롤백"처럼 문장 조각을 입력하고, 문서에는 "Spring transaction rollback", "트랜잭션이 rollback된다"처럼 다른 표기가 섞입니다. 검색 엔진은 문서를 저장하는 순간부터 단어 후보를 뽑아 term 중심의 길을 만들어 둡니다. 나중에 query가 오면 본문 전체를 훑는 대신, query term이 가리키는 posting list를 조합합니다.

```text
RDBMS B-tree mental model
  key(status='PAID') -> matching rows

search inverted-index mental model
  term('rollback') -> documents containing rollback-like token
  term('transaction') -> documents containing transaction-like token
  combine terms -> score and rank candidate documents
```

그래서 analyzer는 저장 전처리 옵션이 아니라 검색 가능성의 일부입니다. 어떤 token을 만들었는지가 나중에 어떤 query가 맞을 수 있는지를 정합니다.

Analyzer는 index time과 search time에 모두 관련됩니다. Index time analyzer가 `running`을 `run`으로 stem 처리했는데 search analyzer가 다르면 query term과 index term이 맞지 않을 수 있습니다. Synonym을 index time에 확장할지 search time에 확장할지도 trade-off가 있습니다. Index time synonym은 query가 단순해질 수 있지만 synonym 변경 시 reindex가 필요할 수 있습니다. Search time synonym은 변경이 유연하지만 query expansion 비용과 scoring 영향을 고려해야 합니다.

`keyword` field는 analyzer로 token을 쪼개지 않고 전체 값을 하나의 term처럼 다룹니다. 그래서 exact match, terms aggregation, sorting에 적합합니다. 반대로 `text` field에 대해 sorting이나 aggregation을 하려면 fielddata를 켜야 할 수 있는데, 이는 memory 비용이 커질 수 있습니다. 일반적으로 같은 논리 field를 `text`와 `keyword` multi-field로 두어 검색과 정렬/집계를 나눕니다.

Analyzer 선택은 데이터가 들어간 뒤의 검색 결과를 바꿉니다.

```text
input: "Spring Transaction Rollback"

standard analyzer 느낌:
  spring, transaction, rollback

keyword field:
  Spring Transaction Rollback

edge ngram autocomplete 예:
  s, sp, spr, spri, sprin, spring ...
```

자동완성을 위해 edge ngram을 쓰면 prefix 검색은 좋아질 수 있지만 index entry가 늘어납니다. 한국어 검색에서 형태소 analyzer를 쓰면 띄어쓰기나 조사 처리를 더 잘할 수 있지만 사전, 동의어, 품질 평가가 필요합니다. Analyzer는 "검색 품질 옵션"이면서 "색인 크기와 reindex 비용을 바꾸는 저장 계약"입니다.

### Mapping은 나중에 쉽게 바뀌지 않는 계약이다

Elasticsearch/OpenSearch는 dynamic mapping으로 field type을 자동 추론할 수 있습니다. 편리하지만 운영에서는 위험합니다. 처음 들어온 값이 `"00123"` 문자열인지 숫자인지에 따라 field type이 잡히고, 이후 다른 형태의 값이 들어오면 indexing error가 발생할 수 있습니다. 날짜 format도 마찬가지입니다. 중요한 index는 explicit mapping과 index template을 사용해 field type을 고정하는 편이 안전합니다.

Mapping 변경은 RDBMS ALTER TABLE처럼 항상 간단하지 않습니다. 이미 index된 field의 analyzer나 type을 바꾸려면 보통 새 index를 만들고 reindex해야 합니다. 이때 alias를 사용해 `products_v1`에서 `products_v2`로 읽기/쓰기 alias를 전환할 수 있습니다. Reindex 중 새 write가 들어오는 시스템이면 dual-write, change capture, version check, alias cutover 순서를 설계해야 합니다.

Nested object도 함정입니다. 일반 object array는 내부 field가 flatten되어 서로 다른 object의 field가 잘못 조합될 수 있습니다. 예를 들어 한 상품에 `variants: [{color:red,size:M}, {color:blue,size:L}]`이 있을 때 일반 object mapping에서는 `color:red AND size:L`이 같은 variant에 있는 것처럼 match될 수 있습니다. `nested` type은 각 object를 별도 hidden document처럼 저장해 object 내부 관계를 유지하지만, query와 storage 비용이 늘어납니다.

### Segment, refresh, merge: index는 즉시 하나의 파일이 되지 않는다

Lucene 기반 검색 엔진은 immutable segment를 만듭니다. 새 document는 buffer에 있다가 refresh 후 검색 가능한 segment가 됩니다. 그래서 Elasticsearch/OpenSearch는 near real-time search라고 부릅니다. Index API가 성공했다고 해서 바로 검색 결과에 보이는 것은 아닙니다. refresh interval이 지나거나 explicit refresh가 일어나야 검색 가능해집니다. 반면 get by id는 translog나 realtime get 경로로 더 빨리 보일 수 있습니다.

Segment가 immutable한 이유는 검색과 색인을 동시에 처리해야 하는 시스템에서 이미 열려 있는 검색 view를 계속 안전하게 읽게 하기 위해서입니다. 쓰기가 들어올 때마다 기존 큰 index 파일을 제자리 수정하면, 검색 중인 reader와 writer가 같은 구조를 두고 복잡하게 충돌합니다. 대신 새 document는 새 segment로 열고, 삭제나 update는 표시를 남긴 뒤, 나중에 merge가 여러 segment를 새 segment로 합칩니다. 이 방식은 읽기 concurrency를 단순하게 만들지만, update-heavy workload에서는 deleted document와 merge I/O라는 비용을 남깁니다.

```text
before update
  segment S1: doc42(title="old") searchable

update doc42
  S1의 doc42는 deleted 표시
  segment S2: doc42(title="new") 추가

later merge
  S1 + S2 -> S3
  deleted old doc42는 merged output에서 빠짐
```

이 흐름 때문에 검색 엔진의 update를 RDBMS row update처럼 제자리 한 줄 변경으로 이해하면 안 됩니다. Refresh, deleted docs, merge, disk watermark가 운영 지표로 나오는 이유가 여기에 있습니다.

Segment는 시간이 지나며 merge됩니다. Delete나 update는 기존 document를 제자리에서 고치는 것이 아니라 old document를 삭제 표시하고 새 document를 추가하는 방식에 가깝습니다. Merge가 old deleted document를 정리합니다. Update-heavy workload에서 segment merge와 deleted document 비율이 성능에 영향을 주는 이유입니다. Force merge는 read-only index에서는 도움이 될 수 있지만, active write index에 무작정 실행하면 큰 I/O 부하를 만들 수 있습니다.

Translog는 crash recovery에 쓰입니다. Index operation이 segment에 완전히 반영되기 전에도 translog에 남아 있으면 recovery 때 재생할 수 있습니다. Search engine도 durability와 near-real-time visibility를 분리해서 봐야 합니다. "Index API 성공"은 durability와 refresh visibility, replica acknowledgement 설정을 함께 읽어야 정확합니다.

Index 요청 하나를 시간축으로 보면 refresh와 flush가 서로 다른 일을 한다는 점이 드러납니다.

```text
T1 index request accepted
  operation is written to indexing buffer / translog path

T2 refresh
  in-memory indexed data becomes a new searchable segment
  search can see the document

T3 more writes and deletes
  old document is marked deleted, new version is added

T4 merge
  smaller segments are merged
  deleted documents can be physically dropped from merged output

T5 flush / fsync boundary
  recovery point and translog relation are advanced according to engine settings
```

이 trace에서 `T2`는 사용자 검색 결과의 가시성에 가깝고, `T5`는 장애 뒤 복구 가능성과 더 가깝습니다. 그래서 "방금 넣었는데 검색이 안 된다"는 refresh 지연일 수 있고, "장애 뒤 복구 가능한가"는 translog, replica, snapshot 경계를 봐야 합니다.

### Segment flush와 OS cache는 visibility와 durability를 분리한다

검색 엔진에서도 운영체제 경계는 중요합니다. Refresh는 새 segment를 검색 가능한 view에 열어 주는 가시성 경계입니다. Flush나 translog fsync는 crash 뒤에도 operation을 재생하거나 segment 상태를 복구할 수 있게 만드는 durability 경계입니다. 이 둘을 섞으면 "문서를 넣었는데 검색이 안 된다"와 "문서를 넣었는데 장애 뒤 사라졌다"를 같은 문제로 오해하게 됩니다.

```text
index request
  |
  v
in-memory indexing buffer
  |
  | refresh
  v
new segment is searchable
  |
  | merge / flush / translog fsync
  v
filesystem + OS page cache + block layer + storage
  |
  v
crash recovery can reconstruct durable operations
```

이 경로는 RDBMS의 WAL과 완전히 같은 구조는 아니지만, 면접에서 사용할 수 있는 공통 질문은 같습니다. "사용자에게 보이는가", "process crash 뒤 다시 만들 수 있는가", "node 장애 뒤 replica가 이어받을 수 있는가", "snapshot으로 과거 상태를 복원할 수 있는가"를 나누어 물어야 합니다. Segment merge는 I/O를 많이 쓰기 때문에 같은 서버에서 DB와 검색 엔진을 함께 돌리면 page cache와 block I/O queue를 두 시스템이 경쟁할 수 있습니다. 운영에서는 refresh interval, translog durability, merge throttling, disk watermark, snapshot repository 상태를 함께 확인해야 합니다.

상태를 네 칸으로 나누면 질문이 더 정확해집니다.

| 상태 질문 | 관련 경계 | 확인할 신호 |
| --- | --- | --- |
| 검색 결과에 보이는가 | refresh, PIT view | refresh interval, explicit refresh, PIT 사용 여부 |
| process crash 뒤 복구되는가 | translog, flush/fsync | translog durability, recovery log |
| node 장애 뒤 이어받는가 | primary/replica shard acknowledgement | replica health, shard allocation |
| 과거 시점으로 되돌릴 수 있는가 | snapshot repository | snapshot success, restore drill |

Replica와 snapshot을 분리해서 보는 이유도 여기에 있습니다. Replica는 현재 cluster 상태를 따라가고, snapshot은 특정 시점으로 되돌아갈 수 있는 별도 복구 표면입니다.

### Shard와 replica: 분산 검색은 fan-out/fan-in이다

Index는 primary shard 여러 개로 나뉩니다. Document id routing hash가 어느 primary shard에 들어갈지 정합니다. Replica shard는 primary shard의 복사본이고, 고가용성과 검색 부하 분산에 도움을 줍니다. Query가 들어오면 coordinating node가 관련 shard copy에 query를 보내고, 각 shard의 top N result를 받아 전체 top N으로 병합합니다.

Fan-out/fan-in은 작은 숫자 예제로 보면 비용이 더 잘 보입니다.

```text
index has 4 primary shards
query asks size=10, from=1000

each shard may need candidates for from+size:
  shard 0 -> top 1010
  shard 1 -> top 1010
  shard 2 -> top 1010
  shard 3 -> top 1010

coordinating node merges up to 4040 candidates
returns only 10 hits to the user
```

사용자는 10개를 요청했지만 cluster는 깊은 offset 앞부분을 버리기 위해 훨씬 많은 후보를 들고 있어야 합니다. 이 구조 때문에 shard 수, sort field, pagination 방식이 함께 성능을 좌우합니다.

이 구조는 세 가지 운영 질문을 만듭니다. 첫째, shard 수는 나중에 쉽게 줄이거나 늘리기 어렵습니다. Split/shrink 같은 기능이 있지만 조건과 비용이 있습니다. 둘째, query는 shard fan-out 비용을 갖습니다. 작은 index를 너무 많은 shard로 쪼개면 query마다 불필요한 overhead가 커집니다. 셋째, scoring은 shard-local statistics와 global ordering의 영향을 받습니다. DFS query then fetch 같은 방식은 더 정확한 term statistics를 얻을 수 있지만 비용이 커질 수 있습니다.

Replica는 backup이 아닙니다. Replica는 primary shard 장애 시 승격될 수 있는 cluster 내부 복제본입니다. 사용자가 실수로 delete query를 실행하면 delete도 replica에 복제됩니다. 장기 보존과 복구에는 snapshot이 필요합니다. "replica가 있으니 backup이 필요 없다"는 검색 엔진 운영에서 위험한 함정입니다.

### Query와 scoring: filter와 query context를 나눈다

Elasticsearch/OpenSearch query DSL에는 scoring이 필요한 query context와 score가 필요 없는 filter context가 있습니다. `match` query는 text relevance score를 계산합니다. `term` query는 exact term match에 가깝고, keyword field에 자주 씁니다. `range` query는 date/number range를 다룹니다. `bool` query는 must, should, filter, must_not을 조합합니다.

Filter context는 score 계산을 하지 않고 cache될 수 있어 structured filtering에 적합합니다. 예를 들어 `status = ACTIVE`, `tenant_id = t1`, `created_at >= ...` 같은 조건은 filter로 두고, 사용자의 text 검색어는 query context로 두는 식입니다. 모든 조건을 scoring query로 넣으면 관련도 계산이 불필요하게 흔들리고 비용이 늘 수 있습니다.

Scoring은 면접에서 깊게 들어갈 수 있는 주제입니다. BM25는 term frequency가 높을수록, 전체 corpus에서 드문 term일수록, field length가 짧을수록 점수를 높이는 경향이 있습니다. 하지만 business search에서는 단순 relevance만으로 충분하지 않습니다. 최신성, popularity, 판매 가능 여부, personalization, exact match boost, phrase match, typo tolerance, synonym을 조합합니다. 이때 scoring function이 설명 가능하고 실험 가능해야 합니다. 검색 품질은 unit test보다 query set과 평가 지표가 필요합니다.

### Pagination: from/size, search_after, PIT의 경계

`from`과 `size`는 앞 페이지를 건너뛰는 직관적인 pagination입니다. 하지만 shard마다 `from + size`만큼 candidate를 유지해야 하므로 깊은 페이지에서 비용이 커집니다. 예를 들어 100개 shard에서 10,000번째 페이지를 요청하면 각 shard가 많은 result를 정렬해 coordinating node로 보내야 합니다. 그래서 검색 엔진은 깊은 pagination에 기본 제한을 둡니다.

`search_after`는 이전 페이지의 마지막 sort value를 다음 query에 넘기는 방식입니다. Offset처럼 앞 결과를 건너뛰지 않고, 정렬 기준 다음부터 이어 갑니다. 안정적으로 쓰려면 sort field가 deterministic해야 하고, tie-breaker로 unique field를 함께 둡니다. 사용자가 임의의 500페이지로 바로 점프해야 하는 UX에는 맞지 않지만, infinite scroll이나 "다음 페이지"에는 적합합니다.

PIT는 point-in-time view를 열어 여러 search_after 요청이 같은 index 상태를 바라보게 합니다. Refresh와 merge가 일어나도 pagination 중 결과가 중복되거나 빠지는 문제를 줄입니다. 이때 다음 page 요청에는 직전 응답이 돌려준 최신 PIT id와 마지막 hit의 전체 `sort` 값을 함께 넘겨야 합니다. PIT id는 page마다 바뀔 수 있고, `search_after`는 sort 배열 전체를 기준으로 다음 위치를 찾기 때문입니다. 단, PIT는 cluster resource를 잡습니다. keep_alive를 짧게 두고, 사용이 끝나면 close API로 닫거나 만료되게 해야 합니다. Scroll API는 대량 export에는 쓰일 수 있지만 사용자-facing deep pagination에는 권장되지 않는 경우가 많습니다.

### Reindex와 alias cutover: 검색 index는 재생 가능한 파생물이어야 한다

Mapping을 바꾸거나 analyzer를 바꾸거나 document shape를 바꾸려면 새 index를 만드는 편이 안전합니다. 예를 들어 `products_v1`에서 `products_v2`를 만들고, source DB나 old index에서 reindex한 뒤, read alias를 `products_current`로 두고 cutover합니다. Write alias를 별도로 두면 새 write의 방향도 제어할 수 있습니다.

Alias cutover가 널리 쓰이는 이유는 검색 index를 운영 중 제자리에서 크게 고치는 일이 위험하기 때문입니다. 검색 field type이나 analyzer를 바꾸면 같은 문서라도 token과 sort 구조가 달라집니다. 운영 traffic이 계속 old mapping을 읽는 동안 새 mapping을 검증하려면, 새 index를 옆에 만들고 같은 query set을 비교한 뒤 읽기 alias를 짧게 돌리는 쪽이 안전합니다. 이 방식은 blue/green 배포와 비슷하지만, 코드가 아니라 검색 가능한 파생 데이터를 바꾼다는 점이 다릅니다.

```text
old path
  products_current -> products_v1
  writes -> products_write -> products_v1

build new path
  create products_v2 with new mapping
  backfill source of truth -> products_v2
  replay changes after backfill start
  compare query results and counts

cutover
  products_current -> products_v2
  products_write   -> products_v2
```

이 흐름에서 가장 위험한 부분은 backfill 중 들어온 변경과 삭제입니다. 새 index를 만들었다는 사실보다, source of truth에서 같은 순서로 다시 만들 수 있고 cutover 뒤 되돌릴 수 있는지가 더 중요합니다.

문제는 reindex 중에도 데이터가 바뀐다는 점입니다. Source of truth가 RDBMS라면 backfill 시작 시점 이후 변경분을 CDC나 outbox event로 new index에 반영해야 합니다. Old index에서 new index로 reindex하면 old index 자체가 stale할 수 있습니다. 삭제 event도 중요합니다. Insert/update만 반영하고 delete를 놓치면 new index에 ghost document가 남습니다.

Consistency 검증은 count만으로 부족합니다. Document count가 같아도 field 값이 다를 수 있고, analyzer tokenization 차이로 query 결과가 달라질 수 있습니다. 중요한 query set을 old/new index에 같이 던져 top result, total hit relation, sort order, aggregation 값을 비교해야 합니다. Alias cutover는 빠르게 되돌릴 수 있어야 하고, cutover 후 write path가 어느 index로 가는지 확인해야 합니다.

### Snapshot, dump, restore: backup과 export의 목적을 분리한다

Snapshot은 cluster의 index data와 metadata를 repository에 저장해 장애 복구나 migration에 쓰는 공식적인 backup 방식입니다. Replica와 다르게 삭제나 corruption이 cluster에 전파되더라도 이전 snapshot으로 복구할 수 있습니다. Snapshot repository의 접근 권한, encryption, retention, restore drill이 중요합니다. Snapshot을 한 번 설정하고 실제 restore를 해 보지 않으면 backup이 있다는 말이 증명되지 않습니다.

Dump/export는 다른 목적일 수 있습니다. 개발 환경 샘플 데이터, 작은 index 이동, data inspection에는 dump tool이 편할 수 있습니다. 하지만 dump 중 index가 계속 바뀌면 consistent snapshot을 보장하지 않을 수 있습니다. PIT나 scroll을 사용해 export view를 안정화할 수 있지만, source DB와 같은 시점 보장은 별도 문제입니다. 검색 index가 파생물이라면 가장 신뢰할 수 있는 복구는 source of truth에서 다시 reindex하는 것일 수 있습니다.

OpenSearch와 Elasticsearch는 API가 비슷한 부분이 많지만 버전과 plugin, 보안 기능, snapshot repository 지원, managed service 제약이 다를 수 있습니다. "둘 다 Lucene 기반"이라는 말은 출발점이지 운영 호환성 보장이 아닙니다. 실제 migration에서는 index settings, mappings, analyzers, ingest pipeline, security roles, snapshot compatibility를 확인해야 합니다.

### Firestore document modeling: query에서 거꾸로 설계한다

Firestore에서는 join이 없습니다. Query가 필요한 모양에서 document와 collection을 설계해야 합니다. 예를 들어 사용자의 주문 목록을 자주 본다면 `/users/{userId}/orders/{orderId}` subcollection이 자연스러울 수 있습니다. 전체 주문을 관리자 화면에서 검색해야 한다면 top-level `/orders/{orderId}` collection이나 collection group query가 필요할 수 있습니다. 같은 데이터를 여러 위치에 denormalize할 수 있지만, update consistency와 비용을 설계해야 합니다.

Firestore가 query에서 거꾸로 설계하게 만드는 배경에는 client SDK, security rules, managed index, 과금 모델이 함께 있습니다. RDBMS 서버 애플리케이션은 보통 backend가 SQL을 실행하고 결과를 가공한 뒤 client에 돌려줍니다. Firestore는 모바일이나 웹 client가 직접 document를 구독하거나 읽는 경로가 흔하고, 이때 security rules가 query 가능성과 권한을 동시에 판단합니다. Query가 rule을 만족할 수 없는 모양이면 "일단 많이 읽고 client에서 거른다"가 안전한 선택이 아닙니다. 읽은 document마다 비용이 생기고, 권한 경계도 DB 앞단에서 막혀야 합니다.

```text
RDBMS backend habit
  client -> API server
  server joins/orders/filters with SQL
  server returns shaped response

Firestore client-facing habit
  client -> collection query
  index must support query shape
  security rules must prove access
  billing follows document reads/writes/indexes
```

그래서 Firestore의 문서 모델은 ERD를 먼저 그리고 나중에 query를 붙이는 방식보다, 화면과 권한, 비용, hotspot을 먼저 놓고 collection/document path를 정하는 방식이 더 안전합니다.

Access pattern에서 거꾸로 설계한다는 말은 다음 질문을 순서대로 답한다는 뜻입니다.

```text
screen: 내 주문 목록 최신순
  query:
    collection = /users/{userId}/orders
    where state in (...)
    orderBy createdAt desc
    limit 20
  model:
    user별 subcollection에 order summary 저장
  cost/risk:
    주문 상세 변경 시 summary 동기화 필요

screen: 관리자 tenant 주문 검색
  query:
    collection = /orders
    where tenantId == ?
    where state == ?
    orderBy createdAt desc
  model:
    top-level orders collection과 composite index 필요
  cost/risk:
    tenant 조건 누락 시 security rules와 query가 충돌하거나 거부될 수 있음
```

두 화면은 같은 주문 데이터를 보지만 document 배치와 index 요구가 다릅니다. RDBMS라면 join과 secondary index로 늦게 보완할 수 있는 질문도, Firestore에서는 처음부터 query shape와 security rule을 함께 놓고 설계해야 합니다.

문서 크기와 write contention도 중요합니다. 한 user document 안에 모든 주문을 array로 넣으면 주문이 늘수록 문서가 커지고, 동시 update가 같은 document에 몰립니다. Firestore는 document 단위 write contention이 있으므로 high-frequency counter나 timeline을 단일 document에 몰아넣으면 hotspot이 생길 수 있습니다. 분산 counter나 time-bucketed document 같은 패턴을 고려합니다.

Index는 query 가능성을 결정합니다. Firestore는 single-field index를 자동으로 만들고, 복합 조건과 orderBy에는 composite index가 필요할 수 있습니다. Index는 읽기를 빠르게 하지만 write마다 index entry를 갱신해야 하므로 cost와 latency에 영향을 줍니다. 불필요한 index를 줄이고, 쿼리 경로를 실제 화면과 API에 맞춰 설계하는 것이 중요합니다.

### Firestore security rules, cost, hotspot

Firestore security rules는 client SDK가 직접 Firestore에 접근하는 구조에서 authorization 경계입니다. Rule은 request의 auth 정보, path 변수, resource data, request data를 사용해 read/write를 허용하거나 거부합니다. 하지만 rules는 query 결과를 필터링하는 사후 장치가 아닙니다. Query가 rule을 만족할 수 있음을 증명할 수 있어야 합니다. 예를 들어 tenant별 접근 rule이 있다면 query도 tenant 조건을 포함해야 합니다.

Admin SDK나 server-side privileged credential은 security rules를 우회할 수 있습니다. 그래서 서버 경로에서는 application authorization을 별도로 구현해야 합니다. "rules가 있으니 서버도 안전하다"가 아닙니다. Client path와 server path의 권한 경계를 분리하고 테스트해야 합니다.

비용은 document read/write/delete와 storage/index를 기준으로 생각합니다. Query가 100개 document를 읽으면 100 read 비용이 생깁니다. Listener는 변경이 발생할 때 read 비용을 만들 수 있습니다. Offset pagination은 건너뛴 document에도 비용이 생길 수 있으므로 cursor 기반 pagination이 중요합니다. 자주 바뀌는 field에 많은 index가 걸려 있으면 write amplification이 커집니다.

Rules와 query의 관계도 작은 반례로 볼 수 있습니다.

```text
rule:
  allow read if resource.data.tenantId == request.auth.token.tenantId

unsafe query:
  db.collection("orders").where("state", "==", "PAID")

safer query:
  db.collection("orders")
    .where("tenantId", "==", tokenTenantId)
    .where("state", "==", "PAID")
```

Security rules는 query 결과를 몰래 필터링해 주는 후처리기가 아닙니다. Query 자체가 rule을 만족할 수 있는 모양이어야 합니다. 이 경계를 모르면 개발 환경에서는 작은 데이터로 되는 것처럼 보이다가, 운영에서 권한 오류나 과한 읽기 비용, tenant leak 위험으로 이어질 수 있습니다.

Hotspot은 sequential document id, monotonically increasing indexed field, 단일 document counter, 특정 collection prefix에 write가 몰릴 때 생길 수 있습니다. Firestore는 자동 ID를 쓰면 분산에 유리합니다. 시간순 query가 필요하다고 timestamp를 document id prefix로 쓰면 write가 한쪽으로 몰릴 수 있습니다. 시간순 조회는 indexed timestamp field와 cursor로 처리하고, id는 분산성 있게 두는 편이 안전합니다.

## DBMS별 경계

여기서 DBMS는 넓은 의미의 data store입니다. 검색 엔진과 document DB를 같은 기준으로 비교하면 실수합니다.

| 주제 | Elasticsearch | OpenSearch | Firestore |
| --- | --- | --- | --- |
| 본질 | Lucene 기반 분산 검색/분석 엔진 | Elasticsearch에서 갈라진 Lucene 기반 검색/분석 엔진 | Google Cloud 기반 document database |
| 저장 단위 | JSON document가 index에 들어가고 `_source`와 index structure가 분리된다 | 유사하지만 버전, plugin, 보안/관리 기능 차이를 확인해야 한다 | collection/document path의 document |
| 주된 index | inverted index, doc values, BKD tree 등 field type별 구조 | Lucene 계열 구조를 공유하지만 API/feature compatibility 확인 필요 | single-field/composite index |
| text 처리 | analyzer, tokenizer, token filter, synonym, BM25 scoring | analyzer와 scoring 모델은 유사하되 plugin availability 확인 | full-text search 엔진이 아니며 별도 검색 연동이 필요할 수 있다 |
| transaction | single document/index operation 중심, refresh visibility 분리 | 유사 | document/transaction/batch 경계가 있고 query consistency 제약을 확인해야 한다 |
| pagination | from/size 제한, search_after, PIT | from/size 제한, search_after, PIT 지원 범위 확인 | cursor 기반 pagination; offset 비용 주의 |
| backup | snapshot repository가 기본 공식 수단 | snapshot repository와 compatibility 확인 | managed backup/export, point-in-time recovery 지원 범위 확인 |
| security | index/cluster roles, document/field security는 license/edition 확인 | security plugin, roles, tenants, TLS 설정 확인 | security rules, IAM, Admin SDK 우회 경계 |
| 비용 모델 | cluster resource, shard/replica/storage/compute | cluster resource, managed service 여부 | document read/write/delete, storage, index, network |
| 설계 시작점 | query relevance, mapping, shard lifecycle | query relevance, compatibility, 운영 plugin | access pattern, security rules, cost, hotspot |

Elasticsearch와 OpenSearch는 유사한 API를 공유하는 부분이 많지만, 장기적으로는 버전과 기능 차이가 벌어질 수 있습니다. 특히 security, alerting, index management, vector search, snapshot repository, client compatibility는 제품과 버전별로 확인해야 합니다. "OpenSearch는 Elasticsearch와 같다"는 말은 면접에서 안전하지 않습니다. 안전한 답은 "Lucene 기반 검색 엔진으로 핵심 개념은 많이 공유하지만, 운영 기능과 API compatibility는 버전별로 검증해야 한다"입니다.

Firestore transaction 함수는 재시도될 수 있고, read는 write보다 먼저 수행되어야 하며, offline client에서는 transaction이 실패합니다. 그래서 transaction callback 안에서 결제 승인, 이메일 발송, 외부 API 호출처럼 한 번만 일어나야 하는 side effect를 직접 실행하면 안 됩니다. DB transaction 안팎의 side effect 경계는 [애플리케이션 경계, 멱등성, outbox](12-application-boundaries-idempotency-money-outbox.md)에서 이어서 봅니다.

Firestore는 검색 엔진이 아닙니다. Prefix 검색, 형태소 검색, relevance ranking, typo tolerance, synonym search를 Firestore query만으로 만들려고 하면 한계가 큽니다. Firestore는 operational document store로 두고, full-text search는 Elasticsearch/OpenSearch/Algolia 같은 검색 시스템으로 projection하는 설계가 흔합니다. 이 경우 source of truth는 Firestore이고, search index는 파생 read model입니다. Sync lag, delete propagation, reindex, idempotent event handling이 필요합니다. 이 지연과 재생 가능성은 [복제, 지연, 백업, failover](09-replication-lag-backup-failover.md)의 replica lag 사고방식과도 닮아 있습니다.

표의 `inverted index`, `doc values`, `BKD tree`는 Lucene 계열 검색 엔진에서 field type과 사용 목적에 따라 달라지는 물리 구조입니다. `text` field는 analyzer가 문장을 token으로 쪼갠 뒤, token에서 document 목록으로 가는 inverted index를 만듭니다. 그래서 "환불 정책"을 검색할 때 `환불`, `정책` 같은 term으로 document 후보를 찾을 수 있습니다. `keyword` field는 분석하지 않은 정확한 값을 집계, filter, sort에 자주 쓰므로 columnar 형태의 doc values가 중요해집니다. `price`, `created_at`, `location`처럼 숫자, 시간, 위치 범위를 묻는 field는 BKD tree 같은 point/range 검색 구조가 관여합니다. 이 구분을 모르면 text에 정렬을 걸거나, keyword에 형태소 검색을 기대하거나, 숫자 range filter를 단순 inverted index 감각으로 설명하는 실수를 하게 됩니다.

RDBMS와도 경계가 다릅니다. RDBMS는 relational constraint, join, transaction, ad hoc query에 강합니다. Search engine은 text relevance와 large-scale retrieval에 강합니다. Firestore는 flexible document model, client SDK, realtime listener, managed scale에 강하지만 query shape와 cost 제약이 강합니다. 면접에서는 "어떤 것이 더 좋다"보다 "어떤 질문에 답하기 위해 어떤 저장 모델을 선택하는가"를 말해야 합니다.

## 직접 재생해 보기

아래 실험은 로컬 Elasticsearch/OpenSearch 또는 managed 개발용 cluster에서 작은 index로 수행합니다. 운영 cluster에서는 mapping 변경, delete, force merge, reindex를 직접 실행하지 않습니다.

### text와 keyword 차이 확인

```json
PUT search_lab
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "status": { "type": "keyword" },
      "created_at": { "type": "date" }
    }
  }
}
```

```json
POST search_lab/_doc/1
{
  "title": "Spring Transaction Rollback",
  "status": "ACTIVE",
  "created_at": "2026-05-20T00:00:00Z"
}
```

`match`와 `term` query를 비교합니다.

```json
GET search_lab/_search
{
  "query": {
    "match": { "title": "spring rollback" }
  }
}
```

```json
GET search_lab/_search
{
  "query": {
    "term": { "title.keyword": "Spring Transaction Rollback" }
  }
}
```

PASS 신호는 `text` field의 match가 analyzer를 거쳐 token 기반으로 찾고, `keyword` field의 term이 exact value에 가깝게 동작한다는 점을 확인하는 것입니다. FAIL 신호는 text field에 term query를 던지고 왜 결과가 없냐고 판단하는 것입니다.

### analyzer 결과 보기

```json
POST search_lab/_analyze
{
  "field": "title",
  "text": "Spring Transaction Rollback"
}
```

PASS 신호는 실제 token 목록을 보고 query와 index가 어떤 term으로 만나는지 이해하는 것입니다. 한국어 검색 품질을 다루는 경우에는 한국어 analyzer를 적용했을 때 token이 어떻게 달라지는지 비교해야 합니다.

### search_after와 PIT 흐름

```json
POST search_lab/_pit?keep_alive=1m
```

응답의 PIT id를 사용해 첫 페이지를 검색합니다.

```json
GET _search
{
  "size": 10,
  "pit": {
    "id": "PIT_ID",
    "keep_alive": "1m"
  },
  "sort": [
    { "created_at": "desc" },
    { "_shard_doc": "desc" }
  ],
  "query": {
    "term": { "status": "ACTIVE" }
  }
}
```

응답 마지막 hit의 `sort` 값을 다음 요청의 `search_after`에 넣습니다. PASS 신호는 offset 없이 다음 페이지로 이어지는 것입니다. FAIL 신호는 sort tie-breaker 없이 같은 timestamp 문서들이 중복되거나 빠지는 것입니다.

### alias 기반 reindex cutover

```json
PUT products_v1
PUT products_v2

POST _aliases
{
  "actions": [
    { "add": { "index": "products_v1", "alias": "products_read" } },
    { "add": { "index": "products_v1", "alias": "products_write" } }
  ]
}
```

Backfill 후 read alias를 바꿉니다.

```json
POST _aliases
{
  "actions": [
    { "remove": { "index": "products_v1", "alias": "products_read" } },
    { "add": { "index": "products_v2", "alias": "products_read" } }
  ]
}
```

PASS 신호는 application이 physical index name이 아니라 alias를 사용하고, rollback 시 alias를 되돌릴 수 있는 것입니다. FAIL 신호는 reindex 중 들어온 write와 delete를 new index에 반영하지 않아 cutover 후 결과가 달라지는 것입니다.

### Firestore document modeling 점검

주문 조회 화면을 기준으로 collection을 설계해 봅니다.

```text
/orders/{orderId}
  userId
  tenantId
  state
  amount
  createdAt

/users/{userId}/orders/{orderId}
  state
  amount
  createdAt
```

관리자는 tenant 전체 주문을 상태와 시간으로 검색하고, 사용자는 자기 주문 목록을 최신순으로 본다고 하자. PASS 신호는 각 query가 어떤 collection과 index를 타는지, 중복 저장된 summary가 언제 갱신되는지, delete가 양쪽에 어떻게 반영되는지 설명하는 것입니다. FAIL 신호는 "나중에 query로 join하면 된다"고 생각하는 것입니다.

### Firestore security rules 사고 실험

```text
allow read: if request.auth != null
            && resource.data.tenantId == request.auth.token.tenantId;
```

이 rule이 있다고 가정하면 query도 tenant 조건을 포함해야 합니다. PASS 신호는 client query가 `where("tenantId", "==", tokenTenantId)`를 포함하고, cross-tenant document를 읽으려 하면 거부되는 테스트를 작성하는 것입니다. FAIL 신호는 rule이 결과를 알아서 필터링해 줄 것이라고 믿는 것입니다.

## 면접 꼬리 질문

### Inverted index를 RDBMS B-tree와 비교하면 어떻게 설명하나요?

B-tree는 key order를 따라 range와 exact lookup을 빠르게 합니다. Inverted index는 term에서 그 term을 포함한 document 목록으로 갑니다. Full-text search에서는 문서를 훑어 단어를 찾는 대신, 미리 만들어 둔 term posting list를 조합합니다. 그래서 text search, relevance scoring, phrase query에는 inverted index가 적합하고, relational join이나 강한 transaction에는 RDBMS가 더 적합합니다.

### `text`와 `keyword` field 차이는 무엇인가요?

`text`는 analyzer를 거쳐 token으로 쪼개져 full-text search에 쓰입니다. `keyword`는 전체 값을 하나의 term처럼 다루어 exact match, sorting, aggregation에 적합합니다. 같은 문자열이라도 검색어 match가 필요한지, 정확한 값 필터와 정렬이 필요한지에 따라 multi-field로 둘 수 있습니다.

### Mapping을 잘못 잡으면 왜 reindex가 필요한가요?

이미 index된 field의 type이나 analyzer는 기존 inverted index 구조에 반영되어 있습니다. analyzer를 바꾸면 token이 달라지고, numeric을 keyword로 바꾸면 저장 구조와 query 방식이 달라집니다. 그래서 기존 index를 제자리에서 간단히 바꾸기보다 새 mapping을 가진 index를 만들고 source data를 다시 색인하는 경우가 많습니다.

### shard 수는 많을수록 좋은가요?

아닙니다. shard가 많으면 병렬성과 분산 여지는 늘지만, query fan-out, cluster metadata, file handle, heap overhead, merge 비용이 늘어납니다. shard가 너무 적으면 index가 커졌을 때 scale-out과 recovery가 어려울 수 있습니다. shard 수는 data size, query pattern, growth, recovery time, node size를 보고 정합니다.

### `from/size` pagination은 왜 깊은 페이지에서 문제가 되나요?

깊은 offset을 건너뛰려면 각 shard가 앞 결과까지 포함한 많은 candidate를 정렬해 coordinating node에 보내야 합니다. 사용자는 10개만 받지만 cluster는 훨씬 많은 작업을 합니다. `search_after`는 이전 페이지의 마지막 sort value 이후부터 이어가므로 deep pagination 비용을 줄일 수 있습니다. PIT를 함께 쓰면 pagination 중 index 변화로 인한 중복/누락을 줄입니다.

### Firestore에서 데이터 모델은 어떻게 시작하나요?

화면과 query에서 시작합니다. 어떤 collection을 어떤 조건과 orderBy로 읽을지, security rules가 그 query를 허용할 수 있는지, 필요한 composite index가 무엇인지, read/write 비용과 hotspot이 어디 생기는지 먼저 봅니다. RDBMS처럼 정규화한 뒤 join으로 해결하는 모델이 아닙니다.

### Firestore security rules가 있으면 server authorization은 필요 없나요?

필요합니다. Client SDK는 security rules의 보호를 받지만, Admin SDK나 privileged server credential은 rules를 우회할 수 있습니다. 서버 경로에서는 application authorization을 직접 수행해야 합니다. Rules와 IAM, server-side checks를 각각의 boundary로 봐야 합니다.

## 함정 질문

### "Elasticsearch에 넣었는데 바로 검색이 안 되면 저장 실패인가요?"

아닙니다. Index operation 성공과 search visibility는 refresh 경계 때문에 분리됩니다. Refresh interval이 지나야 검색 가능해질 수 있습니다. Get by id는 더 빨리 보일 수 있습니다. 저장 실패인지 refresh 지연인지 구분해야 합니다.

### "Replica가 있으니 snapshot backup은 없어도 되죠?"

아닙니다. Replica는 cluster 내부 복제본입니다. 잘못된 delete나 mapping 변경, application bug가 primary에 적용되면 replica에도 반영됩니다. Snapshot은 특정 시점으로 되돌아갈 수 있는 backup입니다. 실제 restore drill까지 해 봐야 backup을 믿을 수 있습니다.

### "OpenSearch는 Elasticsearch와 완전히 호환되죠?"

완전 호환이라고 단정하면 위험합니다. 핵심 개념과 많은 API가 유사하지만, 버전이 갈라지면서 security, plugin, API, client, snapshot compatibility가 달라질 수 있습니다. migration에서는 실제 사용 기능별 compatibility matrix와 실험이 필요합니다.

### "Firestore는 document DB니까 아무 JSON이나 넣고 나중에 query하면 되죠?"

아닙니다. Firestore query는 index와 제한이 있고, document size, collection path, composite index, security rules, read/write cost가 설계를 좌우합니다. 나중에 필요한 query를 만들 수 없거나 비용이 커질 수 있습니다. 먼저 access pattern을 정해야 합니다.

### "Offset pagination은 Firestore에서도 검색 엔진에서도 UI만의 문제 아닌가요?"

아닙니다. Offset은 backend 비용 문제입니다. 검색 엔진에서는 shard별 candidate 유지 비용이 커지고, Firestore에서는 건너뛴 document에도 비용이 생길 수 있습니다. Cursor나 search_after처럼 이전 위치를 기준으로 이어가는 방식을 우선 검토합니다.

### "검색 index를 primary DB로 쓰면 RDBMS가 필요 없나요?"

대부분의 업무 시스템에서는 위험합니다. 검색 엔진은 relevance search와 analytics read model에 강하지만, relational constraint, multi-row transaction, 강한 회계 불변식, write consistency에는 RDBMS가 더 적합한 경우가 많습니다. 검색 index는 source of truth에서 재생 가능한 projection으로 두는 편이 안전합니다.

## 더 깊게 볼 자료

- [Elastic Docs - Text analysis](https://www.elastic.co/docs/manage-data/data-store/text-analysis): analyzer, tokenizer, token filter의 공식 설명을 확인합니다.
- [Elastic Docs - Mapping](https://www.elastic.co/docs/manage-data/data-store/mapping): field type과 mapping contract를 확인합니다.
- [Elastic Docs - Clusters, nodes, and shards](https://www.elastic.co/docs/deploy-manage/distributed-architecture/clusters-nodes-shards): shard/replica와 cluster architecture를 확인합니다.
- [Elastic Docs - Paginate search results](https://www.elastic.co/docs/reference/elasticsearch/rest-apis/paginate-search-results): `from/size`, `search_after`, PIT 기반 pagination을 확인합니다.
- [Elastic Docs - Open a point in time API](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-open-point-in-time): PIT id 갱신, `keep_alive`, resource 경계를 확인합니다.
- [Elastic Docs - Migrate your data](https://www.elastic.co/docs/manage-data/migrate): reindex와 migration 접근을 확인합니다.
- [OpenSearch Docs - Field types](https://docs.opensearch.org/docs/latest/field-types/): OpenSearch mapping과 field type을 확인합니다.
- [OpenSearch Docs - Paginate results](https://docs.opensearch.org/docs/latest/search-plugins/searching-data/paginate/): OpenSearch pagination 지원 범위를 확인합니다.
- [OpenSearch Docs - Reindex data](https://docs.opensearch.org/docs/latest/api-reference/index-apis/reindex/): OpenSearch reindex API와 제약을 확인합니다.
- [OpenSearch Docs - Snapshots](https://docs.opensearch.org/docs/latest/tuning-your-cluster/availability-and-recovery/snapshots/index/): snapshot/restore 운영 경계를 확인합니다.
- [Firebase Docs - Cloud Firestore data model](https://firebase.google.com/docs/firestore/data-model): collection/document/subcollection 모델의 공식 설명입니다.
- [Firebase Docs - Transactions and batched writes](https://firebase.google.com/docs/firestore/manage-data/transactions): transaction 재시도, read-before-write, offline 실패 경계를 확인합니다.
- [Firebase Docs - Cloud Firestore security rules](https://firebase.google.com/docs/firestore/security/get-started): client access와 security rules 경계를 확인합니다.
- [Firebase Docs - Cloud Firestore index overview](https://firebase.google.com/docs/firestore/query-data/index-overview): single-field/composite index와 query 가능성을 확인합니다.
- [Firebase Docs - Cloud Firestore best practices](https://firebase.google.com/docs/firestore/best-practices): hotspot, document id, index fanout, latency 관련 권장사항을 확인합니다.
- [Firebase Docs - Cloud Firestore pricing](https://firebase.google.com/docs/firestore/pricing): read/write/delete, storage, network 비용 모델을 확인합니다.
