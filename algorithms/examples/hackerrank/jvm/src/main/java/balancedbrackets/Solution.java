package balancedbrackets;

import support.Judge;

/**
 * <a href="https://www.hackerrank.com/challenges/balanced-brackets/problem">Balanced Brackets</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 isBalanced 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 시그니처는 HackerRank 원문과 같게 맞췄습니다(String 반환, "YES"/"NO").
 */
class Solution {
    public String isBalanced(String s) {
        return "";
    }

    static void main() {
        var sol = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(sol.isBalanced("{[()]}"), "YES");
        Judge.check(sol.isBalanced("{[(])}"), "NO");
        Judge.check(sol.isBalanced("{{[[(())]]}}"), "YES");
        Judge.check(sol.isBalanced("]"), "NO");
        // 반례를 여기에 추가하세요:
    }
}
