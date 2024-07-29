package nil

import "fmt"

func Nil() {
	var stringNil *string
	var interfaceNil interface{}

	if stringNil == nil {
		println("stringNil is nil")
	}

	if interfaceNil == nil {
		println("interfaceNil is nil")
	}

	if interfaceNil == stringNil {
		println("interfaceNil == stringNil")
	} else {
		println("interfaceNil != stringNil")
	}

	var tmp map[string]interface{}

	if val, ok := tmp["none"]; ok {
		fmt.Println(val)
	} else {
		fmt.Println("not exists")
	}

	if tmp == nil {
		fmt.Println("var tmp map[string]interface{} is nil")
	} else {
		fmt.Println("var tmp map[string]interface{} is NOT nil")
	}

	tmp2 := []string{}
	if tmp2 == nil {
		fmt.Println("var tmp2 []string is nil")
	} else {
		fmt.Println("var tmp2 []string is NOT nil")
	}

	variadicParameter()
	variadicParameter("a", "b", "c")

	var tmp3 []string
	if tmp3 == nil {
		fmt.Println("var tmp3 []string is nil", len(tmp3) == 0)
	} else {
		fmt.Println("var tmp3 []string is NOT nil")
	}
}

func variadicParameter(args ...string) {
	if args == nil {
		fmt.Println("variadicParameter args []string is nil", len(args))
	} else {
		fmt.Println("variadicParameter args []string is NOT nil", args)
	}
}
