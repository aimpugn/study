package main

import (
	"strings"
)

func removeHyphensLoop(uuidWithHyphens string) string {
	result := ""
	for _, char := range uuidWithHyphens {
		if char != '-' {
			result += string(char)
		}
	}
	return result
}

func removeHyphensReplaceAll(uuidWithHyphens string) string {
	return strings.ReplaceAll(uuidWithHyphens, "-", "")
}

func main() {
	BackingArray()
}
