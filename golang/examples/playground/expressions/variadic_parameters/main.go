package main

import "fmt"

func main() {
	tmp := [][]string{
		{
			"test",
			"test2",
		},
		{
			"test2",
			"test3",
		},
	}
	variadicParameter(tmp...)
	variadicParameter([]string{"test, test2"}, []string{"test2", "test3"})
}

func variadicParameter(values ...[]string) {
	fmt.Println(values)
	for _, value := range values {
		fmt.Println(value)
	}
}
