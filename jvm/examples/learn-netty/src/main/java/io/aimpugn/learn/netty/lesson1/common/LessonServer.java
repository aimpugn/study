package io.aimpugn.learn.netty.lesson1.common;

/**
 * 각 예제 서버가 테스트와 실행 환경에서 공통으로 제공하는 최소 계약입니다.
 */
public interface LessonServer extends AutoCloseable {

    String serverName();

    int port();

    @Override
    void close() throws Exception;
}
