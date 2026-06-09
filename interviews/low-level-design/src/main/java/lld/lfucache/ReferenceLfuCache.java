package lld.lfucache;

import java.util.HashMap;
import java.util.LinkedHashSet;
import java.util.Map;

/**
 * LFU 캐시 정답지. 모든 연산 O(1)인 표준 구현.
 *
 * <p>핵심: 빈도별 버킷({@code freqToKeys})을 두고 {@code minFreq}만 추적하면, 가장 낮은 빈도의
 * 가장 오래된 항목을 O(1)에 제거할 수 있다. 같은 빈도 안 순서는 {@link LinkedHashSet}이 삽입순으로 지켜 준다.
 */
public final class ReferenceLfuCache<K, V> implements LfuCache<K, V> {

    private final int capacity;
    private int minFreq;
    private final Map<K, V> values = new HashMap<>();
    private final Map<K, Integer> counts = new HashMap<>();
    private final Map<Integer, LinkedHashSet<K>> freqToKeys = new HashMap<>();

    public ReferenceLfuCache(int capacity) {
        if (capacity <= 0) {
            throw new IllegalArgumentException("capacity must be > 0 but was " + capacity);
        }
        this.capacity = capacity;
    }

    @Override
    public V get(K key) {
        if (!values.containsKey(key)) {
            return null;
        }
        increaseFrequency(key);
        return values.get(key);
    }

    @Override
    public void put(K key, V value) {
        if (values.containsKey(key)) {
            values.put(key, value);
            increaseFrequency(key);
            return;
        }
        if (values.size() == capacity) {
            evictLeastFrequentlyUsed();
        }
        values.put(key, value);
        counts.put(key, 1);
        freqToKeys.computeIfAbsent(1, f -> new LinkedHashSet<>()).add(key);
        minFreq = 1; // 방금 넣은 키의 빈도가 1이므로 최소 빈도는 1.
    }

    @Override
    public int size() {
        return values.size();
    }

    private void increaseFrequency(K key) {
        int freq = counts.get(key);
        counts.put(key, freq + 1);

        LinkedHashSet<K> oldBucket = freqToKeys.get(freq);
        oldBucket.remove(key);
        if (oldBucket.isEmpty()) {
            freqToKeys.remove(freq);
            if (minFreq == freq) {
                minFreq = freq + 1; // 최소 빈도 버킷이 비었으니 한 칸 올린다.
            }
        }
        freqToKeys.computeIfAbsent(freq + 1, f -> new LinkedHashSet<>()).add(key);
    }

    private void evictLeastFrequentlyUsed() {
        LinkedHashSet<K> bucket = freqToKeys.get(minFreq);
        K victim = bucket.iterator().next(); // 가장 오래된(삽입 순서 첫 번째) = 동률 중 LRU.
        bucket.remove(victim);
        if (bucket.isEmpty()) {
            freqToKeys.remove(minFreq);
        }
        values.remove(victim);
        counts.remove(victim);
    }
}
