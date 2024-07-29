package main

import (
	"errors"
	"fmt"
)

func Errorf() {
	err := errors.New("test error")
	fmt.Println(fmt.Errorf("%w", err))
}
