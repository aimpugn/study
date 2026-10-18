package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// 이 파일은 실험 5 화면의 호출 트리와 렌즈입니다(2026-10-15, WORK 실험 5 「실험 5 화면 설계」). 화면은 여기서 낸 값을 고르고
// 그리기만 합니다.
//
// 트리의 마디는 JFR이 기록한 호출과, 기록 없이 다른 사건의 스택에만 보인 호출(지나가는 예외로 끝남, spring_jfr.go 머리 주석)입니다.
// 깊이는 이 트리에서의 깊이 하나만 씁니다. 추적하지 않은 스택 칸(프록시, 반사, 감싸개)은 마디가 아니라 두 마디 사이의 Hidden으로
// 접습니다. javaSpan.Depth(기록된 호출만 셈)는 화면에 쓰지 않습니다. 한 화면에 깊이 숫자가 두 가지면 읽는 사람이 헷갈립니다.

// sceneThread 는 장면 메서드가 도는 스레드입니다. Main.main이 장면의 바깥 메서드를 직접 부르고, 이번 장면들은 다른 스레드에서
// 트랜잭션을 열지 않습니다.
const sceneThread = "main"

// selfLensUs 는 자기 시간 렌즈의 문턱(마이크로초)입니다. 1 ms는 시안(lens.py)에서 렌즈를 켰을 때 남는 줄이 풀 기다림, 서버 답
// 기다림, 연결 검사처럼 시간이 실제로 든 자리였던 처음 값이고, 더 나은 값인지는 확인하지 않았습니다.
const selfLensUs = 1000

// 끝난 꼴. 정상과 예외 둘로 묶을 때는 endReturn만 정상입니다(isExceptionEnd).
const (
	endReturn = "return" // 예외의 흔적이 없음. 이미 있는 예외를 다시 던진 것은 JFR 사건이 없어 여기에 듭니다
	endThrow  = "throw"  // 마지막 자식 뒤에 이 마디의 코드가 예외를 만듦(예외 사건 스택의 가장 안쪽 추적 칸이 이 마디)
	endTwice  = "twice"  // 같은 시작의 사건 둘: 메서드 안의 throw를 같은 메서드의 finally가 받은 뒤 끝남
	endPassed = "passed" // 스택에만 보임: 아래에서 올라온 예외가 지나가며 끝남
	// endRethrow 는 끝 기록은 있지만 부모가 스택에만 보인(endPassed) 마지막 자식입니다. 부모를 지나간 예외는 그때 열려 있던 자식,
	// 곧 마지막 자식에서 올라왔다고 추론합니다(이미 있는 예외를 다시 던지면 사건이 없어 endReturn으로 보이므로). 마지막 자식이 정상으로
	// 끝난 뒤 부모가 추적하지 않는 코드에서 예외를 받았다면 이 추론은 틀립니다(2026-10-15 r5-requires-new-pool1의
	// handleExistingTransaction이 resumeAfterBeginException 뒤 다시 던지는 자리, AbstractPlatformTransactionManager 455~456행).
	endRethrow = "rethrow"
)

// 렌즈 이유. 갈림은 기록끼리 견줘야 알 수 있어 spring-compare가 붙입니다.
const (
	whyEffect = "효과"
	whyBranch = "갈림"
	whyTime   = "시간"
)

// callNode 는 호출 트리의 마디 하나입니다.
//
// T1이 nil이면 스택에만 보인 호출입니다. 그때 T0은 처음 보인 시각(실제 시작은 그보다 앞)이고, Seen은 그 안의 기록 가운데 가장 늦은
// 끝(그때까지는 열려 있었음)입니다. Site는 부모 메서드 안에서 이 호출로 이어진 줄이고, Hidden은 부모와 이 마디 사이의 추적하지 않은
// 스택 칸("클래스::메서드:줄", 바깥부터)이며 HiddenKinds는 그 칸들의 종류(foldKind)입니다. 뿌리 마디의 Hidden은 그 바깥의 칸
// 전부입니다. Self는 자기 시간(마이크로초)이고 셈하지 못하면 nil입니다(annotateTree).
type callNode struct {
	ID          int          `json:"id"`
	Parent      int          `json:"parent"`
	Fn          string       `json:"fn"`
	Desc        string       `json:"desc"`
	Label       string       `json:"label"`
	Lane        string       `json:"lane"`
	Depth       int          `json:"depth"`
	T0          float64      `json:"t0"`
	T1          *float64     `json:"t1"`
	Seen        float64      `json:"seen,omitempty"`
	Site        int          `json:"site"`
	Hidden      []string     `json:"hidden"`
	HiddenKinds []string     `json:"hiddenKinds"`
	End         string       `json:"end"`
	Self        *float64     `json:"self"`
	Kids        []int        `json:"kids"`
	Effects     []callEffect `json:"effects"`
	Why         []string     `json:"why"`
	Branch      *nodeBranch  `json:"branch,omitempty"`
	span        *javaSpan
}

// last 는 마디가 열려 있던 것으로 아는 마지막 시각입니다. 끝 시각이 없으면 Seen입니다.
func (n *callNode) last() float64 {
	if n.T1 != nil {
		return *n.T1
	}
	return n.Seen
}

// callEffect 는 마디가 밖에 드러나게 한 일 하나입니다. Kind마다 쓰는 칸이 다릅니다.
//
//   - wire: 연결 위로 요청을 보냄. Conn과 Seq(그 연결의 springRequest.Seq들)
//   - borrow: 풀에서 연결을 빌림. Conn(시간 초과면 빈 문자열), Result, IdleMs, Validated(springBorrow와 같은 뜻)
//   - return: 연결을 풀에 돌려줌(ProxyConnection.close). Conn은 직전에 스레드에서 풀린 연결(bindingReplay)
//   - suspend, resume: 스레드에서 연결을 떼어 둠, 다시 묶음. Conn
//   - autocommit: 자동 커밋을 바꿈. Value는 부른 줄의 소스 글에서 읽은 값(읽지 못하면 빈 문자열), Sent는 그 안에서 보낸 요청 수
//   - park: 스레드가 잠듦(jdk.ThreadPark). T0, T1, At은 잠든 자리의 추적 대상 칸
//   - throw: 예외를 만듦(jdk.JavaExceptionThrow). T0, Class, Message, At은 만든 자리의 추적 대상 칸
//   - rollbackOnly: 지금 트랜잭션에 롤백 전용 표시를 함(DataSourceTransactionManager.doSetRollbackOnly). 연결 위로는 아무것도
//     보내지 않지만 그 뒤의 커밋을 롤백으로 바꾸는 상태 변화라 효과로 셉니다(사용성 시험 C4, 기본 화면에서 접혀 찾지 못했음)
//   - validate: 빌린 연결이 살아 있는지 검사함(PoolBase.isConnectionDead). 연결 위의 빈 문장 왕복이 이 검사에서 나옵니다(C5)
type callEffect struct {
	Kind      string  `json:"kind"`
	Conn      string  `json:"conn,omitempty"`
	Seq       []int   `json:"seq,omitempty"`
	Result    string  `json:"result,omitempty"`
	IdleMs    float64 `json:"idleMs,omitempty"`
	Validated bool    `json:"validated,omitempty"`
	Value     string  `json:"value,omitempty"`
	Sent      int     `json:"sent,omitempty"`
	T0        float64 `json:"t0,omitempty"`
	T1        float64 `json:"t1,omitempty"`
	At        string  `json:"at,omitempty"`
	Class     string  `json:"class,omitempty"`
	Message   string  `json:"message,omitempty"`
}

// callTree 는 한 스레드의 호출 트리입니다. Nodes[i].ID == i이고, Roots는 시작 순입니다. Ambiguous는 스택 칸 하나에 맞는 기록이 둘
// 이상이었던 수입니다. 0이 아니면 맞춤의 셋째 조건(buildCallTree)이 그 기록을 가르지 못한 것이므로 트리를 믿지 않습니다.
type callTree struct {
	Nodes     []*callNode
	Roots     []int
	Ambiguous int
}

// buildCallTree 는 thread의 호출 가운데 [from, to]와 겹치는 것으로 호출 트리를 만듭니다.
//
// 스택 칸과 기록의 맞춤: 호출 s의 스택 칸 f에 맞는 기록 X는 (1) f와 이름·서술자가 같고, (2) s의 시작을 감싸며, (3) X의 스택이 s의
// 스택에서 f 바깥 부분과 같은 것입니다(줄 번호까지, sameFrames). 셋째 조건이 없으면 같은 메서드가 겹쳐 불렸는데 바깥 호출의
// 기록이 빠진 자리에서 안쪽 기록이 바깥 칸에 붙습니다(2026-10-15 r5-required의 invokeWithinTransaction, WORK 「실험 5 화면 설계」).
// 맞는 기록이 없는 추적 대상 칸은 스택에만 보인 마디이고, (부모, 이름, 서술자, 부른 줄)이 같으면 같은 마디로 봅니다. 그래서 같은
// 부모의 같은 줄에서 두 번 불린 스택에만 보인 호출은 한 마디로 합쳐집니다(이번 기록에는 없음).
//
// Arguments: j는 앱 쪽 기록(spans는 buildSpringJava가 시작 순으로 정렬해 둠), thread는 스레드 이름, from·to는 장면 구간(마이크로초)입니다.
// Returns: 깊이, 자식, 부른 줄, 숨은 칸, 칸, 이름표까지 채운 트리입니다. 끝난 꼴, 자기 시간, 효과, 렌즈는 annotateTree가 채웁니다.
func buildCallTree(j *springJava, thread string, from, to float64) *callTree {
	t := &callTree{}
	var spans []*javaSpan
	byKey := map[[2]string][]*javaSpan{}
	for i := range j.spans {
		s := &j.spans[i]
		if s.Thread == thread && s.T1 >= from && s.T0 <= to {
			spans = append(spans, s)
			k := [2]string{s.Fn, s.Desc}
			byKey[k] = append(byKey[k], s)
		}
	}

	// 마디를 만들고 부모의 자식으로 잇습니다. 기록된 호출은 처음 만날 때 한 번만 만듭니다.
	newNode := func(fn, desc string, parent, site int, hidden []string) *callNode {
		n := &callNode{ID: len(t.Nodes), Parent: parent, Fn: fn, Desc: desc, Site: site, Hidden: orEmpty(hidden), Kids: []int{}}
		t.Nodes = append(t.Nodes, n)
		if parent >= 0 {
			t.Nodes[parent].Kids = append(t.Nodes[parent].Kids, n.ID)
		} else {
			t.Roots = append(t.Roots, n.ID)
		}
		return n
	}
	spanNode := map[*javaSpan]int{}
	nodeOfSpan := func(s *javaSpan, parent, site int, hidden []string) int {
		if id, ok := spanNode[s]; ok {
			return id
		}
		n := newNode(s.Fn, s.Desc, parent, site, hidden)
		t1 := s.T1
		n.span, n.T0, n.T1 = s, s.T0, &t1
		spanNode[s] = n.ID
		return n.ID
	}
	type soKey struct {
		parent   int
		fn, desc string
		site     int
	}
	soNode := map[soKey]int{}

	// 맞춤의 세 조건으로 칸 f에 맞는 기록을 찾습니다. 둘 이상이면 Ambiguous에 세고 가장 늦게 시작한 것(가장 안쪽)을 씁니다.
	match := func(f jfrFrame, outer []jfrFrame, at float64, self *javaSpan) *javaSpan {
		var found *javaSpan
		for _, x := range byKey[[2]string{f.Method.key(), f.Method.Descriptor}] {
			if x == self || x.T0 > at || x.T1 < at || !sameFrames(x.stack, x.truncated, outer, self.truncated) {
				continue
			}
			if found != nil {
				t.Ambiguous++
				if x.T0 <= found.T0 {
					continue
				}
			}
			found = x
		}
		return found
	}

	// 호출마다 스택을 바깥부터 걸으며 추적 대상 칸을 마디로 바꾸고, 그 사이의 칸은 다음 마디의 Hidden으로 모읍니다.
	// 칸의 줄은 그 메서드가 안쪽을 부른 줄이므로, 다음 마디의 Site가 됩니다.
	for _, s := range spans {
		parent, site := -1, 0
		var hidden []string
		for idx := len(s.stack) - 1; idx >= 0; idx-- {
			f := s.stack[idx]
			if !j.traced(f.Method.key()) {
				hidden = append(hidden, f.ref())
				continue
			}
			var id int
			if x := match(f, s.stack[idx+1:], s.T0, s); x != nil {
				id = nodeOfSpan(x, parent, site, hidden)
			} else {
				k := soKey{parent, f.Method.key(), f.Method.Descriptor, site}
				var ok bool
				if id, ok = soNode[k]; !ok {
					n := newNode(k.fn, k.desc, parent, site, hidden)
					n.T0 = s.T0
					id, soNode[k] = n.ID, n.ID
				}
				n := t.Nodes[id]
				n.T0, n.Seen = math.Min(n.T0, s.T0), math.Max(n.Seen, s.T1)
			}
			parent, site, hidden = id, f.Line, nil
		}
		nodeOfSpan(s, parent, site, hidden)
	}

	// 자식과 뿌리를 시작 순으로 놓고, 깊이·칸·숨은 칸의 종류·이름표를 채웁니다.
	byStart := func(ids []int) {
		sort.SliceStable(ids, func(a, b int) bool { return t.Nodes[ids[a]].T0 < t.Nodes[ids[b]].T0 })
	}
	byStart(t.Roots)
	var walk func(id, depth int)
	walk = func(id, depth int) {
		n := t.Nodes[id]
		n.Depth, n.Lane = depth, laneOf(n.Fn)
		n.HiddenKinds = []string{}
		for _, h := range n.Hidden {
			if k := foldKind(h); !contains(n.HiddenKinds, k) {
				n.HiddenKinds = append(n.HiddenKinds, k)
			}
		}
		byStart(n.Kids)
		for _, k := range n.Kids {
			walk(k, depth+1)
		}
	}
	for _, r := range t.Roots {
		walk(r, 0)
	}
	labelNodes(t.Nodes)
	return t
}

// sameFrames 는 두 스택(안쪽부터)이 같은 호출 자리인지 돌려줍니다(메서드, 서술자, 줄). JFR은 스택을 자를 때 바깥 칸을 버리므로, 어느
// 쪽도 잘리지 않았으면 길이까지 같아야 하고, 한쪽이라도 잘렸으면 겹치는 앞부분만 견줍니다. 길이를 보지 않으면 짧은 스택이 긴 스택의
// 앞부분과 우연히 같을 때(바깥 호출이 빠진 자리) 잘못 맞춥니다(TestBuildCallTreeAlignsByOuterStack).
func sameFrames(a []jfrFrame, aCut bool, b []jfrFrame, bCut bool) bool {
	if !aCut && !bCut && len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i].Method.key() != b[i].Method.key() || a[i].Method.Descriptor != b[i].Method.Descriptor || a[i].Line != b[i].Line {
			return false
		}
	}
	return true
}

// laneOf 는 호출이 그려질 칸입니다. 칸은 라이브러리의 역할로 나눕니다. DataSourceTransactionManager처럼 spring-jdbc에 있어도
// 트랜잭션을 다루는 클래스는 Spring 트랜잭션 칸이고, JdbcTemplate만 JdbcTemplate 칸입니다.
func laneOf(fn string) string {
	switch {
	case strings.HasPrefix(fn, "lab."):
		return "app"
	case strings.HasPrefix(fn, "org.springframework.jdbc.core."):
		return "jdbc"
	case strings.HasPrefix(fn, "org.springframework."):
		return "tx"
	case strings.HasPrefix(fn, "com.zaxxer.hikari."):
		return "hikari"
	case strings.HasPrefix(fn, "org.postgresql."):
		return "pg"
	}
	return "other"
}

// foldKind 는 추적하지 않은 스택 칸("클래스::메서드:줄")을 접을 때 붙이는 종류입니다.
func foldKind(ref string) string {
	fn, _, _ := strings.Cut(ref, "::")
	switch {
	case strings.HasPrefix(fn, "lab.Main"):
		return "앱 main"
	case strings.Contains(fn, "$$") || strings.HasPrefix(fn, "0x") ||
		hasAnyPrefix(fn, "java.lang.reflect.", "jdk.internal.reflect.", "org.springframework.aop.", "org.springframework.cglib."):
		return "프록시·반사"
	case strings.HasPrefix(fn, "com.zaxxer."):
		return "HikariCP 감싸개"
	case strings.HasPrefix(fn, "org.postgresql."):
		return "pgjdbc 안쪽"
	case strings.HasPrefix(fn, "org.springframework.jdbc."):
		return "JdbcTemplate 콜백"
	case strings.HasPrefix(fn, "org.springframework."):
		return "Spring 안쪽"
	case hasAnyPrefix(fn, "java.util.concurrent.", "jdk.internal.misc."):
		return "JDK 동시성"
	case hasAnyPrefix(fn, "java.", "jdk.", "sun."):
		return "JDK"
	}
	return "기타"
}

// hasAnyPrefix 는 s가 prefixes 가운데 하나로 시작하는지 돌려줍니다.
func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// labelNodes 는 마디마다 화면에 쓸 짧은 이름("클래스.메서드")을 붙입니다. 같은 이름의 메서드가 서술자가 다른 채로 트리에 둘 이상
// 있으면(오버로드) 인자 형을 붙여 가릅니다(HikariPool.getConnection()과 getConnection(long)).
func labelNodes(nodes []*callNode) {
	descs := map[string]map[string]bool{}
	for _, n := range nodes {
		if descs[n.Fn] == nil {
			descs[n.Fn] = map[string]bool{}
		}
		descs[n.Fn][n.Desc] = true
	}
	for _, n := range nodes {
		n.Label = javaLabel(n.Fn, n.Desc, len(descs[n.Fn]) > 1)
	}
}

// javaLabel 은 "패키지.클래스::메서드"를 "클래스.메서드"로 줄입니다. params가 참이면 서술자의 인자 형을 괄호 안에 붙입니다.
func javaLabel(fn, desc string, params bool) string {
	cls, m, _ := strings.Cut(fn, "::")
	if i := strings.LastIndexByte(cls, '.'); i >= 0 {
		cls = cls[i+1:]
	}
	label := cls + "." + m
	if params {
		label += "(" + strings.Join(descParams(desc), ", ") + ")"
	}
	return label
}

// descParams 는 JVM 메서드 서술자("(JLjava/lang/String;[I)V")의 인자 형을 짧은 이름(long, String, int[])으로 돌려줍니다.
func descParams(desc string) []string {
	open, close := strings.IndexByte(desc, '('), strings.IndexByte(desc, ')')
	if open < 0 || close < open {
		return nil
	}
	prim := map[byte]string{'B': "byte", 'C': "char", 'D': "double", 'F': "float", 'I': "int", 'J': "long", 'S': "short", 'Z': "boolean"}
	var out []string
	s := desc[open+1 : close]
	for len(s) > 0 {
		dims := 0
		for len(s) > 0 && s[0] == '[' {
			dims++
			s = s[1:]
		}
		var name string
		if s[0] == 'L' {
			end := strings.IndexByte(s, ';')
			if end < 0 {
				return out
			}
			name = s[1:end]
			if i := strings.LastIndexByte(name, '/'); i >= 0 {
				name = name[i+1:]
			}
			s = s[end+1:]
		} else {
			name = prim[s[0]]
			s = s[1:]
		}
		out = append(out, name+strings.Repeat("[]", dims))
	}
	return out
}

// bindState 는 스레드의 연결 묶임이 바뀐 때와 그 뒤의 상태입니다. Bound는 스레드에 묶인 연결, Held는 떼어 둔 연결(먼저 떼어 둔 것부터)입니다.
type bindState struct {
	T     float64  `json:"t"`
	Bound string   `json:"bound"`
	Held  []string `json:"held"`
	By    int      `json:"by"`
}

// poolSlot 은 풀의 연결 하나를 이 스레드가 쓴 구간입니다. T0은 빌리기가 끝난 때, T1은 돌려준 때이고 장면 안에 돌려주지 않았으면 0입니다.
type poolSlot struct {
	Conn string  `json:"conn"`
	T0   float64 `json:"t0"`
	T1   float64 `json:"t1"`
}

// bindingReplay 는 스레드에 묶인 연결과 떼어 둔 연결, 풀의 연결을 쓴 구간을 호출 순서로 다시 셉니다(관찰한 값이 아니라 셈한 값).
//
// Spring(6.2.15)은 새 트랜잭션의 연결을 doBegin에서 묶고(DataSourceTransactionManager 311행), doSuspend에서 떼어 두며(328행),
// doResume에서 다시 묶고(333행), doCleanupAfterCompletion의 첫머리에서 풀고(382행) 끝에서 돌려줍니다(403행 → ProxyConnection.close).
// 묶는 줄은 추적하지 않으므로 그 doBegin 안의 빌리기가 끝난 때를, 푸는 줄은 doCleanupAfterCompletion이 시작한 때를 씁니다.
//
// Arguments: t는 호출 트리, borrows는 풀 빌리기(Conn이 정해진 것), from은 장면 시작입니다.
// Returns: 묶임이 바뀔 때마다의 상태(첫 칸은 from의 빈 상태)와, 풀의 연결을 쓴 구간입니다. 돌려주기(close)와 suspend·resume 마디에는
// 옮긴 연결을 효과로 붙입니다.
func bindingReplay(t *callTree, borrows []springBorrow, from float64) ([]bindState, []poolSlot) {
	type step struct {
		t    float64
		kind string
		conn string
		node int
	}
	var steps []step
	for _, b := range borrows {
		if b.Thread == sceneThread && b.Result == "ok" && b.Conn != "" && b.T0 >= from {
			steps = append(steps, step{b.T1, "bind", b.Conn, -1})
		}
	}
	for _, n := range t.Nodes {
		switch {
		case strings.HasSuffix(n.Fn, "DataSourceTransactionManager::doSuspend"):
			steps = append(steps, step{n.T0, "suspend", "", n.ID})
		case strings.HasSuffix(n.Fn, "DataSourceTransactionManager::doResume"):
			steps = append(steps, step{n.T0, "resume", "", n.ID})
		case strings.HasSuffix(n.Fn, "DataSourceTransactionManager::doCleanupAfterCompletion"):
			steps = append(steps, step{n.T0, "unbind", "", n.ID})
		case strings.HasSuffix(n.Fn, "ProxyConnection::close"):
			steps = append(steps, step{n.T0, "close", "", n.ID})
		}
	}
	sort.SliceStable(steps, func(a, b int) bool { return steps[a].t < steps[b].t })

	// 상태를 차례로 바꾸고, 바뀔 때마다 한 칸씩 남깁니다. 풀린 연결은 돌려주기까지 released에 둡니다.
	states := []bindState{{T: from, Held: []string{}, By: -1}}
	var slots []poolSlot
	open := map[string]int{}
	bound, released := "", ""
	var held []string
	for _, s := range steps {
		switch s.kind {
		case "bind":
			bound = s.conn
			open[s.conn] = len(slots)
			slots = append(slots, poolSlot{Conn: s.conn, T0: s.t})
		case "suspend":
			held = append(held, bound)
			t.Nodes[s.node].Effects = append(t.Nodes[s.node].Effects, callEffect{Kind: "suspend", Conn: bound})
			bound = ""
		case "resume":
			if len(held) > 0 {
				bound, held = held[len(held)-1], held[:len(held)-1]
			}
			t.Nodes[s.node].Effects = append(t.Nodes[s.node].Effects, callEffect{Kind: "resume", Conn: bound})
		case "unbind":
			released, bound = bound, ""
		case "close":
			t.Nodes[s.node].Effects = append(t.Nodes[s.node].Effects, callEffect{Kind: "return", Conn: released})
			if k, ok := open[released]; ok {
				slots[k].T1 = s.t
				delete(open, released)
			}
			released = ""
			continue
		}
		states = append(states, bindState{T: s.t, Bound: bound, Held: append([]string{}, held...), By: s.node})
	}
	return states, orEmpty(slots)
}

// serverWait 는 서버의 백엔드가 다른 트랜잭션을 기다린 구간 하나입니다(XactLockTableWait부터 오류 또는 요청의 끝까지).
type serverWait struct {
	Conn string  `json:"conn"`
	Seq  int     `json:"seq"`
	T0   float64 `json:"t0"`
	T1   float64 `json:"t1"`
	Xid  string  `json:"xid"`
	End  string  `json:"end"`
}

// serverWaits 는 연결마다 요청의 서버 사건에서 기다림 구간을 찾습니다. 오류(ereport)로 끝났으면 End는 error, 아니면 done입니다.
func serverWaits(conns []springConn) []serverWait {
	out := []serverWait{}
	for _, c := range conns {
		for _, r := range c.Requests {
			if r.Server == nil {
				continue
			}
			var w *serverWait
			for _, e := range r.Server.Events {
				switch {
				case e.Kind == "xact_wait" && w == nil:
					w = &serverWait{Conn: c.Name, Seq: r.Seq, T0: e.T, T1: r.Done, Xid: e.F["xid"], End: "done"}
				case e.Kind == "ereport" && w != nil && w.End == "done":
					w.T1, w.End = e.T, "error"
				}
			}
			if w != nil {
				out = append(out, *w)
			}
		}
	}
	return out
}

// waitCycle 은 서버의 기다림 하나를 앱 쪽 사실과 이은 기다림 관계입니다(WORK 실험 5 「사용성 시험」 C2: “DB는 왜 교착으로 보지
// 않았나”를 화면이 조각으로만 보여, 참가자가 이유를 추측으로 이었음). 화살표는 셋입니다.
//
//   - Thread → Waiter: 앱 스레드가 Waiter 연결의 응답을 기다린다. Call 마디(그 요청을 보낸 QueryExecutorImpl.execute)가 기다림 내내 열려 있음(JFR 기록)
//   - Waiter → Xid: Waiter의 백엔드가 Xid 트랜잭션의 끝을 기다린다(서버 함수 기록의 XactLockTableWait)
//   - Xid(Holder) → HeldBy: Xid를 받은 연결 Holder를 HeldBy 스레드가 떼어 둔 채 들고 있다(bindingReplay에서 셈, 추론)
//
// Cycle은 HeldBy가 Thread와 같아 앱 스레드를 거쳐 고리가 닫히는지입니다. 서버가 보는 것은 가운데 화살표뿐이므로, 고리가 닫혀도
// 서버의 교착 검사는 고리를 찾지 못합니다(이 해석은 화면에 추론으로 표시).
//
// Kind가 "pool"이면 서버가 아니라 풀의 기다림입니다(poolWaitCycles). 그때 Waiter·Xid는 비고, Holder는 그 순간 풀의 칸을 모두
// 차지한 연결, Pool은 풀의 최대 칸 수, Call은 빈 칸을 기다린 ConcurrentBag.borrow 마디입니다.
type waitCycle struct {
	Kind   string  `json:"kind"`
	Pool   int     `json:"pool,omitempty"`
	T0     float64 `json:"t0"`
	T1     float64 `json:"t1"`
	Waiter string  `json:"waiter"`
	Xid    string  `json:"xid"`
	Holder string  `json:"holder"`
	Thread string  `json:"thread"`
	Call   int     `json:"call"`
	HeldBy string  `json:"heldBy"`
	Cycle  bool    `json:"cycle"`
}

// waitCycles 는 서버의 기다림마다 waitCycle을 만듭니다.
//
// Arguments: t는 호출 트리(효과까지 붙인 것), waits는 서버의 기다림, binding은 bindingReplay의 묶임 상태, conns는 연결입니다.
// Returns: 기다림마다 하나입니다. Holder를 찾지 못하면 빈 글, 기다림 내내 열려 있던 보낸 마디가 없으면 Call은 -1입니다.
func waitCycles(t *callTree, waits []serverWait, binding []bindState, conns []springConn) []waitCycle {
	out := []waitCycle{}
	for _, w := range waits {
		c := waitCycle{Kind: "server", T0: w.T0, T1: w.T1, Waiter: w.Conn, Xid: w.Xid, Call: -1}

		// 그 번호를 서버에서 받은 연결.
		for _, cn := range conns {
			for _, r := range cn.Requests {
				if r.Server == nil {
					continue
				}
				for _, e := range r.Server.Events {
					if e.Kind == "xid" && e.F["xid"] == w.Xid {
						c.Holder = cn.Name
					}
				}
			}
		}

		// 기다리는 요청을 보낸 마디. 그 마디가 기다림 내내 열려 있었을 때만(끝 기록이 기다림의 끝 뒤) 장면 스레드가 그 연결의 응답을
		// 기다렸다고 셉니다. 끝 기록이 없는 마디(스택에만 보임)는 열려 있던 구간을 몰라 세지 않습니다.
		for _, n := range t.Nodes {
			if n.T0 > w.T0 || n.T1 == nil || *n.T1 < w.T1 {
				continue
			}
			for _, e := range n.Effects {
				if e.Kind == "wire" && e.Conn == w.Conn && containsInt(e.Seq, w.Seq) {
					c.Call, c.Thread = n.ID, sceneThread
				}
			}
		}

		// 기다림이 시작될 때 Holder를 떼어 둔 채 들고 있던 스레드. 묶임 상태는 장면 스레드 것뿐입니다.
		var st *bindState
		for i := range binding {
			if binding[i].T <= w.T0 {
				st = &binding[i]
			}
		}
		if st != nil && c.Holder != "" && contains(st.Held, c.Holder) {
			c.HeldBy = sceneThread
		}
		c.Cycle = c.Call >= 0 && c.HeldBy != "" && c.HeldBy == c.Thread
		out = append(out, c)
	}
	return out
}

// poolWaitCycles 는 풀에서 빈 칸을 기다리다 시간 초과로 끝난 빌리기마다, 그 순간 풀의 칸을 누가 쥐고 있었는지를 잇습니다
// (사용성 시험 2차: 풀 1칸 장면에서 “자기가 쥔 연결을 자기가 기다린다”를 참가자가 추측으로만 이었음).
//
// 칸 사용(slots)은 bindingReplay가 장면 스레드의 빌리기~돌려주기로 셈한 것이라, 그 순간 쓰이는 칸이 pool개 이상이면 풀의 칸을
// 모두 장면 스레드가 쥔 것입니다. 그 연결이 묶임 상태에서 떼어 둔 연결(Held)이면 HeldBy는 장면 스레드이고, 빌리기를 기다리는
// 스레드도 장면 스레드이므로 Cycle입니다. 다른 스레드의 빌리기는 셈에 없어, 다른 스레드가 칸을 쥔 경우는 고리로 세지 않습니다.
//
// Arguments: t는 효과를 붙인 호출 트리, slots·binding은 bindingReplay의 결과, pool은 풀의 최대 칸 수(0이면 모름)입니다.
// Returns: 시간 초과로 끝난 빌리기마다 하나입니다. T0·T1은 그 빌리기 마디의 시작과 끝입니다.
func poolWaitCycles(t *callTree, slots []poolSlot, binding []bindState, pool int) []waitCycle {
	out := []waitCycle{}
	for _, n := range t.Nodes {
		timedOut := false
		for _, e := range n.Effects {
			timedOut = timedOut || e.Kind == "borrow" && e.Result == "timeout"
		}
		if !timedOut || n.T1 == nil {
			continue
		}
		c := waitCycle{Kind: "pool", Pool: pool, T0: n.T0, T1: *n.T1, Thread: sceneThread, Call: n.ID}

		// 빌리기가 시작될 때 쓰이던 칸과, 그때 떼어 둔 연결.
		var inUse []string
		for _, s := range slots {
			if s.T0 <= n.T0 && (s.T1 == 0 || s.T1 > n.T0) {
				inUse = append(inUse, s.Conn)
			}
		}
		var st *bindState
		for i := range binding {
			if binding[i].T <= n.T0 {
				st = &binding[i]
			}
		}
		c.Holder = strings.Join(inUse, ", ")
		allHeld := len(inUse) > 0 && st != nil
		for _, conn := range inUse {
			allHeld = allHeld && contains(st.Held, conn)
		}
		if allHeld {
			c.HeldBy = sceneThread
		}
		c.Cycle = pool > 0 && len(inUse) >= pool && allHeld
		out = append(out, c)
	}
	return out
}

// poolConfig 는 소스에서 읽은 풀 설정입니다(사용성 시험 1차: 참가자가 2초·0.5초가 어디서 왔는지 화면에서 찾지 못했음). 값이 리터럴로
// 적힌 줄만 읽고, 줄을 "클래스:줄"로 남깁니다.
//
//   - ConnectionTimeout: 앱 설정(lab.LabConfig)의 setConnectionTimeout. 풀에서 연결을 받지 못하고 기다리는 한도
//   - AliveBypass: HikariCP(HikariPool)의 aliveBypassWindowMs 기본값. 연결이 쉰 시간이 이보다 짧으면 살아 있는지 검사하지 않음.
//     AliveCheckAt은 그 값으로 검사를 건너뛰는 조건의 줄입니다. 시스템 속성 com.zaxxer.hikari.aliveBypassWindowMs로 바꿀 수 있으나
//     txlab은 그 속성을 주지 않습니다(txlab/spring.go의 java 인자)
type poolConfig struct {
	ConnectionTimeoutMs int    `json:"connectionTimeoutMs,omitempty"`
	ConnectionTimeoutAt string `json:"connectionTimeoutAt,omitempty"`
	AliveBypassMs       int    `json:"aliveBypassMs,omitempty"`
	AliveBypassAt       string `json:"aliveBypassAt,omitempty"`
	AliveCheckAt        string `json:"aliveCheckAt,omitempty"`
}

// connTimeoutArg 는 LabConfig의 "c.setConnectionTimeout(2000);"에서 값을 읽습니다.
var connTimeoutArg = regexp.MustCompile(`setConnectionTimeout\((\d+)\)`)

// aliveBypassDefault 는 HikariPool의 "aliveBypassWindowMs = Long.getLong(\"…\", MILLISECONDS.toMillis(500));"에서 기본값을 읽습니다.
var aliveBypassDefault = regexp.MustCompile(`aliveBypassWindowMs = Long\.getLong\("[^"]+", MILLISECONDS\.toMillis\((\d+)\)\)`)

// hikariPool 은 aliveBypassWindowMs가 있는 HikariCP 클래스입니다.
const hikariPool = "com.zaxxer.hikari.pool.HikariPool"

// poolConfigOf 는 src에서 풀 설정을 읽습니다. 파일이 없거나 리터럴이 없는 값은 비어 있습니다.
func poolConfigOf(src *javaSources) poolConfig {
	var pc poolConfig
	for i, l := range src.lines("lab.LabConfig") {
		if m := connTimeoutArg.FindStringSubmatch(l); m != nil {
			pc.ConnectionTimeoutMs, pc.ConnectionTimeoutAt = atoiOr(m[1], 0), fmt.Sprintf("lab.LabConfig:%d", i+1)
		}
	}
	for i, l := range src.lines(hikariPool) {
		if m := aliveBypassDefault.FindStringSubmatch(l); m != nil {
			pc.AliveBypassMs, pc.AliveBypassAt = atoiOr(m[1], 0), fmt.Sprintf("%s:%d", hikariPool, i+1)
		}
		if strings.Contains(l, "> aliveBypassWindowMs") {
			pc.AliveCheckAt = fmt.Sprintf("%s:%d", hikariPool, i+1)
		}
	}
	return pc
}

// javaSources 는 Java 소스를 찾는 뿌리 디렉터리들입니다(그 아래에 패키지 경로가 있음). 파일은 한 번씩만 읽습니다.
type javaSources struct {
	roots []string
	files map[string][]string
}

// fileOf 는 클래스 이름("패키지.클래스" 또는 "패키지.클래스::메서드", 안쪽 클래스는 $)의 소스 상대경로입니다.
func fileOf(class string) string {
	class, _, _ = strings.Cut(class, "::")
	class, _, _ = strings.Cut(class, "$")
	return strings.ReplaceAll(class, ".", "/") + ".java"
}

// lines 는 그 클래스의 소스 줄들입니다. 뿌리 가운데 처음 찾은 파일을 쓰고, 없으면 nil입니다.
func (js *javaSources) lines(class string) []string {
	if js == nil {
		return nil
	}
	rel := fileOf(class)
	if text, ok := js.files[rel]; ok {
		return text
	}
	var text []string
	for _, r := range js.roots {
		if b, err := os.ReadFile(filepath.Join(r, rel)); err == nil {
			text = strings.Split(string(b), "\n")
			break
		}
	}
	js.files[rel] = text
	return text
}

// line 은 그 클래스 소스의 n번째 줄입니다(1부터). 없으면 빈 문자열입니다.
func (js *javaSources) line(class string, n int) string {
	text := js.lines(class)
	if n < 1 || n > len(text) {
		return ""
	}
	return text[n-1]
}

// checkAppSites 는 앱 코드(lab.*)가 다음 칸을 부른 줄의 소스에 그 칸의 메서드 이름이 실제로 있는지 대조합니다. JFR 스택의 줄 번호는
// 기록 때 컴파일한 클래스의 것이라, 읽은 소스가 그 뒤에 바뀌었으면 화면이 엉뚱한 줄을 칠합니다(2026-10-15, 머리 주석이 한 줄 늘어
// 앞선 기록 열두 벌이 모두 한 줄씩 어긋났음). 그런 어긋남을 화면에 알리려는 검사입니다.
//
// 대조하는 쌍: 부모(앱 코드)의 부른 줄 → 다음 칸, 추적하지 않은 칸 가운데 앱 코드 칸의 줄 → 다음 칸. 다음 칸은 추적하지 않은 칸이
// 있으면 그 첫 칸, 없으면 마디 자신입니다. 프록시(이름에 $$)와 줄을 모르는 칸(줄 0 이하), 생성자·람다는 대조하지 않습니다.
//
// Arguments: t는 호출 트리, src는 앱 코드가 든 소스 뿌리입니다(nil이면 대조하지 않음).
// Returns: 어긋난 자리마다 "클래스:줄 → 메서드"입니다. 소스 파일이 없는 클래스도 어긋남으로 셉니다.
func checkAppSites(t *callTree, src *javaSources) []string {
	out := []string{}
	if src == nil {
		return out
	}
	check := func(class string, line int, callee string) {
		if !strings.HasPrefix(class, "lab.") || strings.Contains(class, "$$") || line < 1 ||
			strings.HasPrefix(callee, "<") || strings.HasPrefix(callee, "lambda$") {
			return
		}
		if !strings.Contains(src.line(class, line), callee+"(") {
			out = append(out, fmt.Sprintf("%s:%d → %s", class, line, callee))
		}
	}
	for _, n := range t.Nodes {
		// 추적하지 않은 칸을 바깥부터 늘어놓고 마디 자신을 끝에 둡니다. 칸마다 (클래스, 메서드, 그 칸 안에서 다음 칸을 부른 줄)입니다.
		type frame struct {
			class, method string
			line          int
		}
		var fs []frame
		for _, h := range n.Hidden {
			i := strings.LastIndexByte(h, ':')
			class, method, _ := strings.Cut(h[:i], "::")
			fs = append(fs, frame{class, method, atoiOr(h[i+1:], 0)})
		}
		class, method, _ := strings.Cut(n.Fn, "::")
		fs = append(fs, frame{class, method, 0})
		if n.Parent >= 0 {
			pc, _, _ := strings.Cut(t.Nodes[n.Parent].Fn, "::")
			check(pc, n.Site, fs[0].method)
		}
		for i := 0; i+1 < len(fs); i++ {
			check(fs[i].class, fs[i].line, fs[i+1].method)
		}
	}
	return out
}

// javaExcerptBefore 와 javaExcerptAfter 는 Java 소스 발췌에서 가리킨 줄 앞뒤로 보일 줄 수입니다. 갈린 자리의 if와 그 두 갈래가 함께
// 들어오도록 서버 소스(excerptBefore·excerptAfter)보다 좁게 잡았습니다.
const javaExcerptBefore, javaExcerptAfter = 6, 8

// excerptsFor 는 "클래스::메서드:줄" 목록마다 소스 발췌를 만듭니다. 키는 "클래스:줄"이고, 찾지 못한 것은 빠집니다.
func (js *javaSources) excerptsFor(refs []string) map[string]excerpt {
	out := map[string]excerpt{}
	for _, ref := range refs {
		i := strings.LastIndexByte(ref, ':')
		if i < 0 {
			continue
		}
		n := atoiOr(ref[i+1:], 0)
		class, _, _ := strings.Cut(ref[:i], "::")
		key := class + ":" + ref[i+1:]
		if _, done := out[key]; done || n < 1 {
			continue
		}
		text := js.lines(class)
		if n > len(text) {
			continue
		}
		from := max(1, n-javaExcerptBefore)
		to := min(len(text), n+javaExcerptAfter)
		out[key] = excerpt{From: from, Lines: text[from-1 : to]}
	}
	return out
}

// atoiOr 는 10진 정수를 읽고, 읽지 못하면 def를 돌려줍니다.
func atoiOr(s string, def int) int {
	n := 0
	if s == "" {
		return def
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// autoCommitArg 는 소스 한 줄에서 setAutoCommit의 리터럴 인자를 읽습니다(DataSourceTransactionManager 298행 "con.setAutoCommit(false)").
var autoCommitArg = regexp.MustCompile(`setAutoCommit\((true|false)\)`)

// annotateTree 는 트리에 효과, 끝난 꼴, 자기 시간, 렌즈 이유를 붙입니다.
//
// 효과는 callEffect의 종류마다 이렇게 정합니다. 요청은 가장 바깥 QueryExecutorImpl.execute가 열려 있는 동안 연결 위로 나간 묶음이고,
// 빌리기는 ConcurrentBag.borrow를 감싸는 풀 빌리기(springBorrow)이며, 잠듦과 예외는 그 사건 스택의 가장 안쪽 추적 칸과 이름·서술자가
// 같고 그 시각에 열려 있던 가장 깊은 마디에 붙입니다(같은 메서드가 그 시각에 겹쳐 열려 있으면 안쪽으로 봄). 자동 커밋의 값은 부른 줄의
// 소스 글에서 읽습니다. 연결의 묶임에 따른 효과(suspend, resume, return)는 bindingReplay가 붙입니다.
//
// 자기 시간은 자식이 모두 끝 시각을 가질 때만 셉니다. 길이에서 자식들의 길이를 빼고, 실험 도구의 멈춤(pauses) 가운데 자식 밖에 놓인
// 부분을 뺍니다.
//
// Arguments: t는 bindingReplay까지 마친 트리, j는 앱 쪽 기록, conns는 연결(요청), borrows는 풀 빌리기, pauses는 멈춤 구간,
// src는 Java 소스(nil이면 자동 커밋의 값을 비워 둠)입니다.
func annotateTree(t *callTree, j *springJava, conns []springConn, borrows []springBorrow, pauses [][2]float64, src *javaSources) {
	add := func(n *callNode, e callEffect) { n.Effects = append(n.Effects, e) }
	parentFn := func(n *callNode) string {
		if n.Parent < 0 {
			return ""
		}
		return t.Nodes[n.Parent].Fn
	}

	// 사건 하나를 그 스택의 가장 안쪽 추적 칸에 해당하는 마디에 붙입니다.
	deepestOpen := func(stack []jfrFrame, at float64) (*callNode, jfrFrame) {
		for _, f := range stack {
			if !j.traced(f.Method.key()) {
				continue
			}
			var best *callNode
			for _, n := range t.Nodes {
				if n.Fn == f.Method.key() && n.Desc == f.Method.Descriptor && n.T0 <= at && (n.T1 == nil || at <= *n.T1) &&
					(best == nil || n.Depth > best.Depth) {
					best = n
				}
			}
			return best, f
		}
		return nil, jfrFrame{}
	}

	for _, n := range t.Nodes {
		if n.Effects == nil {
			n.Effects = []callEffect{}
		}
		switch {
		case n.Fn == "org.postgresql.core.v3.QueryExecutorImpl::execute" && parentFn(n) != n.Fn:
			for _, c := range conns {
				var seq []int
				for _, r := range c.Requests {
					if r.Seq > 0 && r.Sent >= n.T0 && r.Sent <= n.last() {
						seq = append(seq, r.Seq)
					}
				}
				if len(seq) > 0 {
					add(n, callEffect{Kind: "wire", Conn: c.Name, Seq: seq})
				}
			}
		case n.Fn == "com.zaxxer.hikari.util.ConcurrentBag::borrow":
			for _, b := range borrows {
				if b.Thread == sceneThread && b.T0 <= n.T0 && n.last() <= b.T1 {
					add(n, callEffect{Kind: "borrow", Conn: b.Conn, Result: b.Result, IdleMs: b.IdleMs, Validated: b.Validated})
				}
			}
		case n.Fn == "org.postgresql.jdbc.PgConnection::setAutoCommit":
			e := callEffect{Kind: "autocommit"}
			if n.Parent >= 0 && src != nil {
				if m := autoCommitArg.FindStringSubmatch(src.line(parentFn(n), n.Site)); m != nil {
					e.Value = m[1]
				}
			}
			for _, c := range conns {
				for _, r := range c.Requests {
					if r.Seq > 0 && r.Sent >= n.T0 && r.Sent <= n.last() {
						e.Sent++
					}
				}
			}
			add(n, e)
		case n.Fn == "org.springframework.jdbc.datasource.DataSourceTransactionManager::doSetRollbackOnly":
			add(n, callEffect{Kind: "rollbackOnly"})
		case n.Fn == "com.zaxxer.hikari.pool.PoolBase::isConnectionDead":
			add(n, callEffect{Kind: "validate"})
		}
	}
	for _, p := range j.parks {
		if p.Thread != sceneThread {
			continue
		}
		if n, f := deepestOpen(p.stack, p.T0); n != nil {
			add(n, callEffect{Kind: "park", T0: p.T0, T1: p.T1, At: f.ref()})
		}
	}
	for _, th := range j.throws {
		if th.Thread != sceneThread {
			continue
		}
		if n, f := deepestOpen(th.stack, th.T); n != nil {
			add(n, callEffect{Kind: "throw", T0: th.T, Class: th.Class, Message: th.Message, At: f.ref()})
		}
	}

	// 끝난 꼴과 자기 시간. 마지막 자식이 끝난 뒤에 이 마디가 예외를 만들었으면 throw입니다.
	pauseIn := func(a, b float64) float64 {
		sum := 0.0
		for _, p := range pauses {
			sum += math.Max(0, math.Min(b, p[1])-math.Max(a, p[0]))
		}
		return sum
	}
	for _, n := range t.Nodes {
		lastKid := n.T0
		for _, k := range n.Kids {
			lastKid = math.Max(lastKid, t.Nodes[k].last())
		}
		switch {
		case n.T1 == nil:
			n.End = endPassed
		case n.span != nil && n.span.Twice:
			n.End = endTwice
		default:
			n.End = endReturn
			for _, e := range n.Effects {
				if e.Kind == "throw" && e.T0 >= lastKid {
					n.End = endThrow
				}
			}
		}
	}
	for _, n := range t.Nodes {
		if n.End == endPassed && len(n.Kids) > 0 {
			if last := t.Nodes[n.Kids[len(n.Kids)-1]]; last.End == endReturn {
				last.End = endRethrow
			}
		}
	}
	for _, n := range t.Nodes {
		if n.T1 == nil {
			continue
		}
		self := *n.T1 - n.T0 - pauseIn(n.T0, *n.T1)
		known := true
		for _, k := range n.Kids {
			kid := t.Nodes[k]
			if kid.T1 == nil {
				known = false
				break
			}
			self -= *kid.T1 - kid.T0 - pauseIn(kid.T0, *kid.T1)
		}
		if known {
			self = math.Max(0, self)
			n.Self = &self
		}
	}

	// 렌즈 이유. 갈림은 spring-compare가 더합니다.
	for _, n := range t.Nodes {
		n.Why = []string{}
		if len(n.Effects) > 0 {
			n.Why = append(n.Why, whyEffect)
		}
		if n.Self != nil && *n.Self >= selfLensUs {
			n.Why = append(n.Why, whyTime)
		}
	}
}

// isExceptionEnd 는 끝난 꼴이 예외 쪽인지 돌려줍니다.
func isExceptionEnd(end string) bool {
	return end != endReturn
}
