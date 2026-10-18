package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// 기록 서버의 파일 자리입니다. 지난 단계의 serve.sh와 같아서 해석기의 경로 규칙(/data/ 아래)을 그대로 씁니다.
const (
	labDir   = "/tmp/pgi"
	traceDir = "/tmp/pgtrace"
)

// straceCalls 는 strace가 기록할 시스템 콜입니다. 지난 단계의 파일·소켓 호출에, 세션 사이의 깨움(kill로 보내는 SIGURG)과
// 잠듦(epoll_pwait)을 더했습니다.
var straceCalls = "openat,pread64,preadv,pwrite64,pwritev,fdatasync,fsync,sendto,recvfrom,kill,epoll_pwait"

// cmdServe 는 컨테이너 안에서 기록 서버를 새로 띄웁니다.
//
// 순서: 이전 서버를 내리고 /tmp/pgi와 /tmp/pgtrace를 비움 → initdb → 기록 없이 띄워 pageinspect와 시나리오의 setup을 실행하고 내림
// → strace 아래에서 함수 기록(PGTRACE_DIR)을 켜고 다시 띄움. 본 실행은 공유 버퍼가 빈 상태에서 시작합니다.
// 서버 실행 파일은 -bin으로 받습니다. uprobe를 걸려면 OrbStack 머신이 같은 파일(같은 inode)을 볼 수 있는 맥 공유 경로여야 합니다.
// 그 경로는 컴파일 때 정한 설치 경로와 모양이 달라 share·lib를 컴파일 때 경로(/usr/local/pgsql)에서 찾습니다.
//
// Errors: initdb, 준비 SQL, 서버 준비 중 하나라도 실패하면 오류를 돌려줍니다.
func cmdServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	bin := fs.String("bin", "/lab/.lab-tmp/postgres", "기록 빌드의 postgres")
	scPath := fs.String("scenario", "", "setup 줄을 읽을 시나리오 파일")
	if err := fs.Parse(args); err != nil {
		return err
	}
	sc, err := parseScenario(*scPath)
	if err != nil {
		return err
	}
	stopServer()
	for _, d := range []string{labDir, traceDir} {
		if err := os.RemoveAll(d); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(labDir, "out"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(traceDir, 0o755); err != nil {
		return err
	}
	if out, err := exec.Command("initdb", "-D", labDir+"/data", "--no-locale", "-E", "UTF8").CombinedOutput(); err != nil {
		return fmt.Errorf("initdb: %v\n%s", err, out)
	}
	srvArgs := []string{"-D", labDir + "/data", "-c", "listen_addresses=", "-c", "unix_socket_directories=" + labDir, "-p", fmt.Sprint(labPort)}

	// 준비: 기록 없이 띄워 확장과 표를 만들고 내립니다.
	setupLog, err := os.Create(labDir + "/server-setup.log")
	if err != nil {
		return err
	}
	pre := exec.Command(*bin, srvArgs...)
	pre.Stdout, pre.Stderr = setupLog, setupLog
	if err := pre.Start(); err != nil {
		return err
	}
	if err := waitReady(30 * time.Second); err != nil {
		return err
	}
	psqlArgs := []string{"-X", "-q", "-v", "ON_ERROR_STOP=1", "-h", labHost, "-p", fmt.Sprint(labPort), "-d", "postgres"}
	for _, q := range setupBatches(sc.Setup) {
		psqlArgs = append(psqlArgs, "-c", q)
	}
	cmd := exec.Command("psql", psqlArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		stopServer()
		return fmt.Errorf("준비 SQL: %v\n%s", err, out)
	}
	stopServer()
	pre.Wait()

	// 본 실행: strace가 postmaster와 그 자식을 모두 따라가고, 기록 함수가 프로세스마다 /tmp/pgtrace/<pid>.bin을 씁니다.
	// setsid로 이 명령의 세션에서 떼어 내야 docker exec가 끝나도 서버가 남습니다.
	log, err := os.Create(labDir + "/server.log")
	if err != nil {
		return err
	}
	st := exec.Command("strace", append([]string{"-f", "-yy", "-s", "1024", "--seccomp-bpf", "-ttt", "-T", "-e", "trace=" + straceCalls,
		"-o", labDir + "/all.strace", *bin}, srvArgs...)...)
	st.Env = append(os.Environ(), "PGTRACE_DIR="+traceDir)
	st.Stdout, st.Stderr = log, log
	st.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := st.Start(); err != nil {
		return err
	}
	if err := os.WriteFile(labDir+"/strace.pid", []byte(fmt.Sprint(st.Process.Pid)), 0o644); err != nil {
		return err
	}
	st.Process.Release()
	if err := waitReady(60 * time.Second); err != nil {
		return err
	}
	fmt.Println("ready", time.Now().UnixNano())
	return nil
}

// setupBatches 는 준비 SQL을 psql의 -c 값들로 묶습니다. psql은 -c 하나를 한 번에 보내고, 그 안의 여러 문장은 한 트랜잭션으로
// 돕니다. 그래서 확장과 표를 만드는 줄은 한 -c로 묶고, 트랜잭션 블록 안에서 돌 수 없는 ALTER SYSTEM 줄만 따로 보냅니다.
//
// 줄마다 -c를 따로 주면 준비가 트랜잭션 번호를 줄마다 하나씩 써서 세션의 첫 번호가 밀립니다. 2026-10-14 실험 3의 첫 시험 기록에서
// S1이 753이 아니라 755를 받았습니다. 한 -c로 묶으면 실험 1·2처럼 S1이 753을 받습니다.
func setupBatches(setup []string) []string {
	batch := []string{"CREATE EXTENSION pageinspect;"}
	var alone []string
	for _, q := range setup {
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(q)), "ALTER SYSTEM") {
			alone = append(alone, q)
		} else {
			batch = append(batch, q)
		}
	}
	return append([]string{strings.Join(batch, "\n")}, alone...)
}

// cmdStop 은 기록 서버를 내리고 strace가 기록을 다 쓰고 끝날 때까지 기다린 뒤, 시나리오와 무관한 백엔드의 함수 기록을 지웁니다.
func cmdStop(args []string) error {
	stopServer()
	b, err := os.ReadFile(labDir + "/strace.pid")
	if err != nil {
		return nil
	}
	var pid int
	fmt.Sscan(string(b), &pid)
	gone := false
	for i := 0; i < 100 && !gone; i++ {
		// 신호 0은 프로세스가 아직 있는지만 봅니다.
		gone = syscall.Kill(pid, 0) != nil
		time.Sleep(100 * time.Millisecond)
	}
	if !gone {
		return fmt.Errorf("strace(pid %d)가 10초 안에 끝나지 않았습니다", pid)
	}
	return pruneTraces()
}

// pruneTraces 는 관찰용 접속과 pg_isready 접속의 백엔드 기록을 지웁니다. 2026-10-14 실험 1의 열 벌에서 이 둘이
// 한 벌 약 490 MB 가운데 410~500 MB를 차지했습니다(관찰 쿼리마다 계획·실행 전체가 기록되고, 접속마다 카탈로그 캐시를 채움).
// 지울 백엔드는 strace에 남은 시작 메시지의 application_name으로 찾고, 관찰용 접속은 steps.json의 observerPids(kill 뒤
// 새로 연 접속까지)로 찾습니다.
func pruneTraces() error {
	drop := map[int]bool{}
	if b, err := os.ReadFile(labDir + "/out/steps.json"); err == nil {
		var r runResult
		if json.Unmarshal(b, &r) == nil {
			for _, p := range append(r.ObserverPids, r.Observer) {
				if p != 0 {
					drop[p] = true
				}
			}
		}
	}
	f, err := os.Open(labDir + "/all.strace")
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if p, ok := isreadyStartup(sc.Text()); ok {
			drop[p] = true
		}
	}
	for p := range drop {
		os.Remove(fmt.Sprintf("%s/%d.bin", traceDir, p))
	}
	return sc.Err()
}

// isreadyStartup 은 strace 한 줄이 pg_isready 접속의 시작 메시지를 받은 recvfrom이면 그 백엔드 pid를 돌려줍니다.
//
// 다른 pid의 줄이 끼면 strace는 한 호출을 "recvfrom(... <unfinished ...>"와 "<... recvfrom resumed>..." 두 줄로 나누고,
// 받은 바이트는 resumed 줄에만 적습니다. 그래서 두 꼴을 모두 봅니다.
func isreadyStartup(l string) (pid int, ok bool) {
	call := strings.Contains(l, " recvfrom(") || strings.Contains(l, "<... recvfrom resumed>")
	if !call || !strings.Contains(l, `application_name\0pg_isready\0`) {
		return 0, false
	}
	if _, err := fmt.Sscan(l, &pid); err != nil {
		return 0, false
	}
	return pid, true
}

// stopServer 는 /tmp/pgi/data의 서버를 빠르게 내립니다. 떠 있지 않으면 아무것도 하지 않습니다.
func stopServer() {
	exec.Command("pg_ctl", "-D", labDir+"/data", "stop", "-m", "fast", "-w").Run()
}

// waitReady 는 서버가 접속을 받을 때까지 기다립니다.
func waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if exec.Command("pg_isready", "-q", "-h", labHost, "-p", fmt.Sprint(labPort)).Run() == nil {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("서버가 %s 안에 준비되지 않았습니다", timeout)
}
