# Document NoSQL modeling

## Firestore/document modeling, consistency, security rules, cost

이 절은 Firestore/document modeling을 RDBMS normalization 감각 그대로 옮기지 않도록 설명한다. Firebase 공식 문서는 Cloud Firestore가 SQL database처럼 table과 row를 쓰지 않고, collection으로 정리되는 document에 key-value pair를 저장한다고 설명한다. 또한 큰 collection의 작은 document 저장에 최적화되어 있다고 설명한다. 이 한 문장이 modeling의 출발점이다. document database에서는 join을 나중에 걸겠다는 감각보다, 화면과 권한과 query와 비용이 함께 읽는 단위를 먼저 정해야 한다.

Firestore의 기본 모양은 collection/document/collection/document가 번갈아 나오는 path다. 로컬 seed `database/firebase/firebase_model.md`와 `database/firebase/firestore.md`는 rooms/messages 예시와 nested data, subcollection, root collection 선택지를 정리한다. 여기서 중요한 것은 어떤 구조가 더 예쁜지가 아니라, 어떤 read가 어떤 document를 몇 개 읽고, 어떤 security rule이 어떤 query를 허용하며, 어떤 write가 hot spot을 만드는지다.

```text
rooms/{roomId}
  name: "Software Chat"
  lastMessage: "..."

rooms/{roomId}/messages/{messageId}
  from: "alice"
  text: "Hello"
  createdAt: 2026-05-19T10:00:00Z

users/{userId}/roomSummaries/{roomId}
  roomName: "Software Chat"
  unreadCount: 3
  lastMessageAt: 2026-05-19T10:00:00Z
```

위 구조는 RDBMS라면 rooms, messages, memberships를 join해서 만들 수 있는 view를 document로 나누어 저장한다. 중복이 생긴다. 하지만 이 중복은 무조건 나쁜 것이 아니다. mobile client가 자주 여는 inbox 화면에서 room summary만 읽으면 된다면, summary document는 읽기 비용과 latency를 줄인다. 반대로 모든 필드를 한 document에 중첩하면 문서 크기, 업데이트 충돌, security rule, 비용이 커질 수 있다.

### 세 가지 구조 선택지

| 구조 | 좋은 경우 | 깨지는 경우 | 검증 질문 |
|---|---|---|---|
| nested data in document | 작고 고정된 목록, 부모와 항상 함께 읽는 값 | 시간이 지나며 목록이 계속 커짐 | 이 배열이 6개월 뒤에도 작은가 |
| subcollection | 부모 아래에서 계속 늘어나는 child, messages/events | 하위 컬렉션 삭제/집계가 번거로움 | parent 삭제 때 child lifecycle을 어떻게 닫는가 |
| root-level collection | 다대다 관계, 독립 query, collection group | path와 권한이 분산됨 | 화면별 read count와 rule 조건이 명확한가 |

Firestore security rules는 query 결과를 사후 필터링하는 장치가 아니다. 공식 문서는 security rules are not filters라고 설명한다. query가 현재 사용자가 읽으면 안 되는 document를 반환할 가능성이 있으면 요청은 실패한다. 따라서 data model과 rule과 query는 함께 설계해야 한다. 예를 들어 `rooms` 전체를 query한 뒤 client에서 member만 필터링하는 방식은 보안 모델이 아니다. query 자체가 `members.{uid} == true` 같은 rule과 맞거나, 사용자의 room summary collection을 따로 둬야 한다.

```text
잘못된 사고:
  클라이언트가 rooms 전체를 가져온다.
  앱에서 내가 속한 room만 필터링한다.
  security rule은 나중에 막아 줄 것이다.

더 안전한 사고:
  users/{uid}/roomSummaries를 읽는다.
  또는 rooms query가 membership 조건을 query 조건으로 포함한다.
  security rule은 query가 반환할 수 있는 전체 후보가 허용 범위인지 판단한다.
```

비용도 modeling의 일부다. Firestore pricing 문서는 query를 만족하기 위해 읽은 document와 index entry, write/delete 작업에 과금이 발생한다고 설명한다. `한 화면을 만들기 위해 몇 document를 읽는가`, `listener가 갱신 때 몇 번 다시 읽히는가`, `security rule에서 get/exists가 얼마나 발생하는가`, `index entry read가 붙는 query인가`를 설계 때 봐야 한다. RDBMS에서는 join 비용을 DB 서버가 먹고 애플리케이션은 결과 row만 보는 경우가 많지만, Firestore에서는 document read 수가 비용과 latency의 직접 언어가 된다.

### sequential index hot spot

로컬 seed `database/firebase/firestore_sharded_timestamps.md`는 timestamp처럼 단조롭게 증가하는 indexed field가 hot spot을 만들 수 있고, shard field를 추가해 workload를 나눌 수 있다고 정리한다. 이 문제는 DB 내부 저장 구조를 모르면 이상하게 보인다. 시간이 증가하는 값이 계속 비슷한 key range에 들어가면 같은 storage partition에 write가 몰릴 수 있다.

```text
without shard
  index key: timestamp=10:00:00 -> tablet A
  index key: timestamp=10:00:01 -> tablet A
  index key: timestamp=10:00:02 -> tablet A
  결과: 최신 timestamp 근처로 write 집중

with shard
  index key: shard=x, timestamp=10:00:00 -> tablet A
  index key: shard=y, timestamp=10:00:01 -> tablet B
  index key: shard=z, timestamp=10:00:02 -> tablet C
  결과: write가 여러 key range로 분산
```

이 trace는 Firestore modeling이 단순 JSON 설계가 아니라 storage/index/write distribution 설계라는 것을 보여 준다.

### 검증 예시

```text
모델링 검증 checklist
  1. 핵심 화면 5개의 document read 수를 손으로 계산한다.
  2. 각 query가 security rule 조건과 같은 방향인지 확인한다.
  3. parent delete 후 subcollection cleanup 경로를 적는다.
  4. timestamp/index hot spot 후보를 찾는다.
  5. 중복 저장 필드마다 update fan-out과 보정 job을 적는다.

PASS: 화면, rule, cost, consistency, cleanup이 같은 model에서 설명된다.
FAIL: JSON 구조는 그럴듯하지만 query와 rule이 서로 맞지 않거나 비용을 계산할 수 없다.
```

### 등장 배경 요약

문서형 모델링이 등장한 배경은 모바일과 웹 클라이언트가 작은 단위의 상태를 빠르게 읽고 동기화해야 하는 요구가 커졌기 때문이다. Firestore는 table join보다 document path, collection, security rules, realtime listener, per-document billing을 중심으로 사고하게 만든다. 그래서 모델링은 JSON을 예쁘게 배치하는 일이 아니라 화면별 read shape, 권한 조건, 비용, hot spot, 중복 전파를 함께 고정하는 작업이다.

### 공식 근거와 로컬 seed

- Firestore Data model: https://firebase.google.com/docs/firestore/data-model
- Firestore Structure data: https://firebase.google.com/docs/firestore/manage-data/structure-data
- Firestore Security Rules conditions: https://firebase.google.com/docs/firestore/security/rules-conditions
- Firestore Pricing: https://firebase.google.com/docs/firestore/pricing
- local seeds: `database/firebase/*`

이 절은 document modeling을 데이터 모양, 보안 규칙, 비용, hot spot, consistency를 함께 결정하는 작업으로 재구성한다.

### Firestore 모델링 replay drill 1

Firestore/document modeling, consistency, security rules, cost의 replay drill 1은 document size boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 document size boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 1의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 1에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 2

Firestore/document modeling, consistency, security rules, cost의 replay drill 2은 read shape first 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 read shape first 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 2의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 2에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 3

Firestore/document modeling, consistency, security rules, cost의 replay drill 3은 denormalization receipt 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 denormalization receipt 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 3의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 3에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 4

Firestore/document modeling, consistency, security rules, cost의 replay drill 4은 security rules as query contract 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 security rules as query contract 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 4의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 4에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 5

Firestore/document modeling, consistency, security rules, cost의 replay drill 5은 rule document access cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 rule document access cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 5의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 5에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 6

Firestore/document modeling, consistency, security rules, cost의 replay drill 6은 subcollection lifecycle 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 subcollection lifecycle 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 6의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 6에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 7

Firestore/document modeling, consistency, security rules, cost의 replay drill 7은 collection group query 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 collection group query 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 7의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 7에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 8

Firestore/document modeling, consistency, security rules, cost의 replay drill 8은 hot timestamp 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 hot timestamp 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 8의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 8에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 9

Firestore/document modeling, consistency, security rules, cost의 replay drill 9은 transaction boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transaction boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 9의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 9에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 10

Firestore/document modeling, consistency, security rules, cost의 replay drill 10은 listener cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 listener cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 10의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 10에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 11

Firestore/document modeling, consistency, security rules, cost의 replay drill 11은 index entry cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 index entry cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 11의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 11에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 12

Firestore/document modeling, consistency, security rules, cost의 replay drill 12은 source of truth 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 source of truth 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 12의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 12에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 13

Firestore/document modeling, consistency, security rules, cost의 replay drill 13은 document size boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 document size boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 13의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 13에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 14

Firestore/document modeling, consistency, security rules, cost의 replay drill 14은 read shape first 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 read shape first 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 14의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 14에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 15

Firestore/document modeling, consistency, security rules, cost의 replay drill 15은 denormalization receipt 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 denormalization receipt 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 15의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 15에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 16

Firestore/document modeling, consistency, security rules, cost의 replay drill 16은 security rules as query contract 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 security rules as query contract 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 16의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 16에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 17

Firestore/document modeling, consistency, security rules, cost의 replay drill 17은 rule document access cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 rule document access cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 17의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 17에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 18

Firestore/document modeling, consistency, security rules, cost의 replay drill 18은 subcollection lifecycle 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 subcollection lifecycle 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 18의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 18에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 19

Firestore/document modeling, consistency, security rules, cost의 replay drill 19은 collection group query 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 collection group query 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 19의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 19에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 20

Firestore/document modeling, consistency, security rules, cost의 replay drill 20은 hot timestamp 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 hot timestamp 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 20의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 20에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 21

Firestore/document modeling, consistency, security rules, cost의 replay drill 21은 transaction boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transaction boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 21의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 21에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 22

Firestore/document modeling, consistency, security rules, cost의 replay drill 22은 listener cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 listener cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 22의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 22에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 23

Firestore/document modeling, consistency, security rules, cost의 replay drill 23은 index entry cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 index entry cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 23의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 23에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 24

Firestore/document modeling, consistency, security rules, cost의 replay drill 24은 source of truth 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 source of truth 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 24의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 24에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 25

Firestore/document modeling, consistency, security rules, cost의 replay drill 25은 document size boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 document size boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 25의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 25에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 26

Firestore/document modeling, consistency, security rules, cost의 replay drill 26은 read shape first 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 read shape first 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 26의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 26에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 27

Firestore/document modeling, consistency, security rules, cost의 replay drill 27은 denormalization receipt 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 denormalization receipt 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 27의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 27에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 28

Firestore/document modeling, consistency, security rules, cost의 replay drill 28은 security rules as query contract 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 security rules as query contract 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 28의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 28에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 29

Firestore/document modeling, consistency, security rules, cost의 replay drill 29은 rule document access cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 rule document access cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 29의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 29에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 30

Firestore/document modeling, consistency, security rules, cost의 replay drill 30은 subcollection lifecycle 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 subcollection lifecycle 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 30의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 30에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 31

Firestore/document modeling, consistency, security rules, cost의 replay drill 31은 collection group query 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 collection group query 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 31의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 31에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 32

Firestore/document modeling, consistency, security rules, cost의 replay drill 32은 hot timestamp 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 hot timestamp 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 32의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 32에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 33

Firestore/document modeling, consistency, security rules, cost의 replay drill 33은 transaction boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transaction boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 33의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 33에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 34

Firestore/document modeling, consistency, security rules, cost의 replay drill 34은 listener cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 listener cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 34의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 34에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 35

Firestore/document modeling, consistency, security rules, cost의 replay drill 35은 index entry cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 index entry cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 35의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 35에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 36

Firestore/document modeling, consistency, security rules, cost의 replay drill 36은 source of truth 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 source of truth 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 36의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 36에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 37

Firestore/document modeling, consistency, security rules, cost의 replay drill 37은 document size boundary 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 document size boundary 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 37의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 37에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### Firestore 모델링 replay drill 38

Firestore/document modeling, consistency, security rules, cost의 replay drill 38은 read shape first 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 read shape first 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 38의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

Firestore/document modeling, consistency, security rules, cost의 replay drill 38에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.
