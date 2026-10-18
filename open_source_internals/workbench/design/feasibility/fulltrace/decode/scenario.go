package main

// 이 파일은 여러 세션의 트랜잭션 시나리오 하나를 해석합니다. 탐색용 코드입니다(2026-10-14, 트랜잭션 관측 실험 1~5).
//
// 재료는 네 가지입니다. txlab의 steps.json(단계마다 보낸 SQL, 시각, psql 출력, 관찰용 접속이 읽은 페이지·세션·락),
// uprobe.trace(OrbStack 머신의 ftrace가 남긴 함수 인자와 반환값), 함수 기록(<pid>.bin, 모든 함수의 입구·출구와 부른 자리),
// strace(소켓과 시스템 콜). uprobe 사건은 pid마다 같은 함수의 몇 번째 호출인지로 함수 기록과 짝짓습니다(시계가 달라도 맞음).
// 짝지은 사건은 함수 기록의 시각(CLOCK_REALTIME)과 호출 스택·부른 줄을 얻습니다.

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// uevent 는 uprobe 사건 하나입니다. Func는 걸린 함수의 실행 때 주소이고, Return은 반환 사건인지입니다.
// RetAddr는 반환 사건에만 있는 원래 돌아갈 주소입니다(repairSites가 씀).
// Pid는 사건마다 읽은 MyProcPid(컨테이너 안 pid), Kpid는 커널의 전역 pid입니다.
type uevent struct {
	Name    string
	Kpid    int
	Ts      float64
	Func    uint64
	Return  bool
	RetAddr uint64
	F       map[string]string
	Pid     int
	// 아래는 짝짓기 뒤에 채웁니다.
	rec   int
	ns    uint64
	trace *traceFile
}

// ueventLine 은 ftrace의 사건 줄입니다. 반환 사건은 "(돌아갈 주소 <- 함수 주소)"입니다.
var ueventLine = regexp.MustCompile(`^\s*(.+?)-(\d+)\s+\[\d+\]\s+\S+\s+([\d.]+):\s+(\w+):\s+\((0x[0-9a-f]+)(?: <- (0x[0-9a-f]+))?\)\s*(.*)$`)

// readUprobes 는 txlab 인스턴스의 trace 파일을 읽습니다. 문자열 필드는 sql="..." 꼴이라 공백을 품을 수 있습니다.
//
// Errors: 파일을 읽지 못하면 오류를 돌려줍니다. 꼴이 다른 줄(머리 주석 등)은 건너뜁니다.
func readUprobes(path string) ([]uevent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []uevent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		m := ueventLine.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		e := uevent{Name: m[4], F: parseFields(m[7])}
		e.Kpid, _ = strconv.Atoi(m[2])
		e.Ts, _ = strconv.ParseFloat(m[3], 64)
		addr := m[5]
		if m[6] != "" {
			e.RetAddr, _ = strconv.ParseUint(m[5][2:], 16, 64)
			addr, e.Return = m[6], true
		}
		e.Func, _ = strconv.ParseUint(addr[2:], 16, 64)
		e.Pid, _ = strconv.Atoi(e.F["pid"])
		out = append(out, e)
	}
	return out, sc.Err()
}

// parseFields 는 "k=v k2="공백 있는 값"" 꼴을 풉니다.
func parseFields(s string) map[string]string {
	out := map[string]string{}
	for s = strings.TrimSpace(s); s != ""; s = strings.TrimSpace(s) {
		eq := strings.IndexByte(s, '=')
		if eq < 0 {
			break
		}
		k := s[:eq]
		s = s[eq+1:]
		if strings.HasPrefix(s, `"`) {
			end := strings.Index(s[1:], `"`)
			if end < 0 {
				out[k] = s[1:]
				break
			}
			out[k] = s[1 : end+1]
			s = s[end+2:]
			continue
		}
		sp := strings.IndexByte(s, ' ')
		if sp < 0 {
			out[k] = s
			break
		}
		out[k] = s[:sp]
		s = s[sp:]
	}
	return out
}

// joinCount 는 pid 하나, 함수 하나, 입구·출구 하나의 짝짓기 결과입니다. 두 수가 같아야 짝짓기를 믿습니다.
type joinCount struct {
	Pid      int    `json:"pid"`
	Fn       string `json:"fn"`
	Return   bool   `json:"return,omitempty"`
	Uprobe   int    `json:"uprobe"`
	Recorder int    `json:"recorder"`
}

// joinUprobes 는 uprobe 사건을 함수 기록과 짝짓습니다. pid마다, 함수마다, 입구·출구마다 k번째 사건을 k번째 기록에 대고,
// 짝지은 사건에 기록의 위치(rec)와 시각(ns)을 채웁니다. 수가 다른 묶음은 짝짓지 않고 결과 목록에 남깁니다.
func joinUprobes(evs []uevent, traces []*traceFile, syms *symbols) []joinCount {
	type key struct {
		pid int
		fn  uint64
		ret bool
	}
	byKey := map[key][]int{}
	for i, e := range evs {
		byKey[key{e.Pid, e.Func, e.Return}] = append(byKey[key{e.Pid, e.Func, e.Return}], i)
	}
	var out []joinCount
	for _, t := range traces {
		recs := map[key][]int{}
		for i, r := range t.Records {
			if r.Kind == kindEnter || r.Kind == kindExit {
				k := key{t.Pid, r.Fn, r.Kind == kindExit}
				if _, want := byKey[k]; want {
					recs[k] = append(recs[k], i)
				}
			}
		}
		for k, idx := range byKey {
			if k.pid != t.Pid {
				continue
			}
			jc := joinCount{Pid: t.Pid, Fn: syms.name(k.fn - t.Slide), Return: k.ret, Uprobe: len(idx), Recorder: len(recs[k])}
			out = append(out, jc)
			if jc.Uprobe != jc.Recorder {
				continue
			}
			for n, i := range idx {
				evs[i].rec = recs[k][n]
				evs[i].ns = t.Records[recs[k][n]].Ns
				evs[i].trace = t
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pid != out[j].Pid {
			return out[i].Pid < out[j].Pid
		}
		if out[i].Fn != out[j].Fn {
			return out[i].Fn < out[j].Fn
		}
		return !out[i].Return
	})
	return out
}

// siteRepair 는 repairSites가 고친 부른 자리의 수입니다. Before는 고치기 전 값마다의 수입니다.
type siteRepair struct {
	Repaired  int            `json:"repaired"`
	Unchanged int            `json:"unchanged"`
	Before    map[string]int `json:"before"`
	Left      int            `json:"leftAfter"`
	LeftFns   map[string]int `json:"leftFns,omitempty"`
}

// repairSites 는 반환 uprobe가 바꿔 놓은 함수 기록의 부른 자리를 원래 값으로 되돌립니다.
//
// 반환 uprobe(%return)를 건 함수는 입구에서 커널이 돌아갈 주소(arm64의 LR)를 uprobe trampoline 주소로 바꿉니다
// (arch/arm64/kernel/probes/uprobes.c의 arch_uretprobe_hijack_return_addr). 함수 기록은 그 뒤 함수 머리에서
// 돌아갈 주소를 부른 자리로 읽으므로, 그 함수들의 입구·출구 기록에는 trampoline 주소가 남습니다.
// ftrace의 반환 사건 줄 "(돌아갈 주소 <- 함수)"에는 커널이 보관한 원래 주소가 있으므로, 짝지은 반환 사건마다
// 그 출구 기록과 짝이 되는 입구 기록의 부른 자리를 그 값으로 바꿉니다.
//
// 입구와 출구의 짝은 기록 순서의 스택으로 찾습니다. elog(ERROR)의 longjmp가 건너뛴 출구는 스택에서 버립니다.
// Returns: 고친 수, 이미 같던 수, 고치기 전 값의 분포, 고친 뒤에도 고치기 전 값(trampoline 주소)이 남은 기록의 수.
// 마지막 수가 0이 아니면 반환 사건이 없는 호출이 있다는 뜻이고(짝짓기 수가 다른 묶음, 또는 오류의 longjmp로 반환하지
// 않은 호출), 그 기록의 부른 줄은 믿을 수 없습니다. LeftFns가 그 함수들입니다.
// 실행 파일 밖을 가리키는 부른 자리가 모두 오류는 아닙니다. libc가 부르는 함수(qsort의 비교 함수, 신호 처리기,
// atexit 함수, main)는 원래 libc 안을 가리킵니다(2026-10-14 실험 1-rc-read에서 남은 767개가 모두 이 넷).
func repairSites(evs []uevent, traces []*traceFile, syms *symbols) siteRepair {
	out := siteRepair{Before: map[string]int{}, LeftFns: map[string]int{}}
	for _, t := range traces {
		// 출구 기록마다 짝이 되는 입구 기록을 찾습니다.
		enterOf := map[int]int{}
		var stack []int
		for i, r := range t.Records {
			switch r.Kind {
			case kindEnter:
				stack = append(stack, i)
			case kindExit:
				for k := len(stack) - 1; k >= 0; k-- {
					if t.Records[stack[k]].Fn == r.Fn {
						enterOf[i] = stack[k]
						stack = stack[:k]
						break
					}
				}
			}
		}
		for _, e := range evs {
			if e.trace != t || !e.Return || e.RetAddr == 0 {
				continue
			}
			ex := e.rec
			for _, i := range []int{ex, enterOf[ex]} {
				if _, ok := enterOf[ex]; !ok && i != ex {
					continue
				}
				if t.Records[i].Site == e.RetAddr {
					out.Unchanged++
					continue
				}
				out.Before[fmt.Sprintf("0x%x", t.Records[i].Site)]++
				t.Records[i].Site = e.RetAddr
				out.Repaired++
			}
		}
	}
	for _, t := range traces {
		for _, r := range t.Records {
			if _, bad := out.Before[fmt.Sprintf("0x%x", r.Site)]; bad && (r.Kind == kindEnter || r.Kind == kindExit) {
				out.Left++
				out.LeftFns[syms.name(r.Fn-t.Slide)]++
			}
		}
	}
	return out
}

// tableMeta 는 관찰하는 표 하나의 이름, 번호(oid), 열입니다.
type tableMeta struct {
	Name    string `json:"name"`
	Oid     int    `json:"oid"`
	Columns []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns"`
	// 실험 4부터 txlab이 남깁니다. SIRead 락의 대상이 인덱스 페이지일 때 그 이름을 붙입니다.
	Indexes []struct {
		Name string `json:"name"`
		Oid  int    `json:"oid"`
	} `json:"indexes"`
}

// stepsFile 은 txlab의 steps.json 가운데 해석에 쓰는 부분입니다.
type stepsFile struct {
	Title    string      `json:"title"`
	Scenario string      `json:"scenario"`
	Tables   []tableMeta `json:"tables"`
	Pauses   []struct {
		BeforeStep int    `json:"beforeStep"`
		StartNs    uint64 `json:"startNs"`
		EndNs      uint64 `json:"endNs"`
	} `json:"pauses"`
	Sessions []struct {
		Name string `json:"name"`
		Pid  int    `json:"pid"`
	} `json:"sessions"`
	Steps []struct {
		N         int      `json:"n"`
		Session   string   `json:"session"`
		SQL       string   `json:"sql"`
		SentNs    uint64   `json:"sentNs"`
		DoneNs    uint64   `json:"doneNs"`
		BlockedNs uint64   `json:"blockedNs"`
		Output    []string `json:"output"`
	} `json:"steps"`
	Kills []struct {
		BeforeStep int    `json:"beforeStep"`
		Session    string `json:"session"`
		Pid        int    `json:"pid"`
		Ns         uint64 `json:"ns"`
		ReadyNs    uint64 `json:"readyNs"`
	} `json:"kills"`
	Captures []struct {
		AfterStep int     `json:"afterStep"`
		AfterKill bool    `json:"afterKill"`
		Wal       *walPos `json:"wal"`
		Ns        uint64  `json:"ns"`
		Pages     map[string][]struct {
			Block     int      `json:"block"`
			Lp        int      `json:"lp"`
			Flags     int      `json:"lpFlags"`
			Xmin      string   `json:"xmin"`
			Xmax      string   `json:"xmax"`
			Ctid      string   `json:"ctid"`
			Infomask  string   `json:"infomask"`
			Infomask2 string   `json:"infomask2"`
			RawFlags  []string `json:"rawFlags"`
			Attrs     []string `json:"attrs"`
		} `json:"pages"`
		Activity json.RawMessage   `json:"activity"`
		Locks    json.RawMessage   `json:"locks"`
		Running  string            `json:"running"`
		Xacts    map[string]string `json:"xacts"`
	} `json:"captures"`
	ObserverPid  int    `json:"observerPid"`
	ObserverPids []int  `json:"observerPids"`
	StartNs      uint64 `json:"startNs"`
	EndNs        uint64 `json:"endNs"`
}

// walPos 는 관찰 한 번의 WAL 위치 셋(넣은 끝, 운영 체제에 쓴 끝, 디스크에 내린 끝)입니다. 실험 3부터 있습니다.
type walPos struct {
	Insert string `json:"insert"`
	Write  string `json:"write"`
	Flush  string `json:"flush"`
}

// 화면에 넘기는 모양입니다. 시각은 모두 첫 단계를 보낸 순간부터의 마이크로초입니다.

type txSnapshot struct {
	T      float64    `json:"t"`
	Kind   string     `json:"kind"`
	Xmin   int64      `json:"xmin"`
	Xmax   int64      `json:"xmax"`
	Xip    []int64    `json:"xip"`
	Curcid int64      `json:"curcid"`
	Stack  []frameRef `json:"stack"`
}

// txDecision 은 줄 판정이 거친 함수 호출 하나입니다. Inner는 그 호출이 직접 부른 함수와 부른 줄입니다(함수 기록).
// 이 빌드(-O0)는 인라인 도우미(TransactionIdPrecedes 등)도 실제 함수로 남아 기록되므로, Inner의 마지막 줄이 그 호출의
// 반환값을 정한 비교에 가깝습니다. 반환을 정한 줄 자체는 기록되지 않으므로 화면은 "마지막으로 부른 줄"로만 씁니다.
type txDecision struct {
	Fn    string     `json:"fn"`
	At    string     `json:"at"`
	Arg   string     `json:"arg,omitempty"`
	Ret   string     `json:"ret,omitempty"`
	Inner []frameRef `json:"inner,omitempty"`
}

type txCheck struct {
	T       float64      `json:"t"`
	Table   string       `json:"table"`
	Ctid    string       `json:"ctid"`
	Xmin    int64        `json:"xmin"`
	Xmax    int64        `json:"xmax"`
	Mask    int64        `json:"mask"`
	SnapMin int64        `json:"snapXmin"`
	SnapMax int64        `json:"snapXmax"`
	Visible bool         `json:"visible"`
	Tup     string       `json:"tup,omitempty"`
	Path    []txDecision `json:"path"`
	Lines   []frameRef   `json:"lines"`
	Stack   []frameRef   `json:"stack"`
}

type txEvent struct {
	T     float64           `json:"t"`
	Kind  string            `json:"kind"`
	F     map[string]string `json:"f,omitempty"`
	Stack []frameRef        `json:"stack,omitempty"`
	T1    float64           `json:"t1,omitempty"`
}

type txRequest struct {
	SQL       string       `json:"sql"`
	T0        float64      `json:"t0"`
	T1        float64      `json:"t1"`
	Aborted   bool         `json:"aborted,omitempty"`
	Snapshots []txSnapshot `json:"snapshots"`
	Catalog   int          `json:"catalogSnapshots"`
	Checks    []txCheck    `json:"checks"`
	Events    []txEvent    `json:"events"`
	In        []pgMessage  `json:"in"`
	Out       []pgMessage  `json:"out"`
}

type txStep struct {
	N        int         `json:"n"`
	Session  string      `json:"session"`
	SQL      string      `json:"sql"`
	T0       float64     `json:"t0"`
	T1       float64     `json:"t1"`
	Blocked  float64     `json:"blocked,omitempty"`
	Output   []string    `json:"output"`
	Requests []txRequest `json:"requests"`
	Check    *stepCheck  `json:"check,omitempty"`
}

type txVersion struct {
	Block  int               `json:"block"`
	Lp     int               `json:"lp"`
	LpFlag int               `json:"lpFlags"`
	Ctid   string            `json:"ctid"`
	Xmin   string            `json:"xmin"`
	Xmax   string            `json:"xmax"`
	Flags  []string          `json:"flags"`
	Values map[string]string `json:"values"`
}

type txCapture struct {
	AfterStep int                    `json:"afterStep"`
	AfterKill bool                   `json:"afterKill,omitempty"`
	Wal       *walPos                `json:"wal,omitempty"`
	T         float64                `json:"t"`
	Pages     map[string][]txVersion `json:"pages"`
	Activity  json.RawMessage        `json:"activity"`
	Locks     json.RawMessage        `json:"locks"`
	Running   string                 `json:"running"`
	Xacts     map[string]string      `json:"xacts"`
	// Reads 는 부분마다 관찰용 접속이 실제로 읽은 구간입니다(captureReads). T는 묻기 시작한 시각일 뿐입니다.
	Reads map[string][2]float64 `json:"reads,omitempty"`
}

// cmdScenario 는 시나리오 기록 하나를 화면이 그릴 JSON으로 바꿉니다.
//
// Arguments: stepsPath는 txlab의 steps.json, uprobePath는 uprobe.trace, sys는 strace 사건, life는 strace의 프로세스 끝 줄,
// logs는 server.log 항목, srcDir가 있으면 소스 발췌를 붙입니다.
// Errors: 파일을 읽지 못하면 오류를 돌려줍니다. 짝짓기 수가 다른 함수가 있어도 끝까지 해석하고 "joins"에 수를 남깁니다.
// 수가 다른 함수의 사건은 시각과 스택이 없으므로 결과에서 빠집니다.
func cmdScenario(w io.Writer, traces []*traceFile, syms *symbols, sys []sysEvent, life []lifeEvent, logs []logLine, stepsPath, uprobePath, srcDir string) error {
	var st stepsFile
	b, err := os.ReadFile(stepsPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return err
	}
	evs, err := readUprobes(uprobePath)
	if err != nil {
		return err
	}
	joins := joinUprobes(evs, traces, syms)
	repair := repairSites(evs, traces, syms)
	base := st.Steps[0].SentNs
	rel := func(ns uint64) float64 { return float64(int64(ns-base)) / 1e3 }

	clock := uprobeClock(evs)

	oids := map[int]string{}
	// ssiRels는 술어 락의 대상 이름입니다. 표 말고 인덱스도 들어가므로 oids(행 판정과 락 사건의 표)와 따로 둡니다.
	ssiRels := map[int]string{}
	for _, t := range st.Tables {
		oids[t.Oid] = t.Name
		ssiRels[t.Oid] = t.Name
		for _, ix := range t.Indexes {
			ssiRels[ix.Oid] = ix.Name
		}
	}
	pidOf := map[string]int{}
	for _, s := range st.Sessions {
		pidOf[s.Name] = s.Pid
	}
	byPid := map[int][]*uevent{}
	for i := range evs {
		if evs[i].trace != nil {
			byPid[evs[i].Pid] = append(byPid[evs[i].Pid], &evs[i])
		}
	}
	for _, l := range byPid {
		sort.SliceStable(l, func(i, j int) bool { return l[i].ns < l[j].ns })
	}
	// 스택은 기록을 앞으로 한 번 따라가며 구하므로, pid마다 시각 순서로 묻습니다.
	samplers := map[int]func(uint64, bool) ([]frameRef, string){}
	for _, t := range traces {
		samplers[t.Pid] = stackSampler(t, syms)
	}
	stackAt := func(e *uevent) []frameRef {
		s, _ := samplers[e.Pid](e.ns, !e.Return)
		return trimStack(s)
	}
	fname := func(e *uevent) string { return syms.name(e.Func - e.trace.Slide) }
	// 부른 줄: 함수 기록의 입구 레코드가 부른 자리 주소를 갖습니다. 반환 사건은 같은 호출의 입구가 부른 자리를 씁니다.
	siteOf := func(e *uevent) string {
		r := e.trace.Records[e.rec]
		return syms.lineAt(r.Site - e.trace.Slide - 4)
	}

	var steps []txStep
	for i, s := range st.Steps {
		pid := pidOf[s.Session]
		end := s.DoneNs
		if i+1 < len(st.Steps) && end == 0 {
			end = st.Steps[i+1].SentNs
		}
		ts := txStep{N: s.N, Session: s.Session, SQL: s.SQL, T0: rel(s.SentNs), T1: rel(s.DoneNs), Output: s.Output}
		if s.BlockedNs != 0 {
			ts.Blocked = rel(s.BlockedNs)
		}
		// 단계 안의 요청: 그 세션 백엔드의 exec_simple_query 입구가 단계 구간 안에 있는 것.
		t, err := findTrace(traces, pid)
		if err != nil {
			return err
		}
		var windows [][2]uint64
		var sqls []string
		for _, e := range byPid[pid] {
			if fname(e) == "exec_simple_query" && !e.Return && e.ns >= s.SentNs && e.ns <= end {
				windows = append(windows, [2]uint64{e.ns, 0})
				sqls = append(sqls, e.F["sql"])
			}
			// 반환은 단계 구간 안의 것만 받습니다. 오류(elog ERROR)로 끝난 요청은 longjmp로 빠져나가 반환 사건이 없으므로,
			// 구간 밖의 다음 요청 반환이 그 요청을 닫으면 안 됩니다.
			if fname(e) == "exec_simple_query" && e.Return && len(windows) > 0 && windows[len(windows)-1][1] == 0 && e.ns >= windows[len(windows)-1][0] && e.ns <= end {
				windows[len(windows)-1][1] = e.ns
			}
		}
		_ = t
		for k, win := range windows {
			aborted := win[1] == 0
			if aborted {
				win[1] = end
			}
			req := txRequest{SQL: sqls[k], T0: rel(win[0]), T1: rel(win[1]), Aborted: aborted, Snapshots: []txSnapshot{}, Checks: []txCheck{}, Events: []txEvent{}}
			decodeRequest(&req, byPid[pid], win, oids, ssiRels, fname, siteOf, stackAt, rel, syms)
			for _, e := range sys {
				if e.Pid != pid || e.Ns < win[0] || e.Ns > win[1]+5e6 {
					continue
				}
				switch {
				case e.Name == "recvfrom" && e.Ret > 0 && e.Ns <= win[0]:
				case e.Name == "sendto" && e.Ret > 0 && e.Ns <= win[1]+5e6:
					req.Out = append(req.Out, parseMessages(e.Data, false)...)
				case e.Name == "kill":
					req.Events = append(req.Events, txEvent{T: rel(e.Ns), Kind: "kill", F: map[string]string{"args": e.Args, "ret": fmt.Sprint(e.Ret)}})
				case e.Name == "fdatasync" || e.Name == "fsync" || (e.Name == "pwrite64" || e.Name == "pwritev") && strings.Contains(e.FdPath, "/pg_wal/"):
					// 디스크로 내려보내기(WAL 쓰기와 동기화). 커밋은 이것이 끝난 뒤에야 커밋 로그를 적고 기다리던 쪽을 깨웁니다.
					// WAL 쓰기는 쓴 범위를 LSN으로도 남깁니다(walIOFields). 그 범위가 커밋 레코드를 덮는지가 실험 3의 판정입니다.
					req.Events = append(req.Events, txEvent{T: rel(e.Ns - uint64(e.DurNs)), T1: rel(e.Ns), Kind: "io", F: walIOFields(e)})
				case e.Name == "epoll_pwait" && e.DurNs > 1e6:
					req.Events = append(req.Events, txEvent{T: rel(e.Ns - uint64(e.DurNs)), T1: rel(e.Ns), Kind: "sleep_syscall", F: map[string]string{"durUs": fmt.Sprint(e.DurNs / 1e3)}})
				}
			}
			// 받은 요청은 그 요청 시작 직전의 recvfrom입니다.
			for j := len(sys) - 1; j >= 0; j-- {
				e := sys[j]
				if e.Pid == pid && e.Name == "recvfrom" && e.Ret > 0 && e.Ns <= win[0] {
					req.In = parseMessages(e.Data, true)
					break
				}
			}
			sort.SliceStable(req.Events, func(a, b int) bool { return req.Events[a].T < req.Events[b].T })
			ts.Requests = append(ts.Requests, req)
		}
		steps = append(steps, ts)
	}

	// 관찰: 페이지의 행마다 값을 열 형식으로 풉니다.
	cols := map[string][][2]string{}
	for _, t := range st.Tables {
		for _, c := range t.Columns {
			cols[t.Name] = append(cols[t.Name], [2]string{c.Name, c.Type})
		}
	}
	var caps []txCapture
	// 관찰마다 부분을 실제로 읽은 구간. 끝을 모르는 마지막 관찰의 마지막 부분은 끝을 시작과 같게 둡니다.
	capNs := make([]uint64, len(st.Captures))
	for k, c := range st.Captures {
		capNs[k] = c.Ns
	}
	// 관찰용 접속은 steps.json의 observerPids(kill 뒤 새로 연 접속까지)입니다. 실험 2까지의 기록에는 observerPid 하나만 있고,
	// 실험 1의 기록에는 그것도 없어 페이지를 묻는 질의를 받은 첫 pid로 찾습니다. 접속이 여럿이면 시각 순으로 이어 붙입니다.
	obsPids := map[int]bool{}
	for _, p := range st.ObserverPids {
		obsPids[p] = true
	}
	if len(obsPids) == 0 && st.ObserverPid != 0 {
		obsPids[st.ObserverPid] = true
	}
	for _, e := range sys {
		if len(obsPids) > 0 {
			break
		}
		if e.Name == "recvfrom" && e.Ret > 0 && strings.Contains(queryText(e.Data), "heap_page_item_attrs") {
			obsPids[e.Pid] = true
		}
	}
	var obsSys []sysEvent
	for _, e := range sys {
		if obsPids[e.Pid] {
			obsSys = append(obsSys, e)
		}
	}
	reads := captureReads(capNs, obsSys)
	for k, c := range st.Captures {
		tc := txCapture{AfterStep: c.AfterStep, AfterKill: c.AfterKill, Wal: c.Wal, T: rel(c.Ns), Pages: map[string][]txVersion{}, Activity: c.Activity, Locks: c.Locks, Running: c.Running, Xacts: c.Xacts, Reads: map[string][2]float64{}}
		for part, w := range reads[k] {
			if w[1] == 0 {
				w[1] = w[0]
			}
			tc.Reads[part] = [2]float64{rel(w[0]), rel(w[1])}
		}
		for tbl, items := range c.Pages {
			for _, it := range items {
				v := txVersion{Block: it.Block, Lp: it.Lp, LpFlag: it.Flags, Ctid: it.Ctid, Xmin: it.Xmin, Xmax: it.Xmax, Flags: it.RawFlags, Values: map[string]string{}}
				for k, a := range it.Attrs {
					if k < len(cols[tbl]) {
						v.Values[cols[tbl][k][0]] = attrValue(a, cols[tbl][k][1])
					}
				}
				tc.Pages[tbl] = append(tc.Pages[tbl], v)
			}
		}
		caps = append(caps, tc)
	}

	// 단계마다 줄 판정을 직전 관찰의 페이지와 psql 결과에 견줍니다. 관찰하는 표가 여럿이면 그 단계에서 판정을 가장 많이 받은 표로
	// 견줍니다. psql 결과의 ctid는 어느 표의 것인지 적혀 있지 않고, 결과를 보는 SELECT는 표 하나만 읽기 때문입니다.
	for i := range steps {
		tbl := busiestTable(st.Tables, steps[i].Requests)
		var before []txVersion
		for _, c := range caps {
			if c.AfterStep == steps[i].N-1 {
				before = c.Pages[tbl]
			}
		}
		steps[i].Check = checkStep(tbl, before, steps[i].Requests, steps[i].Output)
	}
	sess := map[int]bool{}
	for _, p := range pidOf {
		sess[p] = true
	}
	var writes []hintWrite
	sessionOf := map[int]string{}
	for name, p := range pidOf {
		sessionOf[p] = name
	}
	for p, l := range byPid {
		for _, e := range l {
			if name, ok := sessionOf[p]; ok && fname(e) == "SetHintBits" && !e.Return {
				writes = append(writes, hintWrite{T: rel(e.ns), Session: name, Xmin: atoi(e.F["xmin"]), Xmax: atoi(e.F["xmax"]), Bit: hintName(atoi(e.F["bits"])), At: siteOf(e), Tup: e.F["tup"]})
			}
		}
	}
	// 행 머리의 주소는 공유 버퍼 안 자리라서, kill 뒤 서버가 다시 뜨면 같은 주소에 다른 페이지가 들어올 수 있습니다. 그래서 주소로
	// 행을 가리는 일(tupleAddrs, labelHints, hintChanges)은 kill과 kill 사이의 구간마다 따로 합니다.
	var cuts []int
	for _, k := range st.Kills {
		cuts = append(cuts, k.BeforeStep)
	}
	hints := []hintChange{}
	addrRows := []addrRow{}
	for e, g := range epochSteps(steps, cuts) {
		addrs := tupleAddrs(g)
		labelHints(g, addrs)
		for _, t := range st.Tables {
			hints = append(hints, hintChanges(g, caps, t.Name, writes, addrs[t.Name])...)
		}
		for _, r := range addrSummary(addrs) {
			r.Epoch = e
			addrRows = append(addrRows, r)
		}
	}

	// 시나리오의 kill. t는 SIGKILL을 보내기 직전, tReady는 txlab이 server.log에서 접속 재개 줄을 본 시각입니다(50 ms 간격).
	kills := []map[string]any{}
	for _, k := range st.Kills {
		kr := map[string]any{"beforeStep": k.BeforeStep, "session": k.Session, "pid": k.Pid, "t": rel(k.Ns), "tReady": rel(k.ReadyNs)}
		// walWritten은 kill 전에 커널에 넘긴 WAL의 끝, walSynced는 그중 동기화를 마친 끝입니다. SIGKILL은 프로세스만 죽이고
		// 커널의 페이지 캐시는 남기므로, 서버가 다시 뜬 뒤 복구가 읽을 수 있는 WAL은 앞의 것입니다. strace를 붙인 뒤 pg_wal 쓰기가
		// 한 번도 없었으면 두 값을 두지 않습니다.
		if w, sy, wNs := walExtentBefore(sys, k.Ns); w > 0 {
			kr["walWritten"], kr["walWrittenT"] = lsnText(fmt.Sprint(w)), rel(wNs)
			if sy > 0 {
				kr["walSynced"] = lsnText(fmt.Sprint(sy))
			}
		}
		kills = append(kills, kr)
	}
	// 시나리오의 쉼. 화면은 쉼이 든 자리(그 뒤에 온 단계)와 길이를 보입니다.
	pauses := []map[string]any{}
	for _, p := range st.Pauses {
		pauses = append(pauses, map[string]any{"beforeStep": p.BeforeStep, "t": rel(p.StartNs), "t1": rel(p.EndNs)})
	}
	out := map[string]any{
		"title": st.Title, "scenario": st.Scenario, "tables": st.Tables, "sessions": st.Sessions,
		"steps": steps, "captures": caps, "joins": joins, "clock": clock, "siteRepair": repair, "bystanders": bystanders(evs, sess, oids), "hintChanges": hints,
		"tupleAddrs": addrRows, "pauses": pauses, "kills": kills,
	}
	offset, _ := clock["medianOffsetUs"].(float64)
	procs, sevs, spans := serverView(serverInput{Traces: traces, Syms: syms, Evs: evs, Offset: offset * 1e3, Sys: sys, Life: life, Logs: logs,
		Sessions: sessionOf, Observers: obsPids, Win: [2]uint64{st.StartNs, st.EndNs}, Rel: rel})
	out["server"] = map[string]any{"procs": procs, "events": sevs, "spans": spans}
	if srcDir != "" {
		var refs []string
		for _, s := range steps {
			for _, r := range s.Requests {
				for _, sn := range r.Snapshots {
					for _, f := range sn.Stack {
						refs = append(refs, f.At)
					}
				}
				for _, c := range r.Checks {
					for _, p := range c.Path {
						refs = append(refs, p.At)
						for _, f := range p.Inner {
							refs = append(refs, f.At)
						}
					}
					for _, f := range c.Lines {
						refs = append(refs, f.At)
					}
					for _, f := range c.Stack {
						refs = append(refs, f.At)
					}
				}
				for _, e := range r.Events {
					for _, f := range e.Stack {
						refs = append(refs, f.At)
					}
				}
			}
		}
		for _, e := range sevs {
			for _, f := range e.Stack {
				refs = append(refs, f.At)
			}
		}
		for _, sp := range spans {
			refs = append(refs, sp.At)
		}
		out["src"] = excerpts(srcDir, refs)
	}
	return json.NewEncoder(w).Encode(out)
}

// uprobeClock 은 짝지은 uprobe 사건마다 ftrace 시각(tai 시계, 초)과 함수 기록 시각(나노초)의 차이를 봅니다. 중앙값은 짝짓지 못한
// 사건의 시각을 함수 기록 시계로 옮길 때 쓰고, 최대 편차는 그 옮김을 믿을 수 있는지의 근거입니다.
func uprobeClock(evs []uevent) map[string]any {
	var diffs []float64
	for _, e := range evs {
		if e.trace != nil {
			diffs = append(diffs, e.Ts*1e9-float64(e.ns))
		}
	}
	sort.Float64s(diffs)
	clock := map[string]any{"joined": len(diffs), "events": len(evs)}
	if len(diffs) > 0 {
		med := diffs[len(diffs)/2]
		maxDev := 0.0
		for _, d := range diffs {
			maxDev = math.Max(maxDev, math.Abs(d-med))
		}
		clock["medianOffsetUs"], clock["maxDeviationUs"] = med/1e3, maxDev/1e3
	}
	return clock
}

// requestEntries 는 요청 하나를 처리하기 시작하는 함수입니다. 단순 질의는 exec_simple_query 하나이고, 확장 질의(실험 5의 pgjdbc)는
// 메시지마다 exec_parse_message, exec_bind_message, exec_execute_message입니다.
var requestEntries = map[string]bool{"exec_simple_query": true, "exec_parse_message": true, "exec_bind_message": true, "exec_execute_message": true}

// trimStack 은 스택의 바깥쪽에서 요청 처리(requestEntries) 위를 잘라 냅니다. 그 위(main, PostmasterMain, BackendMain 등)는
// 모든 요청이 같아서 판단에 쓰이지 않습니다.
func trimStack(s []frameRef) []frameRef {
	for i, f := range s {
		if requestEntries[f.Fn] {
			return s[i:]
		}
	}
	return s
}

// tmResult, lockType, lockMode 는 18.6 헤더(access/tableam.h, storage/lock.h, storage/lockdefs.h)의 열거값 이름입니다.
// tupleMode, mxStatus, dlState 는 nodes/lockoptions.h의 LockTupleMode, access/multixact.h의 MultiXactStatus,
// storage/lock.h의 DeadLockState입니다.
var (
	tupleMode = []string{"KeyShare", "Share", "NoKeyExclusive", "Exclusive"}
	mxStatus  = []string{"ForKeyShare", "ForShare", "ForNoKeyUpdate", "ForUpdate", "NoKeyUpdate", "Update"}
	dlState   = []string{"NotYetChecked", "NoDeadlock", "SoftDeadlock", "HardDeadlock", "BlockedByAutovacuum"}
	tmResult  = []string{"TM_Ok", "TM_Invisible", "TM_SelfModified", "TM_Updated", "TM_Deleted", "TM_BeingModified", "TM_WouldBlock"}
	lockType  = []string{"relation", "extend", "frozenid", "page", "tuple", "transactionid", "virtualxid", "spectoken", "object", "userlock", "advisory", "applytransaction"}
	lockMode  = []string{"", "AccessShareLock", "RowShareLock", "RowExclusiveLock", "ShareUpdateExclusiveLock", "ShareLock", "ShareRowExclusiveLock", "ExclusiveLock", "AccessExclusiveLock"}
	xltwOper  = []string{"None", "Update", "Delete", "Lock", "LockUpdated", "InsertIndex", "InsertIndexUnique", "FetchUpdated", "RecheckExclusionConstr"}
	acquired  = []string{"NOT_AVAIL", "OK", "ALREADY_HELD", "ALREADY_CLEAR"}
)

func enumName(names []string, v string) string {
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 || n >= len(names) {
		return v
	}
	return names[n]
}

// decodeRequest 는 요청 구간 안의 uprobe 사건을 스냅숏, 행 판정, 그 밖의 사건으로 나눕니다.
//
// 스냅숏과 행 판정은 입구 사건과 반환 사건 사이를 한 묶음으로 봅니다. 두 함수 모두 스스로를 다시 부르지 않으므로
// 입구 뒤의 첫 반환이 그 호출의 반환입니다. 행 판정 묶음 안의 XidInMVCCSnapshot·TransactionIdDidCommit·SetHintBits는
// 그 판정이 거친 길이고, 각각을 부른 줄(함수 기록의 부른 자리)을 함께 남깁니다.
func decodeRequest(req *txRequest, evs []*uevent, win [2]uint64, oids, ssiRels map[int]string,
	fname func(*uevent) string, siteOf func(*uevent) string, stackAt func(*uevent) []frameRef, rel func(uint64) float64, syms *symbols) {
	var snapOpen []*uevent
	var check *txCheck
	var lockOpen []*uevent
	var sleepAt *uevent
	updOpen := -1
	// 행 잠금의 입구마다 사건 자리(관찰하지 않는 표면 -1)를 쌓아, 반환값을 같은 호출의 사건에 붙입니다.
	var tlockOpen []int
	mxOpen := -1
	fpiOpen := -1
	var dlResults []string
	for _, e := range evs {
		if e.ns < win[0] || e.ns > win[1] {
			continue
		}
		fn := fname(e)
		switch {
		case fn == "GetSnapshotData" && !e.Return:
			snapOpen = append(snapOpen, e)
		case fn == "GetSnapshotData" && e.Return && len(snapOpen) > 0:
			in := snapOpen[len(snapOpen)-1]
			snapOpen = snapOpen[:len(snapOpen)-1]
			dst, _ := strconv.ParseUint(strings.TrimPrefix(in.F["dst"], "0x"), 16, 64)
			kind := staticName(syms, dst-in.trace.Slide)
			if kind == "CatalogSnapshotData" {
				req.Catalog++
				continue
			}
			sn := txSnapshot{T: rel(e.ns), Kind: kind, Xmin: atoi(e.F["xmin"]), Xmax: atoi(e.F["xmax"]), Curcid: atoi(e.F["curcid"]), Stack: stackAt(in)}
			for k := 0; k < int(atoi(e.F["xcnt"])) && k < 4; k++ {
				sn.Xip = append(sn.Xip, atoi(e.F[fmt.Sprintf("xip%d", k)]))
			}
			req.Snapshots = append(req.Snapshots, sn)
		case fn == "HeapTupleSatisfiesMVCC" && !e.Return:
			check = nil
			name, ok := oids[int(atoi(e.F["rel"]))]
			if !ok {
				continue
			}
			blk := atoi(e.F["bhi"])<<16 | atoi(e.F["blo"])
			check = &txCheck{T: rel(e.ns), Table: name, Ctid: fmt.Sprintf("(%d,%d)", blk, atoi(e.F["off"])), Xmin: atoi(e.F["xmin"]), Xmax: atoi(e.F["xmax"]),
				Mask: atoi(e.F["mask"]), SnapMin: atoi(e.F["sxmin"]), SnapMax: atoi(e.F["sxmax"]), Tup: e.F["tup"], Lines: directCalls(e.trace, syms, e.rec), Stack: stackAt(e)}
		case fn == "HeapTupleSatisfiesMVCC" && e.Return:
			if check != nil {
				check.Visible = e.F["ret"] == "1"
				req.Checks = append(req.Checks, *check)
			}
			check = nil
		case check != nil && (fn == "XidInMVCCSnapshot" || fn == "TransactionIdDidCommit" || fn == "SetHintBits"):
			if !e.Return {
				d := txDecision{Fn: fn, At: siteOf(e), Arg: e.F["xid"]}
				if fn == "SetHintBits" {
					d.Arg = hintName(atoi(e.F["bits"]))
				} else {
					d.Inner = directCalls(e.trace, syms, e.rec)
				}
				check.Path = append(check.Path, d)
			} else if n := len(check.Path); n > 0 && check.Path[n-1].Fn == fn {
				check.Path[n-1].Ret = e.F["ret"]
			}
		case fn == "LockAcquireExtended" && !e.Return:
			lockOpen = append(lockOpen, e)
		case fn == "LockAcquireExtended" && e.Return && len(lockOpen) > 0:
			in := lockOpen[len(lockOpen)-1]
			lockOpen = lockOpen[:len(lockOpen)-1]
			typ := enumName(lockType, in.F["type"])
			// 관찰할 표의 표·행 락과 트랜잭션 락만 남깁니다. 카탈로그를 여느라 잡는 락은 요청마다 수십 개라 판단을 가립니다.
			_, watched := oids[int(atoi(in.F["f2"]))]
			if !(typ == "transactionid" || ((typ == "relation" || typ == "tuple") && watched)) {
				continue
			}
			f := map[string]string{"type": typ, "mode": enumName(lockMode, in.F["mode"]), "result": enumName(acquired, e.F["ret"])}
			switch typ {
			case "transactionid":
				f["xid"] = in.F["f1"]
			case "relation":
				f["rel"] = oids[int(atoi(in.F["f2"]))]
			case "tuple":
				f["rel"], f["ctid"] = oids[int(atoi(in.F["f2"]))], fmt.Sprintf("(%s,%s)", in.F["f3"], in.F["f4"])
			}
			req.Events = append(req.Events, txEvent{T: rel(in.ns), T1: rel(e.ns), Kind: "lock", F: f, Stack: stackAt(in)})
		// AssignTransactionId: 번호를 받는 트랜잭션의 깊이(1이 맨 바깥). 하위 트랜잭션은 부모에 번호가 없으면 부모가 먼저 받으므로
		// (xact.c 660~678행) 바로 뒤의 GetNewTransactionId가 어느 깊이의 번호인지를 이것으로 가립니다. 실험 5부터 있습니다.
		case fn == "AssignTransactionId":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "assign_xid", F: map[string]string{"level": e.F["level"]}})
		case fn == "GetNewTransactionId" && e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "xid", F: map[string]string{"xid": e.F["xid"]}})
		case fn == "TransactionIdCommitTree":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "commit_clog", F: map[string]string{"xid": e.F["xid"], "nsub": e.F["nsub"]}, Stack: stackAt(e)})
		case fn == "TransactionIdAbortTree":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "abort_clog", F: map[string]string{"xid": e.F["xid"], "nsub": e.F["nsub"]}, Stack: stackAt(e)})
		case fn == "HeapTupleSatisfiesUpdate" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "update_check", F: map[string]string{"off": e.F["off"], "xmin": e.F["xmin"], "xmax": e.F["xmax"]}, Stack: stackAt(e)})
			updOpen = len(req.Events) - 1
		case fn == "HeapTupleSatisfiesUpdate" && e.Return:
			if updOpen >= 0 {
				req.Events[updOpen].F["result"] = enumName(tmResult, e.F["ret"])
			}
			updOpen = -1
		case fn == "XactLockTableWait":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "xact_wait", F: map[string]string{"xid": e.F["xid"], "oper": enumName(xltwOper, e.F["oper"])}, Stack: stackAt(e)})
		case fn == "ProcSleep" && !e.Return:
			sleepAt = e
		case fn == "ProcSleep" && e.Return && sleepAt != nil:
			req.Events = append(req.Events, txEvent{T: rel(sleepAt.ns), T1: rel(e.ns), Kind: "proc_sleep", F: map[string]string{"ret": e.F["ret"]}, Stack: stackAt(sleepAt)})
			sleepAt = nil
		case fn == "ProcWakeup":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "wake", F: map[string]string{"who": e.F["who"], "status": e.F["status"]}, Stack: stackAt(e)})
		case fn == "SetHintBits" && !e.Return:
			// 판정(HeapTupleSatisfiesMVCC) 안의 쓰기는 위에서 판정 경로에 들어가므로, 여기 오는 것은 그 밖의 쓰기입니다(기다린 뒤
			// heap_update의 UpdateXmaxHintBits, 고유 키 검사의 HeapTupleSatisfiesDirty 등). 어느 표의 행인지는 모든 판정의 행 머리
			// 주소를 모은 뒤 labelHints가 정하고, 관찰하지 않는 표의 것은 버립니다.
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "hint", F: map[string]string{"tup": e.F["tup"], "bit": hintName(atoi(e.F["bits"])), "xid": e.F["xid"], "at": siteOf(e)}, Stack: stackAt(e)})
		case fn == "EvalPlanQual":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "epq", Stack: stackAt(e)})
		case fn == "heap_lock_tuple" && !e.Return:
			name, ok := oids[int(atoi(e.F["rel"]))]
			if !ok {
				tlockOpen = append(tlockOpen, -1)
				continue
			}
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "tuple_lock", F: map[string]string{"table": name, "off": e.F["off"], "mode": enumName(tupleMode, e.F["mode"]), "wait": e.F["wait"]}, Stack: stackAt(e)})
			tlockOpen = append(tlockOpen, len(req.Events)-1)
		case fn == "heap_lock_tuple" && e.Return && len(tlockOpen) > 0:
			if k := tlockOpen[len(tlockOpen)-1]; k >= 0 {
				req.Events[k].F["result"] = enumName(tmResult, e.F["ret"])
				req.Events[k].T1 = rel(e.ns)
			}
			tlockOpen = tlockOpen[:len(tlockOpen)-1]
		case fn == "MultiXactIdCreate" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "multixact", F: map[string]string{"xid1": e.F["xid1"], "st1": enumName(mxStatus, e.F["st1"]), "xid2": e.F["xid2"], "st2": enumName(mxStatus, e.F["st2"])}, Stack: stackAt(e)})
			mxOpen = len(req.Events) - 1
		case fn == "MultiXactIdCreate" && e.Return && mxOpen >= 0:
			req.Events[mxOpen].F["id"] = e.F["ret"]
			mxOpen = -1
		case fn == "DeadLockCheck" && e.Return:
			dlResults = append(dlResults, enumName(dlState, e.F["ret"]))
		// WAL과 지속성(실험 3). 커밋은 XLogFlush로 커밋 레코드 끝까지 디스크에 내린 뒤 커밋 로그를 적고, 비동기 커밋과 취소는
		// XLogSetAsyncXactLSN으로 WAL writer에게 맡기기만 합니다(xact.c 1502·1523·1841행).
		case fn == "XLogFlush" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "wal_flush", F: map[string]string{"lsn": lsnText(e.F["lsn"])}, Stack: stackAt(e)})
		case fn == "XLogSetAsyncXactLSN" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "async_lsn", F: map[string]string{"lsn": lsnText(e.F["lsn"])}, Stack: stackAt(e)})
		case fn == "TransactionIdAsyncCommitTree":
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "commit_clog", F: map[string]string{"xid": e.F["xid"], "lsn": lsnText(e.F["lsn"]), "async": "1"}, Stack: stackAt(e)})
		// 체크섬이 켜져 있으면 깨끗한 페이지에 힌트 비트만 쓸 때 페이지 이미지를 WAL에 남길 수 있습니다. 반환 LSN이 0이면 남기지 않았습니다.
		case fn == "XLogSaveBufferForHint" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "hint_fpi", F: map[string]string{"buf": e.F["buf"]}, Stack: stackAt(e)})
			fpiOpen = len(req.Events) - 1
		case fn == "XLogSaveBufferForHint" && e.Return && fpiOpen >= 0:
			req.Events[fpiOpen].F["lsn"], req.Events[fpiOpen].T1 = lsnText(e.F["lsn"]), rel(e.ns)
			fpiOpen = -1
		// SERIALIZABLE(실험 4). 술어 락 요청과 쓰기의 대상 검사는 그 호출 안에서 무엇을 불렀는지가 결과입니다. 요청은
		// CreatePredicateLock을 불렀으면 락을 새로 만든 것이고(이미 같거나 더 굵은 락이 있으면 부르지 않음, predicate.c 2524·2527행),
		// 대상 검사는 FlagRWConflict를 불렀으면 그 대상에서 간선을 그은 것입니다(4244행). 관찰하지 않는 표(카탈로그 등)의 것은 버립니다.
		case fn == "GetSerializableTransactionSnapshotInt" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "ssi_begin", Stack: stackAt(e)})
		case (fn == "PredicateLockRelation" || fn == "PredicateLockPage" || fn == "PredicateLockTID") && !e.Return:
			name, ok := ssiRels[int(atoi(e.F["rel"]))]
			if !ok {
				continue
			}
			f := map[string]string{"rel": name, "gran": map[string]string{"PredicateLockRelation": "relation", "PredicateLockPage": "page", "PredicateLockTID": "tuple"}[fn]}
			switch fn {
			case "PredicateLockPage":
				f["blk"] = e.F["blk"]
			case "PredicateLockTID":
				f["ctid"], f["txid"] = fmt.Sprintf("(%d,%d)", atoi(e.F["bhi"])<<16|atoi(e.F["blo"]), atoi(e.F["off"])), e.F["txid"]
			}
			f["created"] = boolText(callsWithin(e.trace, e.rec, win[1], func(fn uint64) bool { return syms.name(fn-e.trace.Slide) == "CreatePredicateLock" }))
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "pred_lock", F: f, Stack: stackAt(e)})
		case fn == "CheckForSerializableConflictIn" && !e.Return:
			if name, ok := ssiRels[int(atoi(e.F["rel"]))]; ok {
				req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "conflict_in", F: map[string]string{"rel": name, "blk": e.F["blk"]}, Stack: stackAt(e)})
			}
		case fn == "CheckTargetForConflictsIn" && !e.Return:
			name, ok := ssiRels[int(atoi(e.F["rel"]))]
			if !ok {
				continue
			}
			f := predTarget(e.F["blk"], e.F["off"])
			f["rel"] = name
			f["flagged"] = boolText(callsWithin(e.trace, e.rec, win[1], func(fn uint64) bool { return syms.name(fn-e.trace.Slide) == "FlagRWConflict" }))
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "conflict_target", F: f, Stack: stackAt(e)})
		case fn == "CheckForSerializableConflictOut" && !e.Return:
			if name, ok := ssiRels[int(atoi(e.F["rel"]))]; ok {
				req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "conflict_out", F: map[string]string{"rel": name, "xid": e.F["xid"]}, Stack: stackAt(e)})
			}
		case fn == "FlagRWConflict" && !e.Return:
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "rw_conflict", F: map[string]string{
				"rpid": e.F["rpid"], "rxid": e.F["rxid"], "rflags": sxactFlagNames(e.F["rflags"]),
				"wpid": e.F["wpid"], "wxid": e.F["wxid"], "wflags": sxactFlagNames(e.F["wflags"])}, Stack: stackAt(e)})
		case fn == "PreCommit_CheckForSerializationFailure" && !e.Return:
			// RR·RC에는 SERIALIZABLEXACT가 없어 포인터(sx)가 0이고, 그때 읽은 flags는 뜻이 없습니다(읽기에 실패한 숫자 인자는 0으로 남음).
			f := map[string]string{"sxact": "none"}
			if atoi(e.F["sx"]) != 0 {
				f = map[string]string{"flags": sxactFlagNames(e.F["flags"])}
			}
			req.Events = append(req.Events, txEvent{T: rel(e.ns), Kind: "ssi_precommit", F: f, Stack: stackAt(e)})
		}
	}
	// uprobe를 걸지 않은 함수도 함수 기록에는 모두 있습니다. 그중 판단에 쓰는 셋을 찾습니다.
	// errfinish: ereport가 실제로 보고될 때만 불리므로 그것을 부른 줄이 오류를 낸 자리입니다.
	// handle_sig_alarm: SIGALRM 처리기. 락을 기다리기 시작할 때 건 deadlock_timeout 타이머가 울린 것입니다.
	// CheckDeadLock: 그 타이머가 기다림 중에 울렸을 때 교착을 검사합니다(기다림이 이미 끝났으면 불리지 않음).
	if len(evs) > 0 {
		t := evs[0].trace
		sample := stackSampler(t, syms)
		var dlOpen uint64
		for _, r := range t.Records {
			if r.Ns < win[0] || r.Ns > win[1] {
				continue
			}
			switch name := syms.name(r.Fn - t.Slide); {
			case r.Kind == kindEnter && name == "errfinish":
				s, _ := sample(r.Ns, true)
				req.Events = append(req.Events, txEvent{T: rel(r.Ns), Kind: "ereport", F: map[string]string{"at": syms.lineAt(r.Site - t.Slide - 4)}, Stack: trimStack(s)})
			case r.Kind == kindEnter && name == "handle_sig_alarm":
				req.Events = append(req.Events, txEvent{T: rel(r.Ns), Kind: "alarm"})
			// ClearOldPredicateLocks: 끝난 직렬화 트랜잭션의 SIRead 락을 지웁니다. 겹친 직렬화 트랜잭션 가운데 가장 오래된 xmin을
			// 가진 마지막 것이 끝날 때 불립니다(predicate.c 3636~3673행, 실험 4).
			case r.Kind == kindEnter && name == "ClearOldPredicateLocks":
				s, _ := sample(r.Ns, true)
				req.Events = append(req.Events, txEvent{T: rel(r.Ns), Kind: "ssi_clear", Stack: trimStack(s)})
			// _bt_doinsert: B-tree에 항목을 넣습니다. UPDATE가 HOT이면 불리지 않습니다(실험 4 W12).
			case r.Kind == kindEnter && name == "_bt_doinsert":
				s, _ := sample(r.Ns, true)
				req.Events = append(req.Events, txEvent{T: rel(r.Ns), Kind: "index_insert", Stack: trimStack(s)})
			case r.Kind == kindEnter && name == "CheckDeadLock":
				dlOpen = r.Ns
			case r.Kind == kindExit && name == "CheckDeadLock" && dlOpen != 0:
				// 같은 요청 안의 DeadLockCheck 반환값을 순서대로 붙입니다(CheckDeadLock 한 번이 DeadLockCheck를 한 번 부름, proc.c 1857행).
				// 실험 1의 기록에는 그 uprobe가 없어 결과가 비어 있습니다.
				ev := txEvent{T: rel(dlOpen), T1: rel(r.Ns), Kind: "deadlock_check", F: map[string]string{}}
				if len(dlResults) > 0 {
					ev.F["result"], dlResults = dlResults[0], dlResults[1:]
				}
				req.Events = append(req.Events, ev)
				dlOpen = 0
			}
		}
	}
}

// callsWithin 은 함수 기록에서 입구 기록 i의 호출이 끝나기 전에(깊이와 관계없이) want가 참인 함수를 불렀는지 돌려줍니다.
// 호출이 longjmp로 끝나 출구 기록이 빠지면 끝을 모르므로, 요청 구간의 끝(limit, 나노초)에서 멈춥니다.
func callsWithin(t *traceFile, i int, limit uint64, want func(fn uint64) bool) bool {
	depth := 0
	for k := i + 1; k < len(t.Records); k++ {
		r := t.Records[k]
		if r.Ns > limit {
			return false
		}
		switch r.Kind {
		case kindEnter:
			if want(r.Fn) {
				return true
			}
			depth++
		case kindExit:
			if depth == 0 {
				return false
			}
			depth--
		}
	}
	return false
}

// predTarget 은 술어 락 대상 태그의 블록·줄 필드로 굵기를 정합니다. 표 단위는 블록이 InvalidBlockNumber(0xFFFFFFFF),
// 페이지 단위는 줄이 InvalidOffsetNumber(0)입니다(predicate_internals.h 388~404행).
func predTarget(blk, off string) map[string]string {
	b, o := atoi(blk), atoi(off)
	switch {
	case b == 0xFFFFFFFF:
		return map[string]string{"gran": "relation"}
	case o == 0:
		return map[string]string{"gran": "page", "blk": strconv.FormatInt(b, 10)}
	default:
		return map[string]string{"gran": "tuple", "ctid": fmt.Sprintf("(%d,%d)", b, o)}
	}
}

// sxactFlagNames 는 SERIALIZABLEXACT의 flags를 켜진 이름들로 바꿉니다(predicate_internals.h 121~142행). 0이면 빈 글입니다.
// 숫자가 아니면(읽지 못한 값) 그대로 돌려줍니다.
func sxactFlagNames(v string) string {
	n, err := strconv.ParseUint(v, 0, 32)
	if err != nil {
		return v
	}
	names := []string{"COMMITTED", "PREPARED", "ROLLED_BACK", "DOOMED", "CONFLICT_OUT", "READ_ONLY", "DEFERRABLE_WAITING",
		"RO_SAFE", "RO_UNSAFE", "SUMMARY_CONFLICT_IN", "SUMMARY_CONFLICT_OUT", "PARTIALLY_RELEASED"}
	var on []string
	for i, name := range names {
		if n&(1<<i) != 0 {
			on = append(on, name)
		}
	}
	return strings.Join(on, "|")
}

// boolText 는 사건 필드에 쓰는 참·거짓 글("1", "0")입니다.
func boolText(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// directCalls 는 함수 기록에서 입구 기록 i의 호출이 직접 부른 함수와 그 부른 줄을 부른 순서대로 돌려줍니다.
// 호출 안에서 신호 처리기가 돌면 그것도 직접 부른 것처럼 나옵니다(실제로 그 사이에 돈 것이므로 지우지 않음).
// 호출 도중 longjmp로 출구 기록이 빠지면 깊이 계산이 어긋나므로, 오류로 끝난 호출의 결과는 믿지 않습니다.
func directCalls(t *traceFile, syms *symbols, i int) []frameRef {
	out := []frameRef{}
	depth := 0
	for k := i + 1; k < len(t.Records); k++ {
		r := t.Records[k]
		switch r.Kind {
		case kindEnter:
			if depth == 0 {
				out = append(out, frameRef{Fn: syms.name(r.Fn - t.Slide), At: syms.lineAt(r.Site - t.Slide - 4)})
			}
			depth++
		case kindExit:
			if depth == 0 {
				return out
			}
			depth--
		}
	}
	return out
}

// staticName 은 정적 변수의 ELF 주소에서 이름을 찾습니다. 스냅숏을 담을 자리가 문장용(CurrentSnapshotData)인지 카탈로그용인지 가릅니다.
// 실행 파일의 기호표에서 찾고, 없으면 주소를 16진수로 돌려줍니다.
func staticName(syms *symbols, addr uint64) string {
	if n, ok := syms.vars[addr]; ok {
		return n
	}
	return fmt.Sprintf("0x%x", addr)
}

// hintName 은 SetHintBits의 infomask 인자를 이름으로 바꿉니다(access/htup_details.h).
func hintName(bits int64) string {
	names := map[int64]string{0x0100: "HEAP_XMIN_COMMITTED", 0x0200: "HEAP_XMIN_INVALID", 0x0400: "HEAP_XMAX_COMMITTED", 0x0800: "HEAP_XMAX_INVALID"}
	if n, ok := names[bits]; ok {
		return n
	}
	return fmt.Sprintf("0x%04x", bits)
}

// attrValue 는 pageinspect가 돌려준 열 바이트(\x..)를 열 형식으로 풉니다. integer는 4바이트 리틀 엔디언,
// text는 가변 길이 머리(1바이트 머리면 길이 = 첫 바이트 >> 1, 4바이트 머리면 길이 = 첫 4바이트 >> 2)를 벗긴 문자열,
// boolean은 1바이트(0이면 거짓)를 psql처럼 t·f로 씁니다(실험 4의 on_call). 다른 형식은 16진수 그대로 둡니다.
func attrValue(a, typ string) string {
	if a == "NULL" {
		return "NULL"
	}
	b, err := hex.DecodeString(strings.TrimPrefix(a, `\x`))
	if err != nil {
		return a
	}
	switch typ {
	case "integer":
		if len(b) == 4 {
			return fmt.Sprint(int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24))
		}
	case "boolean":
		if len(b) == 1 {
			if b[0] != 0 {
				return "t"
			}
			return "f"
		}
	case "text":
		if len(b) > 0 && b[0]&1 == 1 {
			n := int(b[0] >> 1)
			if n <= len(b) {
				return string(b[1:n])
			}
		} else if len(b) >= 4 {
			n := int((uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24) >> 2)
			if n <= len(b) {
				return string(b[4:n])
			}
		}
	}
	return a
}

func atoi(s string) int64 {
	n, _ := strconv.ParseInt(s, 0, 64)
	return n
}
