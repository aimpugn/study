package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// psql 은 컨테이너 안에서 띄운 psql 프로세스 하나입니다. 문장은 표준 입력으로 보내고, 출력은 줄 단위로 lines에 들어옵니다.
//
// 오류는 psql이 표준 오류로 쓰므로, 표준 출력과 같은 파이프에 묶고 표준 출력을 줄 버퍼로 바꿉니다(stdbuf -oL).
// 그래야 결과와 오류가 psql이 쓴 순서대로 들어옵니다(2026-10-14 컨테이너에서 확인: 결과, ERROR, 표지 순서가 유지됨).
type psql struct {
	cmd   *exec.Cmd
	in    io.WriteCloser
	lines chan string
	once  sync.Once
}

// startPsql 은 psql을 띄웁니다. args는 접속 옵션 뒤에 붙는 psql 인자입니다.
//
// Errors: 프로세스를 띄우지 못하면 오류를 돌려줍니다.
func startPsql(host string, port int, args ...string) (*psql, error) {
	full := append([]string{"-oL", "psql", "-X", "-h", host, "-p", fmt.Sprint(port), "-d", "postgres"}, args...)
	cmd := exec.Command("stdbuf", full...)
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdout, cmd.Stderr = w, w
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	// 자식이 쓰는 쪽을 물려받았으므로 부모는 닫습니다. 그래야 psql이 끝날 때 읽는 쪽이 EOF를 봅니다.
	w.Close()
	p := &psql{cmd: cmd, in: in, lines: make(chan string, 1024)}
	go func() {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			p.lines <- sc.Text()
		}
		close(p.lines)
	}()
	return p, nil
}

// send 는 SQL 한 줄과 끝 표지를 보냅니다. 표지는 psql의 \echo라서 서버로 가지 않습니다.
func (p *psql) send(sql, marker string) error {
	_, err := fmt.Fprintf(p.in, "%s\n\\echo %s\n", sql, marker)
	return err
}

// until 은 표지 줄이 나올 때까지 줄을 모아 돌려줍니다. 표지 줄은 빼고 돌려줍니다.
//
// Errors: 시간 안에 표지가 오지 않거나 psql이 끝나면 오류를 돌려줍니다.
func (p *psql) until(marker string, timeout time.Duration) ([]string, error) {
	var out []string
	deadline := time.After(timeout)
	for {
		select {
		case l, ok := <-p.lines:
			if !ok {
				return out, fmt.Errorf("psql이 표지 %s 전에 끝났습니다", marker)
			}
			if l == marker {
				return out, nil
			}
			out = append(out, l)
		case <-deadline:
			return out, fmt.Errorf("표지 %s가 %s 안에 오지 않았습니다", marker, timeout)
		}
	}
}

// query 는 관찰용 접속(-A -t -F 탭)에서 SQL 하나를 실행하고 결과 행을 탭으로 나눠 돌려줍니다.
//
// Errors: 표지가 오지 않거나 결과에 ERROR 줄이 있으면 오류를 돌려줍니다.
func (p *psql) query(sql string, marker string) ([][]string, error) {
	if err := p.send(sql, marker); err != nil {
		return nil, err
	}
	lines, err := p.until(marker, 30*time.Second)
	if err != nil {
		return nil, err
	}
	var rows [][]string
	for _, l := range lines {
		if strings.HasPrefix(l, "ERROR:") {
			return nil, fmt.Errorf("관찰 쿼리 실패: %s (%s)", l, sql)
		}
		if l == "" {
			continue
		}
		rows = append(rows, strings.Split(l, "\t"))
	}
	return rows, nil
}

// quit 은 psql을 끝내고 프로세스를 거둡니다. 여러 번 불러도 한 번만 끝냅니다. kill 뒤에는 그 구간의 접속을 바로 끝내고,
// run이 끝날 때 남은 접속을 모두 한 번 더 끝내기 때문입니다.
func (p *psql) quit() {
	p.once.Do(func() {
		fmt.Fprintln(p.in, `\q`)
		p.in.Close()
		done := make(chan struct{})
		go func() { p.cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			p.cmd.Process.Kill()
			<-done
		}
	})
}
