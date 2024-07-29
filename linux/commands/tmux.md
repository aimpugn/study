# tmux

- [tmux](#tmux)
    - [세션 생성](#세션-생성)
    - [세션 확인](#세션-확인)
    - [세션 삭제](#세션-삭제)
    - [화면 분할](#화면-분할)
        - [수평 분할](#수평-분할)
        - [수직 분할](#수직-분할)
        - [패널 간 이동](#패널-간-이동)
        - [N개 패널을 균등 분할](#n개-패널을-균등-분할)
            - [`%` 명령어 사용](#-명령어-사용)
            - [`split-window -h` 또는 `split-window -v` 명령어 사용](#split-window--h-또는-split-window--v-명령어-사용)
    - [패널 종료](#패널-종료)
    - [모든 패널에 동시 입력](#모든-패널에-동시-입력)
    - [화면 3분할 함수](#화면-3분할-함수)
    - [conf](#conf)
        - [`~/.tmux.conf` 적용하기](#tmuxconf-적용하기)
        - [bind](#bind)
            - [mac에서 `Cmd` 또는 `option`으로 키 바인드](#mac에서-cmd-또는-option으로-키-바인드)

## 세션 생성

```shell
# new-session [-AdDEPX] [-c start-directory] 
#             [-e environment] [-f flags] [-F format] 
#             [-n window-name] [-s session-name] [-t group-name] 
#             [-x width] [-y height] 
#             [shell-command] 
#             (alias: new)
# `-d`: the initial size comes from the global default-size option
tmux new-session -d -s "$sessionName"
```

## 세션 확인

```shell
# has-session [-t target-session] (alias: has)
tmux has-session -t "$sessionName" 2>/dev/null
```

## 세션 삭제

```shell
# kill-session [-aC] [-t target-session]
tmux kill-session -t "$sessionName" 2>/dev/null
```

## 화면 분할

### 수평 분할

```shell
Ctrl + b, "
```

### 수직 분할

```shell
Ctrl + b, %
```

### 패널 간 이동

```shell
Ctrl + b, <Arrow keys>
```

또는

```shell
Ctrl + b, q 입력 후 pane 선택
```

### N개 패널을 균등 분할

#### `%` 명령어 사용

1. 먼저, `tmux` 세션을 시작
2. 현재 창을 수평으로 균등하게 분할하려면 Prefix + % 키를 누른다 (Prefix는 기본적으로 Ctrl + b입니다. 이 문서에서는 Prefix 키가 Ctrl + b로 설정되어 있다고 가정)
3. 이제 어떤 패널이든 선택하고, 그 패널을 수직으로 균등하게 분할하려면 `Prefix + "` 키를 누른다

#### `split-window -h` 또는 `split-window -v` 명령어 사용

> - `-v`: vertical
> - `-h`: horizontal

1. `tmux` 세션을 시작합니다.
2. `Prefix + :`를 눌러 명령 프롬프트를 엽니다.
3. 다음 명령을 입력하여 현재 창을 수평으로 균등하게 분할합니다:

    ```shell
    split-window -h
    ```

4. 이제 어떤 패널이든 선택하고, 다음 명령을 입력하여 선택한 패널을 수직으로 균등하게 분할합니다:

    ```shell
    split-window -v
    ```

5. 위의 명령은 수동으로 입력하고 실행해야 하므로, 이를 자동화하기 위해 tmux 설정 파일(~/.tmux.conf)에 다음과 같은 스크립트를 추가할 수 있습니다:

```shell
bind e \
  split-window -h \; \
  select-pane -t :.+ \; \
  split-window -v
```

이 스크립트를 ~/.tmux.conf 파일에 추가하고 tmux 세션을 재시작한 후, Prefix + e를 눌러 세 패널을 균등하게 분할할 수 있습니다.

## 패널 종료

`Ctrl + d` 또는 `exit` 명령어 입력

## 모든 패널에 동시 입력

- "synchronized panes" 기능 사용
- 활성화
    1. tmux 세션 열기
    2. `Ctrl + b`를 누른 후, `:`로 tmux 명령 프롬프트 열기
    3. `setw synchronize-panes on` 입력
- 비활성화
    1. `Ctrl + b`를 누른 후, `:`를 눌러 tmux 명령 프롬프트
    2. `setw synchronize-panes off` 입력

## 화면 3분할 함수

```shell
tmux_3pane() {
  local sessionName="tmux_3_split_panes"
  tmux kill-session -t "$sessionName" 2>/dev/null
  tmux new-session -d -s "$sessionName"
  tmux split-window -h -p 66
  tmux split-window -h -p 50
  tmux attach -t "$sessionName"
}
```

## conf

### `~/.tmux.conf` 적용하기

```shell
tmux source-file ~/.tmux.conf
```

### bind

#### mac에서 `Cmd` 또는 `option`으로 키 바인드

- [What is the tmux prefix code for command key on mac?](https://unix.stackexchange.com/questions/453466/what-is-the-tmux-prefix-code-for-command-key-on-mac)
- [bind tmux prefix to OS X cmd key (or any other binding)](https://superuser.com/questions/259614/bind-tmux-prefix-to-os-x-cmd-key-or-any-other-binding)
