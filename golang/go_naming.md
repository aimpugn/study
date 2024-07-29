# Go naming

- [Go naming](#go-naming)
    - [네이밍 컨벤션](#네이밍-컨벤션)
        - [초기어(initialism)?](#초기어initialism)
    - [기준](#기준)
    - [examples](#examples)
        - [`net/http`](#nethttp)
        - [stream types](#stream-types)
        - [statement와 `Stmt`](#statement와-stmt)
        - [`pg_tx_id`](#pg_tx_id)
        - [subscribe와 sbcr](#subscribe와-sbcr)
        - [근거](#근거)
        - [커뮤니티 의견](#커뮤니티-의견)

## 네이밍 컨벤션

Go의 네이밍 컨벤션에 따르면, 대문자와 소문자를 혼합한 카멜케이스(MixedCaps)를 사용하여 이름을 지정한다.
이 규칙은 약어나 초기어(acronyms)에도 적용되며, 이러한 경우 모든 문자를 대문자로 표기한다.
예를 들어,
- "HTTP Server"는 `HTTPServer`
- "ID Processor"는 `IDProcessor`

이는 Go의 공식 문서인 ["Effective Go"](https://go.dev/talks/2014/names.slide) 및 ["Go Code Review Comments"](https://github.com/golang/go/wiki/CodeReviewComments#initialisms)에서 확인할 수 있다.

참고 문서:
- [Effective Go: MixedCaps](https://go.dev/talks/2014/names.slide)
- [Go Code Review Comments: Initialisms](https://github.com/golang/go/wiki/CodeReviewComments#initialisms)

### 초기어(initialism)?

Go에서 "initialisms"는 특정 약어나 초기어를 모두 대문자로 표기하는 컨벤션을 의미한다.
초기어(initialism)란 각 단어의 첫 글자를 취하여 만든 약어를 말한다.
- "HTTP": "HyperText Transfer Protocol"의 초기어
- "URL": "Uniform Resource Locator"의 초기어

이러한 초기어는 Go의 네이밍 컨벤션에 따라 변수명, 함수명, 기타 식별자에서 모두 대문자로 표기된다.

Effective Go와 Go 코드 리뷰 코멘트 가이드라인에 따르면, 초기어는 변수명이나 함수명에서 첫 글자가 대문자인 경우(exported name) 모두 대문자로 써야 하며, 첫 글자가 소문자인 경우(un-exported name)에도 모두 소문자로 작성한다.
이 규칙의 일관된 적용은 코드의 가독성을 높이고, Go 커뮤니티 내에서 널리 통용되는 스타일을 유지하는 데 도움이 된다

- HTTPServerError (올바른 사용)
- ~~HttpServerError~~ (잘못된 사용)
- urlParser (올바른 사용)
- ~~urlparser~~ (잘못된 사용)

## 기준

> **초기문자어는 모두 대문자**

예를 들어, `HTTP Server`가 아닌 `http.Server`, `user ID`가 아닌 `userID`와 같이 사용한다.
이 규칙은 `subscribe_id`와 같은 경우에도 적용되어, `SubscribeID`가 적절한 형태가 된다.

> **널리 인식되고 사용되는 약어는 모두 대문자로 표기한다.**

예를 들어, "HTTP", "URL"과 같은 약어는 그 의미가 널리 알려져 있고 일반적으로 모두 대문자로 표기된다.
따라서 Go에서도 이러한 약어를 변수나 타입 이름에 사용할 때 모두 대문자로 표기한다. (`HTTPResponse`, `ParseURL` 등).

> **일반적으로 널리 인식되지 않는 약어나 초기어는 첫 글자만 대문자로 하여 표기하는 경우가 많다.**

이는 Go의 네이밍 컨벤션에서 유래한 것으로, 가독성과 일관성을 위해 이러한 방식을 채택한다.
예를 들어, "Statement"의 경우 "STMT"보다는 "Stmt"로 표기하는 것이 일반적이다.
이는 "Statement"가 "HTTP"나 "URL"처럼 널리 인식되는 약어는 아니기 때문이다.
따라서 "Stmt"와 같이 첫 글자만 대문자로 하는 표기법을 사용한다.

> **약어가 이름의 시작 부분에 오는 경우에도 모두 대문자**

예를 들어, `IDGenerator`라고 사용한다. `DBTransaction`이나 `IOBuffer`와 같이 두 글자 약어는 모두 대문자를 유지하는 것이 일반적이다.

## examples

### [`net/http`](https://golang.org/pkg/net/http/)

```go
const (
    MethodGet  = "GET"
    MethodHead = "HEAD"
    MethodPost = "POST"
    // ...
)

const (
    StatusContinue           = 100 // RFC 7231, 6.2.1
    StatusSwitchingProtocols = 101 // RFC 7231, 6.2.2
    StatusProcessing         = 102 // RFC 2518, 10.1

    StatusOK                 = 200 // RFC 7231, 6.3.1
    StatusCreated            = 201 // RFC 7231, 6.3.2
    // ...
)
```

### stream types

```go
// stream types
const (
    StreamMPEGDASH = iota
    StreamHLS
    StreamMPEGTSUDP
    StreamMPEGTSRTP
)
```

### statement와 `Stmt`

"Statement"와 같은 단어는 **널리 알려진 약어나 초기어**로 간주되지 않는다.
따라서, "Statement"의 약어인 `Stmt`에서는 첫 글자만 대문자로 표기하는 것이 일반적이다.
이는 Go의 일반적인 네이밍 컨벤션을 따르는 것으로, 특정 단어가 널리 알려진 약어나 초기어가 아닐 경우, 첫 글자만 대문자로 표기하는 것이 일반적인 규칙이다.

Go 언어에서는 네이밍 컨벤션을 매우 중요하게 여긴다. 이 컨벤션에는 변수, 상수, 타입, 함수 등의 이름을 지을 때 따라 룰이 있으며, 이 중 하나가 약어(acronyms)와 초기어(abbreviations)의 사용에 관한 것이다. Go에서는 다음과 같은 기준을 적용합니다:

### `pg_tx_id`

Go 언어에서의 일반적인 관례:
- 초기문자를 대문자로 하여 구조체의 필드를 공개(Public)로 만들고,
- CamelCase (낙타 표기법)를 사용하여 이름을 지정

`pg_tx_id`를 Go 구조체의 속성으로 만들 때, 이를 CamelCase로 변환하고, 약어에 대해서는 모두 대문자를 사용하는 것이 일반적인 관례입니다. 그러나, Go에서는 연속된 대문자 약어의 경우 첫 글자만 대문자로 하는 것이 더 선호됩니다. 이는 `Effective Go` 문서와 Go 커뮤니티에서 권장하는 스타일 가이드를 반영한 것입니다.

따라서, `pg_tx_id`에 대한 Go 스타일의 구조체 필드 이름은 다음과 같이 될 수 있습니다:

- `PgTxID`

이 이름은 다음과 같은 이유로 선택됩니다:

- `Pg`는 `payment gateway`의 약어로, 첫 글자만 대문자로 표기합니다.
- `Tx`는 `transaction`의 약어로, 첫 글자만 대문자로 표기합니다.
- `ID`는 일반적으로 모두 대문자로 표기되는 약어입니다.

이 방식은 Go의 컨벤션을 따르면서도, `pg_tx_id`의 의미를 명확하게 전달합니다. Go의 컨벤션을 따르는 것은 코드의 일관성을 유지하고, 다른 Go 개발자들이 코드를 더 쉽게 이해하고 협업할 수 있게 만듭니다.

### subscribe와 sbcr

그러므로, `subscribe_users`를 `sbcr_users`로 축약했다면, Go 네이밍 컨벤션에 따라 `SBCRUsers`와 `SBCRID`가 적절한 네이밍이 된다.

### 근거

이 규칙에 대한 명확한 근거는 Go의 공식 문서나 [Effective Go](https://golang.org/doc/effective_go#mixed-caps) 섹션에서 찾을 수 있습니다. Effective Go에서는 "MixedCaps" 또는 "mixedCaps" (CamelCase 또는 lowerCamelCase) 방식을 변수명, 함수명, 기타 식별자 이름에 사용할 것을 권장합니다. 그리고 초기문자어의 경우 모두 대문자로 표기하는 것을 권장하고 있습니다.

### 커뮤니티 의견

Go 커뮤니티 내에서도 이러한 규칙은 널리 받아들여지고 있으며, 많은 오픈 소스 프로젝트와 Go로 작성된 라이브러리에서 이러한 네이밍 컨벤션을 따르는 것을 볼 수 있습니다. 또한, Go 코드를 더 읽기 쉽고 유지 관리하기 쉽게 만드는 데 도움이 됩니다.

코드의 일관성과 명확성을 위해, 네이밍 컨벤션을 따르는 것은 매우 중요합니다. 따라서, 축약어를 사용할 때도 이러한 규칙을 고려하여, 코드 전체에서 일관된 방식을 유지하는 것이 좋습니다.
