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
 * Netty가 selector loop, thread handoff, pipeline을 어떻게 정리해 주는지 보여 주는 최소 echo 서버입니다.
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
        bossGroup = new NioEventLoopGroup(1, new DefaultThreadFactory("netty-boss"));
        workerGroup = new NioEventLoopGroup(0, new DefaultThreadFactory("netty-worker"));

        ServerBootstrap bootstrap = new ServerBootstrap()
                .group(bossGroup, workerGroup)
                .channel(NioServerSocketChannel.class)
                .handler(new BossAcceptLogHandler())
                .childHandler(new ChannelInitializer<SocketChannel>() {
                    @Override
                    protected void initChannel(SocketChannel channel) {
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

    private final class EchoHandler extends SimpleChannelInboundHandler<String> {

        @Override
        public void channelActive(ChannelHandlerContext context) {
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

    private final class BossAcceptLogHandler extends ChannelInboundHandlerAdapter {

        @Override
        public void channelRead(ChannelHandlerContext context, Object message) throws Exception {
            observationLog.event("boss-accept", "client-pending", "accepted child channel=" + message);
            super.channelRead(context, message);
        }
    }
}
