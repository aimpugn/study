// Package txlab 는 트랜잭션 실험 화면(concept-5~8)에 박힌 txlab 기록의 해석 결과를 공통 형식으로 옮기고 되돌립니다.
//
// 옮기는 규칙은 설계 메모 5의 3입니다. 클라이언트 쪽은 txlab 장소 아래 세션이고, 서버 쪽은 역할 아래 프로세스와 그 아래
// 함수·프레임·코드 자리입니다. 세션과 백엔드의 짝은 '단계 contains 요청'의 두 장소로 남고 따로 두지 않습니다.
// 원본의 모든 열쇠는 아래 목록으로 받고, 모르는 열쇠나 규칙으로 되살릴 수 없는 꼴이 있으면 조용히 버리지 않고 멈춥니다.
package txlab

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// 해석기가 정한 규칙으로 만든 연결과 주제가 정한 장소의 경로 무리입니다(설계 메모 7, 검사 4).
const (
	// RequestWindow 는 단계 → 요청, 요청 → 사건·스냅숏·줄 판정의 창 묶음입니다(입구가 단계 구간 안인 호출, 창 안의 기록).
	RequestWindow = "rule:request-window"
	// RequestMessages 는 요청 → 메시지의 묶음입니다(요청 입구 직전의 recvfrom, 창 끝 + 5 ms 안의 sendto).
	RequestMessages = "rule:request-messages"
	// WalBeforeKill 은 kill → kill 전에 끝난 pg_wal 쓰기 가운데 쓴 끝이 가장 먼 것의 연결입니다(walExtentBefore).
	WalBeforeKill = "rule:wal-before-kill"
	// ClientPlace 는 기록을 남긴 txlab 프로그램의 장소입니다. 원본에 행으로 있지 않고 주제가 세웁니다.
	ClientPlace = "topic:client"
)

// InferredGroups 함수는 추정 목록(설계 메모 7)에 든 경로 무리를 돌려줍니다: 주소나 값으로 맞춘 힌트 변화 설명과 그 연결.
func InferredGroups() map[string]bool {
	return map[string]bool{"$.hintChanges[]": true, "$.hintChanges[].at": true, "$.hintChanges[].step": true}
}

// 출처 id입니다. 사건마다 그 사건을 남긴 도구가 출처이고, 시계는 출처가 가리킵니다.
const (
	srcTxlab  = "txlab"
	srcRec    = "recorder"
	srcSpan   = "recorder-span"
	srcShift  = "uprobe-ftrace"
	srcStrace = "strace"
	srcLog    = "serverlog"
	srcServer = "server"
	srcTx     = "tx"
	srcHint   = "tx-hint"
	srcSrc    = "src"
	srcTopic  = "topic"
	srcNote   = "pgsql-topic"
)

// 시계 id입니다(설계 메모 3의 표).
const (
	clockRec   = "recorder"
	clockShift = "ftrace-shifted"
	clockLog   = "serverlog"
)

// 사건 이름 가운데 원본의 kind가 아니라 주제가 붙인 것입니다. 원본 목록의 열쇠 이름을 따랐습니다.
const (
	nameStep    = "step"
	nameRequest = "exec_simple_query"
	nameCapture = "capture"
	nameKill    = "kill"
	namePause   = "pause"
	nameProcess = "process"
	nameBystand = "bystander"
	nameCheck   = "check"
	nameLine    = "line"
	namePath    = "path"
	nameInner   = "inner"
	nameHint    = "hint_change"
)

// straceKinds 는 요청 안 사건 가운데 strace가 남긴 종류이고(scenario.go 529~535행), serverStrace 는 서버 사건 가운데
// strace가 남긴 종류입니다(scenario_server.go 425~446행). 나머지는 함수 기록기·uprobe의 사건입니다.
var (
	straceKinds  = map[string]bool{"kill": true, "io": true, "sleep_syscall": true}
	serverStrace = map[string]bool{"wal_io": true, "signal": true, "exit": true}
	walWrites    = map[string]bool{"pwrite64": true, "pwritev": true}
)

// topKeys 는 기록 한 벌의 맨 위 열쇠입니다(재고 조사 3.3.1절).
var topKeys = []string{"bystanders", "captures", "clock", "hintChanges", "joins", "kills", "pauses", "scenario", "server", "sessions", "siteRepair", "src", "steps", "tables", "title", "tupleAddrs"}

// Result 값은 적재 결과입니다. Origins는 검사에만 쓰는 곁 기록이고 형식에는 들어가지 않습니다.
type Result struct {
	Run     *atlas.Run
	Origins atlas.Origins
}

// loader 값은 기록 한 벌을 조립하는 동안의 상태입니다.
type loader struct {
	*load.Loader
	client  int
	session map[string]int       // 세션 이름 → 세션 장소
	backend map[string]string    // 세션 이름 → 그 백엔드 pid(sessions[].pid)
	ownerOf map[string]string    // 백엔드 pid → 세션 이름
	proc    map[string]int       // pid → 프로세스 장소
	step    map[int]int          // 단계 번호 n → 단계 사건
	used    map[string]bool      // 단계를 가진 세션
	kills   []int                // kill 사건(차례대로)
	killAt  map[int]int          // kill 사건 → beforeStep
	writes  map[int64][]walWrite // 끝 시각 → pg_wal 쓰기 사건
	xipNull int
}

// ReadRuns 함수는 트랜잭션 실험 화면(concept-5~8.html)에 박힌 RUNS literal(기록 열쇠 → 해석 결과)을 읽습니다.
// 줄을 찾지 못하거나 JSON으로 읽지 못하면 오류를 돌려줍니다.
func ReadRuns(html string) (map[string]any, error) {
	lit, err := load.Literal(html, "RUNS")
	if err != nil {
		return nil, err
	}
	v, err := load.Decode(lit)
	if err != nil {
		return nil, err
	}
	return load.Object(v, "RUNS")
}

// Import 함수는 기록 한 벌(RUNS의 열쇠 key와 그 값 v)을 run 하나로 옮깁니다.
//
// version은 소스 판(workbench.json의 sourceRoot)이고 정적 장소의 key에 들어갑니다. 원본에 모르는 열쇠가 있거나 규칙(메모 5의 3)으로
// 되살릴 수 없는 꼴(세션마다 백엔드가 하나가 아님, 이을 단계가 없는 kill, 끝 시각이 맞는 WAL 쓰기가 하나가 아님 등)이면 오류를 돌려줍니다.
func Import(key string, v map[string]any, version string) (Result, error) {
	if err := load.Keys(v, "$", topKeys...); err != nil {
		return Result{}, err
	}
	l := &loader{
		Loader: load.NewLoader(atlas.NewBuilder("tx/"+key, "pgsql", "tx", key,
			atlas.Origin{Clock: clockRec, Label: "첫 단계를 보낸 순간(steps[0].sentNs)"}), version),
		session: map[string]int{}, backend: map[string]string{}, ownerOf: map[string]string{}, proc: map[string]int{},
		step: map[int]int{}, used: map[string]bool{}, killAt: map[int]int{}, writes: map[int64][]walWrite{},
	}
	l.clocks()
	l.runAttrs(v)
	server, err := load.Object(v["server"], "$.server")
	if err != nil {
		return Result{}, err
	}
	if err := load.Keys(server, "$.server", "procs", "events", "spans"); err != nil {
		return Result{}, err
	}
	steps, err := load.List(v["steps"], "$.steps")
	if err != nil {
		return Result{}, err
	}

	// 장소를 먼저 세우고, 부른 함수의 위치를 스택에서 모은 뒤, 사건을 원본 목록의 차례로 옮깁니다. 힌트 변화는 추정이라 맨 뒤입니다.
	if err := l.places(v["sessions"], server["procs"]); err != nil {
		return Result{}, err
	}
	if err := l.scanPaths(steps, server); err != nil {
		return Result{}, err
	}
	for _, do := range []func() error{
		func() error { return l.steps(steps) },
		func() error { return l.serverEvents(server["events"]) },
		func() error { return l.serverSpans(server["spans"]) },
		func() error { return l.killsOf(v["kills"]) },
		func() error { return l.pauses(v["pauses"]) },
		func() error { return l.captures(v["captures"]) },
		func() error { return l.bystanders(v["bystanders"]) },
		func() error { return l.hints(v["hintChanges"]) },
	} {
		if err := do(); err != nil {
			return Result{}, fmt.Errorf("%s: %w", key, err)
		}
	}

	// 세션의 pid는 내보낼 때 그 세션 단계의 요청 장소로 다시 세므로, 단계가 없는 세션은 되살릴 수 없습니다.
	for name := range l.session {
		if !l.used[name] {
			return Result{}, fmt.Errorf("%s: 세션 %s에 단계가 없어 pid를 다시 셀 수 없음", key, name)
		}
	}
	l.notes()
	return Result{Run: l.B.Run(), Origins: l.Origins}, nil
}

// clocks 메서드는 시계 셋과 그 관계, 늘 쓰는 출처 둘(txlab, 해석기)을 더합니다.
func (l *loader) clocks() {
	l.Source(atlas.Source{ID: srcTxlab, Tool: "txlab(시나리오 실행기: 단계를 보내고 받은 시각, 관찰, kill, 멈춤)", Clock: clockRec, Nature: atlas.Observed})
	l.Source(atlas.Source{ID: srcTx, Tool: "해석기 decode/scenario.go·scenario_server.go·scenario_check.go", Clock: clockRec, Nature: atlas.Derived})
	l.B.AddClock(atlas.Clock{ID: clockRec, PrecisionNS: 1,
		Note: "해석기가 txlab·함수 기록·strace의 시각을 첫 단계를 보낸 순간부터 센 µs로 옮긴 축(decode/scenario.go의 rel)"})
	l.B.AddClock(atlas.Clock{ID: clockShift, PrecisionNS: 1, Relation: &atlas.ClockRelation{To: clockRec, Basis: srcTx,
		Note: "ftrace 시각에서 짝지은 사건으로 잰 중앙값 차이(run attrs tx.clock.medianOffsetUs)를 뺀 좌표. 짝이 없는 사건의 오차는 남지 않음"}})
	l.B.AddClock(atlas.Clock{ID: clockLog, PrecisionNS: 1_000_000, Relation: &atlas.ClockRelation{To: clockRec, Known: true, LoNS: -999_999, HiNS: 0, Basis: srcTx,
		Note: "server.log의 %m(밀리초까지)을 해석기가 같은 축에 놓은 값이라 1 ms 안쪽으로 내려 있음(scenario_server.go 118행)"}})
}

// runAttrs 메서드는 기록 전체에 붙는 값을 출처마다 run attrs에 둡니다.
func (l *loader) runAttrs(v map[string]any) {
	for _, k := range []string{"title", "scenario", "tables"} {
		l.B.SetRunAttrs(srcTxlab, k, v[k])
	}
	for _, k := range []string{"tupleAddrs", "joins", "clock", "siteRepair"} {
		l.B.SetRunAttrs(srcTx, k, v[k])
	}
	l.Source(atlas.Source{ID: srcSrc, Tool: "PostgreSQL 소스 발췌(이미지 pg-internals:18.6)", Nature: atlas.Observed})
	l.B.SetRunAttrs(srcSrc, "src", v["src"])
}

// places 메서드는 txlab 장소와 세션, 역할 아래의 서버 프로세스와 그 수명 구간을 세웁니다.
//
// 프로세스의 session은 저장하지 않고 다시 세므로, 원본의 값이 sessions[].pid와 맞는지 여기서 확인합니다.
func (l *loader) places(sv, pv any) error {
	l.Source(atlas.Source{ID: srcTopic, Tool: "pgsql 주제(트랜잭션 실험 적재기 topics/pgsql/txlab)", Nature: atlas.Human})
	l.client = l.Place("client\x00txlab", atlas.PlaceSpec{Kind: "client", Name: "txlab", Source: srcTopic}, ClientPlace)
	sessions, err := load.List(sv, "$.sessions")
	if err != nil {
		return err
	}
	for i, raw := range sessions {
		where := fmt.Sprintf("$.sessions[%d]", i)
		s, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(s, where, "name", "pid"); err != nil {
			return err
		}
		name, err := load.Text(s["name"], where+".name")
		if err != nil {
			return err
		}
		pid, ok := s["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		if _, dup := l.session[name]; dup {
			return fmt.Errorf("세션 %s가 두 번", name)
		}
		if _, dup := l.ownerOf[string(pid)]; dup {
			return fmt.Errorf("pid %s가 세션 둘의 백엔드", pid)
		}
		l.session[name] = l.Place("session\x00"+name, atlas.PlaceSpec{Parent: &l.client, Kind: "session", Name: name, Source: srcTxlab}, "$.sessions[]")
		l.backend[name], l.ownerOf[string(pid)] = string(pid), name
	}

	procs, err := load.List(pv, "$.server.procs")
	if err != nil {
		return err
	}
	l.Source(atlas.Source{ID: srcServer, Tool: "함수 기록기와 strace가 본 서버 프로세스(처음·마지막으로 본 기록, 관찰용 접속은 strace, 끝은 strace의 끝 줄)", Clock: clockRec, Nature: atlas.Observed})
	for i, raw := range procs {
		where := fmt.Sprintf("$.server.procs[%d]", i)
		p, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(p, where, "exit", "pid", "role", "t0", "t1", "session"); err != nil {
			return err
		}
		pid, ok := p["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		role, err := load.Text(p["role"], where+".role")
		if err != nil {
			return err
		}
		if _, dup := l.proc[string(pid)]; dup {
			return fmt.Errorf("pid %s가 프로세스 표에 두 번", pid)
		}
		owner, isBackend := l.ownerOf[string(pid)]
		if s, ok := p["session"]; ok != isBackend || (ok && s != owner) {
			return fmt.Errorf("%s.session %v, 다시 센 값 %q", where, s, owner)
		}
		rolePlace := l.Place("role\x00"+role, atlas.PlaceSpec{Kind: "role", Name: role, Source: srcTx}, "$.server.procs[].role")
		place := l.Place("process\x00"+string(pid), atlas.PlaceSpec{Parent: &rolePlace, Kind: "process", Name: string(pid), Source: srcServer}, "$.server.procs[]")
		l.proc[string(pid)] = place
		t0, err := load.Micros(p["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := load.Micros(p["t1"], where+".t1")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: place, Kind: atlas.Span, Name: nameProcess, StartNS: &t0, EndNS: &t1, Source: srcServer}, "$.server.procs[]")
		l.B.EventAttrs(id, srcServer, "exit", p["exit"])
	}
	for name, pid := range l.backend {
		if _, ok := l.proc[pid]; !ok {
			return fmt.Errorf("세션 %s의 백엔드 pid %s가 프로세스 표에 없음", name, pid)
		}
	}
	return nil
}

// scanPaths 메서드는 스택 칸과 서버 함수 구간에서 프로세스마다 함수 이름의 소스 경로를 장부(load.Loader.NotePath)에 모읍니다.
//
// 줄 판정 안의 호출은 부른 함수의 이름과 부른 줄만 남겨, 부른 함수의 위치는 같은 프로세스에서 그 이름이 나온 경로로 찾습니다.
// 그래야 같은 함수가 위치 있는 장소와 위치 없는 장소 둘로 갈리지 않습니다. 프로세스 표에 없는 pid는 뒤에서 오류가 되므로 건너뜁니다.
func (l *loader) scanPaths(steps []any, server map[string]any) error {
	add := func(pid string, raw any) {
		proc, ok := l.proc[pid]
		f, isObj := raw.(map[string]any)
		if !ok || !isObj {
			return
		}
		fn, ok1 := f["fn"].(string)
		at, ok2 := f["at"].(string)
		if ok1 && ok2 {
			l.NotePath(proc, fn, at)
		}
	}
	each := func(v any, fn func(any)) {
		if list, ok := v.([]any); ok {
			for _, x := range list {
				fn(x)
			}
		}
	}
	for _, raw := range steps {
		s, _ := raw.(map[string]any)
		name, _ := s["session"].(string)
		pid := l.backend[name]
		each(s["requests"], func(r any) {
			req, _ := r.(map[string]any)
			for _, list := range []string{"events", "snapshots", "checks"} {
				each(req[list], func(e any) {
					ev, _ := e.(map[string]any)
					each(ev["stack"], func(f any) { add(pid, f) })
				})
			}
		})
	}
	each(server["events"], func(e any) {
		ev, _ := e.(map[string]any)
		pid, _ := ev["pid"].(json.Number)
		each(ev["stack"], func(f any) { add(string(pid), f) })
	})
	each(server["spans"], func(e any) {
		sp, _ := e.(map[string]any)
		pid, _ := sp["pid"].(json.Number)
		add(string(pid), sp)
	})
	return nil
}

// stack 메서드는 백엔드 pid의 원본 스택을 프레임 장소 목록으로 뒤집어 돌려줍니다(load.Loader.Stack). 0번이 가장 안쪽 프레임입니다.
func (l *loader) stack(pid string, v any, src, group, where string) ([]int, error) {
	return l.Stack(l.proc[pid], v, src, group, where)
}

// timed 함수는 원본의 시작 t와 끝 t1(있으면)으로 사건의 종류와 시각을 정합니다. 끝이 있으면 구간, 없으면 순간입니다.
func timed(m map[string]any, start, end, where string) (string, *int64, *int64, error) {
	t0, err := load.Micros(m[start], where+"."+start)
	if err != nil {
		return "", nil, nil, err
	}
	if _, ok := m[end]; !ok {
		return atlas.Instant, &t0, nil, nil
	}
	t1, err := load.Micros(m[end], where+"."+end)
	if err != nil {
		return "", nil, nil, err
	}
	return atlas.Span, &t0, &t1, nil
}

// steps 메서드는 단계를 세션의 구간으로, 그 안의 요청과 요청이 품은 메시지·사건·스냅숏·줄 판정을 옮깁니다.
func (l *loader) steps(steps []any) error {
	l.Source(atlas.Source{ID: srcRec, Tool: "함수 기록기·uprobe(해석기 decode/scenario.go가 읽음)", Clock: clockRec, Nature: atlas.Observed})
	l.Source(atlas.Source{ID: srcStrace, Tool: "strace(시스템 콜, 그 데이터로 푼 연결 위 메시지, 프로세스의 끝 줄)", Clock: clockRec, Nature: atlas.Observed})
	l.B.AddUnrecorded(atlas.Unrecorded{Source: srcRec, Field: "coverage", Note: "어떤 함수에 탐침을 걸었는지 화면 데이터에 남지 않았습니다"})
	for i, raw := range steps {
		where := fmt.Sprintf("$.steps[%d]", i)
		s, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(s, where, "n", "session", "sql", "t0", "t1", "output", "requests", "check", "blocked"); err != nil {
			return err
		}
		n, err := load.Integer(s["n"], where+".n")
		if err != nil {
			return err
		}
		if _, dup := l.step[n]; dup {
			return fmt.Errorf("단계 번호 %d가 두 번", n)
		}
		name, err := load.Text(s["session"], where+".session")
		if err != nil {
			return err
		}
		sess, ok := l.session[name]
		if !ok {
			return fmt.Errorf("%s.session %s가 sessions에 없음", where, name)
		}
		l.used[name] = true
		t0, err := load.Micros(s["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := load.Micros(s["t1"], where+".t1")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: sess, Kind: atlas.Span, Name: nameStep, StartNS: &t0, EndNS: &t1, Source: srcTxlab}, "$.steps[]")
		l.step[n] = id
		for _, k := range []string{"n", "sql", "output", "blocked"} {
			if val, ok := s[k]; ok {
				l.B.EventAttrs(id, srcTxlab, k, val)
			}
		}
		if val, ok := s["check"]; ok {
			l.B.EventAttrs(id, srcTx, "check", val)
		}
		reqs, err := load.List(s["requests"], where+".requests")
		if err != nil {
			return err
		}
		for k, r := range reqs {
			if err := l.request(id, t1, name, r, fmt.Sprintf("%s.requests[%d]", where, k)); err != nil {
				return err
			}
		}
	}
	return nil
}

// request 메서드는 요청 하나(기록기가 본 exec_simple_query 호출)와 그 안의 것을 옮깁니다.
//
// 오류로 끝난 요청(aborted)은 반환 사건이 없어 해석기가 끝을 단계의 끝으로 채웠으므로 끝을 비웁니다. 내보내기가 '끝이 빔'으로 aborted와 t1을 다시 셉니다.
func (l *loader) request(step int, stepEnd int64, sessName string, raw any, where string) error {
	r, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(r, where, "aborted", "catalogSnapshots", "checks", "events", "in", "out", "snapshots", "sql", "t0", "t1"); err != nil {
		return err
	}
	pid := l.backend[sessName]
	backend := l.proc[pid]
	t0, err := load.Micros(r["t0"], where+".t0")
	if err != nil {
		return err
	}
	t1, err := load.Micros(r["t1"], where+".t1")
	if err != nil {
		return err
	}
	end := &t1
	if a, ok := r["aborted"]; ok {
		if a != true || t1 != stepEnd {
			return fmt.Errorf("%s: aborted %v, 끝 %d, 단계 끝 %d", where, a, t1, stepEnd)
		}
		end = nil
	}
	id := l.Event(atlas.Event{Place: backend, Kind: atlas.Span, Name: nameRequest, StartNS: &t0, EndNS: end, Source: srcRec}, "$.steps[].requests[]")
	l.B.EventAttrs(id, srcRec, "sql", r["sql"])
	// 카탈로그 스냅숏 수는 해석기가 요청 안의 GetSnapshotData 가운데 대상이 CatalogSnapshotData인 것을 기호로 가려 센 값이라
	// 해석기의 값입니다(decode/scenario.go 812~822, Spring 코드 검토 지적 3b).
	l.B.EventAttrs(id, srcTx, "catalogSnapshots", r["catalogSnapshots"])
	l.Link(step, id, "contains", srcTx, RequestWindow)

	// 받은 메시지는 세션에서 백엔드로(from), 보낸 메시지는 백엔드에서 세션으로(target) 갑니다. out의 null은 보낸 메시지가 없다는 뜻입니다.
	sess := l.session[sessName]
	in, err := load.List(r["in"], where+".in")
	if err != nil {
		return err
	}
	for k, m := range in {
		if err := l.message(id, backend, &sess, nil, m, fmt.Sprintf("%s.in[%d]", where, k), "$.steps[].requests[].in[]"); err != nil {
			return err
		}
	}
	if r["out"] != nil {
		out, err := load.List(r["out"], where+".out")
		if err != nil || len(out) == 0 {
			return fmt.Errorf("%s.out이 null도 비지 않은 목록도 아님", where)
		}
		for k, m := range out {
			if err := l.message(id, backend, nil, &sess, m, fmt.Sprintf("%s.out[%d]", where, k), "$.steps[].requests[].out[]"); err != nil {
				return err
			}
		}
	}

	evs, err := load.List(r["events"], where+".events")
	if err != nil {
		return err
	}
	for k, e := range evs {
		if err := l.requestEvent(id, pid, e, fmt.Sprintf("%s.events[%d]", where, k)); err != nil {
			return err
		}
	}
	snaps, err := load.List(r["snapshots"], where+".snapshots")
	if err != nil {
		return err
	}
	for k, s := range snaps {
		if err := l.snapshot(id, pid, s, fmt.Sprintf("%s.snapshots[%d]", where, k)); err != nil {
			return err
		}
	}
	checks, err := load.List(r["checks"], where+".checks")
	if err != nil {
		return err
	}
	for k, c := range checks {
		if err := l.check(id, pid, c, fmt.Sprintf("%s.checks[%d]", where, k)); err != nil {
			return err
		}
	}
	return nil
}

// message 메서드는 연결 위 메시지 하나를 백엔드의 순간으로 옮깁니다. 시각은 남지 않았습니다.
func (l *loader) message(req, backend int, from, target *int, raw any, where, group string) error {
	m, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(m, where, "type", "len", "text"); err != nil {
		return err
	}
	typ, err := load.Text(m["type"], where+".type")
	if err != nil {
		return err
	}
	if l.Once("unrecorded:message") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcStrace, Field: "message.start_ns", Note: "메시지는 recvfrom·sendto의 데이터를 푼 것이라 따로 시각이 없습니다"})
	}
	id := l.Event(atlas.Event{Place: backend, From: from, Target: target, Kind: atlas.Instant, Name: typ, Source: srcStrace}, group)
	l.B.EventAttrs(id, srcStrace, "len", m["len"])
	if t, ok := m["text"]; ok {
		l.B.EventAttrs(id, srcStrace, "text", t)
	}
	l.Link(req, id, "contains", srcTx, RequestMessages)
	return nil
}

// requestEvent 메서드는 요청 창 안의 사건 하나를 옮깁니다. 출처는 종류로 정하고, pg_wal 쓰기는 kill의 wal 연결 후보로 적어 둡니다.
func (l *loader) requestEvent(req int, pid string, raw any, where string) error {
	e, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(e, where, "kind", "t", "t1", "f", "stack"); err != nil {
		return err
	}
	name, err := load.Text(e["kind"], where+".kind")
	if err != nil {
		return err
	}
	src := srcRec
	if straceKinds[name] {
		src = srcStrace
	}
	group := "$.steps[].requests[].events[]"
	stack, err := l.stack(pid, e["stack"], src, group, where)
	if err != nil {
		return err
	}
	place := l.proc[pid]
	if stack != nil {
		place = stack[0]
	}
	kind, t0, t1, err := timed(e, "t", "t1", where)
	if err != nil {
		return err
	}
	id := l.Event(atlas.Event{Place: place, Kind: kind, Name: name, StartNS: t0, EndNS: t1, Stack: stack, Source: src}, group)
	if f, ok := e["f"]; ok {
		l.B.EventAttrs(id, src, "f", f)
		l.noteWrite(id, name == "io", f, t1)
	}
	l.Link(req, id, "contains", srcTx, RequestWindow)
	return nil
}

// walWrite 값은 pg_wal 쓰기 사건 하나와 그 쓰기가 닿은 WAL 위치(lsn1, 해석기가 쓴 범위의 끝)입니다.
type walWrite struct {
	id   int
	lsn1 any
}

// noteWrite 메서드는 pg_wal 파일에 쓴 사건(pwrite64·pwritev)을 끝 시각으로 적어 둡니다. kill의 walWrittenT가 이 끝 시각과
// 맞는 것을 찾습니다. 해석기(walExtentBefore)처럼 pg_wal 아래 파일에 쓴 것만 셉니다.
func (l *loader) noteWrite(id int, isIO bool, f any, end *int64) {
	m, ok := f.(map[string]any)
	if !ok || !isIO || end == nil {
		return
	}
	call, _ := m["call"].(string)
	file, _ := m["file"].(string)
	if walWrites[call] && strings.HasPrefix(file, "pg_wal/") {
		l.writes[*end] = append(l.writes[*end], walWrite{id: id, lsn1: m["lsn1"]})
	}
}

// snapshot 메서드는 스냅숏 하나를 순간으로 옮깁니다. xip의 null은 원본이 0개와 기록 안 됨을 가르지 않아 note로 남깁니다.
func (l *loader) snapshot(req int, pid string, raw any, where string) error {
	s, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(s, where, "curcid", "kind", "stack", "t", "xip", "xmax", "xmin"); err != nil {
		return err
	}
	name, err := load.Text(s["kind"], where+".kind")
	if err != nil {
		return err
	}
	group := "$.steps[].requests[].snapshots[]"
	stack, err := l.stack(pid, s["stack"], srcRec, group, where)
	if err != nil {
		return err
	}
	place := l.proc[pid]
	if stack != nil {
		place = stack[0]
	}
	t, err := load.Micros(s["t"], where+".t")
	if err != nil {
		return err
	}
	id := l.Event(atlas.Event{Place: place, Kind: atlas.Instant, Name: name, StartNS: &t, Stack: stack, Source: srcRec}, group)
	for _, k := range []string{"curcid", "xip", "xmax", "xmin"} {
		if _, ok := s[k]; !ok {
			return fmt.Errorf("%s.%s 없음", where, k)
		}
		l.B.EventAttrs(id, srcRec, k, s[k])
	}
	if s["xip"] == nil {
		l.xipNull++
	}
	l.Link(req, id, "contains", srcTx, RequestWindow)
	return nil
}

// check 메서드는 줄 판정 하나와 그 안의 줄·판정 경로·경로 안의 호출을 옮깁니다. 셋은 같은 출처라 parent로 잇는 자식입니다.
func (l *loader) check(req int, pid string, raw any, where string) error {
	c, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(c, where, "ctid", "lines", "mask", "path", "snapXmax", "snapXmin", "stack", "t", "table", "tup", "visible", "xmax", "xmin"); err != nil {
		return err
	}
	group := "$.steps[].requests[].checks[]"
	stack, err := l.stack(pid, c["stack"], srcRec, group, where)
	if err != nil {
		return err
	}
	if stack == nil {
		return fmt.Errorf("%s.stack이 없음", where)
	}
	t, err := load.Micros(c["t"], where+".t")
	if err != nil {
		return err
	}
	id := l.Event(atlas.Event{Place: stack[0], Kind: atlas.Instant, Name: nameCheck, StartNS: &t, Stack: stack, Source: srcRec}, group)
	for _, k := range []string{"ctid", "tup", "table", "mask", "xmin", "xmax", "snapXmin", "snapXmax"} {
		if val, ok := c[k]; ok {
			l.B.EventAttrs(id, srcRec, k, val)
		} else if k != "tup" {
			return fmt.Errorf("%s.%s 없음", where, k)
		}
	}
	l.B.EventAttrs(id, srcTx, "visible", c["visible"])
	l.Link(req, id, "contains", srcTx, RequestWindow)
	if l.Once("unrecorded:calls") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcRec, Field: "check.calls.start_ns", Note: "줄 판정 안의 줄(line)·판정 경로(path)·경로 안의 호출(inner)은 따로 시각이 없습니다"})
	}

	lines, err := load.List(c["lines"], where+".lines")
	if err != nil {
		return err
	}
	for k, raw := range lines {
		if _, err := l.call(id, pid, nameLine, raw, fmt.Sprintf("%s.lines[%d]", where, k), group+".lines[]", "fn", "at"); err != nil {
			return err
		}
	}
	if c["path"] == nil {
		return nil
	}
	path, err := load.List(c["path"], where+".path")
	if err != nil || len(path) == 0 {
		return fmt.Errorf("%s.path가 null도 비지 않은 목록도 아님", where)
	}
	for k, raw := range path {
		pw := fmt.Sprintf("%s.path[%d]", where, k)
		p, err := l.call(id, pid, namePath, raw, pw, group+".path[]", "fn", "at", "arg", "ret", "inner")
		if err != nil {
			return err
		}
		m := raw.(map[string]any)
		for _, a := range []string{"arg", "ret"} {
			if val, ok := m[a]; ok {
				l.B.EventAttrs(p, srcRec, a, val)
			}
		}
		if iv, ok := m["inner"]; ok {
			inner, err := load.List(iv, pw+".inner")
			if err != nil || len(inner) == 0 {
				return fmt.Errorf("%s.inner가 비지 않은 목록이 아님", pw)
			}
			for j, raw := range inner {
				if _, err := l.call(p, pid, nameInner, raw, fmt.Sprintf("%s.inner[%d]", pw, j), group+".path[].inner[]", "fn", "at"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// call 메서드는 줄 판정 안의 호출 하나(부른 함수 fn과 부른 줄 at)를 parent의 자식 순간으로 옮깁니다.
// 장소는 부른 함수이고 스택은 [그 함수, 부른 줄의 코드 자리]입니다(여정 concept-3의 단계와 같은 꼴).
func (l *loader) call(parent int, pid, name string, raw any, where, group string, allowed ...string) (int, error) {
	m, err := load.Object(raw, where)
	if err != nil {
		return 0, err
	}
	if err := load.Keys(m, where, allowed...); err != nil {
		return 0, err
	}
	fn, err := load.Text(m["fn"], where+".fn")
	if err != nil {
		return 0, err
	}
	at, err := load.Text(m["at"], where+".at")
	if err != nil {
		return 0, err
	}
	place, err := l.Callee(l.proc[pid], fn, srcRec, group+".fn")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	site, err := l.Site(l.proc[pid], at, srcRec, group+".at")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	return l.Event(atlas.Event{Place: place, Kind: atlas.Instant, Name: name, Parent: &parent, Stack: []int{place, site}, Source: srcRec}, group), nil
}

// serverEvents 메서드는 관찰 대상이 아닌 서버 프로세스의 사건을 옮깁니다. 출처는 종류로, ftrace 시각을 옮겨 쓴 것은 그 시계의 출처로 정합니다.
//
// f.clock은 출처의 시계로 다시 세므로 저장하지 않습니다. 원본 객체를 고치면 되돌림 비교가 고친 원본과 견주게 되어, 사본에서 뺍니다.
func (l *loader) serverEvents(v any) error {
	evs, err := load.List(v, "$.server.events")
	if err != nil {
		return err
	}
	for i, raw := range evs {
		where := fmt.Sprintf("$.server.events[%d]", i)
		e, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(e, where, "f", "kind", "pid", "stack", "t", "t1"); err != nil {
			return err
		}
		name, err := load.Text(e["kind"], where+".kind")
		if err != nil {
			return err
		}
		pid, ok := e["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		proc, ok := l.proc[string(pid)]
		if !ok {
			return fmt.Errorf("%s.pid %s가 프로세스 표에 없음", where, pid)
		}
		f, err := load.Object(e["f"], where+".f")
		if err != nil {
			return err
		}
		fields := make(map[string]any, len(f))
		for k, val := range f {
			if k != "clock" {
				fields[k] = val
			}
		}
		src := srcRec
		switch {
		case serverStrace[name]:
			src = srcStrace
		case name == "log":
			src = srcLog
			l.Source(atlas.Source{ID: srcLog, Tool: "server.log(log_line_prefix %m [%p])", Clock: clockLog, Nature: atlas.Observed})
		}
		if c, ok := f["clock"]; ok {
			if c != "ftrace" || src != srcRec {
				return fmt.Errorf("%s: %s 사건의 f.clock %v", where, name, c)
			}
			src = srcShift
			l.Source(atlas.Source{ID: srcShift, Tool: "uprobe 사건 가운데 함수 기록에 짝이 없어 ftrace 시각을 옮겨 쓴 것(scenario_server.go 449~505행)", Clock: clockShift, Nature: atlas.Observed})
		}
		stack, err := l.stack(string(pid), e["stack"], src, "$.server.events[]", where)
		if err != nil {
			return err
		}
		place := proc
		if stack != nil {
			place = stack[0]
		}
		kind, t0, t1, err := timed(e, "t", "t1", where)
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: place, Kind: kind, Name: name, StartNS: t0, EndNS: t1, Stack: stack, Source: src}, "$.server.events[]")
		l.B.EventAttrs(id, src, "f", fields)
		l.noteWrite(id, name == "wal_io", fields, t1)
	}
	return nil
}

// serverSpans 메서드는 서버 프로세스가 고른 함수에 머문 구간을 옮깁니다. 장소는 함수 입구 프레임(함수가 정의된 줄)입니다.
func (l *loader) serverSpans(v any) error {
	spans, err := load.List(v, "$.server.spans")
	if err != nil {
		return err
	}
	if len(spans) > 0 {
		l.Source(atlas.Source{ID: srcSpan, Tool: "함수 기록기의 입구·출구 짝(고른 함수에 머문 구간, scenario_server.go 304행)", Clock: clockRec, Nature: atlas.Observed})
	}
	for i, raw := range spans {
		where := fmt.Sprintf("$.server.spans[%d]", i)
		s, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(s, where, "at", "fn", "pid", "t0", "t1"); err != nil {
			return err
		}
		pid, ok := s["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		proc, ok := l.proc[string(pid)]
		if !ok {
			return fmt.Errorf("%s.pid %s가 프로세스 표에 없음", where, pid)
		}
		fn, err := load.Text(s["fn"], where+".fn")
		if err != nil {
			return err
		}
		at, err := load.Text(s["at"], where+".at")
		if err != nil {
			return err
		}
		place, err := l.Frame(proc, fn, at, srcSpan, "$.server.spans[].fn", "$.server.spans[].at")
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		_, t0, t1, err := timed(s, "t0", "t1", where)
		if err != nil {
			return err
		}
		l.Event(atlas.Event{Place: place, Kind: atlas.Span, Name: fn, StartNS: t0, EndNS: t1, Source: srcSpan}, "$.server.spans[]")
	}
	return nil
}

// before 메서드는 사건 id에서 그 뒤에 보낸 단계 n으로 가는 before 연결을 만듭니다. 그 단계가 없으면 오류입니다.
func (l *loader) before(id int, v any, where, group string) (int, error) {
	n, err := load.Integer(v, where+".beforeStep")
	if err != nil {
		return 0, err
	}
	step, ok := l.step[n]
	if !ok {
		return 0, fmt.Errorf("%s.beforeStep %d의 단계가 없음(시나리오가 끝에 kill로 끝나면 단계 수 + 1)", where, n)
	}
	l.Link(id, step, "before", srcTxlab, group)
	return n, nil
}

// killsOf 메서드는 kill을 txlab의 구간(SIGKILL 직전부터 접속 재개 줄을 본 순간까지)으로 옮기고, 그 뒤 단계와 WAL 쓰기에 잇습니다.
func (l *loader) killsOf(v any) error {
	kills, err := load.List(v, "$.kills")
	if err != nil {
		return err
	}
	for i, raw := range kills {
		where := fmt.Sprintf("$.kills[%d]", i)
		k, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(k, where, "beforeStep", "pid", "session", "t", "tReady", "walSynced", "walWritten", "walWrittenT"); err != nil {
			return err
		}
		pid, ok := k["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		target, ok := l.proc[string(pid)]
		if !ok || k["session"] != l.ownerOf[string(pid)] {
			return fmt.Errorf("%s: pid %s의 세션 %v, 다시 센 값 %q", where, pid, k["session"], l.ownerOf[string(pid)])
		}
		_, t0, t1, err := timed(k, "t", "tReady", where)
		if err != nil {
			return err
		}
		if t1 == nil {
			return fmt.Errorf("%s.tReady 없음", where)
		}
		id := l.Event(atlas.Event{Place: l.client, Target: &target, Kind: atlas.Span, Name: nameKill, StartNS: t0, EndNS: t1, Source: srcTxlab}, "$.kills[]")
		n, err := l.before(id, k["beforeStep"], where, "$.kills[].beforeStep")
		if err != nil {
			return err
		}
		l.kills, l.killAt[id] = append(l.kills, id), n
		for _, key := range []string{"walWritten", "walSynced"} {
			if val, ok := k[key]; ok {
				l.B.EventAttrs(id, srcTxlab, key, val)
			}
		}

		// walWrittenT는 그 시각에 끝난 pg_wal 쓰기 사건 하나를 가리킵니다. 없거나 둘 이상이면 되살릴 수 없습니다.
		// 그 쓰기가 닿은 위치(lsn1)는 walWritten과 같아야 합니다. 해석기는 쓴 범위의 끝이 가장 먼 쓰기의 끝을 walWritten으로,
		// 그 쓰기가 끝난 시각을 walWrittenT로 두므로, 같지 않으면 시각만 우연히 맞은 다른 쓰기입니다.
		_, hasW := k["walWritten"]
		wt, hasT := k["walWrittenT"]
		if hasW != hasT {
			return fmt.Errorf("%s: walWritten과 walWrittenT가 함께 있지 않음", where)
		}
		if !hasT {
			continue
		}
		ns, err := load.Micros(wt, where+".walWrittenT")
		if err != nil {
			return err
		}
		c := l.writes[ns]
		if len(c) != 1 {
			return fmt.Errorf("%s: 끝이 walWrittenT인 pg_wal 쓰기 %d개", where, len(c))
		}
		if c[0].lsn1 != k["walWritten"] {
			return fmt.Errorf("%s: walWritten %v와 그 쓰기의 lsn1 %v가 다름", where, k["walWritten"], c[0].lsn1)
		}
		l.Link(id, c[0].id, "wal", srcTx, WalBeforeKill)
	}
	return nil
}

// pauses 메서드는 시나리오의 쉼을 txlab의 구간으로 옮기고 그 뒤에 보낸 단계와 잇습니다.
func (l *loader) pauses(v any) error {
	pauses, err := load.List(v, "$.pauses")
	if err != nil {
		return err
	}
	for i, raw := range pauses {
		where := fmt.Sprintf("$.pauses[%d]", i)
		p, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(p, where, "beforeStep", "t", "t1"); err != nil {
			return err
		}
		_, t0, t1, err := timed(p, "t", "t1", where)
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: l.client, Kind: atlas.Span, Name: namePause, StartNS: t0, EndNS: t1, Source: srcTxlab}, "$.pauses[]")
		if _, err := l.before(id, p["beforeStep"], where, "$.pauses[].beforeStep"); err != nil {
			return err
		}
	}
	return nil
}

// captures 메서드는 관찰을 txlab의 sample로, 부분마다 읽은 구간을 해석기 구간으로 옮기고 관찰에서 contains 연결로 잇습니다.
//
// 읽기 구간은 그 질의를 받은 때부터 관찰용 접속이 다음 질의를 받은 때(없으면 다음 관찰의 시작)까지로 해석기가 정한 위쪽 경계라
// 해석기 출처입니다(decode/scenario_check.go 395~433, Spring 코드 검토 지적 3a). 출처가 관찰과 달라 자식이 아니라 연결입니다.
// 해석기는 끝을 모르는 마지막 관찰의 마지막 부분을 끝이 시작과 같게 두었으므로(scenario.go 560행), 그 부분은 끝을 비웁니다.
func (l *loader) captures(v any) error {
	caps, err := load.List(v, "$.captures")
	if err != nil {
		return err
	}
	for i, raw := range caps {
		where := fmt.Sprintf("$.captures[%d]", i)
		c, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(c, where, "activity", "afterKill", "afterStep", "locks", "pages", "reads", "running", "t", "wal", "xacts"); err != nil {
			return err
		}
		t, err := load.Micros(c["t"], where+".t")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: l.client, Kind: atlas.Sample, Name: nameCapture, StartNS: &t, Source: srcTxlab}, "$.captures[]")
		for _, k := range []string{"activity", "locks", "pages", "running", "xacts", "wal"} {
			if val, ok := c[k]; ok {
				l.B.EventAttrs(id, srcTxlab, k, val)
			}
		}
		reads, err := load.Object(c["reads"], where+".reads")
		if err != nil {
			return err
		}
		parts := make([]string, 0, len(reads))
		for p := range reads {
			parts = append(parts, p)
		}
		sort.Strings(parts)
		for _, p := range parts {
			pair, err := load.List(reads[p], where+".reads."+p)
			if err != nil || len(pair) != 2 {
				return fmt.Errorf("%s.reads.%s가 시각 둘의 목록이 아님", where, p)
			}
			a, err := load.Micros(pair[0], where+".reads."+p+"[0]")
			if err != nil {
				return err
			}
			b, err := load.Micros(pair[1], where+".reads."+p+"[1]")
			if err != nil {
				return err
			}
			end := &b
			if a == b {
				if i != len(caps)-1 {
					return fmt.Errorf("%s.reads.%s: 마지막 관찰이 아닌데 끝이 시작과 같음", where, p)
				}
				end = nil
			}
			read := l.Event(atlas.Event{Place: l.client, Kind: atlas.Span, Name: p, StartNS: &a, EndNS: end, Source: srcTx}, "$.captures[].reads."+p)
			l.Link(id, read, "contains", srcTx, "$.captures[].reads")
		}

		// afterStep 0은 첫 단계 앞의 관찰이라 이을 단계가 없습니다. afterKill 관찰은 kill 앞 마지막 단계와 그 kill 둘에 잇습니다.
		n, err := load.Integer(c["afterStep"], where+".afterStep")
		if err != nil {
			return err
		}
		if n != 0 {
			step, ok := l.step[n]
			if !ok {
				return fmt.Errorf("%s.afterStep %d의 단계가 없음", where, n)
			}
			l.Link(id, step, "after", srcTxlab, "$.captures[].afterStep")
		}
		if ak, ok := c["afterKill"]; ok {
			if ak != true {
				return fmt.Errorf("%s.afterKill이 true가 아님", where)
			}
			var found []int
			for _, k := range l.kills {
				if l.killAt[k] == n+1 {
					found = append(found, k)
				}
			}
			if len(found) != 1 {
				return fmt.Errorf("%s: afterStep %d 바로 뒤의 kill %d개", where, n, len(found))
			}
			l.Link(id, found[0], "after", srcTxlab, "$.captures[].afterKill")
		}
	}
	return nil
}

// bystanders 메서드는 관찰 대상이 아닌 백엔드의 활동 요약을 그 프로세스의 해석기 sample(시각 미기록)로 옮깁니다.
func (l *loader) bystanders(v any) error {
	list, err := load.List(v, "$.bystanders")
	if err != nil {
		return err
	}
	if len(list) > 0 {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcTx, Field: "bystander.start_ns", Note: "기록 전체에 걸친 요약이라 시각이 없습니다"})
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.bystanders[%d]", i)
		b, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		keys := []string{"firstSql", "hints", "outsideHints", "queries", "watchedChecks", "watchedHints"}
		if err := load.Keys(b, where, append(keys, "pid")...); err != nil {
			return err
		}
		pid, ok := b["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		proc, ok := l.proc[string(pid)]
		if !ok {
			return fmt.Errorf("%s.pid %s가 프로세스 표에 없음", where, pid)
		}
		id := l.Event(atlas.Event{Place: proc, Kind: atlas.Sample, Name: nameBystand, Source: srcTx}, "$.bystanders[]")
		for _, k := range keys {
			if val, ok := b[k]; ok {
				l.B.EventAttrs(id, srcTx, k, val)
			}
		}
	}
	return nil
}

// hints 메서드는 힌트 비트 변화 설명을 추정 출처의 순간으로 옮깁니다. 장소는 by 세션 백엔드 아래의 코드 자리이고,
// 같은 자리가 이미 있으면(판정 경로의 SetHintBits) 그것을 씁니다. 단계와는 contains 연결입니다.
func (l *loader) hints(v any) error {
	list, err := load.List(v, "$.hintChanges")
	if err != nil {
		return err
	}
	if len(list) > 0 {
		l.Source(atlas.Source{ID: srcHint, Tool: "해석기가 주소나 값으로 맞춘 힌트 변화 설명(scenario_check.go, via·by가 짐작)", Clock: clockRec, Nature: atlas.Inferred})
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.hintChanges[%d]", i)
		h, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(h, where, "at", "bit", "by", "ctid", "step", "t", "table", "via"); err != nil {
			return err
		}
		by, err := load.Text(h["by"], where+".by")
		if err != nil {
			return err
		}
		pid, ok := l.backend[by]
		if !ok {
			return fmt.Errorf("%s.by %s가 sessions에 없음", where, by)
		}
		at, err := load.Text(h["at"], where+".at")
		if err != nil {
			return err
		}
		site, err := l.Site(l.proc[pid], at, srcHint, "$.hintChanges[].at")
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		t, err := load.Micros(h["t"], where+".t")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: site, Kind: atlas.Instant, Name: nameHint, StartNS: &t, Source: srcHint}, "$.hintChanges[]")
		for _, k := range []string{"bit", "ctid", "table", "via"} {
			l.B.EventAttrs(id, srcHint, k, h[k])
		}
		n, err := load.Integer(h["step"], where+".step")
		if err != nil {
			return err
		}
		step, ok := l.step[n]
		if !ok {
			return fmt.Errorf("%s.step %d의 단계가 없음", where, n)
		}
		l.Link(step, id, "contains", srcHint, "$.hintChanges[].step")
	}
	return nil
}

// notes 메서드는 원본이 뜻을 가르지 못하는 꼴을 run의 ambiguity note로 남깁니다(설계 메모 4).
func (l *loader) notes() {
	if l.xipNull == 0 {
		return
	}
	l.Source(atlas.Source{ID: srcNote, Tool: "pgsql 주제의 트랜잭션 실험 적재기(topics/pgsql/txlab)", Nature: atlas.Human})
	l.B.AddNote(atlas.Note{Scope: atlas.Scope{Kind: "run"}, Kind: "ambiguity", Source: srcNote,
		Text: fmt.Sprintf("스냅숏 %d개의 xip가 null입니다. 해석기는 xcnt만큼 xip를 붙이므로(decode/scenario.go 824~827행) null은 진행 중인 트랜잭션 0개이거나 xcnt를 기록하지 못한 것이고, 원본이 둘을 가르지 않아 값 그대로 옮겼습니다.", l.xipNull)})
}
