package lld.producerconsumer;

import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

/**
 * 용량 제한 블로킹 큐 정답지 - 원형 버퍼 + ReentrantLock + 두 Condition.
 *
 * <p>핵심 두 가지:
 * (1) {@code while}(조건) await - spurious wakeup과 다중 대기자 때문에 깨어난 뒤 조건을 다시 확인한다.
 * (2) notFull / notEmpty 분리 - 자리가 났을 때 생산자만, 원소가 들어왔을 때 소비자만 깨운다.
 */
public final class ReferenceBoundedBlockingQueue<E> implements BoundedBlockingQueue<E> {

    private final Object[] items;
    private int head;
    private int tail;
    private int count;

    private final ReentrantLock lock = new ReentrantLock();
    private final Condition notFull = lock.newCondition();
    private final Condition notEmpty = lock.newCondition();

    public ReferenceBoundedBlockingQueue(int capacity) {
        if (capacity <= 0) {
            throw new IllegalArgumentException("capacity must be > 0 but was " + capacity);
        }
        this.items = new Object[capacity];
    }

    @Override
    public void put(E element) throws InterruptedException {
        if (element == null) {
            throw new NullPointerException("element");
        }
        lock.lockInterruptibly();
        try {
            while (count == items.length) {
                notFull.await(); // 자리가 날 때까지 락을 놓고 잔다.
            }
            items[tail] = element;
            tail = (tail + 1) % items.length;
            count++;
            notEmpty.signal(); // 소비자 한 명을 깨운다.
        } finally {
            lock.unlock();
        }
    }

    @Override
    @SuppressWarnings("unchecked")
    public E take() throws InterruptedException {
        lock.lockInterruptibly();
        try {
            while (count == 0) {
                notEmpty.await(); // 원소가 들어올 때까지 락을 놓고 잔다.
            }
            E element = (E) items[head];
            items[head] = null; // 참조 해제로 GC 도움.
            head = (head + 1) % items.length;
            count--;
            notFull.signal(); // 생산자 한 명을 깨운다.
            return element;
        } finally {
            lock.unlock();
        }
    }

    @Override
    public int size() {
        lock.lock();
        try {
            return count;
        } finally {
            lock.unlock();
        }
    }
}
