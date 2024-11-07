package datastructures.list

fun main() {
    fun printTitle(title: String) = println("${"=".repeat(10)} $title ${"=".repeat(10)}")

    printTitle("초기화")
    // ArrayList<String>
    println(arrayListOf("array", "list", "of")) // [array, list, of]
    val immutableList = listOf("apple", "banana", "cherry")
    println(immutableList) // [apple, banana, cherry]
    // immutableList.add("can not add") => Unresolved reference: add
    println(List(3) {
        "$it as String"
    }) // [0 as String, 1 as String, 2 as String]

    val mutableList = mutableListOf("dog", "cat", "bird")
    printTitle("읽기")
    println(mutableList[0]) // dog
    println(mutableList.get(0)) // dog

    printTitle("수정: 추가, 목록 추가")
    println(mutableList) // [dog, cat, bird]
    mutableList.add("flower")
    println(mutableList) // [dog, cat, bird, flower]
    mutableList.add("fish")
    println(mutableList) // [dog, cat, bird, flower, fish]
    mutableList.addFirst("animal")
    println(mutableList) // [animal, dog, cat, bird, flower, fish]

    val arrayList = ArrayList<String>()
    arrayList.addAll(listOf("red", "green", "blue"))
    println(arrayList) // [red, green, blue]

    printTitle("수정: 대체")
    val replaceTest = mutableList.toMutableList()
    // 원본: [animal, dog, cat, bird, flower, fish]
    replaceTest[4] = "flower_by_idx"
    println(replaceTest) // [animal, dog, cat, bird, flower_by_idx, fish]
    // replaceTest[10] = "added_by_idx" // => java.lang.IndexOutOfBoundsException: Index 10 out of bounds for length 6

    replaceTest.replaceAll({ el -> el.uppercase() })
    println(replaceTest) // [ANIMAL, DOG, CAT, BIRD, FLOWER_BY_IDX, FISH]

    printTitle("수정: 삭제")
    mutableList.removeAt(2)
    println(mutableList) // [animal, dog, bird, flower, fish]: cat 삭제
    println(mutableList.remove("flower")) // true
    println(mutableList.remove("unknown")) // false

    printTitle("탐색")
    // [animal, dog, bird, fish]
    mutableList.add("bird")
    // [animal, dog, bird, fish, bird]: bird를 추가
    println("bird" in mutableList) // true. contains() 메서드 호출과 같습니다.
    println(mutableList.indexOf("bird")) // 2
    println(mutableList.lastIndexOf("bird")) // 4

    // 정렬
    printTitle("정렬")
    mutableList.sort() // 원래 리스트의 순서를 변경합니다.
    println(mutableList) // [animal, bird, bird, dog, fish]: in-place 정렬을 수행합니다.
    println(mutableList.asReversed()) // [fish, dog, bird, bird, animal]: 역정렬된 새로운 리스트를 반환합니다.
    println(mutableList) // [animal, bird, bird, dog, fish]: asReversed() 해도 원래 리스트의 순서는 그대로 유지됩니다.

    printTitle("순회")
    mutableList.forEach({ el -> println("forEach: $el") })
    // 출력:
    //  forEach: animal
    //  forEach: bird
    //  forEach: bird
    //  forEach: cat
    //  forEach: fish
    //  forEach: flower_by_idx
    mutableList.forEachIndexed({ idx, el -> println("forEach: [$idx] $el") })
    // 출력:
    //  forEach: [0] animal
    //  forEach: [1] bird
    //  forEach: [2] bird
    //  forEach: [3] cat
    //  forEach: [4] fish
    //  forEach: [5] flower_by_idx
    for ((idx, el) in mutableList.withIndex()) {
        println("forEach: [$idx] $el")
    }
    // 출력:
    //  forEach: [0] animal
    //  forEach: [1] bird
    //  forEach: [2] bird
    //  forEach: [3] cat
    //  forEach: [4] fish
    //  forEach: [5] flower_by_idx

    printTitle("맵핑")
    val numbers = listOf("one", "two", "three", "four")
    // Map<Int, String>. associate*With*를 사용하면 원래 리스트의 요소를 키, valueSelector의 결과를 값으로 사용하여 연관시킵니다.
    println(numbers.associateWith { it.length }) // {one=3, two=3, three=5, four=4}
    println(numbers.associateBy { it.length }) // {3=two, 5=three, 4=four}

    // mutable -> immutable 변형
    val immutableList2: List<Int> = mutableListOf(1, 2, 3).toList()
    // immutableList2.add(4) => Unresolved reference: add
    // immutableList2[0] = 4 => No set method providing array access
    println(immutableList2) // [1, 2, 3]

    printTitle("리스트 조작: subList")
    println(listOf("one", "two", "three", "four").subList(0, 2)) // [one, two]: toIndex는 포함하지 않습니다.

    printTitle("리스트 조작: slice")
    println(listOf("one", "two", "three", "four").slice(0..2)) // [one, two, three]

    printTitle("리스트 조작: drop")
    println(listOf("one", "two", "three", "four").drop(0)) // [one, two, three, four]: 아무것도 제거하지 않습니다.
    println(listOf("one", "two", "three", "four").drop(2)) // [three, four]: 앞의 두 개를 제거합니다.

    printTitle("리스트 조작: take")
    println(listOf("one", "two", "three", "four").take(0)) // []
    println(listOf("one", "two", "three", "four").take(2)) // [one, two]

    printTitle("리스트 조작: filter")
    println(listOf("one", "two", "three", "four").filter { it.contains("e") }) // [one, three]

    printTitle("리스트 조작: map")
    println(listOf("one", "two", "three", "four").map {
        when (it) {
            "one" -> "1_one"
            "two" -> "2_two"
            "three" -> "3_three"
            else -> it
        }
    }) // [1_one, 2_two, 3_three, four]
    println(
        listOf(
            "one",
            "two",
            null,
            "three",
            "four"
        ).mapIndexedNotNull { idx, it ->
            // 여기서 null을 리턴하지 않아야 새로운 리스트에 포함됩니다.
            it?.let {
                "($idx) $it"
            }
        }
    ) // [(0) one, (1) two, (3) three, (4) four]
    printTitle("리스트 조작: +")
    println(listOf("one", "two") + listOf("three", "four")) // [one, two, three, four]
    // 배열과 리스트를 합칠 수도 있습니다. 이때 좌측의 타입을 따르게 됩니다.
    println(
        (arrayOf("one", "two") + listOf(
            "three",
            "four"
        )).contentToString()
    ) // [one, two, three, four]: String[]이 됩니다.
    println(
        listOf("one", "two") + arrayOf(
            "three",
            "four"
        )
    ) // [one, two, three, four]: List<String>이 됩니다.
    println(
        mutableListOf("one", "two") + arrayOf(
            "three",
            "four"
        )
    ) // [one, two, three, four]: List<String>이 됩니다.
    println(
        mutableSetOf("one", "two") + arrayOf(
            "three",
            "four"
        )
    ) // [one, two, three, four]: Set<String>이 됩니다.
    println(
        (mutableSetOf("one", "two") + arrayOf(
            "three",
            "four"
        )) as LinkedHashSet
    ) // [one, two, three, four]: Set + 연산시 내부적으로 LinkedHashSet를 사용합니다.


    printTitle("리스트 조작: flatMap")
    println(listOf("one", "two", "three", "four", "five").flatMap {
        // 각 요소를 컬렉션과 맵핑하고, 모든 결과는 단일 레벨로 평탄화(flatten) 됩니다.
        when (it) {
            "one" -> listOf(1, "one", 1.5)
            "two" -> listOf(2, "two", 2.5)
            "three" -> listOf(3, "three", 3.5)
            "four" -> listOf(4, "four", 4.5)
            else -> listOf(0, "unknown", 0.0)
        }
    }) // [1, one, 1.5, 2, two, 2.5, 3, three, 3.5, 4, four, 4.5, 0, unknown, 0.0]

    printTitle("리스트 조작: zip")
    println(
        listOf("one", "two", "three", "four").zip(
            listOf(1, 2, 3, 4)
        )
    ) // [(one, 1), (two, 2), (three, 3), (four, 4)]
    println(
        listOf("one", "two", "three", "four").zip(
            listOf(1..2, 3..4) // List<IntRange>
        )
    ) // [(one, 1..2), (two, 3..4)]
    println(
        listOf("one", "two", "three", "four").zip(
            listOf(
                1,
                2
            )
        )
    ) // [(one, 1), (two, 2)]: 길이가 더 작은 리스트에 맞춰서 결과가 생성됩니다.

    val args = arrayOf("test", "value")
    printVararg(*args)
    printVararg(*arrayOf("test", "value"))
}

/**
 * [Variable number of arguments (varargs)﻿](https://kotlinlang.org/docs/functions.html#variable-number-of-arguments-varargs)
 */
fun printVararg(vararg args: String) {
    for (arg in args)
        println("vararg: $arg")
}
