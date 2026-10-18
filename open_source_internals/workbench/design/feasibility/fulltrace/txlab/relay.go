package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

// 이 파일은 맥에서 도는 앱(실험 5의 Java)을 컨테이너 안 기록 서버에 잇습니다(2026-10-15).
//
// 실습 컨테이너는 네트워크가 없고(network none) 서버는 유닉스 소켓만 엽니다. 그래서 맥의 TCP 접속 하나마다
// "docker exec -i pg-lab txlab relay"를 하나 띄우고, 그 표준 입출력을 서버의 유닉스 소켓에 잇습니다.
//
//	앱(pgjdbc) ─TCP 127.0.0.1─ 맥 bridge ─docker exec 표준 입출력─ 컨테이너 relay ─유닉스 소켓─ 백엔드
//
// 양쪽이 같은 바이트를 메시지로 잘라 각자의 시계로 적습니다(wire.go). 같은 연결·방향·순번의 메시지는 같은 메시지이므로,
// 두 기록을 이으면 맥 시계(JFR, 앱 쪽)와 컨테이너 시계(strace, 함수 기록, 서버 쪽)의 차이를 메시지마다 잴 수 있습니다.
// 서버는 이 접속을 유닉스 소켓 접속으로 봅니다. 메시지는 TCP로 직접 붙을 때와 같지만, TCP 자체(연결 맺기, keepalive)는
// 서버 쪽에서 보이지 않습니다. TCP를 보는 일은 실험 6의 몫입니다.

// wireLog 는 메시지를 JSON 한 줄씩 파일에 씁니다. 두 방향의 고루틴이 함께 쓰므로 잠급니다.
type wireLog struct {
	mu  sync.Mutex
	f   *os.File
	w   *bufio.Writer
	enc *json.Encoder
}

// newWireLog 는 path에 새 기록 파일을 만듭니다.
//
// Errors: 파일을 만들지 못하면 오류를 돌려줍니다.
func newWireLog(path string) (*wireLog, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := bufio.NewWriter(f)
	return &wireLog{f: f, w: w, enc: json.NewEncoder(w)}, nil
}

func (l *wireLog) write(ms []wireMsg) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, m := range ms {
		l.enc.Encode(m)
	}
}

// close 는 남은 줄을 내리고 파일을 닫습니다.
func (l *wireLog) close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.w.Flush(); err != nil {
		l.f.Close()
		return err
	}
	return l.f.Close()
}

// pump 는 src에서 읽은 조각을 메시지로 잘라 기록하고 그대로 dst에 씁니다. 자르기를 쓰기보다 먼저 하는 것은, 서버가 이 조각에
// 답하기 전에 요청이 세어져 있어야 하기 때문입니다(SSLRequest의 답 바이트, wireConn 주석). onMsg가 있으면 메시지마다 부릅니다.
//
// Returns: src가 끝나면(EOF) nil, 읽기·쓰기가 실패하면 그 오류입니다.
func pump(dst io.Writer, src io.Reader, c *wireConn, dir byte, lg *wireLog, onMsg func(wireMsg)) error {
	buf := make([]byte, 64<<10)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			ms := c.feed(dir, buf[:n], time.Now().UnixNano())
			lg.write(ms)
			if onMsg != nil {
				for _, m := range ms {
					onMsg(m)
				}
			}
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// cmdRelay 는 컨테이너 안에서 표준 입출력을 서버의 유닉스 소켓에 잇습니다. 맥의 bridge가 접속마다 하나씩 띄웁니다.
//
// 끝은 서버가 정합니다. 앱이 Terminate를 보내거나 앱 쪽이 닫히면(표준 입력 EOF) 소켓의 쓰는 쪽을 닫아 서버에 알리고,
// 서버가 소켓을 닫아 읽기가 끝나면 이 명령도 끝납니다.
//
// Errors: 소켓에 붙지 못하거나 기록 파일을 만들지 못하거나, 서버 쪽 읽기·쓰기가 실패하면 오류를 돌려줍니다.
func cmdRelay(args []string) error {
	fs := flag.NewFlagSet("relay", flag.ContinueOnError)
	id := fs.Int("conn", 0, "연결 번호(맥 bridge가 받은 순서, 1부터)")
	sock := fs.String("sock", fmt.Sprintf("%s/.s.PGSQL.%d", labDir, labPort), "서버의 유닉스 소켓")
	logPath := fs.String("log", "", "메시지 기록 파일(JSON 줄)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 || *logPath == "" {
		return fmt.Errorf("relay: -conn과 -log가 필요합니다")
	}
	uc, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: *sock, Net: "unix"})
	if err != nil {
		return err
	}
	defer uc.Close()
	lg, err := newWireLog(*logPath)
	if err != nil {
		return err
	}
	c := &wireConn{id: *id}

	// 앱 → 서버. 앱 쪽이 끝나면 쓰는 쪽만 닫아, 서버가 EOF를 보고 스스로 끝나게 합니다.
	go func() {
		pump(uc, os.Stdin, c, 'F', lg, nil)
		uc.CloseWrite()
	}()

	// 서버 → 앱. 서버가 닫을 때까지 돕니다.
	berr := pump(os.Stdout, uc, c, 'B', lg, nil)
	if err := lg.close(); err != nil && berr == nil {
		berr = err
	}
	return berr
}

// bridge 는 맥에서 받은 TCP 연결 tc를 컨테이너 안 relay에 잇고, 맥 쪽에서 본 메시지를 macLog에 적습니다.
// 컨테이너 쪽 기록은 /tmp/pgi/out/wire-<id>.jsonl에 남습니다. onMsg는 메시지마다 불립니다(BackendKeyData로 백엔드 pid를 알리는 데 씀).
//
// Errors: relay를 띄우지 못하거나 기록 파일을 만들지 못하면 오류를 돌려줍니다. 앱이 연결을 끊은 뒤 쓰기에 실패하는 것은
// 오류로 치지 않습니다(서버의 마지막 바이트가 앱이 닫은 뒤에 올 수 있음).
func bridge(tc net.Conn, id int, macLog string, onMsg func(wireMsg)) error {
	defer tc.Close()
	lg, err := newWireLog(macLog)
	if err != nil {
		return err
	}
	cmd := exec.Command("docker", "exec", "-i", container, inLabTxlab, "relay",
		"-conn", fmt.Sprint(id), "-log", fmt.Sprintf("%s/out/wire-%d.jsonl", labDir, id))
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	c := &wireConn{id: id}

	// 앱 → 서버. 앱이 닫으면 relay의 표준 입력을 닫아 컨테이너 쪽에 알립니다.
	go func() {
		pump(in, tc, c, 'F', lg, onMsg)
		in.Close()
	}()

	// 서버 → 앱. relay가 끝나면(서버가 닫음) 앱 쪽 연결도 닫습니다.
	pump(tc, out, c, 'B', lg, onMsg)
	tc.Close()
	werr := cmd.Wait()
	if err := lg.close(); err != nil && werr == nil {
		werr = err
	}
	return werr
}
