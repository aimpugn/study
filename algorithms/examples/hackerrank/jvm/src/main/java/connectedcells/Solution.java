package connectedcells;

import support.Judge;

import java.util.List;

/**
 * Connected Cells in a Grid — HackerRank (Interview Prep Kit, DFS/BFS). 격자 flood 순회 굳히기.
 * <a href="https://www.hackerrank.com/challenges/connected-cell-in-a-grid/problem">문제</a>
 * <p>
 * n x m 격자에서 값 1인 칸들이 상하좌우 + 대각선(8방향)으로 이어지면 한 지역(region)입니다.
 * 가장 큰 지역의 칸 개수를 반환합니다.
 * <p>
 * 힌트(막히면 Castle의 그 골격): 방문 안 한 1을 만나면 거기서 BFS(또는 DFS)로 이어진 1을 전부 훑으며
 * 개수를 세고, 방문 표시를 남깁니다. 최대 개수를 기록. Castle과 다른 점 둘 — (1) 최단 거리(dist)가
 * 아니라 "이어진 칸 수"를 세고, (2) 이웃이 미끄러짐이 아니라 인접 8칸(대각 포함). 골격은 같습니다:
 * queue(시작) -> for(8방향) -> 이웃이 1이고 미방문이면 방문표시 + 큐에 넣기.
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

    /*
     * Complete the 'connectedCell' function below.
     *
     * The function is expected to return an INTEGER.
     * The function accepts 2D_INTEGER_ARRAY matrix as parameter.
     *
     * Consider a matrix where each cell contains either a 0 or a 1.
     * - 1: filled cell
     * - Two cells are said to be connected: if they are adjacent to each other horizontally, vertically, or diagonally.
     * In the following grid, all cells marked X are connected to the cell marked Y.
     *
     * If one or more filled cells are also connected,
     * they form a "region".
     * Note that each cell in a "region" is "connected to zero or more cells in the region"
     * but is not necessarily directly connected to all the other cells in the region.
     *
     * The larger region at the top left contains  cells. The smaller one at the bottom right contains .
     * 1 1 0
     * 1 0 0
     * 0 0 1
     * @param matrix
     * - matrix[i] = ith row
     * - 0 < matrix.size < 10
     * - 0 < matrix[i].size < 10
     * @return the area of the largest region
     *
     */
    public static int connectedCell(List<List<Integer>> matrix) {
        // Write your code here
        // 1 1 0 0
        // 0 1 1 0
        // 0 0 1 0
        // 1 0 0 0
        //
        // matrix = [
        //   0 => [1, 1, 0, 0]
        //   1 => [0, 1, 1, 0]
        //   2 => [0, 0, 1, 0]
        //   3 => [1, 0, 0, 0]
        // ]
        //
        // 1. add start point
        // 2. pop
        // 3. move horizontally, vertically, or diagonally
        //    => check connected and visited
        //    => count +1 if filled
        //
        var n = matrix.size();
        var m = matrix.get(0).size();
        // multi dimentional array to check visited
        var visited = new boolean[n][m];

        var max = 0;
        // loop all possible matrix
        for (var x = 0; x < n; x++) {
            for (var y = 0; y < m; y++) {
                // visit, mark, and count
                max = Math.max(max, visit(matrix, visited, n, m, x, y));
            }
        }

        // System.out.println(Arrays.deepToString(visited));

        return max;
    }

    static int visit(List<List<Integer>> matrix, boolean[][] visited, int n, int m, int x, int y) {
        // NOTE: filter visited HERE, because of "java.lang.StackOverflowError"
        if (x < 0 || y < 0 || n <= x || m <= y || visited[x][y] || !isFilled(matrix, x, y)) return 0;

        visited[x][y] = true;
        // System.out.println(Arrays.deepToString(visited));

        // movements
        var movements = new int[][]{
            {-1, 0}, // up
            {-1, -1}, // up-left
            {-1, 1}, // up-right
            {1, 0}, // down
            {1, -1}, // down-left
            {1, 1}, // down-right
            {0, 1}, // right
            {0, -1}, // left
        };
        var cnt = 1;
        for (var movement : movements) {
            cnt += visit(matrix, visited, n, m, x + movement[0], y + movement[1]);
        }

        return cnt;
    }

    static boolean isFilled(List<List<Integer>> matrix, int x, int y) {
        return matrix.get(x).get(y) == 1;
    }

    // 학습자 접근을 고친 동작 버전 (바깥 nested for = 영역 씨앗 찾기 그대로, 안쪽 슬라이드만 큐 flood로):
    //   세워둔 queue를 실제로 쓴다. 씨앗을 큐에 넣고, 꺼낸 칸의 8이웃 중 채워짐+미방문을 큐에 넣으며 센다.
    //   while 슬라이드(newX += movement) 제거 — 이웃은 인접 한 칸이고, 그 이웃도 꺼내지면 자기 이웃을 본다.
    //
    //     public static int connectedCell(List<List<Integer>> matrix) {
    //         int n = matrix.size(), m = matrix.get(0).size();
    //         boolean[][] visited = new boolean[n][m];
    //         int[][] movements = { {-1,0},{-1,-1},{-1,1},{1,0},{1,-1},{1,1},{0,1},{0,-1} }; // 8방향
    //         int max = 0;
    //         for (int x = 0; x < n; x++) {                 // 바깥: 새 영역의 씨앗 찾기
    //             for (int y = 0; y < m; y++) {
    //                 if (visited[x][y] || !isFilled(matrix, x, y)) continue;
    //                 var queue = new ArrayDeque<Point>();  // 씨앗에서 flood 시작
    //                 queue.add(Point.of(x, y));
    //                 visited[x][y] = true;
    //                 int cnt = 0;
    //                 while (!queue.isEmpty()) {            // 안쪽: 한 영역을 전부 훑기
    //                     var c = queue.poll();
    //                     cnt++;                            // 꺼낼 때 +1
    //                     for (int[] d : movements) {
    //                         int nx = c.x + d[0], ny = c.y + d[1];   // 인접 한 칸 (슬라이드 아님)
    //                         if (nx >= 0 && nx < n && ny >= 0 && ny < m
    //                                 && !visited[nx][ny] && isFilled(matrix, nx, ny)) {
    //                             visited[nx][ny] = true;   // 큐에 넣기 전에 방문표시 (중복·무한 방지)
    //                             queue.add(Point.of(nx, ny));
    //                         }
    //                     }
    //                 }
    //                 max = Math.max(max, cnt);
    //             }
    //         }
    //         return max;
    //     }
    //
    //   바뀐 것: while 슬라이드 -> 큐 flood(찾은 칸을 큐에 넣고 꺼내 다시 이웃), visited를 큐 넣기 전에 검사,
    //   cnt는 큐에서 꺼낼 때 +1. 검증: 4x4=5, 2x2=4, 단일=1, 사용자 예제=6.
    //   [데이터 흐름 - BFS 방문 순서] 격자 [1,1,0,0 / 0,1,1,0 / 0,0,1,1], 씨앗 (0,0), 칸에 방문 순서를 매기면:
    //     1 3 . .      큐(FIFO)라 가까운 칸부터 — (0,0)1 (1,1)2 (0,1)3 (2,2)4 (1,2)5 (2,3)6
    //     . 2 5 .      파동이 씨앗에서 바깥으로 한 겹씩 퍼진다(BFS = 겹 우선).
    //     . . 4 6
    //
    // 베스트 프랙티스 (재귀 DFS flood — 완전한 코드):
    //   Stage 2의 큐(BFS)와 같은 flood를 재귀로 쓴 것. Number of Islands에서 쓴 dfs에 '크기 반환'과 '8방향'만
    //   더하면 됩니다. flood는 경계 밖/방문함/안 채워짐이면 0을 돌려주므로, 바깥에서 모든 칸에 그냥 호출해도
    //   씨앗이 아닌 칸은 0을 반환합니다(그래서 바깥 if 검사가 없어도 됨).
    //
    //     static final int[][] DIRS = { {-1,0},{-1,-1},{-1,1},{1,0},{1,-1},{1,1},{0,1},{0,-1} }; // 8방향
    //
    //     public static int connectedCell(List<List<Integer>> matrix) {
    //         int n = matrix.size(), m = matrix.get(0).size();
    //         boolean[][] visited = new boolean[n][m];
    //         int max = 0;
    //         for (int x = 0; x < n; x++) {
    //             for (int y = 0; y < m; y++) {
    //                 max = Math.max(max, flood(matrix, visited, x, y)); // 씨앗 아니면 0 반환
    //             }
    //         }
    //         return max;
    //     }
    //     static int flood(List<List<Integer>> matrix, boolean[][] visited, int x, int y) {
    //         int n = matrix.size(), m = matrix.get(0).size();
    //         if (x < 0 || x >= n || y < 0 || y >= m) return 0;         // 경계 밖
    //         if (visited[x][y] || !isFilled(matrix, x, y)) return 0;   // 방문함 또는 안 채워짐
    //         visited[x][y] = true;
    //         int size = 1;                                             // 이 칸 하나
    //         for (int[] d : DIRS) {
    //             size += flood(matrix, visited, x + d[0], y + d[1]);   // 각 이웃에서 다시 flood (재출발)
    //         }
    //         return size;
    //     }
    //     // isFilled는 위 활성 코드의 것을 그대로 사용.
    //
    //   검증: 4x4=5, 2x2=4, 단일=1, 사용자 예제=6, 전부 0=0.
    //   [데이터 흐름 - DFS 방문 순서] 같은 격자·씨앗 — 같은 6칸이지만 순서가 다르다(깊이 우선):
    //     1 3 . .      한 방향으로 파고든 뒤 되돌아옴 — (0,0)1 (1,1)2 (0,1)3 (1,2)4 (2,2)5 (2,3)6
    //     . 2 4 .      BFS는 겹으로 퍼져 4가 (2,2)였지만, DFS는 한 줄기로 깊이 내려가 4가 (1,2).
    //     . . 5 6
    //   씨앗 깨달음: "연결"은 전이적이라 찾은 각 칸이 다시 출발점이 된다 -> 큐(BFS)든 재귀(DFS)든 "찾은 칸에서
    //   다시 이웃 보기(재출발)"가 핵심. 직선 슬라이드는 그 재출발이 없어 꺾인 연결을 놓친다. BFS(Stage 2)와
    //   DFS는 방문 순서만 다르고 세는 영역·크기는 같다(어느 쪽이든 완전한 코드 한 벌).
    //
    // 최종 시간/공간 복잡도:
    //   - 각 칸은 visited 덕에 큐에 한 번만 들어가고 한 번만 꺼내짐, 꺼낼 때 8이웃 확인 = 칸당 O(8) = O(1)
    //   - 칸 n*m개 -> 시간 O(n*m) (바깥 nested for도 각 칸 1회라 합쳐도 O(n*m))
    //   - 공간 O(n*m): visited 배열 + 최악에 한 영역이 격자 전체라 큐도 O(n*m). 제약 n,m<10이라 아주 작음.
}

/* 오답. 잘못된 풀이법
        // nested loop
        for (var x = 0; x < n; x++) {
            for (var y = 0; y < m; y++) {
                if (visited[x][y]) {
                    continue;
                }
                // mark visited
                visited[x][y] = true;

                // continue if not filled
                if (!isFilled(matrix, x, y)) {
                    continue;
                }

                // start when filled
                var cnt = 1;

                // 첨삭: 큰 틀(위 nested for로 '새 영역의 씨앗'을 찾는 것)은 맞습니다. 그런데 안쪽 확장이
                //   flood fill이 아니라 Castle의 "직선 미끄러짐"입니다 — 아래 for(movement) 안의 while이
                //   newX += movement로 한 방향으로 쭉 긋습니다. 그래서 꺾어야 닿는 칸을 못 잡아 영역이
                //   쪼개지고 개수가 틀립니다(실측: 기대 5, 이 코드 4).
                //
                //   위에 세운 queue(ArrayDeque<Point>)가 정확한 도구였어요 — flood는 그 큐로 합니다. 그런데
                //   큐를 안 쓰고 슬라이드로 새서 길을 잃었습니다.
                //
                //   왜 직선으로는 안 되나 (연결의 정의): "연결"은 이웃의 이웃의 이웃... 전이적 연쇄입니다.
                //   시작점에서 직선을 긋는 게 아니라, 찾은 칸을 큐에 넣고 그 칸을 꺼내 *다시* 이웃을 봅니다 —
                //   그래야 방향을 꺾습니다. 이웃은 인접 한 칸(슬라이드 아님).
                //
                //   [flood 연쇄] 격자 [1,1,0,0 / 0,1,1,0 / 0,0,1,1], 씨앗 (0,0):
                //     큐 [(0,0)]
                //     (0,0) 꺼냄 -> 채워짐+미방문 이웃: (0,1)우, (1,1)우하        큐 [(0,1),(1,1)]
                //     (0,1) 꺼냄 -> (1,2)우하                                     큐 [(1,1),(1,2)]  <- (0,0)에선 못 간 (1,2)를 (0,1)에서 꺾어 도달!
                //     (1,1) 꺼냄 -> (이미 다 방문)                                큐 [(1,2)]
                //     (1,2) 꺼냄 -> (2,2)하, (2,3)우하                            큐 [(2,2),(2,3)]
                //     (2,2),(2,3) 꺼냄 -> 없음                                    => 6칸 (정답)
                //
                //   격자에 방문 순서를 매기면 flood가 퍼지는 모양이 한눈에 (. = 아직 안 센 칸):
                //     1 3 . .      (0,0)1 (1,1)2 (0,1)3 (2,2)4 (1,2)5 (2,3)6  (큐 FIFO라 가까운 칸부터)
                //     . 2 5 .      (1,2)=5는 (0,0)에서 직접이 아니라 (1,1) 등을 거친 연쇄로 닿은 게 번호로 보인다.
                //     . . 4 6
                //
                //   반면 이 코드의 (0,0) 직선 슬라이드: 우 (0,1) | 우하 (1,1),(2,2) | 나머지 막힘  => 4칸.
                //   (1,2),(2,3)은 (0,0)에서 직선으로 못 닿아, 나중에 바깥 for가 (1,2)를 '새 영역'으로 잘못 셉니다.
                //
                //   지뢰 하나 더: 아래 while이 isFilled만 보고 visited는 안 봐서, 다른 씨앗에서 같은 칸을 다시
                //   셀 수 있습니다(중복 카운트). flood에선 큐에 넣기 전에 방문 검사가 필수.
                //
                //   전이: 이건 Number of Islands(격자 flood, DFS로 이미 푼 것)와 같은 뼈대예요 — 거기에 '크기
                //   세기'와 '8방향(대각 포함)'만 더한 것. 섬 세던 그 flood를 그대로 쓰면 됩니다.
                //
                // > 카드: 연결 덩어리 크기 = flood fill. 시작점에서 직선 긋지 말고, 찾은 칸마다 큐에 넣어 거기서 다시 이웃을 본다(연쇄, 꺾임 가능). 이웃은 인접 한 칸. (이동이 직선이면 Castle 슬라이드, 연결이 연쇄면 큐 재확장 — 구별!)

                for (var movement : movements) {
                    var newX = x + movement[0];
                    var newY = y + movement[1];
                    // move when filled
                    while (
                        (0 <= newX && newX < n)
                            && (0 <= newY && newY < m)
                            && isFilled(matrix, newX, newY)
                    ) {
                        // connected
                        cnt++;
                        System.out.println("x: " + x + ", y: " + y + ", newX: " + newX + ", newY: " + newY);

                        visited[newX][newY] = true;
                        newX += movement[0];
                        newY += movement[1];
                    }
                }
                max = Math.max(max, cnt);
            }
        }
 */