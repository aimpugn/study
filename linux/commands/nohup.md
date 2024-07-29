# nohup

## nohup

> invoke a utility immune to hangups

The nohup utility invokes utility with its arguments and at this time sets the signal `SIGHUP` to be ignored.  If the standard output is a terminal, the standard output is appended to the file nohup.out in the current directory. If standard error is a terminal, it is directed to the same place as the standard output.

Some shells may provide a builtin `nohup` command which is similar or identical to this utility.  Consult the builtin(1) manual page.

`nohup` 유틸리티는 인자로 전달된 유틸리티를 호출하는데, 이때 신호 `SIGHUP`을 무시하도록 설정합니다.
표준 출력이 터미널인 경우, 표준 출력은 현재 디렉터리에 있는 `nohup.out` 파일에 추가됩니다.
표준 오류가 터미널인 경우 표준 출력과 같은 위치로 이동합니다.

일부 셸은 이 유틸리티와 유사하거나 동일한 내장 `nohup` 명령을 제공할 수 있습니다.
