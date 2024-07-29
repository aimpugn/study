# encoded to json

## 직관적이고 한 파일로 끝나는 코드

```toml
[package]
name = "url_encoded_to_json"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
url = "2.3"
percent-encoding = "2.2"
encoding_rs = "0.8"
clap = { version = "4.0", features = ["derive"] }
```

```rs
// src/main.rs
use encoding_rs::{EUC_KR, UTF_8, ISO_8859_1, Encoding, CoderResult};
use percent_encoding::percent_decode;
use std::error::Error;
use std::borrow::Cow;
use url::form_urlencoded;
use clap::Parser;
use serde_json::Value;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Input string to convert
    input: String,

    /// Source encoding (utf-8, euc-kr, iso-8859-1)
    #[arg(short, long, default_value = "utf-8")]
    encoding: String,
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let decoded = decode_input(&args.input, &args.encoding)?;
    let json = parse_to_json(&decoded)?;
    println!("{}", serde_json::to_string_pretty(&json)?);
    Ok(())
}

fn decode_input(input: &str, encoding: &str) -> Result<String, Box<dyn Error>> {
    // Percent decode the input to bytes
    let percent_decoded = percent_decode(input.as_bytes()).collect::<Vec<u8>>();

    // Decode from the specified encoding to UTF-8
    let (cow, _, had_errors) = match encoding {
        "utf-8" => UTF_8.decode(&percent_decoded),
        "euc-kr" => EUC_KR.decode(&percent_decoded),
        "iso-8859-1" => ISO_8859_1.decode(&percent_decoded),
        _ => return Err("Unsupported encoding".into()),
    };

    if had_errors {
        return Err("Decoding error".into());
    }

    Ok(unescape_string(&cow))
}

fn unescape_string(input: &str) -> String {
    input.replace("\\x22", "\"")
         .replace("\\x5C", "\\")
         .replace("\\x2F", "/")
}

fn parse_to_json(input: &str) -> Result<Value, Box<dyn Error>> {
    if input.trim_start().starts_with('{') {
        Ok(serde_json::from_str(input)?)
    } else {
        let mut map = serde_json::Map::new();
        for (key, value) in form_urlencoded::parse(input.as_bytes()) {
            insert_into_json_map(&mut map, &key, &value);
        }
        Ok(Value::Object(map))
    }
}

fn insert_into_json_map(map: &mut serde_json::Map<String, Value>, key: &str, value: &str) {
    let parts: Vec<&str> = key.split('[').collect();
    let main_key = parts[0];
    
    if parts.len() == 1 {
        map.insert(main_key.to_string(), Value::String(value.to_string()));
    } else {
        let entry = map.entry(main_key.to_string()).or_insert_with(|| Value::Object(serde_json::Map::new()));
        if let Value::Object(ref mut obj) = entry {
            let sub_key = parts[1].trim_end_matches(']');
            if sub_key.is_empty() {
                let arr = obj.entry("").or_insert_with(|| Value::Array(Vec::new()));
                if let Value::Array(ref mut vec) = arr {
                    vec.push(Value::String(value.to_string()));
                }
            } else {
                obj.insert(sub_key.to_string(), Value::String(value.to_string()));
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_form_urlencoded_to_json() {
        let input = "key1=value1&key2=value2&arr[]=item1&arr[]=item2";
        let expected = r#"{"key1":"value1","key2":"value2","arr":["item1","item2"]}"#;
        let result = parse_to_json(input).unwrap();
        assert_eq!(serde_json::to_string(&result).unwrap(), expected);
    }

    #[test]
    fn test_json_with_escaped_quotes() {
        let input = r#"{\x22key\x22:\x22value\x22}"#;
        let expected = r#"{"key":"value"}"#;
        let decoded = unescape_string(input);
        let result = parse_to_json(&decoded).unwrap();
        assert_eq!(serde_json::to_string(&result).unwrap(), expected);
    }

    #[test]
    fn test_euc_kr_decoding() {
        let input = "%C8%AB%B1%E6%B5%BF=30";  // "홍길동=30" in EUC-KR
        let decoded = decode_input(input, "euc-kr").unwrap();
        assert_eq!(decoded, "홍길동=30");
    }

    #[test]
    fn test_decode_input_euc_kr() {
        // "안녕하세요"를 EUC-KR로 인코딩한 후 URL 인코딩한 값
        let input = "%BE%C8%B3%E7%C7%CF%BC%BC%BF%E4";
        let result = decode_input(input, "euc-kr").unwrap();
        assert_eq!(result, "안녕하세요");
    }

    #[test]
    fn test_decode_input_utf8() {
        // "안녕하세요"를 UTF-8로 URL 인코딩한 값
        let input = "%EC%95%88%EB%85%95%ED%95%98%EC%84%B8%EC%9A%94";
        let result = decode_input(input, "utf-8").unwrap();
        assert_eq!(result, "안녕하세요");
    }

    #[test]
    fn test_decode_input_with_query_params() {
        let input = "key=%BE%C8%B3%E7&arr[]=value2&somemap[key3]=value3";
        let result = decode_input(input, "euc-kr").unwrap();
        assert_eq!(result, "key=안녕&arr[]=value2&somemap[key3]=value3");
    }

    #[test]
    fn test_decode_input_with_mixed_encoding() {
        // "hello"는 ASCII, "안녕"은 EUC-KR
        let input = "hello=%BE%C8%B3%E7&key2=value2";
        let result = decode_input(input, "euc-kr").unwrap();
        assert_eq!(result, "hello=안녕&key2=value2");
    }
}
```

## 멀티 스레드

```toml
// Cargo.toml
[package]
name = "url_encoded_to_json"
version = "0.2.0"
edition = "2021"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
url = "2.3"
percent-encoding = "2.2"
encoding_rs = "0.8"
clap = { version = "4.0", features = ["derive"] }
rayon = "1.5"
```

```rs
// src/main.rs
use clap::Parser;
use encoding_rs::{EUC_KR, UTF_8};
use percent_encoding::percent_decode_str;
use rayon::prelude::*;
use serde_json::Value;
use std::error::Error;
use std::fs::File;
use std::io::{BufRead, BufReader};
use url::form_urlencoded;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Input file path
    input: String,

    /// Source encoding (utf-8, euc-kr)
    #[arg(short, long, default_value = "utf-8")]
    encoding: String,
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let file = File::open(&args.input)?;
    let reader = BufReader::new(file);

    let encoding = args.encoding.as_str();

    reader.lines()
        .par_bridge()  // 병렬 이터레이터로 변환
        .map(|line| {
            let line = line?;
            let decoded = decode_input(&line, encoding)?;
            let json = parse_to_json(&decoded)?;
            Ok(serde_json::to_string(&json)?)
        })
        .try_for_each(|result: Result<String, Box<dyn Error + Send + Sync>>| {
            match result {
                Ok(json_line) => println!("{}", json_line),
                Err(e) => eprintln!("Error processing line: {}", e),
            }
            Ok(())
        })?;

    Ok(())
}


fn decode_input(input: &str, encoding: &str) -> Result<String, Box<dyn Error>> {
    // Percent decode the input to bytes
    let percent_decoded = percent_decode(input.as_bytes()).collect::<Vec<u8>>();

    // Decode from the specified encoding to UTF-8
    let (cow, _, had_errors) = match encoding {
        "utf-8" => UTF_8.decode(&percent_decoded),
        "euc-kr" => EUC_KR.decode(&percent_decoded),
        "iso-8859-1" => ISO_8859_1.decode(&percent_decoded),
        _ => return Err("Unsupported encoding".into()),
    };

    if had_errors {
        return Err("Decoding error".into());
    }

    Ok(unescape_string(&cow))
}


fn unescape_string(input: &str) -> String {
    input.replace("\\x22", "\"")
         .replace("\\x5C", "\\")
         .replace("\\x2F", "/")
}

fn parse_to_json(input: &str) -> Result<Value, Box<dyn Error>> {
    if input.trim_start().starts_with('{') {
        Ok(serde_json::from_str(input)?)
    } else {
        let mut map = serde_json::Map::new();
        for (key, value) in form_urlencoded::parse(input.as_bytes()) {
            insert_into_json_map(&mut map, &key, &value);
        }
        Ok(Value::Object(map))
    }
}

fn insert_into_json_map(map: &mut serde_json::Map<String, Value>, key: &str, value: &str) {
    let parts: Vec<&str> = key.split('[').collect();
    let main_key = parts[0];
    
    if parts.len() == 1 {
        map.insert(main_key.to_string(), Value::String(value.to_string()));
    } else {
        let entry = map.entry(main_key.to_string()).or_insert_with(|| Value::Object(serde_json::Map::new()));
        if let Value::Object(ref mut obj) = entry {
            let sub_key = parts[1].trim_end_matches(']');
            if sub_key.is_empty() {
                let arr = obj.entry("").or_insert_with(|| Value::Array(Vec::new()));
                if let Value::Array(ref mut vec) = arr {
                    vec.push(Value::String(value.to_string()));
                }
            } else {
                obj.insert(sub_key.to_string(), Value::String(value.to_string()));
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_unescape_string() {
        assert_eq!(unescape_string("\\x22Hello\\x22"), "\"Hello\"");
    }

    #[test]
    fn test_parse_to_json_form_urlencoded() {
        let input = "key1=value1&key2=value2&arr[]=item1&arr[]=item2";
        let result = parse_to_json(input).unwrap();
        let expected = serde_json::json!({
            "key1": "value1",
            "key2": "value2",
            "arr": ["item1", "item2"]
        });
        assert_eq!(result, expected);
    }

    #[test]
    fn test_parse_to_json_json() {
        let input = r#"{"key":"value","nested":{"array":[1,2,3]}}"#;
        let result = parse_to_json(input).unwrap();
        let expected = serde_json::json!({
            "key": "value",
            "nested": {
                "array": [1, 2, 3]
            }
        });
        assert_eq!(result, expected);
    }
}
```
