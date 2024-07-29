# Troubleshooting

## Blocking waiting for file lock on package cache

### 문제

```shell
❯ cargo install
    Blocking waiting for file lock on package cache

```

### 원인

캐시 문제인 것으로 보이며, [검색해 보니 그냥 클린](https://stackoverflow.com/a/49634470/8562273)하면 되는 것 같다

### 해결

```shell
❯ cargo clean
❯ cargo install
error: Using `cargo install` to install the binaries for the package in current working directory is no longer supported, use `cargo install --path .` instead. Use `cargo build` if you want to simply build the package.

❯ cargo install --path .
  Installing ostep v0.1.0 (/Users/rody/VscodeProjects/ostep)
    Updating crates.io index
  Downloaded libc v0.2.140
  Downloaded fork v0.1.21
  Downloaded 2 crates (674.0 KB) in 1.71s
   Compiling libc v0.2.140
   Compiling cfg-if v1.0.0
   Compiling autocfg v1.1.0
   Compiling ppv-lite86 v0.2.17
   Compiling pin-utils v0.1.0
   Compiling bitflags v1.3.2
   Compiling static_assertions v1.1.0
   Compiling memoffset v0.7.1
   Compiling getrandom v0.2.8
   Compiling nix v0.26.2
   Compiling fork v0.1.21
   Compiling rand_core v0.6.4
   Compiling rand_chacha v0.3.1
   Compiling rand v0.8.5
   Compiling ostep v0.1.0 (/Users/rody/VscodeProjects/ostep)
    Finished release [optimized + debuginfo] target(s) in 6.01s
  Installing /Users/rody/.cargo/bin/ostep
   Installed package `ostep v0.1.0 (/Users/rody/VscodeProjects/ostep)` (executable `ostep`)
```

## `let...else` statements are unstable

### 문제

```log
   Compiling cargo-show-asm v0.2.17
error[E0658]: `let...else` statements are unstable
   --> /Users/rody/.cargo/registry/src/github.com-1ecc6299db9ec823/cargo-show-asm-0.2.17/src/main.rs:348:17
    |
348 |                 let Some(name) = maybe_origin.file_name() else { continue };
    |                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    |
    = note: see issue #87335 <https://github.com/rust-lang/rust/issues/87335> for more information

error[E0658]: `let...else` statements are unstable
   --> /Users/rody/.cargo/registry/src/github.com-1ecc6299db9ec823/cargo-show-asm-0.2.17/src/main.rs:349:17
    |
349 |                 let Some(name) = name.to_str() else { continue };
    |                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    |
    = note: see issue #87335 <https://github.com/rust-lang/rust/issues/87335> for more information

For more information about this error, try `rustc --explain E0658`.
error: could not compile `cargo-show-asm` due to 2 previous errors
error: failed to compile `cargo-show-asm v0.2.17`, intermediate artifacts can be found at `/var/folders/9x/8djp1ylj221bk02dp8zqsps00000gn/T/cargo-installsXl3w4`
```

### 원인

설치 당시 러스트 버전이 `1.64.0`인데, 이 버전으로는  let...else가 unstable

```shell
❯ rustc --version
rustc 1.64.0 (a55dd71d5 2022-09-19)
```

### 해결

버전 의존적인 내용이 없어서, 버전 업

```shell
rustup update
```

```shell
❯ rustc --version
rustc 1.68.2 (9eb3afe9e 2023-03-27)
```
