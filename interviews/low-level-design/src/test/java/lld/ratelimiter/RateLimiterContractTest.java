package lld.ratelimiter;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.function.LongSupplier;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/**
 * Rate limiter의 공용 계약 테스트. 시간을 {@link FakeNanoClock}으로 주입해 결정적으로 검증한다.
 * 정답지와 내 구현이 같은 케이스를 통과해야 한다.
 */
abstract class RateLimiterContractTest {

    protected abstract RateLimiter newLimiter(long capacity, double refillPerSecond, LongSupplier nanoClock);

    @Test
    @DisplayName("시작은 가득: capacity만큼 연속 통과 후 거부")
    void starts_full_then_denies() {
        FakeNanoClock clock = new FakeNanoClock();
        RateLimiter limiter = newLimiter(5, 10, clock); // 버킷5, 초당 10개 보충
        for (int i = 0; i < 5; i++) {
            assertTrue(limiter.tryAcquire(), "초기 버스트 " + i);
        }
        assertFalse(limiter.tryAcquire(), "버킷이 비면 거부");
    }

    @Test
    @DisplayName("시간이 지나면 보충되어 다시 통과")
    void refills_over_time() {
        FakeNanoClock clock = new FakeNanoClock();
        RateLimiter limiter = newLimiter(5, 10, clock); // 초당 10개
        for (int i = 0; i < 5; i++) {
            limiter.tryAcquire();
        }
        assertFalse(limiter.tryAcquire());
        clock.advanceMillis(500); // 0.5초 -> 5개 보충
        for (int i = 0; i < 5; i++) {
            assertTrue(limiter.tryAcquire(), "보충 후 " + i);
        }
        assertFalse(limiter.tryAcquire());
    }

    @Test
    @DisplayName("보충은 버킷 용량을 넘지 않는다")
    void refill_capped_at_capacity() {
        FakeNanoClock clock = new FakeNanoClock();
        RateLimiter limiter = newLimiter(5, 10, clock);
        for (int i = 0; i < 5; i++) {
            limiter.tryAcquire();
        }
        clock.advanceSeconds(10); // 100개 보충될 시간이지만 상한은 5
        for (int i = 0; i < 5; i++) {
            assertTrue(limiter.tryAcquire());
        }
        assertFalse(limiter.tryAcquire(), "상한을 넘겨 쌓이지 않는다");
    }

    @Test
    @DisplayName("permits>1: 한 번에 여러 토큰 소비")
    void acquires_multiple_permits() {
        FakeNanoClock clock = new FakeNanoClock();
        RateLimiter limiter = newLimiter(5, 10, clock);
        assertTrue(limiter.tryAcquire(5)); // 한 번에 5개
        assertFalse(limiter.tryAcquire(1)); // 남은 0개
    }

    @Test
    @DisplayName("실패: 잘못된 인자")
    void invalid_arguments_rejected() {
        FakeNanoClock clock = new FakeNanoClock();
        assertThrows(IllegalArgumentException.class, () -> newLimiter(0, 10, clock));
        assertThrows(IllegalArgumentException.class, () -> newLimiter(5, 0, clock));
        RateLimiter limiter = newLimiter(5, 10, clock);
        assertThrows(IllegalArgumentException.class, () -> limiter.tryAcquire(0));
    }
}
