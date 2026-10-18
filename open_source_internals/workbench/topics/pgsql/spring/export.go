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

// view 값은 되돌리기에 쓰는 run의 색인입니다. 연결은 종류마다 출발·도착 사건으로 찾고(상대 사건의 id 차례, load.LinkIndex), 자식은 id 차례입니다.
type view struct {
	r        *atlas.Run
	out, in  map[int]map[string][]int
	children map[int][]int
	connOf   map[int]string         // 백엔드 프로세스 장소 → 연결 이름
	seqOf    map[int]int            // 요청 사건 → 그 연결 안의 seq
	treeID   map[int]int            // 트리 마디 사건 → 마디 id
	reqOut   map[int]map[string]any // 요청 사건 → 되살린 요청
}

// newView 함수는 run에서 연결과 부모의 색인을 만들고, 메시지의 백엔드 쪽 끝으로 연결과 백엔드의 짝을 셉니다.
// 연결 하나의 메시지가 두 백엔드를 가리키거나 백엔드 하나가 두 연결의 것이면 오류입니다.
func newView(r *atlas.Run) (*view, error) {
	v := &view{r: r, children: map[int][]int{}, connOf: map[int]string{}, seqOf: map[int]int{}, treeID: map[int]int{}, reqOut: map[int]map[string]any{}}
	v.out, v.in = load.LinkIndex(r)
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

// micros 함수는 사건 시각을 원본의 마이크로초로 씁니다. 시각이 비면 되살릴 수 없어 panic하고, Export가 그것을 오류로 돌려줍니다.
func micros(ns *int64) any {
	if ns == nil {
		panic("되살리는 데 필요한 사건 시각이 비어 있음")
	}
	return atlas.NanosToMicros(*ns)
}

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
// 백엔드가 target이면 앱이 보낸 F, from이면 서버가 보낸 B입니다. f는 "f.<칸>" attrs와, BackendKeyData(K)면 from 백엔드의 pid로 다시 짭니다.
func wireView(r *atlas.Run, mac, cnt atlas.Event) map[string]any {
	m := map[string]any{"t": micros(mac.StartNS), "tc": micros(cnt.StartNS), "type": mac.Name, "dir": "B"}
	if mac.Target != nil {
		m["dir"] = "F"
	}
	for _, k := range []string{"len", "name", "text"} {
		if val, ok := mac.Attrs[srcWireMac][k]; ok {
			m[k] = val
		}
	}
	f := map[string]any{}
	for _, k := range []string{"auth", "code", "severity", "where"} {
		if val, ok := mac.Attrs[srcWireMac]["f."+k]; ok {
			f[k] = val
		}
	}
	if mac.Name == "K" && mac.From != nil {
		f["pid"] = r.Places[*mac.From].Name
	}
	if len(f) > 0 {
		m["f"] = f
	}
	return m
}

// message 메서드는 맥 쪽 메시지 사건 id의 원본 꼴을 돌려줍니다. 컨테이너 쪽 짝이 하나가 아니면 오류입니다.
func (v *view) message(id int) (map[string]any, error) {
	pair := v.out[id]["message"]
	if len(pair) != 1 {
		return nil, fmt.Errorf("메시지 %d의 컨테이너 쪽 짝 %d개", id, len(pair))
	}
	return wireView(v.r, v.r.Events[id], v.r.Events[pair[0]]), nil
}

// Export 함수는 Import가 만든 run에서 기록 한 벌(RUNS의 값 하나)을 다시 만듭니다.
//
// run 밖의 것은 읽지 않습니다. 해석기가 다른 칸에서 셈한 칸은 해석기와 같은 규칙으로 다시 셉니다. 사건을 가를 수 없거나 되살리는 데
// 필요한 값이 없으면 오류를 돌려줍니다. 형식 검사는 통과하지만 적재기가 만들지 않는 꼴(끝이 빈 구간, 빈 스택 등)을 만나 되살리다 멈추면
// 그것도 오류로 돌려줍니다(코드 검토 지적 9).
func Export(r *atlas.Run) (out map[string]any, err error) {
	defer func() {
		if p := recover(); p != nil {
			out, err = nil, fmt.Errorf("되살리다 멈춤(적재기가 만들지 않는 꼴의 run): %v", p)
		}
	}()
	return export(r)
}

// export 함수는 Export의 본문입니다.
func export(r *atlas.Run) (map[string]any, error) {
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
		v.copyAttrs(c, lf, srcSpring, "join")
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
		// 두 중계기는 같은 순번·종류로 잘랐으므로(적재 조건) 마지막 메시지의 종류가 같습니다.
		c["startup"], c["lastMac"], c["lastCont"] = startupOf(ms), ms[len(ms)-1].Name, ms[len(ms)-1].Name
		c["thread"] = v.threadAt(firstQ)
		requests := make([]any, 0, len(reqs[p.ID]))
		for k, q := range reqs[p.ID] {
			v.seqOf[q.ID] = k
			req, err := v.request(q, k)
			if err != nil {
				return nil, err
			}
			v.reqOut[q.ID] = req
			requests = append(requests, req)
		}
		c["requests"] = requests
		out = append(out, c)
	}
	return out, nil
}

// startupOf 함수는 연결의 접속 인증 차례(startup)를 해석기처럼 셉니다(decode/spring.go 495~509): 첫 ReadyForQuery까지의 맥 쪽 메시지
// 가운데 ssl·sslreply·startup·R·K·Z를 종류로 적고, sslreply와 Z는 상태(text)를, 그 밖은 인증 방식(f.auth)을 괄호에 붙입니다(빈 괄호는 뺌).
// 그런 메시지가 없으면 nil(원본의 null)입니다.
func startupOf(ms []atlas.Event) []any {
	var out []any
	for _, m := range ms {
		switch m.Name {
		case "ssl", "sslreply", "startup", "R", "K", "Z":
			step := m.Name
			if m.Name == "sslreply" || m.Name == "Z" {
				step += "(" + fmt.Sprint(m.Attrs[srcWireMac]["text"]) + ")"
			} else if a, ok := m.Attrs[srcWireMac]["f.auth"]; ok {
				step += "(" + fmt.Sprint(a) + ")"
			}
			out = append(out, strings.TrimSuffix(step, "()"))
		}
		if m.Name == "Z" {
			break
		}
	}
	return out
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
			errm = map[string]any{"code": v.text(m, srcWireMac, "f.code"), "message": text, "where": v.text(m, srcWireMac, "f.where")}
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

// window 메서드는 서버 창 하나를 되살립니다. 품은 것은 연결 차례로 가릅니다: 기록기의 sample은 카탈로그 스냅숏 잡기(수는 해석기 값), 이름이 check면
// 줄 판정, curcid가 있으면 스냅숏, 해석기의 기다림은 view.waits라 빼고, 나머지는 사건입니다. sql은 요청의 sql을 "; "로 이은 것이고,
// in·out은 해석기가 Spring 창에 채우지 않아 늘 null입니다(적재기가 null이 아니면 멈춤).
func (v *view) window(w atlas.Event, sql string) (map[string]any, error) {
	s := map[string]any{"sql": sql, "t0": micros(w.StartNS), "t1": micros(w.EndNS), "in": nil, "out": nil}
	events, snaps, checks := []any{}, []any{}, []any{}
	for _, id := range v.out[w.ID]["contains"] {
		m := v.r.Events[id]
		_, isSnap := v.attr(m, srcRec, "curcid")
		switch {
		case m.Source == srcCont:
		case m.Kind == atlas.Sample && m.Name == nameCatalog:
			v.copyAttrs(s, m, srcSpring, "catalogSnapshots")
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
			f, err := v.serverFields(m)
			if err != nil {
				return nil, err
			}
			if len(f) > 0 {
				ev["f"] = f
			}
			events = append(events, ev)
		}
	}
	s["events"], s["snapshots"], s["checks"] = events, snaps, checks
	return s, nil
}

// serverFields 메서드는 서버 사건의 나머지 칸 f를 "f.<칸>" attrs에서 되살립니다. hint·ereport의 at(부른 줄)은 스택(안쪽부터) 1번 칸의
// 코드 자리입니다(적재기가 같은지 확인함). 그 칸이 없으면 오류입니다.
func (v *view) serverFields(m atlas.Event) (map[string]any, error) {
	f := map[string]any{}
	for k, val := range m.Attrs[srcRec] {
		if name, ok := strings.CutPrefix(k, "f."); ok {
			f[name] = val
		}
	}
	if m.Name == "hint" || m.Name == "ereport" {
		st := load.StackOf(v.r.Places, m.Stack)
		if len(st) < 2 {
			return nil, fmt.Errorf("%s 사건 %d의 스택에 부른 줄이 없음", m.Name, m.ID)
		}
		f["at"] = st[len(st)-2].(map[string]any)["at"]
	}
	return f, nil
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
			c, err := v.call(k)
			if err != nil {
				return nil, err
			}
			lines = append(lines, c)
		case namePath:
			p, err := v.call(k)
			if err != nil {
				return nil, err
			}
			v.copyAttrs(p, k, srcRec, "arg", "ret")
			if kids := v.children[id]; len(kids) > 0 {
				inner := []any{}
				for _, j := range kids {
					c, err := v.call(v.r.Events[j])
					if err != nil {
						return nil, err
					}
					inner = append(inner, c)
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

// call 메서드는 줄 판정 안의 호출 하나를 되살립니다. fn은 장소(부른 함수), at은 스택 1번(부른 줄)입니다. 스택이 짧으면 오류입니다.
func (v *view) call(k atlas.Event) (map[string]any, error) {
	if len(k.Stack) < 2 {
		return nil, fmt.Errorf("줄 판정 안의 호출 %d의 스택에 부른 줄이 없음", k.ID)
	}
	return map[string]any{"fn": v.r.Places[k.Place].Name, "at": load.LocText(v.r.Places[k.Stack[1]].Loc)}, nil
}

// transactions 메서드는 트랜잭션을 되살립니다. first·last는 품은 요청의 seq, end는 끝 요청의 마지막 태그(없으면 빈 글), xids·clog는 품은 요청의
// 서버 창 사건, next는 연결한 메시지에서 셉니다(decode/spring.go 311~338).
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
		x := map[string]any{"conn": n, "first": seqs[0], "last": seqs[len(seqs)-1], "t0": micros(e.StartNS), "t1": micros(e.EndNS), "end": ""}
		if tags, _ := v.reqOut[order[len(order)-1]]["tags"].([]any); len(tags) > 0 {
			x["end"] = tags[len(tags)-1]
		}
		xids, clog := []any{}, []any{}
		for _, q := range order {
			for _, w := range v.out[q][nameServer] {
				level := any("1")
				for _, id := range v.out[w]["contains"] {
					m := v.r.Events[id]
					if m.Source != srcRec || m.Name == nameCheck {
						continue
					}
					f := m.Attrs[srcRec]
					switch m.Name {
					case "assign_xid":
						level = f["f.level"]
					case "xid":
						xids = append(xids, map[string]any{"xid": f["f.xid"], "level": level, "t": micros(m.StartNS)})
					case "commit_clog", "abort_clog":
						clog = append(clog, map[string]any{"kind": m.Name, "xid": f["f.xid"], "nsub": f["f.nsub"], "t": micros(m.StartNS)})
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

// captures 메서드는 관찰을 되살립니다. afterStep은 same 연결로 들어온 멈춤 자리의 n(막힘 관찰이면 그 after), span은 [시작, end 연결이
// 가리키는 해석기 순간], reads는 contains로 품은 해석기 구간(끝이 비면 시작과 같음), overlap은 연결한 메시지입니다.
func (v *view) captures(evs []atlas.Event) ([]any, error) {
	out := []any{}
	for _, e := range evs {
		end := v.out[e.ID]["end"]
		if len(end) != 1 {
			return nil, fmt.Errorf("관찰 %d의 end 연결 %d개", e.ID, len(end))
		}
		c := map[string]any{"t": micros(e.StartNS), "span": []any{micros(e.StartNS), micros(v.r.Events[end[0]].StartNS)}}
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
		for _, id := range v.out[e.ID]["contains"] {
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
//
// 구간의 depth는 해석기처럼 셉니다(decode/spring_jfr.go 325~341): 스레드마다 구간을 사건 차례(원본 spans의 차례, 시작 오름·끝 내림)로 보며,
// 시작보다 먼저 끝난 열린 호출을 닫고 남은 열린 호출의 수를 깊이로 씁니다.
func (v *view) java() (map[string]any, error) {
	spans, throws, parks := []any{}, []any{}, []any{}
	open := map[int][]int64{} // 스레드 장소 → 열린 호출의 끝(ns)
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
			s := map[string]any{"fn": m.Name, "desc": place.Name, "thread": th.Name, "t0": micros(e.StartNS), "t1": micros(e.EndNS), "caller": ""}
			if len(e.Stack) > 1 {
				s["caller"] = v.ref(e.Stack[1])
			}
			st := open[th.ID]
			for len(st) > 0 && st[len(st)-1] < *e.StartNS {
				st = st[:len(st)-1]
			}
			s["depth"] = len(st)
			open[th.ID] = append(st, *e.EndNS)
			v.copyAttrs(s, e, srcSpring, "twice")
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
