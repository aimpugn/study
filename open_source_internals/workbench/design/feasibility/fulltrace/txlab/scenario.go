package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// scenario 는 시나리오 파일 하나입니다. Setup은 기록하지 않는 준비 단계의 SQL, Watch는 단계마다 페이지를 읽을 표,
// Steps는 세션 이름과 그 세션에 보낼 SQL의 순서입니다. Pauses는 단계 사이의 쉼이고, 열쇠는 쉼 뒤에 올 단계의 순번(0부터)입니다.
// Kills는 그 순번의 단계 앞에서 백엔드를 SIGKILL로 죽일 세션입니다. 같은 순번에 둘 다 있으면 쉼, kill, 단계 순서로 실행합니다.
// 쉼과 kill을 단계로 세지 않는 것은, 해석기와 화면이 단계마다 세션과 SQL이 있다고 보기 때문입니다.
//
// Spring은 실험 5의 시나리오입니다. 값이 있으면 psql 세션 대신 맥의 Java 앱(spring/)이 그 이름의 장면을 돌리고, 단계는
// 앱이 알리는 멈춤 자리(checkpoint)가 됩니다. 그래서 Spring 시나리오에는 세션 줄·쉼·kill이 없습니다. Pool은 그 앱의 연결 풀 크기입니다.
// Blocked는 go 뒤 다음 알림이 이만큼 오지 않으면 앱이 기다리는 중으로 보고 한 번 더 찍는 때입니다(spring.go). 관찰 한 번이 0.7초쯤
// 걸리므로, 앱이 기다리는 것이 그보다 먼저 끝나는 장면은 이 값을 줄여 관찰이 그 끝과 겹치지 않게 합니다.
//
// Epoch는 세션마다 몇 번째 kill 뒤에 처음 쓰였는지입니다. 백엔드 하나가 신호로 죽으면 postmaster가 다른 백엔드를 모두 내리고
// 다시 뜨므로(postmaster.c HandleChildCrash), kill 앞의 접속은 그 뒤에 쓸 수 없습니다. 그래서 세션은 한 구간 안에서만 씁니다.
type scenario struct {
	Title    string
	Setup    []string
	Watch    []string
	Steps    []step
	Sessions []string
	Pauses   map[int]time.Duration
	Kills    map[int]string
	Epoch    map[string]int
	Spring   string
	Pool     int
	Blocked  time.Duration
}

// step 은 세션 하나에 보내는 SQL 한 줄입니다. 한 줄에 문장이 여럿이면 psql이 문장마다 따로 보냅니다.
type step struct {
	Session string
	SQL     string
}

// sessionLine 은 "S1: SQL" 꼴의 줄입니다. 세션 이름은 S로 시작합니다.
var sessionLine = regexp.MustCompile(`^(S\w*):\s*(.+)$`)

// parseScenario 는 시나리오 파일을 읽습니다.
//
// 형식: 첫 "# " 줄이 제목이고, 나머지 "#" 줄과 빈 줄은 건너뜁니다. "setup: SQL"은 준비, "watch: 표"는 관찰할 표,
// "S1: SQL"은 세션 S1의 단계, "pause: 1500ms"는 다음 단계를 보내기 전에 쉴 시간(Go의 time.ParseDuration 꼴),
// "kill: S1"은 다음 단계를 보내기 전에 세션 S1의 백엔드를 죽이는 줄입니다. 세션은 처음 나온 순서로 번호가 정해집니다.
// "spring: 장면"은 Java 앱이 돌릴 장면 이름이고, "pool: N"은 그 앱의 연결 풀 크기입니다(없으면 2). "blocked: 500ms"는
// 멈춘 채로 찍기까지 기다릴 시간입니다(없으면 1.5초).
//
// Errors: 파일을 읽지 못하거나, 모르는 꼴의 줄이나 읽을 수 없는 쉼 시간이 있거나, 단계가 하나도 없으면 오류를 돌려줍니다.
// Spring 시나리오는 거꾸로 세션 줄·쉼·kill이 하나라도 있으면 오류이고, 풀 크기가 1 이상의 정수가 아니거나 blocked가 양의
// 시간이 아니어도 오류입니다. Spring 시나리오가 아닌데 pool이나 blocked가 있으면 오류입니다.
// kill이 지금 구간에서 쓰인 적 없는 세션을 가리키거나, kill 앞에서 쓴 세션을 그 뒤에 다시 쓰거나, kill 뒤 같은 단계 앞에
// 쉼을 두면(실행 순서가 파일 순서와 달라짐) 오류입니다.
func parseScenario(path string) (*scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := &scenario{Pauses: map[int]time.Duration{}, Kills: map[int]string{}, Epoch: map[string]int{}}
	epoch, lastKill := 0, -1
	in := bufio.NewScanner(f)
	for n := 1; in.Scan(); n++ {
		line := strings.TrimSpace(in.Text())
		switch {
		case line == "":
		case strings.HasPrefix(line, "#"):
			if sc.Title == "" && strings.HasPrefix(line, "# ") {
				sc.Title = strings.TrimPrefix(line, "# ")
			}
		case strings.HasPrefix(line, "setup:"):
			sc.Setup = append(sc.Setup, strings.TrimSpace(strings.TrimPrefix(line, "setup:")))
		case strings.HasPrefix(line, "spring:"):
			sc.Spring = strings.TrimSpace(strings.TrimPrefix(line, "spring:"))
		case strings.HasPrefix(line, "pool:"):
			v := strings.TrimSpace(strings.TrimPrefix(line, "pool:"))
			size, err := strconv.Atoi(v)
			if err != nil || size < 1 {
				return nil, fmt.Errorf("%s:%d: 풀 크기 %q를 읽지 못했습니다", path, n, v)
			}
			sc.Pool = size
		case strings.HasPrefix(line, "blocked:"):
			d, err := time.ParseDuration(strings.TrimSpace(strings.TrimPrefix(line, "blocked:")))
			if err != nil || d <= 0 {
				return nil, fmt.Errorf("%s:%d: 기다릴 시간 %q를 읽지 못했습니다", path, n, line)
			}
			sc.Blocked = d
		case strings.HasPrefix(line, "watch:"):
			sc.Watch = append(sc.Watch, strings.TrimSpace(strings.TrimPrefix(line, "watch:")))
		case strings.HasPrefix(line, "pause:"):
			d, err := time.ParseDuration(strings.TrimSpace(strings.TrimPrefix(line, "pause:")))
			if err != nil || d <= 0 {
				return nil, fmt.Errorf("%s:%d: 쉼 시간 %q를 읽지 못했습니다", path, n, line)
			}
			if lastKill == len(sc.Steps) {
				return nil, fmt.Errorf("%s:%d: 쉼은 같은 단계 앞의 kill보다 앞에 둡니다(실행은 쉼, kill 순서)", path, n)
			}
			sc.Pauses[len(sc.Steps)] += d
		case strings.HasPrefix(line, "kill:"):
			name := strings.TrimSpace(strings.TrimPrefix(line, "kill:"))
			if e, ok := sc.Epoch[name]; !ok || e != epoch {
				return nil, fmt.Errorf("%s:%d: kill할 세션 %q가 이 구간에서 쓰인 적이 없습니다", path, n, name)
			}
			sc.Kills[len(sc.Steps)] = name
			epoch, lastKill = epoch+1, len(sc.Steps)
		default:
			m := sessionLine.FindStringSubmatch(line)
			if m == nil {
				return nil, fmt.Errorf("%s:%d: 모르는 줄 %q", path, n, line)
			}
			if e, ok := sc.Epoch[m[1]]; !ok {
				sc.Epoch[m[1]] = epoch
				sc.Sessions = append(sc.Sessions, m[1])
			} else if e != epoch {
				return nil, fmt.Errorf("%s:%d: 세션 %s는 kill 앞에서 썼으므로 그 뒤에 쓸 수 없습니다(서버가 다시 뜨며 접속이 모두 끊김)", path, n, m[1])
			}
			sc.Steps = append(sc.Steps, step{Session: m[1], SQL: m[2]})
		}
	}
	if err := in.Err(); err != nil {
		return nil, err
	}
	if sc.Spring != "" {
		if len(sc.Steps) > 0 || len(sc.Pauses) > 0 || len(sc.Kills) > 0 {
			return nil, fmt.Errorf("%s: Spring 시나리오에는 세션 줄·쉼·kill을 쓰지 않습니다(단계는 앱의 멈춤 자리)", path)
		}
		if sc.Pool == 0 {
			sc.Pool = 2
		}
		if sc.Blocked == 0 {
			sc.Blocked = 1500 * time.Millisecond
		}
		return sc, nil
	}
	if sc.Pool != 0 || sc.Blocked != 0 {
		return nil, fmt.Errorf("%s: pool과 blocked는 Spring 시나리오에만 씁니다", path)
	}
	if len(sc.Steps) == 0 {
		return nil, fmt.Errorf("%s: 단계가 없습니다", path)
	}
	return sc, nil
}
