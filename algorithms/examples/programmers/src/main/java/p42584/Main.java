package p42584;

import java.util.ArrayList;
import java.util.Arrays;

/**
 * [주식가격](https://school.programmers.co.kr/learn/courses/30/lessons/42584?language=java)
 */
class Solution {
    /**
     * @param prices 초 단위로 기록된 주식 가격이 담긴 배열
     * - 1 <= prices[i] <= 10,000
     * - 2 <= prices.length <= 100,000
     *
     * @return 가격이 떨어지지 않은 기간은 몇 초인지 리턴
     */
    public int[] solution(int[] prices) {
        //  0  1  2  3  4
        // [1, 2, 3, 2, 3]
        // - prices[0]=1 => 4: 끝까지 떨어지지 않습니다.
        // - prices[1]=2 => 3: 끝까지 떨어지지 않습니다.
        // - prices[2]=3 => 1: 1초 뒤에 가격이 떨어지므로, 1초간 가격이 떨어지지 않은 것으로 봅니다.
        // - prices[3]=2 => 1: 1초간 가격이 떨어지지 않습니다.
        // - prices[4]=3 => 0: 0초간 가격이 떨어지지 않습니다.
        //
        // [4, 3, 1, 1, 0]
        //
        // prices[0] 경우 prices[4]까지 가격이 떨어지지 않았고, 멈추지 않고 도달한 인덱스(4) - 0 = 4
        // prices[1] 경우 prices[4]까지 가격이 떨어지지 않았고, 멈추지 않고 도달한 인덱스(4) - 1 = 3
        // prices[2] 경우 prices[3]까지 가격이 떨어지지 않았고, 멈추지 않고 도달한 인덱스(3) - 2 = 1
        // prices[3] 경우 prices[4]까지 가격이 떨어지지 않았고, 멈추지 않고 도달한 인덱스(4) - 3 = 1
        // prices[4] 경우 마지막이므로 떨어지는지 여부를 측정할 수 없으므로 0

        int[] answer = new int[prices.length];
        int i = 0;
        int pricesLength = prices.length;
        int lastIdx = pricesLength - 1;
        while (i < pricesLength) {
            int currPrice = prices[i];
            int maybePriceDownIdx = lastIdx;
            for (int j = i + 1; j < pricesLength; j++) {
                // 현재 가격보다 낮은 가격을 만나면 멈추고 j가 가격 하락 지점에 해당합니다.
                if (prices[j] < currPrice) {
                    maybePriceDownIdx = j;
                    break;
                }
            }
            answer[i] = maybePriceDownIdx - i;

            i++;
        }

        return answer;
    }
}

class TestCase {
    int[] prices;
    int[] answer;

    TestCase(int[] prices, int[] answer) {
        this.prices = prices;
        this.answer = answer;
    }
}

class Main {
    public static void main(String[] args) {
        Solution s = new Solution();

        ArrayList<TestCase> testCases = new ArrayList<>();
        testCases.add(new TestCase(
                new int[]{1, 2, 3, 2, 3},
                new int[]{4, 3, 1, 1, 0}
        ));
        testCases.add(new TestCase(
                new int[]{1, 1, 1, 1, 1},
                new int[]{4, 3, 2, 1, 0}
        ));
        testCases.add(new TestCase(
                new int[]{5, 4, 3, 2, 1},
                new int[]{1, 1, 1, 1, 0}
        ));

        for (TestCase tc : testCases) {
            int[] result = s.solution(tc.prices);
            System.out.printf("%s: %b\n", Arrays.toString(result), Arrays.compare(result, tc.answer));
        }
    }
}
