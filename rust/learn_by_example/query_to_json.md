# Query to JSON

## 최초 코드

```rs
use serde_json::{Map, Value};
use url::form_urlencoded;

fn query_to_json(query: &str) -> Value {
    let pairs = form_urlencoded::parse(query.as_bytes());
    let mut result = serde_json::Map::new();

    for (key, value) in pairs.into_owned() {
        let parts: Vec<&str> = key.split("[").collect();

        match parts.len() {
            1 => {
                // 단순 키-값 쌍
                result.insert(key, Value::String(value));
            }
            _ => {
                // 배열 또는 중첩된 객체
                let main_key = parts[0];
                let sub_key = parts[1].trim_end_matches(']');

                if result.contains_key(main_key) {
                    update_nested_object(&mut result, main_key, sub_key, &value);
                } else {
                    create_nested_object(&mut result, main_key, sub_key, &value);
                }
            }
        }
    }

    Value::Object(result)
}

// 기존 JSON 객체에 새 값을 추가하거나 업데이트합니다.
fn update_nested_object(
    result: &mut Map<String, Value>,
    main_key: &str,
    sub_key: &str,
    value: &str,
) {
    let entry = result
        .entry(main_key.to_string())
        .or_insert_with(|| Value::Object(Map::new()));

    if let Value::Object(ref mut map) = entry {
        match sub_key.is_empty() {
            true => {
                // 배열 처리
                let array = map
                    .entry(main_key.to_string())
                    .or_insert_with(|| Value::Array(Vec::new()));
                if let Value::Array(ref mut arr) = array {
                    arr.push(Value::String(value.to_string()));
                }
            }
            false => {
                // 중첩된 객체 처리
                map.insert(sub_key.to_string(), Value::String(value.to_string()));
            }
        }
    }
}

// 새 배열 또는 중첩된 JSON 객체를 생성하고 초기화합니다.
fn create_nested_object(
    result: &mut Map<String, Value>,
    main_key: &str,
    sub_key: &str,
    value: &str,
) {
    match sub_key.is_empty() {
        true => {
            // 배열인 경우
            let mut arr = Vec::new();
            arr.push(Value::String(value.to_string()));
            result.insert(main_key.to_string(), Value::Array(arr));
        }
        false => {
            // 중첩된 객체인 경우
            let mut map = Map::new();
            map.insert(sub_key.to_string(), Value::String(value.to_string()));
            result.insert(main_key.to_string(), Value::Object(map));
        }
    }
}
```

## 첫번째 리팩토링

```rs
fn query_to_json(query: &str) -> Value {
    let pairs = form_urlencoded::parse(query.as_bytes());
    let mut result = serde_json::Map::new();

    for (key, value) in pairs.into_owned() {
        add_or_update_value(&mut result, &key, &value);
    }

    Value::Object(result)
}

fn add_or_update_value(result: &mut Map<String, Value>, key: &str, value: &str) {
    let parts: Vec<&str> = key.split('[').collect();
    let (main_key, sub_key) = match parts.as_slice() {
        [main_key] => (main_key, None),
        [main_key, sub_key] => (main_key, Some(sub_key.trim_end_matches(']'))),
        _ => panic!("Unexpected key format"),
    };

    match sub_key {
        None => {
            result.insert(main_key.to_string(), Value::String(value.to_string()));
        }
        Some(sub_key) => {
            let entry = result
                .entry(main_key.to_string())
                .or_insert_with(|| Value::Object(Map::new()));
            if let Value::Object(ref mut map) = entry {
                if sub_key.is_empty() {
                    // 배열 처리
                    let array = map
                        .entry(main_key.to_string())
                        .or_insert_with(|| Value::Array(Vec::new()));
                    if let Value::Array(ref mut arr) = array {
                        arr.push(Value::String(value.to_string()));
                    }
                } else {
                    // 중첩된 객체 처리
                    map.insert(sub_key.to_string(), Value::String(value.to_string()));
                }
            }
        }
    }
}
```
