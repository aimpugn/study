package castleonthegrid;

import support.Judge;

import java.util.ArrayDeque;
import java.util.Arrays;
import java.util.List;

/**
 * Castle on the Grid — HackerRank (Interview Prep Kit, Search). 격자 BFS 최단 이동.
 * <a href="https://www.hackerrank.com/challenges/castle-on-the-grid/problem">문제</a>
 * <p>
 * n x n 격자에서 '.'은 통과, 'X'는 막힘. 말(rook처럼)은 한 방향으로 막힐 때까지 미끄러지며, 그 직선 위
 * 아무 칸에나 설 수 있습니다(그게 1회 이동). 시작 (startX, startY)에서 목표 (goalX, goalY)까지 최소 이동
 * 횟수를 구합니다. (좌표는 행, 열 순서. HR 변수명이 X=행, Y=열이라 헷갈리니 주의.)
 * <p>
 * 제약: 1 <= grid.size <= 100, 각 좌표는 0 이상 n 미만.
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

    // BFS 최단 이동: 각 칸의 "처음 닿은 순간의 이동 횟수"를 dist에 파동처럼 채운다. 목표를 큐에서 꺼내는
    //   순간이 곧 최단(파동이 거리 순으로 퍼지니 먼저 닿은 게 가장 가깝다). 이 문제의 한 끗 = 한 이동이
    //   미끄러짐이라 이웃 = 네 방향 직선 위 모든 칸(그래서 for 안에 while로 nx += d 하며 march).
    //   유도(빈 화면에서 뽑아내는 사고):
    //   - "최단 횟수 + 가중치 없음" -> BFS (큐 + dist 배열)
    //   - "먼저 닿은 게 최단" -> dist 처음 쓸 때만 기록하고 다시 안 건드림 (미방문 검사 = 무한루프도 차단)
    //   - "한 이동 = 미끄러지기" -> 이웃이 인접 칸이 아니라 직선 전체 (4방향 while, nx += d)
    //   - "목표에 처음 닿으면 끝" -> 큐에서 목표를 꺼낼 때 dist 반환 (경로도 min도 불필요)
    //   복잡도: 칸 n^2개(각 칸 dist 덕에 큐에 한 번) x 칸당 4방향 미끄러짐 O(n) = O(n^3) 상한
    //   (n<=100이라 약 10^6(1,000,000), 넉넉히 통과). 공간 O(n^2).
    public static int minimumMoves(List<String> grid, int startX, int startY, int goalX, int goalY) {
        var n = grid.size();
        var dist = new int[n][n];
        for (var row : dist) {
            Arrays.fill(row, -1); // -1 = 미방문
        }

        var queue = new ArrayDeque<Pair>();
        queue.offer(Pair.of(startX, startY));
        dist[startX][startY] = 0;

        int[][] directions = {
            {-1, 0}, // up
            {1, 0}, // down
            {0, -1}, // left
            {0, 1} // right
        };
        while (!queue.isEmpty()) {
            var curr = queue.pop();
            if (curr.x == goalX && curr.y == goalY) {
                return dist[curr.x][curr.y]; // 처음 꺼내지는 순간이 최단
            }

            for (var direction : directions) {
                var newX = curr.x + direction[0];
                var newY = curr.y + direction[1];

                // 그 방향으로 막힐 때까지 미끄러지며, 지나치는 미방문 칸을 전부 dist+1로 적는다
                while (
                    (0 <= newX && newX < n)
                        && (0 <= newY && newY < n)
                        && !isBlocked(grid, newX, newY)
                ) {
                    if (dist[newX][newY] == -1) {
                        dist[newX][newY] = dist[curr.x][curr.y] + 1;
                        queue.offer(Pair.of(newX, newY));
                    }
                    newX += direction[0];
                    newY += direction[1];
                }
            }
        }

        return dist[goalX][goalY]; // 시작==목표면 0, 도달 불가면 -1
    }

    // 학습용 대안 — 경로 나열(DFS 백트래킹). 정답은 나오지만 지수 시간이라 n=100이면 TLE(제출 X).
    //   백트래킹은 "모든 경로"를 훑어 지수, BFS는 "칸별 최단"만 한 번씩 채워 O(n^3). 같은 답, 다른 비용.
    //   '모든 경로 나열'의 도구는 BFS 큐가 아니라 이 재귀 + mark/unmark다 — 되돌리기가 재귀 스택에 사는 게 DFS.
    static int minimumMovesByBacktrack(List<String> grid, int startX, int startY, int goalX, int goalY) {
        int n = grid.size();
        boolean[][] onPath = new boolean[n][n];    // 현재 경로에 밟은 칸 (사이클/무한루프 방지)
        onPath[startX][startY] = true;
        int[] best = {Integer.MAX_VALUE};          // 배열에 담아 재귀 사이로 공유
        backtrack(grid, n, startX, startY, goalX, goalY, 0, onPath, best);
        return best[0] == Integer.MAX_VALUE ? -1 : best[0];
    }

    static void backtrack(List<String> grid, int n, int x, int y, int gx, int gy, int moves, boolean[][] onPath, int[] best) {
        if (x == gx && y == gy) { // 목표 도달: 이 경로의 이동 수로 best 갱신
            best[0] = Math.min(best[0], moves);
            return;
        }
        if (moves + 1 >= best[0]) {
            return; // 가지치기: 더 가도 best 못 이기면 그만
        }

        int[][] dirs = new int[][]{
            {1, 0},
            {-1, 0},
            {0, 1},
            {0, -1}
        };
        for (int[] d : dirs) {
            int nx = x + d[0];
            int ny = y + d[1];

            while (nx >= 0 && nx < n && ny >= 0 && ny < n && grid.get(nx).charAt(ny) != 'X') { // 룩: 직선 위 아무 칸
                if (!onPath[nx][ny]) {
                    onPath[nx][ny] = true;         // 밟기
                    backtrack(grid, n, nx, ny, gx, gy, moves + 1, onPath, best);
                    onPath[nx][ny] = false;        // 되돌리기 (이 줄이 백트래킹의 심장)
                }
                nx += d[0];                        // 같은 방향 다음 칸도 착지 후보 (룩은 아무 데나 멈춤)
                ny += d[1];
            }
        }
    }

    static boolean isBlocked(List<String> grid, int x, int y) {
        return grid.get(x).charAt(y) == 'X';
    }

    // ============================================================================
    // 오답 노트 — 첫 시도(minimumMoves_failed)가 어디서 어긋났나
    // ============================================================================
    // 무엇을 생각했나: "목표까지 갈 수 있는 경로를 전부 모아서, 그중 최소 길이를 고르자"(containers에 경로
    //   쌓기 + dfs로 경로 수집 + Math.min(size)). 예전에 백트래킹으로 경우의 수를 조합하던 관성이 나온 것.
    //
    // 어디서 어긋났나 (사고의 갈림길):
    //   - "최단"을 "모든 경로 중 최소"로 읽었다. BFS는 경로를 모으지 않아도 최단을 준다는 걸 놓쳤다 —
    //     파동이 거리 순으로 퍼지니 목표에 '처음 닿은 순간'이 곧 최단이다(경로 수집 자체가 불필요).
    //   - BFS 큐를 세워 놓고 그 위에 '경로 수집(containers/dfs)'을 얹었다 -> 두 모델이 섞여 길을 잃음.
    //     그 결과 visited/dist가 없어 같은 칸을 무한히 다시 큐에 넣어 안 끝났다(무한 루프).
    //   - 데이터가 어떻게 흐르는지가 머릿속에 없었다: "큐에서 한 칸 꺼내 -> 네 방향으로 -> 그 방향으로 쭉
    //     (nx += d) 미끄러지며 미방문 칸에 dist 적기". 이 3중 구조(큐 / for 방향 / while 진행)의 그림이 없어 헤맸다.
    //
    // 무엇을 어떻게 생각했어야 하나 (다음에 이 신호를 잡아라):
    //   - "최단 횟수/거리 + 가중치 없음"을 보면 -> 즉시 BFS. "경로를 모을까?"가 아니라 "dist를 파동으로
    //     채우자"로 튼다. 경로 자체는 오직 경로를 출력해야 할 때만 든다.
    //   - 격자 BFS의 몸통은 늘 같다: queue(시작) -> while(큐 빌 때까지) -> for(방향) -> 이웃 계산 ->
    //     미방문이면 dist 기록 + 큐에 넣기. (이 문제는 '이웃'이 미끄러진 직선이라 for 안에 while이 하나 더.)
    //
    // > 오답 카드: BFS 큐 위에 '경로 수집'을 얹지 마라. 최단은 경로를 모으는 게 아니라 dist를 파동으로 채우는 것.
    //
    // 아래는 실패했던 첫 시도 원본(통째로 주석 보존 — "무엇을 안 해야 하는지"의 기록):
    /*
    public static int minimumMoves_failed(List<String> grid, int startX, int startY, int goalX, int goalY) {
        // to get to goal, need to try possible paths  <- 여기가 갈림길: '모든 경로'로 튼 지점
        var queue = new ArrayDeque<Pair>();
        queue.offer(Pair.of(startX, startY));
        var n = grid.size();
        var cnt = 0;
        var containers = new ArrayList<List<Pair>>();   // 경로 수집 발상의 잔재 (BFS엔 불필요)
        while (!queue.isEmpty()) {
            var curr = queue.pop();
            // 4방향을 '한 칸'씩만 봄(미끄러짐 아님) + visited/dist 없음 -> 무한 루프
            // move down
            if (curr.x + 1 < n && !isBlocked(grid, curr.x + 1, curr.y)) { queue.offer(Pair.of(curr.x + 1, curr.y)); }
            // move up   (0 < curr.x - 1 는 맨 윗줄을 빠뜨리는 경계 버그)
            if (0 < curr.x - 1 && !isBlocked(grid, curr.x - 1, curr.y)) { queue.offer(Pair.of(curr.x - 1, curr.y)); }
            // move right (Pair.of(curr.x + 1, ...) 오타 — 오른쪽인데 행을 늘림)
            if (curr.y + 1 < n && !isBlocked(grid, curr.x, curr.y + 1)) { queue.offer(Pair.of(curr.x + 1, curr.y)); }
            // move left
            if (0 < curr.y - 1 && !isBlocked(grid, curr.x, curr.y - 1)) { queue.offer(Pair.of(curr.x, curr.y - 1)); }
            // as a result, need find path and return list possible paths to goal
            // if (result != null) cnt = Math.min(cnt, result.size());  <- 경로 나열 + min 발상
        }
        return 0;  // 자리표시자 (애초에 무한 루프라 여기 못 닿음)
    }
    */
}

class Pair {
    int x;
    int y;

    Pair(int x, int y) {
        this.x = x;
        this.y = y;
    }

    static Pair of(int x, int y) {
        return new Pair(x, y);
    }
}
