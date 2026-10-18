package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 이 파일은 실험 5 앱의 JFR 기록(spring.jfr)을 읽습니다. JDK의 jfr 명령으로 JSON을 받아 쓰고, 이진 형식은 직접 읽지 않습니다.
//
// JDK 25 jdk.MethodTrace의 두 성질(WORK 실험 5 「방법의 변경과 도구 시험」, 2026-10-15 18:35 기록 열 벌에서 다시 확인)에 맞춰 풉니다.
//   - 메서드 안의 throw를 같은 메서드의 finally가 받으면 같은 시작 시각의 사건이 둘 남고, 긴 쪽이 실제 끝입니다. 그래서 (클래스, 이름,
//     서술자, 스레드, 시작 시각)이 같은 사건은 하나로 합치고 긴 쪽을 씁니다. 이름이 같은 오버로드는 서술자가 달라 합치지 않습니다.
//   - 부른 쪽에서 올라온 예외가 그냥 지나가며 끝난 호출은 사건이 없습니다. 그 호출은 남은 사건의 스택에만 보이므로 stackOnly로 셉니다.

// jfrStackDepth 는 jfr print에 넘기는 스택 깊이입니다. 기본값 5로는 장면 메서드(lab.Outer)까지 닿지 않습니다.
const jfrStackDepth = 64

// jfrNoise 는 예외 목록에서 빼는 던짐입니다. JVM이 메서드 핸들을 이을 때와 pgjdbc가 번역 묶음을 찾을 때 던지고 스스로 받는 것이라
// 장면의 흐름이 아닙니다(2026-10-15 r5 기록에서 장면 구간의 NoSuchMethodError는 모두 java.lang.invoke 안, ClassNotFoundException은
// org.postgresql.util.GT 안). 뺀 수는 dropped에 남깁니다.
var jfrNoise = map[string]bool{
	"java.lang.NoSuchMethodError":      true,
	"java.lang.NoSuchFieldError":       true,
	"java.lang.ClassNotFoundException": true,
}

// jfrMethod 는 JFR이 적은 메서드 하나입니다.
type jfrMethod struct {
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
	Name       string `json:"name"`
	Descriptor string `json:"descriptor"`
}

// key 는 "패키지.클래스::메서드" 꼴의 이름입니다.
func (m jfrMethod) key() string {
	return strings.ReplaceAll(m.Type.Name, "/", ".") + "::" + m.Name
}

// jfrFrame 은 스택의 칸 하나입니다. 메서드 추적 사건의 첫 칸은 그 메서드를 부른 쪽입니다.
type jfrFrame struct {
	Method jfrMethod `json:"method"`
	Line   int       `json:"lineNumber"`
}

// ref 는 "패키지.클래스::메서드:줄" 꼴의 자리입니다.
func (f jfrFrame) ref() string {
	return fmt.Sprintf("%s:%d", f.Method.key(), f.Line)
}

// jfrStack 은 사건의 스택입니다. Frames는 안쪽부터이고, Truncated는 JFR이 바깥 칸을 잘랐다는 표시입니다(--stack-depth 한도).
type jfrStack struct {
	Frames    []jfrFrame `json:"frames"`
	Truncated bool       `json:"truncated"`
}

// jfrEvent 는 jfr print --json의 사건 하나 가운데 쓰는 칸입니다.
type jfrEvent struct {
	Type   string `json:"type"`
	Values struct {
		StartTime   string `json:"startTime"`
		Duration    string `json:"duration"`
		EventThread struct {
			JavaName string `json:"javaName"`
		} `json:"eventThread"`
		Method      *jfrMethod `json:"method"`
		StackTrace  *jfrStack  `json:"stackTrace"`
		ThrownClass *struct {
			Name string `json:"name"`
		} `json:"thrownClass"`
		Message *string `json:"message"`
	} `json:"values"`
}

// javaSpan 은 메서드 호출 하나입니다. Caller는 부른 자리, Depth는 같은 스레드에서 이 호출을 감싼 기록된 호출의 수입니다.
// Twice는 같은 호출이 두 사건으로 남아 하나로 합쳤다는 표시입니다. stack은 부른 쪽부터 바깥으로의 칸이고, truncated는 JFR이 그 스택의
// 바깥 칸을 잘랐다는 표시입니다(--stack-depth 한도).
type javaSpan struct {
	Fn     string  `json:"fn"`
	Desc   string  `json:"desc"`
	Thread string  `json:"thread"`
	T0     float64 `json:"t0"`
	T1     float64 `json:"t1"`
	Caller string  `json:"caller"`
	Depth  int     `json:"depth"`
	Twice  bool    `json:"twice,omitempty"`
	stack  []jfrFrame
	// truncated 는 위 주석의 잘림 표시입니다.
	truncated bool
}

// javaThrow 는 예외가 만들어진 자리 하나입니다. At은 스택의 앞 칸들(만든 쪽부터)입니다. stack은 예외 클래스의 생성자 칸을 뺀
// 스택 전체이고, 호출 트리에서 그 예외를 만든 마디를 찾는 데 씁니다(spring_tree.go).
type javaThrow struct {
	T       float64  `json:"t"`
	Thread  string   `json:"thread"`
	Class   string   `json:"class"`
	Message string   `json:"message"`
	At      []string `json:"at"`
	stack   []jfrFrame
}

// javaPark 는 스레드가 1 ms 넘게 멈춘 자리 하나입니다. Under는 스택에서 처음 만나는 추적 대상 메서드입니다. stack은 스택 전체입니다.
type javaPark struct {
	T0     float64  `json:"t0"`
	T1     float64  `json:"t1"`
	Thread string   `json:"thread"`
	Under  string   `json:"under"`
	At     []string `json:"at"`
	stack  []jfrFrame
}

// javaDup 은 하나로 합친 사건 묶음입니다(파일 머리 주석의 첫 성질). Durs는 합치기 전 사건들의 길이(마이크로초)입니다.
type javaDup struct {
	Fn     string    `json:"fn"`
	Desc   string    `json:"desc"`
	Thread string    `json:"thread"`
	T0     float64   `json:"t0"`
	Durs   []float64 `json:"durs"`
}

// javaStackOnly 는 추적 대상인데 사건 없이 스택에만 보인 메서드입니다. Events는 그 메서드를 감싸는 사건 없이 스택에 품은 사건의 수,
// First·Last는 그 사건들의 시작 시각 범위입니다.
type javaStackOnly struct {
	Fn     string  `json:"fn"`
	Desc   string  `json:"desc"`
	Events int     `json:"events"`
	First  float64 `json:"first"`
	Last   float64 `json:"last"`
}

// springJava 는 앱 쪽 기록 전체입니다.
type springJava struct {
	spans   []javaSpan
	throws  []javaThrow
	parks   []javaPark
	dups    []javaDup
	dropped map[string]int
	traced  func(string) bool
}

// parseISODuration 은 JFR JSON의 길이(PT0.002433417S, PT1M2.5S 꼴)를 나노초로 바꿉니다.
//
// Errors: 꼴이 다르면 오류를 돌려줍니다.
func parseISODuration(s string) (int64, error) {
	rest, ok := strings.CutPrefix(s, "PT")
	if !ok {
		return 0, fmt.Errorf("길이 %q", s)
	}
	var total float64
	for _, unit := range []struct {
		mark string
		ns   float64
	}{{"H", 3600e9}, {"M", 60e9}, {"S", 1e9}} {
		before, after, found := strings.Cut(rest, unit.mark)
		if !found {
			continue
		}
		v, err := strconv.ParseFloat(before, 64)
		if err != nil {
			return 0, fmt.Errorf("길이 %q: %w", s, err)
		}
		total += v * unit.ns
		rest = after
	}
	if rest != "" {
		return 0, fmt.Errorf("길이 %q", s)
	}
	return int64(math.Round(total)), nil
}

// tracedMatcher 는 jdk.MethodTrace 필터(세미콜론으로 나눈 "클래스::메서드" 또는 "클래스")에 메서드 이름이 드는지 가리는 함수를 만듭니다.
func tracedMatcher(filter string) func(string) bool {
	methods, classes := map[string]bool{}, map[string]bool{}
	for _, f := range strings.Split(filter, ";") {
		if f == "" {
			continue
		}
		if strings.Contains(f, "::") {
			methods[f] = true
		} else {
			classes[f] = true
		}
	}
	return func(key string) bool {
		cls, _, _ := strings.Cut(key, "::")
		return methods[key] || classes[cls]
	}
}

// readSpringJFR 은 jfr 명령으로 기록을 JSON으로 받아 앱 쪽 기록을 만듭니다.
//
// Arguments: tool은 jfr 명령, path는 spring.jfr, filter는 기록 때 쓴 메서드 추적 필터(spring.json의 trace), rel은 시각 변환,
// from은 장면이 시작한 시각(start 멈춤 자리의 go, 마이크로초)입니다. 예외는 그 뒤의 것만 담습니다(buildSpringJava).
// Errors: 명령이 실패하거나 JSON·시각·길이를 읽지 못하면 오류를 돌려줍니다.
func readSpringJFR(tool, path, filter string, rel func(uint64) float64, from float64) (*springJava, error) {
	b, err := exec.Command(tool, "print", "--json", "--stack-depth", strconv.Itoa(jfrStackDepth),
		"--events", "jdk.MethodTrace,jdk.ThreadPark,jdk.JavaExceptionThrow", path).Output()
	if err != nil {
		return nil, fmt.Errorf("jfr print: %w", err)
	}
	var doc struct {
		Recording struct {
			Events []jfrEvent `json:"events"`
		} `json:"recording"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("jfr JSON: %w", err)
	}
	return buildSpringJava(doc.Recording.Events, filter, rel, from)
}

// buildSpringJava 는 JFR 사건 목록에서 호출 구간, 예외, 기다림을 만듭니다(readSpringJFR에서 명령 실행을 뺀 부분). 예외는 from 뒤의
// 것만 담습니다. 그 앞은 Spring이 뜨며 리플렉션으로 메서드를 찾아보다 던지고 스스로 받는 것이 대부분이라(2026-10-15 r5-nested-a에서
// 36개 가운데 장면과 관계있는 것 0개) 장면을 가립니다. 뺀 수는 dropped["(장면 앞)"]에 남깁니다. 예외의 자리(At)는 예외 클래스의
// 생성자 칸을 건너뛴 뒤부터라, 첫 칸이 그 예외를 만든 코드입니다.
//
// Errors: 시각이나 길이를 읽지 못하면 오류를 돌려줍니다.
func buildSpringJava(events []jfrEvent, filter string, rel func(uint64) float64, from float64) (*springJava, error) {
	j := &springJava{dropped: map[string]int{}, traced: tracedMatcher(filter)}
	type spanKey struct {
		fn, desc, thread string
		start            uint64
	}
	merged := map[spanKey]int{}
	durs := map[spanKey][]float64{}
	for _, e := range events {
		v := e.Values
		t, err := time.Parse(time.RFC3339Nano, v.StartTime)
		if err != nil {
			return nil, fmt.Errorf("jfr 시각 %q: %w", v.StartTime, err)
		}
		start := uint64(t.UnixNano())
		var frames []jfrFrame
		truncated := false
		if v.StackTrace != nil {
			frames, truncated = v.StackTrace.Frames, v.StackTrace.Truncated
		}
		var dur int64
		if v.Duration != "" {
			if dur, err = parseISODuration(v.Duration); err != nil {
				return nil, err
			}
		}
		switch e.Type {
		case "jdk.MethodTrace":
			if v.Method == nil {
				continue
			}
			k := spanKey{v.Method.key(), v.Method.Descriptor, v.EventThread.JavaName, start}
			durs[k] = append(durs[k], float64(dur)/1e3)
			if i, ok := merged[k]; ok {
				j.spans[i].Twice = true
				j.spans[i].T1 = math.Max(j.spans[i].T1, rel(start+uint64(dur)))
				continue
			}
			s := javaSpan{Fn: k.fn, Desc: k.desc, Thread: k.thread, T0: rel(start), T1: rel(start + uint64(dur)), stack: frames, truncated: truncated}
			if len(frames) > 0 {
				s.Caller = frames[0].ref()
			}
			merged[k] = len(j.spans)
			j.spans = append(j.spans, s)
		case "jdk.JavaExceptionThrow":
			cls := ""
			if v.ThrownClass != nil {
				cls = strings.ReplaceAll(v.ThrownClass.Name, "/", ".")
			}
			if jfrNoise[cls] {
				j.dropped[cls]++
				continue
			}
			if rel(start) < from {
				j.dropped["(장면 앞)"]++
				continue
			}
			th := javaThrow{T: rel(start), Thread: v.EventThread.JavaName, Class: cls}
			if v.Message != nil {
				th.Message = *v.Message
			}
			skip := 0
			for skip < len(frames) && isThrowableInit(frames[skip].Method) {
				skip++
			}
			th.stack = frames[skip:]
			for _, f := range frames[skip:] {
				if len(th.At) == 8 {
					break
				}
				th.At = append(th.At, f.ref())
			}
			j.throws = append(j.throws, th)
		case "jdk.ThreadPark":
			p := javaPark{T0: rel(start), T1: rel(start + uint64(dur)), Thread: v.EventThread.JavaName, stack: frames}
			for k, f := range frames {
				if p.Under == "" && j.traced(f.Method.key()) {
					p.Under = f.ref()
				}
				if k < 6 {
					p.At = append(p.At, f.ref())
				}
			}
			if p.Under == "" {
				j.dropped["jdk.ThreadPark"]++
				continue
			}
			j.parks = append(j.parks, p)
		}
	}
	for k, d := range durs {
		if len(d) > 1 {
			j.dups = append(j.dups, javaDup{Fn: k.fn, Desc: k.desc, Thread: k.thread, T0: rel(k.start), Durs: d})
		}
	}
	sort.Slice(j.dups, func(a, b int) bool { return j.dups[a].T0 < j.dups[b].T0 })
	sort.Slice(j.throws, func(a, b int) bool { return j.throws[a].T < j.throws[b].T })
	sort.Slice(j.parks, func(a, b int) bool { return j.parks[a].T0 < j.parks[b].T0 })

	// 깊이: 스레드마다 시작 순(같으면 긴 것 먼저)으로 놓고, 아직 끝나지 않은 호출의 수를 셉니다.
	sort.SliceStable(j.spans, func(a, b int) bool {
		if j.spans[a].T0 != j.spans[b].T0 {
			return j.spans[a].T0 < j.spans[b].T0
		}
		return j.spans[a].T1 > j.spans[b].T1
	})
	open := map[string][]float64{}
	for i := range j.spans {
		s := &j.spans[i]
		st := open[s.Thread]
		for len(st) > 0 && st[len(st)-1] < s.T0 {
			st = st[:len(st)-1]
		}
		s.Depth = len(st)
		open[s.Thread] = append(st, s.T1)
	}
	return j, nil
}

// isThrowableInit 은 예외 클래스의 생성자인지 돌려줍니다. 이름이 Exception·Error·Throwable로 끝나는 클래스의 <init>을 그렇게 봅니다.
func isThrowableInit(m jfrMethod) bool {
	if m.Name != "<init>" {
		return false
	}
	for _, suffix := range []string{"Exception", "Error", "Throwable"} {
		if strings.HasSuffix(m.Type.Name, suffix) {
			return true
		}
	}
	return false
}

// coverCount 는 thread에서 fn·desc의 기록된 호출 가운데 시각 t를 감싸는 것의 수입니다. self는 세지 않습니다(그 사건 자신).
func (j *springJava) coverCount(thread, fn, desc string, t float64, self *javaSpan) int {
	n := 0
	for i := range j.spans {
		s := &j.spans[i]
		if s != self && s.Thread == thread && s.Fn == fn && s.Desc == desc && s.T0 <= t && t <= s.T1 {
			n++
		}
	}
	return n
}

// stackOnly 는 추적 대상인데 사건 없이 스택에만 보인 호출을 셉니다(파일 머리 주석의 둘째 성질). 사건마다 스택 안에서 같은 메서드가
// 나온 횟수와, 그 사건의 시작을 감싸는 그 메서드의 기록된 호출의 수를 견주어 모자란 만큼을 셉니다. 같은 메서드가 겹쳐 불리면(바깥
// 프록시 안의 안쪽 프록시) 바깥 호출의 기록이 안쪽 호출의 빈자리를 가리기 때문입니다(2026-10-15 r5-nested-a의 안쪽
// TransactionInterceptor.invoke).
func (j *springJava) stackOnly() []javaStackOnly {
	type key struct{ fn, desc string }
	acc := map[key]*javaStackOnly{}
	var order []key
	for i := range j.spans {
		s := &j.spans[i]
		occ := map[key]int{}
		var keys []key
		for _, f := range s.stack {
			k := key{f.Method.key(), f.Method.Descriptor}
			if !j.traced(k.fn) {
				continue
			}
			if occ[k] == 0 {
				keys = append(keys, k)
			}
			occ[k]++
		}
		for _, k := range keys {
			if occ[k] <= j.coverCount(s.Thread, k.fn, k.desc, s.T0, s) {
				continue
			}
			a, ok := acc[k]
			if !ok {
				a = &javaStackOnly{Fn: k.fn, Desc: k.desc, First: s.T0}
				acc[k] = a
				order = append(order, k)
			}
			a.Events++
			a.Last = s.T0
		}
	}
	out := []javaStackOnly{}
	for _, k := range order {
		out = append(out, *acc[k])
	}
	return out
}

// threadAt 은 시각 t를 감싸는 fn 호출의 스레드를 돌려줍니다. 없으면 빈 문자열입니다.
func (j *springJava) threadAt(t float64, fn string) string {
	for _, s := range j.spans {
		if s.Fn == fn && s.T0 <= t && t <= s.T1 {
			return s.Thread
		}
	}
	return ""
}

// springBorrow 는 앱이 풀에서 연결을 빌린 한 번입니다(HikariPool.getConnection(long)). Validated는 그 안에서 살아 있는지 검사했는지
// (PoolBase.isConnectionDead), IdleMs는 빌린 연결 위의 직전 메시지부터 빌리기 시작까지의 시간, Result는 ok 또는 timeout입니다.
type springBorrow struct {
	T0        float64 `json:"t0"`
	T1        float64 `json:"t1"`
	Thread    string  `json:"thread"`
	Validated bool    `json:"validated"`
	Conn      string  `json:"conn"`
	IdleMs    float64 `json:"idleMs"`
	Result    string  `json:"result"`
}

// borrows 는 풀 빌리기마다 어느 연결을 받았는지 연결 위의 메시지로 잇습니다. 검사한 빌리기는 그 안에 빈 문장이 연결 위로 가므로 그
// 연결이고, 검사하지 않은 빌리기는 끝난 뒤 100 ms 안에 앱 메시지가 처음 간 연결입니다. 시간 초과로 끝난 빌리기는 연결이 없습니다.
func (j *springJava) borrows(conns []springConn) []springBorrow {
	out := []springBorrow{}
	for _, s := range j.spans {
		if s.Fn != "com.zaxxer.hikari.pool.HikariPool::getConnection" || !strings.HasPrefix(s.Desc, "(J)") {
			continue
		}
		b := springBorrow{T0: s.T0, T1: s.T1, Thread: s.Thread, Result: "ok"}
		for _, x := range j.spans {
			if x.Thread == s.Thread && x.Fn == "com.zaxxer.hikari.pool.PoolBase::isConnectionDead" && x.T0 >= s.T0 && x.T1 <= s.T1 {
				b.Validated = true
			}
		}
		for _, th := range j.throws {
			if th.Thread == s.Thread && strings.HasSuffix(th.Class, "SQLTransientConnectionException") && th.T >= s.T0 && th.T <= s.T1+100 {
				b.Result = "timeout"
			}
		}
		// 시간 초과로 끝난 빌리기는 연결을 받지 못했으므로 잇지 않습니다.
		best := math.Inf(1)
		if b.Result == "timeout" {
			out = append(out, b)
			continue
		}
		for _, c := range conns {
			for k, m := range c.Messages {
				lo, hi := s.T0, s.T1
				if !b.Validated {
					lo, hi = s.T1, s.T1+100e3
				}
				if m.Dir != "F" || m.T < lo || m.T > hi || m.T >= best {
					continue
				}
				best, b.Conn = m.T, c.Name
				if k > 0 {
					b.IdleMs = (s.T0 - c.Messages[k-1].T) / 1e3
				}
				break
			}
		}
		out = append(out, b)
	}
	return out
}

// view 는 화면과 판정에 넘기는 앱 쪽 기록입니다.
func (j *springJava) view() map[string]any {
	return map[string]any{"spans": orEmpty(j.spans), "throws": orEmpty(j.throws), "parks": orEmpty(j.parks), "dups": orEmpty(j.dups),
		"stackOnly": j.stackOnly(), "dropped": j.dropped}
}

// orEmpty 는 nil 조각을 빈 조각으로 바꿉니다. JSON에서 null 대신 []가 되어 화면이 따로 가리지 않아도 됩니다.
func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
