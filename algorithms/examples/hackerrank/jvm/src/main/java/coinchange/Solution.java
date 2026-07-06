package coinchange;

import support.Judge;

/**
 * Coin Change — 최소 동전 수 (LeetCode 322, 대표 DP·라이브코딩 단골).
 * <a href="https://leetcode.com/problems/coin-change/">문제</a>
 * <p>
 * 동전 종류 coins(무한히 사용 가능)로 금액 amount를 만드는 데 필요한 최소 동전 개수를 반환합니다.
 * 어떤 조합으로도 만들 수 없으면 -1을 반환합니다. amount가 0이면 0개입니다.
 * 예: coins=[1,2,5], amount=11 -> 3  (5 + 5 + 1)
 * - 1 <= coins.length <= 12, 1 <= coins[i]
 * - 0 <= amount <= 10,000 (10^4)
 * <p>
 * 막히면(10분 룰): 코드보다 먼저 값 문장을 고정하세요 — dp[a] = "금액 a를 만드는 최소 동전 수".
 * 그러면 "마지막에 어떤 동전 c를 썼나?"를 물어 dp[a] = min over c of dp[a - c] + 1 이 따라 나옵니다.
 * "만들 수 없음"을 어떤 값으로 표시할지가 승부처. 그리디(큰 동전부터)는 [1,3,4]로 6 만들기에서
 * 깨집니다(4+1+1=3 vs 3+3=2) — 왜 깨지는지가 DP가 필요한 이유입니다.
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.coinChange(new int[]{1, 2, 5}, 11), 3);  // 일반: 5+5+1
        Judge.check(Result.coinChange(new int[]{1, 3, 4}, 6), 2);   // 그리디 함정: 4+1+1(3) 아니라 3+3(2)
        Judge.check(Result.coinChange(new int[]{2}, 3), -1);        // 만들 수 없음 -> -1
        Judge.check(Result.coinChange(new int[]{1, 2, 5}, 0), 0);   // 금액 0 -> 0개
        Judge.check(Result.coinChange(new int[]{5}, 5), 1);         // 단일 동전 딱 맞음
        Judge.check(Result.coinChange(new int[]{5}, 3), -1);        // 단일 동전 불가능
        // 반례를 여기에 추가하세요:
    }
}

class Result {
    public static int coinChange(int[] coins, int amount) {
        // Write your code here
        return -2;
    }
}
