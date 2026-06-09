package lld.lfucache;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/** LFU 캐시의 공용 계약 테스트. 정답지와 내 구현이 글자 그대로 같은 케이스로 검증된다. */
abstract class LfuCacheContractTest {

    protected abstract <K, V> LfuCache<K, V> newCache(int capacity);

    @Test
    @DisplayName("기본: put 한 값을 get으로 다시 읽는다")
    void put_then_get() {
        LfuCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        assertEquals("a", cache.get(1));
        assertEquals(1, cache.size());
    }

    @Test
    @DisplayName("불변식: 용량 초과 시 가장 적게 쓴 키가 사라진다")
    void evicts_least_frequently_used() {
        LfuCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        cache.put(2, "b");
        cache.get(1); // 1의 빈도=2, 2의 빈도=1
        cache.put(3, "c"); // 빈도 최소인 2가 제거
        assertNull(cache.get(2));
        assertEquals("a", cache.get(1));
        assertEquals("c", cache.get(3));
    }

    @Test
    @DisplayName("동률: 같은 빈도면 가장 오래된 키(LRU)가 먼저 제거된다")
    void ties_broken_by_least_recently_used() {
        LfuCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        cache.put(2, "b");
        cache.get(1); // 1 -> 빈도2 (버킷2에 1 먼저 들어감)
        cache.get(2); // 2 -> 빈도2 (버킷2에 2 나중에 들어감), 이제 둘 다 빈도2
        cache.put(3, "c"); // 빈도2 동률 -> 더 오래된 1이 제거
        assertNull(cache.get(1));
        assertEquals("b", cache.get(2));
        assertEquals("c", cache.get(3));
    }

    @Test
    @DisplayName("같은 키 put은 값 갱신 + 빈도 증가")
    void put_existing_updates_value_and_frequency() {
        LfuCache<Integer, String> cache = newCache(2);
        cache.put(1, "a");
        cache.put(2, "b");
        cache.put(1, "a2"); // 1의 빈도=2, 값 갱신
        cache.put(3, "c"); // 빈도 최소인 2 제거
        assertEquals("a2", cache.get(1));
        assertNull(cache.get(2));
        assertEquals("c", cache.get(3));
    }

    @Test
    @DisplayName("경계: 용량 1")
    void capacity_one() {
        LfuCache<Integer, String> cache = newCache(1);
        cache.put(1, "a");
        cache.put(2, "b"); // 1 제거
        assertNull(cache.get(1));
        assertEquals("b", cache.get(2));
        assertEquals(1, cache.size());
    }

    @Test
    @DisplayName("없는 키 조회는 null")
    void missing_key_returns_null() {
        LfuCache<Integer, String> cache = newCache(2);
        assertNull(cache.get(42));
    }

    @Test
    @DisplayName("실패: 용량이 0 이하면 생성 시 예외")
    void invalid_capacity_rejected() {
        assertThrows(IllegalArgumentException.class, () -> newCache(0));
        assertThrows(IllegalArgumentException.class, () -> newCache(-1));
    }
}
