package main.java.p12909;

import java.util.*;

public class Main {
    /**
     * ex:
     * - 올바른 괄호
     *     ()()
     *     (())()
     *
     * - 올바르지 않은 괄호
     *     )()(
     *     (()(
     *
     *
     * @param s "("와 ")"로 이루어진 문자열
     *          1 <= s.length <= 100,000
     * @return 올바른 괄호인지 여부
     */
    boolean solution(String s) {

        // "("로 열리고 ")"로 닫여야 합니다.
        // "("가 있는 만큼 ")"가 있어야 합니다.
        // 쌍을 이뤄야 하므로, 홀수여도 안됩니다.
        //
        // 열린만큼 다시 닫혀야 합니다.
        // 이를 체크하기 위해 괄호가 열릴 때마다 닫힘 기호를 스택에 쌓고
        // `s`에서 닫힘 기호를 만날 때마다 해당 스택에 ')'가 있는지를 판단합니다(LIFO).
        // s    : ( )         (    )
        //          ^              ^
        //          stack에 존재     stack에 존재
        // stack: )           )
        //       add() pop() add() pop()
        //
        int length = s.length();
        if (s.charAt(0) == ')' || s.charAt(length - 1) == '(') {
            return false;
        }

        if (length % 2 != 0) {
            return false;
        }

        Stack<Character> closeParentheses = new Stack<>();
        for (char c : s.toCharArray()) {
            if (c == '(') {
                closeParentheses.add(')');
            } else if (!closeParentheses.isEmpty() // java.util.EmptyStackException 방지
                    && closeParentheses.pop() != ')') {
                return false;
            }
        }

        return closeParentheses.isEmpty();
    }
}
