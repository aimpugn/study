package main

import (
	"reflect"
	"testing"
)

// usRel 은 시험에서 쓰는 시각 변환입니다. 기준 0부터의 마이크로초입니다.
func usRel(ns uint64) float64 { return float64(ns) / 1e3 }

// TestSplitRequestsPairsPipelinedBatches 는 앱이 답을 기다리지 않고 BEGIN 단순 질의와 첫 문장의 확장 질의를 한 번에 보내도, 서버의
// ReadyForQuery 둘이 그 두 묶음에 하나씩 짝지어지는지 봅니다. 메시지 순서는 2026-10-15 r5-nested-a의 C1(pgjdbc 42.7.13)과 같습니다.
// 처음 쓴 해석은 Z가 올 때만 묶음을 닫아, 두 요청을 하나로 합치고 같은 요청을 두 번 냈습니다.
func TestSplitRequestsPairsPipelinedBatches(t *testing.T) {
	// Given
	mac := []wireMsg{
		{Dir: "F", Type: "ssl", Ns: 1000}, {Dir: "B", Type: "sslreply", Ns: 1100}, {Dir: "F", Type: "startup", Ns: 1200},
		{Dir: "B", Type: "R", Ns: 1300}, {Dir: "B", Type: "K", Ns: 1300}, {Dir: "B", Type: "Z", Status: "I", Ns: 1300},
		{Dir: "F", Type: "Q", SQL: "BEGIN", Ns: 5000},
		{Dir: "F", Type: "P", SQL: "UPDATE acct SET balance = balance - 10 WHERE id = 1", Ns: 5000},
		{Dir: "F", Type: "B", Ns: 5000}, {Dir: "F", Type: "D", Ns: 5000}, {Dir: "F", Type: "E", Ns: 5000}, {Dir: "F", Type: "S", Ns: 5000},
		{Dir: "B", Type: "C", Tag: "BEGIN", Ns: 6000}, {Dir: "B", Type: "Z", Status: "T", Ns: 6000},
		{Dir: "B", Type: "1", Ns: 9000}, {Dir: "B", Type: "2", Ns: 9000}, {Dir: "B", Type: "C", Tag: "UPDATE 1", Ns: 9000},
		{Dir: "B", Type: "Z", Status: "T", Ns: 9000},
		{Dir: "F", Type: "X", Ns: 20000},
	}
	cont := make([]wireMsg, len(mac))
	for i, m := range mac {
		m.Ns += 50
		cont[i] = m
	}

	// When
	reqs, zc := splitRequests(mac, cont, usRel)

	// Then
	if len(reqs) != 3 || len(zc) != 3 {
		t.Fatalf("요청 %d개, Z %d개; want 3, 3: %+v", len(reqs), len(zc), reqs)
	}
	if got := reqs[1].SQL; !reflect.DeepEqual(got, []string{"BEGIN"}) || reqs[1].Status != "T" || !reflect.DeepEqual(reqs[1].Tags, []string{"BEGIN"}) {
		t.Fatalf("요청 1 = %+v; want BEGIN 하나, 상태 T, 태그 BEGIN", reqs[1])
	}
	if got := reqs[2].SQL; len(got) != 1 || reqs[2].Sent != 5 || reqs[2].Done != 9 || !reflect.DeepEqual(reqs[2].Tags, []string{"UPDATE 1"}) {
		t.Fatalf("요청 2 = %+v; want UPDATE 하나, 보낸 때 5 µs, 끝 9 µs, 태그 UPDATE 1", reqs[2])
	}
	if zc[2] != 9050 {
		t.Fatalf("요청 2의 컨테이너 쪽 Z = %d; want 9050", zc[2])
	}
}

// TestSplitRequestsDropsUnansweredBatch 는 답이 오지 않은 묶음(끝내기 앞에 보냈지만 Z가 없음)을 요청으로 내지 않는지 봅니다.
func TestSplitRequestsDropsUnansweredBatch(t *testing.T) {
	// Given
	mac := []wireMsg{
		{Dir: "F", Type: "startup", Ns: 1}, {Dir: "B", Type: "Z", Status: "I", Ns: 2},
		{Dir: "F", Type: "Q", SQL: "SELECT 1", Ns: 3},
		{Dir: "F", Type: "X", Ns: 4},
	}

	// When
	reqs, zc := splitRequests(mac, mac, usRel)

	// Then
	if len(reqs) != 1 || len(zc) != 1 || reqs[0].Seq != 0 {
		t.Fatalf("요청 %+v; want 접속 인증 하나", reqs)
	}
}

// TestDeriveTxsFindsTransactionAndNextMessage 는 ReadyForQuery 상태가 I → T → I로 돈 구간을 트랜잭션 하나로 묶고, 끝난 뒤 그 연결의
// 다음 앱 메시지를 붙이는지 봅니다. 실패한 트랜잭션(E)도 같은 구간에 듭니다.
func TestDeriveTxsFindsTransactionAndNextMessage(t *testing.T) {
	// Given
	reqs := []springRequest{
		{Seq: 0, Status: "I"},
		{Seq: 1, Status: "I", Sent: 10, Done: 11},
		{Seq: 2, Status: "T", Sent: 20, Done: 21},
		{Seq: 3, Status: "E", Sent: 30, Done: 31, Tags: nil},
		{Seq: 4, Status: "I", Sent: 40, Done: 41, Tags: []string{"ROLLBACK"}},
	}
	msgs := []wireView{{T: 40, Dir: "F", Type: "P"}, {T: 41, Dir: "B", Type: "Z"}, {T: 90, Dir: "F", Type: "X"}}

	// When
	txs := deriveTxs(1, reqs, msgs)

	// Then
	if len(txs) != 1 {
		t.Fatalf("트랜잭션 %d개; want 1", len(txs))
	}
	tx := txs[0]
	if tx.First != 2 || tx.Last != 4 || tx.End != "ROLLBACK" || tx.T0 != 20 || tx.T1 != 41 || tx.Next == nil || tx.Next.Type != "X" {
		t.Fatalf("트랜잭션 = %+v; want 요청 2~4, 끝 ROLLBACK, 다음 X", tx)
	}
}

// TestCompareClocksBoundsOffset 은 시계 차이 구간이 서버 → 앱 메시지의 최댓값과 앱 → 서버 메시지의 최솟값으로 정해지고, 두 기록의
// 자르기가 다르면 같지 않다고 보는지 봅니다.
func TestCompareClocksBoundsOffset(t *testing.T) {
	// Given: 컨테이너 시계가 맥보다 0.1 ms 앞서고, 앱 → 서버 지연은 0.05 ms, 서버 → 앱 지연은 0.03 ms다.
	mac := []wireMsg{{Dir: "F", N: 1, Type: "Q", Len: 9, Ns: 1_000_000}, {Dir: "B", N: 1, Type: "Z", Len: 5, Ns: 2_000_000}}
	cont := []wireMsg{{Dir: "F", N: 1, Type: "Q", Len: 9, Ns: 1_150_000}, {Dir: "B", N: 1, Type: "Z", Len: 5, Ns: 2_070_000}}

	// When
	c := compareClocks(1, mac, cont)
	cont[1].Len = 6
	bad := compareClocks(1, mac, cont)

	// Then
	if !c.Same || c.Lo != 0.07 || c.Hi != 0.15 {
		t.Fatalf("구간 = %+v; want same, [0.07, 0.15] ms", c)
	}
	if bad.Same {
		t.Fatalf("길이가 다른데 같다고 봄: %+v", bad)
	}
}

// TestVisibleRowsUsesCommitStatus 는 새 트랜잭션이 볼 행 판을 xmin의 커밋과 xmax의 비커밋으로 고르는지 봅니다. 진행 중인 트랜잭션이
// 고친 행은 옛 판이 보이고, 되돌린 트랜잭션이 만든 판은 보이지 않습니다.
func TestVisibleRowsUsesCommitStatus(t *testing.T) {
	// Given: 752가 만든 두 행을 753(진행 중)과 754(되돌림)가 고쳤다.
	vs := []txVersion{
		{Ctid: "(0,1)", Xmin: "752", Xmax: "753", Values: map[string]string{"id": "1", "balance": "100"}},
		{Ctid: "(0,2)", Xmin: "752", Xmax: "754", Values: map[string]string{"id": "2", "balance": "200"}},
		{Ctid: "(0,3)", Xmin: "753", Xmax: "0", Values: map[string]string{"id": "1", "balance": "90"}},
		{Ctid: "(0,4)", Xmin: "754", Xmax: "0", Values: map[string]string{"id": "2", "balance": "210"}},
	}
	xacts := map[string]string{"752": "committed", "753": "in progress", "754": "aborted"}

	// When
	rows := visibleRows(vs, xacts)

	// Then
	var got []string
	for _, r := range rows {
		got = append(got, r["ctid"]+"="+r["balance"])
	}
	if !reflect.DeepEqual(got, []string{"(0,1)=100", "(0,2)=200"}) {
		t.Fatalf("보이는 판 = %v; want 옛 판 둘", got)
	}
}

// TestCaptureSpansIgnoresOtherQueries 는 관찰 구간의 끝이 그 관찰의 마지막 부분 질의에 답한 때이고, 그 뒤 관찰이 아닌 질의(멈춘 채로
// 찍기 전에 묻는 waiters)는 구간을 늘리지 않는지 봅니다. 처음 쓴 해석은 다음 관찰의 시작까지를 구간으로 잡아, 앱이 멈춰 있던 멈춤 자리
// 관찰에도 다음 단계의 메시지가 겹친 것으로 냈습니다.
func TestCaptureSpansIgnoresOtherQueries(t *testing.T) {
	// Given
	q := func(sql string) []byte {
		b := []byte{'Q', 0, 0, 0, byte(len(sql) + 5)}
		return append(append(b, sql...), 0)
	}
	obs := []sysEvent{
		{Name: "recvfrom", Ret: 10, Ns: 110, Data: q("SELECT heap_page_item_attrs(1)")},
		{Name: "sendto", Ret: 10, Ns: 120},
		{Name: "recvfrom", Ret: 10, Ns: 130, Data: q("SELECT pg_xact_status(x)")},
		{Name: "sendto", Ret: 10, Ns: 140},
		{Name: "recvfrom", Ret: 10, Ns: 500, Data: q("SELECT pg_blocking_pids(1)")},
		{Name: "sendto", Ret: 10, Ns: 510},
		{Name: "recvfrom", Ret: 10, Ns: 610, Data: q("SELECT heap_page_item_attrs(1)")},
		{Name: "sendto", Ret: 10, Ns: 620},
	}

	// When
	spans := captureSpans([]uint64{100, 600}, obs)

	// Then
	if !reflect.DeepEqual(spans, [][2]uint64{{100, 140}, {600, 620}}) {
		t.Fatalf("구간 = %v; want [100 140] [600 620]", spans)
	}
}

// TestParseISODurationReadsJFRLengths 는 JFR JSON의 길이 꼴을 나노초로 읽고 다른 꼴은 오류로 보는지 봅니다.
func TestParseISODurationReadsJFRLengths(t *testing.T) {
	for in, want := range map[string]int64{"PT0.002433417S": 2433417, "PT1M2.5S": 62_500_000_000, "PT0S": 0, "PT1H": 3600_000_000_000} {
		if got, err := parseISODuration(in); err != nil || got != want {
			t.Errorf("parseISODuration(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"0.5S", "PT0.5", "PTxS"} {
		if _, err := parseISODuration(in); err == nil {
			t.Errorf("parseISODuration(%q): 오류가 없음", in)
		}
	}
}

// jfrSpanEvent 는 시험용 메서드 추적 사건입니다. callers는 부른 쪽부터 바깥으로의 칸입니다.
func jfrSpanEvent(cls, name, desc, thread, start, dur string, callers ...jfrMethod) jfrEvent {
	var e jfrEvent
	e.Type = "jdk.MethodTrace"
	e.Values.StartTime, e.Values.Duration = start, dur
	e.Values.EventThread.JavaName = thread
	m := jfrMethod{Name: name, Descriptor: desc}
	m.Type.Name = cls
	e.Values.Method = &m
	var frames []jfrFrame
	for _, c := range callers {
		frames = append(frames, jfrFrame{Method: c, Line: 1})
	}
	e.Values.StackTrace = &jfrStack{Frames: frames}
	return e
}

// jm 은 시험용 메서드입니다.
func jm(cls, name, desc string) jfrMethod {
	m := jfrMethod{Name: name, Descriptor: desc}
	m.Type.Name = cls
	return m
}

// TestBuildSpringJavaMergesTwinsAndFindsStackOnlyCalls 는 같은 메서드·서술자·스레드·시작 시각의 사건 둘을 긴 쪽 하나로 합치고, 이름이
// 같은 오버로드는 합치지 않으며, 추적 대상인데 사건 없이 스택에만 보인 호출을 세는지 봅니다. 모양은 2026-10-15 r5-required-a의
// processRollback(finally가 받은 throw)과 HikariPool.getConnection 두 오버로드, 사건이 없는 TransactionInterceptor.invoke입니다.
func TestBuildSpringJavaMergesTwinsAndFindsStackOnlyCalls(t *testing.T) {
	// Given
	invoke := jm("org/springframework/transaction/interceptor/TransactionInterceptor", "invoke", "(Lorg/aopalliance/intercept/MethodInvocation;)Ljava/lang/Object;")
	apt := "org/springframework/transaction/support/AbstractPlatformTransactionManager"
	events := []jfrEvent{
		jfrSpanEvent(apt, "processRollback", "(Lx;Z)V", "main", "2026-10-15T09:00:00.000001000Z", "PT0.000006S", invoke),
		jfrSpanEvent(apt, "processRollback", "(Lx;Z)V", "main", "2026-10-15T09:00:00.000001000Z", "PT0.000009S", invoke),
		jfrSpanEvent("com/zaxxer/hikari/pool/HikariPool", "getConnection", "()Ljava/sql/Connection;", "main", "2026-10-15T09:00:00.000020000Z", "PT0.000010S", invoke),
		jfrSpanEvent("com/zaxxer/hikari/pool/HikariPool", "getConnection", "(J)Ljava/sql/Connection;", "main", "2026-10-15T09:00:00.000020500Z", "PT0.000009S", invoke),
	}
	base := uint64(1792054800000000000) // 2026-10-15T09:00:00Z
	rel := func(ns uint64) float64 { return float64(int64(ns-base)) / 1e3 }
	filter := "org.springframework.transaction.interceptor.TransactionInterceptor::invoke;org.springframework.transaction.support.AbstractPlatformTransactionManager;com.zaxxer.hikari.pool.HikariPool::getConnection"

	// When
	j, err := buildSpringJava(events, filter, rel, 0)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if len(j.spans) != 3 || len(j.dups) != 1 || j.dups[0].Fn != "org.springframework.transaction.support.AbstractPlatformTransactionManager::processRollback" {
		t.Fatalf("구간 %d개, 합친 묶음 %+v; want 구간 셋, processRollback 묶음 하나", len(j.spans), j.dups)
	}
	if s := j.spans[0]; s.T0 != 1 || s.T1 != 10 || !s.Twice {
		t.Fatalf("processRollback = %+v; want 1~10 µs, 합침 표시", s)
	}
	if inner := j.spans[2]; inner.Depth != 1 || inner.Desc != "(J)Ljava/sql/Connection;" {
		t.Fatalf("getConnection(long) = %+v; want 깊이 1", inner)
	}
	only := j.stackOnly()
	if len(only) != 1 || only[0].Fn != "org.springframework.transaction.interceptor.TransactionInterceptor::invoke" || only[0].Events != 3 {
		t.Fatalf("스택에만 보인 호출 = %+v; want invoke 하나, 사건 셋", only)
	}
}

// TestStackOnlyCountsNestedMissingCall 은 같은 메서드가 겹쳐 불렸을 때 바깥 호출만 기록되고 안쪽 호출은 예외가 지나가며 끝나 기록이
// 없으면, 바깥 기록이 감싸더라도 안쪽 호출을 스택에만 보인 호출로 세는지 봅니다. 모양은 2026-10-15 r5-nested-a입니다(바깥 프록시의
// TransactionInterceptor.invoke는 정상으로 끝나 기록되고, 안쪽 프록시의 invoke는 IllegalStateException이 지나가며 끝남).
func TestStackOnlyCountsNestedMissingCall(t *testing.T) {
	// Given: 바깥 invoke(0~100 µs)가 기록됐고, 그 안에서 안쪽 invoke를 거쳐 불린 JdbcTemplate.update가 기록됐다.
	invoke := jm("org/springframework/transaction/interceptor/TransactionInterceptor", "invoke", "(Lm;)Ljava/lang/Object;")
	outer := jm("lab/Outer", "nested", "()V")
	inner := jm("lab/Inner", "nestedFails", "()V")
	events := []jfrEvent{
		jfrSpanEvent("org/springframework/transaction/interceptor/TransactionInterceptor", "invoke", "(Lm;)Ljava/lang/Object;", "main",
			"2026-10-15T09:00:00.000000000Z", "PT0.000100S"),
		jfrSpanEvent("org/springframework/jdbc/core/JdbcTemplate", "update", "(Ljava/lang/String;)I", "main",
			"2026-10-15T09:00:00.000050000Z", "PT0.000010S", inner, invoke, outer, invoke),
	}
	base := uint64(1792054800000000000)
	rel := func(ns uint64) float64 { return float64(int64(ns-base)) / 1e3 }
	filter := "org.springframework.transaction.interceptor.TransactionInterceptor::invoke;org.springframework.jdbc.core.JdbcTemplate::update"

	// When
	j, err := buildSpringJava(events, filter, rel, 0)
	if err != nil {
		t.Fatal(err)
	}
	only := j.stackOnly()

	// Then
	if len(only) != 1 || only[0].Fn != "org.springframework.transaction.interceptor.TransactionInterceptor::invoke" || only[0].Events != 1 {
		t.Fatalf("스택에만 보인 호출 = %+v; want 안쪽 invoke 하나", only)
	}
}
