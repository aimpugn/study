# Settings

- [Settings](#settings)
    - [`GOROOT`와 `GOPATH`](#goroot와-gopath)
    - [GoLand](#goland)

## `GOROOT`와 `GOPATH`

- `GOROOT`:
    - variable that points to the directory **where the Go language is installed on your system**.
    - It contains the Go compiler, standard library, and other tools necessary for working with the language
    - When you install Go, this variable is usually set for you automatically.
- `GOPATH`
    - environment variable that specifies **the workspace directory for your Go projects**
    - since Go 1.11, with the introduction of Go modules, you no longer need to set GOPATH explicitly
    - Go modules cache your dependencies in the `$GOPATH/pkg/mod`

```shell
# 1. GOROOT
GOROOT=/opt/homebrew/Cellar/go/1.19.2/libexec #gosetup
# 2. GOPATH
GOPATH=/Users/rody/IdeaProjects/payment-info-service:/Users/rody/go #gosetup
# 3. go mod
/opt/homebrew/Cellar/go/1.19.2/libexec/bin/go list -modfile=/Users/rody/IdeaProjects/payment-info-service/go.mod -m -json -mod=mod all #gosetup
```

## GoLand

- 내부 패키지, 외부 패키지, std 패키지 별로 정리하기
    - `Settings` > `Editor` > `Code Style` > `Go` > `Imports`에서
    - `Group stdlib imports`
    - `Move all stdlib import in a single group`,
    - `Group` -> `Current proejct packages`
