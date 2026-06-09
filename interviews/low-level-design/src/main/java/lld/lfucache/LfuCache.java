package lld.lfucache;

/**
 * LFU(Least Frequently Used) 캐시 - 면접 암기 카드.
 *
 * <p>정답지는 {@link ReferenceLfuCache}, 직접 연습은 {@link MyLfuCache}.
 * 둘 다 {@code LfuCacheContractTest}의 같은 테스트로 검증된다. LRU와 짝으로 외운다.
 *
 * <pre>
 * 한 줄    : get·put O(1), 가득 차면 '사용 빈도가 가장 낮은' 항목 제거(동률이면 그중 가장 오래된 것=LRU).
 * 자료구조 : Map&lt;K,V&gt; values + Map&lt;K,Integer&gt; counts + Map&lt;Integer, LinkedHashSet&lt;K&gt;&gt; freqToKeys + minFreq.
 * get/put  : 접근하면 freq+1 -> 그 키를 freq 버킷에서 freq+1 버킷으로 옮긴다. 비워진 버킷이 minFreq였으면 minFreq++.
 * 제거     : minFreq 버킷의 가장 오래된 키(LinkedHashSet의 첫 원소)를 뺀다.
 * 왜 LinkedHashSet : 같은 빈도 안에서 삽입순=LRU순 유지 + add·remove·first 모두 O(1).
 * LRU와 차이 : LRU=최근성만. LFU=빈도 우선, 동률에서만 최근성을 본다.
 * 약점     : 한때 인기였다 식은 항목이 높은 freq로 안 빠지는 'cache pollution'. -> 빈도에 시간 감쇠(aging)가 필요.
 * 실무     : Caffeine의 W-TinyLFU(작은 윈도 LRU + 빈도 스케치)가 LRU·LFU 약점을 절충.
 * 꼬리질문 : 동률은 어떻게? minFreq 갱신 시점은? 빈도 오버플로·aging은? 동시성은?
 * </pre>
 */
public interface LfuCache<K, V> {

    /** 없으면 {@code null}. 조회에 성공하면 그 키의 사용 빈도를 1 올린다. */
    V get(K key);

    /** 있으면 값 갱신 후 빈도 1 증가, 없으면 빈도 1로 삽입. 용량 초과 시 LFU 항목 제거. */
    void put(K key, V value);

    /** 현재 보관 중인 항목 수. */
    int size();
}
