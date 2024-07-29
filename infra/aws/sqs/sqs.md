# SQS

- [SQS](#sqs)
    - [create Queue](#create-queue)
    - [Create message](#create-message)
    - [Receive message](#receive-message)
    - [Delete message](#delete-message)

## create Queue

```shell
SQS_ENDPOINT_STRATEGY=path awslocal sqs create-queue --queue-name $QUEUE_NAME
```

```json
// response
{
    "QueueUrl": "http://localhost:4566/000000000000/corewebhook-retry-queue"
}
```

## Create message

```shell
curl -XGET  -H 'Accept: application/json' \
    'http://local.stack:4566/000000000000/corewebhook-retry-queue?Action=SendMessage&MessageBody=hello%2Fworld'
```

## Receive message

```json
// receive_message.json
{
    "AttributeNames": ["All"],
    "ReceiveRequestAttemptId": "string",
    "MaxNumberOfMessages": 10,
    "VisibilityTimeout": 10, // `ReceiveMessage` 요청 수신 후 이 기간 동안 메시지가 안 보여진다.
    "WaitTimeSeconds": 10
}
```

```shell
curl -XPOST \
  -H 'Accept: application/json' \
  -d @receive_message.json \
  'http://local.stack:4566/000000000000/corewebhook-retry-queue?Action=ReceiveMessage'
```

```json
// response
{
    "ReceiveMessageResponse": {
        "ReceiveMessageResult": {
            "Message": {
                "MessageId": "8eb1614e-c2b9-4bed-b2b5-147b78a17b21",
                "ReceiptHandle": "NGI5YzkwOWEtODc0YS00ZDQ3LWI2MmEtYjJjYzNkMWY4YTY3IGFybjphd3M6c3FzOnVzLWVhc3QtMTowMDAwMDAwMDAwMDA6Y29yZXdlYmhvb2stcmV0cnktcXVldWUgOGViMTYxNGUtYzJiOS00YmVkLWIyYjUtMTQ3Yjc4YTE3YjIxIDE2OTU2OTU1NTAuMjYyNzQ4Mg==",
                "MD5OfBody": "c6be4e95a26409675447367b3e79f663",
                "Body": "hello/world"
            }
        },
        "ResponseMetadata": {
            "RequestId": "67a8b6a7-aa1f-4351-b076-9847b356883e"
        }
    }
}
```

## Delete message

```text
# delete_message.txt
ReceiptHandle=YjkzMWNkNmUtYzQyMi00YzZkLWFlMzMtNjA4MTY5NzUxNTJiIGFybjphd3M6c3FzOnVzLWVhc3QtMTowMDAwMDAwMDAwMDA6Y29yZXdlYmhvb2stcmV0cnktcXVldWUgOGViMTYxNGUtYzJiOS00YmVkLWIyYjUtMTQ3Yjc4YTE3YjIxIDE2OTU2OTYxODkuMDIyNjc1
```

```shell
curl -XPOST \
    -H 'Accept: application/json' \
    -d @delete_message.txt \
    'http://local.stack:4566/000000000000/corewebhook-retry-queue?Action=DeleteMessage'
```

```json
// response
{
    "DeleteMessageResponse": {
        "ResponseMetadata": {
            "RequestId": "4c0e78ea-a8c0-4ff7-aa8d-c136d7a1d643"
        }
    }
}
```
