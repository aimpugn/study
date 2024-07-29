# I/O redirection

- [I/O redirection](#io-redirection)
    - [입출력 리디렉션](#입출력-리디렉션)
    - [입력](#입력)
        - [**`<` (입력 리디렉션)**](#-입력-리디렉션)
        - [**`<<` (Here Document)**](#-here-document)
        - [**`<<<` (Here String)**](#-here-string)
    - [출력](#출력)
        - [**`>` (출력 리디렉션)**](#-출력-리디렉션)
        - [**`>>` (출력 리디렉션 추가 모드)**](#-출력-리디렉션-추가-모드)
    - [에러](#에러)
        - [표준 에러 리디렉션 (`2>`, `2>>`)](#표준-에러-리디렉션-2-2)
    - [응용](#응용)
        - [**`<<EOF >` (히어 도큐먼트와 출력 리디렉션 조합)**](#eof--히어-도큐먼트와-출력-리디렉션-조합)
        - [**`<<EOF >>` (히어 도큐먼트와 추가 모드 출력 리디렉션 조합)**](#eof--히어-도큐먼트와-추가-모드-출력-리디렉션-조합)
    - [`<(command)` 프로세스 치환과 복합 리디렉션 (`< <`)](#command-프로세스-치환과-복합-리디렉션--)
        - [`<`와 `<()` 비교](#와--비교)
    - [기타](#기타)
        - [파이프라인 (`|`)](#파이프라인-)
        - [인라인 입력 데이터?](#인라인-입력-데이터)
- [쉘 리디렉션 (Shell Redirection)](#쉘-리디렉션-shell-redirection)
    - [Executive Summary](#executive-summary)
    - [Fundamental Definition and Concept](#fundamental-definition-and-concept)
        - [리디렉션 (Redirection)](#리디렉션-redirection)
        - [파일 디스크립터 (File Descriptors)](#파일-디스크립터-file-descriptors)
        - [Glossary of Key Terms](#glossary-of-key-terms)
    - [Theoretical Framework](#theoretical-framework)
        - [기본 리디렉션 연산자](#기본-리디렉션-연산자)
        - [고급 리디렉션 연산자](#고급-리디렉션-연산자)
        - [`2>&1`의 작동 원리](#21의-작동-원리)
    - [Practical Implementation](#practical-implementation)
    - [Comparative Analysis](#comparative-analysis)
    - [Advanced Topics and Current Research](#advanced-topics-and-current-research)
    - [Further Learning Resources](#further-learning-resources)
    - [References](#references)

## 입출력 리디렉션

Bash와 다른 셸 스크립팅 언어에서 `<`, `>`, `<<`, `>>`, `<<<`, 및 파이프(`|`)와 같은 기호들은 리디렉션과 파이프라인을 다루는 데 사용된다.
이러한 기호들은 입력과 출력의 흐름을 조작하고, 파일이나 다른 프로그램과의 데이터 교환을 가능하게 한다.

## 입력

### **`<` (입력 리디렉션)**

`<` 기호는 파일의 내용을 커맨드의 입력으로 리디렉션 합니다.
이는 커맨드가 표준 입력으로부터 데이터를 읽는 대신 지정된 파일로부터 데이터를 읽게 합니다.
이때 `<`는 파일이나 파일 디스크립터를 기대합니다.

```sh
# `file.txt`의 내용을 `command`의 입력으로 리디렉션
command < file.txt
```

아래 예제에서 `sort` 커맨드는 `input.txt` 파일의 내용을 입력으로 받아 정렬된 결과를 표준 출력으로 보냅니다.

```bash
sort < input.txt
```

### **`<<` (Here Document)**

`<<` 기호는 앞서 설명한 heredoc을 시작하는 데 사용됩니다.
이는 여러 줄의 입력을 직접 스크립트나 커맨드에 제공할 수 있게 해줍니다.

```bash
cat << EOF
This is a
여러 줄의
text
EOF
```

입력 리디렉션에서 사용되며, 마커를 사용하여 인라인 입력 데이터를 명령어로 리디렉션합니다.

> **마커?**
>
> 마커는 히어 도큐먼트의 시작과 끝을 표시하는 식별자를 말한다.
> 일반적으로 마커는 `EOF` (End Of File의 약자)를 사용하지만, 사용자가 원하는 어떤 단어나 문자열도 마커로 사용할 수 있다.
> 마커는 히어 도큐먼트의 내용을 명확하게 구분하기 위해 사용되며, 스크립트 내에서 마커가 나타나는 두 위치 사이에 있는 모든 텍스트가 해당 명령어로 리디렉션된다
>
>
> **인라인 입력 데이터?**
>
> 인라인 입력 데이터는 스크립트나 명령어 내에서 직접적으로 제공되는 텍스트나 데이터를 말한다.
> 히어 도큐먼트를 사용할 때, 마커 사이에 위치한 모든 텍스트가 인라인 입력 데이터로 취급된다.
> 이 데이터는 명령어의 표준 입력으로 전달되며, 파일에서 데이터를 읽어올 필요 없이 스크립트 내에서 직접 명령어에 데이터를 제공할 수 있게 한다.

```bash
command <<EOF
line 1
line 2
EOF
```

이 예에서 `EOF`는 히어 도큐먼트의 끝을 표시하는 마커입니다. `command`는 `EOF` 마커 사이에 있는 모든 텍스트를 입력으로 받습니다.

### **`<<<` (Here String)**

`<<<` 기호는 히어스트링(Here String)을 사용하여 문자열을 커맨드의 표준 입력으로 직접 전달합니다.
문자열을 커맨드의 표준 입력으로 직접 리디렉션합니다.

예: `command <<< "text"`는 문자열 "text"를 `command`의 입력으로 리디렉션합니다.

```bash
grep "hello" <<< "hello world"
```

위 예제에서 `grep` 커맨드는 "hello world"라는 문자열에서 "hello"를 검색합니다.

## 출력

### **`>` (출력 리디렉션)**

`>` 기호는 커맨드의 표준 출력을 파일로 리디렉션합니다.
이는 커맨드의 출력을 터미널이 아닌 파일에 저장하고자 할 때 사용됩니다.
기존 파일 내용은 덮어쓰게 됩니다.

```sh
# `command`의 출력을 `file.txt`로 리디렉션
command > file.txt
```

아래 예제에서 `echo` 커맨드의 출력이 `output.txt` 파일에 저장됩니다.

```bash
echo "Hello, world!" > output.txt
```

### **`>>` (출력 리디렉션 추가 모드)**

`>>` 기호는 커맨드의 표준 출력을 파일에 추가하는 모드로 리디렉션합니다.
이는 기존 파일의 내용을 유지하면서 새로운 내용을 덧붙이고자 할 때 사용됩니다.
파일이 존재하지 않는 경우 새로 생성됩니다.

```sh
# `command`의 출력을 `file.txt`의 끝에 추가
command >> file.txt
```

아래 예제에서 "Another line"이 `output.txt` 파일의 기존 내용 뒤에 추가됩니다.

```bash
echo "Another line" >> output.txt
```

## 에러

### 표준 에러 리디렉션 (`2>`, `2>>`)

- `2>`: 표준 에러를 파일로 리디렉션하여 덮어씁니다.
- `2>>`: 표준 에러를 파일로 리디렉션하여 추가합니다.

```bash
command 2> error.log
command 2>> error.log
```

## 응용

### **`<<EOF >` (히어 도큐먼트와 출력 리디렉션 조합)**

히어 도큐먼트의 내용을 파일로 리디렉션합니다.

예: `cat <<EOF > file.txt`는 `EOF` 마커 사이의 내용을 `file.txt`로 리디렉션합니다.

### **`<<EOF >>` (히어 도큐먼트와 추가 모드 출력 리디렉션 조합)**

히어 도큐먼트의 내용을 기존 파일의 끝에 추가합니다.

예: `cat <<EOF >> file.txt`는 `EOF` 마커 사이의 내용을 기존 `file.txt`의 끝에 추가합니다.

## `<(command)` 프로세스 치환과 복합 리디렉션 (`< <`)

- `<(command)` 프로세스 치환

    `<(...)` 구문은 프로세스 치환(process substitution)이라고 하며, *괄호 안의 명령어가 실행된 결과를 파일처럼 취급*하여 다른 명령어의 입력으로 사용할 수 있게 합니다.

    `$(...)` 구문과 비슷하지만, `$(...)` 구문은 명령어의 실행 결과를 문자열로 반환하는 반면에, `<(...)` 구문은 명령어의 실행 결과를 파일처럼 취급한다는 차이가 있습니다.

    주로 입력 리디렉션(`<`)과 함께 사용합니다.

    ```sh
    diff <(ls dir1) <(ls dir2)

    ❯ echo <(ls)
    /dev/fd/13

    ❯ ls -l <(echo "test")
    prw-rw----  0 rody  staff  5  9 12 00:12 /dev/fd/13

    ❯ print -r -- ${(q):-=(<<<'test')}
    /tmp/zshFwx3HB
    ```

- **`< <(command)` (프로세스 치환과 입력 리디렉션)**:

    외부 커맨드의 출력을 다른 커맨드의 입력으로 리디렉션합니다.
    파이프와 유사하지만, 명령어를 하나의 커맨드 라인에서 사용할 수 있게 해주는 더 복잡한 형태입니다.

    ```sh
    bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
    ```

    `curl` 커맨드를 통해 가져온 스크립트의 출력을 `bash` 커맨드의 입력으로 리디렉션합니다.

### `<`와 `<()` 비교

`<`는 파일과 직접 작동하는 반면,
`<()`는 명령의 출력과 함께 작동하여 임시 파일과 같은 객체(일반적으로 named pipe)를 생성합니다.

`<`는 단순한 입력 리디렉션이지만,
`<()`는 임시 file descriptor를 생성하고 명령을 실행하는 등 더 복잡합니다.

`<`는 파일에 데이터가 있을 때 사용합지만,
`<()`는 *명령의 출력을 파일처럼 사용*하려는 경우에 사용됩니다.

## 기타

### 파이프라인 (`|`)

- **파이프(`|`)**: 한 커맨드의 출력을 다른 커맨드의 입력으로 리디렉션합니다.

  예: `command1 | command2`는 `command1`의 출력을 `command2`의 입력으로 리디렉션합니다.

### 인라인 입력 데이터?

"인라인 입력 데이터"(inline input data)라는 용어는 *특정 데이터가 명령어나 스크립트 내에서 직접적으로, 즉 "인라인"으로 제공되는 경우*를 설명할 때 사용됩니다. 이는 대조적으로 "아웃라인 입력 데이터"(outline input data)라는 명확한 용어는 일반적으로 사용되지 않습니다. 그러나, 이와 반대되는 개념으로 이해할 수 있는 것은 외부 파일이나 다른 소스로부터 데이터를 가져오는 경우입니다.

- **인라인 입력 데이터**: 명령어나 스크립트 내에 직접 포함된 데이터. 히어 도큐먼트(`<<`)나 히어 스트링(`<<<`)을 사용하여 셸 스크립트 내부에 직접 정의된 데이터입니다.

- **외부(비인라인) 입력 데이터**: 파일, 데이터베이스, 인터넷 상의 리소스, 사용자 입력 등 외부 소스로부터 가져오는 데이터입니다. 이 데이터는 스크립트나 프로그램 실행 시 외부에서 읽어들여지며, `<` 리디렉션 연산자를 사용하여 명령어로 입력될 수 있습니다.

이를 요약해 보자면:

- **인라인 입력 데이터**: 스크립트나 명령어 내에 직접 포함된 데이터.
- **외부 입력 데이터**: 외부 소스(파일, 데이터베이스 등)로부터 가져오는 데이터.
- **출력 데이터**: 명령어나 스크립트의 실행 결과로 생성되며, 다양한 방법으로 리디렉션되거나 처리될 수 있습니다.

인라인과 아웃라인(또는 비인라인)의 구분은 데이터가 어디서 오는가(내부적으로 정의되었는가, 외부에서 가져오는가)와 어떻게 처리되는가(직접 포함되었는가, 리디렉션을 통해 입력되는가)에 초점을 맞춥니다.

---

리디렉션 문법에 대한 질문에 대해 상세히 답변하겠습니다. 특히 `2>&1`과 같은 복잡한 리디렉션의 의미와 작동 방식을 중심으로 설명하겠습니다.

# 쉘 리디렉션 (Shell Redirection)

## Executive Summary

- `2>&1`은 표준 에러(stderr, 파일 디스크립터 2)를 표준 출력(stdout, 파일 디스크립터 1)으로 리디렉션합니다.
- `>&`는 파일 디스크립터를 다른 파일 디스크립터로 복사하는 특별한 리디렉션 연산자입니다.
- 리디렉션은 프로세스의 입출력 스트림을 제어하는 강력한 쉘 기능입니다.

## Fundamental Definition and Concept

### 리디렉션 (Redirection)

리디렉션은 프로그램의 입력과 출력을 파일이나 다른 프로그램으로 전환하는 프로세스입니다. 유닉스 계열 운영체제에서는 모든 것이 파일로 취급되며, 이는 입출력 스트림도 포함합니다.

### 파일 디스크립터 (File Descriptors)

파일 디스크립터는 열린 파일을 나타내는 정수입니다. 기본적인 파일 디스크립터는 다음과 같습니다:
- 0: 표준 입력 (stdin)
- 1: 표준 출력 (stdout)
- 2: 표준 에러 (stderr)

### Glossary of Key Terms

- stdin: 표준 입력 스트림
- stdout: 표준 출력 스트림
- stderr: 표준 에러 스트림
- 리디렉션 연산자: `>`, `<`, `>>`, `<<`, `>&`, `<&` 등

## Theoretical Framework

### 기본 리디렉션 연산자

1. `>`: 출력 리디렉션 (덮어쓰기)
2. `>>`: 출력 리디렉션 (추가하기)
3. `<`: 입력 리디렉션

### 고급 리디렉션 연산자

1. `>&`: 파일 디스크립터 복사
2. `<&`: 파일 디스크립터로부터 입력 받기
3. `<<`: Here Document
4. `<<<`: Here String

### `2>&1`의 작동 원리

1. `2>` 부분은 표준 에러(stderr, 파일 디스크립터 2)를 리디렉션합니다.
2. `&1` 부분은 파일 디스크립터 1(stdout)을 참조합니다.
3. 전체적으로, stderr를 stdout으로 리디렉션합니다.

## Practical Implementation

```bash
# 표준 출력과 표준 에러를 모두 파일로 리디렉션
command > output.txt 2>&1

# 표준 출력은 파일로, 표준 에러는 /dev/null로 리디렉션
command > output.txt 2>/dev/null

# Here Document 사용 예
cat << EOF
Hello, World!
This is a here document.
EOF

# Here String 사용 예
grep 'pattern' <<< "This is a here string"
```

## Comparative Analysis

| 리디렉션 | 의미 | 예시 |
|----------|------|------|
| `>` | 출력 리디렉션 (덮어쓰기) | `command > file` |
| `>>` | 출력 리디렉션 (추가) | `command >> file` |
| `<` | 입력 리디렉션 | `command < file` |
| `>&` | 파일 디스크립터 복사 | `2>&1` |
| `<<` | Here Document | `cat << EOF` |
| `<<<` | Here String | `grep 'pattern' <<< "string"` |

## Advanced Topics and Current Research

1. 프로세스 치환 (Process Substitution): `<(command)` 또는 `>(command)` 형식으로 사용
2. 명명된 파이프 (Named Pipes): `mkfifo` 명령어로 생성
3. /dev/fd/ 디렉토리: 파일 디스크립터에 대한 직접 접근

## Further Learning Resources

1. Bash 매뉴얼: `man bash`에서 "REDIRECTION" 섹션 참조
2. Advanced Bash-Scripting Guide: [I/O Redirection](https://tldp.org/LDP/abs/html/io-redirection.html)
3. GNU Coreutils 매뉴얼: [Redirections](https://www.gnu.org/software/coreutils/manual/html_node/Redirections.html)

## References

1. IEEE and The Open Group. (2018). The Open Group Base Specifications Issue 7, 2018 edition. IEEE Std 1003.1-2017 (Revision of IEEE Std 1003.1-2008). <https://pubs.opengroup.org/onlinepubs/9699919799/>
2. Bash Reference Manual. (n.d.). GNU Project. <https://www.gnu.org/software/bash/manual/html_node/Redirections.html>

이 설명을 통해 쉘 리디렉션의 다양한 형태와 그 작동 원리에 대해 이해하셨기를 바랍니다. 리디렉션은 유닉스 계열 시스템에서 매우 강력하고 유용한 기능이며, 이를 잘 이해하고 활용하면 복잡한 작업을 효율적으로 수행할 수 있습니다.
