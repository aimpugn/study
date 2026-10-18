package atlasmap

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/collect"
)

// labDir 와 designDir 는 이 패키지 폴더에서 실습 폴더와 화면 파일 폴더까지의 경로입니다.
const (
	labDir    = "../../../../database/postgresql"
	designDir = "../../../design"
)

// collected 값은 77벌을 한 번만 옮겨 두고 시험들이 함께 씁니다. 옮긴 뒤 JSON으로 쓰고 다시 읽은 run이라 공통 형식에 있는 것만 남습니다.
var collected struct {
	once sync.Once
	runs map[string]*atlas.Run
	err  error
}

// formatRuns 함수는 기록 넷을 공통 형식으로 옮긴 뒤 파일 형식으로 쓰고 다시 읽은 run을 id마다 돌려줍니다.
//
// 적재기의 Go 값이 아니라 형식(atlas.Write → atlas.Read)을 거친 값만 시험이 보게 하려는 것입니다. 실패하면 시험을 멈춥니다.
func formatRuns(t *testing.T) map[string]*atlas.Run {
	t.Helper()
	collected.once.Do(func() {
		runs, err := collect.All(labDir, designDir)
		if err != nil {
			collected.err = err
			return
		}
		out := map[string]*atlas.Run{}
		for id, r := range runs {
			var buf bytes.Buffer
			if err := atlas.Write(&buf, r); err != nil {
				collected.err = err
				return
			}
			back, err := atlas.Read(&buf)
			if err != nil {
				collected.err = err
				return
			}
			out[id] = back
		}
		collected.runs = out
	})
	if collected.err != nil {
		t.Fatal(collected.err)
	}
	return collected.runs
}

// runOf 함수는 id가 같은 run을 돌려주고, 없으면 시험을 멈춥니다.
func runOf(t *testing.T, id string) *atlas.Run {
	t.Helper()
	r, ok := formatRuns(t)[id]
	if !ok {
		t.Fatalf("run %s가 없음", id)
	}
	return r
}

// placeName 함수는 장소 id의 이름을 돌려줍니다.
func placeName(r *atlas.Run, id int) string {
	for _, p := range r.Places {
		if p.ID == id {
			return p.Name
		}
	}
	return ""
}

// attr 함수는 사건의 attrs에서 출처 src의 열쇠 key 값을 돌려줍니다. 없으면 nil입니다.
func attr(e atlas.Event, src, key string) any {
	return e.Attrs[src][key]
}

// ms 함수는 나노초를 0.1 ms까지 반올림한 ms로 돌려줍니다(시안이 적은 자리수).
func ms(ns int64) float64 { return math.Round(float64(ns)/1e5) / 10 }

// us 함수는 나노초를 0.1 µs까지 반올림한 µs로 돌려줍니다.
func us(ns int64) float64 { return math.Round(float64(ns)/100) / 10 }

// querySpan 함수는 여정 run의 질의 구간(해석기 출처의 부모 없는 span)을 돌려줍니다.
func querySpan(t *testing.T, r *atlas.Run) atlas.Event {
	t.Helper()
	for _, e := range r.Events {
		if e.Source == "journey" && e.Kind == atlas.Span && e.Parent == nil {
			return e
		}
	}
	t.Fatalf("%s: 질의 구간이 없음", r.ID)
	return atlas.Event{}
}

// kernelRoots 함수는 시스템 콜 사건에서 kernel 연결로 이어진 커널 마디(ftrace 출처의 뿌리)를 시스템 콜 차례로 돌려줍니다.
func kernelRoots(r *atlas.Run) []atlas.Event {
	byID := map[int]atlas.Event{}
	for _, e := range r.Events {
		byID[e.ID] = e
	}
	var out []atlas.Event
	for _, l := range r.Links {
		if l.Kind == "kernel" {
			out = append(out, byID[l.Dst])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// childrenOf 함수는 부모가 id인 사건을 id 차례로 돌려줍니다(형식 규칙 1: 같은 출처는 기록 순서대로 id가 늘어남).
func childrenOf(r *atlas.Run, id int) []atlas.Event {
	var out []atlas.Event
	for _, e := range r.Events {
		if e.Parent != nil && *e.Parent == id {
			out = append(out, e)
		}
	}
	return out
}

// findPath 함수는 뿌리에서 이름 사슬을 따라 내려가 마지막 마디를 돌려줍니다. 같은 이름의 자식이 여럿이면 가장 오래 머문 것을 고릅니다.
//
// 시안의 사슬(예: fdatasync → … → btrfs_sync_log → mutex_lock)을 그대로 따라가야 같은 이름의 짧은 호출(mutex_lock 42 ns 등)과 섞이지 않습니다.
func findPath(r *atlas.Run, root atlas.Event, names ...string) (atlas.Event, bool) {
	cur := root
	if placeName(r, cur.Place) != names[0] {
		return atlas.Event{}, false
	}
	for _, name := range names[1:] {
		found := false
		var best atlas.Event
		for _, c := range childrenOf(r, cur.ID) {
			if placeName(r, c.Place) == name && (!found || *c.DurNS > *best.DurNS) {
				best, found = c, true
			}
		}
		if !found {
			return atlas.Event{}, false
		}
		cur = best
	}
	return cur, true
}

// TestPremiseInsertRoutes 함수는 시안(t3-merged/notes.md 5절)의 INSERT 숫자가 공통 형식에서 나오는지 봅니다.
//
// Given 여정의 INSERT 1차·2차 run(kernel_insert, kernel_insert2)을 형식으로 쓰고 다시 읽었을 때,
// When 질의 구간, 단계 사건, 이동, 메시지 바이트를 구조 칸과 해당 출처의 attrs로 뽑으면,
// Then 시안의 값(376.8·148.5 ms, 단계 머문 시간 일곱, 범위 A의 시각, Q 35 B·Z 22 B, pid)과 같습니다.
func TestPremiseInsertRoutes(t *testing.T) {
	first, second := runOf(t, "journey/kernel_insert"), runOf(t, "journey/kernel_insert2")
	for _, c := range []struct {
		r    *atlas.Run
		ms   float64
		proc string
	}{{first, 376.8, "6421"}, {second, 148.5, "6576"}} {
		q := querySpan(t, c.r)
		if got := ms(*q.EndNS - *q.StartNS); got != c.ms {
			t.Errorf("%s: 질의 %.1f ms, 시안 %.1f", c.r.ID, got, c.ms)
		}
		if got := placeName(c.r, q.Place); got != c.proc {
			t.Errorf("%s: 백엔드 %s, 시안 %s", c.r.ID, got, c.proc)
		}
	}

	// 단계 머문 시간은 다음 단계 시작(마지막은 질의 끝)까지입니다. 시안은 이 차를 유도값으로 적었습니다.
	var stages []atlas.Event
	for _, e := range first.Events {
		if e.Source == "journey" && e.Kind == atlas.Instant {
			stages = append(stages, e)
		}
	}
	sort.Slice(stages, func(i, j int) bool { return *stages[i].StartNS < *stages[j].StartNS })
	end := *querySpan(t, first).EndNS
	var dwell []float64
	for i, s := range stages {
		next := end
		if i+1 < len(stages) {
			next = *stages[i+1].StartNS
		}
		dwell = append(dwell, ms(next-*s.StartNS))
	}
	if want := []float64{0.6, 1.9, 276.6, 15.3, 19.3, 56.3, 6.8}; fmt.Sprint(dwell) != fmt.Sprint(want) {
		t.Errorf("단계 머문 시간 %v, 시안 %v", dwell, want)
	}

	// 이동은 기록기 출처에서 해석기가 n을 붙인 사건입니다. 범위 A(n 35~39), n 33, n 37의 시각과 메시지 바이트를 봅니다.
	hop := map[int]atlas.Event{}
	for _, e := range first.Events {
		if n, ok := attr(e, "journey", "n").(interface{ String() string }); ok {
			var k int
			fmt.Sscan(n.String(), &k)
			hop[k] = e
		}
	}
	if len(hop) != 42 {
		t.Fatalf("이동 %d줄, 시안 42", len(hop))
	}
	for n, want := range map[int]float64{33: 309.2, 35: 314.2, 37: 315.7, 39: 368.0} {
		if got := ms(*hop[n].StartNS); got != want {
			t.Errorf("n %d 시작 +%.1f ms, 시안 +%.1f", n, got, want)
		}
	}
	for n, want := range map[int]string{1: "35", 42: "22"} {
		if got := fmt.Sprint(attr(hop[n], "rec", "bytes")); got != want {
			t.Errorf("n %d 바이트 %s, 시안 %s", n, got, want)
		}
	}
}

// TestPremiseCatalogFold 함수는 시안의 「카탈로그 조회 ×25」와 「페이지 읽기 ×5」가 형식에서 무엇으로 나오는지 고정합니다.
//
// Given INSERT 1차 run에서,
// When 해석기가 카탈로그로 표시한 이동(journey.catalog)과 의미 분석 단계의 이동을 세면,
// Then 카탈로그 표시는 24줄(원래 이동 71번)이고 의미 분석 단계의 이동은 n 3~27의 25줄입니다(n 13 「락 잡기」는 카탈로그 표시가 없음).
// 그래서 시안의 ×25는 「카탈로그 조회」가 아니라 의미 분석 단계의 이동 줄 수이고, 화면 v1은 형식이 가르는 대로 둘을 나눠 보입니다.
// 카탈로그 이동 가운데 파일에서 온 페이지 읽기는 5번이고 구간은 +2.9 → +266.9 ms입니다.
func TestPremiseCatalogFold(t *testing.T) {
	r := runOf(t, "journey/kernel_insert")
	rows, raw, reads, analyze := 0, 0, 0, 0
	var lo, hi int64 = math.MaxInt64, 0
	files := map[int]bool{}
	for _, p := range r.Places {
		if p.Kind == "file" {
			files[p.ID] = true
		}
	}
	for _, e := range r.Events {
		if attr(e, "journey", "stage") == "의미 분석" {
			analyze++
		}
		if attr(e, "journey", "catalog") != true {
			continue
		}
		rows++
		raw += e.Count
		if e.From != nil && files[*e.From] {
			reads++
		}
		lo = min(lo, *e.StartNS)
		hi = max(hi, *e.EndNS)
	}
	if rows != 24 || raw != 71 || reads != 5 || analyze != 25 {
		t.Errorf("카탈로그 %d줄(이동 %d번), 페이지 읽기 %d, 의미 분석 이동 %d줄; 기대 24(71), 5, 25", rows, raw, reads, analyze)
	}
	if ms(lo) != 2.9 || ms(hi) != 266.9 {
		t.Errorf("카탈로그 구간 +%.1f → +%.1f ms, 시안 +2.9 → +266.9", ms(lo), ms(hi))
	}
}

// TestPremiseKernel 함수는 함수·커널 줌의 숫자와 1차·2차가 갈리는 마디가 공통 형식에서 나오는지 봅니다.
//
// Given INSERT 1차·2차 run의 시스템 콜 → 커널 연결과 ftrace 마디의 부모 나무에서,
// When 뿌리와 이름으로 마디를 찾고, 두 나무를 장소 key의 차례로 견주면,
// Then pwrite64 49.5 µs, fdatasync 51,845.8 µs·6,439.5 µs, mutex_lock 40,815.3 µs, schedule 40,813.3 µs, write_all_supers 10,734.8·5,867.0 µs,
// I/O 기다림 5,349.6 µs, submit_bio 255.7 µs, virtio_queue_rq 254.1 µs가 나오고, 머문 시간 차를 따라간 주된 갈림은 btrfs_sync_log입니다.
func TestPremiseKernel(t *testing.T) {
	first, second := runOf(t, "journey/kernel_insert"), runOf(t, "journey/kernel_insert2")
	r1, r2 := kernelRoots(first), kernelRoots(second)
	if len(r1) != 2 || len(r2) != 2 {
		t.Fatalf("커널 나무 %d·%d개, 기대 2·2(pwrite64, fdatasync)", len(r1), len(r2))
	}
	if got := us(*r1[0].DurNS); got != 49.5 {
		t.Errorf("pwrite64 커널 %.1f µs, 시안 49.5", got)
	}
	sync := []string{"__arm64_sys_fdatasync", "ovl_fsync", "vfs_fsync_range", "btrfs_sync_file", "btrfs_sync_log"}
	path := func(tail ...string) []string { return append(append([]string{}, sync...), tail...) }
	want := []struct {
		r    *atlas.Run
		root atlas.Event
		path []string
		us   float64
	}{
		{first, r1[1], sync[:1], 51845.8},
		{first, r1[1], path("mutex_lock"), 40815.3},
		{first, r1[1], path("mutex_lock", "__mutex_lock_slowpath", "__mutex_lock", "schedule_preempt_disabled", "schedule"), 40813.3},
		{first, r1[1], path("write_all_supers"), 10734.8},
		{first, r1[1], path("write_all_supers", "submit_bio"), 255.7},
		{first, r1[1], path("write_all_supers", "submit_bio", "blk_mq_submit_bio", "blk_mq_try_issue_directly", "virtio_queue_rq"), 254.1},
		{second, r2[1], sync[:1], 6439.5},
		{second, r2[1], path("write_all_supers"), 5867.0},
		{second, r2[1], path("write_all_supers", "wait_for_completion_io"), 5349.6},
	}
	for _, w := range want {
		e, ok := findPath(w.r, w.root, w.path...)
		if !ok {
			t.Errorf("%s: %v 사슬이 없음", w.r.ID, w.path)
			continue
		}
		if got := us(*e.DurNS); got != w.us {
			t.Errorf("%s: %s %.1f µs, 시안 %.1f", w.r.ID, w.path[len(w.path)-1], got, w.us)
		}
	}

	// 갈림은 두 가지로 셉니다. 「처음 갈림」은 두 나무를 먼저 부른 차례대로 같이 걸어 자식들의 장소 key 차례가 처음 다른 마디이고,
	// 「주된 갈림」은 자식 차례가 같은 동안 머문 시간 차가 가장 큰 자식 짝으로 내려가다가 자식 차례가 다른 마디입니다.
	// 시안의 btrfs_sync_log는 주된 갈림이고, 처음 갈림은 그보다 앞의 작은 차이(btrfs_writepages의 자식 차례)입니다.
	key := func(r *atlas.Run, e atlas.Event) string {
		for _, p := range r.Places {
			if p.ID == e.Place {
				return p.Key
			}
		}
		return ""
	}
	same := func(a, b atlas.Event) bool {
		ca, cb := childrenOf(first, a.ID), childrenOf(second, b.ID)
		if len(ca) != len(cb) {
			return false
		}
		for i := range ca {
			if key(first, ca[i]) != key(second, cb[i]) {
				return false
			}
		}
		return true
	}
	var firstSplit func(a, b atlas.Event) (string, bool)
	firstSplit = func(a, b atlas.Event) (string, bool) {
		if !same(a, b) {
			return placeName(first, a.Place), true
		}
		cb := childrenOf(second, b.ID)
		for i, c := range childrenOf(first, a.ID) {
			if at, ok := firstSplit(c, cb[i]); ok {
				return at, true
			}
		}
		return "", false
	}
	a, b := r1[1], r2[1]
	for same(a, b) && len(childrenOf(first, a.ID)) > 0 {
		ca, cb := childrenOf(first, a.ID), childrenOf(second, b.ID)
		i := 0
		for j := range ca {
			if abs(*ca[j].DurNS-*cb[j].DurNS) > abs(*ca[i].DurNS-*cb[i].DurNS) {
				i = j
			}
		}
		a, b = ca[i], cb[i]
	}
	if got := placeName(first, a.Place); got != "btrfs_sync_log" {
		t.Errorf("주된 갈림 %q, 시안 btrfs_sync_log", got)
	}
	if got, ok := firstSplit(r1[1], r2[1]); !ok || got != "btrfs_writepages" {
		t.Errorf("처음 갈림 %q(%v), 기대 btrfs_writepages", got, ok)
	}
}

// abs 함수는 정수의 절댓값입니다.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestPremiseSelectCandidates 함수는 T4 행의 「같은 SELECT의 첫 조회와 다시 조회」가 어느 실행인지 고정합니다.
//
// Given 트랜잭션 실험 1-rc-read의 a run에서,
// When 같은 세션 S1이 보낸 단계 가운데 SQL이 같은 SELECT를 고르고, 단계가 품은 요청(exec_simple_query)을 따라가면,
// Then 단계 둘(첫 조회, 다시 조회)이 같은 백엔드 장소의 요청 둘로 이어지고, 요청 길이는 287.3 ms와 23.6 ms로 다릅니다.
// 여정의 SELECT는 한 벌(select)뿐이라(두 번째 해석 rec-c3은 같은 기록을 다시 푼 것) 다른 경로 후보가 되지 못합니다.
func TestPremiseSelectCandidates(t *testing.T) {
	r := runOf(t, "tx/1-rc-read/a")
	byID := map[int]atlas.Event{}
	for _, e := range r.Events {
		byID[e.ID] = e
	}
	contains := map[int][]int{}
	for _, l := range r.Links {
		if l.Kind == "contains" {
			contains[l.Src] = append(contains[l.Src], l.Dst)
		}
	}
	bySQL := map[string][]atlas.Event{}
	for _, e := range r.Events {
		if e.Source == "txlab" && e.Name == "step" {
			sql := fmt.Sprint(attr(e, "txlab", "sql"))
			bySQL[sql] = append(bySQL[sql], e)
		}
	}
	const sel = "SELECT pg_current_snapshot() AS snap, ctid, xmin, xmax, id, balance FROM acct WHERE balance >= 100 ORDER BY id;"
	steps := bySQL[sel]
	if len(steps) != 2 || placeName(r, steps[0].Place) != "S1" || steps[0].Place != steps[1].Place {
		t.Fatalf("같은 SELECT 단계 %d개, 기대 S1의 둘", len(steps))
	}
	var lens []float64
	var backends []int
	for _, s := range steps {
		for _, id := range contains[s.ID] {
			if q := byID[id]; q.Name == "exec_simple_query" {
				lens = append(lens, ms(*q.EndNS-*q.StartNS))
				backends = append(backends, q.Place)
			}
		}
	}
	if fmt.Sprint(lens) != "[287.3 23.6]" || len(backends) != 2 || backends[0] != backends[1] {
		t.Errorf("요청 길이 %v, 백엔드 %v; 기대 [287.3 23.6], 같은 백엔드", lens, backends)
	}
	sels := 0
	for id, x := range formatRuns(t) {
		if x.Topic == "pgsql" && x.Scene == "journey" && fmt.Sprint(x.Attrs["topic"]["label"]) == "SELECT" {
			sels++
			_ = id
		}
	}
	if sels != 1 {
		t.Errorf("여정의 SELECT run %d벌, 기대 1", sels)
	}
}

// TestPremiseSpringEnds 함수는 Spring 기록이 끝나는 층과 「봤고 0번」을 가를 근거가 형식에 있는지 봅니다.
//
// Given Spring 14벌에서,
// When 출처와 관찰 범위(coverage)를 보면,
// Then 커널을 기록한 출처(ftrace·strace)가 없고, 연결 위 메시지는 연결 장소마다 complete로 지켜봤다는 관찰 범위가 있습니다.
// 그래서 커널 층은 '기록하지 않은 층', 메시지가 지나지 않은 구간의 연결 층은 '기록했지만 사건 없음'으로 가를 수 있습니다.
func TestPremiseSpringEnds(t *testing.T) {
	n := 0
	for _, r := range formatRuns(t) {
		if r.Scene == "" || r.ID[:7] != "spring/" {
			continue
		}
		n++
		for _, s := range r.Sources {
			if s.ID == "ftrace" || s.ID == "strace" {
				t.Errorf("%s: 커널 출처 %s가 있음", r.ID, s.ID)
			}
		}
		complete := false
		for _, c := range r.Coverage {
			if c.Source == "wire-mac" && c.Completeness == "complete" && len(c.Places) > 0 {
				complete = true
			}
		}
		if !complete {
			t.Errorf("%s: 연결 위 메시지의 complete 관찰 범위가 없음", r.ID)
		}
	}
	if n != 14 {
		t.Errorf("Spring run %d벌, 기대 14", n)
	}
}
