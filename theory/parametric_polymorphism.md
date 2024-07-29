# Parametric polymorphism

- [Parametric polymorphism](#parametric-polymorphism)
    - [매개변수 다형성(Parametric polymorphism이란)?](#매개변수-다형성parametric-polymorphism이란)
    - [기본 정의](#기본-정의)
    - [기타 예시](#기타-예시)
    - [바운드(bounded) 타입](#바운드bounded-타입)
        - [타입 클래스와 트레잇](#타입-클래스와-트레잇)
        - [타입 클래스 (Haskell 예시)](#타입-클래스-haskell-예시)
        - [트레잇 (Rust 예시)](#트레잇-rust-예시)
        - [트레잇(scala)](#트레잇scala)
        - [타입 제한](#타입-제한)
    - [함수형에서의 `trait`와 PHP에서의 `trait`](#함수형에서의-trait와-php에서의-trait)
        - [트레잇의 개념 유래](#트레잇의-개념-유래)
        - [vs PHP의 `trait`](#vs-php의-trait)
        - [명명의 이유](#명명의-이유)

## 매개변수 다형성(Parametric polymorphism이란)?

매개변수 다형성(Parametric polymorphism)이라는 개념은 단일 코드 조각에 "일반적인"(제네릭) 타입을 사용할 수 있게 하고, 실제 타입 대신 변수를 사용한 다음 필요에 따라 특정 타입으로 구체화할 수 있도록 합니다.

이런 종류의 *다형성을 갖는 함수나 데이터 타입*을 때때로 일반적인 함수(generic functions)와 일반적인 데이터 타입(generic data types)이라고 부르며, 제네릭 프로그래밍의 기초를 형성합니다.

Parametric polymorphism은 ad hoc polymorphism과 대비될 수 있습니다.
- 매개변수 다형성의 정의는 인스턴스화된 타입에 관계없이 동일하게 동작하는 특징이 있습니다.
- 반면, ad hoc polymorphic 정의는 각 타입에 대해 별도의 정의를 제공합니다.

따라서 ad hoc polymorphism은 일반적으로 지원하는 타입의 수가 제한적이며, 각 타입에 대한 별도의 구현을 제공해야 합니다.

## 기본 정의

타입에 의존하지 않는 함수를 작성할 수 있습니다.

예를 들어, 신원 함수(identity function)는 인자를 수정하지 않고 그대로 반환합니다.

$${\mathsf {id}}(x)=x$$

이것은 다음과 같은 여러 가능한 타입을 자연스럽게 만들어냅니다. 예를 들어:
- `Int -> Int`,
- `Bool -> Bool`,
- `String -> String`

매개변수 다형성은 타입 변수를 사용하여 가장 일반적인 타입을 소개함으로써 `id`에 단일 타입을 제공할 수 있게 합니다.
이 다형성 정의는 모든 구체적인 타입을 대체함으로써 가능한 전체 타입 패밀리를 생성할 수 있습니다.

```go
package main

import (
    "fmt"
)

// Identity 함수는 매개변수 T의 어떤 타입에도 동작합니다.
func Identity[T any](x T) T {
    return x
}

func main() {
    // Int 타입에 대한 Identity 함수 사용
    fmt.Println(Identity[int](42))

    // Bool 타입에 대한 Identity 함수 사용
    fmt.Println(Identity[bool](true))

    // String 타입에 대한 Identity 함수 사용
    fmt.Println(Identity[string]("Hello, World!"))
}
```

## 기타 예시

신원 함수 외에도, 많은 다른 함수들이 parametric polymorphism의 혜택을 받을 수 있습니다.

예를 들어, 두 리스트를 연결하는 `append` 함수는 리스트의 구조만 검사하고 요소를 검사하지 않습니다.

${\mathsf {append}}:\forall \alpha .[\alpha ]\times [\alpha ]\to [\alpha ]$

따라서 `append`는 아래와 같은 유사한 타입 패밀리를 갖게 됩니다.
- `Int` 리스트를 연결하는 `[Int] x [Int] -> [Int]`
- `Bool` 리스트를 연결하는 `[Bool] x [Bool] -> [Bool]`  

가장 일반적인 타입은 $\forall\alpha.[\alpha]\times[\alpha]\to[\alpha]$로 모든 타입에서 인스턴스화할 수 있습니다.

`fst`와 `snd` 함수는 "매개변수적 다형성(parametric polymorphism)"의 예를 보여줍니다.

$${\begin{aligned}
    {\mathsf {fst}}&:\forall \alpha .\forall \beta .\alpha \times \beta \to \alpha \\
    {\mathsf {snd}}&:\forall \alpha .\forall \beta .\alpha \times \beta \to \beta
\end{aligned}}$$

이 개념은 함수가 *하나 이상의 타입 매개변수에 대해 일반화될 수 있음*을 의미합니다.
이 경우, 두 함수 모두 두 가지 타입 변수, 𝛼와 𝛽에 대해 일반화되어 있습니다.
이러한 일반화를 통해 함수는 여러 타입의 입력을 받아들일 수 있으며, 입력된 타입에 따라 다르게 행동할 수 있습니다.

- `fst` 함수는 타입 𝛼와 𝛽의 쌍(pair)을 입력받고, 첫 번째 요소(타입 𝛼의 요소)를 반환합니다.
- `snd` 함수는 타입 𝛼와 𝛽의 쌍(pair)을 입력받고, 두 번째 요소(타입 𝛽의 요소)를 반환합니다.

이러한 함수들은 타입이 서로 다른 두 요소의 쌍을 다룰 수 있기 때문에, 매우 유연하게 사용될 수 있습니다.

예를 들어, `fst` 함수에 `(Int, String)` 타입의 쌍이 주어지면 `Int`를 반환하고, `snd` 함수에 같은 쌍이 주어지면 `String`을 반환합니다.

Go 1.18 버전 이상에서는 제네릭을 사용하여 유사한 기능을 구현할 수 있습니다. 아래는 `fst`와 `snd` 함수를 Go언어로 구현한 예시입니다:

```go
package main

import "fmt"

// Pair는 두 타입의 값을 저장하는 구조체입니다.
type Pair[A any, B any] struct {
    First A
    Second B
}

// Fst는 Pair에서 첫 번째 요소를 반환합니다.
func Fst[A any, B any](p Pair[A, B]) A {
    return p.First
}

// Snd는 Pair에서 두 번째 요소를 반환합니다.
func Snd[A any, B any](p Pair[A, B]) B {
    return p.Second
}

func main() {
    p := Pair[int, string]{First: 1, Second: "apple"}
    fmt.Println("First element:", Fst(p))  // 출력: First element: 1
    fmt.Println("Second element:", Snd(p)) // 출력: Second element: apple
}
```

- `Pair` 구조체는 두 타입의 값을 저장합니다. 이는 `fst`와 `snd` 함수에 대한 입력 타입으로 사용됩니다.
- `Fst` 함수는 `Pair`의 첫 번째 요소를 반환하고, `Snd` 함수는 두 번째 요소를 반환합니다.
- 이 예제는 Go에서 매개변수적 다형성을 사용하는 방법을 보여줍니다. 함수는 어떤 타입의 쌍에 대해서도 동작할 수 있으며, 타입 안전성을 유지합니다.

이러한 방식으로, Go에서 제네릭을 활용하여 다양한 타입의 입력을 처리하고, 유연하게 함수를 구현할 수 있습니다.

> 수식에 대한 설명
>
> 앞서 나온 수식들은 타입 이론에서 사용되는 표기법을 사용하고 있으며, 함수의 타입을 설명하기 위해 수학적 형식을 사용합니다.
>
>
> - `\forall`: 이 기호는 "모든"을 의미하는 '범위 한정자(universal quantifier)'입니다. 프로그래밍에서 제네릭을 정의할 때 모든 가능한 타입에 대해 해당 함수가 적용될 수 있음을 나타냅니다.
>
> - 𝛼 및 𝛽: 이들은 타입 변수로, 아직 구체적인 타입이 지정되지 않은 '일반적인' 타입을 나타냅니다. 예를 들어, 𝛼는 `int`, `string` 등 어떤 타입도 될 수 있습니다.
>
> - `.`: 이 기호는 "다음은"이라는 의미로, "∀𝛼.𝛼"는 "모든 𝛼에 대하여 𝛼"를 의미합니다.
>
> - `\times`: 이 기호는 곱셈을 의미하는 수학적 기호입니다만, 여기서는 타입의 쌍을 나타냅니다. 즉, 𝛼 x 𝛽는 타입 𝛼와 타입 𝛽의 요소를 갖는 쌍(pair)을 의미합니다.
>
> - `\to`: 이 기호는 함수의 반환 타입을 나타냅니다. 예를 들어, 𝛼 -> 𝛼는 타입 𝛼를 입력으로 받고 타입 𝛼를 반환하는 함수의 타입을 설명합니다.
>
> 따라서, 전체 수식 $\forall \alpha .\forall \beta .\alpha \times \beta \to \alpha$는
> "모든 타입 𝛼와 모든 타입 𝛽에 대해, 타입 𝛼와 타입 𝛽의 쌍을 입력으로 받아 타입 𝛼의 요소를 반환하는 함수"를 나타냅니다.
> 이는 `fst` 함수의 타입 서명으로, 모든 가능한 타입의 쌍에서 첫 번째 요소를 선택하여 반환하는 함수입니다.
>
> 이와 유사하게 `snd` 함수의 타입 서명 `\forall \alpha . \forall \beta . \alpha \times \beta \to \beta`는 "모든 타입 𝛼와 모든 타입 𝛽에 대해, 타입 𝛼와 타입 𝛽의 쌍을 입력으로 받아 타입 𝛽의 요소를 반환하는 함수"를 의미합니다.

## 바운드(bounded) 타입

프로그래밍 언어에서 제네릭을 사용할 때 모든 타입에 대해 완전히 자유롭게 함수를 적용하는 것은 때로는 제한적일 수 있습니다.
특정 연산을 수행하기 위해서는 입력 타입이 특정 인터페이스를 구현하거나 특정 특성을 가져야 할 필요가 있기 때문입니다.이러한 문제를 해결하기 위해, 많은 프로그래밍 언어들은 제네릭 타입에 제한을 둘 수 있는 기능을 제공합니다.

이것은 종종 "타입 제한", "타입 제약", 또는 "바운드(bounded) 타입"이라고 불립니다.

### 타입 클래스와 트레잇

**타입 클래스**와 **트레잇(trait)**은 비슷한 개념이지만, 다른 프로그래밍 언어에서 사용됩니다.

- Trait (Scala, Rust, PHP...)

    주로 코드의 재사용과 다중 상속의 이점을 제공하는 방법으로 사용되며, 제네릭 타입에 대한 연산 제한을 직접적으로 다루지는 않습니다.
    Scala와 PHP에서의 `trait`은 클래스 간에 구현 코드를 공유하는 메커니즘을 제공합니다.

    - **Scala**: 클래스에 다중 상속과 비슷한 기능을 제공하면서, 인터페이스처럼 사용될 수 있습니다. 메서드의 실제 구현을 포함할 수 있습니다.
    - **PHP**: 다중 상속의 기능을 모방하여, 클래스에 메서드나 속성을 주입할 수 있습니다. 이는 주로 코드 중복을 방지하고 재사용을 촉진하기 위해 사용됩니다.

- Type Class (Haskell에서)

    타입 클래스는 특정 타입이 수행해야 하는 연산을 정의하는 역할을 합니다.

    특정 연산을 수행할 수 있는 타입을 명시적으로 정의하고,
    제네릭 프로그래밍에서 해당 타입들을 활용할 때 해당 타입의 인스턴스에 대해 정의된 연산을 안전하게 수행할 수 있음을 보장하고,
    같은 함수가 타입에 따라 다르게 행동할 수 있는 함수의 다형성을 지원하고,
    매개변수 다형성에서 타입에 대한 연산 제한을 걸기 위해 사용됩니다.

    이를 통해 같은 연산이지만 서로 다른 타입에 대해 다른 구현을 제공할 수 있습니다.

- Trait vs. Type Class

    - **Trait**: 클래스에 구현 코드를 '믹스인'하는 방식으로 사용되며, 객체 지향 프로그래밍에서 클래스의 기능을 확장하는 데 초점을 맞춥니다.

    - **Type Class**: 타입 별로 다른 구현을 가질 수 있는 연산을 정의하며, 주로 함수형 프로그래밍에서 타입에 대한 일반적인 연산을 가능하게 하는 데 사용됩니다.

결론적으로, Scala와 PHP에서의 `trait`은 코드 재사용과 다중 상속의 이점을 제공하는 반면,
Haskell의 `type class`는 다형성과 타입 안전성을 강화하는 함수형 프로그래밍의 특성을 지원합니다.

### 타입 클래스 (Haskell 예시)

타입 클래스는 Haskell에서 타입에 대한 특정 행위를 정의하는 방법입니다.

예를 들어, `Eq` 타입 클래스는 동등성 비교를 지원하는 타입들을 위한 함수를 정의합니다.
어떤 타입이 `Eq` 클래스의 인스턴스가 되면, 그 타입의 값들을 비교할 수 있는 기능을 갖게 됩니다.

```haskell
class Eq a where
    (==), (/=) :: a -> a -> Bool

    x /= y = not (x == y)
```

위의 예에서 `Eq`는 동등성 비교를 위한 타입 클래스이며, 모든 `Eq`의 인스턴스는 `(==)`와 `(/=)` 함수를 제공해야 합니다.

### 트레잇 (Rust 예시)

Rust의 트레잇은 인터페이스와 유사하며, 특정 타입이 지원해야 하는 동작을 정의합니다.

예를 들어, 다음은 `Summable` 트레잇을 정의하는 예입니다.
이 트레잇은 요소들을 합하는 `sum` 메소드를 갖는 타입에 적용될 수 있습니다.

```rust
trait Summable {
    fn sum(&self) -> Self;
}

impl Summable for Vec<i32> {
    fn sum(&self) -> i32 {
        self.iter().fold(0, |a, &b| a + b)
    }
}
```

위의 예에서 `Vec<i32>` 타입에 대해 `Summable` 트레잇을 구현하면, `Vec<i32>`의 인스턴스에서 `sum` 메소드를 호출할 수 있습니다.

### 트레잇(scala)

제네릭에서 매개변수의 타입에 제한을 거는 것을 타입 클래스(Type Class)라고 합니다.
타입 클래스는 함수형 프로그래밍에서 유래한 개념으로, 특정 타입이 만족해야 하는 연산이나 동작을 정의하는 인터페이스를 말합니다.

스칼라에서는 타입 클래스를 `trait`과 `implicit` parameter를 사용해서 구현할 수 있습니다. 예를 들어:

```scala
trait Addable[T] {
  def add(a: T, b: T): T
}

implicit object IntAddable extends Addable[Int] {
  def add(a: Int, b: Int): Int = a + b
}

def sum[T](list: List[T])(implicit addable: Addable[T]): T = 
  list.reduce(addable.add)
```

여기서 `Addable`이 타입 클래스이고, 더하기 연산을 지원하는 타입 `T`에 대해 `add` 메소드를 제공합니다.
그리고 `IntAddable`은 `Int` 타입에 대한 `Addable`의 구현체입니다.

`sum` 함수는 `implicit` parameter로 `Addable[T]`를 받습니다.
이를 통해 리스트의 원소 타입 `T`가 `Addable`을 만족하는지 컴파일 타임에 검사할 수 있고, 런타임에는 해당하는 Addable 인스턴스를 주입받아 사용할 수 있습니다.

따라서 `trait` 자체가 타입 클래스는 아니지만, 스칼라에서 `trait`과 `implicit`을 활용해 타입 클래스 패턴을 구현하는 것이 일반적입니다.
타입 클래스를 통해 제네릭 코드에서 매개변수 타입에 제한을 걸 수 있습니다.

타입 클래스를 사용하면 다음과 같은 장점이 있습니다:

1. 기존 타입에 대해 추가 기능을 제공할 수 있음 (기존 코드 변경 없이)
2. 타입에 대한 제약 조건을 명시적으로 표현 가능
3. 제네릭 코드의 타입 안정성 향상
4. 런타임 오버헤드 없이 정적 디스패치 가능

이처럼 타입 클래스는 제네릭 프로그래밍에서 매우 강력하고 유연한 도구가 될 수 있습니다.
스칼라 뿐만 아니라 하스켈, 러스트 등 다양한 함수형/다중 패러다임 언어에서 활용되고 있습니다.

### 타입 제한

제네릭 타입에 대한 제한을 설정하여, 특정 연산을 수행할 수 있는 타입만을 허용하도록 할 수 있습니다.

예를 들어, 다음과 같은 제네릭 함수가 있을 때:

```rust
fn print_sum<T: Summable>(items: T) {
    println!("Sum: {}", items.sum());
}
```

여기서 `T: Summable`은 `T`가 `Summable` 트레잇을 구현해야 함을 의미합니다.
이렇게 하면 `print_sum` 함수는 `Summable` 트레잇을 만족하는 어떤 타입의 인스턴스도 받을 수 있습니다.

## 함수형에서의 `trait`와 PHP에서의 `trait`

PHP의 `trait`은 주로 "코드 재사용성"을 목적으로 하는 언어 구조입니다.
PHP는 단일 상속을 지원하는 언어로서, 한 클래스가 여러 클래스로부터 상속을 받을 수 없습니다.
이런 제한을 극복하기 위해, PHP는 `trait`을 도입했습니다.

### 트레잇의 개념 유래

트레잇은 기본적으로 "믹스인(mixin)"의 개념에서 발전한 것입니다.
믹스인은 다른 클래스의 메서드를 현재 클래스에 "믹싱"하여 재사용할 수 있게 하는 프로그래밍 기법입니다.
믹스인은 다중 상속의 어려움을 피하면서 클래스 간 코드를 재사용할 수 있는 방법을 제공합니다.

PHP의 `trait`은 이 믹스인 개념을 차용하고 있으며, 클래스에 필요한 메서드나 속성을 '주입'하는 형식으로 동작합니다.
이를 통해 개발자는 다중 상속의 복잡성 없이 여러 클래스에서 동일한 메서드나 속성을 공유할 수 있게 됩니다.

### vs PHP의 `trait`

PHP의 트레잇(trait)은 다른 프로그래밍 언어에서의 트레잇과는 약간 다른 개념을 가지고 있습니다.
PHP에서 트레잇은 특정 메서드나 속성을 여러 클래스에 재사용하기 위한 방법으로 제공됩니다.
이는 다중 상속의 일부 이점을 제공하면서도 클래스의 상속 구조를 단순하게 유지할 수 있도록 도와줍니다.

PHP에서 트레잇은 클래스와 유사하지만, 스스로 인스턴스화할 수는 없습니다.
대신, 트레잇은 클래스에 포함될 수 있어 하나의 클래스가 여러 트레잇을 사용하거나, 하나의 트레잇이 여러 클래스에 의해 사용될 수 있습니다.
이를 통해 코드 중복을 줄이고 다양한 클래스 간에 코드를 재사용할 수 있습니다.

트레잇은 클래스에 코드를 '주입'하는 방식으로 작동합니다.
즉, 트레잇 내에 정의된 메서드나 속성은 사용하는 클래스에서 직접 정의된 것처럼 동작합니다.
이는 특히 PHP와 같은 단일 상속 언어에서 유용하며, 다중 상속의 혜택을 부분적으로 누릴 수 있게 해줍니다.

```php
trait Logger {
    public function log($message) {
        echo $message;
    }
}

class FileLogger {
    use Logger;
}

class PrintLogger {
    use Logger;
}

$fileLogger = new FileLogger();
$fileLogger->log("Logging to a file.\n");

$printLogger = new PrintLogger();
$printLogger->log("Logging to the printer.\n");
```

- `Logger` 트레잇은 `log` 메서드를 정의하고 있으며, 이 메서드는 메시지를 출력합니다.
- `FileLogger`와 `PrintLogger` 클래스는 `Logger` 트레잇을 사용(`use`)하여 `log` 메서드를 포함합니다. 이는 각 클래스가 `log` 메서드를 자신의 메서드처럼 사용할 수 있음을 의미합니다.

이 예시에서 볼 수 있듯이, PHP의 트레잇은 코드의 재사용성을 높이는 효율적인 수단을 제공합니다.

그러나 트레잇은 기능을 '상속'하는 것이 아니라, 코드를 '복사'하여 사용하는 클래스에 삽입하는 방식으로 작동합니다.
따라서 트레잇은 러스트나 하스켈의 타입 클래스처럼 타입 시스템의 일부로 작동하는 것이 아니라, 더 프로시저적인 방식으로 코드 재사용을 돕는 도구로 사용됩니다.

### 명명의 이유

`trait`이라는 이름은 프로그래밍 언어인 스칼라(Scala)에서 영감을 받은 것으로 추정됩니다. 스칼라에서 트레잇(trait)은 인터페이스와 유사하지만, 구현 코드를 포함할 수 있는 기능을 가지고 있습니다. 스칼라의 트레잇은 객체 지향과 함수형 프로그래밍의 특징을 모두 갖추고 있으며, 이를 통해 클래스의 구현을 돕습니다.

PHP에서 `trait`을 도입하면서 비슷한 목적을 가진 기능을 제공하고자 했으나, 구체적으로는 믹스인과 더 유사하게 작동합니다. 그럼에도 불구하고, `trait`이라는 용어를 사용한 것은 이 개념이 다른 프로그래밍 언어에서 사용되고 있던 용어와 연결성을 가지며, 개발자들에게 좀 더 친숙하게 다가가기 위한 전략적 선택일 수 있습니다.

결론적으로, PHP의 `trait`은 코드의 재사용을 도모하고 다중 상속의 문제를 우회하기 위한 효과적인 수단으로 도입되었으며, 다양한 언어의 개념을 차용하고 통합하는 현대 프로그래밍 언어의 발전 방향을 잘 보여주는 예입니다.

1. PHP와 Scala의 Trait

    PHP의 `trait`은 클래스에 코드를 '주입'하는 수단으로서, 다중 상속을 흉내 낼 수 있게 해줍니다.
    `trait`은 코드 재사용을 목적으로 하며, 다른 클래스에서 공통적으로 사용되는 메서드나 속성을 중복 없이 관리할 수 있도록 합니다.
    PHP에서는 다중 상속이 지원되지 않기 때문에, `trait`을 사용하여 여러 클래스에서 메서드와 속성을 공유할 수 있습니다.

    Scala에서의 `trait`은 인터페이스와 유사하지만, 추상 메서드 뿐만 아니라 구현된 메서드도 포함할 수 있습니다.
    `trait`은 다중 상속의 특성을 제공하며, 클래스에 여러 트레잇을 '믹스인'할 수 있습니다.
    이를 통해 코드 재사용성을 향상시키고, 클래스를 통한 상속보다 유연한 설계가 가능합니다.
    Scala의 트레잇은 클래스에 선형적으로 합성되어, 다중 상속 시 발생할 수 있는 문제들을 해결합니다.

2. Haskell의 Type Class (타입 클래스)

    Haskell의 타입 클래스는 특정 타입이 수행할 수 있는 연산을 정의하는 인터페이스와 유사한 구조입니다.
    이는 주로 함수형 프로그래밍의 다형성을 다루는 데 사용됩니다.

    타입 클래스를 사용하면, 예를 들어 `+` 연산자가 여러 타입에 대해 작동하도록 일반화할 수 있습니다.
    각 타입은 타입 클래스에 명시된 연산을 지원해야 합니다.

    타입 클래스는 특정 연산을 수행할 수 있는 타입들의 그룹을 정의하며, 이를 통해 범용적인 코드 작성이 가능합니다.

3. Mixin

    Mixin은 객체 지향 프로그래밍에서 클래스에 추가적인 속성이나 메서드를 제공하는 클래스를 말합니다.
    Mixin은 스스로 인스턴스화될 수 없으며, 다른 클래스에 기능을 제공하기 위해 사용됩니다.

    다중 상속의 혜택을 제공하지만, 다중 상속의 복잡성과 일부 문제를 피할 수 있습니다.
    Mixin은 코드 재사용과 기능의 모듈화를 촉진합니다.

    일반적으로 Mixin은 특정 기능을 '믹스인'하여 사용하는 클래스에 삽입되어 기능을 확장합니다.