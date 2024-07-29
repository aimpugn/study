package main

func dfs(n, m, x, y, r, c, k int, path string, visited map[[3]int]string) string {
	if x == r && y == c && k == 0 {
		return path
	}
	if x < 1 || x > n || y < 1 || y > m || k < 0 {
		return ""
	}

	state := [3]int{x, y, k}
	if p, ok := visited[state]; ok && p <= path {
		return ""
	}
	visited[state] = path

	directions := []struct {
		dx, dy int
		move   string
	}{{-1, 0, "u"}, {1, 0, "d"}, {0, -1, "l"}, {0, 1, "r"}}

	bestPath := ""
	for _, dir := range directions {
		newPath := dfs(n, m, x+dir.dx, y+dir.dy, r, c, k-1, path+dir.move, visited)
		if newPath == "" {
			continue
		}
		if bestPath == "" || newPath < bestPath {
			bestPath = newPath
		}
	}
	return bestPath
}

func solutionByDFS(n int, m int, x int, y int, r int, c int, k int) string {
	visited := make(map[[3]int]string)
	result := dfs(n, m, x, y, r, c, k, "", visited)
	if result == "" {
		return "impossible"
	}
	return result
}
