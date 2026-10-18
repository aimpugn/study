package lab;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.UncheckedIOException;
import java.nio.charset.StandardCharsets;
import java.time.Instant;

/**
 * 맥의 지휘(txlab의 spring.go)와 표준 입출력으로 주고받는 멈춤 자리입니다.
 *
 * <p>앱은 멈춤 자리에 오면 {@code @@TXLAB AT <이름> <시각>} 한 줄을 쓰고, 지휘가 컨테이너 안에서 서버 상태를 찍은 뒤 보내는
 * {@code go} 한 줄을 받을 때까지 기다립니다. 그래서 각 관찰은 앱의 코드가 정확히 그 줄에 있는 순간의 서버 상태입니다.
 * 표준 출력은 이 약속에만 쓰고, 로그는 표준 오류로 보냅니다(logback.xml).
 *
 * <p>{@code -Dlab.free=true}이면 기다리지 않고 줄만 씁니다. 지휘 없이 앱만 돌려 보는 시험용입니다.
 */
final class Lab {

    private static final BufferedReader IN =
            new BufferedReader(new InputStreamReader(System.in, StandardCharsets.UTF_8));

    private static final boolean FREE = Boolean.getBoolean("lab.free");

    private Lab() {
    }

    /**
     * 멈춤 자리를 알리고 지휘의 {@code go}를 기다립니다.
     *
     * @param label 멈춤 자리 이름. 공백 없이 씁니다(지휘가 공백으로 칸을 나눔).
     * @throws IllegalStateException 지휘가 {@code go}가 아닌 줄을 보내거나 입력이 끝났을 때
     * @throws UncheckedIOException 표준 입력을 읽지 못했을 때
     */
    static synchronized void at(String label) {
        System.out.println("@@TXLAB AT " + label + " " + epochNanos());
        System.out.flush();
        if (FREE) {
            return;
        }
        try {
            String line = IN.readLine();
            if (!"go".equals(line)) {
                throw new IllegalStateException("멈춤 자리 " + label + "에서 go 대신 " + line + "을 받았습니다");
            }
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }
    }

    /**
     * 앱이 받은 예외 하나를 알립니다. 지휘는 기다리지 않고 기록만 합니다.
     *
     * @param where 예외를 받은 자리(공백 없이)
     * @param t 받은 예외. 원인 사슬의 클래스 이름을 모두 적어, 어느 층의 예외가 어느 층의 예외로 바뀌었는지 남깁니다.
     */
    static synchronized void note(String where, Throwable t) {
        StringBuilder chain = new StringBuilder();
        for (Throwable c = t; c != null; c = c.getCause()) {
            if (chain.length() > 0) {
                chain.append(" <- ");
            }
            chain.append(c.getClass().getName());
        }
        String message = String.valueOf(t.getMessage()).replace('\n', ' ');
        System.out.println("@@TXLAB NOTE " + where + " " + epochNanos() + " " + chain + " | " + message);
        System.out.flush();
    }

    /** 지금 시각을 1970년부터의 나노초로 돌려줍니다. JFR과 지휘의 맥 쪽 기록이 같은 벽시계를 씁니다. */
    private static long epochNanos() {
        Instant now = Instant.now();
        return now.getEpochSecond() * 1_000_000_000L + now.getNano();
    }
}
