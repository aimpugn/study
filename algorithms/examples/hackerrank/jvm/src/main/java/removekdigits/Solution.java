package removekdigits;

import support.Judge;

/**
 * Remove K Digits — 자릿수 k개를 지워 가장 작은 수 만들기 (LeetCode 402, String/Stack 라이브코딩 단골).
 * <a href="https://leetcode.com/problems/remove-k-digits/">문제</a>
 * <p>
 * 숫자 문자열 num에서 정확히 k개의 자리를 지워 만들 수 있는 가장 작은 수를 문자열로 반환합니다.
 * 남은 자리의 순서는 유지합니다. 결과에 선행 0이 생기면 제거하고, 전부 지워 비면 "0"을 반환합니다.
 * 예: "1432219", k=3 -> "1219"  ("4","3","2"를 하나씩 지움)
 * - 1 <= num.length <= 100,000 (10^5), num은 '0'-'9'로만 구성
 * - 1 <= k <= num.length
 * <p>
 * 막히면(10분 룰): 정확히 k개를 지우면 모든 후보의 자릿수가 n-k로 같아집니다. 길이가 같은 수는
 * "사전순으로 작다 = 숫자로 작다"라, 큰 수를 실제로 만들(BigInteger/BigDecimal) 필요가 없습니다 —
 * 왼쪽 자리부터 더 작게 만드는 게 씨앗입니다. "선행 0"과 "정확히 k개"를 어떻게 지킬지가 승부처.
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.removeKdigits("1432219", 3), "1219");  // 일반: 4,3,2를 지워 앞자리를 작게
        Judge.check(Result.removeKdigits("10200", 1), "200");     // 선행 0: "0200" -> 앞 0 제거
        Judge.check(Result.removeKdigits("10001", 1), "1");       // 선행 0 다수: "0001" -> "1"
        Judge.check(Result.removeKdigits("10", 2), "0");          // 전부 지움 -> 비면 "0"
        Judge.check(Result.removeKdigits("112", 1), "11");        // 같은 자리: 같으면 지우지 않기(과삭제 금지)
        Judge.check(Result.removeKdigits("12345", 2), "123");     // 이미 오름차순: 안 지워지고 남으면 뒤에서 자르기
        // 반례를 여기에 추가하세요:
    }
}

class Result {
    public static String removeKdigits(String num, int k) {
        // Write your code here
        return "";
    }
}
