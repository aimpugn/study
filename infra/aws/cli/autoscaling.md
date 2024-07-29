# autoscaling

## `describe-tags`

### filter

```bash
aws --profile port-aws-stg  \
    --region ap-northeast-2 \
    autoscaling describe-tags \
    --filters "Name=auto-scaling-group,Values=core-pay" \
    | cat
```

## `create-or-update-tags`

```bash
aws  --profile port-aws-stg  \
    --region ap-northeast-2 \
    autoscaling create-or-update-tags \
    --tags ResourceId=core-pay,ResourceType=auto-scaling-group,Key=Version,Value=hotfix/change-enum-from-point-to-charge,PropagateAtLaunch=true
```

- `ResourceId`
    - auto scaling group의 이름
- `ResourceType`
- `PropagateAtLaunch`
    - `true`
        - Auto Scaling 그룹에 추가된 모든 새 EC2 인스턴스에 태그가 자동으로 적용된다
        - 이는 **그룹의 확장/축소 동안 생성되는 모든 인스턴스에 태그가 전파되어야 할 때** 유용하다
    - `false` 또는 생략
        - 태그는 Auto Scaling 그룹 자체에만 적용되며, 새로 생성되는 EC2 인스턴스에는 전파되지 않는다
        - 이 설정은 태그를 그룹 관리 목적으로만 사용하고자 할 때 적합하다
