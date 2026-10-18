package main

// 탐색용 코드입니다(2026-10-13, 커널 추적 G3~G5). OrbStack 머신의 ftrace function_graph 출력을 읽어,
// strace로 본 시스템 콜 하나하나에 그 안에서 커널이 부른 함수 나무를 붙입니다.
// 추적의 pid와 시각은 strace와 같다고 가정하지 않습니다(컨테이너는 pid 네임스페이스가 따로 있고, 시계도 다릅니다).
// 대신 시스템 콜 이벤트의 인자(fd, count, pos)를 strace의 같은 호출과 순서대로 맞춥니다.

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// kcall 은 커널 함수 호출 하나입니다. Us는 그 함수 안에서 보낸 시간(자식과 잠든 시간 포함, function_graph의 sleep-time 기본값),
// Layer는 정의가 있는 소스 디렉터리로 정한 층, At은 원본 커널 소스에서 찾은 정의 줄입니다.
// Hidden은 화면에 남기지 않은 자손 호출의 수입니다. Guess는 정의를 찾지 못해(매크로로 만든 함수 등) 부모의 층을 물려받았다는 표시입니다.
// Calls는 부모 본문에서 이 함수까지 가는 호출 줄의 후보들입니다. 길 하나가 줄 하나면 부모가 바로 부르는 줄이고, 더 길면
// 추적에 나오지 않은(인라인된) 중간 함수를 거친 길입니다. 추적은 줄 번호를 남기지 않으므로 모두 이름으로 찾은 후보입니다.
type kcall struct {
	Fn    string      `json:"fn"`
	Us    float64     `json:"us"`
	Layer string      `json:"layer,omitempty"`
	Guess bool        `json:"guess,omitempty"`
	At    string      `json:"at,omitempty"`
	Calls [][]srcLine `json:"calls,omitempty"`
	// CallsOf는 형제 호출의 차례로 Calls를 좁혔을 때 좁히기 전 후보의 수입니다(narrow 참고). 좁히지 않았으면 0입니다.
	CallsOf int      `json:"callsOf,omitempty"`
	Hidden  int      `json:"hidden,omitempty"`
	Kids    []*kcall `json:"kids,omitempty"`
}

// srcLine 은 함수 Fn의 본문에 있는 소스 한 줄입니다. At은 파일:줄입니다.
type srcLine struct {
	Fn   string `json:"fn"`
	At   string `json:"at"`
	Text string `json:"text"`
}

// ksyscall 은 추적에서 본 시스템 콜 하나입니다. Name은 strace 이름(pwrite64, fdatasync), Pid는 커널이 본 pid입니다.
type ksyscall struct {
	Name  string
	Pid   int
	Fd    int64
	Count int64
	Pos   int64
	Root  *kcall
}

// kernelSyscalls 는 function_graph의 바깥 함수 이름에서 strace 이름으로 가는 표입니다. set_graph_function에 넣은 것과 같습니다.
var kernelSyscalls = map[string]string{
	"__arm64_sys_pwrite64":  "pwrite64",
	"__arm64_sys_pwritev":   "pwritev",
	"__arm64_sys_fdatasync": "fdatasync",
}

// kernelLayers 는 함수 정의가 있는 소스 경로에서 층 이름으로 가는 표입니다. 먼저 맞는 접두사를 씁니다.
// 함수 이름이 아니라 정의된 파일로 가르므로, btrfs가 정의한 prepare_one_folio는 이름과 달리 btrfs 층입니다.
var kernelLayers = []struct{ Prefix, Layer string }{
	{"arch/", "시스템 콜 입구"},
	{"fs/overlayfs/", "overlayfs"},
	{"fs/backing-file.c", "overlayfs"},
	{"fs/btrfs/", "btrfs"},
	{"mm/", "페이지 캐시·메모리"},
	{"block/", "블록 계층"},
	{"drivers/block/", "virtio 디스크"},
	{"drivers/virtio/", "virtio 디스크"},
	{"kernel/sched/", "잠들고 깨기"},
	{"kernel/locking/", "락"},
	{"security/", "권한 검사"},
	{"fs/", "VFS"},
}

// kernelLine 은 "시각 | CPU) 프로세스-pid | 시간 | 함수" 꼴의 function_graph 줄입니다(funcgraph-abstime, funcgraph-proc).
var kernelLine = regexp.MustCompile(`^\s*([\d.]+)\s*\|\s*\d+\)\s*(\S+)-(\d+)\s*\|([^|]*)\|(.*)$`)

// kernelArg 는 시스템 콜 이벤트 안의 "이름: 값" 인자 하나입니다. 값은 16진수(0x…)나 10진수입니다.
var kernelArg = regexp.MustCompile(`(\w+): (0x[0-9a-f]+|\d+)`)

// readKernelTrace 는 function_graph 출력에서 kernelSyscalls의 시스템 콜을 차례대로 읽습니다.
//
// Arguments: path는 /sys/kernel/tracing/trace를 그대로 받은 파일입니다.
// Returns: 시스템 콜마다 인자와 함수 나무. 시스템 콜 입구 이벤트(sys_enter_*)가 바깥 함수보다 먼저 찍히므로, 그 인자를 다음 바깥 함수에 붙입니다.
// Errors: 파일을 읽지 못하면 오류입니다. 형식에 맞지 않는 줄(문맥 전환 표시 등)은 건너뜁니다.
func readKernelTrace(path string) ([]ksyscall, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []ksyscall
	var stack []*kcall
	var args map[string]int64
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		m := kernelLine.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[3])
		body := strings.TrimSpace(m[5])
		us := parseKernelDur(m[4])
		switch {
		case strings.HasPrefix(body, "/*"):
			// 시스템 콜 입구 이벤트만 인자를 남깁니다. 나가는 이벤트("-> 0x2000")와 끝 표시("} /* fn */")는 쓰지 않습니다.
			if strings.Contains(body, "(") && !strings.Contains(body, "->") {
				args = map[string]int64{}
				for _, a := range kernelArg.FindAllStringSubmatch(body, -1) {
					args[a[1]] = parseNum(a[2])
				}
			}
		case strings.HasSuffix(body, "{"):
			n := &kcall{Fn: strings.TrimSuffix(strings.TrimSpace(strings.TrimSuffix(body, "{")), "()")}
			if len(stack) == 0 {
				name, ok := kernelSyscalls[n.Fn]
				if !ok {
					continue
				}
				out = append(out, ksyscall{Name: name, Pid: pid, Fd: args["fd"], Count: args["count"], Pos: args["pos"], Root: n})
				args = nil
			} else {
				stack[len(stack)-1].Kids = append(stack[len(stack)-1].Kids, n)
			}
			stack = append(stack, n)
		case strings.HasPrefix(body, "}"):
			if len(stack) > 0 {
				stack[len(stack)-1].Us = us
				stack = stack[:len(stack)-1]
			}
		case strings.HasSuffix(body, ";") && len(stack) > 0:
			n := &kcall{Fn: strings.TrimSuffix(strings.TrimSuffix(body, ";"), "()"), Us: us}
			stack[len(stack)-1].Kids = append(stack[len(stack)-1].Kids, n)
		}
	}
	return out, sc.Err()
}

// parseNum 은 이벤트 인자 값을 읽습니다. ftrace는 대부분 16진수(0x…)로, 0은 그냥 0으로 찍습니다.
func parseNum(s string) int64 {
	if h, ok := strings.CutPrefix(s, "0x"); ok {
		v, _ := strconv.ParseInt(h, 16, 64)
		return v
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// parseKernelDur 는 function_graph의 시간 칸("  0.042 us", "! 51845.75 us")에서 마이크로초를 읽습니다. 비어 있으면 0입니다.
func parseKernelDur(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimLeft(s, "+!#*@$ ")
	v, _ := strconv.ParseFloat(strings.TrimSuffix(s, " us"), 64)
	return v
}

// matchKernel 은 strace의 시스템 콜과 추적의 시스템 콜을 순서대로 짝짓습니다.
//
// Arguments: sys는 strace 사건 전체, pid는 컨테이너 안에서 본 대상 백엔드, calls는 readKernelTrace의 결과입니다.
// Returns: sys의 번호에서 짝지은 추적 시스템 콜로 가는 표와, 짝이 된 커널 pid(하나도 없으면 0).
// 같은 이름·fd이고, pwrite64는 쓴 바이트(strace의 반환값)와 위치(pos)까지 같아야 짝입니다.
// 짝이 맞지 않는 추적 시스템 콜(같은 커널의 다른 프로세스)은 건너뜁니다. 첫 짝의 커널 pid만 뒤에도 씁니다.
func matchKernel(sys []sysEvent, pid int, calls []ksyscall) (map[int]*ksyscall, int) {
	out := map[int]*ksyscall{}
	kpid, j := 0, 0
	for i, e := range sys {
		if e.Pid != pid || e.Ret < 0 {
			continue
		}
		if _, traced := map[string]bool{"pwrite64": true, "pwritev": true, "fdatasync": true}[e.Name]; !traced {
			continue
		}
		for k := j; k < len(calls); k++ {
			c := &calls[k]
			same := c.Name == e.Name && c.Fd == e.Fd && (kpid == 0 || c.Pid == kpid)
			if same && e.Name == "pwrite64" {
				same = c.Count == e.Ret && c.Pos == e.Offset
			}
			if same {
				out[i] = c
				kpid, j = c.Pid, k+1
				break
			}
		}
	}
	return out, kpid
}

// kernelIndex 는 원본 커널 소스에서 함수 이름으로 정의 위치("fs/btrfs/file.c:1234")를 찾는 표입니다.
type kernelIndex map[string][]string

// kernelDef 는 C 함수 정의의 첫 줄입니다. 줄 머리에서 시작하고(들여쓰기나 주석이 아님) 이름 뒤에 여는 괄호가 옵니다.
// 커널은 반환 형을 윗줄에 두고 이름을 줄 머리에 쓰기도 하므로("static noinline void __sched" 다음 줄의 "__mutex_lock_slowpath("),
// 이름 앞의 형은 없어도 됩니다. SYSCALL_DEFINEn(이름, …)은 __arm64_sys_이름의 정의입니다.
var kernelDef = regexp.MustCompile(`^(?:SYSCALL_DEFINE\d\((\w+)|(?:[A-Za-z_][\w \*]*?[ \*])?(\w+)\()`)

// buildKernelIndex 는 dir 아래 .c 파일에서 함수 정의 위치를 모읍니다.
//
// 정의로 보는 줄: kernelDef에 맞고, 줄이 세미콜론으로 끝나지 않으며, 다섯 줄 안에 줄 머리의 "{"가 세미콜론보다 먼저 나옵니다.
// Errors: dir를 걸을 수 없으면 오류입니다. 읽지 못한 파일은 건너뜁니다.
func buildKernelIndex(dir string) (kernelIndex, error) {
	idx := kernelIndex{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".c") {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		lines := strings.Split(string(b), "\n")
		for i, ln := range lines {
			m := kernelDef.FindStringSubmatch(ln)
			if m == nil || strings.HasSuffix(strings.TrimSpace(ln), ";") {
				continue
			}
			name := m[2]
			if m[1] != "" {
				name = "__arm64_sys_" + m[1]
			}
			if name == "" || !opensBody(lines[i:min(len(lines), i+6)]) {
				continue
			}
			idx[name] = append(idx[name], fmt.Sprintf("%s:%d", rel, i+1))
		}
		return nil
	})
	return idx, err
}

// opensBody 는 정의 후보 줄부터 몇 줄 안에 함수 본문의 "{"가 세미콜론보다 먼저 오는지 봅니다.
func opensBody(lines []string) bool {
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "{") {
			return true
		}
		if strings.HasSuffix(t, ";") {
			return false
		}
	}
	return false
}

// locate 는 함수 이름의 정의 위치를 고릅니다. 여럿이면(같은 이름의 static 함수) 부모 함수와 같은 디렉터리의 것을, 없으면 첫 것을 씁니다.
func (idx kernelIndex) locate(fn, parentAt string) string {
	locs := idx[fn]
	if len(locs) == 0 {
		return ""
	}
	if parentAt != "" {
		dir := filepath.Dir(parentAt[:strings.LastIndexByte(parentAt, ':')])
		for _, l := range locs {
			if filepath.Dir(l[:strings.LastIndexByte(l, ':')]) == dir {
				return l
			}
		}
	}
	return locs[0]
}

// layerOf 는 정의 위치에서 층 이름을 정합니다. 위치를 모르면 "기타"입니다.
func layerOf(at string) string {
	for _, l := range kernelLayers {
		if strings.HasPrefix(at, l.Prefix) {
			return l.Layer
		}
	}
	return "기타"
}

// kernelKeepUs 는 화면에 남길 커널 호출의 최소 시간입니다. 이보다 짧은 호출은, 그 아래에 블록 계층·virtio·잠듦·더럽힘 표시가 없으면 접습니다.
const kernelKeepUs = 2.0

// landmarks 는 시간이 짧아도 남길 호출입니다. 블록 I/O를 내보내는 자리, 장치에 요청을 넣는 자리, 잠들어 기다리는 자리,
// 페이지를 더럽다고 표시하는 자리입니다. 층 전체를 남기면(1차 시도) 깨우기와 통계 갱신까지 남아 fdatasync 하나가 392칸이 되었습니다.
var landmarks = map[string]bool{
	"submit_bio": true, "blk_mq_submit_bio": true, "virtio_queue_rq": true,
	"io_schedule": true, "schedule": true, "wait_for_completion_io": true, "folio_wait_bit": true, "mutex_lock": true,
	"filemap_dirty_folio": true, "__folio_mark_dirty": true,
}

// landmark 는 n이 landmarks에 든 호출인지입니다.
func landmark(n *kcall) bool { return landmarks[n.Fn] }

// annotate 는 나무의 모든 호출에 정의 위치와 층을 붙입니다. 정의를 찾지 못한 호출은 부모의 층을 물려받고 Guess를 켭니다.
// idx가 nil이면 모든 층이 "기타"입니다.
func annotate(n *kcall, idx kernelIndex, parentAt, parentLayer string) {
	if idx != nil {
		n.At = idx.locate(n.Fn, parentAt)
	}
	n.Layer = layerOf(n.At)
	if n.At == "" && parentLayer != "" {
		n.Layer, n.Guess = parentLayer, true
	}
	for _, k := range n.Kids {
		annotate(k, idx, n.At, n.Layer)
	}
}

// prune 은 화면에 남길 호출만 남기고, 버린 자손의 수를 Hidden에 적습니다. 남긴 호출이 아래에 있으면 그 조상도 남습니다.
// Returns: n 자신을 남겨야 하는지.
func prune(n *kcall, depth int) bool {
	var kept []*kcall
	for _, k := range n.Kids {
		if prune(k, depth+1) {
			kept = append(kept, k)
		} else {
			n.Hidden += 1 + countAll(k)
		}
	}
	n.Kids = kept
	return depth <= 1 || n.Us >= kernelKeepUs || landmark(n) || len(kept) > 0
}

// countAll 은 n 아래 자손 호출의 수입니다.
func countAll(n *kcall) int {
	c := 0
	for _, k := range n.Kids {
		c += 1 + countAll(k)
	}
	return c + n.Hidden
}

// ksource 는 원본 커널 소스를 한 번씩만 읽어 두는 곳입니다. 파일의 줄과, 함수 본문마다 그 안에서 부르는 함수 목록을 모아 둡니다.
// dir이 비어 있으면 아무 줄도 찾지 않습니다.
type ksource struct {
	dir   string
	idx   kernelIndex
	files map[string][]string
	calls map[string][]bodyCall
}

// bodyCall 은 함수 본문의 한 줄에서 부르는 함수 하나입니다.
type bodyCall struct {
	Callee string
	Line   srcLine
}

// calleeName 은 한 줄에서 "이름(" 꼴로 부르는 함수 이름을 찾습니다. if·sizeof 같은 낱말도 걸리지만 정의 색인에 없어 거쳐 갈 함수가 되지 않습니다.
var calleeName = regexp.MustCompile(`\b([A-Za-z_]\w*)\(`)

// maxVia 는 부모와 자식 사이에 끼워 볼 중간 함수의 최대 개수입니다. btrfs의 버퍼 쓰기가 btrfs_file_write_iter에서
// prepare_one_folio까지 세 함수(btrfs_do_write_iter, btrfs_buffered_write, copy_one_range)를 인라인으로 거칩니다.
const maxVia = 3

// maxPaths 는 중간 함수를 거치는 길을 자식 하나에 몇 개까지 남길지입니다. 바로 부르는 줄은 모두 남깁니다(narrow가 좁힙니다).
const maxPaths = 3

// body 는 at(파일:줄)에서 정의한 함수 fn의 본문에서 부르는 함수들입니다. 본문은 정의 다음 줄부터 줄 머리의 "}"까지로 보고,
// 주석 줄("*", "/*", "//"로 시작)은 건너뜁니다. 주석이 "prepare_one_folio()에서"처럼 함수 이름을 부르는 꼴로 쓰기 때문입니다.
func (s *ksource) body(fn, at string) []bodyCall {
	if c, ok := s.calls[at]; ok {
		return c
	}
	i := strings.LastIndexByte(at, ':')
	name := at[:i]
	start, _ := strconv.Atoi(at[i+1:])
	lines, ok := s.files[name]
	if !ok {
		if b, err := os.ReadFile(filepath.Join(s.dir, name)); err == nil {
			lines = strings.Split(string(b), "\n")
		}
		s.files[name] = lines
	}
	var out []bodyCall
	for ln := start; ln < len(lines) && !strings.HasPrefix(lines[ln], "}"); ln++ {
		if t := strings.TrimSpace(lines[ln]); strings.HasPrefix(t, "*") || strings.HasPrefix(t, "/*") || strings.HasPrefix(t, "//") {
			continue
		}
		for _, m := range calleeName.FindAllStringSubmatch(lines[ln], -1) {
			out = append(out, bodyCall{Callee: m[1], Line: srcLine{Fn: fn, At: fmt.Sprintf("%s:%d", name, ln+1), Text: strings.TrimSpace(lines[ln])}})
		}
	}
	s.calls[at] = out
	return out
}

// paths 는 at에서 정의한 함수 fn에서 target까지 가는 호출 줄의 길들입니다. 바로 부르는 줄이 있으면 길이 1인 길만 돌려줍니다.
// 없으면 같은 파일에 정의된 함수를 via개까지 거쳐 가는 길을 찾습니다. 추적에 그 중간 함수가 나오지 않았으니 컴파일러가 그 자리에
// 인라인했다고 봅니다. LTO가 없으면 다른 .c 파일의 함수는 인라인되지 않으므로 중간 함수를 같은 파일로 좁힙니다. 이 커널에 LTO가
// 없다는 근거는 약합니다(축약된 /proc/config.gz에 LTO 항목이 없고 kallsyms에 .llvm. 꼬리가 없음). 작업 가정으로 두고,
// 같은 파일에서 길을 못 찾는 호출이 많아지면 다시 봅니다. skip은 추적에 따로 나온 형제 호출입니다. 그 함수를 거쳤다면 target이 그 아래에 나왔을 것이라 뺍니다.
func (s *ksource) paths(fn, at, target string, skip map[string]bool, via int) [][]srcLine {
	file := at[:strings.LastIndexByte(at, ':')]
	var direct, deeper [][]srcLine
	for _, c := range s.body(fn, at) {
		if c.Callee == target {
			direct = append(direct, []srcLine{c.Line})
			continue
		}
		if via == 0 || skip[c.Callee] || c.Callee == fn {
			continue
		}
		loc := s.idx.inFile(c.Callee, file)
		if loc == "" {
			continue
		}
		for _, p := range s.paths(c.Callee, loc, target, skip, via-1) {
			deeper = append(deeper, append([]srcLine{c.Line}, p...))
		}
	}
	if len(direct) > 0 {
		return direct
	}
	return deeper
}

// inFile 은 fn을 file 안에서 정의한 위치(파일:줄)입니다. 없으면 빈 문자열입니다.
func (idx kernelIndex) inFile(fn, file string) string {
	for _, l := range idx[fn] {
		if strings.HasPrefix(l, file+":") {
			return l
		}
	}
	return ""
}

// callSites 는 남은 호출마다 부모 함수 본문에서 그 함수까지 가는 호출 줄을 찾아 Calls에 넣습니다.
// 중간 함수를 적게 거치는 길부터 찾고(0개, 1개, … maxVia개), 중간 함수를 거친 길은 maxPaths개까지 남깁니다.
// 그다음 형제 호출의 차례로 후보를 좁힙니다(narrow).
// 함수 포인터로 부른 호출(f_op->fsync 등)과 받아 두지 않은 헤더의 함수를 거친 호출은 찾지 못해 Calls가 비어 있습니다.
func callSites(n *kcall, s *ksource) {
	if s.dir == "" || n.At == "" {
		return
	}
	skip := map[string]bool{}
	for _, k := range n.Kids {
		skip[k.Fn] = true
	}
	for _, k := range n.Kids {
		for via := 0; via <= maxVia; via++ {
			if p := s.paths(n.Fn, n.At, k.Fn, skip, via); len(p) > 0 {
				if via > 0 {
					p = p[:min(len(p), maxPaths)]
				}
				k.Calls = p
				break
			}
		}
	}
	narrow(n.Kids)
	for _, k := range n.Kids {
		callSites(k, s)
	}
}

// narrow 는 형제 호출이 불린 차례로 호출 줄 후보를 좁힙니다. 부모 본문을 위에서 아래로 실행했다면(반복이나 뒤로 가는 goto가 없다면)
// 먼저 불린 형제의 호출 줄은 나중에 불린 형제의 호출 줄보다 앞서거나 같습니다. 형제마다 그런 배정이 가능한 후보만 남깁니다.
// 앞에서부터 가능한 가장 이른 줄(lo)과 뒤에서부터 가능한 가장 늦은 줄(hi)을 구하면, 후보 L은 lo ≤ L ≤ hi일 때만 가능합니다.
// 그런 배정이 아예 없으면(반복문 안에서 여러 번 부른 경우 등) 전제가 깨진 것이므로 아무것도 좁히지 않습니다. 후보가 없는 형제는 제약을 주지 않습니다.
// 이 좁히기는 추론이며, 화면은 CallsOf로 좁혔다는 사실을 알립니다.
func narrow(kids []*kcall) {
	line := func(p []srcLine) int {
		at := p[0].At
		v, _ := strconv.Atoi(at[strings.LastIndexByte(at, ':')+1:])
		return v
	}
	lo, hi := make([]int, len(kids)), make([]int, len(kids))
	prev := 0
	for i, k := range kids {
		lo[i] = prev
		if len(k.Calls) == 0 {
			continue
		}
		best := -1
		for _, p := range k.Calls {
			if l := line(p); l >= prev && (best < 0 || l < best) {
				best = l
			}
		}
		if best < 0 {
			return
		}
		lo[i], prev = best, best
	}
	next := int(^uint(0) >> 1)
	for i := len(kids) - 1; i >= 0; i-- {
		hi[i] = next
		if len(kids[i].Calls) == 0 {
			continue
		}
		best := -1
		for _, p := range kids[i].Calls {
			if l := line(p); l <= next && l > best {
				best = l
			}
		}
		if best < 0 {
			return
		}
		hi[i], next = best, best
	}
	for i, k := range kids {
		var keep [][]srcLine
		for _, p := range k.Calls {
			if l := line(p); l >= lo[i] && l <= hi[i] {
				keep = append(keep, p)
			}
		}
		if len(keep) < len(k.Calls) {
			k.CallsOf, k.Calls = len(k.Calls), keep
		}
	}
}

// bodyMax 는 호출 줄 발췌에 담을 함수 본문의 최대 줄 수입니다. 더 길면 호출 줄을 가운데 둔 bodyMax줄만 담습니다.
const bodyMax = 160

// kernelExcerpts 는 나무에 남은 호출들의 소스 발췌입니다. 정의 줄은 앞뒤 몇 줄을(excerpts), 호출 줄은 그 줄이 든 함수의 본문 전체를 담습니다.
// 호출 줄 쪽을 길게 두는 것은, 인라인되어 추적에 나오지 않은 일을 본문에서 읽게 하기 위해서입니다. 예를 들어 pwrite64의 8 KB를
// 페이지 캐시로 복사하는 copy_folio_from_iter_atomic은 copy_one_range 본문(btrfs/file.c)에만 보입니다.
// 키는 파일:줄이고 값의 From은 Lines[0]의 줄 번호입니다.
func kernelExcerpts(roots []*kcall, s *ksource) map[string]excerpt {
	var defs []string
	calls := map[string]srcLine{}
	var walk func(n *kcall)
	walk = func(n *kcall) {
		if n.At != "" {
			defs = append(defs, n.At)
		}
		for _, p := range n.Calls {
			for _, l := range p {
				calls[l.At] = l
			}
		}
		for _, k := range n.Kids {
			walk(k)
		}
	}
	for _, r := range roots {
		walk(r)
	}
	out := excerpts(s.dir, defs)
	for at, l := range calls {
		if ex, ok := s.bodyExcerpt(l); ok {
			out[at] = ex
		}
	}
	return out
}

// bodyExcerpt 는 호출 줄 l이 든 함수 l.Fn의 본문(정의 줄부터 줄 머리의 "}"까지)입니다. 정의를 같은 파일에서 찾지 못하면 false입니다.
func (s *ksource) bodyExcerpt(l srcLine) (excerpt, bool) {
	i := strings.LastIndexByte(l.At, ':')
	file := l.At[:i]
	line, _ := strconv.Atoi(l.At[i+1:])
	def := s.idx.inFile(l.Fn, file)
	lines := s.files[file]
	if def == "" || len(lines) == 0 {
		return excerpt{}, false
	}
	from, _ := strconv.Atoi(def[strings.LastIndexByte(def, ':')+1:])
	// to는 본문을 닫는 "}" 줄의 번호입니다. 정의 줄 다음부터 찾습니다.
	to := from + 1
	for to <= len(lines) && !strings.HasPrefix(lines[to-1], "}") {
		to++
	}
	to = min(to, len(lines))
	if to-from+1 > bodyMax {
		from = max(from, line-bodyMax/2)
		to = min(to, from+bodyMax-1)
	}
	return excerpt{From: from, Lines: lines[from-1 : to]}, true
}
