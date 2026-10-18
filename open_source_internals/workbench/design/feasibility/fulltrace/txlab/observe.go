package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// cmdObserve 는 컨테이너 안에서 관찰용 접속 하나를 열고, 맥의 지휘(spring.go)가 표준 입력으로 보내는 명령대로 관찰합니다(실험 5).
// 실험 1~4의 run은 세션도 스스로 띄웠지만, 실험 5의 세션은 맥의 Java 앱이 연 연결이라 이 명령은 관찰만 맡습니다.
//
// 명령은 한 줄에 하나이고, 답도 한 줄입니다.
//
//	session <이름> <pid>    앱의 연결 하나를 세션으로 더함          → "ok"
//	capture <순번> <이름>   지금의 페이지·세션 상태·락을 찍음        → "ok <찍기 시작한 시각>"
//	waiters                  세션 가운데 락을 기다리는 pid           → "waiters <pid> …"
//	end                      <out>/steps.json을 쓰고 끝냄            → "ok"
//
// Errors: 관찰용 접속을 열지 못하면 오류를 돌려줍니다. 명령이 실패하면 "error <내용>"으로 답한 뒤 그 오류를 돌려주고 끝납니다.
// 표준 입력이 end 없이 끝나도 오류입니다(맥 쪽이 중간에 멈춘 것이므로 반쪽 기록을 남기지 않음).
func cmdObserve(args []string) error {
	fs := flag.NewFlagSet("observe", flag.ContinueOnError)
	scPath := fs.String("scenario", "", "시나리오 파일(제목과 watch 줄을 읽음)")
	out := fs.String("out", "", "출력 디렉터리")
	if err := fs.Parse(args); err != nil {
		return err
	}
	sc, err := parseScenario(*scPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	o, err := openObserver()
	if err != nil {
		return err
	}
	defer o.p.quit()
	res := runResult{Title: sc.Title, Scenario: filepath.Base(*scPath), Observer: o.pid, ObserverPids: []int{o.pid}, StartNs: time.Now().UnixNano()}
	for _, t := range sc.Watch {
		ti, err := o.table(t)
		if err != nil {
			return err
		}
		res.Tables = append(res.Tables, ti)
	}
	reply := func(format string, a ...any) { fmt.Printf(format+"\n", a...) }
	fail := func(err error) error {
		reply("error %s", strings.ReplaceAll(err.Error(), "\n", " "))
		return err
	}
	reply("ready %d", o.pid)

	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		f := strings.Fields(in.Text())
		if len(f) == 0 {
			continue
		}
		switch {
		case f[0] == "session" && len(f) == 3:
			pid, err := strconv.Atoi(f[2])
			if err != nil {
				return fail(fmt.Errorf("세션 pid %q", f[2]))
			}
			res.Sessions = append(res.Sessions, sessionRecord{Name: f[1], Pid: pid})
			o.pids = append(o.pids, pid)
			reply("ok")
		case f[0] == "capture" && len(f) == 3:
			n, err := strconv.Atoi(f[1])
			if err != nil {
				return fail(fmt.Errorf("멈춤 자리 순번 %q", f[1]))
			}
			c, err := o.capture(n, sc.Watch)
			if err != nil {
				return fail(err)
			}
			c.Label = f[2]
			res.Captures = append(res.Captures, c)
			reply("ok %d", c.Ns)
		case f[0] == "waiters" && len(f) == 1:
			w, err := o.lockWaiters()
			if err != nil {
				return fail(err)
			}
			var pids []string
			for _, s := range res.Sessions {
				if w[s.Pid] {
					pids = append(pids, strconv.Itoa(s.Pid))
				}
			}
			reply("waiters %s", strings.Join(pids, " "))
		case f[0] == "end" && len(f) == 1:
			res.EndNs = time.Now().UnixNano()
			b, err := json.MarshalIndent(res, "", " ")
			if err != nil {
				return fail(err)
			}
			if err := os.WriteFile(filepath.Join(*out, "steps.json"), b, 0o644); err != nil {
				return fail(err)
			}
			reply("ok")
			return nil
		default:
			return fail(fmt.Errorf("모르는 명령 %q", in.Text()))
		}
	}
	if err := in.Err(); err != nil {
		return err
	}
	return fmt.Errorf("observe: end 없이 입력이 끝났습니다")
}
