package p42862

/**
 * [탐욕법 - 체육복](https://school.programmers.co.kr/learn/courses/30/lessons/42862)
 */
class Solution {
    /**
     * - 학생들 번호: 체격 순으로 매겨져 있음
     * - 바로 앞 번호 학생 또는 뒷번호 학생에게만 체육복 빌려줄 수 있음
     *      ex: 4번 학생 경우 3번 또느 5번 학생에게만 빌려줄 수 있음
     *
     *      Note: 앞/뒤 번호가 학생에게 빌릴 수 있으므로, 순서가 중요할 수 있다는 것을 염두에 둬야 합니다.
     *            특히 이때 순서에 따라 결과가 달라지는 경우가 있는지 테스트 케이스를 확인해야 할 거 같습니다.
     *
     * Example:
     * 1. 5, [2, 4], [1, 3, 5]
     *      - 1번 학생이 2번 학생에게 빌려줌
     *      - 3번 학생 또는 5번 학생이 4번 학생이게 빌려줌
     *      최대 5명이 체육 수업 들을 수 있음
     * 2. 5, [2, 4], [3]
     *      - 3번 학생이 2번 또는 4번 학생에게 빌려줌
     *      최대 4명이 체육 수업 들을 수 있음
     * 3. 3, [3], [1]
     *      - 1번 학생이 3번 학생에게 빌려줄 수 없음
     *      최대 2명이 체육 수업 들을 수 있음
     * 4. 놓쳤던 엣지 케이스: 5, [2, 4], [3, 5]
     *      정렬된 경우와 정렬하지 않은 경우 상이할 수 있습니다.
     *      - 정렬된 경우
     *        - 3번 학생이 2번 학생에게 빌려줌
     *        - 5번 학생이 3번 학생에게 빌려줌
     *        최대 5명이 체육 수업 들을 수 있음
     *
     *      - 정열하지 않은 경우
     *        - 3번 학생이 4번 학생에게 빌려줌
     *        - 5번 학생이 2번 학생에게 빌려주지 못함
     *        최대 4명이 체육 수업 들을 수 있음
     *
     *      이처럼 결과가 달라질 수가 있으므로, 도난당한 학생 집합을 정렬해서 처리할 필요가 있습니다.
     *
     * @param n 전체 학생 수.
     *  - 1 <= n <= 30
     * @param lost 도난당한 학생들의 번호.
     *  - 1 <= lost.length <= n
     *  - 중복 번호 없음
     * @param reserve 여벌 체육복 있는 학생들의 번호
     *  - 1 <= reserve.length <= n
     *  - 중복 번호 없음
     *  - 여벌 체육복 가져온 학생이 도난당했을 수 있으며, 하나만 도난당했다고 가정. 하나만 남으므로 빌려줄 수 없음.
     *
     * @return 체육 수업 들을 수 있는 학생의 최댓값
     */
    fun solution(n: Int, lost: IntArray, reserve: IntArray): Int {
        var answer = n

        // 도난당한 학생이 빌릴 수 있는지 여부로 판단
        val lostSet = lost.toMutableSet()
        val reserveSet = reserve.toMutableSet()
        val intersectSet = lostSet.intersect(reserveSet)

        // 1. 잃어버렸으면서 여벌이 있는 학생은 체육 수업 들을 수 있는 학생이므로 제외합니다.
        lostSet.removeAll(intersectSet)
        reserveSet.removeAll(intersectSet)

        // 2. 도난당한 학생은 빌릴 수 있는지 확인합니다.
        lostSet.sorted().forEach { number ->
            // # 그리디 알고리즘 특징
            // - 지역적 최적해를 통한 전체 최적해 찾습니다.
            //      각 단계에서 앞뒤 번호의 학생에게 체육복을 빌리는 것이 최선의 선택이 되고,
            //      이를 통해 전체적으로 최대한 많은 학생이 체육 수업을 들을 수 있습니다.
            // - 선택의 되돌림 없습니다. 한 번 체육복을 빌려주면 그 결정을 번복하지 않습니다.
            // - 최적해를 보장하지 않지만, 효율적인 해결책 제공합니다. 즉, 모든 경우의 수를 고려하지 않고도 빠르게 좋은 해결책을 찾을 수 있습니다.
            val prev = number - 1
            val next = number + 1

            if (reserveSet.contains(prev)) {
                reserveSet.remove(prev)
            } else if (reserveSet.contains(next)) {
                reserveSet.remove(next)
            } else {
                // 빌릴 수 없으면, 들을 수 있는 학생 수를 차감합니다.
                answer--
            }
        }

        return answer
    }
}

data class TestCase(val arg1: Int, val arg2: IntArray, val arg3: IntArray)


fun main() {

    val testCases = arrayOf(
        TestCase(
            5,
            intArrayOf(2, 4),
            intArrayOf(1, 3, 5)
        ),
        TestCase(
            5,
            intArrayOf(2, 4),
            intArrayOf(3)
        ),
        TestCase(
            3,
            intArrayOf(3),
            intArrayOf(1)
        ),
        TestCase(
            5,
            intArrayOf(2, 4),
            intArrayOf(3, 5)
        ),
    )

    for (testCase in testCases) {
        println(Solution().solution(testCase.arg1, testCase.arg2, testCase.arg3))
    }

}