package p43105;

import java.util.*;

/**
 * [정수 삼각형](https://school.programmers.co.kr/learn/courses/30/lessons/43105?language=java)
 */
class Solution {
    /**
     * 삼각형 꼭대기에서 바닥까지 이어지는 경로 중, 거쳐간 숫자의 합이 가장 큰 경우 찾기
     * 아래 칸으로 이동 시 대각선 방향으로 한 칸 오른쪽 또는 외쪽으로만 이동 가능
     *
     * @param triangle 삼각형의 정보가 담긴 배열
     *                 - 1 <= triangle.length <= 500
     *                 - 0 <= triangle[i] <= 9,999
     * @return 거쳐간 숫자의 최대값
     */
    public int solution(int[][] triangle) {
        int answer = 0;
        // [[7], [3,8], [8,1,0], [2,7,4,4], [4,5,2,6,5]]
        //         7
        //       3   8
        //     8   1   0
        //   2   7   4   4
        // 4   5   2   6   5 <- 끝까지 도달해야 합니다.
        //
        // 7
        // 3, 8
        // 8, 1, 0
        // 2, 7, 4, 4
        // 4, 5, 2, 6, 5
        //
        // 최대값: 30
        //
        // 7 -> 3 -> 8 -> 7 -> 5
        // triangle[0][0] + triangle[1][0] + triangle[2][0] + triangle[3][1] + triangle[4][1]
        //
        // 직전까지의 최대값을 기반으로 다음 최대값을 구합니다.
        // 최대값은 그 전까지의 최대값을 필요로 합니다.
        // 즉, N 번째 최대값은 N-1 번째 최대값에 기반합니다.
        //
        // N 번째 최대값 = N-1 까지의 최대값 + max(
        //      N 번째 left,
        //      N 번째 right
        // )
        //
        // 이때, 좌/우를 넓혀가는 것보다, 좁혀가는 게 더 편할 것으로 보입니다.
        //
        // 4     5     2      6     5
        //  \   / \   / \    / \   /
        //  6  7  12 9   6 10  10 9 => 각각 최대값을 기록합니다.
        //   2      7     4      4
        //
        // for 현재 가능한 수 in 현재 가능한 수들
        //   n 번째 최대 값 = max(
        //      현재값 + n-1번째 로우의 left,
        //      현재값 + n-1번째 로우의 right,
        //   )
        //
        // 아래에서부터 위로 올라가면서 역으로 추론해 나갑니다.
        int triangleLength = triangle.length;
        int triangleLastIdx = triangle.length - 1;
        int[][] dp = new int[triangleLength][triangleLength];

        // 바닥부터 올라갈 것이므로, 가장 바닥을 초기화합니다.
        for (int i = 0; i < triangle[triangleLastIdx].length; i++) {
            dp[triangleLastIdx][i] = triangle[triangleLastIdx][i];
        }

        for (int depth = triangleLength - 1; depth - 1 >= 0; depth--) {
            int[] nextRow = triangle[depth - 1];
            for (int col = 0; col < nextRow.length; col++) {
                int nextVal = nextRow[col];
                // 이전에 계산된 값을 기반으로 다음 값을 계산합니다.
                dp[depth - 1][col] = Math.max(
                        nextVal + dp[depth][col],
                        nextVal + dp[depth][col + 1]
                );
            }
        }

        return dp[0][0];
    }

    public int solutionByMemoization(int[][] triangle) {
        int n = triangle.length;
        int[][] memo = new int[n][n]; // 메모이제이션 배열 초기화

        // 메모이제이션 배열을 -1로 초기화 (아직 계산되지 않은 상태를 나타냄)
        for (int i = 0; i < n; i++) {
            Arrays.fill(memo[i], -1);
        }

        // 재귀 함수 호출: 꼭대기에서 시작
        return findMaxPath(0, 0, memo, triangle);
    }

    /**
     * @param i        시작 row
     * @param j        시작 col
     * @param memo     최대값 기록 위한 이차원 배열
     * @param triangle 삼각형을 나타내는 이차원 배열
     * @return 각 단계별 최대값을 반환합니다.
     */
    private int findMaxPath(int i, int j, int[][] memo, int[][] triangle) {
        System.out.println("i: " + i + ", j: " + j);
        // 기저 조건: 바닥 행에 도달한 경우, 해당 값 반환
        if (i == triangle.length - 1) {
            return triangle[i][j];
        }

        // 이미 계산된 값이 있다면, 바로 반환
//        if (memo[i][j] != -1) {
//            return memo[i][j];
//        }

        // 현재 위치의 값 + 아래 두 경로 중 최대값을 계산
        // ex: [[7], [3,8], [8,1,0], [2,7,4,4], [4,5,2,6,5]]
        // 0,0
        //    -> 1,0
        //           -> 2,0
        //                 -> 3,0
        //                       -> 4,0
        //                       -> 4,1
        //                 -> 3,1
        //                       -> 4,1
        //                       -> 4,2
        //           -> 2,1
        //                 -> 3,1 => 메모 사용
        //                       -x-> 4,1
        //                       -x-> 4,2
        //                 -> 3,2
        //                       -> 4,2
        //                       -> 4,3
        //    -> 1,1
        //           -> 2,1 => 메모 사용
        //                 -x-> 3,1
        //                       -x-> 4,1
        //                       -x-> 4,2
        //                 -x-> 3,2
        //                       -x-> 4,2
        //                       -x-> 4,3
        //           -> 2,2
        //                 -x-> 3,2 => 메모 사용
        //                       -x-> 4,2
        //                       -x-> 4,3
        //                 --> 3,3
        //                       --> 4,3
        //                       --> 4,4
        //
        int left = findMaxPath(i + 1, j, memo, triangle);     // 왼쪽 아래로 이동
        int right = findMaxPath(i + 1, j + 1, memo, triangle); // 오른쪽 아래로 이동

        // 결과를 메모이제이션하고 반환
        memo[i][j] = triangle[i][j] + Math.max(left, right);
//        System.out.println(
//                "memo: " + Arrays.deepToString(memo) + "\n"
//                        + "\ttriangle[i][j]: " + triangle[i][j] + "\n"
//                        + "\ti: " + i + "\n"
//                        + "\tj: " + j + "\n"
//                        + "\tleft: " + left + "\n"
//                        + "\tright: " + right + "\n"
//                        + "\tmemo[i][j]: " + memo[i][j] + "\n"
//
//        );
        return memo[i][j];
    }
}

class TestCase {
    public int[][] triangle;

    TestCase(int[][] triangle) {
        this.triangle = triangle;
    }
}

public class Main {

    public static void main(String[] args) {

        Solution s = new Solution();
        List<TestCase> testCases = new ArrayList<>();

        testCases.add(new TestCase(new int[][]{
                new int[]{7},
                new int[]{3, 8},
                new int[]{8, 1, 0},
                new int[]{2, 7, 4, 4},
                new int[]{4, 5, 2, 6, 5},
        }));

        for (TestCase tc : testCases) {
            System.out.println(s.solutionByMemoization(tc.triangle));
        }
    }
}
