# corewebhook

- [corewebhook](#corewebhook)
    - [docker-compose.yaml](#docker-composeyaml)
    - [postgresql](#postgresql)
        - [initdb](#initdb)
    - [localstack](#localstack)
        - [`hosts` 수정](#hosts-수정)
        - [functions](#functions)
        - [SQS](#sqs)
            - [큐 만들기](#큐-만들기)
            - [테스트 해보기](#테스트-해보기)
            - [람다 또는 다른 컨테이너에서 큐 접근하기](#람다-또는-다른-컨테이너에서-큐-접근하기)
        - [MSK(kafka)](#mskkafka)

## docker-compose.yaml

```yaml
version: "3.8"

name: corewebhook
services:
    postgres:
        image: postgres:latest
        container_name: corewebhook_db
        ports:
            - "127.0.0.1:55432:5432"
        environment:
            POSTGRES_USER: corewebhook
            POSTGRES_PASSWORD: corewebhook
            POSTGRES_DB: corewebhook
        volumes:
            - /var/run/docker.sock:/var/run/docker.sock
    localstack:
        container_name: "${LOCALSTACK_DOCKER_NAME-localstack_main}"
        image: localstack/localstack
        ports:
            - "127.0.0.1:4566:4566" # LocalStack Gateway
            - "127.0.0.1:4510-4559:4510-4559" # external services port range
        environment:
            - DEBUG=${DEBUG-}
            - DOCKER_HOST=unix:///var/run/docker.sock
            - AWS_ACCESS_KEY_ID=localstack-access-key
            - AWS_SECRET_ACCESS_KEY=localstack-secret-key
            - AWS_DEFAULT_REGION=ap-northeast-2
        volumes:
            - "/var/run/docker.sock:/var/run/docker.sock"
            - "./localstack/volume:/var/lib/localstack"
            # - "${LOCALSTACK_VOLUME_DIR:-./volume}:/var/lib/localstack"
```

- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` :  rate limit 에러가 발생 방지

    ```log
    {"level":"error","@timestamp":"2023-09-26T14:29:52+09:00","caller":"retry/queue.go:125","msg":"unable to receive message from sqs","etc":{"error":"operation error SQS: ReceiveMessage, failed to sign request: failed to retrieve credentials: failed to refresh cached credentials, no EC2 IMDS role found, operation error ec2imds: GetMetadata, failed to get rate limit token, retry quota exceeded, 0 available, 5 requested"},"stacktrace":"github.com/some-org/corewebhook/retry.(*sqsQueue).receiveMessage\n\t/Users/rody/IdeaProjects/some-qwerty-org.io-go/corewebhook/retry/queue.go:125"}
    ```

- `AWS_DEFAULT_REGION`: region이 모호하다고 bad request 에러 나는 경우 방지([참고](https://github.com/localstack/localstack/issues/7869))

    ```log
    localstack_1        | 2023-03-14T21:54:10.489 DEBUG --- [   asgi_gw_0] l.services.sqs.query_api   : Region of queue URL http://localstack:4566/000000000000/queue-name is ambiguous, got region us-east-1 from request
    ```

## postgresql

### initdb

> NOTE: root 유저로는 실행 못한다
>
> ```shell
> root@d0c3831aeba7:/# initdb
> initdb: error: cannot be run as root
> initdb: hint: Please log in (using, e.g., "su") as the (unprivileged) user that will own the server process
> ```

```shell
su postgres

initdb
```

> NOTE: `/var/lib/postgresql/data` 삭제
>
> ```shell
> initdb: error: directory "/var/lib/postgresql/data" exists but is not empty
> initdb: hint: If you want to create a new database system, either remove or empty the directory "/var/lib/postgresql/data" or run initdb with an argument other than "/var/lib/postgresql/data".
> ```
>
> 근데 실제로 삭제하려면 실패
>
> ```shell
> root@d0c3831aeba7:/# rm -rf /var/lib/postgresql/data
> rm: cannot remove '/var/lib/postgresql/data': Device or resource busy
> ```
>
> [스택오버 플로우](https://stackoverflow.com/a/63824663) 참고.

## localstack

### `hosts` 수정

```shell
sudo vi /etc/hosts

# Local stack
127.0.0.1 local.stack
```

### functions

```shell
function localstack_main(){
    docker exec -it localstack_main /bin/bash
}
```

### SQS

#### 큐 만들기

```shell
SQS_ENDPOINT_STRATEGY=path awslocal sqs create-queue --queue-name corewebhook-retry-queue
```

- `SQS_ENDPOINT_STRATEGY`: [Queue URLs 전략](https://docs.localstack.cloud/user-guide/aws/sqs/#queue-urls)

```json
// 결과
{
    "QueueUrl": "http://localhost:4566/000000000000/corewebhook-retry-queue"
}
```

#### 테스트 해보기

- [공통 파라미터](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/CommonParameters.html)
- 메시지 생성하기

    ```shell
    curl -XGET  -H 'Accept: application/json' 'http://local.stack:4566/000000000000/corewebhook-retry-queue?Action=SendMessage&MessageBody=hello%2Fworld'
    ```

    ```json
    {
        "SendMessageResponse": {
            "SendMessageResult": {
                "MD5OfMessageBody": "c6be4e95a26409675447367b3e79f663",
                "MessageId": "8eb1614e-c2b9-4bed-b2b5-147b78a17b21"
            },
            "ResponseMetadata": {
                "RequestId": "5d38e031-6292-4d14-ada4-47c53f66d555"
            }
        }
    }
    ```

- 메시지 수신하기

    ```shell
    curl -XPOST \
        -H 'Accept: application/json' \
        -d @receive_message.json \
        'http://local.stack:4566/000000000000/corewebhook-retry-queue?Action=ReceiveMessage'
    ```

    ```json
    // 요청 body
    {
        "AttributeNames": ["All"],
        "ReceiveRequestAttemptId": "string",
        "MaxNumberOfMessages": 10,
        "VisibilityTimeout": 10, // `ReceiveMessage` 요청 수신 후 이 기간 동안 메시지가 안 보여진다.
        "WaitTimeSeconds": 10
    }
    ```

    ```json
    // response body 
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

    ```json
    // 메시지 삭제된 경우
    {
        "ReceiveMessageResponse": {
            "ReceiveMessageResult": null,
            "ResponseMetadata": {
            "RequestId": "7a2a35a7-b899-4de6-b524-1db1203a3e36"
            }
        }
    }
    ```

- 메시지 삭제하기

    ```shell
    curl -XPOST \
        -H 'Accept: application/json' \
        -d @delete_message.json \
        'http://local.stack:4566/000000000000/corewebhook-retry-queue?Action=DeleteMessage'
    ```

    ```text
    ReceiptHandle=YjkzMWNkNmUtYzQyMi00YzZkLWFlMzMtNjA4MTY5NzUxNTJiIGFybjphd3M6c3FzOnVzLWVhc3QtMTowMDAwMDAwMDAwMDA6Y29yZXdlYmhvb2stcmV0cnktcXVldWUgOGViMTYxNGUtYzJiOS00YmVkLWIyYjUtMTQ3Yjc4YTE3YjIxIDE2OTU2OTYxODkuMDIyNjc1
    ```

    ```json
    {
        "DeleteMessageResponse": {
            "ResponseMetadata": {
                "RequestId": "4c0e78ea-a8c0-4ff7-aa8d-c136d7a1d643"
            }
        }
    }
    ```

#### [람다 또는 다른 컨테이너에서 큐 접근하기](https://docs.localstack.cloud/user-guide/aws/sqs/#accessing-queues-from-lambdas-or-other-containers)

- 내 컨테이너에서 접근하기

    ```shell
    # create the network
    docker network create my-network

    # launch localstack

    DOCKER_FLAGS="--network my-network" localstack start

    # launch your container

    docker run --rm it --network my-network <image name>

    # then your code can access localstack at its container name (by default: localstack_main)
    ```

    ```yaml
    services:
        localstack:
            # ... other configuration here
            networks:
            - ls
        your_container:
            # ... other configuration here
            networks:
            - ls
        networks:
        ls:
            name: ls

        # Your application code can then use
        # <http://localstack:4566> for the endpoint url
    ```

### MSK(kafka)

```json
// brokernodegroupinfo.json
{
    "InstanceType": "kafka.m5.xlarge",
    "BrokerAZDistribution": "DEFAULT",
    // three subnets where you want your local Amazon MSK to distribute the broker nodes.
    "ClientSubnets": [
        "subnet-0123456789111abcd",
        "subnet-0123456789222abcd",
        "subnet-0123456789333abcd"
    ]
}
```

```shell
awslocal kafka create-cluster \
    --cluster-name "EventsCluster" \
    --broker-node-group-info file://brokernodegroupinfo.json \
    --kafka-version "2.2.1" \
    --number-of-broker-nodes 3
```
