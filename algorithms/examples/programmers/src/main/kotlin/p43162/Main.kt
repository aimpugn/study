package p43162

import java.util.LinkedList

/**
 * [네트워크](https://school.programmers.co.kr/learn/courses/30/lessons/43162?language=kotlin)
 */
class Solution {
    /**
     * - A - B 직접 연결
     * - B - C 직접 연결
     * - A - C 간접 연결
     * - A, B, C 모두 같은 네트워크 상에 존재
     *
     * @param n 컴퓨터의 수
     * - 1 <= n <= 200 자연수
     * - 각 컴퓨터는 0 ~ n-1 정수로 표현
     * @param computers 연결에 대한 정보 담긴 2차원 배열
     * - computers.i.j: i 컴퓨터와 j 컴퓨터가 연결
     * - computers.i.i: 항상 1
     * @return 네트워크의 개수를 리턴
     */
    fun solution(n: Int, computers: Array<IntArray>): Int {
        var answer = 0
        // [[1, 1, 0], [1, 1, 0], [0, 0, 1]]
        // 0 - 0
        // 0 - 1
        // 0   2
        // 1   2
        // => 0 - 1 / 2 => 총 2개
        //
        // [
        //  0 => [1],
        //  1 => [0],
        //  2 => []
        // ]
        // 0 -> 1 가보고 종료. 그 다음에 2부터 시작할 수 있어야 합니다.
        //
        // [[1, 1, 0], [1, 1, 1], [0, 1, 1]]
        // 0 - 0
        // 0 - 1
        // 0   2
        // 1 - 2
        // => 0 - 1 - 2 => 총 1개
        //
        // 1. 시작점을 잡고, 해당 시작점에서 어디까지 도달할 수 있는지 체크하고, 그 다음 아직 방문하지 않은 점을 다시 시작점으로 잡기
        // 각 컴퓨터가 도달할 수 있는 부분을 그래프로 정리합니다.
        // 네트워크가 여럿임을 셀 수 있어야 하므로, 새로운 네트워크가 시작되는 것을 알 수 있어야 합니다.
        val graph = List(n) { mutableSetOf<Int>() }

        // 각 컴퓨터별로 방문 가능한 컴퓨터 목록만 남깁니다.
        // 2차원 배열을 그대로 사용해도 되지만, 명확하게 방문할 노드 목록만 남도록 정리합니다.
        computers.forEachIndexed { idx, computerRelation ->
            for (i in 0 until computerRelation.size) {
                if (idx != i && computerRelation[i] == 1) {
                    graph[idx].add(i)
                }
            }
        }

        // 로직:
        // - 시작점 대상 컴퓨터 목록 loop
        //   => 해당 점이 갈 수 있는 곳들 방문
        //      => 방문한 곳 기록하고 queue에 추가
        val visited = BooleanArray(n) { false }
        val queue = LinkedList<Int>()

        visited.indices.forEach { idx ->
            if (visited[idx]) return@forEach
            queue.add(idx)
            visited[idx] = true

            while (queue.isNotEmpty()) {
                val currComputer = queue.poll()

                // 방문하지 않았다면, 방문 처리합니다.
                graph[currComputer].forEach { reachableComputer ->
                    if (!visited[reachableComputer]) {
                        visited[reachableComputer] = true
                        // 해당 컴퓨터를 방문하여 또 도달할 수 있는 컴퓨터가 있는지 확인합니다.
                        queue.add(reachableComputer)
                    }
                }
            }

            // 현재 컴퓨터 기준으로 도달할 수 있는 곳까지 모두 도달했으므로, 네트워크 하나를 구성했다고 보고 +1 합니다.
            answer++
        }

        return answer
    }
}

fun main() {
    val s = Solution()
    listOf(
        Pair(
            3,
            arrayOf(
                intArrayOf(1, 1, 0),
                intArrayOf(1, 1, 0),
                intArrayOf(0, 0, 1),
            )
        ),
        Pair(
            3,
            arrayOf(
                intArrayOf(1, 1, 0),
                intArrayOf(1, 1, 1),
                intArrayOf(0, 1, 1),
            )
        ),
    ).forEach {
        println(s.solution(it.first, it.second))
    }


}