package structs

import "fmt"

func Receiver() {
	t := ByValue{}

	t.SetValue1("test")
	t.SetValue3("1", "2", "3")
	fmt.Println(t) // { 0 []}

	fmt.Println("==================================")

	p := &ByPointer{}

	p.SetValue1("test")
	p.SetValue3("1", "2", "3")
	fmt.Println(p) // &{test 0 [1 2 3]}

	fmt.Println("==================================")

	// p2 := ByPointer{}
	p2 := ByPointer{}
	p2Pointer := Pointer(p2)
	// p2Pointer := &p2
	p2Pointer.SetValue1("pointer of value")
	p2Pointer.SetValue3("111", "112", "113")
	fmt.Println(p2) // {pointer of value 0 [111 112 113]}
}

func Pointer[T any](v T) *T {
	return &v
}

type ByValue struct {
	value1 string
	value2 int
	value3 []string
}

func (r ByValue) SetValue1(s string) {
	r.value1 = s
}

func (r ByValue) SetValue2(i int) {
	r.value2 = i
}

func (r ByValue) SetValue3(values ...string) {
	for _, value := range values {
		r.value3 = append(r.value3, value)
	}
}

type ByPointer struct {
	value1 string
	value2 int
	value3 []string
}

func (r *ByPointer) SetValue1(s string) {
	r.value1 = s
}

func (r *ByPointer) SetValue2(i int) {
	r.value2 = i
}

func (r *ByPointer) SetValue3(values ...string) {
	for _, value := range values {
		r.value3 = append(r.value3, value)
	}
}
