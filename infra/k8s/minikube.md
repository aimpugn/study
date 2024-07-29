# minikube

- [minikube](#minikube)
    - [Quick Start](#quick-start)
    - [애플리케이션 배포](#애플리케이션-배포)

## [Quick Start](https://minikube.sigs.k8s.io/docs/start/)

- macos
- arm64
- Stable
- brew

```shell
brew install minikube
```

```shell
minikube start
```

alias 추가

```.zshrc
alias mkubectl="minikube kubectl --"
```

```shell
kubectl get po -A
```

## 애플리케이션 배포

- ingress

```shell
minikube addons enable ingress
```
