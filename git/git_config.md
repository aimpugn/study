# config

- [config](#config)
    - [구성 파일](#구성-파일)
    - [각 설정 값이 저장된 파일의 경로 확인](#각-설정-값이-저장된-파일의-경로-확인)
    - [구성 정보 확인하기](#구성-정보-확인하기)
    - [특정 구성 값 조회하기](#특정-구성-값-조회하기)
    - [특정 구성 정보 글로벌로 설정하기](#특정-구성-정보-글로벌로-설정하기)
    - [현재 적용된 모든 설정과 각 구성 파일 경로 확인하기](#현재-적용된-모든-설정과-각-구성-파일-경로-확인하기)
    - [git pull 시 rebase 하도록 설정](#git-pull-시-rebase-하도록-설정)
    - [`user` 및 `credential.helper` 설정](#user-및-credentialhelper-설정)
    - [에디터 설정](#에디터-설정)
    - [기타](#기타)
        - [설정 파일 위치](#설정-파일-위치)
        - [오류의 원인](#오류의-원인)
        - [`git config --global`의 역할](#git-config---global의-역할)
        - [`git config`로 할 수 있는 설정들](#git-config로-할-수-있는-설정들)

## 구성 파일

- 전역 Git 구성 파일:

    ```bash
    cat ~/.gitconfig
    ```

    ```ini
    # This is Git's per-user configuration file.
    [user]
    # Please adapt and uncomment the following lines:
        name = rody
        email = rody@some.domain

    [core]
        editor = code --wait
        pager = delta
        excludesFile = /Users/rody/.gitexcludes

    [interactive]
        diffFilter = delta --color-only

    [delta]
        navigate = true

    [merge]
        conflictstyle = diff3

    [diff]
        colorMoved = default

    [pull]
        rebase = true
    [http]
        postBuffer = 524288000
        lowSpeedLimit = 0
        lowSpeedTime = 999999
    [credential]
        helper = osxkeychain
    ```

- 로컬 저장소 구성 파일:

   ```bash
   cat .git/config
   ```

   ```sh
   [core]
        repositoryformatversion = 0
        filemode = true
        bare = false
        logallrefupdates = true
        ignorecase = true
        precomposeunicode = true
    [remote "origin"]
        url = https://github.com/aimpugn/study.git
        fetch = +refs/heads/*:refs/remotes/origin/*
    [branch "main"]
        remote = origin
        merge = refs/heads/main
        vscode-merge-base = origin/main
    [submodule "computer_architecture/ostep/ostep-homework"]
        url = https://github.com/remzi-arpacidusseau/ostep-homework
        active = true
   ```

## 각 설정 값이 저장된 파일의 경로 확인

각 설정 값이 저장된 파일의 경로와 함께 나열됩니다

```bash
git config --list --show-origin
```

## 구성 정보 확인하기

- [Getting Started - First-Time Git Setup](https://git-scm.com/book/en/v2/Getting-Started-First-Time-Git-Setup#Checking-Your-Settings)
- [How do I show my global Git configuration?](https://stackoverflow.com/questions/12254076/how-do-i-show-my-global-git-configuration)
- [git config explained](https://stackoverflow.com/questions/12254076/how-do-i-show-my-global-git-configuration/56506187#56506187)

```shell
git config --global --list
```

## 특정 구성 값 조회하기

```bash
git config --get core.excludesFile
```

## 특정 구성 정보 글로벌로 설정하기

```bash
git config --global core.excludesFile ~/.config/git/ignore
```

## 현재 적용된 모든 설정과 각 구성 파일 경로 확인하기

```bash
git config --list --show-origin
```

## git pull 시 rebase 하도록 설정

- [3.6 Git 브랜치 - Rebase 하기](https://git-scm.com/book/ko/v2/Git-%EB%B8%8C%EB%9E%9C%EC%B9%98-Rebase-%ED%95%98%EA%B8%B0)

```shell
# git fetch && git rebase
git config --global pull.rebase true
```

## `user` 및 `credential.helper` 설정

```bash
# Git Credential Helper를 'store' 모드로 설정해서
# 인증 정보를 디스크에 일반 텍스트 형식으로 임시적으로 저장
git config --global credential.helper store
```

```bash
# Git 커밋을 위해 사용자 이메일과 이름을 설정
# 이 정보는 커밋할 때 커밋 로그에 사용자 정보를 추가하는 데 사용
git config --global user.email "your-email@example.com"
git config --global user.name "your-username"
```

## 에디터 설정

```bash
git config --global core.editor "code --wait"
```

## 기타

### 설정 파일 위치

- Linux 및 macOS
    - `~/.gitconfig`
    - 또는 `~/.config/git/config`
- Windows:
    - `C:\Users\<username>\.gitconfig`
    - 또는 `C:\Users\<username>\.config\git\config`

### 오류의 원인

- Git은 커밋을 생성할 때 커밋 작성자의 이름과 이메일 정보를 사용합니다.
- IntelliJ 또는 기타 Git 인터페이스에서 체리픽과 같은 작업을 할 때, 이러한 정보가 설정되어 있지 않으면 Git은 사용자의 신원을 확인할 수 없으며 오류가 발생합니다.

### `git config --global`의 역할

- `git config --global` 명령어는 Git 사용자의 전역 설정을 정의합니다. 이 설정은 시스템의 모든 Git 저장소에 적용됩니다.
- `git config --global`로 설정된 정보는 사용자의 홈 디렉토리에 위치한 `.gitconfig` 또는 `~/.config/git/config` 파일에 저장됩니다.

### `git config`로 할 수 있는 설정들

1. 사용자 정보 설정:
   - 이름 설정: `git config --global user.name "Your Name"`
   - 이메일 설정: `git config --global user.email "your_email@example.com"`

2. 기본 편집기 설정:
   - 예: `git config --global core.editor "code --wait"`

3. 기본 브랜치 이름 설정:
   - 예: `git config --global init.defaultBranch main`

4. 커밋 메시지 템플릿 설정:
   - 예: `git config --global commit.template ~/.gitmessage.txt`

5. 색상 출력 설정:
   - 예: `git config --global color.ui auto`

6. 별칭(alias) 설정:
   - 예: `git config --global alias.st status`

7. 자동 줄바꿈 설정:
   - 예: `git config --global core.autocrlf true` (Windows)
   - 예: `git config --global core.autocrlf input` (Linux 및 macOS)

8. 원격 저장소의 기본 이름 설정:
   - 예: `git config --global remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"`

이러한 설정들은 Git 작업을 보다 편리하게 만들고, 사용자의 환경에 맞게 Git을 사용자 정의하는 데 도움을 줍니다. 개별 저장소의 설정을 변경하려면 `--global` 플래그를 생략하고 해당 저장소 내에서 명령을 실행합니다.
