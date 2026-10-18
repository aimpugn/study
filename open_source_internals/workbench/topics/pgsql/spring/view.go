package spring

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
)

// view 메서드는 원본의 view를 되살립니다. 장면 구간과 멈춤 구간은 멈춤 자리에서, 트리·묶임·슬롯·기다림은 사건에서, 기다림 고리는 그것들로
// 해석기처럼 다시 셉니다. 해석기의 요약과 발췌, 갈림은 run attrs입니다.
func (v *view) view(get func(src, name string) []atlas.Event, checkpoints, conns []any) (map[string]any, error) {
	out := map[string]any{}
	for src, keys := range map[string][]string{srcSpring: {"ambiguous", "poolConfig", "javaFrom", "siteMismatch"}, srcCompare: {"branches", "compared"}, srcJsrc: {"java", "jsrc"}} {
		for _, k := range keys {
			val, ok := v.r.Attrs[src]["view."+k]
			if !ok {
				return nil, fmt.Errorf("run attrs %s.view.%s 없음", src, k)
			}
			out[k] = val
		}
	}

	// 장면 구간은 [start의 tGo, end의 t], 멈춤 구간은 go가 알림보다 늦은 멈춤 자리마다 [t, tGo]입니다(decode/spring.go 711~724).
	var from, to any
	pauses := []any{}
	for _, raw := range checkpoints {
		c := raw.(map[string]any)
		switch c["label"] {
		case "start":
			from = c["tGo"]
		case "end":
			to = c["t"]
		}
		if c["tGo"].(float64) > c["t"].(float64) {
			pauses = append(pauses, []any{c["t"], c["tGo"]})
		}
	}
	if from == nil || to == nil {
		return nil, fmt.Errorf("start나 end 멈춤 자리가 없어 장면 구간을 셀 수 없음")
	}
	out["window"], out["pauses"] = []any{from, to}, pauses

	tree, roots, err := v.tree(out["branches"])
	if err != nil {
		return nil, err
	}
	out["tree"], out["roots"] = tree, roots
	binding := []any{map[string]any{"t": from, "bound": "", "held": []any{}, "by": -1}}
	held := []any{}
	prev := ""
	for _, e := range get(srcGuess, nameBinding) {
		by := -1
		switch b := v.out[e.ID]["by"]; len(b) {
		case 0:
		case 1:
			by = v.treeID[b[0]]
			// 떼어 둔 연결은 상태를 만든 마디로 다시 셉니다: doSuspend는 앞 상태의 묶인 연결을 쌓고 doResume은 뺍니다(spring_tree.go 442~451).
			th, err := v.up(v.r.Events[b[0]].Place, "method")
			if err != nil {
				return nil, err
			}
			switch {
			case strings.HasSuffix(th.Name, "DataSourceTransactionManager::doSuspend"):
				held = append(append([]any{}, held...), prev)
			case strings.HasSuffix(th.Name, "DataSourceTransactionManager::doResume") && len(held) > 0:
				held = append([]any{}, held[:len(held)-1]...)
			}
		default:
			return nil, fmt.Errorf("묶임 상태 %d의 by 연결 %d개", e.ID, len(b))
		}
		bound := v.placeName(e.Target)
		binding = append(binding, map[string]any{"t": micros(e.StartNS), "bound": bound, "held": append([]any{}, held...), "by": by})
		prev = bound
	}
	out["binding"] = binding
	slots := []any{}
	for _, e := range get(srcGuess, nameSlot) {
		end := any(0.0)
		if e.EndNS != nil {
			end = micros(e.EndNS)
		}
		slots = append(slots, map[string]any{"conn": v.placeName(e.Target), "t0": micros(e.StartNS), "t1": end})
	}
	out["slots"] = slots
	waits, err := v.waits(get(srcCont, nameWait))
	if err != nil {
		return nil, err
	}
	out["waits"] = waits
	pool, err := strconv.Atoi(fmt.Sprint(v.r.Attrs[srcTxlab]["pool"]))
	if err != nil {
		return nil, fmt.Errorf("run attrs txlab.pool: %w", err)
	}
	out["waitCycles"] = waitCycles(tree, waits, binding, conns, slots, pool)
	return out, nil
}

// waits 메서드는 서버 기다림을 되살립니다. conn은 장소(백엔드)의 연결, seq는 그 기다림을 품은 서버 창에 이어진 요청의 seq입니다.
func (v *view) waits(evs []atlas.Event) ([]any, error) {
	out := []any{}
	for _, e := range evs {
		win := v.in[e.ID]["contains"]
		if len(win) != 1 {
			return nil, fmt.Errorf("기다림 %d를 품은 서버 창 %d개", e.ID, len(win))
		}
		req := v.in[win[0]][nameServer]
		if len(req) != 1 {
			return nil, fmt.Errorf("서버 창 %d의 요청 %d개", win[0], len(req))
		}
		w := map[string]any{"conn": v.connOf[e.Place], "seq": v.seqOf[req[0]], "t0": micros(e.StartNS), "t1": micros(e.EndNS)}
		v.copyAttrs(w, e, srcCont, "xid", "end")
		out = append(out, w)
	}
	return out, nil
}

// node 값은 되돌리는 동안의 트리 마디 하나입니다. t0·last는 정렬과 효과 셈에 쓰는 시각(ns)입니다.
type node struct {
	e        atlas.Event
	m        map[string]any
	t0, last int64
	kids     []int // 사건 id
}

// tree 메서드는 해석기 출처의 구간(트리 마디)을 원본의 view.tree와 roots로 되살립니다.
//
// 스택에만 보인 마디의 t0·seen은 기록된 자손의 t0 최솟값·t1 최댓값, kids·roots는 t0 차례(같으면 id 차례), depth는 사건의 깊이,
// lane은 fn의 앞부분, branch는 run attrs의 갈림에서 자식 차례·끝난 꼴로 찾습니다(decode/spring_tree.go 131~246, spring_compare.go 82~111).
func (v *view) tree(branches any) ([]any, []any, error) {
	var evs []atlas.Event
	for _, e := range v.r.Events {
		if e.Source == srcTree && e.Kind == atlas.Span {
			evs = append(evs, e)
		}
	}
	nodes := map[int]*node{}
	for _, e := range evs {
		id, err := strconv.Atoi(fmt.Sprint(e.Attrs[srcTree]["id"]))
		if err != nil {
			return nil, nil, fmt.Errorf("트리 마디 %d의 id: %w", e.ID, err)
		}
		v.treeID[e.ID] = id
		nodes[e.ID] = &node{e: e}
	}
	for _, e := range evs {
		if e.Parent != nil {
			p, ok := nodes[*e.Parent]
			if !ok {
				return nil, nil, fmt.Errorf("트리 마디 %d의 부모 %d가 마디가 아님", e.ID, *e.Parent)
			}
			p.kids = append(p.kids, e.ID)
		}
	}
	// 시각: 기록된 마디는 구간, 스택에만 보인 마디는 기록된 자손에서 셉니다. 자손이 뒤에 오므로 id가 큰 것부터 올립니다.
	for k := len(evs) - 1; k >= 0; k-- {
		n := nodes[evs[k].ID]
		if n.e.StartNS != nil {
			n.t0, n.last = *n.e.StartNS, *n.e.EndNS
			continue
		}
		found := false
		var lo, hi int64
		var walk func(id int)
		walk = func(id int) {
			for _, c := range nodes[id].kids {
				ch := nodes[c]
				if ch.e.StartNS != nil {
					if !found || *ch.e.StartNS < lo {
						lo = *ch.e.StartNS
					}
					if !found || *ch.e.EndNS > hi {
						hi = *ch.e.EndNS
					}
					found = true
				}
				walk(c)
			}
		}
		walk(n.e.ID)
		if !found {
			return nil, nil, fmt.Errorf("스택에만 보인 마디 %d에 기록된 자손이 없어 시각을 셀 수 없음", n.e.ID)
		}
		n.t0, n.last = lo, hi
	}
	byStart := func(ids []int) {
		sort.SliceStable(ids, func(a, b int) bool { return nodes[ids[a]].t0 < nodes[ids[b]].t0 })
	}
	alts := map[string]any{}
	if b, ok := branches.(map[string]any); ok {
		alts = b
	}

	out := make([]any, 0, len(evs))
	var roots []int
	for _, e := range evs {
		n := nodes[e.ID]
		desc := v.r.Places[e.Place]
		method := v.r.Places[*desc.Parent]
		m := map[string]any{"fn": method.Name, "desc": desc.Name, "lane": laneOf(method.Name), "depth": e.Depth, "t0": atlas.NanosToMicros(n.t0)}
		v.copyAttrs(m, e, srcTree, "id", "label", "end", "hiddenKinds", "self", "why")
		if e.EndNS != nil {
			m["t1"] = micros(e.EndNS)
		} else {
			m["t1"], m["seen"] = nil, atlas.NanosToMicros(n.last)
		}

		// 스택: [마디, 숨은 칸(안쪽부터), 부모 메서드 안의 부른 줄]. 뿌리는 부른 줄이 없어 site 0이고 숨은 칸이 나머지 전부입니다.
		hidden := e.Stack[1:]
		m["parent"], m["site"] = -1, 0
		if e.Parent != nil {
			last := e.Stack[len(e.Stack)-1]
			hidden = e.Stack[1 : len(e.Stack)-1]
			m["parent"] = v.treeID[*e.Parent]
			line := -1
			if p := v.r.Places[last]; p.Loc != nil {
				line = p.Loc.Line
			}
			m["site"] = line
		} else {
			roots = append(roots, e.ID)
		}
		h := make([]any, 0, len(hidden))
		for k := len(hidden) - 1; k >= 0; k-- {
			h = append(h, v.ref(hidden[k]))
		}
		m["hidden"] = h
		byStart(n.kids)
		ids := make([]any, 0, len(n.kids))
		for _, c := range n.kids {
			ids = append(ids, v.treeID[c])
		}
		m["kids"] = ids
		n.m = m
		out = append(out, m)
	}
	// 효과와 갈림은 자식의 끝난 꼴과 다른 사건을 보므로 마디를 모두 만든 뒤에 붙입니다.
	for _, e := range evs {
		n := nodes[e.ID]
		effects, err := v.effects(n)
		if err != nil {
			return nil, nil, err
		}
		n.m["effects"] = effects
		fn := n.m["fn"].(string)
		if strings.HasPrefix(fn, "lab.") {
			continue
		}
		key := fn + n.m["desc"].(string)
		alt, err := altOf(n, nodes, alts[key])
		if err != nil {
			return nil, nil, fmt.Errorf("트리 마디 %d(%s): %w", e.ID, key, err)
		}
		n.m["branch"] = map[string]any{"key": key, "alt": alt}
	}
	byStart(roots)
	r := make([]any, 0, len(roots))
	for _, id := range roots {
		r = append(r, v.treeID[id])
	}
	return out, r, nil
}

// laneOf 함수는 호출이 그려질 칸입니다(decode/spring_tree.go 265~279와 같은 규칙).
func laneOf(fn string) string {
	for _, p := range []struct{ prefix, lane string }{{"lab.", "app"}, {"org.springframework.jdbc.core.", "jdbc"}, {"org.springframework.", "tx"}, {"com.zaxxer.hikari.", "hikari"}, {"org.postgresql.", "pg"}} {
		if strings.HasPrefix(fn, p.prefix) {
			return p.lane
		}
	}
	return "other"
}

// altOf 함수는 마디가 간 길의 번호를 갈림 모음(run attrs view.branches[key])에서 찾습니다. 길은 자식(t0 차례)의 fn+desc 차례(앱 코드는
// lab.*), 끝난 꼴(정상 return과 예외), 마지막 자식의 끝난 꼴입니다. 맞는 길이 정확히 하나가 아니면 오류입니다.
func altOf(n *node, nodes map[int]*node, group any) (int, error) {
	g, ok := group.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("갈림 모음에 없음")
	}
	alts, _ := g["alts"].([]any)
	seq := make([]any, 0, len(n.kids))
	for _, c := range n.kids {
		k := nodes[c].m
		if fn := k["fn"].(string); strings.HasPrefix(fn, "lab.") {
			seq = append(seq, "lab.*")
		} else {
			seq = append(seq, fn+k["desc"].(string))
		}
	}
	end := "return"
	if n.m["end"] != "return" {
		end = "exception"
	}
	lastExc := len(n.kids) > 0 && nodes[n.kids[len(n.kids)-1]].m["end"] != "return"
	want, err := json.Marshal(seq)
	if err != nil {
		return 0, err
	}
	hit := -1
	for i, raw := range alts {
		a, _ := raw.(map[string]any)
		got, err := json.Marshal(a["seq"])
		if err != nil {
			return 0, err
		}
		if string(got) == string(want) && a["end"] == end && a["lastExc"] == lastExc {
			if hit >= 0 {
				return 0, fmt.Errorf("맞는 길이 둘 이상")
			}
			hit = i
		}
	}
	if hit < 0 {
		return 0, fmt.Errorf("맞는 길이 없음")
	}
	return hit, nil
}

// effects 메서드는 마디 하나의 효과를 해석기의 차례(묶임 → 종류 → 잠듦 → 예외)로 되살립니다.
//
// 묶임 효과는 contains로 품은 추정 사건, wire는 연결한 요청을 연결 차례로 묶은 것, borrow·park·throw는 연결한 빌림·JFR 사건의 사본,
// autocommit·validate·rollbackOnly는 자식 순간입니다. 빈 값은 원본처럼 칸을 두지 않습니다(callEffect의 omitempty).
func (v *view) effects(n *node) ([]any, error) {
	out := []any{}
	for _, id := range v.out[n.e.ID]["contains"] {
		e := v.r.Events[id]
		m := map[string]any{"kind": e.Name}
		if e.Target != nil {
			m["conn"] = v.placeName(e.Target)
		}
		out = append(out, m)
	}

	// wire: 연결한 요청을 연결(장소) 차례로 묶고, 연결 안에서는 연결 차례입니다.
	byConn := map[int][]any{}
	var connOrder []int
	for _, q := range v.out[n.e.ID]["wire"] {
		p := v.r.Events[q].Place
		if _, ok := byConn[p]; !ok {
			connOrder = append(connOrder, p)
		}
		byConn[p] = append(byConn[p], v.seqOf[q])
	}
	sort.Ints(connOrder)
	for _, p := range connOrder {
		out = append(out, map[string]any{"kind": "wire", "conn": v.r.Places[p].Name, "seq": byConn[p]})
	}
	for _, id := range v.out[n.e.ID]["borrow"] {
		b := v.r.Events[id]
		m := map[string]any{"kind": "borrow"}
		if b.Target != nil {
			m["conn"] = v.placeName(b.Target)
		}
		v.copyAttrs(m, b, srcGuess, "result")
		if x, _ := v.attr(b, srcGuess, "idleMs"); fmt.Sprint(x) != "0" {
			m["idleMs"] = x
		}
		if x, _ := v.attr(b, srcGuess, "validated"); x == true {
			m["validated"] = true
		}
		out = append(out, m)
	}
	for _, id := range v.children[n.e.ID] {
		e := v.r.Events[id]
		if e.Source != srcTree || e.Kind != atlas.Instant {
			continue
		}
		m := map[string]any{"kind": e.Name}
		v.copyAttrs(m, e, srcTree, "value", "sent")
		out = append(out, m)
	}
	for _, id := range v.out[n.e.ID]["park"] {
		p := v.r.Events[id]
		m := map[string]any{"kind": "park", "t0": micros(p.StartNS), "t1": micros(p.EndNS)}
		m["at"], _ = v.attr(p, srcSpring, "under")
		out = append(out, m)
	}
	for _, id := range v.out[n.e.ID]["throw"] {
		t := v.r.Events[id]
		m := map[string]any{"kind": "throw", "t0": micros(t.StartNS)}
		v.copyAttrs(m, t, srcJFR, "class")
		if msg := v.text(t, srcJFR, "message"); msg != "" {
			m["message"] = msg
		}
		// 만든 자리는 예외 스택에서 마디와 메서드가 같은 첫 칸입니다(그 앞 칸은 추적하지 않는 메서드라 마디가 아님).
		fn := n.m["fn"].(string)
		for _, s := range t.Stack {
			if v.r.Places[s].Name == fn {
				m["at"] = v.ref(s)
				break
			}
		}
		if _, ok := m["at"]; !ok {
			return nil, fmt.Errorf("예외 %d의 스택에 마디 %s의 칸이 없음", t.ID, fn)
		}
		out = append(out, m)
	}
	return out, nil
}

// waitCycles 함수는 되살린 트리·기다림·묶임·연결·슬롯으로 기다림 고리를 해석기처럼 셉니다(decode/spring_tree.go 532~626).
//
// 서버 기다림마다 하나(그 xid를 받은 연결, 기다림 내내 열려 있던 보낸 마디, 기다림이 시작될 때 그 연결을 떼어 둔 스레드)와, 시간 초과로
// 끝난 빌리기 마디마다 하나(그 순간 쓰이던 슬롯, 그 연결들을 모두 떼어 둔 스레드)입니다.
func waitCycles(tree, waits, binding, conns, slots []any, pool int) []any {
	num := func(x any) float64 {
		switch n := x.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case json.Number:
			f, _ := n.Float64()
			return f
		}
		return 0
	}
	stateAt := func(t float64) map[string]any {
		var st map[string]any
		for _, raw := range binding {
			b := raw.(map[string]any)
			if num(b["t"]) <= t {
				st = b
			}
		}
		return st
	}
	heldBy := func(st map[string]any, conn string) bool {
		if st == nil {
			return false
		}
		for _, h := range st["held"].([]any) {
			if h == conn {
				return true
			}
		}
		return false
	}
	out := []any{}
	for _, raw := range waits {
		w := raw.(map[string]any)
		c := map[string]any{"kind": "server", "t0": w["t0"], "t1": w["t1"], "waiter": w["conn"], "xid": w["xid"], "holder": "", "thread": "", "call": -1, "heldBy": "", "cycle": false}
		for _, cr := range conns {
			cn := cr.(map[string]any)
			for _, qr := range cn["requests"].([]any) {
				s, _ := qr.(map[string]any)["server"].(map[string]any)
				evs, _ := s["events"].([]any)
				for _, er := range evs {
					e := er.(map[string]any)
					f, _ := e["f"].(map[string]any)
					if e["kind"] == "xid" && f["xid"] == w["xid"] {
						c["holder"] = cn["name"]
					}
				}
			}
		}
		for _, nr := range tree {
			n := nr.(map[string]any)
			if num(n["t0"]) > num(w["t0"]) || n["t1"] == nil || num(n["t1"]) < num(w["t1"]) {
				continue
			}
			for _, er := range n["effects"].([]any) {
				e := er.(map[string]any)
				if e["kind"] != "wire" || e["conn"] != w["conn"] {
					continue
				}
				for _, s := range e["seq"].([]any) {
					if num(s) == num(w["seq"]) {
						c["call"], c["thread"] = n["id"], sceneThread
					}
				}
			}
		}
		if h := c["holder"].(string); h != "" && heldBy(stateAt(num(w["t0"])), h) {
			c["heldBy"] = sceneThread
		}
		c["cycle"] = num(c["call"]) >= 0 && c["heldBy"] != "" && c["heldBy"] == c["thread"]
		out = append(out, c)
	}
	for _, nr := range tree {
		n := nr.(map[string]any)
		timedOut := false
		for _, er := range n["effects"].([]any) {
			e := er.(map[string]any)
			timedOut = timedOut || e["kind"] == "borrow" && e["result"] == "timeout"
		}
		if !timedOut || n["t1"] == nil {
			continue
		}
		t0 := num(n["t0"])
		c := map[string]any{"kind": "pool", "t0": n["t0"], "t1": n["t1"], "waiter": "", "xid": "", "holder": "", "thread": sceneThread, "call": n["id"], "heldBy": "", "cycle": false}
		if pool != 0 {
			c["pool"] = pool
		}
		var inUse []string
		for _, sr := range slots {
			s := sr.(map[string]any)
			if num(s["t0"]) <= t0 && (num(s["t1"]) == 0 || num(s["t1"]) > t0) {
				inUse = append(inUse, s["conn"].(string))
			}
		}
		st := stateAt(t0)
		c["holder"] = strings.Join(inUse, ", ")
		all := len(inUse) > 0 && st != nil
		for _, conn := range inUse {
			all = all && heldBy(st, conn)
		}
		if all {
			c["heldBy"] = sceneThread
		}
		c["cycle"] = pool > 0 && len(inUse) >= pool && all
		out = append(out, c)
	}
	return out
}
