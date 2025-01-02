package p42587

import java.util.*

/**
 * [프로세스](https://school.programmers.co.kr/learn/courses/30/lessons/42587?language=kotlin)
 */
class Solution {
    /**
     * 다음 규칙에 따라 운영체제가 프로세스 관리
     * 1. 실행 큐에서 대기중인 프로세스 꺼내기
     * 2. 큐에 대기중인 프로세스 중 우선순위가 더 높은 프로세스 있다면 방금 프로세스를 다시 큐에 넣기
     * 3. 만약 우선순위 더 높은 프로세스 없다면 방금 꺼낸 프로세스 실행
     *   3.1. 한 번 실행한 프로세스는 종료
     *
     * @param priorities 프로세스의 중요도가 순서대로 담긴 배열
     * - 1 <= priorities.length <= 100
     * - 1 <= priorities.i <= 9 정수
     * - 숫자가 클수록 우선 순위가 높습니다.
     * @param location 몇 번째로 실행되는지 알고 싶은 프로세스의 위치
     * - 0 <= location <= 대기 큐 프로세스 - 1
     * @return 해당 프로스세가 몇 번째로 실행되는지
     */
    fun solution(priorities: IntArray, location: Int): Int {
        // 프로세스: [A, B, C, D]
        // 위치: 2 => C
        // 우선순위: [2, 1, 3, 2]
        //
        // C, D, A, B 순으로 실행
        //
        // 프로세스: [A, B, C, D, E, F]
        // 위치: 0 => A
        // 우선순위: [1, 1, 9, 1, 1, 1]
        //
        // A pop, C가 우선순위 더 높음 => [B, C, D, E, F, A]
        // B pop, C가 우선순위 더 높음 => [C, D, E, F, A, B]
        //
        // C, D, E, F, A, B 순으로 실행
        //
        // 우선순위가 높은 것 먼저 실행해야 하므로, 우선순위 큐를 사용해서 priority 값이 높은 것부터 처리되는지 확인합니다.
        // 동시에 프로세스는 순서대로 처리되어야 하므로, 순서대로 처리하되 끝에 프로세스 추가할 수 있도록 큐를 사용합니다.

        // 우선순위 최대값 추적 위한 우선순위 큐
        // val priorityQueue = PriorityQueue<Int>(compareByDescending { it })
        val priorityQueue = PriorityQueue<Int>(Collections.reverseOrder())
        val queue = LinkedList<Pair<Int, Int>>().apply {
            priorities.forEachIndexed { idx, priority ->
                add(Pair(idx, priority))
                priorityQueue.add(priority)
            }
        }

        var processedCnt = 0

        while (queue.isNotEmpty()) {
            val currPair = queue.poll()

            // 우선 순위가 가장 높지 않으면
            if (priorityQueue.peek() != currPair.second) {
                // 큐에 다시 추가합니다.
                queue.add(currPair)
                continue
            }

            // 우선 순위가 가장 높다면 제거하고,
            priorityQueue.poll()
            // 처리됐다고 카운트합니다.
            processedCnt++

            // 현재 pair 인덱스가 찾으려는 인덱스면 반복문 종료합니다.
            if (location == currPair.first) {
                break
            }
        }

        return processedCnt
    }
}

fun main() {
    val s = Solution()

    listOf(
        Triple(intArrayOf(2, 1, 3, 2), 2, 1),
        Triple(intArrayOf(1, 1, 9, 1, 1, 1), 0, 5),
    ).forEach { tc ->
        println(s.solution(tc.first, tc.second) == tc.third)
    }
}