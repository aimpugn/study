package lld.lrucache;

import org.junit.jupiter.api.Disabled;

/**
 * 내 구현을 정답지와 똑같은 계약 테스트에 연결한다.
 * 구현을 시작하면 아래 {@code @Disabled} 한 줄을 지우고 {@code gradle test}로 빨강-&gt;초록을 만든다.
 */
@Disabled("내 구현(MyLruCache)을 시작하면 이 줄을 지우세요.")
class MyLruCacheTest extends LruCacheContractTest {

    @Override
    protected <K, V> LruCache<K, V> newCache(int capacity) {
        return new MyLruCache<>(capacity);
    }
}
