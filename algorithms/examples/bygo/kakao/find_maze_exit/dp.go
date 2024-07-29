package main

import (
	"fmt"
)

func minStr(a, b string) string {
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

func solutionByDP(n int, m int, x int, y int, r int, c int, k int) string {
	dp := make(map[[3]string]string)

	// Initialize base case
	dp[[3]string{fmt.Sprint(x), fmt.Sprint(y), fmt.Sprint(k)}] = ""

	// Dynamic Programming
	for step := k; step >= 0; step-- {
		for i := 1; i <= n; i++ {
			for j := 1; j <= m; j++ {
				moves := []struct {
					dx, dy int
					move   string
				}{{-1, 0, "u"}, {1, 0, "d"}, {0, -1, "l"}, {0, 1, "r"}}

				for _, move := range moves {
					nx, ny, nk := i+move.dx, j+move.dy, step-1
					if nx >= 1 && nx <= n && ny >= 1 && ny <= m && nk >= 0 {
						prevState := [3]string{fmt.Sprint(i), fmt.Sprint(j), fmt.Sprint(step)}
						currState := [3]string{fmt.Sprint(nx), fmt.Sprint(ny), fmt.Sprint(nk)}
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
	result := dp[[3]string{fmt.Sprint(r), fmt.Sprint(c), "0"}]
	if result == "" {
		return "impossible"
	}
	return result
}
