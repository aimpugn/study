// Command pgsql-atlas 는 pgsql 주제의 기존 기록을 공통 형식(atlas v0)으로 옮겨 파일로 씁니다.
//
// 지금은 gdb 단계 기록(E-05), 요청의 여정(concept-3·4), 트랜잭션 실험의 기록(concept-5~8)을 옮깁니다. 적재 결과는 다시 만들 수 있어 저장소에 넣지 않고,
// 기본으로 실습 폴더의 scripts/.lab-tmp/atlas/ 아래에 장면별로 씁니다(설계 메모 10).
//
//	cd open_source_internals/workbench && go run ./topics/pgsql/cmd/pgsql-atlas
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/gdbtrace"
	"study/open_source_internals/workbench/topics/pgsql/journey"
	"study/open_source_internals/workbench/topics/pgsql/txlab"
)

// main 함수는 인자를 읽고 옮기기를 돌린 뒤, 실패하면 이유를 표준 오류에 쓰고 종료 코드 1로 끝냅니다.
func main() {
	lab := flag.String("lab", "../database/postgresql", "실습 폴더(workbench.json과 evidence/가 있는 곳)")
	design := flag.String("design", "design", "화면 파일(concept-3~8.html)이 있는 폴더")
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

// run 함수는 gdb 기록, 여정 질의, 트랜잭션 기록을 모두 옮기고 형식 검사를 통과한 것만 <out>/<run id>.json으로 씁니다.
//
// 기록 하나라도 적재나 형식 검사에 실패하면 쓰지 않고 그 오류를 돌려줍니다. 반쯤 쓴 결과를 남기지 않기 위해서입니다.
func run(lab, design, out string) error {
	// 소스 판은 실습의 workbench.json이 정한 sourceRoot입니다.
	var subject struct {
		SourceRoot string `json:"sourceRoot"`
	}
	raw, err := os.ReadFile(filepath.Join(lab, "workbench.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &subject); err != nil || subject.SourceRoot == "" {
		return fmt.Errorf("workbench.json의 sourceRoot를 읽지 못함: %v", err)
	}

	// 모두 적재하고 검사한 뒤에 씁니다.
	paths, err := filepath.Glob(filepath.Join(lab, "evidence", "*.trace"))
	if err != nil {
		return err
	}
	runs := map[string]*atlas.Run{}
	// 형식 검사에는 같은 시계·같은 장소의 구간이 어긋나 겹치지 않는지(Perfetto의 넘침 줄)도 듭니다(설계 메모 7의 검사 2).
	// run id가 겹치면 뒤의 것이 앞의 파일을 덮으므로 오류로 봅니다.
	add := func(r *atlas.Run) error {
		if errs := atlas.Validate(r); len(errs) != 0 {
			return fmt.Errorf("%s: 형식 검사 %v", r.ID, errs)
		}
		if ov := atlas.CrossingSpans(r); len(ov) != 0 {
			return fmt.Errorf("%s: 어긋나 겹치는 구간 %d쌍, 첫 쌍 %+v", r.ID, len(ov), ov[0])
		}
		if _, dup := runs[r.ID]; dup {
			return fmt.Errorf("run id %s가 겹침", r.ID)
		}
		runs[r.ID] = r
		return nil
	}
	for _, p := range paths {
		text, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(p), ".trace")
		watch, err := gdbtrace.WatchFor(name, lab)
		if err != nil {
			return err
		}
		res, err := gdbtrace.Import(name, string(text), watch, subject.SourceRoot)
		if err != nil {
			return err
		}
		if err := add(res.Run); err != nil {
			return err
		}
	}

	// 여정: concept-4의 질의마다, concept-3에 같은 기록이 있으면 두 번째 해석으로 함께 옮깁니다.
	c4html, err := os.ReadFile(filepath.Join(design, "concept-4.html"))
	if err != nil {
		return err
	}
	c3html, err := os.ReadFile(filepath.Join(design, "concept-3.html"))
	if err != nil {
		return err
	}
	queries, err := journey.ReadConcept4(string(c4html))
	if err != nil {
		return err
	}
	c3, err := journey.ReadConcept3(string(c3html))
	if err != nil {
		return err
	}
	for _, raw := range queries {
		q, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("QUERIES의 원소가 객체가 아님")
		}
		c3data, _ := c3[fmt.Sprint(q["key"])].(map[string]any)
		res, err := journey.Import(q, c3data, subject.SourceRoot)
		if err != nil {
			return err
		}
		if err := add(res.Run); err != nil {
			return err
		}
	}

	// 트랜잭션 실험: 화면 넷의 RUNS에서 기록 한 벌씩 옮깁니다. 열쇠 차례로 옮겨 오류가 나면 늘 같은 기록에서 멈춥니다.
	for _, name := range []string{"concept-5-tx.html", "concept-6-locks.html", "concept-7-durability.html", "concept-8-serializable.html"} {
		html, err := os.ReadFile(filepath.Join(design, name))
		if err != nil {
			return err
		}
		txRuns, err := txlab.ReadRuns(string(html))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		keys := make([]string, 0, len(txRuns))
		for k := range txRuns {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v, ok := txRuns[k].(map[string]any)
			if !ok {
				return fmt.Errorf("%s: RUNS[%q]가 객체가 아님", name, k)
			}
			res, err := txlab.Import(k, v, subject.SourceRoot)
			if err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
			if err := add(res.Run); err != nil {
				return err
			}
		}
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
