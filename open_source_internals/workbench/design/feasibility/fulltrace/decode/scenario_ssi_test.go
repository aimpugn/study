package main

import (
	"reflect"
	"testing"
)

// TestCallsWithinStopsAtTheCallsOwnExit 는 입구 기록 하나의 호출 안에서만(안긴 호출 포함) 찾고, 그 호출이 끝난 뒤의 같은
// 함수는 세지 않는지 봅니다. 실험 4에서 PredicateLockTID가 표 단위 락 때문에 돌아온 뒤 다음 요청이 CreatePredicateLock을
// 부르면, 앞의 요청이 락을 만든 것으로 잘못 남게 됩니다.
func TestCallsWithinStopsAtTheCallsOwnExit(t *testing.T) {
	// Given: A(0) 안에서 B가 C(만들기)를 부르는 호출과, A(4) 안에서 아무것도 만들지 않고 끝난 뒤 바깥에서 C를 부르는 호출
	tr := &traceFile{Records: []record{
		{Kind: kindEnter, Fn: 0xA, Ns: 10},
		{Kind: kindEnter, Fn: 0xB, Ns: 11},
		{Kind: kindEnter, Fn: 0xC, Ns: 12},
		{Kind: kindExit, Fn: 0xC, Ns: 13},
		{Kind: kindExit, Fn: 0xB, Ns: 14},
		{Kind: kindExit, Fn: 0xA, Ns: 15},
		{Kind: kindEnter, Fn: 0xA, Ns: 20},
		{Kind: kindEnter, Fn: 0xB, Ns: 21},
		{Kind: kindExit, Fn: 0xB, Ns: 22},
		{Kind: kindExit, Fn: 0xA, Ns: 23},
		{Kind: kindEnter, Fn: 0xC, Ns: 30},
	}}
	isC := func(fn uint64) bool { return fn == 0xC }

	// When, Then
	if !callsWithin(tr, 0, 100, isC) {
		t.Fatalf("첫 호출은 안긴 B 아래에서 C를 불렀습니다")
	}
	if callsWithin(tr, 6, 100, isC) {
		t.Fatalf("둘째 호출은 끝난 뒤에야 C가 불렸습니다")
	}
}

// TestCallsWithinStopsAtTheRequestEnd 는 오류로 출구 기록이 빠진 호출이 요청의 끝을 넘어 다음 요청의 함수를 세지 않는지 봅니다.
func TestCallsWithinStopsAtTheRequestEnd(t *testing.T) {
	// Given: 출구가 없는 A와, 요청 끝(50) 뒤에 불린 C
	tr := &traceFile{Records: []record{
		{Kind: kindEnter, Fn: 0xA, Ns: 10},
		{Kind: kindEnter, Fn: 0xC, Ns: 60},
	}}

	// When, Then
	if callsWithin(tr, 0, 50, func(fn uint64) bool { return fn == 0xC }) {
		t.Fatalf("요청 끝 뒤의 C를 셌습니다")
	}
}

// TestPredTargetReadsGranularityFromTag 는 술어 락 대상 태그의 블록·줄로 표·페이지·행을 가르는지 봅니다
// (predicate_internals.h 388~404행의 SET_PREDICATELOCKTARGETTAG_* 매크로).
func TestPredTargetReadsGranularityFromTag(t *testing.T) {
	cases := []struct {
		blk, off string
		want     map[string]string
	}{
		{"4294967295", "0", map[string]string{"gran": "relation"}},
		{"1", "0", map[string]string{"gran": "page", "blk": "1"}},
		{"0", "2", map[string]string{"gran": "tuple", "ctid": "(0,2)"}},
	}
	for _, c := range cases {
		// When
		got := predTarget(c.blk, c.off)

		// Then
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("predTarget(%s, %s) = %v; want %v", c.blk, c.off, got, c.want)
		}
	}
}

// TestSxactFlagNamesReadsDecimalBits 는 uprobe가 10진수로 남긴 flags를 비트 이름으로 바꾸는지 봅니다. 16진수로 읽으면
// 16(CONFLICT_OUT)이 0x16(PREPARED|ROLLED_BACK|CONFLICT_OUT)으로 잘못 풀립니다. 읽지 못한 값은 그대로 둡니다.
func TestSxactFlagNamesReadsDecimalBits(t *testing.T) {
	cases := map[string]string{"0": "", "8": "DOOMED", "16": "CONFLICT_OUT", "9": "COMMITTED|DOOMED", "(fault)": "(fault)"}
	for in, want := range cases {
		if got := sxactFlagNames(in); got != want {
			t.Fatalf("sxactFlagNames(%q) = %q; want %q", in, got, want)
		}
	}
}

// TestAttrValueReadsBooleanAsPsqlDoes 는 boolean 열 1바이트를 psql이 보이는 t·f로 푸는지 봅니다. 실험 4의 on_call이 처음
// boolean 열이었고, 풀지 않으면 페이지 그림에 \x01·\x00이 그대로 남습니다. 길이가 맞지 않는 값은 그대로 둡니다.
func TestAttrValueReadsBooleanAsPsqlDoes(t *testing.T) {
	cases := map[string]string{`\x01`: "t", `\x00`: "f", `\x0100`: `\x0100`, "NULL": "NULL"}
	for in, want := range cases {
		// When
		got := attrValue(in, "boolean")

		// Then
		if got != want {
			t.Fatalf("attrValue(%q, boolean) = %q; want %q", in, got, want)
		}
	}
}
