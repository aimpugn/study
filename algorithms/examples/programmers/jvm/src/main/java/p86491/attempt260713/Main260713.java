package p86491.attempt260713;

import support.TestCase;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/86491?language=java">최소직사각형</a>
 */
class Solution {
    /**
     * 지갑의 크기 정하려 합니다.
     * - 다양한 모양과 크기의 명함들 모두 수납 가능해야 함
     * - 작아서 들고 다니기 편해야 함
     * 모든 명함의 가로/세로 길이를 조사했습니다.
     * 1. 60/50
     * 2. 30/70
     * 3. 60/30
     * 4. 80/40
     * 가장 긴 가로 80, 가장 긴 세로 70으로 만들면 모든 명함 수납이 가능합니다.
     * 하지만 2번 명함을 가로로 눕힌다면 가로 80, 세로 50으로 모든 명함 수납이 가능합니다.
     *
     * @param sizes 모든 명함의 가로 길이와 세로 길이를 나타내는 2차원 배열
     * - 1 <= sizes.length <= 10,000
     * - sizes[i] = [w, h]
     * - w: 가로 길이
     * - h: 세로 길이
     * - 1 <= w, h <= 1,000
     * @return 가능한 가장 작은 지갑의 크기
     */
    public int solution(int[][] sizes) {
        // 가로, 세로 모두 가장 길게 하면 모두 수용 가능하지만,
        // 그러면 크기가 커집니다.
        // 그래서 문제는 카드를 돌려서 "긴 것은 긴 것들끼리, 짧은 것은 짧은 것들끼리 비교"하라고 합니다.
        // 첨삭: 여기서 핵심은 회전할지를 명함마다 따로 탐색하지 않고, 두 변을 긴 변과 짧은 변으로
        // 통일해서 보는 것입니다. 이렇게 방향을 통일하면 서로 다른 명함을 같은 기준으로 비교할 수 있습니다.
        // > 회전 가능한 직사각형은 긴 변끼리, 짧은 변끼리 모은 뒤 각 묶음의 최댓값을 구한다.
        if (sizes == null) return 0;
        if (sizes.length == 1) {
            return sizes[0][0] * sizes[0][1];
        }

        // 첨삭: 문제의 입력에서는 sizes가 null이 아니고 길이도 1 이상입니다. 따라서 위 두 분기는
        // 정답에 꼭 필요하지 않습니다. 명함이 한 장이어도 아래 반복문이 그대로 처리하므로,
        // 일반 풀이에서는 예외 분기를 줄이는 편이 같은 불변식을 모든 입력에 적용하기 쉽습니다.
        // 특히 계약에 없는 null 입력에 0을 반환하면 잘못된 호출을 정상 결과처럼 보이게 할 수 있습니다.

        var maxGreater = 0; // 더 큰 값들 중 가장 큰 값
        var maxSmaller = 0; // 더 작은 값들 중 가장 큰 값
        // 첨삭: 반복문을 k번 마칠 때마다 두 변수는 "앞에서 본 k장의 긴 변 최댓값"과
        // "앞에서 본 k장의 짧은 변 최댓값"입니다. 이 관계가 끝까지 유지되는 것이 루프 불변식입니다.
        // 모든 변의 길이가 1 이상이므로 초깃값 0은 첫 명함을 반영할 때 반드시 실제 길이로 바뀝니다.
        for (var size : sizes) { // O(n): 명함 n장을 한 번씩 확인합니다.
            // 큰 값, 작은 값의 인덱스를 찾습니다.
            // 첨삭: 인덱스를 골라내는 방식도 맞습니다. 다만 Math.max와 Math.min으로 두 변의 값을
            // 바로 구하면 어느 인덱스가 긴 변인지 따로 기억하지 않아도 되어 풀이 의도가 더 잘 드러납니다.
            var smallerIdx = 0;
            var greaterIdx = 1;
            if (size[0] > size[1]) {
                smallerIdx = 1;
                greaterIdx = 0;
            }

            // O(1): 현재 명함 한 장을 반영하는 비교 횟수는 입력 크기와 무관하게 일정합니다.
            maxGreater = Math.max(maxGreater, size[greaterIdx]);
            maxSmaller = Math.max(maxSmaller, size[smallerIdx]);
        }
        // System.out.println("maxGreater: " + maxGreater + ",  maxSmaller: " + maxSmaller);

        /*
         * 첨삭: 첫 번째 예시는 다음처럼 누적됩니다.
         *
         * 명함       긴 변/짧은 변    maxGreater/maxSmaller
         * [60, 50]   60/50            60/50
         * [30, 70]   70/30            70/50
         * [60, 30]   60/30            70/50
         * [80, 40]   80/40            80/50
         *
         * 지갑과 명함을 모두 긴 쪽과 짧은 쪽으로 놓고 보면, 명함의 긴 변은 지갑의 긴 쪽보다
         * 길 수 없고 명함의 짧은 변도 지갑의 짧은 쪽보다 길 수 없습니다. 따라서 지갑의 두 변은
         * 각각 명함의 긴 변 최댓값과 짧은 변 최댓값보다 작아질 수 없습니다. 지금 구한 80 x 50은
         * 이 두 값을 정확히 만족하면서 모든 명함을 담으므로 넓이도 최소입니다.
         *
         * 전체 시간 복잡도는 O(n), 추가 공간 복잡도는 O(1)입니다. 각 변은 최대 1,000이므로
         * 결과의 최댓값 1,000,000은 int 범위에 들어갑니다.
         */
        return maxGreater * maxSmaller;
    }
}

public class Main260713 {
    public static void main(String[] args) {
        var solution = new Solution();

        var testCases = List.of(
            new TestCase<>(new int[][]{{60, 50}, {30, 70}, {60, 30}, {80, 40}}, 4000),
            new TestCase<>(new int[][]{{10, 7}, {12, 3}, {8, 15}, {14, 7}, {5, 15}}, 120),
            new TestCase<>(new int[][]{{14, 4}, {19, 6}, {6, 16}, {18, 7}, {7, 11}}, 133),
            new TestCase<>(new int[][]{{14, 4}}, 56)
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
