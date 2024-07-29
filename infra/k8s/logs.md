# Logs

- [Logs](#logs)
    - [pod log](#pod-log)

## pod log

```shell
# 환경 리스트 보기
kubectl config get-contexts
# 환경으로 세팅
kubectx port-aws-dev
# 팟 이름 보기
kubectl get pods -n crawler-service
# 팟 로그 생성
kubectl logs \
    -f crawler-service-6ffcf768c-wctqc \
    -n crawler-service > dev-crawler-service-2023-02-14-01.log
# 로그 보기
tail -f dev-crawler-service-2023-02-14-01.log
```
