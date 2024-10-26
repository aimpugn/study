import java.util.*
import kotlin.collections.*


fun main() {
    setExamples()
    mapExamples()
    queueExamples()
    sequenceExamples()
    heapExample()
    sortingExamples()
}


fun setExamples() {
    println("\n--- Set Examples ---")
    val immutableSet = setOf(1, 2, 3, 4, 5)
    println("Immutable Set: $immutableSet")

    val mutableSet = mutableSetOf(5, 4, 3, 2, 1)
    println("Original Mutable Set: $mutableSet")

    mutableSet.add(6)
    mutableSet.remove(3)
    val contains = 4 in mutableSet
    val size = mutableSet.size

    println("Updated Mutable Set: $mutableSet")

    val linkedHashSet = LinkedHashSet<Int>()
    linkedHashSet.addAll(listOf(1, 2, 3, 4, 5))
    println("LinkedHashSet: $linkedHashSet")

    val treeSet = sortedSetOf(5, 2, 8, 1, 9)
    println("TreeSet: $treeSet")
}

fun mapExamples() {
    println("\n--- Map Examples ---")
    val immutableMap = mapOf("a" to 1, "b" to 2, "c" to 3)
    println("Immutable Map: $immutableMap")

    val mutableMap = mutableMapOf("x" to 10, "y" to 20, "z" to 30)
    println("Original Mutable Map: $mutableMap")

    mutableMap["w"] = 40
    mutableMap.remove("y")
    val containsKey = "z" in mutableMap
    val containsValue = 30 in mutableMap.values
    val size = mutableMap.size
    val value = mutableMap["x"]

    println("Updated Mutable Map: $mutableMap")

    val linkedHashMap = linkedMapOf("one" to 1, "two" to 2, "three" to 3)
    println("LinkedHashMap: $linkedHashMap")

    val treeMap = sortedMapOf("banana" to 2, "apple" to 1, "cherry" to 3)
    println("TreeMap: $treeMap")

    val map = mapOf("a" to 1, "b" to 2, "c" to 3)
    val keyList = map.keys.toList()
    val valueList = map.values.toList()
    println("Keys: $keyList, Values: $valueList")
}

fun queueExamples() {
    println("\n--- Queue and Deque Examples ---")

    // ArrayDeque를 사용한 큐 구현
    val deque = ArrayDeque(listOf("First", "Second", "Third"))
    println("Original Deque: $deque")

    // 요소 추가 (큐의 끝에 추가)
    deque.addLast("Fourth")
    println("After addLast: $deque")

    // 요소 제거 및 반환 (큐의 앞에서 제거)
    val removed = deque.removeFirst()
    println("Removed first element: $removed")
    println("After removeFirst: $deque")

    // 첫 번째 요소 확인 (제거하지 않음)
    val first = deque.first()
    println("First element (without removal): $first")
    println("Deque after peeking first: $deque")

    // LinkedList as Queue
    val queue: Queue<String> = LinkedList()

    // Queue 인터페이스의 메서드 사용
    queue.offer("Fifth")  // 요소 추가
    println("After offer: $queue")

    val polled = queue.poll()  // 요소 제거 및 반환
    println("Polled: $polled")
    println("After poll: $queue")

    val peeked = queue.peek()  // 첫 번째 요소 확인 (제거하지 않음)
    println("Peeked: $peeked")
    println("After peek: $queue")

    // 스택으로서의 사용 예 (LIFO)
    println("\nUsing ArrayDeque as a stack:")
    val stack = ArrayDeque<String>()
    stack.addFirst("Bottom")
    stack.addFirst("Middle")
    stack.addFirst("Top")
    println("Stack: $stack")

    val popped = stack.removeFirst()
    println("Popped: $popped")
    println("Stack after pop: $stack")
}

fun sequenceExamples() {
    println("\n--- Special Collections Examples ---")
    val sequence = sequenceOf(1, 2, 3, 4, 5)
        .map { it * 2 }
        .filter { it % 3 == 0 }
    println("Sequence result: ${sequence.toList()}")
}

fun heapExample() {
    println("\n--- Heap Example ---")

    // 최소
    val minHeap = PriorityQueue<Int>()
    minHeap.addAll(listOf(5, 3, 7, 1))
    println("Min Heap content: $minHeap") // [1, 3, 7, 5]
    while (minHeap.isNotEmpty()) {
        println("Polled: ${minHeap.poll()}")
    }
    // 출력:
    //  Polled: 1
    //  Polled: 3
    //  Polled: 5
    //  Polled: 7

    // 최대 힙
    val maxHeap = PriorityQueue<Int>(compareByDescending { it })
    maxHeap.addAll(listOf(5, 3, 7, 1))
    println("Max Heap content: $maxHeap") // [7, 3, 5, 1]
    while (maxHeap.isNotEmpty()) {
        println("Polled: ${maxHeap.poll()}")
    }
    // 출력:
    //  Polled: 7
    //  Polled: 5
    //  Polled: 3
    //  Polled: 1

    data class Person(val name: String, val age: Int)

    // 임의 타입의 최소 힙
    val personMinHeap = PriorityQueue<Person>(compareBy { it.age })
    personMinHeap.addAll(
        listOf(
            Person("Alice", 25),
            Person("Bob", 30),
            Person("Charlie", 20)
        )
    )
    println("Person Min Heap content: $personMinHeap")
    while (personMinHeap.isNotEmpty()) {
        println("Polled: ${personMinHeap.poll()}")
    }
    // 출력:
    //  Polled: Person(name=Charlie, age=20)
    //  Polled: Person(name=Alice, age=25)
    //  Polled: Person(name=Bob, age=30)

    // 임의 타입의 최소 힙
    val personMaxHeap = PriorityQueue<Person>(compareByDescending { it.age })
    personMaxHeap.addAll(
        listOf(
            Person("Alice", 25),
            Person("Bob", 30),
            Person("Charlie", 20)
        )
    )
    println("Person Max Heap content: $personMaxHeap")
    while (personMaxHeap.isNotEmpty()) {
        println("Polled: ${personMaxHeap.poll()}")
    }
    // 출력:
    //  Polled: Person(name=Bob, age=30)
    //  Polled: Person(name=Alice, age=25)
    //  Polled: Person(name=Charlie, age=20)
}


fun sortingExamples() {
    println("\n--- Sorting Examples ---")
    val numbers = mutableListOf(5, 2, 8, 1, 9, 3)

    // 오름차순 정렬
    numbers.sort()
    println("Ascending order: $numbers")

    // 내림차순 정렬
    numbers.sortDescending()
    println("Descending order: $numbers")

    // 커스텀 비교자를 사용한 정렬
    numbers.sortWith(compareBy { it % 3 })
    println("Sorted by remainder when divided by 3: $numbers")

    val fruits = listOf("apple", "banana", "cherry", "date")

    // 정렬된 새 리스트 생성 (원본 변경 없음)
    val sortedFruits = fruits.sorted()
    println("Sorted fruits: $sortedFruits")

    // 역순으로 정렬된 새 리스트 생성
    val reverseSortedFruits = fruits.sortedDescending()
    println("Reverse sorted fruits: $reverseSortedFruits")
}