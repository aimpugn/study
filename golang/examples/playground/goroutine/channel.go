package main

import "fmt"

func Channel() {
	w := ChannelWrapper{
		receiver: make(chan string),
		waiter:   make(chan string),
	}

	defer w.close()

	w.loop() // 별도의 고루틴으로 실행되면서 main 고루틴과 독립적으로 동작

	// defer로 등록된 w.close() 메서드가 호출되면서 w.waiter에서 메시지를 수신하고 처리한다.
}

func (w *ChannelWrapper) loop() {
	// 별도의 고루틴으로 main 고루틴이 w.waiter 채널에 메시지를 보내는 작업에 블록되지 않게 된다
	go func() {
		for _, num := range []int{1, 2, 3, 4, 5} {
			fmt.Println(num)
			if num == 4 {
				// w.waiter 채널로 메시지를 보내는 동작이 별도의 고루틴에서 실행된다
				// close 메서드에서 이 메시지를 기다리는 동안에도 프로그램의 다른 부분이 정상적으로 실행될 수 있다
				w.waiter <- fmt.Sprintf("Done, num is %d", num)
			}
		}
	}()
}

func (w *ChannelWrapper) close() error {
	message := <-w.waiter
	fmt.Println("close function is called, message is", message)
	return nil
}

type ChannelWrapper struct {
	receiver chan string
	waiter   chan string
}
