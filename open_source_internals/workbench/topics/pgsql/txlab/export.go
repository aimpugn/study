package txlab

import (
	"encoding/json"
	"fmt"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// view 값은 되돌리기에 쓰는 run의 색인입니다. 연결은 종류마다 출발·도착 사건으로 찾고(상대 사건의 id 차례, load.LinkIndex), 자식은 id 차례입니다.
type view struct {
	r        *atlas.Run
	out, in  map[int]map[string][]int
	children map[int][]int
	owner    map[int]string // 백엔드 프로세스 장소 → 세션 이름
	pidOf    map[string]int // 세션 이름 → 백엔드 프로세스 장소
}

// newView 함수는 run에서 연결과 부모의 색인을 만들고, 단계 contains 요청으로 세션과 백엔드의 짝을 셉니다.
// 세션 하나에 백엔드가 둘이거나 백엔드 하나에 세션이 둘이면 오류입니다.
func newView(r *atlas.Run) (*view, error) {
	v := &view{r: r, children: map[int][]int{}, owner: map[int]string{}, pidOf: map[string]int{}}
	v.out, v.in = load.LinkIndex(r)
	for _, e := range r.Events {
		if e.Parent != nil {
			v.children[*e.Parent] = append(v.children[*e.Parent], e.ID)
		}
	}
	for _, e := range r.Events {
		if e.Source != srcTxlab || e.Name != nameStep {
			continue
		}
		name := r.Places[e.Place].Name
		for _, q := range v.requests(e.ID) {
			proc := r.Events[q].Place
			if p, ok := v.pidOf[name]; ok && p != proc {
				return nil, fmt.Errorf("세션 %s의 백엔드가 둘(%d, %d)", name, p, proc)
			}
			if s, ok := v.owner[proc]; ok && s != name {
				return nil, fmt.Errorf("백엔드 %d의 세션이 둘(%s, %s)", proc, s, name)
			}
			v.pidOf[name], v.owner[proc] = proc, name
		}
	}
	return v, nil
}

// requests 메서드는 단계가 contains로 품은 요청(기록기의 exec_simple_query 구간)을 연결 차례로 돌려줍니다. 힌트 변화는 뺍니다.
func (v *view) requests(step int) []int {
	var out []int
	for _, id := range v.out[step]["contains"] {
		if e := v.r.Events[id]; e.Source == srcRec && e.Name == nameRequest {
			out = append(out, id)
		}
	}
	return out
}

// attr 메서드는 사건의 출처 src가 낸 값 key를 돌려줍니다.
func (v *view) attr(e atlas.Event, src, key string) (any, bool) {
	val, ok := e.Attrs[src][key]
	return val, ok
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

// processOf 메서드는 장소에서 부모를 따라 올라가 프로세스 장소의 pid를 돌려줍니다.
func (v *view) processOf(place int) (json.Number, error) {
	for p := place; ; {
		pl := v.r.Places[p]
		if pl.Kind == "process" {
			return json.Number(pl.Name), nil
		}
		if pl.Parent == nil {
			return "", fmt.Errorf("장소 %d 위에 프로세스가 없음", place)
		}
		p = *pl.Parent
	}
}

// stackOf 메서드는 스택(0번이 가장 안쪽)을 원본의 꼴(바깥 함수부터 {fn, at})로 되돌립니다.
func (v *view) stackOf(stack []int) []any { return load.StackOf(v.r.Places, stack) }

// Export 함수는 Import가 만든 run에서 기록 한 벌(RUNS의 값 하나)을 다시 만듭니다.
//
// run 밖의 것은 읽지 않습니다. 사건은 출처와 이름, 연결로 가르고, 가를 수 없는 사건이나 되살리는 데 필요한 값이 없으면 오류를 돌려줍니다.
func Export(r *atlas.Run) (out map[string]any, err error) {
	// 적재기가 만들지 않는 꼴의 run(빈 시각, 다른 타입의 값)에서 멈추면 panic 대신 오류로 돌려줍니다(앞선 적재기 코드 검토 지적 10, Spring과 같음).
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
	for src, keys := range map[string][]string{srcTxlab: {"title", "scenario", "tables"}, srcTx: {"tupleAddrs", "joins", "clock", "siteRepair"}, srcSrc: {"src"}} {
		for _, k := range keys {
			val, ok := r.Attrs[src][k]
			if !ok {
				return nil, fmt.Errorf("run attrs %s.%s 없음", src, k)
			}
			out[k] = val
		}
	}

	// 세션은 장소 차례이고 pid는 단계가 품은 요청의 장소로 다시 셉니다.
	sessions := []any{}
	for _, p := range r.Places {
		if p.Kind != "session" {
			continue
		}
		proc, ok := v.pidOf[p.Name]
		if !ok {
			return nil, fmt.Errorf("세션 %s의 백엔드를 셀 단계가 없음", p.Name)
		}
		sessions = append(sessions, map[string]any{"name": p.Name, "pid": json.Number(r.Places[proc].Name)})
	}
	out["sessions"] = sessions

	// 사건은 id 차례로 보고 목록마다 모읍니다. 자식과 요청이 품은 것은 그 부모·요청이 되살립니다.
	lists := map[string][]any{}
	for _, e := range r.Events {
		var item map[string]any
		var list string
		switch {
		case e.Parent != nil || len(v.in[e.ID]["contains"]) > 0:
			continue
		case e.Source == srcTxlab && e.Name == nameBlocked:
			// 막힘 관찰은 단계가 blocked 연결로 되살립니다.
		case e.Source == srcTxlab && e.Name == nameStep:
			list = "steps"
			item, err = v.step(e)
		case e.Source == srcTxlab && e.Name == nameCapture:
			list = "captures"
			item, err = v.capture(e)
		case e.Source == srcTxlab && e.Name == nameKill:
			list = "kills"
			item, err = v.kill(e)
		case e.Source == srcTxlab && e.Name == namePause:
			list = "pauses"
			item, err = v.pause(e)
		case e.Source == srcServer && e.Name == nameProcess:
			list = "procs"
			item, err = v.proc(e)
		case e.Source == srcTx && e.Name == nameBystand:
			list = "bystanders"
			item = map[string]any{"pid": json.Number(r.Places[e.Place].Name)}
			v.copyAttrs(item, e, srcTx, "firstSql", "hints", "outsideHints", "queries", "watchedChecks", "watchedHints")
		case e.Source == srcSpan:
			list = "spans"
			item, err = v.span(e)
		case e.Source == srcRec || e.Source == srcShift || e.Source == srcStrace || e.Source == srcLog:
			list = "events"
			item, err = v.serverEvent(e)
		default:
			return nil, fmt.Errorf("사건 %d(%s, %s)를 어느 목록에도 넣을 수 없음", e.ID, e.Source, e.Name)
		}
		if err != nil {
			return nil, err
		}
		lists[list] = append(lists[list], item)
	}
	// 힌트 변화는 단계가 contains로 품어 위에서 건너뛰었으므로 따로 id 차례로 모읍니다.
	for _, e := range r.Events {
		if e.Source == srcHint {
			item, err := v.hint(e)
			if err != nil {
				return nil, err
			}
			lists["hintChanges"] = append(lists["hintChanges"], item)
		}
	}
	for _, k := range []string{"steps", "captures", "kills", "pauses", "bystanders", "hintChanges"} {
		out[k] = orEmpty(lists[k])
	}
	out["server"] = map[string]any{"procs": orEmpty(lists["procs"]), "events": orEmpty(lists["events"]), "spans": orEmpty(lists["spans"])}
	return out, nil
}

// orEmpty 함수는 목록이 nil이면 빈 목록을 돌려줍니다. 원본은 빈 목록을 []로 적습니다.
func orEmpty(l []any) []any {
	if l == nil {
		return []any{}
	}
	return l
}

// step 메서드는 단계 하나와 그 요청을 되살립니다.
func (v *view) step(e atlas.Event) (map[string]any, error) {
	s := map[string]any{"session": v.r.Places[e.Place].Name, "t0": micros(e.StartNS), "t1": micros(e.EndNS)}
	v.copyAttrs(s, e, srcTxlab, "n", "sql", "output")
	for _, b := range v.out[e.ID]["blocked"] {
		s["blocked"] = micros(v.r.Events[b].StartNS)
	}
	v.copyAttrs(s, e, srcTx, "check")
	reqs := []any{}
	for _, id := range v.requests(e.ID) {
		q, err := v.request(v.r.Events[id], *e.EndNS)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, q)
	}
	s["requests"] = reqs
	return s, nil
}

// request 메서드는 요청 하나를 되살립니다. 끝이 비면 오류로 끝난 요청이라 aborted와 단계의 끝을 씁니다.
//
// 요청이 품은 것은 연결 차례대로 가릅니다: from·target이 있으면 메시지, 이름이 check면 줄 판정, curcid가 있으면 스냅숏, 나머지는 사건입니다.
func (v *view) request(e atlas.Event, stepEnd int64) (map[string]any, error) {
	q := map[string]any{"t0": micros(e.StartNS)}
	v.copyAttrs(q, e, srcRec, "sql")
	v.copyAttrs(q, e, srcTx, "catalogSnapshots")
	if e.EndNS == nil {
		q["t1"], q["aborted"] = atlas.NanosToMicros(stepEnd), true
	} else {
		q["t1"] = micros(e.EndNS)
	}
	in, events, snaps, checks := []any{}, []any{}, []any{}, []any{}
	var out []any
	for _, id := range v.out[e.ID]["contains"] {
		m := v.r.Events[id]
		_, isSnap := v.attr(m, srcRec, "curcid")
		_, isMsg := v.attr(m, srcStrace, "len")
		switch {
		case isMsg && m.From != nil:
			in = append(in, v.message(m))
		case isMsg && m.Target != nil:
			out = append(out, v.message(m))
		case m.Source == srcRec && m.Name == nameCheck:
			c, err := v.check(m)
			if err != nil {
				return nil, err
			}
			checks = append(checks, c)
		case isSnap:
			s := map[string]any{"kind": m.Name, "t": micros(m.StartNS), "stack": v.stackOf(m.Stack)}
			v.copyAttrs(s, m, srcRec, "curcid", "xip", "xmax", "xmin")
			snaps = append(snaps, s)
		default:
			events = append(events, v.requestEvent(m))
		}
	}
	q["in"], q["events"], q["snapshots"], q["checks"] = in, events, snaps, checks
	if out == nil {
		q["out"] = nil
	} else {
		q["out"] = out
	}
	return q, nil
}

// message 메서드는 연결 위 메시지 하나를 되살립니다.
func (v *view) message(m atlas.Event) map[string]any {
	msg := map[string]any{"type": m.Name}
	v.copyAttrs(msg, m, srcStrace, "len", "text")
	return msg
}

// requestEvent 메서드는 요청 창 안의 사건 하나를 되살립니다. 스택과 끝, f는 있을 때만 씁니다.
func (v *view) requestEvent(m atlas.Event) map[string]any {
	ev := map[string]any{"kind": m.Name, "t": micros(m.StartNS)}
	if m.EndNS != nil {
		ev["t1"] = micros(m.EndNS)
	}
	if m.Stack != nil {
		ev["stack"] = v.stackOf(m.Stack)
	}
	f := v.fields(m)
	if f == nil && (m.Name == "sleep_syscall" || m.Name == "io" || m.Name == "hint" || m.Name == "ereport") {
		// 다시 세는 칸만 가진 f(sleep_syscall의 durUs, ereport의 at)는 attrs에 남는 칸이 없지만 그릇은 원본에 있습니다.
		f = map[string]any{}
	}
	if f != nil {
		// 다시 세는 칸(적재기의 requestFields): 부른 줄은 스택 1번 칸, 깨운·읽은 쪽 프로세스는 target, 쓴 쪽은 사건의 백엔드, durUs는 (t1 − t)/1000입니다.
		if m.Name == "hint" || m.Name == "ereport" {
			if len(m.Stack) >= 2 {
				f["at"] = load.LocText(v.r.Places[m.Stack[1]].Loc)
			}
		}
		if m.Target != nil {
			switch m.Name {
			case "wake":
				f["who"] = v.r.Places[*m.Target].Name
			case "rw_conflict":
				f["rpid"] = v.r.Places[*m.Target].Name
				if pid, err := v.processOf(m.Place); err == nil {
					f["wpid"] = fmt.Sprint(pid)
				}
			}
		}
		v.durUs(f, m)
		ev["f"] = f
	}
	return ev
}

// fields 메서드는 사건의 attrs에서 "f.<칸>" 열쇠를 모아 나머지 칸 f를 되살립니다. 그런 열쇠가 없으면 nil입니다.
func (v *view) fields(m atlas.Event) map[string]any {
	var f map[string]any
	for k, val := range m.Attrs[m.Source] {
		if name, ok := strings.CutPrefix(k, "f."); ok {
			if f == nil {
				f = map[string]any{}
			}
			f[name] = val
		}
	}
	return f
}

// durUs 메서드는 시스템 콜 사건의 durUs를 (t1 − t) ns를 1000으로 나눈 몫으로 다시 셉니다. 해석기는 call이나 durUs가 있는 종류에만 적었습니다.
func (v *view) durUs(f map[string]any, m atlas.Event) {
	if m.Name != "io" && m.Name != "sleep_syscall" && m.Name != "wal_io" {
		return
	}
	f["durUs"] = fmt.Sprint((*m.EndNS - *m.StartNS) / 1000)
}

// check 메서드는 줄 판정 하나와 그 자식(줄, 판정 경로와 경로 안의 호출)을 되살립니다. 경로 사건이 없으면 path는 null입니다.
func (v *view) check(m atlas.Event) (map[string]any, error) {
	c := map[string]any{"t": micros(m.StartNS), "stack": v.stackOf(m.Stack)}
	v.copyAttrs(c, m, srcRec, "ctid", "tup", "table", "mask", "xmin", "xmax", "snapXmin", "snapXmax")
	v.copyAttrs(c, m, srcTx, "visible")
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

// stepNumber 메서드는 연결 kind로 사건 id에서 나가는 단계의 번호 n을 돌려줍니다. 그런 연결이 없으면 ok가 거짓입니다.
func (v *view) stepNumber(id int, kind string) (any, bool) {
	for _, dst := range v.out[id][kind] {
		if d := v.r.Events[dst]; d.Source == srcTxlab && d.Name == nameStep {
			return v.attr(d, srcTxlab, "n")
		}
	}
	return nil, false
}

// capture 메서드는 관찰 하나를 되살립니다. after 연결이 단계로 가면 afterStep, kill로 가면 afterKill이고, 단계로 가는 것이 없으면 0입니다.
func (v *view) capture(e atlas.Event) (map[string]any, error) {
	c := map[string]any{"t": micros(e.StartNS), "afterStep": json.Number("0")}
	if n, ok := v.stepNumber(e.ID, "after"); ok {
		c["afterStep"] = n
	}
	for _, dst := range v.out[e.ID]["after"] {
		if d := v.r.Events[dst]; d.Source == srcTxlab && d.Name == nameKill {
			c["afterKill"] = true
		}
	}
	v.copyAttrs(c, e, srcTxlab, "activity", "locks", "pages", "running", "xacts", "wal")
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
	return c, nil
}

// kill 메서드는 kill 하나를 되살립니다. session은 target 프로세스의 세션, walWrittenT는 wal 연결이 가리키는 쓰기의 끝입니다.
func (v *view) kill(e atlas.Event) (map[string]any, error) {
	if e.Target == nil || e.EndNS == nil {
		return nil, fmt.Errorf("kill %d의 target이나 끝(tReady) 없음", e.ID)
	}
	k := map[string]any{"pid": json.Number(v.r.Places[*e.Target].Name), "session": v.owner[*e.Target], "t": micros(e.StartNS), "tReady": micros(e.EndNS)}
	n, ok := v.stepNumber(e.ID, "before")
	if !ok {
		return nil, fmt.Errorf("kill %d의 before 연결 없음", e.ID)
	}
	k["beforeStep"] = n
	v.copyAttrs(k, e, srcTxlab, "walWritten", "walSynced")
	if w := v.out[e.ID]["wal"]; len(w) == 1 {
		end := v.r.Events[w[0]].EndNS
		if end == nil {
			return nil, fmt.Errorf("kill %d의 wal 연결이 끝 없는 사건 %d를 가리킴", e.ID, w[0])
		}
		k["walWrittenT"] = micros(end)
	}
	return k, nil
}

// pause 메서드는 쉼 하나를 되살립니다.
func (v *view) pause(e atlas.Event) (map[string]any, error) {
	n, ok := v.stepNumber(e.ID, "before")
	if !ok {
		return nil, fmt.Errorf("쉼 %d의 before 연결 없음", e.ID)
	}
	return map[string]any{"beforeStep": n, "t": micros(e.StartNS), "t1": micros(e.EndNS)}, nil
}

// proc 메서드는 서버 프로세스 하나를 수명 구간에서 되살립니다. 역할은 장소의 부모, session은 그 프로세스를 백엔드로 쓴 세션입니다.
func (v *view) proc(e atlas.Event) (map[string]any, error) {
	pl := v.r.Places[e.Place]
	if pl.Parent == nil {
		return nil, fmt.Errorf("프로세스 %s의 역할 장소 없음", pl.Name)
	}
	p := map[string]any{"pid": json.Number(pl.Name), "role": v.r.Places[*pl.Parent].Name, "t0": micros(e.StartNS), "t1": micros(e.EndNS)}
	v.copyAttrs(p, e, srcServer, "exit")
	if s, ok := v.owner[e.Place]; ok {
		p["session"] = s
	}
	return p, nil
}

// span 메서드는 서버 함수 구간 하나를 되살립니다. at은 장소(함수 입구 프레임)의 위치입니다.
func (v *view) span(e atlas.Event) (map[string]any, error) {
	pid, err := v.processOf(e.Place)
	if err != nil {
		return nil, err
	}
	return map[string]any{"at": load.LocText(v.r.Places[e.Place].Loc), "fn": e.Name, "pid": pid, "t0": micros(e.StartNS), "t1": micros(e.EndNS)}, nil
}

// serverEvent 메서드는 서버 사건 하나를 되살립니다. ftrace 시각을 옮겨 쓴 출처의 사건이면 f에 clock을 다시 붙입니다.
func (v *view) serverEvent(e atlas.Event) (map[string]any, error) {
	pid, err := v.processOf(e.Place)
	if err != nil {
		return nil, err
	}
	ev := map[string]any{"kind": e.Name, "pid": pid, "t": micros(e.StartNS)}
	if e.EndNS != nil {
		ev["t1"] = micros(e.EndNS)
	}
	if e.Stack != nil {
		ev["stack"] = v.stackOf(e.Stack)
	}
	fields := v.fields(e)
	if fields == nil {
		fields = map[string]any{}
	}
	if e.Target != nil {
		fields["to"] = v.r.Places[*e.Target].Name
	}
	v.durUs(fields, e)
	if src, _ := v.r.SourceByID(e.Source); src.Clock == clockShift {
		fields["clock"] = "ftrace"
	}
	ev["f"] = fields
	return ev, nil
}

// hint 메서드는 힌트 변화 하나를 되살립니다. by는 장소(코드 자리)의 부모 프로세스를 백엔드로 쓴 세션, step은 contains로 품은 단계입니다.
func (v *view) hint(e atlas.Event) (map[string]any, error) {
	site := v.r.Places[e.Place]
	if site.Parent == nil {
		return nil, fmt.Errorf("힌트 %d의 장소에 부모 없음", e.ID)
	}
	by, ok := v.owner[*site.Parent]
	if !ok {
		return nil, fmt.Errorf("힌트 %d의 프로세스에 세션 없음", e.ID)
	}
	h := map[string]any{"at": load.LocText(site.Loc), "by": by, "t": micros(e.StartNS)}
	for _, src := range v.in[e.ID]["contains"] {
		if d := v.r.Events[src]; d.Source == srcTxlab && d.Name == nameStep {
			h["step"], _ = v.attr(d, srcTxlab, "n")
		}
	}
	if _, ok := h["step"]; !ok {
		return nil, fmt.Errorf("힌트 %d를 품은 단계 없음", e.ID)
	}
	v.copyAttrs(h, e, srcHint, "bit", "ctid", "table", "via")
	return h, nil
}
