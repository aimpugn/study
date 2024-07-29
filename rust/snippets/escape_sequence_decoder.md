# Escape Sequence Decoder

## 한글 처리 실패/성공 코드 비교

```rust
fn decode_escape_sequences(input: &str) -> String {
    let mut result = String::new();
    let mut chars = input.chars().peekable();

    while let Some(ch) = chars.next() {
        if is_hex_escape_start(&mut chars) {
            process_hex_escape(&mut chars, &mut result);
        } else {
            result.push(ch);
        }
    }

    result
}

fn is_hex_escape_start(chars: &mut std::iter::Peekable<std::str::Chars>) -> bool {
    chars.next_if_eq(&'\\').is_some() && chars.peek() == Some(&'x')
}

fn process_hex_escape(chars: &mut std::iter::Peekable<std::str::Chars>, result: &mut String) {
    // 'x' 문자 건너뛰기
    chars.next();

    let hex_chars: String = chars.take(2).collect();
    if hex_chars.len() == 2 {
        if let Ok(byte) = u8::from_str_radix(&hex_chars, 16) {
            result.push(byte as char);
        } else {
            result.push_str(&format!("\\x{}", hex_chars));
        }
    } else {
        result.push_str("\\x");
        result.push_str(&hex_chars);
    }
}
```

하지만 이 코드는 비정상적인 결과를 출력합니다.

```log
"for":"for":"storeI""store-123-123-123-123-123123""paymentI""some_id""orderNam""NEW BALANCEë\x8´\xE°\x9ë\x9\xE\xA4í\x9
                                                                                                                                                           \xE¦\xBí\x8
ë\xA¼\xE\x8ë\xA¤í\x8° | XSë\xB\xE\x8¬\xE©\xBë\xA\xBì\x9
\xE³\xAê\xB\xE\x8ê\xB\xE\x8ë\x8\xE\xB"}}}"totalAmoun""1290""taxFreeAmoun""""currenc""CURRENCY_KR""noticeUrl":"https://kr-api.xxxx.net/yyyy/dddd/done"]"origi":"userAgen""Mozilla/5.0 (iPhone; CPU iPhone OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E14""ur""nul""platformTyp""PLATFORM_TYPE_MOBIL""ipAddres""111.111.111.112"}"isEscro":true"product":[]}"channelSelecto":"selecto":"fiel""ke""valu":"ke""channel-key-xxx-123-123123-123123-d146a5ebb89e"}}}"appSchem""acloset:/""redirectUr""http://xxxx/yyyy/rn""confirmUr""https://kr-api.xxxx.net/yyyy/dddd/confirm"}"determinedChannelKe""channel-key-8b237670-5932-43f8-9556-d146a5ebb89""paymentMethodFor":"paymentMethodTyp":"fiel""car""valu":{}}}"returnUr":{}}
parse_to_json is done
{"\"for\":\"for\":\"storeI\"\"store-123-123-123-123-123123\"\"paymentI\"\"some_id\"\"orderNam\"\"NEW BALANCEë\\x8´\\xE°\\x9ë\\x9\\xE\\xA4í\\x9
                                                                                                                                                                                   \\xE¦\\xBí\\x8
```

```rust
fn decode_escape_sequences(input: &str) -> String {
    let mut result = String::new();
    let mut chars = input.chars().peekable();
    let mut byte_seq = Vec::new();

    while let Some(ch) = chars.next() {
        // 16진수 이스케이프 시퀀스 시작 여부를 확인
        let maybe_hex_escape_sequence = ch == '\\' && chars.peek() == Some(&'x');

        if !maybe_hex_escape_sequence {
            // 이스케이프 시퀀스가 연속되지 않고 다른 일반 문자로 이어지는 경우,
            // 이전에 수집된 바이트 시퀀스가 있다면,
            // 이를 UTF-8 문자열로 변환하여 result에 추가하고 바이트 시퀀스 비운다.
            if !byte_seq.is_empty() {
                result.push_str(&String::from_utf8_lossy(&byte_seq));
                byte_seq.clear();
            }
            // 이스케이프 시퀀스가 아닌 모든 문자를 결과 문자열에 포함시키기 위해
            // 이스케이프 시퀀스의 일부가 아닌 경우(즉, 일반 문자인 경우), 바로 result 문자열에 추가
            result.push(ch);
            continue;
        }

        // '\\' 이후 'x' 문자 건너뛰기
        chars.next();

        // 다음 첫 번째 문자가 없는 경우
        if chars.peek().is_none() {
            result.push_str("\\x");
            continue;
        }
        let hex_char1 = chars.next().unwrap();

        // 다음 두 번째 문자가 없는 경우
        if chars.peek().is_none() {
            result.push_str(&format!("\\x{}", hex_char1));
            continue;
        }
        let hex_char2 = chars.next().unwrap();

        let hex_chars = hex_char1.to_string() + &hex_char2.to_string();

        if let Ok(byte) = u8::from_str_radix(&hex_chars, 16) {
            // 16진수 문자열을 바이트로 변환 성공하면 byte_seq에 추가
            byte_seq.push(byte);
        } else {
            // 변환 실패 시, 원본 '\x' 이후의 문자열을 결과에 추가
            result.push_str(&format!("\\x{}", hex_chars));
        }
    }

    // 마지막에 남은 바이트 시퀀스가 있다면 문자열로 변환하여 결과에 추가
    if !byte_seq.is_empty() {
        result.push_str(&String::from_utf8_lossy(&byte_seq));
    }

    result
}
```

하지만 아래와 같이 나와야 정상입니다.

```log
{"form":{"form":{"storeId":"store-123-123-123-123-123123","paymentId":"some_id","orderName":"NEW BALANCE 뉴발란스 프린팅 민소매 티 | XS 블랙 A45/ 4-0527-030","customerForm":{"phoneNumber":"01012341234","email":"aaaa@ffff.ddd.com","customerName":{"name":{"field":"fullName","value":"옷사면뭐해입고갈때가없는데"}}},"totalAmount":"12900","taxFreeAmount":"0","currency":"CURRENCY_KRW","noticeUrls":["https://kr-api.acloset.net/transactions/portone/done"],"origin":{"userAgent":"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148","url":"null","platformType":"PLATFORM_TYPE_MOBILE","ipAddress":"111.111.111.112"},"isEscrow":true,"products":[]},"channelSelector":{"selector":{"field":"key","value":{"key":"channel-key-xxx-123-123123-123123-d146a5ebb89e"}}},"appScheme":"acloset://","redirectUrl":"http://xxxx/yyyy/rn","confirmUrl":"https://kr-api.xxxx.net/yyyy/dddd/confirm"},"determinedChannelKey":"channel-key-123-123-1231-123-1234","paymentMethodForm":{"paymentMethodType":{"field":"card","value":{}}},"returnUrl":{}}
```

왜 이런 차이가 발생하는 건가요?

### 연속된 이스케이프 시퀀스 처리

```rust
let mut byte_seq = Vec::new();
// ... (코드 생략)
if let Ok(byte) = u8::from_str_radix(&hex_chars, 16) {
    byte_seq.push(byte);
}
// ... (코드 생략)
if !byte_seq.is_empty() {
    result.push_str(&String::from_utf8_lossy(&byte_seq));
    //                       └─ 잘못된 UTF-8 시퀀스를 안전하게 처리
}
```

- 연속된 이스케이프 시퀀스를 바이트 시퀀스로 수집한 후 한 번에 UTF-8 문자열로 변환합니다.
- 이는 멀티바이트 문자(예: 한글)를 올바르게 처리할 수 있게 합니다.

하지만 실패하는 코드는 이런 처리가 되어 있지 않습니다.

```rust
if let Ok(byte) = u8::from_str_radix(&hex_chars, 16) {
    result.push(byte as char);
}
```

- 단일 바이트 처리:

    `byte as char`는 각 바이트를 개별적인 유니코드 코드 포인트로 해석합니다.
    UTF-8에서 한 문자는 1~4바이트로 구성될 수 있지만, 이 코드는 항상 1바이트만 처리합니다.

- 잘못된 문자 변환:

    0~127 범위의 ASCII 문자는 제대로 처리되지만, 그 이상의 값은 잘못된 문자로 변환됩니다.

    예를 들어, `0xA4`(164)는 유효한 UTF-8 시퀀스의 일부일 수 있지만, 이 코드는 이를 개별 문자로 처리합니다.

- 멀티바이트 문자 분리:

    한글과 같은 멀티바이트 문자는 여러 바이트로 구성되지만, 이 코드는 각 바이트를 별개의 문자로 취급합니다.
    각 이스케이프 시퀀스를 개별적으로 문자로 변환하기 때문에 멀티바이트 문자를 올바르게 처리하지 못할 수 있습니다.

    예를 들어, "안"(`0xEC 0x95 0x88`)이 `\xEC\x95\x88`으로 인코딩되었을 때, 이 코드는 이를 세 개의 개별 문자로 잘못 해석합니다.

    ```rust
    let input = r"\xEC\x95\x88"; // "안" in UTF-8
    let mut result = String::new();

    for i in (0..input.len()).step_by(4) {
        let hex_chars = &input[i+2..i+4];
        if let Ok(byte) = u8::from_str_radix(hex_chars, 16) {
            result.push(byte as char);
        }
    }

    println!("{}", result); // 출력: 쩃 (3개의 잘못된 문자)
    ```

    "안"이라는 한글 문자를 3개의 개별적인, 의미 없는 문자로 잘못 해석합니다.

- UTF-8 규칙 위반:

    UTF-8 인코딩 규칙을 고려하지 않고 각 바이트를 독립적으로 처리합니다.
    이는 잘못된 UTF-8 시퀀스를 생성할 수 있습니다.

아래 방식은 모든 바이트를 수집한 후 UTF-8 문자열로 한 번에 변환합니다.

```rust
let input = r"\xEC\x95\x88"; // "안" in UTF-8
let mut byte_seq = Vec::new();

for i in (0..input.len()).step_by(4) {
    let hex_chars = &input[i+2..i+4];
    if let Ok(byte) = u8::from_str_radix(hex_chars, 16) {
        byte_seq.push(byte);
    }
}

let result = String::from_utf8_lossy(&byte_seq);
println!("{}", result); // 출력: 안
```

- `String::from_utf8_lossy`를 사용하여 UTF-8 디코딩 규칙을 준수합니다.
- 멀티바이트 문자를 올바르게 구성합니다.

### UTF-8 디코딩

```rust
String::from_utf8_lossy(&byte_seq)
```

- `from_utf8_lossy`를 사용하여 잘못된 UTF-8 시퀀스를 안전하게 처리합니다.

하지만 실패하는 코드의 경우 UTF-8 디코딩 과정이 없어 멀티바이트 문자를 올바르게 처리하지 못합니다.

### 예외 처리

```rust
if chars.peek().is_none() {
    result.push_str("\\x");
    continue;
}
// ... (유사한 예외 처리 코드)
```

- 불완전한 이스케이프 시퀀스에 대해 더 강건한 처리를 제공합니다.

하지만 실패하는 코드의 경우 불완전한 이스케이프 시퀀스에 대한 처리가 덜 세밀합니다.

### 개선 제안

다음과 같은 최적화를 고려해볼 수 있습니다:

```rust
fn decode_escape_sequences(input: &str) -> String {
    // 결과 문자열을 위한 공간을 미리 할당하여 재할당 횟수를 줄임
    // 예: input = "hello\x41\x42world" (len = 15)
    // result의 초기 용량 = 15
    let mut result = String::with_capacity(input.len());

    // 유효한 UTF-8 바이트 시퀀스를 저장할 벡터
    // 예: "\x41\x42" 처리 시 [65, 66]을 저장
    let mut byte_seq = Vec::new();

    // 입력 문자열을 문자 단위로 순회할 반복자
    // peekable()을 사용하여 다음 문자를 미리 볼 수 있게 함
    let mut chars = input.chars().peekable();

    // 입력 문자열의 모든 문자를 순회
    while let Some(ch) = chars.next() {
        // 16진수 이스케이프 시퀀스 시작 여부를 확인
        // 예: ch = '\', next char = 'x'
        if ch == '\\' && chars.peek() == Some(&'x') {
            chars.next(); // 'x' 문자 건너뛰기
            // 다음 두 문자를 한 번에 가져와 처리
            // 예: "\x41" -> h1 = '4', h2 = '1'
            if let (Some(h1), Some(h2)) = (chars.next(), chars.next()) {
                let hex_chars = format!("{}{}", h1, h2);
                // 16진수 문자열을 u8로 변환 시도
                // 예: "41" -> Ok(65)
                if let Ok(byte) = u8::from_str_radix(&hex_chars, 16) {
                    // 유효한 16진수 시퀀스인 경우 바이트 시퀀스에 추가
                    // 예: byte_seq.push(65)
                    byte_seq.push(byte);
                    continue; // 다음 문자로 넘어감
                }
                // 유효하지 않은 16진수 시퀀스인 경우 원래 문자열 그대로 추가
                // 예: "\xGG" -> result에 "\\xGG" 추가
                result.push_str("\\x");
                result.push_str(&hex_chars);
            } else {
                // 불완전한 이스케이프 시퀀스 처리
                // 예: 입력의 끝에 "\x"가 있는 경우
                result.push_str("\\x");
            }
        } else {
            // 이스케이프 시퀀스가 아닌 경우
            if !byte_seq.is_empty() {
                // 누적된 바이트 시퀀스가 있으면 UTF-8로 디코딩하여 결과에 추가
                // 예: byte_seq = [65, 66] -> result에 "AB" 추가
                result.push_str(&String::from_utf8_lossy(&byte_seq));
                byte_seq.clear(); // 바이트 시퀀스 초기화
            }
            // 일반 문자는 그대로 결과에 추가
            // 예: ch = 'a' -> result에 'a' 추가
            result.push(ch);
        }
    }

    // 남은 바이트 시퀀스 처리
    // 예: 입력의 끝에 "\x41\x42"가 있었다면 여기서 "AB"가 result에 추가됨
    if !byte_seq.is_empty() {
        result.push_str(&String::from_utf8_lossy(&byte_seq));
    }

    // 최종 결과 반환
    result
}
```

이 코드의 동작을 예시와 함께 단계별로 설명하겠습니다:

1. 입력: `"Hello\x41\x42\xGGWorld\x"`

2. 처리 과정:
   - `result` 초기화: `""`
   - `byte_seq` 초기화: `[]`

3. 문자별 처리:
   - 'H': 일반 문자 -> `result` = `"H"`
   - 'e': 일반 문자 -> `result` = `"He"`
   - 'l': 일반 문자 -> `result` = `"Hel"`
   - 'l': 일반 문자 -> `result` = `"Hell"`
   - 'o': 일반 문자 -> `result` = `"Hello"`
   - '\x41': 이스케이프 시퀀스
     - `byte_seq` = `[65]`
     - `result`는 변경 없음
   - '\x42': 이스케이프 시퀀스
     - `byte_seq` = `[65, 66]`
     - `result`는 변경 없음
   - '\xGG': 잘못된 이스케이프 시퀀스
     - `byte_seq`의 내용을 처리: `result` = `"HelloAB"`
     - 잘못된 시퀀스 추가: `result` = `"HelloAB\xGG"`
   - 'W': 일반 문자 -> `result` = `"HelloAB\xGGW"`
   - 'o': 일반 문자 -> `result` = `"HelloAB\xGGWo"`
   - 'r': 일반 문자 -> `result` = `"HelloAB\xGGWor"`
   - 'l': 일반 문자 -> `result` = `"HelloAB\xGGWorl"`
   - 'd': 일반 문자 -> `result` = `"HelloAB\xGGWorld"`
   - '\x': 불완전한 이스케이프 시퀀스 -> `result` = `"HelloAB\xGGWorld\x"`

4. 최종 결과: `"HelloAB\xGGWorld\x"`

이 예시에서 볼 수 있듯이, 이 함수는:
- 유효한 이스케이프 시퀀스(`\x41`, `\x42`)를 적절히 디코딩합니다.
- 잘못된 이스케이프 시퀀스(`\xGG`)를 그대로 유지합니다.
- 불완전한 이스케이프 시퀀스(`\x` at the end)를 그대로 유지합니다.
- 일반 문자는 그대로 결과에 포함시킵니다.

이 코드의 주요 최적화 포인트와 개선사항:

1. `String::with_capacity(input.len())`:
   - 결과 문자열을 위한 공간을 미리 할당하여 재할당 횟수를 줄입니다.
   - 이는 특히 큰 입력 문자열을 처리할 때 성능을 향상시킵니다.

2. `if let (Some(h1), Some(h2)) = (chars.next(), chars.next())`:
   - 두 문자를 한 번에 처리하여 루프 반복 횟수를 줄입니다.
   - 조건문을 간소화하고 가독성을 높입니다.

3. `continue` 사용:
   - 유효한 16진수 시퀀스를 처리한 후 즉시 다음 반복으로 넘어갑니다.
   - 불필요한 조건 검사를 줄입니다.

4. 바이트 시퀀스 처리 최적화:
   - 이스케이프 시퀀스가 아닌 문자를 처리할 때만 누적된 바이트 시퀀스를 디코딩합니다.
   - 이는 불필요한 UTF-8 디코딩 작업을 줄입니다.

5. `push_str` 사용:
   - 문자열을 추가할 때 `push_str`을 사용하여 개별 문자 추가보다 효율적으로 처리합니다.

이 최적화된 버전은 원래 코드의 기능을 유지하면서도 성능을 개선하고 코드를 더 간결하게 만듭니다. 추가된 주석은 각 부분의 목적과 동작을 명확히 설명하여 코드의 이해도를 높입니다.
