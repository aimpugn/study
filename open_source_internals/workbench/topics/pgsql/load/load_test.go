package load

import (
	"testing"

	"study/open_source_internals/workbench/internal/atlas"
)

// TestParseAtReadsPathAndPositiveLine 함수는 위치 글의 마지막 콜론 뒤만 줄로 읽고 틀린 위치를 거르는지 봅니다.
func TestParseAtReadsPathAndPositiveLine(t *testing.T) {
	// Given: 경로에 콜론이 있는 위치와 틀린 위치들(줄을 모르는 칸의 UnknownAt 포함)
	path, line, err := ParseAt("src/a:b.c:12")

	// When/Then: 마지막 콜론 뒤가 줄이고, 줄이 없거나 양수가 아니면 오류다
	if err != nil || path != "src/a:b.c" || line != 12 {
		t.Fatalf("ParseAt = %q %d %v", path, line, err)
	}
	for _, bad := range []string{"", "a.c", ":3", "a.c:0", "a.c:x", UnknownAt} {
		if _, _, err := ParseAt(bad); err == nil {
			t.Fatalf("%q는 오류여야 함", bad)
		}
	}
}

// TestLoaderReusesPlaceAndKeepsFirstSource 함수는 같은 내용의 장소를 다시 찾으면 처음 만든 장소와 출처를 돌려주는지 봅니다.
func TestLoaderReusesPlaceAndKeepsFirstSource(t *testing.T) {
	// Given: 같은 프로세스 아래 같은 코드 자리를 두 출처가 차례로 찾는 장부
	l := NewLoader(atlas.NewBuilder("r", "t", "s", "n", atlas.Origin{Label: "0"}), "v1")
	proc := l.Process("7", "rec", "$.pid")

	// When: 관찰 출처가 먼저 만들고 추정 출처가 같은 자리를 찾는다
	first, err := l.Site(proc, "a.c:5", "rec", "$.path[].at")
	if err != nil {
		t.Fatal(err)
	}
	again, err := l.Site(proc, "a.c:5", "guess", "$.hint.at")
	if err != nil {
		t.Fatal(err)
	}

	// Then: 장소는 하나이고 처음 출처와 처음 경로 무리가 남는다
	r := l.B.Run()
	if first != again || len(r.Places) != 2 {
		t.Fatalf("장소 %d·%d, 수 %d", first, again, len(r.Places))
	}
	if r.Places[first].Source != "rec" || l.Origins[atlas.RecordRef{Table: "place", ID: first}] != "$.path[].at" {
		t.Fatalf("처음 출처가 남아야 함: %+v", r.Places[first])
	}
}
