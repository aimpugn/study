# ADT

- [ADT](#adt)
    - [대수적 데이터 타입(ADTs)과 더 안전하고 표현력 있는 코드 작성](#대수적-데이터-타입adts과-더-안전하고-표현력-있는-코드-작성)
        - [타입 안전성 향상](#타입-안전성-향상)
        - [도메인 모델링 개선](#도메인-모델링-개선)
        - [패턴 매칭과의 시너지](#패턴-매칭과의-시너지)
        - [런타임 오류 감소](#런타임-오류-감소)
        - [코드 재사용성 향상](#코드-재사용성-향상)
        - [타입 추론 개선](#타입-추론-개선)
    - [기타](#기타)

## 대수적 데이터 타입(ADTs)과 더 안전하고 표현력 있는 코드 작성

대수적 데이터 타입(ADTs)이 타입 시스템의 이해와 안전하고 표현력 있는 코드 작성에 도움이 됩니다.

### 타입 안전성 향상

ADTs는 컴파일러가 모든 가능한 케이스를 처리했는지 확인할 수 있게 합니다.

예를 들어, Rust의 `match` 표현식에서 모든 열거형 변형을 처리하지 않으면 컴파일러가 경고를 줍니다.

```rust
enum TrafficLight {
    Red,
    Yellow,
    Green,
}

fn handle_light(light: TrafficLight) {
    match light {
        TrafficLight::Red => println!("Stop"),
        TrafficLight::Green => println!("Go"),
        // Yellow case not handled, compiler will warn
    }
}
```

- [Rust 공식 문서 - 패턴의 망라성](https://doc.rust-lang.org/book/ch18-03-pattern-syntax.html#matching-named-variables)

### 도메인 모델링 개선

ADTs를 사용하면 복잡한 도메인 개념을 정확히 모델링할 수 있습니다.
이는 코드가 실제 문제 도메인을 더 잘 반영하게 해줍니다.

```rust
// 여러 형태를 하나의 타입으로 표현하면서도 각 형태에 필요한 데이터를 정확히 포함합니다
enum Shape {
    Circle(f64),
    Rectangle(f64, f64),
    Triangle(f64, f64, f64),
}
```

### 패턴 매칭과의 시너지

ADTs는 패턴 매칭과 결합하여 강력한 제어 흐름 구조를 제공합니다.
이는 코드의 가독성과 유지보수성을 향상시킵니다.

```rust
fn area(shape: Shape) -> f64 {
    match shape {
        Shape::Circle(r) => std::f64::consts::PI * r * r,
        Shape::Rectangle(w, h) => w * h,
        Shape::Triangle(a, b, c) => {
            let s = (a + b + c) / 2.0;
            (s * (s - a) * (s - b) * (s - c)).sqrt()
        }
    }
}
```

### 런타임 오류 감소

ADTs를 사용하면 많은 런타임 오류를 컴파일 타임 오류로 변환할 수 있으며, 이를 통해 프로그램의 안정성을 크게 향상시킵니다.

예를 들어, null 참조 대신 `Option<T>`를 사용하면 null 포인터 예외를 방지할 수 있습니다:

```rust
fn divide(a: f64, b: f64) -> Option<f64> {
    if b == 0.0 {
        None
    } else {
        Some(a / b)
    }
}
```

- [Rust 공식 문서 - Option 열거형](https://doc.rust-lang.org/std/option/enum.Option.html)

### 코드 재사용성 향상

ADTs를 사용하면 공통된 동작을 가진 여러 타입을 하나의 인터페이스로 처리할 수 있습니다.

### 타입 추론 개선

ADTs는 컴파일러의 타입 추론 능력을 향상시킵니다.
이는 코드를 더 간결하게 만들면서도 타입 안전성을 유지할 수 있게 합니다.

## 기타

- [Cardelli and Wegner - On Understanding Types, Data Abstraction, and Polymorphism](https://www.cs.cmu.edu/~crary/819-f09/Cardelli85.pdf)
