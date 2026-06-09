package lld.lrucache;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/**
 * LRU 캐시의 '계약'을 담은 공용 테스트. 어떤 구현이든 이 테스트를 통과해야 LRU라고 부를 수 있다.
 *
 * <p>{@link #newCache(int)}만 구현체가 바꿔 끼우면, 정답지({@code ReferenceLruCacheTest})와
 * 내 구현({@code MyLruCacheTest})이 글자 그대로 같은 테스트로 검증된다.
 */
abstract class LruCacheContractTest {

    /** 각 구현체별 테스트가 자기 캐시를 만들어 준다. */
    protected abstract <K, V> LruCache<K, V> newCache(int capacity);

    @Test
    @DisplayName("기본: put 한 값을 get으로 다시 읽는다")
    void put_then_get() {
        LruCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        assertEquals("a", cache.get(1));
        assertEquals(1, cache.size());
    }

    @Test
    @DisplayName("불변식: 용량을 넘으면 가장 오래 안 쓴 키가 사라진다")
    void evicts_least_recently_used() {
        LruCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        cache.put(2, "b");
        cache.put(3, "c"); // 1이 LRU -> 제거
        assertNull(cache.get(1));
        assertEquals("b", cache.get(2));
        assertEquals("c", cache.get(3));
        assertEquals(2, cache.size());
    }

    @Test
    @DisplayName("불변식: get 하면 그 키가 최근으로 갱신되어 제거를 피한다")
    void get_refreshes_recency() {
        LruCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        cache.put(2, "b");
        assertEquals("a", cache.get(1)); // 1을 최근으로 -> 이제 2가 LRU
        cache.put(3, "c"); // 2 제거
        assertNull(cache.get(2));
        assertEquals("a", cache.get(1));
        assertEquals("c", cache.get(3));
    }

    @Test
    @DisplayName("같은 키 put은 값을 갱신하고 최근으로 만든다")
    void put_existing_updates_value_and_recency() {
        LruCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        cache.put(2, "b");
        cache.put(1, "a2"); // 값 갱신 + 1 최근화 -> 2가 LRU
        assertEquals("a2", cache.get(1));
        cache.put(3, "c"); // 2 제거
        assertNull(cache.get(2));
        assertEquals("a2", cache.get(1));
        assertEquals(2, cache.size());
    }

    @Test
    @DisplayName("경계: 용량 1은 새 put이 이전 항목을 즉시 밀어낸다")
    void capacity_one() {
        LruCache<Integer, String> cache = newCache(1);
        cache.put(1, "a");
        cache.put(2, "b");
        assertNull(cache.get(1));
        assertEquals("b", cache.get(2));
        assertEquals(1, cache.size());
    }

    @Test
    @DisplayName("없는 키 조회는 null")
    void missing_key_returns_null() {
        LruCache<Integer, String> cache = newCache(2);
        assertNull(cache.get(42));
    }

    @Test
    @DisplayName("실패: 용량이 0 이하면 생성 시 예외")
    void invalid_capacity_rejected() {
        assertThrows(IllegalArgumentException.class, () -> newCache(0));
        assertThrows(IllegalArgumentException.class, () -> newCache(-1));
    }
}
