package lld.producerconsumer;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/**
 * 블로킹 큐의 공용 계약 테스트. '대기'를 검증하려고 별도 스레드를 띄우고,
 * 짧은 timeout으로 '아직 안 깨어났음(blocked)'을 확인한 뒤 신호를 준다.
 */
abstract class BoundedBlockingQueueContractTest {

    protected abstract <E> BoundedBlockingQueue<E> newQueue(int capacity);

    @Test
    @DisplayName("기본: 넣은 순서대로 꺼낸다(FIFO)")
    void put_then_take_fifo() throws InterruptedException {
        BoundedBlockingQueue<Integer> q = newQueue(3);
        q.put(1);
        q.put(2);
        q.put(3);
        assertEquals(3, q.size());
        assertEquals(1, q.take());
        assertEquals(2, q.take());
        assertEquals(3, q.take());
        assertEquals(0, q.size());
    }

    @Test
    @DisplayName("비면 take가 대기하다가 put 되면 깨어난다")
    void take_blocks_until_put() throws InterruptedException {
        BoundedBlockingQueue<Integer> q = newQueue(2);
        AtomicReference<Integer> got = new AtomicReference<>();
        CountDownLatch consumed = new CountDownLatch(1);
        Thread consumer = new Thread(() -> {
            try {
                got.set(q.take());
                consumed.countDown();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        });
        consumer.start();
        assertFalse(consumed.await(100, TimeUnit.MILLISECONDS), "큐가 비면 take는 대기해야 한다");
        q.put(42);
        assertTrue(consumed.await(1, TimeUnit.SECONDS));
        assertEquals(42, got.get());
        consumer.join(1000);
    }

    @Test
    @DisplayName("가득 차면 put이 대기하다가 take 되면 깨어난다")
    void put_blocks_when_full() throws InterruptedException {
        BoundedBlockingQueue<Integer> q = newQueue(1);
        q.put(1);
        CountDownLatch produced = new CountDownLatch(1);
        Thread producer = new Thread(() -> {
            try {
                q.put(2);
                produced.countDown();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        });
        producer.start();
        assertFalse(produced.await(100, TimeUnit.MILLISECONDS), "가득 차면 put은 대기해야 한다");
        assertEquals(1, q.take());
        assertTrue(produced.await(1, TimeUnit.SECONDS));
        assertEquals(2, q.take());
        producer.join(1000);
    }

    @Test
    @DisplayName("size가 보관 개수를 반영한다")
    void size_reflects_count() throws InterruptedException {
        BoundedBlockingQueue<Integer> q = newQueue(2);
        assertEquals(0, q.size());
        q.put(1);
        assertEquals(1, q.size());
        q.put(2);
        assertEquals(2, q.size());
        q.take();
        assertEquals(1, q.size());
    }

    @Test
    @DisplayName("실패: 용량이 0 이하면 예외")
    void invalid_capacity_rejected() {
        assertThrows(IllegalArgumentException.class, () -> newQueue(0));
        assertThrows(IllegalArgumentException.class, () -> newQueue(-1));
    }

    @Test
    @DisplayName("실패: null 원소는 거부")
    void null_element_rejected() {
        BoundedBlockingQueue<Integer> q = newQueue(2);
        assertThrows(NullPointerException.class, () -> q.put(null));
    }
}
