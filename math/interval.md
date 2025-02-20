# interval

- [interval](#interval)
    - [interval? 구간](#interval-구간)
    - [열린/열림(open)과 닫힌/닫힘(closed)](#열린열림open과-닫힌닫힘closed)
        - [열린(Open)](#열린open)
        - [열린 구간의 수학적 의미](#열린-구간의-수학적-의미)
        - [`열린 구간`의 특징](#열린-구간의-특징)
        - [닫힌(Closed)](#닫힌closed)
        - [`열린 구간`과 `닫힌 구간`의 선택](#열린-구간과-닫힌-구간의-선택)
    - [`열린 구간` (Open Interval)](#열린-구간-open-interval)
    - [`닫힌 구간` (Closed Interval)](#닫힌-구간-closed-interval)
    - [`반열린 구간` (Half-Open Interval, 반개구간)](#반열린-구간-half-open-interval-반개구간)
    - [`반닫힌 구간` (Half-Closed Interval)](#반닫힌-구간-half-closed-interval)
    - [무한 구간과 그 외 특별한 구간](#무한-구간과-그-외-특별한-구간)
    - [Examples](#examples)
        - [예제: 온도 제어 시스템](#예제-온도-제어-시스템)
            - [`열린 구간` `(lower, upper)`: 시스템이 과열되거나 과냉될 위험을 감소](#열린-구간-lower-upper-시스템이-과열되거나-과냉될-위험을-감소)
            - [`닫힌 구간` `[lower, upper]`: 시스템이 설정된 경계값까지 허용](#닫힌-구간-lower-upper-시스템이-설정된-경계값까지-허용)
        - [왜 이렇게 정의됐는지?](#왜-이렇게-정의됐는지)

## interval? 구간

수학과 프로그래밍에서 사용되는 다양한 종류의 구간은 데이터를 다루는 방식에 큰 영향을 미칩니다.
각 구간의 특성을 이해하는 것은 데이터의 범위를 정의하고, 연산의 정확도를 보장하는 데 중요합니다. 다음은 각 구간에 대한 설명과 예시입니다.

## 열린/열림(open)과 닫힌/닫힘(closed)

`열린 구간`과 `닫힌 구간`의 개념은 수학적인 정의와 실용적인 이유 모두에 기반을 두고 있습니다.
이 구분은 *집합론*과 *실수선*의 특성을 처리하는 방식에서 중요한 역할을 합니다.

`열린 구간`과 `닫힌 구간`이 각각 끝점을 포함하거나 포함하지 않는 방식으로 정의된 이유는 다음 이유에 기반을 두고 있습니다.
- 수학적 편의성
- 연속성과 분리성의 처리
- 그리고 역사적 관례 등

이 구분은 특히 실수 집합과 관련된 수학적 개념을 다룰 때 중요하게 적용됩니다.

### 열린(Open)

`(a, b)`는 끝점 `a`와 `b`를 포함하지 않습니다.
이 구간은 `a`보다 크고 `b`보다 작은 모든 실수를 포함합니다.
수학에서 `열린 구간`은 특히 위상수학에서 중요한 역할을 합니다.

### 열린 구간의 수학적 의미

1. 위상학적 연속성:

   `열린 구간`은 "열린 집합"의 개념과 밀접하게 연관되어 있습니다.
   위상공간에서 열린 집합은 *연속성*을 특징으로 합니다.

   `열린 구간`의 모든 점은 해당 구간 내에서 작은 변동이 있어도 여전히 구간 내에 존재합니다.
   이러한 성질은 함수의 연속성을 정의하고 분석하는 데 중요합니다.

2. 무한 분할 가능성:

   `열린 구간`은 임의의 두 점 사이에 항상 다른 점들이 존재합니다.
   이 속성은 극한과 미적분학에서 중요하게 사용됩니다.

   `열린 구간`을 사용하면 *경계에서의 정의나 특별한 처리 없이도* 내부의 모든 점에 대한 처리가 가능합니다.

3. 국부적 유연성:

   `열린 구간`에서는 구간의 모든 점에 대하여 그 점을 중심으로 하는 더 작은 `열린 구간`을 찾을 수 있습니다.
   이는 "국부적 성질"을 탐구할 때 유리하며, 복잡한 수학적 구조에서 지역적 분석을 용이하게 합니다.

### `열린 구간`의 특징

- 국부적 유연성

    `열린 구간`의 각 점 주변에는 해당 점을 중심으로 하는 또 다른 작은 `열린 구간`이 존재합니다.
    이 속성은 *지역성*을 표현하며, *함수의 연속성*과 같은 수학적 개념을 다룰 때 유용합니다.

- 토폴로지적 열림

    `열린 구간`은 수학적 공간에서 "열린 집합"의 기본 예로 사용됩니다.
    열린 집합은 공간의 위상 구조를 정의하는 데 기초가 되며, 연속함수, 수렴성 등의 개념을 정의하는 데 중요합니다.

### 닫힌(Closed)

`닫힌 구간` `[a, b]`는 양 끝점 `a`와 `b`를 포함합니다.
이 구간은 `a` 이상 `b` 이하의 모든 실수를 포함합니다

이는 수학적으로 다음과 같은 이유로 유리합니다:

1. 완전성과 포괄성:

   `닫힌 구간`은 그 구간의 모든 값을 포함합니다.
   이는 극한, 최대/최소 문제에서 특히 중요합니다.

   예를 들어, `닫힌 구간`에서 정의된 연속 함수는 항상 최댓값과 최솟값을 가집니다(볼차노-베어슈트라스 정리).

2. 경계 포함:

   `닫힌 구간`은 경계 조건을 만족하는 문제에 유용합니다.

   예를 들어, 물리학에서 경계가 고정된 상황을 모델링할 때 `닫힌 구간`을 사용하여 경계 조건을 명확히 할 수 있습니다.

3. 수학적 편의성:

   분석학에서는 종종 `닫힌 구간`에서 정의된 함수의 성질을 이용하여 그 함수의 연속성, 미분 가능성 등을 평가합니다.
   `닫힌 구간`에서는 경계에서도 함수의 값이 정의되어 있어야 하므로, 경계에서의 행동을 분석할 수 있습니다.

`닫힌 구간`은 다음과 같은 특성을 가집니다:

- 완전성

    `닫힌 구간`은 경계를 포함함으로써 완전한 범위를 제공합니다.
    이는 극한값, 최댓값 또는 최솟값 같은 수학적 연산을 정의할 때 중요합니다.

- 안정성과 구속

    `닫힌 구간`은 그 경계가 포함되므로, 연산과 함수의 값이 이 범위를 벗어나지 않도록 구속합니다.

    예를 들어, 최적화 문제에서 변수의 값을 특정 범위 내로 제한할 때 `닫힌 구간`이 사용됩니다.

### `열린 구간`과 `닫힌 구간`의 선택

`열린 구간`과 `닫힌 구간`의 정의는 각각의 수학적 성질과 응용 필요성에 따라 발전해 왔습니다. 이러한 정의는 수학적 분석을 간소화하고 특정 문제에 대한 적절한 도구를 제공하는 데 목적이 있습니다. 이 구분은 수학적 설계와 이론적 배경에 깊이 뿌리박고 있으며, 실제 응용에서도 그 유용성이 입증되었습니다.

`열린 구간`과 `닫힌 구간`은 각각의 수학적 문제와 실용적 응용에 따라 선택됩니다.

예를 들어,
- *미분 가능성이나 연속성*을 검토할 때는 `열린 구간`이 유리할 수 있습니다.
- *최댓값이나 최솟값을 확실하게 포함*해야 할 때는 `닫힌 구간`을 사용합니다.

이러한 구간의 사용은 궁극적으로 특정한 문제의 요구사항과 목적에 맞게 선택되며, 구간의 열림과 닫힘은 그 안에서 다루는 값들의 포함 여부를 명확하게 정의함으로써 문제를 명확하게 해결하는 데 도움을 줍니다.

## `열린 구간` (Open Interval)

> `열린 구간` `(a, b)`는 양 끝점 `a`와 `b`를 포함하지 않습니다.

구간 `(0, 1)`은 0과 1 사이의 모든 실수를 포함하지만, 0과 1 자체는 포함하지 않습니다. 즉,

$0 < x < 1$

`열린 구간`은 끝점을 포함하지 않기 때문에, 구간 내의 모든 점이 구간의 내부에 있다고 간주됩니다.
이는 미적분학에서 극한을 다룰 때 유용합니다.

예를 들어, 함수의 연속성을 논할 때 `열린 구간`은 함수가 끝점에서 정의되지 않아도 되는 경우에 사용됩니다.

프로그래밍에서 `열린 구간`은 일반적으로 직접 구현해야 합니다.

```go
a := 0.0
b := 1.0
x := 0.5

// x가 a보다 크고 b보다 작은지 확인합니다. `열린 구간`은 끝점을 포함하지 않습니다.
if x > a && x < b {
    fmt.Println("x is in the open interval (0, 1)")
} else {
    fmt.Println("x is not in the open interval (0, 1)")
}
```

예를 들어, 숫자 배열에서 특정 범위를 검사할 때 양 끝 값을 제외할 수 있습니다.

## `닫힌 구간` (Closed Interval)

> `닫힌 구간` `[a, b]`는 양 끝점 `a`와 `b`를 포함합니다.

구간 `[0, 1]`은 0과 1을 포함한 모든 실수를 포함합니다. 즉,

$0 ≤ x ≤ 1$

`닫힌 구간`은 끝점을 포함하기 때문에, 구간 내의 모든 점이 구간의 경계에 포함될 수 있습니다.
이는 *함수의 정의역*을 명확히 할 때 유용합니다.

예를 들어, 함수가 특정 구간에서 정의되고 그 구간의 끝점에서도 정의될 때 `닫힌 구간`을 사용합니다.

배열의 인덱스를 통해 전체 범위를 선택할 때 `닫힌 구간`의 개념이 사용됩니다.

```go
a := 0
b := 5
array := []int{1, 2, 3, 4, 5}

// 배열의 인덱스를 사용하여 a부터 b까지 접근합니다.
// `닫힌 구간`이므로 a와 b 모두 포함됩니다.
for i := a; i <= b; i++ {
    //        ^^^ 반 `열린 구간`과 조건이 다름
    fmt.Printf("Element at index %d is %d\n", i, array[i])
}
```

예를 들어, 리스트의 첫 요소부터 마지막 요소까지 접근하려면 전체 인덱스를 사용합니다.

## `반열린 구간` (Half-Open Interval, 반개구간)

> `반열린 구간`: `[a, b)`는 시작점 `a`는 포함하지만 끝점 `b`는 포함하지 않습니다.

`[1, 5)`: 1은 포함되지만 5는 포함되지 않으며, 1부터 4.999...까지의 모든 실수를 포함합니다.

$1 ≤ x < 5$

```go
if 1 <= x && x < 5 {}
```

```go
a := 0
b := 5
array := []int{1, 2, 3, 4, 5}

// a부터 b-1까지 반복합니다. 이는 b를 포함하지 않는 `반열린 구간`을 나타냅니다.
for i := a; i < b; i++ {
    //        ^^^ `닫힌 구간`과 조건이 다름
    fmt.Printf("Element at index %d is %d\n", i, array[i])
}
```

Python의 `range(0, 5)` 함수는 0부터 4까지의 정수를 생성하며, 5는 포함하지 않습니다.

## `반닫힌 구간` (Half-Closed Interval)

> `반닫힌 구간` `(a, b]`는 시작점 `a`는 포함하지 않고 끝점 `b`는 포함합니다.

구간 `(1, 5]`은 1은 포함되지 않지만 5는 포함되며, 1.000...부터 5까지의 모든 실수를 포함합니다.

$1 < x ≤ 5$

```go
if 1 < x && x <= 5 {}
```

이 구간은 일반적으로 특정 함수를 통해 조건을 설정함으로써 프로그래밍에서 구현됩니다.

```go
a := 0
b := 5
array := []int{1, 2, 3, 4, 5}

// a+1부터 b까지 반복합니다. 시작점 a는 포함되지 않고, 끝점 b는 포함됩니다.
for i := a + 1; i <= b; i++ {
    fmt.Printf("Element at index %d is %d\n", i, array[i-1])
}
```

## 무한 구간과 그 외 특별한 구간

> 무한 구간은 한 쪽 또는 양쪽 끝이 무한대인 구간입니다.
> `(-∞, b)`, `(a, ∞)`, `(-∞, ∞)` 등으로 표현됩니다.

- 구간 `(-∞, b)`는 음의 무한대부터 `b`까지 모든 실수를 포함하지만 `b`는 포함하지 않습니다.
- 구간 `[a, ∞)`는 `a`부터 양의 무한대까지 모든 실수를 포함하며 `a`도 포함합니다.

예시:
- `(-∞, 5)`: 5보다 작은 모든 실수를 포함합니다.(`x < 5`)
- `[1, ∞)`: 1 이상인 모든 실수를 포함합니다.(`x >= 1`)
- `(-∞, ∞)`: 모든 실수를 포함합니다.(항상 참)

데이터의 범위를 무제한으로 설정할 때 무한 구간을 사용할 수 있습니다.

무한 구간은 실수 전체를 다루거나 특정 값 이상 또는 이하의 모든 값을 포함할 때 사용됩니다.
이는 함수의 정의역이 무한한 경우나, 특정 조건을 만족하는 모든 값을 포함할 때 유용합니다.

예를 들어, 어떤 값 이상의 모든 데이터를 선택하는 필터 조건 등에서 사용됩니다.

```go
// Simulating infinite upper bound
for i := 0; i < 1000000; i++ { // 무한 구간을 시뮬레이션하기 위해 큰 수 사용
    // Simulate some operations
}
fmt.Println("Simulated up to an effectively infinite upper bound")
```

## Examples

### 예제: 온도 제어 시스템

온도를 제어하는 시스템을 예로 들면, 시스템이 특정 온도 범위를 유지하도록 설계되었다고 가정해 보겠습니다.
이 시스템에서는 온도가 너무 낮거나 높아지지 않도록 설정된 온도 범위 내에서 유지해야 합니다.

#### `열린 구간` `(lower, upper)`: 시스템이 과열되거나 과냉될 위험을 감소

`열린 구간`을 사용하는 경우, *경계값 자체는 포함되지 않으므로, 설정된 최대 또는 최소값 바로 전*에서 조치를 취하기 시작합니다. 이는 *경계값에 도달하기 전에 미리 대응*할 수 있도록 해줍니다.

```go
package main

import (
    "fmt"
)

func main() {
    lower := 18.0
    upper := 22.0
    currentTemp := 22.0

    if currentTemp > lower && currentTemp < upper {
        fmt.Println("Temperature is within the safe range.")
    } else {
        fmt.Println("Temperature is out of range!")
    }
}
```

#### `닫힌 구간` `[lower, upper]`: 시스템이 설정된 경계값까지 허용

`닫힌 구간`을 사용하는 경우, 경계값을 포함합니다.
이는 설정된 최대 또는 최소값을 정확히 포함하기 때문에, *경계 상황에서도 시스템이 정상 범위 내로 간주*됩니다.

```go
package main

import (
    "fmt"
)

func main() {
    lower := 18.0
    upper := 22.0
    currentTemp := 22.0

    if currentTemp >= lower && currentTemp <= upper {
        fmt.Println("Temperature is within the safe range.")
    } else {
        fmt.Println("Temperature is out of range!")
    }
}
```

### 왜 이렇게 정의됐는지?

- 진입/이탈 조건

    `닫힌 구간`은 특정 값을 경계로 포함하므로, 그 값에 도달했을 때 문제가 발생하지 않도록 하기 위함입니다.
    `열린 구간`은 경계에 도달하기 전에 미리 조치를 취할 수 있도록 설정된 것입니다.

- 수학적 편의

    수학에서는 `열린 구간`이 *연속성*과 관련된 많은 이론에서 자연스럽게 적용됩니다.

    예를 들어, `열린 구간`에서 정의된 함수는 그 구간의 모든 점에서 연속이라고 쉽게 말할 수 있습니다.
