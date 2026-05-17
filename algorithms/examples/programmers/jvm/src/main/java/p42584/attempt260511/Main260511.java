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
 * - 핵심 패턴: 단조 스택(Monotonic Stack)
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

        // 이 문제는 "뒤에 현재 가격보다 크거나 같은 값이 몇 개 있나"를 세는 문제가 아닙니다.
        // 각 시점마다 "처음으로 가격이 떨어지는 순간이 언제인가"를 찾는 문제입니다.
        // 한 번 떨어진 순간을 만나면 그 인덱스의 답은 바로 확정되고, 그 뒤 가격은 더 볼 필요가 없습니다.
        //
        // 그래서 스택에는 가격 자체가 아니라 "아직 답이 확정되지 않은 시점의 인덱스"를 넣습니다.
        // 인덱스를 넣어야 나중에 하락 시점을 만났을 때 currentIndex - droppedIndex로 몇 초가 지났는지 계산할 수 있습니다.
        // 가격만 저장하면 답을 적어야 할 answer 위치와 지난 시간을 알 수 없습니다.
        //
        // 여기서 쓰는 단조 스택(Monotonic Stack)은 스택 안의 값이 한쪽 방향으로만 정리되게 유지하는 방식입니다.
        // 이 풀이에서는 스택에 남아 있는 인덱스들의 가격이 아래에서 위로 갈수록 낮아지지 않게 유지됩니다.
        // 새 가격이 더 낮게 들어오면, 그 새 가격 때문에 답이 확정되는 인덱스들을 위에서부터 꺼냅니다.
        //
        // [1, 2, 3, 2, 3]을 손으로 따라가면 이렇게 움직입니다.
        //
        // currentIndex=0, price=1
        //   아직 비교할 이전 가격이 없으므로 index 0을 스택에 둡니다. stack=[0]
        //
        // currentIndex=1, price=2
        //   2는 1보다 낮지 않으므로 index 0의 답은 아직 모릅니다. index 1을 쌓습니다. stack=[0, 1]
        //
        // currentIndex=2, price=3
        //   3도 앞 가격들을 떨어뜨리지 못합니다. index 2를 쌓습니다. stack=[0, 1, 2]
        //
        // currentIndex=3, price=2
        //   2는 index 2의 가격 3보다 낮습니다. index 2는 지금 처음으로 떨어진 시점을 만났습니다.
        //   answer[2] = 3 - 2 = 1입니다. index 0의 1과 index 1의 2는 아직 떨어지지 않았으므로 남깁니다.
        //   그런 뒤 index 3도 나중에 답을 찾아야 하므로 스택에 둡니다. stack=[0, 1, 3]
        var waitingIndexes = new ArrayDeque<Integer>();

        for (var currentIndex = 0; currentIndex < prices.length; currentIndex++) {
            var currentPrice = prices[currentIndex];

            // 스택 맨 위에는 아직 답이 확정되지 않은 인덱스 중 가장 최근 인덱스가 있습니다.
            // 현재 가격이 그 인덱스의 가격보다 낮으면, 그 인덱스는 바로 지금 가격이 떨어진 것입니다.
            // 이때는 스택에서 꺼내고, "현재 시점 - 그 인덱스"를 답으로 적습니다.
            //
            // 같은 가격은 떨어진 것이 아닙니다.
            // 그래서 prices[waitingIndexes.peekLast()] >= currentPrice가 아니라
            // prices[waitingIndexes.peekLast()] > currentPrice일 때만 꺼냅니다.
            //
            // [3, 3, 1]에서 currentIndex=2, currentPrice=1을 만나면
            // index 1의 3도, index 0의 3도 지금 처음으로 낮은 가격을 만납니다.
            // 그래서 while이 한 번이 아니라 여러 번 돌 수 있습니다.
            while (!waitingIndexes.isEmpty() && prices[waitingIndexes.peekLast()] > currentPrice) {
                var droppedIndex = waitingIndexes.pollLast();
                answer[droppedIndex] = currentIndex - droppedIndex;
            }

            // 현재 시점의 답은 아직 모릅니다.
            // 지금보다 오른쪽에서 더 낮은 가격을 만나면 그때 답이 확정됩니다.
            // 끝까지 그런 가격을 못 만나면 마지막 시점까지 버틴 것이므로, 아래 정리 단계에서 답을 채웁니다.
            waitingIndexes.offerLast(currentIndex);
        }

        // 반복문이 끝났는데도 스택에 남은 인덱스들은 끝까지 가격이 떨어지지 않은 경우입니다.
        // 예제 [1, 2, 3, 2, 3]에서는 index 0, 1, 3, 4가 여기에 남습니다.
        // 이 인덱스들의 답은 "마지막 인덱스까지 간 시간"이므로 lastIndex - index입니다.
        // 마지막 인덱스 자신은 lastIndex - lastIndex = 0이 되어 자연스럽게 0초가 됩니다.
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
