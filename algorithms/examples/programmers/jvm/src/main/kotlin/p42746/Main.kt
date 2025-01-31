package p42746

import java.util.PriorityQueue

/**
 * [가장 큰 수](https://school.programmers.co.kr/learn/courses/30/lessons/42746?language=kotlin)
 */
class Solution {
    /**
     * 0 또는 양의 정수가 주어졌을 때, 정수를 이어 붙여 만들 수 있는 가장 큰 수
     *
     * 최악의 경우를 고려한다면, "999"가 100,000개인 경우,
     * "999,999,999,...,999" 같은 수가 될 수 있습니다.
     *
     * Examples:
     * - [6, 10, 2]
     *    - [6102, 6210, 1062, 1026, 2610, 2106]
     *    - 가장 큰 수는 6210
     *
     * @param numbers 0 또는 양의 정수가 담긴 배열
     * - 1 <= numbers.size <= 100,000
     * - 0 <= numbers.i <= 1,000
     * @return 순서를 재배치하여 만들 수 있는 가장 큰 수
     */
    fun solution(numbers: IntArray): String {
        // [6, 10, 2]
        // - 6102, 6210
        // - 1061, 1026,
        // - 2610, 2106
        // 가장 큰 수: 6/2/10
        //
        // [3, 30, 34, 5, 9]
        // - 3/30/34/5/9, 3/30/34/9/5, 3/30/5/34/9, 3/30/5/9/34, 3/30/9/34/5, 3/30/9/5/34..
        // 가장 큰 수: 9/5/34/3/30
        //
        // 첫째 자리수가 가장 큰 수를 가장 앞으로 보냅니다.
        // [9000, 8, 81] 배열이 주어질 경우, 9000/8/81 이 8/9000/81이나 8/81/9000보다 큽니다.
        // 하지만 앞자리가 같은 경우:
        // - 자리수가 같은 경우: 단순 비교
        //   - 33/34 (X)
        //   - 34/33 (O)
        //
        // - 자리수가 다른 경우
        //   - 3/34 (X)
        //   - 34/3 (O)
        //      33 < 34 따라서 [34, 3]
        //   - 30/3 (X)
        //   - 3/30 (O)
        //      30 < 33 따라서 [3, 30]
        //   - 321/33 (X)
        //   - 33/321 (O)
        //      321 < 333 따라서 [33, 321]
        //   자리수가 부족한 경우, 부족한 자리수는 그 마지막 수가 계속 이어진다고 보면,
        //   무엇이 어떤 값을 더 크게 해야 할지 쉽게 비교할 수 있을 것 같습니다. => (X) 잘못된 방법
        //
        //   하지만, 제출 실패 후 다시 확인하니, 아래 같은 경우에는 정렬이 잘못될 수 있습니다.
        //   - 40/403 (O)
        //   - 403/40 (X)
        //      400 < 403 따라서 [403, 40], 하지만 [40, 403]이어야 함
        //   - 312/31
        //   - 31/312
        //      311 < 312 따라서 [312, 31], 하지만 [31, 312]이어야 함
        //   - 947/94
        //   - 94/947
        //      944 < 947 따라서 [947, 94], 하지만 [94, 947]이어야 함
        //   중복되는 부분이 있는 경우 정렬이 제대로 되지 않을 가능성이 있었습니다.
        //
        // 대신, 두 수를 합치는 방식을 비교하는 방법이 더 간단하고 확실합니다.
        // - 3/34 vs 34/3 => [34, 3]
        // - 33/34 vs 34/33 => [34, 33]
        // - 321/33 vs 33/321 => [33, 321]
        // - 40/403 vs 403/40 => [40, 403]
        // - 312/31 vs 31/312 => [31, 312]
        // - 947/94 vs 94/947 => [94, 947]
        //
        // 바로 정렬을 하거나, 우선순위 큐를 사용합니다.

        val strNumbers = numbers.map { it.toString() }.toTypedArray()
        quickSort(strNumbers, 0, strNumbers.size - 1)

        // 첫번째 요소가 0이면 모두 0이라는 의미이므로, "0"을 리턴합니다.
        if (strNumbers[0] == "0") {
            return "0"
        }

        val sb = StringBuilder()
        strNumbers.forEach {
            sb.append(it)
        }

        return sb.toString()
    }

    fun quickSort(numbers: Array<String>, low: Int, high: Int) {
        if (high <= low) {
            return
        }

        val pivotIdx = partition(numbers, low, high)
        quickSort(numbers, low, pivotIdx - 1)
        quickSort(numbers, pivotIdx + 1, high)
    }

    /**
     * lomuto partition
     *
     * 왼쪽은 모두 `pivot` 이하, 오른쪽은 모두 `pivot` 초과한 값들이 위치하게 만들고,
     * 피벗은 좌/우 파티션 가운데에 자리를 잡음으로써 위치가 정렬된 상태라 할 수 있습니다.
     * 즉, 파티션 종료 시, 피벗의 왼쪽과 오른쪽에 각각 더 이상 피벗이 갈 수 없는 값들로 나뉘어집니다.
     *
     * Reference:
     * - [Lomuto partition scheme](https://en.wikipedia.org/wiki/Quicksort#Lomuto_partition_scheme)
     */
    fun partition(numbers: Array<String>, low: Int, high: Int): Int {
        // 우측 끝 값을 피벗으로 설정합니다.
        val pivot = numbers[high]
        // 좌측 인덱스부터 비교합니다.
        var i = low

        for (j in low until high) {
            // 원래는 피벗보다 작은 값은 피벗 좌측, 피벗보다 큰 값은 피벗 우측에 위치시킵니다.
            // ```
            // // 일반적인 오름차순 정수 정렬
            // if (numbers[j] < pivot) { swap(...) }
            // ```
            //
            // 하지만 이 문제의 경우 두 숫자를 합쳤을 때 더 큰 값을 찾기 위해 비교해야 합니다.
            // 즉, "a+b"와 "b+a"를 비교하여 더 큰 값이 앞으로 가야 합니다. 이때, 'a=numbers[j]', 'b=피벗'에 해당합니다.
            // 예를 들어,
            // - a = "34", b = "3" 경우,
            //   "34/3" vs "3/34" => "34/3"가 더 크므로,
            //   numbers[j]인 "34"(a)가 피벗보다 크다고 판단되어서 앞으로 보내야 하므로, 스왑해야 합니다.
            //
            // - a = "34", b = "9" 경우,
            //   "34/9" vs "9/34" => "9/34"가 더 크므로,
            //   피벗인 "9"(b)가 더 크다고 판단되므로 스왑하지 않아야 합니다.
            //
            // if((numbers[j] + pivot) < (pivot + numbers[j])) { // 오름차순 정렬인 경우
            if ((numbers[j] + pivot) > (pivot + numbers[j])) {
                // swap
                numbers[i] = numbers[j].also { numbers[j] = numbers[i] }

                i++
            }
        }

        // 이 지점까지 피벗의 위치는 `high`였으므로,
        // 마지막으로 스왑된 지점 바로 다음 인덱스인 `i`와 스왑하여
        // 해당 피벗이 있어야 할 곳에 위치시킵니다.
        numbers[i] = numbers[high].also { numbers[high] = numbers[i] }

        // 그리고 피벗의 최종 위치를 반환합니다.
        return i
    }

    /**
     * 우선순위 큐를 이용한 솔루션입니다.
     */
    fun solutionByPriorityQueue(numbers: IntArray): String {
        val pq = PriorityQueue<String>(numbers.size) { a, b ->
            (b + a).compareTo(a + b)
        }

        numbers.forEach { number ->
            pq.offer(number.toString())
        }

        val sb = StringBuilder()
        while (pq.isNotEmpty()) {
            sb.append(pq.poll())
        }

        if (sb[0] == '0') {
            return "0"
        }

        return sb.toString()
    }
}

fun main() {
    val s = Solution()
    listOf(
        Pair(intArrayOf(6, 10, 2), "6210"),
        Pair(intArrayOf(3, 30, 34, 5, 9), "9534330"),
        Pair(intArrayOf(0, 1, 23, 4), "42310"),
        // 실패하는 케이스들이 많음.
        Pair(intArrayOf(91, 92, 93, 94), "94939291"),
        Pair(intArrayOf(0, 0, 1, 2), "2100"),
        Pair(intArrayOf(9), "9"),
        Pair(intArrayOf(0, 0, 0, 0), "0"),
        Pair(intArrayOf(9, 1), "91"),
        Pair(intArrayOf(3, 1000), "31000"),
        Pair(intArrayOf(300, 1000), "3001000"),
        Pair(intArrayOf(40, 403), "40403"),
    ).forEach { tc ->
        val result = s.solution(tc.first)
        println("result: $result, answer: ${result == tc.second}")
    }
}