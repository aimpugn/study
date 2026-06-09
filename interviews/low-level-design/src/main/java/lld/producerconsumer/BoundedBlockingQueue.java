package lld.producerconsumer;

/**
 * 생산자-소비자: 용량 제한 블로킹 큐 - 면접 암기 카드. 정답지는 {@link ReferenceBoundedBlockingQueue}.
 *
 * <p>직접 연습은 {@link MyBoundedBlockingQueue}. 둘 다 {@code BoundedBlockingQueueContractTest}로 검증된다.
 * 스레드 풀의 작업 큐가 바로 이것이라, 스레드 풀과 한 쌍으로 외운다.
 *
 * <pre>
 * 한 줄    : 용량이 정해진 큐. 가득 차면 put이 대기, 비면 take가 대기. 생산자-소비자 동기화의 핵심.
 * 구조     : 원형 버퍼(배열) + ReentrantLock + 두 Condition(notFull, notEmpty). (또는 synchronized+wait/notify)
 * put      : lock -&gt; while(가득) notFull.await() -&gt; 넣기 -&gt; notEmpty.signal() -&gt; unlock.
 * take     : lock -&gt; while(빔) notEmpty.await() -&gt; 꺼내기 -&gt; notFull.signal() -&gt; unlock.
 * 왜 while  : spurious wakeup + 다중 대기자 때문에 깨어난 뒤 조건을 '다시' 확인해야 한다. if로 쓰면 버그.
 * 왜 Condition 2개 : 생산자·소비자를 따로 깨워 불필요한 경쟁(thundering herd)을 줄인다. signalAll 대신 signal 가능.
 * 함정     : signal 누락 -&gt; 영구 대기. 엉뚱한 조건 깨우기 -&gt; 교착. await는 반드시 lock 보유 중에(락이 자동 해제됨).
 * 실무     : java.util.concurrent.ArrayBlockingQueue가 정확히 이 구조다.
 * 꼬리질문 : if vs while 왜? Condition 1개만 쓰면? 공정성(fairness)은? lock-free 큐는?
 * </pre>
 */
public interface BoundedBlockingQueue<E> {

    /** 큐 끝에 넣는다. 가득 차 있으면 자리가 날 때까지 대기한다. */
    void put(E element) throws InterruptedException;

    /** 큐 앞에서 꺼낸다(FIFO). 비어 있으면 원소가 들어올 때까지 대기한다. */
    E take() throws InterruptedException;

    /** 현재 보관 중인 원소 수. */
    int size();
}
