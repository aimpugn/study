package p12906;

import java.util.ArrayList;
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
         * - `arr` 배열에서 연속으로 나타나는 숫자는 하나만 남기고 전부 제거
         * - 제거된 후 남은 수들을 반환할 때는 배열 arr의 원소들의 순서 유지
         * <p>
         * 예를 들어:
         * <pre>{@code
         * [1, 1, 3, 3, 0, 1, 1] -> [1, 3, 0, 1]
         *    (X)   (X)      (X)
         * [4, 4, 4, 3, 3] -> [4, 3]
         *    (X)(X)   (X)
         * }</pre>
         *
         * @param arr 숫자 배열
         * - 1 <= arr.length <= 1,000,000
         * - 0 <= arr[i] <= 9
         *
         * @return 연속 중복을 제거하되 원래 순서를 유지한 숫자 배열
         */
        public int[] solution(int[] arr) {
            var answer = new ArrayList<Integer>();

            int prev = -1;
            for (var num : arr) {
                // 최초인 경우
                if (prev == -1) {
                    prev = num;
                    answer.add(num);
                    continue;
                }
                if (prev == num) {
                    continue;
                }

                prev = num;
                answer.add(num);
            }

            return answer.stream().mapToInt(Integer::intValue).toArray();
        }
    }
}
