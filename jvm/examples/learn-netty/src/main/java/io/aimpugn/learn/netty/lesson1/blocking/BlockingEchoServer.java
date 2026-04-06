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
 * 연결마다 전용 스레드를 잡고 기다리는 가장 단순한 blocking echo 서버입니다.
 */
public final class BlockingEchoServer implements LessonServer {

    private final ObservationLog observationLog = new ObservationLog("blocking");
    private final AtomicInteger connectionIds = new AtomicInteger();
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
        acceptThread = Thread.ofPlatform().name("blocking-accept").start(this::acceptLoop);
    }

    private void acceptLoop() {
        while (running) {
            try {
                Socket socket = serverSocket.accept();
                int connectionId = connectionIds.incrementAndGet();
                String channel = channelName(connectionId);
                observationLog.event("accept", channel, "accepted " + socket.getRemoteSocketAddress());
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

        try (socket;
             BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), StandardCharsets.UTF_8))) {

            observationLog.event("connection-start", channel, "worker attached");

            String line;
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
