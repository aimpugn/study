package jesseandcookies;

import support.Judge;

import java.util.List;

/**
 * <a href="https://www.hackerrank.com/challenges/jesse-and-cookies/problem">Jesse and Cookies</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 cookies 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 시그니처는 HackerRank 원문과 같게 맞췄습니다(int k, List&lt;Integer&gt; A 받아 int 반환).
 */
class Solution {
    public int cookies(int k, List<Integer> A) {
        return 0;
    }

    static void main() {
        var sol = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(sol.cookies(7, List.of(1, 2, 3, 9, 10, 12)), 2);
        Judge.check(sol.cookies(10, List.of(1, 1, 1)), -1);
        // 반례를 여기에 추가하세요:
    }
}
