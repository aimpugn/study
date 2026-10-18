// Package gdbtrace 는 실습의 gdb 단계 기록(scripts/trace.py와 run-trace.sh가 쓴 .trace 파일)을 공통 형식으로 옮기고 되돌립니다.
//
// 기록 형식의 정본은 trace.py의 머리 주석입니다. 기존 파서(internal/trace)는 "#" 설명 줄을 건너뛰지만,
// 그 줄도 기록 파일의 사실이므로 여기서는 원문 줄을 직접 읽습니다. 되돌리기(Export)가 원문과 글자까지 같은 파일을 만들면
// 옮기면서 잃은 것이 없다는 뜻입니다(설계 메모 9의 검사 1). 옮기는 규칙은 설계 메모 5의 1입니다.
package gdbtrace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/internal/trace"
)

// 출처 id입니다.
const (
	srcGDB     = "gdb"
	srcPSQL    = "psql"
	srcComment = "comment"
	srcTopic   = "topic"
	srcWatch   = "watch"
	srcGuess   = "watch-guess"
)

// 설명의 종류입니다. 되돌리기가 이 종류로 원문의 자리를 찾습니다.
const (
	noteComment   = "comment"
	noteOutput    = "psql-output"
	noteAmbiguity = "ambiguity"
)

// GuessedWatch 는 짐작한 감시 목록의 경로 무리입니다. 설계 메모 7의 추정 목록에 들어, 출처 성격 검사에 이 이름으로 넘깁니다.
const GuessedWatch = "topic:watch-guess"

// traceEnd 는 trace.py가 기록을 끝내려고 멈추는 함수입니다(TRACE_END의 기본값, trace.py 112행이 감시 목록에 더함).
const traceEnd = "ReadyForQuery"

// Watch 값은 그 기록에 쓴 감시 목록입니다.
//
// Names가 nil이면 어떤 감시 목록을 썼는지 기록에 남지 않은 것이고, 적재기는 기록에 나온 함수로 짐작한 목록을
// 추정 출처로 남깁니다. From은 목록의 근거(감시 목록 파일, 사람이 적은 설명)입니다.
type Watch struct {
	Names []string
	From  string
}

// Result 값은 적재 결과입니다. Origins는 검사에만 쓰는 곁 기록이고 형식에는 들어가지 않습니다.
type Result struct {
	Run     *atlas.Run
	Origins atlas.Origins
}

// WatchFor 함수는 기록 이름마다 그 기록에 쓴 감시 목록을 돌려줍니다. labDir는 실습 폴더(database/postgresql)입니다.
//
// 감시 목록이 기록 밖에 남은 것은 둘뿐입니다(재고 조사 4절): select-where-id-2는 재생본의 meta.watchId가 select이고,
// scan-end-order는 기록 파일 1행의 "#" 줄이 select.watch에 세 함수를 더했다고 밝힙니다. 나머지는 Names가 nil인 Watch입니다.
// 감시 목록 파일을 읽지 못하면 오류를 돌려줍니다.
func WatchFor(name, labDir string) (Watch, error) {
	var extra []string
	var from string
	switch name {
	case "select-where-id-2":
		from = "scripts/select.watch(재생본 replays/select-where-id-2.html의 meta.watchId가 select). 파일은 기록과 같은 날(2026-10-12) 만든 뒤 바뀌지 않았지만, 기록한 순간의 판과 같은지는 기록에 없음"
	case "scan-end-order":
		extra = []string{"heap_endscan", "read_stream_end", "read_stream_reset"}
		from = "scripts/select.watch에 heap_endscan, read_stream_end, read_stream_reset을 더한 목록(기록 파일 1행의 # 설명 줄). 더한 목록의 파일은 남아 있지 않음"
	default:
		return Watch{}, nil
	}
	f, err := os.Open(filepath.Join(labDir, "scripts", "select.watch"))
	if err != nil {
		return Watch{}, err
	}
	defer f.Close()
	exprs, err := trace.ParseWatch(f)
	if err != nil {
		return Watch{}, err
	}
	names := extra
	for fn := range exprs {
		names = append(names, fn)
	}
	sort.Strings(names)
	return Watch{Names: names, From: from}, nil
}

// Import 함수는 .trace 파일 하나의 글을 run으로 옮깁니다.
//
// name은 파일 이름(확장자 없이), text는 파일 글 전체, watch는 그 기록의 감시 목록, version은 소스 판(workbench.json의 sourceRoot)입니다.
// 형식에 맞지 않는 줄(빈 줄, 첫 STOP 뒤의 "#" 줄, 차례가 어긋난 STOP·BT, 둘째 END, 이름이 겹치는 END 함수, NOBREAK)이 있거나
// END가 없거나 마지막 줄바꿈이 없으면 오류를 돌려줍니다. NOBREAK는 지금 기록 여덟 개에 없어 옮기는 규칙을 정하지 않았으므로,
// 조용히 버리지 않고 멈춥니다.
func Import(name, text string, watch Watch, version string) (Result, error) {
	if !strings.HasSuffix(text, "\n") {
		return Result{}, fmt.Errorf("%s: 마지막 줄바꿈이 없음", name)
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")

	// 기록기와 psql 출력은 늘 있는 출처이고, 기록에 원래 없는 칸을 먼저 적습니다.
	b := atlas.NewBuilder("gdb/"+name, "pgsql", "gdb", name, atlas.Origin{Label: "시각 없음: gdb 기록은 멈춘 차례만 남깁니다"})
	origins := atlas.Origins{}
	b.AddSource(atlas.Source{ID: srcGDB, Tool: "gdb + scripts/trace.py", Version: version, Nature: atlas.Observed})
	b.AddSource(atlas.Source{ID: srcPSQL, Tool: "psql(scripts/run-trace.sh가 기록 뒤에 붙인 출력)", Nature: atlas.Observed})
	b.AddUnrecorded(atlas.Unrecorded{Source: srcGDB, Field: "start_ns", Note: "trace.py는 시각을 쓰지 않습니다"})
	b.AddUnrecorded(atlas.Unrecorded{Source: srcGDB, Field: "process.pid", Note: "gdb를 붙인 백엔드의 pid를 기록하지 않습니다"})

	// 기록의 주인공인 백엔드 하나를 장소로 둡니다. 함수와 프레임은 그 아래에 둡니다.
	backend := b.Place(atlas.PlaceSpec{Kind: "process", Name: "backend", Source: srcGDB})
	origins[atlas.RecordRef{Table: "place", ID: backend}] = "$"
	functions := map[string][]int{}
	// function 함수는 함수 장소를 찾거나 만듭니다. 처음 만든 자리의 경로 무리를 그 장소의 무리로 둡니다.
	function := func(fn, path, group string) int {
		var loc *atlas.Loc
		if path != "" {
			loc = &atlas.Loc{Path: path}
		}
		id := b.Place(atlas.PlaceSpec{Parent: &backend, Kind: "function", Name: fn, Loc: loc, Source: srcGDB, Static: true, Version: version})
		ref := atlas.RecordRef{Table: "place", ID: id}
		if _, ok := origins[ref]; !ok {
			origins[ref] = group
			functions[fn] = append(functions[fn], id)
		}
		return id
	}
	// frame 함수는 함수 이름과 위치 글("경로:줄" 또는 "?")로 프레임 장소를 찾거나 만듭니다. group은 원본의 경로 무리입니다.
	frame := func(fn, locText, group string) (int, error) {
		path, line, err := parseLoc(locText)
		if err != nil {
			return 0, err
		}
		f := function(fn, path, group+".function")
		var loc *atlas.Loc
		if path != "" {
			loc = &atlas.Loc{Path: path, Line: line}
		}
		id := b.Place(atlas.PlaceSpec{Parent: &f, Kind: "frame", Name: fn, Loc: loc, Source: srcGDB, Static: true, Version: version})
		if _, ok := origins[atlas.RecordRef{Table: "place", ID: id}]; !ok {
			origins[atlas.RecordRef{Table: "place", ID: id}] = group + ".loc"
		}
		return id, nil
	}

	// 줄을 차례로 읽습니다. "--- psql" 아래는 psql의 출력이라 그대로 모읍니다.
	var (
		stops, last                   = 0, -1
		sawEnd, inOutput, hasOutput   bool
		hasValues, addedCommentSource bool
	)
	for i, line := range lines {
		where := fmt.Sprintf("%s:%d", name, i+1)
		if inOutput {
			b.AddNote(atlas.Note{Scope: atlas.Scope{Kind: "run"}, Kind: noteOutput, Text: line, Source: srcPSQL})
			continue
		}
		switch {
		case line == "--- psql":
			inOutput, hasOutput = true, true
		case line == "":
			return Result{}, fmt.Errorf("%s: 빈 줄", where)
		case strings.HasPrefix(line, "#"):
			// 설명 줄은 첫 STOP 앞에만 있다고 보고 그 차례로 되돌립니다. 뒤에 나오면 자리를 잃으므로 멈춥니다.
			if stops > 0 || sawEnd {
				return Result{}, fmt.Errorf("%s: 첫 STOP 뒤의 # 줄", where)
			}
			if !addedCommentSource {
				b.AddSource(atlas.Source{ID: srcComment, Tool: "기록 파일의 # 설명 줄(기록한 사람이 적음)", Nature: atlas.Human})
				addedCommentSource = true
			}
			b.AddNote(atlas.Note{Scope: atlas.Scope{Kind: "run"}, Kind: noteComment, Text: line[1:], Source: srcComment})
		case strings.HasPrefix(line, "STOP "):
			// STOP <순번> <함수> <위치> [값]: 함수에 멈춘 한 순간이고, 장소는 멈춘 프레임입니다.
			f := strings.SplitN(line[len("STOP "):], " ", 4)
			if len(f) < 3 || sawEnd {
				return Result{}, fmt.Errorf("%s: STOP 칸이 모자라거나 END 뒤에 옴", where)
			}
			if f[0] != strconv.Itoa(stops+1) {
				return Result{}, fmt.Errorf("%s: STOP 순번 %s, 기대 %d", where, f[0], stops+1)
			}
			fr, err := frame(f[1], f[2], "$.steps[].place")
			if err != nil {
				return Result{}, fmt.Errorf("%s: %w", where, err)
			}
			last = b.AddEvent(atlas.Event{Place: fr, Kind: atlas.Instant, Name: f[1], Source: srcGDB})
			origins[atlas.RecordRef{Table: "event", ID: last}] = "$.steps[]"
			b.EventAttrs(last, srcGDB, "seq", json.Number(f[0]))
			if len(f) == 4 {
				b.EventAttrs(last, srcGDB, "values", f[3])
				hasValues = true
			}
			stops++
		case strings.HasPrefix(line, "BT "):
			// BT <깊이> <함수> <위치>: 바로 앞 STOP의 스택 한 칸이고, 깊이는 0부터 하나씩 늘어야 합니다.
			f := strings.Split(line[len("BT "):], " ")
			if len(f) != 3 || last < 0 || sawEnd {
				return Result{}, fmt.Errorf("%s: BT 줄 형식 또는 앞 STOP 없음", where)
			}
			if f[0] != strconv.Itoa(len(b.Run().Events[last].Stack)) {
				return Result{}, fmt.Errorf("%s: BT 깊이 %s, 기대 %d", where, f[0], len(b.Run().Events[last].Stack))
			}
			fr, err := frame(f[1], f[2], "$.steps[].stack[]")
			if err != nil {
				return Result{}, fmt.Errorf("%s: %w", where, err)
			}
			ev := &b.Run().Events[last]
			ev.Stack = append(ev.Stack, fr)
		case line == "END" || strings.HasPrefix(line, "END "):
			// END는 trace.py가 끝에 함수별 멈춘 횟수를 한 번 읽은 것입니다(130행). 항목마다 그 함수 장소의 관찰 사건 하나로 둡니다.
			if sawEnd {
				return Result{}, fmt.Errorf("%s: 둘째 END", where)
			}
			sawEnd = true
			items := strings.TrimPrefix(strings.TrimPrefix(line, "END"), " ")
			if items == "" {
				continue
			}
			for _, pair := range strings.Split(items, " ") {
				fn, hits, ok := strings.Cut(pair, "=")
				if _, err := strconv.Atoi(hits); !ok || err != nil || fn == "" {
					return Result{}, fmt.Errorf("%s: END 항목 %q", where, pair)
				}
				// END의 함수는 감시 목록의 이름이라 경로가 없습니다. 같은 이름의 함수 장소가 하나면 그것이고, 없으면 새로 둡니다.
				var place int
				switch ids := functions[fn]; len(ids) {
				case 0:
					place = function(fn, "", "$.counts[].function")
				case 1:
					place = ids[0]
				default:
					return Result{}, fmt.Errorf("%s: END 함수 %s의 장소가 %d개", where, fn, len(ids))
				}
				ev := b.AddEvent(atlas.Event{Place: place, Kind: atlas.Sample, Name: fn, Source: srcGDB})
				origins[atlas.RecordRef{Table: "event", ID: ev}] = "$.counts[]"
				b.EventAttrs(ev, srcGDB, "hits", json.Number(hits))
			}
		default:
			return Result{}, fmt.Errorf("%s: 모르는 줄 %q", where, line)
		}
	}
	if !sawEnd {
		return Result{}, fmt.Errorf("%s: END 줄이 없음", name)
	}
	if !hasOutput {
		b.AddUnrecorded(atlas.Unrecorded{Source: srcPSQL, Field: "output", Note: "--- psql 줄이 없습니다"})
	}

	// 값 글은 나누지 않고 두었으므로, 그 이유를 run의 모호함 설명으로 남깁니다.
	if hasValues {
		b.AddSource(atlas.Source{ID: srcTopic, Tool: "pgsql 주제의 gdb 적재기(topics/pgsql/gdbtrace)", Nature: atlas.Human})
		b.AddNote(atlas.Note{Scope: atlas.Scope{Kind: "run"}, Kind: noteAmbiguity, Source: srcTopic,
			Text: "STOP 사건의 attrs gdb.values는 scripts/trace.py 85~86행이 \"식=값\"을 공백으로 이어 붙인 원문입니다. 식과 값이 공백을 품을 수 있어 감시 목록의 식 없이는 나눌 수 없으므로 나누지 않았습니다(식 목록으로 나누는 규칙: internal/trace의 splitValues)."})
	}

	// 감시 목록을 관찰 범위로 남깁니다. 목록을 모르면 END에 나온 함수로 짐작한 목록을 추정 출처로 둡니다.
	if watch.Names != nil {
		names := append([]string{}, watch.Names...)
		if !contains(names, traceEnd) {
			names = append(names, traceEnd)
		}
		b.AddSource(atlas.Source{ID: srcWatch, Tool: "감시 목록", Nature: atlas.Human, Note: watch.From})
		b.AddCoverage(atlas.Coverage{Source: srcWatch, What: "함수 진입(gdb 중단점)", Names: names, Completeness: "complete",
			Basis: watch.From + ". " + traceEnd + "는 trace.py가 기록을 끝내려고 더함"})
		origins[atlas.RecordRef{Table: "coverage", ID: len(b.Run().Coverage) - 1}] = "topic:watch"
	} else {
		var names []string
		for _, e := range b.Run().Events {
			if e.Kind == atlas.Sample {
				names = append(names, e.Name)
			}
		}
		b.AddSource(atlas.Source{ID: srcGuess, Tool: "gdb 적재기의 짐작", Nature: atlas.Inferred, Note: "감시 목록 파일이 기록에 남지 않았습니다"})
		b.AddCoverage(atlas.Coverage{Source: srcGuess, What: "함수 진입(gdb 중단점)", Names: names, Completeness: "partial",
			Basis: "추정: END에 나온 함수. 감시했지만 한 번도 멈추지 않은 함수는 빠졌을 수 있습니다"})
		origins[atlas.RecordRef{Table: "coverage", ID: len(b.Run().Coverage) - 1}] = GuessedWatch
	}
	return Result{Run: b.Run(), Origins: origins}, nil
}

// Export 함수는 Import가 만든 run에서 .trace 파일의 글을 다시 만듭니다. run 밖의 것은 읽지 않습니다.
//
// 차례는 "#" 설명 줄, STOP과 그 뒤의 BT, END, "--- psql"과 출력입니다. psql 출력이 미기록이면 "--- psql" 줄을 쓰지 않습니다.
// STOP 사건에 순번이 없거나 END 사건에 횟수가 없으면 오류를 돌려줍니다.
func Export(r *atlas.Run) (string, error) {
	var sb strings.Builder
	for _, n := range r.Notes {
		if n.Kind == noteComment {
			sb.WriteString("#" + n.Text + "\n")
		}
	}

	// STOP과 BT는 instant 사건에서, END 항목은 sample 사건에서 기록 차례(id 차례)로 다시 씁니다.
	var counts []string
	for _, e := range r.Events {
		if e.Source != srcGDB {
			continue
		}
		switch e.Kind {
		case atlas.Instant:
			seq, ok := e.Attrs[srcGDB]["seq"].(json.Number)
			if !ok {
				return "", fmt.Errorf("사건 %d: 순번 없음", e.ID)
			}
			sb.WriteString("STOP " + string(seq) + " " + e.Name + " " + locText(r, e.Place))
			if v, ok := e.Attrs[srcGDB]["values"].(string); ok {
				sb.WriteString(" " + v)
			}
			sb.WriteString("\n")
			for depth, p := range e.Stack {
				fmt.Fprintf(&sb, "BT %d %s %s\n", depth, r.Places[p].Name, locText(r, p))
			}
		case atlas.Sample:
			hits, ok := e.Attrs[srcGDB]["hits"].(json.Number)
			if !ok {
				return "", fmt.Errorf("사건 %d: 횟수 없음", e.ID)
			}
			counts = append(counts, r.Places[e.Place].Name+"="+string(hits))
		}
	}
	sb.WriteString(strings.TrimSuffix("END "+strings.Join(counts, " "), " ") + "\n")

	// psql 출력은 미기록이 아니면 표시 줄과 함께 씁니다.
	for _, u := range r.Unrecorded {
		if u.Source == srcPSQL && u.Field == "output" {
			return sb.String(), nil
		}
	}
	sb.WriteString("--- psql\n")
	for _, n := range r.Notes {
		if n.Kind == noteOutput {
			sb.WriteString(n.Text + "\n")
		}
	}
	return sb.String(), nil
}

// parseLoc 함수는 "경로:줄" 또는 위치를 모를 때의 "?"를 읽습니다. "?"면 빈 경로와 0을 돌려줍니다.
// 줄이 양의 정수가 아니면 오류입니다(internal/trace의 parseLocation과 같은 규칙).
func parseLoc(s string) (string, int, error) {
	if s == "?" {
		return "", 0, nil
	}
	i := strings.LastIndexByte(s, ':')
	if i <= 0 {
		return "", 0, fmt.Errorf("위치 %q", s)
	}
	n, err := strconv.Atoi(s[i+1:])
	if err != nil || n <= 0 {
		return "", 0, fmt.Errorf("위치 %q의 줄", s)
	}
	return s[:i], n, nil
}

// locText 함수는 프레임 장소의 위치를 기록의 꼴("경로:줄" 또는 "?")로 씁니다.
func locText(r *atlas.Run, place int) string {
	p := r.Places[place]
	if p.Loc == nil {
		return "?"
	}
	return p.Loc.Path + ":" + strconv.Itoa(p.Loc.Line)
}

// contains 함수는 목록에 값이 있는지 봅니다.
func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
