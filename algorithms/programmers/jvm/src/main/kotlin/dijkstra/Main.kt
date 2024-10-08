package dijkstra

import java.util.*

/**
 *
 * 다익스트라(Dijkstra) 알고리즘은 가중치 있는 방향 그래프에서 한 정점으로부터 다른 모든 정점까지의 최단 경로를 찾는 알고리즘입니다.
 *
 * 그래프 구조 예시:
 *   (1) --4-- (2)
 *    | \     / |
 *    2   1  1  5
 *    |     \   |
 *   (3) --3-- (4)
 *
 * 알고리즘 개요:
 * 1. 모든 정점의 거리를 무한대로 초기화하고, 시작 정점의 거리를 0으로 설정
 * 2. 미방문 정점 중 가장 가까운 정점을 선택 (우선순위 큐 등 사용)
 * 3. 선택된 정점의 이웃들에 대해 더 짧은 경로가 있는지 확인하고 거리 업데이트
 * 4. 모든 정점을 방문할 때까지 2-3 반복
 *
 * 시간 복잡도: O(E + V log V), 여기서 E는 간선의 수, V는 정점의 수
 * 공간 복잡도: O(V)
 */


/**
 * 그래프를 표현하는 클래스
 * @param vertices 그래프의 정점 집합
 * @param edges 그래프의 간선 집합 (시작 정점, 도착 정점, 가중치)
 */
class Graph(val vertices: Set<Int>, val edges: Set<Triple<Int, Int, Int>>) {
    // 정점 u에서 갈 수 있는 인접 정점들의 집합을 반환하는 함수
    fun getAdjacentVertices(u: Int): Set<Int> = edges.filter { it.first == u }.map { it.second }.toSet()

    // 정점 u에서 v로 가는 간선의 가중치를 반환하는 함수
    fun getWeight(u: Int, v: Int): Int = edges.find { it.first == u && it.second == v }?.third ?: Int.MAX_VALUE
}

/**
 * 다익스트라 알고리즘 구현 함수
 * @param graph 주어진 그래프
 * @param start 시작 정점
 * @return 각 정점까지의 최단 거리와 이전 정점을 담은 Pair
 */
fun dijkstra(graph: Graph, start: Int): Pair<Map<Int, Int>, Map<Int, Int>> {
    // 방문한 정점들의 집합
    val visited = mutableSetOf<Int>()

    // 각 정점까지의 최단 거리를 저장하는 맵
    val distances = graph.vertices.associateWith { Int.MAX_VALUE }.toMutableMap()
    distances[start] = 0  // 시작 정점의 거리는 0으로 초기화

    // 각 정점의 이전 정점을 저장하는 맵 (경로 추적용)
    val previousVertices = mutableMapOf<Int, Int>()

    // 모든 정점을 방문할 때까지 반복
    while (visited.size < graph.vertices.size) {
        // 미방문 정점 중 거리가 최소인 정점 선택
        val current = graph.vertices.minus(visited).minByOrNull { distances[it] ?: Int.MAX_VALUE }
            ?: break  // 모든 정점을 방문했거나 연결되지 않은 정점이 있는 경우

        // 선택된 정점을 방문 처리
        visited.add(current)

        // 현재 정점의 인접 정점들에 대해 거리 갱신
        for (neighbor in graph.getAdjacentVertices(current)) {
            if (neighbor !in visited) {
                val newDistance = distances[current]!! + graph.getWeight(current, neighbor)
                if (newDistance < distances[neighbor]!!) {
                    distances[neighbor] = newDistance
                    previousVertices[neighbor] = current
                }
            }
        }

        // 현재 상태 시각화
        visualizeState(graph, visited, distances, current)
    }

    return Pair(distances, previousVertices)
}

/**
 * 알고리즘의 현재 상태를 ASCII 다이어그램으로 시각화하는 함수
 */
fun visualizeState(graph: Graph, visited: Set<Int>, distances: Map<Int, Int>, current: Int) {
    println("\n현재 상태:")
    println("방문한 정점: $visited")
    println("현재 정점: $current")
    println("거리:")
    for (v in graph.vertices) {
        val distance = distances[v]?.toString() ?: "∞"
        println("  $v: $distance")
    }
    println("\n그래프 시각화:")
    // 여기에 그래프의 ASCII 다이어그램을 추가할 수 있습니다.
    // 예를 들어:
    //    (1)--4--(2)
    //     | \     |
    //     2   1   5
    //     |     \ |
    //    (3)--3--(4)
}


/**
 * 다익스트라 알고리즘을 구현한 클래스
 *
 * @param graph 그래프를 나타내는 맵. 키는 정점, 값은 (이웃 정점, 가중치) 쌍의 맵
 */
class Dijkstra(private val graph: Map<Int, Map<Int, Int>>) {

    /**
     * 주어진 시작 정점으로부터 모든 다른 정점까지의 최단 거리를 계산
     *
     * @param start 시작 정점
     * @return 각 정점까지의 최단 거리를 담은 맵
     */
    fun shortestPath(start: Int): Map<Int, Int> {
        // 각 정점까지의 최단 거리를 저장할 맵
        // 초기에는 모든 거리를 무한대로 설정
        val distances = graph.keys.associateWith { Int.MAX_VALUE }.toMutableMap()
        // 시작 정점의 거리는 0으로 설정
        distances[start] = 0

        // 미방문 정점 중 가장 가까운 정점을 효율적으로 선택하기 위한 우선순위 큐
        // Pair의 첫 번째 요소는 거리, 두 번째 요소는 정점 번호
        val queue = PriorityQueue<Pair<Int, Int>>(compareBy { it.first })
        queue.offer(Pair(0, start))

        // 각 정점의 이전 정점을 저장할 맵 (경로 추적용)
        val previous = mutableMapOf<Int, Int>()

        // 큐가 비어있지 않은 동안 계속 실행
        while (queue.isNotEmpty()) {
            // 가장 가까운 미방문 정점을 선택
            val (distance, current) = queue.poll()

            // 이미 더 짧은 경로를 찾은 경우 스킵
            if (distance > distances[current]!!) continue

            // 현재 정점의 모든 이웃을 확인
            graph[current]?.forEach { (neighbor, weight) ->
                val newDistance = distance + weight
                // 더 짧은 경로를 발견한 경우, 거리를 업데이트하고 큐에 추가
                if (newDistance < distances[neighbor]!!) {
                    distances[neighbor] = newDistance
                    previous[neighbor] = current
                    queue.offer(Pair(newDistance, neighbor))
                }
            }
        }

        // 결과 출력
        distances.forEach { (vertex, distance) ->
            println("정점 $vertex 까지의 최단 거리: $distance")
            printPath(start, vertex, previous)
        }

        return distances
    }
}

fun main() {
    // 그래프 정의
    val vertices = setOf(1, 2, 3, 4)
    val edges = setOf(
        Triple(1, 2, 4), Triple(1, 3, 2), Triple(1, 4, 1),
        Triple(2, 4, 5),
        Triple(3, 4, 3)
    )
    val graph = Graph(vertices, edges)

    // 알고리즘 실행
    val (distances, previousVertices) = dijkstra(graph, 1)

    // 결과 출력
    println("\n최종 결과:")
    distances.forEach { (vertex, distance) ->
        println("정점 $vertex 까지의 최단 거리: $distance")
        printPath(1, vertex, previousVertices)
    }
}

/**
 * 시작 정점부터 목표 정점까지의 경로를 출력하는 함수
 */
fun printPath(start: Int, end: Int, previousVertices: Map<Int, Int>) {
    val path = mutableListOf<Int>()
    var current = end
    while (current != start) {
        path.add(current)
        current = previousVertices[current] ?: break
    }
    path.add(start)
    println("경로: ${path.reversed().joinToString(" -> ")}")
}