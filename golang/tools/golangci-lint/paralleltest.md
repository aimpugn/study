# paralleltest

- [paralleltest](#paralleltest)
    - [tparallel](#tparallel)
    - [paralleltest](#paralleltest-1)
    - [관계 및 차이점](#관계-및-차이점)
    - [paralleltest 체크 이유](#paralleltest-체크-이유)
    - [`t.Parallel()`의 동작 원리와 장점](#tparallel의-동작-원리와-장점)
    - [반복문 내에서 `t.Parallel()` 호출](#반복문-내에서-tparallel-호출)
    - [비병렬 테스트를 한 고루틴에서 실행하지 않는 이유](#비병렬-테스트를-한-고루틴에서-실행하지-않는-이유)

## tparallel

`tparallel`은 Go 테스트에서 `t.Parallel()` 호출을 강제하는 린터입니다.
이 도구는 테스트 함수 내에서 `t.Parallel()`이 호출되었는지를 검사하고, 호출되지 않았다면 경고를 발생시킵니다.

`tparallel`의 주요 목적은 모든 테스트가 병렬로 실행될 수 있도록 하여 테스트 실행 시간을 최소화하는 것입니다.

## paralleltest

`paralleltest`는 비슷한 목적을 가지고 있지만, 좀 더 세밀한 설정을 제공합니다.
이 린터는 `t.Parallel()` 호출뿐만 아니라, 테스트 함수 내에서의 루프 변수 처리와 같은 추가적인 문제들을 감지할 수 있습니다.

`paralleltest`는 `t.Parallel()` 호출이 누락된 경우 경고를 발생시키고, 필요에 따라 특정 테스트에서 `t.Parallel()` 호출을 요구하지 않도록 설정할 수 있습니다.

예를 들어, `ignore-missing-subtests` 옵션을 사용하여 하위 테스트(subtests)에서 `t.Parallel()` 호출을 요구하지 않도록 설정할 수 있습니다.

## 관계 및 차이점

두 도구 모두 Go 언어의 테스트 병렬 실행을 지원하며, 테스트 실행 시간을 단축시키는 것을 목표로 합니다. 그러나 `paralleltest`는 `tparallel`보다 더 세밀한 설정과 옵션을 제공하여, 테스트의 특정 요구 사항에 맞게 린터의 동작을 조정할 수 있습니다. 예를 들어, `paralleltest`는 특정 테스트에서 `t.Parallel()` 호출을 생략할 수 있도록 하여, 병렬 실행이 적합하지 않은 테스트 케이스를 유연하게 관리할 수 있습니다.

결론적으로, `tparallel`과 `paralleltest`는 비슷한 기능을 제공하지만, `paralleltest`는 사용자가 테스트 병렬 실행을 더 세밀하게 제어할 수 있는 옵션을 제공한다는 점에서 차이가 있습니다. 이러한 차이는 테스트의 복잡성과 특정 요구 사항에 따라 적절한 도구를 선택하는 데 도움이 될 수 있습니다.

## paralleltest 체크 이유

`t.Parallel()`을 모든 테스트에 붙이도록 강제하는 이유는 여러 가지가 있습니다. 이 접근 방식의 주요 목적과 이점을 살펴보겠습니다:

1. 테스트 실행 속도 향상:
   - 병렬 실행을 통해 전체 테스트 스위트의 실행 시간을 크게 단축할 수 있습니다.
   - 특히 I/O 바운드 테스트나 네트워크 요청이 포함된 테스트에서 효과적입니다.

2. 동시성 문제 조기 발견:
   - 테스트를 병렬로 실행하면 동시성 관련 버그나 경쟁 상태(race condition)를 더 쉽게 발견할 수 있습니다.
   - 이는 실제 운영 환경에서 발생할 수 있는 문제를 미리 파악하는 데 도움이 됩니다.

3. 테스트 독립성 보장:
   - 각 테스트가 독립적으로 실행될 수 있어야 한다는 원칙을 강제합니다.
   - 이는 테스트 간 의존성이나 부작용을 방지하여 더 견고한 테스트 코드를 작성하게 합니다.

4. 리소스 활용 최적화:
   - 현대의 다중 코어 프로세서를 효과적으로 활용할 수 있습니다.
   - CPU 사용률을 높여 테스트 실행의 효율성을 증가시킵니다.

5. CI/CD 파이프라인 최적화:
   - 지속적 통합 및 배포 환경에서 테스트 실행 시간을 줄여 전체 파이프라인의 효율성을 높입니다.

6. 확장성:
   - 프로젝트가 성장하고 테스트 케이스가 증가함에 따라, 병렬 실행의 이점이 더욱 커집니다.

7. 좋은 실천 방법 강제:
   - 개발자들이 테스트를 작성할 때 병렬 실행을 고려하도록 유도하여, 더 나은 테스트 설계를 장려합니다.

그러나 이 접근 방식에도 주의해야 할 점이 있습니다:

1. 모든 테스트가 병렬 실행에 적합하지 않을 수 있습니다 (예: 특정 순서로 실행되어야 하는 테스트).
2. 공유 리소스에 접근하는 테스트의 경우 별도의 동기화 메커니즘이 필요할 수 있습니다.
3. 테스트 환경 설정이 더 복잡해질 수 있습니다.

따라서 `paralleltest` 린터를 사용할 때는 프로젝트의 특성과 요구사항을 고려해야 합니다. 대부분의 경우 병렬 테스트는 이점이 크지만, 특정 상황에서는 예외를 허용하거나 린터 규칙을 조정할 필요가 있을 수 있습니다.

## `t.Parallel()`의 동작 원리와 장점

`t.Parallel()`의 기본 원리:
1. 테스트를 병렬 실행 가능한 상태로 표시합니다.
2. 현재 테스트를 일시 중지하고, 다른 병렬 테스트들이 실행될 수 있도록 합니다.
3. 모든 직렬(non-parallel) 테스트가 완료될 때까지 기다립니다.
4. 그 후 다른 병렬 테스트들과 함께 실행됩니다.

```go
package main

import (
    "testing"
    "time"
)

func TestA(t *testing.T) {
    time.Sleep(2 * time.Second)
    t.Log("TestA completed")
}

func TestB(t *testing.T) {
    `t.Parallel()`
    time.Sleep(2 * time.Second)
    t.Log("TestB completed")
}

func TestC(t *testing.T) {
    `t.Parallel()`
    time.Sleep(2 * time.Second)
    t.Log("TestC completed")
}

func TestD(t *testing.T) {
    time.Sleep(2 * time.Second)
    t.Log("TestD completed")
}
```

이 테스트 파일의 실행 순서와 시간을 분석해보겠습니다:

1. 병렬 실행 없이:
    - 총 실행 시간: 약 8초
    - 순서: TestA -> TestB -> TestC -> TestD

2. `t.Parallel()` 사용:
    - 총 실행 시간: 약 6초
    - 실행 순서:
        1. TestA 실행 (2초)
        2. TestB와 TestC는 `t.Parallel()`을 만나면 대기
        3. TestD 실행 (2초)
        4. TestB와 TestC 동시 실행 (2초)

장점:
1. 실행 시간 단축:
   8초에서 6초로 25% 시간 절약.

2. 리소스 활용 개선:
   TestB와 TestC가 동시에 실행되어 CPU 코어를 더 효율적으로 사용.

3. 동시성 문제 발견:
   만약 TestB와 TestC가 공유 리소스에 접근한다면, 병렬 실행으로 인한 경쟁 조건을 더 쉽게 발견할 수 있습니다.

```go
// https://go.dev/src/testing/testing.go
// 
// Parallel signals that this test is to be run in parallel with (and only with)
// other parallel tests. When a test is run multiple times due to use of
// -test.count or -test.cpu, multiple instances of a single test never run in
// parallel with each other.
func (t *T) Parallel() {
    if t.isParallel {
        panic("testing: t.Parallel called multiple times")
    }
    if t.isEnvSet {
        panic("testing: t.Parallel called after t.Setenv; cannot set environment variables in parallel tests")
    }
    
    t.isParallel = true
    ^^^^^^^^^^^^^^^^^^^ 이 테스트를 병렬 실행 가능하다고 표시합니다.

    if t.parent.barrier == nil {
        // T.Parallel has no effect when fuzzing.
        // Multiple processes may run in parallel, but only one input can run at a
        // time per process so we can attribute crashes to specific inputs.
        return
    }

    // We don't want to include the time we spend waiting for serial tests
    // in the test duration. Record the elapsed time thus far and reset the
    // timer afterwards.
    t.duration += time.Since(t.start)

    // Add to the list of tests to be released by the parent.
    t.parent.sub = append(t.parent.sub, t)
    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ 이 테스트를 병렬 실행할 테스트 목록에 추가합니다.

    // Report any races during execution of this test up to this point.
    //
    // We will assume that any races that occur between here and the point where
    // we unblock are not caused by this subtest. That assumption usually holds,
    // although it can be wrong if the test spawns a goroutine that races in the
    // background while the rest of the test is blocked on the call to Parallel.
    // If that happens, we will misattribute the background race to some other
    // test, or to no test at all — but that false-negative is so unlikely that it
    // is not worth adding race-report noise for the common case where the test is
    // completely suspended during the call to Parallel.
    t.checkRaces()

    if t.chatty != nil {
        t.chatty.Updatef(t.name, "=== PAUSE %s\n", t.name)
    }
    running.Delete(t.name)


    t.signal <- true   // Release calling test.
    ^^^^^^^^^^^^^^^ 현재 테스트를 "일시 중지"하고 제어를 호출한 테스트로 반환합니다.
                    여기서 "일시 중지"는 현재 테스트 함수의 실행을 잠시 멈추고, 다른 테스트들이 실행될 수 있도록 하는 것을 의미합니다. 
                    이는 완전히 종료되는 것이 아니라, 나중에 재개될 준비를 하는 것입니다.

                    `t.signal` 채널에 `true`를 보내는 것은 현재 테스트가 병렬 실행 준비가 되었음을 테스트 러너에게 알리는 신호입니다.
                    현재 고루틴(테스트 함수)은 아래 `<-t.parent.barrier`에서 블록됩니다.
                    이렇게 함으로써, 다른 테스트들(특히 병렬화되지 않은 테스트들)이 실행될 기회를 갖게 됩니다.
                    
                    이렇게 하는 목적은 다음과 같습니다.
                    - 테스트의 순서를 제어합니다. 병렬화된 테스트들은 모든 비병렬 테스트가 완료된 후에 함께 실행되도록 합니다.
                    - 리소스 관리를 돕습니다. 동시에 실행되는 테스트의 수를 제한할 수 있게 해줍니다.
    
    // Go 테스트 프레임워크는 내부적으로 'barrier' 채널을 사용합니다. 
    // 이 채널은 비병렬 테스트가 모두 완료되었음을 알리는 데 사용됩니다.
    // 
    // 여기서 현재 고루틴(테스트 함수)이 블록됩니다.
    <-t.parent.barrier // Wait for the parent test to complete.
    ^^^^^^^^^^^^^^^^^^ 부모 테스트(비병렬 테스트들을 포함한 전체 테스트 스위트)가 완료될 때까지 대기합니다.
                       테스트 프레임워크는 내부적으로 `WaitGroup`과 같은 동기화 메커니즘을 사용하여 모든 비병렬 테스트가 완료되었는지 추적합니다.

                       비병렬 테스트가 완료되면, `t.parent.barrier` 채널이 닫히고,
                       `<-t.parent.barrier`에서 블록된 모든 병렬 테스트들이 깨어나게 됩니다.
    
    t.context.waitParallel()
    ^^^^^^^^^^^^^^^^^^^^^^^^ 다른 병렬 테스트들과 함께 실행을 시작합니다.
                             즉, 이때 실제 병렬 실행이 시작됩니다.

    if t.chatty != nil {
        t.chatty.Updatef(t.name, "=== CONT  %s\n", t.name)
    }
    running.Store(t.name, time.Now())
    t.start = time.Now()

    // Reset the local race counter to ignore any races that happened while this
    // goroutine was blocked, such as in the parent test or in other parallel
    // subtests.
    //
    // (Note that we don't call parent.checkRaces here:
    // if other parallel subtests have already introduced races, we want to
    // let them report those races instead of attributing them to the parent.)
    t.lastRaceErrors.Store(int64(race.Errors()))
}


// 이 함수는 동시에 실행되는 병렬 테스트의 수를 제한합니다.
func (c *testContext) waitParallel() {
    c.mu.Lock()
    
    // 실행 중인 테스트 수 < 최대 병렬 실행 수:
    // `c.maxParallel`(최대 병렬 실행 수)에 도달하지 않았다면, 실행을 허용합니다.
    if c.running < c.maxParallel {
        c.running++
        c.mu.Unlock()
        // `c.running++`를 하고 바로 반환하는 것은, 이 테스트가 실행될 수 있다는 허가를 의미합니다.
        // 실제 테스트 실행은 이 함수를 호출한 고루틴에서 이루어집니다.
        return
    }
    // 실행 중인 테스트 수 >= 최대 병렬 실행 수:
    // 대기 중인 테스트 수를 증가시키고 
    c.numWaiting++
    c.mu.Unlock()

    // `c.startParallel` 채널에서 신호를 기다립니다. 즉, 값을 받을 때까지 블록됩니다.
    // 이 신호는 다른 병렬 테스트가 완료되어 슬롯이 생길 때 보내집니다.
    // 
    // `release` 메서드에서 이 채널에 값을 보내면:
    // - 블록되어 있던 고루틴(테스트)이 깨어나고 실행을 계속합니다.
    // - 이는 곧 새로운 테스트가 실행을 시작할 수 있음을 의미합니다.
    <-c.startParallel
}


// 이 `release` 메서드는 테스트가 완료될 때 호출됩니다.
// 병렬 실행 제어는 주로 `waitParallel` 메서드에서 이루어지고,
// `release` 메서드는 단순히 *완료된 테스트의 자리를 대기 중인 다른 테스트로 채우는 역할*을 합니다.
func (c *testContext) release() {
    c.mu.Lock()
    if c.numWaiting == 0 {
        // `c.running--`로 실행 중인 테스트 수를 감소시키고
        c.running--
        c.mu.Unlock()
        // 반환합니다.
        return
    }

    // 대기 중인 테스트가 있다면 무조건 새로운 테스트 실행을 허용합니다:
    // 대기 중인 테스트가 있다면 (`c.numWaiting > 0`) 대기 중인 테스트 수를 감소시키고
    c.numWaiting--
    c.mu.Unlock()
    // `c.startParallel` 채널에 신호를 보냅니다.
    c.startParallel <- true // Pick a waiting test to be run.
}
```

작동 순서:

1. 테스트가 `waitParallel`을 호출합니다.
2. 실행 중인 테스트 수와 최대치를 비교해서
    1. 실행 중인 테스트 수 < 최대치: 실행 가능함을 의미하므로 바로 반환되어 실행됩니다.
    2. 실행 중인 테스트 수 >= 최대치: `c.numWaiting`이 증가하고, `<-c.startParallel`에서 대기합니다.
    3. 최대치에 도달했다면, `c.startParallel` 채널에서 신호를 기다립니다.
3. 다른 테스트가 완료되면 `release`가 호출됩니다.
    1. 대기중인 테스트가 없다면 테스트 수를 줄이고 리턴합니다.
    2. 대기중인 테스트가 있다면 `c.startParallel`에 신호를 보냅니다.
        1. 이 신호를 받은 대기 중이던 테스트 중 하나가 실행을 시작합니다.

이렇게 `waitParallel`과 `release` 메서드가 상호작용하면서 병렬 테스트의 실행을 조절합니다.
`waitParallel`에서 대기 상태로 들어가고, `release`에서 대기 중인 테스트를 깨우는 구조입니다.

---

네, 그 이해가 정확합니다. `t.Parallel()`의 동작 방식을 좀 더 자세히 설명해 드리겠습니다:

1. 테스트 실행 순서:
   - Go 테스트 러너는 먼저 모든 비병렬(serial) 테스트를 실행합니다.
   - 그 후에 `t.Parallel()`이 호출된 병렬 테스트들을 실행합니다.

2. `t.Parallel()` 호출 시 동작:
   - 해당 테스트 함수의 실행을 일시 중지합니다.
   - 이 테스트를 병렬 실행 대기 목록에 추가합니다.
   - 제어를 테스트 러너에 반환하여 다음 테스트로 넘어갑니다.

3. 비병렬 테스트 완료 후:
   - 모든 비병렬 테스트가 완료되면, 테스트 러너는 병렬 테스트들을 실행하기 시작합니다.
   - 이때 `waitParallel` 메서드를 통해 동시에 실행될 수 있는 테스트의 수를 제어합니다.

4. 병렬 테스트 실행:
   - 대기 중이던 병렬 테스트들이 동시에 실행되기 시작합니다.
   - 단, 동시에 실행되는 테스트의 수는 시스템 설정이나 `GOMAXPROCS` 값에 따라 제한됩니다.

예를 들어:

```go
func TestA(t *testing.T) {
    // 비병렬 테스트
}

func TestB(t *testing.T) {
    t.Parallel()
    // 병렬 테스트
}

func TestC(t *testing.T) {
    // 비병렬 테스트
}

func TestD(t *testing.T) {
    t.Parallel()
    // 병렬 테스트
}
```

실행 순서:
1. TestA 실행 (비병렬)
2. TestB에서 `t.Parallel()` 호출 → 실행 일시 중지 및 대기 목록에 추가
3. TestC 실행 (비병렬)
4. TestD에서 `t.Parallel()` 호출 → 실행 일시 중지 및 대기 목록에 추가
5. 모든 비병렬 테스트 완료
6. TestB와 TestD 동시에 실행 시작

이러한 방식으로 `t.Parallel()`은 비병렬 테스트가 모두 완료된 후에 병렬 테스트들을 그룹으로 실행하게 만듭니다. 이는 테스트 환경의 일관성을 유지하면서도 병렬 실행의 이점을 활용할 수 있게 해줍니다.

## 반복문 내에서 `t.Parallel()` 호출

```go
func TestSomething(t *testing.T) {
    t.Parallel() // TestSomething을 다른 최상위 테스트와 병렬로 실행

    testCases := // ... 테스트 케이스 정의 ...

    for testName, testCase := range testCases {
        testName, testCase := testName, testCase // 변수 섀도잉 (중요!)
        t.Run(testName, func(t *testing.T) {
            t.Parallel() // 각 서브테스트를 병렬로 실행
            // 테스트 로직
        })
    }
}
```

이 패턴은 테스트의 병렬화를 최대화하면서도 구조화된 방식으로 테스트를 조직화할 수 있게 해줍니다.

1. 외부 테스트 함수 (`TestSomething`)에서의 `t.Parallel()`:
    - 이 테스트 함수 자체를 다른 최상위 레벨 테스트 함수들과 병렬로 실행할 수 있게 합니다.

2. 서브테스트 (`t.Run()` 내부)에서의 `t.Parallel()`:
    - 각 서브테스트를 서로 병렬로 실행할 수 있게 합니다.

이렇게 사용하는 이유와 효과:

1. 독립적인 병렬화:
    - 외부 `TestSomething`은 다른 최상위 테스트 함수들과 병렬로 실행됩니다.
    - 각 서브테스트는 `TestSomething` 내의 다른 서브테스트들과 병렬로 실행됩니다.

2. 계층적 병렬화:
    - 이 구조는 테스트의 병렬화를 두 단계로 나눕니다: 최상위 레벨과 서브테스트 레벨.

3. 리소스 활용 최적화:
    - 더 세분화된 병렬화로 시스템 리소스를 더 효율적으로 사용할 수 있습니다.

4. 테스트 격리:
    - 각 서브테스트가 독립적으로 병렬화되므로, 서브테스트 간의 격리가 보장됩니다.

5. 유연성:
    - 필요에 따라 특정 서브테스트만 직렬로 실행하거나, 전체 `TestSomething`을 직렬로 실행할 수 있는 유연성을 제공합니다.

주의사항:
- 루프 내에서 `t.Parallel()`을 사용할 때는 변수 섀도잉(위 코드의 `testName, testCase := testName, testCase`)이 중요합니다. 이는 각 반복에서 고유한 변수 복사본을 사용하도록 보장합니다.

## 비병렬 테스트를 한 고루틴에서 실행하지 않는 이유

- 독립성: 각 테스트는 독립적인 환경에서 실행되어야 합니다. 별도의 고루틴에서 실행함으로써 테스트 간 격리를 보장합니다.
- 에러 처리: 각 테스트는 자체적인 에러 처리 및 보고 메커니즘을 가져야 합니다. 별도의 고루틴에서 실행하면 이를 더 쉽게 관리할 수 있습니다.
- 유연성: 일부 테스트는 `t.Parallel()`을 호출하지 않고도 내부적으로 고루틴을 사용할 수 있습니다. 각 테스트를 별도의 고루틴에서 실행하면 이러한 유연성을 제공합니다.
- 타임아웃 관리: 각 테스트에 대해 개별적인 타임아웃을 설정하고 관리하기가 더 쉽습니다.

Go 테스트 프레임워크는 각 테스트를 `tRunner` 함수를 통해 실행합니다. 이 함수는 각 테스트에 대해 새로운 고루틴을 생성합니다.

```go
func tRunner(t *T, fn func(t *T)) {
    // ...
    if !t.parallel {
        fn(t)
    } else {
        // 병렬 테스트의 경우 새 고루틴에서 실행됩니다.
        go fn(t)
        if t.parent != nil {
            <-t.parent.barrier // 부모 테스트 완료 대기
        }
    }
    // ...
}
```

우선 golang에서의 테스트는
1. 병렬 아닌 테스트를 모두 실행하고,
2. 그 다음에 병렬로 테스트해야 하는 테스트들을 실행하게 되는데요.

그래서 테스트 간 순서를 지켜야 하는 게 아니라면 병렬 아닌 테스트가 없도록 해서 모두 병렬로 테스트하게 하는 게 효율적이라는 거 같습니다. 병렬 아닌 테스트 집합 테스트 후 병렬 테스트 집합 테스트가 되니까요.

여기서 반복문 내에서 실행되는 `t.Run`은 별도의 고루틴에서 주어진 이름으로 `t` 리시버의 서브테스트로 `f`를 실행하게 되고,
이는 `f`가 리턴하거나 또는 `t.Parallel` 호출하여 병렬 테스트가 되기 전까지 블록됩니다.

```go
// Run runs f as a subtest of t called name. It runs f in a separate goroutine
// and blocks until f returns or calls t.Parallel to become a parallel test.
```

따라서
- 함수 최상단에서의 `t.Parallel`은 부모 테스트가 다른 `func Test*` 같은 가장 바깥 뎁스의 테스트 함수들과 병렬로 실행할 것을 지정하는 것이고
- 그래서 여러 병렬로 실행되는 테스트중에서 실제로 실행되면, 다시 그 테스트 내에서 서브 테스트가 다시 병렬로 테스트 되도록, 즉 블록되지 않으려면 `t.Parallel`을 호출해줘야 한다

이렇게 이해했습니다.
