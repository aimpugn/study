package lld.threadpool;

import java.time.Duration;

/**
 * 직접 구현하는 공간. {@link ThreadPool} 헤더 카드와 테스트만 보고 채운다.
 * 막히면 그때만 {@link ReferenceThreadPool}을 연다.
 *
 * <p>핵심 설계 질문: 워커 루프는 어떻게? graceful 종료는 어떻게(poison pill? 플래그+poll?)?
 * 작업 예외를 어떻게 격리? 시작하려면 TODO를 채우고 {@code MyThreadPoolTest}의 {@code @Disabled}를 지운다.
 */
public final class MyThreadPool implements ThreadPool {

    public MyThreadPool(int poolSize) {
        // TODO: poolSize 검증 + 작업 큐 만들고 워커 스레드 N개 시작
    }

    @Override
    public void execute(Runnable task) {
        // TODO: 종료 상태면 거부, 아니면 큐에 넣기
        throw new UnsupportedOperationException("아직 구현 전: MyThreadPool.execute");
    }

    @Override
    public void shutdown() {
        // TODO: 새 작업 거부 표시 + 워커가 빠져나오게(예: poison pill)
        throw new UnsupportedOperationException("아직 구현 전: MyThreadPool.shutdown");
    }

    @Override
    public boolean awaitTermination(Duration timeout) throws InterruptedException {
        // TODO: 모든 워커 종료를 timeout만큼 대기 (CountDownLatch 등)
        throw new UnsupportedOperationException("아직 구현 전: MyThreadPool.awaitTermination");
    }
}
