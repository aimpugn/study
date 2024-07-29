package array_slice

import "fmt"

func Cap() {
	numbers := []int{10, 15, 20}
	another := numbers[0:1]   // 슬라이스는 슬라이스가 생성되는 시작점부터 배열의 끝까지만 볼 수 있다.
	fmt.Println(len(another)) // 1
	fmt.Println(cap(another)) // 3

	numbers = []int{10, 15, 20}
	another = numbers[1:3] // `numbers`의 첫 번째 요소(10)를 "되돌아가서" 참조할 수 없다

	fmt.Println(len(another)) // 2
	fmt.Println(cap(another)) // 2
	// 슬라이스가 생성될 때 그 시작점 이전의 요소들을 "되돌아가서"(loop back) 볼 수 없다
	// 즉, 슬라이스는 자신의 시작점 이후의 요소들만 "볼 수 있다"
	// 슬라이스의 참조 범위가 한 번 정해지면, 그 범위를 넘어서 **이전** 요소로 확장되지 않는다
	// 따라서 슬라이스는 생성 시에 정의된 시작점 이후의 요소들에 대해서만 접근할 수 있고, 이전 요소들로는 "되돌아가" 접근할 수 없다.
}
