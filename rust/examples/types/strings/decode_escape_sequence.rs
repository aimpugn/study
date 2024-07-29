// `\xNN` 형식의 이스케이프 시퀀스를 원래의 문자나 데이터로 변환하는 디코딩 함수
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
