package io.aimpugn.learn.netty.lesson1.common;

import java.util.concurrent.CountDownLatch;

/**
 * 첫 강의의 세 서버를 같은 방식으로 실행하게 맞춰, 비교 포인트가 "부팅 방식"이 아니라
 * "요청을 처리하는 모델 자체"에 머물도록 돕는 공통 지원 코드입니다.
 */
public final class Lesson1Support {

    private Lesson1Support() {
    }

    public static int parsePort(String[] args, int defaultPort) {
        if (args.length == 0) {
            return defaultPort;
        }
        if (args.length != 1) {
            throw new IllegalArgumentException("포트는 하나만 넘겨야 합니다. 예: 9001");
        }
        return Integer.parseInt(args[0]);
    }

    public static void keepRunning(LessonServer server) throws InterruptedException {
        Runtime.getRuntime().addShutdownHook(Thread.ofPlatform()
                .name(server.serverName() + "-shutdown")
                .unstarted(() -> closeQuietly(server)));

        // 세 서버가 모두 같은 ready 메시지를 쓰게 맞춰 두면, 학습자는 포트만 바꿔 가며
        // blocking/NIO/Netty의 로그 차이에 집중할 수 있습니다.
        System.out.printf("[%s] ready on port=%d. Stop with Ctrl+C.%n", server.serverName(), server.port());
        new CountDownLatch(1).await();
    }

    private static void closeQuietly(AutoCloseable closeable) {
        try {
            closeable.close();
        } catch (Exception ignored) {
            // Shutdown hook에서는 종료를 방해하지 않도록 조용히 닫습니다.
        }
    }
}
