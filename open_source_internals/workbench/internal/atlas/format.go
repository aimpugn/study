// Package atlas 는 학습 지도의 공통 사건 형식(v0)입니다.
//
// 기록 한 벌(run)은 장소(place), 사건(event), 연결(link), 설명(note)의 네 표와 run 자신의 머리로 이뤄집니다.
// 주제마다 다른 값은 각 행의 attrs에만 두고, attrs는 그 값을 낸 출처의 id로 나눕니다.
// 이 패키지는 주제를 모릅니다. 주제 쪽 적재기가 이 형식으로 옮기고, 플랫폼은 구조 칸만 읽습니다.
// 설계와 그 이유는 docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md의 「T1 공통 형식 v0와 기존 기록 이식」에 있습니다.
package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Run 값은 기록 한 벌입니다.
//
// Clocks와 Sources는 사건의 시각과 출처를 해석하는 열쇠이고, Coverage는 출처마다 무엇을 지켜봤는지를 적어
// "봤고 0번"과 "안 봄"을 가릅니다. Unrecorded는 그 기록에 원래 없는 칸입니다.
type Run struct {
	ID         string       `json:"id"`
	Topic      string       `json:"topic"`
	Scene      string       `json:"scene"`
	Name       string       `json:"name"`
	Origin     Origin       `json:"origin"`
	Clocks     []Clock      `json:"clocks"`
	Sources    []Source     `json:"sources"`
	Coverage   []Coverage   `json:"coverage"`
	Unrecorded []Unrecorded `json:"unrecorded"`
	Attrs      Attrs        `json:"attrs"`
	Places     []Place      `json:"places"`
	Events     []Event      `json:"events"`
	Links      []Link       `json:"links"`
	Notes      []Note       `json:"notes"`
}

// Origin 값은 run의 시각 0이 무엇인지 적습니다. EpochNS가 nil이면 그 순간의 절대 시각은 기록에 없습니다.
type Origin struct {
	Clock   string `json:"clock,omitempty"`
	Label   string `json:"label"`
	EpochNS *int64 `json:"epoch_ns"`
}

// Clock 값은 시계 하나입니다. PrecisionNS는 그 시계로 찍은 시각을 믿을 수 있는 가장 작은 단위입니다.
// Relation이 nil이면 다른 시계와의 관계를 모릅니다.
type Clock struct {
	ID          string         `json:"id"`
	PrecisionNS int64          `json:"precision_ns"`
	Relation    *ClockRelation `json:"relation"`
	Note        string         `json:"note,omitempty"`
}

// ClockRelation 값은 이 시계의 시각에서 To 시계의 시각을 뺀 차이의 범위입니다.
//
// Known이 false이면 범위를 모릅니다(옮겼다는 사실만 압니다). Basis는 그 범위를 낸 출처의 id입니다.
type ClockRelation struct {
	To    string `json:"to"`
	Known bool   `json:"known"`
	LoNS  int64  `json:"lo_ns"`
	HiNS  int64  `json:"hi_ns"`
	Basis string `json:"basis"`
	Note  string `json:"note,omitempty"`
}

// 출처의 성격입니다. 판정의 뜻은 설계 메모 7에 있습니다.
const (
	// Observed 는 기록기가 남긴 값을 형식만 바꿔 옮긴 출처입니다.
	Observed = "observed"
	// Derived 는 기록 사실에서 정해진 규칙으로 셈한 출처입니다(합친 수, 요청 창, 트리 만들기).
	Derived = "derived"
	// Inferred 는 틀릴 수 있는 짝짓기와 짐작입니다(시간 창으로 붙인 짝, 이름으로 찾은 후보).
	Inferred = "inferred"
	// Human 은 사람이 정한 이름과 글입니다.
	Human = "human"
)

// Source 값은 출처 하나입니다. 사건과 장소의 Source, 연결의 Basis가 ID를 가리킵니다.
type Source struct {
	ID      string `json:"id"`
	Tool    string `json:"tool"`
	Version string `json:"version,omitempty"`
	Clock   string `json:"clock,omitempty"`
	Nature  string `json:"nature"`
	Note    string `json:"note,omitempty"`
}

// Coverage 값은 출처 하나가 무엇을 어디서 지켜봤는지입니다.
//
// What은 지켜본 일의 종류(함수 진입, 시스템 콜 등), Places와 Names는 지켜본 장소와 이름, Window가 nil이면 run 전체입니다.
// Completeness는 complete·partial·sampled 가운데 하나이고, Basis는 이 목록의 근거(파일, 추정)입니다.
type Coverage struct {
	Source       string    `json:"source"`
	What         string    `json:"what"`
	Places       []int     `json:"places,omitempty"`
	Names        []string  `json:"names,omitempty"`
	Window       *[2]int64 `json:"window,omitempty"`
	Completeness string    `json:"completeness"`
	Basis        string    `json:"basis"`
}

// Unrecorded 값은 출처 하나가 원래 기록하지 않는 칸입니다(예: gdb 기록의 시각).
type Unrecorded struct {
	Source string `json:"source"`
	Field  string `json:"field"`
	Note   string `json:"note,omitempty"`
}

// Attrs 값은 출처 id마다 그 출처가 낸 주제 값을 담습니다. 값은 JSON 값 그대로이고, 숫자는 json.Number로 둡니다.
type Attrs map[string]map[string]any

// Loc 값은 소스 트리 안의 한 줄입니다.
type Loc struct {
	Path string `json:"path"`
	Line int    `json:"line,omitempty"`
}

// Place 값은 장소 하나입니다. Parent가 nil이면 맨 위 장소입니다.
//
// Key는 함수·프레임처럼 기록과 무관하게 정해지는 정적 장소에만 있고, 기록을 가로질러 같은 장소를 잇는 데 씁니다.
type Place struct {
	ID     int    `json:"id"`
	Parent *int   `json:"parent"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Key    string `json:"key,omitempty"`
	Loc    *Loc   `json:"loc,omitempty"`
	Source string `json:"source"`
	Attrs  Attrs  `json:"attrs,omitempty"`
}

// 사건의 종류입니다.
const (
	// Span 은 시작과 끝이 있는 구간입니다. 끝을 보지 못했으면 EndNS가 nil입니다.
	Span = "span"
	// Instant 는 한 순간의 일입니다.
	Instant = "instant"
	// Sample 은 관찰 한 번입니다. 관찰이 읽은 행은 attrs에 있습니다.
	Sample = "sample"
	// Group 은 따로 떼어 낼 수 없는 여러 번의 일을 묶은 사건입니다. Count가 2 이상이고 다른 사건의 부모가 될 수 없습니다.
	Group = "group"
)

// Event 값은 사건 하나입니다.
//
// 시각은 출처가 쓰는 시계의 좌표로 적은 정수 나노초입니다. StartNS가 nil이면 그 출처는 시각을 기록하지 않았고,
// DurNS는 시작과 끝을 모르고 길이만 알 때만 씁니다. Parent는 부모와 출처가 같을 때만 쓰고, Depth는 부모 사슬로 셈한 사본입니다.
// Stack은 장소의 목록이고, 있으면 0번은 사건의 장소 자신(가장 안쪽 칸)이고 그 뒤가 바깥쪽으로 부른 쪽입니다.
// Place는 그 일이 실행된 자리이고, Target은 사건이 향하거나 다루는 실행 위치, From은 무엇을 옮기는 사건의 출발 장소입니다.
type Event struct {
	ID      int    `json:"id"`
	Place   int    `json:"place"`
	From    *int   `json:"from,omitempty"`
	Target  *int   `json:"target,omitempty"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	StartNS *int64 `json:"start_ns"`
	EndNS   *int64 `json:"end_ns"`
	DurNS   *int64 `json:"dur_ns,omitempty"`
	Count   int    `json:"count"`
	Parent  *int   `json:"parent"`
	Depth   int    `json:"depth"`
	Stack   []int  `json:"stack,omitempty"`
	Source  string `json:"source"`
	Attrs   Attrs  `json:"attrs,omitempty"`
}

// Link 값은 사건 둘을 잇는 짝입니다. Basis는 이 짝을 정한 출처의 id입니다.
type Link struct {
	ID    int    `json:"id"`
	Src   int    `json:"src"`
	Dst   int    `json:"dst"`
	Kind  string `json:"kind"`
	Basis string `json:"basis"`
	Attrs Attrs  `json:"attrs,omitempty"`
}

// Scope 값은 설명이 붙는 대상입니다. Kind는 run·place·event·topic 가운데 하나이고, run과 topic이면 ID가 nil입니다.
type Scope struct {
	Kind string `json:"kind"`
	ID   *int   `json:"id"`
}

// Note 값은 설명 하나입니다. Kind는 fixed(주제가 정한 글)·ambiguity(원본이 가르지 못하는 꼴) 등입니다.
type Note struct {
	ID     int    `json:"id"`
	Scope  Scope  `json:"scope"`
	Kind   string `json:"kind"`
	Text   string `json:"text"`
	Source string `json:"source"`
	Attrs  Attrs  `json:"attrs,omitempty"`
}

// Write 함수는 run 하나를 JSON으로 씁니다. 같은 run은 언제나 같은 바이트가 됩니다(attrs의 키는 이름순).
//
// 쓰기 오류는 그대로 돌려줍니다.
func Write(w io.Writer, r *Run) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("run %s 쓰기: %w", r.ID, err)
	}
	return nil
}

// Read 함수는 Write가 쓴 JSON을 run으로 읽습니다. attrs의 숫자는 글자 그대로 json.Number로 남습니다.
//
// 형식에 없는 칸이 있으면 오류를 돌려줍니다. 형식이 바뀌었는데 조용히 일부만 읽는 일을 막기 위해서입니다.
func Read(r io.Reader) (*Run, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	dec.DisallowUnknownFields()
	var run Run
	if err := dec.Decode(&run); err != nil {
		return nil, fmt.Errorf("run 읽기: %w", err)
	}
	return &run, nil
}

// ReadFile 함수는 파일 하나에서 run을 읽습니다. 파일을 열 수 없거나 형식이 틀리면 오류를 돌려줍니다.
func ReadFile(path string) (*Run, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Read(bytes.NewReader(b))
}

// WriteFile 함수는 run 하나를 파일로 씁니다. 있던 파일은 덮어씁니다.
func WriteFile(path string, r *Run) error {
	var buf bytes.Buffer
	if err := Write(&buf, r); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// SourceByID 메서드는 id가 같은 출처를 돌려줍니다. 없으면 두 번째 값이 false입니다.
func (r *Run) SourceByID(id string) (Source, bool) {
	for _, s := range r.Sources {
		if s.ID == id {
			return s, true
		}
	}
	return Source{}, false
}

// ClockByID 메서드는 id가 같은 시계를 돌려줍니다. 없으면 두 번째 값이 false입니다.
func (r *Run) ClockByID(id string) (Clock, bool) {
	for _, c := range r.Clocks {
		if c.ID == id {
			return c, true
		}
	}
	return Clock{}, false
}
