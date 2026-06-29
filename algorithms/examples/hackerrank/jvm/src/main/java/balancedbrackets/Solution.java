package balancedbrackets;

import support.Judge;

import java.util.ArrayDeque;
import java.util.List;

/**
 * <a href="https://www.hackerrank.com/challenges/balanced-brackets/problem">Balanced Brackets</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 isBalanced 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 시그니처는 HackerRank 원문과 같게 맞췄습니다(String 반환, "YES"/"NO").
 */
class Solution {

    /*
     * Complete the 'isBalanced' function below.
     *
     * The function is expected to return a STRING.
     * The function accepts STRING s as parameter.
     *
     * 괄호: (, ), {, }, [, or ].
     * 여는 괄호(`(`, `[`, `{`)가 닫는 괄호(), ], or }) 왼쪽에 위치하고 같은 타입인 두 괄호는 한 쌍.
     * - [], {}, ()
     * 이러한 쌍이 매칭되지 않으면 not balanced: {[(])}
     * {와 } 사이가 balanced가 아님
     *
     * balanced 여부 조건
     * - It contains no unmatched brackets.
     * - 부분 집합도 쌍이어야 함
     *
     * @return "YES" || "NO"
     */

    /*
     * Complete the 'isBalanced' function below.
     *
     * The function is expected to return a STRING.
     * The function accepts STRING s as parameter.
     * balanced
     * - It contains no unmatched brackets.
     * @return "YES" || "NO"
     */
    public static String isBalanced(String s) {
        var sLen = s.length();
        if (sLen == 0 || sLen % 2 != 0) return "NO";

        var stack = new ArrayDeque<Character>();

        for (var ch : s.toCharArray()) {
            if (List.of('(', '{', '[').contains(ch)) {
                stack.push(ch);
            } else {
                if (stack.peek() == null) {
                    return "NO";
                }
                var leftBracket = stack.pop();
                if (leftBracket == '(' && ch != ')') {
                    return "NO";
                }
                if (leftBracket == '{' && ch != '}') {
                    return "NO";
                }
                if (leftBracket == '[' && ch != ']') {
                    return "NO";
                }
            }
        }
        // System.out.println("s: " + s + ", stack: " + stack);

        // 첨삭: 정답입니다 — 단 한 곳, 이 마지막 return이 간극이었고 지금은 stack.isEmpty()로 정확히
        //   닫으셨습니다(수정 A 적용 확인). 18/21이 통과하고 3개만 실패했던 건 접근(스택으로 짝 맞추기)이
        //   옳다는 증거였고, 빠졌던 건 "끝났을 때의 조건" 하나뿐이었습니다.
        //
        //   균형의 조건은 둘입니다. (1) 닫는 괄호가 올 때마다 가장 최근 여는 괄호와 짝이 맞아야 하고(위
        //   for 안에서 이미 검사함), (2) 다 돌고 났을 때 열린 채 남은 괄호가 없어야 합니다(스택이 비어야
        //   함). 이 풀이는 (1)만 보고 (2)를 빠뜨려서, 여는 괄호만 남아도 그대로 "YES"를 돌려줍니다.
        //
        //   실패 3건의 정체는 위 홀수-길이 가드(sLen % 2 != 0)와 연결됩니다. 그 가드가 "남는 경우는
        //   처리됐다"는 착각을 주지만, 실제로는 홀수 길이만 거릅니다. 짝수 길이인데 여는 괄호만 남는
        //   입력 — "((((", "{}((", "(())((" 같은 — 은 가드를 통과하고, 닫기가 오지 않으니 (1)의 짝
        //   검사에도 안 걸려, 그대로 끝의 "YES"로 샜습니다. (아래 main에 이 세 입력을 넣어 뒀습니다.)
        //
        //   데이터(스택)가 어떻게 움직이는지 눈으로 — 처리 순서대로, 스택은 밑->위로 적습니다.
        //
        //   통과하는 "{[()]}" (끝에 스택이 빔 -> YES):
        //     입력  동작        스택(밑->위)
        //     {     push {      {
        //     [     push [      { [
        //     (     push (      { [ (
        //     )     pop ( (짝)  { [
        //     ]     pop [ (짝)  {
        //     }     pop { (짝)  (빔)
        //
        //   실패 유형 "(()(" (끝에 스택이 안 빔 -> NO여야 함):
        //     입력  동작        스택(밑->위)
        //     (     push (      (
        //     (     push (      ( (
        //     )     pop ( (짝)  (
        //     (     push (      ( (
        //     끝    --          ( (   <- 열린 ( 둘이 남음
        //
        //   둘째 표가 핵심입니다. 루프 안의 짝 검사(1)는 한 번도 안 깨져 통과하고, 끝에 남은 "( ("를
        //   잡는 건 오직 (2) "스택이 비었는가" 검사뿐입니다. 옛 코드는 그 자리에서 무조건 YES였습니다.
        //
        //   고친 건 한 줄입니다: 끝에서 무조건 "YES"가 아니라 "스택이 비었으면 YES, 아니면 NO"(86줄).
        //
        // > 카드: 스택 짝맞추기는 "루프 중 짝 검사"만으로 안 끝난다 — 끝에서 스택이 비었는지(열린 채 남은 게 없는지)까지 봐야 균형이다. 종료 후 상태도 정답 조건의 일부.
        return stack.isEmpty() ? "YES" : "NO";
    }

    // 첨삭 — 학습자 접근을 고친 동작 버전(핵심은 마지막 줄 하나):
    //   기존 구조를 그대로 두고 끝의 return만 "스택이 비었는지"로 바꾸면 끝납니다. peek() == null도
    //   동작하지만 빈 스택 판정은 isEmpty()가 의도가 더 분명해 같이 정리했습니다(이건 가독성, 버그 아님).
    //
    //     public static String isBalanced(String s) {
    //         var sLen = s.length();
    //         if (sLen == 0 || sLen % 2 != 0) return "NO";   // (선택) 조기 종료 — 끝의 검사가 어차피 다 잡음
    //         var stack = new ArrayDeque<Character>();
    //         for (var ch : s.toCharArray()) {
    //             if (List.of('(', '{', '[').contains(ch)) {
    //                 stack.push(ch);
    //             } else {
    //                 if (stack.isEmpty()) return "NO";       // 닫기인데 열린 게 없음
    //                 var leftBracket = stack.pop();
    //                 if (leftBracket == '(' && ch != ')') return "NO";
    //                 if (leftBracket == '{' && ch != '}') return "NO";
    //                 if (leftBracket == '[' && ch != ']') return "NO";
    //             }
    //         }
    //         return stack.isEmpty() ? "YES" : "NO";          // <- 핵심 수정: 열린 채 남았으면 NO
    //     }
    //
    // 첨삭 — 베스트 프랙티스(열 때 '기대하는 닫기'를 미리 적어둔다):
    //   한 깨달음에서 코드가 따라 나옵니다 — 여는 괄호를 보면, 언젠가 와야 할 "닫는 괄호"를 스택에 적어둡니다.
    //   그러면 닫는 괄호가 왔을 때 할 일은 하나입니다: 스택 맨 위(가장 최근에 적어둔 기대)와 같은가?
    //   다르거나 스택이 비었으면 실패. 다 돌고 스택이 남으면 기대한 닫기가 안 온 것이라 실패. 이렇게 하면
    //   leftBracket을 꺼내 세 번 비교하던 짝 검사가 stack.pop() != ch 한 줄로 줍니다.
    //
    //     public static String isBalanced(String s) {
    //         var stack = new ArrayDeque<Character>();
    //         for (var ch : s.toCharArray()) {
    //             switch (ch) {
    //                 case '(' -> stack.push(')');   // 열면 기대하는 닫기를 적어둠
    //                 case '{' -> stack.push('}');
    //                 case '[' -> stack.push(']');
    //                 default  -> {                  // 닫기: 맨 위 기대와 같아야 함
    //                     if (stack.isEmpty() || stack.pop() != ch) return "NO";
    //                 }
    //             }
    //         }
    //         return stack.isEmpty() ? "YES" : "NO";
    //     }
    //
    //   복잡도 (면접에서 말하는 4박자):
    //   1) n = 문자열 길이로 두면,
    //   2) 시간: 각 문자를 한 번 보고 push 또는 pop 한 번(둘 다 O(1)) -> O(n).
    //   3) 공간: 스택이 최대로 커지는 건 전부 여는 괄호일 때라 최악 O(n).
    //   4) 개선 여지: 짝을 맞추려면 "가장 최근 열기"를 알아야 하고 그게 곧 스택(LIFO)이라 더 줄이긴 어렵습니다.
    //
    // > 카드: 괄호 짝맞추기 = 열 때 '기대하는 닫기'를 push -> 닫을 때 top과 같은지 -> 끝에서 스택 비었는지. 시간 O(n), 공간 O(n).

    static void main() {
        var sol = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(sol.isBalanced("{[()]}"), "YES");
        Judge.check(sol.isBalanced("{[(])}"), "NO");
        Judge.check(sol.isBalanced("{{[[(())]]}}"), "YES");
        Judge.check(sol.isBalanced("]"), "NO");
        //
        Judge.check(sol.isBalanced("{}"), "YES");
        Judge.check(sol.isBalanced("}([[{)[]))]{){}["), "NO");
        Judge.check(sol.isBalanced("{]]{()}{])"), "NO");
        Judge.check(sol.isBalanced("(){}"), "YES");
        Judge.check(sol.isBalanced("{}{()}{{}}"), "YES");
        // 첨삭: HackerRank에서 실패한 유형 — 여는 괄호만 남는 짝수 길이 입력입니다.
        //   지금 코드(끝에서 무조건 YES)로는 아래에서 AssertionError가 나며 그 3건이 "재현"됩니다.
        //   isBalanced의 마지막 return을 stack.isEmpty() 기준으로 고치면 셋 다 green이 됩니다.
        Judge.check(sol.isBalanced("(((("), "NO");
        Judge.check(sol.isBalanced("{}(("), "NO");
        Judge.check(sol.isBalanced("(())(("), "NO");
        // 반례를 여기에 추가하세요:
    }
}
