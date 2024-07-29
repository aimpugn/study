package main

import "fmt"

func main() {
	fmt.Println(globalVar)
	fmt.Println(globalVar)
	fmt.Println(globalVar)
	fmt.Println(getGlobalVar())
	fmt.Println(getGlobalVar())
	fmt.Println(getGlobalVar())
}

type Simple struct {
	Code string
	Name string
}

var globalVar = map[string]*Simple{
	"001": {Code: "001", Name: "Some Name1"},
	"002": {Code: "002", Name: "Some Name2"},
	"003": {Code: "003", Name: "Some Name3"},
}

func getGlobalVar() map[string]*Simple {
	return map[string]*Simple{
		"001": {Code: "001", Name: "Some Name1"},
		"002": {Code: "002", Name: "Some Name2"},
		"003": {Code: "003", Name: "Some Name3"},
	}
}
