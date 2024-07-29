package main

import "fmt"

func BackingArray() {
	// 배열 생성 (이것이 backing array가 됨)
	original := [5]int{1, 2, 3, 4, 5}

	// 슬라이스 생성 (original 배열의 일부를 참조)
	slice1 := original[1:4]
	fmt.Printf("slice1: %v, len: %d, cap: %d\n", slice1, len(slice1), cap(slice1))

	// 새로운 슬라이스 생성 (같은 backing array를 공유)
	slice2 := original[2:5]
	fmt.Printf("slice2: %v, len: %d, cap: %d\n", slice2, len(slice2), cap(slice2))

	// slice1 수정 (backing array도 수정됨)
	slice1[1] = 20
	fmt.Printf("After modifying slice1:\n")
	fmt.Printf("original: %v\n", original)
	fmt.Printf("slice1: %v\n", slice1)
	fmt.Printf("slice2: %v\n", slice2)

	// slice1에 요소 추가 (새로운 backing array 생성)
	slice1 = append(slice1, 30, 40)
	fmt.Printf("\nAfter appending to slice1:\n")
	fmt.Printf("original: %v\n", original)
	fmt.Printf("slice1: %v, len: %d, cap: %d\n", slice1, len(slice1), cap(slice1))
	fmt.Printf("slice2: %v\n", slice2)

	// Output:
	//
	// slice1: [2 3 4], len: 3, cap: 4
	// slice2: [3 4 5], len: 3, cap: 3
	// After modifying slice1:
	// original: [1 2 20 4 5]
	// slice1: [2 20 4]
	// slice2: [20 4 5]
	//
	// After appending to slice1:
	// original: [1 2 20 4 5]
	// slice1: [2 20 4 30 40], len: 5, cap: 8
	// slice2: [20 4 5]
}
