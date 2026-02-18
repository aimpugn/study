# K-BOUNDARY-PARSE-DONT-VALIDATE-TOLERANT-READER

## 1) 이 카드가 답하는 질문

외부 연동에서 아래 질문이 반복됩니다.

"업스트림(API/외부 시스템) 응답을 어디까지 검증해야 합니까? 응답에 예상하지 못한 필드가 추가되면 즉시 실패(fail-fast)해야 합니까, 아니면 무시해도 됩니까? 그리고 그 검증/파싱 로직을 인프라(client)에서 할지, 도메인(use case)에서 할지, 정석에 가까운 기준은 무엇입니까?"

이 질문은 "검증을 많이 하면 안전하다" 같은 감각으로는 답이 나오지 않습니다. 검증은 가용성을 깎을 수도 있고, 반대로 검증을 줄이면 조용한 의미 오류(silent corruption)의 위험이 커질 수 있습니다. 이 카드는 그 균형을 취향이 아니라 제어점과 실패 비용 구조로 고정합니다.

---

## 2) 배경: 왜 이 문제가 실무에서 비싸게 터지는가

외부 연동은 "내가 소유하지 않는 코드"와 대화하는 일입니다. 즉 내가 아무 것도 바꾸지 않아도, 상대가 배포하면 입력 형태(필드 추가/형식 변화/값의 의미)가 달라질 수 있습니다. 이때 엄격하게 "모든 필드가 내가 아는 것만 와야 한다"로 검증하면, 상대의 "추가(additive) 변경"이 곧바로 내 서비스 장애가 됩니다.

반대로 "그냥 파싱만 하고 나머지는 도메인에서 처리"로 밀어버리면, 도메인이 외부 입력의 불안정성과 불일치를 떠안습니다. 결과적으로 도메인 코드가 `null`/공백/형식 분기 같은 방어 코드로 오염되고, 같은 검증이 여러 유스케이스에 흩어져 동작이 갈라집니다. 이 상태에서는 실패가 경계가 아니라 도메인 깊은 곳에서만 관측되어, "어떤 입력이 문제였는지"를 복원하기가 더 어려워집니다.

따라서 핵심은 "검증을 하느냐 마느냐"가 아니라, **무엇을 경계에서 구조화(parse)할지**, **unknown(추가 필드)을 어떻게 다룰지**, **드리프트를 어디서 잡을지(런타임 vs 테스트)**를 일관된 기준으로 고정하는 것입니다.

---

## 3) 핵심 제어점(왜 하필 이 축인가)

이 주제는 디테일이 많기 때문에, 판단을 고정하는 제어점 3개로 축을 좁힙니다.

첫째는 **unknown/extra field를 어떻게 다루는가**입니다. "unknown이면 실패"는 드리프트를 조기에 드러내지만, 제공자의 additive change까지 내 장애로 바꿉니다. "unknown이면 무시"는 가용성을 지키지만, 변화가 생겼다는 사실이 조용히 사라질 수 있습니다. 따라서 이 축은 곧 "가용성 vs 변화 탐지"의 비용 교환입니다.

둘째는 **required field(없으면 의미가 성립하지 않는 필드)를 어디에서 fail-fast로 고정하는가**입니다. required가 도메인 깊은 곳에서야 터지면, 장애의 원인(경계 입력)과 관측(예외/로그)이 멀어지고 복구 비용이 커집니다. 반대로 경계에서 required를 고정하면, 실패가 입력 근처에서 관측되어 원인 추적이 쉬워집니다.

셋째는 **드리프트를 런타임 검증으로 잡을지, 계약 테스트(consumer-driven contract / golden fixture)로 잡을지**입니다. 런타임 검증은 즉시 보호가 되지만, 잘못 설계하면 가용성에 직접 타격을 줍니다. 테스트 기반 탐지는 프로덕션 가용성을 덜 흔들지만, 테스트가 최신 응답을 따라가지 못하면 탐지력이 떨어집니다. 이 축은 "언제 실패를 관측할 것인가"를 결정합니다.

---

## 4) 기본 결론(기본값)과 선택 이유

기본 결론은 다음처럼 닫습니다.

경계에서는 "validate"를 잔뜩 쌓기보다 **parse로 입력을 구조화**하고, 도메인 내부에는 "구조화된 타입"만 들어가게 하십시오. unknown/extra field는 기본적으로 **무시하되 관측(로그/카운터)** 하여 변화가 있었음을 남기고, required field는 **경계에서 fail-fast** 하여 원인을 고정하십시오. 그리고 제공자의 변화는 런타임의 엄격 검증만으로 막으려 하지 말고, **계약 테스트(샘플/fixtures/consumer-driven contracts)**로 조기에 탐지하는 흐름을 함께 두는 편이 안전합니다.

이 선택이 정당화되는 이유는 다음과 같습니다.

첫째, parse는 "불법 상태를 표현하지 못하게" 만들어 도메인을 단순하게 만듭니다. Alexis King이 강조하는 프레임도 이 방향입니다. 값이 유효한지 검사만 하고 통과시키면, 호출자가 검사를 잊는 순간 다시 불법 상태가 내부로 전파됩니다. 반대로 parse는 "유효한 값만 담긴 타입"을 만들어, 검사를 잊어버릴 여지를 줄입니다.

둘째, unknown field에 대한 무조건 실패는 제공자의 additive change에 취약합니다. Martin Fowler가 말하는 Tolerant Reader는 이런 변화에 대해 소비자가 더 관대하게 읽도록(내가 사용하지 않는 정보는 무시) 하여 호환성을 높이는 접근에 가깝습니다.

셋째, unknown을 무시하는 것만으로는 안전하지 않기 때문에 "관측"이 필요합니다. unknown이 늘어나는 순간 그것은 "계약이 변하고 있다"는 신호입니다. 관측이 있으면 조용히 넘어가는 대신, 운영자가 변화를 인지하고 대응할 수 있습니다.

넷째, unknown을 "retryable"로 분류하는 것은 보통 비용이 큽니다. 제공자가 필드를 추가했다고 해서, 같은 요청을 다시 보내면 성공할 확률이 의미 있게 오르지 않습니다. 오히려 재시도로 지연과 부하만 늘어날 수 있습니다. 따라서 retryable 분류는 "재시도로 상태가 바뀔 가능성이 있는 실패"로 제한하는 편이 안전합니다.

---

## 5) 핵심 용어 정의(이 문서 맥락에서)

**Parse(파싱)**는 입력을 받아 "더 강한 구조"로 바꾸는 변환입니다. 예를 들어 `String`을 `NonBlankString`으로, `Map`을 `ResponseEnvelope`로 바꾸는 것이 parse입니다. parse의 결과는 "유효성 조건을 만족한다"는 전제를 가질 수 있어야 합니다.

**Validate(검증)**는 입력을 검사해 통과/실패를 결정하지만, 입력의 형태 자체는 바꾸지 않습니다. 검증만 하고 원래 타입을 그대로 흘려보내면, 이후 호출자가 검증을 잊거나 무시할 수 있습니다.

**Tolerant Reader**는 소비자가 읽을 때 자신이 모르는 정보를 과도하게 가정하지 않고, 추가된 정보(unknown field)를 기본적으로 무시하여 제공자 변경에 덜 깨지게 하는 접근입니다.

**Strict Reader**는 소비자가 자신이 기대한 형태 외에는 거절하여, 드리프트를 강하게 드러내는 접근입니다. 이 접근은 가용성 비용을 치를 수 있습니다.

**계약 테스트(Contract Testing)**는 "소비자가 기대하는 상호작용"을 테스트로 고정하여 제공자/소비자 변경이 호환되는지 조기에 검증하는 방식입니다. consumer-driven contracts는 이 계약을 소비자 관점에서 작성하고, 제공자가 이를 만족하는지 검증하는 흐름을 말합니다.

---

## 6) 메커니즘(입력 -> 변환 -> 산출물 -> 소비자)

외부 응답 처리의 메커니즘을 입력-변환-산출물-소비 흐름으로 닫아야 "검증을 어디에 둘지"가 흔들리지 않습니다.

입력은 보통 bytes 또는 JSON 문자열 같은 "약한 구조"입니다. 변환은 (1) 디코드, (2) 파싱(Map/Node), (3) 필요한 필드를 추출하며 타입화(parse), (4) 도메인 모델로 매핑입니다. 산출물은 도메인이 사용할 수 있는 "구조화된 타입"이며, 소비자는 유스케이스/정책 로직입니다.

unknown field 정책은 이 흐름 중 (2)~(3) 사이에서 결정됩니다. strict라면 (2)에서 전체 키를 검사해 실패시키고, tolerant라면 (3)에서 필요한 키만 읽고 나머지는 무시합니다. 중요한 점은 어떤 정책이든 **required field는 (3)에서 fail-fast로 고정**되어야 한다는 것입니다. 그래야 실패가 "경계 입력"과 가깝게 관측됩니다.

계약 테스트는 이 파서에 대한 외부 검증 장치입니다. 입력(샘플 응답) -> 파서 실행 -> 산출물(도메인 타입) -> 기대 조건(assertions)으로 검증합니다. 이 테스트는 프로덕션 트래픽이 아닌 CI에서 드리프트를 관측하게 해주는 역할을 합니다.

---

## 7) 조건 변경 시 예측(무엇이 깨지고 어떻게 관측되는가)

제어점 단위로 실패 모드와 관측을 닫습니다.

### 7.1 unknown field를 strict로 실패시키면 무엇이 깨지나

제공자가 응답에 필드를 하나 추가하는 additive change를 했다고 가정하겠습니다. strict reader는 그 즉시 예외를 던지고 요청이 실패합니다.

관측 포인트는 "Unknown field" 같은 예외의 급증과, 그 예외가 특정 엔드포인트/특정 제공자 배포 시점과 강하게 상관되는 형태입니다. 즉 내가 배포하지 않았는데도 장애가 발생할 수 있다는 점이 핵심이며, 이는 운영 비용으로 이어집니다.

### 7.2 unknown field를 무시하지만 관측이 없으면 무엇이 깨지나

unknown을 무시하면 가용성은 지키지만, 제공자가 "필드를 추가하면서 기존 필드의 의미를 바꿨다" 같은 변화가 있으면 조용히 잘못 동작할 수 있습니다. 특히 응답의 의미 판정이 일부 필드 조합에 의존할 때, 이 변화는 도메인 레벨에서만 늦게 관측될 수 있습니다.

관측 포인트는 "업스트림 응답은 200인데 결과 판정이 이상하다" 같은 형태입니다. 즉 원인이 경계 입력에 있는데 증상은 도메인 결과에서만 보입니다. 이 상태가 되면 디버깅은 로그/샘플 수집에 의존하고, 재현이 어려워집니다.

### 7.3 unknown field를 retryable로 분류하면 무엇이 깨지나

unknown field는 대개 제공자의 응답 형태 변화(결정적)입니다. 같은 요청을 다시 보내도 같은 unknown field가 다시 들어올 가능성이 높습니다. 그런데 이를 retryable로 분류하면, 재시도가 지연과 부하를 키우고(특히 동시성 높은 시스템에서는) 실패가 더 넓게 전파될 수 있습니다.

관측 포인트는 "같은 실패가 짧은 시간에 반복"되는 로그 패턴과, p95/p99 지연 상승, 재시도 횟수 증가입니다. 즉 원인을 해결하지 못하는 재시도가 시스템을 더 아프게 만드는 형태입니다.

---

## 8) 코드 예시(나쁜 패턴 1 + 권장 패턴 1 이상)

아래 예시는 JSON 라이브러리 없이, `Map<String, Object>`를 "파싱된 응답"이라고 가정한 최소 예시입니다.

### 8.1 unknown 필드를 즉시 실패 + retryable로 분류하는 나쁜 패턴

```java
import java.util.Map;
import java.util.Set;

final class StrictAllKeys {
    static void validateKnownFields(Map<String, Object> root, Set<String> knownKeys) {
        for (String key : root.keySet()) {
            if (!knownKeys.contains(key)) {
                // unknown을 retryable로 보는 것은 위험할 수 있습니다.
                throw new RuntimeException("Unknown field: " + key);
            }
        }
    }
}
```

이 방식은 제공자의 additive change에 취약하고, unknown이 "결정적 실패"임에도 재시도 루프를 만들기 쉽습니다.

### 8.2 필요한 필드만 parse하고, unknown은 관측하는 권장 패턴

```java
import java.util.Map;
import java.util.Objects;
import java.util.Set;

record ResponseEnvelope(String rc, Map<String, Object> body) {}

final class TolerantParser {
    static ResponseEnvelope parse(Map<String, Object> root, Set<String> knownKeys) {
        Objects.requireNonNull(root, "root");

        // required: rc
        Object rcRaw = root.get("rc");
        if (rcRaw == null || String.valueOf(rcRaw).trim().isBlank()) {
            throw new IllegalArgumentException("missing required field: rc");
        }
        String rc = String.valueOf(rcRaw).trim();

        // optional: body (없으면 빈 맵)
        @SuppressWarnings("unchecked")
        Map<String, Object> body = root.get("body") instanceof Map<?, ?> m ? (Map<String, Object>) m : Map.of();

        // unknown은 실패시키지 않되, 관측 신호는 남깁니다(로그/카운터 등).
        for (String key : root.keySet()) {
            if (!knownKeys.contains(key)) {
                System.out.println("[OBSERVE] unknown response key: " + key);
            }
        }

        return new ResponseEnvelope(rc, body);
    }
}
```

핵심은 "unknown을 무시한다"가 아니라, "도메인에 들어가기 전에 required를 고정하고, unknown 변화는 관측 가능하게 만든다"입니다.

---

## 9) 최소 실험(검증)

아래 실험은 바로 실행 가능한 형태로 고정합니다. 하나는 실패를 관측하도록 설계합니다.

### 9.1 strict unknown-key 검증이 additive change에 취약함을 확인(실패 관측)

```java
import java.util.Map;
import java.util.Set;

public class StrictUnknownKeyLab {
    public static void main(String[] args) {
        Set<String> known = Set.of("rc", "body");
        Map<String, Object> response = Map.of(
            "rc", "0",
            "body", Map.of(),
            "newFieldAddedByProvider", "value"
        );

        try {
            for (String key : response.keySet()) {
                if (!known.contains(key)) {
                    throw new RuntimeException("Unknown field: " + key);
                }
            }
            System.out.println("OK");
        } catch (RuntimeException e) {
            System.out.println("FAILED: " + e.getMessage());
        }
    }
}
```

기대 결과는 `FAILED: Unknown field: newFieldAddedByProvider`입니다. 즉 제공자의 "필드 추가"가 내 런타임 실패로 바뀌는 것을 재현합니다.

### 9.2 tolerant parse는 unknown을 넘기되 required 누락은 fail-fast하는지 확인(성공/실패 모두 관측)

```java
import java.util.Map;
import java.util.Set;

record ResponseEnvelope(String rc, Map<String, Object> body) {}

public class TolerantParseLab {
    static ResponseEnvelope parse(Map<String, Object> root, Set<String> known) {
        Object rcRaw = root.get("rc");
        if (rcRaw == null || String.valueOf(rcRaw).trim().isBlank()) {
            throw new IllegalArgumentException("missing required field: rc");
        }
        for (String key : root.keySet()) {
            if (!known.contains(key)) {
                System.out.println("[OBSERVE] unknown key=" + key);
            }
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> body = root.get("body") instanceof Map<?, ?> m ? (Map<String, Object>) m : Map.of();
        return new ResponseEnvelope(String.valueOf(rcRaw).trim(), body);
    }

    public static void main(String[] args) {
        Set<String> known = Set.of("rc", "body");

        // (1) unknown은 관측만 하고 성공해야 합니다.
        var ok = parse(Map.of("rc", "0", "body", Map.of(), "extra", 1), known);
        System.out.println("OK rc=" + ok.rc());

        // (2) required 누락은 즉시 실패해야 합니다.
        try {
            parse(Map.of("body", Map.of(), "extra", 1), known);
        } catch (IllegalArgumentException e) {
            System.out.println("FAILED: " + e.getMessage());
        }
    }
}
```

기대 결과는 (1) `OK rc=0`이 출력되고, (2) `FAILED: missing required field: rc`가 출력되는 것입니다. 즉 unknown은 가용성을 깨지 않지만, 의미 성립에 필요한 필드는 경계에서 고정한다는 결론을 실험으로 재현합니다.

---

## 10) 트레이드오프/대안/Variants

첫째 트레이드오프는 **가용성** vs **드리프트 탐지 강도**입니다. strict unknown-key 실패는 드리프트를 강하게 드러내지만, 제공자의 additive change까지 내 장애로 만들 수 있습니다. tolerant reader는 가용성을 높이지만, 변화가 조용히 넘어갈 수 있습니다. 따라서 tolerant를 기본값으로 두되, 관측과 테스트로 탐지력을 보강하는 쪽이 균형 잡힌 선택이 되는 경우가 많습니다.

둘째 트레이드오프는 **런타임 검증** vs **계약 테스트**입니다. 런타임 검증은 즉시 방어가 되지만, 잘못 설계하면 재시도/장애를 유발합니다. 계약 테스트는 CI에서 드리프트를 잡아 운영 가용성을 덜 흔들지만, 테스트 데이터/스텁이 최신이 아니면 보호력이 떨어집니다. 따라서 "런타임에서는 required fail-fast + 관측" 정도로 최소한을 하고, 상세 호환성은 계약 테스트로 보강하는 조합이 안정적입니다.

자주 등장하는 Variants는 다음과 같습니다.

Variant A: **silent corruption 비용이 매우 큰 도메인**에서는 일부 엔드포인트/일부 필드에 대해 strict를 선택할 수 있습니다. 다만 이 경우에도 unknown을 무조건 retryable로 두기보다, "알림 + 신속 롤백" 같은 운영 플랜과 함께 가져가야 안전합니다.

Variant B: **업스트림이 버전/스키마 식별자를 제공하는 경우**에는 그 축을 파서 분기의 기준으로 삼는 것이 더 명확합니다. 이때 중요한 것은 버전 분기가 도메인 깊숙이 흩어지지 않게, 경계에서 파서 선택을 끝내는 것입니다.

---

## 11) 흔한 오해

첫째 오해는 "unknown field가 있으면 무조건 스키마가 깨진 것이다"입니다. unknown은 많은 경우 additive change입니다. additive change까지 런타임 실패로 만들면, 내 시스템이 제공자의 진화 속도를 감당하지 못할 수 있습니다. 따라서 unknown은 기본값으로 무시하되, 관측/테스트로 변화 사실을 잡는 편이 현실적입니다.

둘째 오해는 "검증을 인프라에서 많이 하면 도메인이 단순해진다"입니다. 검증이 "validate만 하고 원래 타입을 그대로 흘려보내는 방식"이면, 도메인은 여전히 불법 상태를 방어해야 하고 검증 누락이 다시 발생할 수 있습니다. 도메인이 단순해지려면 validate가 아니라 parse로 구조화된 타입을 만들어야 합니다.

---

## 12) 팀 적용 질문(선택)

첫째, 우리 시스템에서 unknown field의 비용은 무엇입니까? "장애"가 더 비싼지, "조용한 의미 오류"가 더 비싼지 먼저 합의해야 정책이 흔들리지 않습니다.

둘째, required field는 무엇이며, 어디에서 fail-fast해야 합니까? 이 질문에 답하려면 "이 필드가 없으면 어떤 의미가 성립하지 않는가"를 먼저 문장으로 고정해야 합니다.

셋째, 드리프트를 어떻게 조기에 탐지할 것입니까? 관측(로그/카운터)만으로 충분한지, 계약 테스트가 필요한지, 샘플 응답을 어디서 얻고 어떻게 최신성을 유지할지를 결정해야 합니다.

---

## 13) 참고(원전/전문가)

이 카드의 "unknown/required/계약 테스트" 관점은 Martin Fowler의 Tolerant Reader 논의(호환성)와, Alexis King의 "Parse, don't validate"(구조화로 불법 상태 제거), 그리고 consumer-driven contract testing(상호작용을 테스트로 고정) 커뮤니티의 영향을 받았습니다.
