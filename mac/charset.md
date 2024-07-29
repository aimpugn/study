# charset

- [charset](#charset)
    - [charset 확인하기](#charset-확인하기)

## charset 확인하기

```shell
# encoding detector library
brew install uchardet

# proto 파일들은 제외
# file -I로 보면 utf-8로 나오는 것들도 있어서 다시 한번 필터링
find ./ -name "*.php"  ! -path './/app/Lib/Proto/*' ! -path './/app/Vendor/google/protobuf/*' \\
    | xargs uchardet | rg -v 'ASCII|UTF-8' | awk '{print $1}' \\
    | sed 's|\\.//||; s|:$||' \\
    | xargs file -I | rg -v 'utf-8|binary'
```
