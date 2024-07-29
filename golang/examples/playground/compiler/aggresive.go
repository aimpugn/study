package main

import "time"

func Aggresive() {
	i := 1
	go func() {
		for {
			i++
		}
	}()
	<-time.After(1 * time.Millisecond)
	println(i)
}
