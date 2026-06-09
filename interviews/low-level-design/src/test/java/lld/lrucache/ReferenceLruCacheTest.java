package lld.lrucache;

/** 정답지를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceLruCacheTest extends LruCacheContractTest {

    @Override
    protected <K, V> LruCache<K, V> newCache(int capacity) {
        return new ReferenceLruCache<>(capacity);
    }
}
