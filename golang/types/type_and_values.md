# Type and values

- [Type and values](#type-and-values)
    - [method set](#method-set)
        - [The Go Programming Language Specification](#the-go-programming-language-specification)
        - [A Tour of Go](#a-tour-of-go)
    - [인터페이스와 리시버](#인터페이스와-리시버)
        - [값 타입 리시버로 구현된 메서드](#값-타입-리시버로-구현된-메서드)
        - [포인터 리시버로 구현된 메서드](#포인터-리시버로-구현된-메서드)
        - [결론](#결론)

## [method set](https://go.dev/ref/spec#Method_sets)

타입의 메서드 집합은 해당 타입의 피연산자에서 호출할 수 있는 메서드를 결정합니다.
모든 타입에는 연결된 (비어 있을 수 있는) 메서드 집합이 있습니다.

- [정의된 타입](https://go.dev/ref/spec#Type_definitions) T의 메서드 집합은 수신자 타입 T로 선언된 모든 [메서드](https://go.dev/ref/spec#Method_declarations)로 구성됩니다.
- 정의된 타입 T에 대한 포인터의 메서드 집합(여기서 T는 포인터도 인터페이스도 아님)은 *리시버*가 `*T` 또는 `T`로 선언된 모든 메서드의 집합입니다.
- [인터페이스 타입](https://go.dev/ref/spec#Interface_types)의 메서드 집합은 인터페이스의 [타입 집합](https://go.dev/ref/spec#Interface_types)에 있는 각 타입의 메서드 집합의 교차점입니다(결과 메서드 집합은 일반적으로 인터페이스에 선언된 메서드의 집합일 뿐입니다).

Go 언어에서 인터페이스와 포인터 리시버와 관련된 문서는 공식 Go 언어 문서에서 찾을 수 있습니다. 특히, 이와 관련된 내용은 **The Go Programming Language Specification** 및 **A Tour of Go**에서 자세히 설명되어 있습니다. 여기서는 메서드 세트와 인터페이스의 구현 방식에 대해 다루며, 포인터 리시버를 사용할 경우 해당 타입의 포인터만이 인터페이스를 구현한다는 규칙을 설명합니다.

### The Go Programming Language Specification

[The Go Programming Language Specification](https://golang.org/ref/spec#Method_sets)에서는 메서드 세트에 대한 정의와 함께, 타입이 어떻게 인터페이스를 구현하는지에 대한 규칙을 설명합니다. 특히, 타입 `T`의 포인터 `*T`가 메서드를 가질 때, 그 메서드 세트가 `*T`에 속하고, 이는 `T`가 아닌 `*T`에 의해서만 인터페이스를 구현할 수 있음을 의미합니다.

### A Tour of Go

[A Tour of Go](https://tour.golang.org/methods/4)에서는 포인터 리시버를 사용하여 메서드를 선언하는 방법과 그 이유에 대해 설명합니다. 이 섹션에서는 메서드가 리시버의 실제 값을 수정해야 하거나, 메서드 호출이 리시버의 복사본을 생성하지 않도록 하기 위해 포인터 리시버를 사용하는 경우를 다룹니다. 또한, 포인터 리시버를 사용함으로써 해당 타입의 포인터만이 메서드 세트를 포함하게 되어, 이 타입의 포인터만이 인터페이스를 만족시킬 수 있음을 설명합니다.

## 인터페이스와 리시버

인터페이스는 메서드 시그니처의 집합으로 정의되며, 특정 타입이 인터페이스가 요구하는 모든 메서드를 구현하는 경우, 그 타입의 인스턴스(포인터 포함)는 해당 인터페이스를 구현한다고 간주된다.

포인터를 사용해야 하는 경우와 그렇지 않은 경우는 주로 *해당 타입의 메서드가 리시버(receiver)로 포인터를 요구하는지, 아니면 값 타입을 요구하는지에 따라* 달라진다.

### 값 타입 리시버로 구현된 메서드

타입의 메서드가 값 리시버를 사용하여 구현된 경우, 그 타입의 인스턴스와 포인터 모두 해당 인터페이스를 구현합니다. 예를 들어, 다음과 같이 값 리시버를 사용한 메서드가 있을 때:

```go
type MyInterface interface {
    DoSomething()
}

type MyType struct{}

func (m MyType) DoSomething() {
    // 구현
}

func main() {
    var i MyInterface = MyType{} // 값 타입 인스턴스로 인터페이스 구현
    var j MyInterface = &MyType{} // 포인터로도 인터페이스 구현
}
```

### 포인터 리시버로 구현된 메서드

메서드가 포인터 리시버를 사용하여 구현된 경우, 해당 타입의 포인터만이 인터페이스를 구현합니다. 값 타입 인스턴스는 인터페이스를 구현하지 않습니다. 예를 들어:

```go
type MyInterface interface {
    DoSomething()
}

type MyType struct{}

func (m *MyType) DoSomething() {
    // 구현
}

func main() {
    // var i MyInterface = MyType{} // 컴파일 에러: MyType 값 타입은 MyInterface를 구현하지 않음
    var j MyInterface = &MyType{} // 올바름: MyType 포인터는 MyInterface를 구현
}
```

### 결론

따라서 `responseOrganizer`의 포인터를 `Organizer` 인터페이스 타입 변수에 할당해야 하는 이유는 `responseOrganizer` 타입이 인터페이스가 요구하는 메서드를 포인터 리시버로 구현했기 때문일 가능성이 높습니다. 만약 `responseOrganizer`가 인터페이스의 메서드를 값 리시버로 구현했다면, 값 타입 인스턴스도 인터페이스 타입 변수에 할당할 수 있습니다.
