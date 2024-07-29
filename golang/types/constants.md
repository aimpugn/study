# constants

## 컨벤션

Go 언어에서 상수(constants)를 선언할 때의 컨벤션은 PHP와는 약간 다릅니다. Go에서는 상수 이름을 작성할 때 CamelCase를 사용하는 것이 일반적입니다. 이는 Go의 공식 코드 스타일 가이드인 [Effective Go](https://golang.org/doc/effective_go#mixed-caps)와 Go 커뮤니티에서 널리 받아들여지는 관례를 반영합니다.

### Go에서의 상수 선언 컨벤션

- **Exported Constants (외부로 공개되는 상수)**: 첫 글자를 대문자로 시작하여, 외부 패키지에서도 접근할 수 있게 합니다. CamelCase를 사용하여 단어 사이의 공백을 대문자로 구분합니다. 예: `MaxSize`, `ReportInterval`
- **Unexported Constants (내부 상수)**: 첫 글자를 소문자로 시작하여, 상수가 정의된 패키지 내부에서만 접근할 수 있게 합니다. CamelCase를 사용하여 단어 사이의 공백을 대문자로 구분합니다. 예: `maxCount`, `defaultTimeout`

### 예시

```go
package main

const (
    // Exported Constants
    MaxConnections = 100
    DefaultName    = "guest"

    // Unexported Constants
    initialBufferSize = 1024
)
```

### 대문자와 언더바 사용

Go에서는 상수 이름을 대문자와 언더바(`_`)로 구성하는 PHP나 다른 언어의 관습을 따르지 않습니다. Go의 관례는 간결하고 읽기 쉬운 코드를 선호하기 때문에, CamelCase를 사용하여 상수 이름을 구성합니다. 그러나, 특정 상황에서는 Go의 커뮤니티에서도 대문자와 언더바를 사용하는 경우가 있습니다. 예를 들어, 환경변수 이름이나, 일부 외부 시스템과의 인터페이스를 위해 이러한 스타일을 사용할 수 있습니다. 그러나 이는 상수 이름에 대한 일반적인 컨벤션이 아닙니다.

### 결론

Go 언어에서 상수를 선언할 때는 CamelCase를 사용하는 것이 일반적인 컨벤션입니다. 이는 코드의 일관성을 유지하고, Go 커뮤니티 내에서 널리 통용되는 스타일을 따르는 것입니다.
