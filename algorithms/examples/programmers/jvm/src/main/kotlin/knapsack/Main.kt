package knapsack


data class Item(val weight: Int, val value: Int)

/**
 * 배낭의 용량을 초과하지 않으면서 아이템의 총 가치를 최대화합니다.
 * 각 아이템은 한 번만 포함될 수 있으며, 분할할 수 없습니다.
 * NP-난해(NP-Hard) 문제로, 모든 경우의 수를 고려해야 최적해를 찾을 수 있습니다.
 */
fun knapsack01(items: List<Item>, capacity: Int): Int {
    val n = items.size
    // dp[i][w]는 첫 번째 i개의 아이템으로 용량 w까지의 최대 가치를 저장합니다.
    val dp = Array(n + 1) { IntArray(capacity + 1) }

    // 아래에서부터 위로 DP 테이블을 채워나갑니다.
    for (i in 1..n) {
        val item = items[i - 1]
        for (w in 0..capacity) {
            if (item.weight <= w) {
                dp[i][w] = maxOf(
                    dp[i - 1][w],  // 현재 아이템을 포함하지 않는 경우
                    dp[i - 1][w - item.weight] + item.value  // 현재 아이템을 포함하는 경우
                )
            } else {
                dp[i][w] = dp[i - 1][w]  // 아이템을 포함할 수 없는 경우
            }
        }
    }

    // 최대 가치는 dp[n][capacity]에 저장됩니다.
    return dp[n][capacity]
}

data class FractionalItem(val weight: Double, val value: Double) {
    val valuePerWeight: Double
        get() = value / weight
}

fun fractionalKnapsack(items: List<FractionalItem>, capacity: Double): Double {
    // Sort items by value per weight in descending order
    val sortedItems = items.sortedByDescending { it.valuePerWeight }
    var totalValue = 0.0
    var remainingCapacity = capacity

    for (item in sortedItems) {
        if (remainingCapacity >= item.weight) {
            // Take the whole item
            totalValue += item.value
            remainingCapacity -= item.weight
        } else {
            // Take the fraction of the item that fits
            totalValue += item.valuePerWeight * remainingCapacity
            break  // Knapsack is full
        }
    }

    return totalValue
}

fun main() {
    val items = listOf(
        Item(weight = 2, value = 3),
        Item(weight = 3, value = 4),
        Item(weight = 4, value = 5),
        Item(weight = 5, value = 6)
    )
    val maxValue = knapsack01(items, 5)
    println("0-1 Knapsack Maximum value achievable: $maxValue")

    val fractionalItems = listOf(
        FractionalItem(weight = 10.0, value = 60.0),
        FractionalItem(weight = 20.0, value = 100.0),
        FractionalItem(weight = 30.0, value = 120.0)
    )
    val fractionalKnapsackMaxValue = fractionalKnapsack(fractionalItems, 50.0)
    println("Fractional Knapsack Maximum value achievable: $fractionalKnapsackMaxValue")
}
