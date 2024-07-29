package main

import (
	"fmt"
	"os"
)

func Assertion() {
	var i interface{} = 42
	j := i.(int)
	fmt.Println(j)

	tmp := map[string]interface{}{"test": "value"}
	if val, ok := tmp["test"].(string); ok {
		fmt.Println(val)
	}

	if val, ok := tmp["none"].(string); ok {
		fmt.Println(val)
	} else {
		fmt.Println("none is not a string")
	}

	none := tmp["none"]
	fmt.Println(none)
	_, ok := none.(string)
	fmt.Println(ok)

	fmt.Println(os.Getenv("GOROOT"))
}
