package lld.threadpool;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/**
 * 스레드 풀의 공용 계약 테스트. 스레드가 얽히므로 latch·barrier·timeout으로 결정적으로 짠다
 * (sleep으로 타이밍을 맞추지 않는다).
 */
abstract class ThreadPoolContractTest {

    protected abstract ThreadPool newPool(int poolSize);

    @Test
    @DisplayName("제출한 작업이 모두 실행된다")
    void runs_all_submitted_tasks() throws InterruptedException {
        ThreadPool pool = newPool(3);
        int n = 200;
        CountDownLatch done = new CountDownLatch(n);
        AtomicInteger counter = new AtomicInteger();
        for (int i = 0; i < n; i++) {
            pool.execute(() -> {
                counter.incrementAndGet();
                done.countDown();
            });
        }
        assertTrue(done.await(2, TimeUnit.SECONDS), "모든 작업이 끝나야 한다");
        assertEquals(n, counter.get());
        pool.shutdown();
    }

    @Test
    @DisplayName("작업을 동시에 실행한다(워커 수만큼 병렬)")
    void executes_tasks_concurrently() throws InterruptedException {
        int size = 3;
        ThreadPool pool = newPool(size);
        CyclicBarrier barrier = new CyclicBarrier(size); // size개가 동시에 도달해야 풀린다
        CountDownLatch done = new CountDownLatch(size);
        for (int i = 0; i < size; i++) {
            pool.execute(() -> {
                try {
                    barrier.await(1, TimeUnit.SECONDS);
                    done.countDown();
                } catch (Exception ignored) {
                    // 직렬 실행이면 barrier가 안 풀려 timeout -> done 미감소 -> 아래 await 실패
                }
            });
        }
        assertTrue(done.await(2, TimeUnit.SECONDS), "동시에 실행되면 barrier가 풀린다");
        pool.shutdown();
    }

    @Test
    @DisplayName("종료 후 제출은 거부된다")
    void rejects_after_shutdown() {
        ThreadPool pool = newPool(2);
        pool.shutdown();
        assertThrows(IllegalStateException.class, () -> pool.execute(() -> { }));
    }

    @Test
    @DisplayName("graceful 종료: 큐에 남은 작업을 마저 처리한다")
    void shutdown_drains_pending_tasks() throws InterruptedException {
        ThreadPool pool = newPool(1);
        int n = 10;
        AtomicInteger counter = new AtomicInteger();
        for (int i = 0; i < n; i++) {
            pool.execute(counter::incrementAndGet);
        }
        pool.shutdown();
        assertTrue(pool.awaitTermination(Duration.ofSeconds(2)), "종료를 기다린다");
        assertEquals(n, counter.get(), "남은 작업이 모두 실행돼야 한다");
    }

    @Test
    @DisplayName("작업 예외가 워커를 죽이지 않는다")
    void task_exception_does_not_kill_worker() throws InterruptedException {
        ThreadPool pool = newPool(1);
        CountDownLatch done = new CountDownLatch(1);
        pool.execute(() -> {
            throw new RuntimeException("boom");
        });
        pool.execute(done::countDown);
        assertTrue(done.await(2, TimeUnit.SECONDS), "예외 작업 뒤에도 다음 작업이 실행돼야 한다");
        pool.shutdown();
    }

    @Test
    @DisplayName("실패: 풀 크기가 0 이하면 예외")
    void invalid_pool_size_rejected() {
        assertThrows(IllegalArgumentException.class, () -> newPool(0));
        assertThrows(IllegalArgumentException.class, () -> newPool(-1));
    }
}
