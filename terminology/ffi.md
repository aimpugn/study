# FFI

- [FFI](#ffi)
    - [FFI (Foreign Function Interface)](#ffi-foreign-function-interface)

## FFI (Foreign Function Interface)

FFI는 한 프로그래밍 언어에서 작성된 코드가 다른 프로그래밍 언어로 작성된 함수나 서비스를 호출할 수 있도록 해주는 메커니즘입니다.

가령 [`libc` crate 페이지](https://crates.io/crates/libc)에는 다음과 같이 설명하고 있습니다.

> `libc` - Raw FFI bindings to platforms' system libraries

여기서 FFI는 Rust가 C로 작성된 라이브러리와 인터페이스할 수 있는 능력을 말합니다.
시스템 라이브러리는 파일 관리, 프로세스 제어, 네트워크 통신과 같은 기본적인 서비스를 제공하는 함수와 절차의 모음입니다. 이러한 라이브러리는 일반적으로 C로 작성되며 운영 체제의 핵심 부분입니다.
Rust의 `libc` crate은 이러한 C 라이브러리 함수와 대응되는 Rust 함수를 정의하여 바인딩을 제공합니다.

FFI를 사용하면, Rust 프로그램에서 새로운 기능을 직접 작성할 필요 없이 C 라이브러리에 이미 존재하는 기능을 사용할 수 있습니다.
