# Assembly

- [Assembly](#assembly)
    - [`cargo-show-asm`](#cargo-show-asm)
    - [`RUSTFLAGS`](#rustflags)
    - [`objdump`](#objdump)
    - [x86 style assembly](#x86-style-assembly)
    - [기타](#기타)
        - [링크](#링크)

## `cargo-show-asm`

```shell
cargo install cargo-show-asm
```

## `RUSTFLAGS`

```shell
RUSTFLAGS="--emit asm -C llvm-args=-x86-asm-syntax=intel" cargo build
```

## `objdump`

```shell
objdump -M intel -d target/release/your_binary | grep -A20 "<virtualization::memory::address_translation::func_to_increase"
```

## x86 style assembly

```shell
# Add a target to a Rust toolchain
rustup target add x86_64-unknown-linux-gnu
```

```shell
RUSTFLAGS="--emit asm -C llvm-args=-x86-asm-syntax=intel" cargo build --target x86_64-unknown-linux-gnu
```

## 기타

### 링크

- [Rust Disassembly: part 1](https://giordi91.github.io/post/disassemlbyrust1/)
- [playground](https://play.rust-lang.org/?version=stable&mode=debug&edition=2021)
