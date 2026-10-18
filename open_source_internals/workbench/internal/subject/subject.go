// Package subject 는 주제 폴더의 workbench.json을 읽고, 소스 경로를 화면의 층으로 분류합니다.
//
// 주제 하나(예: PostgreSQL)에 고유한 값은 모두 그 주제 폴더의 workbench.json이 소유합니다. 작업대 코드는 주제 이름을 알지 못합니다.
package subject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Watch 값은 실행 때 고를 수 있는 감시 목록 하나입니다.
//
// File은 주제 폴더 기준 경로이고, StackAt은 처음 멈출 때 호출 스택 전체를 남길 함수입니다(run-trace.sh의 세 번째 인자).
// PGOptions는 이 감시 목록의 세션에만 더할 서버 설정입니다(예: "-c enable_seqscan=off"). 비어 있어도 됩니다.
// 학습자 입력이 아니라 주제 정의가 정하는 값이고, SQL은 한 문장만 받으므로 SET으로는 바꿀 수 없는 실행 조건을 여기서 정합니다.
type Watch struct {
	ID        string   `json:"id"`
	File      string   `json:"file"`
	Label     string   `json:"label"`
	StackAt   string   `json:"stackAt"`
	PGOptions string   `json:"pgOptions,omitempty"`
	Examples  []string `json:"examples"`
}

// validPGOptions 값은 PGOptions에 받아들일 꼴입니다. "-c 이름=값"을 공백으로 이은 것만 받아, 설정 파일의 오타가 조용히 다른 조건이 되지 않게 합니다.
var validPGOptions = regexp.MustCompile(`^-c [a-z_]+=[A-Za-z0-9_.]+( -c [a-z_]+=[A-Za-z0-9_.]+)*$`)

// Layer 값은 소스 경로 접두사 하나에 붙인 층 이름입니다. 이 이름은 작업대가 디렉터리를 보고 붙인 분류이고, 소스가 정한 이름이 아닙니다.
type Layer struct {
	Prefix string `json:"prefix"`
	Label  string `json:"label"`
}

// Subject 값은 workbench.json 하나의 내용과 그 파일이 있는 주제 폴더입니다.
type Subject struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Guide        string  `json:"guide"`
	Container    string  `json:"container"`
	Image        string  `json:"image"`
	SourceRoot   string  `json:"sourceRoot"`
	SourceNotice string  `json:"sourceNotice"`
	StartCommand string  `json:"startCommand"`
	RunScript    string  `json:"runScript"`
	Watches      []Watch `json:"watches"`
	Layers       []Layer `json:"layers"`

	// Dir은 workbench.json이 있는 폴더의 절대경로입니다. 파일에는 없고 Load가 채웁니다.
	Dir string `json:"-"`
}

// Lane 값은 화면 그림의 세로줄 하나입니다. Index는 왼쪽부터의 순서이고, 같은 Label의 접두사들은 한 줄을 함께 씁니다.
type Lane struct {
	Index int    `json:"index"`
	Label string `json:"label"`
}

// Unclassified 값은 어느 접두사에도 맞지 않는 경로의 층 이름입니다.
const Unclassified = "분류 없음"

// Load 함수는 path의 workbench.json을 읽어 Subject로 돌려줍니다.
//
// 필수 값(id, container, image, sourceRoot, runScript, 감시 목록 하나 이상)이 비었거나, 감시 목록 id가 겹치거나,
// 감시 목록의 pgOptions가 "-c 이름=값" 꼴이 아니거나, runScript나 감시 목록 파일이 주제 폴더에 없으면 오류를 돌려줍니다.
func Load(path string) (Subject, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Subject{}, fmt.Errorf("주제 정의 읽기: %w", err)
	}
	var s Subject
	if err := json.Unmarshal(raw, &s); err != nil {
		return Subject{}, fmt.Errorf("주제 정의 %s 해석: %w", path, err)
	}
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return Subject{}, fmt.Errorf("주제 폴더 경로: %w", err)
	}
	s.Dir = dir

	for name, v := range map[string]string{"id": s.ID, "container": s.Container, "image": s.Image, "sourceRoot": s.SourceRoot, "runScript": s.RunScript} {
		if v == "" {
			return Subject{}, fmt.Errorf("주제 정의 %s에 %s가 없습니다", path, name)
		}
	}
	if len(s.Watches) == 0 {
		return Subject{}, fmt.Errorf("주제 정의 %s에 감시 목록이 없습니다", path)
	}
	files := []string{s.RunScript}
	seen := map[string]bool{}
	for _, w := range s.Watches {
		if w.ID == "" || seen[w.ID] {
			return Subject{}, fmt.Errorf("감시 목록 id가 비었거나 겹칩니다: %q", w.ID)
		}
		seen[w.ID] = true
		if w.PGOptions != "" && !validPGOptions.MatchString(w.PGOptions) {
			return Subject{}, fmt.Errorf("감시 목록 %s의 pgOptions는 \"-c 이름=값\"을 공백으로 이은 꼴이어야 합니다: %q", w.ID, w.PGOptions)
		}
		files = append(files, w.File)
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return Subject{}, fmt.Errorf("주제 폴더에 %s가 없습니다: %w", f, err)
		}
	}
	return s, nil
}

// Watch 함수는 id가 같은 감시 목록을 돌려줍니다. 없으면 false입니다.
func (s Subject) Watch(id string) (Watch, bool) {
	for _, w := range s.Watches {
		if w.ID == id {
			return w, true
		}
	}
	return Watch{}, false
}

// Lanes 함수는 Layers의 이름을 처음 나온 순서대로 한 번씩 세로줄로 만들고, 맨 끝에 Unclassified 줄을 더합니다.
func (s Subject) Lanes() []Lane {
	var lanes []Lane
	index := map[string]bool{}
	for _, l := range s.Layers {
		if !index[l.Label] {
			index[l.Label] = true
			lanes = append(lanes, Lane{Index: len(lanes), Label: l.Label})
		}
	}
	return append(lanes, Lane{Index: len(lanes), Label: Unclassified})
}

// Classify 함수는 소스 상대경로 path에 맞는 가장 긴 접두사의 층 이름을 돌려줍니다.
//
// 맞는 접두사가 없으면 Unclassified를 돌려줍니다. 가장 긴 접두사를 고르므로 파일 하나에 붙인 이름이 그 디렉터리 전체의 이름보다 앞섭니다.
func (s Subject) Classify(path string) string {
	best, label := -1, Unclassified
	for _, l := range s.Layers {
		if strings.HasPrefix(path, l.Prefix) && len(l.Prefix) > best {
			best, label = len(l.Prefix), l.Label
		}
	}
	return label
}
