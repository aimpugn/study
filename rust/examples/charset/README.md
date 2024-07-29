모든 hex 이스케이프 시퀀스를 적절히 처리하려면 더 복잡한 로직이 필요합니다.

## 1. 외부 라이브러리 사용

[`serde_json`](https://crates.io/crates/serde_json) 라이브러리를 사용하면 JSON 문자열 내의 이스케이프 시퀀스를 쉽게 처리할 수 있습니다.

```rust
use serde_json::Value;

fn unescape_string(input: &str) -> Result<String, serde_json::Error> {
    // JSON 문자열로 감싸서 파싱
    let json_str = format!("\"{}\"", input.replace("\"", "\\\""));
    let value: Value = serde_json::from_str(&json_str)?;
    Ok(value.as_str().unwrap().to_string())
}

fn main() {
    let input = r#"\xEB\x89\xB4\xEB\xB0\x9C\xEB\x9E\x80\xEC\x8A\xA4"#;
    match unescape_string(input) {
        Ok(result) => println!("Unescaped: {}", result),
        Err(e) => eprintln!("Error: {}", e),
    }
}
```

이 방법의 장점:
- 안정적이고 잘 테스트된 라이브러리 사용
- JSON 표준을 따르는 이스케이프 처리
- 다양한 이스케이프 시퀀스 처리 가능

단점:
- 추가 의존성 필요
- JSON 파싱 오버헤드 발생 가능

## 2. 직접 구현

더 세밀한 제어가 필요하거나 의존성을 최소화하고 싶다면, 직접 구현할 수 있습니다.

```rust
fn unescape_string(input: &str) -> Result<String, Box<dyn std::error::Error>> {
    let mut result = String::new();
    let mut chars = input.chars().peekable();

    while let Some(&ch) = chars.peek() {
        match ch {
            '\\' => {
                chars.next(); // '\'를 소비
                match chars.next() {
                    Some('x') => {
                        let hex: String = chars.by_ref().take(2).collect();
                        if hex.len() == 2 {
                            let byte = u8::from_str_radix(&hex, 16)?;
                            result.push(byte as char);
                        } else {
                            return Err("Invalid hex escape".into());
                        }
                    }
                    Some('u') => {
                        let hex: String = chars.by_ref().take(4).collect();
                        if hex.len() == 4 {
                            let code = u32::from_str_radix(&hex, 16)?;
                            if let Some(ch) = std::char::from_u32(code) {
                                result.push(ch);
                            } else {
                                return Err("Invalid Unicode escape".into());
                            }
                        } else {
                            return Err("Invalid Unicode escape".into());
                        }
                    }
                    Some(ch) => result.push(ch),
                    None => return Err("Unexpected end of string".into()),
                }
            }
            _ => {
                result.push(ch);
                chars.next();
            }
        }
    }

    Ok(result)
}

fn main() {
    let input = r#"\xEB\x89\xB4\xEB\xB0\x9C\xEB\x9E\x80\xEC\x8A\xA4"#;
    match unescape_string(input) {
        Ok(result) => println!("Unescaped: {}", result),
        Err(e) => eprintln!("Error: {}", e),
    }
}
```

이 구현의 장점:
- 외부 의존성 없음
- 세밀한 제어 가능
- \x 및 \u 이스케이프 시퀀스 모두 처리

단점:
- 복잡성 증가
- 모든 edge case를 처리하기 위해 추가 작업 필요할 수 있음

## 선택 가이드

1. 프로젝트가 이미 `serde_json`을 사용 중이거나 JSON 처리가 필요한 경우: 첫 번째 방법 사용
2. 의존성을 최소화하고 싶거나 특정 이스케이프 시퀀스만 처리하고 싶은 경우: 두 번째 방법 사용 및 필요에 맞게 수정

두 방법 모두 "\xEB\x89\xB4\xEB\xB0\x9C\xEB\x9E\x80\xEC\x8A\xA4"와 같은 hex 이스케이프 시퀀스를 올바르게 처리할 수 있습니다.

```

이 두 가지 접근 방식 중 어느 것을 선택할지는 프로젝트의 요구 사항과 제약 조건에 따라 달라집니다. `serde_json`을 사용하는 방법은 더 안정적이고 표준을 따르지만, 직접 구현하는 방법은 더 유연하고 의존성을 줄일 수 있습니다.

"Error: Error("invalid escape", line: 1, column: 146)" 같은 에러는 JSON 파싱 중에 발생한 것으로 보입니다. 이는 JSON 문자열 내에 올바르게 이스케이프 처리되지 않은 문자가 있을 때 발생할 수 있습니다. 위의 두 방법 중 하나를 사용하여 이스케이프 시퀀스를 먼저 처리한 후 JSON 파싱을 시도하면 이 문제를 해결할 수 있을 것입니다.

추가로, 만약 입력 문자열이 완전한 JSON이 아니라 JSON의 일부(예: 필드 값)라면, 해당 부분만 추출하여 처리하는 로직이 필요할 수 있습니다. 이 경우 정규 표현식을 사용하거나 문자열을 직접 파싱하는 방법을 고려해볼 수 있습니다.
