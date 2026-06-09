package lld.logging;

/** 정답지를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceLoggerTest extends LoggerContractTest {

    @Override
    protected Logger newLogger(Logger.Level minLevel, Logger.Formatter formatter, Logger.Appender appender) {
        return new ReferenceLogger(minLevel, formatter, appender);
    }
}
