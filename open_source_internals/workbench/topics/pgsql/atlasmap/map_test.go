package atlasmap

import (
	"fmt"
	"sync"
	"testing"

	"study/open_source_internals/workbench/internal/atlas/mapview"
)

// built 값은 77벌로 만든 보기 모형을 한 번만 만들어 시험들이 함께 씁니다.
var built struct {
	once sync.Once
	v    *mapview.View
	err  error
}

// view 함수는 주제 자료 파일(map.json)과 형식을 거친 77벌로 보기 모형을 만듭니다. 실패하면 시험을 멈춥니다.
func view(t *testing.T) *mapview.View {
	t.Helper()
	runs := formatRuns(t)
	built.once.Do(func() {
		spec, err := mapview.LoadSpec("map.json")
		if err != nil {
			built.err = err
			return
		}
		built.v, built.err = mapview.Build(mapview.MemStore(runs), spec, ".")
	})
	if built.err != nil {
		t.Fatal(built.err)
	}
	return built.v
}

// routeOf 함수는 run id(뿌리가 있으면 "run#뿌리")가 같은 경로를 돌려줍니다.
func routeOf(t *testing.T, v *mapview.View, id string) mapview.Route {
	t.Helper()
	for _, r := range v.Routes {
		if r.ID == id || r.Run == id {
			return r
		}
	}
	t.Fatalf("경로 %s가 없음", id)
	return mapview.Route{}
}

// TestInsertRouteMatchesDraft 함수는 시안의 INSERT 1차가 보기 모형에서 어떻게 나오는지 고정합니다.
//
// Given 여정의 kernel_insert에서, When 보기 모형을 만들면,
// Then 칸은 20개(카탈로그 ×10·×14, 형식이 가르는 대로), 열 일곱의 머문 시간은 시안과 같고, 디스크까지 동기화 칸은 커널 층까지 기록된 하강(추정 짝)이며
// 잠든 시간이 있고, 페이지 읽기 칸들은 하강이지만 커널을 기록하지 않아 빗금입니다. 시스템 콜 경계 아래 커널 층까지 줌 인할 수 있습니다(C3).
func TestInsertRouteMatchesDraft(t *testing.T) {
	r := routeOf(t, view(t), "journey/kernel_insert")
	var folds []int
	for _, s := range r.Steps {
		if s.Fold > 0 {
			folds = append(folds, s.Fold)
		}
	}
	if len(r.Steps) != 20 || fmt.Sprint(folds) != "[10 14]" {
		t.Errorf("칸 %d개, 접힘 %v; 기대 20, [10 14]", len(r.Steps), folds)
	}
	var dwell []string
	for _, c := range r.Columns {
		dwell = append(dwell, fmt.Sprintf("%.1f", float64(c.DwellNS)/1e6))
	}
	if fmt.Sprint(dwell) != "[0.6 1.9 276.6 15.3 19.3 56.3 6.8]" {
		t.Errorf("열 머문 시간 %v", dwell)
	}
	var sync, reads, readsUnseen int
	for _, s := range r.Steps {
		switch s.Name {
		case "디스크까지 동기화":
			sync++
			if s.Descent == nil || s.Descent.State != mapview.StateEvents || s.Descent.Nature != "inferred" || s.SleepNS == 0 {
				t.Errorf("동기화 칸의 하강 %+v, 잠듦 %d", s.Descent, s.SleepNS)
			}
			if fmt.Sprintf("%.1f", float64(s.DurNS)/1e3) != "51871.0" {
				t.Errorf("동기화 칸 머문 시간 %d", s.DurNS)
			}
		case "페이지 읽기 테이블·인덱스", "카탈로그 조회":
			if s.Descent != nil {
				reads += s.Descent.Count
				if s.Descent.State == mapview.StateUnseen {
					readsUnseen += s.Descent.Count
				}
			}
		}
	}
	if sync != 1 || reads != 6 || readsUnseen != 6 {
		t.Errorf("동기화 칸 %d, 페이지 읽기 하강 %d(빗금 %d), 기대 1, 6(6)", sync, reads, readsUnseen)
	}
	for _, b := range []string{"kernel", "k_btrfs", "k_lock", "k_sched", "k_virtio"} {
		if r.Bands[b] != mapview.StateEvents {
			t.Errorf("띠 %s 상태 %s, 기대 events(커널 층까지 줌 인)", b, r.Bands[b])
		}
	}
}

// TestInsertCandidatesAndDivergence 함수는 같은 INSERT의 다른 기록 셋이 후보이고 1차·2차의 갈림이 btrfs_sync_log인지 봅니다.
//
// Given 친 글이 같은 INSERT 셋(insert, kernel_insert, kernel_insert2)에서, When 무리와 갈림을 보면,
// Then 한 무리이고, 1차와 2차는 주된 갈림이 btrfs_sync_log(tree, 차 45.5 ms)이고, 커널 기록이 없는 insert는 커널 띠가 빗금(X1)입니다.
func TestInsertCandidatesAndDivergence(t *testing.T) {
	v := view(t)
	first := routeOf(t, v, "journey/kernel_insert")
	g := v.Groups[first.Group]
	var runs []string
	for _, i := range g.Routes {
		runs = append(runs, v.Routes[i].Run)
	}
	if fmt.Sprint(runs) != "[journey/insert journey/kernel_insert journey/kernel_insert2]" {
		t.Fatalf("후보 %v", runs)
	}
	for _, d := range first.Divergence {
		if v.Routes[d.Other].Run == "journey/kernel_insert2" {
			if d.Kind != "tree" || d.Name != "btrfs_sync_log" || fmt.Sprintf("%.1f", float64(d.DiffNS)/1e6) != "45.5" {
				t.Errorf("1차·2차 갈림 %+v", d)
			}
		}
	}
	if s := routeOf(t, v, "journey/insert").Bands["kernel"]; s != mapview.StateUnseen {
		t.Errorf("insert의 커널 띠 %s, 기대 unseen", s)
	}
}

// TestSelectFirstAndAgainAreCandidates 함수는 T4 행의 「같은 SELECT의 첫 조회와 다시 조회」가 서로 다른 경로 후보로 고르는지 봅니다.
//
// Given 트랜잭션 실험 1-rc-read의 a·b에서, When S1의 같은 SELECT 단계를 뿌리로 한 경로를 모으면,
// Then 한 무리에 넷(a·b 각각 첫 조회와 다시 조회)이 있고, 같은 벌의 두 후보는 칸 차례가 갈리며, 시각 없는 메시지(Q·T·D·C·Z)는 칸이 아니라 목록에 있습니다.
func TestSelectFirstAndAgainAreCandidates(t *testing.T) {
	v := view(t)
	first := routeOf(t, v, "tx/1-rc-read/a#17")
	g := v.Groups[first.Group]
	var ids []string
	for _, i := range g.Routes {
		ids = append(ids, v.Routes[i].ID)
	}
	if fmt.Sprint(ids) != "[tx/1-rc-read/a#17 tx/1-rc-read/a#92 tx/1-rc-read/b#17 tx/1-rc-read/b#92]" {
		t.Fatalf("후보 %v", ids)
	}
	for _, d := range first.Divergence {
		if v.Routes[d.Other].ID == "tx/1-rc-read/a#92" && d.Kind != "steps" {
			t.Errorf("첫 조회·다시 조회 갈림 %+v, 기대 steps", d)
		}
	}
	var untimed []string
	for _, u := range first.Untimed {
		untimed = append(untimed, u.Name)
	}
	if fmt.Sprint(untimed) != "[Q T D D C Z]" {
		t.Errorf("시각 없는 사건 %v", untimed)
	}
	for _, s := range first.Steps {
		if s.T0 == nil {
			t.Errorf("시각 없는 칸 %s", s.Name)
		}
	}
}

// TestSpringEndsAboveTheKernel 함수는 같은 화면 코드로 Spring 장면을 보이고 기록이 끝나는 층이 빗금인지 봅니다(C3).
//
// Given Spring 5-required의 a·b에서, When 경로를 만들면,
// Then 칸은 요청(맥 시계)이고, 시스템 콜 경계 아래 띠는 모두 안 봄(빗금)이며, 요청을 펼친 나무에서 쓰지 않은 연결 C2 띠는 complete 관찰이라 봤고 0번입니다.
// 앱 쪽 JFR 마디는 추정 연결(wire)로만 닿아 ≈이고 시계가 달라 「시계 다름」 표지가 붙습니다.
func TestSpringEndsAboveTheKernel(t *testing.T) {
	v := view(t)
	r := routeOf(t, v, "spring/5-required/a")
	if r.Clock != "mac" || len(r.Steps) == 0 {
		t.Fatalf("시계 %s, 칸 %d", r.Clock, len(r.Steps))
	}
	for _, b := range v.Bands {
		if b.Below && r.Bands[b.ID] != mapview.StateUnseen {
			t.Errorf("지면 아래 띠 %s 상태 %s, 기대 unseen", b.ID, r.Bands[b.ID])
		}
	}
	watched, wire := 0, 0
	for _, s := range r.Steps {
		if s.Bands["conn2"] == mapview.StateWatched {
			watched++
		}
		for _, n := range s.Tree {
			if n.Via == "wire" {
				wire++
				if n.Nature != "inferred" || !n.OtherClock {
					t.Errorf("wire 마디 %s 성격 %s 시계 다름 %v", n.Name, n.Nature, n.OtherClock)
				}
			}
		}
	}
	if watched == 0 || wire == 0 {
		t.Errorf("C2 봤고 0번 칸 %d, wire 마디 %d", watched, wire)
	}
}
