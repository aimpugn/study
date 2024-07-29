# array

- [array](#array)
    - [일반 배열 선언 및 초기화](#일반-배열-선언-및-초기화)
    - [배열에 요소 추가](#배열에-요소-추가)
    - [배열 요소 접근](#배열-요소-접근)
    - [배열의 길이 확인](#배열의-길이-확인)
    - [연관 배열 선언 및 초기화](#연관-배열-선언-및-초기화)
    - [`array[@]`와 `array[*]`](#array와-array)
        - [`array[@]`와 `array[*]`의 차이점](#array와-array의-차이점)
        - [배열 요소를 문자열로 결합](#배열-요소를-문자열로-결합)
    - [배열인지 확인하는 방법](#배열인지-확인하는-방법)
    - [배열 요소를 유니크하게 만들기](#배열-요소를-유니크하게-만들기)
        - [1. `sort`와 `uniq`를 사용하는 방법](#1-sort와-uniq를-사용하는-방법)
        - [2. `awk`를 사용하는 방법](#2-awk를-사용하는-방법)
        - [3. 연관 배열을 사용하는 방법](#3-연관-배열을-사용하는-방법)
    - [한 배열의 요소가 다른 배열에 포함되어 있는지 확인하는 방법](#한-배열의-요소가-다른-배열에-포함되어-있는지-확인하는-방법)

## 일반 배열 선언 및 초기화

```bash
# 배열 선언 및 초기화
array=()
```

## 배열에 요소 추가

```bash
# 배열에 요소 추가
array+=("file1")
array+=("file2")
```

## 배열 요소 접근

```bash
# 배열 요소 접근
echo "${array[@]}"  # 모든 요소
echo "${array[0]}"  # 첫 번째 요소
```

## 배열의 길이 확인

```bash
# 배열 길이 확인
echo "${#array[@]}"
```

## 연관 배열 선언 및 초기화

```bash
#!/bin/bash

# https://stackoverflow.com/a/10806809
# local associative array
f() {
    declare -A map
    map[x]=a
    map[y]=b
}

# global associative array
# bash 4.2 adds "declare -g" to create global variables from within a function.
```

## `array[@]`와 `array[*]`

쉘 스크립트에서 `array[@]`와 `array[*]`는 배열의 모든 요소를 참조할 때 사용됩니다. 이 둘은 비슷해 보이지만, 실제로는 다르게 작동합니다. 각각의 차이점을 예제를 통해 자세히 설명하겠습니다.

`array[@]`는 배열의 각 요소를 개별적으로 처리할 때 유용하며,
`array[*]`는 배열의 모든 요소를 하나의 문자열로 결합할 때 유용합니다.

### `array[@]`와 `array[*]`의 차이점

- `${array[@]}`

    배열의 모든 요소를 개별적으로 확장합니다.

- `${array[*]}`: 배열의 모든 요소를 하나의 문자열로 확장합니다.

```bash
#!/bin/bash

# `my_array` 배열을 선언하고, "apple", "banana", "cherry" 세 개의 요소를 추가합니다.
my_array=("apple" "banana" "cherry")

# for 루프에서 `"${my_array[@]}"`를 사용하면 배열의 각 요소가 개별적으로 확장됩니다.
# 따라서 각 요소가 별도의 줄에 출력됩니다.
echo "Using \${array[@]}:"
for element in "${my_array[@]}"; do
    echo "$element"
done

# for 루프에서 `"${my_array[*]}"`를 사용하면 배열의 모든 요소가 하나의 문자열로 확장됩니다.
# 따라서 모든 요소가 하나의 줄에 공백으로 구분되어 출력됩니다.
echo "Using \${array[*]}:"
for element in "${my_array[*]}"; do
    echo "$element"
done
```

```bash
Using ${array[@]}:
apple
banana
cherry
Using ${array[*]}:
apple banana cherry
```

### 배열 요소를 문자열로 결합

다음은 `array[@]`와 `array[*]`를 사용하여 배열의 요소를 문자열로 결합하는 예제입니다.

```bash
#!/bin/bash

# 배열 선언
my_array=("apple" "banana" "cherry")

# 사용하면 배열의 각 요소가 개별적으로 확장되지만, 
# `echo` 명령어는 이를 공백으로 구분하여 출력합니다.
echo "Using \${array[@]}:"
echo "${my_array[@]}"

# 배열의 모든 요소가 하나의 문자열로 확장되며, 
# `echo` 명령어는 이를 그대로 출력합니다.
echo "Using \${array[*]}:"
echo "${my_array[*]}"
```

```bash
Using ${array[@]}:
apple banana cherry
Using ${array[*]}:
apple banana cherry
```

## 배열인지 확인하는 방법

Shell script에서 변수가 배열인지 아닌지를 확인하는 방법은 `declare -p` 명령어를 사용하여 변수가 배열인지 아닌지를 검사할 수 있습니다.
`declare -p`는 변수의 선언 상태를 출력해주는데, 배열인 경우에는 배열임을 나타내는 형식으로 출력됩니다.

다음은 변수가 배열인지 확인하는 스크립트 예제입니다:

```bash
is_array() {
    local var="$1"
    if [[ "$(declare -p "$var" 2>/dev/null)" =~ "declare -a" ]]; then
        return 0
    else
        return 1
    fi
}

# 테스트 배열
arr=("x" "y" "z")
var="string"

# 배열인지 확인
if is_array "arr"; then
    echo "arr is an array"
else
    echo "arr is not an array"
fi

# 배열인지 확인
if is_array "var"; then
    echo "var is an array"
else
    echo "var is not an array"
fi
```

1. **is_array 함수**:
    - `declare -p "$var" 2>/dev/null`: 변수가 존재하지 않는 경우 오류 출력을 무시합니다.
    - `[[ "$(declare -p "$var" 2>/dev/null)" =~ "declare -a" ]]`: `declare -p`의 출력이 배열 선언을 포함하는지 확인합니다.

    ```bash
    is_array() {
        local var="$1"
        if [[ "$(declare -p "$var" 2>/dev/null)" =~ "declare -a" ]]; then
            return 0
        else
            return 1
        fi
    }
    ```

2. **테스트 배열과 문자열 변수**:

    ```bash
    arr=("x" "y" "z")
    var="string"
    ```

3. **배열인지 확인**:
    - `if is_array "arr"; then ...`: 배열인 경우 "arr is an array"를 출력합니다.
    - `if is_array "var"; then ...`: 배열이 아닌 경우 "var is not an array"를 출력합니다.

    ```bash
    if is_array "arr"; then
        echo "arr is an array"
    else
        echo "arr is not an array"
    fi

    if is_array "var"; then
        echo "var is an array"
    else
        echo "var is not an array"
    fi
    ```

이 스크립트를 실행하면 다음과 같은 출력이 생성됩니다:

```plaintext
arr is an array
var is not an array
```

이 방법을 사용하면 Shell script에서 변수가 배열인지 아닌지 쉽게 확인할 수 있습니다.

## 배열 요소를 유니크하게 만들기

Shell 스�립트에서 배열의 중복된 요소를 제거하고 유일한 요소만 남기는 방법은 여러 가지가 있습니다. 여기서는 몇 가지 방법을 소개하겠습니다.

### 1. `sort`와 `uniq`를 사용하는 방법

이 방법은 배열의 요소들을 정렬한 후 `uniq` 명령어를 사용하여 중복된 요소를 제거합니다. 이 방법은 배열의 순서를 변경할 수 있습니다.

```bash
#!/bin/bash
arr=("x" "y" "z" "x" "y" "z")

# 배열의 요소들을 한 줄로 변환하고 정렬한 후 중복 제거
unique_arr=($(printf "%s\n" "${arr[@]}" | sort | uniq))

# 결과 출력
echo "${unique_arr[@]}"
```

### 2. `awk`를 사용하는 방법

`awk`를 사용하여 배열의 중복된 요소를 제거할 수 있습니다. 이 방법은 배열의 순서를 유지합니다.

```bash
#!/bin/bash
arr=("x" "y" "z" "x" "y" "z")

# 배열의 요소들을 한 줄로 변환하고 awk를 사용하여 중복 제거
unique_arr=($(printf "%s\n" "${arr[@]}" | awk '!seen[$0]++'))

# 결과 출력
echo "${unique_arr[@]}"
```

### 3. 연관 배열을 사용하는 방법

연관 배열을 사용하여 배열의 중복된 요소를 제거할 수 있습니다. 이 방법은 배열의 순서를 유지합니다.

```bash
#!/bin/bash
arr=("x" "y" "z" "x" "y" "z")

# 연관 배열을 사용하여 중복 제거
declare -A temp_map
for elem in "${arr[@]}"; do
    temp_map["$elem"]=1
done

# 결과 출력
echo "${temp_map[@]}"
```

이 방법들은 배열의 중복된 요소를 제거하고 유일한 요소만 남기는 데 사용할 수 있습니다.
각 방법은 배열의 순서를 유지하는지 여부, 또는 외부 명령어를 사용하는지 여부에 따라 선택할 수 있습니다.

## 한 배열의 요소가 다른 배열에 포함되어 있는지 확인하는 방법

```bash
#!/bin/bash

# array1의 요소들이 array2의 배열에 포함되어 있는지 여부 확인하는 방법
array1=("a" "b" "c" "d")
array2=("b" "c" "d" "e")

# array1의 모든 요소가 array2에 포함되어 있는지 확인
all_included=true
for elem in "${array1[@]}"; do
    # array2의 모든 요소가 공백으로 구분된 문자열로 변환된 후, 
    # 현재 array1의 요소(elem)가 포함되어 있는지 확인합니다. 
    # `=~` 연산자는 정규 표현식을 사용하여 문자열 매칭을 수행합니다. 
    # 만약 elem이 array2에 포함되어 있지 않다면, 조건이 참이 되어 다음 줄로 이동합니다.
    if! [[ " ${array2[*]} " =~ " $elem " ]]; then
        all_included=false
        break
    fi
done

if $all_included; then
    echo "array1의 모든 요소가 array2에 포함되어 있습니다."
else
    echo "array1의 모든 요소가 array2에 포함되어 있지 않습니다."
fi
```
