package p154539;

import support.TestCase;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/154539">뒤에 있는 큰 수 찾기</a>
 * <p>
 * 준비 메타
 * - 공식 레벨: Lv.2
 * - 분류/트랙: Stage3/Stack
 * - 핵심 패턴: Monotonic Stack
 * - 체감 난이도: 체감 Lv.2 중~상
 * - 예상 풀이 시간: 실전 숙련자 25~45분 / 학습 모드 60~90분
 * - 첫 접근: 현재 값이 과거 후보의 답이 되는 순간을 찾아, 아직 더 큰 수를 못 찾은 인덱스를 스택에 둔다.
 */
class Solution {
    /**
     * numbers의 각 위치마다 오른쪽에 있는 수 중에서, 현재 수보다 크면서 가장 가까운 값을 찾습니다.
     * 그런 값이 없으면 -1을 답에 넣습니다.
     *
     * numbers.length는 최대 1,000,000까지 갈 수 있으므로,
     * 각 위치마다 오른쪽을 다시 훑는 O(n^2) 풀이는 먼저 제외하고 시작하는 편이 안전합니다.
     *
     * 42584 주식가격과 연결해서 보면 방향만 바뀝니다.
     * 42584에서는 현재 가격이 과거 인덱스의 "처음 떨어지는 시점"을 확정했습니다.
     * 여기서는 현재 숫자가 과거 인덱스의 "가장 가까운 큰 수"를 확정합니다.
     */
    public int[] solution(int[] numbers) {
        // TODO: answer를 먼저 -1로 채워 두면, 끝까지 더 큰 수를 못 찾은 위치를 따로 처리하지 않아도 됩니다.
        // TODO: 스택에는 값만 넣기보다 "아직 답을 못 찾은 인덱스"를 넣는 쪽을 먼저 생각해 보세요.
        // TODO: 현재 numbers[i]가 스택 맨 위 인덱스의 값보다 클 때, 그 인덱스의 답이 왜 지금 확정되는지 손으로 추적해 보세요.
        // TODO: 풀이를 구현하세요.
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
