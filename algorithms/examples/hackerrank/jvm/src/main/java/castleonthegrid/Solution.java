package castleonthegrid;

import java.util.List;

import support.Judge;

/**
 * Castle on the Grid — HackerRank (Interview Prep Kit, Search). 격자 BFS 최단 이동.
 * <a href="https://www.hackerrank.com/challenges/castle-on-the-grid/problem">문제</a>
 * <p>
 * n x n 격자에서 '.'은 통과, 'X'는 막힘. 말(rook처럼)은 한 번에 상/하/좌/우 한 방향으로 막힐 때까지
 * 쭉 미끄러집니다(그게 1회 이동). 시작 (startX, startY)에서 목표 (goalX, goalY)까지 최소 이동 횟수를 구합니다.
 * (좌표는 행, 열 순서. HR 변수명이 X=행, Y=열이라 헷갈리니 주의.)
 * <p>
 * 핵심(막히면 이걸 떠올리세요): 최단 "횟수"는 BFS입니다 — DFS는 최단을 보장 못 합니다(아무 경로나 먼저 닿음).
 *   그리고 이 문제의 한 끗은 "이웃"의 정의입니다. 보통 BFS 이웃은 인접 4칸이지만, 여기선 한 번의 이동이
 *   "막힐 때까지 미끄러지기"라 한 칸의 이웃 = 네 방향으로 쭉 그은 직선 위의 모든 칸입니다.
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.minimumMoves(List.of(".X.", ".X.", "..."), 0, 0, 0, 2), 3);
        Judge.check(Result.minimumMoves(List.of("..."), 0, 0, 0, 0), 0);
        // 반례를 여기에 추가하세요:
    }
}

class Result {
    public static int minimumMoves(List<String> grid, int startX, int startY, int goalX, int goalY) {
        // Write your code here
        return -1;
    }
}
