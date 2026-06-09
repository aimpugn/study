package lld.ratelimiter;

import java.util.function.LongSupplier;

/**
 * 테스트용 가짜 나노초 시계. 실제 시간 대신 손으로 전진시켜 rate limiter를 결정적으로 검증한다.
 * 운영 코드는 {@code System::nanoTime}을 주입하지만, 테스트는 이걸 주입한다.
 */
final class FakeNanoClock implements LongSupplier {

    private long nanos;

    @Override
    public long getAsLong() {
        return nanos;
    }

    void advanceMillis(long millis) {
        nanos += millis * 1_000_000L;
    }

    void advanceSeconds(double seconds) {
        nanos += (long) (seconds * 1_000_000_000.0);
    }
}
