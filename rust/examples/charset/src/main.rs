use encoding_rs::*;

fn main() {
    // UTF-8로 인코딩된 "동" 문자열
    let utf8_text = "동";

    // 1단계: UTF-8에서 Unicode 코드 포인트로 변환
    // Rust의 문자열은 이미 UTF-8로 인코딩되어 있으므로, 이 단계는 별도의 변환 과정 없이 직접 Unicode 코드 포인트를 확인할 수 있습니다.
    for ch in utf8_text.chars() {
        println!("UTF-8 '동'의 Unicode 코드 포인트: U+{:04X}", ch as u32);
    }

    // 2단계: Unicode에서 EUC-KR로 변환
    // `encoding_rs` 크레이트를 사용하여 변환
    let (cow, _encoding_used, _had_errors) = EUC_KR.encode(utf8_text);

    // 변환된 EUC-KR 바이트 시퀀스 출력
    println!("EUC-KR로 변환된 '동':");
    for byte in cow.iter() {
        println!("{:02X}", byte);
    }
}
