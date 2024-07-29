package strings

import "fmt"

func StringIndex() {
	test1 := "hello"
	fmt.Println(test1[0]) // 104
	fmt.Println(test1[1]) // 101
	fmt.Println(test1[2]) // 108
	fmt.Println(test1[3]) // 108
	fmt.Println(test1[4]) // 111
	// fmt.Println(test1[5]) // panic: runtime error: index out of range [5] with length 5 [recovered]

	fmt.Println(test1[0:4]) // hell
	fmt.Println(test1[0:5]) // hello
}
