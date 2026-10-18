package spring

import (
	"fmt"
	"sort"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// treeKeys 는 호출 트리 마디의 열쇠입니다(decode/spring_tree.go 54~75).
var treeKeys = []string{"branch", "depth", "desc", "effects", "end", "fn", "hidden", "hiddenKinds", "id", "kids", "label", "lane", "parent", "seen", "self", "site", "t0", "t1", "why"}

// tree 메서드는 main 스레드의 호출 트리 마디를 해석기 출처의 구간으로 옮기고 포함을 parent로 둡니다.
//
// 장소는 main 스레드 아래 서술자이고, 스택은 [마디, 숨은 칸(안쪽부터), 부모 메서드 안의 부른 줄]입니다. 스택에만 보인 마디(t1 null)는 처음
// 본 시각을 시작으로 두지 않습니다. 마디 id가 차례가 아니거나 부모가 뒤에 있으면 parent로 둘 수 없어 멈춥니다.
func (l *loader) tree(v any) error {
	list, err := load.List(v, "$.view.tree")
	if err != nil {
		return err
	}
	main, ok := l.thread[sceneThread]
	if !ok && len(list) > 0 {
		return fmt.Errorf("트리가 있는데 JFR에 %s 스레드가 없음", sceneThread)
	}
	fns := make([]string, len(list))
	for i, raw := range list {
		where := fmt.Sprintf("$.view.tree[%d]", i)
		n, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(n, where, treeKeys...); err != nil {
			return err
		}
		if id, err := load.Integer(n["id"], where+".id"); err != nil || id != i {
			return fmt.Errorf("%s.id가 차례 %d가 아님", where, i)
		}
		parent, err := load.Integer(n["parent"], where+".parent")
		if err != nil {
			return err
		}
		if parent >= i || parent < -1 {
			return fmt.Errorf("%s.parent %d가 앞의 마디가 아님", where, parent)
		}
		names, err := texts(n, where, "fn", "desc")
		if err != nil {
			return err
		}
		fns[i] = names[0]
		place := l.descriptor(main, names[0], names[1], srcTree, "$.view.tree[].fn", "$.view.tree[].desc")

		// 스택: 숨은 칸은 바깥부터 적혀 있어 뒤집고, 부모가 있으면 부모 메서드 안의 부른 줄이 마지막 칸입니다. 뿌리의 부른 줄은 0입니다.
		stack := []int{place}
		hidden, err := load.List(n["hidden"], where+".hidden")
		if err != nil {
			return err
		}
		for k := len(hidden) - 1; k >= 0; k-- {
			ref, err := load.Text(hidden[k], fmt.Sprintf("%s.hidden[%d]", where, k))
			if err != nil {
				return err
			}
			f, err := l.frame(main, ref, srcTree, "$.view.tree[].hidden[]")
			if err != nil {
				return fmt.Errorf("%s.hidden[%d]: %w", where, k, err)
			}
			stack = append(stack, f)
		}
		site, err := load.Integer(n["site"], where+".site")
		if err != nil {
			return err
		}
		var parentEvent *int
		if parent >= 0 {
			f, err := l.frame(main, fmt.Sprintf("%s:%d", fns[parent], site), srcTree, "$.view.tree[].site")
			if err != nil {
				return fmt.Errorf("%s.site: %w", where, err)
			}
			stack = append(stack, f)
			p := l.node[parent]
			parentEvent = &p
		} else if site != 0 {
			return fmt.Errorf("%s: 뿌리의 site %d가 0이 아님", where, site)
		}

		// 시각: 기록된 마디는 구간, 스택에만 보인 마디는 시작과 끝이 없고 t0·seen을 자손에서 다시 셉니다.
		var start, end *int64
		_, hasSeen := n["seen"]
		if n["t1"] == nil {
			if !hasSeen {
				return fmt.Errorf("%s: 끝이 없는데 seen이 없음", where)
			}
		} else {
			if hasSeen {
				return fmt.Errorf("%s: 끝이 있는데 seen이 있음", where)
			}
			t0, err := l.micros(n["t0"], where+".t0")
			if err != nil {
				return err
			}
			t1, err := l.micros(n["t1"], where+".t1")
			if err != nil {
				return err
			}
			start, end = &t0, &t1
		}
		id := l.Event(atlas.Event{Place: place, Kind: atlas.Span, Name: names[0], StartNS: start, EndNS: end, Parent: parentEvent, Stack: stack, Source: srcTree}, "$.view.tree[]")
		for _, k := range []string{"id", "label", "end", "hiddenKinds", "self", "why"} {
			l.B.EventAttrs(id, srcTree, k, n[k])
		}
		l.node[i] = id
		if err := l.effects(id, place, n, where); err != nil {
			return err
		}
	}
	if len(list) > 0 && l.Once("unrecorded:passed") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcTree, Field: "tree.start_ns", Note: "스택에만 보인 호출(끝난 꼴 passed)은 시작을 기록하지 않습니다. 처음 본 시각과 마지막으로 본 시각은 기록된 자손에서 다시 셉니다"})
	}
	return nil
}

// effects 메서드는 트리 마디 하나의 효과를 종류마다 옮깁니다(설계 메모 5의 4).
//
// autocommit·validate·rollbackOnly는 마디의 자식 순간, return·suspend·resume은 추정 출처의 순간(옮긴 연결이 target, 마디 → contains),
// wire·borrow·park·throw는 마디에서 요청·빌림·JFR 잠듦·JFR 예외로 가는 연결입니다. 효과의 차례가 묶임 → 종류 → 잠듦 → 예외가 아니거나,
// 종류 효과가 두 종류이거나, 가리킬 사건이 정확히 하나가 아니면 내보내기가 차례를 되살릴 수 없어 멈춥니다.
func (l *loader) effects(node, place int, n map[string]any, where string) error {
	list, err := load.List(n["effects"], where+".effects")
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}
	// 빌림과 맞추는 데 쓰는 마디의 구간: 스택에만 보인 마디는 처음 본 시각과 마지막으로 본 시각입니다(callNode.last).
	t0, err := load.Micros(n["t0"], where+".t0")
	if err != nil {
		return err
	}
	lastKey := "t1"
	if n["t1"] == nil {
		lastKey = "seen"
	}
	last, err := load.Micros(n[lastKey], where+"."+lastKey)
	if err != nil {
		return err
	}
	rank, kindOne := -1, ""
	for k, raw := range list {
		ew := fmt.Sprintf("%s.effects[%d]", where, k)
		e, err := load.Object(raw, ew)
		if err != nil {
			return err
		}
		kind, err := load.Text(e["kind"], ew+".kind")
		if err != nil {
			return err
		}
		r, ok := effectRank[kind]
		if !ok {
			return fmt.Errorf("%s: 모르는 효과 %q", ew, kind)
		}
		if r < rank {
			return fmt.Errorf("%s: 효과 %s가 차례(묶임 → 종류 → 잠듦 → 예외)를 어김", ew, kind)
		}
		rank = r
		if r == 1 {
			if kindOne != "" && kindOne != kind {
				return fmt.Errorf("%s: 한 마디에 종류 효과 %s와 %s", ew, kindOne, kind)
			}
			kindOne = kind
		}
		switch kind {
		case "suspend", "resume", "return":
			if err := load.Keys(e, ew, "kind", "conn"); err != nil {
				return err
			}
			var target *int
			if c, ok := e["conn"]; ok {
				name, err := load.Text(c, ew+".conn")
				if err != nil {
					return err
				}
				p, ok := l.conn[name]
				if !ok {
					return fmt.Errorf("%s.conn %s가 연결이 아님", ew, name)
				}
				target = &p
			}
			id := l.Event(atlas.Event{Place: place, Target: target, Kind: atlas.Instant, Name: kind, Source: srcGuess}, "$.view.tree[].effects[].conn")
			l.Link(node, id, "contains", srcGuess, "$.view.tree[].effects[].conn")
		case "autocommit", "validate", "rollbackOnly":
			allowed := []string{"kind"}
			if kind == "autocommit" {
				allowed = append(allowed, "value", "sent")
			}
			if err := load.Keys(e, ew, allowed...); err != nil {
				return err
			}
			id := l.Event(atlas.Event{Place: place, Kind: atlas.Instant, Name: kind, Parent: &node, Source: srcTree}, "$.view.tree[].effects[]")
			for _, key := range []string{"value", "sent"} {
				if val, ok := e[key]; ok {
					l.B.EventAttrs(id, srcTree, key, val)
				}
			}
		case "wire":
			if err := load.Keys(e, ew, "kind", "conn", "seq"); err != nil {
				return err
			}
			name, err := load.Text(e["conn"], ew+".conn")
			if err != nil {
				return err
			}
			seqs, err := load.List(e["seq"], ew+".seq")
			if err != nil || len(seqs) == 0 {
				return fmt.Errorf("%s.seq가 비지 않은 목록이 아님", ew)
			}
			for _, s := range seqs {
				seq, err := load.Integer(s, ew+".seq[]")
				if err != nil {
					return err
				}
				if seq < 1 || seq >= len(l.reqs[name]) {
					return fmt.Errorf("%s: 연결 %s에 요청 %d가 없음", ew, name, seq)
				}
				l.Link(node, l.reqs[name][seq], "wire", srcGuess, "$.view.tree[].effects[].seq")
			}
		case "borrow":
			if err := load.Keys(e, ew, "kind", "conn", "idleMs", "result", "validated"); err != nil {
				return err
			}
			var found []int
			for _, b := range l.borrows {
				if b.thread == sceneThread && b.t0 <= t0 && last <= b.t1 {
					found = append(found, b.id)
				}
			}
			if len(found) != 1 {
				return fmt.Errorf("%s: 마디를 감싸는 main 빌림 %d개", ew, len(found))
			}
			l.Link(node, found[0], "borrow", srcSpring, "$.view.tree[].effects[]")
		case "park":
			if err := load.Keys(e, ew, "kind", "at", "t0", "t1"); err != nil {
				return err
			}
			a, err := load.Micros(e["t0"], ew+".t0")
			if err != nil {
				return err
			}
			b, err := load.Micros(e["t1"], ew+".t1")
			if err != nil {
				return err
			}
			p := l.parks[[2]int64{a, b}]
			if len(p) != 1 {
				return fmt.Errorf("%s: 같은 구간의 main 잠듦 %d개", ew, len(p))
			}
			l.Link(node, p[0], "park", srcSpring, "$.view.tree[].effects[]")
		case "throw":
			if err := load.Keys(e, ew, "kind", "at", "class", "message", "t0"); err != nil {
				return err
			}
			a, err := load.Micros(e["t0"], ew+".t0")
			if err != nil {
				return err
			}
			t := l.throws[a]
			if len(t) != 1 {
				return fmt.Errorf("%s: 같은 시각의 main 예외 %d개", ew, len(t))
			}
			l.Link(node, t[0], "throw", srcSpring, "$.view.tree[].effects[]")
		}
	}
	// 시각 없는 효과는 출처가 둘입니다: 종류 효과는 해석기의 트리 출처, 묶임 효과는 추정 출처의 사건입니다(코드 검토 지적 6).
	if l.Once("unrecorded:effects") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcTree, Field: "effect.start_ns", Note: "autocommit·validate·rollbackOnly 효과는 따로 시각이 없고 그 마디의 구간 안의 일입니다"})
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcGuess, Field: "effect.start_ns", Note: "return·suspend·resume 효과(연결 묶임의 변화)는 따로 시각이 없고 그 마디의 구간 안의 일입니다"})
	}
	return nil
}

// binding 메서드는 main 스레드의 연결 묶임 상태를 추정 출처의 순간으로 옮깁니다. 묶인 연결이 target이고, 상태를 만든 트리 마디와 by 연결입니다.
//
// 첫 상태는 장면 시작의 빈 상태라(맥 시계) 옮기지 않고 다시 셉니다. 첫 상태가 빈 상태가 아니면 다시 셀 수 없어 멈춥니다. 떼어 둔 연결(held)은
// 상태의 차례와 by 마디로 다시 셉니다.
func (l *loader) binding(v any) error {
	list, err := load.List(v, "$.view.binding")
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return fmt.Errorf("$.view.binding이 비었음(첫 상태가 늘 있음)")
	}
	first, err := load.Object(list[0], "$.view.binding[0]")
	if err != nil {
		return err
	}
	held, _ := first["held"].([]any)
	if first["bound"] != "" || held == nil || len(held) != 0 || fmt.Sprint(first["by"]) != "-1" {
		return fmt.Errorf("$.view.binding[0]이 빈 상태가 아님")
	}
	main := l.thread[sceneThread]
	for i, raw := range list[1:] {
		where := fmt.Sprintf("$.view.binding[%d]", i+1)
		b, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(b, where, "bound", "by", "held", "t"); err != nil {
			return err
		}
		bound, err := load.Text(b["bound"], where+".bound")
		if err != nil {
			return err
		}
		var target *int
		if bound != "" {
			p, ok := l.conn[bound]
			if !ok {
				return fmt.Errorf("%s.bound %s가 연결이 아님", where, bound)
			}
			target = &p
		}
		t, err := l.micros(b["t"], where+".t")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: main, Target: target, Kind: atlas.Instant, Name: nameBinding, StartNS: &t, Source: srcGuess}, "$.view.binding[]")
		by, err := load.Integer(b["by"], where+".by")
		if err != nil {
			return err
		}
		switch node, ok := l.node[by]; {
		case ok:
			l.Link(id, node, "by", srcGuess, "$.view.binding[].by")
		case by != -1:
			return fmt.Errorf("%s.by %d의 트리 마디가 없음", where, by)
		}
	}
	return nil
}

// slots 메서드는 풀의 연결 하나를 main 스레드가 쓴 구간을 추정 출처의 구간으로 옮깁니다. 쓴 연결이 target이고, 장면 안에 돌려주지 않은
// 슬롯(t1 0)은 끝을 비웁니다.
func (l *loader) slots(v any) error {
	list, err := load.List(v, "$.view.slots")
	if err != nil {
		return err
	}
	main := l.thread[sceneThread]
	for i, raw := range list {
		where := fmt.Sprintf("$.view.slots[%d]", i)
		s, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(s, where, "conn", "t0", "t1"); err != nil {
			return err
		}
		name, err := load.Text(s["conn"], where+".conn")
		if err != nil {
			return err
		}
		p, ok := l.conn[name]
		if !ok {
			return fmt.Errorf("%s.conn %s가 연결이 아님", where, name)
		}
		t0, err := l.micros(s["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := load.Micros(s["t1"], where+".t1")
		if err != nil {
			return err
		}
		end := &t1
		if t1 == 0 {
			end = nil
		}
		l.Event(atlas.Event{Place: main, Target: &p, Kind: atlas.Span, Name: nameSlot, StartNS: &t0, EndNS: end, Source: srcGuess}, "$.view.slots[]")
	}
	return nil
}

// finish 메서드는 원본이 뜻을 가르지 못하는 꼴을 run의 ambiguity note로, 원래 기록하지 않는 칸을 미기록으로, 중계기의 관찰 범위를 더합니다.
func (l *loader) finish() {
	note := func(text string) {
		l.Source(atlas.Source{ID: srcNote, Tool: "pgsql 주제의 Spring 적재기(topics/pgsql/spring)", Nature: atlas.Human})
		l.B.AddNote(atlas.Note{Scope: atlas.Scope{Kind: "run"}, Kind: "ambiguity", Source: srcNote, Text: text})
	}
	if l.xipNull > 0 {
		note(fmt.Sprintf("스냅숏 %d개의 xip가 null입니다. 해석기는 xcnt만큼 xip를 붙이므로(decode/scenario.go 824~827행) null은 진행 중인 트랜잭션 0개이거나 xcnt를 기록하지 못한 것이고, 원본이 둘을 가르지 않아 값 그대로 옮겼습니다.", l.xipNull))
	}
	if l.noConn > 0 {
		note(fmt.Sprintf("연결이 빈 글인 빌림 %d개는 target이 없습니다. 해석기는 시간 초과로 끝난 빌림에 연결을 붙이지 않고(decode/spring_jfr.go 454~458), 정상으로 끝났어도 100 ms 안에 앱 메시지가 없으면 빈 글로 둡니다. 원본의 빈 글만으로는 둘을 가르지 못해 result와 함께 읽어야 합니다.", l.noConn))
	}
	if l.zero > 0 {
		note(fmt.Sprintf("시각 %d개가 0입니다. 해석기의 rel은 기록하지 않은 시각(0)도 0으로 돌려주므로(decode/spring.go 408~411) 시작 순간과 값 없음을 가르지 못해 0 그대로 옮겼습니다.", l.zero))
	}
	if n, frames := l.overloadFrames(); frames > 0 {
		note(fmt.Sprintf("자바 프레임 글(메서드:줄)에는 서술자가 없어, 서술자가 둘 이상인 메서드(오버로드) %d개를 가리키는 스택 칸 %d개는 어느 오버로드인지 가르지 않고 메서드 아래 프레임에 두었습니다.", n, frames))
	}
	for _, u := range []atlas.Unrecorded{
		{Source: srcJFR, Field: "frame.loc.path", Note: "JFR 프레임은 클래스·메서드·줄만 적고 소스 파일 경로가 없어, 자바 프레임의 위치는 경로가 빈 줄입니다"},
		{Source: srcJFR, Field: "place.key.version", Note: "앱과 라이브러리의 판이 해석 결과에 없어 자바 메서드·서술자·프레임의 key에 판을 넣지 않았습니다"},
		{Source: srcJFR, Field: "coverage", Note: "jdk.MethodTrace 필터(spring.json의 trace)가 해석 결과에 없습니다"},
		{Source: srcRec, Field: "coverage", Note: "어떤 함수에 탐침을 걸었는지 화면 데이터에 남지 않았습니다"},
		{Source: srcWireCnt, Field: "message.attrs", Note: "컨테이너 쪽 중계기의 메시지 칸(SQL, 태그 등)은 해석 결과에 없습니다. 해석기는 두 쪽의 방향·번호·종류·길이가 같음만 확인했습니다(decode/spring.go 204~220)"},
	} {
		l.B.AddUnrecorded(u)
	}
	// 중계기는 앱의 모든 TCP 접속을 받아 넘기며 메시지마다 적습니다(txlab/spring.go 211~236).
	conns := make([]int, 0, len(l.conn))
	for _, p := range l.conn {
		conns = append(conns, p)
	}
	sort.Ints(conns)
	for _, src := range []string{srcWireMac, srcWireCnt} {
		l.B.AddCoverage(atlas.Coverage{Source: src, What: "연결 위 메시지", Places: conns, Completeness: "complete",
			Basis: "파일: 중계기가 앱의 모든 접속을 받아 넘기며 메시지마다 적음(txlab/spring.go 211~236)"})
		l.Origins[atlas.RecordRef{Table: "coverage", ID: len(l.B.Run().Coverage) - 1}] = "$.conns[].messages[]"
	}
	// JFR의 예외와 잠듦은 남긴 것이 일부입니다(결정 검토 지적 9). 메서드 추적의 필터는 해석 결과에 없어 미기록입니다.
	for _, c := range []struct {
		what, basis, group string
	}{
		{"예외(jdk.JavaExceptionThrow)", "파일: 장면 시작(start 멈춤 자리의 go) 뒤의 것만 남기고 잡음 클래스(NoSuchMethodError, NoSuchFieldError, ClassNotFoundException)는 뺌(decode/spring_jfr.go 24~31·270~282). 뺀 수는 run attrs spring.java.dropped", "$.java.throws[]"},
		{"잠듦(jdk.ThreadPark)", "파일: 1 ms보다 긴 잠듦만 기록하고(txlab/spring.go 248 jdk.ThreadPark#threshold=1ms) 스택에 추적 대상 메서드가 있는 것만 남김(decode/spring_jfr.go 299~313)", "$.java.parks[]"},
	} {
		l.B.AddCoverage(atlas.Coverage{Source: srcJFR, What: c.what, Completeness: "partial", Basis: c.basis})
		l.Origins[atlas.RecordRef{Table: "coverage", ID: len(l.B.Run().Coverage) - 1}] = c.group
	}
}

// overloadFrames 메서드는 서술자가 둘 이상인 자바 메서드의 수와, 그 메서드 아래 프레임을 가리키는 스택 칸의 수를 돌려줍니다.
func (l *loader) overloadFrames() (int, int) {
	r := l.B.Run()
	descs := map[int]int{}
	for _, p := range r.Places {
		if p.Kind == "descriptor" {
			descs[*p.Parent]++
		}
	}
	methods, frames := 0, 0
	for _, n := range descs {
		if n > 1 {
			methods++
		}
	}
	for _, e := range r.Events {
		for _, s := range e.Stack {
			if p := r.Places[s]; p.Kind == "frame" && p.Parent != nil && r.Places[*p.Parent].Kind == "method" && descs[*p.Parent] > 1 {
				frames++
			}
		}
	}
	return methods, frames
}
