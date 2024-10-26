package primitive

fun main() {
    initArrays()
    arrayExtensions()
    typeConversion()
    sliceArray()
    associate()
    group()
}

fun initArrays() {
    println(title("initArrays"))
    // java:
    //  int[] ints = new int[5];
    //  int[] ints2 = {0, 0, 0, 0, 0};
    //  int[] ints3 = new int[]{0, 0, 0, 0, 0};
    val d1IntArray = IntArray(5)
    println(d1IntArray.contentToString()) // [0, 0, 0, 0, 0]

    // java:
    //  Integer[] integers = new Integer[5];
    val d1IntegerArray = Array<Int?>(5) { null }
    println(d1IntegerArray.contentToString()) // [null, null, null, null, null]

    // java:
    //  int[][] d2ints = new int[][]{};
    //  d2ints[0] = new int[3]; => Index 0 out of bounds for length 0
    //
    // Note:
    //  Array<IntArray>(size, init) 사용하여 위와 같은 결과를 낼 수 없습니다.
    val d2IntArray = arrayOf<IntArray>()
    println(d2IntArray.contentToString()) // []
    // d2IntArray[0] = IntArray(3) => Index 0 out of bounds for length 0

    // java:
    //  int[][] d2ints2 = new int[1][];
    val d2IntArray3 = arrayOfNulls<IntArray?>(1)
    println(d2IntArray3.contentToString()) // [null]

    // java:
    //  d2ints2[0] = new int[5];
    d2IntArray3[0] = IntArray(5)
    println(d2IntArray3.contentDeepToString()) // [[0, 0, 0, 0, 0]]

    // java:
    //  int[][] d2ints3 = new int[3][4];
    val d2IntArray4 = Array(3) { IntArray(4) }
    println(d2IntArray4.contentDeepToString()) // [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]
    // [Arrays are always mutable](https://kotlinlang.org/docs/arrays.html#access-and-modify-elements)
    d2IntArray4[0][0] = 1
    println(d2IntArray4.contentDeepToString()) // [[1, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]

    // java:
    //  String[][][] d3String = new String[1][1][1];
    val d3StringArray = Array(1) {
        Array(1) {
            arrayOfNulls<String?>(1)
        }
    }
    println(d3StringArray.contentDeepToString()) // [[[null]]]

    // java:
    //  String[][][] d3String2 = new String[1][1][];
    val d3StringArray2 = Array(1) {
        arrayOfNulls<Array<String?>>(1)
    }
    println(d3StringArray2.contentDeepToString()) // [[null]]
    // d3String2[0][0] = new String[5];
    d3StringArray2[0][0] = arrayOfNulls<String?>(5)
    println(d3StringArray2.contentDeepToString()) // [[[null, null, null, null, null]]]

    // 람다 함수를 사용하여 배열을 초기화할 수 있습니다. 크기가 5인 배열, 각 요소는 인덱스 * 3
    println(Array(5) { idx -> idx * 3 }.joinToString()) // 0, 3, 6, 9, 12
    println(BooleanArray(3) { idx -> idx % 2 == 0 }.joinToString()) // true, false, true
    println(DoubleArray(4) { i -> i * 1.5 }.joinToString()) // 0.0, 1.5, 3.0, 4.5

    // Kotlin에서 Array는 불공변(invariant)입니다.
    // 따라서 `Array<String>`를 `Array<Any>`에 설정할 수 없고, 그 반대도 마찬가지입니다.
    // var anys = arrayOf<Any>()
    // var strings = arrayOf<String>()
    // anys = strings => Type mismatch: inferred type is Array<String> but Array<Any> was expected
    // strings = anys => Type mismatch: inferred type is Array<Any> but Array<String> was expected
    // 더 자세한 내용은 [variance] 패키지를 참고합니다.
}

fun arrayExtensions() {
    // joinToString
    println(title("arrayExtensions: joinToString"))
    println(IntArray(5).joinToString()) // 0, 0, 0, 0, 0
    println(IntArray(5) { i -> i * 2 }.joinToString()) // 0, 2, 4, 6, 8
    println(intArrayOf(1, 2, 3, 4, 5).joinToString()) // 1, 2, 3, 4, 5
    println(listOf(1, 2, 3, 4, 5).joinToString()) // 1, 2, 3, 4, 5
    println(arrayOf(1, 2, 3, 4, 5).joinToString()) // 1, 2, 3, 4, 5
    println(arrayOfNulls<Int>(5).joinToString()) // null, null, null, null, null
    println(intArrayOf(5, 4, 3, 2).sorted().joinToString()) // 2, 3, 4, 5
    println(intArrayOf(5, 4, 3, 2).sortedArray().joinToString()) // 2, 3, 4, 5
    println(intArrayOf(1, 2, 3, 4).sortedArrayDescending().joinToString()) // 4, 3, 2, 1

    val arrToSortDesc = intArrayOf(1, 2, 3, 4)
    arrToSortDesc.sortDescending()
    println(arrToSortDesc.contentToString()) // [4, 3, 2, 1]
    val arrToSortAsc = intArrayOf(10, 9, 8, 7)
    arrToSortAsc.sort()
    println(arrToSortAsc.contentToString()) // [7, 8, 9, 10]
    println(arrToSortAsc.filter { it > 7 }.toList().joinToString()) // 8, 9, 10

    // 복사
    println(title("arrayExtensions: copy"))
    val original = intArrayOf(1, 2, 3)
    println(original.copyOf().contentToString()) // [1, 2, 3]
    println(original.copyOf(5).contentToString()) // [1, 2, 3, 0, 0]
    // 얕은 복사가 이뤄지므로, original2 내부의 배열을 참조하므로 같이 영향을 받습니다.
    val original2 = arrayOf<IntArray>(intArrayOf(0, 1, 2), intArrayOf(3, 4, 5))
    val clonedOriginal2 = original2.clone()
    val copiedOriginal2 = original2.copyOf()
    original2[0][0] = 1
    println(original2.contentDeepToString()) // [[1, 1, 2], [3, 4, 5]]
    println(clonedOriginal2.contentDeepToString()) // [[1, 1, 2], [3, 4, 5]]
    println(copiedOriginal2.contentDeepToString()) // [[1, 1, 2], [3, 4, 5]]
    // 따라서 복사를 해도 내부 배열들의 식별자 또는 hashCode() 메서드 결과값이 동일합니다.
    // ex: [I@4e50df2e
    //  Object 클래스의 기본 toString() 메서드에 의해 생성됩니다.
    //  각 부분의 의미는 다음과 같습니다:
    //  - `[`: 배열임을 의미합니다. `[[I`는 2차원 int 배열 (`int[][]`)을 나타냅니다.
    //  - <타입 식별자>: `I`면 int, `Ljava.lang.String`면 문자열
    //  - @: 구분자
    //  - <해시값>: 객체의 hashCode() 메서드 반환값의 16진수 표현입니다. 객체를 식별하는 데 사용됩니다.
    //      JVM은 이 해시코드를 사용하여 해시 기반 자료구조(ex: HashMap 등)에서 객체를 빠르게 찾습니다.
    println(
        original2[0].hashCode() == clonedOriginal2[0].hashCode()
            && original2[0].hashCode() == copiedOriginal2[0].hashCode()
    ) // true
    // 깊은 복사를 하려면 직접 구현합니다.
    // Reference:
    //  - https://www.baeldung.com/kotlin/clone-object
    fun Array<IntArray>.deepCopy(): Array<IntArray> {
        return Array<IntArray>(this.size) { idx ->
            this[idx].clone()  // 각 IntArray를 복사
        }
    }

    val deepCopiedOriginal2 = original2.deepCopy()
    println(deepCopiedOriginal2[0].hashCode() != original2[0].hashCode()) // true

    // 합, 평균, 최대, 최소 등 연산
    println(title("arrayExtensions: sum, sumOf, avg, max, min"))
    println(intArrayOf(1, 2, 3, 4, 5).sum()) // 15
    println(intArrayOf(1, 2, 3, 4, 5).average()) // 3.0
    println(intArrayOf(1, 2, 3, 4, 5).maxOrNull()) // 5
    println(intArrayOf(1, 2, 3, 4, 5).minOrNull()) // 1
    // println(intArrayOf().max()) => java.util.NoSuchElementException
    println(intArrayOf().maxOrNull()) // null
    println(booleanArrayOf(true, true, false).sumOf { if (it) 3.toInt() else 0 }) // 6
    println(booleanArrayOf(true, true, false).sumOf { if (it) 3L else 0L }) // 1
    println(arrayOf("hello", "world", "!").max()) // world
    println(arrayOf("hello", "world", "!").min()) // !

    // 내용 비교
    println(title("arrayExtensions: contentEquals"))
    println(intArrayOf(1, 2, 3).contentEquals(intArrayOf(1, 2, 3))) // true
    println(intArrayOf(1, 2, 3) === intArrayOf(1, 2, 3)) // false
    println(intArrayOf(1, 2, 3) == intArrayOf(1, 2, 3)) // false. IntArray의 `==`는 Any의 equals()를 사용하여 참조 비교가 이뤄집니다.
    println(listOf(1, 2, 3) == listOf(1, 2, 3)) // true. 반면 List의 equals는 내용을 비교합니다.

    val thatInner = intArrayOf(0, 1, 2)
    val that = arrayOf<IntArray>(thatInner)
    val other = arrayOf<IntArray>(intArrayOf(0, 1, 2))
    println(that.contentEquals(other)) // false. 참조를 비교하기 때문에 다르다고 판단합니다.
    println(that.contentDeepEquals(other)) // true
    val another = arrayOf<IntArray>(thatInner)
    println(that.contentEquals(another)) // true. 같은 내부 배열을 참조하므로 같다고 판단합니다.
    println(that.contentDeepEquals(another)) // true
}

fun typeConversion() {
    println(title("typeCasting"))

    // List<Int>를 IntArray(int[])로 변환할 수 있습니다.
    println(listOf(1, 2, 3, 4, 5).toIntArray().javaClass.canonicalName) // int[]
    // 반대로 IntArray(int[])를 List<Int>로 변환할 수도 있습니다.
    println(arrayOf(1, 2, 3).toList().javaClass.canonicalName) // java.util.ArrayList

    // byte[]를 List<Byte>로 변환할 수 있습니다.
    val byteList: List<Byte> = byteArrayOf(1, 2, 3, 4, 5 /*, 당연히 256 불가*/).toList()
    println(byteList.javaClass.canonicalName) // java.util.ArrayList
    println(byteList) // [1, 2, 3, 4, 5]

    // Double 배열 생성
    println(doubleArrayOf(1.1, 2.2, 3.3, 4.4, 5.5).javaClass.canonicalName) // double[]

    // Boolean 배열 생성
    println(booleanArrayOf(true, false, true).javaClass.canonicalName) // boolean[]

    // Char 배열 생성
    println(charArrayOf('a', 'b', 'c').javaClass.canonicalName) // char[]
}

fun sliceArray() {
    println(title("sliceArray"))

    // sliceArray: 배열 타입에만 사용 가능
    println(intArrayOf(1, 2, 3, 4, 5).sliceArray(1..3).contentToString()) // [2, 3, 4]. IntArray(==int[])
    // slice: 배열과 컬렉션 모두에 사용 가능
    println(intArrayOf(1, 2, 3, 4, 5).slice(1..3)) // [2, 3, 4]. List<Int>
}

fun associate() {
    println(title("associate"))
    println(arrayOf("hello", "world", "!!").associate { s -> Pair(s, s.length) }) // {hello=5, world=5, !!=2}

    // associate*With*~ 경우, valueSelector를 지정하면 기존 배열의 각 요소가 해당 값과 매핑되는 key가 됩니다.
    println(arrayOf("hello", "world", "!!").associateWith { s -> s.length }) // {hello=5, world=5, !!=2}
    println(
        arrayOf(
            "hello",
            "world",
            "!!"
        ).associateWith {
            // valueSelector
            "val_${it}_${it.length}"
        }
    ) // {hello=val_hello_5, world=val_world_5, !!=val_!!_2}

    // associate*By*~ 경우, keySelector를 지정하면 기존 배열의 각 요소가 해당 키와 매핑되는 value가 됩니다.
    println(
        arrayOf("hello", "world", "!!").associateBy {
            // keySelector
            "key_${it}_${it.length}"
        }
    ) // {key_hello_5=hello, key_world_5=world, key_!!_2=!!}

    println(
        arrayOf("hello", "world", "!!").associateByTo(mutableMapOf()) {
            // keySelector
            "key_${it}_${it.length}"
        }
    ) // {key_hello_5=hello, key_world_5=world, key_!!_2=!!}
}

fun group() {
    println(title("group"))

    // Map<Char, List<String>>
    println(arrayOf("hi", "hello", "world").groupBy { it[0] }) // {h=[hi, hello], w=[world]}
    // Map<String, List<String>>
    println(arrayOf("hi", "hello", "world").groupBy { it[0].toString() }) // {h=[hi, hello], w=[world]}
    // MutableMap<String, MutableList<String>>
    println(
        arrayOf(
            "hi",
            "hello",
            "world"
        ).groupByTo(mutableMapOf()) { it[0].toString() }
    ) // {h=[hi, hello], w=[world]}

    // groupingBy 사용하여 Grouping<String, String> 타입을 리턴할 수 있습니다.
    println(arrayOf("hi", "hello", "world").groupingBy { it[0].toString() }.eachCount()) // {h=2, w=1}
}

fun title(title: String): String = "=".repeat(10) + " $title " + "=".repeat(10)
