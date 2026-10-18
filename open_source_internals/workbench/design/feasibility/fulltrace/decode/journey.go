package main

// 이 파일은 요청 하나의 여정(받기부터 응답까지 구성 요소를 지나는 순서)을 만듭니다. 탐색용 코드입니다(P21~P25).
//
// 재료는 두 가지입니다. 함수 기록(<pid>.bin)은 단계의 경계와 부품 안의 일(버퍼 찾기, WAL 기록 등)을,
// strace -f -yy -ttt 기록은 소켓·파일 시스템 콜(누가, 어느 파일, 몇 바이트)을 줍니다. 둘은 같은 시계(CLOCK_REALTIME)입니다.
// 단계와 관측점은 쿼리마다가 아니라 한 번 정합니다. 요청의 경계는 ReadyForQuery, 단계는 요청 처리(tcop)의 단계 함수(stageNames),
// 관측점은 아래 probes 표입니다. 그래서 SELECT, DML, DDL, SHOW 같은 유틸리티 문장, 오류, 확장 프로토콜이 같은 규칙으로 조립됩니다.

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// sysEvent 는 strace 한 줄을 풀어 놓은 것입니다.
type sysEvent struct {
	Pid     int
	Ns      uint64
	Name    string
	Fd      int64
	FdPath  string
	Ret     int64
	RetPath string
	Create  bool
	Data    []byte
	Offset  int64
	DurNs   int64
	Args    string
}

// straceLine 은 "pid sec.usec name(args) = ret" 꼴의 strace 줄입니다.
var straceLine = regexp.MustCompile(`^(\d+)\s+(\d+)\.(\d+)\s+(.*)$`)

// readStrace 는 strace -f -yy -ttt 출력을 읽어 완결된 시스템 콜 목록을 돌려줍니다.
// 여러 프로세스의 호출이 겹치면 strace는 "<unfinished ...>"와 "<... name resumed>"로 나눠 쓰므로, 같은 pid끼리 이어 붙입니다.
// 시각(Ns)은 호출이 끝난 시각입니다. strace는 한 줄로 끝난 호출에는 들어간 시각을, 나뉜 호출의 resumed 줄에는 끝난 시각을
// 적으므로, 한 줄 호출에는 -T 시간을 더합니다. 근거(2026-10-14, 실험 1 rr-read-a의 strace): 5 ms 넘게 걸린 한 줄 호출 9건 모두
// 같은 pid의 다음 줄 시각이 "그 줄 시각 + -T 시간" 이후였고, 나뉜 호출 175건 모두 resumed 줄 시각이 "unfinished 줄 시각 + -T 시간"과
// 2 ms 안에서 같았습니다.
//
// Errors: 파일을 열 수 없으면 오류를 돌려줍니다. 해석하지 못한 줄은 건너뜁니다.
func readStrace(file string) ([]sysEvent, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	pending := map[int]string{}
	var out []sysEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		m := straceLine.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		pid, _ := strconv.Atoi(m[1])
		sec, _ := strconv.ParseUint(m[2], 10, 64)
		usec, _ := strconv.ParseUint(m[3], 10, 64)
		body := m[4]
		if strings.HasSuffix(body, "<unfinished ...>") {
			// strace는 표지 앞에 공백 하나를 둡니다. 그것까지 떼야 "fd<경로>" 바로 뒤에 resumed 줄의 ")"가 붙어 경로를 읽습니다.
			pending[pid] = strings.TrimSuffix(strings.TrimSuffix(body, "<unfinished ...>"), " ")
			continue
		}
		resumed := strings.HasPrefix(body, "<... ")
		if resumed {
			i := strings.Index(body, " resumed>")
			if i < 0 {
				continue
			}
			body = pending[pid] + body[i+len(" resumed>"):]
			delete(pending, pid)
		}
		if ev, ok := parseSyscall(pid, sec*1e9+usec*1e3, body); ok {
			if !resumed {
				ev.Ns += uint64(ev.DurNs)
			}
			out = append(out, ev)
		}
	}
	return out, sc.Err()
}

// fdArg 는 -yy가 붙인 "번호<경로>"에서 번호와 경로를 꺼냅니다. 소켓 표시 안에는 "->"와 따옴표가 들어 있으므로(관측),
// 표시의 끝은 ">" 바로 뒤에 ", "가 오거나 문자열이 끝나는 자리로 봅니다.
var fdArg = regexp.MustCompile(`^(-?\d+)<(.*?)>(, |$)`)

// durSuffix 는 strace -T가 줄 끝에 붙이는 "커널 안에서 걸린 시간"(초)입니다.
var durSuffix = regexp.MustCompile(` <(\d+)\.(\d+)>$`)

// parseSyscall 은 시스템 콜 한 줄의 이름, 대상 경로, 반환값, 데이터를 풉니다.
func parseSyscall(pid int, ns uint64, body string) (sysEvent, bool) {
	open := strings.IndexByte(body, '(')
	eq := strings.LastIndex(body, ") = ")
	if open < 0 || eq < 0 {
		return sysEvent{}, false
	}
	ev := sysEvent{Pid: pid, Ns: ns, Name: body[:open]}
	args := body[open+1 : eq]
	ev.Args = strings.TrimSpace(args)
	ret := body[eq+4:]
	if dm := durSuffix.FindStringSubmatchIndex(ret); dm != nil {
		sec, _ := strconv.ParseInt(ret[dm[2]:dm[3]], 10, 64)
		frac := ret[dm[4]:dm[5]]
		f, _ := strconv.ParseInt((frac + "000000000")[:9], 10, 64)
		ev.DurNs = sec*1e9 + f
		ret = ret[:dm[0]]
	}
	if fm := fdArg.FindStringSubmatch(ret); fm != nil {
		ev.RetPath = fm[2]
	}
	retNum := ret
	if i := strings.IndexAny(retNum, " <"); i >= 0 {
		retNum = retNum[:i]
	}
	ev.Ret, _ = strconv.ParseInt(retNum, 10, 64)
	rest := args
	if fm := fdArg.FindStringSubmatch(args); fm != nil {
		ev.Fd, _ = strconv.ParseInt(fm[1], 10, 64)
		ev.FdPath = fm[2]
		rest = args[len(fm[0]):]
	}
	ev.Create = strings.Contains(args, "O_CREAT")
	if q := strings.IndexByte(rest, '"'); q >= 0 {
		ev.Data = unescapeStrace(rest[q+1:])
	}
	if ev.Name == "pread64" || ev.Name == "pwrite64" || ev.Name == "preadv" || ev.Name == "pwritev" {
		if i := strings.LastIndex(args, ", "); i >= 0 {
			ev.Offset, _ = strconv.ParseInt(strings.TrimSpace(args[i+2:]), 10, 64)
		}
	}
	return ev, true
}

// unescapeStrace 는 strace가 C 문자열처럼 쓴 바이트("\0", "\37", "\n" 등)를 닫는 따옴표까지 원래 바이트로 되돌립니다.
func unescapeStrace(s string) []byte {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			break
		}
		if c != '\\' || i+1 >= len(s) {
			b = append(b, c)
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b = append(b, '\n')
		case 't':
			b = append(b, '\t')
		case 'r':
			b = append(b, '\r')
		case 'v':
			b = append(b, '\v')
		case 'f':
			b = append(b, '\f')
		case '"', '\\':
			b = append(b, s[i])
		default:
			// 8진수는 최대 세 자리입니다.
			j, v := i, 0
			for j < len(s) && j < i+3 && s[j] >= '0' && s[j] <= '7' {
				v = v*8 + int(s[j]-'0')
				j++
			}
			if j == i {
				b = append(b, s[i])
				continue
			}
			b = append(b, byte(v))
			i = j - 1
		}
	}
	return b
}

// pgMessage 는 소켓으로 오간 PostgreSQL 프로토콜 메시지 하나입니다(종류 한 글자, 길이, 내용 요약).
// Cut은 메시지가 보이는 바이트 밖으로 이어진다는 뜻이고, Cont는 앞 송신에서 시작한 메시지의 나머지라는 뜻입니다.
type pgMessage struct {
	Type string `json:"type"`
	Len  int    `json:"len"`
	Text string `json:"text,omitempty"`
	Cut  bool   `json:"cut,omitempty"`
	Cont bool   `json:"cont,omitempty"`
}

// parseMessages 는 소켓 바이트를 프로토콜 메시지로 나눕니다. 메시지는 종류 1바이트와 자기 자신을 포함한 길이 4바이트로 시작합니다.
// 같은 글자가 방향에 따라 다른 메시지이므로(클라이언트의 D는 Describe, 서버의 D는 DataRow) 방향을 받습니다.
func parseMessages(b []byte, fromClient bool) []pgMessage {
	msgs, _ := splitMessages(b, fromClient)
	return msgs
}

// splitMessages 는 parseMessages와 같고, 마지막 메시지가 끝나는 자리(b 기준, 선언된 길이로 계산)를 함께 돌려줍니다.
// 그 자리가 len(b)보다 크면 마지막 메시지는 보이는 바이트 밖으로 이어지고 Cut이 붙습니다.
func splitMessages(b []byte, fromClient bool) ([]pgMessage, int) {
	var out []pgMessage
	off := 0
	for len(b)-off >= 5 {
		t := b[off]
		n := int(binary.BigEndian.Uint32(b[off+1 : off+5]))
		end := off + 1 + n
		msg := pgMessage{Type: string(t), Len: n}
		visible := end
		if n < 4 || end > len(b) {
			visible = len(b)
			msg.Cut = true
		}
		body := b[off+5 : visible]
		msg.Text = messageText(t, body, fromClient)
		out = append(out, msg)
		if msg.Cut {
			return out, end
		}
		off = end
	}
	return out, off
}

// messageText 는 메시지 종류마다 사람이 읽을 요약을 만듭니다. 모르는 종류는 빈 문자열입니다.
func messageText(t byte, body []byte, fromClient bool) string {
	str := func(b []byte) string { return strings.TrimRight(string(b), "\x00") }
	if fromClient {
		switch t {
		case 'Q':
			return str(body)
		case 'P': // Parse: 문장 이름, 쿼리 문자열, 인자 형식
			if parts := strings.SplitN(string(body), "\x00", 3); len(parts) >= 2 {
				return parts[1]
			}
		case 'B': // Bind: 포털 이름, 문장 이름, 인자 형식, 인자 값
			if parts := strings.SplitN(string(body), "\x00", 3); len(parts) == 3 {
				text := "포털 " + quoteName(parts[0]) + " ← 문장 " + quoteName(parts[1])
				if vals := bindValues([]byte(parts[2])); len(vals) > 0 {
					text += " · 값 " + strings.Join(vals, ", ")
				}
				return text
			}
		case 'D', 'C': // Describe와 Close: 대상 종류(S 문장, P 포털)와 이름
			if len(body) > 0 {
				kind := map[byte]string{'S': "문장 ", 'P': "포털 "}[body[0]]
				return kind + quoteName(str(body[1:]))
			}
		case 'E': // Execute: 포털 이름과 최대 행 수
			return "포털 " + quoteName(strings.SplitN(string(body), "\x00", 2)[0])
		}
		return ""
	}
	switch t {
	case 'C':
		return str(body)
	case 'Z':
		if len(body) > 0 {
			return string(body[:1])
		}
	case 'D':
		return dataRow(body)
	case 'T':
		return rowDescription(body)
	case 'E', 'N':
		return errorFields(body)
	}
	return ""
}

// bindValues 는 Bind 메시지에서 이름 두 개 뒤의 인자 값들을 읽습니다.
// 형식은 int16 형식 수, int16 형식들, int16 값 수, 그리고 값마다 int32 길이(-1은 NULL)와 바이트입니다.
// 이진 형식 값은 바이트 수만 적고, 길이가 맞지 않으면 읽은 데까지만 돌려줍니다.
func bindValues(b []byte) []string {
	u16 := func() (int, bool) {
		if len(b) < 2 {
			return 0, false
		}
		v := int(binary.BigEndian.Uint16(b))
		b = b[2:]
		return v, true
	}
	nf, ok := u16()
	if !ok || len(b) < 2*nf {
		return nil
	}
	formats := make([]int, nf)
	for i := range formats {
		formats[i], _ = u16()
	}
	np, ok := u16()
	if !ok {
		return nil
	}
	var vals []string
	for i := 0; i < np && len(b) >= 4; i++ {
		n := int32(binary.BigEndian.Uint32(b))
		b = b[4:]
		if n < 0 {
			vals = append(vals, "NULL")
			continue
		}
		if int(n) > len(b) {
			break
		}
		// 형식이 하나면 모든 값에, 여럿이면 값마다 적용합니다(0은 텍스트, 1은 이진).
		binaryFmt := (nf == 1 && formats[0] == 1) || (nf > 1 && i < nf && formats[i] == 1)
		if binaryFmt {
			vals = append(vals, fmt.Sprintf("(이진 %d B)", n))
		} else {
			vals = append(vals, string(b[:n]))
		}
		b = b[n:]
	}
	return vals
}

// quoteName 은 빈 이름을 "(이름 없음)"으로 바꿉니다. psql의 \bind는 이름 없는 문장과 포털을 씁니다.
func quoteName(s string) string {
	if s == "" {
		return "(이름 없음)"
	}
	return s
}

// parseSend 는 송신 하나의 바이트를 메시지로 나눕니다. 응답 메시지 하나가 여러 송신에 걸칠 수 있으므로
// (통신 버퍼 8 KB가 차면 그 자리에서 내보냄), 앞 송신에서 남은 바이트 수 carry만큼을 먼저 그 메시지의 나머지로 떼어 냅니다.
// sent는 실제로 나간 바이트 수(시스템 콜 반환값)이고, data는 strace가 보여 준 앞부분입니다.
//
// Returns: 메시지 목록, 다음 송신으로 넘어가는 바이트 수(경계를 알 수 없으면 -1), 넘어가는 메시지의 종류와 길이.
func parseSend(data []byte, sent int64, carry int, last pgMessage) ([]pgMessage, int, pgMessage) {
	var out []pgMessage
	off := 0
	if carry < 0 {
		return []pgMessage{{Type: "?", Text: "앞 송신이 strace에 다 보이지 않아 메시지 경계를 알 수 없음"}}, 0, last
	}
	if carry > 0 {
		n := min(int64(carry), sent)
		out = append(out, pgMessage{Type: last.Type, Len: last.Len, Cont: true, Text: fmt.Sprintf("앞 송신에서 이어진 %d B", n)})
		off = int(n)
		if int64(carry) > sent {
			return out, carry - int(sent), last
		}
	}
	if off >= len(data) {
		if int64(off) < sent {
			return out, -1, last
		}
		return out, 0, last
	}
	msgs, end := splitMessages(data[off:], false)
	out = append(out, msgs...)
	abs := int64(off + end)
	switch {
	case abs > sent && len(msgs) > 0:
		return out, int(abs - sent), msgs[len(msgs)-1]
	case abs < sent:
		// 보이는 바이트가 끝났는데 송신은 더 있었다: strace -s 한도로 뒤 메시지가 보이지 않는다.
		return out, -1, last
	}
	return out, 0, last
}

// dataRow 는 DataRow 메시지 내용(열 개수 2바이트, 열마다 길이 4바이트와 값)을 "1|a"처럼 줄입니다.
func dataRow(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	n := int(binary.BigEndian.Uint16(b))
	b = b[2:]
	var cols []string
	for i := 0; i < n && len(b) >= 4; i++ {
		l := int(int32(binary.BigEndian.Uint32(b)))
		b = b[4:]
		if l < 0 {
			cols = append(cols, "NULL")
			continue
		}
		if l > len(b) {
			l = len(b)
		}
		cols = append(cols, string(b[:l]))
		b = b[l:]
	}
	return strings.Join(cols, "|")
}

// rowDescription 은 RowDescription 메시지에서 열 이름만 꺼냅니다(열마다 이름 뒤에 18바이트의 형식 정보가 있음).
func rowDescription(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	n := int(binary.BigEndian.Uint16(b))
	b = b[2:]
	var names []string
	for i := 0; i < n; i++ {
		z := strings.IndexByte(string(b), 0)
		if z < 0 {
			break
		}
		names = append(names, string(b[:z]))
		if z+1+18 > len(b) {
			break
		}
		b = b[z+1+18:]
	}
	return strings.Join(names, ",")
}

// errorFields 는 ErrorResponse의 심각도(S)와 메시지(M) 필드를 꺼냅니다.
func errorFields(b []byte) string {
	var sev, msg string
	for _, f := range strings.Split(string(b), "\x00") {
		if len(f) < 2 {
			continue
		}
		switch f[0] {
		case 'S':
			sev = f[1:]
		case 'M':
			msg = f[1:]
		}
	}
	return strings.TrimSpace(sev + " " + msg)
}

// probe 는 부품 하나에 한 번 정한 관측점입니다. 함수 이름이 불리면 그 부품으로 가는 이동 하나로 셉니다.
type probe struct {
	Fn, Comp, Kind string
}

// probes 는 쿼리와 관계없이 쓰는 관측점 표입니다. 이 표만 바꾸면 모든 쿼리의 여정이 같이 바뀝니다.
var probes = []probe{
	{"ReadBuffer", "shared_buffer", "버퍼 찾기"},
	{"ReadBufferExtended", "shared_buffer", "버퍼 찾기"},
	{"ReleaseAndReadBuffer", "shared_buffer", "버퍼 찾기"},
	{"read_stream_next_buffer", "shared_buffer", "버퍼 찾기"},
	{"heap_insert", "shared_buffer", "행 넣기"},
	{"heap_multi_insert", "shared_buffer", "행 여러 개 넣기"},
	{"heap_update", "shared_buffer", "행 바꾸기"},
	{"heap_delete", "shared_buffer", "행 지우기"},
	{"_bt_doinsert", "shared_buffer", "인덱스 항목 넣기"},
	{"pgaio_worker_submit", "io_worker", "읽기 맡기기"},
	{"XLogInsert", "wal_buffer", "WAL 기록"},
	// 응답 메시지는 하나씩 소켓으로 나가지 않고 통신 버퍼(PqSendBuffer, 8 KB)에 쌓였다가 flush 때 한 번에 나갑니다(pqcomm.c).
	// 결과 행(D)을 만드는 printtup도 이 함수로 버퍼에 쓰므로, 누가 썼는지는 스택에서 봅니다.
	{"socket_putmessage", "comm_buffer", "응답 버퍼에 쓰기"},
	// 공유 메모리의 트랜잭션 표들입니다. 락 테이블, 프로세스 배열(실행 중인 트랜잭션 목록), 커밋 로그 버퍼(CLOG)입니다.
	{"LockAcquireExtended", "lock_table", "락 잡기"},
	{"GetSnapshotData", "proc_array", "스냅숏 만들기"},
	{"GetNewTransactionId", "proc_array", "트랜잭션 번호 받기"},
	{"ProcArrayEndTransaction", "proc_array", "트랜잭션 끝 알림"},
	{"TransactionIdCommitTree", "clog", "커밋 상태 기록"},
	{"TransactionLogFetch", "clog", "커밋 상태 읽기"},
	{"ExecHashTableCreate", "work_mem", "해시 테이블 만들기"},
	// DDL처럼 카탈로그를 바꾸면 다른 백엔드의 카탈로그 캐시를 지우라는 알림을 공유 메모리의 무효화 큐에 넣습니다.
	{"SIInsertDataEntries", "sinval", "캐시 무효화 알림"},
	// ereport가 실제로 보고될 때만 불립니다(오류, 경고, 로그). 어디서 보고했는지는 스택에서 봅니다.
	{"errfinish", "backend", "보고(ereport)"},
	{"tuplesort_begin_index_btree", "work_mem", "정렬 시작"},
	{"tuplesort_begin_heap", "work_mem", "정렬 시작"},
}

// hop 은 여정의 한 이동입니다. 잇달아 일어난 같은 종류의 일은 한 이동으로 묶고 Count를 늘립니다.
// Catalog는 시스템 카탈로그를 읽는 일인지입니다. 화면은 이 표시로 카탈로그 조회를 접어 요청 자체의 길을 먼저 보여 줍니다.
type hop struct {
	N       int `json:"n"`
	Pid     int `json:"pid"`
	ns      uint64
	enter   bool
	Stage   string      `json:"stage"`
	Proc    string      `json:"proc"`
	From    string      `json:"from"`
	To      string      `json:"to"`
	Kind    string      `json:"kind"`
	Catalog bool        `json:"catalog,omitempty"`
	Count   int         `json:"count"`
	Bytes   int64       `json:"bytes,omitempty"`
	Files   []string    `json:"files,omitempty"`
	Msgs    []pgMessage `json:"msgs,omitempty"`
	T0us    float64     `json:"t0"`
	T1us    float64     `json:"t1"`
	Stack   []frameRef  `json:"stack"`
	Site    string      `json:"site"`
	SysUs   float64     `json:"sysUs,omitempty"`
	Kernel  []*kcall    `json:"kernel,omitempty"`
	data    []byte
}

// kernelInput 은 여정에 붙일 커널 추적입니다. Calls가 비어 있으면 커널 나무를 붙이지 않고, Index가 nil이면 소스 위치 없이 붙입니다.
type kernelInput struct {
	Calls []ksyscall
	Index kernelIndex
	Src   string
}

// frameRef 는 호출 스택의 칸 하나입니다. At은 이 함수 안에서 다음 칸을 부른 줄이고, 가장 안쪽 칸은 함수가 정의된 줄입니다.
type frameRef struct {
	Fn string `json:"fn"`
	At string `json:"at"`
}

// stageNames 는 단계 함수에서 단계 이름으로 가는 표입니다. 요청 구간 안에서 그 함수가 처음 들어온 자리에서 단계가 바뀝니다.
// 단순 쿼리(exec_simple_query)와 확장 프로토콜(exec_parse_message, exec_bind_message, exec_execute_message)이
// 같은 함수들을 다른 순서로 부르므로, 단계는 표의 순서가 아니라 시각 순서로 놓입니다.
var stageNames = []struct{ Fn, Stage string }{
	{"pg_parse_query", "구문 분석"},
	{"pg_analyze_and_rewrite_fixedparams", "의미 분석"},
	{"pg_analyze_and_rewrite_varparams", "의미 분석"},
	{"pg_analyze_and_rewrite_withcb", "의미 분석"},
	{"pg_plan_queries", "계획"},
	{"PortalStart", "실행"},
	{"finish_xact_command", "커밋"},
	{"EndCommand", "응답"},
}

// stageMark 는 단계 하나가 시작한 시각입니다. 경계는 각 단계 함수가 처음 들어온 시각입니다.
// Fn은 단계 함수, Caller와 At은 그 함수를 부른 함수와 줄입니다. 단순 쿼리는 exec_simple_query,
// 확장 프로토콜은 메시지마다 exec_parse_message·exec_bind_message·exec_execute_message가 부릅니다.
type stageMark struct {
	Name   string  `json:"name"`
	Ns     uint64  `json:"-"`
	T0us   float64 `json:"t0"`
	Fn     string  `json:"fn,omitempty"`
	Caller string  `json:"caller,omitempty"`
	At     string  `json:"at,omitempty"`
}

// cmdJourney 는 pid의 nth번째 요청에 대해 받기부터 응답까지의 이동 목록을 JSON으로 출력합니다.
//
// Arguments: traces는 모든 프로세스의 함수 기록, sys는 strace 사건 전체, pid와 nth는 대상 백엔드와 접속 뒤 몇 번째 요청인지입니다.
// srcDir가 비어 있지 않으면 스택과 단계가 가리키는 소스 줄의 앞뒤를 그 트리에서 읽어 함께 출력합니다.
// kern에 커널 추적이 있으면 대상 백엔드의 pwrite64·pwritev·fdatasync 이동마다 그 안의 커널 함수 나무를 붙이고(matchKernel),
// kern.Src가 있으면 나무에 남은 함수의 정의 줄 앞뒤를 "ksrc"로 함께 출력합니다.
// Errors: 그 요청의 경계(ReadyForQuery)나 수신·송신을 찾지 못하면 오류를 돌려줍니다.
// 소스 파일을 읽지 못한 줄은 발췌 없이 남깁니다.
func cmdJourney(w io.Writer, traces []*traceFile, syms *symbols, sys []sysEvent, pid, nth int, srcDir string, kern kernelInput) error {
	t, err := findTrace(traces, pid)
	if err != nil {
		return err
	}

	// 요청의 경계는 ReadyForQuery입니다. 서버는 요청 하나를 끝낼 때마다(단순 쿼리의 끝, 확장 프로토콜의 Sync, 오류 뒤)
	// 이 함수로 Z를 보내므로, 프로토콜과 오류에 관계없이 같은 규칙으로 자릅니다. 첫 번째 ReadyForQuery는 접속 인증의 끝입니다.
	// 응답의 끝을 strace의 바이트에서 찾지 않는 이유: strace는 송신의 앞 1024 B만 보여 주므로 큰 응답 끝의 Z가 보이지 않습니다.
	var rfq []uint64
	for _, r := range t.Records {
		if r.Kind == kindEnter && syms.name(r.Fn-t.Slide) == "ReadyForQuery" {
			rfq = append(rfq, r.Ns)
		}
	}
	if nth < 1 || nth >= len(rfq) {
		return fmt.Errorf("pid %d에 %d번째 요청이 없습니다(ReadyForQuery %d번)", pid, nth, len(rfq))
	}
	var recv, send *sysEvent
	for i := range sys {
		e := &sys[i]
		if e.Pid != pid || e.Ret <= 0 {
			continue
		}
		if recv == nil && e.Name == "recvfrom" && e.Ns > rfq[nth-1] && e.Ns < rfq[nth] {
			recv = e
		}
		if e.Name == "sendto" && e.Ns > rfq[nth] {
			send = e
			break
		}
	}
	if recv == nil || send == nil {
		return fmt.Errorf("pid %d의 %d번째 요청의 수신이나 응답 송신을 strace에서 찾지 못했습니다", pid, nth)
	}
	base := recv.Ns

	// 단계 경계: 요청 구간 안에서 단계 함수가 처음 들어온 시각입니다. At은 그 함수를 부른 줄입니다.
	stages := []stageMark{{Name: "받기", Ns: recv.Ns}}
	used := map[string]bool{}
	stageStack := stackSampler(t, syms)
	for _, r := range t.Records {
		if r.Kind != kindEnter || r.Ns < recv.Ns || r.Ns > send.Ns {
			continue
		}
		name := syms.name(r.Fn - t.Slide)
		for _, sn := range stageNames {
			if name == sn.Fn && !used[sn.Stage] {
				used[sn.Stage] = true
				m := stageMark{Name: sn.Stage, Ns: r.Ns, Fn: sn.Fn, At: syms.lineAt(r.Site - t.Slide - 4)}
				if st, _ := stageStack(r.Ns, true); len(st) >= 2 {
					m.Caller = st[len(st)-2].Fn
				}
				stages = append(stages, m)
			}
		}
	}
	stageOf := func(ns uint64) string {
		name := stages[0].Name
		for _, s := range stages {
			if s.Ns <= ns {
				name = s.Name
			}
		}
		return name
	}

	// 이동 후보를 모읍니다. 함수 관측점은 대상 백엔드에서, 시스템 콜은 대상 백엔드와 다른 서버 프로세스에서 구간 안의 것을 씁니다.
	var raw []hop
	for _, r := range t.Records {
		if r.Kind != kindEnter || r.Ns < recv.Ns || r.Ns > send.Ns {
			continue
		}
		name := syms.name(r.Fn - t.Slide)
		for _, p := range probes {
			if p.Fn == name {
				raw = append(raw, hop{Pid: pid, ns: r.Ns, enter: true, Proc: "backend", From: "backend", To: p.Comp, Kind: p.Kind, Count: 1})
			}
		}
	}
	// 커널 나무는 strace 사건과 짝지은 것만 붙입니다. 짝짓기는 요청 구간과 관계없이 처음부터 차례대로 합니다.
	kmatch, kpid := matchKernel(sys, pid, kern.Calls)
	ksrc := &ksource{dir: kern.Src, idx: kern.Index, files: map[string][]string{}, calls: map[string][]bodyCall{}}
	for i, e := range sys {
		if e.Ns < recv.Ns || e.Ns > send.Ns {
			continue
		}
		if h, ok := sysHop(e, pid); ok {
			h.Pid, h.ns = e.Pid, e.Ns
			h.SysUs = float64(e.DurNs) / 1e3
			if c := kmatch[i]; c != nil {
				annotate(c.Root, kern.Index, "", "")
				prune(c.Root, 0)
				callSites(c.Root, ksrc)
				h.Kernel = []*kcall{c.Root}
			}
			raw = append(raw, h)
		}
	}
	sort.SliceStable(raw, func(i, j int) bool { return raw[i].ns < raw[j].ns })

	// 호출 스택은 프로세스마다 기록을 한 번만 앞으로 따라가며 구하므로, 시간 순서로 정렬한 뒤에 채웁니다.
	samplers := map[int]func(uint64, bool) ([]frameRef, string){}
	for _, o := range traces {
		samplers[o.Pid] = stackSampler(o, syms)
	}
	for k := range raw {
		h := &raw[k]
		if sample := samplers[h.Pid]; sample != nil {
			h.Stack, h.Site = sample(h.ns, h.enter)
		}
		h.T0us, h.T1us = us(h.ns, base), us(h.ns, base)
		// 카탈로그 읽기는 카탈로그 캐시(utils/cache)나 systable_* 함수(genam.c)를 거칩니다(P21~P25 실행의 카탈로그 읽기 83개 모두).
		// DDL의 카탈로그 쓰기는 CatalogTuple*·CatalogIndexInsert·CatalogOpenIndexes(catalog/indexing.c)를 거칩니다.
		// 이 파일만 봅니다. index_create 같은 catalog/ 디렉터리의 함수는 새 인덱스 파일 만들기와 정렬까지 부르므로 디렉터리로 가르면 요청 전체가 접힙니다.
		// 캐시나 쓰기 경로가 카탈로그를 여느라 잡는 락도 같은 규칙으로 묶입니다. 다른 프로세스의 읽기는 스택 대신 파일 번호로 가립니다.
		// utils/cache/inval.c는 캐시 조회가 아니라 커밋 때 다른 백엔드에 캐시 무효화를 알리는 길이라 접지 않습니다(DDL이 남에게 미치는 영향).
		h.Catalog = strings.Contains(h.Kind, "카탈로그")
		for _, f := range h.Stack {
			cache := strings.HasPrefix(f.At, "src/backend/utils/cache/") && !strings.HasPrefix(f.At, "src/backend/utils/cache/inval.c:")
			if strings.HasPrefix(f.Fn, "systable_") || cache || strings.HasPrefix(f.At, "src/backend/catalog/indexing.c:") {
				h.Catalog = true
			}
		}
	}

	// 버퍼에 쓴 메시지와 소켓으로 나간 메시지를 맞춥니다. socket_putmessage 한 번이 메시지 하나를 쓰므로,
	// 송신 직전까지 쌓인 쓰기 수와 송신 바이트 안의 메시지 수가 같을 때만 순서대로 짝짓습니다.
	// 앞 송신에서 이어진 나머지(Cont)는 새 쓰기가 아니므로 세지 않습니다.
	var pending []int
	carry, last := 0, pgMessage{}
	for k := range raw {
		h := &raw[k]
		switch {
		case h.Pid == pid && h.From == "backend" && h.To == "comm_buffer":
			pending = append(pending, k)
		case h.Pid == pid && h.Kind == "응답 보냄":
			h.Msgs, carry, last = parseSend(h.data, h.Bytes, carry, last)
			var fresh []pgMessage
			for _, m := range h.Msgs {
				if !m.Cont {
					fresh = append(fresh, m)
				}
			}
			if len(pending) == len(fresh) {
				for j, idx := range pending {
					raw[idx].Msgs = []pgMessage{fresh[j]}
				}
			}
			pending = nil
		}
	}

	// 오류 응답(E)을 버퍼에 쓴 순간부터는 정상 단계가 아니라 오류 처리 단계입니다. 오류 뒤에는 남은 단계 함수가 불리지 않습니다.
	for _, h := range raw {
		if h.Kind == "응답 버퍼에 쓰기" && msgType(h) == "E" {
			stages = append(stages, stageMark{Name: "오류 처리", Ns: h.ns})
			break
		}
	}
	sort.SliceStable(stages, func(i, j int) bool { return stages[i].Ns < stages[j].Ns })
	for k := range raw {
		raw[k].Stage = stageOf(raw[k].ns)
	}

	// 같은 단계에서 잇달아 일어난 같은 종류의 이동은 하나로 묶습니다. 이 규칙이 수십 번의 버퍼 찾기를 한 줄로 만듭니다.
	// 메시지 종류가 다른 버퍼 쓰기(C는 EndCommand, Z는 ReadyForQuery)는 쓴 함수가 달라서 묶지 않습니다.
	var hops []hop
	for _, h := range raw {
		if n := len(hops); n > 0 {
			last := &hops[n-1]
			if last.Stage == h.Stage && last.Kind == h.Kind && last.From == h.From && last.To == h.To && last.Pid == h.Pid && last.Catalog == h.Catalog && msgType(*last) == msgType(h) {
				last.Count += h.Count
				last.Bytes += h.Bytes
				last.T1us = h.T1us
				last.Msgs = append(last.Msgs, h.Msgs...)
				last.SysUs += h.SysUs
				last.Kernel = append(last.Kernel, h.Kernel...)
				for _, f := range h.Files {
					if !contains(last.Files, f) {
						last.Files = append(last.Files, f)
					}
				}
				continue
			}
		}
		h.N = len(hops) + 1
		hops = append(hops, h)
	}
	for i := range stages {
		stages[i].T0us = us(stages[i].Ns, base)
	}
	out := map[string]any{
		"query":  queryText(recv.Data),
		"pid":    pid,
		"endUs":  us(send.Ns, base),
		"stages": stages,
		"hops":   hops,
	}
	if srcDir != "" {
		var refs []string
		for _, s := range stages {
			refs = append(refs, s.At)
		}
		for _, h := range hops {
			for _, f := range h.Stack {
				refs = append(refs, f.At)
			}
		}
		out["src"] = excerpts(srcDir, refs)
	}
	if kpid != 0 {
		out["kernelPid"] = kpid
		if kern.Src != "" {
			var roots []*kcall
			for _, h := range hops {
				roots = append(roots, h.Kernel...)
			}
			out["ksrc"] = kernelExcerpts(roots, ksrc)
		}
	}
	return json.NewEncoder(w).Encode(out)
}

// excerpt 는 소스 한 줄의 앞뒤 발췌입니다. From은 Lines[0]의 줄 번호입니다.
type excerpt struct {
	From  int      `json:"from"`
	Lines []string `json:"lines"`
}

// excerptBefore 와 excerptAfter 는 가리킨 줄 앞뒤로 보여 줄 줄 수입니다.
const excerptBefore, excerptAfter = 8, 12

// excerpts 는 "상대경로:줄" 목록마다 srcDir 아래 파일에서 앞뒤 줄을 읽어 돌려줍니다. 파일은 한 번씩만 읽습니다.
// 형식이 맞지 않거나 읽지 못한 줄은 결과에서 빠집니다.
func excerpts(srcDir string, refs []string) map[string]excerpt {
	files := map[string][]string{}
	out := map[string]excerpt{}
	for _, ref := range refs {
		if _, done := out[ref]; done {
			continue
		}
		i := strings.LastIndexByte(ref, ':')
		if i < 0 {
			continue
		}
		line, err := strconv.Atoi(ref[i+1:])
		if err != nil || line < 1 {
			continue
		}
		name := ref[:i]
		text, ok := files[name]
		if !ok {
			if b, err := os.ReadFile(filepath.Join(srcDir, name)); err == nil {
				text = strings.Split(string(b), "\n")
			}
			files[name] = text
		}
		if line > len(text) {
			continue
		}
		from := max(1, line-excerptBefore)
		to := min(len(text), line+excerptAfter)
		out[ref] = excerpt{From: from, Lines: text[from-1 : to]}
	}
	return out
}

// sysHop 은 시스템 콜 하나를 이동 하나로 옮깁니다. 대상 백엔드의 소켓과 데이터·WAL 파일, 다른 서버 프로세스의 데이터·WAL 파일만 씁니다.
// 다른 프로세스의 일은 Proc을 other로 두고 Pid만 남깁니다. 기록만으로는 그 pid가 I/O 워커인지 WAL 라이터인지 알 수 없으므로,
// 역할은 화면이 기록할 때 ps로 본 역할표로 붙입니다(INSERT의 WAL 쓰기는 WAL 라이터 pid 5519가 했다).
func sysHop(e sysEvent, pid int) (hop, bool) {
	proc := "backend"
	if e.Pid != pid {
		proc = "other"
	}
	path := e.FdPath
	socket := strings.HasPrefix(path, "UNIX") || strings.HasPrefix(path, "TCP")
	dataFile := strings.Contains(path, "/base/") || strings.Contains(path, "/global/")
	read := e.Name == "pread64" || e.Name == "preadv"
	switch {
	case e.Name == "recvfrom" && e.Pid == pid && e.Ret > 0 && socket:
		return hop{Proc: proc, From: "client", To: "comm_buffer", Kind: "요청 받음", Count: 1, Bytes: e.Ret, Msgs: parseMessages(e.Data, true)}, true
	case e.Name == "sendto" && e.Pid == pid && e.Ret > 0 && socket:
		// 메시지는 cmdJourney가 송신 순서대로 앞 송신의 나머지를 이어 붙여 나눕니다(parseSend).
		return hop{Proc: proc, From: "comm_buffer", To: "client", Kind: "응답 보냄", Count: 1, Bytes: e.Ret, data: e.Data}, true
	case e.Name == "openat" && e.Create && strings.Contains(e.RetPath, "/base/"):
		return hop{Proc: proc, From: "backend", To: "data_files", Kind: "새 파일", Count: 1, Files: []string{shortPath(e.RetPath)}}, true
	case read && dataFile:
		from := "data_files"
		to := "shared_buffer"
		return hop{Proc: proc, From: from, To: to, Kind: "페이지 읽기 " + relKind(path), Count: 1, Bytes: e.Ret, Files: []string{shortPath(path)}}, true
	case e.Name == "pwrite64" && strings.Contains(path, "/pg_wal/"):
		return hop{Proc: proc, From: "wal_buffer", To: "wal_files", Kind: "WAL 쓰기", Count: 1, Bytes: e.Ret, Files: []string{shortPath(path)}}, true
	case e.Name == "fdatasync" && strings.Contains(path, "/pg_wal/"):
		return hop{Proc: proc, From: "wal_files", To: "wal_files", Kind: "디스크까지 동기화", Count: 1, Files: []string{shortPath(path)}}, true
	case read && strings.Contains(path, "/pg_xact/"):
		return hop{Proc: proc, From: "clog_files", To: "clog", Kind: "커밋 로그 페이지 읽기", Count: 1, Bytes: e.Ret, Files: []string{shortPath(path)}}, true
	case e.Name == "pwrite64" && strings.Contains(path, "/pg_xact/"):
		return hop{Proc: proc, From: "clog", To: "clog_files", Kind: "커밋 로그 페이지 쓰기", Count: 1, Bytes: e.Ret, Files: []string{shortPath(path)}}, true
	case e.Name == "pwrite64" && strings.Contains(path, "/base/"):
		return hop{Proc: proc, From: "backend", To: "data_files", Kind: "페이지 쓰기", Count: 1, Bytes: e.Ret, Files: []string{shortPath(path)}}, true
	}
	return hop{}, false
}

// relKind 는 파일 이름의 번호로 카탈로그와 사용자 객체를 가릅니다. 16384(FirstNormalObjectId) 미만은 시스템 카탈로그입니다.
// 빈 공간 지도(_fsm), 가시성 지도(_vm), 초기화 fork(_init)와 1GB 넘는 조각(.1, .2)은 본체와 같은 번호에 접미사가 붙으므로 접미사를 떼고 봅니다.
func relKind(path string) string {
	name := path[strings.LastIndexByte(path, '/')+1:]
	if i := strings.IndexAny(name, "_."); i >= 0 {
		name = name[:i]
	}
	n, err := strconv.Atoi(name)
	if err != nil {
		return ""
	}
	if n < 16384 {
		return "카탈로그"
	}
	return "테이블·인덱스"
}

// shortPath 는 데이터 디렉터리 안의 경로만 남깁니다.
func shortPath(p string) string {
	if i := strings.Index(p, "/data/"); i >= 0 {
		return p[i+len("/data/"):]
	}
	return p
}

// queryText 는 요청 바이트 안의 SQL을 돌려줍니다. 단순 쿼리는 Q, 확장 프로토콜은 Parse(P) 메시지에 있습니다.
func queryText(b []byte) string {
	for _, m := range parseMessages(b, true) {
		if m.Type == "Q" || m.Type == "P" {
			return m.Text
		}
	}
	return ""
}

// us 는 base부터 ns까지의 시간을 마이크로초로 돌려줍니다.
func us(ns, base uint64) float64 {
	return float64(int64(ns-base)) / 1e3
}

// msgType 은 이동이 실은 첫 메시지의 종류를 돌려줍니다. 메시지가 없으면 빈 문자열입니다.
func msgType(h hop) string {
	if len(h.Msgs) == 0 {
		return ""
	}
	return h.Msgs[0].Type
}

// contains 는 목록에 s가 있는지 돌려줍니다.
func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// stackSampler 는 기록 하나를 처음부터 한 번 따라가며, 요청한 시각의 호출 스택(바깥부터)과 가장 안쪽 칸을 부른 소스 줄을 돌려주는 함수를 만듭니다.
// 스택의 각 칸에는 그 함수 안에서 다음 칸을 부른 줄을 붙이고, 가장 안쪽 칸에는 함수가 정의된 줄을 붙입니다.
// 시각은 늘어나는 순서로 물어야 합니다. atEnter가 참이면 그 시각에 들어온 함수까지 스택에 넣습니다.
// longjmp로 출구가 빠진 칸은 buildTree와 같은 프레임 주소 규칙으로 닫습니다.
func stackSampler(t *traceFile, syms *symbols) func(ns uint64, atEnter bool) ([]frameRef, string) {
	type frame struct{ fn, site, fp uint64 }
	var stack []frame
	i := 0
	advance := func(limit uint64, inclusive bool) {
		for i < len(t.Records) {
			r := t.Records[i]
			if r.Ns > limit || (!inclusive && r.Ns == limit) {
				return
			}
			switch r.Kind {
			case kindEnter:
				for len(stack) > 0 && stack[len(stack)-1].fp <= r.Frame {
					stack = stack[:len(stack)-1]
				}
				stack = append(stack, frame{r.Fn - t.Slide, r.Site - t.Slide, r.Frame})
			case kindExit:
				for len(stack) > 0 && stack[len(stack)-1].fp < r.Frame {
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 0 && stack[len(stack)-1].fn == r.Fn-t.Slide {
					stack = stack[:len(stack)-1]
				}
			}
			i++
		}
	}
	return func(ns uint64, atEnter bool) ([]frameRef, string) {
		advance(ns, atEnter)
		refs := make([]frameRef, len(stack))
		for k, f := range stack {
			refs[k].Fn = syms.name(f.fn)
			if k+1 < len(stack) {
				refs[k].At = syms.lineAt(stack[k+1].site - 4)
			} else if fi := syms.byLow[f.fn]; fi != nil && fi.Line > 0 {
				refs[k].At = fmt.Sprintf("%s:%d", fi.File, fi.Line)
			}
		}
		site := ""
		if len(stack) > 0 {
			site = syms.lineAt(stack[len(stack)-1].site - 4)
		}
		return refs, site
	}
}
