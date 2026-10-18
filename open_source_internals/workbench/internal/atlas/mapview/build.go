package mapview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
)

// Build 함수는 Store의 run 가운데 자료 파일의 장면 규칙에 드는 것을 모두 읽어 보기 모형을 만듭니다.
//
// specDir는 자료 파일이 있는 폴더이고 고정 설명의 파일을 여기서 찾습니다.
// 장면에 드는 run이 하나도 없거나, 칸의 차례를 정하는 사건이 한 시계로 묶이지 않거나, 경로의 사건이 띠 규칙에 맞지 않으면 오류를 돌려줍니다.
func Build(st Store, spec *Spec, specDir string) (*View, error) {
	v := &View{Title: spec.Title, Zones: spec.Zones}
	ground := groundIndex(spec)
	for i, b := range spec.Bands {
		v.Bands = append(v.Bands, BandView{ID: b.ID, Name: b.Name, Zoom: b.Zoom, Zone: b.Zone, Ground: b.Ground, Below: i > ground, Under: b.Under})
	}
	fixed, err := loadFixed(spec, specDir)
	if err != nil {
		return nil, err
	}
	v.Fixed = fixed

	ids, err := st.List()
	if err != nil {
		return nil, err
	}
	// 장면 규칙에 드는 run마다 경로를 만들고, 정체 값이 같은 경로를 한 무리로 모읍니다.
	groupOf := map[string]int{}
	for _, id := range ids {
		sc := spec.sceneOf(id)
		if sc == nil {
			continue
		}
		r, err := st.Load(id)
		if err != nil {
			return nil, err
		}
		routes, err := buildRoutes(spec, sc, r)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		for _, rt := range routes {
			key := sc.ID + "\x00" + rt.identity
			gi, ok := groupOf[key]
			if !ok {
				gi = len(v.Groups)
				groupOf[key] = gi
				v.Groups = append(v.Groups, Group{ID: fmt.Sprintf("g%d", gi), Scene: sc.ID, Identity: rt.identity, Label: rt.Name, Typed: rt.Typed})
			}
			rt.Group = gi
			v.Groups[gi].Routes = append(v.Groups[gi].Routes, len(v.Routes))
			v.Routes = append(v.Routes, rt.Route)
		}
	}
	if len(v.Routes) == 0 {
		return nil, fmt.Errorf("장면 규칙에 드는 run이 없음")
	}
	for i, rt := range v.Routes {
		if spec.Default != "" && strings.HasPrefix(rt.Run, spec.Default) {
			v.Default = i
			break
		}
	}
	// 같은 무리의 후보끼리 갈리는 곳을 셉니다.
	for _, g := range v.Groups {
		for _, a := range g.Routes {
			for _, b := range g.Routes {
				if a != b {
					v.Routes[a].Divergence = append(v.Routes[a].Divergence, diverge(&v.Routes[a], &v.Routes[b], b))
				}
			}
		}
	}
	return v, nil
}

// groundIndex 함수는 지면 띠의 차례 번호입니다(check가 하나임을 보장).
func groundIndex(spec *Spec) int {
	for i, b := range spec.Bands {
		if b.Ground {
			return i
		}
	}
	return -1
}

// loadFixed 함수는 자료 파일이 가리키는 고정 설명을 읽습니다. 가리킨 열쇠가 없으면 오류입니다.
func loadFixed(spec *Spec, dir string) (FixedView, error) {
	out := FixedView{Bands: map[string]string{}, Names: map[string]string{}}
	var files []string
	for _, f := range spec.Fixed {
		raw, err := os.ReadFile(filepath.Join(dir, f.File))
		if err != nil {
			return out, err
		}
		var cur any
		if err := json.Unmarshal(raw, &cur); err != nil {
			return out, fmt.Errorf("%s: %w", f.File, err)
		}
		for _, k := range f.Path {
			m, ok := cur.(map[string]any)
			if !ok {
				return out, fmt.Errorf("%s: %v에서 %q를 찾을 수 없음", f.File, f.Path, k)
			}
			if cur, ok = m[k]; !ok {
				return out, fmt.Errorf("%s: %v에서 %q가 없음", f.File, f.Path, k)
			}
		}
		switch f.Kind {
		case "bands":
			// 열쇠가 띠 이름인 글 묶음입니다. 띠 이름으로 띠 id를 찾습니다.
			m, _ := cur.(map[string]any)
			for _, b := range spec.Bands {
				if t, ok := m[b.Name].(string); ok {
					out.Bands[b.ID] = t
				}
			}
		case "names":
			// [[장소 이름, 글], …] 꼴입니다.
			list, _ := cur.([]any)
			for _, it := range list {
				if pair, _ := it.([]any); len(pair) == 2 {
					out.Names[fmt.Sprint(pair[0])] = fmt.Sprint(pair[1])
				}
			}
		default:
			if t, ok := cur.(string); ok {
				out.Bands[f.Key] = t
			}
		}
		if !contains(files, f.File) {
			files = append(files, f.File)
		}
	}
	out.Source = strings.Join(files, ", ")
	return out, nil
}

// built 값은 경로 하나와 그 경로의 정체 값입니다.
type built struct {
	Route
	identity string
}

// index 값은 run 하나를 빨리 찾기 위한 색인입니다. band는 사건마다 띠를 한 번만 셈해 둔 것입니다.
type index struct {
	spec   *Spec
	r      *atlas.Run
	events map[int]*atlas.Event
	places map[int]*atlas.Place
	nature map[string]string
	clock  map[string]string
	kids   map[int][]int
	out    map[int][]atlas.Link
	in     map[int][]atlas.Link
	band   map[int]banded
}

// banded 값은 사건 하나의 띠와, 그 띠를 정한 attrs 열쇠의 출처 성격입니다(attrs로 정하지 않았으면 빈 글).
type banded struct {
	band, nature string
}

// newIndex 함수는 run의 사건·장소·연결·출처를 id로 찾게 묶습니다. 연결은 상대 사건의 id 차례입니다(형식 규칙 7).
func newIndex(spec *Spec, r *atlas.Run) *index {
	x := &index{spec: spec, r: r, events: map[int]*atlas.Event{}, places: map[int]*atlas.Place{}, nature: map[string]string{}, clock: map[string]string{},
		kids: map[int][]int{}, out: map[int][]atlas.Link{}, in: map[int][]atlas.Link{}, band: map[int]banded{}}
	for i := range r.Events {
		e := &r.Events[i]
		x.events[e.ID] = e
		if e.Parent != nil {
			x.kids[*e.Parent] = append(x.kids[*e.Parent], e.ID)
		}
	}
	for i := range r.Places {
		x.places[r.Places[i].ID] = &r.Places[i]
	}
	for _, s := range r.Sources {
		x.nature[s.ID] = s.Nature
		x.clock[s.ID] = s.Clock
	}
	for _, l := range r.Links {
		x.out[l.Src] = append(x.out[l.Src], l)
		x.in[l.Dst] = append(x.in[l.Dst], l)
	}
	for k := range x.out {
		sort.Slice(x.out[k], func(i, j int) bool { return x.out[k][i].Dst < x.out[k][j].Dst })
	}
	for k := range x.in {
		sort.Slice(x.in[k], func(i, j int) bool { return x.in[k][i].Src < x.in[k][j].Src })
	}
	for k := range x.kids {
		sort.Ints(x.kids[k])
	}
	return x
}

// weaker 함수는 두 출처 성격 가운데 약한 쪽을 돌려줍니다. 약한 차례는 inferred, derived, human, observed이고 빈 글은 없는 것으로 봅니다.
func weaker(a, b string) string {
	rank := map[string]int{atlas.Inferred: 0, atlas.Derived: 1, atlas.Human: 2, atlas.Observed: 3}
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if rank[b] < rank[a] {
		return b
	}
	return a
}

// eventBand 메서드는 사건의 띠를 돌려줍니다. 도착이 있으면 도착, 없으면 자리로 정합니다. 맞는 규칙이 없으면 빈 글입니다.
//
// 규칙은 띠의 차례로 보되, 대체 규칙(Fallback)은 다른 모든 규칙이 맞지 않을 때만 봅니다.
func (x *index) eventBand(e *atlas.Event) banded {
	if b, ok := x.band[e.ID]; ok {
		return b
	}
	place := e.Place
	if e.Target != nil {
		place = *e.Target
	}
	b := x.match(e, place)
	x.band[e.ID] = b
	return b
}

// match 메서드는 사건 하나(nil이면 장소만)와 장소 하나로 띠를 정합니다.
func (x *index) match(e *atlas.Event, place int) banded {
	for _, fallback := range []bool{false, true} {
		for _, b := range x.spec.Bands {
			for _, rule := range b.Match {
				if rule.Fallback != fallback {
					continue
				}
				if e == nil && (len(rule.Names) > 0 || len(rule.Sources) > 0 || rule.Attr != "") {
					continue
				}
				if x.ruleMatch(rule, e, place) {
					out := banded{band: b.ID}
					if rule.Attr != "" {
						src, _, _ := strings.Cut(rule.Attr, ".")
						out.nature = x.nature[src]
					}
					return out
				}
			}
		}
	}
	return banded{}
}

// ruleMatch 메서드는 규칙의 채운 조건이 모두 맞는지 봅니다. e가 nil이면 장소 조건만 봅니다.
func (x *index) ruleMatch(rule Rule, e *atlas.Event, place int) bool {
	if e != nil {
		if len(rule.Names) > 0 && !contains(rule.Names, e.Name) {
			return false
		}
		if len(rule.Sources) > 0 && !contains(rule.Sources, e.Source) {
			return false
		}
		if rule.Attr != "" {
			val, ok := attrValue(e.Attrs, rule.Attr)
			if !ok || (len(rule.Values) > 0 && !contains(rule.Values, val)) {
				return false
			}
		}
	}
	p, ok := x.places[place]
	if !ok {
		return rule.Place == nil && rule.Ancestor == nil
	}
	if rule.Place != nil && !rule.Place.match(*p) {
		return false
	}
	if rule.Ancestor != nil {
		for q := p; ; q = x.places[*q.Parent] {
			if rule.Ancestor.match(*q) {
				return true
			}
			if q.Parent == nil {
				return false
			}
		}
	}
	return true
}

// buildRoutes 함수는 run 하나에서 장면 규칙대로 경로를 만듭니다. 뿌리 규칙이 없으면 run 전체가 경로 하나입니다.
func buildRoutes(spec *Spec, sc *SceneSpec, r *atlas.Run) ([]built, error) {
	x := newIndex(spec, r)
	roots := []*atlas.Event{nil}
	if sc.Root != nil {
		roots = nil
		for _, row := range atlas.Select(r, *sc.Root) {
			roots = append(roots, x.events[row.Event])
		}
		if len(roots) == 0 {
			return nil, fmt.Errorf("뿌리 규칙에 맞는 사건이 없음")
		}
	}
	var out []built
	for _, root := range roots {
		b, err := buildRoute(sc, x, root)
		if err != nil {
			return nil, err
		}
		if b != nil {
			out = append(out, *b)
		}
	}
	return out, nil
}

// reach 함수는 뿌리에서 kinds 연결을 따라 닿는 사건의 집합입니다(뿌리는 빼고).
func reach(x *index, root int, kinds []string) map[int]bool {
	seen := map[int]bool{}
	queue := []int{root}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, l := range x.out[id] {
			if contains(kinds, l.Kind) && !seen[l.Dst] {
				seen[l.Dst] = true
				queue = append(queue, l.Dst)
			}
		}
	}
	return seen
}

// buildRoute 함수는 뿌리 하나의 경로를 만듭니다. 칸이 될 사건이 없으면 nil입니다.
func buildRoute(sc *SceneSpec, x *index, root *atlas.Event) (*built, error) {
	r := x.r
	// 칸이 될 사건을 고릅니다. Via가 있으면 뿌리에서 그 연결을 따라 닿는 사건만 후보입니다.
	var allowed map[int]bool
	if len(sc.Via) > 0 && root != nil {
		allowed = reach(x, root.ID, sc.Via)
	}
	skip := map[int]bool{}
	if sc.Skip != nil {
		for _, row := range atlas.Select(r, *sc.Skip) {
			skip[row.Event] = true
		}
	}
	var picked []*atlas.Event
	for _, row := range atlas.Select(r, sc.Steps) {
		if skip[row.Event] || (allowed != nil && !allowed[row.Event]) {
			continue
		}
		picked = append(picked, x.events[row.Event])
	}
	if len(picked) == 0 {
		return nil, nil
	}

	// 칸의 가로 차례는 시각 있는 사건만 정하고, 그 사건들은 한 시계여야 합니다(고친 규칙 2). 시각 없는 사건은 칸이 되지 않습니다.
	// 고른 사건이 모두 시각이 없을 때만 id 차례(원본 목록의 차례)로 칸을 만들고 「순서만 기록」으로 표시합니다.
	var timed []*atlas.Event
	for _, e := range picked {
		if e.StartNS != nil {
			timed = append(timed, e)
		}
	}
	rt := &built{Route: Route{Run: r.ID, Scene: sc.ID, SceneLabel: sc.Label, Home: sc.Home}}
	if len(timed) == 0 {
		// 출처 전체가 시각이 없을 때(시계가 없는 출처, 예: gdb)만 순서만 기록된 칸입니다. 시계가 있는 출처인데 이 뿌리에 시각 있는 사건이 없으면 경로가 없습니다.
		for _, e := range picked {
			if x.clock[e.Source] != "" {
				return nil, nil
			}
		}
		timed, rt.OrderOnly = picked, true
	} else {
		for _, e := range timed {
			c := x.clock[e.Source]
			if rt.Clock == "" {
				rt.Clock = c
			} else if c != rt.Clock {
				return nil, fmt.Errorf("칸의 차례를 정하는 사건이 두 시계(%s, %s)에 걸침: 사건 %d", rt.Clock, c, e.ID)
			}
		}
		sort.SliceStable(timed, func(i, j int) bool {
			if *timed[i].StartNS != *timed[j].StartNS {
				return *timed[i].StartNS < *timed[j].StartNS
			}
			return timed[i].ID < timed[j].ID
		})
	}
	if root != nil {
		rt.ID = fmt.Sprintf("%s#%d", r.ID, root.ID)
		rt.Process = x.places[root.Place].Name
	} else {
		rt.ID = r.ID
	}
	rt.identity = identityOf(sc.Identity, x, root)
	rt.Name = identityOf(sc.Name, x, root)
	if sc.Typed != "" {
		if t, ok := attrValue(r.Attrs, sc.Typed); ok {
			rt.Typed = t
		} else if root != nil {
			rt.Typed, _ = attrValue(root.Attrs, sc.Typed)
		}
	}

	// 칸을 만들고, 같은 열 안에서 접기 조건에 맞는 잇단 칸을 한 칸 「×N」으로 접습니다(고친 규칙 3).
	column := func(e *atlas.Event) string {
		if sc.Columns == nil {
			return ""
		}
		v, _ := attrValue(e.Attrs, sc.Columns.Attr)
		return v
	}
	foldable := func(e *atlas.Event) bool {
		if sc.Fold == nil {
			return false
		}
		v, ok := attrValue(e.Attrs, sc.Fold.Attr)
		return ok && v == "true"
	}
	var groups [][]*atlas.Event
	for _, e := range timed {
		if n := len(groups); n > 0 && foldable(e) {
			last := groups[n-1]
			if foldable(last[0]) && column(last[0]) == column(e) {
				groups[n-1] = append(last, e)
				continue
			}
		}
		groups = append(groups, []*atlas.Event{e})
	}
	inTree := map[int]bool{}
	for i, g := range groups {
		st, err := makeStep(sc, x, g, rt.Clock)
		if err != nil {
			return nil, err
		}
		st.Index, st.Column = i, column(g[0])
		if len(g) > 1 {
			st.Fold, st.Name = len(g), sc.Fold.Label
		}
		for _, n := range st.Tree {
			inTree[n.Event] = true
		}
		rt.Steps = append(rt.Steps, st)
	}

	// 경로의 구간은 칸들이 덮는 구간입니다.
	first := true
	for _, s := range rt.Steps {
		for _, t := range []*int64{s.T0, s.T1} {
			if t == nil {
				continue
			}
			if first || *t < rt.StartNS {
				rt.StartNS = *t
			}
			if first || *t > rt.EndNS {
				rt.EndNS = *t
			}
			first = false
		}
	}
	if sc.Columns != nil {
		rt.Columns = columnsOf(sc, x, rt)
	}

	// 뿌리가 품었지만 어느 칸도 품지 않은 시각 없는 사건은 경로의 목록에만 둡니다.
	for id := range allowed {
		e := x.events[id]
		if e.StartNS != nil || inTree[id] || skip[id] {
			continue
		}
		b := x.eventBand(e)
		l := Loose{Event: id, Name: e.Name, Band: b.band, Nature: weaker(x.nature[e.Source], b.nature)}
		l.Text, _ = attrValue(e.Attrs, e.Source+".text")
		rt.Untimed = append(rt.Untimed, l)
	}
	sort.Slice(rt.Untimed, func(i, j int) bool { return rt.Untimed[i].Event < rt.Untimed[j].Event })

	var all []Node
	for _, s := range rt.Steps {
		all = append(all, s.Tree...)
	}
	rt.Bands = bandStates(x, all, rt.Clock)
	rt.Zones = zoneStates(x.spec, rt.Bands)
	rt.System = systemSteps(x.spec, rt.Steps)
	return rt, nil
}

// columnsOf 함수는 처리 단계 열을 만듭니다. 열의 시작은 자료 파일이 고른 사건의 시각이고, 머문 시간은 다음 열의 시작까지(마지막은 경로 끝)입니다.
func columnsOf(sc *SceneSpec, x *index, rt *built) []Column {
	var cols []Column
	for _, row := range atlas.Select(x.r, sc.Columns.Starts) {
		e := x.events[row.Event]
		if e.StartNS == nil {
			continue
		}
		cols = append(cols, Column{Name: e.Name, StartNS: *e.StartNS, Steps: [2]int{-1, -1}, Nature: x.nature[e.Source]})
	}
	sort.SliceStable(cols, func(i, j int) bool { return cols[i].StartNS < cols[j].StartNS })
	for i := range cols {
		next := rt.EndNS
		if i+1 < len(cols) {
			next = cols[i+1].StartNS
		}
		cols[i].DwellNS = next - cols[i].StartNS
		for _, s := range rt.Steps {
			if s.Column != cols[i].Name {
				continue
			}
			if cols[i].Steps[0] < 0 {
				cols[i].Steps[0] = s.Index
			}
			cols[i].Steps[1] = s.Index
		}
	}
	return cols
}

// identityOf 함수는 정체 규칙대로 값을 이어 붙입니다. 값이 없는 칸은 빈 글로 둡니다.
func identityOf(id Identity, x *index, root *atlas.Event) string {
	var parts []string
	switch id.RunField {
	case "scene":
		parts = append(parts, x.r.Scene)
	case "name":
		parts = append(parts, x.r.Name)
	case "id":
		parts = append(parts, x.r.ID)
	case "scenario":
		// 시나리오는 run 이름의 마지막 "/" 앞입니다(같은 시나리오를 여러 번 돌린 a·b…).
		s := x.r.Name
		if i := strings.LastIndex(s, "/"); i >= 0 {
			s = s[:i]
		}
		parts = append(parts, s)
	}
	if id.RunAttr != "" {
		v, _ := attrValue(x.r.Attrs, id.RunAttr)
		parts = append(parts, v)
	}
	if root != nil {
		if id.RootPlace {
			parts = append(parts, x.places[root.Place].Name)
		}
		if id.RootAttr != "" {
			v, _ := attrValue(root.Attrs, id.RootAttr)
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, " · ")
}

// makeStep 함수는 사건 하나(접으면 여럿)로 칸 하나를 만들고, 그 칸을 펼친 함수 줌 나무, 잎 집합, 띠 상태, 하강을 셈합니다.
func makeStep(sc *SceneSpec, x *index, evs []*atlas.Event, clock string) (Step, error) {
	e := evs[0]
	b := x.eventBand(e)
	if b.band == "" {
		return Step{}, fmt.Errorf("사건 %d(%s)이 띠 규칙에 맞지 않음", e.ID, e.Name)
	}
	st := Step{Name: e.Name, Band: b.band}
	if e.From != nil {
		st.Origin = x.match(e, *e.From).band
	}
	if p, ok := x.places[e.Place]; ok {
		st.Place, st.Key = placeLabel(x, p), p.Key
		if p.Loc != nil {
			st.Loc = fmt.Sprintf("%s:%d", p.Loc.Path, p.Loc.Line)
		}
	}
	if e.Target != nil {
		st.Place = placeLabel(x, x.places[*e.Target])
	}
	for _, f := range e.Stack {
		if p, ok := x.places[f]; ok {
			st.Stack = append(st.Stack, placeLabel(x, p))
		}
	}
	st.Nature = b.nature
	for _, ev := range evs {
		st.Events = append(st.Events, ev.ID)
		st.Count += ev.Count
		st.Nature = weaker(st.Nature, weaker(x.nature[ev.Source], x.eventBand(ev).nature))
		if ev.StartNS != nil && (st.T0 == nil || *ev.StartNS < *st.T0) {
			st.T0 = atlas.Ptr(*ev.StartNS)
		}
		end := ev.EndNS
		if end == nil {
			end = ev.StartNS
		}
		if end != nil && (st.T1 == nil || *end > *st.T1) {
			st.T1 = atlas.Ptr(*end)
		}
	}
	st.Attrs = attrViews(x, e)

	tree, err := buildTree(sc, x, evs, clock)
	if err != nil {
		return st, err
	}
	st.Tree = tree
	leaves := map[int]bool{}
	for _, n := range tree {
		if len(n.Leaves) == 1 && n.Leaves[0] == n.Event {
			leaves[n.Event] = true
		}
	}
	st.Leaves = sortedIDs(leaves)
	st.Bands = bandStates(x, tree, clock)

	// 머문 시간은 칸의 구간, 없으면 펼친 나무에서 가장 긴 마디의 길이입니다.
	if st.T0 != nil && st.T1 != nil {
		st.DurNS = *st.T1 - *st.T0
	}
	for _, n := range tree {
		if n.DurNS != nil && *n.DurNS > st.DurNS {
			st.DurNS = *n.DurNS
		}
	}
	st.SleepNS = sleepTime(tree)
	st.Descent = descentOf(x.spec, tree)
	applyLens(tree, st.DurNS, sc.Lens.MinShare)
	return st, nil
}

// placeLabel 함수는 장소를 화면에 적을 이름입니다. 이름이 없는 코드 자리는 경로:줄입니다.
func placeLabel(x *index, p *atlas.Place) string {
	if p.Name == "" && p.Loc != nil {
		return fmt.Sprintf("%s:%d", p.Loc.Path, p.Loc.Line)
	}
	return p.Name
}

// attrViews 함수는 사건의 attrs 가운데 짧은 값만 출처 성격과 함께 고릅니다(자세한 내용 칸). 열쇠 차례로 둡니다.
func attrViews(x *index, e *atlas.Event) []AttrView {
	var out []AttrView
	srcs := make([]string, 0, len(e.Attrs))
	for s := range e.Attrs {
		srcs = append(srcs, s)
	}
	sort.Strings(srcs)
	for _, s := range srcs {
		keys := make([]string, 0, len(e.Attrs[s]))
		for k := range e.Attrs[s] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			val := e.Attrs[s][k]
			text := fmt.Sprint(val)
			switch val.(type) {
			case map[string]any, []any:
				b, _ := json.Marshal(val)
				text = string(b)
			}
			if len(text) > 160 {
				continue
			}
			out = append(out, AttrView{Key: s + "." + k, Value: text, Nature: x.nature[s]})
		}
	}
	return out
}

// buildTree 함수는 칸의 사건들에서 자료 파일의 연결과 parent 자식을 따라 닿는 사건을 먼저 부른 차례(나무의 앞차례)로 놓습니다.
//
// 나가는 연결과 parent 자식은 끝까지 따라가고, 들어오는 연결(dir in)은 그 출발 사건만 붙입니다. 한 사건은 나무에 한 번만 나옵니다.
// 마디의 가로 차례는 나무 차례라 시각을 견주지 않습니다. 그래서 칸과 시계가 다른 시각 있는 마디에는 OtherClock 표지를 붙입니다.
func buildTree(sc *SceneSpec, x *index, roots []*atlas.Event, clock string) ([]Node, error) {
	var tree []Node
	seen := map[int]bool{}
	var add func(e *atlas.Event, parent, depth int, via, linkNature string, follow bool) error
	add = func(e *atlas.Event, parent, depth int, via, linkNature string, follow bool) error {
		if seen[e.ID] {
			return nil
		}
		seen[e.ID] = true
		b := x.eventBand(e)
		if b.band == "" {
			return fmt.Errorf("사건 %d(%s)이 띠 규칙에 맞지 않음", e.ID, e.Name)
		}
		p := x.places[e.Place]
		n := Node{Event: e.ID, Name: p.Name, Band: b.band, Depth: depth, Parent: parent, Via: via, Key: p.Key,
			Nature: weaker(weaker(x.nature[e.Source], linkNature), b.nature), T0: e.StartNS}
		switch p.Kind {
		case "process", "connection", "lane", "client", "session", "site", "role", "app":
			n.Name = e.Name
		}
		if p.Loc != nil {
			n.Loc = fmt.Sprintf("%s:%d", p.Loc.Path, p.Loc.Line)
		}
		switch {
		case e.DurNS != nil:
			n.DurNS = atlas.Ptr(*e.DurNS)
		case e.StartNS != nil && e.EndNS != nil:
			n.DurNS = atlas.Ptr(*e.EndNS - *e.StartNS)
		}
		n.OtherClock = e.StartNS != nil && clock != "" && x.clock[e.Source] != clock
		n.Sleep = contains(x.spec.Sleep, p.Name)
		me := len(tree)
		tree = append(tree, n)
		if !follow {
			return nil
		}
		// 자식: parent 자식, 그다음 자료 파일의 연결 차례로, 각각 상대 사건의 id 차례입니다.
		for _, k := range x.kids[e.ID] {
			if err := add(x.events[k], me, depth+1, "parent", "", true); err != nil {
				return err
			}
		}
		for _, ex := range sc.Expand {
			if ex.Dir == "in" {
				for _, l := range x.in[e.ID] {
					if l.Kind == ex.Link {
						if err := add(x.events[l.Src], me, depth+1, l.Kind, x.nature[l.Basis], false); err != nil {
							return err
						}
					}
				}
				continue
			}
			for _, l := range x.out[e.ID] {
				if l.Kind == ex.Link {
					if err := add(x.events[l.Dst], me, depth+1, l.Kind, x.nature[l.Basis], true); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for _, e := range roots {
		if err := add(e, -1, 0, "", "", true); err != nil {
			return nil, err
		}
	}
	// 마디마다 잎 집합(자기 아래의 잎 사건)을 모읍니다. 뒤에서 앞으로 모으면 자손이 먼저 끝납니다.
	sets := make([]map[int]bool, len(tree))
	hasKid := make([]bool, len(tree))
	for i := range tree {
		if p := tree[i].Parent; p >= 0 {
			hasKid[p] = true
		}
	}
	for i := len(tree) - 1; i >= 0; i-- {
		if sets[i] == nil {
			sets[i] = map[int]bool{}
		}
		if !hasKid[i] {
			sets[i][tree[i].Event] = true
		}
		if p := tree[i].Parent; p >= 0 {
			if sets[p] == nil {
				sets[p] = map[int]bool{}
			}
			for id := range sets[i] {
				sets[p][id] = true
			}
		}
	}
	for i := range tree {
		tree[i].Leaves = sortedIDs(sets[i])
	}
	return tree, nil
}

// applyLens 함수는 렌즈가 켜졌을 때 보이는 마디를 정합니다(고친 규칙 7).
//
// 뿌리, 칸의 머문 시간(base)에 대한 몫이 minShare 이상인 마디, 잠드는 함수, 그리고 그 조상이 보입니다. 길이가 없는 마디는 부모가 보이면 보입니다.
// base가 0이면(길이가 없는 나무) 모두 보입니다. 숨은 마디는 보이는 가장 가까운 조상의 Folded에 셉니다.
func applyLens(tree []Node, base int64, minShare float64) {
	if base <= 0 {
		for i := range tree {
			tree[i].Visible = true
		}
		return
	}
	keep := make([]bool, len(tree))
	for i, n := range tree {
		if n.Parent < 0 || n.Sleep || (n.DurNS != nil && float64(*n.DurNS) >= minShare*float64(base)) {
			keep[i] = true
		}
	}
	for i := len(tree) - 1; i >= 0; i-- {
		if keep[i] {
			for p := tree[i].Parent; p >= 0 && !keep[p]; p = tree[p].Parent {
				keep[p] = true
			}
		}
	}
	for i, n := range tree {
		if !keep[i] && n.DurNS == nil && n.Parent >= 0 && keep[n.Parent] {
			keep[i] = true
		}
	}
	for i := range tree {
		tree[i].Visible = keep[i]
		if keep[i] {
			continue
		}
		for p := tree[i].Parent; p >= 0; p = tree[p].Parent {
			if keep[p] {
				tree[p].Folded++
				break
			}
		}
	}
}

// sortedIDs 함수는 집합의 id를 차례로 돌려줍니다.
func sortedIDs(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for id := range m {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

// sleepTime 함수는 잠든 마디의 길이를 더합니다. 잠든 마디 안의 잠든 마디는 한 번만 셉니다.
func sleepTime(tree []Node) int64 {
	var sum int64
	for _, n := range tree {
		if !n.Sleep || n.DurNS == nil {
			continue
		}
		inner := false
		for p := n.Parent; p >= 0; p = tree[p].Parent {
			if tree[p].Sleep {
				inner = true
				break
			}
		}
		if !inner {
			sum += *n.DurNS
		}
	}
	return sum
}

// descentOf 함수는 칸의 나무가 지면을 건너는지 봅니다(고친 규칙 4). 지면 띠의 마디가 없으면 nil입니다.
//
// 지면 마디가 하나라도 그 아래에 지면보다 아래 띠의 마디를 가지면 기록됨(events), 아니면 안 봄(unseen)입니다.
// Count는 지면 마디의 수이고, Nature는 뿌리에서 지면 마디까지 온 길과 그 아래 마디의 가장 약한 성격입니다.
func descentOf(spec *Spec, tree []Node) *Descent {
	gi := groundIndex(spec)
	order := map[string]int{}
	for i, b := range spec.Bands {
		order[b.ID] = i
	}
	var d *Descent
	for i, n := range tree {
		if order[n.Band] != gi {
			continue
		}
		if d == nil {
			d = &Descent{State: StateUnseen}
		}
		d.Count++
		if n.DurNS != nil {
			d.DurNS += *n.DurNS
		}
		for p := i; p >= 0; p = tree[p].Parent {
			d.Nature = weaker(d.Nature, tree[p].Nature)
		}
		for j := i + 1; j < len(tree); j++ {
			if isDescendant(tree, j, i) && order[tree[j].Band] > gi {
				d.State = StateEvents
				d.Nature = weaker(d.Nature, tree[j].Nature)
				break
			}
		}
	}
	return d
}

// isDescendant 함수는 나무에서 j가 i의 자손인지 봅니다.
func isDescendant(tree []Node, j, i int) bool {
	for p := tree[j].Parent; p >= 0; p = tree[p].Parent {
		if p == i {
			return true
		}
	}
	return false
}

// bandStates 함수는 보이는 범위(마디들)에서 띠마다의 상태를 셈합니다(고친 규칙 5).
//
// 보이는 범위에 그 띠의 마디가 있으면 사건 있음(함수 줌 띠는 합쳐 보일 띠에도 셈)입니다. 없으면 보이는 범위의 시간대(clock의 시각을 가진 마디로 정함)
// 안에 run의 다른 사건이 그 띠에 있으면 흐림, 시간대를 정할 수 있고 complete인 관찰 출처의 관찰 범위가 그 띠의 장소를 덮으며 그 띠에 시간대를 견줄 수
// 없는 사건도 없으면 봤고 0번, 나머지는 안 봄입니다.
func bandStates(x *index, nodes []Node, clock string) map[string]string {
	under := map[string]string{}
	for _, b := range x.spec.Bands {
		under[b.ID] = b.Under
	}
	has := map[string]bool{}
	in := map[int]bool{}
	var lo, hi int64
	span := false
	for _, n := range nodes {
		has[n.Band] = true
		if u := under[n.Band]; u != "" {
			has[u] = true
		}
		in[n.Event] = true
		e := x.events[n.Event]
		if e.StartNS == nil || x.clock[e.Source] != clock || clock == "" {
			continue
		}
		end := *e.StartNS
		if e.EndNS != nil {
			end = *e.EndNS
		}
		if !span || *e.StartNS < lo {
			lo = *e.StartNS
		}
		if !span || end > hi {
			hi = end
		}
		span = true
	}
	// run의 다른 사건을 띠마다 가릅니다: 시간대 안(같은 시계), 시간대를 견줄 수 없음(시각이 없거나 다른 시계).
	other, unknown := map[string]bool{}, map[string]bool{}
	for i := range x.r.Events {
		e := &x.r.Events[i]
		if in[e.ID] {
			continue
		}
		b := x.eventBand(e).band
		if b == "" {
			continue
		}
		// 다른 시계의 사건은 관계 범위만큼 넓혀 이 시계로 옮깁니다(t_c − t ∈ [olo, ohi] 이면 t ∈ [t_c − ohi, t_c − olo]). 관계를 모르면 견줄 수 없습니다.
		olo, ohi, known := atlas.Offset(x.r, x.clock[e.Source], clock)
		if !span || e.StartNS == nil || !known {
			unknown[b] = true
			continue
		}
		start, end := *e.StartNS, *e.StartNS
		if e.EndNS != nil {
			end = *e.EndNS
		}
		start, end = start-ohi, end-olo
		if start <= hi && end >= lo {
			other[b] = true
			if u := under[b]; u != "" {
				other[u] = true
			}
		}
	}
	watched := map[string]bool{}
	if span {
		for _, c := range x.r.Coverage {
			if c.Completeness != "complete" || x.nature[c.Source] != atlas.Observed {
				continue
			}
			if c.Window != nil && (c.Window[0] > lo || c.Window[1] < hi) {
				continue
			}
			for _, p := range c.Places {
				if b := x.match(nil, p).band; b != "" {
					watched[b] = true
				}
			}
		}
	}
	out := map[string]string{}
	for _, b := range x.spec.Bands {
		switch {
		case has[b.ID]:
			out[b.ID] = StateEvents
		case other[b.ID]:
			out[b.ID] = StateOther
		case watched[b.ID] && !unknown[b.ID]:
			out[b.ID] = StateWatched
		default:
			out[b.ID] = StateUnseen
		}
	}
	return out
}

// zoneStates 함수는 구역마다 그 안 띠들의 상태 가운데 가장 강한 것을 돌려줍니다(사건 있음 > 흐림 > 봤고 0번 > 안 봄).
func zoneStates(spec *Spec, bands map[string]string) map[string]string {
	rank := map[string]int{StateUnseen: 0, StateWatched: 1, StateOther: 2, StateEvents: 3}
	out := map[string]string{}
	for _, z := range spec.Zones {
		out[z.ID] = StateUnseen
	}
	for _, b := range spec.Bands {
		if rank[bands[b.ID]] > rank[out[b.Zone]] {
			out[b.Zone] = bands[b.ID]
		}
	}
	return out
}

// systemSteps 함수는 프로세스 안 칸을 구역으로 올립니다. 칸마다 지난 구역의 집합을 셈하고, 집합이 같은 잇단 칸을 한 칸으로 합칩니다.
//
// 지난 구역은 칸의 출발 띠, 펼친 나무의 띠, 칸의 띠가 속한 구역이고, 차례는 처음 만난 차례입니다.
func systemSteps(spec *Spec, steps []Step) []SysStep {
	zoneOf := map[string]string{}
	for _, b := range spec.Bands {
		zoneOf[b.ID] = b.Zone
	}
	var out []SysStep
	var prevKey string
	for i, s := range steps {
		var zs []string
		addZ := func(band string) {
			if z := zoneOf[band]; z != "" && !contains(zs, z) {
				zs = append(zs, z)
			}
		}
		addZ(s.Origin)
		for _, n := range s.Tree {
			addZ(n.Band)
		}
		addZ(s.Band)
		sorted := append([]string{}, zs...)
		sort.Strings(sorted)
		key := strings.Join(sorted, ",")
		if len(out) > 0 && key == prevKey {
			last := &out[len(out)-1]
			last.Steps[1] = i
			last.Leaves = union(last.Leaves, s.Leaves)
			last.DurNS += s.DurNS
			last.Nature = weaker(last.Nature, s.Nature)
			continue
		}
		out = append(out, SysStep{Index: len(out), Zones: zs, Steps: [2]int{i, i}, Leaves: append([]int{}, s.Leaves...), DurNS: s.DurNS, Nature: s.Nature})
		prevKey = key
	}
	return out
}

// union 함수는 두 id 목록의 합집합을 정렬해 돌려줍니다.
func union(a, b []int) []int {
	m := map[int]bool{}
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		m[v] = true
	}
	return sortedIDs(m)
}

// nodeDur 함수는 마디의 길이입니다. 없으면 0입니다(나무의 길이는 dur_ns, 없으면 끝 − 시작으로 이미 셈함).
func nodeDur(n Node) int64 {
	if n.DurNS == nil {
		return 0
	}
	return *n.DurNS
}

// diverge 함수는 경로 a가 같은 무리의 경로 b와 갈리는 곳을 셈합니다(고친 규칙 9).
//
// 칸 차례(띠와 이름)가 다르면 처음 다른 칸이 갈림입니다. 같으면 나무가 있는 칸마다 두 나무를 같이 걸어 봅니다: 뿌리 목록부터 자식 차례(이름과 띠)가
// 같은 동안 머문 시간 차가 가장 큰 자식 짝을 따라가다가, 자식 차례가 다른 마디를 만나면 그 칸의 주된 갈림(tree)이고, 잎까지 같으면 그 길의 끝 마디가
// 「구조 같음, 시간 차 최대 마디」(time)입니다. 한쪽만 아래가 있으면 '한쪽만 기록'(unrecorded)입니다. 고르는 차례는 tree, unrecorded, time이고
// 같은 종류 안에서는 칸의 머문 시간 차가 가장 큰 것입니다. 뿌리만 견준 time은 갈림으로 내지 않습니다.
func diverge(a, b *Route, other int) Divergence {
	d := Divergence{Other: other, Node: -1, Step: -1}
	n := min(len(a.Steps), len(b.Steps))
	for i := 0; i < n; i++ {
		if a.Steps[i].Band != b.Steps[i].Band || a.Steps[i].Name != b.Steps[i].Name {
			d.Kind, d.Step, d.Name = "steps", i, a.Steps[i].Name
			return d
		}
	}
	if len(a.Steps) != len(b.Steps) {
		d.Kind, d.Step = "steps", n
		if n < len(a.Steps) {
			d.Name = a.Steps[n].Name
		}
		return d
	}
	best := map[string]Divergence{}
	bestDiff := map[string]int64{"tree": -1, "time": -1, "unrecorded": -1}
	for i := range a.Steps {
		if len(a.Steps[i].Tree) < 2 || len(b.Steps[i].Tree) < 2 {
			continue
		}
		kind, node, diff := walkTrees(a.Steps[i].Tree, b.Steps[i].Tree)
		if kind == "time" && a.Steps[i].Tree[node].Parent < 0 {
			// 뿌리만 견준 시간 차(펼칠 것이 없는 칸)는 갈림이 아닙니다.
			continue
		}
		if stepDiff := abs(a.Steps[i].DurNS - b.Steps[i].DurNS); stepDiff > bestDiff[kind] {
			bestDiff[kind] = stepDiff
			best[kind] = Divergence{Other: other, Kind: kind, Step: i, Node: node, Name: a.Steps[i].Tree[node].Name, DiffNS: diff}
		}
	}
	if x, ok := best["tree"]; ok {
		return x
	}
	if x, ok := best["unrecorded"]; ok {
		return x
	}
	if x, ok := best["time"]; ok {
		return x
	}
	d.Kind = "none"
	return d
}

// walkTrees 함수는 두 나무를 뿌리 목록부터 같이 걸어 갈림의 종류(tree 또는 time), a 나무의 마디 번호, 그 마디의 머문 시간 차를 돌려줍니다.
func walkTrees(ta, tb []Node) (string, int, int64) {
	kidsOf := func(t []Node, i int) []int {
		var out []int
		for j, nd := range t {
			if nd.Parent == i {
				out = append(out, j)
			}
		}
		return out
	}
	same := func(x, y []int) bool {
		if len(x) != len(y) {
			return false
		}
		for k := range x {
			if ta[x[k]].Name != tb[y[k]].Name || ta[x[k]].Band != tb[y[k]].Band {
				return false
			}
		}
		return true
	}
	i, j := -1, -1
	for {
		ka, kb := kidsOf(ta, i), kidsOf(tb, j)
		if (len(ka) == 0) != (len(kb) == 0) && i >= 0 {
			// 한쪽만 그 아래가 있으면 실행이 갈린 것인지 기록 범위가 다른 것인지 가를 수 없습니다(X1). 따로 표시합니다.
			return "unrecorded", i, abs(nodeDur(ta[i]) - nodeDur(tb[j]))
		}
		if !same(ka, kb) {
			if i < 0 {
				i, j = 0, 0
			}
			return "tree", i, abs(nodeDur(ta[i]) - nodeDur(tb[j]))
		}
		if len(ka) == 0 {
			return "time", i, abs(nodeDur(ta[i]) - nodeDur(tb[j]))
		}
		k := 0
		for m := range ka {
			if abs(nodeDur(ta[ka[m]])-nodeDur(tb[kb[m]])) > abs(nodeDur(ta[ka[k]])-nodeDur(tb[kb[k]])) {
				k = m
			}
		}
		i, j = ka[k], kb[k]
	}
}

// abs 함수는 정수의 절댓값입니다.
func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
