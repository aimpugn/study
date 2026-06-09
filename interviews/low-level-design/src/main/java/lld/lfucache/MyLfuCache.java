package lld.lfucache;

/**
 * 직접 구현하는 공간. {@link LfuCache} 헤더 카드와 테스트만 보고 채운다.
 * 막히면 그때만 {@link ReferenceLfuCache}를 연다.
 *
 * <p>시작: TODO를 채우고 {@code MyLfuCacheTest}의 {@code @Disabled}를 지운 뒤 테스트를 통과시킨다.
 */
public final class MyLfuCache<K, V> implements LfuCache<K, V> {

    public MyLfuCache(int capacity) {
        // TODO: capacity 검증 + (values / counts / freqToKeys / minFreq) 초기화
    }

    @Override
    public V get(K key) {
        // TODO: 없으면 null, 있으면 빈도 +1 후 값 반환
        throw new UnsupportedOperationException("아직 구현 전: MyLfuCache.get");
    }

    @Override
    public void put(K key, V value) {
        // TODO: 있으면 갱신+빈도증가, 없으면 빈도1 삽입(용량 초과 시 minFreq 버킷에서 LFU 제거)
        throw new UnsupportedOperationException("아직 구현 전: MyLfuCache.put");
    }

    @Override
    public int size() {
        // TODO
        throw new UnsupportedOperationException("아직 구현 전: MyLfuCache.size");
    }
}
