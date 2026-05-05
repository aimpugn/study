# Functional Programming

> 원문 보존형 이동본입니다. 이 파일의 source chunk 본문은 원본 `intervie*.md`에서 그대로 복사되었고, 기술적 보강과 딥 리라이트는 다음 단계에서 수행합니다.

## Source Chunks

<!-- source-chunk: sha256=7c5b45abb7d6106d2c28451a32820c77390e1b67a1411b1028aee41bbbd90e87 topic=functional-programming sources=interview_questions2.md:1743-2120, interviews2.md:1743-2120 -->

> Source: `interview_questions2.md:1743-2120`
> Duplicate source aliases: `interview_questions2.md:1743-2120, interviews2.md:1743-2120`

## 함수형

함수형 프로그래밍(Functional Programming, FP)은 수학적 함수의 개념에서 영감을 받은 프로그래밍 패러다임으로, 코드의 명확성, 재사용성, 안전성을 강조합니다.

순수 함수(Pure Function), 불변성(Immutability), 고차 함수(Higher-Order Functions), 일급 객체(First-Class Citizens) 등의 원칙을 중심으로 하는 프로그래밍 패러다임입니다.

1. 순수 함수(Pure Function):
    - 동일한 입력에 대해 항상 동일한 결과를 반환하며, 외부 상태를 변경하지 않음.
    - 부수효과(Side Effects)를 피함으로써 프로그램의 예측 가능성을 증가시킴.

2. 불변성(Immutability):
    - 데이터와 상태는 변경할 수 없으며, 변경하려면 복사본을 생성.
    - 이는 데이터 경쟁(Data Race)과 같은 멀티스레드 이슈를 방지.

3. 고차 함수(Higher-Order Function):
    - 함수를 값처럼 취급하며, 이를 다른 함수의 인자로 전달하거나 반환값으로 사용할 수 있음.

4. 일급 객체(First-Class Citizen):
    - 함수가 변수처럼 할당되고, 다른 함수의 인자나 반환값으로 사용될 수 있음.

5. 참조 투명성(Referential Transparency):
    - 표현식이 항상 동일한 값을 반환하여, 호출을 그 결과값으로 대체해도 프로그램의 의미가 유지됨.

6. 부수효과의 최소화(Side Effects Minimization):
    - 함수의 결과에 영향을 미치지 않는 외부 상태 변경을 최소화.

### 함수형 프로그래밍의 이론적 개념

- Functor
    - 컨테이너 안의 값을 매핑(map)하는 추상화.
    - 값을 직접 변경하지 않고, 컨테이너의 구조를 유지하면서 값을 변환.

    이론적 정의:
    - `map` 연산을 지원하며, 다음 법칙을 만족:
        - 항등 법칙: `F.map(id) == F`
        - 합성 법칙: `F.map(f).map(g) == F.map(f ∘ g)`

- Monad
    - 연속적인 계산을 모델링하는 추상화.
    - Functor를 확장하여, 컨테이너 안의 값을 연결(bind)하고 중첩을 평평하게(flatten) 만듦.

    이론적 정의:
    - `flatMap`(또는 `bind`) 연산을 지원하며, 다음 법칙을 만족:
        - 좌항등 법칙: `M.pure(a).flatMap(f) == f(a)`
        - 우항등 법칙: `M.flatMap(M.pure) == M`
        - 결합 법칙: `M.flatMap(f).flatMap(g) == M.flatMap { x -> f(x).flatMap(g) }`

- 부수효과의 제어
    - Effect Handling: 부수효과를 함수형으로 모델링하여 안전하게 관리.
        - 예: I/O 작업, 네트워크 요청.

### 3.6 Either를 통한 에러 처리

#### 이론 적용

`Either`는 성공과 실패를 명시적으로 처리하는 함수형 대안입니다.

#### 실무 예제

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

fun divide(a: Int, b: Int): Either<String, Int> =
    if (b == 0) "Division by zero".left()
    else (a / b).right()

val result = divide(10, 2) // Right(5)
val error = divide(10, 0) // Left("Division by zero")
```

### 3.7 I/O 작업 제어

#### 이론 적용

부수효과를 함수형으로 모델링하여 안전하게 처리.

#### Arrow-kt에서의 I/O 모델링

```kotlin
import arrow.fx.IO

val program = IO {
    println("Performing side-effect")
    "Result"
}

val result = program.unsafeRunSync() // "Performing side-effect" 출력
```

### 순수 함수 (Pure Function)

동일한 입력에 대해 항상 동일한 출력을 반환하며, 외부 상태를 변경하지 않는 함수.

```kotlin
// 순수 함수는 외부 상태에 의존하지 않으며, 동일한 입력에 대해 항상 동일한 결과를 반환합니다.
fun calculateDiscount(price: Double, discountRate: Double): Double =
    price * discountRate

// 외부 상태에 의존하지 않음
val result = calculateDiscount(100.0, 0.1) // 항상 10.0 반환

fun add(a: Int, b: Int): Int = a + b
```

### 불변성 (Immutability)

함수형 프로그래밍에서는 상태를 변경하지 않고 새로운 값을 생성합니다.

```kotlin
// 데이터를 변경하지 않고, 복사본을 생성하여 상태 변경을 구현.
val originalList = listOf(1, 2, 3)
val updatedList = originalList + 4 // [1, 2, 3, 4]

// 원본은 변경되지 않음
println(originalList) // [1, 2, 3]

val list = listOf(1, 2, 3)
val newList = list.map { it * 2 } // 원래 list는 변경되지 않음
```

### 고차 함수 (Higher-Order Functions)

다른 함수를 인자로 받거나 반환하는 함수.

```kotlin
// 고차 함수는 함수를 인자로 받거나 반환값으로 사용하여 동작을 추상화합니다.
fun applyOperation(numbers: List<Int>, operation: (Int) -> Int): List<Int> {
    return numbers.map(operation)
}

// 두 배로 만드는 함수 전달
val doubled = applyOperation(listOf(1, 2, 3)) { it * 2 } // [2, 4, 6]

fun applyTwice(f: (Int) -> Int, x: Int): Int = f(f(x))

val result = applyTwice({ it * 2 }, 5) // 20
```

### 일급 객체 (First-Class Citizens)

함수가 변수에 할당되거나, 다른 함수의 인자로 전달될 수 있음.

```kotlin
val double: (Int) -> Int = { it * 2 }
println(double(4)) // 8
```

### Functor

컨테이너 안의 값을 매핑할 수 있는 추상화.

Kotlin의 `map` 함수는 Functor의 예입니다.

```kotlin
val list = listOf(1, 2, 3)
val doubled = list.map { it * 2 } // [2, 4, 6]
```

Arrow는 `Option`, `Either`와 같은 컨테이너 타입에 Functor 연산을 제공합니다.

```kotlin
import arrow.core.Option
import arrow.core.some

val option: Option<Int> = 3.some()
val result = option.map { it * 2 } // Option(6)
```

### Monad

컨테이너 안의 값을 매핑할 뿐 아니라 컨테이너의 중첩을 평평하게(flatten) 하는 추상화.
`bind` 또는 `flatMap`을 통해 중첩된 컨테이너를 단일 컨테이너로 변환합니다.

```kotlin
// 컨테이너 안의 중첩된 값을 평평하게(flatten) 하여 연결.
val nested = listOf(listOf(1, 2), listOf(3, 4))
val flattened = nested.flatMap { it } // [1, 2, 3, 4]
```

Arrow-kt에서의 Monad:

```kotlin
import arrow.core.Option
import arrow.core.some

val opt1: Option<Int> = 10.some()
val result = opt1.flatMap { x -> Option.just(x * 2) } // Option(20)
```

```kotlin
import arrow.core.Option
import arrow.core.none
import arrow.core.some

val opt1: Option<Int> = 10.some()
val opt2: Option<Int> = none()

val result = opt1.flatMap { x -> opt2.map { y -> x + y } } // None
```

### Task (Deferred in Kotlin)

비동기 작업을 표현하며, 결과를 지연 평가(Lazy Evaluation)로 처리.

Kotlin의 예:

```kotlin
import kotlinx.coroutines.*

suspend fun fetchUser(): String {
    delay(1000) // Simulating long-running task
    return "User"
}

fun main() = runBlocking {
    val task = async { fetchUser() }
    println(task.await()) // "User"
}
```

Arrow의 `IO` 모나드 또는 `Task`는 비동기 작업을 함수형 스타일로 처리하는 데 사용됩니다.

```kotlin
import arrow.fx.coroutines.*

suspend fun main() {
    val task = IO.invoke { "Result" }
    println(task.unsafeRunSync()) // "Result"
}
```

### Either

실패 또는 성공을 표현하는 양자 타입.
`Either.Left`는 오류를, `Either.Right`는 성공 값을 나타냅니다.

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

fun divide(a: Int, b: Int): Either<String, Int> =
    if (b == 0) "Division by zero".left()
    else (a / b).right()

val result = divide(10, 2)
println(result) // Right(5)
```

`bind()`는 Arrow의 `Monad` 컴퓨테이션을 처리하는 DSL(do notation)에서 사용되는 메서드로, Monad 컨텍스트에서 값을 꺼내고 연속적인 계산을 처리합니다.
Monad 컨텍스트란, 예를 들어, `Either`, `Option`, `IO`와 같은 Monad 타입에서 값을 직접 꺼낼 수 없는 상황에서, `bind()`를 통해 안전하게 값을 꺼내 계산을 이어갑니다.

`Either`는 성공(`Right`)과 실패(`Left`)를 표현하는 타입입니다.

```kotlin
import arrow.core.Either
import arrow.core.continuations.either
import arrow.core.left
import arrow.core.right

fun fetchUserData(userId: String): Either<String, String> =
    if (userId == "123") "User data".right() else "User not found".left()

fun fetchUserPosts(userId: String): Either<String, List<String>> =
    if (userId == "123") listOf("Post1", "Post2").right() else "No posts".left()

suspend fun fetchUserWithPosts(userId: String): Either<String, Pair<String, List<String>>> = either {
    val userData = fetchUserData(userId).bind() // 값을 안전하게 꺼냄
                                                // `bind()`는 Monad(`Either`) 내부에서 값을 꺼내 다음 계산에 전달.
                                                // 만약 `Left`가 발생하면 즉시 실행을 중단하고 에러를 반환.
    val userPosts = fetchUserPosts(userId).bind()
    userData to userPosts
}

fun main() {
    val result = fetchUserWithPosts("123")
    println(result) // Right((User data, [Post1, Post2]))
}
```

Arrow를 사용하여 HTTP 통신 결과를 처리하는 경우도 유사합니다.
예를 들어, HTTP 요청이 `Either` 타입으로 성공과 실패를 반환할 때:

```kotlin
import arrow.core.Either
import arrow.core.continuations.either

suspend fun fetchHttpResponse(url: String): Either<String, String> =
    if (url.startsWith("http")) "Response from $url".right()
    else "Invalid URL".left()

suspend fun fetchData(): Either<String, String> = either {
    val response = fetchHttpResponse("http://example.com").bind() // HTTP 응답을 처리
                                                                  // `bind()`는 성공(`Right`) 값을 꺼내 계산을 이어가며, 실패(`Left`)가 발생하면 즉시 반환.
    "Processed: $response"
}

suspend fun main() {
    val result = fetchData()
    println(result) // Right(Processed: Response from http://example.com)
}
```

`bind()`가 HTTP 통신에서 사용된다면, 이는 Arrow와 같은 함수형 라이브러리를 활용하여 안전한 에러 처리와 Monad 값을 관리하는 방식으로 구현되었을 가능성이 큽니다. 이 방식은 다음과 같은 이점을 제공합니다:

1. 명시적 에러 처리:
    - `Either`와 같은 Monad 타입을 통해, 성공과 실패를 명확히 처리.

2. 안전한 값 추출:
    - `bind()`는 Monad 내부 값을 안전하게 꺼내고, 실패가 발생하면 자동으로 에러를 반환.

3. 연속적인 계산 모델링:
    - 여러 HTTP 요청이 연속적으로 연결될 때, `bind()`를 사용해 간결하고 명확하게 코드 작성 가능.

### Arrow-kt 주요 데이터 타입 및 함수

Arrow-kt는 함수형 프로그래밍의 개념을 구현하는 여러 데이터 타입과 도구를 제공합니다.

- `Option`
    - 정의: 값이 있을 수도, 없을 수도 있는 경우를 표현.
    - `Option.Some`은 값을 나타내고, `Option.None`은 값이 없음을 나타냅니다.

    ```kotlin
    import arrow.core.Option
    import arrow.core.some
    import arrow.core.none

    val someValue: Option<Int> = 5.some()
    val noValue: Option<Int> = none()

    val result = someValue.getOrElse { 0 } // 5
    val fallback = noValue.getOrElse { 0 } // 0
    ```

- `Validated`

    실패 또는 성공을 표현하며, 실패를 누적할 수 있음.

    ```kotlin
    import arrow.core.Validated
    import arrow.core.invalid
    import arrow.core.valid

    val success = 42.valid()
    val failure = "Error".invalid()

    println(success) // Valid(42)
    println(failure) // Invalid(Error)
    ```

- `IO`

    효과(effect)를 표현하며, 부수 효과(side-effect)를 안전하게 처리.

    ```kotlin
    import arrow.fx.IO

    val io = IO { println("Running IO") }
    io.unsafeRunSync() // 실제로 실행
    ```

- Applicative

    독립적인 계산을 결합하여 병렬 실행 지원.

    ```kotlin
    import arrow.core.Tuple2
    import arrow.core.extensions.list.apply.map

    val listA = listOf(1, 2, 3)
    val listB = listOf("A", "B", "C")

    val combined = listA.map(listB) { a, b -> "$a$b" }
    println(combined) // [1A, 1B, 1C, 2A, 2B, 2C, 3A, 3B, 3C]
    ```

<!-- /source-chunk -->
