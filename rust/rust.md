# Rust

- [Rust](#rust)
    - [vscode settings](#vscode-settings)
    - [package layout](#package-layout)
    - [`crate` and `modules`](#crate-and-modules)
        - [crate](#crate)
        - [modules](#modules)
    - [memory](#memory)
        - [`&var`와 `var.as_ref()`](#var와-varas_ref)
        - [`*mut T`](#mut-t)
        - [Strict Provenance](#strict-provenance)
        - [profiling](#profiling)
            - [flamegraph](#flamegraph)
    - [Attributes](#attributes)
        - [Built-in attributes index](#built-in-attributes-index)
    - [`?` operator](#-operator)
    - [reference, dereference, and mutability](#reference-dereference-and-mutability)
        - [reference(`&`)](#reference)
        - [dereference(`*`)](#dereference)

## vscode settings

- `rust-analyzer` 설치
- rust 구현 소스 코드 보기

```shell
rustup component add rust-src
# info: component 'rust-src' is up to date
```

- 만약 std 라이브러리로 진입하지 못한다면, [러스트 커뮤니티 쓰레드](https://users.rust-lang.org/t/solved-how-to-step-into-std-source-code-when-debugging-in-vs-code/25319) 참고

## package layout

- [Package Layout](https://doc.rust-lang.org/cargo/guide/project-layout.html)

## `crate` and `modules`

- [Defining modules to control scope and privacy](https://doc.rust-lang.org/book/ch07-02-defining-modules-to-control-scope-and-privacy.html)
- [`crate` and `modules`](https://www.cs.brandeis.edu/~cs146a/rust/doc-02-21-2015/book/crates-and-modules.html)

`crate` is synonymous with a library or package in other languages. Hence "Cargo" as the name of Rust's package management tool: you ship your crates to others with Cargo. Crates can produce:
1. an executable(`src/main.rs` binary)
2. a shared library(`src/lib.rs` library)

Each `crate` has an implicit root `module` that contains the code for that crate. You can then define a tree of sub-modules under that root module. Modules allow you to partition your code within the crate itself.

```shell
├── phrases # name of crate
    ├── english # module
    │   ├── farewells # module
    │   └── greetings # module
    └── korean # module
        ├── farewells # module
        └── greetings # module
```

```rust
mod english {
    mod greetings {}
    mod farewells {}
}
mod korean {
    mod greetings {}
    mod farewells {}
}
```

### [crate](https://doc.rust-lang.org/rust-by-example/crates.html)

[`crate`는 compilation, linking, versioning, distribution, 그리고 runtime loading 단위](https://doc.rust-lang.org/reference/crates-and-source-files.html)

`rustc some_file.rs` 실행 시, `some_file.rs`는 `crate` 파일
`some_file.rs`가 `mod`로 선언된 코드를 갖고 있다면, 컴파일러가 실행되기 전에 모듈 파일의 컨텐츠는 `crate` 파일에서 `mod`가 선언된 곳에 삽입된다.
즉, 모듈은 개별적으로 컴파일되지 않으며, `crate`만 컴파일된다

`crate`는 binary 또는 library로 컴파일될 수 있다. `rustc`는 `crate`로 바이너리를 만들며, 이는 `--crate-type` 플래그로 조절할 수 있다.

[`crate root`](https://doc.rust-lang.org/book/ch07-02-defining-modules-to-control-scope-and-privacy.html)는 라이브러리인 경우 `src/lib.rs`, 바이너리인 경우 `src/main.rs`

### [modules](https://doc.rust-lang.org/reference/items/modules.html)

> It is encouraged to use the new naming convention as it is more consistent, and avoids having many files named `mod.rs` within a project.

[module and file hierarchy](https://doc.rust-lang.org/reference/items/modules.html)

- module path: `crate`
- file path: `lib.rs`
- file content: `mod util;`

- module path: `crate::util`
- file path: `util.rs`
- file content: `mod config;`

- module path: `crate::util::config.rs`
- file path: `util/config.rs`
- file content: NONE

## memory

### [`&var`와 `var.as_ref()`](https://users.rust-lang.org/t/what-is-the-difference-between-as-ref/76059/4)

- `&`
    - deref 강제를 통해서 이뤄지며, `Deref` trait을 구현했을 때만 작동한다
    - 변수의 값에 대한 참조를 리턴한다
- `as_ref()`
    - `Option<&T>`를 리턴하므로, `Option`을 활용할 수 있다
    - `AsRef::<Target>::as_ref()` 경우에는 저렴하게 타입 컨버전을 할 수 있따

### `*mut T`

- `*mut T`
    - 타입 `T`의 값에 대한 mutable 원시 포인터(raw pointer)를 의미한다.
    - `mut`: mutable, 즉 포인터가 가리키는 값이 변할 수 있음 의미
    - `*`: 포인터 타입을 나타낸다
- `ptr`이 `*mut T` 타입인 경우, `*ptr`
    - `*`: 역참조 연산자(dereference operator)로 사용되며, `ptr`이 가리키는 메모리 공간의 값에 접근할 수 있다.

```rs
//               dereference raw pointer
let x = unsafe { *ptr };
//^^^ value of `T` type at the memory location is assigned to `x`
```

### Strict Provenance

- [rust-lang doc](https://doc.rust-lang.org/nightly/std/ptr/index.html#strict-provenance)
- [Tracking Issue for strict_provenance](https://github.com/rust-lang/rust/issues/95228)
- [FAQ](https://github.com/rust-lang/rust/issues/95228#issuecomment-1075881238)

[Strict Provenance by Aria the cat, twitter thread](https://twitter.com/gankra_/status/1509335249871900678)
Strict Provenance is framed as "a new memory model" but it's actually a set of library APIs that make your code *trivially correct under all coherent memory models*. If you use these APIs "correctly", then it doesn't *matter* what the memory model is, your code supports it!

HUGE CAVEAT TO THE ABOVE: Strict Provenance is mostly about *Unsafe* Rust code. As in, code that's just using raw pointers (*mut T). If you mix in references (&T) then "Stacked Borrows" get layered on top and introduces extra rules that these APIs don't "solve".

Ok! So! What Is Strict Provenance?
In essence, *stop casting integers to pointers*.
Actually ideally assume `ptr as usize` and `usize as ptr` don't exist at all.
That's it. That's the whole model.

Of course, there's lots of code that casts integers to pointers and it's *useful* and sometimes *necessary* so this on its own is an untenable position. This is why strict_provenance the *feature* isn't actually a memory model but a set of APIs that help you do that!

Yes, there is *absolutely nothing* special or magical about strict_provenance (yet). The "key" operation of with_addr is something you could always do, and which you may have seen recommended for people having trouble with miri getting confused about provenance.

People *more* in the know may be familiar with the proposed solution for C, PNVI-ae-udi (which almost assuredly everyone else will implicitly inherit): [A Provenance-aware Memory Object Model for C](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n2676.pdf)

### profiling

#### [flamegraph](https://github.com/flamegraph-rs/flamegraph)

```shell
cargo install flamegraph
```

```shell
# generate binary
cargo build

# dtrace: system integrity protection is on, some features will not be available
# dtrace: failed to initialize dtrace: DTrace requires additional privileges
# 위와 같은 에러가 나면서 실패하지 않으려면, `sudo` 권한 사용
sudo cargo flamegraph --bin ostep  -- --track-allocations
```

## [Attributes](https://doc.rust-lang.org/reference/attributes.html)

inner와 outer 속성이 있으며, 차이점은 작동 방식이 아닌, 적용되는 `item`이다.

- InnerAttribute: `item` 정의 안에 위치한다.
    - `# ! [ Attr ]`: `!`(bang)이 붙어 있다
    - 속성이 선언된 항목에 적용
    - `crate`에 속성을 적용하는 유일한 방법.
- OuterAttribute: `item` 정의 바깥에 위치한다.
    - `# [ Attr ]`: `!`(bang)이 없다
    - 속성 다음에 오는 것에 적용
    - `struct`, `module` 앞에 적용.

[스택오버플로 답변](https://stackoverflow.com/a/27455138)에 의하면:
1. `crate` attribute는 둘러싸는 컨텍스트에 적용되는 속성이므로 InnerAttribute(`#![...]`) 사용하고,
2. `struct`, `module`, `function` 등에 대해선 OuterAttribute(`#[...]`) 사용

### [Built-in attributes index](https://doc.rust-lang.org/reference/attributes.html#built-in-attributes-index)

- Conditional complilation
- Testing
- Derive
- Macros
- Diagnostics
- ABI, linking, symbols, and FFI

## `?` operator

```rs
// the `?` operator can only be applied to values
// that implement `Try` the trait `Try` is not implemented for `&mut
```

## reference, dereference, and mutability

### reference(`&`)

1. `&T`(immutable): read from the referenced value
2. `&mut T`(mutable): read and modify the referenced value.

```rs
let x = 42;
let y = &x; // y is an immutable reference to x
let z = &mut x; // error! cannot borrow `x` as mutable, as it is not declared as mutable
```

reference rules
1. one mutable reference XOR multiple immutable references
2. Mutable and immutable references cannot coexist at the same time.

```rs
let mut x = 42;
let y = &x; // y is an immutable reference to x
let z = &mut x; // z is a mutable reference to x
```

### dereference(`*`)
