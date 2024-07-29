# [rust generic, traits, and lifetimes](https://doc.rust-lang.org/book/ch10-00-generics.html)

- [rust generic, traits, and lifetimes](#rust-generic-traits-and-lifetimes)
    - [Generic Data Types](#generic-data-types)
    - [lifetime](#lifetime)
        - [상세 설명](#상세-설명)
            - [예제](#예제)
        - [적용 시 주의점 및 한계](#적용-시-주의점-및-한계)
        - [예제 및 분석](#예제-및-분석)
            - [1. `impl<'a>`: 수명 매개변수를 제네릭으로 선언](#1-impla-수명-매개변수를-제네릭으로-선언)
            - [2. `Parser<'a, serde_json::Value>`에서의 `'a`](#2-parsera-serde_jsonvalue에서의-a)
            - [3. `input: &'a [u8]`에서의 `'a`](#3-input-a-u8에서의-a)
            - [왜 `'a`가 여러 번 등장하는가?](#왜-a가-여러-번-등장하는가)
    - [Troubleshooting](#troubleshooting)
        - [the value of the associated type `Error` in `Parser` must be specified](#the-value-of-the-associated-type-error-in-parser-must-be-specified)
        - [generic type parameter 대신 구체적 타입 사용](#generic-type-parameter-대신-구체적-타입-사용)

## [Generic Data Types](https://doc.rust-lang.org/book/ch10-01-syntax.html)

## lifetime

Rust는 수동 메모리 관리를 자동화하는 동시에 메모리 안전성을 보장하는 언어로 설계되었습니다.
C/C++ 같은 언어에서 발생하는 메모리 안전성 문제를 해결하기 위해 Rust는 참조자의 유효 기간을 컴파일러가 추론할 수 있도록 수명 시스템을 도입했습니다.
이로 인해 Rust는 컴파일 타임에 메모리 오류를 방지할 수 있는 강력한 안전성을 제공합니다.

수명은 *서로 다른 참조들이 어떻게 연관되어 있는지를 컴파일러에게 알려주는 일종의 **제네릭***입니다.
수명은 우리가 빌린 값(참조된 값)에 대한 충분한 정보를 컴파일러에게 제공할 수 있도록 하여 참조가 유효한지 확인할 수 있도록 해줍니다.

### 상세 설명

수명(Lifetime)은 Rust에서 참조자(reference)가 메모리에 얼마나 오랫동안 유효한지를 나타내는 개념입니다.
Rust의 메모리 관리 시스템은 *수명을 통해 참조가 유효한 범위를 추적*하여, 메모리 오류(예: 댕글링 포인터, 잘못된 메모리 접근)를 방지합니다.

Rust에서 수명(lifetime)은 *참조자가 메모리에 얼마나 오랫동안 유효한지를 컴파일러에게 알려주는 개념*입니다.

Rust는 메모리 안전성을 컴파일 타임에 보장하려고 하며, 이 과정에서 참조자가 유효한지 아닌지를 판단해야 합니다.
그러나 모든 경우에 컴파일러가 참조자의 유효성을 자동으로 파악하기 어려운 경우가 있습니다.
이때 *수명 매개변수는 제네릭 타입처럼 사용*되며, *참조자들이 서로 어떻게 관련되어 있는지를 명시적으로 표현*할 수 있습니다.

예를 들어, 두 개 이상의 참조자가 있을 때, 이 참조자들이 같은 데이터에 대해 동일한 수명을 공유하는지 또는 어느 것이 더 오래 지속되는지를 명확히 할 수 있습니다.

수명을 명시함으로써 컴파일러에게 참조자의 관계를 명확히 설명해 줄 수 있으며, 컴파일러는 이를 바탕으로 메모리 안전성을 더욱 폭넓게 검사할 수 있습니다.
이로 인해 컴파일러는 우리가 수명 정보를 제공하지 않았을 때보다 더 많은 경우에서 참조의 유효성을 확인하고 보장할 수 있습니다.

- **명시적 수명 (Explicit Lifetime)**:

    함수, 구조체, 또는 트레잇 정의에서 수명을 명시적으로 지정하여 참조자의 유효 기간을 설정합니다.

- **암시적 수명 (Implicit Lifetime)**:

    Rust 컴파일러가 자동으로 추론할 수 있는 경우, 수명을 명시하지 않아도 됩니다.
    대부분의 경우, Rust는 참조자의 수명을 자동으로 올바르게 추론할 수 있습니다.

수명은 참조의 유효성을 더 정밀하게 제어하기 위해 다양한 곳에서 사용됩니다.
특히 제네릭 함수나 구조체에서 여러 참조자가 있을 때, 이들이 어떤 관계를 갖는지를 명확히 할 필요가 있습니다.

#### 예제

- 함수에서의 수명 사용

    ```rust
    // 두 참조자의 관계를 명확히 하기 위한 수명 매개변수 사용 예시
    fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
        if x.len() > y.len() {
            x
        } else {
            y
        }
    }
    ```

    여기서 `'a`는 `x`와 `y`의 유효 기간이 같다는 것을 명시적으로 나타내어 반환된 참조가 두 입력 참조 중 가장 짧은 수명을 따르게 합니다.

- 제네릭 사용시 `'a` 수명

    Rust의 제네릭 시스템에서는 *모든 제네릭 매개변수(수명 매개변수 포함)를 먼저 선언한 후에 사용*할 수 있습니다.
    이는 Rust 컴파일러가 수명 추론을 정확히 수행하기 위해 필수적입니다.

    ```rust
    impl<T> MyTrait<T> for MyStruct {
    ```

    여기서 `impl<T>`를 선언하지 않고 `MyTrait<T>`를 사용하려 하면 오류가 발생합니다.
    수명 매개변수 `'a`도 제네릭의 일종으로, `Parser<'a, ...>`에서 사용하기 전에 `impl<'a>`에서 먼저 선언해주어야 합니다.

    `impl<'a>`에서 `'a`를 선언하지 않으면, 해당 수명 매개변수를 사용할 수 없기 때문에 컴파일 오류가 발생합니다.

### 적용 시 주의점 및 한계

1. **수명 오류**: 수명을 잘못 지정하거나 명시하지 않으면 컴파일러는 참조의 유효 기간을 추적할 수 없어 컴파일 오류를 발생시킵니다.
2. **복잡성**: 수명은 처음에 다루기 어려운 개념일 수 있으며, 복잡한 상황에서 올바르게 지정하는 것이 까다로울 수 있습니다.

### 예제 및 분석

```rs
pub struct JsonParser;

impl<'a> Parser<'a, serde_json::Value> for JsonParser {
    type Error = serde_json::Error;

    fn parse(&self, input: &'a [u8]) -> Result<serde_json::Value, Self::Error> {
        let input_str = std::str::from_utf8(input).map_err(|e| serde_json::Error::custom(e.to_string()))?;
        serde_json::from_str(input_str)
    }
}
```

#### 1. `impl<'a>`: 수명 매개변수를 제네릭으로 선언

```rust
impl<'a> Parser<'a, serde_json::Value> for JsonParser {
^^^^^^^^
```

`impl<'a>`에서 `'a`는 *`impl` 블록 내에서 사용되는 **모든** 참조자의 유효 기간*을 지정합니다.
이 `'a`는 이후에 `Parser` 트레잇이나 그 메서드들에서 참조자(`&T` 타입)의 유효 기간을 명시할 때 사용됩니다.

이 수명 매개변수는 `Parser` 트레잇의 수명 매개변수로 전달되며, 파싱할 입력 데이터의 수명을 명시적으로 표현합니다.

만약 `impl<'a>`에서 수명을 정의하지 않으면, `Parser` 트레잇의 수명 매개변수를 지정할 수 없게 됩니다.
이는 컴파일러가 참조의 유효 기간을 추적할 수 없게 되어 컴파일 오류가 발생할 수 있습니다.

#### 2. `Parser<'a, serde_json::Value>`에서의 `'a`

```rust
impl<'a> Parser<'a, serde_json::Value> for JsonParser {
               ^^^

    fn parse(&self, input: &'a [u8]) -> Result<serde_json::Value, Self::Error> {
                           ^^^
```

여기서 `'a`는 `Parser` 트레잇의 수명 매개변수를 지정합니다.
`Parser` 트레잇에서 제네릭 수명 매개변수를 사용하여, 트레잇의 메서드들에서 참조자의 유효 기간을 명시적으로 지정합니다.
`parse` 메서드의 입력으로 전달되는 `&'a [u8]`의 수명을 정의함으로써 `parse` 메서드가 입력 데이터를 처리하는 동안 그 데이터가 유효하게 유지됨을 보장합니다.
입력 데이터가 메서드가 데이터를 처리하는 동안 유효하게 유지하여 댕글링 참조(유효하지 않은 메모리 참조)가 발생하지 않도록 합니다.

#### 3. `input: &'a [u8]`에서의 `'a`

```rust
fn parse(&self, input: &'a [u8]) -> Result<serde_json::Value, Self::Error> {
```

여기서 `'a`는 `parse` 메서드의 입력 데이터 `input`의 수명을 나타냅니다.
`input`은 바이트 슬라이스(`&[u8]`)로, 이 참조자의 유효 기간이 `'a`로 지정되어 있습니다.
이는 이 메서드를 호출하는 동안 `input` 데이터가 유효하게 유지됨을 보장합니다.

#### 왜 `'a`가 여러 번 등장하는가?

Rust의 수명 시스템이 참조의 유효 기간을 명확하게 표현하기 위해 `impl<'a>`, `Parser<'a, serde_json::Value>`, 그리고 `input: &'a [u8]`에 각각 수명 `'a`을 명시해야 합니다.

Rust의 수명 매개변수 시스템은 매우 유연하며, 수명이 여러 곳에서 지정되고 사용될 수 있습니다. 각 위치에서 `'a`가 사용되는 이유는 다음과 같습니다:

1. **제네릭 수명 매개변수 정의 (`impl<'a>`)**:

    여기서 정의된 `'a`는 이후에 사용될 모든 참조자의 수명을 지정할 수 있습니다.
    즉, `impl<'a>`는 수명을 "정의"하며, 모든 참조자들이 동일한 수명 매개변수를 공유하도록 합니다.

    이 수명은 이후 `Parser<'a, serde_json::Value>`와 `parse` 메서드의 입력 파라미터에서 사용될 수 있습니다.

    만약 `impl<'a>`가 없으면?
    이후에 `'a`를 사용할 수 없게 되어, 컴파일러가 수명을 추론할 수 없으므로 컴파일 오류가 발생합니다.

2. **트레잇에서 수명 매개변수 지정 (`Parser<'a, serde_json::Value>`)**:

    트레잇을 구현할 때 이 수명 매개변수를 사용해 트레잇의 메서드들이 일정한 수명을 따르도록 강제합니다.
    이 수명은 트레잇의 메서드(`parse`)에서 참조자를 사용할 때 유효 기간을 정의합니다.

    `impl<'a>`는 수명을 "정의"하고, `Parser<'a, serde_json::Value>`는 이미 정의된 수명을 "사용"합니다.

    `Parser` 트레잇이 사용하는 참조자의 유효 기간을 명시적으로 지정해야, 해당 트레잇을 구현하는 모든 타입이 참조의 수명을 명확히 따를 수 있습니다.
    이 수명을 명시하지 않으면, Rust는 참조자의 유효 기간을 올바르게 추론할 수 없으며, 컴파일러가 메모리 안전성을 보장할 수 없습니다.

    `Parser<'a, serde_json::Value>`에 `'a`가 없으면?
    트레잇에서의 수명 매개변수가 일관되지 않게 되어, 참조자의 유효 기간을 올바르게 지정할 수 없습니다.

3. **메서드의 입력 파라미터 (`input: &'a [u8]`)**:

    메서드 내부에서 수명이 지정된 참조자를 사용하여 메서드가 실행되는 동안 입력 데이터가 유효하게 유지됨을 보장합니다.

    메서드의 입력 참조 `input`이 `'a`라는 수명을 가질 때, 이 수명은 `impl<'a>`와 `Parser<'a, serde_json::Value>`에서 정의된 수명과 일치해야 합니다.
    이 일관성을 통해 Rust는 입력 데이터가 메서드 실행 동안 안전하게 사용될 수 있음을 보장할 수 있습니다.

    `input: &'a [u8]`에서 `'a`가 없으면?
    메서드의 입력 참조가 언제까지 유효한지를 컴파일러가 판단할 수 없게 됩니다.

`impl` 블록, `Parser` trait, `parse` 메서드에서 모두 같은 `'a` 수명을 사용함으로써 참조의 유효 기간이 일관되게 유지됩니다. 각 참조가 언제까지 유효한지를 명확히 표현하는 데 필수적입니다.

## Troubleshooting

### the value of the associated type `Error` in `Parser` must be specified

```rs
pub trait Parser<'a, T> {
    type Error;
    fn parse(&self, input: &'a [u8]) -> Result<T, Self::Error>;
}

pub struct DataProcessor<'a, T> {
    parser: Box<dyn Parser<'a, T> + Send + Sync>,
                    ^^^^^^^^^^^^^
                    the value of the associated type `Error` in `Parser` must be specified
}

impl<'a, T> DataProcessor<'a, T> {
    pub fn new(parser: Box<dyn Parser<'a, T> + Send + Sync>) -> Self {
                               ^^^^^^^^^^^^^
                               the value of the associated type `Error` in `Parser` must be specified
        DataProcessor { parser }
    }

    pub fn process(&self, input: &'a [u8]) -> Result<T, Box<dyn std::error::Error>> {
        self.parser.parse(input).map_err(|e| e.into())
    }
}
```

Rust에서 *트레이트의 연관 타입은 해당 트레이트를 사용할 때 구체적인 타입으로 지정되어야* 합니다.

`Parser` 트레이트에 연관 타입 `Error`가 정의되어 있지만,
이 `Parser` trait을 사용하는 `DataProcessor`에 이 `Error` 타입이 구체적으로 지정되어 있지 않고,
`Error` 타입에 대한 정보가 없으므로 컴파일러가 이를 추론할 수 없습니다.

```rs
pub struct DataProcessor<'a, T, E> {
    parser: Box<dyn Parser<'a, T, Error = E> + Send + Sync>,
}

impl<'a, T, E> DataProcessor<'a, T, E> {
    pub fn new(parser: Box<dyn Parser<'a, T, Error = E> + Send + Sync>) -> Self {
        DataProcessor { parser }
    }

    pub fn process(&self, input: &'a [u8]) -> Result<T, E> {
        self.parser.parse(input)
    }
}
```

이를 해결하기 위해 타입을 매개변수화 하는 `E`를 사용합니다.
이는 실제로 구체적인 타입을 지정하는 것이 아니라, 타입을 매개변수화하는 것입니다.
- `Parser<'a, T, Error = E>`에서 `E`는 구체적인 타입이 아니라, 나중에 지정될 수 있는 "플레이스홀더" 역할을 합니다.
- 이렇게 함으로써 `Parser` 트레이트를 구현하는 각 구조체가 자신만의 오류 타입을 지정할 수 있게 됩니다.

`DataProcessor`를 사용할 때, `E`의 실제 타입은 다음 두 가지 방법 중 하나로 결정됩니다:

- 명시적 타입 지정:

    ```rust
    let processor: DataProcessor<'static, Value, serde_json::Error> = DataProcessor::new(json_parser);
    ```

    여기서 `E`는 `serde_json::Error`로 구체화됩니다.

- 타입 추론:

    ```rust
    let processor = DataProcessor::new(json_parser);
    ```

    이 경우, 컴파일러는 `json_parser`의 구현을 보고 `E`의 타입을 추론합니다.

`Parser` 트레이트를 정의할 때, `Error`를 연관 타입으로 선언하는 대신 제네릭 매개변수로 만듭니다.
`DataProcessor`는 이 제네릭 `E`를 그대로 사용하고, 실제 사용 시점에서 구체적인 타입이 결정됩니다.

예를 들어:

```rust
// 연관 타입을 정의했던 경우
// pub trait Parser<'a, T> {
//     type Error;
//     fn parse(&self, input: &'a [u8]) -> Result<T, Self::Error>;
// }

pub trait Parser<'a, T, E> {
    fn parse(&self, input: &'a [u8]) -> Result<T, E>;
}

pub struct JsonParser;

impl<'a> Parser<'a, Value, serde_json::Error> for JsonParser {
    fn parse(&self, input: &'a [u8]) -> Result<Value, serde_json::Error> {
        // 구현...
    }
}

// 사용 예
let json_parser = Box::new(JsonParser) as Box<dyn Parser<'static, Value, serde_json::Error> + Send + Sync>;
let processor = DataProcessor::new(json_parser);
```

이 예에서:
- `JsonParser`는 `Parser` 트레이트를 구현할 때 `E`를 `serde_json::Error`로 구체화합니다.
- `DataProcessor::new(json_parser)`를 호출할 때, `E`는 `serde_json::Error`로 결정됩니다.

또는 다음과 같이 사용할 수 있습니다.

```rs
let json_parser: Box<dyn Parser<'static, Value, Error = serde_json::Error> + Send + Sync> = Box::new(JsonParser);
let processor: DataProcessor<'static, Value, serde_json::Error> = DataProcessor::new(json_parser);
```

### generic type parameter 대신 구체적 타입 사용

```rs
impl<'a, Utf8Error> Parser<'a, Value, Utf8Error> for FormURLEncodedParser {
         ^^^^^^^^^
    fn parse(&self, input: &'a [u8]) -> Result<Value, Utf8Error> {
        let utf8_str_result = from_utf8(input);
        match utf8_str_result {
            Ok(utf8_str) => {
                println!("utf8_str:{}", utf8_str);
                let decoded = decode_escape_sequences(utf8_str);
                println!("decoded:{}", decoded);
                let pairs = form_urlencoded::parse(decoded.as_bytes());
                let mut result = serde_json::Map::new();

                for (key, value) in pairs.into_owned() {
                    add_or_update_value(&mut result, &key, &value);
                }

                Result::Ok(Value::Object(result))
            }
            Err(e) => {
                eprintln!("Error: {}", e);
                Result::Err(ParserError::from(e))
                            ^^^^^^^^^^^^^^^^^^^^
                            mismatched types
                            expected type parameter `Utf8Error` found enum `ParserError`
            }
        }
    }
}

```

하지만 `impl<'a, Utf8Error>`처럼 `Utf8Error`를 명시적으로 선언할 필요가 없습니다.

1. `impl<'a, Utf8Error> Parser`에서 `Utf8Error`는 제네릭 타입 매개변수로 실제 `std::str::Utf8Error`와는 다른 임의의 타입을 의미합니다.

    ```rust
    #[derive(Debug)]
    pub enum ParserError {
        Utf8Error(std::str::Utf8Error),
        JsonError(serde_json::Error),
        DecodingError(String),
        // 다른 에러 타입들...
    }

    impl<'a, Utf8Error> Parser<'a, Value, Utf8Error> for FormURLEncodedParser
             ^^^^^^^^^
             `std::str::Utf8Error`와는 다른, 임의의 타입
    ```

    `Utf8Error`는 제네릭 타입 매개변수로, 어떤 타입이든 될 수 있습니다.

2. `impl<'a> Parser<'a, Value, ParserError>`에서 `ParserError`는 구체적인 타입으로 사용됩니다.

    ```rust
    #[derive(Debug)]
    pub enum ParserError {
        Utf8Error(std::str::Utf8Error),
        JsonError(serde_json::Error),
        DecodingError(String),
        // 다른 에러 타입들...
    }

    impl<'a> Parser<'a, Value, ParserError> for FormURLEncodedParser
                               ^^^^^^^^^^^
                               이미 정의된 `ParserError` 열거형을 가리킵니다.
    ```

    `ParserError`는 구체적인 타입으로, 이미 정의된 특정 타입을 가리킵니다.

`ParserError`가 이미 `From<std::str::Utf8Error>`를 구현하고 있다면(앞서 정의한 `ParserError` 열거형에서 이를 구현했습니다),
`Utf8Error`를 `ParserError`로 자동 변환할 수 있습니다.

따라서 `Result::Err(ParserError::from(e))`와 같은 코드가 가능해집니다.

이런 방식으로 구체적인 에러 타입을 사용하면
- 코드가 더 명확해지고,
- 타입 안정성이 향상되며,
- 또한 `ParserError`를 통해 다양한 종류의 파싱 관련 오류를 하나의 타입으로 통합할 수 있어, 에러 처리가 더 일관되고 체계적으로 이루어질 수 있습니다.

```rs
impl<'a> Parser<'a, Value, ParserError> for FormURLEncodedParser {
    fn parse(&self, input: &'a [u8]) -> Result<Value, ParserError> {
        let utf8_str_result = from_utf8(input);
        match utf8_str_result {
            Ok(utf8_str) => {
                println!("utf8_str:{}", utf8_str);
                let decoded = decode_escape_sequences(utf8_str);
                println!("decoded:{}", decoded);
                let pairs = form_urlencoded::parse(decoded.as_bytes());
                let mut result = serde_json::Map::new();

                for (key, value) in pairs.into_owned() {
                    add_or_update_value(&mut result, &key, &value);
                }

                Result::Ok(Value::Object(result))
            }
            Err(e) => {
                eprintln!("Error: {}", e);
                Result::Err(ParserError::from(e))
            }
        }
    }
}
```

여기서는 impl에 'a처럼 이 스코프에서 Utf8Error 사용한다는 걸 명시하지 않아도 되는 건가요?
