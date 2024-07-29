package structs

import "fmt"

func Receiver2() {
	byValue := helloImplementedByValue{}
	byValue.Greet()
	byValue.Greet2()

	byValueButPointer := &helloImplementedByValue{}
	byValueButPointer.Greet()
	byValueButPointer.Greet2()

	// 왜 동작하는가?
	// 메소드 호출 시 포인터 리시버가 필요한 경우, Go 컴파일러와 런타임은 자동으로 값을 포인터로 변환해줍니다.
	// 즉, Go 컴파일러는 이를 자동으로 감지하고 byPointerButValue의 주소(&byPointerButValue)를 사용하여 메소드를 호출합니다.
	byPointerButValue := helloImplementedByPointer{}
	byPointerButValue.Greet()
	byPointerButValue.Greet2()

	byPointer := &helloImplementedByPointer{}
	byPointer.Greet()
	byPointer.Greet2()
}

type Hello interface {
	Greet()
	Greet2()
}

type helloImplementedByValue struct{}

func (r helloImplementedByValue) Greet() {
	fmt.Println("helloByValue, Hello, World!")
}

func (r helloImplementedByValue) Greet2() {
	fmt.Println("helloByValue, Hello, Hello, World!")
}

type helloImplementedByPointer struct{}

func (r *helloImplementedByPointer) Greet() {
	fmt.Println("helloByPointer, Hello, World!")
}

func (r helloImplementedByPointer) Greet2() {
	fmt.Println("helloByPointer, Hello, Hello, World!")
}
