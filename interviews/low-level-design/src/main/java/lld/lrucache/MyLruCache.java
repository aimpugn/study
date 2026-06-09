package lld.lrucache;

/**
 * 직접 구현하는 공간. 목표: {@link LruCache} 헤더 카드와 테스트만 보고 채우기.
 * 막히면 그때만 {@link ReferenceLruCache}를 연다.
 *
 * <p>시작 방법: 아래 TODO를 채우고, {@code MyLruCacheTest}의 {@code @Disabled}를 지운 뒤
 * {@code gradle test}로 정답지와 똑같은 테스트를 빨강-&gt;초록으로 통과시킨다.
 */
public final class MyLruCache<K, V> implements LruCache<K, V> {

    public MyLruCache(int capacity) {
        // TODO: capacity 검증 + 자료구조 초기화 (맵 + 이중 연결 리스트 센티넬)
    }

    @Override
    public V get(K key) {
        // TODO: 맵 조회 -> 있으면 최근화 후 값 반환, 없으면 null
        throw new UnsupportedOperationException("아직 구현 전: MyLruCache.get");
    }

    @Override
    public void put(K key, V value) {
        // TODO: 있으면 갱신+최근화, 없으면 삽입(용량 초과 시 LRU 제거)
        throw new UnsupportedOperationException("아직 구현 전: MyLruCache.put");
    }

    @Override
    public int size() {
        // TODO: 현재 항목 수
        throw new UnsupportedOperationException("아직 구현 전: MyLruCache.size");
    }
}
