package main

import (
	"fmt"
	"runtime"
)

func main() {
	test1()
	fmt.Println(getFuncName())
}

func test1() {
	fmt.Println(getFuncName())
	test2()
}

func test2() {
	fmt.Println(getFuncName())
	test3()
}

func test3() {
	fmt.Println(getFuncName())
}

func getFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		return runtime.FuncForPC(pc).Name()
	}

	return "unknown"
}
