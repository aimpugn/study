package io.aimpugn.learn.netty.lesson1.nio;

import io.aimpugn.learn.netty.lesson1.common.Lesson1Support;
import io.aimpugn.learn.netty.lesson1.common.LessonServer;
import io.aimpugn.learn.netty.lesson1.common.ObservationLog;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.ClosedSelectorException;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.charset.StandardCharsets;
import java.util.Iterator;
import java.util.Set;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * selector 한 개가 여러 연결을 번갈아 확인하는 가장 작은 NIO 예제입니다.
 */
public final class NioSelectorEchoServer implements LessonServer {

    private final ObservationLog observationLog = new ObservationLog("nio");
    private final AtomicInteger connectionIds = new AtomicInteger();
    private final int requestedPort;

    private volatile boolean running;
    private Selector selector;
    private ServerSocketChannel serverSocketChannel;
    private Thread selectorThread;

    private NioSelectorEchoServer(int requestedPort) {
        this.requestedPort = requestedPort;
    }

    public static NioSelectorEchoServer start(int port) throws IOException {
        NioSelectorEchoServer server = new NioSelectorEchoServer(port);
        server.startInternal();
        return server;
    }

    private void startInternal() throws IOException {
        selector = Selector.open();
        serverSocketChannel = ServerSocketChannel.open();
        serverSocketChannel.configureBlocking(false);
        serverSocketChannel.bind(new InetSocketAddress(requestedPort));
        serverSocketChannel.register(selector, SelectionKey.OP_ACCEPT);
        running = true;
        observationLog.event("server-start", "server", "bound to port=" + port());
        selectorThread = Thread.ofPlatform().name("nio-selector").start(this::runEventLoop);
    }

    private void runEventLoop() {
        ByteBuffer readBuffer = ByteBuffer.allocate(1024);

        while (running) {
            try {
                selector.select();
                if (!running || !selector.isOpen()) {
                    return;
                }
                Set<SelectionKey> selectedKeys = selector.selectedKeys();
                Iterator<SelectionKey> iterator = selectedKeys.iterator();
                while (iterator.hasNext()) {
                    SelectionKey key = iterator.next();
                    iterator.remove();

                    if (!key.isValid()) {
                        continue;
                    }
                    if (key.isAcceptable()) {
                        acceptClient();
                        continue;
                    }
                    if (key.isReadable()) {
                        readClient(key, readBuffer);
                    }
                }
            } catch (IOException ioException) {
                if (running) {
                    observationLog.event("selector-error", "server", ioException.toString());
                }
                return;
            } catch (ClosedSelectorException ignored) {
                return;
            }
        }
    }

    private void acceptClient() throws IOException {
        SocketChannel socketChannel = serverSocketChannel.accept();
        if (socketChannel == null) {
            return;
        }

        socketChannel.configureBlocking(false);
        int connectionId = connectionIds.incrementAndGet();
        ClientState state = new ClientState(connectionId);
        socketChannel.register(selector, SelectionKey.OP_READ, state);
        observationLog.event("accept", state.channelName(), "accepted " + socketChannel.getRemoteAddress());
    }

    private void readClient(SelectionKey key, ByteBuffer readBuffer) throws IOException {
        SocketChannel channel = (SocketChannel) key.channel();
        ClientState state = (ClientState) key.attachment();

        readBuffer.clear();
        int bytesRead = channel.read(readBuffer);
        if (bytesRead == -1) {
            observationLog.event("client-close", state.channelName(), "peer closed connection");
            closeChannel(channel, key);
            return;
        }

        if (bytesRead == 0) {
            return;
        }

        observationLog.event("read-bytes", state.channelName(), "bytes=" + bytesRead);
        readBuffer.flip();

        while (readBuffer.hasRemaining()) {
            byte next = readBuffer.get();
            if (next == '\n') {
                String line = state.takeLine();
                observationLog.event("read-line", state.channelName(), "line=\"" + ObservationLog.previewText(line) + "\"");
                writeLine(channel, state, line);
                continue;
            }
            if (next != '\r') {
                state.append(next);
            }
        }
    }

    private void writeLine(SocketChannel channel, ClientState state, String line) throws IOException {
        ByteBuffer response = StandardCharsets.UTF_8.encode(line + "\n");
        while (response.hasRemaining()) {
            channel.write(response);
        }
        observationLog.event("write", state.channelName(), "echoed line=\"" + ObservationLog.previewText(line) + "\"");
    }

    private void closeChannel(SocketChannel channel, SelectionKey key) throws IOException {
        key.cancel();
        channel.close();
    }

    @Override
    public String serverName() {
        return "nio";
    }

    @Override
    public int port() {
        try {
            return ((InetSocketAddress) serverSocketChannel.getLocalAddress()).getPort();
        } catch (IOException exception) {
            throw new IllegalStateException("서버 포트를 읽을 수 없습니다.", exception);
        }
    }

    @Override
    public void close() throws Exception {
        running = false;
        if (selector != null) {
            selector.wakeup();
        }
        if (serverSocketChannel != null && serverSocketChannel.isOpen()) {
            serverSocketChannel.close();
        }
        if (selector != null && selector.isOpen()) {
            selector.close();
        }
        if (selectorThread != null && selectorThread != Thread.currentThread()) {
            selectorThread.join(2_000);
        }
        observationLog.event("server-stop", "server", "closed");
    }

    public static void main(String[] args) throws Exception {
        int port = Lesson1Support.parsePort(args, 9002);
        try (NioSelectorEchoServer server = NioSelectorEchoServer.start(port)) {
            Lesson1Support.keepRunning(server);
        }
    }

    private static final class ClientState {
        private final int connectionId;
        private final ByteArrayOutputStream lineBuffer = new ByteArrayOutputStream();

        private ClientState(int connectionId) {
            this.connectionId = connectionId;
        }

        private void append(byte next) {
            lineBuffer.write(next);
        }

        private String takeLine() {
            String line = lineBuffer.toString(StandardCharsets.UTF_8);
            lineBuffer.reset();
            return line;
        }

        private String channelName() {
            return "client-" + connectionId;
        }
    }
}
