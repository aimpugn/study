package main

import (
	"fmt"

	"github.com/aimpugn/snippets/go/example_mod/sub"
	"github.com/aimpugn/snippets/go/example_mod/sub2"
)

func main() {
	sub.ExportedTestFuncOfSub()
	sub2.ExportedTestFuncOfSub2()

	fmt.Println("run main")
}
