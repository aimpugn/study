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

// TestStackReversesAndFindsPathOfUnknownLines 함수는 스택이 안쪽부터의 프레임 장소로 뒤집히고, 줄 모름 칸이 같은 프로세스에서
// 그 함수가 나온 경로의 함수 아래에 놓이며, 원본 꼴로 되돌아오는지 봅니다. 경로가 둘이면 오류여야 합니다.
func TestStackReversesAndFindsPathOfUnknownLines(t *testing.T) {
	// Given: 프로세스 하나와, 그 프로세스의 다른 스택에서 f가 a.c에 있다고 적힌 장부
	l := NewLoader(atlas.NewBuilder("r", "t", "s", "n", atlas.Origin{Label: "0"}), "v1")
	proc := l.Process("7", "rec", "$.pid")
	l.NotePath(proc, "f", "a.c:3")
	raw := []any{map[string]any{"fn": "main", "at": "m.c:1"}, map[string]any{"fn": "f", "at": UnknownAt}}

	// When: 바깥 함수부터 적힌 스택을 옮기면
	stack, err := l.Stack(proc, raw, "rec", "$.events[]", "$.events[0]")
	if err != nil {
		t.Fatal(err)
	}

	// Then: 0번은 안쪽의 f이고 위치는 없지만 그 부모 함수 장소는 a.c에 있으며, 되돌리면 원본과 같다
	r := l.B.Run()
	inner := r.Places[stack[0]]
	if len(stack) != 2 || inner.Name != "f" || inner.Loc != nil || r.Places[*inner.Parent].Loc.Path != "a.c" {
		t.Fatalf("스택 %v, 안쪽 %+v", stack, inner)
	}
	back := StackOf(r.Places, stack)
	if len(back) != 2 || back[0].(map[string]any)["at"] != "m.c:1" || back[1].(map[string]any)["at"] != UnknownAt {
		t.Fatalf("되돌림 %v", back)
	}

	// When/Then: 같은 프로세스에서 f가 다른 경로에도 나오면 어느 것인지 가를 수 없어 오류다
	l.NotePath(proc, "f", "b.c:9")
	if _, err := l.Stack(proc, raw, "rec", "$.events[]", "$.events[0]"); err == nil {
		t.Fatal("경로가 둘이면 오류여야 함")
	}
}

// TestLinkIndexOrdersByOtherEventID 함수는 연결 색인의 목록이 연결 표의 차례가 아니라 상대 사건의 id 차례인지 확인합니다.
func TestLinkIndexOrdersByOtherEventID(t *testing.T) {
	// Given: 사건 0에서 3·1·2로 가는 contains 연결이 그 차례로 적힌 run
	r := &atlas.Run{Links: []atlas.Link{
		{ID: 0, Src: 0, Dst: 3, Kind: "contains"}, {ID: 1, Src: 0, Dst: 1, Kind: "contains"},
		{ID: 2, Src: 0, Dst: 2, Kind: "contains"}, {ID: 3, Src: 2, Dst: 1, Kind: "next"},
	}}

	// When: 색인하면
	out, in := LinkIndex(r)

	// Then: 나가는 목록은 도착 id 차례, 들어오는 목록은 출발 id 차례이고 종류마다 나뉜다
	if got := out[0]["contains"]; len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("out[0].contains = %v, 기대 [1 2 3]", got)
	}
	if got := in[1]["contains"]; len(got) != 1 || got[0] != 0 {
		t.Fatalf("in[1].contains = %v", got)
	}
	if got := in[1]["next"]; len(got) != 1 || got[0] != 2 {
		t.Fatalf("in[1].next = %v", got)
	}
}
