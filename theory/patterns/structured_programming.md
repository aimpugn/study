# Structured Programming

- [Structured Programming](#structured-programming)
    - [구조화된 프로그래밍이란?](#구조화된-프로그래밍이란)
    - [구조화되지 않은 프로그래밍의 특징](#구조화되지-않은-프로그래밍의-특징)
    - [구조화된 프로그래밍의 특징](#구조화된-프로그래밍의-특징)
    - [구조화된/구조화되지 않은 프로그래밍의 차이점](#구조화된구조화되지-않은-프로그래밍의-차이점)
    - [관련 소프트웨어 엔지니어링 이론](#관련-소프트웨어-엔지니어링-이론)

## 구조화된 프로그래밍이란?

구조화된 프로그래밍은 **제어 흐름을 명확하게 하는** 프로그래밍 패러다임으로, **코드의 가독성과 유지보수성을 높이기 위해** 프로그램을 더 작은 단위로 분해하여 작성하는 방법입니다.
이 패러다임은 **Dijkstra, Hoare, Wirth**와 같은 컴퓨터 과학자들에 의해 1960년대 후반에 제안되었습니다.

구조화된 프로그래밍의 기본 요소는 세 가지로 정의됩니다.
- **순차적 실행**: 코드는 순차적으로 위에서 아래로 실행됩니다.
- **분기(조건문)**: 특정 조건에 따라 코드의 실행 흐름을 변경합니다 (e.g., `if`, `switch` 문).
- **반복문**: 특정 조건이 만족될 때까지 코드 블록을 반복 실행합니다 (e.g., `for`, `while` 문).

구조화된 프로그래밍은 *코드의 가독성*과 *유지보수성*을 높이기 위해 고안된 패러다임으로, 순차적 실행, 조건문, 반복문을 사용하여 *명확한 제어 구조를 유지*하는 것을 목표로 합니다.
반면, 구조화되지 않은 프로그래밍은 GOTO 문을 사용하여 제어 흐름이 복잡해질 수 있으며, 이는 코드의 이해와 유지보수를 어렵게 만듭니다.
따라서 구조화된 프로그래밍 원칙을 준수하는 것이 더 나은 소프트웨어 개발을 위한 중요한 기법으로 간주됩니다.

## 구조화되지 않은 프로그래밍의 특징

구조화되지 않은 프로그래밍은 주로 **GOTO 문**을 사용하여 코드의 흐름을 제어하는 방식입니다.
이로 인해 코드의 흐름이 복잡하고 예측하기 어려워지며, 프로그램의 상태를 추적하기 어려워집니다.
이러한 프로그램은 "스파게티 코드"로 불리며, 다음과 같은 특징을 가집니다:

> 스파게티 코드: 코드가 복잡하게 얽히게 되어 이해하기 어렵고, 유지보수가 힘듦

- 비선형적 흐름: 프로그램의 실행 흐름이 GOTO 문을 통해 무작위로 점프함.
- 가독성 저하: 코드의 흐름을 따라가기가 어렵고, 디버깅과 유지보수가 힘듦.
- 복잡성 증가: 프로그램의 상태를 이해하고 예측하기 어렵게 만듦.
- 비예측성: GOTO 문이 프로그램의 제어 흐름을 임의로 변경하여 코드의 동작을 예측하기 어려움.

```go
// 구조화되지 않은 프로그래밍 (비구조적 코드)
package main

import "fmt"

func main() {
    var i int
    i = 0

start:
    fmt.Println(i)
    i = i + 1
    if i < 10 {
        goto start
    }
}
```

위의 코드에서 `goto` 문을 사용하여 루프를 구현하고 있습니다.
이는 코드 흐름을 복잡하게 만들어 가독성과 유지보수성을 떨어뜨립니다.

## 구조화된 프로그래밍의 특징

구조화된 프로그래밍은 다음과 같은 특징을 가집니다:

1. **명확한 제어 구조**: 코드의 흐름이 명확하여 프로그램의 작동 방식을 쉽게 이해할 수 있습니다.
2. **모듈화**: 프로그램을 더 작은 모듈로 나누어 작성함으로써 재사용성과 유지보수성을 높입니다.
3. **디버깅 용이**: 코드의 흐름이 명확하므로 버그를 찾고 수정하기가 더 쉽습니다.

```go
// 구조화된 프로그래밍 (구조적 코드)
package main

import "fmt"

func main() {
    for i := 0; i < 10; i++ {
        fmt.Println(i)
    }
}
```

위의 코드에서는 `for` 문을 사용하여 루프를 구현하고 있습니다.
이는 코드 흐름이 명확하며 가독성과 유지보수성이 향상됩니다.

## 구조화된/구조화되지 않은 프로그래밍의 차이점

1. 코드 흐름:
    - 구조화되지 않은 프로그래밍: GOTO 문을 사용하여 흐름이 복잡하고 비선형적입니다.
    - 구조화된 프로그래밍: 반복문, 조건문, 함수 등을 사용하여 흐름이 명확하고 *선형적*입니다.

2. 가독성:
    - 구조화되지 않은 프로그래밍: 코드의 흐름을 따라가기 어렵고, 이해하기 힘듭니다.
    - 구조화된 프로그래밍: 코드의 흐름이 명확하고, *가독성*이 높습니다.

3. 유지보수성:
    - 구조화되지 않은 프로그래밍: 디버깅과 유지보수가 어렵습니다.
    - 구조화된 프로그래밍: 디버깅과 유지보수가 용이합니다.

4. 복잡성:
    - 구조화되지 않은 프로그래밍: 복잡성이 높고, 프로그램의 상태를 예측하기 어렵습니다.
    - 구조화된 프로그래밍: 복잡성이 낮고, 프로그램의 상태를 *예측하기 쉽습니다*.

## 관련 소프트웨어 엔지니어링 이론

- 모듈화(Modularity):
    - 구조화된 프로그래밍은 코드를 기능별로 나누어 모듈화하는 것을 장려합니다.
    - 모듈화는 코드의 재사용성을 높이고, 변경 시 영향을 최소화합니다.

- 절차적 프로그래밍(Procedural Programming):
    - 구조화된 프로그래밍은 절차적 프로그래밍과 밀접하게 관련되어 있으며, 기능을 절차(또는 함수)로 나누어 구현합니다.
    - 이는 코드를 논리적 단위로 분리하여 가독성과 유지보수성을 높입니다.
