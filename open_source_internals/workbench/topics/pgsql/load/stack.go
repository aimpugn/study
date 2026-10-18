package load

import (
	"fmt"

	"study/open_source_internals/workbench/internal/atlas"
)

// 이 파일은 함수 기록기가 남긴 PostgreSQL 백엔드의 호출 스택을 장소로 옮기는 규칙입니다. 트랜잭션 실험과 Spring 기록이 같은
// 기록기와 해석기(decode/scenario.go의 decodeRequest)를 거쳐 스택 칸이 같은 꼴({fn, at}, 바깥 함수부터)이므로 한 곳에 둡니다.
// 위치를 남기지 않은 함수(줄 모름 칸, 줄 판정 안에서 부른 함수)는 같은 프로세스의 다른 스택 칸에서 그 이름이 나온 경로를 찾아,
// 같은 함수가 위치 있는 장소와 없는 장소 둘로 갈리지 않게 합니다(설계 메모 5의 3).

// NotePath 메서드는 프로세스 장소 proc의 스택 칸 하나(fn, at)에서 함수 fn이 at의 파일에 있다고 적어 둡니다.
// at을 읽지 못하면(줄 모름 칸 포함) 적지 않습니다. Callee가 이 장부로 위치 없는 함수의 경로를 찾으므로, 스택을 옮기기 전에 그
// 프로세스의 스택 칸을 모두 적어 둬야 합니다.
func (l *Loader) NotePath(proc int, fn, at string) {
	path, _, err := ParseAt(at)
	if err != nil {
		return
	}
	if l.paths[proc] == nil {
		l.paths[proc] = map[string]map[string]bool{}
	}
	if l.paths[proc][fn] == nil {
		l.paths[proc][fn] = map[string]bool{}
	}
	l.paths[proc][fn][path] = true
}

// Callee 메서드는 위치를 남기지 않은 함수 fn의 함수 장소를 프로세스 장소 proc 아래에 돌려줍니다.
//
// 경로는 NotePath가 그 프로세스에서 적은 것이 하나일 때만 붙이고, 없으면 위치 없는 함수입니다. 둘 이상이면 어느 것인지 가를 수
// 없어 오류를 돌려줍니다.
func (l *Loader) Callee(proc int, fn, src, group string) (int, error) {
	path := ""
	switch ps := l.paths[proc][fn]; len(ps) {
	case 0:
	case 1:
		for p := range ps {
			path = p
		}
	default:
		return 0, fmt.Errorf("프로세스 장소 %d의 함수 %s가 경로 %d개에 나옴", proc, fn, len(ps))
	}
	return l.Function(proc, fn, path, src, group), nil
}

// Stack 메서드는 원본 스택(바깥 함수부터 {fn, at})을 프로세스 장소 proc 아래 프레임 장소 목록으로 뒤집어 돌려줍니다.
// 0번이 가장 안쪽 프레임(사건의 장소)입니다. group은 스택 칸의 경로 무리 앞부분(예: "$.server.events[]")이고, where는 오류에 쓸 원본 경로입니다.
//
// 스택이 없으면(null) 둘 다 비고, 빈 목록이나 꼴이 다른 칸은 없음과 가를 수 없어 오류입니다. 줄을 모르는 칸(UnknownAt)은 Callee로
// 찾은 함수 아래 위치 없는 프레임입니다.
func (l *Loader) Stack(proc int, v any, src, group, where string) ([]int, error) {
	if v == nil {
		return nil, nil
	}
	list, ok := v.([]any)
	if !ok || len(list) == 0 {
		return nil, fmt.Errorf("%s.stack이 비지 않은 목록이 아님", where)
	}
	out := make([]int, 0, len(list))
	for k := len(list) - 1; k >= 0; k-- {
		f, err := Object(list[k], where+".stack[]")
		if err != nil {
			return nil, err
		}
		if err := Keys(f, where+".stack[]", "fn", "at"); err != nil {
			return nil, err
		}
		fn, err := Text(f["fn"], where+".stack[].fn")
		if err != nil {
			return nil, err
		}
		at, err := Text(f["at"], where+".stack[].at")
		if err != nil {
			return nil, err
		}
		// 줄을 모르는 칸은 같은 프로세스의 스택에서 그 함수가 나온 경로 아래에 둡니다.
		var fr int
		if at == UnknownAt {
			function, err := l.Callee(proc, fn, src, group+".stack[].fn")
			if err != nil {
				return nil, fmt.Errorf("%s: %w", where, err)
			}
			fr = l.UnknownFrame(function, fn, src, group+".stack[].at")
		} else if fr, err = l.Frame(proc, fn, at, src, group+".stack[].fn", group+".stack[].at"); err != nil {
			return nil, fmt.Errorf("%s: %w", where, err)
		}
		out = append(out, fr)
	}
	return out, nil
}

// StackOf 함수는 Stack이 만든 프레임 장소 목록(0번이 가장 안쪽)을 원본의 꼴(바깥 함수부터 {fn, at})로 되돌립니다.
// 위치 없는 프레임의 at은 UnknownAt입니다.
func StackOf(places []atlas.Place, stack []int) []any {
	out := make([]any, 0, len(stack))
	for k := len(stack) - 1; k >= 0; k-- {
		f := places[stack[k]]
		out = append(out, map[string]any{"fn": f.Name, "at": LocText(f.Loc)})
	}
	return out
}
