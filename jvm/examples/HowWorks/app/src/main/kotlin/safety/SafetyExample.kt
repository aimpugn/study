package safety

import kotlinx.coroutines.*
import util.*
import java.util.*
import java.util.concurrent.CopyOnWriteArrayList
import kotlin.concurrent.thread
import kotlin.properties.Delegates
import kotlin.random.Random
import kotlin.reflect.KFunction
import kotlin.reflect.KProperty

@RunExample
fun unsafeCalculateByThread() {
    for (i in 0..2) {
        var num = 0
        val threads = mutableListOf<Thread>()

        repeat(1000) {
            threads.add(thread {
                Thread.sleep(10)
                num += 1
            })
        }

        threads.forEach {
            /**
             * [Thread.join] 시 각 Thread 객체 자체가 monitor/lock으로 사용됩니다.
             * 값이 0이면 현재 synchronized 블록에 진입한 스레드에 대해 `while (isAlive())` 체크를 합니다.
             *
             * 스레드가 살아있다면 [Object.wait]를 호출합니다.
             * [VirtualThread]가 아니므로, [Object.wait0]을 호출하고 리턴합니다.
             */
            it.join(0)
        }
        println(num)
    }
    // Output:
    //  996
    //  992
    //  985
}

/**
 * [suspend] 함수 내에서는 코루틴의 실행이 비동기적으로 진행됩니다.
 *
 * 다음과 같은 [suspend] 함수 `example`:
 * ```
 * suspend fun example() { }
 * ```
 *
 * [kotlin.coroutines.Continuation]을 인자로 받는 메서드로 컴파일 됩니다:
 * ```
 * public final Object example(Continuation<? super Unit> $completion) { }
 * ```
 *
 * 이 변환은 Kotlin 컴파일러가 CPS(Continuation-Passing Style) 변환을 수행한 결과입니다
 *
 * References:
 * - [Asynchronous programming with coroutines](https://kotlinlang.org/spec/asynchronous-programming-with-coroutines.html)
 */
@RunExample
suspend fun unsafeCalculateByCoroutineScope() {
    repeat(3) {
        var num = 0
        /**
         * 새로운 [CoroutineScope]를 생성합니다.
         * [CoroutineScope]가 종료되거나 취소되면 해당 범위 내 모든 코루틴이 함께 종료되는데, 이를 통해 안정적인 동시성 관리를 제공합니다.
         */
        coroutineScope {
            /**
             * [coroutineScope]를 호출하면 새로운 [Job]이 생성되고, 이 [Job]은 외부 범위(outer scope)의 [Job]을 부모로 설정하게 됩니다.
             * [Job]은 [CoroutineScope.coroutineContext]를 통해 접근할 수 있습니다.
             * 여기서 외부 범위(outer scope)는 호출한 suspend [unsafeCalculateByCoroutineScope] 함수가 됩니다.
             */
            repeat(1000) {

                /**
                 * 현재 스레드를 차단하지 않으면서 코루틴을 생성하고, 생성한 코루틴에 대한 참조를 [Job] 타입으로 리턴합니다.
                 */
                launch {
                    delay(10)
                    num += 1
                }
            }
        }
        println(num)
    }
    // Output:
    //  987
    //  981
    //  970
}

@RunExample
fun safeCalculateByCoroutineScope() {
    /**
     * [runBlocking]은 최상위 [CoroutineScope] 역할을 합니다.
     * 내부의 모든 코루틴 작업이 완료될 때까지 현재 스레드를 차단(block)합니다.
     *
     * 일반적으로 최상위 코루틴 스코프에서 비동기 작업을 동기적으로 실행할 때 사용됩니다.
     */
    runBlocking {
        // 따라서 이 경우에는 모든 결과가 정확하게 매번 1000이 출력됩니다.
        unsafeCalculateByCoroutineScope()
        // Output:
        //  1000
        //  1000
        //  1000
    }
}

fun randomInt(): Int {
    print("Calculating... ")
    return Random.nextInt(1000, 9999)
}

/**
 * `val`은 읽기 전용(read-only) 참조를 의미하며, 초기화된 후 참조 자체를 변경할 수 없습니다.
 * 읽기 전용(read-only)는 데이터에 접근하는 방법에 제한을 두지만, 실제 데이터의 변하지 않음을 보장하지 않습니다.
 * 마찬가지로 `val`은 단순히 참조를 고정시켜줄 뿐, 그 참조가 가리키는 객체의 내용까지 불변으로 만드는 것은 아닙니다.
 *
 * [valIsReadOnlyNotImmutable] 호출 전에 이미 'Calculating...'이 출력됩니다.
 */
val fizz = randomInt() // 여기서 한번 실행되고, 최초 한번 계산된 랜덤 값을 갖고 있습니다.
val buzz
    /**
     * 여기서 [buzz]는 초기화된 값이 아니라, get() 접근자를 통해 [randomInt] 함수의 결과를 반환합니다.
     * 즉, [buzz]를 호출할 때마다 [randomInt] 함수가 다시 호출됩니다.
     * 만약 [randomInt]가 각 호출 때마다 다른 결과를 반환하는 함수라면, [buzz]의 값도 매번 달라질 수 있습니다.
     *
     * 즉, `val`은 "참조의 불변성"을 의미할 뿐, 참조가 가리키는 객체나 데이터의 불변(immutable)을 의미하지 않습니다.
     */
    get() = randomInt()

val foo by lazy {
    println("Initialize foo")
    randomInt()
}

class RandomValueDelegate {
    /**
     * 속성의 값을 가져올 때 호출되는 메서드입니다.
     *
     * @param thisRef 이 속성을 소유하고 있는 객체(예: 클래스 인스턴스)를 가리킵니다. `null`일 수도 있습니다.
     * @param property 속성의 메타데이터 정보를 담고 있는 [KProperty] 객체입니다. 이를 통해 속성 이름이나 타입 등 다양한 정보를 가져올 수 있습니다.
     */
    operator fun getValue(thisRef: Example, property: KProperty<*>): Int {
        // println("current property: $property")
        // 속성 이름에 따라 범위를 다르게 설정
        return when (property.name) {
            "smallRange" -> (10..100).random()
            "largeRange" -> (1000..2000).random()
            else -> (1..9).random()  // 기본값
        }
    }
}

class Example {
    /**
     * `by` 키워드는 속성 위임을 설정하는 역할을 합니다.
     * `by`를 사용해 특정 객체에 속성 처리를 맡기면, 속성 접근 시 해당 객체가 정의한 로직에 따라 값을 처리합니다.
     */
    val smallRange: Int by RandomValueDelegate()
    val largeRange: Int by RandomValueDelegate()
    val defaultRange: Int by RandomValueDelegate()
}

val bar = Example()

/**
 * `immutable`은 데이터 자체가 수정 불가능함을 의미합니다.
 *
 * 따라서 `immutable` 객체는:
 * - 생성 이후 절대 변경되지 않습니다.
 * - 참조가 여러 곳에서 이뤄져도 항상 동일한 값을 가지므로 스레드 세이프(Thread-safe)한 특징을 가집니다.
 */
@RunExample
fun valIsReadOnlyNotImmutable() {
    println(fizz)
    println(fizz)
    // 하지만 매번 getter를 사용하여 호출할 때마다 매번 다른 값이 출력될 수 있습니다.
    println(buzz)
    println(buzz)
    // Output:
    //  2494
    //  2494
    //  Calculating... 9655
    //  Calculating... 2871

    // delegate를 사용
    println(foo)
    // Initialize foo
    // Calculating... 8793
    println("Already initialized foo: $foo") // 이미 초기화 됐기 때문에 'Initialize foo' 또는 'Calculating...'는 출력되지 않습니다.
    // 8793

    // by 키워드 통해 delegate 하여 매번 서로 다른 결과를 나오게 할 수도 있습니다.
    var num = 1
    println("${num++}. bar.defaultRange: ${bar.defaultRange}") // 1. bar.defaultRange: 4
    println("${num++}. bar.defaultRange: ${bar.defaultRange}") // 2. bar.defaultRange: 7
    println("${num++}. bar.smallRange: ${bar.smallRange}") // 3. bar.smallRange: 68
    println("${num++}. bar.smallRange: ${bar.smallRange}") // 4. bar.smallRange: 37
    println("${num++}. bar.largeRange: ${bar.largeRange}") // 5. bar.largeRange: 1480
    println("${num++}. bar.largeRange: ${bar.largeRange}") // 6. bar.largeRange: 1029

    String("test".toCharArray())
}

@RunExample
fun safeCalculateBySynchronizedThread() {
    repeat(3) {
        val lock = Any()
        var num = 0
        val threads = List(1000) {
            thread {
                Thread.sleep(10)
                synchronized(lock) {
                    num += 1
                }
            }
        }

        threads.forEach { it.join() }
        println(num)
        // Output
        //  1000
        //  1000
        //  1000
    }
}

var part1 = "first"
var part2 = "second"
val partsConcat
    get() = "$part1 and $part2"

@RunExample
fun valueOfReadOnlyValCanBeMutable() {
    var num = 1
    println("${num++}. partsConcat: $partsConcat")
    part1 = "modified_first"
    println("${num++}. partsConcat: $partsConcat")
    // Output:
    //  1. partsConcat: first and second
    //  2. partsConcat: modified_first and second
}

@RunExample
fun useImmutableAsImmutable() {
    /**
     * 여기서 [list]는 [java.util.Arrays.ArrayList]입니다.
     * [java.util.Arrays.ArrayList]는 [java.util.AbstractList]를 상속하지만,
     * [java.util.AbstractList.add]를 오버라이드하지 않습니다.
     * 따라서 아래 `if`문에서 `add()`로 요소를 추가하려고 하면 [java.lang.UnsupportedOperationException] 익셉션이 발생합니다.
     */
    val list: List<Int> = listOf(1, 2, 3)
    println("Before: $list")
    if (list is MutableList) { // List<Int> 타입을 MutableList<Int> 타입으로 다운 캐스팅
        // list.add(4) // Exception in thread "main" java.lang.UnsupportedOperationException
        /**
         * [java.util.Arrays.ArrayList] 경우 `add()`는 구현하지 않았지만,
         * [java.util.Arrays.ArrayList.set]은 구현이 되어 있어서 억지로 수정한다면 수정할 수 있습니다.
         */
        list[2] = 4
    }
    println("After: $list")
}

class Person(val name: String, val mutableHobbies: MutableList<String>, val immutableHobbies: List<String>) {
    /**
     * [mutableHobbies]에 대한 방어적 복사입니다.
     */
    fun getMutableHobbiesAsImmutable(): List<String> {
        return mutableHobbies.toList()
    }
}

/**
 * `immutable` 경우 데이터가 변경되지 않기 때문에, 방어적 복사가 필요없습니다.
 *
 * > 방어적 복사(defensive copy)?
 * > 원본 객체가 외부에서 의도치 않게 수정되는 것을 방지하는 기법입니다.
 * > 객체 내부의 가변 데이터를 외부에서 안전하게 사용하도록 보호하기 위해 객체의 복사본을 생성하여 제공합니다.
 * > 단, 새로운 객체를 생성하므로 메모리가 추가로 필요하고, 객체 생성과 복사로 인해 성능이 저하될 수 있습니다.
 */
@RunExample
fun immutableDoNotNeedDefensiveCopy() {
    val person = Person("anonymous", mutableListOf(), listOf())
    val hobby = "coding"

    val hobbiesCopy = person.mutableHobbies.toMutableList() // 방어적 복사 필요
    println("person.mutableHobbies: ${person.mutableHobbies}") // person.mutableHobbies: []
    hobbiesCopy.add(hobby)
    println("hobbiesCopy: $hobbiesCopy")// hobbiesCopy: [coding]

    // 방어적 복사 불필요
    person.immutableHobbies + hobby
    println("person.immutableHobbies + hobby: ${person.immutableHobbies + hobby}") // person.immutableHobbies + hobby: [coding]

    val immutableHobbies = person.getMutableHobbiesAsImmutable()
    /* immutableHobbies.add("YOU CAN NOT ADD") */ // Unresolved reference 'add'.
    println("immutableHobbies: $immutableHobbies") // immutableHobbies: []
}

data class User(val name: String, val surname: String)

@RunExample
fun dataClassForCopy() {
    var user = User("anony", "mous")
    println("Original: $user")
    println("Copy: ${user.copy(surname = "mooous")}")
    // Output
    //  Original: User(name=anony, surname=mous)
    //  Copy: User(name=anony, surname=mooous)
}

@RunExample
fun readOnlyListIsNotSynchronized() {
    var readOnlyList = listOf<Int>()
    val threads = mutableListOf<Thread>()
    repeat(1000) {
        threads += thread {
            readOnlyList += 1
        }
    }
    threads.forEach { it.join() }
    println("readOnlyList: ${readOnlyList.size}") // readOnlyList: 990
}

@RunExample
fun threadSafeListExamples() {
    repeat(3) {
        val synchronizedList1 = Collections.synchronizedList(mutableListOf<Int>())
        val synchronizedList1Threads = mutableListOf<Thread>()
        repeat(1000) {
            synchronizedList1Threads += thread {
                synchronizedList1 += 1
            }
        }
        synchronizedList1Threads.forEach { it.join() }
        println("Collections.synchronizedList: ${synchronizedList1.size}") // readOnlyList: 990
    }
    // Output:
    //  Collections.synchronizedList: 1000
    //  Collections.synchronizedList: 1000
    //  Collections.synchronizedList: 1000

    repeat(3) {
        val synchronizedList2 = CopyOnWriteArrayList<Int>()
        val synchronizedList2Threads = mutableListOf<Thread>()
        repeat(1000) {
            synchronizedList2Threads += thread {
                synchronizedList2 += 1
            }
        }
        synchronizedList2Threads.forEach { it.join() }
        println("CopyOnWriteArrayList: ${synchronizedList2.size}") // readOnlyList: 990
    }
    // Output:
    //  CopyOnWriteArrayList: 1000
    //  CopyOnWriteArrayList: 1000
    //  CopyOnWriteArrayList: 1000
}

/**
 * [MutableCollection.plusAssign] 통해서 내부 상태를 직접 수정하게 되는데,
 * 이때의 Thread safety는 해당 [MutableList] 구현체에 의존합니다.
 */
@RunExample
fun synchronizationOfMutableListDependsOnItsImplementation() {
    val mutableList = mutableListOf<Int>()
    val threads = mutableListOf<Thread>()
    repeat(1000) {
        threads += thread {
            mutableList += 1
        }
    }
    threads.forEach { it.join() }
    println("mutableList: ${mutableList.size}") // readOnlyList: 999
}

@RunExample
fun trackMutableProperty() {
    /**
     * mutable property 경우:
     *
     * readonly 리스트의 `+=` 연산은 [Collection.plus]를 통해 이뤄집니다.
     * 새로운 리스트를 만들어서 다시 `names` 변수에 할당하므로, `names`라는 변수가 참조하는 자료구조가 변경됩니다.
     * 이를 통해 오브젝트가 변경되는 방식을 더 잘 제어할 수 있습니다.
     */
    var names by Delegates.observable(listOf<String>()) { _, old, new ->
        println("Names changed from $old to $new")
    }
    names += "Fabio"
    names += "Bill"
    // Output:
    //  Names changed from [] to [Fabio]
    //  Names changed from [Fabio] to [Fabio, Bill]

    /**
     * mutable collection 경우:
     *
     * 반면 이 경우에는 `names2`라는 변수가 변하는 게 아니라,
     * [MutableCollection.plusAssign] 통해 변수에 할당된 자료구조에 추가됩니다.
     * 따라서 감지할 변화가 없으므로 어떤 출력도 없습니다.
     */
    val names2 by Delegates.observable(mutableListOf<String>()) { _, old, new ->
        println("Names changed from $old to $new")
    }
    names2 += "Anony"
    names2 += "Mous"
    // Output: 없음
}


suspend fun main() {
    val mainRef: KFunction<Unit> = ::main
    val javaClass: Class<KFunction<Unit>> = mainRef.javaClass
    RunExampleInvoker.invoke(javaClass.packageName)
}
