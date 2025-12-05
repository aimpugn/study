package p388354

import java.util.LinkedList

class Solution {

    /**
     * 루트 노드가 설정되지 않은 1개 이상의 트리가 있습니다. 즉, 포레스트가 있습니다.
     * 모든 노드들은 서로 다른 번호를 가지고 있습니다.
     *
     * 각 노드는 다음 중 하나입니다.
     * - 홀수 노드: 노드의 번호가 홀수 && 자식 노드의 개수가 홀수
     * - 짝수 노드: 노드의 번호가 짝수 && 자식 노드의 개수가 짝수
     * - 역홀수 노드: 노드의 번호가 홀수 && 자식 노드의 개수가 짝수
     * - 역짝수 노드: 노드의 번호가 짝수 && 자식 노드의 개수가 홀수
     *
     * 0은 짝수입니다.
     *
     * 각 트리에 대해 루트 노드를 설정했을 때, '홀짝 트리'가 될 수 있는 트리의 개수와
     * '역홀짝 트리'가 될 수 있는 트리의 개수를 구하려고 합니다
     *
     * - 홀짝 트리: 홀수 노드와 짝수 노드로만 이루어진 트리입니다.
     * - 역홀짝 트리: 역홀수 노드와 역짝수 노드로만 이루어진 트리입니다.
     *
     * 2 - 3 - 4
     *      └-- 6
     * 1. 3번 노드를 루트로 설정? 모두 홀수/짝수 노드로 이뤄진 홀짝트리
     * 2. 6번 노드를 루트로 설정? 홀짝 트리, 역홀짝 트리 모두 안 됨.
     *
     * 9 - 11
     * 1. 9번 노드를 루트로 설정? 9번 노드는 홀짝 노드, 11번 노드는 역홀짝 노드가 되어서 어떤 트리도 안 됨
     * 2. 11번 노드를 루트로 설정? 11번 노드는 홀짝 노드, 9번 노드는 역홀짝 노드가 되어서 어떤 트리도 안 됨
     *
     * @param nodes 포레스트에 존재하는 노드들의 번호를 담은 1차원 정수 배열
     *              1 <= nodes.length <= 400,000
     *              1 <= nodes[i] <= 1,000,000
     *              노드의 중복은 없음.
     * @param edges 포레스트에 존재하는 간선들의 정보를 담은 2차원 정수 배열
     *              1 <= edges.length <= 1,000,000
     *              `[a, b]`는 a번 노드와 b번 노드 사이에 무방향 간선이 존재함을 의미
     *
     * @return 홀짝 트리가 될 수 있는 트리의 개수 및 역홀짝 트리가 될 수 있는 트리의 개수
     */
    fun solution(nodes: IntArray, edges: Array<IntArray>): IntArray {
        // 어떤 루트를 잡으면 홀짝 트리 또는 역홀짝 트리가 될 수 있는지 효율적으로 판단할 수 있어야 합니다.
        //
        // 우선 그래프에 대해 다시 정리를 한다면, 그래프는 정점들의 집합과 간선들의 집합으로 이루어집니다.
        // 수학적으로는 보통 G = (V, E) 라고 합니다:
        // - V: 정점들의 집합
        // - E: 간선들의 집합
        // 이때 어떤 정점 v ∈ V가 있을 때, v와 직접 연결된 간선들을 모아서 세면 노드 v의 차수(degree)라고 합니다.
        // 이 노드의 차수를 고려하여 자식의 수를 구할 수 있습니다.
        //
        // 트리 그래프는 연결된 비순환(acyclic) 그래프의 한 종류입니다.
        // 각 노드는 부모가 정확히 하나인 구조를 갖는데, 트리의 각 노드는 둘 중 하나가 됩니다:
        // - 하나의 부모와 여러 자식
        //     인접 노드 중 하나는 부모고, 나머지는 모두 자식입니다.
        //     따라서 '자식 노드의 수 = 노드의 차수 - 1'이 됩니다.
        // - 부모 없이 여러 자식
        //     루트 노드가 이에 해당하며, 인접한 노드는 모두 자식입니다.
        //     따라서 '자식 노드의 수 = 노드의 차수'가 됩니다.
        val n = nodes.size

        // 노드 번호를 0..n-1 인덱스로 맵핑하여 노드가 배열의 어디에 위치하는지 정리합니다.
        //    [11, 9, 3, 2, 4, 6]
        // => {2=3, 3=2, 4=4, 6=5, 9=1, 11=0}
        val indicesOfNodes = HashMap<Int, Int>(n * 2)
        for (i in nodes.indices) {
            indicesOfNodes[nodes[i]] = i
        }

        // 인접 리스트 구성
        // [[1], [0], [3, 5, 4], [2], [2], [2]]
        //
        // 주변 노드 개수 미리 계산
        // [1, 1, 3, 1, 1, 1]
        val adjacentNodes = Array(n) { mutableListOf<Int>() }
        val adjacentNodeCnt = IntArray(n)

        for (e in edges) {
            val node1 = indicesOfNodes[e[0]] ?: continue
            val node2 = indicesOfNodes[e[1]] ?: continue
            // 인접 리스트를 구성합니다.
            adjacentNodes[node1].add(node2)
            adjacentNodes[node2].add(node1)
            // 주변 노드 개수를 미리 정리해 둡니다.
            // 여기서 인접 노드의 개수는 '부모 + 자식'을 모두 합한 수, 즉 차수(degree)를 의미합니다.
            adjacentNodeCnt[node1]++
            adjacentNodeCnt[node2]++
        }

        // 각 노드의 번호가 홀짝인지 정리해 둡니다.
        // [1, 1, 1, 0, 0, 0]
        val oddEvenOfNodes = Array(n) { idx ->
            OddEven.of(nodes[idx])
        }

        val queue = LinkedList<Int>()
        val visited = BooleanArray(n)
        var oddEvenTrees = 0
        var reverseOddEvenTrees = 0

        // 노드 배열의 첫번째 노드부터 끝까지 순회하여 각각이 루트 노드가 되도록 합니다.
        for (idx in 0 until n) {
            if (visited[idx]) continue

            // 이어져 있는 노드들, 즉 트리(tree)를 담기 위한 리스트입니다.
            val tree = mutableListOf<Int>()
            visited[idx] = true
            queue.add(idx)

            // BFS로 트리로 구성되는 노드들을 tree 목록에 담습니다.
            while (queue.isNotEmpty()) {
                val u = queue.poll()
                tree.add(u)

                for (v in adjacentNodes[u]) {
                    if (!visited[v]) {
                        visited[v] = true
                        queue.add(v)
                    }
                }
            }

            // 홀짝 트리 및 역홀짝 트리를 다시 정리하면...
            // - 홀짝 트리: 홀수 노드와 짝수 노드로만 이루어진 트리
            //     - 루트 O: 노드 번호의 홀짝 == 주변 노드 수의 홀짝
            //     - 루트 X: 노드 번호의 홀짝 == (주변 노드 수 - 1)
            //
            //     근데 '루트 X' 경우 '주변 노드 수 -1'과 비교합니다.
            //     이때 '주변 노드 수'가 짝수면 홀수가 되고, 홀수면 짝수가 됩니다.
            //
            //     즉, 각각 '노드 번호의 홀짝 == 자식 수의 홀짝'과 '노드 번호의 홀짝 == !자식 수의 홀짝'을 하면
            //     굳이 -1을 하지 않아도 '루트 X'인 경우도 알 수 있습니다.
            //
            //     중요한 점은 "자식 수의 홀짝이 루트냐 비루트냐에 따라 달라진다는 것"입니다.
            //     '루트 O' 노드가 지정되면 자동적으로 그 아래 모든 노드는 '루트 X' 노드가 되기 때문입니다.
            //
            // - 역홀짝 트리: 역홀수 노드와 역짝수 노드로만 이루어진 트리
            //     - 루트 O: 노드 번호의 홀짝 != 주변 노드 수의 홀짝
            //     - 루트 X: 노드 번호의 홀짝 != (주변 노드 수 - 1)
            //
            // 이제 '어떤 루트 r를 잡았더니, 그 루트를 기준으로 트리가 홀짝 트리가 됐다'고 가정을 해봅니다.
            // - '루트 O'인 노드 r은 '노드 번호의 홀짝 == 주변 노드 수의 홀짝'이어야 합니다.
            //   후보 집합 S = { v | 노드 번호의 홀짝 == 주변 노드 수의 홀짝 } 안에 루트가 반드시 포함되어야 합니다.
            // - '루트 X'인 다른 노드들은 '노드 번호의 홀짝 != 주변 노드 수의 홀짝'이어야 합니다.
            //   따라서 후보 집합 S에 포함되지 않습니다.
            // - 그러면 홀짝 트리의 경우 루트 노드 r은 하나만 존재한다는 것을 알 수 있습니다.
            var oddEvenRootCandidateCnt = 0
            for (nodeIdx in tree) {
                // 이 노드 주변의 모든 노드 수의 홀/짝
                val adjacentAllNodeCntOddEven = OddEven.of(adjacentNodeCnt[nodeIdx])

                if (oddEvenOfNodes[nodeIdx] == adjacentAllNodeCntOddEven) {
                    // 홀짝 트리의 루트가 될 자격이 있음
                    oddEvenRootCandidateCnt++
                } else {
                    // 홀짝 트리의 루트가 될 자격이 없음
                }
            }

            // 역홀짝 트리는 홀짝 트리 규칙의 반대이므로,
            // 트리의 노드 수에서 홀짝 트리 루트 후보 수를 제외하면
            // 역홀짝 트리 루트 후보 수를 구할 수 있습니다.
            val reverseOddEvenRootCandidateCnt = tree.size - oddEvenRootCandidateCnt

            // 어떤 트리가 홀짝 트리라면, 임의의 루트 노드 r을 잡았을 때, 그 하위의 모든 노드는
            // '비루트 노드로서 홀짝 노드'가 되어야 합니다.
            // 즉, 홀짝 트리가 구성되면 그 홀짝 트리의 루트 후보는 하나만 존재해야 합니다.
            // 이는 역홀짝 트리도 마찬가지입니다.
            if (oddEvenRootCandidateCnt == 1) {
                oddEvenTrees++
            }
            if (reverseOddEvenRootCandidateCnt == 1) {
                reverseOddEvenTrees++
            }
        }

        return intArrayOf(oddEvenTrees, reverseOddEvenTrees)
    }

    enum class OddEven {
        ODD, EVEN;

        companion object {
            fun of(int: Int): OddEven {
                // 0: 짝수, 1: 홀수
                if ((int and 1) == 0) {
                    return EVEN
                }
                return ODD;
            }
        }
    }

    /**
     * 모든 노드들을 루트 후보로 전부 시도해 보고, 루트마다 트리 전체를 다시 BFS 하면서 자식 수를 셉니다.
     * 그래서 그런지 좋지 않은 성적을 냅니다.
     * ```
     * 테스트 1 〉	통과 (85.04ms, 82.7MB)
     * 테스트 2 〉	통과 (0.29ms, 64.6MB)
     * 테스트 3 〉	통과 (139.99ms, 105MB)
     * 테스트 4 〉	통과 (99.14ms, 84.9MB)
     * 테스트 5 〉	통과 (146.89ms, 104MB)
     * 테스트 6 〉	통과 (106.12ms, 84.5MB)
     * 테스트 7 〉	통과 (123.95ms, 105MB)
     * 테스트 8 〉	통과 (0.25ms, 63.9MB)
     * 테스트 9 〉	통과 (78.40ms, 85MB)
     * 테스트 10 〉	통과 (125.60ms, 105MB)
     * 테스트 11 〉	통과 (95.08ms, 84.9MB)
     * 테스트 12 〉	통과 (136.86ms, 107MB)
     * 테스트 13 〉	통과 (4.44ms, 64.4MB)
     * 테스트 14 〉	통과 (4.31ms, 64.4MB)
     * 테스트 15 〉	통과 (4.00ms, 63.8MB)
     * 테스트 16 〉	통과 (4.45ms, 62.8MB)
     * 테스트 17 〉	통과 (4.59ms, 64.1MB)
     * 테스트 18 〉	통과 (3.24ms, 64.4MB)
     * 테스트 19 〉	통과 (7.94ms, 64.5MB)
     * 테스트 20 〉	통과 (3.48ms, 64.2MB)
     * 테스트 21 〉	통과 (8.28ms, 63.7MB)
     * 테스트 22 〉	통과 (3.15ms, 63.8MB)
     * 테스트 23 〉	통과 (3.75ms, 63.4MB)
     * 테스트 24 〉	실패 (시간 초과)
     * 테스트 25 〉	통과 (0.25ms, 62.3MB)
     * 테스트 26 〉	실패 (시간 초과)
     * 테스트 27 〉	실패 (시간 초과)
     * 테스트 28 〉	실패 (시간 초과)
     * 테스트 29 〉	실패 (시간 초과)
     * 테스트 30 〉	통과 (0.31ms, 64.7MB)
     * 테스트 31 〉	실패 (시간 초과)
     * 테스트 32 〉	통과 (0.20ms, 64.7MB)
     * 테스트 33 〉	실패 (시간 초과)
     * 테스트 34 〉	실패 (시간 초과)
     * 테스트 35 〉	실패 (시간 초과)
     * 테스트 36 〉	실패 (시간 초과)
     * 테스트 37 〉	실패 (시간 초과)
     * 테스트 38 〉	실패 (시간 초과)
     * 테스트 39 〉	실패 (시간 초과)
     * 테스트 40 〉	실패 (시간 초과)
     * 테스트 41 〉	실패 (시간 초과)
     * 테스트 42 〉	실패 (시간 초과)
     * 테스트 43 〉	실패 (시간 초과)
     * 테스트 44 〉	실패 (시간 초과)
     * 테스트 45 〉	실패 (시간 초과)
     * 테스트 46 〉	실패 (시간 초과)
     * 테스트 47 〉	실패 (시간 초과)
     * 테스트 48 〉	실패 (시간 초과)
     * 테스트 49 〉	실패 (시간 초과)
     * 테스트 50 〉	실패 (시간 초과)
     * 테스트 51 〉	실패 (시간 초과)
     * 테스트 52 〉	실패 (시간 초과)
     * 테스트 53 〉	실패 (시간 초과)
     * 테스트 54 〉	실패 (시간 초과)
     * 채점 결과
     * 정확성: 20.0
     * 합계: 20.0 / 100.0
     * ```
     *
     * 노드 개수는 최대 40만 개, 간선 개수는 최대 100만 개까지 올 수 있습니다.
     * 일반적인 BFS는 한 번 도는 데 O(N + E)가 걸리고, 숫자로 치면 대략 1,400,000 정도의 연산입니다.
     *
     * 이런 BFS를 노드 수(최대 400,000번)만큼 반복하게 되면, O(N * (N + E))가 되어 O(400,000 × 1,400,000),
     * 즉 10^11 이상이 됩니다.
     *
     * 각 트리마다 루트를 적절히 잡아 홀짝 트리(또는 역홀짝 트리)가 될 수 있는지를 보고, 트리의 개수를 세야 합니다.
     * 즉, 트리마다 최대 1번씩만 세야 하고, 똑같은 트리를 여러 번 세면 안 됩니다.
     */
    fun solutionFailed(nodes: IntArray, edges: Array<IntArray>): IntArray {
        // 간선을 맵으로 정리
        // [[9, 11], [2, 3], [6, 3], [3, 4]]
        // {
        //   "2": [3]
        //   "3": [2, 4, 6]
        //   "4": [3]
        //   "6": [3]
        //   "9": [11]
        //   "11": [9]
        // }
        var edgeMap = mutableMapOf<Int, MutableList<Int>>()

        edges.forEach { edge ->
            edgeMap.getOrPut(edge[0]) { mutableListOf() }.add(edge[1])
            edgeMap.getOrPut(edge[1]) { mutableListOf() }.add(edge[0])
        }

        var oddEvenTree = 0;
        var reverseOddEvenTree = 0;

        // nodes 배열을 순회할 때 현재 노드를 루트라고 봅니다.
        // [11, 9, 3, 2, 4, 6]
        nodes.forEach { rootNode ->

            // 이떄부터 탐색을 하면서 홀짝 트리인지, 역홀짝 트리인지 판단합니다.

            // 루트가 바뀔 때만 초기화되어야 하는, 방문했는지 여부 체크하는 visited
            var visited = mutableMapOf<Int, Boolean>()

            var queue = LinkedList<Int>()
            queue.add(rootNode)

            var oddEven = 0;
            var reverseOddEven = 0;

            while (!queue.isEmpty()) {
                val currNode = queue.poll()

                if (visited[currNode] == true) {
                    continue
                }

                visited[currNode] = true

                // 자식 노드 개수 확인
                val children = edgeMap[currNode] ?: emptyList()
                val childrenCnt = children.count { el -> visited[el] != true }

                // println("rootNode: $rootNode, currNode: $currNode, children: $children")
                if (isOddEvenTree(currNode, childrenCnt)) {
                    // println("\tisOddEvenTree($currNode, $childrenCnt)")
                    oddEven++;
                } else if (isReverseOddEvenTree(currNode, childrenCnt)) {
                    // println("\tisReverseOddEvenTree($currNode, $childrenCnt)")
                    reverseOddEven++;
                }

                if (children.isNotEmpty()) {
                    queue.addAll(children)
                }
            }

            // println("rootNode: $rootNode, oddEven: $oddEven, reverseOddEven: $reverseOddEven")
            /**
             * 홀짝들만 존재하고 역홀짝은 존재하지 않을 때 홀짝트리
             */
            if (0 < oddEven && reverseOddEven == 0) {
                oddEvenTree++;
            }
            /**
             * 역홀짝들만 존재하고 홀짝은 존재하지 않을 때 역홀짝 트리
             */
            if (0 < reverseOddEven && oddEven == 0) {
                reverseOddEvenTree++;
            }

            // 현재 루트 노드의 방문 여부를 초기화합니다.
            visited[rootNode] = false
        }

        return intArrayOf(oddEvenTree, reverseOddEvenTree)
    }

    /**
     * 홀짝 트리(홀수 노드와 짝수 노드로만 이루어진 트리) 여부를 판단합니다.
     *
     * 각 노드는 다음 중 하나입니다.
     * - 홀수 노드: 노드의 번호가 홀수 && 자식 노드의 개수가 홀수
     * - 짝수 노드: 노드의 번호가 짝수 && 자식 노드의 개수가 짝수
     * - 역홀수 노드: 노드의 번호가 홀수 && 자식 노드의 개수가 짝수
     * - 역짝수 노드: 노드의 번호가 짝수 && 자식 노드의 개수가 홀수
     */
    fun isOddEvenTree(nodeNumber: Int, childrenCnt: Int): Boolean {
        return (isEven(nodeNumber) && isEven(childrenCnt))
                || (isOdd(nodeNumber) && isOdd(childrenCnt))
    }

    /**
     * 역홀짝 트리(역홀수 노드와 역짝수 노드로만 이루어진 트리) 여부를 판단합니다.
     *
     * 각 노드는 다음 중 하나입니다.
     * - 홀수 노드: 노드의 번호가 홀수 && 자식 노드의 개수가 홀수
     * - 짝수 노드: 노드의 번호가 짝수 && 자식 노드의 개수가 짝수
     * - 역홀수 노드: 노드의 번호가 홀수 && 자식 노드의 개수가 짝수
     * - 역짝수 노드: 노드의 번호가 짝수 && 자식 노드의 개수가 홀수
     */
    fun isReverseOddEvenTree(nodeNumber: Int, childrenCnt: Int): Boolean {
        return (isOdd(nodeNumber) && isEven(childrenCnt))
                || (isEven(nodeNumber) && isOdd(childrenCnt))
    }

    fun isEven(target: Int): Boolean {
        return target == 0 || target % 2 == 0
    }

    fun isOdd(target: Int): Boolean {
        return !isEven(target)
    }
}

fun main() {
    val s = Solution()

    listOf(
        Pair(
            intArrayOf(11, 9, 3, 2, 4, 6),
            arrayOf(
                // [[9, 11], [2, 3], [6, 3], [3, 4]]
                intArrayOf(9, 11),
                intArrayOf(2, 3),
                intArrayOf(6, 3),
                intArrayOf(3, 4),
            )
        ),
        Pair(
            intArrayOf(9, 15, 14, 7, 6, 1, 2, 4, 5, 11, 8, 10),
            arrayOf(
                // [[5, 14], [1, 4], [9, 11], [2, 15], [2, 5], [9, 7], [8, 1], [6, 4]]
                intArrayOf(5, 14),
                intArrayOf(1, 4),
                intArrayOf(9, 11),
                intArrayOf(2, 15),
                intArrayOf(2, 5),
                intArrayOf(9, 7),
                intArrayOf(8, 1),
                intArrayOf(6, 4),
            )
        )
    ).forEach { testCase ->
        println(s.solution(testCase.first, testCase.second).contentToString())
    }
}