# rust keywords

## `mod`

Rust에서 모듈을 선언할 때 사용됩니다.
모듈은 코드를 별도의 네임스페이스로 조직하는 방법으로, 다른 파일에 저장될 수 있습니다.

`mod encoding_type`을 선언할 경우:
- 컴파일러에게 `encoding_type.rs` 파일을 찾도록 지시하거나
- `encoding_type` 폴더 내의 `mod.rs` 파일을 찾도록 지시합니다.

```rust
// 모듈 선언
mod encoding_type;
```

## `use`

`use` 키워드는 모듈, 구조체, 열거형 등 경로를 현재 스코프로 가져와 더 짧은 경로로 사용할 수 있게 합니다.

예를 들어, `crate::encoding_type::EncodingType`을 매번 쓰는 대신, `use crate::encoding_type::EncodingType`을 사용하면 `EncodingType`만으로 접근할 수 있습니다.

```rust
// 특정 경로에서 항목 가져오기
use crate::encoding_type::EncodingType;
```

## `as`

`as` 키워드는 가져온 항목에 별칭(alias)을 붙일 때 사용됩니다.
이 기능은 긴 경로나 이름이 충돌할 때 특히 유용합니다.

예시:

```rust
// 별칭 사용
use std::io::Result as IoResult;
```

## `pub`

`pub` 키워드는 모듈, 함수, 구조체, 열거형 등을 공개(public)로 설정하여 다른 모듈에서도 접근할 수 있도록 합니다.
이 키워드를 사용하지 않으면 해당 항목은 기본적으로 비공개(private) 상태입니다.

```rust
// public 모듈로 설정
pub mod encoding_type;

// public 함수로 설정
pub fn parse_encoding() {
    // 함수 내용
}
```

## `super`

`super` 키워드는 부모 모듈을 참조할 때 사용됩니다.
현재 모듈에서 상위 모듈로 접근하고 싶을 때 유용합니다.

예시:

```rust
// 부모 모듈의 함수 호출
super::some_function();
```

## `self`

`self` 키워드는 현재 모듈을 참조할 때 사용됩니다.
이는 현재 모듈의 내부 항목을 명시적으로 가져와야 할 때 유용합니다.

예시:

```rust
// 현재 모듈에서 함수 가져오기
use self::my_function;
```

## `crate`

`crate` 키워드는 현재 크레이트(패키지)의 루트를 가리킵니다.
만약 `parser.rs`와 `encoding_type.rs`가 같은 크레이트의 일부라면, `crate::encoding_type::EncodingType`은 `EncodingType` enum에 대한 루트 경로가 됩니다.

```rust
// 크레이트 루트에서 경로 가져오기
use crate::encoding_type::EncodingType;
```

## `extern crate`

`extern crate`는 외부 크레이트(패키지)를 가져올 때 사용됩니다.
Rust 2015 에디션에서는 외부 크레이트를 명시적으로 가져와야 했지만, Rust 2018 에디션부터는 일반적으로 필요하지 않습니다.
대신 `Cargo.toml` 파일에 크레이트를 추가하면 자동으로 가져올 수 있습니다.

```rust
// Rust 2015 에디션에서 extern crate 사용 예
extern crate serde;
```

## `cfg`

`cfg`는 조건부 컴파일을 할 때 사용되는 지시어입니다.
특정 조건이 충족될 때만 코드가 컴파일되도록 설정할 수 있습니다.

```rust
// 특정 OS에서만 컴파일
#[cfg(target_os = "windows")]
fn windows_only_function() {
    // 함수 내용
}
```

## `#[derive]`

`#[derive]`는 구조체나 열거형에 자동으로 특정 트레이트(예: `Clone`, `Debug`)를 구현할 때 사용됩니다.
이 지시어는 구조체나 열거형 정의 바로 위에 위치합니다.

```rust
#[derive(Debug, Clone)]
struct MyStruct {
    field: i32,
}
```

## `const`

`const` 키워드는 불변의 상수를 선언할 때 사용됩니다.
상수는 *반드시 컴파일 시간에 결정*되어야 하며, *함수 내에서도 선언할 수* 있습니다.

```rust
// 상수 선언
const MAX_LIMIT: u32 = 100;
```

## `let`

`let` 키워드는 변수를 선언할 때 사용됩니다.
`let`으로 선언된 변수는 기본적으로 불변(immutable)이지만, `mut` 키워드를 사용해 가변(변경 가능)으로 만들 수 있습니다.

```rust
// 불변 변수
let x = 5;

// 가변 변수
let mut y = 10;
```

## `fn`

`fn` 키워드는 함수를 선언할 때 사용됩니다.
함수는 Rust 프로그램에서 코드의 재사용과 구조화를 위한 기본 단위입니다.

```rust
// 함수 선언
fn add(a: i32, b: i32) -> i32 {
    a + b
}
```

## `impl`

`impl` 키워드는 *구조체나 열거형에 대한 메서드 구현을 정의*할 때 사용됩니다.
또한 트레이트를 특정 타입에 대해 구현할 때도 사용됩니다.

```rust
struct Circle {
    radius: f64,
}

// 메서드 구현
impl Circle {
    fn area(&self) -> f64 {
        3.14 * self.radius * self.radius
    }
}

// 트레이트 구현
impl Display for Circle {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(f, "Circle with radius: {}", self.radius)
    }
}
```

## `trait`

`trait`는 Rust에서 인터페이스에 해당하는 개념으로, *특정 타입이 구현해야 하는 메서드 집합을 정의*합니다.

```rust
// 트레이트 정의
trait Drawable {
    fn draw(&self);
}
```

## `match`

`match` 키워드는 패턴 매칭을 위한 강력한 제어 구조입니다.
다양한 데이터 형태에 따라 코드를 분기 처리할 수 있습니다.

```rust
fn process_number(n: i32) {
    match n {
        1 => println!("One"),
        2 => println!("Two"),
        _ => println!("Other"),
    }
}
```

## `enum`

`enum`은 여러 상수 값을 하나의 타입으로 묶을 수 있는 열거형을 정의할 때 사용됩니다.

```rust
enum Direction {
    Up,
    Down,
    Left,
    Right,
}
```

## `struct`

`struct`는 여러 필드를 하나의 복합 데이터 타입으로 묶는 구조체를 정의할 때 사용됩니다.

```rust
struct Point {
    x: i32,
    y: i32,
}
```

## `loop`, `while`, `for`

Rust에서 반복문을 만들 때 사용되는 키워드들입니다.

- `loop`: 무한 루프를 생성합니다.
- `while`: 조건이 참인 동안 반복합니다.
- `for`: 반복 가능한 요소에 대해 반복합니다.

```rust
// 무한 루프
loop {
    println!("This will print forever!");
}

// 조건부 반복
while count < 5 {
    count += 1;
}

// 컬렉션 반복
for number in 1..5 {
    println!("{}", number);
}
```
