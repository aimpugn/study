package p42842

import kotlin.math.*

class Solution {
    /**
     * 격자 모양 카펫의 생김새는
     * - 가운데 노란색
     * - 테두리 1줄이 갈색
     * - 노란색과 갈색으로 색칠된 격자의 개수는 기억
     * - 전체 카펫의 크기는 기억하지 못함
     *
     * Example1:
     * brown: #, yellow: *
     * - brown: 10, yellow 2 -> [4, 3]
     *
     *   # # # #
     *   # * * #
     *   # # # #
     *
     *   1 * 2, 2 * 1 어떤 경우든 갈색 타일 10개로 테두리를 감쌀 수 있습니다.
     *   개인적인 편의상 가로 * 세로로 볼 때, 2 * 1 경우
     *   - 전체 높이: 1 + 1 + yellow 타일 높이 => 1 + 1 + 1
     *   - 전체 너비: yello 타일 너비 1 + 1 => 2 + 1 + 1
     *   - 전체 타일 수: (1 + 1 + yellow 타일 높이) * 2 + (yellow 타일 가로) * 2 = (1 + 1 + 1) * 2 + (2) * 2 = 6 + 4 = 10
     *
     * - brown: 8, yellow 1 -> [3, 3]
     *
     *   # # #
     *   # * #
     *   # # #
     *
     *   - 전체 높이: 1 + 1 + yello 타일 높이
     *   - 전체 너비: yello 타일 너비 + 1 + 1
     *
     * - brown: 24, yellow 24 -> [8, 6]
     *
     *   # # # # # # # #
     *   # * * * * * * #
     *   # * * * * * * #
     *   # * * * * * * #
     *   # * * * * * * #
     *   # # # # # # # #
     *
     *   가로 * 세로 경우의 수를 보면,
     *   24 * 1, 12 * 2, 8 * 3, 6 * 4 등이 가능합니다.
     *
     *   1. 24 * 1 경우 한쪽 세로에만 brown 타일 24개를 모두 소진하므로 불가능합니다.
     *   2. 12 * 2 경우 yellow 타일 12개 경우 마찬가지로 양쪽 세로에 brown 타일 24개 모두 소진하므로 불가능합니다.
     *   3. 8 * 3 경우 양쪽 세로 brown 타일 20개지만, 남은 4개로 가로 타일을 덮을 수 없습니다.
     *   4. 6 * 4 경우 양쪽 세로 brown 타일 8 + 8에 가로 4 + 4 = 24개 타일로 덮을 수 있습니다.
     *
     *   - 전체 높이: 1 + 1 + yello 타일 높이
     *   - 전체 너비: yellow 타일 너비 + 1 + 1
     *
     * @param brown 카펫에서 갈색 격자의 수
     *              8 <= brown <= 5,000
     * @param yellow 노란색 격자의 수
     *               1 <= yello <= 2,000,000
     * @return [카펫의 가로 크기, 카펫의 세로 크기]
     */
    fun solution(brown: Int, yellow: Int): IntArray {
        var answer = IntArray(2)

        // 갈색 타일은 1줄이므로, 노란색의 구조에 따라서 갈색 타일이 결정됩니다.
        // 따라서 노란색 타일의 구조를 어떻게 잡는가에 따라 가로와 세로가 결정됩니다.
        val maxYellowHeight = floor(sqrt(yellow.toFloat())).toInt()

        for (yellowHeight in 1..maxYellowHeight) {
            // brown 타일 높이 한쪽 면
            val brownHeight = 2 + yellowHeight
            // yellow 타일의 가로
            val yellowWidth = yellow / yellowHeight

            // 둘레를 모두 감쌀 수 있는지 확인합니다.
            if (((brownHeight * 2) + (yellowWidth * 2)) == brown) {
                answer[0] = yellowWidth + 2
                answer[1] = brownHeight
            }
        }

        return answer
    }
}

fun main() {
    println(Solution().solution(10, 2).contentToString())
}