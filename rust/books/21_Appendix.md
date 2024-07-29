# 21. Appendix

- [21. Appendix](#21-appendix)
    - [21.2. B - Operators and Symbols](#212-b---operators-and-symbols)
        - [Non-operator Symbols](#non-operator-symbols)
            - [문자열 리터럴](#문자열-리터럴)
                - [1. Raw String Literals (`r""`)](#1-raw-string-literals-r)
                - [2. Byte String Literals (`b""`)](#2-byte-string-literals-b)
                - [3. Byte Literals (`b'x'`)](#3-byte-literals-bx)
                - [4. 문자열 리터럴에 접두어 추가하기](#4-문자열-리터럴에-접두어-추가하기)

## 21.2. B - Operators and Symbols

### Non-operator Symbols

#### [문자열 리터럴](https://doc.rust-lang.org/reference/expressions/literal-expr.html)

```rs
"foo"; r"foo";                     // foo
"\"foo\""; r#""foo""#;             // "foo"

"foo #\"# bar";
r##"foo #"# bar"##;                // foo #"# bar

"\x52"; "R"; r"R";                 // R
"\\x52"; r"\x52";                  // \x52
```

##### 1. Raw String Literals (`r""`)

이스케이프 시퀀스를 무시하고 문자열을 그대로 표현하고자 할 때 사용한다.
즉, `\x22`와 같은 이스케이프 시퀀스는 문자열에서 문자 그대로 `\x22`로 처리되고, 특별한 의미를 갖지 않는다.
정규 표현식이나 파일 경로를 작성할 때 유용하다.

```rust
let regex = r"^\d+\.\d+$"; // 숫자.숫자 형태의 문자열을 찾는 정규 표현식
let path = r"C:\Program Files\MyApp"; // 파일 경로
```

`r"..."`에서 사용된 `...` 내의 모든 문자들은 그대로 해석되며, 이스케이프 시퀀스(`\n`, `\t` 등)는 무시된다.
이는 문자열 내에서 많은 이스케이프 시퀀스를 사용해야 하는 경우 유용하다.

```rs
// 성공: "{\x22key1\x22:\x22value2\x22,\x22key2\x22:\x22value2\x22,\x22key3\x22:\x22key3\x22}" 
// 실패: r"{\x22key1\x22:\x22value2\x22,\x22key2\x22:\x22value2\x22,\x22key3\x22:\x22key3\x22}"
pub struct JsonParser;
impl Parser for JsonParser {
    fn parse(&self, input: &str, charset: &EncodingType) -> Value {
        serde_json::from_str(input).unwrap_or(Value::Null)
    }
}
```

위의 주석에서 성공하는 경우와 실패하는 경우의 차이는 이스케이프 시퀀스가 어떻게 인식되는가의 차이다.
`\x22`는 큰따옴표(`"`)를 의미하는 이스케이프 시퀀스다.
원시 문자열 리터럴(`r"..."`)에서는 `\x22`가 문자 그대로 처리되어, 실제 JSON 문자열로 인식되지 않기 때문에 파싱이 실패한다.

##### 2. Byte String Literals (`b""`)

바이트 문자열을 표현하고자 할 때 사용한다. 이는 주로 ASCII 문자열을 바이트 배열로 처리할 때 유용하다.
`b"~"`로 표시된 바이트 문자열 리터럴(byte string literals)은 그 안에 포함된 `\xNN` 형식의 이스케이프 시퀀스를 해당 바이트 값으로 변환한다.
예를 들어, `\x22`는 쌍따옴표(`"`)를 나타내는 ASCII 코드인 34로 변환된다.

```rust
let bytes = b"hello"; // ASCII 문자열 "hello"를 바이트 배열로 표현
```

결과는 `&[u8]` 타입의 슬라이스가 된다.
`b"..."` 내의 모든 문자들은 ASCII 코드로 변환되어 바이트 배열의 요소가 된다.

##### 3. Byte Literals (`b'x'`)

단일 바이트를 표현한다.

```rust
let byte = b'a'; // ASCII 문자 'a'에 해당하는 바이트
```

결과는 `u8` 타입의 값이 된다. `b'x'`에서 `x`는 ASCII 문자여야 하며, 이를 해당하는 바이트 값으로 표현한다.

##### 4. 문자열 리터럴에 접두어 추가하기

`r#""#`, `br#""#` 등과 같이 `#` 기호를 사용하여 더 복잡한 문자열도 쉽게 표현할 수 있다.
이는 리터럴 내부에서 큰따옴표를 사용해야 하는 경우에 유용하다.

```rust
let complex = r#"{"key": "value"}"#; // JSON 문자열
let complex_bytes = br#"{"key": "value"}"#; // JSON 문자열을 바이트 배열로 표현
```
