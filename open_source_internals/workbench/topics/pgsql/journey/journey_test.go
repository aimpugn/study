package journey

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
	"study/open_source_internals/workbench/topics/pgsql/roles"
)

// designDir 는 이 패키지 폴더에서 화면 파일(concept-3.html, concept-4.html)이 있는 폴더까지의 경로입니다.
const designDir = "../../../design/"

// version 은 시험에 쓰는 소스 판입니다. 실습의 workbench.json의 sourceRoot와 같습니다.
const version = "/src/postgresql-18.6"

// loaded 값은 질의 하나의 원본(concept-4와 concept-3)과 적재 결과입니다.
type loaded struct {
	key    string
	q, c3  map[string]any
	result Result
}

// loadAll 함수는 concept-4의 질의 11개를 concept-3의 같은 기록과 함께 적재합니다. 하나라도 실패하면 시험을 멈춥니다.
func loadAll(t *testing.T) []loaded {
	t.Helper()
	read := func(name string) string {
		b, err := os.ReadFile(designDir + name)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	queries, err := ReadConcept4(read("concept-4.html"))
	if err != nil {
		t.Fatal(err)
	}
	c3, err := ReadConcept3(read("concept-3.html"))
	if err != nil {
		t.Fatal(err)
	}
	if len(queries) != 11 || len(c3) != 2 {
		t.Fatalf("질의 %d개(기대 11), concept-3 %d개(기대 2)", len(queries), len(c3))
	}
	var out []loaded
	for _, raw := range queries {
		q := raw.(map[string]any)
		key := q["key"].(string)
		var c3data map[string]any
		if v, ok := c3[key]; ok {
			c3data = v.(map[string]any)
		}
		res, err := Import(q, c3data, version)
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		out = append(out, loaded{key: key, q: q, c3: c3data, result: res})
	}
	return out
}

// TestRoundTripRebuildsBothScreens 함수는 적재 결과를 파일로 쓰고 다시 읽어 되돌린 질의가 concept-4와 concept-3의 원본과 깊이 같은지 확인합니다(검사 1).
// 연결 표의 차례만 뒤집은 사본도 같아야 합니다.
func TestRoundTripRebuildsBothScreens(t *testing.T) {
	for _, l := range loadAll(t) {
		// Given: 적재 결과를 형식의 JSON으로 쓴 것
		var buf bytes.Buffer
		if err := atlas.Write(&buf, l.result.Run); err != nil {
			t.Fatal(err)
		}
		back, err := atlas.Read(&buf)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}

		// When: 그 JSON만 읽어 두 화면의 원본 꼴로 되돌리면
		q, err := ExportC4(back)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}
		c3, hasC3, err := ExportC3(back)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}

		// Then: 원본과 차이가 없다
		assertSame(t, l.key, l.q, q)
		if hasC3 != (l.c3 != nil) {
			t.Fatalf("%s: concept-3 있음 %v, 원본 %v", l.key, hasC3, l.c3 != nil)
		}
		if hasC3 {
			assertSame(t, l.key+"(concept-3)", l.c3, c3)
		}

		// Then: 연결 표의 차례만 뒤집은 사본도 같게 되돌아온다(형식은 연결 표의 차례에 뜻을 두지 않음, Spring 코드 검토 지적 8)
		for i, j := 0, len(back.Links)-1; i < j; i, j = i+1, j-1 {
			back.Links[i], back.Links[j] = back.Links[j], back.Links[i]
		}
		for i := range back.Links {
			back.Links[i].ID = i
		}
		if errs := atlas.Validate(back); len(errs) != 0 {
			t.Fatalf("%s: 연결 차례를 뒤집은 사본의 형식 검사 %v", l.key, errs[0])
		}
		if q, err = ExportC4(back); err != nil {
			t.Fatalf("%s: 연결 차례를 뒤집은 사본: %v", l.key, err)
		}
		assertSame(t, l.key+"(연결 차례 뒤집음)", l.q, q)
		if c3, _, err = ExportC3(back); err != nil {
			t.Fatalf("%s: 연결 차례를 뒤집은 사본: %v", l.key, err)
		}
		if hasC3 {
			assertSame(t, l.key+"(concept-3, 연결 차례 뒤집음)", l.c3, c3)
		}
	}
}

// assertSame 함수는 원본과 되돌린 값을 같은 꼴로 바꿔 깊이 견주고, 다르면 처음 몇 곳을 보이고 멈춥니다.
func assertSame(t *testing.T, name string, want, got any) {
	t.Helper()
	w, err := load.Canonical(want)
	if err != nil {
		t.Fatal(err)
	}
	g, err := load.Canonical(got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := atlas.Diff("$", w, g); len(diff) != 0 {
		if len(diff) > 10 {
			diff = diff[:10]
		}
		t.Fatalf("%s: 차이 %d곳 이상\n%s", name, len(diff), strings.Join(diff, "\n"))
	}
}

// TestImportedRunsPassFormatAndProvenanceChecks 함수는 질의 11개가 형식 검사, 구조 숨김 검사, 출처 성격 검사를 모두 통과하는지 확인합니다(검사 2·3·4).
func TestImportedRunsPassFormatAndProvenanceChecks(t *testing.T) {
	// Given: concept-4와 concept-3의 판정 기준(두 표시 파일과 대응표)을 합친 것
	rs, err := roles.Load("c4")
	if err != nil {
		t.Fatal(err)
	}
	r3, err := roles.Load("c3")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range r3 {
		if _, dup := rs[k]; dup && k != "$" {
			t.Fatalf("두 표시 파일의 경로 %s가 겹침", k)
		}
		rs[k] = v
	}
	for _, l := range loadAll(t) {
		// 추정 목록(설계 메모 7): 순서와 인자로 맞춘 짝, procsGuess인 역할표
		inferred := InferredGroups(l.q)

		// When: 세 검사를 하면
		errs := atlas.Validate(l.result.Run)
		errs = append(errs, atlas.ScanStructure(l.result.Run, l.result.Origins, rs)...)
		errs = append(errs, atlas.CheckNature(l.result.Run, l.result.Origins, rs, inferred)...)

		// Then: 어긴 곳이 없다
		if len(errs) != 0 {
			if len(errs) > 10 {
				errs = errs[:10]
			}
			t.Fatalf("%s: %v", l.key, errs)
		}
	}
}

// TestTopicBlindFactsMatchOriginalScreens 함수는 주제의 선택 규칙(facts.json)만으로 플랫폼이 뽑은 이동·메시지·단계·시스템 콜·커널 마디가
// 원본 화면 데이터에서 바로 센 값과 차례까지 같은지 확인합니다(검사 5).
func TestTopicBlindFactsMatchOriginalScreens(t *testing.T) {
	// Given: 주제의 선택 규칙
	raw, err := os.ReadFile("../facts.json")
	if err != nil {
		t.Fatal(err)
	}
	var rules map[string][]atlas.FactRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		t.Fatal(err)
	}
	for _, l := range loadAll(t) {
		// When: 플랫폼이 규칙대로 뽑으면
		facts, err := atlas.Extract(l.result.Run, rules["journey"])
		if err != nil {
			t.Fatal(err)
		}
		got := map[string][]string{}
		for _, f := range facts {
			for _, r := range f.Rows {
				got[f.Fact] = append(got[f.Fact], rowText(r))
			}
			for _, p := range f.Pairs {
				got[f.Fact] = append(got[f.Fact], fmt.Sprintf("%s|%s→%s", p[0].Nature, p[0].Name, p[1].Name))
			}
		}

		// Then: 원본에서 바로 센 줄과 같다
		want := originalFacts(t, l.q["data"].(map[string]any), "")
		if l.c3 != nil {
			for k, v := range originalFacts(t, l.c3, "(concept-3)") {
				if k == "메시지 사본(concept-3)" {
					// copy 연결은 두 판이 같은 종류라 한 사실로 뽑힙니다(concept-4 이동의 id가 앞).
					want["메시지 사본"] = append(want["메시지 사본"], v...)
					continue
				}
				want[k] = v
			}
		}
		for fact, lines := range want {
			if !reflect.DeepEqual(got[fact], lines) {
				for i := range min(len(got[fact]), len(lines)) {
					if got[fact][i] != lines[i] {
						t.Fatalf("%s %s %d번째 줄\n뽑음 %s\n원래 %s", l.key, fact, i, got[fact][i], lines[i])
					}
				}
				t.Fatalf("%s %s: 뽑음 %d줄, 원래 %d줄", l.key, fact, len(got[fact]), len(lines))
			}
		}
		for fact := range got {
			if _, ok := want[fact]; !ok {
				t.Fatalf("%s: 원본에 없는 사실 %s", l.key, fact)
			}
		}
	}
}

// rowText 함수는 재추출한 행 하나를 견줄 글 한 줄로 씁니다. 출처 성격을 함께 적어 추정이 관측으로 나오지 않는지도 봅니다.
func rowText(r atlas.Row) string {
	ns := func(p *int64) string {
		if p == nil {
			return "-"
		}
		return strconv.FormatInt(*p, 10)
	}
	keys := make([]string, 0, len(r.Attrs))
	for k := range r.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var attrs []string
	for _, k := range keys {
		attrs = append(attrs, fmt.Sprintf("%s=%v", k[strings.Index(k, ".")+1:], r.Attrs[k]))
	}
	return fmt.Sprintf("%s|%s|%d|%s→%s|%s~%s|%s|%d|%s|%s", r.Nature, r.Name, r.Count, r.From, r.Target, ns(r.StartNS), ns(r.EndNS), ns(r.DurNS), r.Depth, r.Kind, strings.Join(attrs, ","))
}

// originalFacts 함수는 원본 한 판의 해석 결과에서 화면이 보이는 사실을 같은 꼴의 줄로 셉니다. 적재기를 거치지 않습니다.
// 이름표 규칙(칸은 "lane 이름", 파일은 "file 경로")만 형식에서 가져옵니다.
func originalFacts(t *testing.T, data map[string]any, suffix string) map[string][]string {
	t.Helper()
	us := func(v any) string {
		ns, err := atlas.MicrosToNanos(v.(json.Number))
		if err != nil {
			t.Fatal(err)
		}
		return strconv.FormatInt(ns, 10)
	}
	end := func(h map[string]any, key string) string {
		name := h[key].(string)
		if fs, ok := h["files"].([]any); ok && fileLanes[name] && (key == "from" || !fileLanes[h["from"].(string)]) {
			return "file " + fs[0].(string)
		}
		return "lane " + name
	}
	out := map[string][]string{}
	for _, raw := range data["hops"].([]any) {
		h := raw.(map[string]any)
		count, _ := strconv.Atoi(string(h["count"].(json.Number)))
		kind := atlas.Instant
		if count > 1 {
			kind = atlas.Group
		}
		// 플랫폼의 줄은 attrs 열쇠를 이름순("출처.키")으로 적으므로 journey.stage가 rec.bytes보다 앞입니다.
		attrs := []string{fmt.Sprintf("stage=%v", h["stage"])}
		if b, ok := h["bytes"]; ok && count == 1 {
			attrs = append(attrs, fmt.Sprintf("bytes=%v", b))
		}
		route := end(h, "from") + "→" + end(h, "to")
		out["이동"+suffix] = append(out["이동"+suffix], fmt.Sprintf("observed|%s|%d|%s|%s~%s|-|0|%s|%s", h["kind"], count, route, us(h["t0"]), us(h["t1"]), kind, strings.Join(attrs, ",")))
		// 버퍼 쓰기(백엔드 → 통신 버퍼)의 메시지는 뒤따른 송신 메시지의 사본이라 메시지 사건이 아니라 추정 copy 연결입니다(앞선 적재기 코드 검토 지적 6).
		bufferWrite := h["proc"] == "backend" && h["from"] == "backend" && h["to"] == "comm_buffer"
		for _, mraw := range asList(h["msgs"]) {
			m := mraw.(map[string]any)
			if bufferWrite {
				out["메시지 사본"+suffix] = append(out["메시지 사본"+suffix], fmt.Sprintf("inferred|%s→%s", h["kind"], m["type"]))
				continue
			}
			if m["cont"] == true {
				// 앞 송신에서 이어진 나머지의 길이·글은 해석기의 값이라 기록기의 len으로 뽑히지 않습니다(재검토 N1).
				out["이어진 메시지"+suffix] = append(out["이어진 메시지"+suffix], fmt.Sprintf("observed|%s|1|%s|-~-|-|0|instant|len=%v", m["type"], route, m["len"]))
				continue
			}
			out["메시지"+suffix] = append(out["메시지"+suffix], fmt.Sprintf("observed|%s|1|%s|-~-|-|0|instant|len=%v", m["type"], route, m["len"]))
		}
		if s, ok := h["sysUs"]; ok {
			if count > 1 {
				out["합친 시스템 콜"+suffix] = append(out["합친 시스템 콜"+suffix], fmt.Sprintf("derived|syscall|%d|→|-~-|%s|0|group|", count, us(s)))
			} else {
				out["시스템 콜"+suffix] = append(out["시스템 콜"+suffix], fmt.Sprintf("observed|syscall|1|→|-~-|%s|0|span|", us(s)))
			}
		}
		var walk func(n map[string]any, depth int)
		walk = func(n map[string]any, depth int) {
			out["커널 마디"+suffix] = append(out["커널 마디"+suffix], fmt.Sprintf("observed|%s|1|→|-~-|%s|%d|span|", n["fn"], us(n["us"]), depth))
			for _, k := range asList(n["kids"]) {
				walk(k.(map[string]any), depth+1)
			}
		}
		for _, k := range asList(h["kernel"]) {
			walk(k.(map[string]any), 0)
		}
	}
	for _, raw := range data["stages"].([]any) {
		s := raw.(map[string]any)
		out["단계"+suffix] = append(out["단계"+suffix], fmt.Sprintf("derived|%s|1|→|%s~%s|-|0|instant|", s["name"], us(s["t0"]), us(s["t0"])))
	}
	return out
}

// asList 함수는 없는 목록을 빈 목록으로 돌려줍니다.
func asList(v any) []any {
	l, _ := v.([]any)
	return l
}

// TestKernelGuessesAreInferred 함수는 커널 마디에 해석기가 이름으로 찾아 붙인 값이 추정 열쇠에만 있는지 봅니다(앞선 적재기 코드 검토 지적 5·9).
//
// calls·callsOf·guess와 guess인 마디의 층은 kernel-guess(inferred) 열쇠에, 위치(at)를 가진 커널 함수·프레임 장소는 kernel-guess 출처여야 합니다.
// 검토자가 센 수(calls 316, guess 31, callsOf 30, 합 377)가 그대로 옮겨졌는지도 봅니다.
func TestKernelGuessesAreInferred(t *testing.T) {
	counts := map[string]int{}
	for _, l := range loadAll(t) {
		// Given: 커널 마디가 있는 run에서
		r := l.result.Run
		nature := map[string]string{}
		for _, s := range r.Sources {
			nature[s.ID] = s.Nature
		}
		for _, e := range r.Events {
			if e.Source != srcFtrace {
				continue
			}
			// When: 추정 값의 열쇠를 보면
			for _, k := range kernelGuessKeys {
				if _, ok := e.Attrs["journey"][k]; ok {
					t.Errorf("%s 사건 %d: %s가 해석기 열쇠에 있음", l.key, e.ID, k)
				}
				if _, ok := e.Attrs[srcKGuess][k]; ok {
					counts[k]++
				}
			}
			// Then: guess인 마디의 층은 추정 열쇠에 있다
			if e.Attrs[srcKGuess]["guess"] == true {
				if _, ok := e.Attrs["journey"]["layer"]; ok {
					t.Errorf("%s 사건 %d: guess인 마디의 층이 해석기 열쇠에 있음", l.key, e.ID)
				}
			}
		}
		for _, p := range r.Places {
			if p.Kind == "frame" && p.Parent != nil && r.Places[*p.Parent].Parent != nil && r.Places[*r.Places[*p.Parent].Parent].Source == srcFtrace && nature[p.Source] != atlas.Inferred {
				t.Errorf("%s 장소 %d(%s): 위치를 가진 커널 프레임의 출처 %s", l.key, p.ID, p.Name, p.Source)
			}
		}
	}
	if counts["calls"] != 316 || counts["guess"] != 31 || counts["callsOf"] != 30 {
		t.Errorf("추정 열쇠의 수 %v, 기대 calls 316, guess 31, callsOf 30", counts)
	}
}

// TestImportRejectsStageThatDisagrees 함수는 이동의 stage가 그 시각의 처리 단계와 다른 원본을 적재기가 받지 않는지 봅니다(앞선 적재기 코드 검토 지적 11).
func TestImportRejectsStageThatDisagrees(t *testing.T) {
	// Given: 고치지 않은 질의 하나와, 첫 이동의 stage를 다른 단계 이름으로 바꾼 사본
	l := loadAll(t)[0]
	data := l.q["data"].(map[string]any)
	hop := data["hops"].([]any)[0].(map[string]any)
	hop["stage"] = "커밋"

	// When: 적재하면
	_, err := Import(l.q, l.c3, version)

	// Then: 멈춘다
	if err == nil {
		t.Fatal("stage가 다른 원본의 적재가 통과함")
	}
}

// TestExportReturnsErrorsInsteadOfPanicking 함수는 형식 검사를 통과하지만 적재기가 만들지 않는 꼴의 run에서 되돌리기가 panic 대신 오류를 돌려주는지 봅니다(앞선 적재기 코드 검토 지적 10).
func TestExportReturnsErrorsInsteadOfPanicking(t *testing.T) {
	for _, c := range []struct {
		name   string
		mutate func(r *atlas.Run)
	}{
		{"질의 구간의 끝을 비움", func(r *atlas.Run) {
			for i, e := range r.Events {
				if e.Name == "query" {
					r.Events[i].EndNS = nil
				}
			}
		}},
		{"시스템 콜을 길이 대신 시작으로", func(r *atlas.Run) {
			for i, e := range r.Events {
				if e.Source == srcStrace {
					r.Events[i].StartNS, r.Events[i].DurNS = r.Events[i].DurNS, nil
					return
				}
			}
		}},
	} {
		// Given: 시스템 콜이 있는 run을 고쳐서
		var l loaded
		for _, x := range loadAll(t) {
			if x.key == "kernel_insert" {
				l = x
			}
		}
		c.mutate(l.result.Run)

		// When: 되돌리면
		_, err := ExportC4(l.result.Run)

		// Then: panic이 아니라 오류다
		if err == nil {
			t.Errorf("%s: 오류가 아님", c.name)
		}
	}
}

// TestKernelPlaceKeysUseKernelVersion 함수는 커널 장소의 key가 PostgreSQL의 판과 무관한지 봅니다(앞선 적재기 코드 검토 지적 12, 재검토 N2).
//
// Given 커널 기록이 있는 질의를 PostgreSQL 판 둘로 적재하면, When 장소의 key를 견줄 때, Then 커널 함수·프레임의 key는 같고 PostgreSQL 함수의 key는 다릅니다.
func TestKernelPlaceKeysUseKernelVersion(t *testing.T) {
	var q, c3 map[string]any
	for _, l := range loadAll(t) {
		if l.key == "kernel_insert" {
			q, c3 = l.q, l.c3
		}
	}
	a, err := Import(q, c3, "/src/postgresql-18.6")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Import(q, c3, "/src/postgresql-99")
	if err != nil {
		t.Fatal(err)
	}
	isKernel := func(r *atlas.Run, p atlas.Place) bool {
		for q := p; q.Parent != nil; q = r.Places[*q.Parent] {
			if pp := r.Places[*q.Parent]; pp.Kind == "process" {
				return pp.Source == srcFtrace
			}
		}
		return false
	}
	kernelSame, pgDiff := 0, 0
	for i, p := range a.Run.Places {
		if p.Key == "" {
			continue
		}
		if isKernel(a.Run, p) {
			if p.Key != b.Run.Places[i].Key {
				t.Fatalf("커널 장소 %s의 key가 PostgreSQL 판에 따라 바뀜", p.Name)
			}
			kernelSame++
		} else if p.Key != b.Run.Places[i].Key {
			pgDiff++
		}
	}
	if kernelSame == 0 || pgDiff == 0 {
		t.Errorf("커널 장소 %d개, 판에 따라 바뀐 PostgreSQL 장소 %d개", kernelSame, pgDiff)
	}
}
