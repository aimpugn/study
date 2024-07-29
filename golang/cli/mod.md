# go mod

- [go mod](#go-mod)
    - [커맨드 목록](#커맨드-목록)
    - [`go mod tidy`](#go-mod-tidy)

## 커맨드 목록

- `download`: download modules to local cache
- `edit`: edit go.mod from tools or scripts
- `graph`: print module requirement graph
- `init`: initialize new module in current directory
- `tidy`: add missing and remove unused modules
- `vendor`: make vendored copy of dependencies
- `verify`: verify dependencies have expected content
- `why`: explain why packages or modules are needed

## `go mod tidy`

> add missing and remove unused modules

```shell
GOROOT=/opt/homebrew/Cellar/go/1.20.3/libexec #gosetup
GOPATH=/Users/rody/go #gosetup
/opt/homebrew/Cellar/go/1.20.3/libexec/bin/go fmt /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/danal.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/eximbay.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/inicis.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/nice.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/payco.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/paymentwall.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/provider.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/settle.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/settle_acc.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/smartro.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/standard.go /Users/rody/IdeaProjects/some-qwerty-org.io-go/some-qwerty-api-service/payments/code/v2.go #gosetup
go: updates to go.mod needed; to update it:
    go mod tidy
```
