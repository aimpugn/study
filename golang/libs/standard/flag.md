# flag

## flag?

`flag` 패키지는 커맨드라인 옵션을 쉽게 정의하고 파싱할 수 있게 해준다.
각 옵션은 기본값을 가질 수 있으며, 사용자가 커맨드라인에서 해당 옵션을 제공할 경우 해당 값으로 설정된다.

## Examples

### `-name`, `-age` 옵션

```go
package main

import (
    "flag"
    "fmt"
)

func main() {
    var name string
    var age int
    flag.StringVar(&name, "name", "Anonymous", "Your name")
    flag.IntVar(&age, "age", 0, "Your age")
    flag.Parse()

    fmt.Printf("Hello, %s! You are %d years old.\n", name, age)
}
```

이 예제에서는 `-name`과 `-age`라는 두 개의 커맨드라인 옵션을 정의한다.
`flag.Parse()`를 호출하여 커맨드라인 인자를 파싱한다.

### 지정한 인자 외의 값 파싱하기

Go의 `flag` 패키지를 사용할 때, 명시적으로 정의된 플래그 이후의 인자들은 `flag.Args()` 함수를 통해 접근할 수 있다.
이 함수는 플래그가 아닌 모든 인자들을 포함하는 슬라이스를 반환한다.
따라서, `-name`과 `-age` 같은 플래그를 처리한 후에 남은 인자들을 `flag.Args()`로 받을 수 있다.

```go
package main

import (
    "flag"
    "fmt"
)

func main() {
    // 플래그 정의
    var name string
    var age int
    flag.StringVar(&name, "name", "", "Your name")
    flag.IntVar(&age, "age", 0, "Your age")
    
    // 플래그 파싱
    flag.Parse()
    
    // 플래그가 아닌 나머지 인자들 처리
    remainingArgs := flag.Args()
    
    fmt.Printf("Name: %s\n", name)
    fmt.Printf("Age: %d\n", age)
    
    // "dddddddddddd" 같은 추가 인자들 출력
    if len(remainingArgs) > 0 {
        fmt.Println("Additional arguments:")
        for _, arg := range remainingArgs {
            fmt.Println(arg)
        }
    }
}
```

이 예제에서 `flag.Parse()`는 모든 플래그를 파싱하고 처리합니다. 이 함수 호출 이후, `flag.Args()`를 호출하여 플래그로 지정되지 않은 모든 추가 인자들을 슬라이스 형태로 받을 수 있습니다.

따라서 위의 코드를 실행할 때 커맨드라인에서

```sh
go run main.go -name test -age 10 dddddddddddd
```

와 같이 입력하면, `-name`에는 "test"가, `-age`에는 `10`이 할당되고, `"dddddddddddd"`는 `flag.Args()`를 통해 접근할 수 있습니다. 이를 통해 프로그램은 모든 플래그와 추가적으로 제공된 인자들을 처리할 수 있습니다.
