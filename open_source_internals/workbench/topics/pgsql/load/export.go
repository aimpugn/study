package load

import (
	"sort"

	"study/open_source_internals/workbench/internal/atlas"
)

// 이 파일은 적재기마다의 내보내기(Export)가 함께 쓰는 run 색인입니다.

// LinkIndex 함수는 run의 연결을 종류마다 색인합니다. out은 출발 사건 → 종류 → 도착 사건 목록, in은 도착 사건 → 종류 → 출발 사건 목록입니다.
//
// 목록은 상대 사건의 id 차례입니다. 형식은 사건의 차례(id)에 뜻을 두지만 연결 표의 차례에는 두지 않으므로(설계 메모 1), 연결 표를
// 섞은 run에서도 내보내기가 같은 원본을 되살리게 합니다(Spring 코드 검토 지적 8). 사건 id 차례가 원본 목록의 차례와 다른 연결이 있으면
// 그 적재기의 되돌림 대조가 어긋남을 드러냅니다.
func LinkIndex(r *atlas.Run) (out, in map[int]map[string][]int) {
	out, in = map[int]map[string][]int{}, map[int]map[string][]int{}
	for _, l := range r.Links {
		if out[l.Src] == nil {
			out[l.Src] = map[string][]int{}
		}
		if in[l.Dst] == nil {
			in[l.Dst] = map[string][]int{}
		}
		out[l.Src][l.Kind] = append(out[l.Src][l.Kind], l.Dst)
		in[l.Dst][l.Kind] = append(in[l.Dst][l.Kind], l.Src)
	}
	for _, idx := range []map[int]map[string][]int{out, in} {
		for _, kinds := range idx {
			for _, ids := range kinds {
				sort.Ints(ids)
			}
		}
	}
	return out, in
}
