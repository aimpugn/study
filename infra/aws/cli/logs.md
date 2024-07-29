# logs

## 로그에 쿼리 요청하고 결과 받기

```shell
aws logs start-query \
 --profile port-aws-prod \
 --log-group-name ${log_group_name} \
 --start-time 1681171200 \
 --end-time `date "+%s"` \
 --query-string 'fields @message | filter @message like "test" | limit 10000' \
| jq '.queryId'
```

```json
{
    "queryId": "cb4e9a27-0937-4a8f-b7f8-6f4876153f59"
}
```

```shell
aws --profile ${profile} logs get-query-results --query-id cb4e9a27-0937-4a8f-b7f8-6f4876153f59
```
