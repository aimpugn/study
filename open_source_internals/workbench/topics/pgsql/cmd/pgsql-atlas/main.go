// Command pgsql-atlas 는 pgsql 주제의 기존 기록을 공통 형식(atlas v0)으로 옮겨 파일로 씁니다.
//
// gdb 단계 기록(E-05), 요청의 여정(concept-3·4), 트랜잭션 실험의 기록(concept-5~8), Spring 전파 실험의 기록(concept-9, E-03)을 옮깁니다. 적재 결과는 다시 만들 수 있어 저장소에 넣지 않고,
// 기본으로 실습 폴더의 scripts/.lab-tmp/atlas/ 아래에 장면별로 씁니다(설계 메모 10).
//
//	cd open_source_internals/workbench && go run ./topics/pgsql/cmd/pgsql-atlas
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/collect"
)

// main 함수는 인자를 읽고 옮기기를 돌린 뒤, 실패하면 이유를 표준 오류에 쓰고 종료 코드 1로 끝냅니다.
func main() {
	lab := flag.String("lab", "../database/postgresql", "실습 폴더(workbench.json과 evidence/가 있는 곳)")
	design := flag.String("design", "design", "화면 파일(concept-3~9)이 있는 폴더")
	out := flag.String("out", "", "적재 결과를 쓸 폴더(비우면 <lab>/scripts/.lab-tmp/atlas)")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "모르는 인자 %v\n", flag.Args())
		os.Exit(2)
	}
	if *out == "" {
		*out = filepath.Join(*lab, "scripts", ".lab-tmp", "atlas")
	}
	if err := run(*lab, *design, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run 함수는 기록 넷을 모두 옮기고, 모두 형식 검사를 통과했을 때만 <out>/<run id>.json으로 씁니다.
//
// 기록 하나라도 적재나 형식 검사에 실패하면 쓰지 않고 그 오류를 돌려줍니다. 반쯤 쓴 결과를 남기지 않기 위해서입니다.
func run(lab, design, out string) error {
	runs, err := collect.All(lab, design)
	if err != nil {
		return err
	}

	// 모두 통과한 뒤에 run id("장면/이름")의 경로로 씁니다.
	for id, r := range runs {
		path := filepath.Join(out, filepath.FromSlash(id)+".json")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := atlas.WriteFile(path, r); err != nil {
			return err
		}
	}
	fmt.Printf("기록 %d개를 %s에 씀\n", len(runs), out)
	return nil
}
