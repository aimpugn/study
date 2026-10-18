// Command decode 는 pgtrace_hook.c가 남긴 함수 입구·출구 기록을 함수 이름, 정의 파일, 소스 디렉터리로 풀어 요약합니다.
//
// 탐색용 코드입니다(2026-10-13, 작업대 관측 재설계의 P15~P20 시험). 시험이 방법을 확정하면
// 작업대의 internal 패키지로 다시 설계해 옮기고 이 파일을 지웁니다.
//
// 사용:
//
//	decode -bin postgres -dir trace summary
//	decode -bin postgres -dir trace query -pid 123 -nth 2 [-find a,b] [-tree 3] [-json 4 [-keep storage/buffer,storage/smgr]]
//	decode -bin postgres -dir trace window -pid 456 -from NS -to NS
//	decode -bin postgres -dir trace journey -pid 123 -nth 1 -strace all.strace
//	decode -bin postgres -dir <기록>/trace spring -rec <기록> -jfr <JDK>/bin/jfr [-src <PostgreSQL 소스>] [-jsrc <Java 소스 뿌리>,...]
//	decode spring-compare -out DIR 장면/실행=<spring 출력> ...
//
// -bin은 기록을 남긴 실행 파일(DWARF 포함), -dir은 <pid>.bin 파일들이 있는 디렉터리입니다.
// 레코드 형식의 정본은 pgtrace_hook.c의 머리 주석입니다.
package main

import (
	"debug/dwarf"
	"debug/elf"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// srcPrefix 는 실습 이미지 안의 소스 트리 위치입니다. 정의 파일 경로에서 이 앞부분을 떼어 상대경로로 보입니다.
const srcPrefix = "/src/postgresql-18.6/"

// 레코드 종류. pgtrace_hook.c의 kind 값과 같습니다.
const (
	kindHead    = 0
	kindEnter   = 1
	kindExit    = 2
	kindDropped = 3
)

// recordSize 는 레코드 하나의 바이트 수입니다(8바이트 정수 다섯).
const recordSize = 40

// record 하나는 기록 파일의 레코드 하나입니다. Frame은 기록된 함수의 프레임 주소입니다.
type record struct {
	Kind, Fn, Site, Ns, Frame uint64
}

// funcInfo 는 DWARF에서 읽은 함수 하나입니다. File은 소스 트리 기준 상대경로이고 Line은 함수 이름이 정의된 줄입니다.
type funcInfo struct {
	Name string
	File string
	Line int64
	Low  uint64
}

// cuRange 는 컴파일 단위 하나가 차지하는 주소 범위이고, 부른 자리의 줄을 찾을 때 씁니다.
type cuRange struct {
	Low, High uint64
	Entry     *dwarf.Entry
}

// symbols 는 실행 파일 하나의 주소 해석 자료입니다.
type symbols struct {
	data   *dwarf.Data
	byLow  map[uint64]*funcInfo
	cus    []cuRange
	mainAt uint64
	// vars 는 ELF 기호표의 데이터 기호(STT_OBJECT)를 주소로 찾는 표입니다. 함수 인자로 넘어온 정적 변수의 주소를 이름으로 바꿀 때 씁니다.
	vars map[uint64]string
}

// loadSymbols 는 실행 파일의 DWARF에서 함수의 시작 주소, 이름, 정의 파일을 읽습니다.
//
// Arguments: bin은 -finstrument-functions로 컴파일한 실행 파일 경로입니다.
// Returns: 시작 주소로 함수를 찾는 표와, PIE 차이를 구할 main의 ELF 주소.
// Errors: 파일을 열 수 없거나 DWARF가 없거나 main 심볼이 없으면 오류를 돌려줍니다.
func loadSymbols(bin string) (*symbols, error) {
	f, err := elf.Open(bin)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 기록 머리의 main 주소와 맞춰 PIE 차이를 구하기 위해 ELF 심볼 표에서 main을 찾습니다.
	syms, err := f.Symbols()
	if err != nil {
		return nil, err
	}
	s := &symbols{byLow: map[uint64]*funcInfo{}, vars: map[uint64]string{}}
	for _, sym := range syms {
		if sym.Name == "main" && elf.ST_TYPE(sym.Info) == elf.STT_FUNC {
			s.mainAt = sym.Value
		}
		if elf.ST_TYPE(sym.Info) == elf.STT_OBJECT {
			s.vars[sym.Value] = sym.Name
		}
	}
	if s.mainAt == 0 {
		return nil, errors.New("main 심볼이 없습니다")
	}

	s.data, err = f.DWARF()
	if err != nil {
		return nil, err
	}

	// 컴파일 단위마다 파일 표를 읽어 두고, 그 안의 함수 정의를 시작 주소로 등록합니다.
	r := s.data.Reader()
	var files []*dwarf.LineFile
	compDir := ""
	for {
		e, err := r.Next()
		if err != nil {
			return nil, err
		}
		if e == nil {
			break
		}
		switch e.Tag {
		case dwarf.TagCompileUnit:
			files = nil
			// 헤더의 파일 이름은 "../../../../src/include/..."처럼 컴파일 디렉터리 기준 상대경로로 남으므로(실행에서 관측) 합쳐서 씁니다.
			compDir, _ = e.Val(dwarf.AttrCompDir).(string)
			if lr, err := s.data.LineReader(e); err == nil && lr != nil {
				files = lr.Files()
			}
			if ranges, err := s.data.Ranges(e); err == nil {
				for _, rg := range ranges {
					s.cus = append(s.cus, cuRange{Low: rg[0], High: rg[1], Entry: e})
				}
			}
		case dwarf.TagSubprogram:
			low, ok := e.Val(dwarf.AttrLowpc).(uint64)
			if !ok {
				continue
			}
			name, _ := e.Val(dwarf.AttrName).(string)
			file := ""
			if idx, ok := e.Val(dwarf.AttrDeclFile).(int64); ok && int(idx) < len(files) && files[idx] != nil {
				file = relSource(compDir, files[idx].Name)
			}
			line, _ := e.Val(dwarf.AttrDeclLine).(int64)
			s.byLow[low] = &funcInfo{Name: name, File: file, Line: line, Low: low}
		}
	}
	sort.Slice(s.cus, func(i, j int) bool { return s.cus[i].Low < s.cus[j].Low })
	return s, nil
}

// relSource 는 DWARF의 파일 이름을 컴파일 디렉터리와 합친 뒤 소스 트리 기준 상대경로로 돌려줍니다.
func relSource(compDir, name string) string {
	if !path.IsAbs(name) {
		name = path.Join(compDir, name)
	}
	return strings.TrimPrefix(path.Clean(name), srcPrefix)
}

// lineAt 은 ELF 주소 pc가 속한 소스 줄을 "상대경로:줄"로 돌려줍니다. 찾지 못하면 "?"입니다.
func (s *symbols) lineAt(pc uint64) string {
	i := sort.Search(len(s.cus), func(i int) bool { return s.cus[i].High > pc })
	if i == len(s.cus) || s.cus[i].Low > pc {
		return "?"
	}
	lr, err := s.data.LineReader(s.cus[i].Entry)
	if err != nil || lr == nil {
		return "?"
	}
	var le dwarf.LineEntry
	if err := lr.SeekPC(pc, &le); err != nil {
		return "?"
	}
	compDir, _ := s.cus[i].Entry.Val(dwarf.AttrCompDir).(string)
	return fmt.Sprintf("%s:%d", relSource(compDir, le.File.Name), le.Line)
}

// module 은 정의 파일의 상대경로를 부품을 고르는 열쇠로 줄입니다.
// src/backend 아래는 두 단계 디렉터리(storage/buffer), src/include 아래는 include/첫 디렉터리, 나머지는 디렉터리입니다.
func module(file string) string {
	if file == "" {
		return "(정의 파일 없음)"
	}
	parts := strings.Split(path.Dir(file), "/")
	switch {
	case len(parts) >= 4 && parts[0] == "src" && parts[1] == "backend":
		return parts[2] + "/" + parts[3]
	case len(parts) == 3 && parts[0] == "src" && parts[1] == "backend":
		return parts[2]
	case len(parts) >= 3 && parts[0] == "src" && parts[1] == "include":
		return "include/" + parts[2]
	case len(parts) == 2 && parts[0] == "src" && parts[1] == "include":
		return "include"
	default:
		return path.Dir(file)
	}
}

// traceFile 은 프로세스 하나의 기록 전체입니다.
type traceFile struct {
	Pid     int
	Slide   uint64
	Records []record
}

// readTrace 는 <pid>.bin 파일 하나를 읽고 머리 레코드로 PIE 차이를 구합니다.
//
// Errors: 파일 크기가 레코드 크기의 배수가 아니거나 첫 레코드가 머리가 아니면 오류를 돌려줍니다.
func readTrace(file string, mainAt uint64) (*traceFile, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if len(raw)%recordSize != 0 {
		return nil, fmt.Errorf("%s: 크기 %d가 %d의 배수가 아닙니다", file, len(raw), recordSize)
	}
	t := &traceFile{}
	for off := 0; off < len(raw); off += recordSize {
		t.Records = append(t.Records, record{
			Kind:  binary.LittleEndian.Uint64(raw[off:]),
			Fn:    binary.LittleEndian.Uint64(raw[off+8:]),
			Site:  binary.LittleEndian.Uint64(raw[off+16:]),
			Ns:    binary.LittleEndian.Uint64(raw[off+24:]),
			Frame: binary.LittleEndian.Uint64(raw[off+32:]),
		})
	}
	if len(t.Records) == 0 || t.Records[0].Kind != kindHead {
		return nil, fmt.Errorf("%s: 머리 레코드가 없습니다", file)
	}
	t.Pid = int(t.Records[0].Site)
	t.Slide = t.Records[0].Fn - mainAt
	return t, nil
}

// node 는 호출 나무의 한 칸입니다. Calls는 이 칸 아래(자신 포함)의 함수 호출 수입니다.
type node struct {
	Fn       uint64
	Frame    uint64
	Site     uint64
	Children []*node
	Calls    int
	StartNs  uint64
	EndNs    uint64
	Unwound  bool
}

// stats 는 한 구간의 호출을 센 결과입니다.
type stats struct {
	Calls, Unresolved, OrphanExits, Unwound, Dropped, MaxDepth int
	ByFn                                                       map[uint64]int
	StartNs, EndNs                                             uint64
	Root                                                       *node
}

// buildTree 는 records[from:]에서 첫 입구의 함수가 끝날 때까지 호출 나무를 만듭니다.
//
// longjmp(elog(ERROR))는 중간 함수들의 출구 기록을 건너뜁니다. 스택은 아래로 자라므로, 살아 있는 조상의
// 프레임 주소는 새로 들어온 함수의 프레임 주소보다 큽니다. 그래서 새 입구나 출구의 프레임 주소 이하인 칸은
// 이미 사라진 것으로 보고 Unwound로 닫습니다. 첫 칸이 이렇게 닫히면 그 시각을 구간의 끝으로 씁니다.
func buildTree(t *traceFile, from int, syms *symbols) *stats {
	st := &stats{ByFn: map[uint64]int{}}
	var stack []*node

	// unwindTo 는 프레임 주소가 frame 이하(같음 포함 여부는 inclusive)인 칸을 스택 위에서부터 닫습니다.
	unwindTo := func(frame uint64, inclusive bool, ns uint64) {
		for len(stack) > 0 {
			top := stack[len(stack)-1]
			if top.Frame > frame || (!inclusive && top.Frame == frame) {
				return
			}
			top.Unwound = true
			top.EndNs = ns
			st.Unwound++
			stack = stack[:len(stack)-1]
		}
	}

	for i := from; i < len(t.Records); i++ {
		r := t.Records[i]
		fn := r.Fn - t.Slide
		switch r.Kind {
		case kindDropped:
			st.Dropped += int(r.Fn)
		case kindEnter:
			if st.Root != nil {
				unwindTo(r.Frame, true, r.Ns)
				if len(stack) == 0 {
					st.EndNs = r.Ns
					countCalls(st.Root)
					return st
				}
			}
			n := &node{Fn: fn, Frame: r.Frame, Site: r.Site - t.Slide, StartNs: r.Ns}
			if st.Root == nil {
				st.Root = n
				st.StartNs = r.Ns
			} else if len(stack) == 0 {
				return st
			} else {
				top := stack[len(stack)-1]
				top.Children = append(top.Children, n)
			}
			stack = append(stack, n)
			st.Calls++
			st.ByFn[fn]++
			if syms.byLow[fn] == nil {
				st.Unresolved++
			}
			if len(stack) > st.MaxDepth {
				st.MaxDepth = len(stack)
			}
		case kindExit:
			if st.Root == nil {
				continue
			}
			unwindTo(r.Frame, false, r.Ns)
			if len(stack) == 0 {
				st.EndNs = r.Ns
				countCalls(st.Root)
				return st
			}
			// 스택 위에서부터 같은 함수를 찾습니다. 바로 위가 아니면 그 사이 칸은 longjmp로 건너뛴 것입니다.
			j := len(stack) - 1
			for j >= 0 && stack[j].Fn != fn {
				j--
			}
			if j < 0 {
				st.OrphanExits++
				continue
			}
			for k := len(stack) - 1; k > j; k-- {
				stack[k].Unwound = true
				stack[k].EndNs = r.Ns
				st.Unwound++
			}
			stack[j].EndNs = r.Ns
			stack = stack[:j]
			if len(stack) == 0 {
				st.EndNs = r.Ns
				countCalls(st.Root)
				return st
			}
		}
	}
	if st.Root != nil {
		countCalls(st.Root)
	}
	return st
}

// countCalls 는 각 칸의 Calls를 자신과 자식들의 합으로 채웁니다.
func countCalls(n *node) int {
	n.Calls = 1
	for _, c := range n.Children {
		n.Calls += countCalls(c)
	}
	return n.Calls
}

// name 은 ELF 주소의 함수 이름을 돌려줍니다. 모르면 주소를 16진수로 보입니다.
func (s *symbols) name(fn uint64) string {
	if f := s.byLow[fn]; f != nil {
		return f.Name
	}
	return fmt.Sprintf("0x%x", fn)
}

// file 은 ELF 주소의 함수가 정의된 상대경로를 돌려줍니다.
func (s *symbols) file(fn uint64) string {
	if f := s.byLow[fn]; f != nil {
		return f.File
	}
	return ""
}

// printTree 는 호출 나무를 maxDepth 단계까지, 같은 함수가 이어지면 한 줄로 묶어 출력합니다.
func printTree(w io.Writer, s *symbols, n *node, depth, maxDepth int) {
	unwound := ""
	if n.Unwound {
		unwound = " [longjmp로 닫힘]"
	}
	fmt.Fprintf(w, "%s%s  calls=%d  %s  from %s%s\n", strings.Repeat("  ", depth), s.name(n.Fn), n.Calls,
		module(s.file(n.Fn)), s.lineAt(n.Site-4), unwound)
	if depth+1 >= maxDepth {
		return
	}
	for i := 0; i < len(n.Children); {
		j := i
		sum := 0
		for j < len(n.Children) && n.Children[j].Fn == n.Children[i].Fn {
			sum += n.Children[j].Calls
			j++
		}
		if j-i > 1 {
			c := n.Children[i]
			fmt.Fprintf(w, "%s%s ×%d  calls=%d  %s\n", strings.Repeat("  ", depth+1), s.name(c.Fn), j-i, sum, module(s.file(c.Fn)))
		} else {
			printTree(w, s, n.Children[i], depth+1, maxDepth)
		}
		i = j
	}
}

// printStats 는 구간의 호출 수, 해석 비율, 부품(디렉터리)별 분포, 많이 불린 함수를 출력합니다.
func printStats(w io.Writer, s *symbols, st *stats, find []string) {
	fmt.Fprintf(w, "calls=%d distinct=%d max_depth=%d duration_ms=%.3f unresolved=%d unwound=%d orphan_exits=%d dropped=%d\n",
		st.Calls, len(st.ByFn), st.MaxDepth, float64(st.EndNs-st.StartNs)/1e6, st.Unresolved, st.Unwound, st.OrphanExits, st.Dropped)

	// 부품 열쇠별 호출 수와, 헤더(src/include)에 정의된 함수의 몫을 셉니다.
	byMod := map[string]int{}
	header, nofile := 0, 0
	for fn, c := range st.ByFn {
		f := s.file(fn)
		byMod[module(f)] += c
		if strings.HasPrefix(f, "src/include/") {
			header += c
		}
		if f == "" {
			nofile += c
		}
	}
	fmt.Fprintf(w, "header_defined_calls=%d (%.1f%%) no_decl_file_calls=%d (%.1f%%)\n",
		header, pct(header, st.Calls), nofile, pct(nofile, st.Calls))
	type kv struct {
		K string
		V int
	}
	var mods []kv
	for k, v := range byMod {
		mods = append(mods, kv{k, v})
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].V > mods[j].V })
	for _, m := range mods {
		fmt.Fprintf(w, "MODULE %-28s %7d %5.1f%%\n", m.K, m.V, pct(m.V, st.Calls))
	}

	var fns []kv
	for fn, c := range st.ByFn {
		fns = append(fns, kv{s.name(fn), c})
	}
	sort.Slice(fns, func(i, j int) bool { return fns[i].V > fns[j].V || fns[i].V == fns[j].V && fns[i].K < fns[j].K })
	for i := 0; i < len(fns) && i < 15; i++ {
		fmt.Fprintf(w, "TOP %-40s %6d\n", fns[i].K, fns[i].V)
	}

	// 이름을 미리 적지 않은 함수가 기록에 나오는지 확인합니다(P18).
	for _, want := range find {
		n := 0
		for fn, c := range st.ByFn {
			if s.name(fn) == want {
				n += c
			}
		}
		fmt.Fprintf(w, "FIND %s=%d\n", want, n)
	}
}

// pct 는 a/b를 백분율로 돌려줍니다. b가 0이면 0입니다.
func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return 100 * float64(a) / float64(b)
}

// loadAll 은 디렉터리 안의 모든 <pid>.bin을 읽습니다.
func loadAll(dir string, syms *symbols) ([]*traceFile, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.bin"))
	if err != nil {
		return nil, err
	}
	var out []*traceFile
	for _, p := range paths {
		t, err := readTrace(p, syms.mainAt)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Pid < out[j].Pid })
	return out, nil
}

// cmdSummary 는 프로세스마다 레코드 수, 첫 함수, 시각 범위를 출력합니다.
func cmdSummary(w io.Writer, traces []*traceFile, syms *symbols) {
	for _, t := range traces {
		enters, exits, dropped := 0, 0, 0
		first := ""
		for _, r := range t.Records {
			switch r.Kind {
			case kindEnter:
				enters++
				if first == "" {
					first = syms.name(r.Fn - t.Slide)
				}
			case kindExit:
				exits++
			case kindDropped:
				dropped += int(r.Fn)
			}
		}
		last := t.Records[len(t.Records)-1].Ns
		fmt.Fprintf(w, "PID %d slide=0x%x enters=%d exits=%d dropped=%d first_enter=%s from_ns=%d to_ns=%d\n",
			t.Pid, t.Slide, enters, exits, dropped, first, t.Records[0].Ns, last)
	}
}

// findTrace 는 pid의 기록을 고릅니다.
func findTrace(traces []*traceFile, pid int) (*traceFile, error) {
	for _, t := range traces {
		if t.Pid == pid {
			return t, nil
		}
	}
	return nil, fmt.Errorf("pid %d의 기록이 없습니다", pid)
}

// flatNode 는 그림 시안에 넘길 호출 나무의 한 칸입니다. 시각은 쿼리 시작부터의 마이크로초입니다.
type flatNode struct {
	Depth   int     `json:"d"`
	Name    string  `json:"fn"`
	Module  string  `json:"mod"`
	Calls   int     `json:"calls"`
	StartUs float64 `json:"t0"`
	EndUs   float64 `json:"t1"`
	From    string  `json:"from"`
}

// flatten 은 호출 나무를 maxDepth 단계까지 시간 순서의 목록으로 펼칩니다.
func flatten(s *symbols, n *node, depth, maxDepth int, base uint64, out []flatNode) []flatNode {
	out = append(out, flatNode{Depth: depth, Name: s.name(n.Fn), Module: module(s.file(n.Fn)), Calls: n.Calls,
		StartUs: float64(n.StartNs-base) / 1e3, EndUs: float64(n.EndNs-base) / 1e3, From: s.lineAt(n.Site - 4)})
	if depth+1 < maxDepth {
		for _, c := range n.Children {
			out = flatten(s, c, depth+1, maxDepth, base, out)
		}
	}
	return out
}

// entry 는 흐름이 keep 모듈 하나에 들어간 순간입니다. Stack은 바깥(exec_simple_query)부터 그 함수까지입니다.
type entry struct {
	Module  string   `json:"mod"`
	StartUs float64  `json:"t0"`
	EndUs   float64  `json:"t1"`
	Calls   int      `json:"calls"`
	From    string   `json:"from"`
	Stack   []string `json:"stack"`
}

// entries 는 keep 밖의 함수가 keep 모듈의 함수를 부른 순간, 곧 그 부품 묶음에 처음 들어간 순간을 깊이와 관계없이 모읍니다.
// keep 안의 모듈끼리 오가는 호출(예: 버퍼에서 파일로)은 그 순간의 아래 칸이므로 따로 모으지 않습니다.
func entries(s *symbols, n *node, parentMod string, stack []string, keep map[string]bool, base uint64, out []entry) []entry {
	mod := module(s.file(n.Fn))
	stack = append(stack, s.name(n.Fn))
	if keep[mod] && !keep[parentMod] {
		out = append(out, entry{Module: mod, StartUs: float64(n.StartNs-base) / 1e3, EndUs: float64(n.EndNs-base) / 1e3,
			Calls: n.Calls, From: s.lineAt(n.Site - 4), Stack: append([]string(nil), stack...)})
	}
	for _, c := range n.Children {
		out = entries(s, c, mod, stack, keep, base, out)
	}
	return out
}

// cmdQuery 는 pid의 nth번째 exec_simple_query 호출 하나를 나무로 만들어 요약합니다.
// jsonDepth가 0보다 크면 요약 대신 그 깊이까지 펼친 나무와, keep 모듈에 들어간 순간들을 JSON으로 출력합니다.
func cmdQuery(w io.Writer, t *traceFile, syms *symbols, nth int, find []string, treeDepth, jsonDepth int, keep map[string]bool) error {
	seen := 0
	for i, r := range t.Records {
		if r.Kind != kindEnter || syms.name(r.Fn-t.Slide) != "exec_simple_query" {
			continue
		}
		seen++
		if seen != nth {
			continue
		}
		st := buildTree(t, i, syms)
		if jsonDepth > 0 {
			return json.NewEncoder(w).Encode(map[string]any{
				"tree":    flatten(syms, st.Root, 0, jsonDepth, st.StartNs, nil),
				"entries": entries(syms, st.Root, "", nil, keep, st.StartNs, nil),
			})
		}
		fmt.Fprintf(w, "QUERY pid=%d nth=%d start_ns=%d end_ns=%d\n", t.Pid, nth, st.StartNs, st.EndNs)
		printStats(w, syms, st, find)
		if treeDepth > 0 {
			printTree(w, syms, st.Root, 0, treeDepth)
		}
		return nil
	}
	return fmt.Errorf("pid %d에 exec_simple_query가 %d번 없습니다(%d번 있음)", t.Pid, nth, seen)
}

// cmdWindow 는 pid가 [from, to] 시각 안에 들어간 함수를 부품별로 세고, storage 아래 함수 이름을 보입니다.
func cmdWindow(w io.Writer, t *traceFile, syms *symbols, from, to uint64) {
	byMod := map[string]int{}
	storage := map[string]int{}
	n := 0
	for _, r := range t.Records {
		if r.Kind != kindEnter || r.Ns < from || r.Ns > to {
			continue
		}
		fn := r.Fn - t.Slide
		m := module(syms.file(fn))
		byMod[m]++
		n++
		if strings.HasPrefix(m, "storage/") {
			storage[syms.name(fn)+" "+m]++
		}
	}
	fmt.Fprintf(w, "WINDOW pid=%d calls=%d\n", t.Pid, n)
	keys := make([]string, 0, len(byMod))
	for k := range byMod {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "MODULE %-28s %7d\n", k, byMod[k])
	}
	keys = keys[:0]
	for k := range storage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "STORAGE %s %d\n", k, storage[k])
	}
}

// main 은 실행 파일과 기록 디렉터리를 읽은 뒤 하위 명령 하나를 실행합니다. 잘못된 사용은 종료 코드 2, 읽기 실패는 1입니다.
func main() {
	bin := flag.String("bin", "", "기록을 남긴 실행 파일(DWARF 포함)")
	dir := flag.String("dir", "", "<pid>.bin 파일들이 있는 디렉터리")
	flag.Parse()

	// spring-compare는 해석기의 출력만 읽으므로 실행 파일과 함수 기록이 필요 없습니다.
	if flag.Arg(0) == "spring-compare" {
		sub := flag.NewFlagSet("spring-compare", flag.ExitOnError)
		out := sub.String("out", "", "spring-compare: 덧붙인 출력을 쓸 디렉터리")
		if err := sub.Parse(flag.Args()[1:]); err != nil {
			os.Exit(2)
		}
		if err := cmdSpringCompare(*out, sub.Args()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *bin == "" || *dir == "" || flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "사용: decode -bin FILE -dir DIR summary|query|window|journey|scenario|spring [옵션] 또는 decode spring-compare -out DIR 이름=경로 ...")
		os.Exit(2)
	}
	syms, err := loadSymbols(*bin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	traces, err := loadAll(*dir, syms)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 하위 명령마다 자기 옵션만 받도록 FlagSet을 따로 둡니다.
	sub := flag.NewFlagSet(flag.Arg(0), flag.ExitOnError)
	pid := sub.Int("pid", 0, "프로세스 번호")
	nth := sub.Int("nth", 1, "몇 번째 exec_simple_query인가")
	findList := sub.String("find", "", "쉼표로 나눈 함수 이름들")
	tree := sub.Int("tree", 0, "호출 나무를 출력할 깊이(0이면 출력하지 않음)")
	jsonDepth := sub.Int("json", 0, "이 깊이까지 펼친 나무를 JSON으로 출력(0이면 요약)")
	keepList := sub.String("keep", "", "-json에서 들어간 순간을 깊이와 관계없이 모을 모듈들(쉼표로 나눔)")
	from := sub.String("from", "0", "구간 시작 나노초")
	to := sub.String("to", "0", "구간 끝 나노초")
	stracePath := sub.String("strace", "", "journey: strace -f -yy -ttt 출력 파일")
	srcDir := sub.String("src", "", "journey: 소스 발췌를 넣을 PostgreSQL 소스 트리(없으면 넣지 않음)")
	kernelPath := sub.String("kernel", "", "journey: ftrace function_graph 출력(없으면 커널 나무를 붙이지 않음)")
	ksrcDir := sub.String("ksrc", "", "journey: 커널 함수의 정의를 찾을 원본 커널 소스 트리")
	stepsPath := sub.String("steps", "", "scenario: txlab의 steps.json")
	uprobePath := sub.String("uprobe", "", "scenario: txlab 인스턴스의 uprobe 기록")
	logPath := sub.String("log", "", "scenario: 기록 서버의 server.log(없으면 로그 사건을 넣지 않음)")
	recDir := sub.String("rec", "", "spring: txlab record의 출력 디렉터리(-dir은 그 아래 trace)")
	jfrTool := sub.String("jfr", "jfr", "spring: JDK의 jfr 명령")
	jsrcList := sub.String("jsrc", "", "spring: Java 소스 뿌리들(쉼표로 나눔, 그 아래에 패키지 경로). 앱 코드와 라이브러리 소스 발췌에 씀")
	if err := sub.Parse(flag.Args()[1:]); err != nil {
		os.Exit(2)
	}
	var find []string
	if *findList != "" {
		find = strings.Split(*findList, ",")
	}
	keep := map[string]bool{}
	if *keepList != "" {
		for _, m := range strings.Split(*keepList, ",") {
			keep[m] = true
		}
	}

	switch flag.Arg(0) {
	case "summary":
		cmdSummary(os.Stdout, traces, syms)
	case "query":
		t, err := findTrace(traces, *pid)
		if err == nil {
			err = cmdQuery(os.Stdout, t, syms, *nth, find, *tree, *jsonDepth, keep)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "window":
		t, err := findTrace(traces, *pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		f, err1 := strconv.ParseUint(*from, 10, 64)
		e, err2 := strconv.ParseUint(*to, 10, 64)
		if err1 != nil || err2 != nil {
			fmt.Fprintln(os.Stderr, "-from과 -to는 나노초 정수여야 합니다")
			os.Exit(2)
		}
		cmdWindow(os.Stdout, t, syms, f, e)
	case "scenario":
		sys, err := readStrace(*stracePath)
		var life []lifeEvent
		if err == nil {
			life, err = readLifecycle(*stracePath)
		}
		var logs []logLine
		if err == nil {
			logs, err = readServerLog(*logPath)
		}
		if err == nil {
			err = cmdScenario(os.Stdout, traces, syms, sys, life, logs, *stepsPath, *uprobePath, *srcDir)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "spring":
		var jsrc []string
		if *jsrcList != "" {
			jsrc = strings.Split(*jsrcList, ",")
		}
		if err := cmdSpring(os.Stdout, traces, syms, *recDir, *jfrTool, *srcDir, jsrc); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "journey":
		sys, err := readStrace(*stracePath)
		var kern kernelInput
		if err == nil && *kernelPath != "" {
			kern.Calls, err = readKernelTrace(*kernelPath)
		}
		if err == nil && *ksrcDir != "" {
			kern.Src = *ksrcDir
			kern.Index, err = buildKernelIndex(*ksrcDir)
		}
		if err == nil {
			err = cmdJourney(os.Stdout, traces, syms, sys, *pid, *nth, *srcDir, kern)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "알 수 없는 하위 명령 %q\n", flag.Arg(0))
		os.Exit(2)
	}
}
