package source

import (
	"strings"
	"testing"
)

// sample은 PostgreSQL 관례(머리 주석, 반환형 줄, 이름으로 시작하는 정의 줄)를 따른 작은 C 파일입니다.
var sample = strings.Split(`#include "postgres.h"

/*
 * add_one
 *
 * Return x + 1.
 */
static int
add_one(int x)
{
	int			y;

	y = x + 1;
	return y;
}`, "\n")

// 머리 주석부터 멈춘 줄까지가 짧으면 한 구간으로 이어 보여 주는지 확인합니다.
func TestComputeShortFunctionIsOneRange(t *testing.T) {
	// Given: add_one이 13번 줄(y = x + 1)에 멈춤

	// When
	w := Compute(sample, "add_one", 13)

	// Then: 3번 줄의 "/*"부터 파일 끝(15번 줄)까지 한 구간, 정의 줄은 9번
	if w.DefLine != 9 {
		t.Fatalf("정의 줄 %d, 9여야 합니다", w.DefLine)
	}
	if len(w.Ranges) != 1 || w.Ranges[0] != (Range{From: 3, To: 15}) {
		t.Fatalf("구간 %+v, [3,15] 하나여야 합니다", w.Ranges)
	}
}

// 멈춘 자리가 정의에서 멀면 머리와 멈춘 자리를 따로 보여 주는지 확인합니다.
func TestComputeLongFunctionSplitsHeadAndBody(t *testing.T) {
	// Given: 정의 뒤에 200줄짜리 본문이 있고(파일 210줄) 150번 줄에 멈춤
	lines := append([]string{}, sample[:10]...) // 1~10번 줄: 주석, 반환형, 정의, "{"
	for i := 0; i < 200; i++ {
		lines = append(lines, "\tx++;")
	}
	stop := 150

	// When
	w := Compute(lines, "add_one", stop)

	// Then: 머리 [3,10]과 멈춘 자리 [138,162]
	want := []Range{{From: 3, To: 10}, {From: stop - contextLines, To: stop + contextLines}}
	if len(w.Ranges) != 2 || w.Ranges[0] != want[0] || w.Ranges[1] != want[1] {
		t.Fatalf("구간 %+v, %+v여야 합니다", w.Ranges, want)
	}
}

// 정의 줄을 찾지 못하면 추측하지 않고 멈춘 자리만 보여 주는지 확인합니다.
func TestComputeWithoutDefinitionShowsOnlyBody(t *testing.T) {
	// When: 파일에 없는 이름으로 계산
	w := Compute(sample, "no_such_function", 13)

	// Then
	if w.DefLine != 0 || len(w.Ranges) != 1 || w.Ranges[0] != (Range{From: 1, To: 15}) {
		t.Fatalf("정의 없이 멈춘 자리 주변만 와야 합니다: %+v", w)
	}
}

// 파일 범위 밖의 줄 번호에는 구간을 만들지 않는지 확인합니다.
func TestComputeOutOfRange(t *testing.T) {
	for _, line := range []int{0, -1, len(sample) + 1} {
		if w := Compute(sample, "add_one", line); len(w.Ranges) != 0 {
			t.Fatalf("줄 %d에서 구간이 없어야 합니다: %+v", line, w)
		}
	}
}

// 컨테이너의 다른 파일을 읽으려는 경로를 거부하는지 확인합니다.
func TestValidatePath(t *testing.T) {
	ok := []string{"src/backend/tcop/postgres.c", "src/include/access/tableam.h", "contrib/pageinspect/heapfuncs.c"}
	bad := []string{
		"/etc/passwd", "../etc/passwd", "src/../../etc/passwd", "src//x.c", "src/./x.c",
		"etc/passwd", "src/backend/x.c; rm -rf /", "src/backend/$(id).c", "src/x.c\n", "",
	}
	for _, p := range ok {
		if err := ValidatePath(p); err != nil {
			t.Errorf("받아야 할 경로를 거부했습니다: %q: %v", p, err)
		}
	}
	for _, p := range bad {
		if err := ValidatePath(p); err == nil {
			t.Errorf("거부해야 할 경로를 받았습니다: %q", p)
		}
	}
}
