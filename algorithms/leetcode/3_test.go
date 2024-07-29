package leetcode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLengthOfLongestSubstring(t *testing.T) {
	testCases := map[string]struct {
		input  string
		output int
	}{
		" ": {
			" ", // The answer is " "
			1,
		},
		"": {
			"", // The answer is " "
			0,
		},
		"aab": {
			"aab", // The answer is "ab"
			2,
		},
		"이미 거쳐간 문자라도, 그 문자부터 다시 체크해야 할 수 있다": {
			"dvdf", // The answer is "vdf"
			3,
		},
		"일부 문자열이 같을 수 있다": {
			"abcabcbb", // The answer is "abc"
			3,
		},
		"bbbbb": {
			"bbbbb", // The answer is "b"
			1,
		},
		"pwwkew": {
			"pwwkew", // The answer is "wke"
			3,
		},
		"abcdddefghiabcdefgh": {
			"abcdddefghiabcdefgh", // The answer is "wke"
			9,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			require.Equal(t,
				testCase.output,
				lengthOfLongestSubstring(testCase.input),
			)
		})
	}
}
