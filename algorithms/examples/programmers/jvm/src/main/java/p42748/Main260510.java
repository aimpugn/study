package p42748;

import java.util.Arrays;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42748?language=java">K번째수</a>
 */
class Solution {
    /**
     * TODO: 문제 조건, 예제, 제약을 먼저 작은 문장으로 분해하세요.
     */
    public int[] solution(int[] array, int[][] commands) {
        // TODO: 풀이를 구현하세요.
        return new int[]{};
    }
}

public class Main260510 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        assertArrayEquals(
                new int[]{5, 6, 3},
                solution.solution(
                        new int[]{1, 5, 2, 6, 3, 7, 4},
                        new int[][]{{2, 5, 3}, {4, 4, 1}, {1, 7, 3}}
                )
        );
    }

    private static void assertArrayEquals(int[] expected, int[] actual) {
        if (!Arrays.equals(expected, actual)) {
            throw new AssertionError("expected=" + Arrays.toString(expected)
                    + ", actual=" + Arrays.toString(actual));
        }
    }
}
