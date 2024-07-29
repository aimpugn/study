package main.java.Collections;

import java.util.*;
import java.util.concurrent.*;

public class Main {

    public static void main(String[] args) {
        listExamples();
        setExamples();
        mapExamples();
        queueExamples();
        stackExamples();
        concurrentCollectionExamples();
        threadSafeCollectionExamples();
        specialPurposeCollectionExamples();
    }

    public static void listExamples() {
        /*
         * ArrayList: 동적 배열 기반의 리스트
         * [apple] -> [banana] -> [cherry] -> [date]
         *  |          |          |          |
         *  V          V          V          V
         * 메모리에 연속적으로 저장, 랜덤 액세스 빠름
         */
        List<String> arrayList =
                new ArrayList<>(List.of("apple", "banana", "cherry"));

        System.out.println("ArrayList: " + arrayList);

        arrayList.add("date"); // O(1) 평균, 최악의 경우 O(n)
        arrayList.remove("banana"); // O(n)
        boolean contains = arrayList.contains("cherry"); // O(n)
        int size = arrayList.size(); // O(1)
        String element = arrayList.get(1); // O(1)
        arrayList.sort(Comparator.naturalOrder()); // O(n log n)

        System.out.println("Updated ArrayList: " + arrayList);

        Collections.sort(arrayList, Collections.reverseOrder()); // O(n)
        System.out.println("Reverse order ArrayList: " + arrayList);

        /*
         * LinkedList: 이중 연결 리스트 기반의 리스트
         *
         * [dog] <-> [cat] <-> [bird]
         *  ^         ^         ^
         *  |         |         |
         *
         * 노드가 이전/다음 노드 참조, 삽입/삭제 빠름
         */
        List<String> linkedList =
                new LinkedList<>(List.of("dog", "cat", "bird"));
        // 불변 리스트 생성
        List<String> immutableList = List.of("red", "green", "blue");

        System.out.println("LinkedList: " + linkedList);
        System.out.println("Immutable List: " + immutableList);

    }

    public static void setExamples() {
        /*
         * HashSet: 해시 테이블 기반의 집합
         *
         * [1] -> [2] -> [3] -> [4] -> [5]
         *  |      |      |      |      |
         *  V      V      V      V      V
         * 버킷0  버킷1  버킷2  버킷3  버킷4
         *
         * - 해시 함수로 버킷 결정
         * - 빠른 삽입/삭제/검색(해시 함수 기반)
         * - 순서 보장 안되므로, 순서 보장이 필요하다면 LinkedHashSet 사용
         */
        Set<Integer> hashSet = new HashSet<>(Set.of(1, 2, 3, 4, 5));

        System.out.println("HashSet: " + hashSet);
        hashSet.add(6); // O(1) 평균
        hashSet.remove(3); // O(1) 평균
        boolean contains = hashSet.contains(4); // O(1) 평균
        int size = hashSet.size(); // O(1)

        System.out.println("Updated HashSet: " + hashSet);
        /*
         * TreeSet: 이진 검색 트리 기반의 정렬된 집합
         *      3
         *    /   \
         *   2     4
         *  /       \
         * 1         5
         * 이진 검색 트리로 구현, 정렬된 순서 유지
         */
        Set<Integer> treeSet = new TreeSet<>(Set.of(5, 4, 3, 2, 1));
        // 불변 집합 생성
        Set<Integer> immutableSet = Set.of(10, 20, 30);
        // LinkedHashSet: 삽입 순서를 보존하는 집합
        Set<Integer> linkedHashSet = new LinkedHashSet<>(Set.of(1, 2, 3, 4, 5));

        System.out.println("TreeSet: " + treeSet);
        System.out.println("Immutable Set: " + immutableSet);
        System.out.println("LinkedHashSet: " + linkedHashSet);
    }

    public static void mapExamples() {
        /*
         * HashMap: 해시 테이블 기반의 맵
         *
         * ["apple" -> 1] -> ["banana" -> 2] -> ["cherry" -> 3]
         *  |               |                  |
         *  V               V                  V
         * 버킷0           버킷1              버킷2
         *
         * - 해시 함수로 버킷 결정
         * - 빠른 삽입/삭제/검색(해시 함수 기반)
         * - 순서 보장 안되므로, 순서 보장이 필요한 경우 LinkedHashMap를 사용
         */
        Map<String, Integer> hashMap =
                new HashMap<>(Map.of("apple", 1, "banana", 2, "cherry", 3));

        System.out.println("HashMap: " + hashMap);

        hashMap.put("date", 4); // O(1) 평균
        hashMap.remove("banana"); // O(1) 평균
        boolean containsKey = hashMap.containsKey("cherry"); // O(1) 평균
        boolean containsValue = hashMap.containsValue(3); // O(n)
        int size = hashMap.size(); // O(1)
        int value = hashMap.get("apple"); // O(1) 평균

        System.out.println("Updated HashMap: " + hashMap);

        /*
         * TreeMap: 레드-블랙 트리 기반의 정렬된 맵
         *
         *          "cat" -> 2
         *        /            \
         * "bird" -> 3      "dog" -> 1
         *
         * 레드-블랙 트리로 구현, 키를 기준으로 정렬 유지
         */
        Map<String, Integer> treeMap =
                new TreeMap<>(Map.of("dog", 1, "cat", 2, "bird", 3));
        // 불변 맵 생성
        Map<String, Integer> immutableMap =
                Map.of("red", 1, "green", 2, "blue", 3);
        // LinkedHashMap: 삽입 순서 보존되는 맵
        Map<String, Integer> linkedHashMap = new LinkedHashMap<>(
                Map.of("apple", 1, "banana", 2, "cherry", 3));

        System.out.println("TreeMap: " + treeMap);
        System.out.println("Immutable Map: " + immutableMap);
        System.out.println("LinkedHashMap: " + linkedHashMap);
    }

    public static void queueExamples() {
        /*
         * Queue (LinkedList): LinkedList를 이용한 Queue
         *
         * [First] -> [Second] -> [Third] -> [Fourth]
         *  ^                                  |
         *  |                                  |
         *  +----------------------------------+
         *
         * FIFO (First-In-First-Out) 순서
         */
        Queue<String> queue =
                new LinkedList<>(List.of("First", "Second", "Third"));

        System.out.println("Queue: " + queue);

        queue.offer("Fourth"); // O(1)
        String first = queue.poll(); // O(1)
        String peek = queue.peek(); // O(1)
        int size = queue.size(); // O(1)
        boolean isEmpty = queue.isEmpty(); // O(1)

        System.out.println("Updated Queue: " + queue);

        /*
         * PriorityQueue: 힙(heap) 기반의 우선순위 큐
         *      1
         *    /   \
         *   1     4
         *  / \   / \
         * 3   5 9
         * 최소 힙 구조로 구현, 항상 가장 작은 요소가 루트에 위치.
         */
        Queue<Integer> priorityQueue =
                new PriorityQueue<>(List.of(3, 1, 4, 1, 5, 9));

        System.out.println("PriorityQueue: " + priorityQueue);
    }

    public static void stackExamples() {
        /*
         * Deque(ArrayDeque)를 이용한 Stack (Java에서는 Stack 클래스 대신 Deque 사용을 권장)
         *
         * [Top]
         * [Middle]
         * [Bottom]
         *
         * LIFO (Last-In-First-Out) 순서
         */
        Deque<String> stack =
                new ArrayDeque<>(List.of("Bottom", "Middle", "Top"));

        System.out.println("Stack: " + stack);

        stack.push("New Top"); // O(1)
        String top = stack.pop(); // O(1)
        String peek = stack.peek(); // O(1)
        int size = stack.size(); // O(1)
        boolean isEmpty = stack.isEmpty(); // O(1)

        System.out.println("Updated Stack: " + stack);
    }

    public static void dequeExamples() {
        // Deque를 스택과 큐로 모두 활용
        Deque<String> deque = new ArrayDeque<>(List.of("A", "B", "C"));

        deque.addFirst("D"); // 스택(FILO)처럼 사용 (push)
        deque.addLast("E"); // 큐(FIFO)처럼 사용 (enqueue)
        System.out.println("Deque after addFirst and addLast: " + deque);

        String first = deque.pollFirst(); // 큐처럼 사용 (dequeue)
        String last = deque.pollLast(); // 스택처럼 사용 (pop)
        System.out.println("First polled: " + first + ", Last polled: " + last);
    }

    public static void enumExamples() {
        // EnumSet 예시: enum 값을 위한 효율적인 집합
        EnumSet<Day> workDays =
                EnumSet.of(Day.MONDAY, Day.TUESDAY, Day.WEDNESDAY);
        System.out.println("EnumSet: " + workDays);

        // EnumMap 예시: enum을 키로 사용하는 맵
        EnumMap<Day, String> dayMap = new EnumMap<>(Day.class);
        dayMap.put(Day.MONDAY, "Work");
        dayMap.put(Day.SUNDAY, "Rest");
        System.out.println("EnumMap: " + dayMap);
    }

    enum Day {
        MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
    }

    public static void concurrentCollectionExamples() {
        /*
         * ConcurrentHashMap: 동시성을 지원하는 해시맵
         *
         * ["a" -> 1] -> ["b" -> 2] -> ["c" -> 3]
         *  |            |            |
         *  V            V            V
         * 세그먼트1    세그먼트2    세그먼트3
         *
         * 각 세그먼트별로 독립적인 락을 사용하여 동시성 지원
         */
        Map<String, Integer> concurrentMap =
                new ConcurrentHashMap<>(Map.of("a", 1, "b", 2, "c", 3));
        concurrentMap.put("d", 4); // 스레드 안전한 삽입
        concurrentMap.putIfAbsent("e", 5); // 키가 없을 때만 삽입
        concurrentMap.computeIfPresent("a", (k, v) -> v + 10); // 스레드 안전한 업데이트
        int value = concurrentMap.get("b"); // 스레드 안전한 조회

        System.out.println("Updated ConcurrentHashMap: " + concurrentMap);

        /*
         * ConcurrentSkipListMap/Set: 동시성을 지원하는 정렬된 맵/셋
         *
         * L4 -------> [9]
         * L3 -------> [5] -------> [9]
         * L2 -------> [2] -------> [5] -------> [9]
         * L1 -> [1] -> [2] -> [5] -> [8] -> [9]
         *
         * 스킵 리스트 구조로 구현, 로그 시간 복잡도의 검색/삽입/삭제
         */
        NavigableMap<String, Integer> skipListMap =
                new ConcurrentSkipListMap<>(Map.of("x", 10, "y", 20, "z", 30));
        NavigableSet<Integer> skipListSet =
                new ConcurrentSkipListSet<>(Set.of(5, 2, 8, 1, 9));

        System.out.println("ConcurrentHashMap: " + concurrentMap);
        System.out.println("ConcurrentSkipListMap: " + skipListMap);
        System.out.println("ConcurrentSkipListSet: " + skipListSet);
    }

    public static void threadSafeCollectionExamples() {
        // NOTE: Vector와 Hashtable은 구식이며 일반적으로 추천되지 않음.
        /*
        * Vector: 스레드 안전한 동적 배열 (ArrayList의 스레드 안전 버전)
        * [safe] -> [thread] -> [list] -> [example]
        *  |          |          |          |
        *  V          V          V          V
        * 동기화된 접근, 모든 메서드가 동기화됨
        */
        List<String> vector = new Vector<>(List.of("safe", "thread", "list"));

        vector.add("example"); // 동기화된 삽입
        String element = vector.get(0); // 동기화된 접근
        vector.remove(1); // 동기화된 삭제

        System.out.println("Updated Vector: " + vector);

        /*
         * Hashtable: 스레드 안전한 해시 테이블 (HashMap의 스레드 안전 버전, 하지만 ConcurrentHashMap이 더 선호됨)
         * ["x" -> 1] -> ["y" -> 2] -> ["z" -> 3]
         *  |            |            |
         *  V            V            V
         * 동기화된 버킷, 모든 메서드가 동기화됨
         */
        Map<String, Integer> hashtable =
                new Hashtable<>(Map.of("x", 1, "y", 2, "z", 3));

        System.out.println("Vector: " + vector);
        System.out.println("Hashtable: " + hashtable);

    }

    public static void specialPurposeCollectionExamples() {
        /*
         * ArrayDeque: 양방향 큐
         *
         * [new first] <-> [first] <-> [second] <-> [third] <-> [new last]
         *
         * 양 끝에서 효율적인 삽입/삭제 가능
         */
        Deque<String> arrayDeque =
                new ArrayDeque<>(List.of("first", "second", "third"));
        /*
         * PriorityBlockingQueue: 동시성을 지원하는 우선순위 큐
         *
         *      1
         *    /   \
         *   1     4
         *  / \   / \
         * 3   5 9
         *
         * 최소 힙 구조 + 스레드 안전성
         */
        Queue<Integer> priorityBlockingQueue =
                new PriorityBlockingQueue<>(List.of(3, 1, 4, 1, 5, 9));
        /*
         * DelayQueue: 지정된 시간이 지난 후에 요소를 꺼낼 수 있는 큐
         *
         * [Task 2 (1000ms)] -> [Task 1 (5000ms)]
         *
         * 시간 기반 우선순위, 지정 시간 후 요소 사용 가능
         */
        DelayQueue<DelayedElement> delayQueue = new DelayQueue<>();
        delayQueue.offer(new DelayedElement("Task 1", 5000));
        delayQueue.offer(new DelayedElement("Task 2", 1000));
        /*
         * LinkedBlockingQueue: 동시성을 지원하는 연결 리스트 기반 큐
         *
         * [blocked] -> [queue] -> [example]
         *  ^                        |
         *  |                        |
         *  +------------------------+
         *
         * FIFO 순서 + 스레드 안전성, 생산자-소비자 패턴에 적합
         */
        BlockingQueue<String> linkedBlockingQueue = new LinkedBlockingQueue<>(
                List.of("blocked", "queue", "example"));

        System.out.println("ArrayDeque: " + arrayDeque);
        System.out.println("PriorityBlockingQueue: " + priorityBlockingQueue);
        System.out.println("DelayQueue: " + delayQueue);
        System.out.println("LinkedBlockingQueue: " + linkedBlockingQueue);

        // 자주 사용되는 메서드 (ArrayDeque 예시)
        arrayDeque.addFirst("new first"); // O(1)
        arrayDeque.addLast("new last"); // O(1)
        String first = arrayDeque.pollFirst(); // O(1)
        String last = arrayDeque.pollLast(); // O(1)

        System.out.println("Updated ArrayDeque: " + arrayDeque);
    }

    // DelayQueue를 위한 지연 요소 클래스
    static class DelayedElement implements Delayed {
        private final String name;
        private final long endTime;

        public DelayedElement(String name, long delay) {
            this.name = name;
            this.endTime = System.currentTimeMillis() + delay;
        }

        @Override
        public long getDelay(TimeUnit unit) {
            long diff = endTime - System.currentTimeMillis();
            return unit.convert(diff, TimeUnit.MILLISECONDS);
        }

        @Override
        public int compareTo(Delayed o) {
            return Long.compare(this.endTime, ((DelayedElement) o).endTime);
        }

        @Override
        public String toString() {
            return name + " (delay until " + endTime + ")";
        }
    }
}
