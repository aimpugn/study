package main

import (
	"fmt"
	"math"
	"math/big"
)

func main() {
	fmt.Println(
		"new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil)",
		new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil),
	)
	fmt.Println(
		"new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil)",
		new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil),
	)
	fmt.Println(
		"10**1",
		math.Pow(10, 1),
	)
	fmt.Println(
		"10**2",
		math.Pow(10, 2),
	)

	f := new(big.Float).SetInt64(int64(1000))
	f.Quo(f, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil)))
	tmp, _ := f.Float64()
	fmt.Println(fmt.Sprintf("%f", tmp))
	fmt.Println(fmt.Sprintf("%d", exp[0]))
}

var exp = map[int]*big.Int{
	1: new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil), // 10^1
	2: new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil), // 10^2
	3: new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil), // 10^3
	4: new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil), // 10^4
}
