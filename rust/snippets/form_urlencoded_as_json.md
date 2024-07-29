# form urlencoded as json

## 코드 수정 과정

```rs
fn insert_into_json_map(map: &mut Map<String, Value>, key: &str, value: &str) {
    let parts: Vec<&str> = key.split('[').collect();
    let main_key = parts[0];

    if parts.len() == 1 {
        map.insert(main_key.to_string(), Value::String(value.to_string()));
    } else {
        let entry = map
            .entry(main_key.to_string())
            .or_insert_with(|| Value::Object(Map::new()));
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
```

1. 배열 처리 방식:

    빈 서브키(`[]`)를 가진 배열 항목을 객체의 빈 문자열 키로 처리합니다.
    그 결과 `"arr": {"": [...]}`가 됩니다.

2. 중첩 구조 처리:

    모든 중첩 구조를 객체로 처리합니다.
    `arr[]`를 `arr` 객체의 빈 키로 해석합니다.

```log
  left: Object {"key1": String("value1"), "key2": String("value2"), "arr": Object {"": Array [String("item1"), String("item2")]}}
 right: Object {"key1": String("value1"), "key2": String("value2"), "arr": Array [String("item1"), String("item2")]}
```

```rs
fn insert_into_json_map(result: &mut Map<String, Value>, key: &str, value: &str) {
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
```

### 두 구현의 주요 차이점

1. 배열 처리:
   - 첫 번째 구현: 빈 서브키를 객체의 빈 문자열 키로 처리
   - 두 번째 구현: 빈 서브키를 배열로 올바르게 처리

2. 키 파싱:
   - 첫 번째 구현: 단순 분할 후 인덱스로 접근
   - 두 번째 구현: 패턴 매칭을 사용한 더 강건한 파싱

3. 에러 처리:
   - 첫 번째 구현: 암시적 에러 처리 (일부 케이스 무시)
   - 두 번째 구현: 명시적 에러 처리 (예상치 못한 형식에 대해 panic)

### 수정된 첫 번째 구현

문제를 해결하기 위해 첫 번째 구현을 다음과 같이 수정할 수 있습니다:

```rust
fn insert_into_json_map(map: &mut Map<String, Value>, key: &str, value: &str) {
    let parts: Vec<&str> = key.split('[').collect();
    let main_key = parts[0];

    if parts.len() == 1 {
        map.insert(main_key.to_string(), Value::String(value.to_string()));
    } else {
        let entry = map
            .entry(main_key.to_string())
            .or_insert_with(|| Value::Array(Vec::new()));
        if let Value::Array(ref mut arr) = entry {
            arr.push(Value::String(value.to_string()));
        }
    }
}
```

주요 변경 사항:
1. 중첩 구조를 항상 배열로 처리합니다.
2. 빈 서브키 검사를 제거하고, 모든 `[]` 형식을 배열로 처리합니다.

두 번째 구현(원래 코드)은 이러한 문제를 올바르게 처리하고 있어 테스트를 통과합니다.
이 구현은 Form URL 인코딩의 다양한 케이스를 더 정확하게 처리하며, 특히 배열과 중첩 객체의 구분을 명확히 합니다.

## 중첩된 구조의 맵 처리

`parent[child]=value1`, `parent[child1][child2]=value2` 같은 복잡한 구조를 처리할 수 있도록 함수 수정

```rust
use serde_json::{Map, Value};

fn insert_into_json_map(map: &mut Map<String, Value>, key: &str, value: &str) {
    let parts: Vec<&str> = key.split('[').collect();
    let main_key = parts[0];

    let mut current = map.entry(main_key.to_string()).or_insert_with(|| Value::Object(Map::new()));

    for part in parts.iter().skip(1) {
        let part = part.trim_end_matches(']');
        
        if part.is_empty() {
            // 배열 요소 추가
            if !current.is_array() {
                *current = Value::Array(Vec::new());
            }
            if let Value::Array(arr) = current {
                arr.push(Value::Null);
                current = arr.last_mut().unwrap();
            }
        } else {
            // 객체 속성 추가/접근
            if !current.is_object() {
                *current = Value::Object(Map::new());
            }
            if let Value::Object(obj) = current {
                current = obj.entry(part.to_string()).or_insert(Value::Null);
            }
        }
    }

    *current = Value::String(value.to_string());
}

fn parse_to_json(input: &str) -> Result<Value, Box<dyn std::error::Error>> {
    let mut map = Map::new();
    for pair in input.split('&') {
        let mut iter = pair.splitn(2, '=');
        if let (Some(key), Some(value)) = (iter.next(), iter.next()) {
            insert_into_json_map(&mut map, key, value);
        }
    }
    Ok(Value::Object(map))
}
```

1. 재귀적 구조 처리:
    - 키를 `[`로 분할하여 각 부분을 순회합니다.
    - 각 부분에 대해 객체 또는 배열 구조를 동적으로 생성/접근합니다.

2. 배열과 객체 구분:
    - 빈 부분(`[]`)은 배열 요소로 처리합니다.
    - 비어있지 않은 부분은 객체의 키로 처리합니다.

3. 동적 타입 전환:
    - 필요에 따라 `Value`를 `Object`나 `Array`로 전환합니다.

4. 중첩 구조 지원:
    - 임의의 깊이의 중첩된 객체와 배열을 처리할 수 있습니다.

## 테스트 케이스

```rust
#[test]
fn test_complex_nested_structure() {
    let input = "parent[child]=value1&parent[child1][child2]=value2&arr[]=item1&arr[]=item2";
    let result = parse_to_json(input).unwrap();
    let expected = serde_json::json!({
        "parent": {
            "child": "value1",
            "child1": {
                "child2": "value2"
            }
        },
        "arr": ["item1", "item2"]
    });
    assert_eq!(result, expected);
}
```

## 결론

1. Form URL 인코딩의 복잡성: 단순해 보이는 형식이지만, 중첩된 구조를 올바르게 처리하려면 신중한 접근이 필요합니다.

2. 동적 타입 처리의 중요성: JSON과 같은 동적 구조를 다룰 때는 런타임에 타입을 유연하게 전환할 수 있어야 합니다.

3. 재귀적 사고의 필요성: 중첩된 구조를 효과적으로 다루려면 재귀적 접근이 효과적입니다.

4. 엣지 케이스 고려: 다양한 입력 패턴을 고려하여 구현해야 합니다.

이 개선된 구현은 복잡한 중첩 구조를 포함한 다양한 Form URL 인코딩 패턴을 올바르게 처리할 수 있습니다. 그러나 실제 사용 시에는 추가적인 엣지 케이스와 에러 처리를 고려해야 할 수 있습니다.
