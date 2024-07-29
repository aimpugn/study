package design_and_philosophy

import (
	"fmt"
	"time"
)

func printCount(c chan int) {
	num := 0
	for num >= 0 {
		num = <-c // 채널을 통해 고루틴 간 데이터를 전달하고 동기화
		fmt.Print(num, " ")
	}
}

func Concurrency() {
	c := make(chan int)
	go printCount(c) // 고루틴을 사용하여 비동기적으로 함수를 실행

	for i := 0; i < 10; i++ {
		c <- i // 채널을 통해 고루틴 간 데이터를 전달하고 동기화
	}

	time.Sleep(time.Millisecond * 1)
	fmt.Println("End of main")
}
