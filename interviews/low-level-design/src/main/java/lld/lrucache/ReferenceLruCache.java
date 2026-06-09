package lld.lrucache;

import java.util.HashMap;
import java.util.Map;

/**
 * LRU 캐시 정답지. 평소엔 {@link MyLruCache}에서 직접 짜고, 막힐 때만 열어 본다.
 *
 * <p>핵심 트릭 두 가지:
 * (1) 키-&gt;노드 맵으로 임의 노드를 O(1)에 찾는다.
 * (2) dummy head/tail 센티넬을 둬서 삽입·삭제에서 null 분기를 없앤다.
 */
public final class ReferenceLruCache<K, V> implements LruCache<K, V> {

    private static final class Node<K, V> {
        K key;
        V value;
        Node<K, V> prev;
        Node<K, V> next;

        Node() {
        }

        Node(K key, V value) {
            this.key = key;
            this.value = value;
        }
    }

    private final int capacity;
    private final Map<K, Node<K, V>> index = new HashMap<>();

    // 센티넬: 실제 데이터가 아니라 경계 표식. head.next가 MRU, tail.prev가 LRU다.
    // 덕분에 모든 실제 노드는 prev/next가 항상 non-null이라 unlink/link에 null 검사가 없다.
    private final Node<K, V> head = new Node<>();
    private final Node<K, V> tail = new Node<>();

    public ReferenceLruCache(int capacity) {
        if (capacity <= 0) {
            throw new IllegalArgumentException("capacity must be > 0 but was " + capacity);
        }
        this.capacity = capacity;
        head.next = tail;
        tail.prev = head;
    }

    @Override
    public V get(K key) {
        Node<K, V> node = index.get(key);
        if (node == null) {
            return null;
        }
        moveToFront(node); // 접근했으니 가장 최근으로 끌어올린다.
        return node.value;
    }

    @Override
    public void put(K key, V value) {
        Node<K, V> existing = index.get(key);
        if (existing != null) {
            existing.value = value;
            moveToFront(existing);
            return;
        }
        if (index.size() == capacity) {
            evictLeastRecentlyUsed();
        }
        Node<K, V> node = new Node<>(key, value);
        index.put(key, node);
        linkToFront(node);
    }

    @Override
    public int size() {
        return index.size();
    }

    private void evictLeastRecentlyUsed() {
        Node<K, V> lru = tail.prev; // tail 바로 앞이 가장 오래된 실제 노드.
        unlink(lru);
        index.remove(lru.key);
    }

    private void moveToFront(Node<K, V> node) {
        unlink(node);
        linkToFront(node);
    }

    private void unlink(Node<K, V> node) {
        // 센티넬 덕분에 node.prev, node.next는 항상 존재한다.
        node.prev.next = node.next;
        node.next.prev = node.prev;
    }

    private void linkToFront(Node<K, V> node) {
        node.prev = head;
        node.next = head.next;
        head.next.prev = node;
        head.next = node;
    }
}
