package lld.ratelimiter;

import java.util.function.LongSupplier;

/** 정답지(토큰 버킷)를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceRateLimiterTest extends RateLimiterContractTest {

    @Override
    protected RateLimiter newLimiter(long capacity, double refillPerSecond, LongSupplier nanoClock) {
        return new ReferenceRateLimiter(capacity, refillPerSecond, nanoClock);
    }
}
