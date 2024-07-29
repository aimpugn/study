# strings

## 특정 문자열에 해당하는 내용만 출력

```shell
#!/bin/bash
# 특정 문자열에 해당하는 내용만 출력
# van_transactions 만 추출된다. https://stackoverflow.com/a/11909057
echo 'CREATE TABLE IF NOT EXISTS`van_transactions` (' | grep -Eo '\`\w+`' | sed 's/\`//g'
```

## 문자열 치환

```shell
# 문자열 치환
service_tmp=${deploy_dir//"$deploy_parent_dir\/"/}
# == "$(echo "$deploy_dir" | sed "s#$deploy_parent_dir/##g")"

# 슬래쉬로 자르고, 마지막 필드 가져오기
# awk
# `-F`: The `-F fs` option defines the input field separator to be the regular expression `fs`
echo "/gas/string" | awk -F/ '{print $NF}'

# 특정 인덱스 필드
echo "/gasg/string" |cut -d/ -f 3
```

## 패턴 매칭

```shell
# 패턴 매칭
# ㄴ https://stackoverflow.com/a/229606
string='My long string'
if [[ $string == *"My long"* ]]; then
  echo "It's there!"
fi

# 패턴 not 매칭
if [[ "local|dev|stg|prod" != *"$PORT_CORE_PROFILE"* ]]
then
  echo "PORT_CORE_PROFILE is '$PORT_CORE_PROFILE'. Please do 'export PORT_CORE_PROFILE=<one of local,dev,stg,prod>' or run with '--profile=<one of dev,stg,prod>'" && exit 1
fi
```
