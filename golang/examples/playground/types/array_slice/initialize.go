package array_slice

import "fmt"

func InitBySubSlice() {
	a := []int{1, 2, 3, 4, 5}
	s := a[1:3] // [2 3]

	fmt.Println("len(a)", a, "len(s)", len(s), "s=a[1:3]", s)
}
