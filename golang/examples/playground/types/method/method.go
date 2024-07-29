package main

import "fmt"

type MyType struct{}

func Method() {
	var tmp *MyType
	fmt.Println(tmp.Test())
}

func (r *MyType) Test() bool {
	if r == nil {
		return false
	}

	return true
}
