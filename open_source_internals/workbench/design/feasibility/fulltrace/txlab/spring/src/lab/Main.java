package lab;

import com.zaxxer.hikari.HikariDataSource;
import java.util.Set;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;

/**
 * 실험 5의 앱입니다. 장면 이름 하나를 받아 Spring을 띄우고, 그 장면의 바깥 메서드를 부른 뒤 끝냅니다. 바깥 메서드는 한 번
 * 부르고, {@code reborrow}만 같은 메서드를 멈춤 없이 두 번 잇달아 부릅니다.
 *
 * <p>멈춤 자리: {@code start}(풀이 다 채워진 뒤), 장면 안의 자리들, {@code end}(바깥 메서드가 끝난 뒤, 풀을 닫기 전),
 * {@code closed}(Spring과 풀을 닫은 뒤). 바깥 메서드가 던진 예외는 {@code outer-threw}로 알리고 삼킵니다. 예외도 관찰할
 * 결과의 하나이기 때문입니다.
 */
public final class Main {

    private static final Set<String> SCENES = Set.of("required", "requires-new", "requires-new-same-row", "nested", "reborrow");

    private Main() {
    }

    /**
     * 장면 하나를 돌립니다.
     *
     * @param args 장면 이름 하나(required, requires-new, requires-new-same-row, nested, reborrow)
     * @throws IllegalArgumentException 장면 이름이 없거나 모를 때(Spring을 띄우기 전에 멈춤)
     * @throws InterruptedException 풀이 채워지기를 기다리다 끊겼을 때
     */
    public static void main(String[] args) throws InterruptedException {
        if (args.length != 1 || !SCENES.contains(args[0])) {
            throw new IllegalArgumentException("장면 이름 하나가 필요합니다: " + SCENES);
        }
        String scene = args[0];
        try (AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext(LabConfig.class)) {
            waitPoolFilled(ctx.getBean(HikariDataSource.class));
            Lab.at("start");
            Outer outer = ctx.getBean(Outer.class);
            try {
                switch (scene) {
                    case "required" -> outer.required();
                    case "requires-new" -> outer.requiresNew();
                    case "requires-new-same-row" -> outer.requiresNewSameRow();
                    case "nested" -> outer.nested();
                    case "reborrow" -> {
                        // 두 호출 사이에 멈춤 자리를 두지 않습니다. 멈춤(관찰 약 0.4초)이 그대로 연결이 쉬는 시간이 되어,
                        // 둘째 빌리기가 HikariCP의 500 ms 안에 오게 하려는 이 장면의 뜻이 없어집니다(WORK 실험 5 「덧붙인 장면 5-reborrow」).
                        outer.once();
                        outer.once();
                    }
                    default -> throw new AssertionError(scene);
                }
            } catch (RuntimeException e) {
                Lab.note("outer-threw", e);
            }
            Lab.at("end");
        }
        Lab.at("closed");
    }

    /**
     * 풀이 처음 채울 연결을 모두 열 때까지 기다립니다. HikariCP는 첫 연결만 풀을 만들 때 열고 나머지는 따로 도는 스레드가
     * 엽니다. 다 열리기 전에 장면을 시작하면, 연결을 여는 바이트가 장면의 바이트와 섞입니다.
     *
     * @throws IllegalStateException 10초 안에 다 열리지 않을 때
     */
    private static void waitPoolFilled(HikariDataSource ds) throws InterruptedException {
        int want = ds.getMinimumIdle();
        for (int i = 0; i < 200; i++) {
            if (ds.getHikariPoolMXBean().getIdleConnections() >= want) {
                return;
            }
            Thread.sleep(50);
        }
        throw new IllegalStateException("풀이 10초 안에 연결 " + want + "개를 열지 못했습니다");
    }
}
