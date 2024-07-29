# DFS & BFS

- [DFS \& BFS](#dfs--bfs)
    - [DFS는 해를 찾지 못할 수도 있다](#dfs는-해를-찾지-못할-수도-있다)
    - [DFS와 BFS 모두 적용하기 어려운 경우](#dfs와-bfs-모두-적용하기-어려운-경우)
        - [1. 동적으로 변하는 그래프](#1-동적으로-변하는-그래프)
        - [2. 가중치가 있는 최단 경로](#2-가중치가-있는-최단-경로)
        - [3. 매우 큰 규모의 그래프](#3-매우-큰-규모의-그래프)
        - [4. 확률적 탐색이 필요한 경우](#4-확률적-탐색이-필요한-경우)

## DFS는 해를 찾지 못할 수도 있다

```text
노드 A에서 노드 E까지 가는 경로를 찾고자 한다

    A
   / \
  B---C
 x     \
D-------E

B-D 경로는  어떤 이유로든 막혀있다고 가정 (예를 들어, 높은 비용이나 장애물 등).
```

- DFS (깊이 우선 탐색)
    1. A에서 시작해서 B로 이동
    2. B에서 D로 가려고 시도하지만, 경로가 막혀있어서 더 이상 진행할 수 없다
    3. 따라서 A로 돌아와서 다음 경로인 C를 탐색
    4. C에서 E로 이동
- BFS (너비 우선 탐색)
    1. A에서 시작해서 B와 C로 이동
    2. B에서는 D로 가려고 하지만, 경로가 막혀있어서 더 이상 진행할 수 없다
    3. C에서 E로 이동

DFS와 BFS 모두 해를 찾을 수 있지만, DFS의 경우 B에서 D로 가는 경로가 막혀있어서 한 번 실패한 후 다시 A로 돌아가야 한다.
이처럼 DFS는 깊이를 우선으로 탐색하기 때문에, 어떤 경로가 막혀있거나 조건을 만족하지 못하면 다시 돌아와야 하는 경우가 발생할 수 있다.

반면에 BFS는 너비를 우선으로 탐색하기 때문에, A에서 B와 C를 동시에 고려한다. 따라서 B에서 실패하더라도, 그와 동시에 C에서 성공하는 경로를 찾을 수 있다

따라서 예를 들어, 탐색 과정에서 시간 제한이 걸려있거나, 메모리가 부족한 경우, 또는 특정 조건을 만족하는 해만 찾아야 하는 경우 등 DFS가 해를 찾지 못할 수 있다.

## DFS와 BFS 모두 적용하기 어려운 경우

DFS(깊이 우선 탐색)와 BFS(너비 우선 탐색)는 그래프 탐색의 기본 알고리즘이지만, 특정 상황에서는 적용이 어렵거나 비효율적일 수 있습니다.

DFS와 BFS는 기본적으로 다음 사항들을 가정합니다.
- 정적일 것
- 가중치가 없을 것
- 메모리에 전체를 저장할 수 있는 그래프일 것
또한, 결정론적인 탐색 방식을 따릅니다.
이러한 가정들이 충족되지 않는 상황에서는 다른 접근 방법이 필요합니다.

### 1. 동적으로 변하는 그래프

- 특징:
- 예시:

동적으로 변하는 그래프는 시간에 따라 노드와 엣지가 추가되거나 제거되는 그래프로, 탐색 과정 중 그래프의 구조가 변경될 수 있습니다.
예를 들어, 실시간 네트워크 토폴로지 변화, 동적 소셜 네트워크 등이 있습니다.

이러한 환경에서 기존의 DFS나 BFS는 다음과 같은 한계를 가집니다:
1. 탐색 도중 그래프가 변경되면 결과의 일관성이 깨질 수 있습니다.
2. 변경될 때마다 전체 그래프를 재탐색하는 것은 비효율적입니다.
3. 빠르게 변화하는 그래프에 대해 즉각적인 대응이 어렵습니다.

이런 경우 다음과 같은 접근이 가능합니다.

1. 증분 업데이트 (Incremental Update):
    - 그래프의 변화를 실시간으로 반영하여 탐색 결과를 업데이트합니다.
    - 전체 그래프를 재탐색하는 대신, 변경된 부분만 효율적으로 갱신합니다.

2. 시간 윈도우 기반 탐색 (Time-Window Based Exploration):
    - 특정 시간 간격 동안의 그래프 스냅샷을 사용하여 탐색을 수행합니다.
    - 주기적으로 그래프를 재구성하고 탐색을 갱신합니다.

3. 적응형 탐색 알고리즘 (Adaptive Exploration Algorithms):
    - 그래프의 변화를 감지하고 탐색 전략을 동적으로 조정합니다.
    - 예: 강화학습 기반의 적응형 탐색 전략

아래와 같이 시간에 따라 변화하는 동적 그래프를 추상화하고 모델링해볼 수 있습니다.

1. 그래프 $G = (V, E, T)$
    - V: 노드 집합
    - E: 엣지 집합
    - T: 시간 변수

2. 변경 연산:
    - addNode(v, t): 시간 t에 노드 v 추가
    - removeNode(v, t): 시간 t에 노드 v 제거
    - addEdge(u, v, t): 시간 t에 엣지 (u, v) 추가
    - removeEdge(u, v, t): 시간 t에 엣지 (u, v) 제거

3. 증분 업데이트 함수 $U(G, \Delta)$:
    - $U$: 업데이트 전 원래 그래프 $G$와 변경 사항 집합 $\Delta$를 입력으로 받아 새로운 그래프 $G'$를 생성하는 함수
    - $\Delta$: 그래프에 적용될 변경 사항들의 집합($\{\cdots \mid \cdots\}$)
    - $G'$: 업데이트 후의 새로운 그래프

    제약 조건:

    $\Delta = \{(op, element, t) \mid op \in \{add, remove\}, element \in V \cup E, t \text{ is time}\}$

    - $(op, element, t)$: $\Delta$ 집합의 각 원소를 나타냅니다.
        - $op$: 연산(operation)
        - $element$: 변경될 그래프의 요소(노드 또는 엣지)
        - $t$: 변경이 일어나는 시간

    - $op \in \{add, remove\}$: 연산 $op$가 'add'(추가) 또는 'remove'(제거) 중 하나임을 나타냅니다.

    - $element \in V \cup E$:
        - $element$는 $V$(노드 집합)와 $E$(엣지 집합)의 합집합에 속합니다.
        - 즉, 변경될 수 있는 요소는 노드이거나 엣지입니다.

    - $t \text{ is time}$: $t$는 각 변경 사항이 언제 적용되는지를 지정합니다.

4. 경로 재계산 함수 $P' = R(G', P, v_s, v_t)$:
    - $R$: 경로 재계산 함수
    - $G'$: 업데이트된 그래프
    - $P$: 원래의 경로
    - $v_s$: 시작 노드
    - $v_t$: 목표 노드
    - $P'$: 재계산된 경로

    제약 조건:

    $\forall v_i, v_{i+1} \in P', (v_i, v_{i+1}) \in E'$

동적 그래프를 구현하면 다음과 같습니다.

```java
/**
 * 동적 그래프에서 증분 업데이트를 수행하는 메서드
 *
 * 이 메서드는 그래프의 변경사항을 받아 기존 탐색 결과를 효율적으로 갱신합니다.
 * 전체 그래프를 재탐색하는 대신, 변경된 부분만 처리하여 계산 비용을 줄입니다.
 *
 * 그래프 변경 처리 과정:
 * 1. 노드 추가/제거
 * 2. 엣지 추가/제거
 * 3. 영향받은 경로 재계산
 *
 * 초기 그래프:   변경 후:
 *    A          A---D
 *   / \   =>   / \ /
 *  B   C      B   C
 *
 * 시간 복잡도:
 * - 최선의 경우: O(1), 변경이 탐색 결과에 영향을 주지 않는 경우
 * - 평균의 경우: O(log N), N은 노드의 수, 변경이 일부 경로만 영향을 줄 때
 * - 최악의 경우: O(N), 변경이 전체 그래프에 영향을 줄 때
 *
 * 공간 복잡도: O(N), N은 노드의 수
 *
 * @param graph 현재 그래프 상태
 * @param changes 그래프에 적용할 변경사항 목록
 * @param currentResults 현재 탐색 결과
 * @return 업데이트된 탐색 결과
 */
public static Map<Node, List<Node>> incrementalUpdate(
    Graph graph,
    List<GraphChange> changes,
    Map<Node, List<Node>> currentResults
) {
    for (GraphChange change : changes) {
        // 변경 유형에 따라 처리
        switch (change.getType()) {
            case NODE_ADDED:
                // 새 노드 추가 로직
                graph.addNode(change.getNode());
                break;
            case NODE_REMOVED:
                // 노드 제거 및 관련 경로 갱신 로직
                graph.removeNode(change.getNode());
                currentResults.remove(change.getNode());
                break;
            case EDGE_ADDED:
            case EDGE_REMOVED:
                // 엣지 변경에 따른 경로 재계산 로직
                Node source = change.getEdge().getSource();
                Node target = change.getEdge().getTarget();
                if (change.getType() == ChangeType.EDGE_ADDED) {
                    graph.addEdge(source, target);
                } else {
                    graph.removeEdge(source, target);
                }
                // 영향받은 경로 재계산
                updateAffectedPaths(graph, source, currentResults);
                updateAffectedPaths(graph, target, currentResults);
                break;
        }
    }

    return currentResults;
}

/**
 * 특정 노드와 관련된 경로를 재계산하는 보조 메서드
 *
 * @param graph 현재 그래프 상태
 * @param node 재계산이 필요한 노드
 * @param results 현재 탐색 결과
 */
private static void updateAffectedPaths(
    Graph graph,
    Node node,
    Map<Node, List<Node>> results
) {
    // 이 노드를 포함하는 모든 경로를 찾아 재계산
    for (Map.Entry<Node, List<Node>> entry : results.entrySet()) {
        List<Node> path = entry.getValue();
        if (path.contains(node)) {
            Node start = path.get(0);
            Node end = path.get(path.size() - 1);
            List<Node> updatedPath = recalculatePath(graph, start, end);
            entry.setValue(updatedPath);
        }
    }
}

/**
 * A* 알고리즘을 사용하여 경로를 재계산하는 메서드
 *
 * A* 알고리즘은 최단 경로를 찾기 위해 휴리스틱 함수를 사용하는 정보 있는 탐색 알고리즘입니다.
 * 이 알고리즘은 다익스트라 알고리즘의 확장 버전으로, 목표까지의 추정 거리를 고려하여 탐색 방향을 결정합니다.
 *
 * @param graph 현재 그래프 상태
 * @param start 시작 노드
 * @param end 목표 노드
 * @return 재계산된 경로
 */
private static List<Node> recalculatePath(
    Graph graph,
    Node start,
    Node end
) {
    // openSet: 탐색할 노드들을 저장하는 우선순위 큐
    // fScore(예상 총 비용)가 가장 낮은 노드를 먼저 탐색합니다.
    PriorityQueue<Node> openSet = new PriorityQueue<>(
        Comparator.comparingDouble(n -> n.gScore + n.hScore));

    // cameFrom: 각 노드에 대한 최적 이전 노드를 저장
    Map<Node, Node> cameFrom = new HashMap<>();

    // gScore: 시작 노드로부터 각 노드까지의 실제 비용
    Map<Node, Double> gScore = new HashMap<>();

    // fScore: 시작 노드로부터 목표 노드까지의 예상 총 비용
    // fScore = gScore + heuristic
    Map<Node, Double> fScore = new HashMap<>();

    // 시작 노드 초기화
    gScore.put(start, 0.0);
    fScore.put(start, heuristic(start, end));
    openSet.add(start);

    /*
     * A* 알고리즘 실행 과정:
     * - S: 시작 노드
     * - E: 목표 노드
     * - (g, h, f)
     *   - g: 시작 노드에서 현재 노드까지의 실제 비용
     *   - h: 현재 노드에서 목표 노드까지의 추정 비용
     *   - f: 전체 경로의 추정 비용 (g + h)
     * - `*`: 현재 선택된 노드
     * - 간선 위의 숫자: 두 노드 사이의 이동 비용
     *
     * !주의: h(추정 비용)는 목표까지의 직선 거리로 계산했다고 가정합니다.
     * 실제 구현에서는 문제에 따라 다른 휴리스틱 함수를 사용할 수 있습니다.
     *
     * 초기 상태: S 노드 시작
     *             S(0,5,5)5*
     *            ╱  |  ╲
     *          ╱    |    ╲
     *        1      2      3
     *      ╱        |        ╲
     *    ╱          |          ╲
     *  A(1,4,5)   B(2,3,5)  C(3,2,5)
     *   |    ╲      |      ╱    |
     *   1      2    1    2      3
     *   |       ╲   |   ╱       |
     *   |         ╲ | ╱         |
     *  D(2,3,5)  E(3,0,3)   F(6,1,7)
     *               ^
     *              목표
     *
     * OpenSet: [S(0,5,5)]
     * ClosedSet: []
     *
     * Step 1: S 노드 처리 후
     *             S(0,5,5)5
     *            ╱  |  ╲
     *          ╱    |    ╲
     *        1      2      3
     *      ╱        |        ╲
     *    ╱          |          ╲
     *  A(1,4,5)*  B(2,3,5)  C(3,2,5)
     *   |    ╲      |      ╱    |
     *   1      2    1    2      3
     *   |       ╲   |   ╱       |
     *   |         ╲ | ╱         |
     *  D(2,3,5)  E(3,0,3)   F(6,1,7)
     *               ^
     *              목표
     *
     * OpenSet: [A(1, 4, 5), B(2, 3, 5), C(3, 2, 5)]
     * ClosedSet: [S]
     *
     * Step 2: A 노드 처리 후
     *             S(0,5,5)5
     *            ╱  |  ╲
     *          ╱    |    ╲
     *        1      2      3
     *      ╱        |        ╲
     *    ╱          |          ╲
     *  A(1,4,5)  B(2,3,5)*  C(3,2,5)
     *   |    ╲      |      ╱    |
     *   1      2    1    2      3
     *   |       ╲   |   ╱       |
     *   |         ╲ | ╱         |
     *  D(2,3,5)  E(3,0,3)   F(6,1,7)
     *               ^
     *              목표
     *
     * OpenSet: [B(2,3,5), C(3,2,5), D(2,3,5), E(3,0,3)]
     * ClosedSet: [S, A]
     *
     * Step 3: E 노드 처리 (A를 통한 경로가 최단)
     *             S(0,5,5)5
     *            ╱  |  ╲
     *          ╱    |    ╲
     *        1      2      3
     *      ╱        |        ╲
     *    ╱          |          ╲
     *  A(1,4,5)  B(2,3,5)  C(3,2,5)
     *   |    ╲      |      ╱    |
     *   1      2    1    2      3
     *   |       ╲   |   ╱       |
     *   |         ╲ | ╱         |
     *  D(2,3,5)  E(3,0,3)*   F(6,1,7)
     *               ^
     *              목표
     *
     * OpenSet: [B(2,3,5), C(3,2,5), D(2,3,5), F(6,1,7)]
     * ClosedSet: [S, A, E]
     *
     * Step 4: 목표 노드 (E) 도달, 알고리즘 종료
     *
     * 최종 경로 역추적:
     * - E의 이전 노드 = A
     * - A의 이전 노드 = S
     * - 최종 경로: S -> A -> E
     */
    while (!openSet.isEmpty()) {
        // fScore가 가장 낮은 노드를 선택
        Node current = openSet.poll();

        // 목표 노드에 도달했다면 경로 재구성
        if (current.equals(end)) {
            return reconstructPath(cameFrom, current);
        }

        // 현재 노드의 모든 이웃 노드를 탐색
        for (Node neighbor : graph.getNeighbors(current)) {
            // 현재 노드를 거쳐 이웃 노드로 가는 비용 계산
            // 여기서는 모든 엣지의 비용을 1로 가정
            double tentativeGScore = gScore.get(current) + 1;

            // 더 나은 경로를 발견했다면 정보 업데이트
            if (tentativeGScore < gScore.getOrDefault(neighbor, Double.MAX_VALUE)) {
                // 경로 정보 업데이트
                cameFrom.put(neighbor, current);
                gScore.put(neighbor, tentativeGScore);
                // fScore 계산: 현재까지의 비용 + 목표까지의 추정 비용
                fScore.put(neighbor, gScore.get(neighbor) + heuristic(neighbor, end));

                // 아직 탐색하지 않은 노드라면 openSet에 추가
                if (!openSet.contains(neighbor)) {
                    openSet.add(neighbor);
                }
            }
        }
    }

    // 경로를 찾지 못한 경우 빈 리스트 반환
    return new ArrayList<>();
}

private static double heuristic(Node a, Node b) {
    // Simple heuristic: could be replaced with a more sophisticated one
    return 1;
}

/**
 * 경로를 재구성합니다.
 *
 * @param cameFrom 현재 노드가 키, 현재 노드 오기 직전의 노드가 값인 맵으로,
 *                 맵<현재_노드, 직전_노드> 구조입니다.
 */
private static List<Node> reconstructPath(
    Map<Node, Node> cameFrom,
    Node current
) {
    List<Node> path = new ArrayList<>();
    // 현재 노드(목표 노드)를 경로에 추가
    // [E]
    path.add(current);

    /*
     * cameFrom = {
     *   E: B,  // E로 오기 직전 노드는 B
     *   B: A,  // B로 오기 직전 노드는 A
     *   A: S   // A로 오기 직전 노드는 S
     * }
     */
    // current 노드가 cameFrom 맵에 있는 동안 (즉, 시작 노드에 도달할 때까지) 반복
    while (cameFrom.containsKey(current)) {
        // 현재 노드의 '이전 노드'를 가져와서 현재 노드로 삼습니다.
        current = cameFrom.get(current);
        // 새로 찾은 이전 노드를 경로의 맨 앞에 추가하여 역순으로 경로를 구성:
        // [B, E]
        // [A, B, E]
        // [S, A, B, E]
        // 이를 통해 시작 노드(S)에서 목표 노드(E)까지의 올바른 순서의 경로를 재구성합니다.
        path.add(0, current);
    }
    return path;  // 완성된 경로 반환
}

// 메인 메서드 예시
public static void main(String[] args) {
    Graph graph = new Graph();
    // 초기 그래프 설정
    graph.addNode(new Node("A"));
    graph.addNode(new Node("B"));
    graph.addNode(new Node("C"));
    graph.addEdge(new Node("A"), new Node("B"));
    graph.addEdge(new Node("A"), new Node("C"));

    // 초기 탐색 결과 (예시)
    Map<Node, List<Node>> initialResults = new HashMap<>();
    initialResults.put(new Node("A"), Arrays.asList(new Node("A"), new Node("B"), new Node("C")));

    // 그래프 변경사항
    List<GraphChange> changes = new ArrayList<>();
    changes.add(new GraphChange(ChangeType.NODE_ADDED, new Node("D")));
    changes.add(new GraphChange(ChangeType.EDGE_ADDED, new Edge(new Node("A"), new Node("D"))));
    changes.add(new GraphChange(ChangeType.EDGE_ADDED, new Edge(new Node("C"), new Node("D"))));

    // 증분 업데이트 수행
    Map<Node, List<Node>> updatedResults = incrementalUpdate(graph, changes, initialResults);

    // 결과 출력
    System.out.println("업데이트된 탐색 결과:");
    for (Map.Entry<Node, List<Node>> entry : updatedResults.entrySet()) {
        System.out.println(entry.getKey() + ": " + entry.getValue());
    }
}
```

### 2. 가중치가 있는 최단 경로

- 특징: 간선에 가중치가 있는 경우 단순 DFS/BFS로는 최단 경로를 찾기 어렵습니다.
- 예: 다익스트라 알고리즘이나 A* 알고리즘이 필요한 경우

### 3. 매우 큰 규모의 그래프

- 특징: 메모리 제한으로 전체 그래프를 저장할 수 없는 경우
- 예: 웹 크롤링, 대규모 소셜 네트워크 분석

### 4. 확률적 탐색이 필요한 경우

- 특징: 결정론적 탐색으로는 해결이 어려운 경우
- 예: 몬테카를로 트리 탐색이 필요한 게임 AI
