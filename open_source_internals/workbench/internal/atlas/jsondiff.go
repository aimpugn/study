package atlas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// Diff 함수는 JSON 값 둘(encoding/json이 any로 읽은 꼴)을 견주어 다른 곳의 경로와 까닭을 모두 돌려줍니다.
// 빈 목록이면 두 값이 같습니다. path는 첫 경로 이름입니다(보통 "$").
//
// 숫자는 글자가 아니라 값으로 견줍니다. JSON에서 `56`과 `56.0`은 같은 값이고, 해석기의 직렬화가 소수 부분이 0인
// 실수를 소수점 없이 쓰기 때문입니다. 숫자·문자열·참거짓·null·목록·객체 사이의 타입은 다르면 다르다고 봅니다.
// 그래서 원래 타입이 다른 식별자(연결 번호 1과 "C1", 문자열 xid)가 바뀌면 잡힙니다.
func Diff(path string, a, b any) []string {
	var out []string
	diffInto(&out, path, a, b)
	return out
}

// diffInto 함수는 Diff의 본체입니다. 다른 곳을 out에 더합니다.
func diffInto(out *[]string, path string, a, b any) {
	switch x := a.(type) {
	case nil:
		if b != nil {
			*out = append(*out, fmt.Sprintf("%s: null ≠ %s", path, brief(b)))
		}
	case bool:
		y, ok := b.(bool)
		if !ok || x != y {
			*out = append(*out, fmt.Sprintf("%s: %s ≠ %s", path, brief(a), brief(b)))
		}
	case string:
		y, ok := b.(string)
		if !ok || x != y {
			*out = append(*out, fmt.Sprintf("%s: %s ≠ %s", path, brief(a), brief(b)))
		}
	case json.Number, float64, int, int64:
		fa, okA := number(a)
		fb, okB := number(b)
		if !okA || !okB || fa != fb {
			*out = append(*out, fmt.Sprintf("%s: %s ≠ %s", path, brief(a), brief(b)))
		}
	case []any:
		y, ok := b.([]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: 목록 ≠ %s", path, brief(b)))
			return
		}
		if len(x) != len(y) {
			*out = append(*out, fmt.Sprintf("%s: 길이 %d ≠ %d", path, len(x), len(y)))
		}
		for i := range min(len(x), len(y)) {
			diffInto(out, fmt.Sprintf("%s[%d]", path, i), x[i], y[i])
		}
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: 객체 ≠ %s", path, brief(b)))
			return
		}
		// 키의 합집합을 이름순으로 돌아, 한쪽에만 있는 키도 빠짐없이 적습니다.
		keys := map[string]bool{}
		for k := range x {
			keys[k] = true
		}
		for k := range y {
			keys[k] = true
		}
		names := make([]string, 0, len(keys))
		for k := range keys {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			va, inA := x[k]
			vb, inB := y[k]
			switch {
			case !inA:
				*out = append(*out, fmt.Sprintf("%s.%s: 없음 ≠ %s", path, k, brief(vb)))
			case !inB:
				*out = append(*out, fmt.Sprintf("%s.%s: %s ≠ 없음", path, k, brief(va)))
			default:
				diffInto(out, path+"."+k, va, vb)
			}
		}
	default:
		*out = append(*out, fmt.Sprintf("%s: 견줄 수 없는 타입 %T", path, a))
	}
}

// number 함수는 숫자 꼴의 값을 float64로 바꿉니다. 숫자가 아니면 두 번째 값이 false입니다.
func number(v any) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := strconv.ParseFloat(string(n), 64)
		return f, err == nil
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// brief 함수는 차이 글에 넣을 짧은 값 표현을 만듭니다. 긴 값은 앞부분만 보입니다.
func brief(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%T", v)
	}
	s := string(b)
	if len(s) > 80 {
		s = s[:77] + "..."
	}
	return s
}
