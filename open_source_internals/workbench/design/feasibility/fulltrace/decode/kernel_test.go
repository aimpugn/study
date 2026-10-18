package main

import (
	"fmt"
	"reflect"
	"testing"
)

// callsAt 는 줄 번호마다 부모 a.c의 그 줄에서 바로 부르는 길 하나씩을 만듭니다.
func callsAt(lines ...int) [][]srcLine {
	var out [][]srcLine
	for _, l := range lines {
		out = append(out, []srcLine{{Fn: "parent", At: fmt.Sprintf("a.c:%d", l)}})
	}
	return out
}

// linesOf 는 호출의 후보 줄 번호들입니다.
func linesOf(k *kcall) []int {
	var out []int
	for _, p := range k.Calls {
		var l int
		fmt.Sscanf(p[0].At, "a.c:%d", &l)
		out = append(out, l)
	}
	return out
}

// TestNarrowKeepsOnlyOrderConsistentLines 는 btrfs_sync_log에서 본 꼴입니다. 같은 이름(mutex_lock)을 여러 줄에서 부르지만,
// 바로 뒤의 형제(write_all_supers)가 한 줄에서만 불리면 그 앞의 후보만 남아야 합니다.
func TestNarrowKeepsOnlyOrderConsistentLines(t *testing.T) {
	// Given: 줄 20의 단일 호출, 후보 10·30·50인 호출, 줄 40의 단일 호출 순서로 불렸습니다.
	first := &kcall{Fn: "a", Calls: callsAt(20)}
	lock := &kcall{Fn: "mutex_lock", Calls: callsAt(10, 30, 50)}
	last := &kcall{Fn: "b", Calls: callsAt(40)}

	// When
	narrow([]*kcall{first, lock, last})

	// Then: 20 ≤ L ≤ 40을 만족하는 30만 남고, 좁히기 전 후보 수 3이 남습니다.
	if got := linesOf(lock); !reflect.DeepEqual(got, []int{30}) || lock.CallsOf != 3 {
		t.Fatalf("lines %v, callsOf %d; want [30], 3", got, lock.CallsOf)
	}
	if first.CallsOf != 0 || last.CallsOf != 0 {
		t.Fatalf("단일 후보는 좁히지 않아야 합니다: %d, %d", first.CallsOf, last.CallsOf)
	}
}

// TestNarrowLeavesAllWhenOrderIsImpossible 는 반복문처럼 위에서 아래로 읽는 전제가 깨진 경우입니다. 아무것도 좁히지 않아야 합니다.
func TestNarrowLeavesAllWhenOrderIsImpossible(t *testing.T) {
	// Given: 줄 40의 호출이 먼저, 줄 20의 호출이 나중에 불렸습니다(반복문으로 다시 올라간 꼴).
	a := &kcall{Fn: "a", Calls: callsAt(40)}
	mid := &kcall{Fn: "m", Calls: callsAt(10, 30)}
	b := &kcall{Fn: "b", Calls: callsAt(20)}

	// When
	narrow([]*kcall{a, mid, b})

	// Then
	if got := linesOf(mid); !reflect.DeepEqual(got, []int{10, 30}) || mid.CallsOf != 0 {
		t.Fatalf("lines %v, callsOf %d; want [10 30], 0", got, mid.CallsOf)
	}
}

// TestNarrowIgnoresSiblingsWithoutCandidates 는 호출 줄을 찾지 못한 형제(함수 포인터 호출 등)가 제약을 주지 않는지 봅니다.
func TestNarrowIgnoresSiblingsWithoutCandidates(t *testing.T) {
	// Given: 후보 없는 형제를 사이에 두고 줄 20, 후보 10·30, 후보 없음, 줄 40 순서입니다.
	a := &kcall{Fn: "a", Calls: callsAt(20)}
	mid := &kcall{Fn: "m", Calls: callsAt(10, 30)}
	none := &kcall{Fn: "ptr"}
	b := &kcall{Fn: "b", Calls: callsAt(40)}

	// When
	narrow([]*kcall{a, mid, none, b})

	// Then
	if got := linesOf(mid); !reflect.DeepEqual(got, []int{30}) {
		t.Fatalf("lines %v; want [30]", got)
	}
}
