package run

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"study/open_source_internals/workbench/internal/subject"
)

// 세션 형식을 깨는 입력을 거부하고, 정상 SQL과 형 변환 '::', 탭, 따옴표 안의 특수 문자는 받는지 확인합니다.
func TestValidateSQL(t *testing.T) {
	ok := []string{"SELECT * FROM t WHERE id = 2;", "SELECT id::text FROM t;", "SELECT\t1;", "SELECT '가';  ",
		"SELECT 'it''s';", "SELECT (1 + (2));", `SELECT 1 AS "a(b";`, "SELECT 'a--b /* $ (';"}
	bad := map[string]string{
		"빈 입력":       "",
		"공백뿐":        "   \t ",
		"백슬래시 메타 명령": `SELECT 1; \! id`,
		"줄바꿈":        "SELECT 1;\n\\! id",
		"캐리지 리턴":     "SELECT 1\r",
		"NUL":        "SELECT 1\x00",
		"유니코드 줄 구분자": "SELECT 1 ",
		"두 문장":       "SELECT 1; SELECT 2;",
		"따옴표 안의 ;":   "SELECT ';';",
		"너무 김":       "SELECT '" + strings.Repeat("a", MaxSQLBytes) + "';",
		"UTF-8 아님":   "SELECT '\xff';",
		"끝에 ; 없음(다음 줄과 이어짐)": "SELECT * FROM t WHERE id = 2",
		"닫히지 않은 작은따옴표":       "SELECT 'a;",
		"닫히지 않은 큰따옴표":        `SELECT 1 AS "x;`,
		"닫히지 않은 괄호":          "SELECT (1;",
		"짝 없는 닫는 괄호":         "SELECT 1);",
		"줄 주석":               "SELECT 1 -- x;",
		"블록 주석":              "SELECT 1 /* x */;",
		"달러 따옴표":             "SELECT $$a;",
		"COPY FROM STDIN":    "COPY t FROM STDIN;",
	}
	for _, s := range ok {
		if err := ValidateSQL(s); err != nil {
			t.Errorf("받아야 할 SQL을 거부했습니다: %q: %v", s, err)
		}
	}
	for name, s := range bad {
		var ie *InputError
		if err := ValidateSQL(s); !errors.As(err, &ie) {
			t.Errorf("%s: 거부해야 할 SQL을 받았습니다: %q", name, s)
		}
	}
}

// fakeDocker 값은 시험용 명령 실행기입니다. 실행한 명령을 기록하고, 추적 스크립트에는 정해 둔 출력을 돌려줍니다.
type fakeDocker struct {
	mu         sync.Mutex
	calls      []string
	running    bool
	inProgress bool
	trace      string
	traceErr   error
	block      chan struct{}
}

// exec 함수는 docker inspect에는 켜짐 여부를, 추적 여부 확인에는 inProgress를, sh(추적 스크립트)에는 정해 둔 기록을 돌려줍니다.
func (f *fakeDocker) exec(ctx context.Context, env []string, name string, args ...string) ([]byte, []byte, error) {
	f.mu.Lock()
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	f.mu.Unlock()
	switch {
	case name == "docker" && args[0] == "inspect":
		if f.running {
			return []byte("true\n"), nil, nil
		}
		return []byte("false\n"), nil, nil
	case name == "docker" && strings.Contains(strings.Join(args, " "), "echo busy"):
		if f.inProgress {
			return []byte("busy\n"), nil, nil
		}
		return []byte("free\n"), nil, nil
	case name == "sh":
		if f.block != nil {
			select {
			case <-f.block:
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
		}
		return []byte(f.trace), []byte("stderr 내용"), f.traceErr
	}
	return nil, nil, nil
}

// cleanups 함수는 남은 프로세스를 정리한 docker exec 호출 횟수를 셉니다.
func (f *fakeDocker) cleanups() int { return f.count("pkill -INT -x gdb") }

// scripts 함수는 추적 스크립트를 부른 횟수를 셉니다.
func (f *fakeDocker) scripts() int { return f.count("sh ") }

// count 함수는 기록한 호출 가운데 s를 담은 것의 수를 셉니다. "sh "는 명령 이름이 sh인 호출만 셉니다.
func (f *fakeDocker) count(s string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if (s == "sh " && strings.HasPrefix(c, s)) || (s != "sh " && strings.Contains(c, s)) {
			n++
		}
	}
	return n
}

// newRunner 함수는 실제 PostgreSQL 주제 정의와 가짜 실행기로 Runner를 만듭니다.
func newRunner(t *testing.T, f *fakeDocker) *Runner {
	t.Helper()
	s, err := subject.Load("../../../database/postgresql/workbench.json")
	if err != nil {
		t.Fatal(err)
	}
	return &Runner{Subject: s, Exec: f.exec, Timeout: 2 * time.Second}
}

// sampleTrace는 run-trace.sh 출력의 최소 꼴입니다. 첫 단계의 query_string이 "SELECT 1;"입니다.
const sampleTrace = `STOP 1 exec_simple_query src/backend/tcop/postgres.c:1014 query_string=0xaaaa "SELECT 1;"
STOP 2 ReadyForQuery src/backend/tcop/dest.c:258
END ReadyForQuery=1 exec_simple_query=1
--- psql
1
1
`

// 정상 실행에서 기록을 읽고, 정리를 하지 않으며, 스크립트에 SQL을 인자 하나로 넘기는지 확인합니다.
func TestRunParsesRecordWithoutCleanup(t *testing.T) {
	// Given
	f := &fakeDocker{running: true, trace: sampleTrace}
	r := newRunner(t, f)

	// When
	res, err := r.Run(context.Background(), "select", "SELECT 1;")
	// Then
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Record.Steps) != 2 || len(res.Warnings) != 0 {
		t.Fatalf("기록 %d단계, 주의 %v; 2단계와 주의 없음이어야 합니다", len(res.Record.Steps), res.Warnings)
	}
	if f.cleanups() != 0 {
		t.Fatalf("성공한 실행에서는 정리하지 않아야 합니다: %d번", f.cleanups())
	}
	last := f.calls[len(f.calls)-1]
	if !strings.HasSuffix(last, "select.watch SELECT 1; PinBufferForBlock") {
		t.Fatalf("스크립트 인자가 다릅니다: %s", last)
	}
}

// PostgreSQL이 받은 SQL이 입력과 다르면(psql 변수 치환 등) 주의 문구를 붙이는지 확인합니다.
func TestRunWarnsWhenQueryDiffers(t *testing.T) {
	// Given: 입력은 :USER였지만 기록의 query_string은 postgres로 바뀐 경우
	f := &fakeDocker{running: true, trace: strings.Replace(sampleTrace, `"SELECT 1;"`, `"SELECT postgres;"`, 1)}
	r := newRunner(t, f)

	// When
	res, err := r.Run(context.Background(), "select", "SELECT :USER;")
	// Then
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], "입력과 다릅니다") {
		t.Fatalf("SQL이 바뀌었다는 주의가 있어야 합니다: %v", res.Warnings)
	}
}

// 컨테이너가 꺼져 있으면 스크립트를 부르지 않고 ErrNotRunning을 돌려주는지 확인합니다.
func TestRunRequiresRunningContainer(t *testing.T) {
	// Given
	f := &fakeDocker{running: false}
	r := newRunner(t, f)

	// When
	_, err := r.Run(context.Background(), "select", "SELECT 1;")

	// Then
	if !errors.Is(err, ErrNotRunning) {
		t.Fatalf("ErrNotRunning이어야 합니다: %v", err)
	}
	for _, c := range f.calls {
		if strings.HasPrefix(c, "sh ") {
			t.Fatal("컨테이너가 꺼져 있는데 스크립트를 불렀습니다")
		}
	}
}

// 컨테이너 안에 이 서버 밖의 추적이 돌고 있으면, 그 추적을 정리하지 않고 ErrTraceInProgress로 거부하는지 확인합니다.
func TestRunRefusesWithoutTouchingTraceInProgress(t *testing.T) {
	// Given: 터미널에서 직접 실행한 run-trace.sh가 돌고 있는 컨테이너
	f := &fakeDocker{running: true, inProgress: true, trace: sampleTrace}
	r := newRunner(t, f)

	// When
	_, err := r.Run(context.Background(), "select", "SELECT 1;")

	// Then
	if !errors.Is(err, ErrTraceInProgress) || f.scripts() != 0 || f.cleanups() != 0 {
		t.Fatalf("스크립트도 정리도 부르지 않고 ErrTraceInProgress여야 합니다: %v, 스크립트 %d번, 정리 %d번", err, f.scripts(), f.cleanups())
	}
}

// 실행 중에 온 두 번째 요청은 ErrBusy로 거부하고, 첫 실행이 끝나면 다시 받는지 확인합니다.
func TestRunRejectsConcurrentRunAndReleasesLock(t *testing.T) {
	// Given: 첫 실행이 block이 닫힐 때까지 끝나지 않음
	f := &fakeDocker{running: true, trace: sampleTrace, block: make(chan struct{})}
	r := newRunner(t, f)
	done := make(chan error, 1)
	go func() {
		_, err := r.Run(context.Background(), "select", "SELECT 1;")
		done <- err
	}()
	waitFor(t, func() bool { return f.scripts() == 1 })

	// When
	_, second := r.Run(context.Background(), "select", "SELECT 1;")
	close(f.block)
	first := <-done
	_, third := r.Run(context.Background(), "select", "SELECT 1;")

	// Then
	if !errors.Is(second, ErrBusy) || first != nil || third != nil {
		t.Fatalf("두 번째는 ErrBusy, 첫째와 셋째는 성공이어야 합니다: %v, %v, %v", second, first, third)
	}
}

// 시간 한도를 넘으면 ErrTimeout을 돌려주고, 남은 프로세스를 정리하고, 다음 실행을 받는지 확인합니다(B9).
func TestRunTimeoutCleansUpAndAllowsNextRun(t *testing.T) {
	// Given: 스크립트가 끝나지 않고, 시간 한도는 짧음
	f := &fakeDocker{running: true, trace: sampleTrace, block: make(chan struct{})}
	r := newRunner(t, f)
	r.Timeout = 50 * time.Millisecond

	// When
	_, err := r.Run(context.Background(), "select", "SELECT 1;")
	f.block = nil
	_, next := r.Run(context.Background(), "select", "SELECT 1;")

	// Then: 정리는 시간 초과 뒤 한 번뿐입니다
	if !errors.Is(err, ErrTimeout) || next != nil {
		t.Fatalf("시간 초과 뒤 다음 실행이 성공해야 합니다: %v, %v", err, next)
	}
	if f.cleanups() != 1 {
		t.Fatalf("정리는 한 번이어야 합니다: %d번", f.cleanups())
	}
}

// 스크립트가 실패하면 표준 오류를 담은 오류를 돌려주고 정리하는지 확인합니다.
func TestRunScriptFailureReportsStderr(t *testing.T) {
	// Given
	f := &fakeDocker{running: true, traceErr: errors.New("exit status 1")}
	r := newRunner(t, f)

	// When
	_, err := r.Run(context.Background(), "select", "SELECT 1;")

	// Then
	if err == nil || !strings.Contains(err.Error(), "stderr 내용") || f.cleanups() != 1 {
		t.Fatalf("표준 오류를 담은 오류와 한 번의 정리가 있어야 합니다: %v, %d", err, f.cleanups())
	}
}

// 최소 환경에 학습자 입력이 들어갈 자리가 없고 PG_OPTIONS가 고정값인지 확인합니다.
func TestMinimalEnvFixesPGOptions(t *testing.T) {
	// Given: 서버 환경에 다른 PG_OPTIONS가 있어도
	t.Setenv("PG_OPTIONS", "-c synchronous_commit=off")
	t.Setenv("SECRET_TOKEN", "x")

	// When
	env := MinimalEnv("")

	// Then
	joined := strings.Join(env, "\n")
	if strings.Count(joined, "PG_OPTIONS=") != 1 || !strings.Contains(joined, "PG_OPTIONS=-c statement_timeout=30s") || strings.Contains(joined, "SECRET_TOKEN") {
		t.Fatalf("환경이 다릅니다: %v", env)
	}
}

// waitFor 함수는 cond가 참이 될 때까지 최대 2초 기다리고, 넘으면 시험을 멈춥니다.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("조건이 2초 안에 참이 되지 않았습니다")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// 시간 한도가 지나면, 끝난 셸의 자식이 출력 파이프를 붙잡고 있어도 곧 돌아오는지 실제 프로세스로 확인합니다.
//
// run-trace.sh는 sh가 docker exec와 grep의 파이프라인을 자식으로 띄웁니다. 시간 한도에 sh만 끝나면 자식이 표준 출력을
// 계속 열어 두므로, 출력이 닫히기를 기다리는 실행기는 자식이 스스로 끝날 때까지 돌아오지 못합니다.
func TestRealExecReturnsSoonAfterTimeoutWhenChildHoldsOutput(t *testing.T) {
	// Given: 자식 sleep이 10초 동안 표준 출력을 붙잡는 셸
	r := &Runner{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// When
	start := time.Now()
	_, _, err := r.exec(ctx, "sh", "-c", "sleep 10 | cat")
	elapsed := time.Since(start)

	// Then
	if err == nil {
		t.Fatal("시간 한도를 넘었는데 오류가 없습니다")
	}
	// 한도는 시간 한도 0.2초 + 출력 파이프를 닫기 전 기다림 2초 + 여유 1초입니다. 자식이 끝나는 10초보다 훨씬 짧습니다.
	if elapsed > 3200*time.Millisecond {
		t.Fatalf("시간 한도 0.2초를 넘긴 뒤 %.1f초 만에 돌아왔습니다. 자식이 출력을 붙잡는 동안 기다렸습니다", elapsed.Seconds())
	}
}

// 오류 문구를 자를 때 한글 글자의 중간에서 자르지 않는지 확인합니다.
func TestTailKeepsValidUTF8(t *testing.T) {
	// Given: "가"는 3바이트이므로 끝 4바이트는 "가"의 중간부터 시작합니다
	b := []byte("가가")

	// When
	got := tail(b, 4)

	// Then
	if !utf8.ValidString(got) || got != "가" {
		t.Fatalf("끝 4바이트 이내의 온전한 글자는 %q여야 합니다: %q", "가", got)
	}
}

// 감시 목록이 정한 서버 설정이 스크립트의 PG_OPTIONS에 시간 한도와 함께 들어가고, 다른 명령에는 들어가지 않는지 확인합니다.
func TestRunPassesWatchPGOptionsOnlyToScript(t *testing.T) {
	// Given: 순차 스캔을 끄는 감시 목록을 가진 주제
	var mu sync.Mutex
	envs := map[string]string{}
	f := &fakeDocker{running: true, trace: sampleTrace}
	r := newRunner(t, f)
	r.Subject.Watches = append(r.Subject.Watches, subject.Watch{ID: "noseq", File: "scripts/select.watch", PGOptions: "-c enable_seqscan=off"})
	r.Exec = func(ctx context.Context, env []string, name string, args ...string) ([]byte, []byte, error) {
		mu.Lock()
		for _, e := range env {
			if v, ok := strings.CutPrefix(e, "PG_OPTIONS="); ok {
				envs[name+" "+args[0]] = v
			}
		}
		mu.Unlock()
		return f.exec(ctx, env, name, args...)
	}

	// When
	_, err := r.Run(context.Background(), "noseq", "SELECT 1;")

	// Then
	if err != nil {
		t.Fatal(err)
	}
	var script, inspect string
	for k, v := range envs {
		switch {
		case strings.HasPrefix(k, "sh "):
			script = v
		case k == "docker inspect":
			inspect = v
		}
	}
	if script != "-c statement_timeout=30s -c enable_seqscan=off" || inspect != "-c statement_timeout=30s" {
		t.Fatalf("PG_OPTIONS가 다릅니다: 스크립트 %q, docker inspect %q", script, inspect)
	}
}
