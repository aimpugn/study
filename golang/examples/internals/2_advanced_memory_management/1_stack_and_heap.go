package main

import "fmt"

func createHugeMemoryToHeap() *[]int {
	huge := make([]int, 1000000)
	return &huge
}

func StackAndHeap() {
	// Go에서 변수들은 기본적으로 스택에 할당된다
	varAssignedToStack := 10 // 스택에 할당되는 기본적인 정수 변수

	// 변수가 함수의 스코프를 벗어날 때 여전히 접근 가능해야 하는 경우, 해당 변수는 힙에 할당된다
	pointerPossibleAssignedToHeap := new(int) // `new` 키워드를 사용하여 힙에 할당될 가능성이 있는 포인터
	*pointerPossibleAssignedToHeap = 20

	fmt.Println(varAssignedToStack, *pointerPossibleAssignedToHeap)

	// 힙에 큰 메모리 블록을 할당
	_ = createHugeMemoryToHeap()

	// 가비지 컬렉터가 필요 없는 메모리를 회수
	// 별도의 메모리 해제 작업이 필요하지 않음
}
