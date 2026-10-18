package mapview

// View 값은 화면 하나가 읽는 보기 모형 전체입니다. 화면은 이것만 읽고 run을 직접 읽지 않습니다.
type View struct {
	Title   string     `json:"title"`
	Default int        `json:"default"`
	Zones   []Zone     `json:"zones"`
	Bands   []BandView `json:"bands"`
	Groups  []Group    `json:"groups"`
	Routes  []Route    `json:"routes"`
	Fixed   FixedView  `json:"fixed"`
}

// BandView 값은 화면이 그리는 띠 하나입니다. Under는 함수 줌 띠가 프로세스 안 줌에서 합쳐 보이는 띠, Below는 지면 아래인지입니다.
type BandView struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Zoom   string `json:"zoom"`
	Zone   string `json:"zone"`
	Ground bool   `json:"ground,omitempty"`
	Below  bool   `json:"below,omitempty"`
	Under  string `json:"under,omitempty"`
}

// FixedView 값은 고정 설명입니다. Bands는 띠 id마다, Names는 장소 이름마다의 글이고, Source는 글이 온 자료 파일입니다.
type FixedView struct {
	Bands  map[string]string `json:"bands"`
	Names  map[string]string `json:"names"`
	Source string            `json:"source"`
}

// Group 값은 같은 요청의 경로 후보 무리입니다. Routes는 View.Routes의 번호입니다.
type Group struct {
	ID       string `json:"id"`
	Scene    string `json:"scene"`
	Identity string `json:"identity"`
	Label    string `json:"label"`
	Typed    string `json:"typed"`
	Routes   []int  `json:"routes"`
}

// 띠의 상태입니다(고친 규칙 5).
const (
	// StateEvents 는 보이는 범위에 그 띠의 사건이 있다는 뜻입니다.
	StateEvents = "events"
	// StateOther 는 보이는 범위 밖에서 같은 시간대에 그 띠의 사건만 있다는 뜻입니다(흐림).
	StateOther = "other"
	// StateWatched 는 관찰 출처가 그 띠의 장소를 complete로 지켜봤고 그 시간에 사건이 0개라는 뜻입니다(봤고 0번).
	StateWatched = "watched"
	// StateUnseen 은 나머지입니다(안 봄, 빗금).
	StateUnseen = "unseen"
)

// Route 값은 경로 하나입니다. Steps가 프로세스 안 줌의 칸, System이 시스템 줌의 칸이고, 칸마다 함수 줌의 나무를 품습니다.
//
// StartNS·EndNS는 칸들이 덮는 구간이고 Clock은 그 시각의 시계입니다. OrderOnly가 참이면 시각이 없어 id 차례로 놓은 칸입니다.
// Bands는 경로 전체에서 본 띠의 상태, Untimed는 뿌리가 품었지만 어느 칸도 품지 않은 시각 없는 사건입니다.
type Route struct {
	ID         string            `json:"id"`
	Run        string            `json:"run"`
	Scene      string            `json:"scene"`
	SceneLabel string            `json:"scene_label"`
	Name       string            `json:"name"`
	Typed      string            `json:"typed"`
	Group      int               `json:"group"`
	Clock      string            `json:"clock"`
	OrderOnly  bool              `json:"order_only,omitempty"`
	Home       string            `json:"home"`
	Process    string            `json:"process,omitempty"`
	StartNS    int64             `json:"start_ns"`
	EndNS      int64             `json:"end_ns"`
	Columns    []Column          `json:"columns,omitempty"`
	Steps      []Step            `json:"steps"`
	System     []SysStep         `json:"system"`
	Bands      map[string]string `json:"bands"`
	Zones      map[string]string `json:"zones"`
	Untimed    []Loose           `json:"untimed,omitempty"`
	Divergence []Divergence      `json:"divergence,omitempty"`
}

// Column 값은 처리 단계 열 하나입니다. Steps는 그 열에 든 칸 번호의 처음과 끝이고, 칸이 없으면 [-1, -1]입니다.
type Column struct {
	Name    string `json:"name"`
	StartNS int64  `json:"start_ns"`
	DwellNS int64  `json:"dwell_ns"`
	Steps   [2]int `json:"steps"`
	Nature  string `json:"nature"`
}

// Loose 값은 칸에 들지 않은 시각 없는 사건 하나입니다.
type Loose struct {
	Event  int    `json:"event"`
	Name   string `json:"name"`
	Band   string `json:"band"`
	Nature string `json:"nature"`
	Text   string `json:"text,omitempty"`
}

// Step 값은 프로세스 안 줌의 칸 하나입니다.
//
// Events는 이 칸의 사건(접힌 칸이면 여럿), Leaves는 그 칸을 펼친 나무의 잎 사건 id(정렬)입니다. Key는 칸 사건 장소의 정적 열쇠이고 T5가 범위의 가지 모양 위치로 바꿀 자리입니다.
type Step struct {
	Index  int    `json:"i"`
	Name   string `json:"name"`
	Band   string `json:"band"`
	Origin string `json:"origin,omitempty"`
	Place  string `json:"place"`
	Key    string `json:"key,omitempty"`
	Column string `json:"column,omitempty"`
	T0     *int64 `json:"t0"`
	T1     *int64 `json:"t1"`
	DurNS  int64  `json:"dur_ns"`
	// DurNature는 머문 시간의 성격입니다. 칸이 순간이라 펼친 나무의 마디 길이를 쓰면 그 마디의 성격이고, 아니면 칸의 성격과 같습니다.
	DurNature string `json:"dur_nature"`
	Count     int    `json:"count"`
	Fold      int    `json:"fold,omitempty"`
	// FoldNature는 접기 조건 열쇠의 출처 성격입니다(접은 판단이 해석기 값이면 「규칙으로 셈」).
	FoldNature string            `json:"fold_nature,omitempty"`
	Events     []int             `json:"events"`
	Leaves     []int             `json:"leaves"`
	Nature     string            `json:"nature"`
	Loc        string            `json:"loc,omitempty"`
	Stack      []string          `json:"stack,omitempty"`
	Attrs      []AttrView        `json:"attrs,omitempty"`
	Descent    *Descent          `json:"descent,omitempty"`
	Tree       []Node            `json:"tree"`
	Bands      map[string]string `json:"bands"`
	SleepNS    int64             `json:"sleep_ns,omitempty"`
}

// AttrView 값은 칸의 자세한 내용에 보일 값 하나와 그 출처 성격입니다.
type AttrView struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Nature string `json:"nature"`
}

// Descent 값은 칸이 지면(시스템 콜 경계)을 건너는지와, 건넜다면 그 아래가 기록됐는지입니다(고친 규칙 4).
//
// State가 events면 지면 아래 사건이 펼친 나무에 있고, unseen이면 시스템 콜은 있으나 그 아래를 기록하지 않았습니다.
// Count는 지면을 건넌 사건의 수(접힌 칸의 「×k」), Recorded는 그 가운데 아래가 기록된 수, Nature는 칸에서 지면까지 온 길의 가장 약한 성격입니다.
// 접힌 칸에서 Recorded < Count면 화면은 「×k 가운데 r 기록」으로 보여 기록되지 않은 하강을 기록된 것처럼 보이지 않습니다.
type Descent struct {
	State    string `json:"state"`
	Count    int    `json:"count"`
	Recorded int    `json:"recorded"`
	Nature   string `json:"nature"`
	DurNS    int64  `json:"dur_ns,omitempty"`
}

// Node 값은 함수 줌 나무의 마디 하나입니다. Parent는 같은 나무의 번호(뿌리는 -1), Via는 부모에서 이 마디로 온 길(parent 또는 연결 종류)입니다.
//
// Nature는 사건의 출처, 띠를 정한 attrs 열쇠의 출처, 뿌리에서 이 마디까지 온 길(조상의 출처와 거친 연결 근거) 가운데 가장 약한 성격입니다(고친 규칙 6).
// OtherClock은 칸과 시계가 달라 가로 차례가 시각이 아니라는 표지, Visible은 렌즈가 켜졌을 때 보이는지, Folded는 이 마디에 접힌 숨은 마디 수입니다.
type Node struct {
	Event      int    `json:"event"`
	Name       string `json:"name"`
	Band       string `json:"band"`
	Depth      int    `json:"depth"`
	Parent     int    `json:"parent"`
	Via        string `json:"via"`
	DurNS      *int64 `json:"dur_ns"`
	T0         *int64 `json:"t0"`
	Sleep      bool   `json:"sleep,omitempty"`
	Nature     string `json:"nature"`
	OtherClock bool   `json:"other_clock,omitempty"`
	Visible    bool   `json:"visible"`
	Folded     int    `json:"folded,omitempty"`
	Key        string `json:"key,omitempty"`
	Loc        string `json:"loc,omitempty"`
	Leaves     []int  `json:"leaves"`
}

// SysStep 값은 시스템 줌의 칸 하나입니다. Steps는 합친 프로세스 안 칸의 처음과 끝 번호, Zones는 지난 구역의 차례입니다.
type SysStep struct {
	Index  int      `json:"i"`
	Zones  []string `json:"zones"`
	Steps  [2]int   `json:"steps"`
	Leaves []int    `json:"leaves"`
	DurNS  int64    `json:"dur_ns"`
	Nature string   `json:"nature"`
}

// Divergence 값은 같은 무리의 다른 후보와 이 경로가 갈리는 곳입니다(고친 규칙 9).
//
// Kind가 steps면 칸 차례가 Step 칸에서 처음 다르고, tree면 칸 차례는 같고 머문 시간 차가 가장 큰 Step 칸의 나무에서 Node 마디의
// 자식 차례가 다르며(주된 갈림), time이면 잎까지 구조가 같고 Node가 시간 차가 가장 큰 마디입니다. unrecorded면 Node 아래를 한쪽만 기록해
// 실행이 갈렸는지 기록 범위가 다른지 가를 수 없습니다(X1). none이면 견줄 나무가 없습니다.
type Divergence struct {
	Other  int    `json:"other"`
	Kind   string `json:"kind"`
	Step   int    `json:"step"`
	Node   int    `json:"node"`
	Name   string `json:"name"`
	DiffNS int64  `json:"diff_ns"`
}
