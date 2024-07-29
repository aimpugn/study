# troubleshooting

## feature `edition2024` is required

### 문제

```toml
[package]
name = "form2json"
version = "0.1.0"
edition = "2021"
```

여기서 edition을 2024로 수정후에 cargo build 하면 아래와 같은 에러가 발생한다.

```log
Caused by:
  feature `edition2024` is required

  The package requires the Cargo feature called `edition2024`, but that feature is not stabilized in this version of Cargo (1.77.1 (e52e36006 2024-03-26)).
  Consider trying a newer version of Cargo (this may require the nightly release).
  See https://doc.rust-lang.org/nightly/cargo/reference/unstable.html#edition-2024 for more information about the status of this feature.
```

### 원인

`edition2024` 기능을 사용하기 위해서는 해당 기능이 안정화(stabilized)된 Cargo 버전이 필요하다.
그러나 현재 사용 중인 Cargo 버전(1.77.1)에서는 `edition2024`가 아직 안정화되지 않았기 때문에, 이 기능을 사용할 수 없다.
Rust의 새로운 edition들은 언어와 도구들의 주요 릴리스와 함께 도입되며, 특정 기능들이 안정화될 때까지는 실험적(experimental) 또는 불안정(unstable) 상태로 유지된다.

에러 메시지는 또한 이 문제를 해결하기 위해 Cargo의 더 최신 버전을 시도해보라고 제안하고 있으며, 이는 일반적으로 nightly 릴리스를 의미한다.
Rust와 Cargo의 nightly 릴리스는 최신 기능들을 포함하고 있지만, 이들 기능은 아직 완전히 안정화되지 않았기 때문에 변동성이 있을 수 있다.

### 해결

1. **Nightly Rust 사용하기**

    `edition2024`와 같은 최신 기능을 사용하려면 Rust의 nightly 버전을 설치하고 사용해야 한다.
    Rust toolchain을 nightly로 변경하려면, 다음 명령어를 실행한다.

    ```sh
    rustup default nightly
    ```

   이후, 프로젝트 디렉토리에서 `cargo build`를 다시 실행한다.

2. **Stable Rust로 돌아가기**

    만약 nightly 기능이 필요하지 않다면, `Cargo.toml` 파일에서 `edition`을 현재 stable Rust 버전이 지원하는 최신 edition으로 되돌릴 수 있다. 예를 들어, `edition = "2021"`로 설정할 수 있다.

3. **Nightly 기능에 대한 문서 참고**

    실험적인 기능들과 그 상태에 대한 자세한 정보는 Rust 문서에서 제공된다.
    에러 메시지에서 제공된 링크([Rust Unstable Features](https://doc.rust-lang.org/nightly/cargo/reference/unstable.html#edition-2024))를 방문하여 더 많은 정보를 얻을 수 있다.
