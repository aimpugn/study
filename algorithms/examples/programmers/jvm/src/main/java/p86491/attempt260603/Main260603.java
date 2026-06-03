package p86491.attempt260603;

import support.TestCase;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/86491?language=java">최소직사각형</a>
 */
class Solution {
    public int solution(int[][] sizes) {
        return 0;
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
