# date

- [date](#date)
    - [MacOS의 `date`는 linux 기반 `date`와 다르다](#macos의-date는-linux-기반-date와-다르다)
    - [날짜를 문자열로 출력](#날짜를-문자열로-출력)
    - [format 지정](#format-지정)
        - [유저 정의 포맷](#유저-정의-포맷)
    - [KST 시간을 epoch 시간으로 출력](#kst-시간을-epoch-시간으로-출력)
    - [epoch to KST](#epoch-to-kst)

## MacOS의 `date`는 linux 기반 `date`와 다르다

```bash
brew install coreutils
```

```bash
# date of linux based OS
alias date='gdate'
```

## 날짜를 문자열로 출력

```bash
# ko_KR.utf8 locale 시
$ date

2020. 04. 12. (일) 13:16:43 KST
```

```bash
# C locale 시
$ date

Sun Apr 12 13:17:34 KST 2020
```

## format 지정

```bash
$ date "+%Y-%m-%d"

2020-04-12
```

```sh
date +'%Y%m%d_%H%M%S'
20240910_231928
```

### 유저 정의 포맷

- 포맷(`-f`) 사용 시 `+` 문자가 가장 앞에 있다면 유저 정의 포맷으로 출력함을 의미
- `+` 사인이 없다면 시스템 개념의 현재 날짜와 시간 설정을 위한 값으로 해석된다

```shell
cc      Century (either 19 or 20) prepended to the abbreviated year.
yy      Year in abbreviated form (e.g., 89 for 1989, 06 for 2006).
mm      Numeric month, a number from 1 to 12.
dd      Day, a number from 1 to 31.
HH      Hour, a number from 0 to 23.
MM      Minutes, a number from 0 to 59.
ss      Seconds, a number from 0 to 60 (59 plus a potential leap second)
```

## KST 시간을 epoch 시간으로 출력

```bash
$ date -d '2023-12-20 15:00:00 KST' +%s
1703052000
```

```bash
function epoch() {
    local time="${1}"
    local timezone="${2:=KST}"
    date -d "$time $timezone" +%s
}
```

## epoch to KST

```bash
gdate -d "@1705299389" "+%Y-%m-%d %H:%M:%S.%3N %:z %Z"
#2024-01-15 15:16:29.000 +09:00 KST
```

```shell
function kst() {
    local epoch="${1//}"
    gdate -d "@${epoch}" "+%Y-%m-%d %H:%M:%S.%3N %:z %Z"
}
```
