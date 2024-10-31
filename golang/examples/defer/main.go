// References:
// - https://go.dev/ref/spec#Defer_statements
// - https://go.dev/blog/defer-panic-and-recover
package main

import (
	"fmt"
)

func foo() {
	defer func() {
		fmt.Println("Deferred function executed")
	}()

	fmt.Println("Before panic")
	panic("Panic occurred")
	fmt.Println("After panic") // 이 줄은 실행되지 않음
}

func main() {
	foo()
	fmt.Println("Main function continued")
}

//Output:
// Before panic
// Deferred function executed
// panic: Panic occurred
