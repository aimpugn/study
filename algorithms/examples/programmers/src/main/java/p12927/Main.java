package p12927;

import java.util.*;

public class Main {
    /**
     * - if 야근, then 피로도 쌓임
     * - 피로도 = 야근 시작 시점에서 남은 일의 작업량을 제곱하여 더한 값
     * - 1시간 동안 작업량 1 처리 가능
     * <p>
     * 주어진 n을 주어진 배열의 요소에서 적절히 차감하여 제곱한 결과 최소값을 찾습니다.
     * ex:
     * - works: [4, 3, 3]
     * n: 4
     * => [4 -2, 3 -1, 3 - 1]
     * => [2, 2, 2]
     * => 4 + 4 + 4 = 12
     * <p>
     * - works: [2, 1, 2]
     * n: 1
     * => [2 - 1, 1, 2]
     * => 1 + 1 + 4 = 6
     *
     * @param n     퇴근까지 남은 시간
     *              1 <= n <= 1,000,000
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
        // 이를 조합 최적화 문제로 접근하려고 했는데, 조합 최적화 문제가 아닌 거로 보입니다.
        //
        // 주어진 n을 각 배열의 요소에 어떻게 차감할 것인지?
        // 지수는 2이므로, 밑이 클수록 그 합도 커질 수밖에 없습니다.
        // 그러면 밑이 큰 수부터 하나씩 차감을 합니다
        // 이를 위해 우선순위 큐(최대 힙)을 사용합니다.

        // 야근이 필요 없는 날입니다.
        int total = Arrays.stream(works).sum();
        if (total <= n) {
            return 0;
        }

        PriorityQueue<Integer> queue =
                new PriorityQueue(Collections.reverseOrder());

        for (int work : works) {
            queue.add(work);
        }

        while (n-- > 0) {
            queue.add(queue.poll() - 1);
        }

        while (!queue.isEmpty()) {
            int curr = queue.poll();
            answer += curr * curr;
        }

        return answer;
    }
}
