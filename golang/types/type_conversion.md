# Type Conversion

- [Type Conversion](#type-conversion)
    - [타입 캐스팅(Type Casting)](#타입-캐스팅type-casting)
        - [C/C++에서의 타입 캐스팅 예시](#cc에서의-타입-캐스팅-예시)
        - [Go에서의 타입 변환](#go에서의-타입-변환)
    - [타입 변환(Type Conversion)](#타입-변환type-conversion)
    - [타입 단언(Type Assertion)](#타입-단언type-assertion)
        - [포인터 타입 vs 값 타입](#포인터-타입-vs-값-타입)
        - [예시](#예시)

## 타입 캐스팅(Type Casting)

- 타입 캐스팅은 일반적으로 하위 수준의 프로그래밍 언어에서 사용되는 개념으로, **메모리에서 데이터를 해석하는 방식을 변경하는 것**을 의미
- Go에서는 타입 캐스팅이라는 용어가 공식적으로 사용되지 않으며, 타입 변환(Type Conversion)이 이에 해당한다
- C나 C++에서 볼 수 있는 타입 캐스팅은 Go에서 지원되지 않는다

C나 C++에서의 타입 캐스팅은 메모리에 저장된 데이터를 다른 타입으로 해석하는 것을 허용하는 강력하면서도 위험한 기능입니다. 이러한 캐스팅은 메모리에서 데이터를 원시적으로 재해석할 수 있게 해주며, 때로는 타입 시스템을 우회할 수 있게 해줍니다. Go에서는 이러한 접근을 지원하지 않으며, 타입 안전성을 유지하려는 설계 철학을 따릅니다.

### C/C++에서의 타입 캐스팅 예시

1. 포인터 캐스팅: 메모리의 특정 부분을 다른 타입의 포인터로 취급하는 것. 예를 들어, `int` 포인터를 `char` 포인터로 캐스팅하여 각 바이트에 접근할 수 있다.

   ```c
   int a = 1025; // `int`
   char* p = (char*)&a; // `int` 포인터를 `char` 포인터로 캐스팅
   ```

2. 재해석 캐스팅(Reinterpret Cast): 데이터의 바이트를 완전히 다른 타입으로 해석한다. 이는 메모리 레이아웃을 완전히 무시하고, 데이터를 새로운 타입으로 해석한다.

   ```cpp
   float f = 3.14;
   int* p = reinterpret_cast<int*>(&f);
   ```

3. C 스타일 캐스팅: C언어의 전통적인 캐스팅 방식으로, 다양한 캐스팅을 혼합한 형태를 가질 수 있다

   ```c
   double d = 1.234;
   int i = (int)d; // C 스타일 캐스팅
   ```

이처럼 *메모리를 임의로 재해석하는 타입 캐스팅*은 프로그램의 예측 불가능성을 증가시키고, 버그 및 보안 문제의 원인이 될 수 있다

### Go에서의 타입 변환

Go는 타입 안전성과 명확성을 중시하기 때문에, C나 C++처럼 **메모리를 직접 재해석하는 형태의 타입 캐스팅은 허용되지 않는다**.
따라서 Go에서의 타입 변환은 다음과 같은 형태로 제한된다:

1. 기본 타입 변환: 기본 타입 간의 변환은 가능하지만, ~~메모리를 재해석~~하는 것이 아니라 **값의 변환**을 의미합니다.

   ```go
   var i int = 42
   var f float64 = float64(i)
   ```

2. 사용자 정의 타입 변환: 기본 타입을 기반으로 한 사용자 정의 타입 간의 변환도 허용된다

   ```go
   type MyInt int
   var x int = 5
   var y MyInt = MyInt(x)
   ```

## 타입 변환(Type Conversion)

- 가능한 경우
    - `int`를 `float64`로 변환하는 것과 같이 호환 가능한 기본 타입 간에는 명시적인 타입 변환을 사용할 수 있다
    - 기본 타입을 기반으로 정의된 사용자 정의 타입 간에도 타입 변환을 사용할 수 있다
- 불가능한 경우
    - 서로 다른 복합 타입(예: 구조체, 사용자 정의 타입) 간의 변환은 직접적으로 허용되지 않는다.

```go
var fromTypeInstance FromType
// 타입 캐스팅(type casting) 또는 타입 변환(type conversion)이 유효하지 않기 때문에 
// "Cannot convert an expression of the type 'FromType' to the type 'ToType'" 오류
var toTypeInstance ToType = ToType(fromTypeInstance) // 이렇게 캐스팅할 수 없음
```

## 타입 단언(Type Assertion)

- 인터페이스 타입의 변수가 특정 타입의 값으로 구성되어 있는지 확인하는 과정
- 인터페이스 타입의 변수가 실제로 어떤 구체적인 타입(예: struct)의 인스턴스를 가지고 있는지 확인하는 것을 의미
- `.(ToType)` 또는 `.(*ToType)` 구문은 인터페이스 값이 `ToType` 타입(또는 `*ToType` 포인터 타입)을 가지고 있다고 단언한다
- 단언 성공과 실패 경우
    - 성공하면: 해당 타입의 값과 true를 반환
    - 실패하면: nil과 false를 반환

```go
var fromTypeInstance SomeInterface

// 타입 단언:
// - `fromTypeInstance`는 `SomeInterface` 인터페이스 타입의 변수
// - `fromTypeInstance`가 `*ToType` 포인터 타입으로 단언을 시도
// - 만약 `fromTypeInstance`가 실제로 `*ToType` 타입의 값을 가지고 있다면, `toTypePtr`는 그 값을 가지며, `ok`는 `true`가 된다
// - 만약 아니라면, `toTypePtr`는 `nil`이 되고, `ok`는 `false`가 된다
toTypePtr, ok := fromTypeInstance.(*ToType) // `fromTypeInstance`가 `*ToType` 포인터를 가지고 있다고 단언

// 타입 단언 실패: 
// - `fromTypeInstance`는 인터페이스 타입
// - `fromTypeInstance`가 구현할 수 있는 구체적인 ToType` 타입으로의 단언을 시도(`ToType` 타입의 값으로 구성되어 있는지 확인)
// - `ToType`은 `SomeInterface` 인터페이스를 만족해야 한다. 즉, `SomeInterface`의 모든 메서드를 구현해야 한다.
// - 만약 `fromTypeInstance`가 `ToType` 타입의 값을 가지고 있지 않다면, `ok`는 `false`가 되고, `toTypeVal`은 `ToType`의 zero value가 된다
toTypeVal, ok := fromTypeInstance.(ToType) //  `fromTypeInstance`가 `ToType`을 구현하고 있다고 단언

// `Impossible type assertion: 'ToType' does not implement 'SomeInterface'` 오류:
// - `fromTypeInstance.(ToType)` 단언이 유효하지 않을 때 발생
// - `ToType`이 `SomeInterface`에 정의된 모든 메서드를 구현하지 않았음을 의미
// - 즉, `ToType`은 `SomeInterface`를 구현하는 타입이 아니므로, 이 단언은 불가능
```

### 포인터 타입 vs 값 타입

1. 포인터 타입(`*ToType`)
    - `fromTypeInstance.(*ToType)`는 `fromTypeInstance`가 `*ToType` 타입, 즉 `ToType` 타입의 포인터를 가리키고 있는 경우를 단언한다
    - 이는 `fromTypeInstance`가 `ToType`의 포인터 인스턴스(`&ToType{}`)로 초기화되었다면 성공
2. 값 타입(`ToType`)
    - `fromTypeInstance.(ToType)`는 `fromTypeInstance`가 `ToType`의 값 인스턴스로 초기화되었음을 단언한다.
    - 즉, `fromTypeInstance`가 `ToType`의 값 타입(`ToType{}`)으로 초기화되었다면 성공한다

### 예시

```go
var fromTypeInstance SomeInterface = &ToType{}
_, ok := fromTypeInstance.(*ToType) // ok == true
_, ok = fromTypeInstance.(ToType)   // ok == false
```
