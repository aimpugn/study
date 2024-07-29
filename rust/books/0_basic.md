# 0. Basic

## Rust 설치하기

Rust의 공식 설치 가이드([The Rust Programming Language](https://www.rust-lang.org/learn/get-started))를 참조하여 Rust 개발 환경을 설정

- Rust는 rustup을 통해 설치합니다. rustup은 Rust 버전 관리와 관련 툴체인을 관리하는 툴입니다.
- 설치 방법은 플랫폼(Windows, macOS, Linux)에 따라 다를 수 있습니다. 보통 다음 명령어를 사용합니다:

    ```bash
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
    ```

## 첫 Rust 프로그램

- 설치가 완료되면, 간단한 "Hello, world!" 프로그램을 작성하여 Rust 개발 환경이 제대로 구성되었는지 확인합니다.
- 터미널에서 다음 명령어를 사용하여 새로운 프로젝트를 생성합니다:

    ```bash
    cargo new hello_world
    cd hello_world
    ```

- `src/main.rs` 파일을 열고, 기본적으로 생성된 코드를 확인합니다:

    ```rust
    fn main() {
        println!("Hello, world!");
    }
    ```

- 프로그램을 실행하기 위해 터미널에서 다음 명령어를 입력합니다:

    ```bash
    cargo run
    ```

- 성공적으로 "Hello, world!"가 출력되면, Rust 개발 환경이 제대로 구성된 것입니다.

## 기본 문법과 개념

- **변수와 불변성**: Rust에서는 변수가 기본적으로 불변(immutable)입니다. 값을 변경하려면 `mut` 키워드를 사용해야 합니다.

    ```rust
    let x = 5; // 불변 변수
    let mut y = 5; // 가변 변수
    y = 6; // 가능
    ```

- **데이터 타입**: Rust는 정적 타입 언어로, 모든 변수의 타입이 컴파일 시에 결정됩니다. 타입 추론이 가능하지만, 때로는 타입을 명시해야 할 때도 있습니다.

    ```rust
    let guess: u32 = "42".parse().expect("Not a number!");
    ```

- **함수**: Rust에서 함수는 `fn` 키워드로 선언합니다. 파라미터의 타입은 반드시 명시해야 합니다.

    ```rust
    fn add_two(x: i32) -> i32 {
        x + 2
    }
    ```

- **소유권**: Rust의 가장 중요한 개념 중 하나로, 메모리 안전성을 보장합니다. 변수의 데이터는 해당 변수가 스코프 밖으로 벗어날 때 자동으로 정리됩니다.
