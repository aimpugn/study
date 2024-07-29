# BFS

- [BFS](#bfs)
    - [BFS?](#bfs-1)
    - [특징](#특징)
    - [장점](#장점)
    - [공간 복잡도: `O(V)`](#공간-복잡도-ov)
    - [시간 복잡도: `O(V + E)`](#시간-복잡도-ov--e)
    - [함수의 콜 스택](#함수의-콜-스택)
    - [예제](#예제)

## BFS?

- Breadth First Search, **너비**를 우선 탐색하는 알고리즘
- 시작 노드에서 시작해서 인접한 모든 노드를 먼저 탐색
- BFS는 그래프나 트리를 탐색하기 위한 알고리즘 중 하나
    - 이 알고리즘은 시작 노드에서 출발하여 인접한 모든 노드를 먼저 방문하고,
    - 그 다음 레벨의 모든 노드를 방문하는 식으로 진행
- 이러한 방식으로 그래프나 트리를 너비 우선으로 탐색하게 됩니다.

## 특징

- BFS는 최단 경로를 찾는 문제에 자주 사용된다. DFS는 모든 경로를 탐색하므로 최단 경로를 찾는 데는 적합하지 않을 수 있다
- 메모리 사용량이 높다. BFS는 탐색할 모든 노드를 저장해야 하므로 메모리 사용량이 높을 수 있다
- 해가 존재한다면 찾는다. BFS는 모든 레벨의 노드를 탐색하기 때문에, 해가 존재한다면 반드시 찾을 수 있다
- BFS는 레벨별로 탐색을 진행하기 때문에, 레벨에 따라 해석이 필요한 문제에 적합합니다.

## 장점

- 최단 경로를 찾을 때 유용하다 (가중치가 없는 경우).
- 두 노드 사이의 최소 경로나 최소 비용을 찾는 문제에 적합하다
- 특정 경우에 시간 복잡도가 DFS보다 예측 가능하다

## 공간 복잡도: `O(V)`

- BFS 알고리즘의 공간 복잡도는 보통 `O(V)`(V = 그래프의 노드 수)
- 이는 방문한 노드와 방문할 노드를 관리하기 위해 큐 자료구조를 사용하기 때문

## 시간 복잡도: `O(V + E)`

- BFS의 시간 복잡도는 모든 노드와 엣지를 한 번씩 방문하므로 `O(V + E)`
    - V = 노드 수
    - E = 엣지 수

## 함수의 콜 스택

- BFS는 일반적으로 재귀 호출을 사용하지 않고, 명시적인 큐 자료구조를 사용해 구현된다
- 따라서 함수의 콜 스택과는 별개로 동작하며, 스택 오버플로우의 위험이 적다

## 예제

```go
package main

import "fmt"

func BFS(graph map[string][]string, start string) {
    var queue []string
    visited := make(map[string]bool)

    // 큐를 기반으로 작동하므로, 큐에 첫 노드를 추가한다
    queue = append(queue, start)
    visited[start] = true

    for len(queue) > 0 {
        // 큐에서 노드를 꺼내 방문하고, 
        vertex := queue[0]
        queue = queue[1:]
        fmt.Println(vertex)

        for _, neighbor := range graph[vertex] {
            if !visited[neighbor] {
                // 그 노드의 이웃을 큐에 추가하는 과정을 반복
                queue = append(queue, neighbor)
                visited[neighbor] = true
            }
        }
    }
}

func main() {
    graph := make(map[string][]string)
    graph["A"] = []string{"B", "C", "E"}
    graph["B"] = []string{"A", "D", "E"}
    graph["C"] = []string{"A", "F", "G"}
    graph["D"] = []string{"B"}
    graph["E"] = []string{"A", "B", "D"}
    graph["F"] = []string{"C"}
    graph["G"] = []string{"C"}

    BFS(graph, "A")
}
```
