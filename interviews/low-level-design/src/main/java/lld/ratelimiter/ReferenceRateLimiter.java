package lld.ratelimiter;

import java.util.function.LongSupplier;

/**
 * Rate limiter 정답지 - 토큰 버킷(token bucket).
 *
 * <p>핵심은 'lazy refill': 별도 타이머 없이 요청이 올 때마다 마지막 보충 이후 경과 시간만큼 토큰을 채운다.
 * 시간 소스({@link LongSupplier})를 주입받아 테스트에서 시간을 손으로 전진시킬 수 있게 한다.
 */
public final class ReferenceRateLimiter implements RateLimiter {

    private final double capacity;
    private final double refillPerNano;
    private final LongSupplier nanoClock;

    private double tokens;
    private long lastRefillNanos;

    public ReferenceRateLimiter(long capacity, double refillPerSecond, LongSupplier nanoClock) {
        if (capacity <= 0) {
            throw new IllegalArgumentException("capacity must be > 0 but was " + capacity);
        }
        if (refillPerSecond <= 0) {
            throw new IllegalArgumentException("refillPerSecond must be > 0 but was " + refillPerSecond);
        }
        this.capacity = capacity;
        this.refillPerNano = refillPerSecond / 1_000_000_000.0;
        this.nanoClock = nanoClock;
        this.tokens = capacity; // 시작은 가득 채워 버스트를 허용한다.
        this.lastRefillNanos = nanoClock.getAsLong();
    }

    /** 운영용: 실제 시계를 쓴다. */
    public ReferenceRateLimiter(long capacity, double refillPerSecond) {
        this(capacity, refillPerSecond, System::nanoTime);
    }

    @Override
    public synchronized boolean tryAcquire(int permits) {
        if (permits <= 0) {
            throw new IllegalArgumentException("permits must be > 0 but was " + permits);
        }
        refill();
        if (tokens >= permits) {
            tokens -= permits;
            return true;
        }
        return false;
    }

    private void refill() {
        long now = nanoClock.getAsLong();
        long elapsed = now - lastRefillNanos;
        if (elapsed <= 0) {
            return; // 시계 역행이나 동일 시각이면 보충하지 않는다.
        }
        tokens = Math.min(capacity, tokens + elapsed * refillPerNano);
        lastRefillNanos = now;
    }
}
