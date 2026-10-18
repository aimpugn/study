package atlas

import (
	"fmt"
)

// Validate 함수는 run 하나가 형식 규칙을 지키는지 보고, 어긴 곳을 모두 돌려줍니다. 빈 목록이면 통과입니다.
//
// 보는 규칙: id가 차례대로 매겨졌는지, 참조(장소·부모·대상·스택·연결·설명 대상·출처·시계)가 모두 풀리는지,
// 장소 사슬과 부모 사슬에 고리가 없는지, 깊이가 부모 사슬로 셈한 값과 같은지, 부모와 자식의 출처가 같은지,
// group이 부모가 아닌지, count가 종류와 맞는지, attrs의 키가 모두 출처 id인지, 시각이 있으면 시작 ≤ 끝인지,
// 시작·끝과 길이가 함께 있지 않은지, 출처의 성격과 시계가 정해진 값인지.
func Validate(r *Run) []error {
	v := validator{run: r}
	v.header()
	v.places()
	v.events()
	v.links()
	v.notes()
	return v.errs
}

// validator 값은 Validate가 어긴 곳을 모으는 그릇입니다.
type validator struct {
	run  *Run
	errs []error
}

// fail 메서드는 어긴 곳 하나를 적습니다.
func (v *validator) fail(format string, a ...any) {
	v.errs = append(v.errs, fmt.Errorf(format, a...))
}

// header 메서드는 run 머리(시계, 출처, 관찰 범위, 미기록, attrs)를 봅니다.
func (v *validator) header() {
	r := v.run
	// 시계의 관계는 있는 시계를 가리켜야 하고, 관계를 낸 출처도 있어야 합니다.
	clocks := map[string]bool{}
	for _, c := range r.Clocks {
		if clocks[c.ID] {
			v.fail("시계 %q가 두 번", c.ID)
		}
		clocks[c.ID] = true
	}
	for _, c := range r.Clocks {
		if c.Relation == nil {
			continue
		}
		if !clocks[c.Relation.To] {
			v.fail("시계 %q의 관계가 없는 시계 %q를 가리킴", c.ID, c.Relation.To)
		}
		if _, ok := r.SourceByID(c.Relation.Basis); !ok {
			v.fail("시계 %q의 관계 근거 출처 %q가 없음", c.ID, c.Relation.Basis)
		}
		if c.Relation.Known && c.Relation.LoNS > c.Relation.HiNS {
			v.fail("시계 %q의 차이 범위가 거꾸로임", c.ID)
		}
	}

	// 출처는 정해진 성격을 가지고, 시계를 적었으면 그 시계가 있어야 합니다.
	sources := map[string]bool{}
	for _, s := range r.Sources {
		if sources[s.ID] {
			v.fail("출처 %q가 두 번", s.ID)
		}
		sources[s.ID] = true
		switch s.Nature {
		case Observed, Derived, Inferred, Human:
		default:
			v.fail("출처 %q의 성격 %q는 정해진 값이 아님", s.ID, s.Nature)
		}
		if s.Clock != "" && !clocks[s.Clock] {
			v.fail("출처 %q의 시계 %q가 없음", s.ID, s.Clock)
		}
	}
	if r.Origin.Clock != "" && !clocks[r.Origin.Clock] {
		v.fail("원점의 시계 %q가 없음", r.Origin.Clock)
	}

	for i, c := range r.Coverage {
		v.source(c.Source, "관찰 범위 %d", i)
		for _, p := range c.Places {
			v.place(p, "관찰 범위 %d", i)
		}
		switch c.Completeness {
		case "complete", "partial", "sampled":
		default:
			v.fail("관찰 범위 %d의 완전성 %q는 정해진 값이 아님", i, c.Completeness)
		}
	}
	for i, u := range r.Unrecorded {
		v.source(u.Source, "미기록 %d", i)
	}
	v.attrs(r.Attrs, "run")
}

// places 메서드는 장소 표를 봅니다.
func (v *validator) places() {
	for i, p := range v.run.Places {
		if p.ID != i {
			v.fail("장소 %d의 id가 %d", i, p.ID)
		}
		if p.Parent != nil {
			v.place(*p.Parent, "장소 %d의 부모", i)
		}
		v.source(p.Source, "장소 %d", i)
		v.attrs(p.Attrs, fmt.Sprintf("장소 %d", i))
	}
	// 장소 사슬에 고리가 있으면 맨 위에 닿지 못합니다. 장소 수만큼 올라가도 끝나지 않으면 고리입니다.
	for i := range v.run.Places {
		if v.cycle(i, len(v.run.Places), func(j int) *int {
			if j < 0 || j >= len(v.run.Places) {
				return nil
			}
			return v.run.Places[j].Parent
		}) {
			v.fail("장소 %d의 부모 사슬에 고리", i)
		}
	}
}

// events 메서드는 사건 표를 봅니다.
func (v *validator) events() {
	evs := v.run.Events
	for i, e := range evs {
		where := fmt.Sprintf("사건 %d", i)
		if e.ID != i {
			v.fail("%s의 id가 %d", where, e.ID)
		}
		v.place(e.Place, "%s의 장소", where)
		if e.From != nil {
			v.place(*e.From, "%s의 출발", where)
		}
		if e.Target != nil {
			v.place(*e.Target, "%s의 대상", where)
		}
		for k, p := range e.Stack {
			v.place(p, "%s의 스택 %d번", where, k)
		}
		// 스택이 있으면 0번은 사건의 장소 자신이고, 그 뒤가 바깥쪽으로 부른 쪽입니다.
		if len(e.Stack) > 0 && e.Stack[0] != e.Place {
			v.fail("%s의 스택 0번 %d가 사건의 장소 %d가 아님", where, e.Stack[0], e.Place)
		}
		v.source(e.Source, "%s", where)
		v.attrs(e.Attrs, where)

		// 종류마다 count와 시각 칸의 약속이 다릅니다.
		switch e.Kind {
		case Span, Instant, Sample:
			if e.Count != 1 {
				v.fail("%s(%s)의 count가 %d", where, e.Kind, e.Count)
			}
		case Group:
			if e.Count < 2 {
				v.fail("%s(group)의 count가 %d", where, e.Count)
			}
		default:
			v.fail("%s의 종류 %q는 정해진 값이 아님", where, e.Kind)
		}
		if e.StartNS != nil && e.EndNS != nil && *e.StartNS > *e.EndNS {
			v.fail("%s의 시작이 끝보다 늦음", where)
		}
		if e.DurNS != nil && (e.StartNS != nil || e.EndNS != nil) {
			v.fail("%s에 시작·끝과 길이가 함께 있음", where)
		}
		if e.DurNS != nil && *e.DurNS < 0 {
			v.fail("%s의 길이가 음수", where)
		}
		if e.Kind == Instant && e.EndNS != nil && e.StartNS != nil && *e.EndNS != *e.StartNS {
			v.fail("%s(instant)의 끝이 시작과 다름", where)
		}

		// 부모는 앞에 있는 사건이고, 출처가 같고, group이 아니며, 깊이는 부모보다 하나 깊습니다.
		if e.Parent == nil {
			if e.Depth != 0 {
				v.fail("%s는 부모가 없는데 깊이가 %d", where, e.Depth)
			}
			continue
		}
		p := *e.Parent
		if p < 0 || p >= i {
			v.fail("%s의 부모 %d가 앞에 있는 사건이 아님", where, p)
			continue
		}
		if evs[p].Source != e.Source {
			v.fail("%s의 출처 %q가 부모 %d의 출처 %q와 다름", where, e.Source, p, evs[p].Source)
		}
		if evs[p].Kind == Group {
			v.fail("%s의 부모 %d가 group", where, p)
		}
		if e.Depth != evs[p].Depth+1 {
			v.fail("%s의 깊이 %d가 부모 깊이 %d + 1과 다름", where, e.Depth, evs[p].Depth)
		}
	}
}

// links 메서드는 연결 표를 봅니다.
func (v *validator) links() {
	n := len(v.run.Events)
	for i, l := range v.run.Links {
		if l.ID != i {
			v.fail("연결 %d의 id가 %d", i, l.ID)
		}
		if l.Src < 0 || l.Src >= n || l.Dst < 0 || l.Dst >= n {
			v.fail("연결 %d의 끝(%d → %d)이 없는 사건", i, l.Src, l.Dst)
		}
		if l.Kind == "" {
			v.fail("연결 %d의 종류가 비어 있음", i)
		}
		v.source(l.Basis, "연결 %d의 근거", i)
		v.attrs(l.Attrs, fmt.Sprintf("연결 %d", i))
	}
}

// notes 메서드는 설명 표를 봅니다.
func (v *validator) notes() {
	for i, n := range v.run.Notes {
		if n.ID != i {
			v.fail("설명 %d의 id가 %d", i, n.ID)
		}
		v.source(n.Source, "설명 %d", i)
		v.attrs(n.Attrs, fmt.Sprintf("설명 %d", i))
		switch n.Scope.Kind {
		case "run", "topic":
			if n.Scope.ID != nil {
				v.fail("설명 %d의 대상(%s)에 id가 있음", i, n.Scope.Kind)
			}
		case "place":
			if n.Scope.ID == nil {
				v.fail("설명 %d의 대상 장소 id가 없음", i)
			} else {
				v.place(*n.Scope.ID, "설명 %d의 대상", i)
			}
		case "event":
			if n.Scope.ID == nil || *n.Scope.ID < 0 || *n.Scope.ID >= len(v.run.Events) {
				v.fail("설명 %d의 대상 사건이 없음", i)
			}
		default:
			v.fail("설명 %d의 대상 종류 %q는 정해진 값이 아님", i, n.Scope.Kind)
		}
	}
}

// place 메서드는 장소 id 하나가 있는지 봅니다.
func (v *validator) place(id int, format string, a ...any) {
	if id < 0 || id >= len(v.run.Places) {
		v.fail("%s: 없는 장소 %d", fmt.Sprintf(format, a...), id)
	}
}

// source 메서드는 출처 id 하나가 있는지 봅니다.
func (v *validator) source(id string, format string, a ...any) {
	if _, ok := v.run.SourceByID(id); !ok {
		v.fail("%s: 없는 출처 %q", fmt.Sprintf(format, a...), id)
	}
}

// attrs 메서드는 attrs의 키가 모두 출처 id인지 봅니다.
func (v *validator) attrs(a Attrs, where string) {
	for k := range a {
		if _, ok := v.run.SourceByID(k); !ok {
			v.fail("%s의 attrs 키 %q가 출처 id가 아님", where, k)
		}
	}
}

// cycle 메서드는 start에서 up을 따라 올라갈 때 limit번 안에 끝나지 않으면 참을 돌려줍니다.
func (v *validator) cycle(start, limit int, up func(int) *int) bool {
	cur := &start
	for range limit + 1 {
		cur = up(*cur)
		if cur == nil {
			return false
		}
	}
	return true
}
