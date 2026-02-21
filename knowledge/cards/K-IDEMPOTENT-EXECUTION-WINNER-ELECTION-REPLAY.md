# K-IDEMPOTENT-EXECUTION-WINNER-ELECTION-REPLAY

## 1) 이 카드가 답하는 질문

외부 시스템 호출(VAN 송금, 결제, 메시지 발송 등)처럼 "실행 자체가 부작용"인 작업을 구현하다 보면 아래 질문이 반복됩니다.

"같은 요청이 재시도/중복 수신으로 두 번 들어올 수 있을 때, 외부 호출을 정확히 1번만 실행하려면 무엇이 필요합니까? 단순히 `find -> 없으면 execute` 같은 형태로는 왜 깨지고, DB 유니크 제약/트랜잭션/락을 어떻게 조합해야 합니까? 그리고 Hexagonal + DDD에서 이 '멱등 실행(guard)' 로직은 core 정책이어야 합니까, adapter 저장소 구현이어야 합니까?"

이 카드는 "멱등"을 정의로 끝내지 않고, 동시성 실패가 언제/어떤 증상으로 나타나는지와, 그 실패를 재현/검증하는 최소 실험까지 한 흐름으로 닫는 것을 목표로 합니다.

---

## 2) 배경: 왜 이 문제가 실무에서 비싸게 터지는가

"외부 호출을 두 번 실행"은 대부분의 도메인에서 치명적입니다. 금융 이체라면 중복 송금이고, 계정 생성이라면 중복 계정이며, webhook 전달이라면 중복 이벤트로 고객사 장애를 유발할 수 있습니다. 문제는 이 버그가 "항상" 재현되지 않는다는 점입니다. 동시 요청이 특정 타이밍에 겹칠 때만 발생하기 때문에, 로컬 테스트에서는 지나가고 운영에서만 간헐적으로 터지기 쉽습니다.

또한 멱등 실행은 단순히 "중복을 무시한다"가 아닙니다. 동일 키(`MessageNo`, `Idempotency-Key` 등)로 들어온 요청이 *같은 의미*인지(동일 requestHash) *다른 의미*인지(충돌)도 구분해야 합니다. 이를 구분하지 않으면, 악의적/실수로 같은 키를 재사용했을 때 조용히 잘못된 결과가 재사용될 수 있습니다.

따라서 멱등 실행은 "정확히 1번 실행"과 "충돌을 즉시 드러내기"라는 두 축을 동시에 만족해야 하고, 이 두 축을 만족시키는 최소 메커니즘(제어점)을 고정하는 것이 실무 비용을 줄입니다.

---

## 3) 핵심 제어점(왜 하필 이 축인가)

이 주제의 판단을 고정하는 제어점은 3개입니다.

첫째는 **winner 선출이 외부 호출 이전에 일어나는가**입니다. 외부 호출이 부작용이라면, DB가 막아야 하는 것은 "row 중복"이 아니라 "부작용 중복"입니다. 이를 위해서는 "조회해서 없으면 실행"이 아니라, "먼저 저장소에서 winner를 선출"하고 winner만 실행하도록 골격을 고정해야 합니다.

둘째는 **중복 키의 의미 충돌을 어떻게 탐지하는가**입니다. 같은 키가 같은 의미로 재시도되는 것은 허용되지만, 다른 의미로 재사용되는 것은 즉시 실패로 드러나야 합니다. 이를 위해 보통 `requestHash`(요청 파라미터 지문)를 저장하고 비교합니다.

셋째는 **리플레이(replay) 정책을 어떻게 닫는가**입니다. 중복 요청(loser)이 들어왔을 때, (a) 저장된 성공 응답을 그대로 반환하는지, (b) "이미 처리됨" 오류로 응답하는지, (c) 진행 중이면 어떻게 표시하는지(진행중 예외/timeout 판정)까지 정책이 필요합니다. 이 정책이 흩어지면 drift가 생기므로, 한 곳(core 정책 서비스)에 모으는 편이 안전합니다.

---

## 4) 기본 결론(기본값)과 선택 이유

기본 결론은 다음처럼 닫습니다.

외부 부작용을 "정확히 1번" 실행하려면, DB(또는 동등한 영속 저장소)를 이용해 "먼저 winner를 선출"하고 winner만 외부 호출을 실행해야 합니다. winner 선출은 보통 `(dedup_key)` 유니크 제약을 활용한 `INSERT PENDING`으로 구현합니다. loser는 저장된 row를 읽어 requestHash/상태/응답을 확인하고, replay/충돌/진행중/과거실패 중 하나로 분기합니다.

Hexagonal + DDD 관점에서는, winner 선출을 위한 SQL/락/트랜잭션은 adapter의 책임(저장소 primitive)이고, "어떤 상태에서 무엇을 반환/예외로 할지"는 core의 책임(정책)입니다. 즉 adapter는 `insertPending`, `updateSuccess`, `updateFailure`, `findByKey` 같은 원자적 연산을 제공하고, core는 이 primitive를 정해진 순서로 조합하여 불변식을 만족시키는 골격을 소유하는 것이 기본값입니다.

이 선택은 다음 비용을 줄입니다.

- 동시성 버그는 "순서" 문제로 터지는데, 순서를 core 정책 서비스로 고정하면 drift가 줄어듭니다.
- 저장소별(MySQL/PostgreSQL/SQLite)로 구현 디테일이 달라지더라도, core 정책은 그대로 유지되고 adapter만 교체됩니다.

---

## 5) 핵심 용어 정의(이 문서 맥락에서)

**멱등 실행(idempotent execution)**은 동일한 dedup 키로 같은 요청을 여러 번 수행하더라도, 외부 부작용이 최대 1번만 발생하도록 만드는 성질입니다.

**winner 선출(winner election)**은 동시에 들어온 여러 요청 중 정확히 하나만 "실행 권한"을 얻도록 결정하는 단계입니다. 이 단계가 외부 호출 이전에 닫히지 않으면, 외부 호출 중복을 막을 수 없습니다.

**requestHash**는 "같은 dedup 키"가 *같은 의미*로 재시도되는지 *다른 의미*로 재사용되는지를 판별하기 위한 요청 지문입니다.

**리플레이(replay)**는 이미 성공한 요청의 결과를 저장소에서 꺼내어, 중복 요청에게 외부 호출 없이 동일 응답을 제공하는 정책입니다.

---

## 6) 메커니즘(입력 -> 변환 규칙 -> 산출물 -> 소비자)

입력은 `(dedup_key, requestHash, action)`입니다. `action`은 외부 부작용을 포함할 수 있는 실행 함수입니다.

변환 규칙은 다음 순서로 고정합니다.

1. 저장소에 `INSERT PENDING(dedup_key, requestHash, status=PENDING)`를 시도합니다.
2. INSERT가 성공하면 이 요청이 winner입니다. winner만 `action`을 실행합니다.
3. winner는 `action` 결과를 `UPDATE SUCCESS(response_payload, status=SUCCESS...)`로 기록합니다. 예외가 나면 `UPDATE FAILURE(status=FAILED_*, resultCode...)`로 기록합니다.
4. INSERT가 유니크 충돌로 실패하면 이 요청은 loser입니다. loser는 `SELECT BY dedup_key`로 기존 row를 읽습니다.
5. loser는 (a) requestHash 불일치면 충돌 예외, (b) 성공 + 응답 존재면 replay 반환, (c) 진행 중이면 진행중 예외 또는 timeout 예외, (d) 과거 실패면 실패 예외로 종료합니다.

산출물은 `action`의 결과(또는 replay된 결과) 또는 명시적 예외입니다.

소비자는 유스케이스 서비스입니다. 소비자는 "외부 호출이 한 번만 실행되었다"는 불변식을 전제로 다음 로직(정규화, 관측, 후속 이벤트 생성)을 수행할 수 있습니다.

---

## 7) 조건 변경 시 예측(무엇이 깨지고 어떻게 관측되는가)

### 7.1 read-then-act("조회 후 실행")로 구현하면 무엇이 깨지나

"없으면 실행" 패턴은 경쟁 창(race window)을 만듭니다. 두 요청 A/B가 동시에 `find(dedup_key)=empty`를 관측한 뒤 각각 `action`을 실행하면, 외부 부작용은 2번 발생합니다. 이후 DB 유니크 제약이 insert/update에서 실패하더라도, 이미 부작용은 발생했기 때문에 늦습니다.

관측 포인트는 "중복 외부 호출"입니다. 예를 들어 외부 시스템 로그에 같은 dedup_key로 두 번 요청이 남거나, 계좌 이체가 두 번 발생하는 형태로 관측됩니다. DB에는 한 row만 남을 수 있기 때문에, DB만 보면 원인이 숨겨질 수 있다는 점이 특히 위험합니다.

### 7.2 requestHash를 저장하지 않으면 무엇이 깨지나

dedup_key만으로 멱등을 구현하면, 같은 키로 다른 요청이 들어왔을 때도 기존 결과를 재사용하거나 진행중으로 처리할 수 있습니다. 이는 "다른 의미를 같은 요청으로 오인"하는 silent corruption으로 이어집니다.

관측 포인트는 "같은 키인데 결과가 이상하다"는 형태의 운영 이슈입니다. 예외가 아니라 잘못된 성공으로 드러날 수 있어 비용이 큽니다.

### 7.3 replay를 포기하면 무엇이 깨지나

replay를 하지 않고 "이미 처리됨"으로만 응답하면, 중복 호출을 막는 데는 도움이 되지만, 소비자(클라이언트)가 결과를 다시 얻기 위해 별도 조회 API를 호출해야 할 수 있습니다. 또한 일부 도메인에서는 "같은 요청이면 같은 응답"이 계약이므로, 계약 변화가 됩니다.

관측 포인트는 클라이언트의 추가 트래픽(결과 조회) 증가 또는 409/409류 응답 증가입니다.

---

## 8) 코드 예시(나쁜 패턴 1 + 권장 패턴 1 이상)

### 8.1 나쁜 패턴: read-then-act

```java
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.Supplier;

final class BadIdempotency {
    private final ConcurrentHashMap<String, String> store = new ConcurrentHashMap<>();

    String executeOnce(String key, Supplier<String> action) {
        Optional<String> existing = Optional.ofNullable(store.get(key));
        if (existing.isPresent()) {
            return existing.get();
        }
        // 위험: 외부 부작용(action)이 이 시점에 실행됩니다.
        String result = action.get();
        store.putIfAbsent(key, result);
        return result;
    }
}
```

이 코드는 "저장"은 중복을 막아도, `action` 자체의 중복을 막지 못합니다.

### 8.2 권장 패턴: 먼저 winner 선출 후 실행

```java
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.Supplier;

final class GoodIdempotency {
    private static final String PENDING = "PENDING";

    private record Row(String requestHash, String status, String response) {}

    private final ConcurrentHashMap<String, Row> store = new ConcurrentHashMap<>();

    String executeOnce(String key, String requestHash, Supplier<String> action) {
        // 1) winner 선출: PENDING row를 먼저 만든다.
        Row pending = new Row(requestHash, PENDING, null);
        Row winner = store.putIfAbsent(key, pending);
        if (winner != null) {
            return handleDuplicate(key, requestHash, winner);
        }

        // 2) winner만 action 실행
        try {
            String result = action.get();
            store.put(key, new Row(requestHash, "SUCCESS", result));
            return result;
        } catch (RuntimeException e) {
            store.put(key, new Row(requestHash, "FAILED", null));
            throw e;
        }
    }

    private String handleDuplicate(String key, String requestHash, Row stored) {
        if (!stored.requestHash.equals(requestHash)) {
            throw new IllegalStateException("CONFLICT: key reused with different requestHash");
        }
        if (stored.response != null) {
            return stored.response; // replay
        }
        throw new IllegalStateException("IN_PROGRESS_OR_FAILED: status=" + stored.status);
    }
}
```

핵심은 "PENDING을 먼저 기록"하여 경쟁 창을 닫는 것입니다. DB에서는 `putIfAbsent`가 아니라 유니크 제약 기반 INSERT로 이 역할을 수행합니다.

---

## 9) 최소 실험(검증)

아래 실험은 "중복 외부 호출"을 실제로 재현하고, 권장 패턴이 이를 막는 것을 관측 가능한 출력으로 고정합니다.

### 9.1 실패 관측: read-then-act는 동시에 2번 action이 실행될 수 있음

```java
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class BadRaceLab {
    static final class Bad {
        private volatile String stored;
        private final CountDownLatch bothChecked;
        private final CountDownLatch proceed;

        Bad(CountDownLatch bothChecked, CountDownLatch proceed) {
            this.bothChecked = bothChecked;
            this.proceed = proceed;
        }

        String run(String key, Runnable action) {
            if (stored != null) return stored;
            // 두 스레드가 모두 stored==null을 관측한 뒤 action으로 진행하도록 강제합니다.
            // 이렇게 "경쟁 창"을 인위적으로 넓히면, read-then-act가 왜 위험한지 결정적으로 관측할 수 있습니다.
            bothChecked.countDown();
            try {
                proceed.await();
            } catch (InterruptedException ignored) {
                Thread.currentThread().interrupt();
            }
            action.run();
            stored = "OK";
            return stored;
        }
    }

    public static void main(String[] args) throws Exception {
        var bothChecked = new CountDownLatch(2);
        var proceed = new CountDownLatch(1);

        var bad = new Bad(bothChecked, proceed);
        var calls = new AtomicInteger(0);

        var pool = Executors.newFixedThreadPool(2);

        Runnable task = () -> bad.run("K", calls::incrementAndGet);

        pool.submit(task);
        pool.submit(task);

        // 두 요청이 모두 "stored==null"을 본 상태에서 동시에 action이 실행되게 만듭니다.
        bothChecked.await();
        proceed.countDown();

        pool.shutdown();
        pool.awaitTermination(1, TimeUnit.SECONDS);

        System.out.println("actionCalls=" + calls.get());
    }
}
```

기대 결과는 항상 `actionCalls=2`입니다. 즉 "조회 후 실행" 골격만으로는 외부 호출 중복을 구조적으로 막을 수 없다는 사실이, 출력으로 고정됩니다.

### 9.2 성공 관측: winner 선출(putIfAbsent)로 action이 1번만 실행됨

```java
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class WinnerElectionLab {
    static final class Good {
        private final ConcurrentHashMap<String, String> store = new ConcurrentHashMap<>();
        String run(String key, Runnable action) {
            // winner 선출(원자적)
            String prev = store.putIfAbsent(key, "PENDING");
            if (prev != null) return prev;
            action.run();
            store.put(key, "OK");
            return "OK";
        }
    }

    public static void main(String[] args) throws Exception {
        var good = new Good();
        var calls = new AtomicInteger(0);

        var start = new CountDownLatch(1);
        var pool = Executors.newFixedThreadPool(2);

        Runnable task = () -> {
            try {
                start.await();
                good.run("K", () -> {
                    calls.incrementAndGet();
                    try { Thread.sleep(50); } catch (InterruptedException ignored) {}
                });
            } catch (InterruptedException ignored) {}
        };

        pool.submit(task);
        pool.submit(task);
        start.countDown();

        pool.shutdown();
        pool.awaitTermination(1, TimeUnit.SECONDS);

        System.out.println("actionCalls=" + calls.get());
    }
}
```

기대 결과는 항상 `actionCalls=1`입니다. 즉 "winner 선출이 외부 호출 이전"에 닫히면, 외부 호출 중복을 구조적으로 막을 수 있습니다.

---

## 10) 트레이드오프/대안/Variants

첫째 대안은 "락 기반"입니다. 예를 들어 DB advisory lock 또는 별도 lock table을 사용해 key 단위로 락을 잡을 수 있습니다. 이 방식은 유니크 제약 기반과 목표는 같지만, DB마다 락 API가 달라 이식성이 떨어질 수 있고, 락 해제/timeout 정책이 별도로 필요합니다.

둘째 대안은 "외부 시스템이 멱등키를 제공"하는 경우입니다. 외부 시스템(VAN)이 idempotency key를 지원하면, 중복 부작용을 외부에서 막을 수 있습니다. 다만 이 경우에도 내부적으로는 "내가 무엇을 보냈는지"를 추적하기 위해 저장소 기록이 필요하므로, 완전한 대체는 아닙니다.

셋째 변형은 replay 정책입니다.

- replay를 유지하면 소비자는 같은 키로 같은 응답을 받으므로 사용성이 좋습니다.
- replay를 포기하면 core/adapter가 응답 직렬화/역직렬화를 덜 알아도 되지만, 계약 변화와 후속 조회 트래픽 증가가 발생할 수 있습니다.

---

## 11) 흔한 오해

첫째 오해는 "DB 유니크 제약이 있으니 중복 실행도 자동으로 막힌다"입니다. 유니크 제약은 DB row 중복은 막아도, 외부 호출이 먼저 실행되면 부작용 중복은 막지 못합니다. 따라서 winner 선출은 외부 호출 이전에 닫혀야 합니다.

둘째 오해는 "select for update로 없는 row도 잠글 수 있다"입니다. `SELECT ... FOR UPDATE`는 보통 "이미 존재하는 row"를 잠급니다. 없는 row에 대해 동일 효과를 얻으려면 별도 락 엔트리 또는 격리수준/갭락 같은 DB 종속 동작을 기대해야 하며, 이는 재현성과 이식성 비용이 큽니다.

---

## 12) 팀 적용 질문(선택)

1. dedup 키는 무엇으로 고정합니까? (`MessageNo`, `Idempotency-Key`, `eventId` 등)
2. winner 선출은 외부 호출 이전에 닫혀 있습니까? (INSERT PENDING / advisory lock 등)
3. requestHash로 충돌을 탐지합니까? 충돌은 예외로 즉시 드러나고 있습니까?
4. replay가 필요한 도메인입니까? 필요하다면 저장 표현과 타입 결정 규칙은 어디에 두고 검증(테스트)할 것입니까?
