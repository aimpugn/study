// Package run 은 화면에서 받은 SQL 한 문장을 주제의 추적 스크립트(run-trace.sh)로 실행하고 기록을 돌려줍니다.
//
// 세션 준비와 gdb 부착의 정본은 주제 폴더의 스크립트입니다. 이 패키지는 그 스크립트를 부르는 얇은 경계로서
// 입력 형식 검사, 한 번에 한 실행, 남은 프로세스 정리, 시간 한도, 최소 환경만 맡습니다.
//
// 보안 경계는 이 패키지가 아니라 컨테이너 격리(--network none, 읽기 전용 마운트)와 서버의 요청 출처 검사입니다.
// 실습의 psql 사용자는 슈퍼유저라서 SQL만으로도 컨테이너 안에서 명령을 실행할 수 있기 때문입니다(설계 기록의 안전 경계 절).
package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"study/open_source_internals/workbench/internal/subject"
	"study/open_source_internals/workbench/internal/trace"
)

// MaxSQLBytes 값은 받아들일 SQL의 최대 바이트 수입니다.
const MaxSQLBytes = 1000

// ErrBusy 값은 다른 실행이 진행 중이라 이번 요청을 받지 않았다는 뜻입니다.
var ErrBusy = errors.New("다른 실행이 진행 중입니다. 끝난 뒤 다시 실행하세요")

// ErrNotRunning 값은 실습 컨테이너가 켜져 있지 않다는 뜻입니다.
var ErrNotRunning = errors.New("실습 컨테이너가 켜져 있지 않습니다")

// ErrTimeout 값은 실행이 시간 한도 안에 끝나지 않았다는 뜻입니다.
var ErrTimeout = errors.New("추적이 시간 한도 안에 끝나지 않았습니다")

// ErrTraceInProgress 값은 실습 컨테이너 안에서 이 서버 밖의 추적(직접 실행한 run-trace.sh 등)이 진행 중이라는 뜻입니다.
var ErrTraceInProgress = errors.New("실습 컨테이너 안에서 다른 추적(gdb나 세션 psql)이 진행 중입니다. " +
	"터미널에서 직접 실행한 run-trace.sh가 있다면 끝난 뒤 다시 실행하세요")

// InputError 값은 요청 자체가 실행할 수 없는 꼴이라는 뜻입니다. 서버는 이것을 400으로 돌려줍니다.
type InputError struct{ Reason string }

// Error 함수는 거부한 이유를 돌려줍니다.
func (e *InputError) Error() string { return e.Reason }

// ValidateSQL 함수는 sql이 세션 스크립트의 한 줄로 들어가 psql이 정확히 한 문장으로 보내는 꼴인지 확인합니다.
//
// 세션 스크립트는 SQL 다음 줄에 psql 메타 명령(\!)을 쓰므로, psql이 이 줄의 끝 ';'를 문장 끝으로 보지 못하면
// 다음 줄들이 SQL에 이어 붙거나 자료로 읽혀 "두 번 실행"이라는 세션 순서가 깨집니다. 실제로 끝의 ';'를 빼자
// query_string이 같은 SQL 두 줄을 이은 한 문장이 되어 구문 오류로 끝났습니다. 그래서 아래를 거부합니다.
//
//   - 빈 입력과 공백뿐인 입력, MaxSQLBytes 초과, UTF-8이 아닌 바이트
//   - 백슬래시: psql 메타 명령을 막기 위해 따옴표 안이든 밖이든 모두 거부하는 단순화입니다.
//   - 줄바꿈과 제어 문자(탭 제외, 유니코드 줄 구분자 포함)
//   - 끝에 ';'가 없거나, 끝이 아닌 자리에 ';'가 있는 것(따옴표 안도 포함하는 의도된 제한)
//   - 닫히지 않은 작은따옴표와 큰따옴표, 짝이 맞지 않는 괄호, 주석(--, /*), 달러 따옴표($): psql은 이것들 안의 ';'를
//     문장 끝으로 보지 않습니다.
//   - COPY ... FROM STDIN: psql이 다음 줄들을 복사할 자료로 읽습니다. 따옴표 안의 stdin도 거부합니다.
//
// 이 검사는 보안 장치가 아니라 세션 형식을 지키는 장치이고, SQL 문법은 해석하지 않습니다.
func ValidateSQL(sql string) error {
	trimmed := strings.TrimSpace(sql)
	switch {
	case trimmed == "":
		return &InputError{"SQL이 비어 있습니다"}
	case len(sql) > MaxSQLBytes:
		return &InputError{fmt.Sprintf("SQL은 %d바이트 이하여야 합니다(%d바이트)", MaxSQLBytes, len(sql))}
	case !utf8.ValidString(sql):
		return &InputError{"SQL에 UTF-8이 아닌 바이트가 있습니다"}
	case strings.ContainsRune(sql, '\\'):
		return &InputError{"SQL에 백슬래시를 쓸 수 없습니다. 작업대는 psql 메타 명령을 막으려고 백슬래시를 모두 거부합니다"}
	}
	for _, r := range sql {
		if (r < 0x20 && r != '\t') || r == 0x7f || r == 0x85 || r == 0x2028 || r == 0x2029 {
			return &InputError{"SQL은 한 줄이어야 합니다. 줄바꿈이나 제어 문자를 뺀 뒤 다시 실행하세요"}
		}
	}
	if i := strings.IndexByte(trimmed, ';'); i != len(trimmed)-1 {
		return &InputError{"SQL은 한 문장이어야 하고 끝에 ';'가 있어야 합니다. ';'는 문장 끝에만 쓸 수 있습니다(따옴표 안도 포함)"}
	}
	if stdin.MatchString(trimmed) {
		return &InputError{"COPY ... FROM STDIN은 실행할 수 없습니다. psql이 세션 스크립트의 다음 줄을 자료로 읽기 때문입니다"}
	}
	return checkClosed(trimmed)
}

// stdin 값은 COPY ... FROM STDIN을 알아보는 꼴입니다. 대소문자를 가리지 않습니다.
var stdin = regexp.MustCompile(`(?i)\bstdin\b`)

// checkClosed 함수는 따옴표와 괄호가 닫혔는지, 따옴표 밖에 주석과 달러 따옴표가 없는지 확인합니다.
// 백슬래시는 앞에서 거부하므로 남는 이스케이프는 작은따옴표를 두 번 이어 쓰는 것뿐이고, 이것은 닫았다가 바로 다시 여는 것으로 세어도 결과가 같습니다.
func checkClosed(sql string) error {
	inSingle, inDouble, depth := false, false, 0
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		next := byte(0)
		if i+1 < len(sql) {
			next = sql[i+1]
		}
		switch {
		case inSingle:
			inSingle = c != '\''
		case inDouble:
			inDouble = c != '"'
		case c == '\'':
			inSingle = true
		case c == '"':
			inDouble = true
		case c == '(':
			depth++
		case c == ')':
			depth--
			if depth < 0 {
				return &InputError{"SQL의 괄호 짝이 맞지 않습니다"}
			}
		case c == '$':
			return &InputError{"SQL에 따옴표 밖의 $를 쓸 수 없습니다. 달러 따옴표 안의 ';'를 psql이 문장 끝으로 보지 않기 때문입니다"}
		case (c == '-' && next == '-') || (c == '/' && next == '*'):
			return &InputError{"SQL에 주석을 쓸 수 없습니다. 주석 안의 ';'를 psql이 문장 끝으로 보지 않기 때문입니다"}
		}
	}
	switch {
	case inSingle || inDouble:
		return &InputError{"SQL의 따옴표가 닫히지 않았습니다"}
	case depth != 0:
		return &InputError{"SQL의 괄호 짝이 맞지 않습니다"}
	}
	return nil
}

// Executor 값은 명령 하나를 실행해 표준 출력과 표준 오류를 돌려줍니다. 시험에서는 가짜로 바꿉니다.
type Executor func(ctx context.Context, env []string, name string, args ...string) (stdout, stderr []byte, err error)

// Runner 값은 주제 하나의 실행을 한 번에 하나씩 맡습니다. 여러 고루틴이 함께 써도 됩니다.
//
// Timeout은 스크립트 한 번의 시간 한도이고, 0이면 90초입니다. Exec가 nil이면 실제 프로세스를 실행합니다.
type Runner struct {
	Subject subject.Subject
	Timeout time.Duration
	Exec    Executor

	mu sync.Mutex
}

// Result 값은 실행 한 번의 기록과, 화면에 함께 보일 주의 문구입니다.
type Result struct {
	Record   trace.Record
	Warnings []string
}

// Running 함수는 실습 컨테이너가 켜져 있는지 docker inspect로 확인합니다. 확인하지 못하면 false입니다.
func (r *Runner) Running(ctx context.Context) bool {
	out, _, err := r.exec(ctx, "docker", "inspect", "-f", "{{.State.Running}}", r.Subject.Container)
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// Run 함수는 감시 목록 watchID로 sql을 추적하고 기록을 돌려줍니다.
//
// 실패: 입력이 맞지 않으면 *InputError, 이 서버의 다른 실행이 진행 중이면 ErrBusy, 컨테이너가 꺼져 있으면 ErrNotRunning,
// 컨테이너 안에 이 서버 밖의 추적이 진행 중이면 ErrTraceInProgress, 시간 한도를 넘으면 ErrTimeout,
// 스크립트가 실패하거나 기록 형식이 맞지 않으면 그 내용을 담은 오류를 돌려줍니다.
// 부수효과: 스크립트가 실패하거나 시간 한도를 넘거나 취소되면 컨테이너 안에 남은 gdb와 세션 psql을 끝냅니다.
// 남은 psql이 다음 실행의 표시 파일에 함께 깨어나 출력이 섞이는 것을 막기 위해서입니다.
//
// 실행 전에는 정리하지 않고 확인만 합니다. 세션은 컨테이너의 고정된 /tmp 파일을 쓰므로 이 서버의 잠금은
// 터미널에서 직접 실행한 run-trace.sh를 보지 못하고, 실행 전에 정리하면 그 실행을 끝내 버리기 때문입니다.
func (r *Runner) Run(ctx context.Context, watchID, sql string) (Result, error) {
	if err := ValidateSQL(sql); err != nil {
		return Result{}, err
	}
	watch, ok := r.Subject.Watch(watchID)
	if !ok {
		return Result{}, &InputError{fmt.Sprintf("감시 목록 %q가 없습니다", watchID)}
	}
	script := filepath.Join(r.Subject.Dir, r.Subject.RunScript)
	watchFile := filepath.Join(r.Subject.Dir, watch.File)
	// run-trace.sh는 감시 목록을 자기 폴더 안의 파일 이름으로 받습니다(컨테이너 안 /lab/<이름>).
	watchArg, err := filepath.Rel(filepath.Dir(script), watchFile)
	if err != nil || strings.HasPrefix(watchArg, "..") {
		return Result{}, fmt.Errorf("감시 목록 %s가 스크립트 폴더 밖에 있습니다", watch.File)
	}

	if !r.mu.TryLock() {
		return Result{}, ErrBusy
	}
	defer r.mu.Unlock()

	if !r.Running(ctx) {
		return Result{}, ErrNotRunning
	}
	if err := r.checkNoTrace(ctx); err != nil {
		return Result{}, err
	}
	exprs, err := readWatch(watchFile)
	if err != nil {
		return Result{}, err
	}

	timeout := r.Timeout
	if timeout == 0 {
		timeout = 90 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	stdout, stderr, err := r.execEnv(runCtx, MinimalEnv(watch.PGOptions), "sh", script, filepath.ToSlash(watchArg), sql, watch.StackAt)
	if err != nil {
		r.cleanup()
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return Result{}, ErrTimeout
		}
		return Result{}, fmt.Errorf("추적 스크립트가 실패했습니다: %w: %s", err, tail(stderr, 400))
	}

	rec, err := trace.Parse(bytes.NewReader(stdout), exprs)
	if err != nil {
		return Result{}, fmt.Errorf("추적 기록의 형식이 예상과 다릅니다: %w", err)
	}
	return Result{Record: rec, Warnings: warnings(rec, sql)}, nil
}

// warnings 함수는 기록을 보고 학습자가 알아야 할 주의 문구를 만듭니다.
//
// 첫 단계의 query_string이 입력한 SQL과 다르면 알립니다. psql이 :이름 꼴을 변수로 바꾸는 것처럼,
// 서버가 받은 SQL과 PostgreSQL이 받은 SQL이 다르면 학습자가 엉뚱한 실행을 보게 되기 때문입니다.
func warnings(rec trace.Record, sql string) []string {
	var w []string
	if len(rec.Steps) == 0 || rec.Steps[0].Function != "exec_simple_query" {
		return w
	}
	for _, v := range rec.Steps[0].Values {
		if v.Expr != "query_string" {
			continue
		}
		got, truncated, ok := trace.CString(v.Text)
		switch {
		case !ok:
			w = append(w, "첫 단계의 query_string 값을 해석하지 못해, 입력한 SQL과 같은지 확인하지 못했습니다.")
		case truncated && !strings.HasPrefix(sql, got):
			w = append(w, "PostgreSQL이 받은 SQL의 앞부분이 입력한 SQL과 다릅니다. 이 기록은 입력한 SQL의 실행이 아닐 수 있습니다.")
		case !truncated && got != sql:
			w = append(w, fmt.Sprintf("PostgreSQL이 받은 SQL이 입력과 다릅니다: %q. 이 기록은 입력한 SQL의 실행이 아닐 수 있습니다(예: psql이 :이름 꼴을 변수 값으로 바꾼 경우).", got))
		}
	}
	return w
}

// checkNoTrace 함수는 실습 컨테이너 안에 gdb나 세션 psql이 이미 돌고 있으면 ErrTraceInProgress를 돌려줍니다.
// 확인 자체를 하지 못하면 그 이유를 담은 오류를 돌려줍니다.
func (r *Runner) checkNoTrace(ctx context.Context) error {
	out, stderr, err := r.exec(ctx, "docker", "exec", r.Subject.Container, "sh", "-c",
		`if pgrep -x gdb >/dev/null || pgrep -f '^psql -X -At -f /tmp/session\.sql$' >/dev/null; then echo busy; else echo free; fi`)
	switch strings.TrimSpace(string(out)) {
	case "free":
		return nil
	case "busy":
		return ErrTraceInProgress
	}
	return fmt.Errorf("실습 컨테이너 안의 추적 여부를 확인하지 못했습니다: %v: %s", err, tail(stderr, 200))
}

// cleanupScript 값은 실습 컨테이너 안에서 남은 gdb와 세션 psql을 끝내는 셸 명령입니다.
//
// 이 명령은 PostgreSQL 서버를 다시 시작하게 만들지 않아야 합니다. 시간 초과 시험에서 두 가지 길로 실제로 다시 시작했습니다.
//
//  1. gdb를 SIGTERM으로 끝냈을 때. 백엔드가 달리는 동안에는 중단점 명령이 백엔드 메모리에 들어 있어서, 백엔드가 그 명령을
//     만나 SIGTRAP으로 죽었습니다. 그래서 gdb에는 SIGINT를 먼저 보냅니다. SIGINT를 받은 gdb는 백엔드를 멈추고(이때 중단점을
//     메모리에서 빼냅니다) trace.py의 continue가 돌아와 END를 찍고 분리합니다. gdb는 SIGINT를 백엔드에 넘기지 않습니다.
//     5초 안에 끝나지 않을 때만 마지막 수단으로 SIGKILL을 보냅니다.
//  2. psql을 감싼 sh까지 함께 끝냈을 때. 부모를 잃은 프로세스는 컨테이너의 1번 프로세스인 postmaster의 자식이 되고,
//     postmaster는 모르는 자식이 신호로 죽으면 서버 전체의 사고로 다룹니다(postmaster.c 2527~2533행). 그래서 psql만
//     정확히 골라 끝내 감싼 sh가 거두게 하고, 끝내기 전에 /tmp/ready를 만들어 psql이 \!로 띄운 대기 반복이
//     부모를 잃더라도 신호 없이 스스로 끝나게 합니다. 다음 실행의 session.sh가 /tmp/ready를 먼저 지우므로 남아도 됩니다.
const cleanupScript = `pkill -INT -x gdb
i=0
while pgrep -x gdb >/dev/null && [ "$i" -lt 25 ]; do sleep 0.2; i=$((i + 1)); done
pkill -KILL -x gdb
touch /tmp/ready
pkill -f '^psql -X -At -f /tmp/session\.sql$'
true`

// cleanup 함수는 cleanupScript로 실습 컨테이너 안에 남은 gdb와 세션 psql을 끝냅니다. 실패해도 오류를 돌려주지 않습니다.
func (r *Runner) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, _, _ = r.exec(ctx, "docker", "exec", "-u", "root", r.Subject.Container, "sh", "-c", cleanupScript)
}

// exec 함수는 더할 서버 설정이 없는 최소 환경으로 명령을 실행합니다.
func (r *Runner) exec(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	return r.execEnv(ctx, MinimalEnv(""), name, args...)
}

// execEnv 함수는 환경 변수 env로 명령을 실행합니다. Exec가 주어졌으면 그것을 씁니다.
func (r *Runner) execEnv(ctx context.Context, env []string, name string, args ...string) ([]byte, []byte, error) {
	if r.Exec != nil {
		return r.Exec(ctx, env, name, args...)
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env
	// 시간 한도에 끝나는 것은 sh 하나뿐이고, sh가 띄운 docker exec와 grep은 표준 출력을 계속 붙잡고 있습니다.
	// 출력이 닫히기를 끝없이 기다리지 않도록, 프로세스를 끝낸 뒤 2초가 지나면 출력 파이프를 닫고 돌아옵니다.
	// 남은 자식은 호출한 쪽의 정리(cleanup)가 컨테이너 안의 gdb와 psql을 끝내면 함께 끝납니다.
	cmd.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// MinimalEnv 함수는 스크립트에 넘길 환경 변수 목록을 돌려줍니다.
//
// docker 명령을 찾고 OrbStack 같은 Docker 연결 설정을 읽는 데 필요한 것만 서버의 환경에서 가져옵니다.
// PG_OPTIONS는 서버가 정한 statement_timeout 30초에, 주제 정의의 감시 목록이 정한 설정 extra(비어 있어도 됨)를 더한 값입니다.
// 학습자 입력은 환경 변수로 넘기지 않습니다.
func MinimalEnv(extra string) []string {
	pg := "-c statement_timeout=30s"
	if extra != "" {
		pg += " " + extra
	}
	env := []string{"PG_OPTIONS=" + pg}
	for _, k := range []string{"PATH", "HOME", "DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CONFIG", "DOCKER_CERT_PATH", "DOCKER_TLS_VERIFY"} {
		if v, ok := os.LookupEnv(k); ok {
			env = append(env, k+"="+v)
		}
	}
	return env
}

// readWatch 함수는 감시 목록 파일을 읽습니다.
func readWatch(path string) (trace.Exprs, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("감시 목록 열기: %w", err)
	}
	defer f.Close()
	return trace.ParseWatch(f)
}

// tail 함수는 b의 끝 n바이트 이내를 문자열로 돌려줍니다. 오류 문구가 지나치게 길어지지 않게 합니다.
// 잘린 자리가 여러 바이트 글자(한글 등)의 중간이면 그 글자의 나머지 바이트도 뺍니다.
func tail(b []byte, n int) string {
	s := strings.TrimSpace(string(b))
	if len(s) > n {
		s = s[len(s)-n:]
		for len(s) > 0 && !utf8.RuneStart(s[0]) {
			s = s[1:]
		}
	}
	return s
}
