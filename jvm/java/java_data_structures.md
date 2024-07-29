# Java Data Structures

- [Java Data Structures](#java-data-structures)
    - [배열 (Array)](#배열-array)
    - [ArrayList](#arraylist)
    - [LinkedList](#linkedlist)
    - [HashMap](#hashmap)
    - [HashSet](#hashset)
    - [Stack](#stack)
    - [PriorityQueue](#priorityqueue)
    - [자료구조 예제 모음](#자료구조-예제-모음)

## 배열 (Array)

```java
int[] arr = new int[5];
for (int i = 0; i < arr.length; i++) {
    arr[i] = i * 2;
}
System.out.println(Arrays.toString(arr));  // [0, 2, 4, 6, 8]
```

```java
메모리 상의 배열 구조:
[0][1][2][3][4]...
 ^  ^  ^  ^  ^
 |  |  |  |  |
 요소들이 연속된 메모리 공간에 저장됨
```

- 고정된 크기: 한 번 생성되면 크기를 변경할 수 없습니다.
- 인덱스 기반 접근: `O(1)` 시간에 어떤 요소든 직접 접근 가능합니다.
- 캐시 효율성: 연속된 메모리 공간을 사용하므로 캐시 지역성이 좋습니다.

## ArrayList

```java
ArrayList<Integer> list = new ArrayList<>();
list.add(1);  // [1]
list.add(2);  // [1, 2]
list.add(1, 3);  // [1, 3, 2] - 인덱스 1에 3 삽입, 기존 요소들은 오른쪽으로 이동
```

```java
내부 구조:
[0][1][2][3][4][...][capacity-1]
 ^  ^  ^  ^  ^
 |  |  |  |  |
 실제 데이터가 저장된 영역
```

- 동적 크기 조절: 내부적으로 배열을 사용하지만, 크기가 부족하면 더 큰 배열을 생성하고 데이터를 복사합니다.
- 용량(Capacity)과 크기(Size): 용량은 내부 배열의 실제 크기, 크기는 저장된 요소의 수입니다.
- 삽입/삭제 시간 복잡도:
    - 끝에서: O(1) (평균)
    - 중간에서: O(n) (요소 이동 필요)

## LinkedList

```java
LinkedList<String> list = new LinkedList<>();
list.addFirst("First");
list.addLast("Last");
list.add(1, "Middle");
System.out.println(list);  // [First, Middle, Last]
```

```java
단일 연결 리스트:
[데이터|다음노드] -> [데이터|다음노드] -> [데이터|다음노드] -> null

이중 연결 리스트 (Java의 LinkedList):
null <- [이전|데이터|다음] <-> [이전|데이터|다음] <-> [이전|데이터|다음] -> null
```

- 노드 기반 구조: 각 노드는 데이터와 다음/이전 노드에 대한 참조를 포함합니다.
- 삽입/삭제 효율성: 노드의 참조만 변경하면 되므로 O(1) 시간에 가능합니다.
- 메모리 사용: 각 노드마다 추가 메모리(참조)가 필요합니다.

## HashMap

```java
HashMap<String, Integer> map = new HashMap<>();
map.put("One", 1);  // "One"의 해시값에 따라 특정 버킷에 저장
map.put("Two", 2);
System.out.println(map.get("One"));  // 1
```

```java
내부 구조:
[0] -> (key1, value1) -> (key2, value2)
[1] -> (key3, value3)
[2]
[3] -> (key4, value4)
...
```

- 해시 함수: 키를 내부 배열의 인덱스로 매핑합니다.
- 충돌 해결: 체이닝 방식을 사용 (동일한 인덱스에 여러 항목을 연결 리스트로 저장)
- 로드 팩터: 기본값 0.75, 이 값을 초과하면 내부 배열 크기를 증가시킵니다.

## HashSet

```java
HashSet<String> set = new HashSet<>();
set.add("Apple");
set.add("Banana");
set.add("Apple");  // 중복, 무시됨
System.out.println(set);  // [Apple, Banana]
```

```java
내부적으로 HashMap을 사용:
[0] -> (element1, PRESENT)
[1] -> (element2, PRESENT)
[2]
[3] -> (element3, PRESENT)
...
```

- HashMap 기반: 내부적으로 HashMap을 사용하며, 모든 값은 동일한 더미 객체(PRESENT)입니다.
- 중복 제거: 동일한 해시코드와 equals() 결과를 가진 요소는 하나만 저장됩니다.

## Stack

```java
Stack<Integer> stack = new Stack<>();
stack.push(1);
stack.push(2);
stack.push(3);
System.out.println(stack.pop());  // 3
System.out.println(stack.peek()); // 2
```

```java
│   │
│ 3 │ <- top
│ 2 │
│ 1 │
└───┘
```

- LIFO (Last In First Out): 가장 최근에 추가된 항목이 가장 먼저 제거됩니다.
- Vector 상속: 스레드 안전하지만, 성능 오버헤드가 있습니다.

## PriorityQueue

```java
// 최대 힙
PriorityQueue<Integer> maxHeap = new PriorityQueue<>(Collections.reverseOrder());
// PriorityQueue<Integer> maxHeap = new PriorityQueue<>((a, b) -> b - a);
maxHeap.offer(3);
maxHeap.offer(1);
maxHeap.offer(4);
System.out.println(maxHeap.poll());  // 4
```

```java
최소 힙 구조:
         1
       /   \
      2     3
     / \   / \
    4   5 6   7
   / \
  8   9
       ┌─────────┐
    ┌──│───┐     │
    │  │  ┌┴─┐  ┌┴─┐
[1, 2, 3, 4, 5, 6, 7, 8, 9]
 │  └┬─┘  │           └┬─┘
 └───┘    └────────────┘

조금 더 복잡한 최소 힙 구조:
       1
     /   \
    2     3
   / \   / \
  4   5 6   7
 / \  |
8  10 9

       ┌─────────┐
    ┌──│───┐     │
    │  │  ┌┴─┐  ┌┴─┐
[1, 2, 3, 4, 5, 6, 7, 8, 10, 9] // 8, 10, 9 사이 순서는 상관 없습니다.
 │  └┬─┘  │  │        └┬─┘   │  // 부모 노드가 자식 노드보다 작기만 하면 됩니다.
 └───┘    └──│─────────┘     │
             └───────────────┘
```

- 힙 기반 구현: 완전 이진 트리 형태의 힙을 사용합니다.
- 최소 힙 (기본): 부모 노드의 값이 항상 자식 노드의 값보다 작거나 같습니다.
- 삽입/삭제 시간 복잡도: O(log n)

## 자료구조 예제 모음

```java
import java.util.*;
import java.util.concurrent.*;

public class ComprehensiveDataStructuresJava14Plus {

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
        List<String> arrayList = new ArrayList<>(List.of("apple", "banana", "cherry"));

        System.out.println("ArrayList: " + arrayList);

        arrayList.add("date");              // O(1) 평균, 최악의 경우 O(n)
        arrayList.remove("banana");         // O(n)
        boolean contains = arrayList.contains("cherry");  // O(n)
        int size = arrayList.size();        // O(1)
        String element = arrayList.get(1);  // O(1)
        arrayList.sort(Comparator.naturalOrder());  // O(n log n)

        System.out.println("Updated ArrayList: " + arrayList);

        /*
         * LinkedList: 이중 연결 리스트 기반의 리스트
         *
         * [dog] <-> [cat] <-> [bird]
         *  ^         ^         ^
         *  |         |         |
         *
         * 노드가 이전/다음 노드 참조, 삽입/삭제 빠름
         */
        List<String> linkedList = new LinkedList<>(List.of("dog", "cat", "bird"));
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
        hashSet.add(6);                 // O(1) 평균
        hashSet.remove(3);              // O(1) 평균
        boolean contains = hashSet.contains(4);  // O(1) 평균
        int size = hashSet.size();      // O(1)

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
        Map<String, Integer> hashMap = new HashMap<>(Map.of("apple", 1, "banana", 2, "cherry", 3));

        System.out.println("HashMap: " + hashMap);

        hashMap.put("date", 4);         // O(1) 평균
        hashMap.remove("banana");       // O(1) 평균
        boolean containsKey = hashMap.containsKey("cherry");  // O(1) 평균
        boolean containsValue = hashMap.containsValue(3);     // O(n)
        int size = hashMap.size();      // O(1)
        int value = hashMap.get("apple");  // O(1) 평균

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
        Map<String, Integer> treeMap = new TreeMap<>(Map.of("dog", 1, "cat", 2, "bird", 3));
        // 불변 맵 생성
        Map<String, Integer> immutableMap = Map.of("red", 1, "green", 2, "blue", 3);
        // LinkedHashMap: 삽입 순서 보존되는 맵
        Map<String, Integer> linkedHashMap = new LinkedHashMap<>(Map.of("apple", 1, "banana", 2, "cherry", 3));

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
        Queue<String> queue = new LinkedList<>(List.of("First", "Second", "Third"));

        System.out.println("Queue: " + queue);

        queue.offer("Fourth");          // O(1)
        String first = queue.poll();    // O(1)
        String peek = queue.peek();     // O(1)
        int size = queue.size();        // O(1)
        boolean isEmpty = queue.isEmpty();  // O(1)

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
        Queue<Integer> priorityQueue = new PriorityQueue<>(List.of(3, 1, 4, 1, 5, 9));

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
        Deque<String> stack = new ArrayDeque<>(List.of("Bottom", "Middle", "Top"));

        System.out.println("Stack: " + stack);

        stack.push("New Top");          // O(1)
        String top = stack.pop();       // O(1)
        String peek = stack.peek();     // O(1)
        int size = stack.size();        // O(1)
        boolean isEmpty = stack.isEmpty();  // O(1)

        System.out.println("Updated Stack: " + stack);
    }

    public static void dequeExamples() {
        // Deque를 스택과 큐로 모두 활용
        Deque<String> deque = new ArrayDeque<>(List.of("A", "B", "C"));

        deque.addFirst("D");  // 스택(FILO)처럼 사용 (push)
        deque.addLast("E");   // 큐(FIFO)처럼 사용 (enqueue)
        System.out.println("Deque after addFirst and addLast: " + deque);

        String first = deque.pollFirst();  // 큐처럼 사용 (dequeue)
        String last = deque.pollLast();    // 스택처럼 사용 (pop)
        System.out.println("First polled: " + first + ", Last polled: " + last);
    }

    public static void enumExamples() {
        // EnumSet 예시: enum 값을 위한 효율적인 집합
        EnumSet<Day> workDays = EnumSet.of(Day.MONDAY, Day.TUESDAY, Day.WEDNESDAY);
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
        Map<String, Integer> concurrentMap = new ConcurrentHashMap<>(Map.of("a", 1, "b", 2, "c", 3));
        concurrentMap.put("d", 4);                // 스레드 안전한 삽입
        concurrentMap.putIfAbsent("e", 5);        // 키가 없을 때만 삽입
        concurrentMap.computeIfPresent("a", (k, v) -> v + 10);  // 스레드 안전한 업데이트
        int value = concurrentMap.get("b");       // 스레드 안전한 조회

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
        NavigableMap<String, Integer> skipListMap = new ConcurrentSkipListMap<>(Map.of("x", 10, "y", 20, "z", 30));
        NavigableSet<Integer> skipListSet = new ConcurrentSkipListSet<>(Set.of(5, 2, 8, 1, 9));

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

        vector.add("example");          // 동기화된 삽입
        String element = vector.get(0); // 동기화된 접근
        vector.remove(1);               // 동기화된 삭제

        System.out.println("Updated Vector: " + vector);

        /*
         * Hashtable: 스레드 안전한 해시 테이블 (HashMap의 스레드 안전 버전, 하지만 ConcurrentHashMap이 더 선호됨)
         * ["x" -> 1] -> ["y" -> 2] -> ["z" -> 3]
         *  |            |            |
         *  V            V            V
         * 동기화된 버킷, 모든 메서드가 동기화됨
         */
        Map<String, Integer> hashtable = new Hashtable<>(Map.of("x", 1, "y", 2, "z", 3));

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
        Deque<String> arrayDeque = new ArrayDeque<>(List.of("first", "second", "third"));
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
        Queue<Integer> priorityBlockingQueue = new PriorityBlockingQueue<>(List.of(3, 1, 4, 1, 5, 9));
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
        BlockingQueue<String> linkedBlockingQueue = new LinkedBlockingQueue<>(List.of("blocked", "queue", "example"));

        System.out.println("ArrayDeque: " + arrayDeque);
        System.out.println("PriorityBlockingQueue: " + priorityBlockingQueue);
        System.out.println("DelayQueue: " + delayQueue);
        System.out.println("LinkedBlockingQueue: " + linkedBlockingQueue);

        // 자주 사용되는 메서드 (ArrayDeque 예시)
        arrayDeque.addFirst("new first");   // O(1)
        arrayDeque.addLast("new last");     // O(1)
        String first = arrayDeque.pollFirst(); // O(1)
        String last = arrayDeque.pollLast();   // O(1)

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
```
