package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// 이 파일은 기록끼리 견준 갈림입니다(2026-10-15, WORK 실험 5 「실험 5 화면 설계」의 「갈림의 규칙」). decode spring의 출력 여러 개를
// 읽어, 같은 메서드(이름, 서술자)의 마디가 간 길을 모으고, 끊김이 아닌 길이 둘 이상인 메서드를 갈림으로 표시해 각 출력에 덧붙입니다.

// appChild 는 자식 차례에서 앱 코드(lab.*) 호출을 대신하는 이름입니다. 장면마다 앱 메서드가 달라, 이름 그대로 견주면 Spring이 같은
// 길을 갔어도 다른 길로 셉니다(19:50 첫 시험판의 invokeWithinTransaction).
const appChild = "lab.*"

// nodeBranch 는 마디가 그 메서드의 길 가운데 어느 것을 갔는지입니다. Key는 springView.Branches의 키, Alt는 그 안의 길 번호입니다.
type nodeBranch struct {
	Key string `json:"key"`
	Alt int    `json:"alt"`
}

// branchAlt 는 한 메서드의 길 하나입니다. Seq는 자식의 (이름+서술자) 차례(앱 코드는 appChild), Labels는 화면용 이름, Sites는 자식마다
// 부른 줄(처음 본 마디의 값)입니다. End는 마디의 끝난 꼴을 정상(return)과 예외(exception)로 묶은 것이고, LastExc는 마지막 자식이
// 예외로 끝났는지입니다. Cut은 끊김(아래에서 올라온 예외로 뒤가 빠짐), N은 이 길을 간 마디 수, Records는 이 길이 나온 기록입니다.
type branchAlt struct {
	Seq     []string `json:"seq"`
	Labels  []string `json:"labels"`
	Sites   []int    `json:"sites"`
	End     string   `json:"end"`
	LastExc bool     `json:"lastExc"`
	Cut     bool     `json:"cut"`
	N       int      `json:"n"`
	Records []string `json:"records"`
}

// branchInfo 는 한 메서드의 길 모음입니다. Flagged는 끊김이 아닌 길이 둘 이상인지입니다. Common은 끊김이 아닌 길들의 자식 차례가
// 같은 앞부분의 길이이고, Lines는 그 바로 다음에 길마다 처음 달라진 자식을 부른 줄(그 메서드의 소스 줄)입니다. 길들이 자식 차례는
// 같고 끝난 꼴만 다르면 Lines는 비어 있습니다.
type branchInfo struct {
	Fn      string       `json:"fn"`
	Desc    string       `json:"desc"`
	Label   string       `json:"label"`
	Alts    []*branchAlt `json:"alts"`
	Flagged bool         `json:"flagged"`
	Common  int          `json:"common"`
	Lines   []int        `json:"lines"`
}

// namedTree 는 견줄 기록 하나의 이름과 호출 트리 마디입니다.
type namedTree struct {
	Name  string
	Nodes []*callNode
}

// computeBranches 는 기록들의 호출 트리에서 메서드마다 길을 모으고 갈림을 정합니다(이 파일 머리 주석).
//
// 끊김: 마지막 자식이 예외로 끝났고, 자식 차례가 마지막 자식이 예외로 끝나지 않은 다른 길의 앞부분(같음 포함)인 길입니다. 그 마디가
// 판단해서가 아니라 아래에서 올라온 예외 때문에 뒤가 빠진 것입니다. 길이 짧아도 마지막 자식이 정상으로 끝났으면 그 마디의 조건이 뒤를
// 건너뛴 것이므로 끊김이 아닙니다(PgConnection.commit()이 아무것도 보내지 않는 길). 자식 차례가 같아도 끝난 꼴이 다르면 다른 길입니다
// (HikariPool.getConnection(long)의 시간 초과와 0.5초 안의 재빌리기).
//
// Arguments: recs는 견줄 기록들입니다. 앱 코드 마디(lab.*)는 장면마다 달라 모으지 않습니다.
// Returns: "이름+서술자"를 키로 한 길 모음입니다. 마디마다 Branch를 채우고, 갈림인 메서드의 마디에는 Why에 갈림을 더합니다.
func computeBranches(recs []namedTree) map[string]*branchInfo {
	groups := map[string]*branchInfo{}
	sigs := map[string]map[string]int{}
	for _, r := range recs {
		for _, n := range r.Nodes {
			if strings.HasPrefix(n.Fn, "lab.") {
				continue
			}
			key := n.Fn + n.Desc
			g, ok := groups[key]
			if !ok {
				g = &branchInfo{Fn: n.Fn, Desc: n.Desc, Label: n.Label, Lines: []int{}}
				groups[key], sigs[key] = g, map[string]int{}
			}

			// 이 마디의 길: 자식 차례, 부른 줄, 끝난 꼴, 마지막 자식의 끝난 꼴.
			alt := &branchAlt{Seq: []string{}, Labels: []string{}, Sites: []int{}, End: endReturn, Records: []string{}}
			for _, k := range n.Kids {
				kid := r.Nodes[k]
				if strings.HasPrefix(kid.Fn, "lab.") {
					alt.Seq, alt.Labels = append(alt.Seq, appChild), append(alt.Labels, "앱 코드")
				} else {
					alt.Seq, alt.Labels = append(alt.Seq, kid.Fn+kid.Desc), append(alt.Labels, kid.Label)
				}
				alt.Sites = append(alt.Sites, kid.Site)
			}
			if isExceptionEnd(n.End) {
				alt.End = "exception"
			}
			if len(n.Kids) > 0 {
				alt.LastExc = isExceptionEnd(r.Nodes[n.Kids[len(n.Kids)-1]].End)
			}
			sig := fmt.Sprintf("%s#%s#%t", strings.Join(alt.Seq, " "), alt.End, alt.LastExc)
			i, seen := sigs[key][sig]
			if !seen {
				i = len(g.Alts)
				sigs[key][sig] = i
				g.Alts = append(g.Alts, alt)
			}
			a := g.Alts[i]
			a.N++
			if !contains(a.Records, r.Name) {
				a.Records = append(a.Records, r.Name)
			}
			n.Branch = &nodeBranch{Key: key, Alt: i}
		}
	}

	// 길마다 끊김을 가리고, 끊김이 아닌 길이 둘 이상이면 갈림입니다.
	for _, g := range groups {
		var real []*branchAlt
		for _, a := range g.Alts {
			for _, b := range g.Alts {
				if a != b && a.LastExc && !b.LastExc && isPrefix(a.Seq, b.Seq) {
					a.Cut = true
				}
			}
			if !a.Cut {
				real = append(real, a)
			}
		}
		if len(real) < 2 {
			continue
		}
		g.Flagged = true
		g.Common = len(real[0].Seq)
		for _, a := range real[1:] {
			k := 0
			for k < g.Common && k < len(a.Seq) && a.Seq[k] == real[0].Seq[k] {
				k++
			}
			g.Common = k
		}
		for _, a := range real {
			if len(a.Seq) > g.Common && !containsInt(g.Lines, a.Sites[g.Common]) {
				g.Lines = append(g.Lines, a.Sites[g.Common])
			}
		}
		sort.Ints(g.Lines)
	}
	for _, r := range recs {
		for _, n := range r.Nodes {
			if n.Branch != nil && groups[n.Branch.Key].Flagged && !contains(n.Why, whyBranch) {
				n.Why = append(n.Why, whyBranch)
			}
		}
	}
	return groups
}

// isPrefix 는 a가 b의 앞부분(같음 포함)인지 돌려줍니다.
func isPrefix(a, b []string) bool {
	if len(a) > len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// containsInt 는 xs에 x가 있는지 돌려줍니다.
func containsInt(xs []int, x int) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// cmdSpringCompare 는 decode spring의 출력들을 읽어 갈림을 셈하고, 각 출력의 view에 Branches·Compared와 마디의 Branch·Why를
// 더해 outDir에 같은 파일 이름으로 씁니다. 원래 파일은 고치지 않습니다.
//
// Arguments: outDir은 쓸 디렉터리, args는 "이름=경로"(이름은 txlab view의 키와 같은 "장면/실행")입니다.
// Errors: 인자의 꼴이 다르거나, 파일을 읽고 쓰지 못하거나, view가 없는 출력(이 파일보다 앞의 해석기로 만든 것)이면 오류입니다.
func cmdSpringCompare(outDir string, args []string) error {
	if outDir == "" || len(args) == 0 {
		return fmt.Errorf("사용: decode spring-compare -out DIR 이름=경로 ...")
	}
	type input struct {
		path string
		raw  map[string]json.RawMessage
		view springView
	}
	var ins []*input
	var recs []namedTree
	for _, a := range args {
		name, path, ok := strings.Cut(a, "=")
		if !ok || name == "" {
			return fmt.Errorf("%q: 이름=경로 꼴이어야 합니다", a)
		}
		in := &input{path: path}
		if err := readJSON(path, &in.raw); err != nil {
			return err
		}
		v, ok := in.raw["view"]
		if !ok {
			return fmt.Errorf("%s: view가 없습니다(decode spring을 다시 돌려야 함)", path)
		}
		if err := json.Unmarshal(v, &in.view); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		ins = append(ins, in)
		recs = append(recs, namedTree{Name: name, Nodes: in.view.Tree})
	}
	branches := computeBranches(recs)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	var names []string
	for _, r := range recs {
		names = append(names, r.Name)
	}
	for _, in := range ins {
		in.view.Branches, in.view.Compared = branches, names
		b, err := json.Marshal(in.view)
		if err != nil {
			return err
		}
		in.raw["view"] = b
		out, err := json.Marshal(in.raw)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(outDir, filepath.Base(in.path)), append(out, '\n'), 0o644); err != nil {
			return err
		}
	}
	return nil
}
