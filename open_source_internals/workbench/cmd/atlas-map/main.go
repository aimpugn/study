// Command atlas-map 은 공통 형식(atlas v0)의 기록 폴더와 주제의 지도 자료 파일로 지도 화면 파일 하나를 씁니다.
//
// 주제를 모르는 플랫폼 명령입니다. 주제는 자료 파일(주제 폴더의 atlasmap/map.json)만 넘기고, 기록은 공통 형식 폴더에서만 읽습니다.
//
//	go run ./cmd/atlas-map -in <기록 폴더> -spec <주제 폴더>/atlasmap/map.json -out <화면 파일>
//
// 주제마다의 예는 그 주제의 README에 있습니다.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"study/open_source_internals/workbench/internal/atlas/mapview"
)

// main 함수는 인자를 읽고 화면 파일을 쓴 뒤, 실패하면 이유를 표준 오류에 쓰고 종료 코드 1로 끝냅니다.
func main() {
	in := flag.String("in", "", "공통 형식 기록 폴더(<run id>.json이 있는 곳)")
	spec := flag.String("spec", "", "주제의 지도 자료 파일")
	out := flag.String("out", "", "쓸 화면 파일(.html)")
	flag.Parse()
	if flag.NArg() != 0 || *in == "" || *spec == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "쓰는 법: atlas-map -in <기록 폴더> -spec <자료 파일> -out <화면 파일>")
		os.Exit(2)
	}
	if err := run(*in, *spec, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run 함수는 자료 파일을 읽고 기록 폴더로 보기 모형을 만든 뒤 화면 파일 하나로 씁니다. 반쯤 쓴 파일을 남기지 않게 다 만든 뒤에 씁니다.
func run(in, specPath, out string) error {
	spec, err := mapview.LoadSpec(specPath)
	if err != nil {
		return err
	}
	v, err := mapview.Build(mapview.DirStore{Dir: in}, spec, filepath.Dir(specPath))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := mapview.WriteHTML(&buf, v); err != nil {
		return err
	}
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Printf("경로 %d개(무리 %d)를 %s에 씀(%d KB)\n", len(v.Routes), len(v.Groups), out, buf.Len()/1024)
	return nil
}
