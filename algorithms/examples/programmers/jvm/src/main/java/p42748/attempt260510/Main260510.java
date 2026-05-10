package p42748.attempt260510;

import java.util.Arrays;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42748?language=java">K번째수</a>
 */
class Solution {
    /**
     *
     * @param array 원본 배열
     * - 1 <= array.length <= 100
     * - 1 <= array[i] <= 100
     * @param commands [i, j, k]를 원소로 갖는 2차원 배열
     * - 1 <= commands.length <= 50
     * - commands[i].length == 3
     *
     * @return `array[i..j]` 자르고 정렬했을 때, k 번째에 있는 수
     */
    public int[] solution(int[] array, int[][] commands) {
        // array: [1, 5, 2, 6, 3, 7, 4]
        // commands: [[2, 5, 3], [4, 4, 1], [1, 7, 3]]
        // return : [5, 6, 3]
        //
        // 1. [2, 5, 3] 경우
        //   => [5, 2, 6, 3] => [2, 3, 5, 6]
        //   => 5
        //
        // 2. [4, 4, 1] 경우
        //   => [6] => [6]
        //   => 6
        //
        // 3. [1, 7, 3] 경우
        //   => [1, 5, 2, 6, 3, 7, 4] => [1, 2, 3, 4, 5, 6, 7]
        //   => 3
        // 첨삭: 문제 크기가 작아서 핵심 판단은 "명령마다 잘라서 정렬해도 충분하다"입니다.
        // 아래처럼 int[]를 유지하면 "자른다 -> 정렬한다 -> k번째를 꺼낸다"가 라인 단위로 바로 보입니다.
        // int[] sliced = Arrays.copyOfRange(array, command[0] - 1, command[1]);
        // Arrays.sort(sliced);
        // answer[idx++] = sliced[command[2] - 1];
        var answer = new int[commands.length];
        var idx = 0;
        for (var command : commands) {
            var sliced = Arrays.stream(Arrays.copyOfRange(array, command[0] - 1, command[1]))
                .boxed()
                .sorted()
                .toArray(Integer[]::new);
            answer[idx++] = sliced[command[2] - 1];
        }

        return answer;
    }
}

public class Main260510 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        // 첨삭: 공식 예제를 통째로만 확인하면 어느 명령에서 깨졌는지 좁히기 어렵습니다.
        // 인덱스 보정 습관을 굳히려면 아래처럼 길이 1 구간과 전체 구간을 따로도 확인해 보세요.
        // assertArrayEquals(new int[]{6}, solution.solution(new int[]{1, 5, 2, 6, 3, 7, 4}, new int[][]{{4, 4, 1}}));
        // assertArrayEquals(new int[]{3}, solution.solution(new int[]{1, 5, 2, 6, 3, 7, 4}, new int[][]{{1, 7, 3}}));
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
