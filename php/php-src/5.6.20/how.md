# run

- [run](#run)
    - [도구 및 라이브러리](#도구-및-라이브러리)
    - [빌드](#빌드)
        - [`buildconf`](#buildconf)
        - [`configure`](#configure)
        - [`make`](#make)
        - [`make install`](#make-install)
    - [기타](#기타)

## 도구 및 라이브러리

```Dockerfile
FROM ubuntu:18.04

RUN apt update && apt install -y \
    software-properties-common \
    git \
    wget \
    pkg-config \
    build-essential \
    autoconf \
    flex \
    libxml2-dev \
    libsqlite3-dev \
    libtool \
    openssl \
    libcurl4-openssl-dev \
    libgd-dev \
    vim

WORKDIR /temp

# Bison 2.7 설치
RUN wget https://ftp.gnu.org/gnu/bison/bison-2.7.tar.gz && tar -xzf bison-2.7.tar.gz
RUN (cd bison-2.7 && ./configure && make && make install)

# Re2c 설치
RUN wget https://sourceforge.net/projects/re2c/files/1.0.1/re2c-1.0.1.tar.gz && tar -xzf re2c-1.0.1.tar.gz
RUN (cd re2c-1.0.1 && ./configure && make && make install)

# Automake 최신 버전 설치
RUN wget https://ftp.gnu.org/gnu/automake/automake-1.16.tar.gz && tar -xzf automake-1.16.tar.gz
RUN (cd automake-1.16 && ./configure && make && make install)

WORKDIR /php-src

# build 시에 실행해도 되고, 도커 안에서 실행해도 된다
# RUN ./configure --disable-all && make
# make 하면 sapi/cli/php 가 생성된다

CMD ["bash"]
```

- `software-properties-common`
    - `add-apt-repository` 등의 명령어를 사용하기 위해 필요한 패키지
- `autoconf`
    - 소스 코드 패키지를 다양한 유닉스 계열 시스템에서 컴파일할 수 있도록 도와주는 도구
    - 시스템의 특성과 기능을 탐지하여 플랫폼에 맞게 소스 코드를 구성하고 적절한 컴파일러 옵션을 설정하는 `configure` 스크립트를 생성한다.
- `bison`(GNU Project parser generator (yacc replacement))
    - `Yacc` 호환 파서 생성기.
    - 이는 소스 코드 내에서 정의된 문법 규칙을 바탕으로 구문 분석기(parser)를 생성한다
    - PHP 언어의 구문을 해석하고 분석하는 데 사용되며, PHP 스크립트의 구문을 분석하고 실행 가능한 코드로 변환하는 데 필요하다.
- `re2c`
    - `lex` 호환 렉서 생성기. 주로 정규 표현식을 이용하여 렉서(lexer)를 생성한다.
    - C 언어 코드로 직접 변환되며, PHP 언어의 렉싱(lexical analysis) 단계에서 사용된다.
    - `re2c`로 생성된 렉서는 PHP 언어의 구문을 해석하고 분석하는 데 사용되며, PHP 스크립트의 구문을 분석하고 실행 가능한 코드로 변환하는 데 필요하다.
- `libxml2`
    - XML 처리를 위한 C 라이브러리
    - XML 파싱, DOM 문서 처리, XPath 쿼리 수행 등을 할 때 사용
- `sqlite3`
    - 경량의 파일 기반 데이터베이스
    - PHP에서 SQLite 데이터베이스를 관리하고 조작하는 데 사용
- `flex`
    - `lex`의 개선된 버전으로, C 언어로 작성된 렉서 생성기
    - 다양한 프로그래밍 언어와 도구에서 소스 코드의 토큰 분석에 사용된다.
    - 사용하기 쉽고, 많은 시스템에서 널리 지원되며, 이식성이 높다.
    - 사용자가 정의한 정규 표현식에 기반하여 렉서를 생성하며, 이를 통해 소스 코드를 토큰화

## 빌드

PHP 소스 코드를 컴파일하고 설치하는 과정에는 `buildconf`, `configure`, `make`, `make install` 등의 명령어가 사용된다.

### `buildconf`

```bash
# "You should not run buildconf in a release package." 라는 에러가 발생해서 `--force` 옵션
./buildconf --force
```

- `buildconf`는 주로 Git에서 직접 체크아웃한 소스 코드와 같이 개발 버전의 PHP를 빌드할 때 사용된다.
- 이 명령어는 `autoconf`와 `automake` 도구를 사용하여, 소스 코드를 컴파일하기 위한 환경 설정 스크립트(`configure` 스크립트)를 생성
- 다양한 시스템 환경에 맞춰 PHP를 컴파일할 수 있도록 필요한 설정을 자동으로 감지하고 구성하기 위해 필요하다.

> **You should not run buildconf in a release package 발생 이유**
>
> 개발 버전은 자주 변경되므로, `configure` 스크립트도 자주 업데이트되어야 한다.
> `configure` 스크립트는 `autoconf` 도구를 사용하여 생성되며, 개발 과정에서 변경된 소스 코드와 설정을 반영해야 한다.
> 따라서 개발 버전에서는 `configure` 스크립트가 포함되지 않고, 대신 `buildconf`를 사용하여 이를 생성한다.
>
> 반면 릴리스 버전에는 이미 생성된 `configure` 스크립트가 포함되어 있다.
> 이 스크립트는 릴리스 시점의 PHP 소스 코드에 맞춰져 있으며, 사용자가 별도로 `buildconf`를 실행할 필요가 없다.
>
> 이 메시지는 릴리스 버전의 PHP 소스 코드에서 `buildconf`를 실행하려 할 때 나타난다.
> 이는 일반적으로 필요하지 않는 작업이기 때문에, 사용자가 실수로 이 명령을 실행하지 않도록 경고하는 것.

### `configure`

```bash
# 굳이 extension 필요 없으므로 disable-all
./configure --disable-all
```

- 시스템 환경을 감지하고 PHP를 해당 환경에 맞게 **컴파일하기 위한 옵션과 변수를 설정**합니다.
- `./configure` 스크립트는 시스템의 컴파일러, 라이브러리, 도구 등을 확인하고, PHP가 사용할 수 있는 기능을 결정한다
- 사용자는 `./configure` 명령어에 다양한 옵션을 전달하여 필요한 PHP 확장, 옵션 등을 선택적으로 활성화하거나 비활성화할 수 있다.
- PHP를 컴파일하기 전에 시스템 환경에 맞게 소스 코드를 구성하고, 사용자의 요구 사항에 맞게 PHP를 맞춤 설정하는 데 필요하다.

### `make`

```bash
make
```

- `make` 명령어는 `Makefile`을 사용하여 소스 코드를 컴파일하고, 필요한 라이브러리를 링크하며, 실행 가능한 PHP 바이너리를 생성한다
- 이 과정에서 소스 코드가 기계어로 번역되고, 필요한 모든 컴포넌트가 하나의 실행 파일로 결합된다.

### `make install`

- `make install` 명령어는 컴파일된 PHP 바이너리, 설정 파일, 라이브러리 등을 시스템의 표준 위치에 복사하여 시스템 전체에서 PHP를 사용할 수 있도록 한다.

## 기타

- [Git Access](https://www.php.net/git.php)
