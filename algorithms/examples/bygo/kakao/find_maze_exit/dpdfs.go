package main

type Cell struct {
	row   int
	col   int
	mvCnt int
}

var directions = []struct {
	toRow    int
	toColumn int
	keyword  string
}{
	// `d`, `l`, `r`, `u의` 순서가 중요하다
	// 가령 `u`, `r`, `l`, `d` 순서로 순회하면 시간 초과가 발생한다
	{1, 0, "d"},
	{0, -1, "l"},
	{0, 1, "r"},
	{-1, 0, "u"},
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func dpdfs(rowLimit, columnLimit, currRow, currColumn, dstRow, dstColumn, mvCnt int, dp map[Cell]string, backtrackingPath string, lexicographicallyBest *string) {
	if mvCnt < 0 {
		return
	}

	manhattanDistance := abs(dstRow-currRow) + abs(dstColumn-currColumn)
	if mvCnt < manhattanDistance || (mvCnt-manhattanDistance)%2 != 0 {
		return
	}

	currentCell := Cell{currRow, currColumn, mvCnt}
	if vlaue, ok := dp[currentCell]; ok && vlaue != "" && vlaue <= backtrackingPath {
		return
	}

	// 목적지에 도착한 경우
	if currRow == dstRow && currColumn == dstColumn && mvCnt == 0 {
		dp[currentCell] = backtrackingPath
		// `lexicographically`가 비어있으면 초기값 개념으로 현재 경로를 넣어주고,
		// 현재 경로가 기존 `lexicographically`보다 값이 더 작다면, 그 더 작은 값으로 최적값을 변경
		if *lexicographicallyBest == "" || backtrackingPath < *lexicographicallyBest {
			*lexicographicallyBest = backtrackingPath
		}
		return
	}

	dp[currentCell] = backtrackingPath

	// 현재 Cell로부터 가능한 모든 방향을 탐색
	for _, direction := range directions {
		newRow, newColumn := currRow+direction.toRow, currColumn+direction.toColumn
		if newRow >= 1 && newRow <= rowLimit && newColumn >= 1 && newColumn <= columnLimit {
			// dfs 안에서 체크하게 될 경로를 추가한다
			newBacktrackingPath := backtrackingPath + direction.keyword
			if *lexicographicallyBest == "" || newBacktrackingPath <= *lexicographicallyBest {
				dpdfs(rowLimit, columnLimit, newRow, newColumn, dstRow, dstColumn, mvCnt-1, dp, newBacktrackingPath, lexicographicallyBest)
			}
		}
	}
}

func solutionDPDFS(n, m, x, y, r, c, k int) string {
	// 각 `Cell`의 최적 경로를 저장하기 위한 DP table(Memoization)
	dp := make(map[Cell]string)
	// 현재 DFS 함수가 탐색 중인 경로를 나타낸다
	var backtrackingPath string
	// `lexicographicallyBest`는 `dp`(map[Cell]string)으로, 특정 `Cell`에 도달할 수 있는 최적의 경로를 문자열로 저장
	// 사전 순으로 가장 빠른 경로를 찾기 위해 사용
	var lexicographicallyBest string

	dpdfs(n, m, x, y, r, c, k, dp, backtrackingPath, &lexicographicallyBest)

	if lexicographicallyBest == "" {
		return "impossible"
	}
	return lexicographicallyBest
}
