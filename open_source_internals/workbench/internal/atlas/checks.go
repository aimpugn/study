package atlas

import (
	"fmt"
	"sort"
	"strings"
)

// RecordRef 값은 run 안의 행 하나를 가리킵니다. Table은 run·place·event·link·note·coverage 가운데 하나입니다.
type RecordRef struct {
	Table string
	ID    int
}

// Origins 값은 적재기가 만든 행마다 그 행이 원본의 어느 경로 무리에서 왔는지 적습니다(예: "$.conns[].messages[]").
// 형식의 일부가 아니라 검사에만 쓰는 곁 기록입니다. 주제의 구성도처럼 원본 밖에서 온 행은 "topic:"으로 시작합니다.
type Origins map[RecordRef]string

// Role 값은 원본 경로 하나에 대한 판정 기준입니다.
//
// Role과 Prov는 재고 조사의 표시 파일이 형식 설계 전에 붙인 역할(장소·연결·시각·사건·값·출처·발췌·이름·모음)과
// 출처(관찰·도출·이름)이고, Decision은 대응표가 정한 목적지(구조 칸 이름 또는 예외 부류 X1~X7)입니다.
type Role struct {
	Role     string
	Prov     string
	Decision string
}

// Roles 값은 경로마다의 판정 기준입니다. 키는 표시 파일의 경로 그대로입니다.
type Roles map[string]Role

// structuralRoles 는 구조 칸으로 가야 하는 역할입니다.
var structuralRoles = map[string]bool{"장소": true, "연결": true, "시각": true, "사건": true}

// ScanStructure 함수는 구조가 attrs에 숨지 않았는지 봅니다(설계 메모 9의 검사 3). 어긴 곳을 모두 돌려줍니다.
//
// 행마다 그 행의 경로 무리와 attrs 맨 윗단 키를 이어 붙인 경로를 Roles에서 찾습니다. 키는 둘 중 하나여야 합니다:
// 역할이 구조 아닌 것(값·출처·발췌·이름·모음), 또는 역할은 구조지만 대응표가 예외 부류(X로 시작)로 정한 것.
// 역할이 구조인데 예외가 아니거나, Roles에 없는 키면 어긴 것입니다. 경로 무리가 없는 행의 attrs도 어긴 것입니다.
// 키의 역할이 구조가 아니어도 그 아래 경로(뒤에 "."이나 "["가 이어지는 경로)에 예외 아닌 구조 역할이 있으면 어긴 것입니다.
// 모음 키에 목록을 통째로 두면 그 안의 장소·사건이 attrs에 숨기 때문입니다.
func ScanStructure(r *Run, origins Origins, roles Roles) []error {
	var errs []error
	// hidden 함수는 경로 path 아래에서 예외 아닌 구조 역할을 받은 경로를 하나 돌려줍니다. 없으면 빈 문자열입니다.
	paths := sortedKeys(roles)
	hidden := func(path string) string {
		for _, q := range paths {
			if len(q) <= len(path) || q[:len(path)] != path || (q[len(path)] != '.' && q[len(path)] != '[') {
				continue
			}
			if role := roles[q]; structuralRoles[role.Role] && !strings.HasPrefix(role.Decision, "X") {
				return q
			}
		}
		return ""
	}
	check := func(ref RecordRef, a Attrs) {
		if len(a) == 0 {
			return
		}
		group, ok := origins[ref]
		if !ok {
			errs = append(errs, fmt.Errorf("%s %d: attrs가 있는데 경로 무리가 없음", ref.Table, ref.ID))
			return
		}
		for _, src := range sortedKeys(a) {
			for _, k := range sortedKeys(a[src]) {
				path := group + "." + k
				role, ok := roles[path]
				switch {
				case !ok:
					errs = append(errs, fmt.Errorf("%s %d: attrs[%s].%s의 경로 %s가 표시 파일에 없음", ref.Table, ref.ID, src, k, path))
				case structuralRoles[role.Role] && !strings.HasPrefix(role.Decision, "X"):
					errs = append(errs, fmt.Errorf("%s %d: attrs[%s].%s는 역할 %s(결정 %s)라 구조 칸에 있어야 함", ref.Table, ref.ID, src, k, role.Role, role.Decision))
				default:
					if q := hidden(path); q != "" {
						errs = append(errs, fmt.Errorf("%s %d: attrs[%s].%s 아래의 %s는 역할 %s(결정 %s)라 구조 칸에 있어야 함", ref.Table, ref.ID, src, k, q, roles[q].Role, roles[q].Decision))
					}
				}
			}
		}
	}
	check(RecordRef{"run", 0}, r.Attrs)
	for _, p := range r.Places {
		check(RecordRef{"place", p.ID}, p.Attrs)
	}
	for _, e := range r.Events {
		check(RecordRef{"event", e.ID}, e.Attrs)
	}
	for _, l := range r.Links {
		check(RecordRef{"link", l.ID}, l.Attrs)
	}
	for _, n := range r.Notes {
		check(RecordRef{"note", n.ID}, n.Attrs)
	}
	return errs
}

// natureOf 는 표시 파일의 출처마다 허용하는 출처 성격입니다.
var natureOf = map[string][]string{
	"관찰": {Observed},
	"도출": {Derived, Inferred},
	"이름": {Human},
}

// CheckNature 함수는 행의 출처 성격이 판정 기준과 맞는지 봅니다(설계 메모 9의 검사 4). 어긴 곳을 모두 돌려줍니다.
//
// 장소·사건은 Source, 연결은 Basis, 관찰 범위(RecordRef의 Table "coverage", ID는 차례)는 Source의 성격을 봅니다. 기대 성격은 그 행의 경로 무리에 붙은 표시 파일의 출처로 정하고
// (관찰 → observed, 도출 → derived 또는 inferred, 이름 → human), inferred에 든 경로 무리는 반드시 inferred여야 합니다.
// "topic:"으로 시작하는 경로 무리는 주제가 정한 것이라 human이어야 하고, "rule:"로 시작하는 경로 무리는 해석기가 정한 규칙으로
// 만든 것(요청 창 묶음 등, 설계 메모 7)이라 derived여야 합니다. 경로 무리가 없거나 Roles에 없으면 어긴 것입니다.
func CheckNature(r *Run, origins Origins, roles Roles, inferred map[string]bool) []error {
	var errs []error
	check := func(ref RecordRef, sourceID string) {
		group, ok := origins[ref]
		if !ok {
			errs = append(errs, fmt.Errorf("%s %d: 경로 무리가 없음", ref.Table, ref.ID))
			return
		}
		s, ok := r.SourceByID(sourceID)
		if !ok {
			errs = append(errs, fmt.Errorf("%s %d: 출처 %q가 없음", ref.Table, ref.ID, sourceID))
			return
		}
		var want []string
		switch {
		case inferred[group]:
			want = []string{Inferred}
		case strings.HasPrefix(group, "topic:"):
			want = []string{Human}
		case strings.HasPrefix(group, "rule:"):
			want = []string{Derived}
		default:
			role, ok := roles[group]
			if !ok {
				errs = append(errs, fmt.Errorf("%s %d: 경로 무리 %s가 표시 파일에 없음", ref.Table, ref.ID, group))
				return
			}
			want = natureOf[role.Prov]
			if want == nil {
				errs = append(errs, fmt.Errorf("%s %d: 경로 무리 %s의 표시 출처 %q를 모름", ref.Table, ref.ID, group, role.Prov))
				return
			}
		}
		for _, w := range want {
			if s.Nature == w {
				return
			}
		}
		errs = append(errs, fmt.Errorf("%s %d(%s): 출처 %q의 성격 %s, 기대 %s", ref.Table, ref.ID, group, s.ID, s.Nature, strings.Join(want, "|")))
	}
	for _, p := range r.Places {
		check(RecordRef{"place", p.ID}, p.Source)
	}
	for _, e := range r.Events {
		check(RecordRef{"event", e.ID}, e.Source)
	}
	for _, l := range r.Links {
		check(RecordRef{"link", l.ID}, l.Basis)
	}
	// 짐작한 감시 목록처럼 관찰 범위도 추정일 수 있어 같은 규칙으로 봅니다.
	for i, c := range r.Coverage {
		check(RecordRef{"coverage", i}, c.Source)
	}
	return errs
}

// sortedKeys 함수는 지도의 키를 이름순으로 돌려줍니다. 검사 결과의 차례를 늘 같게 하려고 씁니다.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
