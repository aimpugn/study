package main

import (
	"os"
	"reflect"
	"testing"
)

// trampoline 은 arm64 Linux가 반환 uprobe를 건 함수의 돌아갈 주소를 바꿔 넣는 값입니다(실험 1의 열 벌에서 모두 이 값).
const trampoline = 0xfffffffff000

// TestRepairSitesRestoresHijackedReturnAddress 는 반환 uprobe를 건 함수 f의 입구·출구 기록에 trampoline 주소가 남은 꼴입니다.
// 짝지은 반환 사건의 원래 주소로 둘 다 고치고, 반환 uprobe가 없는 함수 g의 기록은 그대로 두어야 합니다.
func TestRepairSitesRestoresHijackedReturnAddress(t *testing.T) {
	// Given: main → f → g 순서의 기록. f의 입구(1)와 출구(4)만 부른 자리가 trampoline입니다.
	tr := &traceFile{Pid: 7, Records: []record{
		{Kind: kindEnter, Fn: 0x100, Site: 0x50},
		{Kind: kindEnter, Fn: 0x200, Site: trampoline},
		{Kind: kindEnter, Fn: 0x300, Site: 0x210},
		{Kind: kindExit, Fn: 0x300, Site: 0x210},
		{Kind: kindExit, Fn: 0x200, Site: trampoline},
		{Kind: kindExit, Fn: 0x100, Site: 0x50},
	}}
	evs := []uevent{{Name: "fret", Func: 0x200, Return: true, RetAddr: 0x120, Pid: 7, rec: 4, trace: tr}}

	// When
	got := repairSites(evs, []*traceFile{tr}, &symbols{byLow: map[uint64]*funcInfo{}})

	// Then
	if tr.Records[1].Site != 0x120 || tr.Records[4].Site != 0x120 {
		t.Fatalf("f의 부른 자리 = %#x, %#x; want 0x120", tr.Records[1].Site, tr.Records[4].Site)
	}
	if tr.Records[2].Site != 0x210 || tr.Records[3].Site != 0x210 {
		t.Fatalf("g의 부른 자리가 바뀌었습니다: %#x, %#x", tr.Records[2].Site, tr.Records[3].Site)
	}
	if got.Repaired != 2 || got.Left != 0 || got.Before["0xfffffffff000"] != 2 {
		t.Fatalf("repair = %+v", got)
	}
}

// TestRepairSitesLeavesCallWithoutReturn 는 오류(elog ERROR)의 longjmp로 반환하지 않은 호출입니다.
// 반환 사건이 없으니 고칠 값도 없고, 남은 수와 함수 이름으로 드러나야 합니다.
func TestRepairSitesLeavesCallWithoutReturn(t *testing.T) {
	// Given: f가 두 번 불렸고 첫 호출은 반환 사건이 있으며, 두 번째 호출은 출구 기록도 반환 사건도 없습니다.
	tr := &traceFile{Pid: 7, Records: []record{
		{Kind: kindEnter, Fn: 0x200, Site: trampoline},
		{Kind: kindExit, Fn: 0x200, Site: trampoline},
		{Kind: kindEnter, Fn: 0x200, Site: trampoline},
	}}
	evs := []uevent{{Func: 0x200, Return: true, RetAddr: 0x120, Pid: 7, rec: 1, trace: tr}}
	syms := &symbols{byLow: map[uint64]*funcInfo{0x200: {Name: "exec_simple_query", Low: 0x200}}}

	// When
	got := repairSites(evs, []*traceFile{tr}, syms)

	// Then
	if got.Repaired != 2 || got.Left != 1 || got.LeftFns["exec_simple_query"] != 1 {
		t.Fatalf("repair = %+v", got)
	}
}

// TestResultCtidsReadsPsqlTable 는 psql의 기본 표 출력에서 ctid 열만 읽는지 봅니다. 값 안의 괄호("(0,1)")를
// 끝 표시("(2 rows)")로 오인하면 안 됩니다.
func TestResultCtidsReadsPsqlTable(t *testing.T) {
	// Given
	out := []string{
		"   snap   | ctid  | xmin | xmax | id | balance ",
		"----------+-------+------+------+----+---------",
		" 753:753: | (0,1) |  752 |    0 |  1 |     100",
		" 753:753: | (0,2) |  752 |    0 |  2 |     200",
		"(2 rows)",
		"",
	}

	// When
	got, ok := resultCtids(out)

	// Then
	if !ok || !reflect.DeepEqual(got, []string{"(0,1)", "(0,2)"}) {
		t.Fatalf("got %v %v", got, ok)
	}
	if _, ok := resultCtids([]string{"UPDATE 1"}); ok {
		t.Fatal("ctid 열이 없는 출력은 ok가 거짓이어야 합니다")
	}
}

// TestResultCtidsWhenCtidIsFirstColumn 은 ctid가 첫 열이라 결과 행 자체가 "("로 시작하는 꼴입니다(실험 2의 SELECT).
// 끝 표시는 "(1 row)"처럼 행 수만 든 줄이어야 하고, 값이 "("로 시작하는 줄을 끝으로 보면 결과가 비어 버립니다.
func TestResultCtidsWhenCtidIsFirstColumn(t *testing.T) {
	// Given
	out := []string{" ctid  | xmin | xmax | id | balance ", "-------+------+------+----+---------", " (0,1) |  752 |    0 |  1 |     100", "(1 row)", ""}

	// When
	got, ok := resultCtids(out)

	// Then
	if !ok || !reflect.DeepEqual(got, []string{"(0,1)"}) {
		t.Fatalf("got %v %v; want [(0,1)] true", got, ok)
	}
}

// TestReadStraceGivesEndTimeForBothLineForms 는 strace가 한 줄로 끝난 호출에는 들어간 시각을, 나뉜 호출의 resumed 줄에는
// 끝난 시각을 적는 차이를 readStrace가 지우는지 봅니다. 두 호출 모두 Ns가 끝난 시각이어야, 입출력 막대(Ns-DurNs ~ Ns)가
// 실제 호출 구간에 놓입니다.
func TestReadStraceGivesEndTimeForBothLineForms(t *testing.T) {
	// Given: pid 7의 fdatasync는 한 줄(10.000000에 들어가 20 ms), pid 8의 epoll_pwait는 나뉜 꼴(11.000000에 들어가 11.500000에 끝남).
	f := t.TempDir() + "/all.strace"
	text := "7     10.000000 fdatasync(15</d/pg_wal/0001>) = 0 <0.020000>\n" +
		"8     11.000000 epoll_pwait(8<anon_inode:[eventpoll]>, 0x1, 2, 1000, NULL, 8 <unfinished ...>\n" +
		"7     10.030000 getpid() = 7 <0.000001>\n" +
		"8     11.500000 <... epoll_pwait resumed>[], 2, 1000, NULL, 8) = 0 <0.500000>\n"
	if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	// When
	evs, err := readStrace(f)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	end := map[string]uint64{}
	for _, e := range evs {
		end[e.Name] = e.Ns
	}
	if end["fdatasync"] != 10_020_000_000 {
		t.Fatalf("한 줄 fdatasync의 Ns = %d; want 10020000000(들어간 10 s + 20 ms)", end["fdatasync"])
	}
	if end["epoll_pwait"] != 11_500_000_000 {
		t.Fatalf("나뉜 epoll_pwait의 Ns = %d; want 11500000000(resumed 줄 시각)", end["epoll_pwait"])
	}
}

// TestHintChangesPrefersAddressOverValue 는 xmin이 같은 두 줄 가운데 한 줄에만 비트가 켜졌을 때, 주소를 아는 줄은 주소로
// 쓴 쪽을 찾는지 봅니다. 값으로만 맞추면 다른 줄에 쓴 기록(S2)을 이 줄의 설명으로 잘못 고릅니다.
func TestHintChangesPrefersAddressOverValue(t *testing.T) {
	// Given: 단계 1 전후 관찰. (0,2)에만 XMIN_COMMITTED가 새로 켜졌고, (0,1)과 (0,2)는 xmin이 같습니다.
	row := func(lp int, flags ...string) txVersion {
		return txVersion{Block: 0, Lp: lp, LpFlag: 1, Xmin: "752", Xmax: "0", Flags: flags}
	}
	caps := []txCapture{
		{AfterStep: 0, T: 0, Pages: map[string][]txVersion{"acct": {row(1, "HEAP_XMIN_COMMITTED"), row(2)}}},
		{AfterStep: 1, T: 100, Pages: map[string][]txVersion{"acct": {row(1, "HEAP_XMIN_COMMITTED"), row(2, "HEAP_XMIN_COMMITTED")}}},
	}
	steps := []txStep{{N: 1}}
	// S2가 먼저 다른 자리(0xa0, 그 줄은 이미 켜져 있던 (0,1)이 아닌 카탈로그 행이라 가정)에 같은 값으로 썼고, 그 뒤 S1이 (0,2)의 자리에 썼습니다.
	writes := []hintWrite{
		{T: 10, Session: "S2", Xmin: 752, Bit: "HEAP_XMIN_COMMITTED", Tup: "0xa0"},
		{T: 20, Session: "S1", Xmin: 752, Bit: "HEAP_XMIN_COMMITTED", Tup: "0xb0"},
	}
	addrs := map[string]map[string]bool{"(0,2)": {"0xb0": true}}

	// When
	got := hintChanges(steps, caps, "acct", writes, addrs)
	byValue := hintChanges(steps, caps, "acct", writes, nil)

	// Then
	if len(got) != 1 || got[0].Ctid != "(0,2)" || got[0].By != "S1" || got[0].Via != "addr" {
		t.Fatalf("주소로 맞춘 결과 = %+v; want (0,2)를 S1이 addr로", got)
	}
	if len(byValue) != 1 || byValue[0].By != "S2" || byValue[0].Via != "value" {
		t.Fatalf("값으로 맞춘 결과 = %+v; 이 시험의 전제(값만으로는 S2를 고름)가 깨졌습니다", byValue)
	}
}

// TestLabelHintsKeepsOnlyWatchedRows 는 SetHintBits 사건 가운데 관찰하는 표의 행 주소에 쓴 것만 남고 표와 ctid가 붙는지 봅니다.
// 주소를 모르는 쓰기(카탈로그 행 등)가 남으면 화면의 시간선이 관찰하지 않는 행의 힌트로 덮입니다.
func TestLabelHintsKeepsOnlyWatchedRows(t *testing.T) {
	// Given: 한 요청에 판정 하나와 힌트 쓰기 둘. 하나는 판정한 줄의 주소, 하나는 모르는 주소
	steps := []txStep{{Requests: []txRequest{{
		Checks: []txCheck{{Table: "acct", Ctid: "(0,1)", Tup: "0xa0"}},
		Events: []txEvent{
			{T: 1, Kind: "hint", F: map[string]string{"tup": "0xa0", "bit": "HEAP_XMAX_INVALID"}},
			{T: 2, Kind: "xid", F: map[string]string{"xid": "754"}},
			{T: 3, Kind: "hint", F: map[string]string{"tup": "0xff", "bit": "HEAP_XMIN_COMMITTED"}},
		},
	}}}}

	// When
	labelHints(steps, tupleAddrs(steps))

	// Then
	ev := steps[0].Requests[0].Events
	if len(ev) != 2 || ev[0].Kind != "hint" || ev[1].Kind != "xid" {
		t.Fatalf("events = %+v; want 0xa0의 hint와 xid만", ev)
	}
	if ev[0].F["table"] != "acct" || ev[0].F["ctid"] != "(0,1)" {
		t.Fatalf("hint F = %v; want table acct, ctid (0,1)", ev[0].F)
	}
}

// TestCaptureReadsSplitsOneCaptureIntoParts 는 관찰 하나의 질의들이 부분마다 [받은 시각, 다음 질의를 받은 시각]으로 나뉘는지 봅니다.
// 다음 관찰 전의 다른 질의(드라이버가 기다림을 확인하는 질의)는 부분이 아니지만 앞 부분의 끝이 되어야 합니다.
func TestCaptureReadsSplitsOneCaptureIntoParts(t *testing.T) {
	// Given: 관찰 둘(100, 1000). 첫 관찰은 페이지(110) → 세션 상태(200) → 락(400) → 기다림 확인(500), 둘째는 페이지(1010)만
	qmsg := func(text string) []byte {
		n := len(text) + 5
		return append([]byte{'Q', byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}, append([]byte(text), 0)...)
	}
	recv := func(ns uint64, text string) sysEvent {
		return sysEvent{Name: "recvfrom", Ns: ns, Ret: 10, Data: qmsg(text)}
	}
	obs := []sysEvent{
		recv(110, "SELECT b FROM heap_page_item_attrs(get_raw_page('acct', 0), 'acct')"),
		recv(200, "SELECT pid, coalesce(backend_xid::text, '') FROM pg_stat_activity"),
		recv(400, "SELECT pid, locktype FROM pg_locks"),
		recv(500, "SELECT pid FROM pg_stat_activity WHERE wait_event_type = 'Lock'"),
		recv(1010, "SELECT b FROM heap_page_item_attrs(get_raw_page('acct', 0), 'acct')"),
	}

	// When
	r := captureReads([]uint64{100, 1000}, obs)

	// Then
	want0 := map[string][2]uint64{"pages": {110, 200}, "activity": {200, 400}, "locks": {400, 500}}
	if len(r) != 2 || !reflect.DeepEqual(r[0], want0) {
		t.Fatalf("첫 관찰 = %v; want %v", r[0], want0)
	}
	if !reflect.DeepEqual(r[1], map[string][2]uint64{"pages": {1010, 0}}) {
		t.Fatalf("둘째 관찰 = %v; want pages [1010 0](마지막이라 끝을 모름)", r[1])
	}
}

// TestReadStraceKeepsFdPathOfSplitCall 은 다른 pid의 줄이 끼어 두 줄로 나뉜 호출에서도 대상 파일의 경로를 읽는지 봅니다.
// 2026-10-14 실험 3-crash-sync-on-a에서 나뉜 fdatasync 줄은 "7<경로> <unfinished ...>"처럼 경로 뒤에 공백이 남아, 경로를 읽지
// 못하고 빈 파일 이름으로 나왔습니다.
func TestReadStraceKeepsFdPathOfSplitCall(t *testing.T) {
	// Given: 한 줄 호출과, 같은 호출이 나뉜 두 줄
	f := t.TempDir() + "/s.strace"
	lines := "1909  1791987418.842848 fdatasync(7</tmp/pgi/data/pg_wal/000000010000000000000001>) = 0 <0.001723>\n" +
		"1909  1791987420.054135 fdatasync(7</tmp/pgi/data/pg_wal/000000010000000000000001> <unfinished ...>\n" +
		"1901  1791987420.060000 kill(1901, SIGURG) = 0 <0.000002>\n" +
		"1909  1791987420.118151 <... fdatasync resumed>) = 0 <0.064002>\n"
	if err := os.WriteFile(f, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}

	// When
	evs, err := readStrace(f)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, e := range evs {
		if e.Name == "fdatasync" {
			paths = append(paths, e.FdPath)
		}
	}
	want := "/tmp/pgi/data/pg_wal/000000010000000000000001"
	if len(paths) != 2 || paths[0] != want || paths[1] != want {
		t.Fatalf("fdatasync 경로 = %q; want 둘 다 %q", paths, want)
	}
}
