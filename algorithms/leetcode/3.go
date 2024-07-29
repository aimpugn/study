package leetcode

func lengthOfLongestSubstring(s string) int {
	// 반복되지 않으면서 가장 긴 부분 문자열을 찾는다.
	// - 부분 문자열이 반복되는지 여부 체크해야 하고
	// - 그 문자열의 길이를 체크해서 swap
	// 해야 할 거 같은데...
	if len(s) == 0 {
		return 0
	}
	if len(s) == 1 {
		return 1
	}

	longestLength := 0
	from := 0
	for i := from; i < len(s); i++ {
		newLongestLength := checkSubStringRange(s, i)
		if newLongestLength > longestLength {
			longestLength = newLongestLength
		}
	}

	return longestLength
}

func checkSubStringRange(s string, from int) int {
	longestSubString := string(s[from])

OuterLoop:
	for i := from + 1; i < len(s); i++ {
		newR := rune(s[i])
		for _, r := range longestSubString {
			if r == newR {
				break OuterLoop
			}
		}
		longestSubString += string(newR)
	}

	return len(longestSubString)
}

func lengthOfLongestSubstring1ms(s string) int {
	max := func(x, y int) int {
		if x < y {
			return y
		}
		return x
	}

	// runeAndLastIdx 해시맵은 각 문자(rune)가 마지막으로 등장한 인덱스를 저장한다
	runeAndLastIdx := make(map[rune]int)
	// startIdxOfSubString은 현재 고려 중인 부분 문자열의 시작 인덱스를 나타낸다.
	startIdxOfSubString := 0
	// length는 가장 긴 중복되지 않는 부분 문자열의 길이
	length := 0

	// 문자열을 한번만 순회.
	// 슬라이딩 윈도우(현재 고려 중인 부분 문자열)와 해시맵(각 문자의 마지막 위치 기록) 사용.
	//
	// 각 문자가 마지막으로 등장한 위치를 기록하면서,
	// 중복 문자가 발견될 때 현재 윈도우(부분 문자열)의 시작 위치를 조정
	//
	// Example:
	// 	startIdxOfSubString, length := 0
	// 	abcabcd
	// 	^ lastIdx = 0
	// 	   ^ idx = 3 `a` 문자가 중복됨
	//  length = max(0, 3 - 0) = 3
	// 	startIdxOfSubString = 1
	// 	bcabcd // 새로운 부분 문자열 시작
	//
	// 여러 개의 인덱스가 사용된다
	//  - idx: 현재 rune의 인덱스
	//  - lastIdx: 중복된 rune의 마지막 인덱스
	//  - startIdxOfSubString: 부분 문자열의 시작 인덱스
	for idx, r := range s {
		// 만약 현재 문자가 이미 등장했고, 그 위치가 현재 부분 문자열의 시작 이후라면 중복된 문자임을 의미
		if lastIdx, ok := runeAndLastIdx[r]; ok && startIdxOfSubString <= lastIdx {
			// 이전에 발견된 가장 긴 길이(`length`)와
			// 현재 부분 문자열의 길이(`idx-startIdxOfSubString`)를 비교하여 업데이트
			length = max(length, idx-startIdxOfSubString)
			// 중복된 문자를 포함하지 않는 새로운 부분 문자열의 시작 위치를 조정한다.
			startIdxOfSubString = lastIdx + 1
		}
		runeAndLastIdx[r] = idx // 같은 rune이 있어도 idx 값은 계속 증가
	}
	// 이전에 발견된 가장 긴 길이(`length`)보다
	// 마지막으로 고려 중인 부분 문자열의 길이(`len(s)-startIdxOfSubString`)가 긴지 확인
	return max(length, len(s)-startIdxOfSubString)
}
