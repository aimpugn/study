package main

import "fmt"

func main() {
	fmt.Println(test())
	// test 0
	// "test"는 직접 설정한 값
	// 0은 초기화된 기본값
}

func test() (tmp string, tmp2 int) {
	tmp = "test"

	return
}
