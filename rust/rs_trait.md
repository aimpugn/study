# rust trait

- [rust trait](#rust-trait)
    - [trait](#trait)

## [trait](https://doc.rust-lang.org/book/ch10-02-traits.html)

`trait`는 Rust에서 객체 지향 프로그래밍의 인터페이스에 해당하는 개념입니다.
`trait`는 특정 타입이 반드시 구현해야 하는 메서드, 연관 타입(associated type), 연관 상수, 제네릭 메서드, 수명 매개변수, 디폴트 타입 파라미터 등을 정의합니다.

1. **메서드 (Methods)**:

    `trait`는 타입이 반드시 구현해야 하는 메서드를 정의합니다.
    이러한 메서드는 `trait`을 구현하는 모든 타입에서 동일한 인터페이스를 제공하기 위해 구현되어야 합니다.
    메서드는 반드시 구현해야 하는 추상 메서드일 수도 있고, 기본 구현을 제공하는 메서드일 수도 있습니다.

    ```rust
    trait Example {
        fn required_method(&self); // 추상 메서드

        fn default_method(&self) {  // 기본 구현을 가진 메서드
            println!("This is a default method");
        }
    }
    ```

    `trait`에서 기본 메서드 구현을 제공함으로써 `trait`을 구현하는 타입이 필요에 따라 이 메서드를 재정의할 수 있습니다.

2. **연관 타입 (Associated Types)**:

    `trait`은 연관 타입을 정의할 수 있으며, 이를 통해 `trait`를 구현하는 타입이 자신의 타입 컨텍스트에서 이 연관 타입을 지정하도록 강제할 수 있습니다.

    ```rust
    trait Example {
        type Output;
        fn process(&self) -> Self::Output;
    }
    ```

3. **연관 상수 (Associated Constants)**:

    `trait`는 연관 상수(associated constants)도 정의할 수 있습니다.
    이 상수들은 `trait`를 구현하는 각 타입에 대해 고유한 값을 가질 수 있습니다.

    ```rust
    trait Example {
        const ID: u32;
    }
   ```

4. **제네릭 메서드 (Generic Methods)**:

    `trait`는 제네릭 메서드를 정의할 수 있습니다.
    이를 통해 메서드가 여러 타입에 대해 동작할 수 있도록 유연성을 제공합니다.

    ```rust
    trait Example {
        fn generic_method<T>(&self, value: T);
    }
    ```

5. **수명 매개변수 (Lifetime Parameters)**:

    `trait`는 수명 매개변수를 포함할 수 있습니다.
    이는 *메서드가 사용하는 참조의 유효 기간을 지정*하여 메모리 안전성을 보장합니다.

    ```rust
    trait Example<'a> {
        fn get_reference(&'a self) -> &'a str;
    }
    ```

6. **디폴트 타입 파라미터 (Default Type Parameters)**:

    `trait`는 제네릭 타입 파라미터에 대해 기본값을 제공할 수 있습니다.
    이는 `trait`의 유연성을 높이고, 구현할 때 복잡성을 줄여줍니다.

    ```rust
    trait Example<T = i32> {
        fn process(&self, value: T);
    }
    ```

## `pub trait Parser<'a, T>`

```rs
// Parser 트레잇 정의
// - `'a T`: 트레이트의 제네릭 매개변수를 정의합니다.
//     - `'a`는 수명 매개변수로, 참조의 유효 기간을 나타냅니다.
//     - `T`: 타입 매개변수로, 파싱 결과의 타입을 나타냅니다.
pub trait Parser<'a, T> {
    // 연관 타입(associated type)을 정의합니다. 이를 통해 이 trait을 구현하는 타입은 자신만의 Error 타입을 지정할 수 있습니다.
    type Error;
    // Self? 이 트레이트를 구현하는 타입
    fn parse(&self, input: &'a [u8]) -> Result<T, Self::Error>;
}
```

- 수명 매개변수 (`'a`)

    Rust에서는 메모리 안전성을 보장하기 위해 수명(lifetime) 시스템을 사용합니다.
    수명 매개변수는 참조의 유효 기간을 명시적으로 나타내며, 컴파일러가 이 정보를 바탕으로 메모리 접근의 안전성을 검사합니다.

    `Parser` `trait`에서 `'a`는 입력 참조 `input: &'a [u8]`의 유효 기간을 나타냅니다.
    이는 파서가 입력 데이터를 파싱하는 동안 그 데이터가 유효해야 함을 보장합니다.
    여기서 `input: &'a [u8]` 데이터의 수명은 `'a`로 제한됩니다.
    즉, `input`이 유효하지 않게 되면 파서가 해당 데이터를 안전하게 사용할 수 없게 됩니다.

- 제네릭 타입 매개변수 (`T`)

    제네릭 타입 매개변수는 특정 타입에 종속되지 않는 코드를 작성할 수 있도록 합니다.
    이를 통해 다양한 타입에 대해 동일한 로직을 재사용할 수 있습니다.

    `T`는 파싱 결과의 타입을 나타냅니다.
    이 `trait`를 구현하는 타입은 `T`에 대해 구체적인 타입을 지정하여, 어떤 형태의 데이터를 파싱할 것인지를 명시할 수 있습니다.

- 연관 타입 (`type Error`)

    연관 타입(associated type)은 `trait`를 구현하는 타입이 정의해야 하는 타입을 지정합니다.
    이는 제네릭 인수를 사용하지 않고도 특정 타입을 추상화할 수 있는 방법을 제공합니다.

    `Parser` `trait`에서는 `Error`라는 연관 타입을 정의하여, 파싱이 실패할 경우 반환할 오류 타입을 명시합니다.
    `trait`를 구현하는 타입은 이 `Error` 타입을 자신의 상황에 맞게 정의해야 합니다.

- `Self` 키워드

    `Self`는 `trait`를 구현하는 타입을 가리킵니다.
    이는 객체 지향 프로그래밍의 `this`나 `self`와 비슷한 개념입니다.

    `parse` 메서드에서 사용된 `Self::Error`는 `Parser` `trait`를 구현하는 타입이 정의한 `Error` 타입을 가리킵니다.
    이를 통해 `parse` 메서드는 파싱 실패 시 해당 오류 타입을 반환할 수 있습니다.

- `&self`

    `parse` 메서드가 `Parser`를 구현하는 타입의 불변 참조로 호출됨을 의미합니다.

`Parser` `trait`을 실제로 구현한다면 다음과 같습니다:

```rs
// JSON 데이터를 파싱하는 파서. `Parser` 트레잇을 구현하는 구조체입니다.
pub struct JsonParser;

//                   ┌─ `T`는 `serde_json::Value` 타입으로, JSON 데이터를 파싱한 결과
impl<'a> Parser<'a, serde_json::Value> for JsonParser {
    // 이 구현에서는 `Error` 타입으로 `serde_json::Error`를 지정합니다.
    // 이는 JSON 파싱 중 발생할 수 있는 오류를 처리하기 위한 타입입니다.
    type Error = serde_json::Error;

    // 입력 데이터(`input`)의 수명을 `'a`로 지정하여, 데이터가 파싱하는 동안 유효해야 함을 보장
    fn parse(&self, input: &'a [u8]) -> Result<serde_json::Value, Self::Error> {
        let input_str = std::str::from_utf8(input).map_err(|e| serde_json::Error::custom(e.to_string()))?;
        serde_json::from_str(input_str)
    }
}
```
