# xargs

## xargs?

xargs는 "eXtended ARGuments"의 약자입니다. 이는 명령어 라인에서 사용되는 유닉스 및 유닉스 계열 시스템(예: 리눅스, macOS)의 명령어로, 표준 입력(예: 키보드, 다른 명령어의 출력)에서 데이터를 읽어, 이를 인수로 사용하여 다른 명령어를 실행할 수 있게 해줍니다. 기본적으로, xargs는 입력 데이터를 공백(또는 개행 문자)으로 분리해 각각을 명령어의 인수로 사용합니다. 하지만, `-d` 옵션을 통해 구분자를 지정할 수도 있습니다.

xargs의 가장 일반적인 사용 사례 중 하나는 대량의 파일에 대해 특정 명령어를 실행할 때입니다. 예를 들어, `find` 명령어로 특정 조건을 만족하는 파일들을 찾고, 이 결과를 `xargs`를 통해 다른 명령어에 전달하여 처리할 수 있습니다.

xargs는 유연하고 강력한 도구입니다만, 특히 파일명이나 다른 입력 데이터에 공백, 특수 문자 등이 포함되어 있을 때는 예상치 못한 방식으로 동작할 수 있으므로 사용 시 주의가 필요합니다. 이를 위해, `-0` 옵션과 함께 `find` 명령어의 `-print0` 옵션을 사용하여 NULL 문자로 항목을 구분하도록 설정할 수 있습니다. 이 방법은 파일명에 공백이 포함되어 있을 때 유용합니다.

## 사용 예시

1. `find` 명령어와 함께 사용하여 특정 패턴의 파일에 대해 명령 실행:

    ```bash
    find . -type f -name "*.txt" | xargs grep "특정 문자열"
    ```

2. 사용자 정의 문자로 레코드 구분하기:

    ```bash
    echo "one,two,three" | xargs -d "," -n 1 echo
    ```

    이 예에서는 ','를 구분자로 사용하여 문자열 "one,two,three"를 분리하고, 각각에 대해 'echo' 명령어를 실행합니다. `-n 1` 옵션은 한 번에 하나의 인수만 명령어에 전달하도록 지정합니다.

3. NULL 문자를 사용하여 안전하게 파일 처리하기:

    ```bash
    find . -type f -print0 | xargs -0 rm
    ```

이 명령은 현재 디렉토리에서 모든 파일을 찾아(`find . -type f -print0`), 찾은 파일을 `rm` 명령어로 전달하여 삭제합니다. `-print0`과 `-0` 옵션은 파일명에 공백이나 특수 문자가 포함되어 있어도 안전하게 처리될 수 있도록 합니다.

## Bash -c 옵션을 사용하여 xargs 내에서 함수 호출하기

`kst: command not found` 오류 메시지는 `bash -c`를 사용할 때 발생하는데, 이는 `kst` 함수가 새로운 `bash` 세션에서 인식되지 않기 때문입니다.
`export -f kst`를 사용하더라도, `xargs`와 함께 `bash -c`로 새로운 셸 세션을 시작할 때, 그 세션에서는 `kst` 함수가 정의되어 있지 않습니다.
이는 각 셸 세션이 독립적인 환경을 가지며, 함수의 export가 `bash -c`로 시작된 셸에 상속되지 않기 때문입니다.

### 함수를 인라인으로 전달

함수의 내용을 직접 `bash -c`에 전달하는 방식입니다. 이 방법은 함수가 비교적 간단할 때 유용합니다.

예를 들어, `kst` 함수의 본문을 직접 `bash -c`의 인자로 넣는 방식입니다:

```bash
pbpaste | jq '.response | .[].start' | xargs -I {} bash -c 'local epoch="$1"; gdate -d "@${epoch}" "+%Y-%m-%d %H:%M:%S.%3N %:z %Z"' _ {}
```

### 환경 변수를 사용하여 함수 정의 전달

복잡한 함수의 경우, 함수 정의를 환경 변수에 저장하고 `bash -c` 내에서 이를 평가(eval)하여 사용하는 방법입니다. 이 방법은 함수를 문자열 형태로 환경 변수에 할당하고, 새로운 셸 세션에서 이 환경 변수를 통해 함수를 다시 정의하는 방식을 사용합니다.

1. 함수를 환경 변수에 할당합니다:

    ```bash
    KST_FUNC=$(declare -f kst)
    ```

2. `xargs`와 함께 사용할 때, 환경 변수와 `bash -c`를 사용하여 함수를 호출합니다:

    ```bash
    pbpaste | jq '.response | .[].start' | xargs -I {} bash -c "eval '$KST_FUNC'; kst \"\$@\"" _ {}
    ```

#### 명령어 분석

```bash
pbpaste | jq '.response | .[].start' | xargs -I {} bash -c "eval '$KST_FUNC'; kst \"\$@\"" _ {}
```

1. **`pbpaste`**: macOS에서 사용되는 명령어로, 클립보드에 저장된 내용을 출력합니다. 이 명령어는 파이프(`|`)를 통해 다음 명령어로 데이터를 전달합니다.

2. **`jq '.response | .[].start'`**: `jq`는 JSON 데이터를 처리하는 명령어입니다. 이 부분은 `pbpaste`로부터 받은 JSON 데이터 중에서 `.response` 필드를 선택하고, 그 안의 각 객체에서 `.start` 필드의 값을 추출합니다. 추출된 값들은 다시 파이프를 통해 `xargs` 명령어로 전달됩니다.

3. **`xargs -I {} bash -c "eval '$KST_FUNC'; kst \"\$@\"" _ {}`**: `xargs`는 입력 데이터를 명령어의 인자로 전달하는 역할을 합니다. `-I {}` 옵션은 입력된 데이터를 `{}`로 표시된 위치에 대체하여 명령어를 실행하게 합니다. 여기서는 `bash -c` 명령어가 실행됩니다.

    - **`bash -c "eval '$KST_FUNC'; kst \"\$@\""`**: 새로운 bash 세션을 시작하고, `-c` 옵션으로 주어진 명령어를 실행합니다. 이 명령어는 두 부분으로 구성됩니다.
        - **`eval '$KST_FUNC'`**: `eval`은 문자열을 명령어로 평가하고 실행하는 bash 내장 명령어입니다. 여기서는 `$KST_FUNC` 환경 변수에 저장된 `kst` 함수의 정의를 평가하고 실행하여, 새로운 쉘 세션에서도 `kst` 함수를 사용할 수 있게 합니다.
        - **`kst \"\$@\""`**: `kst` 함수를 호출하며, `\$@`는 모든 인자를 의미합니다. 여기서는 `xargs`로부터 전달받은 값을 인자로 사용합니다.

4. **`_ {}`**: `bash -c`에 전달되는 인자들입니다. `_`는 `bash -c`의 첫 번째 인자로, 일반적으로 스크립트 이름이 위치하지만 여기서는 특별한 의미가 없습니다. `{}`는 `xargs`로부터 전달받은 값을 대체하는 부분으로, `jq`로부터 추출된 `.start` 필드의 값이 여기에 들어갑니다.

### 동작 원리

이 명령어의 핵심은 복잡한 `kst` 함수를 다른 쉘 세션에서도 사용할 수 있게 하는 것입니다. 이를 위해 함수 정의를 환경 변수 `$KST_FUNC`에 저장하고, `eval`을 사용하여 새로운 쉘 세션에서 이 환경 변수를 평가함으로써 함수를 재정의합니다. 이렇게 하면, `xargs`를 통해 전달된 값들을 `kst` 함수의 인자로 사용하여 함수를 호출할 수 있습니다.

이 과정은 복잡한 함수를 다양한 데이터에 대해 반복적으로 사용하고자 할 때 유용합니다. 하지만, 함수 정의를 환경 변수로 전달하는 방식은 보안 측면에서 주의가 필요합니다. 왜냐하면 악의적인 코드가 포함된 함수 정의가 실행될 위험이 있기 때문입니다.

## 특정 길이마다 줄을 나누어 출력

긴 명령어를 `xargs`를 사용하여 특정 길이마다 줄을 나누어 출력하려면, `xargs`와 `printf`를 조합하여 사용할 수 있습니다. 아래는 `xargs`를 사용하여 긴 명령어를 줄바꿈하여 출력하는 방법입니다.

### 방법: `xargs`와 `printf` 조합

```bash
echo arg1 arg2 arg3 arg4 arg5 arg6 arg7 arg8 arg9 arg10 arg11 \
    | xargs -n 5 \
    | sed 's/$/ \\/' \
    | sed '$ s/ \\$//'
```

이 명령어는 다음과 같은 단계를 거칩니다:

1. `echo`로 긴 인자 목록을 출력합니다.
2. `xargs -n 5`로 인자를 5개씩 나누어 각 줄에 배치합니다.
3. `sed 's/$/ \\/'`로 각 줄 끝에 `\`를 추가합니다.
4. 마지막 줄에서 `\`를 제거하기 위해 `sed '$ s/ \\$//'`를 사용합니다.
    - `$`: `sed`에서 줄의 끝을 나타내며, 마지막 줄을 의미합니다.
    - `s/`: `sed`의 `s` 명령은 치환(substitute)을 의미합니다.
    - `\\$`: 이 부분은 줄 끝의 백슬래시(`\`) 문자를 찾는 패턴입니다. sed에서 `$`는 줄 끝을 의미하기 때문에, 이를 문자로 사용하려면 이스케이프(`\`) 해야 합니다.
    - `//`: 치환할 내용을 빈 문자열로 지정합니다. 즉, 찾은 패턴을 빈 문자열로 치환하여 제거합니다.

### 스크립트로 작성

이 작업을 스크립트로 작성하여 재사용할 수 있습니다. 다음은 `wrap_cli.sh`라는 스크립트 예제입니다:

```bash
#!/bin/bash

# Usage: ./wrap_cli.sh <command> <args...>
command=$1
shift
args=("$@")

echo "${args[@]}" | xargs -n 5 | sed 's/^/'"$command "'/;s/$/ \\/' | sed '$ s/ \\$//'
```

스크립트를 저장한 후 실행 권한을 부여합니다:

```bash
chmod +x wrap_cli.sh
```

사용 예:

```bash
./wrap_cli.sh ./cli arg1 arg2 arg3 arg4 arg5 arg6 arg7 arg8 arg9 arg10 arg11
```

이 스크립트를 실행하면 다음과 같은 출력이 생성됩니다:

```plaintext
./cli arg1 arg2 arg3 arg4 arg5 \
./cli arg6 arg7 arg8 arg9 arg10 \
./cli arg11
```

### 한 줄로 해결하는 방법

한 줄 명령으로 처리하려면 아래와 같이 할 수 있습니다:

```bash
echo arg1 arg2 arg3 arg4 arg5 arg6 arg7 arg8 arg9 arg10 arg11 | xargs -n 5 | sed 's/^/./cli /; s/$/ \\/' | sed '$ s/ \\$//'
```

이 명령어는 다음과 같은 출력이 생성됩니다:

```plaintext
./cli arg1 arg2 arg3 arg4 arg5 \
./cli arg6 arg7 arg8 arg9 arg10 \
./cli arg11
```

위의 방법들을 통해 긴 명령어를 줄을 나누어 깔끔하게 출력할 수 있습니다.

### 두번째 줄부터 들여쓰기

```bash
echo 'arg1 arg2 arg3 arg4 arg5 arg6 arg7 arg8 arg9 arg10 arg11' \
    | xargs -n 3 \
    | sed 's/$/ \\/' \
    | sed '$ s/ \\$//' \
    | awk 'NR==1 {print $0} NR>1 {print "    " $0}'
```

- `awk 'NR==1 {print $0} NR>1 {print "    " $0}'`

    첫 번째 줄은 그대로 출력하고, 두 번째 줄부터는 네 칸 띄워쓰기를 추가합니다.

```bash
arg1 arg2 arg3 \
    arg4 arg5 arg6 \
    arg7 arg8 arg9 \
    arg10 arg11
```
