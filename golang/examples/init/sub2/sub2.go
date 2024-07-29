package sub2

import (
	"fmt"
	"math/rand"
)

func init() {
	fmt.Println("sub2.init() called")
}

func ExportedTestFuncOfSub2() {
	fmt.Printf("ExportedTestFuncOfSub2: %d\n", rand.Intn(10))
}
