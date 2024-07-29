# testing

- [testing](#testing)
    - [benchmark](#benchmark)
        - [벤치마크 결과 분석](#벤치마크-결과-분석)
        - [벤치마크 정보 읽기 및 분석 방법](#벤치마크-정보-읽기-및-분석-방법)
    - [`b.ReportAllocs`](#breportallocs)
    - [벤치마크 결과 해석 및 개선](#벤치마크-결과-해석-및-개선)

## benchmark

Go 벤치마크 결과는 각 라인이 다양한 측정치와 메트릭스를 제공하여 함수의 성능을 평가하는 데 사용됩니다.

### 벤치마크 결과 분석

```bash
❯ go test -bench=.
goos: darwin
goarch: arm64
pkg: github.com/aimpugn/a/b/c/d
BenchmarkRemoveHyphensLoop-10            1133557              1032 ns/op             752 B/op         63 allocs/op
BenchmarkRemoveHyphensReplaceAll-10     16412500                72.45 ns/op           32 B/op          1 allocs/op
PASS
ok      github.com/aimpugn/a/b/c/d     3.303s
```

1. **goos: darwin**

    테스트가 실행된 운영 체제를 나타냅니다.
    여기서 `darwin`은 macOS를 의미합니다.

2. **goarch: arm64**

    테스트가 실행된 아키텍처를 나타냅니다.
    여기서 `arm64`는 64비트 ARM 아키텍처를 의미합니다.

3. **pkg: github.com/aimpugn/a/b/c/d**

    벤치마크 테스트가 실행된 패키지의 경로를 나타냅니다.

4. **BenchmarkRemoveHyphensLoop-10**

    `BenchmarkRemoveHyphensLoop` 함수의 벤치마크 이름을 나타냅니다.
    suffix `-10`은 벤치마크 함수가 병렬로 실행된 고루틴의 수를 나타냅니다.
    Go 벤치마크에서 `b.RunParallel`을 사용하여 병렬로 실행하면, 이 숫자가 붙습니다.

    **1133557**:

    벤치마크가 실행된 반복 횟수(operations)를 나타냅니다.
    이 경우, 1,133,557번 반복 실행되었습니다.

    **1032 ns/op**:

    각 반복당 소요된 평균 시간을 나타냅니다.
    이 경우, 각 반복에 평균 1,032 나노초가 소요되었습니다.

    **752 B/op**:

    각 반복당 평균적으로 할당된 바이트 수를 나타냅니다.
    이 경우, 평균 752 바이트가 할당되었습니다.

    **63 allocs/op**:

    각 반복당 평균적으로 발생한 메모리 할당 횟수를 나타냅니다.
    이 경우, 평균 63번의 메모리 할당이 발생했습니다.

5. **BenchmarkRemoveHyphensReplaceAll-10**

    `BenchmarkRemoveHyphensReplaceAll` 함수의 벤치마크 이름을 나타냅니다.
    suffix `-10`은 벤치마크 함수가 병렬로 실행된 고루틴의 수를 나타냅니다.

    **16412500**:

    벤치마크가 실행된 반복 횟수(operations)를 나타냅니다.
    이 경우, 16,412,500번 반복 실행되었습니다.

    **72.45 ns/op**:

    각 반복당 소요된 평균 시간을 나타냅니다.
    이 경우, 각 반복에 평균 72.45 나노초가 소요되었습니다.

    **32 B/op**:

    각 반복당 평균적으로 할당된 바이트 수를 나타냅니다.
    이 경우, 평균 32 바이트가 할당되었습니다.

    **1 allocs/op**:

    각 반복당 평균적으로 발생한 메모리 할당 횟수를 나타냅니다.
    이 경우, 평균 1 번의 메모리 할당이 발생했습니다.

6. **PASS**

    벤치마크 테스트가 성공적으로 완료되었음을 나타냅니다.

7. **ok      github.com/aimpugn/a/b/c/d     3.303s**

    벤치마크가 실행된 패키지와 전체 실행 시간을 나타냅니다.
    이 경우, 전체 실행 시간은 3.303 초입니다.

### 벤치마크 정보 읽기 및 분석 방법

1. **벤치마크 함수 이름과 병렬 수**:

    벤치마크 함수 이름 뒤에 붙은 숫자는 병렬로 실행된 고루틴의 수를 나타냅니다.
    벤치마크 이름이 길 경우, 병렬 수를 확인하여 함수가 병렬로 실행되었는지 확인할 수 있습니다.

2. **반복 횟수**:

    각 벤치마크가 실행된 반복 횟수입니다.
    이 값이 *높을수록 더 많은 샘플을 수집하여 정확한 성능 측정*을 할 수 있습니다.

3. **ns/op (나노초 per operation)**:

    각 반복당 평균 소요 시간을 나타냅니다.
    이 값이 *낮을수록 성능이 더 우수*합니다.
    두 벤치마크를 비교할 때 이 값을 사용하여 상대적인 성능을 평가할 수 있습니다.

4. **B/op (Bytes per operation)**

    각 반복당 평균적으로 할당된 메모리의 바이트 수를 나타냅니다:
    - `BenchmarkRemoveHyphensLoop-10`의 경우, 반복당 평균적으로 `752` 바이트가 할당되었습니다.
    - `BenchmarkRemoveHyphensReplaceAll-10`의 경우, 반복당 평균적으로 `32` 바이트가 할당되었습니다.

    메모리 *할당이 많을수록 메모리 관리 비용이 증가하며, 성능에 부정적인 영향*을 미칠 수 있습니다.

5. **allocs/op (Allocations per operation)**

    각 반복당 평균적으로 발생한 메모리 할당 횟수를 나타냅니다.
    - `BenchmarkRemoveHyphensLoop-10`의 경우, 반복당 평균적으로 `63`번의 메모리 할당이 발생했습니다.
    - `BenchmarkRemoveHyphensReplaceAll-10`의 경우, 반복당 평균적으로 `1`번의 메모리 할당이 발생했습니다.

    메모리 *할당 횟수가 많을수록 메모리 관리 오버헤드가 증가하여 성능에 부정적인 영향*을 미칠 수 있습니다.

6. **전체 실행 시간**:

    벤치마크가 실행된 패키지와 벤치마크 전체 실행 시간을 나타냅니다.
    이 값을 통해 테스트의 전체적인 실행 시간을 확인할 수 있습니다.

이에 따라 앞선 벤치마크 결과를 분석해 보자면:

- `BenchmarkRemoveHyphensLoop-10`
    - `1116459`번 반복되었고, 평균 `1047 ns`가 소요되었습니다.
    - `for` 루프 방식은 매번 새로운 문자열을 생성하기 때문에 많은 메모리 할당이 발생하고, 이로 인해 성능이 저하됩니다.
- `BenchmarkRemoveHyphensReplaceAll-10`
    - `16348495`번 반복되었고, 평균 `75.18 ns`가 소요되었습니다.
    - `strings.ReplaceAll` 방식은 메모리 할당이 적고, 반복당 할당 횟수도 적기 때문에 메모리 관리 오버헤드가 적습니다.
- 이를 통해 `strings.ReplaceAll` 방식이 `for` 루프를 사용한 방식보다 훨씬 더 빠르게 작동함을 알 수 있습니다.

## `b.ReportAllocs`

```go
func BenchmarkRemoveHyphensLoop(b *testing.B) {
    uuid := "123e4567-e89b-12d3-a456-426614174000"
    for i := 0; i < b.N; i++ {
        removeHyphensLoop(uuid)
    }
    b.ReportAllocs()
}
```

이렇게 하면 벤치마크 결과에 메모리 할당 정보도 포함됩니다.

## 벤치마크 결과 해석 및 개선

- 벤치마크 결과를 해석할 때는 각 반복당 소요 시간(ns/op)을 중심으로 성능을 비교합니다.
- 메모리 할당이 빈번하게 발생하는지 확인하려면 `b.ReportAllocs`를 사용하여 메모리 사용 정보를 분석합니다.
- 성능 개선을 위해 알고리즘을 최적화하거나, 보다 효율적인 라이브러리나 함수를 사용하는 방법을 고려할 수 있습니다.
