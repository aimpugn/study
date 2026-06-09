package lld.ratelimiter;

/**
 * Rate Limiter(요청 속도 제한) - 면접 암기 카드. 정답지는 토큰 버킷({@link ReferenceRateLimiter}).
 *
 * <p>직접 연습은 {@link MyRateLimiter}. 둘 다 {@code RateLimiterContractTest}의 같은 테스트로 검증된다.
 * 시간이 얽히는 문제라 '시계를 주입'해 결정적으로 테스트한다.
 *
 * <pre>
 * 한 줄    : 평균 속도는 제한하되 버스트는 버킷 용량까지 허용. 요청마다 토큰 1개 소비, 토큰은 일정 속도로 채워짐.
 * 동작     : 요청 시 (경과시간 × 보충속도)만큼 토큰을 더하고(상한=capacity), 1개 이상이면 소비 후 통과, 아니면 거부.
 * 왜 lazy refill : 타이머 스레드 없이 '요청 시점'에 계산 -> O(1), 스레드 0개. 마지막 보충 시각만 저장.
 * 대안     : Leaky bucket(출력 평탄화, 버스트 흡수 X) / Fixed window(경계에서 최대 2배 버스트 결함) /
 *           Sliding window log(정확하나 메모리 O(요청수)) / Sliding window counter(절충안).
 * 동시성   : 토큰·시각 갱신이 원자적이어야 한다 -> synchronized 또는 CAS. 안 하면 한도 초과 통과.
 * 분산     : 노드마다 버킷이면 전체 한도를 넘는다 -> Redis(Lua 토큰버킷, INCR+TTL)로 중앙화.
 * 테스트 팁: System.nanoTime 대신 시계를 주입해 시간을 손으로 전진시킨다.
 * 꼬리질문 : 버스트 허용량은? 분산에서 정확성은? 시계 역행은? permits&gt;1 요청은?
 * </pre>
 */
public interface RateLimiter {

    /** 토큰 {@code permits}개를 즉시 확보 시도. 성공하면 소비 후 {@code true}, 부족하면 {@code false}. */
    boolean tryAcquire(int permits);

    /** 토큰 1개 확보 시도. */
    default boolean tryAcquire() {
        return tryAcquire(1);
    }
}
