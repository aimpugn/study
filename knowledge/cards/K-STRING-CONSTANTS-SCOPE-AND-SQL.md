# K-STRING-CONSTANTS-SCOPE-AND-SQL

## 1. 이 카드가 답하는 질문
실무에서 다음 질문은 계속 다시 등장합니다.

"코드에 흩어진 문자열 리터럴을 어디까지 상수/enum으로 올려야 하고, 어디부터는 인라인으로 두는 것이 더 낫습니까? 특히 SQL 문자열은 가독성을 해치지 않으면서도 안전하게 유지하려면 어떻게 해야 합니까?"

이 카드는 `"2번 이상 쓰이면 상수"` 같은 기계 규칙을 목표로 하지 않습니다.
대신 문자열이 코드에서 맡는 역할을 "의미 단위"와 "표현 단위"로 나누어, 실수 비용이 큰 곳을 우선 통제하는 기준을 제공합니다.

---

## 2. 적용 범위와 비범위
이 카드의 결론은 아래 같은 문자열에 바로 적용됩니다.

- 설정 키(YAML/JSON), 프로토콜 필드명(TCP 헤더/전문 키), HTTP 헤더명, CLI 옵션명처럼 "식별자"로 쓰이는 문자열
- 정책 키/모드/프로파일처럼 분기 로직을 바꾸는 문자열
- SQL에서 `:namedParameter` 같은 파라미터 이름, 또는 동적으로 조립되는 `ORDER BY`/컬럼명처럼 "값"이 아닌 "식별자"가 개입하는 문자열

반대로, 아래는 이 카드의 기본 결론을 그대로 적용하면 오히려 읽기/유지보수성이 나빠질 수 있어 비범위로 둡니다.

- 사용자에게 보여주는 문구, 로그 메시지처럼 "표현"이 목적이고 재사용 의미가 약한 문자열
- 한 블록 안에서 문장으로 읽히는 것이 중요한 정적 SQL 본문(값 주입은 바인딩으로 통제하되, SQL을 조각 상수로 과분해하는 것은 기본값이 아닙니다)

---

## 3. 기본 결론(기본값)과 선택 이유
여기서의 "기본값"은 "항상"이 아니라, 실패 비용이 큰 영역을 먼저 안전하게 만드는 쪽에 기준을 둔 선택입니다.

첫째, 같은 의미를 여러 위치에서 공유하는 문자열은 상수(또는 닫힌 집합이면 enum/값 객체)로 승격하시는 편이 안전합니다.
이렇게 하면 "어느 한 군데만 수정하고 다른 곳은 놓치는" 부분 수정(partial update)과, 표기 드리프트(같은 의미의 서로 다른 표기)가 만들어내는 런타임 장애를 구조적으로 줄일 수 있습니다.

둘째, 정적 SQL 본문은 읽기 흐름을 유지하는 것이 리뷰/디버깅 속도에 직접적인 영향을 주기 때문에, 본문을 한 덩어리 문장으로 유지하는 것이 기본값입니다.
다만 SQL의 "값"은 반드시 바인딩(PreparedStatement 바인딩, 또는 named parameter 바인딩)으로 주입해야 하며, 문자열 결합으로 값을 끼워 넣는 방식은 안전하지 않습니다.

셋째, SQL에서 동적으로 바뀌는 것이 "값"이 아니라 "식별자"(예: 정렬 컬럼, 테이블/컬럼명)라면 바인딩으로는 해결되지 않습니다.
이 경우는 오히려 타입 시스템을 쓰는 쪽이 안전합니다. 즉, 허용 가능한 식별자 후보를 enum(allow-list)으로 고정하고 그 enum만으로 SQL 조립을 허용해야 합니다.

---

## 4. 용어 정의 (이 카드 기준)
이 카드에서 용어를 다음 의미로 고정합니다.

### 4.1 문자열 리터럴(String literal)
코드에 직접 박혀 있는 문자열 값입니다. 예: `"inbound-profile"`, `"SELECT ..."`.

### 4.2 의미 단위 문자열(semantic identifier string)
프로그램이 "식별자"로 해석하는 문자열입니다. 이 문자열 자체가 분기/검증/매핑을 좌우합니다.

- 포함: 설정 키, 프로토콜 필드명, 모드/프로파일, 정책 키, SQL named parameter 이름, 동적 SQL 식별자(컬럼/정렬 키)
- 제외: 문장 자체가 목적(표현)인 문자열

의미 단위 문자열이 위험한 이유는, 한 글자 오타가 컴파일을 통과하고 런타임까지 숨어 있다가, 종종 "멀리서"(완전히 다른 위치에서) `null`, 분기 누락, 잘못된 라우팅 같은 형태로 터지기 때문입니다.

### 4.3 표현 단위 문자열(human-readable expression string)
사람이 한 번에 읽어야 의미가 살아나는 문자열입니다.
정적 SQL 본문이 대표적입니다. SQL을 상수로 분해하면 "한 문장"이 깨져서, 어디서 무엇을 조회하는지 읽는 사람이 머릿속에서 다시 조립해야 합니다.

### 4.4 드리프트(drift)
"같은 의미"가 시간이 지나며 "서로 다른 표기"로 분화되는 현상입니다.

예: `inbound-profile`, `inboundProfile`, `inbound_profile`.
이 세 값은 사람의 눈에는 비슷하지만, 프로그램 관점에서는 전혀 다른 키입니다.

### 4.5 바인딩(binding)
SQL 문자열에 값을 직접 붙이는 대신, 드라이버/프레임워크가 제공하는 파라미터 자리에 값을 안전하게 결합하는 방식입니다.

- 입력: SQL 템플릿(예: `WHERE name = ?` 또는 `WHERE name = :name`) + 값
- 변환 규칙: 값은 SQL 문법의 일부로 "해석"되지 않고, 데이터로 "취급"됩니다.
- 산출물: DB로 전달되는 실행 계획 + 파라미터

여기서 중요한 경계는 "값(value)"과 "식별자(identifier)"입니다.
바인딩은 값에만 적용됩니다. 컬럼명/정렬키 같은 식별자는 바인딩할 수 없기 때문에 별도의 통제가 필요합니다.

---

## 5. 메커니즘: 상수/enum 승격이 실제로 줄이는 오류의 종류
"상수로 바꾸면 오타가 줄어듭니다"는 너무 얕은 설명이라 실무 의사결정에 도움이 되지 않습니다.
어떤 종류의 오류가 어떤 메커니즘으로 줄어드는지 닫아 보겠습니다.

### 5.1 상수 승격(특히 의미 단위 문자열)
- 입력: 같은 의미의 문자열 리터럴이 여러 파일/함수에 흩어져 있습니다.
- 변환 규칙: 그 문자열에 이름을 붙여(상수화) 한 곳에 모으고, 나머지는 상수를 참조합니다.
- 산출물: 동일한 의미의 문자열이 "항상 같은 값"으로만 사용됩니다(부분 수정이 구조적으로 어려워집니다).
- 소비자: 설정 읽기/검증, 프로토콜 파싱, 분기 로직, 로깅/에러 메시지

이때 상수화가 막아주는 실패는 주로 두 가지입니다.

- 부분 수정(partial update): 키 이름을 바꿔야 할 때, A 파일은 바꿨는데 B 파일은 놓쳐서 런타임에만 깨지는 상황
- 의미 충돌(confusion): `"profile"`이라는 문자열이 여러 의미로 쓰여 전체 검색 시 잘못 수정하는 상황

상수명(예: `FIELD_INBOUND_PROFILE`)은 그 자체가 "의미"를 드러내는 표지(anchor)입니다.
즉, 값(문자열)만 공유하는 것이 아니라 "의도"까지 함께 묶는 효과가 있습니다.

다만 상수화만으로 해결되지 않는 것이 있습니다.
외부 입력이 잘못된 키(`inboundProfile`)로 들어오면, 상수는 그 입력을 "고쳐주지" 않습니다.
이 경우에는 경계에서의 검증(필수 키 존재/허용 키 집합 검사)이 필요합니다.

### 5.2 enum/값 객체 승격(닫힌 집합 식별자)
의미 단위 문자열 중에서도 값 후보가 정해진 "닫힌 집합"(mode/profile/operation type)은 상수보다 enum이 더 강력합니다.

- 입력: 외부에서 `"TRANSLATE"` 같은 문자열이 들어옵니다.
- 변환 규칙: 경계에서 `String -> enum`으로 변환합니다(잘못된 값은 즉시 실패).
- 산출물: 내부 로직은 `String`이 아니라 `enum`을 받습니다.
- 소비자: `switch` 분기

이렇게 하면 "새 값이 추가되었는데 분기 누락" 같은 종류의 버그를 컴파일 단계에서 드러내는 설계가 가능해집니다.
(이 주제는 `K-TYPE-BOUNDARY-ENUM-FAILFAST` 카드에서 더 깊게 다룹니다.)

---

## 6. 의사결정: 무엇을 상수/enum으로 올리고 무엇을 인라인으로 둘 것인가
아래는 실무에서 가장 실패 비용이 낮은 판단 순서입니다.

1) 이 문자열은 프로그램이 "식별자"로 해석합니까, 아니면 사람이 읽기 위한 "문장"입니까?
식별자라면 의미 단위 문자열입니다. 이 경우는 상수/enum 승격을 우선 검토합니다.
문장이라면 표현 단위 문자열입니다. 이 경우는 인라인 유지가 기본값입니다.

2) 식별자라면, 그 값 후보가 닫힌 집합입니까?
닫힌 집합이면 enum/값 객체가 기본값입니다. 상수로만 두면 분기 누락을 컴파일 타임에 강제하기 어렵습니다.

3) 이 문자열이 경계에 걸려 있습니까?
외부 입력(YAML/HTTP/TCP/JSON)에서 들어오는 문자열이라면 "검증 없이 내부로 전파"되기 쉽습니다.
이 경우는 상수/enum 승격과 별개로, 경계에서 fail-fast 검증을 설계해야 합니다.

4) SQL이라면, 이 문자열은 "값"입니까 "식별자"입니까?
값이면 바인딩으로만 주입합니다.
식별자(컬럼명/정렬키)면 바인딩이 불가능하므로, 허용 후보를 enum(allow-list)으로 고정하고 그 enum만으로 조립합니다.

---

## 7. 코드 패턴 예시
아래 예시는 특정 프로젝트 파일에 의존하지 않도록, 카드 안에서 학습 가능한 최소 단위로 구성했습니다.

### 7.1 안 좋은 패턴: 의미 키를 인라인으로 반복(드리프트와 부분 수정을 부릅니다)
```java
import java.util.Map;

class BadConfigReader {
    String readProfile(Map<String, String> listener) {
        // 같은 의미의 키가 서로 다른 표기로 흩어지면, 한쪽은 항상 null이 됩니다.
        String profile = listener.get("inbound-profile");
        if (profile == null) {
            profile = listener.get("inboundProfile"); // 드리프트: '비슷해 보이지만' 완전히 다른 키
        }
        return profile;
    }
}
```

이 패턴의 문제는 "지금은 우연히 동작"할 수 있다는 점입니다.
어느 키가 들어오느냐에 따라 동작이 달라지고, 장애는 설정/입력 데이터에 의존해 간헐적으로 터집니다.

### 7.2 권장 패턴: 의미 단위 문자열을 상수로 승격하고, 경계에서 검증합니다
```java
import java.util.Map;

final class ListenerKeys {
    static final String INBOUND_PROFILE = "inbound-profile";

    private ListenerKeys() {}
}

class GoodConfigReader {
    String readRequired(Map<String, String> map, String key) {
        String value = map.get(key);
        if (value == null || value.isBlank()) {
            // 관측 가능한 실패: "어떤 키"가 비었는지 메시지로 고정됩니다.
            throw new IllegalArgumentException("required key is missing: " + key);
        }
        return value;
    }

    String readProfile(Map<String, String> listener) {
        // 키는 상수로 고정하고, 잘못된 입력은 경계에서 즉시 실패시킵니다.
        return readRequired(listener, ListenerKeys.INBOUND_PROFILE);
    }
}
```

여기서 중요한 점은 "상수화"와 "검증"의 역할이 다르다는 것입니다.

- 상수화는 내부 코드에서의 부분 수정/오타/의미 충돌을 줄입니다.
- 검증은 외부 입력이 틀렸을 때, 실패를 런타임 깊은 곳이 아니라 "경계"에서 터뜨립니다.

### 7.3 SQL 패턴: 정적 본문은 읽기 좋게 유지하고, 값은 바인딩으로만 주입합니다
"정적 SQL 본문"은 아래처럼 한 덩어리로 읽히는 것이 가장 큰 장점입니다.

```java
String sql = """
    SELECT transaction_id, status, result_code
    FROM bridge_transactions
    WHERE operation = :operation
      AND idempotency_key = :idempotency_key
    ORDER BY created_at DESC
    LIMIT 1
    """;

// named parameter 바인딩(예: NamedParameterJdbcTemplate)의 입력 형태를 가정한 예시입니다.
Map<String, Object> params = Map.of(
    "operation", operation,
    "idempotency_key", idempotencyKey
);
```

이 형태의 안전성 핵심은 "SQL 본문을 상수로 쪼갰는가"가 아니라,
`operation`/`idempotencyKey` 같은 값을 문자열 결합으로 붙이지 않고 바인딩으로 넘긴다는 점입니다.

### 7.4 (중요) 동적 SQL 식별자: 바인딩이 불가능하므로 allow-list(enum)로만 조립합니다
정렬 컬럼/정렬 방향 같은 식별자는 바인딩할 수 없습니다.
따라서 외부 입력을 그대로 붙이면 SQL injection과 동일한 성격의 취약점이 됩니다.

안전한 기본값은 "허용 후보를 enum으로 고정"하는 것입니다.

```java
enum SortKey {
    CREATED_AT("created_at"),
    AMOUNT("amount");

    private final String column;

    SortKey(String column) {
        this.column = column;
    }

    String column() {
        return column;
    }
}

class SafeSqlBuilder {
    String orderBy(SortKey key) {
        // 식별자는 바인딩할 수 없으므로, 외부 입력 String을 직접 받지 않습니다.
        return " ORDER BY " + key.column() + " DESC";
    }
}
```

이 예시는 "동적 SQL이 늘어날수록 타입 시스템의 가치가 커진다"는 것을 보여줍니다.

---

## 8. 단점과 트레이드오프 (그리고 왜 그래도 하는가)
상수화/타입화는 만능이 아닙니다. 단점이 있기 때문에, 이 단점을 이해하지 못하면 "무지성 상수화"로 흐르기 쉽습니다.

첫째, 과도한 상수화는 읽기 흐름을 끊습니다.
코드를 읽을 때 매번 상수 정의 파일로 점프해야 하면, 작은 변경에도 인지 부하가 커집니다.

둘째, `Constants` 같은 거대한 중앙 파일은 충돌을 부릅니다.
의미가 다른 문자열들이 한 파일에 모이면, 오히려 "여기에 뭐가 들어있는지"가 불명확해지고, 상수명 네이밍 경쟁이 생깁니다.

셋째, SQL을 조각 상수로 분해해 문자열 결합으로 조립하면, 쿼리 자체를 한 번에 읽기 어려워집니다.
로그로 SQL을 확인할 때도 조립 결과를 다시 추적해야 하므로 디버깅 비용이 늘 수 있습니다.

그럼에도 이 패턴을 채택하는 정당화는 다음 메커니즘에서 나옵니다.

- 의미 단위 문자열은 오타가 컴파일에서 잡히지 않기 때문에, 가장 자주 "운영에서" 비싸게 터집니다.
- 상수/enum 승격은 이러한 실수를 "코드 구조"로 밀어 올려, 리뷰와 리팩터링에서 발견될 확률을 크게 올립니다.
- SQL은 본문 가독성을 유지하면서, 값 주입을 바인딩으로 고정하는 것만으로도 보안/정합성 위험을 크게 줄일 수 있습니다.

---

## 9. 다르게 판단해야 하는 경우(Variants)
기본 결론과 다르게 판단해야 하는 대표 케이스를 미리 고정해두면, 팀이 "예외"를 감정이 아니라 조건으로 다룰 수 있습니다.

- 문자열이 열린 집합(open set)이라면 enum으로 고정하지 않습니다.
  - 예: 사용자 검색어, 자유 입력 라벨, 외부 시스템의 임의 ID
  - 이 경우는 문자열을 유지하되, 길이/포맷/금지 문자 같은 validator 계약을 강화합니다.

- 재사용이 1회라도, 실패 비용이 매우 큰 "식별자"라면 상수/enum 승격을 고려합니다.
  - 예: 동적 SQL 식별자(정렬 컬럼), 보안 경계에서의 키
  - 이 경우는 "횟수"보다 "실패 비용"이 우선입니다.

- SQL이 동적으로 복잡해진다면(필드 선택/조건 조합이 급증) 문자열만으로 관리하지 않습니다.
  - 선택지: enum allow-list 강화, 조립 규칙을 별도 모듈로 격리, SQL DSL 도입
  - 핵심은 "문자열 결합이 유일한 경로"가 되지 않게 만드는 것입니다.

---

## 10. 최소 실험(검증): 이 결론을 직접 확인하기
아래 실험은 특정 프로젝트 도구 없이, 자바 코드만으로 확인할 수 있는 수준으로 고정합니다.

### 10.1 드리프트가 실제 버그가 되는지 확인
```java
import java.util.Map;

final class Keys {
    static final String INBOUND_PROFILE = "inbound-profile";
    private Keys() {}
}

public class DriftLab {
    static String readProfileBad(Map<String, String> m) {
        // 사람이 보기엔 비슷하지만, 프로그램에겐 완전히 다른 키입니다.
        return m.get("inboundProfile");
    }

    static String readProfileGood(Map<String, String> m) {
        return m.get(Keys.INBOUND_PROFILE);
    }

    public static void main(String[] args) {
        Map<String, String> ok = Map.of(Keys.INBOUND_PROFILE, "KPN_200");

        System.out.println(readProfileBad(ok));  // 기대 결과: null
        System.out.println(readProfileGood(ok)); // 기대 결과: KPN_200
    }
}
```

이 실험이 보여주는 것은 단순합니다.
"키 문자열"은 한 글자만 달라도 즉시 `null`이 되고, 그 `null`은 종종 더 깊은 로직에서 NPE/분기 오류로 증폭됩니다.

### 10.2 값 바인딩이 왜 필요한지(문자열 결합의 위험 확인)
```java
public class SqlInjectionLab {
    public static void main(String[] args) {
        String userInput = "x' OR '1'='1";

        String unsafe = "SELECT * FROM users WHERE name = '" + userInput + "'";
        System.out.println(unsafe);
        // 기대 결과:
        // SELECT * FROM users WHERE name = 'x' OR '1'='1'
        // (값이 SQL 문법으로 해석될 여지가 생깁니다)

        String safeTemplate = "SELECT * FROM users WHERE name = ?";
        System.out.println(safeTemplate);
        // 기대 결과:
        // SELECT * FROM users WHERE name = ?
        // (값은 바인딩으로 전달되어 SQL 문법으로 해석되지 않습니다)
    }
}
```

여기서 핵심은 "SQL 문자열을 상수로 뺐는가"가 아니라, 값이 SQL 문법으로 합쳐지는 경로가 있는가/없는가입니다.

### 10.3 동적 식별자(정렬 컬럼)는 바인딩할 수 없음을 확인
"값"은 바인딩할 수 있지만, "컬럼명"은 바인딩할 수 없습니다.
따라서 외부 입력 String을 `ORDER BY <입력>`에 직접 붙이지 말고, enum allow-list로만 통과시키셔야 합니다.

이 결론은 위의 `SortKey` 예시처럼 "외부 String을 받지 않는" 형태의 API로 가장 쉽게 강제됩니다.

---

## 11. 흔한 오해
- 오해: "2번 이상 쓰이면 무조건 상수화"가 정답입니다.
  - 정정: 재사용 횟수는 힌트일 뿐입니다. 핵심은 "같은 의미"를 공유하는지, 그리고 실패 비용이 얼마나 큰지입니다.

- 오해: "상수화하면 입력 검증도 끝"입니다.
  - 정정: 상수화는 내부의 부분 수정/드리프트를 줄입니다. 외부 입력이 잘못된 경우는 경계 검증이 담당해야 합니다.

- 오해: "SQL은 다 상수로 쪼개면 안전"합니다.
  - 정정: 안전성의 핵심은 바인딩입니다. 문자열 조각 상수는 안전을 보장하지 않고, 가독성만 해칠 수 있습니다.

---

## 12. 팀 적용 체크리스트
아래 질문에 팀이 같은 결론을 낼 수 있으면, 상수화 범위를 건강하게 유지하고 계신 상태입니다.

- 이 문자열은 의미 단위입니까, 표현 단위입니까?
- 의미 단위라면, 한 곳에서 이름을 가지고 관리되고 있습니까(상수/enum/값 객체)?
- 외부 입력이 이 값을 만든다면, 경계에서 fail-fast 검증이 있습니까?
- SQL 값 주입이 문자열 결합이 아니라 바인딩으로 통제되고 있습니까?
- 동적 SQL 식별자(정렬/컬럼)는 allow-list(enum)로만 조립되도록 막혀 있습니까?
