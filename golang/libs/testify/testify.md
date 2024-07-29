# testify

## testify?

## `require.EqualValues`와 `require.ElementsMatch`

이 두 함수는 테스트 중에 객체들이 기대하는 조건을 만족하는지 검증하는 데 사용됩니다.

### `require.EqualValues`

> `require.EqualValues`는 타입 변환을 고려한 값의 동등성을 검사합니다.

`require.EqualValues` 함수는 두 객체가 같은 타입으로 변환 가능하거나 같은 값을 가지는지 확인합니다. 이 함수는 타입 변환을 고려하여 두 값이 같은지 비교할 때 유용합니다. 예를 들어, `uint32`와 `int32` 타입의 같은 숫자 값이라도 서로 다른 타입으로 인식되는 문제를 해결할 수 있습니다.

**예제 코드:**

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestEqualValues(t *testing.T) {
    require.EqualValues(t, uint32(123), int32(123), "값이 일치하지 않습니다.")
}
```

이 예제에서는 `uint32(123)`과 `int32(123)`이 같은 값인지를 검사합니다. 타입이 달라도 내부적으로 같은 값을 가지므로 테스트는 통과합니다.

비교하는 대상이 포인터인 경우, 포인터가 가리키는 값의 동등성을 검사합니다. 즉, 두 포인터가 가리키는 실제 값이 같은지를 비교합니다. 포인터 자체의 주소 값이 아니라, 포인터가 가리키는 값의 내용을 비교하는 것입니다. 따라서, 두 포인터가 서로 다른 메모리 주소를 가리키더라도, 그 내용이 같다면 `EqualValues`는 테스트를 통과시킵니다.

**예제 코드:**

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestEqualValuesWithPointers(t *testing.T) {
    a := "hello"
    b := "hello"
    ap := &a
    bp := &b

    require.EqualValues(t, ap, bp) // 값의 내용이 같으므로 테스트 통과
}
```

### `require.ElementsMatch`

> `require.ElementsMatch`는 두 컬렉션의 요소가 순서에 상관없이 동일한지 검사합니다.

`require.ElementsMatch` 함수는 두 컬렉션(배열, 슬라이스 등)이 같은 요소를 가지고 있는지, 순서에 상관없이 확인합니다. 이 함수는 요소의 순서가 다르거나, 같은 요소가 여러 번 등장하는 경우에도 두 컬렉션이 같은 요소를 같은 횟수만큼 포함하고 있는지 검사합니다.

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestElementsMatch(t *testing.T) {
    listA := []int{1, 2, 3, 4}
    listB := []int{4, 3, 2, 1}
    require.ElementsMatch(t, listA, listB, "컬렉션의 요소가 일치하지 않습니다.")
}
```

이 예제에서는 `listA`와 `listB`가 순서와 상관없이 같은 요소를 포함하고 있는지 검사합니다. 두 리스트는 같은 요소를 포함하므로 테스트는 통과합니다.

비교하는 대상이 포인터인 경우, 컬렉션(배열, 슬라이스 등) 내의 요소들이 서로 동등한지를 순서에 상관없이 검사합니다. 포인터를 요소로 포함하는 슬라이스의 경우, `ElementsMatch`는 포인터가 가리키는 값의 동등성을 검사합니다. 즉, 포인터들이 가리키는 값들이 동일한지를 확인하며, 이 값들의 순서는 고려하지 않습니다.

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestElementsMatchWithPointerSlices(t *testing.T) {
    a := "hello"
    b := "world"
    c := "hello"
    d := "world"

    slice1 := []*string{&a, &b}
    slice2 := []*string{&d, &c}

    require.ElementsMatch(t, slice1, slice2) // 포인터가 가리키는 값들이 동일하므로 테스트 통과
}
```

이 예제에서 `slice1`과 `slice2`는 포인터들을 요소로 가지고 있으며, 이 포인터들이 가리키는 값들은 각각 "hello"와 "world"입니다. `ElementsMatch`는 이 값들이 동일한지를 순서에 상관없이 검사하므로, 테스트는 통과합니다.

이렇게 `require.EqualValues`와 `require.ElementsMatch`는 포인터를 포함한 값의 동등성을 검사할 때 유용하게 사용될 수 있습니다.
