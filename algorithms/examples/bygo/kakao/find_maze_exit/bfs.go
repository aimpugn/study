package main

import (
	"container/heap"
	"fmt"
)

type Item struct {
	path    string
	x, y, k int
}
type PriorityQueue []*Item

func (pq *PriorityQueue) Len() int { return len(*pq) }

func (pq *PriorityQueue) Less(i, j int) bool {
	return (*pq)[i].path < (*pq)[j].path
}

func (pq *PriorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Item)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func solutionByBFS(n int, m int, x int, y int, r int, c int, k int) string {
	pq := make(PriorityQueue, 1)
	pq[0] = &Item{path: "", x: x, y: y, k: k}
	heap.Init(&pq)

	visited := make(map[string]string)

	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		state := fmt.Sprintf("%d-%d-%d", item.x, item.y, item.k)

		if prevPath, exists := visited[state]; exists && prevPath <= item.path {
			continue
		}

		if item.x == r && item.y == c && item.k == 0 {
			return item.path
		}

		visited[state] = item.path

		moves := []struct {
			dx, dy int
			move   string
		}{{-1, 0, "u"}, {1, 0, "d"}, {0, -1, "l"}, {0, 1, "r"}}

		for _, move := range moves {
			nx, ny, nk := item.x+move.dx, item.y+move.dy, item.k-1
			if nx >= 1 && nx <= n && ny >= 1 && ny <= m && nk >= 0 {
				newState := fmt.Sprintf("%d-%d-%d", nx, ny, nk)
				newPath := item.path + move.move
				if prevPath, exists := visited[newState]; !exists || prevPath > newPath {
					heap.Push(&pq, &Item{path: newPath, x: nx, y: ny, k: nk})
				}
			}
		}
	}
	return "impossible"
}

