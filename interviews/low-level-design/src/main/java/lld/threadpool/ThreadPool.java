package lld.threadpool;

import java.time.Duration;

/**
 * 스레드 풀 - 면접 암기 카드. 정답지는 고정 크기 풀({@link ReferenceThreadPool}).
 *
 * <p>직접 연습은 {@link MyThreadPool}. 둘 다 {@code ThreadPoolContractTest}의 같은 테스트로 검증된다.
 * 생산자-소비자 큐({@code BoundedBlockingQueue})와 한 쌍으로 외운다: 풀 = 큐 + 워커.
 *
 * <pre>
 * 한 줄    : 스레드를 매번 만들지 않고 N개 워커가 '작업 큐'를 소비. 제출=큐에 넣기, 워커=꺼내 실행.
 * 구조     : BlockingQueue&lt;Runnable&gt; + 고정 워커 스레드 N개. 워커 루프 = while(돌아감) queue.take().run().
 * 왜 풀    : 스레드 생성·소멸 비용 제거 + 동시 실행 수 상한(과부하·OOM 방지).
 * 거부 정책(전략) : 큐가 가득 차면 - 호출자 블록 / 예외 / 호출자가 직접 실행 / 버림. (java RejectedExecutionHandler)
 * 종료     : shutdown=새 작업 거부+남은 작업 처리, shutdownNow=즉시 중단+대기 작업 반환. 여기선 poison pill로 graceful.
 * 크기     : CPU 바운드는 코어 수 근처, IO 바운드는 더 크게(대기 동안 다른 작업 실행).
 * 함정     : 작업이 던진 예외가 워커 스레드를 죽이지 않게 try/catch로 격리한다.
 * 실무     : ThreadPoolExecutor(corePoolSize·workQueue·handler=전략·ThreadFactory=팩토리). 직접 구현은 개념 증명용.
 * 꼬리질문 : 큐는 유한/무한? 거부 정책은? graceful vs 즉시 종료? 작업 예외 격리는?
 * </pre>
 */
public interface ThreadPool {

    /** 작업을 큐에 넣어 워커가 실행하게 한다. 이미 종료됐으면 거부한다. */
    void execute(Runnable task);

    /** 새 작업 수용을 멈추고, 큐에 남은 작업을 모두 처리한 뒤 워커를 종료한다(graceful). */
    void shutdown();

    /** 모든 워커가 끝날 때까지 최대 {@code timeout}만큼 기다린다. 끝났으면 {@code true}. */
    boolean awaitTermination(Duration timeout) throws InterruptedException;
}
