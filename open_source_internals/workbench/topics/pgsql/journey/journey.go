// Package journey 는 요청의 여정 화면(concept-3·concept-4)에 박힌 해석 결과를 공통 형식으로 옮기고 되돌립니다.
//
// 옮기는 규칙은 설계 메모 5의 2입니다. 사건의 place는 그 일이 실행된 자리(코드 자리 → 프로세스)이고,
// 이동의 출발·도착 칸은 from·target입니다. concept-3의 두 질의는 concept-4의 같은 기록에 두 번째 해석기 출처로 들어갑니다.
// 원본의 모든 열쇠는 아래 allowed 목록으로 받고, 모르는 열쇠가 있으면 조용히 버리지 않고 멈춥니다.
package journey

import (
	"encoding/json"
	"fmt"
	"sort"

	"study/open_source_internals/workbench/internal/atlas"
	"study/open_source_internals/workbench/topics/pgsql/load"
)

// 추정 짝의 경로 무리입니다. 설계 메모 7의 추정 목록(순서나 인자로 맞춘 메시지·시스템 콜·커널 짝)에 들어,
// 출처 성격 검사에 이 이름들을 넘깁니다.
const (
	MessageCopy = "topic:message-copy"
	SyscallPair = "topic:syscall-pair"
	KernelPair  = "topic:kernel-pair"
)

// decoder 값은 한 판의 해석기에 매인 출처 id와 원본 경로입니다. concept-4와 concept-3은 같은 기록을 다른 판으로 풀었습니다.
type decoder struct {
	rec, journey, pair, src string
	// pre는 원본에서 한 질의의 해석 결과가 놓인 경로("$.data" 또는 "$")이고, queryKey·srcKey는 run attrs의 열쇠입니다.
	pre, queryKey, srcKey string
}

// c4 와 c3 는 두 판의 해석기입니다.
var (
	c4 = decoder{rec: "rec", journey: "journey", pair: "pair", src: "src", pre: "$.data", queryKey: "data.query", srcKey: "data.src"}
	c3 = decoder{rec: "rec-c3", journey: "journey-c3", pair: "pair-c3", src: "src-c3", pre: "$", queryKey: "query", srcKey: "src"}
)

// 판과 무관한 출처 id입니다.
const (
	srcStrace = "strace"
	srcFtrace = "ftrace"
	srcPS     = "ps"
	srcGuess  = "procs-guess"
	srcTopic  = "topic"
	srcKsrc   = "ksrc"
	srcNote   = "pgsql-topic"
)

// clockRecv 는 그 질의를 받은 순간(recv.Ns)부터 센 시계입니다(decode/journey.go 579·885행).
const clockRecv = "recv"

// fileLanes 는 파일이 놓이는 칸입니다. 파일 장소는 이 칸 아래에 둡니다.
var fileLanes = map[string]bool{"data_files": true, "wal_files": true}

// Result 값은 적재 결과입니다. Origins는 검사에만 쓰는 곁 기록이고 형식에는 들어가지 않습니다.
type Result struct {
	Run     *atlas.Run
	Origins atlas.Origins
}

// loader 값은 run 하나를 조립하는 동안의 상태입니다. 장소는 공용 장부(load.Loader)가 내용으로 찾아 두 판의 해석기가 같이 씁니다.
type loader struct {
	*load.Loader
}

// Import 함수는 concept-4의 질의 하나(q)와, 같은 기록을 앞선 해석기로 푼 concept-3의 해석 결과(c3data, 없으면 nil)를 run 하나로 옮깁니다.
//
// version은 소스 판(workbench.json의 sourceRoot)이고 정적 장소의 key에 들어갑니다. 원본에 모르는 열쇠가 있거나,
// 규칙(메모 5의 2)으로 되살릴 수 없는 꼴(proc가 pid와 맞지 않음, 짝이 맞지 않는 사본 메시지, 이어지는 메시지가 없는 cut 등)이면 오류를 돌려줍니다.
func Import(q map[string]any, c3data map[string]any, version string) (Result, error) {
	if err := load.Keys(q, "$", "key", "label", "typed", "x", "data", "procs", "procsGuess"); err != nil {
		return Result{}, err
	}
	key, ok := q["key"].(string)
	if !ok {
		return Result{}, fmt.Errorf("질의의 key가 글이 아님")
	}
	data, err := load.Object(q["data"], "$.data")
	if err != nil {
		return Result{}, err
	}
	l := &loader{load.NewLoader(atlas.NewBuilder("journey/"+key, "pgsql", "journey", key,
		atlas.Origin{Clock: clockRecv, Label: "그 질의를 받은 순간(recv.Ns)"}), version)}
	l.B.AddClock(atlas.Clock{ID: clockRecv, PrecisionNS: 1, Note: "기록하는 동안 잰 시각이라 실제보다 느립니다(concept-3.html 253행)"})

	// 주제가 붙인 이름표는 사람 출처의 run attrs입니다.
	l.Source(atlas.Source{ID: srcTopic, Tool: "pgsql 주제(화면의 질의 목록)", Nature: atlas.Human})
	for _, k := range []string{"key", "label", "typed", "x"} {
		l.B.SetRunAttrs(srcTopic, k, q[k])
	}

	// 서버 프로세스 역할표: procsGuess가 참이면 ps가 아니라 짐작한 표입니다(concept-4.html 223행 주석).
	guess, ok := q["procsGuess"].(bool)
	if !ok {
		return Result{}, fmt.Errorf("procsGuess가 참·거짓이 아님")
	}
	if err := l.procs(q["procs"], guess); err != nil {
		return Result{}, err
	}

	if err := l.query(c4, data); err != nil {
		return Result{}, fmt.Errorf("%s: %w", key, err)
	}
	if c3data != nil {
		if err := l.query(c3, c3data); err != nil {
			return Result{}, fmt.Errorf("%s(concept-3): %w", key, err)
		}
	}
	return Result{Run: l.B.Run(), Origins: l.Origins}, nil
}

// procs 메서드는 역할표를 역할 장소와 그 아래 프로세스 장소로 옮깁니다. 역할은 이름순으로 만들어 결과를 늘 같게 합니다.
func (l *loader) procs(v any, guess bool) error {
	table, err := load.Object(v, "$.procs")
	if err != nil {
		return err
	}
	src := atlas.Source{ID: srcPS, Tool: "ps(기록할 때 본 서버 프로세스 역할)", Nature: atlas.Observed}
	if guess {
		src = atlas.Source{ID: srcGuess, Tool: "fork 순서와 하는 일로 짐작한 역할표", Nature: atlas.Inferred}
	}
	l.Source(src)
	roles := make([]string, 0, len(table))
	for r := range table {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	for _, r := range roles {
		role := l.Place("role\x00"+r, atlas.PlaceSpec{Kind: "role", Name: r, Source: src.ID}, "$.procs."+r)
		pids, ok := table[r].([]any)
		if !ok {
			return fmt.Errorf("$.procs.%s가 목록이 아님", r)
		}
		for _, p := range pids {
			pid, ok := p.(json.Number)
			if !ok {
				return fmt.Errorf("$.procs.%s의 pid %v", r, p)
			}
			if _, dup := l.Lookup("process\x00" + string(pid)); dup {
				return fmt.Errorf("pid %s가 역할표에 두 번", pid)
			}
			l.Place("process\x00"+string(pid), atlas.PlaceSpec{Parent: &role, Kind: "process", Name: string(pid), Source: src.ID}, "$.procs."+r+"[]")
		}
	}
	return nil
}

// query 메서드는 한 판의 해석 결과(질의 하나)를 옮깁니다: 질의 구간, 이동과 메시지·시스템 콜·커널 마디, 단계, run attrs.
func (l *loader) query(d decoder, data map[string]any) error {
	allowed := []string{"endUs", "hops", "pid", "query", "src", "stages"}
	if d == c4 {
		allowed = append(allowed, "kernelPid", "ksrc")
	}
	if err := load.Keys(data, d.pre, allowed...); err != nil {
		return err
	}
	l.Source(atlas.Source{ID: d.rec, Tool: "uprobe·strace 기록(" + d.journey + " 해석기가 읽음)", Clock: clockRecv, Nature: atlas.Observed})
	l.Source(atlas.Source{ID: d.journey, Tool: decoderTool(d), Clock: clockRecv, Nature: atlas.Derived})
	l.Source(atlas.Source{ID: d.src, Tool: "PostgreSQL 소스 발췌(이미지 pg-internals:18.6)", Nature: atlas.Observed})
	l.B.SetRunAttrs(d.rec, d.queryKey, data["query"])
	l.B.SetRunAttrs(d.src, d.srcKey, data["src"])
	l.B.AddUnrecorded(atlas.Unrecorded{Source: d.rec, Field: "message.start_ns", Note: "메시지는 그 이동(수신·버퍼 쓰기·송신)에 딸려 기록되고 따로 시각이 없습니다"})
	l.B.AddUnrecorded(atlas.Unrecorded{Source: d.rec, Field: "coverage", Note: "어떤 함수에 탐침을 걸었는지 화면 데이터에 남지 않았습니다"})

	// 질의 하나는 백엔드 프로세스의 구간이고, 이 사건의 장소가 그 질의의 백엔드입니다(이동의 proc를 다시 세는 근거).
	pid, ok := data["pid"].(json.Number)
	if !ok {
		return fmt.Errorf("%s.pid가 수가 아님", d.pre)
	}
	backend := l.Process(string(pid), d.rec, d.pre+".pid")
	end, err := load.Micros(data["endUs"], d.pre+".endUs")
	if err != nil {
		return err
	}
	span := l.B.AddEvent(atlas.Event{Place: backend, Kind: atlas.Span, Name: "query", StartNS: atlas.Ptr[int64](0), EndNS: &end, Source: d.journey})
	l.Origins[atlas.RecordRef{Table: "event", ID: span}] = d.pre + ".endUs"

	// 커널 마디의 프로세스는 커널 추적기가 본 pid입니다(concept-4의 kernel_* 셋에만 있음).
	kproc := -1
	if kp, ok := data["kernelPid"]; ok {
		n, ok := kp.(json.Number)
		if !ok {
			return fmt.Errorf("kernelPid가 수가 아님")
		}
		l.Source(atlas.Source{ID: srcFtrace, Tool: "ftrace function_graph(커널 마디마다 보낸 시간)", Nature: atlas.Observed})
		kproc = l.Place("kprocess\x00"+string(n), atlas.PlaceSpec{Kind: "process", Name: string(n), Source: srcFtrace}, d.pre+".kernelPid")
		l.Source(atlas.Source{ID: srcKsrc, Tool: "Linux v7.2.9 소스 발췌(실행 커널은 7.2.9-orbstack, concept-4.html 546행)", Nature: atlas.Observed})
		l.B.SetRunAttrs(srcKsrc, "data.ksrc", data["ksrc"])
	} else if _, ok := data["ksrc"]; ok {
		return fmt.Errorf("kernelPid 없이 ksrc가 있음")
	}

	hops, ok := data["hops"].([]any)
	if !ok {
		return fmt.Errorf("%s.hops가 목록이 아님", d.pre)
	}
	if err := l.hops(d, hops, string(pid), kproc); err != nil {
		return err
	}
	stages, ok := data["stages"].([]any)
	if !ok {
		return fmt.Errorf("%s.stages가 목록이 아님", d.pre)
	}
	return l.stages(d, stages, backend)
}

// decoderTool 함수는 해석기 출처의 도구 설명을 돌려줍니다.
func decoderTool(d decoder) string {
	if d == c3 {
		return "해석기 decode/journey.go의 앞선 판(concept-3 화면)"
	}
	return "해석기 decode/journey.go(concept-4 화면)"
}

// hopKeys 는 이동 하나가 가질 수 있는 열쇠입니다(재고 조사 3.4.3·3.5.3절의 키 트리).
var hopKeys = []string{"n", "pid", "stage", "proc", "from", "to", "kind", "count", "bytes", "msgs", "t0", "t1", "stack", "site", "catalog", "files", "sysUs", "kernel"}

// hops 메서드는 이동을 차례로 옮깁니다. 메시지의 사본 짝과 잘린 메시지의 이어짐은 해석기의 규칙(journey.go 670~693, 338~368행)을 그대로 따라 셉니다.
func (l *loader) hops(d decoder, hops []any, queryPid string, kproc int) error {
	var (
		pending    []int     // 아직 송신과 짝짓지 못한 버퍼 쓰기 메시지
		pendingMsg []msgFact // 그 메시지의 내용
		cut        = -1      // 다음 송신에서 이어져야 할 잘린 송신 메시지
		nullStack  int
	)
	for i, raw := range hops {
		where := fmt.Sprintf("%s.hops[%d]", d.pre, i)
		h, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(h, where, hopKeys...); err != nil {
			return err
		}

		// 실행한 프로세스와 코드 자리가 이동의 장소입니다. proc는 저장하지 않고 다시 세므로, 그 규칙과 맞는지 여기서 확인합니다.
		pid, ok := h["pid"].(json.Number)
		if !ok {
			return fmt.Errorf("%s.pid가 수가 아님", where)
		}
		proc := l.Process(string(pid), d.rec, d.pre+".hops[].pid")
		want := "other"
		if string(pid) == queryPid {
			want = "backend"
		}
		if h["proc"] != want {
			return fmt.Errorf("%s.proc %v, 다시 센 값 %s", where, h["proc"], want)
		}
		site, ok := h["site"].(string)
		if !ok {
			return fmt.Errorf("%s.site가 글이 아님", where)
		}
		place := proc
		if site != "" {
			if place, err = l.Site(proc, site, d.rec, d.pre+".hops[].site"); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}

		// 출발·도착 칸과, 파일이 있으면 그 칸을 좁히는 파일 장소입니다.
		fromName, ok1 := h["from"].(string)
		toName, ok2 := h["to"].(string)
		if !ok1 || !ok2 {
			return fmt.Errorf("%s의 from·to가 글이 아님", where)
		}
		from := l.lane(fromName, d.journey, d.pre+".hops[].from")
		to := l.lane(toName, d.journey, d.pre+".hops[].to")
		if fv, ok := h["files"]; ok {
			files, ok := fv.([]any)
			if !ok || len(files) != 1 {
				return fmt.Errorf("%s.files가 하나짜리 목록이 아님", where)
			}
			path, ok := files[0].(string)
			if !ok {
				return fmt.Errorf("%s.files[0]이 글이 아님", where)
			}
			switch {
			case fileLanes[fromName]:
				from = l.file(from, fromName, path, d.rec, d.pre+".hops[].files[]")
			case fileLanes[toName]:
				to = l.file(to, toName, path, d.rec, d.pre+".hops[].files[]")
			default:
				return fmt.Errorf("%s: 파일 %s가 놓일 파일 칸이 없음(%s → %s)", where, path, fromName, toName)
			}
		}

		// 시각과 횟수: 합친 이동은 group이고, 하나면 instant라 끝이 시작과 같아야 합니다.
		t0, err := load.Micros(h["t0"], where+".t0")
		if err != nil {
			return err
		}
		t1, err := load.Micros(h["t1"], where+".t1")
		if err != nil {
			return err
		}
		count, err := load.Integer(h["count"], where+".count")
		if err != nil {
			return err
		}
		kind := atlas.Instant
		switch {
		case count >= 2:
			kind = atlas.Group
		case count == 1 && t0 == t1:
		default:
			return fmt.Errorf("%s: count %d, t0 %d, t1 %d", where, count, t0, t1)
		}
		name, ok := h["kind"].(string)
		if !ok {
			return fmt.Errorf("%s.kind가 글이 아님", where)
		}

		// 스택은 원본이 바깥 함수부터 적어 뒤집고, 0번에 이동의 장소(코드 자리)를 둡니다. null은 스택을 남기지 않은 것입니다.
		var stack []int
		switch sv := h["stack"].(type) {
		case nil:
			nullStack++
		case []any:
			if len(sv) == 0 {
				return fmt.Errorf("%s.stack이 빈 목록(null과 가를 수 없음)", where)
			}
			stack = []int{place}
			for k := len(sv) - 1; k >= 0; k-- {
				f, err := load.Object(sv[k], where+".stack[]")
				if err != nil {
					return err
				}
				if err := load.Keys(f, where+".stack[]", "fn", "at"); err != nil {
					return err
				}
				fn, ok1 := f["fn"].(string)
				at, ok2 := f["at"].(string)
				if !ok1 || !ok2 {
					return fmt.Errorf("%s.stack[]의 fn·at이 글이 아님", where)
				}
				fr, err := l.Frame(proc, fn, at, d.rec, d.pre+".hops[].stack[].fn", d.pre+".hops[].stack[].at")
				if err != nil {
					return fmt.Errorf("%s: %w", where, err)
				}
				stack = append(stack, fr)
			}
		default:
			return fmt.Errorf("%s.stack이 목록이나 null이 아님", where)
		}
		ev := atlas.Event{Place: place, From: &from, Target: &to, Kind: kind, Name: name, StartNS: atlas.Ptr(t0), EndNS: atlas.Ptr(t1), Count: count, Stack: stack, Source: d.rec}
		hop := l.B.AddEvent(ev)
		l.Origins[atlas.RecordRef{Table: "event", ID: hop}] = d.pre + ".hops[]"
		l.B.EventAttrs(hop, d.journey, "n", h["n"])
		l.B.EventAttrs(hop, d.journey, "stage", h["stage"])
		if v, ok := h["catalog"]; ok {
			l.B.EventAttrs(hop, d.journey, "catalog", v)
		}
		if v, ok := h["bytes"]; ok {
			l.B.EventAttrs(hop, d.rec, "bytes", v)
		}

		// 메시지: 수신·송신의 메시지는 그 이동의 바이트를 나눈 것이고, 버퍼 쓰기의 메시지는 뒤따른 송신의 사본입니다.
		bufferWrite := string(pid) == queryPid && fromName == "backend" && toName == "comm_buffer"
		send := string(pid) == queryPid && name == "응답 보냄"
		recv := fromName == "client" && toName == "comm_buffer"
		var fresh []int
		var freshMsg []msgFact
		if mv, ok := h["msgs"]; ok {
			msgs, ok := mv.([]any)
			if !ok || len(msgs) == 0 {
				return fmt.Errorf("%s.msgs가 비지 않은 목록이 아님", where)
			}
			if !bufferWrite && !send && !recv {
				return fmt.Errorf("%s: 메시지를 가진 이동의 종류 %q를 모름", where, name)
			}
			for k, mraw := range msgs {
				mwhere := fmt.Sprintf("%s.msgs[%d]", where, k)
				m, fact, err := l.message(d, mraw, mwhere, ev)
				if err != nil {
					return err
				}
				basis, group := d.rec, d.pre+".hops[].msgs"
				if bufferWrite {
					basis, group = d.pair, MessageCopy
					l.Source(atlas.Source{ID: d.pair, Tool: d.journey + " 해석기가 차례로 맞춘 짝", Nature: atlas.Inferred})
				}
				link := l.B.AddLink(hop, m, "contains", basis)
				l.Origins[atlas.RecordRef{Table: "link", ID: link}] = group

				switch {
				case bufferWrite:
					if fact.cont {
						return fmt.Errorf("%s: 버퍼 쓰기 메시지에 cont", mwhere)
					}
					pending, pendingMsg = append(pending, m), append(pendingMsg, fact)
				case send:
					// 앞 송신에서 이어진 나머지는 앞 송신의 잘린 메시지에서 오는 continues 연결입니다(첫 메시지에만 올 수 있음).
					if fact.cont {
						if k != 0 || cut < 0 {
							return fmt.Errorf("%s: 잘린 메시지 없이 cont", mwhere)
						}
						link := l.B.AddLink(cut, m, "continues", d.journey)
						l.Origins[atlas.RecordRef{Table: "link", ID: link}] = d.pre + ".hops[].msgs[].cont"
						cut = -1
					} else {
						fresh, freshMsg = append(fresh, m), append(freshMsg, fact)
					}
					if fact.cut {
						if k != len(msgs)-1 {
							return fmt.Errorf("%s: 마지막이 아닌 메시지에 cut", mwhere)
						}
						cut = m
					}
				default:
					if fact.cut || fact.cont {
						return fmt.Errorf("%s: 수신 메시지에 cut·cont", mwhere)
					}
				}
			}
		} else if send && cut >= 0 {
			return fmt.Errorf("%s: 잘린 메시지 뒤의 송신에 메시지가 없음", where)
		}

		// 송신마다 해석기는 기다리던 버퍼 쓰기 메시지를 이 송신의 새 메시지와 수가 같을 때만 차례로 붙이고, 기다림을 비웠습니다.
		if send {
			if len(pending) > 0 {
				if len(pending) != len(fresh) {
					return fmt.Errorf("%s: 버퍼 쓰기 메시지 %d개, 송신의 새 메시지 %d개", where, len(pending), len(fresh))
				}
				for j := range pending {
					if pendingMsg[j] != freshMsg[j] {
						return fmt.Errorf("%s: 사본 %d의 내용이 송신 메시지와 다름", where, j)
					}
					link := l.B.AddLink(fresh[j], pending[j], "copy", d.pair)
					l.Origins[atlas.RecordRef{Table: "link", ID: link}] = MessageCopy
				}
			}
			pending, pendingMsg = nil, nil
		}

		// 시스템 콜 길이와 커널 마디: 해석기가 순서와 인자로 맞춘 짝이라 연결의 근거는 추정입니다.
		if err := l.syscall(d, h, where, hop, place, kproc); err != nil {
			return err
		}
	}
	if len(pending) > 0 || cut >= 0 {
		return fmt.Errorf("%s: 짝을 찾지 못한 버퍼 쓰기 메시지 %d개 또는 이어지지 않은 잘린 메시지", d.pre, len(pending))
	}
	if nullStack > 0 {
		l.Source(atlas.Source{ID: srcNote, Tool: "pgsql 주제의 여정 적재기(topics/pgsql/journey)", Nature: atlas.Human})
		l.B.AddNote(atlas.Note{Scope: atlas.Scope{Kind: "run"}, Kind: "ambiguity", Source: srcNote,
			Text: fmt.Sprintf("%s의 이동 %d개는 stack이 null입니다. 스택을 남기지 않은 것인지 스택이 비어 있는 것인지 원본이 가르지 않아, 스택 없는 사건으로 옮겼습니다(설계 메모 4).", d.journey, nullStack)})
	}
	return nil
}

// msgFact 값은 메시지 하나의 내용입니다. 버퍼 쓰기 사본이 송신 메시지와 같은지 견줄 때 씁니다.
type msgFact struct {
	typ, length, text  string
	hasText, cut, cont bool
}

// message 메서드는 메시지 하나를 사건으로 옮깁니다. 장소와 출발·도착은 그 이동의 것을 그대로 씁니다.
func (l *loader) message(d decoder, raw any, where string, hop atlas.Event) (int, msgFact, error) {
	m, err := load.Object(raw, where)
	if err != nil {
		return 0, msgFact{}, err
	}
	if err := load.Keys(m, where, "type", "len", "text", "cut", "cont"); err != nil {
		return 0, msgFact{}, err
	}
	typ, ok := m["type"].(string)
	if !ok {
		return 0, msgFact{}, fmt.Errorf("%s.type이 글이 아님", where)
	}
	length, ok := m["len"].(json.Number)
	if !ok {
		return 0, msgFact{}, fmt.Errorf("%s.len이 수가 아님", where)
	}
	f := msgFact{typ: typ, length: string(length)}
	for _, flag := range []string{"cut", "cont"} {
		if v, ok := m[flag]; ok && v != true {
			return 0, msgFact{}, fmt.Errorf("%s.%s가 true가 아님", where, flag)
		}
	}
	_, f.cut = m["cut"]
	_, f.cont = m["cont"]
	id := l.B.AddEvent(atlas.Event{Place: hop.Place, From: hop.From, Target: hop.Target, Kind: atlas.Instant, Name: typ, Source: d.rec})
	l.Origins[atlas.RecordRef{Table: "event", ID: id}] = d.pre + ".hops[].msgs[]"
	l.B.EventAttrs(id, d.rec, "len", length)
	if t, ok := m["text"]; ok {
		s, ok := t.(string)
		if !ok {
			return 0, msgFact{}, fmt.Errorf("%s.text가 글이 아님", where)
		}
		f.text, f.hasText = s, true
		l.B.EventAttrs(id, d.rec, "text", s)
	}
	return id, f, nil
}

// syscall 메서드는 이동의 시스템 콜 길이(sysUs)를 strace 사건으로, 커널 마디(kernel)를 ftrace 구간의 나무로 옮깁니다.
func (l *loader) syscall(d decoder, h map[string]any, where string, hop, place, kproc int) error {
	sv, hasSys := h["sysUs"]
	kv, hasKernel := h["kernel"]
	if !hasSys {
		if hasKernel {
			return fmt.Errorf("%s: sysUs 없이 kernel이 있음", where)
		}
		return nil
	}
	dur, err := load.Micros(sv, where+".sysUs")
	if err != nil {
		return err
	}
	l.Source(atlas.Source{ID: srcStrace, Tool: "strace(시스템 콜 길이)", Nature: atlas.Observed})
	l.Source(atlas.Source{ID: d.pair, Tool: d.journey + " 해석기가 순서와 인자로 맞춘 짝", Nature: atlas.Inferred})
	if l.Once("unrecorded:strace") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcStrace, Field: "start_ns", Note: "해석기가 시스템 콜의 길이(sysUs)만 남겼습니다"})
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcStrace, Field: "name", Note: "해석기가 시스템 콜 이름을 남기지 않았습니다"})
	}
	sys := l.B.AddEvent(atlas.Event{Place: place, Kind: atlas.Span, Name: "syscall", DurNS: &dur, Source: srcStrace})
	l.Origins[atlas.RecordRef{Table: "event", ID: sys}] = d.pre + ".hops[].sysUs"
	link := l.B.AddLink(hop, sys, "syscall", d.pair)
	l.Origins[atlas.RecordRef{Table: "link", ID: link}] = SyscallPair
	if !hasKernel {
		return nil
	}
	roots, ok := kv.([]any)
	if !ok || len(roots) == 0 || kproc < 0 {
		return fmt.Errorf("%s.kernel이 비지 않은 목록이 아니거나 kernelPid가 없음", where)
	}
	if l.Once("unrecorded:ftrace") {
		l.B.AddUnrecorded(atlas.Unrecorded{Source: srcFtrace, Field: "start_ns", Note: "커널 마디는 그 함수에서 보낸 시간(us, 잠든 시간 포함)만 남았습니다"})
	}
	for k, r := range roots {
		root, err := l.kernel(r, fmt.Sprintf("%s.kernel[%d]", where, k), nil, 0, kproc)
		if err != nil {
			return err
		}
		link := l.B.AddLink(sys, root, "kernel", d.pair)
		l.Origins[atlas.RecordRef{Table: "link", ID: link}] = KernelPair
	}
	return nil
}

// kernelGroup 함수는 커널 마디의 깊이마다 원본의 경로 무리를 돌려줍니다. 둘째 단 아래는 재고 조사가 한 줄로 합쳤습니다.
func kernelGroup(level int) string {
	switch level {
	case 0:
		return "$.data.hops[].kernel[]"
	case 1:
		return "$.data.hops[].kernel[].kids[]"
	}
	return "$.data.hops[].kernel[].kids[]{2+}"
}

// kernel 메서드는 커널 마디 하나와 그 아래를 ftrace 출처의 구간으로 옮기고 그 사건 id를 돌려줍니다. 자식의 포함은 parent입니다.
func (l *loader) kernel(raw any, where string, parent *int, level, kproc int) (int, error) {
	n, err := load.Object(raw, where)
	if err != nil {
		return 0, err
	}
	if err := load.Keys(n, where, "fn", "us", "layer", "at", "kids", "calls", "callsOf", "hidden", "guess"); err != nil {
		return 0, err
	}
	group := kernelGroup(level)
	fn, ok := n["fn"].(string)
	if !ok {
		return 0, fmt.Errorf("%s.fn이 글이 아님", where)
	}
	// at이 있으면 커널 함수 아래 프레임이, 없으면(층을 물려받은 짐작 마디 등) 커널 함수가 장소입니다.
	var place int
	if at, ok := n["at"]; ok {
		s, ok := at.(string)
		if !ok {
			return 0, fmt.Errorf("%s.at이 글이 아님", where)
		}
		if place, err = l.Frame(kproc, fn, s, srcFtrace, group+".fn", group+".at"); err != nil {
			return 0, fmt.Errorf("%s: %w", where, err)
		}
	} else {
		place = l.Function(kproc, fn, "", srcFtrace, group+".fn")
	}
	dur, err := load.Micros(n["us"], where+".us")
	if err != nil {
		return 0, err
	}
	id := l.B.AddEvent(atlas.Event{Place: place, Kind: atlas.Span, Name: fn, DurNS: &dur, Parent: parent, Source: srcFtrace})
	l.Origins[atlas.RecordRef{Table: "event", ID: id}] = group
	for _, k := range []string{"layer", "calls", "callsOf", "hidden", "guess"} {
		if v, ok := n[k]; ok {
			l.B.EventAttrs(id, "journey", k, v)
		}
	}
	if kv, ok := n["kids"]; ok {
		kids, ok := kv.([]any)
		if !ok || len(kids) == 0 {
			return 0, fmt.Errorf("%s.kids가 비지 않은 목록이 아님(빈 목록은 없음과 가를 수 없음)", where)
		}
		for k, c := range kids {
			if _, err := l.kernel(c, fmt.Sprintf("%s.kids[%d]", where, k), &id, level+1, kproc); err != nil {
				return 0, err
			}
		}
	}
	return id, nil
}

// stages 메서드는 단계를 해석기 출처의 instant로 옮깁니다. 장소는 단계 함수이고 스택은 [단계 함수, 부른 쪽]입니다.
func (l *loader) stages(d decoder, stages []any, backend int) error {
	for i, raw := range stages {
		where := fmt.Sprintf("%s.stages[%d]", d.pre, i)
		s, err := load.Object(raw, where)
		if err != nil {
			return err
		}
		if err := load.Keys(s, where, "name", "t0", "fn", "caller", "at"); err != nil {
			return err
		}
		name, ok := s["name"].(string)
		if !ok {
			return fmt.Errorf("%s.name이 글이 아님", where)
		}
		t0, err := load.Micros(s["t0"], where+".t0")
		if err != nil {
			return err
		}
		ev := atlas.Event{Place: backend, Kind: atlas.Instant, Name: name, StartNS: atlas.Ptr(t0), EndNS: atlas.Ptr(t0), Source: d.journey}
		fn, hasFn := s["fn"].(string)
		at, hasAt := s["at"].(string)
		caller, hasCaller := s["caller"].(string)
		switch {
		case !hasFn && !hasAt && !hasCaller:
		case hasFn && hasAt:
			if ev.Place, err = l.stageFunction(backend, fn, d.rec, d.pre+".stages[].fn"); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
			var outer int
			if hasCaller {
				outer, err = l.Frame(backend, caller, at, d.rec, d.pre+".stages[].caller", d.pre+".stages[].at")
			} else {
				outer, err = l.Site(backend, at, d.rec, d.pre+".stages[].at")
			}
			if err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
			ev.Stack = []int{ev.Place, outer}
		default:
			return fmt.Errorf("%s: fn·at·caller의 짝이 맞지 않음", where)
		}
		id := l.B.AddEvent(ev)
		l.Origins[atlas.RecordRef{Table: "event", ID: id}] = d.pre + ".stages[]"
	}
	return nil
}

// lane 메서드는 구성도의 칸 장소를 돌려줍니다.
func (l *loader) lane(name, src, group string) int {
	return l.Place("lane\x00"+name, atlas.PlaceSpec{Kind: "lane", Name: name, Source: src}, group)
}

// file 메서드는 칸 아래의 파일 장소를 돌려줍니다.
func (l *loader) file(lane int, laneName, path, src, group string) int {
	return l.Place("file\x00"+laneName+"\x00"+path, atlas.PlaceSpec{Parent: &lane, Kind: "file", Name: path, Source: src}, group)
}

// stageFunction 메서드는 단계 함수의 장소를 돌려줍니다. 단계는 함수의 위치를 남기지 않아, 같은 이름의 함수 장소가 하나면 그것을 쓰고
// 없으면 위치 없이 만듭니다. 같은 이름이 둘 이상이면 어느 것인지 가를 수 없어 오류입니다.
func (l *loader) stageFunction(proc int, fn, src, group string) (int, error) {
	switch ids := l.Named(proc, fn); len(ids) {
	case 0:
		return l.Function(proc, fn, "", src, group), nil
	case 1:
		return ids[0], nil
	default:
		return 0, fmt.Errorf("단계 함수 %s의 장소가 %d개", fn, len(ids))
	}
}
