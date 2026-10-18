// Package spring 는 Spring 전파 실험 화면(concept-9)에 박힌 기록 14벌의 해석 결과를 공통 형식으로 옮기고 되돌립니다.
//
// 옮기는 규칙은 설계 메모 5의 4이고, 행마다의 근거는 docs/works/evidence/learning-map-20261017/t1-e03-evidence.md입니다.
// 앱 쪽은 앱 장소 아래 JFR 스레드 → 메서드 → 서술자, 연결 쪽은 연결 장소, 서버 쪽은 백엔드 프로세스와 그 아래 함수·프레임·코드 자리입니다.
// 연결과 백엔드의 짝은 메시지의 백엔드 쪽 끝(target·from)이 나르고 따로 두지 않습니다. 해석기가 다른 칸에서 셈한 칸은 저장하지 않고
// 내보내기가 같은 규칙으로 다시 셉니다. 원본의 모든 열쇠는 목록으로 받고, 모르는 열쇠나 규칙으로 되살릴 수 없는 꼴이 있으면 멈춥니다.
// 적재가 끝나면 내보내기로 되돌려 원본과 견주고, 다르면 적재 결과를 돌려주지 않습니다.
package spring

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// 해석기가 정한 규칙으로 만든 연결과 주제가 세운 장소의 경로 무리입니다(설계 메모 7, 검사 4). 표시 파일에 맞는 경로가 없는 것만 둡니다.
const (
	// WirePair 는 맥 쪽 메시지 → 컨테이너 쪽 메시지의 짝입니다(같은 순번에 방향·번호·종류·길이가 같음, decode/spring.go 204~220).
	WirePair = "rule:wire-pair"
	// RequestBatch 는 요청 → 그 요청의 메시지입니다(Sync나 Q로 끝나는 앱 묶음과 ReadyForQuery까지의 답, splitRequests).
	RequestBatch = "rule:request-batch"
	// ServerWindow 는 요청 → 백엔드의 서버 창입니다(ReadyForQuery 입구 차례로 짝지음, decode/spring.go 516~535).
	ServerWindow = "rule:server-window"
	// RequestWindow 는 서버 창 → 창 안의 서버 사건·스냅숏·줄 판정·기다림입니다.
	RequestWindow = "rule:request-window"
	// ClientPlace 는 기록을 남긴 txlab 프로그램의 장소입니다. 원본에 행으로 있지 않고 주제가 세웁니다.
	ClientPlace = "topic:client"
	// AppPlace 는 txlab이 띄운 Spring 앱(JVM)의 장소입니다.
	AppPlace = "topic:app"
)

// InferredGroups 함수는 추정 목록(설계 메모 7)에 든 경로 무리를 돌려줍니다: 100 ms 창으로 연결을 붙인 빌림, 그 빌림 위에서 다시 센 연결
// 묶임·슬롯·묶임 효과, 맥 시계의 보낸 시각을 JFR 구간에 견준 wire 효과.
func InferredGroups() map[string]bool {
	return map[string]bool{
		"$.borrows[]": true, "$.view.binding[]": true, "$.view.binding[].by": true, "$.view.slots[]": true,
		"$.view.tree[].effects[].conn": true, "$.view.tree[].effects[].seq": true,
	}
}

// 출처 id입니다. 시계는 출처가 가리키므로, 해석기 출처도 시계마다 나눕니다(설계 메모 5의 4).
const (
	srcTxlab   = "txlab"
	srcObserve = "observe"
	srcWireMac = "wire-mac"
	srcWireCnt = "wire-cont"
	srcRec     = "recorder"
	srcJFR     = "jfr"
	srcSpring  = "spring"
	srcMac     = "spring-mac"
	srcCont    = "spring-cont"
	srcTree    = "spring-jfr"
	srcGuess   = "spring-guess"
	srcCompare = "spring-compare"
	srcSrc     = "src"
	srcJsrc    = "jsrc"
	srcTopic   = "topic"
	srcNote    = "pgsql-topic"
)

// 시계 id입니다(설계 메모 3의 표).
const (
	clockMac  = "mac"
	clockCont = "container"
	clockJFR  = "jfr"
)

// 사건 이름 가운데 원본의 kind·type·fn이 아니라 주제가 붙인 것입니다. 원본 목록의 열쇠 이름을 따랐습니다.
const (
	nameCheckpoint = "checkpoint"
	nameNote       = "note"
	nameBlocked    = "blocked"
	nameCapture    = "capture"
	nameConn       = "connection"
	nameRequest    = "request"
	nameServer     = "server"
	nameCheck      = "check"
	nameLine       = "line"
	namePath       = "path"
	nameInner      = "inner"
	nameTx         = "transaction"
	nameBorrow     = "borrow"
	nameThrow      = "throw"
	namePark       = "park"
	nameBinding    = "binding"
	nameSlot       = "slot"
	nameWait       = "wait"
)

// sceneThread 는 장면 메서드가 도는 스레드입니다. 호출 트리, 연결 묶임, 슬롯은 이 스레드의 것뿐입니다(decode/spring_tree.go 22).
const sceneThread = "main"

// authSQL 은 해석기가 접속 인증 요청(seq 0)에 붙이는 글입니다(decode/spring.go 259).
const authSQL = "(접속 인증)"

// topKeys 는 기록 한 벌의 맨 위 열쇠입니다(decode/spring.go 654~676, 재고 조사 2.1절).
var topKeys = []string{"blocked", "borrows", "captures", "checkpoints", "clock", "conns", "java", "javaExit", "javaVersion", "joins", "notes", "pool", "scene", "sessions", "siteRepair", "src", "tables", "title", "transactions", "view"}

// viewKeys 는 view의 열쇠입니다(decode/spring.go 684~701).
var viewKeys = []string{"ambiguous", "binding", "branches", "compared", "java", "javaFrom", "jsrc", "pauses", "poolConfig", "roots", "siteMismatch", "slots", "tree", "waitCycles", "waits", "window"}

// effectRank 는 한 마디 안의 효과 차례입니다. bindingReplay가 묶임 효과를 먼저 붙이고, annotateTree가 종류 효과, 잠듦, 예외 차례로 붙입니다
// (decode/spring.go 730~731, spring_tree.go 847~906).
var effectRank = map[string]int{"suspend": 0, "resume": 0, "return": 0, "wire": 1, "borrow": 1, "autocommit": 1, "rollbackOnly": 1, "validate": 1, "park": 2, "throw": 3}

// Result 값은 적재 결과입니다. Origins는 검사에만 쓰는 곁 기록이고 형식에는 들어가지 않습니다.
type Result struct {
	Run     *atlas.Run
	Origins atlas.Origins
}

// borrowRef 값은 빌림 사건 하나와 효과를 붙일 때 쓰는 시각입니다.
type borrowRef struct {
	id     int
	t0, t1 int64
	thread string
}

// loader 값은 기록 한 벌을 조립하는 동안의 상태입니다.
type loader struct {
	*load.Loader
	client, app int
	thread      map[string]int      // 스레드 이름 → 스레드 장소
	conn        map[string]int      // 연결 이름(C1) → 연결 장소
	proc        map[string]int      // 연결 이름 → 백엔드 프로세스 장소
	msgs        map[string][][2]int // 연결 이름 → 메시지마다 [맥 쪽 사건, 컨테이너 쪽 사건]
	reqs        map[string][]int    // 연결 이름 → 요청 사건(seq 차례)
	done        map[string][]int64  // 연결 이름 → 요청의 done(ns)
	cps         map[int]int         // 멈춤 자리 n → 멈춤 자리 사건
	cpCapture   map[int64][]int     // tCapture(ns) → 멈춤 자리 사건
	blCapture   map[int64][]int     // tCapture(ns) → 막힘 관찰 사건
	throws      map[int64][]int     // main 스레드 예외의 시각 → 예외 사건
	parks       map[[2]int64][]int  // main 스레드 잠듦의 [시작, 끝] → 잠듦 사건
	borrows     []borrowRef
	node        map[int]int                       // 트리 마디 id → 마디 사건
	waits       map[string]map[int]map[string]any // 연결 이름 → seq → view.waits의 원소
	zero        int                               // 0인 시각의 수(rel(0) = 0이 기준 순간과 값 없음을 섞음, decode/spring.go 408~411)
	xipNull     int
	noConn      int
}

// ReadRuns 함수는 Spring 화면(concept-9-spring.html)에 박힌 RUNS literal(기록 열쇠 → 해석 결과)을 읽습니다.
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

// Import 함수는 기록 한 벌(RUNS의 열쇠 key와 그 값 v)을 run 하나로 옮깁니다. v는 고치지 않습니다.
//
// version은 PostgreSQL 소스 판(workbench.json의 sourceRoot)이고 PostgreSQL 정적 장소의 key에 들어갑니다. 원본에 모르는 열쇠가 있거나,
// 메모 5의 4의 규칙으로 되살릴 수 없는 꼴(짝이 없는 관찰, 맞는 길이 하나가 아닌 마디, 끝이 맥 시계인 기다림 등)이거나, 적재 결과를
// 내보낸 것이 원본과 다르면 오류를 돌려줍니다.
func Import(key string, v map[string]any, version string) (Result, error) {
	if err := load.Keys(v, "$", topKeys...); err != nil {
		return Result{}, err
	}
	view, err := load.Object(v["view"], "$.view")
	if err != nil {
		return Result{}, err
	}
	if err := load.Keys(view, "$.view", viewKeys...); err != nil {
		return Result{}, err
	}
	l := &loader{
		Loader: load.NewLoader(atlas.NewBuilder("spring/"+key, "pgsql", "spring", key,
			atlas.Origin{Clock: clockMac, Label: "txlab이 Spring 시나리오를 시작한 순간(spring.json의 startNs, 맥 시계)"}), version),
		thread: map[string]int{}, conn: map[string]int{}, proc: map[string]int{}, msgs: map[string][][2]int{}, reqs: map[string][]int{},
		done: map[string][]int64{}, cps: map[int]int{}, cpCapture: map[int64][]int{}, blCapture: map[int64][]int{}, throws: map[int64][]int{},
		parks: map[[2]int64][]int{}, node: map[int]int{}, waits: map[string]map[int]map[string]any{},
	}

	// 머리(출처, 시계, run attrs)를 먼저, 그다음 장소와 JFR, 앱 쪽 순간, 연결 위, 서버, 해석기의 트리와 효과 차례로 옮깁니다.
	// 효과·빌림·묶임은 가리키는 사건(요청, JFR 예외·잠듦, 트리 마디)이 있어야 해서 뒤에 둡니다.
	if err := l.header(v, view); err != nil {
		return Result{}, fmt.Errorf("%s: %w", key, err)
	}
	conns, err := load.List(v["conns"], "$.conns")
	if err != nil {
		return Result{}, err
	}
	for _, do := range []func() error{
		func() error { return l.places(conns) },
		func() error { return l.java(v["java"]) },
		func() error { return l.checkpoints(v["checkpoints"]) },
		func() error { return l.notes(v["notes"]) },
		func() error { return l.blocked(v["blocked"]) },
		func() error { return l.wire(conns) },
		func() error { return l.indexWaits(view["waits"]) },
		func() error { return l.requests(conns) },
		func() error { return l.captures(v["captures"]) },
		func() error { return l.transactions(v["transactions"]) },
		func() error { return l.borrowsOf(v["borrows"]) },
		func() error { return l.tree(view["tree"]) },
		func() error { return l.binding(view["binding"]) },
		func() error { return l.slots(view["slots"]) },
	} {
		if err := do(); err != nil {
			return Result{}, fmt.Errorf("%s: %w", key, err)
		}
	}
	for name, ws := range l.waits {
		if len(ws) != 0 {
			return Result{}, fmt.Errorf("%s: 연결 %s의 기다림 %d개를 품을 서버 창이 없음", key, name, len(ws))
		}
	}
	l.finish()

	// 적재 결과만으로 되돌린 기록이 원본과 같아야 합니다. 다르면 규칙으로 되살릴 수 없는 꼴이라 결과를 내지 않습니다.
	got, err := Export(l.B.Run())
	if err != nil {
		return Result{}, fmt.Errorf("%s: 되돌리기: %w", key, err)
	}
	want, err := load.Canonical(v)
	if err != nil {
		return Result{}, err
	}
	back, err := load.Canonical(got)
	if err != nil {
		return Result{}, err
	}
	if diff := atlas.Diff("$", want, back); len(diff) != 0 {
		return Result{}, fmt.Errorf("%s: 되돌린 기록이 원본과 다름(%d곳, 처음 %s)", key, len(diff), diff[0])
	}
	return Result{Run: l.B.Run(), Origins: l.Origins}, nil
}

// micros 메서드는 원본의 마이크로초를 정수 나노초로 옮기고, 0인 시각을 셉니다(설계 메모 4의 Spring 시각 0).
func (l *loader) micros(v any, where string) (int64, error) {
	ns, err := load.Micros(v, where)
	if err == nil && ns == 0 {
		l.zero++
	}
	return ns, err
}

// header 메서드는 출처와 시계, run attrs를 더합니다. container → mac 관계는 연결마다 잰 차이 구간을 모두 덮는 범위입니다.
func (l *loader) header(v, view map[string]any) error {
	for _, s := range []atlas.Source{
		{ID: srcTxlab, Tool: "txlab Spring 지휘(spring.json: 멈춤 자리, 예외 알림, 접속의 받음·닫음, 막힘 관찰)", Clock: clockMac, Nature: atlas.Observed},
		{ID: srcObserve, Tool: "txlab 관찰기(steps.json의 관찰, 관찰용 접속의 strace로 해석기가 정한 읽기 구간과 끝)", Clock: clockCont, Nature: atlas.Observed},
		{ID: srcWireMac, Tool: "맥 쪽 중계기 wire-mac-<n>.jsonl(txlab/wire.go)", Clock: clockMac, Nature: atlas.Observed},
		{ID: srcWireCnt, Tool: "컨테이너 쪽 중계기 wire-<n>.jsonl(해석 결과에는 시각만 남음)", Clock: clockCont, Nature: atlas.Observed},
		{ID: srcRec, Tool: "함수 기록기·uprobe(해석기 decode/spring.go·scenario.go가 읽음)", Clock: clockCont, Nature: atlas.Observed},
		{ID: srcJFR, Tool: "JFR spring.jfr(jdk.MethodTrace, jdk.JavaExceptionThrow, jdk.ThreadPark, decode/spring_jfr.go가 읽음)", Clock: clockJFR, Nature: atlas.Observed},
		{ID: srcSpring, Tool: "해석기 decode/spring*.go(짝과 묶음의 규칙, 해석기의 기록 요약)", Nature: atlas.Derived},
		{ID: srcMac, Tool: "해석기 decode/spring.go의 요청 묶음(splitRequests)과 트랜잭션(deriveTxs)", Clock: clockMac, Nature: atlas.Derived},
		{ID: srcCont, Tool: "해석기 decode/spring_tree.go의 서버 기다림(serverWaits)", Clock: clockCont, Nature: atlas.Derived},
		{ID: srcTree, Tool: "해석기 decode/spring_tree.go의 호출 트리(buildCallTree)와 효과(annotateTree)", Clock: clockJFR, Nature: atlas.Derived},
		{ID: srcGuess, Tool: "해석기의 짐작: 100 ms 창으로 붙인 빌림의 연결(spring_jfr.go borrows), 그 위의 연결 묶임·슬롯·묶임 효과(bindingReplay), 맥 시각을 JFR 구간에 견준 wire 효과", Clock: clockJFR, Nature: atlas.Inferred},
		{ID: srcCompare, Tool: "해석기 decode/spring_compare.go(14벌을 견준 갈림)", Nature: atlas.Derived},
		{ID: srcSrc, Tool: "PostgreSQL 소스 발췌(서버 사건 스택이 가리킨 줄)", Nature: atlas.Observed},
		{ID: srcJsrc, Tool: "자바 소스 발췌(기록의 app-src 사본과 라이브러리 소스, view.javaFrom)", Nature: atlas.Observed},
		{ID: srcTopic, Tool: "pgsql 주제(Spring 적재기 topics/pgsql/spring)", Nature: atlas.Human},
	} {
		l.Source(s)
	}

	// container − mac은 연결마다 [lo, hi] ms입니다. 차이가 일정하다면 참값은 모든 구간의 겹침 안이므로, 모두를 덮는 범위는 더 넓은 쪽입니다.
	clock, err := load.Object(v["clock"], "$.clock")
	if err != nil {
		return err
	}
	if err := load.Keys(clock, "$.clock", "uprobe", "wire"); err != nil {
		return err
	}
	wire, err := load.List(clock["wire"], "$.clock.wire")
	if err != nil || len(wire) == 0 {
		return fmt.Errorf("$.clock.wire가 비지 않은 목록이 아님")
	}
	lo, hi := int64(math.MaxInt64), int64(math.MinInt64)
	for i, raw := range wire {
		where := fmt.Sprintf("$.clock.wire[%d]", i)
		w, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if w["same"] != true {
			return fmt.Errorf("%s: 두 중계기가 같은 순번으로 자르지 않아 차이 구간을 믿을 수 없음", where)
		}
		a, err := millisToNanos(w["lo"], where+".lo")
		if err != nil {
			return err
		}
		b, err := millisToNanos(w["hi"], where+".hi")
		if err != nil {
			return err
		}
		if a > b {
			return fmt.Errorf("%s: 차이 구간이 빔(lo > hi)", where)
		}
		lo, hi = min(lo, a), max(hi, b)
	}
	l.B.AddClock(atlas.Clock{ID: clockMac, PrecisionNS: 1000,
		Note: "맥의 벽시계: txlab의 time.Now, 맥 쪽 중계기, 앱이 Instant.now()로 찍어 알린 멈춤 자리·예외 알림(Lab.java 37·67·74~78). 해석기가 startNs부터 센 µs로 옮김"})
	l.B.AddClock(atlas.Clock{ID: clockCont, PrecisionNS: 1, Relation: &atlas.ClockRelation{To: clockMac, Known: true, LoNS: lo, HiNS: hi, Basis: srcSpring,
		Note: "연결마다 맞댄 메시지로 잰 (컨테이너 − 맥) 구간(run attrs spring.clock.wire, decode/spring.go 98~107)을 모두 덮는 범위"}})
	l.B.AddClock(atlas.Clock{ID: clockJFR, PrecisionNS: 1,
		Note: "JFR의 startTime을 해석기가 startNs부터 센 µs로 옮긴 축. Lab.java 74행 주석은 맥 쪽 기록과 같은 벽시계라고 적지만 확인한 기록이 없어 맥 시계와의 관계는 모름"})

	// run attrs: 열쇠에 원본의 아래 경로를 적습니다(run의 경로 무리가 $ 하나).
	for _, k := range []string{"scene", "pool", "javaVersion", "javaExit", "title", "tables"} {
		l.B.SetRunAttrs(srcTxlab, k, v[k])
	}
	for _, k := range []string{"clock", "joins", "siteRepair"} {
		l.B.SetRunAttrs(srcSpring, k, v[k])
	}
	if s, ok := v["src"]; ok {
		l.B.SetRunAttrs(srcSrc, "src", s)
	}
	java, err := load.Object(v["java"], "$.java")
	if err != nil {
		return err
	}
	for _, k := range []string{"dups", "stackOnly", "dropped"} {
		val, ok := java[k]
		if !ok {
			return fmt.Errorf("$.java.%s 없음", k)
		}
		l.B.SetRunAttrs(srcSpring, "java."+k, val)
	}
	for src, keys := range map[string][]string{srcSpring: {"ambiguous", "poolConfig", "javaFrom", "siteMismatch"}, srcCompare: {"branches", "compared"}, srcJsrc: {"java", "jsrc"}} {
		for _, k := range keys {
			val, ok := view[k]
			if !ok {
				return fmt.Errorf("$.view.%s 없음", k)
			}
			l.B.SetRunAttrs(src, "view."+k, val)
		}
	}
	return nil
}

// millisToNanos 함수는 원본의 밀리초 수를 정수 나노초로 옮기고, 되돌린 값이 원본과 같은지 확인합니다(clock.wire의 lo·hi).
func millisToNanos(v any, where string) (int64, error) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s가 수가 아님", where)
	}
	f, err := strconv.ParseFloat(string(n), 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	ns := int64(math.Round(f * 1e6))
	if float64(ns)/1e6 != f {
		return 0, fmt.Errorf("%s: 밀리초 %s를 나노초 %d로 옮기면 되돌린 값이 다름", where, n, ns)
	}
	return ns, nil
}

// places 메서드는 주제 장소 둘(txlab, 앱)과 연결·백엔드 프로세스를 세우고 연결의 수명을 그 연결의 구간으로 옮깁니다.
//
// 연결 이름은 "C" + 번호여야 하고(번호는 이름에서 다시 셈), 백엔드 하나가 연결 둘의 것이면 짝이 갈라져 멈춥니다.
func (l *loader) places(conns []any) error {
	l.client = l.Place("client\x00txlab", atlas.PlaceSpec{Kind: "client", Name: "txlab", Source: srcTopic}, ClientPlace)
	l.app = l.Place("app\x00lab.Main", atlas.PlaceSpec{Kind: "app", Name: "lab.Main", Source: srcTopic}, AppPlace)
	owner := map[string]string{}
	for i, raw := range conns {
		where := fmt.Sprintf("$.conns[%d]", i)
		c, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(c, where, "accept", "bridgeErr", "close", "conn", "join", "lastCont", "lastMac", "messages", "name", "params", "pid", "requests", "startup", "thread"); err != nil {
			return err
		}
		n, err := load.Integer(c["conn"], where+".conn")
		if err != nil {
			return err
		}
		name, err := load.Text(c["name"], where+".name")
		if err != nil {
			return err
		}
		if name != fmt.Sprintf("C%d", n) {
			return fmt.Errorf("%s: 이름 %s가 C%d가 아니라 번호를 다시 셀 수 없음", where, name, n)
		}
		pid, ok := c["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		if _, dup := l.conn[name]; dup {
			return fmt.Errorf("연결 %s가 두 번", name)
		}
		if o, dup := owner[string(pid)]; dup {
			return fmt.Errorf("pid %s가 연결 %s와 %s의 백엔드", pid, o, name)
		}
		owner[string(pid)] = name
		l.conn[name] = l.Place("connection\x00"+name, atlas.PlaceSpec{Kind: "connection", Name: name, Source: srcTxlab}, "$.conns[]")
		l.proc[name] = l.Place("process\x00"+string(pid), atlas.PlaceSpec{Kind: "process", Name: string(pid), Source: srcTxlab}, "$.conns[].pid")

		t0, err := l.micros(c["accept"], where+".accept")
		if err != nil {
			return err
		}
		t1, err := l.micros(c["close"], where+".close")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: l.conn[name], Kind: atlas.Span, Name: nameConn, StartNS: &t0, EndNS: &t1, Source: srcTxlab}, "$.conns[]")
		l.B.EventAttrs(id, srcWireMac, "params", c["params"])
		for _, k := range []string{"startup", "lastMac", "lastCont", "join"} {
			l.B.EventAttrs(id, srcSpring, k, c[k])
		}
		if e, ok := c["bridgeErr"]; ok {
			l.B.EventAttrs(id, srcTxlab, "bridgeErr", e)
		}
	}
	return nil
}

// threadPlace 메서드는 이름이 name인 JFR 스레드 장소(앱 아래)를 돌려줍니다.
func (l *loader) threadPlace(name, src, group string) int {
	id := l.Place("thread\x00"+name, atlas.PlaceSpec{Parent: &l.app, Kind: "thread", Name: name, Source: src}, group)
	l.thread[name] = id
	return id
}

// method 메서드는 스레드 장소 아래 자바 메서드 장소를 돌려줍니다. 기록에 자바 쪽 판이 없어 key에 판을 넣지 않습니다.
func (l *loader) method(thread int, fn, src, group string) int {
	return l.Place(fmt.Sprintf("method\x00%d\x00%s", thread, fn),
		atlas.PlaceSpec{Parent: &thread, Kind: "method", Name: fn, Source: src, Static: true}, group)
}

// descriptor 메서드는 메서드 장소 아래 서술자 장소를 돌려줍니다. JFR 구간과 트리 마디가 놓이는 자리입니다.
func (l *loader) descriptor(thread int, fn, desc, src, fnGroup, descGroup string) int {
	m := l.method(thread, fn, src, fnGroup)
	return l.Place(fmt.Sprintf("descriptor\x00%d\x00%s", m, desc),
		atlas.PlaceSpec{Parent: &m, Kind: "descriptor", Name: desc, Source: src, Static: true}, descGroup)
}

// splitRef 함수는 JFR 프레임 글 "패키지.클래스::메서드:줄"을 메서드와 줄로 나눕니다. 줄은 −1(모름) 이상이어야 합니다.
func splitRef(ref string) (string, int, error) {
	i := strings.LastIndexByte(ref, ':')
	if i <= 0 {
		return "", 0, fmt.Errorf("프레임 글 %q", ref)
	}
	n, err := strconv.Atoi(ref[i+1:])
	if err != nil || n < -1 {
		return "", 0, fmt.Errorf("프레임 글 %q의 줄", ref)
	}
	return ref[:i], n, nil
}

// frame 메서드는 JFR 프레임 글 하나의 프레임 장소를 그 스레드의 메서드 아래에 돌려줍니다. 글에 서술자가 없어 서술자 장소 아래가 아닙니다.
// 줄 −1은 위치 없는 프레임이고, 줄이 있으면 경로가 빈 위치입니다(JFR은 소스 파일을 적지 않음).
func (l *loader) frame(thread int, ref, src, group string) (int, error) {
	fn, line, err := splitRef(ref)
	if err != nil {
		return 0, err
	}
	m := l.method(thread, fn, src, group)
	if line == -1 {
		return l.Place(fmt.Sprintf("jframe\x00%d\x00?", m), atlas.PlaceSpec{Parent: &m, Kind: "frame", Name: fn, Source: src, Static: true}, group), nil
	}
	return l.Place(fmt.Sprintf("jframe\x00%d\x00%d", m, line),
		atlas.PlaceSpec{Parent: &m, Kind: "frame", Name: fn, Loc: &atlas.Loc{Line: line}, Source: src, Static: true}, group), nil
}

// frames 메서드는 프레임 글 목록(안쪽부터)을 프레임 장소 목록으로 옮깁니다. 0번이 사건의 장소입니다.
func (l *loader) frames(thread int, v any, src, group, where string) ([]int, error) {
	list, err := load.List(v, where)
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(list))
	for k, raw := range list {
		ref, err := load.Text(raw, fmt.Sprintf("%s[%d]", where, k))
		if err != nil {
			return nil, err
		}
		f, err := l.frame(thread, ref, src, group)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", where, k, err)
		}
		out = append(out, f)
	}
	return out, nil
}

// java 메서드는 JFR 구간·예외·잠듦을 관찰 사건으로 옮깁니다. 구간은 서술자 장소, 예외·잠듦은 스택 0번 프레임에 놓입니다.
// main 스레드의 예외·잠듦은 효과가 가리키므로 시각으로 적어 둡니다.
func (l *loader) java(v any) error {
	j, err := load.Object(v, "$.java")
	if err != nil {
		return err
	}
	if err := load.Keys(j, "$.java", "dropped", "dups", "parks", "spans", "stackOnly", "throws"); err != nil {
		return err
	}
	spans, err := load.List(j["spans"], "$.java.spans")
	if err != nil {
		return err
	}
	for i, raw := range spans {
		where := fmt.Sprintf("$.java.spans[%d]", i)
		s, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(s, where, "caller", "depth", "desc", "fn", "t0", "t1", "thread", "twice"); err != nil {
			return err
		}
		names, err := texts(s, where, "fn", "desc", "thread")
		if err != nil {
			return err
		}
		th := l.threadPlace(names[2], srcJFR, "$.java.spans[].thread")
		place := l.descriptor(th, names[0], names[1], srcJFR, "$.java.spans[].fn", "$.java.spans[].desc")
		t0, err := l.micros(s["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := l.micros(s["t1"], where+".t1")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: place, Kind: atlas.Span, Name: names[0], StartNS: &t0, EndNS: &t1, Source: srcJFR}, "$.java.spans[]")
		for _, k := range []string{"caller", "depth", "twice"} {
			if val, ok := s[k]; ok {
				l.B.EventAttrs(id, srcSpring, k, val)
			}
		}
	}

	throws, err := load.List(j["throws"], "$.java.throws")
	if err != nil {
		return err
	}
	for i, raw := range throws {
		where := fmt.Sprintf("$.java.throws[%d]", i)
		t, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(t, where, "at", "class", "message", "t", "thread"); err != nil {
			return err
		}
		name, err := load.Text(t["thread"], where+".thread")
		if err != nil {
			return err
		}
		th := l.threadPlace(name, srcJFR, "$.java.throws[].thread")
		stack, err := l.frames(th, t["at"], srcJFR, "$.java.throws[].at[]", where+".at")
		if err != nil {
			return err
		}
		if len(stack) == 0 {
			return fmt.Errorf("%s.at이 비어 사건의 장소가 없음", where)
		}
		ts, err := l.micros(t["t"], where+".t")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: stack[0], Kind: atlas.Instant, Name: nameThrow, StartNS: &ts, Stack: stack, Source: srcJFR}, "$.java.throws[]")
		l.B.EventAttrs(id, srcJFR, "class", t["class"])
		l.B.EventAttrs(id, srcJFR, "message", t["message"])
		if name == sceneThread {
			l.throws[ts] = append(l.throws[ts], id)
		}
	}

	parks, err := load.List(j["parks"], "$.java.parks")
	if err != nil {
		return err
	}
	for i, raw := range parks {
		where := fmt.Sprintf("$.java.parks[%d]", i)
		p, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(p, where, "at", "t0", "t1", "thread", "under"); err != nil {
			return err
		}
		name, err := load.Text(p["thread"], where+".thread")
		if err != nil {
			return err
		}
		th := l.threadPlace(name, srcJFR, "$.java.parks[].thread")
		stack, err := l.frames(th, p["at"], srcJFR, "$.java.parks[].at[]", where+".at")
		if err != nil {
			return err
		}
		if len(stack) == 0 {
			return fmt.Errorf("%s.at이 비어 사건의 장소가 없음", where)
		}
		t0, err := l.micros(p["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := l.micros(p["t1"], where+".t1")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: stack[0], Kind: atlas.Span, Name: namePark, StartNS: &t0, EndNS: &t1, Stack: stack, Source: srcJFR}, "$.java.parks[]")
		l.B.EventAttrs(id, srcSpring, "under", p["under"])
		if name == sceneThread {
			l.parks[[2]int64{t0, t1}] = append(l.parks[[2]int64{t0, t1}], id)
		}
	}
	return nil
}

// texts 함수는 객체의 글 칸 여럿을 차례로 꺼냅니다. 하나라도 글이 아니면 오류입니다.
func texts(m map[string]any, where string, keys ...string) ([]string, error) {
	out := make([]string, len(keys))
	for i, k := range keys {
		s, err := load.Text(m[k], where+"."+k)
		if err != nil {
			return nil, err
		}
		out[i] = s
	}
	return out, nil
}

// checkpoints 메서드는 멈춤 자리를 앱 장소의 구간(앱이 알린 t부터 txlab이 go를 보낸 tGo까지, 둘 다 맥 시계)으로 옮깁니다.
// tCapture는 관찰과 잇는 same 연결로 다시 세므로, 관찰을 옮길 때 찾도록 적어 둡니다.
func (l *loader) checkpoints(v any) error {
	list, err := load.List(v, "$.checkpoints")
	if err != nil {
		return err
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.checkpoints[%d]", i)
		c, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(c, where, "label", "n", "t", "tCapture", "tGo"); err != nil {
			return err
		}
		n, err := load.Integer(c["n"], where+".n")
		if err != nil {
			return err
		}
		if _, dup := l.cps[n]; dup {
			return fmt.Errorf("멈춤 자리 번호 %d가 두 번", n)
		}
		t0, err := l.micros(c["t"], where+".t")
		if err != nil {
			return err
		}
		t1, err := l.micros(c["tGo"], where+".tGo")
		if err != nil {
			return err
		}
		tc, err := l.micros(c["tCapture"], where+".tCapture")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: l.app, Kind: atlas.Span, Name: nameCheckpoint, StartNS: &t0, EndNS: &t1, Source: srcTxlab}, "$.checkpoints[]")
		l.B.EventAttrs(id, srcTxlab, "n", c["n"])
		l.B.EventAttrs(id, srcTxlab, "label", c["label"])
		l.cps[n] = id
		l.cpCapture[tc] = append(l.cpCapture[tc], id)
	}
	return nil
}

// notes 메서드는 앱이 알린 예외를 앱 장소의 순간으로 옮깁니다.
func (l *loader) notes(v any) error {
	list, err := load.List(v, "$.notes")
	if err != nil {
		return err
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.notes[%d]", i)
		n, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(n, where, "chain", "message", "t", "where"); err != nil {
			return err
		}
		t, err := l.micros(n["t"], where+".t")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: l.app, Kind: atlas.Instant, Name: nameNote, StartNS: &t, Source: srcTxlab}, "$.notes[]")
		for _, k := range []string{"where", "chain", "message"} {
			l.B.EventAttrs(id, srcTxlab, k, n[k])
		}
	}
	return nil
}

// blocked 메서드는 막힘 관찰(go 뒤 다음 알림이 오지 않아 찍은 관찰)을 txlab 장소의 순간으로 옮깁니다. 기다린 백엔드가 target이고,
// 앞 멈춤 자리와 after 연결입니다. 기다린 백엔드가 둘 이상이면 target 하나로 둘 수 없어 멈춥니다.
func (l *loader) blocked(v any) error {
	list, err := load.List(v, "$.blocked")
	if err != nil {
		return err
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.blocked[%d]", i)
		b, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(b, where, "after", "t", "tCapture", "waiters"); err != nil {
			return err
		}
		waiters, err := load.List(b["waiters"], where+".waiters")
		if err != nil {
			return err
		}
		var target *int
		switch len(waiters) {
		case 0:
		case 1:
			name, err := load.Text(waiters[0], where+".waiters[0]")
			if err != nil {
				return err
			}
			p, ok := l.proc[name]
			if !ok {
				return fmt.Errorf("%s.waiters의 %s가 연결이 아님", where, name)
			}
			target = &p
		default:
			return fmt.Errorf("%s: 기다린 백엔드 %d개는 target 하나로 둘 수 없음", where, len(waiters))
		}
		t, err := l.micros(b["t"], where+".t")
		if err != nil {
			return err
		}
		tc, err := l.micros(b["tCapture"], where+".tCapture")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: l.client, Target: target, Kind: atlas.Instant, Name: nameBlocked, StartNS: &t, Source: srcTxlab}, "$.blocked[]")
		n, err := load.Integer(b["after"], where+".after")
		if err != nil {
			return err
		}
		cp, ok := l.cps[n]
		if !ok {
			return fmt.Errorf("%s.after %d의 멈춤 자리가 없음", where, n)
		}
		l.Link(id, cp, "after", srcTxlab, "$.blocked[].after")
		l.blCapture[tc] = append(l.blCapture[tc], id)
	}
	return nil
}

// wire 메서드는 연결마다 메시지를 맥 쪽·컨테이너 쪽 사건 둘과 그 짝으로 옮깁니다. 앱이 보낸 F는 백엔드가 target, 서버가 보낸 B는 from입니다.
func (l *loader) wire(conns []any) error {
	for i, raw := range conns {
		c := raw.(map[string]any)
		name := c["name"].(string)
		list, err := load.List(c["messages"], fmt.Sprintf("$.conns[%d].messages", i))
		if err != nil {
			return err
		}
		proc := l.proc[name]
		for k, mr := range list {
			where := fmt.Sprintf("$.conns[%d].messages[%d]", i, k)
			m, err := load.Object(mr, where)
			if err != nil {
				return err
			}
			if err := load.Keys(m, where, "dir", "f", "len", "name", "t", "tc", "text", "type"); err != nil {
				return err
			}
			names, err := texts(m, where, "dir", "type")
			if err != nil {
				return err
			}
			var from, target *int
			switch names[0] {
			case "F":
				target = &proc
			case "B":
				from = &proc
			default:
				return fmt.Errorf("%s.dir %q", where, names[0])
			}
			t, err := l.micros(m["t"], where+".t")
			if err != nil {
				return err
			}
			tc, err := l.micros(m["tc"], where+".tc")
			if err != nil {
				return err
			}
			mac := l.Event(atlas.Event{Place: l.conn[name], From: from, Target: target, Kind: atlas.Instant, Name: names[1], StartNS: &t, Source: srcWireMac}, "$.conns[].messages[]")
			for _, key := range []string{"len", "name", "text", "f"} {
				if val, ok := m[key]; ok {
					l.B.EventAttrs(mac, srcWireMac, key, val)
				}
			}
			cnt := l.Event(atlas.Event{Place: l.conn[name], From: from, Target: target, Kind: atlas.Instant, Name: names[1], StartNS: &tc, Source: srcWireCnt}, "$.conns[].messages[]")
			l.Link(mac, cnt, "message", srcSpring, WirePair)
			l.msgs[name] = append(l.msgs[name], [2]int{mac, cnt})
		}
	}
	return nil
}

// member 값은 요청 하나에 든 메시지 차례와, 그 요청의 첫 F와 ReadyForQuery의 메시지 차례입니다.
type member struct {
	msgs     []int
	first, z int
}

// splitRequests 함수는 연결 하나의 메시지(원본 꼴)를 요청으로 나눕니다(decode/spring.go 258~308과 같은 규칙).
//
// 앱 쪽 메시지는 Sync(S)나 단순 질의(Q)에서 묶음이 끝나고(끝내기 X는 빼고), 접속 인증(ssl, startup)은 seq 0입니다. 서버 쪽 메시지는
// 앞에서부터 ReadyForQuery(Z)까지 차례로 그 요청의 답입니다. 답이 끝나지 않은 묶음은 버립니다. 메시지마다 어느 요청에 드는지 돌려줍니다.
func splitRequests(msgs []map[string]any) []member {
	reqs := []member{{first: -1, z: -1}}
	cur := -1
	for i, m := range msgs {
		if m["dir"] != "F" || m["type"] == "X" {
			continue
		}
		if m["type"] == "ssl" || m["type"] == "startup" {
			if reqs[0].first < 0 {
				reqs[0].first = i
			}
			reqs[0].msgs = append(reqs[0].msgs, i)
			continue
		}
		if cur < 0 {
			reqs = append(reqs, member{first: i, z: -1})
			cur = len(reqs) - 1
		}
		reqs[cur].msgs = append(reqs[cur].msgs, i)
		if m["type"] == "S" || m["type"] == "Q" {
			cur = -1
		}
	}
	// 끝나지 않은 마지막 묶음은 답을 짝지을 때 자리만 차지하고(해석기도 그 자리까지 셈) 결과에서 빠집니다.
	k := 0
	for i, m := range msgs {
		if m["dir"] != "B" || k >= len(reqs) {
			continue
		}
		reqs[k].msgs = append(reqs[k].msgs, i)
		if m["type"] == "Z" {
			reqs[k].z = i
			k++
		}
	}
	return reqs[:k]
}

// indexWaits 메서드는 서버 기다림(view.waits)을 연결과 seq로 적어 둡니다. 기다림은 그 요청의 서버 창이 품으므로 창을 옮길 때 만듭니다.
func (l *loader) indexWaits(v any) error {
	list, err := load.List(v, "$.view.waits")
	if err != nil {
		return err
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.view.waits[%d]", i)
		w, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(w, where, "conn", "end", "seq", "t0", "t1", "xid"); err != nil {
			return err
		}
		name, err := load.Text(w["conn"], where+".conn")
		if err != nil {
			return err
		}
		seq, err := load.Integer(w["seq"], where+".seq")
		if err != nil {
			return err
		}
		if l.waits[name] == nil {
			l.waits[name] = map[int]map[string]any{}
		}
		if _, dup := l.waits[name][seq]; dup {
			return fmt.Errorf("%s: 연결 %s의 요청 %d에 기다림이 둘(해석기는 요청마다 하나)", where, name, seq)
		}
		l.waits[name][seq] = w
	}
	return nil
}

// requests 메서드는 연결마다 요청을 연결 장소의 구간으로, 그 서버 창을 백엔드의 구간으로 옮기고 창 안의 서버 기록을 품게 합니다.
//
// 요청은 품은 메시지(splitRequests)에서 모든 칸을 다시 세므로 attrs가 없습니다. 원본의 요청 수와 차례, 보낸 시각과 받은 시각이 그 규칙과
// 다르면 멈춥니다.
func (l *loader) requests(conns []any) error {
	for i, raw := range conns {
		c := raw.(map[string]any)
		name := c["name"].(string)
		proc := l.proc[name]
		list, err := load.List(c["requests"], fmt.Sprintf("$.conns[%d].requests", i))
		if err != nil {
			return err
		}
		msgs := make([]map[string]any, 0)
		for _, m := range c["messages"].([]any) {
			msgs = append(msgs, m.(map[string]any))
		}
		groups := splitRequests(msgs)
		if len(groups) != len(list) {
			return fmt.Errorf("$.conns[%d]: 요청 %d개, 메시지에서 다시 센 요청 %d개", i, len(list), len(groups))
		}

		// 함수 이름의 경로는 이 백엔드의 모든 창의 스택 칸에서 모은 뒤에 스택을 옮깁니다(load.Loader.Callee).
		for _, qr := range list {
			s, _ := qr.(map[string]any)["server"].(map[string]any)
			for _, key := range []string{"events", "snapshots", "checks"} {
				items, _ := s[key].([]any)
				for _, e := range items {
					stack, _ := e.(map[string]any)["stack"].([]any)
					for _, f := range stack {
						fr, _ := f.(map[string]any)
						fn, _ := fr["fn"].(string)
						at, _ := fr["at"].(string)
						l.NotePath(proc, fn, at)
					}
				}
			}
		}

		for k, qr := range list {
			where := fmt.Sprintf("$.conns[%d].requests[%d]", i, k)
			q, err := load.Object(qr, where)
			if err != nil {
				return err
			}
			if err := load.Keys(q, where, "done", "error", "seq", "sent", "server", "sql", "status", "stmt", "tags", "types"); err != nil {
				return err
			}
			if seq, err := load.Integer(q["seq"], where+".seq"); err != nil || seq != k {
				return fmt.Errorf("%s.seq가 요청 차례 %d가 아님", where, k)
			}
			g := groups[k]
			if g.first < 0 || g.z < 0 {
				return fmt.Errorf("%s: 보낸 메시지나 ReadyForQuery가 없는 요청", where)
			}
			sent, err := l.micros(q["sent"], where+".sent")
			if err != nil {
				return err
			}
			done, err := l.micros(q["done"], where+".done")
			if err != nil {
				return err
			}
			first, z := l.B.Run().Events[l.msgs[name][g.first][0]], l.B.Run().Events[l.msgs[name][g.z][0]]
			if *first.StartNS != sent || *z.StartNS != done {
				return fmt.Errorf("%s: sent·done이 첫 F와 Z의 시각과 다름", where)
			}
			id := l.Event(atlas.Event{Place: l.conn[name], Kind: atlas.Span, Name: nameRequest, StartNS: &sent, EndNS: &done, Source: srcMac}, "$.conns[].requests[].sent")
			for _, m := range g.msgs {
				l.Link(id, l.msgs[name][m][0], "contains", srcSpring, RequestBatch)
			}
			l.reqs[name] = append(l.reqs[name], id)
			l.done[name] = append(l.done[name], done)

			s, ok := q["server"]
			if !ok {
				if _, w := l.waits[name][k]; w {
					return fmt.Errorf("%s: 서버 창이 없는 요청의 기다림", where)
				}
				continue
			}
			if err := l.server(id, name, proc, k, s, where+".server"); err != nil {
				return err
			}
		}
	}
	return nil
}

// server 메서드는 서버 창 하나를 백엔드의 기록기 구간으로 옮기고, 창 안의 서버 사건·스냅숏·줄 판정과 그 요청의 기다림을 품게 합니다.
func (l *loader) server(req int, name string, proc, seq int, raw any, where string) error {
	s, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(s, where, "catalogSnapshots", "checks", "events", "in", "out", "snapshots", "sql", "t0", "t1"); err != nil {
		return err
	}
	t0, err := l.micros(s["t0"], where+".t0")
	if err != nil {
		return err
	}
	t1, err := l.micros(s["t1"], where+".t1")
	if err != nil {
		return err
	}
	win := l.Event(atlas.Event{Place: proc, Kind: atlas.Span, Name: nameServer, StartNS: &t0, EndNS: &t1, Source: srcRec}, "$.conns[].requests[].server")
	l.B.EventAttrs(win, srcRec, "catalogSnapshots", s["catalogSnapshots"])
	for _, k := range []string{"in", "out"} {
		l.B.EventAttrs(win, srcSpring, k, s[k])
	}
	l.Link(req, win, nameServer, srcSpring, ServerWindow)

	evs, err := load.List(s["events"], where+".events")
	if err != nil {
		return err
	}
	for k, e := range evs {
		if err := l.serverEvent(win, proc, e, fmt.Sprintf("%s.events[%d]", where, k)); err != nil {
			return err
		}
	}
	snaps, err := load.List(s["snapshots"], where+".snapshots")
	if err != nil {
		return err
	}
	for k, e := range snaps {
		if err := l.snapshot(win, proc, e, fmt.Sprintf("%s.snapshots[%d]", where, k)); err != nil {
			return err
		}
	}
	checks, err := load.List(s["checks"], where+".checks")
	if err != nil {
		return err
	}
	for k, e := range checks {
		if err := l.check(win, proc, e, fmt.Sprintf("%s.checks[%d]", where, k)); err != nil {
			return err
		}
	}
	if w, ok := l.waits[name][seq]; ok {
		delete(l.waits[name], seq)
		if err := l.wait(win, proc, w); err != nil {
			return err
		}
	}
	return nil
}

// serverEvent 메서드는 서버 창 안의 기록기 사건 하나를 옮깁니다. 장소는 스택의 가장 안쪽 프레임, 스택이 없으면 백엔드입니다.
func (l *loader) serverEvent(win, proc int, raw any, where string) error {
	e, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(e, where, "f", "kind", "stack", "t", "t1"); err != nil {
		return err
	}
	name, err := load.Text(e["kind"], where+".kind")
	if err != nil {
		return err
	}
	group := "$.conns[].requests[].server.events[]"
	stack, err := l.Stack(proc, e["stack"], srcRec, group, where)
	if err != nil {
		return err
	}
	place := proc
	if stack != nil {
		place = stack[0]
	}
	t0, err := l.micros(e["t"], where+".t")
	if err != nil {
		return err
	}
	kind, end := atlas.Instant, (*int64)(nil)
	if v, ok := e["t1"]; ok {
		t1, err := l.micros(v, where+".t1")
		if err != nil {
			return err
		}
		kind, end = atlas.Span, &t1
	}
	id := l.Event(atlas.Event{Place: place, Kind: kind, Name: name, StartNS: &t0, EndNS: end, Stack: stack, Source: srcRec}, group)
	if f, ok := e["f"]; ok {
		l.B.EventAttrs(id, srcRec, "f", f)
	}
	l.Link(win, id, "contains", srcSpring, RequestWindow)
	return nil
}

// snapshot 메서드는 스냅숏 하나를 순간으로 옮깁니다. xip의 null은 원본이 0개와 기록 안 됨을 가르지 않아 note로 남깁니다.
func (l *loader) snapshot(win, proc int, raw any, where string) error {
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
	group := "$.conns[].requests[].server.snapshots[]"
	stack, err := l.Stack(proc, s["stack"], srcRec, group, where)
	if err != nil {
		return err
	}
	place := proc
	if stack != nil {
		place = stack[0]
	}
	t, err := l.micros(s["t"], where+".t")
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
	l.Link(win, id, "contains", srcSpring, RequestWindow)
	return nil
}

// check 메서드는 줄 판정 하나와 그 안의 줄·판정 경로·경로 안의 호출을 옮깁니다(트랜잭션 실험과 같은 꼴). 셋은 같은 출처라 parent로 잇는 자식입니다.
func (l *loader) check(win, proc int, raw any, where string) error {
	c, err := load.Object(raw, where)
	if err != nil {
		return err
	}
	if err := load.Keys(c, where, "ctid", "lines", "mask", "path", "snapXmax", "snapXmin", "stack", "t", "table", "tup", "visible", "xmax", "xmin"); err != nil {
		return err
	}
	group := "$.conns[].requests[].server.checks[]"
	stack, err := l.Stack(proc, c["stack"], srcRec, group, where)
	if err != nil {
		return err
	}
	if stack == nil {
		return fmt.Errorf("%s.stack이 없음", where)
	}
	t, err := l.micros(c["t"], where+".t")
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
	l.B.EventAttrs(id, srcSpring, "visible", c["visible"])
	l.Link(win, id, "contains", srcSpring, RequestWindow)
	if l.Once("unrecorded:calls") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcRec, Field: "check.calls.start_ns", Note: "줄 판정 안의 줄(line)·판정 경로(path)·경로 안의 호출(inner)은 따로 시각이 없습니다"})
	}

	lines, err := load.List(c["lines"], where+".lines")
	if err != nil {
		return err
	}
	for k, raw := range lines {
		if _, err := l.call(id, proc, nameLine, raw, fmt.Sprintf("%s.lines[%d]", where, k), group+".lines[]", "fn", "at"); err != nil {
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
		p, err := l.call(id, proc, namePath, raw, pw, group+".path[]", "fn", "at", "arg", "ret", "inner")
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
				if _, err := l.call(p, proc, nameInner, raw, fmt.Sprintf("%s.inner[%d]", pw, j), group+".path[].inner[]", "fn", "at"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// call 메서드는 줄 판정 안의 호출 하나(부른 함수 fn과 부른 줄 at)를 parent의 자식 순간으로 옮깁니다.
// 장소는 부른 함수이고 스택은 [그 함수, 부른 줄의 코드 자리]입니다(트랜잭션 실험과 같은 꼴).
func (l *loader) call(parent, proc int, name string, raw any, where, group string, allowed ...string) (int, error) {
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
	place, err := l.Callee(proc, fn, srcRec, group+".fn")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	site, err := l.Site(proc, at, srcRec, group+".at")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", where, err)
	}
	return l.Event(atlas.Event{Place: place, Kind: atlas.Instant, Name: name, Parent: &parent, Stack: []int{place, site}, Source: srcRec}, group), nil
}

// wait 메서드는 서버 기다림 하나를 백엔드의 해석기 구간으로 옮기고 그 요청의 서버 창이 품게 합니다.
//
// 끝은 오류(ereport, 컨테이너 시각)여야 합니다. 오류 없이 끝난 기다림의 끝은 요청의 done(맥 시계)이라 컨테이너 시계 구간의 끝이 될 수 없어 멈춥니다.
func (l *loader) wait(win, proc int, w map[string]any) error {
	where := "$.view.waits[]"
	if w["end"] != "error" {
		return fmt.Errorf("%s: 끝 %v인 기다림은 끝이 맥 시계라 옮길 수 없음", where, w["end"])
	}
	t0, err := l.micros(w["t0"], where+".t0")
	if err != nil {
		return err
	}
	t1, err := l.micros(w["t1"], where+".t1")
	if err != nil {
		return err
	}
	id := l.Event(atlas.Event{Place: proc, Kind: atlas.Span, Name: nameWait, StartNS: &t0, EndNS: &t1, Source: srcCont}, "$.view.waits[]")
	l.B.EventAttrs(id, srcCont, "xid", w["xid"])
	l.B.EventAttrs(id, srcCont, "end", w["end"])
	l.Link(win, id, "contains", srcSpring, RequestWindow)
	return nil
}

// captures 메서드는 관찰을 txlab 장소의 sample로(끝은 관찰용 접속이 마지막 답을 보낸 span[1]), 부분마다 실제로 읽은 구간을 그 자식 구간으로
// 옮깁니다. 관찰마다 같은 시각(tCapture)의 멈춤 자리나 막힘 관찰이 정확히 하나여야 하고 그것과 same 연결입니다. 겹친 메시지는 컨테이너 쪽
// 메시지 사건과 overlap 연결입니다.
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
		if err := load.Keys(c, where, "activity", "afterStep", "label", "locks", "overlap", "pages", "reads", "running", "span", "t", "visible", "wal", "xacts"); err != nil {
			return err
		}
		t, err := l.micros(c["t"], where+".t")
		if err != nil {
			return err
		}
		span, err := load.List(c["span"], where+".span")
		if err != nil || len(span) != 2 {
			return fmt.Errorf("%s.span이 시각 둘의 목록이 아님", where)
		}
		s0, err := l.micros(span[0], where+".span[0]")
		if err != nil {
			return err
		}
		s1, err := l.micros(span[1], where+".span[1]")
		if err != nil {
			return err
		}
		if s0 != t || s1 < s0 {
			return fmt.Errorf("%s.span[0]이 관찰 시각이 아니거나 끝이 시작보다 이름", where)
		}
		id := l.Event(atlas.Event{Place: l.client, Kind: atlas.Sample, Name: nameCapture, StartNS: &t, EndNS: &s1, Source: srcObserve}, "$.captures[]")
		for _, k := range []string{"activity", "label", "locks", "pages", "running", "wal", "xacts"} {
			l.B.EventAttrs(id, srcObserve, k, c[k])
		}
		l.B.EventAttrs(id, srcSpring, "visible", c["visible"])

		// 해석기는 끝을 모르는 마지막 관찰의 부분을 끝이 시작과 같게 두었으므로(decode/spring.go 594~597) 그 부분은 끝을 비웁니다.
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
			a, err := l.micros(pair[0], where+".reads."+p+"[0]")
			if err != nil {
				return err
			}
			b, err := l.micros(pair[1], where+".reads."+p+"[1]")
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
			l.Event(atlas.Event{Place: l.client, Kind: atlas.Span, Name: p, Parent: &id, StartNS: &a, EndNS: end, Source: srcObserve}, "$.captures[].reads."+p)
		}

		// 같은 시각에 찍었다고 txlab이 적은 멈춤 자리나 막힘 관찰 하나와 잇습니다. afterStep은 그 짝으로 다시 셉니다.
		cps, bls := l.cpCapture[t], l.blCapture[t]
		switch {
		case len(cps) == 1 && len(bls) == 0:
			l.Link(cps[0], id, "same", srcTxlab, "$.checkpoints[].tCapture")
		case len(cps) == 0 && len(bls) == 1:
			l.Link(bls[0], id, "same", srcTxlab, "$.blocked[].tCapture")
		default:
			return fmt.Errorf("%s: 같은 시각의 멈춤 자리 %d개, 막힘 관찰 %d개", where, len(cps), len(bls))
		}
		delete(l.cpCapture, t)
		delete(l.blCapture, t)

		overlap, err := load.List(c["overlap"], where+".overlap")
		if err != nil {
			return err
		}
		for k, o := range overlap {
			m, err := l.findMessage(o)
			if err != nil {
				return fmt.Errorf("%s.overlap[%d]: %w", where, k, err)
			}
			l.Link(id, m[1], "overlap", srcSpring, "$.captures[].overlap[]")
		}
	}
	if len(l.cpCapture) != 0 || len(l.blCapture) != 0 {
		return fmt.Errorf("관찰과 짝이 없는 멈춤 자리 %d개, 막힘 관찰 %d개", len(l.cpCapture), len(l.blCapture))
	}
	return nil
}

// findMessage 메서드는 메시지 사본 하나(원본 꼴)와 칸이 모두 같은 메시지를 모든 연결에서 찾습니다. 정확히 하나가 아니면 오류입니다.
func (l *loader) findMessage(o any) ([2]int, error) {
	want, err := load.Canonical(o)
	if err != nil {
		return [2]int{}, err
	}
	var found [][2]int
	names := make([]string, 0, len(l.msgs))
	for n := range l.msgs {
		names = append(names, n)
	}
	sort.Strings(names)
	r := l.B.Run()
	for _, n := range names {
		for _, m := range l.msgs[n] {
			got, err := load.Canonical(wireView(r.Events[m[0]], r.Events[m[1]]))
			if err != nil {
				return [2]int{}, err
			}
			if len(atlas.Diff("$", want, got)) == 0 {
				found = append(found, m)
			}
		}
	}
	if len(found) != 1 {
		return [2]int{}, fmt.Errorf("칸이 같은 메시지 %d개", len(found))
	}
	return found[0], nil
}

// transactions 메서드는 트랜잭션을 연결 장소의 해석기 구간으로 옮기고, 품은 요청마다 contains, 끝난 뒤 첫 앱 메시지와 next 연결을 둡니다.
// next는 해석기처럼 마지막 요청의 done보다 늦은 첫 F 메시지입니다(decode/spring.go 326~332). 그런 메시지가 없으면 멈춥니다.
func (l *loader) transactions(v any) error {
	list, err := load.List(v, "$.transactions")
	if err != nil {
		return err
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.transactions[%d]", i)
		x, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(x, where, "clog", "conn", "end", "first", "last", "next", "t0", "t1", "xids"); err != nil {
			return err
		}
		n, err := load.Integer(x["conn"], where+".conn")
		if err != nil {
			return err
		}
		name := fmt.Sprintf("C%d", n)
		place, ok := l.conn[name]
		if !ok {
			return fmt.Errorf("%s.conn %d의 연결이 없음", where, n)
		}
		first, err := load.Integer(x["first"], where+".first")
		if err != nil {
			return err
		}
		last, err := load.Integer(x["last"], where+".last")
		if err != nil {
			return err
		}
		if first < 1 || last < first || last >= len(l.reqs[name]) {
			return fmt.Errorf("%s: 요청 범위 %d~%d", where, first, last)
		}
		t0, err := l.micros(x["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := l.micros(x["t1"], where+".t1")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: place, Kind: atlas.Span, Name: nameTx, StartNS: &t0, EndNS: &t1, Source: srcMac}, "$.transactions[]")
		l.B.EventAttrs(id, srcMac, "end", x["end"])
		for s := first; s <= last; s++ {
			l.Link(id, l.reqs[name][s], "contains", srcSpring, "$.transactions[].first")
		}
		if _, ok := x["next"]; !ok {
			return fmt.Errorf("%s.next 없음(내보내기가 늘 다시 셈)", where)
		}
		next := -1
		r := l.B.Run()
		for _, m := range l.msgs[name] {
			e := r.Events[m[0]]
			if e.Target != nil && *e.StartNS > l.done[name][last] {
				next = m[0]
				break
			}
		}
		if next < 0 {
			return fmt.Errorf("%s: 끝난 뒤의 앱 메시지가 없음", where)
		}
		l.Link(id, next, "next", srcSpring, "$.transactions[].next")
	}
	return nil
}

// borrowsOf 메서드는 풀 빌리기를 스레드의 추정 구간으로 옮기고 받은 연결을 target으로 둡니다. 연결이 빈 빌리기(시간 초과)는 target이 없습니다.
func (l *loader) borrowsOf(v any) error {
	list, err := load.List(v, "$.borrows")
	if err != nil {
		return err
	}
	for i, raw := range list {
		where := fmt.Sprintf("$.borrows[%d]", i)
		b, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(b, where, "conn", "idleMs", "result", "t0", "t1", "thread", "validated"); err != nil {
			return err
		}
		names, err := texts(b, where, "thread", "conn")
		if err != nil {
			return err
		}
		th := l.threadPlace(names[0], srcGuess, "$.borrows[].thread")
		var target *int
		if names[1] == "" {
			l.noConn++
		} else {
			p, ok := l.conn[names[1]]
			if !ok {
				return fmt.Errorf("%s.conn %s가 연결이 아님", where, names[1])
			}
			target = &p
		}
		t0, err := l.micros(b["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := l.micros(b["t1"], where+".t1")
		if err != nil {
			return err
		}
		id := l.Event(atlas.Event{Place: th, Target: target, Kind: atlas.Span, Name: nameBorrow, StartNS: &t0, EndNS: &t1, Source: srcGuess}, "$.borrows[]")
		for _, k := range []string{"validated", "idleMs", "result"} {
			l.B.EventAttrs(id, srcGuess, k, b[k])
		}
		l.borrows = append(l.borrows, borrowRef{id: id, t0: t0, t1: t1, thread: names[0]})
	}
	return nil
}
