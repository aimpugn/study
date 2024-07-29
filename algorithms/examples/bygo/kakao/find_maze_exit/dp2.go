package main

type State struct {
	X, Y, K int
}

var moves = []struct {
	dx   int
	dy   int
	move string
}{
	{-1, 0, "u"},
	{1, 0, "d"},
	{0, -1, "l"},
	{0, 1, "r"},
}

func solutionByDP2(n int, m int, x int, y int, r int, c int, k int) string {
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

	// Initialize base case
	dp[State{x, y, k}] = ""

	// Dynamic Programming
	for step := k; step >= 0; step-- {
		for i := 1; i <= n; i++ {
			for j := 1; j <= m; j++ {
				for _, move := range moves {
					nx, ny, nk := i+move.dx, j+move.dy, step-1
					if nx >= 1 && nx <= n && ny >= 1 && ny <= m && nk >= 0 {
						prevState := State{i, j, step}
						currState := State{nx, ny, nk}
						if prevPath, exists := dp[prevState]; exists {
							newPath := prevPath + move.move
							dp[currState] = minStr(dp[currState], newPath)
						}
					}
				}
			}
		}
	}

	// Get the result
	result, exists := dp[State{r, c, 0}]
	if !exists {
		return "impossible"
	}
	return result
}
