# rust

- [rust](#rust)
    - [vscode \& rust-analyzer \& Message: Failed to spawn ... "rustfmt"](#vscode--rust-analyzer--message-failed-to-spawn--rustfmt)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [vscode \&\& rust-analyzer failed to load workspace](#vscode--rust-analyzer-failed-to-load-workspace)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)

## vscode & rust-analyzer & Message: Failed to spawn ... "rustfmt"

### 문제

```log
[Error - 오후 8:47:44] Request textDocument/formatting failed.
  Message: Failed to spawn cd "/Users/rody/VscodeProjects/study/computer_architecture/ostep/src" && "rustfmt"
  Code: -32603
```

### 원인

근본적인 원인은 fish shell에서 설정한 PATH variable이 vscode에 그대로 적용되지 않았기 때문입니다.

---

`brew install rustup` 으로 `rustup`을 설치했는데, 이때 component를 추가하면 brew 설치 경로 하위에 설치가 됩니다.

`$PATH` 경로에 `"$(brew --prefix rustup)/bin"`을 추가했고, 이 경우 추가된 경로는 다음과 같습니다.

```sh
❯ echo "$(brew --prefix rustup)/bin"
/opt/homebrew/opt/rustup/bin
```

실제로 설치된 컴포넌트들을 보면 다음과 같습니다.

```sh
❯ ll /opt/homebrew/opt/rustup/bin/rust*
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rust-analyzer@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rust-gdb@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rust-gdbgui@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rust-lldb@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rustc@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rustdoc@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rustfmt@ -> rustup-init
lrwxr-xr-x@ 1 rody  admin    11B  4 24 17:31 /opt/homebrew/opt/rustup/bin/rustup@ -> rustup-init
-rwxr-xr-x@ 1 rody  admin   6.8M  4 24 17:31 /opt/homebrew/opt/rustup/bin/rustup-init*
```

### 해결

Vscode 앱을 파인더의 응용프로그램으로 실행하는 대신, [`code` 명령어를 설치해서 `code` 명령어로 작업 공간을 엽니다](https://stackoverflow.com/a/45637716).

이러면 PATH variable들이 정상적으로 반영돼서 아래와 같이 `rustfmt`의 경로를 오버라이드 하지 않아도 정상 작동하게 됩니다.

---

vscode의 rust-analyzer는 `/opt/homebrew/opt/rustup/bin` 경로에 위치한 `rustfmt` 실행 바이너리를 인식 못하는 것으로 보여서,
아래와 같이 설정을 덮어 씁니다.

```json
    "rust-analyzer.rustfmt.overrideCommand": [
        "/opt/homebrew/opt/rustup/bin/rustfmt"
    ],
```

## vscode && rust-analyzer failed to load workspace

### 문제

```log
rust-analyzer failed to load workspace: Failed to load the project at /Users/rody/VscodeProjects/study/computer_architecture/ostep/Cargo.toml: Failed to query rust toolchain version at /Users/rody/VscodeProjects/study/computer_architecture/ostep, is your toolchain setup correctly?: cd "/Users/rody/VscodeProjects/study/computer_architecture/ostep" && "cargo" "--version" failed: No such file or directory (os error 2)
```

### 원인

프로젝트 루트 경로에 Cargo.toml 파일 등이 없어서 rust-analyzer가 작업 공간을 제대로 인식하지 못했습니다.

### 해결

작업 공간을 직접 지정합니다.

```json
    "rust-analyzer.linkedProjects": [
        "/Users/rody/VscodeProjects/study/computer_architecture/ostep/Cargo.toml"
    ],
```
