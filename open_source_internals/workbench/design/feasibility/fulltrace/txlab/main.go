// txlab 은 여러 세션이 정해진 순서로 움직이는 트랜잭션 시나리오를 실제 psql로 실행하고 기록합니다.
// 탐색용 코드입니다(2026-10-14, 트랜잭션 관측 실험 1~5). 작업대의 하위 명령으로 옮길지는 실험 뒤에 정합니다.
//
// 같은 프로그램이 두 곳에서 돕니다.
//
//   - 맥(record): 자기 자신을 linux/arm64로 빌드해 실습 컨테이너에 넣고, OrbStack 머신 tracelab에 uprobe를 걸고,
//     컨테이너 안의 serve·run·stop을 차례로 부른 뒤 기록을 모읍니다.
//   - 컨테이너 안(serve, run, stop): 기록 빌드로 서버를 띄우고(strace + 함수 기록), 시나리오의 세션마다 psql을 하나씩 띄워
//     문장을 보내고, 단계마다 관찰용 접속으로 페이지·세션 상태·락을 읽습니다.
//
// 사용: go run . record -scenario scenarios/1-rc-read.txt -out <출력 디렉터리>
// 결과: <출력>/steps.json(단계와 관찰), uprobe.trace(인자·반환값), all.strace, trace/<pid>.bin, postgres(기록 빌드)
// 화면: go run . view -tmpl <템플릿> -out <HTML> 1-rc-read/a=<decode scenario의 JSON> ...(해석은 decode가 맡음)
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "사용: txlab record|serve|run|stop|uprobes|view|relay|observe [옵션]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "record":
		err = cmdRecord(os.Args[2:])
	case "serve":
		err = cmdServe(os.Args[2:])
	case "run":
		err = cmdRun(os.Args[2:])
	case "stop":
		err = cmdStop(os.Args[2:])
	case "uprobes":
		err = cmdUprobes(os.Args[2:])
	case "view":
		err = cmdView(os.Args[2:])
	case "relay":
		err = cmdRelay(os.Args[2:])
	case "observe":
		err = cmdObserve(os.Args[2:])
	default:
		err = fmt.Errorf("모르는 하위 명령 %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "txlab:", err)
		os.Exit(1)
	}
}
