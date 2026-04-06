package io.aimpugn.learn.netty.lesson1.common;

/**
 * 첫 강의의 서로 다른 서버 구현을 같은 테스트/실행 틀 안에 올려 두기 위한 최소 계약입니다.
 *
 * 비교 학습의 핵심은 "문제를 같게 두고 구현만 바꿔 본다"는 점이므로, 서버 이름/포트/종료 방식만 공통으로 맞춥니다.
 */
public interface LessonServer extends AutoCloseable {

    String serverName();

    int port();

    @Override
    void close() throws Exception;
}
