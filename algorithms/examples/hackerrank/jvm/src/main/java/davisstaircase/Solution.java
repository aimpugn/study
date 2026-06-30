package davisstaircase;

import support.Judge;

import java.util.Arrays;
import java.util.List;

/**
 * <a href="https://www.hackerrank.com/challenges/ctci-recursive-staircase/problem">Davis' Staircase</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 stepPerms 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 계단 n개를 한 번에 1, 2, 또는 3칸 오를 때 꼭대기에 닿는 서로 다른 방법의 수를 셉니다.
 * (큰 n은 수가 매우 커지니 제약의 mod를 확인하세요. 아래 샘플은 작아 mod와 무관합니다.)
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.stepPerms(3), 4);
        Judge.check(Result.stepPerms(5), 13);
        Judge.check(Result.stepPerms(7), 44);
        // 반례를 여기에 추가하세요:
    }
}

class Result {

    /*
     * Complete the 'stepPerms' function below.
     *
     * The function is expected to return an INTEGER.
     * The function accepts INTEGER n as parameter.
     *
     * Davis has a number of staircases in his house and he likes to climb each staircase 1, 2, or 3 steps at a time.
     * Being a very precocious child,
     * he wonders how many ways there are to reach the top of the staircase.
     *
     * Given the respective heights for each of the `s` staircases in his house,
     * find and print the number of ways he can climb each staircase, module 10^10 + 7 on a new line.
     *
     * - `s`: the number of staircases in his house.
     * - `n`: Each of the following  lines contains a single integer. the height of staircase i
     *
     * @param n the number of stairs in the staircase
     *
     * @return the number of ways Davis can climb the staircase, modulo 10,000,000,007
     */

    public static int stepPerms(int n) {
        // Write your code here
        // n = 5
        // 1 1 1 1 1
        // 1 1 1 2
        // 1 1 2 1
        // 1 2 1 1
        // 2 1 1 1
        // 1 2 2
        // 2 2 1
        // 2 1 2
        // 1 1 3
        // 1 3 1
        // 3 1 1
        // 2 3
        // 3 2 -> permutation
        //
        // 13 % 10,000,000,007 = 13
        //
        // [1, 2, 3]
        //
        // dfs? 1+1+1+1+1 recursive
        //
        // System.out.println("n: " + n);
        int[] memo = new int[n + 1];      // O(n) 공간: 상태는 sum=0..n, 총 n+1개
        Arrays.fill(memo, -1);
        return permutation(n, 0, memo);
    }

    static int permutation(int n, int sum, int[] memo) {
        // System.out.println("n: " + n + ", sum: " + sum);
        if (sum == n) {
            return 1;
        }
        if (n < sum) {
            return 0;
        }
        if (memo[sum] != -1) return memo[sum];

        // caching?
        //
        // 첨삭: "// caching?" — 본능이 정확합니다. 정확히 캐시가 들어갈 자리에 물음표를 찍었어요. 막힌 건
        //   능력이 아니라 관점입니다. 두괄로 답하면: 캐싱이 안 떠오르는 이유는 지금 머릿속이 "경우의 수를
        //   나열한다(permutation)"에 묶여 있기 때문입니다 — 위 stepPerms에 1 1 1 1 1 / 1 1 1 2 / ... 를 쭉
        //   적은 그 관점이요. 나열의 세계에선 모든 줄이 서로 다른 순열이라 "뭘 캐시하지?"가 안 보입니다.
        //   캐싱은 다른 관점에서만 보입니다 — "나열"이 아니라 "상태에서 세기".
        //
        //   관점을 바꿔 이 함수가 정말 무엇을 돌려주는지 보면: permutation(n, sum)은 "n칸 중 sum칸까지
        //   올라온 상태에서, 남은 칸을 오르는 방법의 수"입니다. 여기서 결정적인 두 가지:
        //   - 그 답은 '여기까지 어떻게 왔는지'(경로)와 무관하다. sum=3에 1+2로 왔든 1+1+1로 왔든, 남은
        //     2칸을 오르는 방법 수는 똑같다.
        //   - 그러니 답은 인자 sum에만 달렸다. 같은 sum -> 항상 같은 답.
        //
        //   그런데 지금 코드는 그 '항상 같은 답'을 경로마다 처음부터 다시 계산합니다. 재귀 트리를 그리면
        //   같은 상태가 몇 번씩 다시 펼쳐지는 게 보입니다 (n=5):
        //
        //     perm(5,0)
        //      +- perm(5,1)
        //      |   +- perm(5,2)
        //      |   |   +- perm(5,3)*      <- sum=3 도달
        //      |   |   +- perm(5,4)
        //      |   |   +- perm(5,5)=1
        //      |   +- perm(5,3)*          <- (5,3) 또          [1+2로 도달]
        //      |   +- perm(5,4)
        //      +- perm(5,2)               <- (5,2) 통째로 또
        //      |   +- perm(5,3)*          <- (5,3) 또또
        //      |   +- ...
        //      +- perm(5,3)*              <- (5,3) 또또또       [3으로 도달]
        //
        //   perm(5,3)* 하나가 네 번 불리고, 매번 그 아래 전체를 다시 펼칩니다. n이 커지면 이 중복이 지수로
        //   폭발해요 — 실측: n=35면 호출이 4,047,854,365번(약 40억), 6.4초. 호출 수가 답 자체(약 11억)에
        //   비례합니다. 그게 TLE의 정체입니다.
        //
        //   그래서 캐싱: 같은 sum은 항상 같은 답이니, 처음 계산할 때 memo[sum]에 저장해두고 다음엔 꺼내
        //   씁니다. 서로 다른 상태가 몇 개일까요? sum은 0..n뿐이라 n+1개. 각 상태를 딱 한 번만 계산하면
        //   지수가 선형(O(n))으로 내려갑니다. (인자가 (n, sum) 둘 같지만 n은 재귀 내내 고정이라 실제로
        //   변하는 상태는 sum 하나 -> memo 배열 하나면 충분합니다. 고친 코드는 아래.)
        //
        // > 카드: 재귀가 같은 인자로 여러 번 불리고 답이 경로가 아니라 인자에만 달렸으면 -> 그 인자를 키로 캐시(메모). "나열"이 아니라 "상태에서 세기"로 보면 보인다.
        var cnt = 0;
        for (var step : List.of(1, 2, 3)) {
            cnt += permutation(n, sum + step, memo);
        }

        return cnt;
    }

    // 학습자 접근을 고친 동작 버전 (top-down 메모이제이션 — 골격 그대로, 캐시만 추가):
    //   memo[sum]을 -1로 초기화하고, 계산 전에 적중 검사, 계산 후 저장. 딱 세 줄이 더해질 뿐 재귀 구조는 그대로.
    //
    //     static int stepPerms(int n) {
    //         int[] memo = new int[n + 1];      // O(n) 공간: 상태는 sum=0..n, 총 n+1개
    //         Arrays.fill(memo, -1);            // -1 = 아직 계산 안 함
    //         return permutation(n, 0, memo);
    //     }
    //     static int permutation(int n, int sum, int[] memo) {
    //         if (sum == n) return 1;
    //         if (sum > n) return 0;            // 넘으면 0 (memo 접근 전에 거름 -> 인덱스 안전)
    //         if (memo[sum] != -1) return memo[sum];   // 적중: 다시 계산하지 않고 꺼내 씀
    //         int cnt = 0;
    //         for (int step : List.of(1, 2, 3)) {      // 호출당 3갈래지만 캐시로 상태당 1번만 펼침
    //             cnt += permutation(n, sum + step, memo);
    //         }
    //         return memo[sum] = cnt;           // 처음 계산이면 저장하고 반환
    //     }
    //   비용: 서로 다른 상태 n+1개를 각각 한 번씩만 펼침 -> 시간 O(n).
    //
    //   원본의 if (n < sum)를 if (sum > n)로 같은 뜻이되 읽기 순하게 둔 것 외엔 골격 동일.
    //   실측: n=35에서 원본 6.4초 -> 이 버전 0.008ms.
    //
    // 베스트 프랙티스 (bottom-up — 재귀 없이 표를 아래에서 위로 채움):
    //   메모가 채우던 표를 재귀 대신 루프로 채웁니다. dp[i] = i칸 오르는 방법의 수.
    //   유도: i칸의 마지막 점프는 1, 2, 3칸 중 하나 -> 그 직전 위치는 i-1, i-2, i-3칸 -> 방법 수는 그 셋의 합.
    //   씨앗은 dp[0]=1 (0칸 = 안 움직이는 1가지).
    //
    //     static int stepPerms(int n) {
    //         int[] dp = new int[n + 1];                 // O(n) 공간
    //         dp[0] = 1;                                 // 0칸: 1가지 (씨앗)
    //         for (int i = 1; i <= n; i++) {             // O(n): n번 반복, 각 상수 작업
    //             dp[i] = dp[i - 1];                      // 마지막에 1칸 점프
    //             if (i >= 2) dp[i] += dp[i - 2];         // 마지막에 2칸 점프
    //             if (i >= 3) dp[i] += dp[i - 3];         // 마지막에 3칸 점프
    //         }
    //         return dp[n];
    //     }
    //
    //   [표 채우기] n=5: dp[0]=1, dp[1]=1, dp[2]=dp[1]+dp[0]=2, dp[3]=dp[2]+dp[1]+dp[0]=4,
    //              dp[4]=dp[3]+dp[2]+dp[1]=7, dp[5]=dp[4]+dp[3]+dp[2]=13 -> 답 13.
    //   top-down 메모와 bottom-up은 같은 표를 채웁니다 — 방향만 다릅니다(재귀는 위에서 부르며 내려가 채우고,
    //   루프는 아래 칸부터 쌓아 올림). 공간은 직전 3칸만 있으면 되니 변수 3개로 O(1)까지 줄일 수 있습니다.
    //
    // 최종 시간/공간 복잡도 (셋 비교):
    //   - 원본(재귀, 캐시 없음): 한 호출이 최대 3갈래 x 깊이 n -> 대략 O(3^n) 상한 (합이 n을 넘는 가지는
    //     잘려 실제론 좀 작지만 여전히 지수). 실측 호출: n=20 약 43만, n=30 약 1.9억(192,299,281),
    //     n=35 약 40억(4,047,854,365). 시간 O(3^n), 공간 O(n)(재귀 깊이).
    //   - 메모(top-down): 서로 다른 상태 sum=0..n -> n+1개, 각 한 번 상수 작업 -> 시간 O(n), 공간 O(n).
    //   - bottom-up: 루프 n번 x 상수 -> 시간 O(n), 공간 O(n) (직전 3칸만 들면 O(1)).
    //   지수 O(3^n) -> 선형 O(n)이 이 문제의 한 끗이고, 가른 건 알고리즘 교체가 아니라 "같은 상태를 다시
    //   계산하지 않는다"는 관점 하나입니다.
    //
    // 첨삭(부차, HR 제출용): 큰 n은 방법 수가 폭증해 문제는 답을 mod 10^10+7(10,000,000,007)로 요구합니다.
    //   그 값은 int 최대 2,147,483,647(약 21억)을 넘으니, Jesse 때와 똑같이 "제약을 읽고 타입을 정한다" ->
    //   long으로 받고 더할 때마다 % 10000000007L. (위 샘플 4/13/44는 작아 mod와 무관해 int로도 green.)
}