package p42626;

import java.util.*;

/**
 * [더 맵게](https://school.programmers.co.kr/learn/courses/30/lessons/42626?language=java)
 */
class Solution {
    /**
     * - 모든 음식의 스코빌 지수를 K 이상으로 만들고 싶습니다.
     * - 스코빌 지수가 가장 낮은 두 개의 음식을 다음 규칙에 따라 섞어 새로운 음식 만듭니다.
     * <p>
     * 규칙: 섞은 음식 스코빌 지수 = 가장 맵지 않은 음식의 스코빌 지수 + (두 번째로 맵지 않은 음식의 스코빌 지수 * 2)
     * <p>
     * 모든 음식의 스코빌 지수가 K 이상이 될 때까지 반복하여 섞습니다.
     *
     * @param scoville 음식의 스코빌 지수 배열
     * - 2 <= socville.length <= 1,000,000
     * - 0 <= scoville[i] <= 1,000,000
     * @param K 목표로 하는 스코빌 지수
     * - 0 <= K <= 1,000,000,000
     *
     * @return 모든 음식의 스코빌 지수를 K 이상으로 만들기 위해 섞어야 하는 최소 횟수. 모두 변경 불가능할 경우 -1.
     */
    public int solution(int[] scoville, int K) {
        // [1, 2, 3, 9, 10, 12], K=7
        // 1. 1 + (2 * 2) = 5 < 7
        //   => [5, 3, 9, 10, 12]
        // 2. 3 + (5 * 2) = 13 > 7
        //   => [13, 9, 10, 12]
        //   => 2회

        int answer = 0;
        if (K == 0) {
            return answer;
        }

        PriorityQueue<Integer> pq = new PriorityQueue<>();

        // 큐에 담으면서 혹시 이미 모든 스코빌 지수가 K 이상인지 확인합니다.
        boolean alreadyOverK = true;
        for (int el : scoville) {
            if (el <= K) { // 하나라도 K 이하면 false
                alreadyOverK = false;
            }
            pq.offer(el);
        }

        if (alreadyOverK) {
            return answer;
        }

        // 우선순위 큐의 힙은 완전 이진 트리(complete binary tree)의 형태를 가집니다.
        // - 부모 노드의 값은 항상 자식 노드의 값보다 작거나 같습니다.
        // - 완전 이진 트리 형태를 유지합니다. 트리가 항상 왼쪽에서 오른쪽으로 빈 공간 없이 채워지는 것을 보장합니다.
        //   마지막 레벨을 제외한 모든 레벨이 완전히 채워져 있고, 마지막 레벨은 왼쪽부터 채워져 있습니다.
        // - `i`에 해당하는 노드의 왼쪽 자식은 `2i + 1`, 오른쪽 자식은 `2i + 2` 인덱스에 해당합니다.
        // heapify 과정
        // 1. 초기
        //    0  1  2  3  4   5
        //   [1, 2, 3, 9, 10, 12]
        //          1
        //       /     \
        //      2       3
        //     / \     /
        //    9  10   12
        //
        // 2. poll() => 1이 제거됩니다.
        //   a. 루트 노드를 제거하면 트리의 구조가 깨집니다.
        //      빈 자리를 채우기 위해 배열의 마지막 요소를 루트로 이동시키면, 배열의 크기가 하나 줄어들고, 트리의 형태는 유지됩니다.
        //      마지막 노드 12가 루트로 이동
        //      [12, 2, 3, 9, 10]
        //          12
        //       /     \
        //      2       3
        //     / \
        //    9  10
        //
        //   b. 더 작은 2와 교환
        //      [2, 12, 3, 9, 10]
        //          2
        //       /     \
        //      12       3
        //     / \
        //    9  10
        //
        //   c. 더 작은 9와 교환
        //      [2, 9, 3, 12, 10]
        //          2
        //       /     \
        //      9       3
        //     / \
        //    12  10
        //
        // 3. poll() => 2가 제거됩니다.
        //   a. 마지막 노드 10이 루트로 이동
        //      [10, 9, 3, 12]
        //         10
        //       /    \
        //      9      3
        //     /
        //    12
        //
        //   b. 더 작은 값 3과 교환
        //      [3, 9, 10, 12]
        //         3
        //       /   \
        //      9     10
        //     /
        //    12
        //
        // 4. 값 5 추가
        //   a. 트리의 마지막 위치에 5 추가
        //      [3, 9, 10, 12, 5]
        //         3
        //       /   \
        //      9     10
        //     / \
        //    12  5
        //
        //   b. 부모노드 9보다 작으므로 두 값을 교환
        //      [3, 5, 10, 12, 9]
        //         3
        //       /   \
        //      5     10
        //     / \
        //    12  9
        while (!pq.isEmpty()) {
            // 가장 작은 스코빌 지수가 K 이상이면 모두 처리한 것이므로 종료
            if (pq.peek() >= K) break;

            Integer smallest = pq.poll();
            Integer secondSmallest = pq.poll();
            // 두 번째로 작은 값이 없다면 smallest 스코빌 지수를 높일 수 없으므로 -1 리턴
            if (secondSmallest == null) return -1;

            pq.add(smallest + (secondSmallest * 2));
            answer++;
        }

        return answer;
    }
}

class TestCase {
    public int[] scoville;
    public int K;
    public int answer;

    TestCase(int[] scoville, int K, int answer) {
        this.scoville = scoville;
        this.K = K;
        this.answer = answer;
    }
}

class Main {
    public static void main(String[] args) {
        Solution s = new Solution();

        List<TestCase> testCases = new ArrayList<>();
        testCases.add(new TestCase(
                new int[]{1, 2, 3, 9, 10, 12},
                7,
                2
        ));
        testCases.add(new TestCase(
                new int[]{1, 2, 3, 9, 10, 12},
                0,
                0
        ));
        testCases.add(new TestCase(
                new int[]{1, 1, 1, 1},
                100,
                -1
        ));

        for (TestCase tc : testCases) {
            System.out.println(s.solution(tc.scoville, tc.K) == tc.answer);
        }
    }
}
