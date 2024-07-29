# rust attribute

## [속성(Attribute)](https://doc.rust-lang.org/reference/attributes.html)

Rust의 `#[derive(Parser, Debug)]`와 같은 구문들은 속성(attribute)이라고 불리는 메타데이터 구조입니다.
- 코드에 추가 정보를 제공
- 컴파일러에게 특정 동작을 지시

속성은 다음과 같은 형태를 가집니다:

```rust
#[attribute_name(attribute_arguments)]
```

[공식 문서](https://doc.rust-lang.org/reference/attributes.html)에 따르면, 속성을 다음과 같이 설명하고 있습니다:

> An *attribute* is a general, free-form metadatum that is interpreted according to name, convention, language, and compiler version.
> Attributes are modeled on Attributes in [ECMA-335](https://www.ecma-international.org/publications-and-standards/standards/ecma-335/), with the syntax coming from [ECMA-334](https://www.ecma-international.org/publications-and-standards/standards/ecma-334/) (C#).

- general (일반적인):

    속성이 특정 용도에 국한되지 않고 다양한 목적으로 사용될 수 있음을 의미합니다.
    속성은 코드 생성, 조건부 컴파일, 테스트 설정 등 다양한 용도로 사용됩니다.

- Free-form (자유 형식의):

    속성의 구조가 엄격하게 정해져 있지 않으므로 다양한 형태로 작성될 수 있습니다.

    ```rust
    #[attribute_name]
    #[attribute_name = "value"]
    #[attribute_name(key = "value")]
    ```

- Metadatum (메타데이터의 단수형):

    메타데이터란 "데이터에 대한 데이터"로, 다른 데이터에 대한 정보를 제공하는 데이터를 의미합니다.
    Rust의 속성은 코드에 대한 추가 정보를 제공하므로 메타데이터로 볼 수 있습니다.

- Name (이름):

    속성의 이름에 따라 그 의미와 동작이 결정됩니다.

    예를 들어, `#[derive(...)]`와 `#[test]`는 서로 다른 목적으로 사용됩니다.

- Convention (관례):

    Rust 커뮤니티에서 일반적으로 받아들여지는 속성 사용 방식을 의미합니다.
    [Rust API 가이드라인](https://rust-lang.github.io/api-guidelines/interoperability.html)에서 이러한 관례의 예를 볼 수 있습니다.

- Language (언어):

    Rust 언어의 특성과 규칙에 따라 속성이 해석됩니다.

    예를 들어, Rust의 모듈 시스템이나 타입 시스템과 상호작용하는 방식 등이 이에 해당합니다.

- Compiler version (컴파일러 버전):

    Rust 컴파일러의 버전에 따라 속성의 지원 여부나 동작 방식이 달라질 수 있습니다.
    버전별 상세 사항은 [Rust 릴리스 노트](https://github.com/rust-lang/rust/blob/master/RELEASES.md)을 참고합니다.

속성은 크게 내부 속성과 외부 속성으로 나눌 수 있습니다:
1. 내부 속성 (inner attributes):

    내부 속성은 속성이 선언된 항목에 적용됩니다.
    주로 모듈이나 함수의 본문 내에서 사용됩니다.

    ```rs
    # ! [ Attr ]
    ```

    ```rs
    fn main() {
        #![allow(unused_variables)]
        let x = 5;
    }
    ```

2. 외부 속성 (outer attributes):

    속성 다음에 오는 항목에 적용되며, 가장 일반적인 형태의 속성입니다.

    ```rs
    # [ Attr ]
    ```

    ```rs
    #[derive(Debug)]
    struct Point {
        x: i32,
        y: i32,
    }
    ```

## 동작 방식

```rs
#[derive(Parser, Debug)]
struct Args {
    /// Input file path
    input: String,

    /// Source encoding (utf-8, euc-kr, iso-8859-1)
    #[arg(short, long, default_value = "utf-8")]
    encoding: String,
}
```

1. 컴파일러가 `#[derive(Parser, Debug)]` 속성을 만납니다.
2. `derive` 매크로가 호출되어 `Parser`와 `Debug` 트레이트의 구현을 생성합니다.
3. `Parser` 구현은 구조체의 필드를 분석하여 명령줄 인자 파싱 코드를 생성합니다.
4. `Debug` 구현은 구조체의 필드를 문자열로 표현하는 코드를 생성합니다.
5. 생성된 코드는 원래의 구조체 정의에 추가됩니다.

## 속성의 종류

### [내장 속성 (Built-in attributes)](https://doc.rust-lang.org/reference/attributes.html#built-in-attributes-index)

Rust 언어에 기본적으로 포함된 속성으로, 컴파일러에 직접 구현되어 있습니다.

```rs
#[derive(Debug)]
#[cfg(test)]
#[inline]
```

- 컴파일러가 속성을 직접 인식하고 해석합니다.
- 추가적인 코드 생성 단계 없이 컴파일러의 내부 로직에 따라 처리됩니다.
- 일반적으로 매우 효율적으로 처리됩니다.

예를 들어, [`#[inline]` 속성](https://github.com/rust-lang/rust/blob/4fe1e2bd5bf5a6f1cb245f161a5e9d315766f103/compiler/rustc_attr/src/builtin.rs#L43-L49)은 컴파일러에게 함수를 인라인화하라고 지시합니다.
이는 컴파일러의 최적화 단계에서 직접 처리됩니다.

```rs
#[derive(Copy, Clone, PartialEq, Encodable, Decodable, Debug, HashStable_Generic)]
pub enum InlineAttr {
    None,
    Hint,
    Always,
    Never,
}
```

특징은 다음과 같습니다:
- 컴파일러에 의해 직접 해석되므로 처리 속도가 빠릅니다.
- 언어 수준의 기능을 제어할 수 있습니다 (예: 조건부 컴파일, 테스트 설정).
- 추가 의존성 없이 사용할 수 있습니다.

```rust
#[derive(Debug)]
^^^^^^^^^^^^^^^^ 내장 속성
struct Point {
    x: i32,
    y: i32,
}

#[test]
^^^^^^^ 내장 속성
fn test_point() {
    let p = Point { x: 1, y: 2 };
    assert_eq!(format!("{:?}", p), "Point { x: 1, y: 2 }");
}
```

### 매크로 속성 (Macro attributes)

사용자나 외부 라이브러리에 의해 정의된 속성으로, 라이브러리 작성자가 도메인 특화 언어(DSL)를 만들 수 있게 합니다.

```rust
#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
```

- 컴파일 시간에 처리되지만, 컴파일러 외부의 코드(프로시저 매크로)에 의해 처리됩니다.
- 이 프로시저 매크로는 입력 토큰 스트림을 받아 새로운 토큰 스트림을 생성합니다.
- 생성된 코드는 다시 컴파일러에 의해 처리됩니다.

특징은 다음과 같습니다:
- 컴파일 시간에 코드를 생성하거나 변환할 수 있습니다.
- 외부 크레이트에 의해 정의되고 처리됩니다.
- 더 복잡한 메타프로그래밍 작업을 수행할 수 있습니다.

```rust
use clap::Parser;

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ 매크로 속성
struct Cli {
    #[arg(short, long)]
    ^^^^^^^^^^^^^^^^^^^ 매크로 속성
    name: String,
}
```

[Rust 공식 문서의 프로시저 매크로 섹션](https://doc.rust-lang.org/book/ch19-06-macros.html#procedural-macros-for-generating-code-from-attributes)에서 자세한 내용을 확인할 수 있습니다.

### Derive 매크로 헬퍼 속성 (Derive macro helper attributes)

Derive 매크로와 함께 사용되는 보조 속성입니다.
추가적인 커스터마이징을 제공하여 Derive 매크로의 동작을 미세 조정할 수 있습니다.

```rust
#[derive(Serialize, Deserialize)]
struct Person {
    #[serde(rename = "firstName")]
    first_name: String,
}
```

- Derive 매크로에 의해 처리됩니다.
- 특정 필드나 변형에 대한 세부적인 제어를 제공합니다.
- 주로 데이터 구조의 직렬화, 역직렬화, 기타 자동 생성 코드의 동작을 조정하는 데 사용됩니다.

```rust
use serde::Deserialize;

#[derive(Deserialize)]
struct Person {
    #[serde(rename = "firstName")]
    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ Derive 매크로 헬퍼 속성
    first_name: String,
    #[serde(rename = "lastName")]
    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ Derive 매크로 헬퍼 속성
    last_name: String,
}
```

[serde 문서의 속성 섹션](https://serde.rs/attributes.html)에서 이러한 헬퍼 속성의 예를 볼 수 있습니다.

```rs
#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]  // <-- this is a container attribute
struct S {
    #[serde(default)]  // <-- this is a field attribute
    f: i32,
}

#[derive(Serialize, Deserialize)]
#[serde(rename = "e")]  // <-- this is also a container attribute
enum E {
    #[serde(rename = "a")]  // <-- this is a variant attribute
    A(String),
}
```

모두 `#[serde(...)]` 같은 형식이지만, 사용하는 위치에 따라 Container, Variant, Field 속성으로 분류합니다.

### 도구 속성 (Tool attributes)

특정 도구나 IDE를 위한 속성입니다.
컴파일러나 언어 자체와는 직접적인 관련이 없는 메타데이터를 제공합니다.
이 속성을 통해 개발 환경과의 통합이 가능합니다.

```rust
#[cfg_attr(feature = "nightly", feature(test))]
#[allow(clippy::too_many_arguments)]
```

- *컴파일러에 의해 무시*되며, 특정 도구에 의해서만 해석됩니다.
- 코드의 동작에 영향을 주지 않고 추가 정보를 제공합니다.
- IDE 기능, 문서화, 정적 분석 등을 지원하는 데 사용됩니다.

```rust
#[cfg_attr(feature = "nightly", feature(test))]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ 도구 속성
#[cfg(test)]
mod tests {
    #[test]
    fn it_works() {
        assert_eq!(2 + 2, 4);
    }
}
```

[Rust 레퍼런스의 도구 속성 섹션](https://doc.rust-lang.org/reference/attributes.html#tool-attributes)에서 자세한 내용을 확인할 수 있습니다.

## 속성 적용 가능 영역

### 1. 항목 선언 (Item declarations)

모든 항목 선언은 외부 속성을 받아들입니다.

예시:

```rust
#[deprecated(since = "1.0.0", note = "please use `new_function` instead")]
fn old_function() {
    // ...
}
```

### 2. 문장 (Statements)

대부분의 문장은 외부 속성을 받아들입니다.

예시:

```rust
fn main() {
    #[allow(unused_variables)]
    let x = 5;
}
```

### 3. 블록 표현식 (Block expressions)

블록 표현식은 외부 및 내부 속성을 받아들입니다.

예시:

```rust
fn main() {
    let result = #[allow(unused_variables)] {
        let x = 5;
        x * 2
    };
}
```

### 4. 열거형 변형과 구조체, 공용체 필드 (Enum variants and struct and union fields)

예시:

```rust
enum Message {
    #[deprecated]
    Old(String),
    New(String),
}

struct Person {
    #[serde(rename = "firstName")]
    first_name: String,
}
```

### 5. Match 표현식 팔 (Match expression arms)

예시:

```rust
match value {
    #[cfg(feature = "special")]
    Special => println!("Special case"),
    _ => println!("Normal case"),
}
```

### 6. 제네릭 수명 또는 타입 매개변수 (Generic lifetime or type parameter)

예시:

```rust
fn function<#[lifetime_attr] 'a, #[type_attr] T>(x: &'a T) {
    // ...
}
```

### 7. 표현식 (Expressions)

제한된 상황에서 표현식은 외부 속성을 받아들입니다.

예시:

```rust
fn main() {
    let x = #[allow(unused_variables)] 5;
}
```

### 8. 함수, 클로저, 함수 포인터 매개변수 (Function, closure and function pointer parameters)

예시:

```rust
fn function(#[attr] x: i32) {
    // ...
}

let closure = |#[attr] x: i32| x * 2;
```

이러한 다양한 속성과 그 적용 영역을 이해하면, Rust의 메타프로그래밍 기능을 더욱 효과적으로 활용할 수 있습니다. 예를 들어, 사용자 정의 속성을 만들어 코드 생성을 자동화하거나, 조건부 컴파일을 통해 플랫폼별 최적화를 수행하는 등의 고급 기법을 구사할 수 있습니다.

이 속성 시스템의 복잡성과 유연성은 Rust가 다양한 프로그래밍 패러다임과 사용 사례를 지원할 수 있게 해주는 핵심 요소 중 하나입니다. [Rust 공식 문서](https://doc.rust-lang.org/reference/attributes.html)에서 더 자세한 정보를 확인할 수 있습니다.

## 속성 예제

### `#[derive(Parser, Debug)]`

이 속성은 `Parser`와 `Debug` 트레이트(trait)를 자동으로 구현합니다.

- `derive`

    Rust의 프로시저 매크로(procedural macro) 중 하나입니다.

    > 프로시저 매크로? *컴파일 시간에 코드를 생성*하는 메타프로그래밍 도구

    [`derive` 매크로는 지정된 트레이트의 기본 구현을 자동으로 생성](https://doc.rust-lang.org/book/ch19-06-macros.html#how-to-write-a-custom-derive-macro)합니다.

- `Parser` 트레이트

    `Parser` 트레이트는 `clap` 크레이트에서 제공하는 사용자 정의 트레이트입니다.
    이 트레이트는 명령줄 인자를 파싱하는 기능을 제공합니다.

    [`clap` 크레이트의 `Parser` 트레이트를 derive하면 구조체의 필드를 기반으로 명령줄 인자 파싱 코드가 자동으로 생성](https://docs.rs/clap/latest/clap/_derive/index.html)됩니다.

- `Debug` 트레이트

    `Debug` 트레이트는 Rust 표준 라이브러리에서 제공하는 트레이트로, 디버그 목적의 포맷팅을 위한 기능을 제공합니다.

    [`Debug` 트레이트를 구현하면 `{:?}` 포맷 지정자를 사용하여 값을 출력](https://doc.rust-lang.org/std/fmt/trait.Debug.html)할 수 있습니다.

## #[command(author, version, about, long_about = None)]

이 속성은 `clap` 크레이트의 `command` 속성으로, 명령줄 애플리케이션의 메타데이터를 정의합니다.

### 동작 방식

1. 컴파일러가 `#[command(...)]` 속성을 만납니다.
2. `clap` 크레이트의 매크로가 이 속성을 처리합니다.
3. 매크로는 지정된 메타데이터(author, version, about)를 애플리케이션 정보로 설정합니다.
4. `long_about = None`은 긴 설명을 사용하지 않음을 나타냅니다.

[clap 문서](https://docs.rs/clap/latest/clap/_derive/index.html#command-attributes)에 따르면, 이 메타데이터는 `--help` 또는 `--version` 플래그가 사용될 때 표시됩니다.

## 이론적 배경

1. **메타프로그래밍**: 속성과 매크로는 메타프로그래밍의 일종입니다. 메타프로그래밍은 프로그램이 다른 프로그램을 작성하거나 수정하는 기술을 말합니다. 이는 코드의 추상화 수준을 높이고 반복적인 작업을 줄일 수 있게 해줍니다.

2. **선언적 프로그래밍**: 속성을 사용하는 방식은 선언적 프로그래밍의 예입니다. 선언적 프로그래밍은 원하는 결과를 명시하고, 그 결과를 얻기 위한 구체적인 단계는 시스템에 맡기는 프로그래밍 패러다임입니다.

3. **컴파일 시간 코드 생성**: Rust의 매크로 시스템은 컴파일 시간에 코드를 생성합니다. 이는 런타임 오버헤드 없이 코드의 재사용성과 추상화를 높일 수 있게 해줍니다.

4. **트레이트 기반 프로그래밍**: Rust의 트레이트 시스템은 타입 클래스(type class) 개념에 기반합니다. 이는 함수형 프로그래밍 언어인 Haskell에서 유래한 개념으로, 다형성(polymorphism)을 구현하는 강력한 방법을 제공합니다.

이러한 개념들을 이해하고 응용함으로써, 더 유연하고 재사용 가능한 코드를 작성할 수 있습니다. 예를 들어, 자신만의 커스텀 derive 매크로를 만들어 반복적인 코드 작성을 줄이거나, 속성을 사용하여 코드에 메타데이터를 추가하고 이를 활용하는 도구를 만들 수 있습니다.

## 심화 학습 내용 및 관련 개념

- **프로시저 매크로(Procedural Macros)**: Rust의 고급 메타프로그래밍 기능으로, 컴파일 시간에 임의의 코드를 생성할 수 있습니다. [Rust 공식 문서](https://doc.rust-lang.org/reference/procedural-macros.html)에서 자세한 내용을 확인할 수 있습니다.

- **트레이트 객체(Trait Objects)**: Rust에서 동적 디스패치를 구현하는 방법으로, 런타임에 다형성을 구현할 수 있게 해줍니다. [Rust 공식 문서](https://doc.rust-lang.org/book/ch17-02-trait-objects.html)에서 자세한 내용을 확인할 수 있습니다.

- **제네릭 프로그래밍(Generic Programming)**: Rust의 제네릭 시스템은 트레이트와 결합하여 강력한 추상화 도구를 제공합니다. [Rust 공식 문서](https://doc.rust-lang.org/book/ch10-00-generics.html)에서 자세한 내용을 확인할 수 있습니다.

- **컴파일러 플러그인(Compiler Plugins)**: Rust 컴파일러의 동작을 확장하는 방법으로, 더 복잡한 메타프로그래밍 작업을 수행할 수 있습니다. 현재는 불안정한 기능이지만, 향후 Rust의 메타프로그래밍 기능을 더욱 강화할 것으로 예상됩니다.

이러한 개념들을 심도 있게 학습하면, Rust의 메타프로그래밍 기능을 더욱 효과적으로 활용할 수 있을 것입니다.
