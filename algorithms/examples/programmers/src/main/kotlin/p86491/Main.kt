package p86491

import kotlin.math.*

class Solution {
    /**
     * 명함 지갑 크기 정하기
     * 예를 들어, 명함 가로/세로 길이가 다음과 같을 경우
     * - 60 / 50
     * - 30 / 70
     * - 60 / 30
     * - 80 / 40
     *
     * 가로 최대값 80과 세로 최대값 70을 곱하면 모든 명함을 담을 수 있는 명함 지갑을 만들 수 있습니다.
     * 하지만 30/70 명함을 눕혀서 수납한다면 80 * 50 크기의 지갑으로 모든 명함 수납이 가능합니다.
     *
     * @param sizes 배열 [(w, h), ...] 형태의 배열.
     *              1 <= sizes.lenght <= 10,000.
     *              1 <= w, h <= 1,000 자연수.
     *
     * @see 최소직사각형 https://school.programmers.co.kr/learn/courses/30/lessons/86491?language=kotlin
     */
    fun solution(sizes: Array<IntArray>): Int {
        // 모두 담을 수 있는 가로 * 세로를 구합니다.
        var maxWidth = Int.MIN_VALUE
        var maxHeight = Int.MIN_VALUE

        sizes.forEach {
            // 더 큰 값을 너비로
            val w = max(it[0], it[1])
            // 더 작은 값을 높이로
            val h = min(it[0], it[1])

            maxWidth = max(w, maxWidth)
            maxHeight = max(h, maxHeight)
        }

        return maxWidth * maxHeight
    }
}

fun main() {
    // [60, 50], [30, 70], [60, 30], [80, 40]
    println(
        Solution().solution(
            arrayOf(
                intArrayOf(60, 50),
                intArrayOf(30, 70),
                intArrayOf(60, 30),
                intArrayOf(80, 40)
            )
        )
    )
}