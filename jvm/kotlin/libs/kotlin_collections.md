# Kotlin Collections

## [Kotlin Collections](https://kotlinlang.org/docs/collections-overview.html#collection-types)

## Java와 Kotlin 컬렉션 라이브러리 비교

- `java.util.*`: Java의 표준 컬렉션 라이브러리
- `kotlin.collections.*`: Kotlin의 컬렉션 라이브러리

Kotlin은 Java 컬렉션과 완벽하게 호환됩니다.
`java.util.*`의 클래스들을 Kotlin에서 사용할 수 있으며, Kotlin의 확장 함수들을 이 클래스들에 적용할 수 있습니다.

```kotlin
import java.util.ArrayList

val javaArrayList = ArrayList<String>()
javaArrayList.add("Kotlin")
javaArrayList.add("Java")

// Kotlin 확장 함수 사용
val uppercased = javaArrayList.map { it.uppercase() }
```

### (im)mutable

- Java: 기본적으로 모든 컬렉션이 가변(mutable)입니다.
- Kotlin: 불변(immutable)과 가변(mutable) 컬렉션을 명확히 구분합니다.

```kotlin
// Java 스타일 (가변)
val javaList: ArrayList<String> = ArrayList()

// Kotlin 불변 리스트
val immutableList: List<String> = listOf("a", "b", "c")

// Kotlin 가변 리스트
val mutableList: MutableList<String> = mutableListOf("a", "b", "c")
```

### Null Safety

- Java: 널 값을 허용합니다.
- Kotlin: 컬렉션 자체와 요소의 널 허용 여부를 타입 시스템에서 명시적으로 구분합니다.

```kotlin
// 널 허용 리스트 (Java와 유사)
val nullableList: List<String?> = listOf("a", null, "c")

// 널 비허용 리스트
val nonNullList: List<String> = listOf("a", "b", "c")
```

### 확장 함수 (Extension Functions)

- Java: 확장 함수 개념이 없습니다.
- Kotlin: 풍부한 확장 함수를 제공하여 컬렉션 조작을 간편하게 합니다.

```kotlin
val numbers = listOf(1, 2, 3, 4, 5)
val evenNumbers = numbers.filter { it % 2 == 0 }
val doubledNumbers = numbers.map { it * 2 }
```

### 코틀린 특화 컬렉션

- Kotlin은 `Sequence`와 같은 지연 평가(lazy evaluation) 컬렉션을 제공합니다.

```kotlin
val sequence = sequenceOf(1, 2, 3, 4, 5)
    .map { it * 2 }
    .filter { it % 3 == 0 }
    .toList()
```

### 연산자 오버로딩

- Java: 컬렉션에 대한 연산자 오버로딩을 지원하지 않습니다.
- Kotlin: 컬렉션에 대해 연산자 오버로딩을 지원합니다.

```kotlin
val list1 = listOf(1, 2, 3)
val list2 = listOf(4, 5, 6)
val combinedList = list1 + list2  // [1, 2, 3, 4, 5, 6]
```

### 타입 추론과 제네릭

- Java: 타입 추론이 제한적입니다.
- Kotlin: 향상된 타입 추론을 제공합니다.

```kotlin
// Kotlin
val numbers = listOf(1, 2, 3)  // List<Int>로 추론됨

// Java
List<Integer> numbers = Arrays.asList(1, 2, 3);
```

## Java의 Queue와 Kotlin의 ArrayDeque

### Java의 Queue

`java.util` 패키지에 속하는 `Queue`는 Java에서 인터페이스로 정의됩니다.
FIFO(First-In-First-Out) 구조를 나타냅니다.
`LinkedList`, `PriorityQueue`, `ArrayDeque` 등의 클래스가 이를 구현합니다.

주요 메서드는 다음과 같습니다.
- `offer(e)`: 요소 추가
- `poll()`: 첫 번째 요소 제거 및 반환
- `peek()`: 첫 번째 요소 조회 (제거하지 않음)

```java
import java.util.LinkedList;
import java.util.Queue;

Queue<String> queue = new LinkedList<>();
queue.offer("First");
queue.offer("Second");
String first = queue.poll(); // "First"
String peek = queue.peek(); // "Second"
```

## Kotlin의 ArrayDeque

`ArrayDeque`는 Java 구현 클래스 아닌 Kotlin `kotlin.collections` 패키지에 속하는 구현 클래스입니다.
양방향 큐(Deque) 기능을 제공합니다.

주요 메서드는 다음과 같습니다.
- `addFirst(e)`, `addLast(e)`: 앞/뒤에 요소 추가
- `removeFirst()`, `removeLast()`: 앞/뒤 요소 제거 및 반환
- `first()`, `last()`: 앞/뒤 요소 조회

동적 배열 기반으로 구현되어 있어 양 끝에서의 삽입/삭제가 효율적입니다.
리스트, 큐, 스택 등의 기능을 모두 제공합니다.
`null` 요소를 허용합니다.

```kotlin
val deque = ArrayDeque<String>()
deque.addLast("First")
deque.addLast("Second")
val first = deque.removeFirst() // "First"
val last = deque.last() // "Second"
```
