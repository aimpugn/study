package txlab

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

// designDir 는 이 패키지 폴더에서 화면 파일(concept-5~8.html)이 있는 폴더까지의 경로입니다.
const designDir = "../../../design/"

// version 은 시험에 쓰는 소스 판입니다. 실습의 workbench.json의 sourceRoot와 같습니다.
const version = "/src/postgresql-18.6"

// screens 는 트랜잭션 실험 화면 넷입니다(재고 조사 3.3절: 각각 실험 1~4의 기록).
var screens = []string{"concept-5-tx.html", "concept-6-locks.html", "concept-7-durability.html", "concept-8-serializable.html"}

// loaded 값은 기록 한 벌의 원본과 적재 결과입니다. raw는 적재에 넘긴 것과 따로 다시 읽은 원본이라,
// 적재기가 넘겨받은 객체를 고쳐도 되돌림 비교가 그 고친 값과 견주지 않습니다.
type loaded struct {
	key    string
	raw    map[string]any
	result Result
}

// loadAll 함수는 화면 넷의 기록 44벌을 열쇠 차례로 적재합니다. 하나라도 실패하면 시험을 멈춥니다.
func loadAll(t *testing.T) []loaded {
	t.Helper()
	var out []loaded
	for _, name := range screens {
		b, err := os.ReadFile(designDir + name)
		if err != nil {
			t.Fatal(err)
		}
		runs, err := ReadRuns(string(b))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		fresh, err := ReadRuns(string(b))
		if err != nil {
			t.Fatal(err)
		}
		keys := make([]string, 0, len(runs))
		for k := range runs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			res, err := Import(k, runs[k].(map[string]any), version)
			if err != nil {
				t.Fatalf("%s: %v", k, err)
			}
			out = append(out, loaded{key: k, raw: fresh[k].(map[string]any), result: res})
		}
	}
	if len(out) != 44 {
		t.Fatalf("기록 %d벌(기대 44)", len(out))
	}
	return out
}

// TestRoundTripRebuildsEveryRun 은 적재 결과를 파일 꼴로 쓰고 다시 읽어 되돌린 기록이 원본과 깊이 같은지 확인합니다(검사 1).
func TestRoundTripRebuildsEveryRun(t *testing.T) {
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

		// When: 그 JSON만 읽어 원본 꼴로 되돌리면
		got, err := Export(back)
		if err != nil {
			t.Fatalf("%s: %v", l.key, err)
		}

		// Then: 원본과 차이가 없다
		assertSame(t, l.key, l.raw, got)
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
		t.Fatalf("%s: 원본과 다름 %v", name, diff)
	}
}

// TestImportedRunsPassFormatAndProvenanceChecks 는 44벌이 형식 검사, 구조 숨김 검사, 출처 성격 검사를 모두 통과하는지 확인합니다(검사 2·3·4).
func TestImportedRunsPassFormatAndProvenanceChecks(t *testing.T) {
	// Given: 트랜잭션 실험의 판정 기준(표시 파일과 대응표)과 추정 목록
	rs, err := roles.Load("tx")
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
			if len(errs) > 10 {
				errs = errs[:10]
			}
			t.Fatalf("%s: %v", l.key, errs)
		}
	}
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

// linkOf 함수는 종류가 kind이고 출발 사건이 이름 src인 첫 연결의 차례를 돌려줍니다.
func linkOf(t *testing.T, r *atlas.Run, kind, src string) int {
	t.Helper()
	for i, l := range r.Links {
		if l.Kind == kind && r.Events[l.Src].Name == src {
			return i
		}
	}
	t.Fatalf("%s에서 나가는 %s 연결 없음", src, kind)
	return 0
}

// TestRoundTripCatchesBrokenStructure 는 구조 칸을 하나씩 망가뜨린 적재에서 되돌린 기록이 원본과 달라지는지 확인합니다.
// 되돌림 비교가 연결과 장소로 되살리는 값(afterKill, walWrittenT, by, 세션의 pid)을 실제로 보는지 판별력을 봅니다.
func TestRoundTripCatchesBrokenStructure(t *testing.T) {
	// Given: kill과 힌트 변화가 있는 기록
	all := loadAll(t)
	crash := find(t, all, "3-crash-sync-off-wait/a")
	hinted := find(t, all, "2-deadlock/a")
	mutations := []struct {
		name   string
		base   loaded
		mutate func(r *atlas.Run)
	}{
		{"관찰 → kill 연결을 지움", crash, func(r *atlas.Run) {
			for i, l := range r.Links {
				if l.Kind == "after" && r.Events[l.Dst].Name == nameKill {
					r.Links[i].Kind = "gone"
				}
			}
		}},
		{"wal 연결을 다른 사건으로 옮김", crash, func(r *atlas.Run) {
			i := linkOf(t, r, "wal", nameKill)
			r.Links[i].Dst--
		}},
		{"kill 앞 단계 연결을 다음 단계로 옮김", crash, func(r *atlas.Run) {
			i := linkOf(t, r, "before", nameKill)
			for _, e := range r.Events {
				if e.Name == nameStep && e.ID > r.Links[i].Dst {
					r.Links[i].Dst = e.ID
					return
				}
			}
		}},
		{"힌트 장소를 다른 프로세스의 코드 자리로 옮김", hinted, func(r *atlas.Run) {
			for i, e := range r.Events {
				if e.Source == srcHint {
					p := r.Places[e.Place]
					for _, q := range r.Places {
						if q.Kind == "site" && *q.Parent != *p.Parent {
							r.Events[i].Place = q.ID
							return
						}
					}
				}
			}
		}},
		{"요청 하나를 다른 세션의 단계로 옮김", hinted, func(r *atlas.Run) {
			i := linkOf(t, r, "contains", nameStep)
			first := r.Events[r.Links[i].Src].Place
			for _, e := range r.Events {
				if e.Name == nameStep && e.Place != first {
					r.Links[i].Src = e.ID
					return
				}
			}
		}},
	}
	for _, m := range mutations {
		// When: 사본을 망가뜨려 되돌리면
		r := copyRun(t, m.base.result.Run)
		m.mutate(r)
		got, err := Export(r)

		// Then: 되돌리기가 실패하거나 원본과 다르다
		if err != nil {
			continue
		}
		w, _ := load.Canonical(m.base.raw)
		g, _ := load.Canonical(got)
		if len(atlas.Diff("$", w, g)) == 0 {
			t.Fatalf("%s: 원본과 같게 되돌아옴", m.name)
		}
	}
}

// TestChecksCatchHiddenStructureAndWrongNature 는 구조를 attrs에 숨기거나 출처 성격을 바꾼 적재를 검사 3·4가 잡는지 확인합니다.
func TestChecksCatchHiddenStructureAndWrongNature(t *testing.T) {
	// Given: 판정 기준과 힌트 변화가 있는 기록의 사본
	rs, err := roles.Load("tx")
	if err != nil {
		t.Fatal(err)
	}
	base := find(t, loadAll(t), "2-deadlock/a")
	step := -1
	for _, e := range base.result.Run.Events {
		if e.Name == nameStep {
			step = e.ID
			break
		}
	}

	// When/Then: 단계의 attrs에 세션(장소 역할)을 남기면 구조 숨김 검사가 실패한다
	r := copyRun(t, base.result.Run)
	r.Events[step].Attrs[srcTxlab]["session"] = "S1"
	if errs := atlas.ScanStructure(r, base.result.Origins, rs); len(errs) != 1 || !strings.Contains(errs[0].Error(), "구조 칸") {
		t.Fatalf("세션을 잡아야 함: %v", errs)
	}

	// When/Then: 요청 창 묶음의 근거를 txlab(관찰)로 바꾸면 출처 성격 검사가 실패한다
	r = copyRun(t, base.result.Run)
	r.Links[linkOf(t, r, "contains", nameStep)].Basis = srcTxlab
	if errs := atlas.CheckNature(r, base.result.Origins, rs, InferredGroups()); len(errs) != 1 || !strings.Contains(errs[0].Error(), RequestWindow) {
		t.Fatalf("창 묶음 근거를 잡아야 함: %v", errs)
	}

	// When/Then: 힌트 변화의 출처를 도출로 바꾸면 추정 목록 때문에 실패한다
	r = copyRun(t, base.result.Run)
	for i, s := range r.Sources {
		if s.ID == srcHint {
			r.Sources[i].Nature = atlas.Derived
		}
	}
	if errs := atlas.CheckNature(r, base.result.Origins, rs, InferredGroups()); len(errs) == 0 {
		t.Fatal("힌트 변화를 도출로 바꾼 것을 잡아야 함")
	}

	// When/Then: 서버 프로세스 장소의 출처를 해석기로 바꾸면(표시는 관찰) 실패한다
	r = copyRun(t, base.result.Run)
	for i, p := range r.Places {
		if p.Kind == "process" {
			r.Places[i].Source = srcTx
		}
	}
	if errs := atlas.CheckNature(r, base.result.Origins, rs, InferredGroups()); len(errs) == 0 {
		t.Fatal("프로세스 출처를 해석기로 바꾼 것을 잡아야 함")
	}
}

// TestTopicBlindFactsMatchOriginalScreens 는 주제의 선택 규칙(facts.json)만으로 플랫폼이 뽑은 단계·메시지·기다림·관찰·kill·멈춤·힌트 변화가
// 원본 화면 데이터에서 바로 센 줄과 차례까지 같은지 확인합니다(검사 5). 출처 성격도 함께 견줘 추정이 관측으로 나오지 않는지 봅니다.
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
		facts, err := atlas.Extract(l.result.Run, rules["tx"])
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
		want := originalFacts(t, l.raw)
		for _, fact := range []string{"단계", "메시지", "기다림", "관찰", "kill", "멈춤", "힌트 변화"} {
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

// rowText 함수는 재추출한 행 하나를 견줄 글 한 줄로 씁니다: 출처 성격|이름|장소|출발→도착|시작~끝|종류|attrs.
func rowText(r atlas.Row) string {
	keys := make([]string, 0, len(r.Attrs))
	for k := range r.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var attrs []string
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

// originalFacts 함수는 원본 한 벌에서 화면이 보이는 사실을 rowText와 같은 꼴의 줄로 셉니다. 적재기를 거치지 않고,
// 장소 이름표 규칙(종류 이름@경로:줄)만 형식에서 가져옵니다.
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
	frame := func(f map[string]any) string {
		if f["at"] == "?" {
			return "frame " + f["fn"].(string)
		}
		return "frame " + f["fn"].(string) + "@" + f["at"].(string)
	}
	pidOf := map[string]string{}
	for _, s := range run["sessions"].([]any) {
		m := s.(map[string]any)
		pidOf[m["name"].(string)] = string(m["pid"].(json.Number))
	}
	out := map[string][]string{}
	add := func(fact, line string) { out[fact] = append(out[fact], line) }
	for _, raw := range run["steps"].([]any) {
		s := raw.(map[string]any)
		sess := s["session"].(string)
		add("단계", fmt.Sprintf("observed|step|session %s|→|%s~%s|span|%s", sess, us(s["t0"]), us(s["t1"]), attrs(s, "n", "sql", "output", "blocked")))
		for _, rq := range s["requests"].([]any) {
			q := rq.(map[string]any)
			proc := "process " + pidOf[sess]
			for _, m := range q["in"].([]any) {
				mm := m.(map[string]any)
				add("메시지", fmt.Sprintf("observed|%s|%s|session %s→|-~-|instant|len=%v", mm["type"], proc, sess, mm["len"]))
			}
			out, _ := q["out"].([]any)
			for _, m := range out {
				mm := m.(map[string]any)
				add("메시지", fmt.Sprintf("observed|%s|%s|→session %s|-~-|instant|len=%v", mm["type"], proc, sess, mm["len"]))
			}
			for _, e := range q["events"].([]any) {
				ev := e.(map[string]any)
				if ev["kind"] != "proc_sleep" {
					continue
				}
				stack := ev["stack"].([]any)
				add("기다림", fmt.Sprintf("observed|proc_sleep|%s|→|%s~%s|span|", frame(stack[len(stack)-1].(map[string]any)), us(ev["t"]), us(ev["t1"])))
			}
		}
	}
	for _, raw := range run["captures"].([]any) {
		c := raw.(map[string]any)
		add("관찰", fmt.Sprintf("observed|capture|client txlab|→|%s~-|sample|%s", us(c["t"]), attrs(c, "pages")))
	}
	for _, raw := range run["kills"].([]any) {
		k := raw.(map[string]any)
		add("kill", fmt.Sprintf("observed|kill|client txlab|→process %v|%s~%s|span|%s", k["pid"], us(k["t"]), us(k["tReady"]), attrs(k, "walWritten", "walSynced")))
	}
	for _, raw := range run["pauses"].([]any) {
		p := raw.(map[string]any)
		add("멈춤", fmt.Sprintf("observed|pause|client txlab|→|%s~%s|span|", us(p["t"]), us(p["t1"])))
	}
	for _, raw := range run["hintChanges"].([]any) {
		h := raw.(map[string]any)
		add("힌트 변화", fmt.Sprintf("inferred|hint_change|site@%s|→|%s~-|instant|%s", h["at"], us(h["t"]), attrs(h, "bit", "ctid", "table", "via")))
	}
	return out
}

// TestChromeExportHasNoCrossingSpans 는 적재 결과를 Chrome 형식으로 옮겼을 때 같은 시계·같은 장소의 구간이 서로 어긋나 겹치지 않는지 확인합니다(검사 2).
// 어긋난 구간은 Perfetto가 넘침 줄로 따로 그립니다. 포개진 구간(요청 안의 교착 검사, 잠금 프레임 안의 잠듦 등) 하나를
// 어긋나게 만들면 검사가 그 구간을 잡아야 합니다(검사가 볼 것이 실제로 있었는지). 여정은 장소마다 시각 있는 구간이 하나뿐이고
// gdb 기록은 시각이 없어, 이 검사가 볼 것이 있는 적재는 이 패키지뿐입니다.
func TestChromeExportHasNoCrossingSpans(t *testing.T) {
	// Given: 기록 44벌
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

	// When: 포개진 구간이 있는 첫 기록의 사본에서 안쪽 구간의 끝을 바깥 구간의 끝 너머로 민다
	for _, l := range all {
		outer, inner, ok := nestedPair(l.result.Run)
		if !ok {
			continue
		}
		r := copyRun(t, l.result.Run)
		end := *r.Events[outer].EndNS + 1
		r.Events[inner].EndNS = &end

		// Then: 그 구간이 어긋난 쌍으로 잡힌다
		for _, o := range atlas.CrossingSpans(r) {
			if o.B == inner || o.A == inner {
				return
			}
		}
		t.Fatalf("%s: 사건 %d를 사건 %d 밖으로 밀었는데 어긋남을 못 찾음", l.key, inner, outer)
	}
	t.Fatal("포개진 구간이 있는 기록이 없음: 어긋남 검사가 볼 것이 없었음")
}

// nestedPair 함수는 같은 장소·같은 출처에서 끝이 있는 구간이 다른 구간 안에 늦게 시작해 일찍 끝나는 첫 쌍을 찾습니다.
// 출처가 같으면 시계도 같아 CrossingSpans가 같은 줄로 봅니다.
func nestedPair(r *atlas.Run) (outer, inner int, ok bool) {
	timed := func(e atlas.Event) bool { return e.Kind == atlas.Span && e.StartNS != nil && e.EndNS != nil }
	for _, a := range r.Events {
		for _, b := range r.Events {
			if timed(a) && timed(b) && a.Place == b.Place && a.Source == b.Source &&
				*a.StartNS < *b.StartNS && *b.EndNS < *a.EndNS {
				return a.ID, b.ID, true
			}
		}
	}
	return 0, 0, false
}

// TestImportRejectsWalLinkThatDoesNotMatch 는 kill의 walWrittenT가 가리키는 쓰기가 pg_wal 파일이 아니거나, 그 쓰기가 닿은 위치가
// walWritten과 다르면 적재기가 멈추는지 확인합니다. 끝 시각만 우연히 같은 다른 쓰기에 wal 연결을 붙이지 않기 위해서입니다.
func TestImportRejectsWalLinkThatDoesNotMatch(t *testing.T) {
	// Given: walWrittenT가 있는 kill과 끝이 그 시각인 쓰기를 고르는 함수(원본을 화면 파일에서 새로 읽음)
	fresh := func() (string, map[string]any, map[string]any, map[string]any) {
		for _, name := range screens {
			b, err := os.ReadFile(designDir + name)
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
			for _, k := range keys {
				run := runs[k].(map[string]any)
				for _, kr := range run["kills"].([]any) {
					kill := kr.(map[string]any)
					if w := writeEndingAt(run, kill["walWrittenT"]); w != nil {
						return k, run, kill, w
					}
				}
			}
		}
		t.Fatal("walWrittenT가 있는 kill이 없음")
		return "", nil, nil, nil
	}
	if key, run, _, _ := fresh(); func() error { _, err := Import(key, run, version); return err }() != nil {
		t.Fatalf("%s: 고치지 않은 원본은 적재되어야 함", key)
	}

	for _, c := range []struct {
		name   string
		mutate func(kill, write map[string]any)
	}{
		{"쓰기의 파일을 pg_wal 밖으로", func(_, w map[string]any) { w["f"].(map[string]any)["file"] = "base/5/16384" }},
		{"walWritten을 다른 위치로", func(k, _ map[string]any) { k["walWritten"] = "0/1" }},
	} {
		key, run, kill, write := fresh()

		// When: 한 칸을 고쳐 적재하면
		c.mutate(kill, write)
		_, err := Import(key, run, version)

		// Then: 적재기가 멈춘다
		if err == nil {
			t.Errorf("%s: %s 적재가 통과함", c.name, key)
		}
	}
}

// writeEndingAt 함수는 원본 기록에서 끝(t1)이 at인 pwrite64 쓰기(요청 안 io 또는 서버 wal_io)를 찾습니다. 없으면 nil입니다.
func writeEndingAt(run map[string]any, at any) map[string]any {
	if at == nil {
		return nil
	}
	check := func(e map[string]any) bool {
		f, _ := e["f"].(map[string]any)
		return f != nil && f["call"] == "pwrite64" && fmt.Sprint(e["t1"]) == fmt.Sprint(at)
	}
	for _, s := range run["steps"].([]any) {
		reqs, _ := s.(map[string]any)["requests"].([]any)
		for _, q := range reqs {
			evs, _ := q.(map[string]any)["events"].([]any)
			for _, e := range evs {
				if m := e.(map[string]any); m["kind"] == "io" && check(m) {
					return m
				}
			}
		}
	}
	evs, _ := run["server"].(map[string]any)["events"].([]any)
	for _, e := range evs {
		if m := e.(map[string]any); m["kind"] == "wal_io" && check(m) {
			return m
		}
	}
	return nil
}
