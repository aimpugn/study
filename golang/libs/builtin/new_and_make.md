# `new` and `make`

- [`new` and `make`](#new-and-make)
    - [`new`](#new)
        - [`new(time.Time)`와 `time.Time{}`](#newtimetime와-timetime)
            - [`time.Time{}`](#timetime)
            - [`new(time.Time)`](#newtimetime)
    - [`make`](#make)
    - [초기값과 제로 값](#초기값과-제로-값)
        - [Zero Value](#zero-value)
        - [초기화(Initialization)](#초기화initialization)
        - [차이점](#차이점)
    - [`var dao paymentContainer` vs `var dao *paymentContainer` vs `dao := new(paymentContainer)`](#var-dao-paymentcontainer-vs-var-dao-paymentcontainer-vs-dao--newpaymentcontainer)

## [`new`](https://go.dev/doc/effective_go#allocation_new)

메모리를 할당하는 내장 함수이지만 다른 언어의 이름과 달리 *메모리를 초기화하지 않고* 0으로 만들 뿐이다.
즉, `new(T)`는 `T` 타입의 새 항목에 대해 0인 저장소를 할당하고, 그 주소인 `*T` 타입의 값을 반환한다.
Go 용어로 표현하면 새로 할당된 `T` 타입의 zero 값에 대한 포인터를 반환한다.

즉, 어떤 타입 `T`의 *새로운 zero value 인스턴스에 대한 포인터*를 반환한다.

1. `new(TypeName)` 호출하면,
2. `*TypeName` 타입의 새 변수가 생성되고,
3. 해당 타입의 zero value로 초기화된 후,
4. 그 변수의 주소를 반환한다.

특정 타입의 zero value를 가진 새 변수를 초기화하면서, 동시에 그 변수의 포인터를 얻고 싶을 때 `new`를 사용한다.  
이는 특히, 구조체나 큰 배열 등을 할당하고 초기화할 때 유용할 수 있다.
포인터를 사용하면 해당 타입의 인스턴스를 여러 함수나 메서드에서 공유하고 수정할 수 있는 가능성을 제공한다.

```go
type MyStruct struct {
    Field1 int
    Field2 string
}

func main() {
    // MyStruct의 새 인스턴스를 생성하고 그 포인터를 반환합니다.
    ms := new(MyStruct)
    // 포인터를 통해 구조체의 필드에 접근할 수 있습니다.
    ms.Field1 = 10
    ms.Field2 = "Hello"
}
```

요약하면, `time.Time{}`는 `time.Time` 타입의 새 인스턴스를 바로 생성하고, `new(time.Time)`는 같은 타입의 새 인스턴스의 포인터를 반환합니다. 선택은 사용 사례에 따라 달라집니다. 직접 인스턴스를 사용해야 하는 경우 `time.Time{}`를, 포인터를 통해 인스턴스를 다루고 싶은 경우 `new(time.Time)`를 사용할 수 있습니다.

### `new(time.Time)`와 `time.Time{}`

모두 Go에서 `time.Time` 타입의 zero value를 생성하는 방법이지만, 그 결과와 사용 방식에 차이가 있다.

#### `time.Time{}`

- `time.Time{}`는 `time.Time` 타입의 zero value를 직접 생성합니다. 이 표현식은 리터럴 구문을 사용하여, 해당 타입의 새 인스턴스를 초기화하는 방법 중 하나입니다.
- 이 방식으로 생성된 객체는 스택에 할당되거나, 필요에 따라 컴파일러에 의해 힙으로 이동할 수 있습니다. 이 객체는 바로 사용할 수 있으며, `.` 연산자를 사용해 바로 필드에 접근하거나 메서드를 호출할 수 있습니다.
- 예: `var t time.Time = time.Time{}` 또는 간단히 `t := time.Time{}`

#### `new(time.Time)`

- `new(time.Time)`는 `time.Time` 타입의 새 인스턴스를 생성하고, 그 인스턴스의 포인터를 반환합니다. `new` 키워드는 모든 타입(T)에 대해 사용할 수 있으며, `*T` 타입의 새 변수를 생성하고 초기화합니다. 이 경우, 변수는 해당 타입의 zero value로 초기화됩니다.
- `new`를 사용하여 생성된 객체는 항상 힙에 할당됩니다. 이는 해당 객체를 다룰 때 포인터를 통해 접근한다는 것을 의미합니다. 따라서, 이 객체의 필드에 접근하거나 메서드를 호출할 때 포인터 연산자(`->` in C/C++ or `.` in Go with automatic dereferencing)를 사용해야 합니다.
- 예: `t := new(time.Time)`

## [`make`](https://go.dev/doc/effective_go#allocation_make)

내장 함수 `make(T, args)`는 `new(T)`와는 다른 용도로 사용된다.
이 함수는 슬라이스, 맵 및 채널만 생성하며 초기화된(zero 아닌) 유형 T(*T가 아닌)의 값을 반환합니다.

## 초기값과 제로 값

### Zero Value

Go에서 모든 변수는 선언될 때 자동으로 초기 값(zero value)을 갖는다.
이 zero value는 타입에 따라 다르며, 타입의 "빈 상태"를 의미한다.
Zero value 개념은 초기화되지 않은 변수를 사용할 때 발생할 수 있는 예기치 못한 오류를 방지하여 Go의 안전성을 높인다.

- 정수 타입(`int`, `int64` 등)의 zero value는 `0`입니다.
- 부울 타입(`bool`)의 zero value는 `false`입니다.
- 포인터(`*T`)의 zero value는 `nil`입니다.
- 슬라이스(`[]T`), 맵(`map[K]V`), 채널(`chan T`), 함수(`func`), 인터페이스(`interface{}`)의 zero value 역시 `nil`입니다.
- 문자열(`string`)의 zero value는 빈 문자열(`""`)입니다.
- 사용자 정의 구조체(`struct`)의 zero value는 모든 필드가 해당 필드 타입의 zero value를 가지는 상태입니다.

### 초기화(Initialization)

변수의 "초기화"는 개발자가 변수를 선언함과 동시에, 또는 선언 후에 명시적으로 특정 값으로 설정하는 과정을 말한다.
초기화는 zero value 또는 다른 값으로 수행될 수 있으며, 변수의 용도와 프로그램의 로직에 따라 달라집니다.

예를 들어, 다음과 같이 변수를 초기화할 수 있습니다:

```go
var number int = 10 // 명시적 초기화
var truth bool = true // 명시적 초기화

// 짧은 변수 선언을 사용한 초기화
name := "John Doe" 

// 구조체 초기화
type Person struct {
    Name string
    Age  int
}
var person Person = Person{Name: "Alice", Age: 30} // 명시적 초기화
```

### 차이점

- **Zero Value**: 모든 변수는 선언 시 자동으로 타입에 따른 zero value를 가집니다. 이는 명시적인 초기화 과정 없이도 안전하게 사용될 수 있도록 합니다.
- **초기화(Initialization)**: 변수를 특정 값으로 설정하는 과정입니다. 개발자가 명시적으로 값을 제공합니다. 초기화는 zero value로도 할 수 있지만, 주로 zero value와 다른 값을 사용하여 수행됩니다.

요약하면, zero value는 Go가 변수 선언 시 자동으로 제공하는 기본 값이며, 초기화는 개발자가 변수에 명시적으로 값을 할당하는 과정입니다. Zero value는 초기화되지 않은 변수를 안전하게 사용할 수 있게 하는 반면, 초기화 과정은 변수를 특정 값으로 시작하게 만듭니다.

## `var dao paymentContainer` vs `var dao *paymentContainer` vs `dao := new(paymentContainer)`

1. `var dao paymentContainer`: 실제 인스턴스를 선언하고 초기화
    - 이 방식은 `dao`라는 이름의 변수를 `paymentContainer` 타입으로 선언한다.
    - `paymentContainer` 타입의 **zero value로 초기화**된다. 구조체라면 모든 필드가 각 필드 타입의 zero value로 설정된다.
    - 이 선언은 포인터가 아닌, 실제 `paymentContainer` 타입의 인스턴스를 생성한다.

2. `var dao *paymentContainer`: 타입의 포인터를 선언하고 `nil`로 초기화
    - 여기서 `dao`는 `paymentContainer` 타입의 포인터로 선언된다.
    - `dao`는 `nil`로 초기화됩니다. 즉, 아직 실제 `paymentContainer` 객체를 가리키지 않는다.
    - `paymentContainer`의 인스턴스를 생성하지 않는 대신, 해당 타입의 인스턴스를 가리킬 수 있는 포인터 변수를 선언한다.

3. `dao := new(paymentContainer)`: **새로운 인스턴스**를 메모리에 할당하고 그 주소를 포인터 변수에 할당
   - `new` 키워드를 사용하여 `paymentContainer` 타입의 새로운 인스턴스를 생성하고, 이 인스턴스의 주소를 `dao`에 할당한다.
   - 결과적으로, `dao`는 `paymentContainer`의 새 인스턴스를 가리키는 포인터가 된다.
   - 이 방식은 `paymentContainer`의 새 인스턴스를 메모리에 할당하고, 그 주소를 `dao`에 저장한다.
   - 인스턴스의 모든 필드는 각 타입의 zero value로 초기화됩니다.
