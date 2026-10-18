package atlas

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// Order 값은 두 사건의 시작 순서를 견준 결과입니다.
type Order int

// 견준 결과입니다.
const (
	// Incomparable 은 순서를 낼 수 없음입니다(시각 없음, 시계 관계 모름, 차이가 관계 범위 안).
	Incomparable Order = iota
	// Before 는 앞 사건이 먼저 시작함입니다.
	Before
	// Same 은 같은 시계에서 같은 시각에 시작함입니다.
	Same
	// After 는 앞 사건이 나중에 시작함입니다.
	After
)

// CompareStart 함수는 run 안의 사건 a와 b의 시작 순서를 견줍니다. 순서를 낼 수 없으면 Incomparable과 그 까닭을 돌려줍니다.
//
// 같은 시계면 시각으로 바로 견줍니다. 시계가 다르면 a의 시계에서 b의 시계로 가는 관계(차이 범위)가 있어야 하고,
// 두 시각의 차이가 그 범위 밖일 때만 순서를 냅니다. 범위 안이면 어느 쪽이 먼저인지 기록이 말해 주지 않습니다.
// 관계는 한 단계만 따릅니다(a→b 또는 b→a). 해석기가 이미 한 투영을 되돌리지 않습니다.
func CompareStart(r *Run, a, b Event) (Order, string) {
	if a.StartNS == nil || b.StartNS == nil {
		return Incomparable, "시작 시각이 없는 사건"
	}
	ca, cb := clockOf(r, a), clockOf(r, b)
	ta, tb := *a.StartNS, *b.StartNS
	if ca == cb {
		return order(ta, tb), ""
	}

	// 시계가 다르면 b의 시각을 a의 시계로 옮긴 범위를 셉니다. a = b + d (d는 a 시계 − b 시계)
	lo, hi, ok := offset(r, ca, cb)
	if !ok {
		return Incomparable, fmt.Sprintf("시계 %q와 %q의 관계를 모름", ca, cb)
	}
	bLo, bHi := tb+lo, tb+hi
	switch {
	case ta < bLo:
		return Before, ""
	case ta > bHi:
		return After, ""
	}
	return Incomparable, fmt.Sprintf("차이가 시계 관계 범위(%d~%d ns) 안", lo, hi)
}

// clockOf 함수는 사건의 출처가 쓰는 시계 id를 돌려줍니다. 출처가 시계를 적지 않았으면 빈 문자열입니다.
func clockOf(r *Run, e Event) string {
	s, _ := r.SourceByID(e.Source)
	return s.Clock
}

// Offset 함수는 시계 a의 시각에서 시계 b의 시각을 뺀 차이의 범위 [lo, hi]를 돌려줍니다. 같은 시계면 [0, 0]이고, 관계를 모르면 세 번째 값이 false입니다.
//
// 화면이 다른 시계의 사건을 '같은 시간대에 있었을 수 있음'으로 견줄 때 씁니다(T4 고친 규칙 5).
func Offset(r *Run, a, b string) (int64, int64, bool) {
	if a == b && a != "" {
		return 0, 0, true
	}
	return offset(r, a, b)
}

// offset 함수는 시계 a의 시각에서 시계 b의 시각을 뺀 차이의 범위를 돌려줍니다. 관계를 모르면 세 번째 값이 false입니다.
func offset(r *Run, a, b string) (int64, int64, bool) {
	// a가 b를 기준으로 적힌 관계면 그대로 씁니다.
	if c, ok := r.ClockByID(a); ok && c.Relation != nil && c.Relation.To == b && c.Relation.Known {
		return c.Relation.LoNS, c.Relation.HiNS, true
	}
	// b가 a를 기준으로 적힌 관계면 부호를 뒤집습니다(b − a ∈ [lo, hi] 이면 a − b ∈ [−hi, −lo]).
	if c, ok := r.ClockByID(b); ok && c.Relation != nil && c.Relation.To == a && c.Relation.Known {
		return -c.Relation.HiNS, -c.Relation.LoNS, true
	}
	return 0, 0, false
}

// order 함수는 같은 시계의 두 시각을 견줍니다.
func order(a, b int64) Order {
	switch {
	case a < b:
		return Before
	case a > b:
		return After
	}
	return Same
}

// MicrosToNanos 함수는 원본의 마이크로초 수를 정수 나노초로 옮기고, 되돌렸을 때 원본과 같은 값이 되는지 확인합니다.
//
// 되돌린 값(ns / 1000)이 원본 float64와 다르면 오류를 돌려줍니다. 나노초보다 촘촘한 값이 섞여 있다는 뜻이라
// 형식이 그 값을 잃게 되므로 적재를 멈춰야 합니다.
func MicrosToNanos(n json.Number) (int64, error) {
	f, err := strconv.ParseFloat(string(n), 64)
	if err != nil {
		return 0, fmt.Errorf("마이크로초 %q: %w", n, err)
	}
	ns := int64(math.Round(f * 1000))
	if float64(ns)/1000 != f {
		return 0, fmt.Errorf("마이크로초 %q를 나노초 %d로 옮기면 되돌린 값이 다름", n, ns)
	}
	return ns, nil
}

// NanosToMicros 함수는 MicrosToNanos의 반대입니다. 원본과 같은 float64가 됩니다.
func NanosToMicros(ns int64) float64 { return float64(ns) / 1000 }
