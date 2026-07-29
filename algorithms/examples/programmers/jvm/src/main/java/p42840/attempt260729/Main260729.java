package p42840.attempt260729;

import support.Judge;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42840?language=java">모의고사</a>
 */
class Solution {
    public int[] solution(int[] answers) {
        throw new UnsupportedOperationException("TODO: 풀이를 구현하세요.");
    }
}

public class Main260729 {
    public static void main(String[] args) {
        var solution = new Solution();
        Judge.check(solution.solution(new int[]{1, 2, 3, 4, 5}), new int[]{1});
        Judge.check(solution.solution(new int[]{1, 3, 2, 4, 2}), new int[]{1, 2, 3});
    }
}
