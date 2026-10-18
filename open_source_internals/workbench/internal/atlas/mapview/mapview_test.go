package mapview

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"study/open_source_internals/workbench/internal/atlas"
)

// update 깃발을 주면 대조표(testdata/ranges.json)를 다시 씁니다: go test ./internal/atlas/mapview -run Fixture -update
var update = flag.Bool("update", false, "범위 대조표를 다시 씀")

// toySpec 함수는 시험용 가짜 주제의 자료 파일입니다. 주제 이름도 일부러 플랫폼이 모르는 이름(toy)을 씁니다.
//
// 띠는 위에서부터 front(클라이언트), back(프로세스 아래, 대체 규칙), mem(이름이 touch인 사건), ground(지면), kern(커널 요약), 커널 층 둘(kf1·kf2), disk(파일)입니다.
func toySpec() *Spec {
	return &Spec{
		Title: "toy", Default: "toy/s1/a",
		Zones: []Zone{{ID: "cl", Name: "클라"}, {ID: "sv", Name: "서버"}, {ID: "bd", Name: "경계"}, {ID: "kn", Name: "커널"}, {ID: "dk", Name: "디스크"}},
		Bands: []Band{
			{ID: "front", Name: "앞", Zoom: ZoomProcess, Zone: "cl", Match: []Rule{{Place: &PlaceMatch{Kinds: []string{"client"}}}}},
			{ID: "back", Name: "뒤", Zoom: ZoomProcess, Zone: "sv", Match: []Rule{{Fallback: true, Ancestor: &PlaceMatch{Kinds: []string{"process"}}}}},
			{ID: "mem", Name: "메모리", Zoom: ZoomProcess, Zone: "sv", Match: []Rule{{Names: []string{"touch"}}}},
			{ID: "ground", Name: "경계", Zoom: ZoomProcess, Zone: "bd", Ground: true, Match: []Rule{{Sources: []string{"sys"}}}},
			{ID: "kern", Name: "커널", Zoom: ZoomProcess, Zone: "kn", Match: []Rule{{Fallback: true, Sources: []string{"kt"}}}},
			{ID: "kf1", Name: "층1", Zoom: ZoomFunction, Zone: "kn", Under: "kern", Match: []Rule{{Sources: []string{"kt"}, Attr: "kt.layer", Values: []string{"L1"}}, {Sources: []string{"kt"}, Attr: "kguess.layer", Values: []string{"L1"}}}},
			{ID: "kf2", Name: "층2", Zoom: ZoomFunction, Zone: "kn", Under: "kern", Match: []Rule{{Sources: []string{"kt"}, Attr: "kt.layer", Values: []string{"L2"}}}},
			{ID: "disk", Name: "파일", Zoom: ZoomProcess, Zone: "dk", Match: []Rule{{Place: &PlaceMatch{Kinds: []string{"file"}}}}},
		},
		Sleep: []string{"sleepfn"},
		Scenes: []SceneSpec{{
			ID: "s", Label: "장면", Runs: []string{"toy/"}, Home: "back",
			Root:     &atlas.Selector{Sources: []string{"dec"}, Kinds: []string{atlas.Span}},
			Steps:    atlas.Selector{Sources: []string{"rec"}, Require: []string{"dec.n"}},
			Expand:   []Expand{{Link: "syscall"}, {Link: "kernel"}},
			Fold:     &Fold{Attr: "dec.f", Label: "접힘"},
			Columns:  &Columns{Attr: "dec.col", Starts: atlas.Selector{Sources: []string{"dec"}, Kinds: []string{atlas.Instant}}},
			Lens:     Lens{MinShare: 0.1},
			Identity: Identity{RunField: "scenario", RunAttr: "rec.q"},
			Name:     Identity{RunAttr: "rec.q"},
		}},
	}
}

// toyRun 함수는 가짜 주제의 run 하나를 만듭니다. kernel이 거짓이면 시스템 콜 아래 커널 나무를 두지 않습니다.
//
// 칸이 될 사건(rec 출처, dec.n 있음)은 시각 차례로 send(0, 열 A), touch(10, 열 B, 접기), touch(20, 열 B, 접기), work(25, 열 B, 끊음), touch(30, 열 B, 접기),
// io(40, 열 C, 파일에서 읽어 메모리로, 시스템 콜 → 커널 나무)입니다. 시스템 콜·커널 짝은 추정 근거(guess)이고, 시각 없는 rec 사건 하나는 칸이 되지 않습니다.
func toyRun(name string, kernel bool, slow int64) *atlas.Run {
	b := atlas.NewBuilder("toy/"+name, "toy", "toy", name, atlas.Origin{Clock: "c", Label: "0"})
	b.AddClock(atlas.Clock{ID: "c", PrecisionNS: 1})
	b.AddSource(atlas.Source{ID: "rec", Tool: "기록기", Clock: "c", Nature: atlas.Observed})
	b.AddSource(atlas.Source{ID: "dec", Tool: "해석기", Clock: "c", Nature: atlas.Derived})
	b.AddSource(atlas.Source{ID: "sys", Tool: "시스템 콜", Nature: atlas.Observed})
	b.AddSource(atlas.Source{ID: "kt", Tool: "커널 추적", Nature: atlas.Observed})
	b.AddSource(atlas.Source{ID: "guess", Tool: "짝 추정", Nature: atlas.Inferred})
	b.AddSource(atlas.Source{ID: "kguess", Tool: "층 추정", Nature: atlas.Inferred})
	b.SetRunAttrs("rec", "q", "SELECT 1")
	cl := b.Place(atlas.PlaceSpec{Kind: "client", Name: "cli", Source: "rec"})
	proc := b.Place(atlas.PlaceSpec{Kind: "process", Name: "42", Source: "rec"})
	fn := b.Place(atlas.PlaceSpec{Parent: &proc, Kind: "function", Name: "fn", Source: "rec", Static: true})
	file := b.Place(atlas.PlaceSpec{Kind: "file", Name: "f", Source: "rec"})
	kp := b.Place(atlas.PlaceSpec{Kind: "process", Name: "k", Source: "kt"})
	b.AddCoverage(atlas.Coverage{Source: "rec", What: "client", Places: []int{cl}, Completeness: "complete", Basis: "파일"})
	ns := func(v int64) *int64 { return atlas.Ptr(v) }
	b.AddEvent(atlas.Event{Place: proc, Kind: atlas.Span, Name: "req", StartNS: ns(0), EndNS: ns(60 + slow), Source: "dec"})
	for _, c := range []struct {
		name string
		t    int64
	}{{"A", 0}, {"B", 5}, {"C", 35}} {
		b.AddEvent(atlas.Event{Place: fn, Kind: atlas.Instant, Name: c.name, StartNS: ns(c.t), EndNS: ns(c.t), Source: "dec"})
	}
	step := func(n int, place int, from, target *int, name string, t int64, col string, fold bool) int {
		id := b.AddEvent(atlas.Event{Place: place, From: from, Target: target, Kind: atlas.Instant, Name: name, StartNS: ns(t), EndNS: ns(t), Source: "rec"})
		b.EventAttrs(id, "dec", "n", json.Number(fmt.Sprint(n)))
		b.EventAttrs(id, "dec", "col", col)
		if fold {
			b.EventAttrs(id, "dec", "f", true)
		}
		return id
	}
	step(1, cl, nil, nil, "send", 0, "A", false)
	step(2, fn, nil, nil, "touch", 10, "B", true)
	step(3, fn, nil, nil, "touch", 20, "B", true)
	step(4, fn, nil, nil, "work", 25, "B", false)
	step(5, fn, nil, nil, "touch", 30, "B", true)
	io := step(6, fn, &file, &proc, "io", 40, "C", false)
	b.AddEvent(atlas.Event{Place: fn, Kind: atlas.Instant, Name: "untimed", Source: "rec"})
	sys := b.AddEvent(atlas.Event{Place: proc, Kind: atlas.Span, Name: "syscall", DurNS: ns(15 + slow), Source: "sys"})
	b.AddLink(io, sys, "syscall", "guess")
	if kernel {
		root := b.AddEvent(atlas.Event{Place: b.Place(atlas.PlaceSpec{Parent: &kp, Kind: "function", Name: "entry", Source: "kt"}), Kind: atlas.Span, Name: "entry", DurNS: ns(14 + slow), Source: "kt"})
		b.EventAttrs(root, "kt", "layer", "L1")
		mid := b.AddEvent(atlas.Event{Place: b.Place(atlas.PlaceSpec{Parent: &kp, Kind: "function", Name: "mid", Source: "kt"}), Kind: atlas.Span, Name: "mid", DurNS: ns(13 + slow), Parent: &root, Source: "kt"})
		b.EventAttrs(mid, "kguess", "layer", "L1")
		sl := b.AddEvent(atlas.Event{Place: b.Place(atlas.PlaceSpec{Parent: &kp, Kind: "function", Name: "sleepfn", Source: "kt"}), Kind: atlas.Span, Name: "sleepfn", DurNS: ns(10 + slow), Parent: &mid, Source: "kt"})
		b.EventAttrs(sl, "kt", "layer", "L2")
		tiny := b.AddEvent(atlas.Event{Place: b.Place(atlas.PlaceSpec{Parent: &kp, Kind: "function", Name: "tiny", Source: "kt"}), Kind: atlas.Span, Name: "tiny", DurNS: ns(1), Parent: &mid, Source: "kt"})
		b.EventAttrs(tiny, "kt", "layer", "L2")
		if slow > 0 {
			// 느린 판은 mid 아래에 다른 자식이 하나 더 있어 자식 차례가 갈립니다.
			extra := b.AddEvent(atlas.Event{Place: b.Place(atlas.PlaceSpec{Parent: &kp, Kind: "function", Name: "extra", Source: "kt"}), Kind: atlas.Span, Name: "extra", DurNS: ns(1), Parent: &mid, Source: "kt"})
			b.EventAttrs(extra, "kt", "layer", "L2")
		}
		b.AddLink(sys, root, "kernel", "guess")
	}
	return b.Run()
}

// toyView 함수는 가짜 주제의 run 셋(a: 커널 기록, b: 같은 요청을 느리게, c: 커널 기록 없음)으로 보기 모형을 만듭니다.
func toyView(t *testing.T) *View {
	t.Helper()
	st := MemStore{"toy/s1/a": toyRun("s1/a", true, 0), "toy/s1/b": toyRun("s1/b", true, 1000), "toy/s1/c": toyRun("s1/c", false, 0)}
	v, err := Build(st, toySpec(), ".")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// TestStepsFoldWithinColumnAndBreakOnOthers 함수는 접기 규칙(고친 규칙 3)을 봅니다.
//
// Given 열 B 안에서 touch·touch·work·touch(접기 조건은 touch만)인 경로에서, When 칸을 만들면,
// Then touch 둘은 「×2」 한 칸, work는 그 사이를 끊어 한 칸, 마지막 touch는 혼자이고(N = 줄 수), 열의 머문 시간은 다음 열의 시작까지입니다.
func TestStepsFoldWithinColumnAndBreakOnOthers(t *testing.T) {
	r := toyView(t).Routes[0]
	var got []string
	for _, s := range r.Steps {
		got = append(got, fmt.Sprintf("%s/%s/%d", s.Column, s.Name, s.Fold))
	}
	if want := "[A/send/0 B/접힘/2 B/work/0 B/touch/0 C/io/0]"; fmt.Sprint(got) != want {
		t.Errorf("칸 %v, 기대 %s", got, want)
	}
	cols := fmt.Sprint(r.Columns)
	if want := "[{A 0 5 [0 0] derived} {B 5 30 [1 3] derived} {C 35 5 [4 4] derived}]"; cols != want {
		t.Errorf("열 %s, 기대 %s", cols, want)
	}
}

// TestFoldStopsAtColumn 함수는 접기가 열(처리 단계) 경계를 넘지 않는지 봅니다(고친 규칙 3).
//
// Given 열 B의 마지막 칸(touch)과 열 C의 첫 칸(io)이 모두 접기 조건에 맞을 때, When 칸을 만들면, Then 둘은 다른 칸입니다.
func TestFoldStopsAtColumn(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	for i := range r.Events {
		if r.Events[i].Name == "io" {
			r.Events[i].Attrs["dec"]["f"] = true
		}
	}
	v, err := Build(MemStore{"toy/s1/a": r}, toySpec(), ".")
	if err != nil {
		t.Fatal(err)
	}
	steps := v.Routes[0].Steps
	if n := len(steps); n != 5 || steps[n-1].Column != "C" || steps[n-2].Column != "B" {
		t.Errorf("칸 %d개, 끝 둘의 열 %s·%s", len(steps), steps[len(steps)-2].Column, steps[len(steps)-1].Column)
	}
}

// TestUntimedEventsAreNotSteps 함수는 시각 없는 사건이 칸이 되지 않는지(고친 규칙 2) 봅니다. 뿌리 규칙이 run 전체를 고르는 장면이라 목록에도 들지 않습니다.
func TestUntimedEventsAreNotSteps(t *testing.T) {
	for _, s := range toyView(t).Routes[0].Steps {
		if s.Name == "untimed" {
			t.Fatal("시각 없는 사건이 칸이 됨")
		}
	}
}

// TestStepsOnTwoClocksAreRejected 함수는 칸의 차례를 정하는 사건이 두 시계에 걸치면 보기 모형을 만들지 않는지 봅니다(고친 규칙 2).
func TestStepsOnTwoClocksAreRejected(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	r.Clocks = append(r.Clocks, atlas.Clock{ID: "other", PrecisionNS: 1})
	r.Sources = append(r.Sources, atlas.Source{ID: "rec2", Tool: "다른 기록기", Clock: "other", Nature: atlas.Observed})
	for i := range r.Events {
		if r.Events[i].Name == "work" {
			r.Events[i].Source = "rec2"
			r.Events[i].Attrs["rec2"] = r.Events[i].Attrs["dec"]
		}
	}
	spec := toySpec()
	spec.Scenes[0].Steps.Sources = []string{"rec", "rec2"}
	if _, err := Build(MemStore{"toy/s1/a": r}, spec, "."); err == nil || !strings.Contains(err.Error(), "두 시계") {
		t.Fatalf("오류 %v, 기대 두 시계", err)
	}
}

// TestDescentAndHatch 함수는 하강과 빗금 판정(고친 규칙 4·5)을 봅니다.
//
// Given 파일에서 읽어 메모리로 가는 io 칸(도착은 지면 위)에 시스템 콜이 이어진 경로에서,
// When 커널 나무가 있는 기록(a)과 없는 기록(c)을 보면,
// Then 둘 다 하강이고 a는 기록됨·c는 안 봄이며, 하강의 성격은 추정 짝이라 inferred이고, c의 커널 띠는 경로 전체에서 안 봄입니다.
func TestDescentAndHatch(t *testing.T) {
	v := toyView(t)
	byRun := map[string]Route{}
	for _, r := range v.Routes {
		byRun[r.Run] = r
	}
	a, c := byRun["toy/s1/a"], byRun["toy/s1/c"]
	io := func(r Route) Step { return r.Steps[len(r.Steps)-1] }
	if io(a).Band != "mem" && io(a).Band != "back" {
		t.Logf("io 칸의 띠 %s", io(a).Band)
	}
	if d := io(a).Descent; d == nil || d.State != StateEvents || d.Nature != atlas.Inferred {
		t.Errorf("a의 하강 %+v, 기대 events·inferred", d)
	}
	if d := io(c).Descent; d == nil || d.State != StateUnseen {
		t.Errorf("c의 하강 %+v, 기대 unseen", d)
	}
	if a.Bands["kern"] != StateEvents || c.Bands["kern"] != StateUnseen || c.Bands["kf1"] != StateUnseen {
		t.Errorf("커널 띠 a %s, c %s·%s", a.Bands["kern"], c.Bands["kern"], c.Bands["kf1"])
	}
	// 앞 띠는 a에서 사건이 있고, 디스크 띠는 파일이 출발이라 지난 띠로 셉니다.
	if a.Bands["front"] != StateEvents {
		t.Errorf("앞 띠 %s", a.Bands["front"])
	}
}

// TestWatchedNeedsCompleteObservedCoverage 함수는 띠 상태 넷(고친 규칙 5)을 반례로 봅니다.
//
// Given 앞 띠(클라이언트 장소)를 관찰 출처가 complete로 지켜본 run에서,
// When 보이는 범위가 io 칸(시각 40) 하나일 때, Then 앞 띠는 그 시간대에 사건이 없어 봤고 0번이고, 관찰 범위가 없는 층2는 안 봄입니다.
// 관찰 범위를 partial로 바꾸거나 그 출처를 추정으로 바꾸면 봤고 0번이 아니고, 보이는 범위가 요청 구간(0~60)이면 send(0)가 그 시간대에 있어 흐림입니다.
func TestWatchedNeedsCompleteObservedCoverage(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	ids := map[string]int{}
	for _, e := range r.Events {
		ids[e.Name] = e.ID
	}
	io := []Node{{Event: ids["io"], Band: "mem"}}
	states := func() map[string]string { return bandStates(newIndex(toySpec(), r), io, "c") }
	if got := states(); got["front"] != StateWatched || got["kf2"] != StateUnseen {
		t.Errorf("앞 띠 %s(기대 watched), 층2 %s(기대 unseen)", got["front"], got["kf2"])
	}
	r.Coverage[0].Completeness = "partial"
	if s := states()["front"]; s == StateWatched {
		t.Error("partial 관찰 범위가 봤고 0번이 됨")
	}
	r.Coverage[0].Completeness = "complete"
	r.Sources[0].Nature = atlas.Inferred
	if s := states()["front"]; s == StateWatched {
		t.Error("추정 출처의 관찰 범위가 봤고 0번이 됨")
	}
	r.Sources[0].Nature = atlas.Observed
	if s := bandStates(newIndex(toySpec(), r), []Node{{Event: ids["req"], Band: "back"}}, "c")["front"]; s != StateOther {
		t.Errorf("같은 시간대에 경로 밖 사건이 있는 앞 띠 %s, 기대 other", s)
	}
}

// TestBandRulesPrecedenceAndInferredBand 함수는 띠 배정(줌 배정)의 차례와 추정 표시(고친 규칙 6)를 봅니다.
//
// Given touch(프로세스 아래, 이름 규칙)와 work(프로세스 아래, 대체 규칙만 맞음), 커널 층을 추정 열쇠로 받은 mid가 있을 때,
// When 띠를 정하면, Then touch는 메모리(대체 규칙보다 먼저), work는 뒤, mid는 층1이고 그 마디의 성격은 띠를 정한 열쇠 때문에 inferred입니다.
func TestBandRulesPrecedenceAndInferredBand(t *testing.T) {
	r := toyView(t).Routes[0]
	bands := map[string]string{}
	for _, s := range r.Steps {
		bands[s.Name] = s.Band
	}
	if bands["work"] != "back" || bands["touch"] != "mem" || bands["접힘"] != "mem" {
		t.Errorf("띠 %v", bands)
	}
	for _, n := range r.Steps[len(r.Steps)-1].Tree {
		switch n.Name {
		case "mid":
			if n.Band != "kf1" || n.Nature != atlas.Inferred {
				t.Errorf("mid 띠 %s 성격 %s, 기대 kf1·inferred", n.Band, n.Nature)
			}
		case "sleepfn":
			if n.Band != "kf2" || !n.Sleep {
				t.Errorf("sleepfn 띠 %s 잠듦 %v", n.Band, n.Sleep)
			}
		}
	}
}

// TestLensAndLeaves 함수는 렌즈(고친 규칙 7)와 잎 집합(고친 규칙 8)을 봅니다.
//
// Given 커널 나무가 entry → mid → (sleepfn, tiny)이고 몫 문턱이 10%일 때, When io 칸을 펼치면,
// Then tiny는 숨고 mid에 +1로 접히며, 잎 집합은 렌즈와 무관하게 잎(sleepfn, tiny)의 사건입니다.
func TestLensAndLeaves(t *testing.T) {
	st := toyView(t).Routes[0].Steps[4]
	vis := map[string]bool{}
	folded := map[string]int{}
	leafNames := map[int]string{}
	for _, n := range st.Tree {
		vis[n.Name], folded[n.Name] = n.Visible, n.Folded
		leafNames[n.Event] = n.Name
	}
	if vis["tiny"] || !vis["sleepfn"] || !vis["mid"] || folded["mid"] != 1 {
		t.Errorf("보임 %v 접힘 %v", vis, folded)
	}
	var leaves []string
	for _, id := range st.Leaves {
		leaves = append(leaves, leafNames[id])
	}
	if fmt.Sprint(leaves) != "[sleepfn tiny]" {
		t.Errorf("잎 %v, 기대 [sleepfn tiny]", leaves)
	}
}

// TestCandidatesAndDivergence 함수는 후보 무리(고친 규칙 9)와 갈림을 봅니다.
//
// Given 같은 시나리오 s1의 a·b·c(같은 질의 글)에서, When 보기 모형을 만들면,
// Then 셋이 한 무리이고, a와 b는 칸 차례가 같고 커널 나무의 mid 아래 자식 차례가 달라 주된 갈림이 mid(tree)이며,
// a와 c는 c가 시스템 콜 아래를 기록하지 않아 갈림이 아니라 '한쪽만 기록'(unrecorded)입니다.
func TestCandidatesAndDivergence(t *testing.T) {
	v := toyView(t)
	if len(v.Groups) != 1 || len(v.Groups[0].Routes) != 3 {
		t.Fatalf("무리 %+v", v.Groups)
	}
	a := v.Routes[0]
	kinds := map[string]string{}
	for _, d := range a.Divergence {
		kinds[v.Routes[d.Other].Run] = d.Kind + ":" + d.Name
	}
	if kinds["toy/s1/b"] != "tree:mid" {
		t.Errorf("a·b 갈림 %s, 기대 tree:mid", kinds["toy/s1/b"])
	}
	if kinds["toy/s1/c"] != "unrecorded:syscall" {
		t.Errorf("a·c 갈림 %s, 기대 unrecorded:syscall(c는 시스템 콜 아래를 기록하지 않음)", kinds["toy/s1/c"])
	}
	// 질의 글이 같아도 시나리오가 다르면 다른 무리입니다.
	two, err := Build(MemStore{"toy/s1/a": toyRun("s1/a", true, 0), "toy/s2/a": toyRun("s2/a", true, 0)}, toySpec(), ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(two.Groups) != 2 {
		t.Errorf("시나리오가 다른 두 벌의 무리 %d개, 기대 2", len(two.Groups))
	}
	// 시나리오를 정체에서 빼면(run 이름 전체) 셋이 따로 무리가 됩니다.
	spec := toySpec()
	spec.Scenes[0].Identity.RunField = "name"
	st := MemStore{"toy/s1/a": toyRun("s1/a", true, 0), "toy/s1/b": toyRun("s1/b", true, 1000)}
	w, err := Build(st, spec, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Groups) != 2 {
		t.Errorf("run 이름으로 가르면 무리 %d개, 기대 2", len(w.Groups))
	}
}

// TestNormalizeAndRelate 함수는 범위 규칙(「범위 해설의 규칙」 1~5, 고친 규칙 8)을 반례로 봅니다.
func TestNormalizeAndRelate(t *testing.T) {
	cells := []Cell{{0, 9}, {10, 19}, {20, 29}, {30, 39}}
	for _, c := range []struct {
		x0, x1 float64
		lo, hi int
		want   string
	}{
		{12, 12, 0, 3, "1-1 true"},    // 클릭은 칸 하나
		{8, 21, 0, 3, "0-2 true"},     // 일부만 덮어도 바깥쪽으로
		{21, 8, 0, 3, "0-2 true"},     // 거꾸로 드래그
		{8, 21, 1, 3, "1-2 true"},     // 흐린 칸(경로 밖)은 넣지 않음
		{40, 50, 0, 3, "-1--1 false"}, // 아무 칸도 덮지 않음
	} {
		f, l, ok := Normalize(cells, c.x0, c.x1, c.lo, c.hi)
		if got := fmt.Sprintf("%d-%d %v", f, l, ok); got != c.want {
			t.Errorf("Normalize(%v, %v, %d, %d) = %s, 기대 %s", c.x0, c.x1, c.lo, c.hi, got, c.want)
		}
	}
	A := Range{Zoom: "process", Leaves: []int{1, 2, 3}}
	for _, c := range []struct {
		b    Range
		want string
	}{
		{Range{Zoom: "process", Leaves: []int{1, 2, 3}}, RelSame},
		{Range{Zoom: "function", Leaves: []int{1, 2, 3}}, RelSameLeaves},
		{Range{Zoom: "function", Leaves: []int{2}}, RelContains},
		{Range{Zoom: "process", Leaves: []int{1, 2, 3, 4}}, RelInside},
		{Range{Zoom: "process", Leaves: []int{3, 4}}, RelOverlap},
		{Range{Zoom: "process", Leaves: []int{7}}, RelApart},
	} {
		if got := Relate(A, c.b); got != c.want {
			t.Errorf("Relate(A, %v) = %s, 기대 %s", c.b, got, c.want)
		}
	}
}

// fixture 값은 Go 정본의 범위 판정을 화면의 range.js와 견주는 대조표입니다(고친 규칙 8, check-map.mjs가 읽음).
type fixture struct {
	Normalize []normCase `json:"normalize"`
	Relate    []relCase  `json:"relate"`
}

// normCase 값은 칸 맞춤 사례 하나와 Go의 답입니다.
type normCase struct {
	Cells []Cell  `json:"cells"`
	X0    float64 `json:"x0"`
	X1    float64 `json:"x1"`
	Lo    int     `json:"lo"`
	Hi    int     `json:"hi"`
	Want  []int   `json:"want"`
}

// relCase 값은 관계 판정 사례 하나와 Go의 답입니다.
type relCase struct {
	A    Range  `json:"a"`
	B    Range  `json:"b"`
	Want string `json:"want"`
}

// buildFixture 함수는 경계 사례와 씨앗을 고정한 무작위 표본으로 대조표를 만듭니다. 같은 코드는 늘 같은 대조표를 냅니다.
func buildFixture() fixture {
	rng := rand.New(rand.NewSource(20261018))
	var f fixture
	norm := func(cells []Cell, x0, x1 float64, lo, hi int) {
		first, last, ok := Normalize(cells, x0, x1, lo, hi)
		want := []int{}
		if ok {
			want = []int{first, last}
		}
		f.Normalize = append(f.Normalize, normCase{cells, x0, x1, lo, hi, want})
	}
	// 경계 사례: 칸 경계 위, 칸 사이 틈, 흐린 칸 경계, 빈 칸 목록, 거꾸로 드래그.
	cells := []Cell{{0, 9}, {10, 19}, {20, 29}, {30, 39}, {40, 49}}
	for _, c := range [][4]float64{{9, 9, 0, 4}, {9.5, 9.5, 0, 4}, {10, 10, 0, 4}, {0, 49, 1, 3}, {49, 0, 0, 4}, {50, 60, 0, 4}, {19, 20, 2, 2}, {-5, 0, 0, 4}} {
		norm(cells, c[0], c[1], int(c[2]), int(c[3]))
	}
	norm(nil, 0, 10, 0, 3)
	for i := 0; i < 200; i++ {
		n := 1 + rng.Intn(30)
		var cs []Cell
		x := 0.0
		for j := 0; j < n; j++ {
			w := 4 + rng.Float64()*40
			cs = append(cs, Cell{x, x + w - 1})
			x += w
		}
		lo := rng.Intn(n)
		hi := lo + rng.Intn(n-lo)
		norm(cs, rng.Float64()*x, rng.Float64()*x, lo, hi)
	}
	zooms := []string{ZoomSystem, ZoomProcess, ZoomFunction}
	leaves := func() []int {
		m := map[int]bool{}
		for k := rng.Intn(6); k > 0; k-- {
			m[rng.Intn(10)] = true
		}
		return sortedIDs(m)
	}
	for i := 0; i < 300; i++ {
		a := Range{Zoom: zooms[rng.Intn(3)], Leaves: leaves()}
		b := Range{Zoom: zooms[rng.Intn(3)], Leaves: leaves()}
		if i%7 == 0 {
			b.Leaves = append([]int{}, a.Leaves...)
		}
		f.Relate = append(f.Relate, relCase{a, b, Relate(a, b)})
	}
	return f
}

// TestRangeFixtureIsCurrent 함수는 커밋된 대조표가 지금 Go 정본의 답과 같은지 봅니다. 다르면 -update로 다시 쓰고 check-map.mjs를 돌립니다.
func TestRangeFixtureIsCurrent(t *testing.T) {
	got, err := json.MarshalIndent(buildFixture(), "", " ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "ranges.json")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("testdata/ranges.json이 Go 정본의 답과 다름(-update로 다시 쓰고 check-map.mjs로 화면 쪽을 보기)")
	}
}

// TestPlatformDoesNotKnowTopics 함수는 플랫폼과 주제의 경계(고친 규칙 11, T6의 판정)를 봅니다.
//
// 이 패키지와 명령의 의존에 topics/가 없고, 플랫폼 폴더의 파일에 주제 이름 글자가 없으며, 위 시험들이 가짜 주제(toy)만으로 보기 모형을 만듭니다.
func TestPlatformDoesNotKnowTopics(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".", "../../../cmd/atlas-map").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "/topics/") {
		t.Errorf("플랫폼이 주제 패키지를 가져옴:\n%s", out)
	}
	words := []string{"pgsql", "postgres", "spring", "journey", "txlab", "concept-"}
	var files []string
	for _, dir := range []string{".", "screen", "../../../cmd/atlas-map"} {
		m, _ := filepath.Glob(filepath.Join(dir, "*"))
		files = append(files, m...)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		low := strings.ToLower(string(b))
		for _, w := range words {
			if strings.Contains(low, w) {
				t.Errorf("%s에 주제 이름 %q", f, w)
			}
		}
	}
}

// TestWriteHTMLEmbedsViewOnly 함수는 화면 파일이 보기 모형을 한 번 품고 자리 표시를 모두 채우는지, 글 속의 "</"가 스크립트를 닫지 못하는지 봅니다.
func TestWriteHTMLEmbedsViewOnly(t *testing.T) {
	v := toyView(t)
	v.Title = "닫기 시도 </script>"
	var buf bytes.Buffer
	if err := WriteHTML(&buf, v); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, mark := range []string{"/*CSS*/", "/*RANGE*/", "/*MAP*/", "/*DATA*/"} {
		if strings.Contains(page, mark) {
			t.Errorf("자리 표시 %s가 남음", mark)
		}
	}
	if strings.Count(page, "</script>") != 3 {
		t.Errorf("</script> %d개, 기대 3(보기 모형 안의 글이 스크립트를 닫음)", strings.Count(page, "</script>"))
	}
	start := strings.Index(page, `type="application/json">`) + len(`type="application/json">`)
	end := strings.Index(page[start:], "</script>")
	var back View
	if err := json.Unmarshal([]byte(strings.ReplaceAll(page[start:start+end], `<\/`, "</")), &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back.Title, v.Title) || len(back.Routes) != len(v.Routes) {
		t.Error("품은 보기 모형이 원래와 다름")
	}
}
