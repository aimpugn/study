package io.aimpugn.learn.netty.lesson1;

import io.aimpugn.learn.netty.lesson1.blocking.BlockingEchoServer;
import io.aimpugn.learn.netty.lesson1.common.LessonServer;
import io.aimpugn.learn.netty.lesson1.netty.NettyEchoServer;
import io.aimpugn.learn.netty.lesson1.nio.NioSelectorEchoServer;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.util.List;
import java.util.stream.Stream;

class Lesson1EchoServersTest {

    @ParameterizedTest(name = "{0} echoes a single line")
    @MethodSource("serverSpecs")
    void echoesSingleLine(ServerSpec serverSpec) throws Exception {
        try (LessonServer server = serverSpec.starter().start(0)) {
            String response = EchoTestClient.exchangeSingleLine(server.port(), "hello-" + serverSpec.name());
            Assertions.assertEquals("hello-" + serverSpec.name(), response);
        }
    }

    @ParameterizedTest(name = "{0} echoes two lines on the same connection")
    @MethodSource("serverSpecs")
    void echoesTwoLinesOnSameConnection(ServerSpec serverSpec) throws Exception {
        try (LessonServer server = serverSpec.starter().start(0)) {
            List<String> responses = EchoTestClient.exchangeTwoLines(server.port(), "alpha-" + serverSpec.name(), "beta-" + serverSpec.name());
            Assertions.assertEquals(List.of("alpha-" + serverSpec.name(), "beta-" + serverSpec.name()), responses);
        }
    }

    @ParameterizedTest(name = "{0} survives a connect-close client")
    @MethodSource("serverSpecs")
    void survivesClientDisconnectWithoutData(ServerSpec serverSpec) throws Exception {
        try (LessonServer server = serverSpec.starter().start(0)) {
            EchoTestClient.connectAndCloseWithoutSending(server.port());
            String response = EchoTestClient.exchangeSingleLine(server.port(), "after-close-" + serverSpec.name());
            Assertions.assertEquals("after-close-" + serverSpec.name(), response);
        }
    }

    @ParameterizedTest(name = "{0} fails when the port is already in use")
    @MethodSource("serverSpecs")
    void failsToStartOnUsedPort(ServerSpec serverSpec) throws Exception {
        try (LessonServer runningServer = serverSpec.starter().start(0)) {
            Assertions.assertThrows(Exception.class, () -> serverSpec.starter().start(runningServer.port()));
        }
    }

    static Stream<ServerSpec> serverSpecs() {
        return Stream.of(
                new ServerSpec("blocking", BlockingEchoServer::start),
                new ServerSpec("nio", NioSelectorEchoServer::start),
                new ServerSpec("netty", NettyEchoServer::start)
        );
    }

    private record ServerSpec(String name, ServerStarter starter) {
        @Override
        public String toString() {
            return name;
        }
    }

    @FunctionalInterface
    private interface ServerStarter {
        LessonServer start(int port) throws Exception;
    }
}
