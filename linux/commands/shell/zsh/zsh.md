# zsh

- [zsh](#zsh)
    - [zsh?](#zsh-1)
    - [설치 및 구성](#설치-및-구성)
        - [Plugins](#plugins)
    - [Zsh 초기화 파일 로딩 순서](#zsh-초기화-파일-로딩-순서)
        - [`.zshrc`과 `.zprofile`](#zshrc과-zprofile)
        - [`zshenv`](#zshenv)
        - [`zprofile`](#zprofile)
        - [`zshrc`](#zshrc)
        - [`zlogin`](#zlogin)
        - [`zlogout`](#zlogout)
    - [References](#references)

## zsh?

Zsh는 셸(Shell)로, *터미널에서 사용자 명령을 해석하고 실행하는 프로그램*입니다.

## 설치 및 구성

```bash
# Install ZSH
sudo apt install zsh-autosuggestions zsh-syntax-highlighting zsh
```

```bash
# Install Oh my ZSH
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
```

### [Plugins](https://github.com/ohmyzsh/ohmyzsh/wiki/Plugins)

```bash
# Install plugins
# autosuggesions
git clone https://github.com/zsh-users/zsh-autosuggestions.git $ZSH_CUSTOM/plugins/zsh-autosuggestions

# zsh-syntax-highlighting
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git $ZSH_CUSTOM/plugins/zsh-syntax-highlighting

# zsh-fast-syntax-highlighting
git clone https://github.com/zdharma-continuum/fast-syntax-highlighting.git ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/fast-syntax-highlighting

# zsh-autocomplete
git clone --depth 1 -- https://github.com/marlonrichert/zsh-autocomplete.git $ZSH_CUSTOM/plugins/zsh-autocomplete
```

```bash
# # Enable plugins by adding them to .zshrc
plugins=(
    git
    zsh-autosuggestions
    zsh-syntax-highlighting
    fast-syntax-highlighting
    zsh-autocomplete
)
```

## [Zsh 초기화 파일 로딩 순서](https://shreevatsa.wordpress.com/2008/03/30/zshbash-startup-files-loading-order-bashrc-zshrc-etc/)

```ascii
+----------------+-----------+-----------+------+
|                |Interactive|Interactive|Script|
|                |login      |non-login  |      |
+----------------+-----------+-----------+------+
|/etc/zshenv     |    A      |    A      |  A   |
+----------------+-----------+-----------+------+
|~/.zshenv       |    B      |    B      |  B   |
+----------------+-----------+-----------+------+
|/etc/zprofile   |    C      |           |      |
+----------------+-----------+-----------+------+
|~/.zprofile     |    D      |           |      |
+----------------+-----------+-----------+------+
|/etc/zshrc      |    E      |    C      |      |
+----------------+-----------+-----------+------+
|~/.zshrc        |    F      |    D      |      |
+----------------+-----------+-----------+------+
|/etc/zlogin     |    G      |           |      |
+----------------+-----------+-----------+------+
|~/.zlogin       |    H      |           |      |
+----------------+-----------+-----------+------+
|~/.zlogout      |    I      |           |      |
+----------------+-----------+-----------+------+
|/etc/zlogout    |    J      |           |      |
+----------------+-----------+-----------+------+
```

1. 실행 순서: A → B → C → D → E → F → G → H (로그인 시)
2. 비로그인 대화형 셸: A → B → C → D 순으로 실행
3. 스크립트 실행 시: A → B만 실행
4. 시스템 전체(`/etc/*`)와 개별 사용자(`~/.*`)의 설정을 분리되며, 시스템 전체 파일이 사용자별 파일보다 먼저 실행됩니다.
5. `zshenv`는 모든 상황에서 실행되므로, 여기에는 필수적인 환경 설정만 포함해야 합니다.
6. `zshrc`는 대부분의 사용자 정의 설정이 들어가는 곳입니다.

### `.zshrc`과 `.zprofile`

- `~/.zprofile`:

    - 로그인 쉘에서만 실행됩니다.
    - 주로 환경 변수 설정, PATH 수정 등 전체 세션에 영향을 미치는 설정에 사용됩니다.
    - 시스템에 로그인할 때마다 한 번만 실행됩니다.

    ```sh
    export PATH=$PATH:/새로운/경로
    ```

- `~/.zshrc`:

    - 모든 대화형 zsh 세션에서 실행됩니다(로그인 및 비로그인 쉘).
    - 쉘 옵션, 별칭, 함수 정의, 키 바인딩 등 대화형 사용을 위한 설정에 사용됩니다.
    - 새 zsh 터미널을 열 때마다 실행됩니다.
    - `~/.zshrc`는 새 터미널마다 실행되므로, 무거운 작업은 가능한 `~/.zprofile`에 배치하는 것이 좋습니다.
    - 예: 별칭 정의, 함수 선언, 프롬프트 설정 등

로그인 쉘의 경우 `~/.zprofile` (D) -> `~/.zshrc` (F) 순으로 로드됩니다.
비로그인 대화형 쉘의 경우 `~/.zshrc` (D)만 로드됩니다.

### `zshenv`

모든 Zsh 세션에 적용되는 환경 변수를 설정합니다.
스크립트 실행을 포함한 모든 Zsh 인스턴스에서 실행됩니다.
그래서 모든 Zsh 셸 세션에서 공통으로 적용해야 하는 환경 변수나 설정을 지정하는 데 사용됩니다.

예를 들어, `PATH`, `EDITOR` 등 어떤 터미널에서든 동일한 환경 변수를 유지하고 싶을 때 사용합니다.

```zsh
export PATH=$PATH:/usr/local/bin
export EDITOR=vim
```

- `/etc/zshenv` (A)

    - 모든 Zsh 세션에서 항상 가장 먼저 실행됩니다.
    - 시스템 전체에 적용되는 환경 변수를 설정하는 데 사용됩니다.
    - 로그인/비로그인, 대화형/비대화형 모든 경우에 실행됩니다.

- `~/.zshenv` (B)

    - 사용자별 환경 변수 설정에 사용됩니다.
    - `/etc/zshenv` 다음에 실행되며, 역시 로그인/비로그인 가리지 않고 모든 Zsh shell이 시작될 때마다 실행됩니다.
    - `PATH`, `LANG`, `TZ`와 같은 모든 세션에서 필요한 전역 환경 변수 설정이 주로 포함됩니다.

### `zprofile`

`~/.zprofile` 파일은 *로그인 셸*에서만 실행되며, 로그인 셸에서만 필요한 설정이나 초기화 작업을 수행할 때 사용됩니다.
로그인 셸은 사용자가 처음 터미널을 열 때나 SSH 등으로 원격 로그인할 때 시작됩니다.
로그인할 때 한 번만 적용되며, 새로운 서브 셸에서는 다시 실행되지 않습니다.

예를 들어 특정 프로그램의 초기화 스크립트나, `PATH`에 특정 디렉터리를 추가하는 작업 등이 포함됩니다.

```zsh
# 시스템 메시지 표시
echo "Welcome to the system!"
```

- `/etc/zprofile` (C)

    - 로그인 셸에서만 실행됩니다.
    - 시스템 전체에 적용되는 로그인 관련 설정을 포함합니다.

- `~/.zprofile` (D)

    - 사용자별 로그인 셸 설정에 사용됩니다.
    - 로그인 시에만 한 번 실행되어야 하는 명령어를 포함합니다.

### `zshrc`

대화형 셸 세션을 위한 설정 파일입니다.
별칭, 함수, 키 바인딩, 셸 옵션 등 포함합니다.

```zsh
alias ll='ls -la'
bindkey -v  # Vi 모드 활성화
```

- `/etc/zshrc` (E for login, C for non-login)

    - 모든 대화형 Zsh 세션에서 실행됩니다.
    - 시스템 전체에 적용되는 대화형 셸 설정을 포함합니다.

- `~/.zshrc` (F for login, D for non-login)

    - 사용자별 대화형 셸 설정에 사용됩니다.
    - 별칭, 함수, 키 바인딩, 셸 옵션 등 대화형 사용을 위한 설정을 포함합니다.

### `zlogin`

로그인 셸 초기화가 완료된 후 실행되는 명령어입니다.
zshrc 이후에 실행되어, 이미 설정된 환경을 기반으로 작업 가능합니다.
로그인 시 실행할 프로그램이나 메시지에 적합합니다.

```zsh
# 로그인 시간 기록
echo "Logged in at $(date)" >> ~/.login_history
```

- `/etc/zlogin` (G)

    - 로그인 셸에서만 실행됩니다.
    - 시스템 전체에 적용되는 로그인 후 설정을 포함합니다.

- `~/.zlogin` (H)

    - 사용자별 로그인 후 설정에 사용됩니다.
    - `~/.zprofile`과 유사하지만, 다른 시작 파일들이 모두 실행된 후에 실행됩니다.

### `zlogout`

로그인 셸 종료 시 실행되는 명령어로, 사용자가 로그아웃할 때 실행됩니다.
세션 정리, 백업, 로그 기록 등에 사용합니다.

```zsh
# 임시 파일 정리
rm -rf /tmp/my_temp_files
```

- `~/.zlogout` (I)

    - 로그인 셸이 종료될 때 실행됩니다.
    - 세션 정리, 백업 등의 작업에 사용될 수 있습니다.

- `/etc/zlogout` (J)

    - 시스템 전체에 적용되는 로그아웃 스크립트입니다.
    - 모든 사용자의 로그인 셸이 종료될 때 실행됩니다.

## References

- [Oh my ZSH](https://github.com/ohmyzsh/ohmyzsh)
- [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions)
- [zsh-syntax-highlighting](https://github.com/zsh-users/zsh-syntax-highlighting)
- [zsh-fast-syntax-highlighting](https://github.com/zdharma/fast-syntax-highlighting)
- [zsh-autocomplete](https://github.com/marlonrichert/zsh-autocomplete)
