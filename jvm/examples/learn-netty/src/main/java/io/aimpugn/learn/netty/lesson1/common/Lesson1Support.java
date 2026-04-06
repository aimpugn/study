package io.aimpugn.learn.netty.lesson1.common;

import java.util.concurrent.CountDownLatch;

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
