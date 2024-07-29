# Tips

## shell debugging

```shell
#!/bin/bash

# Debugging
#set -x

```

## 현재 스크립트

```shell
# workspace를 일관적으로 반환하기 위한 변수 및 함수들
# bash 경우 https://www.gnu.org/software/bash/manual/html_node/Bash-Variables.html
CURRENT_SCRIPT="${BASH_SOURCE[0]}"
if [ -z "${CURRENT_SCRIPT// }" ]; then
  # zsh 경우 https://unix.stackexchange.com/a/115431
  CURRENT_SCRIPT="${0:a}"
fi
```

## grep 관련

```shell
function is_oneof(){
  local to_check="${1// }"
  
  if echo "opt1|opt2|opt3" | grep -q "$to_check"
  then
    # do something
    printf "%s\n" "$to_check"
  fi
}
```

## arguments

### 인자 파싱 및 특정 인자 제거

```shell
  for i do
    shift
    case "$i" in
      --option1)
        # do something about option 1
        continue
        ;;
      --option2)
        # do something about option 2
        continue
        ;;
    esac
    
    set -- "$@" "$i"
  done
```

### 인자

```shell
# https://stackoverflow.com/a/31307237
# ${variable#glob}
# ㄴ `variable`: bash variable
# ㄴ `#`: deletes the shortest match
# ㄴ `glob`: glob pattern
function args_test(){
  local opt1=""
  local opt2=""
  for i in "$@"; do
    case "$i" in
      --opt1=*)
        opt1="${i#*=}"
        ;;
      --opt2=*)
        opt2="${i#*=}"
        ;;
    esac
  done
  printf "%s\n" "opt1: ${opt1}, opt2: $opt2"
}
```

### 전체 인자 출력하기

```shell
function print_all_args() {
  printf '\t%s\n' "$*"
}
```

### N번째 인자만 사용하기

```shell
function print_Nth_arg(){
  printf "%s\n" "${<SOME INDEX>// }"
}
```

## 특정 인자 이후 인자들만 사용

```shell
  args_idx=<SOME INDEX>
  args_available=$(printf "%s " "${@:$args_idx}") # https://www.shellcheck.net/wiki/SC2124
```

## 타입 알아보기

```shell
type -t some_function # function
type -t ls # alias
type -t test # builtin
```

## 문자열 정리

### 특정 문자 외에는 제거

```shell
# ex:
#   5.6.40-60+ubuntu20.04.1+deb.sury.org+1 -> 56
#   7.4.30 -> 74
php_version=$(php -r 'echo PHP_VERSION;' | grep -oe "^[5-9]\.[0-9]" | sed 's/\.//')
```

## if 조건들

### 실행 가능 여부

```shell
function php_executable() {
  if [ -x php ]
  then
    echo "'php' is not executable"
    return 255
  fi
}
```

## 반복문

## docker 관련

### 컨테이너 이름을 가져오기

```shell

# 컨테이너 이름을 가져오기 위한 함수
#
# 지원 옵션:
# --name
# --port
# --network_id
# ex:
#   container_names_by --port=33060
#   container_names_by --name=testdb
#   container_names_by --name=testdb --port=33060
#   container_names_by --name=testdb --port=33060 --network_id=<NETWORK_ID OF CONTAINER>
container_names_by() {
  local container_name=""
  local name=""
  local port=""
  local network_id=""

  for i in "$@"; do
    case "$i" in
      --name=*)
        name="${i#*=}"
        ;;
      --port=*)
        port="${i#*=}"
        ;;
      --network_id=*)
        network_id="${i#*=}"
        ;;
    esac
  done

  if [[  -z "$container_name" && -n "$port" ]]; then
    container_name=$(docker ps -a --filter "publish=$port" --format '{{.Names}}')
  elif [[ -z "$container_name" && -n "$network_id" ]]; then
    container_name=$(docker ps -a --filter "network=$network_id" --format '{{.Names}}')
  elif [[ -z "$container_name" && -n "${name// }" ]]; then
    container_name=$(docker ps -a --filter "name=$name" --format '{{.Names}}')
  fi

  echo "${container_name// }"
  return 0
}
```

### 컨테이너 이름으로 상태 가져오기

```shell
# ex:
#   status_by_container_name --name="testdb"
status_by_container_name() {
  local name=""
  local container_status="unknown"

  for i in "$@"; do
    case $i in
    --name=*)
      name="${i#*=}"
      ;;
    esac
  done

  if [ -n "${name// }" ]; then
    local Status
    Status=$(docker ps -a --filter "name=$name" --format '{{.Status}}')
    case "$Status" in
      "Up"*"(Paused)")
        container_status="paused"
        ;;
      "Up"*)
        container_status="running"
        ;;
      "Exited"*|"")
        container_status="exited"
        ;;
    esac
  fi

  echo "$container_status"

  return  0
}
```

### mysql 컨테이너 상태 확인하기

```shell
# container_names_by 및 status_by_container_name 사용하여 컨테이너 상태 확인하기
function check_mysql_container_running(){
  # container가 정상적으로 실행되고 있는지 확인
  local max_loop_cnt=30
  local loop_cnt=0
  local mysql_container_status=""
  local mysql_name="$(container_names_by --port=3306)"
  
  mysql_container_status=$(status_by_container_name --name="$mysql_name")

  while [[ "$mysql_container_status" != "running" && $loop_cnt -lt $max_loop_cnt ]]; do
    # cli로 비밀번호 입력 warning, 연결 실패 에러 등 불필요한 메시지는 /dev/null로 버린다
    mysql_container_status=$(status_by_container_name --name="$mysql_name" 2>/dev/null)
    ((loop_cnt++))
    sleep 1
  done
}
```
