# Class

- [Class](#class)
    - [`init` block](#init-block)
    - [공통부와 특정 서브 클래스를 제네릭으로 나누기](#공통부와-특정-서브-클래스를-제네릭으로-나누기)
    - [sealed class, sealed interface](#sealed-class-sealed-interface)
        - [언제 사용?](#언제-사용)
        - [직렬화, 역직렬화](#직렬화-역직렬화)
    - [generics](#generics)
        - [Declaration-site variance](#declaration-site-variance)
            - [`out`](#out)
            - [`in`](#in)
    - [enum](#enum)
        - [역순 탐색](#역순-탐색)
    - [reflection](#reflection)
        - [`isInstance`](#isinstance)
    - [data class](#data-class)
        - [copy](#copy)

## `init` block

- [What is the difference between init block and constructor in kotlin?](https://stackoverflow.com/questions/55356837/what-is-the-difference-between-init-block-and-constructor-in-kotlin)
    - `init` 블록은 primary constructor 직후 실행되며, initializer 블록은 primary constructor의 일부가 된다
    - `constructor`는 보조 생성자

```kotlin
class Sample(private var s : String) { // primary constructor
    init { // initializer
        s += "B"
    }
    constructor(t: String, u: String) : this(t) { // secondary constructor
        this.s += u
    }
}

Sample("T","U") // "TBU"
```

## 공통부와 특정 서브 클래스를 제네릭으로 나누기

```kotlin
interface SomeClient : Client {
  companion object Dto {
    interface SomeResponseWrapper<T> {
      val code: Code
      val message: String
      val data: T?
    }

    data class SomeSpecificResponse(
      override val code: Code,
      override val message: String,
      override val data: Data
    ) : SomeResponseWrapper<SomeSpecificResponse.Data> {
      data class Data(
        val field1: String,
        val field2: String,
        val field3: String,
      )
    }
  }
}
```

## sealed class, sealed interface

- [Sealed classes](https://kotlinlang.org/docs/sealed-classes.html)
- [Enum vs Sealed class — which one to choose?](https://blog.kotlin-academy.com/enum-vs-sealed-class-which-one-to-choose-dc92ce7a4df5)
- [Sealed interfaces in Kotlin](https://jorgecastillo.dev/sealed-interfaces-kotlin)
- [Effective Kotlin Item 39: Use sealed classes and interfaces to express restricted hierarchies](https://kt.academy/article/ek-sealed-classes)
- [Sealed Class vs Enum in Kotlin](https://www.baeldung.com/kotlin/sealed-class-vs-enum)

### 언제 사용?

- 상수를 타입처럼 관리하고 싶을 때
- 해당 타입의 인스턴스에서 데이터를 가지고 있어야 할 때

### 직렬화, 역직렬화

- [How to serialize/deserialize Kotlin sealed class?](https://stackoverflow.com/questions/50157468/how-to-serialize-deserialize-kotlin-sealed-class)
- [Kotlin with Jackson: Deserializing Kotlin Sealed Classes](https://serpro69.medium.com/kotlin-with-jackson-deserializing-kotlin-sealed-classes-c95f837e9164)

## generics

- [Generics: in, out, where](https://kotlinlang.org/docs/generics.html)
- [Java Generics FAQs](http://www.angelikalanger.com/GenericsFAQ/JavaGenericsFAQ.html)

### Declaration-site variance

- [Declaration-site variance](https://kotlinlang.org/docs/generics.html#declaration-site-variance)
- [코틀린 제네릭, in? out?](https://medium.com/mj-studio/%EC%BD%94%ED%8B%80%EB%A6%B0-%EC%A0%9C%EB%84%A4%EB%A6%AD-in-out-3b809869610e)

#### `out`

```kotlin
interface Source<out T> {
    fun nextT(): T
}

fun demo(strs: Source<String>) {
    val objects: Source<Any> = strs // This is OK, since T is an out-parameter
    // ...
}
```

- `Source`의 타입 파라미터 `T`에 주석을 달아서 `Source<T>`의 멤버로부터 반환(produced) 되기만 하고 소비(consumed) 되지 않는다는 것을 보장할 수 있다. 이떄 `out` 수식어를 사용
- `C` 클래스의 타입 파라미터 `T`가 `out`으로 선언되면, `C` 멤버의 out 위치에서만 발생할 수 있지만, 그 대가로(in return) `C<Base>`는 안전하게 `C<Derived>`의 슈퍼타입이 될 수 있다.
- 다르게 말하면 다음과 같이 말할 수 있다
    - `C` 클래스는 `T` 파라미에서 [공변적(covariant)](https://learn.microsoft.com/ko-kr/dotnet/standard/generics/covariance-and-contravariance)이다
    - 또는 그 `T` 파라미터는 공변적인(covariant) 타입 파라미터다.
- `C`를 `T`의 producer로 생각할 수 있지만, `T`의 consumer라고는 생각하지 않을 것이다.

#### `in`

```kotlin
interface Comparable<in T> {
    operator fun compareTo(other: T): Int
}

fun demo(x: Comparable<Number>) {
    x.compareTo(1.0) // 1.0 has type Double, which is a subtype of Number
    // Thus, you can assign x to a variable of type Comparable<Double>
    val y: Comparable<Double> = x // OK!
}
```

## enum

### 역순 탐색

- [Effective Enums in Kotlin with reverse lookup?](https://stackoverflow.com/questions/37794850/effective-enums-in-kotlin-with-reverse-lookup)

## reflection

### `isInstance`

- Returns true if [value] is an instance of this class on a given platform.

```kotlin
String::class.isInstance("a string")  /// true
Int::class.isInstance(42) /// true

// 1::class.isInstance(kotlin.Int) // X
// true::class.isInstance(Boolean) // X
```

## [data class](https://kotlinlang.org/docs/data-classes.html)

### copy

- 특정 값만 바꾸고 나머지는 복사할 수 있는 함수
- 자바 기반 언어에서 primitive type 들 제외하고는 전부 reference copy
- api call 시 연산에 민감하지 않은 서버 특성상 copy 퍼포먼스에 대한 걱정은 하지 않아도 될 거 같다
