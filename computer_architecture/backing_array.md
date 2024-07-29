# Backing array

- [Backing array](#backing-array)
    - [Backing Array("기본 배열" 또는 "기반 배열")의 개념](#backing-array기본-배열-또는-기반-배열의-개념)
    - [여러 언어에서의 사용](#여러-언어에서의-사용)
    - [예시](#예시)
        - [Golang](#golang)

## Backing Array("기본 배열" 또는 "기반 배열")의 개념

**Backing Array**는 동적 배열, 리스트, 슬라이스 등의 자료 구조가 실제 데이터를 저장하기 위해 사용하는 기본 배열을 의미합니다.
이 배열은 고정된 크기를 가지며, 동적 자료 구조는 이 배열을 기반으로 요소를 추가하거나 삭제하면서 동적으로 크기를 조정합니다.

**동적 자료 구조**는 종종 고정된 크기의 배열을 사용하여 데이터를 저장하고, 필요에 따라 새로운 배열을 할당하여 데이터를 복사하고 확장합니다.
이때 기존의 배열이 바로 backing array입니다.

## 여러 언어에서의 사용

- **Java**:

    Java의 `ArrayList`는 내부적으로 `Object[]` 배열을 사용하여 데이터를 저장합니다.
    이 `Object[]` 배열이 바로 `ArrayList`의 backing array입니다.

- **Python**:

    Python의 리스트도 내부적으로 배열을 사용하여 데이터를 저장하며, 이 배열이 Python 리스트의 backing array입니다.

- **C++**:

    `std::vector`도 내부적으로 배열을 사용하며, 이 배열이 벡터의 backing array입니다.

- **Go**

    Go에서는 슬라이스가 backing array를 참조하며, 이 용어가 Go 문서에서 자주 사용됩니다.
    슬라이스는 이 backing array의 일부를 참조함으로써 유연한 배열 사용이 가능하도록 합니다.

## 예시

### Golang

```go
func BackingArray() {
    // 배열 생성 (이것이 backing array가 됨)
    original := [5]int{1, 2, 3, 4, 5}

    // 슬라이스 생성 (original 배열의 일부를 참조)
    slice1 := original[1:4]
    fmt.Printf("slice1: %v, len: %d, cap: %d\n", slice1, len(slice1), cap(slice1))

    // 새로운 슬라이스 생성 (같은 backing array를 공유)
    slice2 := original[2:5]
    fmt.Printf("slice2: %v, len: %d, cap: %d\n", slice2, len(slice2), cap(slice2))

    // slice1 수정 (backing array도 수정됨)
    slice1[1] = 20
    fmt.Printf("After modifying slice1:\n")
    fmt.Printf("original: %v\n", original)
    fmt.Printf("slice1: %v\n", slice1)
    fmt.Printf("slice2: %v\n", slice2)

    // slice1에 요소 추가 (새로운 backing array 생성)
    slice1 = append(slice1, 30, 40)
    fmt.Printf("\nAfter appending to slice1:\n")
    fmt.Printf("original: %v\n", original)
    fmt.Printf("slice1: %v, len: %d, cap: %d\n", slice1, len(slice1), cap(slice1))
    fmt.Printf("slice2: %v\n", slice2)

    // Output:
    //
    // slice1: [2 3 4], len: 3, cap: 4
    // slice2: [3 4 5], len: 3, cap: 3
    // After modifying slice1:
    // original: [1 2 20 4 5]
    // slice1: [2 20 4]
    // slice2: [20 4 5]
    //
    // After appending to slice1:
    // original: [1 2 20 4 5]
    // slice1: [2 20 4 30 40], len: 5, cap: 8
    // slice2: [20 4 5]
}
```

1. `original` 배열이 생성되며, 이것이 초기 backing array가 됩니다.
2. `slice1`과 `slice2`는 같은 backing array(`original`)를 참조하지만, 다른 부분을 가리킵니다.
3. `slice1`을 수정하면 backing array가 수정되어, `original`과 `slice2`에도 영향을 미칩니다.
4. `slice1`에 새 요소를 추가할 때 용량을 초과하면, 새로운 더 큰 backing array가 생성되고 데이터가 복사됩니다. 이때 `slice1`은 새 backing array를 참조하게 되어, `original`과 `slice2`와는 별개가 됩니다.
