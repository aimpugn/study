package main

// 이 파일은 시나리오 해석 결과가 관찰과 맞는지 스스로 견주는 값을 만듭니다. 탐색용 코드입니다(2026-10-14, 트랜잭션 관측 실험 1).
// 화면은 이 값으로 "이 그림의 판정은 기록과 맞다"를 보여 주고, 맞지 않으면 그 단계의 판정을 기록으로 쓰지 않습니다.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// stepCheck 는 결과 행에 ctid가 있는 단계에서, 줄마다의 판정 수와 보인 줄을 페이지·결과와 견준 값입니다.
type stepCheck struct {
	Table      string   `json:"table"`
	LpNormal   int      `json:"lpNormal"`
	Checks     int      `json:"checks"`
	Visible    []string `json:"visible"`
	Result     []string `json:"result"`
	CountMatch bool     `json:"countMatch"`
	SetMatch   bool     `json:"setMatch"`
}

// checkStep 은 단계 하나의 판정을 견줍니다.
//
// Arguments: before는 이 단계 직전 관찰의 페이지(표 이름 → 줄), out은 psql 출력 줄입니다.
// Returns: 출력에 ctid 열이 없거나 직전 관찰이 없으면 nil입니다. LpNormal은 직전 관찰에서 정상 줄(lp_flags 1)의 수,
// Checks는 그 표에 대한 HeapTupleSatisfiesMVCC 호출 수입니다. 페이지 전체가 보임 표시(all_visible)인 페이지는
// 줄마다의 판정을 건너뛰므로 그때는 두 수가 달라도 오류가 아닙니다(이 실험의 표는 VACUUM 전이라 그 표시가 없음).
func checkStep(table string, before []txVersion, reqs []txRequest, out []string) *stepCheck {
	result, ok := resultCtids(out)
	if !ok || before == nil {
		return nil
	}
	c := &stepCheck{Table: table, Visible: []string{}, Result: result}
	for _, v := range before {
		if v.LpFlag == 1 {
			c.LpNormal++
		}
	}
	for _, r := range reqs {
		for _, ch := range r.Checks {
			if ch.Table != table {
				continue
			}
			c.Checks++
			if ch.Visible {
				c.Visible = append(c.Visible, ch.Ctid)
			}
		}
	}
	c.CountMatch = c.Checks == c.LpNormal
	c.SetMatch = sameSet(c.Visible, c.Result)
	return c
}

// rowCount 는 psql 표 출력의 끝 줄입니다.
var rowCount = regexp.MustCompile(`^\(\d+ rows?\)$`)

// resultCtids 는 psql의 표 출력에서 ctid 열을 읽습니다. 머리 줄에 ctid가 없으면 ok가 거짓입니다.
func resultCtids(out []string) (ctids []string, ok bool) {
	col := -1
	for _, l := range out {
		cells := strings.Split(l, "|")
		if col < 0 {
			for k, c := range cells {
				if strings.TrimSpace(c) == "ctid" {
					col = k
				}
			}
			continue
		}
		// 머리 줄 바로 다음은 구분선(----+----)이고, 결과 행은 "(1 row)"·"(N rows)" 앞까지입니다. ctid가 첫 열이면 결과 행도
		// "("로 시작하므로 끝 표시는 행 수만 든 줄로 가립니다.
		if rowCount.MatchString(strings.TrimSpace(l)) {
			break
		}
		if strings.HasPrefix(strings.TrimSpace(l), "-") {
			continue
		}
		if col < len(cells) {
			ctids = append(ctids, strings.TrimSpace(cells[col]))
		}
	}
	if col < 0 {
		return nil, false
	}
	if ctids == nil {
		ctids = []string{}
	}
	return ctids, true
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]int{}
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

// bystander 는 시나리오 세션이 아닌 백엔드(관찰용 접속 등)가 관찰 대상 표를 건드렸는지입니다.
// 관찰용 접속의 함수 기록은 용량 때문에 지우므로(txlab pruneTraces) uprobe 사건만으로 셉니다.
type bystander struct {
	Pid           int    `json:"pid"`
	Queries       int    `json:"queries"`
	FirstSQL      string `json:"firstSql"`
	WatchedChecks int    `json:"watchedChecks"`
	WatchedHints  int    `json:"watchedHints"`
	Hints         int    `json:"hints"`
	OutsideHints  int    `json:"outsideHints"`
}

// bystanders 는 세션이 아닌 pid마다, 관찰 대상 표의 행에 대한 HeapTupleSatisfiesMVCC 호출 수와 그 호출 안의 SetHintBits 수를 셉니다.
//
// SetHintBits는 관계 번호를 인자로 받지 않으므로, 바로 바깥의 HeapTupleSatisfiesMVCC 호출(입구와 반환 사이)이 어느 표의 행인지로 가립니다.
// 그 바깥에서 불린 SetHintBits(다른 판정 함수 안)는 어느 표의 행인지 모르므로 OutsideHints로 따로 셉니다. 이 수가 0이 아니면
// "관찰 대상 표의 힌트 비트를 쓰지 않았다"를 이 값만으로 말할 수 없습니다. 사건은 trace 파일의 순서(시각 순서)로 읽습니다.
func bystanders(evs []uevent, sessions map[int]bool, oids map[int]string) []bystander {
	by := map[int]*bystander{}
	inWatched, inMvcc := map[int]bool{}, map[int]bool{}
	var order []int
	for _, e := range evs {
		if sessions[e.Pid] || e.Pid == 0 {
			continue
		}
		b := by[e.Pid]
		if b == nil {
			b = &bystander{Pid: e.Pid}
			by[e.Pid] = b
			order = append(order, e.Pid)
		}
		switch e.Name {
		case "query":
			b.Queries++
			if b.FirstSQL == "" {
				b.FirstSQL = e.F["sql"]
			}
		case "mvcc":
			_, w := oids[int(atoi(e.F["rel"]))]
			inWatched[e.Pid], inMvcc[e.Pid] = w, true
			if w {
				b.WatchedChecks++
			}
		case "mvccret":
			inWatched[e.Pid], inMvcc[e.Pid] = false, false
		case "hint":
			b.Hints++
			if inWatched[e.Pid] {
				b.WatchedHints++
			}
			if !inMvcc[e.Pid] {
				b.OutsideHints++
			}
		}
	}
	out := []bystander{}
	for _, p := range order {
		out = append(out, *by[p])
	}
	return out
}

// hintChange 는 관찰 두 번 사이에 행 버전 하나에 새로 켜진 힌트 비트입니다. By·At·T는 그 비트를 쓴 세션의 SetHintBits
// (세션, 부른 줄, 시각)이고, By가 비어 있으면 세션의 쓰기로 설명되지 않는 비트입니다(관찰용 접속이나 다른 백엔드가 썼을 수 있음).
type hintChange struct {
	Step  int     `json:"step"`
	Table string  `json:"table"`
	Ctid  string  `json:"ctid"`
	Bit   string  `json:"bit"`
	By    string  `json:"by"`
	At    string  `json:"at"`
	T     float64 `json:"t"`
	Via   string  `json:"via,omitempty"`
}

// hintWrite 는 세션 백엔드의 SetHintBits 호출 하나입니다. Xmin·Xmax는 그때 그 행 머리의 값입니다.
type hintWrite struct {
	T          float64
	Session    string
	Xmin, Xmax int64
	Bit, At    string
	Tup        string
}

// hintBits 는 pageinspect가 돌려주는 힌트 비트 이름입니다(access/htup_details.h).
var hintBits = []string{"HEAP_XMIN_COMMITTED", "HEAP_XMIN_INVALID", "HEAP_XMAX_COMMITTED", "HEAP_XMAX_INVALID"}

// hintChanges 는 관찰 두 번(단계 n 직전과 직후) 사이에, 이미 있던 줄에 새로 켜진 힌트 비트마다 그것을 쓴 세션을 찾습니다.
//
// SetHintBits는 표 번호를 인자로 받지 않아 uprobe만으로는 관찰용 접속이 관찰 대상 표의 비트를 썼는지 가를 수 없습니다.
// 그래서 페이지를 직접 봅니다. 두 관찰 시각 사이에 세션이 같은 값의 행(xmin이 같고, xmax 비트면 xmax도 같음)에 같은 비트를 쓴
// SetHintBits가 있으면 그 세션으로 설명합니다. 세션이 비트를 쓰는 길은 줄 판정만이 아닙니다(2026-10-14 실험 1-rc-write에서
// 깨어난 S1이 heap_update와 heap_lock_tuple 쪽에서 씀). 단계가 아니라 시각으로 견주는 이유도 같습니다. 기다리던 요청은
// 다른 세션의 단계 중에 깨어나 씁니다. 그 사이에 새로 생긴 줄은 만든 쪽이 처음 값을 정하므로 견주지 않습니다.
// 한계: 값으로 맞추므로 xmin·xmax가 같은 다른 표의 행(같은 트랜잭션이 만든 카탈로그 행 등)에 쓴 세션 기록도 설명으로 잡힐 수 있습니다.
//
// 실험 2부터는 줄 판정과 SetHintBits가 행 머리의 주소(tup)를 함께 남깁니다. addrs에 그 줄의 주소가 있으면 값 대신 주소로 맞추고
// Via를 "addr"로, 주소를 모르면 위의 값 맞추기를 하고 Via를 "value"로 적습니다. 주소는 같은 값의 다른 행(카탈로그 행 등)과
// 헷갈리지 않습니다.
//
// kill이 있으면 같은 단계 번호 뒤의 관찰이 둘입니다(단계 뒤, 재시작 뒤). 단계 n의 직전은 n-1 뒤의 마지막 관찰(재시작 뒤가 있으면
// 그것), 직후는 n 뒤의 첫 관찰(재시작 앞)입니다.
func hintChanges(steps []txStep, caps []txCapture, table string, writes []hintWrite, addrs map[string]map[string]bool) []hintChange {
	out := []hintChange{}
	lastAfter, firstAfter := map[int]txCapture{}, map[int]txCapture{}
	for _, c := range caps {
		lastAfter[c.AfterStep] = c
		if _, ok := firstAfter[c.AfterStep]; !ok {
			firstAfter[c.AfterStep] = c
		}
	}
	for _, s := range steps {
		before, ok1 := lastAfter[s.N-1]
		after, ok2 := firstAfter[s.N]
		if !ok1 || !ok2 {
			continue
		}
		old := map[string]txVersion{}
		for _, v := range before.Pages[table] {
			old[fmt.Sprintf("(%d,%d)", v.Block, v.Lp)] = v
		}
		for _, v := range after.Pages[table] {
			ctid := fmt.Sprintf("(%d,%d)", v.Block, v.Lp)
			o, ok := old[ctid]
			if !ok {
				continue
			}
			for _, bit := range hintBits {
				if !hasFlag(v.Flags, bit) || hasFlag(o.Flags, bit) {
					continue
				}
				hc := hintChange{Step: s.N, Table: table, Ctid: ctid, Bit: bit}
				known := len(addrs[ctid]) > 0
				for _, w := range writes {
					if w.Bit != bit || w.T <= before.T || w.T > after.T {
						continue
					}
					var same bool
					if known && w.Tup != "" {
						same, hc.Via = addrs[ctid][w.Tup], "addr"
					} else {
						xmaxOK := !strings.Contains(bit, "XMAX") || fmt.Sprint(w.Xmax) == v.Xmax
						same, hc.Via = fmt.Sprint(w.Xmin) == v.Xmin && xmaxOK, "value"
					}
					if same {
						hc.By, hc.At, hc.T = w.Session, w.At, w.T
						break
					}
				}
				if hc.By == "" {
					hc.Via = ""
				}
				out = append(out, hc)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Step != out[j].Step {
			return out[i].Step < out[j].Step
		}
		return out[i].Ctid+out[i].Bit < out[j].Ctid+out[j].Bit
	})
	return out
}

func hasFlag(flags []string, f string) bool {
	for _, x := range flags {
		if x == f {
			return true
		}
	}
	return false
}

// busiestTable 은 단계의 요청들에서 줄 판정(HeapTupleSatisfiesMVCC)을 가장 많이 받은 표의 이름을 돌려줍니다.
// 판정이 하나도 없거나 같으면 tables의 앞쪽 표입니다.
func busiestTable(tables []tableMeta, reqs []txRequest) string {
	if len(tables) == 0 {
		return ""
	}
	n := map[string]int{}
	for _, r := range reqs {
		for _, c := range r.Checks {
			n[c.Table]++
		}
	}
	best := tables[0].Name
	for _, t := range tables[1:] {
		if n[t.Name] > n[best] {
			best = t.Name
		}
	}
	return best
}

// tupleAddrs 는 줄 판정이 남긴 행 머리 주소를 표 → ctid → 주소 집합으로 모읍니다. 실험 2의 L13은 이 집합이 줄마다 하나인지
// (모든 세션이 같은 공유 버퍼 주소를 보는지) 봅니다.
func tupleAddrs(steps []txStep) map[string]map[string]map[string]bool {
	out := map[string]map[string]map[string]bool{}
	for _, s := range steps {
		for _, r := range s.Requests {
			for _, c := range r.Checks {
				if c.Tup == "" {
					continue
				}
				if out[c.Table] == nil {
					out[c.Table] = map[string]map[string]bool{}
				}
				if out[c.Table][c.Ctid] == nil {
					out[c.Table][c.Ctid] = map[string]bool{}
				}
				out[c.Table][c.Ctid][c.Tup] = true
			}
		}
	}
	return out
}

// labelHints 는 요청마다 남긴 SetHintBits 사건(kind "hint")에 행 머리 주소로 표와 ctid를 붙이고, 관찰하는 표의 행이 아닌
// 것은 지웁니다. SetHintBits는 표 번호를 받지 않아 주소 말고는 어느 행인지 알 길이 없고, 카탈로그 행에 쓰는 것이 대부분입니다.
//
// Arguments: steps는 해석한 단계(사건을 제자리에서 고침), addrs는 tupleAddrs의 결과(표 → ctid → 주소)입니다.
// 주소가 같은 행이 둘이면 첫 것을 씁니다. 한 주소에는 한 줄 머리만 있으므로 그런 일은 페이지를 다시 쓴(VACUUM 등) 뒤에만 생깁니다.
func labelHints(steps []txStep, addrs map[string]map[string]map[string]bool) {
	type row struct{ table, ctid string }
	at := map[string]row{}
	for t, byCtid := range addrs {
		for c, set := range byCtid {
			for a := range set {
				if _, ok := at[a]; !ok {
					at[a] = row{t, c}
				}
			}
		}
	}
	for i := range steps {
		for j := range steps[i].Requests {
			r := &steps[i].Requests[j]
			kept := r.Events[:0]
			for _, e := range r.Events {
				if e.Kind == "hint" {
					w, ok := at[e.F["tup"]]
					if !ok {
						continue
					}
					e.F["table"], e.F["ctid"] = w.table, w.ctid
				}
				kept = append(kept, e)
			}
			r.Events = kept
		}
	}
}

// addrRow 는 tupleAddrs를 JSON으로 내보낼 때의 한 줄입니다. Epoch는 kill로 나뉜 구간의 번호(0부터)입니다.
type addrRow struct {
	Epoch int      `json:"epoch,omitempty"`
	Table string   `json:"table"`
	Ctid  string   `json:"ctid"`
	Addrs []string `json:"addrs"`
}

// addrSummary 는 tupleAddrs를 표·ctid 순서의 목록으로 바꿉니다.
func addrSummary(m map[string]map[string]map[string]bool) []addrRow {
	out := []addrRow{}
	for t, byCtid := range m {
		for c, set := range byCtid {
			r := addrRow{Table: t, Ctid: c}
			for a := range set {
				r.Addrs = append(r.Addrs, a)
			}
			sort.Strings(r.Addrs)
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Table+out[i].Ctid < out[j].Table+out[j].Ctid })
	return out
}

// captureReads 는 관찰마다, 관찰용 접속이 페이지·세션 상태·락 등을 실제로 읽은 구간을 질의를 받은 시각으로 돌려줍니다.
//
// 관찰은 한 순간이 아닙니다. txlab은 표마다 페이지, pg_stat_activity, pg_locks, 스냅숏, 커밋 로그 상태를 차례로 묻고, 관찰용
// 접속도 추적을 받아 느립니다. 2026-10-14 실험 2의 관찰 104개(벌마다 끝을 모르는 마지막 관찰은 뺌)에서 페이지는 묻기 시작한 뒤
// 중앙값 104 ms·최대 441 ms 안에, 세션 상태와 락은 중앙값 367 ms·최대 935 ms 안에 다 읽었습니다. 그래서 묻기 시작한 시각
// 하나로는 그 관찰이 무엇보다 앞인지 가를 수 없습니다.
//
// Arguments: capNs는 관찰마다 txlab이 묻기 시작한 시각(나노초, 오름차순), obs는 관찰용 접속의 strace 사건(시각 순)입니다.
// Returns: 관찰마다 부분 이름("pages", "activity", "locks", "running", "xacts") → [시작, 끝] 나노초입니다. 부분은 질의 글로 가리고,
// 표가 여럿이면 pages는 첫 표의 시작부터 끝 표의 끝까지입니다. 한 질의의 구간은 그 질의를 받은 recvfrom부터 관찰용 접속이
// 다음 질의를 받은 때까지입니다(결과를 다 보낸 뒤라 위쪽 경계). 다음 질의가 없으면 끝은 다음 관찰의 시작, 마지막 관찰이면 0입니다.
func captureReads(capNs []uint64, obs []sysEvent) []map[string][2]uint64 {
	type q struct {
		ns   uint64
		part string
	}
	var qs []q
	for _, e := range obs {
		if e.Name != "recvfrom" || e.Ret <= 0 {
			continue
		}
		text := queryText(e.Data)
		if text == "" {
			continue
		}
		qs = append(qs, q{e.Ns, capturePart(text)})
	}
	out := make([]map[string][2]uint64, len(capNs))
	for k, c := range capNs {
		out[k] = map[string][2]uint64{}
		var next uint64
		if k+1 < len(capNs) {
			next = capNs[k+1]
		}
		for i, x := range qs {
			if x.ns < c || (next != 0 && x.ns >= next) || x.part == "" {
				continue
			}
			end := next
			if i+1 < len(qs) && (next == 0 || qs[i+1].ns < next) {
				end = qs[i+1].ns
			}
			w, ok := out[k][x.part]
			if !ok {
				w[0] = x.ns
			}
			w[1] = end
			out[k][x.part] = w
		}
	}
	return out
}

// capturePart 는 관찰용 접속이 받은 질의 글이 관찰의 어느 부분인지 돌려줍니다. 관찰이 아닌 질의(실험 5의 waiters 등)는 빈 문자열입니다.
func capturePart(text string) string {
	switch {
	case strings.Contains(text, "heap_page_item_attrs"):
		return "pages"
	case strings.Contains(text, "FROM pg_stat_activity") && strings.Contains(text, "backend_xid"):
		return "activity"
	case strings.Contains(text, "FROM pg_locks"):
		return "locks"
	case strings.Contains(text, "pg_current_snapshot"):
		return "running"
	case strings.Contains(text, "pg_xact_status"):
		return "xacts"
	}
	return ""
}

// epochSteps 는 단계를 kill로 나뉜 구간마다 묶습니다. cuts는 kill 뒤 첫 단계의 번호들(오름차순)이고, 단계 n은 n 이하인 cut의 수가
// 그 구간의 번호입니다. 묶음은 steps의 부분 슬라이스라서 묶음 안을 고치면 steps도 바뀝니다.
func epochSteps(steps []txStep, cuts []int) [][]txStep {
	var out [][]txStep
	start := 0
	for _, c := range cuts {
		end := start
		for end < len(steps) && steps[end].N < c {
			end++
		}
		out = append(out, steps[start:end])
		start = end
	}
	return append(out, steps[start:])
}
