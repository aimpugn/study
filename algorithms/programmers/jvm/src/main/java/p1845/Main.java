package main.java.p1845;

import java.util.*;

/**
 * @see <a href=
 *      "https://school.programmers.co.kr/learn/courses/30/lessons/1845">
 *      포켓몬</a>
 */
public class Main {

    public static void main(String[] args) {
        assert Main.solution(new int[] {3, 1, 2, 3}) == 2 : "Wrong when 1st";
        assert Main
                .solution(new int[] {3, 3, 3, 2, 2, 4}) == 2 : "Wrong when 2nd";
        assert Main
                .solution(new int[] {3, 3, 3, 2, 2, 2}) == 3 : "Wrong when 3rd";
        assert Main
                .solution(new int[] {1, 2, 3, 3, 4, 5}) == 3 : "Wrong when 4th";
        System.out.println("Success");
    }

    public static int solution(int[] nums) {
        // 골라야 하는 포켓몬 수는 정해져 있다.
        // if 6 => 3
        // if 7 => 3
        HashMap<Integer, Integer> calculator = new HashMap<>();

        // 선택 방법은 중요하지 않다.
        // [1, 2, 3, 3, 4, 5].length = 6, 3개 선택 가능
        // - 123을 뽑든
        // - 234를 뽑든
        //
        // 유니크한 수의 개수 & numToPick 이하
        //
        int numToPick = nums.length / 2;
        int idx = 0;
        for (int num : nums) {
            if (!calculator.containsKey(num)) {
                idx++;
            }
            calculator.merge(num, calculator.getOrDefault(num, 1),
                    Integer::sum);
            if (numToPick <= idx) {
                return idx;
            }
        }

        return calculator.keySet().size();
    }

}
