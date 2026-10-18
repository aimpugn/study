package spring

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// view 값은 되돌리기에 쓰는 run의 색인입니다. 연결은 종류마다 출발·도착 사건으로 찾고, 자식은 id 차례입니다.
type view struct {
	r        *atlas.Run
	out, in  map[int]map[string][]int
	children map[int][]int
	connOf   map[int]string // 백엔드 프로세스 장소 → 연결 이름
	seqOf    map[int]int    // 요청 사건 → 그 연결 안의 seq
	treeID   map[int]int    // 트리 마디 사건 → 마디 id
}

// newView 함수는 run에서 연결과 부모의 색인을 만들고, 메시지의 백엔드 쪽 끝으로 연결과 백엔드의 짝을 셉니다.
// 연결 하나의 메시지가 두 백엔드를 가리키거나 백엔드 하나가 두 연결의 것이면 오류입니다.
func newView(r *atlas.Run) (*view, error) {
	v := &view{r: r, out: map[int]map[string][]int{}, in: map[int]map[string][]int{}, children: map[int][]int{},
		connOf: map[int]string{}, seqOf: map[int]int{}, treeID: map[int]int{}}
	for _, l := range r.Links {
		if v.out[l.Src] == nil {
			v.out[l.Src] = map[string][]int{}
		}
		if v.in[l.Dst] == nil {
			v.in[l.Dst] = map[string][]int{}
		}
		v.out[l.Src][l.Kind] = append(v.out[l.Src][l.Kind], l.Dst)
		v.in[l.Dst][l.Kind] = append(v.in[l.Dst][l.Kind], l.Src)
	}
	pidOf := map[string]int{}
	for _, e := range r.Events {
		if e.Parent != nil {
			v.children[*e.Parent] = append(v.children[*e.Parent], e.ID)
		}
		if e.Source != srcWireMac {
			continue
		}
		end := e.Target
		if end == nil {
			end = e.From
		}
		if end == nil {
			return nil, fmt.Errorf("메시지 %d의 백엔드 쪽 끝이 없음", e.ID)
		}
		name := r.Places[e.Place].Name
		if p, ok := pidOf[name]; ok && p != *end {
			return nil, fmt.Errorf("연결 %s의 백엔드가 둘(%d, %d)", name, p, *end)
		}
		if c, ok := v.connOf[*end]; ok && c != name {
			return nil, fmt.Errorf("백엔드 %d의 연결이 둘(%s, %s)", *end, c, name)
		}
		pidOf[name], v.connOf[*end] = *end, name
	}
	return v, nil
}

// attr 메서드는 사건의 출처 src가 낸 값 key를 돌려줍니다.
func (v *view) attr(e atlas.Event, src, key string) (any, bool) {
	val, ok := e.Attrs[src][key]
	return val, ok
}

// text 메서드는 사건의 출처 src가 낸 글 값 key를 돌려줍니다. 없거나 글이 아니면 빈 글입니다.
func (v *view) text(e atlas.Event, src, key string) string {
	s, _ := e.Attrs[src][key].(string)
	return s
}

// copyAttrs 메서드는 사건의 출처 src가 낸 값 가운데 keys를 out에 옮깁니다. 없는 열쇠는 옮기지 않습니다.
func (v *view) copyAttrs(out map[string]any, e atlas.Event, src string, keys ...string) {
	for _, k := range keys {
		if val, ok := v.attr(e, src, k); ok {
			out[k] = val
		}
	}
}

// micros 함수는 사건 시각을 원본의 마이크로초로 씁니다.
func micros(ns *int64) any { return atlas.NanosToMicros(*ns) }

// placeName 메서드는 장소 id의 이름을 돌려줍니다. id가 nil이면 빈 글입니다.
func (v *view) placeName(id *int) string {
	if id == nil {
		return ""
	}
	return v.r.Places[*id].Name
}

// up 메서드는 장소 id에서 부모를 따라 올라가 종류가 kind인 첫 장소를 돌려줍니다. 없으면 오류입니다.
func (v *view) up(id int, kind string) (atlas.Place, error) {
	for p := id; ; {
		pl := v.r.Places[p]
		if pl.Kind == kind {
			return pl, nil
		}
		if pl.Parent == nil {
			return atlas.Place{}, fmt.Errorf("장소 %d 위에 %s가 없음", id, kind)
		}
		p = *pl.Parent
	}
}

// ref 메서드는 자바 프레임 장소를 원본의 글 "메서드:줄"로 씁니다. 위치 없는 프레임의 줄은 −1입니다.
func (v *view) ref(id int) string {
	p := v.r.Places[id]
	line := -1
	if p.Loc != nil {
		line = p.Loc.Line
	}
	return p.Name + ":" + strconv.Itoa(line)
}

// refs 메서드는 자바 스택(0번이 사건의 장소)을 프레임 글 목록으로 씁니다. 원본도 안쪽부터입니다.
func (v *view) refs(stack []int) []any {
	out := make([]any, 0, len(stack))
	for _, s := range stack {
		out = append(out, v.ref(s))
	}
	return out
}

// wireView 함수는 맥 쪽·컨테이너 쪽 메시지 사건 짝에서 원본의 메시지 하나({t, tc, dir, type, len, name?, text?, f?})를 만듭니다.
// 백엔드가 target이면 앱이 보낸 F, from이면 서버가 보낸 B입니다.
func wireView(mac, cnt atlas.Event) map[string]any {
	m := map[string]any{"t": micros(mac.StartNS), "tc": micros(cnt.StartNS), "type": mac.Name, "dir": "B"}
	if mac.Target != nil {
		m["dir"] = "F"
	}
	for _, k := range []string{"len", "name", "text", "f"} {
		if val, ok := mac.Attrs[srcWireMac][k]; ok {
			m[k] = val
		}
	}
	return m
}

// message 메서드는 맥 쪽 메시지 사건 id의 원본 꼴을 돌려줍니다. 컨테이너 쪽 짝이 하나가 아니면 오류입니다.
func (v *view) message(id int) (map[string]any, error) {
	pair := v.out[id]["message"]
	if len(pair) != 1 {
		return nil, fmt.Errorf("메시지 %d의 컨테이너 쪽 짝 %d개", id, len(pair))
	}
	return wireView(v.r.Events[id], v.r.Events[pair[0]]), nil
}

// Export 함수는 Import가 만든 run에서 기록 한 벌(RUNS의 값 하나)을 다시 만듭니다.
//
// run 밖의 것은 읽지 않습니다. 해석기가 다른 칸에서 셈한 칸은 해석기와 같은 규칙으로 다시 셉니다. 사건을 가를 수 없거나 되살리는 데
// 필요한 값이 없으면 오류를 돌려줍니다.
func Export(r *atlas.Run) (map[string]any, error) {
	v, err := newView(r)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	for src, keys := range map[string][]string{srcTxlab: {"scene", "pool", "javaVersion", "javaExit", "title", "tables"}, srcSpring: {"clock", "joins", "siteRepair"}} {
		for _, k := range keys {
			val, ok := r.Attrs[src][k]
			if !ok {
				return nil, fmt.Errorf("run attrs %s.%s 없음", src, k)
			}
			out[k] = val
		}
	}
	if s, ok := r.Attrs[srcSrc]["src"]; ok {
		out["src"] = s
	}

	// 사건을 출처와 이름으로 가릅니다. 차례는 id 차례(같은 출처 안에서 기록 차례)입니다.
	by := map[string][]atlas.Event{}
	for _, e := range r.Events {
		by[e.Source+"\x00"+e.Name] = append(by[e.Source+"\x00"+e.Name], e)
	}
	get := func(src, name string) []atlas.Event { return by[src+"\x00"+name] }

	conns, err := v.conns(get)
	if err != nil {
		return nil, err
	}
	out["conns"] = conns
	sessions := map[string]any{}
	for _, c := range conns {
		m := c.(map[string]any)
		sessions[string(m["pid"].(json.Number))] = m["name"]
	}
	out["sessions"] = sessions

	if out["transactions"], err = v.transactions(get(srcMac, nameTx)); err != nil {
		return nil, err
	}
	checkpoints, err := v.checkpoints(get(srcTxlab, nameCheckpoint))
	if err != nil {
		return nil, err
	}
	out["checkpoints"] = checkpoints
	if out["blocked"], err = v.blocked(get(srcTxlab, nameBlocked)); err != nil {
		return nil, err
	}
	notes := []any{}
	for _, e := range get(srcTxlab, nameNote) {
		n := map[string]any{"t": micros(e.StartNS)}
		v.copyAttrs(n, e, srcTxlab, "where", "chain", "message")
		notes = append(notes, n)
	}
	out["notes"] = notes
	if out["captures"], err = v.captures(get(srcObserve, nameCapture)); err != nil {
		return nil, err
	}
	borrows := []any{}
	for _, e := range get(srcGuess, nameBorrow) {
		b := map[string]any{"t0": micros(e.StartNS), "t1": micros(e.EndNS), "thread": r.Places[e.Place].Name, "conn": v.placeName(e.Target)}
		v.copyAttrs(b, e, srcGuess, "validated", "idleMs", "result")
		borrows = append(borrows, b)
	}
	out["borrows"] = borrows
	java, err := v.java()
	if err != nil {
		return nil, err
	}
	out["java"] = java
	vw, err := v.view(get, checkpoints, conns)
	if err != nil {
		return nil, err
	}
	out["view"] = vw
	return out, nil
}

// conns 메서드는 연결 장소마다 연결 하나({conn, name, pid, thread, …, messages, requests})를 되살립니다.
func (v *view) conns(get func(src, name string) []atlas.Event) ([]any, error) {
	r := v.r
	life := map[int]atlas.Event{}
	for _, e := range get(srcTxlab, nameConn) {
		life[e.Place] = e
	}
	msgs, reqs := map[int][]atlas.Event{}, map[int][]atlas.Event{}
	for _, e := range r.Events {
		switch {
		case e.Source == srcWireMac:
			msgs[e.Place] = append(msgs[e.Place], e)
		case e.Source == srcMac && e.Name == nameRequest:
			reqs[e.Place] = append(reqs[e.Place], e)
		}
	}
	out := []any{}
	for _, p := range r.Places {
		if p.Kind != "connection" {
			continue
		}
		lf, ok := life[p.ID]
		if !ok {
			return nil, fmt.Errorf("연결 %s의 수명 구간 없음", p.Name)
		}
		n, err := strconv.Atoi(strings.TrimPrefix(p.Name, "C"))
		if err != nil {
			return nil, fmt.Errorf("연결 이름 %s에서 번호를 읽지 못함", p.Name)
		}
		c := map[string]any{"conn": n, "name": p.Name, "accept": micros(lf.StartNS), "close": micros(lf.EndNS)}
		v.copyAttrs(c, lf, srcWireMac, "params")
		v.copyAttrs(c, lf, srcSpring, "startup", "lastMac", "lastCont", "join")
		v.copyAttrs(c, lf, srcTxlab, "bridgeErr")
		ms := msgs[p.ID]
		if len(ms) == 0 {
			return nil, fmt.Errorf("연결 %s에 메시지가 없어 백엔드를 셀 수 없음", p.Name)
		}
		end := ms[0].Target
		if end == nil {
			end = ms[0].From
		}
		c["pid"] = json.Number(r.Places[*end].Name)
		messages := make([]any, 0, len(ms))
		firstQ := (*int64)(nil)
		for _, m := range ms {
			w, err := v.message(m.ID)
			if err != nil {
				return nil, err
			}
			messages = append(messages, w)
			if firstQ == nil && m.Target != nil && m.Name == "Q" {
				firstQ = m.StartNS
			}
		}
		c["messages"] = messages
		c["thread"] = v.threadAt(firstQ)
		requests := make([]any, 0, len(reqs[p.ID]))
		for k, q := range reqs[p.ID] {
			v.seqOf[q.ID] = k
			req, err := v.request(q, k)
			if err != nil {
				return nil, err
			}
			requests = append(requests, req)
		}
		c["requests"] = requests
		out = append(out, c)
	}
	return out, nil
}

// threadAt 메서드는 연결을 연 스레드를 해석기처럼 셉니다: 첫 Q의 맥 시각을 감싼 QueryExecutorImpl.execute JFR 구간 가운데 구간 차례로
// 처음 것의 스레드입니다(decode/spring.go 511~514, spring_jfr.go 414~421). JFR 구간의 사건 차례는 원본 spans의 차례입니다. Q가 없으면 빈 글입니다.
func (v *view) threadAt(t *int64) string {
	if t == nil {
		return ""
	}
	for _, e := range v.r.Events {
		if e.Source != srcJFR || e.Kind != atlas.Span || e.Name != "org.postgresql.core.v3.QueryExecutorImpl::execute" || v.r.Places[e.Place].Kind != "descriptor" {
			continue
		}
		if *e.StartNS <= *t && *t <= *e.EndNS {
			th, err := v.up(e.Place, "thread")
			if err != nil {
				return ""
			}
			return th.Name
		}
	}
	return ""
}

// request 메서드는 요청 하나를 되살립니다. seq·types·sql·stmt·status·tags·error는 contains로 품은 메시지에서 해석기처럼 다시 셉니다
// (decode/spring.go 258~308). seq 0은 접속 인증입니다.
func (v *view) request(e atlas.Event, seq int) (map[string]any, error) {
	q := map[string]any{"seq": seq, "sent": micros(e.StartNS), "done": micros(e.EndNS)}
	types, stmt, status := []any{}, "", ""
	var sql, tags []any
	var errm map[string]any
	members := v.out[e.ID]["contains"]

	// 요청 구간의 시작은 품은 첫 앱 메시지, 끝은 마지막 메시지(ReadyForQuery)의 시각입니다. 다르면 품은 메시지가 그 요청의 것이 아닙니다.
	firstF := -1
	for _, id := range members {
		if v.r.Events[id].Target != nil {
			firstF = id
			break
		}
	}
	if firstF < 0 || len(members) == 0 {
		return nil, fmt.Errorf("요청 %d가 품은 앱 메시지가 없음", e.ID)
	}
	if z := v.r.Events[members[len(members)-1]]; *v.r.Events[firstF].StartNS != *e.StartNS || z.Name != "Z" || *z.StartNS != *e.EndNS {
		return nil, fmt.Errorf("요청 %d의 시작·끝이 품은 첫 앱 메시지와 ReadyForQuery의 시각과 다름", e.ID)
	}
	for _, id := range members {
		m := v.r.Events[id]
		f, _ := m.Attrs[srcWireMac]["f"].(map[string]any)
		text := v.text(m, srcWireMac, "text")
		if m.Target != nil {
			if seq == 0 {
				continue
			}
			types = append(types, m.Name)
			if (m.Name == "P" || m.Name == "B") && stmt == "" {
				stmt = v.text(m, srcWireMac, "name")
			}
			if m.Name == "Q" || m.Name == "P" {
				sql = append(sql, text)
			}
			continue
		}
		switch m.Name {
		case "C":
			tags = append(tags, text)
		case "E":
			code, _ := f["code"].(string)
			where, _ := f["where"].(string)
			errm = map[string]any{"code": code, "message": text, "where": where}
		case "Z":
			status = text
		}
	}
	if seq == 0 {
		sql = []any{authSQL}
	}
	q["types"], q["status"] = types, status
	if sql == nil {
		q["sql"] = nil
	} else {
		q["sql"] = sql
	}
	if stmt != "" {
		q["stmt"] = stmt
	}
	if tags != nil {
		q["tags"] = tags
	}
	if errm != nil {
		q["error"] = errm
	}
	win := v.out[e.ID][nameServer]
	switch len(win) {
	case 0:
	case 1:
		parts := make([]string, 0, len(sql))
		if seq != 0 {
			for _, s := range sql {
				parts = append(parts, s.(string))
			}
		}
		s, err := v.window(v.r.Events[win[0]], strings.Join(parts, "; "))
		if err != nil {
			return nil, err
		}
		q["server"] = s
	default:
		return nil, fmt.Errorf("요청 %d의 서버 창 %d개", e.ID, len(win))
	}
	return q, nil
}

// window 메서드는 서버 창 하나를 되살립니다. 품은 것은 연결 차례로 가릅니다: 이름이 check면 줄 판정, curcid가 있으면 스냅숏, 해석기의
// 기다림은 view.waits라 빼고, 나머지는 사건입니다. sql은 요청의 sql을 "; "로 이은 것입니다.
func (v *view) window(w atlas.Event, sql string) (map[string]any, error) {
	s := map[string]any{"sql": sql, "t0": micros(w.StartNS), "t1": micros(w.EndNS)}
	v.copyAttrs(s, w, srcRec, "catalogSnapshots")
	v.copyAttrs(s, w, srcSpring, "in", "out")
	events, snaps, checks := []any{}, []any{}, []any{}
	for _, id := range v.out[w.ID]["contains"] {
		m := v.r.Events[id]
		_, isSnap := v.attr(m, srcRec, "curcid")
		switch {
		case m.Source == srcCont:
		case m.Name == nameCheck:
			c, err := v.check(m)
			if err != nil {
				return nil, err
			}
			checks = append(checks, c)
		case isSnap:
			sn := map[string]any{"kind": m.Name, "t": micros(m.StartNS), "stack": load.StackOf(v.r.Places, m.Stack)}
			v.copyAttrs(sn, m, srcRec, "curcid", "xip", "xmax", "xmin")
			snaps = append(snaps, sn)
		default:
			ev := map[string]any{"kind": m.Name, "t": micros(m.StartNS)}
			if m.EndNS != nil {
				ev["t1"] = micros(m.EndNS)
			}
			if m.Stack != nil {
				ev["stack"] = load.StackOf(v.r.Places, m.Stack)
			}
			v.copyAttrs(ev, m, srcRec, "f")
			events = append(events, ev)
		}
	}
	s["events"], s["snapshots"], s["checks"] = events, snaps, checks
	return s, nil
}

// check 메서드는 줄 판정 하나와 그 자식(줄, 판정 경로와 경로 안의 호출)을 되살립니다. 경로 사건이 없으면 path는 null입니다.
func (v *view) check(m atlas.Event) (map[string]any, error) {
	c := map[string]any{"t": micros(m.StartNS), "stack": load.StackOf(v.r.Places, m.Stack)}
	v.copyAttrs(c, m, srcRec, "ctid", "tup", "table", "mask", "xmin", "xmax", "snapXmin", "snapXmax")
	v.copyAttrs(c, m, srcSpring, "visible")
	lines := []any{}
	var path []any
	for _, id := range v.children[m.ID] {
		k := v.r.Events[id]
		switch k.Name {
		case nameLine:
			lines = append(lines, v.call(k))
		case namePath:
			p := v.call(k)
			v.copyAttrs(p, k, srcRec, "arg", "ret")
			if kids := v.children[id]; len(kids) > 0 {
				inner := []any{}
				for _, j := range kids {
					inner = append(inner, v.call(v.r.Events[j]))
				}
				p["inner"] = inner
			}
			path = append(path, p)
		default:
			return nil, fmt.Errorf("줄 판정 %d의 자식 %d 이름 %q", m.ID, id, k.Name)
		}
	}
	c["lines"] = lines
	if path == nil {
		c["path"] = nil
	} else {
		c["path"] = path
	}
	return c, nil
}

// call 메서드는 줄 판정 안의 호출 하나를 되살립니다. fn은 장소(부른 함수), at은 스택 1번(부른 줄)입니다.
func (v *view) call(k atlas.Event) map[string]any {
	return map[string]any{"fn": v.r.Places[k.Place].Name, "at": load.LocText(v.r.Places[k.Stack[1]].Loc)}
}

// transactions 메서드는 트랜잭션을 되살립니다. first·last는 품은 요청의 seq, xids·clog는 품은 요청의 서버 창 사건, next는 연결한 메시지에서 셉니다.
func (v *view) transactions(evs []atlas.Event) ([]any, error) {
	out := []any{}
	for _, e := range evs {
		n, err := strconv.Atoi(strings.TrimPrefix(v.r.Places[e.Place].Name, "C"))
		if err != nil {
			return nil, err
		}
		reqs := v.out[e.ID]["contains"]
		if len(reqs) == 0 {
			return nil, fmt.Errorf("트랜잭션 %d가 품은 요청이 없음", e.ID)
		}
		seqs := make([]int, 0, len(reqs))
		for _, q := range reqs {
			seqs = append(seqs, v.seqOf[q])
		}
		order := append([]int{}, reqs...)
		sort.SliceStable(order, func(a, b int) bool { return v.seqOf[order[a]] < v.seqOf[order[b]] })
		sort.Ints(seqs)
		x := map[string]any{"conn": n, "first": seqs[0], "last": seqs[len(seqs)-1], "t0": micros(e.StartNS), "t1": micros(e.EndNS)}
		v.copyAttrs(x, e, srcMac, "end")
		xids, clog := []any{}, []any{}
		for _, q := range order {
			for _, w := range v.out[q][nameServer] {
				level := any("1")
				for _, id := range v.out[w]["contains"] {
					m := v.r.Events[id]
					if m.Source != srcRec || m.Name == nameCheck {
						continue
					}
					f, _ := m.Attrs[srcRec]["f"].(map[string]any)
					switch m.Name {
					case "assign_xid":
						level = f["level"]
					case "xid":
						xids = append(xids, map[string]any{"xid": f["xid"], "level": level, "t": micros(m.StartNS)})
					case "commit_clog", "abort_clog":
						clog = append(clog, map[string]any{"kind": m.Name, "xid": f["xid"], "nsub": f["nsub"], "t": micros(m.StartNS)})
					}
				}
			}
		}
		x["xids"], x["clog"] = xids, clog
		next := v.out[e.ID]["next"]
		if len(next) != 1 {
			return nil, fmt.Errorf("트랜잭션 %d의 next 연결 %d개", e.ID, len(next))
		}
		if x["next"], err = v.message(next[0]); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}

// checkpoints 메서드는 멈춤 자리를 되살립니다. tCapture는 same 연결이 가리키는 관찰의 시각입니다.
func (v *view) checkpoints(evs []atlas.Event) ([]any, error) {
	out := []any{}
	for _, e := range evs {
		c := map[string]any{"t": micros(e.StartNS), "tGo": micros(e.EndNS)}
		v.copyAttrs(c, e, srcTxlab, "n", "label")
		t, err := v.sameCapture(e.ID)
		if err != nil {
			return nil, err
		}
		c["tCapture"] = t
		out = append(out, c)
	}
	return out, nil
}

// sameCapture 메서드는 사건 id에서 same 연결로 이은 관찰의 시각을 돌려줍니다. 하나가 아니면 오류입니다.
func (v *view) sameCapture(id int) (any, error) {
	s := v.out[id]["same"]
	if len(s) != 1 {
		return nil, fmt.Errorf("사건 %d의 same 연결 %d개", id, len(s))
	}
	return micros(v.r.Events[s[0]].StartNS), nil
}

// blocked 메서드는 막힘 관찰을 되살립니다. after는 after 연결이 가리키는 멈춤 자리의 n, waiters는 target 백엔드의 연결 이름입니다.
func (v *view) blocked(evs []atlas.Event) ([]any, error) {
	out := []any{}
	for _, e := range evs {
		b := map[string]any{"t": micros(e.StartNS), "waiters": []any{}}
		after := v.out[e.ID]["after"]
		if len(after) != 1 {
			return nil, fmt.Errorf("막힘 관찰 %d의 after 연결 %d개", e.ID, len(after))
		}
		b["after"], _ = v.attr(v.r.Events[after[0]], srcTxlab, "n")
		if e.Target != nil {
			name, ok := v.connOf[*e.Target]
			if !ok {
				return nil, fmt.Errorf("막힘 관찰 %d의 백엔드에 연결이 없음", e.ID)
			}
			b["waiters"] = []any{name}
		}
		t, err := v.sameCapture(e.ID)
		if err != nil {
			return nil, err
		}
		b["tCapture"] = t
		out = append(out, b)
	}
	return out, nil
}

// captures 메서드는 관찰을 되살립니다. afterStep은 same 연결로 들어온 멈춤 자리의 n(막힘 관찰이면 그 after), span은 [시작, 끝],
// reads는 자식 구간(끝이 비면 시작과 같음), overlap은 연결한 메시지입니다.
func (v *view) captures(evs []atlas.Event) ([]any, error) {
	out := []any{}
	for _, e := range evs {
		c := map[string]any{"t": micros(e.StartNS), "span": []any{micros(e.StartNS), micros(e.EndNS)}}
		v.copyAttrs(c, e, srcObserve, "activity", "label", "locks", "pages", "running", "wal", "xacts")
		v.copyAttrs(c, e, srcSpring, "visible")
		same := v.in[e.ID]["same"]
		if len(same) != 1 {
			return nil, fmt.Errorf("관찰 %d로 들어온 same 연결 %d개", e.ID, len(same))
		}
		src := v.r.Events[same[0]]
		switch src.Name {
		case nameCheckpoint:
			c["afterStep"], _ = v.attr(src, srcTxlab, "n")
		case nameBlocked:
			after := v.out[src.ID]["after"]
			if len(after) != 1 {
				return nil, fmt.Errorf("막힘 관찰 %d의 after 연결 %d개", src.ID, len(after))
			}
			c["afterStep"], _ = v.attr(v.r.Events[after[0]], srcTxlab, "n")
		default:
			return nil, fmt.Errorf("관찰 %d의 same 짝 %d가 멈춤 자리도 막힘 관찰도 아님", e.ID, src.ID)
		}
		reads := map[string]any{}
		for _, id := range v.children[e.ID] {
			k := v.r.Events[id]
			end := k.EndNS
			if end == nil {
				end = k.StartNS
			}
			reads[k.Name] = []any{micros(k.StartNS), micros(end)}
		}
		c["reads"] = reads
		overlap := []any{}
		for _, id := range v.out[e.ID]["overlap"] {
			mac := v.in[id]["message"]
			if len(mac) != 1 {
				return nil, fmt.Errorf("겹친 메시지 %d의 맥 쪽 짝 %d개", id, len(mac))
			}
			m, err := v.message(mac[0])
			if err != nil {
				return nil, err
			}
			overlap = append(overlap, m)
		}
		c["overlap"] = overlap
		out = append(out, c)
	}
	return out, nil
}

// java 메서드는 JFR 구간·예외·잠듦과 해석기의 JFR 요약(run attrs)을 되살립니다. 구간은 서술자 장소, 잠듦은 프레임 장소에 놓인 사건입니다.
func (v *view) java() (map[string]any, error) {
	spans, throws, parks := []any{}, []any{}, []any{}
	for _, e := range v.r.Events {
		if e.Source != srcJFR {
			continue
		}
		place := v.r.Places[e.Place]
		th, err := v.up(e.Place, "thread")
		if err != nil {
			return nil, err
		}
		switch {
		case e.Kind == atlas.Span && place.Kind == "descriptor":
			m := v.r.Places[*place.Parent]
			s := map[string]any{"fn": m.Name, "desc": place.Name, "thread": th.Name, "t0": micros(e.StartNS), "t1": micros(e.EndNS)}
			v.copyAttrs(s, e, srcSpring, "caller", "depth", "twice")
			spans = append(spans, s)
		case e.Name == nameThrow:
			t := map[string]any{"t": micros(e.StartNS), "thread": th.Name, "at": v.refs(e.Stack)}
			v.copyAttrs(t, e, srcJFR, "class", "message")
			throws = append(throws, t)
		case e.Name == namePark:
			p := map[string]any{"t0": micros(e.StartNS), "t1": micros(e.EndNS), "thread": th.Name, "at": v.refs(e.Stack)}
			v.copyAttrs(p, e, srcSpring, "under")
			parks = append(parks, p)
		default:
			return nil, fmt.Errorf("JFR 사건 %d(%s)를 가를 수 없음", e.ID, e.Name)
		}
	}
	out := map[string]any{"spans": spans, "throws": throws, "parks": parks}
	for _, k := range []string{"dups", "stackOnly", "dropped"} {
		val, ok := v.r.Attrs[srcSpring]["java."+k]
		if !ok {
			return nil, fmt.Errorf("run attrs spring.java.%s 없음", k)
		}
		out[k] = val
	}
	return out, nil
}
