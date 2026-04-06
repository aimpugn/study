package io.aimpugn.learn.netty.lesson1.blocking;

import io.aimpugn.learn.netty.lesson1.common.Lesson1Support;
import io.aimpugn.learn.netty.lesson1.common.LessonServer;
import io.aimpugn.learn.netty.lesson1.common.ObservationLog;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketException;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 첫 강의에서 "연결마다 기다리는 모델"을 보여 주는 blocking echo 서버입니다.
 *
 * 같은 echo 문제를 풀더라도 이 구현은 client마다 worker 하나를 붙이고 그 worker가 {@code readLine()}에서
 * 계속 기다립니다. 그래서 코드 흐름은 가장 읽기 쉽지만, 연결 수가 늘수록 thread도 함께 늘어나는 모습을
 * 로그에서 바로 확인할 수 있습니다.
 */
public final class BlockingEchoServer implements LessonServer {

    private final ObservationLog observationLog = new ObservationLog("blocking");
    // 새 연결이 들어올 때마다 client-1, client-2처럼 번호를 붙여,
    // 로그와 설명 문서에서 같은 연결을 계속 따라갈 수 있게 합니다.
    private final AtomicInteger connectionIds = new AtomicInteger();
    // blocking 모델은 연결마다 worker thread가 하나씩 붙는다는 점이 핵심 비교 포인트이므로,
    // thread 이름에 worker 번호를 넣어 "연결 수가 늘면 기다리는 worker도 늘어난다"는 사실을 드러냅니다.
    private final AtomicInteger workerThreadIds = new AtomicInteger();
    private final int requestedPort;
    private final ExecutorService connectionExecutor = Executors.newCachedThreadPool(task ->
            Thread.ofPlatform()
                    .name("blocking-client-" + workerThreadIds.incrementAndGet())
                    .unstarted(task));

    private volatile boolean running;
    private ServerSocket serverSocket;
    private Thread acceptThread;

    private BlockingEchoServer(int requestedPort) {
        this.requestedPort = requestedPort;
    }

    public static BlockingEchoServer start(int port) throws IOException {
        BlockingEchoServer server = new BlockingEchoServer(port);
        server.startInternal();
        return server;
    }

    private void startInternal() throws IOException {
        serverSocket = new ServerSocket(requestedPort);
        running = true;
        observationLog.event("server-start", "server", "bound to port=" + serverSocket.getLocalPort());
        // accept 전용 thread를 따로 두어 "연결 수락"과 "각 client 처리"가 로그에서 분리되어 보이게 합니다.
        acceptThread = Thread.ofPlatform().name("blocking-accept").start(this::acceptLoop);
    }

    private void acceptLoop() {
        while (running) {
            try {
                Socket socket = serverSocket.accept();
                int connectionId = connectionIds.incrementAndGet();
                String channel = channelName(connectionId);
                observationLog.event("accept", channel, "accepted " + socket.getRemoteSocketAddress());
                // blocking 모델의 핵심은 여기입니다. client 하나마다 worker를 하나 넘기고,
                // 그 worker는 자기 client만 붙잡고 readLine()에서 기다립니다.
                connectionExecutor.submit(() -> handleConnection(socket, connectionId));
            } catch (SocketException socketException) {
                if (running) {
                    observationLog.event("accept-error", "server", socketException.toString());
                }
                return;
            } catch (IOException ioException) {
                observationLog.event("accept-error", "server", ioException.toString());
            }
        }
    }

    private void handleConnection(Socket socket, int connectionId) {
        String channel = channelName(connectionId);

        // 이 worker가 맡은 socket과 그 위에 얹힌 reader/writer를 한 scope 안에서 같이 닫습니다.
        // blocking 예제에서는 "한 worker가 한 연결을 독점하고, 끝나면 여기서 한 번에 정리된다"는 점이 중요합니다.
        try (socket;
             BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), StandardCharsets.UTF_8))) {

            observationLog.event("connection-start", channel, "worker attached");

            String line;
            // 이 loop는 "한 worker가 한 client를 계속 기다린다"는 blocking 모델을 가장 직접적으로 보여 줍니다.
            // NIO 예제는 같은 일을 selector thread 하나가 번갈아 처리하고, Netty 예제는 EventLoop + pipeline이 맡습니다.
            while ((line = reader.readLine()) != null) {
                observationLog.event("read", channel, "line=\"" + ObservationLog.previewText(line) + "\"");
                writer.write(line);
                writer.newLine();
                writer.flush();
                observationLog.event("write", channel, "echoed line=\"" + ObservationLog.previewText(line) + "\"");
            }

            observationLog.event("client-close", channel, "peer closed connection");
        } catch (IOException ioException) {
            observationLog.event("connection-error", channel, ioException.toString());
        }
    }

    @Override
    public String serverName() {
        return "blocking";
    }

    @Override
    public int port() {
        return serverSocket.getLocalPort();
    }

    @Override
    public void close() throws Exception {
        running = false;

        if (serverSocket != null && !serverSocket.isClosed()) {
            serverSocket.close();
        }
        if (acceptThread != null && acceptThread != Thread.currentThread()) {
            acceptThread.join(2_000);
        }

        connectionExecutor.shutdownNow();
        connectionExecutor.awaitTermination(2, TimeUnit.SECONDS);
        observationLog.event("server-stop", "server", "closed");
    }

    private static String channelName(int connectionId) {
        return "client-" + connectionId;
    }

    public static void main(String[] args) throws Exception {
        int port = Lesson1Support.parsePort(args, 9001);
        try (BlockingEchoServer server = BlockingEchoServer.start(port)) {
            Lesson1Support.keepRunning(server);
        }
    }
}
