package main

import "fmt"

func main() {
	// fmt.Println(solutionByDFS(3, 4, 2, 3, 3, 1, 5)) // Should return "dllrl"
	// fmt.Println(solutionByDFS(2, 2, 1, 1, 2, 2, 2)) // Should return "dr"
	// fmt.Println(solutionByDFS(3, 3, 1, 2, 3, 3, 4)) // Should return "impossible"

	// fmt.Println(solutionByBFS(3, 4, 2, 3, 3, 1, 5)) // Should return "dllrl"
	// fmt.Println(solutionByBFS(2, 2, 1, 1, 2, 2, 2)) // Should return "dr"
	// fmt.Println(solutionByBFS(3, 3, 1, 2, 3, 3, 4)) // Should return "impossible"

	// fmt.Println(solutionByDP(3, 4, 2, 3, 3, 1, 5)) // Should return "dllrl"
	// fmt.Println(solutionByDP(2, 2, 1, 1, 2, 2, 2)) // Should return "dr"
	// fmt.Println(solutionByDP(3, 3, 1, 2, 3, 3, 4)) // Should return "impossible"

	// fmt.Println(solutionByDP2(3, 4, 2, 3, 3, 1, 5)) // Should return "dllrl"
	// fmt.Println(solutionByDP2(2, 2, 1, 1, 2, 2, 2)) // Should return "dr"
	// fmt.Println(solutionByDP2(3, 3, 1, 2, 3, 3, 4)) // Should return "impossible"

	// fmt.Println(solutionByDP3(3, 4, 2, 3, 3, 1, 5)) // Should return "dllrl"
	// fmt.Println(solutionByDP3(2, 2, 1, 1, 2, 2, 2)) // Should return "dr"
	// fmt.Println(solutionByDP3(3, 3, 1, 2, 3, 3, 4)) // Should return "impossible"

	fmt.Println("================================")

	fmt.Println(solutionDPDFS(3, 4, 2, 3, 3, 1, 5)) // Should return "dllrl"
	fmt.Println(solutionDPDFS(2, 2, 1, 1, 2, 2, 2)) // Should return "dr"
	fmt.Println(solutionDPDFS(3, 3, 1, 2, 3, 3, 4)) // Should return "impossible"
}
