package journey

import (
	"fmt"
	"regexp"

	"study/open_source_internals/workbench/topics/pgsql/load"
)

// ReadConcept4 함수는 concept-4.html에 박힌 QUERIES literal(질의 11개의 목록)을 읽습니다.
//
// literal은 "const QUERIES = " 뒤의 한 줄이고 JSON과 같은 꼴입니다. 줄을 찾지 못하거나 JSON으로 읽지 못하면 오류를 돌려줍니다.
func ReadConcept4(html string) ([]any, error) {
	lit, err := load.Literal(html, "QUERIES")
	if err != nil {
		return nil, err
	}
	v, err := load.Decode(lit)
	if err != nil {
		return nil, err
	}
	list, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("QUERIES가 목록이 아님")
	}
	return list, nil
}

// topKey 와 nextKey 는 concept-3의 DATA literal에서 따옴표 없이 적힌 맨 위 열쇠 둘(update, select)을 찾습니다.
var (
	topKey  = regexp.MustCompile(`^\{(\w+):`)
	nextKey = regexp.MustCompile(`\},\s*(\w+):\s*\{"endUs"`)
)

// ReadConcept3 함수는 concept-3.html에 박힌 DATA literal(질의 이름 → 해석 결과)을 읽습니다.
//
// 맨 위 열쇠 둘이 JavaScript 꼴(따옴표 없음)이라, 재고 조사의 extract.py와 같은 규칙으로 따옴표를 붙여 읽습니다.
// 그 밖의 곳이 JSON이 아니면 오류를 돌려줍니다.
func ReadConcept3(html string) (map[string]any, error) {
	lit, err := load.Literal(html, "DATA")
	if err != nil {
		return nil, err
	}
	lit = topKey.ReplaceAllString(lit, `{"$1":`)
	lit = nextKey.ReplaceAllString(lit, `},"$1":{"endUs"`)
	v, err := load.Decode(lit)
	if err != nil {
		return nil, err
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("DATA가 객체가 아님")
	}
	return m, nil
}
