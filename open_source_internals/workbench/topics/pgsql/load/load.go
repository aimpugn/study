// Package load 는 pgsql 주제의 적재기들이 함께 쓰는 조립 도구입니다.
//
// 화면 파일에 박힌 데이터 literal 읽기, 원본 JSON 값 꺼내기, 그리고 장소를 내용으로 찾아 한 번만 만드는 장소 장부가 있습니다.
// 같은 함수는 어느 적재기에서 와도 같은 규칙(설계 메모 2의 3·9)으로 장소가 되도록, 장소 규칙은 이 패키지 한 곳이 가집니다.
package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
)

// Literal 함수는 화면 파일에서 "const <name> = " 로 시작하는 줄의 나머지(끝의 ";" 뺌)를 돌려줍니다.
// 그런 줄이 없으면 오류를 돌려줍니다.
func Literal(html, name string) (string, error) {
	prefix := "const " + name + " = "
	for _, line := range strings.Split(html, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimLeft(line, " \t"), prefix); ok {
			return strings.TrimSuffix(strings.TrimRight(rest, " \t\r"), ";"), nil
		}
	}
	return "", fmt.Errorf("%s literal 줄이 없음", name)
}

// Decode 함수는 JSON 값 하나를 숫자를 글자 그대로(json.Number) 두고 읽습니다. 뒤에 다른 값이 더 있으면 오류입니다.
func Decode(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("literal 뒤에 다른 값이 있음")
	}
	return v, nil
}

// Canonical 함수는 원본과 되돌린 값을 같은 꼴로 견주려고 JSON으로 썼다가 숫자를 글자 그대로 두고 다시 읽습니다.
// 쓰거나 읽지 못하면 오류를 돌려줍니다.
func Canonical(v any) (any, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return Decode(strings.TrimSpace(buf.String()))
}

// Object 함수는 JSON 객체를 꺼냅니다. 객체가 아니면 where를 담은 오류입니다.
func Object(v any, where string) (map[string]any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s가 객체가 아님", where)
	}
	return m, nil
}

// List 함수는 JSON 목록을 꺼냅니다. 목록이 아니면 where를 담은 오류입니다.
func List(v any, where string) ([]any, error) {
	l, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%s가 목록이 아님", where)
	}
	return l, nil
}

// Text 함수는 JSON 글을 꺼냅니다. 글이 아니면 where를 담은 오류입니다.
func Text(v any, where string) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s가 글이 아님", where)
	}
	return s, nil
}

// Keys 함수는 객체의 열쇠가 모두 allowed 안에 있는지 봅니다. 모르는 열쇠를 조용히 버리지 않기 위해서입니다.
func Keys(m map[string]any, where string, allowed ...string) error {
	for k := range m {
		ok := false
		for _, a := range allowed {
			if k == a {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("%s에 모르는 열쇠 %q", where, k)
		}
	}
	return nil
}

// Micros 함수는 원본의 마이크로초 수를 정수 나노초로 옮깁니다(형식 규칙 6). 수가 아니거나 되돌린 값이 다르면 오류입니다.
func Micros(v any, where string) (int64, error) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s가 수가 아님", where)
	}
	ns, err := atlas.MicrosToNanos(n)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	return ns, nil
}

// Integer 함수는 원본의 정수를 읽습니다. 정수가 아니면 오류입니다.
func Integer(v any, where string) (int, error) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s가 수가 아님", where)
	}
	i, err := strconv.Atoi(string(n))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	return i, nil
}

// ParseAt 함수는 "경로:줄"을 읽습니다. 줄은 양의 정수여야 합니다.
func ParseAt(s string) (string, int, error) {
	i := strings.LastIndexByte(s, ':')
	if i <= 0 {
		return "", 0, fmt.Errorf("위치 %q", s)
	}
	n, err := strconv.Atoi(s[i+1:])
	if err != nil || n <= 0 {
		return "", 0, fmt.Errorf("위치 %q의 줄", s)
	}
	return s[:i], n, nil
}

// UnknownAt 은 기호 풀이가 줄을 찾지 못한 스택 칸의 위치 글입니다(gdb BT와 함수 기록의 스택이 같게 씀).
const UnknownAt = "?"

// LocText 함수는 위치를 원본의 꼴("경로:줄")로 씁니다. 위치가 없으면(줄을 모르는 프레임) UnknownAt입니다.
func LocText(l *atlas.Loc) string {
	if l == nil {
		return UnknownAt
	}
	return l.Path + ":" + strconv.Itoa(l.Line)
}

// Loader 값은 run 하나를 조립하는 동안의 장부입니다.
//
// 장소는 내용 열쇠로 찾아 한 번만 만들고, 처음 만든 출처가 남습니다(설계 메모 5의 2: 같은 종류·위치·부모의 장소는 출처와
// 상관없이 같이 씀). Origins는 검사에만 쓰는 곁 기록이고, Version은 정적 장소의 key에 넣는 소스 판입니다.
type Loader struct {
	B       *atlas.Builder
	Origins atlas.Origins
	Version string
	index   map[string]int
	byName  map[string][]int
	added   map[string]bool
}

// NewLoader 함수는 빈 장부를 만듭니다. run의 경로 무리는 "$"로 적습니다.
func NewLoader(b *atlas.Builder, version string) *Loader {
	l := &Loader{B: b, Origins: atlas.Origins{}, Version: version, index: map[string]int{}, byName: map[string][]int{}, added: map[string]bool{}}
	l.Origins[atlas.RecordRef{Table: "run", ID: 0}] = "$"
	return l
}

// Once 메서드는 key를 처음 볼 때만 참을 돌려줍니다. 출처나 미기록 칸을 한 번만 더할 때 씁니다.
func (l *Loader) Once(key string) bool {
	if l.added[key] {
		return false
	}
	l.added[key] = true
	return true
}

// Source 메서드는 출처를 처음 쓸 때 한 번만 더합니다.
func (l *Loader) Source(s atlas.Source) {
	if l.Once("source:" + s.ID) {
		l.B.AddSource(s)
	}
}

// Event 메서드는 사건을 더하고 그 경로 무리를 적은 뒤 id를 돌려줍니다.
func (l *Loader) Event(e atlas.Event, group string) int {
	id := l.B.AddEvent(e)
	l.Origins[atlas.RecordRef{Table: "event", ID: id}] = group
	return id
}

// Link 메서드는 연결을 더하고 그 경로 무리를 적은 뒤 id를 돌려줍니다.
func (l *Loader) Link(src, dst int, kind, basis, group string) int {
	id := l.B.AddLink(src, dst, kind, basis)
	l.Origins[atlas.RecordRef{Table: "link", ID: id}] = group
	return id
}

// Lookup 메서드는 내용 열쇠 key의 장소가 이미 있으면 그 id를 돌려줍니다.
func (l *Loader) Lookup(key string) (int, bool) {
	id, ok := l.index[key]
	return id, ok
}

// Place 메서드는 내용 열쇠 key로 장소를 찾고, 없으면 spec으로 만들어 그 경로 무리를 적습니다.
// 출처는 열쇠에 넣지 않습니다(처음 만든 출처가 남음).
func (l *Loader) Place(key string, spec atlas.PlaceSpec, group string) int {
	if id, ok := l.index[key]; ok {
		return id
	}
	id := l.B.Place(spec)
	l.index[key] = id
	l.Origins[atlas.RecordRef{Table: "place", ID: id}] = group
	return id
}

// Process 메서드는 pid의 프로세스 장소를 돌려줍니다. 이미 있으면(역할 아래 등) 그 장소를 씁니다.
func (l *Loader) Process(pid, src, group string) int {
	return l.Place("process\x00"+pid, atlas.PlaceSpec{Kind: "process", Name: pid, Source: src}, group)
}

// Function 메서드는 프로세스 아래 함수 장소를 돌려줍니다. path가 비면 위치를 모르는 함수입니다.
func (l *Loader) Function(proc int, fn, path, src, group string) int {
	key := fmt.Sprintf("function\x00%d\x00%s\x00%s", proc, fn, path)
	_, existed := l.index[key]
	var loc *atlas.Loc
	if path != "" {
		loc = &atlas.Loc{Path: path}
	}
	id := l.Place(key, atlas.PlaceSpec{Parent: &proc, Kind: "function", Name: fn, Loc: loc, Source: src, Static: true, Version: l.Version}, group)
	if !existed {
		nk := fmt.Sprintf("%d\x00%s", proc, fn)
		l.byName[nk] = append(l.byName[nk], id)
	}
	return id
}

// Named 메서드는 프로세스 아래에서 이름이 fn인 함수 장소를 만든 차례로 돌려줍니다.
func (l *Loader) Named(proc int, fn string) []int {
	return l.byName[fmt.Sprintf("%d\x00%s", proc, fn)]
}

// Frame 메서드는 함수 fn 안의 줄 at("경로:줄")의 프레임 장소를 돌려줍니다. at을 읽지 못하면(UnknownAt 포함) 오류입니다.
// 줄을 모르는 칸은 함수의 경로를 다른 기록에서 찾아야 하므로 부른 쪽이 함수 장소를 정해 UnknownFrame으로 둡니다.
func (l *Loader) Frame(proc int, fn, at, src, fnGroup, atGroup string) (int, error) {
	path, line, err := ParseAt(at)
	if err != nil {
		return 0, err
	}
	f := l.Function(proc, fn, path, src, fnGroup)
	return l.Place(fmt.Sprintf("frame\x00%d\x00%d", f, line),
		atlas.PlaceSpec{Parent: &f, Kind: "frame", Name: fn, Loc: &atlas.Loc{Path: path, Line: line}, Source: src, Static: true, Version: l.Version}, atGroup), nil
}

// UnknownFrame 메서드는 함수 장소 f 아래에 줄을 모르는 프레임 장소를 돌려줍니다. 같은 함수가 위치 있는 장소와 없는 장소로
// 갈리지 않도록, f는 같은 프로세스의 다른 기록에서 그 함수의 경로를 찾아 정한 장소여야 합니다(트랜잭션 실험 적재기의 callee).
func (l *Loader) UnknownFrame(f int, fn, src, group string) int {
	return l.Place(fmt.Sprintf("frame\x00%d\x00?", f), atlas.PlaceSpec{Parent: &f, Kind: "frame", Name: fn, Source: src, Static: true, Version: l.Version}, group)
}

// Site 메서드는 함수 이름 없이 경로와 줄만 아는 코드 자리 장소를 프로세스 아래에 돌려줍니다. at을 읽지 못하면 오류입니다.
func (l *Loader) Site(proc int, at, src, group string) (int, error) {
	path, line, err := ParseAt(at)
	if err != nil {
		return 0, err
	}
	return l.Place(fmt.Sprintf("site\x00%d\x00%s\x00%d", proc, path, line),
		atlas.PlaceSpec{Parent: &proc, Kind: "site", Loc: &atlas.Loc{Path: path, Line: line}, Source: src, Static: true, Version: l.Version}, group), nil
}
