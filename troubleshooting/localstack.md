# localstack

## operation error SQS: ReceiveMessage, https response error StatusCode: 403

### 문제

```json
{
    "level": "error",
    "@timestamp": "2023-10-04T16:19:42+09:00",
    "caller": "retry/queue.go:125",
    "msg": "unable to receive message from sqs",
    "etc": {
        "error": "operation error SQS: ReceiveMessage, https response error StatusCode: 403, RequestID: c652a79f-4e38-56a8-899f-fa085a768b63, api error InvalidClientTokenId: The security token included in the request is invalid."
    },
    "stacktrace": "github.com/some-org/corewebhook/retry.(*sqsQueue).receiveMessage\n\t/Users/rody/IdeaProjects/some-qwerty-org.io-go/corewebhook/retry/queue.go:125"
}

```

### 원인

```shell
aws sts get-caller-identity --profile localstack --endpoint-url http://localhost:4566 --no-cli-pager
```

```shell
aws sts get-caller-identity --profile localstack --endpoint-url http://localhost:4566 --no-cli-pager
{
    "UserId": "AKIAIOSFODNN7EXAMPLE",
    "Account": "000000000000",
    "Arn": "arn:aws:iam::000000000000:root"
}
```

### 해결
