# Go

- [Go](#go)
    - [$GOPATH/go.mod exists but should not](#gopathgomod-exists-but-should-not)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [Build constraints exclude all the Go files in](#build-constraints-exclude-all-the-go-files-in)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [inner declaration of var err error](#inner-declaration-of-var-err-error)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [gopls was not able to find modules in your workspace.When outside of GOPATH, gopls needs to know which modules you are working on ~](#gopls-was-not-able-to-find-modules-in-your-workspacewhen-outside-of-gopath-gopls-needs-to-know-which-modules-you-are-working-on-)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [someService'(타입 someService)을(를) 타입 some.Service(으)로 사용할 수 없습니다 'SomeFunction' 메서드가 포인터 리시버이므로 타입이 'some.Service'을(를) 구현하지 않습니다](#someservice타입-someservice을를-타입-someservice으로-사용할-수-없습니다-somefunction-메서드가-포인터-리시버이므로-타입이-someservice을를-구현하지-않습니다)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [GoFmt returned non-zero code on some of the files. Would you like to commit anyway?](#gofmt-returned-non-zero-code-on-some-of-the-files-would-you-like-to-commit-anyway)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
    - [I don't know what to return because the method call was unexpected](#i-dont-know-what-to-return-because-the-method-call-was-unexpected)
        - [문제](#문제-6)
        - [원인](#원인-6)
        - [해결](#해결-6)
    - [go: cannot find main module, but found .git/config in /some/path/to](#go-cannot-find-main-module-but-found-gitconfig-in-somepathto)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-7)
    - [could not import github.com/gofiber/fiber/v2 (invalid package name: "")"](#could-not-import-githubcomgofiberfiberv2-invalid-package-name-)
        - [문제](#문제-8)
        - [원인](#원인-8)
        - [해결](#해결-8)
    - [fatal: could not read Username for `https://github.com`: terminal prompts disabled](#fatal-could-not-read-username-for-httpsgithubcom-terminal-prompts-disabled)
        - [문제](#문제-9)
        - [원인](#원인-9)
        - [해결](#해결-9)
    - [`before_org` -\> `after_org`로 바꾸는 작업 후 에러 발생](#before_org---after_org로-바꾸는-작업-후-에러-발생)
        - [문제](#문제-10)
        - [원인](#원인-10)
        - [해결](#해결-10)
    - [GOPROXY list is not the empty string, but contains no entries](#goproxy-list-is-not-the-empty-string-but-contains-no-entries)
        - [문제](#문제-11)
        - [원인](#원인-11)
            - [주요 원인](#주요-원인)
        - [해결](#해결-11)
    - [gvm\_implode:read:1: -p: no coprocess](#gvm_imploderead1--p-no-coprocess)
        - [문제](#문제-12)
        - [원인](#원인-12)
        - [해결](#해결-12)
    - [could not import slices (current file is not included in a workspace module)](#could-not-import-slices-current-file-is-not-included-in-a-workspace-module)
        - [문제](#문제-13)
        - [원인](#원인-13)
        - [해결](#해결-13)
    - [compile: version 'go1.20.3' does not match go tool version 'go1.22.2'](#compile-version-go1203-does-not-match-go-tool-version-go1222)
        - [문제](#문제-14)
        - [원인](#원인-14)
        - [해결](#해결-14)
    - [failed to decode elasticsearch response: EOF](#failed-to-decode-elasticsearch-response-eof)
        - [문제](#문제-15)
        - [원인](#원인-15)
        - [해결](#해결-15)
            - [`res.Body`의 내용을 바이트 슬라이스에 저장 후 재사용](#resbody의-내용을-바이트-슬라이스에-저장-후-재사용)
            - [`res.Body` 자체를 복사하는 방법](#resbody-자체를-복사하는-방법)
            - [`httputil.DumpResponse` 함수를 사용하는 방법](#httputildumpresponse-함수를-사용하는-방법)

## $GOPATH/go.mod exists but should not

### 문제

intellij에서 go rebuild 시에 에러 발생

```log
GOROOT=/opt/homebrew/Cellar/go/1.19.2/libexec #gosetup
GOPATH=/Users/rody/IdeaProjects/payment-info-service:/Users/rody/go #gosetup
/opt/homebrew/Cellar/go/1.19.2/libexec/bin/go list -modfile=/Users/rody/IdeaProjects/payment-info-service/go.mod -m -json -mod=mod all #gosetup
$GOPATH/go.mod exists but should not
```

### 원인

- [`GOPATH`로 지정한 곳에 go.mod 파일이 있으면 안된다](https://velog.io/@artelee/%EC%98%A4%EB%A5%981.-GOPATHgo.mod-exists-but-should-not)
- `GOPATH`는 워크스페이스 경로를 의미하고, 해당 워크스페이스에서 여러 프로젝트 관리할 수 있다.
- [모듈을 사용할 때, `GOPATH`는 경로 리졸브 시에 더이상 사용되지 않지만, 다운로드한 소스코드(`$GOAPTH/pkg/mod`) 및 컴파일된 커맨드(`$GOPATH/bin`)를 저장하는 데 사용된다](https://stackoverflow.com/a/72709319)

### 해결

- Intellij에서 GOPATH로 프로젝트 경로를 설정했는데, 이를 삭제 후 `go mod tidy`

## Build constraints exclude all the Go files in

### 문제

go로 된 프로젝트를 오랜만에 pull 받았는데, 변경 사항이 매우 많았고, 아래 명령어로 mod 설치 후

```shell
GOROOT=/opt/homebrew/Cellar/go/1.19.2/libexec #gosetup
GOPATH=/Users/rody/go #gosetup
/opt/homebrew/Cellar/go/1.19.2/libexec/bin/go mod tidy #gosetup
```

에러 발생

```log
Build constraints exclude all the Go files in '/Users/rody/IdeaProjects/some-qwerty-org.io-go/corewebhook/domain'
```

```log
Build constraints exclude all the Go files in '/Users/rody/go/pkg/mod/github.com/labstack/echo/v4@v4.9.1'
```

또는 아래 같은 에러가 발생함. 근데 `https://github.com/some-org/corewebhook`는 실제로 존재하는 리파지토리가 아니라, 하나의 큰 리파지토리의 서브 모듈이고, go.mod에 `module github.com/some-org/corewebhook` 라고 선언되어 있을 뿐임

```text
github.com/some-org/corewebhook/webhook imports
        github.com/some-org/corewebhook/domain: cannot find module providing package github.com/some-org/corewebhook/domain: module github.com/some-org/corewebhook/domain: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/d4c53f483c85125ed08bcc43c13e3d5177ab390f66ce231f8eca609a8bea11ef: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/some-org/corewebhook/' not found
github.com/some-org/corewebhook tested by
        github.com/some-org/corewebhook.test imports
        github.com/some-org/corewebhook/webhook/mocks: cannot find module providing package github.com/some-org/corewebhook/webhook/mocks: module github.com/some-org/corewebhook/webhook/mocks: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/d4c53f483c85125ed08bcc43c13e3d5177ab390f66ce231f8eca609a8bea11ef: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/some-org/corewebhook/' not found
```

상위 디렉토리의 go.work에 아래처럼 정의되어 있다

```work
go 1.20

use (
    someotherfeature1
    someotherfeature2
    corewebhook
    someotherfeature3
    someotherfeature4
)
```

어쨌든 위와 같은 에러가 나서 `/Users/rody/go/pkg/mod/cache/vcs/d4c53f483c85125ed08bcc43c13e3d5177ab390f66ce231f8eca609a8bea11ef`를 찾아가 보니 아래 같은 구조로 되어 있음. 이게 뭘까?

```text
╭─░▒▓ ~/go/pkg/mod/cache/vcs                                                                                                                                                                          ✔  system ⬢  at 10:44:07 ▓▒░
╰─ pwd
/Users/rody/go/pkg/mod/cache/vcs

╭─░▒▓ ~/go/pkg/mod/cache/vcs                                                                                                                                                                          ✔  system ⬢  at 10:44:13 ▓▒░
╰─ tree .
.
├── d4c53f483c85125ed08bcc43c13e3d5177ab390f66ce231f8eca609a8bea11ef
│   ├── HEAD
│   ├── config
│   ├── description
│   ├── hooks
│   │   ├── applypatch-msg.sample
│   │   ├── commit-msg.sample
│   │   ├── fsmonitor-watchman.sample
│   │   ├── post-update.sample
│   │   ├── pre-applypatch.sample
│   │   ├── pre-commit.sample
│   │   ├── pre-merge-commit.sample
│   │   ├── pre-push.sample
│   │   ├── pre-rebase.sample
│   │   ├── pre-receive.sample
│   │   ├── prepare-commit-msg.sample
│   │   ├── push-to-checkout.sample
│   │   └── update.sample
│   ├── info
│   │   └── exclude
│   ├── objects
│   │   ├── info
│   │   └── pack
│   └── refs
│       ├── heads
│       └── tags
├── d4c53f483c85125ed08bcc43c13e3d5177ab390f66ce231f8eca609a8bea11ef.info
└── d4c53f483c85125ed08bcc43c13e3d5177ab390f66ce231f8eca609a8bea11ef.lock
```

### 원인

- Go 언어에서 빌드 제약조건(**build constraints**) 때문에 발생하는 것으로 보인다
    - **빌드 제약조건**? 특정 조건 하에서만 파일이나 패키지를 빌드하도록 설정하는데 사용된다
- 이 에러는 특정 디렉토리에 있는 모든 Go 파일이 현재 빌드 환경에 대한 제약조건을 만족하지 않아 제외되었음을 나타낸다
- [build constraints exclude all Go files in](https://stackoverflow.com/questions/55348458/build-constraints-exclude-all-go-files-in)
    - `go`는 `GO_OS` 설정하거나 또는 파일 이름을 `xyz_<OS>.go`로 만들면, 빌드할 파일에 대한 intelligence를 수행하며, 이떄 불행하게도 실제로는 파일이 있음에도 0개의 빌드할 파일이 있다고 착각하게 만들 수 있다
- 가능한 원인?
    - **플랫폼 또는 아키텍처 불일치**
        - Go 파일이나 패키지가 특정 플랫폼이나 아키텍처에 대해 빌드될 수 있도록 설정되었는데, 현재 환경이 그 조건을 만족하지 않을 때 이런 문제가 발생할 수 있다.
        - 예를 들어, 일부 파일이나 패키지는 특정 아키텍처(예: amd64 또는 386)에 대해서만 빌드될 수 있도록 설정될 수 있다
    - **`CGO` 사용**
        - `CGO`는 Go에서 C 코드를 호출할 수 있도록 해주는 도구, 일부 패키지는 `CGO`가 활성화된 상태에서만 빌드될 수 있다
        - `CGO`를 사용하려면 import "C" 구문을 사용하고, 이는 빌드 제약조건에 영향을 줄 수 있다
    - **빌드 태그 사용**
        - 파일 상단에 `// +build` 라인을 사용하여 빌드 제약조건을 지정할 수 있다
        - 이러한 빌드 태그는 파일이 특정 조건에서만 빌드되도록 지시할 수 있다
    - **모듈 모드의 문제**
        - Go 1.12 이후 모듈 모드를 사용하면서 go get -d 명령어와 관련된 이슈가 발생할 수 있다
        - 이 명령어는 모듈의 루트 패키지에 빌드 제약조건이 있는 경우 문제를 유발할 수 있다

### 해결

- 캐시의 문제일까?

    ```shell
    go clean -modcache
    ```

## inner declaration of var err error

### 문제

```log
./aes_gcm.go:34:3: result parameter err not in scope at return
    ./aes_gcm.go:33:8: inner declaration of var err error
```

### 원인

```go
func AesGcmEncrypt(keyInHex string, plaintext string) (ciphertext *string, err error) {
    key, err := hex.DecodeString(keyInHex)
    if err != nil {
        return
    }

    // iv (임의의 값, 12 바이트)
    iv := make([]byte, 12)
    if _, err := rand.Read(iv); err != nil {
    //    ^^^^^^^ 이 부분이 문제
        return
    }
```

- Go 언어에서 `:=` 연산자는 새로운 변수를 선언하고 값을 할당하는데 사용
- `함수의 반환 변수`(err)와 동일한 이름의 변수(err)를 함수 내부에서 `:=` 연산자를 사용하여 선언하면, 함수 내부의 새로운 범위(scope)에 새 변수가 선언된다
- 즉, 함수의 반환 변수와는 **다른 새 변수가 생성**되며, 이 새 변수는 함수 내부의 해당 범위에서만 사용할 수 있다
- 이후의 return 문에서는 **함수의 반환 변수 err가 아닌 함수 내부의 새 err 변수를 참조**하게 됩니다.

### 해결

`:=`를 제거

## gopls was not able to find modules in your workspace.When outside of GOPATH, gopls needs to know which modules you are working on ~

### 문제

```go
package main
// ^^^^^^^^^^^^
// gopls was not able to find modules in your workspace.
//
// When outside of GOPATH, gopls needs to know which modules you are working on.
//
// You can fix this by opening your workspace to a folder inside a Go module, or by using a go.work file to specify multiple modules.
// See the documentation for more information on setting up your workspace: https://github.com/golang/tools/blob/master/gopls/doc/workspace.md.

import (
    "fmt"
)
```

```log
Error loading workspace: 1 modules have errors: github.com/aimpugn/algo/kakao:pattern /Users/rody/IdeaProjects/snippets/algorithms/examples/go/kakao/...: directory prefix . does not contain modules listed in go.work or their selected dependencies
```

### 원인

```log
/Users/rody/IdeaProjects/snippets/algorithms/examples/go
```

[디렉토리 이름이 `go`여서 발생했던 문제로 보임](https://github.com/golang/go/issues/63536#issuecomment-1779982298)

```tree
monorepo/ <-- `go.work`는 `gopls`로 하여금 `monorepo`를 root로 사용하게 한다
  - go.work
  - go/
      - go.mod
      - go files...
  - cache/
  - js/
  - lots/of/other/dirs
```

```go
// go.work
go 1.21.3

use ./go
```

- `go.work` instructs `gopls` to use `monorepo/` as the root rather than `monorepo/go/`
- `gopls` requests to watch files `monorepo/**/*.{go,mod,sum}` (lots and lots of files)
- nvim's lsp now tries to watch way too many directories and files

Since this is kind of a 2-part issue (nvim lsp watch files implementation + gopls overbroad watchfiles glob) I decided to try to change the `root_path` myself and ended up with just a specific config for `gopls`:

```lua
lspconfig.gopls.setup({
  on_attach = [...],
  capabilities = [...],
  settings = [...],

  -- override root_path for issue: https://github.com/golang/go/issues/63536
  root_path = function(fname)
    local root_files = {
      'go/go.mod', -- monorepo override so `root_path` is `./monorepo/go/**` not `./monorepo/**`
      'go.work',
      'go.mod',
      '.git',
    }

    -- return first parent dir that homes a found root_file
    return lspconfig.util.root_pattern(unpack(root_files))(fname) or lspconfig.util.path.dirname(fname)
  end,
})
```

### 해결

```log
/Users/rody/IdeaProjects/snippets/algorithms/examples/bygo
```

경로 이름을 바꿈

## someService'(타입 someService)을(를) 타입 some.Service(으)로 사용할 수 없습니다 'SomeFunction' 메서드가 포인터 리시버이므로 타입이 'some.Service'을(를) 구현하지 않습니다

### 문제

```log
someService'(타입 someService)을(를) 타입 some.Service(으)로 사용할 수 없습니다 'SomeFunction' 메서드가 포인터 리시버이므로 타입이 'some.Service'을(를) 구현하지 않습니다
```

### 원인

- Go 언어에서 이러한 에러 메시지는 일반적으로 인터페이스를 구현하는 구조체와 관련된 문제에서 발생
- 이 에러는 `someService`가 값 형태로 사용되었을 때 `SomeFunction` 메서드의 포인터 리시버와 일치하지 않기 때문에 발생
- 에러 메시지의 핵심은 `'someService'` 타입이 `some.Service` 인터페이스를 구현하려고 했으나, **인터페이스에 정의된 메서드 중 하나가 포인터 리시버를 사용하기 때문에 완전히 구현되지 않았다**는 것
- Go에서 인터페이스 구현은 메서드 집합을 기반으로 하는데, `값 리시버`와 `포인터 리시버`는 서로 다른 메서드 집합을 갖는다
    - 값 리시버(Value Receiver): 값 리시버를 사용하는 메서드는 **해당 타입의 값과 포인터 모두**에서 호출할 수 있다
    - 포인터 리시버(Pointer Receiver): 포인터 리시버를 사용하는 메서드는 **해당 타입의 포인터에서만** 호출할 수 있다

에러 메시지에 나타난 상황에서, `some.Service` 인터페이스에는 포인터 리시버를 사용하는 `SomeFunction` 메서드가 포함되어 있다
이 경우, 인터페이스를 올바르게 구현하려면, `someService` 타입의 변수를 포인터 형태로 사용해야 합니다.

예를 들어, 아래와 같은 상황을 가정해 보겠습니다:

```go
type Service interface {
    SomeFunction()
}

type someService struct {
    // 필드 정의...
}

func (s *someService) SomeFunction() {
    // 구현...
}
```

여기서 `someService` 타입은 `SomeFunction`을 포인터 리시버로 구현한다
따라서, `Service` 인터페이스를 구현하기 위해서는 `someService`의 인스턴스를 포인터로 사용해야 한다

```go
var s Service = &someService{} // 포인터 사용
s.SomeFunction() // 이제 올바르게 작동합니다.
```

값 형태로 `someService` 인스턴스를 사용하려고 하면, 인터페이스 구현이 올바르지 않다는 에러 메시지가 발생

```go
var s Service = someService{} // 잘못된 사용, 에러 발생
```

### 해결

- `someService` 인스턴스를 포인터 형태로 사용하는 것

## GoFmt returned non-zero code on some of the files. Would you like to commit anyway?

### 문제

`git rebase -i` 후에 작업 내용을 수정하고 commit 하려고 하니 경고 창이 뜬다

```log
GoFmt returned non-zero code on some of the files. Would you like to commit anyway?
```

그리고 IntelliJ commit 화면에 나오는 `Check failed`를 클릭하면 "GoFmt returned non-zero code on some of the files. Would you like to commit anyway?"라고 다시 나오고, 다이얼로그 창의 `Details` 버튼을 눌러보면 `Run` 탭에 아래처럼 나온다

```log
go: updates to go.mod needed; to update it:
    go mod tidy
```

### 원인

- Go 소스 코드를 포맷팅하는 도구인 `gofmt`가 일부 파일에서 오류를 발견했음을 의미
- 가능한 원인들
    - 포맷팅 오류
        - 일부 Go 파일이 Go의 공식 스타일 가이드를 따르지 않는 방식으로 작성되었을 수 있다
        - `gofmt`는 이러한 문제를 수정하려고 시도하지만, 때때로 수동으로 수정할 필요가 있을 수 있다
    - 구문 오류
        - 코드에 구문 오류가 있으면 `gofmt`가 제대로 실행되지 않을 수 있으므로, 이 경우 오류를 먼저 수정해야 한다
    - IDE 설정 문제
        - IntelliJ의 Go 플러그인 또는 관련 설정에 문제가 있을 수 있으므로, 이 경우, 설정을 확인하고 필요한 경우 재설정 필요

### 해결

일단 생각나는 가능한 해결 방법은...

1. 작업한 파일들을 각각 열어봐서, 코드에 표시된 오류나 경고를 확인하고, 필요한 경우 직접 수정
2. 직접 `gofmt` 명령 실행해서 오류가 있는 파일을 포맷팅할 수 있다

   ```shell
   # `yourfile.go`를 `gofmt` 스타일로 포맷팅
   gofmt -w yourfile.go
   ```

3. IntelliJ에서 전체 프로젝트에 대한 코드 포맷팅을 수행하여 모든 파일을 일관된 스타일로 포맷팅

근데 그래도 안 된다. 이게 실제로는 상관이 없는 거 같아서 그냥 commit... 하니까 잘 된다

## I don't know what to return because the method call was unexpected

### 문제

```log
=== RUN   Test_some_GetSomething
    mock.go:334:
        assert: mock: I don't know what to return because the method call was unexpected.
            Either do Mock.On("GetSomeRepo").Return(...) first, or remove the GetSomeRepo() call.
            This method was unexpected:
                GetSomeRepo()
```

위와 같은 에러가 발생. 근데 실제 로직 부분과

```go
someRepo, err = command.GetSomeRepo().GetSomethigFromDB(
    command.GetContext(),
    *something,
)
```

테스트 코드 부분을 보면 이상이 없어 보인다

```go
cmd := NewMockCommand(t)

// 테스트 코드
if tc.something != nil {
    mockRepo := mocks.NewRepository(t)
    mockRepo.
        EXPECT().
        GetSomethigFromDB(mock.Anything, *tc.something).
        Return(tc.expected, nil)
    cmd.EXPECT().GetSomeRepo().Return(mockRepo)
}

someStruct.DoMyLogic(cmd)
```

### 원인

- 실제로는 `GetSomeRepo` 자체가 문제가 아닌, `command.GetContext()` 부분이 문제였다.
- for loop를 돌면서 여러 테스트 케이스를 테스트하는데, 문제가 되는 한 건만 남기자 원인이 되는 부분이 제대로 남기 시작했다

```log
=== RUN   Test_some_GetSomething
    mock.go:334:
        assert: mock: I don't know what to return because the method call was unexpected.
            Either do Mock.On("GetContext").Return(...) first, or remove the GetContext() call.
            This method was unexpected:
                GetContext()
```

### 해결

아래와 같이 수정

```go
if tc.something != nil {
    ctx := context.Background()
    mockRepo := mocks.NewRepository(t)
    mockRepo.
        EXPECT().
        GetSomethigFromDB(ctx, *tc.something).
        Return(tc.expected, nil)
    cmd.EXPECT().GetContext().Return(ctx)
    cmd.EXPECT().GetForensicRepository().Return(mockRepo)
}
```

## go: cannot find main module, but found .git/config in /some/path/to

### 문제

```log
Starting: /Users/rody/go/bin/dlv dap --listen=127.0.0.1:52424 --log-dest=3 from /Users/rody/IdeaProjects/snippets/go/learning/1_basic
DAP server listening at: 127.0.0.1:52424
Build Error: go build -o /Users/rody/IdeaProjects/snippets/go/learning/1_basic/__debug_bin1816871114 -gcflags all=-N -l .
go: cannot find main module, but found .git/config in /Users/rody/IdeaProjects/snippets
    to create a module there, run:
    cd ../../.. && go mod init (exit status 1)
```

### 원인

- vscode에서 실행할 때 아래와 같은 명령어가 실행된다

```bash
/Users/rody/go/bin/dlv dap --listen=127.0.0.1:54661 --log-dest=3 from /Users/rody/IdeaProjects/snippets/go/learning/1_basic
```

근데 1.16 이후에는 `go.mod` 사용이 기본이 되었다. 하지만 현재 실행하려는 파일은 단순히 한 개의 실행 파일이어서 `go.mod`가 없고, 없는 상태에서 위 명령어 실행되니 에러 발생

### 해결

단순 파일 실행만 가능하도록 모듈 모드를 끈다.
아래는 `실행 및 디버그` 통해서 추가한 프로젝트 루트 경로의 `.vscode/launch.json` 파일

```json
{
    // IntelliSense를 사용하여 가능한 특성에 대해 알아보세요.
    // 기존 특성에 대한 설명을 보려면 가리킵니다.
    // 자세한 내용을 보려면 https://go.microsoft.com/fwlink/?linkid=830387을(를) 방문하세요.
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch file",
            "type": "go",
            "request": "launch",
            "mode": "debug", // 또는 "auto"
            "env": {
                "GO111MODULE": "off" // GOPATH 모드 강제하여 go.mod 을 굳이 추가하지 않아도 실행되도록 한다
            },
            "program": "${file}" // 또는 "${fileDirname}"
        }
    ]
}
```

## could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")"

### 문제

interface를 private repo에 올라가 있는 경로로 변경했더니, 아래와 같이 CI 단계에서 mockery로 generate 시에 에러 발생.

require 부분을 바꿧는데, 왜 mocker generate가 실패하는지 이해할 수 없지만...

```log
18 Dec 23 08:57 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:57 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:57 UTC INF Walking dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/aggregator.go:6:2: could not import github.com/gofiber/fiber/v2 (invalid package name: \"\")" dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC ERR Unable to find 'Repository' in any go files under this path dry-run=false version=v2.27.1
Error: unable to find interface
Usage:
  mockery [flags]
  mockery [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  showconfig  Show the yaml config

Flags:
      --all                        generates mocks for all found interfaces in all sub-directories
      --boilerplate-file string    File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */
      --case string                name the mocked file using casing convention [camel, snake, underscore] (default "camel")
      --config string              config file to use
      --cpuprofile string          write cpu profile to file
      --dir string                 directory to search for interfaces
      --disable-version-string     Do not insert the version string into the generated mock file.
  -d, --dry-run                    Do a dry run, don't modify any files
      --exclude stringArray        prefixes of subdirectories and files to exclude from search
      --exported                   Generates public mocks for private interfaces.
      --filename string            name of generated file (only works with -name and no regex)
  -h, --help                       help for mockery
      --inpackage                  generate a mock that goes inside the original package
      --inpackage-suffix           use filename '_mock' suffix instead of 'mock_' prefix for InPackage mocks
      --keeptree                   keep the tree structure of the original interface files into a different repository. Must be used with XX
      --log-level string           Level of logging (default "info")
      --name string                name or matching regular expression of interface to generate mock for
      --note string                comment to insert into prologue of each generated file
      --outpkg string              name of generated package (default "mocks")
      --output string              directory to write mocks to (default "./mocks")
      --packageprefix string       prefix for the generated package name, it is ignored if outpkg is also specified.
      --print                      print the generated mock to stdout
      --quiet                      suppresses logger output (equivalent to --log-level="")
  -r, --recursive                  recurse search into sub-directories
      --replace-type stringArray   Replace types
      --srcpkg string              source pkg to search for interfaces
      --structname string          name of generated struct (only works with -name and no regex)
      --tags string                space-separated list of additional build tags to use
      --testonly                   generate a mock in a _test.go file
      --unroll-variadic            For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is. (default true)
      --version                    prints the installed version of mockery
      --with-expecter              Generate expecter utility around mock's On, Run and Return methods with explicit types. This option is NOT compatible with -unroll-variadic=false

Use "mockery [command] --help" for more information about a command.

payments/repository.go:15: running "mockery": exit status 1
18 Dec 23 08:57 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:57 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:57 UTC INF Walking dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF Generating mock dry-run=false interface=Repository qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/payments/extension version=v2.27.1
18 Dec 23 08:57 UTC INF writing mock to file dry-run=false interface=Repository qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/payments/extension version=v2.27.1
18 Dec 23 08:57 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:57 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:57 UTC INF Walking dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF Generating mock dry-run=false interface=Repository qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/payments/forensic version=v2.27.1
18 Dec 23 08:57 UTC INF writing mock to file dry-run=false interface=Repository qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/payments/forensic version=v2.27.1
18 Dec 23 08:57 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:57 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:57 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:57 UTC INF Walking dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:57 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:58 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:58 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
Error: 18 Dec 23 08:58 UTC ERR Error parsing file error="/home/runner/work/go/go/some-qwerty-api-service/payments/receipt/danal_tpay.go:12:2: could not import go.uber.org/zap (invalid package name: \"\")" dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC ERR Unable to find 'Command' in any go files under this path dry-run=false version=v2.27.1
Error: unable to find interface
Usage:
  mockery [flags]
  mockery [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  showconfig  Show the yaml config

Flags:
      --all                        generates mocks for all found interfaces in all sub-directories
      --boilerplate-file string    File to read a boilerplate text from. Text should be a go block comment, i.e. /* ... */
      --case string                name the mocked file using casing convention [camel, snake, underscore] (default "camel")
      --config string              config file to use
      --cpuprofile string          write cpu profile to file
      --dir string                 directory to search for interfaces
      --disable-version-string     Do not insert the version string into the generated mock file.
  -d, --dry-run                    Do a dry run, don't modify any files
      --exclude stringArray        prefixes of subdirectories and files to exclude from search
      --exported                   Generates public mocks for private interfaces.
      --filename string            name of generated file (only works with -name and no regex)
  -h, --help                       help for mockery
      --inpackage                  generate a mock that goes inside the original package
      --inpackage-suffix           use filename '_mock' suffix instead of 'mock_' prefix for InPackage mocks
      --keeptree                   keep the tree structure of the original interface files into a different repository. Must be used with XX
      --log-level string           Level of logging (default "info")
      --name string                name or matching regular expression of interface to generate mock for
      --note string                comment to insert into prologue of each generated file
      --outpkg string              name of generated package (default "mocks")
      --output string              directory to write mocks to (default "./mocks")
      --packageprefix string       prefix for the generated package name, it is ignored if outpkg is also specified.
      --print                      print the generated mock to stdout
      --quiet                      suppresses logger output (equivalent to --log-level="")
  -r, --recursive                  recurse search into sub-directories
      --replace-type stringArray   Replace types
      --srcpkg string              source pkg to search for interfaces
      --structname string          name of generated struct (only works with -name and no regex)
      --tags string                space-separated list of additional build tags to use
      --testonly                   generate a mock in a _test.go file
      --unroll-variadic            For functions with variadic arguments, do not unroll the arguments into the underlying testify call. Instead, pass variadic slice as-is. (default true)
      --version                    prints the installed version of mockery
      --with-expecter              Generate expecter utility around mock's On, Run and Return methods with explicit types. This option is NOT compatible with -unroll-variadic=false

Use "mockery [command] --help" for more information about a command.

payments/receipt/command.go:17: running "mockery": exit status 1
18 Dec 23 08:58 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:58 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:58 UTC INF Walking dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Generating mock dry-run=false interface=Repository qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/payments/sbcr version=v2.27.1
18 Dec 23 08:58 UTC INF writing mock to file dry-run=false interface=Repository qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/payments/sbcr version=v2.27.1
18 Dec 23 08:58 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:58 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:58 UTC INF Walking dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Generating mock dry-run=false interface=RepositoryV1 qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/users version=v2.27.1
18 Dec 23 08:58 UTC INF writing mock to file dry-run=false interface=RepositoryV1 qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/users version=v2.27.1
18 Dec 23 08:58 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:58 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:58 UTC INF Walking dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Generating mock dry-run=false interface=RepositoryV2 qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/users version=v2.27.1
18 Dec 23 08:58 UTC INF writing mock to file dry-run=false interface=RepositoryV2 qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/users version=v2.27.1
18 Dec 23 08:58 UTC INF couldn't read any config file version=v2.27.1
18 Dec 23 08:58 UTC INF Starting mockery dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Using config:  dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF DISCUSSION: dynamic walking of project is being considered for removal in v3. Please provide your feedback at the linked discussion. discussion=https://github.com/vektra/mockery/discussions/549 dry-run=false pr=https://github.com/vektra/mockery/pull/548 version=v2.27.1
18 Dec 23 08:58 UTC INF Walking dry-run=false version=v2.27.1
18 Dec 23 08:58 UTC INF Generating mock dry-run=false interface=Service qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/users version=v2.27.1
18 Dec 23 08:58 UTC INF writing mock to file dry-run=false interface=Service qualified-name=github.com/some-qwerty-org.io/some-qwerty-api-service/users version=v2.27.1
Error: Process completed with exit code 1.
```

### 원인

- `GOPRIVATE` 환경 변수
    - Go 모듈 시스템에서 사용되며, 특정한 패키지 경로들을 비공개로 취급할 때 사용
    - 비공개 또는 내부 저장소에 호스팅되는 Go 패키지들에 대해 Go 도구가 이들을 `공개 모듈 프록시`나 `체크섬 데이터베이스`에 요청하지 않도록 지시한다
    - 외부 공개 저장소에 존재하지 않으므로, Go 도구는 이들을 직접적으로 해당 소스 위치에서 가져와야 한다

`GOPRIVATE`이 `mockery`에서 발생한 문제와 관련이 있는 이유는 다음과 같습니다:

1. 비공개 리포지토리 접근:
    - 만약 `mockery`가 내부적으로 Go의 모듈 시스템을 사용하여 패키지의 종속성을 해결하려고 할 때, `GOPRIVATE` 환경 변수가 설정되지 않은 경우, 비공개 저장소에 있는 모듈들에 대한 접근이 실패할 수 있다
    - 이는 Go 도구가 기본적으로 공개 저장소를 통해 모듈을 검색하기 때문
2. 모듈 검색 경로의 제한
    - `GOPRIVATE` 환경 변수가 설정되지 않았을 때, Go 도구는 내부 또는 비공개 저장소에 있는 패키지를 공개적으로 사용할 수 있는 모듈로 오인하고, 이로 인해 해당 모듈을 찾지 못하는 문제가 발생할 수 있다
3. CI 환경과의 차이
    - 로컬 개발 환경에서는 `GOPRIVATE` 설정이 이미 되어 있을 수 있지만, CI 환경에서는 이 설정이 누락되어 있을 수 있다

[`Error parsing file` 로그가 찍히는 부분](https://github.com/vektra/mockery/blob/446e0bf4ccb760917134bd6fe9d147cbad0b346f/pkg/walker.go#L106-L110C4)을 보면 `.go` 파일을 파싱하다가 발생하는 에러로 보인다

```go
// https://github.com/vektra/mockery/blob/master/pkg/parse.go#L61
func (p *Parser) loadPackages(fpath string) ([]*packages.Package, error) {
    if result, ok := p.packageLoadCache[filepath.Dir(fpath)]; ok {
        return result.pkgs, result.err
    }
    pkgs, err := packages.Load(&p.conf, "file="+fpath)
    p.packageLoadCache[fpath] = packageLoadEntry{pkgs, err}
    return pkgs, err
}
```

### 해결

- 원인은 잘 모르겠고, `go env -w GOPRIVATE="github.com/org/*"`를 설정하니 해결이 됐다. 결과를 통해 원인을 추론해봐야 할 듯...

```yaml
- name: Generate Mocks
  run: |
    go env -w GOPRIVATE="github.com/org/*"
    wget https://github.com/vektra/mockery/releases/download/v2.38.0/mockery_2.38.0_Linux_x86_64.tar.gz -q -O mockery.tar.gz
    tar -xf mockery.tar.gz && mv mockery /usr/local/bin/mockery
    (cd some-qwerty-api-service && mockery)
```

## fatal: could not read Username for `https://github.com`: terminal prompts disabled

### 문제

```log
Error: some-qwerty-api-service/payments/transform.go:8:2: github.com/some-qwerty-org.io/go/interface/channel@v1.31.1: reading github.com/some-qwerty-org.io/go/interface/channel/go.mod at revision interface/channel/v1.31.1: git ls-remote -q origin in /home/runner/go/pkg/mod/cache/vcs/9580286692b696caea77f128ff1e921fc3e09776eb8c476ca5420e9ad1e951b1: exit status 128:
    fatal: could not read Username for 'https://github.com': terminal prompts disabled
Confirm the import path was entered correctly.
If this is a private repository, see https://golang.org/doc/faq#git_https for additional information.
... 반복 ...
Error: Process completed with exit code 1.
```

### 원인

GitHub Actions 에서 테스트 실행시 private repo에 접근하는데, 말 그대로 인증 정보가 추가되어 있지 않기 때문

### 해결

- `PAT`를 사용하도록 한다

```yaml
- name: Go Test
  run: |
    git config --global url.https://${{ secrets.SOME_PAT }}@github.com/.insteadOf https://github.com/
    go test ./some-qwerty-api-service/... -cover .
```

## `before_org` -> `after_org`로 바꾸는 작업 후 에러 발생

### 문제

`before_org` -> `after_org`로 바꾸는 작업을 했는데, 아래와 같은 에러가 발생

```log
╰─ go mod tidy
go: finding module for package github.com/before_org/internal/boot
go: finding module for package github.com/before_org/internal/transformer
go: finding module for package github.com/before_org/internal/client/elasticsearch
go: finding module for package github.com/before_org/some_service_interface/generated/store
go: finding module for package github.com/before_org/internal/env
go: finding module for package github.com/before_org/internal/card
go: finding module for package github.com/before_org/internal/encryption
go: finding module for package github.com/before_org/internal/log
github.com/before_org/some-qwerty-api-service imports
        github.com/before_org/internal/boot: cannot find module providing package github.com/before_org/internal/boot: module github.com/before_org/internal/boot: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
github.com/before_org/some-qwerty-api-service imports
        github.com/before_org/internal/client/elasticsearch: cannot find module providing package github.com/before_org/internal/client/elasticsearch: module github.com/before_org/internal/client/elasticsearch: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
github.com/before_org/some-qwerty-api-service imports
        github.com/before_org/some_service_interface/generated/store: cannot find module providing package github.com/before_org/some_service_interface/generated/store: module github.com/before_org/some_service_interface/generated/store: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/a848c7a7d46e493ce8ce868c7a9754994eec331dc6d58d0f3177f6044025dcd8: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/some_service_interface/' not found
github.com/before_org/some-qwerty-api-service/cmd imports
        github.com/before_org/internal/env: cannot find module providing package github.com/before_org/internal/env: module github.com/before_org/internal/env: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
github.com/before_org/some-qwerty-api-service/payments imports
        github.com/before_org/internal/card: cannot find module providing package github.com/before_org/internal/card: module github.com/before_org/internal/card: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
github.com/before_org/some-qwerty-api-service/payments imports
        github.com/before_org/internal/encryption: cannot find module providing package github.com/before_org/internal/encryption: module github.com/before_org/internal/encryption: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
github.com/before_org/some-qwerty-api-service/payments imports
        github.com/before_org/internal/log: cannot find module providing package github.com/before_org/internal/log: module github.com/before_org/internal/log: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
github.com/before_org/some-qwerty-api-service/payments imports
        github.com/before_org/internal/transformer: cannot find module providing package github.com/before_org/internal/transformer: module github.com/before_org/internal/transformer: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/6afcf8c0b357556a61263df65d781e1a858afb85cc18b350c3fdaf5c10829cac: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/before_org/internal/' not found
```

이때 캐시를 지우고 다시 실행을 해본다

```shell
go clean -modcache
```

그리고 다시 `go mod tidy`를 하면

```log
╰─ go build
go: downloading github.com/gofiber/fiber/v2 v2.49.2
... 생략 ...

<--- 근데 여기서 모듈명만 있고, 실제 리파지토리가 없는 것을 확인해보려고 한다 --->
package github.com/after_org/some-qwerty-api-service
        boot.go:9:2: use of internal package github.com/before_org/internal/boot not allowed
package github.com/after_org/some-qwerty-api-service
        imports github.com/after_org/some-qwerty-api-service/payments
        payments/legacy.go:6:2: use of internal package github.com/before_org/internal/card not allowed
package github.com/after_org/some-qwerty-api-service
        boot.go:10:2: use of internal package github.com/before_org/internal/client/elasticsearch not allowed
package github.com/after_org/some-qwerty-api-service
        imports github.com/after_org/some-qwerty-api-service/payments
        payments/legacy.go:7:2: use of internal package github.com/before_org/internal/transformer not allowed
../interfaces/merchant-service/generated/store/store_service.pb.go:10:2: no required module provides package github.com/before_org/some_service_interface/generated/business_license; to add it:
        cd /Users/rody/IdeaProjects/some-qwerty-org.io-go/interfaces/merchant-service
        go get github.com/before_org/some_service_interface/generated/business_license
../interfaces/merchant-service/generated/store/store_service.pb.go:11:2: no required module provides package github.com/before_org/some_service_interface/generated/error; to add it:
        cd /Users/rody/IdeaProjects/some-qwerty-org.io-go/interfaces/merchant-service
        go get github.com/before_org/some_service_interface/generated/error
boot.go:11:2: no required module provides package github.com/before_org/some_service_interface/generated/store; to add it:
        go get github.com/before_org/some_service_interface/generated/store
../interfaces/merchant-service/generated/store/store_service.pb.go:12:2: no required module provides package github.com/before_org/some_service_interface/generated/term; to add it:
        cd /Users/rody/IdeaProjects/some-qwerty-org.io-go/interfaces/merchant-service
        go get github.com/before_org/some_service_interface/generated/term
```

### 원인

가령 `package github.com/after_org/some-qwerty-api-service`에 대해서 `imports github.com/after_org/some-qwerty-api-service/payments`로 뭔가 서브 패키지를 import 할 때, `use of internal package github.com/before_org/internal/transformer not allowed`라는 에러가 발생하는데, 이때 보면 이전 조직의 모듈명을 찾으려고 한다.

왜 이럴까? Go는 **로컬에 해당 모듈이 없다고 판단하면 원격 리포지토리로 접근하려고 시도**한다. 확인을 해보니, 아직 `github.com/before_org`을 사용하고 있던 부분들이 있었다

### 해결

이전 조직 부분을 모두 다시 수정한 후에도 에러 발생

```log
github.com/some-qwerty-org.io/some-qwerty-api-service imports
        github.com/some-qwerty-org.io/some_service_interface/generated/store imports
        github.com/some-org/some_service_interface/generated/business_license: cannot find module providing package github.com/some-org/some_service_interface/generated/business_license: module github.com/some-org/some_service_interface/generated/business_license: git ls-remote -q origin in /Users/rody/go/pkg/mod/cache/vcs/a848c7a7d46e493ce8ce868c7a9754994eec331dc6d58d0f3177f6044025dcd8: exit status 128:
        remote: Repository not found.
        fatal: repository 'https://github.com/some-org/some_service_interface/' not found
```

`github.com/before_org/some_service_interface/generated/business_license`를 import 하려고 하는데, `some_service_interface`도 조직이 바뀐 리파지토리를 사용하지 않아서 그렇다.

```shell
# 이미 생성된 디렉토리 지우고
rm -rf generated
# 다시 protobuf를 생성
docker run --rm --volume "$(pwd):/protobuf" --workdir /protobuf bufbuild/buf generate
```

## GOPROXY list is not the empty string, but contains no entries

### 문제

```log
... 생략 ...
go: downloading google.golang.org/grpc v1.60.0
go: downloading github.com/google/uuid v1.3.1
go: downloading google.golang.org/protobuf v1.31.0
go: GOPROXY list is not the empty string, but contains no entries
go: GOPROXY list is not the empty string, but contains no entries
go: GOPROXY list is not the empty string, but contains no entries
go: GOPROXY list is not the empty string, but contains no entries
... 반복 ...
```

### 원인

Go의 모듈 프록시 설정과 관련된 문제를 나타낸다.
이 에러는 주로 `GOPROXY` 환경 변수가 설정되었으나, 올바른 모듈 프록시 서버의 URL이 포함되지 않았을 때 발생한다.

Go 1.13 버전부터는 Go 모듈 시스템이 도입되었고, `GOPROXY` 환경 변수를 사용하여 모듈 다운로드를 위한 프록시 서버를 지정할 수 있게 되었다. 이 환경 변수가 설정되면, `go get`과 같은 명령어를 사용할 때 지정된 프록시 서버를 통해 필요한 Go 모듈을 다운로드 한다.

이러한 문제를 방지하기 위해, Go 프로젝트를 진행할 때 다음 사항을 확인하는 것이 좋다.

- `GOPROXY` 환경 변수가 올바르게 설정되어 있는지 확인.
- GVM과 같은 버전 관리 도구를 사용할 때 환경 변수가 올바르게 설정되었는지 확인.
- 필요한 경우 `go.env` 파일을 참조하여 환경 변수를 조정.

#### 주요 원인

1. 잘못된 `GOPROXY` 설정: `GOPROXY` 환경 변수가 설정되었으나, 유효한 URL이 포함되지 않았을 경우, Go는 모듈을 다운로드할 프록시 서버를 찾지 못하고 에러를 발생시킨다.

2. 환경 변수의 위치: Go는 환경 변수 설정을 위해 여러 위치를 참조합니다. 보통 `GOPROXY`는 사용자의 쉘 환경 설정(`.bashrc`, `.zshrc` 등) 또는 `go.env` 파일에서 설정됩니다. `GOROOT` 환경 변수가 지정한 디렉토리 안의 `go.env` 파일에서 `GOPROXY` 설정을 읽을 수도 있습니다. 이 파일에서 올바른 프록시 설정을 하지 않으면 에러가 발생할 수 있습니다.

3. GVM 사용: Go Version Manager(GVM)을 사용할 때, 다른 Go 버전 간 전환 시 환경 변수 설정이 변경될 수 있다. 이 과정에서 `GOPROXY` 설정이 올바르게 유지되지 않을 수 있어, 관련 에러가 발생할 수 있다

### 해결

1. [Go 공식 문서](https://cs.opensource.google/go/go/+/master:go.env)에서 제공하는 `go.env` 파일 내용을 참조하여 `GOPROXY` 환경 변수를 올바르게 설정한다. `go.env` 파일에 정의된 내용은 Go 환경 설정에 직접 영향을 미친다. 이 파일에 올바른 프록시 설정을 추가함으로써 `go get` 명령어가 필요한 모듈을 프록시 서버를 통해 성공적으로 다운로드할 수 있게 된다.
2. `go env -w GOPROXY="https://proxy.golang.org,direct"` 명령어를 사용하여 `GOPROXY` 환경 변수를 설정할 수 있다. 이 명령어는 현재 쉘 세션에만 영향을 미치며, 쉘을 종료하면 설정이 사라진다. 이 방법은 임시적으로 `GOPROXY` 환경 변수를 설정할 때 유용하다.

## gvm_implode:read:1: -p: no coprocess

### 문제

```log
❯ gvm implode -p
gvm_implode:read:1: -p: no coprocess

Action cancelled
```

### 원인

[트러블슈팅 섹션](https://github.com/moovweb/gvm?tab=readme-ov-file#troubleshooting)에 아래와 같은 내용이 있다

> Sometimes especially during upgrades the state of gvm's files can get mixed up. This is mostly true for upgrade from older version than 0.0.8. Changes are slowing down and a LTR is imminent. But for now `rm -rf ~/.gvm` will always remove gvm. Stay tuned!
>
> 때때로, 특히 업그레이드하는 과정에서 GVM의 파일 상태가 혼란스러워질 수 있습니다. 이는 주로 0.0.8 버전 이전에서 업그레이드할 때 발생합니다. 변경 사항은 점차 줄어들고 있으며 장기 지원 버전(LTR)의 출시가 임박했습니다. 하지만 현재로서는 `rm -rf ~/.gvm` 명령을 사용하면 항상 GVM을 제거할 수 있습니다. 계속해서 관심을 가져주시기 바랍니다!

- 파일 및 구성 상태 불일치: 업그레이드 과정에서, GVM의 내부 파일이나 구성이 기대하는 상태와 다르게 변경될 수 있습니다. 이는 새 버전의 GVM이 이전 설정이나 파일 구조와 완전히 호환되지 않을 때 발생할 수 있습니다.
- 업그레이드 과정의 복잡성: 특히 이전 버전에서 새로운 버전으로의 전환은 내부 구조의 변경을 동반할 수 있으며, 이 과정에서 파일이 올바르게 업데이트되지 않거나 구성이 손상될 수 있습니다.

### 해결

**수동 제거**: `gvm implode` 명령이 예상대로 작동하지 않는 경우, GVM이 설치된 디렉토리(일반적으로는 `$HOME/.gvm`)를 수동으로 삭제하여 GVM을 제거할 수 있습니다.

```sh
rm -rf $HOME/.gvm
```

## could not import slices (current file is not included in a workspace module)

### 문제

```go
import (
    "fmt"
    "slices"
    ^^^^^^^^ could not import slices (current file is not included in a workspace module)
    "sort"
    "strings"
)
```

### 원인

[이 사이트](https://www.gopherguides.com/articles/golang-slices-package)에 따르면 1.21부터 표준 라이브러리에 추가되었다고 한다.

> In release 1.21, the slices package will be officially added to the standard library. It includes many useful functions for sorting, managing, and searching slices. In this article, we will cover the more commonly used functions included in the Slices package.

### 해결

```bash
go get golang.org/x/exp/slices
```

`$GOPATH/src` 아래에 패키지가 다운로드되고, 프로젝트에서 해당 패키지를 임포트할 수 있게 된다.

## compile: version 'go1.20.3' does not match go tool version 'go1.22.2'

### 문제

```log
compile: version 'go1.20.3' does not match go tool version 'go1.22.2'
```

### 원인

`go` 도구의 버전이 컴파일하려는 코드와 일치하지 않을 때 발생합니다.

이 문제는 일반적으로 Go 버전 업그레이드 후 발생할 수 있으며, 특히 Go 모듈 파일(`go.mod`)의 `go` 버전 지시문이 현재 시스템의 Go 버전과 일치하지 않을 때 나타납니다.

Go 프로젝트에서 사용하는 Go 버전은 `go.mod` 파일에 명시됩니다. 따라서 `go.mod` 파일의 첫 번째 줄을 확인하여 현재 시스템에 설치된 Go 버전과 일치하는지 확인해야 합니다. 예를 들어, `go1.22`로 변경해야 합니다.

### 해결

문제 해결을 위해 다음 단계를 시도해 볼 수 있습니다:

1. `go.mod` 파일에서 `go` 버전을 현재 설치된 버전(`go1.22`)으로 업데이트합니다.

   ```go
   module my/module

   go 1.22
   ```

2. 변경 사항을 적용하려면 터미널에서 다음 명령을 실행합니다:

   ```bash
   go mod tidy
   ```

   이 명령은 모듈의 의존성을 정리하고 `go.mod` 및 `go.sum` 파일을 최신 상태로 유지합니다.

3. 이후에도 문제가 지속되면 Go 버전을 명시적으로 설정할 수 있는 환경 변수를 확인하고 필요에 따라 조정합니다. 예를 들어, 다음 명령을 사용하여 현재 세션에 대해 Go 버전을 설정할 수 있습니다:

   ```bash
   export GOROOT=/usr/local/go1.22
   export PATH=$GOROOT/bin:$PATH
   ```

4. 모든 변경 후에는 프로젝트를 재컴파일하여 문제가 해결되었는지 확인합니다.

   ```bash
   go build
   ```

이러한 단계를 따르면 일반적으로 버전 불일치 문제가 해결됩니다. 추가로 도움이 필요하다면 실제 파일을 수정하고 컴파일 과정을 다시 진행해 봐야 할 수도 있습니다.

## failed to decode elasticsearch response: EOF

### 문제

```go
type Response struct {
    StatusCode int
    Header     http.Header
    Body       io.ReadCloser
}
```

response body를 한번 읽고 난 후에 아래와 같이 에러가 발생

```go
body, _ := io.ReadAll(res.Body)
fmt.Println(string(body))
```

```bash
"GetPaymentsByStatus.GetESPaymentsToCheckEligible: GetESPaymentsToCheckEligible.GetDocuments: GetDocuments: GetDocuments: failed to decode elasticsearch response: EOF"
```

### 원인

`io.ReadAll` 함수를 사용하여 `res.Body`를 읽은 후에는 `res.Body`의 내용이 소진되어 더 이상 읽을 수 없게 됩니다.

`res.Body`는 스트림 형태의 데이터를 제공하며, 한 번 읽으면 그 위치가 마지막까지 이동하기 때문에 다시 처음부터 읽으려면 리와인드(rewind)를 해야 합니다.

### 해결

#### `res.Body`의 내용을 바이트 슬라이스에 저장 후 재사용

하지만, `http.Response`의 `Body`는 `io.ReadCloser` 인터페이스를 구현하고 있어서, 일반적으로 스트림을 리와인드하는 기능을 지원하지 않습니다. 따라서, `Body`를 다시 읽기 위해서는 내용을 메모리에 저장해두고 필요할 때마다 그 저장된 데이터를 사용해야 합니다.

예를 들어, `res.Body`의 내용을 바이트 슬라이스에 저장한 후, 이를 필요할 때마다 읽는 방법은 다음과 같습니다:

```go
body, err := io.ReadAll(res.Body)
if err != nil {
    log.Fatal(err)
}
res.Body.Close() // Body를 읽은 후에는 반드시 닫아주어야 합니다.

// body 변수에 데이터가 저장되어 있으므로, 이후에는 body를 재사용할 수 있습니다.
fmt.Println(string(body))
```

이 방법을 사용하면 `res.Body`를 직접 다시 읽지 않고도 필요한 데이터에 접근할 수 있습니다. 데이터를 여러 번 처리해야 할 경우, `body` 슬라이스를 사용하여 필요한 작업을 수행하면 됩니다.

#### `res.Body` 자체를 복사하는 방법

아니면 `res.Body`를 복사합니다. 가장 일반적인 방법 중 하나는 `bytes.Buffer`를 사용하여 HTTP 응답을 버퍼에 쓰는 것입니다. 이 방법을 사용하면 응답 본문을 메모리에 저장하고 필요할 때마다 이를 재사용할 수 있습니다.

다음은 `http.Response` 객체의 `Body`를 `bytes.Buffer`에 복사하는 방법을 보여주는 코드 예제입니다:

```go
import (
    "bytes"
    "io"
    "net/http"
)

func copyResponseBody(res *http.Response) (*bytes.Buffer, error) {
    var b bytes.Buffer
    // res.Body를 bytes.Buffer에 복사
    if _, err := io.Copy(&b, res.Body); err != nil {
        return nil, err
    }
    res.Body.Close() // 복사 후에는 Body를 닫아줍니다.

    return &b, nil
}

// 사용 예
resp, err := http.Get("http://example.com")
if err != nil {
    // 에러 처리
}
buf, err := copyResponseBody(resp)
if err != nil {
    // 에러 처리
}

// buf.Bytes()를 사용하여 응답 데이터에 접근
```

위 코드에서는 `io.Copy` 함수를 사용하여 `res.Body`의 내용을 `bytes.Buffer`에 복사합니다. 이렇게 하면 원본 `res.Body`를 닫고 나서도 `bytes.Buffer`에 저장된 데이터를 사용할 수 있습니다. 이 방법은 `res.Body`가 스트림으로부터 모든 데이터를 읽어서 버퍼에 저장하기 때문에, 데이터 크기가 클 경우 메모리 사용에 주의해야 합니다.

이 방식을 통해 `res.Body`의 데이터를 성공적으로 복사하고 재사용할 수 있으며, HTTP 요청과 응답을 처리하는 데 있어 더 큰 유연성을 제공합니다.

#### `httputil.DumpResponse` 함수를 사용하는 방법

더 간단한 방법으로는 `httputil.DumpResponse` 함수를 사용하는 것이 있습니다. 이 함수는 HTTP 응답을 복사하고, 복사된 응답의 바디를 다시 사용할 수 있도록 메모리에 저장합니다. 이 방법은 응답의 헤더와 바디를 모두 복사하므로, 응답을 로깅하거나 디버깅할 때 유용합니다.

다음은 `httputil.DumpResponse`를 사용하여 `res.Body`를 복사하고 재사용하는 예제 코드입니다:

```go
import (
    "net/http"
    "net/http/httputil"
    "bytes"
)

func copyResponse(res *http.Response) (*http.Response, error) {
    dump, err := httputil.DumpResponse(res, true)
    if err != nil {
        return nil, err
    }

    // dump에서 새로운 응답 생성
    newRes := new(http.Response)
    if err := newRes.ReadFrom(bytes.NewReader(dump)); err != nil {
        return nil, err
    }

    return newRes, nil
}

// 사용 예
resp, err := http.Get("http://example.com")
if err != nil {
    // 에러 처리
}
newResp, err := copyResponse(resp)
if err != nil {
    // 에러 처리
}

// newResp.Body를 사용하여 필요한 작업 수행
```

이 코드는 원본 응답을 `httputil.DumpResponse`를 통해 복사하고, 복사된 데이터로부터 새로운 `http.Response` 객체를 생성합니다. 이렇게 하면 원본 응답의 `Body`를 소진하지 않고도 내용을 재사용할 수 있습니다.
