package connectedcells;

import java.util.List;

import support.Judge;

/**
 * Connected Cells in a Grid — HackerRank (Interview Prep Kit, DFS/BFS). 격자 flood 순회 굳히기.
 * <a href="https://www.hackerrank.com/challenges/connected-cell-in-a-grid/problem">문제</a>
 * <p>
 * n x m 격자에서 값 1인 칸들이 상하좌우 + 대각선(8방향)으로 이어지면 한 지역(region)입니다.
 * 가장 큰 지역의 칸 개수를 반환합니다.
 * <p>
 * 힌트(막히면 Castle의 그 골격): 방문 안 한 1을 만나면 거기서 BFS(또는 DFS)로 이어진 1을 전부 훑으며
 *   개수를 세고, 방문 표시를 남깁니다. 최대 개수를 기록. Castle과 다른 점 둘 — (1) 최단 거리(dist)가
 *   아니라 "이어진 칸 수"를 세고, (2) 이웃이 미끄러짐이 아니라 인접 8칸(대각 포함). 골격은 같습니다:
 *   queue(시작) -> for(8방향) -> 이웃이 1이고 미방문이면 방문표시 + 큐에 넣기.
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.connectedCell(List.of(
            List.of(1, 1, 0, 0),
            List.of(0, 1, 1, 0),
            List.of(0, 0, 1, 0),
            List.of(1, 0, 0, 0))), 5);
        Judge.check(Result.connectedCell(List.of(
            List.of(1, 1),
            List.of(1, 1))), 4);
        Judge.check(Result.connectedCell(List.of(
            List.of(1, 0),
            List.of(0, 0))), 1);
        // 반례를 여기에 추가하세요:
    }
}

class Result {
    public static int connectedCell(List<List<Integer>> matrix) {
        // Write your code here
        return 0;
    }
}
