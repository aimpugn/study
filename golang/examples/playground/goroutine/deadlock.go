package main

import "fmt"

func DeadLock() {
	w := DeadLockWrapper{
		receiver: make(chan string),
		waiter:   make(chan string),
	}

	// 교착 상태(Deadlock)는 여러 고루틴(goroutine)이 서로가 다음 단계로 진행하기를 기다리면서
	// 아무도 진행하지 못하는 상황을 의미한다.

	defer w.close() // main 함수의 나머지 부분이 모두 실행된 후에 호출

	for _, num := range []int{1, 2, 3, 4, 5} {
		fmt.Println(num)
		if num == 4 {
			// main 함수 내에서 w.waiter 채널에 메시지를 보내는 동작이 블록된다
			w.waiter <- fmt.Sprintf("Done, num is %d", num) // 블록
			// 하지만 close 메서드가 실행되기 전까지는 이 채널에서 메시지를 수신할 고루틴이 없다
			// 즉, main 고루틴이 w.waiter 채널에 메시지를 보내려고 할 때,
			// 이 채널에서 메시지를 받을 고루틴이 없어서 데드락 발생
		}
	}
}

func (w *DeadLockWrapper) close() error {
	message := <-w.waiter
	fmt.Println("close function is called, message is", message)
	return nil
}

type DeadLockWrapper struct {
	receiver chan string
	waiter   chan string
}
