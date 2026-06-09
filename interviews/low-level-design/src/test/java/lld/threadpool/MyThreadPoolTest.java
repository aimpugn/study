package lld.threadpool;

import org.junit.jupiter.api.Disabled;

/** 내 구현을 정답지와 같은 계약 테스트에 연결한다. 구현을 시작하면 아래 한 줄을 지운다. */
@Disabled("내 구현(MyThreadPool)을 시작하면 이 줄을 지우세요.")
class MyThreadPoolTest extends ThreadPoolContractTest {

    @Override
    protected ThreadPool newPool(int poolSize) {
        return new MyThreadPool(poolSize);
    }
}
