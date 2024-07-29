# shell

- [shell](#shell)
    - [셸(Shell)이란?](#셸shell이란)
    - [셸의 내부 동작](#셸의-내부-동작)
    - [셸의 종류](#셸의-종류)
    - [현재 사용중인 shell 확인하는 방법](#현재-사용중인-shell-확인하는-방법)
    - [zsh -\> fish shell](#zsh---fish-shell)
    - [sub-shell](#sub-shell)
    - [터미널 에뮬레이터, TTY, PTY와 Shell?](#터미널-에뮬레이터-tty-pty와-shell)

## 셸(Shell)이란?

셸은 운영체제에서 *사용자와 커널 사이의 인터페이스 역할을 하는 프로그램*입니다.
사용자가 명령어를 입력하면, 셸이 이를 해석하고 커널에게 전달하여 명령을 실행합니다.
셸은 명령 줄 인터페이스(CLI)를 제공하며, 배치 작업(batch jobs)을 처리하거나 스크립트를 실행하는 등 다양한 작업을 수행할 수 있습니다.

## 셸의 내부 동작

셸의 내부 동작은 다음과 같은 단계로 이루어집니다:

1. 사용자가 명령어를 입력하고 엔터 키를 누르면, 셸은 이 명령어를 읽습니다.
2. 셸은 입력된 명령어를 개별적인 토큰으로 분리합니다.

    예를 들어, `ls -l /home` 명령어는 `ls`, `-l`, `/home`이라는 세 개의 토큰으로 분리됩니다.

3. 토큰을 분석하여 명령어의 구조를 이해하고, 명령어의 종류, 옵션, 인수 등을 파악합니다.
4. 셸은 명령어를 실행하기 위해 시스템 호출을 통해 커널에게 요청합니다.

    이때, 셸은 적절한 실행 파일을 찾아 실행하고, 필요한 인수와 환경 변수를 설정합니다.

5. 실행된 명령어의 결과를 터미널에 출력하고, 명령어의 종료 상태(exit status)를 반환합니다.

## 셸의 종류

- **Bash (Bourne Again Shell)**:

    유닉스 계열 시스템에서 널리 사용되는 셸로, GNU 프로젝트의 일환으로 개발되었습니다.
    POSIX 표준을 따르며, 많은 스크립트와 호환성이 높습니다.

- **Zsh (Z Shell)**:

    Bash와 호환되면서도 더 많은 기능과 유연성을 제공합니다.
    강력한 자동 완성 기능과 다양한 플러그인을 지원합니다.

- **Ksh (Korn Shell)**:

    AT&T 벨 연구소에서 개발된 셸로, 상용 유닉스 시스템에서 많이 사용되었습니다.
    스크립트 작성에 강점을 가지고 있습니다.

- **Tcsh**:

    C 셸의 확장판으로, C 언어 스타일의 구문을 지원하며, 인터랙티브 사용에 강점을 가지고 있습니다.

## 현재 사용중인 shell 확인하는 방법

현재 사용 중인 셸을 확인하는 방법은 여러 가지가 있습니다. 가장 일반적이고 신뢰할 수 있는 방법들을 소개해 드리겠습니다:

1. `echo $SHELL` 명령어 사용:

    `echo $SHELL`은 로그인 셸을 보여주지만, 현재 사용 중인 셸과 다를 수 있음에 주의합니다.

    ```sh
    # 현재 로그인 셸의 경로를 보여줍니다.
    echo $SHELL
    ```

    ```sh
    rody@rodyui-MacBookPro ~> echo $SHELL
    /bin/zsh
    ```

2. `ps` 명령어 사용:

    현재 실행 중인 셸 프로세스를 확인할 수 있습니다.

    ```sh
    ps -p $$
    ```

    ```sh
    rody@rodyui-MacBookPro ~ % ps -p $$
    PID TTY           TIME CMD
    24272 ttys000    0:00.05 zsh
    ```

    또는

    ```bash
    ps -p $$ -o comm=
    ```

    ```sh
    rody@rodyui-MacBookPro ~ % ps -p $$ -o comm=
    zsh
    ```

3. `$0` 변수 확인:

    많은 셸에서 `$0`는 현재 실행 중인 셸의 이름을 나타냅니다.

    ```sh
    echo $0
    ```

    ```sh
    rody@rodyui-MacBookPro ~ % echo $0
    zsh
    ```

4. `/proc/self/exe` 확인 (Linux 시스템에서):

    ```bash
    readlink /proc/self/exe
    ```

5. 셸 특정 변수 확인:
   - Bash: `echo $BASH_VERSION`
   - Zsh: `echo $ZSH_VERSION`
   - Fish: `echo $FISH_VERSION`

6. 명령어 실행:

    ```bash
    echo $SHELL && $SHELL --version
    ```

   이 명령은 셸의 경로와 버전 정보를 함께 보여줍니다.

## zsh -> fish shell

zsh를 삭제하고 fish shell로 대체하는 과정:

1. fish shell 설치:

    Ubuntu/Debian:

    ```sh
    sudo apt-get update
    sudo apt-get install fish
    ```

    macOS (Homebrew 사용):

    ```sh
    brew install fish
    ```

2. fish shell 테스트:

    ```sh
    # fish --version
    rody@rodyui-MacBookPro ~ % fish --version
    fish, version 3.7.1
    ```

3. 기본 셸을 fish로 변경:

    ```sh
    # `chpass`, `chfn`, `chsh`: add or change user database information
    # `-s` newshell: Attempt to change the user's shell to newshell.
    chsh -s $(which fish)
    ```

    ```sh
    rody@rodyui-MacBookPro ~ % chsh -s $(which fish)
    Changing shell for rody.
    Password for rody:
    chsh: /opt/homebrew/bin/fish: non-standard shell
    ```

4. zsh 삭제 (선택사항):

   일부 시스템 도구가 zsh에 의존할 수 있기 때문에,
   시스템에 따라 zsh를 완전히 제거하는 것은 권장되지 않을 수 있습니다.

   Ubuntu/Debian:

   ```sh
   sudo apt-get remove zsh
   ```

   macOS:
   Zsh는 macOS의 기본 셸이므로 제거하지 않는 것이 좋습니다.

5. Oh My Zsh 제거 (설치되어 있는 경우):

   ```sh
   uninstall_oh_my_zsh
   ```

6. zsh 관련 설정 파일 삭제:

   ```sh
   rm -rf ~/.zshrc ~/.oh-my-zsh
   ```

7. 시스템 재시작 또는 로그아웃/로그인:
   변경 사항을 적용하기 위해 시스템을 재시작하거나 로그아웃 후 다시 로그인합니다.

8. fish 설정 (선택사항):
   fish의 웹 기반 설정 인터페이스를 사용하여 기본 설정을 구성할 수 있습니다.

   ```sh
   fish_config
   ```

주의사항:
- 시스템 기본 셸을 변경하기 전에 fish가 `/etc/shells`에 등록되어 있는지 확인하세요. 등록되어 있지 않다면 수동으로 추가해야 할 수 있습니다.
- 일부 스크립트나 애플리케이션이 bash나 zsh 문법에 의존할 수 있으므로, 이들의 호환성을 확인해야 합니다.
- 시스템 관리 작업을 위해 bash나 다른 셸에 대한 기본적인 지식을 유지하는 것이 좋습니다.

fish shell로 전환한 후에는 새로운 문법과 기능에 익숙해지는 시간이 필요할 수 있습니다. fish는 사용자 친화적인 기능이 많지만, 기존의 bash나 zsh 스크립트와 완전히 호환되지 않을 수 있으므로 주의가 필요합니다.

## sub-shell

서브셸에서 변경된 변수는 부모 셸에 영향을 주지 않습니다.

```bash
var=0
(var=999; echo "In sub-shell: $var")  # In sub-shell: 999
echo "In parent shell: $var" # In parent shell: 0
```

## 터미널 에뮬레이터, TTY, PTY와 Shell?

1. **터미널 에뮬레이터(Terminal Emulator)**:

    문자 기반의 사용자 인터페이스를 제공하는 프로그램입니다.
    실제 물리적인 터미널이거나 소프트웨어 터미널 에뮬레이터일 수 있습니다.
    그래픽 사용자 인터페이스(GUI) 환경에서 셸을 실행할 수 있는 창을 제공하는 프로그램입니다.

    GNOME Terminal, Konsole, iTerm2 등이 터미널 에뮬레이터에 해당합니다.

    터미널 에뮬레이터는 셸을 실행할 수 있는 환경을 제공할 뿐, 자체적으로 명령어를 해석하거나 실행하지 않습니다.
    반면 Shell은 사용자가 명령을 입력하고 시스템이 이를 해석하여 실행하는 프로그램입니다.
    즉, 셸은 터미널 내에서 실행됩니다.

2. **TTY (Teletypewriter)**:

    TTY는 전통적으로 텍스트 기반 입력과 출력을 위한 장치입니다.
    초기 컴퓨터 시스템에서는 실제 하드웨어 장치였으며, 사용자는 키보드를 통해 입력하고 출력은 프린터나 모니터를 통해 볼 수 있었습니다.

    현대 시스템에서는 물리적인 하드웨어 대신 소프트웨어적으로 구현된 가상 터미널을 의미합니다.
    사용자는 터미널 에뮬레이터나 콘솔을 통해 TTY와 상호작용합니다.

    TTY는 하드웨어 터미널 또는 소프트웨어 터미널 인터페이스를 의미하며, 셸은 그 안에서 실행되는 프로그램입니다.
    - 사용자의 입력을 받아 셸에 전달하는 역할
    - 셸의 출력을 사용자에게 보여주는 역할

    - **Linux 콘솔**:
        - 사용자가 Ctrl+Alt+F1을 눌러 1번 콘솔로 이동하면, 해당 콘솔은 TTY 장치를 통해 직접 연결됩니다.
    - **agetty**:
        - 시스템이 부팅될 때, `agetty`가 실행되어 사용자가 로그인할 수 있는 TTY 세션을 설정합니다.

3. **Pseudo TTY (PTY)**:

    PTY는 실제 TTY 장치를 에뮬레이트하는 소프트웨어 인터페이스입니다.
    네트워크를 통한 원격 접속을 가능하게 하거나, 터미널 에뮬레이터 프로그램에서 사용됩니다.
    - 네트워크를 통해 터미널 세션을 제공
    - 여러 터미널 세션을 관리할 수 있게 하기 등

    PTY는 두 부분으로 나뉩니다:
    - 마스터: 터미널 에뮬레이터(또는 네트워크 프로그램)와 연결됩니다.
    - 슬레이브: 슬레이브는 실제 TTY처럼 동작하여 셸 또는 다른 프로그램과 상호작용합니다.

    PTY는 소프트웨어적으로 TTY 기능을 제공하는 반면, 셸은 이 PTY 위에서 실행됩니다.
    사용자가 PTY 마스터를 통해 입력을 보내면, 이 입력은 PTY 슬레이브로 전달되어 셸에서 처리됩니다.
    셸의 출력은 다시 PTY 슬레이브를 통해 마스터로 전달되어 터미널 에뮬레이터에 표시됩니다.

    - **SSH**:
        - 사용자가 SSH 클라이언트를 통해 원격 서버에 접속하면, PTY 마스터는 SSH 클라이언트와 연결되고, PTY 슬레이브는 원격 서버의 셸과 연결됩니다.
    - **tmux/screen**:
        - `tmux`나 `screen`과 같은 터미널 멀티플렉서 프로그램은 PTY를 사용하여 여러 터미널 세션을 관리하고, 사용자가 한 세션에서 다른 세션으로 쉽게 전환할 수 있게 합니다.
    - **터미널 에뮬레이터**:
        - GNOME Terminal이나 iTerm2와 같은 터미널 에뮬레이터는 PTY를 사용하여 셸과 상호작용합니다.
