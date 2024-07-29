# Addressable

## `주소 지정 가능`(addressable)

- `주소 지정 가능`이란?
    - 값이 메모리 주소를 가지고
    - 그 주소를 통해 해당 값에 접근할 수 있는지 여부
    - 즉, Go의 주소 연산자(`&`)를 사용하여 그 값의 메모리 주소를 얻을 수 있다
- `주소 지정 가능`한 값들은 예를 들어,
    - 변수: 모든 변수는 `addressable`하다. 변수는 메모리에 저장된 값으로, 그 주소를 통해 접근할 수 있다.
    - 구조체의 필드: 구조체의 인스턴스가 메모리에 있기 때문에, 그 필드들은 주소를 갖는다
    - 배열 요소: 배열이 메모리에 연속적으로 저장되기 때문에, 각 요소는 고유한 주소를 갖는다
    - 인덱싱된 슬라이스 요소: 슬라이스의 요소는 배열 요소처럼 `addressable`하다
- 주소 연산자(`&`)는 오직 "addressable"한 값에 대해서만 사용될 수 있다

## `주소 지정 불가능`(non-addressable)

- 상수 및 리터럴
    - 문자열
    - 숫자
    - 불리언 리터럴 등
- 함수 호출의 결과
- 산술 및 논리 연산의 결과(표현식의 결과)
- 스택에 할당된 임시 변수: 함수 내부에서 생성되는 임시 변수는 때때로 `non-addressable`할 수 있다

### `임시값`

```go
type RFC3339Time time.Time

var test *time.Time = &time.Time{}

// 새로운 인스턴스를 생성하는 것이 아니라 기존 값의 타입을 변환
failedAt = &RFC3339Time(*test)
           ^^^^^^^^^^^^^^^^^^^ Cannot take the address of 'RFC3339Time(*test)'
           RFC3339Time 타입으로 변환하는 임시 값(temporary value) 또는 중간 값(intermediate value)에 해당한다
```

- `임시값`이란 용어는 일반적으로 다음과 같은 경우를 나타내는 데 사용
    - 임시적으로 생성되는 값을 나타낼 때
    - 식(expression)의 결과를 나타낼 때
- 프로그래밍에서 이러한 값들은 일시적으로 메모리에 존재하며, 직접적인 변수 할당 없이 생성된다. 예를 들어,
    - 함수 호출의 결과
    - 계산의 결과,
    - 타입 변환의 결과 등
- 임시값은 일시적으로 생성되고, 명시적인 메모리 주소를 가지지 않는다
    - 주소 연산자를 사용할 수 없고
    - 대개 "non-addressable"로 간주된다
- `주소 지정 불가능`한 값들은 예를 들어
    - 상수: 상수는 리터럴 값이며, 메모리 주소를 갖지 않는다
    - 리터럴: 리터럴(예: `123`, `"hello"`)은 `non-addressable`하다.
    - 임시 값: 직접적인 메모리 주소를 갖지 않는다
        - 함수의 반환 값
        - 타입 변환의 결과 등

```go
// `myFunction()`의 `반환값`은 변수에 할당되기 전까지 임시값
// 따라서 `&myFunction()`과 같은 식은 유효하지 않다
result := myFunction()
```

## 기저의 원리

- `addressable`의 개념은 Go 언어의 메모리 관리 방식과 밀접한 관련이 있다.
    - Go에서 변수나 객체를 선언하면, 그것은 메모리에 저장됩이 메모리 위치는 주소로 표현되고
    - 이 주소를 통해 해당 데이터에 접근할 수 있다
- 변수와 메모리 주소
    - 메모리에 할당된 공간을 나타내며, 이 공간은 고유한 주소를 갖는다.
    - 따라서 변수는 항상 `addressable`하다.
- 임시 값과 메모리 주소
    - 타입 변환의 결과나 함수 호출의 반환 값과 같은 임시 값은 일반적으로 메모리에 명시적으로 주소가 할당되지 않는다.
    - 이러한 값은 변수에 할당되거나 다른 데이터 구조의 일부가 될 때까지 `non-addressable`한 상태로 남는다

### 메모리 관리와 안전성

Go는 효율적이고 안전한 메모리 관리를 중요하게 여깁니다. 이는 Go의 가비지 컬렉션과 엄격한 타입 시스템에서도 잘 드러납니다. `addressable` 개념은 이러한 메모리 관리 철학과 연결됩니다. 메모리에 명확한 위치를 가지고 있는 객체만이 `addressable`하다는 것은, 메모리에 안전하게 접근하고 조작할 수 있는 객체들을 명확히 구분하려는 Go의 의도를 반영합니다.

### 포인터와 간접 참조

Go는 C와 같은 언어들에서 볼 수 있는 포인터를 지원하지만, 포인터 연산(예: 포인터 산술)을 제한합니다. 이는 메모리 안전성을 높이고 프로그램의 신뢰성을 개선하기 위한 결정입니다. `addressable` 개념은 포인터를 사용할 때 메모리 안전성을 유지하는 데 핵심적인 역할을 합니다. 오직 `addressable`한 값만이 주소 연산자를 통해 포인터로 변환될 수 있으며, 이는 메모리를 직접 조작하는 대신 안전한 방법으로 데이터에 접근하도록 유도합니다.

### 타입 시스템과 엄격성

Go의 엄격한 타입 시스템은 프로그램의 안정성과 유지 보수성을 향상시키는 데 기여합니다. `addressable`과 `non-addressable`의 구분은 이러한 타입 시스템의 일부로 볼 수 있습니다. 타입 시스템은 데이터의 사용 방식을 엄격하게 규정하며, `addressable` 여부는 변수나 객체가 어떻게 메모리에 존재하고 어떻게 접근되어야 하는지를 정의합니다.

### 설계 철학: 간결성, 안전성, 효율성

Go의 설계 철학은 간결성(simplicity), 안전성(safety), 효율성(efficiency)을 중시합니다. `addressable` 개념은 이러한 철학을 반영합니다. 간결하고 명확한 규칙을 통해 어떤 객체가 메모리 주소를 가질 수 있는지를 정의함으로써, 프로그래머가 더 안전하고 효율적으로 코드를 작성할 수 있도록 합니다.

### 구현 방식

기술적으로, Go 컴파일러와 런타임은 이러한 규칙을 엄격하게 적용합니다. 컴파일 시점에 타입 검사와 함께 `addressable` 여부가 결정되며, 런타임에서의 메모리 할당 및 접근은 이러한 규칙에 따라 관리됩니다.