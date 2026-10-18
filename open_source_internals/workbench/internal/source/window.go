// Package source 는 소스 파일에서 화면에 보여 줄 줄 구간을 정하고, 실습 컨테이너나 이미지에서 소스 파일을 읽습니다.
package source

import "strings"

// Range 값은 화면에 보여 줄 줄 구간 하나입니다. From과 To는 1부터 세는 줄 번호이고 둘 다 구간에 들어갑니다.
type Range struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// Window 값은 한 단계에서 보여 줄 소스 구간들과, 그 구간 안에서 찾은 함수 정의 줄입니다.
//
// DefLine은 함수 이름이 줄 맨 앞에서 시작하는 정의 줄입니다. 찾지 못하면 0입니다.
// Ranges는 겹치지 않고 줄 번호 순서입니다. 둘 이상이면 그 사이의 줄은 화면에서 생략됩니다.
type Window struct {
	DefLine int     `json:"defLine"`
	Ranges  []Range `json:"ranges"`
}

const (
	// contextLines는 멈춘 줄 위아래로 함께 보여 줄 줄 수입니다.
	contextLines = 12
	// maxSingle은 함수 머리부터 멈춘 자리까지를 한 구간으로 보여 줄 최대 줄 수입니다. 넘으면 머리와 멈춘 자리를 따로 보여 줍니다.
	maxSingle = 70
	// maxDefSearch는 정의 줄을 찾으려고 멈춘 줄에서 위로 올라갈 최대 줄 수입니다.
	maxDefSearch = 800
	// maxCommentLines는 머리 주석의 시작을 찾으려고 올라갈 최대 줄 수입니다.
	maxCommentLines = 80
)

// Compute 함수는 lines(파일 전체, lines[0]이 1번 줄)에서 function이 line에 멈췄을 때 보여 줄 구간을 정합니다.
//
// 구간은 함수의 머리 주석, 반환형과 정의 줄, 멈춘 줄 위아래 contextLines줄을 담습니다.
// 머리부터 멈춘 자리까지가 maxSingle줄을 넘으면, 머리(주석부터 여는 중괄호까지)와 멈춘 자리 둘로 나눕니다.
// 정의 줄은 PostgreSQL 소스 관례대로 함수 이름이 줄 맨 앞에서 "이름(" 꼴로 시작하는 줄로 찾고, 찾지 못하면 멈춘 자리만 보여 줍니다.
// line이 파일 범위 밖이면 구간 없이 돌려줍니다.
func Compute(lines []string, function string, line int) Window {
	if line < 1 || line > len(lines) {
		return Window{}
	}
	body := Range{From: max(1, line-contextLines), To: min(len(lines), line+contextLines)}

	def := findDef(lines, function, line)
	if def == 0 {
		return Window{Ranges: []Range{body}}
	}
	head := Range{From: headStart(lines, def), To: signatureEnd(lines, def)}

	// 머리와 멈춘 자리가 가까우면 한 구간으로 이어서, 함수 머리부터 멈춘 줄까지 끊김 없이 읽히게 합니다.
	if body.To-head.From+1 <= maxSingle || body.From <= head.To+1 {
		return Window{DefLine: def, Ranges: []Range{{From: head.From, To: max(head.To, body.To)}}}
	}
	return Window{DefLine: def, Ranges: []Range{head, body}}
}

// findDef 함수는 line에서 위로 올라가며 function의 정의 줄을 찾습니다. 찾지 못하면 0을 돌려줍니다.
//
// 호출 스택의 바깥 프레임은 함수 본문 깊은 곳을 가리키므로, 정의는 수백 줄 위에 있을 수 있습니다.
func findDef(lines []string, function string, line int) int {
	prefix := function + "("
	for n := line; n >= 1 && line-n <= maxDefSearch; n-- {
		if strings.HasPrefix(lines[n-1], prefix) {
			return n
		}
	}
	return 0
}

// headStart 함수는 정의 줄 def 위의 반환형 줄과 머리 주석을 거슬러 올라가 함수 머리의 첫 줄을 돌려줍니다.
//
// PostgreSQL은 반환형을 정의 줄 바로 위에 따로 쓰고("static void"), 그 위에 "/* ... */" 머리 주석을 둡니다.
// 반환형이 아닌 줄(빈 줄, 다른 함수의 닫는 중괄호)을 만나면 거기서 멈춥니다.
func headStart(lines []string, def int) int {
	start := def
	// 반환형과 속성 줄은 길어야 몇 줄입니다.
	for n := def - 1; n >= 1 && def-n <= 3; n-- {
		t := strings.TrimSpace(lines[n-1])
		if t == "" || t == "}" || strings.HasSuffix(t, "*/") || strings.HasSuffix(t, ";") {
			break
		}
		start = n
	}
	above := start - 1
	if above < 1 || !strings.HasSuffix(strings.TrimSpace(lines[above-1]), "*/") {
		return start
	}
	for n := above; n >= 1 && above-n <= maxCommentLines; n-- {
		if strings.HasPrefix(strings.TrimSpace(lines[n-1]), "/*") {
			return n
		}
	}
	return start
}

// signatureEnd 함수는 정의 줄 def부터 함수 본문을 여는 "{" 줄까지를 머리로 보고 그 줄 번호를 돌려줍니다.
//
// 인자 목록이 여러 줄에 걸칠 수 있으므로 몇 줄 아래까지 찾고, 찾지 못하면 정의 줄 자체를 돌려줍니다.
func signatureEnd(lines []string, def int) int {
	for n := def; n <= len(lines) && n-def <= 10; n++ {
		if strings.TrimSpace(lines[n-1]) == "{" {
			return n
		}
	}
	return def
}
