package p42584.attempt260511;

import support.TestCase;

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
     *
     * @param prices 초 단위로 기록된 주식 가격이 담긴 배열
     * - 1 <= prices[i] <= 10,000
     * - 2 <= prices.length <= 100,000
     *
     * @return 가격이 떨어지지 않은 기간(초)
     */
    public int[] solution(int[] prices) {
        // [1, 2, 3, 2, 3]
        //  ㄴ----------- 안 떨어짐 4
        //     ㄴ-------- 안 떨어짐 3
        //        ㄴ--|   떨어짐   1
        //           ㄴ-- 안 떨어짐 1
        //              ㄴ 0

        // 일단 드는 생각은, 현재 요소를 다음 요소와 계속 비교하며 떨어지는 때를 찾는 겁니다.
        // 첨삭: 이 출발점 자체는 좋습니다. 다만 이 문제의 답은 "끝까지 현재값 이상인 칸의 개수"가 아니라
        // "처음으로 현재값보다 낮아지는 시점까지 몇 초가 지났는가"입니다.
        // 그래서 낮아지는 값을 만나면 그 1초까지 답에 포함하고 바로 멈춰야 합니다.
        var answer = new int[prices.length];
        for (var i = 0; i < prices.length; i++) {
            var curr = prices[i];
            // 3초 시점의 3은 4초 시점에 2로 떨어집니다. 그러면 1초간 가격이 떨어지지 않은 것으로 봅니다.
            // 그렇다면 기본적으로 항상 1초는 유지된다고 봅니다.
            // 첨삭: "기본 1초"로 보정하면 즉시 떨어지는 경우와 마지막 원소를 같은 모양으로 뭉개게 됩니다.
            // 더 안전한 기준은 seconds를 보정하는 것이 아니라, 비교한 다음 시점마다 1초를 먼저 더하고
            // prices[j] < curr가 되는 순간 break하는 것입니다.
            var seconds = 0;

            for (var j = i + 1; j < prices.length; j++) {
                // 다음 가격이 현재 가격보다 크거나 같다면 떨어지지 않은 것
                // 첨삭: 여기서는 "떨어지지 않은 값만 센다"가 아니라 "시간이 1초 흘렀다"를 세야 합니다.
                // 반례: [5, 1, 5, 5, 5]에서 첫 5의 답은 바로 다음 1에서 떨어지므로 1입니다.
                // 현재 방식은 뒤의 5, 5, 5를 세어서 첫 답을 3으로 만들 수 있습니다.
                if (prices[j] < curr) {
                    break;
                }
                seconds++;
            }
            if (seconds == 0) {
                seconds = 1;
            }
            answer[i] = seconds;
        }
        answer[prices.length - 1] = 0;

        return answer;
    }
}

public class Main260511 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        var testCases = List.of(
            new TestCase<>(new int[]{1, 2, 3, 2, 3}, new int[]{4, 3, 1, 1, 0}),
            new TestCase<>(new int[]{1, 1, 1, 1, 1}, new int[]{4, 3, 2, 1, 0}),
            new TestCase<>(new int[]{5, 4, 3, 2, 1}, new int[]{1, 1, 1, 1, 0})
            // 첨삭: 위 세 케이스는 현재 버그를 잡기 어렵습니다.
            // 떨어진 뒤 다시 회복하는 입력도 넣어 보세요.
            // new TestCase<>(new int[]{5, 1, 5, 5, 5}, new int[]{1, 3, 2, 1, 0})
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input());
            System.out.println(
                // 첨삭: 출력으로 true/false를 보는 방식은 실패를 놓치기 쉽습니다.
                // 이 연습 저장소에서는 expected와 actual이 다르면 AssertionError로 바로 멈추는 편이 더 좋습니다.
                Arrays.toString(result) + " == " + Arrays.toString(testCase.answer()) + ": " + Arrays.equals(solution.solution(testCase.input()), testCase.answer())
            );
        }
    }
}
