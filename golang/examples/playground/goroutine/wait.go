package main

import (
	"fmt"
	"strconv"
)

func Wait() {
	w := WaitWrapper{
		receiver: make(chan string),
		waiter:   make(chan string),
	}

	// 1. main 함수 종료 직전에 close 메서드 호출. 프로그램이 종료될 때까지 대기.
	//    main 함수의 모든 고루틴들이 완료될 때까지 기다린 후, w.close()가 실행되어
	//    w.waiter 채널에서 메시지를 기다린다.
	defer w.close()
	go w.loop() // 2. loop 메서드를 별도의 고루틴에서 실행. receiver 채널에서 메시지를 기다리고 처리.

	go func() {
		for _, num := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
			w.receiver <- fmt.Sprintf("%d", num) // 3. receiver 채널에 숫자 1부터 8까지 문자열로 전송.
		}
	}() // 이 익명 고루틴은 별도로 실행되어 receiver 채널로 데이터를 보냄.

	// 이 main 함수 끝나기 전에 w.close() 함수가 실행된다.
	// 즉, w.close() 함수가 main 함수의 종료를 지연시키는 마지막 작업이 된다.
	// 따라서 w.close() 메서드에서 w.waiter 채널로부터 메시지를 수신하는 것은
	// main 함수가 종료되기 전의 마지막 단계가 된다.
}

type WaitWrapper struct {
	receiver chan string
	waiter   chan string
}

func (w *WaitWrapper) loop() {
	for msg := range w.getReceiver() { // 4. receiver 채널에서 메시지를 받아서 순회.
		num, _ := strconv.Atoi(msg)  // 메시지(문자열)를 정수로 변환.
		fmt.Println("received", num) // 받은 숫자를 출력.
		if num == 4 {
			w.waiter <- fmt.Sprintf("Done, num is %d", num) // 5. 숫자가 4일 때 waiter 채널에 메시지 전송.
		}
	}
}

// close 함수는
func (w *WaitWrapper) close() error {
	message := <-w.waiter // 6. waiter 채널에서 메시지 수신 대기. "Done, num is 4" 메시지를 받는 순간 출력 후 대기 종료.
	fmt.Println("close function is called, message is", message)
	return nil
}

func (w *WaitWrapper) getReceiver() <-chan string {
	return w.receiver // 7. receiver 채널을 수신 전용으로 반환.
}
