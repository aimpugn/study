# pipe

## gotcha issue

- [Bash gotcha: cat FILE | while read A ; do ... exit .. done](https://stackoverflow.com/questions/62150604/bash-gotcha-cat-file-while-read-a-do-exit-done)

```sh
cat FILE | while read SPID ; do
    if ls $ODIR/*$SPID*.csv > /dev/null 2>&1 ; then
        echo Error: some files exist in the directory $ODIR:
        ls -la $ODIR/*$SPID*.csv
        exit
    fi
done
```

이 경우 이 코드를 실행한 스크립트가 아닌, `|` 우측의 서브 셸만 종료됩니다.

pipeline에 대한 [bash manual](https://www.gnu.org/savannah-checkouts/gnu/bash/manual/bash.html#Pipelines)을 보면, 파이프라인의 각 명령어는 각자의 서브 셸에서 실행된다고 합니다.

> Each command in a pipeline is executed in its own subshell [...]

```sh
( exit; )                   # exits the subshell only
: | exit                    # exits the subshell only
a=1; ( a=5; ); echo $a      # prints 1, a=5 is executed in a subshell
a=1; : | a=5;  echo $a      # prints 1, a=5 is executed in a subshell
```
