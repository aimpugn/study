package main.java.p12906;

import java.util.*;

/**
 * @see <a href=
 *      "https://school.programmers.co.kr/learn/courses/30/lessons/1845">
 *      포켓몬</a>
 */
public class Main {

    public static void main(String[] args) {
        Main main = new Main();
        System.out.println(
                Arrays.toString(main.solution(new int[] {1, 1, 3, 3, 0, 1})));
    }

    /**
     * <pre>
     * '연속적으로 나타나는 숫자'는 하나만 남기고 전부 제거 제거된 후 남은 수들을 반환시 기존 순서 유지
     *
     * ex:
     * - [1, 1, 3, 3, 0, 1, 1] => [1,3,0,1]
     *    ^^^^ ^^^^^ ^^^^
     *     1     3    1
     * - [4, 4, 4, 3, 3] => [4, 3]
     *    ^^^^^^^  ^^^^
     *     4         3
     * </pre>
     *
     * @param arr 요소 in [0..9] 1 <= arr.length <= 1,000,000
     */
    public int[] solution(int[] arr) {
        ArrayList<Integer> answer = new ArrayList<>();

        // [1, 1, 3, 3, 0, 1, 1]
        //
        // before curr
        // 0   !=   1(추가)
        // 1   ==   1(제외)
        // 1   !=   3(추가)
        // 3   ==   3(제외)
        // 3   !=   0(추가)
        // 0   !=   1(추가)
        // 1   ==   1(제외)
        // [1, 3, 0, 1]
        int before = -1;
        for (int el : arr) {
            if (before != el) {
                answer.add(el);
            }
            before = el;
        }

        return answer.stream().mapToInt(el -> el).toArray();
    }
}
