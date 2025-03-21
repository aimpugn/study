# Pointer

- [Pointer](#pointer)
    - [포인터와 함수](#포인터와-함수)
        - [값 전달 vs. 포인터 전달](#값-전달-vs-포인터-전달)
        - [함수에서 포인터 반환하기](#함수에서-포인터-반환하기)
        - [`new` 사용해서 값으로 리턴하기](#new-사용해서-값으로-리턴하기)
    - [구조체와 포인터](#구조체와-포인터)
        - [구조체 필드 접근하기](#구조체-필드-접근하기)
        - [메서드 리시버로서의 포인터 vs. 값](#메서드-리시버로서의-포인터-vs-값)
        - [구조체 포인터와 복사본 문제](#구조체-포인터와-복사본-문제)
        - [구조체가 큰 경우나 복사 비용이 큰 경우](#구조체가-큰-경우나-복사-비용이-큰-경우)
            - [구조체 크기](#구조체-크기)
                - [Example](#example)
            - [복사 비용](#복사-비용)
            - [결정 방법](#결정-방법)
        - [구조체 속성과 포인터](#구조체-속성과-포인터)
            - [1. 메모리 사용 및 할당](#1-메모리-사용-및-할당)
            - [2. 구조체의 변경 가능성](#2-구조체의-변경-가능성)
            - [3. 프로그램 로직 및 설계](#3-프로그램-로직-및-설계)
    - [기타](#기타)
        - [함수 통해서 포인터 얻는 경우 의도와 다르게 동작할 수 있다](#함수-통해서-포인터-얻는-경우-의도와-다르게-동작할-수-있다)

## 포인터와 함수

### 값 전달 vs. 포인터 전달

Go에서 함수를 호출할 때 인자는 기본적으로 값(value)으로 전달된다.
함수에 인자로 전달된 변수의 복사본이 생성되어 함수 내에서 사용되고, 원본 변수는 변경되지 않는다.

반면, 포인터로 전달하는 경우에는 변수의 메모리 주소를 넘겨주게 되므로, 함수 내에서 해당 주소를 통해 원본 변수의 값을 직접 변경할 수 있다. 이는 특히 큰 데이터 구조를 다루거나, 함수 내에서의 변경사항을 호출한 측에 그대로 반영해야 할 때 유용하다.

### 함수에서 포인터 반환하기

함수가 특정 데이터의 포인터를 반환하면, 호출자는 반환된 포인터를 통해 해당 데이터에 접근할 수 있다.
이 방법은 특히 함수 내에서 새롭게 생성한 데이터를 밖으로 전달하고자 할 때 유용하다.
하지만, 지역 변수의 주소를 반환하는 경우 주의해야 한다.
스택에 할당된 지역 변수의 생명 주기는 함수 호출이 끝나면 종료되기 때문에, 외부에서 접근이 불가능해진다.
따라서 안전하게 반환하려면 힙에 할당된 데이터의 주소를 반환하거나, 지역 변수가 아닌 외부에서 생성된 데이터의 주소를 반환해야 한다.

스택에 할당된 지역 변수가 함수의 스코프를 벗어날 때 힙으로 자동으로 이스케이프(escape)되는 것은 일반적인 상황이 아니다.
여기서 중요한 개념은 "이스케이프 분석(escape analysis)"이다.
Go 컴파일러는 이스케이프 분석을 통해 변수가 함수 스코프 밖에서도 접근될 필요가 있는지를 결정한다.
이 과정에서 컴파일러는 변수가 힙에 할당될 필요가 있는지를 결정하게 되는데, 이는 주로 *변수의 생명 주기가 함수 호출보다 길어야 할 때* 발생한다.

이스케이프 분석은 컴파일 타임에 수행되며, 그 결과로 일부 변수는 스택 대신 힙에 할당될 수 있다.
이는 개발자가 명시적으로 관리하지 않아도 되는 부분이며, 컴파일러가 자동으로 최적의 메모리 할당 전략을 결정한다.

그러나 함수가 지역 변수의 포인터를 반환할 경우 명시적으로 주의해야 하는 이유가 있다.

1. **명시적인 힙 할당 없이 지역 변수의 주소를 반환하는 경우**

    일반적으로 지역 변수는 스택에 할당되며, 함수가 끝나면 해당 스택 프레임은 제거된다.
    만약 컴파일러가 이 변수가 함수 밖에서 접근될 필요가 없다고 판단한다면, 이 변수는 힙에 할당되지 않는다.
    따라서 이 변수의 주소를 반환하는 것은 안전하지 않다.
    외부에서 접근하려고 할 때 이미 해제되었거나 다른 용도로 재사용될 수 있는 메모리 영역을 가리키게 될 수 있다.

2. **이스케이프 분석에 의한 힙 할당**

    변수가 함수 밖에서도 접근될 필요가 있다고 컴파일러가 판단하는 경우, 해당 변수는 힙에 할당된다.
    이 경우에는 함수가 끝나더라도 변수의 메모리는 유지되기 때문에, 반환된 포인터를 통해 안전하게 접근할 수 있다/

결론적으로, 개발자는 이스케이프 분석의 결과를 명시적으로 신경 쓸 필요는 없지만, 함수에서 지역 변수의 포인터를 반환할 때 *해당 변수가 안전하게 접근될 수 있는지 (즉, 컴파일러가 힙에 할당하기로 결정했는지)를 고려해야* 한다.
컴파일러의 이스케이프 분석에만 전적으로 의존하지 않고, 힙에 할당해야 하는 데이터는 명시적으로 힙 할당하는 것이 좋다.
예를 들어, `new`, `make` 함수나 슬라이스, 맵, 채널과 같은 참조 타입을 사용하는 경우가 이에 해당한다.

아래 예시는 지역 변수의 포인터를 반환하고, 이 변수가 스택에 할당될 경우 발생할 수 있는 문제를 보여준다.
*컴파일러의 이스케이프 분석이 변수를 힙으로 이동시키지 않는다면*, 이 코드는 안전하지 않다.

```go
package main

import "fmt"

func unsafePointer() *int {
    a := 10
    return &a
}

func main() {
    p := unsafePointer()
    fmt.Println(*p) // 이 시점에서 *p의 값은 정의되지 않음 or 예상치 못한 값
}
```

`new` 함수나 슬라이스, 맵, 채널과 같은 참조 타입을 사용하여 명시적으로 힙에 데이터를 할당한다.

```go
package main

import "fmt"

func safePointer() *int {
    a := new(int)
    *a = 10
    return a
}

func main() {
    p := safePointer()
    fmt.Println(*p) // 안전하게 10 출력
}
```

함수 외부에서 데이터를 생성하고, 그 포인터를 함수 인자로 받아 작업을 수행한 후 반환한다.
이 경우, 데이터의 생명 주기는 함수 호출과 무관하게 관리된다.

```go
package main

import "fmt"

func modifyValue(p *int) *int {
    *p = 20
    return p
}

func main() {
    a := 10
    p := modifyValue(&a)
    fmt.Println(*p) // 20 출력, 안전함
}
```

이 예시들에서 볼 수 있듯이, Go에서 함수가 지역 변수의 포인터를 반환할 때는 해당 변수가 안전하게 접근될 수 있는지를 신경써야 한다. 컴파일러의 이스케이프 분석이 변수를 자동으로 힙에 할당하기도 하지만, 이를 명시적으로 관리하는 것이 좋다.

### `new` 사용해서 값으로 리턴하기

`new` 키워드를 사용하여 생성한 변수는 해당 타입의 포인터를 반환하고, 이 포인터를 통해 해당 변수를 참조한다.
함수에서 새로운 인스턴스를 생성하고 이를 값으로 반환하고자 할 때, `return *variable` 형태로 포인터가 가리키는 값을 반환할 수 있다.

하지만, 이 방식이 일반적인 프랙티스인지 여부는 사용하는 컨텍스트와 상황에 따라 다르다.
값으로 반환하기 위해 **`*`를 사용하는 것은 객체의 복사본을 만들어 반환한다는 의미**다.
객체가 크거나, 복사 비용이 큰 경우에는 성능에 영향을 줄 수 있으므로 포인터를 그대로 반환하는 것이 더 효율적일 수 있다.

```go
type SomeType struct {
    Field1 int
    Field2 string
}

func instantiate() SomeType {
    p := new(SomeType) // SomeType의 새 인스턴스를 생성하고 포인터를 반환
    p.Field1 = 10      // 필드 초기화
    p.Field2 = "example"
    return *p // p가 가리키는 인스턴스의 복사본을 반환
}
```

위의 예제에서 `instantiate` 함수는 `SomeType`의 인스턴스를 생성하고 초기화한 뒤, 이를 값으로 반환한다.
반환되는 것은 **`SomeType` 인스턴스의 복사본**이며, 이 복사 과정은 특히 구조체가 큰 경우 성능에 영향을 줄 수 있다.

- **성능**

    구조체가 큰 경우나 복사 비용이 큰 경우, 포인터를 반환하는 것이 선호된다.
    포인터는 구조체의 사이즈와 관계없이 일정한 메모리 크기를 가지므로, 성능이 중요한 상황에서는 포인터 반환을 고려해야 한다.

- **불변성**

    값으로 반환하는 경우, 반환된 인스턴스는 원본과 독립적인 복사본이다.
    이는 원본 데이터의 불변성을 보장할 수 있는 방법 중 하나다.
    반환된 복사본을 변경해도 원본 인스턴스에는 영향을 주지 않는다.

- **스레드 안전성**

    복사본을 반환하는 방식은 스레드 간의 데이터 공유에 있어서 안전한 방법일 수 있다.
    각 스레드가 데이터의 독립적인 복사본을 가지고 작업하기 때문에, 복사본을 통해 동시성 문제를 방지할 수 있다.

## 구조체와 포인터

### 구조체 필드 접근하기

구조체의 포인터를 통해 구조체 필드에 접근할 때는 `.` 연산자를 사용합니다. Go는 내부적으로 포인터를 따라가 필드에 접근하는 것을 처리해줍니다. 예를 들어, `person` 구조체의 포인터 `p`가 있을 때, `p.name`은 `(*p).name`을 축약한 형태로, `p`가 가리키는 구조체의 `name` 필드에 접근합니다.

### 메서드 리시버로서의 포인터 vs. 값

메서드 리시버를 정의할 때 포인터를 사용하면, 해당 메서드 내에서 구조체의 필드를 수정할 수 있습니다. 반면 값 리시버를 사용하는 메서드는 원본 구조체의 복사본에서 작동하므로, 구조체의 필드를 변경해도 원본에는 영향을 주지 않습니다. 포인터 리시버는 구조체의 상태를 변경하거나, 메모리를 절약하기 위해 주로 사용됩니다.

### 구조체 포인터와 복사본 문제

구조체를 함수 인자로 전달하면 기본적으로 그 값의 복사본이 생성되어 전달됩니다. 이는 구조체가 크거나, 함수에서의 변경을 원본에 반영하고자 할 때 비효율적일 수 있습니다. 이러한 문제를 해결하기 위해, 구조체의 포인터를 전달하면 원본 구조체에 대한 변경이 가능하고, 데이터 복사에 따른 오버헤드를 줄일 수 있습니다. 구조체를 포인터로 전달하는 것은 특히 상태를 변경하는 메서드나, 성능이 중요한 상황에서 중요합니다.

### 구조체가 큰 경우나 복사 비용이 큰 경우

#### 구조체 크기

구조체의 크기는 그 안에 *포함된 필드의 타입과 수*에 따라 달라진다.
기본적으로, 구조체에 포함된 필드가 많고, 크기가 큰 타입(예: 다른 구조체, 큰 배열, 슬라이스 등)을 포함하는 경우 구조체의 크기도 커진다.

- **작은 구조체**

    몇 개의 기본 타입(예: `int`, `float`, `bool`)만을 필드로 가지는 구조체는 일반적으로 "작은" 것으로 간주된다.
    예를 들어, 2~4개의 `int`나 `float64` 필드만을 가진 구조체는 대체로 작다고 볼 수 있다.
- **큰 구조체**

    대규모 배열, 맵, 슬라이스, 또는 여러 개의 다른 구조체를 필드로 가지는 구조체는 "큰" 구조체로 간주될 수 있다.
    예를 들어, 수십 개의 필드를 가지거나, 큰 배열을 필드로 포함하는 구조체는 복사 비용이 크다고 볼 수 있다.

##### Example

```go
type Parent struct {
    Z         []interface{}
    Child     Child
    ChildMeta ChildMeta
}

type Child struct {
    A time.Time
    B int
    C int
    D SortDirection   // custom type of string
    E SortProperty    // custom type of string
    F StatusCondition // custom type of string
    G time.Time
}

type ChildMeta struct {
    H bool
    I string
    J string
    K string
    L int32
}
```

`Parent` 구조체의 "크기"를 평가할 때 고려해야 할 몇 가지 요소가 있다.

1. **구조체 필드의 종류와 크기**

    `Parent` 구조체는 다음과 같은 필드들을 포함한다
    - 슬라이스(`Z`)
    - 두 개의 다른 구조체(`Child`, `ChildMeta`)
    - `Child`, `ChildMeta` 구조체 내의 여러 기본 타입(`time.Time`,`int`,`bool`,`string`,`int32`)과 사용자 정의 타입 필드(`SortDirection`,`SortProperty`,`StatusCondition`)

2. **동적 할당 요소**

    `Z` 필드는 `interface{}` 타입의 슬라이스로, 다양한 타입의 요소를 동적으로 저장할 수 있다.
    슬라이스와 문자열은 내부적으로 포인터를 포함하기 때문에, 슬라이스의 크기 자체는 고정적이지만, 이 슬라이스가 참조하는 데이터의 양에 따라 실제 메모리 사용량이 달라질 수 있다.

3. **사용자 정의 타입의 크기**

    `SortDirection`, `SortProperty`, `StatusCondition`과 같은 사용자 정의 타입은 기본적으로 문자열을 기반으로 한다. 문자열은 내부적으로 길이와 용량을 관리하기 위한 추가 데이터를 포함하기 때문에, 이러한 필드들 역시 실제 메모리 사용량에 영향을 준다.

`Parent` 구조체는 상당히 많은 데이터를 포함하고 있으며, 특히 슬라이스와 내포된 구조체들은 메모리 사용량을 증가시킨다.
그러나 "큰" 구조체라고 단정 지을 수 있는 명확한 기준은 없으며, 구조체가 "큰"지 여부는 그것을 어떻게 사용하느냐에 따라 달라질 수 있다. 그러나 일반적으로, 여러 복잡한 필드와 동적 데이터 구조를 포함한다는 점에서, `Parent` 구조체는 상대적으로 큰 편에 속한다고 볼 수 있다.

- **값으로 전달**

    값으로 전달할 경우, 함수 호출 시 `Parent` 구조체의 복사본이 생성된다.
    다음과 같은 경우에 유리할 수 있다.
    - 구조체가 비교적 작거나
    - 복사로 인한 오버헤드가 성능에 미치는 영향이 미미한 경우,
    - 혹은 함수에서 구조체의 불변성을 유지하고자 할 경우

    하지만 `Parent`와 같이 크기가 크고 복잡한 구조체의 경우, 값으로 전달하면 메모리 사용량이 증가하고, 복사로 인한 CPU 시간도 늘어날 수 있다.

- **포인터로 전달**

    포인터를 통해 전달하면, 구조체의 실제 데이터가 아닌 메모리 주소만 전달된다.
    이 방법은 메모리 사용량과 CPU 시간을 절약할 수 있어, 큰 구조체를 다룰 때 일반적으로 선호된다.
    또한, 포인터를 통해 전달된 구조체는 함수 내에서 변경할 수 있으며, 이 변경사항이 호출자에게 반영된다.
    하지만, 포인터를 사용할 때는 데이터의 무결성과 동시성 관리에 주의해야 한다.

`Parent` 구조체와 같이 상대적으로 크고 복잡한 구조체를 다룰 때는 **포인터로 전달하는 것이 이득**이다.
이는 메모리와 성능 최적화 측면에서 유리하며, Go의 가비지 컬렉터가 힙에 할당된 메모리를 효율적으로 관리할 수 있게 한다.
값을 변경하거나 크기가 큰 구조체를 다룰 때는 포인터 전달을 통해 성능을 개선할 수 있다.
하지만, 공유 데이터에 대한 접근을 동기화하는 방법을 고려해야 하며, 데이터 무결성을 유지하기 위한 추가적인 메커니즘이 필요할 수 있다.

#### 복사 비용

구조체의 "복사 비용":
- 메모리 상의 크기
- 구조체를 복사할 때 발생하는 CPU 시간 등

구조체가 큰 경우, 메모리에서 한 위치에서 다른 위치로 데이터를 복사하는 데 더 많은 시간이 걸린다.
이는 특히 데이터가 많거나 복잡한 프로그램에서 성능 저하의 원인이 될 수 있다.

구체적인 "크기"에 대한 기준은 없지만, 대략 1KB (1024 바이트)를 넘어가는 구조체는 크다고 생각할 수 있다.
이 기준은 매우 주관적이며, 사용하는 하드웨어, 프로그램의 요구 사항, 성능 테스트 결과 등에 따라 조정될 수 있다.

#### 결정 방법

- **성능 측정**: 실제 애플리케이션에서 성능 측정을 통해 복사 비용이 문제가 되는지 확인하는 것이 중요합니다. 프로파일링 도구를 사용하여 구조체 복사 연산이 성능 병목 현상의 원인인지 분석할 수 있습니다.
- **경험적 판단**: 애플리케이션 개발에 있어서 경험적 판단도 중요합니다. 일반적으로 구조체를 복사하는 것이 코드의 간결성과 가독성을 높이는 데 도움이 되는 경우가 많지만, 성능이 중요한 상황에서는 포인터를 사용하는 것이 더 나을 수 있습니다.

결국, "큰 구조체"의 정의와 "복사 비용이 큰 경우"에 대한 판단은 개발자가 프로그램의 특정 요구 사항과 성능 목표를 고려하여 결정해야 합니다.

### 구조체 속성과 포인터

```go
type Parent struct {
    Z         []interface{}
    Child     Child
    ChildMeta ChildMeta
}

type Child struct {
    A time.Time
    B int
    C int
    D SortDirection   // custom type of string
    E SortProperty    // custom type of string
    F StatusCondition // custom type of string
    G time.Time
}

type ChildMeta struct {
    H bool
    I string
    J string
    K string
    L int32
}
```

이떄 내부의 속성 `Child`도 포인터로 사용해야 할까?

```go
type Parent struct {
    Z         []interface{}
    Child     *Child
    ChildMeta ChildMeta
}
```

#### 1. 메모리 사용 및 할당

- **포인터 사용 (`*Child`)**

    - `Child` 필드를 포인터로 사용하면, `Child` 구조체의 인스턴스가 필요할 때마다 동적으로 할당할 수 있다.
    이는 `Child` 인스턴스가 큰 경우나, `Child`가 `Parent`에 항상 필요하지 않은 선택적(Optional) 필드일 때 유용할 수 있다. 메모리 사용량은 동적으로 관리되며, 필요할 때만 `Child` 인스턴스를 생성하게 된다.

- **값 사용 (`Child`)**

    `Child` 필드를 값으로 사용하면, `Parent` 구조체 인스턴스가 생성될 때마다 `Child`도 함께 메모리에 할당된다.
    이는 `Parent`와 `Child`가 밀접하게 연결되어 있고, 항상 함께 사용되는 경우에 적합하다.
    값 타입으로 사용하면 포인터를 따라가 메모리를 참조하는 추가적인 연산 없이 바로 `Child` 필드에 접근할 수 있으므로, 성능상의 이점을 가질 수 있다.

#### 2. 구조체의 변경 가능성

- **포인터 사용**

    `Child`를 포인터로 사용하면, 해당 포인터를 통해 `Child` 구조체의 필드를 수정할 때 원본 `Child` 인스턴스가 변경된다. 여러 `Parent` 인스턴스가 동일한 `Child` 인스턴스를 공유할 수도 있으므로, 이러한 경우 공유 상태에 대한 변경이 필요하거나 의도된 경우 포인터 사용이 유리할 수 있다.

- **값 사용**

    `Child`를 값으로 사용하면, 각 `Parent` 인스턴스는 자신만의 `Child` 복사본을 갖는다.
    이는 `Child`의 변경이 해당 `Parent` 인스턴스에만 국한되도록 하며, 다른 `Parent` 인스턴스의 `Child`에 영향을 주지 않는다. 데이터의 불변성을 유지하려는 경우에 적합한 방법이다.

#### 3. 프로그램 로직 및 설계

- **선택적 필드**
  
  만약 `Child` 필드가 선택적이어서 모든 `Parent` 인스턴스에 항상 필요하지 않다면, 포인터를 사용하는 것이 낫다.
  이를 통해 `nil`을 사용하여 `Child`가 존재하지 않음을 표현할 수 있다.
- **메모리와 성능 고려사항**

    구조체의 크기가 크고, 성능 및 메모리 사용량이 중요한 애플리케이션에서는 포인터 사용을 통한 메모리 절약이 중요할 수 있다.

## 기타

### 함수 통해서 포인터 얻는 경우 의도와 다르게 동작할 수 있다

```go
func main(){
    value := ByPointer{}
    pointer := Pointer(value)
    pointer.SetValue1("pointer of value")
    pointer.SetValue3("111", "112", "113")
    fmt.Println(value) // { 0 []}
    fmt.Println(*pointer) // {pointer of value 0 [111 112 113]}
}

func Pointer[T any](v T) *T {
    //              ^ 이미 이 시점에 `v`는 `p2`의 복사본
    // 인자로 받은 값, `p2`의 복사본인 `v`의 포인터를 반환
    return &v
}
```

`Pointer`라는 함수를 통해서 포인터를 얻고 나서 `p2`를 출력했을 때 값이 기대와 다르게 나오는 이유는 `Pointer` 함수가 동작하는 방식 때문이다. Go에서 함수 인자는 기본적으로 값으로 전달되므로, `Pointer` 함수에 `p2`를 인자로 넘기면 `p2`의 복사본이 생성되고, 그 복사본의 주소가 반환된다.
