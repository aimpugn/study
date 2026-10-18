// Package trace 는 실습의 scripts/trace.py가 gdb 안에서 쓴 줄 기록을 단계 목록으로 읽습니다.
//
// 기록 형식의 정본은 trace.py의 머리 주석입니다. 이 패키지는 그 형식을 읽기만 하고 새로 정하지 않습니다.
//
//	STOP <순번> <함수> <소스 상대경로>:<줄> [식=값 ...]
//	BT   <깊이> <함수> <소스 상대경로>:<줄>
//	END  <함수>=<횟수> ...
//	NOBREAK <함수>
//
// run-trace.sh는 기록 뒤에 "--- psql" 줄과 psql의 출력을 붙이고, 저장된 기록 파일에는 "#"로 시작하는 설명 줄이 있을 수 있습니다.
package trace

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Location 값은 소스 트리 안의 한 줄을 가리킵니다. 디버그 정보가 없어 위치를 모르면 Path가 빈 문자열이고 Line이 0입니다.
type Location struct {
	Path string `json:"path"`
	Line int    `json:"line"`
}

// Value 하나는 멈춘 순간 gdb가 평가한 식과 그 결과 문자열입니다.
type Value struct {
	Expr string `json:"expr"`
	Text string `json:"text"`
}

// Frame 하나는 호출 스택의 한 칸입니다. Depth 0이 멈춘 함수 자신이고, 숫자가 클수록 바깥 호출자입니다.
type Frame struct {
	Depth    int      `json:"depth"`
	Function string   `json:"function"`
	Loc      Location `json:"loc"`
}

// Step 하나는 중단점 하나에 멈춘 기록입니다.
//
// Values는 감시 목록의 식 순서로 값을 나눌 수 있었을 때만 채워집니다. 나눌 수 없으면 Values는 비고 RawValues에 기록 그대로의 문자열이 남습니다.
// Stack은 trace.py가 이 단계 바로 뒤에 BT 줄을 남긴 경우에만 채워집니다.
type Step struct {
	Seq       int      `json:"seq"`
	Function  string   `json:"function"`
	Loc       Location `json:"loc"`
	Values    []Value  `json:"values,omitempty"`
	RawValues string   `json:"rawValues,omitempty"`
	Stack     []Frame  `json:"stack,omitempty"`
}

// Count 하나는 END 줄에 적힌 함수별 멈춘 횟수입니다.
type Count struct {
	Function string `json:"function"`
	Hits     int    `json:"hits"`
}

// Record 하나는 실행 한 번의 기록 전체입니다.
//
// Counts는 END 줄의 순서를 그대로 지킵니다. trace.py는 한 번도 멈추지 않은 함수를 END에 쓰지 않으므로, 목록에 없는 함수는 0회입니다.
// Output은 "--- psql" 아래의 psql 출력 줄입니다. 캐시를 데우는 실행과 추적한 실행의 결과가 차례로 들어 있습니다.
type Record struct {
	Steps   []Step   `json:"steps"`
	Counts  []Count  `json:"counts"`
	NoBreak []string `json:"noBreak,omitempty"`
	Output  []string `json:"output,omitempty"`
}

// Exprs 값은 함수 이름마다 감시 목록에 적힌 식을 적힌 순서대로 담습니다. Parse는 이것으로 STOP 줄의 값 부분을 나눕니다.
type Exprs map[string][]string

// ParseWatch 함수는 감시 목록 파일("함수<TAB>gdb 식" 줄)을 trace.py와 같은 규칙으로 읽습니다.
//
// 빈 줄과 "#"로 시작하는 줄은 건너뛰고, 같은 함수가 여러 줄에 나오면 식을 차례로 모읍니다.
// 반환한 Exprs에는 식이 없는 함수도 빈 목록으로 들어 있습니다. 읽기 오류가 나면 그 오류를 돌려줍니다.
func ParseWatch(r io.Reader) (Exprs, error) {
	exprs := Exprs{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fn, expr, _ := strings.Cut(line, "\t")
		if expr == "" {
			if _, ok := exprs[fn]; !ok {
				exprs[fn] = nil
			}
			continue
		}
		exprs[fn] = append(exprs[fn], expr)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("감시 목록 읽기: %w", err)
	}
	return exprs, nil
}

// Parse 함수는 trace.py와 run-trace.sh가 쓴 출력을 Record로 읽습니다.
//
// exprs는 그 실행에 쓴 감시 목록입니다. nil이면 값을 나누지 않고 RawValues에만 남깁니다.
// 형식에 맞지 않는 줄이 있으면 몇 번째 줄인지 밝힌 오류를 돌려줍니다. 기록 형식이 바뀌었는데 조용히 일부만 읽는 일을 막기 위해서입니다.
func Parse(r io.Reader, exprs Exprs) (Record, error) {
	rec := Record{}
	sc := bufio.NewScanner(r)
	// 긴 SQL이 query_string 값에 그대로 들어가므로 기본 줄 길이 한도(64KB)보다 넉넉히 둡니다.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	inOutput := false
	n := 0
	for sc.Scan() {
		n++
		line := sc.Text()

		// "--- psql" 아래는 기록이 아니라 psql의 출력이므로 그대로 모읍니다.
		if inOutput {
			rec.Output = append(rec.Output, line)
			continue
		}
		if line == "--- psql" {
			inOutput = true
			continue
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		kind, rest, _ := strings.Cut(line, " ")
		var err error
		switch kind {
		case "STOP":
			err = parseStop(&rec, rest, exprs)
		case "BT":
			err = parseFrame(&rec, rest)
		case "END":
			err = parseEnd(&rec, rest)
		case "NOBREAK":
			rec.NoBreak = append(rec.NoBreak, rest)
		default:
			err = fmt.Errorf("알 수 없는 기록 종류 %q", kind)
		}
		if err != nil {
			return Record{}, fmt.Errorf("%d번째 줄: %w", n, err)
		}
	}
	if err := sc.Err(); err != nil {
		return Record{}, fmt.Errorf("기록 읽기: %w", err)
	}
	return rec, nil
}

// parseStop 함수는 "STOP" 뒤의 "<순번> <함수> <위치> [값]"을 읽어 rec.Steps에 더합니다.
//
// 순번이 1부터 하나씩 늘지 않으면 오류를 돌려줍니다. 순번이 빠졌다는 것은 기록 일부가 잘렸다는 뜻이기 때문입니다.
func parseStop(rec *Record, rest string, exprs Exprs) error {
	fields := strings.SplitN(rest, " ", 4)
	if len(fields) < 3 {
		return fmt.Errorf("STOP 줄에 순번·함수·위치가 모두 있어야 합니다: %q", rest)
	}
	seq, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("STOP 순번이 숫자가 아닙니다: %q", fields[0])
	}
	if want := len(rec.Steps) + 1; seq != want {
		return fmt.Errorf("STOP 순번 %d가 와야 할 자리에 %d가 왔습니다", want, seq)
	}
	loc, err := parseLocation(fields[2])
	if err != nil {
		return err
	}

	step := Step{Seq: seq, Function: fields[1], Loc: loc}
	if len(fields) == 4 {
		step.Values, step.RawValues = splitValues(fields[3], exprs[fields[1]])
	}
	rec.Steps = append(rec.Steps, step)
	return nil
}

// parseFrame 함수는 "BT" 뒤의 "<깊이> <함수> <위치>"를 읽어 바로 앞 단계의 Stack에 더합니다.
//
// BT 줄은 trace.py가 STOP 줄 바로 다음에 깊이 0부터 차례로 쓰므로, 앞 단계가 없거나 깊이가 건너뛰면 오류입니다.
func parseFrame(rec *Record, rest string) error {
	fields := strings.Fields(rest)
	if len(fields) != 3 {
		return fmt.Errorf("BT 줄은 깊이·함수·위치 세 칸이어야 합니다: %q", rest)
	}
	if len(rec.Steps) == 0 {
		return fmt.Errorf("STOP 없이 BT 줄이 나왔습니다")
	}
	depth, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("BT 깊이가 숫자가 아닙니다: %q", fields[0])
	}
	last := &rec.Steps[len(rec.Steps)-1]
	if depth != len(last.Stack) {
		return fmt.Errorf("BT 깊이 %d가 와야 할 자리에 %d가 왔습니다", len(last.Stack), depth)
	}
	loc, err := parseLocation(fields[2])
	if err != nil {
		return err
	}
	last.Stack = append(last.Stack, Frame{Depth: depth, Function: fields[1], Loc: loc})
	return nil
}

// parseEnd 함수는 "END" 뒤의 "<함수>=<횟수>" 목록을 읽어 rec.Counts에 순서대로 넣습니다.
func parseEnd(rec *Record, rest string) error {
	for _, pair := range strings.Fields(rest) {
		fn, hits, ok := strings.Cut(pair, "=")
		if !ok {
			return fmt.Errorf("END 항목에 '='이 없습니다: %q", pair)
		}
		n, err := strconv.Atoi(hits)
		if err != nil {
			return fmt.Errorf("END 항목 %q의 횟수가 숫자가 아닙니다", pair)
		}
		rec.Counts = append(rec.Counts, Count{Function: fn, Hits: n})
	}
	return nil
}

// parseLocation 함수는 "경로:줄" 또는 위치를 모를 때의 "?"를 읽습니다.
func parseLocation(s string) (Location, error) {
	if s == "?" {
		return Location{}, nil
	}
	i := strings.LastIndexByte(s, ':')
	if i <= 0 {
		return Location{}, fmt.Errorf("위치는 '경로:줄' 형식이어야 합니다: %q", s)
	}
	line, err := strconv.Atoi(s[i+1:])
	if err != nil || line <= 0 {
		return Location{}, fmt.Errorf("위치의 줄 번호가 양의 정수가 아닙니다: %q", s)
	}
	return Location{Path: s[:i], Line: line}, nil
}

// splitValues 함수는 trace.py가 "식=값"을 공백으로 이어 쓴 문자열을 식마다 나눕니다.
//
// 식과 값 모두 공백을 품을 수 있어서(예: `(char *) rel->rd_rel->relname.data=0x... "t"`), 감시 목록의 식 순서를 알 때만 나눕니다.
// 각 식은 "식=" 꼴로 차례로 나타나야 하고, 다음 식의 시작은 " 다음식=" 가운데 처음 나오는 자리로 봅니다.
// 이 규칙에 맞지 않거나 식 목록이 없으면 나누지 않고 원문을 돌려줍니다. 값 안에 " 다음식="이 우연히 들어 있으면 잘못 나눌 수 있으며, 이 한계는 trace.py의 출력 형식에서 옵니다.
func splitValues(raw string, exprs []string) ([]Value, string) {
	if len(exprs) == 0 || !strings.HasPrefix(raw, exprs[0]+"=") {
		return nil, raw
	}
	values := make([]Value, 0, len(exprs))
	pos := len(exprs[0]) + 1
	for i, expr := range exprs {
		end := len(raw)
		if i+1 < len(exprs) {
			next := strings.Index(raw[pos:], " "+exprs[i+1]+"=")
			if next < 0 {
				return nil, raw
			}
			end = pos + next
		}
		values = append(values, Value{Expr: expr, Text: raw[pos:end]})
		if i+1 < len(exprs) {
			pos = end + len(exprs[i+1]) + 2
		}
	}
	return values, ""
}

// gdbString 값은 gdb가 C 문자열 포인터를 출력한 꼴입니다: 주소, 공백, 큰따옴표 문자열, 그리고 gdb가 길이 한도로 자른 경우의 "...".
var gdbString = regexp.MustCompile(`^0x[0-9a-f]+ ("(?:[^"\\]|\\.)*")(\.\.\.)?$`)

// CString 함수는 gdb가 출력한 char * 값(예: `0xaaaa... "SELECT 1;"`)에서 문자열 내용을 꺼냅니다.
//
// truncated는 gdb가 출력 길이 한도(print elements, 기본 200자) 때문에 문자열을 잘랐다는 뜻입니다. 그때 s는 앞부분입니다.
// ok가 false이면 이 꼴이 아닙니다. 예를 들어 gdb가 같은 문자가 반복되는 구간을 `'x' <repeats 15 times>`로 줄여 쓰면 꺼내지 않습니다.
func CString(text string) (s string, truncated, ok bool) {
	m := gdbString.FindStringSubmatch(text)
	if m == nil {
		return "", false, false
	}
	// gdb의 이스케이프(\", \\, \n, \t, 세 자리 8진수)는 Go 문자열 리터럴 규칙과 겹치므로 strconv.Unquote로 풉니다.
	s, err := strconv.Unquote(m[1])
	if err != nil {
		return "", false, false
	}
	return s, m[2] != "", true
}
