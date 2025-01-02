package p86971

import java.util.LinkedList
import kotlin.math.*

class Solution {
    /**
     * - n개의 송전탑 -> 전선 통해 트리 형태로 연결
     *   트리 구조는 순환이 없는 연결 그래프
     * - 전선들 중 하나 끊어서 전력망 네트워크를 2개로 분할하는 게 목표
     *      "전선들 중 하나 끊어서"이므로, 모든 전선을 대상으로 시도해야 하므로,
     *      모든 간선에 대해 반복적으로 제거하고 시뮬레이션해야 함을 예상
     * - 두 전력망의 송전탑의 개수 최대한 비슷하게 맞추기
     *
     * Example:
     * 1. [[1,3],[2,3],[3,4],[4,5],[4,6],[4,7],[7,8],[7,9]]
     *
     *      1   5     8
     *      |   |     |
     *      3 - 4 -x- 7
     *      |   |     |
     *      2   6     9
     *
     *      '3과 4' 사이 또는 '4와 7' 사이 전선 끊으면 6개와 3개의 송전탑을 갖습니다.
     *
     *      1 -> 3
     *      2 -> 3
     *      3 -> 1, 2, 4
     *      4 -> 3, 5, 6, 7
     *      5 -> 4
     *      6 -> 4
     *      7 -> 4, 8, 9
     *      8 -> 7
     *      9 -> 7
     *
     * 2. [[1,2],[2,3],[3,4]]
     *
     *      1 - 2 -x- 3 - 4
     *
     *      '2와 3' 사이 전선 끊으면 두 전력망이 모두 2개의 송전탑을 갖습니다.
     *
     *      1 -> 2
     *      2 -> 1, 3
     *      3 -> 2, 4
     *      4 -> 3
     *
     * 3. [[1,2],[2,7],[3,7],[3,4],[4,5],[6,7]]
     *
     *      1 - 2 - 7 - 6
     *              |
     *              x
     *              |
     *              3 - 4 - 5
     *
     *      '7과 3' 사이 전선을 끊으면 각각 3개, 4개의 송전탑을 갖습니다.
     *
     *      1 -> 2
     *      2 -> 1, 7
     *      3 -> 4, 7
     *      4 -> 3, 5
     *      5 -> 4
     *      6 -> 7
     *      7 -> 2, 3, 6
     *
     *  핵심은?
     *  - 네트워크를 두 부분으로 나누는 것
     *  - 두 부분의 크기 차이를 최소화하는 것
     *
     * @param n 송전탑의 개수
     *  - 2 <= n <= 100 자연수
     * @param wires 전선 정보. 전력망 네트워크는 하나의 트리 형태.
     *  - wires.length = n - 1
     *  - [(v1, v2), ...] 2차원 배열. v1 송전탑과 v2 송전탑이 연결되어 있음을 의미.
     *  - 1 <= v1 < v2 <= n
     * @return 송전탑 개수가 가능한 비슷하게 두 전력망으로 나누었을 때, 두 전력망이 갖는 송전탑 개수의 차이(절대값)
     */
    fun solution(n: Int, wires: Array<IntArray>): Int {
        var answer: Int = Int.MAX_VALUE

        // 가장 직관적인 것은 중간 부분을 자른다는 것인데, 그러면 "어떻게" 중간 부분을 잘라야 할지 고민이 됩니다.
        // (X) 1. n의 경우의 수들을 탐색? 가령
        //      n = 4인 경우 (1, 3), (2, 2) 경우가 있습니다. 그리고 최적의 경우부터 찾아봅니다.
        //      하지만, 실제 네트워크 구조가 정말 그렇게 나뉠 수 있는지 판단해야 하는데... 쉽지 않아 보입니다.
        //
        // (X) 2. 하나씩 잘라보며 두 개의 전력망으로 나뉘는지 확인?
        //      실제 네트워크 구조에서 잘라내기 때문에 정말 그런지 별도로 판단할 필요가 없습니다.
        //
        // (X) 3.각 정점별 연결된 리스트 또는 집합으로 정리하고, 가장 많은 노드가 연결된 지점들을 찾습니다.
        //      그리고 상/하/좌/우로 전선을 잘라보며 나뉘는 전력망 개수를 세어 봅니다.
        //
        //      Example1의 4 -> 3, 5, 6, 7에서 3을 지우고, 3 -> 1, 2, 4에서 4를 지우면
        //      4 -> 5, 6, 7이 되고, 3 -> 1, 2가 됩니다. 이때
        //      1. 4를 기점으로 다시 노드를 방문하며 개수를 세고,
        //      2. 3을 기점으로 다시 노드를 방무하여 개수를 세어서
        //      3. 기록하고
        //      4. 나머지 5, 6, 7 등에 대해서도 1번부터 반복합니다.
        //
        //      그리고 모든 노드를 체크할 때까지 반복합니다.
        //
        // 4. 3번은 오히려 모든 경우의 수를 고려하기에 복잡합니다.
        //      따라서 그냥 graph만 배열 형태로 정리하고, 모든 wires를 고려하도록 풀이합니다.

        val nodes = Array<MutableSet<Int>>(n + 1) {
            mutableSetOf()
        }

        // 1. 각 노드별로 연결된 노드를 집합으로 들고 있는 자료 구조로 정리합니다.
        wires.forEach { (v1, v2) ->
            nodes[v1].add(v2)
            nodes[v2].add(v1)
        }
        // println("arranged nodes: ${nodes.contentToString()}")

        // 2. 각 간선을 하나씩 제거하면서 시뮬레이션합니다.
        for ((v1, v2) in wires) {
            // 인접 리스트의 깊은 복사본을 생성합니다.
            val tempGraph = nodes.map { it.toMutableSet() }.toTypedArray()

            // 2.1. 간선을 제거하여 트리를 분할합니다.
            tempGraph[v1].remove(v2)
            tempGraph[v2].remove(v1)

            // 2.2.
            // 이제 서로 연관 관계가 사라졌고, 전선 하나가 사라짐으로써 전력망은 두 개로 나뉘었을 겁니다.
            // v1을 시작점으로 삼아서 도달할 수 있는 노드를 카운트합니다.
            val network1 = countByBFS(n, v1, tempGraph)
            // 다른 전력망은, 전체에서 분할 된 한 쪽 전력망을 차감하면 얻을 수 있습니다.
            val network2 = n - network1

            // 차이가 가능한 비슷하게 두 전력망 나눴을 때, 송전탑 개수의 차이(절대값)을 구합니다.
            // 두 값을 뺀 차이가 적을수록 두 전력망이 비슷하게 나뉘었다고 볼 수 있습니다.
            answer = min(answer, abs(network1 - network2))
        }

        return answer
    }

    private fun countByBFS(n: Int, from: Int, nodes: Array<MutableSet<Int>>): Int {
        val queue = LinkedList<Int>()
        // 0 아닌 1부터 노드가 시작하므로, 편의상 n + 1개의 Boolean 배열로 초기화합니다.
        // 1 <= v1 < v2 <= n 이므로, 송전탑이 인덱스를 벗어날 일은 없습니다.
        val visited = BooleanArray(n + 1) { false }
        // from부터 시작하니 1부터 카운트하고 방문 여부를 기록합니다.
        var count = 1
        visited[from] = true
        queue.add(from)

        while (queue.isNotEmpty()) {
            val currNode = queue.pop()
            nodes[currNode].forEach { linkedNode ->
                if (!visited[linkedNode]) {
                    visited[linkedNode] = true
                    queue.add(linkedNode)
                    count++
                }
            }
        }

        return count
    }
}

// 데이터를 담을 TestCase 데이터 클래스 정의
data class TestCase(val arg1: Int, val arg2: Array<IntArray>)

fun main() {

    val testCases = listOf(
        TestCase(
            9,
            arrayOf(
                intArrayOf(1, 3),
                intArrayOf(2, 3),
                intArrayOf(3, 4),
                intArrayOf(4, 5),
                intArrayOf(4, 6),
                intArrayOf(4, 7),
                intArrayOf(7, 8),
                intArrayOf(7, 9)
            )
        ),
        TestCase(
            4,
            arrayOf(
                intArrayOf(1, 2),
                intArrayOf(2, 3),
                intArrayOf(3, 4),
            )
        ),
        TestCase(
            7,
            arrayOf(
                intArrayOf(1, 2),
                intArrayOf(2, 7),
                intArrayOf(3, 7),
                intArrayOf(3, 4),
                intArrayOf(4, 5),
                intArrayOf(6, 7),
            )
        )
    )

    for (testCase in testCases) {
        println(
            Solution().solution(testCase.arg1, testCase.arg2)
        )
    }
}