package main

import (
	"fmt"
	"testing"
)

func TestLoop(t *testing.T) {
	done := make(chan bool)

	values := []string{"a", "b", "c"}
	for _, v := range values {
		go func() {
			// they usually print “c”, “c”, “c”,
			// instead of printing “a”, “b”, and “c” in some order.
			fmt.Println(v)
			done <- true
		}()
	}

	// wait for all goroutines to complete before exiting
	for range values {
		<-done
	}
}
