# Algorithms skeleton

- [Algorithms skeleton](#algorithms-skeleton)
    - [알고리즘 개요](#알고리즘-개요)
    - [Quick Sort](#quick-sort)
    - [이진 탐색](#이진-탐색)
    - [다익스트라 알고리즘 (Dijkstra's Algorithm)](#다익스트라-알고리즘-dijkstras-algorithm)
    - [최적해 문제(최적의 시간표 짜기)](#최적해-문제최적의-시간표-짜기)
        - [그리디 알고리즘](#그리디-알고리즘)
        - [동적 프로그래밍 (Dynamic Programming)](#동적-프로그래밍-dynamic-programming)
        - [백트래킹 (Backtracking)](#백트래킹-backtracking)
        - [이진 검색 트리 (Binary Search Tree)](#이진-검색-트리-binary-search-tree)

## 알고리즘 개요

1. 퀵 정렬 (Quick Sort)
2. 이진 탐색 (Binary Search)
3. 깊이 우선 탐색 (Depth-First Search, DFS)
4. 너비 우선 탐색 (Breadth-First Search, BFS)
5. 다익스트라 알고리즘 (Dijkstra's Algorithm)
6. 동적 프로그래밍: 피보나치 수열 (Dynamic Programming: Fibonacci Sequence)
7. 최적해 문제(최적의 시간표 짜기)
8. 유니온-파인드 (Union-Find)
9. KMP 알고리즘 (Knuth-Morris-Pratt Algorithm)
10. 세그먼트 트리 (Segment Tree)

## Quick Sort

```java
import java.util.Arrays;
import java.util.Random;

public class QuickSort {

    /**
     * 퀵 정렬 알고리즘의 메인 함수
     *
     * @param arr 정렬할 배열
     * @param low 정렬 범위의 시작 인덱스
     * @param high 정렬 범위의 끝 인덱스
     */
    public static void quickSort(int[] arr, int low, int high) {
        if (low < high) {
            // 피벗을 선택하고 배열을 분할
            int pi = partition(arr, low, high);

            // 피벗을 기준으로 왼쪽과 오른쪽 부분 배열을 재귀적으로 정렬
            // 이 재귀 호출로 인해 평균적으로 O(log n) 깊이의 재귀가 발생합니다.
            // partition 함수가 O(n)이므로, 전체 시간 복잡도는 O(n log n)이 됩니다.
            quickSort(arr, low, pi - 1);
            quickSort(arr, pi + 1, high);
        }
    }

    /**
     * 배열을 피벗을 기준으로 분할하는 함수
     */
    private static int partition(int[] arr, int low, int high) {
        // 피벗 선택 (여기서는 마지막 요소를 피벗으로 사용)
        // 개선 방법: 랜덤 피벗 선택으로 최악의 경우(O(n^2))를 피할 수 있습니다.
        // int pivotIndex = new Random().nextInt(high - low + 1) + low;
        // swap(arr, pivotIndex, high);
        int pivot = arr[high];
        int i = (low - 1); // 피벗보다 작은 요소들의 경계

        // 이 루프가 O(n) 시간 복잡도의 원인입니다.
        // 최악의 경우(이미 정렬된 배열), 이 루프가 n번의 재귀 호출 동안 각각 n번씩 실행되어 O(n^2)가 됩니다.
        for (int j = low; j < high; j++) {
            // 현재 요소가 피벗보다 작거나 같으면 i를 증가시키고 교환
            if (arr[j] <= pivot) {
                // 같은 값도 스왑 대상이 되어 불안정 정렬의 원인이 됩니다.
                // 예: [3, 3, 2]에서 첫 번째 3과 2가 교환되면 두 3의 상대적 순서가 바뀝니다.
                i++;
                swap(arr, i, j);
            }
        }

        // 피벗을 올바른 위치로 이동
        swap(arr, i + 1, high);

        return i + 1;
    }

    // 배열의 두 요소를 교환하는 헬퍼 함수
    private static void swap(int[] arr, int i, int j) {
        int temp = arr[i];
        arr[i] = arr[j];
        arr[j] = temp;
    }

    public static void main(String[] args) {
        int[] arr = {10, 7, 8, 9, 1, 5};
        System.out.println("정렬 전 배열: " + Arrays.toString(arr));

        quickSort(arr, 0, arr.length - 1);

        System.out.println("정렬 후 배열: " + Arrays.toString(arr));

        // 최악의 경우 테스트 (이미 정렬된 배열)
        // 이 경우 partition 함수가 매번 가장 불균형한 분할을 만들어 O(n^2) 시간 복잡도가 됩니다.
        int[] worstCase = {1, 2, 3, 4, 5, 6};
        System.out.println("최악의 경우 정렬 전: " + Arrays.toString(worstCase));
        quickSort(worstCase, 0, worstCase.length - 1);
        System.out.println("최악의 경우 정렬 후: " + Arrays.toString(worstCase));
    }
}
```

추가적인 개선 방법에 대한 설명:

1. 3-way 파티셔닝:

   ```java
   private static void quickSort3Way(int[] arr, int low, int high) {
       if (high <= low) return;
       int lt = low, i = low + 1, gt = high;
       int pivot = arr[low];
       while (i <= gt) {
           if (arr[i] < pivot) swap(arr, lt++, i++);
           else if (arr[i] > pivot) swap(arr, i, gt--);
           else i++;
       }
       quickSort3Way(arr, low, lt - 1);
       quickSort3Way(arr, gt + 1, high);
   }
   ```

   이 방법은 중복 요소가 많은 경우 성능을 크게 개선할 수 있습니다.

2. 삽입 정렬과의 결합:

   ```java
   if (high - low + 1 <= 10) {
       insertionSort(arr, low, high);
       return;
   }
   ```

   작은 부분 배열에 대해 삽입 정렬을 사용하면 재귀 호출의 오버헤드를 줄이고 전체적인 성능을 향상시킬 수 있습니다.

## 이진 탐색

## 다익스트라 알고리즘 (Dijkstra's Algorithm)

**추상화 & 모델링**:

1. 문제 정의:
   - 가중 그래프에서 한 정점에서 다른 모든 정점까지의 최단 경로를 찾아야 함
   - 왜?: 최적해 문제에서 '최단' 또는 '최소 비용' 경로를 찾는 것이 목표인 경우가 많기 때문입니다. 예를 들어, 최소 비용으로 여러 작업을 수행하는 순서를 찾는 문제에 적용할 수 있습니다.

2. 데이터 모델:
   - 정점(Vertex): 그래프의 각 노드 (예: 작업, 위치)
   - 간선(Edge): 정점 간의 연결 (예: 작업 간 전환, 위치 간 이동)
   - 가중치(Weight): 간선의 비용 또는 거리
   - 왜?: 이 모델을 통해 다양한 실제 문제를 그래프로 추상화할 수 있습니다. 가중치는 시간, 비용, 거리 등 다양한 측면을 나타낼 수 있습니다.

3. 핵심 아이디어:
   - 탐욕적(Greedy) 선택을 반복하여 점진적으로 최단 경로를 찾음
   - 왜?: 각 단계에서 가장 가까운 미방문 정점을 선택함으로써, 전체 경로의 최적성을 보장할 수 있습니다. 이는 '최적 부분 구조' 특성을 활용한 것입니다.

**알고리즘 설계**:

1. 초기화:
   - 시작 정점의 거리를 0으로, 나머지 정점의 거리를 무한대로 설정
   - 왜?: 시작점에서의 거리를 기준으로 다른 모든 정점까지의 거리를 계산해 나가기 위함입니다.

2. 미방문 정점 중 최소 거리 정점 선택:
   - 우선순위 큐를 사용하여 효율적으로 선택
   - 왜?: 매 반복마다 최소 거리 정점을 O(log n) 시간에 찾을 수 있어 전체 알고리즘의 효율성이 향상됩니다.

3. 선택된 정점의 인접 정점들의 거리 갱신:
   - 현재까지의 거리와 새로운 경로의 거리를 비교하여 더 짧은 경로로 갱신
   - 왜?: 이 과정을 통해 모든 가능한 경로를 고려하면서 최단 경로를 점진적으로 찾아갈 수 있습니다.

4. 모든 정점을 방문할 때까지 2-3 반복:
   - 왜?: 모든 정점을 처리함으로써 시작점에서 모든 다른 정점까지의 최단 경로를 보장할 수 있습니다.

```java
import java.util.*;

/**
 * 다익스트라 알고리즘을 이용한 최단 경로 찾기
 *
 * 이 알고리즘은 가중 그래프에서 한 정점에서 다른 모든 정점까지의 최단 경로를 찾습니다.
 *
 * 시간 복잡도: O((V + E) log V), 여기서 V는 정점의 수, E는 간선의 수입니다.
 * 공간 복잡도: O(V), 거리 배열과 우선순위 큐를 위한 공간입니다.
 *
 * @param graph 인접 리스트로 표현된 가중 그래프
 * @param start 시작 정점
 * @return 시작 정점으로부터 모든 정점까지의 최단 거리를 담은 배열
 */
public static int[] dijkstra(List<List<Edge>> graph, int start) {
    int n = graph.size();
    int[] dist = new int[n];
    Arrays.fill(dist, Integer.MAX_VALUE);
    dist[start] = 0;

    // 최소 거리를 가진 정점을 빠르게 찾기 위해 우선순위 큐 사용
    PriorityQueue<Node> pq = new PriorityQueue<>((a, b) -> a.cost - b.cost);
    pq.offer(new Node(start, 0));

    while (!pq.isEmpty()) {
        Node node = pq.poll();
        int v = node.vertex;
        int cost = node.cost;

        // 이미 처리된 정점은 무시 (최적화)
        if (cost > dist[v]) continue;

        // 현재 정점의 모든 인접 정점에 대해 거리 갱신
        for (Edge edge : graph.get(v)) {
            int neighbor = edge.to;
            int newCost = cost + edge.weight;
            // 더 짧은 경로를 발견하면 거리 갱신 및 우선순위 큐에 추가
            if (newCost < dist[neighbor]) {
                dist[neighbor] = newCost;
                pq.offer(new Node(neighbor, newCost));
            }
        }
    }

    return dist;
}

// 그래프의 간선을 나타내는 클래스
static class Edge {
    int to, weight;
    Edge(int to, int weight) {
        this.to = to;
        this.weight = weight;
    }
}

// 우선순위 큐에서 사용할 노드 클래스
static class Node {
    int vertex, cost;
    Node(int vertex, int cost) {
        this.vertex = vertex;
        this.cost = cost;
    }
}

public static void main(String[] args) {
    // 그래프 생성 (예: 작업 간 전환 비용을 나타내는 그래프)
    int n = 5; // 정점(작업) 수
    List<List<Edge>> graph = new ArrayList<>();
    for (int i = 0; i < n; i++) {
        graph.add(new ArrayList<>());
    }

    // 간선(작업 간 전환) 추가
    graph.get(0).add(new Edge(1, 4)); // 작업 0에서 작업 1로 전환하는 데 비용 4
    graph.get(0).add(new Edge(2, 1));
    graph.get(1).add(new Edge(3, 1));
    graph.get(2).add(new Edge(1, 2));
    graph.get(2).add(new Edge(3, 5));
    graph.get(3).add(new Edge(4, 3));

    int start = 0; // 시작 작업
    int[] shortestPaths = dijkstra(graph, start);

    System.out.println("작업 0에서 시작하여 각 작업까지의 최소 비용:");
    for (int i = 0; i < n; i++) {
        System.out.println("작업 " + i + ": " + shortestPaths[i]);
    }
}
```

이 다익스트라 알고리즘 구현은 최적해 문제에서 다음과 같은 이점을 제공합니다:

1. 효율성: 우선순위 큐를 사용하여 O((V + E) log V) 시간 복잡도를 달성합니다.
2. 유연성: 다양한 종류의 '최소 비용' 문제에 적용할 수 있습니다.
3. 최적성 보장: 음의 가중치가 없는 그래프에서 항상 최적해를 찾습니다.

최적해 문제에서의 적용:

1. 작업 스케줄링: 각 작업을 정점으로, 작업 간 전환 비용을 간선의 가중치로 모델링하여 최적의 작업 순서를 찾을 수 있습니다.
2. 네트워크 라우팅: 네트워크에서 최소 지연 시간 또는 최대 대역폭 경로를 찾는 데 사용될 수 있습니다.
3. 프로젝트 관리: 프로젝트의 각 단계를 정점으로, 단계 간 소요 시간을 가중치로 모델링하여 프로젝트 완료까지의 최단 시간을 계산할 수 있습니다.

하지만 다음과 같은 제한사항도 고려해야 합니다:

1. 음의 가중치: 다익스트라 알고리즘은 음의 가중치가 있는 그래프에서는 정상적으로 작동하지 않습니다.
2. 메모리 사용: 대규모 그래프에서는 상당한 메모리를 사용할 수 있습니다.
3. 동적 변화: 그래프의 구조나 가중치가 자주 변하는 경우, 매번 알고리즘을 다시 실행해야 합니다.

결론적으로, 다익스트라 알고리즘은 최소 비용 경로를 찾는 다양한 최적화 문제에 적용할 수 있는 강력한 도구입니다. 특히 경로의 '비용'이 누적되는 형태의 문제에 효과적으로 적용할 수 있습니다.

## 최적해 문제(최적의 시간표 짜기)

### 그리디 알고리즘

**추상화 & 모델링**:
1. 문제 분석:

    각 활동은 시작 시간과 종료 시간을 가집니다.
    각 활동이 겹치지 않아야 최대 개수의 활동을 선택할 수 있습니다.

2. 핵심 아이디어 도출:

    더 많은 활동을 선택하려면 각 활동이 차지하는 시간을 최소화해야 합니다.
    따라서 빨리 끝나는 활동을 선택하면 남은 시간에 더 많은 활동을 선택할 가능성이 높아집니다.
    이에 따라 *가장 빨리 끝나는 활동을 우선적으로 선택하는 전략*을 세웁니다.

3. 데이터 모델 설계:

    각 활동을 (시작 시간, 종료 시간)의 쌍으로 표현합니다.
    전체 활동들을 이러한 쌍의 목록으로 모델링합니다.

**알고리즘 설계**:
1. 활동 정렬:

    앞서 수립한 "가장 빨리 끝나는 활동을 우선적으로 선택" 전략을 구현하기 위해 활동들을 종료 시간 기준으로 정렬합니다.
    이는 그리디 선택의 기준이 됩니다.

2. 첫 번째 활동(가장 빨리 끝나는 활동) 선택:

    정렬된 목록의 첫 번째 활동이 가장 빨리 끝나는 활동이므로, 이를 선택합니다.
    이는 남은 시간을 최대화하여 추후 더 많은 활동을 선택할 수 있는 가능성을 높입니다.

3. 다음 활동 선택 과정:

    이전에 선택한 활동과 겹치지 않으면서 가장 빨리 끝나는 활동을 선택해야 합니다.
    마찬가지로 남은 시간을 최대화합니다.

    정렬된 목록을 순회하면서 현재 활동의 시작 시간이 이전에 선택한 활동의 종료 시간 이후인 첫 번째 활동을 선택합니다.

    이 과정을 반복하여 더 이상 선택할 수 있는 활동이 없을 때까지 진행합니다.

4. 결과 반환: 선택된 활동들의 목록을 최종 결과로 반환합니다.

```java
import java.util.*;

/**
 * 최적의 시간표 짜기 (그리디 알고리즘 접근)
 *
 * 이 알고리즘은 주어진 활동들의 목록에서 겹치지 않는 최대 개수의 활동을 선택합니다.
 * 각 활동은 시작 시간과 종료 시간을 가지며, 한 번에 하나의 활동만 수행할 수 있습니다.
 *
 * 시간 복잡도: O(n log n), 여기서 n은 활동의 수입니다.
 * 공간 복잡도: O(n), 입력 활동들을 저장하는 데 필요한 공간입니다.
 *
 * @param activities 각 활동의 시작 시간과 종료 시간을 담은 2차원 배열
 * @return 선택된 활동들의 인덱스 리스트
 */
public static List<Integer> scheduleActivities(int[][] activities) {
    List<Integer> result = new ArrayList<>();

    // 그리디 선택 속성: 가장 빨리 끝나는 활동을 선택하기 위해 종료 시간 기준으로 정렬
    // 이 정렬이 O(n log n) 시간 복잡도의 주요 원인입니다.
    Arrays.sort(activities, (a, b) -> Integer.compare(a[1], b[1]));

    // 첫 번째 활동 선택 (가장 일찍 끝나는 활동)
    // 이는 그리디 선택의 시작점입니다. 가장 빨리 끝나는 활동을 선택함으로써
    // 남은 시간을 최대화하여 더 많은 활동을 선택할 수 있는 기회를 만듭니다.
    result.add(0);
    int lastEndTime = activities[0][1];

    // 나머지 활동들을 순회하며 겹치지 않는 활동 선택
    // 이 루프는 O(n) 시간 복잡도를 가집니다.
    for (int i = 1; i < activities.length; i++) {
        // 현재 활동의 시작 시간이 마지막으로 선택된 활동의 종료 시간보다 늦거나 같으면 선택
        // 이는 최적 부분 구조를 나타냅니다. 각 단계에서의 최적 선택이 전체 문제의 최적해에 기여합니다.
        if (activities[i][0] >= lastEndTime) {
            result.add(i);
            lastEndTime = activities[i][1];
        }
        // 선택되지 않은 활동들은 무시됩니다. 이는 그리디 알고리즘의 특성으로,
        // 현재 단계에서 최적이 아닌 선택은 고려하지 않습니다.
    }

    // 결과 리스트의 크기는 최대 n이 될 수 있으므로,
    // 이 또한 O(n)의 추가 공간 복잡도에 기여합니다.
    return result;
}

public static void main(String[] args) {
    // 테스트를 위한 활동 목록 생성
    // 각 배열은 [시작 시간, 종료 시간]을 나타냅니다.
    int[][] activities = {
        {1, 4}, {3, 5}, {0, 6}, {5, 7}, {3, 8}, {5, 9}, {6, 10}, {8, 11}, {8, 12}, {2, 13}, {12, 14}
    };

    List<Integer> scheduledActivities = scheduleActivities(activities);

    System.out.println("선택된 활동들:");
    for (int i : scheduledActivities) {
        System.out.println("활동 " + i + ": " + activities[i][0] + " - " + activities[i][1]);
    }

    // 안정성: 종료 시간이 같은 활동들의 상대적 순서는 정렬 과정에서 유지됩니다.
    // 이는 Arrays.sort가 안정 정렬을 사용하기 때문입니다.

    // 최적성 증명:
    // 귀류법을 사용하여 이 알고리즘이 최적해를 찾는다는 것을 증명할 수 있습니다.
    // 만약 다른 해가 더 많은 활동을 포함한다고 가정하면, 그 해는 반드시 이 알고리즘이
    // 선택한 첫 번째 활동보다 늦게 끝나는 활동을 포함해야 합니다. 이는 모순이므로
    // 이 알고리즘이 최적해를 찾습니다.

    // 주의 사항:
    // 그러나 활동들 사이에 우선순위나 가중치가 있는 경우에는 이 접근 방식이 최적해를
    // 보장하지 않을 수 있으며, 그런 경우 동적 프로그래밍이나 다른 접근 방식이 필요할 수 있습니다.
}
```

### 동적 프로그래밍 (Dynamic Programming)

**추상화 & 모델링**:
1. 문제 분석:

    기본적인 시간표 짜기 문제에 추가로, 각 활동에 가중치(중요도 등)가 부여되어 있습니다.
    목표는 겹치지 않는 활동들의 가중치 합을 최대화하는 것입니다.

2. 부분 문제 식별:

    그리디 접근으로는 가중치를 고려한 최적해를 보장할 수 없습니다.
    따라서 모든 가능한 조합을 고려해야 하지만, 단순한 전체 탐색은 비효율적입니다.

    이에 문제를 부분 문제로 나누어 해결하는 동적 프로그래밍 접근이 적합해 보입니다.

    부분 문제 정의:

    ```plaintext
    i번째 활동까지 고려했을 때의 최대 가중치 합
    ```

3. 데이터 모델 설계:

    각 활동을 (시작 시간, 종료 시간, 가중치)의 튜플로 표현합니다.
    부분 문제의 해를 저장할 DP 테이블을 정의합니다.

**알고리즘 설계**:

1. 활동 정렬:

    부분 문제를 순차적으로 해결하기 위해 활동들을 종료 시간 기준으로 정렬합니다.

2. DP 테이블 초기화:

    `DP[i]`를 `i`번째 활동까지 고려했을 때의 최대 가중치 합으로 정의합니다.
    이를 통해 각 단계의 최적해를 저장하고 재사용할 수 있습니다.

3. DP 테이블 채우기:

   각 활동에 대해 두 가지 선택을 고려합니다:
    - 현재 활동을 선택하지 않는 경우?

        `DP[i-1]`의 값을 그대로 사용

    - 현재 활동을 선택하는 경우?

        `현재 활동의 가중치 + DP[j]` (`j`는 현재 활동과 겹치지 않는 가장 최근의 활동 인덱스)

    각 단계에서 최적의 선택을 보장하기:

    ```java
    // 두 경우 중 더 큰 값을 `DP[i]`에 저장합니다.
    DP[i] = Math.max(DP[i-1], 현재 활동의 가중치 + DP[j]);
    ```

4. 결과 반환:

    DP 테이블의 마지막 값(`DP[n]`, `n`은 활동의 총 개수)이 전체 문제의 최적해입니다.

```java
/**
 * 동적 프로그래밍을 사용한 최적의 시간표 짜기
 *
 * 이 방법은 모든 부분 문제의 해를 저장하고 재사용합니다.
 * 가중치가 있는 활동 선택 문제에 적합합니다.
 *
 * 시간 복잡도: O(n^2), 여기서 n은 활동의 수입니다.
 * 공간 복잡도: O(n), DP 테이블을 저장하는 데 필요한 공간입니다.
 *
 * @param activities 각 활동의 [시작 시간, 종료 시간, 가중치]를 담은 2차원 배열
 * @return 선택된 활동들의 최대 가중치 합
 */
public static int scheduleActivitiesDP(int[][] activities) {
    // 활동들을 종료 시간 기준으로 정렬
    Arrays.sort(activities, (a, b) -> Integer.compare(a[1], b[1]));

    int n = activities.length;
    int[] dp = new int[n];

    // 첫 번째 활동의 가중치로 초기화
    dp[0] = activities[0][2];

    // 각 활동에 대해
    for (int i = 1; i < n; i++) {
        // i번째 활동을 선택하지 않는 경우
        int notTake = dp[i-1];

        // i번째 활동을 선택하는 경우
        int take = activities[i][2];

        // i번째 활동과 겹치지 않는 이전 활동 찾기
        int j = i - 1;
        while (j >= 0 && activities[j][1] > activities[i][0]) {
            j--;
        }
        if (j >= 0) {
            take += dp[j];
        }

        // 더 큰 가중치를 선택
        dp[i] = Math.max(take, notTake);
    }

    return dp[n-1];
}
```

### 백트래킹 (Backtracking)

**추상화 & 모델링**:

1. 문제 분석:

    겹치지 않는 최대 개수의 활동을 선택해야 하며, 모든 가능한 조합을 고려해야 합니다.
    그러나 단순한 전체 탐색은 비효율적일 수 있으므로, 불가능한 경우를 빠르게 제거할 방법이 필요합니다.

2. 탐색 전략 수립:

    각 활동에 대해 "선택" 또는 "비선택"의 두 가지 경우를 고려합니다.
    이전에 선택한 활동과 겹치는 경우는 즉시 제외하여 탐색 공간을 줄입니다.

3. 데이터 모델 설계:

    각 활동을 (시작 시간, 종료 시간)의 쌍으로 표현합니다.
    현재까지 선택한 활동들의 목록과 마지막으로 선택한 활동의 종료 시간을 추적합니다.

**알고리즘 설계**:

1. 활동 정렬:

    겹침 여부를 쉽게 확인하고 가능한 빨리 유망하지 않은 경로를 제거하기 위해 활동들을 종료 시간 기준으로 정렬합니다.

2. 재귀 함수 정의:

    '현재 고려 중인 활동 인덱스', '지금까지 선택한 활동 수', '마지막으로 선택한 활동의 종료 시간'을 매개변수로 받습니다.
    - 기저 사례: 모든 활동을 고려했을 때 현재까지 선택한 활동 수를 반환합니다.
    - 재귀 단계:
        - 현재 활동을 선택하지 않는 경우? 다음 활동으로 넘어갑니다.
        - 현재 활동을 선택할 수 있는 경우(이전 활동과 겹치지 않음)? 선택하고 다음 활동으로 넘어갑니다.
    - 두 경우의 결과 중 큰 값을 반환합니다.

    ```java
    return Math.max(
        현재 활동을 선택하지 않는 경우,
        (가능한 경우에만) 현재 활동을 선택하는 경우
    );
    ```

3. 초기 호출: 첫 번째 활동부터 시작하여 재귀 함수를 호출합니다.

4. 결과 반환: 재귀 함수의 최종 반환값이 선택 가능한 최대 활동 수입니다.

```java
/**
 * 백트래킹을 사용한 최적의 시간표 짜기
 *
 * 이 방법은 모든 가능한 조합을 탐색하며, 가지치기(pruning)를 통해 효율성을 높입니다.
 *
 * 시간 복잡도: 최악의 경우 O(2^n), 여기서 n은 활동의 수입니다.
 * 공간 복잡도: O(n), 재귀 호출 스택의 깊이입니다.
 *
 * @param activities 각 활동의 [시작 시간, 종료 시간]을 담은 2차원 배열
 * @return 선택 가능한 최대 활동 수
 */
public static int scheduleActivitiesBacktracking(int[][] activities) {
    Arrays.sort(activities, (a, b) -> Integer.compare(a[1], b[1]));
    return backtrack(activities, 0, 0);
}

private static int backtrack(int[][] activities, int index, int lastEndTime) {
    if (index == activities.length) {
        return 0;
    }

    // 현재 활동을 선택하지 않는 경우
    int notTake = backtrack(activities, index + 1, lastEndTime);

    // 현재 활동을 선택하는 경우 (가능한 경우에만)
    int take = 0;
    if (activities[index][0] >= lastEndTime) {
        take = 1 + backtrack(activities, index + 1, activities[index][1]);
    }

    return Math.max(take, notTake);
}
```

### 이진 검색 트리 (Binary Search Tree)

**추상화 & 모델링**:

1. 문제 분석:

    겹치지 않는 최대 개수의 활동을 선택해야 합니다.
    그리디 접근의 아이디어를 유지하면서도 더 효율적인 구현이 필요합니다.

2. 효율성 개선 전략:

    각 단계에서 이전 결과를 빠르게 찾고 업데이트할 수 있는 자료구조가 필요합니다.
    이진 검색 트리가 이러한 요구사항을 만족시킬 수 있습니다.

3. 데이터 모델 설계:

    각 활동을 (시작 시간, 종료 시간)의 쌍으로 표현합니다.

    그리고 이진 검색 트리를 사용하여 각 시점까지의 최대 활동 수를 저장합니다.
    - 키: 활동의 종료 시간
    - 값: 해당 시점까지의 최대 활동 수

**알고리즘 설계**:

1. 활동 정렬:

    이전 활동들과의 겹침 여부를 쉽게 확인하고, 이진 검색 트리의 키로 사용하기 위해 활동들을 종료 시간 기준으로 정렬합니다.

2. 이진 검색 트리 초기화:

    빈 이진 검색 트리를 생성합니다.
    이 트리는 *각 시점까지의 최대 활동 수를 저장*하는 데 사용됩니다.

3. 활동 순회 및 최적해 계산:
    각 활동에 대해:
    1. 현재 활동의 시작 시간보다 작거나 같은 가장 큰 종료 시간을 트리에서 찾습니다.
        이는 $O(\log{n})$ 시간에 가능하며, 현재 활동과 겹치지 않는 이전 활동들 중 가장 최근의 결과를 얻을 수 있습니다.

    2. 찾은 결과에 1을 더하여 현재 활동을 포함했을 때의 최대 활동 수를 계산합니다.

    3. 현재 활동의 종료 시간을 키로, 계산된 최대 활동 수를 값으로 트리에 삽입합니다.
        이를 통해 현재 시점까지의 최적해를 저장하여 이후 계산에 사용할 수 있습니다.

4. 결과 반환:

    트리에 저장된 값 중 가장 큰 값이 전체 문제의 최적해입니다.
    이 값을 최종 결과로 반환합니다.

```java
/**
 * 이진 검색 트리를 사용한 최적의 시간표 짜기
 *
 * 이 방법은 활동들을 종료 시간 기준으로 정렬한 후,
 * 이진 검색 트리를 사용하여 겹치지 않는 이전 활동을 빠르게 찾습니다.
 *
 * 시간 복잡도: O(n log n), 여기서 n은 활동의 수입니다.
 * 공간 복잡도: O(n), 트리를 저장하는 데 필요한 공간입니다.
 *
 * @param activities 각 활동의 [시작 시간, 종료 시간]을 담은 2차원 배열
 * @return 선택 가능한 최대 활동 수
 */
public static int scheduleActivitiesBST(int[][] activities) {
    Arrays.sort(activities, (a, b) -> Integer.compare(a[1], b[1]));
    TreeMap<Integer, Integer> dp = new TreeMap<>();
    int maxActivities = 0;

    for (int[] activity : activities) {
        int start = activity[0], end = activity[1];
        Map.Entry<Integer, Integer> prev = dp.floorEntry(start);
        int currentMax = (prev == null ? 0 : prev.getValue()) + 1;
        maxActivities = Math.max(maxActivities, currentMax);
        dp.put(end, maxActivities);
    }

    return maxActivities;
}
```
