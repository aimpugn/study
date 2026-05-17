package p42584.attempt260511;

import support.TestCase;

import java.util.ArrayDeque;
import java.util.Arrays;
import java.util.List;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42584?language=java">주식가격</a>
 * <p>
 * 준비 메타
 * - 공식 레벨: Lv.2
 * - 분류/트랙: Stage1/StackQueue
 * - 핵심 패턴: Monotonic Stack
 * - 체감 난이도: 체감 Lv.2 중상
 * - 예상 풀이 시간: 실전 숙련자 25~45분 / 학습 모드 70~100분
 * - 첫 접근: 가격이 떨어지는 첫 시점을 아직 답을 못 찾은 인덱스 스택으로 찾는다.
 */
class Solution {
    /**
     * @param prices 초 단위로 기록된 주식 가격이 담긴 배열
     * - 1 <= prices[i] <= 10,000
     * - 2 <= prices.length <= 100,000
     *
     * @return 가격이 떨어지지 않은 기간(초)
     */
    public int[] solution(int[] prices) {
        var answer = new int[prices.length];

        // 이 문제에서 스택에 넣는 것은 가격 자체가 아니라 "아직 답이 확정되지 않은 시점의 인덱스"입니다.
        // 어떤 인덱스 i의 답은 오른쪽에서 처음으로 prices[i]보다 낮은 가격을 만나는 순간 확정됩니다.
        // 반대로 그런 가격을 아직 못 만난 인덱스는 나중 가격을 더 봐야 하므로 스택에 잠시 보관합니다.
        //
        // [1, 2, 3, 2, 3]에서는 index 0, 1, 2가 차례로 스택에 쌓입니다.
        // index 3의 가격 2를 만나는 순간 index 2의 가격 3만 처음으로 떨어진 시점을 만나므로
        // answer[2] = 3 - 2 = 1이 됩니다. index 0의 1과 index 1의 2는 아직 떨어지지 않았으므로 남깁니다.
        var waitingIndexes = new ArrayDeque<Integer>();

        for (var currentIndex = 0; currentIndex < prices.length; currentIndex++) {
            var currentPrice = prices[currentIndex];

            // 현재 가격이 더 낮아졌다면, 스택 위쪽의 가격들은 지금 처음으로 하락 시점을 만난 것입니다.
            // 예를 들어 [3, 3, 1]에서 currentIndex가 2이고 currentPrice가 1이면,
            // index 1의 3도, index 0의 3도 지금 떨어진 시점이 확정됩니다.
            // 같은 가격은 떨어진 것이 아니므로 '>'일 때만 꺼냅니다.
            while (!waitingIndexes.isEmpty() && prices[waitingIndexes.peekLast()] > currentPrice) {
                var droppedIndex = waitingIndexes.pollLast();
                answer[droppedIndex] = currentIndex - droppedIndex;
            }

            // 아직 currentIndex의 답은 모릅니다.
            // 뒤에서 더 낮은 가격을 만나면 그때 확정되고, 끝까지 못 만나면 마지막 정리 단계에서 확정됩니다.
            waitingIndexes.offerLast(currentIndex);
        }

        // 끝까지 스택에 남은 인덱스는 마지막 시점까지 가격이 떨어지지 않은 경우입니다.
        // 따라서 마지막 인덱스까지 간 시간, 즉 lastIndex - index가 답입니다.
        var lastIndex = prices.length - 1;
        while (!waitingIndexes.isEmpty()) {
            var stableIndex = waitingIndexes.pollLast();
            answer[stableIndex] = lastIndex - stableIndex;
        }

        return answer;
    }
}

public class Main260511 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        var testCases = List.of(
            new TestCase<>(new int[]{1, 2, 3, 2, 3}, new int[]{4, 3, 1, 1, 0}),
            new TestCase<>(new int[]{1, 1, 1, 1, 1}, new int[]{4, 3, 2, 1, 0}),
            new TestCase<>(new int[]{5, 4, 3, 2, 1}, new int[]{1, 1, 1, 1, 0}),
            new TestCase<>(new int[]{3, 3, 1}, new int[]{2, 1, 0}),
            new TestCase<>(new int[]{5, 1, 5, 5, 5}, new int[]{1, 3, 2, 1, 0}),
            new TestCase<>(new int[]{2, 2, 2, 1}, new int[]{3, 2, 1, 0}),
            new TestCase<>(new int[]{1, 2}, new int[]{1, 0}),
            new TestCase<>(new int[]{2, 1}, new int[]{1, 0}),
            new TestCase<>(new int[]{2, 2}, new int[]{1, 0}),
            new TestCase<>(new int[]{1, 3, 2, 2, 1}, new int[]{4, 1, 2, 1, 0})
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input());
            assertEquals(testCase, result);
        }
    }

    private static void assertEquals(TestCase<int[], int[]> testCase, int[] actual) {
        if (!Arrays.equals(testCase.answer(), actual)) {
            throw new AssertionError(
                "prices=" + Arrays.toString(testCase.input())
                    + ", expected=" + Arrays.toString(testCase.answer())
                    + ", actual=" + Arrays.toString(actual)
            );
        }
    }
}
