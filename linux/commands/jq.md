# jq

- [jq](#jq)
    - [설치](#설치)
    - [튜토리얼](#튜토리얼)
        - [키로 접근하기](#키로-접근하기)
        - [`?` 마크](#-마크)
        - [이터레이터](#이터레이터)
    - [Named Group](#named-group)
        - [`(?<name>pattern)`의 역할](#namepattern의-역할)
        - [사용례](#사용례)
    - [functions](#functions)
        - [`gsub(regex; tostring)`, `gsub(regex; tostring; flags)`](#gsubregex-tostring-gsubregex-tostring-flags)
    - [Options](#options)
        - [`-n`, `--null-input`](#-n---null-input)
        - [`-r`, `--raw-output`](#-r---raw-output)
        - [`-R`, `--raw-input`](#-r---raw-input)
        - [`-s`, `--slurp`](#-s---slurp)
        - [`slurp`의 뜻과 유래](#slurp의-뜻과-유래)
            - [뜻](#뜻)
            - [유래](#유래)
        - [프로그래밍에서의 `slurp`](#프로그래밍에서의-slurp)
            - [프로그래밍에서의 의미](#프로그래밍에서의-의미)
        - [`jq`에서 `--slurp` 옵션의 사용](#jq에서---slurp-옵션의-사용)
            - [`--slurp` 옵션의 의미](#--slurp-옵션의-의미)
            - [사용 예시](#사용-예시)
        - [왜 `slurp`라는 용어가 사용되었는지](#왜-slurp라는-용어가-사용되었는지)
        - [예시를 통한 설명](#예시를-통한-설명)
            - [예제 1: 여러 줄의 입력을 배열로 처리](#예제-1-여러-줄의-입력을-배열로-처리)
            - [예제 2: JSON 문자열을 배열로 읽어 처리](#예제-2-json-문자열을-배열로-읽어-처리)
        - [요약](#요약)

## [설치](https://stedolan.github.io/jq/download/)

```shell
brew install jq
```

## 튜토리얼

```shell
JSON=$(cat <<EOF
{
    "key":"value",
    "arr":[1, 2, 3],
    "obj":{
        "k2":"val2",
        "k3":"val3",
        "k4":["4","5","6"]
    }
}
EOF
)
```

### 키로 접근하기

`.`: 입력을 받아서 그대로 출력하는 식별자

```shell
echo $JSON | jq '.<KEY>'
```

특수문자 없는 문자열은 `.` 사용해서 바로 키 접근 가능

```shell
❯ echo $JSON | jq '.key'
"value"

❯ echo $JSON | jq '.obj'
{
  "k2": "val2",
  "k3": "val3",
  "k4": [
    "4",
    "5",
    "6"
  ]
}
```

대괄호 사용해서 키 접근 가능

```shell
❯ echo $JSON | jq '.["obj"]'
{
  "k2": "val2",
  "k3": "val3",
  "k4": [
    "4",
    "5",
    "6"
  ]
}
```

### `?` 마크

`.` 배열 또는 오브젝트가 아닌데 접근 시 에러 출력 안함

```shell
❯ echo $JSON | jq '.key.key'
jq: error (at <stdin>:9): Cannot index string with string "key"

❯ echo $JSON | jq '.key.key?'
# 어떤 에러 출력도 없음
```

### 이터레이터

`.[]` 사용하면 배열의 모든 요소를 반환

## Named Group

`jq`에서 사용하는 `(?<name>pattern)`과 같은 구문은 정규 표현식에서 "명명된 그룹" 또는 "명명된 캡처 그룹"을 나타냅니다.
정규 표현식 내에서 특정 패턴과 일치하는 부분을 캡처하여, 이를 `name`이라는 이름으로 참조하거나 변형할 수 있게 합니다.

### `(?<name>pattern)`의 역할

`(?<name>pattern)` 형태로 사용되는 명명된 그룹은 정규 표현식에서 일치하는 부분을 캡처하고, 그 결과를 `name`이라는 이름의 변수에 저장합니다.
이 변수는 이후 `jq`의 문자열 보간(`\(.name)`)을 통해 참조할 수 있습니다.
이는 일반적인 정규 표현식 그룹 `()`와 비슷하지만, 명명된 그룹은 이름을 부여함으로써 나중에 더 직관적으로 참조할 수 있는 장점이 있습니다.

### 사용례

명명된 그룹을 사용하면 `gsub` 함수에서 특정 패턴과 일치하는 문자열을 보다 쉽게 대체할 수 있습니다. 예를 들어:

```bash
echo '"abc"' | jq 'gsub("(?<x>[a-z])"; "\(.x):")'
```

위 코드에서 `"(?<x>[a-z])"`는 `[a-z]`와 일치하는 모든 소문자 알파벳을 `x`라는 이름으로 캡처합니다.
이후 `"\(.x):"`는 이 캡처된 값을 참조하여 `:`과 함께 출력하게 됩니다. 결과적으로 `"a:b:c:"`라는 결과가 출력됩니다.

`.x`는 `jq`에서 `capture`된 값이나 명명된 그룹을 참조하는 구문입니다.

예를 들어, `\(.x)`는 `x`로 명명된 그룹의 값을 참조하고, 이를 보간하여 문자열 내에서 해당 값을 삽입합니다.

```bash
jq 'gsub("(?<x>.)[^a]*"; "+\(.x)-")' <<< '"Abcabc"'
```

이 예시에서 `"(?<x>.)[^a]*"`는 첫 번째 문자(`.`)를 `x`로 캡처한 후, `[^a]*`에 일치하는 문자를 무시합니다.
`\(.x)`는 이 `x`를 참조하여 해당 위치에 대체 문자열을 삽입합니다.
결과적으로 `"Abcabc"`에서 각 일치하는 문자가 `"+A-"`, `"+a-"`로 대체됩니다.

## functions

### `gsub(regex; tostring)`, `gsub(regex; tostring; flags)`

> `gsub` is like sub but all the non-overlapping occurrences of the regex are replaced by `tostring`, after interpolation.
> If the second argument is a stream of `jq` strings, then gsub will produce a corresponding stream of JSON strings.

`gsub`는 `sub`와 유사하지만, 정규 표현식의 모든 겹치지 않는(non-overlapping) 발생을 `tostring`으로 대체합니다.
이때 대체는 보간(interpolation)을 거쳐 수행됩니다.

> **Non-overlapping**?
>
> "겹치지 않는"을 의미합니다.
> 즉, 하나의 대체가 완료된 후 그 대체된 부분은 다시 대체 과정에 포함되지 않음을 뜻합니다.
>
> 예를 들어, 텍스트가 `"aaa"`이고, 정규 표현식이 `"aa"`일 경우, `gsub("aa", "b")`를 적용하면 `"ba"`가 됩니다.
> 첫 번째 `"aa"`가 `"b"`로 대체되었으므로 `"aa"` 중복된 부분은 다시 대체되지 않습니다.
>
> **Non-overlapping Occurrences**?
>
> 정규 표현식과 일치하는, 서로 겹치지 않는 모든 발생을 의미합니다.
> 위의 예시에서 `"aaa"` 중 첫 번째 `"aa"`가 `"b"`로 대체된 후, 남은 `"a"`는 `"aa"`와 일치하지 않으므로 대체되지 않습니다.
>
> **interpolation**?
>
> 보간(interpolation)은 문자열 내부의 변수나 표현식을 평가하여 해당 값으로 치환하는 과정을 의미합니다.
> jq에서 문자열 내부에 `${}` 구문으로 표현된 부분을 변수나 표현식으로 해석하고, 이를 그 값으로 대체합니다.

만약 두 번째 인자가 jq 문자열의 스트림이라면, `gsub`은 해당 스트림에 상응하는 JSON 문자열의 스트림을 생성합니다.

`gsub` 함수는 텍스트 내의 특정 패턴을 찾아 이를 다른 문자열로 대체하는 역할을 합니다.
이 함수는 정규 표현식과 일치하는 모든 부분을 찾아 대체하지만, 중복으로 겹치는 부분은 제외하고 대체를 진행합니다.
이 과정에서 보간(interpolation)을 사용하여 문자열을 변환한 후 대체를 수행합니다.

```sh
❯ echo '"abc"' | jq 'gsub("([a-z])"; "\(.):")'
"{}:{}:{}:"

❯ echo '"abc"' | jq 'gsub("((?<x>[a-z]))"; "\(.x):")'
"a:b:c:"
```

`"abc"`로 출력해야 JSON 문자열로 해석됩니다.
이 명령어는 정상적으로 동작하여 각 문자 뒤에 `:`을 붙인 문자열 "a:b:c:"를 생성해야 하지만, 실제 출력된 결과는 "{}:{}:{}:"였습니다.
이 문제는 정규 표현식에서 ('[a-z]')가 jq에서 정규 표현식 그룹이 아닌, 오브젝트에 대한 포맷팅으로 잘못 해석된 것일 수 있습니다.

```sh
❯ echo "abc" | jq 'gsub("([a-z])"; "\(.):")'
jq: parse error: Invalid numeric literal at line 2, column 0
```

이 경우, jq는 입력된 "abc" 문자열을 JSON 형식으로 해석하려고 시도합니다.
하지만 "abc"는 유효한 JSON이 아닙니다.
JSON 문자열은 반드시 큰따옴표로 둘러싸여 있어야 하며, 여기서는 단순 문자열 "abc"가 아닌 JSON 문자열 '"abc"'가 필요합니다.
이로 인해 jq는 JSON 구문 분석 오류를 발생시킵니다.

```sh
❯ jq 'gsub("(?<x>.)[^a]*"; "+\(.x)-")' <<< '"Abcabc"'
"+A-+a-"

❯ jq '[gsub("p"; "a", "b")]' <<< '"p"'
[
  "a",
  "b"
]
```

## Options

### `-n`, `--null-input`

> use `null` as the single input value

### `-r`, `--raw-output`

### `-R`, `--raw-input`

> read each line as string instead of JSON;

`--raw-input` 옵션을 사용하여 JSON 형식이 아닌 원시 입력 데이터를 받아 처리할 수 있습니다.
즉, `--raw-input`은 입력을 JSON 문자열로 읽습니다.

```sh
echo '{"key":"val"}' | jq --raw-input | jq '.key'
# jq: error (at <stdin>:1): Cannot index string with string "key"
```

이 오류 메시지는 `jq`가 문자열 입력을 JSON 객체로 변환하는 과정에서 발생합니다.
`jq --raw-input`을 사용하면 입력이 그대로 문자열로 처리되기 때문에, `fromjson` 필터를 사용하여 JSON 객체로 변환한 후에 키를 참조해야 합니다.
즉, 원시 입력을 읽고 JSON으로 변환하여 키에 접근해야 합니다.

```sh
jq -Rn 'input | fromjson | .key'
```

- `-R`, `--raw-input`: 입력을 문자열로 읽습니다.
- `-n`: 입력을 무시하고 명령어에서만 동작합니다.
- `input`: `jq`의 빌트인 함수로, 입력을 가져옵니다.
- `fromjson`: 문자열을 JSON 객체로 변환합니다.
- `.key`: 변환된 JSON 객체에서 `key` 값을 추출합니다.

### `-s`, `--slurp`

```bash
echo -e "line 1\nline 2\nline 3" | jq -sR 'split("\n")[:-1]'
# 출력:
# [
#   "line 1",
#   "line 2",
#   "line 3"
# ]
```

- `-s`, `--slurp`: 여러 줄을 읽어 하나의 배열로 만듭니다.
- `-R`, `--raw-input`: 원시 입력을 문자열로 처리합니다.
- `split("\n")[:-1]`: 입력 문자열을 개행 문자로 나누고 마지막 빈 요소를 제거합니다.

```bash
❯ echo -e '{"key":"val1"}\n{"key":"val2"}' | jq -sR 'split("\n")'

[
  "{\"key\":\"val1\"}",
  "{\"key\":\"val2\"}",
  ""
]
```

여기서 단순히 `map` 하면 빈 문자열을 map 하지 못합니다.

```bash
❯ echo -e '{"key":"val1"}\n{"key":"val2"}' | jq -sR 'split("\n") | map(fromjson)'

jq: error (at <stdin>:2): Expected JSON value (while parsing '')
```

문자열 길이가 빈 공백이 아닌 경우 조건으로 줘서 `fromjson`을 하면 정상적으로 동작합니다.

```bash
❯ echo -e '{"key":"val1"}\n{"key":"val2"}' | jq -sR 'split("\n") | map(select(length > 0) | fromjson)'

[
  {
    "key": "val1"
  },
  {
    "key": "val2"
  }
]
```

---

### `slurp`의 뜻과 유래

#### 뜻

`slurp`는 영어 단어로 "후루룩 마시다" 또는 "후루룩 먹다"는 뜻입니다. 이 단어는 액체나 음식물을 큰 소리와 함께 빨아들이는 행위를 묘사할 때 사용됩니다.

#### 유래

`slurp`는 주로 소리를 나타내는 의성어로, 액체를 빠르게 빨아들이는 소리를 표현합니다. 이 단어는 여러 프로그래밍 언어와 도구에서 많은 데이터를 한 번에 읽어들이는 기능을 설명할 때 사용됩니다.

### 프로그래밍에서의 `slurp`

#### 프로그래밍에서의 의미

프로그래밍에서 `slurp`는 파일이나 입력 스트림의 모든 내용을 한 번에 읽어들이는 작업을 의미합니다. 이와 반대되는 작업은 데이터를 한 줄씩 또는 한 블록씩 읽어들이는 것입니다.

### `jq`에서 `--slurp` 옵션의 사용

#### `--slurp` 옵션의 의미

`jq`에서 `--slurp` (`-s`) 옵션은 입력된 모든 내용을 한 번에 읽어서 하나의 JSON 배열로 만드는 기능을 제공합니다. 이 옵션을 사용하면 여러 줄의 입력을 한 번에 처리할 수 있습니다.

#### 사용 예시

`--slurp` 옵션을 사용하면 여러 줄의 입력을 하나의 배열로 읽어들여, 각 줄을 배열의 요소로 취급합니다.

### 왜 `slurp`라는 용어가 사용되었는지

`slurp`라는 용어는 입력된 데이터를 마치 후루룩 빨아들이듯이 한 번에 읽어들이는 동작을 표현하는 데 적합합니다. 이는 데이터를 빠르고 효율적으로 처리하기 위한 방법을 직관적으로 나타냅니다.

### 예시를 통한 설명

#### 예제 1: 여러 줄의 입력을 배열로 처리

```sh
echo -e "line 1\nline 2\nline 3" | jq -s '.'
```

이 명령어는 각 줄을 배열의 요소로 만들어 줍니다.

출력:

```json
[
  "line 1",
  "line 2",
  "line 3"
]
```

#### 예제 2: JSON 문자열을 배열로 읽어 처리

```sh
echo '{"key":"val"}' | jq -sR 'fromjson | .key'
```

이 명령어는 JSON 문자열을 읽어들여 JSON 객체로 변환한 후 키 값을 추출합니다.

출력:

```json
val
```

### 요약

- **`slurp`의 뜻**: "후루룩 마시다" 또는 "후루룩 먹다"는 뜻의 영어 단어.
- **프로그래밍에서의 의미**: 파일이나 입력 스트림의 모든 내용을 한 번에 읽어들이는 작업을 의미.
- **`jq`에서의 사용**: 입력된 모든 내용을 한 번에 읽어서 하나의 JSON 배열로 만드는 기능.
- **유래와 이유**: 데이터를 빠르고 효율적으로 한 번에 처리하는 동작을 직관적으로 나타내기 위해 사용됨.
