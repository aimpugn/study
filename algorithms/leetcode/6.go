package leetcode

import (
	"strings"
)

// The string "PAYPALISHIRING" is written in a zigzag pattern on a given number of rows like this:
// (you may want to display this pattern in a fixed font for better legibility)
//
// P   A   H   N
// A P L S I I G
// Y   I   R
// And then read line by line: "PAHNAPLSIIGYIR"
//
// Write the code that will take a string and make this conversion given a number of rows:
//
// string convert(string s, int numRows);
//
// Example 1:
//
//	Input: s = "PAYPALISHIRING", numRows = 3
//	Output: "PAHNAPLSIIGYIR"
//
// Example 2:
//
//	Input: s = "PAYPALISHIRING", numRows = 4
//	Output: "PINALSIGYAHRPI"
//
// Explanation:
// P     I    N
// A   L S  I G
// Y A   H R
// P     I
//
// Example 3:
//
//	Input: s = "A", numRows = 1
//	Output: "A"
//
// Constraints:
//
// 1 <= s.length <= 1000
// s consists of English letters (lower-case and upper-case), ',' and '.'.
// 1 <= numRows <= 1000
func convert(s string, numRows int) string {
	if numRows == 1 {
		return s
	}

	sLen := len(s)
	// 길이가 같다면, 똑같다.
	if sLen <= numRows {
		return s
	}

	// 지그재그 방식으로, numRows 수만큼 row가 그려져야 한다
	//
	// "PAYPALISHIRING"(14) 문자열은 아래 번호 순서처럼 지그재그로 위치하고,
	// 결과는 각 로우를 합쳐서 한 줄로 만들어 보자
	//
	// (1) P       (5) A       H   N
	// (2) A (4) P (6) L (8) S I I G
	// (3) Y       (7) I       R
	sSlice := strings.Split(s, "")
	zigzag := make([][]string, numRows) // numRows 크기의 2차원 슬라이스 생성
	// 내부 슬라이스 초기화
	for i := range zigzag {
		zigzag[i] = make([]string, sLen)
		// for j := range zigzag[i] {
		// 	zigzag[i][j] = "0"
		// }
	}

	var rowIdx, colIdx int
	downward := true
	for _, char := range sSlice {
		zigzag[rowIdx][colIdx] = char

		// moving: 아래로 끝까지 이동 -> 대각선으로 위로 이동 -> 다시 아래로 끝까지 이동 -> 반복
		if downward {
			// 내려가는중
			if numRows > rowIdx && numRows != rowIdx+1 { // 다음으로 갈 수 있어야 한다
				rowIdx++
			} else { // 끝에 도달
				downward = false
				// 우측으로 이동하고
				colIdx++
				// 다시 위로 올라간다
				rowIdx--
			}
		} else {
			// 올라가는중
			if rowIdx != 0 {
				// 대각선 이동
				rowIdx--
				colIdx++
			} else { // 0번 행에 돌아옴
				downward = true
				// 다시 내려간다
				rowIdx++
			}
		}
	}

	// for _, tmp := range zigzag {
	// 	fmt.Println(tmp)
	// }

	result := ""
	for _, row := range zigzag {
		for _, element := range row {
			if element != "" {
				result += element
			}
		}
	}

	return result
}

func convert0ms(s string, numRows int) string {
	if numRows == 1 {
		return s
	}

	n := len(s)
	result := make([]byte, n)
	charsInSection := 2 * (numRows - 1)
	index := 0

	for i := 0; i < numRows; i++ {
		for j := i; j < n; j += charsInSection {
			result[index] = s[j]
			index++

			if i != 0 && i != (numRows-1) {
				charsInBetween := charsInSection - (2 * i)
				secondIndex := j + charsInBetween
				if secondIndex < n {
					result[index] = s[secondIndex]
					index++
				}
			}
		}
	}

	return string(result)
}

func convert1ms(s string, numRows int) string {
	if numRows == 1 || len(s) <= numRows {
		return s
	}

	arr := make([]string, numRows)
	check := true
	count := 0

	for _, ch := range s {
		arr[count] += string(ch)

		if check {
			count++
		} else {
			count--
		}

		if count == numRows-1 || count == 0 {
			check = !check
		}
	}

	return strings.Join(arr, "")
}
