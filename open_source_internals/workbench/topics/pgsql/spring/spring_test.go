package spring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
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

// screen 은 이 패키지 폴더에서 Spring 화면 파일까지의 경로입니다(재고 조사 2.1절: E-03의 14벌과 값이 같음).
const screen = "../../../design/concept-9-spring.html"

// version 은 시험에 쓰는 소스 판입니다. 실습의 workbench.json의 sourceRoot와 같습니다.
const version = "/src/postgresql-18.6"

// loaded 값은 기록 한 벌의 원본과 적재 결과입니다. raw는 적재에 넘긴 것과 따로 다시 읽은 원본이라,
// 적재기가 넘겨받은 객체를 고쳐도 되돌림 비교가 그 고친 값과 견주지 않습니다.
type loaded struct {
	key    string
	raw    map[string]any
	result Result
}

// freshRuns 함수는 화면 파일에서 RUNS를 새로 읽고 열쇠를 차례로 돌려줍니다.
func freshRuns(t *testing.T) (map[string]any, []string) {
	t.Helper()
	b, err := os.ReadFile(screen)
	if err != nil {
		t.Fatal(err)
	}
	runs, err := ReadRuns(string(b))
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(runs))
	for k := range runs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return runs, keys
}

// loadAll 함수는 Spring 기록 14벌을 열쇠 차례로 적재합니다. 하나라도 실패하면 시험을 멈춥니다.
func loadAll(t *testing.T) []loaded {
	t.Helper()
	runs, keys := freshRuns(t)
	fresh, _ := freshRuns(t)
	out := make([]loaded, 0, len(keys))
	for _, k := range keys {
		res, err := Import(k, runs[k].(map[string]any), version)
		if err != nil {
			t.Fatalf("%s: %v", k, err)
		}
		out = append(out, loaded{key: k, raw: fresh[k].(map[string]any), result: res})
	}
	if len(out) != 14 {
		t.Fatalf("기록 %d벌(기대 14)", len(out))
	}
	return out
}

// find 함수는 적재한 기록 가운데 열쇠가 key인 것을 돌려줍니다.
func find(t *testing.T, all []loaded, key string) loaded {
	t.Helper()
	for _, l := range all {
		if l.key == key {
			return l
		}
	}
	t.Fatalf("기록 %s 없음", key)
	return loaded{}
}

// copyRun 함수는 run을 파일 꼴로 썼다가 다시 읽어 깊은 사본을 만듭니다. 변이 시험이 다른 시험의 적재 결과를 건드리지 않게 합니다.
func copyRun(t *testing.T, r *atlas.Run) *atlas.Run {
	t.Helper()
	var buf bytes.Buffer
	if err := atlas.Write(&buf, r); err != nil {
		t.Fatal(err)
	}
	back, err := atlas.Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	return back
}

// differs 함수는 원본과 되돌린 값을 같은 꼴로 바꿔 견준 차이를 돌려줍니다.
func differs(t *testing.T, want, got any) []string {
	t.Helper()
	w, err := load.Canonical(want)
	if err != nil {
		t.Fatal(err)
	}
	g, err := load.Canonical(got)
	if err != nil {
		t.Fatal(err)
	}
	return atlas.Diff("$", w, g)
}

// TestRoundTripRebuildsEveryRun 함수는 적재 결과를 파일 꼴로 쓰고 다시 읽어 되돌린 기록이 따로 읽은 원본과 깊이 같은지 확인합니다(검사 1).
func TestRoundTripRebuildsEveryRun(t *testing.T) {
	for _, l := range loadAll(t) {
		// Given: 적재 결과를 형식의 JSON으로 쓰고 다시 읽은 것
		back := copyRun(t, l.result.Run)

		// When: 그 JSON만 읽어 원본 꼴로 되돌리면
		got, err := Export(back)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}

		// Then: 원본과 차이가 없다
		if diff := differs(t, l.raw, got); len(diff) != 0 {
			t.Fatalf("%s: 원본과 다름 %v", l.key, diff[:min(10, len(diff))])
		}
	}
}

// TestImportedRunsPassFormatAndProvenanceChecks 함수는 14벌이 형식 검사, 구조 숨김 검사, 출처 성격 검사를 모두 통과하는지 확인합니다(검사 2·3·4).
func TestImportedRunsPassFormatAndProvenanceChecks(t *testing.T) {
	// Given: Spring의 판정 기준(표시 파일과 대응표)과 추정 목록
	rs, err := roles.Load("e03")
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range loadAll(t) {
		// When: 세 검사를 하면
		errs := atlas.Validate(l.result.Run)
		errs = append(errs, atlas.ScanStructure(l.result.Run, l.result.Origins, rs)...)
		errs = append(errs, atlas.CheckNature(l.result.Run, l.result.Origins, rs, InferredGroups())...)

		// Then: 어긴 곳이 없다
		if len(errs) != 0 {
			t.Fatalf("%s: %v", l.key, errs[:min(10, len(errs))])
		}
	}
}

// linkOf 함수는 종류가 kind인 첫 연결의 차례를 돌려줍니다. 출발 사건의 출처를 src로 좁힐 수 있습니다(빈 글이면 모두).
func linkOf(t *testing.T, r *atlas.Run, kind, src string) int {
	t.Helper()
	for i, l := range r.Links {
		if l.Kind == kind && (src == "" || r.Events[l.Src].Source == src) {
			return i
		}
	}
	t.Fatalf("%s에서 나가는 %s 연결 없음", src, kind)
	return 0
}

// eventOf 함수는 출처가 src이고 이름이 name인 사건 가운데 skip번째(0부터)를 돌려줍니다.
func eventOf(t *testing.T, r *atlas.Run, src, name string, skip int) int {
	t.Helper()
	for _, e := range r.Events {
		if e.Source == src && e.Name == name {
			if skip == 0 {
				return e.ID
			}
			skip--
		}
	}
	t.Fatalf("%s의 %s 사건이 모자람", src, name)
	return 0
}

// TestRoundTripCatchesBrokenStructure 함수는 구조 칸을 하나씩 망가뜨린 적재에서 되돌린 기록이 원본과 달라지거나 되돌리기가 실패하는지 확인합니다.
// 되돌림 비교가 연결과 장소로 다시 세는 값(tCapture·afterStep, 연결의 pid, 요청의 칸, wire 효과, 묶임의 by, next, 겹친 메시지, 트리 부모,
// 빌린 연결, 창이 품은 서버 사건, 관찰이 품은 읽기 구간)을 실제로 보는지 판별력을 봅니다.
func TestRoundTripCatchesBrokenStructure(t *testing.T) {
	// Given: 연결이 둘이고 기다림·묶임이 있는 기록과, 겹친 메시지가 있는 기록
	all := loadAll(t)
	row := find(t, all, "5-requires-new-same-row/a")
	pool := find(t, all, "5-requires-new-pool1/a")
	mutations := []struct {
		name   string
		base   loaded
		mutate func(r *atlas.Run)
	}{
		{"멈춤 자리 → 관찰 same 연결을 다음 관찰로", row, func(r *atlas.Run) {
			i := linkOf(t, r, "same", srcTxlab)
			r.Links[i].Dst = eventOf(t, r, srcObserve, nameCapture, 1)
		}},
		{"메시지의 백엔드 쪽 끝을 다른 연결의 백엔드로", row, func(r *atlas.Run) {
			other := r.Events[eventOf(t, r, srcWireMac, "Q", 0)].Target
			for i, e := range r.Events {
				if e.Source == srcWireMac && e.Target != nil && *e.Target != *other {
					r.Events[i].Target = other
					return
				}
			}
		}},
		{"요청이 품은 첫 메시지를 뺌", row, func(r *atlas.Run) { r.Links[linkOf(t, r, "contains", srcMac)].Kind = "gone" }},
		{"wire 효과를 다른 요청으로", row, func(r *atlas.Run) {
			i := linkOf(t, r, "wire", "")
			r.Links[i].Dst = eventOf(t, r, srcMac, nameRequest, 1)
		}},
		{"묶임 상태의 by를 다른 마디로", row, func(r *atlas.Run) {
			i := linkOf(t, r, "by", "")
			r.Links[i].Dst = eventOf(t, r, srcTree, "org.springframework.transaction.interceptor.TransactionInterceptor::invoke", 0)
		}},
		{"트랜잭션의 next를 앞 메시지로", row, func(r *atlas.Run) { r.Links[linkOf(t, r, "next", "")].Dst -= 2 }},
		{"겹친 메시지를 다른 메시지로", pool, func(r *atlas.Run) { r.Links[linkOf(t, r, "overlap", "")].Dst += 2 }},
		{"트리 마디의 부모를 할아버지로", row, func(r *atlas.Run) {
			for i, e := range r.Events {
				if e.Source == srcTree && e.Parent != nil && r.Events[*e.Parent].Parent != nil {
					r.Events[i].Parent = r.Events[*e.Parent].Parent
					return
				}
			}
		}},
		{"빌린 연결을 지움", row, func(r *atlas.Run) { r.Events[eventOf(t, r, srcGuess, nameBorrow, 0)].Target = nil }},
		{"창이 품은 서버 사건을 다른 창으로", row, func(r *atlas.Run) {
			// 창의 첫 contains는 카탈로그 sample이라, 기록기의 서버 사건(sample이 아닌 것)을 품은 연결을 고르고 다른 창으로 옮깁니다.
			for i, l := range r.Links {
				if d := r.Events[l.Dst]; l.Kind == "contains" && r.Events[l.Src].Name == nameServer && d.Source == srcRec && d.Kind != atlas.Sample {
					for k := 0; ; k++ {
						if w := eventOf(t, r, srcCont, nameServer, k); w != l.Src {
							r.Links[i].Src = w
							return
						}
					}
				}
			}
			t.Fatal("서버 창이 품은 서버 사건이 없음")
		}},
		{"관찰이 품은 읽기 구간을 다음 관찰로", row, func(r *atlas.Run) {
			i := linkOf(t, r, "contains", srcObserve)
			r.Links[i].Src = eventOf(t, r, srcObserve, nameCapture, 1)
		}},
		{"관찰의 끝을 다른 관찰의 끝으로", row, func(r *atlas.Run) { r.Links[linkOf(t, r, "end", "")].Dst = eventOf(t, r, srcCont, nameCaptureEnd, 1) }},
		{"JFR 구간의 부른 쪽 프레임을 지움", row, func(r *atlas.Run) {
			r.Events[eventOf(t, r, srcJFR, "org.postgresql.core.v3.QueryExecutorImpl::execute", 0)].Stack = nil
		}},
		{"BackendKeyData의 백엔드를 다른 연결의 것으로", row, func(r *atlas.Run) {
			k := eventOf(t, r, srcWireMac, "K", 0)
			other := r.Events[eventOf(t, r, srcWireMac, "K", 1)].From
			r.Events[k].From = other
		}},
	}
	for _, m := range mutations {
		// When: 사본을 망가뜨려 되돌리면
		r := copyRun(t, m.base.result.Run)
		m.mutate(r)
		got, err := Export(r)

		// Then: 되돌리기가 실패하거나 원본과 다르다
		if err == nil && len(differs(t, m.base.raw, got)) == 0 {
			t.Errorf("%s: 원본과 같게 되돌아옴", m.name)
		}
	}
}

// TestChecksCatchHiddenStructureAndWrongNature 함수는 구조를 attrs에 숨기거나 출처 성격을 바꾼 적재를 검사 3·4가 잡는지 확인합니다.
func TestChecksCatchHiddenStructureAndWrongNature(t *testing.T) {
	// Given: 판정 기준과 연결이 둘인 기록
	rs, err := roles.Load("e03")
	if err != nil {
		t.Fatal(err)
	}
	all := loadAll(t)
	base := find(t, all, "5-requires-new-same-row/a")
	msg := eventOf(t, base.result.Run, srcWireMac, "Q", 0)
	life := eventOf(t, base.result.Run, srcTxlab, nameConn, 0)

	// When/Then: 메시지의 attrs에 시각 t를 남기면 구조 숨김 검사가 실패한다
	r := copyRun(t, base.result.Run)
	r.Events[msg].Attrs[srcWireMac]["t"] = json.Number("1")
	if errs := atlas.ScanStructure(r, base.result.Origins, rs); len(errs) != 1 || !strings.Contains(errs[0].Error(), "구조 칸") {
		t.Fatalf("시각을 잡아야 함: %v", errs)
	}

	// When/Then: 연결의 attrs에 원본의 요청 목록을 통째로 넣으면 그 아래의 구조 역할 때문에 실패한다
	r = copyRun(t, base.result.Run)
	r.Events[life].Attrs[srcSpring]["requests"] = base.raw["conns"].([]any)[0].(map[string]any)["requests"]
	if errs := atlas.ScanStructure(r, base.result.Origins, rs); len(errs) == 0 {
		t.Fatal("원본 조각을 잡아야 함")
	}

	// When/Then: wire 효과의 근거를 관찰 출처로 바꾸면 추정 목록 때문에 출처 성격 검사가 실패한다
	r = copyRun(t, base.result.Run)
	r.Links[linkOf(t, r, "wire", "")].Basis = srcWireMac
	if errs := atlas.CheckNature(r, base.result.Origins, rs, InferredGroups()); len(errs) != 1 || !strings.Contains(errs[0].Error(), "effects[].seq") {
		t.Fatalf("wire 근거를 잡아야 함: %v", errs)
	}

	// When/Then: 추정 출처의 성격을 도출로 바꾸면 빌림·묶임·슬롯·묶임 효과가 모두 실패한다
	r = copyRun(t, base.result.Run)
	for i, s := range r.Sources {
		if s.ID == srcGuess {
			r.Sources[i].Nature = atlas.Derived
		}
	}
	if errs := atlas.CheckNature(r, base.result.Origins, rs, InferredGroups()); len(errs) == 0 {
		t.Fatal("추정을 도출로 바꾼 것을 잡아야 함")
	}

	// When/Then: 요청(해석기가 정한 묶음)의 출처를 맥 쪽 중계기(관찰)로 바꾸면 실패한다
	r = copyRun(t, base.result.Run)
	r.Events[eventOf(t, r, srcMac, nameRequest, 0)].Source = srcWireMac
	if errs := atlas.CheckNature(r, base.result.Origins, rs, InferredGroups()); len(errs) != 1 {
		t.Fatalf("요청 출처를 잡아야 함: %v", errs)
	}
}

// rowText 함수는 재추출한 행 하나를 견줄 글 한 줄로 씁니다: 출처 성격|이름|장소|출발→도착|시작~끝|종류|attrs.
func rowText(r atlas.Row) string {
	keys := make([]string, 0, len(r.Attrs))
	for k := range r.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attrs := make([]string, 0, len(keys))
	for _, k := range keys {
		attrs = append(attrs, fmt.Sprintf("%s=%v", k[strings.Index(k, ".")+1:], r.Attrs[k]))
	}
	return fmt.Sprintf("%s|%s|%s|%s→%s|%s~%s|%s|%s", r.Nature, r.Name, r.Place, r.From, r.Target, nsText(r.StartNS), nsText(r.EndNS), r.Kind, strings.Join(attrs, ","))
}

// nsText 함수는 시각을 글로 씁니다. 없으면 "-"입니다.
func nsText(p *int64) string {
	if p == nil {
		return "-"
	}
	return strconv.FormatInt(*p, 10)
}

// factsOf 함수는 규칙대로 뽑은 사실을 사실 이름마다 줄 목록으로 씁니다.
func factsOf(t *testing.T, r *atlas.Run, rules []atlas.FactRule) map[string][]string {
	t.Helper()
	facts, err := atlas.Extract(r, rules)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][]string{}
	for _, f := range facts {
		for _, row := range f.Rows {
			got[f.Fact] = append(got[f.Fact], rowText(row))
		}
	}
	return got
}

// springRules 함수는 주제의 선택 규칙 가운데 Spring 목록을 읽습니다.
func springRules(t *testing.T) []atlas.FactRule {
	t.Helper()
	raw, err := os.ReadFile("../facts.json")
	if err != nil {
		t.Fatal(err)
	}
	var rules map[string][]atlas.FactRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		t.Fatal(err)
	}
	return rules["spring"]
}

// factNames 는 화면이 보이는 사실 가운데 재추출로 견주는 것입니다(설계 메모 9의 검사 5, 브리프의 Spring 목록).
var factNames = []string{"요청", "트리 마디", "메시지", "기다림", "서버 기다림", "막힘 관찰", "JFR 잠듦", "결과 값"}

// TestTopicBlindFactsMatchOriginalScreens 함수는 주제의 선택 규칙(facts.json)만으로 플랫폼이 뽑은 요청·트리 마디·메시지·기다림·막힘 관찰·
// JFR 잠듦·결과 값이 원본 화면 데이터에서 바로 센 줄과 차례까지 같은지 확인합니다(검사 5). 출처 성격도 함께 견줘 추정이 관측으로 나오지 않는지 봅니다.
func TestTopicBlindFactsMatchOriginalScreens(t *testing.T) {
	// Given: 주제의 선택 규칙
	rules := springRules(t)
	for _, l := range loadAll(t) {
		// When: 플랫폼이 규칙대로 뽑으면
		got := factsOf(t, l.result.Run, rules)

		// Then: 원본에서 바로 센 줄과 같다
		want := originalFacts(t, l.raw)
		for _, fact := range factNames {
			if !reflect.DeepEqual(got[fact], want[fact]) {
				for i := range min(len(got[fact]), len(want[fact])) {
					if got[fact][i] != want[fact][i] {
						t.Fatalf("%s %s %d번째 줄\n뽑음 %.300s\n원래 %.300s", l.key, fact, i, got[fact][i], want[fact][i])
					}
				}
				t.Fatalf("%s %s: 뽑음 %d줄, 원래 %d줄", l.key, fact, len(got[fact]), len(want[fact]))
			}
		}
	}
}

// TestBroadenedFactRuleIsCaught 함수는 재추출 규칙을 넓히면(트리 마디 규칙에서 종류 조건을 빼 해석기 효과의 순간까지 고름) 원본과 달라지는지
// 확인합니다. 재추출 비교가 규칙의 범위를 실제로 판정하는지 봅니다.
func TestBroadenedFactRuleIsCaught(t *testing.T) {
	// Given: 트리 마디 규칙의 kinds를 뺀 규칙
	rules := springRules(t)
	for i := range rules {
		if rules[i].Fact == "트리 마디" {
			s := *rules[i].Select
			s.Kinds = nil
			rules[i].Select = &s
		}
	}
	l := find(t, loadAll(t), "5-nested/a")

	// When: 넓힌 규칙으로 뽑으면
	got := factsOf(t, l.result.Run, rules)

	// Then: 원본에서 센 트리 마디와 다르다
	if reflect.DeepEqual(got["트리 마디"], originalFacts(t, l.raw)["트리 마디"]) {
		t.Fatal("넓힌 규칙이 원본과 같은 줄을 냄")
	}
}

// originalFacts 함수는 원본 한 벌에서 화면이 보이는 사실을 rowText와 같은 꼴의 줄로 셉니다. 적재기를 거치지 않고,
// 장소 이름표 규칙(종류 이름@경로:줄)과 자바 프레임의 위치(경로 없음, 줄 −1은 위치 없음)만 형식에서 가져옵니다.
func originalFacts(t *testing.T, run map[string]any) map[string][]string {
	t.Helper()
	us := func(v any) string {
		ns, err := atlas.MicrosToNanos(v.(json.Number))
		if err != nil {
			t.Fatal(err)
		}
		return strconv.FormatInt(ns, 10)
	}
	attrs := func(m map[string]any, keys ...string) string {
		sort.Strings(keys)
		var out []string
		for _, k := range keys {
			if v, ok := m[k]; ok {
				out = append(out, fmt.Sprintf("%s=%v", k, v))
			}
		}
		return strings.Join(out, ",")
	}
	pgFrame := func(f map[string]any) string {
		if f["at"] == "?" {
			return "frame " + f["fn"].(string)
		}
		return "frame " + f["fn"].(string) + "@" + f["at"].(string)
	}
	javaFrame := func(ref string) string {
		i := strings.LastIndexByte(ref, ':')
		if ref[i+1:] == "-1" {
			return "frame " + ref[:i]
		}
		return "frame " + ref[:i] + "@:" + ref[i+1:]
	}
	pidOf := map[string]string{}
	out := map[string][]string{}
	add := func(fact, line string) { out[fact] = append(out[fact], line) }
	for _, raw := range run["conns"].([]any) {
		c := raw.(map[string]any)
		name, pid := c["name"].(string), string(c["pid"].(json.Number))
		pidOf[name] = pid
		for _, mr := range c["messages"].([]any) {
			m := mr.(map[string]any)
			ends := "process " + pid + "→"
			if m["dir"] == "F" {
				ends = "→process " + pid
			}
			add("메시지", fmt.Sprintf("observed|%s|connection %s|%s|%s~-|instant|%s", m["type"], name, ends, us(m["t"]), attrs(m, "len", "text")))
		}
		for _, qr := range c["requests"].([]any) {
			q := qr.(map[string]any)
			add("요청", fmt.Sprintf("derived|request|connection %s|→|%s~%s|span|", name, us(q["sent"]), us(q["done"])))
			s, ok := q["server"].(map[string]any)
			if !ok {
				continue
			}
			for _, er := range s["events"].([]any) {
				e := er.(map[string]any)
				if e["kind"] != "xact_wait" {
					continue
				}
				place, kind, end := "process "+pid, "instant", "-"
				if st, ok := e["stack"].([]any); ok {
					place = pgFrame(st[len(st)-1].(map[string]any))
				}
				if t1, ok := e["t1"]; ok {
					kind, end = "span", us(t1)
				}
				add("서버 기다림", fmt.Sprintf("observed|xact_wait|%s|→|%s~%s|%s|", place, us(e["t"]), end, kind))
			}
		}
	}
	view := run["view"].(map[string]any)
	for _, nr := range view["tree"].([]any) {
		n := nr.(map[string]any)
		start, end := "-", "-"
		if n["t1"] != nil {
			start, end = us(n["t0"]), us(n["t1"])
		}
		add("트리 마디", fmt.Sprintf("derived|%s|descriptor %s|→|%s~%s|span|%s", n["fn"], n["desc"], start, end, attrs(n, "label", "end")))
	}
	for _, wr := range view["waits"].([]any) {
		w := wr.(map[string]any)
		add("기다림", fmt.Sprintf("derived|wait|process %s|→|%s~%s|span|%s", pidOf[w["conn"].(string)], us(w["t0"]), us(w["t1"]), attrs(w, "xid", "end")))
	}
	for _, br := range run["blocked"].([]any) {
		b := br.(map[string]any)
		target := ""
		if ws := b["waiters"].([]any); len(ws) == 1 {
			target = "process " + pidOf[ws[0].(string)]
		}
		add("막힘 관찰", fmt.Sprintf("observed|blocked|client txlab|→%s|%s~-|instant|", target, us(b["t"])))
	}
	for _, pr := range run["java"].(map[string]any)["parks"].([]any) {
		p := pr.(map[string]any)
		at := p["at"].([]any)
		add("JFR 잠듦", fmt.Sprintf("observed|park|%s|→|%s~%s|span|%s", javaFrame(at[0].(string)), us(p["t0"]), us(p["t1"]), attrs(p, "under")))
	}
	for _, cr := range run["captures"].([]any) {
		c := cr.(map[string]any)
		add("결과 값", fmt.Sprintf("observed|capture|client txlab|→|%s~-|sample|%s", us(c["t"]), attrs(c, "label", "pages", "visible")))
	}
	return out
}

// TestChromeExportHasNoCrossingSpansPerClock 함수는 적재 결과를 Chrome 형식으로 옮겼을 때 시계마다 같은 장소의 구간이 서로 어긋나 겹치지 않는지
// 확인합니다(검사 2). 시계 셋(mac, container, jfr)마다 포개진 구간 하나를 어긋나게 만들면 검사가 그 구간을 잡아야 합니다(볼 것이 실제로 있었는지).
func TestChromeExportHasNoCrossingSpansPerClock(t *testing.T) {
	// Given: 기록 14벌
	all := loadAll(t)

	// When/Then: Chrome JSON으로 쓸 수 있고, 옮기지 못한 사건은 시작 시각이 없는 것뿐이며, 어긋난 구간이 없다
	for _, l := range all {
		_, skipped, err := atlas.ChromeJSON(l.result.Run)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}
		for why := range skipped {
			if why != "시작 시각 없음" {
				t.Errorf("%s: 옮기지 못한 까닭 %q", l.key, why)
			}
		}
		if ov := atlas.CrossingSpans(l.result.Run); len(ov) != 0 {
			t.Errorf("%s: 어긋난 구간 %d쌍, 첫 쌍 %+v", l.key, len(ov), ov[0])
		}
	}

	// When: 시계마다 포개진 구간이 있는 첫 기록의 사본에서 안쪽 구간의 끝을 바깥 구간의 끝 너머로 민다
	for _, clock := range []string{clockMac, clockCont, clockJFR} {
		caught := false
		for _, l := range all {
			outer, inner, ok := nestedPair(l.result.Run, clock)
			if !ok {
				continue
			}
			r := copyRun(t, l.result.Run)
			end := *r.Events[outer].EndNS + 1
			r.Events[inner].EndNS = &end

			// Then: 그 구간이 어긋난 쌍으로 잡힌다
			for _, o := range atlas.CrossingSpans(r) {
				if o.B == inner || o.A == inner {
					caught = true
				}
			}
			if !caught {
				t.Fatalf("%s 시계 %s: 사건 %d를 사건 %d 밖으로 밀었는데 어긋남을 못 찾음", clock, l.key, inner, outer)
			}
			break
		}
		if !caught {
			t.Fatalf("%s 시계에 포개진 구간이 있는 기록이 없음: 어긋남 검사가 볼 것이 없었음", clock)
		}
	}
}

// nestedPair 함수는 시계가 clock인 출처의 사건 가운데 같은 장소에서 끝이 있는 구간이 다른 구간 안에 늦게 시작해 일찍 끝나는 첫 쌍을 찾습니다.
func nestedPair(r *atlas.Run, clock string) (outer, inner int, ok bool) {
	clockOf := map[string]string{}
	for _, s := range r.Sources {
		clockOf[s.ID] = s.Clock
	}
	timed := func(e atlas.Event) bool {
		return e.Kind == atlas.Span && e.StartNS != nil && e.EndNS != nil && clockOf[e.Source] == clock
	}
	for _, a := range r.Events {
		if !timed(a) {
			continue
		}
		for _, b := range r.Events {
			if timed(b) && a.Place == b.Place && *a.StartNS < *b.StartNS && *b.EndNS < *a.EndNS {
				return a.ID, b.ID, true
			}
		}
	}
	return 0, 0, false
}

// TestImportRejectsShapesItCannotRebuild 함수는 규칙으로 되살릴 수 없는 꼴이나 다시 세는 칸이 규칙과 다른 원본을 적재기가 받지 않는지 확인합니다.
// "되돌림 대조"가 붙은 것은 적재기가 끝에 하는 대조(내보낸 기록 = 원본)가 다시 세는 칸의 어긋남을 잡는지 봅니다. want가 있으면 그 까닭으로
// 멈춰야 합니다(다른 검사가 우연히 멈춘 것이 아님).
func TestImportRejectsShapesItCannotRebuild(t *testing.T) {
	cases := []struct {
		name, key, want string
		mutate          func(run map[string]any)
	}{
		{"끝이 맥 시계인 기다림(오류 없이 끝남)", "5-requires-new-same-row/a", "", func(run map[string]any) {
			view := run["view"].(map[string]any)
			view["waits"].([]any)[0].(map[string]any)["end"] = "done"
		}},
		{"기다린 백엔드가 둘인 막힘 관찰", "5-requires-new-same-row/a", "", func(run map[string]any) {
			run["blocked"].([]any)[0].(map[string]any)["waiters"] = []any{"C1", "C2"}
		}},
		{"맞는 길이 없는 갈림", "5-nested/a", "", func(run map[string]any) {
			view := run["view"].(map[string]any)
			n := view["tree"].([]any)[0].(map[string]any)
			view["branches"].(map[string]any)[n["fn"].(string)+n["desc"].(string)].(map[string]any)["alts"] = []any{}
		}},
		{"관찰 시각이 아닌 span[0]", "5-nested/a", "", func(run map[string]any) {
			run["captures"].([]any)[0].(map[string]any)["span"].([]any)[0] = json.Number("1")
		}},
		{"어느 메시지와도 같지 않은 겹친 메시지", "5-requires-new-pool1/a", "", func(run map[string]any) {
			for _, c := range run["captures"].([]any) {
				if o := c.(map[string]any)["overlap"].([]any); len(o) > 0 {
					o[0].(map[string]any)["len"] = json.Number("99999")
					return
				}
			}
		}},
		{"효과의 차례가 바뀐 마디", "5-requires-new-pool1/a", "", func(run map[string]any) {
			for _, n := range run["view"].(map[string]any)["tree"].([]any) {
				if e := n.(map[string]any)["effects"].([]any); len(e) == 2 {
					e[0], e[1] = e[1], e[0]
					return
				}
			}
		}},
		{"null이 아닌 서버 창의 in", "5-nested/a", "", func(run map[string]any) {
			q := run["conns"].([]any)[0].(map[string]any)["requests"].([]any)[1].(map[string]any)
			q["server"].(map[string]any)["in"] = []any{map[string]any{"type": "Q", "len": json.Number("5")}}
		}},
		{"BackendKeyData가 아닌 메시지의 f.pid", "5-nested/a", "", func(run map[string]any) {
			for _, m := range run["conns"].([]any)[0].(map[string]any)["messages"].([]any) {
				if mm := m.(map[string]any); mm["type"] == "R" {
					mm["f"].(map[string]any)["pid"] = "1"
					return
				}
			}
		}},
		{"규칙과 다른 요청의 보낸 시각", "5-nested/a", "", func(run map[string]any) {
			q := run["conns"].([]any)[0].(map[string]any)["requests"].([]any)[1].(map[string]any)
			q["sent"] = json.Number("1")
		}},
		{"규칙과 다른 연결을 연 스레드(되돌림 대조)", "5-nested/a", "", func(run map[string]any) {
			run["conns"].([]any)[0].(map[string]any)["thread"] = "nobody"
		}},
		{"규칙과 다른 떼어 둔 연결(되돌림 대조)", "5-requires-new-same-row/a", "", func(run map[string]any) {
			for _, b := range run["view"].(map[string]any)["binding"].([]any)[1:] {
				if h := b.(map[string]any)["held"].([]any); len(h) > 0 {
					b.(map[string]any)["held"] = []any{}
					return
				}
			}
		}},
		{"규칙과 다른 JFR 구간의 depth(되돌림 대조)", "5-nested/a", "되돌린 기록이 원본과 다름", func(run map[string]any) {
			s := run["java"].(map[string]any)["spans"].([]any)[0].(map[string]any)
			s["depth"] = json.Number("7")
		}},
		{"스택 1번 칸이 아닌 hint의 f.at", "5-requires-new-same-row/a", "스택 1번 칸의 자리가 아님", func(run map[string]any) {
			serverEvents(run, func(e map[string]any) bool {
				if e["kind"] != "hint" {
					return false
				}
				e["f"].(map[string]any)["at"] = "src/x.c:1"
				return true
			})
		}},
		{"서버 사건 f의 모르는 열쇠", "5-requires-new-same-row/a", "모르는 열쇠", func(run map[string]any) {
			serverEvents(run, func(e map[string]any) bool {
				f, ok := e["f"].(map[string]any)
				if ok {
					f["conn"] = "C2"
				}
				return ok
			})
		}},
		{"hint·ereport가 아닌 서버 사건의 f.at", "5-requires-new-same-row/a", "hint·ereport에만", func(run map[string]any) {
			serverEvents(run, func(e map[string]any) bool {
				f, ok := e["f"].(map[string]any)
				if ok && e["kind"] == "lock" {
					f["at"] = "src/x.c:1"
					return true
				}
				return false
			})
		}},
		{"규칙과 다른 트랜잭션의 first", "5-nested/a", "요청에서 다시 센", func(run map[string]any) {
			x := run["transactions"].([]any)[0].(map[string]any)
			x["first"] = json.Number(fmt.Sprint(mustInt(t, x["first"]) + 1))
		}},
		{"규칙과 다른 트랜잭션의 end", "5-nested/a", "요청에서 다시 센", func(run map[string]any) {
			run["transactions"].([]any)[0].(map[string]any)["end"] = "ROLLBACK"
		}},
		{"규칙과 다른 접속 인증 차례(되돌림 대조)", "5-nested/a", "되돌린 기록이 원본과 다름", func(run map[string]any) {
			run["conns"].([]any)[0].(map[string]any)["startup"].([]any)[0] = "ssl(N)"
		}},
		{"규칙과 다른 마지막 맥 쪽 메시지(되돌림 대조)", "5-nested/a", "되돌린 기록이 원본과 다름", func(run map[string]any) {
			run["conns"].([]any)[0].(map[string]any)["lastMac"] = "Q"
		}},
		{"메시지 짝과 다른 연결의 시계 구간", "5-nested/a", "다시 센 구간", func(run map[string]any) {
			run["clock"].(map[string]any)["wire"].([]any)[0].(map[string]any)["lo"] = json.Number("-1")
		}},
		{"규칙과 다른 컨테이너 쪽 마지막 메시지(되돌림 대조)", "5-nested/a", "되돌린 기록이 원본과 다름", func(run map[string]any) {
			run["conns"].([]any)[0].(map[string]any)["lastCont"] = "Q"
		}},
		{"요청 수와 다른 join의 Z 수", "5-nested/a", "요청 수", func(run map[string]any) {
			run["conns"].([]any)[0].(map[string]any)["join"] = []any{json.Number("3"), json.Number("3")}
		}},
		{"ReadyForQuery 입구 수와 Z 수가 다른 연결", "5-nested/a", "차례 짝을 믿을 수 없음", func(run map[string]any) {
			run["conns"].([]any)[0].(map[string]any)["join"] = []any{json.Number("3"), json.Number("4")}
		}},
	}
	for _, c := range cases {
		// Given: 원본을 화면 파일에서 새로 읽어 한 칸을 고친 것
		runs, _ := freshRuns(t)
		run := runs[c.key].(map[string]any)
		if _, err := Import(c.key, run, version); err != nil {
			t.Fatalf("%s: 고치지 않은 원본은 적재되어야 함: %v", c.key, err)
		}
		c.mutate(run)

		// When: 적재하면
		_, err := Import(c.key, run, version)

		// Then: 적재기가 멈춘다(want가 있으면 그 까닭으로)
		switch {
		case err == nil:
			t.Errorf("%s: %s 적재가 통과함", c.name, c.key)
		case c.want != "" && !strings.Contains(err.Error(), c.want):
			t.Errorf("%s: 다른 까닭으로 멈춤: %v", c.name, err)
		}
	}
}

// serverEvents 함수는 원본의 서버 창 사건마다 f를 부르고, f가 참을 돌려주면 멈춥니다. 한 번도 참이 아니면 시험을 멈춥니다.
func serverEvents(run map[string]any, f func(e map[string]any) bool) {
	for _, c := range run["conns"].([]any) {
		for _, q := range c.(map[string]any)["requests"].([]any) {
			s, _ := q.(map[string]any)["server"].(map[string]any)
			evs, _ := s["events"].([]any)
			for _, e := range evs {
				if f(e.(map[string]any)) {
					return
				}
			}
		}
	}
	panic("고칠 서버 사건이 없음")
}

// mustInt 함수는 원본의 수 하나를 정수로 읽습니다. 수가 아니면 시험을 멈춥니다.
func mustInt(t *testing.T, v any) int {
	t.Helper()
	n, err := strconv.Atoi(fmt.Sprint(v))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// TestExportReturnsErrorsInsteadOfPanicking 함수는 형식 검사를 통과하지만 적재기가 만들지 않는 꼴(끝이 빈 빌림·서버 창, 스택이 빈 트리 뿌리)의
// run에서 Export가 panic하지 않고 오류를 돌려주는지 확인합니다(코드 검토 지적 9).
func TestExportReturnsErrorsInsteadOfPanicking(t *testing.T) {
	// Given: 빌림·서버 창·트리가 있는 기록
	l := find(t, loadAll(t), "5-requires-new-same-row/a")
	cases := []struct {
		name   string
		mutate func(r *atlas.Run)
	}{
		{"빌림의 끝을 비움", func(r *atlas.Run) { r.Events[eventOf(t, r, srcGuess, nameBorrow, 0)].EndNS = nil }},
		{"서버 창의 끝을 비움", func(r *atlas.Run) { r.Events[eventOf(t, r, srcCont, nameServer, 0)].EndNS = nil }},
		{"트리 뿌리의 스택을 비움", func(r *atlas.Run) {
			for i, e := range r.Events {
				if e.Source == srcTree && e.Kind == atlas.Span && e.Parent == nil {
					r.Events[i].Stack = nil
					return
				}
			}
		}},
	}
	for _, c := range cases {
		r := copyRun(t, l.result.Run)
		c.mutate(r)

		// When: 되돌리면
		var err error
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("%s: panic %v", c.name, p)
				}
			}()
			_, err = Export(r)
		}()

		// Then: 오류를 돌려준다
		if err == nil {
			t.Errorf("%s: 오류 없이 되돌아옴", c.name)
		}
	}
}

// TestExportIgnoresLinkTableOrder 함수는 연결 표의 차례만 뒤집은 run(사건·장소·연결의 내용은 같음)도 원본과 같게 되돌아오는지 확인합니다.
// 형식은 연결 표의 차례에 뜻을 두지 않으므로 내보내기가 그 차례에 기대면 안 됩니다(코드 검토 지적 8).
func TestExportIgnoresLinkTableOrder(t *testing.T) {
	for _, l := range loadAll(t) {
		// Given: 연결 표를 거꾸로 놓고 id를 다시 매긴 사본(형식 검사는 통과)
		r := copyRun(t, l.result.Run)
		for i, j := 0, len(r.Links)-1; i < j; i, j = i+1, j-1 {
			r.Links[i], r.Links[j] = r.Links[j], r.Links[i]
		}
		for i := range r.Links {
			r.Links[i].ID = i
		}
		if errs := atlas.Validate(r); len(errs) != 0 {
			t.Fatalf("%s: 형식 검사 %v", l.key, errs[0])
		}

		// When: 되돌리면
		got, err := Export(r)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}

		// Then: 원본과 같다
		if d := differs(t, l.raw, got); len(d) != 0 {
			t.Fatalf("%s: 연결 차례만 뒤집었는데 다름 %v", l.key, d[:min(3, len(d))])
		}
	}
}

// relationFromMessages 함수는 run의 메시지 짝(맥 쪽 → 컨테이너 쪽 message 연결)에서 container → mac 관계를 다시 셉니다: 연결 장소마다
// [B의 (tc − t) 최댓값, F의 최솟값]이고 run의 관계는 그 구간들을 모두 덮는 범위입니다(decode/spring.go 204~220, 메모 3).
func relationFromMessages(r *atlas.Run) (lo, hi int64) {
	type span struct{ lo, hi int64 }
	per := map[int]*span{}
	for _, l := range r.Links {
		if l.Kind != "message" {
			continue
		}
		mac, cnt := r.Events[l.Src], r.Events[l.Dst]
		s := per[mac.Place]
		if s == nil {
			s = &span{math.MinInt64, math.MaxInt64}
			per[mac.Place] = s
		}
		d := *cnt.StartNS - *mac.StartNS
		if mac.From != nil {
			s.lo = max(s.lo, d)
		} else {
			s.hi = min(s.hi, d)
		}
	}
	lo, hi = math.MaxInt64, math.MinInt64
	for _, s := range per {
		lo, hi = min(lo, s.lo), max(hi, s.hi)
	}
	return lo, hi
}

// TestClockRelationIsRecountedFromMessages 함수는 run의 container → mac 관계가 그 run의 메시지 짝에서 다시 센 범위와 같은지 확인합니다
// (코드 검토 지적 10). 관계의 부호를 뒤집거나 크기를 1000분의 1로 줄인 사본은 달라야 합니다. 차이가 0에 가까운 기록이 많아 메시지의 방향만
// 보는 시험(F를 컨테이너 쪽이 먼저 읽지 않음)으로는 두 변이를 잡지 못했으므로 값 자체를 견줍니다.
func TestClockRelationIsRecountedFromMessages(t *testing.T) {
	relation := func(r *atlas.Run) *atlas.ClockRelation {
		for i := range r.Clocks {
			if r.Clocks[i].ID == clockCont {
				return r.Clocks[i].Relation
			}
		}
		t.Fatal("컨테이너 시계가 없음")
		return nil
	}
	for _, l := range loadAll(t) {
		// When/Then: 적재한 관계는 메시지에서 다시 센 범위와 같다
		lo, hi := relationFromMessages(l.result.Run)
		if c := relation(l.result.Run); c.LoNS != lo || c.HiNS != hi {
			t.Errorf("%s: 관계 %d~%d ns, 메시지에서 %d~%d ns", l.key, c.LoNS, c.HiNS, lo, hi)
		}

		// When/Then: 관계를 망가뜨린 사본은 다시 센 범위와 다르다
		for name, f := range map[string]func(c *atlas.ClockRelation){
			"부호를 뒤집음":      func(c *atlas.ClockRelation) { c.LoNS, c.HiNS = -c.HiNS, -c.LoNS },
			"1000분의 1로 줄임": func(c *atlas.ClockRelation) { c.LoNS, c.HiNS = c.LoNS/1000, c.HiNS/1000 },
		} {
			r := copyRun(t, l.result.Run)
			f(relation(r))
			if c := relation(r); c.LoNS == lo && c.HiNS == hi {
				t.Errorf("%s: 관계를 %s 사본을 못 가림", l.key, name)
			}
		}
	}
}
