package p12906;

import java.util.Arrays;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/12906">같은 숫자는 싫어</a>
 */
public class Main260508 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        // TODO: 풀이를 구현한 뒤 아래 예제의 기대 결과와 비교하세요.
        System.out.println(Arrays.toString(solution.solution(new int[]{1, 1, 3, 3, 0, 1, 1})));
        System.out.println(Arrays.toString(solution.solution(new int[]{4, 4, 4, 3, 3})));
    }

    static class Solution {
        /**
         * TODO: 연속으로 같은 숫자가 나오는 구간에서 대표 숫자 하나만 남기는 흐름으로 분해하세요.
         *
         * @param arr 숫자 배열
         * - 1 <= arr.length <= 1,000,000
         * - 0 <= arr[i] <= 9
         *
         * @return 연속 중복을 제거하되 원래 순서를 유지한 숫자 배열
         */
        public int[] solution(int[] arr) {
            // TODO: 직전에 답으로 채택한 숫자와 현재 숫자를 비교하세요.
            return new int[]{};
        }
    }
}
