package p42576;

import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42576">완주하지 못한 선수</a>
 */
class Solution {
    /**
     * TODO: 문제 조건, 예제, 제약을 먼저 작은 문장으로 분해하세요.
     */
    public String solution(String[] participant, String[] completion) {
        // TODO: 풀이를 구현하세요.
        return "";
    }
}

public class Main {
    public static void main(String[] args) {
        Solution solution = new Solution();

        check(
                solution,
                new String[]{"leo", "kiki", "eden"},
                new String[]{"eden", "kiki"},
                "leo"
        );
        check(
                solution,
                new String[]{"marina", "josipa", "nikola", "vinko", "filipa"},
                new String[]{"josipa", "filipa", "marina", "nikola"},
                "vinko"
        );
        check(
                solution,
                new String[]{"mislav", "stanko", "mislav", "ana"},
                new String[]{"stanko", "ana", "mislav"},
                "mislav"
        );
    }

    private static void check(Solution solution, String[] participant, String[] completion, String expected) {
        String actual = solution.solution(participant, completion);
        System.out.printf("expected=%s actual=%s pass=%s%n", expected, actual, Objects.equals(expected, actual));
    }
}
