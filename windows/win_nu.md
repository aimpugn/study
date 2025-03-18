# windows nu

- [windows nu](#windows-nu)
    - [시작하기](#시작하기)
        - [설치하기](#설치하기)
        - [실행하기](#실행하기)
        - [Nushell 기본 설정 (환경변수 PATH 포함)](#nushell-기본-설정-환경변수-path-포함)
        - [`config.nu` 설정](#confignu-설정)
            - [커스텀 프롬프트 설정](#커스텀-프롬프트-설정)
        - [`env.nu` 설정](#envnu-설정)
            - [`$env.PATH` 환경변수 설정](#envpath-환경변수-설정)
            - [Nushell에서 Oh My Zsh 스타일의 단축 Git Alias 설정하기](#nushell에서-oh-my-zsh-스타일의-단축-git-alias-설정하기)
    - [Nushell 유용한 기본 명령 및 활용법](#nushell-유용한-기본-명령-및-활용법)
    - [Coming from Bash](#coming-from-bash)
    - [Oh-My-Posh 설정](#oh-my-posh-설정)

## 시작하기

### 설치하기

현재 `winget`, `scoop` 등을 사용할 수 없는 환경이라, [nushell/releases](https://github.com/nushell/nushell/releases)에서 적절한 버전을 다운 받습니다.

그리고 `C:\Program Files\Nushell\`같은 폴더에 `nu.exe`를 이동시킵니다.

### 실행하기

`nu.exe`로 실행하여 터미널을 실행합니다.
그리고 정상적으로 동작하는지 확인합니ㅏㄷ.

```sh
~> version
╭────────────────────┬──────────────────────────────────────────╮
│ version            │ 0.102.0                                  │
│ major              │ 0                                        │
│ minor              │ 102                                      │
│ patch              │ 0                                        │
│ branch             │                                          │
│ commit_hash        │ 1aa2ed1947a0b891398558fcf4e4289849cc5a1d │
│ build_os           │ windows-x86_64                           │
│ build_target       │ x86_64-pc-windows-msvc                   │
│ rust_version       │ rustc 1.82.0 (f6e511eec 2024-10-15)      │
│ rust_channel       │ 1.82.0-x86_64-pc-windows-msvc            │
│ cargo_version      │ cargo 1.82.0 (8f40fc59f 2024-08-21)      │
│ build_time         │ 2025-02-05 01:04:24 +00:00               │
│ build_rust_channel │ release                                  │
│ allocator          │ mimalloc                                 │
│ features           │ default, sqlite, trash                   │
│ installed_plugins  │                                          │
╰────────────────────┴──────────────────────────────────────────╯
```

### Nushell 기본 설정 (환경변수 PATH 포함)

Nushell 설정 파일의 기본 위치는 `$nu.config-path`로 확인 가능합니다.

```sh
~> $nu.config-path
C:\Users\me\AppData\Roaming\nushell\config.nu
```

Nushell은 Windows의 환경변수 PATH를 자동으로 읽어오지만, 직접 설정도 가능합니다.

```nu
# config.nu 설정 예시 (기존 설정에 추가)
$env.PATH = ($env.PATH | append 'C:\Program Files\Git\cmd')
```

### `config.nu` 설정

#### 커스텀 프롬프트 설정

```nu
# config.nu

# Oh My Posh 프롬프트 설정
let-env PROMPT_COMMAND = {||
    oh-my-posh print primary --config "C:\Users\me\AppData\Local\Programs\oh-my-posh\themes\paradox.omp.json"
}
```

### `env.nu` 설정

#### `$env.PATH` 환경변수 설정

```nu
# env.nu
$env.PATH = (
    $env.PATH
    | append 'C:\Program Files\Git\cmd'
    | append 'C:\Program Files\oh-my-posh\bin'
)
```

#### Nushell에서 Oh My Zsh 스타일의 단축 Git Alias 설정하기

Nushell의 Alias는 `config.nu` 또는 `env.nu`에서 정의할 수 있습니다.

```sh
~> $nu.config-path
C:\Users\me\AppData\Roaming\nushell\config.nu
~> $nu.env-path
C:\Users\me\AppData\Roaming\nushell\env.nu
```

아래와 같이 Alias를 추가할 수 있습니다.
우선 bash 문법 사용하는 케이스들은 지워서, `env.nu`에 추가하고 `nu` 쉘 진입 시 문제 없다고 확인됩니다.

```nu
# env.nu 파일
# Oh My Zsh 스타일 Git alias
alias g = git
alias ga = git add
alias gaa = git add --all
alias gam = git am
alias gama = git am --abort
alias gamc = git am --continue
alias gams = git am --skip
alias gamscp = git am --show-current-patch
alias gap = git apply
alias gapa = git add --patch
alias gapt = git apply --3way
alias gau = git add --update
alias gav = git add --verbose
alias gb = git branch
alias gbD = git branch --delete --force
alias gba = git branch --all
alias gbd = git branch --delete
alias gbl = git blame -w
alias gbm = git branch --move
alias gbnm = git branch --no-merged
alias gbr = git branch --remote
alias gbs = git bisect
alias gbsb = git bisect bad
alias gbsg = git bisect good
alias gbsn = git bisect new
alias gbso = git bisect old
alias gbsr = git bisect reset
alias gbss = git bisect start
alias gc = git commit --verbose
alias gcB = git checkout -B
alias gca = git commit --verbose --all
alias gcam = git commit --all --message
alias gcas = git commit --all --signoff
alias gcasm = git commit --all --signoff --message
alias gcb = git checkout -b
alias gcf = git config --list
alias gcl = git clone --recurse-submodules
alias gclean = git clean --interactive -d
alias gclf = git clone --recursive --shallow-submodules --filter=blob:none --also-filter-submodules
alias gcmsg = git commit --message
alias gcn = git commit --verbose --no-edit
alias gco = git checkout
alias gcor = git checkout --recurse-submodules
alias gcount = git shortlog --summary --numbered
alias gcp = git cherry-pick
alias gcpa = git cherry-pick --abort
alias gcpc = git cherry-pick --continue
alias gcs = git commit --gpg-sign
alias gcsm = git commit --signoff --message
alias gcss = git commit --gpg-sign --signoff
alias gcssm = git commit --gpg-sign --signoff --message
alias gd = git diff
alias gdca = git diff --cached
alias gdcw = git diff --cached --word-diff
alias gds = git diff --staged
alias gdt = git diff-tree --no-commit-id --name-only -r
alias gdup = git diff @{upstream}
alias gdw = git diff --word-diff
alias gf = git fetch
alias gfa = git fetch --all --tags --prune --jobs=10
alias gfo = git fetch origin
alias gg = git gui citool
alias gga = git gui citool --amend
alias ghh = git help
alias gignore = git update-index --assume-unchanged
alias gl = git pull
alias glg = git log --stat
alias glgg = git log --graph
alias glgga = git log --graph --decorate --all
alias glgm = git log --graph --max-count=10
alias glgp = git log --stat --patch
alias glo = git log --oneline --decorate
alias glod = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ad) %C(bold blue)<%an>%Creset"
alias glods = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ad) %C(bold blue)<%an>%Creset" --date=short
alias glog = git log --oneline --decorate --graph
alias gloga = git log --oneline --decorate --graph --all
alias glol = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ar) %C(bold blue)<%an>%Creset"
alias glola = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ar) %C(bold blue)<%an>%Creset" --all
alias glols = git log --graph --pretty="%Cred%h%Creset -%C(auto)%d%Creset %s %Cgreen(%ar) %C(bold blue)<%an>%Creset" --stat
alias gm = git merge
alias gma = git merge --abort
alias gmc = git merge --continue
alias gmff = git merge --ff-only
alias gms = git merge --squash
alias gmtl = git mergetool --no-prompt
alias gmtlvim = git mergetool --no-prompt --tool=vimdiff
alias gp = git push
alias gpd = git push --dry-run
alias gpf = git push --force-with-lease --force-if-includes
alias gpoat = git push origin --all and git push origin --tags
alias gpod = git push origin --delete
alias gpr = git pull --rebase
alias gpra = git pull --rebase --autostash
alias gprav = git pull --rebase --autostash -v
alias gpristine = git reset --hard and git clean --force -dfx
alias gprv = git pull --rebase -v
alias gpu = git push upstream
alias gpv = git push --verbose
alias gr = git remote
alias gra = git remote add
alias grb = git rebase
alias grba = git rebase --abort
alias grbc = git rebase --continue
alias grbi = git rebase --interactive
alias grbo = git rebase --onto
alias grbs = git rebase --skip
alias grev = git revert
alias greva = git revert --abort
alias grevc = git revert --continue
alias grf = git reflog
alias grh = git reset
alias grhh = git reset --hard
alias grhk = git reset --keep
alias grhs = git reset --soft
alias grm = git rm
alias grmc = git rm --cached
alias grmv = git remote rename
alias grrm = git remote remove
alias grs = git restore
alias grset = git remote set-url
alias grss = git restore --source
alias grst = git restore --staged
alias gru = git reset --
alias grup = git remote update
alias grv = git remote --verbose
alias gsb = git status --short --branch
alias gsd = git svn dcommit
alias gsh = git show
alias gsi = git submodule init
alias gsps = git show --pretty=short --show-signature
alias gsr = git svn rebase
alias gss = git status --short
alias gst = git status
alias gsta = git stash push
alias gstaa = git stash apply
alias gstall = git stash --all
alias gstc = git stash clear
alias gstd = git stash drop
alias gstl = git stash list
alias gstp = git stash pop
alias gsts = git stash show --patch
alias gsu = git submodule update
alias gsw = git switch
alias gswc = git switch --create
alias gta = git tag --annotate
alias gts = git tag --sign
alias gtv = git tag | sort -V
alias gunignore = git update-index --no-assume-unchanged
alias gwch = git whatchanged -p --abbrev-commit --pretty=medium
alias gwipe = git reset --hard and git clean --force -df
alias gwt = git worktree
alias gwta = git worktree add
alias gwtls = git worktree list
alias gwtmv = git worktree move
alias gwtrm = git worktree remove
```

## Nushell 유용한 기본 명령 및 활용법

기본 명령어는 다음과 같습니다.

```sh
ls           # 디렉토리 목록 표시
cd 폴더명    # 폴더 이동
rm 파일명    # 파일 삭제
open 파일명  # 파일 내용 보기 (cat과 유사)
sys          # 시스템 정보 확인
```

- 수정된 구성을 터미널에서 바로 반영

    ```sh
    source $nu.env-path
    ```

- 명령어 치환

    ```sh
    ~\tmp\for-isolated-network> let branch_name = (git rev-parse --abbrev-ref HEAD)
    ~\tmp\for-isolated-network> echo $branch_name
    main
    ```

- 파일 및 디렉터리 구조 탐색:

    ```sh
    ls | where type == dir
    ```

- 간편한 파일 검색:

    ```sh
    ls **/*.{txt,nu,rs} | where name =~ "pattern"
    ```

- 프로세스 관리:

    ```sh
    ps | where name =~ "git"
    ```

- 시스템 정보

    ```sh
    ~> sys
    View information about the system.

    You must use one of the following subcommands. Using this command as-is will only produce this help message.

    Usage:
      > sys

    Subcommands:
      sys cpu - View information about the system CPUs.
      sys disks - View information about the system disks.
      sys host - View information about the system host.
      sys mem - View information about the system memory.
      sys net - View information about the system network interfaces.
      sys temp - View the temperatures of system components.
      sys users - View information about the users on the system.

    Flags:
      -h, --help: Display the help message for this command

    Input/output types:
      ╭───┬─────────┬────────╮
      │ # │  input  │ output │
      ├───┼─────────┼────────┤
      │ 0 │ nothing │ record │
      ╰───┴─────────┴────────╯

    Examples:
      Show info about the system
      > sys

    ~> sys mem
    ╭────────────┬──────────╮
    │ total      │ 8.5 GB   │
    │ free       │ 5.4 GB   │
    │ used       │ 3.0 GB   │
    │ available  │ 5.4 GB   │
    │ swap total │ 536.8 MB │
    │ swap free  │ 536.8 MB │
    │ swap used  │ 0 B      │
    ╰────────────┴──────────╯
    ```

## [Coming from Bash](https://www.nushell.sh/book/coming_from_bash.html#command-equivalents)

Bash와 Nu에서의 명령어 차이를 보여줍니다.

## Oh-My-Posh 설정

[Oh My Posh GitHub 릴리스 페이지](https://github.com/JanDeDobbeleer/oh-my-posh/releases)에서 적절한 파일을 다운로드합니다.

`C:\path\to\oh-my-posh\`에 `oh-my-posh.exe`로 파일명을 변경하여 저장합니다.
그리고 시스템의 PATH 환경 변수에 추가하여, 터미널에서 oh-my-posh 명령을 전역적으로 사용할 수 있도록 설정합니다.

Nushell의 설정 파일인 `config.nu` 파일에 다음 내용을 추가하여 Oh My Posh를 초기화하고 프롬프트를 설정합니다.

우선 oh-my-posh가 정상적으로 실행되는지 확인합니다.

```sh
C:\Users\me\dev\omp\posh-windows-amd64.exe print primary --config C:\Users\me\dev\omp\theme\agnoster.omp.json
```

참고로 구성 파일 경로는 `""` 쌍따옴표로 감싸면 에러가 발생하므로, 쌍따옴표로 감싸지 않습니다.

```sh
> C:\Users\me\dev\omp\posh-windows-amd64.exe debug --config "C:/Users/me/dev/omp/theme/agnosterplus.omp"

Version: 25.4.3

Shell: nu

Prompt:

 me   for-isolated-network   main ≡  ?3    CONFIG ERROR                                   in nu at 23:41:22
Segments:

ConsoleTitle(true) -   0 ms
Session(true)      -   0 ms
Path(true)         -   5 ms
Git(true)          - 553 ms
Root(true)         -   0 ms
Status(true)       -   0 ms
Node(false)        -   0 ms
Go(false)          -   0 ms
Python(false)      -   0 ms
Shell(true)        -   0 ms
Time(true)         -   0 ms

Run duration: 563.8179ms

Cache path: C:\Users\me\AppData\Local\oh-my-posh

Config path: C:\Users\me\dev\omp\theme\agnosterplus.omp

Logs:

[DEBUG] 23:41:21.490 debug.go:36 → debug mode enabled
[TRACE] 23:41:21.498 load.go:Path() - 0s
[ERROR] 23:41:21.498 load.go:loadConfig:133 → open C:\Users\me\dev\omp\theme\agnosterplus.omp: The system cannot find the file specified.
[TRACE] 23:41:21.498 load.go:loadConfig() - 0s
```

그러면 `config.nu`에 다음 라인을 추가합니다.
여기서는 vscode를 에디터로 사용합니다.

```sh
code $nu.config-path
```

```nu
$env.PROMPT_COMMAND = {|| C:\Users\me\dev\omp\posh-windows-amd64.exe print primary --config C:\Users\me\dev\omp\theme\agnosterplus.omp.json }
```

이모지가 깨질 경우 [nerd hack 폰트](https://www.nerdfonts.com/)가 필요합니다.

윈도우 터미널 사용할 경우:
1. Windows Terminal을 연 후 `Ctrl + ,`(설정 창 열기)
2. Nushell 프로필이 없다면 생성하고, 해당 프로필을 선택합니다.
3. '모양'에서 폰트를 Nerd Font로 수정합니다.
4. 쉘 또는 터미널을 재시작합니다.
