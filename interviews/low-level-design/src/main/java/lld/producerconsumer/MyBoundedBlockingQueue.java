package lld.producerconsumer;

/**
 * 직접 구현하는 공간. {@link BoundedBlockingQueue} 헤더 카드와 테스트만 보고 채운다.
 * 막히면 그때만 {@link ReferenceBoundedBlockingQueue}를 연다.
 *
 * <p>핵심 질문: 왜 if가 아니라 while로 조건을 확인하나? notFull/notEmpty를 왜 나누나?
 * 시작하려면 TODO를 채우고 {@code MyBoundedBlockingQueueTest}의 {@code @Disabled}를 지운다.
 */
public final class MyBoundedBlockingQueue<E> implements BoundedBlockingQueue<E> {

    public MyBoundedBlockingQueue(int capacity) {
        // TODO: capacity 검증 + 배열/락/두 Condition 초기화
    }

    @Override
    public void put(E element) throws InterruptedException {
        // TODO: lock -> while(가득) notFull.await() -> 넣기 -> notEmpty.signal()
        throw new UnsupportedOperationException("아직 구현 전: MyBoundedBlockingQueue.put");
    }

    @Override
    public E take() throws InterruptedException {
        // TODO: lock -> while(빔) notEmpty.await() -> 꺼내기 -> notFull.signal()
        throw new UnsupportedOperationException("아직 구현 전: MyBoundedBlockingQueue.take");
    }

    @Override
    public int size() {
        // TODO: 락 보유 중 현재 개수 반환
        throw new UnsupportedOperationException("아직 구현 전: MyBoundedBlockingQueue.size");
    }
}
