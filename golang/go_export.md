# Export

## 완전 private하게 enum 선언

```go
// ./sub1/private.go

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
```

```go
// ./
package main

import (
    "github.com/aimpugn/snippets/golang/playground/sub1"
)

func main() {
    tmp := sub1.ExportedPrivate0
    // var tmp2 sub1.private
    sub1.Test(tmp)
}

```
