package sub1

import "fmt"

func Test(p private) {
	fmt.Println("Print private", p)
}

type private struct {
	internal uint8
}

var (
	ExportedPrivate0 = private{0}
	ExportedPrivate1 = private{1}
)
