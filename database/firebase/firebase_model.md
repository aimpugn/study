# Firebase Model

- [Firebase Model](#firebase-model)
    - [Cloud Firestore 또는 실시간 데이터베이스](#cloud-firestore-또는-실시간-데이터베이스)
        - [데이터 구조와 쿼리](#데이터-구조와-쿼리)
        - [실시간 업데이트](#실시간-업데이트)
        - [성능 및 확장성](#성능-및-확장성)
        - [쿼리 기능](#쿼리-기능)
        - [비용](#비용)
    - [Cloud Firestore](#cloud-firestore)
        - [데이터 모델](#데이터-모델)
        - [데이터 구조 선택](#데이터-구조-선택)
        - [`collection` 내 하위 `collection`](#collection-내-하위-collection)
    - [실시간 데이터베이스 구조화](#실시간-데이터베이스-구조화)
        - [데이터를 구조화하는 방법: JSON 트리](#데이터를-구조화하는-방법-json-트리)
        - [데이터 구조 권장사항](#데이터-구조-권장사항)
            - [데이터 중첩 피하기](#데이터-중첩-피하기)
                - [Bad](#bad)
                - [Good](#good)
                - [Create data that scales](#create-data-that-scales)
    - [데이터 검색(Retrieving Data)](#데이터-검색retrieving-data)
        - [Flutter - Read data](#flutter---read-data)
    - [Firestore에서 컬렉션과 서브컬렉션 설계](#firestore에서-컬렉션과-서브컬렉션-설계)

## Cloud Firestore 또는 실시간 데이터베이스

### 데이터 구조와 쿼리

- Cloud Firestore는
    - 문서 기반의 데이터 모델을 사용
    - 데이터는 문서와 컬렉션으로 구성되며,
        - 문서는 key-value 쌍으로 데이터를 저장하고,
        - 컬렉션은 문서의 그룹을 나타냅니다.
    - 또한, Cloud Firestore는 문서 내에 중첩된 데이터를 가질 수 있으며, 데이터 구조화 및 쿼리 기능이 뛰어납니다.
    - 주어진 경로에서 문서를 읽을 때 해당 문서의 데이터만 로드되고, 자식 경로의 데이터는 필요할 때마다 수동으로 로드해야 한다
- Firebase Realtime Database는
    - 큰 JSON 트리로 데이터를 저장하며,
    - 간단한 데이터를 처리하는 데 적합하지만,
    - 대량의 데이터 또는 계층적 데이터를 처리하는 데는 한계가 있습니다​
    - 특정 리소스 경로에서 데이터를 로드하면 해당 경로의 모든 자식 리소스가 포함된다

### 실시간 업데이트

- Cloud Firestore
    - 실시간 리스너를 제공하여 데이터 변경을 실시간으로 모니터링할 수 있습니다.
- Realtime Database
    - 실시간으로 데이터의 변경을 반영할 수 있는 강력한 기능을 제공합니다.
    - 데이터 변경 시 클라이언트에 즉시 반영되며, 이는 실시간 채팅 앱과 같은 애플리케이션에 유용합니다.

### 성능 및 확장성

- Cloud Firestore
    - 글로벌 확장성을 제공하며, 대량의 데이터와 사용자를 처리할 수 있는 높은 성능을 제공합니다.
- Realtime Database
    - 단일 리전에서의 확장성을 제공하며, 데이터 크기와 동시 연결 사용자 수가 증가함에 따라 성능이 저하될 수 있습니다.

### 쿼리 기능

- Cloud Firestore
    - 강력한 쿼리 기능을 제공하여 복잡한 쿼리와 정렬, 필터링 등의 작업을 수행할 수 있습니다.
    - 여러 필드를 사용하여 쿼리를 생성하는 것이 쉽고, 여러 필드를 기반으로 쿼리를 필터링할 수 있습니다
- Realtime Database
    - 쿼리 기능이 제한적이며, 복잡한 쿼리를 수행하기 어렵습니다.
    - 여러 필드에 걸쳐 쿼리를 생성하려면 데이터를 비정규화해야 합니다

### 비용

- Cloud Firestore:
    - 읽기, 쓰기 및 삭제 작업에 대한 비용이 발생
    - 저장된 데이터의 양에 따라 비용이 발생합니다.
- Realtime Database:
    - 다운로드된 데이터의 양에 따라 비용이 발생
    - 데이터 저장에도 비용이 발생합니다.

## Cloud Firestore

### [데이터 모델](https://firebase.google.com/docs/firestore/data-model?authuser=0&hl=ko)

- 테이블이나 행이 없으며, `collection`으로 정리되는 문서에 데이터를 저장
- 각 문서에는 키-값 쌍이 들어 있다. Cloud Firestore는 **작은 문서**로 이루어진 **대규모 컬렉션**을 저장하는 데 최적화되어 있다
- 모든 문서는 `collection`에 저장되어야 한다. 문서는 다음 두 가지 포함 가능
    - 하위 `collection`
    - 중첩된 객체
- 하위 `collection` 및 중첩된 객체 둘 다 문자열 같은 기본 필드나 목록 같은 복합 객체를 포함할 수 있다
- `collection`과 문서는 Cloud Firestore에서 암시적으로 생성
- 사용자는 컬렉션 내 문서에 데이터를 할당하기만 하면 된다. 컬렉션이나 문서가 없으면 Cloud Firestore에서 만든다

### [데이터 구조 선택](https://firebase.google.com/docs/firestore/manage-data/structure-data?hl=ko&authuser=0)

### `collection` 내 하위 `collection`

```text
(collections)rooms
    - (class) roomA
    name : "my chat room"
        - (collections) messages
            - (class) message1
            from : "alex"
            msg : "Hello World!"

            - (class) message2
            from : "rody"
            msg : "How are you?"

    ...........

    - (class) roomB
    name : "private chat room"
        - (collections) messages
```

```dart
final messageRef = db
    .collection("rooms")
    .doc("roomA")
    .collection("messages")
    .doc("message1");
```

- **`collection`과 `document`(class)가 교대로 나타나는 패턴**이고, **`collection`과 `document`는 항상 이 패턴을 따라야** 한다
- 가령 아래와 같은 케이스는 참조할 수 없다
    - `collection`에 속한 `collection`
    - `document`에 속한 `document`는
- 🔺 문서를 삭제해도 하위 컬렉션은 삭제되지 않는다. 가령, `coll/doc` 문서가 더 이상 존재하지 않더라도 c`oll/doc/subcoll/subdoc`에는 문서가 있을 수 있다

## [실시간 데이터베이스 구조화](https://firebase.google.com/docs/database/admin/structure-data?hl=ko)

- Firebase 실시간 데이터베이스의 JSON 데이터 구조화와 관련된
    - 주요 데이터 아키텍처 개념과
    - 몇 가지 권장사항을 설명
- **최대한 쉽게 데이터를 저장하고 이후에 검색할 수 있도록 만들 방법을 계획**하는 것이 중요

### 데이터를 구조화하는 방법: JSON 트리

- 데이터베이스를 클라우드 호스팅 **JSON 트리**라고 생각하면 된다
- SQL 데이터베이스와 달리 테이블이나 레코드가 없으며, JSON 트리에 추가된 데이터는 **연결된 키**를 갖는 기존 JSON 구조의 노드가 된다
- 키의 생성은?
    - 사용자 ID 직접 생성
    - 또는 의미 있는 이름을 직접 키로 지정
    - `push()` 메서드를 사용하여 자동으로 지정도 가능

```json
// path: `/users/$uid`
{
  "users": {
    "alovelace": {
      "name": "Ada Lovelace",
      "contacts": { "ghopper": true },
    },
    "ghopper": { ... },
    "eclarke": { ... }
  }
}
```

### 데이터 구조 권장사항

#### 데이터 중첩 피하기

- 최대 32단계의 데이터 중첩을 허용하므로 중첩을 기본 구조로 도입해도 괜찮다고 생각할 수도 있음
- 하지만 데이터베이스의 특정 위치에서 데이터를 가져올 때 그 자식 노드까지 모두 가져오게 된다
- 누군가에게 read/write 권한을 줄때 해당 노드 하위의 자식 노드들에 대해서까지 권한을 부여하게 된다
- 따라서 데이터 구조는 가능한 평면화(flat) 하는 게 좋다

##### Bad

```json
{
    // This is a poorly nested data architecture, because iterating the children
    // of the "chats" node to get a list of conversation titles requires
    // potentially downloading hundreds of megabytes of messages
    "chats": {
        "one": {
            "title": "Historical Tech Pioneers",
            "messages": {
                "m1": { "sender": "ghopper", "message": "Relay malfunction found. Cause: moth." },
                "m2": { ... },
                // a very long list of messages
            }
        },
        "two": { ... }
    }
}
```

- 대화 제목 목록을 가져오기 위해 `chats` 노드의 자식 노드들을 순회할 때 모든 회원과 `messages` 등 수백메가 바이트의 데이터를 다운로드 하게 된다

##### Good

- `chats`는 채팅의 유니크한 아이디(`one`, `two`, `three`) 하위에 각 대화에 대한 메타 정보만 저장

    ```json
    {
        "chats": {
            "one": {
            "title": "Historical Tech Pioneers",
            "lastMessage": "ghopper: Relay malfunction found. Cause: moth.",
            "timestamp": 1459361875666
            },
            "two": { ... },
            "three": { ... }
    },
    ```

- `chats`의 `members`는 채팅의 유니크한 아이디(`one`, `two`, `three`)로 쉽게 저장되고 접근할 수 있다

    ```json
        "members": {
            "one": {
            "ghopper": true,
            "alovelace": true,
            "eclarke": true
            },
            "two": { ... },
            "three": { ... }
        },
    ```

- `messages`는 빠르게 순회하되 쉽게 페이지네이션 되고 조회되길 원하는 원하는 데이터와 분리하되, 채팅의 유니크한 아이디(`one`, `two`, `three`)로 조직화 되어야 한다

    ```json
        "messages": {
            "one": {
            "m1": {
                "name": "eclarke",
                "message": "The relay seems to be malfunctioning.",
                "timestamp": 1459361875337
            },
            "m2": { ... },
            "m3": { ... }
            },
            "two": { ... },
            "three": { ... }
        }
    }
    ```

- 위와 같이 구조화되면 대화당 몇 바이트만 다운로드하고 대화 목록을 순회할 수 있다
- 그리고 빠르게 메타 데이터 목록을 가져와서 UI에 대화 목록 리스팅하거나 보여줄 수 있다
- `messages`는 개별적으로 가져와서 도착하는대로 보여줄 수 있고, UI 반응 속도가 빨라진다

##### Create data that scales

- 목록의 부분 집합만 다운받는 게 보통 더 낫다. 특히 목록이 수천 개의 레코드를 포함할 때 일반적이다.
    - 이 관계가 **정적**(static)이고 **일방향**(one-directional)이면, 자식 객체들을 부모 하위에 중첩하기만 하면 된다.
- 관계가 동적이거나 비정규화(denormalize)를 해줘야 할 필요가 있을 수 있다.
    - 대체적으로 쿼리를 사용하여 데이터의 일부를 [검색](https://firebase.google.com/docs/database/admin/retrieve-data)하면 데이터를 비정규화할 수 있다
- 하지만 `users`와 `groups` 같은 양방향 관계에서 사용자가 어떤 그룹에 속해 있는지 판단하려면 복잡해질 수 있다.
    - 특정 사용자가 속하는 그룹을 나열하고 해당 그룹의 데이터만 가져오는 깔끔한 방법이 필요
    - `groups`의 인덱스가 큰 도움이 될 수 있다

```json
// Ada Lovelace의 멤버쉽을 추적하기 위한 인덱스
{
  "users": {
    "alovelace": {
      "name": "Ada Lovelace",
      // Ada Lovelace 프로필에 그룹을 색인
      "groups": {
         // 여기서 값은 중요하지 않으며 키가 존재한다는 것만 중요합니다.
         "techpioneers": true,
         "womentechmakers": true
      }
    },
    ...
  },
  "groups": {
    "techpioneers": {
      "name": "Historical Tech Pioneers",
      "members": {
        "alovelace": true,
        "ghopper": true,
        "eclarke": true
      }
    },
    ...
  }
}
```

- 위와 같이 Ada Lovelace의 레코드와 `groups` 모두에 관계를 저장하면 일부 데이터가 중복된다
- 현재 `alovelace`는 `groups` 아래에 색인이 생성되어 있고 `techpioneers`는 Ada Lovelace의 프로필에 나열된다
- 따라서 Ada Lovelace를 그룹에서 삭제하려면 프로필과 `groups` 두 위치에서 업데이트가 이루어져야 한다
- **중복성은 양방향 관계에서 불가피**하다.
- `users` 또는 `groups` 목록이 수백만 개로 늘어나거나 실시간 데이터베이스 보안 규칙으로 인해 일부 레코드에 액세스할 수 없더라도 이 중복성 덕분에 Ada Lovelace의 소속 그룹을 빠르고 효율적으로 확인할 수 있다

## 데이터 검색(Retrieving Data)

- 어떻게 데이터가 정렬되고, 간단한 데이터 쿼리를 수행하는 방법을 커버
- Admin SDK의 데이터 검색은 프로그래밍 언어마다 약간 다르게 구현

1. 비동기 리스너
    - Firebase 실시간 데이터베이스에 저장된 데이터는 데이터베이스 참조에 비동기 리스너를 연결하여 검색
    - 리스너는 데이터의 초기 상태가 확인될 때 한 번 트리거된 후 데이터가 변경될 때마다 다시 트리거된다
    - 이벤트 리스너는 여러 가지 다양한 이벤트 유형을 수신할 수 있다.
    - Java, Node.js, 그리고 Python Admin SDKs에서 지원
2. 블로킹 읽기
    - Firebase 실시간 데이터베이스에 저장된 데이터는 데이터베이스 참조의 블로킹 메소드 호출을 통해 검색
    - 이 메소드는 참조에 저장된 데이터를 반환한다.
    - 각 메소드 호출은 일회성 작업이다. 즉, SDK가 후속 데이터 업데이트를 감지(listen)하는 콜백을 등록하지 않는다는 의미
    - 이 데이터 검색 모델은 Python과 Go Admin SDK에서 지원된다

### [Flutter - Read data](https://firebase.google.com/docs/database/flutter/read-and-write#read_data)

- 어떤 경로의 데이터를 읽고 변경 사항을 감지(listen)하려면 `DatabaseReference`의 `onValue` 속성을 사용해서 `DatabaseEvent`를 감지

## Firestore에서 컬렉션과 서브컬렉션 설계

Firestore에서 컬렉션과 서브컬렉션을 설계할 때, 데이터 구조와 이름 짓기는 애플리케이션의 기능과 확장성에 큰 영향을 미칩니다. 여기서 중요한 것은 데이터의 접근성, 쿼리의 효율성, 그리고 확장 가능성입니다.

1. **컬렉션 설계 고려사항**:
   - **접근 패턴**: 데이터를 어떻게 쿼리하고, 어떤 연산이 자주 발생하는지 고려해야 합니다.
   - **보안 및 권한 관리**: 사용자 데이터에 대한 접근을 어떻게 관리할지 고려해야 합니다.
   - **확장성과 유연성**: 애플리케이션의 변화에 따라 데이터 구조가 유연하게 대응할 수 있어야 합니다.

2. **컬렉션 이름 짓기**:
   - **명확성과 간결성**: 컬렉션의 이름은 그 기능을 명확하게 설명하면서도 가능한 간결해야 합니다.
   - **동어 반복 피하기**: 예를 들어, `users/users`와 같은 중복은 피해야 합니다.

3. **설계 예시**:
   - **users 컬렉션**: 모든 사용자의 상세 정보를 포함합니다. 각 문서 ID는 사용자에게 할당된 고유 ID입니다.
   - **usernames 서브컬렉션**: 사용자의 고유한 아이디를 체크하기 위한 목적으로 사용됩니다. 이는 `users/$auto_generated_id/usernames`처럼 구성할 수 있습니다.
   - **phones 서브컬렉션**: 전화번호 체크를 위한 목적으로 사용됩니다. `users/$auto_generated_id/phones`와 같이 구성할 수 있습니다.

4. **경로 구조**:
   - Firestore는 `collection/document/collection/document` 형태를 지원합니다. 따라서 `users/$auto_generated_id/phones/$phone_id`와 같은 경로가 가능합니다.

5. **구조화 제안**:
   - **users**: 사용자의 상세 정보를 저장합니다.
   - **identifiers**: 아이디, 전화번호 등을 체크하기 위한 별도의 컬렉션을 만들 수 있습니다. 예를 들어, `identifiers/usernames`와 `identifiers/phones`와 같이 구성할 수 있습니다. 이렇게 하면 `users` 컬렉션과는 별개로 아이디와 전화번호의 유니크성을 관리할 수 있습니다.

이러한 접근 방식은 데이터 구조를 더욱 명확하고 관리하기 쉽게 만들어줄 수 있습니다. Firestore의 구조는 매우 유연하기 때문에, 애플리케이션의 요구 사항과 사용 패턴에 맞게 데이터 구조를 설계하는 것이 중요합니다.
