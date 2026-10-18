package mapview

import "sort"

// Cell 값은 화면의 한 칸이 가로로 덮는 구간입니다(픽셀이든 칸 번호든 같은 단위).
type Cell struct {
	X0 float64 `json:"x0"`
	X1 float64 `json:"x1"`
}

// Normalize 함수는 드래그한 가로 구간 [x0, x1]을 칸의 경계에 맞춘 범위 [first, last]로 바꿉니다(「범위 해설의 규칙」 1).
//
// 구간이 조금이라도 덮은 칸은 모두 넣습니다(바깥쪽으로 맞춤). 클릭(x0 == x1)은 그 점을 품은 칸 하나입니다.
// 경로 밖으로 흐리게 보이는 칸은 넣지 않으므로, 고를 수 있는 칸 번호를 [lo, hi]로 받습니다. 덮은 칸이 없으면 ok가 거짓입니다.
// 「×N」 칸은 칸 하나라 통째로 들어갑니다.
func Normalize(cells []Cell, x0, x1 float64, lo, hi int) (first, last int, ok bool) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	first, last = -1, -1
	for i := max(lo, 0); i <= hi && i < len(cells); i++ {
		c := cells[i]
		if c.X1 < x0 || c.X0 > x1 {
			continue
		}
		if first < 0 {
			first = i
		}
		last = i
	}
	return first, last, first >= 0
}

// 두 범위의 관계입니다(고친 규칙 8).
const (
	// RelSame 은 줌과 잎 집합이 같은 같은 해설 자리입니다.
	RelSame = "same"
	// RelSameLeaves 는 잎 집합은 같고 줌만 다른 범위입니다(목록에서 한 묶음).
	RelSameLeaves = "same_leaves"
	// RelContains 는 앞 범위의 잎 집합이 뒤 범위의 잎 집합을 진부분집합으로 품는 것입니다.
	RelContains = "contains"
	// RelInside 는 앞 범위가 뒤 범위 안에 드는 것입니다.
	RelInside = "inside"
	// RelOverlap 은 일부만 겹치는 것입니다.
	RelOverlap = "overlap"
	// RelApart 는 겹치지 않는 것입니다.
	RelApart = "apart"
)

// Range 값은 정규화한 범위 하나입니다. Leaves는 범위의 칸들을 펼친 잎 사건 id의 합집합(정렬)입니다.
type Range struct {
	Zoom   string `json:"zoom"`
	First  int    `json:"first"`
	Last   int    `json:"last"`
	Leaves []int  `json:"leaves"`
}

// RangeLeaves 함수는 칸마다의 잎 집합에서 [first, last] 칸들의 합집합을 정렬해 돌려줍니다.
func RangeLeaves(cellLeaves [][]int, first, last int) []int {
	m := map[int]bool{}
	for i := first; i <= last && i < len(cellLeaves); i++ {
		for _, id := range cellLeaves[i] {
			m[id] = true
		}
	}
	out := make([]int, 0, len(m))
	for id := range m {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

// Relate 함수는 범위 a가 범위 b에 대해 어떤 관계인지 돌려줍니다. 판정은 잎 집합으로 하고, 같은 자리만 줌까지 같아야 합니다.
func Relate(a, b Range) string {
	in := map[int]bool{}
	for _, id := range b.Leaves {
		in[id] = true
	}
	common := 0
	for _, id := range a.Leaves {
		if in[id] {
			common++
		}
	}
	switch {
	case common == len(a.Leaves) && common == len(b.Leaves):
		if a.Zoom == b.Zoom {
			return RelSame
		}
		return RelSameLeaves
	case common == len(b.Leaves):
		return RelContains
	case common == len(a.Leaves):
		return RelInside
	case common > 0:
		return RelOverlap
	}
	return RelApart
}
