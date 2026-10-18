package main

import (
	"fmt"
	"reflect"
	"testing"
)

// TestRecordKindNamesReplayedRecords 는 복구가 다시 실행한 레코드의 자원 관리자 번호와 info를 이름으로 바꾸는지 봅니다.
// 값은 2026-10-14 3-crash-sync-on-a에서 startup 프로세스가 다시 실행한 레코드입니다.
func TestRecordKindNamesReplayedRecords(t *testing.T) {
	for _, c := range []struct {
		rmid, info int64
		rmgr, kind string
	}{
		{0, 0xA0, "XLOG", "FPI_FOR_HINT"},
		{10, 0x00, "Heap", "INSERT"},
		{10, 0x80, "Heap", "INSERT"}, // 0x80은 XLOG_HEAP_INIT_PAGE 표지라 종류가 아님
		{11, 0x00, "Btree", "INSERT_LEAF"},
		{1, 0x00, "Transaction", "COMMIT"},
		{1, 0xA0, "Transaction", "ABORT"}, // 0x80은 XLOG_XACT_HAS_INFO 표지
		{9, 0x10, "Heap2", "PRUNE_ON_ACCESS"},
		{30, 0x00, "rmgr30", "info 0x00"},
	} {
		// When
		rmgr, kind := recordKind(c.rmid, c.info)

		// Then
		if rmgr != c.rmgr || kind != c.kind {
			t.Errorf("recordKind(%d, %#x) = %s %s; want %s %s", c.rmid, c.info, rmgr, kind, c.rmgr, c.kind)
		}
	}
}

// TestLsnTextUsesPostgresForm 은 uprobe가 십진수로 남긴 WAL 위치를 server.log와 pg_current_wal_lsn()이 쓰는 꼴로 바꾸는지 봅니다.
func TestLsnTextUsesPostgresForm(t *testing.T) {
	for in, want := range map[string]string{"24761136": "0/179D330", "4294967296": "1/0", "0": "0/0", "x": "x"} {
		if got := lsnText(in); got != want {
			t.Errorf("lsnText(%q) = %q; want %q", in, got, want)
		}
	}
}

// TestWalIOFieldsMapsWriteToLsnRange 는 pg_wal 파일에 쓴 범위를 LSN 범위로 옮기는지 봅니다. 값은 3-crash-sync-off-wait-a에서
// 옛 WAL writer가 비동기 커밋 뒤 첫 주기에 쓴 pwrite64이고, server.log의 redo 위치와 관찰의 WAL 위치로 맞는지 확인한 범위입니다.
func TestWalIOFieldsMapsWriteToLsnRange(t *testing.T) {
	// Given
	e := sysEvent{Name: "pwrite64", FdPath: "/tmp/pgi/data/pg_wal/000000010000000000000001", Offset: 7872512, Ret: 106496}

	// When
	f := walIOFields(e)

	// Then
	if f["lsn0"] != "0/1782000" || f["lsn1"] != "0/179C000" || f["file"] != "pg_wal/000000010000000000000001" {
		t.Fatalf("walIOFields = %v; want lsn0 0/1782000, lsn1 0/179C000", f)
	}
}

// TestParseLogLineReadsPrefix 는 server.log 줄의 시각(UTC, 밀리초), pid, 수준, 글을 읽고 접두어 없는 줄은 받지 않는지 봅니다.
func TestParseLogLineReadsPrefix(t *testing.T) {
	// Given
	l := "2026-10-14 14:14:39.215 UTC [1622] LOG:  all server processes terminated; reinitializing"

	// When
	got, ok := parseLogLine(l)
	_, okCont := parseLogLine("\tProcess 187 waits for ShareLock on transaction 753; blocked by process 185.")

	// Then
	want := logLine{Ns: 1791987279215000000, Pid: 1622, Level: "LOG", Msg: "all server processes terminated; reinitializing"}
	if !ok || got != want || okCont {
		t.Fatalf("parseLogLine = %+v, %v (이어진 줄 %v); want %+v", got, ok, okCont, want)
	}
}

// TestParseLifeLineReadsExits 는 strace의 프로세스 끝 줄 두 꼴을 읽고 시스템 콜 줄은 받지 않는지 봅니다.
func TestParseLifeLineReadsExits(t *testing.T) {
	for _, c := range []struct {
		line string
		want lifeEvent
		ok   bool
	}{
		{"1646  1791987279.215002 +++ killed by SIGKILL +++", lifeEvent{1646, 1791987279215002000, "killed by SIGKILL"}, true},
		{"1630  1791987279.216350 +++ exited with 2 +++", lifeEvent{1630, 1791987279216350000, "exited with 2"}, true},
		{"1622  1791987279.216119 kill(1632, SIGQUIT) = 0 <0.000002>", lifeEvent{}, false},
	} {
		// When
		got, ok := parseLifeLine(c.line)

		// Then
		if got != c.want || ok != c.ok {
			t.Errorf("parseLifeLine(%q) = %+v, %v; want %+v, %v", c.line, got, ok, c.want, c.ok)
		}
	}
}

// TestRoleFromMainsPrefersWorkerFunction 은 백그라운드 워커가 BackgroundWorkerMain 뒤에 부른 워커 함수로 종류를 정하는지 봅니다.
func TestRoleFromMainsPrefersWorkerFunction(t *testing.T) {
	for _, c := range []struct {
		mains []string
		role  string
		done  bool
	}{
		{[]string{"StartupProcessMain"}, "startup", true},
		{[]string{"BackgroundWorkerMain"}, "background worker", false},
		{[]string{"BackgroundWorkerMain", "ApplyLauncherMain"}, "logical replication launcher", true},
		{nil, "", false},
	} {
		// When
		role, done := roleFromMains(c.mains)

		// Then
		if role != c.role || done != c.done {
			t.Errorf("roleFromMains(%v) = %q, %v; want %q, %v", c.mains, role, done, c.role, c.done)
		}
	}
}

// TestEpochStepsSplitsAtKill 은 kill 뒤 첫 단계에서 구간을 나누는지 봅니다(3-crash-sync-off: 단계 1·2 뒤 kill, 단계 3~5).
func TestEpochStepsSplitsAtKill(t *testing.T) {
	// Given
	var steps []txStep
	for n := 1; n <= 5; n++ {
		steps = append(steps, txStep{N: n})
	}

	// When
	groups := epochSteps(steps, []int{3})

	// Then
	var got [][]int
	for _, g := range groups {
		var ns []int
		for _, s := range g {
			ns = append(ns, s.N)
		}
		got = append(got, ns)
	}
	if !reflect.DeepEqual(got, [][]int{{1, 2}, {3, 4, 5}}) {
		t.Fatalf("epochSteps = %v; want [[1 2] [3 4 5]]", got)
	}
}

// TestHintChangesUsesPreKillCaptureAfterLastStep 는 kill 앞 마지막 단계의 힌트 비트를 그 단계 뒤 관찰(재시작 앞)로 견주고, kill 뒤
// 첫 단계는 재시작 뒤 관찰부터 견주는지 봅니다. 같은 단계 번호 뒤의 관찰이 둘이라, 번호로 하나만 고르면 한쪽이 틀립니다.
func TestHintChangesUsesPreKillCaptureAfterLastStep(t *testing.T) {
	// Given: S1이 단계 1에서 (0,1)에 xmin 커밋 비트를 쓰고, 다시 뜬 서버의 페이지에는 그 비트가 없고, S2가 단계 2에서 다시 씀
	row := func(flags ...string) map[string][]txVersion {
		return map[string][]txVersion{"acct": {{Block: 0, Lp: 1, Ctid: "(0,1)", Xmin: "752", Xmax: "0", Flags: flags}}}
	}
	caps := []txCapture{
		{AfterStep: 0, T: 0, Pages: row()},
		{AfterStep: 1, T: 10, Pages: row("HEAP_XMIN_COMMITTED")},
		{AfterStep: 1, AfterKill: true, T: 20, Pages: row()},
		{AfterStep: 2, T: 30, Pages: row("HEAP_XMIN_COMMITTED")},
	}
	steps := []txStep{{N: 1, Session: "S1"}, {N: 2, Session: "S2"}}
	writes := []hintWrite{
		{T: 5, Session: "S1", Xmin: 752, Bit: "HEAP_XMIN_COMMITTED", Tup: "0xa"},
		{T: 25, Session: "S2", Xmin: 752, Bit: "HEAP_XMIN_COMMITTED", Tup: "0xb"},
	}

	// When
	got := hintChanges(steps, caps, "acct", writes, nil)

	// Then
	var by []string
	for _, h := range got {
		by = append(by, h.By)
	}
	if !reflect.DeepEqual(by, []string{"S1", "S2"}) {
		t.Fatalf("hintChanges가 쓴 세션 = %v; want [S1 S2] (단계 1은 재시작 앞 관찰, 단계 2는 재시작 뒤 관찰부터)", by)
	}
}

// TestAttachServerStacksSamplesInTimeOrder 는 서버 사건에 호출 스택을 붙일 때 프로세스마다 시각 순서로 묻고(stackSampler는
// 시각이 늘어나는 순서로만 물을 수 있음), 고른 종류에만 붙이며, 바깥의 공통 부분을 그 프로세스의 가장 안쪽 주 함수에서 자르는지 봅니다.
func TestAttachServerStacksSamplesInTimeOrder(t *testing.T) {
	// Given: postmaster(1)의 신호 둘이 시각과 반대 순서로 들어 있고, 로그 사건은 스택을 받지 않습니다. 가짜 표본기는 물은 시각이
	// 줄어들면 실패를 남기고, 시각마다 다른 가장 안쪽 칸을 돌려줍니다.
	evs := []serverEvent{
		{Kind: "signal", Pid: 1, ns: 300},
		{Kind: "log", Pid: 1, ns: 200},
		{Kind: "signal", Pid: 1, ns: 100},
		{Kind: "redo", Pid: 2, ns: 150},
	}
	var asked []uint64
	last := map[int]uint64{}
	sampler := func(pid int) func(uint64, bool) ([]frameRef, string) {
		return func(ns uint64, atEnter bool) ([]frameRef, string) {
			if ns < last[pid] {
				t.Errorf("pid %d: 시각 %d를 %d 뒤에 물음", pid, ns, last[pid])
			}
			last[pid] = ns
			asked = append(asked, ns)
			top := map[uint64]string{100: "signal_child", 300: "signal_child", 150: "ApplyWalRecord"}[ns]
			outer := []frameRef{{Fn: "main"}, {Fn: "PostmasterMain"}, {Fn: "ServerLoop"}}
			if pid == 2 {
				outer = append(outer, frameRef{Fn: "postmaster_child_launch"}, frameRef{Fn: "StartupProcessMain"}, frameRef{Fn: "StartupXLOG"})
			}
			return append(outer, frameRef{Fn: top}), ""
		}
	}

	// When
	attachServerStacks(evs, sampler)

	// Then
	if !reflect.DeepEqual(asked, []uint64{100, 300, 150}) && !reflect.DeepEqual(asked, []uint64{150, 100, 300}) {
		t.Fatalf("물은 시각 = %v; want pid마다 늘어나는 순서", asked)
	}
	if evs[1].Stack != nil {
		t.Fatalf("로그 사건에 스택이 붙음: %v", evs[1].Stack)
	}
	if got := fnNames(evs[0].Stack); !reflect.DeepEqual(got, []string{"PostmasterMain", "ServerLoop", "signal_child"}) {
		t.Fatalf("postmaster 스택 = %v", got)
	}
	if got := fnNames(evs[3].Stack); !reflect.DeepEqual(got, []string{"StartupProcessMain", "StartupXLOG", "ApplyWalRecord"}) {
		t.Fatalf("startup 스택 = %v; want 가장 안쪽 주 함수 StartupProcessMain부터", got)
	}
}

func fnNames(s []frameRef) []string {
	var out []string
	for _, f := range s {
		out = append(out, f.Fn)
	}
	return out
}

// TestSpansOfPairsEnterAndExit 는 이름난 함수의 입구와 출구를 같은 프레임끼리 짝지어 구간으로 만들고, 출구 없이 기록이 끝난
// 호출은 열린 구간으로 남기며, 구간 밖의 호출은 빼는지 봅니다.
func TestSpansOfPairsEnterAndExit(t *testing.T) {
	// Given: A(프레임 0x90) 안에서 B(0x80)를 두 번 부르고, 두 번째 B는 출구가 없습니다. C는 고르지 않은 함수입니다.
	// 맨 앞의 A는 구간(1000~) 전에 끝난 호출입니다.
	tr := &traceFile{Pid: 5, Records: []record{
		{Kind: kindEnter, Fn: 0xA, Frame: 0x90, Ns: 10},
		{Kind: kindExit, Fn: 0xA, Frame: 0x90, Ns: 20},
		{Kind: kindEnter, Fn: 0xA, Frame: 0x90, Ns: 1000},
		{Kind: kindEnter, Fn: 0xB, Frame: 0x80, Ns: 1100},
		{Kind: kindEnter, Fn: 0xC, Frame: 0x70, Ns: 1150},
		{Kind: kindExit, Fn: 0xC, Frame: 0x70, Ns: 1160},
		{Kind: kindExit, Fn: 0xB, Frame: 0x80, Ns: 1200},
		{Kind: kindExit, Fn: 0xA, Frame: 0x90, Ns: 1300},
		{Kind: kindEnter, Fn: 0xB, Frame: 0x80, Ns: 1400},
		{Kind: kindEnter, Fn: 0xC, Frame: 0x70, Ns: 1500},
	}}
	want := func(fn uint64) (string, string, bool) {
		n, ok := map[uint64]string{0xA: "A", 0xB: "B"}[fn]
		return n, "f.c:" + n, ok
	}

	// When
	got := spansOf(tr, want, [2]uint64{1000, 2000}, func(ns uint64) float64 { return float64(ns) })

	// Then
	exp := []serverSpan{
		{Pid: 5, Fn: "A", T0: 1000, T1: 1300, At: "f.c:A"},
		{Pid: 5, Fn: "B", T0: 1100, T1: 1200, At: "f.c:B"},
		{Pid: 5, Fn: "B", T0: 1400, T1: 1500, At: "f.c:B", Open: true},
	}
	if !reflect.DeepEqual(got, exp) {
		t.Fatalf("spansOf = %+v; want %+v", got, exp)
	}
}

// TestWalExtentBeforeSeparatesWriteFromSync 는 kill 전에 pg_wal에 쓴 끝(pwrite)과, 그 뒤 같은 파일을 동기화까지 마친 끝을
// 따로 구하는지 봅니다. 동기화는 파일 단위라 다른 프로세스가 한 fdatasync도 앞선 쓰기를 내립니다. 값은 3-crash-sync-off-wait-a의
// WAL writer 쓰기(0/1782000~0/179C000)를 본뜬 것입니다.
func TestWalExtentBeforeSeparatesWriteFromSync(t *testing.T) {
	// Given: 프로세스 9가 0/179C000까지 쓰고 프로세스 8이 같은 파일을 fdatasync한 뒤, 프로세스 9가 0/179E000까지 더 썼지만
	// 동기화 전에 kill(시각 900)이 옵니다. kill 뒤의 쓰기와 다른 디렉터리의 쓰기는 세지 않습니다.
	wal := "/d/pg_wal/000000010000000000000001"
	sys := []sysEvent{
		{Pid: 9, Name: "pwrite64", FdPath: wal, Offset: 7872512, Ret: 106496, Ns: 100, DurNs: 10},
		{Pid: 8, Name: "fdatasync", FdPath: wal, Ns: 300, DurNs: 100},
		{Pid: 9, Name: "pwrite64", FdPath: wal, Offset: 7979008, Ret: 8192, Ns: 500, DurNs: 10},
		{Pid: 9, Name: "pwrite64", FdPath: "/d/base/5/16430", Offset: 0, Ret: 1 << 20, Ns: 600, DurNs: 10},
		{Pid: 9, Name: "pwrite64", FdPath: wal, Offset: 7987200, Ret: 8192, Ns: 950, DurNs: 10},
		{Pid: 9, Name: "fdatasync", FdPath: wal, Ns: 990, DurNs: 30},
	}

	// When
	w, s, at := walExtentBefore(sys, 900)

	// Then
	if lsnText(fmt.Sprint(w)) != "0/179E000" || lsnText(fmt.Sprint(s)) != "0/179C000" || at != 500 {
		t.Fatalf("walExtentBefore = %s(시각 %d), 동기화 %s; want 0/179E000(500), 0/179C000", lsnText(fmt.Sprint(w)), at, lsnText(fmt.Sprint(s)))
	}
}

// TestSpansOfMatchesExitByFrame 는 안쪽 호출이 출구 없이 빠져나간 뒤(ereport의 longjmp 등) 바깥 호출의 출구가 오면, 출구를
// 프레임이 같은 바깥 호출에 짝짓는지 봅니다. 이름만으로 짝지으면 가장 최근의 안쪽 호출이 닫힙니다.
func TestSpansOfMatchesExitByFrame(t *testing.T) {
	// Given: B(0x80) 안에서 B(0x60)를 다시 불렀고, 안쪽 B는 출구가 없습니다.
	tr := &traceFile{Pid: 5, Records: []record{
		{Kind: kindEnter, Fn: 0xB, Frame: 0x80, Ns: 100},
		{Kind: kindEnter, Fn: 0xB, Frame: 0x60, Ns: 110},
		{Kind: kindExit, Fn: 0xB, Frame: 0x80, Ns: 200},
		{Kind: kindEnter, Fn: 0xC, Frame: 0x80, Ns: 300},
	}}
	want := func(fn uint64) (string, string, bool) { return "B", "", fn == 0xB }

	// When
	got := spansOf(tr, want, [2]uint64{0, 1000}, func(ns uint64) float64 { return float64(ns) })

	// Then
	exp := []serverSpan{{Pid: 5, Fn: "B", T0: 100, T1: 200}, {Pid: 5, Fn: "B", T0: 110, T1: 300, Open: true}}
	if !reflect.DeepEqual(got, exp) {
		t.Fatalf("spansOf = %+v; want %+v", got, exp)
	}
}
