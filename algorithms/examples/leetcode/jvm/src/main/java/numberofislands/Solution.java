package numberofislands;

import support.Judge;

/**
 * <a href="https://leetcode.com/problems/number-of-islands/">Number of Islands</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     * - "섬"이란? 물로 둘러싸여 있고 수평/수직적으로 인근 땅과 연결되어 있습니다.
     *
     * @param grid m x n 2차 바이너리 그리드
     * - "1"이면 땅, "0"이면 물을 나타냅니다.
     * @return 섬의 개수
     */
    public int numIslands(char[][] grid) {
        // 1. grid = [
        //  ["1","1","1","1","0"],
        //  ["1","1","0","1","0"],
        //  ["1","1","0","0","0"],
        //  ["0","0","0","0","0"]
        // ]
        // => 섬은 한 개입니다.
        //
        // 2. grid = [
        //  ["1","1","0","0","0"],
        //  ["1","1","0","0","0"],
        //  ["0","0","1","0","0"],
        //  ["0","0","0","1","1"]
        // ]
        // => 섬은 세 개입니다.
        //
        // 연속으로 계속 도달 가능하면 한 개의 섬으로 판단합니다.
        // 그러다가 주변이 모두 물이 되면 끝냅니다.
        // bfs로 점차 탐색하고, visited로 방문 여부를 판단합니다.
        // 땅이면 주변을 탐색하고, 땅이 아니면 방문만 체크합니다.
        if (grid == null || grid.length == 0 || grid[0].length == 0) return 0;

        var visited = new boolean[grid.length][grid[0].length];
        // 순회 방법은?
        // (0,0) -> (0, 1), (1, 0) -> (1, 1), (2, 0), (0, 2)
        // BFS 말고 중첩 루프로 체크하는 게 더 직관적일 것으로 보입니다.
        // 물이면 방문 표시하고 스킵하다가, 땅이면 점차 확장하여 파고듭니다.
        // 상/하/좌/우로 탐색하다가 모두 끝 또는 물을 만나면 종료하고 다음 방문하지 않은 인덱스로 이동합니다.
        // 1. 땅이 시작되는 지점을 찾아야 하고
        // 2. 땅인 부분부터 영역을 넓혀 가야 합니다.
        // 첨삭: 정답입니다(네 케이스 모두 통과). 강화할 지점은 도구를 고른 흐름입니다 — 처음 "bfs로
        //   점차 탐색"을 떠올렸다가 "BFS 말고 중첩 루프로 체크"로 역할을 나눈 판단이 이 문제의 골격
        //   그 자체입니다. 격자에서 연결된 덩어리를 세는 문제는 거의 항상 이 두 층입니다 — 바깥은
        //   훑고, 안은 한 땅에서 파고듭니다.
        //
        //   왜 이 구조가 섬을 정확히 세는지는 바깥 루프의 불변식 한 문장으로 닫힙니다: 바깥 루프가
        //   (i, j)까지 훑은 직후, 이미 발견된 모든 섬의 땅 칸은 visited입니다. 이 문장이 참이면 바깥
        //   루프가 "아직 visited 아닌 땅"을 처음 만나는 순간은 곧 "새 섬의 첫 칸"이고, 거기서 한 번만
        //   세면 됩니다. 이어지는 dfs가 그 섬 전체를 visited로 물들이므로 같은 섬의 다른 칸에서는 다시
        //   세지 않습니다 — 불변식이 그 다음 칸에서도 그대로 유지됩니다.
        //
        // > 카드: 격자에서 연결된 덩어리 개수 = 바깥 루프로 미방문 시작점을 찾고, 각 시작점에서 전체를 물들여(flood) 다시 안 세지게. 세는 건 바깥 루프, 물들이는 건 dfs.
        var rowLen = grid.length;
        var colLen = grid[0].length;
        var cnt = 0;

        for (var i = 0; i < rowLen; i++) {
            for (var j = 0; j < colLen; j++) {
                if (visited[i][j]) {
                    continue;
                }
                // 땅인 경우
                if (grid[i][j] == '1') {
                    // 현재 기준으로 상/하/좌/우로 이동하면서 도달 가능한 곳들을 모두 방문 표시합니다.
                    // 계속 파고들어야 하므로 dfs를 사용합니다.
                    cnt += dfs(grid, rowLen, colLen, i, j, visited);
                    // 첨삭: dfs의 return 0/1로 세는 건 동작하지만 한 겹 돌아가는 길입니다. 바깥 루프가
                    //   이미 grid[i][j]=='1' && !visited를 확인하고 부르므로 이 top-level 호출은 항상 1을
                    //   돌려주고, 안쪽 재귀(상/하/좌/우)의 return은 전부 버려집니다. 즉 `cnt += dfs(...)`는
                    //   사실상 `cnt++; dfs(...)`입니다 — 아래 베스트 프랙티스 블록이 이 더 또렷한 형태를 씁니다.
                }
                // System.out.println("visited: " + Arrays.deepToString(visited));
            }
        }
        // System.out.println("cnt: " + cnt);

        return cnt;
    }

    int dfs(char[][] grid, int rowLen, int colLen, int row, int col, boolean[][] visited) {
        // 종료 조건
        if (row < 0 || row == rowLen || col < 0 || col == colLen) {
            return 0;
        }
        // 이미 방문한 적이 있다면 종료
        if (visited[row][col]) {
            return 0;
        }
        visited[row][col] = true;

        // 물이면 종료합니다.
        if (grid[row][col] == '0') {
            return 0;
        }

        dfs(grid, rowLen, colLen, row - 1, col, visited); // 상
        dfs(grid, rowLen, colLen, row + 1, col, visited); // 하
        dfs(grid, rowLen, colLen, row, col - 1, visited); // 좌
        dfs(grid, rowLen, colLen, row, col + 1, visited); // 우

        return 1;
    }

    // 첨삭 — 베스트 프랙티스(제자리 가라앉히기, in-place sink):
    //
    //   visited 배열의 목적은 "이미 본 땅을 다시 보지 않기" 하나입니다. 그런데 본 땅을 그 자리에서 물로
    //   바꿔 버리면 grid 자체가 그 기록이 되어 별도 배열이 필요 없습니다. 이 한 깨달음에서 코드가 따라
    //   나옵니다 — 방문 표시는 grid[r][c]='0'이 되고, "이미 봤거나 물" 판정은 grid[r][c] != '1' 한 줄로
    //   합쳐집니다. 원래 dfs의 경계 체크, visited 체크, 물 체크 세 갈래가 두 줄로 줄어듭니다.
    //
    //     int numIslands(char[][] grid) {
    //         if (grid == null || grid.length == 0) return 0;
    //         int cnt = 0;
    //         for (int r = 0; r < grid.length; r++) {
    //             for (int c = 0; c < grid[0].length; c++) {
    //                 if (grid[r][c] == '1') {    // 아직 안 가라앉은 땅 = 새 섬의 첫 칸
    //                     cnt++;                  // 세는 건 바깥 루프
    //                     sink(grid, r, c);       // 그 섬 전체를 가라앉혀 다시 안 세지게
    //                 }
    //             }
    //         }
    //         return cnt;
    //     }
    //
    //     void sink(char[][] g, int r, int c) {   // 물들이는 건 sink, 반환값 없음
    //         if (r < 0 || r >= g.length || c < 0 || c >= g[0].length) return; // 격자 밖이면 멈춤
    //         if (g[r][c] != '1') return;         // 물('0')이거나 이미 가라앉힌 칸이면 멈춤
    //         g[r][c] = '0';                      // 이 땅을 가라앉혀 방문 표시(다시 안 봄)
    //         sink(g, r - 1, c);                  // 상
    //         sink(g, r + 1, c);                  // 하
    //         sink(g, r, c - 1);                  // 좌
    //         sink(g, r, c + 1);                  // 우
    //     }
    //
    //   공간이 visited 방식의 O(m*n)에서 O(1)(재귀 콜스택은 별개)로 줄어듭니다. 트레이드오프는 입력
    //   grid를 파괴한다는 것 — LeetCode는 이를 허용하지만, 원본을 보존해야 하는 호출자라면 visited
    //   배열 방식이 맞습니다.
    //
    //   정직한 한계 하나 더: 재귀 DFS는 호출 깊이가 곧 한 섬의 칸 수라, 300x300이 전부 땅이면 깊이가
    //   9만까지 가 StackOverflow 위험이 있습니다. 처음 떠올렸던 BFS(큐로 한 겹씩 펴는 반복)는 깊이를
    //   안 쌓아 큰 격자에서 오히려 안전합니다 — 둘 다 맞고 한계가 다릅니다(DFS는 코드가 짧고, BFS는
    //   깊이에 강함).
    //
    // > 카드: visited 배열 대신 방문한 땅을 그 자리에서 '0'으로 덮으면(섬 가라앉히기) 공간 O(1). 대가는 입력 파괴 + 재귀 깊이 한계, 후자는 BFS로 회피.
    // > 카드: 이 flood-fill 골격은 그대로 전이 — 547 Provinces, 695 Max Area of Island, 130 Surrounded Regions, 994 Rotting Oranges(여긴 다중 시작점 BFS).

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(s.numIslands(new char[][]{{'1', '1', '1', '1', '0'}, {'1', '1', '0', '1', '0'}, {'1', '1', '0', '0', '0'}, {'0', '0', '0', '0', '0'}}), 1);
        Judge.check(s.numIslands(new char[][]{{'1', '1', '0', '0', '0'}, {'1', '1', '0', '0', '0'}, {'0', '0', '1', '0', '0'}, {'0', '0', '0', '1', '1'}}), 3);
        Judge.check(s.numIslands(new char[][]{}), 0);
        Judge.check(s.numIslands(new char[][]{{'1'}}), 1);
        // 반례를 여기에 추가하세요:
    }
}
