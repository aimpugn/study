package combination

fun <T> combinations(items: List<T>, k: Int): List<List<T>> {
    if (k == 0) return listOf(emptyList())
    if (items.isEmpty()) return emptyList()

    // 현재 요소(head)
    val head = items.first()
    println("head: $head")
    // 현재 요소(head)를 제외한 나머지 요소들로 새로운 리스트(tail)를 만듭니다.
    // 재귀 호출 시 항상 tail을 사용함으로써 이미 고려한 요소는 다시 고려되지 않습니다.
    val tail = items.drop(1)

    // 현재 요소를 포함하는 경우
    val withHead = combinations(tail, k - 1).map { println("it + head: ${it + head}"); it + head; }
    // 현재 요소를 포함하지 않는 경우
    val withoutHead = combinations(tail, k)

    return withHead + withoutHead
}

fun main() {
    val items = listOf('A', 'B', 'C', 'D')

    for (k in 0..items.size) {
        println("$k-combinations:")
        combinations(items, k).forEach { println(it) }
        println()
    }
}