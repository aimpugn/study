# Numeric

- [Numeric](#numeric)
    - [IEEE 754](#ieee-754)
    - [double이란?](#double이란)
    - [float이란?](#float이란)
    - [double과 float의 차이](#double과-float의-차이)
    - [PHP에서 float과 float의 산술 연산(빼기) 결과 double이 나오는 이유](#php에서-float과-float의-산술-연산빼기-결과-double이-나오는-이유)
    - [PHP와 다른 언어들 비교](#php와-다른-언어들-비교)
        - [Go: `float32`와 `float64` 타입을 명확하게 구분](#go-float32와-float64-타입을-명확하게-구분)
        - [Rust: `f32`와 `f64` 타입을 명확하게 구분](#rust-f32와-f64-타입을-명확하게-구분)
        - [Zig: `f32`와 `f64` 타입을 구분](#zig-f32와-f64-타입을-구분)
        - [Kotlin: `Float`과 `Double` 타입을 구분](#kotlin-float과-double-타입을-구분)
        - [Scala: `Float`과 `Double` 타입을 구분](#scala-float과-double-타입을-구분)
        - [결론](#결론)
    - [코드 예제](#코드-예제)
        - [PHP](#php)
        - [Go](#go)

## IEEE 754

IEEE 754 표준은 부동 소수점 숫자의 표현 방식을 정의하고 산술을 위한 국제 표준입니다.
이 표준에는 단일 정밀도(32비트)와 이중 정밀도(64비트) 부동 소수점 형식을 정의하고 있습니다.
이를 따르는 대부분의 언어는 `float`과 `double`을 각각 단일 정밀도와 이중 정밀도로 구별하여 처리합니다.

이 표준에 따르면 부동 소수점 숫자는 다음 세 가지 부분으로 구성됩니다:
1. **부호 비트(Sign Bit)**: 숫자의 양수/음수를 나타냅니다.
2. **지수(Exponent)**: 숫자의 스케일을 나타냅니다.
3. **가수(Mantissa) 또는 유효 숫자(Significant)**: 숫자의 정밀도를 나타냅니다.

## double이란?

`double`은 *이중 정밀도*를 가지는 부동 소수점 숫자를 나타내는 자료형입니다.

IEEE 754 표준을 따르면, `double` 자료형은 다음과 같이 구성됩니다:
- 총 64비트로 구성
- 이 중 MSB(Most Significant Bit) 1비트는 부호
- 11비트는 지수
- 나머지 52비트는 가수

`double` 자료형은 더 높은 정밀도와 넓은 범위의 숫자를 표현할 수 있습니다.

## float이란?

`float`은 *단일 정밀도*를 가지는 부동 소수점 숫자를 나타내는 자료형입니다.
IEEE 754 표준을 따르면, `float` 자료형은 다음과 같이 구성됩니다:
- 총 32비트로 구성
- 이 중 MSB(Most Significant Bit) 1비트는 부호
- 8비트는 지수
- 나머지 23비트는 가수

`float` 자료형은 `double`보다 정밀도와 범위가 낮습니다.

## double과 float의 차이

`double`과 `float`의 주요 차이점은 정밀도와 비트 수에 있습니다.

- **float (32비트)**:

    ```plaintext
    S EEEEEEEE MMMMMMMMMMMMMMMMMMMMMMM
    1 8        23
    ```

    - **S (1비트)**: 부호 비트, 0이면 양수, 1이면 음수
    - **E (8비트)**: 지수 비트, 지수는 8비트로 표현됩니다.
    - **M (23비트)**: 가수 비트, 유효 숫자는 23비트로 표현됩니다.

    예를 들어 float으로 숫자 1.5는 다음과 같이 나타납니다.

    ```plaintext
    S  EEEEEEEE  MMMMMMMMMMMMMMMMMMMMMMM
    0  01111111  10000000000000000000000
    ```

    - **부호 비트(S)**: 0 (양수)
    - **지수 비트(E)**: 127 (float) - 지수는 127을 기준으로 오프셋된 값입니다.
    - **가수 비트(M)**: 1.5의 유효 숫자 부분

- **double (64비트)**:

    ```plaintext
    S EEEEEEEEEEE MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM
    1 11          52
    ```

    - **S (1비트)**: 부호 비트, 0이면 양수, 1이면 음수
    - **E (11비트)**: 지수 비트, 지수는 11비트로 표현됩니다.
    - **M (52비트)**: 가수 비트, 유효 숫자는 52비트로 표현됩니다.

    예를 들어 double로 숫자 1.5는 다음과 같이 나타납니다.

    ```plaintext
    S  EEEEEEEEEEE  MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM
    0  01111111111  1000000000000000000000000000000000000000000000000000
    ```

    - **부호 비트(S)**: 0 (양수)
    - **지수 비트(E)**: 1023 (double) - 지수는 1023을 기준으로 오프셋된 값입니다.
    - **가수 비트(M)**: 1.5의 유효 숫자 부분

## PHP에서 float과 float의 산술 연산(빼기) 결과 double이 나오는 이유

PHP에서 부동 소수점 숫자는 `float` 타입으로 정의됩니다.
*내부적으로는 이 부동 소수점 숫자가 이중 정밀도(double precision)를 사용하여 64비트로 저장*됩니다.
따라서, PHP에서는 `float`와 `double`을 사실상 구분하지 않습니다.
즉, PHP는 `float`과 `double`을 구분하지 않고, IEEE 754 표준을 따르기 때문에 항상 64비트의 `double`로 처리됩니다.

```php
<?php
$a = 1.0;
$b = 0.1;

$result = $a - $b;

echo "Result: " . $result . "\n"; // Result: 0.9
echo "Result type: " . gettype($result) . "\n"; // Result type: double
?>
```

위 코드에서 `1.0`과 `0.1`은 `float`로 시작하지만, 연산 결과는 `double`로 처리됩니다.
이는 PHP가 모든 부동 소수점 연산을 이중 정밀도로 처리하기 때문입니다.

- PHP에서 `float`은 부동 소수점 숫자를 나타내는 데 사용되는 일반적인 용어입니다.

    PHP에서 `float`이라는 용어를 사용하지만, 실제로는 모든 부동 소수점 숫자를 이중 정밀도(double)로 처리합니다.
    이는 다른 언어에서의 `double`과 동일한 의미로 사용됩니다.

    또한 PHP에는 `floatval()`, `(float)`, `(double)`, `(real)`과 같은 함수와 캐스팅 방법이 있습니다.
    이들은 모두 동일하게 동작하며, 값의 타입을 부동 소수점 숫자로 변환합니다.

- `floatval()` 함수

    `floatval()` 함수는 값을 부동 소수점 숫자로 변환합니다.
    이는 *실제로 값을 이중 정밀도 부동 소수점 숫자로 변환*하는 것을 의미합니다.

    ```php
    <?php
    $intValue = 42;
    $floatValue = floatval($intValue);

    echo "Float value: " . $floatValue . "\n"; // 42.0
    echo "Float value type: " . gettype($floatValue) . "\n"; // double

    $intValue = 42;
    $floatValue = floatval($intValue);
    $floatCast = (float) $intValue;
    $doubleCast = (double) $intValue;

    echo "floatval() value: " . $floatValue . "\n";             // floatval() value: 42
    echo "floatval() type: " . gettype($floatValue) . "\n";     // floatval() type: double

    echo "(float) cast value: " . $floatCast . "\n";            // (float) cast value: 42
    echo "(float) cast type: " . gettype($floatCast) . "\n";    // (float) cast type: double

    echo "(double) cast value: " . $doubleCast . "\n";          // (double) cast value: 42
    echo "(double) cast type: " . gettype($doubleCast) . "\n";  // (double) cast type: double
    ?>
    ```

## PHP와 다른 언어들 비교

PHP에서 `float`과 `double`을 모두 이중 정밀도 부동 소수점 숫자로 처리하는 것은 언어 설계의 일환이며, PHP에서의 표준 동작입니다.
이는 PHP가 부동 소수점 연산에서 일관성을 유지하고, 더 높은 정밀도를 제공하기 위함입니다.
그러나 다른 많은 프로그래밍 언어는 `float`(32비트 단일 정밀도)와 `double`(64비트 이중 정밀도)을 구별하여 처리합니다.

### Go: `float32`와 `float64` 타입을 명확하게 구분

Go에서는 `float32`와 `float64` 타입을 명확하게 구분합니다. `float32`는 단일 정밀도, `float64`는 이중 정밀도를 사용합니다.

```go
package main

import (
    "fmt"
)

func main() {
    var floatValue float32 = 1.23456789
    var doubleValue float64 = 1.2345678901234567

    fmt.Printf("Float value: %.8f\n", floatValue)
    fmt.Printf("Double value: %.16f\n", doubleValue)
}
```

### Rust: `f32`와 `f64` 타입을 명확하게 구분

Rust에서도 `f32`와 `f64` 타입을 명확하게 구분합니다. `f32`는 단일 정밀도, `f64`는 이중 정밀도를 사용합니다.

```rust
fn main() {
    let float_value: f32 = 1.23456789;
    let double_value: f64 = 1.2345678901234567;

    println!("Float value: {:.8}", float_value);
    println!("Double value: {:.16}", double_value);
}
```

### Zig: `f32`와 `f64` 타입을 구분

Zig 또한 `f32`와 `f64` 타입을 구분합니다.

```zig
const std = @import("std");

pub fn main() void {
    var float_value: f32 = 1.23456789;
    var double_value: f64 = 1.2345678901234567;

    std.debug.print("Float value: {:.8f}\n", .{float_value});
    std.debug.print("Double value: {:.16f}\n", .{double_value});
}
```

### Kotlin: `Float`과 `Double` 타입을 구분

Kotlin에서는 `Float`(32비트 단일 정밀도)과 `Double`(64비트 이중 정밀도) 타입을 구분합니다.

```kotlin
fun main() {
    val floatValue: Float = 1.23456789f
    val doubleValue: Double = 1.2345678901234567

    println("Float value: %.8f".format(floatValue))
    println("Double value: %.16f".format(doubleValue))
}
```

### Scala: `Float`과 `Double` 타입을 구분

Scala에서도 `Float`(32비트 단일 정밀도)과 `Double`(64비트 이중 정밀도) 타입을 구분합니다.

```scala
object Main extends App {
    val floatValue: Float = 1.23456789f
    val doubleValue: Double = 1.2345678901234567

    println(f"Float value: $floatValue%.8f")
    println(f"Double value: $doubleValue%.16f")
}
```

### 결론

대부분의 프로그래밍 언어는 IEEE 754 표준을 따르며, 단일 정밀도(`float`/`f32`)와 이중 정밀도(`double`/`f64`) 부동 소수점 숫자를 명확하게 구분하여 처리합니다. PHP는 특별한 경우로, 모든 부동 소수점 숫자를 이중 정밀도 부동 소수점으로 처리합니다. 이는 PHP의 설계 결정에 따른 것이며, 다른 많은 언어는 이를 따르지 않습니다.

따라서, `float`과 `double`을 명확하게 구분하는 것은 일반적인 관행이며, 이는 IEEE 754 표준을 따르는 대부분의 언어에서 일반적인 동작 방식입니다. PHP의 접근 방식은 이와는 다르지만, 이는 PHP의 특수한 설계 결정입니다.

## 코드 예제

### PHP

```php
<?php
$floatValue = 1.23456789;
$doubleValue = 1.2345678901234567;

echo "Float value: " . $floatValue . "\n";
echo "Double value: " . $doubleValue . "\n";

// 차이를 보여주는 예제
$floatValueSmall = 0.1;
$doubleValueSmall = 0.1;
$sumFloat = $floatValueSmall + $floatValueSmall + $floatValueSmall;
$sumDouble = $doubleValueSmall + $doubleValueSmall + $doubleValueSmall;

echo "Sum with float: " . $sumFloat . "\n";
echo "Sum with double: " . $sumDouble . "\n";
?>
```

### Go

```go
package main

import (
    "fmt"
)

func main() {
    var floatValue float32 = 1.23456789
    var doubleValue float64 = 1.2345678901234567

    fmt.Printf("Float value: %.8f\n", floatValue)
    fmt.Printf("Double value: %.16f\n", doubleValue)

    // 차이를 보여주는 예제
    var floatValueSmall float32 = 0.1
    var doubleValueSmall float64 = 0.1
    var sumFloat = floatValueSmall + floatValueSmall + floatValueSmall
    var sumDouble = doubleValueSmall + doubleValueSmall + doubleValueSmall

    fmt.Printf("Sum with float: %.8f\n", sumFloat)
    fmt.Printf("Sum with double: %.16f\n", sumDouble)
}
```
