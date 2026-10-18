package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// treeBase 는 시험 사건의 기준 시각(2026-10-15T09:00:00Z)입니다. treeRel은 그 기준부터의 마이크로초입니다.
const treeBase = uint64(1792054800000000000)

func treeRel(ns uint64) float64 { return float64(int64(ns-treeBase)) / 1e3 }

// at 은 기준부터 us 마이크로초 뒤의 JFR 시각 글입니다.
func at(us int) string {
	return "2026-10-15T09:00:00." + pad6(us) + "000Z"
}

// pad6 은 0~999999를 여섯 자리로 채운 글입니다.
func pad6(n int) string {
	s := []byte("000000")
	for i := 5; i >= 0 && n > 0; i-- {
		s[i] = byte('0' + n%10)
		n /= 10
	}
	return string(s)
}

// jfrThrowEvent 는 시험용 예외 사건입니다. callers는 예외를 만든 쪽부터 바깥으로의 칸입니다.
func jfrThrowEvent(cls, thread, start string, callers ...jfrMethod) jfrEvent {
	var e jfrEvent
	e.Type = "jdk.JavaExceptionThrow"
	e.Values.StartTime = start
	e.Values.EventThread.JavaName = thread
	e.Values.ThrownClass = &struct {
		Name string `json:"name"`
	}{cls}
	var frames []jfrFrame
	for _, c := range callers {
		frames = append(frames, jfrFrame{Method: c, Line: 1})
	}
	e.Values.StackTrace = &jfrStack{Frames: frames}
	return e
}

// nodeByLabel 은 이름표가 label인 마디들입니다.
func nodeByLabel(t *callTree, label string) []*callNode {
	var out []*callNode
	for _, n := range t.Nodes {
		if n.Label == label {
			out = append(out, n)
		}
	}
	return out
}

// TestBuildCallTreeAlignsByOuterStack 은 같은 메서드가 겹쳐 불렸고 바깥 호출들이 스택에만 보일 때, 안쪽 기록이 바깥 칸에 붙지 않고
// 스택에만 보인 마디가 한 벌만 생기는지 봅니다. 모양은 2026-10-15 r5-requires-new-same-row입니다(바깥 invoke·invokeWithinTransaction과
// 앱 메서드가 스택에만 보이고, 안쪽 invokeWithinTransaction과 그 안의 JdbcTemplate.update가 기록됨). 스택을 겹치는 길이만큼만 견주면
// update의 바깥 invokeWithinTransaction 칸이 안쪽 기록과 맞아(바깥 스택 [invoke]가 안쪽 기록 스택의 앞부분과 같음), 그 아래 앱 메서드
// 마디가 안쪽 invokeWithinTransaction 밑에 하나 더 생깁니다.
func TestBuildCallTreeAlignsByOuterStack(t *testing.T) {
	// Given
	ti, tas := "org/springframework/transaction/interceptor/TransactionInterceptor", "org/springframework/transaction/interceptor/TransactionAspectSupport"
	invoke := jm(ti, "invoke", "(Lm;)Ljava/lang/Object;")
	within := jm(tas, "invokeWithinTransaction", "(Lx;)Ljava/lang/Object;")
	outer := jm("lab/Outer", "requiresNewSameRow", "()V")
	proxy := jm("lab/Inner$$SpringCGLIB$$0", "requiresNewSameRow", "()V")
	events := []jfrEvent{
		jfrSpanEvent(tas, "invokeWithinTransaction", "(Lx;)Ljava/lang/Object;", "main", at(10), "PT0.000050S", invoke, proxy, outer, within, invoke),
		jfrSpanEvent("org/springframework/jdbc/core/JdbcTemplate", "update", "(Ljava/lang/String;)I", "main", at(20), "PT0.000010S",
			within, invoke, proxy, outer, within, invoke),
	}
	filter := "org.springframework.transaction.interceptor.TransactionInterceptor::invoke;" +
		"org.springframework.transaction.interceptor.TransactionAspectSupport::invokeWithinTransaction;" +
		"org.springframework.jdbc.core.JdbcTemplate::update;lab.Outer"
	j, err := buildSpringJava(events, filter, treeRel, 0)
	if err != nil {
		t.Fatal(err)
	}

	// When
	tree := buildCallTree(j, "main", 0, 1000)

	// Then: 바깥 invoke → 바깥 invokeWithinTransaction → 앱 메서드 → 안쪽 invoke → 안쪽 invokeWithinTransaction → update.
	if len(tree.Nodes) != 6 || tree.Ambiguous != 0 {
		t.Fatalf("마디 %d개, 모호 %d; want 6, 0", len(tree.Nodes), tree.Ambiguous)
	}
	if apps := nodeByLabel(tree, "Outer.requiresNewSameRow"); len(apps) != 1 || apps[0].T1 != nil || apps[0].Depth != 2 {
		t.Fatalf("앱 메서드 마디 = %+v; want 스택에만 보인 마디 하나, 깊이 2", apps)
	}
	upd := nodeByLabel(tree, "JdbcTemplate.update")
	if len(upd) != 1 || upd[0].Depth != 5 || upd[0].Site != 1 {
		t.Fatalf("update = %+v; want 깊이 5, 부른 줄 1", upd)
	}
	var chain []string
	for n := tree.Nodes[upd[0].Parent]; ; n = tree.Nodes[n.Parent] {
		recorded := n.T1 != nil
		chain = append([]string{n.Label + map[bool]string{true: "(기록)", false: "(스택)"}[recorded]}, chain...)
		if n.Parent < 0 {
			break
		}
	}
	want := []string{"TransactionInterceptor.invoke(스택)", "TransactionAspectSupport.invokeWithinTransaction(스택)", "Outer.requiresNewSameRow(스택)",
		"TransactionInterceptor.invoke(스택)", "TransactionAspectSupport.invokeWithinTransaction(기록)"}
	if !reflect.DeepEqual(chain, want) {
		t.Fatalf("update의 조상 = %v; want %v", chain, want)
	}
	innerInvoke := tree.Nodes[tree.Nodes[upd[0].Parent].Parent]
	if !reflect.DeepEqual(innerInvoke.Hidden, []string{"lab.Inner$$SpringCGLIB$$0::requiresNewSameRow:1"}) ||
		!reflect.DeepEqual(innerInvoke.HiddenKinds, []string{"프록시·반사"}) {
		t.Fatalf("안쪽 invoke의 숨은 칸 = %v %v; want 프록시 칸 하나", innerInvoke.Hidden, innerInvoke.HiddenKinds)
	}
}

// TestAnnotateTreeEndKinds 는 끝난 꼴 다섯 가지를 가르는지 봅니다. 정상(return), 같은 시작의 사건 둘(twice), 마지막 자식 뒤에 자기
// 코드가 만든 예외(throw), 스택에만 보임(passed), 스택에만 보인 부모의 마지막 자식(rethrow, 추론)입니다.
func TestAnnotateTreeEndKinds(t *testing.T) {
	// Given: R(0~100) 안에 A(정상), B(사건 둘), C(예외를 만듦), D(스택에만 보임) 안의 E.
	r := jm("lab/Outer", "r", "()V")
	d := jm("lab/Outer", "d", "()V")
	c := jm("lab/Outer", "c", "()V")
	events := []jfrEvent{
		jfrSpanEvent("lab/Outer", "r", "()V", "main", at(0), "PT0.000100S"),
		jfrSpanEvent("lab/Outer", "a", "()V", "main", at(10), "PT0.000010S", r),
		jfrSpanEvent("lab/Outer", "b", "()V", "main", at(30), "PT0.000005S", r),
		jfrSpanEvent("lab/Outer", "b", "()V", "main", at(30), "PT0.000008S", r),
		jfrSpanEvent("lab/Outer", "c", "()V", "main", at(40), "PT0.000020S", r),
		jfrThrowEvent("java/lang/IllegalStateException", "main", at(55), jm("java/lang/IllegalStateException", "<init>", "()V"), c, r),
		jfrSpanEvent("lab/Outer", "e", "()V", "main", at(70), "PT0.000010S", d, r),
	}
	j, err := buildSpringJava(events, "lab.Outer", treeRel, 0)
	if err != nil {
		t.Fatal(err)
	}
	tree := buildCallTree(j, "main", 0, 1000)

	// When
	annotateTree(tree, j, nil, nil, nil, nil)

	// Then
	got := map[string]string{}
	for _, n := range tree.Nodes {
		got[n.Label] = n.End
	}
	want := map[string]string{"Outer.r": endReturn, "Outer.a": endReturn, "Outer.b": endTwice, "Outer.c": endThrow, "Outer.d": endPassed, "Outer.e": endRethrow}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("끝난 꼴 = %v; want %v", got, want)
	}
	if rr := nodeByLabel(tree, "Outer.r")[0]; rr.Self != nil {
		t.Fatalf("R의 자기 시간 = %v; want nil(자식 D의 끝을 모름)", *rr.Self)
	}
}

// TestAnnotateTreeSelfTimeExcludesKidsAndPauses 는 자기 시간이 자식들의 길이와, 자식 밖에 놓인 실험 도구의 멈춤을 빼는지 봅니다.
// 자식 안의 멈춤은 그 자식의 길이에 들어 있으므로 부모에서 두 번 빼지 않습니다.
func TestAnnotateTreeSelfTimeExcludesKidsAndPauses(t *testing.T) {
	// Given: S(0~100) 안에 K1(10~30), K2(40~50). 멈춤 [15, 25](K1 안)와 [60, 90](자식 밖).
	s := jm("lab/Outer", "s", "()V")
	events := []jfrEvent{
		jfrSpanEvent("lab/Outer", "s", "()V", "main", at(0), "PT0.000100S"),
		jfrSpanEvent("lab/Outer", "k1", "()V", "main", at(10), "PT0.000020S", s),
		jfrSpanEvent("lab/Outer", "k2", "()V", "main", at(40), "PT0.000010S", s),
	}
	j, err := buildSpringJava(events, "lab.Outer", treeRel, 0)
	if err != nil {
		t.Fatal(err)
	}
	tree := buildCallTree(j, "main", 0, 1000)

	// When
	annotateTree(tree, j, nil, nil, [][2]float64{{15, 25}, {60, 90}}, nil)

	// Then: S = 100 − 30(멈춤 둘) − (20 − 10) − 10 = 40, K1 = 20 − 10 = 10.
	self := map[string]float64{}
	for _, n := range tree.Nodes {
		self[n.Label] = *n.Self
	}
	if want := map[string]float64{"Outer.s": 40, "Outer.k1": 10, "Outer.k2": 10}; !reflect.DeepEqual(self, want) {
		t.Fatalf("자기 시간 = %v; want %v", self, want)
	}
}

// TestBindingReplayTracksSuspendAndReturn 은 REQUIRES_NEW의 차례(바깥 C1 빌림 → 떼어 둠 → 안쪽 C2 빌림 → 안쪽 정리·돌려줌 → 바깥 다시
// 묶음 → 바깥 정리·돌려줌)로 묶임과 풀 사용 구간, 옮긴 연결의 효과를 셈하는지 봅니다.
func TestBindingReplayTracksSuspendAndReturn(t *testing.T) {
	// Given
	dstm := "org.springframework.jdbc.datasource.DataSourceTransactionManager::"
	mk := func(id int, fn string, t0 float64) *callNode {
		return &callNode{ID: id, Parent: -1, Fn: fn, T0: t0, Kids: []int{}}
	}
	tree := &callTree{Nodes: []*callNode{
		mk(0, dstm+"doSuspend", 100),
		mk(1, dstm+"doCleanupAfterCompletion", 300),
		mk(2, "com.zaxxer.hikari.pool.ProxyConnection::close", 310),
		mk(3, dstm+"doResume", 320),
		mk(4, dstm+"doCleanupAfterCompletion", 500),
		mk(5, "com.zaxxer.hikari.pool.ProxyConnection::close", 510),
	}}
	borrows := []springBorrow{
		{T0: 10, T1: 20, Thread: "main", Conn: "C1", Result: "ok"},
		{T0: 110, T1: 120, Thread: "main", Conn: "C2", Result: "ok"},
	}

	// When
	states, slots := bindingReplay(tree, borrows, 0)

	// Then
	type st struct {
		t     float64
		bound string
		held  []string
	}
	var got []st
	for _, s := range states {
		got = append(got, st{s.T, s.Bound, s.Held})
	}
	want := []st{{0, "", []string{}}, {20, "C1", []string{}}, {100, "", []string{"C1"}}, {120, "C2", []string{"C1"}},
		{300, "", []string{"C1"}}, {320, "C1", []string{}}, {500, "", []string{}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("묶임 = %v; want %v", got, want)
	}
	if want := []poolSlot{{"C1", 20, 510}, {"C2", 120, 310}}; !reflect.DeepEqual(slots, want) {
		t.Fatalf("풀 사용 = %v; want %v", slots, want)
	}
	moved := []string{}
	for _, n := range tree.Nodes {
		for _, e := range n.Effects {
			moved = append(moved, e.Kind+":"+e.Conn)
		}
	}
	if want := []string{"suspend:C1", "return:C2", "resume:C1", "return:C1"}; !reflect.DeepEqual(moved, want) {
		t.Fatalf("효과 = %v; want %v", moved, want)
	}
}

// branchNode 는 갈림 시험용 마디입니다. kids는 자식 마디 번호입니다.
func branchNode(id int, fn, end string, site int, kids ...int) *callNode {
	return &callNode{ID: id, Parent: -1, Fn: fn, Label: fn, End: end, Site: site, Kids: kids, Why: []string{}}
}

// TestComputeBranchesSeparatesCutFromBranch 는 갈림의 규칙(computeBranches 주석)을 사례 넷으로 봅니다.
//   - M: 짧은 길의 마지막 자식이 정상으로 끝남 → 그 마디의 조건이 뒤를 건너뜀 → 갈림
//   - N: 짧은 길의 마지막 자식이 예외로 끝남 → 아래에서 올라온 예외로 뒤가 빠짐 → 끊김, 갈림 아님
//   - G: 자식 차례가 같아도 끝난 꼴이 다름(HikariPool.getConnection(long)의 재빌리기와 시간 초과) → 길 셋, 갈린 줄은 isConnectionDead를 부른 168
//   - W: 앱 코드 자식의 이름만 다름 → 같은 길
func TestComputeBranchesSeparatesCutFromBranch(t *testing.T) {
	// Given
	r1 := []*callNode{
		branchNode(0, "M", endReturn, 0, 1, 2), branchNode(1, "A", endReturn, 10), branchNode(2, "B", endReturn, 11),
		branchNode(3, "N", endReturn, 0, 4, 5), branchNode(4, "A", endReturn, 20), branchNode(5, "B", endReturn, 21),
		branchNode(6, "G", endReturn, 0, 7, 8), branchNode(7, "borrow", endReturn, 162), branchNode(8, "dead", endReturn, 168),
		branchNode(9, "W", endReturn, 0, 10, 11), branchNode(10, "lab.Outer::x", endReturn, 380), branchNode(11, "C", endReturn, 416),
	}
	r2 := []*callNode{
		branchNode(0, "M", endReturn, 0, 1), branchNode(1, "A", endReturn, 10),
		branchNode(2, "N", endTwice, 0, 3), branchNode(3, "A", endPassed, 20),
		branchNode(4, "G", endReturn, 0, 5), branchNode(5, "borrow", endReturn, 162),
		branchNode(6, "W", endReturn, 0, 7, 8), branchNode(7, "lab.Inner::y", endReturn, 380), branchNode(8, "C", endReturn, 416),
	}
	r3 := []*callNode{branchNode(0, "G", endThrow, 0, 1), branchNode(1, "borrow", endReturn, 162)}

	// When
	b := computeBranches([]namedTree{{"r1", r1}, {"r2", r2}, {"r3", r3}})

	// Then
	if g := b["M"]; !g.Flagged || len(g.Alts) != 2 || g.Common != 1 || !reflect.DeepEqual(g.Lines, []int{11}) {
		t.Fatalf("M = %+v; want 갈림, 길 둘, 공통 1, 갈린 줄 11", g)
	}
	if g := b["N"]; g.Flagged || !g.Alts[1].Cut {
		t.Fatalf("N = %+v; want 갈림 아님, 둘째 길 끊김", g)
	}
	if g := b["G"]; !g.Flagged || len(g.Alts) != 3 || !reflect.DeepEqual(g.Lines, []int{168}) {
		t.Fatalf("G = %+v; want 갈림, 길 셋, 갈린 줄 168", g)
	}
	if g := b["W"]; g.Flagged || len(g.Alts) != 1 || g.Alts[0].N != 2 {
		t.Fatalf("W = %+v; want 길 하나를 두 번", g)
	}
	if !contains(r3[0].Why, whyBranch) || contains(r2[2].Why, whyBranch) {
		t.Fatalf("렌즈 이유: G %v, N %v; want G에만 갈림", r3[0].Why, r2[2].Why)
	}
}

// TestSplitRequestsKeepsTypesAndStatementName 은 요청마다 앱이 보낸 메시지 종류의 차례와 준비 문장 이름을 남기는지 봅니다. 모양은
// 2026-10-15 r5-reborrow-a의 둘째 COMMIT(처음 COMMIT이 만든 S_1을 Parse 없이 Bind로 다시 씀)입니다.
func TestSplitRequestsKeepsTypesAndStatementName(t *testing.T) {
	// Given
	mac := []wireMsg{
		{Dir: "F", Type: "startup", Ns: 1}, {Dir: "B", Type: "Z", Status: "I", Ns: 2},
		{Dir: "F", Type: "P", Name: "S_1", SQL: "COMMIT", Ns: 10}, {Dir: "F", Type: "B", Name: "S_1", Ns: 10}, {Dir: "F", Type: "E", Ns: 10},
		{Dir: "F", Type: "S", Ns: 10}, {Dir: "B", Type: "Z", Status: "I", Ns: 20},
		{Dir: "F", Type: "B", Name: "S_1", Ns: 30}, {Dir: "F", Type: "E", Ns: 30}, {Dir: "F", Type: "S", Ns: 30}, {Dir: "B", Type: "Z", Status: "I", Ns: 40},
	}

	// When
	reqs, _ := splitRequests(mac, mac, usRel)

	// Then
	if len(reqs) != 3 || !reflect.DeepEqual(reqs[1].Types, []string{"P", "B", "E", "S"}) || reqs[1].Stmt != "S_1" ||
		!reflect.DeepEqual(reqs[2].Types, []string{"B", "E", "S"}) || reqs[2].Stmt != "S_1" || len(reqs[2].SQL) != 0 {
		t.Fatalf("요청 = %+v; want P B E S(S_1), B E S(S_1, SQL 없음)", reqs)
	}
}

// TestDescParamsNamesArgumentTypes 는 JVM 서술자의 인자 형을 짧은 이름으로 읽는지 봅니다.
func TestDescParamsNamesArgumentTypes(t *testing.T) {
	if got := descParams("(JLjava/lang/String;[I[[Ljava/lang/Object;Z)V"); !reflect.DeepEqual(got, []string{"long", "String", "int[]", "Object[][]", "boolean"}) {
		t.Fatalf("descParams = %v", got)
	}
	if got := descParams("()V"); len(got) != 0 {
		t.Fatalf("descParams(()V) = %v; want 없음", got)
	}
}

// TestCheckAppSitesFindsShiftedLines 는 앱 코드가 다음 칸을 부른 줄의 대조가, 기록 때와 같은 소스에서는 아무것도 내지 않고 머리에 한
// 줄이 늘어난 소스에서는 어긋난 자리를 모두 내는지 봅니다. 모양은 2026-10-15 앞선 기록 열두 벌입니다(Outer.java 머리 주석이 한 줄
// 늘어 부른 줄이 모두 한 줄 아래로 밀림). 부모가 앱 코드인 마디(update), 추적하지 않은 칸을 거쳐 부른 마디(프록시 뒤의 invoke),
// 추적하지 않은 앱 칸(Main.main)을 함께 넣습니다.
func TestCheckAppSitesFindsShiftedLines(t *testing.T) {
	// Given
	write := func(root, rel, text string) {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, rel), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	outer := "class Outer {\n  void run() {\n    jdbc.update(\"x\");\n    inner.go();\n  }\n}\n"
	main := "class Main {\n  outer.run();\n}\n"
	same, shifted := t.TempDir(), t.TempDir()
	write(same, "lab/Outer.java", outer)
	write(same, "lab/Main.java", main)
	write(shifted, "lab/Outer.java", "// 한 줄 늘어난 머리 주석\n"+outer)
	write(shifted, "lab/Main.java", "// 한 줄 늘어난 머리 주석\n"+main)
	tree := &callTree{Nodes: []*callNode{
		{ID: 0, Parent: -1, Fn: "org.Interceptor::invoke", Hidden: []string{"lab.Main::main:2", "lab.Outer$$Proxy::run:-1"}},
		{ID: 1, Parent: 0, Fn: "lab.Outer::run", Site: 10},
		{ID: 2, Parent: 1, Fn: "org.Jdbc::update", Site: 3},
		{ID: 3, Parent: 1, Fn: "org.Interceptor::invoke", Site: 4, Hidden: []string{"lab.Inner$$Proxy::go:-1"}},
	}}

	// When
	ok := checkAppSites(tree, &javaSources{roots: []string{same}, files: map[string][]string{}})
	miss := checkAppSites(tree, &javaSources{roots: []string{shifted}, files: map[string][]string{}})

	// Then
	if len(ok) != 0 {
		t.Fatalf("같은 소스에서 어긋남 %v; want 없음", ok)
	}
	if want := []string{"lab.Main:2 → run", "lab.Outer:3 → update", "lab.Outer:4 → go"}; !reflect.DeepEqual(miss, want) {
		t.Fatalf("늘어난 소스에서 어긋남 %v; want %v", miss, want)
	}
}

// TestAnnotateTreeMarksRollbackOnlyAndValidate 는 연결 위로 아무것도 보내지 않는 두 상태 변화(롤백 전용 표시, 연결 생존 검사)를
// 효과로 붙이는지 봅니다. 사용성 시험 1차에서 이 둘이 기본 화면에 없어 참가자가 「전부」를 켜야 이유를 찾았습니다(WORK C4, C5).
func TestAnnotateTreeMarksRollbackOnlyAndValidate(t *testing.T) {
	// Given: R 아래에 롤백 전용 표시, 생존 검사, 그리고 효과가 없는 형제 하나.
	r := jm("lab/Outer", "r", "()V")
	events := []jfrEvent{
		jfrSpanEvent("lab/Outer", "r", "()V", "main", at(0), "PT0.000100S"),
		jfrSpanEvent("org/springframework/jdbc/datasource/DataSourceTransactionManager", "doSetRollbackOnly", "()V", "main", at(10), "PT0.000010S", r),
		jfrSpanEvent("com/zaxxer/hikari/pool/PoolBase", "isConnectionDead", "()Z", "main", at(30), "PT0.000010S", r),
		jfrSpanEvent("org/springframework/jdbc/datasource/DataSourceTransactionManager", "doGetTransaction", "()V", "main", at(50), "PT0.000010S", r),
	}
	j, err := buildSpringJava(events, "lab.Outer", treeRel, 0)
	if err != nil {
		t.Fatal(err)
	}
	tree := buildCallTree(j, "main", 0, 1000)

	// When
	annotateTree(tree, j, nil, nil, nil, nil)

	// Then
	got := map[string][]string{}
	for _, n := range tree.Nodes {
		for _, e := range n.Effects {
			got[n.Label] = append(got[n.Label], e.Kind)
		}
	}
	want := map[string][]string{"DataSourceTransactionManager.doSetRollbackOnly": {"rollbackOnly"}, "PoolBase.isConnectionDead": {"validate"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("효과 = %v; want %v", got, want)
	}
}

// TestWaitCyclesClosesOnlyWhenHolderIsHeldByWaitingThread 는 서버의 기다림을 앱 쪽 사실과 이어, 기다림의 고리가 앱 스레드를 거쳐
// 닫힐 때만 Cycle을 세우는지 봅니다. 같은 기다림에서 쥔 연결이 떼어 둔 상태가 아니면(다른 곳에서 쥐고 있으면) 고리가 아닙니다.
func TestWaitCyclesClosesOnlyWhenHolderIsHeldByWaitingThread(t *testing.T) {
	// Given: C1이 번호 753을 받았고, C2의 2번 요청이 +1,000에 753을 기다리기 시작해 +4,000에 오류로 끝남.
	conns := []springConn{
		{Name: "C1", Requests: []springRequest{{Seq: 1, Server: &txRequest{Events: []txEvent{{T: 100, Kind: "xid", F: map[string]string{"xid": "753"}}}}}}},
		{Name: "C2", Requests: []springRequest{{Seq: 2, Done: 4010, Server: &txRequest{Events: []txEvent{
			{T: 1000, Kind: "xact_wait", F: map[string]string{"xid": "753"}},
			{T: 4000, Kind: "ereport"},
		}}}}},
	}
	waits := serverWaits(conns)
	tree := &callTree{Nodes: []*callNode{
		{ID: 0, Parent: -1, Fn: "lab.Outer::run"},
		{ID: 1, Parent: 0, Fn: "org.postgresql.core.v3.QueryExecutorImpl::execute", T0: 900, T1: f64(4020), Effects: []callEffect{{Kind: "wire", Conn: "C2", Seq: []int{2}}}},
	}}
	// 같은 요청을 보냈지만 기다림이 끝나기 전에 끝난 마디(보낸 뒤 응답을 기다리지 않은 꼴)는 Call이 되지 않아야 합니다.
	early := &callTree{Nodes: []*callNode{
		{ID: 0, Parent: -1, Fn: "lab.Outer::run"},
		{ID: 1, Parent: 0, Fn: "org.postgresql.core.v3.QueryExecutorImpl::execute", T0: 900, T1: f64(1500), Effects: []callEffect{{Kind: "wire", Conn: "C2", Seq: []int{2}}}},
	}}
	held := []bindState{{T: 0, Held: []string{}}, {T: 500, Bound: "C2", Held: []string{"C1"}}}
	free := []bindState{{T: 0, Held: []string{}}, {T: 500, Bound: "C2", Held: []string{}}, {T: 2000, Bound: "C2", Held: []string{"C1"}}}

	// When
	got := waitCycles(tree, waits, held, conns)
	other := waitCycles(tree, waits, free, conns)

	// Then
	want := []waitCycle{{Kind: "server", T0: 1000, T1: 4000, Waiter: "C2", Xid: "753", Holder: "C1", Thread: sceneThread, Call: 1, HeldBy: sceneThread, Cycle: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("떼어 둔 C1: %+v; want %+v", got, want)
	}
	if len(other) != 1 || other[0].Cycle || other[0].HeldBy != "" {
		t.Fatalf("기다림이 시작될 때 C1을 떼어 두지 않음: %+v; want Cycle false, HeldBy 빈 글", other)
	}
	if e := waitCycles(early, waits, held, conns); len(e) != 1 || e[0].Call != -1 || e[0].Cycle {
		t.Fatalf("기다림 전에 끝난 마디: %+v; want Call -1, Cycle false", e)
	}
}

// TestPoolConfigOfReadsLiteralTimeout 는 앱 설정 소스에서 연결 시간 초과를, HikariCP 소스에서 생존 검사를 건너뛰는 창의 기본값과
// 그 값을 쓰는 조건의 줄을 읽는지 봅니다.
func TestPoolConfigOfReadsLiteralTimeout(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "lab"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := "class LabConfig {\n  void ds() {\n    c.setMaximumPoolSize(pool);\n    c.setConnectionTimeout(2000);\n  }\n}\n"
	if err := os.WriteFile(filepath.Join(root, "lab", "LabConfig.java"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	pool := "class HikariPool {\n   private final long aliveBypassWindowMs = Long.getLong(\"com.zaxxer.hikari.aliveBypassWindowMs\", MILLISECONDS.toMillis(500));\n" +
		"   Connection getConnection(long hardTimeout) {\n      final var now = currentTime();\n" +
		"      if (poolEntry.isMarkedEvicted() || (elapsedMillis(poolEntry.lastAccessed, now) > aliveBypassWindowMs && isConnectionDead(poolEntry.connection))) {\n   }\n}\n"
	if err := os.MkdirAll(filepath.Join(root, "com", "zaxxer", "hikari", "pool"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "com", "zaxxer", "hikari", "pool", "HikariPool.java"), []byte(pool), 0o644); err != nil {
		t.Fatal(err)
	}

	// When
	got := poolConfigOf(&javaSources{roots: []string{root}, files: map[string][]string{}})
	none := poolConfigOf(&javaSources{roots: []string{t.TempDir()}, files: map[string][]string{}})

	// Then
	if want := (poolConfig{ConnectionTimeoutMs: 2000, ConnectionTimeoutAt: "lab.LabConfig:4", AliveBypassMs: 500,
		AliveBypassAt: hikariPool + ":2", AliveCheckAt: hikariPool + ":5"}); got != want {
		t.Fatalf("poolConfigOf = %+v; want %+v", got, want)
	}
	if none != (poolConfig{}) {
		t.Fatalf("소스가 없을 때 = %+v; want 빈 값", none)
	}
}

// f64 는 끝 시각처럼 없을 수 있는 값(*float64)을 시험에서 만듭니다.
func f64(v float64) *float64 { return &v }

// TestPoolWaitCyclesNeedsEveryInUseSlotHeldBySameThread 는 시간 초과로 끝난 빌리기를, 풀의 칸을 모두 장면 스레드가 떼어 둔 채
// 쥐고 있을 때만 고리로 세는지 봅니다. 풀이 2칸인데 1칸만 쓰이면 기다림의 원인이 이 스레드가 아니므로 고리가 아닙니다.
func TestPoolWaitCyclesNeedsEveryInUseSlotHeldBySameThread(t *testing.T) {
	// Given: C1을 +100에 빌렸고 +500에 떼어 둔 뒤, +600~+2,600 빌리기가 시간 초과로 끝남. 시간 초과가 아닌 빌리기도 하나 있음.
	tree := &callTree{Nodes: []*callNode{
		{ID: 0, Parent: -1, Fn: "lab.Outer::run"},
		{ID: 1, Parent: 0, Fn: "com.zaxxer.hikari.util.ConcurrentBag::borrow", T0: 100, T1: f64(110), Effects: []callEffect{{Kind: "borrow", Result: "ok", Conn: "C1"}}},
		{ID: 2, Parent: 0, Fn: "com.zaxxer.hikari.util.ConcurrentBag::borrow", T0: 600, T1: f64(2600), Effects: []callEffect{{Kind: "borrow", Result: "timeout"}}},
	}}
	slots := []poolSlot{{Conn: "C1", T0: 110, T1: 3000}}
	binding := []bindState{{T: 0, Held: []string{}}, {T: 110, Bound: "C1", Held: []string{}}, {T: 500, Held: []string{"C1"}}}
	bound := []bindState{{T: 0, Held: []string{}}, {T: 110, Bound: "C1", Held: []string{}}}

	// When
	one := poolWaitCycles(tree, slots, binding, 1)
	two := poolWaitCycles(tree, slots, binding, 2)
	notHeld := poolWaitCycles(tree, slots, bound, 1)

	// Then
	want := []waitCycle{{Kind: "pool", Pool: 1, T0: 600, T1: 2600, Holder: "C1", Thread: sceneThread, Call: 2, HeldBy: sceneThread, Cycle: true}}
	if !reflect.DeepEqual(one, want) {
		t.Fatalf("풀 1칸: %+v; want %+v", one, want)
	}
	if len(two) != 1 || two[0].Cycle {
		t.Fatalf("풀 2칸에 1칸만 쓰임: %+v; want Cycle false", two)
	}
	if len(notHeld) != 1 || notHeld[0].Cycle || notHeld[0].HeldBy != "" {
		t.Fatalf("C1을 떼어 두지 않음(묶인 채): %+v; want Cycle false, HeldBy 빈 글", notHeld)
	}
}
