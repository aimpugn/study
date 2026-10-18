package mapview

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
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
		Title: "toy", Default: "toy/s1/a", EventNamedKinds: []string{"process", "client"},
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
//
// Given 칸 출처의 사건 가운데 시각 없는 untimed가 낀 가짜 주제 run에서, When 보기 모형을 만들면, Then 어느 칸의 이름도 untimed가 아닙니다.
func TestUntimedEventsAreNotSteps(t *testing.T) {
	for _, s := range toyView(t).Routes[0].Steps {
		if s.Name == "untimed" {
			t.Fatal("시각 없는 사건이 칸이 됨")
		}
	}
}

// TestStepsOnTwoClocksAreRejected 함수는 칸의 차례를 정하는 사건이 두 시계에 걸치면 보기 모형을 만들지 않는지 봅니다(고친 규칙 2).
//
// Given work 사건만 다른 시계(other)의 출처로 옮기고 칸 출처에 두 출처를 다 넣은 장면에서, When 보기 모형을 만들면, Then 「두 시계」 오류로 거절합니다.
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
//
// Given 0~39를 열 칸씩 덮는 칸 넷과 잎 집합 사례들에서, When 드래그 구간을 맞추고 두 범위를 견주면, Then 사례마다 적은 [first, last]와 관계가 나옵니다.
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
	Leaves    []leafCase `json:"leaves"`
}

// leafCase 값은 칸들의 잎 집합 합치기 사례 하나와 Go의 답입니다.
type leafCase struct {
	CellLeaves [][]int `json:"cell_leaves"`
	First      int     `json:"first"`
	Last       int     `json:"last"`
	Want       []int   `json:"want"`
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
	// 보기 모형에서 뽑은 사례: 칸마다의 잎 집합(접힌 칸, 커널 나무를 펼친 칸 포함)을 모든 [first, last]로 합치고,
	// 시스템 칸 하나와 그 칸이 합친 프로세스 안 칸들의 범위(줌만 다르고 잎이 같음), 함수 줌 마디 하나와 그 칸의 범위(안에 듦)를 견줍니다.
	v, err := Build(MemStore{"toy/s1/a": toyRun("s1/a", true, 0)}, toySpec(), ".")
	if err != nil {
		panic(err)
	}
	r := v.Routes[0]
	var cl [][]int
	for _, st := range r.Steps {
		cl = append(cl, st.Leaves)
	}
	for i := range cl {
		for j := i; j < len(cl); j++ {
			f.Leaves = append(f.Leaves, leafCase{cl, i, j, RangeLeaves(cl, i, j)})
		}
	}
	for _, sy := range r.System {
		a := Range{Zoom: ZoomSystem, First: sy.Index, Last: sy.Index, Leaves: sy.Leaves}
		b := Range{Zoom: ZoomProcess, First: sy.Steps[0], Last: sy.Steps[1], Leaves: RangeLeaves(cl, sy.Steps[0], sy.Steps[1])}
		f.Relate = append(f.Relate, relCase{a, b, Relate(a, b)})
	}
	for i, st := range r.Steps {
		for k, n := range st.Tree {
			if len(st.Tree) < 2 || n.Parent < 0 {
				continue
			}
			a := Range{Zoom: ZoomFunction, First: k, Last: k, Leaves: n.Leaves}
			b := Range{Zoom: ZoomProcess, First: i, Last: i, Leaves: st.Leaves}
			f.Relate = append(f.Relate, relCase{a, b, Relate(a, b)}, relCase{b, a, Relate(b, a)})
		}
	}
	return f
}

// TestRangeFixtureIsCurrent 함수는 커밋된 대조표가 지금 Go 정본의 답과 같은지 봅니다. 다르면 -update로 다시 쓰고 check-map.mjs를 돌립니다.
//
// Given 경계 사례와 씨앗 고정 무작위 표본으로 지금 정본이 셈한 대조표에서, When 커밋된 testdata/ranges.json과 견주면, Then 바이트까지 같습니다.
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
// Given 이 패키지와 명령 cmd/atlas-map, 플랫폼 폴더의 시험 아닌 파일에서, When 의존 목록과 파일 글을 보면,
// Then 의존에 topics/가 없고 주제 이름 글자가 없습니다. 위 시험들은 가짜 주제(toy)만으로 보기 모형을 만듭니다.
// 또 저장소의 모든 주제 자료 파일(topics/*/*/map.json)이 쓰는 장소 종류·출처 id·attrs 열쇠·run 앞부분의 낱말이 플랫폼 Go 코드의 글자 상수로
// 나오지 않습니다(코드 검토 지적 4, 재검토 N1). 글자 상수도 같은 규칙으로 쪼개 조각마다 봅니다. 플랫폼이 스스로 정한 상수(줌·띠 상태 이름)는 빼고 봅니다.
// 그래서 그 상수와 같은 글(예: process)이나 자료 파일에 없는 글(run 이름 전체, 낱말의 앞부분)로 주제를 가르는 코드는 이 시험이 잡지 못합니다.
// 이 시험은 디스크의 소스를 읽으므로 go test -overlay로 넣은 변이는 보지 못합니다.
func TestPlatformDoesNotKnowTopics(t *testing.T) {
	vocab := topicVocabulary(t)
	reserved, lits := platformLiterals(t)
	// 글자 상수도 자료 파일 값과 같은 규칙("."과 "/")으로 쪼개 조각마다 봅니다. 그래야 "strace.text"나 e.Source+".text" 같은 열쇠 조각도 걸립니다.
	for lit, where := range lits {
		for _, w := range strings.FieldsFunc(lit, func(r rune) bool { return r == '.' || r == '/' }) {
			if vocab[w] && !reserved[w] {
				t.Errorf("%s: 플랫폼 코드의 글자 %q에 주제 자료 파일의 낱말 %q", where, lit, w)
			}
		}
	}

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
//
// Given 제목에 "</script>"를 넣은 가짜 주제 보기 모형에서, When 화면 파일을 쓰면, Then 자리 표시가 남지 않고 </script>는 3개뿐이며 품은 JSON을 되읽으면 제목이 같습니다.
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

// TestNatureFollowsThePath 함수는 마디와 칸 길이의 성격이 거쳐 온 길을 따르는지 봅니다(고친 규칙 6, 코드 검토 지적 1).
//
// Given io 칸(순간) 아래 시스템 콜이 추정 근거(guess)로, 커널 나무가 다시 추정 근거로 붙은 경로에서, When 보기 모형을 만들면,
// Then 관찰 출처인 커널 마디(sleepfn, tiny)도 추정이고, io 칸의 머문 시간은 가장 긴 마디(syscall)의 길이이며 그 성격이 추정입니다.
// io 칸을 품은 시스템 칸도 길이의 성격을 받아 추정이고, 접힌 칸(touch ×2)은 접기 열쇠(dec.f)의 출처대로 「규칙으로 셈」입니다.
func TestNatureFollowsThePath(t *testing.T) {
	r := toyView(t).Routes[0]
	if sy := r.System[len(r.System)-1]; sy.Nature != atlas.Inferred {
		t.Errorf("io를 품은 시스템 칸 성격 %s, 기대 inferred", sy.Nature)
	}
	if f := r.Steps[1]; f.Fold != 2 || f.FoldNature != atlas.Derived {
		t.Errorf("접힌 칸 ×%d 성격 %q, 기대 ×2 derived", f.Fold, f.FoldNature)
	}
	st := r.Steps[4]
	for _, n := range st.Tree {
		if n.Parent >= 0 && n.Nature != atlas.Inferred {
			t.Errorf("마디 %s 성격 %s, 기대 inferred(추정 연결 아래)", n.Name, n.Nature)
		}
	}
	if st.DurNS != 15 || st.DurNature != atlas.Inferred {
		t.Errorf("io 칸 머문 시간 %d(%s), 기대 15(inferred)", st.DurNS, st.DurNature)
	}
}

// TestStepDwellPrefersItsOwnSpan 함수는 칸에 구간이 있으면 그 구간을 머문 시간으로 쓰는지 봅니다(코드 검토 지적 2).
//
// Given io 칸이 40~42의 구간을 갖고 그 아래 시스템 콜이 15만큼 걸린 경로에서, When 보기 모형을 만들면,
// Then io 칸의 머문 시간은 더 긴 마디가 아니라 칸의 구간 2이고, 그 성격은 칸의 성격입니다.
func TestStepDwellPrefersItsOwnSpan(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	for i := range r.Events {
		if r.Events[i].Name == "io" {
			r.Events[i].EndNS = atlas.Ptr(int64(42))
		}
	}
	v, err := Build(MemStore{"toy/s1/a": r}, toySpec(), ".")
	if err != nil {
		t.Fatal(err)
	}
	st := v.Routes[0].Steps[4]
	if st.DurNS != 2 || st.DurNature != st.Nature {
		t.Errorf("io 칸 머문 시간 %d(%s), 기대 2(%s)", st.DurNS, st.DurNature, st.Nature)
	}
}

// TestTimeDivergenceNamesTheLargestDifference 함수는 구조가 같은 두 나무의 갈림이 길 위에서 시간 차가 가장 큰 마디인지 봅니다(코드 검토 지적 3).
//
// Given 구조가 같고 mid의 길이만 5 다른 a·d와, 모든 것이 같은 a·e에서, When 보기 모형을 만들면,
// Then a·d의 갈림은 time:mid(차 5)이고 잎(sleepfn, 차 0)이 아니며, 뿌리 길이까지 다른 a·f도 mid이고, a·e는 나무가 같음(same)입니다.
func TestTimeDivergenceNamesTheLargestDifference(t *testing.T) {
	d := toyRun("s1/d", true, 0)
	for i := range d.Events {
		if d.Events[i].Name == "mid" {
			d.Events[i].DurNS = atlas.Ptr(int64(18))
		}
	}
	// f는 mid 차 5에 더해 뿌리(io)의 길이가 20 달라도, 뿌리의 차는 칸 머문 시간으로 이미 보이므로 갈림은 mid여야 합니다.
	f := toyRun("s1/f", true, 0)
	for i := range f.Events {
		switch f.Events[i].Name {
		case "mid":
			f.Events[i].DurNS = atlas.Ptr(int64(18))
		case "io":
			f.Events[i].EndNS = atlas.Ptr(int64(60))
		}
	}
	v, err := Build(MemStore{"toy/s1/a": toyRun("s1/a", true, 0), "toy/s1/d": d, "toy/s1/e": toyRun("s1/e", true, 0), "toy/s1/f": f}, toySpec(), ".")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Divergence{}
	for _, dv := range v.Routes[0].Divergence {
		got[v.Routes[dv.Other].Run] = dv
	}
	if g := got["toy/s1/d"]; g.Kind != "time" || g.Name != "mid" || g.DiffNS != 5 {
		t.Errorf("a·d 갈림 %+v, 기대 time:mid 차 5", g)
	}
	if g := got["toy/s1/f"]; g.Kind != "time" || g.Name != "mid" || g.DiffNS != 5 {
		t.Errorf("a·f 갈림 %+v, 기대 time:mid 차 5(뿌리 io의 차 20은 갈림이 아님)", g)
	}
	if g := got["toy/s1/e"]; g.Kind != "same" {
		t.Errorf("a·e 갈림 %+v, 기대 same(나무를 견줬고 모두 같음)", g)
	}
}

// TestMissingIdentityIsAnError 함수는 정체로 고른 attrs가 없는 run을 빈 정체로 묶지 않고 멈추는지 봅니다(코드 검토 지적 11).
//
// Given 정체 열쇠 rec.q가 없는 run에서, When 보기 모형을 만들면, Then rec.q를 이름 대는 오류로 멈춥니다.
func TestMissingIdentityIsAnError(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	delete(r.Attrs["rec"], "q")
	if _, err := Build(MemStore{"toy/s1/a": r}, toySpec(), "."); err == nil || !strings.Contains(err.Error(), "rec.q") {
		t.Fatalf("오류 %v, 기대 rec.q가 없음", err)
	}
}

// TestSystemStepsMergeSameZones 함수는 시스템 줌이 지난 구역 집합이 같은 잇단 칸만 합치는지 봅니다(코드 검토 지적 5).
//
// Given 칸이 cl, sv, sv, sv(접힌 칸 포함), 디스크·서버·경계·커널을 지나는 io 차례인 경로에서, When 보기 모형을 만들면,
// Then 시스템 칸은 [cl] 0–0, [sv] 1–3, [dk sv bd kn] 4–4 셋이고, 둘째 칸의 잎은 합친 칸들의 잎의 합집합입니다.
func TestSystemStepsMergeSameZones(t *testing.T) {
	r := toyView(t).Routes[0]
	var got []string
	for _, s := range r.System {
		got = append(got, fmt.Sprint(s.Zones, s.Steps))
	}
	if fmt.Sprint(got) != "[[cl] [0 0] [sv] [1 3] [dk sv bd kn] [4 4]]" {
		t.Errorf("시스템 칸 %v", got)
	}
	var cl [][]int
	for _, st := range r.Steps {
		cl = append(cl, st.Leaves)
	}
	if want := RangeLeaves(cl, 1, 3); !reflect.DeepEqual(r.System[1].Leaves, want) {
		t.Errorf("시스템 칸 잎 %v, 기대 %v", r.System[1].Leaves, want)
	}
}

// TestIncomingLinkAddsOnlyItsSource 함수는 들어오는 연결(dir in)이 출발 사건만 붙이고 더 따라가지 않는지 봅니다(코드 검토 지적 5).
//
// Given io로 들어오는 call 연결의 출발 사건 caller와 그 parent 자식 inner가 있고 장면이 call을 dir in으로 펼칠 때,
// When 보기 모형을 만들면, Then io 칸의 나무에 caller는 있고 inner는 없습니다.
func TestIncomingLinkAddsOnlyItsSource(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	var io, proc int
	for _, e := range r.Events {
		if e.Name == "io" {
			io, proc = e.ID, *e.Target
		}
	}
	caller := len(r.Events)
	r.Events = append(r.Events, atlas.Event{ID: caller, Place: proc, Kind: atlas.Instant, Name: "caller", Source: "rec"})
	r.Events = append(r.Events, atlas.Event{ID: caller + 1, Place: proc, Kind: atlas.Instant, Name: "inner", Parent: atlas.Ptr(caller), Source: "rec"})
	r.Links = append(r.Links, atlas.Link{ID: len(r.Links), Src: caller, Dst: io, Kind: "call", Basis: "rec"})
	spec := toySpec()
	spec.Scenes[0].Expand = append(spec.Scenes[0].Expand, Expand{Link: "call", Dir: "in"})
	v, err := Build(MemStore{"toy/s1/a": r}, spec, ".")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, n := range v.Routes[0].Steps[4].Tree {
		names[n.Name] = true
	}
	if !names["caller"] || names["inner"] {
		t.Errorf("io 나무 %v, 기대 caller 있음·inner 없음", names)
	}
}

// TestOrderOnlyRouteKeepsIDOrder 함수는 시계가 없는 출처만으로 된 경로가 id 차례의 「순서만 기록」 칸이 되는지 봅니다(고친 규칙 2).
//
// Given 시계 없는 출처 g의 시각 없는 사건 셋(id 차례 c, a, b)만 칸이 되는 장면에서, When 보기 모형을 만들면,
// Then 경로는 OrderOnly이고 칸 차례는 c, a, b이며 시계가 비어 있습니다.
func TestOrderOnlyRouteKeepsIDOrder(t *testing.T) {
	b := atlas.NewBuilder("toy/o/a", "toy", "toy", "o/a", atlas.Origin{Clock: "", Label: "없음"})
	b.AddSource(atlas.Source{ID: "g", Tool: "멈춤 기록", Nature: atlas.Observed})
	proc := b.Place(atlas.PlaceSpec{Kind: "process", Name: "42", Source: "g"})
	for _, n := range []string{"c", "a", "b"} {
		b.AddEvent(atlas.Event{Place: proc, Kind: atlas.Instant, Name: n, Source: "g"})
	}
	spec := toySpec()
	sc := &spec.Scenes[0]
	sc.Root, sc.Columns, sc.Fold = nil, nil, nil
	sc.Steps = atlas.Selector{Sources: []string{"g"}}
	sc.Identity, sc.Name = Identity{RunField: "scenario"}, Identity{RunField: "name"}
	v, err := Build(MemStore{"toy/o/a": b.Run()}, spec, ".")
	if err != nil {
		t.Fatal(err)
	}
	r := v.Routes[0]
	var names []string
	for _, s := range r.Steps {
		names = append(names, s.Name)
	}
	if !r.OrderOnly || fmt.Sprint(names) != "[c a b]" || r.Clock != "" {
		t.Errorf("순서만 %v 칸 %v 시계 %q", r.OrderOnly, names, r.Clock)
	}
}

// topicVocabulary 함수는 저장소의 모든 주제 자료 파일이 장소 종류·출처·attrs 열쇠·run 앞부분 자리에 쓴 낱말("."과 "/"로 나눈 조각)을 모읍니다.
func topicVocabulary(t *testing.T) map[string]bool {
	t.Helper()
	files, err := filepath.Glob("../../../topics/*/*/map.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("주제 자료 파일을 못 찾음: %v", err)
	}
	keys := map[string]bool{"kinds": true, "sources": true, "attr": true, "event_named_kinds": true, "untimed_text": true, "runs": true, "run_attr": true, "root_attr": true, "typed": true, "stage": true}
	vocab := map[string]bool{}
	var walk func(v any, key string)
	walk = func(v any, key string) {
		switch x := v.(type) {
		case map[string]any:
			for k, c := range x {
				walk(c, k)
			}
		case []any:
			for _, c := range x {
				walk(c, key)
			}
		case string:
			if keys[key] {
				for _, w := range strings.FieldsFunc(x, func(r rune) bool { return r == '.' || r == '/' }) {
					vocab[w] = true
				}
			}
		}
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var v any
		if err := json.Unmarshal(b, &v); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		walk(v, "")
	}
	return vocab
}

// platformLiterals 함수는 플랫폼 Go 파일(시험 제외)의 글자 상수를 모읍니다. 구조체 꼬리표는 빼고, 이름 있는 const 선언의 값은 reserved로 따로 돌려줍니다.
func platformLiterals(t *testing.T) (reserved map[string]bool, lits map[string]string) {
	t.Helper()
	reserved, lits = map[string]bool{}, map[string]string{}
	files, _ := filepath.Glob("*.go")
	files = append(files, "../../../cmd/atlas-map/main.go")
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		tags := map[*ast.BasicLit]bool{}
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Field:
				if x.Tag != nil {
					tags[x.Tag] = true
				}
			case *ast.GenDecl:
				if x.Tok == token.CONST {
					for _, sp := range x.Specs {
						for _, v := range sp.(*ast.ValueSpec).Values {
							if bl, ok := v.(*ast.BasicLit); ok && bl.Kind == token.STRING {
								s, _ := strconv.Unquote(bl.Value)
								reserved[s] = true
							}
						}
					}
				}
			case *ast.BasicLit:
				if x.Kind == token.STRING && !tags[x] {
					s, _ := strconv.Unquote(x.Value)
					lits[s] = fset.Position(x.Pos()).String()
				}
			}
			return true
		})
	}
	return reserved, lits
}

// TestViaBasisWeakensSteps 함수는 뿌리에서 칸까지 거친 연결 근거가 칸의 성격에 들어가는지 봅니다(고친 규칙 6, 코드 검토 지적 1의 c).
//
// Given 뿌리 req에서 칸 사건마다 해석기 근거(dec, 도출)의 contains 연결이 있고 장면이 그 연결로만 칸을 고를 때, When 보기 모형을 만들면,
// Then 관찰 출처의 칸과 그 나무의 마디(뿌리 포함), 시각 없는 사건 목록도 모두 「규칙으로 셈」(derived) 이하입니다.
func TestViaBasisWeakensSteps(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	root := -1
	for _, e := range r.Events {
		if e.Name == "req" {
			root = e.ID
		}
	}
	for _, e := range r.Events {
		if e.Source == "rec" {
			r.Links = append(r.Links, atlas.Link{ID: len(r.Links), Src: root, Dst: e.ID, Kind: "contains", Basis: "dec"})
		}
	}
	spec := toySpec()
	spec.Scenes[0].Via = []string{"contains"}
	v, err := Build(MemStore{"toy/s1/a": r}, spec, ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range v.Routes[0].Steps {
		if st.Nature == atlas.Observed {
			t.Errorf("칸 %s 성격 observed, 기대 derived 이하(도출 연결로 고른 칸)", st.Name)
		}
		for _, n := range st.Tree {
			if n.Nature == atlas.Observed {
				t.Errorf("칸 %s의 마디 %s 성격 observed, 기대 derived 이하(나무 뿌리도 칸까지 온 길을 받음)", st.Name, n.Name)
			}
		}
	}
	un := v.Routes[0].Untimed
	if len(un) != 1 || un[0].Nature == atlas.Observed {
		t.Errorf("시각 없는 사건 %+v, 기대 하나이고 derived 이하", un)
	}
}

// TestFoldedDescentCountsRecorded 함수는 접힌 칸의 하강 가운데 아래가 기록된 수를 따로 세는지 봅니다(코드 검토 지적 9).
//
// Given 접힌 칸(touch ×2)의 두 사건이 각각 시스템 콜로 내려가고 그 가운데 하나만 커널 나무를 가진 경로에서, When 보기 모형을 만들면,
// Then 그 칸의 하강은 지면 마디 2개 가운데 1개 기록(state events, count 2, recorded 1)입니다.
func TestFoldedDescentCountsRecorded(t *testing.T) {
	r := toyRun("s1/a", true, 0)
	var touches []int
	kernelRoot := -1
	for _, e := range r.Events {
		switch {
		case e.Name == "touch" && len(touches) < 2:
			touches = append(touches, e.ID)
		case e.Name == "entry":
			kernelRoot = e.ID
		}
	}
	proc := r.Events[0].Place
	for k, id := range touches {
		sys := len(r.Events)
		r.Events = append(r.Events, atlas.Event{ID: sys, Place: proc, Kind: atlas.Span, Name: "syscall", DurNS: atlas.Ptr(int64(2)), Source: "sys"})
		r.Links = append(r.Links, atlas.Link{ID: len(r.Links), Src: id, Dst: sys, Kind: "syscall", Basis: "guess"})
		if k == 0 {
			child := len(r.Events)
			r.Events = append(r.Events, atlas.Event{ID: child, Place: r.Events[kernelRoot].Place, Kind: atlas.Span, Name: "entry", DurNS: atlas.Ptr(int64(1)), Source: "kt",
				Attrs: atlas.Attrs{"kt": {"layer": "L1"}}})
			r.Links = append(r.Links, atlas.Link{ID: len(r.Links), Src: sys, Dst: child, Kind: "kernel", Basis: "guess"})
		}
	}
	v, err := Build(MemStore{"toy/s1/a": r}, toySpec(), ".")
	if err != nil {
		t.Fatal(err)
	}
	d := v.Routes[0].Steps[1].Descent
	if d == nil || d.State != StateEvents || d.Count != 2 || d.Recorded != 1 {
		t.Errorf("접힌 칸 하강 %+v, 기대 events·count 2·recorded 1", d)
	}
}
