package lld.logging;

import static lld.logging.Logger.Level.DEBUG;
import static lld.logging.Logger.Level.INFO;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.ArrayList;
import java.util.List;
import lld.logging.Logger.Appender;
import lld.logging.Logger.Formatter;
import lld.logging.Logger.Level;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/**
 * 로거의 공용 계약 테스트. 출력 전략(Appender)으로 메모리 리스트를 주입해 무엇이 찍혔는지 가로챈다.
 * Formatter·Appender가 함수형 인터페이스라 람다로 바로 만든다.
 */
abstract class LoggerContractTest {

    protected abstract Logger newLogger(Level minLevel, Formatter formatter, Appender appender);

    private final List<String> captured = new ArrayList<>();
    private final Formatter simpleFormat = (level, message) -> "[" + level + "] " + message;

    private Appender capture() {
        return captured::add;
    }

    @Test
    @DisplayName("기준 미달 레벨은 걸러진다")
    void below_min_level_is_filtered() {
        Logger logger = newLogger(INFO, simpleFormat, capture());
        logger.debug("hidden");
        assertTrue(captured.isEmpty());
    }

    @Test
    @DisplayName("기준 이상 레벨은 통과한다")
    void at_or_above_min_level_passes() {
        Logger logger = newLogger(INFO, simpleFormat, capture());
        logger.info("a");
        logger.warn("b");
        logger.error("c");
        assertEquals(List.of("[INFO] a", "[WARN] b", "[ERROR] c"), captured);
    }

    @Test
    @DisplayName("Formatter 전략이 적용된다")
    void formatter_is_applied() {
        Logger logger = newLogger(DEBUG, simpleFormat, capture());
        logger.debug("hi");
        assertEquals(List.of("[DEBUG] hi"), captured);
    }

    @Test
    @DisplayName("출력 순서가 보존된다")
    void order_is_preserved() {
        Logger logger = newLogger(DEBUG, (level, message) -> message, capture());
        logger.info("1");
        logger.info("2");
        logger.info("3");
        assertEquals(List.of("1", "2", "3"), captured);
    }

    @Test
    @DisplayName("편의 메서드가 올바른 레벨로 매핑된다")
    void convenience_methods_map_to_levels() {
        Logger logger = newLogger(DEBUG, simpleFormat, capture());
        logger.warn("w");
        assertEquals(List.of("[WARN] w"), captured);
    }

    @Test
    @DisplayName("실패: 의존성이 null이면 거부")
    void null_dependencies_rejected() {
        assertThrows(NullPointerException.class, () -> newLogger(null, simpleFormat, capture()));
        assertThrows(NullPointerException.class, () -> newLogger(INFO, null, capture()));
        assertThrows(NullPointerException.class, () -> newLogger(INFO, simpleFormat, null));
    }
}
