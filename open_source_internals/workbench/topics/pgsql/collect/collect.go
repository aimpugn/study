// Package collect 는 pgsql 주제의 기존 기록 넷(gdb 단계 기록, 요청의 여정, 트랜잭션 실험, Spring 전파 실험)을 공통 형식(atlas v0)으로 옮겨 모읍니다.
//
// 실행 명령 pgsql-atlas가 이것으로 파일을 쓰고, 지도 화면의 전제 시험이 같은 함수로 적재 결과를 얻습니다.
package collect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/gdbtrace"
	"study/open_source_internals/workbench/topics/pgsql/journey"
	"study/open_source_internals/workbench/topics/pgsql/spring"
	"study/open_source_internals/workbench/topics/pgsql/txlab"
)

// All 함수는 gdb 기록, 여정 질의, 트랜잭션 기록, Spring 기록을 모두 공통 형식으로 옮기고, 형식 검사를 통과한 run을 id마다 돌려줍니다.
//
// lab은 실습 폴더(workbench.json과 evidence/가 있는 곳), design은 화면 파일(concept-3~9)이 있는 폴더입니다.
// 기록 하나라도 적재나 형식 검사에 실패하거나 run id가 겹치면 아무것도 돌려주지 않고 그 오류를 돌려줍니다.
func All(lab, design string) (map[string]*atlas.Run, error) {
	// 소스 판은 실습의 workbench.json이 정한 sourceRoot입니다.
	var subject struct {
		SourceRoot string `json:"sourceRoot"`
	}
	raw, err := os.ReadFile(filepath.Join(lab, "workbench.json"))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &subject); err != nil || subject.SourceRoot == "" {
		return nil, fmt.Errorf("workbench.json의 sourceRoot를 읽지 못함: %v", err)
	}

	// 모두 적재하고 검사한 뒤에 씁니다.
	paths, err := filepath.Glob(filepath.Join(lab, "evidence", "*.trace"))
	if err != nil {
		return nil, err
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
			return nil, err
		}
		name := strings.TrimSuffix(filepath.Base(p), ".trace")
		watch, err := gdbtrace.WatchFor(name, lab)
		if err != nil {
			return nil, err
		}
		res, err := gdbtrace.Import(name, string(text), watch, subject.SourceRoot)
		if err != nil {
			return nil, err
		}
		if err := add(res.Run); err != nil {
			return nil, err
		}
	}

	// 여정: concept-4의 질의마다, concept-3에 같은 기록이 있으면 두 번째 해석으로 함께 옮깁니다.
	c4html, err := os.ReadFile(filepath.Join(design, "concept-4.html"))
	if err != nil {
		return nil, err
	}
	c3html, err := os.ReadFile(filepath.Join(design, "concept-3.html"))
	if err != nil {
		return nil, err
	}
	queries, err := journey.ReadConcept4(string(c4html))
	if err != nil {
		return nil, err
	}
	c3, err := journey.ReadConcept3(string(c3html))
	if err != nil {
		return nil, err
	}
	for _, raw := range queries {
		q, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("QUERIES의 원소가 객체가 아님")
		}
		c3data, _ := c3[fmt.Sprint(q["key"])].(map[string]any)
		res, err := journey.Import(q, c3data, subject.SourceRoot)
		if err != nil {
			return nil, err
		}
		if err := add(res.Run); err != nil {
			return nil, err
		}
	}

	// 트랜잭션 실험: 화면 넷의 RUNS에서 기록 한 벌씩 옮깁니다. 열쇠 차례로 옮겨 오류가 나면 늘 같은 기록에서 멈춥니다.
	for _, name := range []string{"concept-5-tx.html", "concept-6-locks.html", "concept-7-durability.html", "concept-8-serializable.html"} {
		html, err := os.ReadFile(filepath.Join(design, name))
		if err != nil {
			return nil, err
		}
		txRuns, err := txlab.ReadRuns(string(html))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		keys := make([]string, 0, len(txRuns))
		for k := range txRuns {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v, ok := txRuns[k].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s: RUNS[%q]가 객체가 아님", name, k)
			}
			res, err := txlab.Import(k, v, subject.SourceRoot)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
			if err := add(res.Run); err != nil {
				return nil, err
			}
		}
	}

	// Spring 전파 실험: concept-9의 RUNS(E-03과 값이 같아 한 번만 옮김, 재고 조사 2.1절)에서 기록 한 벌씩 옮깁니다.
	html, err := os.ReadFile(filepath.Join(design, "concept-9-spring.html"))
	if err != nil {
		return nil, err
	}
	springRuns, err := spring.ReadRuns(string(html))
	if err != nil {
		return nil, fmt.Errorf("concept-9-spring.html: %w", err)
	}
	keys := make([]string, 0, len(springRuns))
	for k := range springRuns {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v, ok := springRuns[k].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("concept-9-spring.html: RUNS[%q]가 객체가 아님", k)
		}
		res, err := spring.Import(k, v, subject.SourceRoot)
		if err != nil {
			return nil, err
		}
		if err := add(res.Run); err != nil {
			return nil, err
		}
	}
	return runs, nil
}
