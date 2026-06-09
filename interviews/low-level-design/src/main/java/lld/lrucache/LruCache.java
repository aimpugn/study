package lld.lrucache;

/**
 * LRU(Least Recently Used) 캐시 - 면접 암기 카드.
 *
 * <p>이 헤더만 외워도 면접장에서 LRU는 흔들리지 않는다. 정답지는 {@link ReferenceLruCache},
 * 직접 연습은 {@link MyLruCache}. 둘 다 {@code LruCacheContractTest}의 같은 테스트로 검증된다.
 *
 * <pre>
 * 한 줄    : get·put 모두 O(1) + 용량 초과 시 LRU 제거도 O(1). 이 둘을 동시에 만족시키는 게 전부다.
 * 자료구조 : HashMap&lt;K, Node&gt; + 직접 만든 '이중 연결 리스트'. head=가장 최근(MRU), tail=가장 오래됨(LRU).
 * get(k)  : 맵에서 노드를 O(1)로 찾고 -> 리스트에서 head로 이동(최근화). 없으면 null.
 * put(k,v): 있으면 값 갱신 + head 이동. 없으면 head 삽입, 용량 초과면 tail 앞 노드(LRU) 제거.
 * 왜 HashMap : 키로 노드를 O(1)에 찾아야 리스트에서 바로 이동 가능. 리스트만 쓰면 탐색이 O(n).
 * 왜 이중연결: 임의 노드를 O(1)에 떼려면 prev가 필요. 단일 연결은 앞 노드 찾기가 O(n).
 * 리소스   : 항목당 (맵 엔트리 + 포인터 2개 노드)의 메모리를 써서 '모든 연산 O(1)'을 산다.
 * 실무 대안: LinkedHashMap(accessOrder=true) + removeEldestEntry 오버라이드 / 고성능은 Caffeine.
 *           "면접은 직접 구현, 실무는 라이브러리"를 한 문장으로 말할 수 있어야 한다.
 * 동시성   : 기본은 thread-safe 아님. 필요하면 전체 lock(경합↑)·분할 lock, 실무는 Caffeine.
 * 꼬리질문: TTL 만료를 더하면? / 동시 접근이면? / LFU와 차이(빈도 vs 최근성)? / 용량 0이면?
 * </pre>
 */
public interface LruCache<K, V> {

    /** 없으면 {@code null}. 조회에 성공하면 그 키를 '가장 최근 사용'으로 만든다. */
    V get(K key);

    /** 있으면 값 갱신 후 최근화, 없으면 삽입. 용량을 넘으면 LRU 항목을 제거한다. */
    void put(K key, V value);

    /** 현재 보관 중인 항목 수. */
    int size();
}
