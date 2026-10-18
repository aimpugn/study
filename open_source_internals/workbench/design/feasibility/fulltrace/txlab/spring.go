package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 이 파일은 실험 5(Spring 전파)의 지휘입니다(2026-10-15). record가 기록 서버를 띄우고 uprobe를 켠 뒤, psql 시나리오의 run 대신
// springRun을 부릅니다.
//
// 지휘는 맥에서 세 프로세스를 잇습니다.
//
//   - Java 앱(spring/): 장면 하나를 돌리며 멈춤 자리마다 표준 출력으로 알리고 go를 기다립니다(Lab.java).
//   - 연결마다의 bridge(relay.go): 앱의 TCP 접속을 컨테이너의 서버에 잇고, 오가는 메시지를 맥 쪽에서 적습니다.
//   - 컨테이너의 observe(observe.go): 멈춤 자리마다 페이지·세션 상태·락을 찍습니다.
//
// 멈춤 자리마다 순서는 "앱이 알림 → 새 연결을 세션으로 알림 → 찍기 → go"입니다. 그래서 각 관찰은 앱의 코드가 그 줄에 멈춘
// 동안의 서버 상태입니다. go 뒤 다음 알림이 시나리오의 blocked 안에 오지 않으면, 앱이 어딘가에서 기다리는 중이므로 그 상태를 한 번 더
// 찍습니다(락을 기다리는지 함께 물음).

// springJar 는 앱이 쓰는 jar 하나입니다. M2는 ~/.m2/repository 안의 상대 경로이고, 비어 있으면 실습 jar 자리에 내려받아 둔
// 것만 씁니다. Sha1이 있으면 그 값과 맞아야 합니다(없으면 ~/.m2의 .sha1 파일과 견줌).
type springJar struct {
	Name string
	M2   string
	Sha1 string
}

// springJars 는 앱의 클래스 경로입니다. 판의 선택과 이유는 WORK 실험 5의 덧붙임(2026-10-15 01:36)에 있습니다.
// pgjdbc의 sha1은 2026-10-14 20:08 KST에 이 맥의 Gradle 캐시가 그 파일에 붙인 디렉터리 이름(Gradle은 파일의 sha1로 디렉터리를 만듦)입니다.
var springJars = []springJar{
	{Name: "spring-aop-6.2.15.jar", M2: "org/springframework/spring-aop/6.2.15"},
	{Name: "spring-beans-6.2.15.jar", M2: "org/springframework/spring-beans/6.2.15"},
	{Name: "spring-context-6.2.15.jar", M2: "org/springframework/spring-context/6.2.15"},
	{Name: "spring-core-6.2.15.jar", M2: "org/springframework/spring-core/6.2.15"},
	{Name: "spring-expression-6.2.15.jar", M2: "org/springframework/spring-expression/6.2.15"},
	{Name: "spring-jcl-6.2.15.jar", M2: "org/springframework/spring-jcl/6.2.15"},
	{Name: "spring-jdbc-6.2.15.jar", M2: "org/springframework/spring-jdbc/6.2.15"},
	{Name: "spring-tx-6.2.15.jar", M2: "org/springframework/spring-tx/6.2.15"},
	{Name: "micrometer-observation-1.14.14.jar", M2: "io/micrometer/micrometer-observation/1.14.14"},
	{Name: "micrometer-commons-1.14.14.jar", M2: "io/micrometer/micrometer-commons/1.14.14"},
	{Name: "HikariCP-5.1.0.jar", M2: "com/zaxxer/HikariCP/5.1.0"},
	{Name: "slf4j-api-2.0.17.jar", M2: "org/slf4j/slf4j-api/2.0.17"},
	{Name: "logback-classic-1.5.22.jar", M2: "ch/qos/logback/logback-classic/1.5.22"},
	{Name: "logback-core-1.5.22.jar", M2: "ch/qos/logback/logback-core/1.5.22"},
	{Name: "postgresql-42.7.13.jar", Sha1: "a6e1bd21b412d6ffb3df23cd13d507bc2cc9e37d"},
}

// springTrace 는 JFR이 들어감·나옴을 적을 메서드입니다(JDK 25 jdk.MethodTrace, 세미콜론으로 나눔). 클래스 이름만 쓰면 그 클래스의
// 모든 메서드입니다. pgjdbc의 이름은 그 소스를 읽기 전이라 추론이고, 없는 이름은 JFR이 오류 없이 넘깁니다(2026-10-15 시험).
//
// 한계(2026-10-15 시험, GraalVM 25.0.4와 OpenJDK 25.0.2): 아래 메서드가 부른 쪽에서 올라온 예외로 끝나면 그 호출은 기록에 없습니다.
// 메서드 안의 throw로 끝난 호출은 남고, 그 throw를 같은 메서드의 finally가 받으면 같은 시작 시각으로 두 번 남습니다(긴 쪽이 실제 끝).
// 그래서 예외가 만들어진 자리(jdk.JavaExceptionThrow)와 Spring·HikariCP 로그를 함께 남깁니다.
var springTrace = strings.Join([]string{
	"org.springframework.transaction.interceptor.TransactionInterceptor::invoke",
	"org.springframework.transaction.interceptor.TransactionAspectSupport::invokeWithinTransaction",
	"org.springframework.transaction.interceptor.TransactionAspectSupport::createTransactionIfNecessary",
	"org.springframework.transaction.interceptor.TransactionAspectSupport::commitTransactionAfterReturning",
	"org.springframework.transaction.interceptor.TransactionAspectSupport::completeTransactionAfterThrowing",
	"org.springframework.transaction.support.AbstractPlatformTransactionManager",
	"org.springframework.jdbc.datasource.DataSourceTransactionManager",
	"org.springframework.jdbc.datasource.ConnectionHolder::createSavepoint",
	"org.springframework.jdbc.datasource.JdbcTransactionObjectSupport::createSavepoint",
	"org.springframework.jdbc.datasource.JdbcTransactionObjectSupport::rollbackToSavepoint",
	"org.springframework.jdbc.datasource.JdbcTransactionObjectSupport::releaseSavepoint",
	"org.springframework.jdbc.core.JdbcTemplate::update",
	"org.springframework.jdbc.core.JdbcTemplate::execute",
	"com.zaxxer.hikari.HikariDataSource::getConnection",
	"com.zaxxer.hikari.pool.HikariPool::getConnection",
	"com.zaxxer.hikari.pool.PoolBase::isConnectionDead",
	"com.zaxxer.hikari.pool.ProxyConnection::close",
	"com.zaxxer.hikari.util.ConcurrentBag::borrow",
	"com.zaxxer.hikari.util.ConcurrentBag::requite",
	"org.postgresql.jdbc.PgConnection::setAutoCommit",
	"org.postgresql.jdbc.PgConnection::commit",
	"org.postgresql.jdbc.PgConnection::rollback",
	"org.postgresql.jdbc.PgConnection::setSavepoint",
	"org.postgresql.jdbc.PgConnection::releaseSavepoint",
	"org.postgresql.jdbc.PgConnection::isValid",
	"org.postgresql.core.v3.QueryExecutorImpl::execute",
	"lab.Outer",
	"lab.Inner",
}, ";")

// 지휘의 기다림입니다. 멈춘 채로 찍기까지의 시간은 시나리오의 blocked(기본 1.5초)입니다. 기본값은 실습 서버의 deadlock_timeout(기본 1초)
// 뒤라서, 같은 행을 기다리는 장면에서는 교착 검사가 한 번 돈 뒤를 찍습니다. 기다림을 묻고 찍는 데 0.7초쯤 걸리므로, 풀이 1개인 장면은
// 연결 기다림(2초, LabConfig)이 끝나기 전에 다 찍도록 0.5초로 줄입니다(2026-10-15 18:36 기록에서 1.5초로 찍다가 관찰이 그 끝과 겹침, WORK 실험 5).
// endCheckpoint와 lastCheckpoint는 앱의 마지막 두 멈춤 자리(Main.java)입니다. end 뒤는 Spring과 풀을 닫는 구간이고 closed 뒤는 JVM이
// 끝나며 JFR 기록을 쓰는 구간이라, 사전 등록한 질문 가운데 그 사이의 서버 상태를 묻는 것이 없습니다. 그래서 그 뒤로는 멈춘 채로 찍지 않습니다
// (풀을 닫는 데 1.5초가 넘게 걸려도 관찰이 하나 끼지 않게, WORK 실험 5 「첫 시험에서 볼 것」).
const (
	stepLimit      = 20 * time.Second
	endCheckpoint  = "end"
	lastCheckpoint = "closed"
)

// checkpoint 는 앱의 멈춤 자리 하나입니다. JavaNs는 앱이 알린 시각(맥 시계), AtNs는 지휘가 그 줄을 읽은 시각(맥 시계),
// CaptureNs는 관찰을 시작한 시각(컨테이너 시계), GoNs는 go를 보낸 시각(맥 시계)입니다.
type checkpoint struct {
	N         int    `json:"n"`
	Label     string `json:"label"`
	JavaNs    int64  `json:"javaNs"`
	AtNs      int64  `json:"atNs"`
	CaptureNs int64  `json:"captureNs"`
	GoNs      int64  `json:"goNs"`
}

// javaNote 는 앱이 받은 예외 하나입니다. Chain은 원인 사슬의 클래스 이름입니다(바깥 ← 원인).
type javaNote struct {
	Where   string `json:"where"`
	JavaNs  int64  `json:"javaNs"`
	Chain   string `json:"chain"`
	Message string `json:"message"`
}

// connRecord 는 앱이 연 TCP 접속 하나입니다. Conn은 받은 순서(1부터)이고 wire-<Conn>.jsonl·wire-mac-<Conn>.jsonl과 같은 번호입니다.
type connRecord struct {
	Conn     int    `json:"conn"`
	Pid      int    `json:"pid"`
	AcceptNs int64  `json:"acceptNs"`
	CloseNs  int64  `json:"closeNs"`
	Err      string `json:"err,omitempty"`
}

// blockedRecord 는 go 뒤 다음 알림이 오지 않아 찍은 관찰 하나입니다. After는 그 앞 멈춤 자리의 순번, Waiters는 그때 락을 기다린 백엔드입니다.
type blockedRecord struct {
	After     int   `json:"after"`
	Ns        int64 `json:"ns"`
	CaptureNs int64 `json:"captureNs"`
	Waiters   []int `json:"waiters"`
}

// springResult 는 spring.json 전체입니다. 컨테이너 쪽 관찰(steps.json)과 메시지 기록(wire-*.jsonl)은 따로 남고, 해석기가 잇습니다.
type springResult struct {
	Scene       string          `json:"scene"`
	Pool        int             `json:"pool"`
	Port        int             `json:"port"`
	Java        string          `json:"java"`
	Jars        []string        `json:"jars"`
	Trace       string          `json:"trace"`
	Checkpoints []checkpoint    `json:"checkpoints"`
	Notes       []javaNote      `json:"notes"`
	Conns       []connRecord    `json:"conns"`
	Blocked     []blockedRecord `json:"blocked"`
	Stdout      []string        `json:"stdout,omitempty"`
	JavaExit    int             `json:"javaExit"`
	StartNs     int64           `json:"startNs"`
	EndNs       int64           `json:"endNs"`
}

// springRun 은 Spring 시나리오 하나를 돌리고 <out>에 spring.json, spring.jfr, java.log, wire-mac-*.jsonl을 씁니다.
// labTmp는 기록 빌드가 있는 실습 임시 자리이고, 그 아래 spring/에 jar와 컴파일한 클래스를 둡니다.
//
// Errors: jar가 없거나 sha1이 다르거나, 컴파일·실행이 실패하거나, 앱이 멈춤 자리 없이 끝나거나, 한 멈춤 자리 뒤 20초 안에 다음
// 알림이 오지 않거나, 관찰이 실패하면 오류를 돌려줍니다. 오류로 끝날 때도 그때까지의 spring.json은 씁니다.
func springRun(sc *scenario, out, labTmp string) error {
	res := springResult{Scene: sc.Spring, Pool: sc.Pool, Trace: springTrace, StartNs: time.Now().UnixNano()}
	defer func() {
		res.EndNs = time.Now().UnixNano()
		if b, err := json.MarshalIndent(res, "", " "); err == nil {
			os.WriteFile(filepath.Join(out, "spring.json"), b, 0o644)
		}
	}()
	javaHome, err := exec.Command("/usr/libexec/java_home", "-v", "25").Output()
	if err != nil {
		return fmt.Errorf("JDK 25를 찾지 못했습니다(JFR 메서드 추적이 JDK 25부터): %w", err)
	}
	java := filepath.Join(strings.TrimSpace(string(javaHome)), "bin")
	if v, err := exec.Command(filepath.Join(java, "java"), "-version").CombinedOutput(); err == nil {
		res.Java = strings.TrimSpace(string(v))
	}
	cp, jars, err := prepareSpring(java, filepath.Join(labTmp, "spring"), filepath.Join(out, appSrcDir))
	if err != nil {
		return err
	}
	res.Jars = jars

	// 앱의 접속을 받을 자리. 포트는 운영 체제가 고릅니다.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer ln.Close()
	res.Port = ln.Addr().(*net.TCPAddr).Port

	// 컨테이너의 관찰. 첫 줄 "ready <pid>"를 받아야 시작합니다.
	obs, err := startObserver()
	if err != nil {
		return fmt.Errorf("observe: %w", err)
	}
	defer obs.stop()

	// 접속마다 bridge를 띄웁니다. BackendKeyData를 보면 그 접속의 백엔드 pid를 pids로 알립니다.
	var mu sync.Mutex
	var wg sync.WaitGroup
	conns := map[int]*connRecord{}
	pids := make(chan [2]int, 64)
	go func() {
		for id := 1; ; id++ {
			tc, err := ln.Accept()
			if err != nil {
				return
			}
			rec := &connRecord{Conn: id, AcceptNs: time.Now().UnixNano()}
			mu.Lock()
			conns[id] = rec
			mu.Unlock()
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				err := bridge(tc, id, filepath.Join(out, fmt.Sprintf("wire-mac-%d.jsonl", id)), func(m wireMsg) {
					if m.Dir == "B" && m.Type == "K" {
						pids <- [2]int{id, m.Pid}
					}
				})
				mu.Lock()
				rec.CloseNs = time.Now().UnixNano()
				if err != nil {
					rec.Err = err.Error()
				}
				mu.Unlock()
			}(id)
		}
	}()

	// 앱. 표준 출력은 멈춤 자리의 약속만, 표준 오류는 로그입니다.
	jlog, err := os.Create(filepath.Join(out, "java.log"))
	if err != nil {
		return err
	}
	defer jlog.Close()
	cmd := exec.Command(filepath.Join(java, "java"),
		"-Dlab.url=jdbc:postgresql://127.0.0.1:"+strconv.Itoa(res.Port)+"/postgres",
		"-Dlab.pool="+strconv.Itoa(sc.Pool),
		"-XX:StartFlightRecording:filename="+filepath.Join(out, "spring.jfr")+",exceptions=all,method-trace="+springTrace+",jdk.ThreadPark#threshold=1ms",
		"-cp", cp, "lab.Main", sc.Spring)
	cmd.Stderr = jlog
	jin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	jout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	lines := make(chan string, 64)
	go func() {
		s := bufio.NewScanner(jout)
		for s.Scan() {
			lines <- s.Text()
		}
		close(lines)
	}()
	stopJava := func() { cmd.Process.Kill(); cmd.Wait() }

	// 새로 알게 된 백엔드를 관찰에 세션으로 알립니다. 멈춤 자리마다 찍기 전에 부릅니다.
	known := map[int]bool{}
	syncSessions := func() error {
		for {
			select {
			case p := <-pids:
				if known[p[0]] {
					continue
				}
				known[p[0]] = true
				mu.Lock()
				conns[p[0]].Pid = p[1]
				mu.Unlock()
				if _, err := obs.ask(fmt.Sprintf("session C%d %d", p[0], p[1])); err != nil {
					return err
				}
			default:
				return nil
			}
		}
	}

	// 기다림의 끝: go 뒤 아직 멈춘 채로 찍지 않았으면 sc.Blocked, 첫 멈춤 자리 전이면 JVM 시작과 풀 채우기에 60초, 그 밖에는 stepLimit.
	n, awaiting, blockedTaken := 0, false, false
	startAt, goAt := time.Now(), time.Now()
loop:
	for {
		deadline := goAt.Add(stepLimit)
		switch {
		case awaiting && !blockedTaken:
			deadline = goAt.Add(sc.Blocked)
		case n == 0:
			deadline = startAt.Add(60 * time.Second)
		}
		select {
		case line, ok := <-lines:
			if !ok {
				break loop
			}
			f := strings.SplitN(line, " ", 4)
			switch {
			case len(f) == 4 && f[0] == "@@TXLAB" && f[1] == "AT":
				at := time.Now().UnixNano()
				if err := syncSessions(); err != nil {
					stopJava()
					return err
				}
				n++
				reply, err := obs.ask(fmt.Sprintf("capture %d %s", n, f[2]))
				if err != nil {
					stopJava()
					return fmt.Errorf("멈춤 자리 %d %s: %w", n, f[2], err)
				}
				c := checkpoint{N: n, Label: f[2], AtNs: at, CaptureNs: replyNs(reply)}
				c.JavaNs, _ = strconv.ParseInt(f[3], 10, 64)
				c.GoNs = time.Now().UnixNano()
				res.Checkpoints = append(res.Checkpoints, c)
				fmt.Fprintln(jin, "go")
				goAt, blockedTaken = time.Now(), false
				awaiting = f[2] != endCheckpoint && f[2] != lastCheckpoint
			case len(f) == 4 && f[0] == "@@TXLAB" && f[1] == "NOTE":
				note := javaNote{Where: f[2]}
				rest := strings.SplitN(f[3], " ", 2)
				note.JavaNs, _ = strconv.ParseInt(rest[0], 10, 64)
				if len(rest) == 2 {
					note.Chain, note.Message, _ = strings.Cut(rest[1], " | ")
				}
				res.Notes = append(res.Notes, note)
			default:
				res.Stdout = append(res.Stdout, line)
			}
		case <-time.After(time.Until(deadline)):
			if !awaiting || blockedTaken {
				stopJava()
				return fmt.Errorf("멈춤 자리 %d 뒤 제한 시간 안에 앱이 다음 자리에 오지도 끝나지도 않았습니다", n)
			}
			// 앱이 go 뒤 어딘가에서 기다리는 중입니다. 서버 쪽에서 락을 기다리는지 묻고 지금 상태를 찍습니다.
			if err := syncSessions(); err != nil {
				stopJava()
				return err
			}
			b := blockedRecord{After: n, Ns: time.Now().UnixNano()}
			reply, err := obs.ask("waiters")
			if err != nil {
				stopJava()
				return err
			}
			for _, p := range strings.Fields(strings.TrimPrefix(reply, "waiters")) {
				v, _ := strconv.Atoi(p)
				b.Waiters = append(b.Waiters, v)
			}
			if reply, err = obs.ask(fmt.Sprintf("capture %d blocked", n)); err != nil {
				stopJava()
				return err
			}
			b.CaptureNs = replyNs(reply)
			res.Blocked = append(res.Blocked, b)
			blockedTaken = true
		}
	}
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.JavaExit = ee.ExitCode()
		} else {
			return err
		}
	}
	// 앱이 끝났으면 풀이 연결을 모두 닫았습니다. bridge가 서버 쪽 끝까지 받고 끝나기를 기다립니다.
	ln.Close()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("앱이 끝난 뒤 10초 안에 연결이 모두 닫히지 않았습니다")
	}
	mu.Lock()
	for id := 1; id <= len(conns); id++ {
		res.Conns = append(res.Conns, *conns[id])
	}
	mu.Unlock()
	if _, err := obs.ask("end"); err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("앱이 멈춤 자리 없이 끝났습니다(종료 코드 %d, %s)", res.JavaExit, filepath.Join(out, "java.log"))
	}
	if res.JavaExit != 0 {
		return fmt.Errorf("앱이 종료 코드 %d로 끝났습니다(%s)", res.JavaExit, filepath.Join(out, "java.log"))
	}
	return nil
}

// replyNs 는 observe의 "ok <시각>" 답에서 시각을 읽습니다. 읽지 못하면 0입니다.
func replyNs(reply string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(reply, "ok")), 10, 64)
	return v
}

// observerProc 은 컨테이너에서 도는 observe 하나입니다. 명령 한 줄에 답 한 줄을 받습니다.
type observerProc struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out *bufio.Scanner
}

// startObserver 는 컨테이너에서 observe를 띄우고 첫 답(ready)을 기다립니다.
//
// Errors: 프로세스를 띄우지 못하거나 첫 답이 ready가 아니면 오류를 돌려줍니다.
func startObserver() (*observerProc, error) {
	cmd := exec.Command("docker", "exec", "-i", container, inLabTxlab, "observe", "-scenario", "/tmp/txlab-scenario.txt", "-out", labDir+"/out")
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	p := &observerProc{cmd: cmd, in: in, out: bufio.NewScanner(out)}
	if !p.out.Scan() || !strings.HasPrefix(p.out.Text(), "ready ") {
		p.stop()
		return nil, fmt.Errorf("observe가 준비되지 않았습니다: %q", p.out.Text())
	}
	return p, nil
}

// ask 는 명령 한 줄을 보내고 답 한 줄을 받습니다.
//
// Errors: 답이 오지 않거나 "error "로 시작하면 오류를 돌려줍니다.
func (p *observerProc) ask(line string) (string, error) {
	if _, err := fmt.Fprintln(p.in, line); err != nil {
		return "", err
	}
	if !p.out.Scan() {
		return "", fmt.Errorf("observe가 %q에 답하지 않고 끝났습니다", line)
	}
	if r := p.out.Text(); strings.HasPrefix(r, "error ") {
		return "", fmt.Errorf("observe: %s", strings.TrimPrefix(r, "error "))
	} else {
		return r, nil
	}
}

// stop 은 observe의 입력을 닫고 끝나기를 기다립니다. end를 보낸 뒤라면 이미 끝나 있습니다.
func (p *observerProc) stop() {
	p.in.Close()
	p.cmd.Wait()
}

// appSrcDir 는 기록 디렉터리 안에서 그 기록을 만든 앱 소스의 사본이 놓이는 자리입니다. 해석기(decode spring)는 앱 코드의 소스를
// 여기서 먼저 읽습니다. 저장소의 소스는 기록 뒤에 바뀔 수 있어, JFR 스택의 줄 번호와 맞는 것은 이 사본뿐입니다(2026-10-15, 머리
// 주석이 한 줄 늘어 앞선 기록 열두 벌의 부른 줄이 화면에서 한 줄씩 어긋났음).
const appSrcDir = "app-src"

// prepareSpring 은 jar를 dir/jars로 모으고(sha1 확인), 앱 소스를 srcOut/lab에 복사한 뒤 그 사본을 dir/classes로 컴파일합니다.
// 앱 소스는 이 디렉터리의 spring/에 있습니다. 사본을 컴파일하므로 기록에 남은 소스가 실행한 클래스와 같다는 것이 보장됩니다.
// 내려받은 jar는 ~/.m2에 두지 않고 dir/jars에만 둡니다. 사용자가 캐시를 정리해도 기록을 다시 만들 수 있게 하려는 것입니다.
//
// Returns: 클래스 경로와, jar마다 "이름 sha1" 줄입니다.
// Errors: jar가 없거나 sha1이 다르거나, 소스를 복사하지 못하거나 컴파일이 실패하면 오류를 돌려줍니다. pgjdbc가 없으면 내려받기
// 허락이 필요하다고 알립니다.
func prepareSpring(javaBin, dir, srcOut string) (string, []string, error) {
	jarDir := filepath.Join(dir, "jars")
	if err := os.MkdirAll(jarDir, 0o755); err != nil {
		return "", nil, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}
	var cp, recs []string
	for _, j := range springJars {
		dst := filepath.Join(jarDir, j.Name)
		want := j.Sha1
		if _, err := os.Stat(dst); err != nil && j.M2 != "" {
			src := filepath.Join(home, ".m2/repository", j.M2, j.Name)
			b, err := os.ReadFile(src + ".sha1")
			if err != nil {
				return "", nil, fmt.Errorf("%s의 sha1을 읽지 못했습니다: %w", src, err)
			}
			want = strings.Fields(string(b))[0]
			data, err := os.ReadFile(src)
			if err != nil {
				return "", nil, err
			}
			if err := os.WriteFile(dst, data, 0o644); err != nil {
				return "", nil, err
			}
		} else if err != nil {
			return "", nil, fmt.Errorf("%s가 없습니다. pgjdbc는 이 맥에 없어 Maven Central에서 내려받아야 하고, 사용자 허락을 기다립니다(WORK 실험 5)", dst)
		}
		data, err := os.ReadFile(dst)
		if err != nil {
			return "", nil, err
		}
		sum := sha1.Sum(data)
		got := hex.EncodeToString(sum[:])
		if want != "" && got != want {
			return "", nil, fmt.Errorf("%s의 sha1이 %s입니다(%s이어야 함)", dst, got, want)
		}
		cp = append(cp, dst)
		recs = append(recs, j.Name+" "+got)
	}

	// 앱을 매번 다시 컴파일합니다. 몇 초면 끝나고, 소스와 기록이 어긋날 일을 없앱니다.
	classes := filepath.Join(dir, "classes")
	if err := os.RemoveAll(classes); err != nil {
		return "", nil, err
	}
	orig, err := filepath.Glob("spring/src/lab/*.java")
	if err != nil || len(orig) == 0 {
		return "", nil, fmt.Errorf("앱 소스(spring/src/lab)를 찾지 못했습니다: %v", err)
	}
	srcs, err := copyFiles(orig, filepath.Join(srcOut, "lab"))
	if err != nil {
		return "", nil, fmt.Errorf("앱 소스를 기록에 복사하지 못했습니다: %w", err)
	}
	args := append([]string{"-encoding", "UTF-8", "-Xlint:all", "-d", classes, "-cp", strings.Join(cp, ":")}, srcs...)
	if b, err := exec.Command(filepath.Join(javaBin, "javac"), args...).CombinedOutput(); err != nil {
		return "", nil, fmt.Errorf("javac: %v\n%s", err, b)
	}
	lb, err := os.ReadFile("spring/resources/logback.xml")
	if err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(filepath.Join(classes, "logback.xml"), lb, 0o644); err != nil {
		return "", nil, err
	}
	return strings.Join(append([]string{classes}, cp...), ":"), recs, nil
}

// copyFiles 는 files를 dst 디렉터리에 같은 이름으로 복사합니다. dst는 없으면 만듭니다.
//
// Returns: 복사한 파일의 경로들(files와 같은 차례)입니다.
// Errors: 읽거나 쓰지 못하면 그때까지 복사한 것과 함께 오류를 돌려줍니다.
func copyFiles(files []string, dst string) ([]string, error) {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return nil, err
	}
	var out []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return out, err
		}
		p := filepath.Join(dst, filepath.Base(f))
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return out, err
		}
		out = append(out, p)
	}
	return out, nil
}
