package gdbtrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/internal/trace"
	"study/open_source_internals/workbench/topics/pgsql/roles"
)

// labDir 는 이 패키지 폴더에서 실습 폴더까지의 경로입니다. 기록 원본과 감시 목록, workbench.json이 거기 있습니다.
const labDir = "../../../../database/postgresql"

// loaded 값은 기록 하나의 원문과 적재 결과입니다.
type loaded struct {
	name, text string
	res        Result
}

// loadAll 함수는 evidence/*.trace 여덟 개를 모두 적재합니다. 하나라도 실패하면 시험을 멈춥니다.
func loadAll(t *testing.T) []loaded {
	t.Helper()
	var subject struct {
		SourceRoot string `json:"sourceRoot"`
	}
	raw, err := os.ReadFile(filepath.Join(labDir, "workbench.json"))
	if err != nil || json.Unmarshal(raw, &subject) != nil || subject.SourceRoot == "" {
		t.Fatalf("workbench.json: %v", err)
	}
	paths, err := filepath.Glob(filepath.Join(labDir, "evidence", "*.trace"))
	if err != nil || len(paths) != 8 {
		t.Fatalf("기록 %d개(기대 8): %v", len(paths), err)
	}
	var out []loaded
	for _, p := range paths {
		text, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(filepath.Base(p), ".trace")
		watch, err := WatchFor(name, labDir)
		if err != nil {
			t.Fatal(err)
		}
		res, err := Import(name, string(text), watch, subject.SourceRoot)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		out = append(out, loaded{name: name, text: string(text), res: res})
	}
	return out
}

// TestRoundTripRebuildsEveryTraceByte 함수는 적재 결과를 파일로 쓰고 다시 읽어 되돌린 글이 원문과 글자까지 같은지 확인합니다(검사 1).
func TestRoundTripRebuildsEveryTraceByte(t *testing.T) {
	for _, l := range loadAll(t) {
		// Given: 적재 결과를 형식의 JSON으로 쓴 것
		var buf bytes.Buffer
		if err := atlas.Write(&buf, l.res.Run); err != nil {
			t.Fatal(err)
		}

		// When: 그 JSON만 읽어 되돌리면
		back, err := atlas.Read(&buf)
		if err != nil {
			t.Fatalf("%s: %v", l.name, err)
		}
		got, err := Export(back)
		if err != nil {
			t.Fatalf("%s: %v", l.name, err)
		}

		// Then: 원문과 같다. 다르면 처음 다른 줄을 보인다
		if got != l.text {
			a, b := strings.Split(l.text, "\n"), strings.Split(got, "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Fatalf("%s %d행\n원문 %q\n되돌림 %q", l.name, i+1, a[i], b[i])
				}
			}
			t.Fatalf("%s: 줄 수 %d, 되돌림 %d", l.name, len(a), len(b))
		}
	}
}

// TestImportedRunsPassFormatAndProvenanceChecks 함수는 여덟 기록이 형식 검사, 구조 숨김 검사, 출처 성격 검사를 모두 통과하고
// 감시 목록을 아는 둘만 완전한 관찰 범위를 갖는지 확인합니다(검사 2·3·4).
func TestImportedRunsPassFormatAndProvenanceChecks(t *testing.T) {
	// Given: 여덟 기록과 E-05의 판정 기준(표시 파일과 대응표)
	rs, err := roles.Load("e05")
	if err != nil {
		t.Fatal(err)
	}
	complete := map[string]bool{}
	for _, l := range loadAll(t) {
		// When: 세 검사를 하면
		errs := atlas.Validate(l.res.Run)
		errs = append(errs, atlas.ScanStructure(l.res.Run, l.res.Origins, rs)...)
		errs = append(errs, atlas.CheckNature(l.res.Run, l.res.Origins, rs, map[string]bool{GuessedWatch: true})...)

		// Then: 어긴 곳이 없다
		if len(errs) != 0 {
			t.Fatalf("%s: %v", l.name, errs)
		}
		if c := l.res.Run.Coverage; len(c) != 1 {
			t.Fatalf("%s: 관찰 범위 %d개", l.name, len(c))
		} else if c[0].Completeness == "complete" {
			complete[l.name] = true
		}
	}
	if want := map[string]bool{"select-where-id-2": true, "scan-end-order": true}; !reflect.DeepEqual(complete, want) {
		t.Fatalf("완전한 관찰 범위 %v", complete)
	}
}

// TestChecksCatchHiddenStructureAndWrongNature 함수는 적재 결과를 일부러 망가뜨리면 구조 숨김 검사와 출처 성격 검사가 실패하는지 확인합니다(변이 시험).
func TestChecksCatchHiddenStructureAndWrongNature(t *testing.T) {
	rs, err := roles.Load("e05")
	if err != nil {
		t.Fatal(err)
	}
	inferred := map[string]bool{GuessedWatch: true}
	var known, guessed Result
	for _, l := range loadAll(t) {
		switch l.name {
		case "select-where-id-2":
			known = l.res
		case "commit-sync-on":
			guessed = l.res
		}
	}

	// When/Then: STOP 사건의 attrs에 멈춘 자리(장소 역할)를 남기면 구조 숨김 검사가 실패한다
	known.Run.Events[0].Attrs[srcGDB]["place"] = "src/backend/tcop/postgres.c:1014"
	if errs := atlas.ScanStructure(known.Run, known.Origins, rs); len(errs) != 1 || !strings.Contains(errs[0].Error(), "구조 칸") {
		t.Fatalf("장소를 잡아야 함: %v", errs)
	}
	delete(known.Run.Events[0].Attrs[srcGDB], "place")

	// When/Then: 원본의 단계 목록을 run attrs에 통째로 넣으면, 모음 키라도 그 안의 사건 때문에 실패한다
	known.Origins[atlas.RecordRef{Table: "run", ID: 0}] = "$"
	known.Run.Attrs[srcGDB] = map[string]any{"steps": []any{map[string]any{"seq": 1, "place": map[string]any{"function": "exec_simple_query"}}}}
	if errs := atlas.ScanStructure(known.Run, known.Origins, rs); len(errs) != 1 || !strings.Contains(errs[0].Error(), "아래의") {
		t.Fatalf("통째로 넣은 단계를 잡아야 함: %v", errs)
	}
	delete(known.Run.Attrs, srcGDB)

	// When/Then: 기록기 출처를 도출로 바꾸면 출처 성격 검사가 실패한다
	known.Run.Sources[0].Nature = atlas.Derived
	if errs := atlas.CheckNature(known.Run, known.Origins, rs, inferred); len(errs) == 0 {
		t.Fatal("관찰을 도출로 바꾼 것을 잡아야 함")
	}

	// When/Then: 짐작한 감시 목록의 출처를 사람이 정한 것으로 바꾸면 실패한다
	for i, s := range guessed.Run.Sources {
		if s.ID == srcGuess {
			guessed.Run.Sources[i].Nature = atlas.Human
		}
	}
	if errs := atlas.CheckNature(guessed.Run, guessed.Origins, rs, inferred); len(errs) != 1 || !strings.Contains(errs[0].Error(), "coverage 0") {
		t.Fatalf("짐작을 사람 글로 바꾼 것을 잡아야 함: %v", errs)
	}
}

// TestTopicBlindFactsMatchOriginalParser 함수는 주제의 선택 규칙(facts.json)만으로 플랫폼이 뽑은 사실이,
// 원래 화면이 쓰는 파서(internal/trace)로 원문을 읽은 값과 같은지 확인합니다(검사 5).
func TestTopicBlindFactsMatchOriginalParser(t *testing.T) {
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
		// When: 플랫폼이 규칙대로 뽑고, 원래 파서가 원문을 읽으면
		facts, err := atlas.Extract(l.res.Run, rules["gdb"])
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]atlas.Fact{}
		for _, f := range facts {
			got[f.Fact] = f
		}
		rec, err := trace.Parse(strings.NewReader(l.text), nil)
		if err != nil {
			t.Fatalf("%s: %v", l.name, err)
		}

		// Then: 단계마다 함수, 멈춘 자리, 순번, 값 글, 스택이 같고 모두 관찰이다
		steps := got["단계"].Rows
		if len(steps) != len(rec.Steps) {
			t.Fatalf("%s: 단계 %d, 원래 %d", l.name, len(steps), len(rec.Steps))
		}
		for i, s := range rec.Steps {
			want := atlas.Row{Name: s.Function, Place: "frame " + label(s.Function, s.Loc), Nature: atlas.Observed,
				Attrs: map[string]any{"gdb.seq": json.Number(strconv.Itoa(s.Seq))}}
			if s.RawValues != "" {
				want.Attrs["gdb.values"] = s.RawValues
			}
			for _, f := range s.Stack {
				want.Stack = append(want.Stack, label(f.Function, f.Loc))
			}
			row := steps[i]
			if row.Name != want.Name || row.Place != want.Place || row.Nature != want.Nature ||
				!reflect.DeepEqual(row.Attrs, want.Attrs) || !reflect.DeepEqual(row.Stack, want.Stack) {
				t.Fatalf("%s 단계 %d\n뽑음 %+v\n원래 %+v", l.name, i+1, row, want)
			}
		}

		// Then: END 항목이 차례까지 같고, 함수마다의 횟수가 그 함수의 STOP 수와 같다
		var counts []string
		stopsOf := map[string]int{}
		for _, s := range steps {
			stopsOf[s.Name]++
		}
		for _, r := range got["END 횟수"].Rows {
			counts = append(counts, fmt.Sprintf("%s=%s", r.Name, r.Attrs["gdb.hits"]))
			if n, _ := strconv.Atoi(string(r.Attrs["gdb.hits"].(json.Number))); n != stopsOf[r.Name] {
				t.Fatalf("%s: %s의 END %d, STOP %d", l.name, r.Name, n, stopsOf[r.Name])
			}
		}
		var want []string
		for _, c := range rec.Counts {
			want = append(want, fmt.Sprintf("%s=%d", c.Function, c.Hits))
		}
		if !reflect.DeepEqual(counts, want) {
			t.Fatalf("%s: END %v, 원래 %v", l.name, counts, want)
		}

		// Then: psql 출력과 # 설명 줄이 같다
		if !sameLines(got["psql 출력"].Notes, rec.Output) {
			t.Fatalf("%s: 출력 %q, 원래 %q", l.name, got["psql 출력"].Notes, rec.Output)
		}
		var comments []string
		for _, line := range strings.Split(l.text, "\n") {
			if strings.HasPrefix(line, "#") {
				comments = append(comments, line[1:])
			}
		}
		if !sameLines(got["기록 설명"].Notes, comments) {
			t.Fatalf("%s: 설명 %q, 원래 %q", l.name, got["기록 설명"].Notes, comments)
		}

		// Then: 완전한 관찰 범위라면 멈춘 함수가 모두 그 목록에 있다
		if c := l.res.Run.Coverage[0]; c.Completeness == "complete" {
			for fn := range stopsOf {
				if !contains(c.Names, fn) {
					t.Fatalf("%s: 멈춘 함수 %s가 감시 목록에 없음", l.name, fn)
				}
			}
		}
	}
}

// TestImportRejectsMalformedTraces 함수는 형식에 맞지 않는 기록을 조용히 옮기지 않고 거부하는지 확인합니다.
func TestImportRejectsMalformedTraces(t *testing.T) {
	ok := "STOP 1 f a.c:1\nBT 0 f a.c:1\nEND f=1\n--- psql\n1\n"
	// Given: 올바른 기록은 통과한다
	if _, err := Import("ok", ok, Watch{}, "v"); err != nil {
		t.Fatalf("올바른 기록: %v", err)
	}
	cases := map[string]string{
		"마지막 줄바꿈 없음":  strings.TrimSuffix(ok, "\n"),
		"빈 줄":         "STOP 1 f a.c:1\n\nEND f=1\n",
		"STOP 뒤의 # 줄": "STOP 1 f a.c:1\n# 설명\nEND f=1\n",
		"순번 건너뜀":      "STOP 2 f a.c:1\nEND f=1\n",
		"BT 깊이 건너뜀":   "STOP 1 f a.c:1\nBT 1 f a.c:1\nEND f=1\n",
		"END 없음":      "STOP 1 f a.c:1\n",
		"둘째 END":      "STOP 1 f a.c:1\nEND f=1\nEND f=1\n",
		"NOBREAK":     "NOBREAK g\nSTOP 1 f a.c:1\nEND f=1\n",
		"줄 번호 0":      "STOP 1 f a.c:0\nEND f=1\n",
		"END 횟수 글자":   "STOP 1 f a.c:1\nEND f=x\n",
	}
	for name, text := range cases {
		// When/Then: 형식을 어긴 기록은 오류다
		if _, err := Import("bad", text, Watch{}, "v"); err == nil {
			t.Fatalf("%s: 거부해야 함", name)
		}
	}
}

// label 함수는 원래 파서의 위치를 플랫폼의 스택 칸 꼴("이름@경로:줄", 위치를 모르면 이름)로 씁니다.
func label(fn string, loc trace.Location) string {
	if loc.Path == "" {
		return fn
	}
	return fmt.Sprintf("%s@%s:%d", fn, loc.Path, loc.Line)
}

// sameLines 함수는 두 줄 목록이 같은지 봅니다. nil과 빈 목록은 같게 봅니다(둘 다 줄이 없음).
func sameLines(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
