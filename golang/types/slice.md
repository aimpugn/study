# slice

- [slice](#slice)
    - [슬라이스(slice)?](#슬라이스slice)
    - [슬라이스의 네 가지 주요 요소](#슬라이스의-네-가지-주요-요소)
    - [반개구간](#반개구간)
    - [슬라이스 초기화 방식](#슬라이스-초기화-방식)
        - [Composite Literal](#composite-literal)
        - [nil 슬라이스](#nil-슬라이스)
        - [`make` 함수를 사용한 초기화](#make-함수를-사용한-초기화)
        - [Sub-slicing](#sub-slicing)
    - [배열과 슬라이스](#배열과-슬라이스)
    - [`dst = dst[:n]`로 슬라이싱 하는 이유](#dst--dstn로-슬라이싱-하는-이유)
    - [`append`](#append)
        - [Backing Array와 슬라이스의 관계](#backing-array와-슬라이스의-관계)
        - [예시를 통한 설명](#예시를-통한-설명)
        - [중요한 점](#중요한-점)
    - [기타](#기타)

## 슬라이스(slice)?

Go에서 슬라이스는 실제로 배열(backing array)의 뷰(view) 또는 배열의 부분집합을 나타내는 자료구조입니다.

> You can see a slice as a “window” into an array

- 슬라이스는 *배열의 일부분을 참조*하는 **가변 길이**의 구조
- 슬라이스는 동적으로 크기가 변할 수 있으며, 내부적으로 배열의 한 부분을 '참조'한다
- 슬라이스는 **배열의 동적인 뷰**로 볼 수 있으며, 배열의 연속된 부분에 대한 참조 정보를 포함한다

## 슬라이스의 네 가지 주요 요소

슬라이스를 구조체로 표현한다면 다음과 같습니다.

```go
type slice struct {
    array unsafe.Pointer // backing array를 가리키는 포인터
    len   int            // 슬라이스의 현재 길이(현재 요소 수)
    cap   int            // backing array의 남은 용량. 현재 위치부터 backing array 끝까지의 요소 수
}
```

- 배열에 대한 참조(Reference to an Array), `포인터`

    슬라이스가 참조하는 배열(backing array)을 가리킵니다.
    Backing array는 슬라이스가 참조하는 실제 메모리 상의 연속된 데이터 저장 공간입니다.
    이 Backing array에 슬라이스의 데이터가 실제로 저장됩니다.

- `오프셋`(Offset)

    슬라이스가 참조하는 배열의 첫 번째 요소의 인덱스입니다.
    이는 슬라이스가 배열의 어느 부분을 참조하는지 정의합니다.

    `window`를 설명하는 데 필요합니다.

- `길이`(Length)

    슬라이스에 포함된 요소의 수입니다.
    이는 슬라이스가 참조하는 배열의 특정 부분의 길이를 나타냅니다.
    길이는 슬라이스에 *직접 접근할 수 있는 요소의 수*를 나타내며, `len` 함수를 사용하여 얻을 수 있습니다.

    `window`를 설명하는 데 필요합니다.

- `용량`(Capacity)

    슬라이스의 시작점(`오프셋`)부터 해당 슬라이스의 backing array 끝까지 가능한 요소의 최대 수를 나타냅니다.
    즉, 용량은 슬라이스가 확장될 수 있는 최대 길이(공간)입니다.

    이는 슬라이스가 참조하는 배열의 끝까지 남아 있는 여유 공간을 의미하며, `cap` 함수를 사용하여 얻을 수 있다.

    요소 추가시 backing array를 관리하는 데 필요합니다.

## [반개구간](../../math/interval.md#반열린-구간-half-open-interval-반개구간)

```go
test := "hello" // 길이 5
fmt.Println(test[0:4]) // hell
fmt.Println(test[0:5]) // hello
```

## 슬라이스 초기화 방식

### Composite Literal

```go
// Composite Literal을 사용한 초기화
slice1 := []int{1,2,3} // len=3, cap=3

// nil이 아닌 슬라이스를 만든다.
// 해당 슬라이스는 길이와 용량이 0이지만, 실제로는 메모리에 할당된 빈 슬라이스를 가리킨다.
// nil 슬라이스는 아니며, 이로 인해 작은 메모리 할당이 발생할 수 있다.
emptyButNotNil := []byte{}
```

### nil 슬라이스

```go
// 아무런 메모리도 할당하지 않으며, 길이와 용량이 모두 0이다.
// 요소를 추가하기 위해서는 
//  - 별도의 초기화 과정(예: cipher = make([]byte, 0)로 용량을 지정
//  - 새로운 요소를 append
//  - 그 외 방법? 등이 필요
var emptyAndNil []byte
```

### `make` 함수를 사용한 초기화

`make` 함수를 사용하여 슬라이스를 초기화할 때 Go 런타임은 내부적으로 새 배열을 생성합니다.
이 배열은 슬라이스에 의해 참조되며, 슬라이스의 용량(`cap`)이 배열의 전체 크기를 결정합니다.
슬라이스의 길이(`len`)는 이 배열 내에서 슬라이스가 즉시 접근하고 사용할 수 있는 부분의 길이를 나타냅니다.

- 슬라이스에 의해 접근 가능한 부분 = from 배열의 첫 요소 to `len` 수만큼의 요소
- 예비 공간 = from `len` to `cap`까지의 구간

슬라이스는 이 구간의 요소들을 직접 읽고 쓸 수 있습니다.
즉, 슬라이스는 배열의 시작부터 `len` 위치까지의 요소들을 **보고** 사용할 수 있습니다.

예비 공간은 현재는 접근할 수 없지만, `append` 함수 등을 통해 슬라이스의 길이를 확장할 때 사용될 수 있는 공간입니다.

1. **배열 생성**:

    `make`는 요청된 용량(`cap`) 크기의 배열을 생성합니다.
    이 배열은 타입의 기본값(예: `int`의 경우 `0`)으로 초기화됩니다.

2. **슬라이스 참조**:

    생성된 배열의 첫 번째 요소부터 `len` 위치까지의 요소를 포함하는 슬라이스를 반환합니다.
    이 부분이 슬라이스가 초기에 접근하고 조작할 수 있는 범위입니다.

3. **접근 가능 범위**:

    슬라이스의 `len`은 슬라이스가 현재 접근할 수 있는 요소들의 수를 나타내며, 이 요소들은 배열의 시작 부분에서부터 연속적으로 위치합니다.

4. **예비 공간**:

    슬라이스의 용량(`cap`)은 내부 배열의 총 크기를 나타내며, 이 중 `len`을 초과하는 부분은 추가 데이터를 저장하기 위한 공간으로 남아 있습니다.

```go
s := make([]int, 5) // 길이와 용량이 5인 슬라이스

// 길이는 5이지만, 추가 요소를 저장할 수 있도록 배열의 실제 길이는 10이 된다.
//  - 용량(cap)이 10인 배열 생성:
//    
//    10 개의 `int` 요소를 저장할 수 있는 공간을 갖고, 각 요소는 초기값 `0`으로 설정된다.
//  
//  - 길이(len)가 5인 슬라이스 반환:
//
//    슬라이스 `t`는 생성된 배열의 첫 번째 요소부터 다섯 번째 요소까지를 포함한다.
//    즉, 슬라이스 `t`는 배열의 첫 5개 요소(인덱스 0, 1, 2, 3, 4)에 접근하고 이를 조작할 수 있다.
// 
t := make([]int, 5, 10) // 길이가 5이고, 용량이 10인 슬라이스
```

### Sub-slicing

기존 슬라이스나 배열에서 하위 슬라이스를 생성할 수 있습니다.

```go
a := []int{1, 2, 3, 4, 5}
// 하위 슬라이스는 원본 슬라이스의 일부를 참조합니다.
s := a[1:3]  // [2 3]: len? 2 cap? 4(a의 전체 길이에서 시작 인덱스를 뺀 값)
```

## 배열과 슬라이스

- Go 언어에서 배열과 슬라이스는 다른 타입
    - 배열은 고정된 크기를 가진 연속된 요소의 집합
    - 슬라이스는 동적으로 크기가 변할 수 있는 배열의 레퍼런스 형태
- 슬라이스는 내부적으로 세 가지 주요 속성을 갖는다
    - 포인터: 배열의 특정 요소를 가리킨다
    - 길이(length): 슬라이스가 포함하는 요소의 수
    - 그리고 용량(capacity):

```go
arr := [5]int{1, 2, 3, 4, 5} // 크기가 5인 정수 배열 선언
s := arr[:] // 배열의 모든 요소를 포함하는 슬라이스 생성
```

- 이 코드에서 `arr`는 정확히 5개의 정수를 가진 배열이며,
- `s`는 `arr` 배열의 모든 요소를 가리키는 슬라이스
- 여기서 `[:]` 연산은 **배열의 모든 요소를 포함하는 슬라이스를 만들라는 의미**입니다.
- 이 슬라이스는 배열의 뷰(view)와 같다
- 슬라이스를 통해 배열의 요소를 수정하면 **원본 배열도 변경**됩니다. 하지만 슬라이스 자체는 배열과는 다른 타입으로, 다양한 내장 함수와 메서드를 사용할 수 있으며, 슬라이스만을 인자로 받는 함수에 전달할 수 있다

```go
// `append` 함수는 슬라이스에 새 요소를 추가할 때 사용된다
s = append(s, 6) // 슬라이스에 새 요소 추가
```

- 슬라이스는 동적으로 크기가 변할 수 있으며, 배열과 달리 길이가 고정되어 있지 않다
- 따라서, 배열을 슬라이스로 변환하는 것은 다음 같은 경우에 필요
    - Go에서 배열에 내장된 함수나 메서드를 사용하고자 할 때,
    - 또는 슬라이스만을 인자로 받는 함수에 배열을 전달하고자 할 때

## `dst = dst[:n]`로 슬라이싱 하는 이유

```go
func main() {
    str := "SGVsbG8sIHdvcmxkIQ=="

    // `dst` 슬라이스를 생성할 때 `make` 함수로 초기 길이를 설정하는 것이 중요한데, 이는 디코딩될 데이터를 저장할 충분한 공간을 확보하기 위함
    // `dst` 슬라이스는 처음에 `base64.StdEncoding.DecodedLen(len(str))`에 의해 계산된 길이로 생성된다 
    // base64.StdEncoding.DecodedLen(len(str)): str의 길이를 기반으로 디코딩된 데이터의 최대 길이를 계산
    dst := make([]byte, base64.StdEncoding.DecodedLen(len(str)))
    n, err := base64.StdEncoding.Decode(dst, []byte(str))
//  ^ 
//  실제로 디코딩된 바이트의 수를 반환하며, 
//  실제 디코딩된 데이터는 이보다 짧을 수 있다

    if err != nil {
        fmt.Println("decode error:", err)
        return
    }

    // 디코딩된 데이터의 실제 바이트 수에 맞게 `dst` 슬라이스의 길이를 조정
    dst = dst[:n]
    //    ^^^^^^^
    //    슬라이스의 길이를 재조정하여 Base64 디코딩으로 인해 
    //    실제로 채워진 바이트만 포함하도록 하는 것
    fmt.Printf("%q\n", dst)

    //Output:
    //
    //"Hello, world!"
}
```

- Go에서는 슬라이스 길이를 재조정하여 작업의 결과를 정확하게 표현하는 것이 일반적인 패턴
- 함수가 작업의 결과와 함께 실제로 처리된 바이트의 수를 반환할 때 유용하며, 호출자는 이 정보를 사용하여 슬라이스를 적절하게 재조정할 수 있다

## `append`

- ~~adding elements to a slice~~ -> always creates a new slice.
    - new slice = original slice + 추가하고 싶은 요소
    - 원본 슬라이스는 건드리지 않고 그대로 남아 있는다.

---

네, 말씀하신 대로 정확합니다. 제가 설명을 부정확하게 표현한 점 사과드립니다. 귀하의 지적이 매우 정확하며, Go에서 슬라이스의 본질에 대해 더 정확하고 상세하게 설명드리겠습니다.

### Backing Array와 슬라이스의 관계

1. Backing Array:
   - 실제 데이터를 저장하는 연속된 메모리 공간입니다.
   - 고정된 크기를 가집니다.

2. 슬라이스:
   - Backing Array의 일부 또는 전체를 참조합니다.
   - 동적으로 크기를 조절할 수 있는 것처럼 보이지만, 실제로는 새로운 배열을 할당하고 데이터를 복사하는 과정을 통해 이루어집니다.

### 예시를 통한 설명

```go
package main

import "fmt"

func main() {
    // 배열 생성
    arr := [5]int{1, 2, 3, 4, 5}
    
    // 배열을 기반으로 슬라이스 생성
    slice1 := arr[1:4]
    fmt.Printf("slice1: %v, len: %d, cap: %d\n", slice1, len(slice1), cap(slice1))
    
    // 슬라이스를 수정하면 backing array도 수정됨
    slice1[0] = 20
    fmt.Printf("arr after modifying slice1: %v\n", arr)
    
    // 용량을 초과하여 요소 추가
    slice1 = append(slice1, 60, 70)
    fmt.Printf("slice1 after append: %v, len: %d, cap: %d\n", slice1, len(slice1), cap(slice1))
    fmt.Printf("arr after appending to slice1: %v\n", arr)
}
```

이 예시의 출력 결과:

```
slice1: [2 3 4], len: 3, cap: 4
arr after modifying slice1: [1 20 3 4 5]
slice1 after append: [20 3 4 60 70], len: 5, cap: 8
arr after appending to slice1: [1 20 3 4 5]
```

설명:
1. `slice1`은 처음에 `arr`의 일부를 참조합니다.
2. `slice1`을 수정하면 `arr`도 변경됩니다. 이는 같은 backing array를 공유하기 때문입니다.
3. `slice1`에 요소를 추가하여 용량을 초과하면, 새로운 더 큰 배열이 할당되고 데이터가 복사됩니다. 이때 `slice1`은 새로운 backing array를 참조하게 되어, 원래의 `arr`와는 별개가 됩니다.

### 중요한 점

1. 메모리 효율성: 슬라이스는 배열의 뷰로 작동하기 때문에, 큰 배열의 일부만 참조할 때 매우 효율적입니다.

2. 유연성: 배열은 고정 크기인 반면, 슬라이스는 동적으로 크기를 조절할 수 있습니다(새 배열 할당을 통해).

3. 공유와 수정: 여러 슬라이스가 같은 backing array를 공유할 수 있어, 한 슬라이스의 수정이 다른 슬라이스에 영향을 줄 수 있습니다.

4. 용량 증가: 슬라이스의 용량을 초과하여 요소를 추가하면, 새로운 배열이 할당되고 데이터가 복사됩니다. 이는 성능에 영향을 줄 수 있습니다.

결론적으로, Go의 슬라이스는 배열의 뷰로 작동하며, backing array를 통해 실제 데이터에 접근합니다. 이러한 구조는 유연성과 효율성을 동시에 제공하지만, 메모리 관리와 성능 최적화를 위해서는 이러한 내부 동작을 잘 이해하고 있어야 합니다.

## 기타

- [Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [Slices/arrays explained: create, index, slice, iterate](https://yourbasic.org/golang/slices-explained/)
