# Dynamic Progrmming

- [Dynamic Progrmming](#dynamic-progrmming)
    - [Dynamic Programming](#dynamic-programming)
    - [다이나믹 프로그래밍의 기본 원리](#다이나믹-프로그래밍의-기본-원리)
    - [구현 방식](#구현-방식)
        - [**탑다운 방식 (Top-Down)**: `Memoization`](#탑다운-방식-top-down-memoization)
        - [**바텀업 방식 (Bottom-Up)**: 테이블 저장](#바텀업-방식-bottom-up-테이블-저장)
    - [최적해 보장 원리](#최적해-보장-원리)
    - [동적 프로그래밍 문제인지 판별하기](#동적-프로그래밍-문제인지-판별하기)
    - [작은 문제로 나누기와 점화식](#작은-문제로-나누기와-점화식)
        - [최적 부분 구조](#최적-부분-구조)
    - [테이블 설계하기](#테이블-설계하기)
        - [Bottom-up DP](#bottom-up-dp)
        - [Top-down DP](#top-down-dp)
    - [예제](#예제)
        - [피보나치 수열](#피보나치-수열)
        - [Knapsack 문제](#knapsack-문제)
            - [1. 문제 이해 및 분석](#1-문제-이해-및-분석)
            - [2. 추상화와 모델링](#2-추상화와-모델링)
            - [3. 알고리즘 설계](#3-알고리즘-설계)
            - [4. 알고리즘 구현](#4-알고리즘-구현)
        - [미로 탈출](#미로-탈출)
        - [미로 탈출 문제 DP + DFS](#미로-탈출-문제-dp--dfs)

## Dynamic Programming

동적 프로그래밍이란?
1. *큰 문제의 해답에 작은 문제의 해답이 포함*되어 있고,

    > 이를 최적 부분 구조(Optimal Substructure)를 가졌다고 합니다.

2. 이를 *재귀호출 알고리즘으로 구현*하면 지나친 중복이 발생하는 경우에
3. 이 재귀적 중복을 해결하는 방법입니다.

최적해를 구하는 효율적인 알고리즘 기법 중 하나입니다.

예를 들어, 임의 그래프의 두 정점 s, t 간의 최단 경로 s -> t에서,
t에 이르기 직전에 방문된 정점을 x(s -> x -> t)라고 가정합니다.
s와 t 간의 최단 경로는 s와 x 간의 최단 경로를 포함합니다.
s에서 x까지 최단 경로가 아니라면, s에서 t까지도 최단 경로가 될 수 없기 때문입니다.

일반적인 DP 문제에서는 *각 상태의 최적 해가 그 상태에 이르게 만든 이전 상태에 독립적*입니다.
DP는 순차적인 결정 과정을 모델링하는 데 적합합니다.

예를 들어, `n`번째 단계에서의 최적의 선택은 1부터 `n-1`까지의 선택에만 의존합니다.

## 다이나믹 프로그래밍의 기본 원리

1. **최적 부분 구조 (Optimal Substructure)**:

   문제의 최적해가 하위 문제들의 최적해로부터 만들어질 수 있어야 합니다.

   예를 들어, 최단 경로 문제에서 *전체 경로의 최단 거리는 그 경로를 구성하는 하위 경로들의 최단 거리의 합으로 결정*됩니다.

2. **중복되는 부분 문제 (Overlapping Subproblems)**:

   동일한 작은 문제들이 반복적으로 계산되는 경우가 많습니다.

   예를 들어, 피보나치 수열 계산에서 동일한 피보나치 수를 여러 번 계산하게 됩니다.

## 구현 방식

### **탑다운 방식 (Top-Down)**: `Memoization`

*재귀적 접근법*을 사용하여 문제를 해결합니다.
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

### **바텀업 방식 (Bottom-Up)**: 테이블 저장

*반복문을 사용*하여 작은 문제부터 차례대로 해결해 나갑니다.
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

### 최적 부분 구조

DP로 풀 수 있는 문제의 요건은 다음과 같습니다.
- 전체 문제를 부분 문제로 나눌 수 있어야 합니다.
- 나누어진 부분 문제의 최적해가 전체 문제의 최적해를 구성할 수 있어야 합니다.

부분 문제, 전체 문제, 그리고 둘 사이의 최적 부분 구조를 정리하면 다음과 같습니다.
- **부분 문제**란 전체 문제를 해결하기 위해 나눈 작은 단위의 문제를 가리킵니다.
    - 원래 문제와 유사한 형태를 가집니다.
    - 독립적으로 해결이 가능하면서 전체 문제 해결에 기여합니다.

        가령 피보나치 수열에서 `F(n-1)`, `F(n-2)`는 `F(n)`의 부분 문제입니다.

- **전체 문제**란 그 작은 **부분 문제**들이 합쳐져서 궁극적으로 달성하려는 목표입니다.
    - 여러 부분 문제로 나눌 수 있습니다.
    - 부분 문제들의 해를 조합하여 해결할 수 있습니다.

- **최적 부분 구조**란 부분 문제들의 최적해가 전체 문제의 최적해를 구성하는 성질을 의미합니다.
    - "즉, 작은 문제에서 최적의 해를 구했다면, 그 결과를 바탕으로 전체 문제의 최적해를 구할 수 있는가?"에 대답할 수 있어야 합니다.
    - 최단 경로 문제에서 A에서 C로 가는 최단 경로가 B를 거친다면, A에서 B까지의 경로와 B에서 C까지의 경로도 각각 최단 경로입니다.

## 테이블 설계하기

테이블 설계 과정은 다음과 같습니다:
1. 문제의 구조를 이해하고 [*최적 부분 구조*](#최적-부분-구조)를 식별합니다.
2. 상태 정의: 문제를 해결하는 데 필요한 정보를 상태로 정의합니다.
3. 점화식 도출: 현재 상태와 이전 상태 간의 관계를 수식으로 표현합니다.
4. 기저 상태 설정: 가장 작은 하위 문제의 해를 정의합니다.
5. 계산 순서 결정: 하위 문제부터 상위 문제로 진행하는 순서를 결정합니다.

피보나치 수열을 예로 들어 DP 테이블 설계 과정을 살펴보겠습니다.

1. 문제 분석: F(n) = F(n-1) + F(n-2)라는 구조를 가집니다.
2. 상태 정의: DP[i] = i번째 피보나치 수
3. 점화식: DP[i] = DP[i-1] + DP[i-2]
4. 기저 상태: DP[0] = 0, DP[1] = 1
5. 계산 순서: 0부터 n까지 순차적으로 계산

### Bottom-up DP

### Top-down DP

## 예제

### 피보나치 수열

문제: n번째 피보나치 수를 구하라.

- 문제 분석:
    주어진 문제가 어떤 작은 부분으로 나눌 수 있는지 생각합니다.
    피보나치 수열은 이전 두 수의 합으로 다음 수가 결정됩니다.

- 부분 문제 정의:
    큰 문제를 해결하기 위해 필요한 작은 문제를 정의합니다.

    $F(i) = i\text{번째 피보나치 수}$

- 관계식(점화식) 도출:
    부분 문제들 사이의 관계를 수학적으로 표현합니다.

    $F(i) = F(i-1) + F(i-2) \;\; (i ≥ 2)$

- 기저 조건:
    가장 작은 부분 문제의 해답을 정의합니다.

    $F(0) = 0, F(1) = 1$

- 테이블 설계:
    부분 문제의 해답을 저장할 테이블을 설계합니다.

    ```java
    DP[i] = i번째 피보나치 수를 저장할 배열
    ```

- 테이블 채우기:
    작은 문제부터 큰 문제까지 테이블을 채워나갑니다.

    ```java
    for i from 2 to n:
        DP[i] = DP[i-1] + DP[i-2]
    ```

- 최종 해답:
    테이블의 마지막 값이나 특정 값을 통해 전체 문제의 해답을 얻습니다.

    `DP[n]` = n번째 피보나치 수

```java
public int fibonacci(int n) {
    if (n <= 1) return n;
    int[] dp = new int[n + 1];
    dp[0] = 0;
    dp[1] = 1;
    for (int i = 2; i <= n; i++) {
        dp[i] = dp[i-1] + dp[i-2];
    }
    return dp[n];
}
```

이 예시에서 DP[i]는 i번째 피보나치 수를 저장하고 있으며, 이를 재사용하여 더 큰 수의 피보나치 수를 계산합니다.

### Knapsack 문제

주어진 물건들의 가치와 무게, 그리고 배낭의 무게 한도가 있을 때,
배낭에 넣을 수 있는 물건들의 가치의 합을 최대화하는 조합 최적화의 문제입니다.

- **0/1 Knapsack**

    아이템을 통째로 넣거나 아예 뺍니다.
    즉, 부분적으로 넣을 수 없습니다.

- **Fractional Knapsack**

    아이템을 부분적으로 넣을 수 있습니다.

다이나믹 프로그래밍 접근법은 0/1 Knapsack 문제 해결 방법에 사용됩니다.
2차원 테이블을 사용하여 '`i`-행, `j`-열의 항목 = i개의 아이템'을 고려했을 때,
최대 무게가 `j`인 Knapsack에 넣을 수 있는 최대 가치를 찾습니다.

중복 계산을 피함으로써 문제 해결 과정을 최적화하는 데 유리합니다.

#### 1. 문제 이해 및 분석

```plaintext
여행 가방에 물건을 담으려 합니다.
각 물건에는 무게와 가치가 있고, 가방에는 무게 제한이 있습니다.
가방에 담을 수 있는 물건들의 가치 합계를 최대화하고 싶습니다.
단, 각 물건은 온전히 담거나 아예 담지 않아야 합니다.
```

- 이 문제의 핵심

    가방에 물건을 담는 건데 무게 제한이 있습니다.
    그리고 각 물건에는 가치가 있습니다.
    결국 가치를 최대화하는 게 목표입니다.

    '온전히 담거나 아예 담지 않아야 한다'는 조건은 물건을 분할할 수 없음을 의미합니다.

- 예시

    - 가방 무게 제한: 10kg
    - 물건 목록:
        - 책: 4kg, 가치 5
        - 노트북: 3kg, 가치 7
        - 카메라: 2kg, 가치 4
        - 옷: 5kg, 가치 6

    가장 비싼 것을 먼저 담을지, 아니면 가장 가벼운 것을 먼저 담을지를 생각하게 되기 쉽습니다.
    하지만 그러한 우선순위로는 가치 최대화를 보장할 수 없습니다.
    '가치를 최대화'하기 위해서는 가능한 모든 조합의 가치를 알아야 합니다.

#### 2. 추상화와 모델링

1. 문제의 일반화

    먼저 주어진 조건들을 추상화하여 일반화된 형태로 표현합니다.

    ```java
    n: 물건의 총 개수
    W: 가방의 무게 제한
    각 물건 i (i = 1, 2, ..., n):
        w_i: i번째 물건의 무게
        v_i: i번째 물건의 가치
    ```

    예: (책, 4kg, 5), (노트북, 3kg, 7), ...

    각 물건을 선택할지 말지를 결정해야 하므로, 선택 여부를 나타내는 변수를 도입합니다:

    ```java
    x_i: i번째 물건의 선택 여부 (선택하면 1, 선택하지 않으면 0)
    ```

2. 목표와 제약 조건 정의

    - 목표: 선택한 물건들의 가치 합을 최대화하는 것입니다.
    - 제약 조건: 선택한 물건들의 무게 합이 가방의 무게 제한을 초과하지 않아야 합니다.

3. 수학적 모델링

    - 목적 함수 (최대화하고자 하는 값):

        ```java
        (선택1 * v_i) + (선택2 * 가치2) + ... + (선택n * 가치n)
        ```

        최대화: 선택한 물건들의 가치 합을 나타냅니다.

        $\sum_{i=1}^{n}(x_i \times v_i)$

    - 제약 조건:

        ```java
        (선택1 * 무게1) + (선택2 * 무게2) + ... + (선택n * 무게n) ≤ 가방 무게 제한 (W)
        ```

        $\sum_{i=1}^{n}(x_i \times w_i) \leq W$

        $선택i \in \{0, 1\} \quad \forall.i \in \{1, 2, ..., n\}$

   첫 번째 조건은 선택한 물건들의 무게 합이 가방의 무게 제한을 초과하지 않아야 함을 나타냅니다.
   두 번째 조건은 각 물건을 선택하거나(1) 선택하지 않아야(0) 함을 나타냅니다.

4. 문제 해결 접근법

    가능한 모든 조합을 시도해보는 완전 탐색 방법은 $2^n$의 시간 복잡도를 가지므로, 물건의 수가 많아지면 비효율적입니다.
    이 경우 동적 프로그래밍 접근법을 사용하여 문제를 더 효율적으로 해결할 수 있습니다.

    동적 프로그래밍 접근법의 핵심 아이디어는 문제를 더 작은 부분 문제로 나누고, 이 부분 문제들의 해를 이용하여 전체 문제의 해를 구하는 것입니다.

    - 부분 문제 정의

        이 문제를 작은 문제로 나누면 다음과 같습니다:
        - 첫 번째 물건까지만 고려했을 때의 최적해
        - 두 번째 물건까지 고려했을 때의 최적해
        - N 번째 물건까지 고려했을 때의 최적해

        $K(i, w)$ = `i`번째 물건까지 고려했을 때, 무게(`w`) 이하에서 얻을 수 있는 최대 가치
        - `i`: 현재까지 고려한 물건의 수
        - `w`: 현재 고려 중인 무게 제한

    - 부분 문제 간의 관계

        $K(i, w)$를 구하는 방법은 다음 두 가지 경우로 나눌 수 있습니다:

        - i번째 물건을 선택하지 않는 경우:
            이 경우, `i-1`번째 물건까지 고려했을 때의 최적해와 같습니다.

            $K(i, w) = K(i - 1, w)$

        - i번째 물건을 선택하는 경우:
            이 경우, `i`번째 물건의 가치를 더하고, 남은 무게에 대해 `i-1`번째 물건까지 고려했을 때의 최적해를 더합니다.

            $K(i, w) = v_i + K(i - 1, w - w_i)$

            단, 이 경우는 $w_i \leq w$일 때만 가능합니다.

   따라서, K(i, w)의 최종 식은 다음과 같습니다:

    $$
    K(i, w)=
    \begin{cases}
    \max(\;
        K(i-1, w),\;
        v_i + K(i-1, w - w_i)
    \;), & w_i \leq w \\
    K(i-1, w), & w_i > w
    \end{cases}
    $$

   이 관계를 2차원 표로 나타낼 수 있습니다. 표의 각 칸 (i, w)는 K(i, w)의 값을 나타냅니다.

    ```java
    입력:
    최대 용량 한도 = 7
    무게들 = [3, 4, 5]
    가치들 = [4, 5, 6]
    물건_수 = 3

    DP 테이블 채우기 과정:
    - i: 물건 개수
    - w: 무게
    - 각 셀: 물건 개수별 무게별 가치

      w → 0   1   2   3   4   5   6   7
      i
      ↓ ┌───┬───┬───┬───┬───┬───┬───┬───┐
      0 │ 0 │ 0 │ 0 │ 0 │ 0 │ 0 │ 0 │ 0 │
        ├───┼───┼───┼───┼───┼───┼───┼───┤
      1 │ 0 │ 0 │ 0 │ 4 │ 4 │ 4 │ 4 │ 4 │
        ├───┼───┼───┼───┼───┼───┼───┼───┤
      2 │ 0 │ 0 │ 0 │ 4 │ 5 │ 5 │ 5 │ 9 │
        ├───┼───┼───┼───┼───┼───┼───┼───┤
      3 │ 0 │ 0 │ 0 │ 4 │ 5 │ 6 │ 6 │ 9 │
        └───┴───┴───┴───┴───┴───┴───┴───┘

    DP[i][w] 계산 예시 (i=2, w=7):
    max(가치들[1] + DP[1][7-무게들[1]], DP[1][7])
    = max(5 + DP[1][3], DP[1][7])
    = max(5 + 4, 4) = max(9, 4) = 9

    선택된 물건 추적:
    1. DP[3][7] != DP[2][7]이므로 물건 3 선택하지 않음
    2. DP[2][7] != DP[1][7]이므로 물건 2 선택 (w = 7 - 4 = 3)
    3. DP[1][3] != DP[0][3]이므로 물건 1 선택

    최종 결과:
    최대 가치: 9
    선택된 물건: [0, 1] (첫 번째와 두 번째 물건)

    ```

   - 0번째 행은 모두 0으로 초기화 (물건을 하나도 선택하지 않은 경우)
   - 0번째 열도 모두 0으로 초기화 (무게 제한이 0인 경우)
   - 나머지 칸들은 위에서 정의한 관계식에 따라 채워집니다.

   표를 채우는 과정에서 각 칸의 값은 해당 부분 문제의 최적해를 나타내며,
   최종적으로 $K(n, W)$가 전체 문제의 최적해(최대 가치)가 됩니다.

#### 3. 알고리즘 설계

1. 알고리즘의 주요 단계

    - DP 테이블 초기화
    - DP 테이블 채우기
    - 최적해(최대 가치) 찾기
    - 선택된 물건 추적 (선택 사항)

2. DP 테이블 채우기(최대 가치 채워넣기)

    ```java
    for 물건_인덱스 = 1 to 물건_수:
        for 현재_무게_제한 = 1 to 최대_용량:
            // 현재 순서의 물건의 무게와 현재 고려중인 무게를 비교합니다.
            // 1. 무게가 현재 용량 이하라서 물건을 선택할 수 있는 경우, 둘 중 더 좋은 것을 선택합니다:
            //     - 이 물건을 선택하지 않는 경우
            //     - 이 물건을 선택하는 경우
            // 2. 무게가 현재 용량을 초과해서 물건을 선택할 수 없는 경우: 이 물건을 무시하고 이전 상태를 그대로 유지해야 합니다.
            //
            // 1차원 배열로 최적화한 경우에는 다음과 같습니다:
            // dp[w] = max(
            //     dp[w], // 이 물건을 선택하지 않는 경우
            //     v_i + dp[w - w_i] // 이 물건을 선택하는 경우.
            //     // w: 현재 고려하고 있는 무게 제한,
            //     // w_i: 현재 물건의 무게
            // )

            // i번째 물건을 넣을 수 있는 경우
            if (물건들[물건_인덱스 - 1].무게 <= 현재_무게_제한):
                // 배낭에는 물건인덱스, 현재_무게_제한 공간에 최대 가치가 저장됩니다.
                배낭[물건_인덱스][현재_무게_제한] = max(
                    // 이 물건을 선택하지 않는 경우 이전 가치를 그대로 사용합니다.
                    배낭[물건_인덱스 - 1][현재_무게_제한],
                    // 이 물건을 선택하는 경우 선택했을 때의 가치를 계산합니다.
                    물건들[물건_인덱스].가치 + 배낭[물건_인덱스 - 1][현재_무게_제한 - 물건들[물건_인덱스].무게]
                )
            else: // i번째 물건을 넣을 수 없는 경우
                배낭[물건_인덱스][현재_무게_제한] = 배낭[물건_인덱스 - 1][현재_무게_제한]
    ```

3. 선택된 물건 추적(어떤 물건들을 선택했는지):

    최대 가치를 찾은 후, 어떤 물건들이 선택되었는지 추적할 수 있습니다:

    ```java
    물건_인덱스, 현재_무게_제한 = 개수_n, 최대_무게_W
    선택된것들 = []

    while 물건_인덱스 > 0 && 현재_무게_제한 > 0
        // '현재 무게에서의 최대 가치'와 '이전 물건까지의 최대 가치'를 비교하여 그 결과가:
        // 1. 다르다면? 이는 현재 물건을 선택했다는 의미입니다.
        // 2. 같다면? 현재 물건을 선택하지 않은 것입니다.
        if 배낭[물건_인덱스][현재_무게_제한] != 배낭[물건_인덱스 - 1][현재_무게_제한]:
            선택된것들.append(i - 1)
            현재_무게_제한 -= 무게들[물건_인덱스 - 1]
        물건_인덱스 -= 1

    return 선택된_물건들.reverse()  // 물건들을 선택한 순서대로 정렬
    ```

4. 시/공간 복잡도 분석

    표의 크기가 $(n+1) \times (W+1)$이고, 각 칸을 채우는 데 상수 시간이 걸립니다.
    - 시간복잡도: $O(nW)$

        DP 테이블의 각 셀을 한 번씩 채우며, 각 셀을 채우는 데 상수 시간이 걸립니다.

    - 공간복잡도: $O(nW)$

        n+1 행, W+1 열의 2차원 배열을 사용합니다.

    - 공간복잡도 최적화

        각 단계에서 이전 행의 정보만 필요하므로,
        1차원 배열을 사용하여 공간 복잡도를 $O(W)$로 줄일 수 있습니다:

        ```python
        K = [0 for w in range(W + 1)]

        for i in range(1, n + 1):
            for w in range(W, weights[i - 1] - 1, -1):
                K[w] = max(K[w], values[i - 1] + K[w - weights[i-1]])

        return K[W]
        ```

#### 4. 알고리즘 구현

```java
/**
 * 0-1 Knapsack 문제를 해결하는 클래스입니다.
 * 이 클래스는 동적 프로그래밍을 사용하여 주어진 무게 제한 내에서
 * 최대 가치를 가지는 물건들의 조합을 찾습니다.
 */
public class Knapsack01 {

    /**
     * 0-1 Knapsack 문제를 해결하는 메인 메서드입니다.
     * 이 메서드는 동적 프로그래밍 접근법을 사용하여 최적의 해를 찾습니다.
     *
     * @param W 배낭의 최대 무게 제한
     * @param weights 각 물건의 무게 배열
     * @param values 각 물건의 가치 배열
     * @param n 물건의 개수
     * @return 최대 가치와 선택된 물건들의 인덱스를 포함하는 Result 객체
     */
    public static Result solve(int W, int[] weights, int[] values, int n) {
        // 입력 유효성 검사
        if (
            W <= 0
            || n <= 0
            || weights == null
            || values == null
            || weights.length != n
            || values.length != n
        ) {
            throw new IllegalArgumentException("Invalid input parameters");
        }

        // DP 테이블 초기화 (공간 최적화를 위해 1차원 배열 사용)
        // dp[w]는 무게 w까지 사용했을 때 얻을 수 있는 최대 가치를 저장합니다.
        int[] dp = new int[W + 1];

        /*
        * DP 테이블 채우기
        * 각 물건에 대해 순차적으로 처리합니다.
        *
        * DP 테이블 갱신 과정
        *  w →
        * i 0 1 2 3 4 5 ... W
        * ↓ ┌─┬─┬─┬─┬─┬─┬───┐
        * 0 │0│0│0│0│0│0│...│
        *   ├─┼─┼─┼─┼─┼─┼───┤
        * 1 │0│ │ │ │ │ │...│  ← 첫 번째 물건 처리 후
        *   ├─┼─┼─┼─┼─┼─┼───┤
        * 2 │0│ │ │ │ │ │...│  ← 두 번째 물건 처리 후
        *   ├─┼─┼─┼─┼─┼─┼───┤
        * . │.│ │ │ │ │ │...│
        * . │.│ │ │ │ │ │...│
        *   ├─┼─┼─┼─┼─┼─┼───┤
        * n │0│ │ │ │ │ │...│  ← 마지막 물건 처리 후
        *   └─┴─┴─┴─┴─┴─┴───┘
        */
        for (int i = 1; i <= n; i++) {
            // 각 단계에서 이전 단계(i-1)의 결과를 사용하여 현재 단계(i)의 결과를 계산합니다.
            // 현재 물건의 무게와 가치
            int currentWeight = weights[i - 1];
            int currentValue = values[i - 1];

            // 정방향으로 순회하면, 현재 단계의 계산이 이전 단계의 값을 덮어쓰게 되어 잘못된 결과를 얻게 됩니다:
            // 물건: 무게 = 2, 가치 = 3
            //
            // 초기 상태:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   0   1   1   2
            //
            // 1. 가방_무게 w = 2일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   3   1   1   2
            //                 ↑
            //                 │ 원래 0이었던 값을 3으로 업데이트
            //                 │
            //                 max(0, 3 + dp[가방_무게(2) - 현재_무게(2)]) = 3
            //
            // 2. 가방_무게 w = 3일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   3   3   1   2
            //                     ↑
            //                     │ 원래 1이었던 값을 3으로 업데이트
            //                     │
            //                     max(1, 3 + dp[가방_무게(3) - 현재_무게(2)]) = 3
            //
            // 3. 가방_무게 w = 4일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   3   3   6   2
            //                         ↑
            //                         │ 원래 1이었던 값을 6으로 업데이트
            //                         │
            //                         max(1, 3 + dp[가방_무게(4) - 현재_무게(2)]) = 6
            //                         초기 상태 값이 0이었다가 현재 단계에서는 수정된 dp[2]의 값을 사용하여 잘못된 계산이 이뤄집니다.
            //
            // 4. 가방_무게 w = 5일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   3   3   6   6
            //                             ↑
            //                             │ 원래 2였던 값을 6으로 업데이트
            //                             │
            //                             max(2, 3 + dp[가방_무게(5) - 현재_무게(2)]) = 6
            //                             초기 상태 값이 1이었다가 현재 단계에서는 수정된 dp[3]의 값을 사용하여 잘못된 계산이 이뤄집니다.
            //
            // 반면 역순으로 순회하여 이전 단계의 값을 덮어쓰지 않습니다.
            // 물건: 무게 = 2, 가치 = 3
            //
            // 초기 상태:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   0   1   1   2
            //
            // 1. 가방_무게 w = 5일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   0   1   1   4
            //                             ↑
            //                             │ 원래 2였던 값을 4로 업데이트
            //                             │
            //                             max(2, 3 + dp[가방_무게(5) - 현재_무게(2)])
            //                             = max(2, 3 + dp[3])
            //                             = max(2, 3 + 1) = 4
            //
            // 2. 가방_무게 w = 4일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   0   1   3   4
            //                         ↑
            //                         │ 원래 1이었던 값을 3으로 업데이트
            //                         │
            //                         max(1, 3 + dp[가방_무게(4) - 현재_무게(2)])
            //                         = max(2, 3 + dp[2])
            //                         = max(2, 3 + 0) = 3
            //
            // 3. 가방_무게 w = 3일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   0   3   3   4
            //                     ↑
            //                     │ 원래 1이었던 값을 3으로 업데이트
            //                     │
            //                     max(1, 3 + dp[가방_무게(3) - 현재_무게(2)])
            //                     = max(2, 3 + dp[1])
            //                     = max(2, 3 + 0) = 3
            //
            // 4. 가방_무게 w = 2일 때:
            // Index:  0   1   2   3   4   5 ← 가방_무게
            // Value:  0   0   3   3   3   4
            //                 ↑
            //                 │ 원래 0이었던 값을 3으로 업데이트
            //                 │
            //                 max(0, 3 + dp[가방_무게(2) - 현재_무게(2)])
            //                 = max(2, 3 + dp[0])
            //                 = max(2, 3 + 0) = 3
            for (int w = W; w >= currentWeight; w--) {
                // 현재 물건을 선택했을 때의 가치와 선택하지 않았을 때의 가치 중 큰 값 선택
                // - dp[w]: 현재 물건을 선택하지 않았을 때의 가치
                // - currentValue + dp[w - currentWeight]: 현재 물건을 선택했을 때의 가치
                dp[w] = Math.max(dp[w], currentValue + dp[w - currentWeight]);
            }
        }

        // // 2차원 배열 사용하는 경우
        // int[][] dp = new int[n + 1][W + 1];
        //
        // for (int i = 1; i <= n; i++) {
        //     int currentWeight = weights[i-1];
        //     int currentValue = values[i-1];
        //
        //     for (int w = 0; w <= W; w++) {
        //         if (currentWeight <= w) {
        //             // 현재 물건을 선택할 수 있는 경우
        //             dp[i][w] = Math.max(dp[i-1][w], currentValue + dp[i-1][w - currentWeight]);
        //         } else {
        //             // 현재 물건을 선택할 수 없는 경우
        //             dp[i][w] = dp[i-1][w];
        //         }
        //     }
        // }

        /*
         * 선택된 물건들 추적하기
         * - n: 물건의 개수
         * - w: 추적중인 무게 한도
         *
         * n ───┐
         *      │
         * .    │ 역순으로
         * .    │ 추적
         * 2    │
         * 1 ───┘
         *  w → W..0
         */
        boolean[] selected = new boolean[n];
        int w = W;
        // 마지막 물건부터 시작하여 어떤 물건들이 선택되었는지 역추적합니다.
        // i: 물건의 인덱스
        for (int i = n; i > 0 && w > 0; i--) {
            // 현재 물건의 무게
            int currentWeight = weights[i - 1];
            int nextWeight = w - currentWeight;

            if (
            // 다음 무게가 끝인 경우를 의미합니다.
            // 즉, dp[0]은 아무 무게도 선택하지 않았음을 의미하므로, 현재가 끝입니다.
            nextWeight == 0
                    // 차감한 무게가 다르다면 선택했음을 의미합니다.
                    || ((dp[w] - dp[nextWeight]) != dp[w])) {
                selected[i - 1] = true; // 현재 물건 선택
                w -= currentWeight; // 남은 무게 갱신
            }
        }

        // 최종 결과 반환
        return new Result(dp[W], selected);
    }

    /**
     * Knapsack 문제의 결과를 저장하는 내부 클래스입니다.
     */
    public static class Result {
        /** 얻을 수 있는 최대 가치 */
        public final int maxValue;
        /** 각 물건의 선택 여부를 나타내는 불리언 배열 */
        public final boolean[] selectedItems;

        /**
         * Result 객체를 생성합니다.
         * @param maxValue 최대 가치
         * @param selectedItems 선택된 물건들의 배열
         */
        public Result(int maxValue, boolean[] selectedItems) {
            this.maxValue = maxValue;
            this.selectedItems = selectedItems;
        }
    }

    /**
     * 메인 메서드: 예제를 실행하고 결과를 출력합니다.
     */
    public static void main(String[] args) {
        int[] values = {60, 100, 120};
        int[] weights = {10, 20, 30};
        int W = 50;
        int n = values.length;

        Result result = solve(W, weights, values, n);

        System.out.println("최대 가치: " + result.maxValue);
        System.out.println("선택된 물건:");
        for (int i = 0; i < n; i++) {
            if (result.selectedItems[i]) {
                System.out.println("물건 " + (i+1) + ": 무게 = " + weights[i] + ", 가치 = " + values[i]);
            }
        }
    }
}
```

### 미로 탈출

[미로 탈출 문제](./examples/bygo/kakao/find_maze_exit/README.md)

- 목표: 시작 위치 (x, y)에서 목표 위치 (r, c)까지 이동하는 데 필요한 최소 이동 경로를 찾는 것
- 최대 `k`번의 이동이 가능
- 이동 방향은 위(`u`), 아래(`d`), 왼쪽(`l`), 오른쪽(`r`) 중 하나

```go
// 어떤 상태
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
    // `lexicographicallyBest`는 `dp`(map[Cell]string)으로, 특정 `Cell`에 도달할 수 있는 최적의 경로를 문자열로 저장
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

(2, 3)에서 시작해서 k=5, (3, 1)이 목표입니다.
moves는 left, right, down, up 순서로 정의되어 있습니다.

DFS의 순서를 따라가보면 다음과 같습니다:
1. start, dfs(2, 3, 5)
2. left(2, 3, 5), dfs(2, 3-1, 5-1)
3. left(2, 2, 4), dfs(2, 2-1, 4-1)
4. right(2, 1, 3), dfs(2, 1+1, 3-1)
5. left(2, 2, 2), dfs(2, 2-1, 2-1)
6. down(2, 1, 1), dfs(2+1, 1, 0)
7. (3, 1, 0) == (3, 1, 0) 목적지에 도착 => "" 리턴
8. 그럼 이때 map[(2, 1, 1)]="llrld" 가 되어야 합니다.

DP를 활용한다면 Memoization 또는 Tabulation 위해 DP 테이블 설계가 필요합니다.
가령 (3, 1)까지 찾아가는 과정에서 (2, 3)에서 (2, 1)까지 가는 경로는 아래처럼 다양합니다.
- ll
- ulld
- dllu
- rullld
- rdlllu 등

그렇다면 이때 dfs와 DP를 같이 사용한다고 할 때, 어떤 상태를 저장해야 할까?
단순히 (2, 1)이라는 좌표로만 저장한다면 아래처럼 된다

```go
map[(2, 1)] = []string{"ll", "ulld", "dllu", "rullld", "rdlllu"}
```

근데 (3, 1)에 도달하기까지 (2, 1)을 거치는 경로가 다양하기 때문에 어떤 값을 사용해야 할지 모르게 됩니다.
이를 알기 위해 `map[(2, 1, k)]`로 한다면 아래처럼 됩니다.

```go
// k = 6
map[(2, 1, 4)] = []string{"ll"}
map[(2, 1, 2)] = []string{"ulld", "dllu"}
map[(2, 1, 0)] = []string{"rullld", "rdlllu"}
```

`map[(3, 1, 0)]`까지 도달하는 가장 작은 값을 찾아야 하므로, 이 경로까지 찾아간는 데 사용되는 연산에 도움이 되는 값들을 저장해야 합니다.

만약 경로까지 고려하게 된다면 결국 모든 경우의 수를 저장하게 되므로, 최적화의 의미가 없어집니다.

```go
map[(2, 1, 4, "ll")] = "ll"
map[(2, 1, 2, "ulld")] = "ulld"
map[(2, 1, 2, "dllu")] = "dllu"
map[(2, 1, 0, "rullld")] = "rullld"
map[(2, 1, 0, "rdlllu")] = "rdlllu"

// 아래처럼 두 개의 중복되는 경우가 사라져야 최적화에 의미가 있습니다.
map[(2, 1, 4, "ll")] = "ll"
map[(2, 1, 2, "dllu")] = "dllu"
// map[(2, 1, 2, "ulld")] = "ulld"
map[(2, 1, 0, "rdlllu")] = "rdlllu"
// map[(2, 1, 0, "rullld")] = "rullld"
```
