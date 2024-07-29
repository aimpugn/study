# Go bench

## 벤치마크?

벤치마크는 코드 세그먼트를 여러 번 실행하고 각 출력을 표준과 비교하여 코드의 전반적인 성능 수준을 평가하는 일종의 함수.

Golang에는 `testing` 패키지와 `go` tool에 벤치마크 작성을 위한 내장 도구들이 있으므로 종속성을 설치하지 않고도 유용한 벤치마크를 작성할 수 있다.

## 벤치마크 위한 올바른 조건 설정

벤치마킹이 유용하려면 실행할 때마다 결과가 일관되고 유사해야 하며, 그렇지 않으면 테스트 중인 코드의 실제 성능을 측정하기 어렵다.

벤치마킹 결과는 벤치마크가 실행되는 컴퓨터의 상태에 따라 크게 영향을 받을 수 있다. 테스트 결과에 영향을 미쳐 부정확하고 불안정한 결과를 초래할 수 있다.
- 전원 관리
- 백그라운드 프로세스
- 열 관리 등

따라서 환경에 미치는 영향을 최대한 최소화해야 한다.
가능하면 아무것도 실행되지 않는 물리적 컴퓨터나 원격 서버를 사용하여 벤치마크를 수행해야 한다.

그러나 그런 별도의 컴퓨터나 서버가 없는 경우 벤치마크를 실행하기 전에 가능한 한 많은 프로그램을 종료하여 다른 프로세스가 벤치마크 결과에 미치는 영향을 최소화해야 한다.

또한 보다 안정적인 결과를 얻으려면 측정값을 기록하기 전에 벤치마크를 여러 번 실행하여 시스템이 충분히 예열되었는지 확인해야 한다.

마지막으로, 네트워크 요청을 모의하는 등 벤치마킹하는 코드를 프로그램의 나머지 부분과 분리하는 것이 중요하다.

## 벤치마크 작성하기

```go
// main.go
package main

func primeNumbers(max int) []int {
    var primes []int

    for i := 2; i < max; i++ {
        isPrime := true

        for j := 2; j <= int(math.Sqrt(float64(i))); j++ {
            if i%j == 0 {
                isPrime = false
                break
            }
        }

        if isPrime {
            primes = append(primes, i)
        }
    }

    return primes
}
```

```go
// main_test.go
package main

import (
    "testing"
)

var num = 1000

func BenchmarkPrimeNumbers(b *testing.B) {
    for i := 0; i < b.N; i++ {
        primeNumbers(num)
    }
}
```

### 벤치마크 실행하기

```bash
$ go test -bench=. -count 5 -run=^#
goos: linux
goarch: amd64
pkg: github.com/ayoisaiah/random
cpu: Intel(R) Core(TM) i7-7560U CPU @ 2.40GHz
BenchmarkPrimeNumbers-4            14485             82484 ns/op
BenchmarkPrimeNumbers-4            14557             82456 ns/op
BenchmarkPrimeNumbers-4            14520             82702 ns/op
BenchmarkPrimeNumbers-4            14407             87850 ns/op
BenchmarkPrimeNumbers-4            14446             82525 ns/op
PASS
ok      github.com/ayoisaiah/random     10.259s
```

- `-count 5`: 벤치마크가 일관된 결과를 생성하는지 확인하기 위해 5번 벤치마크 실행
- `-run=^#`:
    - `-run`: 어떤 테스트가 실행되어야 하는지 지정
    - `^#`: 유닛테스트 함수 제외
- `goos`: OS
- `goarch`: 아키텍처
- `pkg`: 패키지
- `cpu`: CPU
- `BenchmarkPrimeNumbers-4`
    - `BenchmarkPrimeNumbers`: 벤치마크 함수명
    - `-4`: 벤치마크 실행에 사용된 CPU(`GOMAXPROCS`로 지정)
- `14588`: 수행된 총 반복 수
- `82798 ns/op`: 각 반복이 완료되는 데 소요된 평균 시간(nanoseconds)

### 다양한 입력으로 벤치마크하기

코드를 벤치마킹할 때는 다양한 입력이 주어졌을 때 함수가 어떻게 동작하는지 테스트하는 것이 필수적이다.
Go에서 단위 테스트를 작성할 때 일반적으로 사용되는 테이블 기반 테스트 패턴을 활용하여 다양한 입력을 지정할 수 있다.
다음으로 `b.Run()` 메서드를 사용하여 각 입력에 대한 하위 벤치마크를 만들 수 있다.

```go
var table = []struct { // 테이블 기반 테스트 패턴
    input int
}{
    {input: 100},
    {input: 1000},
    {input: 74382},
    {input: 382399},
}

func BenchmarkPrimeNumbers(b *testing.B) {
    for _, v := range table {
        b.Run(fmt.Sprintf("input_size_%d", v.input), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                primeNumbers(v.input)
            }
        })
    }
}
```

## [하위 테스트와 하위 벤치마크](https://go.dev/blog/subtests)

### 하위 테스트

```go
func TestTime(t *testing.T) {
    testCases := []struct {
        gmt  string
        loc  string
        want string
    }{
        {"12:31", "Europe/Zuri", "13:31"},
        {"12:31", "America/New_York", "7:31"},
        {"08:08", "Australia/Sydney", "18:08"},
    }
    for _, tc := range testCases {
        t.Run(fmt.Sprintf("%s in %s", tc.gmt, tc.loc), func(t *testing.T) {
            loc, err := time.LoadLocation(tc.loc)
            if err != nil {
                t.Fatal("could not load location")
            }
            gmt, _ := time.Parse("15:04", tc.gmt)
            if got := gmt.In(loc).Format("15:04"); got != tc.want {
                t.Errorf("got %s; want %s", got, tc.want)
            }
        })
    }
}
```

특정 테스트만 실행할 수도 있다

```bash
# 유럽의 타임존 사용하는 테스트만 실행
$ go test -run=TestTime/"in Europe"
--- FAIL: TestTime (0.00s)
    --- FAIL: TestTime/12:31_in_Europe/Zuri (0.00s)
        time_test.go:85: could not load location

# 정오 이후의 시간대에 대한 테스트만 실행
$ go test -run=Time/12:[0-9] -v
=== RUN   TestTime
=== RUN   TestTime/12:31_in_Europe/Zuri
=== RUN   TestTime/12:31_in_America/New_York
--- FAIL: TestTime (0.00s)
    --- FAIL: TestTime/12:31_in_Europe/Zuri (0.00s)
        time_test.go:85: could not load location
    --- FAIL: TestTime/12:31_in_America/New_York (0.00s)
        time_test.go:89: got 07:31; want 7:31

# 반면 `-run=TestTime/New_York`는 어떤 테스트에도 매치되지 않는다. 
# `/`가 구분자로 사용되므로 아래와 같이 사용해야 한다.
$ go test -run=Time//New_York
--- FAIL: TestTime (0.00s)
    --- FAIL: TestTime/12:31_in_America/New_York (0.00s)
        time_test.go:88: got 07:31; want 7:31
```

### 병렬 처리 제어

## 벤치마크 테스트 진행 방법

1. **별도의 함수로 구현**: 가장 일반적인 방법은 각 구현을 별도의 함수로 만들고, 각각에 대해 벤치마크 테스트를 작성하는 것입니다. 예를 들어, `function1`과 `function2`라는 두 가지 구현이 있다면, 각각에 대해 별도의 벤치마크 함수를 만들어서 테스트합니다.

2. **조건부 컴파일 (빌드 태그 사용)**: 조건부 컴파일을 사용하여 다른 구현을 포함할 수 있습니다. 이 방법은 빌드 태그를 사용하여 서로 다른 파일에 있는 같은 함수의 다른 구현을 선택적으로 컴파일하게 할 수 있습니다. 예를 들어, `implementation1.go`와 `implementation2.go` 파일이 있고 각 파일에 같은 함수의 다른 구현이 있다면, 빌드 태그를 사용하여 벤치마크할 구현을 선택할 수 있습니다.

3. **인터페이스 사용**: 공통 인터페이스를 정의하고, 각 구현이 이 인터페이스를 만족하도록 구현할 수 있습니다. 그런 다음 벤치마크 테스트에서 이 인터페이스를 통해 각 구현을 테스트할 수 있습니다. 이 방법은 다형성을 활용하여 더 유연하게 구현을 비교할 수 있게 합니다.

### 코드 보관 방법

성능이 떨어지는 코드에 대한 처리 방법은 상황에 따라 다릅니다. 일반적으로 고려할 수 있는 몇 가지 방법은 다음과 같습니다.

- **성능이 떨어지는 구현 삭제**: 성능이 현저히 떨어지고, 특별한 이유(예: 가독성이 더 좋음, 추후 최적화 가능성 등)가 없다면, 더 나은 성능의 구현을 선택하고 나머지를 삭제하는 것이 일반적입니다.
  
- **버전 관리 시스템에 보관**: 벤치마크 테스트 결과와 함께 모든 구현을 버전 관리 시스템에 보관할 수 있습니다. 이 방법은 추후 참고할 목적으로, 혹은 나중에 다른 컨텍스트에서 성능이 더 우수할 수 있는 구현을 재검토하기 위해 유용합니다.

- **문서화**: 각 구현의 특성, 장단점, 벤치마크 결과를 문서화하는 것도 좋은 방법입니다. 이는 결정 과정을 투명하게 하고, 나중에 이러한 결정을 재고할 때 도움이 될 수 있습니다.

결론적으로, 벤치마크한 두 구현을 모두 버전 관리할지 여부는 프로젝트의 요구 사항, 팀의 정책, 구현의 가치에 따라 결정됩니다. 벤치마크 결과를 기반으로 의사결정 과정을 문서화하고, 이해관계자와 공유하여 투명한 결정을 내리는 것이 중요합니다.
