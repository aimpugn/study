package atlas

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Selector 값은 재추출에서 고를 사건의 조건과 함께 낼 값입니다. 빈 조건은 모두를 받습니다.
//
// 주제는 화면이 보이는 사실(단계, 연결 위 메시지, 기다림, 결과 값)마다 Selector 하나를 정해 넘깁니다.
// 플랫폼은 구조 칸(출처, 이름, 종류, 장소의 종류)으로만 고르고, Attrs에 적힌 키만 attrs에서 꺼냅니다.
type Selector struct {
	Sources    []string `json:"sources,omitempty"`
	Names      []string `json:"names,omitempty"`
	NamePrefix string   `json:"name_prefix,omitempty"`
	Kinds      []string `json:"kinds,omitempty"`
	PlaceKinds []string `json:"place_kinds,omitempty"`
	// Require는 "출처 id.키" 꼴의 목록이고, 사건에 그 값이 모두 있어야 고릅니다(같은 출처의 이동과 메시지를 가를 때).
	Require []string `json:"require,omitempty"`
	// Attrs는 "출처 id.키" 꼴의 목록입니다. 사건에 그 값이 없으면 결과의 Attrs에서 그 키만 빠집니다.
	Attrs []string `json:"attrs,omitempty"`
}

// Row 값은 재추출한 사건 하나입니다. Nature는 그 사건의 출처 성격이라, 화면이 추정을 관측처럼 그리지 않게 합니다.
type Row struct {
	Event   int            `json:"event"`
	Name    string         `json:"name"`
	Kind    string         `json:"kind"`
	Place   string         `json:"place"`
	From    string         `json:"from,omitempty"`
	Target  string         `json:"target,omitempty"`
	StartNS *int64         `json:"start_ns"`
	EndNS   *int64         `json:"end_ns"`
	DurNS   *int64         `json:"dur_ns,omitempty"`
	Count   int            `json:"count"`
	Depth   int            `json:"depth"`
	Stack   []string       `json:"stack,omitempty"`
	Nature  string         `json:"nature"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

// Select 함수는 조건에 맞는 사건을 id 차례로 돌려줍니다. 같은 출처 안에서 id 차례는 기록 차례입니다.
//
// 장소·대상·스택은 장소의 이름으로 풀어 냅니다. 스택 칸은 "이름@경로:줄" 꼴입니다.
func Select(r *Run, s Selector) []Row {
	var out []Row
	for _, e := range r.Events {
		if !s.match(r, e) {
			continue
		}
		src, _ := r.SourceByID(e.Source)
		row := Row{Event: e.ID, Name: e.Name, Kind: e.Kind, Place: placeLabel(r, e.Place),
			StartNS: e.StartNS, EndNS: e.EndNS, DurNS: e.DurNS, Count: e.Count, Depth: e.Depth, Nature: src.Nature}
		if e.From != nil {
			row.From = placeLabel(r, *e.From)
		}
		if e.Target != nil {
			row.Target = placeLabel(r, *e.Target)
		}
		for _, p := range e.Stack {
			row.Stack = append(row.Stack, frameLabel(r, p))
		}
		// 주제가 고른 attrs 키만 꺼냅니다.
		for _, want := range s.Attrs {
			src, key, _ := strings.Cut(want, ".")
			if v, ok := e.Attrs[src][key]; ok {
				if row.Attrs == nil {
					row.Attrs = map[string]any{}
				}
				row.Attrs[want] = v
			}
		}
		out = append(out, row)
	}
	return out
}

// Count 함수는 조건에 맞는 사건의 수를 돌려줍니다. group은 묶은 횟수(count)가 아니라 한 행으로 셉니다.
func Count(r *Run, s Selector) int { return len(Select(r, s)) }

// Linked 함수는 kind가 같은 연결의 양 끝 사건을 출발 사건의 id, 같으면 도착 사건의 id 차례로 돌려줍니다. 근거의 성격도 함께 냅니다.
//
// 연결 표의 차례에는 뜻이 없어(형식 규칙 7) 사건의 id 차례로 놓습니다(앞선 적재기 코드 검토 지적 13).
func Linked(r *Run, kind string) [][2]Row {
	var out [][2]Row
	all := Select(r, Selector{})
	links := make([]Link, 0, len(r.Links))
	for _, l := range r.Links {
		if l.Kind == kind {
			links = append(links, l)
		}
	}
	sort.Slice(links, func(i, j int) bool {
		if links[i].Src != links[j].Src {
			return links[i].Src < links[j].Src
		}
		return links[i].Dst < links[j].Dst
	})
	for _, l := range links {
		a, b := all[l.Src], all[l.Dst]
		basis, _ := r.SourceByID(l.Basis)
		a.Nature, b.Nature = basis.Nature, basis.Nature
		out = append(out, [2]Row{a, b})
	}
	return out
}

// FactRule 값은 주제가 정한 재추출 규칙 하나입니다. Select·Notes·Linked 가운데 꼭 하나만 씁니다.
//
// 주제는 화면이 보이는 사실마다 규칙 하나를 주제 디렉터리의 자료 파일에 적고, 플랫폼은 그 규칙만 받아 뽑습니다(설계 메모 9의 검사 5).
type FactRule struct {
	Fact   string    `json:"fact"`
	Select *Selector `json:"select,omitempty"`
	Notes  string    `json:"notes,omitempty"`
	Linked string    `json:"linked,omitempty"`
}

// Fact 값은 규칙 하나로 뽑은 사실입니다. 규칙의 종류에 맞는 칸 하나만 찹니다.
type Fact struct {
	Fact  string   `json:"fact"`
	Rows  []Row    `json:"rows,omitempty"`
	Notes []string `json:"notes,omitempty"`
	Pairs [][2]Row `json:"pairs,omitempty"`
}

// Extract 함수는 규칙마다 사실을 뽑아 규칙 차례로 돌려줍니다.
//
// 규칙 하나가 Select·Notes·Linked를 하나도 쓰지 않거나 둘 이상 쓰면, 무엇을 뽑을지 모르므로 오류를 돌려줍니다.
func Extract(r *Run, rules []FactRule) ([]Fact, error) {
	out := make([]Fact, 0, len(rules))
	for _, rule := range rules {
		n := 0
		if rule.Select != nil {
			n++
		}
		if rule.Notes != "" {
			n++
		}
		if rule.Linked != "" {
			n++
		}
		if n != 1 {
			return nil, fmt.Errorf("규칙 %q: select·notes·linked 가운데 하나만 써야 함(%d개)", rule.Fact, n)
		}
		f := Fact{Fact: rule.Fact}
		switch {
		case rule.Select != nil:
			f.Rows = Select(r, *rule.Select)
		case rule.Notes != "":
			f.Notes = NotesOf(r, rule.Notes)
		default:
			f.Pairs = Linked(r, rule.Linked)
		}
		out = append(out, f)
	}
	return out, nil
}

// NotesOf 함수는 kind가 같은 설명의 글을 id 차례로 돌려줍니다.
func NotesOf(r *Run, kind string) []string {
	var out []string
	for _, n := range r.Notes {
		if n.Kind == kind {
			out = append(out, n.Text)
		}
	}
	return out
}

// match 메서드는 사건 하나가 조건에 맞는지 봅니다.
func (s Selector) match(r *Run, e Event) bool {
	if len(s.Sources) > 0 && !contains(s.Sources, e.Source) {
		return false
	}
	if len(s.Names) > 0 && !contains(s.Names, e.Name) {
		return false
	}
	if s.NamePrefix != "" && !strings.HasPrefix(e.Name, s.NamePrefix) {
		return false
	}
	if len(s.Kinds) > 0 && !contains(s.Kinds, e.Kind) {
		return false
	}
	if len(s.PlaceKinds) > 0 && !contains(s.PlaceKinds, r.Places[e.Place].Kind) {
		return false
	}
	for _, want := range s.Require {
		src, key, _ := strings.Cut(want, ".")
		if _, ok := e.Attrs[src][key]; !ok {
			return false
		}
	}
	return true
}

// placeLabel 함수는 장소를 "종류 이름"(이름이 없으면 "종류")으로 쓰고, 위치가 있으면 "@경로:줄"(줄이 없으면 "@경로")을 붙입니다.
// 화면이 단계마다 보이는 소스 위치가 장소에 있기 때문입니다.
func placeLabel(r *Run, id int) string {
	p := r.Places[id]
	label := p.Kind
	if p.Name != "" {
		label += " " + p.Name
	}
	if p.Loc != nil {
		label += "@" + p.Loc.Path
		if p.Loc.Line != 0 {
			label += ":" + strconv.Itoa(p.Loc.Line)
		}
	}
	return label
}

// frameLabel 함수는 스택 칸의 장소를 "이름@경로:줄"로 씁니다. 위치가 없으면 이름만 씁니다.
func frameLabel(r *Run, id int) string {
	p := r.Places[id]
	if p.Loc == nil {
		return p.Name
	}
	return p.Name + "@" + p.Loc.Path + ":" + strconv.Itoa(p.Loc.Line)
}

// contains 함수는 목록에 값이 있는지 봅니다.
func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
