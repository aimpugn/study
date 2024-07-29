# cargo

- [cargo](#cargo)
    - [cargo?](#cargo-1)
    - [initialize new Cargo manifest](#initialize-new-cargo-manifest)
    - [build cache](#build-cache)
    - [cargo build](#cargo-build)
    - [cargo install](#cargo-install)
    - [Cargo 업데이트](#cargo-업데이트)

## cargo?

Cargo는 Rust의 패키지 매니저이며, Rust 설치 시 함께 설치됩니다.

## [initialize new Cargo manifest](https://doc.rust-lang.org/cargo/commands/cargo-init.html)

```shell
CARGO_INIT_OPTIONS=(
    --bin # Create a package with a binary target
    --name ostep # set package name
)
cargo init $CARGO_INIT_OPTIONS
```

## [build cache](https://doc.rust-lang.org/cargo/guide/build-cache.html)

Cargo stores the output of a build into the `target` directory

```shell
├── debug # output for `dev` profile
│   ├── build
│   ├── deps
│   ├── examples
│   └── incremental
├── release # output for `release` profile
└── foo # output for `foo` profile when `--profile=foo`
```

## cargo build

Compile local packages and all of their dependencies.

```bash
cargo build
```

## cargo install

```sh
❯ cargo install
error: Using `cargo install` to install the binaries from the package in current working directory is no longer supported, use `cargo install --path .` instead. Use `cargo build` if you want to simply build the package.
```

과거에는 `cargo install`을 현재 디렉토리에서 실행하면, 해당 디렉토리의 프로젝트를 빌드하고 그 바이너리를 시스템에 설치했습니다.

이로 인해 다음과 같은 상황들로 인해 혼란이 있었다고 합니다:
- 사용자가 단순히 프로젝트를 빌드하려고 했지만, 실수로 `cargo install`을 실행하여 시스템에 설치할 수 있었습니다.
- 다른 프로젝트의 의존성을 설치하려다가 현재 디렉토리의 프로젝트를 설치할 수 있었습니다.

**변경 전**:

```sh
# 의도: crates.io에서 'serde' 설치
# 실제 결과: 현재 디렉토리의 프로젝트가 설치됨
$ cd my_project
$ cargo install serde
```

**변경 후**:

```sh
# crates.io에서 'serde' 설치
$ cargo install serde

# 현재 디렉토리의 프로젝트를 빌드하고 그 바이너리를 시스템에 설치합니다.
# 이전 `cargo install`의 동작과 동일합니다.
$ cargo install --path .
```

`--path` 옵션을 사용하면 Cargo는 다음과 같이 동작합니다:
1. 지정된 경로의 `Cargo.toml` 파일을 명시적으로 찾아 해당 프로젝트의 정보를 읽습니다.
2. 이 프로젝트가 실행 파일(binary)을 생성하는지 확인합니다.
3. 프로젝트를 빌드하고, 생성된 실행 파일을 Cargo의 bin 디렉토리(보통 `~/.cargo/bin/`)에 복사합니다.

`--path` 옵션을 요구함으로써 다음과 같은 이점이 있습니다:
- 사용자가 명확히 "이 경로의 프로젝트를 설치하겠다"고 지정해야 합니다.
- 실수로 현재 디렉토리의 프로젝트를 설치하는 일을 방지합니다.
- 다른 경로의 프로젝트도 쉽게 설치할 수 있습니다.

## Cargo 업데이트

Cargo (및 Rust)를 업데이트하려면 Rust의 공식 설치 및 버전 관리 도구 rustup을 사용합니다.

```bash
rustup update
```
