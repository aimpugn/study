# [set](https://linuxcommand.org/lc3_man_pages/seth.html)

## set?

> Set or unset values of shell options and positional parameters.

쉘(Shell)에서 옵션을 설정하거나 스크립트의 동작 방식을 변경하는 데 사용된다

## options

### `-e`(errexit)

스크립트가 오류를 만났을 때 즉시 종료하도록 한다
기본적으로 쉘 스크립트는 명령어가 실패해도 계속 실행되지만, `-e` 옵션을 사용하면 어떤 명령어가 실패하면 스크립트 실행이 중단되고 오류 코드와 함께 종료된다.

### `-u`(nounset)

선언되지 않은 변수를 사용하려고 하면 오류 메시지와 함께 스크립트가 종료되도록 한다.
이를 통해 오타나 변수 초기화 오류를 쉽게 발견할 수 있다.

### `-o pipefail`

파이프라인(pipeline)의 **모든** 명령어가 성공적으로 실행되어야 전체 파이프라인이 성공한 것으로 간주되도록 한다.
기본적으로 쉘은 파이프라인의 마지막 명령어의 종료 상태만을 고려한다.
`pipefail` 옵션을 사용하면 파이프라인 내의 모든 명령어가 성공해야 전체가 성공한 것으로 간주된다.

## `set -euo pipefail; IFS=$'\r\n'`

```shell
# 쉘 스크립트의 견고성과 정확성을 향상시키기 위해 자주 사용되는 구문
set -euo pipefail; IFS=$'\r\n'
```

- `IFS`(Internal Field Separator)
    - 쉘 스크립트에서 필드 구분자를 정의하는 특별한 변수
    - 기본적으로 `IFS`는 공백, 탭, 개행 문자로 설정되어 있다
    - 쉘에서 문자열을 단어나 라인으로 분할할 때 사용

## "안전 모드" 설정

```shell
# Bash 스크립트의 시작 부분에 자주 사용되는 "안전 모드" 설정
set -euo pipefail; IFS=$'\n\t'; set -x
```

1. `set -e`

   - 의미: "errexit" 옵션을 활성화합니다.
   - 동작: 스크립트 실행 중 어떤 명령어라도 0이 아닌 종료 상태(에러)를 반환하면 즉시 스크립트 실행을 중단합니다.
   - 이유: 에러가 발생했을 때 스크립트가 계속 실행되어 예상치 못한 결과를 만들어내는 것을 방지합니다.
   - 예시:

     ```bash
     #!/bin/bash
     set -e
     echo "This will print"
     non_existent_command  # 이 명령어는 실패하고, 스크립트는 여기서 종료됩니다.
     echo "This will not print"
     ```

2. `set -u`

   - 의미: "nounset" 옵션을 활성화합니다.
   - 동작: 정의되지 않은 변수를 사용하려고 하면 스크립트 실행을 중단합니다.
   - 이유: 변수 이름을 잘못 입력하거나 초기화하지 않은 변수를 사용하는 실수를 방지합니다.
   - 예시:

     ```bash
     #!/bin/bash
     set -u
     echo $DEFINED_VAR  # 이 변수가 정의되지 않았다면, 여기서 스크립트가 중단됩니다.
     ```

3. `set -o pipefail`

   - 의미: 파이프라인의 실패를 감지하도록 합니다.
   - 동작: 파이프라인에서 마지막 명령어뿐만 아니라 모든 명령어의 종료 상태를 확인합니다.
   - 이유: 파이프라인 중간에 발생한 오류를 놓치지 않도록 합니다.
   - 예시:

     ```bash
     #!/bin/bash
     set -eo pipefail
     non_existent_command | echo "Won't save you"  # 첫 번째 명령어의 실패로 스크립트가 중단됩니다.
     ```

4. `IFS=$'\n\t'`

   - 의미: Internal Field Separator(IFS)를 새 줄과 탭으로 설정합니다.
   - 동작: 단어 분할 시 공백을 구분자로 사용하지 않고, 새 줄과 탭만을 구분자로 사용합니다.
   - 이유: 파일명이나 기타 데이터에 포함된 공백으로 인한 예기치 않은 동작을 방지합니다.
   - 예시:

     ```bash
     #!/bin/bash
     IFS=$'\n\t'
     for file in $(ls); do
         echo "Processing $file"  # 파일명에 공백이 있어도 정상적으로 처리됩니다.
     done
     ```

    `IFS=$'\n\t'`에서 `$'string'`는 문자열 내의 이스케이프 시퀀스를 해석하기 위해 사용되는 Bash 구문([ANSI-C Quoting](https://www.gnu.org/software/bash/manual/bash.html#ANSI_002dC-Quoting))입니다.
    `$'string'` 문법에서 백슬래시로 이스케이프 된 문자들은 ANSI-C 표준으로 치환됩니다.
    `$'...'` 구문으로 사용할 수 있는 주요 이스케이프 시퀀스는 다음과 같습니다:
    - `$'\n'`: 새 줄(newline)

        ```sh
        ❯ echo $'Hello\nWorld'
        Hello
        World
        ```

    - `$'\r'`: 캐리지 리턴(carriage return)

        ```sh
        ❯ echo $'Hello\rWorld'
        World
        ```

    - `$'\t'`: 탭(tab)

        ```sh
        ❯ echo $'Hello\tWorld'
        Hello    World
        ```

    - `$'\0'`: 널 문자(null character)
    - `$'\a'`: 경고음(alert, bell)

        ```sh
        ❯ echo $'Hello\0World'; echo $'Hello\0World' | wc -c
        HelloWorld
              12
        ❯ echo $'HelloWorld'; echo $'HelloWorld' | wc -c
        HelloWorld
              11
        ```

    - `$'\b'`: 백스페이스(backspace)

        ```sh
        ❯ echo $'Hello\bWorld'
        # 문자 'o' 삭제
        HellWorld
        ```

    - `$'\f'`: 폼 피드(form feed)
    - `$'\v'`: 수직 탭(vertical tab)

        ```sh
        ❯ echo $'Hello\vWorld'
        Hello
            World
        ```

5. `set -x`

   - 의미: "xtrace" 옵션을 활성화합니다.
   - 동작: 실행되는 각 명령어를 확장된 형태로 표준 오류에 출력합니다.
   - 이유: 스크립트의 실행 과정을 추적하고 디버깅하는 데 도움이 됩니다.
   - 예시:

     ```bash
     #!/bin/bash
     set -x
     echo "Hello"
     # 출력:
     # + echo 'Hello'
     # Hello
     ```

이 모든 옵션을 함께 사용함으로써, 스크립트는 더욱 엄격하고 예측 가능한 방식으로 실행됩니다. 에러를 빠르게 감지하고, 디버깅을 용이하게 하며, 예상치 못한 동작을 최소화합니다. 이는 특히 복잡한 스크립트나 중요한 작업을 수행하는 스크립트에서 매우 유용합니다.
