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
 * 첫 강의에서 "한 selector thread가 여러 연결을 번갈아 확인하는 모델"을 보여 주는 NIO echo 서버입니다.
 *
 * blocking 예제와 비교하면 client마다 worker를 늘리지 않고, selector 하나가 accept/read 준비가 된 연결만
 * 순서대로 집어 듭니다. 그래서 로그를 보면 대부분의 이벤트가 같은 {@code nio-selector} thread에서
 * 찍히고, line 경계와 write loop도 우리가 직접 처리해야 한다는 차이가 드러납니다.
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
        // 첫 강의에서는 selector thread가 하나뿐이므로 read buffer 하나를 재사용해도 됩니다.
        // 이 단순화 덕분에 "연결마다 thread를 두지 않아도 된다"는 차이에 집중할 수 있습니다.
        ByteBuffer readBuffer = ByteBuffer.allocate(1024);

        while (running) {
            try {
                // selector는 "지금 바로 처리할 수 있는 채널"이 생길 때까지 기다립니다.
                // blocking 예제처럼 client마다 잠드는 것이 아니라, 한 thread가 준비된 일만 번갈아 집어 듭니다.
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

        // non-blocking read는 한 줄이 한 번에 오지 않을 수 있으므로 line 경계를 직접 조립해야 합니다.
        // blocking 예제는 readLine()이 이 일을 숨겨 주고, Netty 예제는 LineBasedFrameDecoder가 맡습니다.
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
        // SocketChannel.write(...)는 한 번에 다 못 쓸 수 있으므로, 첫 강의에서는 직접 끝까지 밀어 넣습니다.
        // Netty 예제의 writeAndFlush()가 실제로 이런 종류의 반복과 상태 관리를 감싸 준다는 점이 비교 포인트입니다.
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

    /**
     * selector 기반 서버는 read 경계가 애매하게 끊길 수 있으므로, client별로 "아직 줄이 완성되지 않은 바이트"를
     * 잠깐 들고 있어야 합니다. 이 작은 상태 객체가 blocking의 readLine()과 Netty의 decoder 사이를 메워 줍니다.
     */
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
