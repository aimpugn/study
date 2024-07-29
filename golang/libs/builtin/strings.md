# strings

## `ReplaceAll`

### vs for loop

Golang에서 문자열이 충분히 큰 경우, 문자열에서 하이픈을 제거할 때 `strings.ReplaceAll`을 사용하는 것이 직접 루프를 돌면서 문자를 제거하는 방법보다 더 빠릅니다. 그 이유는 `strings.ReplaceAll` 함수가 더 효율적으로 메모리 할당 및 문자열 처리를 하기 때문입니다.

직접 루프를 돌면서 문자를 제거하는 방법은 각 반복마다 새로운 문자열을 생성해야 하기 때문에 메모리 할당과 복사 과정에서 성능이 저하될 수 있습니다.

1. **루프를 사용한 문자 제거**:

   각 문자를 하나씩 확인하면서, 조건에 맞는 문자만 결과 문자열에 추가하는 방식입니다.
   이 방법은 문자열을 반복하면서 매번 새로운 문자열을 생성하게 되어, 메모리 할당 및 복사가 빈번히 발생합니다.

   시간 복잡도는 O(n)입니다만, 각 반복마다 새로운 문자열을 생성하므로 성능 저하가 발생할 수 있습니다.

2. **strings.ReplaceAll 사용**:

   내부적으로 최적화된 알고리즘을 사용하여 문자열을 한 번에 처리합니다.
   *메모리 할당을 최소화하고, 내부적으로 더 효율적인 방법으로 문자열을 처리*합니다.

   시간 복잡도는 O(n)으로 동일하지만, 실질적인 수행 속도는 더 빠를 수 있습니다.

아래 코드는 두 가지 방법을 비교하는 벤치마크 테스트입니다.

```go
package main

import (
    "strings"
    "testing"
)

func removeHyphensLoop(uuidWithHyphens string) string {
    result := ""
    for _, char := range uuidWithHyphens {
        if char != '-' {
            result += string(char)
        }
    }
    return result
}

func removeHyphensReplaceAll(uuidWithHyphens string) string {
    return strings.ReplaceAll(uuidWithHyphens, "-", "")
}

func BenchmarkRemoveHyphensLoop(b *testing.B) {
    uuid := "123e4567-e89b-12d3-a456-426614174000"
    for i := 0; i < b.N; i++ {
        removeHyphensLoop(uuid)
    }
}

func BenchmarkRemoveHyphensReplaceAll(b *testing.B) {
    uuid := "123e4567-e89b-12d3-a456-426614174000"
    for i := 0; i < b.N; i++ {
        removeHyphensReplaceAll(uuid)
    }
}
```

### 벤치마크 실행

위 벤치마크를 실행하면, `strings.ReplaceAll`이 `for` 루프를 사용한 방법보다 더 빠르게 동작하는 것을 확인할 수 있습니다.

```bash
❯ go test -bench=.
goos: darwin
goarch: arm64
pkg: github.com/aimpugn/snippets/golang/examples/builtin
BenchmarkRemoveHyphensLoop-10            1116459              1047 ns/op
BenchmarkRemoveHyphensReplaceAll-10     16348495                75.18 ns/op
PASS
ok      github.com/aimpugn/snippets/golang/examples/builtin     3.494s
```
