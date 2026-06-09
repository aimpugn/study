package lld.ratelimiter;

import java.util.function.LongSupplier;

/**
 * 직접 구현하는 공간. {@link RateLimiter} 헤더 카드와 테스트만 보고 토큰 버킷을 채운다.
 * 막히면 그때만 {@link ReferenceRateLimiter}를 연다.
 *
 * <p>생성자 시그니처는 정답지와 같다(테스트가 시계를 주입하기 때문). TODO를 채우고
 * {@code MyRateLimiterTest}의 {@code @Disabled}를 지운 뒤 통과시킨다.
 */
public final class MyRateLimiter implements RateLimiter {

    public MyRateLimiter(long capacity, double refillPerSecond, LongSupplier nanoClock) {
        // TODO: capacity·refill 검증 + 토큰/마지막보충시각 초기화 (tokens는 capacity로 시작)
    }

    @Override
    public boolean tryAcquire(int permits) {
        // TODO: refill(경과시간×보충속도, 상한 capacity) 후 tokens>=permits면 소비하고 true
        throw new UnsupportedOperationException("아직 구현 전: MyRateLimiter.tryAcquire");
    }
}
