package main

import "fmt"

// solutionByDP3
//
// 1. 가지치기(Pruning) 추가: 맨해튼 거리를 사용하여 불가능한 상태를 미리 제거
// 2. 불필요한 연산 제거: 상태 업데이트를 할 때, 이미 계산된 상태는 무시할 수 있다
// 3. 상태 저장 최적화: 현재 구현에서는 문자열을 상태로 저장하고 있다. 이를 개선하여 메모리 사용량과 연산 시간을 줄일 수 있다
func solutionByDP3(n int, m int, x int, y int, r int, c int, k int) string {
	type State struct {
		X, Y, K int
	}

	minStr := func(a, b string) string {
		if a == "" {
			return b
		}
		if b == "" {
			return a
		}
		if a < b {
			return a
		}
		return b
	}

	dp := make(map[State]string)

	dp[State{x, y, k}] = ""

	moves := []struct {
		dx, dy int
		move   string
	}{
		{-1, 0, "u"},
		{1, 0, "d"},
		{0, -1, "l"},
		{0, 1, "r"},
	}

	for step := k; step >= 0; step-- {
		for i := 1; i <= n; i++ {
			for j := 1; j <= m; j++ {
				for _, move := range moves {
					nx, ny, nk := i+move.dx, j+move.dy, step-1
					if nx >= 1 && nx <= n && ny >= 1 && ny <= m && nk >= 0 {
						manhattan := abs(nx-r) + abs(ny-c)
						if nk < manhattan || (nk-manhattan)%2 != 0 {
							continue
						}

						prevState := State{i, j, step}
						currState := State{nx, ny, nk}
						if prevPath, exists := dp[prevState]; exists {
							newPath := prevPath + move.move
							fmt.Println(nx, ny, nk, prevPath, newPath)
							dp[currState] = minStr(dp[currState], newPath)
						}
					}
				}
			}
		}
	}

	result, exists := dp[State{r, c, 0}]
	if !exists {
		return "impossible"
	}
	return result
}
