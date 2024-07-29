# file

## grep 결과로 파일 만들기

```bash
#!/bin/bash

# 특정 결과로 빈 파일 생성하기
# https://stackoverflow.com/a/11909057 참고
grep "CREATE TABLE" resources/sqls/meta/tables.sql \
    | grep -Eo '\`\w+`' \
    | sed 's/\`//g' \
    | xargs -n 1 -I {} echo './resources/sqls/data/'{}'.sql' \
    | xargs touch
```

## 파일의 인코딩 정보 확인하기

```sh
file -i filename
```

`filename`을 확인하려는 파일의 이름으로 바꿉니다. 예를 들어, `example.txt` 파일의 인코딩을 확인하려면:

```sh
file -i example.txt
```

출력 예시:

```sh
example.txt: text/plain; charset=utf-8
```

여기서 `charset=utf-8`이 파일의 인코딩을 나타냅니다.
