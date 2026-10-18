package atlas

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// sampleRun 함수는 시험에 쓸 작은 run을 만듭니다. 시계 둘(mac, container), 출처 셋, 장소 셋, 사건 넷, 연결 하나입니다.
func sampleRun() *Run {
	b := NewBuilder("r1", "pgsql", "s", "a", Origin{Clock: "mac", Label: "시작"})
	b.AddSource(Source{ID: "wire", Tool: "relay", Clock: "mac", Nature: Observed})
	b.AddSource(Source{ID: "rec", Tool: "uprobe", Clock: "container", Nature: Observed})
	b.AddSource(Source{ID: "dec", Tool: "decode", Nature: Derived})
	b.AddClock(Clock{ID: "mac", PrecisionNS: 1000})
	b.AddClock(Clock{ID: "container", PrecisionNS: 1, Relation: &ClockRelation{To: "mac", Known: true, LoNS: -2000, HiNS: 3000, Basis: "dec"}})
	conn := b.Place(PlaceSpec{Kind: "connection", Name: "C1", Source: "wire"})
	backend := b.Place(PlaceSpec{Kind: "process", Name: "backend 5531", Source: "rec"})
	fn := b.Place(PlaceSpec{Parent: &backend, Kind: "function", Name: "exec_simple_query", Loc: &Loc{Path: "src/backend/tcop/postgres.c"}, Source: "rec", Static: true, Version: "18.6"})
	q := b.AddEvent(Event{Place: conn, Kind: Instant, Name: "F:Q", StartNS: Ptr[int64](1000), EndNS: Ptr[int64](1000), Source: "wire"})
	b.EventAttrs(q, "wire", "len", json.Number("27"))
	w := b.AddEvent(Event{Place: backend, Kind: Span, Name: "window", StartNS: Ptr[int64](900), EndNS: Ptr[int64](9000), Source: "dec"})
	_ = b.AddEvent(Event{Place: backend, Kind: Span, Name: "inner", StartNS: Ptr[int64](1000), EndNS: Ptr[int64](2000), Parent: &w, Source: "dec"})
	s := b.AddEvent(Event{Place: fn, Kind: Instant, Name: "lock", StartNS: Ptr[int64](1500), EndNS: Ptr[int64](1500), Stack: []int{fn}, Source: "rec"})
	b.AddLink(w, s, "contains", "dec")
	return b.Run()
}

// TestValidateAcceptsWellFormedRun 함수는 규칙을 지켜 조립한 run이 형식 검사를 통과하는지 확인합니다.
func TestValidateAcceptsWellFormedRun(t *testing.T) {
	// Given: 형식 규칙을 지켜 조립한 run
	r := sampleRun()

	// When: 형식 검사를 하면
	errs := Validate(r)

	// Then: 어긴 곳이 없다
	if len(errs) != 0 {
		t.Fatalf("어긴 곳 %v", errs)
	}
}

// TestValidateCatchesEachBrokenRule 함수는 형식 규칙을 하나씩 어긴 run마다 그 규칙의 지적이 나오는지 확인합니다. 검사가 실제로 결함을 가려내는지 보는 시험입니다.
func TestValidateCatchesEachBrokenRule(t *testing.T) {
	cases := []struct {
		name   string
		break_ func(r *Run)
		want   string
	}{
		{"다른 출처의 부모", func(r *Run) { r.Events[2].Source = "rec" }, "부모"},
		{"틀린 깊이", func(r *Run) { r.Events[2].Depth = 5 }, "깊이"},
		{"group 부모", func(r *Run) { r.Events[1].Kind = Group; r.Events[1].Count = 3 }, "group"},
		{"count 1인 group", func(r *Run) { r.Events[0].Kind = Group }, "count"},
		{"시작과 길이 함께", func(r *Run) { r.Events[0].DurNS = Ptr[int64](5) }, "길이"},
		{"거꾸로 된 구간", func(r *Run) { r.Events[1].EndNS = Ptr[int64](1) }, "시작이 끝보다"},
		{"출처가 아닌 attrs 키", func(r *Run) { r.Events[0].Attrs["nope"] = map[string]any{"x": 1} }, "출처 id가 아님"},
		{"없는 장소", func(r *Run) { r.Events[0].Place = 99 }, "없는 장소"},
		{"없는 출발 장소", func(r *Run) { r.Events[0].From = Ptr(77) }, "출발"},
		{"없는 스택 장소", func(r *Run) { r.Events[3].Stack = []int{42} }, "스택"},
		{"스택 0번이 장소가 아님", func(r *Run) { r.Events[3].Stack = []int{1, 2} }, "스택 0번"},
		{"없는 연결 끝", func(r *Run) { r.Links[0].Dst = 99 }, "없는 사건"},
		{"모르는 성격", func(r *Run) { r.Sources[0].Nature = "guess" }, "성격"},
		{"장소 고리", func(r *Run) { r.Places[1].Parent = Ptr(2) }, "고리"},
		{"관계 시계 없음", func(r *Run) { r.Clocks[1].Relation.To = "x" }, "없는 시계"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Given: 규칙 하나를 어긴 run
			r := sampleRun()
			c.break_(r)

			// When: 형식 검사를 하면
			errs := Validate(r)

			// Then: 그 규칙을 어겼다고 적는다
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), c.want) {
					found = true
				}
			}
			if !found {
				t.Fatalf("%q가 든 지적이 없음: %v", c.want, errs)
			}
		})
	}
}

// TestWriteReadRoundTripKeepsNumbersAsWritten 함수는 쓰고 다시 읽은 run에서 attrs의 숫자가 글자 그대로 남는지 확인합니다.
func TestWriteReadRoundTripKeepsNumbersAsWritten(t *testing.T) {
	// Given: attrs에 json.Number가 있는 run
	r := sampleRun()

	// When: 쓰고 다시 읽으면
	var buf bytes.Buffer
	if err := Write(&buf, r); err != nil {
		t.Fatal(err)
	}
	got, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Then: 숫자의 글자까지 그대로다
	if v := got.Events[0].Attrs["wire"]["len"]; v != json.Number("27") {
		t.Fatalf("len = %#v", v)
	}
	if len(Validate(got)) != 0 {
		t.Fatal("다시 읽은 run이 형식 검사를 통과하지 못함")
	}
}

// TestReadRejectsUnknownField 함수는 형식에 없는 칸을 조용히 버리지 않고 오류로 내는지 확인합니다.
func TestReadRejectsUnknownField(t *testing.T) {
	// Given: 형식에 없는 칸이 든 JSON
	in := `{"id":"x","surprise":1}`

	// When: 읽으면
	_, err := Read(strings.NewReader(in))

	// Then: 조용히 버리지 않고 오류다
	if err == nil {
		t.Fatal("모르는 칸을 받아들임")
	}
}

// TestStaticPlaceKeyIgnoresDynamicParent 함수는 같은 함수가 다른 프로세스 아래에 있어도 같은 key를 받는지 확인합니다.
func TestStaticPlaceKeyIgnoresDynamicParent(t *testing.T) {
	// Given: 서로 다른 프로세스 아래의 같은 함수
	b := NewBuilder("r", "t", "s", "a", Origin{})
	b.AddSource(Source{ID: "rec", Tool: "x", Nature: Observed})
	p1 := b.Place(PlaceSpec{Kind: "process", Name: "5531", Source: "rec"})
	p2 := b.Place(PlaceSpec{Kind: "process", Name: "5903", Source: "rec"})
	spec := func(parent int) PlaceSpec {
		return PlaceSpec{Parent: &parent, Kind: "function", Name: "f", Loc: &Loc{Path: "a.c"}, Source: "rec", Static: true, Version: "18.6"}
	}

	// When: 두 장소를 만들면
	f1, f2 := b.Place(spec(p1)), b.Place(spec(p2))

	// Then: 장소는 둘이지만 key는 같다(기록을 가로질러 잇는 열쇠)
	r := b.Run()
	if f1 == f2 || r.Places[f1].Key != r.Places[f2].Key || r.Places[f1].Key == "" {
		t.Fatalf("f1=%d f2=%d key1=%q key2=%q", f1, f2, r.Places[f1].Key, r.Places[f2].Key)
	}
}

// TestDiffComparesNumbersByValueAndTypesStrictly 함수는 깊은 비교가 숫자는 값으로, 타입은 엄격하게 견주는지 확인합니다.
func TestDiffComparesNumbersByValueAndTypesStrictly(t *testing.T) {
	// Given: 글자는 다르지만 값이 같은 숫자, 타입만 다른 식별자, 한쪽에만 있는 키
	a := map[string]any{"t": json.Number("56"), "conn": json.Number("1"), "only": true}
	b := map[string]any{"t": 56.0, "conn": "1"}

	// When: 견주면
	d := Diff("$", a, b)

	// Then: 숫자는 같다고 보고, 타입이 다른 식별자와 빠진 키를 잡는다
	joined := strings.Join(d, "\n")
	if strings.Contains(joined, "$.t") || !strings.Contains(joined, "$.conn") || !strings.Contains(joined, "$.only") {
		t.Fatalf("차이: %v", d)
	}
}

// TestMicrosToNanosRoundTrips 함수는 원본 마이크로초를 나노초로 옮겼다 되돌리면 같은 float64가 되는지 확인합니다.
func TestMicrosToNanosRoundTrips(t *testing.T) {
	for _, s := range []string{"0", "95.172", "4479425.793", "4456776", "-3.5"} {
		// Given: 원본의 마이크로초 글
		n := json.Number(s)

		// When: 나노초로 옮기고 되돌리면
		ns, err := MicrosToNanos(n)
		if err != nil {
			t.Fatal(err)
		}
		f, _ := n.Float64()

		// Then: 원본 float64와 같다
		if NanosToMicros(ns) != f {
			t.Fatalf("%s → %d → %v", s, ns, NanosToMicros(ns))
		}
	}
}

// TestMicrosToNanosRefusesSubNanosecond 함수는 나노초보다 촘촘한 값을 조용히 버리지 않고 오류로 내는지 확인합니다.
func TestMicrosToNanosRefusesSubNanosecond(t *testing.T) {
	// Given: 나노초보다 촘촘한 마이크로초
	n := json.Number("1.0004")

	// When: 옮기면
	_, err := MicrosToNanos(n)

	// Then: 값을 잃는다고 오류다
	if err == nil {
		t.Fatal("나노초 아래 자리를 조용히 버림")
	}
}

// TestCompareStartRespectsClockRelation 함수는 시계가 다른 두 사건의 순서를 관계 범위가 허락할 때만 내는지 확인합니다.
func TestCompareStartRespectsClockRelation(t *testing.T) {
	r := sampleRun()
	ev := func(src string, ns int64) Event {
		return Event{Kind: Instant, StartNS: Ptr(ns), Source: src}
	}
	cases := []struct {
		name string
		a, b Event
		want Order
	}{
		{"같은 시계", ev("wire", 10), ev("wire", 20), Before},
		{"시계가 달라도 범위 밖이면 순서", ev("wire", 0), ev("rec", 10_000), Before},
		{"범위 안이면 견줄 수 없음", ev("wire", 1000), ev("rec", 0), Incomparable},
		{"반대 방향도 범위로", ev("rec", 10_000), ev("wire", 0), After},
		{"시계 없는 출처", ev("dec", 0), ev("wire", 5), Incomparable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// When: 두 사건의 시작을 견주면
			got, why := CompareStart(r, c.a, c.b)

			// Then: 시계 관계가 허락하는 만큼만 순서를 낸다
			if got != c.want {
				t.Fatalf("got %v want %v (%s)", got, c.want, why)
			}
		})
	}
}

// TestScanStructureAllowsOnlyListedKeys 함수는 구조 숨김 검사가 허용 목록 밖의 키(구조 역할, 표시 파일에 없는 조각)를 잡고 예외 부류는 통과시키는지 확인합니다.
func TestScanStructureAllowsOnlyListedKeys(t *testing.T) {
	// Given: 메시지 사건 하나와 그 경로 무리의 판정 기준
	r := sampleRun()
	origins := Origins{{"event", 0}: "$.conns[].messages[]"}
	roles := Roles{
		"$.conns[].messages[].len":  {Role: "값", Prov: "관찰"},
		"$.conns[].messages[].t":    {Role: "시각", Prov: "관찰", Decision: "start_ns"},
		"$.conns[].messages[].name": {Role: "장소", Prov: "관찰", Decision: "X2"},
	}

	// When/Then: 값 역할의 키만 있으면 통과한다
	if errs := ScanStructure(r, origins, roles); len(errs) != 0 {
		t.Fatalf("통과해야 함: %v", errs)
	}

	// When/Then: 시각 역할의 키를 attrs에 남기면 실패한다(설계 메모의 변이 시험)
	r.Events[0].Attrs["wire"]["t"] = json.Number("1")
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "구조 칸") {
		t.Fatalf("t를 잡아야 함: %v", errs)
	}

	// When/Then: 표시 파일에 없는 키(원본 조각을 통째로 넣은 경우)도 실패한다
	delete(r.Events[0].Attrs["wire"], "t")
	r.Events[0].Attrs["wire"]["raw"] = map[string]any{"t": 1, "dir": "F"}
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "표시 파일에 없음") {
		t.Fatalf("조각을 잡아야 함: %v", errs)
	}

	// When/Then: 구조 역할이어도 예외 부류로 정한 키는 통과한다
	delete(r.Events[0].Attrs["wire"], "raw")
	r.Events[0].Attrs["wire"]["name"] = "S_1"
	if errs := ScanStructure(r, origins, roles); len(errs) != 0 {
		t.Fatalf("예외 부류는 통과: %v", errs)
	}

	// When/Then: 모음 키라도 그 아래에 예외 아닌 구조 역할이 있으면 실패하고, 아래가 예외 부류뿐이면 통과한다
	roles["$.conns[].messages[].frames"] = Role{Role: "모음", Prov: "관찰"}
	roles["$.conns[].messages[].frames[].fn"] = Role{Role: "장소", Prov: "관찰", Decision: "stack"}
	roles["$.conns[].messages[].vals"] = Role{Role: "모음", Prov: "관찰"}
	roles["$.conns[].messages[].vals[]"] = Role{Role: "사건", Prov: "관찰", Decision: "X7"}
	r.Events[0].Attrs["wire"]["frames"] = []any{map[string]any{"fn": "f"}}
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "frames[].fn") {
		t.Fatalf("모음 안의 장소를 잡아야 함: %v", errs)
	}
	delete(r.Events[0].Attrs["wire"], "frames")
	r.Events[0].Attrs["wire"]["vals"] = []any{"x=1"}
	if errs := ScanStructure(r, origins, roles); len(errs) != 0 {
		t.Fatalf("아래가 예외 부류뿐이면 통과: %v", errs)
	}
}

// TestScanStructureWalksIntoValues 함수는 구조 숨김 검사가 허용된 키의 값 안까지 걸어 내려가, 표시 파일에 없는 아래 경로를 잡는지 확인합니다
// (Spring 코드 검토 지적 5의 세 사례: 그릇에 원본에 없던 열쇠를 더함, 허용된 열쇠 아래에 다른 경로의 목록을 넣음, run attrs 안에 넣음).
func TestScanStructureWalksIntoValues(t *testing.T) {
	// Given: 메시지 사건 하나와 run, 그 경로 무리의 판정 기준. f는 칸이 정해진 그릇, params는 동적 열쇠의 그릇, tables는 run attrs의 목록입니다.
	r := sampleRun()
	origins := Origins{{"event", 0}: "$.conns[].messages[]", {"run", 0}: "$"}
	roles := Roles{
		"$.conns[].messages[].len":        {Role: "값", Prov: "관찰"},
		"$.conns[].messages[].f":          {Role: "모음", Prov: "관찰"},
		"$.conns[].messages[].f.code":     {Role: "값", Prov: "관찰"},
		"$.conns[].messages[].params":     {Role: "모음", Prov: "관찰"},
		"$.conns[].messages[].params.{*}": {Role: "값", Prov: "관찰"},
		"$.tables":                        {Role: "모음", Prov: "관찰"},
		"$.tables[]":                      {Role: "값", Prov: "관찰"},
		"$.tables[].name":                 {Role: "이름", Prov: "관찰"},
	}
	r.Events[0].Attrs["wire"]["f"] = map[string]any{"code": "40001"}
	r.Events[0].Attrs["wire"]["params"] = map[string]any{"user": "lab", "database": "lab"}
	r.Attrs = Attrs{"txlab": {"tables": []any{map[string]any{"name": "acct"}}}}

	// When/Then: 값 안의 경로가 모두 표시 파일에 있으면(동적 열쇠는 {*}) 통과한다
	if errs := ScanStructure(r, origins, roles); len(errs) != 0 {
		t.Fatalf("통과해야 함: %v", errs)
	}

	// When/Then: 그릇에 원본에 없던 열쇠(장소를 가리키는 conn)를 더하면 실패한다
	r.Events[0].Attrs["wire"]["f"].(map[string]any)["conn"] = "C2"
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "f.conn가 표시 파일에 없음") {
		t.Fatalf("그릇에 더한 열쇠를 잡아야 함: %v", errs)
	}
	delete(r.Events[0].Attrs["wire"]["f"].(map[string]any), "conn")

	// When/Then: 동적 열쇠 그릇이라도 그 값 안에 다른 경로의 목록(요청)을 넣으면 실패한다
	r.Events[0].Attrs["wire"]["params"].(map[string]any)["user"] = []any{map[string]any{"seq": 1}}
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "params.{*}[]") {
		t.Fatalf("허용된 열쇠 아래의 목록을 잡아야 함: %v", errs)
	}
	r.Events[0].Attrs["wire"]["params"].(map[string]any)["user"] = "lab"

	// When/Then: run attrs의 목록 원소에 다른 경로(연결)를 넣어도 실패한다
	r.Attrs["txlab"]["tables"].([]any)[0].(map[string]any)["conns"] = []any{}
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "$.tables[].conns") {
		t.Fatalf("run attrs 안의 조각을 잡아야 함: %v", errs)
	}
	delete(r.Attrs["txlab"]["tables"].([]any)[0].(map[string]any), "conns")

	// When/Then: Go 타입이 정해진 그릇은 안을 걸을 수 없어 실패한다(Spring 재검토 N1: 같은 내용을 map[string]any로 넣으면 위처럼 잡힘)
	r.Events[0].Attrs["wire"]["f"] = map[string]string{"conn": "C2"}
	if errs := ScanStructure(r, origins, roles); len(errs) != 1 || !strings.Contains(errs[0].Error(), "걸을 수 없는 그릇") {
		t.Fatalf("타입이 정해진 그릇을 잡아야 함: %v", errs)
	}
}

// TestCheckNatureFollowsMarksAndInferredList 함수는 출처 성격 검사가 표시 출처와 추정 목록에 어긋난 근거를 잡는지 확인합니다.
func TestCheckNatureFollowsMarksAndInferredList(t *testing.T) {
	// Given: 사건 넷과 장소 셋의 경로 무리, 표시 출처
	r := sampleRun()
	origins := Origins{
		{"place", 0}: "$.conns[]", {"place", 1}: "$.pid", {"place", 2}: "$.stack[].fn",
		{"event", 0}: "$.conns[].messages[]", {"event", 1}: "$.server", {"event", 2}: "$.server.inner",
		{"event", 3}: "$.events[]", {"link", 0}: "$.server.events",
	}
	roles := Roles{
		"$.conns[]": {Prov: "관찰"}, "$.pid": {Prov: "관찰"}, "$.stack[].fn": {Prov: "관찰"},
		"$.conns[].messages[]": {Prov: "관찰"}, "$.server": {Prov: "도출"}, "$.server.inner": {Prov: "도출"},
		"$.events[]": {Prov: "관찰"}, "$.server.events": {Prov: "도출"},
	}

	// When/Then: 기대와 맞으면 통과한다
	if errs := CheckNature(r, origins, roles, nil); len(errs) != 0 {
		t.Fatalf("통과해야 함: %v", errs)
	}

	// When/Then: 해석기가 정한 연결의 근거를 관찰 출처로 바꾸면 실패한다(설계 메모의 변이 시험)
	r.Links[0].Basis = "rec"
	if errs := CheckNature(r, origins, roles, nil); len(errs) != 1 {
		t.Fatalf("근거 바꿈을 잡아야 함: %v", errs)
	}

	// When/Then: 추정 목록에 든 무리가 derived면 실패한다
	r.Links[0].Basis = "dec"
	if errs := CheckNature(r, origins, roles, map[string]bool{"$.server.events": true}); len(errs) != 1 {
		t.Fatalf("추정이어야 함: %v", errs)
	}

	// When/Then: 추정 목록에 든 관찰 범위(짐작한 감시 목록)의 출처가 추정이 아니면 실패하고, 추정이면 통과한다
	r.Coverage = append(r.Coverage, Coverage{Source: "dec", What: "함수 진입", Completeness: "partial", Basis: "추정"})
	origins[RecordRef{"coverage", 0}] = "topic:watch-guess"
	guess := map[string]bool{"topic:watch-guess": true}
	if errs := CheckNature(r, origins, roles, guess); len(errs) != 1 || !strings.Contains(errs[0].Error(), "coverage 0") {
		t.Fatalf("짐작한 관찰 범위를 잡아야 함: %v", errs)
	}
	r.Sources = append(r.Sources, Source{ID: "guess", Tool: "짐작", Nature: Inferred})
	r.Coverage[0].Source = "guess"
	if errs := CheckNature(r, origins, roles, guess); len(errs) != 0 {
		t.Fatalf("추정 출처면 통과: %v", errs)
	}

	// When/Then: 해석기 규칙 무리(rule:)는 표시 파일 없이 derived만 받는다. 관찰이나 추정 근거면 실패한다
	origins[RecordRef{"link", 0}] = "rule:window"
	if errs := CheckNature(r, origins, roles, guess); len(errs) != 0 {
		t.Fatalf("해석기 근거면 통과: %v", errs)
	}
	for _, basis := range []string{"rec", "guess"} {
		r.Links[0].Basis = basis
		if errs := CheckNature(r, origins, roles, guess); len(errs) != 1 || !strings.Contains(errs[0].Error(), "rule:window") {
			t.Fatalf("근거 %s를 잡아야 함: %v", basis, errs)
		}
	}
}

// TestCrossingSpansFindsOnlyMisnestedPairs 함수는 같은 장소에서 어긋나 겹치는 구간만 찾고 바르게 포개진 구간은 내지 않는지 확인합니다.
func TestCrossingSpansFindsOnlyMisnestedPairs(t *testing.T) {
	// Given: 같은 장소에 바르게 포개진 구간 둘(sampleRun)과, 어긋나 겹치는 구간 하나
	r := sampleRun()
	b := &Builder{run: r, places: map[string]int{}}
	x := b.AddEvent(Event{Place: 1, Kind: Span, Name: "cross", StartNS: Ptr[int64](1500), EndNS: Ptr[int64](2500), Source: "dec"})

	// When: 어긋난 쌍을 찾으면
	got := CrossingSpans(r)

	// Then: 새 구간과 그것이 걸친 구간 하나만 나온다
	if len(got) != 1 || got[0].B != x || got[0].A != 2 {
		t.Fatalf("어긋난 쌍 %v", got)
	}
}

// TestChromeExportSeparatesClocksAndSkipsUntimed 함수는 Chrome 내보내기가 시계마다 묶음을 나누고 시각 없는 사건을 세어 빼는지 확인합니다.
func TestChromeExportSeparatesClocksAndSkipsUntimed(t *testing.T) {
	// Given: 두 시계의 사건과 시각 없는 사건 하나
	r := sampleRun()
	b := &Builder{run: r, places: map[string]int{}}
	b.AddEvent(Event{Place: 2, Kind: Instant, Name: "untimed", Source: "rec"})

	// When: 내보내면
	tr, skipped := ChromeExport(r)

	// Then: 시계마다 프로세스 묶음이 다르고, 시각 없는 사건은 세어서 뺀다
	pids := map[string]int{}
	for _, e := range tr.TraceEvents {
		if e["ph"] == "M" || e["cat"] == nil {
			continue
		}
		pids[e["cat"].(string)] = e["pid"].(int)
	}
	if pids["wire"] == pids["rec"] || skipped["시작 시각 없음"] != 1 {
		t.Fatalf("pid %v, 뺀 것 %v", pids, skipped)
	}
}

// TestSelectReturnsStructureAndNature 함수는 재추출이 구조 칸과 고른 값, 출처 성격을 함께 내는지 확인합니다.
func TestSelectReturnsStructureAndNature(t *testing.T) {
	// Given: 출처가 다른 사건들
	r := sampleRun()

	// When: 연결 위 메시지를 고르면
	rows := Select(r, Selector{PlaceKinds: []string{"connection"}, Attrs: []string{"wire.len"}})

	// Then: 구조 칸과 고른 값, 출처 성격이 나온다
	if len(rows) != 1 || rows[0].Name != "F:Q" || rows[0].Attrs["wire.len"] != json.Number("27") || rows[0].Nature != Observed {
		t.Fatalf("rows %+v", rows)
	}

	// When/Then: Require는 그 값이 있는 사건만 고른다
	if got := Count(r, Selector{Require: []string{"wire.len"}}); got != 1 {
		t.Fatalf("Require로 고른 수 %d", got)
	}

	// When/Then: 출발·도착 장소는 위치를 붙인 이름표로 나온다
	r.Events[0].From, r.Events[0].Target = Ptr(2), Ptr(1)
	rows = Select(r, Selector{PlaceKinds: []string{"connection"}})
	if rows[0].From != "function exec_simple_query@src/backend/tcop/postgres.c" || rows[0].Target != "process backend 5531" {
		t.Fatalf("출발 %q, 도착 %q", rows[0].From, rows[0].Target)
	}
	links := Linked(r, "contains")
	if len(links) != 1 || links[0][0].Name != "window" || links[0][1].Nature != Derived {
		t.Fatalf("links %+v", links)
	}
}

// TestExtractFollowsOneRuleKind 함수는 재추출 규칙이 종류 하나(고르기·설명·연결)만 쓸 때 그 결과를 내고, 둘을 쓰거나 아무것도 쓰지 않으면 거부하는지 확인합니다.
func TestExtractFollowsOneRuleKind(t *testing.T) {
	// Given: 설명 하나가 있는 기록과 규칙 셋
	r := sampleRun()
	b := &Builder{run: r, places: map[string]int{}}
	b.AddNote(Note{Scope: Scope{Kind: "run"}, Kind: "output", Text: "2|b", Source: "rec"})
	rules := []FactRule{
		{Fact: "메시지", Select: &Selector{PlaceKinds: []string{"connection"}}},
		{Fact: "출력", Notes: "output"},
		{Fact: "포함", Linked: "contains"},
	}

	// When: 규칙대로 뽑으면
	facts, err := Extract(r, rules)

	// Then: 규칙 차례로 해당 칸 하나만 찬다
	if err != nil || len(facts) != 3 || len(facts[0].Rows) != 1 || facts[1].Notes[0] != "2|b" || len(facts[2].Pairs) != 1 {
		t.Fatalf("facts %+v, err %v", facts, err)
	}

	// When/Then: 종류를 둘 쓰거나 아무것도 쓰지 않은 규칙은 거부한다
	for _, bad := range []FactRule{{Fact: "둘", Notes: "output", Linked: "contains"}, {Fact: "없음"}} {
		if _, err := Extract(r, []FactRule{bad}); err == nil {
			t.Fatalf("규칙 %q를 거부해야 함", bad.Fact)
		}
	}
}
