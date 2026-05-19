package p154539;

import support.TestCase;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/154539">뒤에 있는 큰 수 찾기</a>
 */
class Solution {
    public int[] solution(int[] numbers) {
        return new int[]{};
    }
}

public class Main260519 {
    public static void main(String[] args) {
        var solution = new Solution();

        var testCases = List.of(
            new TestCase<>(new int[]{2, 3, 3, 5}, new int[]{3, 5, 5, -1}),
            new TestCase<>(new int[]{9, 1, 5, 3, 6, 2}, new int[]{-1, 5, 6, 6, -1, -1})
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
