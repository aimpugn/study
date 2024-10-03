# Functions

- [Functions](#functions)
    - [scoping function](#scoping-function)
        - [`run{...}` and `let{...}`](#run-and-let)
        - [`with()`](#with)
    - [`inline`, `noline`, `crossinline`, `reified`](#inline-noline-crossinline-reified)
        - [`inline`](#inline)
            - [`inline` 없는 경우](#inline-없는-경우)
            - [`inline` 있는 경우](#inline-있는-경우)
        - [`noinline`](#noinline)
        - [`crossinline`](#crossinline)
        - [`reified`](#reified)
    - [`inline` \& Expected performance impact from inlining is insignificant](#inline--expected-performance-impact-from-inlining-is-insignificant)

## scoping function

### `run{...}` and `let{...}`

- [Example of when should we use run, let, apply, also and with on Kotlin](https://stackoverflow.com/questions/45977011/example-of-when-should-we-use-run-let-apply-also-and-with-on-kotlin)
- [Mastering Kotlin standard functions: run, with, let, also and apply](https://medium.com/mobile-app-development-publication/mastering-kotlin-standard-functions-run-with-let-also-and-apply-9cd334b0ef84)

why scoping function? to **provide an inner scope** for the caller function.

```kotlin
fun test() {
    var mood = "I am sad"
    run {
        val mood = "I am happy"
        println(mood) // I am happy
    }
    println(mood)  // I am sad
}
```

The `T.let{}` does provide a clearer distinguish use the given variable function/member vs. the external class function/member

### `with()`

> The **context object** is available as a receiver (`this`).
> The **return value** is the lambda result.

## `inline`, `noline`, `crossinline`, `reified`

- [Kotlin’s Noinline & Crossline, once for all](https://ncorti.com/blog/noinline-and-crossline-once-for-all)
- [`inline`, `noline`, `crossinline`, `reified`에 대하여](https://leveloper.tistory.com/171)
- [Kotlin-inline, crossline, noinline function and reified: Everything you need to know(Android) — Part 1](https://ashiqulislamshaon.medium.com/kotlin-inline-crossline-noinline-function-and-reified-everything-you-need-to-know-android-3743dafa476)

### [`inline`](https://kotlinlang.org/docs/inline-functions.html)

> [고차 함수(higher-order function)](https://kotlinlang.org/docs/lambdas.html)를 사용하면 *런타임 패널티*가 있기 때문에 함수 구현 자체를 코드에 넣음으로써 오버헤드를 없앨 수 있다.

- *런타임 패널티* -> 각 함수는 오브젝트고, 이는 클로저를 캡처한다

#### `inline` 없는 경우

```kotlin
fun doSomething(action: () -> Unit) {
    action()
}

fun callFunc() {
    doSomething {
        println("문자열 출력!")
    }
}
```

```java
public void doSomething(Function action) {
    action.invoke();
}

public void callFunc() {
    doSomething(System.out.println("문자열 출력!");
}
```

```java
public void callFunc() {
    // Function 인스턴스 생성 && 함수 호출 -> 오버헤드 생길 수 있다
    doSomething(new Function() {
        @Override
        public void invoke() {
            System.out.println("문자열 출력!");
        }
    }
}
```

#### `inline` 있는 경우

```kotlin
inline fun doSomething(action: () -> Unit) {
    action()
}

fun callFunc() {
    doSomething {
        println("문자열 출력!")
    }
}
```

```java
public void callFunc() {
    // Function 인스턴스 생성하지 않고, `callFunc` 내부에 삽입되어 바로 선언된다(inline)
    System.out.println("문자열 출력!");
}
```

내부적으로 코드를 복사하기 때문에, 인자로 전달 받은 함수는 다른 함수로 전달되거나 참조할 수 없다

### `noinline`

모든 함수를 `inince`으로 처리하고 싶지 않을 때, `noinline` 키워드 붙이면 `inline`에서 제외된다

```kotlin
inline fun doSomething(action1: () -> Unit, noinline action2: () -> Unit) {
    action1()
    anotherFunc(action2) // `inline` 되지 않으므로, 인자로 전달할 수 있다
}

fun anotherFunc(action: () -> Unit) {
    action()
}

fun main() {
    doSomething({
        println("1")
    }, {
        println("2")
    })
}
```

### `crossinline`

```kotlin
inline fun View.click(block: (View) -> Unit) {
    // 다른 실행 컨텍스트 통해서 실행 시,
    // 함수 안에서 non local 흐름을 제어할 수 없다
    setOnClickListener { view ->
        block(view) // error
    }
}
```

```kotlin
inline fun View.click(crossinline block: (View) -> Unit) {
    setOnClickListener { view ->
        block(view) // work
    }
}
```

### `reified`

- [Kotlin - Reified를 사용하는 이유?](https://codechacha.com/ko/kotlin-reified-keyword/)

1. 컴파일 시에는 Generic의 타입을 알고 있음
2. 컴파일 후에는 타입 정보를 제거하여 그저 T로 정해진 객체가 존재할 뿐, Runtime 시에는 타입 알 수 없음

`inline` 함수에서 사용되며, Runtime에 타입 정보를 알고 싶을 때 사용

```kotlin
// compile time에 compiler는 T가 어떤 타입인지 알고 있음
// 컴파일하면서 타입 정보를 제거하여 runtime 시에 T가 어떤 타입인지 모름
// 그냥 T로 정해진 객체가 존재할 뿐
fun <T> function(argument: T)
```

`reified` 키워드 사용하면 Generics function에서 런타임에 타입 정보 알 수 있음.
`reified` 없이 런타임에 타입 정보 알고 싶은 경우 아래와 같은 방법이 가능하다

```kotlin
/*
!에러 발생 코드
fun <T> printGenerics(value: T) {
    ^^^: T라는 객체가 있을 뿐, 실제 어떤 타입인지는 런타임에 알 수 없다
    when (value::class) {  // compile error! 런타임에 타입 정보가 지워지기 때문
        String::class.java -> {
            println("String : $value")
        }
        Int::class.java -> {
            println("Integer : $value")
        }
    }
}
*/

fun <T> printGenericsWithClassType(value: T, classType: Class<T>) {
    when (classType) {
        String::class.java -> {
            println("String : $value")
        }
        Int::class.java -> {
            println("Int : $value")
        }
    }
}

inline fun <reified T> printGenericsWithReified(value: T) {
    when (T::class) {
        String::class -> {
            println("String : $value")
        }
        Int::class -> {
            println("Int : $value")
        }
    }
}
```

```kotlin
// `Unchecked cast: View! as T`: inline 함수에서 특정 타입 가졌는지 판단할 수 없기 때문
inline fun <T: View> T.click(crossinline block: (T) -> Unit) {
    setOnClickListener { view ->
        block(view as T) 
    }
}
```

## `inline` & Expected performance impact from inlining is insignificant

```log
Warning:(98, 13) Expected performance impact from inlining is insignificant. Inlining works best for functions with parameters of functional types
```
