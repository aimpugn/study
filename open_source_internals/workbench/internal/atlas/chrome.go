package atlas

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ChromeTrace 값은 Chrome의 Trace Event 형식(JSON 객체 꼴)입니다. Perfetto UI와 chrome://tracing이 엽니다.
type ChromeTrace struct {
	TraceEvents     []map[string]any `json:"traceEvents"`
	DisplayTimeUnit string           `json:"displayTimeUnit"`
}

// ChromeExport 함수는 run 하나를 Chrome Trace Event 형식으로 옮기고, 옮기지 못한 사건의 수를 까닭별로 돌려줍니다.
//
// 시계마다 프로세스 묶음(pid) 하나를 두어 서로 다른 시계의 사건이 한 시간축에 섞이지 않게 하고, 장소마다 줄(tid) 하나를 둡니다.
// 끝이 있는 구간과 group은 X(시작과 길이), 끝을 보지 못한 구간은 B, 한 순간과 관찰은 i로 씁니다.
// 시작 시각이 없는 사건(gdb 기록, 길이만 아는 커널 마디)은 시간축에 놓을 수 없어 옮기지 않습니다.
func ChromeExport(r *Run) (ChromeTrace, map[string]int) {
	skipped := map[string]int{}
	pidOf := map[string]int{}
	var evs []map[string]any

	// 시계마다 pid를 정하고 이름을 붙입니다. 시계를 적지 않은 출처는 "시계 없음" 묶음으로 갑니다.
	pid := func(clock string) int {
		if p, ok := pidOf[clock]; ok {
			return p
		}
		p := len(pidOf) + 1
		pidOf[clock] = p
		name := clock
		if name == "" {
			name = "시계 없음"
		}
		evs = append(evs, map[string]any{"ph": "M", "name": "process_name", "pid": p, "tid": 0,
			"args": map[string]any{"name": fmt.Sprintf("%s · %s", r.ID, name)}})
		return p
	}
	named := map[[2]int]bool{}

	for _, e := range r.Events {
		if e.StartNS == nil {
			skipped["시작 시각 없음"]++
			continue
		}
		p := pid(clockOf(r, e))
		// 장소마다 줄 이름을 한 번 붙입니다.
		if !named[[2]int{p, e.Place}] {
			named[[2]int{p, e.Place}] = true
			pl := r.Places[e.Place]
			evs = append(evs, map[string]any{"ph": "M", "name": "thread_name", "pid": p, "tid": e.Place,
				"args": map[string]any{"name": fmt.Sprintf("%s %s", pl.Kind, pl.Name)}})
		}
		ev := map[string]any{"name": e.Name, "cat": e.Source, "pid": p, "tid": e.Place,
			"ts": NanosToMicros(*e.StartNS), "args": map[string]any{"event": e.ID, "kind": e.Kind, "count": e.Count}}
		switch {
		case (e.Kind == Span || e.Kind == Group) && e.EndNS != nil:
			ev["ph"] = "X"
			ev["dur"] = NanosToMicros(*e.EndNS - *e.StartNS)
		case e.Kind == Span:
			ev["ph"] = "B"
		default:
			ev["ph"] = "i"
			ev["s"] = "t"
		}
		evs = append(evs, ev)
	}
	return ChromeTrace{TraceEvents: evs, DisplayTimeUnit: "ns"}, skipped
}

// ChromeJSON 함수는 ChromeExport의 결과를 JSON 바이트로 씁니다.
func ChromeJSON(r *Run) ([]byte, map[string]int, error) {
	t, skipped := ChromeExport(r)
	b, err := json.Marshal(t)
	return b, skipped, err
}

// Overlap 값은 같은 시계·같은 장소에서 서로 어긋나 겹치는 구간 둘입니다(한쪽이 다른 쪽을 품지도, 떨어져 있지도 않음).
type Overlap struct {
	A, B int
}

// CrossingSpans 함수는 같은 시계·같은 장소의 끝이 있는 구간 가운데 서로 어긋나 겹치는 쌍을 모두 찾습니다.
//
// Perfetto는 한 줄의 구간이 바르게 포개지지 않으면 넘침 줄로 따로 그립니다. 이 함수는 그 상황을 파일을 열지 않고 찾습니다.
// group은 따로 떼어 낼 수 없는 일의 묶음이라 포개짐을 따지지 않습니다.
func CrossingSpans(r *Run) []Overlap {
	type track struct {
		clock string
		place int
	}
	byTrack := map[track][]Event{}
	for _, e := range r.Events {
		if e.Kind == Span && e.StartNS != nil && e.EndNS != nil {
			t := track{clockOf(r, e), e.Place}
			byTrack[t] = append(byTrack[t], e)
		}
	}
	var out []Overlap
	for _, evs := range byTrack {
		// 시작이 이른 것부터, 같으면 긴 것부터 놓으면 품는 구간이 먼저 옵니다.
		sort.Slice(evs, func(i, j int) bool {
			if *evs[i].StartNS != *evs[j].StartNS {
				return *evs[i].StartNS < *evs[j].StartNS
			}
			return *evs[i].EndNS > *evs[j].EndNS
		})
		// 열린 구간의 더미를 유지합니다. 새 구간이 맨 위 구간 안에서 끝나지 않으면 어긋난 것입니다.
		var stack []Event
		for _, e := range evs {
			for len(stack) > 0 && *stack[len(stack)-1].EndNS <= *e.StartNS {
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 && *e.EndNS > *stack[len(stack)-1].EndNS {
				out = append(out, Overlap{A: stack[len(stack)-1].ID, B: e.ID})
			}
			stack = append(stack, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].A != out[j].A {
			return out[i].A < out[j].A
		}
		return out[i].B < out[j].B
	})
	return out
}
