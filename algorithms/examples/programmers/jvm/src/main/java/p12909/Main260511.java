package p12909;

import java.util.List;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/12909?language=java">올바른 괄호</a>
 */
class Solution {
    /**
     * 괄호 문자열은 전체 `(` 개수와 `)` 개수가 같다고 해서 곧바로 올바르지 않습니다.
     * 왼쪽에서 오른쪽으로 읽는 모든 순간에, 아직 닫히지 않은 `(` 개수가 음수가 되지 않아야 합니다.
     * 이 조건을 prefix invariant라고 부를 수 있습니다. prefix는 "문자열의 앞에서부터 여기까지 읽은 부분"입니다.
     * <p>
     * 이번 풀이에서 헤맨 지점은 상태를 "최종 개수"로만 본 데 있었습니다.
     * 예를 들어 "())(()"는 전체 개수만 보면 `(` 3개, `)` 3개라서 마지막 상태는 0입니다.
     * 하지만 세 번째 문자까지 읽은 순간 이미 닫을 `(`가 없는데 `)`가 나옵니다.
     *
     * <pre>
     * s        (  )  )  (  (  )
     * open     1  0 -1
     *                여기서 이미 실패. 뒤에 `(`가 나와도 앞의 잘못된 `)`를 복구할 수 없다.
     * </pre>
     *
     * 그래서 이 문제의 상태는 스택 자체가 아니라 "아직 닫히지 않은 열린 괄호 수"입니다.
     * 스택에 같은 값 `)`만 계속 넣고 뺀다면, 실제로 필요한 정보는 stack.size() 하나입니다.
     * 여기서는 그 크기를 `parenCnt` 정수 하나로 들고 갑니다.
     * <p>
     * 효율성에서 한 번 더 배운 점은, 명백히 불가능한 입력을 긴 루프 전에 버릴 수 있다는 것입니다.
     * 길이가 홀수인 문자열, 첫 글자가 `)`, 마지막 글자가 `(`인 문자열은 올바른 괄호가 될 수 없습니다.
     * 이 guard는 정답의 본질은 아니지만, 실패가 확정된 입력을 빨리 잘라내는 방어선입니다.
     *
     * @param s `(` 또는 `)`로만 이루어진 문자열
     * - s.length <= 100,000
     *
     * @return 올바른 괄호이면 true, 아니면 false
     */
    public boolean solution(String s) {
        var sLen = s.length();

        // 짝을 만들어야 하므로 길이가 홀수이면 어떤 순서로 읽어도 성공할 수 없습니다.
        // 첫 글자가 `)`인 경우는 바로 닫을 괄호가 없고, 마지막 글자가 `(`인 경우는 마지막에 새로 열린 괄호를 닫을 문자가 없습니다.
        if (sLen % 2 != 0) return false;
        if (s.charAt(0) == ')' || s.charAt(sLen - 1) == '(') {
            return false;
        }

        var parenCnt = 0;
        for (int i = 0; i < sLen; i++) {
            if (s.charAt(i) == '(') {
                parenCnt++;
            } else {
                // `)`는 지금까지 열린 `(` 하나를 소비합니다.
                // 소비할 열린 괄호가 없으면, 이후 문자가 무엇이든 이미 올바른 괄호가 될 수 없습니다.
                if (parenCnt == 0) {
                    return false;
                }
                parenCnt--;
            }
        }

        // 끝까지 prefix invariant를 지켰더라도, 아직 닫히지 않은 `(`가 남아 있으면 실패입니다.
        return parenCnt == 0;
    }
}

record TestCase(String input, boolean answer) {
}

public class Main260511 {
    public static void main(String[] args) {
        var solution = new Solution();

        // System.out.println(Arrays.toString(solution.solution(...)));
        var testCases = List.of(
            new TestCase("()()", true),
            new TestCase("(())()", true),
            new TestCase(")()(", false),
            new TestCase("(()(", false),
            new TestCase("())(()", false),
            new TestCase("())())", false),
            new TestCase("((((()))", false)
        );

        for (var testCase : testCases) {
            assertSolution(solution, testCase);
        }
    }

    private static void assertSolution(Solution solution, TestCase testCase) {
        boolean actual = solution.solution(testCase.input());
        if (actual != testCase.answer()) {
            throw new AssertionError(
                "input=%s, expected=%s, actual=%s".formatted(testCase.input(), testCase.answer(), actual)
            );
        }
    }
}
