# prealloc

- [prealloc](#prealloc)
    - [`prealloc`?](#prealloc-1)
    - [슬라이스 사전 할당(Pre-allocation)의 장점](#슬라이스-사전-할당pre-allocation의-장점)
        - [메모리 재할당 방지](#메모리-재할당-방지)
        - [성능 향상](#성능-향상)
        - [예측 가능한 메모리 사용](#예측-가능한-메모리-사용)
    - [작동 방식](#작동-방식)
    - [`prealloc` 명령어 사용법](#prealloc-명령어-사용법)
    - [슬라이스 미리 할당하기](#슬라이스-미리-할당하기)
        - [**`make` 함수 사용**](#make-함수-사용)
        - [**슬라이스 리터럴 사용**](#슬라이스-리터럴-사용)
    - [주의사항](#주의사항)
    - [경고 예시](#경고-예시)
        - [`Consider pre-allocating 'ret' (prealloc)`](#consider-pre-allocating-ret-prealloc)
        - [작동 원리](#작동-원리)
        - [구현 예시](#구현-예시)
        - [성능 고려사항](#성능-고려사항)
        - [실제 적용시 주의점](#실제-적용시-주의점)
        - [최적화 기법](#최적화-기법)
        - [결론](#결론)

## `prealloc`?

`prealloc`은 Go 언어에서 *슬라이스 선언 시 성능 최적화를 위해 미리 할당할 수 있는 가능성*을 찾아내는 정적 분석 도구입니다.

Go의 슬라이스는 동적 배열로, 내부적으로 다음 세 가지 요소로 구성됩니다:
1. 포인터: 기본 배열을 가리키는 포인터
2. 길이(len): 현재 슬라이스에 포함된 요소의 수
3. 용량(cap): 기본 배열의 전체 크기

슬라이스에 새 요소를 추가할 때 용량이 부족하면, Go 런타임은 더 큰 배열을 할당하고 기존 데이터를 복사합니다. 이 과정은 비용이 많이 듭니다.

이 도구는 특히 반복문 내에서 슬라이스에 요소를 추가하는 패턴을 분석하여, 슬라이스의 용량을 미리 설정할 수 있는 기회를 제공함으로써 메모리 재할당을 줄이고 성능을 향상시킬 수 있는지를 검사합니다

## 슬라이스 사전 할당(Pre-allocation)의 장점

### 메모리 재할당 방지

슬라이스의 내부 구조는 다음과 같습니다:

1. 포인터: 실제 데이터를 저장하는 배열을 가리킵니다.
2. 길이(length): 현재 슬라이스에 포함된 요소의 수입니다.
3. 용량(capacity): 내부 배열의 총 크기입니다.

슬라이스에 새로운 요소를 추가할 때, 현재 용량을 초과하면 Go 런타임은 다음과 같은 작업을 수행합니다:

1. 새로운, 더 큰 배열을 할당합니다.
2. 기존 데이터를 새 배열로 복사합니다.
3. 슬라이스의 내부 포인터를 새 배열로 업데이트합니다.

이 과정은 계산 비용이 많이 들고, 메모리 사용량을 일시적으로 증가시킵니다.

```go
// 사전 할당을 하지 않은 경우
func withoutPreallocation() []int {
    var numbers []int
    for i := 0; i < 10000; i++ {
        numbers = append(numbers, i)
    }
    return numbers
}

// 사전 할당을 한 경우
func withPreallocation() []int {
    numbers := make([]int, 0, 10000)
    for i := 0; i < 10000; i++ {
        numbers = append(numbers, i)
    }
    return numbers
}
```

`withoutPreallocation` 함수에서는 슬라이스가 여러 번 재할당됩니다.
Go의 성장 전략에 따라, 용량이 불충분할 때마다 대략 2배씩 증가합니다.

예를 들어, 용량이 0, 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384로 증가할 수 있습니다.

반면 `withPreallocation` 함수에서는 처음부터 10000개의 요소를 저장할 수 있는 용량을 가진 슬라이스를 생성합니다.
따라서 재할당이 전혀 발생하지 않습니다.

### 성능 향상

슬라이스 재할당은 세 가지 주요 비용을 수반합니다:

1. 메모리 할당: 새로운 배열을 위한 메모리 할당
2. 데이터 복사: 기존 데이터를 새 배열로 복사
3. 가비지 컬렉션: 이전 배열이 더 이상 사용되지 않을 때 가비지 컬렉션의 대상이 됨

이러한 작업들은 CPU 시간을 소비하며, 특히 슬라이스 크기가 클수록 그 비용이 커집니다.

```go
package main

import (
    "fmt"
    "testing"
)

func BenchmarkWithoutPreallocation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var s []int
        for j := 0; j < 10000; j++ {
            s = append(s, j)
        }
    }
}

func BenchmarkWithPreallocation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 10000)
        for j := 0; j < 10000; j++ {
            s = append(s, j)
        }
    }
}

func main() {
    result := testing.Benchmark(BenchmarkWithoutPreallocation)
    fmt.Printf("Without Preallocation: %d ns/op\n", result.NsPerOp())

    result = testing.Benchmark(BenchmarkWithPreallocation)
    fmt.Printf("With Preallocation: %d ns/op\n", result.NsPerOp())
}
```

이 벤치마크 테스트를 실행하면, 사전 할당을 사용한 버전이 그렇지 않은 버전보다 훨씬 빠른 것을 확인할 수 있습니다. 예를 들어:

```sh
Without Preallocation: 98765 ns/op
With Preallocation: 54321 ns/op
```

이러한 결과는 시스템에 따라 다를 수 있지만, 일반적으로 사전 할당 버전이 상당히 더 빠릅니다.

### 예측 가능한 메모리 사용

사전 할당을 통해 프로그램의 메모리 사용을 더 정확하게 예측하고 제어할 수 있습니다.
동적 재할당은 일시적으로 메모리 사용량을 증가시키고, 이전 배열이 가비지 컬렉션될 때까지 추가 메모리를 점유합니다.

```go
package main

import (
    "fmt"
    "runtime"
)

func memoryUsage() uint64 {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return m.Alloc
}

func main() {
    before := memoryUsage()

    // 사전 할당 없이 슬라이스 생성
    var s1 []int
    for i := 0; i < 1000000; i++ {
        s1 = append(s1, i)
    }

    middle := memoryUsage()

    // 사전 할당으로 슬라이스 생성
    s2 := make([]int, 0, 1000000)
    for i := 0; i < 1000000; i++ {
        s2 = append(s2, i)
    }

    after := memoryUsage()

    fmt.Printf("Memory without preallocation: %d bytes\n", middle-before)
    fmt.Printf("Memory with preallocation: %d bytes\n", after-middle)
}
```

이 예제를 실행하면 사전 할당을 사용한 경우가 그렇지 않은 경우보다 메모리 사용량이 더 적고 일정한 것을 볼 수 있습니다.
예를 들어:

```sh
Memory without preallocation: 9876543 bytes
Memory with preallocation: 8765432 bytes
```

사전 할당 없이 슬라이스를 생성할 때는 여러 번의 재할당으로 인해 추가적인 메모리 오버헤드가 발생합니다. 반면, 사전 할당을 사용하면 필요한 메모리를 한 번에 할당하므로 더 효율적입니다.

## 작동 방식

`prealloc`은 주로 `for` 또는 `range` 반복문을 사용하여 슬라이스에 요소를 추가하는 코드 패턴을 분석합니다.
만약 반복문의 반복 횟수를 예측할 수 있다면, `prealloc`은 해당 슬라이스의 초기 용량을 미리 할당하도록 제안합니다. 이는 다음과 같은 형태의 코드에서 유용합니다:

```go
var to []string
for _, t := range s.To {
    to = append(to, t.String())
}
```

위 코드에서 `to` 슬라이스는 반복문을 통해 동적으로 요소가 추가되고 있습니다.
`prealloc`은 이러한 패턴을 감지하고, 가능하다면 `make` 함수를 사용하여 초기 용량을 할당하도록 제안합니다

## `prealloc` 명령어 사용법

`prealloc` 도구를 사용하는 방법은 다음과 같습니다:

```bash
prealloc go/src/....
```

이 명령은 지정된 디렉토리 내의 Go 소스 파일을 분석하고, 슬라이스의 미리 할당을 고려할 수 있는 위치를 보고합니다

## 슬라이스 미리 할당하기

### **`make` 함수 사용**

`make` 함수를 사용하여 슬라이스의 초기 용량을 설정할 수 있습니다.
이 방법은 슬라이스에 저장될 요소의 최대 개수를 알고 있을 때 유용합니다.

```go
var ret = make([]MyType, 0, expectedSize)
```

여기서 `expectedSize`는 슬라이스에 저장될 요소의 예상 개수입니다.
이렇게 하면 슬라이스 `ret`는 길이가 0이고 용량이 `expectedSize`인 상태로 시작합니다.

```go
func collectData() []MyType {
    expectedSize := 100 // 예상되는 요소의 개수
    ret := make([]MyType, 0, expectedSize) // 용량을 미리 할당

    for i := 0; i < expectedSize; i++ {
        element := MyType{ /* 데이터 초기화 */ }
        ret = append(ret, element)
    }
    return ret
}
```

이 코드에서 `ret` 슬라이스는 `make` 함수를 사용하여 미리 할당되었습니다.
`expectedSize`만큼의 용량을 미리 설정함으로써, 슬라이스에 요소를 추가할 때 재할당이 발생하지 않습니다.
이는 메모리 사용을 최적화하고 성능을 향상시키는 데 도움이 됩니다.

### **슬라이스 리터럴 사용**

슬라이스 리터럴을 사용하여 초기 요소를 포함한 슬라이스를 생성할 수 있습니다.
이 방법은 초기 요소의 개수가 명확할 때 사용할 수 있습니다.

```go
var ret = []MyType{ /* 초기 요소들 */ }
```

## 주의사항

`prealloc` 사용 시 주의해야 할 점은, 모든 경우에 미리 할당이 최적의 선택은 아닐 수 있다는 것입니다.
특히, 반복 횟수가 매우 적거나 예측할 수 없는 경우에는 미리 할당이 오히려 메모리 낭비를 초래할 수 있습니다.
따라서, `prealloc`의 제안을 적용하기 전에 성능 프로파일링을 수행하는 것이 좋습니다

## 경고 예시

### `Consider pre-allocating 'ret' (prealloc)`

Go 코드에서 슬라이스를 사용할 때 발생합니다.
이 경고는 *슬라이스가 루프나 다른 반복적인 데이터 추가 과정에서 여러 번 재할당될 수 있음을 지적*하며, 이로 인해 성능 저하가 발생할 수 있습니다.
슬라이스의 용량(capacity)이 초과되어 재할당되는 것을 방지하기 위해, 가능한 경우 슬라이스를 미리 할당(pre-allocate)하는 것이 좋습니다.

---

네, 질문하신 내용이 정확합니다. "pre-allocating `list`" 또는 "prealloc"이라는 표현은 `make` 함수를 사용하여 슬라이스를 미리 할당하라는 의미입니다. 이 최적화 기법에 대해 자세히 설명해 드리겠습니다.

### 작동 원리

1. `make` 함수를 사용하여 슬라이스를 생성합니다.
2. 예상되는 최종 크기를 지정하여 기본 배열을 미리 할당합니다.
3. 이렇게 하면 슬라이스가 확장될 때 새로운 메모리 할당과 데이터 복사를 피할 수 있습니다.

### 구현 예시

```go
// 사전 할당을 하지 않은 경우
var list []string
for i := 0; i < 10000; i++ {
    list = append(list, fmt.Sprintf("item %d", i))
}

// 사전 할당을 한 경우
list := make([]string, 0, 10000)
for i := 0; i < 10000; i++ {
    list = append(list, fmt.Sprintf("item %d", i))
}
```

이 예시에서 두 번째 방법은 `make([]string, 0, 10000)`을 사용하여 용량이 10,000인 슬라이스를 미리 할당합니다. 이렇게 하면 `append` 작업 중에 재할당이 발생하지 않습니다.

### 성능 고려사항

1. 메모리 사용: 사전 할당은 초기에 더 많은 메모리를 사용하지만, 재할당 횟수를 줄여 전체적인 메모리 사용을 최적화할 수 있습니다.
2. 실행 시간: 재할당과 데이터 복사를 피함으로써 실행 시간을 단축할 수 있습니다.
3. 예측 가능성: 최종 크기를 정확히 알 수 있는 경우에 가장 효과적입니다.

### 실제 적용시 주의점

1. 과대 추정: 필요 이상으로 큰 용량을 할당하면 메모리 낭비가 발생할 수 있습니다.
2. 과소 추정: 너무 작게 추정하면 여전히 재할당이 필요할 수 있습니다.
3. 동적 크기: 최종 크기를 예측할 수 없는 경우, 점진적으로 용량을 늘리는 전략이 더 적합할 수 있습니다.

### 최적화 기법

1. 점진적 확장: 정확한 크기를 모를 때는 예상 크기의 일정 비율(예: 25% 더 크게)로 할당하고, 필요시 확장합니다.
2. 벤치마킹: 다양한 사전 할당 크기로 벤치마크 테스트를 수행하여 최적의 초기 용량을 결정합니다.

```go
// 점진적 확장 예시
initialCap := 100
list := make([]string, 0, initialCap)
for i := 0; ; i++ {
    if i >= cap(list) {
        newCap := cap(list) * 2
        newList := make([]string, len(list), newCap)
        copy(newList, list)
        list = newList
    }
    list = append(list, fmt.Sprintf("item %d", i))
    // 종료 조건 체크
}
```

이 최적화 기법은 특히 대규모 데이터 처리, 고성능 애플리케이션, 메모리 사용이 중요한 환경에서 유용합니다. 하지만 항상 실제 사용 사례와 데이터 특성을 고려하여 적용해야 합니다.

---

네, 슬라이스를 미리 할당하는 것의 장점과 이론적 근거에 대해 자세히 설명드리겠습니다.

### 결론

슬라이스 사전 할당은 메모리 효율성, 실행 속도, 그리고 예측 가능한 리소스 사용 측면에서 상당한 이점을 제공합니다. 특히 대규모 데이터를 다루거나 성능이 중요한 애플리케이션에서 이 기법은 매우 유용합니다. 그러나 슬라이스의 최종 크기를 정확히 예측할 수 없는 경우에는 과도한 메모리 할당을 피하기 위해 점진적 확장 전략을 고려해야 합니다.
