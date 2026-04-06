package io.aimpugn.learn.netty.lesson1.netty;

import io.aimpugn.learn.netty.lesson1.common.Lesson1Support;
import io.aimpugn.learn.netty.lesson1.common.LessonServer;
import io.aimpugn.learn.netty.lesson1.common.ObservationLog;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelInboundHandlerAdapter;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.LineBasedFrameDecoder;
import io.netty.handler.codec.string.StringDecoder;
import io.netty.handler.codec.string.StringEncoder;
import io.netty.util.AttributeKey;
import io.netty.util.concurrent.DefaultThreadFactory;

import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 첫 강의에서 "같은 echo 문제를 Netty는 어떤 추상화로 정리하는가"를 보여 주는 최소 서버입니다.
 *
 * NIO 예제에서는 selector loop, line 경계 조립, string 변환, write 반복을 우리가 직접 적었습니다.
 * 여기서는 boss/worker EventLoop와 pipeline이 그 일을 나눠 맡아, business handler가 더 짧아지는 대신
 * 내부에서 어떤 handoff가 일어나는지 로그와 주석으로 따라가야 합니다.
 */
public final class NettyEchoServer implements LessonServer {

    private static final AttributeKey<Integer> CONNECTION_ID = AttributeKey.valueOf("lesson1.connectionId");

    private final ObservationLog observationLog = new ObservationLog("netty");
    private final AtomicInteger connectionIds = new AtomicInteger();
    private final int requestedPort;

    private NioEventLoopGroup bossGroup;
    private NioEventLoopGroup workerGroup;
    private Channel serverChannel;

    private NettyEchoServer(int requestedPort) {
        this.requestedPort = requestedPort;
    }

    public static NettyEchoServer start(int port) throws Exception {
        NettyEchoServer server = new NettyEchoServer(port);
        server.startInternal();
        return server;
    }

    private void startInternal() throws Exception {
        // 첫 강의에서는 accept 담당과 client IO 담당을 분리해, boss/worker 차이가 thread 이름으로 바로 보이게 합니다.
        bossGroup = new NioEventLoopGroup(1, new DefaultThreadFactory("netty-boss"));
        workerGroup = new NioEventLoopGroup(0, new DefaultThreadFactory("netty-worker"));

        ServerBootstrap bootstrap = new ServerBootstrap()
                .group(bossGroup, workerGroup)
                .channel(NioServerSocketChannel.class)
                .handler(new BossAcceptLogHandler())
                .childHandler(new ChannelInitializer<SocketChannel>() {
                    @Override
                    protected void initChannel(SocketChannel channel) {
                        // 첫 강의의 핵심 비교 포인트입니다.
                        // NIO에서 직접 하던 "줄 경계 찾기 -> 문자열 decode -> 문자열 encode"를
                        // Netty는 pipeline handler 체인으로 잘게 나눠 붙입니다.
                        channel.pipeline()
                                .addLast("lineDecoder", new LineBasedFrameDecoder(1024))
                                .addLast("stringDecoder", new StringDecoder(StandardCharsets.UTF_8))
                                .addLast("stringEncoder", new StringEncoder(StandardCharsets.UTF_8))
                                .addLast("echoHandler", new EchoHandler());
                    }
                });

        var bindFuture = bootstrap.bind(requestedPort).awaitUninterruptibly();
        if (!bindFuture.isSuccess()) {
            Throwable cause = bindFuture.cause();
            close();
            if (cause instanceof Exception exception) {
                throw exception;
            }
            throw new IllegalStateException("Netty 서버 바인딩에 실패했습니다.", cause);
        }

        serverChannel = bindFuture.channel();
        observationLog.event("server-start", "server", "bound to port=" + port());
    }

    @Override
    public String serverName() {
        return "netty";
    }

    @Override
    public int port() {
        return ((InetSocketAddress) serverChannel.localAddress()).getPort();
    }

    @Override
    public void close() {
        if (serverChannel != null) {
            serverChannel.close().awaitUninterruptibly();
        }
        if (bossGroup != null) {
            bossGroup.shutdownGracefully(0, 5, TimeUnit.SECONDS).awaitUninterruptibly();
        }
        if (workerGroup != null) {
            workerGroup.shutdownGracefully(0, 5, TimeUnit.SECONDS).awaitUninterruptibly();
        }
        observationLog.event("server-stop", "server", "closed");
    }

    public static void main(String[] args) throws Exception {
        int port = Lesson1Support.parsePort(args, 9003);
        try (NettyEchoServer server = NettyEchoServer.start(port)) {
            Lesson1Support.keepRunning(server);
        }
    }

    /**
     * pipeline의 마지막 business handler입니다.
     *
     * 앞단 decoder/encoder 덕분에 여기서는 이미 "한 줄 문자열"만 다루면 됩니다.
     * 첫 강의에서는 이 단순함 자체가 비교 포인트이므로, NIO 예제에서 직접 처리하던 일들이
     * 앞선 handler로 이동했다는 사실을 같이 봐야 합니다.
     */
    private final class EchoHandler extends SimpleChannelInboundHandler<String> {

        @Override
        public void channelActive(ChannelHandlerContext context) {
            // child channel의 실제 처리 owner가 worker thread로 정해진 뒤에 connection id를 붙입니다.
            // 그래서 boss-accept 로그와 channel-active 로그를 나란히 보면 handoff 시점을 비교할 수 있습니다.
            int connectionId = connectionIds.incrementAndGet();
            context.channel().attr(CONNECTION_ID).set(connectionId);
            observationLog.event("channel-active", channelName(context), "remote=" + context.channel().remoteAddress());
            context.fireChannelActive();
        }

        @Override
        protected void channelRead0(ChannelHandlerContext context, String message) {
            observationLog.event("read", channelName(context), "line=\"" + ObservationLog.previewText(message) + "\"");
            String response = message + "\n";
            observationLog.event("write", channelName(context), "echoed line=\"" + ObservationLog.previewText(message) + "\"");
            // Netty는 writeAndFlush 뒤에서 outbound 경로와 실제 channel write를 이어 줍니다.
            // 첫 강의에서는 NIO처럼 직접 while(write)를 돌리지 않는다는 차이를 눈여겨봅니다.
            context.writeAndFlush(response);
        }

        @Override
        public void channelInactive(ChannelHandlerContext context) {
            observationLog.event("channel-inactive", channelName(context), "peer closed connection");
            context.fireChannelInactive();
        }

        @Override
        public void exceptionCaught(ChannelHandlerContext context, Throwable cause) {
            observationLog.event("channel-error", channelName(context), cause.toString());
            context.close();
        }

        private String channelName(ChannelHandlerContext context) {
            Integer connectionId = context.channel().attr(CONNECTION_ID).get();
            if (connectionId == null) {
                return "client-pending";
            }
            return "client-" + connectionId;
        }
    }

    /**
     * business 기능에는 꼭 필요하지 않지만, 첫 강의에서는 accept가 boss thread에서 먼저 보이고
     * 이후 child channel 처리가 worker thread로 넘어간다는 차이를 드러내기 위해 둔 관측용 handler입니다.
     */
    private final class BossAcceptLogHandler extends ChannelInboundHandlerAdapter {

        @Override
        public void channelRead(ChannelHandlerContext context, Object message) throws Exception {
            observationLog.event("boss-accept", "client-pending", "accepted child channel=" + message);
            super.channelRead(context, message);
        }
    }
}
