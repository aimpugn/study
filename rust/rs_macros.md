# Macros

- [Macros](#macros)
    - [매크로와 함수의 차이](#매크로와-함수의-차이)
    - [`macro_rules!` 사용한 선언적 매크로](#macro_rules-사용한-선언적-매크로)
    - [`println!()`](#println)
        - [`{}` 포맷](#-포맷)
        - [`{:p}` 포맷](#p-포맷)
        - [`{:?}` 또는 `{:#?}`](#-또는-)

`macro`는 Rust의 기능군(a family of features)을 가리킨다

- `macro_rules!` 사용한 *선언적* 매크로(*declarative* macros)
- 세 종류의 *절차적* 매크로(*procedural* macros)
    - Custom `#[derive]` macros: struct와 enum에 사용되는 `derive`와 함께 추가되는 코드 명시
    - Attribute-like macros: 모든 아이템에 사용할 수 있는 custom attributes
    - Function-like macros: 함수 호출 같지만 인자로 지정된 토큰에 작동

## [매크로와 함수의 차이](https://doc.rust-lang.org/book/ch19-06-macros.html#the-difference-between-macros-and-functions)

매크로는 *metaprogramming*이라는, 다른 코드를 작성하는 코드를 작성하는 방법이다.
`derive` 속성, `println!`과 `vec!` 매크로들은 직접 작성한 코드보다 더 많은 코드를 만들어 낸다

metaprogramming은 코드의 양을 줄이고 관리하는 데 유용하다. 함수도 비슷한 역할을 하지만, 매크로는 함수가 갖지 않은 추가적인 강력함이 있다

- 함수
    - 함수가 가져야 하는 파라미터의 개수와 타입을 반드시 선언해야 한다.
    - 함수는 런타임에 호출되고 trait은 컴파일 시에 구현되어야 하기 때문에 주어진 타입에 trait 구현할 수 없다
- 매크로
    - `println!("hello")`,  `println!("hello {}", name)` 처럼 가변적인 개수의 파라미터를 받을 수 있다
    - 컴파일러가 코드의 의미를 해석하기 전에 확장되므로, 매크로는 주어진 타입에 대한 trait을 구현할 수 있다
    - Rust 코드를 작성하는 Rust 코드를 작성해야 하므로, 함수를 정의하는 것보다 매크로를 정의하는 것이 더 복잡하며, 이 때문에 매크로 정의를 읽고, 이해하고, 관리하기가 더 어렵다.
    - 반드시 파일에서 매크로를 호출하기 ***전***에 매크로를 정의하거나 scope로 가져와야 한다

## `macro_rules!` 사용한 선언적 매크로

`match` 표현식을 작성하는 것과 비슷하다. 값을 특정 토드와 연관된 패턴과 비교한다. 이 경우 그 값은 매크로로 전달된 문자열 Rust 소스 코드다. 즉, 러스트 소스코드와 패턴을 비교하고, 만약 매칭된다면, 각 패턴과 연관된 러스트 코드는 매크로로 전달된 코드를 대체한다. 이 모든 것은 컴파일 동안 발생한다.

매크로를 정의하기 위해, `macro_rules!`를 사용한다

```rs
#[cfg(all(not(no_global_oom_handling), not(test)))]
#[macro_export]
^^^^^^^^^^^^^^^
매크로가 정의된 crate가 scope로 들어올 때마다, 이 매크로가 사용할 수 있어야 함을 나타낸다. 이 어노테이션이 없으면 매크로를 scope로 가져올 수 없다.
#[stable(feature = "rust1", since = "1.0.0")]
#[rustc_diagnostic_item = "vec_macro"]
#[allow_internal_unstable(rustc_attrs, liballoc_internals)]
macro_rules! vec {
             ^^^
             `!` 없이 선언된다
    () => (
        $crate::__rust_force_expr!($crate::vec::Vec::new())
    );
    ($elem:expr; $n:expr) => (
        $crate::__rust_force_expr!($crate::vec::from_elem($elem, $n))
    );
     ⎾ 패턴과 매칭되는 러스트 소스 코드를 갖는 변수를 `$` 사용하여 선언
    ($($x:expr),+ $(,)?) => (
    ^ │ │     │        ^ 전체 패턴을 감싸는 괄호
      └─│─ 대체하는 코드에서 사용하기 위해 괄호 안의 패턴과 매칭되는 값들을 캡처하기 위한 괄호
        └── 모든 러스트 표현식과 매칭되며 그 식에 `$x`라는 이름을 부여한다

        $crate::__rust_force_expr!(<[_]>::into_vec(
            #[rustc_box]
            $crate::boxed::Box::new([$($x),+])
        ))
    );
}

vec![1, 2, 3];
    ^^^^^^^^^
    `1`, `2`, `3` 표현식과 `$x` 패턴이 세 번 매칭된다
```

- 매크로의 패턴 문법은 러스트 코드 구조에 대해 매칭해야 하기 때문에, 일반적인 러스트 코드상의 패턴 문법과는 다르다
    - [일반적인 코드상에서의 패턴 문법](https://doc.rust-lang.org/book/ch18-03-pattern-syntax.html)
    - [매크로의 패턴 문법](https://doc.rust-lang.org/reference/macros-by-example.html)
- [The Little Book of Rust Macros](https://veykril.github.io/tlborm/)

## `println!()`

- [Formatted print](https://doc.rust-lang.org/rust-by-example/hello/print.html)
- [Module std::fmt](https://doc.rust-lang.org/std/fmt/)

### `{}` 포맷

### `{:p}` 포맷

- 포인터 출력

```rs
let x_layout = Layout::new::<i32>();
let x_ptr = alloc(x_layout);
// x_ptr type: `*mut u8`
println!("{:p}", x_ptr);
```

### `{:?}` 또는 `{:#?}`

- `?`: [Debug](https://doc.rust-lang.org/std/fmt/trait.Debug.html)
- `#`을 붙이면 pretty print
