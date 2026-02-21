# K-ASYNC-EVENTS-OUTBOX-INBOX-CLAIM-LEASE

## 1) 이 카드가 답하는 질문

비동기 콜백(webhook) 수신이나 외부 시스템으로의 비동기 전달(재시도 포함)을 설계/구현하다 보면 아래 질문이 반복됩니다.

"왜 inbox/outbox 같은 테이블(또는 저장소)이 필요합니까? 단순히 요청/응답을 `bridge_transactions` 같은 트랜잭션 테이블에 같이 저장하면 안 됩니까? 여러 워커가 동시에 돌 때 같은 이벤트를 두 번 처리하지 않으려면(중복 전달/중복 처리) 저장소는 어떤 원자적 primitive를 제공해야 합니까? 그리고 Hexagonal + DDD에서 '상태기계/재시도 정책'은 core가 가져가고, adapter(JDBC)는 무엇만 제공해야 경계가 깨지지 않습니까?"

이 카드는 "outbox가 좋다" 수준이 아니라, 왜 이 문제가 운영에서 비싸게 터지는지(실패 비용), 어디를 제어점으로 고정해야 하는지(최소 축), 실제로 어떤 메커니즘으로 동작하는지(입력->변환->산출물->소비자), 그리고 그 결론을 작은 실험으로 어떻게 검증하는지까지 닫는 것을 목표로 합니다.

---

## 2) 배경: 왜 이 문제가 실무에서 비싸게 터지는가

비동기 전달은 동기 요청/응답과 달리 "실패를 정상으로 취급"해야 합니다. 네트워크 타임아웃, 상대 시스템 장애, 일시적 5xx, 재시도 폭주 같은 일이 규칙적으로 발생하기 때문에, 구현이 조금만 느슨해도 아래 두 가지 비용이 급격히 증가합니다.

첫째, 전달이 누락되거나(유실) 반대로 중복 전달됩니다. 유실은 고객 입장에서 "안 왔다"로 관측되지만, 원인은 내부 스레드/프로세스/배포/재시작 타이밍과 얽혀 재현이 어렵습니다. 중복은 더 위험합니다. 고객사에서 webhook을 2번 처리해 데이터가 두 번 반영되거나, 동일 메시지가 2번 발송되는 식으로 "외부 피해"로 번지기 때문입니다.

둘째, 동시성(여러 워커)에서 문제가 간헐적으로만 터집니다. 한 워커만 돌 때는 문제가 없는데, 스케일아웃(워커 2개, 3개) 순간부터 중복 처리나 lease 만료 경쟁이 나타납니다. 이런 버그는 로컬 테스트에서 잘 안 보이고, 운영에서 특정 타이밍에만 관측되기 때문에 RCA 비용이 큽니다.

따라서 비동기 전달은 "기능이 된다"가 아니라, "재시작/장애/동시성"까지 포함한 실패 모델을 설계에 포함해야 실무 비용이 내려갑니다.

---

## 3) 핵심 제어점(왜 하필 이 축인가)

이 주제는 넓기 때문에, 판단을 고정하는 제어점 4개로 축을 좁힙니다.

첫째는 **SSOT 분리**입니다. 동기 업무 트랜잭션(요청/응답) 원장과 비동기 이벤트(수신/전달) 원장은 조회 패턴, 인덱스, 업데이트 경합 패턴이 다릅니다. 비동기를 동기 트랜잭션 테이블에 합치면 워커가 `status/lease`를 계속 업데이트하면서 핫스팟이 생기고, 동기 조회까지 영향을 받습니다.

둘째는 **이벤트 Envelope(공통 컬럼) 고정**입니다. "범용 JSON 덩어리"로만 저장하면, 상태 전이/중복 제거/관측을 SQL로 강제하기 어렵습니다. 최소한 `event_id`, `direction`, `event_type`, `dedup_key`, `status`, `attempt_count`, `next_attempt_at`, `lease_until`, `last_error`, `created_at` 같은 Envelope는 고정돼야, 재시도/claim/중복 제거가 데이터 레벨에서 닫힙니다.

셋째는 **claim lease(원자적 선출)** 입니다. 여러 워커가 동일 이벤트를 동시에 처리하지 않으려면, "조회 후 처리"가 아니라 "저장소에서 먼저 처리권을 선출"해야 합니다. 즉 `UPDATE ... WHERE status IN (...) AND lease_until < now` 같은 조건부 업데이트(또는 동등한 원자 연산)가 primitive로 필요합니다.

넷째는 **정책과 primitive 분리**입니다. "어떤 status에서 언제 retry로 바꾸고 backoff를 어떻게 계산하는지"는 정책이므로 core가 소유해야 drift가 줄어듭니다. 반대로 "조건부 claim/update/insert"는 DB 종속 구현이므로 adapter가 제공해야 합니다.

---

## 4) 기본 결론(기본값)과 선택 이유

기본 결론은 다음처럼 닫습니다.

비동기 전달 내구성을 얻으면서 테이블 폭증을 막고 싶다면, 저장소(테이블) 구조는 보통 2개 SSOT로 수렴합니다.

1) 동기 업무 트랜잭션 SSOT: `bridge_transactions`  
2) 비동기 이벤트 SSOT: `bridge_async_events` (inbox/outbox 통합)

핵심은 "통합"을 하되, 무한 범용 JSON 단일 컬럼으로 뭉개는 것이 아니라, 공통 Envelope 컬럼을 고정해 상태기계를 데이터로 드러내는 것입니다. 이렇게 하면 inbox/outbox를 테이블로는 분리하지 않으면서도, `direction`(INBOUND/OUTBOUND) + `status` + `lease_until`을 통해 수신/전달/재시도/중복 제거를 일관된 규칙으로 구현할 수 있습니다.

Hexagonal + DDD 관점에서 이 기본값이 유리한 이유는, core는 "이벤트 전달 정책"을 테스트 가능한 순수 로직으로 고정할 수 있고, adapter는 그 정책이 필요로 하는 원자적 primitive를 제공하는 얇은 구현으로 남을 수 있기 때문입니다. 즉 정책은 core에, 저장소 primitive는 adapter에 둬서 변화 주기를 분리합니다.

---

## 5) 핵심 용어 정의(이 문서 맥락에서)

**비동기 이벤트(Async Event)**는 외부 시스템과의 통신 결과를 "즉시 완료"시키지 않고, 저장소에 기록한 뒤 워커가 나중에 전달/처리/재시도를 수행하는 단위입니다. 이 카드에서는 webhook 수신(inbox)과 외부 전달(outbox)을 모두 포함합니다.

**Outbox 패턴**은 "외부로 보내야 할 것"을 먼저 로컬 저장소에 기록하고, 별도 워커가 이를 읽어 전달하며 성공/실패 상태를 저장하는 패턴입니다. 핵심은 프로세스 재시작이 발생해도 전달 의도가 저장소에 남아 복구 가능하다는 점입니다.

**Inbox 패턴**은 외부에서 들어온 이벤트를 저장소에 기록하고, dedup_key로 중복 처리를 방지하는 패턴입니다. 핵심은 "중복 수신"이 정상인 환경(재시도/네트워크)에서도 내부 처리가 중복되지 않게 하는 것입니다.

**claim lease**는 워커가 특정 이벤트를 일정 시간 동안 "내가 처리 중"으로 선출(claim)하고, 그 시간(lease) 동안 다른 워커가 같은 이벤트를 처리하지 못하게 하는 규칙입니다. 이 규칙은 원자적 업데이트가 없으면 깨집니다.

---

## 6) 메커니즘(입력 -> 변환 규칙 -> 산출물 -> 소비자)

이벤트 SSOT를 기준으로 한 비동기 전달 흐름을 입력/변환/산출물/소비자로 닫으면 다음과 같습니다.

입력은 두 종류입니다. (1) 외부에서 들어온 콜백 요청(HTTP body 등)과 (2) 내부 유스케이스가 "외부로 보내야 한다"고 결정한 전달 요청(예: webhook notify payload)입니다.

변환 규칙은 core 정책 서비스가 소유합니다. 수신(inbound)이면 `dedup_key`를 계산해 "이미 처리했는지"를 결정하고, 신규면 `RECEIVED/PENDING` 같은 상태로 이벤트를 기록합니다. 전달(outbound)이면 이벤트를 `PENDING`으로 기록하고, 워커가 `PENDING/RETRYING` 이벤트 중 `next_attempt_at <= now`인 것만 claim하도록 합니다. claim은 저장소 primitive(조건부 UPDATE)로 원자적으로 수행되어야 하며, 성공하면 상태가 `DELIVERING`으로 바뀌고 `lease_until`이 설정됩니다.

산출물은 이벤트 row의 상태 변화입니다. 성공하면 `SENT`로, 실패하면 backoff 정책에 따라 `RETRYING(next_attempt_at=...)` 또는 `FAILED`로 전이됩니다. 이 상태 전이는 워커 재시작/크래시가 발생해도 저장소에서 복원 가능합니다.

소비자는 워커(배치/스케줄러)와 운영자입니다. 워커는 상태에 따라 재처리를 수행하고, 운영자는 `FAILED` 이벤트와 `attempt_count`를 관측해 장애를 복원합니다.

---

## 7) 조건 변경 시 예측(무엇이 깨지고 어떻게 관측되는가)

### 7.1 claim lease 없이 "조회 후 처리"로 구현하면 무엇이 깨지나

두 워커가 동시에 같은 pending 이벤트를 읽고 둘 다 전달을 실행할 수 있습니다. 이 실패는 DB에 "최종 상태"가 하나만 남을 수 있어(마지막 업데이트가 이김) DB만 보면 원인이 숨겨질 수 있습니다.

관측 포인트는 외부 시스템(고객사)에서의 중복 수신, 또는 내부 전송 로그에서 동일 `event_id/dedup_key`로 2번 전송이 남는 형태입니다.

### 7.2 비동기 이벤트를 동기 트랜잭션 테이블에 합치면 무엇이 깨지나

워커가 `status/attempt/lease`를 자주 업데이트하면서 트랜잭션 테이블이 hot row/인덱스 경합 지점이 됩니다. 이 경합은 동기 API 조회 지연으로 전이될 수 있습니다.

관측 포인트는 DB에서 특정 인덱스/row에 대한 lock wait 증가, 그리고 동기 조회 쿼리 latency 증가입니다(예: p95가 상승).

### 7.3 Envelope 없이 payload_json만 저장하면 무엇이 깨지나

재시도/중복 제거/관측이 "애플리케이션 로직"에만 존재하게 되고, 운영 중 상태를 SQL로 빠르게 확인하기 어렵습니다. 또한 잘못된 상태 전이가 발생했을 때 복구 도구(쿼리/대시보드) 작성이 어려워집니다.

관측 포인트는 장애 시 "어느 상태에서 멈췄는지"를 확인하기 위해 코드/로그를 따라가야 하는 상황입니다.

---

## 8) 코드 예시(나쁜 패턴 1 + 권장 패턴 1 이상)

### 8.1 나쁜 패턴: loadPending() 후 처리(중복 가능)

```java
import java.util.List;

final class BadOutboxWorker {
    private final OutboxStore store;
    private final Sender sender;

    BadOutboxWorker(OutboxStore store, Sender sender) {
        this.store = store;
        this.sender = sender;
    }

    void tick() {
        // 문제: 여러 워커가 동시에 같은 목록을 읽을 수 있습니다.
        List<OutboxEvent> pending = store.loadPending();
        for (OutboxEvent e : pending) {
            sender.send(e.payload());
            store.markSent(e.eventId());
        }
    }

    interface OutboxStore {
        List<OutboxEvent> loadPending();
        void markSent(String eventId);
    }

    interface Sender {
        void send(String payload);
    }

    record OutboxEvent(String eventId, String payload) {}
}
```

이 코드는 "DB에 상태가 남는다"는 이유로 안전해 보이지만, 중복 전달을 막는 원자적 선출이 없어서 동시에 2번 실행될 수 있습니다.

### 8.2 권장 패턴: claim(원자) -> send -> markSent/markRetry

```java
import java.time.Instant;
import java.util.Optional;

final class GoodOutboxWorker {
    private final AsyncEventRepository repo;
    private final Sender sender;
    private final InstantSource clock;

    GoodOutboxWorker(AsyncEventRepository repo, Sender sender, InstantSource clock) {
        this.repo = repo;
        this.sender = sender;
        this.clock = clock;
    }

    void tick(String workerId) {
        Optional<AsyncEvent> claimed = repo.claimNextDeliverable(workerId, clock.now(), /*leaseSeconds=*/30);
        if (claimed.isEmpty()) return;

        AsyncEvent e = claimed.get();
        try {
            sender.send(e.payloadJson());
            repo.markSent(e.eventId(), clock.now());
        } catch (RuntimeException ex) {
            repo.markRetry(e.eventId(), clock.now(), ex.getMessage());
        }
    }

    interface AsyncEventRepository {
        // adapter는 이 메서드를 "조건부 UPDATE + 반환" 같은 원자 primitive로 구현해야 합니다.
        Optional<AsyncEvent> claimNextDeliverable(String workerId, Instant now, long leaseSeconds);
        void markSent(String eventId, Instant now);
        void markRetry(String eventId, Instant now, String lastError);
    }

    interface Sender {
        void send(String payloadJson);
    }

    interface InstantSource {
        Instant now();
    }

    record AsyncEvent(String eventId, String payloadJson) {}
}
```

핵심은 "조회(load)"가 아니라 "선출(claim)"을 저장소 원자 연산으로 고정하는 것입니다.

---

## 9) 최소 실험(검증)

아래 실험은 프로젝트 명령/DB에 의존하지 않고, 작은 Java 코드로 "중복 전달"과 "claim로 중복이 사라짐"을 출력으로 고정합니다. 하나는 실패를 관측하도록 설계합니다.

### 9.1 실패 관측: loadPending()은 두 워커가 같은 이벤트를 2번 처리할 수 있음

```java
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class BadOutboxLab {
    static final class Store {
        private final List<String> pending = new CopyOnWriteArrayList<>(List.of("E1"));

        List<String> loadPending() {
            return List.copyOf(pending);
        }

        void markSent(String eventId) {
            pending.remove(eventId);
        }
    }

    public static void main(String[] args) throws Exception {
        var store = new Store();
        var deliveries = new AtomicInteger(0);

        // 두 워커가 "같은 pending 목록"을 확보한 뒤 동시에 send하도록 강제합니다.
        var bothLoaded = new CountDownLatch(2);
        var proceed = new CountDownLatch(1);

        Runnable worker = () -> {
            List<String> list = store.loadPending();
            bothLoaded.countDown();
            try { proceed.await(); } catch (InterruptedException ignored) { Thread.currentThread().interrupt(); }

            for (String e : list) {
                deliveries.incrementAndGet();
                store.markSent(e);
            }
        };

        var pool = Executors.newFixedThreadPool(2);
        pool.submit(worker);
        pool.submit(worker);

        bothLoaded.await();
        proceed.countDown();

        pool.shutdown();
        pool.awaitTermination(1, TimeUnit.SECONDS);

        System.out.println("deliveries=" + deliveries.get());
    }
}
```

기대 결과는 항상 `deliveries=2`입니다. 즉 "조회 후 처리" 골격은 워커 수가 2가 되는 순간 중복 전달을 구조적으로 허용한다는 사실이 출력으로 고정됩니다.

### 9.2 성공 관측: claim(원자 선출)을 제공하면 deliveries가 1로 고정됨

```java
import java.time.Instant;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class GoodOutboxLab {
    static final class Store {
        private final ConcurrentHashMap<String, String> status = new ConcurrentHashMap<>();

        Store() {
            status.put("E1", "PENDING");
        }

        Optional<String> claim(String eventId) {
            // winner 선출(원자적): PENDING -> DELIVERING 전이
            boolean claimed = status.replace(eventId, "PENDING", "DELIVERING");
            return claimed ? Optional.of(eventId) : Optional.empty();
        }

        void markSent(String eventId) {
            status.put(eventId, "SENT");
        }
    }

    public static void main(String[] args) throws Exception {
        var store = new Store();
        var deliveries = new AtomicInteger(0);

        var start = new CountDownLatch(1);

        Runnable worker = () -> {
            try { start.await(); } catch (InterruptedException ignored) { Thread.currentThread().interrupt(); }
            store.claim("E1").ifPresent(e -> {
                deliveries.incrementAndGet();
                store.markSent(e);
            });
        };

        var pool = Executors.newFixedThreadPool(2);
        pool.submit(worker);
        pool.submit(worker);
        start.countDown();

        pool.shutdown();
        pool.awaitTermination(1, TimeUnit.SECONDS);

        System.out.println("deliveries=" + deliveries.get());
        System.out.println("status=" + store.status.get("E1") + " at " + Instant.now());
    }
}
```

기대 결과는 항상 `deliveries=1`이고 `status=SENT`입니다. 즉 "저장소가 claim primitive를 제공하고, 워커가 그 primitive를 먼저 호출"하는 골격으로 중복 전달을 구조적으로 막을 수 있습니다.

---

## 10) 트레이드오프/대안/Variants

첫째 대안은 inbox/outbox를 테이블로 분리하는 것입니다. 분리하면 각 테이블의 인덱스/제약/상태기계가 더 단순해질 수 있지만, 테이블/코드가 늘고 "공통 처리(lease/retry/관측)"가 중복되기 쉽습니다. 운영 팀이 테이블 폭증을 문제로 보고 있다면, "통합 테이블 + 고정 Envelope"가 더 안정적인 기본값이 됩니다.

둘째 대안은 비동기 이벤트를 메시지 큐(Kafka 등)로만 처리하는 것입니다. 큐가 내구성과 분산 소비를 제공하지만, 큐의 재처리/오프셋/중복 의미를 도메인과 일치시키려면 별도 원장(최소한의 DB 기록)이 결국 필요해지는 경우가 많습니다. 또한 "내가 무엇을 언제 보냈는지"를 운영에서 재구성하려면 DB 관측 지점이 있는 편이 복원력이 높습니다.

셋째 변형은 claim 구현 방식입니다. 유니크 제약 기반(INSERT/UPDATE)과 조건부 업데이트 기반(WHERE status=...)은 둘 다 가능하지만, 핵심은 "외부 호출 이전에 처리권이 선출되어야 한다"는 불변식입니다.

---

## 11) 흔한 오해

첫째 오해는 "outbox는 로깅이고, 운영 최적화일 뿐이다"입니다. outbox의 본질은 로깅이 아니라, **재시작 안전성과 재시도 가능성**입니다. outbox가 없으면 프로세스 크래시/배포 중간에 "보내야 했던 것"이 메모리에서 사라져 유실됩니다.

둘째 오해는 "동기 트랜잭션 테이블 하나로 다 합치면 SSOT니까 더 안전하다"입니다. SSOT는 안전성을 높일 수 있지만, 서로 다른 경합/업데이트 패턴을 한 테이블로 합치면 동기 경로를 느리게 하거나 락 경합을 키우는 형태로 실패할 수 있습니다. 그래서 "SSOT는 줄이되 의미/패턴은 분리"가 운영 안전성에 더 직접적입니다.

---

## 12) 팀 적용 질문(선택)

1. 비동기 이벤트의 dedup_key는 무엇으로 고정합니까? (콜백이면 외부가 제공하는 event id, 없으면 payload 지문)
2. 워커 스케일아웃(2개 이상) 시에도 중복 처리가 발생하지 않도록 claim lease primitive가 존재합니까?
3. Envelope 컬럼(상태/시도/리스/오류)을 고정했습니까? 운영에서 SQL로 "지금 어디서 막혔는지"를 확인할 수 있습니까?
4. 정책(상태기계/백오프/최대 시도)은 core에 있고, 원자 primitive(조건부 update/insert)는 adapter에 있습니까?
