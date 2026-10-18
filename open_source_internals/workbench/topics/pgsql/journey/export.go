package journey

import (
	"encoding/json"
	"fmt"
	"strconv"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// view 값은 되돌리기에 쓰는 run의 색인입니다. 연결은 종류마다 출발·도착 사건으로 찾습니다(상대 사건의 id 차례, load.LinkIndex).
type view struct {
	r        *atlas.Run
	out, in  map[int]map[string][]int
	children map[int][]int
}

// newView 함수는 run에서 연결과 부모의 색인을 만듭니다. 차례는 id 차례(기록 차례)입니다.
func newView(r *atlas.Run) *view {
	v := &view{r: r, children: map[int][]int{}}
	v.out, v.in = load.LinkIndex(r)
	for _, e := range r.Events {
		if e.Parent != nil {
			v.children[*e.Parent] = append(v.children[*e.Parent], e.ID)
		}
	}
	return v
}

// ExportC4 함수는 Import가 만든 run에서 concept-4의 질의 하나({key, label, typed, x, data, procs, procsGuess})를 다시 만듭니다.
//
// run 밖의 것은 읽지 않습니다. 되살리는 데 필요한 사건·장소·값이 없으면 오류를 돌려줍니다.
func ExportC4(r *atlas.Run) (out map[string]any, err error) {
	defer recoverExport(&out, &err)
	v := newView(r)
	q := map[string]any{}
	for _, k := range []string{"key", "label", "typed", "x"} {
		val, ok := r.Attrs[srcTopic][k]
		if !ok {
			return nil, fmt.Errorf("run attrs %s.%s 없음", srcTopic, k)
		}
		q[k] = val
	}
	data, err := v.data(c4)
	if err != nil {
		return nil, err
	}
	q["data"] = data

	// 역할표는 역할 장소와 그 아래 프로세스 장소에서, procsGuess는 역할 장소의 출처 성격에서 되살립니다.
	procs := map[string]any{}
	guess := false
	for _, p := range r.Places {
		if p.Kind != "role" {
			continue
		}
		src, _ := r.SourceByID(p.Source)
		guess = src.Nature == atlas.Inferred
		pids := []any{}
		for _, c := range r.Places {
			if c.Parent != nil && *c.Parent == p.ID && c.Kind == "process" {
				pids = append(pids, json.Number(c.Name))
			}
		}
		procs[p.Name] = pids
	}
	q["procs"] = procs
	q["procsGuess"] = guess
	return q, nil
}

// ExportC3 함수는 같은 run에서 concept-3의 해석 결과(질의 하나)를 다시 만듭니다. concept-3 출처가 없으면 두 번째 값이 false입니다.
func ExportC3(r *atlas.Run) (data map[string]any, found bool, err error) {
	if _, ok := r.SourceByID(c3.journey); !ok {
		return nil, false, nil
	}
	defer recoverExport(&data, &err)
	data, err = newView(r).data(c3)
	return data, true, err
}

// recoverExport 함수는 적재기가 만들지 않는 꼴의 run(빈 시각 등)에서 되살리다 멈추면 panic 대신 오류를 돌려주게 합니다(앞선 적재기 코드 검토 지적 10, Spring과 같음).
func recoverExport(out *map[string]any, err *error) {
	if p := recover(); p != nil {
		*out, *err = nil, fmt.Errorf("되살리다 멈춤(적재기가 만들지 않는 꼴의 run): %v", p)
	}
}

// data 메서드는 한 판의 해석 결과({endUs, hops, pid, query, src, stages}와 concept-4의 kernelPid·ksrc)를 되살립니다.
func (v *view) data(d decoder) (map[string]any, error) {
	r := v.r
	out := map[string]any{"query": r.Attrs[d.rec][d.queryKey], "src": r.Attrs[d.src][d.srcKey]}

	// 질의 구간이 끝 시각과 백엔드 pid를 줍니다.
	backend := -1
	for _, e := range r.Events {
		if e.Source == d.journey && e.Kind == atlas.Span && e.Name == "query" {
			backend = e.Place
			out["endUs"] = atlas.NanosToMicros(*e.EndNS)
			out["pid"] = json.Number(r.Places[e.Place].Name)
		}
	}
	if backend < 0 {
		return nil, fmt.Errorf("%s의 질의 구간이 없음", d.journey)
	}
	if d == c4 {
		for _, p := range r.Places {
			if p.Kind == "process" && p.Source == srcFtrace {
				out["kernelPid"] = json.Number(p.Name)
				out["ksrc"] = r.Attrs[srcKsrc]["data.ksrc"]
			}
		}
	}

	// 이동은 기록기 출처의 사건 가운데 contains 연결의 도착(메시지)이 아닌 것입니다.
	hops := []any{}
	stages := []any{}
	for _, e := range r.Events {
		switch {
		case e.Source == d.rec && len(v.in[e.ID]["contains"]) == 0:
			h, err := v.hop(d, e, backend)
			if err != nil {
				return nil, err
			}
			hops = append(hops, h)
		case e.Source == d.journey && e.Kind == atlas.Instant:
			stages = append(stages, v.stage(e))
		}
	}
	out["hops"] = hops
	out["stages"] = stages
	return out, nil
}

// hop 메서드는 이동 하나를 되살립니다. proc는 이동의 프로세스가 질의의 백엔드인지로 다시 셉니다.
func (v *view) hop(d decoder, e atlas.Event, backend int) (map[string]any, error) {
	r := v.r
	place := r.Places[e.Place]
	proc, site := e.Place, ""
	if place.Kind == "site" {
		proc, site = *place.Parent, load.LocText(place.Loc)
	}
	h := map[string]any{
		"n": e.Attrs[d.journey]["n"], "stage": e.Attrs[d.journey]["stage"],
		"pid": json.Number(r.Places[proc].Name), "proc": "other",
		"kind": e.Name, "count": json.Number(strconv.Itoa(e.Count)),
		"t0": atlas.NanosToMicros(*e.StartNS), "t1": atlas.NanosToMicros(*e.EndNS), "site": site,
	}
	if proc == backend {
		h["proc"] = "backend"
	}
	if val, ok := e.Attrs[d.journey]["catalog"]; ok {
		h["catalog"] = val
	}
	if val, ok := e.Attrs[d.rec]["bytes"]; ok {
		h["bytes"] = val
	}
	if val, ok := e.Attrs[d.journey]["bytes"]; ok {
		h["bytes"] = val
	}

	// 출발·도착이 파일 장소면 그 부모 칸이 from·to이고 파일 이름이 files입니다.
	for _, end := range []struct {
		key string
		id  *int
	}{{"from", e.From}, {"to", e.Target}} {
		if end.id == nil {
			return nil, fmt.Errorf("사건 %d의 %s 없음", e.ID, end.key)
		}
		p := r.Places[*end.id]
		if p.Kind == "file" {
			h[end.key] = r.Places[*p.Parent].Name
			h["files"] = []any{p.Name}
		} else {
			h[end.key] = p.Name
		}
	}

	// 스택은 0번(이동의 장소)을 빼고 바깥 함수부터 되돌립니다. 스택이 없던 이동은 null입니다.
	if e.Stack == nil {
		h["stack"] = nil
	} else {
		stack := []any{}
		for k := len(e.Stack) - 1; k >= 1; k-- {
			f := r.Places[e.Stack[k]]
			stack = append(stack, map[string]any{"fn": f.Name, "at": load.LocText(f.Loc)})
		}
		h["stack"] = stack
	}

	// 메시지: 수신·송신은 contains로 품은 메시지이고, 버퍼 쓰기는 copy 연결이 가리키는 송신 메시지의 사본입니다.
	// cont는 들어오는 continues 연결, cut은 나가는 continues 연결이 있는 것입니다(사본은 그 송신 메시지의 cut을 따름).
	ms := v.out[e.ID]["contains"]
	if len(ms) == 0 {
		ms = v.out[e.ID]["copy"]
	}
	if len(ms) > 0 {
		msgs := []any{}
		for _, id := range ms {
			m := r.Events[id]
			src := d.rec
			if _, ok := m.Attrs[d.journey]["len"]; ok {
				src = d.journey
			}
			msg := map[string]any{"type": m.Name, "len": m.Attrs[src]["len"]}
			if t, ok := m.Attrs[src]["text"]; ok {
				msg["text"] = t
			}
			if len(v.out[id]["continues"]) > 0 {
				msg["cut"] = true
			}
			if len(v.in[id]["continues"]) > 0 && len(v.out[e.ID]["contains"]) > 0 {
				msg["cont"] = true
			}
			msgs = append(msgs, msg)
		}
		h["msgs"] = msgs
	}

	// 시스템 콜 길이와 커널 마디는 syscall·kernel 연결을 따라 찾습니다.
	if sys := v.out[e.ID]["syscall"]; len(sys) == 1 {
		s := r.Events[sys[0]]
		h["sysUs"] = atlas.NanosToMicros(*s.DurNS)
		if roots := v.out[s.ID]["kernel"]; len(roots) > 0 {
			kernel := []any{}
			for _, root := range roots {
				kernel = append(kernel, v.kernelNode(root))
			}
			h["kernel"] = kernel
		}
	}
	return h, nil
}

// kernelNode 메서드는 커널 마디 하나와 그 아래를 되살립니다. at은 장소가 프레임일 때만 있습니다.
func (v *view) kernelNode(id int) map[string]any {
	e := v.r.Events[id]
	p := v.r.Places[e.Place]
	n := map[string]any{"fn": p.Name, "us": atlas.NanosToMicros(*e.DurNS)}
	if p.Kind == "frame" {
		n["at"] = load.LocText(p.Loc)
	}
	for _, src := range []string{"journey", srcKGuess} {
		for k, val := range e.Attrs[src] {
			n[k] = val
		}
	}
	if kids := v.children[id]; len(kids) > 0 {
		list := []any{}
		for _, c := range kids {
			list = append(list, v.kernelNode(c))
		}
		n["kids"] = list
	}
	return n
}

// stage 메서드는 단계 하나를 되살립니다. 스택 1번이 프레임이면 caller와 at, 코드 자리면 at만 있습니다.
func (v *view) stage(e atlas.Event) map[string]any {
	s := map[string]any{"name": e.Name, "t0": atlas.NanosToMicros(*e.StartNS)}
	if len(e.Stack) == 2 {
		s["fn"] = v.r.Places[e.Place].Name
		outer := v.r.Places[e.Stack[1]]
		if outer.Kind == "frame" {
			s["caller"] = outer.Name
		}
		s["at"] = load.LocText(outer.Loc)
	}
	return s
}
