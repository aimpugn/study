# hyphen(`-`)

- [hyphen(`-`)](#hyphen-)
    - [`-` (하이픈)의 의미와 사용](#--하이픈의-의미와-사용)
    - [주요 사용 사례](#주요-사용-사례)
    - [동작 방식](#동작-방식)
    - [구체적인 예시와 비교](#구체적인-예시와-비교)
    - [INVOCATION](#invocation)
    - [기타](#기타)

## `-` (하이픈)의 의미와 사용

`-`는 많은 Unix/Linux 명령어에서 "로그인 셸" 또는 "새로운 환경"을 시뮬레이트하라는 의미로 사용됩니다.
이는 옵션이라기보다는 특별한 인자(argument)로 취급됩니다.

POSIX 표준의 일부입니다. 역사적으로 Bourne 셸에서 유래했으며, 현대의 셸들에서도 이 관행을 유지하고 있습니다.

`-`를 사용하면 새로운 환경을 제공하므로, 이전 환경의 잠재적인 보안 위험을 줄일 수 있습니다.

## 주요 사용 사례

1. su 명령어

    ```sh
    # switch to root user
    su -

    su - username
    ```

2. ssh 명령어

    ```sh
    ssh -
    ```

3. login 명령어

    ```sh
    login -
    ```

4. bash 명령어

    ```sh
    # 여기서 `-l`은 `-`와 동일합니다.
    bash -l
    ```

    - `-l`: bash가 로그인 셸로 호출된 것처럼 동작하게 만듭니다. (see `INVOCATION` below)

## 동작 방식

- 사용자의 환경을 완전히 초기화합니다.
- 대상 사용자의 프로필 스크립트(예: `.profile`, `.bash_profile`)를 실행합니다.
- 새로운 환경 변수 세트를 로드합니다.
- 작업 디렉토리를 사용자의 홈 디렉토리로 변경합니다.

## 구체적인 예시와 비교

1. su와 su - 비교

    ```sh
    # 현재 환경:

    ~$ export PATH="$PATH:/home/test/bin"
    ~$ echo $PATH
    /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/home/test/bin

    # root의 기본 PATH
    ~$ sudo su -
    root@ip-w-x-y-z:~# echo $PATH
    /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin # root의 기본 PATH
    ```

2. bash와 bash - 비교

    ```sh
    # 현재 셸:
    echo $CUSTOM_VAR
    my_custom_value

    # 새 bash 세션
    bash
    echo $CUSTOM_VAR
    my_custom_value  # 상속됨

    # 로그인 셸로 bash 시작
    exit
    bash -
    echo $CUSTOM_VAR
    # 출력 없음 (변수가 설정되지 않음)
    ```

## INVOCATION

A login shell is one whose first character of argument zero is a `-`, or one started with the `--login` option.

An interactive shell is one started without non-option arguments and without the `-c` option whose standard input and error are both connected to terminals (as determined by isatty(3)), or one started with the `-i` option.
*PS1* is set and `$-` includes `i` if `bash` is interactive, allowing a shell script or a startup file to test this state.

The following paragraphs describe how bash executes its startup files.  If any of the files exist but cannot be read, bash reports an error.  Tildes are expanded in file names as described below under *Tilde Expansion* in the *EXPANSION* section.

When `bash` is invoked as an interactive login shell, or as a non-interactive shell with the `--login` option, it first reads and executes commands from the file `/etc/profile`, if that file exists. After reading that file, it looks for `~/.bash_profile`, `~/.bash_login`, and `~/.profile`, in that order, and reads and executes commands from the first one that exists and is readable. The `--noprofile` option may be used when the shell is started to inhibit this behavior.

When a login shell exits, bash reads and executes commands from the file `~/.bash_logout`, if it exists.

When an interactive shell that is not a login shell is started, bash reads and executes commands from `~/.bashrc`, if that file exists. This may be inhibited by using the `--norc` option. The `--rcfile` file option will force bash to read and execute commands from file instead of `~/.bashrc`.

When `bash` is started non-interactively, to run a shell script, for example, it looks for the variable *BASH_ENV* in the environment, expands its value if it appears there, and uses the expanded value as the name of a file to read and execute. *Bash* behaves as if the following command were executed:

```sh
if [ -n "$BASH_ENV" ]; then . "$BASH_ENV"; fi
```

but the value of the `PATH` variable is not used to search for the file name.

If `bash` is invoked with the name `sh`, it tries to mimic the startup behavior of historical versions of `sh` as closely as possible, while conforming to the POSIX standard as well. When invoked as an interactive login shell, or a non-interactive shell with the `--login` option, it first attempts to read and execute commands from `/etc/profile` and `~/.profile`, in that order. The `--noprofile` option may be used to inhibit this behavior.  When invoked as an interactive shell with the name `sh`, `bash` looks for the variable *ENV*, expands its value if it is defined, and uses the expanded value as the name of a file to read and execute. Since a shell invoked as `sh` does not attempt to read and execute commands from any other startup files, the `--rcfile` option has no effect. A non-interactive shell invoked with the name sh does not attempt to read any other startup files.  When invoked as sh, bash enters posix mode after the startup files are read.

When `bash` is started in posix mode, as with the `--posix` command line option, it follows the POSIX standard for startup files.  In this mode, interactive shells expand the ENV variable and commands are read and executed from the file whose name is the expanded value.  No other startup files are read.

*Bash* attempts to determine when it is being run by the remote shell daemon, usually `rshd`.  If bash determines it is being run by `rshd`, it reads and executes commands from `~/.bashrc`, if that file exists and is readable. It will not do this if invoked as `sh`. The `--norc` option may be used to inhibit this behavior, and the `--rcfile` option may be used to force another file to be read, but rshd does not generally invoke the shell with those options or allow them to be specified.

If the shell is started with the effective user (group) id not equal to the real user (group) id, and the `-p` option is not supplied, no startup files are read, shell functions are not inherited from the environment, the *SHELLOPTS* variable, if it appears in the environment, is ignored, and the effective user id is set to the real user id. If the `-p` option is supplied at invocation, the startup behavior is the same, but the effective user id is not reset.

## 기타

- [What's the magic of "-" (a dash) in command-line parameters?](https://stackoverflow.com/questions/8045479/whats-the-magic-of-a-dash-in-command-line-parameters)
