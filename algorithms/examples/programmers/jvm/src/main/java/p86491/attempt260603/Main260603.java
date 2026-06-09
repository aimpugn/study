package p86491.attempt260603;

import support.TestCase;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/86491?language=java">최소직사각형</a>
 */
class Solution {
    /**
     * - 다양한 모양과 크기의 명함들을 "모두" 수납할 수 있어야 합니다.
     * - 작아서 들고 다니기 편해야 합니다.
     * - 모든 명함의 가로/세로 길이
     * <p>
     * 1. 가로 60, 세로 50
     * 2. 가로 30, 세로 70
     * 3. 가로 60, 세로 30
     * 4. 가로 80, 세로 40
     * <p>
     * 가장 긴 가로는 4번 명함의 80, 가장 긴 세로는 2번 명함의 70
     * - 80x70이면 모든 명함 수납 가능합니다.
     * - 하지만 2번 명함을 가로로 눕혀 수납하면 80x50으로 모든 명함 수납 가능합니다.
     *
     * @param sizes 모든 명함의 가로 길이와 세로 길이 나타내는 2차원 배열
     * - 1 <= sizes.length <= 10,000
     * - sizes[i]는 [width, height] 형식
     * - 1 <= w, h <= 1,000
     *
     * @return 모든 명함을 회전 포함해 수납할 수 있는 가장 작은 지갑의 넓이
     *
     * <p>풀이 핵심: 명함은 회전할 수 있어 가로와 세로는 서로 맞바꿔도 되는 대칭이다.
     * 그래서 각 명함을 (긴 변, 짧은 변)으로 먼저 정규화하면 "어느 쪽이 기냐"는 분기가
     * 통째로 사라지고, 긴 변끼리의 최댓값 × 짧은 변끼리의 최댓값만 남는다. 시간 O(n), 공간 O(1).
     * 비교 횟수 트레이드오프와 branchless, 정규화의 전이 사례는 같은 패키지의 PROCESS.md
     * "재풀이 회고" 절에 정리해 두었다.
     */
    public int solution(int[][] sizes) {
        // 가로가 길어도 세로로 돌릴 수 있고,
        // 세로가 길어도 가로로 돌릴 수 있습니다.
        // 일단 한 쪽을 가장 길게.
        // [막아 둔 1차 시도] if로 가로·세로 중 누가 긴지 매번 갈랐다. 명함당 비교는 3번으로
        // 아래 max/min(4번)보다 1번 적지만, 진짜 분기라 의도("긴 변/짧은 변")가 드러나지 않는다.
        /*var resultWidth = 0;
        var resultHeight = 0;

        for (var size : sizes) {
            //
            var width = size[0];
            var height = size[1];

            // 둘 중 더 큰 값을 가로로, 그러면 세로는 70 -> 30
            if (height < width) {
                resultWidth = Math.max(resultWidth, width);
                resultHeight = Math.max(resultHeight, height);
            } else {
                resultWidth = Math.max(resultWidth, height);
                resultHeight = Math.max(resultHeight, width);
            }
        }*/

        // 각 명함에서 긴 변과 짧은 변을 가려, 긴 변끼리의 최대, 짧은 변끼리의 최대를 따로 모은다.
        // [채택] 회전 = 가로/세로 대칭이므로 (긴 변, 짧은 변)으로 정규화하면 분기가 사라진다.
        // Math.max/min은 같은 두 값을 두 번 비교하지만, 보통 분기 없는 명령(CMOV)으로 컴파일되어
        // 이 규모(n<=10,000)에서 비교 한 번 차이는 무의미하다. 자세한 트레이드오프는 PROCESS.md.
        var longestWidth = 0;
        var longestHeight = 0;
        for (var size : sizes) {
            // 더 큰 값을 너비로 봅니다.
            var width = Math.max(size[0], size[1]);
            // 더 작은 값을 높이로 봅니다.
            var height = Math.min(size[0], size[1]);

            longestWidth = Math.max(width, longestWidth);
            longestHeight = Math.max(height, longestHeight);
        }

        return longestWidth * longestHeight;
    }
}

public class Main260603 {
    public static void main(String[] args) {
        var solution = new Solution();

        var testCases = List.of(
            new TestCase<>(new int[][]{{60, 50}, {30, 70}, {60, 30}, {80, 40}}, 4000),
            new TestCase<>(new int[][]{{10, 7}, {12, 3}, {8, 15}, {14, 7}, {5, 15}}, 120),
            new TestCase<>(new int[][]{{14, 4}, {19, 6}, {6, 16}, {18, 7}, {7, 11}}, 133)
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input());
            assertEquals(testCase, result);
        }
    }

    private static void assertEquals(TestCase<?, ?> testCase, Object actual) {
        if (!valuesEqual(testCase.answer(), actual)) {
            throw new AssertionError(
                "input=" + valueToString(testCase.input())
                    + ", expected=" + valueToString(testCase.answer())
                    + ", actual=" + valueToString(actual)
            );
        }
    }

    private static boolean valuesEqual(Object expected, Object actual) {
        if (expected instanceof int[] expectedArray && actual instanceof int[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof long[] expectedArray && actual instanceof long[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof Object[] expectedArray && actual instanceof Object[] actualArray) {
            return Arrays.deepEquals(expectedArray, actualArray);
        }
        return Objects.equals(expected, actual);
    }

    private static String valueToString(Object value) {
        if (value instanceof int[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof long[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof Object[] values) {
            return Arrays.deepToString(values);
        }
        return String.valueOf(value);
    }
}
