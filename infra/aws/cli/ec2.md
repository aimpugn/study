# aws ec2

- [aws ec2](#aws-ec2)
    - [`describe-instances`](#describe-instances)
        - [instance ip 확인](#instance-ip-확인)
        - [filter](#filter)
        - [query](#query)
            - [인스턴스 아이디, IP, 특정 태그 값, 실행 상태 테이블로 출력](#인스턴스-아이디-ip-특정-태그-값-실행-상태-테이블로-출력)
    - [`create-tags`](#create-tags)
        - [태그 업데이트](#태그-업데이트)

## `describe-instances`

```bash
aws --profile port-aws-stg \
    --region ap-northeast-2 \
    ec2 describe-instances
```

### instance ip 확인

```shell
# 이 명령으로 조회 가능 (stg대신 prod 적으면 prod 조회)
aws --profile port-aws-stg \
    --region ap-northeast-2 \
    ec2 describe-instances \
    --filters "Name=tag:Name,Values=core*asg" \
    # --filters "Name=tag:Name,Values=*api*"\
    # --filters "Name=tag:Name,Values=*core-pay-asg" \
    --query "Reservations[*].Instances[*].[PrivateIpAddress, Tags[?Key=='Name'] | [0].Value, State.Name]"\
    --output table \
    | cat
    # --output text | nl
```

### filter

`describe-instances` 결과중 특정 값에 해당하는 경우로 결과를 제한할 수 있다

```bash
aws --profile port-aws-stg \
    --region ap-northeast-2 \
    ec2 describe-instances \
    --filters "Name=tag:Name,Values=*pay*"
```

### query

`describe-instances` 결과 json이 아래와 같을 때, 다양한 쿼리 실행 가능

```json
{
    "Reservations": [
        {
            "Instances": [
                {
                    "PrivateIpAddress": "10.0.1.1",
                    "Tags": [
                        {
                            "Key": "Name",
                            "Value": "tag-name-what-i-want"
                        }
                    ]
                }
            ]
        }
    ]   
}
```

```bash
aws --profile port-aws-stg \
    --region ap-northeast-2 \
    ec2 describe-instances \
    --query "Reservations[*].Instances[*].[PrivateIpAddress, Tags[?Key=='Name'] | [0].Value, State.Name]"
```

- `Reservations[*].Instances[*]`:
    - `Reservations` 배열의 모든 요소(각 예약)에서
    - `Instances` 배열의 모든 요소(각 인스턴스)를 선택
- `[PrivateIpAddress, Tags[?Key=='Name'] | [0].Value, State.Name]`
    - `Instances` 배열의 각 요소를 `[]`로 보고,
    - 각 요소(`[]`) 안에 있는 `PrivateIpAddress`, `Tags`, `State` 키들을 선택

#### 인스턴스 아이디, IP, 특정 태그 값, 실행 상태 테이블로 출력

```bash
❯ aws --profile port-aws-stg \
    --region ap-northeast-2 \
    ec2 describe-instances \
    --filters "Name=tag:Name,Values=*pay*"\
    --query "Reservations[*].Instances[*].[InstanceId, PrivateIpAddress, Tags[?Key=='Version'] | [0].Value, State.Name]" \
    --output table | cat
------------------------------------------------------------
|                     DescribeInstances                    |
+----------------------+----------------+-------+----------+
|  i-0bda4aaabdf532507 |  10.16.75.54   |  main |  running |
|  i-0517c3e24de477695 |  10.16.107.64  |  main |  running |
+----------------------+----------------+-------+----------+
```

```bash
aws --profile port-aws-stg  \
    --region ap-northeast-2 \
    ec2 describe-tags \
    --filters "Name=resource-id,Values=i-0bda4aaabdf532507"
```

## `create-tags`

```bash
aws ec2 create-tags --resources <InstanceId> --tags Key=<TagName>,Value=<NewValue>
```

### 태그 업데이트

```bash
aws --profile port-aws-stg\
    --region ap-northeast-2 \
    ec2 create-tags \
    --resources i-0bda4aaabdf532507 \
    --tags Key=Version,Value=hotfix/change-enum-from-point-to-charge
```
