package lld.producerconsumer;

/** 정답지를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceBoundedBlockingQueueTest extends BoundedBlockingQueueContractTest {

    @Override
    protected <E> BoundedBlockingQueue<E> newQueue(int capacity) {
        return new ReferenceBoundedBlockingQueue<>(capacity);
    }
}
