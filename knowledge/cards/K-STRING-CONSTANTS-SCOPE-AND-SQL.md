# K-STRING-CONSTANTS-SCOPE-AND-SQL

## 1) 이 카드가 답하는 질문
실무에서는 아래 질문이 반복됩니다.

"코드에 흩어진 문자열 리터럴을 어디까지 상수/enum으로 올려야 하고, 어디부터는 인라인으로 두는 것이 더 낫습니까? 특히 SQL 문자열은 가독성을 해치지 않으면서도 안전하게 유지하려면 어떻게 해야 합니까?"

이 질문이 어렵게 느껴지는 이유는, 문자열을 상수화하면 오타/수정 누락은 줄지만 읽기 흐름이 끊길 수 있고, 반대로 인라인을 유지하면 지금은 빠르지만 시간이 지날수록 드리프트(같은 의미의 서로 다른 표기)와 부분 수정(한 군데만 바꾸고 다른 곳은 놓침)이 누적되기 때문입니다.
이 카드는 이 균형을 "취향"이 아니라 "실패 비용을 어디에서 줄일지"라는 관점으로 정리합니다.

---

## 2) 배경: 왜 이 문제가 실무에서 비싸게 터지는가
문자열은 컴파일러가 의미를 알지 못합니다.
따라서 오타가 있어도 컴파일은 통과하고, 실패는 런타임까지 숨어 있다가 더 깊은 곳에서 관측됩니다.
이때 관측되는 증상은 보통 `null`, 분기 누락, 잘못된 라우팅, SQL 에러처럼 "2차 증상"이고, 원인(문자열 한 글자 차이)은 경계나 설정 근처에 숨어 있기 때문에 디버깅 비용이 커집니다.

또 한 가지 비용은 팀 협업에서 발생합니다.
문자열이 여기저기 흩어져 있으면, 변경 시 전체 검색에 의존하게 되고, 이름이 비슷한 문자열을 잘못 수정하는 위험이 올라갑니다.
반대로 모든 문자열을 상수로 빼면, 코드가 한 문장으로 읽히지 않아 의도를 파악하기 어렵고 리뷰 속도가 떨어질 수 있습니다.

특히 SQL은 실무에서 실패 비용이 더 큽니다.
값을 문자열 결합으로 붙이면 입력이 SQL 문법으로 해석될 수 있어 보안 문제가 생기고, 동적 조립이 늘면 "어떤 SQL이 실제로 실행됐는지"를 재현하기 어려워집니다.
따라서 SQL에서는 "가독성"뿐 아니라 "경계(값/식별자)"를 어떻게 지키는지가 안전성의 핵심이 됩니다.

---

## 3) 핵심 제어점(왜 하필 이 축인가)
이 주제는 요소를 나열하면 끝이 없습니다.
그래서 먼저 "이 축을 바꾸면 시스템의 성질이 바뀐다"는 제어점을 3개로 고정하고, 나머지 판단은 이 제어점 위에서 파생시키는 방식이 안전합니다.

이 카드의 핵심 제어점은 세 가지이고, 각각은 서로 다른 실패 비용을 설명하는 "최소 축"입니다.
즉 이 세 축을 먼저 고정해 두면, 뒤에서 용어(의미 단위/표현 단위, 바인딩)와 메커니즘(입력 -> 변환 -> 산출물 -> 소비자), 조건 변경 시 예측이 같은 뼈대 위에서 연결됩니다.

첫째는 "문자열을 의미 단위(식별자)와 표현 단위(문장)로 나눌 수 있는가"입니다.
의미 단위는 프로그램이 키로 해석하므로 오타가 런타임까지 숨어 비용이 커지고, 표현 단위는 사람이 한 덩어리로 읽어야 하므로 과도한 상수화가 인지 비용을 키웁니다.
따라서 이 구분이 무너지면, 식별자는 인라인으로 방치되어 드리프트가 쌓이거나, 반대로 문장은 조각으로 분해되어 읽기 흐름이 깨지는 두 실패 중 하나로 흔들립니다.

둘째는 "외부 입력이 들어오는 경계에서 fail-fast로 검증하는가"입니다.
상수화는 내부 코드에서 같은 의미를 같은 이름으로 쓰게 만드는 장치이지만, 외부 입력이 이미 잘못된 형태로 들어오는 문제를 고쳐주지 못합니다.
그래서 경계에서 필수 키/허용 후보/표준화 정책을 검증해 실패를 경계로 당기지 않으면, 실패가 내부 깊은 곳에서 2차 증상으로만 관측되어 디버깅 비용이 커집니다.

셋째는 "SQL에서 값(value)과 식별자(identifier)를 분리해 통제하는가"입니다.
값은 바인딩으로 데이터로만 취급되게 만들 수 있지만, 식별자(컬럼/정렬 키)는 바인딩할 수 없습니다.
따라서 식별자를 외부 `String`으로 받아 SQL에 붙이는 순간, 그 지점은 입력이 SQL 구조를 바꾸는 취약점이 되며, 이 영역은 allow-list(enum) 같은 형태로만 통제해야 합니다.

왜 하필 이 세 가지를 먼저 보느냐는 질문에 대한 답은, 실무에서 관측되는 실패가 대체로 이 세 축으로 설명되기 때문입니다.
첫째는 "설정/프로토콜이 있는데도 동작이 안 된다"(드리프트/부분 수정으로 인한 정합성 실패),
둘째는 "어떤 입력에서만 간헐적으로 깨진다"(경계 검증 부재로 인한 실패 위치 분산),
셋째는 "SQL이 입력에 의해 바뀐다"(보안/정합성 실패)입니다.
이 세 축은 모두 로그/테스트로 관측 가능하고, 코드 레벨에서 즉시 실험할 수 있어 학습의 출발점으로 고정하기에 적합합니다.

---

## 4) 기본 결론(기본값)과 선택 이유
기본 결론은 다음처럼 닫습니다.

문자열을 먼저 "의미 단위(식별자)"와 "표현 단위(문장)"로 나눈 뒤, 의미 단위는 상수/enum으로 통제하고 표현 단위는 읽기 흐름을 우선해 인라인으로 둡니다.
그리고 SQL은 정적 본문 가독성을 유지하되, 값은 반드시 바인딩으로 주입하고, 식별자는 allow-list(enum)로만 조립되도록 제한합니다.

이 결론의 근거는 실패 비용 구조에 있습니다.
의미 단위 문자열은 오타가 컴파일에서 잡히지 않으므로 운영에서 비싸게 터지고, 관측 지점이 원인과 멀어지는 경향이 있습니다.
반면 표현 단위 문자열(특히 정적 SQL 본문)은 사람이 한 덩어리로 읽을 때 이해 속도가 가장 빠르므로, 과도한 상수화는 안전성보다 인지 비용을 더 크게 올릴 수 있습니다.
SQL은 여기에 보안까지 얹히기 때문에, "문장을 어떻게 보관하느냐"보다 "값/식별자 경계를 어떻게 강제하느냐"가 더 중요합니다.

---

## 5) 용어 정의 (이 카드 기준)
### 5.1 의미 단위 문자열(semantic identifier string)
프로그램이 "식별자"로 해석하는 문자열입니다. 이 문자열 자체가 분기/검증/매핑의 기준이 됩니다.
예를 들어 설정 키, 프로토콜 필드명, 모드/프로파일, 정책 키, SQL named parameter 이름, 동적 SQL 식별자(컬럼/정렬 키)는 모두 의미 단위 문자열입니다.

여기서 중요한 경계는 "사람이 비슷해 보인다"와 "프로그램이 같은 값으로 취급한다"가 완전히 다르다는 점입니다.
키 조회는 완전 일치가 기본이므로, 한 글자만 달라도 즉시 `null`이 되고, 그 `null`은 멀리서 NPE나 분기 오류로 증폭될 수 있습니다.

### 5.2 표현 단위 문자열(human-readable expression string)
사람이 한 번에 읽어야 의미가 살아나는 문자열입니다.
정적 SQL 본문이 대표적입니다. SQL을 조각 상수로 과분해하면 "한 문장"이 깨져서, 읽는 사람이 머릿속에서 다시 조립해야 합니다.

### 5.3 바인딩(binding)
SQL 문자열에 값을 직접 붙이지 않고, 드라이버/프레임워크가 제공하는 파라미터 자리에 값을 안전하게 전달하는 방식입니다.
바인딩의 핵심은 값이 SQL 문법의 일부로 해석되지 않고 데이터로 취급되도록 "경계"를 유지하는 것입니다.
따라서 바인딩은 값(value)에만 적용할 수 있고, 컬럼명/정렬키 같은 식별자(identifier)에는 적용할 수 없습니다.

---

## 6) 메커니즘(입력 -> 변환 규칙 -> 산출물 -> 소비자)
여기서는 상수화/바인딩이 "왜" 안전한지, 실제 데이터 흐름으로 닫습니다.

의미 단위 문자열에 대한 상수화의 입력은 "같은 의미가 여러 곳에 흩어진 문자열 리터럴"입니다.
변환 규칙은 그 문자열에 이름을 붙여 한 곳에 모으고, 나머지 사용처는 상수를 참조하도록 만드는 것입니다.
산출물은 "동일 의미는 동일 상수"라는 구조이고, 소비자는 설정 읽기/검증, 프로토콜 파싱, 분기 로직, 에러 메시지 조립입니다.
이 구조의 효과는 부분 수정과 표기 드리프트를 구조적으로 줄이는 것입니다.

다만 상수화는 외부 입력을 고쳐주지 못합니다.
외부에서 `inboundProfile`이라는 키로 값이 들어오면, 내부 상수(`INBOUND_PROFILE`)는 그 입력을 자동으로 표준화하지 않습니다.
따라서 외부 입력이 들어오는 경계에서는 "필수 키 존재" 또는 "허용 키 집합"을 검증해 실패를 경계로 당겨야 합니다.

SQL 바인딩의 입력은 SQL 템플릿과 값입니다.
변환 규칙은 값이 SQL 문법으로 해석되지 않도록, DB 드라이버가 파라미터로 전달하는 것입니다.
산출물은 "실행할 SQL + 파라미터"이며, 소비자는 DB의 파서/실행기입니다.
이때 값이 데이터로 취급되기 때문에, 문자열 결합 방식에서만 발생하는 SQL injection 계열의 문제가 구조적으로 줄어듭니다.

동적 SQL 식별자의 경우 입력은 "정렬 컬럼" 같은 식별자 후보입니다.
바인딩은 식별자에 적용할 수 없으므로, 변환 규칙은 "외부 String을 받지 않고 enum(allow-list)만 받는 API"로 제한하는 것입니다.
산출물은 enum이 보장하는 안전한 문자열 조각이고, 소비자는 SQL 조립 코드입니다.
즉, 이 경우의 안전성은 바인딩이 아니라 타입 시스템(허용 후보 제한)이 담당합니다.

---

## 7) 조건 변경 시 예측(무엇이 깨지고 어떻게 관측되는가)
이 절은 위에서 고정한 제어점을 일부러 흔들어 보면서, 무엇이 깨지고 어떤 증상으로 관측되는지 설명합니다.

### 7.1 의미/표현 구분을 하지 않으면 무엇이 깨지나
의미 단위와 표현 단위를 구분하지 않으면, 팀은 두 극단 중 하나로 흔들리기 쉽습니다.

의미 단위를 인라인으로 방치하면 드리프트가 쌓입니다.
예를 들어 설정에서 실제 키가 `inbound-profile`인데 코드가 `inboundProfile`로 읽으면, 프로그램은 즉시 `null`을 반환하고, 그 `null`은 훨씬 뒤에서 NPE로 관측될 수 있습니다.
관측 포인트는 "값이 비었다"가 아니라 "키가 달라서 못 읽었다"라는 메시지를 경계에서 고정할 수 있는지입니다.

표현 단위를 과도하게 상수화하면 가독성이 깨집니다.
정적 SQL 본문을 `SELECT`, `FROM`, `WHERE` 같은 조각으로 나눠 문자열 결합으로 조립하면, 리뷰어는 쿼리를 머릿속에서 다시 합쳐야 합니다.
이때 문제는 "읽기 어렵다"에서 끝나지 않고, 문자열 조립 버그로도 이어질 수 있습니다.
예를 들어 조각을 결합할 때 공백이 하나 빠지면 최종 SQL이 `FROMbridge_transactions`처럼 토큰 경계를 잃고, DB는 이를 문법 오류로 즉시 거절합니다.
따라서 관측은 느낌이 아니라 출력으로 고정하는 편이 안전합니다. 실행 직전에 로깅/출력한 SQL 문자열에서 토큰 경계가 붙어 있는지(예: `FROMbridge_transactions`)를 보거나, DB가 반환하는 문법 오류가 키워드 경계 주변에서 발생하는지로 조립 실수를 빠르게 의심할 수 있어야 합니다.

### 7.2 상수화는 했는데 경계 검증을 하지 않으면 무엇이 깨지나
상수화는 내부 코드의 일관성을 강제하지만, 외부 입력이 틀린 경우를 잡아주지 못합니다.
따라서 경계 검증이 없으면 잘못된 키/값이 내부로 전파되고, 실패 위치가 분산됩니다.

가장 흔한 증상은 "설정은 있는데도 동작이 안 된다"입니다.
내부에서는 값이 `null`이거나 기본값으로 대체되어 조용히 잘못된 경로로 흐를 수 있고, 그 결과가 한참 뒤에만 관측됩니다.
관측 포인트는 "필수 키가 없을 때" 즉시 실패하는지, 그리고 에러 메시지가 어떤 키가 문제인지 고정하는지입니다.
예를 들어 `"required key is missing: inbound-profile"`처럼 키 이름이 메시지에 박혀 있으면, 장애의 원인 위치가 경계로 고정됩니다.

### 7.3 SQL 값 주입을 문자열 결합으로 하면 무엇이 깨지나
값을 문자열 결합으로 붙이면 입력이 SQL 문법으로 해석될 수 있습니다.
즉 입력이 데이터가 아니라 쿼리 구조를 바꾸는 토큰으로 취급될 수 있고, 이것이 SQL injection의 메커니즘입니다.

관측 포인트는 매우 직접적입니다.
입력 값에 작은 따옴표 같은 문법 문자가 들어갔을 때, SQL 문자열 자체가 변해버리는지 출력으로 확인할 수 있습니다.
바인딩을 사용하면 SQL 템플릿은 변하지 않고 값만 파라미터로 전달됩니다.

### 7.4 동적 식별자를 외부 String으로 받으면 무엇이 깨지나
정렬 컬럼/테이블명/컬럼명 같은 식별자는 바인딩할 수 없습니다.
따라서 이 값을 외부 String으로 받아 그대로 SQL에 붙이는 순간, 그 지점은 "SQL 구조를 외부 입력이 결정하는" 취약점이 됩니다.

관측 포인트는 "허용 후보가 아닌 값"이 들어왔을 때의 동작입니다.
allow-list(enum)로 제한되어 있으면 컴파일 타임 또는 파싱 단계에서 값이 걸러지고, 외부 입력이 SQL 구조를 바꿀 수 없습니다.
이때 중요한 것은 "거절"이 조용히 사라지지 않도록, 거절이 관측 가능한 신호로 남는 것입니다. 예를 들어 `"unknown sort key: <raw>"` 같은 메시지로 실패를 고정하거나, 거절 카운터를 올려 외부 입력이 SQL 구조를 바꾸려 했음을 조기에 감지할 수 있어야 합니다.

반대로 String으로 받으면, 어떤 입력이 들어왔는지에 따라 SQL이 달라지고, 장애/취약점이 재현되기 어려워집니다.

---

## 8) 코드 예시(나쁜 패턴 1 + 권장 패턴 1 이상)
아래 예시는 카드 안에서 학습 가능한 최소 단위로 구성했습니다.

### 8.1 의미 키 드리프트(나쁜 패턴)
```java
import java.util.Map;

class BadConfigReader {
    String readProfile(Map<String, String> listener) {
        // 사람이 보기엔 비슷하지만 프로그램에겐 완전히 다른 키입니다.
        return listener.get("inboundProfile");
    }
}
```

### 8.2 의미 키 상수화 + 경계 검증(권장 패턴)
```java
import java.util.Map;

final class ListenerKeys {
    static final String INBOUND_PROFILE = "inbound-profile";
    private ListenerKeys() {}
}

class GoodConfigReader {
    static String readRequired(Map<String, String> map, String key) {
        String value = map.get(key);
        if (value == null || value.isBlank()) {
            // 실패를 경계로 당겨 원인을 고정합니다.
            throw new IllegalArgumentException("required key is missing: " + key);
        }
        return value;
    }

    String readProfile(Map<String, String> listener) {
        return readRequired(listener, ListenerKeys.INBOUND_PROFILE);
    }
}
```

### 8.3 정적 SQL 본문 가독성 유지 + 값 바인딩(권장 패턴)
```java
String sql = """
    SELECT transaction_id, status, result_code
    FROM bridge_transactions
    WHERE operation = :operation
      AND idempotency_key = :idempotency_key
    ORDER BY created_at DESC
    LIMIT 1
    """;

Map<String, Object> params = Map.of(
    "operation", operation,
    "idempotency_key", idempotencyKey
);
```

### 8.4 동적 식별자 allow-list(enum)로 제한(권장 패턴)
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
        // 식별자는 바인딩할 수 없으므로, 외부 String을 직접 받지 않습니다.
        return " ORDER BY " + key.column() + " DESC";
    }
}
```

---

## 9) 최소 실험(검증)
아래 실험은 특정 프로젝트 도구 없이, 자바 코드만으로 확인할 수 있는 수준으로 고정합니다.

### 9.1 드리프트가 실제 버그가 되는지 확인
```java
import java.util.Map;

final class Keys {
    static final String INBOUND_PROFILE = "inbound-profile";
    private Keys() {}
}

public class DriftLab {
    static String readProfileBad(Map<String, String> m) {
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

### 9.2 값 바인딩이 왜 필요한지(문자열 결합의 위험 확인)
```java
public class SqlInjectionLab {
    public static void main(String[] args) {
        String userInput = "x' OR '1'='1";

        String unsafe = "SELECT * FROM users WHERE name = '" + userInput + "'";
        System.out.println(unsafe);
        // 기대 결과:
        // SELECT * FROM users WHERE name = 'x' OR '1'='1'

        String safeTemplate = "SELECT * FROM users WHERE name = ?";
        System.out.println(safeTemplate);
        // 기대 결과:
        // SELECT * FROM users WHERE name = ?
    }
}
```

### 9.3 동적 식별자를 String으로 받을 때와 allow-list로 막을 때의 차이 확인
동적 SQL 식별자(정렬 키/컬럼명)는 바인딩할 수 없기 때문에, 이 값이 외부 입력으로 들어오면 "문자열 결합"이 곧 SQL 구조 변경으로 이어질 수 있습니다.
따라서 핵심은 "식별자를 바인딩하자"가 아니라, "식별자는 애초에 외부 String을 받지 않게 하자"입니다.

아래 실험은 그 차이를 출력으로 확인합니다.

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

    static SortKey fromExternal(String raw) {
        if (raw == null || raw.isBlank()) {
            throw new IllegalArgumentException("sort key is required");
        }
        // 외부 입력은 표준화 후 enum으로 변환하고, 허용 후보 밖이면 즉시 거절합니다.
        return SortKey.valueOf(raw.trim().toUpperCase());
    }
}

public class IdentifierLab {
    static String orderByUnsafe(String raw) {
        // 식별자를 외부 String으로 받는 순간, 입력이 SQL 구조를 바꿀 수 있습니다.
        return " ORDER BY " + raw + " DESC";
    }

    static String orderBySafe(String raw) {
        SortKey key = SortKey.fromExternal(raw);
        return " ORDER BY " + key.column() + " DESC";
    }

    public static void main(String[] args) {
        String injected = "amount DESC; DROP TABLE users; --";

        System.out.println(orderByUnsafe(injected));
        // 기대 결과:
        // ORDER BY amount DESC; DROP TABLE users; -- DESC

        try {
            System.out.println(orderBySafe(injected));
        } catch (Exception e) {
            System.out.println(e.getClass().getSimpleName());
            // 기대 결과: IllegalArgumentException
        }
    }
}
```

---

## 10) 트레이드오프/대안/Variants
기본 결론은 강한 기본값이지만, 항상 같은 선택을 하면 오히려 비용이 커질 수 있습니다.
따라서 조건이 달라질 때의 대안을 미리 닫아 두는 것이 안전합니다.

값 후보가 열린 집합(open set)이라면 enum으로 고정하지 않습니다.
이때는 문자열을 유지하되 길이/포맷/금지 문자 같은 validator 계약을 강화합니다. 이 조건에서는 타입 고정이 아니라 입력 검증이 안정성의 중심이 됩니다.

동적 SQL이 급격히 복잡해진다면(필드 선택/조건 조합이 급증) 문자열만으로 관리하지 않습니다.
이때는 enum allow-list를 더 촘촘히 하거나, 조립 규칙을 별도 모듈로 격리하거나, SQL DSL을 도입하는 쪽이 "조립 실수"라는 실패 모드 면적을 줄입니다.

상수화가 과도해져 읽기 흐름이 깨진다면, 상수를 중앙에 몰아넣기보다 "사용되는 맥락"에 가깝게 두는 것이 낫습니다.
즉 전역 `Constants` 파일을 키우기보다, 도메인/모듈 경계에 맞춘 작은 상수 집합으로 분산시키는 것이 탐색성과 충돌 비용을 함께 줄입니다.

---

## 11) 흔한 오해
"2번 이상 쓰이면 무조건 상수화"가 정답이라는 오해가 대표적입니다.
재사용 횟수는 힌트일 뿐이고, 핵심은 같은 의미를 공유하는지와 실패 비용이 얼마나 큰지입니다.

"상수화하면 입력 검증도 끝"이라는 오해도 자주 나옵니다.
상수화는 내부의 부분 수정/드리프트를 줄이지만, 외부 입력이 잘못된 경우는 경계 검증이 담당해야 합니다.

"SQL은 다 상수로 쪼개면 안전"이라는 오해도 있습니다.
안전성의 핵심은 바인딩 경계를 유지하는 것이고, 문자열 조각 상수는 안전을 보장하지 않으며 가독성만 해칠 수 있습니다.

---

## 12) 팀 적용 질문(선택)
이 기준을 팀 규칙으로 쓰려면, 질문을 "각각 따로" 체크하기보다 "순서대로" 통과시키는 형태가 더 안전합니다.
첫 질문에서 의미 단위/표현 단위가 갈리면, 이후 판단(상수화/인라인 유지)의 방향이 바뀌고, SQL은 마지막에서 값/식별자 경계로 다시 한 번 갈립니다.
즉 아래 질문들은 독립적인 체크박스가 아니라, 앞의 답이 뒤의 선택지를 제한하는 하나의 논리 흐름입니다.

- `이 문자열은 의미 단위입니까, 표현 단위입니까?`: 의미 단위(식별자)라면 상수/enum 쪽으로, 표현 단위(문장)라면 읽기 흐름을 지키는 쪽으로 출발점을 고정합니다.
- `의미 단위라면, 한 곳에서 이름을 가지고 관리되고 있습니까(상수/enum/값 객체)?`: 여기서 "한 곳"이 흔들리면 부분 수정과 표기 드리프트가 다시 생깁니다.
- `외부 입력이 이 값을 만든다면, 경계에서 fail-fast 검증이 있습니까?`: 상수화만으로는 잘못된 입력을 고칠 수 없으므로, 실패를 경계로 당겨 원인을 고정해야 합니다.
- `SQL 값 주입이 문자열 결합이 아니라 바인딩으로 통제되고 있습니까?`: 쿼리 본문의 상수화 여부보다 값/문법 경계가 먼저입니다.
- `동적 SQL 식별자(정렬/컬럼)는 allow-list(enum)로만 조립되도록 막혀 있습니까?`: 식별자는 바인딩할 수 없으므로, API 형태로 통제해야 합니다.
