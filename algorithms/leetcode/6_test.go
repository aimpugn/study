package leetcode

import (
	"fmt"
	"testing"
)

func TestConvert(t *testing.T) {
	testCases := map[string]struct {
		input    string
		numRows  int
		expected string
	}{
		" -> ": {
			"",
			1,
			"",
		},
		"A -> A": {
			"A",
			1,
			"A",
		},
		"PAYPALISHIRING -> PAHNAPLSIIGYIR": {
			"PAYPALISHIRING",
			3,
			"PAHNAPLSIIGYIR",
		},
		"PAYPALISHIRING -> PINALSIGYAHRPI": {
			"PAYPALISHIRING",
			4,
			"PINALSIGYAHRPI",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			fmt.Println(convert(testCase.input, testCase.numRows), testCase.expected)
		})
	}
}
