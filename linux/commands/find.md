
# find

- [find](#find)
    - [경로 제외하기](#경로-제외하기)
    - [`.//`가 붙는 이유](#가-붙는-이유)
    - [찾은 디렉토리를 `-exec`으로 삭제하기](#찾은-디렉토리를--exec으로-삭제하기)

## 경로 제외하기

- [How do I exclude a directory when using `find`?](https://stackoverflow.com/a/15736463)

```shell
find -name "*.js" -not -path "./directory/*"
```

## `.//`가 붙는 이유

```sh
$ find ./ -typf f

.//.idea/shelf/Changes4.xml
.//.idea/shelf/Changes18/shelved.patch
.//.idea/shelf/Changes20/shelved.patch
```

단순히 현재 디렉토리를 나타내는 표시. 입력 경로가 `./`로 시작하면 출력 경로도 `.//`로 시작할 수 있다.
일반적으로 Unix/Linux 시스템에서는 경로 내의 여러 연속된 슬래시를 하나의 슬래시로 취급함
그러나 find 명령어에서는 입력 경로와 출력 경로가 정확하게 일치해야 하므로, 경로 내의 슬래시 수도 정확히 일치해야 한다

- `.//`: `./`와 동일하게 작동
    - `.`은 현재 디렉토리
    - 뒤에 붙은 `/`는 경로 구분자
    - 그리고 그 다음에 붙은 `/`는 루트 경로에서 시작하는 하위 경로를 나타낸다

```sh
find . -type f ! -path './.idea/*'

./.git/index
./.git/packed-refs
./.git/COMMIT_EDITMSG
./.git/FETCH_HEAD
```

그래서 경로를 `.`로 지정하면 `./`로 붙지 않기 때문에 원하는대로 나온다

## 찾은 디렉토리를 `-exec`으로 삭제하기

```shell
find . -name "*Psalm*" 2>/dev/null -exec rm -r "{}" \;
```

```bash
find . -name "ELB-API*" -exec mv {} ./ELB-API/ \;
```
