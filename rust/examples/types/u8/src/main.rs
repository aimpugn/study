fn main() {
    // 16진수 문자열을 u8로 변환
    let hex = "1A";
    // 문자열 슬라이스(&str)를 받아서 주어진 진법(radix)에 따라 해당 문자열을 u8 타입의 숫자로 변환
    match u8::from_str_radix(hex, 16) {
        Ok(n) => println!("16진수 '{}'는 10진수로 {}입니다.", hex, n),
        Err(e) => println!("변환 실패: {}", e),
    }

    // 2진수 문자열을 u8로 변환
    let binary = "10101100";
    match u8::from_str_radix(binary, 2) {
        Ok(n) => println!("2진수 '{}'는 10진수로 {}입니다.", binary, n),
        Err(e) => println!("변환 실패: {}", e),
    }

    // 8진수 문자열을 u8로 변환
    let octal = "254";
    match u8::from_str_radix(octal, 8) {
        Ok(n) => println!("8진수 '{}'는 10진수로 {}입니다.", octal, n),
        Err(e) => println!("변환 실패: {}", e),
    }

    // 잘못된 진법의 문자열을 변환하려고 시도
    let invalid = "1G"; // 16진수에서 'G'는 유효하지 않은 값입니다.
    match u8::from_str_radix(invalid, 16) {
        Ok(n) => println!("'{}'는 10진수로 {}입니다.", invalid, n),
        Err(e) => println!("변환 실패: {}", e),
    }
}
