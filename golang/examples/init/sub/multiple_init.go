package sub

import (
	"fmt"
)

// 첫 번째 init 함수
func init() {
	fmt.Println("multiple_init 첫 번째 init 함수 실행")
	// 초기화 작업 1
}

// 두 번째 init 함수
func init() {
	fmt.Println("multiple_init 두 번째 init 함수 실행")
	// 초기화 작업 2
}

// 세 번째 init 함수
func init() {
	fmt.Println("multiple_init 세 번째 init 함수 실행")
	// 초기화 작업 3
}
