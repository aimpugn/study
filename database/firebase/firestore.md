# Firestore

## [데이터 구조화](https://firebase.google.com/docs/firestore/manage-data/structure-data?hl=ko)

- 옵션
    - 문서
    - 여러 컬렉션
    - 문서 내의 하위 컬렉션

### 문서의 중첩 데이터

- 문서 내에 복합 객체를 중첩 가능
    - 배열
    - 지도 등
- 장점
    - 문서 안에 **단순한 고정 데이터 목록**을 보관하려는 경우 데이터 구조를 손쉽게 설정하고 간소화할 수 있다
- 제한사항
    - 시간에 따라 데이터가 증가하는 경우 다른 옵션보다 확장성이 부족하다
    - 목록이 커지면 문서도 커지므로 문서 검색 속도가 느려질 수 있다
- 가능한 사용 사례
    - 예를 들어 채팅 앱에서 사용자가 가장 최근에 입장한 대화방 3개를 프로필에 중첩 목록으로 저장할 수 있다

```text
📘 alovelace
    name :
        first : "Ada"
        last : "Lovelace"
    born : 1815
    📑 rooms :
        0 : "Software Chat"
        1 : "Famous Figures"
        2 : "Famous SWEs"
```

### 하위 컬렉션

```text
📑 science
    📘 software
        name : "software chat"
        📑 users
            📘 alovelace
                first : "Ada"
                last : "Lovelace"
            📘 sride
                first : "Sally"
                last : "Ride"`


    📘 astrophysics
        name : "astrophysics chat"
...
```

- 데이터가 **시간에 따라 증가할 가능성**이 있다면 문서 내에 컬렉션을 만들 수 있다
- 장점
    - 목록이 커져도 상위 문서의 크기는 그대로이다
    - 또한 하위 컬렉션에서 모든 쿼리 기능을 사용할 수 있고
    - 하위 컬렉션 간에 컬렉션 그룹 쿼리를 실행할 수 있다
- 제한사항
    - 하위 컬렉션을 손쉽게 삭제할 수 없다
- 가능한 사용 사례
    - 동일한 채팅 앱에서 채팅방 문서 안에 사용자 또는 메시지의 컬렉션을 만들 수 있다

### 루트 수준 컬렉션

```text
📑 users
    📘 alovelace
        first : "Ada"
        last : "Lovelace"
        born : 1815
    📘 sride
        first : "Sally"
        last : "Ride"
        born : 1951
📑 rooms
    📘 software
        📑 messages
            📘 message1
                from : "alovelace"
                content : "..."
            📘 message2
                from : "sride"
                content : "..."
```

- 데이터베이스 루트 수준에 컬렉션을 만들어 상이한 데이터 세트를 정리한다
- 장점
    - 루트 수준 컬렉션은 다대다 관계에 적합하며 각 컬렉션 내에서 강력한 쿼리를 제공한다
- 제한사항
    - 데이터베이스가 커지면 내재적으로 계층 구조를 가진 데이터 가져오기가 더욱 복잡해질 수 있다
- 가능한 사용 사례
    - 동일한 채팅 앱에서 사용자 컬렉션 하나와 채팅방 및 메시지 컬렉션 하나를 만들 수 있다

## [데이터 추가](https://firebase.google.com/docs/firestore/manage-data/add-data?hl=ko#dart)

```dart
// 인스턴스 초기화
db = FirebaseFirestore.instance;

// `set()` 메서드 사용
final city = <String, String>{
  "name": "Los Angeles",
  "state": "CA",
  "country": "USA"
};

db
    .collection("cities")
    .doc("LA")
    .set(city)
    .onError((e, _) => print("Error writing document: $e"));
```

- 문서가
    - 없으면? 생성
    - 있으면? 새로 제공한 데이터로 내용을 덮어쓴다

```dart
// Update one field, creating the document if it does not already exist.
final data = {"capital": true};

db.collection("cities").doc("BJ").set(data, SetOptions(merge: true));
```

- 단, 기존 문서와 병합(`SetOptions(merge: true)`)하도록 지정한 경우는 덮어쓰지 않는다
- 문서가 있는지 확실하지 않은 경우 전체 문서를 실수로 덮어쓰지 않도록 새 데이터를 기존 문서와 병합하는 옵션을 전달
- 맵이 포함된 문서의 경우, 빈 맵을 포함한 필드와 함께 필드 집합을 지정하면, 대상 문서의 맵 필드를 덮어쓴다

### 데이터 타입

```dart
final docData = {
  "stringExample": "Hello world!",
  "booleanExample": true,
  "numberExample": 3.14159265,
  "dateExample": Timestamp.now(),
  "listExample": [1, 2, 3],
  "nullExample": null
};

final nestedData = {
  "a": 5,
  "b": true,
};

docData["objectExample"] = nestedData;

db
    .collection("data")
    .doc("one")
    .set(docData)
    .onError((e, _) => print("Error writing document: $e"));
```

- Cloud Firestore를 사용하면 문서 안에 다양한 데이터 타입 사용 가능
    - 문자열
    - 불리언
    - 숫자
    - 날짜
    - null
    - 중첩 배열
    - 객체 등
- Cloud Firestore는 코드에서 사용하는 숫자 유형과 관계없이 숫자를 항상 double 부동 소수점으로 저장

### 커스텀 객체

```dart
class City {
  final String? name;
  final String? state;
  final String? country;
  final bool? capital;
  final int? population;
  final List<String>? regions;

  City({
    this.name,
    this.state,
    this.country,
    this.capital,
    this.population,
    this.regions,
  });

  factory City.fromFirestore(
    DocumentSnapshot<Map<String, dynamic>> snapshot,
    SnapshotOptions? options,
  ) {
    final data = snapshot.data();
    return City(
      name: data?['name'],
      state: data?['state'],
      country: data?['country'],
      capital: data?['capital'],
      population: data?['population'],
      regions:
          data?['regions'] is Iterable ? List.from(data?['regions']) : null,
    );
  }

  Map<String, dynamic> toFirestore() {
    return {
      if (name != null) "name": name,
      if (state != null) "state": state,
      if (country != null) "country": country,
      if (capital != null) "capital": capital,
      if (population != null) "population": population,
      if (regions != null) "regions": regions,
    };
  }
}

final city = City(
  name: "Los Angeles",
  state: "CA",
  country: "USA",
  capital: false,
  population: 5000000,
  regions: ["west_coast", "socal"],
);
final docRef = db
    .collection("cities")
    .withConverter(
      fromFirestore: City.fromFirestore,
      toFirestore: (City city, options) => city.toFirestore(),
    )
    .doc("LA");
await docRef.set(city);
```

- Map 또는 Dictionary 객체로는 문서를 표현하기가 불편한 경우가 많으므로 커스텀 클래스를 사용하여 문서를 작성하는 방식도 지원
- Cloud Firestore에서 객체를 지원되는 데이터 유형으로 변환한다

## [문서 추가](https://firebase.google.com/docs/firestore/manage-data/add-data?hl=ko#add_a_document)

- `set()`:
    - 아이디 지정 필요
    - 데이터 저장
- `add()`:
    - 아이디 자동 생성
    - 데이터 저장
- `doc()`:
    - 아이디 자동 생성
- `.add(...)` === `.doc().set(...)`

> Note:
>
> Firebase 실시간 데이터베이스의 '푸시 ID'와 달리, Cloud Firestore에서 자동으로 생성한 ID에서는 자동 정렬을 지원하지 않습니다.
> 생성일에 따라 문서를 정렬하려면 타임스탬프를 문서의 필드로 저장해야 합니다.

```dart
// set()을 사용하여 문서를 만들 때는 만들 문서의 ID를 지정해야 한다
db.collection("cities").doc("new-city-id").set({"name": "Chicago"});


// 자동으로 ID를 생성하도록 하려면 `add()` 메서드를 호출
final data = {"name": "Tokyo", "country": "Japan"};

db.collection("cities").add(data).then((documentSnapshot) =>
    print("Added Data with ID: ${documentSnapshot.id}"));


// `doc()` 사용하여 자동 생성 ID를 만들고, 그 ID를 사용하여 문서 참조를 만든 후 사용
// `.add(...)` === `.doc().set(...)`
final data = <String, dynamic>{};

final newCityRef = db.collection("cities").doc();

// Later...
newCityRef.set(data);
```

## 문서 업데이트

```dart
final washingtonRef = db.collection("cites").doc("DC");

// 문서의 일부 필드를 업데이트하려면 `update()` 메서드를 사용
washingtonRef.update({"capital": true}).then(
    (value) => print("DocumentSnapshot successfully updated!"),
    onError: (e) => print("Error updating document $e"));
```

### 서버 타임스탬프

```dart
final docRef = db.collection("objects").doc("some-id");

// 문서의 필드를 서버 업데이트 수신 시점을 추적하는 서버 타임스탬프로 설정할 수 있다
// 트랜잭션 안의 여러 타임스탬프 필드를 업데이트할 때 각 필드는 동일한 서버 타임스탬프 값을 수신
final updates = <String, dynamic>{
  "timestamp": FieldValue.serverTimestamp(),
};

docRef.update(updates).then(
    (value) => print("DocumentSnapshot successfully updated!"),
    onError: (e) => print("Error updating document $e"));
```

### 중첩된 객체의 필드 업데이트

> Note
>
> 문서에서 중첩 필드를 업데이트하려는 경우 사용 가능한 옵션 간 미묘한 시맨틱 차이에 유의해야 합니다.

```dart
// 문서에 중첩된 객체가 있으면 `점(.) 표기법`을 사용하여 문서 내 중첩 필드를 참조 가능
// Assume the document contains:
// {
//   name: "Frank",
//   favorites: { food: "Pizza", color: "Blue", subject: "recess" }
//   age: 12
// }
// `favorites.color`는 Blue에서 Red로 업데이트 된다
db
    .collection("users")
    .doc("frank")
    .update({"age": 13, "favorites.color": "Red"});
```

```dart
// Create our initial doc
db.collection("users").doc("frank").set({
  name: "Frank",
  favorites: {
    food: "Pizza",
    color: "Blue",
    subject: "Recess"
  },
  age: 12
}).then(function() {
  console.log("Frank created");
});

// 점 표기법 없이 중첩 객체 필드를 업데이트 하면, 전체 맵 필드를 덮어쓴다
db.collection("users").doc("frank").update({
  favorites: {
    food: "Ice Cream"
    // `color`, `subject`가 사라진다
  }
}).then(function() {
  console.log("Frank food updated");
});

/*
Ending State, favorite.color and favorite.subject are no longer present:
/users
    /frank
        {
            name: "Frank",
            favorites: {
                food: "Ice Cream",
            },
            age: 12
        }
*/
```

### 배열 요소 업데이트

```dart
final washingtonRef = db.collection("cities").doc("DC");

// `arrayUnion`: 배열에 없는 요소만 추가
washingtonRef.update({
  "regions": FieldValue.arrayUnion(["greater_virginia"]),
});

// `arrayRemove`: 제공된 각 요소의 모든 인스턴스를 삭제
washingtonRef.update({
  "regions": FieldValue.arrayRemove(["east_coast"]),
});
```

### 숫자 증가

> Note:
>
> 필드가 없거나 현재 필드 값이 숫자가 아니면 작업을 통해 필드가 지정된 값으로 설정됩니다.
>
> 증분 작업은 카운터를 구현하는 데 유용하지만, 초당 한 번만 단일 문서를 업데이트할 수 있다
> 카운터를 이보다 자주 업데이트해야 하는 경우 [분산 카운터 페이지](https://firebase.google.com/docs/firestore/solutions/counters?hl=ko)를 참조

```dart
var washingtonRef = db.collection('cities').doc('DC');

// Atomically increment the population of the city by 50.
washingtonRef.update(
  {"population": FieldValue.increment(50)},
);
```

#### [솔루션: 분산 카운터](https://firebase.google.com/docs/firestore/solutions/counters?hl=ko#solution_distributed_counters)

- 각 카운터는 `shards`라는 하위 컬렉션을 갖는 문서
- 카운터의 값은 샤드들의 값의 합계
- 쓰기 처리량은 샤드의 수에 따라 선형적으로 증가. 10개의 샤드가 있는 분산 카운터는 기존 카운터보다 10배 많은 쓰기 처리 가능

```json
// 분산 카운터의 구조
// counters/${ID}
{
  "num_shards": NUM_SHARDS, // 전체 샤드의 개수
  "shards": [subcollection]
}

// counters/${ID}/shards/${NUM}
{
  "count": 123
}
```

```dart
// `ref`? db.collection('counters').doc()
function createCounter(ref, num_shards) {
    var batch = db.batch();

    // Initialize the counter document
    /*
     * {
     *   "num_shards": NUM_SHARDS, // 전체 샤드의 개수
     *   "shards": [subcollection]
     * }
     */
    batch.set(ref, { num_shards: num_shards });

    // Initialize each shard with count=0
    for (let i = 0; i < num_shards; i++) {
        // 숫자는 부동소수점 double로 처리되므로, "0", "1", ... "N" 형식의 문자열로 만든다
        const shardRef = ref.collection('shards').doc(i.toString());
        batch.set(shardRef, { count: 0 });
    }

    // Commit the write batch
    return batch.commit();
}

function incrementCounter(ref, num_shards) {
    // Select a shard of the counter at random
    const shard_id = Math.floor(Math.random() * num_shards).toString();
    const shard_ref = ref.collection('shards').doc(shard_id);

    // Update count
    return shard_ref.update("count", firebase.firestore.FieldValue.increment(1));
}

function getCount(ref) {
    // Sum the count of each shard in the subcollection
    return ref.collection('shards').get().then((snapshot) => {
        let total_count = 0;
        snapshot.forEach((doc) => {
            total_count += doc.data().count;
        });

        return total_count;
    });
}

var app = firebase.initializeApp(
    {apiKey: "...", authDomain: "...", projectId: "..."}, 
    "solution-counters",
);
db = firebase.firestore(app);

// Create a counter with 10 shards
createCounter(db.collection('counters').doc(), 10);

// Create a counter, then increment it
createCounter(db.collection('counters').doc(), 10).then(() => {
    return incrementCounter(ref, 10);
});

// Create a counter, increment it, then get the count
return createCounter(db.collection('counters').doc(), 10).then(() => {
    return incrementCounter(ref, 10);
}).then(() => {
    return getCount(ref);
});
```

#### 제한사항

위 솔루션은 Cloud Firestore에서 공유 카운터를 만드는 확장 가능한 방법이지만 다음과 같은 제한사항에 유의해야 합니다.

- 샤드 수
    - 샤드 수는 분산 카운터의 성능을 좌우한다
    - 샤드가 너무 적으면 일부 트랜잭션을 재시도해야 하므로 쓰기 작업이 느려진다.
    - 샤드가 너무 많으면 읽기가 느려지고 비용이 증가한다
    - 더 느린 주기로 업데이트되는 개별 롤업 문서에서 카운터 합계를 유지하고 클라이언트가 이 문서를 읽어 합계를 가져오도록 하면 읽기 비용을 상쇄할 수 있다
    - 하지만 이 경우 클라이언트가 업데이트 직후 모든 샤드를 읽어 합계를 계산하지 않고 롤업 문서가 업데이트될 때까지 기다려야 한다는 단점이 있다
- 비용
    - 카운터 값을 읽을 때 전체 샤드 하위 컬렉션이 로드되어야 하므로 **샤드 수에 따라 비용이 선형으로 증가**한다
