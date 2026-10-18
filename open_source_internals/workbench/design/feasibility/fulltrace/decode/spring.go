package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// 이 파일은 실험 5(Spring 전파) 기록의 해석입니다(2026-10-15). txlab의 Spring 지휘(txlab/spring.go)가 남긴 세 층을 한 시간축에 놓습니다.
//
//   - 앱: spring.json(멈춤 자리, 예외 알림, 접속)과 spring.jfr(메서드 추적, 예외, 기다림. spring_jfr.go)
//   - 연결 위: wire-mac-<n>.jsonl(맥 쪽 중계기)과 wire-<n>.jsonl(컨테이너 쪽 중계기). 형식의 정본은 txlab/wire.go의 wireMsg입니다.
//   - 서버: steps.json의 관찰, uprobe.trace, 백엔드마다의 함수 기록(<pid>.bin), all.strace
//
// 시각은 모두 spring.json의 startNs(맥 시계)부터의 마이크로초입니다. 맥 시계와 컨테이너 시계를 한 축에 놓는 근거는 연결마다 잰
// 차이 구간(clock.wire)이고, 그 구간이 1 ms를 넘으면 그렇게 놓지 말아야 합니다(WORK 실험 5의 P10).
//
// 탐색용 코드입니다. 작업대 관측 재설계가 방법을 확정하면 decode 전체와 함께 다시 설계해 옮깁니다(main.go 머리 주석).

// springFile 은 txlab의 spring.json 가운데 해석에 쓰는 부분입니다(정본은 txlab/spring.go의 springResult).
type springFile struct {
	Scene       string `json:"scene"`
	Pool        int    `json:"pool"`
	Java        string `json:"java"`
	Trace       string `json:"trace"`
	Checkpoints []struct {
		N         int    `json:"n"`
		Label     string `json:"label"`
		JavaNs    uint64 `json:"javaNs"`
		CaptureNs uint64 `json:"captureNs"`
		GoNs      uint64 `json:"goNs"`
	} `json:"checkpoints"`
	Notes []struct {
		Where   string `json:"where"`
		JavaNs  uint64 `json:"javaNs"`
		Chain   string `json:"chain"`
		Message string `json:"message"`
	} `json:"notes"`
	Conns []struct {
		Conn     int    `json:"conn"`
		Pid      int    `json:"pid"`
		AcceptNs uint64 `json:"acceptNs"`
		CloseNs  uint64 `json:"closeNs"`
		Err      string `json:"err"`
	} `json:"conns"`
	Blocked []struct {
		After     int    `json:"after"`
		Ns        uint64 `json:"ns"`
		CaptureNs uint64 `json:"captureNs"`
		Waiters   []int  `json:"waiters"`
	} `json:"blocked"`
	JavaExit int    `json:"javaExit"`
	StartNs  uint64 `json:"startNs"`
	EndNs    uint64 `json:"endNs"`
}

// wireMsg 는 중계기가 적은 메시지 한 줄입니다(정본은 txlab/wire.go의 wireMsg). Ns는 그 메시지의 끝 바이트를 읽은 시각입니다.
type wireMsg struct {
	Conn     int               `json:"conn"`
	Dir      string            `json:"dir"`
	N        int               `json:"n"`
	Type     string            `json:"type"`
	Len      int               `json:"len"`
	Ns       uint64            `json:"ns"`
	Name     string            `json:"name,omitempty"`
	SQL      string            `json:"sql,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	Tag      string            `json:"tag,omitempty"`
	Status   string            `json:"status,omitempty"`
	Pid      int               `json:"pid,omitempty"`
	Auth     *int              `json:"auth,omitempty"`
	Count    int               `json:"count,omitempty"`
	Severity string            `json:"severity,omitempty"`
	Code     string            `json:"code,omitempty"`
	Message  string            `json:"message,omitempty"`
	Where    string            `json:"where,omitempty"`
}

// wireView 는 화면과 판정에 넘기는 메시지 하나입니다. T는 맥 쪽이 읽은 시각, Tc는 컨테이너 쪽이 읽은 시각입니다.
// Text는 그 메시지의 뜻 있는 칸 하나(SQL, 태그, 상태, 오류 글)입니다.
type wireView struct {
	T    float64           `json:"t"`
	Tc   float64           `json:"tc"`
	Dir  string            `json:"dir"`
	Type string            `json:"type"`
	Len  int               `json:"len"`
	Name string            `json:"name,omitempty"`
	Text string            `json:"text,omitempty"`
	F    map[string]string `json:"f,omitempty"`
}

// wireClock 은 연결 하나의 두 시계 비교입니다. Same은 두 쪽이 메시지를 같은 순번·종류·길이로 잘랐는지이고, Lo·Hi는 (컨테이너 − 맥)
// 차이가 들어 있는 구간(밀리초)입니다. 앱 → 서버 메시지는 차이 + 지연, 서버 → 앱 메시지는 차이 − 지연이라 구간은
// [서버 → 앱 값의 최댓값, 앱 → 서버 값의 최솟값]입니다. Lo > Hi면 구간이 빈 것입니다.
type wireClock struct {
	Conn int     `json:"conn"`
	N    int     `json:"n"`
	Same bool    `json:"same"`
	Lo   float64 `json:"lo"`
	Hi   float64 `json:"hi"`
}

// springRequest 는 연결 위의 요청 하나입니다. 앱이 보낸 묶음(Sync 또는 단순 질의로 끝남)과 서버의 답(ReadyForQuery까지)입니다.
// Seq 0은 접속 인증이고 그 뒤로 1부터 셉니다. Sent·Done은 맥 쪽이 묶음의 첫 메시지와 ReadyForQuery를 읽은 시각, Status는
// ReadyForQuery의 트랜잭션 상태(I 밖, T 안, E 실패한 트랜잭션 안)입니다. Server는 백엔드가 그 요청을 처리한 구간(앞 ReadyForQuery부터
// 이번 ReadyForQuery까지)의 함수와 사건이고, 접속 인증에는 두지 않습니다. Types는 앱이 보낸 메시지 종류의 차례(Q 하나, 또는 P B D E S),
// Stmt는 그 묶음이 쓴 준비 문장의 이름(Parse 또는 Bind에 적힌 것, 이름 없는 문장이면 빈 문자열)입니다.
type springRequest struct {
	Seq    int               `json:"seq"`
	SQL    []string          `json:"sql"`
	Types  []string          `json:"types"`
	Stmt   string            `json:"stmt,omitempty"`
	Sent   float64           `json:"sent"`
	Done   float64           `json:"done"`
	Status string            `json:"status"`
	Tags   []string          `json:"tags,omitempty"`
	Error  map[string]string `json:"error,omitempty"`
	Server *txRequest        `json:"server,omitempty"`
}

// springTx 는 연결 하나에서 ReadyForQuery의 상태가 I에서 벗어났다 I로 돌아온 구간, 곧 트랜잭션 하나입니다. First·Last는 그 안의
// 첫·끝 요청의 Seq입니다. Xids는 그 구간에서 서버가 준 트랜잭션 번호를 준 순서대로 담고, Level은 그 번호를 받은 트랜잭션의 깊이
// (1이 맨 바깥, AssignTransactionId의 nestingLevel)입니다. Clog는 커밋 로그에 적힌 결과입니다. Next는 끝난 뒤 그 연결에 앱이 처음
// 보낸 메시지이고, 풀에 돌려줄 때 무언가 보냈는지를 이것으로 봅니다.
type springTx struct {
	Conn  int              `json:"conn"`
	First int              `json:"first"`
	Last  int              `json:"last"`
	T0    float64          `json:"t0"`
	T1    float64          `json:"t1"`
	End   string           `json:"end"`
	Xids  []springXid      `json:"xids"`
	Clog  []map[string]any `json:"clog"`
	Next  *wireView        `json:"next,omitempty"`
}

// springXid 는 트랜잭션 번호 하나와 그 번호를 받은 깊이입니다.
type springXid struct {
	Xid   string  `json:"xid"`
	Level string  `json:"level"`
	T     float64 `json:"t"`
}

// springConn 은 앱의 접속 하나입니다. Thread는 그 접속을 연 앱 스레드(접속 인증 직후의 첫 질의를 보낸 JFR 사건의 스레드)입니다.
type springConn struct {
	Conn      int               `json:"conn"`
	Pid       int               `json:"pid"`
	Name      string            `json:"name"`
	Thread    string            `json:"thread"`
	Startup   []string          `json:"startup"`
	Params    map[string]string `json:"params"`
	Accept    float64           `json:"accept"`
	Close     float64           `json:"close"`
	BridgeErr string            `json:"bridgeErr,omitempty"`
	LastMac   string            `json:"lastMac"`
	LastCont  string            `json:"lastCont"`
	Messages  []wireView        `json:"messages"`
	Requests  []springRequest   `json:"requests"`
	// Join은 함수 기록의 ReadyForQuery 수와 연결 위의 ReadyForQuery(Z) 수입니다. 둘이 다르면 요청마다의 서버 구간을 믿지 않습니다.
	Join [2]int `json:"join"`
}

// springCapture 는 서버 관찰 하나입니다(scenario의 txCapture에 이름과 겹침을 더함). Overlap은 관찰이 실제로 읽은 구간 안에 연결 위로
// 오간 메시지입니다. 멈춤 자리의 관찰은 앱이 멈춰 있으므로 비어야 하고, 비지 않으면 관찰 도중에 상태가 바뀌었을 수 있습니다
// (2026-10-15 18:36 5-requires-new-pool1 a·b에서 실제로 그랬음, WORK 실험 5). Span은 그 구간(captureSpans), Visible은 관찰 때 새로
// 시작한 트랜잭션이 볼 값입니다.
type springCapture struct {
	txCapture
	Label   string                         `json:"label"`
	Span    [2]float64                     `json:"span"`
	Overlap []wireView                     `json:"overlap"`
	Visible map[string][]map[string]string `json:"visible"`
}

// readWire 는 중계기 기록 한 파일을 읽습니다.
//
// Errors: 파일을 열지 못하거나 줄을 JSON으로 읽지 못하면 오류를 돌려줍니다.
func readWire(path string) ([]wireMsg, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []wireMsg
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var m wireMsg
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		out = append(out, m)
	}
	return out, sc.Err()
}

// compareClocks 는 같은 연결의 두 기록을 메시지마다 맞대어 시계 차이 구간을 잽니다(wireClock 주석).
func compareClocks(conn int, mac, cont []wireMsg) wireClock {
	c := wireClock{Conn: conn, N: len(mac), Same: len(mac) == len(cont), Lo: math.Inf(-1), Hi: math.Inf(1)}
	for i := 0; c.Same && i < len(mac); i++ {
		a, b := mac[i], cont[i]
		if a.Dir != b.Dir || a.N != b.N || a.Type != b.Type || a.Len != b.Len {
			c.Same = false
			break
		}
		d := float64(int64(b.Ns-a.Ns)) / 1e6
		if a.Dir == "B" {
			c.Lo = math.Max(c.Lo, d)
		} else {
			c.Hi = math.Min(c.Hi, d)
		}
	}
	return c
}

// wireText 는 메시지의 뜻 있는 칸 하나와 나머지 칸을 고릅니다.
func wireText(m wireMsg) (string, map[string]string) {
	f := map[string]string{}
	text := ""
	switch {
	case m.SQL != "" || m.Type == "Q" || m.Type == "P":
		text = m.SQL
	case m.Tag != "":
		text = m.Tag
	case m.Status != "":
		text = m.Status
	case m.Message != "":
		text = m.Message
	}
	if m.Code != "" {
		f["code"], f["severity"], f["where"] = m.Code, m.Severity, m.Where
	}
	if m.Pid != 0 {
		f["pid"] = fmt.Sprint(m.Pid)
	}
	if m.Auth != nil {
		f["auth"] = fmt.Sprint(*m.Auth)
	}
	if len(f) == 0 {
		f = nil
	}
	return text, f
}

// splitRequests 는 연결 하나의 메시지를 요청으로 나눕니다. 앱 쪽 메시지는 Sync(S)나 단순 질의(Q)에서 묶음이 끝나고, 서버 쪽
// ReadyForQuery(Z)가 그 묶음의 답의 끝입니다. 앱은 답을 기다리지 않고 묶음 여럿을 한 번에 보낼 수 있으므로(pgjdbc는 BEGIN 단순 질의와
// 첫 문장의 확장 질의를 한 번에 씀, WORK 실험 5 「pgjdbc 소스 확인」) 앱 쪽 묶음을 먼저 순서대로 만들고, 서버 쪽 답을 앞에서부터
// 하나씩 짝짓습니다. 접속 인증(SSLRequest·StartupMessage부터 첫 Z까지)은 Seq 0입니다. 답이 오지 않은 묶음(끝내기 앞의 것)은 버립니다.
//
// Arguments: mac과 cont는 같은 연결의 맥 쪽·컨테이너 쪽 기록(같은 순번), rel은 맥 시계의 시각을 마이크로초로 바꿉니다.
// Returns: 답이 온 요청마다의 springRequest(Server는 비어 있음)와, 요청마다 컨테이너 쪽 Z 시각입니다.
func splitRequests(mac, cont []wireMsg, rel func(uint64) float64) ([]springRequest, []uint64) {
	reqs := []springRequest{{Seq: 0, SQL: []string{"(접속 인증)"}, Types: []string{}}}
	var cur *springRequest
	for _, m := range mac {
		if m.Dir != "F" || m.Type == "X" {
			continue
		}
		if m.Type == "ssl" || m.Type == "startup" {
			if reqs[0].Sent == 0 {
				reqs[0].Sent = rel(m.Ns)
			}
			continue
		}
		if cur == nil {
			cur = &springRequest{Seq: len(reqs), Sent: rel(m.Ns)}
		}
		cur.Types = append(cur.Types, m.Type)
		if (m.Type == "P" || m.Type == "B") && cur.Stmt == "" {
			cur.Stmt = m.Name
		}
		if m.Type == "Q" || m.Type == "P" {
			cur.SQL = append(cur.SQL, m.SQL)
		}
		if m.Type == "S" || m.Type == "Q" {
			reqs = append(reqs, *cur)
			cur = nil
		}
	}
	var zc []uint64
	k := 0
	for i, m := range mac {
		if m.Dir != "B" || k >= len(reqs) {
			continue
		}
		switch m.Type {
		case "C":
			reqs[k].Tags = append(reqs[k].Tags, m.Tag)
		case "E":
			reqs[k].Error = map[string]string{"code": m.Code, "message": m.Message, "where": m.Where}
		case "Z":
			reqs[k].Done, reqs[k].Status = rel(m.Ns), m.Status
			var c uint64
			if i < len(cont) {
				c = cont[i].Ns
			}
			zc = append(zc, c)
			k++
		}
	}
	return reqs[:k], zc
}

// deriveTxs 는 요청의 ReadyForQuery 상태로 트랜잭션 구간을 찾습니다(springTx 주석). 끝난 뒤 그 연결의 다음 앱 메시지를 Next에 둡니다.
func deriveTxs(conn int, reqs []springRequest, msgs []wireView) []springTx {
	var out []springTx
	var cur *springTx
	for _, r := range reqs {
		if r.Seq == 0 {
			continue
		}
		if cur == nil && r.Status != "I" {
			cur = &springTx{Conn: conn, First: r.Seq, T0: r.Sent, Xids: []springXid{}, Clog: []map[string]any{}}
		}
		if cur != nil && r.Status == "I" {
			cur.Last, cur.T1 = r.Seq, r.Done
			if len(r.Tags) > 0 {
				cur.End = r.Tags[len(r.Tags)-1]
			}
			for k := range msgs {
				if msgs[k].Dir == "F" && msgs[k].T > r.Done {
					next := msgs[k]
					cur.Next = &next
					break
				}
			}
			out = append(out, *cur)
			cur = nil
		}
	}
	return out
}

// visibleRows 는 관찰 한 번의 페이지에서 그때 새로 시작한 트랜잭션이 볼 행 판을 고릅니다. xmin의 트랜잭션이 커밋됐고 xmax가 없거나
// 그 트랜잭션이 커밋되지 않았으면(진행 중, 되돌림) 보입니다. 커밋 여부는 관찰의 xacts(pg_xact_status)로 보고, xacts에 없는 번호는
// 그 판의 HEAP_XMIN_COMMITTED 힌트로만 커밋으로 봅니다. 진행 중인 트랜잭션의 스냅숏이 보는 값과는 다를 수 있습니다.
func visibleRows(vs []txVersion, xacts map[string]string) []map[string]string {
	committed := func(xid string, flags []string, hint string) bool {
		if s, ok := xacts[xid]; ok {
			return s == "committed"
		}
		for _, f := range flags {
			if f == hint {
				return true
			}
		}
		return false
	}
	var out []map[string]string
	for _, v := range vs {
		if !committed(v.Xmin, v.Flags, "HEAP_XMIN_COMMITTED") {
			continue
		}
		if v.Xmax != "0" && committed(v.Xmax, v.Flags, "HEAP_XMAX_COMMITTED") {
			continue
		}
		row := map[string]string{"ctid": v.Ctid}
		for k, x := range v.Values {
			row[k] = x
		}
		out = append(out, row)
	}
	return out
}

// readyForQueryEnters 는 백엔드 하나의 함수 기록에서 ReadyForQuery에 들어간 시각을 차례로 돌려줍니다. 백엔드는 요청 하나를 마칠 때마다
// 그것으로 Z를 보내므로(postgres.c PostgresMain), 두 시각 사이가 한 요청의 처리 구간입니다.
func readyForQueryEnters(t *traceFile, syms *symbols) []uint64 {
	var out []uint64
	for _, r := range t.Records {
		if r.Kind == kindEnter && syms.name(r.Fn-t.Slide) == "ReadyForQuery" {
			out = append(out, r.Ns)
		}
	}
	return out
}

// cmdSpring 은 실험 5의 기록 하나를 화면과 판정이 쓸 JSON으로 바꿉니다.
//
// Arguments: traces와 syms는 -dir(<기록>/trace)과 -bin에서 읽은 함수 기록, rec은 기록 디렉터리, jfrTool은 JDK의 jfr 명령,
// srcDir가 있으면 서버 소스 발췌를, jsrc(Java 소스 뿌리들)가 있으면 앱 코드와 라이브러리 소스 발췌를 view에 붙입니다.
// Errors: 기록 파일을 읽지 못하거나 jfr 명령이 실패하면 오류를 돌려줍니다. 짝이 맞지 않는 부분(uprobe와 함수 기록, Z와 ReadyForQuery)은
// 끝까지 해석하고 joins·conns[].join에 수를 남깁니다.
func cmdSpring(w io.Writer, traces []*traceFile, syms *symbols, rec, jfrTool, srcDir string, jsrc []string) error {
	var sp springFile
	if err := readJSON(filepath.Join(rec, "spring.json"), &sp); err != nil {
		return err
	}
	var st stepsFile
	if err := readJSON(filepath.Join(rec, "steps.json"), &st); err != nil {
		return err
	}
	evs, err := readUprobes(filepath.Join(rec, "uprobe.trace"))
	if err != nil {
		return err
	}
	sys, err := readStrace(filepath.Join(rec, "all.strace"))
	if err != nil {
		return err
	}
	base := sp.StartNs
	rel := func(ns uint64) float64 {
		if ns == 0 {
			return 0
		}
		return float64(int64(ns-base)) / 1e3
	}
	joins := joinUprobes(evs, traces, syms)
	repair := repairSites(evs, traces, syms)

	// 앱 쪽: JFR. 접속을 연 스레드를 고르는 데 쓰므로 연결보다 먼저 읽습니다. 장면의 시작은 start 멈춤 자리의 go입니다.
	from := 0.0
	for _, c := range sp.Checkpoints {
		if c.Label == "start" {
			from = rel(c.GoNs)
		}
	}
	java, err := readSpringJFR(jfrTool, filepath.Join(rec, "spring.jfr"), sp.Trace, rel, from)
	if err != nil {
		return err
	}

	// 서버 쪽 사건을 백엔드마다 시각 순으로 모읍니다(cmdScenario와 같은 방식).
	byPid := map[int][]*uevent{}
	for i := range evs {
		if evs[i].trace != nil {
			byPid[evs[i].Pid] = append(byPid[evs[i].Pid], &evs[i])
		}
	}
	for _, l := range byPid {
		sort.SliceStable(l, func(i, j int) bool { return l[i].ns < l[j].ns })
	}
	samplers := map[int]func(uint64, bool) ([]frameRef, string){}
	for _, t := range traces {
		samplers[t.Pid] = stackSampler(t, syms)
	}
	stackAt := func(e *uevent) []frameRef {
		s, _ := samplers[e.Pid](e.ns, !e.Return)
		return trimStack(s)
	}
	fname := func(e *uevent) string { return syms.name(e.Func - e.trace.Slide) }
	siteOf := func(e *uevent) string {
		r := e.trace.Records[e.rec]
		return syms.lineAt(r.Site - e.trace.Slide - 4)
	}
	oids, ssiRels := map[int]string{}, map[int]string{}
	for _, t := range st.Tables {
		oids[t.Oid], ssiRels[t.Oid] = t.Name, t.Name
	}

	// 연결마다: 두 기록, 시계, 요청, 트랜잭션.
	var conns []springConn
	var clocks []wireClock
	var txs []springTx
	var allCont []wireView
	sessions := map[int]string{}
	for _, c := range sp.Conns {
		mac, err := readWire(filepath.Join(rec, fmt.Sprintf("wire-mac-%d.jsonl", c.Conn)))
		if err != nil {
			return err
		}
		cont, err := readWire(filepath.Join(rec, fmt.Sprintf("wire-%d.jsonl", c.Conn)))
		if err != nil {
			return err
		}
		clocks = append(clocks, compareClocks(c.Conn, mac, cont))
		sc := springConn{Conn: c.Conn, Pid: c.Pid, Name: fmt.Sprintf("C%d", c.Conn), Accept: rel(c.AcceptNs), Close: rel(c.CloseNs), BridgeErr: c.Err, Messages: []wireView{}}
		sessions[c.Pid] = sc.Name
		if len(mac) > 0 {
			sc.LastMac = mac[len(mac)-1].Type
		}
		if len(cont) > 0 {
			sc.LastCont = cont[len(cont)-1].Type
		}
		var appName uint64
		authDone := false
		for i, m := range mac {
			text, f := wireText(m)
			v := wireView{T: rel(m.Ns), Dir: m.Dir, Type: m.Type, Len: m.Len, Name: m.Name, Text: text, F: f}
			if i < len(cont) {
				v.Tc = rel(cont[i].Ns)
				allCont = append(allCont, v)
			}
			sc.Messages = append(sc.Messages, v)
			switch {
			case m.Type == "startup":
				sc.Params = m.Params
			case m.Dir == "F" && m.Type == "Q" && appName == 0:
				appName = m.Ns
			}
			if !authDone {
				authDone = m.Type == "Z"
				switch m.Type {
				case "ssl", "sslreply", "startup", "R", "K", "Z":
					step := m.Type
					if m.Type == "sslreply" || m.Type == "Z" {
						step += "(" + m.Status + ")"
					} else if m.Auth != nil {
						step += fmt.Sprintf("(%d)", *m.Auth)
					}
					sc.Startup = append(sc.Startup, strings.TrimSuffix(step, "()"))
				}
			}
		}
		// 접속 인증 뒤 첫 질의(pgjdbc의 SET application_name)를 보낸 JFR 사건의 스레드가 그 접속을 연 스레드입니다.
		if appName != 0 {
			sc.Thread = java.threadAt(rel(appName), "org.postgresql.core.v3.QueryExecutorImpl::execute")
		}

		reqs, zc := splitRequests(mac, cont, rel)
		if t, err := findTrace(traces, c.Pid); err == nil {
			rfq := readyForQueryEnters(t, syms)
			sc.Join = [2]int{len(rfq), len(zc)}
			if len(rfq) == len(zc) {
				for k := 1; k < len(reqs); k++ {
					win := [2]uint64{rfq[k-1], rfq[k]}
					req := txRequest{SQL: strings.Join(reqs[k].SQL, "; "), T0: rel(win[0]), T1: rel(win[1]), Snapshots: []txSnapshot{}, Checks: []txCheck{}, Events: []txEvent{}}
					var inWin []*uevent
					for _, e := range byPid[c.Pid] {
						if e.ns > win[0] && e.ns <= win[1] {
							inWin = append(inWin, e)
						}
					}
					decodeRequest(&req, inWin, win, oids, ssiRels, fname, siteOf, stackAt, rel, syms)
					sort.SliceStable(req.Events, func(a, b int) bool { return req.Events[a].T < req.Events[b].T })
					reqs[k].Server = &req
				}
			}
		}
		sc.Requests = reqs
		for _, tx := range deriveTxs(c.Conn, reqs, sc.Messages) {
			for _, r := range reqs[tx.First : tx.Last+1] {
				if r.Server == nil {
					continue
				}
				level := "1"
				for _, e := range r.Server.Events {
					switch e.Kind {
					case "assign_xid":
						level = e.F["level"]
					case "xid":
						tx.Xids = append(tx.Xids, springXid{Xid: e.F["xid"], Level: level, T: e.T})
					case "commit_clog", "abort_clog":
						tx.Clog = append(tx.Clog, map[string]any{"kind": e.Kind, "xid": e.F["xid"], "nsub": e.F["nsub"], "t": e.T})
					}
				}
			}
			txs = append(txs, tx)
		}
		conns = append(conns, sc)
	}
	sort.SliceStable(allCont, func(i, j int) bool { return allCont[i].Tc < allCont[j].Tc })

	// 관찰: scenario와 같은 풀이에 이름, 겹침, 보이는 값을 더합니다.
	cols := map[string][][2]string{}
	for _, t := range st.Tables {
		for _, c := range t.Columns {
			cols[t.Name] = append(cols[t.Name], [2]string{c.Name, c.Type})
		}
	}
	obsPids := map[int]bool{}
	for _, p := range st.ObserverPids {
		obsPids[p] = true
	}
	var obsSys []sysEvent
	for _, e := range sys {
		if obsPids[e.Pid] {
			obsSys = append(obsSys, e)
		}
	}
	capNs := make([]uint64, len(st.Captures))
	for k, c := range st.Captures {
		capNs[k] = c.Ns
	}
	reads := captureReads(capNs, obsSys)
	spans := captureSpans(capNs, obsSys)
	var labels []string
	if err := readCaptureLabels(filepath.Join(rec, "steps.json"), &labels); err != nil {
		return err
	}
	var caps []springCapture
	for k, c := range st.Captures {
		tc := springCapture{txCapture: txCapture{AfterStep: c.AfterStep, Wal: c.Wal, T: rel(c.Ns), Pages: map[string][]txVersion{}, Activity: c.Activity, Locks: c.Locks, Running: c.Running, Xacts: c.Xacts, Reads: map[string][2]float64{}},
			Overlap: []wireView{}, Visible: map[string][]map[string]string{}}
		if k < len(labels) {
			tc.Label = labels[k]
		}
		for part, w := range reads[k] {
			if w[1] == 0 {
				w[1] = w[0]
			}
			tc.Reads[part] = [2]float64{rel(w[0]), rel(w[1])}
		}
		tc.Span = [2]float64{rel(spans[k][0]), rel(spans[k][1])}
		for _, m := range allCont {
			if m.Tc >= tc.Span[0] && m.Tc <= tc.Span[1] {
				tc.Overlap = append(tc.Overlap, m)
			}
		}
		for tbl, items := range c.Pages {
			for _, it := range items {
				v := txVersion{Block: it.Block, Lp: it.Lp, LpFlag: it.Flags, Ctid: it.Ctid, Xmin: it.Xmin, Xmax: it.Xmax, Flags: it.RawFlags, Values: map[string]string{}}
				for k, a := range it.Attrs {
					if k < len(cols[tbl]) {
						v.Values[cols[tbl][k][0]] = attrValue(a, cols[tbl][k][1])
					}
				}
				tc.Pages[tbl] = append(tc.Pages[tbl], v)
			}
			tc.Visible[tbl] = visibleRows(tc.Pages[tbl], c.Xacts)
		}
		caps = append(caps, tc)
	}

	// 멈춤 자리, 멈춘 채로 찍은 관찰, 앱이 받은 예외.
	var cps []map[string]any
	for _, c := range sp.Checkpoints {
		cps = append(cps, map[string]any{"n": c.N, "label": c.Label, "t": rel(c.JavaNs), "tCapture": rel(c.CaptureNs), "tGo": rel(c.GoNs)})
	}
	blocked := []map[string]any{}
	for _, b := range sp.Blocked {
		// 기다리는 세션이 없으면 기록에는 null이 남지만(물음은 성공함), 화면이 한 가지 뜻만 읽게 빈 목록으로 냅니다.
		names := []string{}
		for _, p := range b.Waiters {
			names = append(names, sessions[p])
		}
		blocked = append(blocked, map[string]any{"after": b.After, "t": rel(b.Ns), "tCapture": rel(b.CaptureNs), "waiters": names})
	}
	notes := []map[string]any{}
	for _, n := range sp.Notes {
		notes = append(notes, map[string]any{"where": n.Where, "t": rel(n.JavaNs), "chain": n.Chain, "message": n.Message})
	}

	borrows := java.borrows(conns)

	// 앱 코드의 소스는 기록에 남은 사본(txlab의 appSrcDir, "app-src")을 먼저 읽습니다. JFR 스택의 줄 번호와 맞는 것은 그 사본뿐입니다.
	// 사본이 없는 기록(그 전의 txlab으로 만든 것)은 저장소 소스를 쓰고 javaFrom을 repo로 남겨, 줄이 어긋날 수 있음을 화면이 알리게 합니다.
	javaFrom := "repo"
	if fi, err := os.Stat(filepath.Join(rec, "app-src")); err == nil && fi.IsDir() {
		jsrc = append([]string{filepath.Join(rec, "app-src")}, jsrc...)
		javaFrom = "record"
	}
	view := springViewOf(java, conns, borrows, sp, rel, jsrc)
	view.JavaFrom = javaFrom
	for _, m := range view.SiteMiss {
		fmt.Fprintf(os.Stderr, "앱 코드가 부른 줄의 소스에 다음 칸이 없음(소스와 기록이 다른 빌드일 수 있음): %s\n", m)
	}
	out := map[string]any{
		"scene": sp.Scene, "pool": sp.Pool, "javaVersion": firstLine(sp.Java), "javaExit": sp.JavaExit, "title": st.Title,
		"tables": st.Tables, "sessions": sessions, "checkpoints": cps, "blocked": blocked, "notes": notes,
		"conns": conns, "transactions": txs, "borrows": borrows, "captures": caps,
		"clock": map[string]any{"wire": clocks, "uprobe": uprobeClock(evs)}, "joins": joins, "siteRepair": repair,
		"java": java.view(), "view": view,
	}
	if srcDir != "" {
		var refs []string
		for _, c := range conns {
			for _, r := range c.Requests {
				if r.Server == nil {
					continue
				}
				for _, e := range r.Server.Events {
					for _, f := range e.Stack {
						refs = append(refs, f.At)
					}
				}
			}
		}
		out["src"] = excerpts(srcDir, refs)
	}
	return json.NewEncoder(w).Encode(out)
}

// springView 는 화면이 고르고 그리기만 하도록 해석기가 미리 셈한 값입니다(WORK 실험 5 「실험 5 화면 설계」). Window는 장면 구간
// (start의 go ~ end 도착), Tree·Roots는 main 스레드의 호출 트리(spring_tree.go), Binding·Slots는 연결 묶임과 풀 사용(bindingReplay,
// 셈한 값), Waits는 서버의 기다림, Pauses는 실험 도구가 앱을 멈춘 구간입니다. Java는 앱 코드(lab.*)의 소스 전체, Jsrc는 부른 줄
// 둘레의 라이브러리 소스 발췌("클래스:줄")입니다. Branches·Compared는 spring-compare가 채웁니다.
type springView struct {
	Window    [2]float64             `json:"window"`
	Tree      []*callNode            `json:"tree"`
	Roots     []int                  `json:"roots"`
	Ambiguous int                    `json:"ambiguous"`
	Binding   []bindState            `json:"binding"`
	Slots     []poolSlot             `json:"slots"`
	Waits     []serverWait           `json:"waits"`
	Pauses    [][2]float64           `json:"pauses"`
	Java      map[string]javaFile    `json:"java"`
	Jsrc      map[string]excerpt     `json:"jsrc"`
	Branches  map[string]*branchInfo `json:"branches,omitempty"`
	Compared  []string               `json:"compared,omitempty"`
	JavaFrom  string                 `json:"javaFrom"`
	SiteMiss  []string               `json:"siteMismatch"`
	Cycles    []waitCycle            `json:"waitCycles"`
	PoolConf  poolConfig             `json:"poolConfig"`
}

// javaFile 은 앱 코드 소스 파일 하나입니다. Path는 소스 뿌리 아래의 상대경로입니다.
type javaFile struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

// springViewOf 는 앱 쪽 기록과 연결로 springView를 만듭니다. 장면 구간과 멈춤 구간은 멈춤 자리에서 읽습니다. 끝 자리 이름 "end"는
// txlab의 endCheckpoint(txlab/spring.go)와 같아야 합니다.
func springViewOf(java *springJava, conns []springConn, borrows []springBorrow, sp springFile, rel func(uint64) float64, jsrc []string) springView {
	var from, to float64
	pauses := [][2]float64{}
	for _, c := range sp.Checkpoints {
		switch c.Label {
		case "start":
			from = rel(c.GoNs)
		case "end":
			to = rel(c.JavaNs)
		}
		if c.GoNs > c.JavaNs {
			pauses = append(pauses, [2]float64{rel(c.JavaNs), rel(c.GoNs)})
		}
	}
	var src *javaSources
	if len(jsrc) > 0 {
		src = &javaSources{roots: jsrc, files: map[string][]string{}}
	}
	tree := buildCallTree(java, sceneThread, from, to)
	binding, slots := bindingReplay(tree, borrows, from)
	annotateTree(tree, java, conns, borrows, pauses, src)

	// 소스: 앱 코드는 파일 전체, 라이브러리는 부른 줄과 잠든 자리·예외를 만든 자리의 둘레만.
	v := springView{Window: [2]float64{from, to}, Tree: tree.Nodes, Roots: tree.Roots, Ambiguous: tree.Ambiguous, Binding: binding,
		Slots: slots, Waits: serverWaits(conns), Pauses: pauses, Java: map[string]javaFile{}, Jsrc: map[string]excerpt{}}
	v.Cycles = append(waitCycles(tree, v.Waits, binding, conns), poolWaitCycles(tree, slots, binding, sp.Pool)...)
	if src == nil {
		return v
	}
	var refs []string
	for _, n := range tree.Nodes {
		cls, _, _ := strings.Cut(n.Fn, "::")
		if strings.HasPrefix(n.Fn, "lab.") {
			if _, done := v.Java[cls]; !done {
				if lines := src.lines(cls); lines != nil {
					v.Java[cls] = javaFile{Path: fileOf(cls), Lines: lines}
				}
			}
		}
		if n.Parent >= 0 && n.Site > 0 && !strings.HasPrefix(tree.Nodes[n.Parent].Fn, "lab.") {
			refs = append(refs, fmt.Sprintf("%s:%d", tree.Nodes[n.Parent].Fn, n.Site))
		}
		for _, e := range n.Effects {
			if e.At != "" {
				refs = append(refs, e.At)
			}
		}
	}
	// 풀 설정이 적힌 줄도 발췌해 둡니다(화면의 빌리기·생존 검사 설명이 가리킴). 앱 설정(lab.LabConfig)은 호출 트리에 없어 앱 코드 창에
	// 나오지 않으므로 함께 발췌합니다. excerptsFor는 "클래스::메서드:줄"을 받으므로 메서드 자리를 채웁니다.
	v.PoolConf = poolConfigOf(src)
	for _, at := range []string{v.PoolConf.ConnectionTimeoutAt, v.PoolConf.AliveBypassAt, v.PoolConf.AliveCheckAt} {
		if c, line, ok := strings.Cut(at, ":"); ok {
			refs = append(refs, c+"::-:"+line)
		}
	}
	v.Jsrc = src.excerptsFor(refs)
	v.SiteMiss = checkAppSites(tree, src)
	return v
}

// captureSpans 는 관찰마다 관찰용 접속이 실제로 일한 구간입니다. 시작은 txlab이 묻기 시작한 시각이고, 끝은 그 관찰의 마지막 질의에
// 답을 다 보낸 sendto가 끝난 시각입니다. captureReads의 끝(다음 질의를 받은 때, 없으면 다음 관찰의 시작)은 위쪽 경계라서, 마지막
// 질의 뒤 다음 관찰까지의 시간 전체가 들어가 겹침을 가리는 데 쓰지 못합니다. 관찰이 아닌 질의(capturePart가 빈 문자열)와 그 답은
// 세지 않습니다. 멈춘 채로 찍기 전에 묻는 waiters가 앞 관찰의 구간을 늘리지 않게 하려는 것입니다.
//
// Arguments: capNs는 관찰마다 묻기 시작한 시각(오름차순), obs는 관찰용 접속의 strace 사건(시각 순)입니다.
// Returns: 관찰마다 [시작, 끝] 나노초입니다. 그 관찰 안에서 질의를 받은 적이 없으면 끝은 시작과 같습니다.
func captureSpans(capNs []uint64, obs []sysEvent) [][2]uint64 {
	out := make([][2]uint64, len(capNs))
	for k, c := range capNs {
		out[k] = [2]uint64{c, c}
		var next uint64 = math.MaxUint64
		if k+1 < len(capNs) {
			next = capNs[k+1]
		}
		inPart := false
		for _, e := range obs {
			if e.Ns < c || e.Ns >= next || e.Ret <= 0 {
				continue
			}
			if e.Name == "recvfrom" {
				if text := queryText(e.Data); text != "" {
					inPart = capturePart(text) != ""
				}
			}
			if inPart && (e.Name == "recvfrom" || e.Name == "sendto") {
				out[k][1] = e.Ns
			}
		}
	}
	return out
}

// readJSON 은 JSON 파일 하나를 v로 읽습니다.
func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// readCaptureLabels 는 steps.json의 관찰마다 이름(멈춤 자리 이름 또는 blocked)을 읽습니다. stepsFile은 psql 시나리오와 함께 쓰는
// 형이라 이름 칸을 두지 않았습니다.
func readCaptureLabels(path string, labels *[]string) error {
	var v struct {
		Captures []struct {
			Label string `json:"label"`
		} `json:"captures"`
	}
	if err := readJSON(path, &v); err != nil {
		return err
	}
	for _, c := range v.Captures {
		*labels = append(*labels, c.Label)
	}
	return nil
}

// firstLine 은 여러 줄 글의 첫 줄입니다(java -version의 판 줄).
func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
