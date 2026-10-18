package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 서버의 접속 자리는 serve가 띄우는 기록 서버와 같습니다.
const (
	labHost = "/tmp/pgi"
	labPort = 5433
)

// stepRecord 는 실행한 단계 하나입니다. 시각은 컨테이너의 CLOCK_REALTIME 나노초라서 strace(-ttt)와 함수 기록과 같은 시계입니다.
// SentNs는 psql에 문장을 쓴 시각, DoneNs는 끝 표지를 읽은 시각, BlockedNs는 관찰용 접속이 그 세션의 락 기다림을 처음 본 시각입니다.
type stepRecord struct {
	N         int      `json:"n"`
	Session   string   `json:"session"`
	SQL       string   `json:"sql"`
	SentNs    int64    `json:"sentNs"`
	DoneNs    int64    `json:"doneNs"`
	BlockedNs int64    `json:"blockedNs,omitempty"`
	Output    []string `json:"output"`
}

// pauseRecord 는 시나리오의 쉼 하나입니다. BeforeStep은 쉼 뒤에 보낸 단계의 번호(1부터)입니다.
type pauseRecord struct {
	BeforeStep int   `json:"beforeStep"`
	StartNs    int64 `json:"startNs"`
	EndNs      int64 `json:"endNs"`
}

// killRecord 는 시나리오의 kill 하나입니다. BeforeStep은 kill 뒤에 보낸 단계의 번호(1부터, 시나리오가 kill로 끝나면 단계 수 + 1)이고,
// Ns는 SIGKILL을 보내기 직전, ReadyNs는 server.log에서 서버가 다시 접속을 받는다는 줄을 본 시각입니다(50 ms 간격으로 봄).
type killRecord struct {
	BeforeStep int    `json:"beforeStep"`
	Session    string `json:"session"`
	Pid        int    `json:"pid"`
	Ns         int64  `json:"ns"`
	ReadyNs    int64  `json:"readyNs"`
}

// sessionRecord 는 세션 하나와 그 백엔드의 컨테이너 안 pid입니다.
type sessionRecord struct {
	Name string `json:"name"`
	Pid  int    `json:"pid"`
}

// tableInfo 는 관찰할 표의 번호(oid)와 열, 그 표의 인덱스입니다. 해석기가 페이지 안 행의 바이트를 값으로 풀 때와,
// SIRead 락의 대상이 어느 인덱스의 페이지인지 이름을 붙일 때 씁니다(실험 4).
type tableInfo struct {
	Name    string      `json:"name"`
	Oid     int         `json:"oid"`
	Columns []column    `json:"columns"`
	Indexes []indexInfo `json:"indexes,omitempty"`
}

// indexInfo 는 관찰할 표에 걸린 인덱스 하나의 이름과 번호입니다.
type indexInfo struct {
	Name string `json:"name"`
	Oid  int    `json:"oid"`
}

type column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// capture 는 단계 하나가 끝난 뒤(또는 락을 기다리기 시작한 뒤) 관찰용 접속이 읽은 상태입니다.
// 표마다 페이지의 줄, 세션마다의 상태, 락, 그 순간 진행 중인 트랜잭션(관찰용 접속의 스냅숏), WAL의 위치, 나온 번호의 커밋 로그 상태입니다.
// AfterKill이 참이면 kill 뒤 서버가 다시 뜬 다음 찍은 관찰이고, AfterStep은 kill 앞의 마지막 단계입니다(그 단계 뒤의 관찰과 번호가 같음).
// Label은 실험 5에서 앱의 멈춤 자리 이름(또는 앱이 멈춘 채 기다리는 동안 찍은 "blocked")이고, 그때 AfterStep은 멈춤 자리의 순번입니다.
type capture struct {
	AfterStep int                   `json:"afterStep"`
	AfterKill bool                  `json:"afterKill,omitempty"`
	Label     string                `json:"label,omitempty"`
	Ns        int64                 `json:"ns"`
	Pages     map[string][]pageItem `json:"pages"`
	Activity  []activity            `json:"activity"`
	Locks     []lockRow             `json:"locks"`
	Running   string                `json:"running"`
	Wal       *walPos               `json:"wal,omitempty"`
	Xacts     map[string]string     `json:"xacts"`
}

// walPos 는 WAL의 세 위치입니다. Insert는 WAL 버퍼에 레코드를 넣은 끝, Write는 운영 체제에 쓴 끝, Flush는 디스크에 내린 끝입니다
// (pg_current_wal_insert_lsn, pg_current_wal_lsn, pg_current_wal_flush_lsn). 비동기 커밋은 Insert만 앞서 나가고 Flush가 뒤따릅니다.
type walPos struct {
	Insert string `json:"insert"`
	Write  string `json:"write"`
	Flush  string `json:"flush"`
}

type pageItem struct {
	Block     int      `json:"block"`
	Lp        int      `json:"lp"`
	Flags     int      `json:"lpFlags"`
	Xmin      string   `json:"xmin"`
	Xmax      string   `json:"xmax"`
	Ctid      string   `json:"ctid"`
	Infomask  string   `json:"infomask"`
	Infomask2 string   `json:"infomask2"`
	RawFlags  []string `json:"rawFlags"`
	Attrs     []string `json:"attrs"`
}

type activity struct {
	Pid       int    `json:"pid"`
	Xid       string `json:"xid"`
	Xmin      string `json:"xmin"`
	State     string `json:"state"`
	WaitType  string `json:"waitType"`
	WaitEvent string `json:"waitEvent"`
}

type lockRow struct {
	Pid      int    `json:"pid"`
	Type     string `json:"type"`
	Relation string `json:"relation"`
	Page     string `json:"page"`
	Tuple    string `json:"tuple"`
	Xid      string `json:"xid"`
	Vxid     string `json:"vxid"`
	Mode     string `json:"mode"`
	Granted  bool   `json:"granted"`
}

// runResult 는 run이 남기는 steps.json 전체입니다. Observer는 처음 연 관찰용 접속이고, ObserverPids는 kill 뒤 새로 연 것까지
// 연 순서대로 모은 것입니다(앞선 기록을 읽는 해석기를 위해 Observer를 남김).
type runResult struct {
	Title        string          `json:"title"`
	Scenario     string          `json:"scenario"`
	Tables       []tableInfo     `json:"tables"`
	Sessions     []sessionRecord `json:"sessions"`
	Steps        []stepRecord    `json:"steps"`
	Pauses       []pauseRecord   `json:"pauses,omitempty"`
	Kills        []killRecord    `json:"kills,omitempty"`
	Captures     []capture       `json:"captures"`
	Observer     int             `json:"observerPid"`
	ObserverPids []int           `json:"observerPids"`
	StartNs      int64           `json:"startNs"`
	EndNs        int64           `json:"endNs"`
}

// session 은 실행 중인 세션 하나입니다. pending은 보냈지만 끝 표지를 아직 받지 못한 단계 번호(0이면 없음)입니다.
type session struct {
	sessionRecord
	p       *psql
	pending int
}

// cmdRun 은 기록 서버가 떠 있는 컨테이너 안에서 시나리오를 실행하고 <out>/steps.json을 씁니다.
//
// 순서: 관찰용 접속을 열고 → 첫 구간의 세션마다 psql을 띄워 백엔드 pid를 받고 → 단계마다 문장을 보내고, 모든 세션이 끝났거나 락을
// 기다리는 상태가 되면(settle) 관찰합니다. 락을 기다리는 세션은 나중 단계가 그 락을 풀면 끝나고, 그 출력은 원래 단계에 붙습니다.
// 단계 앞에 kill이 있으면 그 세션의 백엔드를 죽이고(crash), 서버가 다시 뜬 뒤 관찰 하나를 찍고, 다음 구간의 세션을 접속합니다.
//
// Errors: 서버에 접속하지 못하거나, 세션이 20초 안에 끝나지도 기다리지도 않거나, kill 뒤 서버가 다시 뜨지 않으면 오류를 돌려줍니다.
func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	scPath := fs.String("scenario", "", "시나리오 파일")
	out := fs.String("out", "", "출력 디렉터리")
	if err := fs.Parse(args); err != nil {
		return err
	}
	sc, err := parseScenario(*scPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	o, err := openObserver()
	if err != nil {
		return err
	}
	// kill 뒤에는 o.p가 새 접속으로 바뀌므로, 끝낼 때 그때의 접속을 끝냅니다.
	defer func() { o.p.quit() }()

	res := runResult{Title: sc.Title, Scenario: filepath.Base(*scPath), Observer: o.pid, ObserverPids: []int{o.pid}, StartNs: time.Now().UnixNano()}
	for _, t := range sc.Watch {
		ti, err := o.table(t)
		if err != nil {
			return err
		}
		res.Tables = append(res.Tables, ti)
	}

	// 세션은 구간(kill과 kill 사이)마다 그 구간의 첫 단계 전에 한꺼번에 접속합니다. live는 지금 구간의 세션입니다.
	byName := map[string]*session{}
	defer func() {
		for _, s := range byName {
			s.p.quit()
		}
	}()
	var live []*session
	connectEpoch := func(epoch int) error {
		live = nil
		for _, name := range sc.Sessions {
			if sc.Epoch[name] != epoch {
				continue
			}
			s, err := connectSession(name)
			if err != nil {
				return err
			}
			byName[name] = s
			live = append(live, s)
			res.Sessions = append(res.Sessions, s.sessionRecord)
		}
		o.pids = nil
		for _, s := range live {
			o.pids = append(o.pids, s.Pid)
		}
		return nil
	}
	if err := connectEpoch(0); err != nil {
		return err
	}

	// 시나리오 전의 상태를 0번 관찰로 남깁니다.
	c0, err := o.capture(0, sc.Watch)
	if err != nil {
		return err
	}
	res.Captures = append(res.Captures, c0)

	// i번째 단계 앞에서 쉼, kill, 단계 순서로 실행합니다. i가 단계 수와 같은 마지막 차례는 시나리오 끝의 쉼과 kill만 실행합니다.
	epoch := 0
	for i := 0; i <= len(sc.Steps); i++ {
		// 쉼은 그대로 잠듭니다. 기다리는 세션의 교착 검사(deadlock_timeout)나 WAL writer의 주기(wal_writer_delay)가 그 사이에 돌게 하려는 것입니다.
		if d := sc.Pauses[i]; d > 0 {
			p := pauseRecord{BeforeStep: i + 1, StartNs: time.Now().UnixNano()}
			time.Sleep(d)
			p.EndNs = time.Now().UnixNano()
			res.Pauses = append(res.Pauses, p)
		}
		if name, ok := sc.Kills[i]; ok {
			k, err := crash(byName[name], live, o, res.Steps)
			if err != nil {
				return fmt.Errorf("단계 %d 앞의 kill: %w", i+1, err)
			}
			k.BeforeStep = i + 1
			res.Kills = append(res.Kills, k)
			res.ObserverPids = append(res.ObserverPids, o.pid)
			// 다시 뜬 서버를 어느 세션도 건드리기 전에 관찰합니다. 이 구간의 세션은 그 뒤에 접속합니다.
			live, o.pids = nil, nil
			c, err := o.capture(i, sc.Watch)
			if err != nil {
				return err
			}
			c.AfterKill = true
			res.Captures = append(res.Captures, c)
			epoch++
			if err := connectEpoch(epoch); err != nil {
				return err
			}
		}
		if i == len(sc.Steps) {
			break
		}
		st := sc.Steps[i]
		// 이 세션이 앞 단계를 아직 끝내지 않았으면(교착으로 오류를 받을 세션 등) 끝날 때까지 기다린 뒤 보냅니다.
		// psql에 먼저 써 두면 앞 문장이 끝나자마자 실행되지만, 그러면 단계의 보낸 시각과 출력이 어느 단계 것인지 갈리지 않습니다.
		s := byName[st.Session]
		if s.pending != 0 {
			if err := waitIdle(s, live, res.Steps, 20*time.Second); err != nil {
				return fmt.Errorf("단계 %d: %w", i+1, err)
			}
		}
		rec := stepRecord{N: i + 1, Session: st.Session, SQL: st.SQL, SentNs: time.Now().UnixNano()}
		res.Steps = append(res.Steps, rec)
		if err := s.p.send(st.SQL, marker(i+1)); err != nil {
			return err
		}
		s.pending = i + 1
		if err := settle(live, o, res.Steps, 20*time.Second); err != nil {
			return fmt.Errorf("단계 %d: %w", i+1, err)
		}
		c, err := o.capture(i+1, sc.Watch)
		if err != nil {
			return err
		}
		res.Captures = append(res.Captures, c)
	}
	for _, s := range live {
		if s.pending != 0 {
			return fmt.Errorf("세션 %s의 단계 %d가 시나리오 끝까지 끝나지 않았습니다", s.Name, s.pending)
		}
	}
	res.EndNs = time.Now().UnixNano()
	b, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*out, "steps.json"), b, 0o644)
}

// connectSession 은 세션 하나의 psql을 띄우고 백엔드 pid를 받습니다. pid는 그 세션의 첫 요청(SELECT pg_backend_pid())으로 받고,
// 해석기는 시나리오 단계보다 앞선 이 요청을 건너뜁니다.
//
// Errors: psql을 띄우지 못하거나 30초 안에 pid를 받지 못하면 오류를 돌려줍니다.
func connectSession(name string) (*session, error) {
	p, err := startPsql(labHost, labPort)
	if err != nil {
		return nil, err
	}
	if err := p.send(`SELECT pg_backend_pid() AS pid \gset`+"\n"+`\echo __TXLAB_PID__ :pid`, "__TXLAB_READY__"); err != nil {
		p.quit()
		return nil, err
	}
	lines, err := p.until("__TXLAB_READY__", 30*time.Second)
	if err != nil {
		p.quit()
		return nil, err
	}
	pid := 0
	for _, l := range lines {
		if v, ok := strings.CutPrefix(l, "__TXLAB_PID__ "); ok {
			pid, _ = strconv.Atoi(v)
		}
	}
	if pid == 0 {
		p.quit()
		return nil, fmt.Errorf("세션 %s의 백엔드 pid를 받지 못했습니다: %q", name, lines)
	}
	return &session{sessionRecord: sessionRecord{Name: name, Pid: pid}, p: p}, nil
}

// crash 는 세션 s의 백엔드를 SIGKILL로 죽이고, 서버가 다시 접속을 받을 때까지 기다린 뒤 관찰용 접속을 새로 엽니다.
//
// 서버가 다시 뜨기 전에는 이 구간의 psql(죽인 세션, 다른 세션, 관찰용 접속)에 아무것도 보내지 않습니다. \q도 psql이 백엔드에
// 종료 메시지를 보내게 해서, postmaster가 그 백엔드를 내리기 전에 스스로 끝나는 일을 만들 수 있기 때문입니다.
// 다시 떴는지는 server.log의 "ready to accept connections" 줄 수가 kill 앞보다 늘었는지로 봅니다. pg_isready는 postmaster가
// 자식의 죽음을 처리하기 전 잠깐 동안에도 성공할 수 있어서 쓰지 않습니다.
//
// Returns: BeforeStep을 뺀 kill 기록입니다. o는 새 접속으로 바뀝니다.
// Errors: 이 구간에 끝나지 않은 단계가 있거나, 신호를 보내지 못하거나, 60초 안에 서버가 다시 뜨지 않거나, 관찰용 접속을 열지 못하면
// 오류를 돌려줍니다.
func crash(s *session, live []*session, o *observer, steps []stepRecord) (killRecord, error) {
	// 락을 기다리던 단계가 남은 채로 죽이면 그 단계의 끝과 출력이 정해지지 않으므로 받지 않습니다.
	if _, err := drain(live, steps); err != nil {
		return killRecord{}, err
	}
	for _, x := range live {
		if x.pending != 0 {
			return killRecord{}, fmt.Errorf("세션 %s의 단계 %d가 끝나지 않았습니다", x.Name, x.pending)
		}
	}
	log, err := os.ReadFile(labDir + "/server.log")
	if err != nil {
		return killRecord{}, err
	}
	before := readyCount(log)
	k := killRecord{Session: s.Name, Pid: s.Pid, Ns: time.Now().UnixNano()}
	if err := syscall.Kill(s.Pid, syscall.SIGKILL); err != nil {
		return k, fmt.Errorf("pid %d에 SIGKILL: %w", s.Pid, err)
	}
	if k.ReadyNs, err = waitRestart(before, 60*time.Second); err != nil {
		return k, err
	}
	for _, x := range live {
		x.p.quit()
	}
	return k, o.reopen()
}

// readyCount 는 서버 로그에서 postmaster가 접속을 받기 시작했다고 남긴 줄의 수입니다(postmaster.c 2332행).
// 처음 뜰 때 한 줄, 크래시 뒤 복구를 마치고 다시 받을 때마다 한 줄씩 늘어납니다.
func readyCount(log []byte) int {
	return bytes.Count(log, []byte("database system is ready to accept connections"))
}

// waitRestart 는 server.log의 접속 받기 줄이 before보다 많아질 때까지 50 ms마다 보고, 그것을 본 시각을 돌려줍니다.
//
// Errors: 로그를 읽지 못하거나 timeout 안에 줄이 늘지 않으면 오류를 돌려줍니다.
func waitRestart(before int, timeout time.Duration) (int64, error) {
	deadline := time.Now().Add(timeout)
	for {
		log, err := os.ReadFile(labDir + "/server.log")
		if err != nil {
			return 0, err
		}
		if readyCount(log) > before {
			return time.Now().UnixNano(), nil
		}
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("kill 뒤 %s 안에 서버가 다시 접속을 받지 않았습니다", timeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// marker 는 단계 n의 끝 표지입니다.
func marker(n int) string { return fmt.Sprintf("__TXLAB_STEP_%d__", n) }

// settle 은 모든 세션이 보낸 단계를 끝냈거나 락을 기다리는 상태가 될 때까지 기다립니다.
// 세션의 출력은 그 세션이 기다리던 단계에 붙고, 표지를 읽은 시각이 그 단계의 DoneNs가 됩니다.
//
// 락 기다림은 관찰용 접속이 pg_stat_activity의 wait_event_type = 'Lock'으로 봅니다. 문장이 아직 락까지 가지 않은 순간에는
// 그 값이 비어 있으므로, 끝나지도 기다리지도 않은 세션이 있으면 100 ms 뒤에 다시 봅니다.
//
// Errors: timeout 안에 정리되지 않거나 psql이 끝나면 오류를 돌려줍니다.
func settle(sessions []*session, o *observer, steps []stepRecord, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		busy, err := drain(sessions, steps)
		if err != nil {
			return err
		}
		if !busy {
			return nil
		}
		waiting, err := o.lockWaiters()
		if err != nil {
			return err
		}
		stuck := false
		for _, s := range sessions {
			if s.pending == 0 {
				continue
			}
			if waiting[s.Pid] {
				if rec := &steps[s.pending-1]; rec.BlockedNs == 0 {
					rec.BlockedNs = time.Now().UnixNano()
				}
				continue
			}
			stuck = true
		}
		if !stuck {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s 안에 끝나지도 락을 기다리지도 않은 세션이 있습니다", timeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// drain 은 세션마다 지금까지 온 출력 줄을 그 세션이 기다리던 단계에 붙이고, 끝 표지를 읽은 단계를 끝냅니다.
// 기다리지 않고 바로 돌아오며, 아직 끝나지 않은 단계가 있으면 busy가 참입니다.
//
// Errors: 어느 세션의 psql이 끝났으면 오류를 돌려줍니다.
func drain(sessions []*session, steps []stepRecord) (busy bool, err error) {
	for _, s := range sessions {
		if s.pending == 0 {
			continue
		}
		rec := &steps[s.pending-1]
	lines:
		for {
			select {
			case l, ok := <-s.p.lines:
				if !ok {
					return false, fmt.Errorf("세션 %s의 psql이 끝났습니다", s.Name)
				}
				if l == marker(s.pending) {
					rec.DoneNs = time.Now().UnixNano()
					s.pending = 0
					break lines
				}
				rec.Output = append(rec.Output, l)
			default:
				break lines
			}
		}
		if s.pending != 0 {
			busy = true
		}
	}
	return busy, nil
}

// waitIdle 은 세션 s의 앞 단계가 끝날 때까지 모든 세션의 출력을 받으며 기다립니다. 다른 세션도 같은 순간에 깨어나 끝날 수 있어서
// (교착의 피해자가 오류를 받으면 상대가 깨어남) s만 보지 않고 모두 받습니다.
//
// Errors: timeout 안에 끝나지 않거나 psql이 끝나면 오류를 돌려줍니다.
func waitIdle(s *session, sessions []*session, steps []stepRecord, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := drain(sessions, steps); err != nil {
			return err
		}
		if s.pending == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("세션 %s의 단계 %d가 %s 안에 끝나지 않았습니다", s.Name, s.pending, timeout)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// observer 는 관찰용 접속입니다. 시나리오의 세션과 다른 백엔드라서 그 세션들의 트랜잭션에 끼어들지 않습니다.
// 페이지는 get_raw_page가 만든 사본을 읽으므로 힌트 비트를 쓰지 않습니다(실험 1의 E8이 이것을 확인합니다).
// pids는 지금 구간의 세션 백엔드이고, n은 끝 표지의 번호입니다(접속을 새로 열어도 이어 셉니다).
type observer struct {
	p    *psql
	pid  int
	pids []int
	n    int
}

// openObserver 는 관찰용 접속을 열고 그 백엔드 pid를 받습니다. 기록을 마친 뒤 이 pid의 함수 기록을 지우므로 pid는 꼭 받아야 합니다.
//
// Errors: psql을 띄우지 못하거나 pid를 받지 못하면 오류를 돌려줍니다.
func openObserver() (*observer, error) {
	p, err := startPsql(labHost, labPort, "-A", "-t", "-F", "\t", "-q")
	if err != nil {
		return nil, err
	}
	o := &observer{p: p}
	rows, err := o.q(`SELECT pg_backend_pid()`)
	if err == nil && (len(rows) == 0 || len(rows[0]) == 0) {
		err = fmt.Errorf("관찰용 접속의 pid가 비었습니다")
	}
	if err == nil {
		o.pid, err = strconv.Atoi(rows[0][0])
	}
	if err != nil {
		p.quit()
		return nil, err
	}
	return o, nil
}

// reopen 은 지금의 관찰용 접속을 끝내고 새로 엽니다. kill 뒤 서버가 다시 뜨면 앞의 접속은 끊겨 있습니다.
func (o *observer) reopen() error {
	o.p.quit()
	n, err := openObserver()
	if err != nil {
		return err
	}
	o.p, o.pid = n.p, n.pid
	return nil
}

// q 는 관찰 쿼리 하나를 실행합니다. psql은 세미콜론을 만나야 쿼리를 보내므로 없으면 붙입니다(없으면 뒤의 \echo만 먼저 실행됩니다).
func (o *observer) q(sql string) ([][]string, error) {
	if !strings.HasSuffix(strings.TrimSpace(sql), ";") {
		sql += ";"
	}
	o.n++
	return o.p.query(sql, fmt.Sprintf("__TXLAB_OBS_%d__", o.n))
}

// table 은 표의 oid와 열 이름·형식, 그 표의 인덱스를 읽습니다.
func (o *observer) table(name string) (tableInfo, error) {
	rows, err := o.q(fmt.Sprintf(`SELECT c.oid, a.attname, format_type(a.atttypid, a.atttypmod) FROM pg_class c JOIN pg_attribute a ON a.attrelid = c.oid
WHERE c.oid = %s::regclass AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum`, quote(name)))
	if err != nil {
		return tableInfo{}, err
	}
	if len(rows) == 0 {
		return tableInfo{}, fmt.Errorf("표 %s를 찾지 못했습니다", name)
	}
	ti := tableInfo{Name: name}
	ti.Oid, _ = strconv.Atoi(rows[0][0])
	for _, r := range rows {
		ti.Columns = append(ti.Columns, column{Name: r[1], Type: r[2]})
	}
	rows, err = o.q(fmt.Sprintf(`SELECT indexrelid, indexrelid::regclass::text FROM pg_index WHERE indrelid = %d ORDER BY 1`, ti.Oid))
	if err != nil {
		return tableInfo{}, err
	}
	for _, r := range rows {
		oid, _ := strconv.Atoi(r[0])
		ti.Indexes = append(ti.Indexes, indexInfo{Name: r[1], Oid: oid})
	}
	return ti, nil
}

// lockWaiters 는 세션 pid 가운데 지금 락을 기다리는 것을 돌려줍니다.
func (o *observer) lockWaiters() (map[int]bool, error) {
	rows, err := o.q(fmt.Sprintf(`SELECT pid FROM pg_stat_activity WHERE pid IN (%s) AND wait_event_type = 'Lock'`, joinInts(o.pids)))
	if err != nil {
		return nil, err
	}
	out := map[int]bool{}
	for _, r := range rows {
		pid, _ := strconv.Atoi(r[0])
		out[pid] = true
	}
	return out, nil
}

// capture 는 지금의 페이지, 세션 상태, 락, 진행 중인 트랜잭션, 커밋 로그 상태를 읽습니다.
func (o *observer) capture(after int, tables []string) (capture, error) {
	c := capture{AfterStep: after, Ns: time.Now().UnixNano(), Pages: map[string][]pageItem{}, Xacts: map[string]string{}}
	xids := map[string]bool{}
	for _, t := range tables {
		rows, err := o.q(fmt.Sprintf(`SELECT b, h.lp, h.lp_flags, coalesce(h.t_xmin::text, ''), coalesce(h.t_xmax::text, ''), coalesce(h.t_ctid::text, ''),
 coalesce(h.t_infomask::text, ''), coalesce(h.t_infomask2::text, ''), coalesce(array_to_string(f.raw_flags, ','), ''),
 coalesce(array_to_string(h.t_attrs::text[], ',', 'NULL'), '')
FROM generate_series(0, (pg_relation_size(%[1]s) / 8192)::int - 1) b
CROSS JOIN LATERAL heap_page_item_attrs(get_raw_page(%[1]s, b), %[1]s::regclass) h
LEFT JOIN LATERAL heap_tuple_infomask_flags(h.t_infomask, h.t_infomask2) f ON true ORDER BY b, h.lp`, quote(t)))
		if err != nil {
			return c, err
		}
		for _, r := range rows {
			it := pageItem{Xmin: r[3], Xmax: r[4], Ctid: r[5], Infomask: r[6], Infomask2: r[7]}
			it.Block, _ = strconv.Atoi(r[0])
			it.Lp, _ = strconv.Atoi(r[1])
			it.Flags, _ = strconv.Atoi(r[2])
			if r[8] != "" {
				it.RawFlags = strings.Split(r[8], ",")
			}
			if r[9] != "" {
				it.Attrs = strings.Split(r[9], ",")
			}
			c.Pages[t] = append(c.Pages[t], it)
			xids[it.Xmin], xids[it.Xmax] = true, true
		}
	}
	rows, err := o.q(fmt.Sprintf(`SELECT pid, coalesce(backend_xid::text, ''), coalesce(backend_xmin::text, ''), coalesce(state, ''),
 coalesce(wait_event_type, ''), coalesce(wait_event, '') FROM pg_stat_activity WHERE pid IN (%s) ORDER BY pid`, joinInts(o.pids)))
	if err != nil {
		return c, err
	}
	for _, r := range rows {
		a := activity{Xid: r[1], Xmin: r[2], State: r[3], WaitType: r[4], WaitEvent: r[5]}
		a.Pid, _ = strconv.Atoi(r[0])
		c.Activity = append(c.Activity, a)
		xids[a.Xid] = true
	}
	rows, err = o.q(fmt.Sprintf(`SELECT pid, locktype, coalesce(relation::regclass::text, ''), coalesce(page::text, ''), coalesce(tuple::text, ''),
 coalesce(transactionid::text, ''), coalesce(virtualxid, ''), mode, granted FROM pg_locks WHERE pid IN (%s)
 ORDER BY pid, locktype, relation::text, transactionid::text`, joinInts(o.pids)))
	if err != nil {
		return c, err
	}
	for _, r := range rows {
		l := lockRow{Type: r[1], Relation: r[2], Page: r[3], Tuple: r[4], Xid: r[5], Vxid: r[6], Mode: r[7], Granted: r[8] == "t"}
		l.Pid, _ = strconv.Atoi(r[0])
		c.Locks = append(c.Locks, l)
	}
	// 스냅숏과 WAL의 세 위치를 한 질의로 읽어 같은 순간에 가깝게 둡니다.
	rows, err = o.q(`SELECT pg_current_snapshot(), pg_current_wal_insert_lsn(), pg_current_wal_lsn(), pg_current_wal_flush_lsn()`)
	if err != nil {
		return c, err
	}
	if len(rows) > 0 && len(rows[0]) == 4 {
		c.Running = rows[0][0]
		c.Wal = &walPos{Insert: rows[0][1], Write: rows[0][2], Flush: rows[0][3]}
	}
	// 커밋 로그는 일반 번호(3 이상)만 묻습니다. 0은 "없음", 1과 2는 부트스트랩과 동결 번호입니다.
	// pg_xact_status는 다음에 줄 번호 이상이면 오류를 내지만(xid8funcs.c 116~120행), 여기서 묻는 번호는 이 관찰이 페이지와 세션
	// 상태에서 막 읽은 것이라 늘 그보다 작습니다. kill 뒤 다음 번호가 되돌아가도 관찰은 페이지를 새로 읽고, 잃은 번호는 디스크의
	// 페이지에 없습니다(WAL을 먼저 내려야 페이지를 씀). 스냅숏 xmax로 거르면 안 됩니다. xmax는 마지막으로 끝난 번호 + 1이라
	// (procarray.c 2239·2248~2249행) 진행 중인 번호도 xmax 이상일 수 있습니다(2026-10-14 3-rollback의 첫 두 벌이 진행 중인 753을 미래 번호로 잘못 남겼음).
	var list []string
	for x := range xids {
		if n, err := strconv.Atoi(x); err == nil && n >= 3 {
			list = append(list, x)
		}
	}
	sort.Strings(list)
	if len(list) > 0 {
		rows, err = o.q(fmt.Sprintf(`SELECT x, coalesce(pg_xact_status(x::text::xid8), '') FROM unnest(ARRAY[%s]::bigint[]) x`, strings.Join(list, ",")))
		if err != nil {
			return c, err
		}
		for _, r := range rows {
			c.Xacts[r[0]] = r[1]
		}
	}
	return c, nil
}

// quote 는 SQL 문자열 상수를 만듭니다. 표 이름은 시나리오 파일에서 오므로 작은따옴표를 겹쳐 씁니다.
func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// joinInts 는 SQL의 IN 목록을 만듭니다. 비면 "NULL"을 돌려줍니다. "IN ()"은 문법 오류이고 "IN (NULL)"은 어느 행도 맞지 않습니다
// (kill 뒤 관찰은 세션이 접속하기 전에 찍음).
func joinInts(v []int) string {
	if len(v) == 0 {
		return "NULL"
	}
	var s []string
	for _, n := range v {
		s = append(s, strconv.Itoa(n))
	}
	return strings.Join(s, ",")
}
