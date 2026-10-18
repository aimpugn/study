package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// cmdView 는 해석기(decode scenario)가 만든 JSON 여러 벌을 화면 템플릿 하나에 넣어 HTML 파일 하나를 만듭니다.
//
// Arguments: -tmpl은 템플릿 경로(자리 표시 "__RUNS__"를 정확히 한 번 가짐), -out은 쓸 HTML 경로, 나머지 인자는
// "이름=JSON 경로"입니다. 이름은 "시나리오/실행"(예: 1-rc-read/a)이고, 화면은 시나리오마다 실행을 고르는 단추를 만듭니다.
// Errors: 템플릿의 자리 표시가 한 번이 아니거나, JSON을 읽거나 풀 수 없거나, 이름에 "/"가 없으면 오류를 돌려줍니다.
//
// JSON은 한 번 풀었다가 다시 씁니다. encoding/json이 <, >, &를 < 꼴로 바꾸므로, 소스 발췌에 든 "</script>" 같은
// 문자열이 HTML의 script 요소를 닫지 않습니다.
func cmdView(args []string) error {
	fs := flag.NewFlagSet("view", flag.ContinueOnError)
	tmpl := fs.String("tmpl", "", "화면 템플릿")
	out := fs.String("out", "", "쓸 HTML")
	if err := fs.Parse(args); err != nil {
		return err
	}
	b, err := os.ReadFile(*tmpl)
	if err != nil {
		return err
	}
	if n := strings.Count(string(b), "__RUNS__"); n != 1 {
		return fmt.Errorf("템플릿 %s에 __RUNS__가 %d번 있습니다(1번이어야 함)", *tmpl, n)
	}

	// 이름 순서대로 넣어 같은 입력이면 같은 HTML이 나오게 합니다.
	runs := map[string]any{}
	var names []string
	for _, a := range fs.Args() {
		name, path, ok := strings.Cut(a, "=")
		if !ok || !strings.Contains(name, "/") {
			return fmt.Errorf("인자 %q는 시나리오/실행=경로 꼴이어야 합니다", a)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		runs[name] = v
		names = append(names, name)
	}
	sort.Strings(names)
	data, err := json.Marshal(runs)
	if err != nil {
		return err
	}
	html := strings.Replace(string(b), "__RUNS__", string(data), 1)
	if err := os.WriteFile(*out, []byte(html), 0o644); err != nil {
		return err
	}
	fmt.Printf("%s: %d벌, %d KB\n", *out, len(names), len(html)>>10)
	return nil
}
