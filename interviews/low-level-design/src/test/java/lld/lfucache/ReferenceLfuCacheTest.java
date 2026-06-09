package lld.lfucache;

/** 정답지를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceLfuCacheTest extends LfuCacheContractTest {

    @Override
    protected <K, V> LfuCache<K, V> newCache(int capacity) {
        return new ReferenceLfuCache<>(capacity);
    }
}
