# Parse arguments

## 인자 파싱하기

```bash
for i in "$@"; do
  case $i in
  -n=*|--network=*)
    NETWORK_ID="${i#*=}"
    ;;
  --mysql-user=*)
    MYSQL_USERNAME="${i#*=}"
    ;;
  --mysql-password=*)
    MYSQL_PASSWORD="${i#*=}"
    ;;
  esac
done
```

## 인자 목록에서 특정 인자 제거하기

```bash
# 쉘 스크립트 인자 목록에서 특정 인자 제거하기
# https://unix.stackexchange.com/questions/403890/how-to-drop-an-argument-from-the-list-of-arguments-in-a-shell-script
# https://unix.stackexchange.com/questions/569510/remove-argument-from-if-it-is-an-option
for a; do
  shift
  case $a in
    -*) opts+=("$a");;
    # set -- to update the params
    *) set -- "$@" "$a";;
  esac
done

for i do
  [ "$i" = -D ] || set -- "$@" "$i"
  shift
done
printf '<%s>\n' "$@"
```

## `@`와 `*`의 차이

```bash
# `@`와 `*`의 차이
# https://stackoverflow.com/a/21071995
echo "\${@:2}:" "${@:2}"
echo "\$*: $*"
# https://github.com/koalaman/shellcheck/issues/214#issuecomment-52276612
echo "\${*:2}: ${*:2}"
```
