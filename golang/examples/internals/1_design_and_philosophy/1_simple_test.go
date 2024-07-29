package design_and_philosophy

// Go 언어의 컴파일러 디자인은
// 1. 빠른 컴파일 시간과
// 2. 실행 효율성을 강조

// Go의 패키지 시스템은 종속성을 최소화하며, 각 패키지는 컴파일될 때 필요한 최소한의 종속성만 포함 -> 전체 프로그램의 컴파일 시간을 단축
import (
	"testing"
)

func TestSimple(t *testing.T) {
	Simple()
}
