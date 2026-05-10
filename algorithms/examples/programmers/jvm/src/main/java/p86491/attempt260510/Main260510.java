package p86491.attempt260510;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/86491?language=java">최소직사각형</a>
 */
class Solution {
    /**
     * @param sizes 모든 명함의 [가로, 세로] 길이
     * - 1 <= sizes.length <= 10,000
     * - 1 <= sizes[i][0], sizes[i][1] <= 1,000
     * @return 모든 명함을 수납할 수 있는 가장 작은 지갑의 넓이
     */
    public int solution(int[][] sizes) {
        // TODO: 각 명함은 회전할 수 있습니다. 한 명함 안에서 큰 쪽/작은 쪽을 먼저 고정해 보세요.
        // TODO: 풀이를 구현하세요.
        return 0;
    }
}

public class Main260510 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        // TODO: 풀이를 구현한 뒤 아래 예제를 하나씩 주석 해제해 확인하세요.
        // assertEquals(4000, solution.solution(new int[][]{{60, 50}, {30, 70}, {60, 30}, {80, 40}}));
        // assertEquals(120, solution.solution(new int[][]{{10, 7}, {12, 3}, {8, 15}, {14, 7}, {5, 15}}));
        // assertEquals(133, solution.solution(new int[][]{{14, 4}, {19, 6}, {6, 16}, {18, 7}, {7, 11}}));
    }

    private static void assertEquals(int expected, int actual) {
        if (expected != actual) {
            throw new AssertionError("expected=" + expected + ", actual=" + actual);
        }
    }
}
