package p43165

/**
 * [타겟 넘버](https://school.programmers.co.kr/learn/courses/30/lessons/43165?language=kotlin)
 */
class Solution {
    /**
     * - n 개의 음이 아닌 정수 배열
     * - 순서 바꾸지 않음
     * - 더하거나 빼서 목표 숫자를 만들기
     *
     * @param numbers 사용할 수 있는 숫자 배열
     * - 2 <= numbers.length <= 20
     * - 1 <= numbers.i <= 50
     * @param target 만들어야 하는 목표 숫자
     * - 1 <= target <= 1,000
     *
     * @return 방법의 수
     */
    fun solution(numbers: IntArray, target: Int): Int {
        // [1, 1, 1, 1, 1] | 3
        // -1 +1 +1 +1 +1 = 3
        // +1 -1 +1 +1 +1 = 3
        // +1 +1 -1 +1 +1 = 3
        // +1 +1 +1 -1 +1 = 3
        // +1 +1 +1 +1 -1 = 3
        //
        // [4, 1, 2, 1] | 4
        // 4 -1 +2 -1 = 4
        // 4 +1 -2 +1 = 4
        //
        // 1. 순서는 바꾸지 않으므로 numbers 배열에 대해 순차적으로 가능한 경우의 수 따져보기?
        // DFS 통해 각 자리별로 +/- 만들기
        //
        // 2. 다 더한 결과와 만들어야 하는 숫자 사이의 차를 구하고, 그 차를 만들 수 있는 경우의 수 구하기?
        // 복잡할 거 같고, 조건을 많이 따져야 할 거 같음

        return dfs(numbers, target, 0, 0)
    }

    fun dfs(numbers: IntArray, target: Int, currIdx: Int, currentSum: Int): Int {
        if (currIdx == numbers.size) {
            if (currentSum == target) {
                return 1
            }
            return 0
        }

        var answer = dfs(numbers, target, currIdx + 1, currentSum + numbers[currIdx])
        answer += dfs(numbers, target, currIdx + 1, currentSum - numbers[currIdx])

        return answer
    }

    /**
     * 다른 사람 풀이
     */
    fun solution2(numbers: IntArray, target: Int): Int {
        val answer = numbers.fold(listOf(0)) { initial, int ->
            println("before initial: $initial, int: $int")
            val tmp2 = initial.run {
                val tmp3 = map { it + int }
                val tmp4 = map { it - int }
                println("\ttmp3: $tmp3, tmp4: $tmp4")

                tmp3 + tmp4
            }
            println("after initial: $tmp2, int: $int")

            tmp2
        }.count { it == target }
        println(answer)

        return answer
    }
}

fun main() {
    val s = Solution()

    listOf(
        Triple(intArrayOf(1, 1, 1, 1, 1), 3, 5),
        Triple(intArrayOf(4, 1, 2, 1), 4, 2),
    ).forEach { tc ->
        println(s.solution(tc.first, tc.second) == tc.third)
    }
}