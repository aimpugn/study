package functions

import util.RunExample

/**
 * [fold]는 함수형 프로그래밍에서 재귀적 데이터 구조를 순회하면서 값을 결합하는 고차 함수를 뜻합니다.
 *
 * `fold`라는 단어는 일반적으로 접다, 겹치다라는 의미를 가지고 있습니다.
 * 함수형 프로그래밍에서 `fold`는 데이터 구조의 계층적 구조(hierarchy)를 접어서(`combine`, `reduce`) 하나의 값으로 합치는 과정을 나타냅니다.
 * 즉, 각 요소를 순회하면서 이전 단계의 누적 값에 현재 요소를 적용하여 최종적으로 하나의 결과를 만들어내는 연산입니다.
 *
 * `fold`는 right-fold와 left-fold 두 가지가 있습니다:
 * - right-fold: 첫 번째 요소와 나머지 요소를 재귀적으로 결합한 결과를 결합합니다.
 *   오른쪽에서 왼쪽으로 순회합니다.
 *    ```
 *    1 + (2 + (3 + (4 + (5 + 0))))
 *    ```
 * - left-fold: 마지막 요소를 제외한 모든 요소를 재귀적으로 결합한 결과와 마지막 요소를 결합합니다.
 *   왼쪽에서 오른쪽으로 순회합니다.
 *    ```
 *    ((((0 + 1) + 2) + 3) + 4) + 5
 *    ```
 *
 * 코틀린에서 [kotlin.collections.fold]는 left-fold 방식으로 구현되어 있고,
 * [kotlin.collections.foldRight]는 right-fold 방식으로 구현되어 있습니다.
 *
 * ```
 * public static final int product(@NotNull Iterable $this$product) {
 *    Intrinsics.checkNotNullParameter($this$product, "<this>");
 *    int initial$iv = 1;
 *    int $i$f$fold = 0;
 *    int accumulator$iv = initial$iv;
 *
 *    for(Object element$iv : $this$product) {
 *       int element = ((Number)element$iv).intValue();
 *       int var9 = 0;
 *       accumulator$iv *= element;
 *    }
 *
 *    return accumulator$iv;
 * }
 * ```
 *
 * References:
 * - [Fold(higher-order function)](https://en.wikipedia.org/wiki/Fold_(higher-order_function))
 * - [Higher-order function](https://en.wikipedia.org/wiki/Higher-order_function)
 */
fun Iterable<Int>.productByFold() = fold(1) { acc, element -> acc * element }

/**
 * [reduce]는 [fold]와 유사하지만 초기값을 제공하지 않고, 컬렉션의 첫 번째 요소를 초기값으로 사용합니다.
 *
 * References:
 * - [stdlib - reduce](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/reduce.html)
 */
fun Iterable<Int>.productByReduce() = reduce { acc, element -> acc * element }

/**
 * [map]은 컬렉션의 각 요소에 주어진 변환 함수를 적용하여 새로운 컬렉션을 생성합니다.
 *
 * References:
 * - [stdlib - map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/map.html)
 */
fun Iterable<Int>.squaredByMap() = map { it * it }

/**
 * [flatMap]은 각 요소에 변환 함수를 적용한 후, 결과를 평탄화(flatten)합니다.
 *
 * References:
 * - [stdlib - flatMap](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/flat-map.html)
 */
fun Iterable<Int>.duplicated() = flatMap { listOf(it, it) }


@RunExample
fun extensionsExample() {
    println("[2, 5, 10].productByFold(): ${listOf(2, 5, 10).productByFold()}") // listOf(2, 5, 10).product(): 100
    println("[2, 5, 10].productByReduce(): ${listOf(2, 5, 10).productByReduce()}") // listOf(2, 5, 10).product(): 100
}