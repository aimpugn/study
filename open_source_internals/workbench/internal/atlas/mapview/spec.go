// Package mapview 는 공통 형식(atlas v0)의 기록에서 지도 화면의 보기 모형을 만들고, 그 모형을 품은 정적 화면 파일을 씁니다.
//
// 주제를 모릅니다. 주제는 자료 파일(Spec) 하나로 띠·구역·경로·후보·접기의 규칙을 넘기고, 이 패키지는 형식의 구조 칸과
// 자료 파일이 이름을 댄 attrs 열쇠만 읽습니다. 설계는 docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md의 「T4 지도 화면 v1」에 있습니다.
package mapview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"study/open_source_internals/workbench/internal/atlas"
)

// 줌의 이름입니다. 시스템(L1)은 구역, 프로세스 안(L2)은 띠, 함수·커널(L3)은 한 칸을 펼친 나무입니다.
const (
	// ZoomSystem 은 구역 사이의 이동을 보는 줌입니다.
	ZoomSystem = "system"
	// ZoomProcess 는 경로의 사건을 한 칸씩 띠 위에 보는 줌입니다.
	ZoomProcess = "process"
	// ZoomFunction 은 칸 하나를 함수·커널 호출로 펼친 줌입니다.
	ZoomFunction = "function"
)

// Spec 값은 주제가 넘기는 지도 자료 파일 한 벌입니다.
//
// Bands의 차례가 위에서 아래로의 깊이 차례이고, 띠 규칙은 앞의 것이 먼저 맞습니다.
type Spec struct {
	Title  string      `json:"title"`
	Zones  []Zone      `json:"zones"`
	Bands  []Band      `json:"bands"`
	Sleep  []string    `json:"sleep"`
	Scenes []SceneSpec `json:"scenes"`
	// Fixed는 고정 설명의 자리입니다. 열쇠는 띠 id나 장소 이름이고 값은 주제 자료 파일 안의 위치입니다.
	Fixed []FixedRef `json:"fixed,omitempty"`
	// Default는 화면이 처음 보일 경로의 run id 앞부분입니다.
	Default string `json:"default,omitempty"`
}

// Zone 값은 시스템 줌의 구역 하나입니다. X·Y·W·H는 구역 지도의 칸(0~100의 비율 좌표)입니다.
type Zone struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	W    float64 `json:"w"`
	H    float64 `json:"h"`
}

// Band 값은 띠 하나입니다. Zoom은 이 띠가 처음 보이는 줌(process 또는 function), Zone은 시스템 줌에서 속하는 구역입니다.
//
// Ground가 참이면 시스템 콜 경계(지면)이고, 이 띠보다 아래 띠로 가는 사건만 하강선으로 그립니다.
// Under는 함수 줌 띠가 프로세스 안 줌에서 합쳐 보일 띠입니다(커널의 층들이 프로세스 안 줌에서는 커널 띠 하나).
type Band struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Zoom   string `json:"zoom"`
	Zone   string `json:"zone"`
	Ground bool   `json:"ground,omitempty"`
	Under  string `json:"under,omitempty"`
	Match  []Rule `json:"match"`
}

// PlaceMatch 값은 장소 하나에 거는 조건입니다. 빈 조건은 모두를 받습니다.
type PlaceMatch struct {
	Kinds      []string `json:"kinds,omitempty"`
	Names      []string `json:"names,omitempty"`
	NamePrefix []string `json:"name_prefix,omitempty"`
}

// Rule 값은 띠 규칙 하나입니다. 채운 조건이 모두 맞아야 맞습니다.
//
// Names는 사건의 이름, Place는 사건의 자리(도착이 있으면 도착), Ancestor는 그 장소나 조상 가운데 하나에 거는 조건입니다.
// Attr은 "출처 id.열쇠" 꼴이고, 사건의 attrs에 그 값이 있어야 하며 Values가 있으면 그 가운데 하나여야 맞습니다.
// Fallback이 참인 규칙은 다른 모든 규칙이 맞지 않을 때만 봅니다(띠의 깊이 차례와 맞추는 차례를 떼어 놓으려는 것).
type Rule struct {
	Fallback bool        `json:"fallback,omitempty"`
	Names    []string    `json:"names,omitempty"`
	Place    *PlaceMatch `json:"place,omitempty"`
	Ancestor *PlaceMatch `json:"ancestor,omitempty"`
	Sources  []string    `json:"sources,omitempty"`
	Attr     string      `json:"attr,omitempty"`
	Values   []string    `json:"values,omitempty"`
}

// Columns 값은 프로세스 안 줌의 열(처리 단계) 규칙입니다. Attr("출처 id.열쇠")의 값이 같은 칸이 한 열이고,
// Starts가 고른 사건(이름이 열 이름)의 시각이 그 열의 시작입니다.
type Columns struct {
	Attr   string         `json:"attr"`
	Starts atlas.Selector `json:"starts"`
}

// Lens 값은 함수 줌에서 보이는 마디를 고르는 규칙입니다. 칸의 머문 시간에 대한 몫이 MinShare 이상인 마디와 잠드는 함수, 그 조상이 보입니다.
type Lens struct {
	MinShare float64 `json:"min_share"`
}

// Expand 값은 칸을 펼칠 때 따라갈 연결 하나입니다. Dir이 in이면 그 사건으로 들어오는 연결의 출발 사건을 붙이고 더 따라가지 않습니다.
type Expand struct {
	Link string `json:"link"`
	Dir  string `json:"dir,omitempty"`
}

// SceneSpec 값은 장면 무리 하나의 경로 규칙입니다.
//
// Runs는 run id의 앞부분입니다. Root가 nil이면 run 전체가 경로 하나이고, 아니면 Root가 고른 사건마다 경로 하나입니다.
// Steps는 경로의 칸이 될 사건을 고릅니다. Via가 있으면 뿌리에서 그 연결을 따라 닿는 사건 가운데서, 없으면 run 전체에서 고릅니다.
type SceneSpec struct {
	ID       string          `json:"id"`
	Label    string          `json:"label"`
	Runs     []string        `json:"runs"`
	Home     string          `json:"home"`
	Root     *atlas.Selector `json:"root,omitempty"`
	Via      []string        `json:"via,omitempty"`
	Steps    atlas.Selector  `json:"steps"`
	Skip     *atlas.Selector `json:"skip,omitempty"`
	Expand   []Expand        `json:"expand,omitempty"`
	Fold     *Fold           `json:"fold,omitempty"`
	Columns  *Columns        `json:"columns,omitempty"`
	Lens     Lens            `json:"lens"`
	Identity Identity        `json:"identity"`
	// Name과 Typed는 경로의 이름과 친 글을 고르는 곳이고, Stage는 단계별 안내를 묶는 attrs 열쇠입니다.
	Name  Identity `json:"name"`
	Typed string   `json:"typed,omitempty"`
	Stage string   `json:"stage,omitempty"`
}

// Fold 값은 잇단 칸을 한 칸 「×N」으로 접는 조건입니다. 같은 열 안에서 Attr("출처 id.열쇠")의 값이 참인 칸이 잇달면 접고, N은 줄 수입니다.
type Fold struct {
	Attr  string `json:"attr"`
	Label string `json:"label"`
}

// Identity 값은 같은 요청을 가르는 값입니다. 채운 것을 모두 이어 붙여 정체로 씁니다.
//
// RunField는 run의 머리 칸(scene, name, id, scenario = name의 마지막 "/" 앞), RunAttr과 RootAttr은 run과 뿌리 사건의 attrs("출처 id.열쇠"),
// RootPlace가 참이면 뿌리 사건의 장소 이름도 정체에 들어갑니다.
type Identity struct {
	RunField  string `json:"run_field,omitempty"`
	RunAttr   string `json:"run_attr,omitempty"`
	RootAttr  string `json:"root_attr,omitempty"`
	RootPlace bool   `json:"root_place,omitempty"`
}

// FixedRef 값은 고정 설명 하나의 자리입니다. File은 자료 파일의 폴더에서 본 경로, Path는 그 JSON 안의 열쇠 사슬입니다.
//
// Key가 띠 id면 그 띠의 설명, Kind가 "names"이면 Path의 값이 [[이름, 글], …] 꼴이라 장소 이름마다 붙습니다.
type FixedRef struct {
	Key  string   `json:"key"`
	Kind string   `json:"kind,omitempty"`
	File string   `json:"file"`
	Path []string `json:"path"`
}

// LoadSpec 함수는 자료 파일을 읽습니다. 모르는 칸이 있거나 띠·구역의 참조가 풀리지 않으면 오류를 돌려줍니다.
func LoadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var s Spec
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := s.check(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &s, nil
}

// check 메서드는 자료 파일 안의 참조를 봅니다. 띠의 구역, 장면의 집 띠, 줌 이름, 지면 띠의 수가 맞아야 합니다.
func (s *Spec) check() error {
	zones := map[string]bool{}
	for _, z := range s.Zones {
		zones[z.ID] = true
	}
	bands := map[string]bool{}
	ground := 0
	for _, b := range s.Bands {
		if bands[b.ID] {
			return fmt.Errorf("띠 %s가 둘", b.ID)
		}
		bands[b.ID] = true
		if !zones[b.Zone] {
			return fmt.Errorf("띠 %s의 구역 %s가 없음", b.ID, b.Zone)
		}
		if b.Zoom != ZoomProcess && b.Zoom != ZoomFunction {
			return fmt.Errorf("띠 %s의 줌 %q", b.ID, b.Zoom)
		}
		if b.Ground {
			ground++
		}
	}
	for _, b := range s.Bands {
		if b.Zoom == ZoomFunction && !bands[b.Under] {
			return fmt.Errorf("함수 줌 띠 %s의 합쳐 보일 띠 %q가 없음", b.ID, b.Under)
		}
	}
	if ground != 1 {
		return fmt.Errorf("지면 띠가 %d개(하나여야 함)", ground)
	}
	for _, sc := range s.Scenes {
		if !bands[sc.Home] {
			return fmt.Errorf("장면 %s의 집 띠 %s가 없음", sc.ID, sc.Home)
		}
	}
	return nil
}

// sceneOf 메서드는 run id가 앞부분에 맞는 장면 규칙을 돌려줍니다. 없으면 nil입니다.
func (s *Spec) sceneOf(runID string) *SceneSpec {
	for i := range s.Scenes {
		for _, p := range s.Scenes[i].Runs {
			if strings.HasPrefix(runID, p) {
				return &s.Scenes[i]
			}
		}
	}
	return nil
}

// match 메서드는 장소 하나가 조건에 맞는지 봅니다.
func (m *PlaceMatch) match(p atlas.Place) bool {
	if len(m.Kinds) > 0 && !contains(m.Kinds, p.Kind) {
		return false
	}
	if len(m.Names) > 0 && !contains(m.Names, p.Name) {
		return false
	}
	if len(m.NamePrefix) > 0 {
		ok := false
		for _, pre := range m.NamePrefix {
			if strings.HasPrefix(p.Name, pre) {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// contains 함수는 목록에 값이 있는지 봅니다.
func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// attrValue 함수는 사건이나 run의 attrs에서 "출처 id.열쇠"의 값을 글로 돌려줍니다. 없으면 두 번째 값이 false입니다.
func attrValue(a atlas.Attrs, path string) (string, bool) {
	src, key, ok := strings.Cut(path, ".")
	if !ok {
		return "", false
	}
	v, ok := a[src][key]
	if !ok || v == nil {
		return "", false
	}
	return fmt.Sprint(v), true
}
