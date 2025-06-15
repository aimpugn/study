# [std::string::String](https://doc.rust-lang.org/std/string/struct.String.html)

- [std::string::String](#stdstringstring)
    - [from\_utf8\_lossy의 동작 원리](#from_utf8_lossy의-동작-원리)
        - [대안적 접근 방식](#대안적-접근-방식)

## from_utf8_lossy의 동작 원리

`String::from_utf8_lossy`는 바이트 슬라이스를 UTF-8 문자열로 변환하는 함수입니다. 그 동작 원리는 다음과 같습니다:

- 유효한 UTF-8 시퀀스 경우: 정상적으로 UTF-8 문자로 변환됩니다.

- 잘못된 UTF-8 시퀀스 경우: U+FFFD REPLACEMENT CHARACTER (�)로 대체됩니다.

- 처리 과정:
    - 바이트 슬라이스를 순회하며 UTF-8 디코딩을 시도합니다.
    - 유효한 시퀀스는 그대로 유지됩니다.
    - 잘못된 시퀀스를 만나면 해당 부분을 � 문자로 대체하고 계속 진행합니다.

```rust
use std::str::from_utf8_lossy;

fn main() {
    // 유효한 UTF-8
    let valid = [240, 159, 146, 150];
    println!("Valid: {}", from_utf8_lossy(&valid));  // 출력: "💖"

    // 잘못된 UTF-8 포함
    let invalid = [0, 159, 146, 150];
    println!("Invalid: {}", from_utf8_lossy(&invalid));  // 출력: "����"

    // 혼합된 경우
    let mixed = [240, 159, 146, 150, 0, 65, 66, 67];
    println!("Mixed: {}", from_utf8_lossy(&mixed));  // 출력: "💖�ABC"
}
```

`from_utf8_lossy`를 사용하는 이유는 다음과 같습니다.

1. 안전성:
    - 패닉 없이 항상 유효한 UTF-8 문자열을 생성합니다.
    - 데이터 손실 가능성을 감수하고 프로그램의 안정성을 우선시합니다.

2. 데이터 보존:
    - 가능한 한 많은 원본 데이터를 유지하려고 합니다.
    - 잘못된 부분만 대체하고 나머지는 보존합니다.

3. 디버깅 용이성:
    - `�` 문자의 존재로 잘못된 UTF-8 시퀀스의 위치를 쉽게 파악할 수 있습니다.

4. 성능:
    - 최적화된 구현으로 대부분의 경우 효율적입니다.

### 대안적 접근 방식

때에 따라 `from_utf8_lossy` 대신 다른 방법을 선택할 수 있습니다:

1. `std::str::from_utf8`: UTF-8 검증만 수행하고 오류 시 실패합니다.
2. `String::from_utf8`: 소유권을 가진 String을 생성하며, 실패 시 원본 바이트를 반환합니다.
3. 직접 구현: 특정 요구사항에 맞는 사용자 정의 디코딩 로직을 만듭니다.

    ```rust
    fn custom_decode(bytes: &[u8]) -> String {
        bytes.iter().map(|&b| b as char).collect()
    }
    ```
