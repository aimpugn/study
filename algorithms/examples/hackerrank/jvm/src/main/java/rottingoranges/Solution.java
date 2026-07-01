package rottingoranges;

import support.Judge;

/**
 * Rotting Oranges — LeetCode 994 (격자 BFS "파동", 다중 시작점). Castle에서 미끄러짐만 뺀 순수 BFS 골격 굳히기.
 * <a href="https://leetcode.com/problems/rotting-oranges/">문제</a>
 * <p>
 * 격자 값: 0 = 빈 칸, 1 = 신선한 오렌지, 2 = 썩은 오렌지. 매 분(minute)마다 썩은 오렌지가 상하좌우로
 * 인접한 신선한 오렌지를 썩게 합니다. 모든 신선한 오렌지가 썩을 때까지 걸리는 최소 분을 반환하고,
 * 영영 못 썩는 게 있으면 -1을 반환합니다.
 * <p>
 * 힌트(막히면 Castle의 그 골격 그대로): 이건 "파동이 몇 겹 퍼지나 = 최소 분"이라 BFS입니다. Castle과 딱
 *   두 가지만 다릅니다 — (1) 시작점이 여럿(처음부터 썩은 칸 전부를 큐에 넣고 시작), (2) 이웃이 미끄러짐이
 *   아니라 인접 4칸(그래서 while 없이 한 칸씩). dist 대신 "몇 번째 파동인지"를 세고, 남은 신선 개수가
 *   0이 되는 순간이 답. 처음부터 신선이 없으면 0.
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.orangesRotting(new int[][]{{2, 1, 1}, {1, 1, 0}, {0, 1, 1}}), 4);
        Judge.check(Result.orangesRotting(new int[][]{{2, 1, 1}, {0, 1, 1}, {1, 0, 1}}), -1);
        Judge.check(Result.orangesRotting(new int[][]{{0, 2}}), 0);
        // 반례를 여기에 추가하세요:
    }
}

class Result {
    public static int orangesRotting(int[][] grid) {
        // Write your code here
        return -1;
    }
}
