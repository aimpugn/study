package main

import (
	"fmt"
	"math/rand"
	"os"
)

func RandInt() {
	// Seed 설정
	seed := int64(os.Getpid()) * int64(rand.Intn(999)+1)
	r := rand.New(rand.NewSource(seed))

	// 1000부터 9999까지의 난수 생성
	randomNumber := r.Intn(9000) + 1000
	fmt.Println(randomNumber)
}
