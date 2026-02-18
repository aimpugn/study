# K-VERTICAL-SLICE-POLICY-COMPOSITION-READABILITY

## 1) 이 카드가 답하는 질문

코드베이스가 커지면 아래 질문이 반복됩니다.

"Hexagonal + DDD로 시작했는데, 시간이 지나면서 패키지/클래스/메서드가 일관되지 않고 정책 객체가 늘어나 '조합(composition)'이 복잡해졌습니다. 이때 어떻게 구조를 더 flat하게 만들고, 한 눈에 처리 흐름이 보이는 코드로 정리할 수 있습니까? 유스케이스 전용 정책 클래스(예: 특정 작업만 라우팅하는 Policy)는 언제 괜찮고, 언제는 일반화/재사용 방향으로 바꿔야 합니까?"

여기서 핵심은 디자인 패턴 이름을 붙이는 것이 아니라, 독자가 "이 요청이 어디를 지나 어떤 정책이 언제 적용되는지"를 같은 논리로 재구성할 수 있게 만드는 것입니다.

---

## 2) 배경: 왜 이 문제가 실무에서 비싸게 터지는가

가독성 저하는 취향 문제가 아니라 운영 비용으로 연결됩니다. 처리 흐름이 한 눈에 보이지 않으면, 장애가 났을 때 "정책이 어디서 적용되었는지"를 복원하기 어렵고, 수정은 더 많은 추측을 포함하게 됩니다. 특히 금융/연동 시스템에서는 timeout/retry/멱등성/failover 같은 정책이 곧 안전장치이기 때문에, "언제 무엇이 적용되는지"가 불명확하면 회귀 위험이 큽니다.

복잡한 composition은 보통 두 가지 방향에서 생깁니다.

첫째, 공통 관심사(cross-cutting concern: timeout/retry/관측/보안)를 유스케이스마다 따로 구현해 중복이 늘고, 그 중복을 줄이기 위해 정책 타입을 계속 추가하면서 조합이 복잡해집니다.

둘째, "일반화"를 성급히 하면서, 서로 다른 유스케이스의 규칙을 하나의 추상화로 묶어버립니다. 이 경우 추상화는 커지는데, 정작 읽을 때는 "이 경우에 어떤 분기가 선택되는지"를 따라가야 해서 인지 복잡도는 더 커질 수 있습니다.

따라서 정리는 "정책 클래스를 없애자"가 아니라, **유스케이스의 실행 스토리(순서)**를 코드에서 드러내고, 공통화는 **실행 골격(skeleton)**까지만 하며, 유스케이스별 의미는 **슬라이스 안에 남겨** 변화 반경을 줄이는 쪽이 실무적으로 안전한 경우가 많습니다.

---

## 3) 핵심 제어점(왜 하필 이 축인가)

이 주제는 범위가 넓으므로, 판단을 고정하는 제어점 3개로 축을 좁힙니다.

첫째는 **코드의 조직 기준이 무엇인가(레이어 vs 유스케이스/요청)**입니다. 레이어 기준 패키징은 같은 기술(컨트롤러/서비스/리포지토리)을 모으지만, "하나의 요청"을 이해하려면 여러 패키지를 왕복해야 합니다. 유스케이스 기준(Vertical Slice)은 한 요청의 처리에 필요한 것들을 가까이 둬서, 변경 반경과 독해 비용을 줄이는 방향입니다.

둘째는 **공통화의 단위가 무엇인가(의미 vs 실행 골격)**입니다. timeout/retry/관측 같은 것은 많은 유스케이스에 공통이지만, "어떤 실패에서 retry하는가", "failover를 언제 허용하는가" 같은 의미는 유스케이스마다 다를 수 있습니다. 공통화는 실행 골격까지로 제한하고, 의미는 유스케이스에 남겨야 잘못된 추상화 비용을 줄일 수 있습니다.

셋째는 **정책 적용 순서가 코드에서 명시적으로 드러나는가**입니다. 정책이 인터셉터/데코레이터/DI 조합으로 흩어지면, 독자는 "정책이 언제 적용되는지"를 머릿속에서 합성해야 합니다. 반대로 유스케이스 메서드가 "resolve -> guard -> execute(with policy) -> normalize -> observe"처럼 읽히면, 디버깅과 리뷰가 쉬워집니다.

---

## 4) 기본 결론(기본값)과 선택 이유

기본 결론은 다음처럼 닫습니다.

패키지/구조를 더 flat하고 읽기 쉽게 만들려면, "레이어별로 모으기"보다 "유스케이스(요청)별로 모으기"를 기본값으로 두고, 공통화는 "실행 골격"까지만 하십시오. 그리고 정책의 핵심은 "무엇이 있느냐"가 아니라 "언제 적용되느냐"이므로, 정책 적용 순서를 유스케이스 코드에서 명시적으로 드러내는 형태(파이프라인/러너)를 기본값으로 두는 편이 안전합니다.

이 결론은 다음 비용 구조를 줄입니다.

유스케이스 중심(Vertical Slice)은 한 요청을 이해할 때 필요한 파일 이동을 줄이고, 변경이 한 슬라이스에 머무를 확률을 올립니다. Jimmy Bogard가 말하는 Vertical Slice의 핵심 메시지도 "요청을 중심으로 코드를 조직하면, 기능 변경의 조합 폭발을 줄일 수 있다"에 가깝습니다.

또한 실행 골격 공통화(예: timeout+retry+관측을 수행하는 실행기)는 중복을 줄이되, 의미 차이(어떤 실패를 retry하는가)는 유스케이스에 남길 수 있어 잘못된 추상화 비용을 줄입니다.

---

## 5) 핵심 용어 정의(이 문서 맥락에서)

**Vertical Slice(유스케이스 중심 구조)**는 "컨트롤러/서비스/리포지토리" 같은 기술 레이어가 아니라, "하나의 요청(유스케이스)" 단위로 코드를 모으는 조직 방식입니다. 목적은 재사용이 아니라, 독해/변경 반경을 줄이는 것입니다.

**실행 골격(skeleton)**은 여러 유스케이스가 공유할 수 있는 "실행 프레임"입니다. 예: `timeout -> retry -> attempt 기록 -> 실패 분류` 같은 구조는 골격이 될 수 있지만, 어떤 예외가 retryable인지 같은 규칙은 유스케이스 의미에 가깝습니다.

**정책(policy)**은 값 객체(설정)일 수도 있고, 실행기(로직)일 수도 있습니다. 이 카드에서 중요한 구분은 "정책이 존재한다"가 아니라 "정책이 어디에서 적용되는지"입니다.

---

## 6) 메커니즘(입력 -> 변환 -> 산출물 -> 소비자)

유스케이스 실행을 한 흐름으로 복원할 수 있어야, 복잡한 composition을 줄일 수 있습니다.

입력은 "유스케이스 요청"입니다(예: 이체/조회/등록 요청). 변환은 (1) 대상 결정(resolve: 라우팅/스티키/기본값), (2) 단일 실행 보호(guard: 멱등성/중복 방지), (3) 정책 실행(execute: timeout/retry/관측), (4) 결과 정규화(normalize), (5) 관측(로그/메트릭/감사)입니다. 산출물은 유스케이스 결과이며, 소비자는 inbound(HTTP/TCP/CLI) 또는 후속 유스케이스입니다.

복잡도가 올라가는 지점은 보통 (1)~(3)입니다. 이때 핵심은 "정책 타입을 더 만든다"가 아니라, (1)~(3)의 순서를 **코드 한 곳에서 읽히게** 만드는 것입니다. 순서가 보이면, "이 정책이 여기서 적용되니, 이 실패는 여기서 관측된다"가 연결됩니다.

---

## 7) 조건 변경 시 예측(무엇이 깨지고 어떻게 관측되는가)

제어점 단위로 실패 모드와 관측을 닫습니다.

### 7.1 레이어 중심 구조를 유지하면 무엇이 깨지기 쉬운가

레이어 기준으로 패키징하면, 유스케이스 변경이 컨트롤러/서비스/리포지토리를 동시에 건드리기 쉽고, 공통 로직을 서비스 레이어에 누적시키기 쉽습니다. 시간이 지나면 "서비스"가 의미 없는 집합이 되고, 특정 기능을 이해하려면 여러 파일을 왕복해야 합니다.

관측 포인트는 "작은 기능 변경인데 수정 파일이 과도하게 많다"입니다. 예를 들어 한 정책을 바꾸는데 컨트롤러/서비스/리포지토리/유틸이 동시에 바뀐다면, 코드 조직 축이 변경 축과 맞지 않을 가능성이 큽니다.

### 7.2 공통화를 의미까지 해버리면 무엇이 깨지나

예를 들어 "모든 작업은 동일한 failover/retry 규칙을 가진다" 같은 추상화로 통일하면, 어떤 작업은 안전하지만 어떤 작업은 중복 실행 위험이 커질 수 있습니다.

관측 포인트는 "정책 변경이 예상하지 못한 유스케이스까지 함께 바꾼다"입니다. 즉 변경 영향이 기능 경계를 넘어 전파됩니다. 이때 장애는 특정 작업에서만 터지고, 원인은 공유 코드에 있습니다.

### 7.3 정책 적용 순서가 숨겨지면 무엇이 깨지나

정책이 데코레이터/인터셉터 체인으로 숨겨지면, "timeout이 먼저냐, 멱등성이 먼저냐" 같은 질문에 코드가 답을 주지 못합니다. 이 상태에서는 리뷰/디버깅이 사람의 기억과 추측에 의존합니다.

관측 포인트는 장애 분석 시 "원인 위치"가 계속 바뀌는 것입니다. 같은 유형의 실패가 어떤 때는 timeout으로, 어떤 때는 guard로, 어떤 때는 라우팅으로 보이는 식의 혼선이 생기면, 정책 순서가 명시적으로 고정돼 있지 않을 가능성이 큽니다.

---

## 8) 코드 예시(나쁜 패턴 1 + 권장 패턴 1 이상)

### 8.1 정책 적용이 분산되어 순서가 보이지 않는 나쁜 패턴

```java
import java.time.Duration;
import java.util.function.Supplier;

final class BadUseCase {
    private final TimeoutRunner timeout;
    private final RetryRunner retry;
    private final IdempotencyGuard guard;
    private final Router router;

    BadUseCase(TimeoutRunner timeout, RetryRunner retry, IdempotencyGuard guard, Router router) {
        this.timeout = timeout;
        this.retry = retry;
        this.guard = guard;
        this.router = router;
    }

    String handle(String input) {
        // 순서가 여기저기 섞이고, 일부 분기는 다른 경로를 탈 수 있습니다.
        String target = router.resolve(input);
        return guard.executeOnce(input, () ->
            retry.run(() ->
                timeout.run(Duration.ofSeconds(1), () -> callUpstream(target))
            )
        );
    }

    private String callUpstream(String target) { return "ok:" + target; }
}

interface TimeoutRunner { <T> T run(Duration d, Supplier<T> action); }
interface RetryRunner { <T> T run(Supplier<T> action); }
interface IdempotencyGuard { <T> T executeOnce(String key, Supplier<T> action); }
interface Router { String resolve(String key); }
```

이 코드는 동작할 수 있지만, "어떤 정책이 먼저 적용되는지"를 한 눈에 파악하기 어렵습니다. 특히 훅(관측/쿨다운/스티키 매핑)이 끼어들면 더 복잡해집니다.

### 8.2 실행 골격을 파이프라인으로 고정하는 권장 패턴

```java
import java.time.Duration;
import java.util.Objects;
import java.util.function.Supplier;

final class OperationPipeline {
    private final TimeoutRunner timeout;
    private final RetryRunner retry;
    private final IdempotencyGuard guard;

    OperationPipeline(TimeoutRunner timeout, RetryRunner retry, IdempotencyGuard guard) {
        this.timeout = Objects.requireNonNull(timeout);
        this.retry = Objects.requireNonNull(retry);
        this.guard = Objects.requireNonNull(guard);
    }

    <T> T run(OperationContext ctx, Supplier<T> action) {
        // "순서"가 여기에서 고정됩니다.
        return guard.executeOnce(ctx.idempotencyKey(), () ->
            retry.run(() ->
                timeout.run(ctx.timeout(), action)
            )
        );
    }
}

record OperationContext(String idempotencyKey, Duration timeout) {}

final class GoodUseCase {
    private final Router router;
    private final OperationPipeline pipeline;

    GoodUseCase(Router router, OperationPipeline pipeline) {
        this.router = router;
        this.pipeline = pipeline;
    }

    String handle(String input) {
        String target = router.resolve(input);
        OperationContext ctx = new OperationContext("op:" + input, Duration.ofSeconds(1));
        return pipeline.run(ctx, () -> callUpstream(target));
    }

    private String callUpstream(String target) { return "ok:" + target; }
}

interface TimeoutRunner { <T> T run(Duration d, Supplier<T> action); }
interface RetryRunner { <T> T run(Supplier<T> action); }
interface IdempotencyGuard { <T> T executeOnce(String key, Supplier<T> action); }
interface Router { String resolve(String key); }
```

핵심은 "정책이 더 많다/적다"가 아니라, **정책 적용 순서가 파이프라인 코드로 고정되어 독자가 추론을 재현할 수 있다**는 점입니다.

---

## 9) 최소 실험(검증)

### 9.1 파이프라인이 호출 순서를 고정하는지 확인(성공/실패 관측)

```java
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Supplier;

public class PipelineOrderLab {
    interface TimeoutRunner { <T> T run(Duration d, Supplier<T> action); }
    interface RetryRunner { <T> T run(Supplier<T> action); }
    interface IdempotencyGuard { <T> T executeOnce(String key, Supplier<T> action); }

    static final class RecordingTimeout implements TimeoutRunner {
        private final List<String> events;
        RecordingTimeout(List<String> events) { this.events = events; }
        public <T> T run(Duration d, Supplier<T> action) {
            events.add("timeout.start");
            try { return action.get(); } finally { events.add("timeout.end"); }
        }
    }

    static final class RecordingRetry implements RetryRunner {
        private final List<String> events;
        RecordingRetry(List<String> events) { this.events = events; }
        public <T> T run(Supplier<T> action) {
            events.add("retry.start");
            try { return action.get(); } finally { events.add("retry.end"); }
        }
    }

    static final class RecordingGuard implements IdempotencyGuard {
        private final List<String> events;
        RecordingGuard(List<String> events) { this.events = events; }
        public <T> T executeOnce(String key, Supplier<T> action) {
            events.add("guard.start:" + key);
            try { return action.get(); } finally { events.add("guard.end"); }
        }
    }

    static final class Pipeline {
        private final TimeoutRunner timeout;
        private final RetryRunner retry;
        private final IdempotencyGuard guard;
        Pipeline(TimeoutRunner timeout, RetryRunner retry, IdempotencyGuard guard) {
            this.timeout = timeout; this.retry = retry; this.guard = guard;
        }
        <T> T run(String key, Supplier<T> action) {
            return guard.executeOnce(key, () -> retry.run(() -> timeout.run(Duration.ofMillis(1), action)));
        }
    }

    public static void main(String[] args) {
        List<String> events = new ArrayList<>();
        Pipeline p = new Pipeline(new RecordingTimeout(events), new RecordingRetry(events), new RecordingGuard(events));

        p.run("K", () -> {
            events.add("action");
            return "ok";
        });

        System.out.println(String.join(" -> ", events));
    }
}
```

기대 결과는 이벤트 순서가 항상 동일하다는 것입니다. 예: `guard.start:K -> retry.start -> timeout.start -> action -> timeout.end -> retry.end -> guard.end`. 즉 파이프라인이 "정책 적용 순서"를 코드로 고정함을 재현합니다.

### 9.2 유스케이스 전용 전략이 누락되면 fail-fast하는지 확인(실패 관측)

```java
import java.util.EnumMap;
import java.util.Objects;

public class StrategyRegistryLab {
    enum Operation { DEPOSIT, WITHDRAWAL }
    interface Strategy { String run(); }

    static final class Registry {
        private final EnumMap<Operation, Strategy> map = new EnumMap<>(Operation.class);

        Registry() {
            map.put(Operation.DEPOSIT, () -> "deposit");
            // WITHDRAWAL을 일부러 누락
        }

        Strategy getRequired(Operation op) {
            return Objects.requireNonNull(map.get(op), "missing strategy: " + op);
        }
    }

    public static void main(String[] args) {
        Registry r = new Registry();
        System.out.println(r.getRequired(Operation.DEPOSIT).run());
        try {
            System.out.println(r.getRequired(Operation.WITHDRAWAL).run());
        } catch (NullPointerException e) {
            System.out.println("FAILED: " + e.getMessage());
        }
    }
}
```

기대 결과는 `FAILED: missing strategy: WITHDRAWAL`입니다. 이 실험은 "유스케이스 전용 정책/전략은 괜찮지만, 그 누락이 조용히 통과하면 위험하므로 fail-fast로 고정해야 한다"는 판단을 검증합니다.

---

## 10) 트레이드오프/대안/Variants

첫째 트레이드오프는 **유스케이스 중심 구조의 중복** vs **레이어 중심 구조의 재사용**입니다. Vertical Slice는 슬라이스 간 코드 공유를 줄이기 때문에, 일부 중복이 생길 수 있습니다. 대신 그 중복은 "기능 변화가 다른 것"을 억지로 묶지 않는 비용이며, 인지 복잡도와 변경 반경을 줄여줍니다. 반대로 레이어 중심 구조는 재사용이 쉬워 보이지만, 시간이 지나면 공통 레이어가 비대해지고 기능 간 결합이 증가할 수 있습니다.

둘째 트레이드오프는 **파이프라인을 코드에 드러내기(명시성)** vs **프레임워크/인터셉터로 숨기기(추상화)**입니다. 인터셉터는 공통화에 유리하지만, 순서가 숨겨져 디버깅/리뷰 비용이 올라갈 수 있습니다. 파이프라인은 명시성이 높지만, 코드가 길어질 수 있습니다. 기본값은 안전과 관측을 우선해 "핵심 유스케이스는 파이프라인으로 명시"하는 편이 보수적입니다.

Variant A: 팀이 레이어 단위로 나뉘어 있고 ownership이 강한 조직에서는, vertical slice를 한 번에 적용하면 충돌이 큽니다. 이때는 "가장 위험/핵심 유스케이스"부터 슬라이스로 묶고, 나머지는 레이어 구조를 유지하는 점진 적용이 현실적입니다.

Variant B: 정책이 매우 많고 순서/조건이 복잡해지는 경우, 파이프라인을 더 얇게 만들기 위해 "정책 표(설정) + 실행기" 조합을 쓸 수 있습니다. 다만 이 경우에도 실행 순서가 문서/코드로 복원 가능해야 합니다.

---

## 11) 흔한 오해

첫째 오해는 "유스케이스 전용 정책 클래스는 재사용성이 없으니 나쁘다"입니다. 유스케이스 전용 정책은 오히려 "의미가 다른 규칙을 섞지 않는다"는 점에서 안전할 수 있습니다. 문제는 전용 정책이 아니라, 전용 정책의 누락/정합성 붕괴가 조용히 통과하는 것입니다. 그래서 전략 레지스트리/기본값/테스트로 fail-fast를 고정하는 것이 중요합니다.

둘째 오해는 "정책 타입이 많으면 무조건 과설계다"입니다. 정책이 많아도 "순서가 읽히고, 경계가 분명하고, 실패가 관측 가능"하면 인지 비용이 통제됩니다. 반대로 타입이 적어도 거대한 메서드/거대한 switch가 정책 순서를 숨기면 더 위험할 수 있습니다. 중요한 것은 개수가 아니라 실행 스토리의 복원 가능성입니다.

---

## 12) 팀 적용 질문(선택)

첫째, 우리가 줄이고 싶은 것은 "코드 줄 수"입니까, 아니면 "한 유스케이스를 이해하기 위해 필요한 점프(파일 이동/추론)"입니까? 후자라면 구조 축을 유스케이스로 바꾸는 편이 더 직접적입니다.

둘째, 공통화는 "어떤 의미"까지 할 것입니까? timeout/retry 자체는 골격으로 공통화하되, retry 조건/분류 같은 의미는 유스케이스에 남길지 합의해야 합니다.

셋째, 정책 적용 순서를 누가/어떻게 검증할 것입니까? 최소한 핵심 유스케이스는 "순서를 테스트로 고정"하거나, 아키텍처 규칙(정적 분석/테스트)으로 drift를 조기 탐지하는 장치가 필요합니다.

---

## 13) 참고(원전/전문가)

이 카드의 유스케이스 중심 구조(Vertical Slice) 관점은 Jimmy Bogard의 논의(요청 중심 조직)와, 경계/의존성 방향을 중시하는 Hexagonal 사고방식의 영향을 받았습니다. 또한 "규칙을 문서가 아니라 실행 가능한 검증으로 고정"한다는 생각은 evolutionary architecture에서 말하는 fitness function(아키텍처 특성을 자동 검증하는 함수) 논의와도 연결됩니다.
