package lld.logging;

/**
 * 직접 구현하는 공간. {@link Logger} 헤더 카드와 테스트만 보고 채운다.
 * 막히면 그때만 {@link ReferenceLogger}를 연다.
 *
 * <p>핵심 질문: 로거가 직접 갖는 정책은 무엇이고(레벨 필터), 무엇을 전략으로 위임하나(포맷·출력)?
 * 시작하려면 TODO를 채우고 {@code MyLoggerTest}의 {@code @Disabled}를 지운다.
 */
public final class MyLogger implements Logger {

    public MyLogger(Level minLevel, Formatter formatter, Appender appender) {
        // TODO: null 검증 + 필드 보관
    }

    @Override
    public void log(Level level, String message) {
        // TODO: 레벨이 minLevel 이상이면 formatter로 만든 줄을 appender에 넘긴다
        throw new UnsupportedOperationException("아직 구현 전: MyLogger.log");
    }
}
