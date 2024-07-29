# [Sharded Timestamps](https://firebase.google.com/docs/firestore/solutions/shard-timestamp)

- [Sharded Timestamps](#sharded-timestamps)
    - [순차적으로 색인이 지정된 필드](#순차적으로-색인이-지정된-필드)
        - [데이터 모델](#데이터-모델)
        - [쿼리 예시](#쿼리-예시)
    - [timestamp 필드 샤딩](#timestamp-필드-샤딩)
        - [`shard` 필드 추가](#shard-필드-추가)
        - [색인 정의 업데이트](#색인-정의-업데이트)
            - [복합 색인 정의 삭제](#복합-색인-정의-삭제)
            - [단일 필드 색인 정의 업데이트](#단일-필드-색인-정의-업데이트)
        - [새 복합 색인 만들기](#새-복합-색인-만들기)
    - [순차적으로 색인이 지정된 필드의 쓰기 제한 이해](#순차적으로-색인이-지정된-필드의-쓰기-제한-이해)

## 순차적으로 색인이 지정된 필드

- '순차적으로 색인이 지정된 필드'(Sequential indexed fields)
    - 단조롭게 증가 또는 감소하는 색인이 지정된 필드가 포함된 문서 컬렉션을 의미
    - 대부분의 경우 timestamp 필드
- 다음과 같이 userid 값을 할당하는 경우 userid라는 색인 생성 필드가 있는 user 문서 컬렉션에 이 제한이 적용
    - 1281, 1282, 1283, 1284, 1285, ...
- 필드가 단조롭게 증가 또는 감소한다는 사실 외에 필드의 실제 값은 중요하지 않다.
    - timestamp 필드에 이 제한이 적용되지 않는 경우도 있다
    - timestamp 필드가 임의로 분산된 값을 추적하면 쓰기 제한이 적용되지 않는다.
- 예를 들어 다음과 같이 단조롭게 증가하는 필드 값 집합에는 모두 쓰기 제한이 적용됩니다.
    - 100000, 100001, 100002, 100003, ...
    - 0, 1, 2, 3, ...

### 데이터 모델

```json
// 통화, 보통주, ETF 같은 금융 상품을 거의 실시간으로 분석하는 앱 예제
// instruments/${ID}
{
    symbol: 'AAA',
    price: {
        currency: 'USD',
        micros: 34790000
    },
    exchange: 'EXCHG1',
    instrumentType: 'commonstock',
    timestamp: 1546350323010
}
```

- 통화, 보통주, ETF 같은 금융 상품을 거의 실시간으로 분석하는 앱 예제에서 `timestamp`는 `instruments` 컬렉션 내에서 단조롭게 증가
    - 2019-01-01T13:45:23.010Z
    - 2019-01-01T13:45:23.101Z
    - 2019-01-01T13:45:23.001Z
- 앱이 금융 상품에 대한 업데이트를 초당 1,000~1,500개 수신한다면?
    - 색인이 지정된 `timestamp` 필드가 있는 문서를 포함하는 컬렉션 `instruments`에 허용되는 초당 500회의 쓰기 제한을 훨씬 초과

```js
async function insertData() {
  const instruments = [
    {
      symbol: 'AAA',
      price: {
        currency: 'USD',
        micros: 34790000
      },
      exchange: 'EXCHG1',
      instrumentType: 'commonstock',
      timestamp: Timestamp.fromMillis(
          Date.parse('2019-01-01T13:45:23.010Z'))
    },
    {
      symbol: 'BBB',
      price: {
        currency: 'JPY',
        micros: 64272000000
      },
      exchange: 'EXCHG2',
      instrumentType: 'commonstock',
      timestamp: Timestamp.fromMillis(
          Date.parse('2019-01-01T13:45:23.101Z'))
    },
    {
      symbol: 'Index1 ETF',
      price: {
        currency: 'USD',
        micros: 473000000
      },
      exchange: 'EXCHG1',
      instrumentType: 'etf',
      timestamp: Timestamp.fromMillis(
          Date.parse('2019-01-01T13:45:23.001Z'))
    }
  ];

  const batch = fs.batch();
  for (const inst of instruments) {
    const ref = fs.collection('instruments').doc();
    batch.set(ref, inst);
  }

  await batch.commit();
}
```

### 쿼리 예시

## timestamp 필드 샤딩

### `shard` 필드 추가

```js
// Define our 'K' shard values
const shards = ['x', 'y', 'z'];
// Define a function to help 'chunk' our shards for use in queries.
// When using the 'in' query filter there is a max number of values that can be
// included in the value. If our number of shards is higher than that limit
// break down the shards into the fewest possible number of chunks.
function shardChunks() {
  const chunks = [];
  let start = 0;
  while (start < shards.length) {
    const elements = Math.min(MAX_IN_VALUES, shards.length - start);
    const end = start + elements;
    chunks.push(shards.slice(start, end));
    start = end;
  }
  return chunks;
}

// Add a convenience function to select a random shard
function randomShard() {
  return shards[Math.floor(Math.random() * Math.floor(shards.length))];
}

async function insertData() {
  const instruments = [
    {
      shard: randomShard(),  // add the new shard field to the document
      symbol: 'AAA',
      price: {
        currency: 'USD',
        micros: 34790000
      },
      exchange: 'EXCHG1',
      instrumentType: 'commonstock',
      timestamp: Timestamp.fromMillis(
          Date.parse('2019-01-01T13:45:23.010Z'))
    },
    {
      shard: randomShard(),  // add the new shard field to the document
      symbol: 'BBB',
      price: {
        currency: 'JPY',
        micros: 64272000000
      },
      exchange: 'EXCHG2',
      instrumentType: 'commonstock',
      timestamp: Timestamp.fromMillis(
          Date.parse('2019-01-01T13:45:23.101Z'))
    },
    {
      shard: randomShard(),  // add the new shard field to the document
      symbol: 'Index1 ETF',
      price: {
        currency: 'USD',
        micros: 473000000
      },
      exchange: 'EXCHG1',
      instrumentType: 'etf',
      timestamp: Timestamp.fromMillis(
          Date.parse('2019-01-01T13:45:23.001Z'))
    }
  ];

  const batch = fs.batch();
  for (const inst of instruments) {
    const ref = fs.collection('instruments').doc();
    batch.set(ref, inst);
  }

  await batch.commit();
}
```

### 색인 정의 업데이트

- 초당 500회의 쓰기 제한을 제거하려면 `timestamp` 필드를 사용하는 기존의 단일 필드 색인 및 복합 색인을 삭제

#### [복합 색인 정의 삭제](https://firebase.google.com/docs/firestore/solutions/shard-timestamp?hl=ko#update_index_definitions)

#### [단일 필드 색인 정의 업데이트](https://firebase.google.com/docs/firestore/solutions/shard-timestamp?hl=ko#update_single-field_index_definitions)

### [새 복합 색인 만들기](https://firebase.google.com/docs/firestore/solutions/shard-timestamp?hl=ko#create_new_composite_indexes)

- `timestamp`를 포함하는 인덱스는 반드시 `shard`도 포함해야 한다 예를 들어 위의 쿼리를 지원하려면 다음 색인을 추가해야 합니다.

## 순차적으로 색인이 지정된 필드의 쓰기 제한 이해

- 순차적으로 색인이 지정된 필드의 쓰기 속도 제한은 Cloud Firestore가
    - 어떻게 색인 값(index values)을 저장하고
    - 어떻게 색인 쓰기(index writes)를 확장(scale, 조정)하는지에 따라 좌우된다
- Cloud Firestore는 색인 쓰기별로 문서 이름과 각 색인이 지정된 필드의 값을 연결하는 키-값 항목을 정의한다
- Cloud Firestore는 이러한 색인 항목을 태블릿이라는 데이터 그룹으로 정리한다
    - 각 Cloud Firestore 서버에는 1개 이상의 태블릿이 있다
    - 특정 태블릿에 대한 쓰기 부하가 너무 높은 경우 Cloud Firestore는 태블릿을 더 작은 태블릿으로 분할하고 서로 다른 Cloud Firestore 서버에 새 태블릿을 분산하여 수평으로 확장한다
- Cloud Firestore는 사전순으로 가까운 색인 항목을 동일한 태블릿에 배치한다
    - `timestamp` 필드에서와 같이 **태블릿의 색인 값들이 서로 너무 가까이**에 있으면 Cloud Firestore에서 태블릿을 더 작은 **태블릿으로 효율적으로 분할하지 못한다**
    - 이로 인해 **단일 태블릿에서 너무 많은 트래픽을 수신하는 핫스팟**이 생성되고 핫스팟에 대한 읽기 및 쓰기 작업 속도가 더 느려진다
- `timestamp` 필드를 샤딩하면 Cloud Firestore에서 워크로드를 여러 태블릿에 효율적으로 분할할 수 있다
    - `timestamp` **필드의 값**은 여전히 서로 가깝게 유지
    - 하지만 **연결된 샤드 및 색인 값** 덕분에 Cloud Firestore가 색인 항목들 간에 충분한 공간을 확보하여 여러 태블릿에 항목들을 분할할 수 있다
