// Package replay 는 실행 기록 하나를 화면이 그대로 그릴 수 있는 자료(Bundle)로 묶습니다.
//
// 실시간 실행의 응답과 누구나 여는 재생본이 같은 Bundle을 씁니다. 그래서 화면 코드는 자료가 어디서 왔는지 몰라도 됩니다.
package replay

import (
	"context"
	"fmt"
	"sort"

	"study/open_source_internals/workbench/internal/source"
	"study/open_source_internals/workbench/internal/subject"
	"study/open_source_internals/workbench/internal/trace"
)

// Lines 값은 소스 파일 하나에서 화면에 필요한 줄만 줄 번호로 담습니다.
type Lines map[int]string

// Place 값은 기록 속 위치 하나(멈춘 자리 또는 호출 스택의 한 칸)를 화면에 그리는 데 필요한 것입니다.
//
// Layer는 subject.Classify가 디렉터리로 붙인 분류이고, Lane은 그 분류의 세로줄 번호입니다.
// Window는 Files에 들어 있는 줄 가운데 이 위치에서 보여 줄 구간입니다. 위치를 모르면 Window가 비어 있습니다.
type Place struct {
	Function string         `json:"function"`
	Loc      trace.Location `json:"loc"`
	Layer    string         `json:"layer"`
	Lane     int            `json:"lane"`
	Window   source.Window  `json:"window"`
}

// Step 값은 기록의 단계 하나와, 그 단계와 호출 스택 각 칸의 Place입니다.
type Step struct {
	Seq       int           `json:"seq"`
	Place     Place         `json:"place"`
	Values    []trace.Value `json:"values,omitempty"`
	RawValues string        `json:"rawValues,omitempty"`
	Stack     []Place       `json:"stack,omitempty"`
}

// Meta 값은 기록이 어디서 왔는지를 화면에 밝히는 정보입니다.
//
// Origin은 "PostgreSQL 실습의 evidence/select-where-id-2.trace"나 "실시간 실행"처럼 사람이 읽는 출처이고, 호스트의 절대경로를 담지 않습니다.
// Guide는 이 실행을 설명하는 문서로 가는 상대 링크입니다. 재생본 파일의 위치를 기준으로 합니다.
type Meta struct {
	SQL        string `json:"sql"`
	WatchID    string `json:"watchId"`
	WatchLabel string `json:"watchLabel"`
	Origin     string `json:"origin"`
	Guide      string `json:"guide,omitempty"`
	RecordedAt string `json:"recordedAt,omitempty"`
}

// Bundle 값은 화면이 실행 한 번을 그리는 데 필요한 모든 자료입니다.
//
// Files에는 각 Place의 Window에 든 줄만 들어 있습니다. 재생본 크기를 줄이기 위해서이고, 화면은 이 범위 밖의 줄을 지어내지 않습니다.
type Bundle struct {
	SubjectID    string           `json:"subjectId"`
	SubjectTitle string           `json:"subjectTitle"`
	SourceRoot   string           `json:"sourceRoot"`
	SourceNotice string           `json:"sourceNotice"`
	Lanes        []subject.Lane   `json:"lanes"`
	Meta         Meta             `json:"meta"`
	Steps        []Step           `json:"steps"`
	Counts       []trace.Count    `json:"counts"`
	NoBreak      []string         `json:"noBreak,omitempty"`
	Output       []string         `json:"output,omitempty"`
	Files        map[string]Lines `json:"files"`
}

// LineReader 값은 소스 상대경로의 파일 전체를 줄 목록으로 돌려줍니다. source.Docker가 이 꼴을 따릅니다.
type LineReader interface {
	Lines(ctx context.Context, path string) ([]string, error)
}

// Build 함수는 기록 rec을 주제 subj의 층 분류와 reader가 읽은 소스로 묶어 Bundle을 만듭니다.
//
// 위치가 있는 모든 단계와 호출 스택 칸에 대해 소스 구간을 계산하고, 그 구간의 줄만 Files에 담습니다.
// 소스 파일 하나라도 읽지 못하면 그 파일을 밝힌 오류를 돌려줍니다. 일부 줄이 빠진 화면은 기록과 소스가 맞는지 판정할 수 없기 때문입니다.
func Build(ctx context.Context, rec trace.Record, subj subject.Subject, reader LineReader, meta Meta) (Bundle, error) {
	lanes := subj.Lanes()
	laneOf := map[string]int{}
	for _, l := range lanes {
		laneOf[l.Label] = l.Index
	}
	b := Bundle{
		SubjectID: subj.ID, SubjectTitle: subj.Title, SourceRoot: subj.SourceRoot, SourceNotice: subj.SourceNotice, Lanes: lanes, Meta: meta,
		Counts: rec.Counts, NoBreak: rec.NoBreak, Output: rec.Output, Files: map[string]Lines{},
	}

	// place 함수는 위치 하나의 분류와 구간을 계산하고, 구간의 줄을 b.Files에 모읍니다.
	place := func(fn string, loc trace.Location) (Place, error) {
		layer := subject.Unclassified
		p := Place{Function: fn, Loc: loc}
		if loc.Path == "" {
			p.Layer, p.Lane = layer, laneOf[layer]
			return p, nil
		}
		layer = subj.Classify(loc.Path)
		p.Layer, p.Lane = layer, laneOf[layer]
		lines, err := reader.Lines(ctx, loc.Path)
		if err != nil {
			return Place{}, err
		}
		p.Window = source.Compute(lines, fn, loc.Line)
		keep := b.Files[loc.Path]
		if keep == nil {
			keep = Lines{}
			b.Files[loc.Path] = keep
		}
		for _, r := range p.Window.Ranges {
			for n := r.From; n <= r.To; n++ {
				keep[n] = lines[n-1]
			}
		}
		return p, nil
	}

	for _, s := range rec.Steps {
		p, err := place(s.Function, s.Loc)
		if err != nil {
			return Bundle{}, fmt.Errorf("%d번 단계(%s): %w", s.Seq, s.Function, err)
		}
		step := Step{Seq: s.Seq, Place: p, Values: s.Values, RawValues: s.RawValues}
		for _, f := range s.Stack {
			fp, err := place(f.Function, f.Loc)
			if err != nil {
				return Bundle{}, fmt.Errorf("%d번 단계의 호출 스택 %d번 칸(%s): %w", s.Seq, f.Depth, f.Function, err)
			}
			step.Stack = append(step.Stack, fp)
		}
		b.Steps = append(b.Steps, step)
	}
	return b, nil
}

// LineCount 함수는 Files에 든 줄 수를 파일 이름 순서로 돌려줍니다. 재생본 크기와 소스 대조 범위를 보고할 때 씁니다.
func (b Bundle) LineCount() (files []string, total int) {
	for path, lines := range b.Files {
		files = append(files, path)
		total += len(lines)
	}
	sort.Strings(files)
	return files, total
}
