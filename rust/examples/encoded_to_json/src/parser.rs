use crate::encoding_type::EncodingType;
use encoding_rs::EUC_KR;
use percent_encoding::percent_decode;
use serde_json::{Map, Value};
use std::{fmt, str::from_utf8};
use url::form_urlencoded;

pub trait Parser<'a, T, E> {
    fn parse(&self, input: &'a [u8]) -> Result<T, E>;
}

pub struct DataProcessor<'a, T, E> {
    parser: Box<dyn Parser<'a, T, E> + Send + Sync>,
}

impl<'a, T, E> DataProcessor<'a, T, E> {
    pub fn new(parser: Box<dyn Parser<'a, T, E> + Send + Sync>) -> Self {
        DataProcessor { parser }
    }

    pub fn process(&self, input: &'a [u8]) -> Result<T, E> {
        self.parser.parse(input)
    }
}

#[derive(Debug)]
pub enum ParserError {
    Utf8Error(std::str::Utf8Error),
    JsonError(serde_json::Error),
    DecodingError(String),
    // 다른 에러 타입들...
}

impl std::error::Error for ParserError {}

impl fmt::Display for ParserError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            ParserError::Utf8Error(e) => write!(f, "UTF-8 error: {}", e),
            ParserError::JsonError(e) => write!(f, "JSON error: {}", e),
            ParserError::DecodingError(s) => write!(f, "Decoding error: {}", s),
            // 다른 에러 타입들에 대한 처리...
        }
    }
}

impl From<std::str::Utf8Error> for ParserError {
    fn from(error: std::str::Utf8Error) -> Self {
        ParserError::Utf8Error(error)
    }
}

impl From<serde_json::Error> for ParserError {
    fn from(error: serde_json::Error) -> Self {
        ParserError::JsonError(error)
    }
}

pub fn decode_input(
    input: &[u8],
    encoding: EncodingType,
) -> Result<String, Box<dyn std::error::Error>> {
    match encoding {
        EncodingType::Utf8 => {
            let decoded = percent_decode(input).decode_utf8()?.to_string();
            Ok(decode_escape_sequences(&decoded))
        }
        EncodingType::EucKr => {
            let decoded_bytes = percent_decode(input).collect::<Vec<u8>>();
            let (cow, _, had_errors) = EUC_KR.decode(&decoded_bytes);
            if had_errors {
                Err("EUC-KR decoding error".into())
            } else {
                Ok(decode_escape_sequences(&cow))
            }
        }
        EncodingType::Iso8859_1 => {
            // ISO-8859-1 implementation
            todo!()
        }
    }
}

// JsonParser 구현체
pub struct JsonParser;

impl<'a> Parser<'a, Value, ParserError> for JsonParser {
    fn parse(&self, input: &'a [u8]) -> Result<Value, ParserError> {
        let utf8_str_result = from_utf8(input);
        match utf8_str_result {
            Ok(utf8_str) => {
                let unescaped = decode_escape_sequences(utf8_str);
                Result::Ok(serde_json::from_slice(unescaped.as_bytes()).unwrap_or(Value::Null))
            }
            Err(e) => {
                eprintln!("Error: {}", e);
                Result::Err(ParserError::from(e))
            }
        }
    }
}

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

fn decode_escape_sequences2(input: &str) -> String {
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

// FormURLEncodedParser 구현체
pub struct FormURLEncodedParser;

impl<'a> Parser<'a, Value, ParserError> for FormURLEncodedParser {
    fn parse(&self, input: &'a [u8]) -> Result<Value, ParserError> {
        let utf8_str_result = from_utf8(input);
        match utf8_str_result {
            Ok(utf8_str) => {
                println!("utf8_str:{}", utf8_str);
                let decoded = decode_escape_sequences(utf8_str);
                println!("decoded:{}", decoded);
                let pairs = form_urlencoded::parse(decoded.as_bytes());
                let mut result = serde_json::Map::new();

                for (key, value) in pairs.into_owned() {
                    add_or_update_value(&mut result, &key, &value);
                }

                Result::Ok(Value::Object(result))
            }
            Err(e) => {
                eprintln!("Error: {}", e);
                Result::Err(ParserError::from(e))
            }
        }
    }
}

fn add_or_update_value(result: &mut Map<String, Value>, key: &str, value: &str) {
    // parts examples:
    // key=value: ["key"]
    // a[]=value: ["a", "]"]
    // foo[bar]=value: ["foo", "bar]"]
    let parts: Vec<&str> = key.split('[').collect();

    let (main_key, sub_key) = match parts.as_slice() {
        // key=value
        [main_key] => (main_key, None),
        // key[]=value OR key[sub_key]=value
        [main_key, sub_key] => (main_key, Some(sub_key.trim_end_matches(']'))),
        _ => panic!("Unexpected key format"),
    };

    match sub_key {
        None => {
            result.insert(main_key.to_string(), Value::String(value.to_string()));
        }
        Some(sub_key) => {
            if sub_key.is_empty() {
                // 배열 처리
                let array = result
                    .entry(main_key.to_string())
                    .or_insert_with(|| Value::Array(Vec::new()));

                if let Value::Array(ref mut arr) = array {
                    arr.push(Value::String(value.to_string()));
                }
            } else {
                // 중첩된 객체 처리
                let nested_object = result
                    .entry(main_key.to_string())
                    .or_insert_with(|| Value::Object(Map::new()));

                if let Value::Object(ref mut nested_map) = nested_object {
                    nested_map.insert(sub_key.to_string(), Value::String(value.to_string()));
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_decode_no_escapes() {
        assert_eq!(decode_escape_sequences("hello"), "hello");
    }

    #[test]
    fn test_decode_simple_escape() {
        assert_eq!(decode_escape_sequences("\\x48ello"), "Hello");
    }

    #[test]
    fn test_decode_multiple_escapes() {
        assert_eq!(decode_escape_sequences("H\\x65l\\x6Co\\x21"), "Hello!");
    }

    #[test]
    fn test_decode_unfinished_escape() {
        assert_eq!(decode_escape_sequences("hello\\x6"), "hello\\x6");
    }

    #[test]
    fn test_decode_invalid_hex() {
        // 유효하지 않은 16진수 처리에 대한 예상되는 결과
        // 여기서는 실패한 변환을 무시하고 나머지 문자열을 그대로 반환한다고 가정합니다.
        assert_eq!(decode_escape_sequences("hello\\xGG"), "hello\\xGG");
    }

    #[test]
    fn test_decode_adjacent_escapes() {
        assert_eq!(
            decode_escape_sequences("\\x48\\x65\\x6C\\x6C\\x6F"),
            "Hello"
        );
    }

    #[test]
    fn test_decode_mixed_content() {
        assert_eq!(
            decode_escape_sequences("Data: \\x44\\x61\\x74\\x61"),
            "Data: Data"
        );
    }

    #[test]
    fn test_decode_escapes_with_normal_text() {
        assert_eq!(
            decode_escape_sequences("\\x54his is \\x61 test\\x21"),
            "This is a test!"
        );
    }

    #[test]
    fn test_empty_input() {
        assert_eq!(decode_escape_sequences(""), "");
    }

    #[test]
    fn test_only_escape() {
        assert_eq!(decode_escape_sequences("\\x41\\x42\\x43"), "ABC");
    }
}
