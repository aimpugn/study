# locale

- [locale](#locale)
    - [locale](#locale-1)
    - [로케일 카테고리](#로케일-카테고리)
    - [`locale-gen`](#locale-gen)
    - [`update-locale`](#update-locale)
    - [기타](#기타)

## [locale](https://man7.org/linux/man-pages/man1/locale.1.html)

로케일이란 특정 위치(국가, 지역 등)에 대한 모든 설정과 정보에 대한 추상화입니다.
가령, 리눅스에서 로케일 인프라스트럭처는 [GNC C 라이브러리](https://en.wikipedia.org/wiki/Glibc)가 제공합니다.
이때 다양한 것들이 로케일 인프라스트럭처의 일부로 함께 동작합니다.
- 현재 로케일에 기반하여 동작이 변하는 함수 집합. (ex. `printf` 함수 패밀리)
- 서로 다른 지역에서 어떻게 동작할 것인지 정보를 갖는 데이터베이스
- 환경변수 집합 등 현재 로케일을 설정하는 방법 등

그리고 "`locale` 명령어"는 현재 지역, 모든 지역에 대한 정보를 보여줍니다.
아무 인자 없이 호출되면, 현재 환경 변수를 기준으로 각 지역 카테고리([locale(5)](https://man7.org/linux/man-pages/man5/locale.5.html))에 대한 현재 지역 설정 정보를 보여줍니다.

## 로케일 카테고리

- `LC_COLLATE`: `locale` 을 이용하여 두 개의 문자열 비교하는 `strcoll`과 `locale` 을 사용해 문자열을 변환하는 `strxfrm` 함수에 적용됩니다. [Collation Functions](https://www.gnu.org/software/libc/manual/html_node/Collation-Functions.html)를 참고합니다.

- `LC_CTYPE`: 문자열의 분류와 변환과 멀티바이트와 [확장형 문자](https://seowoosung.github.io/etc/2018/10/10/wchar.html)에 적용됩니다. [Character Handling](https://www.gnu.org/software/libc/manual/html_node/Character-Handling.html)와 [Character Set Handling](https://www.gnu.org/software/libc/manual/html_node/Character-Set-Handling.html)를 참고합니다.

- `LC_MONETARY`: 화폐 서식에 적용됩니다. [Generic Numeric Formatting Parameters](https://www.gnu.org/software/libc/manual/html_node/General-Numeric.html)를 참고합니다.

- `LC_NUMERIC`: 화폐가 아닌 숫자 서식에 적용됩니다. [Generic Numeric Formatting Parameters](https://www.gnu.org/software/libc/manual/html_node/General-Numeric.html)를 참고합니다.

- `LC_TIME`: 날짜와 시간 서식에 적용됩니다. [Formatting Calendar Time](https://www.gnu.org/software/libc/manual/html_node/Formatting-Calendar-Time.html)를 참고합니다.

- `LC_MESSAGES`: UI에서 메시지 번역에 사용될 언어를 선택하는 데 적용되고, 긍정 및 부정 응답을 위한 정규 표현식을 포함합니다.
    - 영어(English) 로케일에서는 "yes", "y", "no", "n" 등의 응답을 처리.
    - 독일어(German) 로케일에서는 "ja", "j", "nein", "n" 등을 처리.

- `LC_ALL`: 카테고리가 아니고, `setlocale`를 사용하여 모든 경우에 대해 단일 로케일을 설정할 때 사용하는 매크로입니다.
    - 이 환경 변수를 설정하면, 기존의 `LC_*`변수 및 `LANG` 값보다 우선하여 모든 로케일 설정을 덮어씁니다.

- `LANG`: 이 환경 변수가 정의되어 있으면, 위에서 설정한 `LC_*` 변수에 의해 덮어씌워지지 않는 한 모든 로케일에 적용됩니다.

## `locale-gen`

`localedef`를 호출하여, `/usr/lib/locale/ko_KR.UTF-8` 등의 로케일 데이터를 생성하는 역할을 합니다.

따라서, 만약 `locale-gen`을 하지 않는다면 환경변수로 무엇을 설정하든 해당 로케일 정보를 사용할 수 없습니다.

`locale-gen` 전을 본다면 다음과 같이 `C.utf8`라는 디렉토리만 존재합니다.

```sh
➜  ~ cd /usr/lib/locale
➜  locale ll
total 0
drwxr-xr-x 1 root root 250 Feb 16 14:42 C.utf8
➜  locale ll C.utf8
total 396K
-rw-r--r-- 1 root root  127 Jan 28 17:07 LC_ADDRESS
-rw-r--r-- 1 root root 1.4K Jan 28 17:07 LC_COLLATE
-rw-r--r-- 1 root root 353K Jan 28 17:07 LC_CTYPE
-rw-r--r-- 1 root root  258 Jan 28 17:07 LC_IDENTIFICATION
-rw-r--r-- 1 root root   23 Jan 28 17:07 LC_MEASUREMENT
drwxr-xr-x 1 root root   30 Feb 16 15:48 LC_MESSAGES
-rw-r--r-- 1 root root  270 Jan 28 17:07 LC_MONETARY
-rw-r--r-- 1 root root   62 Jan 28 17:07 LC_NAME
-rw-r--r-- 1 root root   50 Jan 28 17:07 LC_NUMERIC
-rw-r--r-- 1 root root   34 Jan 28 17:07 LC_PAPER
-rw-r--r-- 1 root root   47 Jan 28 17:07 LC_TELEPHONE
-rw-r--r-- 1 root root 3.3K Jan 28 17:07 LC_TIME
```

`locale-gen ko_KR.UTF-8`을 실행하면 하나의 파일로 컴파일된 로케일 데이터를 생성합니다.

```sh
➜  locale sudo locale-gen ko_KR.UTF-8
Generating locales (this might take a while)...
  ko_KR.UTF-8... done
Generation complete.
➜  locale ll
total 1.4M
drwxr-xr-x 1 root root   44 Feb 16 15:48 C.utf8
-rw-r--r-- 1 root root 1.4M Feb 16 15:51 locale-archive
```

만약, `--no-archive` 옵션을 사용하면 `C.utf8`처럼 `LC_*` 파일들을 갖는 디렉토리로 생성합니다.

## `update-locale`

`/etc/default/locale` 파일을 수정하여 로그인 시 시스템에서 기본적으로 적용할 로케일을 설정합니다.
적용한 다음, 사용자가 로그인할 때 (`ssh`, `tty` 등) `/etc/default/locale`을 읽어 환경 변수로 설정됩니다.

- 업데이트 전

    ```sh
    ➜  C.utf8 cat /etc/default/locale
    LANG=C.UTF-8
    ```

- 업데이트 후

    ```sh
    ➜  ~ sudo update-locale LANG=ko_KR.UTF-8 LC_MESSAGES=POSIX
    ➜  ~ cat /etc/default/locale
    LANG=ko_KR.UTF-8
    LC_MESSAGES=POSIX
    ```

## 기타

- [What does locale-gen generate in Linux?](https://superuser.com/a/1552086)
- [7 Locales and Internationalization](https://www.gnu.org/software/libc/manual/html_node/Locales.html#Locales)
