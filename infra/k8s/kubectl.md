# commands

- [commands](#commands)
    - [공통](#공통)
        - [JSON 형태 출력: `-o json`](#json-형태-출력--o-json)
    - [`get namespaces`](#get-namespaces)
    - [파드의 상세 정보를 조회: `describe pod`](#파드의-상세-정보를-조회-describe-pod)
        - [특정 파드의 상세 정보 조회: `describe pod <pod-name>`](#특정-파드의-상세-정보-조회-describe-pod-pod-name)
    - [파드 목록 가져오기: `get pods`](#파드-목록-가져오기-get-pods)
        - [전체 파드 목록 가져오기: `get pods -A`](#전체-파드-목록-가져오기-get-pods--a)
        - [더 많은 정보와 함께 출력: `get pods -o wide`](#더-많은-정보와-함께-출력-get-pods--o-wide)
        - [모든 파드와 그에 연관된 레이블을 나열: `get pods --show-labels`](#모든-파드와-그에-연관된-레이블을-나열-get-pods---show-labels)
        - [레이블 선택자: `get pods -l app=<something>`](#레이블-선택자-get-pods--l-appsomething)
        - [특정 네임스페이스의 파드를 나열: `get pods -n <namespace>`](#특정-네임스페이스의-파드를-나열-get-pods--n-namespace)
        - [커스텀 칼럼을 지정하여 원하는 정보만을 출력: `get pods -o custom-columns=`](#커스텀-칼럼을-지정하여-원하는-정보만을-출력-get-pods--o-custom-columns)
    - [쿠버네티스 컨트롤메쉬 API를 사용](#쿠버네티스-컨트롤메쉬-api를-사용)

## 공통

### JSON 형태 출력: `-o json`

```shell
kubectl get pods -A -o json
```

## `get namespaces`

```shell
kubectl get namespaces
```

## 파드의 상세 정보를 조회: `describe pod`

### 특정 파드의 상세 정보 조회: `describe pod <pod-name>`

## 파드 목록 가져오기: `get pods`

### 전체 파드 목록 가져오기: `get pods -A`

```shell
kubectl get pods -A
```

```log
NAMESPACE                      NAME                                                   READY   STATUS      RESTARTS   AGE
argo-events                    argo-events-eventbus-controller-6d4666c9f7-jxrf5       1/1     Running     0          118d
```

### 더 많은 정보와 함께 출력: `get pods -o wide`

```shell
kubectl get pods -A
```

```log
NAME                                  READY   STATUS    RESTARTS   AGE    IP             NODE                                               NOMINATED NODE   READINESS GATES
tx-gateway-service-5d8d48bdd9-wvxbs   2/2     Running   0          11d    10.32.78.14    ip-10-32-78-28.ap-northeast-2.compute.internal     <none>           <none>
```

### 모든 파드와 그에 연관된 레이블을 나열: `get pods --show-labels`

```shell
kubectl get pods --show-labels
```

### 레이블 선택자: `get pods -l app=<something>`

```shell
kubectl get pods -l app=nginx
```

### 특정 네임스페이스의 파드를 나열: `get pods -n <namespace>`

```shell
kubectl get pods -n kube-system
```

```log
NAME                                  READY   STATUS    RESTARTS   AGE
aws-node-9bppp                        1/1     Running   0          124d
aws-node-jpjhs                        1/1     Running   0          118d
aws-node-kxbqf                        1/1     Running   0          124d
```

### 커스텀 칼럼을 지정하여 원하는 정보만을 출력: `get pods -o custom-columns=`

```shell
kubectl get pods \
    -o custom-columns='NAME:.metadata.name,CONTAINERS:.spec.containers[*].name'
```

```log
NAME                                  CONTAINERS
tx-gateway-service-5d8d48bdd9-wvxbs   istio-proxy,tx-gateway-service
tx-gateway-service-postgresql-0       postgresql
tx-gateway-service-redis-cluster-0    tx-gateway-service-redis-cluster
tx-gateway-service-redis-cluster-1    tx-gateway-service-redis-cluster
tx-gateway-service-redis-cluster-2    tx-gateway-service-redis-cluster
```

## 쿠버네티스 컨트롤메쉬 API를 사용

```bash
printf blabla | kubectl apply -f
```
