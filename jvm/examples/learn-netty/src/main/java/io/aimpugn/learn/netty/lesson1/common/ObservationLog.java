package io.aimpugn.learn.netty.lesson1.common;

import java.time.LocalTime;
import java.time.format.DateTimeFormatter;
import java.util.Objects;
import java.util.concurrent.atomic.AtomicLong;

/**
 * 첫 강의의 핵심은 "요청이 어디서 어떻게 흘렀는지 눈으로 보이게 만드는 것"이므로
 * 각 서버가 같은 형태로 로그를 남기도록 맞춰 둡니다.
 */
public final class ObservationLog {

    private static final DateTimeFormatter TIME_FORMATTER = DateTimeFormatter.ofPattern("HH:mm:ss.SSS");
    private final AtomicLong eventSequence = new AtomicLong();
    private final String serverName;
    private final long startedAtNanos = System.nanoTime();

    public ObservationLog(String serverName) {
        this.serverName = serverName;
    }

    public void event(String phase, String channel, String detail) {
        long elapsedMillis = (System.nanoTime() - startedAtNanos) / 1_000_000;
        long sequence = eventSequence.incrementAndGet();

        synchronized (System.out) {
            System.out.printf(
                    "#%04d %s | +%5dms | %-8s | %-12s | %-10s | thread=%-18s | %s%n",
                    sequence,
                    LocalTime.now().format(TIME_FORMATTER),
                    elapsedMillis,
                    serverName,
                    phase,
                    Objects.requireNonNullElse(channel, "-"),
                    Thread.currentThread().getName(),
                    detail
            );
        }
    }

    public static String previewText(String text) {
        String escaped = text
                .replace("\r", "\\r")
                .replace("\n", "\\n");
        if (escaped.length() <= 80) {
            return escaped;
        }
        return escaped.substring(0, 77) + "...";
    }
}
