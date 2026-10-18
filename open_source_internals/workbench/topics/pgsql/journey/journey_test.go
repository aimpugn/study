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

// TestRoundTripRebuildsBothScreens 는 적재 결과를 파일로 쓰고 다시 읽어 되돌린 질의가 concept-4와 concept-3의 원본과 깊이 같은지 확인합니다(검사 1).
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

// TestImportedRunsPassFormatAndProvenanceChecks 는 질의 11개가 형식 검사, 구조 숨김 검사, 출처 성격 검사를 모두 통과하는지 확인합니다(검사 2·3·4).
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
		inferred := map[string]bool{MessageCopy: true, SyscallPair: true, KernelPair: true}
		if l.q["procsGuess"] == true {
			for role := range l.q["procs"].(map[string]any) {
				inferred["$.procs."+role] = true
				inferred["$.procs."+role+"[]"] = true
			}
		}

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

// TestTopicBlindFactsMatchOriginalScreens 는 주제의 선택 규칙(facts.json)만으로 플랫폼이 뽑은 이동·메시지·단계·시스템 콜·커널 마디가
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
		}

		// Then: 원본에서 바로 센 줄과 같다
		want := originalFacts(t, l.q["data"].(map[string]any), "")
		if l.c3 != nil {
			for k, v := range originalFacts(t, l.c3, "(concept-3)") {
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
		if b, ok := h["bytes"]; ok {
			attrs = append(attrs, fmt.Sprintf("bytes=%v", b))
		}
		route := end(h, "from") + "→" + end(h, "to")
		out["이동"+suffix] = append(out["이동"+suffix], fmt.Sprintf("observed|%s|%d|%s|%s~%s|-|0|%s|%s", h["kind"], count, route, us(h["t0"]), us(h["t1"]), kind, strings.Join(attrs, ",")))
		for _, mraw := range asList(h["msgs"]) {
			m := mraw.(map[string]any)
			out["메시지"+suffix] = append(out["메시지"+suffix], fmt.Sprintf("observed|%s|1|%s|-~-|-|0|instant|len=%v", m["type"], route, m["len"]))
		}
		if s, ok := h["sysUs"]; ok {
			out["시스템 콜"+suffix] = append(out["시스템 콜"+suffix], fmt.Sprintf("observed|syscall|1|→|-~-|%s|0|span|", us(s)))
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
