# dirname

- [dirname](#dirname)
    - [현재 경로 가져오기](#현재-경로-가져오기)
    - [현재 스크립트의 path 가져오기](#현재-스크립트의-path-가져오기)

## 현재 경로 가져오기

```shell
CURRENT_DIR=$(cd "$(dirname "$0")" && pwd -P)
```

## 현재 스크립트의 path 가져오기

```shell
# ㄴ https://stackoverflow.com/a/4774063
SCRIPTPATH="$( cd -- "$(dirname "$0" )" >/dev/null 2>&1 ; pwd -P )"
```
