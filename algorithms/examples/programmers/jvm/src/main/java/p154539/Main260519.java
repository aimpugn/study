package p154539;

import support.TestCase;

import java.util.ArrayDeque;
import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/154539">뒤에 있는 큰 수 찾기</a>
 */
class Solution {
    /**
     * 뒷 큰수: 정수 배열에서 각 원소들에 대해 자신보다 뒤에 있는 숫자 중에서 자신보다 크면서 가장 가까이 있는 수
     *
     * @param numbers 정수로 이루어진 배열
     * - 4 <= numbers.length <= 1,000,000
     * - 1 <= numbers[i] <= 1,000,000
     *
     * @return 모든 원소에 대해 뒷 큰수들을 차례로 담은 배열
     * - 뒷 큰수가 존재하지 않는 원소는 -1
     */
    public int[] solution(int[] numbers) {

        // 자신보다 크면서 가장 가까이 있는 수를 찾아야 합니다.
        // 어떻게 찾을 것인가? 일단 이중 for문이 생각납니다.
        // 항상 자신보다 뒤의 수들만 보면 될 것으로 보입니다.
        // => 실패
        // 단순 이중 for문을 돌면 시간 초과 케이스가 발생합니다.
        // 현재 위치부터 항상 그 뒤를 전부 비교하면 비효율적이게 됩니다.O(n^2)
        //
        var answer = new int[numbers.length];
        var stack = new ArrayDeque<Integer>(); // 아직 답을 못 찾은 인덱스들을 보관합니다.

        for (var i = 0; i < numbers.length; i++) { // 왼쪽에서 오른쪽으로 한 번만 훑습니다.
            var curr = numbers[i];

            while (!stack.isEmpty() && numbers[stack.peekFirst()] < curr) {
                answer[stack.pop()] = curr;
            }

            stack.push(i);
        }
        while (!stack.isEmpty()) {
            answer[stack.pop()] = -1;
        }

        return answer;
    }

    /**
     * 같은 단조 스택 풀이지만, 오른쪽에서 왼쪽으로 보며 "뒷 큰수가 될 수 있는 값"만 남깁니다.
     * 현재 문제는 답으로 인덱스나 거리를 묻지 않고 실제 값을 묻기 때문에, 값만 쌓아도 충분합니다.
     */
    public int[] solutionWithValueStack(int[] numbers) {
        var answer = new int[numbers.length];
        var candidates = new int[numbers.length];
        var top = -1;

        for (var i = numbers.length - 1; i >= 0; i--) {
            var current = numbers[i];

            // current보다 작거나 같은 값은 current의 뒷 큰수가 될 수 없습니다.
            // 더 왼쪽 원소 입장에서도 current가 더 가까운 위치에서 그 후보들을 대신하거나 막기 때문에 버려도 됩니다.
            while (top >= 0 && candidates[top] <= current) {
                top--;
            }

            answer[i] = top >= 0 ? candidates[top] : -1;
            candidates[++top] = current;
        }

        return answer;
    }
}

public class Main260519 {
    public static void main(String[] args) {
        var solution = new Solution();

        var testCases = List.of(
            new TestCase<>(new int[]{2, 3, 3, 5}, new int[]{3, 5, 5, -1}),
            new TestCase<>(new int[]{9, 1, 5, 3, 6, 2}, new int[]{-1, 5, 6, 6, -1, -1}),
            new TestCase<>(new int[]{4, 4, 4, 4}, new int[]{-1, -1, -1, -1}),
            new TestCase<>(new int[]{1, 2, 3, 4}, new int[]{2, 3, 4, -1}),
            new TestCase<>(new int[]{4, 3, 2, 1}, new int[]{-1, -1, -1, -1}),
            new TestCase<>(new int[]{2, 2, 3, 2, 4}, new int[]{3, 3, 4, 4, -1}),
            new TestCase<>(new int[]{5, 1, 5, 2, 5, 3}, new int[]{-1, 5, -1, 5, -1, -1})
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input());
            var valueStackResult = solution.solutionWithValueStack(testCase.input());
            System.out.println(
                "input=" + valueToString(testCase.input())
                    + ", expected=" + valueToString(testCase.answer())
                    + ", solution=" + valueToString(result)
                    + ", valueStack=" + valueToString(valueStackResult)
            );
            assertEquals("solution", testCase, result);
            assertEquals("solutionWithValueStack", testCase, valueStackResult);
        }
    }

    private static void assertEquals(String label, TestCase<?, ?> testCase, Object actual) {
        if (!valuesEqual(testCase.answer(), actual)) {
            throw new AssertionError(
                label
                    + ": input=" + valueToString(testCase.input())
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
