package trace

import (
	"os"
	"strings"
	"testing"
)

// 실습의 실제 기록을 그대로 읽어, 기록 파일과 Parse 결과의 단계 수·스택·횟수가 같은지 확인합니다.
func TestParseRealSelectRecord(t *testing.T) {
	// Given: 질문 1에서 쓴 실제 기록과 그 실행의 감시 목록
	watch := mustOpen(t, "../../../database/postgresql/scripts/select.watch")
	exprs, err := ParseWatch(watch)
	if err != nil {
		t.Fatal(err)
	}
	raw := mustRead(t, "../../../database/postgresql/evidence/select-where-id-2.trace")

	// When
	rec, err := Parse(strings.NewReader(raw), exprs)
	// Then: STOP 줄 27개, 18번 단계의 BT 29칸, END의 printtup=1, psql 출력 두 줄
	if err != nil {
		t.Fatal(err)
	}
	if got := len(rec.Steps); got != strings.Count(raw, "\nSTOP ")+1 || got != 27 {
		t.Fatalf("단계 수 %d, 기록의 STOP 줄과 같아야 하고 27이어야 합니다", got)
	}
	pin := rec.Steps[17]
	if pin.Function != "PinBufferForBlock" || len(pin.Stack) != 29 || pin.Stack[28].Function != "main" {
		t.Fatalf("18번 단계의 스택이 기록과 다릅니다: %s, %d칸", pin.Function, len(pin.Stack))
	}
	if !hasCount(rec.Counts, "printtup", 1) || !hasCount(rec.Counts, "heap_getnextslot", 4) {
		t.Fatalf("END 횟수가 기록과 다릅니다: %+v", rec.Counts)
	}
	if strings.Join(rec.Output, ",") != "2|b,2|b" {
		t.Fatalf("psql 출력이 기록과 다릅니다: %q", rec.Output)
	}
}

// 식에 공백이 있고 한 함수에 식이 둘인 경우에도 값을 식마다 나누는지 확인합니다.
func TestSplitValuesWithSpacesAndTwoExprs(t *testing.T) {
	// Given: index.watch의 PinBufferForBlock 식 두 개와 실제 기록의 값 부분
	exprs := []string{"(char *) rel->rd_rel->relname.data", "blockNum"}
	raw := `(char *) rel->rd_rel->relname.data=0xffffa632815c "t_id_idx" blockNum=1`

	// When
	values, rest := splitValues(raw, exprs)

	// Then
	if rest != "" || len(values) != 2 {
		t.Fatalf("두 값으로 나뉘어야 합니다: %+v, 남은 원문 %q", values, rest)
	}
	if values[0].Text != `0xffffa632815c "t_id_idx"` || values[1].Text != "1" {
		t.Fatalf("값이 잘못 나뉘었습니다: %+v", values)
	}
}

// 감시 목록과 맞지 않는 값은 추측해서 나누지 않고 원문을 남기는지 확인합니다.
func TestSplitValuesKeepsRawWhenExprsDoNotMatch(t *testing.T) {
	// Given: 기록에는 blockNum 하나뿐인데 감시 목록은 다른 식을 말하는 경우
	raw := "blockNum=0"

	// When
	values, rest := splitValues(raw, []string{"forkNum"})

	// Then
	if values != nil || rest != raw {
		t.Fatalf("나누지 않고 원문을 남겨야 합니다: %+v, %q", values, rest)
	}
}

// 기록이 잘렸거나 형식이 바뀐 경우를 조용히 넘기지 않고 오류로 알리는지 확인합니다.
func TestParseRejectsBrokenRecords(t *testing.T) {
	cases := map[string]string{
		"순번이 건너뜀":    "STOP 1 f a.c:1\nSTOP 3 g a.c:2\n",
		"STOP 없는 BT": "BT 0 f a.c:1\n",
		"BT 깊이 건너뜀":  "STOP 1 f a.c:1\nBT 0 f a.c:1\nBT 2 g a.c:9\n",
		"줄 번호 없음":    "STOP 1 f a.c\n",
		"알 수 없는 종류":  "STEP 1 f a.c:1\n",
		"END 횟수 이상":  "STOP 1 f a.c:1\nEND f=x\n",
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			// When
			_, err := Parse(strings.NewReader(input), nil)

			// Then
			if err == nil {
				t.Fatalf("오류여야 합니다: %q", input)
			}
		})
	}
}

// 디버그 정보가 없는 프레임의 "?" 위치와 "#" 설명 줄을 받아들이는지 확인합니다.
func TestParseAcceptsUnknownLocationAndComments(t *testing.T) {
	// Given
	input := "# 설명 줄\nSTOP 1 f a.c:3\nBT 0 f a.c:3\nBT 1 ?? ?\nEND f=1\n"

	// When
	rec, err := Parse(strings.NewReader(input), nil)
	// Then
	if err != nil {
		t.Fatal(err)
	}
	if got := rec.Steps[0].Stack[1].Loc; got != (Location{}) {
		t.Fatalf("위치를 모르는 프레임은 빈 Location이어야 합니다: %+v", got)
	}
}

// mustOpen 함수는 시험에 쓸 파일을 열고, 시험이 끝나면 닫습니다. 열지 못하면 시험을 멈춥니다.
func mustOpen(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// mustRead 함수는 시험에 쓸 파일 전체를 문자열로 읽습니다. 읽지 못하면 시험을 멈춥니다.
func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// hasCount 함수는 END 횟수 목록에서 fn의 횟수가 hits와 같은지 봅니다. fn이 목록에 없으면 false입니다.
func hasCount(counts []Count, fn string, hits int) bool {
	for _, c := range counts {
		if c.Function == fn {
			return c.Hits == hits
		}
	}
	return false
}

// gdb가 출력한 char * 값에서 문자열을 꺼내고, 잘림과 해석할 수 없는 꼴을 구분하는지 확인합니다.
func TestCString(t *testing.T) {
	cases := []struct {
		name, in, want string
		truncated, ok  bool
	}{
		{"실제 기록", `0xaaaafe32e8e8 "SELECT * FROM t WHERE id = 2;"`, "SELECT * FROM t WHERE id = 2;", false, true},
		{"따옴표와 역슬래시", `0x1 "SELECT 'a\\b', \"x\";"`, `SELECT 'a\b', "x";`, false, true},
		{"8진수 이스케이프(UTF-8 바이트)", `0x1 "SELECT '\352\260\200';"`, "SELECT '가';", false, true},
		{"gdb가 잘라낸 긴 문자열", `0x1 "SELECT 1"...`, "SELECT 1", true, true},
		{"반복 축약", `0x1 'x' <repeats 15 times>`, "", false, false},
		{"포인터가 아님", `42`, "", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// When
			got, truncated, ok := CString(c.in)

			// Then
			if got != c.want || truncated != c.truncated || ok != c.ok {
				t.Fatalf("CString(%q) = %q, %v, %v; %q, %v, %v여야 합니다", c.in, got, truncated, ok, c.want, c.truncated, c.ok)
			}
		})
	}
}
