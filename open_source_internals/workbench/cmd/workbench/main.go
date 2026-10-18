// Command workbench 는 오픈소스 내부 작업대의 명령입니다.
//
//	workbench serve  -subject <workbench.json> [-addr 127.0.0.1:8742] [-timeout 90s]
//	workbench export -subject <workbench.json> -trace <기록 파일> -watch <감시 목록 id> -out <재생본.html> [-sql <SQL>] [-guide <상대 링크>]
//	workbench check  -subject <workbench.json> -replay <재생본.html>
//
// serve는 브라우저에서 SQL을 실행하고 추적하는 로컬 서버를 엽니다. export는 저장된 기록을 서버 없이 여는 재생본 한 파일로 만들고,
// check는 재생본에 담긴 모든 소스 줄을 이미지 안의 소스와 다른 방법(sed)으로 다시 읽어 대조합니다.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// main 함수는 첫 인자로 하위 명령을 골라 실행하고, 실패하면 이유를 표준 오류에 쓰고 종료 코드 1로 끝냅니다.
// 하위 명령이 없거나 모르는 이름이면 사용법을 보이고 종료 코드 2로 끝냅니다.
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch os.Args[1] {
	case "serve":
		err = serve(ctx, os.Args[2:])
	case "export":
		err = export(ctx, os.Args[2:])
	case "check":
		err = check(ctx, os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "workbench:", err)
		os.Exit(1)
	}
}

// usage 함수는 하위 명령 목록을 표준 오류로 보입니다.
func usage() {
	fmt.Fprintln(os.Stderr, "사용: workbench serve|export|check [옵션]. 각 명령의 옵션은 workbench <명령> -h로 봅니다.")
}

// parse 함수는 하위 명령의 옵션을 읽고, 빈 필수 옵션이 있으면 그 이름을 밝힌 오류를 돌려줍니다.
//
// 알 수 없는 옵션이나 값의 형식 오류는 flag 패키지가 오류로 돌려줍니다. 위치 인자가 남아도 오류입니다.
// 오타가 난 옵션이 조용히 무시되어 다른 조건으로 실행되는 일을 막기 위해서입니다.
func parse(fs *flag.FlagSet, args []string, required map[string]*string) error {
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("%s: 알 수 없는 인자 %q", fs.Name(), fs.Arg(0))
	}
	for name, v := range required {
		if *v == "" {
			return fmt.Errorf("%s: -%s 옵션이 필요합니다", fs.Name(), name)
		}
	}
	return nil
}
