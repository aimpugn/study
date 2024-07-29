# scope

## 변수 범위(scope)와 변수 선언

User

```go
var a *string

if a = command.GetCardIssuerCode(); a != nil && len(*a) == 4 {
   ^^^
    if info := command.codeProvider.GetCardIssuerInfo(command); info != nil {
        a = &info.Code
    }
}
another = a
```

```go
var a *string

if a := command.GetSomeValue(); a != nil && len(*a) == 4 {
   ^^^^
    if info := command.someProvider.GetSomeInfo(command); info != nil {
        a = &info.Code
    }
}
another = a
```

위의 코드중 전자는 `another`에 `a`가 잘 설정되지만, 후자는 `another`에 `a`가 설정되지 않는다. 이는 `=`와 `:=` 연산자의 사용의 차이 때문이며, 이 차이는 Go 언어의 변수 범위(scope)와 변수 선언에 관한 규칙 때문에 발생한다.

- `=` 연산자 사용 (첫 번째 코드)
    - 이 경우 `a`는 함수의 시작 부분에서 `var a *string`으로 선언되어 있다
    - `if` 블록 내에서 `a = command.GetCardIssuerCode()`는 이미 선언된 `a`에 새 값을 할당한다
    - 따라서 `if` 블록을 벗어난 후에도 `a`는 새로운 값을 유지하고, `another`에 할당된다
- `:=` 연산자 사용 (두 번째 코드)
    - 이 경우 `:=`는 **새로운 변수 `a`**를 `if` 블록의 지역 범위(scope) 내에서 선언하고 초기화한다
    - 이렇게 선언된 `a`는 `if` 블록 내에서만 유효하며, `if` 블록을 벗어나면 이 변수에 대한 접근은 더 이상 유효하지 않다.
    - 따라서 `if` 블록 밖의 `another`에는 함수 시작 부분에 선언된 `a`의 값이 할당되며, 이는 `nil` 또는 변경되지 않은 초기 값이 된다
