# cURL and jq

## 액세스 토큰 뽑아서 사용하기

```bash
#!/bin/bash

TOKEN=$(curl -s -H 'Content-Type: application/json' \
  -d '{"apiKey":"apiKeyValue","apiSecret":"apiSecretValue"}' \
  https://core-api.dev.domain.co/users/getToken)

ACCESS_TOKEN=$(jq -r -n --arg token "$TOKEN" '$token | fromjson | .response.access_token')

echo "access_token: $ACCESS_TOKEN"

curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  'https://core-api.dev.domain.co/payments/status/all?from=1704034800&to=1706713200&page=1&limit=100' |
  jq '{code:.code,total:.response.total}'
```

- `jq`
    - `-r`: raw-output
