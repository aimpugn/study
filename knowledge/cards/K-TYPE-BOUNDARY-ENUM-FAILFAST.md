# K-TYPE-BOUNDARY-ENUM-FAILFAST

## 1. 이 카드가 답하는 질문
이 카드가 답하려는 질문은 다음입니다.

"외부에서 문자열로 들어오는 식별자(mode/profile/operation type)를 내부에서도 계속 `String`으로 들고 가면 왜 위험해지고, 어디에서 어떤 타입(enum/값 객체)으로 바꿔야 안전합니까? 그리고 fail-fast 검증은 정확히 어느 지점에서 어떤 형태로 구현해야 합니까?"

여기서 핵심은 `enum` 문법 자체가 아니라, 장애를 언제 발견하고(시점), 어디에서 원인을 고정하고(범위), 어느 정도 비용으로 복구할지(운영 비용)를 설계로 통제하는 것입니다.

---

## 2. 적용 범위와 비범위
이 카드의 결론은 외부 입력(YAML/JSON/HTTP/TCP 등)에서 들어오는 문자열이 내부 분기 로직을 바꾸는 경우에 적용됩니다.
예를 들어 `mode=BYPASS|TRANSLATE`, `profile=KPN_200|HECTO_300`처럼 값 후보가 정해진 식별자가 있고, 그 값에 따라 라우팅/핸들러 선택/정책 적용 같은 핵심 경로가 달라지는 상황이 이에 해당합니다.

반대로 값 후보가 열린 집합(open set)인 문자열에는 그대로 적용하지 않습니다.
사용자 검색어, 자유 입력 라벨, 외부 시스템의 임의 ID처럼 "새 값이 계속 나올 수밖에 없는" 데이터는 enum으로 고정하는 순간 운영 자체가 불가능해집니다.
이 영역은 타입화가 아니라 포맷/길이/금지 문자 같은 validator 계약이 중심이 되어야 합니다.

---

## 3. 기본 결론(기본값)과 선택 이유
이 카드의 기본 결론은 세 가지가 한 묶음으로 움직입니다.

첫째, 외부 입력으로 들어오는 닫힌 집합 식별자는 경계에서 `String -> enum`으로 변환하고, 잘못된 값은 즉시 실패(fail-fast)시켜야 합니다.
둘째, 변환 이후 내부 로직은 `String`을 받지 말고 enum만 받도록 API를 고정해야 합니다. 이렇게 해야 문자열 비교가 내부에 퍼지지 않고, "검증이 끝난 값"만 핵심 로직으로 들어가게 됩니다.
셋째, 분기 로직은 enum `switch`로 작성하되, 가능하면 `default`를 두지 않아 누락 분기가 빌드 단계에서 드러나게 만들어야 합니다. 특히 값을 반환하는 switch expression 형태로 작성하면, case 누락을 컴파일 단계에서 강제하기가 더 쉽습니다.

이 선택이 정당화되는 이유는 단순합니다.
문자열 분기에서 가장 비싼 실패는 오타/누락이 컴파일을 통과해 런타임 깊은 곳에서 터지는 것이고, 그때 관측되는 증상(NPE, 잘못된 라우팅, 알 수 없는 정책 적용)은 원인(입력 값)과 멀어지는 경향이 있습니다.
경계 변환 + enum-only + exhaustiveness는 이 실패를 가장 이른 시점(경계/컴파일)으로 끌어올려, 원인 위치를 고정하고 복구 비용을 낮추는 메커니즘입니다.

---

## 4. 용어 정의 (이 카드 기준)
여기서 용어를 다음 의미로 고정합니다.

### 4.1 닫힌 집합(Closed Set)
허용 가능한 값 후보가 미리 정해져 있고, 그 외의 값이 들어오면 "새 기능"이 아니라 "오류"로 취급해야 하는 식별자입니다.
이 카드에서 닫힌 집합에는 `mode`, `profile`, `operation type`, 프로토콜 메시지 타입 같은 분기 키가 포함됩니다.
반대로 사용자 입력 문자열, 임의의 외부 ID처럼 값 후보가 열려 있는 데이터는 닫힌 집합에서 제외합니다.

닫힌 집합의 본질은 "새 값이 추가되면 코드도 함께 바뀌어야 한다"는 점입니다.
따라서 코드를 바꾸지 않았는데도 새 값이 조용히 들어와 통과하는 것은, 기능 확장이 아니라 정합성 붕괴의 시작점이 됩니다.

### 4.2 경계(Boundary)
외부 세계의 자유로운 입력이 내부 도메인 모델로 들어오는 첫 지점입니다.
설정 파일(YAML/JSON)을 읽어 자바 객체로 만드는 지점, HTTP 요청 본문을 DTO로 바인딩하는 지점, TCP 헤더/전문을 파싱해 내부 모델로 만드는 지점이 모두 경계입니다.

경계는 검증 책임이 모여야 하는 곳입니다.
검증이 경계 밖으로 흩어지면, 잘못된 값이 내부 깊은 곳까지 전파되어 장애가 간헐적이고 추적하기 어려워집니다.

### 4.3 Fail-Fast
잘못된 입력을 "나중"이 아니라 "지금"(경계)에서 실패시키는 전략입니다.
fail-fast의 목표는 엄격함 자체가 아니라, 장애 원인 위치를 경계로 고정해 디버깅 범위를 줄이는 것입니다.

### 4.4 분기 완결성(Exhaustiveness)
가능한 모든 경우의 수를 빠짐없이 처리했는지 컴파일 타임에 검증받는 성질입니다.
이 카드에서 분기 완결성은 "새 enum 상수가 추가되었을 때, 분기 누락이 빌드에서 바로 드러나는가"라는 운영 의미로 사용합니다.

---

## 5. 메커니즘 (입력 -> 변환 규칙 -> 산출물 -> 소비자)
여기서는 "enum을 쓰면 좋다"가 아니라, 실제로 시스템이 어떻게 안정성을 얻는지 데이터 흐름으로 닫습니다.

입력은 외부에서 들어오는 raw 문자열입니다. 예를 들어 `profile="KPN_200"` 같은 값은 처음에는 단지 텍스트일 뿐이고, 그 텍스트가 유효한지 여부는 아직 확정되지 않았습니다.

경계에서의 변환 규칙은 다음 세 단계를 묶어 구현하는 것이 안전합니다.
먼저 `null/blank` 같은 결측을 잡고, 그 다음 대소문자/공백 같은 표현 차이를 허용할지 정책을 고정합니다. 정책이 허용이라면 `trim().toUpperCase()` 같은 표준화를 적용해 "표현"을 하나로 모읍니다.
마지막으로 표준화된 값을 enum으로 변환합니다. 이 변환이 실패하면 그 시점에서 "입력이 잘못됐다"가 확정되므로, 내부 로직으로 들어가기 전에 실패를 고정할 수 있습니다.

산출물은 enum입니다. 이때 중요한 설계는 내부 API가 enum만 받도록 고정하는 것입니다.
내부가 여전히 `String`을 받으면, 문자열 비교가 퍼지고, 어떤 경로는 검증을 거치고 어떤 경로는 검증을 건너뛰는 상태가 되어, 결국 fail-fast의 효과가 사라집니다.

소비자는 내부 분기 로직입니다.
분기 로직이 enum `switch`로 작성되어 있으면, 새 값이 추가될 때 누락 분기를 빌드에서 드러내는 형태(특히 switch expression)를 설계할 수 있습니다.

---

## 6. 코드 패턴 예시
아래 예시는 프로젝트에 종속되지 않도록, 카드 안에서 학습 가능한 최소 단위로 구성했습니다.

### 6.1 안 좋은 패턴: 내부까지 String이 전파되고 분기가 흩어집니다
```java
class BadRouter {
    void route(String rawProfile) {
        if ("HECTO_300".equals(rawProfile)) {
            handleHecto();
            return;
        }
        if ("KPN_200".equals(rawProfile)) {
            handleKpn();
            return;
        }

        // 오타/누락은 컴파일에서 잡히지 않습니다.
        // rawProfile의 원인 위치(경계)가 멀어질수록, 장애 추적 비용이 급격히 증가합니다.
        handleUnknown(rawProfile);
    }

    void handleHecto() {}
    void handleKpn() {}
    void handleUnknown(String raw) {}
}
```

이 코드는 실행은 되지만, 시간이 지날수록 유지보수 비용이 올라갑니다.
프로파일이 하나 추가될 때마다 문자열 비교가 퍼지고, 어느 분기가 빠졌는지 빌드 단계에서 강제하기 어렵기 때문입니다.

### 6.2 권장 패턴: 경계에서 enum으로 변환하고, 내부는 enum-only로 고정합니다
```java
enum InboundProfile {
    HECTO_300,
    KPN_200;

    static InboundProfile fromExternal(String raw) {
        // 경계에서 결측을 먼저 잡습니다.
        if (raw == null || raw.isBlank()) {
            throw new IllegalArgumentException("inbound-profile is required");
        }

        // 표현 차이를 허용하는 정책이라면 여기서 표준화합니다.
        // (정책이 "대소문자 구분"이면 toUpperCase는 하지 말아야 합니다.)
        String normalized = raw.trim().toUpperCase();

        // 허용 후보가 아니면 IllegalArgumentException으로 실패합니다.
        return InboundProfile.valueOf(normalized);
    }
}

class GoodRouter {
    void route(String rawProfile) {
        InboundProfile profile = InboundProfile.fromExternal(rawProfile);

        // 내부는 enum-only: 더 이상 String 비교가 퍼지지 않습니다.
        switch (profile) {
            case HECTO_300 -> handleHecto();
            case KPN_200 -> handleKpn();
        }
    }

    void handleHecto() {}
    void handleKpn() {}
}
```

여기서 fail-fast는 "예외를 던졌다"가 아니라, "문제가 있으면 내부 로직으로 들어가기 전에 경계에서 실패했다"라는 성질로 이해하셔야 합니다.

---

## 7. 단점과 트레이드오프 (그리고 왜 그래도 하는가)
이 패턴에도 비용이 있습니다. 이 비용을 모르고 적용하면, 오히려 운영에서 곤란해질 수 있습니다.

가장 먼저 부딪히는 비용은, enum이 값 추가를 코드 변경 + 배포로 묶는다는 점입니다.
외부 시스템이 예고 없이 새 값을 보내는 환경에서는, fail-fast가 곧 거절 폭증으로 보일 수 있습니다.

다음 비용은 표준화 정책(대소문자/공백 허용)이 "편의"가 아니라 "프로토콜 계약"이라는 점입니다.
팀이 이 정책을 합의하지 않으면 어떤 입력은 통과하고 어떤 입력은 실패하는 애매한 상태가 되고, 그 애매함은 결국 운영 장애로 돌아옵니다.

또 한 가지는 default의 유혹입니다.
default를 습관적으로 두면 "새 값 추가"가 조용히 흘러가며, 누락 분기를 빌드에서 잡을 기회를 잃습니다.

그럼에도 이 패턴을 기본값으로 두는 정당화는 "닫힌 집합"이 갖는 성질에서 나옵니다.
닫힌 집합 식별자는 모르는 값을 만났을 때 안전하게 처리할 방법이 없는 경우가 대부분입니다.
예를 들어 어떤 profile이 들어왔는지에 따라 전문 포맷/라우팅이 달라지는데 profile이 unknown이면 무엇을 보내야 할지 결정 자체가 불가능합니다.
이때 내부에서 "대충" 처리하면 더 큰 정합성 장애를 만들 수 있으므로, 경계에서 실패시키는 것이 오히려 안전합니다.

### 7.1 언제 fail-fast 대신 fail-soft(완화)가 필요한가
외부가 새 값을 보낼 가능성을 기술적으로 막을 수 없고, 거절이 곧 큰 장애가 되는 환경도 있습니다.
그 경우에는 enum에 `UNKNOWN`을 두거나, 파싱 결과를 별도 타입으로 분리하는 방법이 있습니다.

다만 이 선택은 문제를 숨기기 쉬우므로, 관측을 함께 설계해야 합니다.
`UNKNOWN`이 발생하면 반드시 로그/메트릭으로 집계되어야 하고, 증가 추세는 곧 외부 계약 변화의 신호가 되어야 합니다.

```java
enum Mode {
    BYPASS,
    TRANSLATE,
    UNKNOWN;

    static Mode tryParse(String raw) {
        if (raw == null || raw.isBlank()) return UNKNOWN;
        try {
            return Mode.valueOf(raw.trim().toUpperCase());
        } catch (IllegalArgumentException e) {
            return UNKNOWN;
        }
    }
}
```

---

## 8. 다르게 판단해야 하는 경우(Variants)
기본 결론이 항상 정답이 되지는 않습니다. 어떤 전제가 바뀌면 결론도 바뀝니다.

값 후보가 열린 집합이면 enum을 쓰지 않습니다.
이 경우에는 문자열을 유지하면서 validator를 강화하는 편이 맞습니다. 여기서의 안정성은 타입이 아니라 검증 규칙(길이/포맷/금지 문자/정규식 등)에서 나옵니다.

값 후보는 닫혀 있지만 변경 주기가 코드 배포 주기보다 빠르면, enum만으로 운영이 불가능할 수 있습니다.
이 경우는 레지스트리/설정 기반 매핑 + 계약 테스트로 통제하는 방향을 고려해야 합니다.

프로토콜 호환(구버전/신버전 공존)이 필요한 경우에는, fail-fast의 지점을 "요청 거절"이 아니라 "안전한 폴백"으로 설계할 수 있습니다.
단, 폴백이 가능한지(정합성 보장 가능한지)를 먼저 증명해야 합니다.

---

## 9. 최소 실험(검증): 이 결론을 직접 확인하기
이 카드의 핵심 주장(컴파일 타임 누락 검출, 경계 fail-fast)을 작은 실험으로 확인해 보겠습니다.

### 9.1 switch expression이 누락 분기를 컴파일 단계에서 막는지 확인
아래 코드는 일부러 `TRANSLATE` 분기를 빼 둔 예시입니다.

```java
enum Mode { BYPASS, TRANSLATE }

class ExhaustiveLab {
    static int handle(Mode mode) {
        // switch expression은 모든 값을 처리해야 값을 만들 수 있습니다.
        // 따라서 case 하나라도 빠지면 컴파일이 막히는 형태로 만들 수 있습니다.
        return switch (mode) {
            case BYPASS -> 0;
            // case TRANSLATE -> 1;  // 일부러 누락
        };
    }
}
```

기대 결과는 "컴파일 실패"입니다.
즉 새 enum 상수가 추가되거나(case가 누락되면) 빌드 단계에서 바로 드러나게 만들 수 있습니다.

### 9.2 Enum.valueOf가 어떤 입력에서 어떻게 실패하는지 확인
아래 코드는 `valueOf`의 실패 형태를 관측하기 위한 실험입니다.

```java
enum Mode { BYPASS, TRANSLATE }

public class ValueOfLab {
    public static void main(String[] args) {
        try {
            System.out.println(Mode.valueOf("BYPASS"));
        } catch (Exception e) {
            System.out.println("unexpected: " + e);
        }

        try {
            System.out.println(Mode.valueOf("bypass"));
        } catch (Exception e) {
            System.out.println(e.getClass().getSimpleName());
            // 기대 결과: IllegalArgumentException
        }

        try {
            System.out.println(Mode.valueOf(null));
        } catch (Exception e) {
            System.out.println(e.getClass().getSimpleName());
            // 기대 결과: NullPointerException
        }
    }
}
```

이 실험이 알려주는 결론은 간단합니다.
외부 입력을 그대로 `valueOf`에 넣으면 표현 차이(대소문자)와 결측(null)에서 바로 실패할 수 있습니다.
따라서 경계에서 표준화/결측 검사를 먼저 해야 한다는 것이 메커니즘으로 확인됩니다.

---

## 10. 흔한 오해
오해는 대부분 "어느 책임이 어디에 있는지"가 섞일 때 생깁니다.
따라서 오해를 바로잡을 때는, 타입(enum), 검증(경계), 분기(exhaustiveness)가 각각 무엇을 해결하고 무엇을 해결하지 못하는지를 분리해 닫아야 합니다.

"enum을 쓰면 검증이 자동으로 끝난다"는 오해가 대표적입니다.
외부 입력은 여전히 문자열이므로, enum은 내부 모델의 안전성을 높이지만 경계 검증은 별개로 필요합니다.

"default를 두면 안전하다"는 오해도 자주 나옵니다.
default는 새 값 추가/분기 누락을 숨길 수 있습니다. 누락을 빌드에서 잡고 싶다면 default를 두지 않는 방향(특히 switch expression)이 더 안전합니다.

"fail-fast는 항상 예외를 던진다"는 오해도 있습니다.
fail-fast의 본질은 경계에서 실패시키는 것이고, 예외를 던질 수도 있으며 검증 에러를 모아 요청을 거절하는 방식으로도 구현할 수 있습니다.

---

## 11. 팀 적용 체크리스트
마지막으로 팀이 같은 논리로 판단할 수 있도록 질문을 남깁니다.

- 이 식별자는 닫힌 집합입니까(허용 후보가 정해져 있습니까)?
- 외부 입력은 경계에서 `String -> enum`으로 변환됩니까?
- 잘못된 값은 내부 로직으로 들어가기 전에 실패합니까(fail-fast)?
- 내부 핵심 로직은 enum-only로 고정되어 있습니까?
- 분기 누락을 컴파일 단계에서 드러내는 구조(switch expression 등)입니까?
