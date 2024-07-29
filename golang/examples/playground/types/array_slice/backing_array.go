package array_slice

import "fmt"

func BackingArray() {
	original := [5]int{1, 2, 3, 4, 5} // 고정된 길이를 갖는 연속적인 메모리 공간 배열 생성

	// 배열의 일부분을 참조하는 가변 길이의 구조인 슬라이스 `a`를 만들어 배열 전체를 참조
	a := original[:] // `a`의 `backing array`는 원본 배열 `original`

	// 슬라이스 `a`의 일부분을 참조하는 새로운 슬라이스 `b`를 생성
	b := a[1:4] // `b`의 `backing array`는 `a`가 참조하는 같은 원본 배열 `original`

	// `b`의 첫 번째 요소를 변경
	b[0] = 20 // `b`의 변경이 `a`와 `original`에도 영향을 미친다

	fmt.Println(original) // [1 20 3 4 5]
	fmt.Println(a)        // [1 20 3 4 5]
	fmt.Println(b)        // [20 3 4]
}
