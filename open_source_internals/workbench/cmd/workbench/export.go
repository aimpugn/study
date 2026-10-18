package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"study/open_source_internals/workbench/internal/replay"
	"study/open_source_internals/workbench/internal/source"
	"study/open_source_internals/workbench/internal/subject"
	"study/open_source_internals/workbench/internal/trace"
	"study/open_source_internals/workbench/web"
)

// export 함수는 저장된 기록 파일 하나를 재생본 HTML로 만듭니다.
//
// 소스는 실습 컨테이너가 아니라 이미지에서 직접 읽으므로(docker run --rm --network none), 컨테이너가 꺼져 있어도 됩니다.
// -sql이 없으면 첫 단계의 query_string 값에서 SQL을 꺼내고, 꺼낼 수 없으면 오류로 멈춥니다.
func export(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	subjPath := fs.String("subject", "", "주제 정의 workbench.json의 경로")
	tracePath := fs.String("trace", "", "run-trace.sh가 남긴 기록 파일")
	watchID := fs.String("watch", "", "그 기록에 쓴 감시 목록 id(workbench.json의 watches)")
	out := fs.String("out", "", "만들 재생본 HTML 경로")
	sql := fs.String("sql", "", "기록한 SQL. 없으면 첫 단계의 query_string에서 꺼냅니다")
	guide := fs.String("guide", "", "재생본 위치 기준으로, 이 실행을 설명하는 문서의 상대 링크")
	if err := parse(fs, args, map[string]*string{"subject": subjPath, "trace": tracePath, "watch": watchID, "out": out}); err != nil {
		return err
	}

	subj, err := subject.Load(*subjPath)
	if err != nil {
		return err
	}
	watch, ok := subj.Watch(*watchID)
	if !ok {
		return fmt.Errorf("주제 %s에 감시 목록 %q가 없습니다", subj.ID, *watchID)
	}
	exprs, err := readWatch(filepath.Join(subj.Dir, watch.File))
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(*tracePath)
	if err != nil {
		return fmt.Errorf("기록 파일 읽기: %w", err)
	}
	rec, err := trace.Parse(bytes.NewReader(raw), exprs)
	if err != nil {
		return fmt.Errorf("기록 파일 %s: %w", *tracePath, err)
	}
	if len(rec.Steps) == 0 {
		return fmt.Errorf("기록 파일 %s에 STOP 줄이 없습니다", *tracePath)
	}

	// SQL을 따로 받지 않았으면 기록 자체에서 꺼내, 화면의 SQL이 기록과 어긋나지 않게 합니다.
	if *sql == "" {
		q, truncated, ok := firstQuery(rec)
		if !ok || truncated {
			return fmt.Errorf("첫 단계의 query_string에서 SQL을 꺼내지 못했습니다. -sql로 주십시오")
		}
		*sql = q
	}

	// 출처에는 호스트의 절대경로 대신 주제 폴더 기준 경로를 씁니다.
	origin := filepath.Base(*tracePath)
	if abs, err := filepath.Abs(*tracePath); err == nil {
		if rel, err := filepath.Rel(subj.Dir, abs); err == nil && !strings.HasPrefix(rel, "..") {
			origin = filepath.ToSlash(rel)
		}
	}
	meta := replay.Meta{
		SQL: *sql, WatchID: watch.ID, WatchLabel: watch.Label,
		Origin: fmt.Sprintf("%s 실습의 %s", subj.Title, origin), Guide: *guide,
	}

	reader := &source.Docker{Command: []string{"docker", "run", "--rm", "--network", "none", subj.Image, "cat"}, Root: subj.SourceRoot}
	bundle, err := replay.Build(ctx, rec, subj, reader, meta)
	if err != nil {
		return err
	}
	page, err := web.Standalone(bundle)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*out, page, 0o644); err != nil {
		return fmt.Errorf("재생본 쓰기: %w", err)
	}
	files, lines := bundle.LineCount()
	fmt.Printf("재생본 %s: 단계 %d개, 소스 파일 %d개, 소스 줄 %d개, %d바이트\n", *out, len(bundle.Steps), len(files), lines, len(page))
	return nil
}

// readWatch 함수는 감시 목록 파일을 읽습니다.
func readWatch(path string) (trace.Exprs, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("감시 목록 열기: %w", err)
	}
	defer f.Close()
	return trace.ParseWatch(f)
}

// firstQuery 함수는 기록의 첫 단계가 exec_simple_query이면 그 query_string 값을 문자열로 꺼냅니다.
func firstQuery(rec trace.Record) (q string, truncated, ok bool) {
	if len(rec.Steps) == 0 || rec.Steps[0].Function != "exec_simple_query" {
		return "", false, false
	}
	for _, v := range rec.Steps[0].Values {
		if v.Expr == "query_string" {
			return trace.CString(v.Text)
		}
	}
	return "", false, false
}

// bundleScript 값은 재생본 안의 자료 스크립트입니다. web.Standalone이 이 꼴로 씁니다.
var bundleScript = regexp.MustCompile(`<script id="bundle" type="application/json">(.*?)</script>`)

// check 함수는 재생본에 담긴 모든 소스 줄을 이미지 안의 소스와 대조합니다(판정 기준 B2의 첫 단계).
//
// 재생본을 만들 때는 파일 전체를 cat으로 읽어 나눴으므로, 대조는 다른 방법으로 합니다.
// 파일마다 sed -n '줄번호{=;p}'로 그 줄만 다시 읽어, 줄 번호를 잘못 세는 실수가 두 경로에 함께 숨지 않게 합니다.
// 다른 줄이 하나라도 있으면 그 목록을 보이고 오류를 돌려줍니다.
func check(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	subjPath := fs.String("subject", "", "주제 정의 workbench.json의 경로")
	replayPath := fs.String("replay", "", "대조할 재생본 HTML")
	if err := parse(fs, args, map[string]*string{"subject": subjPath, "replay": replayPath}); err != nil {
		return err
	}
	subj, err := subject.Load(*subjPath)
	if err != nil {
		return err
	}
	page, err := os.ReadFile(*replayPath)
	if err != nil {
		return fmt.Errorf("재생본 읽기: %w", err)
	}
	m := bundleScript.FindSubmatch(page)
	if m == nil {
		return fmt.Errorf("재생본 %s에 자료 스크립트가 없습니다", *replayPath)
	}
	var bundle replay.Bundle
	if err := json.Unmarshal(m[1], &bundle); err != nil {
		return fmt.Errorf("재생본 자료 해석: %w", err)
	}

	// 컨테이너 안에서 실행할 셸 스크립트 하나를 만듭니다. 경로는 ValidatePath를 통과한 것만 넣어 따옴표 문제를 없앱니다.
	paths := make([]string, 0, len(bundle.Files))
	for p := range bundle.Files {
		if err := source.ValidatePath(p); err != nil {
			return err
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var script strings.Builder
	total := 0
	for _, p := range paths {
		nums := make([]int, 0, len(bundle.Files[p]))
		for n := range bundle.Files[p] {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		total += len(nums)
		fmt.Fprintf(&script, "echo '@@FILE %s'\nsed -n", p)
		for _, n := range nums {
			fmt.Fprintf(&script, " -e '%d{=;p;}'", n)
		}
		fmt.Fprintf(&script, " '%s/%s'\n", strings.TrimSuffix(subj.SourceRoot, "/"), p)
	}
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--network", "none", subj.Image, "sh", "-c", script.String())
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("이미지에서 소스 줄 읽기: %w", err)
	}

	// 출력은 "@@FILE 경로" 뒤에 "줄 번호", "그 줄" 두 줄씩입니다.
	got := map[string]map[int]string{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var file string
	for sc.Scan() {
		line := sc.Text()
		if rest, ok := strings.CutPrefix(line, "@@FILE "); ok {
			file = rest
			got[file] = map[int]string{}
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil || !sc.Scan() {
			return fmt.Errorf("sed 출력의 형식이 예상과 다릅니다(%s 근처)", file)
		}
		got[file][n] = sc.Text()
	}

	var bad []string
	for _, p := range paths {
		for n, text := range bundle.Files[p] {
			if g, ok := got[p][n]; !ok || g != text {
				bad = append(bad, fmt.Sprintf("%s:%d", p, n))
			}
		}
	}
	sort.Strings(bad)
	fmt.Printf("소스 대조: 파일 %d개, 줄 %d개, 다른 줄 %d개\n", len(paths), total, len(bad))
	if len(bad) > 0 {
		return fmt.Errorf("재생본의 소스 줄이 이미지와 다릅니다: %s", strings.Join(bad, ", "))
	}
	return nil
}
