# K-HEXAGONAL-PORT-TYPES-AND-MAPPING-OWNERSHIP

## 1) 이 카드가 답하는 질문

Hexagonal + DDD로 시스템을 설계/리팩토링하다 보면 아래 질문이 반복됩니다.

"Repository(Port)의 메서드 시그니처에는 어떤 타입이 등장해야 합니까? DB 테이블(row)을 나타내는 타입은 core에 둬야 합니까, adapter에 둬야 합니까? 그리고 DB 컬럼에 JSON payload 같은 문자열 표현이 들어 있을 때, 그 문자열을 누가 파싱하여 core 모델로 변환하는 것이 정석에 가깝습니까?"

이 카드는 "좋아 보이는 구조"가 아니라, 의존성 방향(컴파일 가능성), 변경 비용(연쇄 수정), 실패 모드(언제/어떻게 깨지는지), 검증 경로(작은 실험)로 결론을 재현 가능하게 만드는 것을 목표로 합니다.

---

## 2) 배경: 왜 이 문제가 실무에서 비싸게 터지는가

경계(boundary)가 있는 구조에서 타입을 잘못 배치하면, 문제가 두 가지 형태로 비싸게 터집니다.

첫째, 의존성 방향이 깨지면 "바깥(인프라) 변경"이 "안쪽(core) 변경"으로 전이됩니다. DB 컬럼 이름 하나가 바뀌거나, JSON 직렬화 정책(필드명/누락/기본값)이 바뀌는 순간 core 코드가 같이 바뀌면, core는 더 이상 "보호된 안쪽"이 아니게 됩니다. 이 상태에서는 테스트도 함께 오염됩니다. core 단위 테스트가 DB row/JSON 형태를 알게 되면, 테스트는 도메인 계약이 아니라 인프라 표현을 재현하게 되고, 리팩토링이 곧바로 테스트 붕괴로 이어집니다.

둘째, 반대로 모든 변환을 core로 밀어 넣으면 core가 "표현(Representation)"을 소유하지 않은 상태에서 표현을 해석해야 합니다. 표현을 소유하지 않는데 해석만 하면, 변화가 생길 때 실패가 어디서 터지는지(관측)가 불명확해지고, "왜 그 지점에서" 실패하는지 설명이 닫히지 않습니다. 예를 들어 DB에 저장된 payload가 Jackson 기본 직렬화 결과처럼 "인프라 편의로 만들어진 스냅샷"이라면, core가 이를 직접 파싱하는 순간 core는 Jackson/클래스 구조/애노테이션 변화의 영향을 직접 받습니다.

따라서 이 문제는 "어디에 클래스를 두느냐"가 아니라, 시스템의 비용 구조(변경/장애/테스트)를 결정하는 핵심 제어점입니다.

---

## 3) 핵심 제어점(왜 하필 이 축인가)

이 주제는 범위가 넓기 때문에, 판단을 고정하는 제어점 3개로 축을 좁힙니다.

첫째는 **Port 시그니처의 타입 소유권**입니다. Port는 core가 외부에 "필요한 능력"을 선언하는 API이므로, Port에 등장하는 타입은 core가 소유해야 합니다. Port가 adapter의 타입을 참조하면 core는 adapter를 알아야 컴파일되며, Hexagonal의 의존성 방향이 즉시 붕괴합니다.

둘째는 **표현(Representation)의 소유권**입니다. DB row의 컬럼 집합, JSON payload의 스키마/버전 규칙을 누가 결정하고 장기 호환을 책임지는지가 포인트입니다. 표현이 "인프라가 내부 편의로 선택한 스냅샷"이면 해석 책임은 인프라(adapter)에 두는 편이 자연스럽습니다. 반대로 표현이 "도메인이 정의한 계약(타입/버전/필수 필드)"이면 core가 해석을 소유해도 경계가 흔들리지 않습니다.

셋째는 **정책(policy)과 저장소 primitive의 분리**입니다. 예를 들어 "멱등 실행(단일 실행)" 같은 것은 도메인/유스케이스 정책입니다. 하지만 그 정책이 필요로 하는 것은 DB 트랜잭션/유니크 제약/조건부 업데이트 같은 저장소 primitive입니다. 저장소 primitive는 adapter가 제공하되, 그 primitive를 어떤 순서로 조합해 불변식을 만족시키는지는 core가 소유해야 drift(미세한 순서 차이/누락)가 줄어듭니다.

---

## 4) 기본 결론(기본값)과 선택 이유

기본 결론은 다음처럼 닫습니다.

Port(Repository) 시그니처에는 **core가 소유한 타입만** 등장하게 하십시오. adapter 내부의 DB row 타입(JDBC row record, JPA entity 등)은 adapter에 숨기고, adapter가 row를 core 타입으로 매핑하여 Port를 구현하는 것이 기본값입니다. 이때 core 타입은 "도메인 엔티티"일 수도 있고, "Port DTO(저장소에서 읽은 스냅샷/원장 레코드)"일 수도 있습니다. 중요한 점은 core가 그 타입의 진화(필드/의미)를 통제할 수 있어야 한다는 것입니다.

JSON payload 파싱 책임은 "JSON이냐 아니냐"가 아니라, 그 payload가 **도메인 계약인지 인프라 스냅샷인지**로 결정하십시오.

- payload가 인프라 스냅샷(예: "객체를 Jackson 기본 규칙으로 덤프")이라면, 그 표현을 선택한 쪽이 adapter이므로 파싱/직렬화도 adapter에 두는 편이 안정적입니다. core는 "의미 있는 타입"만 받습니다.
- payload가 도메인 계약(예: type/version/envelope가 고정되고 하위 호환 규칙이 정의됨)이라면, core가 파싱을 소유해도 됩니다. 다만 이 경우에는 core가 스키마 버전, 필수 필드, 하위 호환 전략을 함께 소유해야 합니다.

마지막으로, 도메인 타입(예: `DepositTransferResult`)이 DB 레코드 타입(예: `TransactionRecord`)을 직접 받아 `fromRecord(...)` 같은 변환을 제공하는 형태는 기본값으로 피하십시오. 이는 도메인 타입이 저장 표현을 알게 만들어 결합을 증가시킵니다. 변환은 core의 "정책/오케스트레이션" 컴포넌트(예: replay/guard 서비스) 또는 adapter 매퍼에서 닫는 편이 안전합니다.

---

## 5) 핵심 용어 정의(이 문서 맥락에서)

**Core**는 업무 의미(도메인 모델, 유스케이스 정책)를 소유하는 안쪽입니다. core는 DB/JDBC/HTTP 같은 구체 기술을 몰라도 컴파일되고 테스트될 수 있어야 합니다.

**Adapter(인프라)**는 core가 선언한 Port를 구현하여 실제 DB/네트워크/프레임워크와 연결하는 바깥입니다. adapter는 구체 기술 변화(JDBC->JPA, MySQL->PostgreSQL, JSON 라이브러리 교체)를 흡수하는 위치입니다.

**Port(Repository Port)**는 core가 외부에 요구하는 "능력"의 선언입니다. 여기서 Port는 저장소라는 구현을 말하는 것이 아니라, core 입장에서 필요한 추상 연산(조회/삽입/조건부 갱신/스트리밍 등)을 말합니다.

**표현(Representation)**은 "DB row 컬럼 집합"이나 "JSON payload 문자열"처럼 저장/전송을 위해 선택된 형태입니다. 이 카드에서 중요한 것은 표현이 "도메인 계약"인지 "인프라 스냅샷"인지의 경계입니다.

**Port DTO**는 도메인 엔티티와 달리, Port 경계를 통과하기 위해 core가 정의한 데이터 구조입니다. 예를 들어 "원장 레코드"처럼 여러 유스케이스에서 조회되는 데이터를 담을 수 있습니다.

---

## 6) 메커니즘(입력 -> 변환 규칙 -> 산출물 -> 소비자)

Port 타입과 파싱 책임을 고정하려면, 데이터 흐름을 이렇게 닫아야 합니다.

입력은 DB가 반환하는 row입니다. row는 대개 문자열/숫자/타임스탬프처럼 단순 값으로 구성됩니다.

변환 규칙은 adapter에 있습니다. adapter는 row를 읽어 core가 소유한 타입(도메인 엔티티 또는 Port DTO)으로 매핑합니다. JSON payload가 "인프라 스냅샷"이면 여기서 역직렬화도 수행하고, core에는 의미 있는 타입으로 전달합니다.

산출물은 core 타입입니다. core의 유스케이스/정책 서비스는 산출물을 입력으로 받아, 정책(멱등/타임아웃/재시도/라우팅)을 적용한 결과를 만듭니다.

소비자는 inbound(HTTP/TCP/CLI) 또는 후속 유스케이스입니다. 이 소비자는 adapter의 row 타입이나 JSON 문자열 표현을 몰라도 동작해야 합니다.

동일한 원리를 "core가 파싱을 소유"하는 케이스로 바꾸면, adapter는 row를 Port DTO로만 매핑하고(문자열 payload 포함), core의 정책 서비스가 "도메인 계약 스키마"로 payload를 해석합니다. 이때 중요한 점은 payload 스키마의 소유권이 core로 넘어왔다는 사실까지 함께 닫혀야 한다는 것입니다.

---

## 7) 조건 변경 시 예측(무엇이 깨지고 어떻게 관측되는가)

### 7.1 Port가 adapter 타입을 노출하면 무엇이 깨지나

Port가 adapter 타입을 반환하면, core는 adapter를 알아야 컴파일됩니다. 멀티 모듈/패키지 경계에서 이는 보통 "컴파일 실패"로 관측됩니다. 컴파일이 통과하도록 억지로 의존성을 추가하면, 그 순간부터 DB row 타입 변경(JDBC 매핑 변경/JPA entity 리네임)이 core 변경으로 전파됩니다.

관측 포인트는 두 가지입니다.

첫째, core 모듈이 adapter 모듈을 의존해야만 빌드가 된다는 형태의 설정 변경이 발생합니다.

둘째, DB 컬럼 이름/타입 변경이 "도메인 로직 변경"으로 같이 나타납니다. 예를 들어 단순 컬럼 rename이 유스케이스 코드까지 수정하게 만들면, 경계가 무너진 것입니다.

### 7.2 core가 소유하지 않는 JSON 표현을 core가 파싱하면 무엇이 깨지나

표현을 소유하지 않으면서 파싱만 하면, 표현이 조금만 바뀌어도 core에서 런타임 예외가 터집니다. 인프라 스냅샷(JSON dump)의 경우, 필드명 변화/nullable 변화/기본값 변화는 보통 "배포 없이"도 발생할 수 있습니다. 예를 들어 adapter에서 직렬화 정책을 바꾸거나, 클래스 구조를 리팩토링하면 과거 데이터가 새 파서에 맞지 않게 됩니다.

관측 포인트는 저장된 payload replay 시점의 역직렬화 실패입니다. 실패 형태는 보통 "필수 필드 누락" 또는 "형식 불일치" 예외로 드러나며, 이 예외는 "과거 데이터"를 만나야만 발생하므로 테스트에서 놓치기 쉽습니다.

### 7.3 도메인 타입이 Port DTO/DB 레코드를 직접 알면 무엇이 깨지나

도메인 타입이 `fromRecord(PortDto)` 같은 변환을 직접 제공하면, 도메인 타입이 저장 표현 변화에 끌려갑니다. 이 경우 실패는 컴파일 단계(필드 rename)에서 관측되기도 하고, 더 위험하게는 런타임 단계(의미 변경, null 의미 변경)에서 관측되기도 합니다.

관측 포인트는 "도메인 타입이 DB 컬럼 이름을 알고 있다"는 코드 형태입니다. 도메인 타입의 생성/변환 코드가 `record.responsePayload()` 같은 표현을 직접 소비한다면, 저장 표현이 도메인으로 침투한 것입니다.

---

## 8) 코드 예시(나쁜 패턴 1 + 권장 패턴 1 이상)

### 8.1 Port가 adapter 타입을 노출하는 나쁜 패턴

```java
// core 모듈
package core;

import adapter.JdbcTransactionRow; // adapter 타입이 core로 들어옵니다.

public interface TransactionRepository {
    JdbcTransactionRow findById(String id);
}
```

이 설계는 "Port의 타입 소유권" 제어점을 깨뜨립니다. core는 adapter 없이는 컴파일될 수 없고, 의존성 방향이 역전됩니다.

### 8.2 도메인 타입이 저장 레코드를 직접 해석하는 나쁜 패턴

```java
package domain;

import core.TransactionRecord;

public record DepositTransferResult(String messageNo, String resultCode) {
    public static DepositTransferResult fromRecord(TransactionRecord record) {
        // 도메인 타입이 persistence 표현을 직접 소유합니다.
        // record의 필드/컬럼/표현이 바뀌면 도메인 타입이 함께 흔들립니다.
        return new DepositTransferResult(record.messageNo(), record.resultCode());
    }
}
```

### 8.3 adapter가 row를 core 타입으로 매핑하고, core 서비스가 정책을 담당하는 권장 패턴

```java
// core 모듈
package core;

import java.time.Instant;
import java.util.Optional;

public record TransactionRecord(
    String transactionId,
    String operation,
    String sendDate,
    String messageNo,
    String requestHash,
    String status,
    String responsePayload,
    Instant createdAt,
    Instant updatedAt
) {}

public interface TransactionRepository {
    Optional<TransactionRecord> findByMessageNo(String sendDate, String messageNo);
    void insertPending(TransactionRecord pending);
    void updateSuccess(String transactionId, String status, String responsePayload);
}

// core 모듈(정책 서비스)
package core;

import java.util.Objects;

public final class IdempotentExecutionService {
    private final TransactionRepository repo;

    public IdempotentExecutionService(TransactionRepository repo) {
        this.repo = Objects.requireNonNull(repo);
    }

    // 여기서 멱등/충돌/리플레이 정책을 닫습니다.
}

// adapter 모듈
package adapter;

import core.TransactionRecord;
import core.TransactionRepository;

import java.util.Optional;

public final class JdbcTransactionRepository implements TransactionRepository {
    @Override
    public Optional<TransactionRecord> findByMessageNo(String sendDate, String messageNo) {
        // ResultSet -> TransactionRecord 매핑
        return Optional.empty();
    }

    @Override
    public void insertPending(TransactionRecord pending) {
        // 유니크 제약으로 winner 선출
    }

    @Override
    public void updateSuccess(String transactionId, String status, String responsePayload) {
        // UPDATE
    }
}
```

핵심은 adapter가 "저장소 표현"을 흡수하고, core는 "정책"을 흡수해 각각의 변화 주기를 분리한다는 점입니다.

---

## 9) 최소 실험(검증)

아래 실험은 특정 프로젝트 명령에 의존하지 않고, 표준 `javac`로 관측 가능한 형태로 고정합니다. 하나는 실패를 관측하도록 설계합니다.

### 9.1 Port가 adapter 타입을 노출하면 core 단독 컴파일이 깨짐(실패 관측)

아래처럼 파일을 만들고 컴파일해 보십시오.

```java
// lab/core/PortBad.java
package core;

import adapter.JdbcTransactionRow;

public interface PortBad {
    JdbcTransactionRow find();
}
```

```bash
mkdir -p lab/core
cat > lab/core/PortBad.java <<'JAVA'
package core;

import adapter.JdbcTransactionRow;

public interface PortBad {
    JdbcTransactionRow find();
}
JAVA

javac lab/core/PortBad.java
```

기대 결과는 컴파일 실패이며, 에러는 "`package adapter does not exist`"처럼 관측됩니다. 즉 core는 adapter 없이 컴파일될 수 없다는 사실이 고정됩니다.

### 9.2 Port 타입을 core 소유 타입으로 바꾸면 core 단독 컴파일이 가능해짐(성공 관측)

```java
// lab/core/TransactionRecord.java
package core;

public record TransactionRecord(String id) {}
```

```java
// lab/core/PortGood.java
package core;

public interface PortGood {
    TransactionRecord find();
}
```

```bash
cat > lab/core/TransactionRecord.java <<'JAVA'
package core;

public record TransactionRecord(String id) {}
JAVA

cat > lab/core/PortGood.java <<'JAVA'
package core;

public interface PortGood {
    TransactionRecord find();
}
JAVA

javac lab/core/*.java
```

기대 결과는 컴파일 성공입니다. 즉 Port 타입 소유권을 core로 가져오면, 의존성 방향이 컴파일 레벨에서 강제됩니다.

---

## 10) 트레이드오프/대안/Variants

첫째 트레이드오프는 **adapter 파싱(표현 흡수)** vs **core 파싱(도메인 계약 소유)** 입니다.

- adapter 파싱은 core의 이식성(표현/라이브러리 결합 최소화)을 강화합니다. 대신 adapter 구현이 조금 더 두꺼워지고, "정책"과 "표현"이 섞이지 않도록 내부 구조(매퍼 분리)가 필요합니다.
- core 파싱은 adapter를 얇게 만들 수 있지만, core가 표현 스키마와 하위 호환을 소유해야 합니다. 이 전제가 없으면 과거 데이터 replay에서 런타임 실패가 발생합니다.

둘째 트레이드오프는 **Port DTO를 넓게(테이블 1:1)** vs **Port DTO를 좁게(필요 필드만)** 입니다.

- 1:1 DTO는 구현이 단순하고 조회/운영 요구가 많을 때 유리하지만, 스키마 변화가 core 계약 변화로 전이될 수 있습니다.
- 좁은 DTO는 core 결합을 줄이지만, 새로운 조회 요구가 생길 때 DTO 확장이 필요하고, 매핑 코드가 조금 더 복잡해질 수 있습니다.

---

## 11) 흔한 오해

첫째 오해는 "DDD면 Repository는 항상 도메인 엔티티만 반환해야 한다"입니다. 실제로는 Port의 목적이 "도메인 저장"인지 "원장/검색"인지에 따라, Port DTO(스냅샷/레코드)를 반환하는 설계가 더 안전할 수 있습니다. 중요한 것은 타입이 core에 의해 소유되고, 도메인 정책이 표현에 끌려가지 않게 경계를 닫는 것입니다.

둘째 오해는 "JSON이면 무조건 core가 파싱하면 안 된다"입니다. JSON 자체가 금지되는 것이 아니라, 그 JSON 스키마를 누가 소유하는지가 핵심입니다. 도메인이 type/version/필수 필드/호환 규칙을 정의했다면 core가 파싱을 소유해도 경계는 유지됩니다.

---

## 12) 팀 적용 질문(선택)

이 결정을 팀 규칙으로 적용할 때는 아래 질문 순서로 좁히는 편이 안전합니다.

1. Port가 반환/입력으로 받는 타입은 "누가 소유"합니까? core가 소유하지 못하면 Port 시그니처에 넣지 않습니다.
2. 저장된 payload(문자열/JSON)의 스키마/버전 규칙을 "누가" 정의하고 장기 호환을 책임집니까?
3. 정책(멱등/리스/재시도)은 core에 있고, 저장소 primitive(조건부 insert/update)는 adapter에 있는가요? 순서/불변식이 drift 하지 않게 골격이 고정돼 있나요?
