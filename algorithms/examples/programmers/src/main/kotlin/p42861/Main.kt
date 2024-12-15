package p42861

import java.util.PriorityQueue

/**
 * [섬 연결하기](https://school.programmers.co.kr/learn/courses/30/lessons/42861?language=kotlin)
 * 최소 신장 트리(Minimum Spanning Tree) 문제
 */
class Solution {
    /**
     * `n`개의 섬 사이에 다리 건설 위해 건설 비용 `cost`가 주어질 때,
     * 최소 비용으로 모든 섬 서로 통행 가능하도록 만들기
     *
     * 다리를 여러 번 건너더라도, 도달할 수만 있으면 통행 가능
     * - A - B - C 경우 A와 C는 서로 통행 가능
     *
     * 같은 연결은 두 번 주어지지 않고, 순서가 바뀌어도 같은 연결
     * - 0과 1 연결 시 비용 주어졌을 경우, 1과 0 연결 비용은 주어지지 않음
     *
     * 모든 섬 사이의 다리 건설 비용이 주어지지 않음. 이 경우 두 섬 사이의 건설이 불가능한 것으로 봄
     * - 즉, 직접적으로 연결되지 않은 경우가 있음.
     *
     * 연결할 수 없는 섬은 주어지지 않음
     *
     * @param n 섬의 개수
     * - 1 <= n <= 100
     * @param costs 다리 건설 비용 이차원 배열
     * - costs.length <= ((n-1) * n ) / 2. ex: n = 4 경우, 6
     * - costs[i][0]과 costs[i][1]에는 다리가 연결되는 두 섬의 번호
     * - costs[i][2]는 다리 건설 비용
     *
     * @return 모든 섬 서로 통행 가능하도록 만들 때 필요한 최소 비용
     */
    fun solution(n: Int, costs: Array<IntArray>): Int {
        var answer = 0

        // 예제:
        // [[0,1,1],[0,2,2],[1,2,5],[1,3,1],[2,3,8]]
        // 0  -1-  1
        // |     / |
        // 2   5   1
        // | /     |
        // 2  -8-  3
        //
        // [0,1,1], [0,2,2], [1,3,1] 이렇게 셋을 연결하는 게 가장 적은 비용 4
        //
        // 시작은 어디서부터든 할 수 있는 것으로 보입니다.
        // 그런데 모두 연결되어야 하므로, 시작점이 어디든 상관없습니다.
        // 0에서 시작하든, 1에서 시작하든, 2에서 시작하든, 3에서 시작하든 연결이 되어야 하고, 결국 [0,1,1], [0,2,2], [1,3,1]이 최소값이 됩니다.
        //
        // "모든 섬 서로 통행 가능"과 "연결할 수 없는 섬은 주어지지 않음" 조건.
        // 그러면 어떻게든 이어지게 데이터가 주어지므로, 다음 지점으로 이동 시 최저의 비용으로 이동하도록 하면 되지 않을까?
        // `costs`를 인접 배열로 만들어 본다면:
        // ex: [[1, 2], [0, 2, 3], [0, 1, 3], [1, 2]]
        // [
        //  0 => [Pair(1, 1), Pair(2, 2)],
        //  1 => [Pair(0, 1), Pair(2, 5), Pair(3, 1)],
        //  2 => [Pair(0, 2), Pair(1, 5), Pair(3, 8)],
        //  3 => [Pair(1, 1), Pair(2, 8)]
        // ]
        //
        // 1. 방문하지 않은 최소값만 따라간다면? X
        // 0부터 출발한다면? 0 -(1)-> 1 -(1)-> 3 -(8)-> 2 => 오답
        //
        // 2. DFS로 방문한다면? X
        // 0 -(1)-> 1 -(5)-> 2 -(8)-> 3 X
        //          1 -(1)-> 3 -(8)-> 2 X
        // 0 -(2)-> 2 -(5)-> 1 -(1)-> 3 X
        //          2 -(8)-> 3 -(1)-> 1 X
        //
        // 3. 최소값만 따라갈 때, 0 -(1)-> 1 -(1)-> 3 -(8)-> 2 에서 `3 -(8)-> 2` 가 아닌 `2 -(2) -> 0` 이 된다면 문제가 해결됩니다.
        //   a.          0 -(1)-> 1
        //   b.                     -(1)-> 3
        //   c. 2 -(2)->
        //
        //   => 2 -(2)-> 0 -(1)-> 1 -(1)-> 3
        //
        // 현재 노드와 연결된 노드중 비용이 최소인 값을 찾아갑니다.
        // 하지만 단순히 최저값을 따라가는 게 아니라, 그래프 전체 관점에서 아직 방문하지 않은 노드에 대해 비용이 최소인 간선이 있다면 우선하여 방문할 수 있어야 합니다.
        // 방문할 수 있는 간선을 큐에 쌓되, 비용이 저렴한 순서로 먼저 처리할 수 있도록 우선순위 큐를 사용합니다.
        val adjacentList = MutableList(n) { mutableSetOf<Pair<Int, Int>>() }
        costs.forEach { (from, to, cost) ->
            adjacentList[from].add(Pair(to, cost))
            adjacentList[to].add(Pair(from, cost))
        }

        // 노드별로 이동할 수 있는 노드가 비용이 낮은 순서대로 우선순위 큐에 쌓이도록 합니다.
        //
        // - 큐에 쌓이는 순서
        // ex: [[1, 2], [0, 2, 3], [0, 1, 3], [1, 2]]
        // [
        //  0 => [Pair(1, 1), Pair(2, 2)],
        //  1 => [Pair(0, 1), Pair(2, 5), Pair(3, 1)],
        //  2 => [Pair(0, 2), Pair(1, 5), Pair(3, 8)],
        //  3 => [Pair(1, 1), Pair(2, 8)]
        // ]
        // 1.
        //  visited {true, false, false, false}
        //  answer = 0
        //  Pair(1, 1)
        //  Pair(2, 2)
        // 2.
        //  visited {true, true, false, false}
        //  answer = 1
        //  Pair(3, 1)
        //  Pair(2, 2)
        //  Pair(2, 5)
        // 3.
        //  visited {true, true, false, true}
        //  answer = 2
        //  Pair(2, 2)
        //  Pair(2, 5)
        //  Pair(2, 8)
        // 5.
        //  visited {true, true, true, true}
        //  answer = 4
        //  Pair(2, 5)
        //  Pair(2, 8)
        val pq = PriorityQueue<Pair<Int, Int>>(compareBy { it.second })
        val visited = BooleanArray(n) { false }
        // 모든 노드를 방문했다면 더이상 확인할 필요 없으므로, 이를 카운트해서 반복문을 종료합니다.
        var visitedCnt = 0
        pq.add(Pair(0, 0)) // 비용 없이 0번부터 시작합니다.

        while (pq.isNotEmpty()) {
            val currPair = pq.poll()

            if (visitedCnt == n) break
            // 이미 방문한 노드면 스킵합니다.
            if (visited[currPair.first]) {
                continue
            }

            // 현재 노드를 방문 표시합니다.
            visited[currPair.first] = true
            visitedCnt++
            answer += currPair.second // 현재 비용이 현재 노드로 오는 최소 비용이므로, 바로 답에 더합니다.

            for (pair in adjacentList[currPair.first]) {
                if (!visited[pair.first]) {
                    pq.add(pair)
                }
            }
        }

        return answer
    }

    /**
     * @param graph 인접 리스트 형식의 그래프
     * - `graphs.i = List<Pair<Vertex, Weight>>`에서 `i` 정점과 연결된 다른 `Vertex`과 가중치 목록을 나타냅니다.
     *   즉, 간선 edge(i, vertex)를 나타냅니다.
     */
    fun primAlgorithm(graph: List<List<Pair<Int, Int>>>, start: Int): Pair<Int, List<Int>> {
        val n = graph.size // 정점의 수
        // 각 정점을 최소 신장 트리에 연결하는 최소 비용.
        // 즉, 현재까지 정점 i를 MST(최소 신장 트리)에 포함시키기 위한 최소 비용.
        val minimumWeightToVertex = MutableList(n) { Int.MAX_VALUE }
        val parent = MutableList(n) { -1 } // 각 정점의 부모 노들 추적하기 위한 리스트
        val visited = MutableList(n) { false } // 정점이 최소 신장 트리에 포함되었는지 여부 추적하는 리스트

        minimumWeightToVertex[start] = 0 // 시작하는 정점에 대해 자기 자신으로 이동할 필요가 없으므로 비용은 0입니다.
        var totalCost = 0 // 최소 신장 트리의 총 비용

        // ex:
        // 시작
        //  0 -(2)- 1 -(5)- 2
        //  |     /        /
        // (6)  (8)       /
        //  |  /         /
        //  3 ---------(3)
        //
        // 1. vertex: 0
        //    minimumWeightToVertex[0] = [0, 2147483647, 2147483647, 2147483647]
        //    parent: [-1, -1, -1, -1]
        //    visited: [true, false, false, false]
        //
        //    아직 방문하지 않은 인접 노드의 가중치 업데이트
        //    minimumWeightToVertex[1] = 2
        //    minimumWeightToVertex[3] = 6
        //
        // 2. vertex: 1
        //    minimumWeightToVertex: [0, 2, 2147483647, 6]
        //    parent: [-1, 0, -1, 0]
        //    visited: [true, true, false, false]
        //
        //    아직 방문하지 않은 인접 노드의 가중치 업데이트
        //    minimumWeightToVertex[2] = 5
        //    1 -> 3 이동 경우 0 -> 3 이동보다 비용이 더 크므로 업데이트하지 않습니다.
        //
        // 3. vertex: 2
        //    minimumWeightToVertex: [0, 2, 5, 6]
        //    parent: [-1, 0, 1, 0]
        //    visited: [true, true, true, false]
        //
        //    아직 방문하지 않은 인접 노드의 가중치 업데이트
        //    minimumWeightToVertex[3] = 3 // 0 -> 3 이동 비용보다 2 -> 3 이동 비용이 더 저렴하므로 업데이트합니다.
        //
        // 3. vertex: 3
        //    minimumWeightToVertex: [0, 2, 5, 6]
        //    parent: [-1, 0, 1, 2]
        //    visited: [true, true, true, true]
        //
        // 0 -(2)-> 1 -(5)-> 2 -(3)-> 3
        repeat(n) {
            // 1. 아직 방문하지 않았고, `minimumWeightToVertex[i]`가 최소인 정점 `i`를 선택합니다.
            var currVertex = -1
            for (i in 0 until n) {
                if (!visited[i] // 방문하지 않은 정점에 대해
                    && (
                            currVertex == -1 // 현재 정점이 선택되지 않았거나
                                    // 또는 현재까지 선택된 정점(`currVertex`)의 최소 비용보다 더 작은 비용을 가진 정점 `i`를 선택
                                    || minimumWeightToVertex[i] < minimumWeightToVertex[currVertex]
                            )
                ) {
                    currVertex = i
                }
            }
            println("currVertex: $currVertex, minimumWeightToVertex: $minimumWeightToVertex, parent: $parent, visited: $visited")

            if (currVertex == -1) return@repeat // 모든 정점이 처리 됐다면 반복문을 종료합니다.

            // 2. `i`를 방문했다고 표시하고 총 최소 신장 트리 비용에 간선 비용을 추가합니다.
            visited[currVertex] = true
            totalCost += minimumWeightToVertex[currVertex]

            // 3. `i` 정점의 모든 인접 정점들에 대해, 각 정점에 대한 가중치를 업데이트합니다.
            for ((adjacentVertex, weight) in graph[currVertex]) { // `i`와 `adjacentVertex` 사이의 간선의 가중치
                if (!visited[adjacentVertex] && weight < minimumWeightToVertex[adjacentVertex]) {
                    println("Update adjacentVertex: $adjacentVertex from ${minimumWeightToVertex[adjacentVertex]} to $weight")
                    // 아직 `adjacentVertex` 정점을 방문하지 않았고 새로운 간선의 가중치가 더 저렴한 경우에만 업데이트합니다.
                    minimumWeightToVertex[adjacentVertex] = weight
                    // 그리고 `i` 정점을 `vertex`에 대한 부모로 설정합니다.
                    parent[adjacentVertex] = currVertex
                }
            }
        }

        return Pair(totalCost, parent) // 총 비용과 최소 신장 트리를 나타내는 부모 배열을 리턴
    }
}

class TestCase(val n: Int, val costs: Array<IntArray>)

fun main() {
    val s = Solution()

    val cases = listOf(
        TestCase(
            4,
            arrayOf(
                intArrayOf(0, 1, 1),
                intArrayOf(0, 2, 2),
                intArrayOf(1, 2, 5),
                intArrayOf(1, 3, 1),
                intArrayOf(2, 3, 8),
            )
        ),
        // 4
        // | \
        // 5  1
        // |    \
        // 0 -1-  1
        // | \    |
        // 1  3   1
        // |    \ |
        // 2 -7-  3
        // [
        //   0 => [Pair(1, 1), Pair(2, 1), Pair(3, 3), Pair(4, 5)],
        //   1 => [Pair(0, 1), Pair(3, 1), Pair(4, 1)],
        //   2 => [Pair(0, 1), Pair(3, 7)],
        //   3 => [Pair(0, 3), Pair(1, 1), Pair(2, 7)],
        //   4 => [Pair(0, 5), Pair(1, 1)],
        // ]
        //
        // [1, 4, 1], [1, 3, 1], [0, 2, 1], [0, 1, 1]: 4
        TestCase(
            5,
            arrayOf(
                intArrayOf(0, 1, 1),
                intArrayOf(0, 2, 1),
                intArrayOf(0, 4, 5),
                intArrayOf(1, 3, 1),
                intArrayOf(1, 4, 1),
                intArrayOf(2, 3, 7),
            )
        )
    )

    cases.forEach { case ->
        println(s.solution(case.n, case.costs))
    }

    // 프림 알고리즘
    println(
        s.primAlgorithm(
            listOf(
                listOf(Pair(1, 2), Pair(3, 6)), // 0번 정점과 연결된 정점
                listOf(Pair(0, 2), Pair(2, 5), Pair(3, 8)), // 1번 정점과 연결된 정점
                listOf(Pair(1, 5), Pair(3, 3)), // 2번 정점과 연결된 정점
                listOf(Pair(0, 6), Pair(1, 8), Pair(2, 3))  // 3번 정점과 연결된 정점
            ),
            0
        )
    )
}