package main

import (
	"fmt"
	"runtime"
	"strconv"
)

func main() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Println("The specific set of memory included:")
	fmt.Println("Alloc:", formatNumber(memStats.Alloc))
	fmt.Println("TotalAlloc:", formatNumber(memStats.TotalAlloc))
	fmt.Println("Sys:", formatNumber(memStats.Sys))
	fmt.Println("HeapAlloc:", formatNumber(memStats.HeapAlloc))
	fmt.Println("HeapSys:", formatNumber(memStats.HeapSys))
	fmt.Println("HeapIdle:", formatNumber(memStats.HeapIdle))
	fmt.Println("HeapInuse:", formatNumber(memStats.HeapInuse))
	fmt.Println("HeapReleased:", formatNumber(memStats.HeapReleased))
	fmt.Println("HeapObjects:", formatNumber(memStats.HeapObjects))
}

func formatNumber(n uint64) string {
	str := strconv.FormatUint(n, 10)
	length := len(str)
	if length <= 3 {
		return str
	}

	commaCount := (length - 1) / 3
	formatted := make([]byte, length+commaCount)
	copy(formatted, str)

	for i := length - 1; i >= 0; i-- {
		if (length-i-1)%3 == 0 && i != length-1 {
			formatted[i+commaCount] = ','
			commaCount--
		}
		formatted[i+commaCount] = str[i]
	}

	return string(formatted)
}
