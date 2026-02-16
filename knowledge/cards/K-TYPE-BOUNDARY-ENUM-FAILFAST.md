# K-TYPE-BOUNDARY-ENUM-FAILFAST

## 1. 이 카드가 답하는 질문
이 카드가 답하려는 질문은 다음입니다.

"외부에서 문자열로 들어오는 식별자(mode/profile/operation type)를 내부에서도 계속 `String`으로 들고 가면 왜 위험해지고, 어디에서 어떤 타입(enum/값 객체)으로 바꿔야 안전합니까? 그리고 fail-fast 검증은 정확히 어느 지점에서 어떤 형태로 구현해야 합니까?"

여기서 핵심은 `enum` 문법 자체가 아니라, 장애가 언제/어디서/어떤 비용으로 터지는지를 설계로 통제하는 것입니다.

---

## 2. 적용 범위와 비범위
이 카드의 결론은 아래 상황에서 그대로 적용됩니다.

- 외부 입력(YAML/JSON/HTTP/TCP 등)에서 들어오는 "식별자" 문자열이 내부 분기 로직을 바꾸는 경우
  - 예: `mode=BYPASS|TRANSLATE`, `profile=KPN_200|HECTO_300` 같은 값 후보가 정해진 식별자
- "값이 다르면 동작이 달라지는" 핵심 경로(라우팅/핸들러 선택/정책 적용)에서 사용되는 경우

반대로 아래는 비범위입니다.

- 값 후보가 열려 있는 문자열(open set)
  - 예: 사용자 검색어, 자유 입력 라벨, 외부 시스템의 임의 ID
  - 이 경우는 enum 고정이 과도해지고, 대신 포맷/길이/금지 문자 같은 validator 계약이 중요해집니다.

---

## 3. 기본 결론(기본값)과 선택 이유
기본 결론은 아래 세 문장으로 닫습니다.

외부 입력으로 들어오는 닫힌 집합 식별자는 경계에서 `String -> enum`으로 변환하고, 잘못된 값은 즉시 실패시켜야 합니다.
이 변환 이후 내부 로직은 `String`을 받지 말고 enum만 받도록 API를 고정하는 것이 안전합니다.
그리고 분기 로직은 enum `switch`로 작성하되, 가능하면 `default`를 두지 않아(특히 switch expression) 누락 분기가 컴파일 단계에서 드러나게 해야 합니다.

이 선택이 정당화되는 이유는 "타입이 멋져서"가 아닙니다.
문자열 분기에서 가장 비싼 실패는 오타/누락이 컴파일을 통과해 런타임 깊은 곳에서 터지는 것이고, enum + 경계 변환 + exhaustiveness는 이 실패를 가장 이른 시점(경계/컴파일)으로 끌어올리는 메커니즘이기 때문입니다.

---

## 4. 용어 정의 (이 카드 기준)
여기서 용어를 다음 의미로 고정합니다.

### 4.1 닫힌 집합(Closed Set)
허용 가능한 값 후보가 미리 정해져 있고, 그 외의 값이 들어오면 "새 기능"이 아니라 "오류"로 취급해야 하는 식별자입니다.

- 포함: `mode`, `profile`, `operation type`, 프로토콜 메시지 타입
- 제외: 사용자 입력 문자열, 임의의 외부 ID

닫힌 집합은 "새 값이 추가되면 코드가 반드시 함께 바뀌어야" 합니다.
따라서 코드를 바꾸지 않고 새 값이 조용히 들어오는 것은 안전하지 않습니다.

### 4.2 경계(Boundary)
외부 세계의 자유로운 입력이 내부 도메인 모델로 들어오는 첫 지점입니다.

예:
- 설정 파일(YAML/JSON)을 읽어 자바 객체로 만드는 지점
- HTTP 요청 본문을 DTO로 바인딩하는 지점
- TCP 헤더/전문을 파싱해 내부 모델로 만드는 지점

경계는 "검증 책임"이 모여야 하는 곳입니다.
검증이 경계 밖으로 흩어지면, 잘못된 값이 내부 깊은 곳까지 전파되어 장애가 간헐적이고 추적하기 어려워집니다.

### 4.3 Fail-Fast
잘못된 입력을 "나중"이 아니라 "지금"(경계)에서 실패시키는 전략입니다.

- 입력: raw 문자열
- 변환 규칙: 필수 값/포맷/허용 후보를 검사
- 산출물: 유효하면 enum, 아니면 즉시 실패(예외 또는 검증 에러 수집 후 요청 자체를 거절)

fail-fast의 목표는 "엄격함"이 아니라 "장애 원인 위치를 경계로 고정"하는 것입니다.

### 4.4 분기 완결성(Exhaustiveness)
가능한 모든 경우의 수를 빠짐없이 처리했는지 컴파일 타임에 검증받는 성질입니다.

여기서 중요한 구분이 하나 있습니다.

- `switch statement`는 경우에 따라 누락이 "경고"로 남을 수 있습니다(컴파일이 막히지 않을 수 있습니다).
- `switch expression`은 값을 만들어야 하므로, 모든 경우를 다루지 않으면 컴파일이 막히는 형태로 쓸 수 있습니다.

따라서 "누락을 빌드에서 반드시 잡고 싶다"면, 분기 핵심은 switch expression 형태로 고정하는 편이 안전합니다.

---

## 5. 메커니즘 (입력 -> 변환 규칙 -> 산출물 -> 소비자)
이 절은 "왜 enum이 운영 안정성으로 이어지는지"를 실제 데이터 흐름으로 닫는 부분입니다.

- 입력: 외부에서 들어오는 raw 문자열(예: `profile="KPN_200"`)
- 변환 규칙(경계):
  - `null/blank` 같은 결측을 먼저 잡습니다.
  - 대소문자/공백 같은 표현 차이가 허용되는지 정책을 고정하고(예: `trim().toUpperCase()`), 허용하면 표준화(canonicalize)합니다.
  - 표준화된 값을 enum으로 변환합니다.
- 산출물: 내부 모델은 `String`이 아니라 enum을 갖습니다.
- 소비자: 내부 핵심 로직은 enum `switch`로만 분기합니다.

이 구조의 효과는 "유효하지 않은 값이 내부 핵심 로직까지 들어갈 통로"를 줄이는 것입니다.
통로가 줄면 장애가 늦게 터지는 경우도 줄고, 터지더라도 경계에서 메시지가 고정되므로 디버깅 비용이 크게 줄어듭니다.

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

        // 문제가 여기서부터 시작됩니다.
        // - 오타/누락은 컴파일에서 잡히지 않습니다.
        // - rawProfile이 어디서 왔는지(경계)가 멀어질수록, 원인 추적 비용이 급격히 증가합니다.
        handleUnknown(rawProfile);
    }

    void handleHecto() {}
    void handleKpn() {}
    void handleUnknown(String raw) {}
}
```

이 코드는 "실행은 되지만" 유지보수 비용이 시간이 갈수록 올라갑니다.
프로파일이 하나 추가될 때마다 문자열 비교가 퍼지고, 어느 분기가 빠졌는지 빌드 단계에서 강제하기 어렵습니다.

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

        // Enum.valueOf는 허용 후보가 아니면 IllegalArgumentException을 던집니다.
        // 즉, 여기서 실패하면 "입력이 잘못됐다"가 확정됩니다.
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

여기서 fail-fast는 "예외를 던졌다"가 아니라,
"문제가 있으면 내부 로직으로 들어가기 전에 경계에서 실패했다"라는 성질로 이해하셔야 합니다.

---

## 7. 단점과 트레이드오프 (그리고 왜 그래도 하는가)
이 패턴에도 비용이 있습니다. 이 비용을 모르고 적용하면, 오히려 운영에서 곤란해질 수 있습니다.

첫째, enum은 값 추가가 코드 변경 + 배포를 필요로 합니다.
외부 시스템이 예고 없이 새 값을 보내는 환경에서는, fail-fast가 곧 장애(거절 폭증)로 보일 수 있습니다.

둘째, 경계에서의 표준화 정책(대소문자/공백 허용)은 "편의"가 아니라 "프로토콜 계약"입니다.
팀이 이 정책을 합의하지 않으면, 어떤 입력은 통과하고 어떤 입력은 실패하는 애매한 상태가 됩니다.

셋째, default를 습관적으로 두면 "새 값 추가"가 조용히 흘러가며, 누락 분기를 빌드에서 잡을 기회를 잃습니다.

그럼에도 이 패턴을 기본값으로 두는 정당화는 다음 메커니즘에서 나옵니다.

- 닫힌 집합 식별자는 "모르는 값"을 만나면 안전하게 처리할 방법이 없는 경우가 대부분입니다.
  - 예: 어떤 profile이 들어왔는지에 따라 전문 포맷/라우팅이 달라지는데, profile이 unknown이면 무엇을 보내야 할지 결정 자체가 불가능합니다.
- 따라서 fail-fast로 경계에서 터뜨리는 것이, 내부에서 "대충" 처리하다가 더 큰 정합성 장애를 만드는 것보다 안전합니다.

### 7.1 언제 fail-fast 대신 fail-soft(완화)가 필요한가
외부가 새 값을 보낼 가능성을 기술적으로 막을 수 없고, 거절이 곧 큰 장애가 되는 환경도 있습니다.
그 경우에는 enum에 `UNKNOWN`을 두거나, 파싱 결과를 별도 타입으로 분리하는 방법이 있습니다.

단, 이 선택은 "문제를 숨기는" 쪽으로 흐르기 쉬우므로 관측을 함께 설계해야 합니다.

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

이 경우의 관측 포인트는 간단합니다.
`UNKNOWN`이 발생하면 반드시 로그/메트릭으로 집계되어야 하며, 증가 추세는 곧 "외부 계약 변화"의 신호입니다.

---

## 8. 다르게 판단해야 하는 경우(Variants)
- 값 후보가 열린 집합이면 enum을 쓰지 않습니다.
  - 대신 문자열을 유지하면서 validator를 강화합니다(길이/포맷/금지 문자/정규식 등).

- 값 후보는 닫혀 있지만, 변경 주기가 코드 배포 주기보다 빠르면 enum만으로는 운영이 불가능할 수 있습니다.
  - 이 경우는 레지스트리/설정 기반 매핑 + 계약 테스트로 통제하는 방향을 고려해야 합니다.

- 프로토콜 호환(구버전/신버전 공존)이 필요한 경우에는, fail-fast의 지점을 "요청 거절"이 아니라 "안전한 폴백"으로 설계할 수 있습니다.
  - 단, 폴백이 가능한지(정합성 보장 가능한지)를 먼저 증명해야 합니다.

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
즉, 새 enum 상수가 추가되거나(case가 누락되면), 빌드 단계에서 바로 드러나게 만들 수 있습니다.

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
외부 입력을 그대로 `valueOf`에 넣으면 "표현 차이(대소문자)"와 "결측(null)"에서 바로 실패할 수 있고,
따라서 경계에서 표준화/결측 검사를 먼저 해야 한다는 것이 메커니즘으로 확인됩니다.

---

## 10. 흔한 오해
- 오해: "enum을 쓰면 검증이 자동으로 끝납니다."
  - 정정: 외부 입력은 여전히 문자열입니다. enum은 내부 모델의 안전성을 높이지만, 경계 검증은 별개로 필요합니다.

- 오해: "default를 두면 안전합니다."
  - 정정: default는 새 값 추가/분기 누락을 숨길 수 있습니다. 누락을 빌드에서 잡고 싶다면 default를 두지 않는 방향(특히 switch expression)이 더 안전합니다.

- 오해: "fail-fast는 항상 예외를 던지는 것입니다."
  - 정정: fail-fast의 본질은 "경계에서 실패"입니다. 예외를 던질 수도 있고, 검증 에러를 모아 400/거절로 응답할 수도 있습니다.

---

## 11. 팀 적용 체크리스트
- 이 식별자는 닫힌 집합입니까(허용 후보가 정해져 있습니까)?
- 외부 입력은 경계에서 `String -> enum`으로 변환됩니까?
- 잘못된 값은 내부 로직으로 들어가기 전에 실패합니까(fail-fast)?
- 내부 핵심 로직은 enum-only로 고정되어 있습니까?
- 분기 누락을 컴파일 단계에서 드러내는 구조(switch expression 등)입니까?
