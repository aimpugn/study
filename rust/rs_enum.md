# rust enum

- [rust enum](#rust-enum)
    - [대수적 데이터 타입(Algebraic Data Type, ADT)와 `enum`](#대수적-데이터-타입algebraic-data-type-adt와-enum)
    - [기타](#기타)

## 대수적 데이터 타입(Algebraic Data Type, ADT)와 `enum`

대수적 데이터 타입(Algebraic Data Type, ADT)은 수학의 대수학과 프로그래밍의 타입 시스템을 연결하는 개념으로,
복합 데이터 타입을 정의하는 방법을 제공합니다.

대수적 데이터 타입의 두 가지 주요 형태:
- 합 타입(Sum Type):

    'OR' 관계를 나타냅니다.
    Rust의 `enum`이 이에 해당합니다.

- 곱 타입(Product Type):

    'AND' 관계를 나타냅니다.
    Rust의 `struct`가 이에 해당합니다.

Rust의 열거형(`enum`)과 대수적 데이터 타입의 관계:

1. 합 타입으로서의 열거형:

    열거형의 각 variant는 서로 다른 경우를 나타냅니다.
    이는 수학적으로 합집합(union)과 유사합니다.

2. 값을 포함하는 열거형:

    Rust의 열거형은 각 variant에 다른 타입의 값을 포함할 수 있습니다.
    이는 합 타입과 곱 타입의 조합을 가능하게 합니다.

예시:

```rust
enum Result<T, E> {
    Ok(T),
    Err(E),
}
```

- `Result<T, E> = Ok(T) | Err(E)`
- 여기서 `|`는 합 연산을 나타냅니다.

수학적 관점:
- 만약 `T`가 가질 수 있는 값의 수가 $t$이고, `E`가 가질 수 있는 값의 수가 $e$라면,
- `Result<T, E>`가 가질 수 있는 전체 값의 수는 $t + e$ 입니다.

이러한 특성은 타입 시스템에 강력한 표현력을 제공합니다:
1. 타입 안전성: 컴파일러가 모든 가능한 경우를 처리했는지 확인할 수 있습니다.
2. 패턴 매칭: 대수적 데이터 타입은 패턴 매칭과 자연스럽게 결합됩니다.
3. 코드의 명확성: 데이터의 구조와 의미를 명확하게 표현할 수 있습니다.

## 기타

- [Rust 공식 문서 - 열거형과 패턴 매칭](https://doc.rust-lang.org/book/ch06-00-enums.html)
- [Haskell 위키 - 대수적 데이터 타입](https://wiki.haskell.org/Algebraic_data_type)
- [Cornell University CS3110 - 대수적 데이터 타입](https://www.cs.cornell.edu/courses/cs3110/2019sp/textbook/data/algebraic_data_types.html)
