package main

// 이 파일은 시나리오 기록에서 세션이 아닌 서버 프로세스(postmaster, startup, WAL writer, checkpointer 등)의 일을 꺼냅니다.
// 탐색용 코드입니다(2026-10-14, 트랜잭션 관측 실험 3: 취소와 커밋, 그리고 서버가 죽었을 때).
//
// 재료는 넷입니다. 프로세스마다의 함수 기록(어느 종류의 프로세스인지, 언제 살아 있었는지), strace(신호, pg_wal 쓰기와 동기화,
// 프로세스의 끝), uprobe(복구가 다시 실행한 레코드, 커밋 로그, WAL writer의 쓰기), server.log(재시작의 이정표).

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// procRoles 는 프로세스 종류를 정하는 주 함수입니다. postmaster는 fork한 자식에서 종류에 맞는 주 함수를 부릅니다
// (launch_backend.c 180~207행의 child_process_kinds). 백그라운드 워커는 BackgroundWorkerMain이 워커의 함수를 다시 부르므로
// 그 뒤의 더 구체적인 함수(ApplyLauncherMain 등)로 정합니다.
var procRoles = map[string]string{
	"PostmasterMain":       "postmaster",
	"StartupProcessMain":   "startup",
	"CheckpointerMain":     "checkpointer",
	"BackgroundWriterMain": "bgwriter",
	"WalWriterMain":        "walwriter",
	"IoWorkerMain":         "io worker",
	"AutoVacLauncherMain":  "autovacuum launcher",
	"AutoVacWorkerMain":    "autovacuum worker",
	"ApplyLauncherMain":    "logical replication launcher",
	"BackendMain":          "backend",
	"BackgroundWorkerMain": "background worker",
	"PgArchiverMain":       "archiver",
	"WalSummarizerMain":    "walsummarizer",
	"SysLoggerMain":        "syslogger",
}

// roleFromMains 는 함수 기록에서 기록 순서대로 만난 주 함수 이름들로 프로세스 종류를 정합니다.
//
// Returns: 처음 만난 주 함수의 종류입니다. BackgroundWorkerMain만 만났으면 "background worker"이고 done이 거짓입니다
// (그 뒤에 워커의 함수가 올 수 있음). 아무것도 없으면 빈 문자열입니다.
func roleFromMains(mains []string) (role string, done bool) {
	for _, m := range mains {
		if r := procRoles[m]; r != "" && r != "background worker" {
			return r, true
		}
	}
	if len(mains) > 0 {
		return "background worker", false
	}
	return "", false
}

// procRole 은 함수 기록의 입구를 앞에서부터 보며 주 함수를 만나는 대로 roleFromMains에 넘깁니다.
func procRole(t *traceFile, syms *symbols) string {
	var mains []string
	for _, r := range t.Records {
		if r.Kind != kindEnter {
			continue
		}
		name := syms.name(r.Fn - t.Slide)
		if _, ok := procRoles[name]; !ok {
			continue
		}
		mains = append(mains, name)
		if role, done := roleFromMains(mains); done {
			return role
		}
	}
	role, _ := roleFromMains(mains)
	return role
}

// lifeEvent 는 strace가 프로세스의 끝에 남기는 줄 하나입니다. How는 "killed by SIGKILL", "exited with 2" 같은 끝난 꼴입니다.
type lifeEvent struct {
	Pid int
	Ns  uint64
	How string
}

// lifeLine 은 "pid sec.usec +++ 끝난 꼴 +++" 줄입니다. readStrace는 괄호가 없는 이 줄을 건너뛰므로 따로 읽습니다.
var lifeLine = regexp.MustCompile(`^(\d+)\s+(\d+)\.(\d+)\s+\+\+\+ (.+) \+\+\+$`)

// parseLifeLine 은 strace 한 줄이 프로세스의 끝이면 그것을 돌려줍니다.
func parseLifeLine(l string) (lifeEvent, bool) {
	m := lifeLine.FindStringSubmatch(l)
	if m == nil {
		return lifeEvent{}, false
	}
	pid, _ := strconv.Atoi(m[1])
	sec, _ := strconv.ParseUint(m[2], 10, 64)
	usec, _ := strconv.ParseUint(m[3], 10, 64)
	return lifeEvent{Pid: pid, Ns: sec*1e9 + usec*1e3, How: m[4]}, true
}

// readLifecycle 은 strace 출력에서 프로세스의 끝 줄을 모두 읽습니다.
//
// Errors: 파일을 열 수 없으면 오류를 돌려줍니다.
func readLifecycle(file string) ([]lifeEvent, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []lifeEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if ev, ok := parseLifeLine(sc.Text()); ok {
			out = append(out, ev)
		}
	}
	return out, sc.Err()
}

// logLine 은 server.log의 한 항목입니다. 시각은 log_line_prefix의 %m(밀리초까지, 이 실험대는 UTC)이라 1 ms 단위로 내림된 값입니다.
// Msg에는 탭으로 시작하는 이어진 줄(교착의 DETAIL 등)을 줄바꿈으로 붙입니다.
type logLine struct {
	Ns    uint64
	Pid   int
	Level string
	Msg   string
}

// logPrefix 는 PostgreSQL의 기본 log_line_prefix('%m [%p] ')로 시작하는 줄입니다.
var logPrefix = regexp.MustCompile(`^(\d{4}-\d\d-\d\d \d\d:\d\d:\d\d\.\d{3}) (\S+) \[(\d+)\] ([A-Z]+):  (.*)$`)

// parseLogLine 은 로그 한 줄의 시각, pid, 수준, 글을 풉니다. 접두어가 없는 줄이면 ok가 거짓입니다.
func parseLogLine(l string) (logLine, bool) {
	m := logPrefix.FindStringSubmatch(l)
	if m == nil {
		return logLine{}, false
	}
	t, err := time.Parse("2006-01-02 15:04:05.000 MST", m[1]+" "+m[2])
	if err != nil {
		return logLine{}, false
	}
	pid, _ := strconv.Atoi(m[3])
	return logLine{Ns: uint64(t.UnixNano()), Pid: pid, Level: m[4], Msg: m[5]}, true
}

// readServerLog 는 server.log를 항목 목록으로 읽습니다. path가 비면 빈 목록입니다.
//
// Errors: 파일을 열 수 없으면 오류를 돌려줍니다.
func readServerLog(path string) ([]logLine, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []logLine
	for _, l := range strings.Split(string(b), "\n") {
		if e, ok := parseLogLine(l); ok {
			out = append(out, e)
		} else if strings.HasPrefix(l, "\t") && len(out) > 0 {
			out[len(out)-1].Msg += "\n" + strings.TrimPrefix(l, "\t")
		}
	}
	return out, nil
}

// lsnText 는 uprobe가 십진수로 남긴 WAL 위치를 PostgreSQL이 쓰는 "상위32비트/하위32비트" 16진 꼴로 바꿉니다(예: 0/179D330).
// 읽지 못하면 받은 글을 그대로 돌려줍니다.
func lsnText(dec string) string {
	v, err := strconv.ParseUint(dec, 10, 64)
	if err != nil {
		return dec
	}
	return fmt.Sprintf("%X/%X", v>>32, v&0xffffffff)
}

// rmgrNames 는 자원 관리자 번호의 이름입니다(rmgrlist.h의 순서).
var rmgrNames = []string{"XLOG", "Transaction", "Storage", "CLOG", "Database", "Tablespace", "MultiXact", "RelMap", "Standby",
	"Heap2", "Heap", "Btree", "Hash", "Gin", "Gist", "Sequence", "SPGist", "BRIN", "CommitTs", "ReplicationOrigin", "Generic", "LogicalMessage"}

// recordKinds 는 자원 관리자마다 info의 위 4비트가 가리키는 레코드 종류입니다. 값은 각 헤더의 #define입니다(pg_control.h,
// xact.h, heapam_xlog.h, nbtxlog.h, standbydefs.h, storage_xlog.h, clog.h). Transaction과 Heap은 0x70 안의 값이 종류이고
// 0x80은 덧붙은 표지입니다(XLOG_XACT_HAS_INFO, XLOG_HEAP_INIT_PAGE).
var recordKinds = map[string]map[int64]string{
	"XLOG": {0x00: "CHECKPOINT_SHUTDOWN", 0x10: "CHECKPOINT_ONLINE", 0x20: "NOOP", 0x30: "NEXTOID", 0x40: "SWITCH", 0x50: "BACKUP_END",
		0x60: "PARAMETER_CHANGE", 0x70: "RESTORE_POINT", 0x80: "FPW_CHANGE", 0x90: "END_OF_RECOVERY", 0xA0: "FPI_FOR_HINT", 0xB0: "FPI",
		0xD0: "OVERWRITE_CONTRECORD", 0xE0: "CHECKPOINT_REDO"},
	"Transaction": {0x00: "COMMIT", 0x10: "PREPARE", 0x20: "ABORT", 0x30: "COMMIT_PREPARED", 0x40: "ABORT_PREPARED", 0x50: "ASSIGNMENT", 0x60: "INVALIDATIONS"},
	"Heap":        {0x00: "INSERT", 0x10: "DELETE", 0x20: "UPDATE", 0x30: "TRUNCATE", 0x40: "HOT_UPDATE", 0x50: "CONFIRM", 0x60: "LOCK", 0x70: "INPLACE"},
	"Heap2": {0x00: "REWRITE", 0x10: "PRUNE_ON_ACCESS", 0x20: "PRUNE_VACUUM_SCAN", 0x30: "PRUNE_VACUUM_CLEANUP", 0x40: "VISIBLE",
		0x50: "MULTI_INSERT", 0x60: "LOCK_UPDATED", 0x70: "NEW_CID"},
	"Btree": {0x00: "INSERT_LEAF", 0x10: "INSERT_UPPER", 0x20: "INSERT_META", 0x30: "SPLIT_L", 0x40: "SPLIT_R", 0x50: "INSERT_POST",
		0x60: "DEDUP", 0x70: "DELETE", 0x80: "UNLINK_PAGE", 0x90: "UNLINK_PAGE_META", 0xA0: "NEWROOT", 0xB0: "MARK_PAGE_HALFDEAD",
		0xC0: "VACUUM", 0xD0: "REUSE_PAGE", 0xE0: "META_CLEANUP"},
	"Standby": {0x00: "LOCK", 0x10: "RUNNING_XACTS", 0x20: "INVALIDATIONS"},
	"Storage": {0x10: "CREATE", 0x20: "TRUNCATE"},
	"CLOG":    {0x00: "ZEROPAGE", 0x10: "TRUNCATE"},
}

// recordKind 는 레코드의 자원 관리자 번호와 info로 이름을 돌려줍니다. 모르는 종류는 info를 16진으로 적습니다.
func recordKind(rmid, info int64) (rmgr, kind string) {
	rmgr = fmt.Sprintf("rmgr%d", rmid)
	if rmid >= 0 && int(rmid) < len(rmgrNames) {
		rmgr = rmgrNames[rmid]
	}
	op := info & 0xF0
	if rmgr == "Transaction" || rmgr == "Heap" {
		op = info & 0x70
	}
	if k, ok := recordKinds[rmgr][op]; ok {
		return rmgr, k
	}
	return rmgr, fmt.Sprintf("info 0x%02X", info)
}

// serverProc 는 시나리오 구간에 살아 있던 프로세스 하나입니다. T0·T1은 처음과 마지막으로 본 시각이고, Exit는 strace가 남긴
// 끝난 꼴입니다(없으면 구간이 끝날 때까지 살아 있었음).
type serverProc struct {
	Pid     int     `json:"pid"`
	Role    string  `json:"role"`
	Session string  `json:"session,omitempty"`
	T0      float64 `json:"t0"`
	T1      float64 `json:"t1"`
	Exit    string  `json:"exit,omitempty"`
}

// serverEvent 는 서버 프로세스의 일 하나입니다. Kind는 signal(kill 시스템 콜), exit(프로세스의 끝), wal_io(pg_wal 쓰기와 동기화),
// redo(복구가 다시 실행한 레코드), commit_clog·abort_clog(커밋 로그), bg_flush(WAL writer의 한 주기), wal_flush(XLogFlush),
// async_lsn, hint_fpi, log(server.log 항목)입니다.
//
// Stack은 그 순간 그 프로세스의 호출 스택이고(함수 기록이 있는 프로세스만), ns는 스택을 물을 절대 시각입니다.
type serverEvent struct {
	T     float64           `json:"t"`
	T1    float64           `json:"t1,omitempty"`
	Pid   int               `json:"pid"`
	Kind  string            `json:"kind"`
	F     map[string]string `json:"f,omitempty"`
	Stack []frameRef        `json:"stack,omitempty"`
	ns    uint64
}

// stackKinds 는 스택을 붙일 사건 종류와, 그 시각에 들어온 함수까지 스택에 넣을지(uprobe는 함수 입구라 참, 시스템 콜은 그 함수
// 안에서 부른 것이라 거짓)입니다. WAL writer의 주기(bg_flush)는 매번 같은 길이라, 로그와 끝은 함수 안의 일이 아니라 빼었습니다.
var stackKinds = map[string]bool{"signal": false, "wal_io": false, "redo": true, "commit_clog": true, "abort_clog": true, "wal_flush": true}

// attachServerStacks 는 stackKinds의 사건에 호출 스택을 붙입니다. 표본기는 시각이 늘어나는 순서로만 물을 수 있으므로
// 프로세스마다 사건을 시각 순으로 묻습니다. 스택은 그 프로세스의 가장 안쪽 주 함수(procRoles)부터 남깁니다. 그 위(main,
// PostmasterMain, ServerLoop, fork한 자리)는 postmaster에서 fork한 모든 자식이 같습니다.
//
// Arguments: sampler는 pid의 표본기를 돌려주고, 함수 기록이 없는 프로세스면 nil입니다.
func attachServerStacks(evs []serverEvent, sampler func(pid int) func(uint64, bool) ([]frameRef, string)) {
	byPid := map[int][]int{}
	for i, e := range evs {
		if _, ok := stackKinds[e.Kind]; ok {
			byPid[e.Pid] = append(byPid[e.Pid], i)
		}
	}
	for pid, idx := range byPid {
		sample := sampler(pid)
		if sample == nil {
			continue
		}
		sort.SliceStable(idx, func(a, b int) bool { return evs[idx[a]].ns < evs[idx[b]].ns })
		for _, i := range idx {
			s, _ := sample(evs[i].ns, stackKinds[evs[i].Kind])
			evs[i].Stack = trimAtMain(s)
		}
	}
}

// trimAtMain 은 스택에서 가장 안쪽 주 함수(procRoles의 이름) 앞을 잘라 냅니다. 주 함수가 없으면 그대로 돌려줍니다.
func trimAtMain(s []frameRef) []frameRef {
	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := procRoles[s[i].Fn]; ok {
			return s[i:]
		}
	}
	return s
}

// serverSpanFns 는 서버 프로세스가 머문 구간을 낼 함수입니다. 실험 3에서 kill부터 접속 재개까지 함수 기록에서 오래 걸린 함수와
// 그 길의 갈림목을 골랐습니다(2026-10-14 3-crash-sync-on-a의 기록).
//   - postmaster: 자식이 죽은 것을 처리하고(HandleChildCrash) 다른 자식을 모두 끝낸 뒤(HandleFatalError, TerminateChildren),
//     공유 메모리를 다시 만들고(CreateSharedMemoryAndSemaphores) 자식을 다시 띄웁니다(StartChildProcess).
//   - startup: 데이터 디렉터리를 동기화하고(SyncDataDirectory) WAL을 다시 실행한 뒤(PerformWalRecovery), 복구 끝 체크포인트를
//     요청하고 기다립니다(RequestCheckpoint). 모두 StartupXLOG 안입니다.
//   - checkpointer: 체크포인트(CreateCheckPoint) 안에서 더러운 버퍼를 쓰고(CheckPointBuffers) 파일을 동기화합니다(ProcessSyncRequests).
var serverSpanFns = map[string]bool{
	"HandleChildCrash": true, "HandleFatalError": true, "TerminateChildren": true, "CreateSharedMemoryAndSemaphores": true, "StartChildProcess": true,
	"StartupXLOG": true, "SyncDataDirectory": true, "PerformWalRecovery": true, "RequestCheckpoint": true,
	"CreateCheckPoint": true, "CheckPointBuffers": true, "ProcessSyncRequests": true,
}

// serverSpan 은 서버 프로세스가 고른 함수 하나에 머문 구간입니다. At은 함수가 정의된 줄이고, Open이 참이면 기록이 끝날 때까지
// 출구가 없었습니다.
type serverSpan struct {
	Pid  int     `json:"pid"`
	Fn   string  `json:"fn"`
	T0   float64 `json:"t0"`
	T1   float64 `json:"t1"`
	At   string  `json:"at"`
	Open bool    `json:"open,omitempty"`
}

// spansOf 는 함수 기록 하나에서 want가 고른 함수의 입구와 출구를 같은 함수·같은 프레임끼리 짝지어 구간으로 만듭니다.
//
// Arguments: want는 함수 주소(PIE 차이를 뺀 값)로 이름과 정의된 줄, 고를지 여부를 돌려줍니다. win은 구간이고, 구간과 겹치는
// 호출만 냅니다.
// Returns: 시작 시각 순서의 구간입니다. 출구가 없는 호출은 마지막 기록의 시각에서 끝난 열린 구간입니다.
func spansOf(t *traceFile, want func(fn uint64) (name, at string, ok bool), win [2]uint64, rel func(uint64) float64) []serverSpan {
	type open struct {
		fn, frame, ns uint64
		name, at      string
	}
	var stack []open
	var out []serverSpan
	emit := func(o open, end uint64, isOpen bool) {
		if end >= win[0] && o.ns <= win[1] {
			out = append(out, serverSpan{Pid: t.Pid, Fn: o.name, T0: rel(o.ns), T1: rel(end), At: o.at, Open: isOpen})
		}
	}
	for _, r := range t.Records {
		if r.Kind != kindEnter && r.Kind != kindExit {
			continue
		}
		fn := r.Fn - t.Slide
		name, at, ok := want(fn)
		if !ok {
			continue
		}
		if r.Kind == kindEnter {
			stack = append(stack, open{fn, r.Frame, r.Ns, name, at})
			continue
		}
		for k := len(stack) - 1; k >= 0; k-- {
			if stack[k].fn == fn && stack[k].frame == r.Frame {
				emit(stack[k], r.Ns, false)
				stack = append(stack[:k], stack[k+1:]...)
				break
			}
		}
	}
	if n := len(t.Records); n > 0 {
		for _, o := range stack {
			emit(o, t.Records[n-1].Ns, true)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].T0 < out[j].T0 })
	return out
}

// serverInput 은 serverView의 재료입니다. Evs는 짝짓기를 마친 uprobe 사건 전체, Offset은 짝지은 사건의 ftrace 시각과 함수 기록
// 시각의 차이(나노초, 중앙값), Sessions는 세션 백엔드 pid → 세션 이름, Observers는 관찰용 접속 pid, Win은 run의 시작과 끝입니다.
type serverInput struct {
	Traces    []*traceFile
	Syms      *symbols
	Evs       []uevent
	Offset    float64
	Sys       []sysEvent
	Life      []lifeEvent
	Logs      []logLine
	Sessions  map[int]string
	Observers map[int]bool
	Win       [2]uint64
	Rel       func(uint64) float64
}

// serverView 는 시나리오 구간에 살아 있던 프로세스와, 세션·관찰용 접속이 아닌 프로세스가 그 구간에 한 일을 모읍니다.
// 프로세스의 끝(exit)은 세션과 관찰용 접속의 것도 냅니다. 죽인 세션이 언제 끝났는지가 크래시 장면의 시작이기 때문입니다.
//
// Returns: 프로세스(pid 순), 사건(시각 순), 서버 프로세스가 serverSpanFns의 함수에 머문 구간(시작 순)입니다. 사건은 구간 안의
// 것만 담고, 로그는 시각이 1 ms로 내림되어 있어 구간을 양쪽으로 1 ms 넓혀 담습니다.
func serverView(in serverInput) ([]serverProc, []serverEvent, []serverSpan) {
	inWin := func(ns uint64) bool { return ns >= in.Win[0] && ns <= in.Win[1] }
	exit := map[int]lifeEvent{}
	for _, l := range in.Life {
		exit[l.Pid] = l
	}
	// 처음과 마지막으로 본 시각: 함수 기록이 있으면 그것으로, 없으면(관찰용 접속은 기록을 지움) strace로 봅니다.
	first, last := map[int]uint64{}, map[int]uint64{}
	seen := func(pid int, ns uint64) {
		if v, ok := first[pid]; !ok || ns < v {
			first[pid] = ns
		}
		if ns > last[pid] {
			last[pid] = ns
		}
	}
	role := map[int]string{}
	for _, t := range in.Traces {
		if len(t.Records) < 2 {
			continue
		}
		seen(t.Pid, t.Records[1].Ns)
		seen(t.Pid, t.Records[len(t.Records)-1].Ns)
		role[t.Pid] = procRole(t, in.Syms)
	}
	for _, e := range in.Sys {
		if in.Observers[e.Pid] {
			seen(e.Pid, e.Ns)
		}
	}
	for p := range in.Observers {
		role[p] = "observer"
	}
	var procs []serverProc
	for pid, r := range role {
		end := last[pid]
		ex, ok := exit[pid]
		if ok {
			end = ex.Ns
		}
		if end < in.Win[0] || first[pid] > in.Win[1] {
			continue
		}
		p := serverProc{Pid: pid, Role: r, Session: in.Sessions[pid], T0: in.Rel(first[pid]), T1: in.Rel(end)}
		if ok {
			p.Exit = ex.How
		}
		procs = append(procs, p)
	}
	sort.Slice(procs, func(i, j int) bool { return procs[i].Pid < procs[j].Pid })

	server := func(pid int) bool { _, s := in.Sessions[pid]; return !s && !in.Observers[pid] }
	var evs []serverEvent
	for _, l := range in.Life {
		if inWin(l.Ns) {
			evs = append(evs, serverEvent{T: in.Rel(l.Ns), Pid: l.Pid, Kind: "exit", F: map[string]string{"how": l.How}, ns: l.Ns})
		}
	}
	for _, e := range in.Sys {
		if !server(e.Pid) || !inWin(e.Ns) {
			continue
		}
		switch {
		case e.Name == "kill":
			to, sig, _ := strings.Cut(e.Args, ", ")
			// postmaster가 신호 처리기에서 자기 래치를 깨우는 kill(자기, SIGURG)은 매번 같아서 뺍니다.
			if to == strconv.Itoa(e.Pid) && sig == "SIGURG" {
				continue
			}
			f := map[string]string{"to": strings.TrimPrefix(to, "-"), "sig": sig, "ret": fmt.Sprint(e.Ret)}
			if strings.HasPrefix(to, "-") {
				f["group"] = "1"
			}
			evs = append(evs, serverEvent{T: in.Rel(e.Ns), Pid: e.Pid, Kind: "signal", F: f, ns: e.Ns - uint64(e.DurNs)})
		case strings.Contains(e.FdPath, "/pg_wal/") && (e.Name == "pwrite64" || e.Name == "pwritev" || e.Name == "fdatasync" || e.Name == "fsync"):
			evs = append(evs, serverEvent{T: in.Rel(e.Ns - uint64(e.DurNs)), T1: in.Rel(e.Ns), Pid: e.Pid, Kind: "wal_io", F: walIOFields(e), ns: e.Ns - uint64(e.DurNs)})
		}
	}
	// 배경 프로세스는 uprobe를 걸기 전부터 떠 있고 끈 뒤에도 돌아서, 함수 기록에만 있는 호출이 앞뒤에 생겨 짝짓기 수가 어긋납니다
	// (2026-10-14 3-crash-sync-on-a: 옛 WAL writer의 XLogBackgroundFlush 반환이 uprobe 10번, 함수 기록 19번). 그런 사건은
	// ftrace 시각에서 Offset을 빼 씁니다. 짝지은 사건으로 잰 두 시계의 차이는 흔들림이 10 µs 안팎입니다(clock.maxDeviationUs).
	slide := map[int]uint64{}
	for _, t := range in.Traces {
		slide[t.Pid] = t.Slide
	}
	byPid := map[int][]*uevent{}
	for i := range in.Evs {
		e := &in.Evs[i]
		if _, ok := slide[e.Pid]; ok && server(e.Pid) {
			byPid[e.Pid] = append(byPid[e.Pid], e)
		}
	}
	nsOf := func(e *uevent) uint64 {
		if e.trace != nil {
			return e.ns
		}
		return uint64(e.Ts*1e9 - in.Offset)
	}
	for pid, l := range byPid {
		sort.SliceStable(l, func(i, j int) bool { return nsOf(l[i]) < nsOf(l[j]) })
		var fpiOpen *uevent
		for _, e := range l {
			ns := nsOf(e)
			if !inWin(ns) {
				continue
			}
			ev := serverEvent{T: in.Rel(ns), Pid: pid, ns: ns}
			switch fn := in.Syms.name(e.Func - slide[pid]); {
			case fn == "ApplyWalRecord" && !e.Return:
				rmgr, kind := recordKind(atoi(e.F["rmid"]), atoi(e.F["info"]))
				ev.Kind, ev.F = "redo", map[string]string{"lsn": lsnText(e.F["lsn"]), "end": lsnText(e.F["end"]), "rmgr": rmgr, "kind": kind, "xid": e.F["xid"], "len": e.F["len"]}
			case fn == "TransactionIdCommitTree" || fn == "TransactionIdAbortTree":
				ev.Kind, ev.F = map[bool]string{true: "commit_clog", false: "abort_clog"}[fn == "TransactionIdCommitTree"], map[string]string{"xid": e.F["xid"]}
			case fn == "TransactionIdAsyncCommitTree":
				ev.Kind, ev.F = "commit_clog", map[string]string{"xid": e.F["xid"], "lsn": lsnText(e.F["lsn"]), "async": "1"}
			case fn == "XLogBackgroundFlush" && e.Return:
				ev.Kind, ev.F = "bg_flush", map[string]string{"ret": e.F["ret"]}
			case fn == "XLogFlush" && !e.Return:
				ev.Kind, ev.F = "wal_flush", map[string]string{"lsn": lsnText(e.F["lsn"])}
			case fn == "XLogSetAsyncXactLSN" && !e.Return:
				ev.Kind, ev.F = "async_lsn", map[string]string{"lsn": lsnText(e.F["lsn"])}
			case fn == "XLogSaveBufferForHint" && !e.Return:
				fpiOpen = e
				continue
			case fn == "XLogSaveBufferForHint" && e.Return && fpiOpen != nil:
				ev.T, ev.T1, ev.ns = in.Rel(nsOf(fpiOpen)), in.Rel(ns), nsOf(fpiOpen)
				ev.Kind, ev.F = "hint_fpi", map[string]string{"lsn": lsnText(e.F["lsn"])}
				fpiOpen = nil
			default:
				continue
			}
			if e.trace == nil {
				ev.F["clock"] = "ftrace"
			}
			evs = append(evs, ev)
		}
	}
	for _, l := range in.Logs {
		if l.Ns+1e6 >= in.Win[0] && l.Ns <= in.Win[1]+1e6 {
			evs = append(evs, serverEvent{T: in.Rel(l.Ns), Pid: l.Pid, Kind: "log", F: map[string]string{"level": l.Level, "msg": l.Msg}, ns: l.Ns})
		}
	}
	// 함수 기록이 있는 서버 프로세스의 사건에 스택을 붙이고, 고른 함수에 머문 구간을 냅니다. ftrace 시각으로 쓴 사건(clock
	// "ftrace")의 스택은 두 시계 차이의 흔들림만큼(실험 3에서 최대 1 ms 안팎) 어긋날 수 있습니다.
	traceOf := map[int]*traceFile{}
	for _, t := range in.Traces {
		if server(t.Pid) {
			traceOf[t.Pid] = t
		}
	}
	attachServerStacks(evs, func(pid int) func(uint64, bool) ([]frameRef, string) {
		if t := traceOf[pid]; t != nil {
			return stackSampler(t, in.Syms)
		}
		return nil
	})
	want := func(fn uint64) (string, string, bool) {
		name := in.Syms.name(fn)
		if !serverSpanFns[name] {
			return "", "", false
		}
		at := ""
		if fi := in.Syms.byLow[fn]; fi != nil && fi.Line > 0 {
			at = fmt.Sprintf("%s:%d", fi.File, fi.Line)
		}
		return name, at, true
	}
	spans := []serverSpan{}
	for _, t := range in.Traces {
		if server(t.Pid) {
			spans = append(spans, spansOf(t, want, in.Win, in.Rel)...)
		}
	}
	sort.SliceStable(spans, func(i, j int) bool { return spans[i].T0 < spans[j].T0 })
	sort.SliceStable(evs, func(i, j int) bool { return evs[i].T < evs[j].T })
	return procs, evs, spans
}

// walSegment 는 WAL 파일 하나의 크기입니다. initdb 기본값이고, 이 실험대의 pg_wal 파일 이름(000000010000000000000001 등)과
// 파일 안 위치가 이 크기를 넘지 않는 것으로 기록에서 확인합니다.
const walSegment = 16 << 20

// walIOFields 는 pg_wal에 쓴 시스템 콜 하나를 화면이 쓸 값으로 바꿉니다. 쓰기는 파일 안 위치와 길이에 더해, 그 범위를
// WAL 위치(LSN)로 옮긴 값(lsn0~lsn1)을 둡니다. 파일 이름의 끝 16자리가 몇 번째 파일인지이므로 LSN = 번호 × 16 MB + 위치입니다.
func walIOFields(e sysEvent) map[string]string {
	f := map[string]string{"call": e.Name, "file": shortPath(e.FdPath), "durUs": fmt.Sprint(e.DurNs / 1e3)}
	if e.Name != "pwrite64" && e.Name != "pwritev" {
		return f
	}
	f["off"], f["len"] = fmt.Sprint(e.Offset), fmt.Sprint(e.Ret)
	if lo, hi, ok := walRange(e); ok {
		f["lsn0"], f["lsn1"] = lsnText(fmt.Sprint(lo)), lsnText(fmt.Sprint(hi))
	}
	return f
}

// walRange 는 pg_wal 파일에 쓴 시스템 콜 하나가 덮은 WAL 위치 범위 [lo, hi)를 돌려줍니다. 파일 이름이 WAL 파일 꼴(16진 24자리)이
// 아니면 ok가 거짓입니다.
func walRange(e sysEvent) (lo, hi uint64, ok bool) {
	name := e.FdPath[strings.LastIndex(e.FdPath, "/")+1:]
	if len(name) != 24 {
		return 0, 0, false
	}
	h, err1 := strconv.ParseUint(name[8:16], 16, 64)
	l, err2 := strconv.ParseUint(name[16:24], 16, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	// 이름의 가운데 8자리는 LSN 상위 32비트, 끝 8자리는 그 안의 파일 번호입니다(16 MB 파일이면 0~255).
	base := (h<<32/walSegment + l) * walSegment
	return base + uint64(e.Offset), base + uint64(e.Offset) + uint64(max(e.Ret, 0)), true
}

// walExtentBefore 는 시각 ns 전에 끝난 pg_wal 쓰기로 WAL이 어디까지 커널에 넘어갔는지와, 그중 어디까지 동기화를 마쳤는지를 구합니다.
// WAL은 앞에서부터 차례로 쓰므로 쓴 범위의 끝 가운데 가장 먼 곳이 쓴 끝입니다. 동기화는 파일 단위라(fdatasync는 그 파일의
// 더러운 데이터를 모두 내림) 쓰기가 끝난 뒤 같은 파일에서 시작해 ns 전에 끝난 fdatasync·fsync가 있으면 프로세스와 관계없이
// 그 쓰기를 내린 것으로 봅니다.
//
// Returns: 쓴 끝과 동기화한 끝(LSN 값, 없으면 0), 쓴 끝을 만든 쓰기가 끝난 시각입니다.
func walExtentBefore(sys []sysEvent, ns uint64) (written, synced, writtenNs uint64) {
	isWal := func(e sysEvent) bool { return strings.Contains(e.FdPath, "/pg_wal/") }
	for _, w := range sys {
		if w.Ns >= ns || !isWal(w) || (w.Name != "pwrite64" && w.Name != "pwritev") {
			continue
		}
		_, hi, ok := walRange(w)
		if !ok {
			continue
		}
		if hi > written {
			written, writtenNs = hi, w.Ns
		}
		for _, s := range sys {
			if s.FdPath == w.FdPath && (s.Name == "fdatasync" || s.Name == "fsync") && s.Ns-uint64(s.DurNs) >= w.Ns && s.Ns < ns {
				synced = max(synced, hi)
				break
			}
		}
	}
	return written, synced, writtenNs
}
