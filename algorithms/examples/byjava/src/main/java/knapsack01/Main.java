package main.java.knapsack01;

import java.util.Arrays;

/**
 * Main
 */
public class Main {

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
        if (W <= 0 || n <= 0 || weights == null || values == null
                || weights.length != n || values.length != n) {
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
         * i 0 1 2 3 4 5 ... W 무게
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
         * 개수
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
        // 남은 무게 한도를 추적하기 위해 전체 무게 한도로 초기화합니다.
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
        // int[] weights = {3, 4, 5};
        // int[] values = {4, 5, 6};
        // int W = 7;
        int n = values.length;

        Result result = solve(W, weights, values, n);

        System.out.println("최대 가치: " + result.maxValue);
        System.out.println("선택된 물건:");
        for (int i = 0; i < n; i++) {
            if (result.selectedItems[i]) {
                System.out.println("물건 " + (i + 1) + ": 무게 = " + weights[i]
                        + ", 가치 = " + values[i]);
            }
        }
    }

}
