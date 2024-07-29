# Factory Method

## by Rust

### `Parser` 트레잇을 구현

이 패턴은 생성할 파서 타입을 결정하는 로직을 캡슐화하여, 사용자가 직접적으로 구체 클래스에 의존하지 않도록 만든다.

```rs
trait Parser {
    fn parse(&self, input: &str) -> Value;
}

struct JsonParser;
impl Parser for JsonParser {
    fn parse(&self, input: &str) -> Value {
        // JSON 파싱 로직
    }
}

struct FormURLEncodedParser;
impl Parser for FormURLEncodedParser {
    fn parse(&self, input: &str) -> Value {
        // Form URL Encoded 파싱 로직
    }
}

enum ParserType {
    Json,
    FormURLEncoded,
}

fn create_parser(parser_type: ParserType) -> Box<dyn Parser> {
    match parser_type {
        ParserType::Json => Box::new(JsonParser {}),
        ParserType::FormURLEncoded => Box::new(FormURLEncodedParser {}),
    }
}
```
