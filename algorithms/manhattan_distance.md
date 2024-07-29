# Manhattan Distance

## 맨해튼 거리?

- 맨해튼 거리는 **격자 형태의 그래프**에서 **두 점 사이의 거리**를 측정할 때 사용되는 방법 중 하나
- 이 거리는 두 점을 연결하는 경로가 격자의 선을 따라야 하며, 대각선 이동이 허용되지 않는 경우에 주로 사용된다. 즉, 수평 또는 수직으로만 이동할 수 있다

## 수학적 정의

$A(x_1,y_1)$와 $B(x_2, y_2)$ 사이의 맨해튼 거리는:

$$\text{Manhattan Distance} = |x_1 - x_2| + |y_1 - y_2|$$

예를 들어 A(1, 1)과 B(4, 5)라는 두 점이 있다면, 맨해튼 거리는 |4 - 1| + |5 - 1| = 3 + 5 = 7이 된다

## 응용 분야

- 로봇 경로 계획: 로봇이 격자 형태의 환경에서 목표 지점까지 이동할 때 사용될 수 있다
- 게임 개발: 격자 형태의 맵에서 캐릭터가 이동할 때 최단 경로를 계산하는 데 사용될 수 있다
- 도시 거리 측정: 도로가 수직, 수평으로만 구성된 도시에서 두 지점 사이의 거리를 측정하는 데 사용될 수 있다
- 데이터 마이닝: 클러스터링, 분류 등에서 다차원 공간에서의 거리 측정 방법으로 사용될 수 있다

이러한 맨해튼 거리는 특정 문제 상황에 따라 유용하게 쓰이며, 맨해튼 거리를 사용하여 가지치기(pruning)를 수행할 수 있다

## 예제

```go
func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

manhattan := abs(row-rowDst) + abs(column-columnDst)
// `k < manhattan`: 남은 이동 횟수(k)가 맨해튼 거리보다 작은 경우, 목표에 도달할 수 없다
// `(k-manhattan)%2 != 0`: 남은 이동 횟수(k)와 맨해튼 거리의 짝수/홀수성이 다른 경우, 목표에 정확히 도달할 수 없다
if k < manhattan || (k-manhattan)%2 != 0 {
    dp[currGrid] = impossible
    return impossible
}
```

```py
# `remain`: 남은 이동 횟수
# `shortest_path`: 현재 위치에서 목표까지의 맨해튼 거리
remain, shortest_path = k - len(path), abs(x_pos - r) + abs(y_pos - c)

# `remain < shortest_path`: 남은 이동 횟수가 맨해튼 거리보다 작은 경우, 목표에 도달할 수 없다
# `remain % 2 != shortest_path % 2`: 남은 이동 횟수와 맨해튼 거리의 짝수/홀수성이 다른 경우, 목표에 정확히 도달할 수 없다
if remain < shortest_path or remain % 2 != shortest_path % 2:
    continue
```

둘은 같은 동작을 한다. 이를 수식으로 표현하면 아래와 같을 것이다.

```mathematica
remain % 2 != shortest_path % 2
-> remain % 2 - shortest_path % 2 != 0
-> (remain - shortest_path) % 2 != 0
```
