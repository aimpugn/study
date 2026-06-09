package lld.logging;

/**
 * 로깅 라이브러리 - 면접 암기 카드. 디자인 패턴(전략) 대표 문제. 정답지는 {@link ReferenceLogger}.
 *
 * <p>직접 연습은 {@link MyLogger}. 둘 다 {@code LoggerContractTest}로 검증된다.
 *
 * <pre>
 * 한 줄    : 레벨로 거르고 -&gt; 포맷을 입혀 -&gt; 목적지에 쓴다. 핵심은 '바꿔 끼우는 축'을 인터페이스로 여는 것.
 * 축(전략) : Formatter(어떻게 보일까) + Appender(어디에 쓸까)를 분리 -&gt; 콘솔/파일/네트워크 × 평문/JSON 자유 조합.
 * 레벨 필터 : minLevel 이상만 통과. enum 순서로 비교(DEBUG &lt; INFO &lt; WARN &lt; ERROR).
 * 편의 메서드 : debug/info/warn/error는 default 메서드로 log(level, msg)에 위임 -&gt; 구현은 한 곳만.
 * 안티패턴  : 전역 Singleton 로거에 출력·포맷·레벨을 하드코딩 -&gt; 테스트 불가, 교체 불가. 의존성 주입으로 푼다.
 * 체인(확장) : Appender 여러 개(멀티 출력)나 레벨별 핸들러를 Chain of Responsibility로 연결할 수 있다.
 * 동시성    : 같은 Appender에 여러 스레드 -&gt; append 동기화 또는 비동기 큐(로깅이 핫패스를 막지 않게).
 * 실무      : SLF4J(파사드) + Logback/Log4j2. 비동기 appender, MDC, 레벨별 라우팅.
 * 꼬리질문  : 새 출력 추가는? 런타임에 포맷/레벨 변경은? 멀티 appender는? 비동기 로깅은?
 * </pre>
 */
public interface Logger {

    /** 심각도. 순서가 곧 우선순위다(아래로 갈수록 심각). */
    enum Level { DEBUG, INFO, WARN, ERROR }

    /** 메시지를 어떻게 한 줄 문자열로 만들지 (전략). */
    @FunctionalInterface
    interface Formatter {
        String format(Level level, String message);
    }

    /** 만들어진 한 줄을 어디에 쓸지 (전략). */
    @FunctionalInterface
    interface Appender {
        void append(String line);
    }

    /** 레벨이 기준 이상이면 포맷해서 목적지에 쓴다. */
    void log(Level level, String message);

    default void debug(String message) {
        log(Level.DEBUG, message);
    }

    default void info(String message) {
        log(Level.INFO, message);
    }

    default void warn(String message) {
        log(Level.WARN, message);
    }

    default void error(String message) {
        log(Level.ERROR, message);
    }
}
