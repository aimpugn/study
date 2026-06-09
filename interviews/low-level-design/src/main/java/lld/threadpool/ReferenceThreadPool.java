package lld.threadpool;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;

/**
 * 고정 크기 스레드 풀 정답지.
 *
 * <p>작업 큐는 표준 {@link LinkedBlockingQueue}를 쓴다(블로킹 큐 자체를 만드는 건 producerconsumer 문제다).
 * 종료는 'poison pill'(독약 작업)로 처리한다: 워커 수만큼 sentinel을 큐 끝에 넣으면, 앞의 실제 작업을
 * 모두 처리한 뒤 sentinel을 만나 워커가 스스로 빠져나가므로 graceful drain이 된다.
 */
public final class ReferenceThreadPool implements ThreadPool {

    /** 종료 신호용 sentinel. 실제로 실행되지 않는다. */
    private static final Runnable POISON = () -> { };

    private final BlockingQueue<Runnable> queue = new LinkedBlockingQueue<>();
    private final List<Thread> workers = new ArrayList<>();
    private final CountDownLatch terminated;
    private volatile boolean shutdown = false;

    public ReferenceThreadPool(int poolSize) {
        if (poolSize <= 0) {
            throw new IllegalArgumentException("poolSize must be > 0 but was " + poolSize);
        }
        this.terminated = new CountDownLatch(poolSize);
        for (int i = 0; i < poolSize; i++) {
            Thread worker = new Thread(this::runWorker, "ref-pool-worker-" + i);
            workers.add(worker);
            worker.start();
        }
    }

    @Override
    public void execute(Runnable task) {
        if (task == null) {
            throw new NullPointerException("task");
        }
        if (shutdown) {
            throw new IllegalStateException("pool is already shut down");
        }
        queue.offer(task);
    }

    @Override
    public void shutdown() {
        if (shutdown) {
            return;
        }
        shutdown = true;
        // 워커 수만큼 독약을 큐 끝에 넣는다. 앞의 실제 작업을 모두 처리한 뒤 만난다.
        for (int i = 0; i < workers.size(); i++) {
            queue.offer(POISON);
        }
    }

    @Override
    public boolean awaitTermination(Duration timeout) throws InterruptedException {
        return terminated.await(timeout.toNanos(), TimeUnit.NANOSECONDS);
    }

    private void runWorker() {
        try {
            while (true) {
                Runnable task;
                try {
                    task = queue.take();
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
                if (task == POISON) {
                    break;
                }
                runSafely(task);
            }
        } finally {
            terminated.countDown();
        }
    }

    private void runSafely(Runnable task) {
        try {
            task.run();
        } catch (Throwable ignored) {
            // 작업 예외를 삼켜 워커 스레드를 살린다. 실무라면 여기서 로깅/uncaughtExceptionHandler.
        }
    }
}
