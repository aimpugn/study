# compiler

- [compiler](#compiler)
    - [go compiler by golang](#go-compiler-by-golang)
    - [부트스트래핑(Bootstrapping)?](#부트스트래핑bootstrapping)
    - [Go 언어 컴파일러의 경우](#go-언어-컴파일러의-경우)
    - [구체적인 구현 과정 설명](#구체적인-구현-과정-설명)
        - [1단계: 초기 컴파일러 구현](#1단계-초기-컴파일러-구현)
        - [2단계: 자기 컴파일러 구현](#2단계-자기-컴파일러-구현)
        - [3단계: 반복적 개선과 최적화](#3단계-반복적-개선과-최적화)

## go compiler by golang

"부트스트래핑(Bootstrapping)"이라는 컴파일러 개발 방법 통해 golang의 컴파일러는 golang으로 작성되었다.

## 부트스트래핑(Bootstrapping)?

부트스트래핑은 컴파일러 또는 인터프리터를 그것이 지원하는 동일한 프로그래밍 언어로 작성하는 과정이다.
이 방법은 특히 새로운 프로그래밍 언어를 개발하거나 기존 언어의 컴파일러를 개선할 때 사용된다.

부트스트래핑 과정은 다음 단계로 진행된다:

1. **초기 구현**:
    - 가장 먼저, 타겟 언어의 최소한의 기능을 지원하는 기초 컴파일러를 다른 언어로 작성한다.
    - 예를 들어, Go 언어 컴파일러의 초기 버전은 C 언어로 작성되었을 수 있다.

2. **자기 컴파일**:
    - 초기 컴파일러를 사용하여 타겟 언어(이 경우 Go)로 작성된 더 완전하거나 개선된 컴파일러 버전을 컴파일한다.
    - 이 단계에서는 Go 언어로 작성된 컴파일러를 Go 언어로 컴파일합니다.

3. **반복 개선**:
    - 새로운 컴파일러를 사용하여 자기 자신을 다시 컴파일하고 개선하는 과정을 반복한다.
    - 이 과정을 통해 언어의 기능을 확장하고 성능을 향상시킬 수 있다.

## Go 언어 컴파일러의 경우

Go 언어 컴파일러는 이런 **부트스트래핑** 과정을 통해 개발되었다.
초기에는 다른 언어(예: C)로 작성된 컴파일러로 시작하여, Go 언어의 기능이 충분히 발전하면, Go 언어로 작성된 컴파일러로 자기 자신을 컴파일할 수 있게 된다. 이러한 방식으로, Go 컴파일러는 시간이 지남에 따라 Go 언어 자체로 구현되고 개선되었다.

이 과정은 자기 참조적으로 보일 수 있지만, 실제로는 프로그래밍 언어의 발전과 컴파일러 기술의 진화를 가능하게 하는 효과적인 방법이다. 부트스트래핑을 통해 언어 설계자들은 타겟 언어의 능력을 최대한 활용하여 컴파일러를 개발하고, 동시에 언어의 발전을 촉진할 수 있다.

## 구체적인 구현 과정 설명

### 1단계: 초기 컴파일러 구현

첫 단계에서는 다른 언어로 작성된 기본 컴파일러를 사용한다. 예를 들어, Go 언어의 초기 컴파일러는 Ken Thompson에 의해 Plan 9 C 언어로 작성되었다. 이 컴파일러는 Go 언어의 기본 구문과 기능을 처리할 수 있었다.

```c
// C 언어로 작성된 간단한 Go 컴파일러의 코드 예시(가상)
void compileGoCode(char* source) {
    // C 언어로 작성된 Go 코드 분석 및 실행
}
```

### 2단계: 자기 컴파일러 구현

Go 언어가 충분히 발전하고, 기본 컴파일러로 Go 언어의 기능을 충분히 지원할 수 있게 되면, Go 언어 자체를 사용하여 더 발전된 컴파일러를 작성할 수 있다. 이 단계에서는 Go 언어로 작성된 컴파일러로 Go 코드를 컴파일하게 된다.

```go
// Go 언어로 작성된 컴파일러의 간단한 코드 예시(가상)
package main

import "fmt"

func compileGoCode(source string) {
    // Go 언어로 작성된 컴파일러 로직
    fmt.Println("Compiling Go code...")
}

func main() {
    // Go 컴파일러 자기 자신을 컴파일하는 과정
    compileGoCode("some Go code")
}
```

### 3단계: 반복적 개선과 최적화

Go 언어로 작성된 컴파일러는 이후에도 지속적으로 자기 자신을 컴파일하고 개선하는 과정을 거친다.
이 과정을 통해 성능 최적화, 새로운 언어 기능의 지원, 버그 수정 등이 이루어진다.
이 단계에서는 Go 커뮤니티와 개발자들의 기여를 통해 컴파일러가 지속적으로 발전한다.

이 과정은 Go 언어의 발전과 함께 컴파일러 기술의 발전을 가능하게 한다.
Go 컴파일러가 Go로 작성되었다는 사실은 이 언어가 자체적인 발전을 통해 자신의 컴파일러를 개발하고 최적화할 수 있는 충분한 성숙도에 도달했음을 의미한다.

이러한 부트스트래핑 과정은 프로그래밍 언어의 성숙도와 자립성을 나타내는 중요한 이정표다. Go 언어 컴파일러의 경우, 이 과정을 통해 언어의 효율성, 안정성, 그리고 확장성이 지속적으로 향상되었다.
