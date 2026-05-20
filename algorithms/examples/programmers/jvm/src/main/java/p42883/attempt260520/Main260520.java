package p42883.attempt260520;

import support.TestCase2;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42883?language=java">큰 수 만들기</a>
 */
class Solution {
    public String solution(String number, int k) {
        return "";
    }
}

public class Main260520 {
    public static void main(String[] args) {
        var solution = new Solution();

        var testCases = List.of(
            new TestCase2<>("1924", 2, "94"),
            new TestCase2<>("1231234", 3, "3234"),
            new TestCase2<>("4177252841", 4, "775841")
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input(), testCase.input2());
            assertEquals(testCase, result);
        }
    }

    private static void assertEquals(TestCase2<?, ?, ?> testCase, Object actual) {
        if (!valuesEqual(testCase.answer(), actual)) {
            throw new AssertionError(
                "input=" + valueToString(testCase.input())
                    + ", input2=" + valueToString(testCase.input2())
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
