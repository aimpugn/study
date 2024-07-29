# Dijkstra

- [Dijkstra](#dijkstra)
    - [다익스트라 알고리즘](#다익스트라-알고리즘)
    - [기본 원리](#기본-원리)
    - [비교: 배낭 문제 예제 (동적 프로그래밍)](#비교-배낭-문제-예제-동적-프로그래밍)

## 다익스트라 알고리즘

다익스트라 알고리즘은 그래프에서 한 정점에서 다른 모든 정점으로의 최단 경로를 찾는 알고리즘입니다.
이 알고리즘은 **가중치가 있는 그래프에서 동작하며, 음의 가중치를 허용하지 않습니다**.

다익스트라 알고리즘은 그래프에서 한 정점에서 다른 모든 정점으로의 최단 경로를 찾는 데 사용됩니다.
이는 *최적의 경로를 찾는 문제*로, 가중치가 있는 그래프에서 최단 경로를 계산하는 데 매우 유용합니다.

## 기본 원리

다익스트라 알고리즘의 기본 원리는 다음과 같습니다:

1. 시작 정점을 설정하고, 그 정점의 거리를 0으로 설정합니다.
2. 나머지 모든 정점의 거리를 무한대로 설정합니다.
3. 방문하지 않은 정점 중 가장 거리가 짧은 정점을 선택합니다.
4. 해당 정점을 통해 다른 정점으로 가는 경로를 계산하고, 더 짧은 경로가 발견되면 거리를 갱신합니다.
5. 모든 정점을 방문할 때까지 3번과 4번을 반복합니다.

다음과 같은 그래프를 예제로 사용하겠습니다:

```plaintext
A --1--> B --2--> C
|      / |        |
4     3  1        5
|    /   |        |
D --2--> E --1--> F
```

다음은 Go 언어로 다익스트라 알고리즘을 구현한 예제입니다:

```go
package main

import (
    "container/heap"
    "fmt"
    "math"
)

// Edge 구조체는 그래프의 간선을 나타냅니다.
// 'to'는 목적지 정점을, 'weight'는 간선의 가중치를 나타냅니다.
type Edge struct {
    to     int    // 목적지 정점의 인덱스
    weight int    // 간선의 가중치
}

// Graph 구조체는 인접 리스트로 그래프를 나타냅니다.
// 'vertices'는 정점의 개수를, 'edges'는 각 정점의 간선 리스트를 나타냅니다.
type Graph struct {
    vertices int    // 그래프 내의 정점 수
    edges    [][]Edge    // 각 정점에 연결된 간선 리스트
}

// NewGraph 함수는 주어진 정점 수로 새로운 그래프를 생성합니다.
// 그래프를 초기화하고 빈 간선 리스트를 생성합니다.
func NewGraph(vertices int) *Graph {
    return &Graph{
        vertices: vertices,    // 정점 수 설정
        edges:    make([][]Edge, vertices),    // 각 정점에 대한 간선 리스트를 초기화
    }
}

// AddEdge 함수는 그래프에 간선을 추가합니다.
// 'from' 정점에서 'to' 정점으로 가는 가중치 'weight'의 간선을 추가합니다.
func (g *Graph) AddEdge(from, to, weight int) {
    g.edges[from] = append(g.edges[from], Edge{to, weight})    // from 정점에서 to 정점으로 가는 간선을 추가
}

// Item 구조체는 우선순위 큐의 항목을 나타냅니다.
// 'vertex'는 정점을, 'dist'는 현재까지의 최단 거리를 나타냅니다.
type Item struct {
    vertex int    // 정점 인덱스
    dist   int    // 최단 거리
    index  int    // 우선순위 큐 내의 인덱스
}

// PriorityQueue는 우선순위 큐를 구현합니다.
// Go의 heap 패키지를 사용하여 구현합니다.
// 우선순위 큐를 선택한 이유
//
// 다익스트라 알고리즘에서는 최단 거리를 찾는 과정에서
// 각 정점까지의 최단 거리를 효율적으로 갱신하고 관리해야 합니다.
// 이때 우선순위 큐를 사용하면 최소 힙을 통해 `O(log N)`의 시간복잡도로
// 최단 거리를 갱신할 수 있습니다. 이는 간단한 배열을 사용하는 경우
// `O(N)`의 시간복잡도가 소요되는 것에 비해 효율적입니다.
// 따라서 다익스트라 알고리즘의 성능을 최적화하기 위해 우선순위 큐를 사용합니다.
type PriorityQueue []*Item    // Item 포인터를 요소로 가지는 슬라이스 정의

// Len 메서드는 우선순위 큐의 길이를 반환합니다.
func (pq PriorityQueue) Len() int { return len(pq) }

// Less 메서드는 두 항목의 우선순위를 비교합니다.
// dist 값이 작은 항목이 더 높은 우선순위를 갖습니다.
func (pq PriorityQueue) Less(i, j int) bool {
    return pq[i].dist < pq[j].dist
}

// Swap 메서드는 두 항목의 위치를 교환합니다.
func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}

// Push 메서드는 우선순위 큐에 새로운 항목을 추가합니다.
func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    item := x.(*Item)
    item.index = n
    *pq = append(*pq, item)
}

// Pop 메서드는 우선순위 큐에서 항목을 제거하고 반환합니다.
func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1] // 마지막 항목을 꺼냅니다.
    old[n-1] = nil   // 메모리 누수를 방지하기 위해 nil로 설정
    *pq = old[0 : n-1] // 마지막 항목을 제거합니다.
    return item
}

// update 메서드는 우선순위 큐에서 항목의 거리를 갱신하고 큐를 재정렬합니다.
func (pq *PriorityQueue) update(item *Item, dist int) {
    item.dist = dist    // 항목의 거리 갱신
    heap.Fix(pq, item.index)    // 우선순위 큐 재정렬
}

// Dijkstra 함수는 다익스트라 알고리즘을 구현합니다.
// 시작 정점 'start'로부터 모든 정점까지의 최단 거리를 계산합니다.
func Dijkstra(g *Graph, start int) []int {
    // 각 정점까지의 최단 거리를 저장할 배열을 초기화
    dist := make([]int, g.vertices)
    for i := range dist {
        dist[i] = math.MaxInt32 // 초기 거리를 무한대로 설정
    }
    dist[start] = 0 // 시작 정점의 거리를 0으로 설정

    // 우선순위 큐 초기화
    pq := make(PriorityQueue, 0)
    heap.Init(&pq)
    heap.Push(&pq, &Item{vertex: start, dist: 0})

    // 우선순위 큐가 빌 때까지 반복
    for pq.Len() > 0 {
        // 가장 거리가 짧은 정점을 선택
        u := heap.Pop(&pq).(*Item).vertex

        // 선택된 정점과 인접한 모든 정점을 탐색
        for _, edge := range g.edges[u] {
            v := edge.to
            weight := edge.weight
            // 새로운 경로가 더 짧으면 거리 갱신
            if dist[u]+weight < dist[v] {
                dist[v] = dist[u] + weight
                heap.Push(&pq, &Item{vertex: v, dist: dist[v]})
            }
        }
    }

    return dist
}

func main() {
    // 예제 그래프를 생성
    g := NewGraph(6)
    g.AddEdge(0, 1, 1)
    g.AddEdge(0, 3, 4)
    g.AddEdge(1, 2, 2)
    g.AddEdge(1, 3, 3)
    g.AddEdge(1, 4, 1)
    g.AddEdge(2, 5, 5)
    g.AddEdge(3, 4, 2)
    g.AddEdge(4, 5, 1)

    start := 0
    // 다익스트라 알고리즘을 실행하여 각 정점까지의 최단 거리를 계산
    distances := Dijkstra(g, start)

    // 결과를 출력
    for i, d := range distances {
        fmt.Printf("Distance from %d to %d: %d\n", start, i, d)
    }
}
```

## 비교: 배낭 문제 예제 (동적 프로그래밍)

- 다익스트라 알고리즘

    그래프에서 최단 경로를 찾는 탐욕 알고리즘으로, 각 단계에서 최단 경로를 확장합니다.

    - **목적**: 그래프에서 한 정점에서 다른 모든 정점으로의 최단 경로를 찾는 것.
    - **알고리즘 유형**: 탐욕 알고리즘 (Greedy Algorithm).
    - **작동 방식**: 각 단계에서 현재까지의 최단 경로를 확장하여 최단 경로를 찾음.
    - **시간 복잡도**: O(E log V) (E는 간선의 수, V는 정점의 수).

- 배낭 문제 (Knapsack Problem)

    동적 프로그래밍을 사용하여 문제를 작은 하위 문제로 나누고, 각 하위 문제의 해를 저장하여 중복 계산을 피합니다.

    - **목적**: 주어진 무게 제한 내에서 최대 가치를 가지는 아이템의 집합을 찾는 것.
    - **알고리즘 유형**: 동적 프로그래밍 (Dynamic Programming) 또는 분기 한정법 (Branch and Bound).
    - **작동 방식**: 동적 프로그래밍은 문제를 작은 하위 문제로 나누고, 각 하위 문제의 해를 저장하여 중복 계산을 피함. 분기 한정법은 최적화 문제를 해결하기 위해 트리 구조를 사용하여 가능한 해를 탐색하고, 가지치기를 통해 불필요한 경로를 제거함.
    - **시간 복잡도**: 동적 프로그래밍은 O(nW) (n은 아이템의 수, W는 배낭의 최대 무게), 분기 한정법은 최악의 경우 O(2^n).

다음은 동적 프로그래밍을 사용하여 배낭 문제를 해결하는 예제입니다.

```go
package main

import "fmt"

func knapsack(weights []int, values []int, W int) int {
    n := len(weights)
    dp := make([][]int, n+1)
    for i := range dp {
        dp[i] = make([]int, W+1)
    }

    for i := 1; i <= n; i++ {
        for w := 0; w <= W; w++ {
            if weights[i-1] <= w {
                dp[i][w] = max(dp[i-1][w], dp[i-1][w-weights[i-1]]+values[i-1])
            } else {
                dp[i][w] = dp[i-1][w]
            }
        }
    }

    return dp[n][W]
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func main() {
    weights := []int{2, 3, 4, 5}
    values := []int{3, 4, 5, 6}
    W := 5
    fmt.Printf("Maximum value in Knapsack = %d\n", knapsack(weights, values, W))
}
```
