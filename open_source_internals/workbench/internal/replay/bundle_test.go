package replay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"study/open_source_internals/workbench/internal/subject"
	"study/open_source_internals/workbench/internal/trace"
)

// fakeReader 값은 시험용 소스 트리입니다. 경로마다 파일 내용을 들고 있고, 없는 경로는 오류로 답합니다.
type fakeReader map[string][]string

// Lines 함수는 path의 줄 목록을 돌려주고, 없으면 오류를 돌려줍니다.
func (f fakeReader) Lines(_ context.Context, path string) ([]string, error) {
	if lines, ok := f[path]; ok {
		return lines, nil
	}
	return nil, errors.New("없는 파일")
}

// 단계와 호출 스택의 위치마다 분류·세로줄·구간이 붙고, 구간의 줄만 Files에 담기는지 확인합니다.
func TestBuildAttachesLayersWindowsAndOnlyNeededLines(t *testing.T) {
	// Given: 40줄짜리 파일 하나(18번 빈 줄, 19번 반환형, 20번 정의)와, 30번 줄에 멈춘 단계 하나(스택 두 칸, 하나는 위치를 모름)
	file := make([]string, 40)
	for i := range file {
		file[i] = "line"
	}
	file[17] = ""
	file[18] = "static void"
	file[19] = "f(void)"
	reader := fakeReader{"src/a/x.c": file}
	subj := subject.Subject{ID: "s", Layers: []subject.Layer{{Prefix: "src/a/", Label: "가"}}}
	rec := trace.Record{Steps: []trace.Step{{
		Seq: 1, Function: "f", Loc: trace.Location{Path: "src/a/x.c", Line: 30},
		Stack: []trace.Frame{
			{Depth: 0, Function: "f", Loc: trace.Location{Path: "src/a/x.c", Line: 30}},
			{Depth: 1, Function: "??"},
		},
	}}}

	// When
	b, err := Build(context.Background(), rec, subj, reader, Meta{SQL: "SELECT 1;"})
	// Then: 분류 "가"(세로줄 0), 정의 줄 20, 위치 없는 칸은 분류 없음, 파일의 줄은 구간(반환형 19번~40번) 안의 것만
	if err != nil {
		t.Fatal(err)
	}
	p := b.Steps[0].Place
	if p.Layer != "가" || p.Lane != 0 || p.Window.DefLine != 20 {
		t.Fatalf("단계의 분류·구간이 다릅니다: %+v", p)
	}
	if got := b.Steps[0].Stack[1]; got.Layer != subject.Unclassified || len(got.Window.Ranges) != 0 {
		t.Fatalf("위치를 모르는 칸은 분류 없음이고 구간이 없어야 합니다: %+v", got)
	}
	lines := b.Files["src/a/x.c"]
	if _, ok := lines[18]; ok || len(lines) != 40-19+1 {
		t.Fatalf("구간(19~40번 줄)의 줄만 담겨야 합니다: %d줄", len(lines))
	}
}

// 소스 파일을 읽지 못하면 일부만 담은 자료를 만들지 않고 오류를 돌려주는지 확인합니다.
func TestBuildFailsWhenSourceIsMissing(t *testing.T) {
	// Given
	rec := trace.Record{Steps: []trace.Step{{Seq: 1, Function: "f", Loc: trace.Location{Path: "src/none.c", Line: 3}}}}

	// When
	_, err := Build(context.Background(), rec, subject.Subject{}, fakeReader{}, Meta{})

	// Then
	if err == nil || !strings.Contains(err.Error(), "1번 단계") {
		t.Fatalf("어느 단계에서 실패했는지 밝힌 오류여야 합니다: %v", err)
	}
}
