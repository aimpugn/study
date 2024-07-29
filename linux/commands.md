# Commands

- [Commands](#commands)
    - [alias 내용 알아내기](#alias-내용-알아내기)
        - [`type` 명령어 사용](#type-명령어-사용)
        - [또는 `alias` 명령어 사용](#또는-alias-명령어-사용)
    - [/sbin/init](#sbininit)

## alias 내용 알아내기

- [How to see the command attached to a bash alias?](https://askubuntu.com/a/103524)

### `type` 명령어 사용

```shell
$ type grbc
grbc is an alias for git rebase --continue
$ type -a grbc
grbc is an alias for git rebase --continue
```

### 또는 `alias` 명령어 사용

```shell
$ alias

... 너무 많아서 생략 ...
md='mkdir -p'
port-restart='yarn --cwd /Users/rody/IdeaProjects/core-r1-docker restart-bg'
portrody='totp get vpn | vpn port rody'
vim=nvim
vimdiff='nvim -d'
which-command=whence
```

## /sbin/init
