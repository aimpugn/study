# JSON 여부

- [JSON 여부](#json-여부)
    - [JSON 표준(RFC 8259)](#json-표준rfc-8259)
    - [특정 형태의 문자열인 경우 JSON으로 간주](#특정-형태의-문자열인-경우-json으로-간주)
    - [무조건 파싱 시도](#무조건-파싱-시도)
    - [최적화, 성능 고려사항, 그리고 적용시 주의점](#최적화-성능-고려사항-그리고-적용시-주의점)

## JSON 표준([RFC 8259](https://tools.ietf.org/html/rfc8259))

JSON 표준([RFC 8259](https://tools.ietf.org/html/rfc8259))에 따르면, JSON의 유효한 형태는 다음과 같습니다.
1. 객체: `{...}`
2. 배열: `[...]`
3. 문자열: `"..."`
4. 숫자: `5`, `-3.14`, `1.2e+3` 등
5. 불리언: `true` 또는 `false`
6. null: `null`

> The literal names MUST be lowercase.  
> No other literal names are allowed.
>
> - value = false / null / true / object / array / number / string
>
> ```sh
>           f  a  l  s  e
> false = %x66.61.6c.73.65   ; false
> ```
>
> ```sh
>           n  u  l  l
> null  = %x6e.75.6c.6c      ; null
> ```
>
> ```sh
>           t  r  u  e
> true  = %x74.72.75.65      ; true
> ```
>
> ABNF(Augmented Backus-Naur Form) 표기법을 사용한 구문 정의입니다.
> `%x`는 16진수 표현을 나타내며, 각 쌍의 숫자는 하나의 ASCII 문자를 나타냅니다.
>
> 이는 JSON의 리터럴 값들("false", "null", "true")이 정확히 어떤 ASCII 문자로 구성되어야 하는지를 명확하게 정의합니다.
> 모든 문자가 소문자여야 함을 명시적으로 나타냅니다.
>
> 예를 들어, "FALSE", "True", "falsE", 또는 "nul" 등은 유효하지 않습니다.

## 특정 형태의 문자열인 경우 JSON으로 간주

```rust
use serde_json::{Value, Error as JsonError};

fn is_likely_json(input: &str) -> bool {
    let trimmed = input.trim();
    if trimmed.is_empty() {
        return false;
    }

    match trimmed.chars().next().unwrap() {
        '{' | '[' => trimmed.ends_with('}') || trimmed.ends_with(']'),
        '"' => trimmed.ends_with('"') && trimmed.len() > 1,
        't' => trimmed == "true",
        'f' => trimmed == "false",
        'n' => trimmed == "null",
        '0'..='9' | '-' => trimmed.parse::<f64>().is_ok(),
        _ => false,
    }
}

fn parse_to_json(input: &str) -> Result<Value, JsonError> {
    if is_likely_json(input) {
        serde_json::from_str(input)
    } else {
        Err(JsonError::syntax(serde_json::error::ErrorCode::ExpectedSomeValue, 0, 0))
    }
}

fn main() {
    let inputs = vec![
        r#"{"name": "Alice", "age": 30}"#,
        r#"[1, 2, 3]"#,
        r#""Hello, World!""#,
        "42",
        "-3.14",
        "true",
        "false",
        "null",
        "name=Alice&age=30",
    ];

    for input in inputs {
        println!("Input: {}", input);
        println!("Is likely JSON: {}", is_likely_json(input));
        println!("Parse result: {:?}", parse_to_json(input));
        println!();
    }
}
```

다만, 이 방법도 100% 정확하지는 않습니다.

예를 들어, `"true"` (따옴표로 묶인 문자열)도 유효한 JSON이지만, 이 구현에서는 불리언 `true`로 잘못 판단할 수 있습니다.

또한 간단한 검사만 수행하므로, 복잡한 JSON 구조(예: 중첩된 객체나 배열)에 대해서는 정확도가 떨어질 수 있습니다.

따라서 가장 확실한 방법은 여전히 실제로 파싱을 시도하는 것이지만, 큰 입력에 대해 성능 문제가 있을 수 있습니다.

## 무조건 파싱 시도

```rs
pub fn parse_to_json(input: &str) -> Result<Value, Box<dyn Error>> {
    // 무조건 json 
    match serde_json::from_str(input) {
        Ok(json) => Ok(json),
        Err(_) => {
            let mut map = serde_json::Map::new();
            for (key, value) in form_urlencoded::parse(input.as_bytes()) {
                insert_into_json_map(&mut map, &key, &value);
            }

            Ok(Value::Object(map))
        }
    }
}
```

## 최적화, 성능 고려사항, 그리고 적용시 주의점

1. **지연 파싱**: 대용량 JSON을 다룰 때는 전체를 즉시 파싱하는 대신 필요한 부분만 파싱하는 것이 효율적일 수 있습니다.

2. **스트리밍 파서 사용**: `serde_json`의 스트리밍 파서를 사용하면 전체 JSON을 메모리에 로드하지 않고도 파싱할 수 있습니다.

3. **사용자 정의 역직렬화**: 특정 구조를 자주 파싱한다면, 사용자 정의 역직렬화를 구현하여 성능을 향상시킬 수 있습니다.

4. **대용량 JSON 처리**: 위 방법은 전체 JSON을 메모리에 로드하므로 매우 큰 JSON 파일에는 적합하지 않을 수 있습니다.

5. **부분 파싱 불가능**: 위 방법은 전체 JSON이 유효해야 하므로, 부분적으로 유효한 JSON을 처리할 수 없습니다.

6. **성능 오버헤드**: 유효성 검사를 위해 전체 JSON을 파싱하므로, 단순히 형식만 확인하는 것보다 느릴 수 있습니다.

7. **Zero-copy 역직렬화**: `serde_json`의 고급 기능으로, 문자열을 복사하지 않고 직접 참조하여 성능을 향상시킬 수 있습니다.
