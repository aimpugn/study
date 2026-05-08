package p12906;

import java.util.ArrayList;
import java.util.Arrays;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/12906">같은 숫자는 싫어</a>
 */
public class Main {

    public static void main(String[] args) {
        Main main = new Main();
        main.check(new int[]{1, 1, 3, 3, 0, 1, 1}, new int[]{1, 3, 0, 1});
        main.check(new int[]{4, 4, 4, 3, 3}, new int[]{4, 3});
    }

    private void check(int[] arr, int[] expected) {
        int[] actual = solution(arr);
        System.out.printf(
                "expected=%s actual=%s pass=%s%n",
                Arrays.toString(expected),
                Arrays.toString(actual),
                Arrays.equals(expected, actual));
    }

    /**
     * <pre>
     * '연속적으로 나타나는 숫자'는 하나만 남기고, 남은 수들을 반환할 때는 기존 순서를 유지합니다.
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
