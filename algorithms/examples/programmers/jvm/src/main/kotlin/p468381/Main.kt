package p468381

/**
 * [2025 카카오 하반기 2차 - 기차선로](https://school.programmers.co.kr/learn/courses/30/lessons/468381?language=kotlin)
 *
 * 이 파일은 정답만 남기는 제출용 코드라기보다, 어려운 상태 탐색 문제를 머릿속에 각인하기 위한 학습용 풀이입니다.
 * 따라서 주석은 문법 설명보다 "이 줄이 어떤 문제 조건을 코드 상태로 바꾸는가"에 초점을 둡니다.
 */
class Solution {
    /**
     * 이 문제는 "도착 가능한가"가 아니라 "선로를 놓는 방법이 몇 가지인가"를 묻습니다.
     * 그래서 BFS로 도착 여부만 보지 않고, DFS/백트래킹으로 실제 배치를 하나씩 만들어 보며 경우의 수를 셉니다.
     *
     * 핵심 사고 흐름은 다음과 같습니다.
     *
     * 1. 선로 번호를 열린 방향 집합으로 번역합니다.
     * 2. 현재 칸에서 한 방향으로 이동할 수 있는지 확인합니다.
     * 3. 다음 칸이 빈칸이면, 필요한 방향을 받아 줄 수 있는 선로만 후보로 놓아 봅니다.
     * 4. 후보를 놓고 DFS로 한 칸 들어간 뒤, 돌아오면 반드시 원복합니다.
     * 5. 도착점에 도착하면 모든 선로 방문 여부와 전체 연결 상태를 함께 검사합니다.
     *
     * `visited`와 `inPath`는 서로 다른 질문에 답합니다.
     * `visited`는 "이번 완성 후보에서 이 선로를 한 번 이상 지났는가"를 묻고,
     * `inPath`는 "현재 재귀 호출 스택 안에서 같은 이동 상태가 반복되고 있는가"를 묻습니다.
     */
    fun solution(grid: Array<IntArray>): Int {
        // 프로그래머스 기본 템플릿에서 남은 변수입니다.
        // 현재 풀이는 `dfs(...)`가 직접 경우의 수를 반환하므로 이 변수는 실제 계산에 쓰이지 않습니다.
        var answer: Int = 0

        // 선로 번호를 "어느 방향으로 문이 열려 있는가"로 바꾸는 표입니다.
        // 이 표를 먼저 만들면 뒤의 코드는 1번, 2번 같은 숫자 모양을 직접 외우지 않고 방향 집합만 보고 판단할 수 있습니다.
        val tracks = mapOf(
            // 1번 선로는 왼쪽과 오른쪽이 연결됩니다.
            1 to setOf(Direction.LEFT, Direction.RIGHT),
            // 2번 선로는 위와 아래가 연결됩니다.
            2 to setOf(Direction.UP, Direction.DOWN),
            // 3번 선로는 네 방향이 모두 열려 있습니다. 최종 검증에서는 네 방향 모두 실제로 이어져야 합니다.
            3 to setOf(Direction.UP, Direction.DOWN, Direction.LEFT, Direction.RIGHT),
            // 4번 선로는 왼쪽과 위가 연결됩니다.
            4 to setOf(Direction.LEFT, Direction.UP),
            // 5번 선로는 오른쪽과 위가 연결됩니다.
            5 to setOf(Direction.RIGHT, Direction.UP),
            // 6번 선로는 오른쪽과 아래가 연결됩니다.
            6 to setOf(Direction.RIGHT, Direction.DOWN),
            // 7번 선로는 왼쪽과 아래가 연결됩니다.
            7 to setOf(Direction.LEFT, Direction.DOWN),
        )

        // 이 배열은 "결과 조건"을 위한 방문 기록입니다.
        // 문제는 모든 선로를 한 번 이상 지나야 한다고 했으므로, 선로 칸을 처음 지날 때만 카운트를 올립니다.
        val visited = Array(grid.size) { BooleanArray(grid[0].size) }

        // 이 배열은 "무한 재귀 방지"를 위한 현재 경로 기록입니다.
        // 같은 칸이라도 어느 방향에서 들어왔는지에 따라 다음 선택이 달라질 수 있으므로 세 번째 차원을 방향 수로 둡니다.
        val inPath = Array(grid.size) {
            Array(grid[0].size) {
                BooleanArray(Direction.entries.size)
            }
        }

        // 처음부터 격자에 놓여 있던 선로 수입니다.
        // DFS 중 빈칸에 새 선로를 놓으면, 그 분기 안에서는 `totalTrackCount`가 추가로 늘어납니다.
        val initialTrackCount = grid.sumOf { row ->
            row.count { cell -> cell > 0 }
        }

        // 시작점 `(0, 0)`의 1번 선로는 기차가 이미 지나기 시작한 칸으로 봅니다.
        visited[0][0] = true

        // 기차는 시작 칸의 왼쪽 바깥에서 들어와 오른쪽으로 출발한다고 모델링합니다.
        // 그래서 `enteredFrom = Direction.LEFT`, `visitedCount = 1`로 DFS를 시작합니다.
        return dfs(
            grid,
            tracks,
            0,
            0,
            Direction.LEFT,
            visited,
            1,
            inPath,
            initialTrackCount,
        )
    }

    /**
     * 현재 칸 `(row, col)`에서 `direction`으로 한 칸 이동할 때,
     * 다음 칸에 놓을 수 있는 선로 번호 후보를 반환합니다.
     *
     * 이 함수가 닫는 질문은 하나입니다.
     * "현재 칸에서 나갈 수 있고, 다음 칸이 그 반대 방향을 받아 줄 수 있는가?"
     *
     * 빈 리스트를 반환하면 이 방향으로는 이동할 수 없다는 뜻입니다.
     */
    fun moveCandidates(
        grid: Array<IntArray>,
        tracks: Map<Int, Set<Direction>>,
        row: Int,
        col: Int,
        currentTrack: Int,
        direction: Direction
    ): List<Int> {
        // 이동하려는 다음 좌표입니다. 격자 이동 문제는 항상 "현재 좌표 + 방향 변화량"에서 시작합니다.
        val nextRow = row + direction.rowDelta
        val nextCol = col + direction.colDelta

        // 실제로 접근할 칸은 현재 칸이 아니라 다음 칸입니다.
        // 다음 칸이 격자 밖이면 이 이동은 성립하지 않으므로 후보가 없습니다.
        if (!(nextRow in grid.indices && nextCol in grid[nextRow].indices)) return emptyList()

        // 현재 선로가 나가려는 방향으로 열려 있어야 합니다.
        // `!= true`를 쓰면 선로 번호가 없거나, 방향이 없거나, 맵 조회가 실패한 경우를 모두 이동 불가로 처리합니다.
        if (tracks[currentTrack]?.contains(direction) != true) return emptyList()

        // 다음 칸은 현재 이동 방향의 반대 방향을 열고 있어야 합니다.
        // 예를 들어 현재 칸에서 RIGHT로 나가면, 다음 칸은 LEFT가 열려 있어야 합니다.
        return when (val cell = grid[nextRow][nextCol]) {
            // 장애물은 지나갈 수도 없고 선로를 놓을 수도 없습니다.
            -1 -> emptyList()

            // 빈칸에는 모든 선로를 놓아 보지 않습니다.
            // 방금 들어오는 방향, 즉 `direction.opposite()`을 열 수 있는 선로만 후보로 둡니다.
            0 -> tracks
                .filter { (_, openedDirections) -> direction.opposite() in openedDirections }
                .keys
                .sorted()

            // 이미 선로가 놓여 있다면 그 선로가 반대 방향을 열 때만 이동할 수 있습니다.
            // 가능한 경우에도 후보는 기존 선로 하나뿐입니다.
            else -> if (tracks[cell]?.contains(direction.opposite()) == true) listOf(cell) else emptyList()
        }
    }

    /**
     * 현재 칸에 들어온 방향을 기준으로, 현재 선로에서 나갈 수 있는 방향을 반환합니다.
     *
     * `moveCandidates`가 "다음 칸이 받아 줄 수 있는가"를 본다면,
     * 이 함수는 "현재 칸에서 어디로 나갈 것인가"를 봅니다.
     */
    fun outgoingDirections(
        tracks: Map<Int, Set<Direction>>,
        currentTrack: Int,
        enteredFrom: Direction?
    ): List<Direction> {
        // 현재 선로 번호가 여는 방향 집합입니다. 선로가 아니면 나갈 방향이 없습니다.
        val openedDirections = tracks[currentTrack] ?: return emptyList()

        // 들어온 방향이 없는 일반화된 시작 상태입니다.
        // 현재 풀이에서는 시작을 LEFT에서 들어온 것으로 모델링하지만, 함수는 null 시작도 표현할 수 있게 되어 있습니다.
        if (enteredFrom == null) {
            return openedDirections.toList()
        }

        // 들어온 방향이 현재 선로의 열린 방향에 없다면 애초에 이 칸에 들어올 수 없는 상태입니다.
        if (enteredFrom !in openedDirections) {
            return emptyList()
        }

        // 3번 선로는 네 방향이 모두 열려 있지만, 이 DFS 이동에서는 들어온 방향의 반대편으로 직진시킵니다.
        // 3번 선로의 네 방향 연결 의무는 마지막 `allConnectionsValid`에서 별도로 검사합니다.
        return if (currentTrack == 3) {
            listOf(enteredFrom.opposite())
        } else {
            // 일반 선로는 열린 두 방향 중 들어온 방향이 아닌 쪽으로 나갑니다.
            openedDirections.filter { direction -> direction != enteredFrom }
        }
    }

    /**
     * DFS는 이 풀이의 중심입니다.
     *
     * 한 호출은 "현재 칸에 들어온 상태에서, 도착점까지 만들 수 있는 유효한 선로 배치 수"를 반환합니다.
     * 백트래킹 리듬은 항상 다음 네 단계입니다.
     *
     * 1. 후보를 고릅니다.
     * 2. `grid`, `visited`, `inPath` 같은 상태를 바꿉니다.
     * 3. 재귀로 한 칸 더 들어갑니다.
     * 4. 재귀가 끝나면 바꾼 상태를 되돌립니다.
     */
    fun dfs(
        grid: Array<IntArray>,
        tracks: Map<Int, Set<Direction>>,
        row: Int,
        col: Int,
        enteredFrom: Direction?,
        visited: Array<BooleanArray>,
        visitedCount: Int,
        inPath: Array<Array<BooleanArray>>,
        totalTrackCount: Int
    ): Int {
        // 도착점에 왔다고 바로 정답이 아닙니다.
        // 모든 선로를 한 번 이상 지났고, 완성된 grid의 모든 열린 방향이 합법적으로 연결되어야 한 가지 배치입니다.
        if (row == grid.lastIndex && col == grid[0].lastIndex) {
            return if (visitedCount == totalTrackCount && allConnectionsValid(grid, tracks)) 1 else 0
        }

        // `enteredFrom?.ordinal`은 방향을 `inPath`의 세 번째 인덱스로 바꿉니다.
        // `enteredFrom`이 null이면 "어느 방향에서 들어왔는가"라는 질문 자체가 없으므로 기록할 인덱스도 없습니다.
        val enteredDirectionIndex = enteredFrom?.ordinal

        // `inPath`는 방문 금지가 아니라, 현재 재귀 경로 안에서 같은 상태가 반복되는 것을 막는 장치입니다.
        // 같은 `(row, col, enteredFrom)` 상태가 다시 나오면 이후 선택이 반복될 수 있으므로 이 가지는 0으로 끊습니다.
        if (enteredDirectionIndex != null) {
            if (inPath[row][col][enteredDirectionIndex]) {
                return 0
            }

            // 지금부터 이 재귀 호출이 끝날 때까지 이 상태는 현재 경로 안에서 사용 중입니다.
            inPath[row][col][enteredDirectionIndex] = true
        }

        // 이 상태에서 가능한 모든 하위 배치 수를 누적합니다.
        var count = 0

        // 현재 칸의 선로 번호입니다. 빈칸이었다면 이전 호출에서 이미 후보 선로가 놓인 상태여야 합니다.
        val currentTrack = grid[row][col]

        // 현재 선로에서 나갈 수 있는 방향을 하나씩 시도합니다.
        for (direction in outgoingDirections(tracks, currentTrack, enteredFrom)) {
            // 이 방향으로 한 칸 이동할 때 다음 칸에 놓을 수 있는 후보 선로들입니다.
            val candidates = moveCandidates(grid, tracks, row, col, currentTrack, direction)

            // 각 후보는 하나의 분기입니다. 후보를 실제로 놓고 내려갔다가 돌아오면 원복합니다.
            for (nextTrack in candidates) {
                // 다음 좌표는 이 방향으로 한 칸 이동한 위치입니다.
                val nextRow = row + direction.rowDelta
                val nextCol = col + direction.colDelta

                // 원래 칸 값이 0이면 이번 분기에서 새 선로를 놓은 것입니다.
                // 이미 선로가 있던 칸이면 그 선로를 그대로 지나가는 분기입니다.
                val original = grid[nextRow][nextCol]
                val isEmptyCell = original == 0

                // 빈칸인 경우에만 후보 선로를 실제 격자에 반영합니다.
                // 이렇게 해야 다음 재귀 호출이 이 칸의 나가는 방향을 계산할 수 있습니다.
                if (isEmptyCell) {
                    grid[nextRow][nextCol] = nextTrack
                }

                // 이 칸을 이번 경로에서 이미 지나갔는지 기억해 둡니다.
                // 재방문은 허용되므로, 이미 true였던 칸은 나중에 false로 되돌리면 안 됩니다.
                val wasVisited = visited[nextRow][nextCol]
                if (!wasVisited) {
                    visited[nextRow][nextCol] = true
                }

                // 처음 방문한 칸이면 방문한 선로 수가 1 늘어납니다.
                val nextVisitedCount = visitedCount + if (wasVisited) 0 else 1

                // 빈칸에 새 선로를 놓았다면, 이 분기에서 전체 선로 수 역시 1 늘어납니다.
                val nextTotalTrackCount = totalTrackCount + if (isEmptyCell) 1 else 0

                // 다음 칸으로 들어갑니다.
                // 현재 칸에서 `direction`으로 나갔으므로, 다음 칸 입장에서는 `direction.opposite()` 방향에서 들어온 것입니다.
                count += dfs(
                    grid,
                    tracks,
                    nextRow,
                    nextCol,
                    direction.opposite(),
                    visited,
                    nextVisitedCount,
                    inPath,
                    nextTotalTrackCount,
                )

                // 이번 호출에서 처음 방문 처리한 칸만 방문 표시를 되돌립니다.
                // 원래 방문되어 있던 칸까지 false로 만들면 같은 경로의 과거 방문 정보가 깨집니다.
                if (!wasVisited) {
                    visited[nextRow][nextCol] = false
                }

                // 이번 분기에서 빈칸에 놓은 선로만 다시 0으로 되돌립니다.
                // 이 원복이 있어야 다음 후보 선로가 같은 칸에서 독립적으로 시도됩니다.
                if (isEmptyCell) {
                    grid[nextRow][nextCol] = 0
                }
            }
        }

        // 이 재귀 호출이 맡은 상태 탐색이 끝났으므로 현재 경로 사용 표시를 지웁니다.
        // 이 원복이 있어야 다른 분기에서 같은 상태를 다시 합법적으로 사용할 수 있습니다.
        if (enteredDirectionIndex != null) {
            inPath[row][col][enteredDirectionIndex] = false
        }

        // 현재 상태에서 가능한 모든 분기의 유효 배치 수를 반환합니다.
        return count
    }

    /**
     * 완성된 `grid`의 모든 선로 연결이 닫혀 있는지 검사합니다.
     *
     * DFS가 도착점에 도착했다는 사실만으로는 충분하지 않습니다.
     * 어떤 선로의 열린 방향이 빈칸, 장애물, 격자 밖으로 새고 있을 수 있기 때문입니다.
     */
    fun allConnectionsValid(
        grid: Array<IntArray>,
        tracks: Map<Int, Set<Direction>>,
    ): Boolean {
        // 모든 칸을 훑으며 선로 칸만 검사합니다.
        for (row in grid.indices) {
            for (col in grid[row].indices) {
                val track = grid[row][col]

                // 빈칸과 장애물은 선로 연결 검증 대상이 아닙니다.
                if (track <= 0) continue

                // 현재 선로가 열고 있는 모든 방향을 검사합니다.
                for (direction in tracks[track].orEmpty()) {
                    val nextRow = row + direction.rowDelta
                    val nextCol = col + direction.colDelta

                    // 열린 방향이 격자 밖이면 보통 잘못된 배치입니다.
                    // 다만 시작점의 왼쪽 바깥과 도착점의 오른쪽/아래 바깥은 이 코드의 출발/도착 모델에서 허용합니다.
                    if (!(nextRow in grid.indices && nextCol in grid[nextRow].indices)) {
                        if (!isAllowedOuterOpening(grid, row, col, direction)) return false
                        continue
                    }

                    // 현재 선로가 어떤 방향으로 열렸다면, 이웃 선로는 반드시 그 반대 방향을 열어야 합니다.
                    // 이 조건이 깨지면 겉으로는 도착했더라도 선로 배치 자체는 완성된 상태가 아닙니다.
                    val nextTrack = grid[nextRow][nextCol]
                    if (tracks[nextTrack]?.contains(direction.opposite()) != true) return false
                }
            }
        }

        // 모든 선로의 모든 열린 방향이 합법적으로 맞물렸습니다.
        return true
    }

    /**
     * 최종 연결 검증에서 격자 밖으로 열려 있어도 허용할 예외를 정합니다.
     *
     * 시작 칸은 왼쪽 바깥에서 기차가 들어오는 입구를 허용하고,
     * 도착 칸은 오른쪽 또는 아래쪽으로 남는 출구를 허용합니다.
     */
    fun isAllowedOuterOpening(
        grid: Array<IntArray>,
        row: Int,
        col: Int,
        direction: Direction,
    ): Boolean {
        val isStart = row == 0 && col == 0
        val isEnd = row == grid.lastIndex && col == grid[row].lastIndex

        return (isStart && direction == Direction.LEFT) ||
                (isEnd && (direction == Direction.RIGHT || direction == Direction.DOWN))
    }

    /**
     * 방향은 이름만 있으면 부족합니다.
     * 격자 문제에서 방향은 다음 칸 좌표를 만들기 위한 `(rowDelta, colDelta)`와,
     * 다음 칸 입장에서의 진입 방향을 계산하기 위한 `opposite()`을 함께 가져야 합니다.
     */
    enum class Direction(val rowDelta: Int, val colDelta: Int) {
        UP(-1, 0),
        DOWN(1, 0),
        LEFT(0, -1),
        RIGHT(0, 1);

        // 현재 칸에서 어떤 방향으로 나가면, 다음 칸 입장에서는 그 반대 방향에서 들어온 것입니다.
        fun opposite(): Direction = when (this) {
            UP -> DOWN
            DOWN -> UP
            LEFT -> RIGHT
            RIGHT -> LEFT
        }
    }
}

fun main() {
    // 로컬 replay용 예제입니다. 프로그래머스 제출에는 필요 없지만, 학습 저장소에서는 검증 경로가 됩니다.
    val testCases = listOf(
        // 기본 3x3 예제입니다. 빈칸, 장애물, 도착점 2번 선로가 함께 나옵니다.
        Pair(
            arrayOf(
                intArrayOf(1, 0, -1),
                intArrayOf(0, 0, 7),
                intArrayOf(0, 0, 2),
            ),
            2
        ),
        // 기존 선로와 빈칸이 섞인 가로형 예제입니다. 최종 연결 검증이 없으면 과대계산되기 쉽습니다.
        Pair(
            arrayOf(
                intArrayOf(1, 0, 0, 0, 0, -1, -1),
                intArrayOf(-1, 0, 0, 1, 0, 0, 1),
            ),
            2
        ),
        // 3번 선로가 포함된 예제입니다. 이동 중 3번 처리와 최종 네 방향 연결 검증을 함께 확인합니다.
        Pair(
            arrayOf(
                intArrayOf(1, 0, 0, 0, 0),
                intArrayOf(0, 0, 3, 0, 2),
                intArrayOf(0, 0, 0, 0, 2),
            ),
            4
        ),
        // 빈칸이 많은 예제입니다. `count += dfs(...)`로 경우의 수를 누적하는 구조를 검증합니다.
        Pair(
            arrayOf(
                intArrayOf(1, 0, 0, 0),
                intArrayOf(0, 0, 0, 0),
                intArrayOf(0, 0, 0, 0),
                intArrayOf(0, 0, 0, 0),
                intArrayOf(0, 0, 0, 1),
            ),
            644
        ),
        // 이미 선로가 완성된 예제입니다. 선로를 새로 놓지 않는 것도 한 가지 방법입니다.
        Pair(
            arrayOf(
                intArrayOf(1, 7),
                intArrayOf(0, 2),
            ),
            1
        ),
        // 불가능한 배치입니다. 갈 수 있는 길이 막혀 있으면 0을 반환해야 합니다.
        Pair(
            arrayOf(
                intArrayOf(1, -1, 0, 0),
                intArrayOf(-1, 0, 0, 0),
                intArrayOf(0, 0, 0, -1),
                intArrayOf(0, 0, -1, 1),
            ),
            0
        ),
    )

    // `assert`는 JVM에서 `-ea` 옵션을 켜야 실패를 던집니다.
    // 검증할 때는 `java -ea -jar ...` 형태로 실행해야 합니다.
    testCases.forEach { testCase ->
        assert(
            Solution().solution(testCase.first) == testCase.second
        )
    }
}
