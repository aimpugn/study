# PHP source

- [PHP source](#php-source)
    - [php-src 5.6.20 빌드 및 실행](#php-src-5620-빌드-및-실행)
        - [clone `php-src` and checkout `php-5.6.40`](#clone-php-src-and-checkout-php-5640)
        - [Dockerfile](#dockerfile)
        - [build \& run](#build--run)
        - [configure \& make \& test](#configure--make--test)
    - [테스트 코드 실행해보기](#테스트-코드-실행해보기)
        - [특정 테스트 케이스 실행해보기](#특정-테스트-케이스-실행해보기)
        - [코드 수정 후 출력해보기](#코드-수정-후-출력해보기)
    - [그냥 php5.6에서 테스트 실행해보기](#그냥-php56에서-테스트-실행해보기)
    - [참고](#참고)

## php-src 5.6.20 빌드 및 실행

### clone `php-src` and checkout `php-5.6.40`

```bash
git clone https://github.com/php/php-src.git

git checkout php-5.6.40
```

### Dockerfile

```dockerfile
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

## Bison 2.7 설치
RUN wget https://ftp.gnu.org/gnu/bison/bison-2.7.tar.gz && tar -xzf bison-2.7.tar.gz
RUN (cd bison-2.7 && ./configure && make && make install)

## Re2c 설치
RUN wget https://sourceforge.net/projects/re2c/files/1.0.1/re2c-1.0.1.tar.gz && tar -xzf re2c-1.0.1.tar.gz
RUN (cd re2c-1.0.1 && ./configure && make && make install)

## Automake 최신 버전 설치
RUN wget https://ftp.gnu.org/gnu/automake/automake-1.16.tar.gz && tar -xzf automake-1.16.tar.gz
RUN (cd automake-1.16 && ./configure && make && make install)

WORKDIR /php-src

## build 시에 실행해도 되고, 도커 안에서 실행해도 된다
## RUN ./configure --disable-all && make
## make 하면 sapi/cli/php 가 생성된다

CMD ["bash"]
```

- 참고로 bison, re2c 등의 버전은, [INSTALL](https://github.com/php/php-src/blob/PHP-5.6.20/INSTALL) 이란 문서 참고 했는데, 그대로 되지는 않아서 몇 시도해 봐서 되는 걸로 사용해봤습니다.
- gcc 7.5.0 으로도 정상 실행 돼서, gcc++-5 설치 부분 삭제

### build & run

```bash
docker build . --tag php-src

docker run -v "$(pwd):/php-src" --name php-src --rm -it php-src
```

### configure & make & test

```bash
## generate configure script
## You should not run buildconf in a release package 라고 막혀서 `--force` 사용
./buildconf --forcem

## `./configure --help` 사용해서 추가 옵션 확인 가능. 여기서는 모두 비활성화
./configure --disable-all

## make executable
make

## 빌드 완료후 실행 가능 파일이 `./sapi/cli/php`에 위치
./sapi/cli/php --version
## PHP 5.6.40 (cli) (built: Jan 16 2024 08:27:08) 
## Copyright (c) 1997-2016 The PHP Group
## Zend Engine v2.6.0, Copyright (c) 1998-2016 Zend Technologies
```

> You should not run buildconf in a release package?
>
>
> 개발 버전은 자주 변경되므로, `configure` 스크립트도 자주 업데이트된다. `configure` 스크립트는 `autoconf` 도구를 사용하여 생성되는데, 개발 과정에서 변경된 소스 코드와 설정을 반영해야 한다. 따라서 개발 버전에서는 `configure` 스크립트가 포함되지 않고, 대신 `buildconf`를 사용하여 이를 생성하게 한다.
>
> 반면 릴리즈 버전에는 이미 생성된 `configure` 스크립트가 포함되어 있다.(근데 php-5.6.20 으로 체크아웃하면 없다…) 이 스크립트는 릴리스 시점의 PHP 소스 코드에 맞춰져 있으며, 사용자가 별도로 `buildconf`를 실행할 필요가 없다.
>
> 이 메시지는 릴리스 버전의 PHP 소스 코드에서 `buildconf`를 실행하려 할 때 나타나는데, 릴리즈 버전에서 일반적으로 필요하지 않는 작업이기 때문에, 사용자가 실수로 이 명령을 실행하지 않도록 경고하는 것이라고 한다.
>

## 테스트 코드 실행해보기

### 특정 테스트 케이스 실행해보기

```bash
## ERROR: environment variable TEST_PHP_EXECUTABLE must be set to specify PHP executable!
export TEST_PHP_EXECUTABLE=/php-src/sapi/cli/php

$TEST_PHP_EXECUTABLE ./run-tests.php ./ext/pcre/tests/004.phpt
```

### 코드 수정 후 출력해보기

- `zval` 타입별로 출력하는 함수

    ```c
    ##include <stdio.h>
    void print_zval(zval *val, const char *spec) {
        // zval의 타입에 따라 다른 출력
        switch (Z_TYPE_P(val)) {
            case IS_NULL:
                printf("Spec: %s, \t\t\tType: NULL\n", spec);
                break;
            case IS_LONG:
                printf("Spec: %s, \t\t\tType: Long, Value: %ld\n", spec, Z_LVAL_P(val));
                break;
            case IS_DOUBLE:
                printf("Spec: %s, \t\t\tType: Double, Value: %f\\n", spec, Z_DVAL_P(val)); // %f를 사용하여 double 값을 출력
                break;
            case IS_BOOL:
                printf("Spec: %s, \t\t\tType: Bool, Value: %s\\n", spec, Z_BVAL_P(val) ? "true" : "false"); // 불리언 값을 문자열로 출력
                break;
            case IS_ARRAY:
                printf("Spec: %s, \t\t\tType: Array\n", spec);
                // 배열의 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
                break;
            case IS_OBJECT:
                printf("Spec: %s, \t\t\tType: Object\n", spec);
                // 객체의 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
                break;
            case IS_STRING:
                printf("Spec: %s, \t\t\tType: String, Value: %s\n", spec, Z_STRVAL_P(val));
                break;
            case IS_RESOURCE:
                printf("Spec: %s, \t\t\tType: Resource, Resource ID: %ld\n", spec, Z_RESVAL_P(val));
                break;
            case IS_CONSTANT:
                // 상수 타입의 경우, 상수의 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
                printf("Spec: %s, \t\t\tType: Constant\n", spec);
                break;
            case IS_CONSTANT_AST:
                // AST 노드의 경우, 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
                printf("Spec: %s, \t\t\tType: Constant AST\n", spec);
                break;
            case IS_CALLABLE:
                // 콜백의 경우, 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
                printf("Spec: %s, \t\t\tType: Callable\n", spec);
                break;
            default:
                printf("Spec: %s, \t\t\tUnknown type\n", spec);
                break;
        }
    }
    ```

- 코드 수정

    ```c
    ##include <stdio.h>
    
    static const char *zend_parse_arg_impl(int arg_num, zval **arg, va_list *va, const char **spec, char **error, int *severity TSRMLS_DC) /* {{{ */
    {
        const char *spec_walk = *spec;
        char c = *spec_walk++;
        int check_null = 0;
    
        print_zval(*arg);
    ```

- 다시 빌드

    ```bash
    make
    ```

- test.phpt 파일

    ```c
    --TEST--
    simple test
    --FILE--
    <?php
    
    ?>
    --EXPECTF--
    ```

- 실행

    ```bash
    $TEST_PHP_EXECUTABLE ./run-tests.php ./test.phpt
    ```

    - 결과

        ```bash
        Spec: zH,                       Type: String, Value: GZIP_POST
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: DEFLATE_POST
        Spec: H,                        Type: Array
        Spec: z|l,                      Type: NULL
        Spec: Z,                        Type: String, Value: /php-src/test.phpt
        Spec: z|l,                      Type: NULL
        Spec: pz/|lr!,                  Type: String, Value: /php-src/test.php
        Spec: z/|lr!,                   Type: String, Value: <?php
        
        ?>
        
        Spec: lr!,                      Type: Long, Value: 0
        Spec: zH,                       Type: String, Value: GET
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: COOKIE
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: POST_RAW
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: PUT
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: POST
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: GZIP_POST
        Spec: H,                        Type: Array
        Spec: zH,                       Type: String, Value: DEFLATE_POST
        Spec: H,                        Type: Array
        Spec: saz|s!a!a!,                       Type: String, Value: /php-src/sapi/cli/php   -d "output_handler=" -d "open_basedir=" -d "safe_mode=0" -d "disable_functions=" -d "output_buffering=Off" -d "error_reporting=32767" -d "display_errors=1" -d "display_startup_errors=1" -d "log_errors=0" -d "html_errors=0" -d "track_errors=1" -d "report_memleaks=1" -d "report_zend_debug=0" -d "docref_root=" -d "docref_ext=.html" -d "error_prepend_string=" -d "error_append_string=" -d "auto_prepend_file=" -d "auto_append_file=" -d "magic_quotes_runtime=0" -d "ignore_repeated_errors=0" -d "precision=14" -d "memory_limit=128M" -d "log_errors_max_len=0" -d "opcache.fast_shutdown=0" -d "opcache.file_update_protection=0" -f "/php-src/test.php"  2>&1
        Spec: az|s!a!a!,                        Type: Array
        Spec: z|s!a!a!,                         Type: NULL
        Spec: s!a!a!,                   Type: String, Value: /php-src
        Spec: a!a!,                     Type: Array
        Spec: a!,                       Type: Array
        ```

## 그냥 php5.6에서 테스트 실행해보기

```bash
## 설치: https://github.com/shivammathur/homebrew-php
brew tap shivammathur/php

brew install shivammathur/php/php@5.6

brew link --overwrite php@5.6

which php
## /opt/homebrew/bin/php
export TEST_PHP_EXECUTABLE=/opt/homebrew/bin/php
$TEST_PHP_EXECUTABLE ./run-tests.php ./ext/pcre/tests/004.phpt
```

## 참고

- [https://github.com/php/php-src/blob/PHP-5.6.20/INSTALL](https://github.com/php/php-src/blob/PHP-5.6.20/INSTALL)
