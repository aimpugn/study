package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Builder 값은 적재기가 run 하나를 조립할 때 쓰는 도구입니다.
//
// id를 차례로 매기고, 같은 장소는 한 번만 만들고, 사건의 깊이를 부모 사슬로 셈하고, 정적 장소의 key를 만듭니다.
// 형식의 규칙을 지키는지는 조립이 끝난 뒤 Validate가 따로 봅니다.
type Builder struct {
	run    *Run
	places map[string]int
}

// NewBuilder 함수는 빈 run을 가진 Builder를 만듭니다. id는 run의 식별자이고 나머지는 run 머리의 칸입니다.
func NewBuilder(id, topic, scene, name string, origin Origin) *Builder {
	return &Builder{
		run: &Run{
			ID: id, Topic: topic, Scene: scene, Name: name, Origin: origin,
			Clocks: []Clock{}, Sources: []Source{}, Coverage: []Coverage{}, Unrecorded: []Unrecorded{},
			Attrs: Attrs{}, Places: []Place{}, Events: []Event{}, Links: []Link{}, Notes: []Note{},
		},
		places: map[string]int{},
	}
}

// Run 메서드는 조립 중인 run을 돌려줍니다. 돌려받은 값을 고치면 조립 중인 run이 바뀝니다.
func (b *Builder) Run() *Run { return b.run }

// AddClock 메서드는 시계 하나를 더합니다.
func (b *Builder) AddClock(c Clock) { b.run.Clocks = append(b.run.Clocks, c) }

// AddSource 메서드는 출처 하나를 더합니다.
func (b *Builder) AddSource(s Source) { b.run.Sources = append(b.run.Sources, s) }

// SetRunAttrs 메서드는 run의 attrs에 출처 source가 낸 값 하나를 넣습니다.
func (b *Builder) SetRunAttrs(source, key string, v any) { setAttr(b.run.Attrs, source, key, v) }

// PlaceSpec 값은 Builder.Place에 넘기는 장소의 내용입니다.
//
// Static이 참이면 기록과 무관하게 정해지는 장소(함수, 프레임 등)이고 key를 받습니다. Version은 그 장소가 속한 소스의 판입니다.
type PlaceSpec struct {
	Parent  *int
	Kind    string
	Name    string
	Loc     *Loc
	Source  string
	Static  bool
	Version string
}

// Place 메서드는 장소 하나의 id를 돌려줍니다. 부모·종류·이름·위치·출처가 같은 장소가 이미 있으면 새로 만들지 않습니다.
//
// 정적 장소의 key는 H(종류, 이름, 소스 경로, 줄, 정적 부모의 key, 소스 판)입니다. 부모가 정적이 아니면 부모 자리는 비워,
// 같은 함수는 어느 프로세스 아래에 있어도 같은 key가 됩니다.
func (b *Builder) Place(s PlaceSpec) int {
	// 같은 장소를 두 번 만들지 않으려고 장소의 내용으로 찾습니다.
	parent := -1
	if s.Parent != nil {
		parent = *s.Parent
	}
	path, line := "", 0
	if s.Loc != nil {
		path, line = s.Loc.Path, s.Loc.Line
	}
	ident := fmt.Sprintf("%d\x00%s\x00%s\x00%s\x00%d\x00%s", parent, s.Kind, s.Name, path, line, s.Source)
	if id, ok := b.places[ident]; ok {
		return id
	}

	// 정적 장소면 기록을 가로질러 같은 장소를 잇는 key를 만듭니다.
	key := ""
	if s.Static {
		parentKey := ""
		if s.Parent != nil {
			parentKey = b.run.Places[*s.Parent].Key
		}
		key = contentKey(s.Kind, s.Name, path, fmt.Sprint(line), parentKey, s.Version)
	}

	id := len(b.run.Places)
	b.run.Places = append(b.run.Places, Place{ID: id, Parent: s.Parent, Kind: s.Kind, Name: s.Name, Key: key, Loc: s.Loc, Source: s.Source})
	b.places[ident] = id
	return id
}

// PlaceAttrs 메서드는 장소 id의 attrs에 출처 source가 낸 값 하나를 넣습니다.
func (b *Builder) PlaceAttrs(id int, source, key string, v any) {
	p := &b.run.Places[id]
	if p.Attrs == nil {
		p.Attrs = Attrs{}
	}
	setAttr(p.Attrs, source, key, v)
}

// AddEvent 메서드는 사건 하나를 더하고 그 id를 돌려줍니다. ID와 Depth는 Builder가 정하고, Count가 0이면 1로 둡니다.
//
// 부모가 아직 없는 사건을 가리키면 panic합니다. 적재기의 순서 잘못이라 기록을 끝까지 만들 이유가 없습니다.
func (b *Builder) AddEvent(e Event) int {
	e.ID = len(b.run.Events)
	if e.Count == 0 {
		e.Count = 1
	}
	// 깊이는 부모 사슬로 셈한 사본이라, 부모의 깊이에 1을 더합니다.
	e.Depth = 0
	if e.Parent != nil {
		if *e.Parent >= e.ID {
			panic(fmt.Sprintf("사건 %d의 부모 %d가 아직 없음", e.ID, *e.Parent))
		}
		e.Depth = b.run.Events[*e.Parent].Depth + 1
	}
	b.run.Events = append(b.run.Events, e)
	return e.ID
}

// EventAttrs 메서드는 사건 id의 attrs에 출처 source가 낸 값 하나를 넣습니다.
func (b *Builder) EventAttrs(id int, source, key string, v any) {
	e := &b.run.Events[id]
	if e.Attrs == nil {
		e.Attrs = Attrs{}
	}
	setAttr(e.Attrs, source, key, v)
}

// AddLink 메서드는 연결 하나를 더하고 그 id를 돌려줍니다.
func (b *Builder) AddLink(src, dst int, kind, basis string) int {
	id := len(b.run.Links)
	b.run.Links = append(b.run.Links, Link{ID: id, Src: src, Dst: dst, Kind: kind, Basis: basis})
	return id
}

// LinkAttrs 메서드는 연결 id의 attrs에 출처 source가 낸 값 하나를 넣습니다.
func (b *Builder) LinkAttrs(id int, source, key string, v any) {
	l := &b.run.Links[id]
	if l.Attrs == nil {
		l.Attrs = Attrs{}
	}
	setAttr(l.Attrs, source, key, v)
}

// AddNote 메서드는 설명 하나를 더하고 그 id를 돌려줍니다.
func (b *Builder) AddNote(n Note) int {
	n.ID = len(b.run.Notes)
	b.run.Notes = append(b.run.Notes, n)
	return n.ID
}

// AddCoverage 메서드는 관찰 범위 하나를 더합니다.
func (b *Builder) AddCoverage(c Coverage) { b.run.Coverage = append(b.run.Coverage, c) }

// AddUnrecorded 메서드는 출처가 원래 기록하지 않는 칸 하나를 적습니다.
func (b *Builder) AddUnrecorded(u Unrecorded) { b.run.Unrecorded = append(b.run.Unrecorded, u) }

// setAttr 함수는 출처별 attrs에 값 하나를 넣습니다. 그 출처의 칸이 없으면 만듭니다.
func setAttr(a Attrs, source, key string, v any) {
	m := a[source]
	if m == nil {
		m = map[string]any{}
		a[source] = m
	}
	m[key] = v
}

// contentKey 함수는 내용에서 정적 장소의 key를 만듭니다. 칸 사이에 0바이트를 넣어 칸 경계가 섞이지 않게 합니다.
func contentKey(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(h[:12])
}

// Ptr 함수는 값의 주소를 돌려줍니다. 선택 칸(*int, *int64)을 채울 때 씁니다.
func Ptr[T any](v T) *T { return &v }
