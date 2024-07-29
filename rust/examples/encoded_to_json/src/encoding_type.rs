use std::fmt;
use std::str::FromStr;

#[derive(Clone, Debug, PartialEq)]
pub enum EncodingType {
    Utf8,
    EucKr,
    Iso8859_1,
    // 추가 인코딩...
}

// 사용자 친화적인 문자열 표현을 위해 Display 트레이트를 구현
impl fmt::Display for EncodingType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            EncodingType::Utf8 => write!(f, "UTF-8"),
            EncodingType::EucKr => write!(f, "EUC-KR"),
            EncodingType::Iso8859_1 => write!(f, "ISO-8859-1"),
            // 추가 인코딩...
        }
    }
}

// 문자열에서 EncodingType으로의 변환을 지원하기 위해 FromStr 트레이트를 구현
impl FromStr for EncodingType {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "utf-8" | "utf8" => Ok(EncodingType::Utf8),
            "euc-kr" | "euckr" => Ok(EncodingType::EucKr),
            "iso-8859-1" | "iso8859-1" => Ok(EncodingType::Iso8859_1),
            // 추가 인코딩...
            _ => Err(format!("Unknown encoding: {}", s)),
        }
    }
}
