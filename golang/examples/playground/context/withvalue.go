package main

import (
	"context"
	"fmt"
)

const myStringKey = "myKey"

type myKeyType string

const myTypeKey = myKeyType("myKey")

func WithValue() {

	parentCtx := context.Background()
	ctx1 := context.WithValue(parentCtx, myStringKey, "myStringKey")
	ctx2 := context.WithValue(parentCtx, myTypeKey, "myTypeKey")

	fmt.Println(ctx1.Value(myTypeKey))   // <nil>
	fmt.Println(ctx2.Value(myStringKey)) // <nil>

	fmt.Println(ctx1.Value(myStringKey)) // myStringKey
	fmt.Println(ctx2.Value(myTypeKey))   // myTypeKey
}
