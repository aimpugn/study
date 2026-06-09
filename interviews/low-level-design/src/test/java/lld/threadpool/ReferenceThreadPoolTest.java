package lld.threadpool;

/** 정답지를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceThreadPoolTest extends ThreadPoolContractTest {

    @Override
    protected ThreadPool newPool(int poolSize) {
        return new ReferenceThreadPool(poolSize);
    }
}
