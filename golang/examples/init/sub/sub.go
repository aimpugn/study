package sub

import "fmt"

func init() {
	fmt.Println("sub.init() called")
}

func ExportedTestFuncOfSub() {
	fmt.Println("ExportedTestFuncOfSub")
}
