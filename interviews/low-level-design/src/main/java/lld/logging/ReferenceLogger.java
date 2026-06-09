package lld.logging;

/**
 * 로깅 라이브러리 정답지.
 *
 * <p>로거 자신은 '레벨 필터'라는 정책 하나만 갖고, '어떻게 보일지(Formatter)'와 '어디에 쓸지(Appender)'는
 * 주입받은 전략에 위임한다. 그래서 콘솔/파일/JSON을 조합으로 바꿔 끼울 수 있고 테스트에서 출력을 가로챌 수 있다.
 */
public final class ReferenceLogger implements Logger {

    private final Level minLevel;
    private final Formatter formatter;
    private final Appender appender;

    public ReferenceLogger(Level minLevel, Formatter formatter, Appender appender) {
        if (minLevel == null || formatter == null || appender == null) {
            throw new NullPointerException("minLevel, formatter, appender must not be null");
        }
        this.minLevel = minLevel;
        this.formatter = formatter;
        this.appender = appender;
    }

    @Override
    public void log(Level level, String message) {
        if (level.ordinal() < minLevel.ordinal()) {
            return; // 기준 미달 레벨은 버린다.
        }
        appender.append(formatter.format(level, message));
    }
}
