package pointer

import (
	"fmt"
	"time"
)

func NestedPointer() {
	parent := new(Parent)
	fmt.Println("1.", parent)
	testParent(parent)
	fmt.Println("2.", parent)
	testChildValue(parent.Child)
	fmt.Println("3.", parent)
	testChildPointer(&parent.Child)
	fmt.Println("4.", parent)
}

func testParent(p *Parent) {
	// `Parent` 구조체가 참조하는 메모리 위치에 직접 적용
	// 따라서 이 함수를 호출한 후 parent 인스턴스의 `Child` 필드가 변경된다
	p.Child = Child{
		B: 100,
		D: "generated in testParent function",
	}
}

func testChildValue(c Child) {
	c.D = "modified in testChildValue function"
}

func testChildPointer(c *Child) {
	// `Child` 포인터를 통해 필드를 변경하면, 이 변경사항이 포인터가 가리키는 원본 `Child` 인스턴스에 적용된다.
	// 는 `Parent` 인스턴스 내의 `Child` 필드도 포인터가 아니라 값이라 할지라도,
	// `Parent` 인스턴스 자체가 포인터를 통해 참조되고 있기 때문에 가능하다.
	c.D = "modified in testChildPointer function"
}

type Parent struct {
	Z         []interface{}
	Child     Child
	ChildMeta ChildMeta
}

type Child struct {
	A time.Time
	B int
	C int
	D string
	E string
	F string
	G time.Time
}

type ChildMeta struct {
	H bool
	I string
	J string
	K string
	L int32
}
