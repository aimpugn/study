# process

```java
/**
 * - if 야근, then 피로도 쌓임
 * - 피로도 = 야근 시작 시점에서 남은 일의 작업량을 제곱하여 더한 값
 * - 1시간 동안 작업량 1 처리 가능
 *
 * 주어진 n을 주어진 배열의 요소에서 적절히 차감하여 제곱한 결과 최소값을 찾습니다.
 * ex:
 * - works: [4, 3, 3]
 *   n: 4
 *   => [4 -2, 3 -1, 3 - 1]
 *   => [2, 2, 2]
 *   => 4 + 4 + 4 = 12
 *
 * - works: [2, 1, 2]
 *   n: 1
 *   => [2 - 1, 1, 2]
 *   => 1 + 1 + 4 = 6
 *
 * @param n 퇴근까지 남은 시간
 *          1 <= n <= 1,000,000
 * @param works 각 일에 대한 작업량
 *              1 <= works.length <= 20,000
 *              1 <= work <= 50,000
 * @return 야근 피로도를 최소화한 값
 */
public long solution(int n, int[] works) {
    long answer = 0;

    // n    : 4
    // works: [4, 3, 3]
    // cases:
    // - [0, 3, 3] = 18
    //    ^
    //   -4
    // - [1, 2, 3] = 14
    //    ^  ^
    //   -3 -1
    // - [1, 3, 2] = 14 // [1, 2, 3]과 같지만 결과는 같습니다. memoization
    //    ^  ^  ^
    //   -3  0 -1
    // - [2, 1, 3] = 14 // memoization
    //    ^  ^
    //   -2 -2
    // - [2, 1, 3] = 14 // memoization
    //    ^  ^
    //   -2 -2
    // - [3, 0, 3] = 18
    //
    // 주어진 n을 각 배열의 요소에 어떻게 차감할 것인지


    return answer;
}
```

문제는 여기서 "어떻게" 할지를 떠올려야 합니다.

일단 목표는 *주어진 시간 내에 작업량을 효율적으로 줄여 전체 작업의 제곱합을 최소화하는 것*입니다.
그러면 어떤 자료구조를 왜 선택해서, 어떻게 자료를 탐색하며 어떻게 값이 최소화되도록 할 것인가?

우선 목표에서 중요한 것은 *제곱합을 최소화*하는 것으로 보입니다.
"제곱합"은 지수가 2로 고정이 되어 있고, 제곱합을 줄이는 방법은 밑을 차감하는 것입니다.
이미 작은 값을 차감할 필요는 없고, 그 순간 가장 큰 값을 차감하면 제곱합이 작아질 것으로 예상됩니다.

1. 그렇다면 그 순간 최대값을 효과적으로 판단하기 위해 (최대 힙) 우선순위 큐를 사용합니다.

    매 단계에서 현재 가장 큰 작업량을 선택한다는 점에서 그리디 알고리즘에 해당합니다.

    우선순위 큐(최대 힙)를 사용하는 이유는 다음과 같습니다:
    1. 매 단계에서 최대값을 O(log n) 시간에 탐색 가능
    2. 값을 갱신하고 다시 삽입하는 것도 O(log n) 시간에 가능
    3. 전체 과정이 O(n log m) 시간에 수행됩니다. (n은 주어진 시간, m은 작업의 수)

2. n번 반복하면서 가장 큰 작업량을 차감해 갑니다.
3. 큐에 남아있는 모든 작업량을 제곱하여 합산합니다.

```plaintext
초기 상태:
[4, 3, 3]  (피로도: 4² + 3² + 3² = 34)
  4
 / \
3   3   n = 4

1단계: 4를 선택하여 감소하고 다시 삽입
[3, 3, 3]  (피로도: 3² + 3² + 3² = 27)
  3
 / \
3   3   n = 3

2단계: 3 중 하나를 선택하여 감소하고 2 삽입
[3, 2, 3]  (피로도: 3² + 2² + 3² = 22)
  3
 / \
3   2   n = 2

3단계: 3 중 하나를 선택하여 감소하고 2삽입
[2, 2, 3]  (피로도: 2² + 2² + 3² = 17)
  3
 / \
2   2   n = 1

4단계: 3을 선택하여 감소하고 2 삽입
[2, 2, 2]  (피로도: 2² + 2² + 2² = 12)
  2
 / \
2   2   n = 0
```

Abstraction and Modeling

이 문제는 수학적으로 다음과 같이 모델링할 수 있습니다:

목적 함수: $\min \sum_{i=1}^{m} w_i^2$

제약 조건: $\sum_{i=1}^{m} (o_i - w_i) = n$, $w_i \geq 0$

- $m$: 작업의 수
- $w_i$: i번째 작업의 최종 작업량
- $o_i$: i번째 작업의 초기 작업량
- $n$: 사용 가능한 총 시간

이 모델링을 통해 우리는 다음과 같은 특성을 파악할 수 있습니다:

1. 최적화 문제의 특성:
   - 이는 제약 조건 하에서 목적 함수를 최소화하는 최적화 문제입니다.
   - 특히, 각 작업량의 제곱의 합을 최소화해야 하므로, 작업량 간의 균형을 맞추는 것이 중요합니다.
   - 이는 "볼록 최적화(Convex Optimization)" 문제의 한 형태로, 지역 최적해가 전역 최적해가 됩니다.

2. 자료 구조 선택:
   - 매 단계에서 가장 큰 작업량을 찾아 줄여야 합니다.
   - 따라서, 최대값을 빠르게 찾고 갱신할 수 있는 자료 구조가 필요합니다.
   - 우선순위 큐(최대 힙)가 이 요구사항을 가장 잘 만족시킵니다:
     - 최대값 추출: O(1)
     - 요소 삽입/갱신: O(log n)

이러한 분석을 통해, 우리는 그리디 알고리즘과 우선순위 큐를 결합한 접근 방식이 이 문제에 적합하다는 결론을 내릴 수 있습니다.

Algorithm Design

```java
function solution(n, works):
    max_heap = create_max_heap(works)

    for i = 1 to n:
        if max_heap is empty:
            break
        max_work = max_heap.extract_max()
        if max_work > 0:
            max_heap.insert(max_work - 1)

    result = 0
    while max_heap is not empty:
        work = max_heap.extract_max()
        result += work * work

    return result
```

Time Complexity: O(n log m), where n is the given time and m is the number of works.
Space Complexity: O(m) for the priority queue.

1. Implementation

```java
import java.util.PriorityQueue;
import java.util.Collections;

public class Solution {
    /**
     * 야근 지수를 최소화하는 함수
     *
     * @param n     퇴근까지 남은 시간
     * @param works 각 일에 대한 작업량 배열
     * @return 야근 피로도를 최소화한 값
     */
    public long solution(int n, int[] works) {
        // 최대 힙으로 사용할 우선순위 큐 생성
        PriorityQueue<Integer> maxHeap = new PriorityQueue<>(Collections.reverseOrder());

        // 모든 작업을 우선순위 큐에 추가
        for (int work : works) {
            maxHeap.offer(work);
        }

        // n시간 동안 작업량 감소
        while (n > 0 && !maxHeap.isEmpty()) {
            int maxWork = maxHeap.poll();
            if (maxWork > 0) {
                maxHeap.offer(maxWork - 1);
            }
            n--;
        }

        // 남은 작업량의 제곱의 합 계산
        long result = 0;
        while (!maxHeap.isEmpty()) {
            int remainingWork = maxHeap.poll();
            result += (long) remainingWork * remainingWork;
        }

        return result;
    }

    public static void main(String[] args) {
        Solution solution = new Solution();

        // 테스트 케이스
        int n1 = 4;
        int[] works1 = {4, 3, 3};
        System.out.println("Test case 1: " + solution.solution(n1, works1)); // Expected: 12

        int n2 = 1;
        int[] works2 = {2, 1, 2};
        System.out.println("Test case 2: " + solution.solution(n2, works2)); // Expected: 6

        int n3 = 3;
        int[] works3 = {1, 1};
        System.out.println("Test case 3: " + solution.solution(n3, works3)); // Expected: 0
    }
}
```

이 구현에서 우리는 Java의 PriorityQueue를 최대 힙으로 사용하여 매 순간 가장 큰 작업량을 효율적으로 찾고 감소시킵니다.

- 시간 복잡도: O(n log m), 여기서 n은 주어진 시간, m은 작업의 수입니다.
- 공간 복잡도: O(m), 우선순위 큐의 크기입니다.

이 접근 방식은 문제의 최적화 특성(작업량 간의 균형)을 활용하며, 우선순위 큐라는 적절한 자료 구조를 선택하여 효율적으로 문제를 해결합니다.
