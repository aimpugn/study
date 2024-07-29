# Dynamic Progrmming

- [Dynamic Progrmming](#dynamic-progrmming)
    - [Dynamic Programming](#dynamic-programming)
    - [다이나믹 프로그래밍의 기본 원리](#다이나믹-프로그래밍의-기본-원리)
    - [다이나믹 프로그래밍의 구현 방식](#다이나믹-프로그래밍의-구현-방식)
    - [최적해 보장 원리](#최적해-보장-원리)
    - [예제: 배낭 문제 (Knapsack Problem)](#예제-배낭-문제-knapsack-problem)
    - [동적 프로그래밍 문제인지 판별하기](#동적-프로그래밍-문제인지-판별하기)
    - [작은 문제로 나누기와 점화식](#작은-문제로-나누기와-점화식)
    - [수학적으로 검증하기](#수학적으로-검증하기)
    - [진행 방향](#진행-방향)
        - [Bottom-up DP](#bottom-up-dp)
        - [Top-down DP](#top-down-dp)
    - [knapsack problem](#knapsack-problem)
        - [Knapsack 문제의 종류](#knapsack-문제의-종류)
        - [다이나믹 프로그래밍 접근법 (0/1 Knapsack에 대한)](#다이나믹-프로그래밍-접근법-01-knapsack에-대한)
        - [예제](#예제)
            - [기본 예제](#기본-예제)
            - [`0/1 Knapsack` 문제 다이나믹 프로그래밍 접근법](#01-knapsack-문제-다이나믹-프로그래밍-접근법)
    - [코드를 통해 분석하기](#코드를-통해-분석하기)
        - [미로 탈출](#미로-탈출)
        - [미로 탈출 문제 DP + DFS](#미로-탈출-문제-dp--dfs)

## Dynamic Programming

다이나믹 프로그래밍(Dynamic Programming, DP)은 최적해를 구하는 효율적인 알고리즘 기법 중 하나입니다.
다이나믹 프로그래밍은 문제를 더 작은 하위 문제들로 나누고, 각 하위 문제의 해를 저장하여 필요할 때 다시 사용함으로써 전체 문제를 해결합니다.
이를 통해 중복 계산을 방지하고 시간 복잡도를 줄일 수 있습니다.

 동적 프로그래밍은 크게 두 가지 요소로 나뉩니다:
1. 하위 문제를 정의하고 해결한다. (Divide-and-Conquer)
2. 하위 문제의 해결책을 저장하고 재사용한다. (Memoization)

일반적인 DP 문제에서는 각 상태의 최적 해가 그 상태에 이르게 만든 이전 상태에 독립적입니다.
DP는 순차적인 결정 과정을 모델링하는 데 적합합니다.

예를 들어, `n`번째 단계에서의 최적의 선택은 1부터 `n-1`까지의 선택에만 의존합니다.

## 다이나믹 프로그래밍의 기본 원리

1. **최적 부분 구조 (Optimal Substructure)**:

   문제의 최적해가 하위 문제들의 최적해로부터 만들어질 수 있어야 합니다.

   예를 들어, 최단 경로 문제에서 *전체 경로의 최단 거리는 그 경로를 구성하는 하위 경로들의 최단 거리의 합으로 결정*됩니다.

2. **중복되는 부분 문제 (Overlapping Subproblems)**:

   동일한 작은 문제들이 반복적으로 계산되는 경우가 많습니다.

   예를 들어, 피보나치 수열 계산에서 동일한 피보나치 수를 여러 번 계산하게 됩니다.

## 다이나믹 프로그래밍의 구현 방식

1. **탑다운 방식 (Top-Down)**: `Memoization`

   재귀적 접근법을 사용하여 문제를 해결합니다.
   이때 메모이제이션(Memoization)을 통해 이미 계산된 하위 문제의 해를 저장하고 재사용합니다.

   ```python
   def fib(n, memo={}):
       if n in memo:
           return memo[n]
       if n <= 2:
           return 1

       memo[n] = fib(n-1, memo) + fib(n-2, memo)

       return memo[n]
   ```

2. **바텀업 방식 (Bottom-Up)**: 테이블 저장

   반복문을 사용하여 작은 문제부터 차례대로 해결해 나갑니다.
   이를 통해 하위 문제의 해를 차례로 구하고, 이를 이용해 전체 문제를 해결합니다.

   ```python
   def fib(n):
       if n <= 2:
           return 1

       dp = [0] * (n + 1)
       dp[1], dp[2] = 1, 1
       for i in range(3, n + 1):
           dp[i] = dp[i-1] + dp[i-2]

       return dp[n]
   ```

## 최적해 보장 원리

다이나믹 프로그래밍이 최적해를 보장하는 이유는 다음과 같습니다:

1. **재귀적 관계 정의**:

   문제를 하위 문제들로 나누고, 그 관계를 재귀적으로 정의합니다.
   이때 각 하위 문제의 해가 최적해를 제공하도록 정의됩니다.

   예를 들어, 배낭 문제(Knapsack Problem)에서 최대 가치를 얻기 위해 아이템을 선택하는 문제는 현재 아이템을 포함할지 여부에 따라 문제를 나누어 재귀적으로 해결합니다.

2. **하위 문제의 최적해 사용**:

   각 하위 문제는 최적해를 저장하고, 이를 기반으로 상위 문제를 해결합니다.
   이는 메모이제이션이나 테이블 저장을 통해 구현됩니다.

   예를 들어, 피보나치 수열 계산에서 이전 두 수의 합을 사용하여 다음 수를 계산함으로써 최적해를 구합니다.

3. **정확한 상태 저장 및 갱신**:

   DP 테이블을 사용하여 각 상태에서의 최적해를 저장하고, 이 값을 사용하여 다음 상태의 해를 구합니다.
   이를 통해 최적해가 보장됩니다.

   예를 들어, 최단 경로 문제에서 각 노드까지의 최단 거리를 저장하고 이를 이용하여 다른 노드까지의 최단 거리를 갱신합니다.

## 예제: 배낭 문제 (Knapsack Problem)

배낭 문제는 무게 제한이 있는 배낭에 최대 가치를 가지도록 물건을 넣는 문제입니다. 이 문제는 다이나믹 프로그래밍을 사용하여 해결할 수 있습니다.

```python
def knapsack(values, weights, W):
    n = len(values)
    dp = [[0 for _ in range(W + 1)] for _ in range(n + 1)]
    for i in range(1, n + 1):
        for w in range(1, W + 1):
            if weights[i-1] <= w:
                dp[i][w] = max(dp[i-1][w], dp[i-1][w-weights[i-1]] + values[i-1])
            else:
                dp[i][w] = dp[i-1][w]
    return dp[n][W]
```

이 예제에서 `dp[i][w]`는 `i`번째 아이템까지 고려했을 때 배낭의 무게가 `w`일 때의 최대 가치를 저장합니다.
이처럼 다이나믹 프로그래밍은 문제를 작은 문제로 분할하고, 각 작은 문제의 최적해를 저장하여 전체 문제의 최적해를 구하는 방식으로 동작합니다.

## 동적 프로그래밍 문제인지 판별하기

1. 최적 부분 구조 (Optimal Substructure): 큰 문제의 최적 해결책이 작은 문제의 최적 해결책으로부터 구성된다
2. 중복되는 부분 문제 (Overlapping Sub-problems): 동일한 작은 문제가 반복적으로 발생한다

## 작은 문제로 나누기와 점화식

- 문제를 작은 문제로 나누는 것은 문제의 성격에 따라 다르다
    - 일반적으로 입력 데이터를 줄이거나,
    - 해결해야 하는 문제의 크기를 줄인다
- 이 작은 문제들을 해결한 뒤에는 그 해결책을 이용하여 점화식(recurrence relation)을 만든다

예를 들어, 피보나치 수열의 경우:

$$F(n) = F(n - 1) + F(n - 2)$$

## 수학적으로 검증하기

- 수학적인 검증은 종종 점화식을 사용하여 이루어진다
- 점화식이 정확하다면, 기저 조건(base case)부터 시작하여 점화식을 사용해 다음 값을 계산함으로써 문제를 해결할 수 있다

## 진행 방향

### Bottom-up DP

### Top-down DP

## knapsack problem

- 조합 최적화의 문제
- 주어진 물건들의 가치와 무게, 그리고 배낭의 무게 한도가 있을 때, 배낭에 넣을 수 있는 물건들의 가치의 합을 최대화하는 문제

### Knapsack 문제의 종류

- **0/1 Knapsack**
    - 아이템을 통째로 넣거나 아예 뺀다
    - 즉, 부분적으로 넣을 수 없다
- **Fractional Knapsack**
    - 아이템을 부분적으로 넣을 수 있다

### 다이나믹 프로그래밍 접근법 (0/1 Knapsack에 대한)

- 다이나믹 프로그래밍을 사용한 **0/1 Knapsack** 문제의 해결 방법은 2차원 테이블을 사용하는 것
    - i-행, j-열의 항목 = i개의 아이템을 고려했을 때, 최대 무게가 j인 Knapsack에 넣을 수 있는 최대 가치
- 중복 계산을 피함으로써 문제 해결 과정을 최적화하는 데 유리

### 예제

#### 기본 예제

```go
package main

import "fmt"

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func knapSack(W int, wt []int, val []int, n int) int {
    if n == 0 || W == 0 {
        return 0
    }
    if wt[n-1] > W {
        return knapSack(W, wt, val, n-1)
    } else {
        return max(
            val[n-1]+knapSack(W-wt[n-1], wt, val, n-1),
            knapSack(W, wt, val, n-1),
        )
    }
}

func main() {
    val := []int{60, 100, 120}
    wt := []int{10, 20, 30}
    W := 50
    n := len(val)
    fmt.Println(knapSack(W, wt, val, n))
}
```

#### `0/1 Knapsack` 문제 다이나믹 프로그래밍 접근법

```go
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// knapSack
// - `W`: Knapsack이 견딜 수 있는 최대 무게
// - `wt`: 아이템 무게의 배열
// - `val`: 아이템 가치의 배열
// - `n`: 아이템 수
// 
// 시간 복잡도: O(nW), n은 아이템의 수, W는 최대 무게
// 공간 복잡도: O(nW), DP 테이블을 저장하기 위해 필요
func knapSack(W int, wt []int, val []int, n int) int {
    K := make([][]int, n+1)
    for i := range K {
        K[i] = make([]int, W+1)
    }

    for i := 0; i <= n; i++ {
        for w := 0; w <= W; w++ {
            if i == 0 || w == 0 {
                K[i][w] = 0
            } else if wt[i-1] <= w {
                K[i][w] = max(val[i-1]+K[i-1][w-wt[i-1]], K[i-1][w])
            } else {
                K[i][w] = K[i-1][w]
            }
        }
    }
    return K[n][W]
}
```

## 코드를 통해 분석하기

### 미로 탈출

[미로 탈출 문제](./examples/bygo/kakao/find_maze_exit/README.md)

- 목표: 시작 위치 (x, y)에서 목표 위치 (r, c)까지 이동하는 데 필요한 최소 이동 경로를 찾는 것
- 최대 `k`번의 이동이 가능
- 이동 방향은 위(`u`), 아래(`d`), 왼쪽(`l`), 오른쪽(`r`) 중 하나

```go
// 어떤 상태
type State struct {
    X int // 현재 위치의 X 좌표
    Y int // 현재 위치의 Y 좌표
    K int // 이동 횟수
}

var moves = []struct {
    dx   int // x축 이동 방향
    dy   int // y축 이동 방향
    move string // 이에 대응하는 문자열
}{
    {1, 0, "d"},
    {0, -1, "l"},
    {0, 1, "r"},
    {-1, 0, "u"},
}

func solutionByDP2(n int, m int, x int, y int, r int, c int, k int) string {
    minStr := func(a, b string) string {
        if a == "" {
            return b
        }
        if b == "" {
            return a
        }
        if a < b {
            return a
        }
        return b
    }

    // dp 맵을 사용하여 이전에 계산한 상태의 값을 저장하고 재사용
    // Key: 어떤 `State` 
    // Value: 해당 상태에서 목적지까지 가는 최단 경로
    dp := make(map[State]string)

    // Initialize Base Case
    // 시작 위치(x, y)와 남은 이동 횟수 k를 가진 상태의 값을 빈 문자열로 초기화한다
    // 이는 시작 위치에서 아무런 이동도 하지 않았기 때문
    dp[State{x, y, k}] = ""

    // Dynamic Programming
    // `k`: 최대 이동 가능 횟수
    for step := k; step >= 0; step-- { // 남은 이동 횟수 `step`을 기준으로 반복문을 시작
        // 미로의 모든 위치 (i, j)에 대해 반복문을 실행
        for i := 1; i <= n; i++ {
            for j := 1; j <= m; j++ {
                for _, move := range moves {
                    // 현재 위치 (i, j)에서 이동 후의 위치는 (nx, ny)이고, 이동 후의 남은 이동 횟수는 nk
                    nx, ny, nk := i+move.dx, j+move.dy, step-1
                    // 만약 이동 후의 상태 (nx, ny, nk)가 미로 내에 있고, nk >= 0이라면, 상태를 업데이트
                    if nx >= 1 && nx <= n && ny >= 1 && ny <= m && nk >= 0 {
                        prevState := State{i, j, step}
                        currState := State{nx, ny, nk}
                        if prevPath, exists := dp[prevState]; exists {
                            newPath := prevPath + move.move
                            dp[currState] = minStr(dp[currState], newPath)
                        }
                    }
                }
            }
        }
    }

    // Get the result
    // 목표 위치 (r, c)와 남은 이동 횟수 0을 가진 상태의 값을 반환
    result, exists := dp[State{r, c, 0}]
    if !exists {
        return "impossible"
    }
    return result
}
```

### 미로 탈출 문제 DP + DFS

```go
package main

const impossible = "xxxxxxxxxxxxxxxxxx"

type Cell struct {
    row    int
    column int
    k      int
}

func min(a, b string) string {
    if a < b {
        return a
    }
    return b
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

// 1 2 3 4
// . . . . 1
// . . . . 2
// . . . . 3
var lrdu = []struct {
    toRow    int
    toColumn int
    keyword  string
    symbol   string
}{
    // d, l, r, u 순서가 중요하다
    {0, -1, "l", "<-"},
    {0, 1, "r", "->"},
    {1, 0, "d", "↓"},
    {-1, 0, "u", "↑"},
}

func dpdfs(rowLimit, columnLimit, row, column, rowDst, columnDst, k int, dp map[Cell]string, path string) {
    currGrid := Cell{row, column, k}


    if dpValue, exists := dp[currGrid]; exists && dpValue < path {
        return
    }


type Cell struct {
    row, col, k int
}

var directions = []struct {
    toRow    int
    toColumn int
    keyword  string
}{
    // `d`, `l`, `r`, `u의` 순서가 중요하다
    // 가령 `u`, `r`, `l`, `d` 순서로 순회하면 시간 초과가 발생한다
    {1, 0, "d"},
    {0, -1, "l"},
    {0, 1, "r"},
    {-1, 0, "u"},
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

func dpdfs(rowLimit, columnLimit, currRow, currColumn, dstRow, dstColumn, mvCnt int, dp map[Cell]string, backtrackingPath string, lexicographicallyBest *string) {
    if mvCnt < 0 {
        return
    }

    manhattanDistance := abs(dstRow-currRow) + abs(dstColumn-currColumn)
    if mvCnt < manhattanDistance || (mvCnt-manhattanDistance)%2 != 0 {
        return
    }

    // 여기서 아래처럼 단순히 방문 여부만 체크하면 안된다.
    // 다른 경로로 같은 `Cell`에 도달할 수 있으므로 더 작은 값인지 여부도 체크해야 한다.
    // (X)
    // if dpValue, exists := dp[currGrid]; exists {
    //     dp[currGrid] = min(dpValue, path)
    //     return
    // }
    currentCell := Cell{currRow, currColumn, mvCnt}
    if vlaue, ok := dp[currentCell]; ok && vlaue != "" && vlaue <= backtrackingPath {
        return
    }

    // 목적지에 도착한 경우
    if currRow == dstRow && currColumn == dstColumn && mvCnt == 0 {
        dp[currentCell] = backtrackingPath
        // `lexicographically`가 비어있으면 초기값 개념으로 현재 경로를 넣어주고,
        // 현재 경로가 기존 `lexicographically`보다 값이 더 작다면, 그 더 작은 값으로 최적값을 변경
        if *lexicographicallyBest == "" || backtrackingPath < *lexicographicallyBest {
            *lexicographicallyBest = backtrackingPath
        }
        return
    }

    dp[currentCell] = backtrackingPath

    // 현재 Cell로부터 가능한 모든 방향을 탐색
    for _, direction := range directions {
        newRow, newColumn := currRow+direction.toRow, currColumn+direction.toColumn
        if newRow >= 1 && newRow <= rowLimit && newColumn >= 1 && newColumn <= columnLimit {
            // dfs 안에서 체크하게 될 경로를 추가한다
            backtrackingPath = backtrackingPath + direction.keyword
            // `*lexicographicallyBest == ""`: 아직 최적 값이 나오지 않은 경우나
            // `backtrackingPath <= *lexicographicallyBest` 이번에 체크할 경로가 최적값보다 더 작은 경우
            if *lexicographicallyBest == "" || backtrackingPath <= *lexicographicallyBest {
                dpdfs(rowLimit, columnLimit, newRow, newColumn, dstRow, dstColumn, mvCnt-1, dp, backtrackingPath, lexicographicallyBest)
            }
            backtrackingPath = backtrackingPath[:len(backtrackingPath)-1]
        }
    }
}

func solutionDPDFS(n, m, x, y, r, c, k int) string {
    // 각 `Cell`의 최적 경로를 저장하기 위한 DP table(Memoization)
    dp := make(map[Cell]string)
    // 현재 DFS 함수가 탐색 중인 경로를 나타낸다
    var backtrackingPath string
    // `lexicographicallyBest`는 `dp`(map[Cell]string)으로, 특정 `Cell`에 도달할 수 있는 최적의 경로를 문자열로 저장
    // 사전 순으로 가장 빠른 경로를 찾기 위해 사용
    var lexicographicallyBest string

    dpdfs(n, m, x, y, r, c, k, dp, backtrackingPath, &lexicographicallyBest)

    if lexicographicallyBest == "" {
        return "impossible"
    }
    return lexicographicallyBest
}

```

```text
  1 2 3 4
1 . . . .
2 . . S .
3 E . . .
```

(2, 3)에서 시작해서 k=5, (3, 1)이 목표
moves는 left, right, down, up 순서로 정의되어 있다

DFS는 결국 계속 파고 들어간다

1. start, dfs(2, 3, 5)
2. left(2, 3, 5), dfs(2, 3-1, 5-1)
3. left(2, 2, 4), dfs(2, 2-1, 4-1)
4. right(2, 1, 3), dfs(2, 1+1, 3-1)
5. left(2, 2, 2), dfs(2, 2-1, 2-1)
6. down(2, 1, 1), dfs(2+1, 1, 0)
7. (3, 1, 0) == (3, 1, 0) 목적지에 도착 => "" 리턴
8. 그럼 이때 map[(2, 1, 1)]="llrld" 가 되어야 한다

근데 Dynamic Programming을 활용한다면 어떤 상태들을 메모이제이션 또는 tabulation 해야 효율화가 가능할 것이다
가령 (3, 1)까지 찾아가는 과정에서 (2, 3)에서 (2, 1)까지 가는 경로는 아래처럼 다양하다.

ll, ulld, dllu, rullld, rdlllu 등 다양할 수 있다

그렇다면 이때 dfs와 DP를 같이 사용한다고 할 때, 어떤 상태를 저장해야 할까?
단순히 (2, 1)이라는 좌표로만 저장한다면 아래처럼 된다

```go
map[(2, 1)] = []string{"ll", "ulld", "dllu", "rullld", "rdlllu"}
```

근데 (3, 1)에 도달하기까지 (2, 1)을 거치는 경로가 다양할 것이다. 이때 어떤 값을 사용해야 할지 모르게 된다

map[(2, 1, k)]로 한다면?

k = 6이라고한다면, 아래처럼 될 것이다

```go
map[(2, 1, 4)] = []string{"ll"}
map[(2, 1, 2)] = []string{"ulld", "dllu"}
map[(2, 1, 0)] = []string{"rullld", "rdlllu"}
```

`map[(3, 1, 0)]`까지 도달하는 가장 작은 값을 찾아야 하므로, 이 경로까지 찾아간는 데 사용되는 연산에 도움이 되는 값들을 저장해야 한다

그렇다면 경로까지 고려한다면?

```go
map[(2, 1, 4, "ll")] = "ll"
map[(2, 1, 2, "ulld")] = "ulld"
map[(2, 1, 2, "dllu")] = "dllu"
map[(2, 1, 0, "rullld")] = "rullld"
map[(2, 1, 0, "rdlllu")] = "rdlllu"
```

결국 모든 경우의 수를 저장하게 되므로, 최적화의 의미가 없어진다. 따라서 아래처럼 두 개의 경우가 사라져야 최적화에 의미가 있다

```go
map[(2, 1, 4, "ll")] = "ll"
map[(2, 1, 2, "dllu")] = "dllu"
// map[(2, 1, 2, "ulld")] = "ulld"
map[(2, 1, 0, "rdlllu")] = "rdlllu"
// map[(2, 1, 0, "rullld")] = "rullld"
```

그렇다면 이제 상태에 대한 테이블 값을 업데이트 하는 부분이 중요하다.
DP와 DFS의 조합에서 각 상태에 대한 DP 테이블의 값을 어떻게 업데이트해야 할까?

```go
    if k < manhattan || (k-manhattan)%2 != 0 {
        dp[currGrid] = impossible
        return impossible
    }

    if row == rowDst && column == columnDst && k == 0 {
        dp[currGrid] = path
        return path
    }

    ... 생략 ...
    
        if newRow >= 1 && newRow <= rowLimit && column >= 1 && column <= columnLimit && newK >= 0 {
            nextPath := dpdfs(rowLimit, columnLimit, newRow, newColumn, rowDst, columnDst, newK, dp, path+move.keyword)
            dp[currGrid] = min(dp[currGrid], nextPath)
```

- **각 상태의 DP 값**은 **그 상태에서 목적지까지 가는 최단 경로**를 나타내야 한다
- 이 최단 경로가 상**위 호출(즉, 이 상태에 이르게 만든 이전 상태)**에 따라 달라질 수 있다.
