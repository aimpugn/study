# Dependency Injection

- [Dependency Injection](#dependency-injection)
    - [의존성 주입](#의존성-주입)
    - [by Rust](#by-rust)
        - [`DataProcessor`](#dataprocessor)

## 의존성 주입

의존성 주입은 객체 간의 결합도를 낮추고, 유연성 및 테스트 용이성을 향상시키기 위한 기법이다.
객체가 필요로 하는 의존성(종속 객체)을 외부에서 주입받음으로써, 객체는 자신의 의존성을 직접 생성하거나 관리할 필요가 없게 된다.
이 방식은 주로 "구조"의 변화에 초점을 맞추며, 객체가 서로 느슨하게 연결될 수 있도록 한다.
의존성 주입을 통해, 특히 테스트 시에 모의 객체(mock object)를 주입하는 것이 용이해지며, 유닛 테스트의 품질이 향상된다.

## by Rust

### `DataProcessor`

의존성 주입을 통해 `Parser` 구현체를 사용하는 곳에 직접 주입할 수 있다.
이는 테스트 용이성을 향상시키고, 느슨한 결합을 제공한다.

```rs
struct DataProcessor {
    parser: Box<dyn Parser>,
}

impl DataProcessor {
    fn process_data(&self, input: &str) -> Value {
        self.parser.parse(input)
    }
}
```
