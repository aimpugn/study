# Customs

## `ftoj`

> Form to JSON

```bash
function ftoj() {
    local formData="$1"  # 입력된 form 데이터를 인자로부터 받아옵니다. 이는 함수의 첫 번째 인자로 전달되는 값입니다.
    
    # Python의 urllib.parse.parse_qs를 사용하여 form data를 파싱하고, 결과를 JSON으로 변환합니다.
    # 이 때, 각 키에 대응하는 값이 JSON 값인지를 판별하고, 이에 따라 적절하게 처리합니다.
    local json=$(python3 -c "
import json, urllib.parse

def try_parse_json(s):
    try:
        # s가 JSON 문자열이면 이를 파싱하여 반환합니다.
        return json.loads(s)
    except json.JSONDecodeError:
        # s가 JSON 문자열이 아니면 그대로 반환합니다.
        return s

data = urllib.parse.parse_qs('$formData')  # form data를 파싱하여 딕셔너리로 변환합니다. 이 때, 각 키에 대응하는 값은 리스트입니다.

# 각 키에 대응하는 값이 JSON 문자열인지 판별하고, 이에 따라 적절하게 처리합니다.
# 이를 위해, 각 키에 대응하는 값(리스트의 첫 번째 요소)을 try_parse_json 함수에 전달하여 JSON 파싱을 시도합니다.
# 만약 값이 JSON 문자열이면 파싱된 결과가, 그렇지 않으면 원래의 문자열이 딕셔너리의 값이 됩니다.
# data = {k: try_parse_json(v[0]) for k, v in data.items()}
data = {k: v[0] if k in ['buyer_tel', 'transid', 'refundtransid', 'resdt'] else try_parse_json(v[0]) for k, v in data.items()}

print(json.dumps(data))  # 딕셔너리를 JSON 문자열로 변환하여 출력합니다.")
    
    # 최종적으로 생성된 JSON 객체를 jq를 사용하여 pretty print로 출력합니다.
    echo "$json" | jq .
}
```

`jq`를 사용하여 더 간소화시킬 수 있다. 대신 `key[]=value&key[]=value2&key[]=value3` 같은 경우에 대해서는 대응이 안 된다

```bash
ftoj() {

    echo "$1" \
        # & 문자를 개행 문자(\n)로 변환
        | tr '&' '\n' \
        | jq -Rn '[inputs | split("=") | {(.[0]): .[1]}] | add'
}
```
