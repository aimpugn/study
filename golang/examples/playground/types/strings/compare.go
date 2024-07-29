package strings

import "fmt"

func min(a, b string) string {
	if a < b {
		return a
	}
	return b
}

func Compare() {
	fmt.Println(min("llr", "dl"))
	fmt.Println(min("l", "dxxxxx"))
	fmt.Println(min("dxy", "dxxxxx"))
}
