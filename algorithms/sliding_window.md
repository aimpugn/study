# Sliding Window

## 슬라이딩 윈도우?

슬라이딩 윈도우 기법은 배열이나 리스트와 같은 연속된 데이터에서 일정 범위의 요소들을 효율적으로 처리하는 방법이다.
이 기법은 다음과 같은 경우에 유용하다:
- 부분 배열이나 부분 문자열을 처리할 때
- 특히 연속된 데이터 내에서 최대값, 최소값, 평균 등을 찾을 때
- 특정 조건을 만족하는 데이터의 범위를 조사할 때

## 기초: 고정 크기의 슬라이딩 윈도우

고정 크기의 슬라이딩 윈도우는 윈도우(즉, 데이터의 연속된 부분)의 크기가 고정되어 있으며, 이 윈도우가 데이터를 통해 "슬라이드"하면서 문제를 해결한다.

가령 크기가 3인 윈도우를 사용해 배열의 각 부분의 합계를 찾는 문제를 생각해보자.

```plaintext
배열: [1, 3, 5, 2, 8, 10]
윈도우 크기: 3

결과:
[1, 3, 5]의 합 = 9
[3, 5, 2]의 합 = 10
[5, 2, 8]의 합 = 15
...
```

이런 경우 각 슬라이딩 윈도우의 합계를 쉽게 계산할 수 있으며, 배열을 한 번만 순회하면서 연산을 수행할 수 있다.

```go
package main

import (
    "fmt"
)

func findMaxSumSubarray(arr []int, k int) int {
    var maxSum, windowSum int
    for i := 0; i < k; i++ {
        windowSum += arr[i]
    }

    maxSum = windowSum
    for i := k; i < len(arr); i++ {
        windowSum += arr[i] - arr[i-k]
        if windowSum > maxSum {
            maxSum = windowSum
        }
    }
    return maxSum
}

func main() {
    arr := []int{2, 1, 5, 1, 3, 2}
    k := 3
    fmt.Println("Maximum sum of a subarray of size K:", findMaxSumSubarray(arr, k))
}
```

## 중급: 중복 문자 없는 가장 긴 부분 문자열의 길이를 찾기

"중복 문자 없는 가장 긴 부분 문자열의 길이를 찾기"와 같은 문제가 있다.
이 경우, 해시맵을 사용하여 문자의 마지막 위치를 기록하면서 슬라이딩 윈도우를 조정한다.
이렇게 하면 각 문자가 윈도우 내에서 유일한지 빠르게 확인하고, 중복 발생 시 윈도우의 시작점을 조정하여 문제를 효율적으로 해결할 수 있다.

```go
package main

import (
    "fmt"
)

// 문제: 정수 배열과 S가 주어졌을 때, 합이 S 이상이 되는 가장 작은 연속 부분 배열의 길이를 찾으세요.
// 윈도우의 크기를 조건을 만족할 때까지 확장하고, 
// 다시 조건을 만족하지 않을 때까지 축소하는 방식으로 조절합니다.

// minSubArrayLen 함수는 주어진 배열 nums 내에서 합이 s 이상이 되는 최소 길이의 연속된 부분 배열을 찾는다
func minSubArrayLen(s int, nums []int) int {
    // start: 슬라이딩 윈도우의 시작 인덱스
    // sum: 현재 윈도우의 합계
    // minLength: 현재까지 발견된 조건을 만족하는 최소 길이 (초기값은 len(nums)+1로 설정하여, 아직 유효한 윈도우가 발견되지 않았음을 나타냄)
    start, sum, minLength := 0, 0, len(nums)+1

    // 배열을 순회하며, 각 요소를 끝점으로 하는 슬라이딩 윈도우를 조정합니다.
    for end := 0; end < len(nums); end++ {
        sum += nums[end] // 현재 요소를 윈도우의 합계에 추가
        // 현재 윈도우의 합계가 s 이상인 경우, 윈도우의 크기를 가능한 한 줄이면서 조건을 만족시키는지 확인
        for sum >= s {
            if end-start+1 < minLength {
                minLength = end - start + 1 // 새로운 최소 길이를 발견한 경우, 갱신
            }
            sum -= nums[start] // 윈도우의 시작 부분을 제거하여 윈도우 크기를 줄임
            start++
        }
    }

    // 유효한 부분 배열이 발견되지 않은 경우, 0을 반환
    if minLength == len(nums)+1 {
        return 0
    }
    return minLength // 발견된 최소 길이 반환
}

func main() {
    nums := []int{2, 3, 1, 2, 4, 3} // 입력 배열
    s := 7 // 타겟 합계
    fmt.Println("Minimum size subarray sum:", minSubArrayLen(s, nums))
}

```

## 심화: 가변 크기의 슬라이딩 윈도우

가변 크기의 슬라이딩 윈도우는 윈도우의 크기가 고정되어 있지 않고, 문제의 조건을 만족시키기 위해 윈도우의 크기가 동적으로 변한다.

"최소 길이의 부분 배열의 합이 S 이상이 되는 것을 찾아라"와 같은 문제를 해결할 때 사용할 수 있다.

```plaintext
배열: [2, 1, 5, 2, 3, 2], S = 7
결과: [5, 2] 또는 [2, 5] (최소 길이 부분 배열의 합이 7 이상)
```

이 문제는 두 개의 포인터(시작점과 끝점)를 사용하여 동적으로 윈도우의 크기를 조정하면서, 주어진 조건(S 이상의 합)을 만족하는 최소 길이의 부분 배열을 찾아내는 방식이다.
합계가 S 이상이 될 때까지 끝점을 확장하고, S 이상이 되면 시작점을 조정하여 윈도우의 크기를 최소화한다.
시작점과 끝점을 조정하는 로직은 상대적으로 직관적이지만, *윈도우의 크기를 동적으로 관리*한다는 점에서 최적화된 해결책이 요구된다.

이는 특정 알고리즘 문제에서 시간 복잡도나 공간 복잡도를 줄이기 위해 사용된다.